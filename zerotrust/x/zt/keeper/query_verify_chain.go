package keeper

import (
	"context"

	"zerotrust/x/zt/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) VerifyChain(ctx context.Context, req *types.QueryVerifyChainRequest) (*types.QueryVerifyChainResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// TODO: Process the query

	return &types.QueryVerifyChainResponse{}, nil
}
