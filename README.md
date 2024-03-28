# block-explorer-rpc-cosmos
Block Explorer RPC for Cosmos chains, as a module.

### Setup
The following methods must be called:
```go
config.AddBeJsonRpcFlags(rootCmd)
```
```go
server.StartBeJsonRPC(...)
```