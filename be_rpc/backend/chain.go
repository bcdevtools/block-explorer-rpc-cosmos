package backend

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

func (m *Backend) GetChainInfo() (berpctypes.GenericBackendResponse, error) {
	statusInfo, err := m.clientCtx.Client.Status(m.ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	denoms := make(map[string]string)
	var intercepted bool
	if m.interceptor != nil {
		intercepted, _, denoms, err = m.interceptor.GetDenomsInformation()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if denoms == nil {
			denoms = make(map[string]string) // re-initialize
		}
	}

	if !intercepted {
		stakingParams, err := m.queryClient.StakingQueryClient.Params(m.ctx, &stakingtypes.QueryParamsRequest{})
		if err != nil {
			return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get staking params").Error())
		}
		denoms["bond"] = stakingParams.Params.BondDenom
	}

	return berpctypes.GenericBackendResponse{
		"chainType":               m.externalServices.ChainType,
		"chainId":                 statusInfo.NodeInfo.Network,
		"latestBlock":             statusInfo.SyncInfo.LatestBlockHeight,
		"latestBlockTimeEpochUTC": statusInfo.SyncInfo.LatestBlockTime.UTC().Unix(),
		"bech32": map[string]string{
			"addr": sdk.GetConfig().GetBech32AccountAddrPrefix(),
			"val":  sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
			"cons": sdk.GetConfig().GetBech32ConsensusAddrPrefix(),
		},
		"denoms": denoms,
	}, nil
}

func (m *Backend) GetModuleParams(moduleName string) (berpctypes.GenericBackendResponse, error) {
	moduleName = strings.TrimSpace(strings.ToLower(moduleName))

	if m.interceptor != nil {
		intercepted, response, err := m.interceptor.GetModuleParams(moduleName)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if intercepted {
			return response, nil
		}
	}

	var params any
	var err error

	switch moduleName {
	case "bank":
		bankParams, errFetch := m.queryClient.BankQueryClient.Params(m.ctx, &banktypes.QueryParamsRequest{})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get bank params")
		} else {
			params = bankParams.Params
		}
		break
	case "staking":
		stakingParams, errFetch := m.queryClient.StakingQueryClient.Params(m.ctx, &stakingtypes.QueryParamsRequest{})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get staking params")
		} else {
			params = stakingParams.Params
		}
		break
	case "distribution":
		distributionParams, errFetch := m.queryClient.DistributionQueryClient.Params(m.ctx, &disttypes.QueryParamsRequest{})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get distribution params")
		} else {
			params = distributionParams.Params
		}
		break
	case "gov":
		govParams, errFetch := m.queryClient.GovV1QueryClient.Params(m.ctx, &govv1types.QueryParamsRequest{})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get gov params")
		} else {
			params = govParams
		}
		break
	case "mint":
		mintParams, errFetch := m.queryClient.MintQueryClient.Params(m.ctx, &minttypes.QueryParamsRequest{})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get mint params")
		} else {
			params = mintParams.Params
		}
		break
	default:
		err = errors.Errorf("not yet support module %s", moduleName)
		break
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res, err := berpctypes.NewGenericBackendResponseFrom(params)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "module params").Error())
	}

	return res, nil
}
