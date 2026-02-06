package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// LoanStatus represents the status of a loan
type LoanStatus int32

const (
	LoanStatusPending     LoanStatus = iota // Loan request submitted
	LoanStatusActive                        // Loan approved and active
	LoanStatusRepaid                        // Loan fully repaid
	LoanStatusDefaulted                     // Loan defaulted
	LoanStatusLiquidated                    // Collateral liquidated
	LoanStatusCancelled                     // Loan cancelled before funding
)

func (s LoanStatus) String() string {
	switch s {
	case LoanStatusPending:
		return "pending"
	case LoanStatusActive:
		return "active"
	case LoanStatusRepaid:
		return "repaid"
	case LoanStatusDefaulted:
		return "defaulted"
	case LoanStatusLiquidated:
		return "liquidated"
	case LoanStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// CollateralType represents the type of collateral
type CollateralType int32

const (
	CollateralTypeHODL   CollateralType = iota // HODL tokens
	CollateralTypeEquity                       // Equity shares
)

func (c CollateralType) String() string {
	switch c {
	case CollateralTypeHODL:
		return "hodl"
	case CollateralTypeEquity:
		return "equity"
	default:
		return "unknown"
	}
}

// InterestRateType represents how interest is calculated
type InterestRateType int32

const (
	InterestRateFixed    InterestRateType = iota // Fixed interest rate
	InterestRateVariable                         // Variable based on utilization
)

// Loan represents a lending position
type Loan struct {
	ID             uint64         `json:"id"`
	Borrower       string         `json:"borrower"`        // Address of borrower
	Lender         string         `json:"lender"`          // Address of lender (or pool)

	// Loan terms
	Principal      math.Int       `json:"principal"`       // Amount borrowed (in HODL)
	InterestRate   math.LegacyDec `json:"interest_rate"`   // Annual interest rate (e.g., 0.08 = 8%)
	RateType       InterestRateType `json:"rate_type"`
	Duration       time.Duration  `json:"duration"`        // Loan duration

	// Collateral
	Collateral     []Collateral   `json:"collateral"`
	CollateralValue math.LegacyDec `json:"collateral_value"` // Total collateral value in HODL
	CollateralRatio math.LegacyDec `json:"collateral_ratio"` // Current collateral ratio

	// Repayment tracking
	AccruedInterest math.LegacyDec `json:"accrued_interest"` // Interest accrued so far
	PrincipalPaid   math.Int       `json:"principal_paid"`   // Principal already repaid
	InterestPaid    math.LegacyDec `json:"interest_paid"`    // Interest already paid
	TotalOwed       math.LegacyDec `json:"total_owed"`       // Total amount owed

	// Status
	Status         LoanStatus     `json:"status"`

	// Timestamps
	CreatedAt      time.Time      `json:"created_at"`
	ActivatedAt    time.Time      `json:"activated_at"`
	MaturityAt     time.Time      `json:"maturity_at"`       // When loan is due
	LastAccrualAt  time.Time      `json:"last_accrual_at"`   // Last interest accrual
	RepaidAt       time.Time      `json:"repaid_at"`

	// Liquidation thresholds
	LiquidationRatio     math.LegacyDec `json:"liquidation_ratio"`     // e.g., 1.25 = 125%
	LiquidationPenalty   math.LegacyDec `json:"liquidation_penalty"`   // e.g., 0.10 = 10%
	MinCollateralRatio   math.LegacyDec `json:"min_collateral_ratio"`  // e.g., 1.50 = 150%
}

// Collateral represents collateral for a loan
type Collateral struct {
	Type        CollateralType `json:"type"`
	Denom       string         `json:"denom"`        // Token denomination
	Amount      math.Int       `json:"amount"`       // Amount locked
	Value       math.LegacyDec `json:"value"`        // Value in HODL at time of locking
	CompanyID   uint64         `json:"company_id"`   // For equity collateral
	ShareClass  string         `json:"share_class"`  // For equity collateral
}

// LendingPool represents a pool of funds for lending
type LendingPool struct {
	ID             uint64         `json:"id"`
	Name           string         `json:"name"`

	// Pool assets
	TotalDeposits  math.Int       `json:"total_deposits"`    // Total HODL deposited
	TotalBorrowed  math.Int       `json:"total_borrowed"`    // Total HODL borrowed
	AvailableLiquidity math.Int   `json:"available_liquidity"` // Available for borrowing

	// Pool tokens
	PoolTokenSupply math.Int      `json:"pool_token_supply"` // LP token supply

	// Interest rates
	BaseRate       math.LegacyDec `json:"base_rate"`         // Base interest rate
	MultiplierRate math.LegacyDec `json:"multiplier_rate"`   // Rate multiplier based on utilization
	JumpRate       math.LegacyDec `json:"jump_rate"`         // Jump rate above kink
	Kink           math.LegacyDec `json:"kink"`              // Utilization kink point

	// Pool statistics
	UtilizationRate math.LegacyDec `json:"utilization_rate"` // Current utilization
	SupplyAPY       math.LegacyDec `json:"supply_apy"`       // Annual yield for suppliers
	BorrowAPY       math.LegacyDec `json:"borrow_apy"`       // Annual rate for borrowers

	// Reserves
	ReserveFactor   math.LegacyDec `json:"reserve_factor"`   // % of interest to reserves
	TotalReserves   math.Int       `json:"total_reserves"`   // Accumulated reserves

	// Timestamps
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// PoolDeposit represents a user's deposit in a lending pool
type PoolDeposit struct {
	User           string         `json:"user"`
	PoolID         uint64         `json:"pool_id"`
	DepositAmount  math.Int       `json:"deposit_amount"`    // Original deposit
	PoolTokens     math.Int       `json:"pool_tokens"`       // LP tokens received
	DepositedAt    time.Time      `json:"deposited_at"`
	LastClaimAt    time.Time      `json:"last_claim_at"`
}

// LoanOffer represents a peer-to-peer loan offer
type LoanOffer struct {
	ID                      uint64           `json:"id"`
	Lender                  string           `json:"lender"`
	Amount                  math.Int         `json:"amount"`                    // Amount willing to lend
	MinInterestRate         math.LegacyDec   `json:"min_interest_rate"`         // Minimum acceptable interest
	MaxDuration             time.Duration    `json:"max_duration"`
	RequiredCollateralRatio math.LegacyDec   `json:"required_collateral_ratio"`
	AcceptedCollateralTypes []CollateralType `json:"accepted_collateral_types"`
	Active                  bool             `json:"active"`
	CreatedAt               time.Time        `json:"created_at"`
	ExpiresAt               time.Time        `json:"expires_at"`

	// Stake-as-Trust-Ceiling: Stake backing this offer
	StakeLocked bool     `json:"stake_locked"` // True if stake is locked for this offer
	StakeAmount math.Int `json:"stake_amount"` // Amount of stake backing this offer
}

// LoanRequest represents a borrower's loan request
type LoanRequest struct {
	ID              uint64         `json:"id"`
	Borrower        string         `json:"borrower"`
	Amount          math.Int       `json:"amount"`             // Amount requested
	MaxInterestRate math.LegacyDec `json:"max_interest_rate"`  // Maximum acceptable interest
	Duration        time.Duration  `json:"duration"`
	Collateral      []Collateral   `json:"collateral"` // Offered collateral
	Active          bool           `json:"active"`
	CreatedAt       time.Time      `json:"created_at"`
	ExpiresAt       time.Time      `json:"expires_at"`

	// Stake-as-Trust-Ceiling: Stake backing this request
	StakeLocked bool     `json:"stake_locked"` // True if stake is locked for this request
	StakeAmount math.Int `json:"stake_amount"` // Amount of stake backing this request
}

// Validation methods

func (l Loan) Validate() error {
	if l.Borrower == "" {
		return fmt.Errorf("borrower cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(l.Borrower); err != nil {
		return fmt.Errorf("invalid borrower address: %v", err)
	}
	if l.Principal.IsNil() || l.Principal.LTE(math.ZeroInt()) {
		return fmt.Errorf("principal must be positive")
	}
	if l.InterestRate.IsNil() || l.InterestRate.IsNegative() {
		return fmt.Errorf("interest rate cannot be negative")
	}
	if len(l.Collateral) == 0 {
		return fmt.Errorf("loan must have collateral")
	}
	for _, c := range l.Collateral {
		if err := c.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c Collateral) Validate() error {
	if c.Denom == "" {
		return fmt.Errorf("collateral denom cannot be empty")
	}
	if c.Amount.IsNil() || c.Amount.LTE(math.ZeroInt()) {
		return fmt.Errorf("collateral amount must be positive")
	}
	return nil
}

func (p LendingPool) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("pool name cannot be empty")
	}
	return nil
}

// Constructor functions

// NewLoan creates a new loan
func NewLoan(
	id uint64,
	borrower string,
	principal math.Int,
	interestRate math.LegacyDec,
	duration time.Duration,
	collateral []Collateral,
) Loan {
	now := time.Now()

	// Calculate total collateral value
	collateralValue := math.LegacyZeroDec()
	for _, c := range collateral {
		collateralValue = collateralValue.Add(c.Value)
	}

	// Calculate initial collateral ratio
	collateralRatio := collateralValue.QuoInt(principal)

	return Loan{
		ID:               id,
		Borrower:         borrower,
		Principal:        principal,
		InterestRate:     interestRate,
		RateType:         InterestRateFixed,
		Duration:         duration,
		Collateral:       collateral,
		CollateralValue:  collateralValue,
		CollateralRatio:  collateralRatio,
		AccruedInterest:  math.LegacyZeroDec(),
		PrincipalPaid:    math.ZeroInt(),
		InterestPaid:     math.LegacyZeroDec(),
		TotalOwed:        math.LegacyNewDecFromInt(principal),
		Status:           LoanStatusPending,
		CreatedAt:        now,
		LastAccrualAt:    now,
		LiquidationRatio: math.LegacyNewDecWithPrec(125, 2),    // 125%
		LiquidationPenalty: math.LegacyNewDecWithPrec(10, 2),   // 10%
		MinCollateralRatio: math.LegacyNewDecWithPrec(150, 2),  // 150%
	}
}

// NewLendingPool creates a new lending pool
func NewLendingPool(id uint64, name string) LendingPool {
	now := time.Now()
	return LendingPool{
		ID:              id,
		Name:            name,
		TotalDeposits:   math.ZeroInt(),
		TotalBorrowed:   math.ZeroInt(),
		AvailableLiquidity: math.ZeroInt(),
		PoolTokenSupply: math.ZeroInt(),
		BaseRate:        math.LegacyNewDecWithPrec(2, 2),       // 2% base rate
		MultiplierRate:  math.LegacyNewDecWithPrec(20, 2),      // 20% multiplier
		JumpRate:        math.LegacyNewDecWithPrec(100, 2),     // 100% jump rate
		Kink:            math.LegacyNewDecWithPrec(80, 2),      // 80% kink
		UtilizationRate: math.LegacyZeroDec(),
		SupplyAPY:       math.LegacyZeroDec(),
		BorrowAPY:       math.LegacyZeroDec(),
		ReserveFactor:   math.LegacyNewDecWithPrec(10, 2),      // 10% reserves
		TotalReserves:   math.ZeroInt(),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// CalculateBorrowRate calculates the borrow rate based on utilization
func (p LendingPool) CalculateBorrowRate() math.LegacyDec {
	if p.TotalDeposits.IsZero() {
		return p.BaseRate
	}

	utilization := math.LegacyNewDecFromInt(p.TotalBorrowed).QuoInt(p.TotalDeposits)

	if utilization.LTE(p.Kink) {
		// Normal rate: base + (utilization * multiplier)
		return p.BaseRate.Add(utilization.Mul(p.MultiplierRate))
	} else {
		// Above kink: base + (kink * multiplier) + ((utilization - kink) * jumpRate)
		normalRate := p.BaseRate.Add(p.Kink.Mul(p.MultiplierRate))
		excessUtilization := utilization.Sub(p.Kink)
		return normalRate.Add(excessUtilization.Mul(p.JumpRate))
	}
}

// CalculateSupplyRate calculates the supply rate based on borrow rate and utilization
func (p LendingPool) CalculateSupplyRate() math.LegacyDec {
	borrowRate := p.CalculateBorrowRate()
	if p.TotalDeposits.IsZero() {
		return math.LegacyZeroDec()
	}

	utilization := math.LegacyNewDecFromInt(p.TotalBorrowed).QuoInt(p.TotalDeposits)
	// Supply rate = borrow rate * utilization * (1 - reserve factor)
	return borrowRate.Mul(utilization).Mul(math.LegacyOneDec().Sub(p.ReserveFactor))
}

// CalculateAccruedInterest calculates interest accrued since last accrual
func (l Loan) CalculateAccruedInterest(currentTime time.Time) math.LegacyDec {
	if l.LastAccrualAt.IsZero() || currentTime.Before(l.LastAccrualAt) {
		return math.LegacyZeroDec()
	}

	// Time in years
	duration := currentTime.Sub(l.LastAccrualAt)
	yearsElapsed := math.LegacyNewDec(int64(duration.Hours())).QuoInt64(8760) // 365 * 24

	// Interest = Principal * Rate * Time
	outstandingPrincipal := l.Principal.Sub(l.PrincipalPaid)
	interest := l.InterestRate.MulInt(outstandingPrincipal).Mul(yearsElapsed)

	return interest
}

// IsLiquidatable checks if a loan should be liquidated
func (l Loan) IsLiquidatable() bool {
	return l.Status == LoanStatusActive && l.CollateralRatio.LT(l.LiquidationRatio)
}

// IsOverdue checks if a loan is past maturity
func (l Loan) IsOverdue(currentTime time.Time) bool {
	return l.Status == LoanStatusActive && currentTime.After(l.MaturityAt)
}

// ============ Stake-as-Trust-Ceiling: P2P Stake Structures ============

// LenderStake represents a lender's security deposit for P2P lending
// Trust ceiling: lenders cannot post offers exceeding their stake
type LenderStake struct {
	Address         string    `json:"address"`
	TotalStake      math.Int  `json:"total_stake"`       // Total HODL staked
	AvailableStake  math.Int  `json:"available_stake"`   // Stake not backing active offers
	LockedStake     math.Int  `json:"locked_stake"`      // Stake backing active offers
	ActiveOffers    uint64    `json:"active_offers"`     // Number of active offers
	TotalOfferValue math.Int  `json:"total_offer_value"` // Total value of active offers
	StakedAt        time.Time `json:"staked_at"`

	// Unbonding state (14-day withdrawal period)
	UnbondingAmount math.Int  `json:"unbonding_amount"`
	UnbondingAt     time.Time `json:"unbonding_at,omitempty"`
}

// BorrowerStake represents a borrower's security deposit for P2P lending
// Trust ceiling: borrowers cannot request loans exceeding their stake
type BorrowerStake struct {
	Address          string    `json:"address"`
	TotalStake       math.Int  `json:"total_stake"`
	AvailableStake   math.Int  `json:"available_stake"`
	LockedStake      math.Int  `json:"locked_stake"`
	ActiveRequests   uint64    `json:"active_requests"`
	TotalRequestValue math.Int `json:"total_request_value"`
	StakedAt         time.Time `json:"staked_at"`

	// Unbonding state
	UnbondingAmount math.Int  `json:"unbonding_amount"`
	UnbondingAt     time.Time `json:"unbonding_at,omitempty"`
}

// NewLenderStake creates a new lender stake record
func NewLenderStake(address string, amount math.Int) LenderStake {
	return LenderStake{
		Address:         address,
		TotalStake:      amount,
		AvailableStake:  amount,
		LockedStake:     math.ZeroInt(),
		ActiveOffers:    0,
		TotalOfferValue: math.ZeroInt(),
		StakedAt:        time.Now(),
		UnbondingAmount: math.ZeroInt(),
	}
}

// NewBorrowerStake creates a new borrower stake record
func NewBorrowerStake(address string, amount math.Int) BorrowerStake {
	return BorrowerStake{
		Address:          address,
		TotalStake:       amount,
		AvailableStake:   amount,
		LockedStake:      math.ZeroInt(),
		ActiveRequests:   0,
		TotalRequestValue: math.ZeroInt(),
		StakedAt:         time.Now(),
		UnbondingAmount:  math.ZeroInt(),
	}
}

// GetTrustCeiling returns the maximum offer/request value the stake holder can post
func (s LenderStake) GetTrustCeiling() math.Int {
	return s.AvailableStake
}

// GetTrustCeiling returns the maximum request value the stake holder can post
func (s BorrowerStake) GetTrustCeiling() math.Int {
	return s.AvailableStake
}

// CanPostOffer checks if lender can post an offer of the given value
func (s LenderStake) CanPostOffer(value math.Int) bool {
	return s.AvailableStake.GTE(value)
}

// CanPostRequest checks if borrower can post a request of the given value
func (s BorrowerStake) CanPostRequest(value math.Int) bool {
	return s.AvailableStake.GTE(value)
}

// IsUnbondingComplete checks if the 14-day unbonding period has passed
func (s LenderStake) IsUnbondingComplete(now time.Time) bool {
	if s.UnbondingAmount.IsZero() {
		return false
	}
	unbondingEnd := s.UnbondingAt.Add(14 * 24 * time.Hour)
	return now.After(unbondingEnd) || now.Equal(unbondingEnd)
}

// IsUnbondingComplete checks if the 14-day unbonding period has passed
func (s BorrowerStake) IsUnbondingComplete(now time.Time) bool {
	if s.UnbondingAmount.IsZero() {
		return false
	}
	unbondingEnd := s.UnbondingAt.Add(14 * 24 * time.Hour)
	return now.After(unbondingEnd) || now.Equal(unbondingEnd)
}

// Validation methods for stakes

func (s LenderStake) Validate() error {
	if s.Address == "" {
		return fmt.Errorf("lender address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(s.Address); err != nil {
		return fmt.Errorf("invalid lender address: %v", err)
	}
	if s.TotalStake.IsNil() || s.TotalStake.IsNegative() {
		return fmt.Errorf("total stake cannot be negative")
	}
	return nil
}

func (s BorrowerStake) Validate() error {
	if s.Address == "" {
		return fmt.Errorf("borrower address cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(s.Address); err != nil {
		return fmt.Errorf("invalid borrower address: %v", err)
	}
	if s.TotalStake.IsNil() || s.TotalStake.IsNegative() {
		return fmt.Errorf("total stake cannot be negative")
	}
	return nil
}
