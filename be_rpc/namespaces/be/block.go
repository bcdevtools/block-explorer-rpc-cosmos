package be

import berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"

func (api *API) GetBlockByNumber(height int64) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getBlockByNumber")
	return api.backend.GetBlockByNumber(height)
}
