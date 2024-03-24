package utils

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetIncomingIBCCoin(t *testing.T) {
	testcases := []struct {
		name       string
		srcPort    string
		srcChannel string
		dstPort    string
		dstChannel string
		denom      string
		amount     string
		wantCoin   sdk.Coin
	}{
		{
			name:       "transfer unwrapped coin to destination which is not its source",
			srcPort:    "transfer",
			srcChannel: "channel-2",
			dstPort:    "transfer",
			dstChannel: "channel-2",
			denom:      "uosmo",
			amount:     "10",
			wantCoin: sdk.Coin{
				Denom: transfertypes.DenomTrace{
					Path:      "transfer/channel-2",
					BaseDenom: "uosmo",
				}.IBCDenom(),
				Amount: math.NewInt(10),
			},
		},
		{
			name:       "transfer ibc wrapped coin to destination which is its source",
			srcPort:    "transfer",
			srcChannel: "channel-0",
			dstPort:    "transfer",
			dstChannel: "channel-0",
			denom:      "transfer/channel-0/adym",
			amount:     "10",
			wantCoin: sdk.Coin{
				Denom:  "adym",
				Amount: math.NewInt(10),
			},
		},
		{
			name:       "transfer 2x ibc wrapped coin to destination which is its source",
			srcPort:    "transfer",
			srcChannel: "channel-1",
			dstPort:    "transfer",
			dstChannel: "channel-1",
			denom:      "transfer/channel-1/transfer/channel-1/uatom",
			amount:     "10",
			wantCoin: sdk.Coin{
				Denom: transfertypes.DenomTrace{
					Path:      "transfer/channel-1",
					BaseDenom: "uatom",
				}.IBCDenom(),
				Amount: math.NewInt(10),
			},
		},
		{
			name:       "transfer ibc wrapped coin to destination which is not its source",
			srcPort:    "transfer",
			srcChannel: "channel-1",
			dstPort:    "transfer",
			dstChannel: "channel-1",
			denom:      "transfer/channel-2/uatom",
			amount:     "10",
			wantCoin: sdk.Coin{
				Denom: transfertypes.DenomTrace{
					Path:      "transfer/channel-1/transfer/channel-2",
					BaseDenom: "uatom",
				}.IBCDenom(),
				Amount: math.NewInt(10),
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			coin, err := GetIncomingIBCCoin(tt.srcPort, tt.srcChannel, tt.dstPort, tt.dstChannel, tt.denom, tt.amount)
			require.NoError(t, err)
			require.Equal(t, tt.wantCoin, coin)
		})
	}
}
