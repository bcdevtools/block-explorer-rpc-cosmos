package backend

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	berpcutils "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
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

	balancesInfo, err := m.GetAccountBalances(accAddrStr, nil)
	if err != nil {
		return nil, err
	}

	res["address"] = addressInfo
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
	}

	return res, nil
}
