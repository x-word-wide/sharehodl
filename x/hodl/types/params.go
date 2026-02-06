package types

import (
	"fmt"
	"gopkg.in/yaml.v2"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// Parameter store keys
var (
	KeyMintingEnabled      = []byte("MintingEnabled")
	KeyCollateralRatio     = []byte("CollateralRatio")
	KeyStabilityFee        = []byte("StabilityFee")
	KeyLiquidationRatio    = []byte("LiquidationRatio")
	KeyMintFee             = []byte("MintFee")
	KeyBurnFee             = []byte("BurnFee")
	KeyLiquidationPenalty  = []byte("LiquidationPenalty")
	KeyLiquidatorReward    = []byte("LiquidatorReward")
	KeyMaxPriceDeviation   = []byte("MaxPriceDeviation")
	KeyBlocksPerYear       = []byte("BlocksPerYear")
	KeyMaxBadDebtLimit     = []byte("MaxBadDebtLimit")
)

// Default parameter values - all governance-controllable
const (
	DefaultBlocksPerYear uint64 = 5_256_000 // ~6 seconds per block
)

var (
	DefaultMaxBadDebtLimit = math.NewInt(1_000_000_000_000) // 1 trillion uhodl circuit breaker
)

// DefaultParams returns default parameters
// ALL values can be changed via governance proposals
func DefaultParams() Params {
	return Params{
		MintingEnabled:     true,
		CollateralRatio:    math.LegacyNewDecWithPrec(150, 2), // 150% collateral requirement
		StabilityFee:       math.LegacyNewDecWithPrec(5, 4),   // 0.05% annual fee
		LiquidationRatio:   math.LegacyNewDecWithPrec(130, 2), // 130% liquidation threshold
		MintFee:            math.LegacyNewDecWithPrec(1, 4),   // 0.01% mint fee
		BurnFee:            math.LegacyNewDecWithPrec(1, 4),   // 0.01% burn fee
		LiquidationPenalty: math.LegacyNewDecWithPrec(10, 2),  // 10% penalty on liquidation
		LiquidatorReward:   math.LegacyNewDecWithPrec(5, 2),   // 5% of penalty to liquidator
		MaxPriceDeviation:  math.LegacyNewDecWithPrec(20, 2),  // 20% max price change per update
		BlocksPerYear:      DefaultBlocksPerYear,
		MaxBadDebtLimit:    DefaultMaxBadDebtLimit,
	}
}

// ParamKeyTable returns the parameter key table
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// String implements the Stringer interface
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs returns param set pairs
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintingEnabled, &p.MintingEnabled, validateBool),
		paramtypes.NewParamSetPair(KeyCollateralRatio, &p.CollateralRatio, validateDec),
		paramtypes.NewParamSetPair(KeyStabilityFee, &p.StabilityFee, validateDec),
		paramtypes.NewParamSetPair(KeyLiquidationRatio, &p.LiquidationRatio, validateDec),
		paramtypes.NewParamSetPair(KeyMintFee, &p.MintFee, validateDec),
		paramtypes.NewParamSetPair(KeyBurnFee, &p.BurnFee, validateDec),
		paramtypes.NewParamSetPair(KeyLiquidationPenalty, &p.LiquidationPenalty, validateDec),
		paramtypes.NewParamSetPair(KeyLiquidatorReward, &p.LiquidatorReward, validateDec),
		paramtypes.NewParamSetPair(KeyMaxPriceDeviation, &p.MaxPriceDeviation, validateDec),
		paramtypes.NewParamSetPair(KeyBlocksPerYear, &p.BlocksPerYear, validateUint64),
		paramtypes.NewParamSetPair(KeyMaxBadDebtLimit, &p.MaxBadDebtLimit, validateInt),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	// Check if params are initialized (non-zero values)
	if p.CollateralRatio.IsNil() || p.LiquidationRatio.IsNil() || p.StabilityFee.IsNil() || p.MintFee.IsNil() || p.BurnFee.IsNil() {
		return fmt.Errorf("params are not properly initialized")
	}

	if !p.CollateralRatio.IsPositive() {
		return fmt.Errorf("collateral ratio must be positive: %s", p.CollateralRatio)
	}
	if !p.LiquidationRatio.IsPositive() {
		return fmt.Errorf("liquidation ratio must be positive: %s", p.LiquidationRatio)
	}
	if p.CollateralRatio.LTE(p.LiquidationRatio) {
		return fmt.Errorf("collateral ratio must be greater than liquidation ratio")
	}

	// Validate new params
	if !p.LiquidationPenalty.IsNil() && p.LiquidationPenalty.IsNegative() {
		return fmt.Errorf("liquidation penalty must be non-negative: %s", p.LiquidationPenalty)
	}
	if !p.LiquidatorReward.IsNil() && p.LiquidatorReward.IsNegative() {
		return fmt.Errorf("liquidator reward must be non-negative: %s", p.LiquidatorReward)
	}
	if !p.LiquidatorReward.IsNil() && !p.LiquidationPenalty.IsNil() && p.LiquidatorReward.GT(p.LiquidationPenalty) {
		return fmt.Errorf("liquidator reward cannot exceed liquidation penalty")
	}
	if !p.MaxPriceDeviation.IsNil() && (p.MaxPriceDeviation.IsNegative() || p.MaxPriceDeviation.GT(math.LegacyOneDec())) {
		return fmt.Errorf("max price deviation must be between 0 and 100%%: %s", p.MaxPriceDeviation)
	}
	if p.BlocksPerYear == 0 {
		return fmt.Errorf("blocks per year must be positive")
	}
	if !p.MaxBadDebtLimit.IsNil() && !p.MaxBadDebtLimit.IsPositive() {
		return fmt.Errorf("max bad debt limit must be positive: %s", p.MaxBadDebtLimit)
	}

	return nil
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateDec(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("parameter must be positive: %s", v)
	}
	return nil
}

func validateUint64(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateInt(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("parameter must be non-negative: %s", v)
	}
	return nil
}