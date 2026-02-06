package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/validator/types"
)

// Keeper of the validator store
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	memKey     storetypes.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
	// hodlKeeper removed for now - use bank keeper directly
}

// NewKeeper creates a new validator Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}
	
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)
}

// ValidatorTier operations

// GetValidatorTierInfo returns a validator's tier information
func (k Keeper) GetValidatorTierInfo(ctx sdk.Context, valAddr sdk.ValAddress) (types.ValidatorTierInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ValidatorTierKey(valAddr))
	if bz == nil {
		return types.ValidatorTierInfo{}, false
	}
	
	var tierInfo types.ValidatorTierInfo
	k.cdc.MustUnmarshal(bz, &tierInfo)
	return tierInfo, true
}

// SetValidatorTierInfo sets a validator's tier information
func (k Keeper) SetValidatorTierInfo(ctx sdk.Context, tierInfo types.ValidatorTierInfo) {
	store := ctx.KVStore(k.storeKey)
	valAddr, _ := sdk.ValAddressFromBech32(tierInfo.ValidatorAddress)
	bz := k.cdc.MustMarshal(&tierInfo)
	store.Set(types.ValidatorTierKey(valAddr), bz)
}

// GetAllValidatorTiers returns all validator tier information
func (k Keeper) GetAllValidatorTiers(ctx sdk.Context) []types.ValidatorTierInfo {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorTierPrefix)
	defer iterator.Close()
	
	tiers := []types.ValidatorTierInfo{}
	for ; iterator.Valid(); iterator.Next() {
		var tierInfo types.ValidatorTierInfo
		k.cdc.MustUnmarshal(iterator.Value(), &tierInfo)
		tiers = append(tiers, tierInfo)
	}
	return tiers
}

// Business verification operations

// GetBusinessVerification returns a business verification
func (k Keeper) GetBusinessVerification(ctx sdk.Context, businessID string) (types.BusinessVerification, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BusinessVerificationKey(businessID))
	if bz == nil {
		return types.BusinessVerification{}, false
	}
	
	var verification types.BusinessVerification
	k.cdc.MustUnmarshal(bz, &verification)
	return verification, true
}

// SetBusinessVerification sets a business verification
func (k Keeper) SetBusinessVerification(ctx sdk.Context, verification types.BusinessVerification) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&verification)
	store.Set(types.BusinessVerificationKey(verification.BusinessID), bz)
}

// GetAllBusinessVerifications returns all business verifications
func (k Keeper) GetAllBusinessVerifications(ctx sdk.Context) []types.BusinessVerification {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.BusinessVerificationPrefix)
	defer iterator.Close()
	
	verifications := []types.BusinessVerification{}
	for ; iterator.Valid(); iterator.Next() {
		var verification types.BusinessVerification
		k.cdc.MustUnmarshal(iterator.Value(), &verification)
		verifications = append(verifications, verification)
	}
	return verifications
}

// Validator business stake operations

// GetValidatorBusinessStake returns a validator's stake in a business
func (k Keeper) GetValidatorBusinessStake(ctx sdk.Context, valAddr sdk.ValAddress, businessID string) (types.ValidatorBusinessStake, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ValidatorBusinessStakeKey(valAddr, businessID))
	if bz == nil {
		return types.ValidatorBusinessStake{}, false
	}
	
	var stake types.ValidatorBusinessStake
	k.cdc.MustUnmarshal(bz, &stake)
	return stake, true
}

// SetValidatorBusinessStake sets a validator's stake in a business
func (k Keeper) SetValidatorBusinessStake(ctx sdk.Context, stake types.ValidatorBusinessStake) {
	store := ctx.KVStore(k.storeKey)
	valAddr, _ := sdk.ValAddressFromBech32(stake.ValidatorAddress)
	bz := k.cdc.MustMarshal(&stake)
	store.Set(types.ValidatorBusinessStakeKey(valAddr, stake.BusinessID), bz)
}

