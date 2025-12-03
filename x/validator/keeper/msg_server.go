package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/validator/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) *msgServer {
	return &msgServer{Keeper: keeper}
}

// RegisterValidatorTier handles validator tier registration
func (k msgServer) RegisterValidatorTier(goCtx context.Context, msg *types.SimpleMsgRegisterValidatorTier) (*types.MsgRegisterValidatorTierResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate validator address
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid validator address: %v", err)
	}

	// Check if validator exists in staking module
	_, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return nil, fmt.Errorf("validator not found in staking module")
	}

	// Check if validator already registered
	_, exists := k.GetValidatorTierInfo(ctx, valAddr)
	if exists {
		return nil, types.ErrValidatorAlreadyRegistered
	}

	// Validate stake amount meets tier requirement
	actualTier := k.DetermineValidatorTier(msg.StakeAmount)
	if actualTier < msg.DesiredTier {
		return nil, types.ErrInsufficientStakeForTier
	}

	// Convert validator address to account address for token operations
	accAddr := sdk.AccAddress(valAddr)

	// Check if validator has sufficient HODL balance
	hodlBalance := k.bankKeeper.GetBalance(ctx, accAddr, "hodl")
	if hodlBalance.Amount.LT(msg.StakeAmount) {
		return nil, fmt.Errorf("insufficient HODL balance: have %s, need %s", hodlBalance.Amount.String(), msg.StakeAmount.String())
	}

	// Transfer HODL tokens to module for staking
	hodlCoins := sdk.NewCoins(sdk.NewCoin("hodl", msg.StakeAmount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, accAddr, types.ModuleName, hodlCoins); err != nil {
		return nil, fmt.Errorf("failed to stake HODL tokens: %v", err)
	}

	// Create validator tier info
	tierInfo := types.ValidatorTierInfo{
		ValidatorAddress:        msg.ValidatorAddress,
		Tier:                   actualTier,
		StakedAmount:           msg.StakeAmount,
		VerificationCount:      0,
		SuccessfulVerifications: 0,
		TotalRewards:           math.ZeroInt(),
		LastVerification:       time.Time{},
		ReputationScore:        math.LegacyNewDec(100), // Start with perfect reputation
	}

	// Save tier info
	k.SetValidatorTierInfo(ctx, tierInfo)

	// Initialize verification rewards
	rewards := types.VerificationRewards{
		ValidatorAddress:       msg.ValidatorAddress,
		TotalHODLRewards:       math.ZeroInt(),
		TotalEquityRewards:     math.ZeroInt(),
		VerificationsCompleted: 0,
		LastRewardClaim:        time.Time{},
	}
	k.SetVerificationRewards(ctx, rewards)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"register_validator_tier",
			sdk.NewAttribute("validator", msg.ValidatorAddress),
			sdk.NewAttribute("tier", actualTier.String()),
			sdk.NewAttribute("staked_amount", msg.StakeAmount.String()),
		),
	)

	return &types.MsgRegisterValidatorTierResponse{
		Tier:         actualTier,
		StakedAmount: msg.StakeAmount,
	}, nil
}

// SubmitBusinessVerification handles business verification submission
func (k msgServer) SubmitBusinessVerification(goCtx context.Context, msg *types.SimpleMsgSubmitBusinessVerification) (*types.MsgSubmitBusinessVerificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate submitter address
	submitterAddr, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		return nil, fmt.Errorf("invalid submitter address: %v", err)
	}

	// Check if business already exists
	_, exists := k.GetBusinessVerification(ctx, msg.BusinessID)
	if exists {
		return nil, types.ErrBusinessAlreadyVerified
	}

	// Validate verification fee payment
	hodlBalance := k.bankKeeper.GetBalance(ctx, submitterAddr, "hodl")
	if hodlBalance.Amount.LT(msg.VerificationFee) {
		return nil, types.ErrVerificationFeeInsufficient
	}

	// Transfer verification fee to module
	feeCoins := sdk.NewCoins(sdk.NewCoin("hodl", msg.VerificationFee))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, submitterAddr, types.ModuleName, feeCoins); err != nil {
		return nil, fmt.Errorf("failed to pay verification fee: %v", err)
	}

	// Get parameters
	params := k.GetParams(ctx)

	// Select validators for verification
	selectedValidators := k.SelectValidatorsForVerification(ctx, msg.RequiredTier, params.MinValidatorsRequired)
	if uint64(len(selectedValidators)) < params.MinValidatorsRequired {
		return nil, types.ErrInsufficientValidators
	}

	// Calculate deadline
	deadline := time.Now().Add(time.Duration(params.VerificationTimeoutDays) * 24 * time.Hour)

	// Create business verification
	verification := types.BusinessVerification{
		BusinessID:           msg.BusinessID,
		CompanyName:          msg.CompanyName,
		BusinessType:         msg.BusinessType,
		RegistrationCountry:  msg.RegistrationCountry,
		DocumentHashes:       msg.DocumentHashes,
		RequiredTier:         msg.RequiredTier,
		StakeRequired:        msg.RequiredTier.GetStakeThreshold(),
		VerificationFee:      msg.VerificationFee,
		Status:              types.VerificationPending,
		AssignedValidators:  selectedValidators,
		VerificationDeadline: deadline,
		SubmittedAt:         time.Now(),
		VerificationNotes:   "",
		DisputeReason:       "",
	}

	// Save verification
	k.SetBusinessVerification(ctx, verification)

	// Generate verification ID (simplified)
	verificationID := fmt.Sprintf("verify_%s_%d", msg.BusinessID, ctx.BlockHeight())

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"submit_business_verification",
			sdk.NewAttribute("business_id", msg.BusinessID),
			sdk.NewAttribute("company_name", msg.CompanyName),
			sdk.NewAttribute("verification_id", verificationID),
			sdk.NewAttribute("assigned_validators", fmt.Sprintf("%v", selectedValidators)),
			sdk.NewAttribute("deadline", deadline.String()),
		),
	)

	return &types.MsgSubmitBusinessVerificationResponse{
		BusinessID:         msg.BusinessID,
		VerificationID:     verificationID,
		AssignedValidators: selectedValidators,
		Deadline:           deadline,
	}, nil
}

