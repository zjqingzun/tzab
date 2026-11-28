package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"zerotrust/x/zt/keeper"
	"zerotrust/x/zt/types"
)

func TestMsgRecordAccess(t *testing.T) {
	f := initFixture(t)
	server := keeper.NewMsgServerImpl(f.keeper)

	creator := newKeeperTestAddress(t, f, 4)

	t.Run("successful ALLOW record", func(t *testing.T) {
		msg := &types.MsgRecordAccess{
			Creator:  creator,
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "READ",
			Decision: "ALLOW",
		}

		res, err := server.RecordAccess(f.ctx, msg)

		require.NoError(t, err)
		require.Equal(t, uint64(0), res.AuditId)

		record, err := f.keeper.AuditRecord.Get(f.ctx, res.AuditId)
		require.NoError(t, err)

		require.Equal(t, "ACCESS_ATTEMPT", record.EventType)
		require.Equal(t, "did:alice", record.Subject)
		require.Equal(t, "resource-1", record.Resource)
		require.Equal(t, "READ", record.Action)
		require.Equal(t, "ALLOW", record.Decision)
		require.Equal(t, creator, record.Actor)
		require.NotEmpty(t, record.PayloadHash)
	})

	t.Run("successful DENY record", func(t *testing.T) {
		res, err := server.RecordAccess(f.ctx, &types.MsgRecordAccess{
			Creator:  creator,
			Subject:  "did:bob",
			Resource: "resource-2",
			Action:   "WRITE",
			Decision: "DENY",
			Reason:   "insufficient permission",
		})

		require.NoError(t, err)

		record, err := f.keeper.AuditRecord.Get(f.ctx, res.AuditId)
		require.NoError(t, err)
		require.Equal(t, "DENY", record.Decision)
		require.Equal(t, "insufficient permission", record.Reason)
	})

	t.Run("nil message", func(t *testing.T) {
		_, err := server.RecordAccess(f.ctx, nil)
		require.Error(t, err)
	})

	t.Run("missing required fields", func(t *testing.T) {
		_, err := server.RecordAccess(f.ctx, &types.MsgRecordAccess{
			Creator: creator,
			Subject: "did:alice",
		})
		require.Error(t, err)
	})

	t.Run("invalid decision", func(t *testing.T) {
		_, err := server.RecordAccess(f.ctx, &types.MsgRecordAccess{
			Creator:  creator,
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "READ",
			Decision: "UNKNOWN",
		})
		require.Error(t, err)
	})

	t.Run("DENY requires reason", func(t *testing.T) {
		_, err := server.RecordAccess(f.ctx, &types.MsgRecordAccess{
			Creator:  creator,
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "READ",
			Decision: "DENY",
		})
		require.Error(t, err)
	})

	t.Run("invalid creator", func(t *testing.T) {
		_, err := server.RecordAccess(f.ctx, &types.MsgRecordAccess{
			Creator:  "invalid-address",
			Subject:  "did:alice",
			Resource: "resource-1",
			Action:   "READ",
			Decision: "ALLOW",
		})
		require.Error(t, err)
	})

	t.Run("invalid payload hash", func(t *testing.T) {
		_, err := server.RecordAccess(f.ctx, &types.MsgRecordAccess{
			Creator:     creator,
			Subject:     "did:alice",
			Resource:    "resource-1",
			Action:      "READ",
			Decision:    "ALLOW",
			PayloadHash: "invalid-hash",
		})
		require.Error(t, err)
	})

	t.Run("valid supplied payload hash", func(t *testing.T) {
		sum := sha256.Sum256([]byte("test-payload"))
		hash := hex.EncodeToString(sum[:])

		res, err := server.RecordAccess(f.ctx, &types.MsgRecordAccess{
			Creator:     creator,
			Subject:     "did:alice",
			Resource:    "resource-1",
			Action:      "READ",
			Decision:    "ALLOW",
			PayloadHash: hash,
		})

		require.NoError(t, err)

		record, err := f.keeper.AuditRecord.Get(f.ctx, res.AuditId)
		require.NoError(t, err)
		require.Equal(t, hash, record.PayloadHash)
	})
}
