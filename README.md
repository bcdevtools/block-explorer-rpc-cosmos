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
```