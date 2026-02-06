package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
)

// ValidatorDeposit allows validators to deposit assets into the reserve
// No fee is charged for validator deposits
func (k Keeper) ValidatorDeposit(ctx sdk.Context, validator sdk.AccAddress, denom string, amount math.Int) error {
	// Verify sender is a validator
	if !k.IsValidator(ctx, validator) {
		return types.ErrNotValidator
	}

	// Check if asset is supported
	asset, found := k.GetSupportedAsset(ctx, denom)
	if !found {
		return types.ErrAssetNotSupported
	}

	if !asset.Enabled {
		return types.ErrAssetNotSupported
	}

	// Transfer asset from validator to bridge module
	coin := sdk.NewCoin(denom, amount)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, validator, types.ModuleName, sdk.NewCoins(coin)); err != nil {
		return fmt.Errorf("failed to transfer asset: %w", err)
	}

	// Add to reserve
	k.AddToReserve(ctx, denom, amount)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_validator_deposit",
			sdk.NewAttribute("validator", validator.String()),
			sdk.NewAttribute("denom", denom),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	k.Logger(ctx).Info("validator deposit completed",
		"validator", validator.String(),
		"denom", denom,
		"amount", amount.String(),
	)

	return nil
}

// GetTotalReserveValue returns the total USD value of all reserves
// This is a simplified implementation - in production would use oracle prices
func (k Keeper) GetTotalReserveValue(ctx sdk.Context) math.LegacyDec {
	totalValue := math.LegacyZeroDec()

	reserves := k.GetAllReserveBalances(ctx)
	for _, reserve := range reserves {
		asset, found := k.GetSupportedAsset(ctx, reserve.Denom)
		if !found {
			continue
		}

		// Calculate USD value using swap rate as proxy for USD value
		// In production, this would use oracle price feeds
		value := math.LegacyNewDecFromInt(reserve.Balance).Mul(asset.SwapRateToHODL)
		totalValue = totalValue.Add(value)
	}

	return totalValue
}

// CheckReserveRatio verifies that the reserve ratio meets minimum requirements
func (k Keeper) CheckReserveRatio(ctx sdk.Context) error {
	params := k.GetParams(ctx)

	// Get total reserve value
	totalReserveValue := k.GetTotalReserveValue(ctx)

	// Get total HODL supply (from bank keeper)
	hodlSupply := k.bankKeeper.GetSupply(ctx, HODLDenom)
	totalHODLSupply := math.LegacyNewDecFromInt(hodlSupply.Amount)

	// If no HODL in circulation, reserve ratio is infinite (OK)
	if totalHODLSupply.IsZero() {
		return nil
	}

	// Calculate reserve ratio: total_reserve_value / total_hodl_supply
	reserveRatio := totalReserveValue.Quo(totalHODLSupply)

	// Check if ratio meets minimum requirement
	if reserveRatio.LT(params.MinReserveRatio) {
		return types.ErrInvalidReserveRatio
	}

	return nil
}

// CheckReserveRatioAfterWithdrawal checks if reserve ratio would be maintained after a withdrawal
func (k Keeper) CheckReserveRatioAfterWithdrawal(ctx sdk.Context, denom string, amount math.Int) error {
	params := k.GetParams(ctx)

	// Get current reserve value
	currentReserveValue := k.GetTotalReserveValue(ctx)

	// Calculate value being withdrawn
	asset, found := k.GetSupportedAsset(ctx, denom)
	if !found {
		return types.ErrAssetNotSupported
	}

	withdrawalValue := math.LegacyNewDecFromInt(amount).Mul(asset.SwapRateToHODL)

	// Calculate new reserve value after withdrawal
	newReserveValue := currentReserveValue.Sub(withdrawalValue)
	if newReserveValue.IsNegative() {
		newReserveValue = math.LegacyZeroDec()
	}

	// Get total HODL supply
	hodlSupply := k.bankKeeper.GetSupply(ctx, HODLDenom)
	totalHODLSupply := math.LegacyNewDecFromInt(hodlSupply.Amount)

	// If no HODL in circulation, any withdrawal is OK
	if totalHODLSupply.IsZero() {
		return nil
	}

	// Calculate new reserve ratio
	newReserveRatio := newReserveValue.Quo(totalHODLSupply)

	// Check if ratio meets minimum requirement
	if newReserveRatio.LT(params.MinReserveRatio) {
		return types.ErrInvalidReserveRatio
	}

	return nil
}

// GetReserveRatio returns the current reserve ratio
func (k Keeper) GetReserveRatio(ctx sdk.Context) math.LegacyDec {
	totalReserveValue := k.GetTotalReserveValue(ctx)
	hodlSupply := k.bankKeeper.GetSupply(ctx, HODLDenom)
	totalHODLSupply := math.LegacyNewDecFromInt(hodlSupply.Amount)

	if totalHODLSupply.IsZero() {
		return math.LegacyNewDec(1000000) // Effectively infinite
	}

	return totalReserveValue.Quo(totalHODLSupply)
}
