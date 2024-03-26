package types

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrBadRequest              = status.Error(codes.InvalidArgument, "bad request")
	ErrBadPageSize             = status.Error(codes.InvalidArgument, "bad page size")
	ErrBadPageNo               = status.Error(codes.InvalidArgument, "bad page number")
	ErrNotSupportedMessageType = status.Error(codes.Unimplemented, "message parser not found")
	ErrBadAddress              = status.Error(codes.InvalidArgument, "bad address")
)
