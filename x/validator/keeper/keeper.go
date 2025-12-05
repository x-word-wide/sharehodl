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