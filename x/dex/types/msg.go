package types

import (
	"strconv"
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

// ===== BLOCKCHAIN-NATIVE MESSAGE TYPES =====

// MsgPlaceAtomicSwapOrder executes instant cross-asset swaps
type MsgPlaceAtomicSwapOrder struct {
	Trader       string                 `json:"trader"`
	FromSymbol   string                 `json:"from_symbol"`
	ToSymbol     string                 `json:"to_symbol"`
	Quantity     uint64                 `json:"quantity"`
	MaxSlippage  math.LegacyDec         `json:"max_slippage"`
	TimeInForce  TimeInForce           `json:"time_in_force"`
}

type MsgPlaceAtomicSwapOrderResponse struct {
	OrderId               string         `json:"order_id"`
	EstimatedExchangeRate math.LegacyDec `json:"estimated_exchange_rate"`
}

// MsgPlaceFractionalOrder handles fractional share trading
type MsgPlaceFractionalOrder struct {
	Trader             string         `json:"trader"`
	Symbol             string         `json:"symbol"`
	Side               OrderSide      `json:"side"`
	FractionalQuantity math.LegacyDec `json:"fractional_quantity"`
	Price              math.LegacyDec `json:"price"`
	MaxPrice           math.LegacyDec `json:"max_price"`    // Maximum acceptable price for buy orders
	TimeInForce        TimeInForce    `json:"time_in_force"`
}

type MsgPlaceFractionalOrderResponse struct {
	OrderId        string         `json:"order_id"`
	EffectivePrice math.LegacyDec `json:"effective_price"`
}

// MsgCreateTradingStrategy creates programmable trading strategies
type MsgCreateTradingStrategy struct {
	Owner       string           `json:"owner"`
	Name        string           `json:"name"`
	Conditions  []TriggerCondition `json:"conditions"`
	Actions     []TradingAction    `json:"actions"`
	MaxExposure math.Int         `json:"max_exposure"`
}

type MsgCreateTradingStrategyResponse struct {
	StrategyId string `json:"strategy_id"`
}

// MsgExecuteStrategy triggers strategy execution
type MsgExecuteStrategy struct {
	Executor       string `json:"executor"`
	StrategyId     string `json:"strategy_id"`
	ForceExecution bool   `json:"force_execution"`
}

type MsgExecuteStrategyResponse struct {
	ExecutedOrderIds []string `json:"executed_order_ids"`
	OrdersPlaced     uint64   `json:"orders_placed"`
}

// Supporting types for atomic swaps and strategies

type AtomicSwapDetails struct {
	TargetSymbol      string         `json:"target_symbol"`
	ExchangeRate      math.LegacyDec `json:"exchange_rate"`
	SlippageTolerance math.LegacyDec `json:"slippage_tolerance"`
}

type FractionalDetails struct {
	FractionalQuantity math.LegacyDec `json:"fractional_quantity"`
	MinFraction        math.LegacyDec `json:"min_fraction"`
}

type TradingStrategy struct {
	StrategyId  string           `json:"strategy_id"`
	Name        string           `json:"name"`
	Owner       string           `json:"owner"`
	Conditions  []TriggerCondition `json:"conditions"`
	Actions     []TradingAction    `json:"actions"`
	MaxExposure math.Int         `json:"max_exposure"`
	IsActive    bool             `json:"is_active"`
	CreatedAt   time.Time        `json:"created_at"`
}

type TriggerCondition struct {
	Symbol         string         `json:"symbol"`
	PriceThreshold math.LegacyDec `json:"price_threshold"`
	ConditionType  string         `json:"condition_type"` // "ABOVE", "BELOW", "EQUALS"
	IsPercentage   bool           `json:"is_percentage"`
}

type TradingAction struct {
	ActionType  OrderType      `json:"action_type"`
	Symbol      string         `json:"symbol"`
	Side        OrderSide      `json:"side"`
	Quantity    math.LegacyDec `json:"quantity"`
	PriceOffset math.LegacyDec `json:"price_offset"`
}

// Validation methods for blockchain-native messages

func (msg MsgPlaceAtomicSwapOrder) ValidateBasic() error {
	if msg.Trader == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Trader); err != nil {
		return ErrUnauthorized
	}
	if msg.FromSymbol == "" || msg.ToSymbol == "" {
		return ErrInvalidAsset
	}
	if msg.FromSymbol == msg.ToSymbol {
		return ErrSameAssetSwap
	}
	if msg.Quantity == 0 {
		return ErrInvalidOrderSize
	}
	if msg.MaxSlippage.IsNegative() {
		return ErrInvalidSlippage
	}
	// Limit slippage to 50% to prevent unreasonable trades
	maxAllowedSlippage := math.LegacyNewDecWithPrec(50, 2) // 50%
	if msg.MaxSlippage.GT(maxAllowedSlippage) {
		return ErrSlippageExceeded
	}
	return nil
}

