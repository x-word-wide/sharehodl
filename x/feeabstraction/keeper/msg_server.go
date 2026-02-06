package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateParams updates module parameters (governance only)
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority
	if msg.Authority != ms.Keeper.GetAuthority() {
		return nil, types.ErrInvalidAuthority
	}

	// Update parameters
	if err := ms.Keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"params_updated",
			sdk.NewAttribute("authority", msg.Authority),
		),
	)

	return &types.MsgUpdateParamsResponse{}, nil
}

// GrantTreasuryAllowance grants a fee allowance (governance only)
func (ms msgServer) GrantTreasuryAllowance(goCtx context.Context, msg *types.MsgGrantTreasuryAllowance) (*types.MsgGrantTreasuryAllowanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority
	if msg.Authority != ms.Keeper.GetAuthority() {
		return nil, types.ErrInvalidAuthority
	}

	// Create grant
	if err := ms.Keeper.CreateGrant(
		ctx,
		msg.Grantee,
		msg.AllowedAmount,
		msg.ExpirationHeight,
		msg.AllowedEquities,
		0, // No proposal ID in direct message
	); err != nil {
		return nil, err
	}

	return &types.MsgGrantTreasuryAllowanceResponse{}, nil
}

// RevokeTreasuryAllowance revokes a fee allowance (governance only)
func (ms msgServer) RevokeTreasuryAllowance(goCtx context.Context, msg *types.MsgRevokeTreasuryAllowance) (*types.MsgRevokeTreasuryAllowanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority
	if msg.Authority != ms.Keeper.GetAuthority() {
		return nil, types.ErrInvalidAuthority
	}

	// Revoke grant
	if err := ms.Keeper.RevokeGrant(ctx, msg.Grantee); err != nil {
		return nil, err
	}

	return &types.MsgRevokeTreasuryAllowanceResponse{}, nil
}

// FundTreasury adds funds to the protocol treasury
func (ms msgServer) FundTreasury(goCtx context.Context, msg *types.MsgFundTreasury) (*types.MsgFundTreasuryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	funderAddr, err := sdk.AccAddressFromBech32(msg.Funder)
	if err != nil {
		return nil, types.ErrInvalidAddress
	}

	// Fund treasury
	if err := ms.Keeper.FundTreasury(ctx, funderAddr, msg.Amount); err != nil {
		return nil, err
	}

	return &types.MsgFundTreasuryResponse{}, nil
}
