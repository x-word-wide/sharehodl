package keeper

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
)

// CheckRateLimit checks if a withdrawal would exceed rate limits
func (k Keeper) CheckRateLimit(ctx sdk.Context, chainID string, assetSymbol string, amount math.Int) error {
	params := k.GetParams(ctx)
	currentTime := ctx.BlockTime()

	// Get or create rate limit
	rateLimit := k.GetRateLimit(ctx, chainID, assetSymbol, currentTime)

	// Check if window is expired, reset if needed
	if rateLimit.IsExpired(currentTime) {
		rateLimit = types.NewRateLimit(
			chainID,
			assetSymbol,
			currentTime,
			params.RateLimitWindowDuration(),
			params.MaxWithdrawalPerWindow,
		)
	}

	// Check capacity
	if !rateLimit.HasCapacity(amount) {
		return types.ErrRateLimitExceeded
	}

	return nil
}

// UpdateRateLimit updates the rate limit after a withdrawal
func (k Keeper) UpdateRateLimit(ctx sdk.Context, chainID string, assetSymbol string, amount math.Int) {
	params := k.GetParams(ctx)
	currentTime := ctx.BlockTime()

	rateLimit := k.GetRateLimit(ctx, chainID, assetSymbol, currentTime)

	// If expired, create new window
	if rateLimit.IsExpired(currentTime) {
		rateLimit = types.NewRateLimit(
			chainID,
			assetSymbol,
			currentTime,
			params.RateLimitWindowDuration(),
			params.MaxWithdrawalPerWindow,
		)
	}

	// Update usage
	rateLimit.AmountUsed = rateLimit.AmountUsed.Add(amount)
	rateLimit.TxCount++

	k.SetRateLimit(ctx, rateLimit)
}

// GetRateLimit retrieves the current rate limit
func (k Keeper) GetRateLimit(ctx sdk.Context, chainID string, assetSymbol string, currentTime time.Time) types.RateLimit {
	store := ctx.KVStore(k.storeKey)

	// Calculate window start (aligned to rate limit window)
	params := k.GetParams(ctx)
	windowDuration := params.RateLimitWindowDuration()
	windowStart := currentTime.Truncate(windowDuration)

	bz := store.Get(types.RateLimitKey(chainID, assetSymbol, windowStart.Unix()))
	if bz == nil {
		return types.NewRateLimit(
			chainID,
			assetSymbol,
			windowStart,
			windowDuration,
			params.MaxWithdrawalPerWindow,
		)
	}

	var rateLimit types.RateLimit
	if err := json.Unmarshal(bz, &rateLimit); err != nil {
		return types.NewRateLimit(
			chainID,
			assetSymbol,
			windowStart,
			windowDuration,
			params.MaxWithdrawalPerWindow,
		)
	}
	return rateLimit
}

// SetRateLimit stores a rate limit
func (k Keeper) SetRateLimit(ctx sdk.Context, rateLimit types.RateLimit) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(rateLimit)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal rate limit", "error", err)
		return
	}
	store.Set(types.RateLimitKey(rateLimit.ChainID, rateLimit.AssetSymbol, rateLimit.WindowStart.Unix()), bz)
}

// CheckDepositRateLimit checks if a deposit would exceed rate limits
func (k Keeper) CheckDepositRateLimit(ctx sdk.Context, chainID string, assetSymbol string, amount math.Int) error {
	params := k.GetParams(ctx)
	currentTime := ctx.BlockTime()

	// Get or create rate limit for deposits (use same window as withdrawals)
	rateLimit := k.GetRateLimit(ctx, chainID, assetSymbol, currentTime)

	// Check if window is expired, reset if needed
	if rateLimit.IsExpired(currentTime) {
		rateLimit = types.NewRateLimit(
			chainID,
			assetSymbol,
			currentTime,
			params.RateLimitWindowDuration(),
			params.MaxWithdrawalPerWindow, // Using same limit for deposits
		)
	}

	// Check capacity
	if !rateLimit.HasCapacity(amount) {
		return types.ErrRateLimitExceeded
	}

	return nil
}

// UpdateDepositRateLimit updates the rate limit after a deposit
func (k Keeper) UpdateDepositRateLimit(ctx sdk.Context, chainID string, assetSymbol string, amount math.Int) {
	params := k.GetParams(ctx)
	currentTime := ctx.BlockTime()

	rateLimit := k.GetRateLimit(ctx, chainID, assetSymbol, currentTime)

	// If expired, create new window
	if rateLimit.IsExpired(currentTime) {
		rateLimit = types.NewRateLimit(
			chainID,
			assetSymbol,
			currentTime,
			params.RateLimitWindowDuration(),
			params.MaxWithdrawalPerWindow, // Using same limit for deposits
		)
	}

	// Update usage
	rateLimit.AmountUsed = rateLimit.AmountUsed.Add(amount)
	rateLimit.TxCount++

	k.SetRateLimit(ctx, rateLimit)
}