func (msg MsgPlaceFractionalOrder) ValidateBasic() error {
	if msg.Trader == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Trader); err != nil {
		return ErrUnauthorized
	}
	if msg.Symbol == "" {
		return ErrInvalidAsset
	}
	if msg.FractionalQuantity.IsZero() || msg.FractionalQuantity.IsNegative() {
		return ErrInvalidOrderSize
	}
	if msg.Price.IsZero() || msg.Price.IsNegative() {
		return ErrInvalidPrice
	}
	return nil
}

func (msg MsgCreateTradingStrategy) ValidateBasic() error {
	if msg.Owner == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return ErrUnauthorized
	}
	if msg.Name == "" {
		return ErrInvalidOrderID
	}
	if len(msg.Conditions) == 0 {
		return ErrInvalidOrderType
	}
	if len(msg.Actions) == 0 {
		return ErrInvalidOrderType
	}
	if msg.MaxExposure.IsZero() || msg.MaxExposure.IsNegative() {
		return ErrInvalidOrderSize
	}
	return nil
}

func (msg MsgExecuteStrategy) ValidateBasic() error {
	if msg.Executor == "" {
		return ErrUnauthorized
	}
	if _, err := sdk.AccAddressFromBech32(msg.Executor); err != nil {
		return ErrUnauthorized
	}
	if msg.StrategyId == "" {
		return ErrInvalidOrderID
	}
	return nil
}

// ParseOrderID converts a string order ID to uint64
func ParseOrderID(orderIdStr string) (uint64, error) {
	return strconv.ParseUint(orderIdStr, 10, 64)
}

// ===== QUERY MESSAGE TYPES =====

// QueryGetOrderRequest requests order information by ID
type QueryGetOrderRequest struct {
	OrderId string `json:"order_id"`
}

type QueryGetOrderResponse struct {
	Order Order `json:"order"`
}

// QueryGetUserOrdersRequest requests orders for a specific user
type QueryGetUserOrdersRequest struct {
	User       string `json:"user"`
	Status     string `json:"status,omitempty"`     // Filter by status
	MarketSymbol string `json:"market_symbol,omitempty"` // Filter by market
	Limit      uint64 `json:"limit,omitempty"`      // Pagination limit
	Offset     uint64 `json:"offset,omitempty"`     // Pagination offset
}

type QueryGetUserOrdersResponse struct {
	Orders     []Order `json:"orders"`
	Pagination PaginationResponse `json:"pagination"`
}

// QueryGetMarketRequest requests market information
type QueryGetMarketRequest struct {
	BaseSymbol  string `json:"base_symbol"`
	QuoteSymbol string `json:"quote_symbol"`
}

type QueryGetMarketResponse struct {
	Market Market `json:"market"`
}

