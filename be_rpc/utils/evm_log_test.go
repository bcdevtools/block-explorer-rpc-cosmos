package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsEvmEventMatch(t *testing.T) {
	var topic0Transfer = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	var topic0Deposit = common.HexToHash("0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c")

	t.Run("number of topic must match", func(t *testing.T) {
		require.Equal(t,
			false,
			IsEvmEventMatch(
				[]common.Hash{}, []byte{},
				1, common.Hash{},
				false, false, false,
				false,
			),
		)

		require.Equal(t,
			false,
			IsEvmEventMatch(
				[]common.Hash{}, []byte{},
				2, common.Hash{},
				false, false, false,
				false,
			),
		)

		require.Equal(t,
			false,
			IsEvmEventMatch(
				[]common.Hash{
					{},
					{},
				}, []byte{},
				1, common.Hash{},
				false, false, false,
				false,
			),
		)

		require.Equal(t,
			true,
			IsEvmEventMatch(
				[]common.Hash{
					{},
					{},
				}, []byte{},
				2, common.Hash{},
				true, false, false,
				false,
			),
		)
	})

	t.Run("want data must match", func(t *testing.T) {
		require.Equal(t,
			true,
			IsEvmEventMatch(
				[]common.Hash{
					{},
					{},
				}, []byte{},
				2, common.Hash{},
				true, false, false,
				false,
			),
		)

		require.Equal(t,
			true,
			IsEvmEventMatch(
				[]common.Hash{
					{},
					{},
				}, []byte{1},
				2, common.Hash{},
				true, false, false,
				true,
			),
		)

		require.Equal(t,
			false,
			IsEvmEventMatch(
				[]common.Hash{
					{},
					{},
				}, []byte{1},
				2, common.Hash{},
				true, false, false,
				false,
			),
		)

		require.Equal(t,
			false,
			IsEvmEventMatch(
				[]common.Hash{
					{},
					{},
				}, []byte{},
				2, common.Hash{},
				true, false, false,
				true,
			),
		)
	})

	t.Run("topic0 must match", func(t *testing.T) {
		require.Equal(t,
			true,
			IsEvmEventMatch(
				[]common.Hash{
					topic0Transfer,
					{},
					{},
				}, []byte{1},
				3, topic0Transfer,
				true, true, false,
				true,
			),
		)
		require.Equal(t,
			false,
			IsEvmEventMatch(
				[]common.Hash{
					topic0Transfer,
					{},
					{},
				}, []byte{1},
				3, topic0Deposit,
				true, true, false,
				true,
			),
		)
	})

	addrTopic := common.HexToHash("0x0000000000000000000000007af33266ef0f967b01376f613387fc7c88699967")
	nonAddrTopic := common.HexToHash("0x0000000000000000000000017af33266ef0f967b01376f613387fc7c88699967")

	t.Run("want address at topic 1 must match", func(t *testing.T) {
		require.Equal(t,
			true,
			IsEvmEventMatch(
				[]common.Hash{
					topic0Transfer,
					addrTopic,
					addrTopic,
				}, []byte{1},
				3, topic0Transfer,
				true, true, false,
				true,
			),
		)
		require.Equal(t,
			false,
			IsEvmEventMatch(
				[]common.Hash{
					topic0Transfer,
					nonAddrTopic,
					addrTopic,
				}, []byte{1},
				3, topic0Transfer,
				true, true, false,
				true,
			),
		)
	})

	t.Run("want address at topic 2 must match", func(t *testing.T) {
		require.Equal(t,
			true,
			IsEvmEventMatch(
				[]common.Hash{
					topic0Transfer,
					addrTopic,
					addrTopic,
				}, []byte{1},
				3, topic0Transfer,
				true, true, false,
				true,
			),
		)
		require.Equal(t,
			false,
			IsEvmEventMatch(
				[]common.Hash{
					topic0Transfer,
					addrTopic,
					nonAddrTopic,
				}, []byte{1},
				3, topic0Transfer,
				true, true, false,
				true,
			),
		)
	})

	t.Run("want address at topic 3 must match", func(t *testing.T) {
		require.Equal(t,
			true,
			IsEvmEventMatch(
				[]common.Hash{
					topic0Transfer,
					addrTopic,
					addrTopic,
					addrTopic,
				}, []byte{1},
				4, topic0Transfer,
				true, true, true,
				true,
			),
		)
		require.Equal(t,
			false,
			IsEvmEventMatch(
				[]common.Hash{
					topic0Transfer,
					addrTopic,
					addrTopic,
					nonAddrTopic,
				}, []byte{1},
				4, topic0Transfer,
				true, true, true,
				true,
			),
		)
	})

	t.Run("when topic is address but not specific must be address, it is ok", func(t *testing.T) {
		require.Equal(t,
			true,
			IsEvmEventMatch(
				[]common.Hash{
					topic0Transfer,
					addrTopic,
					addrTopic,
					addrTopic,
				}, []byte{1},
				4, topic0Transfer,
				false, false, false,
				true,
			),
		)
	})
}

func TestAccAddressFromTopic(t *testing.T) {
	addrTopic := common.HexToHash("0x0000000000000000000000007af33266ef0f967b01376f613387fc7c88699967")
	addr := common.HexToAddress("0x7af33266ef0f967b01376f613387fc7c88699967")
	require.Equal(t,
		sdk.AccAddress(addr.Bytes()),
		AccAddressFromTopic(addrTopic),
	)
}
