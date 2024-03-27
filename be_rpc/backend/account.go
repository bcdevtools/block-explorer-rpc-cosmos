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
	"strings"
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
	accountAddressStr = berpcutils.NormalizeAddress(accountAddressStr)
	if !m.isAccAddrOr0x(accountAddressStr) {
		return nil, berpctypes.ErrBadAddress
	}

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
		res = res.ReInitializeIfNil()
	}

	accAddrStr := m.bech32Cfg.ConvertToAccAddressIfHexOtherwiseKeepAsIs(accountAddressStr)

	addressInfo := berpctypes.GenericBackendResponse{
		"cosmos": accAddrStr,
	}
	if m.externalServices.ChainType == berpctypes.ChainTypeEvm {
		accAddr, err := sdk.AccAddressFromBech32(accAddrStr)
		if err != nil {
			return nil, berpctypes.ErrBadAddress
		}
		addressInfo["evm"] = common.BytesToAddress(accAddr.Bytes()).Hex()
	}

	res["address"] = addressInfo

	_, isSmartContract := res["contract"] // "contract" is a key in the response if the account is a smart contract, returned by EVM interceptor

	// get account balance

	balancesInfo, err := m.GetAccountBalances(accAddrStr, nil)
	if err != nil {
		return nil, err
	}

	res["balances"] = balancesInfo

	// get account transaction count

	if !isSmartContract {
		resAccount, err := m.queryClient.AuthQueryClient.Account(m.ctx, &authtypes.QueryAccountRequest{
			Address: accAddrStr,
		})

		if err == nil && resAccount != nil && resAccount.Account != nil {
			fakeBaseAccount := &berpctypes.FakeBaseAccount{}
			extractedSuccess, err := fakeBaseAccount.TryUnmarshalFromProto(resAccount.Account, m.clientCtx.Codec)
			if err == nil && extractedSuccess {
				res["txsCount"] = fakeBaseAccount.Sequence + 1
			} else if err != nil {
				m.GetLogger().Error("failed to extract base account", "error", err)
			}
		}
	}

	// get staking information

	if !isSmartContract {
		stakingInfo, err := m.GetStakingInfo(accAddrStr)
		if err != nil {
			return nil, err
		}
		res["staking"] = stakingInfo
	}

	return res, nil
}

func (m *Backend) GetValidatorAccount(consOrValAddr string) (berpctypes.GenericBackendResponse, error) {
	consOrValAddr = berpcutils.NormalizeAddress(consOrValAddr)
	if !m.bech32Cfg.IsValAddr(consOrValAddr) && !m.bech32Cfg.IsConsAddr(consOrValAddr) {
		return nil, berpctypes.ErrBadAddress
	}

	res := make(berpctypes.GenericBackendResponse)

	if m.interceptor != nil {
		var intercepted bool
		var err error
		intercepted, _, res, err = m.interceptor.GetAccount(consOrValAddr)
		if err != nil {
			return nil, err
		}
		if intercepted {
			return res, nil
		}
		res = res.ReInitializeIfNil()
	}

	valAddr, consAddr, found, err := m.validatorsConsAddrToValAddr.GetValAddrAndConsAddr(consOrValAddr)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, status.Error(codes.NotFound, "validator could not be found")
	}

	res["address"] = berpctypes.GenericBackendResponse{
		"validatorAddress": valAddr,
		"consensusAddress": consAddr,
	}

	resValInfo, err := m.queryClient.StakingQueryClient.Validator(m.ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: valAddr,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get validator info").Error())
	}

	validatorInfo := berpctypes.GenericBackendResponse{
		"jailed":          resValInfo.Validator.Jailed,
		"status":          resValInfo.Validator.Status.String(),
		"tokens":          resValInfo.Validator.Tokens.String(),
		"delegatorShares": resValInfo.Validator.DelegatorShares.String(),
		"description": berpctypes.GenericBackendResponse{
			"moniker":          resValInfo.Validator.Description.Moniker,
			"identity":         resValInfo.Validator.Description.Identity,
			"website":          resValInfo.Validator.Description.Website,
			"security_contact": resValInfo.Validator.Description.SecurityContact,
			"details":          resValInfo.Validator.Description.Details,
		},
		"unbondingHeight":   resValInfo.Validator.UnbondingHeight,
		"unbondingTime":     resValInfo.Validator.UnbondingTime,
		"commission":        resValInfo.Validator.Commission,
		"minSelfDelegation": resValInfo.Validator.MinSelfDelegation.String(),
	}

	res["validator"] = validatorInfo

	consensusPubKeyMap, err := berpcutils.FromAnyToJsonMap(resValInfo.Validator.ConsensusPubkey, m.clientCtx.Codec)
	if err == nil {
		validatorInfo["consensusPubkey"] = consensusPubKeyMap
	}

	tmVals, err := m.tendermintValidatorsCache.GetValidators()
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to get validators").Error())
	}
	for _, val := range tmVals {
		if sdk.ConsAddress(val.Address).String() == consAddr {
			validatorInfo["votingPower"] = val.VotingPower
			break
		}
	}

	stakingInfo, err := m.GetStakingInfo(valAddr)
	if err != nil {
		return nil, err
	}
	res["staking"] = stakingInfo

	return res, nil
}

func (m *Backend) isAccAddrOr0x(addr string) bool {
	addr = berpcutils.NormalizeAddress(addr)

	if strings.HasPrefix(addr, "0x") {
		return true
	}

	if m.bech32Cfg.IsAccountAddr(addr) {
		return true
	}

	return false
}
