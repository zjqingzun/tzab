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

func createNIdentity(keeper keeper.Keeper, ctx context.Context, n int) []types.Identity {
	items := make([]types.Identity, n)
	for i := range items {
		items[i].Did = strconv.Itoa(i)
		items[i].Pubkey = strconv.Itoa(i)
		items[i].Role = strconv.Itoa(i)
		items[i].Status = strconv.Itoa(i)
		items[i].CreatedAt = int64(i)
		_ = keeper.Identity.Set(ctx, items[i].Did, items[i])
	}
	return items
}

func TestIdentityQuerySingle(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNIdentity(f.keeper, f.ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetIdentityRequest
		response *types.QueryGetIdentityResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetIdentityRequest{
				Did: msgs[0].Did,
			},
			response: &types.QueryGetIdentityResponse{Identity: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetIdentityRequest{
				Did: msgs[1].Did,
			},
			response: &types.QueryGetIdentityResponse{Identity: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetIdentityRequest{
				Did: strconv.Itoa(100000),
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
			response, err := qs.GetIdentity(f.ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.EqualExportedValues(t, tc.response, response)
			}
		})
	}
}

func TestIdentityQueryPaginated(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)
	msgs := createNIdentity(f.keeper, f.ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllIdentityRequest {
		return &types.QueryAllIdentityRequest{
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
			resp, err := qs.ListIdentity(f.ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Identity), step)
			require.Subset(t, msgs, resp.Identity)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := qs.ListIdentity(f.ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Identity), step)
			require.Subset(t, msgs, resp.Identity)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := qs.ListIdentity(f.ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.EqualExportedValues(t, msgs, resp.Identity)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := qs.ListIdentity(f.ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
