package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Default parameter values
var (
	DefaultSwapFeeRate              = math.LegacyNewDecWithPrec(1, 3) // 0.1%
	DefaultMinReserveRatio          = math.LegacyNewDecWithPrec(20, 2) // 20%
	DefaultMaxWithdrawalPercent     = math.LegacyNewDecWithPrec(10, 2) // 10%
	DefaultWithdrawalVotingPeriod   = uint64(7 * 24 * 60 * 60)        // 7 days in seconds
	DefaultRequiredValidatorApproval = math.LegacyNewDecWithPrec(667, 3) // 66.7%
	DefaultBridgeEnabled            = true
)

// Params defines the parameters for the bridge module
type Params struct {
	// SwapFeeRate is the fee charged on swaps (0.1% default)
	SwapFeeRate math.LegacyDec `json:"swap_fee_rate" yaml:"swap_fee_rate"`

	// MinReserveRatio is the minimum reserve ratio that must be maintained (20% default)
	MinReserveRatio math.LegacyDec `json:"min_reserve_ratio" yaml:"min_reserve_ratio"`

	// MaxWithdrawalPercent is the maximum percentage of reserve that can be withdrawn per proposal (10% default)
	MaxWithdrawalPercent math.LegacyDec `json:"max_withdrawal_percent" yaml:"max_withdrawal_percent"`

	// WithdrawalVotingPeriod is the voting period for withdrawal proposals in seconds (7 days default)
	WithdrawalVotingPeriod uint64 `json:"withdrawal_voting_period" yaml:"withdrawal_voting_period"`

	// RequiredValidatorApproval is the percentage of validators required to approve withdrawal (66.7% default)
	RequiredValidatorApproval math.LegacyDec `json:"required_validator_approval" yaml:"required_validator_approval"`

	// BridgeEnabled controls whether bridge operations are enabled
	BridgeEnabled bool `json:"bridge_enabled" yaml:"bridge_enabled"`
}

// NewParams creates a new Params instance
func NewParams(
	swapFeeRate,
	minReserveRatio,
	maxWithdrawalPercent,
	requiredValidatorApproval math.LegacyDec,
	withdrawalVotingPeriod uint64,
	bridgeEnabled bool,
) Params {
	return Params{
		SwapFeeRate:               swapFeeRate,
		MinReserveRatio:           minReserveRatio,
		MaxWithdrawalPercent:      maxWithdrawalPercent,
		WithdrawalVotingPeriod:    withdrawalVotingPeriod,
		RequiredValidatorApproval: requiredValidatorApproval,
		BridgeEnabled:             bridgeEnabled,
	}
}

// DefaultParams returns default module parameters
func DefaultParams() Params {
	return NewParams(
		DefaultSwapFeeRate,
		DefaultMinReserveRatio,
		DefaultMaxWithdrawalPercent,
		DefaultRequiredValidatorApproval,
		DefaultWithdrawalVotingPeriod,
		DefaultBridgeEnabled,
	)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateSwapFeeRate(p.SwapFeeRate); err != nil {
		return err
	}

	if err := validateMinReserveRatio(p.MinReserveRatio); err != nil {
		return err
	}

	if err := validateMaxWithdrawalPercent(p.MaxWithdrawalPercent); err != nil {
		return err
	}

	if err := validateRequiredValidatorApproval(p.RequiredValidatorApproval); err != nil {
		return err
	}

	if p.WithdrawalVotingPeriod == 0 {
		return fmt.Errorf("withdrawal voting period must be positive")
	}

	return nil
}

func validateSwapFeeRate(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("swap fee rate cannot be negative: %s", v)
	}

	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("swap fee rate cannot exceed 100%%: %s", v)
	}

	return nil
}

func validateMinReserveRatio(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("minimum reserve ratio cannot be negative: %s", v)
	}

	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("minimum reserve ratio cannot exceed 100%%: %s", v)
	}

	return nil
}

func validateMaxWithdrawalPercent(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max withdrawal percent cannot be negative: %s", v)
	}

	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("max withdrawal percent cannot exceed 100%%: %s", v)
	}

	return nil
}

func validateRequiredValidatorApproval(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("required validator approval cannot be negative: %s", v)
	}

	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("required validator approval cannot exceed 100%%: %s", v)
	}

	// Minimum 50% approval required
	if v.LT(math.LegacyNewDecWithPrec(50, 2)) {
		return fmt.Errorf("required validator approval must be at least 50%%: %s", v)
	}

	return nil
}
