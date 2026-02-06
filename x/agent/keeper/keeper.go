package keeper

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/agent/types"
)

// Keeper maintains the agent module state
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	logger       log.Logger
	authority    string

	// Other module keepers for executing actions
	bankKeeper BankKeeper
	dexKeeper  DEXKeeper
	hodlKeeper HODLKeeper
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// DEXKeeper defines the expected DEX keeper interface
type DEXKeeper interface {
	PlaceOrder(ctx sdk.Context, owner string, marketID uint64, side string, orderType string, price, quantity math.LegacyDec) (uint64, error)
	CancelOrder(ctx sdk.Context, owner string, orderID uint64) error
}

// HODLKeeper defines the expected HODL keeper interface
type HODLKeeper interface {
	MintHODL(ctx sdk.Context, owner string, collateral sdk.Coins, hodlAmount math.Int) error
	BurnHODL(ctx sdk.Context, owner string, hodlAmount math.Int) error
}

// NewKeeper creates a new agent keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
	bankKeeper BankKeeper,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		logger:       logger,
		authority:    authority,
		bankKeeper:   bankKeeper,
	}
}

// SetDEXKeeper sets the DEX keeper (called after initialization to avoid circular deps)
func (k *Keeper) SetDEXKeeper(dexKeeper DEXKeeper) {
	k.dexKeeper = dexKeeper
}

// SetHODLKeeper sets the HODL keeper (called after initialization to avoid circular deps)
func (k *Keeper) SetHODLKeeper(hodlKeeper HODLKeeper) {
	k.hodlKeeper = hodlKeeper
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return k.logger.With("module", "x/"+types.ModuleName)
}

// GetAuthority returns the module authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// ============================================================================
// Agent Management
// ============================================================================

// CreateAgent creates a new agent for a user
func (k Keeper) CreateAgent(ctx sdk.Context, owner, name, description, subscriptionTier string) (types.Agent, error) {
	// Validate owner address
	if _, err := sdk.AccAddressFromBech32(owner); err != nil {
		return types.Agent{}, types.ErrInvalidOwner
	}

	// Validate subscription tier
	tier, found := types.GetSubscriptionTier(subscriptionTier)
	if !found {
		return types.Agent{}, types.ErrInvalidSubscription
	}

	// Check if user already has max agents for their tier
	agentCount := k.GetAgentCountByOwner(ctx, owner)
	if agentCount >= uint64(tier.MaxAgents) {
		return types.Agent{}, types.ErrMaxAgentsReached
	}

	// Generate unique agent key
	agentKey, err := generateAgentKey()
	if err != nil {
		return types.Agent{}, fmt.Errorf("failed to generate agent key: %w", err)
	}

	// Check if key already exists (unlikely but possible)
	if k.AgentExistsByKey(ctx, agentKey) {
		return types.Agent{}, types.ErrAgentAlreadyExists
	}

	// Get next agent ID
	agentID := k.GetNextAgentID(ctx)

	// Calculate expiry (1 month from now)
	expiry := ctx.BlockTime().AddDate(0, 1, 0)

	// Create agent with tier-appropriate limits
	agent := types.NewAgent(agentID, owner, agentKey, name, description, subscriptionTier, expiry)

	// Set tier-appropriate limits
	agent.RateLimits.MaxActionsPerDay = tier.MaxActionsPerDay
	agent.SpendingLimits.DailyLimit = tier.MaxDailySpend

	// Set default permissions for allowed modules
	for _, module := range tier.AllowedModules {
		agent.Permissions = append(agent.Permissions, types.Permission{
			Module:    module,
			Actions:   []string{"*"}, // All actions within module
			MaxAmount: tier.MaxDailySpend,
		})
	}

	// Save agent
	if err := k.SetAgent(ctx, agent); err != nil {
		return types.Agent{}, err
	}

	// Create index by key
	k.SetAgentByKey(ctx, agentKey, agentID)

	// Increment counter
	k.SetNextAgentID(ctx, agentID+1)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"agent_created",
			sdk.NewAttribute("agent_id", fmt.Sprintf("%d", agentID)),
			sdk.NewAttribute("owner", owner),
			sdk.NewAttribute("name", name),
			sdk.NewAttribute("tier", subscriptionTier),
		),
	)

	return agent, nil
}

