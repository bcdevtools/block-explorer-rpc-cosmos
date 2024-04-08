package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsZeroAddress(t *testing.T) {
	addr1 := common.HexToAddress("0x0000000000000000000000000000000000000000")
	require.True(t, IsZeroEvmAddress(addr1))
	require.True(t, IsZeroAccAddress(addr1.Bytes()))

	addr2 := common.HexToAddress("0x0000000000000000000000000000000000000001")
	require.False(t, IsZeroEvmAddress(addr2))
	require.False(t, IsZeroAccAddress(addr2.Bytes()))
}
