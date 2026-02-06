package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	// Methods to retrieve account information
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// UniversalStakingKeeper defines the expected universal staking keeper interface
// Used for tier checks and stake locking for company listings
type UniversalStakingKeeper interface {
	// Tier checks
	CanSubmitListing(ctx sdk.Context, addr sdk.AccAddress) bool
	CanVerifyBusinessByShares(ctx sdk.Context, addr sdk.AccAddress, totalShares int64) bool
	GetUserTierInt(ctx sdk.Context, addr sdk.AccAddress) int

	// Reputation
	GetReputation(ctx sdk.Context, addr sdk.AccAddress) math.LegacyDec
	RewardSuccessfulVerification(ctx sdk.Context, addr sdk.AccAddress, businessID string) error
	PenalizeFailedVerification(ctx sdk.Context, addr sdk.AccAddress, businessID string) error

	// Stake locks for company listings
	LockForPendingListing(ctx sdk.Context, addr sdk.AccAddress, listingID string) error
	UnlockPendingListing(ctx sdk.Context, addr sdk.AccAddress, listingID string) error
	LockForCompanyListing(ctx sdk.Context, addr sdk.AccAddress, companyID string) error
	UnlockCompanyListing(ctx sdk.Context, addr sdk.AccAddress, companyID string) error
}

// Stake tier constants for equity verification
const (
	StakeTierHolder    = 0
	StakeTierKeeper    = 1 // Can submit listings, verify small businesses
	StakeTierWarden    = 2 // Can verify medium businesses
	StakeTierSteward   = 3 // Can verify large businesses
	StakeTierArchon    = 4 // Can verify enterprise businesses
	StakeTierValidator = 5
)

// BeneficialOwnerKeeper defines the interface for other modules to register beneficial ownership
// This allows escrow, lending, and DEX modules to track who really owns shares they hold
type BeneficialOwnerKeeper interface {
	// RegisterBeneficialOwner records that shares in a module belong to a specific owner
	RegisterBeneficialOwner(
		ctx sdk.Context,
		moduleAccount string,
		companyID uint64,
		classID string,
		beneficialOwner string,
		shares math.Int,
		refID uint64,
		refType string,
	) error

	// UnregisterBeneficialOwner removes a beneficial ownership record when shares are released
	UnregisterBeneficialOwner(
		ctx sdk.Context,
		moduleAccount string,
		companyID uint64,
		classID string,
		beneficialOwner string,
		refID uint64,
	) error

	// UpdateBeneficialOwnerShares updates the share count for an existing beneficial ownership record
	UpdateBeneficialOwnerShares(
		ctx sdk.Context,
		moduleAccount string,
		companyID uint64,
		classID string,
		beneficialOwner string,
		refID uint64,
		newShares math.Int,
	) error
}

// ValidatorKeeper defines the interface for checking validator status
// This is used for audit verification to ensure only validators can verify audits
type ValidatorKeeper interface {
	// IsValidator checks if an address is currently a validator
	IsValidator(ctx sdk.Context, addr sdk.AccAddress) bool
}