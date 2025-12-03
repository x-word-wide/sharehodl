package types

import (
	"fmt"
	"time"
	"gopkg.in/yaml.v2"

	"cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// Parameter store keys
var (
	KeyMinVerificationDelay    = []byte("MinVerificationDelay")
	KeyMaxVerificationDelay    = []byte("MaxVerificationDelay")
	KeyVerificationTimeoutDays = []byte("VerificationTimeoutDays")
	KeyMinValidatorsRequired   = []byte("MinValidatorsRequired")
	KeyReputationDecayRate     = []byte("ReputationDecayRate")
	KeyBaseVerificationReward  = []byte("BaseVerificationReward")
	KeyEquityStakePercentage   = []byte("EquityStakePercentage")
)

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MinVerificationDelay:    24 * time.Hour,        // 1 day minimum
		MaxVerificationDelay:    30 * 24 * time.Hour,   // 30 days maximum
		VerificationTimeoutDays: 14,                    // 14 days timeout
		MinValidatorsRequired:   3,                     // Minimum 3 validators
		ReputationDecayRate:     math.LegacyNewDecWithPrec(1, 2), // 1% monthly decay
		BaseVerificationReward:  math.NewInt(1000),     // 1000 HODL base reward
		EquityStakePercentage:   math.LegacyNewDecWithPrec(5, 2), // 5% equity stake
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
		paramtypes.NewParamSetPair(KeyMinVerificationDelay, &p.MinVerificationDelay, validateDuration),
		paramtypes.NewParamSetPair(KeyMaxVerificationDelay, &p.MaxVerificationDelay, validateDuration),
		paramtypes.NewParamSetPair(KeyVerificationTimeoutDays, &p.VerificationTimeoutDays, validateUint64),
		paramtypes.NewParamSetPair(KeyMinValidatorsRequired, &p.MinValidatorsRequired, validateUint64),
		paramtypes.NewParamSetPair(KeyReputationDecayRate, &p.ReputationDecayRate, validateDec),
		paramtypes.NewParamSetPair(KeyBaseVerificationReward, &p.BaseVerificationReward, validateInt),
		paramtypes.NewParamSetPair(KeyEquityStakePercentage, &p.EquityStakePercentage, validateDec),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.MinVerificationDelay >= p.MaxVerificationDelay {
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
func validateDuration(i interface{}) error {
	_, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
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