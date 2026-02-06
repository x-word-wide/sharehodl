package types

import (
	"fmt"
	"time"
	
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Company represents a business entity on ShareHODL
type Company struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Symbol      string    `json:"symbol"` // Trading symbol (e.g., "APPLE", "TESLA")
	Description string    `json:"description"`
	Founder     string    `json:"founder"`     // Creator address
	Website     string    `json:"website"`
	Industry    string    `json:"industry"`
	Country     string    `json:"country"`
	Status      CompanyStatus `json:"status"`
	Verified    bool      `json:"verified"`   // Verified by validators
	ListingFee  sdk.Coins `json:"listing_fee"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Equity details
	TotalShares     math.Int `json:"total_shares"`
	AuthorizedShares math.Int `json:"authorized_shares"`
	ParValue        math.LegacyDec `json:"par_value"` // Par value per share

	// Financial info (optional)
	MarketCap       math.Int `json:"market_cap,omitempty"`
	LastValuation   math.Int `json:"last_valuation,omitempty"`
	ValuationDate   time.Time `json:"valuation_date,omitempty"`

	// Authorized responders (Issue #11)
	AuthorizedResponders []string `json:"authorized_responders,omitempty"` // Addresses that can respond to warnings
}

// CompanyStatus represents the current status of a company
type CompanyStatus int32

const (
	CompanyStatusPending CompanyStatus = iota // Awaiting verification
	CompanyStatusActive                       // Verified and trading
	CompanyStatusSuspended                    // Temporarily suspended
	CompanyStatusDelisted                     // Removed from trading
)

func (cs CompanyStatus) String() string {
	switch cs {
	case CompanyStatusPending:
		return "pending"
	case CompanyStatusActive:
		return "active"
	case CompanyStatusSuspended:
		return "suspended"
	case CompanyStatusDelisted:
		return "delisted"
	default:
		return "unknown"
	}
}

// ShareClass represents a class of shares (e.g., Common, Preferred)
type ShareClass struct {
	CompanyID     uint64 `json:"company_id"`
	ClassID       string `json:"class_id"`     // e.g., "COMMON", "PREFERRED_A"
	Name          string `json:"name"`         // Human readable name
	Description   string `json:"description"`

	// Share characteristics
	VotingRights  bool   `json:"voting_rights"`
	DividendRights bool  `json:"dividend_rights"`
	LiquidationPreference bool `json:"liquidation_preference"`
	ConversionRights bool `json:"conversion_rights"`

	// Share counts
	AuthorizedShares math.Int `json:"authorized_shares"`
	IssuedShares     math.Int `json:"issued_shares"`
	OutstandingShares math.Int `json:"outstanding_shares"` // Issued - Treasury

	// Pricing
	ParValue      math.LegacyDec `json:"par_value"`
	CurrentPrice  math.LegacyDec `json:"current_price,omitempty"`

	// Transfer restrictions
	Transferable  bool     `json:"transferable"`
	RestrictedTo  []string `json:"restricted_to,omitempty"` // List of allowed holders
	LockupPeriod  time.Duration `json:"lockup_period,omitempty"`

	// Anti-dilution protection
	HasAntiDilution    bool             `json:"has_anti_dilution"`            // Whether this class has protection
	AntiDilutionType   AntiDilutionType `json:"anti_dilution_type,omitempty"` // Type of protection

	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Shareholding represents ownership of shares by an address
type Shareholding struct {
	CompanyID     uint64   `json:"company_id"`
	ClassID       string   `json:"class_id"`
	Owner         string   `json:"owner"`      // Bech32 address
	
	// Share ownership
	Shares        math.Int `json:"shares"`      // Total shares owned
	VestedShares  math.Int `json:"vested_shares"` // Shares that are vested
	LockedShares  math.Int `json:"locked_shares"` // Shares under lockup
	
	// Purchase info
	CostBasis     math.LegacyDec `json:"cost_basis"`     // Average cost per share
	TotalCost     math.LegacyDec `json:"total_cost"`     // Total amount paid
	AcquisitionDate time.Time `json:"acquisition_date"`
	
	// Restrictions
	TransferRestricted bool      `json:"transfer_restricted"`
	LockupExpiry      time.Time  `json:"lockup_expiry,omitempty"`
	
	UpdatedAt        time.Time  `json:"updated_at"`
}

// VestingSchedule represents a vesting schedule for shares
type VestingSchedule struct {
	CompanyID     uint64    `json:"company_id"`
	ClassID       string    `json:"class_id"`
	Owner         string    `json:"owner"`
	
	// Vesting details
	TotalShares   math.Int  `json:"total_shares"`   // Total shares to vest
	VestedShares  math.Int  `json:"vested_shares"`  // Currently vested shares
	StartDate     time.Time `json:"start_date"`     // When vesting starts
	CliffDate     time.Time `json:"cliff_date"`     // Cliff vesting date
	EndDate       time.Time `json:"end_date"`       // When fully vested
	
	// Vesting parameters
	CliffPercent  math.LegacyDec `json:"cliff_percent"`  // % vested at cliff
	VestingPeriod time.Duration `json:"vesting_period"` // How often vesting occurs
	
	CreatedAt     time.Time `json:"created_at"`
	LastVested    time.Time `json:"last_vested"`
}

// Validation methods

// Validate validates a Company
func (c Company) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("company name cannot be empty")
	}
	if c.Symbol == "" {
		return fmt.Errorf("company symbol cannot be empty")
	}
	if c.Founder == "" {
		return fmt.Errorf("company founder cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(c.Founder); err != nil {
		return fmt.Errorf("invalid founder address: %v", err)
	}
	if c.TotalShares.IsNil() || c.TotalShares.LTE(math.ZeroInt()) {
		return fmt.Errorf("total shares must be positive")
	}
	if c.AuthorizedShares.IsNil() || c.AuthorizedShares.LT(c.TotalShares) {
		return fmt.Errorf("authorized shares must be >= total shares")
	}
	if c.ParValue.IsNegative() {
		return fmt.Errorf("par value cannot be negative")
	}
	return nil
}

// Validate validates a ShareClass
func (sc ShareClass) Validate() error {
	if sc.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if sc.ClassID == "" {
		return fmt.Errorf("class ID cannot be empty")
	}
	if sc.Name == "" {
		return fmt.Errorf("class name cannot be empty")
	}
	if sc.AuthorizedShares.IsNil() || sc.AuthorizedShares.LTE(math.ZeroInt()) {
		return fmt.Errorf("authorized shares must be positive")
	}
	if sc.IssuedShares.IsNil() || sc.IssuedShares.GT(sc.AuthorizedShares) {
		return fmt.Errorf("issued shares cannot exceed authorized shares")
	}
	if sc.OutstandingShares.IsNil() || sc.OutstandingShares.GT(sc.IssuedShares) {
		return fmt.Errorf("outstanding shares cannot exceed issued shares")
	}
	if sc.ParValue.IsNegative() {
		return fmt.Errorf("par value cannot be negative")
	}
	return nil
}

// Validate validates a Shareholding
func (sh Shareholding) Validate() error {
	if sh.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if sh.ClassID == "" {
		return fmt.Errorf("class ID cannot be empty")
	}
	if sh.Owner == "" {
		return fmt.Errorf("owner cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(sh.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if sh.Shares.IsNil() || sh.Shares.LTE(math.ZeroInt()) {
		return fmt.Errorf("shares must be positive")
	}
	if sh.VestedShares.IsNil() || sh.VestedShares.GT(sh.Shares) {
		return fmt.Errorf("vested shares cannot exceed total shares")
	}
	if sh.LockedShares.IsNil() || sh.LockedShares.GT(sh.Shares) {
		return fmt.Errorf("locked shares cannot exceed total shares")
	}
	return nil
}

// NewCompany creates a new Company
func NewCompany(
	id uint64,
	name, symbol, description, founder, website, industry, country string,
	totalShares, authorizedShares math.Int,
	parValue math.LegacyDec,
	listingFee sdk.Coins,
) Company {
	now := time.Now()
	return Company{
		ID:               id,
		Name:             name,
		Symbol:           symbol,
		Description:      description,
		Founder:          founder,
		Website:          website,
		Industry:         industry,
		Country:          country,
		Status:           CompanyStatusPending,
		Verified:         false,
		ListingFee:       listingFee,
		TotalShares:      totalShares,
		AuthorizedShares: authorizedShares,
		ParValue:         parValue,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// NewShareClass creates a new ShareClass
func NewShareClass(
	companyID uint64,
	classID, name, description string,
	votingRights, dividendRights, liquidationPreference, conversionRights bool,
	authorizedShares math.Int,
	parValue math.LegacyDec,
	transferable bool,
) ShareClass {
	now := time.Now()
	return ShareClass{
		CompanyID:            companyID,
		ClassID:              classID,
		Name:                 name,
		Description:          description,
		VotingRights:         votingRights,
		DividendRights:       dividendRights,
		LiquidationPreference: liquidationPreference,
		ConversionRights:     conversionRights,
		AuthorizedShares:     authorizedShares,
		IssuedShares:         math.ZeroInt(),
		OutstandingShares:    math.ZeroInt(),
		ParValue:             parValue,
		Transferable:         transferable,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// NewShareholding creates a new Shareholding
func NewShareholding(
	companyID uint64,
	classID, owner string,
	shares math.Int,
	costBasis math.LegacyDec,
) Shareholding {
	now := time.Now()
	totalCost := costBasis.MulInt(shares)

	return Shareholding{
		CompanyID:       companyID,
		ClassID:         classID,
		Owner:           owner,
		Shares:          shares,
		VestedShares:    shares, // Assume fully vested by default
		LockedShares:    math.ZeroInt(),
		CostBasis:       costBasis,
		TotalCost:       totalCost,
		AcquisitionDate: now,
		UpdatedAt:       now,
	}
}

// =============================================================================
// Anti-Dilution Rights Types
// =============================================================================

// AntiDilutionType represents the type of anti-dilution protection
type AntiDilutionType int32

const (
	// AntiDilutionNone - No anti-dilution protection
	AntiDilutionNone AntiDilutionType = iota
	// AntiDilutionFullRatchet - Full ratchet protection (most favorable to investors)
	// Adjusts conversion price to the new lower price
	AntiDilutionFullRatchet
	// AntiDilutionWeightedAverage - Weighted average protection (more balanced)
	// Adjusts based on how many shares were issued at the lower price
	AntiDilutionWeightedAverage
	// AntiDilutionBroadBasedWeightedAverage - Broad-based weighted average (most common)
	// Includes all common shares, options, warrants in the calculation
	AntiDilutionBroadBasedWeightedAverage
)

func (t AntiDilutionType) String() string {
	switch t {
	case AntiDilutionNone:
		return "none"
	case AntiDilutionFullRatchet:
		return "full_ratchet"
	case AntiDilutionWeightedAverage:
		return "weighted_average"
	case AntiDilutionBroadBasedWeightedAverage:
		return "broad_based_weighted_average"
	default:
		return "unknown"
	}
}

// AntiDilutionProvision defines anti-dilution protection for a share class
type AntiDilutionProvision struct {
	CompanyID        uint64           `json:"company_id"`
	ClassID          string           `json:"class_id"`           // Protected share class
	ProvisionType    AntiDilutionType `json:"provision_type"`     // Type of protection

	// Protection parameters
	TriggerThreshold math.LegacyDec   `json:"trigger_threshold"`  // % below last price to trigger (e.g., 0.0 = any decrease)
	AdjustmentCap    math.LegacyDec   `json:"adjustment_cap"`     // Max adjustment per round (e.g., 0.5 = 50% max)
	MinimumPrice     math.LegacyDec   `json:"minimum_price"`      // Floor price for adjustments

	// State
	IsActive         bool             `json:"is_active"`
	LastTriggerDate  time.Time        `json:"last_trigger_date,omitempty"`
	TotalAdjustments uint64           `json:"total_adjustments"`  // Count of adjustments applied

	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// Validate validates an AntiDilutionProvision
func (p AntiDilutionProvision) Validate() error {
	if p.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if p.ClassID == "" {
		return fmt.Errorf("class ID cannot be empty")
	}
	if p.ProvisionType < AntiDilutionNone || p.ProvisionType > AntiDilutionBroadBasedWeightedAverage {
		return fmt.Errorf("invalid anti-dilution type")
	}
	if p.TriggerThreshold.IsNegative() || p.TriggerThreshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("trigger threshold must be between 0 and 1")
	}
	if p.AdjustmentCap.IsNegative() || p.AdjustmentCap.GT(math.LegacyOneDec()) {
		return fmt.Errorf("adjustment cap must be between 0 and 1")
	}
	if p.MinimumPrice.IsNegative() {
		return fmt.Errorf("minimum price cannot be negative")
	}
	return nil
}

// IssuanceRecord tracks historical share issuances for anti-dilution calculations
type IssuanceRecord struct {
	ID              uint64         `json:"id"`
	CompanyID       uint64         `json:"company_id"`
	ClassID         string         `json:"class_id"`

	// Issuance details
	SharesIssued    math.Int       `json:"shares_issued"`
	IssuePrice      math.LegacyDec `json:"issue_price"`       // Price per share at issuance
	TotalRaised     math.LegacyDec `json:"total_raised"`      // Total funds raised
	Recipient       string         `json:"recipient"`         // Who received the shares

	// Classification
	IsDownRound     bool           `json:"is_down_round"`     // Was this a down round?
	PreviousPrice   math.LegacyDec `json:"previous_price"`    // Price of previous issuance
	PriceDropPercent math.LegacyDec `json:"price_drop_percent"` // % decrease from previous

	// Metadata
	Purpose         string         `json:"purpose,omitempty"` // Reason for issuance
	IssuedAt        time.Time      `json:"issued_at"`
}

// Validate validates an IssuanceRecord
func (r IssuanceRecord) Validate() error {
	if r.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if r.ClassID == "" {
		return fmt.Errorf("class ID cannot be empty")
	}
	if r.SharesIssued.IsNil() || r.SharesIssued.LTE(math.ZeroInt()) {
		return fmt.Errorf("shares issued must be positive")
	}
	if r.IssuePrice.IsNegative() {
		return fmt.Errorf("issue price cannot be negative")
	}
	if r.Recipient == "" {
		return fmt.Errorf("recipient cannot be empty")
	}
	return nil
}

// AntiDilutionAdjustment records an adjustment made to protect shareholders
type AntiDilutionAdjustment struct {
	ID                uint64           `json:"id"`
	CompanyID         uint64           `json:"company_id"`
	ClassID           string           `json:"class_id"`           // Protected share class
	IssuanceRecordID  uint64           `json:"issuance_record_id"` // Which issuance triggered this

	// Adjustment details
	AdjustmentType    AntiDilutionType `json:"adjustment_type"`
	Shareholder       string           `json:"shareholder"`        // Who received adjustment
	OriginalShares    math.Int         `json:"original_shares"`    // Shares before adjustment
	AdjustedShares    math.Int         `json:"adjusted_shares"`    // Shares after adjustment
	SharesAdded       math.Int         `json:"shares_added"`       // Additional shares granted

	// Calculation details
	OldPrice          math.LegacyDec   `json:"old_price"`          // Price before down round
	NewPrice          math.LegacyDec   `json:"new_price"`          // Price at down round
	AdjustmentRatio   math.LegacyDec   `json:"adjustment_ratio"`   // Multiplier applied

	// Metadata
	AppliedAt         time.Time        `json:"applied_at"`
}

// Validate validates an AntiDilutionAdjustment
func (a AntiDilutionAdjustment) Validate() error {
	if a.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if a.ClassID == "" {
		return fmt.Errorf("class ID cannot be empty")
	}
	if a.Shareholder == "" {
		return fmt.Errorf("shareholder cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(a.Shareholder); err != nil {
		return fmt.Errorf("invalid shareholder address: %v", err)
	}
	if a.OriginalShares.IsNil() || a.OriginalShares.LT(math.ZeroInt()) {
		return fmt.Errorf("original shares cannot be negative")
	}
	if a.AdjustedShares.IsNil() || a.AdjustedShares.LT(a.OriginalShares) {
		return fmt.Errorf("adjusted shares must be >= original shares")
	}
	return nil
}

// NewAntiDilutionProvision creates a new anti-dilution provision
func NewAntiDilutionProvision(
	companyID uint64,
	classID string,
	provisionType AntiDilutionType,
	triggerThreshold, adjustmentCap, minimumPrice math.LegacyDec,
) AntiDilutionProvision {
	now := time.Now()
	return AntiDilutionProvision{
		CompanyID:        companyID,
		ClassID:          classID,
		ProvisionType:    provisionType,
		TriggerThreshold: triggerThreshold,
		AdjustmentCap:    adjustmentCap,
		MinimumPrice:     minimumPrice,
		IsActive:         true,
		TotalAdjustments: 0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// NewIssuanceRecord creates a new issuance record
func NewIssuanceRecord(
	id, companyID uint64,
	classID string,
	sharesIssued math.Int,
	issuePrice math.LegacyDec,
	recipient string,
	previousPrice math.LegacyDec,
	purpose string,
) IssuanceRecord {
	totalRaised := issuePrice.MulInt(sharesIssued)
	isDownRound := issuePrice.LT(previousPrice)

	var priceDropPercent math.LegacyDec
	if previousPrice.IsPositive() {
		priceDropPercent = previousPrice.Sub(issuePrice).Quo(previousPrice)
		if priceDropPercent.IsNegative() {
			priceDropPercent = math.LegacyZeroDec()
		}
	} else {
		priceDropPercent = math.LegacyZeroDec()
	}

	return IssuanceRecord{
		ID:               id,
		CompanyID:        companyID,
		ClassID:          classID,
		SharesIssued:     sharesIssued,
		IssuePrice:       issuePrice,
		TotalRaised:      totalRaised,
		Recipient:        recipient,
		IsDownRound:      isDownRound,
		PreviousPrice:    previousPrice,
		PriceDropPercent: priceDropPercent,
		Purpose:          purpose,
		IssuedAt:         time.Now(),
	}
}

// NewAntiDilutionAdjustment creates a new adjustment record
func NewAntiDilutionAdjustment(
	id, companyID uint64,
	classID string,
	issuanceRecordID uint64,
	adjustmentType AntiDilutionType,
	shareholder string,
	originalShares, adjustedShares math.Int,
	oldPrice, newPrice, adjustmentRatio math.LegacyDec,
) AntiDilutionAdjustment {
	return AntiDilutionAdjustment{
		ID:               id,
		CompanyID:        companyID,
		ClassID:          classID,
		IssuanceRecordID: issuanceRecordID,
		AdjustmentType:   adjustmentType,
		Shareholder:      shareholder,
		OriginalShares:   originalShares,
		AdjustedShares:   adjustedShares,
		SharesAdded:      adjustedShares.Sub(originalShares),
		OldPrice:         oldPrice,
		NewPrice:         newPrice,
		AdjustmentRatio:  adjustmentRatio,
		AppliedAt:        time.Now(),
	}
}

// CalculateFullRatchetAdjustment calculates shares after full ratchet adjustment
// Formula: New Shares = Original Shares × (Old Price / New Price)
func CalculateFullRatchetAdjustment(
	originalShares math.Int,
	oldPrice, newPrice math.LegacyDec,
) (adjustedShares math.Int, adjustmentRatio math.LegacyDec) {
	if newPrice.IsZero() || newPrice.GTE(oldPrice) {
		return originalShares, math.LegacyOneDec()
	}

	adjustmentRatio = oldPrice.Quo(newPrice)
	adjustedSharesDec := adjustmentRatio.MulInt(originalShares)
	adjustedShares = adjustedSharesDec.TruncateInt()

	return adjustedShares, adjustmentRatio
}

// CalculateWeightedAverageAdjustment calculates shares after weighted average adjustment
// Formula: New Price = ((Old Shares × Old Price) + (New Shares × New Price)) / (Old Shares + New Shares)
// Then: Adjusted Shares = Original Shares × (Old Price / Adjusted Price)
func CalculateWeightedAverageAdjustment(
	originalShares math.Int,
	oldPrice, newPrice math.LegacyDec,
	totalSharesBefore math.Int, // Total shares outstanding before new issuance
	newSharesIssued math.Int,   // Shares issued in the down round
) (adjustedShares math.Int, adjustmentRatio math.LegacyDec) {
	if newPrice.IsZero() || newPrice.GTE(oldPrice) {
		return originalShares, math.LegacyOneDec()
	}

	// Calculate weighted average price
	oldValue := oldPrice.MulInt(totalSharesBefore)
	newValue := newPrice.MulInt(newSharesIssued)
	totalShares := totalSharesBefore.Add(newSharesIssued)

	if totalShares.IsZero() {
		return originalShares, math.LegacyOneDec()
	}

	weightedAvgPrice := oldValue.Add(newValue).QuoInt(totalShares)

	// Calculate adjustment ratio
	if weightedAvgPrice.IsZero() {
		return originalShares, math.LegacyOneDec()
	}

	adjustmentRatio = oldPrice.Quo(weightedAvgPrice)
	adjustedSharesDec := adjustmentRatio.MulInt(originalShares)
	adjustedShares = adjustedSharesDec.TruncateInt()

	return adjustedShares, adjustmentRatio
}

// AddressHolding represents all shareholdings for a specific address
// Used for inheritance and portfolio queries
type AddressHolding struct {
	CompanyID uint64   `json:"company_id"`
	ClassID   string   `json:"class_id"`
	Shares    math.Int `json:"shares"`
}

// =============================================================================
// Beneficial Ownership Types
// =============================================================================

// BeneficialOwnership tracks the true owner of shares held in module accounts
// This is critical for dividend distribution when shares are locked in escrow, lending, or DEX
// Also used by governance to calculate voting power for LP providers
type BeneficialOwnership struct {
	ModuleAccount   string    `json:"module_account"`   // Module holding the shares (escrow, lending, dex)
	CompanyID       uint64    `json:"company_id"`       // Company ID
	ClassID         string    `json:"class_id"`         // Share class ID
	BeneficialOwner string    `json:"beneficial_owner"` // The REAL owner (who gets dividends and voting rights)
	Shares          math.Int  `json:"shares"`           // Number of shares locked
	LockedAt        time.Time `json:"locked_at"`        // When the shares were locked
	ReferenceID     uint64    `json:"reference_id"`     // ID in the locking module (escrow_id, loan_id, order_id)
	ReferenceType   string    `json:"reference_type"`   // Type of lock: "escrow", "loan", "order", "liquidity_provision"
}

// Validate validates a BeneficialOwnership record
func (bo BeneficialOwnership) Validate() error {
	if bo.ModuleAccount == "" {
		return fmt.Errorf("module account cannot be empty")
	}
	if bo.CompanyID == 0 {
		return fmt.Errorf("company ID cannot be zero")
	}
	if bo.ClassID == "" {
		return fmt.Errorf("class ID cannot be empty")
	}
	if bo.BeneficialOwner == "" {
		return fmt.Errorf("beneficial owner cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(bo.BeneficialOwner); err != nil {
		return fmt.Errorf("invalid beneficial owner address: %v", err)
	}
	if bo.Shares.IsNil() || bo.Shares.LTE(math.ZeroInt()) {
		return fmt.Errorf("shares must be positive")
	}
	if bo.ReferenceType == "" {
		return fmt.Errorf("reference type cannot be empty")
	}
	// Reference type should be one of the known types
	validTypes := map[string]bool{
		"escrow":              true,
		"loan":                true,
		"order":               true,
		"treasury_investment": true,
		"liquidity_provision": true, // LP pools
	}
	if !validTypes[bo.ReferenceType] {
		return fmt.Errorf("invalid reference type: %s", bo.ReferenceType)
	}
	return nil
}