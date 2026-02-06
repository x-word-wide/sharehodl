package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
)

const (
	HODLDenom = "uhodl"
)

// SwapIn swaps external asset for HODL
// User deposits external asset -> receives HODL
func (k Keeper) SwapIn(ctx sdk.Context, user sdk.AccAddress, inputDenom string, inputAmount math.Int) (math.Int, math.Int, error) {
	// Check if bridge is enabled
	params := k.GetParams(ctx)
	if !params.BridgeEnabled {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("bridge is currently disabled")
	}

	// Validate asset is supported
	asset, found := k.GetSupportedAsset(ctx, inputDenom)
	if !found {
		return math.ZeroInt(), math.ZeroInt(), types.ErrAssetNotSupported
	}

	if !asset.Enabled {
		return math.ZeroInt(), math.ZeroInt(), types.ErrAssetNotSupported
	}

	// Validate swap amount
	if inputAmount.LT(asset.MinSwapAmount) {
		return math.ZeroInt(), math.ZeroInt(), types.ErrInvalidSwapAmount
	}

	if !asset.MaxSwapAmount.IsZero() && inputAmount.GT(asset.MaxSwapAmount) {
		return math.ZeroInt(), math.ZeroInt(), types.ErrInvalidSwapAmount
	}

	// Calculate HODL output amount (before fees)
	hodlAmount := k.CalculateSwapAmount(inputAmount, asset.SwapRateToHODL)

	// Calculate and deduct swap fee
	fee := k.CalculateSwapFee(ctx, hodlAmount)
	hodlAfterFee := hodlAmount.Sub(fee)

	if hodlAfterFee.IsNegative() || hodlAfterFee.IsZero() {
		return math.ZeroInt(), math.ZeroInt(), types.ErrInvalidSwapAmount
	}

	// Transfer input asset from user to bridge module
	inputCoin := sdk.NewCoin(inputDenom, inputAmount)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, user, types.ModuleName, sdk.NewCoins(inputCoin)); err != nil {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("failed to transfer input asset: %w", err)
	}

	// Add input asset to reserve
	k.AddToReserve(ctx, inputDenom, inputAmount)

	// Mint HODL to user
	hodlCoin := sdk.NewCoin(HODLDenom, hodlAfterFee)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(hodlCoin)); err != nil {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("failed to mint HODL: %w", err)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, user, sdk.NewCoins(hodlCoin)); err != nil {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("failed to send HODL to user: %w", err)
	}

	// Record swap
	swapID := k.GetNextSwapID(ctx)
	record := types.NewSwapRecord(
		swapID,
		user.String(),
		types.SwapDirectionIn,
		inputDenom,
		inputAmount,
		HODLDenom,
		hodlAfterFee,
		fee,
		asset.SwapRateToHODL,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)
	k.SetSwapRecord(ctx, record)
	k.SetNextSwapID(ctx, swapID+1)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_swap_in",
			sdk.NewAttribute("user", user.String()),
			sdk.NewAttribute("input_denom", inputDenom),
			sdk.NewAttribute("input_amount", inputAmount.String()),
			sdk.NewAttribute("hodl_received", hodlAfterFee.String()),
			sdk.NewAttribute("fee", fee.String()),
			sdk.NewAttribute("swap_id", fmt.Sprintf("%d", swapID)),
		),
	)

	k.Logger(ctx).Info("swap in completed",
		"user", user.String(),
		"input", inputAmount.String()+inputDenom,
		"output", hodlAfterFee.String()+HODLDenom,
		"fee", fee.String(),
	)

	return hodlAfterFee, fee, nil
}

