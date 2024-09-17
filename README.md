## Service Overview: Optimistic Blockchain Listener

This service is designed to listen to Optimistic blockchain networks for the latest blocks and transactions (Txs). Optimistic blockchains generally operate with two types of nodes:

1. [**Replica Node**](https://docs.optimism.io/builders/chain-operators/architecture#replica-node)
2. [**Sequencer Node**](https://docs.optimism.io/builders/chain-operators/architecture#sequencer)

### Network Components

Optimistic networks are composed of multiple [**Permissioned Components**](https://docs.optimism.io/builders/chain-operators/architecture#permissioned-components):
1. **op-geth**
2. **op-node**
3. **op-batcher**
4. **op-proposer**

### Node Roles: Sequencer vs. Replica Node

The main difference between Sequencer and Replica nodes is their access to `op-batcher` and `op-proposer`. All replica nodes send their transactions (Txs) to the Sequencer (there is only one Sequencer deployed by the Optimism team). The Sequencer then adds the transactions into a block, broadcasts the block to all network nodes, and periodically writes these blocks to L1.

### Service Behavior

This service functions like a lightweight replica node. It fetches the latest blocks from the Sequencer (or other trusted peers through P2P gossip) upon block publication.

The following callback function in `bootnode` is triggered when a new block is received:

```go
func (g *gossipIn) OnUnsafeL2Payload(_ context.Context, _ peer.ID, msg *eth.ExecutionPayloadEnvelope) error
```

In this function, we can access the `msg` (a pointer to `eth.ExecutionPayloadEnvelope`):

```go
type ExecutionPayloadEnvelope struct {
	ParentBeaconBlockRoot *common.Hash      `json:"parentBeaconBlockRoot,omitempty"`
	ExecutionPayload      *ExecutionPayload `json:"executionPayload"`
}

type ExecutionPayload struct {
	ParentHash    common.Hash     `json:"parentHash"`
	FeeRecipient  common.Address  `json:"feeRecipient"`
	StateRoot     Bytes32         `json:"stateRoot"`
	ReceiptsRoot  Bytes32         `json:"receiptsRoot"`
	LogsBloom     Bytes256        `json:"logsBloom"`
	PrevRandao    Bytes32         `json:"prevRandao"`
	BlockNumber   Uint64Quantity  `json:"blockNumber"`
	GasLimit      Uint64Quantity  `json:"gasLimit"`
	GasUsed       Uint64Quantity  `json:"gasUsed"`
	Timestamp     Uint64Quantity  `json:"timestamp"`
	ExtraData     BytesMax32      `json:"extraData"`
	BaseFeePerGas Uint256Quantity `json:"baseFeePerGas"`
	BlockHash     common.Hash     `json:"blockHash"`
	Transactions  []Data          `json:"transactions"`
	Withdrawals   *types.Withdrawals `json:"withdrawals,omitempty"`
	BlobGasUsed   *Uint64Quantity `json:"blobGasUsed,omitempty"`
	ExcessBlobGas *Uint64Quantity `json:"excessBlobGas,omitempty"`
}
```

### Extracting Transactions (Txs)

To extract the transactions from the received block, we can access `msg.ExecutionPayload.Transactions`, which is an array of `Data`.

The `Data` type is defined as:
```go
type Data = hexutil.Bytes
```

We can iterate through the `msg.ExecutionPayload.Transactions` and parse the `hexutil.Bytes` into Ethereum transactions.

Ethereum transaction types can be found in the [EIPs](https://eips.ethereum.org/).

In the `bootnode.go` file, the `decodeTransaction()` method decodes the transactions into:

```go
type Transaction struct {
	inner TxData    // Consensus contents of a transaction
	time  time.Time // Time first seen locally (spam avoidance)

	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value

	// cache of details to compute the data availability fee
	rollupCostData atomic.Value
}
```

### Accessing Transaction Details

The `TxData` interface provides methods to access details of the transaction:

```go
type TxData interface {
	txType() byte // returns the type ID
	copy() TxData // creates a deep copy and initializes all fields

	chainID() *big.Int
	accessList() AccessList
	data() []byte
	gas() uint64
	gasPrice() *big.Int
	gasTipCap() *big.Int
	gasFeeCap() *big.Int
	value() *big.Int
	nonce() uint64
	to() *common.Address
	isSystemTx() bool
}
```

Using this, we can access transaction details such as:

```go
	p := txDto{
		Hash:  tx.Hash().Hex(),
		Nonce: tx.Nonce(),
		Value: tx.Value().String(),
		Data:  hex.EncodeToString(tx.Data()),
	}
```

### Pushing to Redis

We then stringify and push the extracted transaction data to a Redis queue (`LPush`) using the following type:

```go
type txDto struct {
	Hash  string `json:"hash"`
	To    string `json:"to"`
	Nonce uint64 `json:"nonce"`
	Value string `json:"value"`
	Data  string `json:"data"`
}
```

---

## How to Run the Node

### Prerequisites

To run the `op-node`, you need to prepare the following:

1. A `.config` file for the network you want the node to participate in (SHOULD place in the root directory of the project).
2. A `rollup.json` file for the network (place in the root directory of the project or anywhere else).
3. `P2P_SEQUENCER_ADDRESS` for the network (set as an environment variable).
4. `REDIS_CONN_STRING` (set as an environment variable).
5. `REDIS_STREAM_NAME` (set as an environment variable).

**Note: Golang reads environment variables directly from the OS environment, not from a .env file (unlike Node.js, etc.).**

### Building the Project

To build the project, run:
```shell
$ go build main.go
```

### Viewing Help

You can see the help options by running:
```shell
$ ./op-light-node --help
```

The `op-light-node` requires a flag `--rollup.config` which should point to the `.config` file of the network you want to participate in.

### Example Command

To start the node with the Mantle network configuration:
```shell
$ ./op-light-node --rollup.config="./config/mantle/rollup.json"
```

New transactions will be logged to the console and pushed to the Redis queue as well.
