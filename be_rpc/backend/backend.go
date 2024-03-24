package backend

import (
	"context"
	"github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/config"
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/tendermint/tendermint/libs/log"
)

type BackendI interface {
	// Chain

	// GetChainInfo returns the chain information.
	GetChainInfo() (berpctypes.GenericBackendResponse, error)

	// GetModuleParams returns the module parameters by module name.
	GetModuleParams(moduleName string) (berpctypes.GenericBackendResponse, error)

	// Account

	GetAccountBalances(accountAddressStr string, denom *string) (berpctypes.GenericBackendResponse, error)

	// TODO BE get account

	// Block

	// GetBlockByNumber returns a block by its height.
	GetBlockByNumber(height int64) (berpctypes.GenericBackendResponse, error)

	// Transactions

	// GetTransactionsInBlockRange returns the list transaction info within a block range.
	// The range is inclusive, specified clearly.
	GetTransactionsInBlockRange(fromHeightIncluded, toHeightIncluded int64) (berpctypes.GenericBackendResponse, error)

	// GetTransactionByHash returns a transaction by its hash.
	GetTransactionByHash(hash string) (berpctypes.GenericBackendResponse, error)

	// TODO get all ERC-20 & NFT contracts and involvers within tx

	// Staking

	// GetStakingInfo returns the staking information, includes:
	// - Delegator's staking information
	// - Validator's commission & outstanding rewards
	GetStakingInfo(delegatorAddr string) (berpctypes.GenericBackendResponse, error)

	// Gov

	GetGovProposals(pageNo int) (berpctypes.GenericBackendResponse, error)

	// Misc

	GetDenomMetadata(base string) (berpctypes.GenericBackendResponse, error)
	GetDenomsMetadata(pageNo int) (berpctypes.GenericBackendResponse, error)
	GetTotalSupply(pageNo int) (berpctypes.GenericBackendResponse, error)

	// Export fields

	GetContext() context.Context
	GetClientContext() client.Context
	GetQueryClient() *berpctypes.QueryClient
	GetLogger() log.Logger
	GetConfig() config.BeJsonRpcConfig
	GetExternalServices() berpctypes.ExternalServices
}

var _ BackendI = (*Backend)(nil)

// Backend implements the BackendI interface
type Backend struct {
	ctx                        context.Context
	clientCtx                  client.Context
	queryClient                *berpctypes.QueryClient // gRPC query client
	logger                     log.Logger
	cfg                        config.BeJsonRpcConfig
	interceptor                RequestInterceptor
	messageParsers             map[string]berpctypes.MessageParser
	messageInvolversExtractors map[string]berpctypes.MessageInvolversExtractor
	externalServices           berpctypes.ExternalServices
}

// NewBackend creates a new Backend instance for RollApp Block Explorer
func NewBackend(
	ctx *server.Context,
	logger log.Logger,
	clientCtx client.Context,
	messageParsers map[string]berpctypes.MessageParser,
	messageInvolversExtractors map[string]berpctypes.MessageInvolversExtractor,
	externalServices berpctypes.ExternalServices,
) *Backend {
	appConf, err := config.GetConfig(ctx.Viper)
	if err != nil {
		panic(err)
	}

	return &Backend{
		ctx:                        context.Background(),
		clientCtx:                  clientCtx,
		queryClient:                berpctypes.NewQueryClient(clientCtx),
		logger:                     logger.With("module", "be_rpc"),
		cfg:                        appConf,
		messageParsers:             messageParsers,
		messageInvolversExtractors: messageInvolversExtractors,
		externalServices:           externalServices,
	}
}

func (m *Backend) WithInterceptor(interceptor RequestInterceptor) *Backend {
	m.interceptor = interceptor
	return m
}

func (m *Backend) GetContext() context.Context {
	return m.ctx
}

func (m *Backend) GetClientContext() client.Context {
	return m.clientCtx
}

func (m *Backend) GetQueryClient() *berpctypes.QueryClient {
	return m.queryClient
}

func (m *Backend) GetLogger() log.Logger {
	return m.logger
}

func (m *Backend) GetConfig() config.BeJsonRpcConfig {
	return m.cfg
}

func (m *Backend) GetExternalServices() berpctypes.ExternalServices {
	return m.externalServices
}
