package backend

import (
	"cosmossdk.io/errors"
	"encoding/hex"
	"fmt"
	"github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/constants"
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	berpcutils "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/utils"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypeslegacy "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/big"
	"regexp"
	"strings"
)

var patternTxHash = regexp.MustCompile(`^(0[xX])?[\da-fA-F]{64}$`)

func (m *Backend) GetTransactionsInBlockRange(fromHeightIncluded, toHeightIncluded int64) (berpctypes.GenericBackendResponse, error) {
	if toHeightIncluded == 0 {
		toHeightIncluded = fromHeightIncluded
	}
	if fromHeightIncluded <= 0 || toHeightIncluded <= 0 || fromHeightIncluded > toHeightIncluded {
		return nil, berpctypes.ErrBadRequest
	}

	res := make(berpctypes.GenericBackendResponse)

	const maxPageSize = 100

	if toHeightIncluded-fromHeightIncluded+1 > maxPageSize {
		originalToHeightIncluded := toHeightIncluded
		toHeightIncluded = fromHeightIncluded + maxPageSize - 1
		res["skippedBlockRange"] = []int64{toHeightIncluded + 1, originalToHeightIncluded}
	}

	statusInfo, err := m.clientCtx.Client.Status(m.ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res["chainId"] = statusInfo.NodeInfo.Network

	missingBlocks := make(berpctypes.Tracker[int64])
	errorBlocks := make(berpctypes.Tracker[int64])

	blockInfoByHeight := make(map[int64]map[string]any)
	for height := fromHeightIncluded; height <= toHeightIncluded; height++ {
		blockInfo, missing, err := m.getTransactionsInBlock(height)
		if err != nil {
			m.GetLogger().Error("failed to get transactions in block", "height", height, "error", err)
		}

		if missing {
			missingBlocks.Add(height)
			continue
		}

		if err != nil {
			errorBlocks.Add(height)
			continue
		}

		blockInfoByHeight[height] = blockInfo
	}

	res["blocks"] = blockInfoByHeight

	if len(missingBlocks) > 0 {
		res["missingBlocks"] = missingBlocks.ToSortedSlice()
	}
	if len(errorBlocks) > 0 {
		res["errorBlocks"] = errorBlocks.ToSortedSlice()
	}

	return res, nil
}

func (m *Backend) getTransactionsInBlock(height int64) (blockInfo map[string]any, missing bool, err error) {
	defer func() {
		recv := recover()
		if recv != nil {
			err = fmt.Errorf("panic: %v", recv)
		}
	}()

	resBlock, errGetBlock := m.queryClient.GetBlockWithTxs(m.ctx, &tx.GetBlockWithTxsRequest{
		Height: height,
	})
	if errGetBlock != nil {
		missing = true
		err = errors.Wrap(errGetBlock, "failed to get block with txs")
		return
	}
	if resBlock == nil {
		missing = true
		err = fmt.Errorf("block not found")
		return
	}

	const txTypeCosmos = "cosmos"
	const txTypeEvm = "evm"
	const txTypeWasm = "wasm"

	var txsInfo []map[string]any
	for txIdx := 0; txIdx < len(resBlock.Block.Data.Txs); txIdx++ {
		tx := resBlock.Txs[txIdx]
		tmTx := tmtypes.Tx(resBlock.Block.Data.Txs[txIdx])
		txHash := strings.ToUpper(hex.EncodeToString(tmTx.Hash()))
		txType := txTypeCosmos

		evmTxAction := constants.EvmActionNone
		var evmTxSignature string

		wasmTxAction := constants.WasmActionNone
		var wasmTxSignature string

		var ibcPacketsInfo []map[string]any

		txValue := sdk.Coins{}

		var optionalTxResult *coretypes.ResultTx

		if berpcutils.IsEvmTx(tx) {
			var errTxResult error
			optionalTxResult, errTxResult = m.clientCtx.Client.Tx(m.ctx, tmTx.Hash(), false)
			if errTxResult != nil {
				m.GetLogger().Error("failed to query tx for evm information", "hash", tmTx.Hash(), "height", height, "error", errTxResult)
				// TODO BE: find another way to handle properly when error
			} else if optionalTxResult == nil {
				// ignore
			} else if evmTxHash := berpcutils.GetEvmTransactionHashFromEvent(optionalTxResult.TxResult.Events); evmTxHash != nil {
				txHash = berpcutils.NormalizeTransactionHash(evmTxHash.String(), false)
				txType = txTypeEvm

				// if the tx is an EVM tx, we need to get method signature
				txByHash, err := m.GetTransactionByHash(txHash)
				if err == nil && txByHash != nil {
					if evmTx, errConvert := berpcutils.TryConvertAnyStructToMap(txByHash["evmTx"]); errConvert == nil && len(evmTx) > 0 {
						var toStr, inputSigStr string
						if to, ok := berpcutils.TryGetMapValueAsType[string](evmTx, "to"); ok && len(to) > 0 {
							toStr = to
						}
						if input, ok := berpcutils.TryGetMapValueAsType[string](evmTx, "input"); ok && len(input) > 0 {
							if strings.HasPrefix(input, "0x") {
								if len(input) >= 10 {
									inputSigStr = input[:10]
								}
							} else {
								if len(input) >= 8 {
									inputSigStr = "0x" + input[:8]
								}
							}
						}

						if value, ok := berpcutils.TryGetMapValueAsType[string](evmTx, "value"); ok && len(value) > 0 {
							if strings.HasPrefix(value, "0x") && len(value) > 2 {
								if bi, ok := new(big.Int).SetString(value[2:], 16); ok && bi.Sign() > 0 {
									if evmModuleParams, err := m.GetModuleParams("evm"); err == nil && len(evmModuleParams) > 0 {
										if evmDenom, found := evmModuleParams["evm_denom"].(string); found && len(evmDenom) > 0 {
											txValue = txValue.Add(sdk.NewCoin(evmDenom, sdk.NewIntFromBigInt(bi)))
										}
									}
								}
							}
						}
						if toStr == "" {
							evmTxAction = constants.EvmActionCreate
						} else if inputSigStr == "" {
							evmTxAction = constants.EvmActionTransfer
						} else {
							evmTxAction = constants.EvmActionCall
							evmTxSignature = inputSigStr
						}
					}
				}
			}
		}

		var involvers berpctypes.MessageInvolversResult
		var messagesType []string

		for _, msg := range tx.Body.Messages {
			messagesType = append(messagesType, msg.TypeUrl)

			var cosmosMsg sdk.Msg
			errUnpack := m.clientCtx.Codec.UnpackAny(msg, &cosmosMsg)
			if errUnpack != nil {
				err = errors.Wrap(errUnpack, fmt.Sprintf("failed to unpack message %s", msg.TypeUrl))
				return
			}

			if wasmTxAction == constants.WasmActionNone {
				switch msg.TypeUrl {
				case "/cosmwasm.wasm.v1.MsgInstantiateContract":
					txType = txTypeWasm
					wasmTxAction = constants.WasmActionCreate
					wasmTxSignature = ""
				case "/cosmwasm.wasm.v1.MsgExecuteContract":
					txType = txTypeWasm
					wasmTxAction = constants.WasmActionCall
					if msgContent, errDecode := berpcutils.FromAnyToJsonMap(msg, m.clientCtx.Codec); errDecode == nil {
						if execMsg, errConvert := berpcutils.TryConvertAnyStructToMap(msgContent["msg"]); errConvert == nil && len(execMsg) > 0 {
							for k := range execMsg {
								wasmTxSignature = k
								break
							}
						}
					}
				}
			}

			var messageInvolversExtractor berpctypes.MessageInvolversExtractor
			if extractor, found := m.messageInvolversExtractors[berpcutils.ProtoMessageName(cosmosMsg)]; found {
				messageInvolversExtractor = extractor
			} else {
				messageInvolversExtractor = m.defaultMessageInvolversExtractor
			}

			resInvolvers, errExtractInvolvers := messageInvolversExtractor(cosmosMsg, tx, tmTx, m.clientCtx)
			if errExtractInvolvers == nil {
				if _, found := resInvolvers.GenericInvolvers()[berpctypes.MessageSenderSigner]; !found {
					// if no signer found, try to get it from the signers
					func() {
						defer func() {
							_ = recover() // omit any error
						}()
						if len(cosmosMsg.GetSigners()) > 0 {
							resInvolvers.AddGenericInvolvers(berpctypes.MessageSenderSigner, cosmosMsg.GetSigners()[0].String())
						}
					}()
				}
				involvers = resInvolvers
			} else {
				m.GetLogger().Error("failed to extract involvers", "error", errExtractInvolvers)
			}

			switch ibcMsg := cosmosMsg.(type) {
			case *channeltypes.MsgRecvPacket:
				ibcPacketsInfo = append(ibcPacketsInfo, buildIbcPacketInfoFromPacket(ibcMsg.Packet, true))
				if coin := getTxValueInIbcPacketInfo(ibcMsg.Packet); !coin.IsZero() {
					txValue = txValue.Add(coin)
				}
			case *channeltypes.MsgAcknowledgement:
				ibcPacketsInfo = append(ibcPacketsInfo, buildIbcPacketInfoFromPacket(ibcMsg.Packet, false))
				if coin := getTxValueInIbcPacketInfo(ibcMsg.Packet); !coin.IsZero() {
					txValue = txValue.Add(coin)
				}
			case *channeltypes.MsgTimeout:
				ibcPacketsInfo = append(ibcPacketsInfo, buildIbcPacketInfoFromPacket(ibcMsg.Packet, false))
				if coin := getTxValueInIbcPacketInfo(ibcMsg.Packet); !coin.IsZero() {
					txValue = txValue.Add(coin)
				}
			case *channeltypes.MsgTimeoutOnClose:
				ibcPacketsInfo = append(ibcPacketsInfo, buildIbcPacketInfoFromPacket(ibcMsg.Packet, false))
				if coin := getTxValueInIbcPacketInfo(ibcMsg.Packet); !coin.IsZero() {
					txValue = txValue.Add(coin)
				}
			case *ibctransfertypes.MsgTransfer:
				if optionalTxResult == nil {
					var errTxResult error
					optionalTxResult, errTxResult = m.clientCtx.Client.Tx(m.ctx, tmTx.Hash(), false)
					if errTxResult != nil {
						err = errors.Wrap(errTxResult, fmt.Sprintf("failed to query tx %s", hex.EncodeToString(tmTx.Hash())))
						return
					}
				}

				for _, event := range optionalTxResult.TxResult.Events {
					ok, kv := berpcutils.IsEventTypeWithAllAttributes(
						event,
						channeltypes.EventTypeSendPacket,
						channeltypes.AttributeKeySequence,
						channeltypes.AttributeKeySrcPort,
						channeltypes.AttributeKeySrcChannel,
						channeltypes.AttributeKeyDstPort,
						channeltypes.AttributeKeyDstChannel,
					)

					if ok {
						ibcPacketsInfo = append(ibcPacketsInfo, buildIbcPacketInfo(
							kv[channeltypes.AttributeKeySequence],
							kv[channeltypes.AttributeKeySrcPort], kv[channeltypes.AttributeKeySrcChannel],
							kv[channeltypes.AttributeKeyDstPort], kv[channeltypes.AttributeKeyDstChannel],
							false,
						))
						break
					}
				}

				if !ibcMsg.Token.IsZero() {
					txValue = txValue.Add(ibcMsg.Token)
				}
			}

			switch msgWithValue := cosmosMsg.(type) {
			case *banktypes.MsgSend:
				if !msgWithValue.Amount.IsZero() {
					txValue = txValue.Add(msgWithValue.Amount...)
				}
			case *banktypes.MsgMultiSend:
				var totalValue sdk.Coins
				for _, output := range msgWithValue.Outputs {
					totalValue = totalValue.Add(output.Coins...)
				}
				if !totalValue.IsZero() {
					txValue = txValue.Add(totalValue...)
				}
			case *stakingtypes.MsgDelegate:
				if !msgWithValue.Amount.IsZero() {
					txValue = txValue.Add(msgWithValue.Amount)
				}
			case *stakingtypes.MsgUndelegate:
				if !msgWithValue.Amount.IsZero() {
					txValue = txValue.Add(msgWithValue.Amount)
				}
			case *stakingtypes.MsgBeginRedelegate:
				if !msgWithValue.Amount.IsZero() {
					txValue = txValue.Add(msgWithValue.Amount)
				}
			}
		}

		involvers.Finalize()

		txInfo := map[string]any{
			"hash":         txHash,
			"type":         txType,
			"involvers":    involvers.ToResponseObject(),
			"messagesType": messagesType,
		}
		if len(ibcPacketsInfo) > 0 {
			txInfo["ibcPacketsInfo"] = ibcPacketsInfo
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
		if !txValue.IsZero() {
			txInfo["value"] = txValue.String()
		}
		txsInfo = append(txsInfo, txInfo)
	}

	blockInfo = map[string]any{
		"timeEpochUTC": resBlock.Block.Header.Time.UTC().Unix(),
		"txs":          txsInfo,
	}
	return
}

func (m *Backend) GetTransactionByHash(hashStr string) (berpctypes.GenericBackendResponse, error) {
	if !patternTxHash.MatchString(hashStr) {
		return nil, berpctypes.ErrBadRequest
	}

	if m.interceptor != nil {
		intercepted, response, err := m.interceptor.GetTransactionByHash(hashStr)
		if intercepted {
			return response, err
		}
	}

	hash := berpcutils.NormalizeTransactionHash(hashStr, true)

	res, err := m.queryClient.GetTx(m.ctx, &tx.GetTxRequest{
		Hash: hash[2:],
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if res == nil {
		return nil, status.Error(codes.NotFound, "transaction not found")
	}

	tx := res.Tx
	txRes := res.TxResponse
	txEvents := berpctypes.ConvertTxEvent(txRes.Events).RemoveUnnecessaryEvmTxEvents()

	msgsInfo := make([]map[string]any, 0)
	for msgIdx, msg := range tx.Body.Messages {
		var cosmosMsg sdk.Msg

		err := m.clientCtx.Codec.UnpackAny(msg, &cosmosMsg)
		if err != nil {
			return nil, status.Error(codes.Internal, errors.Wrapf(err, "failed to unpack message").Error())
		}

		protoType := berpcutils.ProtoMessageName(cosmosMsg)

		msgInfo := map[string]any{
			"idx":  msgIdx,
			"type": protoType,
		}
		msgsInfo = append(msgsInfo, msgInfo)

		var messageParser berpctypes.MessageParser
		if customParser, found := m.messageParsers[protoType]; found {
			messageParser = customParser
		} else {
			messageParser = m.defaultMessageParser
		}

		parsedContent, err := messageParser(cosmosMsg, uint(msgIdx), tx, txRes)
		if err != nil {
			msgInfo["contentError"] = err.Error()
		} else {
			msgInfo["content"] = parsedContent
		}

		{
			msgContent, err := berpcutils.FromAnyToJsonMap(msg, m.clientCtx.Codec)
			if err != nil {
				msgInfo["protoContentError"] = err.Error()
			} else {
				msgInfo["protoContent"] = msgContent
			}
		}
	}

	response := berpctypes.GenericBackendResponse{
		"height": txRes.Height,
		"hash":   txRes.TxHash,
		"msgs":   msgsInfo,
		"result": berpctypes.GenericBackendResponse{
			"code":    txRes.Code,
			"success": txRes.Code == 0,
			"gas": berpctypes.GenericBackendResponse{
				"limit": txRes.GasWanted,
				"used":  txRes.GasUsed,
			},
			"events": txEvents,
		},
	}

	if len(tx.Body.Memo) > 0 {
		response["memo"] = tx.Body.Memo
	}

	return response, nil
}

func (m *Backend) defaultMessageParser(msg sdk.Msg, msgIdx uint, tx *tx.Tx, txResponse *sdk.TxResponse) (res berpctypes.GenericBackendResponse, err error) {
	switch msg := msg.(type) {
	case *banktypes.MsgSend:
		res = berpctypes.GenericBackendResponse{
			"transfer": map[string]any{
				"from": []string{msg.FromAddress},
				"to": []map[string]any{
					{
						"address": msg.ToAddress,
						"amount":  berpcutils.CoinsToMap(msg.Amount...),
					},
				},
			},
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.FromAddress).
			WriteText(" transfers ").
			WriteCoins(msg.Amount, m.getBankDenomsMetadata(msg.Amount)).
			WriteText(" to ").
			WriteAddress(msg.ToAddress).
			BuildIntoResponse(res)

		return
	case *banktypes.MsgMultiSend:
		rb := berpctypes.NewFriendlyResponseContentBuilder()

		var fromAddresses []string
		for i, input := range msg.Inputs {
			fromAddresses = append(fromAddresses, input.Address)

			if i > 0 {
				rb.WriteText(", ")
			}
			rb.WriteAddress(input.Address)
		}

		rb.WriteText(" transfer ")

		var allCoins sdk.Coins
		for _, output := range msg.Outputs {
			allCoins = allCoins.Add(output.Coins...)
		}

		toInfo := make([]map[string]any, 0)
		for i, output := range msg.Outputs {
			toInfo = append(toInfo, map[string]any{
				"address": output.Address,
				"amount":  berpcutils.CoinsToMap(output.Coins...),
			})

			if i > 0 {
				rb.WriteText(", ")
			}
			rb.WriteCoins(output.Coins, m.getBankDenomsMetadata(allCoins)).
				WriteText(" to ").
				WriteAddress(output.Address)
		}

		res = berpctypes.GenericBackendResponse{
			"transfer": map[string]any{
				"from": fromAddresses,
				"to":   toInfo,
			},
		}

		rb.BuildIntoResponse(res)

		return
	case *crisistypes.MsgVerifyInvariant:
		res = berpctypes.GenericBackendResponse{
			"sender":              msg.Sender,
			"invariantModuleName": msg.InvariantModuleName,
			"invariantRoute":      msg.InvariantRoute,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Sender).
			WriteText(" verifies invariant ").
			WriteText(msg.InvariantModuleName).
			WriteText(" at route ").
			WriteText(msg.InvariantRoute).
			BuildIntoResponse(res)

		return
	case *disttypes.MsgSetWithdrawAddress:
		res = berpctypes.GenericBackendResponse{
			"delegatorAddress": msg.DelegatorAddress,
			"withdrawAddress":  msg.WithdrawAddress,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.DelegatorAddress).
			WriteText(" sets withdraw address to ").
			WriteAddress(msg.WithdrawAddress).
			BuildIntoResponse(res)

		return
	case *disttypes.MsgWithdrawDelegatorReward:
		res = berpctypes.GenericBackendResponse{
			"delegatorAddress": msg.DelegatorAddress,
			"validatorAddress": msg.ValidatorAddress,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.DelegatorAddress).
			WriteText(" withdraws rewards from ").
			WriteAddress(msg.ValidatorAddress).
			BuildIntoResponse(res)

		return
	case *disttypes.MsgWithdrawValidatorCommission:
		res = berpctypes.GenericBackendResponse{
			"validatorAddress": msg.ValidatorAddress,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.ValidatorAddress).
			WriteText(" withdraws commission").
			BuildIntoResponse(res)

		return
	case *disttypes.MsgFundCommunityPool:
		res = berpctypes.GenericBackendResponse{
			"depositor": msg.Depositor,
			"amount":    berpcutils.CoinsToMap(msg.Amount...),
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Depositor).
			WriteText(" funds community pool with ").
			WriteCoins(msg.Amount, m.getBankDenomsMetadata(msg.Amount)).
			BuildIntoResponse(res)

		return
	case *evidencetypes.MsgSubmitEvidence:
		res = berpctypes.GenericBackendResponse{
			"submitter": msg.Submitter,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Submitter).
			WriteText(" submits evidence").
			BuildIntoResponse(res)

		return
	case *govtypesv1.MsgSubmitProposal:
		var messageTypes []string
		for _, message := range msg.Messages {
			messageTypes = append(messageTypes, message.TypeUrl)
		}

		res = berpctypes.GenericBackendResponse{
			"proposer":     msg.Proposer,
			"metadata":     msg.Metadata,
			"deposit":      berpcutils.CoinsToMap(msg.InitialDeposit...),
			"messageTypes": messageTypes,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Proposer).
			WriteText(" submits proposal of message types [").
			WriteText(strings.Join(messageTypes, ", ")).
			WriteText("] with initial deposit ").
			WriteCoins(msg.InitialDeposit, m.getBankDenomsMetadata(msg.InitialDeposit)).
			BuildIntoResponse(res)

		return
	case *govtypeslegacy.MsgSubmitProposal:
		res = berpctypes.GenericBackendResponse{
			"proposer":     msg.Proposer,
			"deposit":      berpcutils.CoinsToMap(msg.InitialDeposit...),
			"messageTypes": []string{msg.Content.TypeUrl},
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Proposer).
			WriteText(" submits proposal of message types [").
			WriteText(msg.Content.TypeUrl).
			WriteText("] with initial deposit ").
			WriteCoins(msg.InitialDeposit, m.getBankDenomsMetadata(msg.InitialDeposit)).
			BuildIntoResponse(res)

		return
	case *govtypesv1.MsgDeposit:
		res = berpctypes.GenericBackendResponse{
			"depositor":  msg.Depositor,
			"proposalId": msg.ProposalId,
			"amount":     berpcutils.CoinsToMap(msg.Amount...),
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Depositor).
			WriteText(" deposits ").
			WriteCoins(msg.Amount, m.getBankDenomsMetadata(msg.Amount)).
			WriteText(" to proposal ").
			WriteText(fmt.Sprintf("%d", msg.ProposalId)).
			BuildIntoResponse(res)

		return
	case *govtypeslegacy.MsgDeposit:
		res = berpctypes.GenericBackendResponse{
			"depositor":  msg.Depositor,
			"proposalId": msg.ProposalId,
			"amount":     berpcutils.CoinsToMap(msg.Amount...),
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Depositor).
			WriteText(" deposits ").
			WriteCoins(msg.Amount, m.getBankDenomsMetadata(msg.Amount)).
			WriteText(" to proposal ").
			WriteText(fmt.Sprintf("%d", msg.ProposalId)).
			BuildIntoResponse(res)

		return
	case *govtypesv1.MsgVote:
		res = berpctypes.GenericBackendResponse{
			"voter":      msg.Voter,
			"proposalId": msg.ProposalId,
			"option":     msg.Option.String(),
			"metadata":   msg.Metadata,
		}

		rb := berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Voter)

		switch msg.Option {
		case govtypesv1.OptionYes:
			rb.WriteText(" votes YES")
		case govtypesv1.OptionNo:
			rb.WriteText(" votes NO")
		case govtypesv1.OptionAbstain:
			rb.WriteText(" votes Abstains")
		case govtypesv1.OptionNoWithVeto:
			rb.WriteText(" votes NO with VETO")
		default:
			rb.WriteText(" votes ").
				WriteText(msg.Option.String())
		}

		rb.WriteText(" to proposal ").
			WriteText(fmt.Sprintf("%d", msg.ProposalId)).
			BuildIntoResponse(res)

		return
	case *govtypeslegacy.MsgVote:
		res = berpctypes.GenericBackendResponse{
			"voter":      msg.Voter,
			"proposalId": msg.ProposalId,
			"option":     msg.Option.String(),
		}

		rb := berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Voter)

		switch msg.Option {
		case govtypeslegacy.OptionYes:
			rb.WriteText(" votes YES")
		case govtypeslegacy.OptionNo:
			rb.WriteText(" votes NO")
		case govtypeslegacy.OptionAbstain:
			rb.WriteText(" votes Abstains")
		case govtypeslegacy.OptionNoWithVeto:
			rb.WriteText(" votes NO with VETO")
		default:
			rb.WriteText(" votes ").
				WriteText(msg.Option.String())
		}

		rb.WriteText(" to proposal ").
			WriteText(fmt.Sprintf("%d", msg.ProposalId)).
			BuildIntoResponse(res)

		return
	case *govtypesv1.MsgVoteWeighted:
		res = berpctypes.GenericBackendResponse{
			"voter":      msg.Voter,
			"proposalId": msg.ProposalId,
			"metadata":   msg.Metadata,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Voter).
			WriteText(" votes with weight to proposal ").
			WriteText(fmt.Sprintf("%d", msg.ProposalId)).
			BuildIntoResponse(res)

		return
	case *govtypeslegacy.MsgVoteWeighted:
		res = berpctypes.GenericBackendResponse{
			"voter":      msg.Voter,
			"proposalId": msg.ProposalId,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Voter).
			WriteText(" votes with weight to proposal ").
			WriteText(fmt.Sprintf("%d", msg.ProposalId)).
			BuildIntoResponse(res)

		return
	case *ibctypes.MsgCreateClient:
		res = berpctypes.GenericBackendResponse{
			"signer": msg.Signer,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" creates IBC client").
			BuildIntoResponse(res)

		return
	case *ibctypes.MsgUpdateClient:
		res = berpctypes.GenericBackendResponse{
			"signer":   msg.Signer,
			"clientId": msg.ClientId,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" updates IBC client ").
			WriteText(msg.ClientId).
			BuildIntoResponse(res)

		return
	case *ibctypes.MsgUpgradeClient:
		res = berpctypes.GenericBackendResponse{
			"signer":   msg.Signer,
			"clientId": msg.ClientId,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" upgrades IBC client ").
			WriteText(msg.ClientId).
			BuildIntoResponse(res)

		return
	case *ibctypes.MsgSubmitMisbehaviour:
		res = berpctypes.GenericBackendResponse{
			"signer":   msg.Signer,
			"clientId": msg.ClientId,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" submits IBC misbehaviour for client ").
			WriteText(msg.ClientId).
			BuildIntoResponse(res)

		return
	case *ibctransfertypes.MsgTransfer:
		res = berpctypes.GenericBackendResponse{
			"sender":          msg.Sender,
			"receiver":        msg.Receiver,
			"amount":          berpcutils.CoinsToMap(msg.Token),
			"timeoutHeight":   msg.TimeoutHeight,
			"timeoutEpochUTC": msg.TimeoutTimestamp,
			"sourcePort":      msg.SourcePort,
			"sourceChannel":   msg.SourceChannel,
		}

		if len(msg.Memo) > 0 {
			res["memo"] = msg.Memo
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Sender).
			WriteText(" transfers ").
			WriteCoins(sdk.Coins{msg.Token}, m.getBankDenomsMetadata(sdk.Coins{msg.Token})).
			WriteText(" to ").
			WriteAddress(msg.Receiver).
			WriteText(" through IBC via ").
			WriteText(msg.SourcePort).WriteText("/").WriteText(msg.SourceChannel).
			BuildIntoResponse(res)

		return
	case *connectiontypes.MsgConnectionOpenAck:
		res = berpctypes.GenericBackendResponse{
			"signer":                   msg.Signer,
			"connectionId":             msg.ConnectionId,
			"counterpartyConnectionId": msg.CounterpartyConnectionId,
			"proofHeight":              msg.ProofHeight,
			"consensusHeight":          msg.ConsensusHeight,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" acknowledges open connection ").
			WriteText(msg.ConnectionId).
			WriteText(" with counterparty connection ").
			WriteText(msg.CounterpartyConnectionId).
			BuildIntoResponse(res)

		return
	case *connectiontypes.MsgConnectionOpenInit:
		res = berpctypes.GenericBackendResponse{
			"signer":                   msg.Signer,
			"clientId":                 msg.ClientId,
			"counterpartyClientId":     msg.Counterparty.ClientId,
			"counterpartyConnectionId": msg.Counterparty.ConnectionId,
			"delayPeriod":              msg.DelayPeriod,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" initializes open connection ").
			WriteText(msg.ClientId).
			WriteText(" with counterparty client ").
			WriteText(msg.Counterparty.ClientId).
			WriteText(" and connection ").
			WriteText(msg.Counterparty.ConnectionId).
			BuildIntoResponse(res)

		return
	case *connectiontypes.MsgConnectionOpenConfirm:
		res = berpctypes.GenericBackendResponse{
			"signer":       msg.Signer,
			"connectionId": msg.ConnectionId,
			"proofHeight":  msg.ProofHeight,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" confirms open connection ").
			WriteText(msg.ConnectionId).
			BuildIntoResponse(res)

		return
	case *connectiontypes.MsgConnectionOpenTry:
		res = berpctypes.GenericBackendResponse{
			"signer":                   msg.Signer,
			"clientId":                 msg.ClientId,
			"counterpartyClientId":     msg.Counterparty.ClientId,
			"counterpartyConnectionId": msg.Counterparty.ConnectionId,
			"proofHeight":              msg.ProofHeight,
			"consensusHeight":          msg.ConsensusHeight,
			"delayPeriod":              msg.DelayPeriod,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" tries open connection ").
			WriteText(msg.ClientId).
			WriteText(" with counterparty client ").
			WriteText(msg.Counterparty.ClientId).
			WriteText(" and connection ").
			WriteText(msg.Counterparty.ConnectionId).
			BuildIntoResponse(res)

		return
	case *channeltypes.MsgChannelOpenInit:
		res = berpctypes.GenericBackendResponse{
			"signer":                msg.Signer,
			"portId":                msg.PortId,
			"counterPartyPortId":    msg.Channel.Counterparty.PortId,
			"counterPartyChannelId": msg.Channel.Counterparty.ChannelId,
			"connectionHops":        msg.Channel.ConnectionHops,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" initializes open channel for port ").
			WriteText(msg.PortId).
			WriteText(" connects to counterparty port ").
			WriteText(msg.Channel.Counterparty.PortId).
			WriteText(" and channel ").
			WriteText(msg.Channel.Counterparty.ChannelId).
			BuildIntoResponse(res)

		return
	case *channeltypes.MsgChannelOpenConfirm:
		res = berpctypes.GenericBackendResponse{
			"signer":      msg.Signer,
			"portId":      msg.PortId,
			"channelId":   msg.ChannelId,
			"proofHeight": msg.ProofHeight,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" confirms open channel for port ").
			WriteText(msg.PortId).
			WriteText(" and channel ").
			WriteText(msg.ChannelId).
			BuildIntoResponse(res)

		return
	case *channeltypes.MsgChannelOpenTry:
		res = berpctypes.GenericBackendResponse{
			"signer":                msg.Signer,
			"portId":                msg.PortId,
			"counterPartyPortId":    msg.Channel.Counterparty.PortId,
			"counterPartyChannelId": msg.Channel.Counterparty.ChannelId,
			"connectionHops":        msg.Channel.ConnectionHops,
			"proofHeight":           msg.ProofHeight,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" tries open channel for port ").
			WriteText(msg.PortId).
			WriteText(" connects to counterparty port ").
			WriteText(msg.Channel.Counterparty.PortId).
			WriteText(" and channel ").
			WriteText(msg.Channel.Counterparty.ChannelId).
			BuildIntoResponse(res)

		return
	case *channeltypes.MsgAcknowledgement:
		res = berpctypes.GenericBackendResponse{
			"signer":           msg.Signer,
			"proofHeight":      msg.ProofHeight,
			"timeoutHeight":    msg.Packet.TimeoutHeight,
			"timeoutTimestamp": msg.Packet.TimeoutTimestamp,
		}

		rb := berpctypes.NewFriendlyResponseContentBuilder()
		rb.WriteAddress(msg.Signer).WriteText(" acknowledges packet:")

		m.addIbcPacketInfoIntoResponse(msg.Packet, false, res, rb)

		rb.BuildIntoResponse(res)

		return
	case *channeltypes.MsgChannelOpenAck:
		res = berpctypes.GenericBackendResponse{
			"signer":              msg.Signer,
			"portId":              msg.PortId,
			"channelId":           msg.ChannelId,
			"counterPartyChannel": msg.CounterpartyChannelId,
			"proofHeight":         msg.ProofHeight,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Signer).
			WriteText(" acknowledges open channel for port ").
			WriteText(msg.PortId).
			WriteText(" and channel ").
			WriteText(msg.ChannelId).
			WriteText(" with counterparty channel ").
			WriteText(msg.CounterpartyChannelId).
			BuildIntoResponse(res)

		return
	case *channeltypes.MsgRecvPacket:
		res = berpctypes.GenericBackendResponse{
			"signer":          msg.Signer,
			"proofHeight":     msg.ProofHeight,
			"timeoutHeight":   msg.Packet.TimeoutHeight,
			"timeoutEpochUTC": msg.Packet.TimeoutTimestamp,
		}

		rb := berpctypes.NewFriendlyResponseContentBuilder()
		rb.WriteAddress(msg.Signer).WriteText(" informs receive packet:")

		m.addIbcPacketInfoIntoResponse(msg.Packet, true, res, rb)

		rb.BuildIntoResponse(res)

		return
	case *channeltypes.MsgTimeout:
		res = berpctypes.GenericBackendResponse{
			"signer":           msg.Signer,
			"proofHeight":      msg.ProofHeight,
			"timeoutHeight":    msg.Packet.TimeoutHeight,
			"timeoutEpochUTC":  msg.Packet.TimeoutTimestamp,
			"nextSequenceRecv": msg.NextSequenceRecv,
		}

		rb := berpctypes.NewFriendlyResponseContentBuilder()
		rb.WriteAddress(msg.Signer).WriteText(" informs packet timed out:")

		m.addIbcPacketInfoIntoResponse(msg.Packet, false, res, rb)

		rb.BuildIntoResponse(res)

		return
	case *channeltypes.MsgTimeoutOnClose:
		res = berpctypes.GenericBackendResponse{
			"signer":           msg.Signer,
			"proofHeight":      msg.ProofHeight,
			"timeoutHeight":    msg.Packet.TimeoutHeight,
			"timeoutEpochUTC":  msg.Packet.TimeoutTimestamp,
			"nextSequenceRecv": msg.NextSequenceRecv,
		}

		rb := berpctypes.NewFriendlyResponseContentBuilder()
		rb.WriteAddress(msg.Signer).WriteText(" informs closing timed out packet:")

		m.addIbcPacketInfoIntoResponse(msg.Packet, false, res, rb)

		rb.BuildIntoResponse(res)

		return
	case *slashingtypes.MsgUnjail:
		res = berpctypes.GenericBackendResponse{
			"validator": msg.ValidatorAddr,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.ValidatorAddr).
			WriteText(" un-jails").
			BuildIntoResponse(res)

		return
	case *stakingtypes.MsgCreateValidator:
		res = berpctypes.GenericBackendResponse{
			"validator": msg.ValidatorAddress,
			"delegator": msg.DelegatorAddress,
			"delegate":  berpcutils.CoinsToMap(msg.Value),
			"commission": map[string]string{
				"rate":          msg.Commission.Rate.String(),
				"maxRate":       msg.Commission.MaxRate.String(),
				"maxChangeRate": msg.Commission.MaxChangeRate.String(),
			},
			"minSelfDelegation": msg.MinSelfDelegation.String(),
		}

		res["description"] = map[string]string{}
		if msg.Description.Moniker != "" {
			res["description"].(map[string]string)["moniker"] = msg.Description.Moniker
		}
		if msg.Description.Identity != "" {
			res["description"].(map[string]string)["identity"] = msg.Description.Identity
		}
		if msg.Description.Website != "" {
			res["description"].(map[string]string)["website"] = msg.Description.Website
		}
		if msg.Description.SecurityContact != "" {
			res["description"].(map[string]string)["securityContact"] = msg.Description.SecurityContact
		}
		if msg.Description.Details != "" {
			res["description"].(map[string]string)["details"] = msg.Description.Details
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.DelegatorAddress).
			WriteText(" creates validator ").
			WriteAddress(msg.ValidatorAddress).
			WriteText(" with delegation ").
			WriteCoins(sdk.Coins{msg.Value}, m.getBankDenomsMetadata(sdk.Coins{msg.Value})).
			BuildIntoResponse(res)

		return
	case *stakingtypes.MsgEditValidator:
		res = berpctypes.GenericBackendResponse{
			"validator": msg.ValidatorAddress,
		}

		res["description"] = map[string]string{}
		if msg.Description.Moniker != "" {
			res["description"].(map[string]string)["moniker"] = msg.Description.Moniker
		}
		if msg.Description.Identity != "" {
			res["description"].(map[string]string)["identity"] = msg.Description.Identity
		}
		if msg.Description.Website != "" {
			res["description"].(map[string]string)["website"] = msg.Description.Website
		}
		if msg.Description.SecurityContact != "" {
			res["description"].(map[string]string)["securityContact"] = msg.Description.SecurityContact
		}
		if msg.Description.Details != "" {
			res["description"].(map[string]string)["details"] = msg.Description.Details
		}

		if msg.CommissionRate != nil {
			res["commission"] = map[string]string{
				"rate": msg.CommissionRate.String(),
			}
		}
		if msg.MinSelfDelegation != nil {
			res["minSelfDelegation"] = msg.MinSelfDelegation.String()
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.ValidatorAddress).
			WriteText(" updates validator information").
			BuildIntoResponse(res)

		return
	case *stakingtypes.MsgDelegate:
		res = berpctypes.GenericBackendResponse{
			"delegator": msg.DelegatorAddress,
			"validator": msg.ValidatorAddress,
			"amount":    berpcutils.CoinsToMap(msg.Amount),
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.DelegatorAddress).
			WriteText(" delegates ").
			WriteCoins(sdk.Coins{msg.Amount}, m.getBankDenomsMetadata(sdk.Coins{msg.Amount})).
			WriteText(" to ").
			WriteAddress(msg.ValidatorAddress).
			BuildIntoResponse(res)

		return
	case *stakingtypes.MsgBeginRedelegate:
		res = berpctypes.GenericBackendResponse{
			"delegator":     msg.DelegatorAddress,
			"validatorFrom": msg.ValidatorSrcAddress,
			"validatorTo":   msg.ValidatorDstAddress,
			"amount":        berpcutils.CoinsToMap(msg.Amount),
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.DelegatorAddress).
			WriteText(" re-delegates ").
			WriteCoins(sdk.Coins{msg.Amount}, m.getBankDenomsMetadata(sdk.Coins{msg.Amount})).
			WriteText(" from ").
			WriteAddress(msg.ValidatorSrcAddress).
			WriteText(" to ").
			WriteAddress(msg.ValidatorDstAddress).
			BuildIntoResponse(res)

		return
	case *stakingtypes.MsgUndelegate:
		res = berpctypes.GenericBackendResponse{
			"delegator": msg.DelegatorAddress,
			"validator": msg.ValidatorAddress,
			"amount":    berpcutils.CoinsToMap(msg.Amount),
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.DelegatorAddress).
			WriteText(" un-delegates ").
			WriteCoins(sdk.Coins{msg.Amount}, m.getBankDenomsMetadata(sdk.Coins{msg.Amount})).
			WriteText(" from ").
			WriteAddress(msg.ValidatorAddress).
			BuildIntoResponse(res)

		return
	case *stakingtypes.MsgCancelUnbondingDelegation:
		res = berpctypes.GenericBackendResponse{
			"delegator": msg.DelegatorAddress,
			"validator": msg.ValidatorAddress,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.DelegatorAddress).
			WriteText(" cancels unbonding delegation from ").
			WriteAddress(msg.ValidatorAddress).
			BuildIntoResponse(res)

		return
	case *authztypes.MsgGrant:
		res = berpctypes.GenericBackendResponse{
			"granter":       msg.Granter,
			"grantee":       msg.Grantee,
			"authorization": msg.Grant.Authorization.TypeUrl,
		}

		rb := berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Granter).
			WriteText(" grants ").
			WriteText(msg.Grant.Authorization.TypeUrl).
			WriteText(" to ").
			WriteAddress(msg.Grantee)

		if msg.Grant.Expiration != nil {
			res["expirationEpochUTC"] = msg.Grant.Expiration.UTC().Unix()
			rb.WriteText(" with expiration ").
				WriteText(msg.Grant.Expiration.UTC().String()).
				WriteText(" UTC")
		}

		rb.BuildIntoResponse(res)

		return
	case *authztypes.MsgExec:
		res = berpctypes.GenericBackendResponse{
			"grantee": msg.Grantee,
		}

		rb := berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Grantee).
			WriteText(" executes authorized messages")

		if len(msg.Msgs) > 0 {
			authorizedMessagesContent := make([]berpctypes.GenericBackendResponse, 0)

			for _, authorizedMsgAny := range msg.Msgs {
				var cosmosMsg sdk.Msg

				innerRes := berpctypes.GenericBackendResponse{
					"type": berpcutils.ProtoMessageName(msg),
				}

				authorizedMessagesContent = append(authorizedMessagesContent, innerRes)

				err := m.clientCtx.Codec.UnpackAny(authorizedMsgAny, &cosmosMsg)
				if err != nil {
					innerRes["error"] = errors.Wrap(err, "failed to unpack authorized message").Error()
					continue
				}

				parsedInnerRes, err := m.defaultMessageParser(cosmosMsg, 0, tx, txResponse)
				if err != nil {
					innerRes["error"] = errors.Wrap(err, "failed to parse authorized message").Error()
					continue
				}

				innerRes["content"] = parsedInnerRes
			}

			res["authorized-messages"] = authorizedMessagesContent
		}

		rb.BuildIntoResponse(res)

		return
	case *authztypes.MsgRevoke:
		res = berpctypes.GenericBackendResponse{
			"granter":       msg.Granter,
			"grantee":       msg.Grantee,
			"authorization": msg.MsgTypeUrl,
		}

		berpctypes.NewFriendlyResponseContentBuilder().
			WriteAddress(msg.Granter).
			WriteText(" revokes permission ").
			WriteText(msg.MsgTypeUrl).
			WriteText(" from ").
			WriteAddress(msg.Grantee).
			BuildIntoResponse(res)

		return
	}

	return nil, berpctypes.ErrNotSupportedMessageType
}

