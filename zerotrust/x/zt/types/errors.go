package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/zt module sentinel errors
var (
	ErrInvalidSigner = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")

	ErrUnauthorized = errors.Register(
		ModuleName,
		1101,
		"unauthorized signer",
	)

	ErrInvalidIdentity = errors.Register(
		ModuleName,
		1102,
		"invalid identity",
	)

	ErrIdentityExists = errors.Register(
		ModuleName,
		1103,
		"identity already exists",
	)
)
