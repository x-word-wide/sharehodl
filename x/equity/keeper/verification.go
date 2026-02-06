package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// =============================================================================
// VERIFICATION PARAMS
// =============================================================================

// GetVerificationParams returns the verification parameters
func (k Keeper) GetVerificationParams(ctx sdk.Context) types.VerificationParams {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.VerificationParamsKey)
	if bz == nil {
		return types.DefaultVerificationParams()
	}

	var params types.VerificationParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultVerificationParams()
	}
	return params
}

// SetVerificationParams sets the verification parameters
func (k Keeper) SetVerificationParams(ctx sdk.Context, params types.VerificationParams) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.VerificationParamsKey, bz)
	return nil
}

// =============================================================================
// BUSINESS VERIFICATION CRUD
// =============================================================================

// GetBusinessVerification returns a business verification by ID
func (k Keeper) GetBusinessVerification(ctx sdk.Context, businessID string) (types.BusinessVerification, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetBusinessVerificationKey(businessID))
	if bz == nil {
		return types.BusinessVerification{}, false
	}

	var verification types.BusinessVerification
	if err := json.Unmarshal(bz, &verification); err != nil {
		return types.BusinessVerification{}, false
	}
	return verification, true
}

// SetBusinessVerification stores a business verification
func (k Keeper) SetBusinessVerification(ctx sdk.Context, verification types.BusinessVerification) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(verification)
	if err != nil {
		return err
	}
	store.Set(types.GetBusinessVerificationKey(verification.BusinessID), bz)
	return nil
}

// DeleteBusinessVerification removes a business verification
func (k Keeper) DeleteBusinessVerification(ctx sdk.Context, businessID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetBusinessVerificationKey(businessID))
}

// GetAllBusinessVerifications returns all business verifications
func (k Keeper) GetAllBusinessVerifications(ctx sdk.Context) []types.BusinessVerification {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.BusinessVerificationPrefix)
	defer iterator.Close()

	var verifications []types.BusinessVerification
	for ; iterator.Valid(); iterator.Next() {
		var verification types.BusinessVerification
		if err := json.Unmarshal(iterator.Value(), &verification); err != nil {
			continue
		}
		verifications = append(verifications, verification)
	}
	return verifications
}

// =============================================================================
// VERIFIER STAKE CRUD
// =============================================================================

// GetVerifierStake returns a verifier's stake in a business
func (k Keeper) GetVerifierStake(ctx sdk.Context, verifierAddr, businessID string) (types.VerifierStake, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetVerifierStakeKey(verifierAddr, businessID))
	if bz == nil {
		return types.VerifierStake{}, false
	}

	var stake types.VerifierStake
	if err := json.Unmarshal(bz, &stake); err != nil {
		return types.VerifierStake{}, false
	}
	return stake, true
}

// SetVerifierStake stores a verifier's stake
func (k Keeper) SetVerifierStake(ctx sdk.Context, stake types.VerifierStake) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(stake)
	if err != nil {
		return err
	}
	store.Set(types.GetVerifierStakeKey(stake.VerifierAddress, stake.BusinessID), bz)
	return nil
}

// GetAllVerifierStakes returns all verifier stakes
func (k Keeper) GetAllVerifierStakes(ctx sdk.Context) []types.VerifierStake {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.VerifierStakePrefix)
	defer iterator.Close()

	var stakes []types.VerifierStake
	for ; iterator.Valid(); iterator.Next() {
		var stake types.VerifierStake
		if err := json.Unmarshal(iterator.Value(), &stake); err != nil {
			continue
		}
		stakes = append(stakes, stake)
	}
	return stakes
}

// =============================================================================
// VERIFICATION REWARDS CRUD
// =============================================================================

// GetVerificationRewards returns verification rewards for a verifier
func (k Keeper) GetVerificationRewards(ctx sdk.Context, verifierAddr string) (types.VerificationRewards, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetVerificationRewardsKey(verifierAddr))
	if bz == nil {
		return types.VerificationRewards{}, false
	}

	var rewards types.VerificationRewards
	if err := json.Unmarshal(bz, &rewards); err != nil {
		return types.VerificationRewards{}, false
	}
	return rewards, true
}

// SetVerificationRewards stores verification rewards
func (k Keeper) SetVerificationRewards(ctx sdk.Context, rewards types.VerificationRewards) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(rewards)
	if err != nil {
		return err
	}
	store.Set(types.GetVerificationRewardsKey(rewards.VerifierAddress), bz)
	return nil
}

// =============================================================================
// BUSINESS VERIFICATION LOGIC
// =============================================================================

