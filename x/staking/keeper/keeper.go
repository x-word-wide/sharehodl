package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/staking/types"
)

// Keeper of the universal staking store
type Keeper struct {
	cdc       codec.BinaryCodec
	storeKey  storetypes.StoreKey
	memKey    storetypes.StoreKey
	authority string

	accountKeeper        types.AccountKeeper
	bankKeeper           types.BankKeeper
	governanceKeeper     types.GovernanceKeeper
	feeAbstractionKeeper types.FeeAbstractionKeeper
	hooks                types.StakingHooks
}

// NewKeeper creates a new universal staking Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	memKey storetypes.StoreKey,
	authority string,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	governanceKeeper types.GovernanceKeeper,
	_ interface{}, // Placeholder for removed ValidatorKeeper - kept for API compatibility
	feeAbstractionKeeper types.FeeAbstractionKeeper,
) *Keeper {
	return &Keeper{
		cdc:                  cdc,
		storeKey:             storeKey,
		memKey:               memKey,
		authority:            authority,
		accountKeeper:        accountKeeper,
		bankKeeper:           bankKeeper,
		governanceKeeper:     governanceKeeper,
		feeAbstractionKeeper: feeAbstractionKeeper,
	}
}

// SetHooks sets the staking hooks
func (k *Keeper) SetHooks(hooks types.StakingHooks) {
	k.hooks = hooks
}

// GetAuthority returns the module authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// IsTestnet checks if running on testnet
func (k Keeper) IsTestnet(ctx sdk.Context) bool {
	chainID := ctx.ChainID()
	return chainID == "sharehodl-testnet-1" || chainID == "sharehodl-testnet-2" || chainID == "sharehodl-local"
}

// =============================================================================
// PARAMS
// =============================================================================

// GetParams returns the staking parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

// SetParams sets the staking parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}

// =============================================================================
// USER STAKE MANAGEMENT
// =============================================================================

// GetUserStake returns a user's stake
func (k Keeper) GetUserStake(ctx sdk.Context, owner sdk.AccAddress) (types.UserStake, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetUserStakeKey(owner)

	bz := store.Get(key)
	if bz == nil {
		return types.UserStake{}, false
	}

	var stake types.UserStake
	if err := json.Unmarshal(bz, &stake); err != nil {
		return types.UserStake{}, false
	}
	return stake, true
}

// SetUserStake stores a user's stake
func (k Keeper) SetUserStake(ctx sdk.Context, stake types.UserStake) error {
	owner, err := sdk.AccAddressFromBech32(stake.Owner)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetUserStakeKey(owner)

	bz, err := json.Marshal(stake)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// DeleteUserStake removes a user's stake
func (k Keeper) DeleteUserStake(ctx sdk.Context, owner sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetUserStakeKey(owner)
	store.Delete(key)
}

// IterateUserStakes iterates over all user stakes
func (k Keeper) IterateUserStakes(ctx sdk.Context, fn func(stake types.UserStake) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.UserStakePrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var stake types.UserStake
		if err := json.Unmarshal(iterator.Value(), &stake); err != nil {
			continue
		}
		if fn(stake) {
			break
		}
	}
}

// GetAllUserStakes returns all user stakes
func (k Keeper) GetAllUserStakes(ctx sdk.Context) []types.UserStake {
	var stakes []types.UserStake
	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		stakes = append(stakes, stake)
		return false
	})
	return stakes
}

// =============================================================================
// STAKING OPERATIONS
// =============================================================================

