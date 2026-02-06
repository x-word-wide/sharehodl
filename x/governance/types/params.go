package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Default parameter values - ALL governance-controllable via validator voting
const (
	// Voting periods (in days)
	DefaultVotingPeriodDays          uint64 = 14 // Default 14 days voting period
	DefaultValidatorVotingPeriodDays uint64 = 7  // Default 7 days for validator proposals

	// Quorum thresholds (basis points, 1000 = 10%)
	DefaultQuorumBasisPoints         uint64 = 3340 // 33.4% quorum
	DefaultCompanyQuorumBasisPoints  uint64 = 2500 // 25% for company proposals
	DefaultValidatorQuorumBasisPoints uint64 = 5000 // 50% for validator proposals
	DefaultProtocolQuorumBasisPoints uint64 = 4000 // 40% for protocol proposals

	// Pass thresholds (basis points)
	DefaultThresholdBasisPoints          uint64 = 5000 // 50% to pass
	DefaultCompanyThresholdBasisPoints   uint64 = 5000 // 50% for company proposals
	DefaultValidatorThresholdBasisPoints uint64 = 6700 // 67% for validator proposals
	DefaultProtocolThresholdBasisPoints  uint64 = 5000 // 50% for protocol proposals

	// Veto thresholds (basis points)
	DefaultVetoBasisPoints        uint64 = 3340 // 33.4% veto threshold
	DefaultCompanyVetoBasisPoints uint64 = 3340 // 33.4% for company proposals

	// Executive compensation threshold
	DefaultExecutiveCompThresholdBasisPoints uint64 = 5000 // 50%

	// ==========================================================================
	// ANTI-SPAM PARAMETERS - Prevent governance abuse
	// ==========================================================================

	// Deposit requirements (in uhodl - micro HODL)
	DefaultMinDeposit uint64 = 10000000000 // 10,000 HODL minimum deposit (refundable if proposal passes/voting)
	DefaultProposalFee uint64 = 1000000    // 1 HODL non-refundable fee (burned on submission)

	// Rate limiting
	DefaultMaxProposalsPerDay    uint64 = 3   // Max 3 proposals per address per day
	DefaultProposalCooldownHours uint64 = 24  // 24 hour cooldown between proposals from same address

	// Stake requirements
	DefaultMinProposerStake uint64 = 1000000000 // 1,000 HODL minimum balance to submit proposals

	// Deposit period (days to reach minimum deposit)
	DefaultDepositPeriodDays uint64 = 7 // 7 days to reach minimum deposit or proposal is deleted
)

