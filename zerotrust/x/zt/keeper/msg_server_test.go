package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func TestNewMsgServerImpl(t *testing.T) {
	f := initFixture(t)

	server := keeper.NewMsgServerImpl(f.keeper)

	require.NotNil(t, server)

	// Kiểm tra implementation thỏa mãn MsgServer interface.
	var _ types.MsgServer = server
}

func TestMsgServerNilRequest(t *testing.T) {
	f := initFixture(t)

	server := keeper.NewMsgServerImpl(f.keeper)

	t.Run("UpdatePolicy", func(t *testing.T) {
		_, err := server.UpdatePolicy(f.ctx, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "message cannot be nil")
	})

	t.Run("RegisterIdentity", func(t *testing.T) {
		_, err := server.RegisterIdentity(f.ctx, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "message cannot be nil")
	})

	t.Run("RegisterResource", func(t *testing.T) {
		_, err := server.RegisterResource(f.ctx, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "message cannot be nil")
	})

	t.Run("GrantAccess", func(t *testing.T) {
		_, err := server.GrantAccess(f.ctx, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "message cannot be nil")
	})

	t.Run("RevokeAccess", func(t *testing.T) {
		_, err := server.RevokeAccess(f.ctx, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "message cannot be nil")
	})

	t.Run("RecordAccess", func(t *testing.T) {
		_, err := server.RecordAccess(f.ctx, nil)

		require.Error(t, err)
		require.Contains(t, err.Error(), "message cannot be nil")
	})
}
