package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"zerotrust/x/zt/types"
)

func computeAuditHash(r types.AuditRecord) string {
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

// AppendAudit tạo và lưu một audit record vào hash-chain.
// Id, BlockHeight, Timestamp, PrevHash và RecordHash được chain thiết lập.
func (k Keeper) AppendAudit(
	ctx sdk.Context,
	record types.AuditRecord,
) (uint64, error) {
	// Next trả về ID hiện tại và tăng sequence.
	id, err := k.AuditRecordSeq.Next(ctx)
	if err != nil {
		return 0, err
	}

	prevHash := "GENESIS"

	if id > 0 {
		previous, err := k.AuditRecord.Get(ctx, id-1)
		if err != nil {
			return 0, err
		}

		prevHash = previous.RecordHash
	}

	record.Id = id
	record.BlockHeight = ctx.BlockHeight()
	record.Timestamp = ctx.BlockTime().Unix()
	record.PrevHash = prevHash
	record.RecordHash = computeAuditHash(record)

	if err := k.AuditRecord.Set(ctx, id, record); err != nil {
		return 0, err
	}

	return id, nil
}
