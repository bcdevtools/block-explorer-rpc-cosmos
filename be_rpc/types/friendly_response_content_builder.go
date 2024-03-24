package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"regexp"
	"strings"
)

type FriendlyResponseContentBuilderI interface {
	WriteText(string) FriendlyResponseContentBuilderI
	WriteAddress(string) FriendlyResponseContentBuilderI
	WriteCoins(coins sdk.Coins, denomsMetadata map[string]banktypes.Metadata) FriendlyResponseContentBuilderI

	Build() (friendlySimple string, friendlyMachine string)
	BuildIntoResponse(res GenericBackendResponse)
}

var _ FriendlyResponseContentBuilderI = &friendlyResponseContentBuilder{}

type friendlyResponseContentBuilder struct {
	// friendlySimple is the simple friendly response content, without ability to inject HTML code.
	friendlySimple strings.Builder

	// friendlyMachine is the friendly response content, with ability to inject HTML using pattern replacement.
	friendlyMachine strings.Builder
}

func NewFriendlyResponseContentBuilder() FriendlyResponseContentBuilderI {
	return &friendlyResponseContentBuilder{
		friendlySimple:  strings.Builder{},
		friendlyMachine: strings.Builder{},
	}
}

func (f *friendlyResponseContentBuilder) WriteText(s string) FriendlyResponseContentBuilderI {
	return f.addBoth(s)
}

var regexAlphaNumericOnly = regexp.MustCompile(`^[a-zA-Z\d]+$`)

func (f *friendlyResponseContentBuilder) WriteAddress(a string) FriendlyResponseContentBuilderI {
	f.friendlySimple.WriteString(a)
	if regexAlphaNumericOnly.MatchString(a) { // only write pattern if content is sanitized
		f.addMachinePattern("address", a)
	} else {
		f.friendlyMachine.WriteString(a)
	}
	return f
}

func (f *friendlyResponseContentBuilder) WriteCoins(coins sdk.Coins, denomsMetadata map[string]banktypes.Metadata) FriendlyResponseContentBuilderI {
	for i, coin := range coins {
		if i > 0 {
			f.addBoth(", ")
		}
		metadata, found := denomsMetadata[coin.Denom]
		if !found || !isTextSatisfyCoinDenom(coin.Denom) {
			f.addBoth("(raw) ").addBoth(coin.String())
			continue
		}
		display, highestExponent := getHighestExponent(metadata)
		if highestExponent == 0 || display == coin.Denom {
			f.addBoth("(raw) ").addBoth(coin.String())
			continue
		}
		if coin.Amount.IsZero() {
			f.addBoth("0 ").addBoth(display)
			continue
		}

		f.addBoth(getDisplayNumber(coin.Amount, highestExponent)).addBoth(" ").addBoth(display)
	}
	return f
}

func (f *friendlyResponseContentBuilder) Build() (friendlySimple string, friendlyMachine string) {
	friendlySimple = f.friendlySimple.String()
	friendlyMachine = f.friendlyMachine.String()
	return
}

func (f *friendlyResponseContentBuilder) BuildIntoResponse(res GenericBackendResponse) {
	friendlySimple, friendlyMachine := f.Build()
	res["cts"] = friendlySimple
	res["ctm"] = friendlyMachine
	return
}

func (f *friendlyResponseContentBuilder) addMachinePattern(_type, content string) {
	f.friendlyMachine.WriteString("{[{ .[")
	f.friendlyMachine.WriteString(_type)
	f.friendlyMachine.WriteString("].[")
	f.friendlyMachine.WriteString(content)
	f.friendlyMachine.WriteString("]. }]}")
}

func (f *friendlyResponseContentBuilder) addBoth(s string) *friendlyResponseContentBuilder {
	f.friendlySimple.WriteString(s)
	f.friendlyMachine.WriteString(s)
	return f
}

var regexCoinDenom = regexp.MustCompile(`^([a-zA-Z\d]+/)?[a-zA-Z\d]+$`)

func isTextSatisfyCoinDenom(s string) bool {
	return regexCoinDenom.MatchString(s)
}

func getHighestExponent(denomMetadata banktypes.Metadata) (displayOfHighest string, highestExponent uint32) {
	for _, denomUnit := range denomMetadata.DenomUnits {
		if denomUnit.Exponent > highestExponent {
			displayOfHighest = denomUnit.Denom
			highestExponent = denomUnit.Exponent
		}
	}

	return
}

func getDisplayNumber(num math.Int, exponent uint32) string {
	if num.IsZero() {
		return "0"
	}

	str := math.LegacyNewDecFromIntWithPrec(num, int64(exponent)).String()
	spl := strings.Split(str, ".")

	if len(spl) == 1 {
		return str
	}

	for strings.HasSuffix(spl[1], "0") {
		spl[1] = spl[1][:len(spl[1])-1]
	}

	if len(spl[1]) == 0 {
		return spl[0]
	}

	return spl[0] + "." + spl[1]
}
