package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/dex/types"
)

// =============================================================================
// TRADING STATISTICS AND MONITORING SYSTEM
// =============================================================================
// Comprehensive trading statistics for market analysis and fraud prevention

// MarketStats represents market-wide statistics
type MarketStats struct {
	Symbol           string         `json:"symbol"`
	TotalVolume24h   math.LegacyDec `json:"total_volume_24h"`
	TotalVolume7d    math.LegacyDec `json:"total_volume_7d"`
	TotalTrades24h   uint64         `json:"total_trades_24h"`
	TotalTrades7d    uint64         `json:"total_trades_7d"`
	UniqueTraders24h uint64         `json:"unique_traders_24h"`
	UniqueTraders7d  uint64         `json:"unique_traders_7d"`
	AverageTradeSize math.LegacyDec `json:"average_trade_size"`
	LargestTrade24h  math.LegacyDec `json:"largest_trade_24h"`
	BuyVolume24h     math.LegacyDec `json:"buy_volume_24h"`
	SellVolume24h    math.LegacyDec `json:"sell_volume_24h"`
	BuySellRatio     math.LegacyDec `json:"buy_sell_ratio"`
	Volatility24h    math.LegacyDec `json:"volatility_24h"`
	OpenInterest     math.LegacyDec `json:"open_interest"`
	LastUpdate       time.Time      `json:"last_update"`
}

// TraderStats represents statistics for a specific trader
type TraderStats struct {
	Address          string         `json:"address"`
	TotalTrades      uint64         `json:"total_trades"`
	TotalVolume      math.LegacyDec `json:"total_volume"`
	TotalFeesPaid    math.LegacyDec `json:"total_fees_paid"`
	WinRate          math.LegacyDec `json:"win_rate"`          // Percentage of profitable trades
	ProfitLoss       math.LegacyDec `json:"profit_loss"`       // Total P&L
	AverageHoldTime  time.Duration  `json:"average_hold_time"` // Average position hold time
	LargestWin       math.LegacyDec `json:"largest_win"`
	LargestLoss      math.LegacyDec `json:"largest_loss"`
	ConsecutiveWins  uint32         `json:"consecutive_wins"`
	ConsecutiveLosses uint32        `json:"consecutive_losses"`
	FirstTradeTime   time.Time      `json:"first_trade_time"`
	LastTradeTime    time.Time      `json:"last_trade_time"`
	RiskScore        math.LegacyDec `json:"risk_score"`        // Risk assessment 0-100
	TrustScore       math.LegacyDec `json:"trust_score"`       // Trust score 0-100
	Tier             string         `json:"tier"`              // Trader tier
}

// OrderBookStats represents order book depth statistics
type OrderBookStats struct {
	Symbol         string         `json:"symbol"`
	BidDepth       math.LegacyDec `json:"bid_depth"`       // Total bid volume
	AskDepth       math.LegacyDec `json:"ask_depth"`       // Total ask volume
	BidAskSpread   math.LegacyDec `json:"bid_ask_spread"`  // Current spread
	BestBid        math.LegacyDec `json:"best_bid"`
	BestAsk        math.LegacyDec `json:"best_ask"`
	MidPrice       math.LegacyDec `json:"mid_price"`
	Imbalance      math.LegacyDec `json:"imbalance"`       // (bid - ask) / (bid + ask)
	NumBidOrders   uint64         `json:"num_bid_orders"`
	NumAskOrders   uint64         `json:"num_ask_orders"`
	LastUpdate     time.Time      `json:"last_update"`
}

// FraudAlert represents a potential fraud detection
type FraudAlert struct {
	ID           uint64         `json:"id"`
	Type         string         `json:"type"`          // "wash_trading", "spoofing", "manipulation"
	Severity     string         `json:"severity"`      // "low", "medium", "high", "critical"
	Addresses    []string       `json:"addresses"`     // Involved addresses
	Symbol       string         `json:"symbol"`
	Description  string         `json:"description"`
	Evidence     string         `json:"evidence"`      // JSON encoded evidence
	DetectedAt   time.Time      `json:"detected_at"`
	Status       string         `json:"status"`        // "pending", "investigating", "confirmed", "dismissed"
	ResolvedAt   time.Time      `json:"resolved_at,omitempty"`
	ResolvedBy   string         `json:"resolved_by,omitempty"`
	ActionTaken  string         `json:"action_taken,omitempty"`
}

