package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func TestQueryVerifyChain(t *testing.T) {
	f := initFixture(t)
	server := keeper.NewQueryServerImpl(f.keeper)
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)

	appendRecord := func(event string) {
		t.Helper()

		_, err := f.keeper.AppendAudit(sdkCtx, types.AuditRecord{
			EventType: event,
			Subject:   "did:alice",
			Action:    "READ",
			Decision:  "ALLOW",
			Actor:     "test",
		})
		require.NoError(t, err)
	}

	t.Run("nil request", func(t *testing.T) {
		_, err := server.VerifyChain(f.ctx, nil)
		require.Error(t, err)
	})

	t.Run("invalid range", func(t *testing.T) {
		_, err := server.VerifyChain(f.ctx, &types.QueryVerifyChainRequest{
			From: 2,
			To:   1,
		})
		require.Error(t, err)
	})

	t.Run("valid chain", func(t *testing.T) {
		appendRecord("TEST_1")
		appendRecord("TEST_2")
		appendRecord("TEST_3")

		res, err := server.VerifyChain(f.ctx, &types.QueryVerifyChainRequest{
			From: 0,
			To:   2,
		})

		require.NoError(t, err)
		require.True(t, res.Valid)
		require.NotEmpty(t, res.LastHash)
	})

	t.Run("valid partial chain", func(t *testing.T) {
		res, err := server.VerifyChain(f.ctx, &types.QueryVerifyChainRequest{
			From: 1,
			To:   2,
		})

		require.NoError(t, err)
		require.True(t, res.Valid)
	})

	t.Run("missing previous record", func(t *testing.T) {
		emptyFixture := initFixture(t)
		emptyServer := keeper.NewQueryServerImpl(emptyFixture.keeper)

		res, err := emptyServer.VerifyChain(emptyFixture.ctx, &types.QueryVerifyChainRequest{
			From: 1,
			To:   1,
		})

		require.NoError(t, err)
		require.False(t, res.Valid)
		require.Equal(t, uint64(1), res.BrokenAt)
	})

	t.Run("missing record in range", func(t *testing.T) {
		res, err := server.VerifyChain(f.ctx, &types.QueryVerifyChainRequest{
			From: 0,
			To:   10,
		})

		require.NoError(t, err)
		require.False(t, res.Valid)
		require.Equal(t, uint64(3), res.BrokenAt)
	})

	t.Run("tampered record hash", func(t *testing.T) {
		record, err := f.keeper.AuditRecord.Get(f.ctx, uint64(1))
		require.NoError(t, err)

		record.RecordHash = "tampered-hash"
		require.NoError(t, f.keeper.AuditRecord.Set(f.ctx, uint64(1), record))

		res, err := server.VerifyChain(f.ctx, &types.QueryVerifyChainRequest{
			From: 0,
			To:   2,
		})

		require.NoError(t, err)
		require.False(t, res.Valid)
		require.Equal(t, uint64(1), res.BrokenAt)
	})
}
