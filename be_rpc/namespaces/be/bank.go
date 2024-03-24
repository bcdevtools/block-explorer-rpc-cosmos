package be

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
)

func (api *API) GetDenomMetadata(base string) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getDenomMetadata")
	return api.backend.GetDenomMetadata(base)
}

func (api *API) GetDenomsMetadata(pageNoOptional *int) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getDenomsMetadata")

	pageNo, err := getPageNumber(pageNoOptional)
	if err != nil {
		return nil, err
	}

	return api.backend.GetDenomsMetadata(pageNo)
}

func (api *API) GetTotalSupply(pageNoOptional *int) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getTotalSupply")

	pageNo, err := getPageNumber(pageNoOptional)
	if err != nil {
		return nil, err
	}

	return api.backend.GetTotalSupply(pageNo)
}
