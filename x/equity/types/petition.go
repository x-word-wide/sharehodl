package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PetitionType represents different types of shareholder petitions
type PetitionType int32

const (
	PetitionTypeFraudConcern PetitionType = iota
	PetitionTypeUnusualActivity
	PetitionTypeManagementMisconduct
)

func (pt PetitionType) String() string {
	switch pt {
	case PetitionTypeFraudConcern:
		return "fraud_concern"
	case PetitionTypeUnusualActivity:
		return "unusual_activity"
	case PetitionTypeManagementMisconduct:
		return "management_misconduct"
	default:
		return "unknown"
	}
}

// PetitionStatus represents the current status of a petition
type PetitionStatus int32

const (
	PetitionStatusOpen PetitionStatus = iota
	PetitionStatusThresholdMet
	PetitionStatusConverted  // Converted to formal report
	PetitionStatusExpired    // Didn't reach threshold
	PetitionStatusWithdrawn  // Creator withdrew
)

func (ps PetitionStatus) String() string {
	switch ps {
	case PetitionStatusOpen:
		return "open"
	case PetitionStatusThresholdMet:
		return "threshold_met"
	case PetitionStatusConverted:
		return "converted"
	case PetitionStatusExpired:
		return "expired"
	case PetitionStatusWithdrawn:
		return "withdrawn"
	default:
		return "unknown"
	}
}

// PetitionSignature represents a signature/support on a petition
type PetitionSignature struct {
	Signer     string         `json:"signer"`
	SharesHeld math.Int       `json:"shares_held"` // How many shares they hold (for weighted voting)
	SignedAt   time.Time      `json:"signed_at"`
	Comment    string         `json:"comment,omitempty"` // Optional comment
}

// Validate validates a PetitionSignature
func (ps PetitionSignature) Validate() error {
	if ps.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(ps.Signer); err != nil {
		return fmt.Errorf("invalid signer address: %v", err)
	}
	if ps.SharesHeld.IsNil() || ps.SharesHeld.LTE(math.ZeroInt()) {
		return fmt.Errorf("shares held must be positive")
	}
	return nil
}