// Stake allows a user to stake HODL
func (k Keeper) Stake(ctx sdk.Context, staker sdk.AccAddress, amount math.Int) error {
	params := k.GetParams(ctx)

	// Check minimum stake
	if amount.LT(params.MinStakeAmount) {
		return types.ErrBelowMinimumStake
	}

	// Get existing stake or create new
	existingStake, exists := k.GetUserStake(ctx, staker)

	var newAmount math.Int
	var oldTier types.StakeTier

	if exists {
		newAmount = existingStake.StakedAmount.Add(amount)
		oldTier = existingStake.Tier
	} else {
		newAmount = amount
		oldTier = types.StakeTier(-1)
	}

	// Determine new tier using governance-controlled thresholds
	var newTier types.StakeTier
	if k.IsTestnet(ctx) {
		// Use hardcoded testnet thresholds (1000x lower for testing)
		newTier = types.GetTierFromStake(newAmount, true)
	} else {
		// Use governance-controlled thresholds from params
		newTier = types.GetTierFromStakeWithParams(newAmount, params)
	}
	if newTier < 0 {
		return types.ErrBelowMinimumStake
	}

	// Transfer tokens from user to staking pool
	coins := sdk.NewCoins(sdk.NewCoin("uhodl", amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, staker, types.StakingPoolName, coins); err != nil {
		return err
	}

	// Update or create stake
	var stake types.UserStake
	if exists {
		stake = existingStake
		stake.StakedAmount = newAmount
		stake.Tier = newTier
	} else {
		stake = types.NewUserStake(staker.String(), newAmount, newTier)
	}

	if err := k.SetUserStake(ctx, stake); err != nil {
		return err
	}

	// Update tier counts
	if oldTier >= 0 && oldTier != newTier {
		k.decrementTierCount(ctx, oldTier)
	}
	if !exists || oldTier != newTier {
		k.incrementTierCount(ctx, newTier)
	}

	// Emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStake,
			sdk.NewAttribute(types.AttributeKeyStaker, staker.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyTier, newTier.String()),
		),
	)

	if oldTier >= 0 && oldTier != newTier {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTierChange,
				sdk.NewAttribute(types.AttributeKeyStaker, staker.String()),
				sdk.NewAttribute(types.AttributeKeyOldTier, oldTier.String()),
				sdk.NewAttribute(types.AttributeKeyNewTier, newTier.String()),
			),
		)

		// Call hooks
		if k.hooks != nil {
			if err := k.hooks.AfterTierChange(ctx, staker, oldTier, newTier); err != nil {
				k.Logger(ctx).Error("tier change hook failed", "error", err)
			}
		}
	}

	// Call hooks
	if k.hooks != nil {
		if err := k.hooks.AfterStake(ctx, staker, amount, newTier); err != nil {
			k.Logger(ctx).Error("stake hook failed", "error", err)
		}
	}

	k.Logger(ctx).Info("user staked",
		"staker", staker.String(),
		"amount", amount.String(),
		"total", newAmount.String(),
		"tier", newTier.String(),
	)

	return nil
}

