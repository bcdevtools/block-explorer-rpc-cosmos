package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFakeBaseAccount_TryUnmarshalFromProto(t *testing.T) {
	testEncodingConfig := simapp.MakeTestEncodingConfig()

	fakeBaseAccount := &FakeBaseAccount{}

	t.Run("base account", func(t *testing.T) {
		baseAccount := &authtypes.BaseAccount{
			Address:       "a",
			AccountNumber: 1,
			Sequence:      2,
		}

		any, err := codectypes.NewAnyWithValue(baseAccount)
		require.NoError(t, err)

		success, err := fakeBaseAccount.TryUnmarshalFromProto(any, testEncodingConfig.Codec)
		require.NoError(t, err)
		require.True(t, success)

		require.Equal(t, "a", fakeBaseAccount.Address)
		require.Equal(t, uint64(1), fakeBaseAccount.AccountNumber)
		require.Equal(t, uint64(2), fakeBaseAccount.Sequence)
	})

	t.Run("account contains base account", func(t *testing.T) {
		moduleAccount := &authtypes.ModuleAccount{
			BaseAccount: &authtypes.BaseAccount{
				Address:       "a",
				AccountNumber: 1,
				Sequence:      2,
			},
		}

		any, err := codectypes.NewAnyWithValue(moduleAccount)
		require.NoError(t, err)

		success, err := fakeBaseAccount.TryUnmarshalFromProto(any, testEncodingConfig.Codec)
		require.NoError(t, err)
		require.True(t, success)

		require.Equal(t, "a", fakeBaseAccount.Address)
		require.Equal(t, uint64(1), fakeBaseAccount.AccountNumber)
		require.Equal(t, uint64(2), fakeBaseAccount.Sequence)
	})

	t.Run("continous vesting account", func(t *testing.T) {
		continousVestingAccount := &vestingtypes.ContinuousVestingAccount{
			BaseVestingAccount: &vestingtypes.BaseVestingAccount{
				BaseAccount: &authtypes.BaseAccount{
					Address:       "b",
					AccountNumber: 1,
					Sequence:      2,
				},
				OriginalVesting:  nil,
				DelegatedFree:    nil,
				DelegatedVesting: nil,
				EndTime:          0,
			},
			StartTime: 0,
		}

		any, err := codectypes.NewAnyWithValue(continousVestingAccount)
		require.NoError(t, err)

		success, err := fakeBaseAccount.TryUnmarshalFromProto(any, testEncodingConfig.Codec)
		require.NoError(t, err)
		require.True(t, success)

		require.Equal(t, "b", fakeBaseAccount.Address)
		require.Equal(t, uint64(1), fakeBaseAccount.AccountNumber)
		require.Equal(t, uint64(2), fakeBaseAccount.Sequence)
	})

	t.Run("delayed vesting account", func(t *testing.T) {
		delayedVestingAccount := &vestingtypes.DelayedVestingAccount{
			BaseVestingAccount: &vestingtypes.BaseVestingAccount{
				BaseAccount: &authtypes.BaseAccount{
					Address:       "b",
					AccountNumber: 1,
					Sequence:      2,
				},
				OriginalVesting:  nil,
				DelegatedFree:    nil,
				DelegatedVesting: nil,
				EndTime:          0,
			},
		}

		any, err := codectypes.NewAnyWithValue(delayedVestingAccount)
		require.NoError(t, err)

		success, err := fakeBaseAccount.TryUnmarshalFromProto(any, testEncodingConfig.Codec)
		require.NoError(t, err)
		require.True(t, success)

		require.Equal(t, "b", fakeBaseAccount.Address)
		require.Equal(t, uint64(1), fakeBaseAccount.AccountNumber)
		require.Equal(t, uint64(2), fakeBaseAccount.Sequence)
	})

	t.Run("periodic vesting account", func(t *testing.T) {
		periodicVestingAccount := &vestingtypes.PeriodicVestingAccount{
			BaseVestingAccount: &vestingtypes.BaseVestingAccount{
				BaseAccount: &authtypes.BaseAccount{
					Address:       "b",
					AccountNumber: 1,
					Sequence:      2,
				},
				OriginalVesting:  nil,
				DelegatedFree:    nil,
				DelegatedVesting: nil,
				EndTime:          0,
			},
		}

		any, err := codectypes.NewAnyWithValue(periodicVestingAccount)
		require.NoError(t, err)

		success, err := fakeBaseAccount.TryUnmarshalFromProto(any, testEncodingConfig.Codec)
		require.NoError(t, err)
		require.True(t, success)

		require.Equal(t, "b", fakeBaseAccount.Address)
		require.Equal(t, uint64(1), fakeBaseAccount.AccountNumber)
		require.Equal(t, uint64(2), fakeBaseAccount.Sequence)
	})

	t.Run("permanent locked vesting account", func(t *testing.T) {
		permanentLockedAccount := &vestingtypes.PermanentLockedAccount{
			BaseVestingAccount: &vestingtypes.BaseVestingAccount{
				BaseAccount: &authtypes.BaseAccount{
					Address:       "b",
					AccountNumber: 1,
					Sequence:      2,
				},
				OriginalVesting:  nil,
				DelegatedFree:    nil,
				DelegatedVesting: nil,
				EndTime:          0,
			},
		}

		any, err := codectypes.NewAnyWithValue(permanentLockedAccount)
		require.NoError(t, err)

		success, err := fakeBaseAccount.TryUnmarshalFromProto(any, testEncodingConfig.Codec)
		require.NoError(t, err)
		require.True(t, success)

		require.Equal(t, "b", fakeBaseAccount.Address)
		require.Equal(t, uint64(1), fakeBaseAccount.AccountNumber)
		require.Equal(t, uint64(2), fakeBaseAccount.Sequence)
	})
}
