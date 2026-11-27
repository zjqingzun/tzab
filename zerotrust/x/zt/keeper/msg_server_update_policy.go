package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zerotrust/x/zt/types"
)

func (k msgServer) UpdatePolicy(ctx context.Context, msg *types.MsgUpdatePolicy) (*types.MsgUpdatePolicyResponse, error) {
	if msg == nil {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"message cannot be nil",
		)
	}

	creator := strings.TrimSpace(msg.Creator)
	policyID := strings.TrimSpace(msg.PolicyId)
	version := strings.TrimSpace(msg.Version)
	contentHash := strings.TrimSpace(msg.ContentHash)

	if policyID == "" || version == "" || contentHash == "" {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"policy ID, version and content hash are required",
		)
	}

	if _, err := k.addressCodec.StringToBytes(creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid creator address")
	}

	if err := k.mustBeAdmin(ctx, creator); err != nil {
		return nil, err
	}

	// Load existing policy.
	policy, err := k.Policy.Get(ctx, policyID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"policy %s does not exist",
				policyID,
			)
		}

		return nil, errorsmod.Wrap(err, "failed to get policy")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime().Unix()

	// Update policy fields.
	policy.Version = version
	policy.ContentHash = contentHash
	policy.UpdatedBy = creator
	policy.UpdatedAt = now

	if err := k.Policy.Set(ctx, policyID, policy); err != nil {
		return nil, errorsmod.Wrap(err, "failed to update policy")
	}

	// Create audit payload hash.
	payload := fmt.Sprintf(
		"%s|%s|%s|%s",
		policyID,
		version,
		contentHash,
		creator,
	)

	payloadSum := sha256.Sum256([]byte(payload))
	payloadHash := hex.EncodeToString(payloadSum[:])

	_, err = k.AppendAudit(sdkCtx, types.AuditRecord{
		EventType:   "POLICY_UPDATE",
		Subject:     creator,
		Resource:    policyID,
		Action:      "UPDATE",
		Decision:    "ALLOW",
		Actor:       creator,
		PayloadHash: payloadHash,
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to append audit record")
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"policy_updated",
			sdk.NewAttribute("policy_id", policyID),
			sdk.NewAttribute("version", version),
			sdk.NewAttribute("creator", creator),
		),
	)

	return &types.MsgUpdatePolicyResponse{Ok: true}, nil
}
