package keeper

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/dex/types"
)

// =============================================================================
// PRICE ORACLE SYSTEM
// =============================================================================
// Price oracle provides price feeds from multiple sources with confidence scoring

// PriceSource represents a source of price data
type PriceSource struct {
	Name       string         `json:"name"`
	Address    string         `json:"address"`     // Validator address submitting prices
	Weight     math.LegacyDec `json:"weight"`      // Weight in aggregation
	IsActive   bool           `json:"is_active"`
	LastUpdate time.Time      `json:"last_update"`
	Accuracy   math.LegacyDec `json:"accuracy"`    // Historical accuracy score
	TotalFeeds uint64         `json:"total_feeds"` // Total price feeds submitted
}

// PriceFeed represents a single price submission
type PriceFeed struct {
	ID         uint64         `json:"id"`
	Symbol     string         `json:"symbol"`
	Price      math.LegacyDec `json:"price"`
	Volume     math.LegacyDec `json:"volume"`     // 24h volume
	Source     string         `json:"source"`     // Source identifier
	Submitter  string         `json:"submitter"`  // Validator who submitted
	Timestamp  time.Time      `json:"timestamp"`
	BlockHeight int64         `json:"block_height"`
	Confidence math.LegacyDec `json:"confidence"` // Confidence level 0-1
}

// AggregatedPrice represents the final aggregated price
type AggregatedPrice struct {
	Symbol         string         `json:"symbol"`
	Price          math.LegacyDec `json:"price"`
	PriceUSD       math.LegacyDec `json:"price_usd"`
	Change24h      math.LegacyDec `json:"change_24h"`      // 24h price change percentage
	Change7d       math.LegacyDec `json:"change_7d"`       // 7d price change percentage
	High24h        math.LegacyDec `json:"high_24h"`        // 24h high
	Low24h         math.LegacyDec `json:"low_24h"`         // 24h low
	Volume24h      math.LegacyDec `json:"volume_24h"`      // 24h volume
	MarketCap      math.LegacyDec `json:"market_cap"`
	TotalSupply    math.Int       `json:"total_supply"`
	CirculatingSupply math.Int    `json:"circulating_supply"`
	NumSources     uint32         `json:"num_sources"`     // Number of sources used
	Confidence     math.LegacyDec `json:"confidence"`      // Overall confidence
	LastUpdate     time.Time      `json:"last_update"`
	BlockHeight    int64          `json:"block_height"`
}

// PriceHistory stores historical price data for analytics
type PriceHistory struct {
	Symbol    string         `json:"symbol"`
	Timestamp time.Time      `json:"timestamp"`
	Open      math.LegacyDec `json:"open"`
	High      math.LegacyDec `json:"high"`
	Low       math.LegacyDec `json:"low"`
	Close     math.LegacyDec `json:"close"`
	Volume    math.LegacyDec `json:"volume"`
	Interval  string         `json:"interval"` // "1m", "5m", "1h", "1d"
}

// Key prefixes for oracle storage
var (
	PriceSourcePrefix     = []byte{0xB0}
	PriceFeedPrefix       = []byte{0xB1}
	AggregatedPricePrefix = []byte{0xB2}
	PriceHistoryPrefix    = []byte{0xB3}
	OracleParamsKey       = []byte{0xB4}
)

// OracleParams contains oracle configuration
type OracleParams struct {
	MinSources          uint32         `json:"min_sources"`           // Minimum sources for valid price
	MaxPriceAge         time.Duration  `json:"max_price_age"`         // Maximum age of price feed
	PriceDeviationLimit math.LegacyDec `json:"price_deviation_limit"` // Max deviation from median
	UpdateFrequency     time.Duration  `json:"update_frequency"`      // Minimum time between updates
	SlashForBadFeed     math.LegacyDec `json:"slash_for_bad_feed"`    // Slash amount for bad feeds
}

// DefaultOracleParams returns default oracle parameters
func DefaultOracleParams() OracleParams {
	return OracleParams{
		MinSources:          3,
		MaxPriceAge:         time.Minute * 5,
		PriceDeviationLimit: math.LegacyNewDecWithPrec(5, 2),  // 5%
		UpdateFrequency:     time.Second * 30,
		SlashForBadFeed:     math.LegacyNewDecWithPrec(1, 3),  // 0.1%
	}
}

// GetOracleParams returns oracle parameters
func (k Keeper) GetOracleParams(ctx sdk.Context) OracleParams {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(OracleParamsKey)
	if bz == nil {
		return DefaultOracleParams()
	}
	var params OracleParams
	json.Unmarshal(bz, &params)
	return params
}

// SetOracleParams sets oracle parameters
func (k Keeper) SetOracleParams(ctx sdk.Context, params OracleParams) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(params)
	store.Set(OracleParamsKey, bz)
}

