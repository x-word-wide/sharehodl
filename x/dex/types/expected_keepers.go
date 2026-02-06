package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// AccountKeeper defines the expected account keeper used for simulations
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	GetModuleAddress(name string) sdk.AccAddress
}

// EquityKeeper defines the expected interface for interacting with equity tokens
type EquityKeeper interface {
	GetCompany(ctx context.Context, companyID uint64) (interface{}, bool)
	GetShareClass(ctx context.Context, companyID uint64, classID string) (interface{}, bool)
	GetShareholding(ctx context.Context, companyID uint64, classID, owner string) (interface{}, bool)
	IsSymbolTaken(ctx context.Context, symbol string) bool

	// Beneficial owner registry for dividend tracking when equity is locked in DEX orders
	RegisterBeneficialOwner(ctx sdk.Context, moduleAccount string, companyID uint64, classID string, beneficialOwner string, shares math.Int, referenceID uint64, referenceType string) error
	UnregisterBeneficialOwner(ctx sdk.Context, moduleAccount string, companyID uint64, classID string, beneficialOwner string, referenceID uint64) error
	UpdateBeneficialOwnerShares(ctx sdk.Context, moduleAccount string, companyID uint64, classID string, beneficialOwner string, referenceID uint64, newShares math.Int) error

	// Helper to get company ID from trading symbol
	GetCompanyBySymbol(ctx sdk.Context, symbol string) (uint64, bool)

	// Trading halt integration - check if trading is halted for a company
	IsTradingHalted(ctx sdk.Context, companyID uint64) bool

	// Blacklist integration - check if an address is blacklisted from trading a company's shares
	IsBlacklisted(ctx sdk.Context, companyID uint64, address string) bool
}

// HODLKeeper defines the expected interface for interacting with HODL stablecoin
type HODLKeeper interface {
	GetTotalSupply(ctx context.Context) interface{}
}