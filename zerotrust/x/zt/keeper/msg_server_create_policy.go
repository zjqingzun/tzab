package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zerotrust/x/zt/types"
)

func (k msgServer) CreatePolicy(
	ctx context.Context,
	msg *types.MsgCreatePolicy,
) (*types.MsgCreatePolicyResponse, error) {
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

	// Check policy ID uniqueness.
	exists, err := k.Policy.Has(ctx, policyID)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to check policy ID")
	}

	if exists {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"policy %s already exists",
			policyID,
		)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime().Unix()

	policy := types.Policy{
		PolicyId:    policyID,
		Version:     version,
		ContentHash: contentHash,
		UpdatedBy:   creator,
		UpdatedAt:   now,
	}

	if err := k.Policy.Set(ctx, policyID, policy); err != nil {
		return nil, errorsmod.Wrap(err, "failed to save policy")
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
		EventType:   "POLICY_CREATE",
		Subject:     creator,
		Resource:    policyID,
		Action:      "CREATE",
		Decision:    "ALLOW",
		Actor:       creator,
		PayloadHash: payloadHash,
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to append audit record")
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"policy_created",
			sdk.NewAttribute("policy_id", policyID),
			sdk.NewAttribute("version", version),
			sdk.NewAttribute("creator", creator),
		),
	)

	return &types.MsgCreatePolicyResponse{Ok: true}, nil
}
