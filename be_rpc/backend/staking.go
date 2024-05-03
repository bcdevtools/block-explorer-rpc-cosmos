package backend

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

func (m *Backend) GetStakingInfo(delegatorAddr string) (berpctypes.GenericBackendResponse, error) {
	delegatorAddr = strings.ToLower(strings.TrimSpace(delegatorAddr))
	unsafeDelegatorAddr := m.bech32Cfg.FromAnyToBech32AccountAddrUnsafe(delegatorAddr)

	resDd, err := m.queryClient.StakingQueryClient.DelegatorDelegations(m.ctx, &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: unsafeDelegatorAddr,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get delegator delegations").Error())
	}

	resDist, err := m.queryClient.DistributionQueryClient.DelegationTotalRewards(m.ctx, &disttypes.QueryDelegationTotalRewardsRequest{
		DelegatorAddress: unsafeDelegatorAddr,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get delegation total rewards").Error())
	}

	validatorCommission := sdk.DecCoins{}
	validatorOutstandingRewards := sdk.DecCoins{}

	isValidatorAddress := m.bech32Cfg.IsValAddr(delegatorAddr)
	if isValidatorAddress {
		resCom, err := m.queryClient.DistributionQueryClient.ValidatorCommission(m.ctx, &disttypes.QueryValidatorCommissionRequest{
			ValidatorAddress: delegatorAddr,
		})
		if err != nil {
			m.GetLogger().Error("failed to get validator commission", "error", err)
		} else {
			validatorCommission = resCom.Commission.Commission
		}

		resOutRew, err := m.queryClient.DistributionQueryClient.ValidatorOutstandingRewards(m.ctx, &disttypes.QueryValidatorOutstandingRewardsRequest{
			ValidatorAddress: delegatorAddr,
		})
		if err != nil {
			m.GetLogger().Error("failed to get validator outstanding rewards", "error", err)
		} else {
			validatorOutstandingRewards = resOutRew.Rewards.Rewards
		}
	}

	totalRewards := sdk.DecCoins{}
	for _, reward := range resDist.Rewards {
		totalRewards = totalRewards.Add(reward.Reward...)
	}

	resStakingInfo := make(berpctypes.GenericBackendResponse)
	for _, delegation := range resDd.DelegationResponses {
		resStakingInfo[delegation.Delegation.ValidatorAddress] = delegation.Balance.Amount.String()
	}

	res := berpctypes.GenericBackendResponse{
		"staking": resStakingInfo,
		"rewards": totalRewards.String(),
	}

	if !validatorCommission.IsZero() {
		res["validatorCommission"] = validatorCommission.String()
	}
	if !validatorOutstandingRewards.IsZero() {
		res["validatorOutstandingRewards"] = validatorOutstandingRewards.String()
	}

	return res, nil
}

func (m *Backend) GetValidators() (berpctypes.GenericBackendResponse, error) {
	tmValidators, err := m.tendermintValidatorsCache.GetValidators()
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get tendermint validators").Error())
	}

	stakingValidators, err := m.stakingValidatorsCache.GetValidators()
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get staking validators").Error())
	}

	res := make(berpctypes.GenericBackendResponse)

	for _, stakingValidator := range stakingValidators {
		consAddr := stakingValidator.consAddr
		valInfo := map[string]any{
			"consAddress": consAddr,
			"valAddress":  stakingValidator.validator.OperatorAddress,
			"pubKeyType":  "",
			"votingPower": -1,
		}

		for _, tmValidator := range tmValidators {
			if sdk.ConsAddress(tmValidator.Address).String() != consAddr {
				continue
			}

			valInfo["pubKeyType"] = tmValidator.PubKey.Type()
			valInfo["votingPower"] = tmValidator.VotingPower
			break
		}

		res[consAddr] = valInfo
	}

	return res, nil
}
