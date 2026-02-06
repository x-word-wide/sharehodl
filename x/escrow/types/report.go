package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Report escalation constants
const (
	MaxReportExtensions  = 2                       // Maximum deadline extensions before escalation
	MaxReportEscalations = 3                       // Maximum tier escalations before governance
	ExtensionDuration    = 3 * 24 * time.Hour      // 3 days per extension
)

// ReportType represents the type of report being submitted
type ReportType int32

const (
	ReportTypeFraud               ReportType = iota // Fraudulent activity
	ReportTypeScam                                  // Scam company/listing
	ReportTypeModeratorMisconduct                   // Bad moderator behavior
	ReportTypeMarketManipulation                    // Trading manipulation
	ReportTypeCollusion                             // Moderator-party collusion
	ReportTypeWrongResolution                       // Wrong dispute resolution - user lost funds unfairly
)

func (t ReportType) String() string {
	switch t {
	case ReportTypeFraud:
		return "fraud"
	case ReportTypeScam:
		return "scam"
	case ReportTypeModeratorMisconduct:
		return "moderator_misconduct"
	case ReportTypeMarketManipulation:
		return "market_manipulation"
	case ReportTypeCollusion:
		return "collusion"
	case ReportTypeWrongResolution:
		return "wrong_resolution"
	default:
		return "unknown"
	}
}

// ReportStatus represents the current status of a report
type ReportStatus int32

const (
	ReportStatusOpen                  ReportStatus = iota // Newly submitted
	ReportStatusPendingVoluntaryReturn                    // Waiting for counterparty to voluntarily return funds
	ReportStatusVoluntarilyResolved                       // Counterparty returned funds voluntarily - no penalties
	ReportStatusUnderInvestigation                        // Assigned to reviewers (counterparty refused/ignored)
	ReportStatusConfirmed                                 // Fraud confirmed
	ReportStatusDismissed                                 // Report dismissed
	ReportStatusAppealed                                  // Reporter appealed dismissal
)

func (s ReportStatus) String() string {
	switch s {
	case ReportStatusOpen:
		return "open"
	case ReportStatusPendingVoluntaryReturn:
		return "pending_voluntary_return"
	case ReportStatusVoluntarilyResolved:
		return "voluntarily_resolved"
	case ReportStatusUnderInvestigation:
		return "under_investigation"
	case ReportStatusConfirmed:
		return "confirmed"
	case ReportStatusDismissed:
		return "dismissed"
	case ReportStatusAppealed:
		return "appealed"
	default:
		return "unknown"
	}
}

// ReportTargetType represents what entity is being reported
type ReportTargetType int32

const (
	ReportTargetTypeEscrow    ReportTargetType = iota // Specific escrow transaction
	ReportTargetTypeCompany                           // Listed company
	ReportTargetTypeModerator                         // Moderator
	ReportTargetTypeUser                              // User address
)

func (t ReportTargetType) String() string {
	switch t {
	case ReportTargetTypeEscrow:
		return "escrow"
	case ReportTargetTypeCompany:
		return "company"
	case ReportTargetTypeModerator:
		return "moderator"
	case ReportTargetTypeUser:
		return "user"
	default:
		return "unknown"
	}
}

