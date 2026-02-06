package keeper

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/dex/types"
)

// MatchingEngine handles order matching with price-time priority
type MatchingEngine struct {
	keeper *Keeper
}

// NewMatchingEngine creates a new matching engine
func NewMatchingEngine(k *Keeper) *MatchingEngine {
	return &MatchingEngine{keeper: k}
}

// OrderBookEntry represents an order in the order book
type OrderBookEntry struct {
	OrderID   uint64
	Price     math.LegacyDec
	Quantity  math.Int
	Timestamp time.Time
	User      string
}

// GetBuyOrders returns all open buy orders for a market, sorted by price descending (best bid first)
// PERFORMANCE FIX: Uses market-specific index for O(log n) instead of O(n) full table scan
func (k Keeper) GetBuyOrders(ctx sdk.Context, baseSymbol, quoteSymbol string) []types.Order {
	store := ctx.KVStore(k.storeKey)

	// Use market-specific prefix for efficient lookups
	marketPrefix := types.GetMarketOrderPrefix(baseSymbol, quoteSymbol, types.OrderSideBuy)
	iterator := prefix.NewStore(store, marketPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var orders []types.Order
	for ; iterator.Valid(); iterator.Next() {
		var order types.Order
		if err := json.Unmarshal(iterator.Value(), &order); err != nil {
			continue
		}
		// Only include fillable orders
		if order.IsFillable() {
			orders = append(orders, order)
		}
	}

	// Sort by price descending (best bid first), then by time ascending
	sort.Slice(orders, func(i, j int) bool {
		if orders[i].Price.Equal(orders[j].Price) {
			return orders[i].CreatedAt.Before(orders[j].CreatedAt)
		}
		return orders[i].Price.GT(orders[j].Price)
	})

	return orders
}

// GetSellOrders returns all open sell orders for a market, sorted by price ascending (best ask first)
// PERFORMANCE FIX: Uses market-specific index for O(log n) instead of O(n) full table scan
func (k Keeper) GetSellOrders(ctx sdk.Context, baseSymbol, quoteSymbol string) []types.Order {
	store := ctx.KVStore(k.storeKey)

	// Use market-specific prefix for efficient lookups
	marketPrefix := types.GetMarketOrderPrefix(baseSymbol, quoteSymbol, types.OrderSideSell)
	iterator := prefix.NewStore(store, marketPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var orders []types.Order
	for ; iterator.Valid(); iterator.Next() {
		var order types.Order
		if err := json.Unmarshal(iterator.Value(), &order); err != nil {
			continue
		}
		// Only include fillable orders
		if order.IsFillable() {
			orders = append(orders, order)
		}
	}

	// Sort by price ascending (best ask first), then by time ascending
	sort.Slice(orders, func(i, j int) bool {
		if orders[i].Price.Equal(orders[j].Price) {
			return orders[i].CreatedAt.Before(orders[j].CreatedAt)
		}
		return orders[i].Price.LT(orders[j].Price)
	})

	return orders
}

// MatchOrder attempts to match an incoming order against the order book
// Returns list of trades executed and the remaining unfilled portion of the order
func (k Keeper) MatchOrder(ctx sdk.Context, incomingOrder types.Order) ([]types.Trade, types.Order, error) {
	var trades []types.Trade
	remainingQty := incomingOrder.RemainingQuantity
	if remainingQty.IsNil() {
		remainingQty = incomingOrder.Quantity
	}

	// Check if trading is halted for this equity (before attempting to match)
	if k.equityKeeper != nil && incomingOrder.BaseSymbol != "HODL" && incomingOrder.QuoteSymbol != "HODL" {
		companyID, found := k.GetCompanyIDBySymbol(ctx, incomingOrder.BaseSymbol)
		if found {
			// SECURITY: Check trading halt
			if k.equityKeeper.IsTradingHalted(ctx, companyID) {
				k.Logger(ctx).Info("skipping order matching - trading halted",
					"market", incomingOrder.MarketSymbol,
					"company_id", companyID,
					"order_id", incomingOrder.ID,
				)
				// Return order unchanged, no trades executed
				return trades, incomingOrder, nil
			}

			// SECURITY: Check if order user is blacklisted from this company
			if k.equityKeeper.IsBlacklisted(ctx, companyID, incomingOrder.User) {
				k.Logger(ctx).Info("rejecting order - user is blacklisted",
					"market", incomingOrder.MarketSymbol,
					"company_id", companyID,
					"order_id", incomingOrder.ID,
					"user", incomingOrder.User,
				)
				return nil, incomingOrder, types.ErrUnauthorized
			}
		}
	}

	// Get market for fee calculation
	market, found := k.GetMarket(ctx, incomingOrder.BaseSymbol, incomingOrder.QuoteSymbol)
	if !found {
		return nil, incomingOrder, types.ErrMarketNotFound
	}

	// Get opposite side orders
	var oppositeOrders []types.Order
	if incomingOrder.Side == types.OrderSideBuy {
		oppositeOrders = k.GetSellOrders(ctx, incomingOrder.BaseSymbol, incomingOrder.QuoteSymbol)
	} else {
		oppositeOrders = k.GetBuyOrders(ctx, incomingOrder.BaseSymbol, incomingOrder.QuoteSymbol)
	}

	// Track total filled for average price calculation
	totalFilled := math.ZeroInt()
	totalValue := math.LegacyZeroDec()

	// Match against existing orders
	for _, existingOrder := range oppositeOrders {
		if remainingQty.IsZero() {
			break
		}

		// Check price compatibility
		if !k.isPriceCompatible(incomingOrder, existingOrder) {
			// For limit orders, stop when prices no longer match
			if incomingOrder.Type == types.OrderTypeLimit {
				break
			}
			continue
		}

		// Calculate fill quantity (minimum of both remaining quantities)
		existingRemaining := existingOrder.RemainingQuantity
		if existingRemaining.IsNil() || existingRemaining.IsZero() {
			existingRemaining = existingOrder.Quantity.Sub(existingOrder.FilledQuantity)
		}

		fillQty := math.MinInt(remainingQty, existingRemaining)
		if fillQty.IsZero() {
			continue
		}

		// Determine execution price (maker's price has priority)
		execPrice := existingOrder.Price

		// Execute the trade
		trade, err := k.executeTrade(ctx, incomingOrder, existingOrder, fillQty, execPrice, market)
		if err != nil {
			k.Logger(ctx).Error("trade execution failed", "error", err)
			continue
		}

		trades = append(trades, trade)
		remainingQty = remainingQty.Sub(fillQty)
		totalFilled = totalFilled.Add(fillQty)
		totalValue = totalValue.Add(execPrice.MulInt(fillQty))

		// Update existing order
		existingOrder.FilledQuantity = existingOrder.FilledQuantity.Add(fillQty)
		existingOrder.RemainingQuantity = existingOrder.Quantity.Sub(existingOrder.FilledQuantity)
		existingOrder.UpdatedAt = ctx.BlockTime()

		if existingOrder.RemainingQuantity.IsZero() {
			existingOrder.Status = types.OrderStatusFilled
		} else {
			existingOrder.Status = types.OrderStatusPartiallyFilled
		}
		k.SetOrder(ctx, existingOrder)

		// Update beneficial owner registry for sell orders
		// When a sell order is filled (partially or fully), update the beneficial owner tracking
		if existingOrder.Side == types.OrderSideSell && k.equityKeeper != nil {
			companyID, found := k.GetCompanyIDBySymbol(ctx, existingOrder.BaseSymbol)
			if found {
				if existingOrder.RemainingQuantity.IsZero() {
					// Fully filled - unregister beneficial owner
					err := k.equityKeeper.UnregisterBeneficialOwner(
						ctx,
						types.ModuleName,
						companyID,
						"COMMON",
						existingOrder.User,
						existingOrder.ID,
					)
					if err != nil {
						k.Logger(ctx).Error("failed to unregister beneficial owner for filled order",
							"order_id", existingOrder.ID,
							"error", err,
						)
					} else {
						k.Logger(ctx).Info("unregistered beneficial owner for filled order",
							"order_id", existingOrder.ID,
						)
					}
				} else {
					// Partially filled - update remaining shares
					err := k.equityKeeper.UpdateBeneficialOwnerShares(
						ctx,
						types.ModuleName,
						companyID,
						"COMMON",
						existingOrder.User,
						existingOrder.ID,
						existingOrder.RemainingQuantity,
					)
					if err != nil {
						k.Logger(ctx).Error("failed to update beneficial owner shares",
							"order_id", existingOrder.ID,
							"remaining", existingOrder.RemainingQuantity.String(),
							"error", err,
						)
					} else {
						k.Logger(ctx).Info("updated beneficial owner shares for partial fill",
							"order_id", existingOrder.ID,
							"remaining", existingOrder.RemainingQuantity.String(),
						)
					}
				}
			}
		}
	}

	// Update incoming order
	incomingOrder.FilledQuantity = totalFilled
	incomingOrder.RemainingQuantity = remainingQty

	if totalFilled.IsPositive() {
		incomingOrder.AveragePrice = totalValue.QuoInt(totalFilled)
	}

	if remainingQty.IsZero() {
		incomingOrder.Status = types.OrderStatusFilled
	} else if totalFilled.IsPositive() {
		incomingOrder.Status = types.OrderStatusPartiallyFilled
	}
	incomingOrder.UpdatedAt = ctx.BlockTime()

	// Handle unfilled portion based on order type
	if !remainingQty.IsZero() {
		switch incomingOrder.TimeInForce {
		case types.TimeInForceIOC:
			// Immediate-or-Cancel: cancel remaining
			incomingOrder.Status = types.OrderStatusCancelled
			if totalFilled.IsPositive() {
				incomingOrder.Status = types.OrderStatusPartiallyFilled
			}
		case types.TimeInForceFOK:
			// Fill-or-Kill: if not fully filled, cancel entire order
			if totalFilled.IsZero() {
				incomingOrder.Status = types.OrderStatusCancelled
				return nil, incomingOrder, nil
			}
			// Shouldn't reach here for FOK, but handle gracefully
		default:
			// GTC/GTD: keep order open for remaining quantity
			if incomingOrder.Status != types.OrderStatusFilled {
				incomingOrder.Status = types.OrderStatusOpen
			}
		}
	}

	k.SetOrder(ctx, incomingOrder)

	// Update beneficial owner tracking for incoming sell order
	// This handles the case where a sell order was placed and matched
	if incomingOrder.Side == types.OrderSideSell && k.equityKeeper != nil && totalFilled.IsPositive() {
		companyID, found := k.GetCompanyIDBySymbol(ctx, incomingOrder.BaseSymbol)
		if found {
			if incomingOrder.Status == types.OrderStatusFilled {
				// Fully filled - unregister beneficial owner
				err := k.equityKeeper.UnregisterBeneficialOwner(
					ctx,
					types.ModuleName,
					companyID,
					"COMMON",
					incomingOrder.User,
					incomingOrder.ID,
				)
				if err != nil {
					k.Logger(ctx).Error("failed to unregister beneficial owner for incoming filled order",
						"order_id", incomingOrder.ID,
						"error", err,
					)
				}
			} else if incomingOrder.Status == types.OrderStatusPartiallyFilled {
				// Partially filled - update remaining shares
				err := k.equityKeeper.UpdateBeneficialOwnerShares(
					ctx,
					types.ModuleName,
					companyID,
					"COMMON",
					incomingOrder.User,
					incomingOrder.ID,
					incomingOrder.RemainingQuantity,
				)
				if err != nil {
					k.Logger(ctx).Error("failed to update beneficial owner shares for incoming order",
						"order_id", incomingOrder.ID,
						"remaining", incomingOrder.RemainingQuantity.String(),
						"error", err,
					)
				}
			} else if incomingOrder.Status == types.OrderStatusCancelled && totalFilled.IsPositive() {
				// IOC order that was partially filled then cancelled - unregister
				err := k.equityKeeper.UnregisterBeneficialOwner(
					ctx,
					types.ModuleName,
					companyID,
					"COMMON",
					incomingOrder.User,
					incomingOrder.ID,
				)
				if err != nil {
					k.Logger(ctx).Error("failed to unregister beneficial owner for cancelled IOC order",
						"order_id", incomingOrder.ID,
						"error", err,
					)
				}
			}
		}
	}

	return trades, incomingOrder, nil
}

// isPriceCompatible checks if two orders can be matched based on price
func (k Keeper) isPriceCompatible(incoming, existing types.Order) bool {
	// Market orders match any price
	if incoming.Type == types.OrderTypeMarket {
		return true
	}

	// For limit orders, check price compatibility
	if incoming.Side == types.OrderSideBuy {
		// Buy order matches if bid >= ask
		return incoming.Price.GTE(existing.Price)
	} else {
		// Sell order matches if ask <= bid
		return incoming.Price.LTE(existing.Price)
	}
}

// executeTrade executes a trade between two orders
func (k Keeper) executeTrade(
	ctx sdk.Context,
	takerOrder, makerOrder types.Order,
	quantity math.Int,
	price math.LegacyDec,
	market types.Market,
) (types.Trade, error) {
	tradeID := k.GetNextTradeID(ctx)
	tradeValue := price.MulInt(quantity)

	// Determine buyer and seller
	var buyerAddr, sellerAddr string
	var buyOrderID, sellOrderID uint64

	if takerOrder.Side == types.OrderSideBuy {
		buyerAddr = takerOrder.User
		sellerAddr = makerOrder.User
		buyOrderID = takerOrder.ID
		sellOrderID = makerOrder.ID
	} else {
		buyerAddr = makerOrder.User
		sellerAddr = takerOrder.User
		buyOrderID = makerOrder.ID
		sellOrderID = takerOrder.ID
	}

	// Calculate fees (taker pays taker fee, maker pays maker fee)
	takerFee := market.TakerFee.Mul(tradeValue)
	makerFee := market.MakerFee.Mul(tradeValue)

	var buyerFee, sellerFee math.LegacyDec
	if takerOrder.Side == types.OrderSideBuy {
		buyerFee = takerFee
		sellerFee = makerFee
	} else {
		buyerFee = makerFee
		sellerFee = takerFee
	}

	// Transfer assets
	buyer, err := sdk.AccAddressFromBech32(buyerAddr)
	if err != nil {
		return types.Trade{}, err
	}
	seller, err := sdk.AccAddressFromBech32(sellerAddr)
	if err != nil {
		return types.Trade{}, err
	}

	// SECURITY FIX: Use cache context for atomic operations
	// If any transfer fails, the entire operation is rolled back
	cacheCtx, writeCache := ctx.CacheContext()

	// Transfer base asset (equity) from seller to buyer
	baseCoins := sdk.NewCoins(sdk.NewCoin(takerOrder.BaseSymbol, quantity))
	if err := k.bankKeeper.SendCoins(cacheCtx, seller, buyer, baseCoins); err != nil {
		return types.Trade{}, fmt.Errorf("failed to transfer base asset: %w", err)
	}

	// Calculate quote amount (buyer pays price * quantity)
	quoteAmount := price.MulInt(quantity).TruncateInt()
	quoteCoins := sdk.NewCoins(sdk.NewCoin(takerOrder.QuoteSymbol, quoteAmount))
	if err := k.bankKeeper.SendCoins(cacheCtx, buyer, seller, quoteCoins); err != nil {
		// No manual rollback needed - cache context handles it
		return types.Trade{}, fmt.Errorf("failed to transfer quote asset: %w", err)
	}

	// Commit the cache context - all transfers succeeded
	writeCache()

	// Collect fees to module account
	totalFeeAmount := buyerFee.Add(sellerFee).TruncateInt()
	if totalFeeAmount.IsPositive() {
		// Collect from buyer
		if buyerFee.TruncateInt().IsPositive() {
			buyerFeeCoins := sdk.NewCoins(sdk.NewCoin(takerOrder.QuoteSymbol, buyerFee.TruncateInt()))
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, buyer, types.ModuleName, buyerFeeCoins); err != nil {
				k.Logger(ctx).Error("failed to collect buyer fee", "error", err)
			}
		}
		// Collect from seller
		if sellerFee.TruncateInt().IsPositive() {
			sellerFeeCoins := sdk.NewCoins(sdk.NewCoin(takerOrder.QuoteSymbol, sellerFee.TruncateInt()))
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, seller, types.ModuleName, sellerFeeCoins); err != nil {
				k.Logger(ctx).Error("failed to collect seller fee", "error", err)
			}
		}
	}

	// Create trade record
	trade := types.NewTrade(
		tradeID,
		takerOrder.MarketSymbol,
		buyOrderID,
		sellOrderID,
		buyerAddr,
		sellerAddr,
		quantity,
		price,
		buyerFee,
		sellerFee,
		takerOrder.Side == types.OrderSideSell, // buyer is maker if taker is selling
	)

	// Store trade
	if err := k.SetTrade(ctx, trade); err != nil {
		return types.Trade{}, fmt.Errorf("failed to store trade: %w", err)
	}

	// Update market statistics
	k.UpdateMarketStats(ctx, market, trade)

	// Emit trade event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"trade_executed",
			sdk.NewAttribute("trade_id", fmt.Sprintf("%d", tradeID)),
			sdk.NewAttribute("market", takerOrder.MarketSymbol),
			sdk.NewAttribute("buyer", buyerAddr),
			sdk.NewAttribute("seller", sellerAddr),
			sdk.NewAttribute("quantity", quantity.String()),
			sdk.NewAttribute("price", price.String()),
			sdk.NewAttribute("value", tradeValue.String()),
			sdk.NewAttribute("buyer_fee", buyerFee.String()),
			sdk.NewAttribute("seller_fee", sellerFee.String()),
		),
	)

	return trade, nil
}