// GetAgent retrieves an agent by ID
func (k Keeper) GetAgent(ctx sdk.Context, agentID uint64) (types.Agent, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetAgentKey(agentID)

	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.Agent{}, false
	}

	var agent types.Agent
	if err := json.Unmarshal(bz, &agent); err != nil {
		k.Logger(ctx).Error("failed to unmarshal agent", "error", err)
		return types.Agent{}, false
	}

	return agent, true
}

// GetAgentByKey retrieves an agent by its API key
func (k Keeper) GetAgentByKey(ctx sdk.Context, agentKey string) (types.Agent, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetAgentByAPIKeyKey(agentKey)

	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.Agent{}, false
	}

	// bz contains the agent ID
	var agentID uint64
	if err := json.Unmarshal(bz, &agentID); err != nil {
		return types.Agent{}, false
	}

	return k.GetAgent(ctx, agentID)
}

// SetAgent stores an agent
func (k Keeper) SetAgent(ctx sdk.Context, agent types.Agent) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetAgentKey(agent.ID)

	bz, err := json.Marshal(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent: %w", err)
	}

	return store.Set(key, bz)
}

// SetAgentByKey creates an index from agent key to agent ID
func (k Keeper) SetAgentByKey(ctx sdk.Context, agentKey string, agentID uint64) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetAgentByAPIKeyKey(agentKey)
	bz, _ := json.Marshal(agentID)
	store.Set(key, bz)
}

// AgentExistsByKey checks if an agent with the given key exists
func (k Keeper) AgentExistsByKey(ctx sdk.Context, agentKey string) bool {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetAgentByAPIKeyKey(agentKey)
	has, _ := store.Has(key)
	return has
}

// DeleteAgent removes an agent
func (k Keeper) DeleteAgent(ctx sdk.Context, agentID uint64) error {
	agent, found := k.GetAgent(ctx, agentID)
	if !found {
		return types.ErrAgentNotFound
	}

	store := k.storeService.OpenKVStore(ctx)

	// Delete main record
	if err := store.Delete(types.GetAgentKey(agentID)); err != nil {
		return err
	}

	// Delete key index
	if err := store.Delete(types.GetAgentByAPIKeyKey(agent.AgentKey)); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"agent_deleted",
			sdk.NewAttribute("agent_id", fmt.Sprintf("%d", agentID)),
			sdk.NewAttribute("owner", agent.Owner),
		),
	)

	return nil
}

// GetAgentCountByOwner returns the number of agents owned by an address
func (k Keeper) GetAgentCountByOwner(ctx sdk.Context, owner string) uint64 {
	var count uint64
	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Owner == owner {
			count++
		}
		return false
	})
	return count
}

// GetAgentsByOwner returns all agents owned by an address
func (k Keeper) GetAgentsByOwner(ctx sdk.Context, owner string) []types.Agent {
	var agents []types.Agent
	k.IterateAgents(ctx, func(agent types.Agent) bool {
		if agent.Owner == owner {
			agents = append(agents, agent)
		}
		return false
	})
	return agents
}

// IterateAgents iterates over all agents
func (k Keeper) IterateAgents(ctx sdk.Context, fn func(agent types.Agent) bool) {
	store := k.storeService.OpenKVStore(ctx)
	iter, err := store.Iterator(types.AgentPrefix, types.PrefixEndBytes(types.AgentPrefix))
	if err != nil {
		return
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var agent types.Agent
		if err := json.Unmarshal(iter.Value(), &agent); err != nil {
			continue
		}
		if fn(agent) {
			break
		}
	}
}

// GetNextAgentID returns the next available agent ID
func (k Keeper) GetNextAgentID(ctx sdk.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.AgentCounterKey)
	if err != nil || bz == nil {
		return 1
	}

	var id uint64
	if err := json.Unmarshal(bz, &id); err != nil {
		return 1
	}
	return id
}

// SetNextAgentID sets the next agent ID
func (k Keeper) SetNextAgentID(ctx sdk.Context, id uint64) {
	store := k.storeService.OpenKVStore(ctx)
	bz, _ := json.Marshal(id)
	store.Set(types.AgentCounterKey, bz)
}

// ============================================================================
// Agent Authentication and Authorization
// ============================================================================

