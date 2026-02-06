package types

import (
	"time"

	"cosmossdk.io/math"
)

// BusinessVerificationStatus represents the status of business verification
type BusinessVerificationStatus int32

const (
	VerificationPending    BusinessVerificationStatus = iota
	VerificationInProgress
	VerificationApproved
	VerificationRejected
	VerificationDisputed
)

// String returns the string representation of verification status
func (s BusinessVerificationStatus) String() string {
	switch s {
	case VerificationPending:
		return "Pending"
	case VerificationInProgress:
		return "In Progress"
	case VerificationApproved:
		return "Approved"
	case VerificationRejected:
		return "Rejected"
	case VerificationDisputed:
		return "Disputed"
	default:
		return "Unknown"
	}
}

// BusinessVerification represents a business verification request
// Migrated from x/validator module - now uses x/staking tiers
type BusinessVerification struct {
	BusinessID           string                     `json:"business_id"`
	CompanyName          string                     `json:"company_name"`
	BusinessType         string                     `json:"business_type"`
	RegistrationCountry  string                     `json:"registration_country"`
	DocumentHashes       []string                   `json:"document_hashes"`
	RequiredTier         int                        `json:"required_tier"` // Uses staking tiers (Keeper=1, Warden=2, Steward=3, Archon=4)
	StakeRequired        math.Int                   `json:"stake_required"`
	VerificationFee      math.Int                   `json:"verification_fee"`
	Status               BusinessVerificationStatus `json:"status"`
	AssignedVerifiers    []string                   `json:"assigned_verifiers"`
	VerificationDeadline time.Time                  `json:"verification_deadline"`
	SubmittedAt          time.Time                  `json:"submitted_at"`
	CompletedAt          time.Time                  `json:"completed_at"`
	VerificationNotes    string                     `json:"verification_notes"`
	DisputeReason        string                     `json:"dispute_reason"`
}

// ProtoMessage implements proto.Message interface
func (m *BusinessVerification) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *BusinessVerification) Reset() { *m = BusinessVerification{} }

// String implements proto.Message interface
func (m *BusinessVerification) String() string { return m.BusinessID }

// VerifierStake represents a verifier's stake in a verified business
type VerifierStake struct {
	VerifierAddress  string         `json:"verifier_address"`
	BusinessID       string         `json:"business_id"`
	StakeAmount      math.Int       `json:"stake_amount"`
	EquityPercentage math.LegacyDec `json:"equity_percentage"`
	StakedAt         time.Time      `json:"staked_at"`
	VestingPeriod    time.Duration  `json:"vesting_period"`
	IsVested         bool           `json:"is_vested"`
}

// ProtoMessage implements proto.Message interface
func (m *VerifierStake) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *VerifierStake) Reset() { *m = VerifierStake{} }

// String implements proto.Message interface
func (m *VerifierStake) String() string { return m.VerifierAddress + ":" + m.BusinessID }

// VerificationRewards tracks rewards earned by verifiers
type VerificationRewards struct {
	VerifierAddress        string    `json:"verifier_address"`
	TotalHODLRewards       math.Int  `json:"total_hodl_rewards"`
	TotalEquityRewards     math.Int  `json:"total_equity_rewards"`
	VerificationsCompleted uint64    `json:"verifications_completed"`
	LastRewardClaim        time.Time `json:"last_reward_claim"`
}

// ProtoMessage implements proto.Message interface
func (m *VerificationRewards) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *VerificationRewards) Reset() { *m = VerificationRewards{} }

// String implements proto.Message interface
func (m *VerificationRewards) String() string { return m.VerifierAddress }

// VerificationParams defines parameters for business verification
type VerificationParams struct {
	MinVerificationDelay    time.Duration  `json:"min_verification_delay"`
	MaxVerificationDelay    time.Duration  `json:"max_verification_delay"`
	VerificationTimeoutDays uint64         `json:"verification_timeout_days"`
	MinVerifiersRequired    uint64         `json:"min_verifiers_required"`
	BaseVerificationReward  math.Int       `json:"base_verification_reward"`
	EquityStakePercentage   math.LegacyDec `json:"equity_stake_percentage"`
}

// DefaultVerificationParams returns default verification parameters
func DefaultVerificationParams() VerificationParams {
	return VerificationParams{
		MinVerificationDelay:    24 * time.Hour,
		MaxVerificationDelay:    7 * 24 * time.Hour,
		VerificationTimeoutDays: 30,
		MinVerifiersRequired:    3,
		BaseVerificationReward:  math.NewInt(1000000000), // 1000 HODL
		EquityStakePercentage:   math.LegacyNewDecWithPrec(1, 2), // 0.01 = 1%
	}
}

// Tier mapping from old validator tiers to staking tiers
// Old: Bronze=0, Silver=1, Gold=2, Platinum=3
// New: Holder=0, Keeper=1, Warden=2, Steward=3, Archon=4, Validator=5
// Mapping: Bronze→Keeper(1), Silver→Warden(2), Gold→Steward(3), Platinum→Archon(4)

// GetRequiredStakingTier maps business verification requirements to staking tiers
// businessSize: "small" → Keeper (1), "medium" → Warden (2), "large" → Steward (3), "enterprise" → Archon (4)
func GetRequiredStakingTier(businessSize string) int {
	switch businessSize {
	case "small":
		return StakeTierKeeper // 10K HODL
	case "medium":
		return StakeTierWarden // 100K HODL
	case "large":
		return StakeTierSteward // 1M HODL
	case "enterprise":
		return StakeTierArchon // 10M HODL
	default:
		return StakeTierKeeper
	}
}

// GetBusinessSizeFromTotalShares determines business size based on total shares
func GetBusinessSizeFromTotalShares(totalShares int64) string {
	if totalShares >= 10000000 { // 10M+ shares
		return "enterprise"
	} else if totalShares >= 1000000 { // 1M+ shares
		return "large"
	} else if totalShares >= 100000 { // 100K+ shares
		return "medium"
	}
	return "small"
}

// Key prefixes for verification storage
var (
	BusinessVerificationPrefix = []byte{0x30}
	VerifierStakePrefix        = []byte{0x31}
	VerificationRewardsPrefix  = []byte{0x32}
	VerificationParamsKey      = []byte{0x33}
)

// GetBusinessVerificationKey returns the store key for a business verification
func GetBusinessVerificationKey(businessID string) []byte {
	return append(BusinessVerificationPrefix, []byte(businessID)...)
}

// GetVerifierStakeKey returns the store key for a verifier stake
func GetVerifierStakeKey(verifierAddr, businessID string) []byte {
	key := append(VerifierStakePrefix, []byte(verifierAddr)...)
	key = append(key, []byte("|")...)
	return append(key, []byte(businessID)...)
}

// GetVerificationRewardsKey returns the store key for verification rewards
func GetVerificationRewardsKey(verifierAddr string) []byte {
	return append(VerificationRewardsPrefix, []byte(verifierAddr)...)
}
