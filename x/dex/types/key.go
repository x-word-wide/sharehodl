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