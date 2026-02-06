package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/dex/types"
)

// FeeHandler manages fee collection and auto-swap from equity
type FeeHandler struct {
	keeper *Keeper
}

// NewFeeHandler creates a new fee handler
func NewFeeHandler(k *Keeper) *FeeHandler {
	return &FeeHandler{keeper: k}
}

// FeeConfig holds fee configuration (derived from governance params)
type FeeConfig struct {
	BaseFee          math.Int       // Base transaction fee in uhodl
	FeePerByte       math.Int       // Fee per byte of transaction
	MinFee           math.Int       // Minimum fee
	MaxFee           math.Int       // Maximum fee
	AutoSwapEnabled  bool           // Whether auto-swap from equity is enabled
	AutoSwapSlippage math.LegacyDec // Maximum slippage for auto-swap
}

// GetFeeConfig returns fee configuration from governance-controllable params
func (k Keeper) GetFeeConfig(ctx sdk.Context) FeeConfig {
	params := k.GetParams(ctx)
	return FeeConfig{
		BaseFee:          math.NewInt(int64(params.BaseFee)),
		FeePerByte:       math.NewInt(int64(params.FeePerByte)),
		MinFee:           math.NewInt(int64(params.MinFee)),
		MaxFee:           math.NewInt(int64(params.MaxFee)),
		AutoSwapEnabled:  params.TradingEnabled, // Use trading enabled flag
		AutoSwapSlippage: params.AutoSwapSlippage,
	}
}

// DefaultFeeConfig returns default fee configuration (for backwards compatibility)
func DefaultFeeConfig() FeeConfig {
	return FeeConfig{
		BaseFee:          math.NewInt(int64(types.DefaultBaseFee)),
		FeePerByte:       math.NewInt(int64(types.DefaultFeePerByte)),
		MinFee:           math.NewInt(int64(types.DefaultMinFee)),
		MaxFee:           math.NewInt(int64(types.DefaultMaxFee)),
		AutoSwapEnabled:  true,
		AutoSwapSlippage: types.DefaultAutoSwapSlippage,
	}
}

// CalculateTransactionFee calculates the fee for a transaction using governance params
func (k Keeper) CalculateTransactionFee(ctx sdk.Context, txSize int) math.Int {
	config := k.GetFeeConfig(ctx)

	// Calculate fee: base + (size * perByte)
	sizeFee := config.FeePerByte.MulRaw(int64(txSize))
	totalFee := config.BaseFee.Add(sizeFee)

	// Clamp to min/max
	if totalFee.LT(config.MinFee) {
		totalFee = config.MinFee
	}
	if totalFee.GT(config.MaxFee) {
		totalFee = config.MaxFee
	}

	return totalFee
}

// DeductFeeWithAutoSwap deducts fees from user, auto-swapping equity if needed
// This is the key function that allows users without HODL to transact
func (k Keeper) DeductFeeWithAutoSwap(
	ctx sdk.Context,
	payer sdk.AccAddress,
	feeAmount math.Int,
) error {
	// First, check if user has enough HODL
	hodlBalance := k.bankKeeper.GetBalance(ctx, payer, "hodl")

	if hodlBalance.Amount.GTE(feeAmount) {
		// User has enough HODL, pay directly
		feeCoins := sdk.NewCoins(sdk.NewCoin("hodl", feeAmount))
		return k.bankKeeper.SendCoinsFromAccountToModule(ctx, payer, types.ModuleName, feeCoins)
	}

	// User doesn't have enough HODL, try auto-swap from equity
	config := k.GetFeeConfig(ctx)
	if !config.AutoSwapEnabled {
		return types.ErrInsufficientFunds
	}

	// Calculate HODL shortfall
	hodlShortfall := feeAmount.Sub(hodlBalance.Amount)

	// Get all user's equity holdings
	allBalances := k.bankKeeper.GetAllBalances(ctx, payer)

	// Try to swap equity for the shortfall
	for _, balance := range allBalances {
		// Skip HODL itself
		if balance.Denom == "hodl" {
			continue
		}

		// Check if this is an equity token (has a market with HODL)
		market, found := k.GetMarket(ctx, balance.Denom, "HODL")
		if !found {
			continue
		}

		if !market.Active || market.TradingHalted {
			continue
		}

		// Calculate how much equity we need to sell
		equityPrice := market.LastPrice
		if equityPrice.IsZero() {
			continue
		}

		// Amount of equity needed = hodlShortfall / equityPrice
		// Add slippage buffer
		slippageMultiplier := math.LegacyOneDec().Add(config.AutoSwapSlippage)
		equityNeeded := math.LegacyNewDecFromInt(hodlShortfall).Quo(equityPrice).Mul(slippageMultiplier)
		equityNeededInt := equityNeeded.TruncateInt()

		// Check if user has enough of this equity
		if balance.Amount.LT(equityNeededInt) {
			// Try with what they have
			equityNeededInt = balance.Amount
		}

		// Execute the swap
		swapResult, err := k.executeEquityToHODLSwap(ctx, payer, balance.Denom, equityNeededInt)
		if err != nil {
			k.Logger(ctx).Debug("auto-swap failed", "equity", balance.Denom, "error", err)
			continue
		}

		// Add swap result to available HODL
		hodlBalance.Amount = hodlBalance.Amount.Add(swapResult)

		// Check if we now have enough
		if hodlBalance.Amount.GTE(feeAmount) {
			break
		}

		// Update remaining shortfall
		hodlShortfall = feeAmount.Sub(hodlBalance.Amount)
	}

	// Final check - do we have enough HODL now?
	if hodlBalance.Amount.LT(feeAmount) {
		return fmt.Errorf("insufficient funds after auto-swap: have %s, need %s",
			hodlBalance.Amount.String(), feeAmount.String())
	}

	// Deduct the fee
	feeCoins := sdk.NewCoins(sdk.NewCoin("hodl", feeAmount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, payer, types.ModuleName, feeCoins); err != nil {
		return err
	}

	// Emit auto-swap event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"fee_auto_swap",
			sdk.NewAttribute("payer", payer.String()),
			sdk.NewAttribute("fee_amount", feeAmount.String()),
			sdk.NewAttribute("auto_swapped", "true"),
		),
	)

	return nil
}

