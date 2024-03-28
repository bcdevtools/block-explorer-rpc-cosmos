package server

import (
	"github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc"
	"github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/backend"
	berpccfg "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/config"
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	tmlog "github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"golang.org/x/net/netutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/types"
	ethlog "github.com/ethereum/go-ethereum/log"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
)

// StartBeJsonRPC starts the BE-JSON-RPC server.
// Legacy TODO BE: call to this function to start server
func StartBeJsonRPC(ctx *server.Context,
	clientCtx client.Context,
	tmRPCAddr,
	tmEndpoint string,
	config berpccfg.BeJsonRpcConfig,
	requestInterceptorCreator func(backend.BackendI) backend.RequestInterceptor,
	externalServices berpctypes.ExternalServices,
) (*http.Server, chan struct{}, error) {
	tmWsClient := connectTmWS(tmRPCAddr, tmEndpoint, ctx.Logger)

	logger := ctx.Logger.With("module", "geth")
	ethlog.Root().SetHandler(ethlog.FuncHandler(func(r *ethlog.Record) error {
		switch r.Lvl {
		case ethlog.LvlTrace, ethlog.LvlDebug:
			logger.Debug(r.Msg, r.Ctx...)
		case ethlog.LvlInfo, ethlog.LvlWarn:
			logger.Info(r.Msg, r.Ctx...)
		case ethlog.LvlError, ethlog.LvlCrit:
			logger.Error(r.Msg, r.Ctx...)
		}
		return nil
	}))

	rpcServer := ethrpc.NewServer()

	apis := be_rpc.GetBeRpcAPIs(ctx, clientCtx, tmWsClient, requestInterceptorCreator, externalServices)

	for _, api := range apis {
		if err := rpcServer.RegisterName(api.Namespace, api.Service); err != nil {
			ctx.Logger.Error(
				"failed to register service in JSON RPC namespace",
				"namespace", api.Namespace,
				"service", api.Service,
			)
			return nil, nil, err
		}
	}

	r := mux.NewRouter()

	var handlerFunc func(http.ResponseWriter, *http.Request)
	if config.AllowCORS {
		handlerFunc = func(writer http.ResponseWriter, request *http.Request) {
			addCorsHeaders(request.Method, writer)
			rpcServer.ServeHTTP(writer, request)
		}
	} else {
		handlerFunc = rpcServer.ServeHTTP
	}

	r.HandleFunc("/", handlerFunc).Methods("POST")

	handlerWithCors := cors.Default()
	if config.AllowCORS {
		handlerWithCors = cors.AllowAll()
	}

	httpSrv := &http.Server{
		Addr:              config.Address,
		Handler:           handlerWithCors.Handler(r),
		ReadHeaderTimeout: config.HTTPTimeout,
		ReadTimeout:       config.HTTPTimeout,
		WriteTimeout:      config.HTTPTimeout,
		IdleTimeout:       config.HTTPIdleTimeout,
	}
	httpSrvDone := make(chan struct{}, 1)

	ln, err := listen(httpSrv.Addr, config)
	if err != nil {
		return nil, nil, err
	}

	errCh := make(chan error)
	go func() {
		ctx.Logger.Info("Starting BE-JSON-RPC server", "address", config.Address)
		if err := httpSrv.Serve(ln); err != nil {
			if err == http.ErrServerClosed {
				close(httpSrvDone)
				return
			}

			ctx.Logger.Error("failed to start BE-JSON-RPC server", "error", err.Error())
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		ctx.Logger.Error("failed to boot BE-JSON-RPC server", "error", err.Error())
		return nil, nil, err
	case <-time.After(types.ServerStartTime): // assume BE JSON RPC server started successfully
	}

	return httpSrv, httpSrvDone, nil
}

// listen starts a net.Listener on the tcp network on the given address.
// If there is a specified MaxOpenConnections in the config, it will also set the limitListener.
func listen(addr string, config berpccfg.BeJsonRpcConfig) (net.Listener, error) {
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	if config.MaxOpenConnections > 0 {
		ln = netutil.LimitListener(ln, config.MaxOpenConnections)
	}
	return ln, err
}

func connectTmWS(tmRPCAddr, tmEndpoint string, logger tmlog.Logger) *rpcclient.WSClient {
	tmWsClient, err := rpcclient.NewWS(tmRPCAddr, tmEndpoint,
		rpcclient.MaxReconnectAttempts(256),
		rpcclient.ReadWait(120*time.Second),
		rpcclient.WriteWait(120*time.Second),
		rpcclient.PingPeriod(50*time.Second),
		rpcclient.OnReconnect(func() {
			logger.Debug("BE Json RPC reconnects to Tendermint WS", "address", tmRPCAddr+tmEndpoint)
		}),
	)

	if err != nil {
		logger.Error(
			"Tendermint WS client could not be created",
			"address", tmRPCAddr+tmEndpoint,
			"error", err,
		)
	} else if err := tmWsClient.OnStart(); err != nil {
		logger.Error(
			"Tendermint WS client could not start",
			"address", tmRPCAddr+tmEndpoint,
			"error", err,
		)
	}

	return tmWsClient
}

func addCorsHeaders(httpMethod string, responseWriter http.ResponseWriter) {
	httpMethod = strings.ToUpper(httpMethod)

	if httpMethod == http.MethodPost {
		responseWriter.Header().Set("Access-Control-Allow-Origin", "*")
		responseWriter.Header().Set("Access-Control-Allow-Methods", "POST")
		responseWriter.Header().Set("Access-Control-Allow-Headers", "DNT,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Origin,Accept,X-Server-Time")
	}

	return
}
