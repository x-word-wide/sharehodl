package types

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v2"
)

// Params defines the parameters for the inheritance module
type Params struct {
	// Minimum inactivity period before switch can trigger (e.g., 180 days)
	MinInactivityPeriod time.Duration `json:"min_inactivity_period" yaml:"min_inactivity_period"`

	// Minimum grace period after trigger (e.g., 30 days)
	MinGracePeriod time.Duration `json:"min_grace_period" yaml:"min_grace_period"`

	// Minimum claim window for beneficiaries (e.g., 30 days)
	MinClaimWindow time.Duration `json:"min_claim_window" yaml:"min_claim_window"`

	// Maximum claim window for beneficiaries (e.g., 365 days)
	MaxClaimWindow time.Duration `json:"max_claim_window" yaml:"max_claim_window"`

	// Ultra-long inactivity period for charity transfer (e.g., 50 years)
	UltraLongInactivityPeriod time.Duration `json:"ultra_long_inactivity_period" yaml:"ultra_long_inactivity_period"`

	// Default charity address if none specified
	DefaultCharityAddress string `json:"default_charity_address" yaml:"default_charity_address"`

	// Maximum number of beneficiaries per plan
	MaxBeneficiaries uint32 `json:"max_beneficiaries" yaml:"max_beneficiaries"`
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{
		MinInactivityPeriod:       180 * 24 * time.Hour, // 180 days
		MinGracePeriod:            30 * 24 * time.Hour,  // 30 days minimum
		MinClaimWindow:            30 * 24 * time.Hour,  // 30 days minimum
		MaxClaimWindow:            365 * 24 * time.Hour, // 1 year maximum
		UltraLongInactivityPeriod: 50 * 365 * 24 * time.Hour, // 50 years
		DefaultCharityAddress:     "", // Will be set to community pool module address
		MaxBeneficiaries:          10, // Max 10 beneficiaries per plan
	}
}

// String implements the Stringer interface
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.MinInactivityPeriod <= 0 {
		return fmt.Errorf("min inactivity period must be positive")
	}
	if p.MinGracePeriod < 30*24*time.Hour {
		return fmt.Errorf("min grace period must be at least 30 days")
	}
	if p.MinClaimWindow <= 0 {
		return fmt.Errorf("min claim window must be positive")
	}
	if p.MaxClaimWindow < p.MinClaimWindow {
		return fmt.Errorf("max claim window must be >= min claim window")
	}
	if p.UltraLongInactivityPeriod <= p.MinInactivityPeriod {
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
