package types

import (
	"time"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "dex"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key  
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_dex"
)

// Store keys
var (
	// MarketPrefix stores market information
	MarketPrefix = []byte{0x01}
	
	// OrderBookPrefix stores order book data
	OrderBookPrefix = []byte{0x02}
	
	// OrderPrefix stores individual orders
	OrderPrefix = []byte{0x03}
	
	// TradePrefix stores executed trades
	TradePrefix = []byte{0x04}
	
	// LiquidityPoolPrefix stores liquidity pool data
	LiquidityPoolPrefix = []byte{0x05}
	
	// PositionPrefix stores user positions
	PositionPrefix = []byte{0x06}
	
	// MarketStatsPrefix stores market statistics
	MarketStatsPrefix = []byte{0x07}
	
	// TradingStrategyPrefix stores programmable trading strategies
	TradingStrategyPrefix = []byte{0x08}
	
	// OrderCounterKey stores the global order counter
	OrderCounterKey = []byte{0x09}
	
	// TradeCounterKey stores the global trade counter  
	TradeCounterKey = []byte{0x0A}
	
	// Trading controls keys
	TradingControlsPrefix = []byte{0x10}
	DailyVolumePrefix     = []byte{0x11}
	DailyHaltCountPrefix  = []byte{0x12}

	// Authorization keys
	AuthorizedMarketCreatorPrefix = []byte{0x20}

	// Params key
	ParamsKey = []byte{0x30}

	// Equity locking keys for atomic swaps
	LockedEquityPrefix = []byte{0x40}

	// Market-specific order book prefixes for efficient lookups
	// Key format: MarketOrderBuyPrefix + marketSymbol + price (descending) + orderID
	MarketOrderBuyPrefix  = []byte{0x50}
	// Key format: MarketOrderSellPrefix + marketSymbol + price (ascending) + orderID
	MarketOrderSellPrefix = []byte{0x51}

	// LP Position tracking for beneficial ownership
	LPPositionPrefix        = []byte{0x60}
	LPPositionCounterKey    = []byte{0x61}
	LPPositionByUserPrefix  = []byte{0x62}
)

// GetMarketKey returns the store key for a market
func GetMarketKey(baseSymbol, quoteSymbol string) []byte {
	return append(MarketPrefix, []byte(baseSymbol+"/"+quoteSymbol)...)
}

// GetOrderBookKey returns the store key for an order book
func GetOrderBookKey(baseSymbol, quoteSymbol string, side OrderSide) []byte {
	key := append(OrderBookPrefix, []byte(baseSymbol+"/"+quoteSymbol)...)
	if side == OrderSideBuy {
		key = append(key, []byte("/buy")...)
	} else {
		key = append(key, []byte("/sell")...)
	}
	return key
}

// GetOrderKey returns the store key for an order
func GetOrderKey(orderID uint64) []byte {
	return append(OrderPrefix, sdk.Uint64ToBigEndian(orderID)...)
}

// GetUserOrdersKey returns the store key for user orders
func GetUserOrdersKey(user string) []byte {
	return append(OrderPrefix, []byte("user:")...)
}

// GetTradeKey returns the store key for a trade
func GetTradeKey(tradeID uint64) []byte {
	return append(TradePrefix, sdk.Uint64ToBigEndian(tradeID)...)
}

// GetMarketTradesKey returns the store key for market trades
func GetMarketTradesKey(baseSymbol, quoteSymbol string) []byte {
	return append(TradePrefix, []byte("market:"+baseSymbol+"/"+quoteSymbol)...)
}

// GetLiquidityPoolKey returns the store key for a liquidity pool
func GetLiquidityPoolKey(baseSymbol, quoteSymbol string) []byte {
	return append(LiquidityPoolPrefix, []byte(baseSymbol+"/"+quoteSymbol)...)
}

// GetUserPositionKey returns the store key for a user position
func GetUserPositionKey(user, baseSymbol, quoteSymbol string) []byte {
	key := append(PositionPrefix, []byte(user)...)
	key = append(key, []byte(":")...)
	return append(key, []byte(baseSymbol+"/"+quoteSymbol)...)
}

// GetMarketStatsKey returns the store key for market statistics
func GetMarketStatsKey(baseSymbol, quoteSymbol string) []byte {
	return append(MarketStatsPrefix, []byte(baseSymbol+"/"+quoteSymbol)...)
}

// GetTradingStrategyKey returns the store key for a trading strategy
func GetTradingStrategyKey(strategyID string) []byte {
	return append(TradingStrategyPrefix, []byte(strategyID)...)
}

// MarketKey returns the key for a market by symbol
func MarketKey(marketSymbol string) []byte {
	return append(MarketPrefix, []byte(marketSymbol)...)
}

