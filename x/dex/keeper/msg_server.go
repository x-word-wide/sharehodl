package keeper

import (
	"context"
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