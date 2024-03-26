package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

type Bech32Config struct {
	prefixAccountAddr string
	prefixValAddr     string
	prefixConsAddr    string
}

func NewBech32Config() Bech32Config {
	sdkCfg := sdk.GetConfig()
	return Bech32Config{
		prefixAccountAddr: sdkCfg.GetBech32AccountAddrPrefix(),
		prefixValAddr:     sdkCfg.GetBech32ValidatorAddrPrefix(),
		prefixConsAddr:    sdkCfg.GetBech32ConsensusAddrPrefix(),
	}
}

func (m Bech32Config) GetBech32AccountAddrPrefix() string {
	return m.prefixAccountAddr
}

func (m Bech32Config) GetBech32ValidatorAddrPrefix() string {
	return m.prefixValAddr
}

func (m Bech32Config) GetBech32ConsensusAddrPrefix() string {
	return m.prefixConsAddr
}

func (m Bech32Config) IsAccountAddr(addr string) bool {
	return strings.HasPrefix(addr, m.prefixAccountAddr+"1")
}

func (m Bech32Config) IsValAddr(addr string) bool {
	return strings.HasPrefix(addr, m.prefixValAddr+"1")
}

func (m Bech32Config) IsConsAddr(addr string) bool {
	return strings.HasPrefix(addr, m.prefixConsAddr+"1")
}

func (m Bech32Config) FromAnyToBech32AccountAddrUnsafe(addr string) string {
	if strings.HasPrefix(addr, "0x") {
		return sdk.AccAddress(common.HexToAddress(addr).Bytes()).String()
	}

	if m.IsAccountAddr(addr) {
		return addr
	}

	if m.IsValAddr(addr) {
		accAddr, err := sdk.GetFromBech32(addr, m.prefixValAddr)
		if err != nil {
			// ignore
			return addr
		}
		return sdk.AccAddress(accAddr).String()
	}

	return addr
}

func (m Bech32Config) ConvertToAccAddressIfHexOtherwiseKeepAsIs(addr string) string {
	if strings.HasPrefix(addr, "0x") {
		return sdk.AccAddress(common.HexToAddress(addr).Bytes()).String()
	}

	return addr
}
