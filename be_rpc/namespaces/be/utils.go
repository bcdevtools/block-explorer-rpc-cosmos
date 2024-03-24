package be

import (
	berpctypes "github.com/bcdevtools/block-explorer-rpc-cosmos/be_rpc/types"
	tmmath "github.com/tendermint/tendermint/libs/math"
)

func getPageNumber(pageNoOptional *int) (int, error) {
	if pageNoOptional == nil {
		return 1, nil
	}

	pageNo := *pageNoOptional

	if pageNo < 0 {
		return 0, berpctypes.ErrBadPageNo
	}

	return tmmath.MaxInt(1, pageNo), nil
}