// ProcessMarketOrder processes a market order immediately
func (k Keeper) ProcessMarketOrder(ctx sdk.Context, order types.Order) ([]types.Trade, types.Order, error) {
	// Market orders execute immediately at best available price
	return k.MatchOrder(ctx, order)
}

// ProcessLimitOrder processes a limit order
func (k Keeper) ProcessLimitOrder(ctx sdk.Context, order types.Order) ([]types.Trade, types.Order, error) {
	// First try to match against existing orders
	trades, updatedOrder, err := k.MatchOrder(ctx, order)
	if err != nil {
		return nil, order, err
	}

	// If order not fully filled and is GTC/GTD, it stays in the book (already saved in MatchOrder)
	return trades, updatedOrder, nil
}

// ProcessStopOrder checks and processes stop orders when price triggers
func (k Keeper) ProcessStopOrder(ctx sdk.Context, order types.Order, currentPrice math.LegacyDec) ([]types.Trade, types.Order, error) {
	// Check if stop price has been triggered
	triggered := false

	if order.Side == types.OrderSideBuy {
		// Buy stop triggers when price rises above stop price
		triggered = currentPrice.GTE(order.StopPrice)
	} else {
		// Sell stop triggers when price falls below stop price
		triggered = currentPrice.LTE(order.StopPrice)
	}

	if !triggered {
		return nil, order, nil
	}

	// Convert to market or limit order based on type
	if order.Type == types.OrderTypeStop {
		order.Type = types.OrderTypeMarket
	} else if order.Type == types.OrderTypeStopLimit {
		order.Type = types.OrderTypeLimit
	}

	// Emit stop triggered event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"stop_order_triggered",
			sdk.NewAttribute("order_id", fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute("stop_price", order.StopPrice.String()),
			sdk.NewAttribute("trigger_price", currentPrice.String()),
		),
	)

	return k.MatchOrder(ctx, order)
}

