package be

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
)

func (api *API) GetAccountBalances(accountAddressStr string, denom *string) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getAccountBalances")
	return api.backend.GetAccountBalances(accountAddressStr, denom)
}

func (api *API) GetAccount(accountAddressStr string) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getAccount")
	return api.backend.GetAccount(accountAddressStr)
}

func (api *API) GetValidatorAccount(consOrValAddr string) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getValidatorAccount")
	return api.backend.GetValidatorAccount(consOrValAddr)
}
