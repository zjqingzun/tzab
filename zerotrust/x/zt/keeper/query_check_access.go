package keeper

import (
	"context"
	"errors"
	"strings"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"zerotrust/x/zt/types"
)

func (q queryServer) CheckAccess(ctx context.Context, req *types.QueryCheckAccessRequest) (*types.QueryCheckAccessResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	subject := strings.TrimSpace(req.Subject)
	resource := strings.TrimSpace(req.Resource)
	action := strings.TrimSpace(req.Action)

	if subject == "" || resource == "" || action == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"subject, resource and action are required",
		)
	}

	deny := func(reason string) *types.QueryCheckAccessResponse {
		return &types.QueryCheckAccessResponse{
			Allowed: false,
			Reason:  reason,
		}
	}

	// Check subject identity.
	identity, err := q.k.Identity.Get(ctx, subject)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return deny("subject identity does not exist"), nil
		}

		return nil, status.Error(
			codes.Internal,
			"failed to get subject identity",
		)
	}

	if identity.Status != "ACTIVE" {
		return deny("subject identity is not active"), nil
	}

	// Check resource.
	if _, err := q.k.Resource.Get(ctx, resource); err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return deny("resource does not exist"), nil
		}

		return nil, status.Error(
			codes.Internal,
			"failed to get resource",
		)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	now := sdkCtx.BlockTime().Unix()

	// Default denial response.
	response := deny("no valid grant found")

	err = q.k.Grant.Walk(ctx, nil, func(_ string, grant types.Grant) (bool, error) {
		matches :=
			strings.TrimSpace(grant.Subject) == subject &&
				strings.TrimSpace(grant.Resource) == resource &&
				strings.TrimSpace(grant.Action) == action

		if !matches {
			return false, nil
		}

		if grant.Status != "ACTIVE" {
			return false, nil
		}

		if grant.ExpiresAt != 0 && grant.ExpiresAt <= now {
			return false, nil
		}

		response = &types.QueryCheckAccessResponse{
			Allowed: true,
			Reason:  "access granted",
			GrantId: grant.GrantId,
		}

		return true, nil
	})
	if err != nil {
		return nil, status.Error(
			codes.Internal,
			"failed to check access grants",
		)
	}

	return response, nil
}
