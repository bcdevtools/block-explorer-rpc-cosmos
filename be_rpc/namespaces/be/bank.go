package be

import berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"

func (api *API) GetDenomsMetadata() (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getDenomsMetadata")
	return api.backend.GetDenomsMetadata()
}
