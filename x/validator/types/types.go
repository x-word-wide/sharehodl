package types

import (
	"context"
	"time"

	"cosmossdk.io/math"
)

// ValidatorTier represents the different validator tiers
type ValidatorTier int32

const (
	TierBronze ValidatorTier = iota
	TierSilver
	TierGold
	TierPlatinum
)

// String returns the string representation of the tier
func (t ValidatorTier) String() string {
	switch t {
	case TierBronze:
		return "Bronze"
	case TierSilver:
		return "Silver"
	case TierGold:
		return "Gold"
	case TierPlatinum:
		return "Platinum"
	default:
		return "Unknown"
	}
}

// GetStakeThreshold returns the minimum stake required for each tier
func (t ValidatorTier) GetStakeThreshold(isTestnet bool) math.Int {
	if isTestnet {
		switch t {
		case TierBronze:
			return math.NewInt(TestnetBronzeTierThreshold)
		case TierSilver:
			return math.NewInt(TestnetSilverTierThreshold)
		case TierGold:
			return math.NewInt(TestnetGoldTierThreshold)
		case TierPlatinum:
			return math.NewInt(TestnetPlatinumTierThreshold)
		default:
			return math.ZeroInt()
		}
	}
	
	// Mainnet thresholds
	switch t {
	case TierBronze:
		return math.NewInt(BronzeTierThreshold)
	case TierSilver:
		return math.NewInt(SilverTierThreshold)
	case TierGold:
		return math.NewInt(GoldTierThreshold)
	case TierPlatinum:
		return math.NewInt(PlatinumTierThreshold)
	default:
		return math.ZeroInt()
	}
}

// GetTierFromStake determines the appropriate tier based on staked amount
func GetTierFromStake(stakeAmount math.Int, isTestnet bool) ValidatorTier {
	stake := stakeAmount.Int64()
	
	if isTestnet {
		if stake >= TestnetPlatinumTierThreshold {
			return TierPlatinum
		} else if stake >= TestnetGoldTierThreshold {
			return TierGold
		} else if stake >= TestnetSilverTierThreshold {
			return TierSilver
		} else if stake >= TestnetBronzeTierThreshold {
			return TierBronze
		}
	} else {
		// Mainnet thresholds
		if stake >= PlatinumTierThreshold {
			return TierPlatinum
		} else if stake >= GoldTierThreshold {
			return TierGold
		} else if stake >= SilverTierThreshold {
			return TierSilver
		} else if stake >= BronzeTierThreshold {
			return TierBronze
		}
	}
	
	return ValidatorTier(-1) // Invalid tier
}

// ValidatorTierInfo stores information about a validator's tier and capabilities
type ValidatorTierInfo struct {
	ValidatorAddress    string        `json:"validator_address" yaml:"validator_address"`
	Tier               ValidatorTier  `json:"tier" yaml:"tier"`
	StakedAmount       math.Int       `json:"staked_amount" yaml:"staked_amount"`
	VerificationCount  uint64         `json:"verification_count" yaml:"verification_count"`
	SuccessfulVerifications uint64    `json:"successful_verifications" yaml:"successful_verifications"`
	TotalRewards       math.Int       `json:"total_rewards" yaml:"total_rewards"`
	LastVerification   time.Time      `json:"last_verification" yaml:"last_verification"`
	ReputationScore    math.LegacyDec `json:"reputation_score" yaml:"reputation_score"`
}

// ProtoMessage implements proto.Message interface
func (m *ValidatorTierInfo) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *ValidatorTierInfo) Reset() { *m = ValidatorTierInfo{} }

// String implements proto.Message interface
func (m *ValidatorTierInfo) String() string {
	return m.ValidatorAddress
}

// BusinessVerificationStatus represents the status of business verification
type BusinessVerificationStatus int32

