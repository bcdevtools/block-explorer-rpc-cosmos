package types

import abci "github.com/tendermint/tendermint/abci/types"

type TxEvent struct {
	Type       string             `json:"type,omitempty"`
	Attributes []TxEventAttribute `json:"attributes,omitempty"`
}

type TxEventAttribute struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func ConvertTxEvent(events []abci.Event) []TxEvent {
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
