package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BridgeReserve tracks all reserve balances for the bridge
type BridgeReserve struct {
	Denom   string   `json:"denom" yaml:"denom"`
	Balance math.Int `json:"balance" yaml:"balance"`
}

// NewBridgeReserve creates a new BridgeReserve
func NewBridgeReserve(denom string, balance math.Int) BridgeReserve {
	return BridgeReserve{
		Denom:   denom,
		Balance: balance,
	}
}

// Validate validates the bridge reserve
func (br BridgeReserve) Validate() error {
	if br.Denom == "" {
		return fmt.Errorf("denom cannot be empty")
	}
	if br.Balance.IsNegative() {
		return fmt.Errorf("balance cannot be negative")
	}
	return nil
}

// SupportedAsset represents a whitelisted asset with its swap configuration
type SupportedAsset struct {
	Denom            string         `json:"denom" yaml:"denom"`
	SwapRateToHODL   math.LegacyDec `json:"swap_rate_to_hodl" yaml:"swap_rate_to_hodl"` // How many HODL per 1 unit of asset
	SwapRateFromHODL math.LegacyDec `json:"swap_rate_from_hodl" yaml:"swap_rate_from_hodl"` // How many asset units per 1 HODL
	MinSwapAmount    math.Int       `json:"min_swap_amount" yaml:"min_swap_amount"`
	MaxSwapAmount    math.Int       `json:"max_swap_amount" yaml:"max_swap_amount"`
	Enabled          bool           `json:"enabled" yaml:"enabled"`
	AddedAt          time.Time      `json:"added_at" yaml:"added_at"`
}

// NewSupportedAsset creates a new SupportedAsset
func NewSupportedAsset(
	denom string,
	swapRateToHODL,
	swapRateFromHODL math.LegacyDec,
	minSwapAmount,
	maxSwapAmount math.Int,
	enabled bool,
) SupportedAsset {
	return SupportedAsset{
		Denom:            denom,
		SwapRateToHODL:   swapRateToHODL,
		SwapRateFromHODL: swapRateFromHODL,
		MinSwapAmount:    minSwapAmount,
		MaxSwapAmount:    maxSwapAmount,
		Enabled:          enabled,
		AddedAt:          time.Now(),
	}
}

// Validate validates the supported asset
func (sa SupportedAsset) Validate() error {
	if sa.Denom == "" {
		return fmt.Errorf("denom cannot be empty")
	}
	if sa.SwapRateToHODL.IsNil() || sa.SwapRateToHODL.IsNegative() || sa.SwapRateToHODL.IsZero() {
		return fmt.Errorf("swap rate to HODL must be positive")
	}
	if sa.SwapRateFromHODL.IsNil() || sa.SwapRateFromHODL.IsNegative() || sa.SwapRateFromHODL.IsZero() {
		return fmt.Errorf("swap rate from HODL must be positive")
	}
	if sa.MinSwapAmount.IsNegative() {
		return fmt.Errorf("min swap amount cannot be negative")
	}
	if sa.MaxSwapAmount.IsNegative() {
		return fmt.Errorf("max swap amount cannot be negative")
	}
	if !sa.MaxSwapAmount.IsZero() && sa.MaxSwapAmount.LT(sa.MinSwapAmount) {
		return fmt.Errorf("max swap amount must be greater than or equal to min swap amount")
	}
	return nil
}

// SwapDirection represents the direction of a swap
type SwapDirection int

const (
	SwapDirectionIn  SwapDirection = 0 // External asset -> HODL
	SwapDirectionOut SwapDirection = 1 // HODL -> External asset
)

// String returns the string representation of SwapDirection
func (sd SwapDirection) String() string {
	switch sd {
	case SwapDirectionIn:
		return "swap_in"
	case SwapDirectionOut:
		return "swap_out"
	default:
		return "unknown"
	}
}

