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
	MessageSenderSigner InvolversType = "s"
	MessageInvolvers    InvolversType = "0"
	Erc20Involvers      InvolversType = "erc20"
	NftInvolvers        InvolversType = "nft"
)

type MessageInvolversResult interface {
	Merge(other MessageInvolversResult)
	AddGenericInvolvers(t InvolversType, addresses ...string)
	AddContractInvolvers(t InvolversType, contractAddr ContractAddress, addresses ...string)
	Finalize()

	GenericInvolvers() MessageGenericInvolvers
	ContractsInvolvers() MessageContractsInvolvers
}

type MessageGenericInvolvers map[InvolversType][]string

type MessageContractsInvolvers map[InvolversType]MessageContractsInvolversByType

type ContractAddress string
type MessageContractsInvolversByType map[ContractAddress][]string

var _ MessageInvolversResult = &messageInvolversResult{}

type messageInvolversResult struct {
	genericInvolvers  MessageGenericInvolvers
	contractInvolvers MessageContractsInvolvers
}

func NewMessageInvolversResult() MessageInvolversResult {
	return newMessageInvolversResult()
}

func newMessageInvolversResult() *messageInvolversResult {
	return &messageInvolversResult{
		genericInvolvers:  make(MessageGenericInvolvers),
		contractInvolvers: make(MessageContractsInvolvers),
	}
}

func (m *messageInvolversResult) Merge(other MessageInvolversResult) {
	for ivt, ivs := range other.GenericInvolvers() {
		m.genericInvolvers[ivt] = append(m.genericInvolvers[ivt], ivs...)
	}
	for ivt, ivc := range other.ContractsInvolvers() {
		if _, exists := m.contractInvolvers[ivt]; !exists {
			m.contractInvolvers[ivt] = make(MessageContractsInvolversByType)
		}
		for contract, ivs := range ivc {
			m.contractInvolvers[ivt][contract] = append(m.contractInvolvers[ivt][contract], ivs...)
		}
	}
}

func (m *messageInvolversResult) AddGenericInvolvers(t InvolversType, addresses ...string) {
	for _, address := range addresses {
		address := normalizeAddress(address)
		if len(address) == 0 {
			continue
		}

		spl := strings.Split(address, "/") // remove the suffix
		address = spl[0]

		m.genericInvolvers[t] = append(m.genericInvolvers[t], address)
	}
}

func (m *messageInvolversResult) AddContractInvolvers(t InvolversType, contractAddr ContractAddress, addresses ...string) {
	contractAddr = ContractAddress(normalizeAddress(string(contractAddr)))

	var involvedInContract MessageContractsInvolversByType
	if ci, exists := m.contractInvolvers[t]; exists {
		involvedInContract = ci
	} else {
		involvedInContract = make(MessageContractsInvolversByType)
		m.contractInvolvers[t] = involvedInContract
	}

	for _, address := range addresses {
		address := normalizeAddress(address)
		if len(address) == 0 {
			continue
		}

		spl := strings.Split(address, "/") // remove the suffix
		address = spl[0]

		involvedInContract[contractAddr] = append(involvedInContract[contractAddr], address)
	}
}

// Finalize normalize addresses and removes duplicates from the result
func (m *messageInvolversResult) Finalize() {
	if m == nil {
		return
	}

	r := newMessageInvolversResult()
	r.genericInvolvers = func() MessageGenericInvolvers {
		distinctMap := make(MessageGenericInvolvers)
		for ivt, ivs := range m.genericInvolvers {
			unique := make(map[string]bool)
			for _, addr := range ivs {
				unique[normalizeAddress(addr)] = true
			}
			distinctMap[ivt] = make([]string, 0, len(unique))
			for addr := range unique {
				distinctMap[ivt] = append(distinctMap[ivt], addr)
			}
		}
		return distinctMap
	}()
	r.contractInvolvers = func() MessageContractsInvolvers {
		res := make(MessageContractsInvolvers)
		for ivt, ivc := range m.contractInvolvers {
			distinctMap := make(MessageContractsInvolversByType)
			for contract, ivs := range ivc {
				unique := make(map[string]bool)
				for _, addr := range ivs {
					unique[normalizeAddress(addr)] = true
				}
				distinctMap[contract] = make([]string, 0, len(unique))
				for addr := range unique {
					distinctMap[contract] = append(distinctMap[contract], addr)
				}
			}
			res[ivt] = distinctMap
		}
		return res
	}()

	m.genericInvolvers = r.genericInvolvers
	m.contractInvolvers = r.contractInvolvers
}

func (m *messageInvolversResult) GenericInvolvers() MessageGenericInvolvers {
	return m.genericInvolvers
}

func (m *messageInvolversResult) ContractsInvolvers() MessageContractsInvolvers {
	return m.contractInvolvers
}

func normalizeAddress(address string) string {
	return strings.ToLower(strings.TrimSpace(address))
}