// ExtendedParams defines extended parameters for the governance module
// ALL parameters are governance-controllable via validator voting
// Note: This extends the existing GovernanceParams in governance.go
type ExtendedParams struct {
	// Voting periods
	VotingPeriodDays          uint64 `json:"voting_period_days" yaml:"voting_period_days"`
	ValidatorVotingPeriodDays uint64 `json:"validator_voting_period_days" yaml:"validator_voting_period_days"`

	// Quorum thresholds (basis points)
	QuorumBasisPoints          uint64 `json:"quorum_basis_points" yaml:"quorum_basis_points"`
	CompanyQuorumBasisPoints   uint64 `json:"company_quorum_basis_points" yaml:"company_quorum_basis_points"`
	ValidatorQuorumBasisPoints uint64 `json:"validator_quorum_basis_points" yaml:"validator_quorum_basis_points"`
	ProtocolQuorumBasisPoints  uint64 `json:"protocol_quorum_basis_points" yaml:"protocol_quorum_basis_points"`

	// Pass thresholds (basis points)
	ThresholdBasisPoints          uint64 `json:"threshold_basis_points" yaml:"threshold_basis_points"`
	CompanyThresholdBasisPoints   uint64 `json:"company_threshold_basis_points" yaml:"company_threshold_basis_points"`
	ValidatorThresholdBasisPoints uint64 `json:"validator_threshold_basis_points" yaml:"validator_threshold_basis_points"`
	ProtocolThresholdBasisPoints  uint64 `json:"protocol_threshold_basis_points" yaml:"protocol_threshold_basis_points"`

	// Veto thresholds (basis points)
	VetoBasisPoints        uint64 `json:"veto_basis_points" yaml:"veto_basis_points"`
	CompanyVetoBasisPoints uint64 `json:"company_veto_basis_points" yaml:"company_veto_basis_points"`

	// Special thresholds
	ExecutiveCompThresholdBasisPoints uint64 `json:"executive_comp_threshold_basis_points" yaml:"executive_comp_threshold_basis_points"`

	// Master toggles
	GovernanceEnabled bool `json:"governance_enabled" yaml:"governance_enabled"`

	// ==========================================================================
	// ANTI-SPAM PARAMETERS
	// ==========================================================================

	// Deposit requirements (in uhodl)
	MinDeposit  uint64 `json:"min_deposit" yaml:"min_deposit"`   // Minimum deposit to create proposal (refundable)
	ProposalFee uint64 `json:"proposal_fee" yaml:"proposal_fee"` // Non-refundable fee burned on submission

	// Rate limiting
	MaxProposalsPerDay    uint64 `json:"max_proposals_per_day" yaml:"max_proposals_per_day"`       // Max proposals per address per day
	ProposalCooldownHours uint64 `json:"proposal_cooldown_hours" yaml:"proposal_cooldown_hours"`   // Hours between proposals

	// Stake requirements
	MinProposerStake uint64 `json:"min_proposer_stake" yaml:"min_proposer_stake"` // Minimum HODL balance to propose

	// Deposit period
	DepositPeriodDays uint64 `json:"deposit_period_days" yaml:"deposit_period_days"` // Days to reach min deposit
}

// ProtoMessage implements proto.Message interface
func (m *ExtendedParams) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *ExtendedParams) Reset() { *m = ExtendedParams{} }

// String implements proto.Message interface
func (m *ExtendedParams) String() string { return "governance_extended_params" }

// DefaultExtendedParams returns default extended governance parameters
// ALL values can be changed via governance proposals
func DefaultExtendedParams() ExtendedParams {
	return ExtendedParams{
		VotingPeriodDays:                  DefaultVotingPeriodDays,
		ValidatorVotingPeriodDays:         DefaultValidatorVotingPeriodDays,
		QuorumBasisPoints:                 DefaultQuorumBasisPoints,
		CompanyQuorumBasisPoints:          DefaultCompanyQuorumBasisPoints,
		ValidatorQuorumBasisPoints:        DefaultValidatorQuorumBasisPoints,
		ProtocolQuorumBasisPoints:         DefaultProtocolQuorumBasisPoints,
		ThresholdBasisPoints:              DefaultThresholdBasisPoints,
		CompanyThresholdBasisPoints:       DefaultCompanyThresholdBasisPoints,
		ValidatorThresholdBasisPoints:     DefaultValidatorThresholdBasisPoints,
		ProtocolThresholdBasisPoints:      DefaultProtocolThresholdBasisPoints,
		VetoBasisPoints:                   DefaultVetoBasisPoints,
		CompanyVetoBasisPoints:            DefaultCompanyVetoBasisPoints,
		ExecutiveCompThresholdBasisPoints: DefaultExecutiveCompThresholdBasisPoints,
		GovernanceEnabled:                 true,
		// Anti-spam parameters
		MinDeposit:            DefaultMinDeposit,
		ProposalFee:           DefaultProposalFee,
		MaxProposalsPerDay:    DefaultMaxProposalsPerDay,
		ProposalCooldownHours: DefaultProposalCooldownHours,
		MinProposerStake:      DefaultMinProposerStake,
		DepositPeriodDays:     DefaultDepositPeriodDays,
	}
}

