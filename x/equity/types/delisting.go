package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DelistingReason represents why a company was delisted
type DelistingReason int32

const (
	DelistingReasonVoluntary      DelistingReason = iota // Company chose to delist
	DelistingReasonFraud                                  // Fraud confirmed by report investigation
	DelistingReasonScam                                   // Scam company confirmed
	DelistingReasonGovernance                             // Governance proposal passed
	DelistingReasonRegulatoryIssue                        // Regulatory compliance failure
	DelistingReasonAbandonment                            // Company abandoned (no activity)
)

func (r DelistingReason) String() string {
	switch r {
	case DelistingReasonVoluntary:
		return "voluntary"
	case DelistingReasonFraud:
		return "fraud"
	case DelistingReasonScam:
		return "scam"
	case DelistingReasonGovernance:
		return "governance"
	case DelistingReasonRegulatoryIssue:
		return "regulatory_issue"
	case DelistingReasonAbandonment:
		return "abandonment"
	default:
		return "unknown"
	}
}

// RequiresCompensation returns true if shareholders should be compensated
func (r DelistingReason) RequiresCompensation() bool {
	switch r {
	case DelistingReasonFraud, DelistingReasonScam:
		return true
	default:
		return false
	}
}

// CompensationStatus represents the status of a delisting compensation process
type CompensationStatus int32

const (
	CompensationStatusPending    CompensationStatus = iota // Waiting for claims
	CompensationStatusProcessing                            // Processing claims
	CompensationStatusCompleted                             // All claims processed
	CompensationStatusExpired                               // Claim window expired
)

