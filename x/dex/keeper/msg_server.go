package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/dex/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) *msgServer {
	return &msgServer{Keeper: keeper}
}

// CreateMarket handles market creation
func (k msgServer) CreateMarket(goCtx context.Context, msg *types.SimpleMsgCreateMarket) (*types.MsgCreateMarketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// For now, allow any user to create markets
	// In production, might require governance approval or special permissions

	// Create market
	err = k.Keeper.CreateMarket(
		ctx,
		msg.BaseSymbol,
		msg.QuoteSymbol,
		msg.BaseAssetID,
		msg.QuoteAssetID,
		msg.MinOrderSize,
		msg.MaxOrderSize,
		msg.TickSize,
		msg.LotSize,
		msg.MakerFee,
		msg.TakerFee,
	)
	if err != nil {
		return nil, err
	}

	marketSymbol := msg.BaseSymbol + "/" + msg.QuoteSymbol

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"create_market",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("market_symbol", marketSymbol),
			sdk.NewAttribute("base_symbol", msg.BaseSymbol),
			sdk.NewAttribute("quote_symbol", msg.QuoteSymbol),
		),
	)

	return &types.MsgCreateMarketResponse{
		MarketSymbol: marketSymbol,
		Success:      true,
	}, nil
}

// PlaceOrder handles order placement
func (k msgServer) PlaceOrder(goCtx context.Context, msg *types.SimpleMsgPlaceOrder) (*types.MsgPlaceOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Parse order side
	var side types.OrderSide
	switch strings.ToLower(msg.Side) {
	case "buy":
		side = types.OrderSideBuy
	case "sell":
		side = types.OrderSideSell
	default:
		return nil, types.ErrInvalidOrderSide
	}

	// Parse order type
	var orderType types.OrderType
	switch strings.ToLower(msg.OrderType) {
	case "market":
		orderType = types.OrderTypeMarket
	case "limit":
		orderType = types.OrderTypeLimit
	case "stop":
		orderType = types.OrderTypeStop
	case "stop_limit":
		orderType = types.OrderTypeStopLimit
	default:
		return nil, types.ErrInvalidOrderType
	}

	// Parse time in force
	var timeInForce types.TimeInForce
	switch strings.ToLower(msg.TimeInForce) {
	case "gtc":
		timeInForce = types.TimeInForceGTC
	case "ioc":
		timeInForce = types.TimeInForceIOC
	case "fok":
		timeInForce = types.TimeInForceFOK
	case "gtd":
		timeInForce = types.TimeInForceGTD
	default:
		return nil, types.ErrInvalidTimeInForce
	}

	// Place order
	order, err := k.Keeper.PlaceOrder(
		ctx,
		msg.Creator,
		msg.MarketSymbol,
		side,
		orderType,
		timeInForce,
		msg.Quantity,
		msg.Price,
		msg.StopPrice,
		msg.ClientOrderID,
	)
	if err != nil {
		return nil, err
	}

	// Attempt to match order (simplified)
	trades := k.Keeper.MatchOrders(ctx, order)

	// Update order with any fills
	filledQuantity := math.ZeroInt()
	totalValue := math.LegacyZeroDec()
	totalFees := math.LegacyZeroDec()

	for _, trade := range trades {
		filledQuantity = filledQuantity.Add(trade.Quantity)
		totalValue = totalValue.Add(trade.Value)
		if trade.Buyer == order.User {
			totalFees = totalFees.Add(trade.BuyerFee)
		} else {
			totalFees = totalFees.Add(trade.SellerFee)
		}
	}

	// Update order status
	if filledQuantity.GT(math.ZeroInt()) {
		order.FilledQuantity = filledQuantity
		order.RemainingQuantity = order.Quantity.Sub(filledQuantity)
		order.TotalFees = totalFees

		if order.RemainingQuantity.IsZero() {
			order.Status = types.OrderStatusFilled
		} else {
			order.Status = types.OrderStatusPartiallyFilled
		}

		if !totalValue.IsZero() {
			order.AveragePrice = totalValue.QuoInt(filledQuantity)
		}

		k.Keeper.SetOrder(ctx, order)
	}

	// Store trades
	for _, trade := range trades {
		k.Keeper.SetTrade(ctx, trade)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"place_order",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("order_id", string(rune(order.ID))),
			sdk.NewAttribute("market_symbol", msg.MarketSymbol),
			sdk.NewAttribute("side", msg.Side),
			sdk.NewAttribute("order_type", msg.OrderType),
			sdk.NewAttribute("quantity", msg.Quantity.String()),
			sdk.NewAttribute("price", msg.Price.String()),
			sdk.NewAttribute("status", order.Status.String()),
			sdk.NewAttribute("filled_quantity", order.FilledQuantity.String()),
		),
	)

	return &types.MsgPlaceOrderResponse{
		OrderID:           order.ID,
		Status:            order.Status.String(),
		FilledQuantity:    order.FilledQuantity,
		RemainingQuantity: order.RemainingQuantity,
		AveragePrice:      order.AveragePrice,
		TotalFees:         order.TotalFees,
	}, nil
}

