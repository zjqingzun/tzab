package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func TestMsgRegisterResource(t *testing.T) {
	f := initFixture(t)
	server := keeper.NewMsgServerImpl(f.keeper)

	creator := newKeeperTestAddress(t, f, 3)

	t.Run("successful registration", func(t *testing.T) {
		msg := &types.MsgRegisterResource{
			Creator:        creator,
			ResourceId:     "resource-1",
			Classification: "CONFIDENTIAL",
		}

		res, err := server.RegisterResource(f.ctx, msg)

		require.NoError(t, err)
		require.True(t, res.Ok)

		resource, err := f.keeper.Resource.Get(f.ctx, "resource-1")
		require.NoError(t, err)

		require.Equal(t, "resource-1", resource.ResourceId)
		require.Equal(t, creator, resource.Owner)
		require.Equal(t, "CONFIDENTIAL", resource.Classification)

		record, err := f.keeper.AuditRecord.Get(f.ctx, 0)
		require.NoError(t, err)
		require.Equal(t, "RESOURCE_REGISTER", record.EventType)
		require.Equal(t, "resource-1", record.Resource)
	})

	t.Run("nil message", func(t *testing.T) {
		_, err := server.RegisterResource(f.ctx, nil)
		require.Error(t, err)
	})

	t.Run("missing resource ID", func(t *testing.T) {
		_, err := server.RegisterResource(f.ctx, &types.MsgRegisterResource{
			Creator:        creator,
			Classification: "CONFIDENTIAL",
		})
		require.Error(t, err)
	})

	t.Run("missing classification", func(t *testing.T) {
		_, err := server.RegisterResource(f.ctx, &types.MsgRegisterResource{
			Creator:    creator,
			ResourceId: "resource-2",
		})
		require.Error(t, err)
	})

	t.Run("invalid creator", func(t *testing.T) {
		_, err := server.RegisterResource(f.ctx, &types.MsgRegisterResource{
			Creator:        "invalid-address",
			ResourceId:     "resource-2",
			Classification: "PUBLIC",
		})
		require.Error(t, err)
	})

	t.Run("duplicate resource", func(t *testing.T) {
		_, err := server.RegisterResource(f.ctx, &types.MsgRegisterResource{
			Creator:        creator,
			ResourceId:     "resource-1",
			Classification: "PUBLIC",
		})

		require.Error(t, err)
	})

	t.Run("trimmed values", func(t *testing.T) {
		_, err := server.RegisterResource(f.ctx, &types.MsgRegisterResource{
			Creator:        creator,
			ResourceId:     " resource-trim ",
			Classification: " INTERNAL ",
		})

		require.NoError(t, err)

		resource, err := f.keeper.Resource.Get(f.ctx, "resource-trim")
		require.NoError(t, err)
		require.Equal(t, "INTERNAL", resource.Classification)
	})
}
