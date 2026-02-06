package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// =============================================================================
// Dividend Audit Document Types
// =============================================================================

// AuditDocumentType represents the type of audit document
type AuditDocumentType int32

const (
	// AuditDocTypeUnknown - Unknown document type
	AuditDocTypeUnknown AuditDocumentType = iota
	// AuditDocTypePDF - PDF document
	AuditDocTypePDF
	// AuditDocTypeDOCX - Microsoft Word document
	AuditDocTypeDOCX
	// AuditDocTypeXLSX - Microsoft Excel spreadsheet
	AuditDocTypeXLSX
	// AuditDocTypePNG - PNG image (for scanned documents)
	AuditDocTypePNG
	// AuditDocTypeJPG - JPG image (for scanned documents)
	AuditDocTypeJPG
	// AuditDocTypeJSON - JSON document (for structured data)
	AuditDocTypeJSON
)

func (t AuditDocumentType) String() string {
	switch t {
	case AuditDocTypePDF:
		return "pdf"
	case AuditDocTypeDOCX:
		return "docx"
	case AuditDocTypeXLSX:
		return "xlsx"
	case AuditDocTypePNG:
		return "png"
	case AuditDocTypeJPG:
		return "jpg"
	case AuditDocTypeJSON:
		return "json"
	default:
		return "unknown"
	}
}

// ParseAuditDocumentType parses a string to AuditDocumentType
func ParseAuditDocumentType(s string) AuditDocumentType {
	switch strings.ToLower(s) {
	case "pdf":
		return AuditDocTypePDF
	case "docx", "doc":
		return AuditDocTypeDOCX
	case "xlsx", "xls":
		return AuditDocTypeXLSX
	case "png":
		return AuditDocTypePNG
	case "jpg", "jpeg":
		return AuditDocTypeJPG
	case "json":
		return AuditDocTypeJSON
	default:
		return AuditDocTypeUnknown
	}
}

// IsValidAuditDocType checks if the document type is valid for audit purposes
func (t AuditDocumentType) IsValid() bool {
	return t != AuditDocTypeUnknown
}

// AuditVerificationStatus represents the verification status of an audit
type AuditVerificationStatus int32

const (
	// AuditStatusPending - Audit submitted but not yet verified
	AuditStatusPending AuditVerificationStatus = iota
	// AuditStatusVerified - Audit verified by validators
	AuditStatusVerified
	// AuditStatusRejected - Audit rejected as invalid/fraudulent
	AuditStatusRejected
	// AuditStatusDisputed - Audit under dispute
	AuditStatusDisputed
)

func (s AuditVerificationStatus) String() string {
	switch s {
	case AuditStatusPending:
		return "pending"
	case AuditStatusVerified:
		return "verified"
	case AuditStatusRejected:
		return "rejected"
	case AuditStatusDisputed:
		return "disputed"
	default:
		return "unknown"
	}
}