// Report represents a user-submitted fraud/misconduct report
type Report struct {
	ID         uint64           `json:"id"`
	ReportType ReportType       `json:"report_type"`
	Reporter   string           `json:"reporter"` // Who filed the report (anonymous to target initially)
	TargetType ReportTargetType `json:"target_type"`
	TargetID   string           `json:"target_id"` // Entity ID being reported

	// Report details
	Reason      string     `json:"reason"`      // Description of the issue
	Evidence    []Evidence `json:"evidence"`    // Reuses existing Evidence type
	Priority    int        `json:"priority"`    // 1-5, calculated from reporter tier + evidence
	Severity    int        `json:"severity"`    // User-estimated severity 1-5

	// Investigation
	Status            ReportStatus `json:"status"`
	AssignedReviewers []string     `json:"assigned_reviewers"` // Validators assigned
	VotesRequired     int          `json:"votes_required"`     // Based on priority
	ReviewVotes       []ReviewVote `json:"review_votes"`       // Reviewer votes

	// Resolution
	Resolution       string    `json:"resolution"`        // Final resolution description
	ActionsTaken     []string  `json:"actions_taken"`     // List of actions executed
	ResolvedBy       string    `json:"resolved_by"`       // Who finalized the resolution

	// Timeline
	CreatedAt  time.Time `json:"created_at"`
	DeadlineAt time.Time `json:"deadline_at"` // Investigation deadline
	ResolvedAt time.Time `json:"resolved_at"`

	// Voluntary Return (for WrongResolution reports)
	// Gives counterparty a chance to return funds OR reject the claim
	Counterparty            string    `json:"counterparty,omitempty"`              // Who received funds (Bob)
	VoluntaryReturnDeadline time.Time `json:"voluntary_return_deadline,omitempty"` // 48-hour grace period
	AmountToReturn          math.Int  `json:"amount_to_return,omitempty"`          // How much counterparty should return
	VoluntaryReturnComplete bool      `json:"voluntary_return_complete"`           // Did they return voluntarily?
	CounterpartyRejected    bool      `json:"counterparty_rejected"`               // Did counterparty reject the claim?
	CounterpartyResponse    string    `json:"counterparty_response,omitempty"`     // Bob's response/reasoning

	// Reporter reward/penalty tracking
	ReporterReputationChange int      `json:"reporter_reputation_change"` // +2000 confirmed, -1000 false
	ReporterRewardAmount     math.Int `json:"reporter_reward_amount"`     // 10% of recovered funds (max 1000 HODL)

	// Escalation tracking (prevents false negatives)
	EscalationCount   int          `json:"escalation_count"`              // How many times escalated
	ExtensionCount    int          `json:"extension_count"`               // How many times deadline extended
	OriginalTier      int          `json:"original_tier"`                 // Starting reviewer tier
	CurrentTier       int          `json:"current_tier"`                  // Current reviewer tier
	EscalatedAt       time.Time    `json:"escalated_at,omitempty"`        // When last escalated
	PreviousVotes     []ReviewVote `json:"previous_votes,omitempty"`      // Votes from lower tiers

	// Evidence locking (Issue #5)
	EvidenceSnapshot  string    `json:"evidence_snapshot,omitempty"`   // Hash of evidence when first vote cast
	EvidenceLockedAt  time.Time `json:"evidence_locked_at,omitempty"`  // When evidence was locked
}

// ReviewVote represents a validator's vote on a report
type ReviewVote struct {
	Reviewer    string    `json:"reviewer"`    // Validator address
	Tier        int       `json:"tier"`        // Reviewer's staking tier
	Confirmed   bool      `json:"confirmed"`   // true = fraud confirmed
	Comments    string    `json:"comments"`    // Reasoning
	VotedAt     time.Time `json:"voted_at"`
}

// Appeal represents an appeal of a dismissed report
type Appeal struct {
	ID                uint64            `json:"id"`
	DisputeID         uint64            `json:"dispute_id,omitempty"`  // For dispute appeals
	ReportID          uint64            `json:"report_id,omitempty"`   // For report appeals
	Appellant         string            `json:"appellant"`             // Who filed the appeal
	AppealReason      string            `json:"appeal_reason"`
	NewEvidence       []Evidence        `json:"new_evidence"`
	AppealLevel       int               `json:"appeal_level"`      // 1 = first, 2 = escalated, 3 = final
	RequiredTier      int               `json:"required_tier"`     // Minimum reviewer tier
	ReviewerCount     int               `json:"reviewer_count"`    // Required number of reviewers
	AssignedReviewers []string          `json:"assigned_reviewers"` // Validators assigned to review
	Votes             []AppealVote      `json:"votes"`
	Status            AppealStatus      `json:"status"`
	OriginalResolution DisputeResolution `json:"original_resolution"`
	NewResolution     DisputeResolution `json:"new_resolution"`
	CreatedAt         time.Time         `json:"created_at"`
	DeadlineAt        time.Time         `json:"deadline_at"`
	ResolvedAt        time.Time         `json:"resolved_at"`

	// Evidence locking (Issue #5)
	EvidenceSnapshot  string    `json:"evidence_snapshot,omitempty"`   // Hash of evidence when first vote cast
	EvidenceLockedAt  time.Time `json:"evidence_locked_at,omitempty"`  // When evidence was locked
}

