package keeper

import (
	"context"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sharehodl/sharehodl-blockchain/x/staking/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// UpdateParams updates the staking module parameters
// This is a governance-controlled action that requires the governance module authority
func (k msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority is governance module
	if k.authority != msg.Authority {
		return nil, errors.Wrapf(sdkerrors.ErrUnauthorized, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	// Validate the params
	if err := msg.Params.Validate(); err != nil {
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// Set the new params
	if err := k.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeParamsUpdated,
			sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
		),
	)

	k.Logger(ctx).Info("staking params updated",
		"authority", msg.Authority,
		"holder_threshold", msg.Params.HolderThreshold.String(),
		"keeper_threshold", msg.Params.KeeperThreshold.String(),
		"warden_threshold", msg.Params.WardenThreshold.String(),
		"steward_threshold", msg.Params.StewardThreshold.String(),
		"archon_threshold", msg.Params.ArchonThreshold.String(),
		"validator_threshold", msg.Params.ValidatorThreshold.String(),
	)

	return &types.MsgUpdateParamsResponse{}, nil
}

// Stake allows a user to stake HODL
func (k msgServer) Stake(goCtx context.Context, msg *types.MsgStake) (*types.MsgStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid staker address: %s", err)
	}

	if err := k.Keeper.Stake(ctx, staker, msg.Amount); err != nil {
		return nil, err
	}

	// Get updated stake info
	stake, _ := k.GetUserStake(ctx, staker)

	return &types.MsgStakeResponse{
		NewTier:     stake.Tier.String(),
		TotalStaked: stake.StakedAmount,
	}, nil
}

// Unstake allows a user to unstake HODL
func (k msgServer) Unstake(goCtx context.Context, msg *types.MsgUnstake) (*types.MsgUnstakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid staker address: %s", err)
	}

	if err := k.Keeper.Unstake(ctx, staker, msg.Amount); err != nil {
		return nil, err
	}

	// Get updated stake info
	stake, exists := k.GetUserStake(ctx, staker)
	if !exists {
		return &types.MsgUnstakeResponse{
			NewTier:         "None",
			RemainingStaked: math.ZeroInt(),
		}, nil
	}

	return &types.MsgUnstakeResponse{
		NewTier:         stake.Tier.String(),
		RemainingStaked: stake.StakedAmount,
	}, nil
}

// ClaimRewards allows a user to claim their pending rewards
func (k msgServer) ClaimRewards(goCtx context.Context, msg *types.MsgClaimRewards) (*types.MsgClaimRewardsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	staker, err := sdk.AccAddressFromBech32(msg.Staker)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid staker address: %s", err)
	}

	claimed, err := k.Keeper.ClaimRewards(ctx, staker)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimRewardsResponse{
		ClaimedAmount: claimed,
	}, nil
}
