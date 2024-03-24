package backend

import (
	"github.com/cosmos/cosmos-sdk/types/query"
)

const defaultPageSize = 20

func getDefaultPagination(pageNo int) *query.PageRequest {
	return &query.PageRequest{
		Offset:  uint64(defaultPageSize * (pageNo - 1)),
		Limit:   defaultPageSize,
		Reverse: true,
	}
}