// GetValidatorBusinessStakes returns all stakes for a validator
func (k Keeper) GetValidatorBusinessStakes(ctx sdk.Context, valAddr sdk.ValAddress) []types.ValidatorBusinessStake {
	store := ctx.KVStore(k.storeKey)
	prefix := append(types.ValidatorBusinessStakePrefix, valAddr.Bytes()...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	
	stakes := []types.ValidatorBusinessStake{}
	for ; iterator.Valid(); iterator.Next() {
		var stake types.ValidatorBusinessStake
		k.cdc.MustUnmarshal(iterator.Value(), &stake)
		stakes = append(stakes, stake)
	}
	return stakes
}

// Verification rewards operations

// GetVerificationRewards returns verification rewards for a validator
func (k Keeper) GetVerificationRewards(ctx sdk.Context, valAddr sdk.ValAddress) (types.VerificationRewards, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.VerificationRewardsKey(valAddr))
	if bz == nil {
		return types.VerificationRewards{}, false
	}
	
	var rewards types.VerificationRewards
	k.cdc.MustUnmarshal(bz, &rewards)
	return rewards, true
}

// SetVerificationRewards sets verification rewards for a validator
func (k Keeper) SetVerificationRewards(ctx sdk.Context, rewards types.VerificationRewards) {
	store := ctx.KVStore(k.storeKey)
	valAddr, _ := sdk.ValAddressFromBech32(rewards.ValidatorAddress)
	bz := k.cdc.MustMarshal(&rewards)
	store.Set(types.VerificationRewardsKey(valAddr), bz)
}

// Helper functions

// DetermineValidatorTier determines the tier based on staked amount
func (k Keeper) DetermineValidatorTier(ctx sdk.Context, stakedAmount math.Int) types.ValidatorTier {
	// Check if we're on testnet based on chain ID
	isTestnet := ctx.ChainID() == "sharehodl-testnet-1" || ctx.ChainID() == "sharehodl-testnet-2"
	
	return types.GetTierFromStake(stakedAmount, isTestnet)
}

// SelectValidatorsForVerification selects appropriate validators for business verification
func (k Keeper) SelectValidatorsForVerification(ctx sdk.Context, requiredTier types.ValidatorTier, count uint64) []string {
	allTiers := k.GetAllValidatorTiers(ctx)
	
	// Filter validators by tier and reputation
	eligibleValidators := []types.ValidatorTierInfo{}
	for _, tierInfo := range allTiers {
		if tierInfo.Tier >= requiredTier && tierInfo.ReputationScore.GTE(math.LegacyNewDec(70)) {
			eligibleValidators = append(eligibleValidators, tierInfo)
		}
	}
	
	// Select validators (simplified selection for now)
	selectedValidators := []string{}
	for i, validator := range eligibleValidators {
		if uint64(i) >= count {
			break
		}
		selectedValidators = append(selectedValidators, validator.ValidatorAddress)
	}
	
	return selectedValidators
}

// CalculateVerificationReward calculates the reward for completing a verification
func (k Keeper) CalculateVerificationReward(ctx sdk.Context, tier types.ValidatorTier, businessValue math.Int) math.Int {
	params := k.GetParams(ctx)
	baseReward := params.BaseVerificationReward
	
	// Tier multiplier
	tierMultiplier := math.LegacyNewDec(1)
	switch tier {
	case types.TierSilver:
		tierMultiplier = math.LegacyNewDecWithPrec(15, 1) // 1.5x
	case types.TierGold:
		tierMultiplier = math.LegacyNewDec(2) // 2x
	case types.TierPlatinum:
		tierMultiplier = math.LegacyNewDecWithPrec(25, 1) // 2.5x
	}
	
	// Business value bonus (small percentage)
	valueBonus := math.LegacyNewDecFromInt(businessValue).Mul(math.LegacyNewDecWithPrec(1, 4)) // 0.01%
	
	totalReward := tierMultiplier.MulInt(baseReward).Add(valueBonus)
	return totalReward.TruncateInt()
}

// UpdateValidatorReputation updates a validator's reputation score
func (k Keeper) UpdateValidatorReputation(ctx sdk.Context, valAddr sdk.ValAddress, successful bool) {
	tierInfo, exists := k.GetValidatorTierInfo(ctx, valAddr)
	if !exists {
		return
	}
	
	tierInfo.VerificationCount++
	if successful {
		tierInfo.SuccessfulVerifications++
	}
	
	// Calculate new reputation score (simple success rate for now)
	if tierInfo.VerificationCount > 0 {
		successRate := math.LegacyNewDec(int64(tierInfo.SuccessfulVerifications)).QuoInt64(int64(tierInfo.VerificationCount))
		tierInfo.ReputationScore = successRate.Mul(math.LegacyNewDec(100))
	}
	
	tierInfo.LastVerification = time.Now()
	k.SetValidatorTierInfo(ctx, tierInfo)
}

