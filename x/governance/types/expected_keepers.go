package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgServer defines the server API for Msg service
type MsgServer interface {
	// SubmitProposal submits a governance proposal
	SubmitProposal(context.Context, *MsgSubmitProposal) (*MsgSubmitProposalResponse, error)
	
	// Vote votes on a governance proposal
	Vote(context.Context, *MsgVote) (*MsgVoteResponse, error)
	
	// VoteWeighted casts a weighted vote on a governance proposal
	VoteWeighted(context.Context, *MsgVoteWeighted) (*MsgVoteWeightedResponse, error)
	
	// Deposit deposits tokens to a governance proposal
	Deposit(context.Context, *MsgDeposit) (*MsgDepositResponse, error)
	
	// CancelProposal cancels a governance proposal
	CancelProposal(context.Context, *MsgCancelProposal) (*MsgCancelProposalResponse, error)
	
	// SetGovernanceParams sets governance module parameters
	SetGovernanceParams(context.Context, *MsgSetGovernanceParams) (*MsgSetGovernanceParamsResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) SubmitProposal(context.Context, *MsgSubmitProposal) (*MsgSubmitProposalResponse, error) {
	panic("implement me")
}

func (UnimplementedMsgServer) Vote(context.Context, *MsgVote) (*MsgVoteResponse, error) {
	panic("implement me")
}

func (UnimplementedMsgServer) VoteWeighted(context.Context, *MsgVoteWeighted) (*MsgVoteWeightedResponse, error) {
	panic("implement me")
}

func (UnimplementedMsgServer) Deposit(context.Context, *MsgDeposit) (*MsgDepositResponse, error) {
	panic("implement me")
}

func (UnimplementedMsgServer) CancelProposal(context.Context, *MsgCancelProposal) (*MsgCancelProposalResponse, error) {
	panic("implement me")
}

func (UnimplementedMsgServer) SetGovernanceParams(context.Context, *MsgSetGovernanceParams) (*MsgSetGovernanceParamsResponse, error) {
	panic("implement me")
}