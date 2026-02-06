package keeper

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/staking/types"
)

// =============================================================================
// STAKE LOCK MANAGEMENT
// =============================================================================

// GetUserLocks returns all locks for a user
func (k Keeper) GetUserLocks(ctx sdk.Context, addr sdk.AccAddress) types.UserStakeLocks {
	store := ctx.KVStore(k.storeKey)
	key := types.GetUserLocksKey(addr)

	bz := store.Get(key)
	if bz == nil {
		return types.UserStakeLocks{
			Owner: addr.String(),
			Locks: []types.StakeLock{},
		}
	}

	var locks types.UserStakeLocks
	if err := json.Unmarshal(bz, &locks); err != nil {
		return types.UserStakeLocks{Owner: addr.String()}
	}
	return locks
}

// SetUserLocks stores a user's locks
func (k Keeper) SetUserLocks(ctx sdk.Context, locks types.UserStakeLocks) error {
	addr, err := sdk.AccAddressFromBech32(locks.Owner)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetUserLocksKey(addr)

	bz, err := json.Marshal(locks)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// AddLock adds a lock to a user's stake
func (k Keeper) AddLock(ctx sdk.Context, addr sdk.AccAddress, lockType types.LockType, referenceID, description string) error {
	locks := k.GetUserLocks(ctx, addr)

	// Check if lock already exists for this reference
	if locks.HasLockForReference(lockType, referenceID) {
		return nil // Already locked
	}

	newLock := types.StakeLock{
		LockType:    lockType,
		ReferenceID: referenceID,
		Description: description,
		LockedAt:    ctx.BlockTime(),
	}

	locks.Locks = append(locks.Locks, newLock)

	if err := k.SetUserLocks(ctx, locks); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStakeLocked,
			sdk.NewAttribute(types.AttributeKeyStaker, addr.String()),
			sdk.NewAttribute(types.AttributeKeyLockType, lockType.String()),
			sdk.NewAttribute(types.AttributeKeyReferenceID, referenceID),
		),
	)

	k.Logger(ctx).Info("stake lock added",
		"staker", addr.String(),
		"lock_type", lockType.String(),
		"reference", referenceID,
	)

	return nil
}

// AddLockWithExpiry adds a lock with an expiration time (for bans)
func (k Keeper) AddLockWithExpiry(ctx sdk.Context, addr sdk.AccAddress, lockType types.LockType, referenceID, description string, expiresAt time.Time) error {
	locks := k.GetUserLocks(ctx, addr)

	// Check if lock already exists for this reference
	if locks.HasLockForReference(lockType, referenceID) {
		return nil // Already locked
	}

	newLock := types.StakeLock{
		LockType:    lockType,
		ReferenceID: referenceID,
		Description: description,
		LockedAt:    ctx.BlockTime(),
		ExpiresAt:   expiresAt,
	}

	locks.Locks = append(locks.Locks, newLock)

	if err := k.SetUserLocks(ctx, locks); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStakeLocked,
			sdk.NewAttribute(types.AttributeKeyStaker, addr.String()),
			sdk.NewAttribute(types.AttributeKeyLockType, lockType.String()),
			sdk.NewAttribute(types.AttributeKeyReferenceID, referenceID),
			sdk.NewAttribute(types.AttributeKeyExpiresAt, expiresAt.String()),
		),
	)

	return nil
}

// RemoveLock removes a specific lock from a user's stake
func (k Keeper) RemoveLock(ctx sdk.Context, addr sdk.AccAddress, lockType types.LockType, referenceID string) error {
	locks := k.GetUserLocks(ctx, addr)

	// Find and remove the lock
	found := false
	var newLocks []types.StakeLock
	for _, lock := range locks.Locks {
		if lock.LockType == lockType && lock.ReferenceID == referenceID {
			found = true
			continue // Skip this lock (remove it)
		}
		newLocks = append(newLocks, lock)
	}

	if !found {
		return types.ErrLockNotFound
	}

	locks.Locks = newLocks

	if err := k.SetUserLocks(ctx, locks); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStakeUnlocked,
			sdk.NewAttribute(types.AttributeKeyStaker, addr.String()),
			sdk.NewAttribute(types.AttributeKeyLockType, lockType.String()),
			sdk.NewAttribute(types.AttributeKeyReferenceID, referenceID),
		),
	)

	k.Logger(ctx).Info("stake lock removed",
		"staker", addr.String(),
		"lock_type", lockType.String(),
		"reference", referenceID,
	)

	return nil
}