const (
	VerificationPending BusinessVerificationStatus = iota
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
type BusinessVerification struct {
	BusinessID          string                      `json:"business_id" yaml:"business_id"`
	CompanyName         string                      `json:"company_name" yaml:"company_name"`
	BusinessType        string                      `json:"business_type" yaml:"business_type"`
	RegistrationCountry string                      `json:"registration_country" yaml:"registration_country"`
	DocumentHashes      []string                    `json:"document_hashes" yaml:"document_hashes"`
	RequiredTier        ValidatorTier               `json:"required_tier" yaml:"required_tier"`
	StakeRequired       math.Int                    `json:"stake_required" yaml:"stake_required"`
	VerificationFee     math.Int                    `json:"verification_fee" yaml:"verification_fee"`
	Status              BusinessVerificationStatus `json:"status" yaml:"status"`
	AssignedValidators  []string                    `json:"assigned_validators" yaml:"assigned_validators"`
	VerificationDeadline time.Time                  `json:"verification_deadline" yaml:"verification_deadline"`
	SubmittedAt         time.Time                   `json:"submitted_at" yaml:"submitted_at"`
	CompletedAt         time.Time                   `json:"completed_at" yaml:"completed_at"`
	VerificationNotes   string                      `json:"verification_notes" yaml:"verification_notes"`
	DisputeReason       string                      `json:"dispute_reason" yaml:"dispute_reason"`
}

// ProtoMessage implements proto.Message interface
func (m *BusinessVerification) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *BusinessVerification) Reset() { *m = BusinessVerification{} }

// String implements proto.Message interface
func (m *BusinessVerification) String() string {
	return m.BusinessID
}

// ValidatorBusinessStake represents a validator's equity stake in a verified business
type ValidatorBusinessStake struct {
	ValidatorAddress string        `json:"validator_address" yaml:"validator_address"`
	BusinessID       string        `json:"business_id" yaml:"business_id"`
	StakeAmount      math.Int      `json:"stake_amount" yaml:"stake_amount"`
	EquityPercentage math.LegacyDec `json:"equity_percentage" yaml:"equity_percentage"`
	StakedAt         time.Time     `json:"staked_at" yaml:"staked_at"`
	VestingPeriod    time.Duration `json:"vesting_period" yaml:"vesting_period"`
	IsVested         bool          `json:"is_vested" yaml:"is_vested"`
}

// ProtoMessage implements proto.Message interface
func (m *ValidatorBusinessStake) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *ValidatorBusinessStake) Reset() { *m = ValidatorBusinessStake{} }

// String implements proto.Message interface
func (m *ValidatorBusinessStake) String() string {
	return m.ValidatorAddress + ":" + m.BusinessID
}

// VerificationRewards tracks rewards earned by validators
type VerificationRewards struct {
	ValidatorAddress     string    `json:"validator_address" yaml:"validator_address"`
	TotalHODLRewards     math.Int  `json:"total_hodl_rewards" yaml:"total_hodl_rewards"`
	TotalEquityRewards   math.Int  `json:"total_equity_rewards" yaml:"total_equity_rewards"`
	VerificationsCompleted uint64  `json:"verifications_completed" yaml:"verifications_completed"`
	LastRewardClaim      time.Time `json:"last_reward_claim" yaml:"last_reward_claim"`
}

// ProtoMessage implements proto.Message interface
func (m *VerificationRewards) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *VerificationRewards) Reset() { *m = VerificationRewards{} }

// String implements proto.Message interface
func (m *VerificationRewards) String() string {
	return m.ValidatorAddress
}

// GenesisState defines the validator module's genesis state
type GenesisState struct {
	Params               Params                   `json:"params" yaml:"params"`
	ValidatorTiers       []ValidatorTierInfo      `json:"validator_tiers" yaml:"validator_tiers"`
	BusinessVerifications []BusinessVerification  `json:"business_verifications" yaml:"business_verifications"`
	ValidatorBusinessStakes []ValidatorBusinessStake `json:"validator_business_stakes" yaml:"validator_business_stakes"`
	VerificationRewards  []VerificationRewards    `json:"verification_rewards" yaml:"verification_rewards"`
}

// ProtoMessage implements proto.Message interface
func (m *GenesisState) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *GenesisState) Reset() { *m = GenesisState{} }

// String implements proto.Message interface
func (m *GenesisState) String() string {
	return "validator_genesis"
}

// Params defines the parameters for the validator module
// Duration fields use int64 seconds for proper JSON serialization
type Params struct {
	MinVerificationDelaySeconds int64          `json:"min_verification_delay_seconds" yaml:"min_verification_delay_seconds"`
	MaxVerificationDelaySeconds int64          `json:"max_verification_delay_seconds" yaml:"max_verification_delay_seconds"`
	VerificationTimeoutDays     uint64         `json:"verification_timeout_days" yaml:"verification_timeout_days"`
	MinValidatorsRequired       uint64         `json:"min_validators_required" yaml:"min_validators_required"`
	ReputationDecayRate         math.LegacyDec `json:"reputation_decay_rate" yaml:"reputation_decay_rate"`
	BaseVerificationReward      math.Int       `json:"base_verification_reward" yaml:"base_verification_reward"`
	EquityStakePercentage       math.LegacyDec `json:"equity_stake_percentage" yaml:"equity_stake_percentage"`
}

