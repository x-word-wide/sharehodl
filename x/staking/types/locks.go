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

	// ==========================================================================
	// COMMITMENT-BASED LOCKS (Stake-as-Trust-Ceiling)
	// These lock types track amounts committed, not just boolean locks.
	// Users cannot unstake below their total committed amount.
	// ==========================================================================
	LockTypeEscrowCommitment   // Active escrow (as sender/recipient) - commits stake value
	LockTypeLendingCommitment  // Active lending position (lender/borrower) - commits loan value
	LockTypeP2PCommitment      // Active P2P trade - commits trade value
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
	case LockTypeEscrowCommitment:
		return "escrow_commitment"
	case LockTypeLendingCommitment:
		return "lending_commitment"
	case LockTypeP2PCommitment:
		return "p2p_commitment"
	default:
		return "unknown"
	}
}

// IsCommitmentLock returns true if this lock type tracks a committed amount
func (l LockType) IsCommitmentLock() bool {
	switch l {
	case LockTypeEscrowCommitment, LockTypeLendingCommitment, LockTypeP2PCommitment:
		return true
	default:
		return false
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

// =============================================================================
// STAKE COMMITMENTS (Stake-as-Trust-Ceiling)
// =============================================================================

// StakeCommitment represents a specific amount committed from a user's stake
// for an activity (escrow, loan, P2P trade). Users cannot unstake below
// their total committed amount.
type StakeCommitment struct {
	LockType        LockType  `json:"lock_type"`         // Type of commitment
	ReferenceID     string    `json:"reference_id"`      // Escrow ID, Loan ID, Trade ID
	CommittedAmount math.Int  `json:"committed_amount"`  // Amount committed in uhodl
	Description     string    `json:"description"`       // Human-readable description
	CreatedAt       time.Time `json:"created_at"`        // When commitment was created
	ExpiresAt       time.Time `json:"expires_at,omitempty"` // Optional expiration
}

// ProtoMessage implements proto.Message interface
func (m *StakeCommitment) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *StakeCommitment) Reset() { *m = StakeCommitment{} }

// String implements proto.Message interface
func (m *StakeCommitment) String() string { return m.ReferenceID }

// IsExpired checks if the commitment has expired
func (c StakeCommitment) IsExpired(currentTime time.Time) bool {
	if c.ExpiresAt.IsZero() {
		return false
	}
	return currentTime.After(c.ExpiresAt)
}

// UserStakeCommitments holds all commitments for a user
type UserStakeCommitments struct {
	Owner           string            `json:"owner"`
	Commitments     []StakeCommitment `json:"commitments"`
	TotalCommitted  math.Int          `json:"total_committed"`   // Sum of all active commitments
	EscrowCommitted math.Int          `json:"escrow_committed"`  // Sum of escrow commitments
	LendingCommitted math.Int         `json:"lending_committed"` // Sum of lending commitments
	P2PCommitted    math.Int          `json:"p2p_committed"`     // Sum of P2P commitments
}

// ProtoMessage implements proto.Message interface
func (m *UserStakeCommitments) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *UserStakeCommitments) Reset() { *m = UserStakeCommitments{} }

// String implements proto.Message interface
func (m *UserStakeCommitments) String() string { return m.Owner }

// NewUserStakeCommitments creates a new empty commitments struct
func NewUserStakeCommitments(owner string) UserStakeCommitments {
	return UserStakeCommitments{
		Owner:            owner,
		Commitments:      []StakeCommitment{},
		TotalCommitted:   math.ZeroInt(),
		EscrowCommitted:  math.ZeroInt(),
		LendingCommitted: math.ZeroInt(),
		P2PCommitted:     math.ZeroInt(),
	}
}

// AddCommitment adds a new commitment and updates totals
func (u *UserStakeCommitments) AddCommitment(commitment StakeCommitment) {
	u.Commitments = append(u.Commitments, commitment)
	u.TotalCommitted = u.TotalCommitted.Add(commitment.CommittedAmount)

	switch commitment.LockType {
	case LockTypeEscrowCommitment:
		u.EscrowCommitted = u.EscrowCommitted.Add(commitment.CommittedAmount)
	case LockTypeLendingCommitment:
		u.LendingCommitted = u.LendingCommitted.Add(commitment.CommittedAmount)
	case LockTypeP2PCommitment:
		u.P2PCommitted = u.P2PCommitted.Add(commitment.CommittedAmount)
	}
}

// RemoveCommitment removes a commitment by reference ID and updates totals
func (u *UserStakeCommitments) RemoveCommitment(lockType LockType, referenceID string) bool {
	for i, c := range u.Commitments {
		if c.LockType == lockType && c.ReferenceID == referenceID {
			// Remove from totals
			u.TotalCommitted = u.TotalCommitted.Sub(c.CommittedAmount)
			switch c.LockType {
			case LockTypeEscrowCommitment:
				u.EscrowCommitted = u.EscrowCommitted.Sub(c.CommittedAmount)
			case LockTypeLendingCommitment:
				u.LendingCommitted = u.LendingCommitted.Sub(c.CommittedAmount)
			case LockTypeP2PCommitment:
				u.P2PCommitted = u.P2PCommitted.Sub(c.CommittedAmount)
			}

			// Remove from slice
			u.Commitments = append(u.Commitments[:i], u.Commitments[i+1:]...)
			return true
		}
	}
	return false
}

// GetCommitment returns a specific commitment by reference
func (u UserStakeCommitments) GetCommitment(lockType LockType, referenceID string) (StakeCommitment, bool) {
	for _, c := range u.Commitments {
		if c.LockType == lockType && c.ReferenceID == referenceID {
			return c, true
		}
	}
	return StakeCommitment{}, false
}

// GetActiveCommitments returns all non-expired commitments
func (u UserStakeCommitments) GetActiveCommitments(currentTime time.Time) []StakeCommitment {
	var active []StakeCommitment
	for _, c := range u.Commitments {
		if !c.IsExpired(currentTime) {
			active = append(active, c)
		}
	}
	return active
}

// GetCommitmentsByType returns commitments of a specific type
func (u UserStakeCommitments) GetCommitmentsByType(lockType LockType) []StakeCommitment {
	var result []StakeCommitment
	for _, c := range u.Commitments {
		if c.LockType == lockType {
			result = append(result, c)
		}
	}
	return result
}

// RecalculateTotals recalculates all totals from commitments (for consistency)
func (u *UserStakeCommitments) RecalculateTotals() {
	u.TotalCommitted = math.ZeroInt()
	u.EscrowCommitted = math.ZeroInt()
	u.LendingCommitted = math.ZeroInt()
	u.P2PCommitted = math.ZeroInt()

	for _, c := range u.Commitments {
		u.TotalCommitted = u.TotalCommitted.Add(c.CommittedAmount)
		switch c.LockType {
		case LockTypeEscrowCommitment:
			u.EscrowCommitted = u.EscrowCommitted.Add(c.CommittedAmount)
		case LockTypeLendingCommitment:
			u.LendingCommitted = u.LendingCommitted.Add(c.CommittedAmount)
		case LockTypeP2PCommitment:
			u.P2PCommitted = u.P2PCommitted.Add(c.CommittedAmount)
		}
	}
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
