package utils

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cmcryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcrypto "github.com/tendermint/tendermint/crypto"
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

func FromAnyPubKeyToConsensusAddress(protoAny *codectypes.Any, codec codec.Codec) (consAddr sdk.ConsAddress, success bool) {
	var err error

	var cosmosPubKey cmcryptotypes.PubKey
	err = codec.UnpackAny(protoAny, &cosmosPubKey)
	if err == nil {
		return sdk.ConsAddress(cosmosPubKey.Address()), true
	}

	var tmPubKey tmcrypto.PubKey
	err = codec.UnpackAny(protoAny, &tmPubKey)
	if err == nil {
		return sdk.ConsAddress(tmPubKey.Address()), true
	}

	return nil, false
}