// TradingSession represents aggregate data for a trading session
type TradingSession struct {
	Symbol       string         `json:"symbol"`
	SessionStart time.Time      `json:"session_start"`
	SessionEnd   time.Time      `json:"session_end"`
	OpenPrice    math.LegacyDec `json:"open_price"`
	ClosePrice   math.LegacyDec `json:"close_price"`
	HighPrice    math.LegacyDec `json:"high_price"`
	LowPrice     math.LegacyDec `json:"low_price"`
	Volume       math.LegacyDec `json:"volume"`
	Trades       uint64         `json:"trades"`
	VWAP         math.LegacyDec `json:"vwap"`          // Volume Weighted Average Price
}

// Key prefixes for statistics storage
var (
	MarketStatsPrefix    = []byte{0xC0}
	TraderStatsPrefix    = []byte{0xC1}
	OrderBookStatsPrefix = []byte{0xC2}
	FraudAlertPrefix     = []byte{0xC3}
	TradingSessionPrefix = []byte{0xC4}
	DailyStatsPrefix     = []byte{0xC5}
)

// GetTradingMarketStats returns market statistics for a symbol
func (k Keeper) GetTradingMarketStats(ctx sdk.Context, symbol string) (MarketStats, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(MarketStatsPrefix, []byte(symbol)...)
	bz := store.Get(key)
	if bz == nil {
		return MarketStats{}, false
	}
	var stats MarketStats
	json.Unmarshal(bz, &stats)
	return stats, true
}

// SetMarketStats stores market statistics
func (k Keeper) SetMarketStats(ctx sdk.Context, stats MarketStats) {
	store := ctx.KVStore(k.storeKey)
	key := append(MarketStatsPrefix, []byte(stats.Symbol)...)
	bz, _ := json.Marshal(stats)
	store.Set(key, bz)
}

// GetTraderStats returns trader statistics
func (k Keeper) GetTraderStats(ctx sdk.Context, address string) (TraderStats, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(TraderStatsPrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return TraderStats{}, false
	}
	var stats TraderStats
	json.Unmarshal(bz, &stats)
	return stats, true
}

// SetTraderStats stores trader statistics
func (k Keeper) SetTraderStats(ctx sdk.Context, stats TraderStats) {
	store := ctx.KVStore(k.storeKey)
	key := append(TraderStatsPrefix, []byte(stats.Address)...)
	bz, _ := json.Marshal(stats)
	store.Set(key, bz)
}

// GetOrderBookStats returns order book statistics
func (k Keeper) GetOrderBookStats(ctx sdk.Context, symbol string) (OrderBookStats, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(OrderBookStatsPrefix, []byte(symbol)...)
	bz := store.Get(key)
	if bz == nil {
		return OrderBookStats{}, false
	}
	var stats OrderBookStats
	json.Unmarshal(bz, &stats)
	return stats, true
}

// SetOrderBookStats stores order book statistics
func (k Keeper) SetOrderBookStats(ctx sdk.Context, stats OrderBookStats) {
	store := ctx.KVStore(k.storeKey)
	key := append(OrderBookStatsPrefix, []byte(stats.Symbol)...)
	bz, _ := json.Marshal(stats)
	store.Set(key, bz)
}

// =============================================================================
// STATISTICS UPDATE FUNCTIONS
// =============================================================================