// OrderBookKey returns the key for an order book
func OrderBookKey(marketSymbol string) []byte {
	return append(OrderBookPrefix, []byte(marketSymbol)...)
}

// TradingControlsKey returns the key for trading controls
func TradingControlsKey(marketSymbol string) []byte {
	return append(TradingControlsPrefix, []byte(marketSymbol)...)
}

// DailyVolumeKey returns the key for daily volume
func DailyVolumeKey(marketSymbol string, date time.Time) []byte {
	dateStr := date.Format("2006-01-02")
	key := append(DailyVolumePrefix, []byte(marketSymbol)...)
	return append(key, []byte(dateStr)...)
}

// DailyHaltCountKey returns the key for daily halt count
func DailyHaltCountKey(marketSymbol string, date time.Time) []byte {
	dateStr := date.Format("2006-01-02")
	key := append(DailyHaltCountPrefix, []byte(marketSymbol)...)
	return append(key, []byte(dateStr)...)
}

// GetAuthorizedMarketCreatorKey returns the key for an authorized market creator
func GetAuthorizedMarketCreatorKey(addr sdk.AccAddress) []byte {
	return append(AuthorizedMarketCreatorPrefix, addr.Bytes()...)
}

// GetLockedEquityKey returns the key for locked equity shares
// Key format: prefix + trader_addr + symbol
func GetLockedEquityKey(trader sdk.AccAddress, symbol string) []byte {
	key := append(LockedEquityPrefix, trader.Bytes()...)
	return append(key, []byte(symbol)...)
}

// GetLockedEquityByTraderPrefix returns the prefix for iterating all locked equity for a trader
func GetLockedEquityByTraderPrefix(trader sdk.AccAddress) []byte {
	return append(LockedEquityPrefix, trader.Bytes()...)
}

// GetMarketOrderKey returns the composite key for a market-specific order (for efficient indexing)
// This allows O(log n) lookups instead of O(n) full table scans
func GetMarketOrderKey(baseSymbol, quoteSymbol string, side OrderSide, price string, orderID uint64) []byte {
	marketSymbol := baseSymbol + "/" + quoteSymbol
	var prefix []byte
	if side == OrderSideBuy {
		prefix = MarketOrderBuyPrefix
	} else {
		prefix = MarketOrderSellPrefix
	}

	key := append(prefix, []byte(marketSymbol)...)
	key = append(key, []byte("|")...)
	key = append(key, []byte(price)...)
	key = append(key, []byte("|")...)
	return append(key, sdk.Uint64ToBigEndian(orderID)...)
}

// GetMarketOrderPrefix returns the prefix for iterating orders in a specific market and side
func GetMarketOrderPrefix(baseSymbol, quoteSymbol string, side OrderSide) []byte {
	marketSymbol := baseSymbol + "/" + quoteSymbol
	var prefix []byte
	if side == OrderSideBuy {
		prefix = MarketOrderBuyPrefix
	} else {
		prefix = MarketOrderSellPrefix
	}

	key := append(prefix, []byte(marketSymbol)...)
	return append(key, []byte("|")...)
}

// GetLPTokenDenom returns the LP token denomination for a liquidity pool
// Format: "lp/{base}-{quote}" (e.g., "lp/APPLE-HODL")
// LP tokens are minted to users when they add liquidity and burned when removed
func GetLPTokenDenom(baseSymbol, quoteSymbol string) string {
	return "lp/" + baseSymbol + "-" + quoteSymbol
}

// ParseLPTokenDenom parses an LP token denom and returns base and quote symbols
// Returns empty strings if the denom is not a valid LP token
func ParseLPTokenDenom(denom string) (baseSymbol, quoteSymbol string, valid bool) {
	if len(denom) < 4 || denom[:3] != "lp/" {
		return "", "", false
	}

	symbols := denom[3:] // Remove "lp/" prefix
	for i := 0; i < len(symbols); i++ {
		if symbols[i] == '-' {
			return symbols[:i], symbols[i+1:], true
		}
	}
	return "", "", false
}

// GetLPPositionKey returns the store key for an LP position
func GetLPPositionKey(positionID uint64) []byte {
	return append(LPPositionPrefix, sdk.Uint64ToBigEndian(positionID)...)
}

// GetLPPositionByUserKey returns the store key for user's LP positions
func GetLPPositionByUserKey(user string, positionID uint64) []byte {
	key := append(LPPositionByUserPrefix, []byte(user)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(positionID)...)
}

// GetLPPositionByUserPrefix returns the prefix for iterating user's LP positions
func GetLPPositionByUserPrefix(user string) []byte {
	key := append(LPPositionByUserPrefix, []byte(user)...)
	return append(key, []byte(":")...)
}