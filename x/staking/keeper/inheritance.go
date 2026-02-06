package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/staking/types"
)

// UnbondingInheritance tracks an unbonding request with a different recipient
type UnbondingInheritance struct {
	Owner       string    `json:"owner"`
	Recipient   string    `json:"recipient"`
	Amount      math.Int  `json:"amount"`
	InitiatedAt time.Time `json:"initiated_at"`
	CompletesAt time.Time `json:"completes_at"`
	Status      string    `json:"status"` // "active", "completed"
}

// UnstakeForInheritance initiates unstaking with a different recipient
// This is called during inheritance to transfer staked assets to beneficiaries
// The 14-day unbonding period still applies, but funds go to the recipient
func (k Keeper) UnstakeForInheritance(ctx sdk.Context, owner sdk.AccAddress, recipient string, amount math.Int) error {
	stake, exists := k.GetUserStake(ctx, owner)
	if !exists {
		return types.ErrStakeNotFound
	}

	if amount.GT(stake.StakedAmount) {
		return types.ErrInvalidAmount
	}

	// Validate recipient address
	recipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}

	// Check for active locks - inheritance overrides this check
	// NOTE: Inheritance is a special case where we allow unstaking even with active obligations
	// The beneficiary inherits both the stake AND the obligations

	// Check for existing unbonding request
	if _, exists := k.GetUnbondingRequest(ctx, owner); exists {
		return types.ErrUnbondingInProgress
	}

	newAmount := stake.StakedAmount.Sub(amount)
	oldTier := stake.Tier

	// Calculate new tier using governance-controlled thresholds
	var newTier types.StakeTier
	if newAmount.IsZero() {
		newTier = types.StakeTier(-1)
	} else if k.IsTestnet(ctx) {
		newTier = types.GetTierFromStake(newAmount, true)
	} else {
		params := k.GetParams(ctx)
		newTier = types.GetTierFromStakeWithParams(newAmount, params)
	}

	// Update stake record (tokens remain in staking pool during unbonding)
	if newAmount.IsZero() {
		k.DeleteUserStake(ctx, owner)
		k.decrementTierCount(ctx, oldTier)
	} else {
		stake.StakedAmount = newAmount
		stake.Tier = newTier
		if err := k.SetUserStake(ctx, stake); err != nil {
			return err
		}

		if oldTier != newTier {
			k.decrementTierCount(ctx, oldTier)
			if newTier >= 0 {
				k.incrementTierCount(ctx, newTier)
			}
		}
	}

	// Create inheritance unbonding request (14-day waiting period)
	// The recipient will receive the funds instead of the owner
	if err := k.CreateInheritanceUnbondingRequest(ctx, owner, recipientAddr, amount); err != nil {
		return err
	}

	// Emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(types.AttributeKeyStaker, owner.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute("recipient", recipient),
			sdk.NewAttribute("inheritance", "true"),
		),
	)

	if oldTier != newTier {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTierChange,
				sdk.NewAttribute(types.AttributeKeyStaker, owner.String()),
				sdk.NewAttribute(types.AttributeKeyOldTier, oldTier.String()),
				sdk.NewAttribute(types.AttributeKeyNewTier, newTier.String()),
			),
		)

		if k.hooks != nil {
			if err := k.hooks.AfterTierChange(ctx, owner, oldTier, newTier); err != nil {
				k.Logger(ctx).Error("tier change hook failed", "error", err)
			}
		}
	}

	if k.hooks != nil {
		if err := k.hooks.AfterUnstake(ctx, owner, amount, oldTier); err != nil {
			k.Logger(ctx).Error("unstake hook failed", "error", err)
		}
	}

	k.Logger(ctx).Info("unstaking initiated for inheritance (14-day unbonding period)",
		"owner", owner.String(),
		"recipient", recipient,
		"amount", amount.String(),
		"remaining", newAmount.String(),
		"old_tier", oldTier.String(),
		"new_tier", newTier.String(),
	)

	return nil
}

// CreateInheritanceUnbondingRequest creates an unbonding request with a different recipient
func (k Keeper) CreateInheritanceUnbondingRequest(ctx sdk.Context, owner sdk.AccAddress, recipient sdk.AccAddress, amount math.Int) error {
	// SECURITY FIX: Use configured unbonding period from params instead of hardcoded value
	params := k.GetParams(ctx)
	// Assuming 6-second block times, convert blocks to duration
	unbondingDuration := time.Duration(params.UnbondingPeriodBlocks) * 6 * time.Second

	unbonding := UnbondingInheritance{
		Owner:       owner.String(),
		Recipient:   recipient.String(),
		Amount:      amount,
		InitiatedAt: ctx.BlockTime(),
		CompletesAt: ctx.BlockTime().Add(unbondingDuration),
		Status:      "active",
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetInheritanceUnbondingKey(owner)
	bz, err := json.Marshal(unbonding)
	if err != nil {
		return fmt.Errorf("failed to marshal inheritance unbonding: %w", err)
	}
	store.Set(key, bz)

	return nil
}

// GetInheritanceUnbondingRequest returns an inheritance unbonding request
func (k Keeper) GetInheritanceUnbondingRequest(ctx sdk.Context, owner sdk.AccAddress) (UnbondingInheritance, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetInheritanceUnbondingKey(owner)
	bz := store.Get(key)
	if bz == nil {
		return UnbondingInheritance{}, false
	}

	var unbonding UnbondingInheritance
	if err := json.Unmarshal(bz, &unbonding); err != nil {
		return UnbondingInheritance{}, false
	}
	return unbonding, true
}

// ProcessInheritanceUnbondings processes all completed inheritance unbondings
// This should be called in EndBlock
func (k Keeper) ProcessInheritanceUnbondings(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.InheritanceUnbondingPrefix)
	defer iterator.Close()

	currentTime := ctx.BlockTime()

	for ; iterator.Valid(); iterator.Next() {
		var unbonding UnbondingInheritance
		if err := json.Unmarshal(iterator.Value(), &unbonding); err != nil {
			continue
		}

		// Check if unbonding period is complete
		if currentTime.After(unbonding.CompletesAt) || currentTime.Equal(unbonding.CompletesAt) {
			// Transfer funds to recipient
			recipientAddr, err := sdk.AccAddressFromBech32(unbonding.Recipient)
			if err != nil {
				k.Logger(ctx).Error("invalid recipient address in inheritance unbonding",
					"recipient", unbonding.Recipient,
					"error", err,
				)
				continue
			}

			coins := sdk.NewCoins(sdk.NewCoin("uhodl", unbonding.Amount))
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.StakingPoolName, recipientAddr, coins); err != nil {
				k.Logger(ctx).Error("failed to send inheritance unbonding to recipient",
					"recipient", unbonding.Recipient,
					"amount", unbonding.Amount.String(),
					"error", err,
				)
				continue
			}

			// Delete unbonding request
			store.Delete(iterator.Key())

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"inheritance_unbonding_completed",
					sdk.NewAttribute("owner", unbonding.Owner),
					sdk.NewAttribute("recipient", unbonding.Recipient),
					sdk.NewAttribute("amount", unbonding.Amount.String()),
				),
			)

			k.Logger(ctx).Info("inheritance unbonding completed",
				"owner", unbonding.Owner,
				"recipient", unbonding.Recipient,
				"amount", unbonding.Amount.String(),
			)
		}
	}
}