// SwapRecord tracks individual swap transactions
type SwapRecord struct {
	ID            uint64        `json:"id" yaml:"id"`
	User          string        `json:"user" yaml:"user"`
	Direction     SwapDirection `json:"direction" yaml:"direction"`
	InputDenom    string        `json:"input_denom" yaml:"input_denom"`
	InputAmount   math.Int      `json:"input_amount" yaml:"input_amount"`
	OutputDenom   string        `json:"output_denom" yaml:"output_denom"`
	OutputAmount  math.Int      `json:"output_amount" yaml:"output_amount"`
	FeeAmount     math.Int      `json:"fee_amount" yaml:"fee_amount"`
	SwapRate      math.LegacyDec `json:"swap_rate" yaml:"swap_rate"`
	Timestamp     time.Time     `json:"timestamp" yaml:"timestamp"`
	BlockHeight   int64         `json:"block_height" yaml:"block_height"`
}

// NewSwapRecord creates a new SwapRecord
func NewSwapRecord(
	id uint64,
	user string,
	direction SwapDirection,
	inputDenom string,
	inputAmount math.Int,
	outputDenom string,
	outputAmount math.Int,
	feeAmount math.Int,
	swapRate math.LegacyDec,
	timestamp time.Time,
	blockHeight int64,
) SwapRecord {
	return SwapRecord{
		ID:           id,
		User:         user,
		Direction:    direction,
		InputDenom:   inputDenom,
		InputAmount:  inputAmount,
		OutputDenom:  outputDenom,
		OutputAmount: outputAmount,
		FeeAmount:    feeAmount,
		SwapRate:     swapRate,
		Timestamp:    timestamp,
		BlockHeight:  blockHeight,
	}
}