// GetValidator returns validator information for equity keeper compatibility
func (k Keeper) GetValidator(ctx sdk.Context, address string) (interface{}, bool) {
	// Convert address string to validator address
	valAddr, err := sdk.ValAddressFromBech32(address)
	if err != nil {
		return nil, false
	}
	
	// Check if validator has tier info
	tierInfo, exists := k.GetValidatorTierInfo(ctx, valAddr)
	if !exists {
		return nil, false
	}
	
	// Return simplified validator info
	validatorInfo := struct {
		Address string
		Active  bool
		Tier    int
	}{
		Address: address,
		Active:  true,
		Tier:    int(tierInfo.Tier),
	}
	
	return validatorInfo, true
}

// GetAllValidators returns all validator information for equity keeper compatibility
func (k Keeper) GetAllValidators(ctx sdk.Context) []interface{} {
	validators := []interface{}{}

	// Iterate through all validator tier info
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorTierInfoPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tierInfo types.ValidatorTierInfo
		k.cdc.MustUnmarshal(iterator.Value(), &tierInfo)

		validatorInfo := struct {
			Address string
			Active  bool
			Tier    int
		}{
			Address: tierInfo.ValidatorAddress,
			Active:  true,
			Tier:    int(tierInfo.Tier),
		}

		validators = append(validators, validatorInfo)
	}

	return validators
}

// GetValidatorTier returns the tier for a validator by string address (governance compatibility)
func (k Keeper) GetValidatorTier(ctx sdk.Context, addr string) types.ValidatorTier {
	valAddr, err := sdk.ValAddressFromBech32(addr)
	if err != nil {
		return types.TierBronze // Default to Bronze if address invalid
	}

	tierInfo, exists := k.GetValidatorTierInfo(ctx, valAddr)
	if !exists {
		return types.TierBronze // Default to Bronze if not found
	}

	return tierInfo.Tier
}

// GetValidatorTierInt returns the validator's tier as an int (for interface compatibility)
func (k Keeper) GetValidatorTierInt(ctx sdk.Context, addr string) int {
	return int(k.GetValidatorTier(ctx, addr))
}

// IsValidatorActive checks if a validator is active by string address
func (k Keeper) IsValidatorActive(ctx sdk.Context, addr string) bool {
	valAddr, err := sdk.ValAddressFromBech32(addr)
	if err != nil {
		return false
	}

	tierInfo, exists := k.GetValidatorTierInfo(ctx, valAddr)
	if !exists {
		return false
	}

	// Consider validator active if they have tier info and were not recently demoted
	return tierInfo.Tier >= types.TierBronze
}

// ============ Vesting Management ============

// ProcessVestingSchedules checks and processes all vesting schedules
func (k Keeper) ProcessVestingSchedules(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorBusinessStakePrefix)
	defer iterator.Close()

	currentTime := ctx.BlockTime()

	for ; iterator.Valid(); iterator.Next() {
		var stake types.ValidatorBusinessStake
		k.cdc.MustUnmarshal(iterator.Value(), &stake)

		// Skip already vested stakes
		if stake.IsVested {
			continue
		}

		// Check if vesting period has completed
		vestingEndTime := stake.StakedAt.Add(stake.VestingPeriod)
		if currentTime.After(vestingEndTime) {
			stake.IsVested = true
			valAddr, _ := sdk.ValAddressFromBech32(stake.ValidatorAddress)
			k.SetValidatorBusinessStake(ctx, stake)

			// Emit vesting complete event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"equity_vested",
					sdk.NewAttribute("validator", stake.ValidatorAddress),
					sdk.NewAttribute("business_id", stake.BusinessID),
					sdk.NewAttribute("stake_amount", stake.StakeAmount.String()),
					sdk.NewAttribute("equity_percentage", stake.EquityPercentage.String()),
				),
			)

			k.Logger(ctx).Info("Equity vested",
				"validator", stake.ValidatorAddress,
				"business", stake.BusinessID,
				"amount", stake.StakeAmount.String())
			_ = valAddr // silence unused warning
		}
	}
}