// DividendAudit represents an audit document for a dividend declaration
// This ensures company owners provide proper documentation for dividends
type DividendAudit struct {
	ID          uint64            `json:"id"`
	DividendID  uint64            `json:"dividend_id"`  // Associated dividend ID
	CompanyID   uint64            `json:"company_id"`   // Company ID
	Submitter   string            `json:"submitter"`    // Who submitted the audit

	// Document reference (stored off-chain, referenced by hash)
	ContentHash   string            `json:"content_hash"`   // IPFS hash, SHA256, or content-addressable hash
	StorageURI    string            `json:"storage_uri"`    // IPFS URI, Arweave URI, or other storage location
	DocumentType  AuditDocumentType `json:"document_type"`  // Type of document
	FileName      string            `json:"file_name"`      // Original file name
	FileSize      uint64            `json:"file_size"`      // File size in bytes

	// Audit content summary (on-chain, for quick reference)
	Title         string            `json:"title"`          // Audit title/description
	AuditorName   string            `json:"auditor_name"`   // Name of auditing firm/individual
	AuditorID     string            `json:"auditor_id"`     // Auditor registration/license ID
	AuditPeriod   string            `json:"audit_period"`   // e.g., "Q1 2025", "FY 2024"
	ReportDate    time.Time         `json:"report_date"`    // Date of audit report

	// Financial summary from the audit
	ReportedRevenue  string          `json:"reported_revenue,omitempty"`   // Revenue as string for large numbers
	ReportedProfit   string          `json:"reported_profit,omitempty"`    // Profit/loss as string
	ReportedDividend string          `json:"reported_dividend,omitempty"`  // Dividend amount in audit

	// Verification
	Status        AuditVerificationStatus `json:"status"`
	VerifiedBy    []string                `json:"verified_by,omitempty"`    // Validators who verified
	RejectionReason string                `json:"rejection_reason,omitempty"`

	// Timestamps
	SubmittedAt   time.Time         `json:"submitted_at"`
	VerifiedAt    time.Time         `json:"verified_at,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// Validate validates a DividendAudit
func (a DividendAudit) Validate() error {
	if a.DividendID == 0 {
		return fmt.Errorf("dividend ID cannot be zero")
	}
	if a.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if a.Submitter == "" {
		return fmt.Errorf("submitter cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(a.Submitter); err != nil {
		return fmt.Errorf("invalid submitter address: %v", err)
	}
	if a.ContentHash == "" {
		return fmt.Errorf("content hash cannot be empty")
	}
	if len(a.ContentHash) < 32 {
		return fmt.Errorf("content hash too short, minimum 32 characters")
	}
	// Max length validation for content hash (IPFS CID can be up to 64 chars)
	const maxContentHashLength = 128
	if len(a.ContentHash) > maxContentHashLength {
		return fmt.Errorf("content hash exceeds maximum length of %d characters", maxContentHashLength)
	}
	if a.StorageURI == "" {
		return fmt.Errorf("storage URI cannot be empty")
	}
	// Max length validation for storage URI
	const maxStorageURILength = 512
	if len(a.StorageURI) > maxStorageURILength {
		return fmt.Errorf("storage URI exceeds maximum length of %d characters", maxStorageURILength)
	}
	if !a.DocumentType.IsValid() {
		return fmt.Errorf("invalid document type: %s", a.DocumentType)
	}
	if a.FileName == "" {
		return fmt.Errorf("file name cannot be empty")
	}
	// Max length validation for file name
	const maxFileNameLength = 256
	if len(a.FileName) > maxFileNameLength {
		return fmt.Errorf("file name exceeds maximum length of %d characters", maxFileNameLength)
	}
	if a.Title == "" {
		return fmt.Errorf("audit title cannot be empty")
	}
	// Max length validation for title
	const maxTitleLength = 200
	if len(a.Title) > maxTitleLength {
		return fmt.Errorf("title exceeds maximum length of %d characters", maxTitleLength)
	}
	if a.AuditorName == "" {
		return fmt.Errorf("auditor name cannot be empty")
	}
	// Max length validation for auditor name
	const maxAuditorNameLength = 200
	if len(a.AuditorName) > maxAuditorNameLength {
		return fmt.Errorf("auditor name exceeds maximum length of %d characters", maxAuditorNameLength)
	}
	// Max length validation for auditor ID
	const maxAuditorIDLength = 100
	if len(a.AuditorID) > maxAuditorIDLength {
		return fmt.Errorf("auditor ID exceeds maximum length of %d characters", maxAuditorIDLength)
	}
	if a.AuditPeriod == "" {
		return fmt.Errorf("audit period cannot be empty")
	}
	// Max length validation for audit period
	const maxAuditPeriodLength = 50
	if len(a.AuditPeriod) > maxAuditPeriodLength {
		return fmt.Errorf("audit period exceeds maximum length of %d characters", maxAuditPeriodLength)
	}
	if a.ReportDate.IsZero() {
		return fmt.Errorf("report date cannot be zero")
	}
	// Max length validation for rejection reason
	const maxRejectionReasonLength = 1000
	if len(a.RejectionReason) > maxRejectionReasonLength {
		return fmt.Errorf("rejection reason exceeds maximum length of %d characters", maxRejectionReasonLength)
	}
	// Note: Time validation moved to keeper methods where we have access to block time
	return nil
}

// NewDividendAudit creates a new dividend audit
// Note: blockTime parameter added to avoid using time.Now() which varies by node
func NewDividendAudit(
	id, dividendID, companyID uint64,
	submitter string,
	contentHash, storageURI string,
	documentType AuditDocumentType,
	fileName string,
	fileSize uint64,
	title, auditorName, auditorID, auditPeriod string,
	reportDate time.Time,
	blockTime time.Time, // Use block time for consistency across nodes
) DividendAudit {
	return DividendAudit{
		ID:           id,
		DividendID:   dividendID,
		CompanyID:    companyID,
		Submitter:    submitter,
		ContentHash:  contentHash,
		StorageURI:   storageURI,
		DocumentType: documentType,
		FileName:     fileName,
		FileSize:     fileSize,
		Title:        title,
		AuditorName:  auditorName,
		AuditorID:    auditorID,
		AuditPeriod:  auditPeriod,
		ReportDate:   reportDate,
		Status:       AuditStatusPending,
		SubmittedAt:  blockTime,
		UpdatedAt:    blockTime,
	}
}

// AuditVerification represents a validator's verification of an audit
type AuditVerification struct {
	AuditID     uint64    `json:"audit_id"`
	Verifier    string    `json:"verifier"`     // Validator address
	Approved    bool      `json:"approved"`     // true = verified, false = rejected
	Comments    string    `json:"comments"`
	VerifiedAt  time.Time `json:"verified_at"`
}

// Validate validates an AuditVerification
func (v AuditVerification) Validate() error {
	if v.AuditID == 0 {
		return fmt.Errorf("audit ID cannot be zero")
	}
	if v.Verifier == "" {
		return fmt.Errorf("verifier cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(v.Verifier); err != nil {
		return fmt.Errorf("invalid verifier address: %v", err)
	}
	return nil
}

// =============================================================================
// Audit Request Message (for declaring dividend with audit)
// =============================================================================

// AuditInfo contains audit information for dividend declaration
type AuditInfo struct {
	// Required fields
	ContentHash   string `json:"content_hash"`   // IPFS hash or SHA256 of document
	StorageURI    string `json:"storage_uri"`    // URI where document is stored
	DocumentType  string `json:"document_type"`  // "pdf", "docx", "xlsx", "png", "jpg", "json"
	FileName      string `json:"file_name"`      // Original file name

	// Required audit metadata
	Title         string    `json:"title"`          // Audit title
	AuditorName   string    `json:"auditor_name"`   // Auditing firm/individual name
	AuditorID     string    `json:"auditor_id"`     // Auditor registration ID (optional)
	AuditPeriod   string    `json:"audit_period"`   // e.g., "Q1 2025", "FY 2024"
	ReportDate    time.Time `json:"report_date"`    // Date of the audit report

	// Optional financial summary
	ReportedRevenue  string `json:"reported_revenue,omitempty"`
	ReportedProfit   string `json:"reported_profit,omitempty"`
	ReportedDividend string `json:"reported_dividend,omitempty"`
	FileSize         uint64 `json:"file_size,omitempty"`
}

// Validate validates AuditInfo
func (a AuditInfo) Validate() error {
	if a.ContentHash == "" {
		return ErrAuditContentHashRequired
	}
	if len(a.ContentHash) < 32 {
		return ErrAuditContentHashTooShort
	}
	// Max length validation for content hash
	const maxContentHashLength = 128
	if len(a.ContentHash) > maxContentHashLength {
		return fmt.Errorf("content hash exceeds maximum length of %d characters", maxContentHashLength)
	}
	if a.StorageURI == "" {
		return ErrAuditStorageURIRequired
	}
	// Max length validation for storage URI
	const maxStorageURILength = 512
	if len(a.StorageURI) > maxStorageURILength {
		return fmt.Errorf("storage URI exceeds maximum length of %d characters", maxStorageURILength)
	}
	docType := ParseAuditDocumentType(a.DocumentType)
	if !docType.IsValid() {
		return ErrAuditInvalidDocumentType
	}
	// Max length validation for file name
	const maxFileNameLength = 256
	if len(a.FileName) > maxFileNameLength {
		return fmt.Errorf("file name exceeds maximum length of %d characters", maxFileNameLength)
	}
	if a.Title == "" {
		return ErrAuditTitleRequired
	}
	// Max length validation for title
	const maxTitleLength = 200
	if len(a.Title) > maxTitleLength {
		return fmt.Errorf("title exceeds maximum length of %d characters", maxTitleLength)
	}
	if a.AuditorName == "" {
		return ErrAuditAuditorNameRequired
	}
	// Max length validation for auditor name
	const maxAuditorNameLength = 200
	if len(a.AuditorName) > maxAuditorNameLength {
		return fmt.Errorf("auditor name exceeds maximum length of %d characters", maxAuditorNameLength)
	}
	// Max length validation for auditor ID
	const maxAuditorIDLength = 100
	if len(a.AuditorID) > maxAuditorIDLength {
		return fmt.Errorf("auditor ID exceeds maximum length of %d characters", maxAuditorIDLength)
	}
	if a.AuditPeriod == "" {
		return ErrAuditPeriodRequired
	}
	// Max length validation for audit period
	const maxAuditPeriodLength = 50
	if len(a.AuditPeriod) > maxAuditPeriodLength {
		return fmt.Errorf("audit period exceeds maximum length of %d characters", maxAuditPeriodLength)
	}
	if a.ReportDate.IsZero() {
		return ErrAuditReportDateRequired
	}
	// Note: Time validation moved to keeper methods where we have access to block time
	return nil
}

// =============================================================================
// Audit Summary for Shareholder View
// =============================================================================

// AuditSummary provides a simplified view of audit for shareholders
type AuditSummary struct {
	DividendID    uint64                  `json:"dividend_id"`
	CompanyID     uint64                  `json:"company_id"`
	Title         string                  `json:"title"`
	AuditorName   string                  `json:"auditor_name"`
	AuditPeriod   string                  `json:"audit_period"`
	ReportDate    time.Time               `json:"report_date"`
	DocumentType  string                  `json:"document_type"`
	StorageURI    string                  `json:"storage_uri"`    // Where to download
	ContentHash   string                  `json:"content_hash"`   // To verify integrity
	Status        AuditVerificationStatus `json:"status"`
	VerifiedBy    int                     `json:"verified_by"`    // Number of validators who verified
	SubmittedAt   time.Time               `json:"submitted_at"`
}

// ToSummary converts a DividendAudit to AuditSummary
func (a DividendAudit) ToSummary() AuditSummary {
	return AuditSummary{
		DividendID:   a.DividendID,
		CompanyID:    a.CompanyID,
		Title:        a.Title,
		AuditorName:  a.AuditorName,
		AuditPeriod:  a.AuditPeriod,
		ReportDate:   a.ReportDate,
		DocumentType: a.DocumentType.String(),
		StorageURI:   a.StorageURI,
		ContentHash:  a.ContentHash,
		Status:       a.Status,
		VerifiedBy:   len(a.VerifiedBy),
		SubmittedAt:  a.SubmittedAt,
	}
}

// =============================================================================
// Message Types for Audit Management
// =============================================================================

// MsgVerifyAudit allows validators to verify or reject an audit
type MsgVerifyAudit struct {
	AuditID  uint64 `json:"audit_id"`
	Verifier string `json:"verifier"`  // Validator address
	Approved bool   `json:"approved"`  // true = verified, false = rejected
	Comments string `json:"comments"`  // Optional verification comments
}

// ValidateBasic performs basic validation of MsgVerifyAudit
func (m MsgVerifyAudit) ValidateBasic() error {
	if m.AuditID == 0 {
		return fmt.Errorf("audit ID cannot be zero")
	}
	if m.Verifier == "" {
		return fmt.Errorf("verifier cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Verifier); err != nil {
		return fmt.Errorf("invalid verifier address: %v", err)
	}
	// Validate max length on comments
	const maxCommentsLength = 1000
	if len(m.Comments) > maxCommentsLength {
		return fmt.Errorf("comments exceed maximum length of %d characters", maxCommentsLength)
	}
	return nil
}

// MsgDisputeAudit allows shareholders to dispute an audit
type MsgDisputeAudit struct {
	AuditID   uint64 `json:"audit_id"`
	Disputer  string `json:"disputer"`   // Shareholder address
	Reason    string `json:"reason"`     // Reason for dispute
}

// ValidateBasic performs basic validation of MsgDisputeAudit
func (m MsgDisputeAudit) ValidateBasic() error {
	if m.AuditID == 0 {
		return fmt.Errorf("audit ID cannot be zero")
	}
	if m.Disputer == "" {
		return fmt.Errorf("disputer cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.Disputer); err != nil {
		return fmt.Errorf("invalid disputer address: %v", err)
	}
	if m.Reason == "" {
		return fmt.Errorf("reason cannot be empty")
	}
	// Validate max length on reason
	const maxReasonLength = 500
	if len(m.Reason) > maxReasonLength {
		return fmt.Errorf("reason exceeds maximum length of %d characters", maxReasonLength)
	}
	return nil
}
