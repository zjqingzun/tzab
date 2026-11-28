package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func TestMsgGrantAccess(t *testing.T) {
	f := initFixture(t)

	admin := newKeeperTestAddress(t, f, 9)

	params := types.DefaultParams()
	params.Admin = admin
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	server := keeper.NewMsgServerImpl(f.keeper)
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)

	require.NoError(t, f.keeper.Identity.Set(f.ctx, "did:alice", types.Identity{
		Did:    "did:alice",
		Pubkey: "alice-key",
		Role:   "USER",
		Status: "ACTIVE",
	}))

	require.NoError(t, f.keeper.Resource.Set(f.ctx, "resource-1", types.Resource{
		ResourceId:     "resource-1",
		Owner:          admin,
		Classification: "CONFIDENTIAL",
	}))

	t.Run("successful grant", func(t *testing.T) {
		res, err := server.GrantAccess(f.ctx, &types.MsgGrantAccess{
			Creator:   admin,
			GrantId:   "grant-1",
			Subject:   "did:alice",
			Resource:  "resource-1",
			Action:    "READ",
			ExpiresAt: 0,
		})

		require.NoError(t, err)
		require.NotNil(t, res)

		grant, err := f.keeper.Grant.Get(f.ctx, "grant-1")
		require.NoError(t, err)

		require.Equal(t, "grant-1", grant.GrantId)
		require.Equal(t, "did:alice", grant.Subject)
		require.Equal(t, "resource-1", grant.Resource)
		require.Equal(t, "READ", grant.Action)
		require.Equal(t, admin, grant.Grantor)
		require.Equal(t, "ACTIVE", grant.Status)
		require.Zero(t, grant.ExpiresAt)

		record, err := f.keeper.AuditRecord.Get(f.ctx, 0)
		require.NoError(t, err)
		require.Equal(t, "ACCESS_GRANT", record.EventType)
	})

	t.Run("nil message", func(t *testing.T) {
		_, err := server.GrantAccess(f.ctx, nil)
		require.Error(t, err)
	})

	t.Run("missing required fields", func(t *testing.T) {
		_, err := server.GrantAccess(f.ctx, &types.MsgGrantAccess{
			Creator: admin,
			Subject: "did:alice",
		})
		require.Error(t, err)
	})

	t.Run("invalid creator", func(t *testing.T) {
		_, err := server.GrantAccess(f.ctx, &types.MsgGrantAccess{
			Creator:  "invalid-address",
			GrantId:  "grant-invalid",
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "READ",
		})
		require.Error(t, err)
	})

	t.Run("non-admin cannot grant access", func(t *testing.T) {
		other := newKeeperTestAddress(t, f, 10)

		_, err := server.GrantAccess(f.ctx, &types.MsgGrantAccess{
			Creator:  other,
			GrantId:  "grant-unauthorized",
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "WRITE",
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "not admin")
	})

	t.Run("negative expiration", func(t *testing.T) {
		_, err := server.GrantAccess(f.ctx, &types.MsgGrantAccess{
			Creator:   admin,
			GrantId:   "grant-negative",
			Subject:   "did:alice",
			Resource:  "resource-1",
			Action:    "WRITE",
			ExpiresAt: -1,
		})
		require.Error(t, err)
	})

	t.Run("expiration in the past", func(t *testing.T) {
		_, err := server.GrantAccess(f.ctx, &types.MsgGrantAccess{
			Creator:   admin,
			GrantId:   "grant-past",
			Subject:   "did:alice",
			Resource:  "resource-1",
			Action:    "WRITE",
			ExpiresAt: sdkCtx.BlockTime().Unix() - 1,
		})
		require.Error(t, err)
	})

	t.Run("identity does not exist", func(t *testing.T) {
		_, err := server.GrantAccess(f.ctx, &types.MsgGrantAccess{
			Creator:  admin,
			GrantId:  "grant-no-identity",
			Subject:  "did:unknown",
			Resource: "resource-1",
			Action:   "READ",
		})
		require.Error(t, err)
	})

	t.Run("inactive identity", func(t *testing.T) {
		require.NoError(t, f.keeper.Identity.Set(f.ctx, "did:inactive", types.Identity{
			Did:    "did:inactive",
			Pubkey: "inactive-key",
			Role:   "USER",
			Status: "INACTIVE",
		}))

		_, err := server.GrantAccess(f.ctx, &types.MsgGrantAccess{
			Creator:  admin,
			GrantId:  "grant-inactive",
			Subject:  "did:inactive",
			Resource: "resource-1",
			Action:   "READ",
		})
		require.Error(t, err)
	})

	t.Run("resource does not exist", func(t *testing.T) {
		_, err := server.GrantAccess(f.ctx, &types.MsgGrantAccess{
			Creator:  admin,
			GrantId:  "grant-no-resource",
			Subject:  "did:alice",
			Resource: "unknown-resource",
			Action:   "READ",
		})
		require.Error(t, err)
	})

	t.Run("duplicate grant ID", func(t *testing.T) {
		_, err := server.GrantAccess(f.ctx, &types.MsgGrantAccess{
			Creator:  admin,
			GrantId:  "grant-1",
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "WRITE",
		})
		require.Error(t, err)
	})

	t.Run("duplicate active permission", func(t *testing.T) {
		_, err := server.GrantAccess(f.ctx, &types.MsgGrantAccess{
			Creator:  admin,
			GrantId:  "grant-duplicate",
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "READ",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "active grant already exists")
	})
}
