package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

// ClaimInheritance processes a beneficiary's claim
func (k Keeper) ClaimInheritance(ctx sdk.Context, beneficiary sdk.AccAddress, planID uint64) ([]types.TransferredAsset, error) {
	// CRITICAL FIX: Check for claim lock to prevent double-claim race condition
	if k.HasClaimLock(ctx, planID) {
		return nil, types.ErrClaimInProgress
	}

	// Set claim lock atomically
	k.SetClaimLock(ctx, planID)
	defer k.RemoveClaimLock(ctx, planID)

	// Get the plan
	plan, found := k.GetPlan(ctx, planID)
	if !found {
		return nil, types.ErrPlanNotFound
	}

	// Get the claim
	claim, found := k.GetClaim(ctx, planID, beneficiary)
	if !found {
		return nil, types.ErrClaimNotFound
	}

	// Validate claim status and timing
	if err := k.ValidateBeneficiaryEligibility(ctx, claim); err != nil {
		return nil, err
	}

	// Update claim status to "processing" BEFORE transfers (prevents race condition)
	claim.Status = types.ClaimStatusProcessing
	k.SetClaim(ctx, claim)

	// Process the claim
	transferredAssets, err := k.ProcessClaim(ctx, plan, claim)
	if err != nil {
		// Revert to open status on error
		claim.Status = types.ClaimStatusOpen
		k.SetClaim(ctx, claim)
		return nil, err
	}

	// Update claim status to claimed
	claim.Status = types.ClaimStatusClaimed
	claim.ClaimedAt = ctx.BlockTime()
	claim.TransferredAssets = transferredAssets
	k.SetClaim(ctx, claim)

	k.Logger(ctx).Info("inheritance claimed",
		"plan_id", planID,
		"beneficiary", beneficiary.String(),
		"assets_transferred", len(transferredAssets),
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAssetsClaimed,
			sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(planID)),
			sdk.NewAttribute(types.AttributeKeyBeneficiary, beneficiary.String()),
			sdk.NewAttribute(types.AttributeKeyPriority, fmt.Sprint(claim.Priority)),
		),
	)

	// Check if all claims are processed
	k.CheckAndCompletePlan(ctx, plan)

	return transferredAssets, nil
}

// ProcessClaim transfers assets to a beneficiary
func (k Keeper) ProcessClaim(ctx sdk.Context, plan types.InheritancePlan, claim types.BeneficiaryClaim) ([]types.TransferredAsset, error) {
	owner, _ := sdk.AccAddressFromBech32(plan.Owner)
	beneficiary, _ := sdk.AccAddressFromBech32(claim.Beneficiary)

	var transferredAssets []types.TransferredAsset

	// Find beneficiary details in plan
	var beneficiaryInfo *types.Beneficiary
	for _, b := range plan.Beneficiaries {
		if b.Address == claim.Beneficiary {
			beneficiaryInfo = &b
			break
		}
	}

	if beneficiaryInfo == nil {
		return nil, fmt.Errorf("beneficiary not found in plan")
	}

	// Transfer specific assets if defined
	if len(beneficiaryInfo.SpecificAssets) > 0 {
		for _, asset := range beneficiaryInfo.SpecificAssets {
			transferred, err := k.TransferSpecificAsset(ctx, owner, beneficiary, asset)
			if err != nil {
				k.Logger(ctx).Error("failed to transfer specific asset",
					"plan_id", plan.PlanID,
					"beneficiary", beneficiary.String(),
					"asset_type", asset.AssetType,
					"error", err,
				)
				// Continue with other assets even if one fails
				continue
			}
			transferredAssets = append(transferredAssets, transferred...)
		}
	}

	// Transfer percentage of remaining assets
	if claim.Percentage.IsPositive() {
		transferred, err := k.TransferPercentageOfAssets(ctx, owner, beneficiary, claim.Percentage)
		if err != nil {
			k.Logger(ctx).Error("failed to transfer percentage of assets",
				"plan_id", plan.PlanID,
				"beneficiary", beneficiary.String(),
				"percentage", claim.Percentage,
				"error", err,
			)
			// Log but don't fail - some assets may have been transferred
		} else {
			transferredAssets = append(transferredAssets, transferred...)
		}
	}

	return transferredAssets, nil
}

