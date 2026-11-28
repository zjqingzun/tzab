package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func TestQueryCheckAccess(t *testing.T) {
	f := initFixture(t)
	server := keeper.NewQueryServerImpl(f.keeper)

	// Prepare identity and resource.
	require.NoError(t, f.keeper.Identity.Set(f.ctx, "did:alice", types.Identity{
		Did:    "did:alice",
		Pubkey: "alice-public-key",
		Role:   "USER",
		Status: "ACTIVE",
	}))

	require.NoError(t, f.keeper.Resource.Set(f.ctx, "resource-1", types.Resource{
		ResourceId:     "resource-1",
		Owner:          "owner",
		Classification: "CONFIDENTIAL",
	}))

	t.Run("nil request", func(t *testing.T) {
		_, err := server.CheckAccess(f.ctx, nil)
		require.Error(t, err)
	})

	t.Run("missing required fields", func(t *testing.T) {
		_, err := server.CheckAccess(f.ctx, &types.QueryCheckAccessRequest{
			Subject: "did:alice",
		})
		require.Error(t, err)
	})

	t.Run("identity does not exist", func(t *testing.T) {
		res, err := server.CheckAccess(f.ctx, &types.QueryCheckAccessRequest{
			Subject:  "did:unknown",
			Resource: "resource-1",
			Action:   "READ",
		})

		require.NoError(t, err)
		require.False(t, res.Allowed)
		require.Equal(t, "subject identity does not exist", res.Reason)
	})

	t.Run("resource does not exist", func(t *testing.T) {
		res, err := server.CheckAccess(f.ctx, &types.QueryCheckAccessRequest{
			Subject:  "did:alice",
			Resource: "unknown-resource",
			Action:   "READ",
		})

		require.NoError(t, err)
		require.False(t, res.Allowed)
		require.Equal(t, "resource does not exist", res.Reason)
	})

	t.Run("identity is inactive", func(t *testing.T) {
		require.NoError(t, f.keeper.Identity.Set(f.ctx, "did:inactive", types.Identity{
			Did:    "did:inactive",
			Pubkey: "inactive-key",
			Role:   "USER",
			Status: "INACTIVE",
		}))

		res, err := server.CheckAccess(f.ctx, &types.QueryCheckAccessRequest{
			Subject:  "did:inactive",
			Resource: "resource-1",
			Action:   "READ",
		})

		require.NoError(t, err)
		require.False(t, res.Allowed)
		require.Equal(t, "subject identity is not active", res.Reason)
	})

	t.Run("no valid grant", func(t *testing.T) {
		res, err := server.CheckAccess(f.ctx, &types.QueryCheckAccessRequest{
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "READ",
		})

		require.NoError(t, err)
		require.False(t, res.Allowed)
		require.Equal(t, "no valid grant found", res.Reason)
	})

	t.Run("valid grant allows access", func(t *testing.T) {
		require.NoError(t, f.keeper.Grant.Set(f.ctx, "grant-1", types.Grant{
			GrantId:   "grant-1",
			Subject:   "did:alice",
			Resource:  "resource-1",
			Action:    "READ",
			Status:    "ACTIVE",
			ExpiresAt: 0,
		}))

		res, err := server.CheckAccess(f.ctx, &types.QueryCheckAccessRequest{
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "READ",
		})

		require.NoError(t, err)
		require.True(t, res.Allowed)
		require.Equal(t, "access granted", res.Reason)
		require.Equal(t, "grant-1", res.GrantId)
	})

	t.Run("expired grant denies access", func(t *testing.T) {
		sdkCtx := sdk.UnwrapSDKContext(f.ctx)

		require.NoError(t, f.keeper.Grant.Set(f.ctx, "grant-expired", types.Grant{
			GrantId:   "grant-expired",
			Subject:   "did:alice",
			Resource:  "resource-1",
			Action:    "WRITE",
			Status:    "ACTIVE",
			ExpiresAt: sdkCtx.BlockTime().Unix() - 1,
		}))

		res, err := server.CheckAccess(f.ctx, &types.QueryCheckAccessRequest{
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "WRITE",
		})

		require.NoError(t, err)
		require.False(t, res.Allowed)
		require.Equal(t, "no valid grant found", res.Reason)
	})

	t.Run("revoked grant denies access", func(t *testing.T) {
		require.NoError(t, f.keeper.Grant.Set(f.ctx, "grant-revoked", types.Grant{
			GrantId:  "grant-revoked",
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "DELETE",
			Status:   "REVOKED",
		}))

		res, err := server.CheckAccess(f.ctx, &types.QueryCheckAccessRequest{
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "DELETE",
		})

		require.NoError(t, err)
		require.False(t, res.Allowed)
	})
}
