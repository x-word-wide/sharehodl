package types

import (
	"time"
	
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types for DEX operations (simplified implementation avoiding protobuf)

// SimpleMsgCreateMarket creates a new trading market
type SimpleMsgCreateMarket struct {
	Creator         string          `json:"creator"`
	BaseSymbol      string          `json:"base_symbol"`       // Symbol being traded
	QuoteSymbol     string          `json:"quote_symbol"`      // Symbol used for pricing
	BaseAssetID     uint64          `json:"base_asset_id"`     // Company ID for equity tokens
	QuoteAssetID    uint64          `json:"quote_asset_id"`    // Asset ID for quote currency
	MinOrderSize    math.Int        `json:"min_order_size"`    // Minimum order size
	MaxOrderSize    math.Int        `json:"max_order_size"`    // Maximum order size
	TickSize        math.LegacyDec  `json:"tick_size"`         // Minimum price increment
	LotSize         math.Int        `json:"lot_size"`          // Minimum quantity increment
	MakerFee        math.LegacyDec  `json:"maker_fee"`         // Fee for makers
	TakerFee        math.LegacyDec  `json:"taker_fee"`         // Fee for takers
}

// SimpleMsgPlaceOrder places a new trading order
type SimpleMsgPlaceOrder struct {
	Creator         string          `json:"creator"`
	MarketSymbol    string          `json:"market_symbol"`     // Market to trade on
	Side            string          `json:"side"`              // "buy" or "sell"
	OrderType       string          `json:"order_type"`        // "market", "limit", "stop", "stop_limit"
	TimeInForce     string          `json:"time_in_force"`     // "gtc", "ioc", "fok", "gtd"
	Quantity        math.Int        `json:"quantity"`          // Order quantity
	Price           math.LegacyDec  `json:"price"`             // Limit price (zero for market orders)
	StopPrice       math.LegacyDec  `json:"stop_price"`        // Stop trigger price
	ExpirationTime  int64           `json:"expiration_time"`   // Unix timestamp for GTD orders
	ClientOrderID   string          `json:"client_order_id"`   // Client-provided order ID
}

// SimpleMsgCancelOrder cancels an existing order
type SimpleMsgCancelOrder struct {
	Creator         string          `json:"creator"`
	OrderID         uint64          `json:"order_id"`          // Order to cancel
	ClientOrderID   string          `json:"client_order_id"`   // Alternative: cancel by client order ID
}

// SimpleMsgCancelAllOrders cancels all orders for a user in a market
type SimpleMsgCancelAllOrders struct {
	Creator         string          `json:"creator"`
	MarketSymbol    string          `json:"market_symbol"`     // Market to cancel orders in (optional)
}

// SimpleMsgCreateLiquidityPool creates a new liquidity pool
type SimpleMsgCreateLiquidityPool struct {
	Creator         string          `json:"creator"`
	MarketSymbol    string          `json:"market_symbol"`     // Market for liquidity pool
	BaseAmount      math.Int        `json:"base_amount"`       // Initial base asset amount
	QuoteAmount     math.Int        `json:"quote_amount"`      // Initial quote asset amount
	Fee             math.LegacyDec  `json:"fee"`               // Pool trading fee
}

// SimpleMsgAddLiquidity adds liquidity to an existing pool
type SimpleMsgAddLiquidity struct {
	Creator         string          `json:"creator"`
	MarketSymbol    string          `json:"market_symbol"`     // Target market
	BaseAmount      math.Int        `json:"base_amount"`       // Base asset to add
	QuoteAmount     math.Int        `json:"quote_amount"`      // Quote asset to add
	MinLPTokens     math.Int        `json:"min_lp_tokens"`     // Minimum LP tokens to receive
}

// SimpleMsgRemoveLiquidity removes liquidity from a pool
type SimpleMsgRemoveLiquidity struct {
	Creator         string          `json:"creator"`
	MarketSymbol    string          `json:"market_symbol"`     // Target market
	LPTokenAmount   math.Int        `json:"lp_token_amount"`   // LP tokens to burn
	MinBaseAmount   math.Int        `json:"min_base_amount"`   // Minimum base to receive
	MinQuoteAmount  math.Int        `json:"min_quote_amount"`  // Minimum quote to receive
}

// SimpleMsgSwap performs an AMM swap
type SimpleMsgSwap struct {
	Creator         string          `json:"creator"`
	MarketSymbol    string          `json:"market_symbol"`     // Target market
	InputAsset      string          `json:"input_asset"`       // Asset to swap from
	OutputAsset     string          `json:"output_asset"`      // Asset to swap to
	InputAmount     math.Int        `json:"input_amount"`      // Amount to swap
	MinOutputAmount math.Int        `json:"min_output_amount"` // Minimum output to accept
}

// Response types

// MsgCreateMarketResponse returns market creation result
type MsgCreateMarketResponse struct {
	MarketSymbol    string          `json:"market_symbol"`
	Success         bool            `json:"success"`
}

// MsgPlaceOrderResponse returns order placement result
type MsgPlaceOrderResponse struct {
	OrderID         uint64          `json:"order_id"`
	Status          string          `json:"status"`
	FilledQuantity  math.Int        `json:"filled_quantity"`
	RemainingQuantity math.Int      `json:"remaining_quantity"`
	AveragePrice    math.LegacyDec  `json:"average_price"`
	TotalFees       math.LegacyDec  `json:"total_fees"`
}

// MsgCancelOrderResponse returns order cancellation result
type MsgCancelOrderResponse struct {
	OrderID         uint64          `json:"order_id"`
	Success         bool            `json:"success"`
}

// MsgCancelAllOrdersResponse returns bulk cancellation result
type MsgCancelAllOrdersResponse struct {
	OrdersCancelled uint64          `json:"orders_cancelled"`
	Success         bool            `json:"success"`
}

// MsgCreateLiquidityPoolResponse returns pool creation result
type MsgCreateLiquidityPoolResponse struct {
	PoolAddress     string          `json:"pool_address"`
	LPTokensIssued  math.Int        `json:"lp_tokens_issued"`
	Success         bool            `json:"success"`
}

// MsgAddLiquidityResponse returns liquidity addition result
type MsgAddLiquidityResponse struct {
	LPTokensReceived math.Int       `json:"lp_tokens_received"`
	BaseUsed        math.Int        `json:"base_used"`
	QuoteUsed       math.Int        `json:"quote_used"`
	Success         bool            `json:"success"`
}

// MsgRemoveLiquidityResponse returns liquidity removal result
type MsgRemoveLiquidityResponse struct {
	BaseReceived    math.Int        `json:"base_received"`
	QuoteReceived   math.Int        `json:"quote_received"`
	Success         bool            `json:"success"`
}

// MsgSwapResponse returns swap result
type MsgSwapResponse struct {
	OutputAmount    math.Int        `json:"output_amount"`
	Fee             math.LegacyDec  `json:"fee"`
	PriceImpact     math.LegacyDec  `json:"price_impact"`
	Success         bool            `json:"success"`
}

// Message validation methods

// ValidateBasic validates SimpleMsgCreateMarket
func (msg SimpleMsgCreateMarket) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.BaseSymbol == "" {
		return ErrInvalidMarket
	}
	if msg.QuoteSymbol == "" {
		return ErrInvalidMarket
	}
	if msg.BaseSymbol == msg.QuoteSymbol {
		return ErrInvalidMarket
	}
	if msg.MinOrderSize.IsNil() || msg.MinOrderSize.LTE(math.ZeroInt()) {
		return ErrInvalidOrderSize
	}
	if msg.MaxOrderSize.IsNil() || msg.MaxOrderSize.LT(msg.MinOrderSize) {
		return ErrInvalidOrderSize
	}
	if msg.TickSize.IsNil() || msg.TickSize.LTE(math.LegacyZeroDec()) {
		return ErrInvalidPrice
	}
	if msg.LotSize.IsNil() || msg.LotSize.LTE(math.ZeroInt()) {
		return ErrInvalidOrderSize
	}
	if msg.MakerFee.IsNegative() || msg.TakerFee.IsNegative() {
		return ErrInvalidFee
	}
	return nil
}

