package keeper_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func createNGrant(keeper keeper.Keeper, ctx context.Context, n int) []types.Grant {
	items := make([]types.Grant, n)
	for i := range items {
		items[i].GrantId = strconv.Itoa(i)
		items[i].Subject = strconv.Itoa(i)
		items[i].Resource = strconv.Itoa(i)
		items[i].Action = strconv.Itoa(i)
		items[i].Grantor = strconv.Itoa(i)
		items[i].Status = strconv.Itoa(i)
		items[i].ExpiresAt = int64(i)
		items[i].CreatedAt = int64(i)
		items[i].RevokedAt = int64(i)
		_ = keeper.Grant.Set(ctx, items[i].GrantId, items[i])
	}
	return items
}

func TestGrantQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNGrant(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetGrantRequest
		response *types.QueryGetGrantResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetGrantRequest{
				GrantId: msgs[0].GrantId,
			},
			response: &types.QueryGetGrantResponse{Grant: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetGrantRequest{
				GrantId: msgs[1].GrantId,
			},
			response: &types.QueryGetGrantResponse{Grant: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetGrantRequest{
				GrantId: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := qs.GetGrant(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestGrantQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNGrant(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllGrantRequest {
		return &types.QueryAllGrantRequest{
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
			resp, err := qs.ListGrant(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Grant), step)
			require.Subset(t, msgs, resp.Grant)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListGrant(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Grant), step)
			require.Subset(t, msgs, resp.Grant)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListGrant(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.Grant)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListGrant(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
