package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"zerotrust/x/zt/types"
)

func expectedAuditHash(r types.AuditRecord) string {
	data := fmt.Sprintf(
		"%d|%s|%s|%s|%s|%s|%s|%s|%d|%d|%s",
		r.Id,
		r.EventType,
		r.Subject,
		r.Resource,
		r.Action,
		r.Decision,
		r.Actor,
		r.PayloadHash,
		r.BlockHeight,
		r.Timestamp,
		r.PrevHash,
	)

	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func TestAppendAudit(t *testing.T) {
	f := initFixture(t)
	ctx := sdk.UnwrapSDKContext(f.ctx)

	t.Run("FirstRecord", func(t *testing.T) {
		id, err := f.keeper.AppendAudit(ctx, types.AuditRecord{
			EventType:   "IDENTITY_REGISTER",
			Subject:     "did:example:1",
			Action:      "REGISTER",
			Decision:    "ALLOW",
			Actor:       "admin",
			PayloadHash: "payload-1",
		})

		require.NoError(t, err)
		require.Equal(t, uint64(0), id)

		record, err := f.keeper.AuditRecord.Get(ctx, id)
		require.NoError(t, err)

		require.Equal(t, uint64(0), record.Id)
		require.Equal(t, "GENESIS", record.PrevHash)
		require.Equal(t, ctx.BlockHeight(), record.BlockHeight)
		require.Equal(t, ctx.BlockTime().Unix(), record.Timestamp)
		require.Equal(t, expectedAuditHash(record), record.RecordHash)
		require.Len(t, record.RecordHash, 64)
	})

	t.Run("SecondRecord", func(t *testing.T) {
		first, err := f.keeper.AuditRecord.Get(ctx, 0)
		require.NoError(t, err)

		id, err := f.keeper.AppendAudit(ctx, types.AuditRecord{
			EventType:   "ACCESS_ATTEMPT",
			Subject:     "did:example:2",
			Resource:    "resource-1",
			Action:      "READ",
			Decision:    "ALLOW",
			Actor:       "gateway",
			PayloadHash: "payload-2",
		})

		require.NoError(t, err)
		require.Equal(t, uint64(1), id)

		second, err := f.keeper.AuditRecord.Get(ctx, id)
		require.NoError(t, err)

		require.Equal(t, first.RecordHash, second.PrevHash)
		require.Equal(t, expectedAuditHash(second), second.RecordHash)
	})

	t.Run("SequentialIDs", func(t *testing.T) {
		id, err := f.keeper.AppendAudit(ctx, types.AuditRecord{
			EventType: "TEST",
		})

		require.NoError(t, err)
		require.Equal(t, uint64(2), id)
	})
}

func TestComputeExpectedAuditHash(t *testing.T) {
	record := types.AuditRecord{
		Id:          1,
		EventType:   "ACCESS_ATTEMPT",
		Subject:     "subject",
		Resource:    "resource",
		Action:      "READ",
		Decision:    "ALLOW",
		Actor:       "gateway",
		PayloadHash: "abc123",
		BlockHeight: 100,
		Timestamp:   1700000000,
		PrevHash:    "GENESIS",
	}

	hash := expectedAuditHash(record)

	require.Len(t, hash, 64)

	decoded, err := hex.DecodeString(hash)
	require.NoError(t, err)
	require.Len(t, decoded, sha256.Size)
}