// SubmitBusinessVerification submits a new business for verification
func (k Keeper) SubmitBusinessVerification(
	ctx sdk.Context,
	businessID string,
	companyName string,
	businessType string,
	registrationCountry string,
	documentHashes []string,
	totalShares int64,
	submitter string,
) error {
	// Check if verification already exists
	if _, exists := k.GetBusinessVerification(ctx, businessID); exists {
		return fmt.Errorf("verification already exists for business %s", businessID)
	}

	// Determine required tier based on business size
	businessSize := types.GetBusinessSizeFromTotalShares(totalShares)
	requiredTier := types.GetRequiredStakingTier(businessSize)

	// Check if submitter has required tier (using staking keeper)
	if k.stakingKeeper != nil {
		submitterAddr, err := sdk.AccAddressFromBech32(submitter)
		if err != nil {
			return fmt.Errorf("invalid submitter address: %w", err)
		}
		if !k.stakingKeeper.CanVerifyBusinessByShares(ctx, submitterAddr, totalShares) {
			return fmt.Errorf("submitter does not have required tier for this business size")
		}
	}

	params := k.GetVerificationParams(ctx)

	// Create verification
	verification := types.BusinessVerification{
		BusinessID:           businessID,
		CompanyName:          companyName,
		BusinessType:         businessType,
		RegistrationCountry:  registrationCountry,
		DocumentHashes:       documentHashes,
		RequiredTier:         requiredTier,
		StakeRequired:        math.ZeroInt(), // Could be calculated based on business size
		VerificationFee:      params.BaseVerificationReward,
		Status:               types.VerificationPending,
		AssignedVerifiers:    []string{},
		VerificationDeadline: ctx.BlockTime().Add(time.Duration(params.VerificationTimeoutDays) * 24 * time.Hour),
		SubmittedAt:          ctx.BlockTime(),
	}

	if err := k.SetBusinessVerification(ctx, verification); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"business_verification_submitted",
			sdk.NewAttribute("business_id", businessID),
			sdk.NewAttribute("company_name", companyName),
			sdk.NewAttribute("required_tier", fmt.Sprintf("%d", requiredTier)),
			sdk.NewAttribute("deadline", verification.VerificationDeadline.String()),
		),
	)

	k.Logger(ctx).Info("business verification submitted",
		"business_id", businessID,
		"company_name", companyName,
		"required_tier", requiredTier,
	)

	return nil
}

// SelectVerifiersForBusiness selects appropriate verifiers for a business verification
func (k Keeper) SelectVerifiersForBusiness(ctx sdk.Context, businessID string) ([]string, error) {
	verification, exists := k.GetBusinessVerification(ctx, businessID)
	if !exists {
		return nil, fmt.Errorf("business verification not found: %s", businessID)
	}

	params := k.GetVerificationParams(ctx)
	requiredCount := params.MinVerifiersRequired
	requiredTier := verification.RequiredTier

	// Get eligible verifiers from staking keeper
	var eligibleVerifiers []string
	if k.stakingKeeper != nil {
		// Get users with required tier and good reputation
		// For now, we'll use a simplified approach
		// In a full implementation, this would query the staking module for eligible users
		k.Logger(ctx).Info("selecting verifiers",
			"business_id", businessID,
			"required_tier", requiredTier,
			"required_count", requiredCount,
		)
	}

	// Update verification with selected verifiers
	verification.AssignedVerifiers = eligibleVerifiers
	verification.Status = types.VerificationInProgress
	if err := k.SetBusinessVerification(ctx, verification); err != nil {
		return nil, err
	}

	return eligibleVerifiers, nil
}