// QueryGetMarketsRequest requests all markets
type QueryGetMarketsRequest struct {
	Limit  uint64 `json:"limit,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
}

type QueryGetMarketsResponse struct {
	Markets    []Market `json:"markets"`
	Pagination PaginationResponse `json:"pagination"`
}

// QueryGetOrderBookRequest requests order book data
type QueryGetOrderBookRequest struct {
	BaseSymbol  string `json:"base_symbol"`
	QuoteSymbol string `json:"quote_symbol"`
	Depth       uint32 `json:"depth,omitempty"` // Number of levels to return
}

type QueryGetOrderBookResponse struct {
	OrderBook OrderBook `json:"order_book"`
}

// QueryGetTradeRequest requests trade information by ID
type QueryGetTradeRequest struct {
	TradeId string `json:"trade_id"`
}

type QueryGetTradeResponse struct {
	Trade Trade `json:"trade"`
}

// QueryGetMarketTradesRequest requests trades for a specific market
type QueryGetMarketTradesRequest struct {
	BaseSymbol   string `json:"base_symbol"`
	QuoteSymbol  string `json:"quote_symbol"`
	Limit        uint64 `json:"limit,omitempty"`
	Offset       uint64 `json:"offset,omitempty"`
	FromTime     int64  `json:"from_time,omitempty"` // Unix timestamp
	ToTime       int64  `json:"to_time,omitempty"`   // Unix timestamp
}

type QueryGetMarketTradesResponse struct {
	Trades     []Trade `json:"trades"`
	Pagination PaginationResponse `json:"pagination"`
}

// QueryGetExchangeRateRequest requests current exchange rate
type QueryGetExchangeRateRequest struct {
	FromSymbol string `json:"from_symbol"`
	ToSymbol   string `json:"to_symbol"`
}

type QueryGetExchangeRateResponse struct {
	FromSymbol   string         `json:"from_symbol"`
	ToSymbol     string         `json:"to_symbol"`
	ExchangeRate math.LegacyDec `json:"exchange_rate"`
	Timestamp    int64          `json:"timestamp"` // Unix timestamp
}

// QueryGetSwapHistoryRequest requests atomic swap history for user
type QueryGetSwapHistoryRequest struct {
	User       string `json:"user"`
	FromSymbol string `json:"from_symbol,omitempty"` // Filter by source asset
	ToSymbol   string `json:"to_symbol,omitempty"`   // Filter by target asset
	Limit      uint64 `json:"limit,omitempty"`
	Offset     uint64 `json:"offset,omitempty"`
	FromTime   int64  `json:"from_time,omitempty"`
	ToTime     int64  `json:"to_time,omitempty"`
}

type QueryGetSwapHistoryResponse struct {
	Swaps      []AtomicSwapRecord `json:"swaps"`
	Pagination PaginationResponse `json:"pagination"`
}

// QueryGetTradingStrategyRequest requests trading strategy by ID
type QueryGetTradingStrategyRequest struct {
	StrategyId string `json:"strategy_id"`
}

type QueryGetTradingStrategyResponse struct {
	Strategy TradingStrategy `json:"strategy"`
}

// QueryGetUserStrategiesRequest requests all strategies for a user
type QueryGetUserStrategiesRequest struct {
	User   string `json:"user"`
	Active bool   `json:"active,omitempty"` // Filter by active status
	Limit  uint64 `json:"limit,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
}

type QueryGetUserStrategiesResponse struct {
	Strategies []TradingStrategy  `json:"strategies"`
	Pagination PaginationResponse `json:"pagination"`
}

// ===== SUPPORTING TYPES =====

// PaginationResponse provides pagination metadata
type PaginationResponse struct {
	NextKey string `json:"next_key,omitempty"` // Key for next page
	Total   uint64 `json:"total"`              // Total number of items
	Limit   uint64 `json:"limit"`              // Items per page
	Offset  uint64 `json:"offset"`             // Current offset
}

// AtomicSwapRecord represents a completed atomic swap
type AtomicSwapRecord struct {
	OrderId      uint64         `json:"order_id"`
	TradeId      uint64         `json:"trade_id"`
	User         string         `json:"user"`
	FromSymbol   string         `json:"from_symbol"`
	ToSymbol     string         `json:"to_symbol"`
	InputAmount  math.Int       `json:"input_amount"`
	OutputAmount math.LegacyDec `json:"output_amount"`
	ExchangeRate math.LegacyDec `json:"exchange_rate"`
	Slippage     math.LegacyDec `json:"slippage"`
	Fee          math.LegacyDec `json:"fee"`
	ExecutedAt   time.Time      `json:"executed_at"`
}