package keeper

import (
	"context"
	"errors"

	"zerotrust/x/zt/types"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) ListGrant(ctx context.Context, req *types.QueryAllGrantRequest) (*types.QueryAllGrantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	grants, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.Grant,
		req.Pagination,
		func(_ string, value types.Grant) (types.Grant, error) {
			return value, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllGrantResponse{Grant: grants, Pagination: pageRes}, nil
}

func (q queryServer) GetGrant(ctx context.Context, req *types.QueryGetGrantRequest) (*types.QueryGetGrantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, err := q.k.Grant.Get(ctx, req.GrantId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetGrantResponse{Grant: val}, nil
}
