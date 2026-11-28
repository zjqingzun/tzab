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

func (k msgServer) RegisterResource(ctx context.Context, msg *types.MsgRegisterResource) (*types.MsgRegisterResourceResponse, error) {
	if msg == nil {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"message cannot be nil",
		)
	}

	creator := strings.TrimSpace(msg.Creator)
	resourceID := strings.TrimSpace(msg.ResourceId)
	classification := strings.TrimSpace(msg.Classification)

	if resourceID == "" || classification == "" {
		return nil, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"resource ID and classification are required",
		)
	}

	if _, err := k.addressCodec.StringToBytes(creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid creator address")
	}

	// Check resource ID uniqueness.
	exists, err := k.Resource.Has(ctx, resourceID)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to check resource ID")
	}

	if exists {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"resource %s already exists",
			resourceID,
		)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	resource := types.Resource{
		ResourceId:     resourceID,
		Owner:          creator,
		Classification: classification,
	}

	// Save resource to blockchain.
	if err := k.Resource.Set(ctx, resourceID, resource); err != nil {
		return nil, errorsmod.Wrap(err, "failed to save resource")
	}

	// Create audit payload hash.
	payload := fmt.Sprintf(
		"%s|%s|%s",
		resourceID,
		creator,
		classification,
	)

	payloadSum := sha256.Sum256([]byte(payload))
	payloadHash := hex.EncodeToString(payloadSum[:])

	_, err = k.AppendAudit(sdkCtx, types.AuditRecord{
		EventType:   "RESOURCE_REGISTER",
		Subject:     creator,
		Resource:    resourceID,
		Action:      "REGISTER",
		Decision:    "ALLOW",
		Actor:       creator,
		PayloadHash: payloadHash,
	})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to append audit record")
	}

	// Emit resource registered event.
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"resource_registered",
			sdk.NewAttribute("resource_id", resourceID),
			sdk.NewAttribute("owner", creator),
			sdk.NewAttribute("classification", classification),
		),
	)

	return &types.MsgRegisterResourceResponse{Ok: true}, nil
}