// executeEquityToHODLSwap executes a swap from equity to HODL for fee payment
func (k Keeper) executeEquityToHODLSwap(
	ctx sdk.Context,
	trader sdk.AccAddress,
	equitySymbol string,
	equityAmount math.Int,
) (math.Int, error) {
	// Get current market rate
	market, found := k.GetMarket(ctx, equitySymbol, "HODL")
	if !found {
		return math.ZeroInt(), types.ErrMarketNotFound
	}

	// Calculate HODL output
	hodlOutput := market.LastPrice.MulInt(equityAmount)

	// Apply swap fee (governance-controllable, default 0.3%)
	swapFeeRate := k.GetEquitySwapFee(ctx)
	swapFee := hodlOutput.Mul(swapFeeRate)
	hodlOutput = hodlOutput.Sub(swapFee)

	hodlOutputInt := hodlOutput.TruncateInt()
	if hodlOutputInt.IsZero() {
		return math.ZeroInt(), fmt.Errorf("swap output too small")
	}

	// Transfer equity from trader to module (this equity becomes part of DEX reserves)
	equityCoins := sdk.NewCoins(sdk.NewCoin(equitySymbol, equityAmount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, trader, types.ModuleName, equityCoins); err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to transfer equity: %w", err)
	}

	// SECURITY FIX: Instead of minting new HODL (which violates collateralization),
	// we use HODL from the DEX module's liquidity reserves.
	// The module must have sufficient HODL balance from trading fees and liquidity provision.
	hodlCoins := sdk.NewCoins(sdk.NewCoin("uhodl", hodlOutputInt))

	// Check module has sufficient HODL balance
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	moduleBalance := k.bankKeeper.GetBalance(ctx, moduleAddr, "uhodl")
	if moduleBalance.Amount.LT(hodlOutputInt) {
		// Rollback equity transfer
		k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, trader, equityCoins)
		return math.ZeroInt(), fmt.Errorf("insufficient HODL liquidity in DEX module for auto-swap")
	}

	// Send HODL from module reserves to trader
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, trader, hodlCoins); err != nil {
		// Rollback equity transfer
		k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, trader, equityCoins)
		return math.ZeroInt(), fmt.Errorf("failed to send HODL: %w", err)
	}

	// Update market stats
	market.Volume24h = market.Volume24h.Add(equityAmount)
	market.LastPrice = market.LastPrice // Price unchanged for auto-swap
	k.SetMarket(ctx, market)

	// Emit swap event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"equity_to_fee_swap",
			sdk.NewAttribute("trader", trader.String()),
			sdk.NewAttribute("equity_symbol", equitySymbol),
			sdk.NewAttribute("equity_amount", equityAmount.String()),
			sdk.NewAttribute("hodl_received", hodlOutputInt.String()),
			sdk.NewAttribute("fee_collected", swapFee.TruncateInt().String()),
		),
	)

	return hodlOutputInt, nil
}