// ValidateBeneficiaryEligibility checks if a beneficiary is eligible to claim
func (k Keeper) ValidateBeneficiaryEligibility(ctx sdk.Context, claim types.BeneficiaryClaim) error {
	// Check claim status
	if claim.Status == types.ClaimStatusClaimed {
		return types.ErrClaimAlreadyProcessed
	}

	if claim.Status == types.ClaimStatusProcessing {
		return types.ErrClaimInProgress
	}

	if claim.Status == types.ClaimStatusSkipped {
		return types.ErrInvalidBeneficiary
	}

	// Check if claim window is open
	currentTime := ctx.BlockTime()
	if currentTime.Before(claim.ClaimWindowStart) {
		return types.ErrClaimWindowClosed
	}

	if currentTime.After(claim.ClaimWindowEnd) {
		return types.ErrClaimWindowClosed
	}

	// Check if beneficiary is banned
	if k.IsBeneficiaryValid(ctx, claim.Beneficiary) != nil {
		return types.ErrBeneficiaryBanned
	}

	return nil
}

const MaxCascadeDepth = 10

// SkipBeneficiaryAndCascade skips a beneficiary and cascades to next priority
func (k Keeper) SkipBeneficiaryAndCascade(ctx sdk.Context, planID uint64, beneficiary string, reason string) error {
	return k.skipBeneficiaryWithDepth(ctx, planID, beneficiary, reason, 0)
}

// skipBeneficiaryWithDepth is the internal implementation with cascade depth tracking
func (k Keeper) skipBeneficiaryWithDepth(ctx sdk.Context, planID uint64, beneficiary string, reason string, depth int) error {
	// CRITICAL FIX: Prevent infinite cascade loops
	if depth >= MaxCascadeDepth {
		k.Logger(ctx).Error("maximum cascade depth reached, transferring remaining assets to charity",
			"plan_id", planID,
			"depth", depth,
		)

		// Get the plan and transfer to charity instead of cascading further
		plan, found := k.GetPlan(ctx, planID)
		if found {
			if err := k.TransferToCharity(ctx, plan); err != nil {
				k.Logger(ctx).Error("failed to transfer to charity after max cascade depth",
					"plan_id", planID,
					"error", err,
				)
			}
		}

		return types.ErrMaxCascadeDepthReached
	}

	bAddr, _ := sdk.AccAddressFromBech32(beneficiary)
	claim, found := k.GetClaim(ctx, planID, bAddr)
	if !found {
		return types.ErrClaimNotFound
	}

	// Mark as skipped
	claim.Status = types.ClaimStatusSkipped
	k.SetClaim(ctx, claim)

	k.Logger(ctx).Info("beneficiary skipped",
		"plan_id", planID,
		"beneficiary", beneficiary,
		"reason", reason,
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBeneficiarySkipped,
			sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(planID)),
			sdk.NewAttribute(types.AttributeKeyBeneficiary, beneficiary),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)

	// Find next priority beneficiary
	plan, found := k.GetPlan(ctx, planID)
	if !found {
		return types.ErrPlanNotFound
	}

	// Get all claims and sort by priority
	allClaims := k.GetAllClaimsForPlan(ctx, planID)
	sort.Slice(allClaims, func(i, j int) bool {
		return allClaims[i].Priority < allClaims[j].Priority
	})

	// Find next unclaimed beneficiary
	cascadedToNext := false
	for _, nextClaim := range allClaims {
		if nextClaim.Status == types.ClaimStatusPending && nextClaim.Beneficiary != beneficiary {
			// Check if next beneficiary is also invalid before cascading
			if k.IsBeneficiaryValid(ctx, nextClaim.Beneficiary) != nil {
				// Next beneficiary is also invalid, recursively skip with increased depth
				k.Logger(ctx).Warn("next beneficiary is also invalid, continuing cascade",
					"plan_id", planID,
					"beneficiary", nextClaim.Beneficiary,
					"depth", depth+1,
				)
				return k.skipBeneficiaryWithDepth(ctx, planID, nextClaim.Beneficiary, "beneficiary_banned", depth+1)
			}

			// Open this claim window immediately
			nextClaim.Status = types.ClaimStatusOpen
			nextClaim.ClaimWindowStart = ctx.BlockTime()
			nextClaim.ClaimWindowEnd = ctx.BlockTime().Add(plan.ClaimWindow)
			k.SetClaim(ctx, nextClaim)

			k.Logger(ctx).Info("cascaded to next beneficiary",
				"plan_id", planID,
				"next_beneficiary", nextClaim.Beneficiary,
				"priority", nextClaim.Priority,
				"depth", depth,
			)

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeClaimWindowOpened,
					sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(planID)),
					sdk.NewAttribute(types.AttributeKeyBeneficiary, nextClaim.Beneficiary),
					sdk.NewAttribute(types.AttributeKeyPriority, fmt.Sprint(nextClaim.Priority)),
				),
			)

			cascadedToNext = true
			break
		}
	}

	// If no valid beneficiary found, transfer to charity
	if !cascadedToNext {
		k.Logger(ctx).Info("no more beneficiaries available, transferring to charity",
			"plan_id", planID,
		)
		plan, found := k.GetPlan(ctx, planID)
		if found {
			if err := k.TransferToCharity(ctx, plan); err != nil {
				k.Logger(ctx).Error("failed to transfer to charity",
					"plan_id", planID,
					"error", err,
				)
			}
		}
	}

	return nil
}