func (m *Backend) getBankDenomsMetadata(coins sdk.Coins) map[string]banktypes.Metadata {
	denomsMetadata := make(map[string]banktypes.Metadata)
	for _, coin := range coins {
		res, err := m.queryClient.BankQueryClient.DenomMetadata(m.ctx, &banktypes.QueryDenomMetadataRequest{
			Denom: coin.Denom,
		})
		if err != nil || res == nil || coin.Denom == "" {
			continue
		}
		denomsMetadata[coin.Denom] = res.Metadata
	}

	if len(denomsMetadata) == 0 {
		// trying to insert denom metadata for the default RollApp coin
		const defaultDenom = "urax"
		const defaultDisplay = "RAX"
		if found, _ := coins.Find(defaultDenom); found {
			denomsMetadata[defaultDenom] = banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    defaultDenom,
					Exponent: 0,
				}, {
					Denom:    defaultDisplay,
					Exponent: 18,
				}},
				Base:    defaultDenom,
				Display: defaultDisplay,
				Name:    defaultDisplay,
				Symbol:  defaultDisplay,
			}
		}
	}

	return denomsMetadata

}

func (m *Backend) addIbcPacketInfoIntoResponse(packet channeltypes.Packet, incomingPacket bool, res berpctypes.GenericBackendResponse, rb berpctypes.FriendlyResponseContentBuilderI) {
	res["ibcPacketInfo"] = buildIbcPacketInfoFromPacket(packet, incomingPacket)

	var data ibctransfertypes.FungibleTokenPacketData
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.Data, &data); err == nil {
		token, err := berpcutils.GetIncomingIBCCoin(
			packet.SourcePort, packet.SourceChannel,
			packet.DestinationPort, packet.DestinationChannel,
			data.Denom, data.Amount,
		)
		if err == nil {
			tokens := sdk.Coins{token}

			res["transfer"] = map[string]any{
				"from": []string{data.Sender},
				"to": []map[string]any{
					{
						"address": data.Receiver,
						"amount":  berpcutils.CoinsToMap(tokens...),
					},
				},
			}

			rb.WriteText(data.Sender).
				WriteText(" transfers ").
				WriteCoins(tokens, m.getBankDenomsMetadata(tokens)).
				WriteText(" to ").
				WriteAddress(data.Receiver)
		} else {
			m.GetLogger().Error("failed to get incoming IBC coin", "error", err)

			rb.WriteText(data.Sender).
				WriteText(" transfers unknown amount (parse error)").
				WriteText(" to ").
				WriteAddress(data.Receiver)
		}

		rb.WriteText(" through IBC via ")

		if incomingPacket {
			// MsgRecvPacket
			rb.WriteText(packet.DestinationPort).WriteText("/").WriteText(packet.DestinationChannel).
				WriteText(" from ").
				WriteText(packet.SourcePort).WriteText("/").WriteText(packet.SourceChannel)
		} else {
			// MsgAcknowledgement
			rb.WriteText(packet.SourcePort).WriteText("/").WriteText(packet.SourceChannel).
				WriteText(" to ").
				WriteText(packet.DestinationPort).WriteText("/").WriteText(packet.DestinationChannel)
		}

		if data.Memo != "" {
			res["memo"] = data.Memo
			rb.WriteText(" with memo ").WriteText(data.Memo)
		}
	}
}

