package types

import (
	"time"

	"cosmossdk.io/math"
)

// ReputationAction defines actions that affect reputation
type ReputationAction int32

const (
	// Positive actions
	ActionSuccessfulVerification ReputationAction = iota
	ActionGoodVote                                // Voted with majority on passed proposal
	ActionUptimeBonus                             // Validator uptime bonus
	ActionSuccessfulLoan                          // Loan repaid on time
	ActionLiquidityProvided                       // Provided liquidity to DEX

	// Moderator actions
	ActionSuccessfulDispute    // Dispute resolved fairly (voted with majority)
	ActionUnfairDispute        // Dispute resolution overturned

	// Negative actions
	ActionFailedVerification   // Verification rejected
	ActionBadVote              // Consistently voting against community
	ActionDowntime             // Validator downtime
	ActionLoanDefault          // Failed to repay loan
	ActionSlashed              // Got slashed for misbehavior
	ActionFraudAttempt         // Attempted fraudulent activity
	ActionBadModeration        // Consistently unfair dispute resolution
)

// Reputation score changes (in basis points, 100 = 1 point)
const (
	SuccessfulVerificationBonus  int64 = 500  // +5 points
	GoodVoteBonus                int64 = 100  // +1 point
	UptimeBonusPerEpoch          int64 = 200  // +2 points per epoch with 100% uptime
	SuccessfulLoanBonus          int64 = 300  // +3 points
	LiquidityProvidedBonus       int64 = 150  // +1.5 points
	SuccessfulDisputeBonus       int64 = 400  // +4 points for fair dispute resolution

	FailedVerificationPenalty    int64 = -2000 // -20 points
	BadVotePenalty               int64 = -200  // -2 points
	DowntimePenalty              int64 = -1000 // -10 points
	LoanDefaultPenalty           int64 = -1500 // -15 points
	SlashedPenalty               int64 = -3000 // -30 points
	FraudAttemptPenalty          int64 = -5000 // -50 points
	UnfairDisputePenalty         int64 = -800  // -8 points for overturned resolution
	BadModerationPenalty         int64 = -2500 // -25 points for pattern of unfair moderation
)

// String returns the action name
func (a ReputationAction) String() string {
	switch a {
	case ActionSuccessfulVerification:
		return "successful_verification"
	case ActionGoodVote:
		return "good_vote"
	case ActionUptimeBonus:
		return "uptime_bonus"
	case ActionSuccessfulLoan:
		return "successful_loan"
	case ActionLiquidityProvided:
		return "liquidity_provided"
	case ActionSuccessfulDispute:
		return "successful_dispute"
	case ActionFailedVerification:
		return "failed_verification"
	case ActionBadVote:
		return "bad_vote"
	case ActionDowntime:
		return "downtime"
	case ActionLoanDefault:
		return "loan_default"
	case ActionSlashed:
		return "slashed"
	case ActionFraudAttempt:
		return "fraud_attempt"
	case ActionUnfairDispute:
		return "unfair_dispute"
	case ActionBadModeration:
		return "bad_moderation"
	default:
		return "unknown"
	}
}

// GetScoreChange returns the reputation score change for this action
func (a ReputationAction) GetScoreChange() math.LegacyDec {
	var change int64
	switch a {
	case ActionSuccessfulVerification:
		change = SuccessfulVerificationBonus
	case ActionGoodVote:
		change = GoodVoteBonus
	case ActionUptimeBonus:
		change = UptimeBonusPerEpoch
	case ActionSuccessfulLoan:
		change = SuccessfulLoanBonus
	case ActionLiquidityProvided:
		change = LiquidityProvidedBonus
	case ActionSuccessfulDispute:
		change = SuccessfulDisputeBonus
	case ActionFailedVerification:
		change = FailedVerificationPenalty
	case ActionBadVote:
		change = BadVotePenalty
	case ActionDowntime:
		change = DowntimePenalty
	case ActionLoanDefault:
		change = LoanDefaultPenalty
	case ActionSlashed:
		change = SlashedPenalty
	case ActionFraudAttempt:
		change = FraudAttemptPenalty
	case ActionUnfairDispute:
		change = UnfairDisputePenalty
	case ActionBadModeration:
		change = BadModerationPenalty
	default:
		change = 0
	}

	return math.LegacyNewDec(change).QuoInt64(100)
}

