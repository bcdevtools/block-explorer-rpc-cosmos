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
	// Enable defines if the Be Json RPC server should be enabled.
	Enable bool `mapstructure:"enable"`
	// HTTPTimeout is the read/write timeout of http json-rpc server.
	HTTPTimeout time.Duration `mapstructure:"http-timeout"`
	// HTTPIdleTimeout is the idle timeout of http json-rpc server.
	HTTPIdleTimeout time.Duration `mapstructure:"http-idle-timeout"`
	// MaxOpenConnections sets the maximum number of simultaneous connections
	// for the server listener.
	MaxOpenConnections int `mapstructure:"max-open-connections"`
	// AllowCORS defines if the server should allow CORS requests. Allowed by default.
	AllowCORS bool `mapstructure:"allow-cors"`
}

// DefaultBeJsonRpcConfig returns Block Explorer JSON-RPC API config with default values
func DefaultBeJsonRpcConfig() *BeJsonRpcConfig {
	return &BeJsonRpcConfig{
		Enable:             DefaultEnable,
		Address:            DefaultJSONRPCAddress,
		HTTPTimeout:        DefaultHTTPTimeout,
		HTTPIdleTimeout:    DefaultHTTPIdleTimeout,
		MaxOpenConnections: DefaultMaxOpenConnections,
		AllowCORS:          DefaultAllowCORS,
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
		HTTPTimeout:        v.GetDuration(FlagBeJsonRpcHttpTimeout),
		HTTPIdleTimeout:    v.GetDuration(FlagBeJsonRpcHttpIdleTimeout),
		MaxOpenConnections: v.GetInt(FlagBeJsonRpcMaxOpenConnection),
		AllowCORS:          v.GetBool(FlagBeJsonRpcAllowCORS),
	}

	return cfg, cfg.Validate()
}

// AddBeJsonRpcFlags add Block Explorer Json-RPC flags into the cmd
// Legacy TODO BE: call this to register flags
func AddBeJsonRpcFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagBeJsonRpcEnable, DefaultEnable, "define if the Block Explorer Json-RPC server should be enabled")
	cmd.Flags().String(FlagBeJsonRpcAddress, DefaultJSONRPCAddress, "define the address for the Block Explorer Json-RPC server to listen on")
	cmd.Flags().Duration(FlagBeJsonRpcHttpTimeout, DefaultHTTPTimeout, "sets a read/write timeout for Block Explorer Json-RPC http server (0 is no timeout)")
	cmd.Flags().Duration(FlagBeJsonRpcHttpIdleTimeout, DefaultHTTPIdleTimeout, "sets an idle timeout for Block Explorer Json-RPC http server (0 is no timeout)")
	cmd.Flags().Duration(FlagBeJsonRpcMaxOpenConnection, DefaultMaxOpenConnections, "sets maximum open connection for Block Explorer Json-RPC http server (0 is unlimited)")
	cmd.Flags().Bool(FlagBeJsonRpcAllowCORS, DefaultAllowCORS, "define if the Block Explorer Json-RPC should allow CORS requests")
}