// ClaimVestedEquity allows validator to claim their vested equity
func (k Keeper) ClaimVestedEquity(ctx sdk.Context, valAddr sdk.ValAddress, businessID string) error {
	stake, exists := k.GetValidatorBusinessStake(ctx, valAddr, businessID)
	if !exists {
		return fmt.Errorf("no stake found for business %s", businessID)
	}

	if !stake.IsVested {
		// Check if vesting has completed
		vestingEndTime := stake.StakedAt.Add(stake.VestingPeriod)
		if ctx.BlockTime().Before(vestingEndTime) {
			return fmt.Errorf("equity not yet vested, vesting completes at %s", vestingEndTime.String())
		}
		// Update vested status
		stake.IsVested = true
		k.SetValidatorBusinessStake(ctx, stake)
	}

	// In a full implementation, this would:
	// 1. Calculate the actual equity shares based on stake.EquityPercentage
	// 2. Transfer equity tokens from company treasury to validator
	// 3. Update cap table

	// For now, emit event that equity is claimable
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"equity_claimed",
			sdk.NewAttribute("validator", stake.ValidatorAddress),
			sdk.NewAttribute("business_id", businessID),
			sdk.NewAttribute("equity_percentage", stake.EquityPercentage.String()),
		),
	)

	return nil
}

// ============ Slashing Mechanism ============

// SlashValidator slashes a validator's stake and demotes their tier
func (k Keeper) SlashValidator(ctx sdk.Context, valAddr sdk.ValAddress, reason string, slashFraction math.LegacyDec) error {
	tierInfo, exists := k.GetValidatorTierInfo(ctx, valAddr)
	if !exists {
		return types.ErrValidatorTierNotFound
	}

	// Validate slash fraction
	if slashFraction.LT(math.LegacyZeroDec()) || slashFraction.GT(math.LegacyOneDec()) {
		return fmt.Errorf("invalid slash fraction: must be between 0 and 1")
	}

	// Calculate slash amount
	slashAmount := slashFraction.MulInt(tierInfo.StakedAmount).TruncateInt()
	if slashAmount.IsZero() {
		return nil // Nothing to slash
	}

	// CRITICAL: Burn slashed tokens FIRST - must succeed before modifying state
	slashCoins := sdk.NewCoins(sdk.NewCoin("hodl", slashAmount))
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, slashCoins); err != nil {
		k.Logger(ctx).Error("failed to burn slashed tokens, aborting slash", "error", err)
		return fmt.Errorf("failed to burn slashed tokens: %w", err)
	}

	// Update state ONLY after successful burn
	// Reduce staked amount
	tierInfo.StakedAmount = tierInfo.StakedAmount.Sub(slashAmount)
	if tierInfo.StakedAmount.IsNegative() {
		tierInfo.StakedAmount = math.ZeroInt()
	}

	// Reduce reputation score significantly
	tierInfo.ReputationScore = tierInfo.ReputationScore.Sub(math.LegacyNewDec(20))
	if tierInfo.ReputationScore.LT(math.LegacyZeroDec()) {
		tierInfo.ReputationScore = math.LegacyZeroDec()
	}

	// Recalculate tier based on new stake
	isTestnet := ctx.ChainID() == "sharehodl-testnet-1" || ctx.ChainID() == "sharehodl-testnet-2"
	newTier := types.GetTierFromStake(tierInfo.StakedAmount, isTestnet)

	// Demote tier
	tierInfo.Tier = newTier

	// Save updated tier info
	k.SetValidatorTierInfo(ctx, tierInfo)

	// Emit slashing event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_slashed",
			sdk.NewAttribute("validator", tierInfo.ValidatorAddress),
			sdk.NewAttribute("reason", reason),
			sdk.NewAttribute("slash_amount", slashAmount.String()),
			sdk.NewAttribute("slash_fraction", slashFraction.String()),
			sdk.NewAttribute("new_tier", newTier.String()),
			sdk.NewAttribute("new_reputation", tierInfo.ReputationScore.String()),
		),
	)

	k.Logger(ctx).Info("Validator slashed",
		"validator", tierInfo.ValidatorAddress,
		"reason", reason,
		"amount", slashAmount.String(),
		"new_tier", newTier.String())

	return nil
}

