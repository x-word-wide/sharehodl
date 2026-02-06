package types

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, name string) sdk.ModuleAccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin

	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error

	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error

	GetSupply(ctx context.Context, denom string) sdk.Coin
}

// StakingKeeper defines the expected interface for validator tier checks
type StakingKeeper interface {
	// GetValidatorTier returns the staking tier for a validator
	GetValidatorTier(ctx sdk.Context, validator sdk.AccAddress) (int32, error)

	// IsValidator checks if an address is a validator
	IsValidator(ctx sdk.Context, addr sdk.AccAddress) bool

	// GetTotalValidators returns the total number of active validators
	GetTotalValidators(ctx sdk.Context) uint64

	// GetValidatorsByTier returns validators at or above a certain tier
	GetValidatorsByTier(ctx sdk.Context, minTier int32) ([]sdk.AccAddress, error)
}

// EscrowKeeper defines the expected interface for ban/blacklist checking
type EscrowKeeper interface {
	// IsAddressBanned checks if an address is banned from the platform
	// Returns (isBanned, reason) - reason explains why the address is banned
	IsAddressBanned(ctx sdk.Context, address string, blockTime time.Time) (bool, string)
}
