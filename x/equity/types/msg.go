package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types for simplified implementation (avoiding protobuf for now)

// SimpleMsgCreateCompany creates a new company listing
type SimpleMsgCreateCompany struct {
	Creator          string          `json:"creator"`
	Name             string          `json:"name"`
	Symbol           string          `json:"symbol"`
	Description      string          `json:"description"`
	Website          string          `json:"website"`
	Industry         string          `json:"industry"`
	Country          string          `json:"country"`
	TotalShares      math.Int        `json:"total_shares"`
	AuthorizedShares math.Int        `json:"authorized_shares"`
	ParValue         math.LegacyDec  `json:"par_value"`
	ListingFee       sdk.Coins       `json:"listing_fee"`
}

// SimpleMsgCreateShareClass creates a new share class
type SimpleMsgCreateShareClass struct {
	Creator               string     `json:"creator"`
	CompanyID             uint64     `json:"company_id"`
	ClassID               string     `json:"class_id"`
	Name                  string     `json:"name"`
	Description           string     `json:"description"`
	VotingRights          bool       `json:"voting_rights"`
	DividendRights        bool       `json:"dividend_rights"`
	LiquidationPreference bool       `json:"liquidation_preference"`
	ConversionRights      bool       `json:"conversion_rights"`
	AuthorizedShares      math.Int   `json:"authorized_shares"`
	ParValue              math.LegacyDec `json:"par_value"`
	Transferable          bool       `json:"transferable"`
}

// SimpleMsgIssueShares issues shares to an address
type SimpleMsgIssueShares struct {
	Creator     string          `json:"creator"`
	CompanyID   uint64          `json:"company_id"`
	ClassID     string          `json:"class_id"`
	Recipient   string          `json:"recipient"`
	Shares      math.Int        `json:"shares"`
	Price       math.LegacyDec  `json:"price"`       // Price per share
	PaymentDenom string         `json:"payment_denom"` // HODL, USDC, etc.
}

// SimpleMsgTransferShares transfers shares between addresses
type SimpleMsgTransferShares struct {
	Creator   string   `json:"creator"`
	CompanyID uint64   `json:"company_id"`
	ClassID   string   `json:"class_id"`
	From      string   `json:"from"`
	To        string   `json:"to"`
	Shares    math.Int `json:"shares"`
}

// SimpleMsgUpdateCompany updates company information (founder only)
type SimpleMsgUpdateCompany struct {
	Creator     string `json:"creator"`
	CompanyID   uint64 `json:"company_id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Website     string `json:"website,omitempty"`
	Industry    string `json:"industry,omitempty"`
	Country     string `json:"country,omitempty"`
}

// Response types

// MsgCreateCompanyResponse returns the company ID
type MsgCreateCompanyResponse struct {
	CompanyID uint64 `json:"company_id"`
}

// MsgCreateShareClassResponse confirms share class creation
type MsgCreateShareClassResponse struct {
	Success bool `json:"success"`
}

// MsgIssueSharesResponse returns shares issued info
type MsgIssueSharesResponse struct {
	SharesIssued math.Int       `json:"shares_issued"`
	TotalCost    math.LegacyDec `json:"total_cost"`
}

// MsgTransferSharesResponse confirms transfer
type MsgTransferSharesResponse struct {
	Success bool `json:"success"`
}

// MsgUpdateCompanyResponse confirms update
type MsgUpdateCompanyResponse struct {
	Success bool `json:"success"`
}

// Message validation methods

// ValidateBasic validates SimpleMsgCreateCompany
func (msg SimpleMsgCreateCompany) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.Name == "" {
		return ErrInvalidSymbol
	}
	if msg.Symbol == "" {
		return ErrInvalidSymbol
	}
	if msg.TotalShares.IsNil() || msg.TotalShares.LTE(math.ZeroInt()) {
		return ErrInsufficientShares
	}
	if msg.AuthorizedShares.IsNil() || msg.AuthorizedShares.LT(msg.TotalShares) {
		return ErrExceedsAuthorized
	}
	if msg.ParValue.IsNegative() {
		return ErrInvalidSymbol
	}
	if !msg.ListingFee.IsValid() {
		return ErrInsufficientFunds
	}
	return nil
}

// ValidateBasic validates SimpleMsgCreateShareClass
func (msg SimpleMsgCreateShareClass) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	if msg.ClassID == "" {
		return ErrShareClassNotFound
	}
	if msg.Name == "" {
		return ErrInvalidSymbol
	}
	if msg.AuthorizedShares.IsNil() || msg.AuthorizedShares.LTE(math.ZeroInt()) {
		return ErrInsufficientShares
	}
	if msg.ParValue.IsNegative() {
		return ErrInvalidSymbol
	}
	return nil
}

// ValidateBasic validates SimpleMsgIssueShares
func (msg SimpleMsgIssueShares) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	if msg.ClassID == "" {
		return ErrShareClassNotFound
	}
	if msg.Recipient == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return ErrUnauthorized
	}
	if msg.Shares.IsNil() || msg.Shares.LTE(math.ZeroInt()) {
		return ErrInsufficientShares
	}
	if msg.Price.IsNegative() {
		return ErrInvalidSymbol
	}
	if msg.PaymentDenom == "" {
		return ErrInsufficientFunds
	}
	return nil
}

// ValidateBasic validates SimpleMsgTransferShares
func (msg SimpleMsgTransferShares) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	if msg.ClassID == "" {
		return ErrShareClassNotFound
	}
	if msg.From == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.From); err != nil {
		return ErrUnauthorized
	}
	if msg.To == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.To); err != nil {
		return ErrUnauthorized
	}
	if msg.From == msg.To {
		return ErrTransferRestricted
	}
	if msg.Shares.IsNil() || msg.Shares.LTE(math.ZeroInt()) {
		return ErrInsufficientShares
	}
	return nil
}

// ValidateBasic validates SimpleMsgUpdateCompany
func (msg SimpleMsgUpdateCompany) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	// At least one field should be updated
	if msg.Name == "" && msg.Description == "" && msg.Website == "" && 
	   msg.Industry == "" && msg.Country == "" {
		return ErrUnauthorized
	}
	return nil
}