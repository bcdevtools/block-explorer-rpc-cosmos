package be

import berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"

func (api *API) GetStakingInfo(delegatorAddr string) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getStakingInfo")
	return api.backend.GetStakingInfo(delegatorAddr)
}

func (api *API) GetValidators() (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getValidators")
	return api.backend.GetValidators()
}