// RemoveAllLocksOfType removes all locks of a specific type
func (k Keeper) RemoveAllLocksOfType(ctx sdk.Context, addr sdk.AccAddress, lockType types.LockType) error {
	locks := k.GetUserLocks(ctx, addr)

	var newLocks []types.StakeLock
	for _, lock := range locks.Locks {
		if lock.LockType != lockType {
			newLocks = append(newLocks, lock)
		}
	}

	locks.Locks = newLocks
	return k.SetUserLocks(ctx, locks)
}

// CleanExpiredLocks removes all expired locks
func (k Keeper) CleanExpiredLocks(ctx sdk.Context, addr sdk.AccAddress) {
	locks := k.GetUserLocks(ctx, addr)
	currentTime := ctx.BlockTime()

	var activeLocks []types.StakeLock
	for _, lock := range locks.Locks {
		if !lock.IsExpired(currentTime) {
			activeLocks = append(activeLocks, lock)
		}
	}

	if len(activeLocks) != len(locks.Locks) {
		locks.Locks = activeLocks
		k.SetUserLocks(ctx, locks)
	}
}

// HasActiveLocks checks if a user has any active (non-expired) locks
func (k Keeper) HasActiveLocks(ctx sdk.Context, addr sdk.AccAddress) bool {
	locks := k.GetUserLocks(ctx, addr)
	return locks.HasActiveLocks(ctx.BlockTime())
}

// GetActiveLocks returns all active locks for a user
func (k Keeper) GetActiveLocks(ctx sdk.Context, addr sdk.AccAddress) []types.StakeLock {
	locks := k.GetUserLocks(ctx, addr)
	return locks.GetActiveLocks(ctx.BlockTime())
}

// CanUnstake checks if a user can unstake (no active locks)
func (k Keeper) CanUnstake(ctx sdk.Context, addr sdk.AccAddress) (bool, []types.StakeLock) {
	// Clean expired locks first
	k.CleanExpiredLocks(ctx, addr)

	activeLocks := k.GetActiveLocks(ctx, addr)
	return len(activeLocks) == 0, activeLocks
}

// =============================================================================
// UNBONDING REQUEST MANAGEMENT
// =============================================================================

// GetUnbondingRequest returns a user's unbonding request
func (k Keeper) GetUnbondingRequest(ctx sdk.Context, addr sdk.AccAddress) (types.UnbondingRequest, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetUnbondingRequestKey(addr)

	bz := store.Get(key)
	if bz == nil {
		return types.UnbondingRequest{}, false
	}

	var request types.UnbondingRequest
	if err := json.Unmarshal(bz, &request); err != nil {
		return types.UnbondingRequest{}, false
	}
	return request, true
}

// SetUnbondingRequest stores an unbonding request
func (k Keeper) SetUnbondingRequest(ctx sdk.Context, request types.UnbondingRequest) error {
	addr, err := sdk.AccAddressFromBech32(request.Owner)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetUnbondingRequestKey(addr)

	bz, err := json.Marshal(request)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// DeleteUnbondingRequest removes an unbonding request
func (k Keeper) DeleteUnbondingRequest(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetUnbondingRequestKey(addr)
	store.Delete(key)
}

// CreateUnbondingRequest initiates the unbonding period (governance-controllable)
func (k Keeper) CreateUnbondingRequest(ctx sdk.Context, addr sdk.AccAddress, amount math.Int) error {
	// Check for existing unbonding request
	if _, exists := k.GetUnbondingRequest(ctx, addr); exists {
		return types.ErrUnbondingInProgress
	}

	// Get unbonding period from governance-controllable params
	params := k.GetParams(ctx)
	unbondingSeconds := params.UnbondingPeriodBlocks * 6 // 6 seconds per block
	completesAt := ctx.BlockTime().Add(time.Duration(unbondingSeconds) * time.Second)

	request := types.UnbondingRequest{
		Owner:       addr.String(),
		Amount:      amount,
		RequestedAt: ctx.BlockTime(),
		CompletesAt: completesAt,
		Completed:   false,
	}

	if err := k.SetUnbondingRequest(ctx, request); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbondingStarted,
			sdk.NewAttribute(types.AttributeKeyStaker, addr.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletesAt, completesAt.String()),
		),
	)

	k.Logger(ctx).Info("unbonding started",
		"staker", addr.String(),
		"amount", amount.String(),
		"completes_at", completesAt.String(),
	)

	return nil
}

