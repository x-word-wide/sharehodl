package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WithdrawalStatus represents the status of a withdrawal
type WithdrawalStatus int

const (
	WithdrawalStatusPending    WithdrawalStatus = 0 // Request submitted, waiting for timelock
	WithdrawalStatusTimelocked WithdrawalStatus = 1 // In timelock period
	WithdrawalStatusReady      WithdrawalStatus = 2 // Ready for TSS signing
	WithdrawalStatusSigning    WithdrawalStatus = 3 // TSS signing in progress
	WithdrawalStatusSigned     WithdrawalStatus = 4 // Signed, ready to broadcast
	WithdrawalStatusBroadcast  WithdrawalStatus = 5 // Broadcasted to external chain
	WithdrawalStatusCompleted  WithdrawalStatus = 6 // Confirmed on external chain
	WithdrawalStatusCancelled  WithdrawalStatus = 7 // Cancelled
	WithdrawalStatusFailed     WithdrawalStatus = 8 // Failed
	WithdrawalStatusTimeout    WithdrawalStatus = 9 // TSS signing timed out
	WithdrawalStatusRefunded   WithdrawalStatus = 10 // Escrowed HODL refunded
)

// String returns the string representation of WithdrawalStatus
func (ws WithdrawalStatus) String() string {
	switch ws {
	case WithdrawalStatusPending:
		return "pending"
	case WithdrawalStatusTimelocked:
		return "timelocked"
	case WithdrawalStatusReady:
		return "ready"
	case WithdrawalStatusSigning:
		return "signing"
	case WithdrawalStatusSigned:
		return "signed"
	case WithdrawalStatusBroadcast:
		return "broadcast"
	case WithdrawalStatusCompleted:
		return "completed"
	case WithdrawalStatusCancelled:
		return "cancelled"
	case WithdrawalStatusFailed:
		return "failed"
	case WithdrawalStatusTimeout:
		return "timeout"
	case WithdrawalStatusRefunded:
		return "refunded"
	default:
		return "unknown"
	}
}

// Withdrawal represents a withdrawal to an external chain
type Withdrawal struct {
	ID              uint64           `json:"id" yaml:"id"`
	ChainID         string           `json:"chain_id" yaml:"chain_id"`             // Destination chain ID
	AssetSymbol     string           `json:"asset_symbol" yaml:"asset_symbol"`     // Asset to withdraw
	Sender          string           `json:"sender" yaml:"sender"`                 // ShareHODL address burning HODL
	Recipient       string           `json:"recipient" yaml:"recipient"`           // Address on external chain
	HODLAmount      math.Int         `json:"hodl_amount" yaml:"hodl_amount"`       // HODL burned
	ExternalAmount  math.Int         `json:"external_amount" yaml:"external_amount"` // Amount to send on external chain
	FeeAmount       math.Int         `json:"fee_amount" yaml:"fee_amount"`         // Bridge fee
	Status          WithdrawalStatus `json:"status" yaml:"status"`
	RequestedAt     time.Time        `json:"requested_at" yaml:"requested_at"`
	TimelockExpiry  time.Time        `json:"timelock_expiry" yaml:"timelock_expiry"` // When timelock expires
	TSSSessionID    *uint64          `json:"tss_session_id,omitempty" yaml:"tss_session_id,omitempty"` // Associated TSS session
	ExternalTxHash  string           `json:"external_tx_hash,omitempty" yaml:"external_tx_hash,omitempty"` // Tx hash on external chain
	CompletedAt     *time.Time       `json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
	FailureReason   string           `json:"failure_reason,omitempty" yaml:"failure_reason,omitempty"`
}

// NewWithdrawal creates a new Withdrawal
func NewWithdrawal(
	id uint64,
	chainID string,
	assetSymbol string,
	sender string,
	recipient string,
	hodlAmount math.Int,
	externalAmount math.Int,
	feeAmount math.Int,
	timelockDuration time.Duration,
) Withdrawal {
	now := time.Now()
	return Withdrawal{
		ID:             id,
		ChainID:        chainID,
		AssetSymbol:    assetSymbol,
		Sender:         sender,
		Recipient:      recipient,
		HODLAmount:     hodlAmount,
		ExternalAmount: externalAmount,
		FeeAmount:      feeAmount,
		Status:         WithdrawalStatusPending,
		RequestedAt:    now,
		TimelockExpiry: now.Add(timelockDuration),
	}
}

// Validate validates the withdrawal
func (w Withdrawal) Validate() error {
	if w.ChainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if w.AssetSymbol == "" {
		return fmt.Errorf("asset symbol cannot be empty")
	}
	if w.Sender == "" {
		return fmt.Errorf("sender cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(w.Sender); err != nil {
		return fmt.Errorf("invalid sender address: %w", err)
	}
	if w.Recipient == "" {
		return fmt.Errorf("recipient cannot be empty")
	}
	if w.HODLAmount.IsNil() || !w.HODLAmount.IsPositive() {
		return fmt.Errorf("HODL amount must be positive")
	}
	if w.ExternalAmount.IsNil() || !w.ExternalAmount.IsPositive() {
		return fmt.Errorf("external amount must be positive")
	}
	if w.FeeAmount.IsNil() || w.FeeAmount.IsNegative() {
		return fmt.Errorf("fee amount must be non-negative")
	}
	return nil
}

// IsTimelockExpired returns whether the timelock has expired
func (w Withdrawal) IsTimelockExpired(currentTime time.Time) bool {
	return currentTime.After(w.TimelockExpiry) || currentTime.Equal(w.TimelockExpiry)
}

// CanSign returns whether the withdrawal is ready for TSS signing
func (w Withdrawal) CanSign(currentTime time.Time) bool {
	return w.Status == WithdrawalStatusReady && w.IsTimelockExpired(currentTime)
}

// IsCompleted returns whether the withdrawal is completed
func (w Withdrawal) IsCompleted() bool {
	return w.Status == WithdrawalStatusCompleted
}

// IsCancelled returns whether the withdrawal is cancelled
func (w Withdrawal) IsCancelled() bool {
	return w.Status == WithdrawalStatusCancelled
}

// IsFailed returns whether the withdrawal has failed
func (w Withdrawal) IsFailed() bool {
	return w.Status == WithdrawalStatusFailed
}

// IsFinal returns whether the withdrawal is in a final state
func (w Withdrawal) IsFinal() bool {
	return w.IsCompleted() || w.IsCancelled() || w.IsFailed()
}
