package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func newKeeperTestAddress(t *testing.T, f *fixture, marker byte) string {
	t.Helper()

	addressBytes := make([]byte, 20)
	addressBytes[0] = marker

	address, err := f.addressCodec.BytesToString(addressBytes)
	require.NoError(t, err)

	return address
}

func TestMsgRegisterIdentity(t *testing.T) {
	f := initFixture(t)

	admin := newKeeperTestAddress(t, f, 1)

	params := types.DefaultParams()
	params.Admin = admin
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	server := keeper.NewMsgServerImpl(f.keeper)

	t.Run("successful registration", func(t *testing.T) {
		msg := &types.MsgRegisterIdentity{
			Creator: admin,
			Did:     "did:alice",
			Pubkey:  "alice-public-key",
			Role:    "USER",
		}

		res, err := server.RegisterIdentity(f.ctx, msg)

		require.NoError(t, err)
		require.True(t, res.Ok)

		identity, err := f.keeper.Identity.Get(f.ctx, "did:alice")
		require.NoError(t, err)
		require.Equal(t, "did:alice", identity.Did)
		require.Equal(t, "alice-public-key", identity.Pubkey)
		require.Equal(t, "USER", identity.Role)
		require.Equal(t, "ACTIVE", identity.Status)

		record, err := f.keeper.AuditRecord.Get(f.ctx, 0)
		require.NoError(t, err)
		require.Equal(t, "IDENTITY_REGISTER", record.EventType)
		require.Equal(t, "did:alice", record.Subject)
	})

	t.Run("nil message", func(t *testing.T) {
		_, err := server.RegisterIdentity(f.ctx, nil)
		require.Error(t, err)
	})

	t.Run("missing DID", func(t *testing.T) {
		_, err := server.RegisterIdentity(f.ctx, &types.MsgRegisterIdentity{
			Creator: admin,
			Pubkey:  "public-key",
			Role:    "USER",
		})
		require.Error(t, err)
	})

	t.Run("missing public key", func(t *testing.T) {
		_, err := server.RegisterIdentity(f.ctx, &types.MsgRegisterIdentity{
			Creator: admin,
			Did:     "did:bob",
			Role:    "USER",
		})
		require.Error(t, err)
	})

	t.Run("missing role", func(t *testing.T) {
		_, err := server.RegisterIdentity(f.ctx, &types.MsgRegisterIdentity{
			Creator: admin,
			Did:     "did:bob",
			Pubkey:  "public-key",
		})
		require.Error(t, err)
	})

	t.Run("invalid creator address", func(t *testing.T) {
		_, err := server.RegisterIdentity(f.ctx, &types.MsgRegisterIdentity{
			Creator: "invalid-address",
			Did:     "did:bob",
			Pubkey:  "public-key",
			Role:    "USER",
		})
		require.Error(t, err)
	})

	t.Run("non-admin cannot register identity", func(t *testing.T) {
		other := newKeeperTestAddress(t, f, 2)

		_, err := server.RegisterIdentity(f.ctx, &types.MsgRegisterIdentity{
			Creator: other,
			Did:     "did:unauthorized",
			Pubkey:  "public-key",
			Role:    "USER",
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "not admin")
	})

	t.Run("duplicate identity", func(t *testing.T) {
		msg := &types.MsgRegisterIdentity{
			Creator: admin,
			Did:     "did:alice",
			Pubkey:  "another-public-key",
			Role:    "USER",
		}

		_, err := server.RegisterIdentity(f.ctx, msg)

		require.Error(t, err)
	})
}
