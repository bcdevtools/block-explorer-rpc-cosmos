package config

import "time"

const (
	FlagBeJsonRpcEnable            = "be-json-rpc.enable"
	FlagBeJsonRpcAddress           = "be-json-rpc.address"
	FlagBeJsonRpcHttpTimeout       = "be-json-rpc.http-timeout"
	FlagBeJsonRpcHttpIdleTimeout   = "be-json-rpc.http-idle-timeout"
	FlagBeJsonRpcMaxOpenConnection = "be-json-rpc.max-open-connections"
	FlagBeJsonRpcAllowCORS         = "be-json-rpc.allow-cors"
)

const (
	// DefaultEnable is the default value for enabling the BE-JSON-RPC server
	DefaultEnable = true

	// DefaultJSONRPCAddress is the default address the BE-JSON-RPC server binds to.
	DefaultJSONRPCAddress = "127.0.0.1:11100"

	// DefaultHTTPTimeout is the default read/write timeout of the http be-json-rpc server
	DefaultHTTPTimeout = 30 * time.Second

	// DefaultHTTPIdleTimeout is the default idle timeout of the http be-json-rpc server
	DefaultHTTPIdleTimeout = 120 * time.Second

	// DefaultMaxOpenConnections represents the amount of open connections (unlimited = 0)
	DefaultMaxOpenConnections = 0

	// DefaultAllowCORS represents the default value for allowing CORS requests
	DefaultAllowCORS = true
)
