package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

// EndBlocker processes inheritance-related tasks at the end of each block
func (k Keeper) EndBlocker(ctx sdk.Context) error {
	// 1. Process inactivity checks for all active plans
	k.ProcessInactivityChecks(ctx)

	// 2. Process grace period expiries
	k.ProcessGracePeriodExpiries(ctx)

	// 3. Process claim window expiries
	k.ProcessClaimWindowExpiries(ctx)

	// 4. Process ultra-long inactivity (50+ years)
	k.ProcessUltraLongInactivity(ctx)

	return nil
}

// ProcessInactivityChecks checks for owners who have been inactive and triggers switches
func (k Keeper) ProcessInactivityChecks(ctx sdk.Context) {
	k.IteratePlans(ctx, func(plan types.InheritancePlan) bool {
		// Skip if not active
		if plan.Status != types.PlanStatusActive {
			return false
		}

		// Check if owner is banned - skip if banned
		if k.IsOwnerBanned(ctx, plan) {
			return false
		}

		// Check if owner is inactive
		owner, _ := sdk.AccAddressFromBech32(plan.Owner)
		if k.IsInactive(ctx, owner, plan.InactivityPeriod) {
			// Check if already triggered
			if _, found := k.GetActiveTrigger(ctx, plan.PlanID); !found {
				// Trigger the switch
				err := k.TriggerSwitch(ctx, plan.PlanID)
				if err != nil {
					k.Logger(ctx).Error("failed to trigger switch",
						"plan_id", plan.PlanID,
						"owner", plan.Owner,
						"error", err,
					)
				}
			}
		}

		return false
	})
}

// ProcessGracePeriodExpiries processes triggers whose grace period has expired
func (k Keeper) ProcessGracePeriodExpiries(ctx sdk.Context) {
	triggers := k.GetAllActiveTriggers(ctx)

	for _, trigger := range triggers {
		// Skip if not active
		if trigger.Status != types.TriggerStatusActive {
			continue
		}

		// Check if grace period has expired
		if ctx.BlockTime().After(trigger.GracePeriodEnd) {
			err := k.ProcessGracePeriodExpiry(ctx, trigger)
			if err != nil {
				k.Logger(ctx).Error("failed to process grace period expiry",
					"plan_id", trigger.PlanID,
					"error", err,
				)
			}
		}
	}
}

// ProcessClaimWindowExpiries processes expired claim windows
func (k Keeper) ProcessClaimWindowExpiries(ctx sdk.Context) {
	// Iterate through all plans in executing status
	k.IteratePlans(ctx, func(plan types.InheritancePlan) bool {
		if plan.Status != types.PlanStatusExecuting {
			return false
		}

		// Get all claims for this plan
		claims := k.GetAllClaimsForPlan(ctx, plan.PlanID)

		for _, claim := range claims {
			// Skip already processed claims
			if claim.Status == types.ClaimStatusClaimed || claim.Status == types.ClaimStatusSkipped || claim.Status == types.ClaimStatusExpired {
				continue
			}

			// Check if claim window has expired
			if ctx.BlockTime().After(claim.ClaimWindowEnd) {
				// Check if beneficiary is banned before expiring
				if k.IsBeneficiaryValid(ctx, claim.Beneficiary) != nil {
					// Beneficiary is banned - skip them
					err := k.SkipBeneficiaryAndCascade(ctx, plan.PlanID, claim.Beneficiary, "beneficiary_banned")
					if err != nil {
						k.Logger(ctx).Error("failed to skip banned beneficiary",
							"plan_id", plan.PlanID,
							"beneficiary", claim.Beneficiary,
							"error", err,
						)
					}
				} else {
					// Process expiry
					err := k.ProcessClaimWindowExpiry(ctx, claim)
					if err != nil {
						k.Logger(ctx).Error("failed to process claim window expiry",
							"plan_id", plan.PlanID,
							"beneficiary", claim.Beneficiary,
							"error", err,
						)
					}
				}
			}

			// Open claim window if it's time
			if claim.Status == types.ClaimStatusPending && ctx.BlockTime().After(claim.ClaimWindowStart) {
				claim.Status = types.ClaimStatusOpen
				k.SetClaim(ctx, claim)

				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeClaimWindowOpened,
						sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(plan.PlanID)),
						sdk.NewAttribute(types.AttributeKeyBeneficiary, claim.Beneficiary),
					),
				)
			}
		}

		return false
	})
}

// ProcessUltraLongInactivity handles ultra-long inactivity (50+ years)
// After this period, assets go to charity/community pool
func (k Keeper) ProcessUltraLongInactivity(ctx sdk.Context) {
	params := k.GetParams(ctx)

	k.IteratePlans(ctx, func(plan types.InheritancePlan) bool {
		// Only process completed plans or plans that have been inactive for ultra-long time
		owner, _ := sdk.AccAddressFromBech32(plan.Owner)
		activity, found := k.GetLastActivity(ctx, owner)

		var timeSinceActivity time.Duration
		if found {
			timeSinceActivity = ctx.BlockTime().Sub(activity.LastActivityTime)
		} else {
			// If no activity found, use plan creation time
			timeSinceActivity = ctx.BlockTime().Sub(plan.CreatedAt)
		}

		// Check if ultra-long inactivity period has passed
		if timeSinceActivity >= params.UltraLongInactivityPeriod {
			// Transfer remaining assets to charity
			err := k.TransferToCharity(ctx, plan)
			if err != nil {
				k.Logger(ctx).Error("failed to transfer ultra-long inactive assets to charity",
					"plan_id", plan.PlanID,
					"owner", plan.Owner,
					"error", err,
				)
				return false
			}

			// Mark plan as completed
			plan.Status = types.PlanStatusCompleted
			plan.UpdatedAt = ctx.BlockTime()
			k.SetPlan(ctx, plan)

			k.Logger(ctx).Info("ultra-long inactive assets transferred to charity",
				"plan_id", plan.PlanID,
				"owner", plan.Owner,
				"inactivity_duration", timeSinceActivity,
			)
		}

		return false
	})
}
