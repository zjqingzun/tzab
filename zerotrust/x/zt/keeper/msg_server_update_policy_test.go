package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func TestMsgUpdatePolicy(t *testing.T) {
	f := initFixture(t)
	admin := newKeeperTestAddress(t, f, 7)

	params := types.DefaultParams()
	params.Admin = admin
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	server := keeper.NewMsgServerImpl(f.keeper)

	require.NoError(t, f.keeper.Policy.Set(f.ctx, "policy-1", types.Policy{
		PolicyId:    "policy-1",
		Version:     "v1",
		ContentHash: "old-hash",
		UpdatedBy:   admin,
	}))

	t.Run("successful update", func(t *testing.T) {
		res, err := server.UpdatePolicy(f.ctx, &types.MsgUpdatePolicy{
			Creator:     admin,
			PolicyId:    "policy-1",
			Version:     "v2",
			ContentHash: "new-hash",
		})

		require.NoError(t, err)
		require.True(t, res.Ok)

		policy, err := f.keeper.Policy.Get(f.ctx, "policy-1")
		require.NoError(t, err)

		require.Equal(t, "v2", policy.Version)
		require.Equal(t, "new-hash", policy.ContentHash)
		require.Equal(t, admin, policy.UpdatedBy)

		record, err := f.keeper.AuditRecord.Get(f.ctx, 0)
		require.NoError(t, err)
		require.Equal(t, "POLICY_UPDATE", record.EventType)
	})

	t.Run("nil message", func(t *testing.T) {
		_, err := server.UpdatePolicy(f.ctx, nil)
		require.Error(t, err)
	})

	t.Run("missing policy ID", func(t *testing.T) {
		_, err := server.UpdatePolicy(f.ctx, &types.MsgUpdatePolicy{
			Creator:     admin,
			Version:     "v2",
			ContentHash: "hash",
		})
		require.Error(t, err)
	})

	t.Run("missing version", func(t *testing.T) {
		_, err := server.UpdatePolicy(f.ctx, &types.MsgUpdatePolicy{
			Creator:     admin,
			PolicyId:    "policy-1",
			ContentHash: "hash",
		})
		require.Error(t, err)
	})

	t.Run("missing content hash", func(t *testing.T) {
		_, err := server.UpdatePolicy(f.ctx, &types.MsgUpdatePolicy{
			Creator:  admin,
			PolicyId: "policy-1",
			Version:  "v2",
		})
		require.Error(t, err)
	})

	t.Run("invalid creator", func(t *testing.T) {
		_, err := server.UpdatePolicy(f.ctx, &types.MsgUpdatePolicy{
			Creator:     "invalid-address",
			PolicyId:    "policy-1",
			Version:     "v2",
			ContentHash: "hash",
		})
		require.Error(t, err)
	})

	t.Run("non-admin cannot update policy", func(t *testing.T) {
		other := newKeeperTestAddress(t, f, 8)

		_, err := server.UpdatePolicy(f.ctx, &types.MsgUpdatePolicy{
			Creator:     other,
			PolicyId:    "policy-1",
			Version:     "v3",
			ContentHash: "hash",
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "not admin")
	})

	t.Run("policy does not exist", func(t *testing.T) {
		_, err := server.UpdatePolicy(f.ctx, &types.MsgUpdatePolicy{
			Creator:     admin,
			PolicyId:    "unknown-policy",
			Version:     "v1",
			ContentHash: "hash",
		})

		require.Error(t, err)
	})
}