// ValidateBasic validates SimpleMsgPlaceOrder
func (msg SimpleMsgPlaceOrder) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.MarketSymbol == "" {
		return ErrInvalidMarket
	}
	if msg.Side != "buy" && msg.Side != "sell" {
		return ErrInvalidOrderSide
	}
	if msg.OrderType != "market" && msg.OrderType != "limit" && 
	   msg.OrderType != "stop" && msg.OrderType != "stop_limit" {
		return ErrInvalidOrderType
	}
	if msg.TimeInForce != "gtc" && msg.TimeInForce != "ioc" && 
	   msg.TimeInForce != "fok" && msg.TimeInForce != "gtd" {
		return ErrInvalidTimeInForce
	}
	if msg.Quantity.IsNil() || msg.Quantity.LTE(math.ZeroInt()) {
		return ErrInvalidOrderSize
	}
	if (msg.OrderType == "limit" || msg.OrderType == "stop_limit") &&
	   (msg.Price.IsNil() || msg.Price.LTE(math.LegacyZeroDec())) {
		return ErrInvalidPrice
	}
	if (msg.OrderType == "stop" || msg.OrderType == "stop_limit") &&
	   (msg.StopPrice.IsNil() || msg.StopPrice.LTE(math.LegacyZeroDec())) {
		return ErrInvalidPrice
	}
	if msg.TimeInForce == "gtd" && msg.ExpirationTime <= time.Now().Unix() {
		return ErrOrderExpired
	}
	return nil
}

