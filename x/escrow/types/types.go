package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EscrowStatus represents the status of an escrow
type EscrowStatus int32

const (
	EscrowStatusPending    EscrowStatus = iota // Escrow created, waiting for funding
	EscrowStatusFunded                         // Escrow fully funded
	EscrowStatusReleased                       // Funds released to recipient
	EscrowStatusRefunded                       // Funds refunded to sender
	EscrowStatusDisputed                       // Escrow in dispute
	EscrowStatusResolved                       // Dispute resolved
	EscrowStatusCancelled                      // Escrow cancelled
	EscrowStatusExpired                        // Escrow expired without action
)

func (s EscrowStatus) String() string {
	switch s {
	case EscrowStatusPending:
		return "pending"
	case EscrowStatusFunded:
		return "funded"
	case EscrowStatusReleased:
		return "released"
	case EscrowStatusRefunded:
		return "refunded"
	case EscrowStatusDisputed:
		return "disputed"
	case EscrowStatusResolved:
		return "resolved"
	case EscrowStatusCancelled:
		return "cancelled"
	case EscrowStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// DisputeStatus represents the status of a dispute
type DisputeStatus int32

const (
	DisputeStatusOpen       DisputeStatus = iota // Dispute opened
	DisputeStatusUnderReview                     // Moderator reviewing
	DisputeStatusVoting                          // Moderators voting
	DisputeStatusResolved                        // Resolution decided
	DisputeStatusAppealed                        // Resolution appealed
	DisputeStatusFinal                           // Final resolution (no more appeals)
)

func (s DisputeStatus) String() string {
	switch s {
	case DisputeStatusOpen:
		return "open"
	case DisputeStatusUnderReview:
		return "under_review"
	case DisputeStatusVoting:
		return "voting"
	case DisputeStatusResolved:
		return "resolved"
	case DisputeStatusAppealed:
		return "appealed"
	case DisputeStatusFinal:
		return "final"
	default:
		return "unknown"
	}
}

// DisputeResolution represents how a dispute was resolved
type DisputeResolution int32

const (
	DisputeResolutionNone          DisputeResolution = iota // No resolution yet
	DisputeResolutionReleaseBuyer                           // Release funds to buyer
	DisputeResolutionReleaseSeller                          // Release funds to seller
	DisputeResolutionSplit                                  // Split funds between parties
	DisputeResolutionRefund                                 // Full refund to sender
)

func (r DisputeResolution) String() string {
	switch r {
	case DisputeResolutionNone:
		return "none"
	case DisputeResolutionReleaseBuyer:
		return "release_to_buyer"
	case DisputeResolutionReleaseSeller:
		return "release_to_seller"
	case DisputeResolutionSplit:
		return "split"
	case DisputeResolutionRefund:
		return "refund"
	default:
		return "unknown"
	}
}

// AssetType represents the type of asset in escrow
type AssetType int32

const (
	AssetTypeHODL   AssetType = iota // HODL tokens
	AssetTypeEquity                  // Equity shares
	AssetTypeMixed                   // Both HODL and equity
)

// Escrow represents an escrow agreement
type Escrow struct {
	ID           uint64         `json:"id"`
	Sender       string         `json:"sender"`        // Address of the sender (buyer in P2P trade)
	Recipient    string         `json:"recipient"`     // Address of the recipient (seller in P2P trade)
	Moderator    string         `json:"moderator"`     // Optional moderator for dispute resolution

	// Assets in escrow
	Assets       []EscrowAsset  `json:"assets"`
	TotalValue   math.LegacyDec `json:"total_value"`   // Total value in HODL

	// Escrow terms
	Description  string         `json:"description"`   // Description of the escrow agreement
	Terms        string         `json:"terms"`         // Terms and conditions

	// Status
	Status       EscrowStatus   `json:"status"`

	// Timeline
	CreatedAt    time.Time      `json:"created_at"`
	FundedAt     time.Time      `json:"funded_at"`
	ExpiresAt    time.Time      `json:"expires_at"`
	CompletedAt  time.Time      `json:"completed_at"`

	// Fees
	EscrowFee    math.LegacyDec `json:"escrow_fee"`    // Fee for escrow service
	ModeratorFee math.LegacyDec `json:"moderator_fee"` // Fee for moderator (if used)

	// Conditions for automatic release
	Conditions   []EscrowCondition `json:"conditions"`

	// Signatures
	SenderConfirmed    bool `json:"sender_confirmed"`
	RecipientConfirmed bool `json:"recipient_confirmed"`
}

// EscrowAsset represents an asset held in escrow
type EscrowAsset struct {
	AssetType   AssetType      `json:"asset_type"`
	Denom       string         `json:"denom"`        // Token denomination (e.g., "hodl", "AAPL")
	Amount      math.Int       `json:"amount"`
	CompanyID   uint64         `json:"company_id"`   // For equity assets
	ShareClass  string         `json:"share_class"`  // For equity assets
}

// EscrowCondition represents a condition for automatic release
type EscrowCondition struct {
	Type        ConditionType  `json:"type"`
	Description string         `json:"description"`
	Satisfied   bool           `json:"satisfied"`
	SatisfiedAt time.Time      `json:"satisfied_at"`
	Evidence    string         `json:"evidence"`     // Hash of evidence
}

// ConditionType represents types of escrow conditions
type ConditionType int32

const (
	ConditionTypeManual      ConditionType = iota // Manual confirmation required
	ConditionTypeTimelock                         // Release after time period
	ConditionTypeSignature                        // Release after multi-sig
	ConditionTypeOracle                           // Release based on oracle data
	ConditionTypeDelivery                         // Release after delivery confirmation
)

// Dispute represents a dispute on an escrow
type Dispute struct {
	ID           uint64         `json:"id"`
	EscrowID     uint64         `json:"escrow_id"`
	Initiator    string         `json:"initiator"`     // Who opened the dispute
	Reason       string         `json:"reason"`        // Reason for dispute

	// Status
	Status       DisputeStatus  `json:"status"`
	Resolution   DisputeResolution `json:"resolution"`

	// Evidence
	Evidence     []Evidence     `json:"evidence"`

	// Moderator votes
	Votes        []ModeratorVote `json:"votes"`
	VotesRequired int           `json:"votes_required"` // Number of moderator votes needed

	// Split amounts (if resolution is split)
	SenderAmount    math.LegacyDec `json:"sender_amount"`
	RecipientAmount math.LegacyDec `json:"recipient_amount"`

	// Timeline
	CreatedAt    time.Time      `json:"created_at"`
	DeadlineAt   time.Time      `json:"deadline_at"`   // Deadline for moderator decision
	ResolvedAt   time.Time      `json:"resolved_at"`

	// Appeals
	AppealsCount int            `json:"appeals_count"`
	MaxAppeals   int            `json:"max_appeals"`   // Maximum number of appeals allowed
}

// Evidence represents evidence submitted in a dispute
type Evidence struct {
	Submitter   string    `json:"submitter"`
	Hash        string    `json:"hash"`        // IPFS or other hash of evidence
	Description string    `json:"description"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// ModeratorVote represents a moderator's vote on a dispute
type ModeratorVote struct {
	Moderator   string            `json:"moderator"`
	Vote        DisputeResolution `json:"vote"`
	Reason      string            `json:"reason"`
	VotedAt     time.Time         `json:"voted_at"`
}

// Moderator represents a registered moderator
type Moderator struct {
	Address         string         `json:"address"`
	Stake           math.Int       `json:"stake"`            // Staked HODL as collateral
	ReputationScore math.LegacyDec `json:"reputation_score"` // 0-100 reputation
	DisputesHandled uint64         `json:"disputes_handled"`
	SuccessRate     math.LegacyDec `json:"success_rate"` // % of decisions upheld
	Active          bool           `json:"active"`
	RegisteredAt    time.Time      `json:"registered_at"`

	// Tiers (kept for reputation tracking, but not used for trust ceiling)
	Tier           ModeratorTier  `json:"tier"`
	MaxEscrowValue math.LegacyDec `json:"max_escrow_value"` // Deprecated: use Stake as trust ceiling

	// Stake-as-Trust-Ceiling: Active dispute tracking
	ActiveDisputeValue math.LegacyDec `json:"active_dispute_value"` // Total value of active disputes
	ActiveDisputeCount uint64         `json:"active_dispute_count"` // Number of active disputes

	// Unbonding state
	UnbondingStake math.Int `json:"unbonding_stake"` // Amount currently in unbonding

	// Blacklist state
	Blacklisted   bool      `json:"blacklisted"`
	BlacklistedAt time.Time `json:"blacklisted_at,omitempty"`
}

// ModeratorTier represents moderator experience levels
type ModeratorTier int32

const (
	ModeratorTierBronze   ModeratorTier = iota // New moderators
	ModeratorTierSilver                        // 10+ disputes, 80%+ success
	ModeratorTierGold                          // 50+ disputes, 90%+ success
	ModeratorTierPlatinum                      // 200+ disputes, 95%+ success
)

func (t ModeratorTier) String() string {
	switch t {
	case ModeratorTierBronze:
		return "bronze"
	case ModeratorTierSilver:
		return "silver"
	case ModeratorTierGold:
		return "gold"
	case ModeratorTierPlatinum:
		return "platinum"
	default:
		return "unknown"
	}
}

// UnbondingStatus represents the status of a moderator's unbonding request
type UnbondingStatus int32

const (
	UnbondingStatusActive    UnbondingStatus = iota // Unbonding in progress
	UnbondingStatusCancelled                        // Unbonding cancelled by moderator
	UnbondingStatusCompleted                        // Unbonding completed, stake returned
	UnbondingStatusSlashed                          // Stake slashed during unbonding
)

func (s UnbondingStatus) String() string {
	switch s {
	case UnbondingStatusActive:
		return "active"
	case UnbondingStatusCancelled:
		return "cancelled"
	case UnbondingStatusCompleted:
		return "completed"
	case UnbondingStatusSlashed:
		return "slashed"
	default:
		return "unknown"
	}
}

// UnbondingModerator tracks moderators in the 14-day unbonding period
type UnbondingModerator struct {
	Address         string          `json:"address"`
	UnbondingAmount math.Int        `json:"unbonding_amount"`
	InitiatedAt     time.Time       `json:"initiated_at"`
	CompletesAt     time.Time       `json:"completes_at"` // 14 days from initiated_at
	Status          UnbondingStatus `json:"status"`
}

// ModeratorBlacklist tracks banned moderators
type ModeratorBlacklist struct {
	Address     string    `json:"address"`
	BannedBy    string    `json:"banned_by"` // Validator address who banned
	BannedAt    time.Time `json:"banned_at"`
	Reason      string    `json:"reason"`
	SlashAmount math.Int  `json:"slash_amount"` // Amount slashed when banned
	Permanent   bool      `json:"permanent"`
	UnbanAt     time.Time `json:"unban_at,omitempty"` // For temporary bans
}

// ValidatorAction logs validator oversight actions for audit trail
type ValidatorAction struct {
	ID            uint64    `json:"id"`
	Validator     string    `json:"validator"`
	Action        string    `json:"action"` // "slash", "blacklist", "unblacklist"
	Target        string    `json:"target"` // Moderator address
	Reason        string    `json:"reason"`
	Amount        math.Int  `json:"amount,omitempty"` // For slash actions
	Timestamp     time.Time `json:"timestamp"`
}

// Validation methods

func (e Escrow) Validate() error {
	if e.Sender == "" {
		return fmt.Errorf("sender cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(e.Sender); err != nil {
		return fmt.Errorf("invalid sender address: %v", err)
	}
	if e.Recipient == "" {
		return fmt.Errorf("recipient cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(e.Recipient); err != nil {
		return fmt.Errorf("invalid recipient address: %v", err)
	}
	if e.Sender == e.Recipient {
		return fmt.Errorf("sender and recipient cannot be the same")
	}
	if len(e.Assets) == 0 {
		return fmt.Errorf("escrow must have at least one asset")
	}
	for _, asset := range e.Assets {
		if err := asset.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a EscrowAsset) Validate() error {
	if a.Denom == "" {
		return fmt.Errorf("asset denom cannot be empty")
	}
	if a.Amount.IsNil() || a.Amount.LTE(math.ZeroInt()) {
		return fmt.Errorf("asset amount must be positive")
	}
	return nil
}

func (d Dispute) Validate() error {
	if d.EscrowID == 0 {
		return fmt.Errorf("escrow ID cannot be zero")
	}
	if d.Initiator == "" {
		return fmt.Errorf("initiator cannot be empty")
	}
	if d.Reason == "" {
		return fmt.Errorf("dispute reason cannot be empty")
	}
	return nil
}

func (m Moderator) Validate() error {
	if m.Address == "" {
		return fmt.Errorf("moderator address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Address); err != nil {
		return fmt.Errorf("invalid moderator address: %v", err)
	}
	if m.Stake.IsNil() || m.Stake.LTE(math.ZeroInt()) {
		return fmt.Errorf("moderator stake must be positive")
	}
	return nil
}

// NewEscrow creates a new escrow
// Note: Use ctx.BlockTime() for the blockTime parameter in production
func NewEscrow(
	id uint64,
	sender, recipient string,
	assets []EscrowAsset,
	description, terms string,
	expiresAt time.Time,
	blockTime time.Time, // Use ctx.BlockTime() from sdk.Context
) Escrow {
	// Calculate total value
	totalValue := math.LegacyZeroDec()
	for _, asset := range assets {
		if asset.AssetType == AssetTypeHODL {
			totalValue = totalValue.Add(math.LegacyNewDecFromInt(asset.Amount))
		}
	}

	return Escrow{
		ID:          id,
		Sender:      sender,
		Recipient:   recipient,
		Assets:      assets,
		TotalValue:  totalValue,
		Description: description,
		Terms:       terms,
		Status:      EscrowStatusPending,
		CreatedAt:   blockTime,
		ExpiresAt:   expiresAt,
		EscrowFee:   math.LegacyNewDecWithPrec(1, 2),   // 1% default fee
		ModeratorFee: math.LegacyNewDecWithPrec(5, 3),  // 0.5% moderator fee
	}
}

// NewDispute creates a new dispute
// Note: Use ctx.BlockTime() for the blockTime parameter in production
func NewDispute(id, escrowID uint64, initiator, reason string, deadline time.Time, blockTime time.Time) Dispute {
	return Dispute{
		ID:            id,
		EscrowID:      escrowID,
		Initiator:     initiator,
		Reason:        reason,
		Status:        DisputeStatusOpen,
		Resolution:    DisputeResolutionNone,
		Evidence:      []Evidence{},
		Votes:         []ModeratorVote{},
		VotesRequired: 3, // Default: need 3 moderator votes
		CreatedAt:     blockTime,
		DeadlineAt:    deadline,
		MaxAppeals:    1, // Default: 1 appeal allowed
	}
}

// NewModerator creates a new moderator
// Note: Use ctx.BlockTime() for the blockTime parameter in production
func NewModerator(address string, stake math.Int, blockTime time.Time) Moderator {
	return Moderator{
		Address:         address,
		Stake:           stake,
		ReputationScore: math.LegacyNewDec(50), // Start at 50
		DisputesHandled: 0,
		SuccessRate:     math.LegacyZeroDec(),
		Active:          true,
		RegisteredAt:    blockTime,
		Tier:            ModeratorTierBronze,
		MaxEscrowValue:  math.LegacyNewDecFromInt(stake), // Trust ceiling = stake

		// Initialize stake-as-trust-ceiling fields
		ActiveDisputeValue: math.LegacyZeroDec(),
		ActiveDisputeCount: 0,
		UnbondingStake:     math.ZeroInt(),
		Blacklisted:        false,
	}
}

// NewUnbondingModerator creates a new unbonding record
func NewUnbondingModerator(address string, amount math.Int, initiatedAt time.Time) UnbondingModerator {
	return UnbondingModerator{
		Address:         address,
		UnbondingAmount: amount,
		InitiatedAt:     initiatedAt,
		CompletesAt:     initiatedAt.Add(14 * 24 * time.Hour), // 14 days
		Status:          UnbondingStatusActive,
	}
}

// NewModeratorBlacklist creates a new blacklist entry
// Note: Use ctx.BlockTime() for the blockTime parameter in production
func NewModeratorBlacklist(address, bannedBy, reason string, permanent bool, banDuration time.Duration, blockTime time.Time) ModeratorBlacklist {
	bl := ModeratorBlacklist{
		Address:     address,
		BannedBy:    bannedBy,
		BannedAt:    blockTime,
		Reason:      reason,
		SlashAmount: math.ZeroInt(),
		Permanent:   permanent,
	}
	if !permanent && banDuration > 0 {
		bl.UnbanAt = blockTime.Add(banDuration)
	}
	return bl
}

// GetAvailableStake returns the stake available for new disputes (total - unbonding)
func (m Moderator) GetAvailableStake() math.Int {
	if m.Stake.LT(m.UnbondingStake) {
		return math.ZeroInt()
	}
	return m.Stake.Sub(m.UnbondingStake)
}

// GetTrustCeiling returns the maximum dispute value this moderator can handle
// Trust ceiling = available stake - currently committed dispute value
func (m Moderator) GetTrustCeiling() math.LegacyDec {
	availableStake := math.LegacyNewDecFromInt(m.GetAvailableStake())
	if availableStake.LT(m.ActiveDisputeValue) {
		return math.LegacyZeroDec()
	}
	return availableStake.Sub(m.ActiveDisputeValue)
}

// CanHandleDispute checks if moderator can handle a dispute of given value
func (m Moderator) CanHandleDispute(disputeValue math.LegacyDec) bool {
	if !m.Active || m.Blacklisted {
		return false
	}
	return m.GetTrustCeiling().GTE(disputeValue)
}

// IsUnbondingComplete checks if the unbonding period has passed
func (u UnbondingModerator) IsUnbondingComplete(now time.Time) bool {
	return now.After(u.CompletesAt) || now.Equal(u.CompletesAt)
}

// IsBanExpired checks if a temporary ban has expired
func (b ModeratorBlacklist) IsBanExpired(now time.Time) bool {
	if b.Permanent {
		return false
	}
	return now.After(b.UnbanAt) || now.Equal(b.UnbanAt)
}