// GetOrderBookAggregated returns aggregated order book levels
func (k Keeper) GetOrderBookAggregated(ctx sdk.Context, baseSymbol, quoteSymbol string, levels int) types.OrderBook {
	buyOrders := k.GetBuyOrders(ctx, baseSymbol, quoteSymbol)
	sellOrders := k.GetSellOrders(ctx, baseSymbol, quoteSymbol)

	// Aggregate buy orders by price level
	bidLevels := k.aggregateOrdersByPrice(buyOrders, levels)
	// Aggregate sell orders by price level
	askLevels := k.aggregateOrdersByPrice(sellOrders, levels)

	return types.OrderBook{
		MarketSymbol: baseSymbol + "/" + quoteSymbol,
		Bids:         bidLevels,
		Asks:         askLevels,
		LastUpdated:  ctx.BlockTime(),
	}
}

// aggregateOrdersByPrice aggregates orders by price level
func (k Keeper) aggregateOrdersByPrice(orders []types.Order, maxLevels int) []types.OrderBookLevel {
	if len(orders) == 0 {
		return []types.OrderBookLevel{}
	}

	// Group by price
	priceGroups := make(map[string]types.OrderBookLevel)
	for _, order := range orders {
		priceKey := order.Price.String()
		remaining := order.RemainingQuantity
		if remaining.IsNil() {
			remaining = order.Quantity.Sub(order.FilledQuantity)
		}

		if level, exists := priceGroups[priceKey]; exists {
			level.Quantity = level.Quantity.Add(remaining)
			level.OrderCount++
			priceGroups[priceKey] = level
		} else {
			priceGroups[priceKey] = types.OrderBookLevel{
				Price:      order.Price,
				Quantity:   remaining,
				OrderCount: 1,
			}
		}
	}

	// Convert to slice
	levels := make([]types.OrderBookLevel, 0, len(priceGroups))
	for _, level := range priceGroups {
		levels = append(levels, level)
	}

	// Limit to maxLevels
	if maxLevels > 0 && len(levels) > maxLevels {
		levels = levels[:maxLevels]
	}

	return levels
}

