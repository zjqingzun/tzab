package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"zerotrust/x/zt/types"
)

func (k Keeper) mustBeAdmin(ctx context.Context, signer string) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return errorsmod.Wrap(err, "failed to get module params")
	}

	if signer != params.Admin {
		return errorsmod.Wrapf(
			sdkerrors.ErrUnauthorized,
			"signer %s is not admin",
			signer,
		)
	}

	return nil
}

func (k Keeper) mustBeGateway(ctx context.Context, signer string) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return errorsmod.Wrap(err, "failed to get module params")
	}

	if signer != params.Gateway {
		return errorsmod.Wrapf(
			sdkerrors.ErrUnauthorized,
			"signer %s is not a registered gateway",
			signer,
		)
	}

	return nil
}

var _ = types.ModuleName
