package keeper

import (
	"context"

	"zerotrust/x/zt/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) CheckAccess(ctx context.Context, req *types.QueryCheckAccessRequest) (*types.QueryCheckAccessResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Process the query

	return &types.QueryCheckAccessResponse{}, nil
}
