package types

import (
	"context"
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
}

// EquityKeeper defines the expected interface for interacting with equity tokens
type EquityKeeper interface {
	GetCompany(ctx context.Context, companyID uint64) (interface{}, bool)
	GetShareClass(ctx context.Context, companyID uint64, classID string) (interface{}, bool)
	GetShareholding(ctx context.Context, companyID uint64, classID, owner string) (interface{}, bool)
	IsSymbolTaken(ctx context.Context, symbol string) bool
}

// HODLKeeper defines the expected interface for interacting with HODL stablecoin
type HODLKeeper interface {
	GetTotalSupply(ctx context.Context) interface{}
}