package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func TestUpdatePolicyAuthorization(t *testing.T) {
	f := initFixture(t)

	// Tạo hai địa chỉ Bech32 hợp lệ.
	adminBytes := make([]byte, 20)
	adminBytes[0] = 1

	otherBytes := make([]byte, 20)
	otherBytes[0] = 2

	admin, err := f.addressCodec.BytesToString(adminBytes)
	require.NoError(t, err)

	other, err := f.addressCodec.BytesToString(otherBytes)
	require.NoError(t, err)

	// Cấu hình admin.
	params := types.DefaultParams()
	params.Admin = admin

	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	server := keeper.NewMsgServerImpl(f.keeper)

	t.Run("admin is authorized", func(t *testing.T) {
		msg := &types.MsgUpdatePolicy{
			Creator:     admin,
			PolicyId:    "policy-admin",
			Version:     "v1",
			ContentHash: "hash-admin",
		}

		_, err := server.UpdatePolicy(f.ctx, msg)

		// Không được bị từ chối vì không phải admin.
		if err != nil {
			require.NotContains(t, err.Error(), "not admin")
		}
	})

	t.Run("non-admin is unauthorized", func(t *testing.T) {
		msg := &types.MsgUpdatePolicy{
			Creator:     other,
			PolicyId:    "policy-other",
			Version:     "v1",
			ContentHash: "hash-other",
		}

		_, err := server.UpdatePolicy(f.ctx, msg)

		require.Error(t, err)
		require.Contains(t, err.Error(), "not admin")
	})
}
