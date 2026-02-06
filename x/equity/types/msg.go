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

// =============================================================================
// Treasury Share Dilution Messages
// =============================================================================

// SimpleMsgProposeShareDilution proposes adding new shares to treasury (requires governance)
type SimpleMsgProposeShareDilution struct {
	Creator          string         `json:"creator"`
	CompanyID        uint64         `json:"company_id"`
	ClassID          string         `json:"class_id"`
	AdditionalShares math.Int       `json:"additional_shares"`
	Purpose          string         `json:"purpose"`          // Reason for dilution
	ProposalTitle    string         `json:"proposal_title"`   // Governance proposal title
	ProposalDesc     string         `json:"proposal_desc"`    // Governance proposal description
}

// SimpleMsgExecuteShareDilution executes approved dilution (called by governance)
type SimpleMsgExecuteShareDilution struct {
	Authority        string   `json:"authority"`  // Governance module address
	CompanyID        uint64   `json:"company_id"`
	ClassID          string   `json:"class_id"`
	AdditionalShares math.Int `json:"additional_shares"`
}

// Response types

type MsgProposeShareDilutionResponse struct {
	ProposalID uint64 `json:"proposal_id"`
	Success    bool   `json:"success"`
}

type MsgExecuteShareDilutionResponse struct {
	NewTreasuryShares math.Int `json:"new_treasury_shares"`
	TotalAuthorized   math.Int `json:"total_authorized"`
	Success           bool     `json:"success"`
}

// Validation

func (msg SimpleMsgProposeShareDilution) ValidateBasic() error {
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
	if msg.AdditionalShares.IsNil() || msg.AdditionalShares.LTE(math.ZeroInt()) {
		return ErrInsufficientShares
	}
	if msg.Purpose == "" {
		return ErrInvalidReason
	}
	if msg.ProposalTitle == "" || msg.ProposalDesc == "" {
		return ErrInvalidReason
	}
	return nil
}

func (msg SimpleMsgExecuteShareDilution) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	if msg.ClassID == "" {
		return ErrShareClassNotFound
	}
	if msg.AdditionalShares.IsNil() || msg.AdditionalShares.LTE(math.ZeroInt()) {
		return ErrInsufficientShares
	}
	return nil
}

// =============================================================================
// Anti-Dilution Message Types
// =============================================================================

// SimpleMsgRegisterAntiDilution registers anti-dilution protection for a share class
type SimpleMsgRegisterAntiDilution struct {
	Creator          string           `json:"creator"`
	CompanyID        uint64           `json:"company_id"`
	ClassID          string           `json:"class_id"`
	ProvisionType    AntiDilutionType `json:"provision_type"`    // Full ratchet, weighted average, etc.
	TriggerThreshold math.LegacyDec   `json:"trigger_threshold"` // % price drop to trigger (0.0 = any decrease)
	AdjustmentCap    math.LegacyDec   `json:"adjustment_cap"`    // Max adjustment per round (e.g., 0.5 = 50%)
	MinimumPrice     math.LegacyDec   `json:"minimum_price"`     // Floor price for adjustments
}

// SimpleMsgUpdateAntiDilution updates an existing anti-dilution provision
type SimpleMsgUpdateAntiDilution struct {
	Creator          string         `json:"creator"`
	CompanyID        uint64         `json:"company_id"`
	ClassID          string         `json:"class_id"`
	IsActive         *bool          `json:"is_active,omitempty"`         // Enable/disable provision
	TriggerThreshold math.LegacyDec `json:"trigger_threshold,omitempty"` // Updated threshold
	AdjustmentCap    math.LegacyDec `json:"adjustment_cap,omitempty"`    // Updated cap
	MinimumPrice     math.LegacyDec `json:"minimum_price,omitempty"`     // Updated floor
}

// SimpleMsgIssueSharesWithProtection issues shares with anti-dilution processing
type SimpleMsgIssueSharesWithProtection struct {
	Creator      string         `json:"creator"`
	CompanyID    uint64         `json:"company_id"`
	ClassID      string         `json:"class_id"`
	Recipient    string         `json:"recipient"`
	Shares       math.Int       `json:"shares"`
	Price        math.LegacyDec `json:"price"`         // Price per share
	PaymentDenom string         `json:"payment_denom"` // HODL, USDC, etc.
	Purpose      string         `json:"purpose"`       // Reason for issuance
}

// Anti-Dilution Response Types