// AppealStatus represents the status of an appeal
type AppealStatus int32

const (
	AppealStatusOpen      AppealStatus = iota // Appeal submitted
	AppealStatusReviewing                     // Under review
	AppealStatusUpheld                        // Original decision upheld
	AppealStatusOverturned                    // Original decision overturned
	AppealStatusEscalated                     // Escalated to higher tier
)

func (s AppealStatus) String() string {
	switch s {
	case AppealStatusOpen:
		return "open"
	case AppealStatusReviewing:
		return "reviewing"
	case AppealStatusUpheld:
		return "upheld"
	case AppealStatusOverturned:
		return "overturned"
	case AppealStatusEscalated:
		return "escalated"
	default:
		return "unknown"
	}
}

// AppealVote represents a reviewer's vote on an appeal
type AppealVote struct {
	Reviewer       string            `json:"reviewer"`
	Tier           int               `json:"tier"`
	UpholdOriginal bool              `json:"uphold_original"` // true = keep original decision
	NewResolution  DisputeResolution `json:"new_resolution"`  // If overturning
	Reasoning      string            `json:"reasoning"`
	VotedAt        time.Time         `json:"voted_at"`
}

// ModeratorMetrics tracks moderator decision quality for accountability
type ModeratorMetrics struct {
	Address              string         `json:"address"`
	TotalDecisions       uint64         `json:"total_decisions"`
	OverturnedDecisions  uint64         `json:"overturned_decisions"`
	OverturnRate         math.LegacyDec `json:"overturn_rate"` // overturned / total
	ConsecutiveOverturns int            `json:"consecutive_overturns"` // Reset on upheld
	LastOverturnAt       time.Time      `json:"last_overturn_at"`
	ReportsAgainst       uint64         `json:"reports_against"`   // Reports filed against
	ConfirmedReports     uint64         `json:"confirmed_reports"` // Reports confirmed true
	WarningCount         int            `json:"warning_count"`     // Warnings before action
	LastWarningAt        time.Time      `json:"last_warning_at"`
}

// EscrowReserve tracks the reserve fund for compensating wronged users
type EscrowReserve struct {
	TotalFunds   math.Int `json:"total_funds"`   // Funded by 0.1% of escrow fees
	PendingClaims []ReserveClaim `json:"pending_claims"`
	TotalPaidOut math.Int `json:"total_paid_out"`
}

// ReserveClaim represents a claim against the escrow reserve
type ReserveClaim struct {
	ID           uint64    `json:"id"`
	Claimant     string    `json:"claimant"`
	DisputeID    uint64    `json:"dispute_id"`
	AppealID     uint64    `json:"appeal_id"`
	Amount       math.Int  `json:"amount"`
	Reason       string    `json:"reason"`
	Status       string    `json:"status"` // pending, approved, paid, rejected
	CreatedAt    time.Time `json:"created_at"`
	ProcessedAt  time.Time `json:"processed_at"`
}

// Validation methods