// RegisterPriceSource registers a new price source
func (k Keeper) RegisterPriceSource(ctx sdk.Context, name, address string, weight math.LegacyDec) error {
	source := PriceSource{
		Name:       name,
		Address:    address,
		Weight:     weight,
		IsActive:   true,
		LastUpdate: ctx.BlockTime(),
		Accuracy:   math.LegacyOneDec(),
		TotalFeeds: 0,
	}

	store := ctx.KVStore(k.storeKey)
	key := append(PriceSourcePrefix, []byte(address)...)
	bz, _ := json.Marshal(source)
	store.Set(key, bz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"price_source_registered",
			sdk.NewAttribute("name", name),
			sdk.NewAttribute("address", address),
		),
	)

	return nil
}

// SubmitPriceFeed submits a price feed from a validator
func (k Keeper) SubmitPriceFeed(ctx sdk.Context, submitter, symbol string, price, volume math.LegacyDec) error {
	params := k.GetOracleParams(ctx)

	// Verify submitter is a registered price source
	source, found := k.GetPriceSource(ctx, submitter)
	if !found || !source.IsActive {
		return types.ErrUnauthorized
	}

	// Check update frequency
	if ctx.BlockTime().Sub(source.LastUpdate) < params.UpdateFrequency {
		return types.ErrTooFrequentUpdate
	}

	// Get current aggregated price to check deviation
	currentPrice, hasPrice := k.GetAggregatedPrice(ctx, symbol)
	if hasPrice {
		deviation := price.Sub(currentPrice.Price).Abs().Quo(currentPrice.Price)
		if deviation.GT(params.PriceDeviationLimit) {
			// Price deviates too much - could be manipulation or error
			// Don't reject, but lower confidence
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"price_deviation_warning",
					sdk.NewAttribute("symbol", symbol),
					sdk.NewAttribute("submitted_price", price.String()),
					sdk.NewAttribute("current_price", currentPrice.Price.String()),
					sdk.NewAttribute("deviation", deviation.String()),
				),
			)
		}
	}

	// Create price feed
	feed := PriceFeed{
		ID:          k.getNextPriceFeedID(ctx),
		Symbol:      symbol,
		Price:       price,
		Volume:      volume,
		Source:      source.Name,
		Submitter:   submitter,
		Timestamp:   ctx.BlockTime(),
		BlockHeight: ctx.BlockHeight(),
		Confidence:  source.Accuracy,
	}

	// Store feed
	k.setPriceFeed(ctx, feed)

	// Update source
	source.LastUpdate = ctx.BlockTime()
	source.TotalFeeds++
	k.setPriceSource(ctx, source)

	// Recalculate aggregated price
	k.aggregatePrice(ctx, symbol)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"price_feed_submitted",
			sdk.NewAttribute("symbol", symbol),
			sdk.NewAttribute("price", price.String()),
			sdk.NewAttribute("submitter", submitter),
		),
	)

	return nil
}

// GetAggregatedPrice returns the aggregated price for a symbol
func (k Keeper) GetAggregatedPrice(ctx sdk.Context, symbol string) (AggregatedPrice, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(AggregatedPricePrefix, []byte(symbol)...)
	bz := store.Get(key)
	if bz == nil {
		return AggregatedPrice{}, false
	}
	var price AggregatedPrice
	json.Unmarshal(bz, &price)
	return price, true
}

// setAggregatedPrice stores an aggregated price
func (k Keeper) setAggregatedPrice(ctx sdk.Context, price AggregatedPrice) {
	store := ctx.KVStore(k.storeKey)
	key := append(AggregatedPricePrefix, []byte(price.Symbol)...)
	bz, _ := json.Marshal(price)
	store.Set(key, bz)
}

// GetPriceSource returns a price source
func (k Keeper) GetPriceSource(ctx sdk.Context, address string) (PriceSource, bool) {
	store := ctx.KVStore(k.storeKey)
	key := append(PriceSourcePrefix, []byte(address)...)
	bz := store.Get(key)
	if bz == nil {
		return PriceSource{}, false
	}
	var source PriceSource
	json.Unmarshal(bz, &source)
	return source, true
}

// setPriceSource stores a price source
func (k Keeper) setPriceSource(ctx sdk.Context, source PriceSource) {
	store := ctx.KVStore(k.storeKey)
	key := append(PriceSourcePrefix, []byte(source.Address)...)
	bz, _ := json.Marshal(source)
	store.Set(key, bz)
}

// setPriceFeed stores a price feed
func (k Keeper) setPriceFeed(ctx sdk.Context, feed PriceFeed) {
	store := ctx.KVStore(k.storeKey)
	key := append(PriceFeedPrefix, sdk.Uint64ToBigEndian(feed.ID)...)
	bz, _ := json.Marshal(feed)
	store.Set(key, bz)
}

// getNextPriceFeedID returns the next price feed ID
func (k Keeper) getNextPriceFeedID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := append(PriceFeedPrefix, []byte("counter")...)
	bz := store.Get(key)
	var id uint64 = 1
	if bz != nil {
		id = sdk.BigEndianToUint64(bz) + 1
	}
	store.Set(key, sdk.Uint64ToBigEndian(id))
	return id
}

