package config

import (
	"errors"
	"github.com/spf13/cobra"
	"time"

	"github.com/spf13/viper"
)

// BeJsonRpcConfig defines configuration for the BE-RPC server.
type BeJsonRpcConfig struct {
	// Address defines the HTTP server to listen on
	Address string `mapstructure:"address"`
	// WsAddress defines the WebSocket server to listen on
	WsAddress string `mapstructure:"ws-address"`
	// Enable defines if the Be Json RPC server should be enabled.
	Enable bool `mapstructure:"enable"`
	// HTTPTimeout is the read/write timeout of http json-rpc server.
	HTTPTimeout time.Duration `mapstructure:"http-timeout"`
	// HTTPIdleTimeout is the idle timeout of http json-rpc server.
	HTTPIdleTimeout time.Duration `mapstructure:"http-idle-timeout"`
	// MaxOpenConnections sets the maximum number of simultaneous connections
	// for the server listener.
	MaxOpenConnections int `mapstructure:"max-open-connections"`
	// EnableUnsafeCORS defines if the server should allow unsafe CORS requests.
	EnableUnsafeCORS bool `mapstructure:"-"`
}

// DefaultBeJsonRpcConfig returns Block Explorer JSON-RPC API config enabled by default
func DefaultBeJsonRpcConfig() *BeJsonRpcConfig {
	return &BeJsonRpcConfig{
		Enable:             true,
		Address:            DefaultJSONRPCAddress,
		WsAddress:          DefaultJSONRPCWsAddress,
		HTTPTimeout:        DefaultHTTPTimeout,
		HTTPIdleTimeout:    DefaultHTTPIdleTimeout,
		MaxOpenConnections: DefaultMaxOpenConnections,
		EnableUnsafeCORS:   false,
	}
}

// Validate returns an error if the JSON-RPC configuration fields are invalid.
func (c BeJsonRpcConfig) Validate() error {
	if c.HTTPTimeout < 0 {
		return errors.New("BE-JSON-RPC HTTP timeout duration cannot be negative")
	}

	if c.HTTPIdleTimeout < 0 {
		return errors.New("BE-JSON-RPC HTTP idle timeout duration cannot be negative")
	}

	return nil
}

// GetConfig returns a fully parsed BeJsonRpcConfig object.
func GetConfig(v *viper.Viper) (BeJsonRpcConfig, error) {
	cfg := BeJsonRpcConfig{
		Enable:             v.GetBool(FlagBeJsonRpcEnable),
		Address:            v.GetString(FlagBeJsonRpcAddress),
		WsAddress:          v.GetString(FlagBeJsonRpcWsAddress),
		HTTPTimeout:        v.GetDuration(FlagBeJsonRpcHttpTimeout),
		HTTPIdleTimeout:    v.GetDuration(FlagBeJsonRpcHttpIdleTimeout),
		MaxOpenConnections: v.GetInt(FlagBeJsonRpcHttpTimeout),
		EnableUnsafeCORS:   false,
	}

	return cfg, cfg.Validate()
}

// AddBeJsonRpcFlags add Block Explorer Json-RPC flags into the cmd
// Legacy TODO BE: call this to register flags
func AddBeJsonRpcFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagBeJsonRpcEnable, true, "define if the Block Explorer JSON-RPC server should be enabled")
	cmd.Flags().String(FlagBeJsonRpcAddress, DefaultJSONRPCAddress, "the Block Explorer JSON-RPC server address to listen on")
	cmd.Flags().String(FlagBeJsonRpcWsAddress, DefaultJSONRPCWsAddress, "the Block Explorer JSON-RPC WS server address to listen on")
	cmd.Flags().Duration(FlagBeJsonRpcHttpTimeout, DefaultHTTPTimeout, "Sets a read/write timeout for block explorer json-rpc http server (0=infinite)")
	cmd.Flags().Duration(FlagBeJsonRpcHttpIdleTimeout, DefaultHTTPIdleTimeout, "Sets a idle timeout for block explorer json-rpc http server (0=infinite)")
	cmd.Flags().Duration(FlagBeJsonRpcMaxOpenConnection, DefaultMaxOpenConnections, "Maximum open connection for block explorer json-rpc http server")
}