// CompleteUnbonding completes an unbonding request if the period has passed
func (k Keeper) CompleteUnbonding(ctx sdk.Context, addr sdk.AccAddress) (math.Int, error) {
	request, exists := k.GetUnbondingRequest(ctx, addr)
	if !exists {
		return math.ZeroInt(), types.ErrNoUnbondingInProgress
	}

	if request.Completed {
		return math.ZeroInt(), types.ErrNoUnbondingInProgress
	}

	if !request.IsComplete(ctx.BlockTime()) {
		return math.ZeroInt(), types.ErrUnbondingNotComplete
	}

	// Transfer tokens from staking pool to user
	coins := sdk.NewCoins(sdk.NewCoin("uhodl", request.Amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.StakingPoolName, addr, coins); err != nil {
		return math.ZeroInt(), err
	}

	// Mark as completed and delete
	k.DeleteUnbondingRequest(ctx, addr)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbondingCompleted,
			sdk.NewAttribute(types.AttributeKeyStaker, addr.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, request.Amount.String()),
		),
	)

	k.Logger(ctx).Info("unbonding completed",
		"staker", addr.String(),
		"amount", request.Amount.String(),
	)

	return request.Amount, nil
}

// CancelUnbonding cancels an unbonding request (re-stakes the tokens)
func (k Keeper) CancelUnbonding(ctx sdk.Context, addr sdk.AccAddress) error {
	request, exists := k.GetUnbondingRequest(ctx, addr)
	if !exists {
		return types.ErrNoUnbondingInProgress
	}

	if request.Completed {
		return types.ErrNoUnbondingInProgress
	}

	// Delete the request (tokens remain staked)
	k.DeleteUnbondingRequest(ctx, addr)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnbondingCancelled,
			sdk.NewAttribute(types.AttributeKeyStaker, addr.String()),
			sdk.NewAttribute(types.AttributeKeyAmount, request.Amount.String()),
		),
	)

	k.Logger(ctx).Info("unbonding cancelled",
		"staker", addr.String(),
		"amount", request.Amount.String(),
	)

	return nil
}

// =============================================================================
// LISTING TIER CHECKS
// =============================================================================

// CanSubmitListing checks if a user can submit a company listing
// Requires Keeper+ tier (10K HODL)
func (k Keeper) CanSubmitListing(ctx sdk.Context, addr sdk.AccAddress) bool {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return false
	}
	return stake.Tier >= types.MinTierForListing
}

// CanVerifyBusiness checks if a user can verify a business of given size
func (k Keeper) CanVerifyBusinessByShares(ctx sdk.Context, addr sdk.AccAddress, totalShares int64) bool {
	stake, exists := k.GetUserStake(ctx, addr)
	if !exists {
		return false
	}

	requiredTier := types.GetMinTierForBusinessSize(totalShares)
	return stake.Tier >= requiredTier
}

// GetVerifierTierForBusiness returns the minimum tier required to verify a business
func (k Keeper) GetVerifierTierForBusiness(totalShares int64) types.StakeTier {
	return types.GetMinTierForBusinessSize(totalShares)
}

// =============================================================================
// COMPANY LISTING LOCK HELPERS
// =============================================================================

// LockForCompanyListing adds a lock when a company is listed
func (k Keeper) LockForCompanyListing(ctx sdk.Context, addr sdk.AccAddress, companyID string) error {
	return k.AddLock(ctx, addr, types.LockTypeCompanyListing, companyID, "Active company listing: "+companyID)
}

