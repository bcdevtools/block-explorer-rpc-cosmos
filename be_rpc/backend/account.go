package backend

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	berpcutils "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *Backend) GetAccountBalances(accountAddressStr string, denom *string) (berpctypes.GenericBackendResponse, error) {
	accAddrStr := m.bech32Cfg.ConvertToAccAddressIfHexOtherwiseKeepAsIs(accountAddressStr)

	if denom == nil || len(*denom) == 0 {
		resAllBalances, err := m.queryClient.BankQueryClient.AllBalances(m.ctx, &banktypes.QueryAllBalancesRequest{
			Address: accAddrStr,
		})
		if err != nil {
			return nil, err
		}

		res := make(berpctypes.GenericBackendResponse)
		for _, coin := range resAllBalances.Balances {
			res[coin.Denom] = coin.Amount.String()
		}

		return res, nil
	}

	resBalance, err := m.queryClient.BankQueryClient.Balance(m.ctx, &banktypes.QueryBalanceRequest{
		Address: accAddrStr,
		Denom:   *denom,
	})
	if err != nil {
		return nil, err
	}

	res := make(berpctypes.GenericBackendResponse)
	res[resBalance.Balance.Denom] = resBalance.Balance.Amount.String()
	return res, nil
}

func (m *Backend) GetAccount(accountAddressStr string) (berpctypes.GenericBackendResponse, error) {
	res := make(berpctypes.GenericBackendResponse)

	if m.interceptor != nil {
		var intercepted bool
		var err error
		intercepted, _, res, err = m.interceptor.GetAccount(accountAddressStr)
		if err != nil {
			return nil, err
		}
		if intercepted {
			return res, nil
		}
		if res == nil {
			// re-initialize
			res = make(berpctypes.GenericBackendResponse)
		}
	}

	accAddrStr := m.bech32Cfg.ConvertToAccAddressIfHexOtherwiseKeepAsIs(accountAddressStr)

	addressInfo := berpctypes.GenericBackendResponse{
		"cosmos": accAddrStr,
	}
	if m.externalServices.ChainType == berpctypes.ChainTypeEvm {
		accAddr, err := sdk.AccAddressFromBech32(accAddrStr)
		if err == nil {
			addressInfo["evm"] = common.BytesToAddress(accAddr.Bytes()).Hex()
		}
	}

	res["address"] = addressInfo

	_, isSmartContract := res["contract"]
	isValidatorAddress := !isSmartContract && m.bech32Cfg.IsValAddr(accAddrStr)

	// get account balance

	if !isValidatorAddress {
		balancesInfo, err := m.GetAccountBalances(accAddrStr, nil)
		if err != nil {
			return nil, err
		}

		res["balances"] = balancesInfo
	}

	// get account transaction count

	if !isSmartContract && !isValidatorAddress {
		resAccount, err := m.queryClient.AuthQueryClient.Account(m.ctx, &authtypes.QueryAccountRequest{
			Address: accAddrStr,
		})

		if err == nil && resAccount != nil && resAccount.Account != nil {
			fakeBaseAccount := &berpctypes.FakeBaseAccount{}
			extractedSuccess, err := fakeBaseAccount.TryUnmarshalFromProto(resAccount.Account, m.clientCtx.Codec)
			if err == nil && extractedSuccess {
				res["txs_count"] = fakeBaseAccount.Sequence + 1
			} else if err != nil {
				m.GetLogger().Error("failed to extract base account", "error", err)
			}
		}
	}

	// get staking information

	if !isSmartContract && !isValidatorAddress {
		stakingInfo, err := m.GetStakingInfo(accAddrStr)
		if err != nil {
			return nil, err
		}
		res["staking"] = stakingInfo
	}

	// get validator information

	if isValidatorAddress {
		resValInfo, err := m.queryClient.StakingQueryClient.Validator(m.ctx, &stakingtypes.QueryValidatorRequest{
			ValidatorAddr: accAddrStr,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get validator info").Error())
		}

		validatorInfo := berpctypes.GenericBackendResponse{
			"operator_address": resValInfo.Validator.OperatorAddress,
			"jailed":           resValInfo.Validator.Jailed,
			"status":           resValInfo.Validator.Status.String(),
			"tokens":           resValInfo.Validator.Tokens.String(),
			"delegator_shares": resValInfo.Validator.DelegatorShares.String(),
			"description": berpctypes.GenericBackendResponse{
				"moniker":          resValInfo.Validator.Description.Moniker,
				"identity":         resValInfo.Validator.Description.Identity,
				"website":          resValInfo.Validator.Description.Website,
				"security_contact": resValInfo.Validator.Description.SecurityContact,
				"details":          resValInfo.Validator.Description.Details,
			},
			"unbonding_height":    resValInfo.Validator.UnbondingHeight,
			"unbonding_time":      resValInfo.Validator.UnbondingTime,
			"commission":          resValInfo.Validator.Commission,
			"min_self_delegation": resValInfo.Validator.MinSelfDelegation.String(),
		}

		consensusPubKeyMap, err := berpcutils.FromAnyToJsonMap(resValInfo.Validator.ConsensusPubkey, m.clientCtx.Codec)
		if err == nil {
			validatorInfo["consensus_pubkey"] = consensusPubKeyMap
		}

		consAddr, success := berpcutils.FromAnyPubKeyToConsensusAddress(resValInfo.Validator.ConsensusPubkey, m.clientCtx.Codec)
		if success {
			validatorInfo["consensus_address"] = consAddr.String()
			tmVals, err := m.tendermintValidatorsCache.GetValidators()
			if err != nil {
				return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get validators").Error())
			}
			for _, val := range tmVals {
				if sdk.ConsAddress(val.Address).String() == consAddr.String() {
					validatorInfo["voting_power"] = val.VotingPower
					break
				}
			}
		}

		res["validator"] = validatorInfo
	}

	return res, nil
}
