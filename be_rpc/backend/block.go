package backend

import (
	"encoding/hex"
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	berpcutils "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/utils"
	"github.com/cosmos/cosmos-sdk/types/tx"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

func (m *Backend) GetLatestBlockNumber() (berpctypes.GenericBackendResponse, error) {
	statusInfo, err := m.clientCtx.Client.Status(m.ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return berpctypes.GenericBackendResponse{
		"latestBlock":             statusInfo.SyncInfo.LatestBlockHeight,
		"latestBlockTimeEpochUTC": statusInfo.SyncInfo.LatestBlockTime.UTC().Unix(),
	}, nil
}

func (m *Backend) GetBlockByNumber(height int64) (berpctypes.GenericBackendResponse, error) {
	resBlock, err := m.queryClient.ServiceClient.GetBlockWithTxs(m.ctx, &tx.GetBlockWithTxsRequest{
		Height: height,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if resBlock == nil {
		return nil, status.Error(codes.NotFound, "block not found")
	}

	resBlockResults, err := m.clientCtx.Client.BlockResults(m.ctx, &height)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if resBlockResults == nil {
		return nil, status.Error(codes.NotFound, "block results not found")
	}

	block := resBlock.Block

	response := berpctypes.GenericBackendResponse{
		"height":       block.Header.Height,
		"hash":         strings.ToUpper(hex.EncodeToString(resBlock.BlockId.Hash)),
		"timeEpochUTC": block.Header.Time.UTC().Unix(),
	}

	txsInfo := make([]map[string]any, 0)
	for i, tx := range resBlock.Txs {
		tmTx := tmtypes.Tx(resBlock.Block.Data.Txs[i])
		txHash := strings.ToUpper(hex.EncodeToString(tmTx.Hash()))
		txType := "cosmos"

		txResult := resBlockResults.TxsResults[i]

		if berpcutils.IsEvmTx(tx) {
			if evmTxHash := berpcutils.GetEvmTransactionHashFromEvent(txResult.Events); evmTxHash != nil {
				txHash = berpcutils.NormalizeTransactionHash(evmTxHash.String(), false)
				txType = "evm"
			}
		}

		txInfo := map[string]any{
			"hash":     txHash,
			"type":     txType,
			"code":     txResult.Code,
			"gasUsed":  txResult.GasUsed,
			"gasLimit": txResult.GasWanted,
		}

		msgTypes := make([]string, 0)
		for _, msg := range tx.Body.Messages {
			msgTypes = append(msgTypes, msg.TypeUrl)
		}
		txInfo["messages"] = msgTypes

		if tx.AuthInfo != nil {
			if tx.AuthInfo.Fee != nil {
				txInfo["fee"] = map[string]any{
					"gasLimit": tx.AuthInfo.Fee.GasLimit,
					"amount":   berpcutils.CoinsToMap(tx.AuthInfo.Fee.Amount...),
				}
			}
			if tx.AuthInfo.Tip != nil {
				txInfo["tip"] = map[string]any{
					"tipper": tx.AuthInfo.Tip.Tipper,
					"amount": berpcutils.CoinsToMap(tx.AuthInfo.Tip.Amount...),
				}
			}
		}

		txsInfo = append(txsInfo, txInfo)
	}

	response["txs"] = txsInfo

	return response, nil
}
