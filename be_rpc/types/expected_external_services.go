package types

import (
	"github.com/ethereum/go-ethereum/common"
)

type ExternalServices struct {
	ChainType    string
	EvmTxIndexer ExpectedEVMTxIndexer
}

const (
	ChainTypeDefault  = "cosmos"
	ChainTypeEvm      = "evm"
	ChainTypeCosmWasm = "wasm"
)

// ExpectedEVMTxIndexer defines the interface of custom eth tx indexer.
type ExpectedEVMTxIndexer interface {
	LastIndexedBlock() (int64, error)

	// GetByTxHashForExternal returns nil if tx not found.
	GetByTxHashForExternal(common.Hash) (TxResultForExternal, error)

	// GetIndexer returns the indexer instance.
	GetIndexer() any
}

type TxResultForExternal interface {
	GetHeight() int64
	GetTxIndex() uint32
	GetMsgIndex() uint32
	GetEthTxIndex() int32
	GetFailed() bool
	GetGasUsed() uint64
	GetCumulativeGasUsed() uint64
}
