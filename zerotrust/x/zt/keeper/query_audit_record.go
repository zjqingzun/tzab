package keeper

import (
	"context"
	"errors"

	"zerotrust/x/zt/types"

	"cosmossdk.io/collections"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) ListAuditRecord(ctx context.Context, req *types.QueryAllAuditRecordRequest) (*types.QueryAllAuditRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	auditRecords, pageRes, err := query.CollectionPaginate(
		ctx,
		q.k.AuditRecord,
		req.Pagination,
		func(_ uint64, value types.AuditRecord) (types.AuditRecord, error) {
			return value, nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllAuditRecordResponse{AuditRecord: auditRecords, Pagination: pageRes}, nil
}

func (q queryServer) GetAuditRecord(ctx context.Context, req *types.QueryGetAuditRecordRequest) (*types.QueryGetAuditRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	auditRecord, err := q.k.AuditRecord.Get(ctx, req.Id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, sdkerrors.ErrKeyNotFound
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &types.QueryGetAuditRecordResponse{AuditRecord: auditRecord}, nil
}