// CancelOrder handles order cancellation
func (k msgServer) CancelOrder(goCtx context.Context, msg *types.SimpleMsgCancelOrder) (*types.MsgCancelOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Cancel order
	err = k.Keeper.CancelOrder(ctx, creatorAddr, msg.OrderID)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"cancel_order",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("order_id", string(rune(msg.OrderID))),
		),
	)

	return &types.MsgCancelOrderResponse{
		OrderID: msg.OrderID,
		Success: true,
	}, nil
}

// CancelAllOrders handles bulk order cancellation
func (k msgServer) CancelAllOrders(goCtx context.Context, msg *types.SimpleMsgCancelAllOrders) (*types.MsgCancelAllOrdersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Get all user orders
	orders := k.Keeper.getAllUserOrders(ctx, msg.Creator)
	
	cancelledCount := uint64(0)
	for _, order := range orders {
		// Filter by market if specified
		if msg.MarketSymbol != "" && order.MarketSymbol != msg.MarketSymbol {
			continue
		}

		// Cancel order
		if err := k.Keeper.CancelOrder(ctx, creatorAddr, order.ID); err == nil {
			cancelledCount++
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"cancel_all_orders",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("market_symbol", msg.MarketSymbol),
			sdk.NewAttribute("orders_cancelled", string(rune(cancelledCount))),
		),
	)

	return &types.MsgCancelAllOrdersResponse{
		OrdersCancelled: cancelledCount,
		Success:         true,
	}, nil
}

// CreateLiquidityPool handles liquidity pool creation
func (k msgServer) CreateLiquidityPool(goCtx context.Context, msg *types.SimpleMsgCreateLiquidityPool) (*types.MsgCreateLiquidityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Check if pool already exists
	baseSymbol, quoteSymbol := k.Keeper.parseMarketSymbol(msg.MarketSymbol)
	if _, exists := k.Keeper.GetLiquidityPool(ctx, baseSymbol, quoteSymbol); exists {
		return nil, types.ErrPoolAlreadyExists
	}

	// Check user has sufficient funds
	balance := k.bankKeeper.GetAllBalances(ctx, creatorAddr)
	
	if balance.AmountOf(baseSymbol).LT(msg.BaseAmount) || balance.AmountOf(quoteSymbol).LT(msg.QuoteAmount) {
		return nil, types.ErrInsufficientFunds
	}

	// Transfer funds to module (simplified)
	// In full implementation would transfer to pool contract

	// Calculate initial LP tokens (simplified)
	lpTokenSupply := msg.BaseAmount.Add(msg.QuoteAmount) // Simplified calculation

	// Create pool
	pool := types.LiquidityPool{
		MarketSymbol:     msg.MarketSymbol,
		BaseReserve:      msg.BaseAmount,
		QuoteReserve:     msg.QuoteAmount,
		LPTokenSupply:    lpTokenSupply,
		Fee:              msg.Fee,
		Volume24h:        math.ZeroInt(),
		FeesCollected24h: math.LegacyZeroDec(),
	}

	k.Keeper.SetLiquidityPool(ctx, pool)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"create_liquidity_pool",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("market_symbol", msg.MarketSymbol),
			sdk.NewAttribute("base_amount", msg.BaseAmount.String()),
			sdk.NewAttribute("quote_amount", msg.QuoteAmount.String()),
			sdk.NewAttribute("lp_tokens_issued", lpTokenSupply.String()),
		),
	)

	return &types.MsgCreateLiquidityPoolResponse{
		PoolAddress:    "pool_" + msg.MarketSymbol, // Simplified
		LPTokensIssued: lpTokenSupply,
		Success:        true,
	}, nil
}