func (s CompensationStatus) String() string {
	switch s {
	case CompensationStatusPending:
		return "pending"
	case CompensationStatusProcessing:
		return "processing"
	case CompensationStatusCompleted:
		return "completed"
	case CompensationStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// DelistingCompensation tracks the compensation pool for a delisted company
type DelistingCompensation struct {
	ID                   uint64             `json:"id"`
	CompanyID            uint64             `json:"company_id"`
	CompanySymbol        string             `json:"company_symbol"`
	DelistingReason      DelistingReason    `json:"delisting_reason"`
	Status               CompensationStatus `json:"status"`

	// Total shares that need compensation
	TotalShares          math.Int           `json:"total_shares"`

	// Compensation pool breakdown (in uhodl)
	TreasuryContribution math.Int           `json:"treasury_contribution"` // From frozen company treasury
	VerifierContribution math.Int           `json:"verifier_contribution"` // From slashed verifier stakes
	CommunityContribution math.Int          `json:"community_contribution"` // From community pool (backup)
	TotalPool            math.Int           `json:"total_pool"`

	// Per-share compensation rate
	CompensationPerShare math.LegacyDec     `json:"compensation_per_share"`

	// Claim tracking
	TotalClaims          uint64             `json:"total_claims"`
	ProcessedClaims      uint64             `json:"processed_claims"`
	ClaimedAmount        math.Int           `json:"claimed_amount"`
	RemainingPool        math.Int           `json:"remaining_pool"`
	TotalClaimedShares   math.Int           `json:"total_claimed_shares"` // Track total shares claimed

	// Timeline
	CreatedAt            time.Time          `json:"created_at"`
	ClaimDeadline        time.Time          `json:"claim_deadline"`        // Legacy: 90 days (now using ClaimWindowEnd)
	ClaimWindowStart     time.Time          `json:"claim_window_start"`    // When claims can start
	ClaimWindowEnd       time.Time          `json:"claim_window_end"`      // When claim window closes (30 days)
	DistributionAt       time.Time          `json:"distribution_at"`       // When pro-rata distribution happens
	IsDistributed        bool               `json:"is_distributed"`        // Whether distribution has occurred
	CompletedAt          time.Time          `json:"completed_at,omitempty"`

	// Associated report (if triggered by fraud report)
	ReportID             uint64             `json:"report_id,omitempty"`
}

// Validate validates a DelistingCompensation
func (c DelistingCompensation) Validate() error {
	if c.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if c.TotalShares.IsNil() || c.TotalShares.IsNegative() {
		return fmt.Errorf("total shares cannot be negative")
	}
	if c.TotalPool.IsNil() || c.TotalPool.IsNegative() {
		return fmt.Errorf("total pool cannot be negative")
	}
	if c.ClaimDeadline.Before(c.CreatedAt) {
		return fmt.Errorf("claim deadline must be after creation time")
	}
	return nil
}

// IsClaimable returns true if the compensation is still claimable
func (c DelistingCompensation) IsClaimable(currentTime time.Time) bool {
	// Check if we're within the claim window and not yet distributed
	return c.Status == CompensationStatusPending &&
		currentTime.After(c.ClaimWindowStart) &&
		currentTime.Before(c.ClaimWindowEnd) &&
		!c.IsDistributed
}

// CompensationClaim represents a shareholder's claim for compensation
type CompensationClaim struct {
	ID               uint64    `json:"id"`
	CompensationID   uint64    `json:"compensation_id"`
	CompanyID        uint64    `json:"company_id"`
	Claimant         string    `json:"claimant"`

	// Claim details
	SharesHeld       math.Int  `json:"shares_held"`   // Shares held at delisting
	ShareClass       string    `json:"share_class"`   // Which class of shares
	ClaimAmount      math.Int  `json:"claim_amount"`  // Calculated compensation

	// Status
	Status           string    `json:"status"` // pending, approved, paid, rejected
	ProcessedAt      time.Time `json:"processed_at,omitempty"`
	TransactionHash  string    `json:"transaction_hash,omitempty"` // Payment tx hash

	// Timeline
	CreatedAt        time.Time `json:"created_at"`
}

// Validate validates a CompensationClaim
func (c CompensationClaim) Validate() error {
	if c.CompensationID == 0 {
		return fmt.Errorf("compensation ID cannot be zero")
	}
	if c.Claimant == "" {
		return fmt.Errorf("claimant cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(c.Claimant); err != nil {
		return fmt.Errorf("invalid claimant address: %v", err)
	}
	if c.SharesHeld.IsNil() || c.SharesHeld.LTE(math.ZeroInt()) {
		return fmt.Errorf("shares held must be positive")
	}
	if c.ClaimAmount.IsNil() || c.ClaimAmount.IsNegative() {
		return fmt.Errorf("claim amount cannot be negative")
	}
	return nil
}

// TradingHalt represents a temporary trading halt for a company
type TradingHalt struct {
	CompanyID        uint64    `json:"company_id"`
	HaltedAt         time.Time `json:"halted_at"`
	Reason           string    `json:"reason"`
	HaltedBy         string    `json:"halted_by"`     // Who initiated the halt
	ReportID         uint64    `json:"report_id,omitempty"` // Associated report if any
	ExpectedDuration string    `json:"expected_duration"`   // e.g., "7 days" or "indefinite"
	ResumeAt         time.Time `json:"resume_at,omitempty"` // When trading will resume
}

// FrozenTreasury represents frozen company funds during investigation
type FrozenTreasury struct {
	CompanyID        uint64    `json:"company_id"`
	FrozenAmount     math.Int  `json:"frozen_amount"`
	Currency         string    `json:"currency"`
	FrozenAt         time.Time `json:"frozen_at"`
	Reason           string    `json:"reason"`
	ReportID         uint64    `json:"report_id,omitempty"`
	ReleasedAt       time.Time `json:"released_at,omitempty"`
	ReleasedTo       string    `json:"released_to,omitempty"` // "compensation_pool" or "company"
}

// VerifierPenalty tracks penalties applied to verifiers who approved a fraudulent listing
type VerifierPenalty struct {
	ID               uint64    `json:"id"`
	CompanyID        uint64    `json:"company_id"`
	VerifierAddress  string    `json:"verifier_address"`
	OriginalStake    math.Int  `json:"original_stake"`
	SlashAmount      math.Int  `json:"slash_amount"`
	SlashPercent     math.LegacyDec `json:"slash_percent"` // e.g., 0.50 for 50%
	Reason           string    `json:"reason"`
	AppliedAt        time.Time `json:"applied_at"`
}

// NewDelistingCompensation creates a new compensation record
func NewDelistingCompensation(
	id uint64,
	companyID uint64,
	companySymbol string,
	reason DelistingReason,
	totalShares math.Int,
	blockTime time.Time,
) DelistingCompensation {
	claimWindowEnd := blockTime.Add(30 * 24 * time.Hour) // 30-day claim window
	distributionAt := claimWindowEnd.Add(24 * time.Hour) // Distribute 1 day after window closes
	claimDeadline := blockTime.Add(90 * 24 * time.Hour) // Legacy field for backward compatibility

	return DelistingCompensation{
		ID:                    id,
		CompanyID:             companyID,
		CompanySymbol:         companySymbol,
		DelistingReason:       reason,
		Status:                CompensationStatusPending,
		TotalShares:           totalShares,
		TreasuryContribution:  math.ZeroInt(),
		VerifierContribution:  math.ZeroInt(),
		CommunityContribution: math.ZeroInt(),
		TotalPool:             math.ZeroInt(),
		CompensationPerShare:  math.LegacyZeroDec(),
		TotalClaims:           0,
		ProcessedClaims:       0,
		ClaimedAmount:         math.ZeroInt(),
		RemainingPool:         math.ZeroInt(),
		TotalClaimedShares:    math.ZeroInt(),
		CreatedAt:             blockTime,
		ClaimDeadline:         claimDeadline,
		ClaimWindowStart:      blockTime,
		ClaimWindowEnd:        claimWindowEnd,
		DistributionAt:        distributionAt,
		IsDistributed:         false,
	}
}

// AddTreasuryContribution adds funds from company treasury
func (c *DelistingCompensation) AddTreasuryContribution(amount math.Int) {
	c.TreasuryContribution = c.TreasuryContribution.Add(amount)
	c.TotalPool = c.TotalPool.Add(amount)
	c.RemainingPool = c.RemainingPool.Add(amount)
	c.recalculatePerShare()
}

// AddVerifierContribution adds funds from slashed verifier stakes
func (c *DelistingCompensation) AddVerifierContribution(amount math.Int) {
	c.VerifierContribution = c.VerifierContribution.Add(amount)
	c.TotalPool = c.TotalPool.Add(amount)
	c.RemainingPool = c.RemainingPool.Add(amount)
	c.recalculatePerShare()
}

// AddCommunityContribution adds funds from community pool
func (c *DelistingCompensation) AddCommunityContribution(amount math.Int) {
	c.CommunityContribution = c.CommunityContribution.Add(amount)
	c.TotalPool = c.TotalPool.Add(amount)
	c.RemainingPool = c.RemainingPool.Add(amount)
	c.recalculatePerShare()
}

// recalculatePerShare recalculates the compensation per share
func (c *DelistingCompensation) recalculatePerShare() {
	if c.TotalShares.IsPositive() {
		c.CompensationPerShare = math.LegacyNewDecFromInt(c.TotalPool).QuoInt(c.TotalShares)
	}
}

// CalculateClaimAmount calculates compensation for a given share count
func (c DelistingCompensation) CalculateClaimAmount(shares math.Int) math.Int {
	if shares.IsNil() || shares.IsNegative() || c.CompensationPerShare.IsZero() {
		return math.ZeroInt()
	}
	return c.CompensationPerShare.MulInt(shares).TruncateInt()
}

// ProcessClaim updates the compensation after a claim is processed
func (c *DelistingCompensation) ProcessClaim(claimAmount math.Int) {
	c.ProcessedClaims++
	c.ClaimedAmount = c.ClaimedAmount.Add(claimAmount)
	c.RemainingPool = c.RemainingPool.Sub(claimAmount)
}

// NewCompensationClaim creates a new claim
func NewCompensationClaim(
	id uint64,
	compensationID uint64,
	companyID uint64,
	claimant string,
	shareClass string,
	sharesHeld math.Int,
	compensationPerShare math.LegacyDec,
	blockTime time.Time,
) CompensationClaim {
	claimAmount := compensationPerShare.MulInt(sharesHeld).TruncateInt()

	return CompensationClaim{
		ID:             id,
		CompensationID: compensationID,
		CompanyID:      companyID,
		Claimant:       claimant,
		ShareClass:     shareClass,
		SharesHeld:     sharesHeld,
		ClaimAmount:    claimAmount,
		Status:         "pending",
		CreatedAt:      blockTime,
	}
}

// NewTradingHalt creates a new trading halt
func NewTradingHalt(
	companyID uint64,
	reason string,
	haltedBy string,
	reportID uint64,
	duration time.Duration,
	blockTime time.Time,
) TradingHalt {
	var resumeAt time.Time
	var durationStr string

	if duration > 0 {
		resumeAt = blockTime.Add(duration)
		durationStr = duration.String()
	} else {
		durationStr = "indefinite"
	}

	return TradingHalt{
		CompanyID:        companyID,
		HaltedAt:         blockTime,
		Reason:           reason,
		HaltedBy:         haltedBy,
		ReportID:         reportID,
		ExpectedDuration: durationStr,
		ResumeAt:         resumeAt,
	}
}

// FreezeWarningStatus represents the status of a freeze warning
type FreezeWarningStatus int32

const (
	FreezeWarningStatusPending   FreezeWarningStatus = iota // Warning active
	FreezeWarningStatusResponded                            // Company responded
	FreezeWarningStatusExecuted                             // Freeze executed (no response or invalid)
	FreezeWarningStatusEscalated                            // Escalated to Archon
	FreezeWarningStatusCleared                              // Cleared, no freeze
)

func (s FreezeWarningStatus) String() string {
	switch s {
	case FreezeWarningStatusPending:
		return "pending"
	case FreezeWarningStatusResponded:
		return "responded"
	case FreezeWarningStatusExecuted:
		return "executed"
	case FreezeWarningStatusEscalated:
		return "escalated"
	case FreezeWarningStatusCleared:
		return "cleared"
	default:
		return "unknown"
	}
}

// FreezeWarning represents a 24-hour warning before company treasury freeze
type FreezeWarning struct {
	ID                uint64              `json:"id"`
	CompanyID         uint64              `json:"company_id"`
	InvestigationID   uint64              `json:"investigation_id"` // Links to CompanyInvestigation
	ReportID          uint64              `json:"report_id"`

	// Warning details
	Status            FreezeWarningStatus `json:"status"`
	WarningIssuedAt   time.Time           `json:"warning_issued_at"`
	WarningExpiresAt  time.Time           `json:"warning_expires_at"` // 24 hours after issued

	// Company response
	CompanyResponded      bool        `json:"company_responded"`
	ResponseSubmittedAt   time.Time   `json:"response_submitted_at,omitempty"`
	CounterEvidence       []Evidence  `json:"counter_evidence,omitempty"` // Reuse Evidence type
	ResponseText          string      `json:"response_text,omitempty"`

	// Outcome
	FreezeExecuted    bool      `json:"freeze_executed"`
	EscalatedToArchon bool      `json:"escalated_to_archon"`
	ResolvedAt        time.Time `json:"resolved_at,omitempty"`
	Resolution        string    `json:"resolution,omitempty"`
}

// Validate validates a FreezeWarning
func (w FreezeWarning) Validate() error {
	if w.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if w.ReportID == 0 {
		return fmt.Errorf("report ID cannot be zero")
	}
	if w.WarningExpiresAt.Before(w.WarningIssuedAt) {
		return fmt.Errorf("warning expiration must be after issuance")
	}
	return nil
}

// IsExpired checks if the warning period has expired
func (w FreezeWarning) IsExpired(currentTime time.Time) bool {
	return currentTime.After(w.WarningExpiresAt)
}

// Evidence represents evidence submitted (reusing for consistency with report system)
type Evidence struct {
	Submitter   string    `json:"submitter"`
	Hash        string    `json:"hash"`
	Description string    `json:"description"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// Event types for delisting
const (
	EventTypeTradingHalted           = "trading_halted"
	EventTypeTradingResumed          = "trading_resumed"
	EventTypeTreasuryFrozen          = "treasury_frozen"
	EventTypeTreasuryReleased        = "treasury_released"
	EventTypeCompanyDelisted         = "company_delisted"
	EventTypeCompensationCreated     = "compensation_created"
	EventTypeCompensationClaimed     = "compensation_claimed"
	EventTypeCompensationCompleted   = "compensation_completed"
	EventTypeVerifierPenalized       = "verifier_penalized"
	EventTypeFreezeWarningIssued     = "freeze_warning_issued"
	EventTypeFreezeWarningResponse   = "freeze_warning_response"
	EventTypeFreezeExecuted          = "freeze_executed"
	EventTypeFreezeEscalated         = "freeze_escalated"
	EventTypeFreezeWarningCleared    = "freeze_warning_cleared"
	EventTypeResponderAuthorized     = "responder_authorized"  // Issue #11
	EventTypeResponderRevoked        = "responder_revoked"     // Issue #11
)

// Attribute keys for delisting events
const (
	AttributeKeyCompanyID         = "company_id"
	AttributeKeyCompanySymbol     = "company_symbol"
	AttributeKeyDelistingReason   = "delisting_reason"
	AttributeKeyHaltReason        = "halt_reason"
	AttributeKeyFrozenAmount      = "frozen_amount"
	AttributeKeyCompensationID    = "compensation_id"
	AttributeKeyTotalPool         = "total_pool"
	AttributeKeyClaimAmount       = "claim_amount"
	AttributeKeyClaimant          = "claimant"
	AttributeKeySharesHeld        = "shares_held"
	AttributeKeyVerifier          = "verifier"
	AttributeKeySlashAmount       = "slash_amount"
	AttributeKeySlashPercent      = "slash_percent"
	AttributeKeyReason            = "reason"
	AttributeKeyAlertType         = "alert_type"
	AttributeKeyWarningID         = "warning_id"
	AttributeKeyExpiresAt         = "expires_at"
	AttributeKeyMessage           = "message"
)

// EventTypeShareholderAlert is emitted when shareholders need to be notified
const (
	EventTypeShareholderAlert = "shareholder_alert"
)

// ShareholderAlert represents an alert for a shareholder about their holdings
type ShareholderAlert struct {
	CompanyID     uint64    `json:"company_id"`
	CompanySymbol string    `json:"company_symbol"`
	AlertType     string    `json:"alert_type"` // "fraud_investigation_warning", "trading_halted", etc.
	WarningID     uint64    `json:"warning_id,omitempty"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
	Message       string    `json:"message"`
	SharesHeld    math.Int  `json:"shares_held"`
	ClassID       string    `json:"class_id"`
}

// TreasuryWithdrawalLimit restricts withdrawals during freeze warning period
type TreasuryWithdrawalLimit struct {
	CompanyID       uint64    `json:"company_id"`
	MaxWithdrawal   math.Int  `json:"max_withdrawal"`   // Max single withdrawal (10% of treasury)
	DailyLimit      math.Int  `json:"daily_limit"`      // Max daily withdrawal (20% of treasury)
	WithdrawnToday  math.Int  `json:"withdrawn_today"`  // Track daily withdrawals
	LastResetDate   time.Time `json:"last_reset_date"`  // When daily counter was reset
	WarningID       uint64    `json:"warning_id"`       // Associated warning
	CreatedAt       time.Time `json:"created_at"`
}

// Validate validates a TreasuryWithdrawalLimit
func (t TreasuryWithdrawalLimit) Validate() error {
	if t.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if t.MaxWithdrawal.IsNil() || t.MaxWithdrawal.IsNegative() {
		return fmt.Errorf("max withdrawal cannot be negative")
	}
	if t.DailyLimit.IsNil() || t.DailyLimit.IsNegative() {
		return fmt.Errorf("daily limit cannot be negative")
	}
	if t.WithdrawnToday.IsNil() || t.WithdrawnToday.IsNegative() {
		return fmt.Errorf("withdrawn today cannot be negative")
	}
	if t.WarningID == 0 {
		return fmt.Errorf("warning ID cannot be zero")
	}
	return nil
}
