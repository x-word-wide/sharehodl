package types

import (
	"time"

	"cosmossdk.io/math"
)

// DividendType represents the type of dividend
type DividendType int32

const (
	DividendTypeCash     DividendType = 0
	DividendTypeStock    DividendType = 1
	DividendTypeProperty DividendType = 2
	DividendTypeSpecial  DividendType = 3
)

// String returns the string representation of DividendType
func (dt DividendType) String() string {
	switch dt {
	case DividendTypeCash:
		return "cash"
	case DividendTypeStock:
		return "stock"
	case DividendTypeProperty:
		return "property"
	case DividendTypeSpecial:
		return "special"
	default:
		return "unknown"
	}
}

// DividendStatus represents the status of a dividend distribution
type DividendStatus int32

const (
	DividendStatusDeclared   DividendStatus = 0
	DividendStatusRecorded   DividendStatus = 1
	DividendStatusProcessing DividendStatus = 2
	DividendStatusPaid       DividendStatus = 3
	DividendStatusCancelled  DividendStatus = 4
)

// String returns the string representation of DividendStatus
func (ds DividendStatus) String() string {
	switch ds {
	case DividendStatusDeclared:
		return "declared"
	case DividendStatusRecorded:
		return "recorded"
	case DividendStatusProcessing:
		return "processing"
	case DividendStatusPaid:
		return "paid"
	case DividendStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Dividend represents a dividend distribution to shareholders
type Dividend struct {
	ID              uint64         `json:"id"`
	CompanyID       uint64         `json:"company_id"`
	ClassID         string         `json:"class_id"`         // Share class, empty for all classes
	Type            DividendType   `json:"type"`
	Currency        string         `json:"currency"`         // Denomination for cash dividends
	AmountPerShare  math.LegacyDec `json:"amount_per_share"` // Amount per share
	TotalAmount     math.LegacyDec `json:"total_amount"`     // Total dividend amount
	PaidAmount      math.LegacyDec `json:"paid_amount"`      // Amount already paid
	RemainingAmount math.LegacyDec `json:"remaining_amount"` // Amount remaining to pay
	
	// Important dates
	DeclarationDate time.Time `json:"declaration_date"`
	ExDividendDate  time.Time `json:"ex_dividend_date"`
	RecordDate      time.Time `json:"record_date"`
	PaymentDate     time.Time `json:"payment_date"`
	
	// Status and metadata
	Status              DividendStatus `json:"status"`
	EligibleShares      math.Int       `json:"eligible_shares"`      // Total shares eligible for dividend
	SharesProcessed     math.Int       `json:"shares_processed"`     // Shares processed so far
	ShareholdersEligible uint64        `json:"shareholders_eligible"` // Number of eligible shareholders
	ShareholdersPaid     uint64        `json:"shareholders_paid"`     // Number of shareholders paid
	
	// For stock dividends
	StockRatio          math.LegacyDec `json:"stock_ratio,omitempty"`          // New shares per existing share
	NewSharesIssued     math.Int       `json:"new_shares_issued,omitempty"`    // Total new shares issued
	
	// Administrative
	Declarant    string    `json:"declarant"`     // Address that declared the dividend
	TaxRate      math.LegacyDec `json:"tax_rate"`  // Withholding tax rate
	Description  string    `json:"description"`   // Human-readable description
	Notes        string    `json:"notes"`         // Additional notes
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CompletedAt  time.Time `json:"completed_at,omitempty"`
}

// DividendPayment represents a dividend payment to a specific shareholder
type DividendPayment struct {
	ID           uint64         `json:"id"`
	DividendID   uint64         `json:"dividend_id"`
	CompanyID    uint64         `json:"company_id"`
	ClassID      string         `json:"class_id"`
	Shareholder  string         `json:"shareholder"`   // Bech32 address
	SharesHeld   math.Int       `json:"shares_held"`   // Shares held on record date
	Amount       math.LegacyDec `json:"amount"`        // Dividend amount
	Currency     string         `json:"currency"`
	TaxWithheld  math.LegacyDec `json:"tax_withheld"`  // Tax withheld
	NetAmount    math.LegacyDec `json:"net_amount"`    // Net amount paid
	PaidAt       time.Time      `json:"paid_at"`
	TxHash       string         `json:"tx_hash"`       // Transaction hash of payment
	Status       string         `json:"status"`        // "pending", "paid", "failed"
	FailureReason string        `json:"failure_reason,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

// DividendClaim represents a claim for dividend payment (for manual claims)
type DividendClaim struct {
	ID          uint64         `json:"id"`
	DividendID  uint64         `json:"dividend_id"`
	CompanyID   uint64         `json:"company_id"`
	ClassID     string         `json:"class_id"`
	Claimant    string         `json:"claimant"`      // Bech32 address
	SharesHeld  math.Int       `json:"shares_held"`   // Shares held on record date
	Amount      math.LegacyDec `json:"amount"`        // Dividend amount claimed
	Status      string         `json:"status"`        // "pending", "approved", "rejected", "paid"
	ClaimedAt   time.Time      `json:"claimed_at"`
	ProcessedAt time.Time      `json:"processed_at,omitempty"`
	ProcessedBy string         `json:"processed_by,omitempty"`
	Notes       string         `json:"notes,omitempty"`
}

// DividendSnapshot represents a snapshot of shareholders at the record date
type DividendSnapshot struct {
	DividendID       uint64    `json:"dividend_id"`
	CompanyID        uint64    `json:"company_id"`
	ClassID          string    `json:"class_id"`
	RecordDate       time.Time `json:"record_date"`
	TotalShares      math.Int  `json:"total_shares"`
	TotalShareholders uint64   `json:"total_shareholders"`
	SnapshotTaken    bool      `json:"snapshot_taken"`
	CreatedAt        time.Time `json:"created_at"`
}

// ShareholderSnapshot represents a shareholder's position at record date
type ShareholderSnapshot struct {
	DividendID   uint64   `json:"dividend_id"`
	CompanyID    uint64   `json:"company_id"`
	ClassID      string   `json:"class_id"`
	Shareholder  string   `json:"shareholder"`
	SharesHeld   math.Int `json:"shares_held"`
	RecordDate   time.Time `json:"record_date"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewDividend creates a new dividend
func NewDividend(
	id, companyID uint64,
	classID string,
	dividendType DividendType,
	currency string,
	amountPerShare math.LegacyDec,
	declarationDate, exDividendDate, recordDate, paymentDate time.Time,
	declarant string,
) Dividend {
	now := time.Now()
	
	return Dividend{
		ID:              id,
		CompanyID:       companyID,
		ClassID:         classID,
		Type:            dividendType,
		Currency:        currency,
		AmountPerShare:  amountPerShare,
		TotalAmount:     math.LegacyZeroDec(),
		PaidAmount:      math.LegacyZeroDec(),
		RemainingAmount: math.LegacyZeroDec(),
		DeclarationDate: declarationDate,
		ExDividendDate:  exDividendDate,
		RecordDate:      recordDate,
		PaymentDate:     paymentDate,
		Status:          DividendStatusDeclared,
		EligibleShares:  math.ZeroInt(),
		SharesProcessed: math.ZeroInt(),
		ShareholdersEligible: 0,
		ShareholdersPaid:     0,
		StockRatio:      math.LegacyZeroDec(),
		NewSharesIssued: math.ZeroInt(),
		Declarant:       declarant,
		TaxRate:         math.LegacyZeroDec(),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Validate validates a dividend
func (d Dividend) Validate() error {
	if d.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	
	if d.AmountPerShare.IsNegative() {
		return ErrInvalidDividendDate // Reuse error for simplicity
	}
	
	if d.DeclarationDate.IsZero() || d.RecordDate.IsZero() || d.PaymentDate.IsZero() {
		return ErrInvalidDividendDate
	}
	
	if d.ExDividendDate.After(d.RecordDate) {
		return ErrInvalidDividendDate
	}
	
	if d.RecordDate.After(d.PaymentDate) {
		return ErrInvalidDividendDate
	}
	
	if d.Type == DividendTypeCash && d.Currency == "" {
		return ErrInvalidDividendDate // Reuse error
	}
	
	return nil
}

// IsEligible checks if a shareholder is eligible for dividend based on ex-dividend date
func (d Dividend) IsEligible(purchaseDate time.Time) bool {
	return purchaseDate.Before(d.ExDividendDate)
}

// CalculatePayment calculates the dividend payment for a given number of shares
func (d Dividend) CalculatePayment(shares math.Int) math.LegacyDec {
	return d.AmountPerShare.MulInt(shares)
}

// CalculateWithholding calculates tax withholding for a payment
func (d Dividend) CalculateWithholding(grossAmount math.LegacyDec) math.LegacyDec {
	return grossAmount.Mul(d.TaxRate)
}

// CalculateNetPayment calculates net payment after withholding
func (d Dividend) CalculateNetPayment(grossAmount math.LegacyDec) math.LegacyDec {
	withholding := d.CalculateWithholding(grossAmount)
	return grossAmount.Sub(withholding)
}

// IsReadyForPayment checks if dividend is ready for payment processing
func (d Dividend) IsReadyForPayment(currentTime time.Time) bool {
	return d.Status == DividendStatusRecorded && 
		   currentTime.After(d.PaymentDate) ||
		   currentTime.Equal(d.PaymentDate)
}

// IsRecordDateReached checks if record date has been reached
func (d Dividend) IsRecordDateReached(currentTime time.Time) bool {
	return currentTime.After(d.RecordDate) || currentTime.Equal(d.RecordDate)
}

// NewDividendPayment creates a new dividend payment
func NewDividendPayment(
	id, dividendID, companyID uint64,
	classID, shareholder string,
	sharesHeld math.Int,
	amount math.LegacyDec,
	currency string,
) DividendPayment {
	return DividendPayment{
		ID:          id,
		DividendID:  dividendID,
		CompanyID:   companyID,
		ClassID:     classID,
		Shareholder: shareholder,
		SharesHeld:  sharesHeld,
		Amount:      amount,
		Currency:    currency,
		TaxWithheld: math.LegacyZeroDec(),
		NetAmount:   amount,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
}

// NewDividendSnapshot creates a new dividend snapshot
func NewDividendSnapshot(
	dividendID, companyID uint64,
	classID string,
	recordDate time.Time,
) DividendSnapshot {
	return DividendSnapshot{
		DividendID:        dividendID,
		CompanyID:         companyID,
		ClassID:           classID,
		RecordDate:        recordDate,
		TotalShares:       math.ZeroInt(),
		TotalShareholders: 0,
		SnapshotTaken:     false,
		CreatedAt:         time.Now(),
	}
}

// NewShareholderSnapshot creates a new shareholder snapshot
func NewShareholderSnapshot(
	dividendID, companyID uint64,
	classID, shareholder string,
	sharesHeld math.Int,
	recordDate time.Time,
) ShareholderSnapshot {
	return ShareholderSnapshot{
		DividendID:  dividendID,
		CompanyID:   companyID,
		ClassID:     classID,
		Shareholder: shareholder,
		SharesHeld:  sharesHeld,
		RecordDate:  recordDate,
		CreatedAt:   time.Now(),
	}
}

// DividendStatistics represents dividend statistics for reporting
type DividendStatistics struct {
	CompanyID               uint64         `json:"company_id"`
	ClassID                 string         `json:"class_id,omitempty"`
	TotalDividendsDeclared  uint64         `json:"total_dividends_declared"`
	TotalDividendsPaid      uint64         `json:"total_dividends_paid"`
	TotalAmountDeclared     math.LegacyDec `json:"total_amount_declared"`
	TotalAmountPaid         math.LegacyDec `json:"total_amount_paid"`
	AverageYield            math.LegacyDec `json:"average_yield"`
	LastDividendDate        time.Time      `json:"last_dividend_date"`
	LastDividendAmount      math.LegacyDec `json:"last_dividend_amount"`
	DividendFrequency       string         `json:"dividend_frequency"` // "monthly", "quarterly", "annual", etc.
	PayoutRatio             math.LegacyDec `json:"payout_ratio"`       // Dividend/Earnings ratio
	ConsecutivePayments     uint64         `json:"consecutive_payments"`
	DividendGrowthRate      math.LegacyDec `json:"dividend_growth_rate"`
	UpdatedAt               time.Time      `json:"updated_at"`
}

// DividendPolicy represents a company's dividend policy
type DividendPolicy struct {
	CompanyID               uint64         `json:"company_id"`
	Policy                  string         `json:"policy"`                    // "fixed", "variable", "progressive", "stable"
	MinimumPayout           math.LegacyDec `json:"minimum_payout,omitempty"`   // Minimum payout percentage
	MaximumPayout           math.LegacyDec `json:"maximum_payout,omitempty"`   // Maximum payout percentage
	TargetPayoutRatio       math.LegacyDec `json:"target_payout_ratio,omitempty"` // Target payout ratio
	PaymentFrequency        string         `json:"payment_frequency"`          // "monthly", "quarterly", "semi-annual", "annual"
	PaymentMethod           string         `json:"payment_method"`             // "automatic", "manual", "claim"
	ReinvestmentAvailable   bool           `json:"reinvestment_available"`     // DRIP availability
	AutoReinvestmentDefault bool           `json:"auto_reinvestment_default"`  // Default to DRIP
	MinimumSharesForPayment math.Int       `json:"minimum_shares_for_payment,omitempty"` // Minimum shares to receive payment
	FractionalShareHandling string         `json:"fractional_share_handling"`  // "cash", "round", "carry"
	TaxOptimization         bool           `json:"tax_optimization"`           // Optimize for tax efficiency
	Description             string         `json:"description"`
	EffectiveDate           time.Time      `json:"effective_date"`
	LastUpdated             time.Time      `json:"last_updated"`
	UpdatedBy               string         `json:"updated_by"`
}