package keeper

import (
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

// TriggerSwitch triggers the dead man switch for a plan
// CRITICAL: This function checks if owner is banned - banned users cannot trigger switches
func (k Keeper) TriggerSwitch(ctx sdk.Context, planID uint64) error {
	plan, found := k.GetPlan(ctx, planID)
	if !found {
		return types.ErrPlanNotFound
	}

	// Check plan status
	if plan.Status != types.PlanStatusActive {
		return types.ErrPlanNotActive
	}

	// CRITICAL BAN CHECK: Dead man switch MUST NOT work if owner is banned
	if k.banKeeper != nil && k.banKeeper.IsAddressBanned(ctx, plan.Owner) {
		k.Logger(ctx).Error("cannot trigger switch for banned owner",
			"plan_id", planID,
			"owner", plan.Owner,
		)
		return types.ErrOwnerBanned
	}

	// Check if owner is actually inactive
	owner, _ := sdk.AccAddressFromBech32(plan.Owner)
	if !k.IsInactive(ctx, owner, plan.InactivityPeriod) {
		return types.ErrOwnerStillActive
	}

	// Check if already triggered
	if _, found := k.GetActiveTrigger(ctx, planID); found {
		return types.ErrPlanAlreadyTriggered
	}

	// Create the trigger
	trigger := types.SwitchTrigger{
		PlanID:              planID,
		TriggeredAt:         ctx.BlockTime(),
		GracePeriodEnd:      ctx.BlockTime().Add(plan.GracePeriod),
		Status:              types.TriggerStatusActive,
		LastInactivityCheck: ctx.BlockTime(),
	}

	// Store the trigger
	k.SetActiveTrigger(ctx, trigger)

	// Update plan status
	plan.Status = types.PlanStatusTriggered
	plan.UpdatedAt = ctx.BlockTime()
	k.SetPlan(ctx, plan)

	k.Logger(ctx).Info("dead man switch triggered",
		"plan_id", planID,
		"owner", plan.Owner,
		"grace_period_end", trigger.GracePeriodEnd,
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwitchTriggered,
			sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(planID)),
			sdk.NewAttribute(types.AttributeKeyOwner, plan.Owner),
			sdk.NewAttribute(types.AttributeKeyTriggeredAt, trigger.TriggeredAt.String()),
			sdk.NewAttribute(types.AttributeKeyGracePeriodEnd, trigger.GracePeriodEnd.String()),
		),
	)

	return nil
}

// CancelTrigger allows the owner to manually cancel a triggered switch
func (k Keeper) CancelTrigger(ctx sdk.Context, owner sdk.AccAddress, planID uint64) error {
	plan, found := k.GetPlan(ctx, planID)
	if !found {
		return types.ErrPlanNotFound
	}

	// Check ownership
	if plan.Owner != owner.String() {
		return types.ErrUnauthorized
	}

	// Get the trigger
	trigger, found := k.GetActiveTrigger(ctx, planID)
	if !found {
		return types.ErrTriggerNotFound
	}

	// Check trigger status
	if trigger.Status != types.TriggerStatusActive {
		return fmt.Errorf("trigger is not active")
	}

	// Cancel the trigger
	trigger.Status = types.TriggerStatusCancelled
	k.SetActiveTrigger(ctx, trigger)

	// Update plan status back to active
	plan.Status = types.PlanStatusActive
	plan.UpdatedAt = ctx.BlockTime()
	k.SetPlan(ctx, plan)

	// Record activity (since owner is clearly active)
	k.RecordActivity(ctx, owner, "cancel_trigger")

	k.Logger(ctx).Info("trigger cancelled by owner",
		"plan_id", planID,
		"owner", owner.String(),
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSwitchCancelled,
			sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(planID)),
			sdk.NewAttribute(types.AttributeKeyOwner, plan.Owner),
			sdk.NewAttribute(types.AttributeKeyReason, "manual_cancel"),
		),
	)

	return nil
}

// GetActiveTrigger returns an active trigger for a plan
func (k Keeper) GetActiveTrigger(ctx sdk.Context, planID uint64) (types.SwitchTrigger, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ActiveTriggerKey(planID))
	if bz == nil {
		return types.SwitchTrigger{}, false
	}

	var trigger types.SwitchTrigger
	k.cdc.MustUnmarshal(bz, &trigger)
	return trigger, true
}

// SetActiveTrigger stores an active trigger
func (k Keeper) SetActiveTrigger(ctx sdk.Context, trigger types.SwitchTrigger) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&trigger)
	store.Set(types.ActiveTriggerKey(trigger.PlanID), bz)
}

// DeleteActiveTrigger deletes an active trigger
func (k Keeper) DeleteActiveTrigger(ctx sdk.Context, planID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.ActiveTriggerKey(planID))
}