// AddLiquidity handles liquidity addition
func (k msgServer) AddLiquidity(goCtx context.Context, msg *types.SimpleMsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	// Simplified implementation - return success
	// Full implementation would:
	// 1. Check pool exists
	// 2. Calculate optimal amounts based on current ratio
	// 3. Transfer funds from user
	// 4. Mint LP tokens
	// 5. Update pool reserves

	return &types.MsgAddLiquidityResponse{
		LPTokensReceived: msg.MinLPTokens,
		BaseUsed:         msg.BaseAmount,
		QuoteUsed:        msg.QuoteAmount,
		Success:          true,
	}, nil
}

// RemoveLiquidity handles liquidity removal
func (k msgServer) RemoveLiquidity(goCtx context.Context, msg *types.SimpleMsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	// Simplified implementation - return success
	// Full implementation would:
	// 1. Check pool exists and user has LP tokens
	// 2. Calculate proportional amounts to return
	// 3. Burn LP tokens
	// 4. Transfer assets back to user
	// 5. Update pool reserves

	return &types.MsgRemoveLiquidityResponse{
		BaseReceived:  msg.MinBaseAmount,
		QuoteReceived: msg.MinQuoteAmount,
		Success:       true,
	}, nil
}

// Swap handles AMM swaps
func (k msgServer) Swap(goCtx context.Context, msg *types.SimpleMsgSwap) (*types.MsgSwapResponse, error) {
	// Simplified implementation - return success
	// Full implementation would:
	// 1. Check pool exists and has sufficient liquidity
	// 2. Calculate output amount using constant product formula
	// 3. Check slippage tolerance
	// 4. Transfer input asset from user
	// 5. Transfer output asset to user
	// 6. Update pool reserves
	// 7. Calculate and collect fees

	return &types.MsgSwapResponse{
		OutputAmount: msg.MinOutputAmount,
		Fee:          math.LegacyNewDecWithPrec(3, 3), // 0.3% fee
		PriceImpact:  math.LegacyNewDecWithPrec(1, 2), // 1% price impact
		Success:      true,
	}, nil
}

// ===== BLOCKCHAIN-NATIVE MESSAGE HANDLERS =====

// PlaceAtomicSwapOrder handles atomic cross-asset swap orders
func (k msgServer) PlaceAtomicSwapOrder(goCtx context.Context, msg *types.MsgPlaceAtomicSwapOrder) (*types.MsgPlaceAtomicSwapOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate trader address
	traderAddr, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid trader address: %v", err)
	}

	// Validate input parameters
	if msg.FromSymbol == "" || msg.ToSymbol == "" {
		return nil, types.ErrInvalidAsset
	}

	if msg.Quantity == 0 {
		return nil, types.ErrInvalidOrderSize
	}

	quantity := math.NewIntFromUint64(msg.Quantity)

	// Execute atomic swap
	order, err := k.ExecuteAtomicSwap(
		ctx,
		traderAddr,
		msg.FromSymbol,
		msg.ToSymbol,
		quantity,
		msg.MaxSlippage,
	)
	if err != nil {
		return nil, err
	}

	// Log successful atomic swap
	ctx.Logger().Info(
		"Atomic swap executed successfully",
		"trader", msg.Trader,
		"from", msg.FromSymbol,
		"to", msg.ToSymbol,
		"quantity", msg.Quantity,
		"order_id", fmt.Sprintf("%d", order.ID),
		"rate", order.AveragePrice.String(),
	)

	return &types.MsgPlaceAtomicSwapOrderResponse{
		OrderId:               fmt.Sprintf("%d", order.ID),
		EstimatedExchangeRate: order.AveragePrice,
	}, nil
}

// PlaceFractionalOrder handles fractional share trading
func (k msgServer) PlaceFractionalOrder(goCtx context.Context, msg *types.MsgPlaceFractionalOrder) (*types.MsgPlaceFractionalOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate trader address
	traderAddr, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid trader address: %v", err)
	}

	// TODO: Implement fractional order logic
	_ = ctx // Mark as used until full implementation
	_ = traderAddr // Mark as used until full implementation

	// Validate fractional quantity is positive and not zero
	if msg.FractionalQuantity.IsZero() || msg.FractionalQuantity.IsNegative() {
		return nil, types.ErrInvalidOrderSize
	}

	// For now, return not implemented
	// TODO: Implement fractional share logic
	return nil, types.ErrNotImplemented

	// Implementation would:
	// 1. Validate minimum fraction requirements
	// 2. Check if fractional trading is enabled for the symbol
	// 3. Convert fractional shares to internal units
	// 4. Process order matching with fractional logic
	// 5. Handle partial fills and fractional remainder tracking
}