// AuthenticateAgent validates an agent key and returns the agent if valid
func (k Keeper) AuthenticateAgent(ctx sdk.Context, agentKey string) (types.Agent, error) {
	agent, found := k.GetAgentByKey(ctx, agentKey)
	if !found {
		return types.Agent{}, types.ErrAgentNotFound
	}

	// Check if agent is active
	if !agent.IsActive {
		return types.Agent{}, types.ErrAgentNotActive
	}

	// Check if subscription expired
	if ctx.BlockTime().After(agent.ExpiresAt) {
		return types.Agent{}, types.ErrAgentExpired
	}

	return agent, nil
}

// AuthorizeAction checks if an agent can perform a specific action
func (k Keeper) AuthorizeAction(ctx sdk.Context, agent types.Agent, module, action string, amount math.Int) error {
	// Reset rate limits if needed
	k.resetRateLimitsIfNeeded(ctx, &agent)
	k.resetSpendingLimitsIfNeeded(ctx, &agent)

	// Check permission
	if !agent.HasPermission(module, action) {
		return types.ErrPermissionDenied
	}

	// Check module allowed for tier
	tier, _ := types.GetSubscriptionTier(agent.SubscriptionTier)
	moduleAllowed := false
	for _, m := range tier.AllowedModules {
		if m == module {
			moduleAllowed = true
			break
		}
	}
	if !moduleAllowed {
		return types.ErrModuleNotAllowed
	}

	// Check rate limits
	if !agent.CanExecute() {
		return types.ErrRateLimitExceeded
	}

	// Check spending limits if amount specified
	if !amount.IsZero() && !agent.CanSpend(amount) {
		return types.ErrSpendingLimitExceeded
	}

	return nil
}

// resetRateLimitsIfNeeded resets rate limit counters if periods have elapsed
// SECURITY FIX: Added validation to prevent manipulation of reset times
func (k Keeper) resetRateLimitsIfNeeded(ctx sdk.Context, agent *types.Agent) {
	now := ctx.BlockTime()

	// SECURITY: Validate that LastMinuteReset is not in the future (prevent manipulation)
	// If reset time is in the future, reset it to now (agent creation time or last valid reset)
	if agent.RateLimits.LastMinuteReset.After(now) {
		k.Logger(ctx).Warn("agent rate limit reset time in future, resetting",
			"agent_id", agent.ID,
			"invalid_reset", agent.RateLimits.LastMinuteReset,
			"current_time", now)
		agent.RateLimits.LastMinuteReset = now
		agent.RateLimits.ActionsThisMinute = 0
	}

	if agent.RateLimits.LastHourReset.After(now) {
		k.Logger(ctx).Warn("agent rate limit reset time in future, resetting",
			"agent_id", agent.ID)
		agent.RateLimits.LastHourReset = now
		agent.RateLimits.ActionsThisHour = 0
	}

	if agent.RateLimits.LastDayReset.After(now) {
		k.Logger(ctx).Warn("agent rate limit reset time in future, resetting",
			"agent_id", agent.ID)
		agent.RateLimits.LastDayReset = now
		agent.RateLimits.ActionsThisDay = 0
	}

	// Reset minute counter
	if now.Sub(agent.RateLimits.LastMinuteReset) >= time.Minute {
		agent.RateLimits.ActionsThisMinute = 0
		agent.RateLimits.LastMinuteReset = now
	}

	// Reset hour counter
	if now.Sub(agent.RateLimits.LastHourReset) >= time.Hour {
		agent.RateLimits.ActionsThisHour = 0
		agent.RateLimits.LastHourReset = now
	}

	// Reset day counter
	if now.Sub(agent.RateLimits.LastDayReset) >= 24*time.Hour {
		agent.RateLimits.ActionsThisDay = 0
		agent.RateLimits.LastDayReset = now
	}
}

