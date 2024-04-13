package constants

type EvmAction string

const (
	EvmActionNone     EvmAction = ""
	EvmActionTransfer EvmAction = "transfer"
	EvmActionCall     EvmAction = "call"
	EvmActionCreate   EvmAction = "create"
)
