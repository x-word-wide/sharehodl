package types

import (
	"fmt"
	"time"
	
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OrderSide represents whether an order is a buy or sell
type OrderSide int32

const (
	OrderSideBuy  OrderSide = iota
	OrderSideSell
)

func (os OrderSide) String() string {
	switch os {
	case OrderSideBuy:
		return "buy"
	case OrderSideSell:
		return "sell"
	default:
		return "unknown"
	}
}

// OrderType represents the type of order
type OrderType int32

const (
	OrderTypeMarket OrderType = iota // Execute immediately at best available price
	OrderTypeLimit                   // Execute only at specified price or better
	OrderTypeStop                    // Trigger market order when stop price reached
	OrderTypeStopLimit              // Trigger limit order when stop price reached
)

func (ot OrderType) String() string {
	switch ot {
	case OrderTypeMarket:
		return "market"
	case OrderTypeLimit:
		return "limit"
	case OrderTypeStop:
		return "stop"
	case OrderTypeStopLimit:
		return "stop_limit"
	default:
		return "unknown"
	}
}

// OrderStatus represents the current status of an order
type OrderStatus int32

const (
	OrderStatusPending   OrderStatus = iota // Order submitted, waiting to be processed
	OrderStatusOpen                         // Order active in order book
	OrderStatusPartiallyFilled              // Order partially filled
	OrderStatusFilled                       // Order completely filled
	OrderStatusCancelled                    // Order cancelled by user
	OrderStatusRejected                     // Order rejected (insufficient funds, etc.)
	OrderStatusExpired                      // Order expired
)

func (os OrderStatus) String() string {
	switch os {
	case OrderStatusPending:
		return "pending"
	case OrderStatusOpen:
		return "open"
	case OrderStatusPartiallyFilled:
		return "partially_filled"
	case OrderStatusFilled:
		return "filled"
	case OrderStatusCancelled:
		return "cancelled"
	case OrderStatusRejected:
		return "rejected"
	case OrderStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// TimeInForce represents how long an order remains active
type TimeInForce int32

const (
	TimeInForceGTC TimeInForce = iota // Good Till Cancelled
	TimeInForceIOC                    // Immediate Or Cancel
	TimeInForceFOK                    // Fill Or Kill
	TimeInForceGTD                    // Good Till Date
)

func (tif TimeInForce) String() string {
	switch tif {
	case TimeInForceGTC:
		return "gtc"
	case TimeInForceIOC:
		return "ioc"
	case TimeInForceFOK:
		return "fok"
	case TimeInForceGTD:
		return "gtd"
	default:
		return "unknown"
	}
}

// Market represents a trading pair
type Market struct {
	BaseSymbol    string `json:"base_symbol"`    // Symbol being traded (e.g., "APPLE")
	QuoteSymbol   string `json:"quote_symbol"`   // Symbol used for pricing (e.g., "HODL")
	BaseAssetID   uint64 `json:"base_asset_id"`  // Company ID for equity tokens
	QuoteAssetID  uint64 `json:"quote_asset_id"` // Asset ID for quote currency
	
	// Market configuration
	MinOrderSize    math.Int       `json:"min_order_size"`    // Minimum order size in base asset
	MaxOrderSize    math.Int       `json:"max_order_size"`    // Maximum order size in base asset
	TickSize        math.LegacyDec `json:"tick_size"`         // Minimum price increment
	LotSize         math.Int       `json:"lot_size"`          // Minimum quantity increment
	
	// Market status
	Active          bool           `json:"active"`            // Whether market is active for trading
	TradingHalted   bool           `json:"trading_halted"`    // Whether trading is temporarily halted
	
	// Market statistics
	LastPrice       math.LegacyDec `json:"last_price"`        // Last traded price
	Volume24h       math.Int       `json:"volume_24h"`        // 24h trading volume
	High24h         math.LegacyDec `json:"high_24h"`          // 24h high price
	Low24h          math.LegacyDec `json:"low_24h"`           // 24h low price
	PriceChange24h  math.LegacyDec `json:"price_change_24h"`  // 24h price change percentage
	
	// Market making
	MakerFee        math.LegacyDec `json:"maker_fee"`         // Fee for market makers (negative = rebate)
	TakerFee        math.LegacyDec `json:"taker_fee"`         // Fee for market takers
	
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// Order represents a trading order
type Order struct {
	ID              uint64         `json:"id"`                // Unique order ID
	MarketSymbol    string         `json:"market_symbol"`     // Market symbol (e.g., "APPLE/HODL")
	BaseSymbol      string         `json:"base_symbol"`       // Base asset symbol
	QuoteSymbol     string         `json:"quote_symbol"`      // Quote asset symbol
	
	// Order details
	User            string         `json:"user"`              // Order creator address
	Side            OrderSide      `json:"side"`              // Buy or sell
	Type            OrderType      `json:"type"`              // Market, limit, stop, etc.
	Status          OrderStatus    `json:"status"`            // Current order status
	TimeInForce     TimeInForce    `json:"time_in_force"`     // How long order stays active
	
	// Quantities and prices
	Quantity        math.Int       `json:"quantity"`          // Total order quantity
	FilledQuantity  math.Int       `json:"filled_quantity"`   // Quantity already filled
	RemainingQuantity math.Int     `json:"remaining_quantity"` // Quantity remaining to fill
	Price           math.LegacyDec `json:"price"`             // Limit price (zero for market orders)
	StopPrice       math.LegacyDec `json:"stop_price"`        // Stop trigger price
	
	// Order execution
	AveragePrice    math.LegacyDec `json:"average_price"`     // Average fill price
	TotalFees       math.LegacyDec `json:"total_fees"`        // Total fees paid
	
	// Timestamps
	CreatedAt       time.Time      `json:"created_at"`        // Order creation time
	UpdatedAt       time.Time      `json:"updated_at"`        // Last update time
	ExpiresAt       time.Time      `json:"expires_at"`        // Order expiration time
	
	// Client tracking
	ClientOrderID   string         `json:"client_order_id"`   // Client-provided order ID
}

// Trade represents an executed trade
type Trade struct {
	ID              uint64         `json:"id"`                // Unique trade ID
	MarketSymbol    string         `json:"market_symbol"`     // Market where trade occurred
	
	// Trade details
	BuyOrderID      uint64         `json:"buy_order_id"`      // Buy order ID
	SellOrderID     uint64         `json:"sell_order_id"`     // Sell order ID
	Buyer           string         `json:"buyer"`             // Buyer address
	Seller          string         `json:"seller"`            // Seller address
	
	// Trade execution
	Quantity        math.Int       `json:"quantity"`          // Traded quantity
	Price           math.LegacyDec `json:"price"`             // Trade price
	Value           math.LegacyDec `json:"value"`             // Total trade value
	
	// Fees
	BuyerFee        math.LegacyDec `json:"buyer_fee"`         // Fee paid by buyer
	SellerFee       math.LegacyDec `json:"seller_fee"`        // Fee paid by seller
	
	// Market making
	BuyerIsMaker    bool           `json:"buyer_is_maker"`    // Whether buyer was market maker
	
	ExecutedAt      time.Time      `json:"executed_at"`       // Trade execution time
}

// OrderBook represents a market order book
type OrderBook struct {
	MarketSymbol    string         `json:"market_symbol"`     // Market symbol
	
	// Buy orders (bids) - sorted by price descending
	Bids            []OrderBookLevel `json:"bids"`
	
	// Sell orders (asks) - sorted by price ascending  
	Asks            []OrderBookLevel `json:"asks"`
	
	LastUpdated     time.Time      `json:"last_updated"`      // Last update timestamp
}

// OrderBookLevel represents a price level in the order book
type OrderBookLevel struct {
	Price           math.LegacyDec `json:"price"`             // Price level
	Quantity        math.Int       `json:"quantity"`          // Total quantity at this price
	OrderCount      int            `json:"order_count"`       // Number of orders at this price
}

// LiquidityPool represents a liquidity pool for a market
type LiquidityPool struct {
	MarketSymbol    string         `json:"market_symbol"`     // Market symbol
	
	// Pool reserves
	BaseReserve     math.Int       `json:"base_reserve"`      // Base asset reserve
	QuoteReserve    math.Int       `json:"quote_reserve"`     // Quote asset reserve
	
	// Pool tokens
	LPTokenSupply   math.Int       `json:"lp_token_supply"`   // Total LP token supply
	
	// Pool configuration
	Fee             math.LegacyDec `json:"fee"`               // Pool trading fee
	
	// Pool statistics
	Volume24h       math.Int       `json:"volume_24h"`        // 24h trading volume
	FeesCollected24h math.LegacyDec `json:"fees_collected_24h"` // 24h fees collected
	
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// UserPosition represents a user's position in a market
type UserPosition struct {
	User            string         `json:"user"`              // User address
	MarketSymbol    string         `json:"market_symbol"`     // Market symbol
	
	// Asset holdings
	BaseBalance     math.Int       `json:"base_balance"`      // Base asset balance
	QuoteBalance    math.Int       `json:"quote_balance"`     // Quote asset balance
	
	// Locked balances (in open orders)
	BaseLockedBuy   math.Int       `json:"base_locked_buy"`   // Base locked in buy orders
	BaseLockedSell  math.Int       `json:"base_locked_sell"`  // Base locked in sell orders
	QuoteLockedBuy  math.Int       `json:"quote_locked_buy"`  // Quote locked in buy orders
	QuoteLockedSell math.Int       `json:"quote_locked_sell"` // Quote locked in sell orders
	
	// P&L tracking
	RealizedPnL     math.LegacyDec `json:"realized_pnl"`      // Realized profit/loss
	UnrealizedPnL   math.LegacyDec `json:"unrealized_pnl"`    // Unrealized profit/loss
	
	// Trading statistics
	TotalTrades     uint64         `json:"total_trades"`      // Total number of trades
	TotalVolume     math.LegacyDec `json:"total_volume"`      // Total trading volume
	TotalFees       math.LegacyDec `json:"total_fees"`        // Total fees paid
	
	UpdatedAt       time.Time      `json:"updated_at"`
}

// MarketStats represents market statistics
type MarketStats struct {
	MarketSymbol    string         `json:"market_symbol"`     // Market symbol
	
	// Price data
	OpenPrice       math.LegacyDec `json:"open_price"`        // Opening price (24h)
	HighPrice       math.LegacyDec `json:"high_price"`        // High price (24h)
	LowPrice        math.LegacyDec `json:"low_price"`         // Low price (24h)
	LastPrice       math.LegacyDec `json:"last_price"`        // Last trade price
	PriceChange     math.LegacyDec `json:"price_change"`      // Price change (24h)
	PriceChangePercent math.LegacyDec `json:"price_change_percent"` // Price change % (24h)
	
	// Volume data
	Volume          math.Int       `json:"volume"`            // Trading volume (24h)
	QuoteVolume     math.LegacyDec `json:"quote_volume"`      // Quote volume (24h)
	TradeCount      uint64         `json:"trade_count"`       // Number of trades (24h)
	
	// Order book data
	BidPrice        math.LegacyDec `json:"bid_price"`         // Best bid price
	AskPrice        math.LegacyDec `json:"ask_price"`         // Best ask price
	Spread          math.LegacyDec `json:"spread"`            // Bid-ask spread
	SpreadPercent   math.LegacyDec `json:"spread_percent"`    // Spread percentage
	
	LastUpdated     time.Time      `json:"last_updated"`
}

// Validation methods

// Validate validates a Market
func (m Market) Validate() error {
	if m.BaseSymbol == "" {
		return fmt.Errorf("base symbol cannot be empty")
	}
	if m.QuoteSymbol == "" {
		return fmt.Errorf("quote symbol cannot be empty")
	}
	if m.BaseSymbol == m.QuoteSymbol {
		return fmt.Errorf("base and quote symbols cannot be the same")
	}
	if m.MinOrderSize.IsNil() || m.MinOrderSize.LTE(math.ZeroInt()) {
		return fmt.Errorf("minimum order size must be positive")
	}
	if m.MaxOrderSize.IsNil() || m.MaxOrderSize.LT(m.MinOrderSize) {
		return fmt.Errorf("maximum order size must be >= minimum order size")
	}
	if m.TickSize.IsNil() || m.TickSize.LTE(math.LegacyZeroDec()) {
		return fmt.Errorf("tick size must be positive")
	}
	if m.LotSize.IsNil() || m.LotSize.LTE(math.ZeroInt()) {
		return fmt.Errorf("lot size must be positive")
	}
	return nil
}

// Validate validates an Order
func (o Order) Validate() error {
	if o.MarketSymbol == "" {
		return fmt.Errorf("market symbol cannot be empty")
	}
	if o.User == "" {
		return fmt.Errorf("user cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(o.User); err != nil {
		return fmt.Errorf("invalid user address: %v", err)
	}
	if o.Quantity.IsNil() || o.Quantity.LTE(math.ZeroInt()) {
		return fmt.Errorf("quantity must be positive")
	}
	if o.Type == OrderTypeLimit || o.Type == OrderTypeStopLimit {
		if o.Price.IsNil() || o.Price.LTE(math.LegacyZeroDec()) {
			return fmt.Errorf("limit orders must have positive price")
		}
	}
	if o.Type == OrderTypeStop || o.Type == OrderTypeStopLimit {
		if o.StopPrice.IsNil() || o.StopPrice.LTE(math.LegacyZeroDec()) {
			return fmt.Errorf("stop orders must have positive stop price")
		}
	}
	return nil
}

// IsExpired checks if an order has expired
func (o Order) IsExpired() bool {
	if o.TimeInForce == TimeInForceGTD && !o.ExpiresAt.IsZero() {
		return time.Now().After(o.ExpiresAt)
	}
	return false
}

// IsFillable checks if an order can be filled
func (o Order) IsFillable() bool {
	return o.Status == OrderStatusOpen || o.Status == OrderStatusPartiallyFilled
}

// GetSymbol returns the market symbol for an order
func (o Order) GetSymbol() string {
	return o.BaseSymbol + "/" + o.QuoteSymbol
}

// NewMarket creates a new Market
func NewMarket(
	baseSymbol, quoteSymbol string,
	baseAssetID, quoteAssetID uint64,
	minOrderSize, maxOrderSize math.Int,
	tickSize math.LegacyDec,
	lotSize math.Int,
	makerFee, takerFee math.LegacyDec,
) Market {
	now := time.Now()
	return Market{
		BaseSymbol:      baseSymbol,
		QuoteSymbol:     quoteSymbol,
		BaseAssetID:     baseAssetID,
		QuoteAssetID:    quoteAssetID,
		MinOrderSize:    minOrderSize,
		MaxOrderSize:    maxOrderSize,
		TickSize:        tickSize,
		LotSize:         lotSize,
		Active:          true,
		TradingHalted:   false,
		LastPrice:       math.LegacyZeroDec(),
		Volume24h:       math.ZeroInt(),
		High24h:         math.LegacyZeroDec(),
		Low24h:          math.LegacyZeroDec(),
		PriceChange24h:  math.LegacyZeroDec(),
		MakerFee:        makerFee,
		TakerFee:        takerFee,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// NewOrder creates a new Order
func NewOrder(
	id uint64,
	marketSymbol, baseSymbol, quoteSymbol string,
	user string,
	side OrderSide,
	orderType OrderType,
	timeInForce TimeInForce,
	quantity math.Int,
	price math.LegacyDec,
	stopPrice math.LegacyDec,
	clientOrderID string,
) Order {
	now := time.Now()
	return Order{
		ID:                id,
		MarketSymbol:      marketSymbol,
		BaseSymbol:        baseSymbol,
		QuoteSymbol:       quoteSymbol,
		User:              user,
		Side:              side,
		Type:              orderType,
		Status:            OrderStatusPending,
		TimeInForce:       timeInForce,
		Quantity:          quantity,
		FilledQuantity:    math.ZeroInt(),
		RemainingQuantity: quantity,
		Price:             price,
		StopPrice:         stopPrice,
		AveragePrice:      math.LegacyZeroDec(),
		TotalFees:         math.LegacyZeroDec(),
		CreatedAt:         now,
		UpdatedAt:         now,
		ClientOrderID:     clientOrderID,
	}
}

// NewTrade creates a new Trade
func NewTrade(
	id uint64,
	marketSymbol string,
	buyOrderID, sellOrderID uint64,
	buyer, seller string,
	quantity math.Int,
	price math.LegacyDec,
	buyerFee, sellerFee math.LegacyDec,
	buyerIsMaker bool,
) Trade {
	value := price.MulInt(quantity)
	return Trade{
		ID:           id,
		MarketSymbol: marketSymbol,
		BuyOrderID:   buyOrderID,
		SellOrderID:  sellOrderID,
		Buyer:        buyer,
		Seller:       seller,
		Quantity:     quantity,
		Price:        price,
		Value:        value,
		BuyerFee:     buyerFee,
		SellerFee:    sellerFee,
		BuyerIsMaker: buyerIsMaker,
		ExecutedAt:   time.Now(),
	}
}