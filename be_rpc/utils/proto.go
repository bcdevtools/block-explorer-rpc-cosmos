package utils

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func FromAnyToJsonMap(protoAny *codectypes.Any, codec codec.Codec) (map[string]any, error) {
	bz, err := codec.MarshalJSON(protoAny)
	if err != nil {
		return nil, err
	}
	res := make(map[string]any)
	err = json.Unmarshal(bz, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
