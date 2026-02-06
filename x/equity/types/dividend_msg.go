package types

import (
	"encoding/json"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types for dividend operations

// MsgDeclareDividend represents a message to declare a dividend
type MsgDeclareDividend struct {
	Creator         string         `json:"creator"`
	CompanyID       uint64         `json:"company_id"`
	ClassID         string         `json:"class_id"`           // Empty for all share classes
	DividendType    string         `json:"dividend_type"`      // "cash", "stock", "property", "special"
	Currency        string         `json:"currency"`           // For cash dividends
	AmountPerShare  math.LegacyDec `json:"amount_per_share"`
	ExDividendDays  uint32         `json:"ex_dividend_days"`   // Days from declaration to ex-dividend
	RecordDays      uint32         `json:"record_days"`        // Days from declaration to record date
	PaymentDays     uint32         `json:"payment_days"`       // Days from declaration to payment date
	TaxRate         math.LegacyDec `json:"tax_rate"`           // Withholding tax rate
	Description     string         `json:"description"`
	PaymentMethod   string         `json:"payment_method"`     // "automatic", "manual", "claim"
	StockRatio      math.LegacyDec `json:"stock_ratio,omitempty"` // For stock dividends: new shares per existing share

	// Audit document (MANDATORY for dividend declaration)
	// This ensures company owners provide proper documentation before distributing dividends
	Audit           AuditInfo      `json:"audit"`              // Audit document information
}

// GetSigners implements the Msg interface
func (msg MsgDeclareDividend) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// ValidateBasic implements the Msg interface
func (msg MsgDeclareDividend) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}

	if msg.AmountPerShare.IsNegative() {
		return ErrInvalidDividendDate
	}

	if msg.DividendType == "cash" && msg.Currency == "" {
		return ErrInvalidDividendDate
	}

	if msg.ExDividendDays == 0 || msg.RecordDays == 0 || msg.PaymentDays == 0 {
		return ErrInvalidDividendDate
	}

	if msg.ExDividendDays >= msg.RecordDays || msg.RecordDays >= msg.PaymentDays {
		return ErrInvalidDividendDate
	}

	// Validate audit document (MANDATORY)
	if err := msg.Audit.Validate(); err != nil {
		return err
	}

	return nil
}

// MsgDeclareDividendResponse is the response type for MsgDeclareDividend
type MsgDeclareDividendResponse struct {
	DividendID uint64 `json:"dividend_id"`
	Success    bool   `json:"success"`
}

// MsgProcessDividend represents a message to process dividend payments
type MsgProcessDividend struct {
	Creator    string `json:"creator"`
	DividendID uint64 `json:"dividend_id"`
	BatchSize  uint32 `json:"batch_size"`  // Number of payments to process in this batch
}

// GetSigners implements the Msg interface
func (msg MsgProcessDividend) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// ValidateBasic implements the Msg interface
func (msg MsgProcessDividend) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.DividendID == 0 {
		return ErrDividendNotFound
	}

	if msg.BatchSize == 0 || msg.BatchSize > 1000 {
		return ErrInvalidDividendDate // Reuse error
	}

	return nil
}

// MsgProcessDividendResponse is the response type for MsgProcessDividend
type MsgProcessDividendResponse struct {
	PaymentsProcessed uint32         `json:"payments_processed"`
	AmountPaid        math.LegacyDec `json:"amount_paid"`
	IsComplete        bool           `json:"is_complete"`
	Success           bool           `json:"success"`
}

// MsgClaimDividend represents a message to claim dividend payment
type MsgClaimDividend struct {
	Creator    string `json:"creator"`
	DividendID uint64 `json:"dividend_id"`
}

// GetSigners implements the Msg interface
func (msg MsgClaimDividend) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// ValidateBasic implements the Msg interface
func (msg MsgClaimDividend) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.DividendID == 0 {
		return ErrDividendNotFound
	}

	return nil
}

// MsgClaimDividendResponse is the response type for MsgClaimDividend
type MsgClaimDividendResponse struct {
	Amount    math.LegacyDec `json:"amount"`
	Currency  string         `json:"currency"`
	TaxWithheld math.LegacyDec `json:"tax_withheld"`
	NetAmount math.LegacyDec `json:"net_amount"`
	Success   bool           `json:"success"`
}

// MsgCancelDividend represents a message to cancel a dividend
type MsgCancelDividend struct {
	Creator    string `json:"creator"`
	DividendID uint64 `json:"dividend_id"`
	Reason     string `json:"reason"`
}

// GetSigners implements the Msg interface
func (msg MsgCancelDividend) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// ValidateBasic implements the Msg interface
func (msg MsgCancelDividend) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.DividendID == 0 {
		return ErrDividendNotFound
	}

	if msg.Reason == "" {
		return ErrInvalidReason
	}

	return nil
}

// MsgCancelDividendResponse is the response type for MsgCancelDividend
type MsgCancelDividendResponse struct {
	Success bool `json:"success"`
}

