package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func TestMsgCreatePolicy(t *testing.T) {
	f := initFixture(t)
	admin := newKeeperTestAddress(t, f, 5)

	params := types.DefaultParams()
	params.Admin = admin
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	server := keeper.NewMsgServerImpl(f.keeper)

	// CreatePolicy is implemented by msgServer but is not in the
	// generated types.MsgServer interface.
	createPolicyServer, ok := server.(interface {
		CreatePolicy(context.Context, *types.MsgCreatePolicy) (*types.MsgCreatePolicyResponse, error)
	})
	require.True(t, ok)

	t.Run("successful creation", func(t *testing.T) {
		res, err := createPolicyServer.CreatePolicy(f.ctx, &types.MsgCreatePolicy{
			Creator:     admin,
			PolicyId:    "policy-1",
			Version:     "v1",
			ContentHash: "hash-1",
		})

		require.NoError(t, err)
		require.True(t, res.Ok)

		policy, err := f.keeper.Policy.Get(f.ctx, "policy-1")
		require.NoError(t, err)

		require.Equal(t, "policy-1", policy.PolicyId)
		require.Equal(t, "v1", policy.Version)
		require.Equal(t, "hash-1", policy.ContentHash)
		require.Equal(t, admin, policy.UpdatedBy)

		record, err := f.keeper.AuditRecord.Get(f.ctx, 0)
		require.NoError(t, err)
		require.Equal(t, "POLICY_CREATE", record.EventType)
	})

	t.Run("nil message", func(t *testing.T) {
		_, err := createPolicyServer.CreatePolicy(f.ctx, nil)
		require.Error(t, err)
	})

	t.Run("missing policy ID", func(t *testing.T) {
		_, err := createPolicyServer.CreatePolicy(f.ctx, &types.MsgCreatePolicy{
			Creator:     admin,
			Version:     "v1",
			ContentHash: "hash",
		})
		require.Error(t, err)
	})

	t.Run("missing version", func(t *testing.T) {
		_, err := createPolicyServer.CreatePolicy(f.ctx, &types.MsgCreatePolicy{
			Creator:     admin,
			PolicyId:    "policy-2",
			ContentHash: "hash",
		})
		require.Error(t, err)
	})

	t.Run("missing content hash", func(t *testing.T) {
		_, err := createPolicyServer.CreatePolicy(f.ctx, &types.MsgCreatePolicy{
			Creator:  admin,
			PolicyId: "policy-2",
			Version:  "v1",
		})
		require.Error(t, err)
	})

	t.Run("invalid creator", func(t *testing.T) {
		_, err := createPolicyServer.CreatePolicy(f.ctx, &types.MsgCreatePolicy{
			Creator:     "invalid-address",
			PolicyId:    "policy-2",
			Version:     "v1",
			ContentHash: "hash",
		})
		require.Error(t, err)
	})

	t.Run("non-admin cannot create policy", func(t *testing.T) {
		other := newKeeperTestAddress(t, f, 6)

		_, err := createPolicyServer.CreatePolicy(f.ctx, &types.MsgCreatePolicy{
			Creator:     other,
			PolicyId:    "policy-unauthorized",
			Version:     "v1",
			ContentHash: "hash",
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "not admin")
	})

	t.Run("duplicate policy", func(t *testing.T) {
		_, err := createPolicyServer.CreatePolicy(f.ctx, &types.MsgCreatePolicy{
			Creator:     admin,
			PolicyId:    "policy-1",
			Version:     "v2",
			ContentHash: "hash-2",
		})

		require.Error(t, err)
	})
}