// CollectTradingFees collects trading fees to module account
func (k Keeper) CollectTradingFees(
	ctx sdk.Context,
	trader sdk.AccAddress,
	tradeValue math.LegacyDec,
	isMaker bool,
	market types.Market,
) (math.Int, error) {
	var feeRate math.LegacyDec
	if isMaker {
		feeRate = market.MakerFee
	} else {
		feeRate = market.TakerFee
	}

	feeAmount := tradeValue.Mul(feeRate).TruncateInt()
	if feeAmount.IsZero() {
		return math.ZeroInt(), nil
	}

	// Try to deduct fee with auto-swap
	if err := k.DeductFeeWithAutoSwap(ctx, trader, feeAmount); err != nil {
		return math.ZeroInt(), err
	}

	return feeAmount, nil
}

// DistributeFees distributes collected fees to LP providers, validators, and treasury
// Fee distribution:
//   - 70% to LP providers (added to pool reserves, increasing LP token value)
//   - 20% to validators (for network security)
//   - 10% to treasury (for protocol development)
func (k Keeper) DistributeFees(ctx sdk.Context) {
	// Get module account balance
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	moduleBalance := k.bankKeeper.GetBalance(ctx, moduleAddr, "uhodl")

	if moduleBalance.Amount.IsZero() {
		return
	}

	// Fee distribution ratios
	lpProviderShare := moduleBalance.Amount.MulRaw(70).QuoRaw(100)                    // 70% to LP providers
	validatorShare := moduleBalance.Amount.MulRaw(20).QuoRaw(100)                     // 20% to validators
	treasuryShare := moduleBalance.Amount.Sub(lpProviderShare).Sub(validatorShare)   // 10% to treasury

	// Distribute to LP pools (added to reserves, making LP tokens more valuable)
	if lpProviderShare.IsPositive() {
		k.DistributeFeesToLPPools(ctx, lpProviderShare)
	}

	// Emit distribution event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"fees_distributed",
			sdk.NewAttribute("total", moduleBalance.Amount.String()),
			sdk.NewAttribute("lp_provider_share", lpProviderShare.String()),
			sdk.NewAttribute("validator_share", validatorShare.String()),
			sdk.NewAttribute("treasury_share", treasuryShare.String()),
		),
	)

	// Note: In production, validator and treasury shares would be sent to respective accounts
}

// DistributeFeesToLPPools distributes fees to all liquidity pools proportionally by TVL
// This increases pool reserves, making LP tokens more valuable when redeemed
func (k Keeper) DistributeFeesToLPPools(ctx sdk.Context, totalFees math.Int) {
	// Get all liquidity pools and calculate total TVL
	pools := k.GetAllLiquidityPools(ctx)
	if len(pools) == 0 {
		return
	}

	// Calculate total TVL across all pools (in quote currency, typically HODL)
	totalTVL := math.ZeroInt()
	poolTVLs := make(map[string]math.Int)

	for _, pool := range pools {
		// TVL = quote reserve (direct HODL value)
		tvl := pool.QuoteReserve
		poolTVLs[pool.MarketSymbol] = tvl
		totalTVL = totalTVL.Add(tvl)
	}

	if totalTVL.IsZero() {
		return
	}

	// Distribute fees proportionally to each pool
	for _, pool := range pools {
		poolTVL := poolTVLs[pool.MarketSymbol]
		if poolTVL.IsZero() {
			continue
		}

		// Calculate this pool's share of fees
		poolFeeShare := totalFees.Mul(poolTVL).Quo(totalTVL)
		if poolFeeShare.IsZero() {
			continue
		}

		// Add fees to pool's quote reserve (HODL)
		// This increases the value of LP tokens
		pool.QuoteReserve = pool.QuoteReserve.Add(poolFeeShare)
		pool.FeesCollected24h = pool.FeesCollected24h.Add(math.LegacyNewDecFromInt(poolFeeShare))
		pool.UpdatedAt = ctx.BlockTime()
		k.SetLiquidityPool(ctx, pool)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"lp_fees_distributed",
				sdk.NewAttribute("market_symbol", pool.MarketSymbol),
				sdk.NewAttribute("fee_amount", poolFeeShare.String()),
				sdk.NewAttribute("new_quote_reserve", pool.QuoteReserve.String()),
			),
		)
	}
}

// GetFeeStats returns fee statistics
func (k Keeper) GetFeeStats(ctx sdk.Context) map[string]interface{} {
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	moduleBalance := k.bankKeeper.GetBalance(ctx, moduleAddr, "hodl")

	return map[string]interface{}{
		"collected_fees": moduleBalance.Amount.String(),
		"auto_swap_enabled": true,
		"base_fee": DefaultFeeConfig().BaseFee.String(),
		"min_fee": DefaultFeeConfig().MinFee.String(),
		"max_fee": DefaultFeeConfig().MaxFee.String(),
	}
}
