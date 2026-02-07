package types

import (
	"context"

	"cosmossdk.io/math"
)

// MsgServer defines the module's message service.
type MsgServer interface {
	CreateCompany(context.Context, *SimpleMsgCreateCompany) (*MsgCreateCompanyResponse, error)
	CreateShareClass(context.Context, *SimpleMsgCreateShareClass) (*MsgCreateShareClassResponse, error)
	IssueShares(context.Context, *SimpleMsgIssueShares) (*MsgIssueSharesResponse, error)
	TransferShares(context.Context, *SimpleMsgTransferShares) (*MsgTransferSharesResponse, error)
	UpdateCompany(context.Context, *SimpleMsgUpdateCompany) (*MsgUpdateCompanyResponse, error)

	// Anti-dilution message handlers
	RegisterAntiDilution(context.Context, *SimpleMsgRegisterAntiDilution) (*MsgRegisterAntiDilutionResponse, error)
	UpdateAntiDilution(context.Context, *SimpleMsgUpdateAntiDilution) (*MsgUpdateAntiDilutionResponse, error)
	IssueSharesWithProtection(context.Context, *SimpleMsgIssueSharesWithProtection) (*MsgIssueSharesWithProtectionResponse, error)
}

// RegisterMsgServer registers the message server
func RegisterMsgServer(server interface{}, impl MsgServer) {
	// In a real implementation this would register the gRPC server
	// For our simplified version, we'll handle this in the module
}

// =============================================================================
// Query Service Types
// =============================================================================

// QueryServer defines the module's query service
type QueryServer interface {
	// Company queries
	Company(context.Context, *QueryCompanyRequest) (*QueryCompanyResponse, error)
	CompanyBySymbol(context.Context, *QueryCompanyBySymbolRequest) (*QueryCompanyResponse, error)
	Companies(context.Context, *QueryCompaniesRequest) (*QueryCompaniesResponse, error)

	// Share class queries
	ShareClass(context.Context, *QueryShareClassRequest) (*QueryShareClassResponse, error)
	ShareClasses(context.Context, *QueryShareClassesRequest) (*QueryShareClassesResponse, error)

	// Shareholding queries
	Shareholding(context.Context, *QueryShareholdingRequest) (*QueryShareholdingResponse, error)
	ShareholdingsByOwner(context.Context, *QueryShareholdingsByOwnerRequest) (*QueryShareholdingsByOwnerResponse, error)
	ShareholdingsByCompany(context.Context, *QueryShareholdingsByCompanyRequest) (*QueryShareholdingsByCompanyResponse, error)

	// Anti-dilution queries
	AntiDilutionProvision(context.Context, *QueryAntiDilutionProvisionRequest) (*QueryAntiDilutionProvisionResponse, error)
	AntiDilutionProvisions(context.Context, *QueryAntiDilutionProvisionsRequest) (*QueryAntiDilutionProvisionsResponse, error)
	IssuanceHistory(context.Context, *QueryIssuanceHistoryRequest) (*QueryIssuanceHistoryResponse, error)
	AdjustmentHistory(context.Context, *QueryAdjustmentHistoryRequest) (*QueryAdjustmentHistoryResponse, error)
	CompanyCapTable(context.Context, *QueryCompanyCapTableRequest) (*QueryCompanyCapTableResponse, error)

	// Dividend queries
	PendingDividends(context.Context, *QueryPendingDividendsRequest) (*QueryPendingDividendsResponse, error)
	DividendsByCompany(context.Context, *QueryDividendsByCompanyRequest) (*QueryDividendsByCompanyResponse, error)
}

// =============================================================================
// Company Query Types
// =============================================================================

type QueryCompanyRequest struct {
	CompanyID uint64 `json:"company_id"`
}

type QueryCompanyBySymbolRequest struct {
	Symbol string `json:"symbol"`
}

type QueryCompanyResponse struct {
	Company Company `json:"company"`
}

type QueryCompaniesRequest struct {
	Status     string `json:"status,omitempty"`     // Filter by status
	Industry   string `json:"industry,omitempty"`   // Filter by industry
	Pagination *PageRequest `json:"pagination,omitempty"`
}