// LockForPendingListing adds a lock when a listing is submitted
func (k Keeper) LockForPendingListing(ctx sdk.Context, addr sdk.AccAddress, listingID string) error {
	return k.AddLock(ctx, addr, types.LockTypePendingListing, listingID, "Pending listing under review: "+listingID)
}

// UnlockPendingListing removes the pending listing lock (approved or rejected)
func (k Keeper) UnlockPendingListing(ctx sdk.Context, addr sdk.AccAddress, listingID string) error {
	return k.RemoveLock(ctx, addr, types.LockTypePendingListing, listingID)
}

// UnlockCompanyListing removes the company listing lock (company delisted)
func (k Keeper) UnlockCompanyListing(ctx sdk.Context, addr sdk.AccAddress, companyID string) error {
	return k.RemoveLock(ctx, addr, types.LockTypeCompanyListing, companyID)
}

// =============================================================================
// OTHER MODULE LOCK HELPERS
// =============================================================================

// LockForLoan adds a lock for an active loan
func (k Keeper) LockForLoan(ctx sdk.Context, addr sdk.AccAddress, loanID string) error {
	return k.AddLock(ctx, addr, types.LockTypeActiveLoan, loanID, "Active loan: "+loanID)
}

// UnlockLoan removes a loan lock
func (k Keeper) UnlockLoan(ctx sdk.Context, addr sdk.AccAddress, loanID string) error {
	return k.RemoveLock(ctx, addr, types.LockTypeActiveLoan, loanID)
}

// LockForDispute adds a lock for an active dispute
func (k Keeper) LockForDispute(ctx sdk.Context, addr sdk.AccAddress, disputeID string) error {
	return k.AddLock(ctx, addr, types.LockTypeActiveDispute, disputeID, "Active dispute: "+disputeID)
}

// UnlockDispute removes a dispute lock
func (k Keeper) UnlockDispute(ctx sdk.Context, addr sdk.AccAddress, disputeID string) error {
	return k.RemoveLock(ctx, addr, types.LockTypeActiveDispute, disputeID)
}

// LockForVote adds a lock for a pending governance vote
func (k Keeper) LockForVote(ctx sdk.Context, addr sdk.AccAddress, proposalID string) error {
	return k.AddLock(ctx, addr, types.LockTypePendingVote, proposalID, "Pending vote on proposal: "+proposalID)
}

// UnlockVote removes a vote lock
func (k Keeper) UnlockVote(ctx sdk.Context, addr sdk.AccAddress, proposalID string) error {
	return k.RemoveLock(ctx, addr, types.LockTypePendingVote, proposalID)
}

// LockForBan adds a ban lock with optional expiry
func (k Keeper) LockForBan(ctx sdk.Context, addr sdk.AccAddress, reason string, permanent bool, duration time.Duration) error {
	var expiresAt time.Time
	if !permanent {
		expiresAt = ctx.BlockTime().Add(duration)
	}
	return k.AddLockWithExpiry(ctx, addr, types.LockTypeBan, "ban", "User banned: "+reason, expiresAt)
}

// UnlockBan removes a ban lock
func (k Keeper) UnlockBan(ctx sdk.Context, addr sdk.AccAddress) error {
	return k.RemoveLock(ctx, addr, types.LockTypeBan, "ban")
}

// LockForValidator adds a validator lock
func (k Keeper) LockForValidator(ctx sdk.Context, addr sdk.AccAddress, valAddr string) error {
	return k.AddLock(ctx, addr, types.LockTypeValidator, valAddr, "Running validator node: "+valAddr)
}

// UnlockValidator removes a validator lock
func (k Keeper) UnlockValidator(ctx sdk.Context, addr sdk.AccAddress, valAddr string) error {
	return k.RemoveLock(ctx, addr, types.LockTypeValidator, valAddr)
}

// LockForModerator adds a moderator lock
func (k Keeper) LockForModerator(ctx sdk.Context, addr sdk.AccAddress) error {
	return k.AddLock(ctx, addr, types.LockTypeModerator, "moderator", "Registered as escrow moderator")
}

// UnlockModerator removes a moderator lock
func (k Keeper) UnlockModerator(ctx sdk.Context, addr sdk.AccAddress) error {
	return k.RemoveLock(ctx, addr, types.LockTypeModerator, "moderator")
}
