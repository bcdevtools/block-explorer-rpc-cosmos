package be_rpc

import (
	"fmt"
	"github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/namespaces/be"
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	berpcutils "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/utils"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/backend"
	rpcclient "github.com/tendermint/tendermint/rpc/jsonrpc/client"
)

// RPC namespaces and API version
const (
	DymRollAppBlockExplorerNamespace = "be"

	ApiVersion = "1.0"
)

// APICreator creates the JSON-RPC API implementations.
type APICreator = func(
	ctx *server.Context,
	clientCtx client.Context,
	tendermintWebsocketClient *rpcclient.WSClient,
	messageParsers map[string]berpctypes.MessageParser,
	messageInvolversExtractors map[string]berpctypes.MessageInvolversExtractor,
	externalServices berpctypes.ExternalServices,
) []rpc.API

// apiCreators defines the JSON-RPC API namespaces.
var apiCreators map[string]APICreator

// messageParsers defines the message parsers.
var messageParsers map[string]berpctypes.MessageParser

// messageInvolversExtractors defines the message involvers extractors.
var messageInvolversExtractors map[string]berpctypes.MessageInvolversExtractor

func init() {
	apiCreators = map[string]APICreator{
		DymRollAppBlockExplorerNamespace: func(ctx *server.Context,
			clientCtx client.Context,
			tmWSClient *rpcclient.WSClient,
			messageParsers map[string]berpctypes.MessageParser,
			messageInvolversExtractors map[string]berpctypes.MessageInvolversExtractor,
			externalServices berpctypes.ExternalServices,
		) []rpc.API {
			backend := backend.NewBackend(ctx, ctx.Logger, clientCtx, messageParsers, messageInvolversExtractors, externalServices)
			return []rpc.API{
				{
					Namespace: DymRollAppBlockExplorerNamespace,
					Version:   ApiVersion,
					Service:   be.NewBeAPI(ctx, backend),
					Public:    true,
				},
			}
		},
	}

	messageParsers = make(map[string]berpctypes.MessageParser)
}

// GetBeRpcAPIs returns the list of all BE-Json-APIs
func GetBeRpcAPIs(ctx *server.Context,
	clientCtx client.Context,
	tendermintWebsocketClient *rpcclient.WSClient,
	externalServices berpctypes.ExternalServices,
) []rpc.API {
	var apis []rpc.API

	for _, creator := range apiCreators {
		apis = append(apis, creator(
			ctx,
			clientCtx,
			tendermintWebsocketClient,
			messageParsers,
			messageInvolversExtractors,
			externalServices,
		)...)
	}

	return apis
}

// RegisterAPINamespace registers a new API namespace with the API creator.
// This function fails if the namespace is already registered.
// Legacy TODO BE: call to this function to register before startup
func RegisterAPINamespace(ns string, creator APICreator, allowOverride bool) error {
	if !allowOverride {
		if _, ok := apiCreators[ns]; ok {
			panic(fmt.Sprintf("duplicated api namespace %s", ns))
		}
	}

	apiCreators[ns] = creator
	return nil
}

// RegisterMessageParser registers a new parser for the given message type.
// This overrides any existing parser for the given message type.
// Contract: the parser must be registered before the server starts.
// Legacy TODO BE: call to this function to register before startup
func RegisterMessageParser(m sdk.Msg, parser berpctypes.MessageParser) {
	messageParsers[berpcutils.ProtoMessageName(m)] = parser
}

// RegisterMessageInvolversExtractor registers a new involvers extractor for the given message type.
// This overrides any existing parser for the given message type.
// Contract: the parser must be registered before the server starts.
// Legacy TODO BE: call to this function to register before startup
func RegisterMessageInvolversExtractor(m sdk.Msg, extractor berpctypes.MessageInvolversExtractor) {
	messageInvolversExtractors[berpcutils.ProtoMessageName(m)] = extractor
}
