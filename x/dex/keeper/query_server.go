package keeper

import (
	"context"
	"strconv"
	"time"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/dex/types"
)

// queryServer implements the QueryServer interface
type queryServer struct {
	keeper Keeper
}

// NewQueryServerImpl creates a new query server implementation
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{keeper: keeper}
}

// ===== ORDER QUERIES =====

// GetOrder returns order information by ID
func (q queryServer) GetOrder(goCtx context.Context, req *types.QueryGetOrderRequest) (*types.QueryGetOrderResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidOrderID, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Parse order ID
	orderID, err := strconv.ParseUint(req.OrderId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(types.ErrInvalidOrderID, "invalid order ID format")
	}

	// Get order from store
	order, found := q.keeper.GetOrder(ctx, orderID)
	if !found {
		return nil, errors.Wrap(types.ErrOrderNotFound, "order not found")
	}

	return &types.QueryGetOrderResponse{
		Order: order,
	}, nil
}

// GetUserOrders returns orders for a specific user with filtering and pagination
func (q queryServer) GetUserOrders(goCtx context.Context, req *types.QueryGetUserOrdersRequest) (*types.QueryGetUserOrdersResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrUnauthorized, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate user address
	userAddr, err := sdk.AccAddressFromBech32(req.User)
	if err != nil {
		return nil, errors.Wrap(types.ErrUnauthorized, "invalid user address")
	}

	// Set default pagination
	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Limit > 500 {
		req.Limit = 500 // Cap at 500
	}

	// Get user orders (simplified implementation)
	allOrders := q.keeper.getAllUserOrders(ctx, userAddr.String())
	
	// Filter by status if provided
	var filteredOrders []types.Order
	for _, order := range allOrders {
		if req.Status != "" {
			statusStr := order.Status.String()
			if statusStr != req.Status {
				continue
			}
		}
		
		// Filter by market if provided
		if req.MarketSymbol != "" && order.MarketSymbol != req.MarketSymbol {
			continue
		}
		
		filteredOrders = append(filteredOrders, order)
	}

	// Apply pagination
	start := req.Offset
	if start >= uint64(len(filteredOrders)) {
		start = uint64(len(filteredOrders))
	}

	end := start + req.Limit
	if end > uint64(len(filteredOrders)) {
		end = uint64(len(filteredOrders))
	}

	pagedOrders := filteredOrders[start:end]

	return &types.QueryGetUserOrdersResponse{
		Orders: pagedOrders,
		Pagination: types.PaginationResponse{
			Total:  uint64(len(filteredOrders)),
			Limit:  req.Limit,
			Offset: req.Offset,
		},
	}, nil
}

// ===== MARKET QUERIES =====

// GetMarket returns market information
func (q queryServer) GetMarket(goCtx context.Context, req *types.QueryGetMarketRequest) (*types.QueryGetMarketResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidMarket, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get market from store
	market, found := q.keeper.GetMarket(ctx, req.BaseSymbol, req.QuoteSymbol)
	if !found {
		return nil, errors.Wrap(types.ErrMarketNotFound, "market not found")
	}

	return &types.QueryGetMarketResponse{
		Market: market,
	}, nil
}

