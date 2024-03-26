package utils

import (
	abci "github.com/tendermint/tendermint/abci/types"
)

func IsEventTypeWithAllAttributes(event abci.Event, eventType string, attributeKeys ...string) (bool, map[string]string) {
	if event.Type != eventType {
		return false, nil
	}

	allAttrs := make(map[string]string)
	for _, attr := range event.Attributes {
		allAttrs[string(attr.Key)] = string(attr.Value)
	}

	keyToValue := make(map[string]string)
	for _, key := range attributeKeys {
		v, ok := allAttrs[key]
		if !ok {
			return false, nil
		}
		keyToValue[key] = v
	}

	return true, keyToValue
}