// MsgRegisterAntiDilutionResponse confirms registration
type MsgRegisterAntiDilutionResponse struct {
	Success       bool   `json:"success"`
	ProvisionType string `json:"provision_type"`
}

// MsgUpdateAntiDilutionResponse confirms update
type MsgUpdateAntiDilutionResponse struct {
	Success bool `json:"success"`
}

// MsgIssueSharesWithProtectionResponse returns issuance info with adjustments
type MsgIssueSharesWithProtectionResponse struct {
	SharesIssued     math.Int       `json:"shares_issued"`
	TotalCost        math.LegacyDec `json:"total_cost"`
	IsDownRound      bool           `json:"is_down_round"`
	AdjustmentsCount uint64         `json:"adjustments_count"` // Number of shareholders adjusted
}

// Anti-Dilution Message Validation

// ValidateBasic validates SimpleMsgRegisterAntiDilution
func (msg SimpleMsgRegisterAntiDilution) ValidateBasic() error {
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
	if msg.ProvisionType < AntiDilutionNone || msg.ProvisionType > AntiDilutionBroadBasedWeightedAverage {
		return ErrInvalidAntiDilutionType
	}
	if msg.ProvisionType == AntiDilutionNone {
		return ErrInvalidAntiDilutionType
	}
	if msg.TriggerThreshold.IsNegative() || msg.TriggerThreshold.GT(math.LegacyOneDec()) {
		return ErrInvalidAntiDilutionType
	}
	if msg.AdjustmentCap.IsNegative() || msg.AdjustmentCap.GT(math.LegacyOneDec()) {
		return ErrExceedsAdjustmentCap
	}
	if msg.MinimumPrice.IsNegative() {
		return ErrBelowMinimumPrice
	}
	return nil
}

// ValidateBasic validates SimpleMsgUpdateAntiDilution
func (msg SimpleMsgUpdateAntiDilution) ValidateBasic() error {
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
	// At least one field should be updated
	if msg.IsActive == nil && msg.TriggerThreshold.IsNil() &&
	   msg.AdjustmentCap.IsNil() && msg.MinimumPrice.IsNil() {
		return ErrAntiDilutionNotFound
	}
	return nil
}

// ValidateBasic validates SimpleMsgIssueSharesWithProtection
func (msg SimpleMsgIssueSharesWithProtection) ValidateBasic() error {
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
		return ErrInvalidInitialPrice
	}
	if msg.PaymentDenom == "" {
		return ErrInsufficientFunds
	}
	return nil
}
// =============================================================================
// Shareholder Blacklist Message Types
// =============================================================================

// SimpleMsgBlacklistShareholder adds a shareholder to the blacklist
type SimpleMsgBlacklistShareholder struct {
	Authority string `json:"authority"` // Company owner or governance
	CompanyID uint64 `json:"company_id"`
	Address   string `json:"address"`
	Reason    string `json:"reason"`
	Duration  int64  `json:"duration"` // Duration in seconds, 0 = permanent
}

// Validate validates the blacklist shareholder message
func (msg SimpleMsgBlacklistShareholder) Validate() error {
	if msg.Authority == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	if msg.Address == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return ErrUnauthorized
	}
	if msg.Reason == "" {
		return ErrInvalidReason
	}
	if msg.Duration < 0 {
		return ErrInvalidDividendDate // Reuse error for simplicity
	}
	return nil
}

// SimpleMsgUnblacklistShareholder removes a shareholder from the blacklist
type SimpleMsgUnblacklistShareholder struct {
	Authority string `json:"authority"` // Company owner or governance
	CompanyID uint64 `json:"company_id"`
	Address   string `json:"address"`
}

// Validate validates the unblacklist shareholder message
func (msg SimpleMsgUnblacklistShareholder) Validate() error {
	if msg.Authority == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	if msg.Address == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Address); err != nil {
		return ErrUnauthorized
	}
	return nil
}

// SimpleMsgSetDividendRedirection configures dividend redirection for blacklisted shareholders
type SimpleMsgSetDividendRedirection struct {
	Authority      string `json:"authority"` // Company owner or governance
	CompanyID      uint64 `json:"company_id"`
	CharityWallet  string `json:"charity_wallet,omitempty"`
	FallbackAction string `json:"fallback_action"` // "community_pool", "burn", "pro_rata"
	Description    string `json:"description,omitempty"`
}

