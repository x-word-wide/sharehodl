package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
)

// RateLimit tracks withdrawal rate limits per asset
type RateLimit struct {
	ChainID       string    `json:"chain_id" yaml:"chain_id"`
	AssetSymbol   string    `json:"asset_symbol" yaml:"asset_symbol"`
	WindowStart   time.Time `json:"window_start" yaml:"window_start"`   // Start of current window
	WindowEnd     time.Time `json:"window_end" yaml:"window_end"`       // End of current window
	AmountUsed    math.Int  `json:"amount_used" yaml:"amount_used"`     // Amount withdrawn in window
	MaxAmount     math.Int  `json:"max_amount" yaml:"max_amount"`       // Max allowed per window
	TxCount       uint64    `json:"tx_count" yaml:"tx_count"`           // Number of transactions
}

// NewRateLimit creates a new RateLimit
func NewRateLimit(
	chainID string,
	assetSymbol string,
	windowStart time.Time,
	windowDuration time.Duration,
	maxAmount math.Int,
) RateLimit {
	return RateLimit{
		ChainID:     chainID,
		AssetSymbol: assetSymbol,
		WindowStart: windowStart,
		WindowEnd:   windowStart.Add(windowDuration),
		AmountUsed:  math.ZeroInt(),
		MaxAmount:   maxAmount,
		TxCount:     0,
	}
}

// Validate validates the rate limit
func (rl RateLimit) Validate() error {
	if rl.ChainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if rl.AssetSymbol == "" {
		return fmt.Errorf("asset symbol cannot be empty")
	}
	if rl.WindowEnd.Before(rl.WindowStart) {
		return fmt.Errorf("window end must be after window start")
	}
	if rl.AmountUsed.IsNil() || rl.AmountUsed.IsNegative() {
		return fmt.Errorf("amount used must be non-negative")
	}
	if rl.MaxAmount.IsNil() || !rl.MaxAmount.IsPositive() {
		return fmt.Errorf("max amount must be positive")
	}
	return nil
}

// IsExpired returns whether the current window has expired
func (rl RateLimit) IsExpired(currentTime time.Time) bool {
	return currentTime.After(rl.WindowEnd) || currentTime.Equal(rl.WindowEnd)
}

// HasCapacity returns whether there's capacity for the given amount
func (rl RateLimit) HasCapacity(amount math.Int) bool {
	newTotal := rl.AmountUsed.Add(amount)
	return newTotal.LTE(rl.MaxAmount)
}

// RemainingCapacity returns the remaining capacity in the window
func (rl RateLimit) RemainingCapacity() math.Int {
	remaining := rl.MaxAmount.Sub(rl.AmountUsed)
	if remaining.IsNegative() {
		return math.ZeroInt()
	}
	return remaining
}

// CircuitBreaker represents the emergency pause mechanism
type CircuitBreaker struct {
	Enabled        bool      `json:"enabled" yaml:"enabled"`
	Reason         string    `json:"reason" yaml:"reason"`
	TriggeredBy    string    `json:"triggered_by" yaml:"triggered_by"`    // Governance address or validator
	TriggeredAt    time.Time `json:"triggered_at" yaml:"triggered_at"`
	CanDeposit     bool      `json:"can_deposit" yaml:"can_deposit"`      // Whether deposits are allowed
	CanWithdraw    bool      `json:"can_withdraw" yaml:"can_withdraw"`    // Whether withdrawals are allowed
	CanAttest      bool      `json:"can_attest" yaml:"can_attest"`        // Whether attestations are allowed
	ExpiresAt      *time.Time `json:"expires_at,omitempty" yaml:"expires_at,omitempty"` // Optional auto-expiry
}

// NewCircuitBreaker creates a new CircuitBreaker
func NewCircuitBreaker(
	enabled bool,
	reason string,
	triggeredBy string,
	canDeposit bool,
	canWithdraw bool,
	canAttest bool,
	duration *time.Duration,
) CircuitBreaker {
	now := time.Now()
	cb := CircuitBreaker{
		Enabled:     enabled,
		Reason:      reason,
		TriggeredBy: triggeredBy,
		TriggeredAt: now,
		CanDeposit:  canDeposit,
		CanWithdraw: canWithdraw,
		CanAttest:   canAttest,
	}

	if duration != nil {
		expiry := now.Add(*duration)
		cb.ExpiresAt = &expiry
	}

	return cb
}

// Validate validates the circuit breaker
func (cb CircuitBreaker) Validate() error {
	if cb.Enabled && cb.Reason == "" {
		return fmt.Errorf("reason must be provided when circuit breaker is enabled")
	}
	if cb.Enabled && cb.TriggeredBy == "" {
		return fmt.Errorf("triggered by must be provided when circuit breaker is enabled")
	}
	return nil
}

// IsExpired returns whether the circuit breaker has expired
func (cb CircuitBreaker) IsExpired(currentTime time.Time) bool {
	if !cb.Enabled || cb.ExpiresAt == nil {
		return false
	}
	return currentTime.After(*cb.ExpiresAt) || currentTime.Equal(*cb.ExpiresAt)
}

// IsOperationAllowed checks if an operation is allowed
func (cb CircuitBreaker) IsOperationAllowed(operation string) bool {
	if !cb.Enabled {
		return true
	}

	switch operation {
	case "deposit":
		return cb.CanDeposit
	case "withdraw":
		return cb.CanWithdraw
	case "attest":
		return cb.CanAttest
	default:
		return false
	}
}

// AnomalyDetection tracks suspicious activity
type AnomalyDetection struct {
	ChainID        string    `json:"chain_id" yaml:"chain_id"`
	AssetSymbol    string    `json:"asset_symbol" yaml:"asset_symbol"`
	DetectionType  string    `json:"detection_type" yaml:"detection_type"`  // e.g., "rapid_withdrawal", "large_deposit"
	Severity       string    `json:"severity" yaml:"severity"`              // "low", "medium", "high", "critical"
	Description    string    `json:"description" yaml:"description"`
	DetectedAt     time.Time `json:"detected_at" yaml:"detected_at"`
	RelatedTxs     []string  `json:"related_txs" yaml:"related_txs"`        // Related transaction IDs
	ActionTaken    string    `json:"action_taken" yaml:"action_taken"`      // Action taken (if any)
	Resolved       bool      `json:"resolved" yaml:"resolved"`
}

// NewAnomalyDetection creates a new anomaly detection record
func NewAnomalyDetection(
	chainID string,
	assetSymbol string,
	detectionType string,
	severity string,
	description string,
	relatedTxs []string,
) AnomalyDetection {
	return AnomalyDetection{
		ChainID:       chainID,
		AssetSymbol:   assetSymbol,
		DetectionType: detectionType,
		Severity:      severity,
		Description:   description,
		DetectedAt:    time.Now(),
		RelatedTxs:    relatedTxs,
		Resolved:      false,
	}
}

// Validate validates the anomaly detection record
func (ad AnomalyDetection) Validate() error {
	if ad.ChainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if ad.DetectionType == "" {
		return fmt.Errorf("detection type cannot be empty")
	}
	if ad.Severity == "" {
		return fmt.Errorf("severity cannot be empty")
	}
	validSeverities := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
	if !validSeverities[ad.Severity] {
		return fmt.Errorf("invalid severity: %s", ad.Severity)
	}
	return nil
}

// IsCritical returns whether this is a critical anomaly
func (ad AnomalyDetection) IsCritical() bool {
	return ad.Severity == "critical"
}
