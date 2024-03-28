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
simd start --be-json-rpc.enable true
# Ports will be opened at 11100
```

#### Optional configurations
_(the following values are default values)_
```bash
simd start \
    --be-json-rpc.enable true \
    --be-json-rpc.address 0.0.0.0:11100 \
    --be-json-rpc.http-timeout 30s \
    --be-json-rpc.http-idle-timeout 120s \
    --be-json-rpc.max-open-connections 0 \
    --be-json-rpc.allow-cors true
```