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