package types

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// FakeBaseAccount is fake version of x/auth BaseAccount. With ability to marshal and unmarshal.
type FakeBaseAccount struct {
	Address       string `json:"address"`
	AccountNumber uint64 `json:"account_number,string"`
	Sequence      uint64 `json:"sequence,string"`
}

func (m *FakeBaseAccount) TryUnmarshalFromProto(protoAccount *codectypes.Any, codec codec.Codec) (success bool, err error) {
	type genericWrappedAccount struct {
		BaseAccount    *FakeBaseAccount       `json:"base_account,omitempty"`
		VestingAccount *genericWrappedAccount `json:"base_vesting_account,omitempty"`
	}

	bz, err := codec.MarshalJSON(protoAccount)
	if err != nil {
		return false, err
	}

	var wrappedAccount genericWrappedAccount
	err = json.Unmarshal(bz, &wrappedAccount)
	if err != nil {
		return false, err
	}

	if wrappedAccount.BaseAccount != nil && wrappedAccount.BaseAccount.Address != "" {

		m.Address = wrappedAccount.BaseAccount.Address
		m.AccountNumber = wrappedAccount.BaseAccount.AccountNumber
		m.Sequence = wrappedAccount.BaseAccount.Sequence

		return true, nil
	} else if wrappedAccount.VestingAccount != nil &&
		wrappedAccount.VestingAccount.BaseAccount != nil &&
		wrappedAccount.VestingAccount.BaseAccount.Address != "" {

		m.Address = wrappedAccount.VestingAccount.BaseAccount.Address
		m.AccountNumber = wrappedAccount.VestingAccount.BaseAccount.AccountNumber
		m.Sequence = wrappedAccount.VestingAccount.BaseAccount.Sequence

		return true, nil
	}

	// try again with the base account
	var baseAccount FakeBaseAccount
	err = json.Unmarshal(bz, &baseAccount)
	if err != nil {
		return false, err
	}

	if baseAccount.Address != "" {
		m.Address = baseAccount.Address
		m.AccountNumber = baseAccount.AccountNumber
		m.Sequence = baseAccount.Sequence

		return true, nil
	}

	return false, nil
}