// resetSpendingLimitsIfNeeded resets spending counters if periods have elapsed
// SECURITY FIX: Added validation to prevent manipulation of reset times
func (k Keeper) resetSpendingLimitsIfNeeded(ctx sdk.Context, agent *types.Agent) {
	now := ctx.BlockTime()

	// SECURITY: Validate that reset times are not in the future (prevent manipulation)
	if agent.SpendingLimits.ResetDay.After(now) {
		k.Logger(ctx).Warn("agent spending limit reset time in future, resetting",
			"agent_id", agent.ID,
			"invalid_reset", agent.SpendingLimits.ResetDay,
			"current_time", now)
		agent.SpendingLimits.ResetDay = now
		agent.SpendingLimits.TotalSpentDay = math.ZeroInt()
	}

	if agent.SpendingLimits.ResetMonth.After(now) {
		k.Logger(ctx).Warn("agent spending limit reset time in future, resetting",
			"agent_id", agent.ID)
		agent.SpendingLimits.ResetMonth = now
		agent.SpendingLimits.TotalSpentMonth = math.ZeroInt()
	}

	// Reset daily spending
	if now.Sub(agent.SpendingLimits.ResetDay) >= 24*time.Hour {
		agent.SpendingLimits.TotalSpentDay = math.ZeroInt()
		agent.SpendingLimits.ResetDay = now
	}

	// Reset monthly spending (30 days)
	if now.Sub(agent.SpendingLimits.ResetMonth) >= 30*24*time.Hour {
		agent.SpendingLimits.TotalSpentMonth = math.ZeroInt()
		agent.SpendingLimits.ResetMonth = now
	}
}

// ============================================================================
// Action Execution
// ============================================================================

// ExecuteAction executes an action on behalf of a user via their agent
func (k Keeper) ExecuteAction(ctx sdk.Context, agentKey, module, action string, params map[string]interface{}) (types.AgentAction, error) {
	// Authenticate agent
	agent, err := k.AuthenticateAgent(ctx, agentKey)
	if err != nil {
		return types.AgentAction{}, err
	}

	// Parse amount from params if present
	amount := math.ZeroInt()
	if amtStr, ok := params["amount"].(string); ok {
		parsedAmt, ok := math.NewIntFromString(amtStr)
		if ok {
			amount = parsedAmt
		}
	}

	// Authorize action
	if err := k.AuthorizeAction(ctx, agent, module, action, amount); err != nil {
		return types.AgentAction{}, err
	}

	// Create action record
	actionID := k.GetNextActionID(ctx)
	paramsJSON, _ := json.Marshal(params)

	agentAction := types.AgentAction{
		ID:          actionID,
		AgentID:     agent.ID,
		Owner:       agent.Owner,
		Module:      module,
		Action:      action,
		Params:      string(paramsJSON),
		Amount:      amount,
		Status:      "pending",
		Timestamp:   ctx.BlockTime(),
		BlockHeight: ctx.BlockHeight(),
	}

	// Execute the action based on module
	var result interface{}
	var execErr error

	switch module {
	case "dex":
		result, execErr = k.executeDEXAction(ctx, agent.Owner, action, params)
	case "hodl":
		result, execErr = k.executeHODLAction(ctx, agent.Owner, action, params)
	case "bank":
		result, execErr = k.executeBankAction(ctx, agent.Owner, action, params)
	default:
		execErr = fmt.Errorf("unsupported module: %s", module)
	}

	// Update action record with result
	if execErr != nil {
		agentAction.Status = "failed"
		agentAction.Result = execErr.Error()
	} else {
		agentAction.Status = "executed"
		resultJSON, _ := json.Marshal(result)
		agentAction.Result = string(resultJSON)

		// Update agent stats
		agent.TotalActions++
		agent.LastActiveAt = ctx.BlockTime()
		agent.RateLimits.ActionsThisMinute++
		agent.RateLimits.ActionsThisHour++
		agent.RateLimits.ActionsThisDay++

		// Update spending
		if !amount.IsZero() {
			agent.SpendingLimits.TotalSpentDay = agent.SpendingLimits.TotalSpentDay.Add(amount)
			agent.SpendingLimits.TotalSpentMonth = agent.SpendingLimits.TotalSpentMonth.Add(amount)
		}

		// Save updated agent
		k.SetAgent(ctx, agent)
	}

	// Save action record
	k.SetAction(ctx, agentAction)
	k.SetNextActionID(ctx, actionID+1)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"agent_action",
			sdk.NewAttribute("action_id", fmt.Sprintf("%d", actionID)),
			sdk.NewAttribute("agent_id", fmt.Sprintf("%d", agent.ID)),
			sdk.NewAttribute("owner", agent.Owner),
			sdk.NewAttribute("module", module),
			sdk.NewAttribute("action", action),
			sdk.NewAttribute("status", agentAction.Status),
		),
	)

	if execErr != nil {
		return agentAction, types.ErrActionFailed
	}

	return agentAction, nil
}

