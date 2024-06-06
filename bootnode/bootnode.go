package bootnode

import (
	"context"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/urfave/cli/v2"
	"github.com/zigbalthazar/op-light-p2p-node/queue"
	"github.com/zigbalthazar/op-light-p2p-node/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	opnode "github.com/ethereum-optimism/optimism/op-node"
	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/p2p"
	p2pcli "github.com/ethereum-optimism/optimism/op-node/p2p/cli"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	opflags "github.com/ethereum-optimism/optimism/op-service/flags"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/opio"
	oprpc "github.com/ethereum-optimism/optimism/op-service/rpc"
)

var stream *queue.Queue

type gossipIn struct{}

func (g *gossipIn) OnUnsafeL2Payload(_ context.Context, _ peer.ID, msg *eth.ExecutionPayloadEnvelope) error {
	fmt.Println("New block received")
	stringData, err := utils.StructToMap(msg)
	if err != nil {
		panic(err)
	}

	fmt.Println(stringData)

	stream.Add("mantle", stringData) // TODO : env
	return nil
}

type gossipConfig struct{}

func (g *gossipConfig) P2PSequencerAddress() common.Address {
	return common.HexToAddress("0xAAC979CBeE00C75C35DE9a2635d8B75940F466dc") // base-mainnet // TODO : env
}

type l2Chain struct{}

func (l *l2Chain) PayloadByNumber(_ context.Context, _ uint64) (*eth.ExecutionPayloadEnvelope, error) {
	return nil, errors.New("P2P req/resp is not supported in bootnodes")
}

func Main(cliCtx *cli.Context) error {
	log.Info("Initializing bootnode")
	logCfg := oplog.ReadCLIConfig(cliCtx)
	logger := oplog.NewLogger(oplog.AppOut(cliCtx), logCfg)
	oplog.SetGlobalLogHandler(logger.Handler())
	m := metrics.NewMetrics("default")
	ctx := context.Background()

	log.Info("Connecting to redis stream")
	stream = queue.Init("redis://localhost:6379", ctx) // TODO : env

	network := cliCtx.String(opflags.NetworkFlagName)
	rollupConfigPath := cliCtx.String(opflags.RollupConfigFlagName)

	fmt.Println(network)
	fmt.Println(rollupConfigPath)

	config, err := opnode.NewRollupConfigFromCLI(logger, cliCtx)
	if err != nil {
		return err
	}
	if err = validateConfig(config); err != nil {
		return err
	}

	p2pConfig, err := p2pcli.NewConfig(cliCtx, config)
	if err != nil {
		return fmt.Errorf("failed to load p2p config: %w", err)
	}
	if p2pConfig.EnableReqRespSync {
		logger.Warn("req-resp sync is enabled, bootnode does not support this feature")
		p2pConfig.EnableReqRespSync = false
	}

	gossipHandler := &gossipIn{}
	p2pNode, err := p2p.NewNodeP2P(ctx, config, logger, p2pConfig, gossipHandler, &l2Chain{}, &gossipConfig{}, m, false)
	if err != nil || p2pNode == nil {
		return err
	}
	if p2pNode.Dv5Udp() == nil {
		return fmt.Errorf("uninitialized discovery service")
	}

	rpcCfg := oprpc.ReadCLIConfig(cliCtx)
	if err := rpcCfg.Check(); err != nil {
		return fmt.Errorf("failed to validate RPC config")
	}
	rpcServer := oprpc.NewServer(rpcCfg.ListenAddr, rpcCfg.ListenPort, "", oprpc.WithLogger(logger))
	if rpcCfg.EnableAdmin {
		logger.Info("Admin RPC enabled but does nothing for the bootnode")
	}
	rpcServer.AddAPI(rpc.API{
		Namespace:     p2p.NamespaceRPC,
		Version:       "",
		Service:       p2p.NewP2PAPIBackend(p2pNode, logger, m),
		Authenticated: false,
	})
	if err := rpcServer.Start(); err != nil {
		return fmt.Errorf("failed to start the RPC server")
	}
	defer func() {
		if err := rpcServer.Stop(); err != nil {
			log.Error("failed to stop RPC server", "err", err)
		}
	}()

	go p2pNode.DiscoveryProcess(ctx, logger, config, p2pConfig.TargetPeers())

	metricsCfg := opmetrics.ReadCLIConfig(cliCtx)
	if metricsCfg.Enabled {
		log.Debug("starting metrics server", "addr", metricsCfg.ListenAddr, "port", metricsCfg.ListenPort)
		metricsSrv, err := m.StartServer(metricsCfg.ListenAddr, metricsCfg.ListenPort)
		if err != nil {
			return fmt.Errorf("failed to start metrics server: %w", err)
		}
		defer func() {
			if err := metricsSrv.Stop(context.Background()); err != nil {
				log.Error("failed to stop metrics server", "err", err)
			}
		}()
		log.Info("started metrics server", "addr", metricsSrv.Addr())
		m.RecordUp()
	}

	opio.BlockOnInterrupts()

	return nil
}

func validateConfig(config *rollup.Config) error {
	if config.L2ChainID == nil || config.L2ChainID.Uint64() == 0 {
		return errors.New("chain ID is not set")
	}
	if config.Genesis.L2Time <= 0 {
		return errors.New("genesis timestamp is not set")
	}
	if config.BlockTime <= 0 {
		return errors.New("block time is not set")
	}
	return nil
}
