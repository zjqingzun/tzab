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

func (k msgServer) RevokeAccess(ctx context.Context, msg *types.MsgRevokeAccess) (*types.MsgRevokeAccessResponse, error) {
	if msg == nil {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"message cannot be nil",
		)
	}

	creator := strings.TrimSpace(msg.Creator)
	grantID := strings.TrimSpace(msg.GrantId)
	reason := strings.TrimSpace(msg.Reason)

	if creator == "" || grantID == "" {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"creator and grant ID are required",
		)
	}

	if _, err := k.addressCodec.StringToBytes(creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid creator address")
	}

	if err := k.mustBeAdmin(ctx, creator); err != nil {
		return nil, err
	}

	// Get grant and distinguish not-found from storage errors.
	grant, err := k.Grant.Get(ctx, grantID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"grant %s does not exist",
				grantID,
			)
		}

		return nil, errorsmod.Wrap(err, "failed to get grant")
	}

	if grant.Status != "ACTIVE" {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"grant %s is not active",
			grantID,
		)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	grant.Status = "REVOKED"
	grant.RevokedAt = sdkCtx.BlockTime().Unix()

	if err := k.Grant.Set(ctx, grantID, grant); err != nil {
		return nil, errorsmod.Wrap(err, "failed to revoke grant")
	}

	// Include grant ID and reason in the audit payload hash.
	payload := fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s|%s",
		grantID,
		grant.Subject,
		grant.Resource,
		grant.Action,
		creator,
		reason,
		grant.Status,
	)

	payloadSum := sha256.Sum256([]byte(payload))
	payloadHash := hex.EncodeToString(payloadSum[:])

	_, err = k.AppendAudit(sdkCtx, types.AuditRecord{
		EventType:   "ACCESS_REVOKE",
		Subject:     grant.Subject,
		Resource:    grant.Resource,
		Action:      grant.Action,
		Decision:    "REVOKED",
		Actor:       creator,
		PayloadHash: payloadHash,
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to append audit record")
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"access_revoked",
			sdk.NewAttribute("grant_id", grantID),
			sdk.NewAttribute("subject", grant.Subject),
			sdk.NewAttribute("resource", grant.Resource),
			sdk.NewAttribute("action", grant.Action),
			sdk.NewAttribute("revoked_by", creator),
			sdk.NewAttribute("reason", reason),
		),
	)

	return &types.MsgRevokeAccessResponse{Ok: true}, nil
}
