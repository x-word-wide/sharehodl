package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InheritancePlanStatus represents the current status of an inheritance plan
type InheritancePlanStatus int32

const (
	PlanStatusActive InheritancePlanStatus = iota   // Active, monitoring for inactivity
	PlanStatusTriggered                             // Dead man switch triggered, in grace period
	PlanStatusExecuting                             // Claims being processed
	PlanStatusCompleted                             // All assets transferred
	PlanStatusCancelled                             // Cancelled by owner
)

func (s InheritancePlanStatus) String() string {
	switch s {
	case PlanStatusActive:
		return "active"
	case PlanStatusTriggered:
		return "triggered"
	case PlanStatusExecuting:
		return "executing"
	case PlanStatusCompleted:
		return "completed"
	case PlanStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// InheritancePlan defines a complete inheritance plan
type InheritancePlan struct {
	PlanID              uint64                    `json:"plan_id"`
	Owner               string                    `json:"owner"`
	Beneficiaries       []Beneficiary             `json:"beneficiaries"`
	InactivityPeriod    time.Duration             `json:"inactivity_period"`     // e.g., 365 days
	GracePeriod         time.Duration             `json:"grace_period"`          // e.g., 30 days minimum
	ClaimWindow         time.Duration             `json:"claim_window"`          // e.g., 180 days per beneficiary
	CharityAddress      string                    `json:"charity_address"`       // Where unclaimed assets go
	Status              InheritancePlanStatus     `json:"status"`
	CreatedAt           time.Time                 `json:"created_at"`
	UpdatedAt           time.Time                 `json:"updated_at"`
}

// Beneficiary defines a beneficiary in an inheritance plan
type Beneficiary struct {
	Address      string         `json:"address"`
	Priority     uint32         `json:"priority"`         // 1 = highest, lower number = first in line
	Percentage   math.LegacyDec `json:"percentage"`       // Share of assets (0.0-1.0)
	SpecificAssets []SpecificAsset `json:"specific_assets,omitempty"` // Specific asset allocations
}

// AssetType defines types of assets that can be inherited
type AssetType int32

const (
	AssetTypeHODL AssetType = iota      // Native HODL stablecoin
	AssetTypeEquity                     // Company shares
	AssetTypeCustomToken                // Custom tokens
	AssetTypeAllAssets                  // All assets (used for percentage-based allocation)
	AssetTypeStaked                     // Staked HODL
	AssetTypeLoanPosition               // Loan positions (borrower or lender)
	AssetTypeEscrowPosition             // Escrow positions (buyer or seller)
)

func (t AssetType) String() string {
	switch t {
	case AssetTypeHODL:
		return "hodl"
	case AssetTypeEquity:
		return "equity"
	case AssetTypeCustomToken:
		return "custom_token"
	case AssetTypeAllAssets:
		return "all_assets"
	case AssetTypeStaked:
		return "staked"
	case AssetTypeLoanPosition:
		return "loan_position"
	case AssetTypeEscrowPosition:
		return "escrow_position"
	default:
		return "unknown"
	}
}

// SpecificAsset defines a specific asset allocation
type SpecificAsset struct {
	AssetType AssetType `json:"asset_type"`
	// For AssetTypeHODL or AssetTypeCustomToken
	Denom     string    `json:"denom,omitempty"`
	Amount    math.Int  `json:"amount,omitempty"`
	// For AssetTypeEquity
	CompanyID uint64    `json:"company_id,omitempty"`
	ClassID   string    `json:"class_id,omitempty"`
	Shares    math.Int  `json:"shares,omitempty"`
}

// TriggerStatus represents the status of a switch trigger
type TriggerStatus int32

const (
	TriggerStatusActive TriggerStatus = iota  // Trigger is active, in grace period
	TriggerStatusExpired                      // Grace period expired, claims can begin
	TriggerStatusCancelled                    // Owner cancelled the trigger
)

func (s TriggerStatus) String() string {
	switch s {
	case TriggerStatusActive:
		return "active"
	case TriggerStatusExpired:
		return "expired"
	case TriggerStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// SwitchTrigger represents an activated dead man switch
type SwitchTrigger struct {
	PlanID              uint64        `json:"plan_id"`
	TriggeredAt         time.Time     `json:"triggered_at"`
	GracePeriodEnd      time.Time     `json:"grace_period_end"`
	Status              TriggerStatus `json:"status"`
	LastInactivityCheck time.Time     `json:"last_inactivity_check"`
}

// ClaimStatus represents the status of a beneficiary claim
type ClaimStatus int32

const (
	ClaimStatusPending ClaimStatus = iota  // Waiting for claim window
	ClaimStatusOpen                        // Claim window open
	ClaimStatusProcessing                  // Claim being processed (prevents double-claim)
	ClaimStatusClaimed                     // Assets claimed
	ClaimStatusExpired                     // Claim window expired
	ClaimStatusSkipped                     // Beneficiary skipped (banned/invalid)
)

func (s ClaimStatus) String() string {
	switch s {
	case ClaimStatusPending:
		return "pending"
	case ClaimStatusOpen:
		return "open"
	case ClaimStatusProcessing:
		return "processing"
	case ClaimStatusClaimed:
		return "claimed"
	case ClaimStatusExpired:
		return "expired"
	case ClaimStatusSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// BeneficiaryClaim tracks a beneficiary's claim on an inheritance
type BeneficiaryClaim struct {
	PlanID          uint64             `json:"plan_id"`
	Beneficiary     string             `json:"beneficiary"`
	Priority        uint32             `json:"priority"`
	Percentage      math.LegacyDec     `json:"percentage"`
	ClaimWindowStart time.Time         `json:"claim_window_start"`
	ClaimWindowEnd   time.Time         `json:"claim_window_end"`
	Status          ClaimStatus        `json:"status"`
	ClaimedAt       time.Time          `json:"claimed_at,omitempty"`
	TransferredAssets []TransferredAsset `json:"transferred_assets,omitempty"`
}

// TransferredAsset records an asset that was transferred
type TransferredAsset struct {
	AssetType AssetType `json:"asset_type"`
	// For tokens
	Denom     string    `json:"denom,omitempty"`
	Amount    math.Int  `json:"amount,omitempty"`
	// For equity
	CompanyID uint64    `json:"company_id,omitempty"`
	ClassID   string    `json:"class_id,omitempty"`
	Shares    math.Int  `json:"shares,omitempty"`
	// Metadata
	TransferredAt time.Time `json:"transferred_at"`
}

// ActivityRecord tracks the last activity of an address
type ActivityRecord struct {
	Address          string    `json:"address"`
	LastActivityTime time.Time `json:"last_activity_time"`
	LastBlockHeight  int64     `json:"last_block_height"`
	ActivityType     string    `json:"activity_type"` // "transaction", "sign", etc.
}

// LockedAssetDuringInheritance tracks assets locked during the inheritance process
type LockedAssetDuringInheritance struct {
	PlanID        uint64      `json:"plan_id"`
	Owner         string      `json:"owner"`
	LockedCoins   sdk.Coins   `json:"locked_coins"`
	LockedEquity  []EquityLock `json:"locked_equity"`
	LockedAt      time.Time   `json:"locked_at"`
}

// EquityLock represents locked equity shares
type EquityLock struct {
	CompanyID uint64   `json:"company_id"`
	ClassID   string   `json:"class_id"`
	Shares    math.Int `json:"shares"`
}

// Validation methods

// Validate validates an InheritancePlan
func (p InheritancePlan) Validate() error {
	if p.PlanID == 0 {
		return fmt.Errorf("plan ID cannot be zero")
	}
	if p.Owner == "" {
		return fmt.Errorf("owner cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(p.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if len(p.Beneficiaries) == 0 {
		return fmt.Errorf("must have at least one beneficiary")
	}
	if p.InactivityPeriod <= 0 {
		return fmt.Errorf("inactivity period must be positive")
	}
	if p.GracePeriod < 30*24*time.Hour {
		return fmt.Errorf("grace period must be at least 30 days")
	}
	if p.ClaimWindow <= 0 {
		return fmt.Errorf("claim window must be positive")
	}

	// Validate beneficiaries
	totalPercentage := math.LegacyZeroDec()
	priorityMap := make(map[uint32]bool)
	for _, b := range p.Beneficiaries {
		if err := b.Validate(); err != nil {
			return fmt.Errorf("invalid beneficiary: %w", err)
		}
		if priorityMap[b.Priority] {
			return fmt.Errorf("duplicate priority %d", b.Priority)
		}
		priorityMap[b.Priority] = true
		totalPercentage = totalPercentage.Add(b.Percentage)
	}

	// SECURITY FIX: Total percentage must equal exactly 100%
	if !totalPercentage.Equal(math.LegacyOneDec()) {
		return fmt.Errorf("total beneficiary percentage must equal 100%%, got: %s", totalPercentage)
	}

	return nil
}

// Validate validates a Beneficiary
func (b Beneficiary) Validate() error {
	if b.Address == "" {
		return fmt.Errorf("beneficiary address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(b.Address); err != nil {
		return fmt.Errorf("invalid beneficiary address: %v", err)
	}
	if b.Priority == 0 {
		return fmt.Errorf("priority must be positive")
	}
	if b.Percentage.IsNegative() || b.Percentage.GT(math.LegacyOneDec()) {
		return fmt.Errorf("percentage must be between 0 and 1: %s", b.Percentage)
	}

	// Validate specific assets
	for _, asset := range b.SpecificAssets {
		if err := asset.Validate(); err != nil {
			return fmt.Errorf("invalid specific asset: %w", err)
		}
	}

	return nil
}

// Validate validates a SpecificAsset
func (a SpecificAsset) Validate() error {
	switch a.AssetType {
	case AssetTypeHODL, AssetTypeCustomToken:
		if a.Denom == "" {
			return fmt.Errorf("denom cannot be empty for token asset")
		}
		if a.Amount.IsNil() || a.Amount.IsNegative() {
			return fmt.Errorf("amount must be non-negative")
		}
	case AssetTypeEquity:
		if a.CompanyID == 0 {
			return fmt.Errorf("company ID cannot be zero for equity asset")
		}
		if a.ClassID == "" {
			return fmt.Errorf("class ID cannot be empty for equity asset")
		}
		if a.Shares.IsNil() || a.Shares.IsNegative() {
			return fmt.Errorf("shares must be non-negative")
		}
	case AssetTypeAllAssets:
		// No specific validation needed
	default:
		return fmt.Errorf("invalid asset type: %d", a.AssetType)
	}
	return nil
}

// Validate validates a SwitchTrigger
func (t SwitchTrigger) Validate() error {
	if t.PlanID == 0 {
		return fmt.Errorf("plan ID cannot be zero")
	}
	if t.TriggeredAt.IsZero() {
		return fmt.Errorf("triggered at time cannot be zero")
	}
	if t.GracePeriodEnd.IsZero() {
		return fmt.Errorf("grace period end time cannot be zero")
	}
	if t.GracePeriodEnd.Before(t.TriggeredAt) {
		return fmt.Errorf("grace period end must be after triggered time")
	}
	return nil
}

// Validate validates a BeneficiaryClaim
func (c BeneficiaryClaim) Validate() error {
	if c.PlanID == 0 {
		return fmt.Errorf("plan ID cannot be zero")
	}
	if c.Beneficiary == "" {
		return fmt.Errorf("beneficiary cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(c.Beneficiary); err != nil {
		return fmt.Errorf("invalid beneficiary address: %v", err)
	}
	if c.Percentage.IsNegative() || c.Percentage.GT(math.LegacyOneDec()) {
		return fmt.Errorf("percentage must be between 0 and 1: %s", c.Percentage)
	}
	return nil
}

// ProtoMessage implementations for compatibility
func (m *InheritancePlan) ProtoMessage()                  {}
func (m *InheritancePlan) Reset()                         { *m = InheritancePlan{} }
func (m *InheritancePlan) String() string                 { return fmt.Sprintf("InheritancePlan{ID:%d,Owner:%s}", m.PlanID, m.Owner) }
func (m *Beneficiary) ProtoMessage()                      {}
func (m *Beneficiary) Reset()                             { *m = Beneficiary{} }
func (m *Beneficiary) String() string                     { return fmt.Sprintf("Beneficiary{Address:%s}", m.Address) }
func (m *SpecificAsset) ProtoMessage()                    {}
func (m *SpecificAsset) Reset()                           { *m = SpecificAsset{} }
func (m *SpecificAsset) String() string                   { return fmt.Sprintf("SpecificAsset{Type:%s}", m.AssetType) }
func (m *SwitchTrigger) ProtoMessage()                    {}
func (m *SwitchTrigger) Reset()                           { *m = SwitchTrigger{} }
func (m *SwitchTrigger) String() string                   { return fmt.Sprintf("SwitchTrigger{PlanID:%d}", m.PlanID) }
func (m *BeneficiaryClaim) ProtoMessage()                 {}
func (m *BeneficiaryClaim) Reset()                        { *m = BeneficiaryClaim{} }
func (m *BeneficiaryClaim) String() string                { return fmt.Sprintf("BeneficiaryClaim{PlanID:%d}", m.PlanID) }
func (m *TransferredAsset) ProtoMessage()                 {}
func (m *TransferredAsset) Reset()                        { *m = TransferredAsset{} }
func (m *TransferredAsset) String() string                { return "TransferredAsset{}" }
func (m *ActivityRecord) ProtoMessage()                   {}
func (m *ActivityRecord) Reset()                          { *m = ActivityRecord{} }
func (m *ActivityRecord) String() string                  { return fmt.Sprintf("ActivityRecord{Address:%s}", m.Address) }
func (m *LockedAssetDuringInheritance) ProtoMessage()     {}
func (m *LockedAssetDuringInheritance) Reset()            { *m = LockedAssetDuringInheritance{} }
func (m *LockedAssetDuringInheritance) String() string    { return fmt.Sprintf("LockedAsset{PlanID:%d}", m.PlanID) }
func (m *EquityLock) ProtoMessage()                       {}
func (m *EquityLock) Reset()                              { *m = EquityLock{} }
func (m *EquityLock) String() string                      { return "EquityLock{}" }