// Unstake initiates the unstaking process with a 14-day unbonding period
// User must have no active locks (company listings, loans, disputes, etc.)
func (k Keeper) Unstake(ctx sdk.Context, staker sdk.AccAddress, amount math.Int) error {
	stake, exists := k.GetUserStake(ctx, staker)
	if !exists {
		return types.ErrStakeNotFound
	}

	if amount.GT(stake.StakedAmount) {
		return types.ErrInvalidAmount
	}

	// CHECK FOR ACTIVE LOCKS - Cannot unstake if any obligations exist
	canUnstake, activeLocks := k.CanUnstake(ctx, staker)
	if !canUnstake {
		// Return detailed error about what's locking the stake
		if len(activeLocks) > 0 {
			lock := activeLocks[0]
			switch lock.LockType {
			case types.LockTypeCompanyListing:
				return types.ErrActiveCompanyListing
			case types.LockTypePendingListing:
				return types.ErrPendingListing
			case types.LockTypeActiveLoan:
				return types.ErrActiveLoan
			case types.LockTypeActiveDispute:
				return types.ErrActiveDispute
			case types.LockTypePendingVote:
				return types.ErrPendingVote
			case types.LockTypeBan:
				return types.ErrUserBanned
			case types.LockTypeValidator:
				return types.ErrActiveValidator
			case types.LockTypeModerator:
				return types.ErrActiveModerator
			default:
				return types.ErrStakeLocked
			}
		}
		return types.ErrStakeLocked
	}

	// Check for existing unbonding request
	if _, exists := k.GetUnbondingRequest(ctx, staker); exists {
		return types.ErrUnbondingInProgress
	}

	// CHECK STAKE-AS-TRUST-CEILING COMMITMENTS
	// Cannot unstake below total committed amount (escrow, lending, P2P)
	canUnstakeAmount, reason := k.CanUnstakeAmount(ctx, staker, amount)
	if !canUnstakeAmount {
		k.Logger(ctx).Info("unstake blocked by commitments", "staker", staker.String(), "reason", reason)
		return types.ErrInsufficientAvailable
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
		k.DeleteUserStake(ctx, staker)
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

	// Create unbonding request (14-day waiting period)
	if err := k.CreateUnbondingRequest(ctx, staker, amount); err != nil {
		return err
	}

	// Emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(types.AttributeKeyStaker, staker.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	if oldTier != newTier {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeTierChange,
				sdk.NewAttribute(types.AttributeKeyStaker, staker.String()),
				sdk.NewAttribute(types.AttributeKeyOldTier, oldTier.String()),
				sdk.NewAttribute(types.AttributeKeyNewTier, newTier.String()),
			),
		)

		if k.hooks != nil {
			if err := k.hooks.AfterTierChange(ctx, staker, oldTier, newTier); err != nil {
				k.Logger(ctx).Error("tier change hook failed", "error", err)
			}
		}
	}

	if k.hooks != nil {
		if err := k.hooks.AfterUnstake(ctx, staker, amount, oldTier); err != nil {
			k.Logger(ctx).Error("unstake hook failed", "error", err)
		}
	}

	k.Logger(ctx).Info("unstaking initiated (14-day unbonding period)",
		"staker", staker.String(),
		"amount", amount.String(),
		"remaining", newAmount.String(),
		"old_tier", oldTier.String(),
		"new_tier", newTier.String(),
	)

	return nil
}

// =============================================================================
// TIER MANAGEMENT
// =============================================================================

// GetTierCount returns the number of users in a tier
func (k Keeper) GetTierCount(ctx sdk.Context, tier types.StakeTier) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTierCountKey(tier)
	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

func (k Keeper) incrementTierCount(ctx sdk.Context, tier types.StakeTier) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTierCountKey(tier)
	count := k.GetTierCount(ctx, tier)
	store.Set(key, sdk.Uint64ToBigEndian(count+1))
}

func (k Keeper) decrementTierCount(ctx sdk.Context, tier types.StakeTier) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTierCountKey(tier)
	count := k.GetTierCount(ctx, tier)
	if count > 0 {
		store.Set(key, sdk.Uint64ToBigEndian(count-1))
	}
}

// GetTierStats returns statistics for a tier
func (k Keeper) GetTierStats(ctx sdk.Context, tier types.StakeTier) types.TierStats {
	stats := types.TierStats{
		Tier:        tier,
		MemberCount: k.GetTierCount(ctx, tier),
		TotalStaked: math.ZeroInt(),
		TotalWeight: math.LegacyZeroDec(),
	}

	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		if stake.Tier == tier {
			stats.TotalStaked = stats.TotalStaked.Add(stake.StakedAmount)
			stats.TotalWeight = stats.TotalWeight.Add(stake.CalculateWeight())
		}
		return false
	})

	return stats
}

// GetAllTierStats returns statistics for all tiers
func (k Keeper) GetAllTierStats(ctx sdk.Context) []types.TierStats {
	tiers := []types.StakeTier{
		types.TierHolder,
		types.TierKeeper,
		types.TierWarden,
		types.TierSteward,
		types.TierArchon,
		types.TierValidator,
	}

	var allStats []types.TierStats
	for _, tier := range tiers {
		allStats = append(allStats, k.GetTierStats(ctx, tier))
	}
	return allStats
}

// =============================================================================
// TIER CHECKS (for other modules)
// =============================================================================

