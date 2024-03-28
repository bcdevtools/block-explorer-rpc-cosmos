package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
)

const (
	FlagBeJsonRpcEnable            = "be.enable"
	FlagBeJsonRpcAddress           = "be.address"
	FlagBeJsonRpcHttpTimeout       = "be.http-timeout"
	FlagBeJsonRpcHttpIdleTimeout   = "be.http-idle-timeout"
	FlagBeJsonRpcMaxOpenConnection = "be.max-open-connections"
	FlagBeJsonRpcAllowCORS         = "be.allow-cors"
)

const (
	// DefaultEnable is the default value for enabling the BE-JSON-RPC server
	DefaultEnable = true

	// DefaultJSONRPCAddress is the default address the BE-JSON-RPC server binds to.
	DefaultJSONRPCAddress = "0.0.0.0:11100"

	// DefaultHTTPTimeout is the default read/write timeout of the http be-json-rpc server
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultHTTPIdleTimeout is the default idle timeout of the http be-json-rpc server
	DefaultHTTPIdleTimeout = 120 * time.Second

	// DefaultMaxOpenConnections represents the amount of open connections (unlimited = 0)
	DefaultMaxOpenConnections = 0

	// DefaultAllowCORS represents the default value for allowing CORS requests
	DefaultAllowCORS = true
)

func bindFlagsToViper(cmd *cobra.Command, v *viper.Viper) error {
	if err := v.BindPFlag("enable", cmd.Flags().Lookup(FlagBeJsonRpcEnable)); err != nil {
		return err
	}
	if err := v.BindPFlag("address", cmd.Flags().Lookup(FlagBeJsonRpcAddress)); err != nil {
		return err
	}
	if err := v.BindPFlag("http-timeout", cmd.Flags().Lookup(FlagBeJsonRpcHttpTimeout)); err != nil {
		return err
	}
	if err := v.BindPFlag("http-idle-timeout", cmd.Flags().Lookup(FlagBeJsonRpcHttpIdleTimeout)); err != nil {
		return err
	}
	if err := v.BindPFlag("max-open-connections", cmd.Flags().Lookup(FlagBeJsonRpcMaxOpenConnection)); err != nil {
		return err
	}
	if err := v.BindPFlag("allow-cors", cmd.Flags().Lookup(FlagBeJsonRpcAllowCORS)); err != nil {
		return err
	}
	return nil
}
