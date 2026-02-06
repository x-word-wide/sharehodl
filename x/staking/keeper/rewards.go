package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/staking/types"
)

// =============================================================================
// EPOCH MANAGEMENT
// =============================================================================

// GetEpochInfo returns current epoch information
func (k Keeper) GetEpochInfo(ctx sdk.Context) types.EpochInfo {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.EpochInfoKey)
	if bz == nil {
		return types.EpochInfo{
			CurrentEpoch:     0,
			EpochStartHeight: ctx.BlockHeight(),
			EpochStartTime:   ctx.BlockTime(),
			TotalDistributed: math.ZeroInt(),
			LastDistribution: time.Time{},
		}
	}

	var info types.EpochInfo
	if err := json.Unmarshal(bz, &info); err != nil {
		return types.EpochInfo{}
	}
	return info
}

// SetEpochInfo stores epoch information
func (k Keeper) SetEpochInfo(ctx sdk.Context, info types.EpochInfo) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(info)
	store.Set(types.EpochInfoKey, bz)
}

// =============================================================================
// REWARDS DISTRIBUTION
// =============================================================================

// DistributeRewards distributes protocol revenue to all stakers
// Called from EndBlocker when epoch ends
func (k Keeper) DistributeRewards(ctx sdk.Context, treasuryBalance math.Int) error {
	params := k.GetParams(ctx)

	// Check if it's time to distribute
	if !params.ShouldDistribute(uint64(ctx.BlockHeight())) {
		return nil
	}

	// Check minimum distribution threshold
	if treasuryBalance.LT(params.MinRewardsDistribution) {
		k.Logger(ctx).Debug("treasury below minimum for distribution",
			"balance", treasuryBalance.String(),
			"minimum", params.MinRewardsDistribution.String(),
		)
		return nil
	}

	// Calculate pool amounts
	stakerPool := params.GetStakerPoolMultiplier().MulInt(treasuryBalance).TruncateInt()
	validatorPool := params.GetValidatorPoolMultiplier().MulInt(treasuryBalance).TruncateInt()
	governancePool := params.GetGovernancePoolMultiplier().MulInt(treasuryBalance).TruncateInt()

	k.Logger(ctx).Info("distributing rewards",
		"total", treasuryBalance.String(),
		"staker_pool", stakerPool.String(),
		"validator_pool", validatorPool.String(),
		"governance_pool", governancePool.String(),
	)

	// Distribute to all stakers
	stakerDistributed := k.distributeToStakers(ctx, stakerPool)

	// Distribute bonus to validators
	validatorDistributed := k.distributeToValidators(ctx, validatorPool)

	// Distribute to governance participants (TODO: integrate with governance module)
	governanceDistributed := k.distributeToGovernanceParticipants(ctx, governancePool)

	totalDistributed := stakerDistributed.Add(validatorDistributed).Add(governanceDistributed)

	// Update epoch info
	epochInfo := k.GetEpochInfo(ctx)
	epochInfo.CurrentEpoch++
	epochInfo.EpochStartHeight = ctx.BlockHeight()
	epochInfo.EpochStartTime = ctx.BlockTime()
	epochInfo.TotalDistributed = epochInfo.TotalDistributed.Add(totalDistributed)
	epochInfo.LastDistribution = ctx.BlockTime()
	k.SetEpochInfo(ctx, epochInfo)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRewardsDistributed,
			sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", epochInfo.CurrentEpoch)),
			sdk.NewAttribute(types.AttributeKeyTotalDistributed, totalDistributed.String()),
			sdk.NewAttribute(types.AttributeKeyStakerPool, stakerDistributed.String()),
			sdk.NewAttribute(types.AttributeKeyValidatorPool, validatorDistributed.String()),
			sdk.NewAttribute(types.AttributeKeyGovernancePool, governanceDistributed.String()),
		),
	)

	k.Logger(ctx).Info("rewards distributed",
		"epoch", epochInfo.CurrentEpoch,
		"total", totalDistributed.String(),
	)

	return nil
}

// distributeToStakers distributes rewards to all stakers proportionally
func (k Keeper) distributeToStakers(ctx sdk.Context, pool math.Int) math.Int {
	if pool.IsZero() {
		return math.ZeroInt()
	}

	// Calculate total weight
	totalWeight := k.GetTotalWeight(ctx)
	if totalWeight.IsZero() {
		return math.ZeroInt()
	}

	distributed := math.ZeroInt()

	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		// Calculate this staker's share
		weight := stake.CalculateWeight()
		share := weight.Quo(totalWeight).MulInt(pool).TruncateInt()

		if share.IsPositive() {
			// Add to pending rewards
			stake.PendingRewards = stake.PendingRewards.Add(share)
			k.SetUserStake(ctx, stake)
			distributed = distributed.Add(share)

			k.Logger(ctx).Debug("staker reward allocated",
				"staker", stake.Owner,
				"share", share.String(),
				"weight", weight.String(),
			)
		}
		return false
	})

	return distributed
}

