package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"strings"
)

type TxEvents []TxEvent

type TxEvent struct {
	Type       string             `json:"type,omitempty"`
	Attributes []TxEventAttribute `json:"attributes,omitempty"`
}

type TxEventAttribute struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func ConvertTxEvent(events []abci.Event) TxEvents {
	var res []TxEvent
	for _, event := range events {
		var attributes []TxEventAttribute
		for _, attr := range event.Attributes {
			attributes = append(attributes, TxEventAttribute{
				Key:   string(attr.Key),
				Value: string(attr.Value),
			})
		}
		res = append(res, TxEvent{
			Type:       event.Type,
			Attributes: attributes,
		})
	}
	return res
}

func (m TxEvents) RemoveUnnecessaryEvmTxEvents() TxEvents {
	remove := func() TxEvents {
		txEventsTruncatedEvm := make([]TxEvent, 0)
		for _, event := range m {
			// remove unnecessary events

			if event.Type == EventTypeEthereumTx {
				continue
			}

			if event.Type == EventTypeTxLog {
				continue
			}

			if event.Type == sdk.EventTypeMessage {
				var ignore bool
				for _, attribute := range event.Attributes {
					if attribute.Key == sdk.AttributeKeyModule && attribute.Value == AttributeValueCategory {
						ignore = true
						break
					} else if attribute.Key == sdk.AttributeKeyAction && strings.HasSuffix(attribute.Value, "MsgEthereumTx") {
						ignore = true
						break
					}
				}
				if ignore {
					continue
				}
			}

			txEventsTruncatedEvm = append(txEventsTruncatedEvm, event)
		}

		return txEventsTruncatedEvm
	}

	for _, event := range m {
		if event.Type == sdk.EventTypeMessage {
			for _, attribute := range event.Attributes {
				if attribute.Key == sdk.AttributeKeyAction && strings.HasSuffix(attribute.Value, "MsgEthereumTx") {
					return remove()
				}
			}
		}
	}

	return m
}
