package constants

type WasmAction string

const (
	WasmActionNone     WasmAction = ""
	WasmActionTransfer WasmAction = "transfer"
	WasmActionCall     WasmAction = "call"
	WasmActionCreate   WasmAction = "create"
)
