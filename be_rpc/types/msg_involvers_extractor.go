package types

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	tmtypes "github.com/tendermint/tendermint/types"
	"strings"
)

type MessageInvolversExtractor func(msg sdk.Msg, tx *tx.Tx, tmTx tmtypes.Tx, clientCtx client.Context) (MessageInvolversResult, error)

type InvolversType string

const (
	MessageInvolvers InvolversType = "0"
	Erc20Involvers   InvolversType = "erc20"
	NftInvolvers     InvolversType = "nft"
)

type MessageInvolversResult map[InvolversType][]string

func (m MessageInvolversResult) Merge(other MessageInvolversResult) MessageInvolversResult {
	for k, v := range other {
		m[k] = append(m[k], v...)
	}
	return m
}

func (m MessageInvolversResult) Add(t InvolversType, addresses ...string) {
	for _, address := range addresses {
		address := normalizeAddress(address)
		if len(address) == 0 {
			continue
		}

		spl := strings.Split(address, "/") // remove the suffix
		address = spl[0]

		if _, found := m[t]; found {
			m[t] = append(m[t], address)
		} else {
			m[t] = []string{address}
		}
	}
}

// Finalize normalize addresses and removes duplicates from the result
func (m MessageInvolversResult) Finalize() MessageInvolversResult {
	r := make(MessageInvolversResult)
	for k, v := range m {
		unique := make(map[string]bool)
		for _, addr := range v {
			unique[normalizeAddress(addr)] = true
		}
		r[k] = make([]string, 0, len(unique))
		for addr := range unique {
			r[k] = append(r[k], addr)
		}
	}
	return r
}

func normalizeAddress(address string) string {
	return strings.ToLower(strings.TrimSpace(address))
}
