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
}

// RegisterMsgServer registers the message server
func RegisterMsgServer(server interface{}, impl MsgServer) {
	// In a real implementation this would register the gRPC server
	// For our simplified version, we'll handle this in the module
}