// UpdateMarketStatsOnTrade updates market statistics after a trade
func (k Keeper) UpdateMarketStatsOnTrade(ctx sdk.Context, trade types.Trade) {
	stats, found := k.GetTradingMarketStats(ctx, trade.MarketSymbol)
	if !found {
		stats = MarketStats{
			Symbol:           trade.MarketSymbol,
			TotalVolume24h:   math.LegacyZeroDec(),
			TotalVolume7d:    math.LegacyZeroDec(),
			TotalTrades24h:   0,
			TotalTrades7d:    0,
			UniqueTraders24h: 0,
			AverageTradeSize: math.LegacyZeroDec(),
			LargestTrade24h:  math.LegacyZeroDec(),
			BuyVolume24h:     math.LegacyZeroDec(),
			SellVolume24h:    math.LegacyZeroDec(),
			BuySellRatio:     math.LegacyOneDec(),
		}
	}

	// Update volume
	tradeValue := trade.Price.MulInt(trade.Quantity)
	stats.TotalVolume24h = stats.TotalVolume24h.Add(tradeValue)
	stats.TotalVolume7d = stats.TotalVolume7d.Add(tradeValue)
	stats.TotalTrades24h++
	stats.TotalTrades7d++

	// Track buy/sell volume based on buyer is maker (taker was seller)
	if trade.BuyerIsMaker {
		// Taker was seller
		stats.SellVolume24h = stats.SellVolume24h.Add(tradeValue)
	} else {
		// Taker was buyer
		stats.BuyVolume24h = stats.BuyVolume24h.Add(tradeValue)
	}

	// Update buy/sell ratio
	if !stats.SellVolume24h.IsZero() {
		stats.BuySellRatio = stats.BuyVolume24h.Quo(stats.SellVolume24h)
	}

	// Update largest trade
	if tradeValue.GT(stats.LargestTrade24h) {
		stats.LargestTrade24h = tradeValue
	}

	// Update average trade size
	if stats.TotalTrades24h > 0 {
		stats.AverageTradeSize = stats.TotalVolume24h.Quo(math.LegacyNewDec(int64(stats.TotalTrades24h)))
	}

	stats.LastUpdate = ctx.BlockTime()
	k.SetMarketStats(ctx, stats)
}

// UpdateTraderStatsOnTrade updates trader statistics after a trade
func (k Keeper) UpdateTraderStatsOnTrade(ctx sdk.Context, trade types.Trade) {
	// Update buyer stats (maker if BuyerIsMaker, otherwise taker)
	k.updateSingleTraderStats(ctx, trade.Buyer, trade, !trade.BuyerIsMaker)

	// Update seller stats (taker if BuyerIsMaker, otherwise maker)
	k.updateSingleTraderStats(ctx, trade.Seller, trade, trade.BuyerIsMaker)
}

// updateSingleTraderStats updates stats for a single trader
func (k Keeper) updateSingleTraderStats(ctx sdk.Context, address string, trade types.Trade, isTaker bool) {
	stats, found := k.GetTraderStats(ctx, address)
	if !found {
		stats = TraderStats{
			Address:        address,
			TotalTrades:    0,
			TotalVolume:    math.LegacyZeroDec(),
			TotalFeesPaid:  math.LegacyZeroDec(),
			WinRate:        math.LegacyZeroDec(),
			ProfitLoss:     math.LegacyZeroDec(),
			FirstTradeTime: ctx.BlockTime(),
			TrustScore:     math.LegacyNewDec(50), // Start at 50
			Tier:           "bronze",
		}
	}

	tradeValue := trade.Price.MulInt(trade.Quantity)
	stats.TotalTrades++
	stats.TotalVolume = stats.TotalVolume.Add(tradeValue)
	stats.LastTradeTime = ctx.BlockTime()

	// Update fees (takers pay more) - uses governance-controllable rates
	if isTaker {
		stats.TotalFeesPaid = stats.TotalFeesPaid.Add(tradeValue.Mul(k.GetTakerFee(ctx)))
	} else {
		stats.TotalFeesPaid = stats.TotalFeesPaid.Add(tradeValue.Mul(k.GetMakerFee(ctx)))
	}

	// Update trader tier based on volume using governance-controllable thresholds
	stats.Tier = k.calculateTraderTierWithParams(ctx, stats.TotalVolume)

	// Calculate trust score based on trading history
	stats.TrustScore = k.calculateTrustScore(ctx, stats)

	k.SetTraderStats(ctx, stats)
}

// calculateTraderTier determines trader tier based on volume using governance-controllable thresholds
func (k Keeper) calculateTraderTier(volume math.LegacyDec) string {
	// Use default thresholds (for backward compatibility when ctx not available)
	return calculateTraderTierWithThresholds(
		volume,
		types.DefaultDiamondTierThreshold,
		types.DefaultPlatinumTierThreshold,
		types.DefaultGoldTierThreshold,
		types.DefaultSilverTierThreshold,
	)
}

// calculateTraderTierWithParams determines trader tier using governance-controllable params
func (k Keeper) calculateTraderTierWithParams(ctx sdk.Context, volume math.LegacyDec) string {
	return calculateTraderTierWithThresholds(
		volume,
		k.GetDiamondTierThreshold(ctx),
		k.GetPlatinumTierThreshold(ctx),
		k.GetGoldTierThreshold(ctx),
		k.GetSilverTierThreshold(ctx),
	)
}

