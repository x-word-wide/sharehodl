package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DepositStatus represents the status of a deposit
type DepositStatus int

const (
	DepositStatusPending   DepositStatus = 0 // Observed but not yet attested
	DepositStatusAttesting DepositStatus = 1 // Collecting attestations
	DepositStatusCompleted DepositStatus = 2 // Attested and HODL minted
	DepositStatusRejected  DepositStatus = 3 // Rejected (invalid or fraud)
)

// String returns the string representation of DepositStatus
func (ds DepositStatus) String() string {
	switch ds {
	case DepositStatusPending:
		return "pending"
	case DepositStatusAttesting:
		return "attesting"
	case DepositStatusCompleted:
		return "completed"
	case DepositStatusRejected:
		return "rejected"
	default:
		return "unknown"
	}
}

// Deposit represents a deposit from an external chain
type Deposit struct {
	ID              uint64        `json:"id" yaml:"id"`
	ChainID         string        `json:"chain_id" yaml:"chain_id"`             // External chain ID
	AssetSymbol     string        `json:"asset_symbol" yaml:"asset_symbol"`     // Asset symbol (USDT, BTC, etc.)
	ExternalTxHash  string        `json:"external_tx_hash" yaml:"external_tx_hash"` // Transaction hash on external chain
	ExternalBlockHeight uint64    `json:"external_block_height" yaml:"external_block_height"` // Block height on external chain
	Sender          string        `json:"sender" yaml:"sender"`                 // Address on external chain
	Recipient       string        `json:"recipient" yaml:"recipient"`           // ShareHODL address to receive HODL
	Amount          math.Int      `json:"amount" yaml:"amount"`                 // Amount deposited (in external asset units)
	HODLAmount      math.Int      `json:"hodl_amount" yaml:"hodl_amount"`       // Amount of HODL to mint
	Status          DepositStatus `json:"status" yaml:"status"`
	Attestations    uint64        `json:"attestations" yaml:"attestations"`     // Number of validator attestations
	RequiredAttestations uint64   `json:"required_attestations" yaml:"required_attestations"` // Required attestations
	ObservedAt      time.Time     `json:"observed_at" yaml:"observed_at"`       // When first observed
	CompletedAt     *time.Time    `json:"completed_at,omitempty" yaml:"completed_at,omitempty"` // When completed
	Confirmations   uint64        `json:"confirmations" yaml:"confirmations"`   // Current confirmations on external chain
}

// NewDeposit creates a new Deposit
func NewDeposit(
	id uint64,
	chainID string,
	assetSymbol string,
	externalTxHash string,
	externalBlockHeight uint64,
	sender string,
	recipient string,
	amount math.Int,
	hodlAmount math.Int,
	requiredAttestations uint64,
	observedAt time.Time,
) Deposit {
	return Deposit{
		ID:                   id,
		ChainID:              chainID,
		AssetSymbol:          assetSymbol,
		ExternalTxHash:       externalTxHash,
		ExternalBlockHeight:  externalBlockHeight,
		Sender:               sender,
		Recipient:            recipient,
		Amount:               amount,
		HODLAmount:           hodlAmount,
		Status:               DepositStatusPending,
		Attestations:         0,
		RequiredAttestations: requiredAttestations,
		ObservedAt:           observedAt,
		Confirmations:        0,
	}
}

// Validate validates the deposit
func (d Deposit) Validate() error {
	if d.ChainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if d.AssetSymbol == "" {
		return fmt.Errorf("asset symbol cannot be empty")
	}
	if d.ExternalTxHash == "" {
		return fmt.Errorf("external tx hash cannot be empty")
	}
	if d.Sender == "" {
		return fmt.Errorf("sender cannot be empty")
	}
	if d.Recipient == "" {
		return fmt.Errorf("recipient cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(d.Recipient); err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}
	if d.Amount.IsNil() || !d.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive")
	}
	if d.HODLAmount.IsNil() || !d.HODLAmount.IsPositive() {
		return fmt.Errorf("HODL amount must be positive")
	}
	if d.RequiredAttestations == 0 {
		return fmt.Errorf("required attestations must be positive")
	}
	return nil
}

// IsCompleted returns whether the deposit is completed
func (d Deposit) IsCompleted() bool {
	return d.Status == DepositStatusCompleted
}

// IsRejected returns whether the deposit is rejected
func (d Deposit) IsRejected() bool {
	return d.Status == DepositStatusRejected
}

// CanAttest returns whether attestations are still being collected
func (d Deposit) CanAttest() bool {
	return d.Status == DepositStatusPending || d.Status == DepositStatusAttesting
}

// HasEnoughAttestations returns whether enough attestations have been collected
func (d Deposit) HasEnoughAttestations() bool {
	return d.Attestations >= d.RequiredAttestations
}

// DepositAttestation represents a validator's attestation of a deposit
type DepositAttestation struct {
	DepositID       uint64    `json:"deposit_id" yaml:"deposit_id"`
	Validator       string    `json:"validator" yaml:"validator"`     // Validator address
	ChainID         string    `json:"chain_id" yaml:"chain_id"`       // External chain ID
	AssetSymbol     string    `json:"asset_symbol" yaml:"asset_symbol"`
	ExternalTxHash  string    `json:"external_tx_hash" yaml:"external_tx_hash"`
	Amount          math.Int  `json:"amount" yaml:"amount"`           // Attested amount
	Approved        bool      `json:"approved" yaml:"approved"`       // true = approve, false = reject
	AttestedAt      time.Time `json:"attested_at" yaml:"attested_at"`
	Signature       []byte    `json:"signature,omitempty" yaml:"signature,omitempty"` // Optional signature
}

// NewDepositAttestation creates a new DepositAttestation
func NewDepositAttestation(
	depositID uint64,
	validator string,
	chainID string,
	assetSymbol string,
	externalTxHash string,
	amount math.Int,
	approved bool,
	attestedAt time.Time,
) DepositAttestation {
	return DepositAttestation{
		DepositID:      depositID,
		Validator:      validator,
		ChainID:        chainID,
		AssetSymbol:    assetSymbol,
		ExternalTxHash: externalTxHash,
		Amount:         amount,
		Approved:       approved,
		AttestedAt:     attestedAt,
	}
}

// Validate validates the attestation
func (da DepositAttestation) Validate() error {
	if da.Validator == "" {
		return fmt.Errorf("validator cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(da.Validator); err != nil {
		return fmt.Errorf("invalid validator address: %w", err)
	}
	if da.ChainID == "" {
		return fmt.Errorf("chain ID cannot be empty")
	}
	if da.AssetSymbol == "" {
		return fmt.Errorf("asset symbol cannot be empty")
	}
	if da.ExternalTxHash == "" {
		return fmt.Errorf("external tx hash cannot be empty")
	}
	if da.Amount.IsNil() || !da.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}