// GetMarkets returns all markets with pagination
func (q queryServer) GetMarkets(goCtx context.Context, req *types.QueryGetMarketsRequest) (*types.QueryGetMarketsResponse, error) {
	if req == nil {
		req = &types.QueryGetMarketsRequest{}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Set default pagination
	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Limit > 500 {
		req.Limit = 500
	}

	// Get all markets (simplified implementation)
	allMarkets := q.keeper.getAllMarkets(ctx)

	// Apply pagination
	start := req.Offset
	if start >= uint64(len(allMarkets)) {
		start = uint64(len(allMarkets))
	}

	end := start + req.Limit
	if end > uint64(len(allMarkets)) {
		end = uint64(len(allMarkets))
	}

	pagedMarkets := allMarkets[start:end]

	return &types.QueryGetMarketsResponse{
		Markets: pagedMarkets,
		Pagination: types.PaginationResponse{
			Total:  uint64(len(allMarkets)),
			Limit:  req.Limit,
			Offset: req.Offset,
		},
	}, nil
}

// GetOrderBook returns order book data
func (q queryServer) GetOrderBook(goCtx context.Context, req *types.QueryGetOrderBookRequest) (*types.QueryGetOrderBookResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidMarket, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get order book
	orderBook := q.keeper.GetOrderBook(ctx, req.BaseSymbol, req.QuoteSymbol)

	// Apply depth limit if specified
	if req.Depth > 0 {
		if len(orderBook.Bids) > int(req.Depth) {
			orderBook.Bids = orderBook.Bids[:req.Depth]
		}
		if len(orderBook.Asks) > int(req.Depth) {
			orderBook.Asks = orderBook.Asks[:req.Depth]
		}
	}

	return &types.QueryGetOrderBookResponse{
		OrderBook: orderBook,
	}, nil
}

// ===== TRADE QUERIES =====

// GetTrade returns trade information by ID
func (q queryServer) GetTrade(goCtx context.Context, req *types.QueryGetTradeRequest) (*types.QueryGetTradeResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidOrderID, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Parse trade ID
	tradeID, err := strconv.ParseUint(req.TradeId, 10, 64)
	if err != nil {
		return nil, errors.Wrap(types.ErrInvalidOrderID, "invalid trade ID format")
	}

	// Get trade from store
	trade, found := q.keeper.GetTrade(ctx, tradeID)
	if !found {
		return nil, errors.Wrap(types.ErrOrderNotFound, "trade not found")
	}

	return &types.QueryGetTradeResponse{
		Trade: trade,
	}, nil
}

// GetMarketTrades returns trades for a specific market with filtering and pagination
func (q queryServer) GetMarketTrades(goCtx context.Context, req *types.QueryGetMarketTradesRequest) (*types.QueryGetMarketTradesResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidMarket, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Set default pagination
	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Limit > 500 {
		req.Limit = 500
	}

	// Get market trades (simplified implementation)
	marketSymbol := req.BaseSymbol + "/" + req.QuoteSymbol
	allTrades := q.keeper.getMarketTrades(ctx, marketSymbol)

	// Filter by time range if provided
	var filteredTrades []types.Trade
	for _, trade := range allTrades {
		if req.FromTime > 0 && trade.ExecutedAt.Unix() < req.FromTime {
			continue
		}
		if req.ToTime > 0 && trade.ExecutedAt.Unix() > req.ToTime {
			continue
		}
		filteredTrades = append(filteredTrades, trade)
	}

	// Apply pagination
	start := req.Offset
	if start >= uint64(len(filteredTrades)) {
		start = uint64(len(filteredTrades))
	}

	end := start + req.Limit
	if end > uint64(len(filteredTrades)) {
		end = uint64(len(filteredTrades))
	}

	pagedTrades := filteredTrades[start:end]

	return &types.QueryGetMarketTradesResponse{
		Trades: pagedTrades,
		Pagination: types.PaginationResponse{
			Total:  uint64(len(filteredTrades)),
			Limit:  req.Limit,
			Offset: req.Offset,
		},
	}, nil
}

// ===== ATOMIC SWAP QUERIES =====

// GetExchangeRate returns the current exchange rate between two assets
func (q queryServer) GetExchangeRate(goCtx context.Context, req *types.QueryGetExchangeRateRequest) (*types.QueryGetExchangeRateResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrAssetNotFound, "empty request")
	}

	if req.FromSymbol == "" || req.ToSymbol == "" {
		return nil, errors.Wrap(types.ErrAssetNotFound, "missing asset symbols")
	}

	if req.FromSymbol == req.ToSymbol {
		return nil, errors.Wrap(types.ErrSameAssetSwap, "cannot get rate for same asset")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get current exchange rate
	exchangeRate, err := q.keeper.GetAtomicSwapRate(ctx, req.FromSymbol, req.ToSymbol)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get exchange rate")
	}

	return &types.QueryGetExchangeRateResponse{
		FromSymbol:   req.FromSymbol,
		ToSymbol:     req.ToSymbol,
		ExchangeRate: exchangeRate,
		Timestamp:    time.Now().Unix(),
	}, nil
}

// GetSwapHistory returns atomic swap history for a user
func (q queryServer) GetSwapHistory(goCtx context.Context, req *types.QueryGetSwapHistoryRequest) (*types.QueryGetSwapHistoryResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrUnauthorized, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate user address
	userAddr, err := sdk.AccAddressFromBech32(req.User)
	if err != nil {
		return nil, errors.Wrap(types.ErrUnauthorized, "invalid user address")
	}

	// Set default pagination
	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Limit > 500 {
		req.Limit = 500
	}

	// Get user's atomic swap records
	swapRecords := q.keeper.getUserAtomicSwaps(ctx, userAddr.String())

	// Filter by asset symbols if provided
	var filteredSwaps []types.AtomicSwapRecord
	for _, swap := range swapRecords {
		if req.FromSymbol != "" && swap.FromSymbol != req.FromSymbol {
			continue
		}
		if req.ToSymbol != "" && swap.ToSymbol != req.ToSymbol {
			continue
		}
		if req.FromTime > 0 && swap.ExecutedAt.Unix() < req.FromTime {
			continue
		}
		if req.ToTime > 0 && swap.ExecutedAt.Unix() > req.ToTime {
			continue
		}
		filteredSwaps = append(filteredSwaps, swap)
	}

	// Apply pagination
	start := req.Offset
	if start >= uint64(len(filteredSwaps)) {
		start = uint64(len(filteredSwaps))
	}

	end := start + req.Limit
	if end > uint64(len(filteredSwaps)) {
		end = uint64(len(filteredSwaps))
	}

	pagedSwaps := filteredSwaps[start:end]

	return &types.QueryGetSwapHistoryResponse{
		Swaps: pagedSwaps,
		Pagination: types.PaginationResponse{
			Total:  uint64(len(filteredSwaps)),
			Limit:  req.Limit,
			Offset: req.Offset,
		},
	}, nil
}

