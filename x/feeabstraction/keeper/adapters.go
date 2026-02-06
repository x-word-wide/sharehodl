package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dexkeeper "github.com/sharehodl/sharehodl-blockchain/x/dex/keeper"
	equitykeeper "github.com/sharehodl/sharehodl-blockchain/x/equity/keeper"
	equitytypes "github.com/sharehodl/sharehodl-blockchain/x/equity/types"
	"github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/types"
)

// =============================================================================
// DEX KEEPER ADAPTER
// =============================================================================

// DEXKeeperAdapter adapts the DEX keeper to the types.DEXKeeper interface
type DEXKeeperAdapter struct {
	keeper *dexkeeper.Keeper
}

// NewDEXKeeperAdapter creates a new DEX keeper adapter
func NewDEXKeeperAdapter(keeper *dexkeeper.Keeper) *DEXKeeperAdapter {
	return &DEXKeeperAdapter{keeper: keeper}
}

// GetMarket returns a market for trading between two tokens
func (a *DEXKeeperAdapter) GetMarket(ctx sdk.Context, baseSymbol, quoteSymbol string) (types.Market, bool) {
	market, found := a.keeper.GetMarket(ctx, baseSymbol, quoteSymbol)
	if !found {
		return types.Market{}, false
	}
	return types.Market{
		BaseSymbol:    market.BaseSymbol,
		QuoteSymbol:   market.QuoteSymbol,
		Active:        market.Active,
		TradingHalted: market.TradingHalted,
		LastPrice:     market.LastPrice,
	}, true
}

// ExecuteSwap executes a swap from one token to another
func (a *DEXKeeperAdapter) ExecuteSwap(ctx sdk.Context, trader sdk.AccAddress, fromDenom string, fromAmount math.Int, toDenom string, minReceived math.Int) (math.Int, error) {
	// Calculate max slippage from minReceived
	// Get current rate to determine expected output
	exchangeRate, err := a.keeper.GetAtomicSwapRate(ctx, fromDenom, toDenom)
	if err != nil {
		return math.ZeroInt(), err
	}
	expectedOutput := exchangeRate.MulInt(fromAmount).TruncateInt()

	// Calculate slippage tolerance
	var maxSlippage math.LegacyDec
	if expectedOutput.IsZero() {
		maxSlippage = math.LegacyNewDecWithPrec(5, 2) // 5% default
	} else {
		maxSlippage = math.LegacyNewDecFromInt(expectedOutput.Sub(minReceived)).
			Quo(math.LegacyNewDecFromInt(expectedOutput))
		if maxSlippage.IsNegative() {
			maxSlippage = math.LegacyZeroDec()
		}
	}

	// Execute via DEX atomic swap
	order, err := a.keeper.ExecuteAtomicSwap(ctx, trader, fromDenom, toDenom, fromAmount, maxSlippage)
	if err != nil {
		return math.ZeroInt(), err
	}

	// Return the filled amount
	return order.FilledQuantity, nil
}

// GetMarketPrice returns the current price of a token in HODL
func (a *DEXKeeperAdapter) GetMarketPrice(ctx sdk.Context, symbol string) (math.LegacyDec, bool) {
	market, found := a.keeper.GetMarket(ctx, symbol, "HODL")
	if !found {
		return math.LegacyZeroDec(), false
	}
	if market.LastPrice.IsZero() {
		return math.LegacyZeroDec(), false
	}
	return market.LastPrice, true
}

// Verify interface implementation
var _ types.DEXKeeper = (*DEXKeeperAdapter)(nil)

// =============================================================================
// EQUITY KEEPER ADAPTER
// =============================================================================

// EquityKeeperAdapter adapts the Equity keeper to the types.EquityKeeper interface
type EquityKeeperAdapter struct {
	keeper equitykeeper.Keeper
}

// NewEquityKeeperAdapter creates a new Equity keeper adapter
func NewEquityKeeperAdapter(keeper equitykeeper.Keeper) *EquityKeeperAdapter {
	return &EquityKeeperAdapter{keeper: keeper}
}

// IsEquityToken checks if a denom is a registered equity token
func (a *EquityKeeperAdapter) IsEquityToken(ctx sdk.Context, denom string) bool {
	// Check if the denom is a company equity token
	// Search through all companies to find a matching symbol
	companies := a.keeper.GetAllCompanies(ctx)
	for _, company := range companies {
		if company.Symbol == denom && company.Status == equitytypes.CompanyStatusActive {
			return true
		}
	}
	return false
}

// GetEquityInfo returns information about an equity token
func (a *EquityKeeperAdapter) GetEquityInfo(ctx sdk.Context, symbol string) (types.EquityInfo, bool) {
	// Search through all companies to find a matching symbol
	companies := a.keeper.GetAllCompanies(ctx)
	for _, company := range companies {
		if company.Symbol == symbol {
			return types.EquityInfo{
				Symbol:      company.Symbol,
				CompanyID:   company.ID,
				Active:      company.Status == equitytypes.CompanyStatusActive,
				TotalSupply: company.TotalShares,
			}, true
		}
	}
	return types.EquityInfo{}, false
}

// Verify interface implementation
var _ types.EquityKeeper = (*EquityKeeperAdapter)(nil)