// Validate validates the set dividend redirection message
func (msg SimpleMsgSetDividendRedirection) Validate() error {
	if msg.Authority == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrUnauthorized
	}
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	if msg.CharityWallet != "" {
		if _, err := sdk.AccAddressFromBech32(msg.CharityWallet); err != nil {
			return ErrInvalidCharityWallet
		}
	}
	if msg.FallbackAction != "" &&
	   msg.FallbackAction != "community_pool" &&
	   msg.FallbackAction != "burn" &&
	   msg.FallbackAction != "pro_rata" {
		return ErrInvalidPaymentMethod
	}
	return nil
}

// SimpleMsgSetCharityWallet sets the protocol-level default charity wallet
type SimpleMsgSetCharityWallet struct {
	Authority string `json:"authority"` // Governance only
	Wallet    string `json:"wallet"`
}

// Validate validates the set charity wallet message
func (msg SimpleMsgSetCharityWallet) Validate() error {
	if msg.Authority == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrUnauthorized
	}
	if msg.Wallet == "" {
		return ErrInvalidCharityWallet
	}
	if _, err := sdk.AccAddressFromBech32(msg.Wallet); err != nil {
		return ErrInvalidCharityWallet
	}
	return nil
}

// Response types for blacklist operations

// MsgBlacklistShareholderResponse confirms blacklisting
type MsgBlacklistShareholderResponse struct {
	Success bool `json:"success"`
}

// MsgUnblacklistShareholderResponse confirms unblacklisting
type MsgUnblacklistShareholderResponse struct {
	Success bool `json:"success"`
}

// MsgSetDividendRedirectionResponse confirms redirection configuration
type MsgSetDividendRedirectionResponse struct {
	Success bool `json:"success"`
}

// MsgSetCharityWalletResponse confirms charity wallet configuration
type MsgSetCharityWalletResponse struct {
	Success bool `json:"success"`
}

// =============================================================================
// Treasury Deposit Message Types
// =============================================================================

// SimpleMsgDepositToTreasury deposits funds into company treasury for dividends/operations
type SimpleMsgDepositToTreasury struct {
	CompanyID uint64    `json:"company_id"`
	Depositor string    `json:"depositor"`
	Amount    sdk.Coins `json:"amount"`
}

func (msg SimpleMsgDepositToTreasury) ValidateBasic() error {
	if msg.CompanyID == 0 {
		return ErrCompanyNotFound
	}
	if _, err := sdk.AccAddressFromBech32(msg.Depositor); err != nil {
		return ErrUnauthorized
	}
	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return ErrInvalidWithdrawalAmount
	}
	return nil
}

type MsgDepositToTreasuryResponse struct {
	Success bool `json:"success"`
}

// =============================================================================
// Shareholder Petition Message Types
// =============================================================================

// SimpleMsgCreatePetition creates a new shareholder petition
type SimpleMsgCreatePetition struct {
	Creator      string       `json:"creator"`
	CompanyID    uint64       `json:"company_id"`
	ClassID      string       `json:"class_id"`
	PetitionType PetitionType `json:"petition_type"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
}

// SimpleMsgSignPetition signs/supports a petition
type SimpleMsgSignPetition struct {
	Signer     string `json:"signer"`
	PetitionID uint64 `json:"petition_id"`
	Comment    string `json:"comment,omitempty"`
}

// SimpleMsgWithdrawPetition allows creator to withdraw their petition
type SimpleMsgWithdrawPetition struct {
	Creator    string `json:"creator"`
	PetitionID uint64 `json:"petition_id"`
}

// Response types

type MsgCreatePetitionResponse struct {
	PetitionID uint64 `json:"petition_id"`
	Success    bool   `json:"success"`
}

type MsgSignPetitionResponse struct {
	Success        bool `json:"success"`
	SignatureCount int  `json:"signature_count"`
}

type MsgWithdrawPetitionResponse struct {
	Success bool `json:"success"`
}

// Validation

func (msg SimpleMsgCreatePetition) ValidateBasic() error {
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
	if msg.Title == "" {
		return ErrInvalidReason
	}
	if msg.Description == "" {
		return ErrInvalidReason
	}
	return nil
}

func (msg SimpleMsgSignPetition) ValidateBasic() error {
	if msg.Signer == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return ErrUnauthorized
	}
	if msg.PetitionID == 0 {
		return ErrPetitionNotFound
	}
	return nil
}

func (msg SimpleMsgWithdrawPetition) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.PetitionID == 0 {
		return ErrPetitionNotFound
	}
	return nil
}
