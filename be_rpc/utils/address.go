package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

func FromAnyToBech32AddressUnsafe(addr string) string {
	if strings.HasPrefix(addr, "0x") {
		return sdk.AccAddress(common.HexToAddress(addr).Bytes()).String()
	}

	sdkCfg := sdk.GetConfig()

	if strings.HasPrefix(addr, sdkCfg.GetBech32AccountAddrPrefix()+"1") {
		return addr
	}

	if strings.HasPrefix(addr, sdkCfg.GetBech32ValidatorAddrPrefix()+"1") {
		accAddr, err := sdk.GetFromBech32(addr, sdkCfg.GetBech32ValidatorAddrPrefix())
		if err != nil {
			// ignore
			return addr
		}
		return sdk.AccAddress(accAddr).String()
	}

	return addr
}
