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
	KeyMinVerificationDelaySeconds = []byte("MinVerificationDelaySeconds")
	KeyMaxVerificationDelaySeconds = []byte("MaxVerificationDelaySeconds")
	KeyVerificationTimeoutDays     = []byte("VerificationTimeoutDays")
	KeyMinValidatorsRequired       = []byte("MinValidatorsRequired")
	KeyReputationDecayRate         = []byte("ReputationDecayRate")
	KeyBaseVerificationReward      = []byte("BaseVerificationReward")
	KeyEquityStakePercentage       = []byte("EquityStakePercentage")
)

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MinVerificationDelaySeconds: 86400,      // 1 day minimum (24 * 60 * 60)
		MaxVerificationDelaySeconds: 2592000,    // 30 days maximum (30 * 24 * 60 * 60)
		VerificationTimeoutDays:     14,         // 14 days timeout
		MinValidatorsRequired:       3,          // Minimum 3 validators
		ReputationDecayRate:         math.LegacyNewDecWithPrec(1, 2), // 1% monthly decay
		BaseVerificationReward:      math.NewInt(1000),     // 1000 HODL base reward
		EquityStakePercentage:       math.LegacyNewDecWithPrec(5, 2), // 5% equity stake
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
		paramtypes.NewParamSetPair(KeyMinVerificationDelaySeconds, &p.MinVerificationDelaySeconds, validateInt64),
		paramtypes.NewParamSetPair(KeyMaxVerificationDelaySeconds, &p.MaxVerificationDelaySeconds, validateInt64),
		paramtypes.NewParamSetPair(KeyVerificationTimeoutDays, &p.VerificationTimeoutDays, validateUint64),
		paramtypes.NewParamSetPair(KeyMinValidatorsRequired, &p.MinValidatorsRequired, validateUint64),
		paramtypes.NewParamSetPair(KeyReputationDecayRate, &p.ReputationDecayRate, validateDec),
		paramtypes.NewParamSetPair(KeyBaseVerificationReward, &p.BaseVerificationReward, validateInt),
		paramtypes.NewParamSetPair(KeyEquityStakePercentage, &p.EquityStakePercentage, validateDec),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.MinVerificationDelaySeconds <= 0 {
		return fmt.Errorf("min verification delay must be positive")
	}
	if p.MaxVerificationDelaySeconds <= 0 {
		return fmt.Errorf("max verification delay must be positive")
	}
	if p.MinVerificationDelaySeconds >= p.MaxVerificationDelaySeconds {
		return fmt.Errorf("min verification delay must be less than max verification delay")
	}
	if p.VerificationTimeoutDays == 0 {
		return fmt.Errorf("verification timeout days must be positive")
	}
	if p.MinValidatorsRequired == 0 {
		return fmt.Errorf("minimum validators required must be positive")
	}
	if p.ReputationDecayRate.IsNegative() || p.ReputationDecayRate.GT(math.LegacyOneDec()) {
		return fmt.Errorf("reputation decay rate must be between 0 and 1")
	}
	if p.BaseVerificationReward.IsNegative() {
		return fmt.Errorf("base verification reward must be non-negative")
	}
	if p.EquityStakePercentage.IsNegative() || p.EquityStakePercentage.GT(math.LegacyOneDec()) {
		return fmt.Errorf("equity stake percentage must be between 0 and 1")
	}
	return nil
}

// Validation functions
func validateInt64(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("parameter must be positive")
	}
	return nil
}

func validateUint64(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("parameter must be positive")
	}
	return nil
}

func validateDec(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNegative() {
		return fmt.Errorf("parameter must be non-negative: %s", v)
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