package utils

import (
	"cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"strings"
)

func CoinsToMap(coins ...sdk.Coin) map[string]string {
	m := make(map[string]string)
	for _, coin := range coins {
		m[coin.Denom] = coin.Amount.String()
	}
	return m
}

func GetIncomingIBCCoin(srcPort, srcChannel string, dstPort, dstChannel string, denom, amt string) (sdk.Coin, error) {
	amount, ok := math.NewIntFromString(amt)
	if !ok {
		return sdk.Coin{}, fmt.Errorf("invalid amount: %s", amt)
	}

	if transfertypes.ReceiverChainIsSource(srcPort, srcChannel, denom) {
		spl := strings.SplitN(denom, "/", 3)
		denom := spl[2]

		denomTrace := transfertypes.ParseDenomTrace(denom)
		if denomTrace.Path != "" {
			denom = denomTrace.IBCDenom()
		}

		return sdk.Coin{
			Denom:  denom,
			Amount: amount,
		}, nil
	}

	prefixedDenom := fmt.Sprintf("%s/%s/%s", dstPort, dstChannel, denom)
	denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	voucherDenom := denomTrace.IBCDenom()

	return sdk.Coin{
		Denom:  voucherDenom,
		Amount: amount,
	}, nil
}
