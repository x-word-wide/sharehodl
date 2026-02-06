package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

// IsWalletBanned checks if a wallet address is banned
func (k Keeper) IsWalletBanned(ctx sdk.Context, address string) bool {
	if k.banKeeper == nil {
		return false
	}
	return k.banKeeper.IsAddressBanned(ctx, address)
}

// IsOwnerBanned checks if the plan owner is banned
// CRITICAL: Banned owners cannot trigger or use the dead man switch
func (k Keeper) IsOwnerBanned(ctx sdk.Context, plan types.InheritancePlan) bool {
	return k.IsWalletBanned(ctx, plan.Owner)
}

// IsBeneficiaryValid checks if a beneficiary is valid (not banned)
func (k Keeper) IsBeneficiaryValid(ctx sdk.Context, beneficiaryAddress string) error {
	if k.IsWalletBanned(ctx, beneficiaryAddress) {
		k.Logger(ctx).Info("beneficiary is banned",
			"beneficiary", beneficiaryAddress,
		)
		return types.ErrBeneficiaryBanned
	}
	return nil
}

// ValidateAllBeneficiaries checks if all beneficiaries in a plan are valid
// Returns list of invalid beneficiaries
func (k Keeper) ValidateAllBeneficiaries(ctx sdk.Context, plan types.InheritancePlan) []string {
	var invalidBeneficiaries []string

	for _, beneficiary := range plan.Beneficiaries {
		if k.IsWalletBanned(ctx, beneficiary.Address) {
			invalidBeneficiaries = append(invalidBeneficiaries, beneficiary.Address)
		}
	}

	return invalidBeneficiaries
}

// CheckAndSkipBannedBeneficiaries checks all beneficiaries and skips banned ones
func (k Keeper) CheckAndSkipBannedBeneficiaries(ctx sdk.Context, planID uint64) error {
	plan, found := k.GetPlan(ctx, planID)
	if !found {
		return types.ErrPlanNotFound
	}

	invalidBeneficiaries := k.ValidateAllBeneficiaries(ctx, plan)

	for _, beneficiary := range invalidBeneficiaries {
		err := k.SkipBeneficiaryAndCascade(ctx, planID, beneficiary, "beneficiary_banned")
		if err != nil {
			k.Logger(ctx).Error("failed to skip banned beneficiary",
				"plan_id", planID,
				"beneficiary", beneficiary,
				"error", err,
			)
		}
	}

	return nil
}