// CompleteVerification marks a verification as complete
func (k Keeper) CompleteVerification(
	ctx sdk.Context,
	businessID string,
	verifierAddr string,
	approved bool,
	notes string,
) error {
	verification, exists := k.GetBusinessVerification(ctx, businessID)
	if !exists {
		return fmt.Errorf("business verification not found: %s", businessID)
	}

	if verification.Status != types.VerificationPending && verification.Status != types.VerificationInProgress {
		return fmt.Errorf("verification is not in a completable state: %s", verification.Status.String())
	}

	// Check if verifier has required tier
	if k.stakingKeeper != nil {
		verifier, err := sdk.AccAddressFromBech32(verifierAddr)
		if err != nil {
			return fmt.Errorf("invalid verifier address: %w", err)
		}
		verifierTier := k.stakingKeeper.GetUserTierInt(ctx, verifier)
		if verifierTier < verification.RequiredTier {
			return fmt.Errorf("verifier tier %d is below required tier %d", verifierTier, verification.RequiredTier)
		}
	}

	// Update verification status
	if approved {
		verification.Status = types.VerificationApproved
	} else {
		verification.Status = types.VerificationRejected
	}
	verification.CompletedAt = ctx.BlockTime()
	verification.VerificationNotes = notes

	if err := k.SetBusinessVerification(ctx, verification); err != nil {
		return err
	}

	// Update verifier reputation through staking keeper
	if k.stakingKeeper != nil {
		verifier, _ := sdk.AccAddressFromBech32(verifierAddr)
		if approved {
			k.stakingKeeper.RewardSuccessfulVerification(ctx, verifier, businessID)
		} else {
			// No penalty for rejections - only for bad verifications
		}
	}

	// Calculate and award verification reward
	reward := k.CalculateVerificationReward(ctx, verification.RequiredTier)
	if err := k.AwardVerificationReward(ctx, verifierAddr, reward); err != nil {
		k.Logger(ctx).Error("failed to award verification reward", "error", err)
	}

	// Emit event
	status := "approved"
	if !approved {
		status = "rejected"
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"business_verification_completed",
			sdk.NewAttribute("business_id", businessID),
			sdk.NewAttribute("verifier", verifierAddr),
			sdk.NewAttribute("status", status),
			sdk.NewAttribute("reward", reward.String()),
		),
	)

	k.Logger(ctx).Info("business verification completed",
		"business_id", businessID,
		"verifier", verifierAddr,
		"approved", approved,
	)

	return nil
}

// DisputeVerification allows disputing a verification decision
func (k Keeper) DisputeVerification(ctx sdk.Context, businessID, disputeReason string) error {
	verification, exists := k.GetBusinessVerification(ctx, businessID)
	if !exists {
		return fmt.Errorf("business verification not found: %s", businessID)
	}

	verification.Status = types.VerificationDisputed
	verification.DisputeReason = disputeReason

	if err := k.SetBusinessVerification(ctx, verification); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"business_verification_disputed",
			sdk.NewAttribute("business_id", businessID),
			sdk.NewAttribute("reason", disputeReason),
		),
	)

	return nil
}

// =============================================================================
// REWARD CALCULATIONS
// =============================================================================

// CalculateVerificationReward calculates reward based on tier
func (k Keeper) CalculateVerificationReward(ctx sdk.Context, tier int) math.Int {
	params := k.GetVerificationParams(ctx)
	baseReward := params.BaseVerificationReward

	// Tier multiplier
	var multiplier int64 = 100 // 1.0x base
	switch tier {
	case types.StakeTierKeeper:
		multiplier = 100 // 1.0x
	case types.StakeTierWarden:
		multiplier = 150 // 1.5x
	case types.StakeTierSteward:
		multiplier = 200 // 2.0x
	case types.StakeTierArchon:
		multiplier = 250 // 2.5x
	}

	return baseReward.MulRaw(multiplier).QuoRaw(100)
}

// AwardVerificationReward awards HODL reward to a verifier
func (k Keeper) AwardVerificationReward(ctx sdk.Context, verifierAddr string, reward math.Int) error {
	// Get or create rewards record
	rewards, exists := k.GetVerificationRewards(ctx, verifierAddr)
	if !exists {
		rewards = types.VerificationRewards{
			VerifierAddress:        verifierAddr,
			TotalHODLRewards:       math.ZeroInt(),
			TotalEquityRewards:     math.ZeroInt(),
			VerificationsCompleted: 0,
			LastRewardClaim:        ctx.BlockTime(),
		}
	}

	rewards.TotalHODLRewards = rewards.TotalHODLRewards.Add(reward)
	rewards.VerificationsCompleted++

	return k.SetVerificationRewards(ctx, rewards)
}

// =============================================================================
// VESTING MANAGEMENT
// =============================================================================

// ProcessVestingSchedules checks and processes all vesting schedules
func (k Keeper) ProcessVestingSchedules(ctx sdk.Context) {
	stakes := k.GetAllVerifierStakes(ctx)
	currentTime := ctx.BlockTime()

	for _, stake := range stakes {
		if stake.IsVested {
			continue
		}

		vestingEndTime := stake.StakedAt.Add(stake.VestingPeriod)
		if currentTime.After(vestingEndTime) {
			stake.IsVested = true
			if err := k.SetVerifierStake(ctx, stake); err != nil {
				k.Logger(ctx).Error("failed to update vested stake", "error", err)
				continue
			}

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"equity_vested",
					sdk.NewAttribute("verifier", stake.VerifierAddress),
					sdk.NewAttribute("business_id", stake.BusinessID),
					sdk.NewAttribute("stake_amount", stake.StakeAmount.String()),
					sdk.NewAttribute("equity_percentage", stake.EquityPercentage.String()),
				),
			)
		}
	}
}
