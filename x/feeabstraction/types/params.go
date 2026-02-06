package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Default parameter values - ALL governance-controllable via validator voting
const (
	// DefaultFeeMarkupBasisPoints is the default protocol profit (500 = 5%)
	DefaultFeeMarkupBasisPoints uint64 = 500

	// DefaultMaxSlippageBasisPoints is the default max swap slippage allowed (500 = 5%)
	DefaultMaxSlippageBasisPoints uint64 = 500

	// DefaultEnabled is whether fee abstraction is enabled by default
	DefaultEnabled bool = true
)

var (
	// DefaultMinEquityValueUhodl is the minimum equity value for fee abstraction
	DefaultMinEquityValueUhodl = math.NewInt(1000) // 1000 uhodl minimum

	// DefaultMaxFeeAbstractionPerBlockUhodl is the DoS protection limit per block
	DefaultMaxFeeAbstractionPerBlockUhodl = math.NewInt(100000000000) // 100k HODL

	// DefaultTreasuryFundingMinimum is the minimum HODL treasury must maintain
	DefaultTreasuryFundingMinimum = math.NewInt(1000000000000) // 1M HODL
)

// Params defines the parameters for the fee abstraction module
// ALL parameters are governance-controllable via validator voting
type Params struct {
	// FeeMarkupBasisPoints: Protocol profit (100 = 1%, 500 = 5%)
	FeeMarkupBasisPoints uint64 `json:"fee_markup_basis_points" yaml:"fee_markup_basis_points"`

	// MaxSlippageBasisPoints: Max swap slippage allowed (100 = 1%)
	MaxSlippageBasisPoints uint64 `json:"max_slippage_basis_points" yaml:"max_slippage_basis_points"`

	// MinEquityValueUhodl: Minimum equity value for fee abstraction
	MinEquityValueUhodl math.Int `json:"min_equity_value_uhodl" yaml:"min_equity_value_uhodl"`

	// EnabledEquitySymbols: Whitelist (empty = all active equities)
	EnabledEquitySymbols []string `json:"enabled_equity_symbols" yaml:"enabled_equity_symbols"`

	// Enabled: Master toggle for fee abstraction
	Enabled bool `json:"enabled" yaml:"enabled"`

	// MaxFeeAbstractionPerBlockUhodl: DoS protection limit per block
	MaxFeeAbstractionPerBlockUhodl math.Int `json:"max_fee_abstraction_per_block_uhodl" yaml:"max_fee_abstraction_per_block_uhodl"`

	// TreasuryFundingMinimum: Min HODL treasury must maintain
	TreasuryFundingMinimum math.Int `json:"treasury_funding_minimum" yaml:"treasury_funding_minimum"`
}

// ProtoMessage implements proto.Message interface
func (m *Params) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *Params) Reset() { *m = Params{} }

// String implements proto.Message interface
func (m *Params) String() string { return "feeabstraction_params" }

// DefaultParams returns default fee abstraction parameters
// ALL values can be changed via governance proposals
func DefaultParams() Params {
	return Params{
		FeeMarkupBasisPoints:           DefaultFeeMarkupBasisPoints,
		MaxSlippageBasisPoints:         DefaultMaxSlippageBasisPoints,
		MinEquityValueUhodl:            DefaultMinEquityValueUhodl,
		EnabledEquitySymbols:           []string{}, // Empty = all equities allowed
		Enabled:                        DefaultEnabled,
		MaxFeeAbstractionPerBlockUhodl: DefaultMaxFeeAbstractionPerBlockUhodl,
		TreasuryFundingMinimum:         DefaultTreasuryFundingMinimum,
	}
}

// Validate validates the params
func (p Params) Validate() error {
	if p.FeeMarkupBasisPoints > 5000 {
		return fmt.Errorf("fee markup cannot exceed 50%% (5000 basis points)")
	}
	if p.MaxSlippageBasisPoints > 2000 {
		return fmt.Errorf("max slippage cannot exceed 20%% (2000 basis points)")
	}
	if !p.MinEquityValueUhodl.IsNil() && p.MinEquityValueUhodl.IsNegative() {
		return fmt.Errorf("min equity value cannot be negative")
	}
	if !p.MaxFeeAbstractionPerBlockUhodl.IsNil() && p.MaxFeeAbstractionPerBlockUhodl.IsNegative() {
		return fmt.Errorf("max fee abstraction per block cannot be negative")
	}
	if !p.TreasuryFundingMinimum.IsNil() && p.TreasuryFundingMinimum.IsNegative() {
		return fmt.Errorf("treasury funding minimum cannot be negative")
	}
	return nil
}

// GetFeeMarkupMultiplier returns the fee markup as a decimal multiplier
func (p Params) GetFeeMarkupMultiplier() math.LegacyDec {
	return math.LegacyNewDec(int64(p.FeeMarkupBasisPoints)).QuoInt64(10000)
}

// GetMaxSlippageMultiplier returns the max slippage as a decimal multiplier
func (p Params) GetMaxSlippageMultiplier() math.LegacyDec {
	return math.LegacyNewDec(int64(p.MaxSlippageBasisPoints)).QuoInt64(10000)
}
