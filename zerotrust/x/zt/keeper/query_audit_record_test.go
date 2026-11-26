package keeper_test

import (
	"context"
	"strconv"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func createNAuditRecord(keeper keeper.Keeper, ctx context.Context, n int) []types.AuditRecord {
	items := make([]types.AuditRecord, n)
	for i := range items {
		iu := uint64(i)
		items[i].Id = iu
		items[i].EventType = strconv.Itoa(i)
		items[i].Subject = strconv.Itoa(i)
		items[i].Resource = strconv.Itoa(i)
		items[i].Action = strconv.Itoa(i)
		items[i].Decision = strconv.Itoa(i)
		items[i].Reason = strconv.Itoa(i)
		items[i].Actor = strconv.Itoa(i)
		items[i].PayloadHash = strconv.Itoa(i)
		items[i].PrevHash = strconv.Itoa(i)
		items[i].RecordHash = strconv.Itoa(i)
		items[i].BlockHeight = int64(i)
		items[i].Timestamp = int64(i)
		_ = keeper.AuditRecord.Set(ctx, iu, items[i])
		_ = keeper.AuditRecordSeq.Set(ctx, iu)
	}
	return items
}

func TestAuditRecordQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNAuditRecord(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetAuditRecordRequest
		response *types.QueryGetAuditRecordResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetAuditRecordRequest{Id: msgs[0].Id},
			response: &types.QueryGetAuditRecordResponse{AuditRecord: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetAuditRecordRequest{Id: msgs[1].Id},
			response: &types.QueryGetAuditRecordResponse{AuditRecord: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetAuditRecordRequest{Id: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := qs.GetAuditRecord(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestAuditRecordQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNAuditRecord(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllAuditRecordRequest {
		return &types.QueryAllAuditRecordRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListAuditRecord(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.AuditRecord), step)
			require.Subset(t, msgs, resp.AuditRecord)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListAuditRecord(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.AuditRecord), step)
			require.Subset(t, msgs, resp.AuditRecord)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListAuditRecord(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.AuditRecord)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListAuditRecord(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
