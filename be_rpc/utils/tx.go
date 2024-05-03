package utils

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"strings"
)

// NormalizeTransactionHash normalizes the transaction hash into '0xHASH' format.
// Contract: input hash is a valid transaction hash, with or without '0x' prefix.
func NormalizeTransactionHash(hash string, upper bool) string {
	hash = strings.ToLower(hash)
	if strings.HasPrefix(hash, "0x") {
		if upper {
			return "0x" + strings.ToUpper(hash[2:])
		} else {
			return strings.ToLower(hash)
		}
	} else {
		if upper {
			return "0x" + strings.ToUpper(hash)
		} else {
			return "0x" + hash
		}
	}
}

func ProtoMessageName(msg sdk.Msg) string {
	return proto.MessageName(msg)
}

func IsEvmTx(tx *tx.Tx) bool {
	for _, msg := range tx.Body.Messages {
		if strings.HasSuffix(msg.TypeUrl, ".MsgEthereumTx") {
			return true
		}
	}
	return false
}

func GetEvmTransactionHashFromEvent(events []abci.Event) *common.Hash {
	for _, event := range events {
		if event.Type != berpctypes.EventTypeEthereumTx {
			continue
		}
		for _, attr := range event.Attributes {
			if string(attr.Key) == berpctypes.AttributeKeyEthereumTxHash {
				hash := common.HexToHash(string(attr.Value))
				return &hash
			}
		}
	}
	return nil
}

func ConvertTxIntoTmTx(tx *tx.Tx, txConfig client.TxConfig) (tmtypes.Tx, error) {
	return txConfig.TxEncoder()(authtx.WrapTx(tx).(sdk.Tx))
}
