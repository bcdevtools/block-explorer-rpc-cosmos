package types

import banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

type RpcDenomMetadata struct {
	// denom_units represents the list of DenomUnit's for a given coin
	DenomUnits []RpcDenomMetadataUnit `json:"denom_units,omitempty"`
	// base represents the base denom (should be the DenomUnit with exponent = 0).
	Base string `json:"base"`
	// name defines the name of the token (eg: Cosmos Atom)
	Name string `json:"name"`
	// symbol is the token symbol usually shown on exchanges (eg: ATOM). This can
	// be the same as the display.
	Symbol string `json:"symbol"`
}

type RpcDenomMetadataUnit struct {
	// denom represents the string name of the given denom unit (e.g uatom).
	Denom string `json:"denom,omitempty"`
	// exponent represents power of 10 exponent that one must
	// raise the base_denom to in order to equal the given DenomUnit's denom
	// 1 denom = 10^exponent base_denom
	// (e.g. with a base_denom of uatom, one can create a DenomUnit of 'atom' with
	// exponent = 6, thus: 1 atom = 10^6 uatom).
	Exponent uint32 `json:"exponent"`
}

func NewRpcDenomMetadataFromBankMetadata(metadata banktypes.Metadata) RpcDenomMetadata {
	denomUnits := make([]RpcDenomMetadataUnit, len(metadata.DenomUnits))
	for i, unit := range metadata.DenomUnits {
		denomUnits[i] = RpcDenomMetadataUnit{
			Denom:    unit.Denom,
			Exponent: unit.Exponent,
		}
	}

	res := RpcDenomMetadata{
		DenomUnits: denomUnits,
		Base:       metadata.Base,
		Symbol:     metadata.Symbol,
	}

	if metadata.Display != "" {
		res.Name = metadata.Display
	} else {
		res.Name = metadata.Name
	}

	return res
}
