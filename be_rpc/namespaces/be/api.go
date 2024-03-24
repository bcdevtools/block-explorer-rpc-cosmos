package be

import (
	"fmt"
	"github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/backend"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/tendermint/tendermint/libs/log"
)

// API is the Block Explorer JSON-RPC.
type API struct {
	ctx     *server.Context
	logger  log.Logger
	backend backend.BackendI
}

// NewBeAPI creates an instance of the Block Explorer API.
func NewBeAPI(
	ctx *server.Context,
	backend backend.BackendI,
) *API {
	return &API{
		ctx:     ctx,
		logger:  ctx.Logger.With("api", "be"),
		backend: backend,
	}
}

func (api *API) Echo(text string) string {
	api.logger.Debug("be_echo")
	return fmt.Sprintf("hello \"%s\" from RollApp Block Explorer API", text)
}
