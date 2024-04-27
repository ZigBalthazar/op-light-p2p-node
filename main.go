package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/urfave/cli/v2"

	opnode "github.com/ethereum-optimism/optimism/op-node"
	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/p2p"
	p2pcli "github.com/ethereum-optimism/optimism/op-node/p2p/cli"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/opio"
	oprpc "github.com/ethereum-optimism/optimism/op-service/rpc"
)




func main(){
	log.Info("Initializing bootnode")
	logCfg := oplog.ReadCLIConfig(cliCtx)
	logger := oplog.NewLogger(oplog.AppOut(cliCtx), logCfg)
	oplog.SetGlobalLogHandler(logger.Handler())
	m := metrics.NewMetrics("default")
	ctx := context.Background()

	config, err := opnode.NewRollupConfigFromCLI(logger, cliCtx)
	if err != nil {
		return err
	}
	if err = validateConfig(config); err != nil {
		return err
	}
}