package backend

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *Backend) GetDenomMetadata(base string) (berpctypes.GenericBackendResponse, error) {
	resDenomMetadata, err := m.queryClient.BankQueryClient.DenomMetadata(m.ctx, &banktypes.QueryDenomMetadataRequest{
		Denom: base,
	})
	if err != nil {
		return nil, err
	}

	res := make(berpctypes.GenericBackendResponse)
	rpcDenomMetadata := berpctypes.NewRpcDenomMetadataFromBankMetadata(resDenomMetadata.Metadata)
	res[resDenomMetadata.Metadata.Base] = rpcDenomMetadata

	return res, nil
}

func (m *Backend) GetDenomsMetadata(pageNo int) (berpctypes.GenericBackendResponse, error) {
	if pageNo < 1 {
		return nil, berpctypes.ErrBadPageNo
	}

	resDenomMetadata, err := m.queryClient.BankQueryClient.DenomsMetadata(m.ctx, &banktypes.QueryDenomsMetadataRequest{
		Pagination: getDefaultPagination(pageNo),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get denoms metadata").Error())
	}

	res := make(berpctypes.GenericBackendResponse)
	for _, metadata := range resDenomMetadata.Metadatas {
		rpcDenomMetadata := berpctypes.NewRpcDenomMetadataFromBankMetadata(metadata)
		res[metadata.Base] = rpcDenomMetadata
	}

	return res, nil
}

func (m *Backend) GetTotalSupply(pageNo int) (berpctypes.GenericBackendResponse, error) {
	if pageNo < 1 {
		return nil, berpctypes.ErrBadPageNo
	}

	resTotalSupply, err := m.queryClient.BankQueryClient.TotalSupply(m.ctx, &banktypes.QueryTotalSupplyRequest{
		Pagination: getDefaultPagination(pageNo),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get total supply").Error())
	}

	res := make(berpctypes.GenericBackendResponse)
	for _, coin := range resTotalSupply.Supply {
		res[coin.Denom] = coin.Amount.String()
	}

	return res, nil
}