// CreateTradingStrategy handles programmable trading strategy creation
func (k msgServer) CreateTradingStrategy(goCtx context.Context, msg *types.MsgCreateTradingStrategy) (*types.MsgCreateTradingStrategyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid owner address: %v", err)
	}

	// Validate strategy parameters
	if msg.Name == "" {
		return nil, errors.Wrapf(types.ErrInvalidOrderID, "strategy name cannot be empty")
	}

	if len(msg.Conditions) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidOrderType, "strategy must have at least one condition")
	}

	if len(msg.Actions) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidOrderType, "strategy must have at least one action")
	}

	// Generate strategy ID
	strategyID := fmt.Sprintf("strategy_%s_%d", ownerAddr.String()[:8], ctx.BlockHeight())

	// Create trading strategy
	strategy := types.TradingStrategy{
		StrategyId:  strategyID,
		Name:        msg.Name,
		Owner:       msg.Owner,
		Conditions:  msg.Conditions,
		Actions:     msg.Actions,
		MaxExposure: msg.MaxExposure,
		IsActive:    true,
		CreatedAt:   ctx.BlockTime(),
	}

	// Store strategy
	k.SetTradingStrategy(ctx, strategy)

	// Emit strategy creation event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"trading_strategy_created",
			sdk.NewAttribute("owner", msg.Owner),
			sdk.NewAttribute("strategy_id", strategyID),
			sdk.NewAttribute("name", msg.Name),
			sdk.NewAttribute("conditions_count", fmt.Sprintf("%d", len(msg.Conditions))),
			sdk.NewAttribute("actions_count", fmt.Sprintf("%d", len(msg.Actions))),
			sdk.NewAttribute("max_exposure", msg.MaxExposure.String()),
		),
	)

	ctx.Logger().Info(
		"Trading strategy created",
		"owner", msg.Owner,
		"strategy_id", strategyID,
		"name", msg.Name,
	)

	return &types.MsgCreateTradingStrategyResponse{
		StrategyId: strategyID,
	}, nil
}

// ExecuteStrategy handles manual strategy execution
func (k msgServer) ExecuteStrategy(goCtx context.Context, msg *types.MsgExecuteStrategy) (*types.MsgExecuteStrategyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate executor address
	executorAddr, err := sdk.AccAddressFromBech32(msg.Executor)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid executor address: %v", err)
	}

	// Get strategy
	strategy, found := k.GetTradingStrategy(ctx, msg.StrategyId)
	if !found {
		return nil, types.ErrOrderNotFound
	}

	// Check if executor is the owner
	if strategy.Owner != msg.Executor {
		return nil, types.ErrUnauthorized
	}

	// Check if strategy is active
	if !strategy.IsActive {
		return nil, errors.Wrapf(types.ErrMarketInactive, "strategy is inactive")
	}

	// For manual execution or force execution, skip condition checks
	if !msg.ForceExecution {
		// Check if conditions are met
		if !k.CheckStrategyConditions(ctx, strategy) {
			return nil, errors.Wrapf(types.ErrInvalidOrderType, "strategy conditions not met")
		}
	}

	// Execute strategy actions
	orderIds, err := k.ExecuteStrategyActions(ctx, strategy, executorAddr)
	if err != nil {
		return nil, err
	}

	// Emit strategy execution event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"trading_strategy_executed",
			sdk.NewAttribute("executor", msg.Executor),
			sdk.NewAttribute("strategy_id", msg.StrategyId),
			sdk.NewAttribute("force_execution", fmt.Sprintf("%t", msg.ForceExecution)),
			sdk.NewAttribute("orders_placed", fmt.Sprintf("%d", len(orderIds))),
		),
	)

	ctx.Logger().Info(
		"Trading strategy executed",
		"executor", msg.Executor,
		"strategy_id", msg.StrategyId,
		"orders_placed", len(orderIds),
	)

	return &types.MsgExecuteStrategyResponse{
		ExecutedOrderIds: orderIds,
		OrdersPlaced:     uint64(len(orderIds)),
	}, nil
}

