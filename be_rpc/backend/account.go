package backend

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	berpcutils "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/utils"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (m *Backend) GetAccountBalances(accountAddressStr string, denom *string) (berpctypes.GenericBackendResponse, error) {
	unsafeAccAddr := berpcutils.FromAnyToBech32AddressUnsafe(accountAddressStr)

	if denom == nil || len(*denom) == 0 {
		resAllBalances, err := m.queryClient.BankQueryClient.AllBalances(m.ctx, &banktypes.QueryAllBalancesRequest{
			Address: unsafeAccAddr,
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
		Address: unsafeAccAddr,
		Denom:   *denom,
	})
	if err != nil {
		return nil, err
	}

	res := make(berpctypes.GenericBackendResponse)
	res[resBalance.Balance.Denom] = resBalance.Balance.Amount.String()
	return res, nil
}
