package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type MessageParser func(msg sdk.Msg, msgIdx uint, tx *tx.Tx, txResponse *sdk.TxResponse) (GenericBackendResponse, error)
