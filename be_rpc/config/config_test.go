package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultBeJsonRpcConfig(t *testing.T) {
	cfg := DefaultBeJsonRpcConfig()
	require.True(t, cfg.Enable)
	require.Equal(t, cfg.Address, DefaultJSONRPCAddress)
	require.Equal(t, cfg.WsAddress, DefaultJSONRPCWsAddress)
}