// SlashForDowntime slashes a validator for being offline
func (k Keeper) SlashForDowntime(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return k.SlashValidator(ctx, valAddr, "downtime", math.LegacyNewDecWithPrec(1, 2)) // 1% slash
}

// SlashForDoubleSign slashes a validator for double signing (more severe)
func (k Keeper) SlashForDoubleSign(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return k.SlashValidator(ctx, valAddr, "double_sign", math.LegacyNewDecWithPrec(5, 2)) // 5% slash
}

// SlashForBadVerification slashes a validator for fraudulent verification
func (k Keeper) SlashForBadVerification(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return k.SlashValidator(ctx, valAddr, "bad_verification", math.LegacyNewDecWithPrec(10, 2)) // 10% slash
}

// ============ Tier Demotion ============

// CheckAndDemoteTiers checks all validators and demotes those with low reputation
func (k Keeper) CheckAndDemoteTiers(ctx sdk.Context) {
	allTiers := k.GetAllValidatorTiers(ctx)

	for _, tierInfo := range allTiers {
		// Check if reputation is too low for current tier
		minReputation := k.getMinReputationForTier(tierInfo.Tier)

		if tierInfo.ReputationScore.LT(minReputation) {
			// Demote to lower tier
			newTier := k.getDemotedTier(tierInfo.Tier)
			if newTier != tierInfo.Tier {
				tierInfo.Tier = newTier
				k.SetValidatorTierInfo(ctx, tierInfo)

				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						"validator_demoted",
						sdk.NewAttribute("validator", tierInfo.ValidatorAddress),
						sdk.NewAttribute("old_tier", tierInfo.Tier.String()),
						sdk.NewAttribute("new_tier", newTier.String()),
						sdk.NewAttribute("reason", "low_reputation"),
					),
				)
			}
		}
	}
}

// getMinReputationForTier returns minimum reputation required for a tier
func (k Keeper) getMinReputationForTier(tier types.ValidatorTier) math.LegacyDec {
	switch tier {
	case types.TierPlatinum:
		return math.LegacyNewDec(95)
	case types.TierGold:
		return math.LegacyNewDec(85)
	case types.TierSilver:
		return math.LegacyNewDec(70)
	default:
		return math.LegacyNewDec(50)
	}
}

// getDemotedTier returns the next lower tier
func (k Keeper) getDemotedTier(tier types.ValidatorTier) types.ValidatorTier {
	switch tier {
	case types.TierPlatinum:
		return types.TierGold
	case types.TierGold:
		return types.TierSilver
	case types.TierSilver:
		return types.TierBronze
	default:
		return types.TierBronze
	}
}

// ============ All Business Stakes ============

// GetAllValidatorBusinessStakes returns all business stakes for all validators
func (k Keeper) GetAllValidatorBusinessStakes(ctx sdk.Context) []types.ValidatorBusinessStake {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorBusinessStakePrefix)
	defer iterator.Close()

	stakes := []types.ValidatorBusinessStake{}
	for ; iterator.Valid(); iterator.Next() {
		var stake types.ValidatorBusinessStake
		k.cdc.MustUnmarshal(iterator.Value(), &stake)
		stakes = append(stakes, stake)
	}
	return stakes
}

// ============ Governance Integration Methods ============

// IterateValidators iterates through all validators
// Used by governance module to calculate voting power
func (k Keeper) IterateValidators(ctx sdk.Context, fn func(valAddr string) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ValidatorTierPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var tierInfo types.ValidatorTierInfo
		k.cdc.MustUnmarshal(iterator.Value(), &tierInfo)

		// Only include active validators
		if tierInfo.Tier >= types.TierBronze {
			if fn(tierInfo.ValidatorAddress) {
				break
			}
		}
	}
}

// GetValidatorBondedTokens returns the bonded tokens for a validator
// Used by governance module to calculate voting power
func (k Keeper) GetValidatorBondedTokens(ctx sdk.Context, valAddr string) math.Int {
	addr, err := sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return math.ZeroInt()
	}

	tierInfo, exists := k.GetValidatorTierInfo(ctx, addr)
	if !exists {
		return math.ZeroInt()
	}

	return tierInfo.StakedAmount
}