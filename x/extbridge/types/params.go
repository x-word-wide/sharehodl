package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	"gopkg.in/yaml.v2"
)

// Params defines the parameters for the extbridge module
type Params struct {
	// BridgingEnabled controls whether bridging is active
	BridgingEnabled bool `json:"bridging_enabled" yaml:"bridging_enabled"`

	// AttestationThreshold is the fraction of validators needed to approve a deposit (e.g., 0.67 for 2/3)
	AttestationThreshold math.LegacyDec `json:"attestation_threshold" yaml:"attestation_threshold"`

	// MinValidatorTier is the minimum staking tier required to submit attestations (e.g., TierArchon)
	MinValidatorTier int32 `json:"min_validator_tier" yaml:"min_validator_tier"`

	// WithdrawalTimelock is the time a withdrawal must wait before being signed (in seconds)
	WithdrawalTimelock uint64 `json:"withdrawal_timelock" yaml:"withdrawal_timelock"`

	// RateLimitWindow is the time window for rate limiting (in seconds)
	RateLimitWindow uint64 `json:"rate_limit_window" yaml:"rate_limit_window"`

	// MaxWithdrawalPerWindow is the maximum amount that can be withdrawn per asset per window
	MaxWithdrawalPerWindow math.Int `json:"max_withdrawal_per_window" yaml:"max_withdrawal_per_window"`

	// BridgeFee is the fee charged for bridging (as a decimal, e.g., 0.001 for 0.1%)
	BridgeFee math.LegacyDec `json:"bridge_fee" yaml:"bridge_fee"`

	// TSSThreshold is the threshold for TSS signing (e.g., 2/3)
	TSSThreshold math.LegacyDec `json:"tss_threshold" yaml:"tss_threshold"`

	// EmergencyPauseEnabled allows circuit breaker to halt operations
	EmergencyPauseEnabled bool `json:"emergency_pause_enabled" yaml:"emergency_pause_enabled"`
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		BridgingEnabled:        true,
		AttestationThreshold:   math.LegacyNewDecWithPrec(67, 2), // 67% = 2/3
		MinValidatorTier:       4,                                 // TierArchon (10M+ HODL)
		WithdrawalTimelock:     3600,                              // 1 hour
		RateLimitWindow:        86400,                             // 24 hours
		MaxWithdrawalPerWindow: math.NewInt(1_000_000_000_000),    // 1M HODL per day
		BridgeFee:              math.LegacyNewDecWithPrec(1, 3),   // 0.1%
		TSSThreshold:           math.LegacyNewDecWithPrec(67, 2),  // 67% = 2/3
		EmergencyPauseEnabled:  false,
	}
}

// String implements the Stringer interface
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.AttestationThreshold.IsNil() || p.AttestationThreshold.LTE(math.LegacyNewDecWithPrec(50, 2)) || p.AttestationThreshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("attestation threshold must be between 0.5 and 1.0: %s", p.AttestationThreshold)
	}

	if p.MinValidatorTier < 0 || p.MinValidatorTier > 5 {
		return fmt.Errorf("min validator tier must be between 0 and 5: %d", p.MinValidatorTier)
	}

	if p.WithdrawalTimelock == 0 {
		return fmt.Errorf("withdrawal timelock must be positive")
	}

	if p.RateLimitWindow == 0 {
		return fmt.Errorf("rate limit window must be positive")
	}

	if p.MaxWithdrawalPerWindow.IsNil() || !p.MaxWithdrawalPerWindow.IsPositive() {
		return fmt.Errorf("max withdrawal per window must be positive")
	}

	if p.BridgeFee.IsNil() || p.BridgeFee.IsNegative() {
		return fmt.Errorf("bridge fee must be non-negative")
	}

	if p.BridgeFee.GTE(math.LegacyOneDec()) {
		return fmt.Errorf("bridge fee must be less than 100%%")
	}

	if p.TSSThreshold.IsNil() || p.TSSThreshold.LTE(math.LegacyNewDecWithPrec(50, 2)) || p.TSSThreshold.GT(math.LegacyOneDec()) {
		return fmt.Errorf("TSS threshold must be between 0.5 and 1.0: %s", p.TSSThreshold)
	}

	return nil
}

// DefaultWithdrawalTimelockDuration returns the default timelock as a duration
func (p Params) WithdrawalTimelockDuration() time.Duration {
	return time.Duration(p.WithdrawalTimelock) * time.Second
}

// RateLimitWindowDuration returns the rate limit window as a duration
func (p Params) RateLimitWindowDuration() time.Duration {
	return time.Duration(p.RateLimitWindow) * time.Second
}