// CheckAndProcessStopOrders checks all stop orders and triggers those that meet conditions
func (k Keeper) CheckAndProcessStopOrders(ctx sdk.Context, baseSymbol, quoteSymbol string) {
	market, found := k.GetMarket(ctx, baseSymbol, quoteSymbol)
	if !found || market.LastPrice.IsZero() {
		return
	}

	currentPrice := market.LastPrice

	// Get all stop orders for this market
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.OrderPrefix).Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var order types.Order
		if err := json.Unmarshal(iterator.Value(), &order); err != nil {
			continue
		}

		// Only process stop orders for this market
		if order.BaseSymbol != baseSymbol || order.QuoteSymbol != quoteSymbol {
			continue
		}
		if order.Type != types.OrderTypeStop && order.Type != types.OrderTypeStopLimit {
			continue
		}
		if !order.IsFillable() {
			continue
		}

		// Process the stop order
		k.ProcessStopOrder(ctx, order, currentPrice)
	}
}

// CancelExpiredOrders cancels all expired GTD orders
func (k Keeper) CancelExpiredOrders(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.OrderPrefix).Iterator(nil, nil)
	defer iterator.Close()

	currentTime := ctx.BlockTime()

	for ; iterator.Valid(); iterator.Next() {
		var order types.Order
		if err := json.Unmarshal(iterator.Value(), &order); err != nil {
			continue
		}

		if !order.IsFillable() {
			continue
		}

		// Check if order has expired
		if order.TimeInForce == types.TimeInForceGTD && !order.ExpiresAt.IsZero() {
			if currentTime.After(order.ExpiresAt) {
				order.Status = types.OrderStatusExpired
				order.UpdatedAt = currentTime
				k.SetOrder(ctx, order)

				// Emit expiry event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeOrderExpired,
						sdk.NewAttribute("order_id", fmt.Sprintf("%d", order.ID)),
						sdk.NewAttribute("user", order.User),
						sdk.NewAttribute("market", order.MarketSymbol),
					),
				)
			}
		}
	}
}

// GetAllOpenOrders returns all open orders (for debugging/admin)
func (k Keeper) GetAllOpenOrders(ctx sdk.Context) []types.Order {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.OrderPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var orders []types.Order
	for ; iterator.Valid(); iterator.Next() {
		var order types.Order
		if err := json.Unmarshal(iterator.Value(), &order); err != nil {
			continue
		}
		if order.IsFillable() {
			orders = append(orders, order)
		}
	}

	return orders
}

// GetUserOpenOrders returns all open orders for a user
func (k Keeper) GetUserOpenOrders(ctx sdk.Context, user string) []types.Order {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.OrderPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var orders []types.Order
	for ; iterator.Valid(); iterator.Next() {
		var order types.Order
		if err := json.Unmarshal(iterator.Value(), &order); err != nil {
			continue
		}
		if order.User == user && order.IsFillable() {
			orders = append(orders, order)
		}
	}

	return orders
}