type QueryCompaniesResponse struct {
	Companies  []Company     `json:"companies"`
	Pagination *PageResponse `json:"pagination,omitempty"`
}

// =============================================================================
// Share Class Query Types
// =============================================================================

type QueryShareClassRequest struct {
	CompanyID uint64 `json:"company_id"`
	ClassID   string `json:"class_id"`
}

type QueryShareClassResponse struct {
	ShareClass ShareClass `json:"share_class"`
}

type QueryShareClassesRequest struct {
	CompanyID uint64 `json:"company_id"`
}

type QueryShareClassesResponse struct {
	ShareClasses []ShareClass `json:"share_classes"`
}

// =============================================================================
// Shareholding Query Types
// =============================================================================

type QueryShareholdingRequest struct {
	CompanyID uint64 `json:"company_id"`
	ClassID   string `json:"class_id"`
	Owner     string `json:"owner"`
}

type QueryShareholdingResponse struct {
	Shareholding Shareholding `json:"shareholding"`
}

type QueryShareholdingsByOwnerRequest struct {
	Owner      string       `json:"owner"`
	Pagination *PageRequest `json:"pagination,omitempty"`
}

type QueryShareholdingsByOwnerResponse struct {
	Shareholdings []Shareholding `json:"shareholdings"`
	Pagination    *PageResponse  `json:"pagination,omitempty"`
}

type QueryShareholdingsByCompanyRequest struct {
	CompanyID  uint64       `json:"company_id"`
	ClassID    string       `json:"class_id,omitempty"` // Optional: filter by class
	Pagination *PageRequest `json:"pagination,omitempty"`
}

type QueryShareholdingsByCompanyResponse struct {
	Shareholdings []Shareholding `json:"shareholdings"`
	Pagination    *PageResponse  `json:"pagination,omitempty"`
}

// =============================================================================
// Anti-Dilution Query Types
// =============================================================================

type QueryAntiDilutionProvisionRequest struct {
	CompanyID uint64 `json:"company_id"`
	ClassID   string `json:"class_id"`
}

type QueryAntiDilutionProvisionResponse struct {
	Provision AntiDilutionProvision `json:"provision"`
	Found     bool                  `json:"found"`
}

type QueryAntiDilutionProvisionsRequest struct {
	CompanyID uint64 `json:"company_id"`
}

type QueryAntiDilutionProvisionsResponse struct {
	Provisions []AntiDilutionProvision `json:"provisions"`
}

type QueryIssuanceHistoryRequest struct {
	CompanyID  uint64       `json:"company_id"`
	ClassID    string       `json:"class_id,omitempty"` // Optional: filter by class
	Pagination *PageRequest `json:"pagination,omitempty"`
}

type QueryIssuanceHistoryResponse struct {
	Records    []IssuanceRecord `json:"records"`
	Pagination *PageResponse    `json:"pagination,omitempty"`
}

type QueryAdjustmentHistoryRequest struct {
	CompanyID   uint64       `json:"company_id"`
	ClassID     string       `json:"class_id,omitempty"`     // Optional filter
	Shareholder string       `json:"shareholder,omitempty"`  // Optional filter
	Pagination  *PageRequest `json:"pagination,omitempty"`
}

type QueryAdjustmentHistoryResponse struct {
	Adjustments []AntiDilutionAdjustment `json:"adjustments"`
	Pagination  *PageResponse            `json:"pagination,omitempty"`
}

// =============================================================================
// Cap Table Query Types (comprehensive company equity view)
// =============================================================================

type QueryCompanyCapTableRequest struct {
	CompanyID uint64 `json:"company_id"`
}

type QueryCompanyCapTableResponse struct {
	Company        Company                     `json:"company"`
	ShareClasses   []ShareClassSummary         `json:"share_classes"`
	TopHolders     []ShareholderSummary        `json:"top_holders"`
	IssuanceStats  IssuanceStatistics          `json:"issuance_stats"`
	AntiDilution   AntiDilutionSummary         `json:"anti_dilution"`
}

