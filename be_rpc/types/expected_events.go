package types

const (
	EventTypeEthereumTx = "ethereum_tx"
	EventTypeBlockBloom = "block_bloom"
	EventTypeTxLog      = "tx_log"

	AttributeKeyContractAddress = "contract"
	AttributeKeyRecipient       = "recipient"
	AttributeKeyTxHash          = "txHash"
	AttributeKeyEthereumTxHash  = "ethereumTxHash"
	AttributeKeyTxIndex         = "txIndex"
	AttributeKeyTxGasUsed       = "txGasUsed"
	AttributeKeyTxType          = "txType"
	AttributeKeyTxLog           = "txLog"

	AttributeKeyEthereumTxFailed = "ethereumTxFailed"
	AttributeValueCategory       = "evm"
	AttributeKeyEthereumBloom    = "bloom"

	MetricKeyTransitionDB = "transition_db"
	MetricKeyStaticCall   = "static_call"
)