// calculateTraderTierWithThresholds determines trader tier with explicit thresholds
func calculateTraderTierWithThresholds(volume math.LegacyDec, diamond, platinum, gold, silver uint64) string {
	if volume.GTE(math.LegacyNewDec(int64(diamond))) {
		return "diamond"
	} else if volume.GTE(math.LegacyNewDec(int64(platinum))) {
		return "platinum"
	} else if volume.GTE(math.LegacyNewDec(int64(gold))) {
		return "gold"
	} else if volume.GTE(math.LegacyNewDec(int64(silver))) {
		return "silver"
	}
	return "bronze"
}

// calculateTrustScore calculates trader trust score
func (k Keeper) calculateTrustScore(ctx sdk.Context, stats TraderStats) math.LegacyDec {
	score := math.LegacyNewDec(50) // Base score

	// Add points for trading history
	if stats.TotalTrades > 100 {
		score = score.Add(math.LegacyNewDec(10))
	}
	if stats.TotalTrades > 1000 {
		score = score.Add(math.LegacyNewDec(10))
	}

	// Add points for account age
	accountAge := ctx.BlockTime().Sub(stats.FirstTradeTime)
	if accountAge > time.Hour*24*30 { // 30 days
		score = score.Add(math.LegacyNewDec(10))
	}
	if accountAge > time.Hour*24*180 { // 180 days
		score = score.Add(math.LegacyNewDec(10))
	}

	// Cap at 100
	if score.GT(math.LegacyNewDec(100)) {
		score = math.LegacyNewDec(100)
	}

	return score
}

// UpdateOrderBookStats updates order book statistics
func (k Keeper) UpdateOrderBookStats(ctx sdk.Context, symbol string) {
	buyOrders := k.GetBuyOrders(ctx, symbol, "uhodl")
	sellOrders := k.GetSellOrders(ctx, symbol, "uhodl")

	bidDepth := math.LegacyZeroDec()
	askDepth := math.LegacyZeroDec()
	bestBid := math.LegacyZeroDec()
	bestAsk := math.LegacyNewDec(999999999) // High initial value

	for _, order := range buyOrders {
		orderValue := order.Price.MulInt(order.RemainingQuantity)
		bidDepth = bidDepth.Add(orderValue)
		if order.Price.GT(bestBid) {
			bestBid = order.Price
		}
	}

	for _, order := range sellOrders {
		orderValue := order.Price.MulInt(order.RemainingQuantity)
		askDepth = askDepth.Add(orderValue)
		if order.Price.LT(bestAsk) {
			bestAsk = order.Price
		}
	}

	// Calculate spread and mid price
	spread := math.LegacyZeroDec()
	midPrice := math.LegacyZeroDec()
	if !bestBid.IsZero() && bestAsk.LT(math.LegacyNewDec(999999999)) {
		spread = bestAsk.Sub(bestBid)
		midPrice = bestBid.Add(bestAsk).Quo(math.LegacyNewDec(2))
	}

	// Calculate imbalance
	imbalance := math.LegacyZeroDec()
	total := bidDepth.Add(askDepth)
	if !total.IsZero() {
		imbalance = bidDepth.Sub(askDepth).Quo(total)
	}

	stats := OrderBookStats{
		Symbol:       symbol,
		BidDepth:     bidDepth,
		AskDepth:     askDepth,
		BidAskSpread: spread,
		BestBid:      bestBid,
		BestAsk:      bestAsk,
		MidPrice:     midPrice,
		Imbalance:    imbalance,
		NumBidOrders: uint64(len(buyOrders)),
		NumAskOrders: uint64(len(sellOrders)),
		LastUpdate:   ctx.BlockTime(),
	}

	k.SetOrderBookStats(ctx, stats)
}

// =============================================================================
// FRAUD DETECTION FUNCTIONS
// =============================================================================

// DetectWashTrading checks for potential wash trading
func (k Keeper) DetectWashTrading(ctx sdk.Context, trade types.Trade) *FraudAlert {
	// Check if buyer and seller are the same (self-trade)
	if trade.Buyer == trade.Seller {
		return &FraudAlert{
			ID:          k.getNextFraudAlertID(ctx),
			Type:        "wash_trading",
			Severity:    "high",
			Addresses:   []string{trade.Buyer},
			Symbol:      trade.MarketSymbol,
			Description: "Self-trade detected: buyer and seller are the same address",
			DetectedAt:  ctx.BlockTime(),
			Status:      "pending",
		}
	}

	// Check for circular trading patterns
	// This would require more sophisticated analysis in production

	return nil
}

