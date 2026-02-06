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
	SendCoinsFromModuleToModule(ctx context.Context, senderModule string, recipientModule string, amt sdk.Coins) error
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

	// Beneficial Owner Registry - for dividend distribution on locked collateral
	RegisterBeneficialOwner(ctx sdk.Context, moduleAccount string, companyID uint64, classID string, beneficialOwner string, shares math.Int, referenceID uint64, referenceType string) error
	UnregisterBeneficialOwner(ctx sdk.Context, moduleAccount string, companyID uint64, classID string, beneficialOwner string, referenceID uint64) error
	UpdateBeneficialOwnerShares(ctx sdk.Context, moduleAccount string, companyID uint64, classID string, beneficialOwner string, referenceID uint64, newShares math.Int) error
}

// DEXKeeper defines the expected DEX keeper interface
type DEXKeeper interface {
	GetAssetPrice(ctx context.Context, symbol string) interface{}
	GetMarket(ctx context.Context, baseSymbol, quoteSymbol string) (interface{}, bool)
}

// UniversalStakingKeeper defines the expected universal staking keeper interface
// Used to check tier requirements for lending participation
type UniversalStakingKeeper interface {
	// CanLend checks if user can participate in lending (requires Keeper+ tier)
	CanLend(ctx sdk.Context, addr sdk.AccAddress) bool
	// CanBorrow checks if user can borrow (requires Keeper+ tier)
	CanBorrow(ctx sdk.Context, addr sdk.AccAddress) bool
	// GetUserTierInt returns the user's current stake tier as int
	GetUserTierInt(ctx sdk.Context, addr sdk.AccAddress) int
	// GetReputation returns the user's reputation score
	GetReputation(ctx sdk.Context, addr sdk.AccAddress) math.LegacyDec
	// RewardSuccessfulLoan rewards a user for repaying a loan on time
	RewardSuccessfulLoan(ctx sdk.Context, addr sdk.AccAddress, loanID string) error
	// PenalizeLoanDefault penalizes a user for defaulting on a loan
	PenalizeLoanDefault(ctx sdk.Context, addr sdk.AccAddress, loanID string) error
	// SlashForLoanDefault slashes a user's stake for loan default
	SlashForLoanDefault(ctx sdk.Context, addr sdk.AccAddress, defaultAmount math.Int) error
}

// Stake tier constants matching x/staking/types
const (
	StakeTierHolder    = 0 // Cannot lend/borrow
	StakeTierKeeper    = 1 // Can lend/borrow
	StakeTierWarden    = 2
	StakeTierSteward   = 3
	StakeTierArchon    = 4
	StakeTierValidator = 5
)