func (m *Backend) defaultMessageInvolversExtractor(msg sdk.Msg, tx *tx.Tx, tmTx tmtypes.Tx, clientCtx client.Context) (res berpctypes.MessageInvolversResult, err error) {
	res = berpctypes.NewMessageInvolversResult()

	switch msg := msg.(type) {
	case *banktypes.MsgSend:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.FromAddress, msg.ToAddress)
		return
	case *banktypes.MsgMultiSend:
		for _, input := range msg.Inputs {
			res.AddGenericInvolvers(berpctypes.MessageInvolvers, input.Address)
		}
		for _, output := range msg.Outputs {
			res.AddGenericInvolvers(berpctypes.MessageInvolvers, output.Address)
		}
		return
	case *crisistypes.MsgVerifyInvariant:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Sender)
		return
	case *disttypes.MsgSetWithdrawAddress:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.DelegatorAddress, msg.WithdrawAddress)
		return
	case *disttypes.MsgWithdrawDelegatorReward:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.DelegatorAddress, msg.ValidatorAddress)
		return
	case *disttypes.MsgWithdrawValidatorCommission:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.ValidatorAddress)
		return
	case *disttypes.MsgFundCommunityPool:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Depositor)
		return
	case *evidencetypes.MsgSubmitEvidence:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Submitter)
		return
	case *govtypesv1.MsgSubmitProposal:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Proposer)
		return
	case *govtypeslegacy.MsgSubmitProposal:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Proposer)
		return
	case *govtypesv1.MsgDeposit:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Depositor)
		return
	case *govtypeslegacy.MsgDeposit:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Depositor)
		return
	case *govtypesv1.MsgVote:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Voter)
		return
	case *govtypeslegacy.MsgVote:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Voter)
		return
	case *govtypesv1.MsgVoteWeighted:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Voter)
		return
	case *govtypeslegacy.MsgVoteWeighted:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Voter)
		return
	case *ibctypes.MsgCreateClient:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *ibctypes.MsgUpdateClient:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *ibctypes.MsgUpgradeClient:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *ibctypes.MsgSubmitMisbehaviour:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *ibctransfertypes.MsgTransfer:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Sender, msg.Receiver)
		return
	case *connectiontypes.MsgConnectionOpenAck:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *connectiontypes.MsgConnectionOpenInit:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *connectiontypes.MsgConnectionOpenConfirm:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *connectiontypes.MsgConnectionOpenTry:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *channeltypes.MsgChannelOpenInit:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *channeltypes.MsgChannelOpenConfirm:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *channeltypes.MsgChannelOpenTry:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *channeltypes.MsgAcknowledgement:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		res.Merge(getInvolversInIbcPacketInfo(msg.Packet))
		return
	case *channeltypes.MsgChannelOpenAck:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		return
	case *channeltypes.MsgRecvPacket:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		res.Merge(getInvolversInIbcPacketInfo(msg.Packet))
		return
	case *channeltypes.MsgTimeout:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		res.Merge(getInvolversInIbcPacketInfo(msg.Packet))
		return
	case *channeltypes.MsgTimeoutOnClose:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Signer)
		res.Merge(getInvolversInIbcPacketInfo(msg.Packet))
		return
	case *slashingtypes.MsgUnjail:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.ValidatorAddr)
		return
	case *stakingtypes.MsgCreateValidator:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.DelegatorAddress, msg.ValidatorAddress)
		return
	case *stakingtypes.MsgEditValidator:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.ValidatorAddress)
		return
	case *stakingtypes.MsgDelegate:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.DelegatorAddress, msg.ValidatorAddress)
		return
	case *stakingtypes.MsgBeginRedelegate:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.DelegatorAddress, msg.ValidatorSrcAddress, msg.ValidatorDstAddress)
		return
	case *stakingtypes.MsgUndelegate:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.DelegatorAddress, msg.ValidatorAddress)
		return
	case *stakingtypes.MsgCancelUnbondingDelegation:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.DelegatorAddress, msg.ValidatorAddress)
		return
	case *authztypes.MsgGrant:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Granter, msg.Grantee)
		return
	case *authztypes.MsgExec:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Grantee)

		if len(msg.Msgs) > 0 {
			for _, authorizedMsgAny := range msg.Msgs {
				var cosmosMsg sdk.Msg
				err := m.clientCtx.Codec.UnpackAny(authorizedMsgAny, &cosmosMsg)
				if err != nil {
					continue
				}
				resChild, err := m.defaultMessageInvolversExtractor(cosmosMsg, tx, tmTx, clientCtx)
				if err != nil {
					continue
				}

				res.Merge(resChild)
			}
		}
		return
	case *authztypes.MsgRevoke:
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, msg.Granter, msg.Grantee)
		return
	default:
		m.GetLogger().Error("missing message involvers extractor", "msg-type", berpcutils.ProtoMessageName(msg))
		resTxResult, errTxResult := clientCtx.Client.Tx(m.ctx, tmTx.Hash(), false)
		if errTxResult != nil {
			return nil, status.Error(
				codes.Internal,
				fmt.Sprintf(
					"failed to get tx result for tx %x when processing msg %s: %s",
					tmTx.Hash(),
					berpcutils.ProtoMessageName(msg),
					errTxResult.Error(),
				),
			)
		}
		for _, event := range resTxResult.TxResult.Events {
			for _, attribute := range event.Attributes {
				if m.bech32Cfg.IsAccountAddr(string(attribute.Value)) {
					res.AddGenericInvolvers(berpctypes.MessageInvolvers, string(attribute.Value))
				}
			}
		}
		return
	}
}

