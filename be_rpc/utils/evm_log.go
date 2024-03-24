package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func IsEvmEventMatch(
	topics []common.Hash, data []byte,
	wantTopicSize int,
	wantTopic0 common.Hash,
	wantTopic1MustBeAddress, wantTopic2MustBeAddress, wantTopic3MustBeAddress, wantData bool,
) bool {
	if len(topics) != wantTopicSize {
		return false
	}
	if (len(data) > 0) != wantData {
		return false
	}
	if topics[0] != wantTopic0 {
		return false
	}

	isTopicAddress := func(topic common.Hash) bool {
		bz := topic.Bytes()
		return new(big.Int).SetBytes(bz[:12]).Sign() == 0
	}

	if wantTopic1MustBeAddress {
		if len(topics) < 2 {
			panic("wrong number of topics, expected at least 2")
		}
		if !isTopicAddress(topics[1]) {
			return false
		}
	}

	if wantTopic2MustBeAddress {
		if len(topics) < 3 {
			panic("wrong number of topics, expected at least 3")
		}
		if !isTopicAddress(topics[2]) {
			return false
		}
	}

	if wantTopic3MustBeAddress {
		if len(topics) < 4 {
			panic("wrong number of topics, expected 4")
		}
		if !isTopicAddress(topics[3]) {
			return false
		}
	}

	return true
}
