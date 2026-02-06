package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// SwapIn handles SimpleMsgSwapIn
func (ms msgServer) SwapIn(goCtx context.Context, msg *types.SimpleMsgSwapIn) (*types.MsgSwapInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	user, err := sdk.AccAddressFromBech32(msg.User)
	if err != nil {
		return nil, err
	}

	hodlReceived, fee, err := ms.Keeper.SwapIn(ctx, user, msg.InputDenom, msg.InputAmount)
	if err != nil {
		return nil, err
	}

	return &types.MsgSwapInResponse{
		HodlReceived: hodlReceived,
		FeeCharged:   fee,
		SwapID:       ms.Keeper.GetNextSwapID(ctx) - 1, // Return the ID that was just used
	}, nil
}

// SwapOut handles SimpleMsgSwapOut
func (ms msgServer) SwapOut(goCtx context.Context, msg *types.SimpleMsgSwapOut) (*types.MsgSwapOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	user, err := sdk.AccAddressFromBech32(msg.User)
	if err != nil {
		return nil, err
	}

	assetReceived, fee, err := ms.Keeper.SwapOut(ctx, user, msg.HodlAmount, msg.OutputDenom)
	if err != nil {
		return nil, err
	}

	return &types.MsgSwapOutResponse{
		AssetReceived: assetReceived,
		FeeCharged:    fee,
		SwapID:        ms.Keeper.GetNextSwapID(ctx) - 1,
	}, nil
}

// ValidatorDeposit handles SimpleMsgValidatorDeposit
func (ms msgServer) ValidatorDeposit(goCtx context.Context, msg *types.SimpleMsgValidatorDeposit) (*types.MsgValidatorDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validator, err := sdk.AccAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	if err := ms.Keeper.ValidatorDeposit(ctx, validator, msg.Denom, msg.Amount); err != nil {
		return nil, err
	}

	return &types.MsgValidatorDepositResponse{
		Success: true,
	}, nil
}

// ProposeWithdrawal handles SimpleMsgProposeWithdrawal
func (ms msgServer) ProposeWithdrawal(goCtx context.Context, msg *types.SimpleMsgProposeWithdrawal) (*types.MsgProposeWithdrawalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	proposer, err := sdk.AccAddressFromBech32(msg.Proposer)
	if err != nil {
		return nil, err
	}

	recipient, err := sdk.AccAddressFromBech32(msg.Recipient)
	if err != nil {
		return nil, err
	}

	proposalID, err := ms.Keeper.ProposeWithdrawal(ctx, proposer, recipient, msg.Denom, msg.Amount, msg.Reason)
	if err != nil {
		return nil, err
	}

	return &types.MsgProposeWithdrawalResponse{
		ProposalID: proposalID,
	}, nil
}

// VoteWithdrawal handles SimpleMsgVoteWithdrawal
func (ms msgServer) VoteWithdrawal(goCtx context.Context, msg *types.SimpleMsgVoteWithdrawal) (*types.MsgVoteWithdrawalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validator, err := sdk.AccAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	if err := ms.Keeper.VoteOnWithdrawal(ctx, validator, msg.ProposalID, msg.Vote); err != nil {
		return nil, err
	}

	return &types.MsgVoteWithdrawalResponse{
		Success: true,
	}, nil
}
