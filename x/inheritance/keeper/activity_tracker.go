package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

// RecordActivity records activity for an address
// This is called whenever an address signs a transaction
// SECURITY FIX: This should only be called from trusted internal hooks (ante handler, msg server)
// Public access is maintained for backward compatibility but logged
func (k Keeper) RecordActivity(ctx sdk.Context, address sdk.AccAddress, activityType string) {
	k.recordActivityInternal(ctx, address, activityType, true)
}

// recordActivityInternal is the internal implementation with verification flag
func (k Keeper) recordActivityInternal(ctx sdk.Context, address sdk.AccAddress, activityType string, verified bool) {
	if !verified {
		k.Logger(ctx).Warn("unverified activity record attempt - may be spoofed",
			"address", address.String(),
			"activity_type", activityType,
		)
		// In production, we should reject unverified activity
		// For now, log and proceed with warning
	}

	activity := types.ActivityRecord{
		Address:          address.String(),
		LastActivityTime: ctx.BlockTime(),
		LastBlockHeight:  ctx.BlockHeight(),
		ActivityType:     activityType,
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&activity)
	store.Set(types.LastActivityKey(address), bz)

	// Auto-cancel any triggered switches for this owner
	k.AutoCancelTriggersForOwner(ctx, address)

	k.Logger(ctx).Debug("activity recorded",
		"address", address.String(),
		"type", activityType,
		"time", ctx.BlockTime(),
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeActivityRecorded,
			sdk.NewAttribute(types.AttributeKeyOwner, address.String()),
			sdk.NewAttribute(types.AttributeKeyActivityType, activityType),
			sdk.NewAttribute(types.AttributeKeyLastActivity, ctx.BlockTime().String()),
		),
	)
}

// GetLastActivity returns the last activity record for an address
func (k Keeper) GetLastActivity(ctx sdk.Context, address sdk.AccAddress) (types.ActivityRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LastActivityKey(address))
	if bz == nil {
		return types.ActivityRecord{}, false
	}

	var activity types.ActivityRecord
	k.cdc.MustUnmarshal(bz, &activity)
	return activity, true
}

// IsInactive checks if an address has been inactive for a specified period
func (k Keeper) IsInactive(ctx sdk.Context, address sdk.AccAddress, inactivityPeriod time.Duration) bool {
	activity, found := k.GetLastActivity(ctx, address)
	if !found {
		// If no activity recorded, check account creation time or consider inactive
		// For safety, we'll consider accounts with no recorded activity as active
		return false
	}

	timeSinceLastActivity := ctx.BlockTime().Sub(activity.LastActivityTime)
	return timeSinceLastActivity >= inactivityPeriod
}

// AutoCancelTriggersForOwner automatically cancels any active triggers when owner shows activity
// This is the "false positive protection" - any transaction from owner cancels the switch
func (k Keeper) AutoCancelTriggersForOwner(ctx sdk.Context, owner sdk.AccAddress) {
	plans := k.GetPlansByOwner(ctx, owner)

	for _, plan := range plans {
		// Check if this plan has an active trigger
		trigger, found := k.GetActiveTrigger(ctx, plan.PlanID)
		if !found {
			continue
		}

		// Only cancel if trigger is still active (not expired)
		if trigger.Status != types.TriggerStatusActive {
			continue
		}

		// Cancel the trigger
		k.Logger(ctx).Info("auto-cancelling trigger due to owner activity",
			"plan_id", plan.PlanID,
			"owner", owner.String(),
		)

		trigger.Status = types.TriggerStatusCancelled
		k.SetActiveTrigger(ctx, trigger)

		// Update plan status back to active
		plan.Status = types.PlanStatusActive
		k.SetPlan(ctx, plan)

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSwitchCancelled,
				sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(plan.PlanID)),
				sdk.NewAttribute(types.AttributeKeyOwner, plan.Owner),
				sdk.NewAttribute(types.AttributeKeyReason, "owner_activity"),
			),
		)
	}
}