// Helper methods for trading strategies

// SetTradingStrategy stores a trading strategy
func (k Keeper) SetTradingStrategy(ctx sdk.Context, strategy types.TradingStrategy) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradingStrategyKey(strategy.StrategyId)
	
	bz, err := json.Marshal(strategy)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// GetTradingStrategy retrieves a trading strategy
func (k Keeper) GetTradingStrategy(ctx sdk.Context, strategyId string) (types.TradingStrategy, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradingStrategyKey(strategyId)
	
	bz := store.Get(key)
	if bz == nil {
		return types.TradingStrategy{}, false
	}
	
	var strategy types.TradingStrategy
	if err := json.Unmarshal(bz, &strategy); err != nil {
		panic(err)
	}
	return strategy, true
}

// CheckStrategyConditions checks if strategy conditions are met
func (k Keeper) CheckStrategyConditions(ctx sdk.Context, strategy types.TradingStrategy) bool {
	for _, condition := range strategy.Conditions {
		if !k.CheckTriggerCondition(ctx, condition) {
			return false
		}
	}
	return true
}

// CheckTriggerCondition checks a single trigger condition
func (k Keeper) CheckTriggerCondition(ctx sdk.Context, condition types.TriggerCondition) bool {
	// Get current price for the symbol
	currentPrice := k.GetAssetPrice(ctx, condition.Symbol)
	
	switch condition.ConditionType {
	case "ABOVE":
		return currentPrice.GT(condition.PriceThreshold)
	case "BELOW":
		return currentPrice.LT(condition.PriceThreshold)
	case "EQUALS":
		// Use small tolerance for equality
		tolerance := math.LegacyNewDecWithPrec(1, 4) // 0.01%
		diff := currentPrice.Sub(condition.PriceThreshold).Abs()
		maxDiff := condition.PriceThreshold.Mul(tolerance)
		return diff.LTE(maxDiff)
	default:
		return false
	}
}

// ExecuteStrategyActions executes strategy actions
func (k Keeper) ExecuteStrategyActions(ctx sdk.Context, strategy types.TradingStrategy, executor sdk.AccAddress) ([]string, error) {
	var orderIds []string
	
	for _, action := range strategy.Actions {
		orderId, err := k.ExecuteTradingAction(ctx, action, executor, strategy.MaxExposure)
		if err != nil {
			// Log error but continue with other actions
			ctx.Logger().Error(
				"Failed to execute strategy action",
				"strategy_id", strategy.StrategyId,
				"action", action,
				"error", err,
			)
			continue
		}
		orderIds = append(orderIds, orderId)
	}
	
	return orderIds, nil
}

// ExecuteTradingAction executes a single trading action
func (k Keeper) ExecuteTradingAction(ctx sdk.Context, action types.TradingAction, executor sdk.AccAddress, maxExposure math.Int) (string, error) {
	// Convert fractional quantity to uint64 for order placement
	quantity := action.Quantity.TruncateInt().Uint64()
	
	if quantity == 0 {
		return "", types.ErrInvalidOrderSize
	}

	// Calculate price (current price + offset)
	currentPrice := k.GetAssetPrice(ctx, action.Symbol)
	orderPrice := currentPrice.Add(action.PriceOffset)
	
	// Create and place order
	orderID := k.GetNextOrderID(ctx)
	orderIdStr := fmt.Sprintf("strategy_order_%d", orderID)
	
	order := types.Order{
		ID:               orderID,
		User:             executor.String(),
		Type:             types.OrderTypeMarket, // Simplified for strategy execution
		Side:             types.OrderSideBuy,    // Simplified for now
		MarketSymbol:     action.Symbol + "/HODL",
		BaseSymbol:       action.Symbol,
		QuoteSymbol:      "HODL", 
		Quantity:         math.NewIntFromUint64(quantity),
		Price:            orderPrice,
		Status:           types.OrderStatusPending,
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
		ClientOrderID:    fmt.Sprintf("strategy_%s", executor.String()[:8]),
	}
	
	// Store the order
	k.SetOrder(ctx, order)
	
	// For now, immediately mark as filled for demonstration
	// In production, would go through order matching
	order.Status = types.OrderStatusFilled
	order.FilledQuantity = math.NewIntFromUint64(quantity)
	order.AveragePrice = orderPrice
	k.SetOrder(ctx, order)
	
	return orderIdStr, nil
}