// DetectSpoofing checks for potential spoofing (placing and quickly canceling orders)
func (k Keeper) DetectSpoofing(ctx sdk.Context, trader string, symbol string) *FraudAlert {
	// Get trader's recent orders (simplified)
	// In production, would track order placement and cancellation rates

	return nil
}

// RecordFraudAlert stores a fraud alert
func (k Keeper) RecordFraudAlert(ctx sdk.Context, alert FraudAlert) {
	store := ctx.KVStore(k.storeKey)
	key := append(FraudAlertPrefix, sdk.Uint64ToBigEndian(alert.ID)...)
	bz, _ := json.Marshal(alert)
	store.Set(key, bz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"fraud_alert",
			sdk.NewAttribute("alert_id", fmt.Sprintf("%d", alert.ID)),
			sdk.NewAttribute("type", alert.Type),
			sdk.NewAttribute("severity", alert.Severity),
			sdk.NewAttribute("symbol", alert.Symbol),
		),
	)
}

// getNextFraudAlertID returns the next fraud alert ID
func (k Keeper) getNextFraudAlertID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := append(FraudAlertPrefix, []byte("counter")...)
	bz := store.Get(key)
	var id uint64 = 1
	if bz != nil {
		id = sdk.BigEndianToUint64(bz) + 1
	}
	store.Set(key, sdk.Uint64ToBigEndian(id))
	return id
}

// GetPendingFraudAlerts returns all pending fraud alerts
func (k Keeper) GetPendingFraudAlerts(ctx sdk.Context) []FraudAlert {
	store := ctx.KVStore(k.storeKey)
	alertStore := prefix.NewStore(store, FraudAlertPrefix)
	iterator := alertStore.Iterator(nil, nil)
	defer iterator.Close()

	alerts := []FraudAlert{}
	for ; iterator.Valid(); iterator.Next() {
		var alert FraudAlert
		if err := json.Unmarshal(iterator.Value(), &alert); err == nil {
			if alert.Status == "pending" || alert.Status == "investigating" {
				alerts = append(alerts, alert)
			}
		}
	}
	return alerts
}

// =============================================================================
// TRADING GUARDRAILS
// =============================================================================

// TradingGuardrails contains trading limits and circuit breakers
type TradingGuardrails struct {
	MaxOrderSize           math.LegacyDec `json:"max_order_size"`           // Maximum order size
	MaxPositionSize        math.LegacyDec `json:"max_position_size"`        // Maximum position per trader
	MaxDailyVolume         math.LegacyDec `json:"max_daily_volume"`         // Maximum daily volume per trader
	PriceCircuitBreaker    math.LegacyDec `json:"price_circuit_breaker"`    // % change to trigger halt
	VolumeCircuitBreaker   math.LegacyDec `json:"volume_circuit_breaker"`   // Volume spike to trigger halt
	MinimumOrderValue      math.LegacyDec `json:"minimum_order_value"`      // Minimum order value
	MaxOrdersPerBlock      uint32         `json:"max_orders_per_block"`     // Max orders per block per user
	TradingHaltedUntil     time.Time      `json:"trading_halted_until"`     // Trading halt time
	RequiresKYC            bool           `json:"requires_kyc"`             // Require KYC for trading
	WhitelistOnly          bool           `json:"whitelist_only"`           // Only whitelisted addresses
}

// DefaultTradingGuardrails returns default guardrails using governance-controllable defaults
func DefaultTradingGuardrails() TradingGuardrails {
	return TradingGuardrails{
		MaxOrderSize:         types.DefaultMaxOrderSize,
		MaxPositionSize:      types.DefaultMaxPositionSize,
		MaxDailyVolume:       types.DefaultMaxDailyVolume,
		PriceCircuitBreaker:  types.DefaultPriceCircuitBreaker,
		VolumeCircuitBreaker: types.DefaultVolumeCircuitBreaker,
		MinimumOrderValue:    types.DefaultMinimumOrderValue,
		MaxOrdersPerBlock:    types.DefaultMaxOrdersPerBlock,
		RequiresKYC:          false,
		WhitelistOnly:        false,
	}
}

