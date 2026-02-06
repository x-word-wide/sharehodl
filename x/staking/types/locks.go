package types

import (
	"time"

	"cosmossdk.io/math"
)

// LockType defines the type of stake lock
type LockType int32

const (
	LockTypeNone           LockType = iota
	LockTypeCompanyListing          // Active company listing
	LockTypePendingListing          // Listing under review
	LockTypeActiveLoan              // Active loan (lender or borrower)
	LockTypeActiveDispute           // Moderating or party in dispute
	LockTypePendingVote             // Participating in active governance vote
	LockTypeBan                     // User is banned/blacklisted
	LockTypeValidator               // Running a validator node
	LockTypeModerator               // Registered as escrow moderator
)

// String returns the lock type name
func (l LockType) String() string {
	switch l {
	case LockTypeNone:
		return "none"
	case LockTypeCompanyListing:
		return "company_listing"
	case LockTypePendingListing:
		return "pending_listing"
	case LockTypeActiveLoan:
		return "active_loan"
	case LockTypeActiveDispute:
		return "active_dispute"
	case LockTypePendingVote:
		return "pending_vote"
	case LockTypeBan:
		return "ban"
	case LockTypeValidator:
		return "validator"
	case LockTypeModerator:
		return "moderator"
	default:
		return "unknown"
	}
}

// StakeLock represents a single lock on a user's stake
type StakeLock struct {
	LockType    LockType  `json:"lock_type"`
	ReferenceID string    `json:"reference_id"` // Company ID, Loan ID, Dispute ID, etc.
	Description string    `json:"description"`
	LockedAt    time.Time `json:"locked_at"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"` // Optional expiration (for bans)
}

// ProtoMessage implements proto.Message interface
func (m *StakeLock) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *StakeLock) Reset() { *m = StakeLock{} }

// String implements proto.Message interface
func (m *StakeLock) String() string { return m.LockType.String() }

// IsExpired checks if the lock has expired
func (l StakeLock) IsExpired(currentTime time.Time) bool {
	if l.ExpiresAt.IsZero() {
		return false // No expiration
	}
	return currentTime.After(l.ExpiresAt)
}

// UserStakeLocks holds all locks for a user
type UserStakeLocks struct {
	Owner string      `json:"owner"`
	Locks []StakeLock `json:"locks"`
}

// ProtoMessage implements proto.Message interface
func (m *UserStakeLocks) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *UserStakeLocks) Reset() { *m = UserStakeLocks{} }

// String implements proto.Message interface
func (m *UserStakeLocks) String() string { return m.Owner }

// HasActiveLocks checks if user has any non-expired locks
func (u UserStakeLocks) HasActiveLocks(currentTime time.Time) bool {
	for _, lock := range u.Locks {
		if !lock.IsExpired(currentTime) {
			return true
		}
	}
	return false
}

// GetActiveLocks returns all non-expired locks
func (u UserStakeLocks) GetActiveLocks(currentTime time.Time) []StakeLock {
	var active []StakeLock
	for _, lock := range u.Locks {
		if !lock.IsExpired(currentTime) {
			active = append(active, lock)
		}
	}
	return active
}

// HasLockType checks if user has a specific type of lock
func (u UserStakeLocks) HasLockType(lockType LockType, currentTime time.Time) bool {
	for _, lock := range u.Locks {
		if lock.LockType == lockType && !lock.IsExpired(currentTime) {
			return true
		}
	}
	return false
}

// HasLockForReference checks if user has a lock for a specific reference
func (u UserStakeLocks) HasLockForReference(lockType LockType, referenceID string) bool {
	for _, lock := range u.Locks {
		if lock.LockType == lockType && lock.ReferenceID == referenceID {
			return true
		}
	}
	return false
}

// CountLocksByType returns count of locks by type
func (u UserStakeLocks) CountLocksByType(lockType LockType, currentTime time.Time) int {
	count := 0
	for _, lock := range u.Locks {
		if lock.LockType == lockType && !lock.IsExpired(currentTime) {
			count++
		}
	}
	return count
}

// UnbondingRequest represents a pending unstake request
type UnbondingRequest struct {
	Owner          string    `json:"owner"`
	Amount         math.Int  `json:"amount"`
	RequestedAt    time.Time `json:"requested_at"`
	CompletesAt    time.Time `json:"completes_at"`
	Completed      bool      `json:"completed"`
}

// ProtoMessage implements proto.Message interface
func (m *UnbondingRequest) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *UnbondingRequest) Reset() { *m = UnbondingRequest{} }

// String implements proto.Message interface
func (m *UnbondingRequest) String() string { return m.Owner }

// IsComplete checks if unbonding period is complete
func (u UnbondingRequest) IsComplete(currentTime time.Time) bool {
	return currentTime.After(u.CompletesAt) || u.Completed
}

// Default unbonding period (14 days)
const (
	DefaultUnbondingPeriodDays    = 14
	DefaultUnbondingPeriodSeconds = 14 * 24 * 60 * 60 // 14 days in seconds
	// Note: DefaultUnbondingPeriodBlocks is defined in params.go
)

// Minimum tier requirements for various actions
const (
	MinTierForListing       = TierKeeper  // 10K HODL to submit company listing
	MinTierForSmallBusiness = TierKeeper  // 10K HODL to verify small business
	MinTierForMediumBusiness = TierWarden // 100K HODL to verify medium business
	MinTierForLargeBusiness = TierSteward // 1M HODL to verify large business
	MinTierForEnterprise    = TierArchon  // 10M HODL to verify enterprise
)

// GetMinTierForBusinessSize returns the minimum tier required to verify a business
func GetMinTierForBusinessSize(totalShares int64) StakeTier {
	// Small: < 10M shares
	// Medium: 10M - 100M shares
	// Large: 100M - 1B shares
	// Enterprise: > 1B shares
	switch {
	case totalShares >= 1_000_000_000:
		return MinTierForEnterprise
	case totalShares >= 100_000_000:
		return MinTierForLargeBusiness
	case totalShares >= 10_000_000:
		return MinTierForMediumBusiness
	default:
		return MinTierForSmallBusiness
	}
}