// CompleteVerification handles verification completion by validators
func (k msgServer) CompleteVerification(goCtx context.Context, msg *types.SimpleMsgCompleteVerification) (*types.MsgCompleteVerificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate validator address
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid validator address: %v", err)
	}

	// Get business verification
	verification, exists := k.GetBusinessVerification(ctx, msg.BusinessID)
	if !exists {
		return nil, types.ErrBusinessVerificationNotFound
	}

	// Check if verification is still pending
	if verification.Status != types.VerificationPending && verification.Status != types.VerificationInProgress {
		return nil, types.ErrVerificationAlreadyCompleted
	}

	// Check if validator is assigned to this verification
	isAssigned := false
	for _, assignedVal := range verification.AssignedValidators {
		if assignedVal == msg.ValidatorAddress {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return nil, types.ErrValidatorNotAssigned
	}

	// Check if verification has expired
	if time.Now().After(verification.VerificationDeadline) {
		verification.Status = types.VerificationRejected
		verification.DisputeReason = "Verification expired"
		k.SetBusinessVerification(ctx, verification)
		return nil, types.ErrVerificationExpired
	}

	// Get validator tier info
	tierInfo, exists := k.GetValidatorTierInfo(ctx, valAddr)
	if !exists {
		return nil, types.ErrValidatorTierNotFound
	}

	// Update verification with result
	verification.Status = msg.VerificationResult
	verification.VerificationNotes = msg.VerificationNotes
	verification.CompletedAt = time.Now()

	// Calculate rewards
	var rewardEarned math.Int
	var equityAllocated math.Int
	
	if msg.VerificationResult == types.VerificationApproved {
		// Calculate HODL reward
		rewardEarned = k.CalculateVerificationReward(ctx, tierInfo.Tier, verification.StakeRequired)
		
		// Calculate equity allocation if validator stakes
		if msg.StakeInBusiness.IsPositive() {
			params := k.GetParams(ctx)
			equityAllocated = params.EquityStakePercentage.MulInt(msg.StakeInBusiness).TruncateInt()
			
			// Create validator business stake
			stake := types.ValidatorBusinessStake{
				ValidatorAddress: msg.ValidatorAddress,
				BusinessID:       msg.BusinessID,
				StakeAmount:      msg.StakeInBusiness,
				EquityPercentage: params.EquityStakePercentage,
				StakedAt:         time.Now(),
				VestingPeriod:    365 * 24 * time.Hour, // 1 year vesting
				IsVested:         false,
			}
			k.SetValidatorBusinessStake(ctx, stake)
			
			// Transfer stake amount to module
			accAddr := sdk.AccAddress(valAddr)
			stakeCoins := sdk.NewCoins(sdk.NewCoin("hodl", msg.StakeInBusiness))
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, accAddr, types.ModuleName, stakeCoins); err != nil {
				return nil, fmt.Errorf("failed to stake in business: %v", err)
			}
		}
		
		// Mint HODL rewards to validator
		accAddr := sdk.AccAddress(valAddr)
		rewardCoins := sdk.NewCoins(sdk.NewCoin("hodl", rewardEarned))
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, rewardCoins); err != nil {
			return nil, fmt.Errorf("failed to mint rewards: %v", err)
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, accAddr, rewardCoins); err != nil {
			return nil, fmt.Errorf("failed to send rewards: %v", err)
		}
		
		// Update validator rewards
		rewards, exists := k.GetVerificationRewards(ctx, valAddr)
		if exists {
			rewards.TotalHODLRewards = rewards.TotalHODLRewards.Add(rewardEarned)
			rewards.TotalEquityRewards = rewards.TotalEquityRewards.Add(equityAllocated)
			rewards.VerificationsCompleted++
			k.SetVerificationRewards(ctx, rewards)
		}
	}

	// Update validator reputation
	k.UpdateValidatorReputation(ctx, valAddr, msg.VerificationResult == types.VerificationApproved)

	// Save updated verification
	k.SetBusinessVerification(ctx, verification)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"complete_verification",
			sdk.NewAttribute("validator", msg.ValidatorAddress),
			sdk.NewAttribute("business_id", msg.BusinessID),
			sdk.NewAttribute("result", msg.VerificationResult.String()),
			sdk.NewAttribute("reward_earned", rewardEarned.String()),
			sdk.NewAttribute("equity_allocated", equityAllocated.String()),
		),
	)

	return &types.MsgCompleteVerificationResponse{
		BusinessID:      msg.BusinessID,
		Status:          msg.VerificationResult,
		RewardEarned:    rewardEarned,
		EquityAllocated: equityAllocated,
	}, nil
}

