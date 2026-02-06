package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
)

// ProposalType represents the type of governance proposal
type ProposalType int32

const (
	ProposalTypeText                ProposalType = 0
	ProposalTypeParameterChange     ProposalType = 1
	ProposalTypeSoftwareUpgrade     ProposalType = 2
	ProposalTypeCommunityPoolSpend  ProposalType = 3
	ProposalTypeCompanyGovernance   ProposalType = 4
	ProposalTypeValidatorTierChange ProposalType = 5
	ProposalTypeListingRequirement  ProposalType = 6
	ProposalTypeTradingHalt         ProposalType = 7
	ProposalTypeEmergencyAction     ProposalType = 8
	ProposalTypeSetCharityWallet    ProposalType = 9  // Set protocol default charity wallet
	// Fee Abstraction proposal types
	ProposalTypeFeeAbstractionParams ProposalType = 50 // Update fee abstraction parameters
	ProposalTypeTreasuryGrant        ProposalType = 51 // Grant treasury allowance
	ProposalTypeTreasuryFunding      ProposalType = 52 // Fund protocol treasury
)

// String returns the string representation of ProposalType
func (pt ProposalType) String() string {
	switch pt {
	case ProposalTypeText:
		return "text"
	case ProposalTypeParameterChange:
		return "parameter_change"
	case ProposalTypeSoftwareUpgrade:
		return "software_upgrade"
	case ProposalTypeCommunityPoolSpend:
		return "community_pool_spend"
	case ProposalTypeCompanyGovernance:
		return "company_governance"
	case ProposalTypeValidatorTierChange:
		return "validator_tier_change"
	case ProposalTypeListingRequirement:
		return "listing_requirement"
	case ProposalTypeTradingHalt:
		return "trading_halt"
	case ProposalTypeEmergencyAction:
		return "emergency_action"
	case ProposalTypeSetCharityWallet:
		return "set_charity_wallet"
	case ProposalTypeFeeAbstractionParams:
		return "fee_abstraction_params"
	case ProposalTypeTreasuryGrant:
		return "treasury_grant"
	case ProposalTypeTreasuryFunding:
		return "treasury_funding"
	default:
		return "unknown"
	}
}

// ProposalStatus represents the current status of a proposal
type ProposalStatus int32

const (
	ProposalStatusDepositPeriod ProposalStatus = 0
	ProposalStatusVotingPeriod  ProposalStatus = 1
	ProposalStatusPassed        ProposalStatus = 2
	ProposalStatusRejected      ProposalStatus = 3
	ProposalStatusFailed        ProposalStatus = 4
	ProposalStatusCanceled      ProposalStatus = 5
)

// String returns the string representation of ProposalStatus
func (ps ProposalStatus) String() string {
	switch ps {
	case ProposalStatusDepositPeriod:
		return "deposit_period"
	case ProposalStatusVotingPeriod:
		return "voting_period"
	case ProposalStatusPassed:
		return "passed"
	case ProposalStatusRejected:
		return "rejected"
	case ProposalStatusFailed:
		return "failed"
	case ProposalStatusCanceled:
		return "canceled"
	default:
		return "unknown"
	}
}

// VoteOption represents a vote choice
type VoteOption int32

const (
	VoteOptionEmpty      VoteOption = 0
	VoteOptionYes        VoteOption = 1
	VoteOptionAbstain    VoteOption = 2
	VoteOptionNo         VoteOption = 3
	VoteOptionNoWithVeto VoteOption = 4
)

// String returns the string representation of VoteOption
func (vo VoteOption) String() string {
	switch vo {
	case VoteOptionYes:
		return "yes"
	case VoteOptionAbstain:
		return "abstain"
	case VoteOptionNo:
		return "no"
	case VoteOptionNoWithVeto:
		return "no_with_veto"
	default:
		return "empty"
	}
}

