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

func (k msgServer) RecordAccess(ctx context.Context, msg *types.MsgRecordAccess) (*types.MsgRecordAccessResponse, error) {
	if msg == nil {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"message cannot be nil",
		)
	}

	creator := strings.TrimSpace(msg.Creator)
	subject := strings.TrimSpace(msg.Subject)
	resource := strings.TrimSpace(msg.Resource)
	action := strings.TrimSpace(msg.Action)
	decision := strings.ToUpper(strings.TrimSpace(msg.Decision))
	reason := strings.TrimSpace(msg.Reason)
	payloadHash := strings.TrimSpace(msg.PayloadHash)

	// Validate required fields.
	if subject == "" || resource == "" || action == "" {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"subject, resource and action are required",
		)
	}

	// Validate decision.
	if decision != "ALLOW" && decision != "DENY" {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"decision must be ALLOW or DENY",
		)
	}

	if decision == "DENY" && reason == "" {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"reason is required for DENY decision",
		)
	}

	// Validate transaction creator.
	if _, err := k.addressCodec.StringToBytes(creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid creator address")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Generate payload hash if not provided.
	if payloadHash == "" {
		payload := fmt.Sprintf(
			"%s|%s|%s|%s|%s|%s",
			subject,
			resource,
			action,
			decision,
			reason,
			creator,
		)

		sum := sha256.Sum256([]byte(payload))
		payloadHash = hex.EncodeToString(sum[:])
	} else {
		// Validate supplied SHA-256 hash.
		decoded, err := hex.DecodeString(payloadHash)
		if err != nil || len(decoded) != sha256.Size {
			return nil, errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"payload hash must be a valid SHA-256 hash",
			)
		}

		payloadHash = strings.ToLower(payloadHash)
	}

	// Append access audit record.
	auditID, err := k.AppendAudit(sdkCtx, types.AuditRecord{
		EventType:   "ACCESS_ATTEMPT",
		Subject:     subject,
		Resource:    resource,
		Action:      action,
		Decision:    decision,
		Reason:      reason,
		Actor:       creator,
		PayloadHash: payloadHash,
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to append access audit record")
	}

	// Emit access audit event.
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"access_recorded",
			sdk.NewAttribute("audit_id", fmt.Sprintf("%d", auditID)),
			sdk.NewAttribute("subject", subject),
			sdk.NewAttribute("resource", resource),
			sdk.NewAttribute("action", action),
			sdk.NewAttribute("decision", decision),
			sdk.NewAttribute("actor", creator),
		),
	)

	return &types.MsgRecordAccessResponse{AuditId: auditID, }, nil
}