// IsPositive returns true if this action improves reputation
func (a ReputationAction) IsPositive() bool {
	return a.GetScoreChange().IsPositive()
}

// ReputationRecord tracks a single reputation change event
type ReputationRecord struct {
	Owner       string           `json:"owner"`
	Action      ReputationAction `json:"action"`
	ScoreChange math.LegacyDec   `json:"score_change"`
	NewScore    math.LegacyDec   `json:"new_score"`
	Reason      string           `json:"reason"`
	BlockHeight int64            `json:"block_height"`
	Timestamp   time.Time        `json:"timestamp"`
	RelatedTx   string           `json:"related_tx,omitempty"`
}

// ProtoMessage implements proto.Message interface
func (m *ReputationRecord) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *ReputationRecord) Reset() { *m = ReputationRecord{} }

// String implements proto.Message interface
func (m *ReputationRecord) String() string { return m.Owner }

// ReputationHistory stores reputation history for a user
type ReputationHistory struct {
	Owner        string             `json:"owner"`
	Records      []ReputationRecord `json:"records"`
	TotalGains   math.LegacyDec     `json:"total_gains"`
	TotalLosses  math.LegacyDec     `json:"total_losses"`
	HighestScore math.LegacyDec     `json:"highest_score"`
	LowestScore  math.LegacyDec     `json:"lowest_score"`
}

// ProtoMessage implements proto.Message interface
func (m *ReputationHistory) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *ReputationHistory) Reset() { *m = ReputationHistory{} }

// String implements proto.Message interface
func (m *ReputationHistory) String() string { return m.Owner }

// ReputationTier defines reputation-based access levels
type ReputationTier int32

const (
	ReputationPoor      ReputationTier = iota // 0-29: Restricted access
	ReputationLow                              // 30-49: Limited access
	ReputationMedium                           // 50-69: Standard access
	ReputationGood                             // 70-84: Enhanced access
	ReputationExcellent                        // 85-100: Full access
)

// GetReputationTier returns the reputation tier for a score
func GetReputationTier(score math.LegacyDec) ReputationTier {
	s := score.TruncateInt64()
	if s >= 85 {
		return ReputationExcellent
	} else if s >= 70 {
		return ReputationGood
	} else if s >= 50 {
		return ReputationMedium
	} else if s >= 30 {
		return ReputationLow
	}
	return ReputationPoor
}

// String returns the reputation tier name
func (t ReputationTier) String() string {
	switch t {
	case ReputationExcellent:
		return "Excellent"
	case ReputationGood:
		return "Good"
	case ReputationMedium:
		return "Medium"
	case ReputationLow:
		return "Low"
	case ReputationPoor:
		return "Poor"
	default:
		return "Unknown"
	}
}

// MinimumReputationRequirements defines minimum reputation for various actions
var MinimumReputationRequirements = map[string]int64{
	"lend":                 60, // Minimum reputation to lend
	"borrow":               50, // Minimum reputation to borrow
	"verify_small":         50, // Minimum to verify small businesses
	"verify_medium":        60, // Minimum to verify medium businesses
	"verify_large":         70, // Minimum to verify large businesses
	"verify_enterprise":    80, // Minimum to verify enterprise businesses
	"validator":            80, // Minimum to be a validator
	"governance_proposer":  70, // Minimum to submit governance proposals
	"large_trade":          40, // Minimum for large trades
	"moderate":             65, // Minimum reputation to be a moderator
	"moderate_large":       75, // Minimum to moderate large disputes (> 100K HODL)
	"slash_moderator":      80, // Minimum to slash other moderators
	"blacklist_moderator":  85, // Minimum to blacklist moderators
}

// MeetsReputationRequirement checks if a score meets the requirement
func MeetsReputationRequirement(score math.LegacyDec, action string) bool {
	required, exists := MinimumReputationRequirements[action]
	if !exists {
		return true // No requirement
	}
	return score.GTE(math.LegacyNewDec(required))
}
