package backend

import (
	"encoding/hex"
	"fmt"
	"github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/constants"
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	berpcutils "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/libs/math"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
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

func (m *Backend) GetRecentBlocks(pageNo, pageSize int) (berpctypes.GenericBackendResponse, error) {
	pageNo = math.MaxInt(1, pageNo)
	pageSize = math.MaxInt(1, pageSize)

	const maxPageSize = 100
	if pageSize > maxPageSize {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("page size exceeds maximum allowed value %d", maxPageSize))
	}

	statusInfo, err := m.clientCtx.Client.Status(m.ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	latestBlockNumber := statusInfo.SyncInfo.LatestBlockHeight

	startBlockNumber := latestBlockNumber - int64(pageNo)*int64(pageSize) + 1
	startBlockNumber = math.MaxInt64(1, startBlockNumber)

	endBlockNumber := startBlockNumber + int64(pageSize) - 1
	endBlockNumber = math.MinInt64(latestBlockNumber, endBlockNumber)

	blocksInfo := make([]berpctypes.GenericBackendResponse, 0)
	for h := startBlockNumber; h <= endBlockNumber; h++ {
		resBlock, err := m.queryClient.ServiceClient.GetBlockWithTxs(m.ctx, &tx.GetBlockWithTxsRequest{
			Height: h,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if resBlock == nil {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("block not found %d", h))
		}
		blocksInfo = append(blocksInfo, m.getBasicBlockInformation(resBlock))
	}

	return berpctypes.GenericBackendResponse{
		"latestBlock":             latestBlockNumber,
		"latestBlockTimeEpochUTC": statusInfo.SyncInfo.LatestBlockTime.UTC().Unix(),
		"pageNumber":              pageNo,
		"pageSize":                pageSize,
		"blocks":                  blocksInfo,
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
		return nil, status.Error(codes.NotFound, fmt.Sprintf("block not found %d", height))
	}

	response := m.getBasicBlockInformation(resBlock)
	delete(response, "txsCount") // maintains legacy format

	sdkCtx := sdk.NewContext(nil, resBlock.Block.Header, false, nil).
		WithBlockHeight(resBlock.Block.Header.Height).
		WithBlockTime(resBlock.Block.Header.Time)
	queryCtx := sdk.WrapSDKContext(sdkCtx)

	txsInfo := make([]map[string]any, 0)
	for _, txBz := range resBlock.Block.Data.Txs {
		tmTx := tmtypes.Tx(txBz)
		txResponse, errGetTx := m.queryClient.GetTx(queryCtx, &tx.GetTxRequest{
			Hash: fmt.Sprintf("%X", tmTx.Hash()),
		})
		if errGetTx != nil {
			if strings.Contains(errGetTx.Error(), "not found") {
				continue
			}
			err := errors.Wrap(errGetTx, fmt.Sprintf("failed to get tx %X", tmTx.Hash()))
			return nil, status.Error(codes.Internal, err.Error())
		}

		tx := txResponse.Tx
		recheckTx := resBlock.Txs[len(txsInfo)]

		// TODO BE: remove this check once confirmed the issue not happens again
		if !reflect.DeepEqual(*tx, *recheckTx) {
			err := fmt.Errorf("txs mismatch, expected %v, got %v", tx, recheckTx)
			return nil, status.Error(codes.Internal, err.Error())
		}

		txHash := strings.ToUpper(hex.EncodeToString(tmTx.Hash()))

		const txTypeCosmos = "cosmos"
		const txTypeEvm = "evm"
		const txTypeWasm = "wasm"
		txType := txTypeCosmos

		txResult := txResponse.TxResponse

		evmTxAction := constants.EvmActionNone
		var evmTxSignature string

		wasmTxAction := constants.WasmActionNone
		var wasmTxSignature string

		if berpcutils.IsEvmTx(tx) {
			if evmTxHash := berpcutils.GetEvmTransactionHashFromEvent(txResult.Events); evmTxHash != nil {
				txHash = berpcutils.NormalizeTransactionHash(evmTxHash.String(), false)
				txType = txTypeEvm

				_absolutelyEvmTx, _evmTxAction, _evmTxSignature, _, errEvmTxInfo := m.getEvmTransactionInfo(txHash)
				if errEvmTxInfo == nil && _absolutelyEvmTx && _evmTxAction != constants.EvmActionNone {
					evmTxAction = _evmTxAction
					evmTxSignature = _evmTxSignature
				}
			}
		}

		for _, msg := range tx.Body.Messages {
			_absolutelyWasmTx, _wasmTxAction, _wasmTxSignature := m.getWasmTransactionInfo(msg)
			if _absolutelyWasmTx && _wasmTxAction != constants.WasmActionNone {
				txType = txTypeWasm
				wasmTxAction = _wasmTxAction
				wasmTxSignature = _wasmTxSignature
				break
			}
		}

		txInfo := map[string]any{
			"hash":     txHash,
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

		if txType == txTypeEvm {
			evmTxInfo := make(map[string]any)
			txInfo["evmTx"] = evmTxInfo
			if len(evmTxAction) > 0 {
				evmTxInfo["action"] = evmTxAction
			}
			if len(evmTxSignature) > 0 {
				evmTxInfo["sig"] = strings.TrimSpace(strings.ToLower(evmTxSignature))
			}
		} else if txType == txTypeWasm {
			wasmTxInfo := make(map[string]any)
			txInfo["wasmTx"] = wasmTxInfo
			if len(wasmTxAction) > 0 {
				wasmTxInfo["action"] = wasmTxAction
			}
			if len(wasmTxSignature) > 0 {
				wasmTxInfo["sig"] = strings.TrimSpace(strings.ToLower(wasmTxSignature))
			}
		}

		txInfo["type"] = txType

		txsInfo = append(txsInfo, txInfo)
	}

	response["txs"] = txsInfo

	return response, nil
}

func (m *Backend) getBasicBlockInformation(resBlock *tx.GetBlockWithTxsResponse) berpctypes.GenericBackendResponse {
	block := resBlock.Block
	result := berpctypes.GenericBackendResponse{
		"height":       block.Header.Height,
		"hash":         strings.ToUpper(hex.EncodeToString(resBlock.BlockId.Hash)),
		"timeEpochUTC": block.Header.Time.UTC().Unix(),
		"txsCount":     len(resBlock.Txs),
	}

	proposerConsAddr := sdk.ConsAddress(block.Header.GetProposerAddress()).String()
	var proposerMoniker string
	if stakingValidators, err := m.stakingValidatorsCache.GetValidators(); err == nil {
		for _, stakingValidator := range stakingValidators {
			if stakingValidator.consAddr == proposerConsAddr {
				proposerMoniker = stakingValidator.validator.Description.Moniker
				break
			}
		}
	}

	result["proposer"] = berpctypes.GenericBackendResponse{
		"consensusAddress": proposerConsAddr,
		"moniker":          proposerMoniker,
	}

	return result
}
