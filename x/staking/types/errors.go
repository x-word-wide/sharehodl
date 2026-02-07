package types

import (
	"cosmossdk.io/errors"
)

// x/staking module sentinel errors
var (
	// Staking errors
	ErrInsufficientStake      = errors.Register(ModuleName, 2, "insufficient stake for tier")
	ErrStakeNotFound          = errors.Register(ModuleName, 3, "stake not found")
	ErrAlreadyStaked          = errors.Register(ModuleName, 4, "user already has stake")
	ErrInvalidAmount          = errors.Register(ModuleName, 5, "invalid amount")
	ErrBelowMinimumStake      = errors.Register(ModuleName, 6, "amount below minimum stake")
	ErrUnbondingInProgress    = errors.Register(ModuleName, 7, "unbonding already in progress")
	ErrNoUnbondingInProgress  = errors.Register(ModuleName, 8, "no unbonding in progress")
	ErrUnbondingNotComplete   = errors.Register(ModuleName, 9, "unbonding period not complete")

	// Tier errors
	ErrInsufficientTier       = errors.Register(ModuleName, 10, "insufficient tier for action")
	ErrTierDemotion           = errors.Register(ModuleName, 11, "stake would cause tier demotion")

	// Reputation errors
	ErrInsufficientReputation = errors.Register(ModuleName, 20, "insufficient reputation for action")
	ErrReputationTooLow       = errors.Register(ModuleName, 21, "reputation too low")
	ErrReputationLocked       = errors.Register(ModuleName, 22, "reputation changes locked")

	// Validator errors
	ErrNotValidator           = errors.Register(ModuleName, 30, "user is not a validator")
	ErrAlreadyValidator       = errors.Register(ModuleName, 31, "user is already a validator")
	ErrValidatorNotActive     = errors.Register(ModuleName, 32, "validator is not active")

	// Reward errors
	ErrNoRewardsToClaim       = errors.Register(ModuleName, 40, "no rewards to claim")
	ErrRewardsLocked          = errors.Register(ModuleName, 41, "rewards are locked")
	ErrInsufficientTreasury   = errors.Register(ModuleName, 42, "insufficient treasury balance")

	// Lending permission errors
	ErrCannotLend             = errors.Register(ModuleName, 50, "user cannot participate in lending")
	ErrCannotBorrow           = errors.Register(ModuleName, 51, "user cannot borrow")

	// Authorization errors
	ErrUnauthorized           = errors.Register(ModuleName, 60, "unauthorized")
	ErrInvalidAuthority       = errors.Register(ModuleName, 61, "invalid authority")

	// Stake lock errors
	ErrStakeLocked            = errors.Register(ModuleName, 70, "stake is locked and cannot be unstaked")
	ErrActiveCompanyListing   = errors.Register(ModuleName, 71, "stake locked: active company listing")
	ErrPendingListing         = errors.Register(ModuleName, 72, "stake locked: pending listing under review")
	ErrActiveLoan             = errors.Register(ModuleName, 73, "stake locked: active loan")
	ErrActiveDispute          = errors.Register(ModuleName, 74, "stake locked: active dispute")
	ErrPendingVote            = errors.Register(ModuleName, 75, "stake locked: pending governance vote")
	ErrUserBanned             = errors.Register(ModuleName, 76, "stake locked: user is banned")
	ErrActiveValidator        = errors.Register(ModuleName, 77, "stake locked: running validator node")
	ErrActiveModerator        = errors.Register(ModuleName, 78, "stake locked: registered as moderator")
	ErrLockNotFound           = errors.Register(ModuleName, 79, "lock not found")

	// Listing requirement errors
	ErrInsufficientTierForListing = errors.Register(ModuleName, 80, "must be Keeper tier or higher (10K HODL) to submit company listing")
	ErrInsufficientTierToVerify   = errors.Register(ModuleName, 81, "insufficient tier to verify this business size")

	// Stake-as-Trust-Ceiling commitment errors
	ErrInvalidLockType       = errors.Register(ModuleName, 90, "invalid lock type for commitment")
	ErrExceedsTrustCeiling   = errors.Register(ModuleName, 91, "amount exceeds available stake (trust ceiling)")
	ErrCommitmentNotFound    = errors.Register(ModuleName, 92, "commitment not found")
	ErrInsufficientAvailable = errors.Register(ModuleName, 93, "insufficient available stake after commitments")
)

// Event types
const (
	EventTypeStake              = "stake"
	EventTypeUnstake            = "unstake"
	EventTypeClaimRewards       = "claim_rewards"
	EventTypeTierChange         = "tier_change"
	EventTypeReputationChange   = "reputation_change"
	EventTypeRewardsDistributed = "rewards_distributed"
	EventTypeSlash              = "slash"
	EventTypeValidatorRegistered = "validator_registered"
	EventTypeValidatorRemoved   = "validator_removed"
	EventTypeParamsUpdated      = "params_updated"
	EventTypeStakeLocked        = "stake_locked"
	EventTypeStakeUnlocked      = "stake_unlocked"
	EventTypeUnbondingStarted   = "unbonding_started"
	EventTypeUnbondingCompleted = "unbonding_completed"
	EventTypeUnbondingCancelled = "unbonding_cancelled"

	AttributeKeyStaker          = "staker"
	AttributeKeyAuthority       = "authority"
	AttributeKeyLockType        = "lock_type"
	AttributeKeyReferenceID     = "reference_id"
	AttributeKeyExpiresAt       = "expires_at"
	AttributeKeyCompletesAt     = "completes_at"
	AttributeKeyAmount          = "amount"
	AttributeKeyTier            = "tier"
	AttributeKeyOldTier         = "old_tier"
	AttributeKeyNewTier         = "new_tier"
	AttributeKeyReputation      = "reputation"
	AttributeKeyReputationChange = "reputation_change"
	AttributeKeyAction          = "action"
	AttributeKeyReason          = "reason"
	AttributeKeyRewards         = "rewards"
	AttributeKeyEpoch           = "epoch"
	AttributeKeyTotalDistributed = "total_distributed"
	AttributeKeyStakerPool      = "staker_pool"
	AttributeKeyValidatorPool   = "validator_pool"
	AttributeKeyGovernancePool  = "governance_pool"
	AttributeKeySlashAmount     = "slash_amount"
	AttributeKeySlashReason     = "slash_reason"
)