// ShareClassSummary provides a summary of a share class for cap table
type ShareClassSummary struct {
	ClassID           string           `json:"class_id"`
	Name              string           `json:"name"`
	AuthorizedShares  math.Int         `json:"authorized_shares"`
	IssuedShares      math.Int         `json:"issued_shares"`
	OutstandingShares math.Int         `json:"outstanding_shares"`
	HolderCount       uint64           `json:"holder_count"`
	HasAntiDilution   bool             `json:"has_anti_dilution"`
	AntiDilutionType  AntiDilutionType `json:"anti_dilution_type,omitempty"`
	VotingRights      bool             `json:"voting_rights"`
	DividendRights    bool             `json:"dividend_rights"`
}

// ShareholderSummary provides a summary of a shareholder's holdings
type ShareholderSummary struct {
	Owner            string         `json:"owner"`
	TotalShares      math.Int       `json:"total_shares"`
	OwnershipPercent math.LegacyDec `json:"ownership_percent"`
	ClassID          string         `json:"class_id"`
	VotingPower      math.LegacyDec `json:"voting_power"`
}

// IssuanceStatistics provides statistics about share issuances
type IssuanceStatistics struct {
	TotalIssuances     uint64         `json:"total_issuances"`
	TotalSharesIssued  math.Int       `json:"total_shares_issued"`
	TotalRaised        math.LegacyDec `json:"total_raised"`
	DownRoundCount     uint64         `json:"down_round_count"`
	LastIssuancePrice  math.LegacyDec `json:"last_issuance_price"`
	AverageIssuePrice  math.LegacyDec `json:"average_issue_price"`
}

// AntiDilutionSummary provides a summary of anti-dilution status
type AntiDilutionSummary struct {
	ProtectedClasses    uint64         `json:"protected_classes"`
	TotalAdjustments    uint64         `json:"total_adjustments"`
	TotalSharesAdjusted math.Int       `json:"total_shares_adjusted"`
	ActiveProvisions    uint64         `json:"active_provisions"`
}

// =============================================================================
// Pagination Types
// =============================================================================

type PageRequest struct {
	Key    []byte `json:"key,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
	Limit  uint64 `json:"limit,omitempty"`
}

type PageResponse struct {
	NextKey []byte `json:"next_key,omitempty"`
	Total   uint64 `json:"total,omitempty"`
}

// =============================================================================
// Dividend Query Types
// =============================================================================

// QueryPendingDividendsRequest queries pending dividends for a shareholder
type QueryPendingDividendsRequest struct {
	Shareholder string `json:"shareholder"` // Shareholder address
}

// QueryPendingDividendsResponse returns pending dividends
type QueryPendingDividendsResponse struct {
	Dividends []PendingDividendInfo `json:"dividends"`
}

// PendingDividendInfo contains dividend info with estimated payment
type PendingDividendInfo struct {
	DividendID      uint64  `json:"dividend_id"`
	CompanyID       uint64  `json:"company_id"`
	CompanyName     string  `json:"company_name"`
	CompanySymbol   string  `json:"company_symbol"`
	Type            string  `json:"type"`             // "cash" or "stock"
	Status          string  `json:"status"`           // "pending_approval", "declared", etc.
	AmountPerShare  string  `json:"amount_per_share"` // Amount per share
	Currency        string  `json:"currency"`
	SharesHeld      string  `json:"shares_held"`      // User's shares at record date (estimate)
	EstimatedAmount string  `json:"estimated_amount"` // Estimated dividend payment
	RecordDate      string  `json:"record_date"`
	PaymentDate     string  `json:"payment_date"`
	ProposalID      uint64  `json:"proposal_id,omitempty"` // Governance proposal ID
	AuditHash       string  `json:"audit_hash,omitempty"`  // Audit document hash
	Description     string  `json:"description"`
}

// QueryDividendsByCompanyRequest queries dividends for a company
type QueryDividendsByCompanyRequest struct {
	CompanyID uint64 `json:"company_id"`
	Status    string `json:"status,omitempty"` // Filter by status
}

// QueryDividendsByCompanyResponse returns company dividends
type QueryDividendsByCompanyResponse struct {
	Dividends []Dividend `json:"dividends"`
}