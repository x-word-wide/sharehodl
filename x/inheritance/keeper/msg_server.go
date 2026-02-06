package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreatePlan creates a new inheritance plan
func (ms msgServer) CreatePlan(goCtx context.Context, msg *types.MsgCreatePlan) (*types.MsgCreatePlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Record activity for creator
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	ms.RecordActivity(ctx, creator, "create_plan")

	// Create the plan
	planID := ms.GetNextPlanID(ctx)

	plan := types.InheritancePlan{
		PlanID:           planID,
		Owner:            msg.Creator,
		Beneficiaries:    msg.Beneficiaries,
		InactivityPeriod: msg.InactivityPeriod,
		GracePeriod:      msg.GracePeriod,
		ClaimWindow:      msg.ClaimWindow,
		CharityAddress:   msg.CharityAddress,
		Status:           types.PlanStatusActive,
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
	}

	if err := ms.Keeper.CreatePlan(ctx, plan); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePlanCreated,
			sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(planID)),
			sdk.NewAttribute(types.AttributeKeyOwner, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyInactivityPeriod, msg.InactivityPeriod.String()),
			sdk.NewAttribute(types.AttributeKeyGracePeriod, msg.GracePeriod.String()),
		),
	)

	return &types.MsgCreatePlanResponse{
		PlanID: planID,
	}, nil
}

// UpdatePlan updates an existing inheritance plan
func (ms msgServer) UpdatePlan(goCtx context.Context, msg *types.MsgUpdatePlan) (*types.MsgUpdatePlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Get the plan
	plan, found := ms.GetPlan(ctx, msg.PlanID)
	if !found {
		return nil, types.ErrPlanNotFound
	}

	// Check ownership
	if plan.Owner != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	// Can only update active plans
	if plan.Status != types.PlanStatusActive {
		return nil, types.ErrCannotModifyPlan
	}

	// Record activity
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	ms.RecordActivity(ctx, creator, "update_plan")

	// Delete old beneficiary indices
	for _, beneficiary := range plan.Beneficiaries {
		bAddr, _ := sdk.AccAddressFromBech32(beneficiary.Address)
		ms.DeleteBeneficiaryPlanIndex(ctx, bAddr, plan.PlanID)
	}

	// Update the plan
	plan.Beneficiaries = msg.Beneficiaries
	plan.InactivityPeriod = msg.InactivityPeriod
	plan.GracePeriod = msg.GracePeriod
	plan.ClaimWindow = msg.ClaimWindow
	plan.CharityAddress = msg.CharityAddress
	plan.UpdatedAt = ctx.BlockTime()

	// Validate the updated plan
	if err := plan.Validate(); err != nil {
		return nil, err
	}

	// Validate against params
	params := ms.GetParams(ctx)
	if plan.InactivityPeriod < params.MinInactivityPeriod {
		return nil, types.ErrInvalidInactivity
	}
	if plan.GracePeriod < params.MinGracePeriod {
		return nil, types.ErrInvalidGracePeriod
	}
	if plan.ClaimWindow < params.MinClaimWindow || plan.ClaimWindow > params.MaxClaimWindow {
		return nil, types.ErrInvalidClaimWindow
	}

	// Store the plan
	ms.SetPlan(ctx, plan)

	// Create new beneficiary indices
	for _, beneficiary := range plan.Beneficiaries {
		bAddr, _ := sdk.AccAddressFromBech32(beneficiary.Address)
		ms.SetBeneficiaryPlanIndex(ctx, bAddr, plan.PlanID)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePlanUpdated,
			sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(msg.PlanID)),
			sdk.NewAttribute(types.AttributeKeyOwner, msg.Creator),
		),
	)

	return &types.MsgUpdatePlanResponse{}, nil
}

// CancelPlan cancels an inheritance plan
func (ms msgServer) CancelPlan(goCtx context.Context, msg *types.MsgCancelPlan) (*types.MsgCancelPlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Get the plan
	plan, found := ms.GetPlan(ctx, msg.PlanID)
	if !found {
		return nil, types.ErrPlanNotFound
	}

	// Check ownership
	if plan.Owner != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	// Record activity
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	ms.RecordActivity(ctx, creator, "cancel_plan")

	// Cancel any active trigger
	if trigger, found := ms.GetActiveTrigger(ctx, msg.PlanID); found {
		trigger.Status = types.TriggerStatusCancelled
		ms.SetActiveTrigger(ctx, trigger)
	}

	// Update plan status
	plan.Status = types.PlanStatusCancelled
	plan.UpdatedAt = ctx.BlockTime()
	ms.SetPlan(ctx, plan)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePlanCancelled,
			sdk.NewAttribute(types.AttributeKeyPlanID, fmt.Sprint(msg.PlanID)),
			sdk.NewAttribute(types.AttributeKeyOwner, msg.Creator),
		),
	)

	return &types.MsgCancelPlanResponse{}, nil
}

// ClaimAssets allows a beneficiary to claim their inheritance
func (ms msgServer) ClaimAssets(goCtx context.Context, msg *types.MsgClaimAssets) (*types.MsgClaimAssetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	claimer, _ := sdk.AccAddressFromBech32(msg.Claimer)

	// Record activity for claimer
	ms.RecordActivity(ctx, claimer, "claim_assets")

	// Process the claim
	transferredAssets, err := ms.ClaimInheritance(ctx, claimer, msg.PlanID)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimAssetsResponse{
		TransferredAssets: transferredAssets,
	}, nil
}

// CancelTrigger allows owner to cancel a triggered switch
func (ms msgServer) CancelTrigger(goCtx context.Context, msg *types.MsgCancelTrigger) (*types.MsgCancelTriggerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	creator, _ := sdk.AccAddressFromBech32(msg.Creator)

	// Cancel the trigger (this also records activity)
	err := ms.Keeper.CancelTrigger(ctx, creator, msg.PlanID)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelTriggerResponse{}, nil
}