// GetUserTier returns a user's current tier
func (k Keeper) GetUserTier(ctx sdk.Context, addr sdk.AccAddress) types.StakeTier {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return types.StakeTier(-1)
	}
	return stake.Tier
}

// HasMinimumTier checks if user has at least the specified tier
func (k Keeper) HasMinimumTier(ctx sdk.Context, addr sdk.AccAddress, minTier types.StakeTier) bool {
	tier := k.GetUserTier(ctx, addr)
	return tier >= minTier
}

// CanLend checks if user can participate in lending
func (k Keeper) CanLend(ctx sdk.Context, addr sdk.AccAddress) bool {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return false
	}
	return stake.CanLend() && types.MeetsReputationRequirement(stake.ReputationScore, "lend")
}

// CanBorrow checks if user can borrow
func (k Keeper) CanBorrow(ctx sdk.Context, addr sdk.AccAddress) bool {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return false
	}
	return stake.CanBeBorrower() && types.MeetsReputationRequirement(stake.ReputationScore, "borrow")
}

// CanVerifyBusiness checks if user can verify a business of given size
func (k Keeper) CanVerifyBusiness(ctx sdk.Context, addr sdk.AccAddress, businessSize string) bool {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return false
	}

	if !stake.CanVerifyBusiness(businessSize) {
		return false
	}

	action := "verify_" + businessSize
	return types.MeetsReputationRequirement(stake.ReputationScore, action)
}

// =============================================================================
// ESCROW MODERATOR CHECKS
// =============================================================================

// CanModerate checks if user can be a moderator (requires Warden+ tier, 100K HODL)
func (k Keeper) CanModerate(ctx sdk.Context, addr sdk.AccAddress) bool {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return false
	}
	return stake.CanModerate() && types.MeetsReputationRequirement(stake.ReputationScore, "moderate")
}

// CanModerateLargeDisputes checks if user can handle large disputes (requires Steward+ tier, 1M HODL)
func (k Keeper) CanModerateLargeDisputes(ctx sdk.Context, addr sdk.AccAddress) bool {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return false
	}
	return stake.CanModerateLarge() && types.MeetsReputationRequirement(stake.ReputationScore, "moderate_large")
}

// CanSlashModerators checks if user can slash moderators (requires Archon+ tier, 10M HODL)
func (k Keeper) CanSlashModerators(ctx sdk.Context, addr sdk.AccAddress) bool {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return false
	}
	return stake.Tier >= types.TierArchon
}

// =============================================================================
// LENDING MODULE ADAPTER METHODS (return int instead of StakeTier)
// =============================================================================

// GetUserTierInt returns the user's tier as an int (for lending module interface)
func (k Keeper) GetUserTierInt(ctx sdk.Context, addr sdk.AccAddress) int {
	return int(k.GetUserTier(ctx, addr))
}

// GetStakeAge returns how long the user has been staked at their current tier (in seconds)
// Used for anti-Sybil checks - prevents banned users from creating new accounts to bypass bans
func (k Keeper) GetStakeAge(ctx sdk.Context, addr sdk.AccAddress) int64 {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return 0
	}

	// Calculate seconds since stake was created/last modified
	currentTime := ctx.BlockTime()
	duration := currentTime.Sub(stake.StakedAt)
	return int64(duration.Seconds())
}

// Note: GetReputation, RewardSuccessfulLoan, PenalizeLoanDefault, and SlashForLoanDefault
// are defined in reputation.go and rewards.go

// =============================================================================
// TOTAL CALCULATIONS
// =============================================================================

// GetTotalStaked returns the total HODL staked across all users
func (k Keeper) GetTotalStaked(ctx sdk.Context) math.Int {
	total := math.ZeroInt()
	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		total = total.Add(stake.StakedAmount)
		return false
	})
	return total
}

// GetTotalWeight returns the total reward weight across all users
func (k Keeper) GetTotalWeight(ctx sdk.Context) math.LegacyDec {
	total := math.LegacyZeroDec()
	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		total = total.Add(stake.CalculateWeight())
		return false
	})
	return total
}

