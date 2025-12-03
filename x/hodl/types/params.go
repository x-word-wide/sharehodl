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
	KeyMintingEnabled     = []byte("MintingEnabled")
	KeyCollateralRatio    = []byte("CollateralRatio") 
	KeyStabilityFee       = []byte("StabilityFee")
	KeyLiquidationRatio   = []byte("LiquidationRatio")
	KeyMintFee           = []byte("MintFee")
	KeyBurnFee           = []byte("BurnFee")
)

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MintingEnabled:     true,
		CollateralRatio:    math.LegacyNewDecWithPrec(150, 2), // 150% collateral requirement
		StabilityFee:       math.LegacyNewDecWithPrec(5, 4),   // 0.05% annual fee
		LiquidationRatio:   math.LegacyNewDecWithPrec(130, 2), // 130% liquidation threshold
		MintFee:           math.LegacyNewDecWithPrec(1, 4),    // 0.01% mint fee
		BurnFee:           math.LegacyNewDecWithPrec(1, 4),    // 0.01% burn fee
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
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if !p.CollateralRatio.IsPositive() {
		return fmt.Errorf("collateral ratio must be positive: %s", p.CollateralRatio)
	}
	if !p.LiquidationRatio.IsPositive() {
		return fmt.Errorf("liquidation ratio must be positive: %s", p.LiquidationRatio)
	}
	if p.CollateralRatio.LTE(p.LiquidationRatio) {
		return fmt.Errorf("collateral ratio must be greater than liquidation ratio")
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