// Validate validates the swap record
func (sr SwapRecord) Validate() error {
	if sr.User == "" {
		return fmt.Errorf("user cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(sr.User); err != nil {
		return fmt.Errorf("invalid user address: %w", err)
	}
	if sr.InputDenom == "" {
		return fmt.Errorf("input denom cannot be empty")
	}
	if sr.OutputDenom == "" {
		return fmt.Errorf("output denom cannot be empty")
	}
	if sr.InputAmount.IsNegative() {
		return fmt.Errorf("input amount cannot be negative")
	}
	if sr.OutputAmount.IsNegative() {
		return fmt.Errorf("output amount cannot be negative")
	}
	if sr.FeeAmount.IsNegative() {
		return fmt.Errorf("fee amount cannot be negative")
	}
	if sr.SwapRate.IsNil() || sr.SwapRate.IsNegative() || sr.SwapRate.IsZero() {
		return fmt.Errorf("swap rate must be positive")
	}
	return nil
}

// ProposalStatus represents the status of a withdrawal proposal
type ProposalStatus int

const (
	ProposalStatusActive   ProposalStatus = 0
	ProposalStatusApproved ProposalStatus = 1
	ProposalStatusRejected ProposalStatus = 2
	ProposalStatusExpired  ProposalStatus = 3
	ProposalStatusExecuted ProposalStatus = 4
)

// String returns the string representation of ProposalStatus
func (ps ProposalStatus) String() string {
	switch ps {
	case ProposalStatusActive:
		return "active"
	case ProposalStatusApproved:
		return "approved"
	case ProposalStatusRejected:
		return "rejected"
	case ProposalStatusExpired:
		return "expired"
	case ProposalStatusExecuted:
		return "executed"
	default:
		return "unknown"
	}
}

// ReserveWithdrawal represents a governance proposal to withdraw from the reserve
type ReserveWithdrawal struct {
	ID              uint64         `json:"id" yaml:"id"`
	Proposer        string         `json:"proposer" yaml:"proposer"` // Validator proposing withdrawal
	Recipient       string         `json:"recipient" yaml:"recipient"` // Address to receive withdrawn funds
	Denom           string         `json:"denom" yaml:"denom"`
	Amount          math.Int       `json:"amount" yaml:"amount"`
	Reason          string         `json:"reason" yaml:"reason"`
	Status          ProposalStatus `json:"status" yaml:"status"`
	YesVotes        uint64         `json:"yes_votes" yaml:"yes_votes"`
	NoVotes         uint64         `json:"no_votes" yaml:"no_votes"`
	TotalValidators uint64         `json:"total_validators" yaml:"total_validators"`
	CreatedAt       time.Time      `json:"created_at" yaml:"created_at"`
	VotingEndsAt    time.Time      `json:"voting_ends_at" yaml:"voting_ends_at"`
	ExecutedAt      *time.Time     `json:"executed_at,omitempty" yaml:"executed_at,omitempty"`
}

// NewReserveWithdrawal creates a new ReserveWithdrawal proposal
func NewReserveWithdrawal(
	id uint64,
	proposer,
	recipient,
	denom string,
	amount math.Int,
	reason string,
	totalValidators uint64,
	votingPeriod uint64,
) ReserveWithdrawal {
	now := time.Now()
	return ReserveWithdrawal{
		ID:              id,
		Proposer:        proposer,
		Recipient:       recipient,
		Denom:           denom,
		Amount:          amount,
		Reason:          reason,
		Status:          ProposalStatusActive,
		YesVotes:        0,
		NoVotes:         0,
		TotalValidators: totalValidators,
		CreatedAt:       now,
		VotingEndsAt:    now.Add(time.Duration(votingPeriod) * time.Second),
	}
}

// Validate validates the reserve withdrawal
func (rw ReserveWithdrawal) Validate() error {
	if rw.Proposer == "" {
		return fmt.Errorf("proposer cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(rw.Proposer); err != nil {
		return fmt.Errorf("invalid proposer address: %w", err)
	}
	if rw.Recipient == "" {
		return fmt.Errorf("recipient cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(rw.Recipient); err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}
	if rw.Denom == "" {
		return fmt.Errorf("denom cannot be empty")
	}
	if rw.Amount.IsNegative() || rw.Amount.IsZero() {
		return fmt.Errorf("amount must be positive")
	}
	if rw.Reason == "" {
		return fmt.Errorf("reason cannot be empty")
	}
	return nil
}

// IsActive returns whether the proposal is active
func (rw ReserveWithdrawal) IsActive() bool {
	return rw.Status == ProposalStatusActive && time.Now().Before(rw.VotingEndsAt)
}

// IsExpired returns whether the proposal has expired
func (rw ReserveWithdrawal) IsExpired() bool {
	return rw.Status == ProposalStatusActive && time.Now().After(rw.VotingEndsAt)
}

// GetApprovalRate returns the approval rate as a decimal
func (rw ReserveWithdrawal) GetApprovalRate() math.LegacyDec {
	if rw.TotalValidators == 0 {
		return math.LegacyZeroDec()
	}
	return math.LegacyNewDec(int64(rw.YesVotes)).QuoInt64(int64(rw.TotalValidators))
}

// WithdrawalVote represents a validator's vote on a withdrawal proposal
type WithdrawalVote struct {
	ProposalID uint64    `json:"proposal_id" yaml:"proposal_id"`
	Validator  string    `json:"validator" yaml:"validator"`
	Vote       bool      `json:"vote" yaml:"vote"` // true = yes, false = no
	VotedAt    time.Time `json:"voted_at" yaml:"voted_at"`
}

// NewWithdrawalVote creates a new WithdrawalVote
func NewWithdrawalVote(proposalID uint64, validator string, vote bool) WithdrawalVote {
	return WithdrawalVote{
		ProposalID: proposalID,
		Validator:  validator,
		Vote:       vote,
		VotedAt:    time.Now(),
	}
}

// Validate validates the withdrawal vote
func (wv WithdrawalVote) Validate() error {
	if wv.Validator == "" {
		return fmt.Errorf("validator cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(wv.Validator); err != nil {
		return fmt.Errorf("invalid validator address: %w", err)
	}
	return nil
}
