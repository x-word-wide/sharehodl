package types

import (
	"time"

	"cosmossdk.io/math"
)

// ListingStatus represents the status of a business listing application
type ListingStatus int32

const (
	ListingStatusPending     ListingStatus = 0
	ListingStatusUnderReview ListingStatus = 1
	ListingStatusApproved    ListingStatus = 2
	ListingStatusRejected    ListingStatus = 3
	ListingStatusSuspended   ListingStatus = 4
)

// String returns the string representation of ListingStatus
func (ls ListingStatus) String() string {
	switch ls {
	case ListingStatusPending:
		return "pending"
	case ListingStatusUnderReview:
		return "under_review"
	case ListingStatusApproved:
		return "approved"
	case ListingStatusRejected:
		return "rejected"
	case ListingStatusSuspended:
		return "suspended"
	default:
		return "unknown"
	}
}

// BusinessListing represents a company's application to list on ShareHODL
type BusinessListing struct {
	ID                  uint64                 `json:"id"`
	CompanyName         string                 `json:"company_name"`
	LegalEntityName     string                 `json:"legal_entity_name"`
	RegistrationNumber  string                 `json:"registration_number"`
	Jurisdiction        string                 `json:"jurisdiction"`
	Industry            string                 `json:"industry"`
	BusinessDescription string                 `json:"business_description"`
	FoundedDate         time.Time              `json:"founded_date"`
	Website             string                 `json:"website"`
	ContactEmail        string                 `json:"contact_email"`
	TokenSymbol         string                 `json:"token_symbol"`
	TokenName           string                 `json:"token_name"`
	TotalShares         math.Int               `json:"total_shares"`
	InitialPrice        math.LegacyDec         `json:"initial_price"`
	ListingFee          math.Int               `json:"listing_fee"`
	Applicant           string                 `json:"applicant"`
	Status              ListingStatus          `json:"status"`
	SubmittedAt         time.Time              `json:"submitted_at"`
	ReviewStartedAt     time.Time              `json:"review_started_at"`
	ReviewCompletedAt   time.Time              `json:"review_completed_at"`
	AssignedValidators  []string               `json:"assigned_validators"`
	VerificationResults []VerificationResult   `json:"verification_results"`
	Documents           []BusinessDocument     `json:"documents"`
	RejectionReason     string                 `json:"rejection_reason"`
	Notes               string                 `json:"notes"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// VerificationResult represents the result of validator verification
type VerificationResult struct {
	ValidatorAddress string                 `json:"validator_address"`
	ValidatorTier    int                    `json:"validator_tier"`
	Approved         bool                   `json:"approved"`
	Comments         string                 `json:"comments"`
	Score            int                    `json:"score"`
	VerifiedAt       time.Time              `json:"verified_at"`
	VerificationData map[string]interface{} `json:"verification_data"`
}

// BusinessDocument represents required documentation
type BusinessDocument struct {
	ID           uint64    `json:"id"`
	DocumentType string    `json:"document_type"`
	FileName     string    `json:"file_name"`
	FileHash     string    `json:"file_hash"`
	FileSize     int64     `json:"file_size"`
	UploadedAt   time.Time `json:"uploaded_at"`
	Verified     bool      `json:"verified"`
	VerifiedBy   string    `json:"verified_by"`
	VerifiedAt   time.Time `json:"verified_at"`
}

// RequiredDocuments defines the required documents for listing
var RequiredDocuments = []string{
	"certificate_of_incorporation",
	"articles_of_association",
	"financial_statements_latest",
	"financial_statements_previous",
	"auditors_report",
	"board_resolution",
	"directors_list",
	"shareholders_list",
	"business_plan",
	"risk_assessment",
	"compliance_certificate",
	"insurance_certificate",
}

// VerificationCriteria defines what validators need to verify
type VerificationCriteria struct {
	LegalCompliance    bool `json:"legal_compliance"`
	FinancialHealth    bool `json:"financial_health"`
	BusinessViability  bool `json:"business_viability"`
	DocumentVeracity   bool `json:"document_veracity"`
	ManagementQuality  bool `json:"management_quality"`
	RiskAssessment     bool `json:"risk_assessment"`
	ComplianceProgram  bool `json:"compliance_program"`
	CorporateGovernance bool `json:"corporate_governance"`
}

// NewBusinessListing creates a new business listing application
func NewBusinessListing(
	id uint64,
	companyName, legalEntityName, registrationNumber, jurisdiction, industry string,
	businessDescription, website, contactEmail, tokenSymbol, tokenName string,
	totalShares math.Int,
	initialPrice math.LegacyDec,
	listingFee math.Int,
	applicant string,
) BusinessListing {
	now := time.Now()
	
	return BusinessListing{
		ID:                  id,
		CompanyName:         companyName,
		LegalEntityName:     legalEntityName,
		RegistrationNumber:  registrationNumber,
		Jurisdiction:        jurisdiction,
		Industry:            industry,
		BusinessDescription: businessDescription,
		Website:             website,
		ContactEmail:        contactEmail,
		TokenSymbol:         tokenSymbol,
		TokenName:           tokenName,
		TotalShares:         totalShares,
		InitialPrice:        initialPrice,
		ListingFee:          listingFee,
		Applicant:           applicant,
		Status:              ListingStatusPending,
		SubmittedAt:         now,
		AssignedValidators:  []string{},
		VerificationResults: []VerificationResult{},
		Documents:           []BusinessDocument{},
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// Validate validates a business listing
func (bl BusinessListing) Validate() error {
	if bl.CompanyName == "" {
		return ErrInvalidCompanyName
	}
	
	if bl.LegalEntityName == "" {
		return ErrInvalidLegalEntityName
	}
	
	if bl.RegistrationNumber == "" {
		return ErrInvalidRegistrationNumber
	}
	
	if bl.TokenSymbol == "" || len(bl.TokenSymbol) < 2 || len(bl.TokenSymbol) > 10 {
		return ErrInvalidTokenSymbol
	}
	
	if bl.TotalShares.IsZero() || bl.TotalShares.IsNegative() {
		return ErrInvalidTotalShares
	}
	
	if bl.InitialPrice.IsZero() || bl.InitialPrice.IsNegative() {
		return ErrInvalidInitialPrice
	}
	
	return nil
}

// HasRequiredDocuments checks if all required documents are uploaded
func (bl BusinessListing) HasRequiredDocuments() bool {
	docTypes := make(map[string]bool)
	for _, doc := range bl.Documents {
		docTypes[doc.DocumentType] = true
	}
	
	for _, required := range RequiredDocuments {
		if !docTypes[required] {
			return false
		}
	}
	
	return true
}

// GetVerificationScore calculates the overall verification score
func (bl BusinessListing) GetVerificationScore() (int, int) {
	if len(bl.VerificationResults) == 0 {
		return 0, 0
	}
	
	totalScore := 0
	maxScore := 0
	
	for _, result := range bl.VerificationResults {
		totalScore += result.Score
		maxScore += 100 // Assuming max score per validator is 100
	}
	
	return totalScore, maxScore
}

// CanApprove checks if the listing can be approved based on verification results
func (bl BusinessListing) CanApprove() bool {
	if len(bl.VerificationResults) < 3 {
		return false
	}
	
	approvals := 0
	for _, result := range bl.VerificationResults {
		if result.Approved {
			approvals++
		}
	}
	
	// Require at least 2/3 approval
	return approvals >= (len(bl.VerificationResults)*2)/3
}

// CanReject checks if the listing should be rejected
func (bl BusinessListing) CanReject() bool {
	if len(bl.VerificationResults) < 3 {
		return false
	}
	
	rejections := 0
	for _, result := range bl.VerificationResults {
		if !result.Approved {
			rejections++
		}
	}
	
	// Reject if more than 1/3 validators reject
	return rejections > len(bl.VerificationResults)/3
}

// ListingRequirements defines the requirements for listing on ShareHODL
type ListingRequirements struct {
	MinimumListingFee    math.Int       `json:"minimum_listing_fee"`
	MinimumShares        math.Int       `json:"minimum_shares"`
	MinimumPrice         math.LegacyDec `json:"minimum_price"`
	MinimumBusinessAge   int            `json:"minimum_business_age_months"`
	RequiredValidators   int            `json:"required_validators"`
	ReviewPeriodDays     int            `json:"review_period_days"`
	AllowedJurisdictions []string       `json:"allowed_jurisdictions"`
	ProhibitedIndustries []string       `json:"prohibited_industries"`
}

// DefaultListingRequirements returns the default listing requirements
func DefaultListingRequirements() ListingRequirements {
	return ListingRequirements{
		MinimumListingFee:    math.NewInt(10000000000), // 10,000 HODL
		MinimumShares:        math.NewInt(1000000),      // 1M shares minimum
		MinimumPrice:         math.LegacyNewDecWithPrec(1, 0), // $1 minimum
		MinimumBusinessAge:   12, // 12 months minimum
		RequiredValidators:   3,  // 3 validator approvals needed
		ReviewPeriodDays:     30, // 30 days review period
		AllowedJurisdictions: []string{
			"US", "CA", "GB", "AU", "DE", "FR", "JP", "SG", "CH", "NL",
		},
		ProhibitedIndustries: []string{
			"gambling", "adult_entertainment", "weapons", "illegal_substances",
			"pyramid_schemes", "multi_level_marketing",
		},
	}
}