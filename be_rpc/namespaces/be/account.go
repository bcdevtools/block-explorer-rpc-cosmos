package be

import berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"

func (api *API) GetAccountBalances(accountAddressStr string, denom *string) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getAccountBalances")
	return api.backend.GetAccountBalances(accountAddressStr, denom)
}