// executeDEXAction executes a DEX-related action
func (k Keeper) executeDEXAction(ctx sdk.Context, owner, action string, params map[string]interface{}) (interface{}, error) {
	if k.dexKeeper == nil {
		return nil, fmt.Errorf("DEX keeper not initialized")
	}

	switch action {
	case "place_order":
		marketID := uint64(params["market_id"].(float64))
		side := params["side"].(string)
		orderType := params["order_type"].(string)
		price, _ := math.LegacyNewDecFromStr(params["price"].(string))
		quantity, _ := math.LegacyNewDecFromStr(params["quantity"].(string))

		orderID, err := k.dexKeeper.PlaceOrder(ctx, owner, marketID, side, orderType, price, quantity)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"order_id": orderID}, nil

	case "cancel_order":
		orderID := uint64(params["order_id"].(float64))
		err := k.dexKeeper.CancelOrder(ctx, owner, orderID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"cancelled": true}, nil

	default:
		return nil, fmt.Errorf("unsupported DEX action: %s", action)
	}
}

// executeHODLAction executes a HODL-related action
func (k Keeper) executeHODLAction(ctx sdk.Context, owner, action string, params map[string]interface{}) (interface{}, error) {
	if k.hodlKeeper == nil {
		return nil, fmt.Errorf("HODL keeper not initialized")
	}

	switch action {
	case "mint":
		amountStr := params["amount"].(string)
		amount, ok := math.NewIntFromString(amountStr)
		if !ok {
			return nil, fmt.Errorf("invalid amount")
		}

		collateralStr := params["collateral"].(string)
		collateral, err := sdk.ParseCoinsNormalized(collateralStr)
		if err != nil {
			return nil, fmt.Errorf("invalid collateral: %w", err)
		}

		err = k.hodlKeeper.MintHODL(ctx, owner, collateral, amount)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"minted": amount.String()}, nil

	case "burn":
		amountStr := params["amount"].(string)
		amount, ok := math.NewIntFromString(amountStr)
		if !ok {
			return nil, fmt.Errorf("invalid amount")
		}

		err := k.hodlKeeper.BurnHODL(ctx, owner, amount)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"burned": amount.String()}, nil

	default:
		return nil, fmt.Errorf("unsupported HODL action: %s", action)
	}
}

// executeBankAction executes a bank-related action
func (k Keeper) executeBankAction(ctx sdk.Context, owner, action string, params map[string]interface{}) (interface{}, error) {
	switch action {
	case "send":
		toAddr := params["to"].(string)
		amountStr := params["amount"].(string)

		amount, err := sdk.ParseCoinsNormalized(amountStr)
		if err != nil {
			return nil, fmt.Errorf("invalid amount: %w", err)
		}

		fromAcc, _ := sdk.AccAddressFromBech32(owner)
		toAcc, err := sdk.AccAddressFromBech32(toAddr)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient: %w", err)
		}

		err = k.bankKeeper.SendCoins(ctx, fromAcc, toAcc, amount)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"sent": amount.String()}, nil

	case "balance":
		acc, _ := sdk.AccAddressFromBech32(owner)
		balances := k.bankKeeper.GetAllBalances(ctx, acc)
		return map[string]interface{}{"balances": balances.String()}, nil

	default:
		return nil, fmt.Errorf("unsupported bank action: %s", action)
	}
}

// ============================================================================
// Action Records
// ============================================================================

// SetAction stores an action record
func (k Keeper) SetAction(ctx sdk.Context, action types.AgentAction) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetActionKey(action.ID)

	bz, err := json.Marshal(action)
	if err != nil {
		return fmt.Errorf("failed to marshal action: %w", err)
	}

	// Store main record
	if err := store.Set(key, bz); err != nil {
		return err
	}

	// Store index by agent
	indexKey := types.GetActionByAgentKey(action.AgentID, action.ID)
	return store.Set(indexKey, bz)
}

// GetAction retrieves an action by ID
func (k Keeper) GetAction(ctx sdk.Context, actionID uint64) (types.AgentAction, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetActionKey(actionID)

	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.AgentAction{}, false
	}

	var action types.AgentAction
	if err := json.Unmarshal(bz, &action); err != nil {
		return types.AgentAction{}, false
	}

	return action, true
}