// aggregatePrice aggregates price feeds for a symbol using weighted median
func (k Keeper) aggregatePrice(ctx sdk.Context, symbol string) {
	params := k.GetOracleParams(ctx)
	maxAge := ctx.BlockTime().Add(-params.MaxPriceAge)

	// Get all recent feeds for this symbol
	feeds := k.getRecentPriceFeeds(ctx, symbol, maxAge)

	if uint32(len(feeds)) < params.MinSources {
		// Not enough sources for valid price
		return
	}

	// Calculate weighted median
	totalWeight := math.LegacyZeroDec()
	weightedSum := math.LegacyZeroDec()
	volumeSum := math.LegacyZeroDec()
	confidenceSum := math.LegacyZeroDec()

	for _, feed := range feeds {
		source, found := k.GetPriceSource(ctx, feed.Submitter)
		if !found {
			continue
		}

		weight := source.Weight.Mul(feed.Confidence)
		weightedSum = weightedSum.Add(feed.Price.Mul(weight))
		totalWeight = totalWeight.Add(weight)
		volumeSum = volumeSum.Add(feed.Volume)
		confidenceSum = confidenceSum.Add(feed.Confidence)
	}

	if totalWeight.IsZero() {
		return
	}

	// Calculate weighted average price
	avgPrice := weightedSum.Quo(totalWeight)
	avgConfidence := confidenceSum.Quo(math.LegacyNewDec(int64(len(feeds))))

	// Get previous price for change calculation
	prevPrice, hasPrev := k.GetAggregatedPrice(ctx, symbol)

	change24h := math.LegacyZeroDec()
	if hasPrev && !prevPrice.Price.IsZero() {
		change24h = avgPrice.Sub(prevPrice.Price).Quo(prevPrice.Price).Mul(math.LegacyNewDec(100))
	}

	// Create aggregated price
	aggregated := AggregatedPrice{
		Symbol:      symbol,
		Price:       avgPrice,
		PriceUSD:    avgPrice, // For HODL, price equals USD
		Change24h:   change24h,
		Volume24h:   volumeSum,
		NumSources:  uint32(len(feeds)),
		Confidence:  avgConfidence,
		LastUpdate:  ctx.BlockTime(),
		BlockHeight: ctx.BlockHeight(),
	}

	k.setAggregatedPrice(ctx, aggregated)

	// Store historical data
	k.storePriceHistory(ctx, symbol, avgPrice, volumeSum)
}

// getRecentPriceFeeds returns recent price feeds for a symbol
func (k Keeper) getRecentPriceFeeds(ctx sdk.Context, symbol string, minTime time.Time) []PriceFeed {
	store := ctx.KVStore(k.storeKey)
	feedStore := prefix.NewStore(store, PriceFeedPrefix)
	iterator := feedStore.Iterator(nil, nil)
	defer iterator.Close()

	feeds := []PriceFeed{}
	for ; iterator.Valid(); iterator.Next() {
		var feed PriceFeed
		if err := json.Unmarshal(iterator.Value(), &feed); err == nil {
			if feed.Symbol == symbol && feed.Timestamp.After(minTime) {
				feeds = append(feeds, feed)
			}
		}
	}
	return feeds
}

// storePriceHistory stores historical price data
func (k Keeper) storePriceHistory(ctx sdk.Context, symbol string, price, volume math.LegacyDec) {
	history := PriceHistory{
		Symbol:    symbol,
		Timestamp: ctx.BlockTime(),
		Open:      price,
		High:      price,
		Low:       price,
		Close:     price,
		Volume:    volume,
		Interval:  "1m",
	}

	store := ctx.KVStore(k.storeKey)
	key := append(PriceHistoryPrefix, []byte(symbol)...)
	key = append(key, sdk.Uint64ToBigEndian(uint64(ctx.BlockTime().Unix()))...)
	bz, _ := json.Marshal(history)
	store.Set(key, bz)
}

// GetPriceHistory returns historical price data for a symbol
func (k Keeper) GetPriceHistory(ctx sdk.Context, symbol string, startTime, endTime time.Time) []PriceHistory {
	store := ctx.KVStore(k.storeKey)
	prefixKey := append(PriceHistoryPrefix, []byte(symbol)...)

	iterator := store.Iterator(
		append(prefixKey, sdk.Uint64ToBigEndian(uint64(startTime.Unix()))...),
		append(prefixKey, sdk.Uint64ToBigEndian(uint64(endTime.Unix()))...),
	)
	defer iterator.Close()

	history := []PriceHistory{}
	for ; iterator.Valid(); iterator.Next() {
		var h PriceHistory
		if err := json.Unmarshal(iterator.Value(), &h); err == nil {
			history = append(history, h)
		}
	}
	return history
}