// MsgSetDividendPolicy represents a message to set company dividend policy
type MsgSetDividendPolicy struct {
	Creator                 string         `json:"creator"`
	CompanyID               uint64         `json:"company_id"`
	Policy                  string         `json:"policy"`
	MinimumPayout           math.LegacyDec `json:"minimum_payout,omitempty"`
	MaximumPayout           math.LegacyDec `json:"maximum_payout,omitempty"`
	TargetPayoutRatio       math.LegacyDec `json:"target_payout_ratio,omitempty"`
	PaymentFrequency        string         `json:"payment_frequency"`
	PaymentMethod           string         `json:"payment_method"`
	ReinvestmentAvailable   bool           `json:"reinvestment_available"`
	AutoReinvestmentDefault bool           `json:"auto_reinvestment_default"`
	MinimumSharesForPayment math.Int       `json:"minimum_shares_for_payment,omitempty"`
	FractionalShareHandling string         `json:"fractional_share_handling"`
	TaxOptimization         bool           `json:"tax_optimization"`
	Description             string         `json:"description"`
}

// GetSigners implements the Msg interface
func (msg MsgSetDividendPolicy) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// ValidateBasic implements the Msg interface
func (msg MsgSetDividendPolicy) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}

	validPolicies := map[string]bool{
		"fixed": true, "variable": true, "progressive": true, "stable": true,
	}
	if !validPolicies[msg.Policy] {
		return ErrInvalidDividendDate // Reuse error
	}

	validFrequencies := map[string]bool{
		"monthly": true, "quarterly": true, "semi-annual": true, "annual": true,
	}
	if !validFrequencies[msg.PaymentFrequency] {
		return ErrInvalidDividendDate
	}

	validMethods := map[string]bool{
		"automatic": true, "manual": true, "claim": true,
	}
	if !validMethods[msg.PaymentMethod] {
		return ErrInvalidDividendDate
	}

	return nil
}

// MsgSetDividendPolicyResponse is the response type for MsgSetDividendPolicy
type MsgSetDividendPolicyResponse struct {
	Success bool `json:"success"`
}

// MsgCreateRecordSnapshot represents a message to create shareholder snapshot
type MsgCreateRecordSnapshot struct {
	Creator    string `json:"creator"`
	DividendID uint64 `json:"dividend_id"`
}

// GetSigners implements the Msg interface
func (msg MsgCreateRecordSnapshot) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// ValidateBasic implements the Msg interface
func (msg MsgCreateRecordSnapshot) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.DividendID == 0 {
		return ErrDividendNotFound
	}

	return nil
}

// MsgCreateRecordSnapshotResponse is the response type for MsgCreateRecordSnapshot
type MsgCreateRecordSnapshotResponse struct {
	TotalShareholders uint64   `json:"total_shareholders"`
	TotalShares       math.Int `json:"total_shares"`
	SnapshotComplete  bool     `json:"snapshot_complete"`
	Success           bool     `json:"success"`
}

// MsgEnableDividendReinvestment represents a message to enable/disable DRIP
type MsgEnableDividendReinvestment struct {
	Creator   string `json:"creator"`
	CompanyID uint64 `json:"company_id"`
	ClassID   string `json:"class_id"`
	Enabled   bool   `json:"enabled"`
}

// GetSigners implements the Msg interface
func (msg MsgEnableDividendReinvestment) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// ValidateBasic implements the Msg interface
func (msg MsgEnableDividendReinvestment) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}

	return nil
}

// MsgEnableDividendReinvestmentResponse is the response type for MsgEnableDividendReinvestment
type MsgEnableDividendReinvestmentResponse struct {
	Success bool `json:"success"`
}

// Helper functions

// NewMsgDeclareDividend creates a new MsgDeclareDividend instance
func NewMsgDeclareDividend(
	creator string,
	companyID uint64,
	classID string,
	dividendType string,
	currency string,
	amountPerShare math.LegacyDec,
	exDividendDays, recordDays, paymentDays uint32,
	taxRate math.LegacyDec,
	description string,
	paymentMethod string,
) *MsgDeclareDividend {
	return &MsgDeclareDividend{
		Creator:        creator,
		CompanyID:      companyID,
		ClassID:        classID,
		DividendType:   dividendType,
		Currency:       currency,
		AmountPerShare: amountPerShare,
		ExDividendDays: exDividendDays,
		RecordDays:     recordDays,
		PaymentDays:    paymentDays,
		TaxRate:        taxRate,
		Description:    description,
		PaymentMethod:  paymentMethod,
	}
}

// String returns a human readable string representation of the message
func (msg MsgDeclareDividend) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// String returns a human readable string representation of the message
func (msg MsgProcessDividend) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// String returns a human readable string representation of the message
func (msg MsgClaimDividend) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// String returns a human readable string representation of the message
func (msg MsgCancelDividend) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// String returns a human readable string representation of the message
func (msg MsgSetDividendPolicy) String() string {
	out, _ := json.Marshal(msg)
	return string(out)
}

// =============================================================================
// Audit Verification Message Response Types
// =============================================================================

// MsgVerifyAuditResponse is the response from verifying an audit
type MsgVerifyAuditResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"` // Current audit status after verification
}

// MsgDisputeAuditResponse is the response from disputing an audit
type MsgDisputeAuditResponse struct {
	Success bool `json:"success"`
}