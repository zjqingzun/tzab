package keeper

import (
	"context"

	"zerotrust/x/zt/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) GrantAccess(ctx context.Context, msg *types.MsgGrantAccess) (*types.MsgGrantAccessResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgGrantAccessResponse{}, nil
}
