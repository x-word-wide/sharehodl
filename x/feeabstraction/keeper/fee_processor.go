package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/types"
)

// ProcessFeeAbstraction is the main entry point for fee abstraction
// It handles the complete flow: validate -> select equity -> swap -> pay fee -> return remainder
func (k Keeper) ProcessFeeAbstraction(
	ctx sdk.Context,
	payer sdk.AccAddress,
	requiredFeeUhodl math.Int,
) (processedFee math.Int, err error) {
	// 1. Validate fee abstraction is enabled
	params := k.GetParams(ctx)
	if !params.Enabled {
		return math.ZeroInt(), types.ErrFeeAbstractionDisabled
	}

	// 2. Check block limit not exceeded (DoS protection)
	if err := k.CheckBlockLimit(ctx, requiredFeeUhodl); err != nil {
		return math.ZeroInt(), err
	}

	// 3. Calculate total required (fee + markup)
	totalRequired, markup := k.CalculateTotalWithMarkup(ctx, requiredFeeUhodl)

	// 4. Get user's eligible equity value
	totalEquityValue, eligibleEquities, err := k.GetUserEquityValue(ctx, payer)
	if err != nil {
		return math.ZeroInt(), err
	}

	// 5. Validate sufficient equity value
	if totalEquityValue.LT(totalRequired) {
		return math.ZeroInt(), sdkerrors.Wrapf(
			types.ErrInsufficientEquityValue,
			"equity value %s < required %s (fee %s + markup %s)",
			totalEquityValue, totalRequired, requiredFeeUhodl, markup,
		)
	}

	// 6. Select optimal equity to swap
	equitySymbol, equityAmount, _, err := k.SelectOptimalEquity(
		ctx, payer, totalRequired, eligibleEquities,
	)
	if err != nil {
		return math.ZeroInt(), err
	}

	// 7. Execute atomic swap via DEX
	hodlReceived, err := k.ExecuteEquitySwap(ctx, payer, equitySymbol, equityAmount)
	if err != nil {
		return math.ZeroInt(), sdkerrors.Wrap(err, "equity swap failed")
	}

	// 8. Pay gas fee from swapped HODL
	if err := k.PayGasFee(ctx, payer, requiredFeeUhodl); err != nil {
		return math.ZeroInt(), err
	}

	// 9. Return remainder to protocol treasury
	remainder := hodlReceived.Sub(requiredFeeUhodl)
	if remainder.IsPositive() {
		if err := k.ReturnToTreasury(ctx, payer, remainder); err != nil {
			return math.ZeroInt(), err
		}
	}

	// 10. Record transaction and update block usage
	k.RecordFeeAbstraction(ctx, payer, equitySymbol, equityAmount,
		hodlReceived, requiredFeeUhodl, markup, remainder)
	k.IncrementBlockUsage(ctx, totalRequired)

	// 11. Emit events
	k.EmitFeeAbstractionEvent(ctx, payer, equitySymbol, requiredFeeUhodl, markup)

	return totalRequired, nil
}

// CalculateTotalWithMarkup returns fee + markup
func (k Keeper) CalculateTotalWithMarkup(
	ctx sdk.Context,
	baseFee math.Int,
) (total math.Int, markup math.Int) {
	params := k.GetParams(ctx)
	// markup = baseFee * markupBasisPoints / 10000
	markupDec := math.LegacyNewDecFromInt(baseFee).
		MulInt64(int64(params.FeeMarkupBasisPoints)).
		QuoInt64(10000)
	markup = markupDec.TruncateInt()
	total = baseFee.Add(markup)
	return
}

// GetUserEquityValue calculates total equity value and returns eligible equities
func (k Keeper) GetUserEquityValue(
	ctx sdk.Context,
	user sdk.AccAddress,
) (totalValue math.Int, eligibleEquities []types.EligibleEquity, err error) {
	allBalances := k.bankKeeper.GetAllBalances(ctx, user)
	totalValue = math.ZeroInt()
	eligibleEquities = []types.EligibleEquity{}

	for _, balance := range allBalances {
		// Skip native HODL token
		if balance.Denom == "uhodl" || balance.Denom == "hodl" {
			continue
		}

		// Check if this is an equity token
		if !k.equityKeeper.IsEquityToken(ctx, balance.Denom) {
			continue
		}

		// Check if equity is enabled for fee abstraction
		if !k.IsEquityEnabled(ctx, balance.Denom) {
			continue
		}

		// Get market price
		price, found := k.dexKeeper.GetMarketPrice(ctx, balance.Denom)
		if !found || price.IsZero() {
			continue
		}

		// Check if market is active
		market, found := k.dexKeeper.GetMarket(ctx, balance.Denom, "HODL")
		if !found || !market.Active || market.TradingHalted {
			continue
		}

		// Calculate HODL value
		hodlValue := price.MulInt(balance.Amount).TruncateInt()
		if hodlValue.IsZero() {
			continue
		}

		totalValue = totalValue.Add(hodlValue)
		eligibleEquities = append(eligibleEquities, types.EligibleEquity{
			Symbol:       balance.Denom,
			Balance:      balance.Amount,
			HODLValue:    hodlValue,
			Price:        price,
			MarketActive: market.Active && !market.TradingHalted,
		})
	}

	if len(eligibleEquities) == 0 {
		return math.ZeroInt(), nil, types.ErrNoEligibleEquity
	}

	return totalValue, eligibleEquities, nil
}