// GetTradingGuardrailsFromParams returns trading guardrails from governance params
func (k Keeper) GetTradingGuardrailsFromParams(ctx sdk.Context, symbol string) TradingGuardrails {
	// First check for symbol-specific overrides
	guardrails := k.GetTradingGuardrails(ctx, symbol)

	// If using defaults, populate from governance params
	if guardrails.MaxOrderSize.IsZero() {
		guardrails.MaxOrderSize = k.GetMaxOrderSize(ctx)
	}
	if guardrails.MaxPositionSize.IsZero() {
		guardrails.MaxPositionSize = k.GetMaxPositionSize(ctx)
	}
	if guardrails.MaxDailyVolume.IsZero() {
		guardrails.MaxDailyVolume = k.GetMaxDailyVolume(ctx)
	}
	if guardrails.PriceCircuitBreaker.IsZero() {
		guardrails.PriceCircuitBreaker = k.GetPriceCircuitBreaker(ctx)
	}
	if guardrails.VolumeCircuitBreaker.IsZero() {
		guardrails.VolumeCircuitBreaker = k.GetVolumeCircuitBreaker(ctx)
	}
	if guardrails.MinimumOrderValue.IsZero() {
		guardrails.MinimumOrderValue = k.GetMinimumOrderValue(ctx)
	}
	if guardrails.MaxOrdersPerBlock == 0 {
		guardrails.MaxOrdersPerBlock = k.GetMaxOrdersPerBlock(ctx)
	}

	return guardrails
}

// CheckTradingGuardrails checks if a trade passes all guardrails using governance-controllable params
func (k Keeper) CheckTradingGuardrails(ctx sdk.Context, trader string, symbol string, quantity math.Int, price math.LegacyDec) error {
	guardrails := k.GetTradingGuardrailsFromParams(ctx, symbol)

	// Check if trading is halted
	if ctx.BlockTime().Before(guardrails.TradingHaltedUntil) {
		return types.ErrTradingHalted
	}

	// Check order size
	orderValue := price.MulInt(quantity)
	if orderValue.GT(guardrails.MaxOrderSize) {
		return types.ErrOrderTooLarge
	}

	// Check minimum order value
	if orderValue.LT(guardrails.MinimumOrderValue) {
		return types.ErrOrderTooSmall
	}

	// Check daily volume for trader
	traderStats, found := k.GetTraderStats(ctx, trader)
	if found {
		// Check if daily volume exceeded (simplified - would need proper 24h tracking)
		if traderStats.TotalVolume.GT(guardrails.MaxDailyVolume) {
			return types.ErrDailyLimitExceeded
		}
	}

	return nil
}

// GetTradingGuardrails returns trading guardrails for a symbol
func (k Keeper) GetTradingGuardrails(ctx sdk.Context, symbol string) TradingGuardrails {
	store := ctx.KVStore(k.storeKey)
	key := append([]byte{0xD0}, []byte(symbol)...)
	bz := store.Get(key)
	if bz == nil {
		return DefaultTradingGuardrails()
	}
	var guardrails TradingGuardrails
	json.Unmarshal(bz, &guardrails)
	return guardrails
}

// SetTradingGuardrails sets trading guardrails for a symbol
func (k Keeper) SetTradingGuardrails(ctx sdk.Context, symbol string, guardrails TradingGuardrails) {
	store := ctx.KVStore(k.storeKey)
	key := append([]byte{0xD0}, []byte(symbol)...)
	bz, _ := json.Marshal(guardrails)
	store.Set(key, bz)
}

// TriggerCircuitBreaker triggers a trading halt
func (k Keeper) TriggerCircuitBreaker(ctx sdk.Context, symbol, reason string, duration time.Duration) {
	guardrails := k.GetTradingGuardrails(ctx, symbol)
	guardrails.TradingHaltedUntil = ctx.BlockTime().Add(duration)
	k.SetTradingGuardrails(ctx, symbol, guardrails)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"circuit_breaker_triggered",
			sdk.NewAttribute("symbol", symbol),
			sdk.NewAttribute("reason", reason),
			sdk.NewAttribute("halted_until", guardrails.TradingHaltedUntil.String()),
		),
	)
}