// GetValidatorTotalWeight returns total weight of validators only
func (k Keeper) GetValidatorTotalWeight(ctx sdk.Context) math.LegacyDec {
	total := math.LegacyZeroDec()
	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		if stake.Tier == types.TierValidator && stake.IsValidator {
			total = total.Add(stake.CalculateWeight())
		}
		return false
	})
	return total
}

// =============================================================================
// EXTBRIDGE MODULE INTERFACE METHODS
// =============================================================================

// GetValidatorTier returns the staking tier for a validator as int32
func (k Keeper) GetValidatorTier(ctx sdk.Context, validator sdk.AccAddress) (int32, error) {
	stake, exists := k.GetUserStake(ctx, validator)
	if !exists {
		return -1, types.ErrStakeNotFound
	}
	return int32(stake.Tier), nil
}

// IsValidator checks if an address is a registered validator
func (k Keeper) IsValidator(ctx sdk.Context, addr sdk.AccAddress) bool {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return false
	}
	return stake.IsValidator && stake.Tier >= types.TierValidator
}

// GetTotalValidators returns the total number of active validators
func (k Keeper) GetTotalValidators(ctx sdk.Context) uint64 {
	count := uint64(0)
	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		if stake.IsValidator && stake.Tier >= types.TierValidator {
			count++
		}
		return false
	})
	return count
}

// GetValidatorsByTier returns validators at or above a certain tier
func (k Keeper) GetValidatorsByTier(ctx sdk.Context, minTier int32) ([]sdk.AccAddress, error) {
	var validators []sdk.AccAddress
	k.IterateUserStakes(ctx, func(stake types.UserStake) bool {
		if stake.IsValidator && int32(stake.Tier) >= minTier {
			addr, err := sdk.AccAddressFromBech32(stake.Owner)
			if err == nil {
				validators = append(validators, addr)
			}
		}
		return false
	})
	return validators, nil
}

// =============================================================================
// STAKE-AS-TRUST-CEILING: COMMITMENT MANAGEMENT
// =============================================================================
// Users cannot handle transactions above their staked amount.
// When a user creates an escrow, loan, or P2P trade, that value is "committed"
// from their stake. They cannot unstake below their total committed amount.

// GetUserCommitments returns a user's stake commitments
func (k Keeper) GetUserCommitments(ctx sdk.Context, owner sdk.AccAddress) (types.UserStakeCommitments, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetUserCommitmentsKey(owner)

	bz := store.Get(key)
	if bz == nil {
		return types.NewUserStakeCommitments(owner.String()), false
	}

	var commitments types.UserStakeCommitments
	if err := json.Unmarshal(bz, &commitments); err != nil {
		return types.NewUserStakeCommitments(owner.String()), false
	}
	return commitments, true
}

