package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	"zerotrust/x/zt/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q queryServer) VerifyChain(ctx context.Context, req *types.QueryVerifyChainRequest) (*types.QueryVerifyChainResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.From > req.To {
		return nil, status.Error(
			codes.InvalidArgument,
			"from must be less than or equal to to",
		)
	}

	expectedPrevHash := "GENESIS"

	// Get previous record when verification does not start at zero.
	if req.From > 0 {
		previous, err := q.k.AuditRecord.Get(ctx, req.From-1)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				return &types.QueryVerifyChainResponse{
					Valid:    false,
					BrokenAt: req.From,
				}, nil
			}

			return nil, status.Error(
				codes.Internal,
				"failed to get previous audit record",
			)
		}

		expectedPrevHash = previous.RecordHash
	}

	lastHash := ""

	// Verify records in the requested range.
	for id := req.From; ; id++ {
		record, err := q.k.AuditRecord.Get(ctx, id)
		if err != nil {
			if errors.Is(err, collections.ErrNotFound) {
				return &types.QueryVerifyChainResponse{
					Valid:    false,
					BrokenAt: id,
					LastHash: lastHash,
				}, nil
			}

			return nil, status.Error(
				codes.Internal,
				"failed to get audit record",
			)
		}

		// Verify record ID.
		if record.Id != id {
			return &types.QueryVerifyChainResponse{
				Valid:    false,
				BrokenAt: id,
				LastHash: lastHash,
			}, nil
		}

		// Verify previous record hash.
		if record.PrevHash != expectedPrevHash {
			return &types.QueryVerifyChainResponse{
				Valid:    false,
				BrokenAt: id,
				LastHash: lastHash,
			}, nil
		}

		// Recalculate record hash.
		calculatedHash := computeAuditHash(record)

		if record.RecordHash != calculatedHash {
			return &types.QueryVerifyChainResponse{
				Valid:    false,
				BrokenAt: id,
				LastHash: lastHash,
			}, nil
		}

		// Current record is valid.
		lastHash = record.RecordHash
		expectedPrevHash = record.RecordHash

		if id == req.To {
			break
		}
	}

	return &types.QueryVerifyChainResponse{
		Valid:    true,
		LastHash: lastHash,
	}, nil
}
