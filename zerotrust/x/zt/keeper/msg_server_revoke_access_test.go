package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func TestMsgRevokeAccess(t *testing.T) {
	f := initFixture(t)

	admin := newKeeperTestAddress(t, f, 11)

	params := types.DefaultParams()
	params.Admin = admin
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	server := keeper.NewMsgServerImpl(f.keeper)

	require.NoError(t, f.keeper.Grant.Set(f.ctx, "grant-1", types.Grant{
		GrantId:   "grant-1",
		Subject:   "did:alice",
		Resource:  "resource-1",
		Action:    "READ",
		Grantor:   admin,
		Status:    "ACTIVE",
		ExpiresAt: 0,
	}))

	t.Run("successful revoke", func(t *testing.T) {
		res, err := server.RevokeAccess(f.ctx, &types.MsgRevokeAccess{
			Creator: admin,
			GrantId: "grant-1",
			Reason:  "permission no longer required",
		})

		require.NoError(t, err)
		require.True(t, res.Ok)

		grant, err := f.keeper.Grant.Get(f.ctx, "grant-1")
		require.NoError(t, err)

		require.Equal(t, "REVOKED", grant.Status)
		require.NotZero(t, grant.RevokedAt)

		record, err := f.keeper.AuditRecord.Get(f.ctx, 0)
		require.NoError(t, err)
		require.Equal(t, "ACCESS_REVOKE", record.EventType)
		require.Equal(t, "REVOKED", record.Decision)
	})

	t.Run("nil message", func(t *testing.T) {
		_, err := server.RevokeAccess(f.ctx, nil)
		require.Error(t, err)
	})

	t.Run("missing creator", func(t *testing.T) {
		_, err := server.RevokeAccess(f.ctx, &types.MsgRevokeAccess{
			GrantId: "grant-1",
		})
		require.Error(t, err)
	})

	t.Run("missing grant ID", func(t *testing.T) {
		_, err := server.RevokeAccess(f.ctx, &types.MsgRevokeAccess{
			Creator: admin,
		})
		require.Error(t, err)
	})

	t.Run("invalid creator", func(t *testing.T) {
		_, err := server.RevokeAccess(f.ctx, &types.MsgRevokeAccess{
			Creator: "invalid-address",
			GrantId: "grant-1",
		})
		require.Error(t, err)
	})

	t.Run("non-admin cannot revoke", func(t *testing.T) {
		other := newKeeperTestAddress(t, f, 12)

		_, err := server.RevokeAccess(f.ctx, &types.MsgRevokeAccess{
			Creator: other,
			GrantId: "grant-1",
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "not admin")
	})

	t.Run("grant does not exist", func(t *testing.T) {
		_, err := server.RevokeAccess(f.ctx, &types.MsgRevokeAccess{
			Creator: admin,
			GrantId: "unknown-grant",
		})
		require.Error(t, err)
	})

	t.Run("grant already revoked", func(t *testing.T) {
		require.NoError(t, f.keeper.Grant.Set(f.ctx, "grant-revoked", types.Grant{
			GrantId:  "grant-revoked",
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "READ",
			Status:   "REVOKED",
		}))

		_, err := server.RevokeAccess(f.ctx, &types.MsgRevokeAccess{
			Creator: admin,
			GrantId: "grant-revoked",
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "is not active")
	})

	t.Run("empty reason is accepted", func(t *testing.T) {
		require.NoError(t, f.keeper.Grant.Set(f.ctx, "grant-no-reason", types.Grant{
			GrantId:  "grant-no-reason",
			Subject:  "did:bob",
			Resource: "resource-2",
			Action:   "WRITE",
			Status:   "ACTIVE",
		}))

		res, err := server.RevokeAccess(f.ctx, &types.MsgRevokeAccess{
			Creator: admin,
			GrantId: "grant-no-reason",
		})

		require.NoError(t, err)
		require.True(t, res.Ok)
	})
}