// Helper methods to get durations
func (p Params) MinVerificationDelay() time.Duration {
	return time.Duration(p.MinVerificationDelaySeconds) * time.Second
}

func (p Params) MaxVerificationDelay() time.Duration {
	return time.Duration(p.MaxVerificationDelaySeconds) * time.Second
}

// ProtoMessage implements proto.Message interface
func (m *Params) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *Params) Reset() { *m = Params{} }

// MsgServer defines the Msg service for validator operations
type MsgServer interface {
	RegisterValidatorTier(goCtx context.Context, msg *SimpleMsgRegisterValidatorTier) (*MsgRegisterValidatorTierResponse, error)
	SubmitBusinessVerification(goCtx context.Context, msg *SimpleMsgSubmitBusinessVerification) (*MsgSubmitBusinessVerificationResponse, error)
	CompleteVerification(goCtx context.Context, msg *SimpleMsgCompleteVerification) (*MsgCompleteVerificationResponse, error)
	StakeInBusiness(goCtx context.Context, msg *SimpleMsgStakeInBusiness) (*MsgStakeInBusinessResponse, error)
	ClaimVerificationRewards(goCtx context.Context, msg *SimpleMsgClaimVerificationRewards) (*MsgClaimVerificationRewardsResponse, error)
}

// Response types
type MsgRegisterValidatorTierResponse struct {
	Tier         ValidatorTier `json:"tier" yaml:"tier"`
	StakedAmount math.Int      `json:"staked_amount" yaml:"staked_amount"`
}

type MsgSubmitBusinessVerificationResponse struct {
	BusinessID         string    `json:"business_id" yaml:"business_id"`
	VerificationID     string    `json:"verification_id" yaml:"verification_id"`
	AssignedValidators []string  `json:"assigned_validators" yaml:"assigned_validators"`
	Deadline           time.Time `json:"deadline" yaml:"deadline"`
}

type MsgCompleteVerificationResponse struct {
	BusinessID      string                      `json:"business_id" yaml:"business_id"`
	Status          BusinessVerificationStatus `json:"status" yaml:"status"`
	RewardEarned    math.Int                   `json:"reward_earned" yaml:"reward_earned"`
	EquityAllocated math.Int                   `json:"equity_allocated" yaml:"equity_allocated"`
}

type MsgStakeInBusinessResponse struct {
	StakeAmount      math.Int       `json:"stake_amount" yaml:"stake_amount"`
	EquityPercentage math.LegacyDec `json:"equity_percentage" yaml:"equity_percentage"`
	VestingPeriod    time.Duration  `json:"vesting_period" yaml:"vesting_period"`
}

type MsgClaimVerificationRewardsResponse struct {
	HODLRewardsClaimed   math.Int `json:"hodl_rewards_claimed" yaml:"hodl_rewards_claimed"`
	EquityRewardsClaimed math.Int `json:"equity_rewards_claimed" yaml:"equity_rewards_claimed"`
}

// Query types
type QueryServer interface {
	Params(goCtx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
	ValidatorTier(goCtx context.Context, req *QueryValidatorTierRequest) (*QueryValidatorTierResponse, error)
	BusinessVerification(goCtx context.Context, req *QueryBusinessVerificationRequest) (*QueryBusinessVerificationResponse, error)
	ValidatorRewards(goCtx context.Context, req *QueryValidatorRewardsRequest) (*QueryValidatorRewardsResponse, error)
}

type QueryParamsRequest struct{}
type QueryParamsResponse struct {
	Params Params `json:"params" yaml:"params"`
}

type QueryValidatorTierRequest struct {
	ValidatorAddress string `json:"validator_address" yaml:"validator_address"`
}

type QueryValidatorTierResponse struct {
	TierInfo ValidatorTierInfo `json:"tier_info" yaml:"tier_info"`
}

type QueryBusinessVerificationRequest struct {
	BusinessID string `json:"business_id" yaml:"business_id"`
}

type QueryBusinessVerificationResponse struct {
	Verification BusinessVerification `json:"verification" yaml:"verification"`
}

type QueryValidatorRewardsRequest struct {
	ValidatorAddress string `json:"validator_address" yaml:"validator_address"`
}

type QueryValidatorRewardsResponse struct {
	Rewards VerificationRewards `json:"rewards" yaml:"rewards"`
}