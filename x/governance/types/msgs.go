package types

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types for governance module

// MsgServer defines the governance message service
type MsgServer interface {
	SubmitProposal(ctx context.Context, msg *MsgSubmitProposal) (*MsgSubmitProposalResponse, error)
	Vote(ctx context.Context, msg *MsgVote) (*MsgVoteResponse, error)
	VoteWeighted(ctx context.Context, msg *MsgVoteWeighted) (*MsgVoteWeightedResponse, error)
	Deposit(ctx context.Context, msg *MsgDeposit) (*MsgDepositResponse, error)
	CancelProposal(ctx context.Context, msg *MsgCancelProposal) (*MsgCancelProposalResponse, error)
	SetGovernanceParams(ctx context.Context, msg *MsgSetGovernanceParams) (*MsgSetGovernanceParamsResponse, error)
}

// MsgSetGovernanceParams defines a message to update governance parameters
type MsgSetGovernanceParams struct {
	Authority string           `json:"authority"`
	Params    GovernanceParams `json:"params"`
}

func (msg MsgSetGovernanceParams) Route() string { return ModuleName }
func (msg MsgSetGovernanceParams) Type_() string { return "set_governance_params" }
func (msg MsgSetGovernanceParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return fmt.Errorf("invalid authority address: %v", err)
	}
	return nil
}

func (msg MsgSetGovernanceParams) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgSetGovernanceParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// MsgSetGovernanceParamsResponse is the response for setting params
type MsgSetGovernanceParamsResponse struct{}

// MsgSubmitProposal defines a message to submit a governance proposal
type MsgSubmitProposal struct {
	Proposer       string         `json:"proposer"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Type           ProposalType   `json:"type"`
	InitialDeposit sdk.Coins      `json:"initial_deposit"`
	Content        interface{}    `json:"content,omitempty"`
}

func (msg MsgSubmitProposal) Route() string { return ModuleName }
func (msg MsgSubmitProposal) Type_() string { return "submit_proposal" }
func (msg MsgSubmitProposal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Proposer); err != nil {
		return fmt.Errorf("invalid proposer address: %v", err)
	}
	if msg.Title == "" {
		return fmt.Errorf("proposal title cannot be empty")
	}
	if msg.Description == "" {
		return fmt.Errorf("proposal description cannot be empty")
	}
	if !msg.InitialDeposit.IsValid() {
		return fmt.Errorf("invalid initial deposit")
	}
	return nil
}

func (msg MsgSubmitProposal) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Proposer)
	return []sdk.AccAddress{addr}
}

// MsgSubmitProposalResponse is the response for submitting a proposal
type MsgSubmitProposalResponse struct {
	ProposalID uint64 `json:"proposal_id"`
}

// MsgVote defines a message to vote on a proposal
type MsgVote struct {
	ProposalID uint64     `json:"proposal_id"`
	Voter      string     `json:"voter"`
	Option     VoteOption `json:"option"`
}

func (msg MsgVote) Route() string { return ModuleName }
func (msg MsgVote) Type_() string { return "vote" }
func (msg MsgVote) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Voter); err != nil {
		return fmt.Errorf("invalid voter address: %v", err)
	}
	if msg.ProposalID == 0 {
		return fmt.Errorf("invalid proposal id")
	}
	if msg.Option == VoteOptionEmpty {
		return fmt.Errorf("invalid vote option")
	}
	return nil
}

func (msg MsgVote) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgVote) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Voter)
	return []sdk.AccAddress{addr}
}

// MsgVoteResponse is the response for voting
type MsgVoteResponse struct{}

// MsgVoteWeighted defines a message to vote with weights on a proposal
type MsgVoteWeighted struct {
	ProposalID uint64               `json:"proposal_id"`
	Voter      string               `json:"voter"`
	Options    []WeightedVoteOption `json:"options"`
}

func (msg MsgVoteWeighted) Route() string { return ModuleName }
func (msg MsgVoteWeighted) Type_() string { return "vote_weighted" }
func (msg MsgVoteWeighted) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Voter); err != nil {
		return fmt.Errorf("invalid voter address: %v", err)
	}
	if msg.ProposalID == 0 {
		return fmt.Errorf("invalid proposal id")
	}
	if len(msg.Options) == 0 {
		return fmt.Errorf("weighted vote options cannot be empty")
	}
	totalWeight := math.LegacyZeroDec()
	for _, opt := range msg.Options {
		if opt.Weight.IsNegative() || opt.Weight.GT(math.LegacyOneDec()) {
			return fmt.Errorf("vote weight must be between 0 and 1")
		}
		totalWeight = totalWeight.Add(opt.Weight)
	}
	if !totalWeight.Equal(math.LegacyOneDec()) {
		return fmt.Errorf("total vote weight must equal 1")
	}
	return nil
}

func (msg MsgVoteWeighted) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgVoteWeighted) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Voter)
	return []sdk.AccAddress{addr}
}

// MsgVoteWeightedResponse is the response for weighted voting
type MsgVoteWeightedResponse struct{}

// MsgDeposit defines a message to deposit on a proposal
type MsgDeposit struct {
	ProposalID uint64    `json:"proposal_id"`
	Depositor  string    `json:"depositor"`
	Amount     sdk.Coins `json:"amount"`
}

func (msg MsgDeposit) Route() string { return ModuleName }
func (msg MsgDeposit) Type_() string { return "deposit" }
func (msg MsgDeposit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Depositor); err != nil {
		return fmt.Errorf("invalid depositor address: %v", err)
	}
	if msg.ProposalID == 0 {
		return fmt.Errorf("invalid proposal id")
	}
	if !msg.Amount.IsValid() || msg.Amount.IsZero() {
		return fmt.Errorf("invalid deposit amount")
	}
	return nil
}

func (msg MsgDeposit) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Depositor)
	return []sdk.AccAddress{addr}
}

// MsgDepositResponse is the response for depositing
type MsgDepositResponse struct{}

// MsgCancelProposal defines a message to cancel a proposal
type MsgCancelProposal struct {
	ProposalID uint64 `json:"proposal_id"`
	Proposer   string `json:"proposer"`
}

func (msg MsgCancelProposal) Route() string { return ModuleName }
func (msg MsgCancelProposal) Type_() string { return "cancel_proposal" }
func (msg MsgCancelProposal) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Proposer); err != nil {
		return fmt.Errorf("invalid proposer address: %v", err)
	}
	if msg.ProposalID == 0 {
		return fmt.Errorf("invalid proposal id")
	}
	return nil
}

func (msg MsgCancelProposal) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgCancelProposal) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Proposer)
	return []sdk.AccAddress{addr}
}

// MsgCancelProposalResponse is the response for canceling
type MsgCancelProposalResponse struct{}
