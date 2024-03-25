package be

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
)

func (api *API) GetGovProposal(proposal uint64) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getGovProposal")
	return api.backend.GetGovProposal(proposal)
}

func (api *API) GetGovProposals(pageNoOptional *int) (berpctypes.GenericBackendResponse, error) {
	api.logger.Debug("be_getGovProposals")

	pageNo, err := getPageNumber(pageNoOptional)
	if err != nil {
		return nil, err
	}

	return api.backend.GetGovProposals(pageNo)
}
