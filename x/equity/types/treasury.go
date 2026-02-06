package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// =============================================================================
// Company Treasury
// =============================================================================

// CompanyTreasury represents the treasury balance and state for a company
type CompanyTreasury struct {
	CompanyID       uint64              `json:"company_id"`
	Balance         sdk.Coins           `json:"balance"`           // Current treasury balance (HODL/USD from share sales)
	TotalDeposited  sdk.Coins           `json:"total_deposited"`   // Total funds deposited (lifetime)
	TotalWithdrawn  sdk.Coins           `json:"total_withdrawn"`   // Total funds withdrawn (lifetime)
	LockedAmount    sdk.Coins           `json:"locked_amount"`     // Amount reserved for pending proposals/dividends
	IsLocked        bool                `json:"is_locked"`         // Emergency lock flag

	// Share holdings in treasury (OWN company shares not yet sold)
	TreasuryShares  map[string]math.Int `json:"treasury_shares"`   // classID -> shares held in treasury
	SharesSoldTotal map[string]math.Int `json:"shares_sold_total"` // classID -> total shares sold from treasury

	// Investment holdings (OTHER companies' equity held as investments)
	// Structure: map[companyID] -> map[classID] -> shares
	// These shares ARE eligible for dividends from those companies
	InvestmentHoldings map[uint64]map[string]math.Int `json:"investment_holdings"`

	// Locked equity for pending withdrawal proposals
	// Structure: map[companyID] -> map[classID] -> locked shares
	LockedEquity map[uint64]map[string]math.Int `json:"locked_equity"`

	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Validate validates a CompanyTreasury
func (t CompanyTreasury) Validate() error {
	if t.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if !t.Balance.IsValid() {
		return fmt.Errorf("invalid balance")
	}
	if !t.TotalDeposited.IsValid() {
		return fmt.Errorf("invalid total deposited")
	}
	if !t.TotalWithdrawn.IsValid() {
		return fmt.Errorf("invalid total withdrawn")
	}
	if !t.LockedAmount.IsValid() {
		return fmt.Errorf("invalid locked amount")
	}
	// Ensure locked amount does not exceed balance
	if t.LockedAmount.IsAnyGT(t.Balance) {
		return fmt.Errorf("locked amount cannot exceed balance")
	}
	// Validate share holdings
	for classID, shares := range t.TreasuryShares {
		if classID == "" {
			return fmt.Errorf("empty class ID in treasury shares")
		}
		if shares.IsNil() || shares.IsNegative() {
			return fmt.Errorf("invalid share count for class %s", classID)
		}
	}
	for classID, shares := range t.SharesSoldTotal {
		if classID == "" {
			return fmt.Errorf("empty class ID in shares sold total")
		}
		if shares.IsNil() || shares.IsNegative() {
			return fmt.Errorf("invalid sold share count for class %s", classID)
		}
	}
	return nil
}

// CanWithdraw checks if the treasury can withdraw the specified amount
func (t CompanyTreasury) CanWithdraw(amount sdk.Coins) bool {
	if t.IsLocked {
		return false
	}
	availableBalance := t.GetAvailableBalance()
	return availableBalance.IsAllGTE(amount)
}

// GetAvailableBalance returns the available balance (balance - locked amount)
func (t CompanyTreasury) GetAvailableBalance() sdk.Coins {
	// Calculate available = balance - locked
	available := t.Balance.Sub(t.LockedAmount...)
	if available.IsAnyNegative() {
		return sdk.NewCoins() // Return empty if calculation results in negative
	}
	return available
}

// NewCompanyTreasury creates a new CompanyTreasury
func NewCompanyTreasury(companyID uint64) CompanyTreasury {
	now := time.Now()
	return CompanyTreasury{
		CompanyID:          companyID,
		Balance:            sdk.NewCoins(),
		TotalDeposited:     sdk.NewCoins(),
		TotalWithdrawn:     sdk.NewCoins(),
		LockedAmount:       sdk.NewCoins(),
		IsLocked:           false,
		TreasuryShares:     make(map[string]math.Int),
		SharesSoldTotal:    make(map[string]math.Int),
		InvestmentHoldings: make(map[uint64]map[string]math.Int),
		LockedEquity:       make(map[uint64]map[string]math.Int),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// GetInvestmentShares returns shares held as investment in another company
func (t CompanyTreasury) GetInvestmentShares(targetCompanyID uint64, classID string) math.Int {
	if t.InvestmentHoldings == nil {
		return math.ZeroInt()
	}
	if holdings, exists := t.InvestmentHoldings[targetCompanyID]; exists {
		if shares, ok := holdings[classID]; ok {
			return shares
		}
	}
	return math.ZeroInt()
}

// GetLockedEquity returns locked equity for a company/class
func (t CompanyTreasury) GetLockedEquity(targetCompanyID uint64, classID string) math.Int {
	if t.LockedEquity == nil {
		return math.ZeroInt()
	}
	if locked, exists := t.LockedEquity[targetCompanyID]; exists {
		if shares, ok := locked[classID]; ok {
			return shares
		}
	}
	return math.ZeroInt()
}

// GetAvailableInvestmentShares returns investment shares minus locked
func (t CompanyTreasury) GetAvailableInvestmentShares(targetCompanyID uint64, classID string) math.Int {
	total := t.GetInvestmentShares(targetCompanyID, classID)
	locked := t.GetLockedEquity(targetCompanyID, classID)
	available := total.Sub(locked)
	if available.IsNegative() {
		return math.ZeroInt()
	}
	return available
}

// GetAllInvestmentHoldings returns all investment holdings as a flat list
func (t CompanyTreasury) GetAllInvestmentHoldings() []InvestmentHolding {
	holdings := []InvestmentHolding{}
	if t.InvestmentHoldings == nil {
		return holdings
	}
	for companyID, classMap := range t.InvestmentHoldings {
		for classID, shares := range classMap {
			if shares.GT(math.ZeroInt()) {
				holdings = append(holdings, InvestmentHolding{
					OwnerCompanyID:  t.CompanyID,
					TargetCompanyID: companyID,
					ClassID:         classID,
					Shares:          shares,
				})
			}
		}
	}
	return holdings
}

// InvestmentHolding represents a treasury's investment in another company
type InvestmentHolding struct {
	OwnerCompanyID  uint64   `json:"owner_company_id"`  // Company that owns this investment
	TargetCompanyID uint64   `json:"target_company_id"` // Company whose shares are held
	ClassID         string   `json:"class_id"`          // Share class
	Shares          math.Int `json:"shares"`            // Number of shares held
}

// GetTreasuryShares returns the number of shares held in treasury for a class
func (t CompanyTreasury) GetTreasuryShares(classID string) math.Int {
	if shares, exists := t.TreasuryShares[classID]; exists {
		return shares
	}
	return math.ZeroInt()
}

// GetSharesSold returns the total number of shares sold from treasury for a class
func (t CompanyTreasury) GetSharesSold(classID string) math.Int {
	if shares, exists := t.SharesSoldTotal[classID]; exists {
		return shares
	}
	return math.ZeroInt()
}

// CanSellShares checks if treasury has enough shares to sell
func (t CompanyTreasury) CanSellShares(classID string, amount math.Int) bool {
	available := t.GetTreasuryShares(classID)
	return available.GTE(amount)
}

// =============================================================================
// Treasury Withdrawal
// =============================================================================

// WithdrawalStatus represents the status of a treasury withdrawal
type WithdrawalStatus int32

const (
	WithdrawalStatusPending   WithdrawalStatus = iota // Awaiting approval
	WithdrawalStatusApproved                          // Approved by governance
	WithdrawalStatusRejected                          // Rejected by governance
	WithdrawalStatusExecuted                          // Funds transferred
	WithdrawalStatusCancelled                         // Cancelled by requester
)

// String returns the string representation of WithdrawalStatus
func (s WithdrawalStatus) String() string {
	switch s {
	case WithdrawalStatusPending:
		return "pending"
	case WithdrawalStatusApproved:
		return "approved"
	case WithdrawalStatusRejected:
		return "rejected"
	case WithdrawalStatusExecuted:
		return "executed"
	case WithdrawalStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// TreasuryWithdrawal represents a request to withdraw funds from company treasury
type TreasuryWithdrawal struct {
	ID          uint64           `json:"id"`
	CompanyID   uint64           `json:"company_id"`
	ProposalID  uint64           `json:"proposal_id,omitempty"` // Associated governance proposal ID
	Recipient   string           `json:"recipient"`             // Bech32 address
	Amount      sdk.Coins        `json:"amount"`                // Amount to withdraw
	Purpose     string           `json:"purpose"`               // Reason for withdrawal
	Status      WithdrawalStatus `json:"status"`
	RequestedBy string           `json:"requested_by"`          // Bech32 address of requester
	RequestedAt time.Time        `json:"requested_at"`
	ExecutedAt  time.Time        `json:"executed_at,omitempty"`
}

// Validate validates a TreasuryWithdrawal
func (w TreasuryWithdrawal) Validate() error {
	if w.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if w.Recipient == "" {
		return fmt.Errorf("recipient cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(w.Recipient); err != nil {
		return fmt.Errorf("invalid recipient address: %v", err)
	}
	if !w.Amount.IsValid() || w.Amount.IsZero() {
		return fmt.Errorf("amount must be valid and non-zero")
	}
	if w.Amount.IsAnyNegative() {
		return fmt.Errorf("amount cannot be negative")
	}
	if w.Purpose == "" {
		return fmt.Errorf("purpose cannot be empty")
	}
	if w.RequestedBy == "" {
		return fmt.Errorf("requested by cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(w.RequestedBy); err != nil {
		return fmt.Errorf("invalid requester address: %v", err)
	}
	return nil
}

// IsExecutable checks if the withdrawal can be executed
func (w TreasuryWithdrawal) IsExecutable() bool {
	return w.Status == WithdrawalStatusApproved && w.ExecutedAt.IsZero()
}

// NewTreasuryWithdrawal creates a new treasury withdrawal request
func NewTreasuryWithdrawal(
	id uint64,
	companyID uint64,
	recipient string,
	amount sdk.Coins,
	purpose string,
	requestedBy string,
) TreasuryWithdrawal {
	return TreasuryWithdrawal{
		ID:          id,
		CompanyID:   companyID,
		ProposalID:  0, // Will be set when governance proposal is created
		Recipient:   recipient,
		Amount:      amount,
		Purpose:     purpose,
		Status:      WithdrawalStatusPending,
		RequestedBy: requestedBy,
		RequestedAt: time.Now(),
	}
}

// =============================================================================
// Equity Withdrawal Request (for treasury investment holdings)
// =============================================================================

// EquityWithdrawalRequest represents a request to withdraw equity from treasury
// This requires governance vote like HODL withdrawals
type EquityWithdrawalRequest struct {
	ID              uint64           `json:"id"`
	CompanyID       uint64           `json:"company_id"`        // Treasury owner company
	TargetCompanyID uint64           `json:"target_company_id"` // Company whose shares to withdraw
	ClassID         string           `json:"class_id"`          // Share class
	Shares          math.Int         `json:"shares"`            // Number of shares to withdraw
	Recipient       string           `json:"recipient"`         // Who receives the shares
	Purpose         string           `json:"purpose"`           // Reason for withdrawal
	ProposalID      uint64           `json:"proposal_id,omitempty"`
	Status          WithdrawalStatus `json:"status"`
	RequestedBy     string           `json:"requested_by"`
	RequestedAt     time.Time        `json:"requested_at"`
	ExecutedAt      time.Time        `json:"executed_at,omitempty"`
}

// Validate validates an EquityWithdrawalRequest
func (r EquityWithdrawalRequest) Validate() error {
	if r.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if r.TargetCompanyID == 0 {
		return fmt.Errorf("target company ID cannot be zero")
	}
	if r.ClassID == "" {
		return fmt.Errorf("class ID cannot be empty")
	}
	if r.Shares.IsNil() || r.Shares.LTE(math.ZeroInt()) {
		return fmt.Errorf("shares must be positive")
	}
	if r.Recipient == "" {
		return fmt.Errorf("recipient cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(r.Recipient); err != nil {
		return fmt.Errorf("invalid recipient address: %v", err)
	}
	if r.Purpose == "" {
		return fmt.Errorf("purpose cannot be empty")
	}
	if r.RequestedBy == "" {
		return fmt.Errorf("requested by cannot be empty")
	}
	return nil
}

// NewEquityWithdrawalRequest creates a new equity withdrawal request
func NewEquityWithdrawalRequest(
	id uint64,
	companyID uint64,
	targetCompanyID uint64,
	classID string,
	shares math.Int,
	recipient string,
	purpose string,
	requestedBy string,
) EquityWithdrawalRequest {
	return EquityWithdrawalRequest{
		ID:              id,
		CompanyID:       companyID,
		TargetCompanyID: targetCompanyID,
		ClassID:         classID,
		Shares:          shares,
		Recipient:       recipient,
		Purpose:         purpose,
		Status:          WithdrawalStatusPending,
		RequestedBy:     requestedBy,
		RequestedAt:     time.Now(),
	}
}

// =============================================================================
// Primary Sale Offer
// =============================================================================

// OfferStatus represents the status of a primary sale offer
type OfferStatus int32

const (
	OfferStatusDraft     OfferStatus = iota // Being prepared
	OfferStatusOpen                         // Accepting purchases
	OfferStatusClosed                       // Sale period ended
	OfferStatusCancelled                    // Cancelled by company
)

// String returns the string representation of OfferStatus
func (s OfferStatus) String() string {
	switch s {
	case OfferStatusDraft:
		return "draft"
	case OfferStatusOpen:
		return "open"
	case OfferStatusClosed:
		return "closed"
	case OfferStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// PrimarySaleOffer represents a primary share offering from company treasury
type PrimarySaleOffer struct {
	ID             uint64         `json:"id"`
	CompanyID      uint64         `json:"company_id"`
	ClassID        string         `json:"class_id"`          // Share class being offered
	SharesOffered  math.Int       `json:"shares_offered"`    // Total shares available
	SharesSold     math.Int       `json:"shares_sold"`       // Shares sold so far
	PricePerShare  math.LegacyDec `json:"price_per_share"`   // Price per share
	QuoteDenom     string         `json:"quote_denom"`       // Currency (e.g., "uhodl")
	MinPurchase    math.Int       `json:"min_purchase"`      // Minimum shares per purchase
	MaxPerAddress  math.Int       `json:"max_per_address"`   // Maximum shares per address (0 = unlimited)
	StartTime      time.Time      `json:"start_time"`        // Sale start time
	EndTime        time.Time      `json:"end_time"`          // Sale end time
	Status         OfferStatus    `json:"status"`
	CreatedBy      string         `json:"created_by"`        // Bech32 address
	CreatedAt      time.Time      `json:"created_at"`
}

// Validate validates a PrimarySaleOffer
func (o PrimarySaleOffer) Validate() error {
	if o.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if o.ClassID == "" {
		return fmt.Errorf("class ID cannot be empty")
	}
	if o.SharesOffered.IsNil() || o.SharesOffered.LTE(math.ZeroInt()) {
		return fmt.Errorf("shares offered must be positive")
	}
	if o.SharesSold.IsNil() || o.SharesSold.IsNegative() {
		return fmt.Errorf("shares sold cannot be negative")
	}
	if o.SharesSold.GT(o.SharesOffered) {
		return fmt.Errorf("shares sold cannot exceed shares offered")
	}
	if o.PricePerShare.IsNil() || o.PricePerShare.LTE(math.LegacyZeroDec()) {
		return fmt.Errorf("price per share must be positive")
	}
	if o.QuoteDenom == "" {
		return fmt.Errorf("quote denomination cannot be empty")
	}
	if o.MinPurchase.IsNil() || o.MinPurchase.IsNegative() {
		return fmt.Errorf("minimum purchase cannot be negative")
	}
	if o.MaxPerAddress.IsNil() || o.MaxPerAddress.IsNegative() {
		return fmt.Errorf("maximum per address cannot be negative")
	}
	if !o.MaxPerAddress.IsZero() && o.MaxPerAddress.LT(o.MinPurchase) {
		return fmt.Errorf("maximum per address must be >= minimum purchase")
	}
	if o.EndTime.Before(o.StartTime) {
		return fmt.Errorf("end time must be after start time")
	}
	if o.CreatedBy == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(o.CreatedBy); err != nil {
		return fmt.Errorf("invalid creator address: %v", err)
	}
	return nil
}

// RemainingShares returns the number of shares still available for purchase
func (o PrimarySaleOffer) RemainingShares() math.Int {
	remaining := o.SharesOffered.Sub(o.SharesSold)
	if remaining.IsNegative() {
		return math.ZeroInt()
	}
	return remaining
}

// IsActive checks if the offer is currently active for purchases
func (o PrimarySaleOffer) IsActive(currentTime time.Time) bool {
	if o.Status != OfferStatusOpen {
		return false
	}
	if currentTime.Before(o.StartTime) {
		return false
	}
	if currentTime.After(o.EndTime) {
		return false
	}
	// Check if shares are still available
	return o.RemainingShares().GT(math.ZeroInt())
}

// NewPrimarySaleOffer creates a new primary sale offer
func NewPrimarySaleOffer(
	id uint64,
	companyID uint64,
	classID string,
	sharesOffered math.Int,
	pricePerShare math.LegacyDec,
	quoteDenom string,
	minPurchase math.Int,
	maxPerAddress math.Int,
	startTime time.Time,
	endTime time.Time,
	createdBy string,
) PrimarySaleOffer {
	return PrimarySaleOffer{
		ID:            id,
		CompanyID:     companyID,
		ClassID:       classID,
		SharesOffered: sharesOffered,
		SharesSold:    math.ZeroInt(),
		PricePerShare: pricePerShare,
		QuoteDenom:    quoteDenom,
		MinPurchase:   minPurchase,
		MaxPerAddress: maxPerAddress,
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        OfferStatusDraft,
		CreatedBy:     createdBy,
		CreatedAt:     time.Now(),
	}
}

// =============================================================================
// Event Types
// =============================================================================

const (
	EventTypeTreasuryCreated            = "treasury_created"
	EventTypeTreasuryDeposit            = "treasury_deposit"
	EventTypeTreasuryWithdrawalRequest  = "treasury_withdrawal_request"
	EventTypeTreasuryWithdrawalApproved = "treasury_withdrawal_approved"
	EventTypeTreasuryWithdrawalExecuted = "treasury_withdrawal_executed"
	EventTypeTreasuryWithdrawalRejected = "treasury_withdrawal_rejected"
	EventTypeTreasuryWithdrawalCancelled = "treasury_withdrawal_cancelled"
	EventTypeTreasuryLocked             = "treasury_locked"
	EventTypeTreasuryUnlocked           = "treasury_unlocked"
	EventTypePrimarySaleCreated         = "primary_sale_created"
	EventTypePrimarySaleOpened          = "primary_sale_opened"
	EventTypePrimarySalePurchase        = "primary_sale_purchase"
	EventTypePrimarySaleClosed          = "primary_sale_closed"
)

// Attribute keys for treasury events
const (
	AttributeKeyTreasuryCompanyID     = "company_id"
	AttributeKeyTreasuryAmount        = "amount"
	AttributeKeyTreasuryBalance       = "balance"
	AttributeKeyTreasuryDenom         = "treasury_denom"
	AttributeKeyDepositAmount         = "deposit_amount"
	AttributeKeyDepositor             = "depositor"
	AttributeKeyWithdrawalID          = "withdrawal_id"
	AttributeKeyWithdrawalRecipient   = "recipient"
	AttributeKeyWithdrawalAmount      = "withdrawal_amount"
	AttributeKeyWithdrawalPurpose     = "purpose"
	AttributeKeyWithdrawalProposer    = "withdrawal_proposer"
	AttributeKeyWithdrawalStatus      = "status"
	AttributeKeyLockReason            = "lock_reason"
	AttributeKeyLockedUntil           = "locked_until"
	AttributeKeyOfferID               = "offer_id"
	AttributeKeyOfferClassID          = "class_id"
	AttributeKeySharesOffered         = "shares_offered"
	AttributeKeySharesSold            = "shares_sold"
	AttributeKeySharesRemaining       = "shares_remaining"
	AttributeKeyPricePerShare         = "price_per_share"
	AttributeKeyOfferPrice            = "price_per_share" // Alias for compatibility
	AttributeKeyPurchaser             = "purchaser"
	AttributeKeyOfferBuyer            = "buyer" // Alias for compatibility
	AttributeKeyPurchaseAmount        = "purchase_amount"
	AttributeKeyOfferSharesPurchased  = "shares_purchased" // Alias for compatibility
	AttributeKeyPurchaseCost          = "purchase_cost"
	AttributeKeyTotalRevenue          = "total_revenue"
	AttributeKeyTotalPurchasers       = "total_purchasers"
	AttributeKeyOfferStartTime        = "offer_start_time"
	AttributeKeyOfferEndTime          = "offer_end_time"
	AttributeKeyOfferStatus           = "offer_status"
	AttributeKeyShareClass            = "share_class"
)
