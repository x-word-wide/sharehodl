package types

import (
	"fmt"
	"time"
)

// TSSSessionStatus represents the status of a TSS signing session
type TSSSessionStatus int

const (
	TSSSessionStatusPending   TSSSessionStatus = 0 // Created, waiting for participants
	TSSSessionStatusActive    TSSSessionStatus = 1 // Active signing in progress
	TSSSessionStatusCompleted TSSSessionStatus = 2 // Signing completed successfully
	TSSSessionStatusFailed    TSSSessionStatus = 3 // Signing failed
	TSSSessionStatusTimeout   TSSSessionStatus = 4 // Session timed out
)

// String returns the string representation of TSSSessionStatus
func (tss TSSSessionStatus) String() string {
	switch tss {
	case TSSSessionStatusPending:
		return "pending"
	case TSSSessionStatusActive:
		return "active"
	case TSSSessionStatusCompleted:
		return "completed"
	case TSSSessionStatusFailed:
		return "failed"
	case TSSSessionStatusTimeout:
		return "timeout"
	default:
		return "unknown"
	}
}

// TSSSession represents a threshold signature signing session
type TSSSession struct {
	ID               uint64           `json:"id" yaml:"id"`
	WithdrawalID     uint64           `json:"withdrawal_id" yaml:"withdrawal_id"`     // Associated withdrawal
	ChainID          string           `json:"chain_id" yaml:"chain_id"`               // External chain
	Status           TSSSessionStatus `json:"status" yaml:"status"`
	Participants     []string         `json:"participants" yaml:"participants"`       // Validator addresses participating
	RequiredSigs     uint64           `json:"required_sigs" yaml:"required_sigs"`     // Number of signatures required
	ReceivedSigs     uint64           `json:"received_sigs" yaml:"received_sigs"`     // Number of signatures received
	SignatureShares  []TSSSignatureShare `json:"signature_shares" yaml:"signature_shares"` // Collected signature shares
	CombinedSignature []byte          `json:"combined_signature,omitempty" yaml:"combined_signature,omitempty"` // Final combined signature
	CreatedAt        time.Time        `json:"created_at" yaml:"created_at"`
	CompletedAt      *time.Time       `json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
	TimeoutAt        time.Time        `json:"timeout_at" yaml:"timeout_at"`           // When session expires
	Message          []byte           `json:"message" yaml:"message"`                 // Message being signed (tx data)
}

// NewTSSSession creates a new TSS session
func NewTSSSession(
	id uint64,
	withdrawalID uint64,
	chainID string,
	participants []string,
	requiredSigs uint64,
	timeout time.Duration,
	message []byte,
	createdAt time.Time,
) TSSSession {
	return TSSSession{
		ID:           id,
		WithdrawalID: withdrawalID,
		ChainID:      chainID,
		Status:       TSSSessionStatusPending,
		Participants: participants,
		RequiredSigs: requiredSigs,
		ReceivedSigs: 0,
		SignatureShares: []TSSSignatureShare{},
		CreatedAt:    createdAt,
		TimeoutAt:    createdAt.Add(timeout),
		Message:      message,
	}
}

// Validate validates the TSS session
func (ts TSSSession) Validate() error {
	if ts.ChainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if len(ts.Participants) == 0 {
		return fmt.Errorf("participants cannot be empty")
	}
	if ts.RequiredSigs == 0 {
		return fmt.Errorf("required sigs must be positive")
	}
	if ts.RequiredSigs > uint64(len(ts.Participants)) {
		return fmt.Errorf("required sigs cannot exceed participants")
	}
	if len(ts.Message) == 0 {
		return fmt.Errorf("message cannot be empty")
	}
	return nil
}

// IsCompleted returns whether the session is completed
func (ts TSSSession) IsCompleted() bool {
	return ts.Status == TSSSessionStatusCompleted
}

// IsFailed returns whether the session has failed
func (ts TSSSession) IsFailed() bool {
	return ts.Status == TSSSessionStatusFailed
}

// IsTimeout returns whether the session has timed out
func (ts TSSSession) IsTimeout(currentTime time.Time) bool {
	return currentTime.After(ts.TimeoutAt)
}

// HasEnoughSignatures returns whether enough signatures have been collected
func (ts TSSSession) HasEnoughSignatures() bool {
	return ts.ReceivedSigs >= ts.RequiredSigs
}

// CanAddSignature returns whether more signatures can be added
func (ts TSSSession) CanAddSignature() bool {
	return ts.Status == TSSSessionStatusPending || ts.Status == TSSSessionStatusActive
}

// TSSSignatureShare represents a validator's signature share
type TSSSignatureShare struct {
	SessionID     uint64    `json:"session_id" yaml:"session_id"`
	Validator     string    `json:"validator" yaml:"validator"`         // Validator address
	SignatureData []byte    `json:"signature_data" yaml:"signature_data"` // Signature share
	SubmittedAt   time.Time `json:"submitted_at" yaml:"submitted_at"`
}

// NewTSSSignatureShare creates a new signature share
func NewTSSSignatureShare(
	sessionID uint64,
	validator string,
	signatureData []byte,
) TSSSignatureShare {
	return TSSSignatureShare{
		SessionID:     sessionID,
		Validator:     validator,
		SignatureData: signatureData,
		SubmittedAt:   time.Now(),
	}
}

// Validate validates the signature share
func (tss TSSSignatureShare) Validate() error {
	if tss.Validator == "" {
		return fmt.Errorf("validator cannot be empty")
	}
	if len(tss.SignatureData) == 0 {
		return fmt.Errorf("signature data cannot be empty")
	}
	return nil
}

// TSSKeyGeneration represents a TSS key generation session
// This is used to generate/rotate the shared TSS key
type TSSKeyGeneration struct {
	ID           uint64    `json:"id" yaml:"id"`
	ChainID      string    `json:"chain_id" yaml:"chain_id"`
	Status       TSSSessionStatus `json:"status" yaml:"status"`
	Participants []string  `json:"participants" yaml:"participants"`
	Threshold    uint64    `json:"threshold" yaml:"threshold"`
	PublicKey    []byte    `json:"public_key,omitempty" yaml:"public_key,omitempty"`
	CreatedAt    time.Time `json:"created_at" yaml:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" yaml:"completed_at,omitempty"`
}