// ShareholderPetition represents a petition that allows shareholders to collectively flag concerns
type ShareholderPetition struct {
	ID         uint64        `json:"id"`
	CompanyID  uint64        `json:"company_id"`
	ClassID    string        `json:"class_id"`     // Which share class (e.g., "COMMON")
	PetitionType PetitionType `json:"petition_type"`
	Title      string        `json:"title"`
	Description string       `json:"description"`
	Creator    string        `json:"creator"` // Original petitioner

	// Signature tracking
	Signatures     []PetitionSignature `json:"signatures"`
	SignatureCount int                 `json:"signature_count"` // Total unique signers

	// Thresholds (either/or)
	ShareholderThreshold int `json:"shareholder_threshold"` // e.g., 10% of shareholders
	AbsoluteThreshold    int `json:"absolute_threshold"`    // OR 100 signatures minimum
	ThresholdMet         bool `json:"threshold_met"`

	// Timeline
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`   // 7 days to gather signatures
	ConvertedAt time.Time `json:"converted_at,omitempty"` // When converted to formal report

	// Conversion tracking
	ConvertedToReportID uint64          `json:"converted_to_report_id,omitempty"` // If threshold met, auto-creates report
	Status              PetitionStatus  `json:"status"`
}

// NewShareholderPetition creates a new petition
func NewShareholderPetition(
	id, companyID uint64,
	classID string,
	petitionType PetitionType,
	title, description, creator string,
	shareholderThreshold, absoluteThreshold int,
	blockTime time.Time,
) ShareholderPetition {
	return ShareholderPetition{
		ID:                   id,
		CompanyID:            companyID,
		ClassID:              classID,
		PetitionType:         petitionType,
		Title:                title,
		Description:          description,
		Creator:              creator,
		Signatures:           []PetitionSignature{},
		SignatureCount:       0,
		ShareholderThreshold: shareholderThreshold,
		AbsoluteThreshold:    absoluteThreshold,
		ThresholdMet:         false,
		CreatedAt:            blockTime,
		ExpiresAt:            blockTime.Add(7 * 24 * time.Hour), // 7 days
		Status:               PetitionStatusOpen,
	}
}

// Validate validates a ShareholderPetition
func (p ShareholderPetition) Validate() error {
	if p.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if p.ClassID == "" {
		return fmt.Errorf("class ID cannot be empty")
	}
	if p.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	if p.Description == "" {
		return fmt.Errorf("description cannot be empty")
	}
	if p.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(p.Creator); err != nil {
		return fmt.Errorf("invalid creator address: %v", err)
	}
	if p.ShareholderThreshold <= 0 && p.AbsoluteThreshold <= 0 {
		return fmt.Errorf("at least one threshold must be positive")
	}
	return nil
}

// HasSigned checks if an address has already signed this petition
func (p ShareholderPetition) HasSigned(address string) bool {
	for _, sig := range p.Signatures {
		if sig.Signer == address {
			return true
		}
	}
	return false
}

// AddSignature adds a signature to the petition
func (p *ShareholderPetition) AddSignature(signature PetitionSignature) error {
	// Check for duplicate
	if p.HasSigned(signature.Signer) {
		return fmt.Errorf("address has already signed this petition")
	}

	// Validate signature
	if err := signature.Validate(); err != nil {
		return err
	}

	// Add signature
	p.Signatures = append(p.Signatures, signature)
	p.SignatureCount++

	return nil
}

// IsExpired checks if the petition has expired
func (p ShareholderPetition) IsExpired(blockTime time.Time) bool {
	return blockTime.After(p.ExpiresAt)
}

// CheckThreshold checks if the petition has met its threshold
// Returns (thresholdMet bool, reason string)
func (p ShareholderPetition) CheckThreshold(totalShareholders int) (bool, string) {
	// Check absolute threshold first (easier)
	if p.AbsoluteThreshold > 0 && p.SignatureCount >= p.AbsoluteThreshold {
		return true, fmt.Sprintf("absolute threshold met: %d/%d signatures", p.SignatureCount, p.AbsoluteThreshold)
	}

	// Check shareholder percentage threshold
	if p.ShareholderThreshold > 0 && totalShareholders > 0 {
		percentageThreshold := float64(p.ShareholderThreshold) / 100.0
		requiredSignatures := int(float64(totalShareholders) * percentageThreshold)

		if p.SignatureCount >= requiredSignatures {
			return true, fmt.Sprintf("percentage threshold met: %d/%d signatures (%.1f%% of %d shareholders)",
				p.SignatureCount, requiredSignatures, percentageThreshold*100, totalShareholders)
		}
	}

	return false, "threshold not met"
}

// GetTotalSharesSigned returns the total number of shares represented by signatures
func (p ShareholderPetition) GetTotalSharesSigned() math.Int {
	total := math.ZeroInt()
	for _, sig := range p.Signatures {
		total = total.Add(sig.SharesHeld)
	}
	return total
}

// ConvertToPriority converts petition to report priority based on signatures
// More signatures = higher priority
func (p ShareholderPetition) ConvertToPriority() int {
	// Base priority is 3 (medium)
	priority := 3

	// Increase priority based on signature count
	if p.SignatureCount >= 200 {
		priority = 5 // Emergency
	} else if p.SignatureCount >= 150 {
		priority = 4 // High
	} else if p.SignatureCount >= 100 {
		priority = 3 // Medium
	}

	return priority
}

// Petition threshold constants
const (
	// DefaultShareholderThreshold is the default percentage of shareholders needed (10%)
	DefaultShareholderThreshold = 10

	// DefaultAbsoluteThreshold is the default absolute number of signatures needed (100)
	DefaultAbsoluteThreshold = 100

	// PetitionExpirationDays is how long a petition stays open (7 days)
	PetitionExpirationDays = 7
)