// SelectOptimalEquity selects the best equity to swap for covering the fee
// Strategy: Use the equity with the highest liquidity and smallest impact
func (k Keeper) SelectOptimalEquity(
	ctx sdk.Context,
	user sdk.AccAddress,
	amountNeeded math.Int,
	eligibleEquities []types.EligibleEquity,
) (symbol string, amount math.Int, expectedHODL math.Int, err error) {
	if len(eligibleEquities) == 0 {
		return "", math.ZeroInt(), math.ZeroInt(), types.ErrNoEligibleEquity
	}

	params := k.GetParams(ctx)
	maxSlippage := params.GetMaxSlippageMultiplier()

	// Find the best equity - prefer one that can cover the full amount
	var bestEquity *types.EligibleEquity
	for i := range eligibleEquities {
		eq := &eligibleEquities[i]
		if eq.HODLValue.GTE(amountNeeded) {
			if bestEquity == nil || eq.HODLValue.GT(bestEquity.HODLValue) {
				bestEquity = eq
			}
		}
	}

	// If no single equity can cover it, use the one with highest value
	if bestEquity == nil {
		for i := range eligibleEquities {
			eq := &eligibleEquities[i]
			if bestEquity == nil || eq.HODLValue.GT(bestEquity.HODLValue) {
				bestEquity = eq
			}
		}
	}

	if bestEquity == nil {
		return "", math.ZeroInt(), math.ZeroInt(), types.ErrNoEligibleEquity
	}

	// Calculate equity amount needed (with slippage buffer)
	// equityNeeded = amountNeeded / price * (1 + slippage)
	slippageMultiplier := math.LegacyOneDec().Add(maxSlippage)
	equityNeededDec := math.LegacyNewDecFromInt(amountNeeded).
		Quo(bestEquity.Price).
		Mul(slippageMultiplier)
	equityNeeded := equityNeededDec.Ceil().TruncateInt()

	// Cap at user's balance
	if equityNeeded.GT(bestEquity.Balance) {
		equityNeeded = bestEquity.Balance
	}

	// Calculate expected HODL output
	expectedHODL = bestEquity.Price.MulInt(equityNeeded).TruncateInt()

	return bestEquity.Symbol, equityNeeded, expectedHODL, nil
}

// ExecuteEquitySwap executes the swap from equity to HODL
func (k Keeper) ExecuteEquitySwap(
	ctx sdk.Context,
	trader sdk.AccAddress,
	equitySymbol string,
	equityAmount math.Int,
) (hodlReceived math.Int, err error) {
	params := k.GetParams(ctx)
	maxSlippage := params.GetMaxSlippageMultiplier()

	// Get expected output
	price, found := k.dexKeeper.GetMarketPrice(ctx, equitySymbol)
	if !found {
		return math.ZeroInt(), types.ErrPriceUnavailable
	}

	expectedOutput := price.MulInt(equityAmount).TruncateInt()
	minReceived := math.LegacyNewDecFromInt(expectedOutput).
		Mul(math.LegacyOneDec().Sub(maxSlippage)).
		TruncateInt()

	// Execute swap via DEX
	hodlReceived, err = k.dexKeeper.ExecuteSwap(
		ctx,
		trader,
		equitySymbol,
		equityAmount,
		"uhodl",
		minReceived,
	)
	if err != nil {
		return math.ZeroInt(), sdkerrors.Wrap(types.ErrSwapFailed, err.Error())
	}

	// Verify slippage
	if hodlReceived.LT(minReceived) {
		return math.ZeroInt(), types.ErrSlippageExceeded
	}

	return hodlReceived, nil
}

// PayGasFee pays the gas fee from the user's HODL balance
func (k Keeper) PayGasFee(ctx sdk.Context, payer sdk.AccAddress, feeAmount math.Int) error {
	feeCoins := sdk.NewCoins(sdk.NewCoin("uhodl", feeAmount))

	// Send fee to fee collector module
	return k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		payer,
		"fee_collector",
		feeCoins,
	)
}

// ReturnToTreasury returns the remainder to the protocol treasury
func (k Keeper) ReturnToTreasury(ctx sdk.Context, from sdk.AccAddress, amount math.Int) error {
	if amount.IsZero() || amount.IsNegative() {
		return nil
	}

	treasuryCoins := sdk.NewCoins(sdk.NewCoin("uhodl", amount))
	return k.bankKeeper.SendCoinsFromAccountToModule(
		ctx,
		from,
		types.TreasuryPoolName,
		treasuryCoins,
	)
}

// RecordFeeAbstraction saves an audit record of the fee abstraction
func (k Keeper) RecordFeeAbstraction(
	ctx sdk.Context,
	user sdk.AccAddress,
	equitySymbol string,
	equitySwapped, hodlReceived, gasFee, markup, treasuryReturned math.Int,
) {
	record := types.FeeAbstractionRecord{
		ID:               k.GetNextRecordID(ctx),
		User:             user.String(),
		EquitySymbol:     equitySymbol,
		EquitySwapped:    equitySwapped,
		HODLReceived:     hodlReceived,
		GasFeePaid:       gasFee,
		MarkupPaid:       markup,
		TreasuryReturned: treasuryReturned,
		BlockHeight:      uint64(ctx.BlockHeight()),
		Timestamp:        ctx.BlockTime(),
	}

	k.SaveRecord(ctx, record)
}

// EmitFeeAbstractionEvent emits an event for fee abstraction
func (k Keeper) EmitFeeAbstractionEvent(
	ctx sdk.Context,
	payer sdk.AccAddress,
	equitySymbol string,
	feeAmount, markup math.Int,
) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"fee_abstraction",
			sdk.NewAttribute("payer", payer.String()),
			sdk.NewAttribute("equity_symbol", equitySymbol),
			sdk.NewAttribute("fee_amount", feeAmount.String()),
			sdk.NewAttribute("markup", markup.String()),
		),
	)
}