// ===== TRADING STRATEGY QUERIES =====

// GetTradingStrategy returns a trading strategy by ID
func (q queryServer) GetTradingStrategy(goCtx context.Context, req *types.QueryGetTradingStrategyRequest) (*types.QueryGetTradingStrategyResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrInvalidStrategyID, "empty request")
	}

	if req.StrategyId == "" {
		return nil, errors.Wrap(types.ErrInvalidStrategyID, "missing strategy ID")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get strategy from store
	strategy, found := q.keeper.GetTradingStrategy(ctx, req.StrategyId)
	if !found {
		return nil, errors.Wrap(types.ErrInvalidStrategyID, "strategy not found")
	}

	return &types.QueryGetTradingStrategyResponse{
		Strategy: strategy,
	}, nil
}

// GetUserStrategies returns all trading strategies for a user
func (q queryServer) GetUserStrategies(goCtx context.Context, req *types.QueryGetUserStrategiesRequest) (*types.QueryGetUserStrategiesResponse, error) {
	if req == nil {
		return nil, errors.Wrap(types.ErrUnauthorized, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate user address
	userAddr, err := sdk.AccAddressFromBech32(req.User)
	if err != nil {
		return nil, errors.Wrap(types.ErrUnauthorized, "invalid user address")
	}

	// Set default pagination
	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Limit > 500 {
		req.Limit = 500
	}

	// Get user's strategies
	userStrategies := q.keeper.getUserTradingStrategies(ctx, userAddr.String())

	// Filter by active status if provided
	var filteredStrategies []types.TradingStrategy
	for _, strategy := range userStrategies {
		if req.Active && !strategy.IsActive {
			continue
		}
		filteredStrategies = append(filteredStrategies, strategy)
	}

	// Apply pagination
	start := req.Offset
	if start >= uint64(len(filteredStrategies)) {
		start = uint64(len(filteredStrategies))
	}

	end := start + req.Limit
	if end > uint64(len(filteredStrategies)) {
		end = uint64(len(filteredStrategies))
	}

	pagedStrategies := filteredStrategies[start:end]

	return &types.QueryGetUserStrategiesResponse{
		Strategies: pagedStrategies,
		Pagination: types.PaginationResponse{
			Total:  uint64(len(filteredStrategies)),
			Limit:  req.Limit,
			Offset: req.Offset,
		},
	}, nil
}

// ===== HELPER METHODS =====

// getAllMarkets returns all markets (simplified implementation)
func (k Keeper) getAllMarkets(ctx sdk.Context) []types.Market {
	// In a full implementation, this would iterate through the store
	// For now, return a simplified list of supported markets
	return []types.Market{
		{
			BaseSymbol:   "APPLE",
			QuoteSymbol:  "HODL",
			BaseAssetID:  1,
			QuoteAssetID: 0,
			Active:       true,
			LastPrice:    k.GetAssetPrice(ctx, "APPLE"),
		},
		{
			BaseSymbol:   "TSLA", 
			QuoteSymbol:  "HODL",
			BaseAssetID:  2,
			QuoteAssetID: 0,
			Active:       true,
			LastPrice:    k.GetAssetPrice(ctx, "TSLA"),
		},
	}
}

// getMarketTrades returns trades for a market (simplified implementation)
func (k Keeper) getMarketTrades(ctx sdk.Context, marketSymbol string) []types.Trade {
	// In a full implementation, this would query trades by market from store
	// For now, return empty slice
	return []types.Trade{}
}

// getUserAtomicSwaps returns atomic swap records for a user (simplified implementation)
func (k Keeper) getUserAtomicSwaps(ctx sdk.Context, userAddr string) []types.AtomicSwapRecord {
	// In a full implementation, this would query swap records from store
	// For now, return empty slice
	return []types.AtomicSwapRecord{}
}

// getUserTradingStrategies returns trading strategies for a user (simplified implementation)
func (k Keeper) getUserTradingStrategies(ctx sdk.Context, userAddr string) []types.TradingStrategy {
	// In a full implementation, this would query strategies from store
	// For now, return empty slice
	return []types.TradingStrategy{}
}