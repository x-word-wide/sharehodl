package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants
const (
	TypeMsgSwapIn            = "swap_in"
	TypeMsgSwapOut           = "swap_out"
	TypeMsgValidatorDeposit  = "validator_deposit"
	TypeMsgProposeWithdrawal = "propose_withdrawal"
	TypeMsgVoteWithdrawal    = "vote_withdrawal"
)

// ============================================================================
// SimpleMsgSwapIn - Swap external asset for HODL
// ============================================================================

type SimpleMsgSwapIn struct {
	User        string   `json:"user"`
	InputDenom  string   `json:"input_denom"`
	InputAmount math.Int `json:"input_amount"`
}

func NewMsgSwapIn(user, inputDenom string, inputAmount math.Int) *SimpleMsgSwapIn {
	return &SimpleMsgSwapIn{
		User:        user,
		InputDenom:  inputDenom,
		InputAmount: inputAmount,
	}
}

func (msg SimpleMsgSwapIn) ValidateBasic() error {
	if msg.User == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.User); err != nil {
		return ErrUnauthorized
	}
	if msg.InputDenom == "" {
		return ErrInvalidAsset
	}
	if msg.InputAmount.IsNil() || msg.InputAmount.IsNegative() || msg.InputAmount.IsZero() {
		return ErrZeroAmount
	}
	return nil
}

func (msg SimpleMsgSwapIn) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.User)
	return []sdk.AccAddress{addr}
}

// ============================================================================
// SimpleMsgSwapOut - Swap HODL for external asset
// ============================================================================

type SimpleMsgSwapOut struct {
	User        string   `json:"user"`
	HodlAmount  math.Int `json:"hodl_amount"`
	OutputDenom string   `json:"output_denom"`
}

func NewMsgSwapOut(user string, hodlAmount math.Int, outputDenom string) *SimpleMsgSwapOut {
	return &SimpleMsgSwapOut{
		User:        user,
		HodlAmount:  hodlAmount,
		OutputDenom: outputDenom,
	}
}

func (msg SimpleMsgSwapOut) ValidateBasic() error {
	if msg.User == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.User); err != nil {
		return ErrUnauthorized
	}
	if msg.HodlAmount.IsNil() || msg.HodlAmount.IsNegative() || msg.HodlAmount.IsZero() {
		return ErrZeroAmount
	}
	if msg.OutputDenom == "" {
		return ErrInvalidAsset
	}
	return nil
}

func (msg SimpleMsgSwapOut) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.User)
	return []sdk.AccAddress{addr}
}

// ============================================================================
// SimpleMsgValidatorDeposit - Validator deposits assets to reserve
// ============================================================================

type SimpleMsgValidatorDeposit struct {
	Validator string   `json:"validator"`
	Denom     string   `json:"denom"`
	Amount    math.Int `json:"amount"`
}

func NewMsgValidatorDeposit(validator, denom string, amount math.Int) *SimpleMsgValidatorDeposit {
	return &SimpleMsgValidatorDeposit{
		Validator: validator,
		Denom:     denom,
		Amount:    amount,
	}
}

func (msg SimpleMsgValidatorDeposit) ValidateBasic() error {
	if msg.Validator == "" {
		return ErrNotValidator
	}
	if _, err := sdk.AccAddressFromBech32(msg.Validator); err != nil {
		return ErrInvalidValidator
	}
	if msg.Denom == "" {
		return ErrInvalidAsset
	}
	if msg.Amount.IsNil() || msg.Amount.IsNegative() || msg.Amount.IsZero() {
		return ErrZeroAmount
	}
	return nil
}

func (msg SimpleMsgValidatorDeposit) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Validator)
	return []sdk.AccAddress{addr}
}

// ============================================================================
// SimpleMsgProposeWithdrawal - Propose a withdrawal from reserve
// ============================================================================

type SimpleMsgProposeWithdrawal struct {
	Proposer  string   `json:"proposer"`
	Recipient string   `json:"recipient"`
	Denom     string   `json:"denom"`
	Amount    math.Int `json:"amount"`
	Reason    string   `json:"reason"`
}

func NewMsgProposeWithdrawal(proposer, recipient, denom string, amount math.Int, reason string) *SimpleMsgProposeWithdrawal {
	return &SimpleMsgProposeWithdrawal{
		Proposer:  proposer,
		Recipient: recipient,
		Denom:     denom,
		Amount:    amount,
		Reason:    reason,
	}
}

func (msg SimpleMsgProposeWithdrawal) ValidateBasic() error {
	if msg.Proposer == "" {
		return ErrNotValidator
	}
	if _, err := sdk.AccAddressFromBech32(msg.Proposer); err != nil {
		return ErrInvalidValidator
	}
	if msg.Recipient == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return ErrUnauthorized
	}
	if msg.Denom == "" {
		return ErrInvalidAsset
	}
	if msg.Amount.IsNil() || msg.Amount.IsNegative() || msg.Amount.IsZero() {
		return ErrZeroAmount
	}
	if msg.Reason == "" {
		return ErrInvalidProposal
	}
	return nil
}

func (msg SimpleMsgProposeWithdrawal) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Proposer)
	return []sdk.AccAddress{addr}
}

// ============================================================================
// SimpleMsgVoteWithdrawal - Vote on a withdrawal proposal
// ============================================================================

type SimpleMsgVoteWithdrawal struct {
	Validator  string `json:"validator"`
	ProposalID uint64 `json:"proposal_id"`
	Vote       bool   `json:"vote"` // true = yes, false = no
}

func NewMsgVoteWithdrawal(validator string, proposalID uint64, vote bool) *SimpleMsgVoteWithdrawal {
	return &SimpleMsgVoteWithdrawal{
		Validator:  validator,
		ProposalID: proposalID,
		Vote:       vote,
	}
}

func (msg SimpleMsgVoteWithdrawal) ValidateBasic() error {
	if msg.Validator == "" {
		return ErrNotValidator
	}
	if _, err := sdk.AccAddressFromBech32(msg.Validator); err != nil {
		return ErrInvalidValidator
	}
	if msg.ProposalID == 0 {
		return ErrProposalNotFound
	}
	return nil
}

func (msg SimpleMsgVoteWithdrawal) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Validator)
	return []sdk.AccAddress{addr}
}

// ============================================================================
// Message Responses
// ============================================================================

type MsgSwapInResponse struct {
	HodlReceived math.Int `json:"hodl_received"`
	FeeCharged   math.Int `json:"fee_charged"`
	SwapID       uint64   `json:"swap_id"`
}

type MsgSwapOutResponse struct {
	AssetReceived math.Int `json:"asset_received"`
	FeeCharged    math.Int `json:"fee_charged"`
	SwapID        uint64   `json:"swap_id"`
}

type MsgValidatorDepositResponse struct {
	Success bool `json:"success"`
}

type MsgProposeWithdrawalResponse struct {
	ProposalID uint64 `json:"proposal_id"`
}

type MsgVoteWithdrawalResponse struct {
	Success bool `json:"success"`
}