// Proposal represents a governance proposal
type Proposal struct {
	ID          uint64       `json:"id"`
	Type        ProposalType `json:"type"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Proposer    string       `json:"proposer"`
	Status      ProposalStatus `json:"status"`
	
	// Voting parameters
	VotingStartTime    time.Time `json:"voting_start_time"`
	VotingEndTime      time.Time `json:"voting_end_time"`
	DepositEndTime     time.Time `json:"deposit_end_time"`
	
	// Deposit information
	TotalDeposit    math.Int `json:"total_deposit"`
	MinDeposit      math.Int `json:"min_deposit"`
	Deposits        []Deposit `json:"deposits"`
	
	// Vote tallying
	YesVotes        math.Int `json:"yes_votes"`
	NoVotes         math.Int `json:"no_votes"`
	AbstainVotes    math.Int `json:"abstain_votes"`
	NoWithVetoVotes math.Int `json:"no_with_veto_votes"`
	TotalVotingPower math.Int `json:"total_voting_power"`
	
	// Quorum and threshold requirements
	Quorum          math.LegacyDec `json:"quorum"`           // Minimum participation
	Threshold       math.LegacyDec `json:"threshold"`        // Minimum yes votes
	VetoThreshold   math.LegacyDec `json:"veto_threshold"`   // Veto threshold
	
	// Execution information
	Executed        bool      `json:"executed"`
	ExecutionTime   time.Time `json:"execution_time,omitempty"`
	ExecutionResult string    `json:"execution_result,omitempty"`
	
	// Metadata
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CompanyID       uint64    `json:"company_id,omitempty"`     // For company-specific proposals
	ValidatorAddr   string    `json:"validator_addr,omitempty"` // For validator-specific proposals
	
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	FinalizedAt     time.Time `json:"finalized_at,omitempty"`
}

// Vote represents a single vote on a proposal
type Vote struct {
	ProposalID   uint64     `json:"proposal_id"`
	Voter        string     `json:"voter"`
	Option       VoteOption `json:"option"`
	VotingPower  math.Int   `json:"voting_power"`
	Weight       math.LegacyDec `json:"weight"`     // For weighted votes
	Justification string    `json:"justification,omitempty"`
	CompanyID    uint64     `json:"company_id,omitempty"` // For company governance votes
	VotedAt      time.Time  `json:"voted_at"`
}

// Deposit represents a deposit on a proposal
type Deposit struct {
	ProposalID uint64   `json:"proposal_id"`
	Depositor  string   `json:"depositor"`
	Amount     math.Int `json:"amount"`
	Currency   string   `json:"currency"`
	DepositedAt time.Time `json:"deposited_at"`
}

// WeightedVote represents a vote with multiple options and weights
type WeightedVote struct {
	ProposalID  uint64              `json:"proposal_id"`
	Voter       string              `json:"voter"`
	Options     []WeightedVoteOption `json:"options"`
	VotingPower math.Int            `json:"voting_power"`
	VotedAt     time.Time           `json:"voted_at"`
}

// WeightedVoteOption represents a single option in a weighted vote
type WeightedVoteOption struct {
	Option VoteOption     `json:"option"`
	Weight math.LegacyDec `json:"weight"`
}

// TallyResult represents the result of vote tallying
type TallyResult struct {
	YesVotes        math.Int `json:"yes_votes"`
	NoVotes         math.Int `json:"no_votes"`
	AbstainVotes    math.Int `json:"abstain_votes"`
	NoWithVetoVotes math.Int `json:"no_with_veto_votes"`
	TotalVotingPower math.Int `json:"total_voting_power"`
	Turnout         math.LegacyDec `json:"turnout"`
	YesPercentage   math.LegacyDec `json:"yes_percentage"`
	NoPercentage    math.LegacyDec `json:"no_percentage"`
	VetoPercentage  math.LegacyDec `json:"veto_percentage"`
}

// GovernanceParams defines governance parameters
type GovernanceParams struct {
	// Proposal parameters
	MinDepositRatio    math.LegacyDec `json:"min_deposit_ratio"`    // Minimum deposit as ratio of total supply
	MaxDepositPeriod   time.Duration  `json:"max_deposit_period"`   // Maximum deposit period
	VotingPeriod       time.Duration  `json:"voting_period"`        // Voting period duration
	
	// Voting thresholds
	Quorum            math.LegacyDec `json:"quorum"`             // Minimum participation
	Threshold         math.LegacyDec `json:"threshold"`          // Minimum yes votes to pass
	VetoThreshold     math.LegacyDec `json:"veto_threshold"`     // Veto threshold
	
	// Emergency parameters
	EmergencyQuorum   math.LegacyDec `json:"emergency_quorum"`   // Emergency proposal quorum
	EmergencyThreshold math.LegacyDec `json:"emergency_threshold"` // Emergency proposal threshold
	EmergencyVotingPeriod time.Duration `json:"emergency_voting_period"` // Emergency voting period
	
	// Company governance parameters
	CompanyQuorum     math.LegacyDec `json:"company_quorum"`     // Company proposal quorum
	CompanyThreshold  math.LegacyDec `json:"company_threshold"`  // Company proposal threshold
	ShareholderVoting bool           `json:"shareholder_voting"` // Enable shareholder voting
	
	// Validator governance parameters
	ValidatorQuorum   math.LegacyDec `json:"validator_quorum"`   // Validator proposal quorum
	ValidatorThreshold math.LegacyDec `json:"validator_threshold"` // Validator proposal threshold
}

// CompanyGovernanceProposal represents a company-specific governance proposal
type CompanyGovernanceProposal struct {
	ProposalID        uint64                    `json:"proposal_id"`
	CompanyID         uint64                    `json:"company_id"`
	CompanySymbol     string                    `json:"company_symbol"`
	Type              CompanyProposalType       `json:"type"`
	Title             string                    `json:"title"`
	Description       string                    `json:"description"`
	Proposer          string                    `json:"proposer"`
	ShareClassVoting  map[string]math.LegacyDec `json:"share_class_voting"` // Voting weights by share class
	RequiredApproval  math.LegacyDec            `json:"required_approval"`  // Required approval percentage
	BoardApprovalReq  bool                      `json:"board_approval_req"` // Requires board approval
	QuorumRequirement math.LegacyDec            `json:"quorum_requirement"`
	VotingPeriod      time.Duration             `json:"voting_period"`
	ExecutionDelay    time.Duration             `json:"execution_delay"`    // Delay before execution
	ProposalData      map[string]interface{}    `json:"proposal_data"`     // Type-specific data
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
}

// ValidatorGovernanceProposal represents a validator-specific governance proposal
type ValidatorGovernanceProposal struct {
	ProposalID    uint64    `json:"proposal_id"`
	ValidatorAddr string    `json:"validator_addr"`
	Type          string    `json:"type"` // "tier_change", "removal", "penalty", etc.
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Proposer      string    `json:"proposer"`
	Evidence      string    `json:"evidence,omitempty"`
	NewTier       int       `json:"new_tier,omitempty"`
	PenaltyAmount math.Int  `json:"penalty_amount,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// ProposalQueue represents the queue of proposals for execution
type ProposalQueue struct {
	ProposalID    uint64    `json:"proposal_id"`
	ExecutionTime time.Time `json:"execution_time"`
	Priority      int       `json:"priority"`
}

// NewProposal creates a new proposal
func NewProposal(
	id uint64,
	proposalType ProposalType,
	title, description, proposer string,
	depositEndTime, votingEndTime time.Time,
	minDeposit math.Int,
) Proposal {
	now := time.Now()
	
	return Proposal{
		ID:              id,
		Type:            proposalType,
		Title:           title,
		Description:     description,
		Proposer:        proposer,
		Status:          ProposalStatusDepositPeriod,
		DepositEndTime:  depositEndTime,
		VotingEndTime:   votingEndTime,
		MinDeposit:      minDeposit,
		TotalDeposit:    math.ZeroInt(),
		Deposits:        []Deposit{},
		YesVotes:        math.ZeroInt(),
		NoVotes:         math.ZeroInt(),
		AbstainVotes:    math.ZeroInt(),
		NoWithVetoVotes: math.ZeroInt(),
		TotalVotingPower: math.ZeroInt(),
		Quorum:          math.LegacyNewDecWithPrec(334, 3), // 33.4%
		Threshold:       math.LegacyNewDecWithPrec(5, 1),   // 50%
		VetoThreshold:   math.LegacyNewDecWithPrec(334, 3), // 33.4%
		Executed:        false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Validate validates a proposal
func (p Proposal) Validate() error {
	if p.Title == "" {
		return ErrInvalidProposal
	}
	
	if p.Description == "" {
		return ErrInvalidProposal
	}
	
	if p.Proposer == "" {
		return fmt.Errorf("invalid proposer: cannot be empty")
	}
	
	if p.MinDeposit.IsNegative() {
		return ErrInvalidDeposit
	}
	
	if p.VotingEndTime.Before(time.Now()) {
		return ErrInvalidVotingPeriod
	}
	
	return nil
}

// HasMetDeposit checks if the proposal has met the minimum deposit requirement
func (p Proposal) HasMetDeposit() bool {
	return p.TotalDeposit.GTE(p.MinDeposit)
}

// IsInVotingPeriod checks if the proposal is in voting period
func (p Proposal) IsInVotingPeriod(currentTime time.Time) bool {
	return p.Status == ProposalStatusVotingPeriod &&
		currentTime.After(p.VotingStartTime) &&
		currentTime.Before(p.VotingEndTime)
}

// IsInDepositPeriod checks if the proposal is in deposit period
func (p Proposal) IsInDepositPeriod(currentTime time.Time) bool {
	return p.Status == ProposalStatusDepositPeriod &&
		currentTime.Before(p.DepositEndTime)
}

// CalculateTallyResult calculates the tally result
func (p Proposal) CalculateTallyResult() TallyResult {
	totalVotes := p.YesVotes.Add(p.NoVotes).Add(p.AbstainVotes).Add(p.NoWithVetoVotes)
	
	var turnout, yesPercentage, noPercentage, vetoPercentage math.LegacyDec
	
	if p.TotalVotingPower.GT(math.ZeroInt()) {
		turnout = math.LegacyNewDecFromInt(totalVotes).QuoInt(p.TotalVotingPower)
	}
	
	if totalVotes.GT(math.ZeroInt()) {
		yesPercentage = math.LegacyNewDecFromInt(p.YesVotes).QuoInt(totalVotes)
		noPercentage = math.LegacyNewDecFromInt(p.NoVotes).QuoInt(totalVotes)
		vetoPercentage = math.LegacyNewDecFromInt(p.NoWithVetoVotes).QuoInt(totalVotes)
	}
	
	return TallyResult{
		YesVotes:         p.YesVotes,
		NoVotes:          p.NoVotes,
		AbstainVotes:     p.AbstainVotes,
		NoWithVetoVotes:  p.NoWithVetoVotes,
		TotalVotingPower: p.TotalVotingPower,
		Turnout:          turnout,
		YesPercentage:    yesPercentage,
		NoPercentage:     noPercentage,
		VetoPercentage:   vetoPercentage,
	}
}

// ShouldPass determines if the proposal should pass based on vote tallies
func (p Proposal) ShouldPass() bool {
	tally := p.CalculateTallyResult()
	
	// Check if quorum is met
	if tally.Turnout.LT(p.Quorum) {
		return false
	}
	
	// Check if veto threshold is exceeded
	if tally.VetoPercentage.GTE(p.VetoThreshold) {
		return false
	}
	
	// Check if threshold is met
	return tally.YesPercentage.GTE(p.Threshold)
}

// NewVote creates a new vote
func NewVote(proposalID uint64, voter string, option VoteOption, votingPower math.Int) Vote {
	return Vote{
		ProposalID:  proposalID,
		Voter:       voter,
		Option:      option,
		VotingPower: votingPower,
		Weight:      math.LegacyNewDec(1),
		VotedAt:     time.Now(),
	}
}

// NewDeposit creates a new deposit
func NewDeposit(proposalID uint64, depositor string, amount math.Int, currency string) Deposit {
	return Deposit{
		ProposalID:  proposalID,
		Depositor:   depositor,
		Amount:      amount,
		Currency:    currency,
		DepositedAt: time.Now(),
	}
}

// CompanyProposalType represents different types of company governance proposals
type CompanyProposalType int32

const (
	CompanyProposalTypeBoardElection    CompanyProposalType = 0
	CompanyProposalTypeDividendPolicy   CompanyProposalType = 1
	CompanyProposalTypeMergerAcquisition CompanyProposalType = 2
	CompanyProposalTypeShareSplit       CompanyProposalType = 3
	CompanyProposalTypeShareBuyback     CompanyProposalType = 4
	CompanyProposalTypeBylaw           CompanyProposalType = 5
	CompanyProposalTypeExecutiveComp   CompanyProposalType = 6
	CompanyProposalTypeAuditorSelection CompanyProposalType = 7
	CompanyProposalTypeCapitalStructure CompanyProposalType = 8
	CompanyProposalTypeAssetSale       CompanyProposalType = 9
)

// String returns the string representation of CompanyProposalType
func (cpt CompanyProposalType) String() string {
	switch cpt {
	case CompanyProposalTypeBoardElection:
		return "board_election"
	case CompanyProposalTypeDividendPolicy:
		return "dividend_policy"
	case CompanyProposalTypeMergerAcquisition:
		return "merger_acquisition"
	case CompanyProposalTypeShareSplit:
		return "share_split"
	case CompanyProposalTypeShareBuyback:
		return "share_buyback"
	case CompanyProposalTypeBylaw:
		return "bylaw_change"
	case CompanyProposalTypeExecutiveComp:
		return "executive_compensation"
	case CompanyProposalTypeAuditorSelection:
		return "auditor_selection"
	case CompanyProposalTypeCapitalStructure:
		return "capital_structure"
	case CompanyProposalTypeAssetSale:
		return "asset_sale"
	default:
		return "unknown"
	}
}

// VoteDelegation represents voting power delegation
type VoteDelegation struct {
	Delegator     string             `json:"delegator"`
	Delegate      string             `json:"delegate"`
	CompanyID     uint64             `json:"company_id"`     // Company-specific delegation
	ProposalType  ProposalType       `json:"proposal_type"`  // Type-specific delegation
	VotingPower   math.LegacyDec     `json:"voting_power"`   // Delegated voting power
	ExpiryHeight  uint64             `json:"expiry_height"`  // Delegation expiry
	Revocable     bool               `json:"revocable"`      // Can be revoked by delegator
	CreatedAt     time.Time          `json:"created_at"`
	LastUsed      time.Time          `json:"last_used,omitempty"`
}

// GovernanceStats tracks governance participation statistics
type GovernanceStats struct {
	TotalProposals       uint64             `json:"total_proposals"`
	ActiveProposals      uint64             `json:"active_proposals"`
	TotalVotes           uint64             `json:"total_votes"`
	AverageParticipation math.LegacyDec     `json:"average_participation"`
	TopValidatorsByVotes []string           `json:"top_validators_by_votes"`
	ProposalSuccessRate  math.LegacyDec     `json:"proposal_success_rate"`
	LastUpdated          time.Time          `json:"last_updated"`
}

// EmergencyProposal represents emergency proposals with expedited procedures
type EmergencyProposal struct {
	ProposalID      uint64         `json:"proposal_id"`
	EmergencyType   EmergencyType  `json:"emergency_type"`
	SeverityLevel   int            `json:"severity_level"`    // 1-5, 5 being most severe
	Justification   string         `json:"justification"`
	RequiredTier    string         `json:"required_tier"`     // Minimum validator tier to vote
	FastTrack       bool           `json:"fast_track"`        // Use reduced voting period
	ExecuteOnPass   bool           `json:"execute_on_pass"`   // Execute immediately on pass
	TimeLimit       time.Duration  `json:"time_limit"`        // Maximum execution time
	CreatedAt       time.Time      `json:"created_at"`
}

// EmergencyType represents different types of emergencies
type EmergencyType int32

const (
	EmergencyTypeSecurityBreach  EmergencyType = 0
	EmergencyTypeMarketHalt      EmergencyType = 1
	EmergencyTypeValidatorSlash  EmergencyType = 2
	EmergencyTypeProtocolUpgrade EmergencyType = 3
	EmergencyTypeCompanyDelisting EmergencyType = 4
	EmergencyTypeSystemOutage    EmergencyType = 5
)

// DefaultGovernanceParams returns default governance parameters
func DefaultGovernanceParams() GovernanceParams {
	return GovernanceParams{
		MinDepositRatio:       math.LegacyNewDecWithPrec(1, 4),      // 0.01%
		MaxDepositPeriod:      time.Hour * 24 * 14,                 // 2 weeks
		VotingPeriod:          time.Hour * 24 * 7,                  // 1 week
		Quorum:                math.LegacyNewDecWithPrec(334, 3),   // 33.4%
		Threshold:             math.LegacyNewDecWithPrec(5, 1),     // 50%
		VetoThreshold:         math.LegacyNewDecWithPrec(334, 3),   // 33.4%
		EmergencyQuorum:       math.LegacyNewDecWithPrec(5, 1),     // 50%
		EmergencyThreshold:    math.LegacyNewDecWithPrec(67, 2),    // 67%
		EmergencyVotingPeriod: time.Hour * 24 * 3,                 // 3 days
		CompanyQuorum:         math.LegacyNewDecWithPrec(5, 1),     // 50%
		CompanyThreshold:      math.LegacyNewDecWithPrec(67, 2),    // 67%
		ShareholderVoting:     true,
		ValidatorQuorum:       math.LegacyNewDecWithPrec(67, 2),    // 67%
		ValidatorThreshold:    math.LegacyNewDecWithPrec(75, 2),    // 75%
	}
}

// =============================================================================
// STATUS ALIASES (for keeper compatibility)
// =============================================================================

// Status aliases for keeper compatibility
type Status string

const (
	StatusDepositPeriod   Status = "deposit_period"
	StatusVotingPeriod    Status = "voting_period"
	StatusPassed          Status = "passed"
	StatusRejected        Status = "rejected"
	StatusCancelled       Status = "cancelled"
	StatusVetoed          Status = "vetoed"
	StatusExecuted        Status = "executed"
	StatusExecutionFailed Status = "execution_failed"
)

// VoteOption aliases for keeper compatibility
type VoteOptionAlias string

const (
	OptionYes     VoteOptionAlias = "yes"
	OptionNo      VoteOptionAlias = "no"
	OptionAbstain VoteOptionAlias = "abstain"
	OptionVeto    VoteOptionAlias = "no_with_veto"
)

// ProposalOutcome represents the outcome of a proposal
type ProposalOutcome string

const (
	OutcomeVotingInProgress ProposalOutcome = "voting_in_progress"
	OutcomePassed           ProposalOutcome = "passed"
	OutcomeRejected         ProposalOutcome = "rejected"
	OutcomeQuorumNotMet     ProposalOutcome = "quorum_not_met"
	OutcomeVetoed           ProposalOutcome = "vetoed"
)

// Extended ProposalType constants for keeper compatibility
const (
	ProposalTypeCompanyListing    ProposalType = 10
	ProposalTypeCompanyDelisting  ProposalType = 11
	ProposalTypeCompanyParameter  ProposalType = 12
	ProposalTypeValidatorPromotion ProposalType = 20
	ProposalTypeValidatorDemotion  ProposalType = 21
	ProposalTypeValidatorRemoval   ProposalType = 22
	ProposalTypeProtocolParameter  ProposalType = 30
	ProposalTypeProtocolUpgrade    ProposalType = 31
	ProposalTypeTreasurySpend      ProposalType = 40
)

// =============================================================================
// EXTENDED TYPES FOR KEEPER
// =============================================================================

// Extended Proposal for keeper compatibility
type KeeperProposal struct {
	ID                uint64         `json:"id"`
	Type              ProposalType   `json:"type"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	Submitter         string         `json:"submitter"`
	Status            Status         `json:"status"`
	SubmissionTime    time.Time      `json:"submission_time"`
	DepositEndTime    time.Time      `json:"deposit_end_time"`
	VotingStartTime   time.Time      `json:"voting_start_time"`
	VotingEndTime     time.Time      `json:"voting_end_time"`
	TotalDeposit      math.Int       `json:"total_deposit"`
	QuorumRequired    math.LegacyDec `json:"quorum_required"`
	ThresholdRequired math.LegacyDec `json:"threshold_required"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// Extended Vote for keeper compatibility
type KeeperVote struct {
	ProposalID   uint64          `json:"proposal_id"`
	Voter        string          `json:"voter"`
	Option       VoteOptionAlias `json:"option"`
	Weight       math.LegacyDec  `json:"weight"`
	VotingPower  math.LegacyDec  `json:"voting_power"`
	Reason       string          `json:"reason,omitempty"`
	SubmittedAt  time.Time       `json:"submitted_at"`
}

// Extended WeightedVote for keeper compatibility
type KeeperWeightedVote struct {
	ProposalID   uint64                  `json:"proposal_id"`
	Voter        string                  `json:"voter"`
	Options      []WeightedVoteOption    `json:"options"`
	VotingPower  math.LegacyDec          `json:"voting_power"`
	Reason       string                  `json:"reason,omitempty"`
	SubmittedAt  time.Time               `json:"submitted_at"`
}

// Extended TallyResult for keeper compatibility
type KeeperTallyResult struct {
	ProposalID         uint64          `json:"proposal_id"`
	YesVotes           math.LegacyDec  `json:"yes_votes"`
	NoVotes            math.LegacyDec  `json:"no_votes"`
	AbstainVotes       math.LegacyDec  `json:"abstain_votes"`
	VetoVotes          math.LegacyDec  `json:"veto_votes"`
	TotalVotes         math.LegacyDec  `json:"total_votes"`
	TotalEligibleVotes math.LegacyDec  `json:"total_eligible_votes"`
	ParticipationRate  math.LegacyDec  `json:"participation_rate"`
	YesPercentage      math.LegacyDec  `json:"yes_percentage"`
	NoPercentage       math.LegacyDec  `json:"no_percentage"`
	AbstainPercentage  math.LegacyDec  `json:"abstain_percentage"`
	VetoPercentage     math.LegacyDec  `json:"veto_percentage"`
	Outcome            ProposalOutcome `json:"outcome"`
}

// Extended Deposit for keeper compatibility
type KeeperDeposit struct {
	ProposalID  uint64    `json:"proposal_id"`
	Depositor   string    `json:"depositor"`
	Amount      math.Int  `json:"amount"`
	DepositedAt time.Time `json:"deposited_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Extended GovernanceParams for keeper compatibility
type KeeperGovernanceParams struct {
	MinDeposit             math.Int       `json:"min_deposit"`
	MaxDepositPeriodDays   uint32         `json:"max_deposit_period_days"`
	VotingPeriodDays       uint32         `json:"voting_period_days"`
	Quorum                 math.LegacyDec `json:"quorum"`
	Threshold              math.LegacyDec `json:"threshold"`
	VetoThreshold          math.LegacyDec `json:"veto_threshold"`
	BurnDeposits           bool           `json:"burn_deposits"`
	BurnVoteVeto           bool           `json:"burn_vote_veto"`
	MinInitialDepositRatio math.LegacyDec `json:"min_initial_deposit_ratio"`
}

// DefaultParams returns the default governance parameters for module initialization
func DefaultParams() GovernanceParams {
	return GovernanceParams{
		MinDepositRatio:       math.LegacyNewDecWithPrec(1, 4),      // 0.01%
		MaxDepositPeriod:      time.Hour * 24 * 7,                   // 1 week
		VotingPeriod:          time.Hour * 24 * 14,                  // 2 weeks
		Quorum:                math.LegacyNewDecWithPrec(334, 3),    // 33.4%
		Threshold:             math.LegacyNewDecWithPrec(5, 1),      // 50%
		VetoThreshold:         math.LegacyNewDecWithPrec(334, 3),    // 33.4%
		EmergencyQuorum:       math.LegacyNewDecWithPrec(5, 1),      // 50%
		EmergencyThreshold:    math.LegacyNewDecWithPrec(67, 2),     // 67%
		EmergencyVotingPeriod: time.Hour * 24 * 3,                   // 3 days
		CompanyQuorum:         math.LegacyNewDecWithPrec(5, 1),      // 50%
		CompanyThreshold:      math.LegacyNewDecWithPrec(67, 2),     // 67%
		ShareholderVoting:     true,
		ValidatorQuorum:       math.LegacyNewDecWithPrec(67, 2),     // 67%
		ValidatorThreshold:    math.LegacyNewDecWithPrec(75, 2),     // 75%
	}
}