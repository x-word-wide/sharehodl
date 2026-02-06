package types

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
)

// Params defines the parameters for the inheritance module
// Duration fields use int64 seconds for proper JSON serialization
type Params struct {
	// Minimum inactivity period before switch can trigger (e.g., 180 days) - in seconds
	MinInactivityPeriodSeconds int64 `json:"min_inactivity_period_seconds" yaml:"min_inactivity_period_seconds"`

	// Minimum grace period after trigger (e.g., 30 days) - in seconds
	MinGracePeriodSeconds int64 `json:"min_grace_period_seconds" yaml:"min_grace_period_seconds"`

	// Minimum claim window for beneficiaries (e.g., 30 days) - in seconds
	MinClaimWindowSeconds int64 `json:"min_claim_window_seconds" yaml:"min_claim_window_seconds"`

	// Maximum claim window for beneficiaries (e.g., 365 days) - in seconds
	MaxClaimWindowSeconds int64 `json:"max_claim_window_seconds" yaml:"max_claim_window_seconds"`

	// Ultra-long inactivity period for charity transfer (e.g., 50 years) - in seconds
	UltraLongInactivityPeriodSeconds int64 `json:"ultra_long_inactivity_period_seconds" yaml:"ultra_long_inactivity_period_seconds"`

	// Default charity address if none specified
	DefaultCharityAddress string `json:"default_charity_address" yaml:"default_charity_address"`

	// Maximum number of beneficiaries per plan
	MaxBeneficiaries uint32 `json:"max_beneficiaries" yaml:"max_beneficiaries"`
}

// Helper methods to get durations
func (p Params) MinInactivityPeriod() time.Duration {
	return time.Duration(p.MinInactivityPeriodSeconds) * time.Second
}

func (p Params) MinGracePeriod() time.Duration {
	return time.Duration(p.MinGracePeriodSeconds) * time.Second
}

func (p Params) MinClaimWindow() time.Duration {
	return time.Duration(p.MinClaimWindowSeconds) * time.Second
}

func (p Params) MaxClaimWindow() time.Duration {
	return time.Duration(p.MaxClaimWindowSeconds) * time.Second
}

func (p Params) UltraLongInactivityPeriod() time.Duration {
	return time.Duration(p.UltraLongInactivityPeriodSeconds) * time.Second
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MinInactivityPeriodSeconds:       15552000,    // 180 days (180 * 24 * 60 * 60)
		MinGracePeriodSeconds:            2592000,     // 30 days (30 * 24 * 60 * 60)
		MinClaimWindowSeconds:            2592000,     // 30 days (30 * 24 * 60 * 60)
		MaxClaimWindowSeconds:            31536000,    // 365 days (365 * 24 * 60 * 60)
		UltraLongInactivityPeriodSeconds: 1576800000,  // 50 years (50 * 365 * 24 * 60 * 60)
		DefaultCharityAddress:            "",          // Will be set to community pool module address
		MaxBeneficiaries:                 10,          // Max 10 beneficiaries per plan
	}
}

// String implements the Stringer interface
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.MinInactivityPeriodSeconds <= 0 {
		return fmt.Errorf("min inactivity period must be positive")
	}
	if p.MinGracePeriodSeconds < 2592000 { // 30 days in seconds
		return fmt.Errorf("min grace period must be at least 30 days")
	}
	if p.MinClaimWindowSeconds <= 0 {
		return fmt.Errorf("min claim window must be positive")
	}
	if p.MaxClaimWindowSeconds < p.MinClaimWindowSeconds {
		return fmt.Errorf("max claim window must be >= min claim window")
	}
	if p.UltraLongInactivityPeriodSeconds <= p.MinInactivityPeriodSeconds {
		return fmt.Errorf("ultra long inactivity period must be greater than min inactivity period")
	}
	if p.MaxBeneficiaries == 0 || p.MaxBeneficiaries > 50 {
		return fmt.Errorf("max beneficiaries must be between 1 and 50")
	}
	return nil
}

// ProtoMessage implements proto.Message interface
func (m *Params) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *Params) Reset() { *m = Params{} }
