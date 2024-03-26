package utils

import (
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"
)

func TestIsEventTypeWithAllAttributes(t *testing.T) {
	var ok bool
	var kv map[string]string

	ok, kv = IsEventTypeWithAllAttributes(abci.Event{
		Type: "test",
		Attributes: []abci.EventAttribute{
			{Key: []byte("key1"), Value: []byte("value1")},
			{Key: []byte("key2"), Value: []byte("value2")},
		},
	}, "test", "key1", "key2")
	require.True(t, ok)
	require.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, kv)

	ok, kv = IsEventTypeWithAllAttributes(abci.Event{
		Type: "test",
		Attributes: []abci.EventAttribute{
			{Key: []byte("key1"), Value: []byte("value1")},
			{Key: []byte("key2"), Value: []byte("value2")},
		},
	}, "test", "key1")
	require.True(t, ok)
	require.Equal(t, map[string]string{"key1": "value1"}, kv)

	ok, kv = IsEventTypeWithAllAttributes(abci.Event{
		Type: "test",
		Attributes: []abci.EventAttribute{
			{Key: []byte("key1"), Value: []byte("value1")},
			{Key: []byte("key2"), Value: []byte("value2")},
		},
	}, "test")
	require.True(t, ok)
	require.Equal(t, map[string]string{}, kv)

	ok, _ = IsEventTypeWithAllAttributes(abci.Event{
		Type: "test",
		Attributes: []abci.EventAttribute{
			{Key: []byte("key1"), Value: []byte("value1")},
			{Key: []byte("key2"), Value: []byte("value2")},
		},
	}, "test", "key1", "key3")
	require.False(t, ok)

	ok, _ = IsEventTypeWithAllAttributes(abci.Event{
		Type: "test",
		Attributes: []abci.EventAttribute{
			{Key: []byte("key1"), Value: []byte("value1")},
			{Key: []byte("key2"), Value: []byte("value2")},
		},
	}, "test", "key3", "key4")
	require.False(t, ok)
}