// GetActionsByAgent returns all actions for an agent
func (k Keeper) GetActionsByAgent(ctx sdk.Context, agentID uint64) []types.AgentAction {
	var actions []types.AgentAction
	store := k.storeService.OpenKVStore(ctx)

	prefix := types.AgentActionIteratorKey(agentID)
	iter, err := store.Iterator(prefix, types.PrefixEndBytes(prefix))
	if err != nil {
		return actions
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var action types.AgentAction
		if err := json.Unmarshal(iter.Value(), &action); err != nil {
			continue
		}
		actions = append(actions, action)
	}

	return actions
}

// GetNextActionID returns the next available action ID
func (k Keeper) GetNextActionID(ctx sdk.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ActionCounterKey)
	if err != nil || bz == nil {
		return 1
	}

	var id uint64
	if err := json.Unmarshal(bz, &id); err != nil {
		return 1
	}
	return id
}

// SetNextActionID sets the next action ID
func (k Keeper) SetNextActionID(ctx sdk.Context, id uint64) {
	store := k.storeService.OpenKVStore(ctx)
	bz, _ := json.Marshal(id)
	store.Set(types.ActionCounterKey, bz)
}

// ============================================================================
// Subscription Management
// ============================================================================

// Subscription represents a user's subscription
type Subscription struct {
	Owner     string    `json:"owner"`
	Tier      string    `json:"tier"`
	ExpiresAt time.Time `json:"expires_at"`
	AutoRenew bool      `json:"auto_renew"`
}

// GetSubscription retrieves a user's subscription
func (k Keeper) GetSubscription(ctx sdk.Context, owner string) (Subscription, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetSubscriptionKey(owner)

	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return Subscription{}, false
	}

	var sub Subscription
	if err := json.Unmarshal(bz, &sub); err != nil {
		return Subscription{}, false
	}

	return sub, true
}

// SetSubscription stores a user's subscription
func (k Keeper) SetSubscription(ctx sdk.Context, sub Subscription) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.GetSubscriptionKey(sub.Owner)

	bz, err := json.Marshal(sub)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription: %w", err)
	}

	return store.Set(key, bz)
}

// SubscribeUser creates or upgrades a user's subscription
func (k Keeper) SubscribeUser(ctx sdk.Context, owner, tierName string, autoRenew bool) error {
	tier, found := types.GetSubscriptionTier(tierName)
	if !found {
		return types.ErrInvalidSubscription
	}

	// Charge subscription fee
	ownerAcc, _ := sdk.AccAddressFromBech32(owner)
	balance := k.bankKeeper.GetBalance(ctx, ownerAcc, "uhodl")

	if balance.Amount.LT(tier.MonthlyFeeHODL) {
		return fmt.Errorf("insufficient balance for subscription: need %s uhodl", tier.MonthlyFeeHODL.String())
	}

	// Transfer fee to module account (in production, this would go to a fee collector)
	// For now we just burn it
	// k.bankKeeper.SendCoinsFromAccountToModule(ctx, ownerAcc, types.ModuleName, sdk.NewCoins(sdk.NewCoin("uhodl", tier.MonthlyFeeHODL)))

	// Create/update subscription
	sub := Subscription{
		Owner:     owner,
		Tier:      tierName,
		ExpiresAt: ctx.BlockTime().AddDate(0, 1, 0), // 1 month
		AutoRenew: autoRenew,
	}

	if err := k.SetSubscription(ctx, sub); err != nil {
		return err
	}

	// Update all user's agents with new tier limits
	agents := k.GetAgentsByOwner(ctx, owner)
	for _, agent := range agents {
		agent.SubscriptionTier = tierName
		agent.ExpiresAt = sub.ExpiresAt
		agent.RateLimits.MaxActionsPerDay = tier.MaxActionsPerDay
		agent.SpendingLimits.DailyLimit = tier.MaxDailySpend

		// Update permissions
		agent.Permissions = []types.Permission{}
		for _, module := range tier.AllowedModules {
			agent.Permissions = append(agent.Permissions, types.Permission{
				Module:    module,
				Actions:   []string{"*"},
				MaxAmount: tier.MaxDailySpend,
			})
		}

		k.SetAgent(ctx, agent)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"subscription_updated",
			sdk.NewAttribute("owner", owner),
			sdk.NewAttribute("tier", tierName),
			sdk.NewAttribute("expires_at", sub.ExpiresAt.String()),
		),
	)

	return nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateAgentKey generates a random API key for an agent
func generateAgentKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "agent_" + hex.EncodeToString(bytes), nil
}