// ValidateBasic validates SimpleMsgCancelOrder
func (msg SimpleMsgCancelOrder) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.OrderID == 0 && msg.ClientOrderID == "" {
		return ErrInvalidOrderID
	}
	return nil
}

// ValidateBasic validates SimpleMsgCancelAllOrders
func (msg SimpleMsgCancelAllOrders) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	return nil
}

// ValidateBasic validates SimpleMsgCreateLiquidityPool
func (msg SimpleMsgCreateLiquidityPool) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.MarketSymbol == "" {
		return ErrInvalidMarket
	}
	if msg.BaseAmount.IsNil() || msg.BaseAmount.LTE(math.ZeroInt()) {
		return ErrInvalidLiquidity
	}
	if msg.QuoteAmount.IsNil() || msg.QuoteAmount.LTE(math.ZeroInt()) {
		return ErrInvalidLiquidity
	}
	if msg.Fee.IsNegative() || msg.Fee.GT(math.LegacyNewDecWithPrec(1, 0)) { // Fee > 100%
		return ErrInvalidFee
	}
	return nil
}

// ValidateBasic validates SimpleMsgAddLiquidity
func (msg SimpleMsgAddLiquidity) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.MarketSymbol == "" {
		return ErrInvalidMarket
	}
	if msg.BaseAmount.IsNil() || msg.BaseAmount.LTE(math.ZeroInt()) {
		return ErrInvalidLiquidity
	}
	if msg.QuoteAmount.IsNil() || msg.QuoteAmount.LTE(math.ZeroInt()) {
		return ErrInvalidLiquidity
	}
	if msg.MinLPTokens.IsNil() || msg.MinLPTokens.IsNegative() {
		return ErrInvalidLiquidity
	}
	return nil
}

// ValidateBasic validates SimpleMsgRemoveLiquidity
func (msg SimpleMsgRemoveLiquidity) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.MarketSymbol == "" {
		return ErrInvalidMarket
	}
	if msg.LPTokenAmount.IsNil() || msg.LPTokenAmount.LTE(math.ZeroInt()) {
		return ErrInvalidLiquidity
	}
	if msg.MinBaseAmount.IsNil() || msg.MinBaseAmount.IsNegative() {
		return ErrInvalidLiquidity
	}
	if msg.MinQuoteAmount.IsNil() || msg.MinQuoteAmount.IsNegative() {
		return ErrInvalidLiquidity
	}
	return nil
}

// ValidateBasic validates SimpleMsgSwap
func (msg SimpleMsgSwap) ValidateBasic() error {
	if msg.Creator == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if msg.MarketSymbol == "" {
		return ErrInvalidMarket
	}
	if msg.InputAsset == "" || msg.OutputAsset == "" {
		return ErrInvalidAsset
	}
	if msg.InputAsset == msg.OutputAsset {
		return ErrInvalidAsset
	}
	if msg.InputAmount.IsNil() || msg.InputAmount.LTE(math.ZeroInt()) {
		return ErrInvalidSwapAmount
	}
	if msg.MinOutputAmount.IsNil() || msg.MinOutputAmount.IsNegative() {
		return ErrInvalidSwapAmount
	}
	return nil
}