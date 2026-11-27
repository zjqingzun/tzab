package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"zerotrust/x/zt/types"
)

func (k msgServer) RegisterIdentity(
	ctx context.Context,
	msg *types.MsgRegisterIdentity,
) (*types.MsgRegisterIdentityResponse, error) {
	if msg == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidIdentity, "message cannot be nil")
	}

	// Validate required fields.
	if strings.TrimSpace(msg.Did) == "" ||
		strings.TrimSpace(msg.Pubkey) == "" ||
		strings.TrimSpace(msg.Role) == "" {
		return nil, errorsmod.Wrap(types.ErrInvalidIdentity, "DID, public key and role are required")
	}

	// Validate creator address.
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid creator address")
	}

	// Only admin can register identity.
	if err := k.mustBeAdmin(ctx, msg.Creator); err != nil {
		return nil, err
	}

	// Check whether identity already exists.
	exists, err := k.Identity.Has(ctx, msg.Did)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to check identity")
	}
	if exists {
		return nil, errorsmod.Wrap(types.ErrIdentityExists, msg.Did)
	}

	// Convert context.Context to sdk.Context for block metadata and events.
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Create identity.
	identity := types.Identity{
		Did:       msg.Did,
		Pubkey:    msg.Pubkey,
		Role:      msg.Role,
		Status:    "ACTIVE",
		CreatedAt: sdkCtx.BlockTime().Unix(),
	}

	if err := k.Identity.Set(ctx, msg.Did, identity); err != nil {
		return nil, errorsmod.Wrap(err, "failed to save identity")
	}

	// Hash identity payload.
	payload := fmt.Sprintf(
		"%s|%s|%s",
		identity.Did,
		identity.Pubkey,
		identity.Role,
	)
	payloadHash := sha256.Sum256([]byte(payload))

	// Append audit record.
	_, err = k.AppendAudit(sdkCtx, types.AuditRecord{
		EventType:   "IDENTITY_REGISTER",
		Subject:     identity.Did,
		Decision:    "N/A",
		Actor:       msg.Creator,
		PayloadHash: hex.EncodeToString(payloadHash[:]),
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to append audit record")
	}

	// Emit event.
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"identity_registered",
			sdk.NewAttribute("did", identity.Did),
			sdk.NewAttribute("role", identity.Role),
			sdk.NewAttribute("actor", msg.Creator),
		),
	)

	return &types.MsgRegisterIdentityResponse{
		Ok: true,
	}, nil
}