// SetUserCommitments stores a user's stake commitments
func (k Keeper) SetUserCommitments(ctx sdk.Context, commitments types.UserStakeCommitments) error {
	owner, err := sdk.AccAddressFromBech32(commitments.Owner)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetUserCommitmentsKey(owner)

	bz, err := json.Marshal(commitments)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// DeleteUserCommitments removes a user's commitments (when all cleared)
func (k Keeper) DeleteUserCommitments(ctx sdk.Context, owner sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetUserCommitmentsKey(owner)
	store.Delete(key)
}

// GetAvailableStake returns the stake amount available for new commitments
// Available = Staked - TotalCommitted
func (k Keeper) GetAvailableStake(ctx sdk.Context, owner sdk.AccAddress) math.Int {
	stake, exists := k.GetUserStake(ctx, owner)
	if !exists {
		return math.ZeroInt()
	}

	commitments, _ := k.GetUserCommitments(ctx, owner)
	available := stake.StakedAmount.Sub(commitments.TotalCommitted)

	if available.IsNegative() {
		return math.ZeroInt()
	}
	return available
}

// GetTotalCommitted returns the total amount committed by a user
func (k Keeper) GetTotalCommitted(ctx sdk.Context, owner sdk.AccAddress) math.Int {
	commitments, exists := k.GetUserCommitments(ctx, owner)
	if !exists {
		return math.ZeroInt()
	}
	return commitments.TotalCommitted
}

// CanCommit checks if user can commit the specified amount
// Returns true if stake - currentCommitments >= newAmount
func (k Keeper) CanCommit(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) bool {
	available := k.GetAvailableStake(ctx, owner)
	return available.GTE(amount)
}

// AddCommitment adds a new commitment to a user's stake
// This should be called when creating escrow, activating loan, starting P2P trade
func (k Keeper) AddCommitment(
	ctx sdk.Context,
	owner sdk.AccAddress,
	lockType types.LockType,
	referenceID string,
	amount math.Int,
	description string,
) error {
	// Verify this is a commitment-type lock
	if !lockType.IsCommitmentLock() {
		return types.ErrInvalidLockType
	}

	// Check if user can commit this amount
	if !k.CanCommit(ctx, owner, amount) {
		return types.ErrExceedsTrustCeiling
	}

	// Get or create commitments
	commitments, _ := k.GetUserCommitments(ctx, owner)
	if commitments.Owner == "" {
		commitments = types.NewUserStakeCommitments(owner.String())
	}

	// Create the commitment
	commitment := types.StakeCommitment{
		LockType:        lockType,
		ReferenceID:     referenceID,
		CommittedAmount: amount,
		Description:     description,
		CreatedAt:       ctx.BlockTime(),
	}

	// Add commitment
	commitments.AddCommitment(commitment)

	// Save
	if err := k.SetUserCommitments(ctx, commitments); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"commitment_added",
			sdk.NewAttribute("owner", owner.String()),
			sdk.NewAttribute("lock_type", lockType.String()),
			sdk.NewAttribute("reference_id", referenceID),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("total_committed", commitments.TotalCommitted.String()),
		),
	)

	return nil
}

// ReleaseCommitment removes a commitment when the activity is complete
// This should be called when escrow completes, loan repaid, P2P trade finished
func (k Keeper) ReleaseCommitment(
	ctx sdk.Context,
	owner sdk.AccAddress,
	lockType types.LockType,
	referenceID string,
) error {
	commitments, exists := k.GetUserCommitments(ctx, owner)
	if !exists {
		return nil // No commitments to release
	}

	// Get the commitment amount before removing (for event)
	commitment, found := commitments.GetCommitment(lockType, referenceID)
	if !found {
		return nil // Commitment doesn't exist
	}
	releasedAmount := commitment.CommittedAmount

	// Remove the commitment
	if !commitments.RemoveCommitment(lockType, referenceID) {
		return nil // Wasn't found
	}

	// Save or delete if empty
	if len(commitments.Commitments) == 0 {
		k.DeleteUserCommitments(ctx, owner)
	} else {
		if err := k.SetUserCommitments(ctx, commitments); err != nil {
			return err
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"commitment_released",
			sdk.NewAttribute("owner", owner.String()),
			sdk.NewAttribute("lock_type", lockType.String()),
			sdk.NewAttribute("reference_id", referenceID),
			sdk.NewAttribute("amount", releasedAmount.String()),
			sdk.NewAttribute("total_committed", commitments.TotalCommitted.String()),
		),
	)

	return nil
}

// GetCommitmentsByType returns all commitments of a specific type for a user
func (k Keeper) GetCommitmentsByType(ctx sdk.Context, owner sdk.AccAddress, lockType types.LockType) []types.StakeCommitment {
	commitments, exists := k.GetUserCommitments(ctx, owner)
	if !exists {
		return nil
	}
	return commitments.GetCommitmentsByType(lockType)
}

