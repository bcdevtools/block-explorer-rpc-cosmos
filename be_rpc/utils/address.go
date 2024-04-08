package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

func NormalizeAddress(address string) string {
	return strings.ToLower(strings.TrimSpace(address))
}

// IsZeroAccAddress returns true if the address is 0x00..00
func IsZeroAccAddress(accAddr sdk.AccAddress) bool {
	for _, b := range accAddr.Bytes() {
		if b != 0x00 {
			return false
		}
	}

	return true
}

// IsZeroEvmAddress returns true if the address is 0x00..00
func IsZeroEvmAddress(addr common.Address) bool {
	for _, b := range addr.Bytes() {
		if b != 0x00 {
			return false
		}
	}

	return true
}