func (r Report) Validate() error {
	if r.Reporter == "" {
		return fmt.Errorf("reporter cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(r.Reporter); err != nil {
		return fmt.Errorf("invalid reporter address: %v", err)
	}
	if r.TargetID == "" {
		return fmt.Errorf("target ID cannot be empty")
	}
	if r.Reason == "" {
		return fmt.Errorf("report reason cannot be empty")
	}
	if r.Severity < 1 || r.Severity > 5 {
		return fmt.Errorf("severity must be between 1 and 5")
	}
	return nil
}

func (a Appeal) Validate() error {
	if a.Appellant == "" {
		return fmt.Errorf("appellant cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(a.Appellant); err != nil {
		return fmt.Errorf("invalid appellant address: %v", err)
	}
	if a.DisputeID == 0 && a.ReportID == 0 {
		return fmt.Errorf("appeal must reference a dispute or report")
	}
	if a.AppealReason == "" {
		return fmt.Errorf("appeal reason cannot be empty")
	}
	if a.AppealLevel < 1 || a.AppealLevel > 3 {
		return fmt.Errorf("appeal level must be between 1 and 3")
	}
	return nil
}

// Helper constructors

// NewReport creates a new report with calculated priority
func NewReport(
	id uint64,
	reportType ReportType,
	reporter string,
	targetType ReportTargetType,
	targetID string,
	reason string,
	severity int,
	reporterTier int,
	blockTime time.Time,
) Report {
	// Calculate priority based on reporter tier + severity
	// Higher tier reporters and higher severity = higher priority
	priority := CalculateReportPriority(reporterTier, severity, 0)

	// Determine votes required and deadline based on priority
	votesRequired, deadlineDays := getReportRequirements(priority)

	// Determine starting tier for reviewers
	requiredTier := getRequiredReviewerTierFromPriority(priority)

	return Report{
		ID:            id,
		ReportType:    reportType,
		Reporter:      reporter,
		TargetType:    targetType,
		TargetID:      targetID,
		Reason:        reason,
		Severity:      severity,
		Priority:      priority,
		Status:        ReportStatusOpen,
		VotesRequired: votesRequired,
		Evidence:      []Evidence{},
		ReviewVotes:   []ReviewVote{},
		ActionsTaken:  []string{},
		CreatedAt:     blockTime,
		DeadlineAt:    blockTime.Add(time.Duration(deadlineDays) * 24 * time.Hour),
		ReporterRewardAmount: math.ZeroInt(),
		// Escalation tracking
		EscalationCount: 0,
		ExtensionCount:  0,
		OriginalTier:    requiredTier,
		CurrentTier:     requiredTier,
		PreviousVotes:   []ReviewVote{},
	}
}

// getRequiredReviewerTierFromPriority returns minimum tier based on priority
func getRequiredReviewerTierFromPriority(priority int) int {
	switch priority {
	case 5:
		return 4 // Archon for emergency
	case 4:
		return 3 // Steward for high priority
	default:
		return 2 // Warden for standard
	}
}

// CalculateReportPriority calculates report priority
// Priority = ReporterTier + Severity + (EvidenceCount / 2)
func CalculateReportPriority(reporterTier, severity, evidenceCount int) int {
	priority := reporterTier + severity + (evidenceCount / 2)

	// Clamp to 1-5 range
	if priority < 1 {
		priority = 1
	}
	if priority > 5 {
		priority = 5
	}

	return priority
}

// getReportRequirements returns votes required and deadline days based on priority
func getReportRequirements(priority int) (votesRequired int, deadlineDays int) {
	switch priority {
	case 5:
		return 7, 3 // Emergency: 7 reviewers, 3 days
	case 4:
		return 5, 5 // High: 5 reviewers, 5 days
	case 3:
		return 5, 5 // Medium: 5 reviewers, 5 days
	default:
		return 3, 7 // Low: 3 reviewers, 7 days
	}
}

// NewAppeal creates a new appeal
func NewAppeal(
	id uint64,
	disputeID uint64,
	reportID uint64,
	appellant string,
	reason string,
	appealLevel int,
	originalResolution DisputeResolution,
	blockTime time.Time,
) Appeal {
	requiredTier, reviewerCount, deadlineDays := getAppealRequirements(appealLevel)

	return Appeal{
		ID:                 id,
		DisputeID:          disputeID,
		ReportID:           reportID,
		Appellant:          appellant,
		AppealReason:       reason,
		AppealLevel:        appealLevel,
		RequiredTier:       requiredTier,
		ReviewerCount:      reviewerCount,
		Status:             AppealStatusOpen,
		OriginalResolution: originalResolution,
		NewEvidence:        []Evidence{},
		Votes:              []AppealVote{},
		CreatedAt:          blockTime,
		DeadlineAt:         blockTime.Add(time.Duration(deadlineDays) * 24 * time.Hour),
	}
}

// getAppealRequirements returns tier requirement, reviewer count, and deadline based on level
func getAppealRequirements(appealLevel int) (requiredTier int, reviewerCount int, deadlineDays int) {
	switch appealLevel {
	case 1: // Standard appeal
		return 2, 5, 5 // Warden+, 5 voters, 5 days
	case 2: // Escalated appeal
		return 3, 7, 7 // Steward+, 7 voters, 7 days
	case 3: // Final appeal
		return 4, 9, 10 // Archon+, 9 voters, 10 days
	default:
		return 2, 5, 5
	}
}

// NewModeratorMetrics creates initial metrics for a moderator
func NewModeratorMetrics(address string) ModeratorMetrics {
	return ModeratorMetrics{
		Address:              address,
		TotalDecisions:       0,
		OverturnedDecisions:  0,
		OverturnRate:         math.LegacyZeroDec(),
		ConsecutiveOverturns: 0,
		ReportsAgainst:       0,
		ConfirmedReports:     0,
		WarningCount:         0,
	}
}

// UpdateOverturnRate recalculates the overturn rate
func (m *ModeratorMetrics) UpdateOverturnRate() {
	if m.TotalDecisions == 0 {
		m.OverturnRate = math.LegacyZeroDec()
		return
	}
	m.OverturnRate = math.LegacyNewDec(int64(m.OverturnedDecisions)).Quo(
		math.LegacyNewDec(int64(m.TotalDecisions)),
	)
}

// RecordDecision updates metrics when a decision is made
func (m *ModeratorMetrics) RecordDecision(overturned bool, blockTime time.Time) {
	m.TotalDecisions++
	if overturned {
		m.OverturnedDecisions++
		m.ConsecutiveOverturns++
		m.LastOverturnAt = blockTime
	} else {
		m.ConsecutiveOverturns = 0
	}
	m.UpdateOverturnRate()
}

// ShouldAutoBlacklist checks if moderator should be automatically blacklisted
func (m ModeratorMetrics) ShouldAutoBlacklist() (bool, string) {
	// 3+ overturns triggers temporary blacklist
	if m.ConsecutiveOverturns >= 3 {
		return true, fmt.Sprintf("3+ consecutive overturned decisions (%d)", m.ConsecutiveOverturns)
	}

	// 2+ confirmed fraud reports triggers permanent blacklist
	if m.ConfirmedReports >= 2 {
		return true, fmt.Sprintf("2+ confirmed fraud reports (%d)", m.ConfirmedReports)
	}

	// 30%+ overturn rate with 10+ decisions triggers warning/review
	if m.TotalDecisions >= 10 && m.OverturnRate.GTE(math.LegacyNewDecWithPrec(30, 2)) {
		return true, fmt.Sprintf("overturn rate %.2f%% exceeds 30%% threshold", m.OverturnRate.MulInt64(100).MustFloat64())
	}

	return false, ""
}

// NewEscrowReserve creates an empty reserve
func NewEscrowReserve() EscrowReserve {
	return EscrowReserve{
		TotalFunds:    math.ZeroInt(),
		PendingClaims: []ReserveClaim{},
		TotalPaidOut:  math.ZeroInt(),
	}
}

// =============================================================================
// REPORTER HISTORY - Track report abuse and spam
// =============================================================================

// ReporterHistory tracks a user's reporting behavior for anti-spam
type ReporterHistory struct {
	Address               string         `json:"address"`
	TotalReports          uint64         `json:"total_reports"`
	ConfirmedReports      uint64         `json:"confirmed_reports"`      // Reports that were validated
	DismissedReports      uint64         `json:"dismissed_reports"`      // Reports that were dismissed
	ConsecutiveDismissed  int            `json:"consecutive_dismissed"`  // Reset on confirmed report
	FalseReportRate       math.LegacyDec `json:"false_report_rate"`      // dismissed / total
	LastReportAt          time.Time      `json:"last_report_at"`
	LastDismissedAt       time.Time      `json:"last_dismissed_at"`

	// Penalty tracking
	TotalReputationLost   int64          `json:"total_reputation_lost"`  // Cumulative reputation penalties
	TotalStakeSlashed     math.Int       `json:"total_stake_slashed"`    // Cumulative stake slashed
	WarningCount          int            `json:"warning_count"`          // Warnings issued

	// Ban status
	IsBanned              bool           `json:"is_banned"`
	BanReason             string         `json:"ban_reason,omitempty"`
	BannedAt              time.Time      `json:"banned_at,omitempty"`
	BanExpiresAt          time.Time      `json:"ban_expires_at,omitempty"` // Zero = permanent
	BanCount              int            `json:"ban_count"`               // How many times banned
}

// NewReporterHistory creates a new reporter history
func NewReporterHistory(address string) ReporterHistory {
	return ReporterHistory{
		Address:              address,
		TotalReports:         0,
		ConfirmedReports:     0,
		DismissedReports:     0,
		ConsecutiveDismissed: 0,
		FalseReportRate:      math.LegacyZeroDec(),
		TotalReputationLost:  0,
		TotalStakeSlashed:    math.ZeroInt(),
		WarningCount:         0,
		IsBanned:             false,
		BanCount:             0,
	}
}

// RecordReport updates history when a report is resolved
func (h *ReporterHistory) RecordReport(confirmed bool, blockTime time.Time) {
	h.TotalReports++
	h.LastReportAt = blockTime

	if confirmed {
		h.ConfirmedReports++
		h.ConsecutiveDismissed = 0 // Reset on success
	} else {
		h.DismissedReports++
		h.ConsecutiveDismissed++
		h.LastDismissedAt = blockTime
	}

	// Update false report rate
	if h.TotalReports > 0 {
		h.FalseReportRate = math.LegacyNewDec(int64(h.DismissedReports)).Quo(
			math.LegacyNewDec(int64(h.TotalReports)),
		)
	}
}

// GetEscalatedPenalty calculates escalating penalty based on history
// Returns: (reputationPenalty, slashPercent, shouldBan, banDuration)
func (h ReporterHistory) GetEscalatedPenalty(reportType ReportType) (int, math.LegacyDec, bool, time.Duration) {
	basePenalty := 1000
	slashPercent := math.LegacyZeroDec()
	shouldBan := false
	banDuration := time.Duration(0)

	// WrongResolution reports have higher base penalty (attempted theft)
	if reportType == ReportTypeWrongResolution {
		basePenalty = 3000
		slashPercent = math.LegacyNewDecWithPrec(5, 2) // 5% base slash
	}

	// Escalate based on consecutive dismissals
	switch h.ConsecutiveDismissed {
	case 1:
		// First offense: base penalty + warning
		// No escalation yet
	case 2:
		// Second offense: 1.5x penalty + warning
		basePenalty = basePenalty * 3 / 2
		if reportType == ReportTypeWrongResolution {
			slashPercent = math.LegacyNewDecWithPrec(10, 2) // 10% slash
		}
	case 3:
		// Third offense: 2x penalty + temporary ban (7 days)
		basePenalty = basePenalty * 2
		slashPercent = math.LegacyNewDecWithPrec(15, 2) // 15% slash
		shouldBan = true
		banDuration = 7 * 24 * time.Hour
	case 4:
		// Fourth offense: 3x penalty + longer ban (30 days)
		basePenalty = basePenalty * 3
		slashPercent = math.LegacyNewDecWithPrec(25, 2) // 25% slash
		shouldBan = true
		banDuration = 30 * 24 * time.Hour
	default:
		if h.ConsecutiveDismissed >= 5 {
			// 5+ offenses: permanent ban + max slash
			basePenalty = basePenalty * 5
			slashPercent = math.LegacyNewDecWithPrec(50, 2) // 50% slash
			shouldBan = true
			banDuration = 0 // 0 = permanent
		}
	}

	// Additional escalation based on overall false report rate
	// If >50% of reports are false with 5+ reports, increase penalty
	if h.TotalReports >= 5 && h.FalseReportRate.GT(math.LegacyNewDecWithPrec(50, 2)) {
		basePenalty = basePenalty * 3 / 2 // Additional 50%

		// If >80% false with 10+ reports, that's systematic abuse
		if h.TotalReports >= 10 && h.FalseReportRate.GT(math.LegacyNewDecWithPrec(80, 2)) {
			shouldBan = true
			if banDuration > 0 || banDuration == 0 {
				banDuration = 0 // Permanent ban for systematic abusers
			}
		}
	}

	return basePenalty, slashPercent, shouldBan, banDuration
}

// ShouldBlockReport checks if this reporter should be blocked from submitting
func (h ReporterHistory) ShouldBlockReport(blockTime time.Time) (bool, string) {
	// Check if permanently banned
	if h.IsBanned && h.BanExpiresAt.IsZero() {
		return true, fmt.Sprintf("permanently banned for abuse: %s", h.BanReason)
	}

	// Check if temporarily banned
	if h.IsBanned && blockTime.Before(h.BanExpiresAt) {
		remaining := h.BanExpiresAt.Sub(blockTime)
		return true, fmt.Sprintf("banned for %v (reason: %s)", remaining.Round(time.Hour), h.BanReason)
	}

	// Rate limiting: max 3 reports per day
	if h.TotalReports > 0 {
		timeSinceLastReport := blockTime.Sub(h.LastReportAt)
		reportsLast24h := h.countRecentReports(blockTime, 24*time.Hour)
		if reportsLast24h >= 3 && timeSinceLastReport < time.Hour {
			return true, "rate limited: maximum 3 reports per day"
		}
	}

	// If user has 3+ consecutive dismissed reports, require cooldown
	if h.ConsecutiveDismissed >= 3 {
		cooldownPeriod := time.Duration(h.ConsecutiveDismissed) * 24 * time.Hour
		timeSinceLastDismissed := blockTime.Sub(h.LastDismissedAt)
		if timeSinceLastDismissed < cooldownPeriod {
			remaining := cooldownPeriod - timeSinceLastDismissed
			return true, fmt.Sprintf("cooldown: %v remaining after %d dismissed reports",
				remaining.Round(time.Hour), h.ConsecutiveDismissed)
		}
	}

	return false, ""
}

// countRecentReports is a placeholder - in production would track timestamps
func (h ReporterHistory) countRecentReports(blockTime time.Time, window time.Duration) int {
	// Simplified: check if last report was within window
	if blockTime.Sub(h.LastReportAt) < window {
		return 1 // At least 1 recent report
	}
	return 0
}

// IsPermanentlyBanned checks if user is permanently banned
func (h ReporterHistory) IsPermanentlyBanned() bool {
	return h.IsBanned && h.BanExpiresAt.IsZero()
}

// Ban applies a ban to the reporter
func (h *ReporterHistory) Ban(reason string, duration time.Duration, blockTime time.Time) {
	h.IsBanned = true
	h.BanReason = reason
	h.BannedAt = blockTime
	h.BanCount++

	if duration > 0 {
		h.BanExpiresAt = blockTime.Add(duration)
	} else {
		h.BanExpiresAt = time.Time{} // Zero = permanent
	}
}

// Unban removes the ban (for expired temporary bans)
func (h *ReporterHistory) Unban() {
	h.IsBanned = false
	h.BanReason = ""
}

// =============================================================================
// TIERED INVESTIGATION SYSTEM - Multi-tier voting for fraud freeze authority
// =============================================================================

// InvestigationStatus represents the current phase of a company investigation
type InvestigationStatus int32

const (
	InvestigationStatusPreliminary    InvestigationStatus = iota // Report filed, no action yet
	InvestigationStatusWardenReview                              // Awaiting Warden review (48hr)
	InvestigationStatusStewardReview                             // Escalated to Steward review (72hr)
	InvestigationStatusFreezeApproved                            // Freeze approved, pending 24hr warning
	InvestigationStatusFrozen                                    // Treasury frozen, trading halted
	InvestigationStatusCleared                                   // Investigation cleared, no fraud
)

func (s InvestigationStatus) String() string {
	switch s {
	case InvestigationStatusPreliminary:
		return "preliminary"
	case InvestigationStatusWardenReview:
		return "warden_review"
	case InvestigationStatusStewardReview:
		return "steward_review"
	case InvestigationStatusFreezeApproved:
		return "freeze_approved"
	case InvestigationStatusFrozen:
		return "frozen"
	case InvestigationStatusCleared:
		return "cleared"
	default:
		return "unknown"
	}
}

// TierVote represents a vote by a specific tier (Warden, Steward, or Archon)
type TierVote struct {
	Voter     string    `json:"voter"`
	Tier      int       `json:"tier"`       // Validator tier (Warden=2, Steward=3, Archon=4)
	Approve   bool      `json:"approve"`    // true = escalate/freeze, false = clear
	Reason    string    `json:"reason"`     // Voting rationale
	VotedAt   time.Time `json:"voted_at"`
}

// CompanyInvestigation tracks the multi-tier voting process for fraud investigations
type CompanyInvestigation struct {
	ID        uint64              `json:"id"`
	CompanyID uint64              `json:"company_id"`
	ReportID  uint64              `json:"report_id"`
	Status    InvestigationStatus `json:"status"`

	// Warden Review Phase (Tier 2: 10K HODL stake)
	// 3 Wardens vote, 2/3 required to escalate to Steward review
	WardenVotes    []TierVote `json:"warden_votes"`
	WardenDeadline time.Time  `json:"warden_deadline"`
	WardenApproved bool       `json:"warden_approved"`

	// Steward Review Phase (Tier 3: 50K HODL stake)
	// 5 Stewards vote, 3/5 required to approve freeze
	StewardVotes    []TierVote `json:"steward_votes"`
	StewardDeadline time.Time  `json:"steward_deadline"`
	StewardApproved bool       `json:"steward_approved"`

	// Warning Phase
	// 24-hour notice before freeze takes effect
	WarningIssuedAt  time.Time `json:"warning_issued_at,omitempty"`
	WarningExpiresAt time.Time `json:"warning_expires_at,omitempty"` // 24 hours after issued
	CompanyNotified  bool      `json:"company_notified"`

	// Timeline
	CreatedAt  time.Time `json:"created_at"`
	FrozenAt   time.Time `json:"frozen_at,omitempty"`
	ResolvedAt time.Time `json:"resolved_at,omitempty"`
}

// NewCompanyInvestigation creates a new company investigation
func NewCompanyInvestigation(id, companyID, reportID uint64, blockTime time.Time) CompanyInvestigation {
	return CompanyInvestigation{
		ID:              id,
		CompanyID:       companyID,
		ReportID:        reportID,
		Status:          InvestigationStatusPreliminary,
		WardenVotes:     []TierVote{},
		WardenDeadline:  blockTime.Add(48 * time.Hour), // 48 hours for Warden review
		WardenApproved:  false,
		StewardVotes:    []TierVote{},
		StewardDeadline: time.Time{}, // Set when escalated to Steward review
		StewardApproved: false,
		CompanyNotified: false,
		CreatedAt:       blockTime,
	}
}

// Validate validates a company investigation
func (i CompanyInvestigation) Validate() error {
	if i.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if i.ReportID == 0 {
		return fmt.Errorf("report ID cannot be zero")
	}
	return nil
}

// CountWardenVotes returns (approveCount, rejectCount)
func (i CompanyInvestigation) CountWardenVotes() (int, int) {
	approve := 0
	reject := 0
	for _, vote := range i.WardenVotes {
		if vote.Approve {
			approve++
		} else {
			reject++
		}
	}
	return approve, reject
}

// CountStewardVotes returns (approveCount, rejectCount)
func (i CompanyInvestigation) CountStewardVotes() (int, int) {
	approve := 0
	reject := 0
	for _, vote := range i.StewardVotes {
		if vote.Approve {
			approve++
		} else {
			reject++
		}
	}
	return approve, reject
}

// HasVoted checks if a voter has already voted in the current phase
func (i CompanyInvestigation) HasVoted(voter string) bool {
	// Check current phase votes
	if i.Status == InvestigationStatusWardenReview {
		for _, vote := range i.WardenVotes {
			if vote.Voter == voter {
				return true
			}
		}
	} else if i.Status == InvestigationStatusStewardReview {
		for _, vote := range i.StewardVotes {
			if vote.Voter == voter {
				return true
			}
		}
	}
	return false
}
