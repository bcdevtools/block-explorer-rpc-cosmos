package be

import berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"

func (api *API) GetTransactionsInBlockRange(fromHeightIncluded int64, toHeightIncluded *int64) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getTransactionsInBlockRange")

	var toHeightIncluded2 int64
	if toHeightIncluded == nil {
		toHeightIncluded2 = fromHeightIncluded
	} else {
		toHeightIncluded2 = *toHeightIncluded
	}
	return api.backend.GetTransactionsInBlockRange(fromHeightIncluded, toHeightIncluded2)
}

func (api *API) GetTransactionByHash(hash string) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getTransactionByHash")
	return api.backend.GetTransactionByHash(hash)
}