// SwapOut swaps HODL for external asset
// User deposits HODL -> receives external asset
func (k Keeper) SwapOut(ctx sdk.Context, user sdk.AccAddress, hodlAmount math.Int, outputDenom string) (math.Int, math.Int, error) {
	// Check if bridge is enabled
	params := k.GetParams(ctx)
	if !params.BridgeEnabled {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("bridge is currently disabled")
	}

	// Validate asset is supported
	asset, found := k.GetSupportedAsset(ctx, outputDenom)
	if !found {
		return math.ZeroInt(), math.ZeroInt(), types.ErrAssetNotSupported
	}

	if !asset.Enabled {
		return math.ZeroInt(), math.ZeroInt(), types.ErrAssetNotSupported
	}

	// Calculate output amount (before fees)
	outputAmount := k.CalculateSwapAmount(hodlAmount, asset.SwapRateFromHODL)

	// Calculate and deduct swap fee (in output asset terms)
	fee := k.CalculateSwapFee(ctx, outputAmount)
	outputAfterFee := outputAmount.Sub(fee)

	if outputAfterFee.IsNegative() || outputAfterFee.IsZero() {
		return math.ZeroInt(), math.ZeroInt(), types.ErrInvalidSwapAmount
	}

	// Validate swap amount
	if outputAfterFee.LT(asset.MinSwapAmount) {
		return math.ZeroInt(), math.ZeroInt(), types.ErrInvalidSwapAmount
	}

	if !asset.MaxSwapAmount.IsZero() && outputAfterFee.GT(asset.MaxSwapAmount) {
		return math.ZeroInt(), math.ZeroInt(), types.ErrInvalidSwapAmount
	}

	// Check reserve has enough balance
	reserveBalance := k.GetReserveBalance(ctx, outputDenom)
	if reserveBalance.LT(outputAfterFee) {
		return math.ZeroInt(), math.ZeroInt(), types.ErrInsufficientReserve
	}

	// Check reserve ratio after withdrawal
	if err := k.CheckReserveRatioAfterWithdrawal(ctx, outputDenom, outputAfterFee); err != nil {
		return math.ZeroInt(), math.ZeroInt(), err
	}

	// Burn HODL from user
	hodlCoin := sdk.NewCoin(HODLDenom, hodlAmount)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, user, types.ModuleName, sdk.NewCoins(hodlCoin)); err != nil {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("failed to receive HODL from user: %w", err)
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(hodlCoin)); err != nil {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("failed to burn HODL: %w", err)
	}

	// Subtract from reserve
	if err := k.SubtractFromReserve(ctx, outputDenom, outputAfterFee); err != nil {
		return math.ZeroInt(), math.ZeroInt(), err
	}

	// Send output asset to user
	outputCoin := sdk.NewCoin(outputDenom, outputAfterFee)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, user, sdk.NewCoins(outputCoin)); err != nil {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("failed to send output asset to user: %w", err)
	}

	// Record swap
	swapID := k.GetNextSwapID(ctx)
	record := types.NewSwapRecord(
		swapID,
		user.String(),
		types.SwapDirectionOut,
		HODLDenom,
		hodlAmount,
		outputDenom,
		outputAfterFee,
		fee,
		asset.SwapRateFromHODL,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)
	k.SetSwapRecord(ctx, record)
	k.SetNextSwapID(ctx, swapID+1)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_swap_out",
			sdk.NewAttribute("user", user.String()),
			sdk.NewAttribute("hodl_amount", hodlAmount.String()),
			sdk.NewAttribute("output_denom", outputDenom),
			sdk.NewAttribute("output_received", outputAfterFee.String()),
			sdk.NewAttribute("fee", fee.String()),
			sdk.NewAttribute("swap_id", fmt.Sprintf("%d", swapID)),
		),
	)

	k.Logger(ctx).Info("swap out completed",
		"user", user.String(),
		"input", hodlAmount.String()+HODLDenom,
		"output", outputAfterFee.String()+outputDenom,
		"fee", fee.String(),
	)

	return outputAfterFee, fee, nil
}

// CalculateSwapAmount calculates the output amount for a swap
func (k Keeper) CalculateSwapAmount(inputAmount math.Int, rate math.LegacyDec) math.Int {
	return rate.MulInt(inputAmount).TruncateInt()
}

// CalculateSwapFee calculates the swap fee
func (k Keeper) CalculateSwapFee(ctx sdk.Context, amount math.Int) math.Int {
	params := k.GetParams(ctx)
	return params.SwapFeeRate.MulInt(amount).TruncateInt()
}