// StakeInBusiness handles additional stakes in verified businesses
func (k msgServer) StakeInBusiness(goCtx context.Context, msg *types.SimpleMsgStakeInBusiness) (*types.MsgStakeInBusinessResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate validator address
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid validator address: %v", err)
	}

	// Check if business is verified and approved
	verification, exists := k.GetBusinessVerification(ctx, msg.BusinessID)
	if !exists {
		return nil, types.ErrBusinessVerificationNotFound
	}
	if verification.Status != types.VerificationApproved {
		return nil, types.ErrBusinessNotApproved
	}

	// Get validator tier info
	_, exists = k.GetValidatorTierInfo(ctx, valAddr)
	if !exists {
		return nil, types.ErrValidatorTierNotFound
	}

	// Check if validator has sufficient balance
	accAddr := sdk.AccAddress(valAddr)
	hodlBalance := k.bankKeeper.GetBalance(ctx, accAddr, "hodl")
	if hodlBalance.Amount.LT(msg.StakeAmount) {
		return nil, fmt.Errorf("insufficient HODL balance")
	}

	// Transfer stake to module
	stakeCoins := sdk.NewCoins(sdk.NewCoin("hodl", msg.StakeAmount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, accAddr, types.ModuleName, stakeCoins); err != nil {
		return nil, fmt.Errorf("failed to transfer stake: %v", err)
	}

	// Calculate equity percentage
	params := k.GetParams(ctx)
	equityPercentage := params.EquityStakePercentage

	// Create or update validator business stake
	existingStake, exists := k.GetValidatorBusinessStake(ctx, valAddr, msg.BusinessID)
	if exists {
		existingStake.StakeAmount = existingStake.StakeAmount.Add(msg.StakeAmount)
		existingStake.VestingPeriod = msg.VestingPeriod
		k.SetValidatorBusinessStake(ctx, existingStake)
	} else {
		stake := types.ValidatorBusinessStake{
			ValidatorAddress: msg.ValidatorAddress,
			BusinessID:       msg.BusinessID,
			StakeAmount:      msg.StakeAmount,
			EquityPercentage: equityPercentage,
			StakedAt:         time.Now(),
			VestingPeriod:    msg.VestingPeriod,
			IsVested:         false,
		}
		k.SetValidatorBusinessStake(ctx, stake)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"stake_in_business",
			sdk.NewAttribute("validator", msg.ValidatorAddress),
			sdk.NewAttribute("business_id", msg.BusinessID),
			sdk.NewAttribute("stake_amount", msg.StakeAmount.String()),
			sdk.NewAttribute("equity_percentage", equityPercentage.String()),
		),
	)

	return &types.MsgStakeInBusinessResponse{
		StakeAmount:      msg.StakeAmount,
		EquityPercentage: equityPercentage,
		VestingPeriod:    msg.VestingPeriod,
	}, nil
}

// ClaimVerificationRewards handles reward claiming
func (k msgServer) ClaimVerificationRewards(goCtx context.Context, msg *types.SimpleMsgClaimVerificationRewards) (*types.MsgClaimVerificationRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate validator address
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid validator address: %v", err)
	}

	// Get verification rewards
	rewards, exists := k.GetVerificationRewards(ctx, valAddr)
	if !exists {
		return nil, fmt.Errorf("no rewards found for validator")
	}

	// For now, return the current reward totals (claiming mechanism to be implemented)
	return &types.MsgClaimVerificationRewardsResponse{
		HODLRewardsClaimed:   rewards.TotalHODLRewards,
		EquityRewardsClaimed: rewards.TotalEquityRewards,
	}, nil
}