func getInvolversInIbcPacketInfo(packet channeltypes.Packet) (res berpctypes.MessageInvolversResult) {
	res = berpctypes.NewMessageInvolversResult()

	var data ibctransfertypes.FungibleTokenPacketData
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.Data, &data); err == nil {
		res.AddGenericInvolvers(berpctypes.MessageInvolvers, data.Sender, data.Receiver)
	}

	return res
}

func getTxValueInIbcPacketInfo(packet channeltypes.Packet) sdk.Coin {
	var data ibctransfertypes.FungibleTokenPacketData

	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.Data, &data); err == nil && len(data.Denom) > 0 && len(data.Amount) > 0 {
		if amount, ok := sdk.NewIntFromString(data.Amount); ok {
			return sdk.NewCoin(data.Denom, amount)
		}
	}

	return sdk.Coin{}
}

func buildIbcPacketInfoFromPacket(packet channeltypes.Packet, incoming bool) map[string]any {
	return buildIbcPacketInfo(
		fmt.Sprintf("%d", packet.Sequence),
		packet.SourcePort, packet.SourceChannel,
		packet.DestinationPort, packet.DestinationChannel,
		incoming,
	)
}

func buildIbcPacketInfo(sequence, sourcePort, sourceChannel, destPort, destChannel string, incoming bool) map[string]any {
	return map[string]any{
		"sequence":           sequence,
		"sourcePort":         sourcePort,
		"sourceChannel":      sourceChannel,
		"destinationPort":    destPort,
		"destinationChannel": destChannel,
		"incoming":           incoming,
	}
}