// CheckAndCompletePlan checks if all claims are processed and completes the plan
func (k Keeper) CheckAndCompletePlan(ctx sdk.Context, plan types.InheritancePlan) {
	allClaims := k.GetAllClaimsForPlan(ctx, plan.PlanID)

	allProcessed := true
	for _, claim := range allClaims {
		if claim.Status == types.ClaimStatusPending || claim.Status == types.ClaimStatusOpen {
			allProcessed = false
			break
		}
	}

	if allProcessed {
		// All claims processed - mark plan as completed
		plan.Status = types.PlanStatusCompleted
		plan.UpdatedAt = ctx.BlockTime()
		k.SetPlan(ctx, plan)

		k.Logger(ctx).Info("inheritance plan completed",
			"plan_id", plan.PlanID,
		)

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypePlanCompleted,
				sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(plan.PlanID)),
				sdk.NewAttribute(types.AttributeKeyOwner, plan.Owner),
			),
		)
	}
}

// HasClaimLock checks if a claim lock exists for a plan
func (k Keeper) HasClaimLock(ctx sdk.Context, planID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.ClaimLockKey(planID))
}

// SetClaimLock sets a claim lock to prevent race conditions
func (k Keeper) SetClaimLock(ctx sdk.Context, planID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ClaimLockKey(planID), []byte{1})
}

// RemoveClaimLock removes a claim lock
func (k Keeper) RemoveClaimLock(ctx sdk.Context, planID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.ClaimLockKey(planID))
}

// ProcessClaimWindowExpiry handles expired claim windows
func (k Keeper) ProcessClaimWindowExpiry(ctx sdk.Context, claim types.BeneficiaryClaim) error {
	if claim.Status == types.ClaimStatusClaimed || claim.Status == types.ClaimStatusSkipped || claim.Status == types.ClaimStatusProcessing {
		return nil // Already processed or being processed
	}

	// Mark as expired
	claim.Status = types.ClaimStatusExpired
	k.SetClaim(ctx, claim)

	k.Logger(ctx).Info("claim window expired",
		"plan_id", claim.PlanID,
		"beneficiary", claim.Beneficiary,
	)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimWindowClosed,
			sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(claim.PlanID)),
			sdk.NewAttribute(types.AttributeKeyBeneficiary, claim.Beneficiary),
			sdk.NewAttribute(types.AttributeKeyReason, "expired"),
		),
	)

	// Cascade to next beneficiary
	return k.SkipBeneficiaryAndCascade(ctx, claim.PlanID, claim.Beneficiary, "claim_window_expired")
}