// distributeToValidators distributes bonus rewards to active validators
func (k Keeper) distributeToValidators(ctx sdk.Context, pool math.Int) math.Int {
	if pool.IsZero() {
		return math.ZeroInt()
	}

	// Get total validator weight
	validatorWeight := k.GetValidatorTotalWeight(ctx)
	if validatorWeight.IsZero() {
		return math.ZeroInt()
	}

	distributed := math.ZeroInt()

	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		// Only validators get this bonus
		if stake.Tier != types.TierValidator || !stake.IsValidator {
			return false
		}

		weight := stake.CalculateWeight()
		share := weight.Quo(validatorWeight).MulInt(pool).TruncateInt()

		if share.IsPositive() {
			stake.PendingRewards = stake.PendingRewards.Add(share)
			k.SetUserStake(ctx, stake)
			distributed = distributed.Add(share)

			k.Logger(ctx).Debug("validator bonus allocated",
				"validator", stake.Owner,
				"bonus", share.String(),
			)
		}
		return false
	})

	return distributed
}

// distributeToGovernanceParticipants distributes rewards to active governance participants
func (k Keeper) distributeToGovernanceParticipants(ctx sdk.Context, pool math.Int) math.Int {
	if pool.IsZero() {
		return math.ZeroInt()
	}

	// For now, distribute governance pool equally to all stakers with good reputation
	// TODO: Integrate with governance module to track actual voters

	goodReputationWeight := math.LegacyZeroDec()
	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		if stake.ReputationScore.GTE(math.LegacyNewDec(70)) {
			goodReputationWeight = goodReputationWeight.Add(stake.CalculateWeight())
		}
		return false
	})

	if goodReputationWeight.IsZero() {
		return math.ZeroInt()
	}

	distributed := math.ZeroInt()

	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		if stake.ReputationScore.GTE(math.LegacyNewDec(70)) {
			weight := stake.CalculateWeight()
			share := weight.Quo(goodReputationWeight).MulInt(pool).TruncateInt()

			if share.IsPositive() {
				stake.PendingRewards = stake.PendingRewards.Add(share)
				k.SetUserStake(ctx, stake)
				distributed = distributed.Add(share)
			}
		}
		return false
	})

	return distributed
}

// =============================================================================
// REWARD CLAIMING
// =============================================================================

// ClaimRewards allows a user to claim their pending rewards
func (k Keeper) ClaimRewards(ctx sdk.Context, staker sdk.AccAddress) (math.Int, error) {
	stake, exists := k.GetUserStake(ctx, staker)
	if !exists {
		return math.ZeroInt(), types.ErrStakeNotFound
	}

	if stake.PendingRewards.IsZero() {
		return math.ZeroInt(), types.ErrNoRewardsToClaim
	}

	rewards := stake.PendingRewards

	// Transfer rewards from rewards pool to user
	coins := sdk.NewCoins(sdk.NewCoin("uhodl", rewards))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.RewardsPoolName, staker, coins); err != nil {
		return math.ZeroInt(), err
	}

	// Update stake
	stake.PendingRewards = math.ZeroInt()
	stake.TotalRewardsClaimed = stake.TotalRewardsClaimed.Add(rewards)
	stake.LastRewardClaim = ctx.BlockTime()
	if err := k.SetUserStake(ctx, stake); err != nil {
		return math.ZeroInt(), err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimRewards,
			sdk.NewAttribute(types.AttributeKeyStaker, staker.String()),
			sdk.NewAttribute(types.AttributeKeyRewards, rewards.String()),
		),
	)

	k.Logger(ctx).Info("rewards claimed",
		"staker", staker.String(),
		"amount", rewards.String(),
		"total_claimed", stake.TotalRewardsClaimed.String(),
	)

	return rewards, nil
}

// GetPendingRewards returns pending rewards for a user
func (k Keeper) GetPendingRewards(ctx sdk.Context, staker sdk.AccAddress) math.Int {
	stake, exists := k.GetUserStake(ctx, staker)
	if !exists {
		return math.ZeroInt()
	}
	return stake.PendingRewards
}

// GetTotalRewardsClaimed returns total rewards claimed by a user
func (k Keeper) GetTotalRewardsClaimed(ctx sdk.Context, staker sdk.AccAddress) math.Int {
	stake, exists := k.GetUserStake(ctx, staker)
	if !exists {
		return math.ZeroInt()
	}
	return stake.TotalRewardsClaimed
}

// =============================================================================
// SLASHING
// =============================================================================