// Validate validates the extended params
func (p ExtendedParams) Validate() error {
	if p.VotingPeriodDays == 0 {
		return fmt.Errorf("voting period must be positive")
	}
	if p.ValidatorVotingPeriodDays == 0 {
		return fmt.Errorf("validator voting period must be positive")
	}
	// Quorum cannot exceed 100%
	if p.QuorumBasisPoints > 10000 || p.CompanyQuorumBasisPoints > 10000 ||
		p.ValidatorQuorumBasisPoints > 10000 || p.ProtocolQuorumBasisPoints > 10000 {
		return fmt.Errorf("quorum cannot exceed 100%%")
	}
	// Thresholds cannot exceed 100%
	if p.ThresholdBasisPoints > 10000 || p.CompanyThresholdBasisPoints > 10000 ||
		p.ValidatorThresholdBasisPoints > 10000 || p.ProtocolThresholdBasisPoints > 10000 {
		return fmt.Errorf("threshold cannot exceed 100%%")
	}
	// Veto cannot exceed 100%
	if p.VetoBasisPoints > 10000 || p.CompanyVetoBasisPoints > 10000 {
		return fmt.Errorf("veto threshold cannot exceed 100%%")
	}

	// Validate anti-spam parameters
	if p.MaxProposalsPerDay == 0 {
		return fmt.Errorf("max proposals per day must be positive")
	}
	if p.MaxProposalsPerDay > 100 {
		return fmt.Errorf("max proposals per day cannot exceed 100")
	}
	if p.DepositPeriodDays == 0 {
		return fmt.Errorf("deposit period must be positive")
	}
	// Note: MinDeposit, ProposalFee, MinProposerStake can be 0 (disabled)

	return nil
}

// Helper methods to convert basis points to decimals

// GetQuorum returns quorum as a decimal (e.g., 0.334 for 33.4%)
func (p ExtendedParams) GetQuorum() math.LegacyDec {
	return math.LegacyNewDec(int64(p.QuorumBasisPoints)).QuoInt64(10000)
}

// GetCompanyQuorum returns company quorum as a decimal
func (p ExtendedParams) GetCompanyQuorum() math.LegacyDec {
	return math.LegacyNewDec(int64(p.CompanyQuorumBasisPoints)).QuoInt64(10000)
}

// GetValidatorQuorum returns validator quorum as a decimal
func (p ExtendedParams) GetValidatorQuorum() math.LegacyDec {
	return math.LegacyNewDec(int64(p.ValidatorQuorumBasisPoints)).QuoInt64(10000)
}

// GetProtocolQuorum returns protocol quorum as a decimal
func (p ExtendedParams) GetProtocolQuorum() math.LegacyDec {
	return math.LegacyNewDec(int64(p.ProtocolQuorumBasisPoints)).QuoInt64(10000)
}

// GetThreshold returns pass threshold as a decimal
func (p ExtendedParams) GetThreshold() math.LegacyDec {
	return math.LegacyNewDec(int64(p.ThresholdBasisPoints)).QuoInt64(10000)
}

// GetCompanyThreshold returns company threshold as a decimal
func (p ExtendedParams) GetCompanyThreshold() math.LegacyDec {
	return math.LegacyNewDec(int64(p.CompanyThresholdBasisPoints)).QuoInt64(10000)
}

// GetValidatorThreshold returns validator threshold as a decimal
func (p ExtendedParams) GetValidatorThreshold() math.LegacyDec {
	return math.LegacyNewDec(int64(p.ValidatorThresholdBasisPoints)).QuoInt64(10000)
}

// GetProtocolThreshold returns protocol threshold as a decimal
func (p ExtendedParams) GetProtocolThreshold() math.LegacyDec {
	return math.LegacyNewDec(int64(p.ProtocolThresholdBasisPoints)).QuoInt64(10000)
}

// GetVetoThreshold returns veto threshold as a decimal
func (p ExtendedParams) GetVetoThreshold() math.LegacyDec {
	return math.LegacyNewDec(int64(p.VetoBasisPoints)).QuoInt64(10000)
}

// GetCompanyVetoThreshold returns company veto threshold as a decimal
func (p ExtendedParams) GetCompanyVetoThreshold() math.LegacyDec {
	return math.LegacyNewDec(int64(p.CompanyVetoBasisPoints)).QuoInt64(10000)
}

// ParamsKey is the store key for governance params
var ParamsKey = []byte{0x40}
