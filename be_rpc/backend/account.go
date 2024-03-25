package backend

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	berpcutils "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

func (m *Backend) GetAccountBalances(accountAddressStr string, denom *string) (berpctypes.GenericBackendResponse, error) {
	accAddrStr := berpcutils.ConvertToAccAddressIfHexOtherwiseKeepAsIs(accountAddressStr)

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
	}

	accAddrStr := berpcutils.ConvertToAccAddressIfHexOtherwiseKeepAsIs(accountAddressStr)

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

	balancesInfo, err := m.GetAccountBalances(accAddrStr, nil)
	if err != nil {
		return nil, err
	}

	res["balances"] = balancesInfo

	if res["contract"] == nil {
		stakingInfo, err := m.GetStakingInfo(accAddrStr)
		if err != nil {
			return nil, err
		}
		res["staking"] = stakingInfo

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

		if strings.HasPrefix(accAddrStr, sdk.GetConfig().GetBech32ValidatorAddrPrefix()+"1") {
			resVal, err := m.queryClient.StakingQueryClient.Validator(m.ctx, &stakingtypes.QueryValidatorRequest{
				ValidatorAddr: accAddrStr,
			})
			if err == nil && resVal != nil {
				validatorInfo := berpctypes.GenericBackendResponse{
					"operator_address": resVal.Validator.OperatorAddress,
					"consensus_pubkey": resVal.Validator.ConsensusPubkey,
					"jailed":           resVal.Validator.Jailed,
					"status":           resVal.Validator.Status.String(),
					"tokens":           resVal.Validator.Tokens.String(),
					"delegator_shares": resVal.Validator.DelegatorShares.String(),
					"description": berpctypes.GenericBackendResponse{
						"moniker":          resVal.Validator.Description.Moniker,
						"identity":         resVal.Validator.Description.Identity,
						"website":          resVal.Validator.Description.Website,
						"security_contact": resVal.Validator.Description.SecurityContact,
						"details":          resVal.Validator.Description.Details,
					},
					"unbonding_height":    resVal.Validator.UnbondingHeight,
					"unbonding_time":      resVal.Validator.UnbondingTime,
					"commission":          resVal.Validator.Commission,
					"min_self_delegation": resVal.Validator.MinSelfDelegation.String(),
				}

				res["validator"] = validatorInfo
			} else if err != nil {
				m.GetLogger().Error("failed to get validator", "error", err)
			}
		}
	}

	return res, nil
}
