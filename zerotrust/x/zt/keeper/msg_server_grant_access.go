package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zerotrust/x/zt/types"
)

func (k msgServer) GrantAccess(ctx context.Context, msg *types.MsgGrantAccess) (*types.MsgGrantAccessResponse, error) {
	if msg == nil {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"message cannot be nil",
		)
	}

	grantID := strings.TrimSpace(msg.GrantId)
	subject := strings.TrimSpace(msg.Subject)
	resource := strings.TrimSpace(msg.Resource)
	action := strings.TrimSpace(msg.Action)
	creator := strings.TrimSpace(msg.Creator)

	if grantID == "" || subject == "" || resource == "" || action == "" {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"grant ID, subject, resource and action are required",
		)
	}

	if _, err := k.addressCodec.StringToBytes(creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid creator address")
	}

	if err := k.mustBeAdmin(ctx, creator); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime().Unix()

	if msg.ExpiresAt < 0 {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"grant expiration cannot be negative",
		)
	}

	if msg.ExpiresAt != 0 && msg.ExpiresAt <= now {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"grant expiration must be in the future",
		)
	}

	// Check subject identity.
	identity, err := k.Identity.Get(ctx, subject)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"subject identity does not exist",
			)
		}

		return nil, errorsmod.Wrap(err, "failed to get subject identity")
	}

	if identity.Status != "ACTIVE" {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"subject identity %s is not active",
			subject,
		)
	}

	// Check resource.
	if _, err := k.Resource.Get(ctx, resource); err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"resource does not exist",
			)
		}

		return nil, errorsmod.Wrap(err, "failed to get resource")
	}

	// Check grant ID uniqueness.
	exists, err := k.Grant.Has(ctx, grantID)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to check grant ID")
	}

	if exists {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"grant %s already exists",
			grantID,
		)
	}

	// Prevent duplicate active and unexpired permissions.
	duplicate := false

	err = k.Grant.Walk(ctx, nil, func(_ string, existing types.Grant) (bool, error) {
		samePermission :=
			strings.TrimSpace(existing.Subject) == subject &&
				strings.TrimSpace(existing.Resource) == resource &&
				strings.TrimSpace(existing.Action) == action

		stillValid :=
			existing.Status == "ACTIVE" &&
				(existing.ExpiresAt == 0 || existing.ExpiresAt > now)

		if samePermission && stillValid {
			duplicate = true
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to check existing grants")
	}

	if duplicate {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"an active grant already exists for this subject, resource and action",
		)
	}

	grant := types.Grant{
		GrantId:   grantID,
		Subject:   subject,
		Resource:  resource,
		Action:    action,
		Grantor:   creator,
		Status:    "ACTIVE",
		ExpiresAt: msg.ExpiresAt,
		CreatedAt: now,
		RevokedAt: 0,
	}

	if err := k.Grant.Set(ctx, grantID, grant); err != nil {
		return nil, errorsmod.Wrap(err, "failed to save grant")
	}

	// Hash the grant details for the audit record.
	payload := fmt.Sprintf(
		"%s|%s|%s|%s|%d|%s",
		grantID,
		subject,
		resource,
		action,
		msg.ExpiresAt,
		creator,
	)

	payloadSum := sha256.Sum256([]byte(payload))
	payloadHash := hex.EncodeToString(payloadSum[:])

	_, err = k.AppendAudit(sdkCtx, types.AuditRecord{
		EventType:   "ACCESS_GRANT",
		Subject:     subject,
		Resource:    resource,
		Action:      action,
		Decision:    "ALLOW",
		Actor:       creator,
		PayloadHash: payloadHash,
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to append audit record")
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"access_granted",
			sdk.NewAttribute("grant_id", grantID),
			sdk.NewAttribute("subject", subject),
			sdk.NewAttribute("resource", resource),
			sdk.NewAttribute("action", action),
			sdk.NewAttribute("grantor", creator),
			sdk.NewAttribute(
				"expires_at",
				strconv.FormatInt(grant.ExpiresAt, 10),
			),
		),
	)

	return &types.MsgGrantAccessResponse{}, nil
}