// GetAllActiveTriggers returns all active triggers
func (k Keeper) GetAllActiveTriggers(ctx sdk.Context) []types.SwitchTrigger {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ActiveTriggerPrefix)
	defer iterator.Close()

	var triggers []types.SwitchTrigger
	for ; iterator.Valid(); iterator.Next() {
		var trigger types.SwitchTrigger
		k.cdc.MustUnmarshal(iterator.Value(), &trigger)
		triggers = append(triggers, trigger)
	}

	return triggers
}

// ValidateTriggerConditions checks if a trigger should be activated
func (k Keeper) ValidateTriggerConditions(ctx sdk.Context, planID uint64) error {
	plan, found := k.GetPlan(ctx, planID)
	if !found {
		return types.ErrPlanNotFound
	}

	// Check if owner is banned
	if k.banKeeper != nil && k.banKeeper.IsAddressBanned(ctx, plan.Owner) {
		return types.ErrOwnerBanned
	}

	// Check if owner is inactive
	owner, _ := sdk.AccAddressFromBech32(plan.Owner)
	if !k.IsInactive(ctx, owner, plan.InactivityPeriod) {
		return types.ErrOwnerStillActive
	}

	return nil
}

// ProcessGracePeriodExpiry processes a trigger whose grace period has expired
func (k Keeper) ProcessGracePeriodExpiry(ctx sdk.Context, trigger types.SwitchTrigger) error {
	plan, found := k.GetPlan(ctx, trigger.PlanID)
	if !found {
		return types.ErrPlanNotFound
	}

	// Update trigger status
	trigger.Status = types.TriggerStatusExpired
	k.SetActiveTrigger(ctx, trigger)

	// Update plan status
	plan.Status = types.PlanStatusExecuting
	plan.UpdatedAt = ctx.BlockTime()
	k.SetPlan(ctx, plan)

	// Initialize claims for all beneficiaries (by priority)
	err := k.InitializeBeneficiaryClaims(ctx, plan)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info("grace period expired, claims initialized",
		"plan_id", plan.PlanID,
		"beneficiaries", len(plan.Beneficiaries),
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGracePeriodExpired,
			sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(plan.PlanID)),
			sdk.NewAttribute(types.AttributeKeyOwner, plan.Owner),
		),
	)

	return nil
}

// InitializeBeneficiaryClaims creates initial claim records for all beneficiaries
func (k Keeper) InitializeBeneficiaryClaims(ctx sdk.Context, plan types.InheritancePlan) error {
	// Sort beneficiaries by priority (lower number = higher priority)
	// For simplicity, we'll process them in order and create cascading claim windows

	currentTime := ctx.BlockTime()

	for i, beneficiary := range plan.Beneficiaries {
		// Calculate claim window
		// Priority 1 gets first window, priority 2 gets second window, etc.
		windowStart := currentTime.Add(time.Duration(i) * plan.ClaimWindow)
		windowEnd := windowStart.Add(plan.ClaimWindow)

		claim := types.BeneficiaryClaim{
			PlanID:           plan.PlanID,
			Beneficiary:      beneficiary.Address,
			Priority:         beneficiary.Priority,
			Percentage:       beneficiary.Percentage,
			ClaimWindowStart: windowStart,
			ClaimWindowEnd:   windowEnd,
			Status:           types.ClaimStatusPending,
		}

		k.SetClaim(ctx, claim)

		k.Logger(ctx).Debug("beneficiary claim initialized",
			"plan_id", plan.PlanID,
			"beneficiary", beneficiary.Address,
			"priority", beneficiary.Priority,
			"window_start", windowStart,
			"window_end", windowEnd,
		)
	}

	return nil
}

// SetClaim stores a beneficiary claim
func (k Keeper) SetClaim(ctx sdk.Context, claim types.BeneficiaryClaim) {
	store := ctx.KVStore(k.storeKey)
	beneficiary, _ := sdk.AccAddressFromBech32(claim.Beneficiary)
	bz := k.cdc.MustMarshal(&claim)
	store.Set(types.ClaimKey(claim.PlanID, beneficiary), bz)
}

// GetClaim returns a beneficiary claim
func (k Keeper) GetClaim(ctx sdk.Context, planID uint64, beneficiary sdk.AccAddress) (types.BeneficiaryClaim, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ClaimKey(planID, beneficiary))
	if bz == nil {
		return types.BeneficiaryClaim{}, false
	}

	var claim types.BeneficiaryClaim
	k.cdc.MustUnmarshal(bz, &claim)
	return claim, true
}

// GetAllClaimsForPlan returns all claims for a plan
func (k Keeper) GetAllClaimsForPlan(ctx sdk.Context, planID uint64) []types.BeneficiaryClaim {
	store := ctx.KVStore(k.storeKey)
	prefix := types.ClaimPrefixForPlan(planID)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var claims []types.BeneficiaryClaim
	for ; iterator.Valid(); iterator.Next() {
		var claim types.BeneficiaryClaim
		k.cdc.MustUnmarshal(iterator.Value(), &claim)
		claims = append(claims, claim)
	}

	return claims
}
