package backend

import (
	"github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/constants"
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
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
			"addr": m.bech32Cfg.GetBech32AccountAddrPrefix(),
			"val":  m.bech32Cfg.GetBech32ValidatorAddrPrefix(),
			"cons": m.bech32Cfg.GetBech32ConsensusAddrPrefix(),
		},
		"denoms": denoms,
		"version": map[string]string{
			"be-rpc-cosmos": constants.BlockExplorerRpcCosmosVersion,
		},
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
			break
		}

		params = bankParams.Params
	case "staking":
		stakingParams, errFetch := m.queryClient.StakingQueryClient.Params(m.ctx, &stakingtypes.QueryParamsRequest{})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get staking params")
			break
		}

		params = stakingParams.Params
	case "distribution":
		distributionParams, errFetch := m.queryClient.DistributionQueryClient.Params(m.ctx, &disttypes.QueryParamsRequest{})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get distribution params")
			break
		}

		params = distributionParams.Params
	case "gov":
		votingParams, errFetch := m.queryClient.GovV1QueryClient.Params(m.ctx, &govv1types.QueryParamsRequest{
			ParamsType: govv1types.ParamVoting,
		})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get gov voting params")
			break
		}

		tallyParams, errFetch := m.queryClient.GovV1QueryClient.Params(m.ctx, &govv1types.QueryParamsRequest{
			ParamsType: govv1types.ParamTallying,
		})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get gov tallying params")
			break
		}

		depositParams, errFetch := m.queryClient.GovV1QueryClient.Params(m.ctx, &govv1types.QueryParamsRequest{
			ParamsType: govv1types.ParamDeposit,
		})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get gov deposit params")
			break
		}

		params = &govv1types.QueryParamsResponse{
			VotingParams:  votingParams.VotingParams,
			TallyParams:   tallyParams.TallyParams,
			DepositParams: depositParams.DepositParams,
		}
	case "mint":
		mintParams, errFetch := m.queryClient.MintQueryClient.Params(m.ctx, &minttypes.QueryParamsRequest{})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get mint params")
			break
		}

		params = mintParams.Params
	case "auth":
		authParams, errFetch := m.queryClient.AuthQueryClient.Params(m.ctx, &authtypes.QueryParamsRequest{})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get auth params")
			break
		}

		params = authParams.Params
	case "ibc-transfer":
		ibcTransferParams, errFetch := m.queryClient.IbcTransferQueryClient.Params(m.ctx, &ibctransfertypes.QueryParamsRequest{})
		if errFetch != nil {
			err = errors.Wrap(errFetch, "failed to get ibc-transfer params")
			break
		}

		params = ibcTransferParams.Params
	default:
		err = errors.Errorf("not yet support module %s", moduleName)
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
