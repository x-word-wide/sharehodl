package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// EquityKeeper defines the expected equity keeper interface
type EquityKeeper interface {
	GetShareholding(ctx context.Context, companyID uint64, classID, holder string) (interface{}, bool)
	TransferShares(ctx sdk.Context, companyID uint64, classID string, from, to string, shares math.Int) error
	IsSymbolTaken(ctx context.Context, symbol string) bool
	GetCompany(ctx context.Context, companyID uint64) (interface{}, bool)

	// Fraud investigation and delisting (Phase 3)
	InitiateFraudInvestigation(ctx sdk.Context, companyID uint64, reportID uint64, initiator string) error
	ConfirmFraudAndDelist(ctx sdk.Context, companyID uint64, reportID uint64, isFraud bool, authority string) error
	HaltTrading(ctx sdk.Context, companyID uint64, reason string, initiator string, reportID uint64) error
	ResumeTrading(ctx sdk.Context, companyID uint64, authority string) error
	IsTradingHalted(ctx sdk.Context, companyID uint64) bool

	// Beneficial owner tracking for dividend distribution
	// When escrow locks equity shares, register the beneficial owner so they still receive dividends
	RegisterBeneficialOwner(ctx sdk.Context, moduleAccount string, companyID uint64, classID string, beneficialOwner string, shares math.Int, referenceID uint64, referenceType string) error
	UnregisterBeneficialOwner(ctx sdk.Context, moduleAccount string, companyID uint64, classID string, beneficialOwner string, referenceID uint64) error
	UpdateBeneficialOwnerShares(ctx sdk.Context, moduleAccount string, companyID uint64, classID string, beneficialOwner string, referenceID uint64, newShares math.Int) error
}

// UniversalStakingKeeper defines the expected universal staking keeper interface
// Used to check tier requirements for moderator registration and dispute resolution
type UniversalStakingKeeper interface {
	// CanModerate checks if user can be a moderator (requires Warden+ tier)
	CanModerate(ctx sdk.Context, addr sdk.AccAddress) bool
	// CanModerateLargeDisputes checks if user can handle large disputes (requires Steward+ tier)
	CanModerateLargeDisputes(ctx sdk.Context, addr sdk.AccAddress) bool
	// GetUserTierInt returns the user's current stake tier as int
	GetUserTierInt(ctx sdk.Context, addr sdk.AccAddress) int
	// GetReputation returns the user's reputation score
	GetReputation(ctx sdk.Context, addr sdk.AccAddress) math.LegacyDec
	// RewardSuccessfulDispute rewards a moderator for resolving a dispute fairly
	RewardSuccessfulDispute(ctx sdk.Context, addr sdk.AccAddress, disputeID string) error
	// PenalizeBadDispute penalizes a moderator for unfair resolution
	PenalizeBadDispute(ctx sdk.Context, addr sdk.AccAddress, disputeID string) error
	// GetStakeAge returns how long the user has been at their current tier (anti-sybil)
	GetStakeAge(ctx sdk.Context, addr sdk.AccAddress) int64 // Returns seconds staked at current tier

	// ==========================================================================
	// STAKE-AS-TRUST-CEILING: Commitment Management
	// Only the SENDER (market maker) needs stake commitment, not the recipient
	// ==========================================================================

	// CanCommit checks if user can commit the specified amount (has enough available stake)
	CanCommit(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) bool
	// GetAvailableStake returns stake minus current commitments
	GetAvailableStake(ctx sdk.Context, owner sdk.AccAddress) math.Int
	// AddEscrowCommitment adds a commitment when escrow is funded by sender
	AddEscrowCommitment(ctx sdk.Context, owner sdk.AccAddress, escrowID uint64, amount math.Int) error
	// ReleaseEscrowCommitment releases commitment when escrow is completed/cancelled
	ReleaseEscrowCommitment(ctx sdk.Context, owner sdk.AccAddress, escrowID uint64) error
}

// Universal staking tier constants (matching x/staking/types)
const (
	StakeTierHolder    = 0 // Cannot moderate
	StakeTierKeeper    = 1 // Can lend/borrow only
	StakeTierWarden    = 2 // Can be moderator (disputes < 100K HODL)
	StakeTierSteward   = 3 // Can moderate large disputes (< 1M HODL)

	// Anti-sybil: Minimum stake age to submit reports (7 days in seconds)
	// Prevents banned users from creating new accounts and immediately reporting
	MinStakeAgeForReporting = 7 * 24 * 60 * 60 // 7 days

	// Anti-sybil: Minimum stake age to vote on investigations (3 days in seconds)
	// Prevents reviewers from unbonding to escape penalties
	MinStakeAgeForVoting = 3 * 24 * 60 * 60 // 3 days

	StakeTierArchon    = 4 // Can moderate any dispute + slash moderators
	StakeTierValidator = 5 // Full oversight
)
