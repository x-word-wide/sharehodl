package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(name string) sdk.AccAddress
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin

	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// EquityKeeper defines the expected equity keeper for share transfers
// Note: Uses context.Context to match the actual equity keeper implementation
type EquityKeeper interface {
	// GetShareholding returns the shareholding for an owner
	GetShareholding(ctx sdk.Context, companyID uint64, classID string, owner string) (interface{}, bool)

	// TransferShares transfers shares from one owner to another
	TransferShares(ctx sdk.Context, companyID uint64, classID string, from string, to string, shares math.Int) error

	// GetCompany returns company information (uses context.Context)
	GetCompany(ctx context.Context, companyID uint64) (interface{}, bool)

	// GetAllHoldingsByAddress returns all shareholdings for a given address
	// Returns slice of interface{} to avoid import cycles (actual type is types.AddressHolding)
	GetAllHoldingsByAddress(ctx sdk.Context, owner string) []interface{}
}

// BanKeeper defines the expected ban keeper interface
type BanKeeper interface {
	// IsAddressBanned checks if an address is banned
	IsAddressBanned(ctx sdk.Context, address string) bool
}

// HODLKeeper defines the expected HODL keeper interface
type HODLKeeper interface {
	// GetBalance returns the HODL balance for an address
	GetBalance(ctx sdk.Context, addr sdk.AccAddress) math.Int

	// SendCoins sends HODL tokens from one address to another
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount math.Int) error
}

// StakingKeeper defines the expected staking keeper interface for inheritance
// Uses adapters to handle concrete type returns
type StakingKeeper interface {
	// GetUserStake returns a user's stake (returns interface{} for adapter compatibility)
	GetUserStake(ctx sdk.Context, owner sdk.AccAddress) (interface{}, bool)

	// UnstakeForInheritance initiates unstaking with beneficiary as recipient
	UnstakeForInheritance(ctx sdk.Context, owner sdk.AccAddress, recipient string, amount math.Int) error
}

// LendingKeeper defines the expected lending keeper interface for inheritance
// Uses adapters to handle concrete type returns
type LendingKeeper interface {
	// GetUserLoans returns all loans where user is borrower or lender
	GetUserLoans(ctx sdk.Context, user string) []interface{}

	// TransferBorrowerPosition transfers a loan's borrower to a new address
	TransferBorrowerPosition(ctx sdk.Context, loanID uint64, from, to string) error

	// TransferLenderPosition transfers a lending position to a new address
	TransferLenderPosition(ctx sdk.Context, loanID uint64, from, to string) error
}

// EscrowKeeper defines the expected escrow keeper interface for inheritance
// Uses adapters to handle concrete type returns
type EscrowKeeper interface {
	// GetAllEscrows returns all escrows
	GetAllEscrows(ctx sdk.Context) []interface{}

	// TransferBuyerPosition transfers escrow buyer position for inheritance
	TransferBuyerPosition(ctx sdk.Context, escrowID uint64, from, to string) error

	// TransferSellerPosition transfers escrow seller position for inheritance
	TransferSellerPosition(ctx sdk.Context, escrowID uint64, from, to string) error
}
