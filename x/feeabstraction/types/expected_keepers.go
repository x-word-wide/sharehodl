package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// DEXKeeper defines the expected DEX keeper interface
type DEXKeeper interface {
	// GetMarket returns a market for trading between two tokens
	GetMarket(ctx sdk.Context, baseSymbol, quoteSymbol string) (Market, bool)

	// ExecuteSwap executes a swap from one token to another
	ExecuteSwap(ctx sdk.Context, trader sdk.AccAddress, fromDenom string, fromAmount math.Int, toDenom string, minReceived math.Int) (math.Int, error)

	// GetMarketPrice returns the current price of a token in HODL
	GetMarketPrice(ctx sdk.Context, symbol string) (math.LegacyDec, bool)
}

// Market represents a trading market (simplified interface)
type Market struct {
	BaseSymbol    string
	QuoteSymbol   string
	Active        bool
	TradingHalted bool
	LastPrice     math.LegacyDec
}

// EquityKeeper defines the expected equity keeper interface
type EquityKeeper interface {
	// IsEquityToken checks if a denom is a registered equity token
	IsEquityToken(ctx sdk.Context, denom string) bool

	// GetEquityInfo returns information about an equity token
	GetEquityInfo(ctx sdk.Context, symbol string) (EquityInfo, bool)
}

// EquityInfo represents equity token information (simplified interface)
type EquityInfo struct {
	Symbol      string
	CompanyID   uint64
	Active      bool
	TotalSupply math.Int
}

// StakingKeeper defines the expected staking keeper interface (optional)
type StakingKeeper interface {
	// GetUserStake returns a user's stake information
	GetUserStake(ctx sdk.Context, staker sdk.AccAddress) (UserStake, bool)
}

// UserStake represents a user's stake (simplified interface)
type UserStake struct {
	StakedAmount math.Int
	Tier         string
}
