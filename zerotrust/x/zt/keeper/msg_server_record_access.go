package keeper

import (
	"context"

	"zerotrust/x/zt/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) RecordAccess(ctx context.Context, msg *types.MsgRecordAccess) (*types.MsgRecordAccessResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgRecordAccessResponse{}, nil
}
