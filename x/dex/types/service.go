package types

import (
	"context"
)

// MsgServer defines the module's message service.
type MsgServer interface {
	CreateMarket(context.Context, *SimpleMsgCreateMarket) (*MsgCreateMarketResponse, error)
	PlaceOrder(context.Context, *SimpleMsgPlaceOrder) (*MsgPlaceOrderResponse, error)
	CancelOrder(context.Context, *SimpleMsgCancelOrder) (*MsgCancelOrderResponse, error)
	CancelAllOrders(context.Context, *SimpleMsgCancelAllOrders) (*MsgCancelAllOrdersResponse, error)
	CreateLiquidityPool(context.Context, *SimpleMsgCreateLiquidityPool) (*MsgCreateLiquidityPoolResponse, error)
	AddLiquidity(context.Context, *SimpleMsgAddLiquidity) (*MsgAddLiquidityResponse, error)
	RemoveLiquidity(context.Context, *SimpleMsgRemoveLiquidity) (*MsgRemoveLiquidityResponse, error)
	Swap(context.Context, *SimpleMsgSwap) (*MsgSwapResponse, error)
	
	// Blockchain-native trading features
	PlaceAtomicSwapOrder(context.Context, *MsgPlaceAtomicSwapOrder) (*MsgPlaceAtomicSwapOrderResponse, error)
	PlaceFractionalOrder(context.Context, *MsgPlaceFractionalOrder) (*MsgPlaceFractionalOrderResponse, error)
	CreateTradingStrategy(context.Context, *MsgCreateTradingStrategy) (*MsgCreateTradingStrategyResponse, error)
}

// QueryServer defines the module's query service.
type QueryServer interface {
	// Order queries
	GetOrder(context.Context, *QueryGetOrderRequest) (*QueryGetOrderResponse, error)
	GetUserOrders(context.Context, *QueryGetUserOrdersRequest) (*QueryGetUserOrdersResponse, error)
	
	// Market queries
	GetMarket(context.Context, *QueryGetMarketRequest) (*QueryGetMarketResponse, error)
	GetMarkets(context.Context, *QueryGetMarketsRequest) (*QueryGetMarketsResponse, error)
	GetOrderBook(context.Context, *QueryGetOrderBookRequest) (*QueryGetOrderBookResponse, error)
	
	// Trade queries
	GetTrade(context.Context, *QueryGetTradeRequest) (*QueryGetTradeResponse, error)
	GetMarketTrades(context.Context, *QueryGetMarketTradesRequest) (*QueryGetMarketTradesResponse, error)
	
	// Atomic swap specific queries
	GetExchangeRate(context.Context, *QueryGetExchangeRateRequest) (*QueryGetExchangeRateResponse, error)
	GetSwapHistory(context.Context, *QueryGetSwapHistoryRequest) (*QueryGetSwapHistoryResponse, error)
	
	// Trading strategy queries
	GetTradingStrategy(context.Context, *QueryGetTradingStrategyRequest) (*QueryGetTradingStrategyResponse, error)
	GetUserStrategies(context.Context, *QueryGetUserStrategiesRequest) (*QueryGetUserStrategiesResponse, error)
}

// RegisterMsgServer registers the message server
func RegisterMsgServer(server interface{}, impl MsgServer) {
	// In a real implementation this would register the gRPC server
	// For our simplified version, we'll handle this in the module
}

// RegisterQueryServer registers the query server
func RegisterQueryServer(server interface{}, impl QueryServer) {
	// In a real implementation this would register the gRPC server
	// For our simplified version, we'll handle this in the module
}