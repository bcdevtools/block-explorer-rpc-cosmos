package types

import (
	"github.com/stretchr/testify/require"
	"slices"
	"strings"
	"testing"
)

func Test_messageInvolversResult_Merge(t *testing.T) {
	t.Run("merge both", func(t *testing.T) {
		one := NewMessageInvolversResult()
		one.AddGenericInvolvers(MessageInvolvers, "a1", "a2")
		one.AddContractInvolvers(Erc20Involvers, "c1", "ca1", "ca2")

		other := NewMessageInvolversResult()
		other.AddGenericInvolvers(MessageInvolvers, "a1", "b1", "b2")
		other.AddGenericInvolvers(Erc20Involvers, "a1", "b2", "b3", "b1", "a2")
		other.AddContractInvolvers(Erc20Involvers, "c1", "ca1", "cb1", "cb2")
		other.AddContractInvolvers(Erc20Involvers, "c2", "ca1", "ca1", "cb2")
		other.AddContractInvolvers(NftInvolvers, "c3", "ca2", "ca1", "cb3")

		one.Merge(other)

		require.Equal(
			t,
			MessageGenericInvolvers{
				MessageInvolvers: {"a1", "a2", "a1", "b1", "b2"},
				Erc20Involvers:   {"a1", "b2", "b3", "b1", "a2"},
			},
			one.(*messageInvolversResult).genericInvolvers,
		)

		require.Equal(
			t,
			MessageContractsInvolvers{
				Erc20Involvers: {
					"c1": {"ca1", "ca2", "ca1", "cb1", "cb2"},
					"c2": {"ca1", "ca1", "cb2"},
				},
				NftInvolvers: {
					"c3": {"ca2", "ca1", "cb3"},
				},
			},
			one.(*messageInvolversResult).contractInvolvers,
		)
	})

	t.Run("merge with empty", func(t *testing.T) {
		one := NewMessageInvolversResult()
		one.AddGenericInvolvers(MessageInvolvers, "a1", "b1", "b2")
		one.AddGenericInvolvers(Erc20Involvers, "a1", "b2", "b3", "b1", "a2")
		one.AddContractInvolvers(Erc20Involvers, "c1", "ca1", "cb1", "cb2")
		one.AddContractInvolvers(Erc20Involvers, "c2", "ca1", "cb3", "cb2")
		one.AddContractInvolvers(NftInvolvers, "c3", "ca2", "ca1", "cb3")

		other := NewMessageInvolversResult()

		one.Merge(other)

		require.Equal(
			t,
			MessageGenericInvolvers{
				MessageInvolvers: {"a1", "b1", "b2"},
				Erc20Involvers:   {"a1", "b2", "b3", "b1", "a2"},
			},
			one.(*messageInvolversResult).genericInvolvers,
		)

		require.Equal(
			t,
			MessageContractsInvolvers{
				Erc20Involvers: {
					"c1": {"ca1", "cb1", "cb2"},
					"c2": {"ca1", "cb3", "cb2"},
				},
				NftInvolvers: {
					"c3": {"ca2", "ca1", "cb3"},
				},
			},
			one.(*messageInvolversResult).contractInvolvers,
		)
	})

	t.Run("empty merge with non empty", func(t *testing.T) {
		one := NewMessageInvolversResult()

		other := NewMessageInvolversResult()
		other.AddGenericInvolvers(MessageInvolvers, "a1", "b1", "b2")
		other.AddGenericInvolvers(Erc20Involvers, "a1", "b2", "b3", "b1", "a2")
		other.AddContractInvolvers(Erc20Involvers, "c1", "ca1", "cb1", "cb2")
		other.AddContractInvolvers(Erc20Involvers, "c2", "ca1", "cb3", "cb2")
		other.AddContractInvolvers(NftInvolvers, "c3", "ca2", "ca1", "cb3")

		one.Merge(other)

		require.Equal(
			t,
			MessageGenericInvolvers{
				MessageInvolvers: {"a1", "b1", "b2"},
				Erc20Involvers:   {"a1", "b2", "b3", "b1", "a2"},
			},
			one.(*messageInvolversResult).genericInvolvers,
		)

		require.Equal(
			t,
			MessageContractsInvolvers{
				Erc20Involvers: {
					"c1": {"ca1", "cb1", "cb2"},
					"c2": {"ca1", "cb3", "cb2"},
				},
				NftInvolvers: {
					"c3": {"ca2", "ca1", "cb3"},
				},
			},
			one.(*messageInvolversResult).contractInvolvers,
		)
	})

	t.Run("empty merge with empty", func(t *testing.T) {
		one := NewMessageInvolversResult()

		other := NewMessageInvolversResult()

		one.Merge(other)

		require.Equal(
			t,
			MessageGenericInvolvers{},
			one.(*messageInvolversResult).genericInvolvers,
		)

		require.Equal(
			t,
			MessageContractsInvolvers{},
			one.(*messageInvolversResult).contractInvolvers,
		)
	})
}

func Test_messageInvolversResult_Finalize(t *testing.T) {
	one := NewMessageInvolversResult()
	one.AddGenericInvolvers(MessageInvolvers, "a1", "a2", "a4", "a2")
	one.AddContractInvolvers(Erc20Involvers, "c1", "ca1", "ca2", "ca1")

	one.Finalize()

	slices.SortFunc(one.(*messageInvolversResult).genericInvolvers[MessageInvolvers], func(l, r string) int {
		return strings.Compare(l, r)
	})

	slices.SortFunc(one.(*messageInvolversResult).contractInvolvers[Erc20Involvers]["c1"], func(l, r string) int {
		return strings.Compare(l, r)
	})

	require.Equal(
		t,
		MessageGenericInvolvers{
			MessageInvolvers: {"a1", "a2", "a4"},
		},
		one.(*messageInvolversResult).genericInvolvers,
	)

	require.Equal(
		t,
		MessageContractsInvolvers{
			Erc20Involvers: {
				"c1": {"ca1", "ca2"},
			},
		},
		one.(*messageInvolversResult).contractInvolvers,
	)
}

func Test_messageInvolversResult_AddGenericInvolvers_AddContractInvolvers(t *testing.T) {
	one := NewMessageInvolversResult()
	one.AddGenericInvolvers(MessageInvolvers, "A1", "A2", "A4", "A2")
	one.AddContractInvolvers(Erc20Involvers, "C1", "cA1", "cA2", "cA1")

	require.Equal(
		t,
		MessageGenericInvolvers{
			MessageInvolvers: {"a1", "a2", "a4", "a2"},
		},
		one.(*messageInvolversResult).genericInvolvers,
		"all addresses must be added and lower-cased",
	)

	require.Equal(
		t,
		MessageContractsInvolvers{
			Erc20Involvers: {
				"c1": {"ca1", "ca2", "ca1"},
			},
		},
		one.(*messageInvolversResult).contractInvolvers,
		"all addresses must be added and lower-cased, include contract address",
	)
}
