# block-explorer-rpc-cosmos
Block Explorer RPC for Cosmos chains, as a module.

### Setup
The following methods must be called:
```go
config.EnsureRoot(home, config.DefaultBeJsonRpcConfig())
// in root.go
```
```go
config.AddBeJsonRpcFlags(rootCmd)
// in start.go
```
```go
server.StartBeJsonRPC(...)
// in start.go
```

### Start
```bash
simd start --be.enable true
# Port will be opened at 11100
```

#### Optional configurations
_(the following values are default values)_
```bash
simd start \
    --be.enable true \
    --be.address 0.0.0.0:11100 \
    --be.http-timeout 30s \
    --be.http-idle-timeout 120s \
    --be.max-open-connections 0 \
    --be.allow-cors true
# Configuration file is located at ~/$NODE_HOME/config/be-json-rpc.toml
```
