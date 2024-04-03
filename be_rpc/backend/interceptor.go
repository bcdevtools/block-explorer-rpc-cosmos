package backend

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
)

// RequestInterceptor is the interface for the request interceptor.
// It is used to intercept the request and return the response.
//   - If the intercepted is true, the response from this interceptor must be respected.
//   - If the intercepted is false, the response from this interceptor must be ignored and processed as usual.
type RequestInterceptor interface {
	GetTransactionByHash(hashStr string) (intercepted bool, response berpctypes.GenericBackendResponse, err error)

	GetDenomsInformation() (intercepted, append bool, denoms map[string]string, err error)

	GetModuleParams(moduleName string) (intercepted bool, params berpctypes.GenericBackendResponse, err error)

	GetAccount(accountAddressStr string) (intercepted, append bool, response berpctypes.GenericBackendResponse, err error)
	// TODO BE: module wasm
}
