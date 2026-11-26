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

func createNResource(keeper keeper.Keeper, ctx context.Context, n int) []types.Resource {
	items := make([]types.Resource, n)
	for i := range items {
		items[i].ResourceId = strconv.Itoa(i)
		items[i].Owner = strconv.Itoa(i)
		items[i].Classification = strconv.Itoa(i)
		_ = keeper.Resource.Set(ctx, items[i].ResourceId, items[i])
	}
	return items
}

func TestResourceQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNResource(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetResourceRequest
		response *types.QueryGetResourceResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetResourceRequest{
				ResourceId: msgs[0].ResourceId,
			},
			response: &types.QueryGetResourceResponse{Resource: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetResourceRequest{
				ResourceId: msgs[1].ResourceId,
			},
			response: &types.QueryGetResourceResponse{Resource: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetResourceRequest{
				ResourceId: strconv.Itoa(100000),
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
			response, err := qs.GetResource(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestResourceQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNResource(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllResourceRequest {
		return &types.QueryAllResourceRequest{
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
			resp, err := qs.ListResource(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Resource), step)
			require.Subset(t, msgs, resp.Resource)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListResource(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Resource), step)
			require.Subset(t, msgs, resp.Resource)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListResource(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.Resource)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListResource(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
