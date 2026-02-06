package types

import (
	"cosmossdk.io/errors"
)

// x/bridge module sentinel errors
var (
	// Asset errors
	ErrAssetNotSupported     = errors.Register(ModuleName, 1, "asset not supported for bridging")
	ErrInvalidAsset          = errors.Register(ModuleName, 2, "invalid asset")
	ErrAssetAlreadySupported = errors.Register(ModuleName, 3, "asset already supported")

	// Swap errors
	ErrInvalidSwapAmount     = errors.Register(ModuleName, 10, "invalid swap amount")
	ErrInsufficientReserve   = errors.Register(ModuleName, 11, "insufficient reserve balance")
	ErrSwapFailed            = errors.Register(ModuleName, 12, "swap execution failed")
	ErrInvalidSwapDirection  = errors.Register(ModuleName, 13, "invalid swap direction")
	ErrZeroAmount            = errors.Register(ModuleName, 14, "amount must be greater than zero")

	// Rate errors
	ErrInvalidSwapRate       = errors.Register(ModuleName, 20, "invalid swap rate")
	ErrRateNotSet            = errors.Register(ModuleName, 21, "swap rate not set for asset")

	// Reserve errors
	ErrInvalidReserveRatio   = errors.Register(ModuleName, 30, "reserve ratio below minimum threshold")
	ErrReserveNotFound       = errors.Register(ModuleName, 31, "reserve balance not found")
	ErrInvalidDeposit        = errors.Register(ModuleName, 32, "invalid reserve deposit")

	// Withdrawal errors
	ErrProposalNotFound      = errors.Register(ModuleName, 40, "withdrawal proposal not found")
	ErrProposalExpired       = errors.Register(ModuleName, 41, "withdrawal proposal expired")
	ErrProposalNotActive     = errors.Register(ModuleName, 42, "withdrawal proposal not active")
	ErrProposalAlreadyVoted  = errors.Register(ModuleName, 43, "validator already voted on this proposal")
	ErrInvalidProposal       = errors.Register(ModuleName, 44, "invalid withdrawal proposal")
	ErrExceedsMaxWithdrawal  = errors.Register(ModuleName, 45, "withdrawal amount exceeds maximum allowed")
	ErrProposalNotApproved   = errors.Register(ModuleName, 46, "withdrawal proposal not approved")

	// Authorization errors
	ErrNotValidator          = errors.Register(ModuleName, 50, "only validators can perform this action")
	ErrUnauthorized          = errors.Register(ModuleName, 51, "unauthorized")
	ErrInvalidValidator      = errors.Register(ModuleName, 52, "invalid validator address")

	// Parameter errors
	ErrInvalidParams         = errors.Register(ModuleName, 60, "invalid module parameters")
	ErrInvalidFeeRate        = errors.Register(ModuleName, 61, "invalid fee rate")
	ErrInvalidMinReserve     = errors.Register(ModuleName, 62, "invalid minimum reserve ratio")
)
