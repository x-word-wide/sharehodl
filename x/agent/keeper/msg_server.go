package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/agent/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateAgent creates a new AI agent for a user
func (k msgServer) CreateAgent(goCtx context.Context, msg *types.MsgCreateAgent) (*types.MsgCreateAgentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrInvalidParams, err.Error())
	}

	// Check subscription
	sub, found := k.GetSubscription(ctx, msg.Owner)
	if !found {
		// Auto-create basic subscription
		if err := k.Keeper.SubscribeUser(ctx, msg.Owner, "basic", false); err != nil {
			return nil, fmt.Errorf("failed to create default subscription: %w", err)
		}
		sub, _ = k.GetSubscription(ctx, msg.Owner)
	}

	// Check if subscription is expired
	if ctx.BlockTime().After(sub.ExpiresAt) {
		return nil, types.ErrAgentExpired
	}

	// Create agent
	agent, err := k.Keeper.CreateAgent(ctx, msg.Owner, msg.Name, msg.Description, sub.Tier)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateAgentResponse{
		AgentId:  agent.ID,
		AgentKey: agent.AgentKey,
	}, nil
}

// UpdateAgent updates an existing agent
func (k msgServer) UpdateAgent(goCtx context.Context, msg *types.MsgUpdateAgent) (*types.MsgUpdateAgentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrInvalidParams, err.Error())
	}

	// Get existing agent
	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, types.ErrAgentNotFound
	}

	// Verify ownership
	if agent.Owner != msg.Owner {
		return nil, types.ErrUnauthorized
	}

	// Update fields
	if msg.Name != "" {
		agent.Name = msg.Name
	}
	if msg.Description != "" {
		agent.Description = msg.Description
	}

	// Save updated agent
	if err := k.SetAgent(ctx, agent); err != nil {
		return nil, err
	}

	return &types.MsgUpdateAgentResponse{}, nil
}

// DeleteAgent removes an agent
func (k msgServer) DeleteAgent(goCtx context.Context, msg *types.MsgDeleteAgent) (*types.MsgDeleteAgentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrInvalidParams, err.Error())
	}

	// Get existing agent
	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, types.ErrAgentNotFound
	}

	// Verify ownership
	if agent.Owner != msg.Owner {
		return nil, types.ErrUnauthorized
	}

	// Delete agent
	if err := k.Keeper.DeleteAgent(ctx, msg.AgentId); err != nil {
		return nil, err
	}

	return &types.MsgDeleteAgentResponse{}, nil
}

// ActivateAgent activates or deactivates an agent
func (k msgServer) ActivateAgent(goCtx context.Context, msg *types.MsgActivateAgent) (*types.MsgActivateAgentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrInvalidParams, err.Error())
	}

	// Get existing agent
	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, types.ErrAgentNotFound
	}

	// Verify ownership
	if agent.Owner != msg.Owner {
		return nil, types.ErrUnauthorized
	}

	// Update active status
	agent.IsActive = msg.Active

	// Save updated agent
	if err := k.SetAgent(ctx, agent); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"agent_activation_changed",
			sdk.NewAttribute("agent_id", fmt.Sprintf("%d", msg.AgentId)),
			sdk.NewAttribute("active", fmt.Sprintf("%t", msg.Active)),
		),
	)

	return &types.MsgActivateAgentResponse{}, nil
}

// ExecuteAction executes an action via an agent
func (k msgServer) ExecuteAction(goCtx context.Context, msg *types.MsgExecuteAction) (*types.MsgExecuteActionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrInvalidParams, err.Error())
	}

	// Execute the action
	action, err := k.Keeper.ExecuteAction(ctx, msg.AgentKey, msg.Module, msg.Action, msg.Params)
	if err != nil {
		return nil, err
	}

	return &types.MsgExecuteActionResponse{
		ActionId: action.ID,
		Status:   action.Status,
		Result:   action.Result,
	}, nil
}

// Subscribe creates or updates a subscription
func (k msgServer) Subscribe(goCtx context.Context, msg *types.MsgSubscribe) (*types.MsgSubscribeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrInvalidParams, err.Error())
	}

	// Create/update subscription
	if err := k.Keeper.SubscribeUser(ctx, msg.Owner, msg.Tier, msg.AutoRenew); err != nil {
		return nil, err
	}

	// Get updated subscription
	sub, _ := k.GetSubscription(ctx, msg.Owner)

	return &types.MsgSubscribeResponse{
		ExpiresAt: sub.ExpiresAt.Unix(),
	}, nil
}

// RegenerateAgentKey generates a new API key for an agent
func (k msgServer) RegenerateAgentKey(goCtx context.Context, msg *types.MsgRegenerateAgentKey) (*types.MsgRegenerateAgentKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrInvalidParams, err.Error())
	}

	// Get existing agent
	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, types.ErrAgentNotFound
	}

	// Verify ownership
	if agent.Owner != msg.Owner {
		return nil, types.ErrUnauthorized
	}

	// Delete old key index
	store := k.storeService.OpenKVStore(ctx)
	store.Delete(types.GetAgentByAPIKeyKey(agent.AgentKey))

	// Generate new key
	newKey, err := generateAgentKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new key: %w", err)
	}

	// Update agent
	agent.AgentKey = newKey

	// Save agent
	if err := k.SetAgent(ctx, agent); err != nil {
		return nil, err
	}

	// Create new key index
	k.SetAgentByKey(ctx, newKey, agent.ID)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"agent_key_regenerated",
			sdk.NewAttribute("agent_id", fmt.Sprintf("%d", msg.AgentId)),
		),
	)

	return &types.MsgRegenerateAgentKeyResponse{
		NewAgentKey: newKey,
	}, nil
}

// SetPermissions sets permissions for an agent
func (k msgServer) SetPermissions(goCtx context.Context, msg *types.MsgSetPermissions) (*types.MsgSetPermissionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, errors.Wrap(types.ErrInvalidParams, err.Error())
	}

	// Get existing agent
	agent, found := k.GetAgent(ctx, msg.AgentId)
	if !found {
		return nil, types.ErrAgentNotFound
	}

	// Verify ownership
	if agent.Owner != msg.Owner {
		return nil, types.ErrUnauthorized
	}

	// Get tier to validate modules are allowed
	tier, _ := types.GetSubscriptionTier(agent.SubscriptionTier)

	// Validate and set permissions
	for _, perm := range msg.Permissions {
		moduleAllowed := false
		for _, m := range tier.AllowedModules {
			if m == perm.Module {
				moduleAllowed = true
				break
			}
		}
		if !moduleAllowed {
			return nil, errors.Wrapf(types.ErrModuleNotAllowed, "module %s not allowed for tier %s", perm.Module, agent.SubscriptionTier)
		}
	}

	agent.Permissions = msg.Permissions

	// Save agent
	if err := k.SetAgent(ctx, agent); err != nil {
		return nil, err
	}

	return &types.MsgSetPermissionsResponse{}, nil
}
