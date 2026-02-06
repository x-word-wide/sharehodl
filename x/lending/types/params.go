package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
)

// Default parameter values - ALL governance-controllable via validator voting
const (
	DefaultLoanDefaultSlashThreshold uint64 = 1_000_000_000 // 1000 HODL
	DefaultDefaultLoanDurationDays   uint64 = 365           // 1 year
	DefaultUnbondingPeriodDays       uint64 = 14            // 14 days
)

// Params defines the parameters for the lending module
// ALL parameters are governance-controllable via validator voting
type Params struct {
	// Loan configuration
	LoanDefaultSlashThreshold math.Int `json:"loan_default_slash_threshold" yaml:"loan_default_slash_threshold"` // Slash stake if default > this
	DefaultLoanDurationDays   uint64   `json:"default_loan_duration_days" yaml:"default_loan_duration_days"`     // Default loan term

	// Unbonding periods
	LenderUnbondingPeriodDays   uint64 `json:"lender_unbonding_period_days" yaml:"lender_unbonding_period_days"`
	BorrowerUnbondingPeriodDays uint64 `json:"borrower_unbonding_period_days" yaml:"borrower_unbonding_period_days"`

	// Interest rate bounds
	MinInterestRateBasisPoints uint64 `json:"min_interest_rate_basis_points" yaml:"min_interest_rate_basis_points"` // Minimum interest rate
	MaxInterestRateBasisPoints uint64 `json:"max_interest_rate_basis_points" yaml:"max_interest_rate_basis_points"` // Maximum interest rate

	// Collateralization
	MinCollateralRatioBasisPoints uint64 `json:"min_collateral_ratio_basis_points" yaml:"min_collateral_ratio_basis_points"` // Minimum collateral ratio

	// Master toggle
	LendingEnabled bool `json:"lending_enabled" yaml:"lending_enabled"`
}

// ProtoMessage implements proto.Message interface
func (m *Params) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *Params) Reset() { *m = Params{} }

// String implements proto.Message interface
func (m *Params) String() string { return "lending_params" }

// DefaultParams returns default lending parameters
// ALL values can be changed via governance proposals
func DefaultParams() Params {
	return Params{
		LoanDefaultSlashThreshold:     math.NewInt(int64(DefaultLoanDefaultSlashThreshold)),
		DefaultLoanDurationDays:       DefaultDefaultLoanDurationDays,
		LenderUnbondingPeriodDays:     DefaultUnbondingPeriodDays,
		BorrowerUnbondingPeriodDays:   DefaultUnbondingPeriodDays,
		MinInterestRateBasisPoints:    100,   // 1% minimum
		MaxInterestRateBasisPoints:    5000,  // 50% maximum
		MinCollateralRatioBasisPoints: 11000, // 110% minimum collateral
		LendingEnabled:                true,
	}
}

// Validate validates the params
func (p Params) Validate() error {
	if p.DefaultLoanDurationDays == 0 {
		return fmt.Errorf("default loan duration must be positive")
	}
	if p.LenderUnbondingPeriodDays == 0 {
		return fmt.Errorf("lender unbonding period must be positive")
	}
	if p.BorrowerUnbondingPeriodDays == 0 {
		return fmt.Errorf("borrower unbonding period must be positive")
	}
	if p.MinInterestRateBasisPoints > p.MaxInterestRateBasisPoints {
		return fmt.Errorf("min interest rate cannot exceed max interest rate")
	}
	if p.MinCollateralRatioBasisPoints < 10000 {
		return fmt.Errorf("min collateral ratio must be at least 100%%")
	}
	return nil
}

// GetDefaultLoanDuration returns the default loan duration as time.Duration
func (p Params) GetDefaultLoanDuration() time.Duration {
	return time.Duration(p.DefaultLoanDurationDays) * 24 * time.Hour
}

// GetLenderUnbondingPeriod returns the lender unbonding period as time.Duration
func (p Params) GetLenderUnbondingPeriod() time.Duration {
	return time.Duration(p.LenderUnbondingPeriodDays) * 24 * time.Hour
}

// GetBorrowerUnbondingPeriod returns the borrower unbonding period as time.Duration
func (p Params) GetBorrowerUnbondingPeriod() time.Duration {
	return time.Duration(p.BorrowerUnbondingPeriodDays) * 24 * time.Hour
}

// GetMinInterestRate returns minimum interest rate as a decimal
func (p Params) GetMinInterestRate() math.LegacyDec {
	return math.LegacyNewDec(int64(p.MinInterestRateBasisPoints)).QuoInt64(10000)
}

// GetMaxInterestRate returns maximum interest rate as a decimal
func (p Params) GetMaxInterestRate() math.LegacyDec {
	return math.LegacyNewDec(int64(p.MaxInterestRateBasisPoints)).QuoInt64(10000)
}

// GetMinCollateralRatio returns minimum collateral ratio as a decimal
func (p Params) GetMinCollateralRatio() math.LegacyDec {
	return math.LegacyNewDec(int64(p.MinCollateralRatioBasisPoints)).QuoInt64(10000)
}

// Note: ParamsKey is already defined in keys.go
