package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToAccount(ctx sdk.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// AccountKeeper defines the expected account keeper used for simulations
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx sdk.Context, acc sdk.AccountI)
}

// EquityKeeper defines the expected interface for interacting with equity tokens
type EquityKeeper interface {
	GetCompany(ctx sdk.Context, companyID uint64) (interface{}, bool)
	GetShareClass(ctx sdk.Context, companyID uint64, classID string) (interface{}, bool)
	GetShareholding(ctx sdk.Context, companyID uint64, classID, owner string) (interface{}, bool)
	IsSymbolTaken(ctx sdk.Context, symbol string) bool
}

// HODLKeeper defines the expected interface for interacting with HODL stablecoin
type HODLKeeper interface {
	GetTotalSupply(ctx sdk.Context) interface{}
}