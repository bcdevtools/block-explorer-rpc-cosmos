package utils

import (
	"context"
	"fmt"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"google.golang.org/grpc/metadata"
)

func QueryContextWithHeight(height int64) context.Context {
	if height == 0 {
		return context.Background()
	}

	return metadata.AppendToOutgoingContext(
		context.Background(),
		grpctypes.GRPCBlockHeightHeader,
		fmt.Sprintf("%d", height),
	)
}
