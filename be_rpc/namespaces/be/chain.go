package be

import berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"

func (api *API) GetChainInfo() (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getChainInfo")
	return api.backend.GetChainInfo()
}

func (api *API) GetModuleParams(moduleName string) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getModuleParams")
	return api.backend.GetModuleParams(moduleName)
}