// CanUnstakeAmount checks if user can unstake the specified amount
// considering their current commitments
func (k Keeper) CanUnstakeAmount(ctx sdk.Context, owner sdk.AccAddress, unstakeAmount math.Int) (bool, string) {
	stake, exists := k.GetUserStake(ctx, owner)
	if !exists {
		return false, "no stake found"
	}

	commitments, _ := k.GetUserCommitments(ctx, owner)

	// After unstaking, remaining stake must cover commitments
	remainingAfterUnstake := stake.StakedAmount.Sub(unstakeAmount)
	if remainingAfterUnstake.LT(commitments.TotalCommitted) {
		return false, fmt.Sprintf("cannot unstake: %s uhodl committed to active positions (escrow: %s, lending: %s, p2p: %s)",
			commitments.TotalCommitted.String(),
			commitments.EscrowCommitted.String(),
			commitments.LendingCommitted.String(),
			commitments.P2PCommitted.String())
	}

	return true, ""
}

// =============================================================================
// ESCROW COMMITMENT WRAPPERS
// These are called by the escrow module for stake-as-trust-ceiling
// Only the SENDER (market maker who posts escrow) needs commitment
// =============================================================================

// AddEscrowCommitment adds a commitment when an escrow is funded by sender
func (k Keeper) AddEscrowCommitment(ctx sdk.Context, owner sdk.AccAddress, escrowID uint64, amount math.Int) error {
	referenceID := fmt.Sprintf("escrow:%d", escrowID)
	description := fmt.Sprintf("Escrow #%d commitment", escrowID)
	return k.AddCommitment(ctx, owner, types.LockTypeEscrowCommitment, referenceID, amount, description)
}

// ReleaseEscrowCommitment releases commitment when escrow is completed/refunded/cancelled
func (k Keeper) ReleaseEscrowCommitment(ctx sdk.Context, owner sdk.AccAddress, escrowID uint64) error {
	referenceID := fmt.Sprintf("escrow:%d", escrowID)
	return k.ReleaseCommitment(ctx, owner, types.LockTypeEscrowCommitment, referenceID)
}

// =============================================================================
// LENDING COMMITMENT WRAPPERS
// These are called by the lending module for stake-as-trust-ceiling
// Only the LENDER (who posts the offer) needs commitment
// Borrowers provide collateral instead of stake commitment
// =============================================================================

// AddLendingCommitment adds a commitment when a lender activates a loan
func (k Keeper) AddLendingCommitment(ctx sdk.Context, owner sdk.AccAddress, loanID uint64, amount math.Int) error {
	referenceID := fmt.Sprintf("loan:%d", loanID)
	description := fmt.Sprintf("Loan #%d lending commitment", loanID)
	return k.AddCommitment(ctx, owner, types.LockTypeLendingCommitment, referenceID, amount, description)
}

// ReleaseLendingCommitment releases commitment when loan is repaid/liquidated/defaulted
func (k Keeper) ReleaseLendingCommitment(ctx sdk.Context, owner sdk.AccAddress, loanID uint64) error {
	referenceID := fmt.Sprintf("loan:%d", loanID)
	return k.ReleaseCommitment(ctx, owner, types.LockTypeLendingCommitment, referenceID)
}

// =============================================================================
// P2P TRADE COMMITMENT WRAPPERS
// These are called by P2P trading module for stake-as-trust-ceiling
// Only the SELLER (who posts the ad) needs commitment
// Buyers who respond to ads are free to participate
// =============================================================================

// AddP2PCommitment adds a commitment when a seller posts a P2P trade ad
func (k Keeper) AddP2PCommitment(ctx sdk.Context, owner sdk.AccAddress, tradeID uint64, amount math.Int) error {
	referenceID := fmt.Sprintf("p2p:%d", tradeID)
	description := fmt.Sprintf("P2P Trade #%d commitment", tradeID)
	return k.AddCommitment(ctx, owner, types.LockTypeP2PCommitment, referenceID, amount, description)
}

// ReleaseP2PCommitment releases commitment when P2P trade is completed/cancelled
func (k Keeper) ReleaseP2PCommitment(ctx sdk.Context, owner sdk.AccAddress, tradeID uint64) error {
	referenceID := fmt.Sprintf("p2p:%d", tradeID)
	return k.ReleaseCommitment(ctx, owner, types.LockTypeP2PCommitment, referenceID)
}
