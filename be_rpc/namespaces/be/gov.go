package be

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	tmmath "github.com/tendermint/tendermint/libs/math"
)

func (api *API) GetGovProposals(pageNoOptional *int) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getGovProposals")

	var pageNo int
	if pageNoOptional == nil {
		pageNo = 1
	} else {
		pageNo = tmmath.MaxInt(1, *pageNoOptional)
	}

	return api.backend.GetGovProposals(pageNo)
}
