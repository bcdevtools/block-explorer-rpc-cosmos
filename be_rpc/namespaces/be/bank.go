package be

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	tmmath "github.com/tendermint/tendermint/libs/math"
)

func (api *API) GetDenomMetadata(base string) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getDenomMetadata")
	return api.backend.GetDenomMetadata(base)
}

func (api *API) GetDenomsMetadata(pageNoOptional *int) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getDenomsMetadata")

	var pageNo int
	if pageNoOptional == nil {
		pageNo = 1
	} else {
		pageNo = tmmath.MaxInt(1, *pageNoOptional)
	}

	return api.backend.GetDenomsMetadata(pageNo)
}