// Slash reduces a user's stake for misbehavior
func (k Keeper) Slash(ctx sdk.Context, staker sdk.AccAddress, reason string, slashFraction math.LegacyDec) error {
	stake, exists := k.GetUserStake(ctx, staker)
	if !exists {
		return types.ErrStakeNotFound
	}

	// Get maximum slash for tier
	maxSlashBps := stake.Tier.GetSlashRisk()
	maxSlashFraction := math.LegacyNewDec(int64(maxSlashBps)).QuoInt64(10000)

	// Cap slash at tier maximum
	if slashFraction.GT(maxSlashFraction) {
		slashFraction = maxSlashFraction
	}

	// Calculate slash amount
	slashAmount := slashFraction.MulInt(stake.StakedAmount).TruncateInt()
	if slashAmount.IsZero() {
		return nil
	}

	oldTier := stake.Tier
	oldAmount := stake.StakedAmount
	newAmount := oldAmount.Sub(slashAmount)

	// Burn slashed tokens
	coins := sdk.NewCoins(sdk.NewCoin("uhodl", slashAmount))
	if err := k.bankKeeper.BurnCoins(ctx, types.StakingPoolName, coins); err != nil {
		return err
	}

	// Update stake
	stake.StakedAmount = newAmount

	// Check for tier demotion using governance-controlled thresholds
	var newTier types.StakeTier
	if newAmount.IsZero() {
		newTier = types.StakeTier(-1)
	} else if k.IsTestnet(ctx) {
		newTier = types.GetTierFromStake(newAmount, true)
	} else {
		params := k.GetParams(ctx)
		newTier = types.GetTierFromStakeWithParams(newAmount, params)
	}
	stake.Tier = newTier

	if err := k.SetUserStake(ctx, stake); err != nil {
		return err
	}

	// Update tier counts if tier changed
	if oldTier != newTier {
		k.decrementTierCount(ctx, oldTier)
		if newTier >= 0 {
			k.incrementTierCount(ctx, newTier)
		}
	}

	// Penalize reputation
	k.PenalizeSlash(ctx, staker, reason)

	// Emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeKeyStaker, staker.String()),
			sdk.NewAttribute(types.AttributeKeySlashAmount, slashAmount.String()),
			sdk.NewAttribute(types.AttributeKeySlashReason, reason),
		),
	)

	if oldTier != newTier {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTierChange,
				sdk.NewAttribute(types.AttributeKeyStaker, staker.String()),
				sdk.NewAttribute(types.AttributeKeyOldTier, oldTier.String()),
				sdk.NewAttribute(types.AttributeKeyNewTier, newTier.String()),
				sdk.NewAttribute(types.AttributeKeyReason, "slashed: "+reason),
			),
		)

		if k.hooks != nil {
			k.hooks.AfterTierChange(ctx, staker, oldTier, newTier)
		}
	}

	if k.hooks != nil {
		k.hooks.AfterSlash(ctx, staker, slashAmount, reason)
	}

	k.Logger(ctx).Info("staker slashed",
		"staker", staker.String(),
		"amount", slashAmount.String(),
		"reason", reason,
		"old_tier", oldTier.String(),
		"new_tier", newTier.String(),
	)

	return nil
}

// GetDowntimeSlashFraction returns the governance-controllable downtime slash fraction
func (k Keeper) GetDowntimeSlashFraction(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	return math.LegacyNewDec(int64(params.DowntimeSlashBasisPoints)).QuoInt64(10000)
}

// GetDoubleSignSlashFraction returns the governance-controllable double sign slash fraction
func (k Keeper) GetDoubleSignSlashFraction(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	return math.LegacyNewDec(int64(params.DoubleSignSlashBasisPoints)).QuoInt64(10000)
}

// GetBadVerificationSlashFraction returns the governance-controllable bad verification slash fraction
func (k Keeper) GetBadVerificationSlashFraction(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	return math.LegacyNewDec(int64(params.BadVerificationSlashBasisPoints)).QuoInt64(10000)
}

// GetGoodReputationThreshold returns the governance-controllable good reputation threshold
func (k Keeper) GetGoodReputationThreshold(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).GoodReputationThreshold
}

// SlashForDowntime slashes a validator for downtime (governance-controllable rate)
func (k Keeper) SlashForDowntime(ctx sdk.Context, staker sdk.AccAddress) error {
	return k.Slash(ctx, staker, "validator_downtime", k.GetDowntimeSlashFraction(ctx))
}

// SlashForDoubleSign slashes a validator for double signing (governance-controllable rate)
func (k Keeper) SlashForDoubleSign(ctx sdk.Context, staker sdk.AccAddress) error {
	return k.Slash(ctx, staker, "double_sign", k.GetDoubleSignSlashFraction(ctx))
}

// SlashForBadVerification slashes a user for bad business verification (governance-controllable rate)
func (k Keeper) SlashForBadVerification(ctx sdk.Context, staker sdk.AccAddress) error {
	return k.Slash(ctx, staker, "bad_verification", k.GetBadVerificationSlashFraction(ctx))
}

// SlashForLoanDefault slashes a user for loan default
func (k Keeper) SlashForLoanDefault(ctx sdk.Context, staker sdk.AccAddress, defaultAmount math.Int) error {
	stake, exists := k.GetUserStake(ctx, staker)
	if !exists {
		return types.ErrStakeNotFound
	}

	// Slash proportional to default amount vs stake, up to tier max
	slashFraction := math.LegacyNewDecFromInt(defaultAmount).Quo(math.LegacyNewDecFromInt(stake.StakedAmount))
	return k.Slash(ctx, staker, "loan_default", slashFraction)
}
