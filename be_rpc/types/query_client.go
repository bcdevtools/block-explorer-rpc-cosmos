package types

import (
	"github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
)

// QueryClient defines a gRPC Client used for:
//   - Transaction simulation
type QueryClient struct {
	tx.ServiceClient

	BankQueryClient         banktypes.QueryClient
	StakingQueryClient      stakingtypes.QueryClient
	DistributionQueryClient disttypes.QueryClient
	GovV1QueryClient        govv1types.QueryClient
	MintQueryClient         minttypes.QueryClient
}

// NewQueryClient creates a new gRPC query client
func NewQueryClient(clientCtx client.Context) *QueryClient {
	return &QueryClient{
		ServiceClient:           tx.NewServiceClient(clientCtx),
		BankQueryClient:         banktypes.NewQueryClient(clientCtx),
		StakingQueryClient:      stakingtypes.NewQueryClient(clientCtx),
		DistributionQueryClient: disttypes.NewQueryClient(clientCtx),
		GovV1QueryClient:        govv1types.NewQueryClient(clientCtx),
		MintQueryClient:         minttypes.NewQueryClient(clientCtx),
	}
}
