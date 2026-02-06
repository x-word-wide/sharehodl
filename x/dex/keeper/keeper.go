package keeper

import (
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/dex/types"
)

// Keeper of the dex store
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	memKey     storetypes.StoreKey
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
	equityKeeper  types.EquityKeeper
	hodlKeeper    types.HODLKeeper
}

// NewKeeper creates a new dex Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	equityKeeper types.EquityKeeper,
	hodlKeeper types.HODLKeeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		equityKeeper:  equityKeeper,
		hodlKeeper:    hodlKeeper,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// =============================================================================
// PARAMS MANAGEMENT - All values governance-controllable
// =============================================================================

// GetParams returns the current DEX parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

// SetParams sets the DEX parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}

// Param getters for convenience

// GetOrderBookDepth returns the governance-controllable order book depth
func (k Keeper) GetOrderBookDepth(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).OrderBookDepth
}

// GetBaseFee returns the governance-controllable base transaction fee
func (k Keeper) GetBaseFee(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).BaseFee
}

// GetTradingFee returns the governance-controllable trading fee
func (k Keeper) GetTradingFee(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.TradingFee.IsNil() {
		return types.DefaultTradingFee
	}
	return params.TradingFee
}

// GetAtomicSwapFee returns the governance-controllable atomic swap fee
func (k Keeper) GetAtomicSwapFee(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.AtomicSwapFee.IsNil() {
		return types.DefaultAtomicSwapFee
	}
	return params.AtomicSwapFee
}

// GetEquitySwapFee returns the governance-controllable equity swap fee
func (k Keeper) GetEquitySwapFee(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.EquitySwapFee.IsNil() {
		return types.DefaultEquitySwapFee
	}
	return params.EquitySwapFee
}

// GetAutoSwapSlippage returns the governance-controllable auto-swap slippage
func (k Keeper) GetAutoSwapSlippage(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.AutoSwapSlippage.IsNil() {
		return types.DefaultAutoSwapSlippage
	}
	return params.AutoSwapSlippage
}

// GetValidatorFeeShare returns the governance-controllable validator fee share
func (k Keeper) GetValidatorFeeShare(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).ValidatorFeeShare
}

// IsTradingEnabled returns whether trading is enabled
func (k Keeper) IsTradingEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).TradingEnabled
}

// GetMakerFee returns the governance-controllable maker fee
func (k Keeper) GetMakerFee(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.MakerFee.IsNil() {
		return types.DefaultMakerFee
	}
	return params.MakerFee
}

// GetTakerFee returns the governance-controllable taker fee
func (k Keeper) GetTakerFee(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.TakerFee.IsNil() {
		return types.DefaultTakerFee
	}
	return params.TakerFee
}

// Trader tier threshold getters

// GetDiamondTierThreshold returns the governance-controllable diamond tier threshold
func (k Keeper) GetDiamondTierThreshold(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).DiamondTierThreshold
}

// GetPlatinumTierThreshold returns the governance-controllable platinum tier threshold
func (k Keeper) GetPlatinumTierThreshold(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).PlatinumTierThreshold
}

// GetGoldTierThreshold returns the governance-controllable gold tier threshold
func (k Keeper) GetGoldTierThreshold(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).GoldTierThreshold
}

// GetSilverTierThreshold returns the governance-controllable silver tier threshold
func (k Keeper) GetSilverTierThreshold(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).SilverTierThreshold
}

// Trading guardrail getters

// GetMaxOrderSize returns the governance-controllable max order size
func (k Keeper) GetMaxOrderSize(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.MaxOrderSize.IsNil() {
		return types.DefaultMaxOrderSize
	}
	return params.MaxOrderSize
}

// GetMaxPositionSize returns the governance-controllable max position size
func (k Keeper) GetMaxPositionSize(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.MaxPositionSize.IsNil() {
		return types.DefaultMaxPositionSize
	}
	return params.MaxPositionSize
}

// GetMaxDailyVolume returns the governance-controllable max daily volume
func (k Keeper) GetMaxDailyVolume(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.MaxDailyVolume.IsNil() {
		return types.DefaultMaxDailyVolume
	}
	return params.MaxDailyVolume
}

// GetPriceCircuitBreaker returns the governance-controllable price circuit breaker
func (k Keeper) GetPriceCircuitBreaker(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.PriceCircuitBreaker.IsNil() {
		return types.DefaultPriceCircuitBreaker
	}
	return params.PriceCircuitBreaker
}

// GetVolumeCircuitBreaker returns the governance-controllable volume circuit breaker
func (k Keeper) GetVolumeCircuitBreaker(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.VolumeCircuitBreaker.IsNil() {
		return types.DefaultVolumeCircuitBreaker
	}
	return params.VolumeCircuitBreaker
}

// GetMinimumOrderValue returns the governance-controllable minimum order value
func (k Keeper) GetMinimumOrderValue(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.MinimumOrderValue.IsNil() {
		return types.DefaultMinimumOrderValue
	}
	return params.MinimumOrderValue
}

// GetMaxOrdersPerBlock returns the governance-controllable max orders per block
func (k Keeper) GetMaxOrdersPerBlock(ctx sdk.Context) uint32 {
	params := k.GetParams(ctx)
	if params.MaxOrdersPerBlock == 0 {
		return types.DefaultMaxOrdersPerBlock
	}
	return params.MaxOrdersPerBlock
}

// Market management methods

// GetNextOrderID returns the next order ID and increments the counter
func (k Keeper) GetNextOrderID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.OrderCounterKey)
	
	var counter uint64 = 1 // Start from 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}
	
	// Increment counter for next use
	store.Set(types.OrderCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetNextTradeID returns the next trade ID and increments the counter
func (k Keeper) GetNextTradeID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TradeCounterKey)
	
	var counter uint64 = 1 // Start from 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}
	
	// Increment counter for next use
	store.Set(types.TradeCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetMarket returns a market by base and quote symbols
func (k Keeper) GetMarket(ctx sdk.Context, baseSymbol, quoteSymbol string) (types.Market, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetMarketKey(baseSymbol, quoteSymbol)
	bz := store.Get(key)
	if bz == nil {
		return types.Market{}, false
	}

	var market types.Market
	if err := json.Unmarshal(bz, &market); err != nil {
		return types.Market{}, false
	}
	return market, true
}

// GetMarketBySymbol returns a market by combined symbol (e.g., "AAPL/HODL")
// Parses the symbol to extract base and quote components
func (k Keeper) GetMarketBySymbol(ctx sdk.Context, marketSymbol string) (types.Market, bool) {
	// Parse market symbol (format: "BASE/QUOTE")
	parts := strings.Split(marketSymbol, "/")
	if len(parts) != 2 {
		return types.Market{}, false
	}
	return k.GetMarket(ctx, parts[0], parts[1])
}

// SetMarket stores a market
// SECURITY FIX: Returns error instead of panicking
func (k Keeper) SetMarket(ctx sdk.Context, market types.Market) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetMarketKey(market.BaseSymbol, market.QuoteSymbol)
	bz, err := json.Marshal(market)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal market", "error", err)
		return fmt.Errorf("failed to marshal market: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// DeleteMarket removes a market
func (k Keeper) DeleteMarket(ctx sdk.Context, baseSymbol, quoteSymbol string) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetMarketKey(baseSymbol, quoteSymbol)
	store.Delete(key)
}

// IsMarketActive checks if a market is active for trading
func (k Keeper) IsMarketActive(ctx sdk.Context, baseSymbol, quoteSymbol string) bool {
	market, found := k.GetMarket(ctx, baseSymbol, quoteSymbol)
	if !found {
		return false
	}
	return market.Active && !market.TradingHalted
}

// Order management methods

// GetOrder returns an order by ID
func (k Keeper) GetOrder(ctx sdk.Context, orderID uint64) (types.Order, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetOrderKey(orderID)
	bz := store.Get(key)
	if bz == nil {
		return types.Order{}, false
	}
	
	var order types.Order
	if err := json.Unmarshal(bz, &order); err != nil {
		return types.Order{}, false
	}
	return order, true
}

// SetOrder stores an order
// SECURITY FIX: Returns error instead of panicking
// PERFORMANCE FIX: Also stores in market-specific index for efficient lookups
func (k Keeper) SetOrder(ctx sdk.Context, order types.Order) error {
	store := ctx.KVStore(k.storeKey)

	// Store in primary key (orderID)
	key := types.GetOrderKey(order.ID)
	bz, err := json.Marshal(order)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal order", "error", err)
		return fmt.Errorf("failed to marshal order: %w", err)
	}
	store.Set(key, bz)

	// PERFORMANCE FIX: Store in market-specific index for efficient lookups
	// Only index fillable orders
	if order.IsFillable() {
		marketKey := types.GetMarketOrderKey(
			order.BaseSymbol,
			order.QuoteSymbol,
			order.Side,
			order.Price.String(),
			order.ID,
		)
		store.Set(marketKey, bz)
	}

	return nil
}

// DeleteOrder removes an order
// PERFORMANCE FIX: Also removes from market-specific index
func (k Keeper) DeleteOrder(ctx sdk.Context, orderID uint64) {
	store := ctx.KVStore(k.storeKey)

	// Get the order first to remove from market index
	order, found := k.GetOrder(ctx, orderID)
	if found {
		// Delete from market-specific index
		marketKey := types.GetMarketOrderKey(
			order.BaseSymbol,
			order.QuoteSymbol,
			order.Side,
			order.Price.String(),
			order.ID,
		)
		store.Delete(marketKey)
	}

	// Delete from primary key
	key := types.GetOrderKey(orderID)
	store.Delete(key)
}

// GetTrade returns a trade by ID
func (k Keeper) GetTrade(ctx sdk.Context, tradeID uint64) (types.Trade, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradeKey(tradeID)
	bz := store.Get(key)
	if bz == nil {
		return types.Trade{}, false
	}
	
	var trade types.Trade
	if err := json.Unmarshal(bz, &trade); err != nil {
		return types.Trade{}, false
	}
	return trade, true
}

// SetTrade stores a trade
// SECURITY FIX: Returns error instead of panicking
func (k Keeper) SetTrade(ctx sdk.Context, trade types.Trade) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradeKey(trade.ID)
	bz, err := json.Marshal(trade)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal trade", "error", err)
		return fmt.Errorf("failed to marshal trade: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// Order book management

// GetOrderBook returns the order book for a market
func (k Keeper) GetOrderBook(ctx sdk.Context, baseSymbol, quoteSymbol string) types.OrderBook {
	// Use the aggregated order book from matching engine
	return k.GetOrderBookAggregated(ctx, baseSymbol, quoteSymbol, 50) // Return top 50 levels
}

// Business logic methods

// CreateMarket creates a new trading market
func (k Keeper) CreateMarket(
	ctx sdk.Context,
	baseSymbol, quoteSymbol string,
	baseAssetID, quoteAssetID uint64,
	minOrderSize, maxOrderSize math.Int,
	tickSize math.LegacyDec,
	lotSize math.Int,
	makerFee, takerFee math.LegacyDec,
) error {
	// Check if market already exists
	if _, exists := k.GetMarket(ctx, baseSymbol, quoteSymbol); exists {
		return types.ErrMarketAlreadyExists
	}
	
	// Create and validate market
	market := types.NewMarket(
		baseSymbol, quoteSymbol,
		baseAssetID, quoteAssetID,
		minOrderSize, maxOrderSize,
		tickSize, lotSize,
		makerFee, takerFee,
	)
	
	if err := market.Validate(); err != nil {
		return err
	}

	// Store market
	return k.SetMarket(ctx, market)
}

// PlaceOrder places a new order in the order book
func (k Keeper) PlaceOrder(
	ctx sdk.Context,
	user string,
	marketSymbol string,
	side types.OrderSide,
	orderType types.OrderType,
	timeInForce types.TimeInForce,
	quantity math.Int,
	price math.LegacyDec,
	stopPrice math.LegacyDec,
	clientOrderID string,
) (types.Order, error) {
	// Parse market symbol
	baseSymbol, quoteSymbol := k.parseMarketSymbol(marketSymbol)

	// Check if market exists and is active
	if !k.IsMarketActive(ctx, baseSymbol, quoteSymbol) {
		return types.Order{}, types.ErrMarketInactive
	}

	// Check if trading is halted for this equity
	if k.equityKeeper != nil && baseSymbol != "HODL" && quoteSymbol != "HODL" {
		// Extract company ID from the trading pair
		companyID, found := k.GetCompanyIDBySymbol(ctx, baseSymbol)
		if found && k.equityKeeper.IsTradingHalted(ctx, companyID) {
			k.Logger(ctx).Info("order rejected - trading halted",
				"market", marketSymbol,
				"company_id", companyID,
				"user", user,
			)

			// Emit order rejected event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeOrderRejected,
					sdk.NewAttribute("market", marketSymbol),
					sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyID)),
					sdk.NewAttribute("user", user),
					sdk.NewAttribute("reason", "trading_halted"),
				),
			)

			return types.Order{}, types.ErrTradingHalted
		}
	}
	
	// Get next order ID
	orderID := k.GetNextOrderID(ctx)
	
	// Create order
	order := types.NewOrder(
		orderID,
		marketSymbol,
		baseSymbol,
		quoteSymbol,
		user,
		side,
		orderType,
		timeInForce,
		quantity,
		price,
		stopPrice,
		clientOrderID,
	)
	
	// Validate order
	if err := order.Validate(); err != nil {
		return types.Order{}, err
	}
	
	// Check user has sufficient funds
	userAddr, err := sdk.AccAddressFromBech32(user)
	if err != nil {
		return types.Order{}, types.ErrUnauthorized
	}
	
	// For simplified implementation, check basic balance
	// In full implementation would lock funds and check trading balances
	if err := k.checkOrderFunds(ctx, userAddr, order); err != nil {
		return types.Order{}, err
	}
	
	// Set order status to open
	order.Status = types.OrderStatusOpen

	// Lock funds for the order
	if err := k.lockOrderFunds(ctx, userAddr, order); err != nil {
		return types.Order{}, err
	}

	// Store order
	k.SetOrder(ctx, order)

	// Execute order matching based on order type
	var trades []types.Trade
	var updatedOrder types.Order
	var matchErr error

	switch order.Type {
	case types.OrderTypeMarket:
		trades, updatedOrder, matchErr = k.ProcessMarketOrder(ctx, order)
	case types.OrderTypeLimit:
		trades, updatedOrder, matchErr = k.ProcessLimitOrder(ctx, order)
	case types.OrderTypeStop, types.OrderTypeStopLimit:
		// Stop orders wait for trigger, don't match immediately
		updatedOrder = order
	default:
		trades, updatedOrder, matchErr = k.MatchOrder(ctx, order)
	}

	if matchErr != nil {
		k.Logger(ctx).Error("order matching failed", "error", matchErr, "order_id", order.ID)
	}

	// Emit order placed event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"order_placed",
			sdk.NewAttribute("order_id", fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute("user", user),
			sdk.NewAttribute("market", marketSymbol),
			sdk.NewAttribute("side", order.Side.String()),
			sdk.NewAttribute("type", order.Type.String()),
			sdk.NewAttribute("quantity", order.Quantity.String()),
			sdk.NewAttribute("price", order.Price.String()),
			sdk.NewAttribute("status", updatedOrder.Status.String()),
			sdk.NewAttribute("trades_count", fmt.Sprintf("%d", len(trades))),
		),
	)

	return updatedOrder, nil
}

// CancelOrder cancels an existing order
func (k Keeper) CancelOrder(ctx sdk.Context, userAddr sdk.AccAddress, orderID uint64) error {
	// Get order
	order, found := k.GetOrder(ctx, orderID)
	if !found {
		return types.ErrOrderNotFound
	}
	
	// Check user owns the order
	if order.User != userAddr.String() {
		return types.ErrUnauthorized
	}
	
	// Check order can be cancelled
	if order.Status != types.OrderStatusOpen && order.Status != types.OrderStatusPartiallyFilled {
		return types.ErrCannotCancelOrder
	}
	
	// Release locked funds
	if err := k.unlockOrderFunds(ctx, order); err != nil {
		return fmt.Errorf("failed to unlock funds: %w", err)
	}

	// Update order status
	order.Status = types.OrderStatusCancelled
	order.UpdatedAt = ctx.BlockTime()
	k.SetOrder(ctx, order)

	// Emit cancel event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOrderCancelled,
			sdk.NewAttribute("order_id", fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute("user", order.User),
			sdk.NewAttribute("market", order.MarketSymbol),
		),
	)

	return nil
}

// MatchOrders performs order matching using the matching engine
func (k Keeper) MatchOrders(ctx sdk.Context, newOrder types.Order) []types.Trade {
	// Use the full matching engine
	trades, _, err := k.MatchOrder(ctx, newOrder)
	if err != nil {
		k.Logger(ctx).Error("order matching failed", "error", err, "order_id", newOrder.ID)
		return []types.Trade{}
	}
	return trades
}

// Liquidity pool methods (simplified)

// GetLiquidityPool returns a liquidity pool
func (k Keeper) GetLiquidityPool(ctx sdk.Context, baseSymbol, quoteSymbol string) (types.LiquidityPool, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLiquidityPoolKey(baseSymbol, quoteSymbol)
	bz := store.Get(key)
	if bz == nil {
		return types.LiquidityPool{}, false
	}
	
	var pool types.LiquidityPool
	if err := json.Unmarshal(bz, &pool); err != nil {
		return types.LiquidityPool{}, false
	}
	return pool, true
}

// SetLiquidityPool stores a liquidity pool
func (k Keeper) SetLiquidityPool(ctx sdk.Context, pool types.LiquidityPool) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLiquidityPoolKey(pool.MarketSymbol, "")
	bz, err := json.Marshal(pool)
	if err != nil {
		return fmt.Errorf("failed to marshal liquidity pool: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// GetAllLiquidityPools returns all liquidity pools
func (k Keeper) GetAllLiquidityPools(ctx sdk.Context) []types.LiquidityPool {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.LiquidityPoolPrefix, nil)
	defer iterator.Close()

	var pools []types.LiquidityPool
	for ; iterator.Valid(); iterator.Next() {
		var pool types.LiquidityPool
		if err := json.Unmarshal(iterator.Value(), &pool); err != nil {
			continue
		}
		pools = append(pools, pool)
	}
	return pools
}

// LP Position Management for Beneficial Ownership Tracking

// getNextLPPositionID returns the next LP position ID and increments the counter
func (k Keeper) getNextLPPositionID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LPPositionCounterKey)

	var id uint64 = 1
	if bz != nil {
		id = sdk.BigEndianToUint64(bz) + 1
	}

	store.Set(types.LPPositionCounterKey, sdk.Uint64ToBigEndian(id))
	return id
}

// SetLPPosition stores an LP position
func (k Keeper) SetLPPosition(ctx sdk.Context, position types.LPPosition) error {
	store := ctx.KVStore(k.storeKey)

	bz, err := json.Marshal(position)
	if err != nil {
		return fmt.Errorf("failed to marshal LP position: %w", err)
	}

	// Store by position ID
	store.Set(types.GetLPPositionKey(position.ID), bz)

	// Store index by user for efficient lookups
	store.Set(types.GetLPPositionByUserKey(position.Provider, position.ID), sdk.Uint64ToBigEndian(position.ID))

	return nil
}

// GetLPPosition retrieves an LP position by ID
func (k Keeper) GetLPPosition(ctx sdk.Context, positionID uint64) (types.LPPosition, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetLPPositionKey(positionID))
	if bz == nil {
		return types.LPPosition{}, false
	}

	var position types.LPPosition
	if err := json.Unmarshal(bz, &position); err != nil {
		return types.LPPosition{}, false
	}
	return position, true
}

// GetLPPositionsByUser returns all LP positions for a user
func (k Keeper) GetLPPositionsByUser(ctx sdk.Context, user string) []types.LPPosition {
	store := ctx.KVStore(k.storeKey)
	iterator := store.Iterator(types.GetLPPositionByUserPrefix(user), nil)
	defer iterator.Close()

	var positions []types.LPPosition
	for ; iterator.Valid(); iterator.Next() {
		positionID := sdk.BigEndianToUint64(iterator.Value())
		if position, found := k.GetLPPosition(ctx, positionID); found {
			positions = append(positions, position)
		}
	}
	return positions
}

// GetLPPositionByUserAndMarket finds an LP position for a user in a specific market
func (k Keeper) GetLPPositionByUserAndMarket(ctx sdk.Context, user, marketSymbol string) (types.LPPosition, bool) {
	positions := k.GetLPPositionsByUser(ctx, user)
	for _, pos := range positions {
		if pos.MarketSymbol == marketSymbol {
			return pos, true
		}
	}
	return types.LPPosition{}, false
}

// DeleteLPPosition removes an LP position
func (k Keeper) DeleteLPPosition(ctx sdk.Context, position types.LPPosition) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetLPPositionKey(position.ID))
	store.Delete(types.GetLPPositionByUserKey(position.Provider, position.ID))
}

// UpdateLPPositionShares updates the shares in an LP position (for partial withdrawals)
func (k Keeper) UpdateLPPositionShares(ctx sdk.Context, positionID uint64, newBaseAmount, newLPTokenAmount math.Int) error {
	position, found := k.GetLPPosition(ctx, positionID)
	if !found {
		return fmt.Errorf("LP position not found: %d", positionID)
	}

	position.BaseAmount = newBaseAmount
	position.LPTokenAmount = newLPTokenAmount
	position.UpdatedAt = ctx.BlockTime()

	return k.SetLPPosition(ctx, position)
}

// Helper methods

// parseMarketSymbol parses a market symbol like "APPLE/HODL" into base and quote symbols
func (k Keeper) parseMarketSymbol(marketSymbol string) (string, string) {
	// Simple parsing - in production would use more robust parsing
	for i := 0; i < len(marketSymbol); i++ {
		if marketSymbol[i] == '/' {
			return marketSymbol[:i], marketSymbol[i+1:]
		}
	}
	return marketSymbol, "HODL" // Default to HODL as quote
}

// lockOrderFunds locks user funds in the module account for an order
func (k Keeper) lockOrderFunds(ctx sdk.Context, userAddr sdk.AccAddress, order types.Order) error {
	var lockCoins sdk.Coins

	if order.Side == types.OrderSideBuy {
		// For buy orders, lock quote currency (HODL)
		requiredAmount := order.Price.MulInt(order.Quantity).TruncateInt()
		lockCoins = sdk.NewCoins(sdk.NewCoin(order.QuoteSymbol, requiredAmount))
	} else {
		// For sell orders, lock base currency (equity)
		lockCoins = sdk.NewCoins(sdk.NewCoin(order.BaseSymbol, order.Quantity))

		// Register beneficial owner for dividend tracking
		// When equity shares are locked in a sell order, the seller remains the beneficial owner
		// and should continue to receive dividends until the order fills
		if k.equityKeeper != nil {
			companyID, found := k.GetCompanyIDBySymbol(ctx, order.BaseSymbol)
			if found {
				// Register with "COMMON" as default share class
				// In production, might want to parse class from symbol or add to order
				err := k.equityKeeper.RegisterBeneficialOwner(
					ctx,
					types.ModuleName,    // "dex" module holds the shares
					companyID,
					"COMMON",            // Default share class
					order.User,          // Seller is the true owner
					order.Quantity,
					order.ID,            // Order ID as reference
					"sell_order",        // Reference type
				)
				if err != nil {
					k.Logger(ctx).Error("failed to register beneficial owner for sell order",
						"order_id", order.ID,
						"user", order.User,
						"symbol", order.BaseSymbol,
						"error", err,
					)
					// Don't fail the order placement - beneficial owner tracking is best-effort
				} else {
					k.Logger(ctx).Info("registered beneficial owner for sell order",
						"order_id", order.ID,
						"user", order.User,
						"symbol", order.BaseSymbol,
						"shares", order.Quantity.String(),
					)
				}
			}
		}
	}

	// Transfer to module account
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, userAddr, types.ModuleName, lockCoins)
}

// unlockOrderFunds returns locked funds to user when order is cancelled
func (k Keeper) unlockOrderFunds(ctx sdk.Context, order types.Order) error {
	userAddr, err := sdk.AccAddressFromBech32(order.User)
	if err != nil {
		return err
	}

	var unlockCoins sdk.Coins
	remaining := order.RemainingQuantity
	if remaining.IsNil() {
		remaining = order.Quantity.Sub(order.FilledQuantity)
	}

	if order.Side == types.OrderSideBuy {
		// Return unfilled quote currency
		unlockAmount := order.Price.MulInt(remaining).TruncateInt()
		unlockCoins = sdk.NewCoins(sdk.NewCoin(order.QuoteSymbol, unlockAmount))
	} else {
		// Return unfilled base currency
		unlockCoins = sdk.NewCoins(sdk.NewCoin(order.BaseSymbol, remaining))

		// Unregister beneficial owner when sell order is cancelled/completed
		if k.equityKeeper != nil && remaining.GT(math.ZeroInt()) {
			companyID, found := k.GetCompanyIDBySymbol(ctx, order.BaseSymbol)
			if found {
				err := k.equityKeeper.UnregisterBeneficialOwner(
					ctx,
					types.ModuleName,
					companyID,
					"COMMON",
					order.User,
					order.ID,
				)
				if err != nil {
					k.Logger(ctx).Error("failed to unregister beneficial owner",
						"order_id", order.ID,
						"user", order.User,
						"error", err,
					)
					// Don't fail the unlock - best effort
				} else {
					k.Logger(ctx).Info("unregistered beneficial owner for cancelled/completed order",
						"order_id", order.ID,
						"user", order.User,
					)
				}
			}
		}
	}

	if unlockCoins.IsAllPositive() {
		return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, userAddr, unlockCoins)
	}
	return nil
}

// checkOrderFunds checks if user has sufficient funds for an order
func (k Keeper) checkOrderFunds(ctx sdk.Context, userAddr sdk.AccAddress, order types.Order) error {
	// Get user balance
	balance := k.bankKeeper.GetAllBalances(ctx, userAddr)
	
	// Calculate required amount based on order side
	if order.Side == types.OrderSideBuy {
		// For buy orders, need quote currency
		quoteDenom := order.QuoteSymbol
		if order.Type == types.OrderTypeMarket {
			// For market orders, would need to estimate cost
			// Simplified: assume price from last trade or market data
			requiredAmount := order.Price.MulInt(order.Quantity).TruncateInt()
			if balance.AmountOf(quoteDenom).LT(requiredAmount) {
				return types.ErrInsufficientFunds
			}
		} else {
			// For limit orders, know exact cost
			requiredAmount := order.Price.MulInt(order.Quantity).TruncateInt()
			if balance.AmountOf(quoteDenom).LT(requiredAmount) {
				return types.ErrInsufficientFunds
			}
		}
	} else {
		// For sell orders, need base currency (equity tokens)
		baseDenom := order.BaseSymbol
		if balance.AmountOf(baseDenom).LT(order.Quantity) {
			return types.ErrInsufficientFunds
		}
	}
	
	return nil
}

// calculateTradingFees calculates maker and taker fees for a trade
func (k Keeper) calculateTradingFees(
	ctx sdk.Context,
	market types.Market,
	quantity math.Int,
	price math.LegacyDec,
	isMaker bool,
) math.LegacyDec {
	tradeValue := price.MulInt(quantity)
	
	if isMaker {
		return market.MakerFee.Mul(tradeValue)
	} else {
		return market.TakerFee.Mul(tradeValue)
	}
}

// getAllUserOrders returns all orders for a user
func (k Keeper) getAllUserOrders(ctx sdk.Context, user string) []types.Order {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.OrderPrefix)
	defer iterator.Close()

	var orders []types.Order
	for ; iterator.Valid(); iterator.Next() {
		var order types.Order
		if err := json.Unmarshal(iterator.Value(), &order); err != nil {
			k.Logger(ctx).Error("failed to unmarshal order", "error", err)
			continue
		}
		// Filter by user
		if order.User == user {
			orders = append(orders, order)
		}
	}
	return orders
}

// Market statistics methods (simplified)

// UpdateMarketStats updates market statistics after a trade
func (k Keeper) UpdateMarketStats(ctx sdk.Context, market types.Market, trade types.Trade) {
	// Update last price
	market.LastPrice = trade.Price
	
	// Update 24h volume
	market.Volume24h = market.Volume24h.Add(trade.Quantity)
	
	// Update 24h high/low (simplified)
	if market.High24h.IsZero() || trade.Price.GT(market.High24h) {
		market.High24h = trade.Price
	}
	if market.Low24h.IsZero() || trade.Price.LT(market.Low24h) {
		market.Low24h = trade.Price
	}
	
	// Store updated market
	k.SetMarket(ctx, market)
}

// GetMarketStats returns market statistics
func (k Keeper) GetMarketStats(ctx sdk.Context, baseSymbol, quoteSymbol string) types.MarketStats {
	// Simplified implementation
	return types.MarketStats{
		MarketSymbol: baseSymbol + "/" + quoteSymbol,
		// Would populate with real data in full implementation
	}
}

// ===== BLOCKCHAIN-NATIVE ATOMIC SWAP IMPLEMENTATION =====

// ExecuteAtomicSwap performs instant cross-asset swaps with zero counterparty risk
func (k Keeper) ExecuteAtomicSwap(
	ctx sdk.Context,
	trader sdk.AccAddress,
	fromSymbol, toSymbol string,
	quantity math.Int,
	maxSlippage math.LegacyDec,
) (types.Order, error) {
	// 1. Validate assets exist and are swappable
	if err := k.ValidateSwapAssets(ctx, fromSymbol, toSymbol); err != nil {
		return types.Order{}, err
	}

	// 2. Get real-time exchange rate
	exchangeRate, err := k.GetAtomicSwapRate(ctx, fromSymbol, toSymbol)
	if err != nil {
		return types.Order{}, err
	}

	// 3. Calculate output amount and check slippage
	outputAmount := exchangeRate.MulInt(quantity)
	currentRate := k.GetCurrentMarketRate(ctx, fromSymbol, toSymbol)
	slippage := exchangeRate.Sub(currentRate).Quo(currentRate).Abs()
	
	if slippage.GT(maxSlippage) {
		return types.Order{}, types.ErrSlippageExceeded
	}

	// 4. Lock input tokens
	if err := k.LockAssetForSwap(ctx, trader, fromSymbol, quantity); err != nil {
		return types.Order{}, err
	}

	// 5. Execute atomic swap
	orderID := k.GetNextOrderID(ctx)
	
	// Create atomic swap order
	order := types.Order{
		ID:           orderID,
		MarketSymbol: fromSymbol + "/" + toSymbol,
		BaseSymbol:   fromSymbol,
		QuoteSymbol:  toSymbol,
		User:         trader.String(),
		Side:         types.OrderSideSell, // Selling fromSymbol
		Type:         types.OrderTypeAtomicSwap,
		Status:       types.OrderStatusFilled,
		Quantity:     quantity,
		Price:        exchangeRate,
		FilledQuantity: quantity,
		AveragePrice: exchangeRate,
		TimeInForce:  types.TimeInForceGTC,
		CreatedAt:    ctx.BlockTime(),
		UpdatedAt:    ctx.BlockTime(),
	}

	// 6. Transfer assets atomically
	if err := k.ExecuteSwapTransfer(ctx, trader, fromSymbol, toSymbol, math.LegacyNewDecFromInt(quantity), outputAmount); err != nil {
		// Rollback locked tokens
		k.UnlockAssetForSwap(ctx, trader, fromSymbol, quantity)
		return types.Order{}, err
	}

	// 7. Store order and emit events
	k.SetOrder(ctx, order)
	
	// Create trade record  
	trade := types.Trade{
		ID:           orderID,
		MarketSymbol: fromSymbol + "/" + toSymbol,
		Buyer:        trader.String(),
		Seller:       "atomic_swap_pool",
		BuyOrderID:   orderID,
		SellOrderID:  orderID,
		Quantity:     quantity,
		Price:        exchangeRate,
		Value:        outputAmount,
		BuyerFee:     math.LegacyZeroDec(),
		SellerFee:    math.LegacyZeroDec(),
		BuyerIsMaker: false,
		ExecutedAt:   ctx.BlockTime(),
	}
	k.SetTrade(ctx, trade)

	// 8. Emit atomic swap event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"atomic_swap_executed",
			sdk.NewAttribute("trader", trader.String()),
			sdk.NewAttribute("from_symbol", fromSymbol),
			sdk.NewAttribute("to_symbol", toSymbol),
			sdk.NewAttribute("input_amount", quantity.String()),
			sdk.NewAttribute("output_amount", outputAmount.String()),
			sdk.NewAttribute("exchange_rate", exchangeRate.String()),
			sdk.NewAttribute("slippage", slippage.String()),
			sdk.NewAttribute("order_id", fmt.Sprintf("%d", order.ID)),
		),
	)

	return order, nil
}

// ValidateSwapAssets ensures both assets can be swapped
func (k Keeper) ValidateSwapAssets(ctx sdk.Context, fromSymbol, toSymbol string) error {
	// Check if fromSymbol is a valid equity or HODL
	if fromSymbol == "HODL" {
		// Validate HODL exists
		supply := k.hodlKeeper.GetTotalSupply(ctx)
		if supply == nil {
			return types.ErrAssetNotFound
		}
	} else {
		// Check if company/equity exists
		if !k.equityKeeper.IsSymbolTaken(ctx, fromSymbol) {
			return types.ErrAssetNotFound
		}
	}

	// Check if toSymbol is valid
	if toSymbol == "HODL" {
		supply := k.hodlKeeper.GetTotalSupply(ctx)
		if supply == nil {
			return types.ErrAssetNotFound
		}
	} else {
		if !k.equityKeeper.IsSymbolTaken(ctx, toSymbol) {
			return types.ErrAssetNotFound
		}
	}

	// Ensure they're different assets
	if fromSymbol == toSymbol {
		return types.ErrSameAssetSwap
	}

	return nil
}

// GetAtomicSwapRate calculates real-time exchange rate for atomic swaps
func (k Keeper) GetAtomicSwapRate(ctx sdk.Context, fromSymbol, toSymbol string) (math.LegacyDec, error) {
	// Get market prices for both assets
	fromPrice := k.GetAssetPrice(ctx, fromSymbol)
	toPrice := k.GetAssetPrice(ctx, toSymbol)

	if fromPrice.IsZero() || toPrice.IsZero() {
		return math.LegacyZeroDec(), types.ErrNoMarketPrice
	}

	// Calculate exchange rate: fromPrice / toPrice
	rate := fromPrice.Quo(toPrice)

	// Apply atomic swap fee (governance-controllable, default 0.1%)
	fee := k.GetAtomicSwapFee(ctx)
	rate = rate.Mul(math.LegacyOneDec().Sub(fee))

	return rate, nil
}

// GetCurrentMarketRate gets current market rate for slippage calculation
func (k Keeper) GetCurrentMarketRate(ctx sdk.Context, fromSymbol, toSymbol string) math.LegacyDec {
	// Use order book mid-price as reference
	orderBook := k.GetOrderBook(ctx, fromSymbol, toSymbol)
	// Check if order book has data
	if len(orderBook.Bids) == 0 && len(orderBook.Asks) == 0 {
		// Fallback to last trade price
		return k.GetLastTradePrice(ctx, fromSymbol, toSymbol)
	}

	if len(orderBook.Bids) > 0 && len(orderBook.Asks) > 0 {
		bestBid := orderBook.Bids[0].Price
		bestAsk := orderBook.Asks[0].Price
		return bestBid.Add(bestAsk).QuoInt64(2) // Mid price
	}

	return k.GetLastTradePrice(ctx, fromSymbol, toSymbol)
}

// GetAssetPrice gets current price for an asset in HODL terms
func (k Keeper) GetAssetPrice(ctx sdk.Context, symbol string) math.LegacyDec {
	if symbol == "HODL" {
		return math.LegacyOneDec() // HODL = 1.0 (stable)
	}

	// For equity assets, get market price
	market, found := k.GetMarket(ctx, symbol, "HODL")
	if found && !market.LastPrice.IsZero() {
		return market.LastPrice
	}

	// Fallback to last trade price
	return k.GetLastTradePrice(ctx, symbol, "HODL")
}

// GetLastTradePrice gets the last trade price for an asset
func (k Keeper) GetLastTradePrice(ctx sdk.Context, symbol, quoteSymbol string) math.LegacyDec {
	// Simplified implementation - in a full implementation this would
	// search through stored trades to find the most recent one
	// For now, check market last price first
	market, found := k.GetMarket(ctx, symbol, quoteSymbol)
	if found && !market.LastPrice.IsZero() {
		return market.LastPrice
	}

	// Default to $1 HODL if no trades exist
	return math.LegacyOneDec()
}

// LockAssetForSwap temporarily locks asset during swap
func (k Keeper) LockAssetForSwap(ctx sdk.Context, trader sdk.AccAddress, symbol string, amount math.Int) error {
	if symbol == "HODL" {
		// Lock HODL tokens
		hodlCoins := sdk.NewCoins(sdk.NewCoin("hodl", amount))
		return k.bankKeeper.SendCoinsFromAccountToModule(ctx, trader, types.ModuleName, hodlCoins)
	}

	// Lock equity shares
	// 1. Get company ID from symbol for beneficial owner tracking
	companyID, found := k.equityKeeper.GetCompanyBySymbol(ctx, symbol)
	if !found {
		return types.ErrAssetNotFound
	}

	// 2. Transfer equity tokens to module account (this locks them)
	equityCoins := sdk.NewCoins(sdk.NewCoin(symbol, amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, trader, types.ModuleName, equityCoins); err != nil {
		k.Logger(ctx).Error("failed to lock equity shares",
			"trader", trader.String(),
			"symbol", symbol,
			"amount", amount.String(),
			"error", err,
		)
		return err
	}

	// 3. Store locked equity record for tracking
	store := ctx.KVStore(k.storeKey)
	key := types.GetLockedEquityKey(trader, symbol)

	// Get existing locked amount if any
	var existingLocked math.Int
	bz := store.Get(key)
	if bz != nil {
		existingLocked = math.NewIntFromUint64(sdk.BigEndianToUint64(bz))
	} else {
		existingLocked = math.ZeroInt()
	}

	// Add to locked amount
	newLocked := existingLocked.Add(amount)
	store.Set(key, sdk.Uint64ToBigEndian(newLocked.Uint64()))

	// 4. Register beneficial owner for dividend tracking
	// Even though shares are locked for a swap, the trader remains beneficial owner
	// until the swap completes
	if k.equityKeeper != nil {
		classID := "COMMON" // Default share class

		// Use block height as reference ID for atomic swaps
		// In production, might want a more unique identifier
		swapRefID := uint64(ctx.BlockHeight())

		err := k.equityKeeper.RegisterBeneficialOwner(
			ctx,
			types.ModuleName,      // "dex" module holds the shares
			companyID,
			classID,
			trader.String(),       // Trader is the beneficial owner
			amount,
			swapRefID,
			"atomic_swap",         // Reference type
		)
		if err != nil {
			k.Logger(ctx).Error("failed to register beneficial owner for atomic swap",
				"trader", trader.String(),
				"symbol", symbol,
				"company_id", companyID,
				"amount", amount.String(),
				"error", err,
			)
			// Don't fail the lock - beneficial owner tracking is best-effort
			// They've already transferred the shares to module account
		} else {
			k.Logger(ctx).Info("registered beneficial owner for atomic swap",
				"trader", trader.String(),
				"symbol", symbol,
				"company_id", companyID,
				"amount", amount.String(),
			)
		}
	}

	k.Logger(ctx).Info("locked equity shares for atomic swap",
		"trader", trader.String(),
		"symbol", symbol,
		"amount", amount.String(),
		"total_locked", newLocked.String(),
	)

	return nil
}

// UnlockAssetForSwap unlocks asset if swap fails
func (k Keeper) UnlockAssetForSwap(ctx sdk.Context, trader sdk.AccAddress, symbol string, amount math.Int) {
	if symbol == "HODL" {
		// Return HODL tokens
		hodlCoins := sdk.NewCoins(sdk.NewCoin("hodl", amount))
		k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, trader, hodlCoins)
		return
	}

	// Unlock equity shares
	// 1. Return equity tokens to trader
	equityCoins := sdk.NewCoins(sdk.NewCoin(symbol, amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, trader, equityCoins); err != nil {
		k.Logger(ctx).Error("failed to return equity shares to trader",
			"symbol", symbol,
			"trader", trader.String(),
			"amount", amount.String(),
			"error", err,
		)
		// Continue with cleanup even if transfer failed
	}

	// 2. Get company ID from symbol for beneficial owner cleanup
	companyID, found := k.equityKeeper.GetCompanyBySymbol(ctx, symbol)
	if !found {
		k.Logger(ctx).Error("failed to unlock equity - company not found",
			"symbol", symbol,
			"trader", trader.String(),
		)
		// Continue with record cleanup
	}

	// 3. Update locked equity record
	store := ctx.KVStore(k.storeKey)
	key := types.GetLockedEquityKey(trader, symbol)

	bz := store.Get(key)
	if bz == nil {
		k.Logger(ctx).Error("no locked equity record found",
			"symbol", symbol,
			"trader", trader.String(),
		)
		// Continue anyway
	} else {
		existingLocked := math.NewIntFromUint64(sdk.BigEndianToUint64(bz))
		if existingLocked.LT(amount) {
			k.Logger(ctx).Error("attempting to unlock more than locked",
				"symbol", symbol,
				"trader", trader.String(),
				"locked", existingLocked.String(),
				"unlock_amount", amount.String(),
			)
			amount = existingLocked // Unlock all available
		}

		// Subtract from locked amount
		newLocked := existingLocked.Sub(amount)
		if newLocked.IsZero() {
			// Remove lock record if no longer locked
			store.Delete(key)
		} else {
			store.Set(key, sdk.Uint64ToBigEndian(newLocked.Uint64()))
		}

		k.Logger(ctx).Info("updated locked equity record",
			"trader", trader.String(),
			"symbol", symbol,
			"unlocked_amount", amount.String(),
			"remaining_locked", newLocked.String(),
		)
	}

	// 4. Unregister beneficial owner
	if k.equityKeeper != nil && found {
		classID := "COMMON"
		swapRefID := uint64(ctx.BlockHeight())

		err := k.equityKeeper.UnregisterBeneficialOwner(
			ctx,
			types.ModuleName,
			companyID,
			classID,
			trader.String(),
			swapRefID,
		)
		if err != nil {
			k.Logger(ctx).Error("failed to unregister beneficial owner",
				"trader", trader.String(),
				"symbol", symbol,
				"error", err,
			)
			// Don't fail - best effort
		} else {
			k.Logger(ctx).Info("unregistered beneficial owner for failed atomic swap",
				"trader", trader.String(),
				"symbol", symbol,
				"amount", amount.String(),
			)
		}
	}

	k.Logger(ctx).Info("unlocked equity shares",
		"trader", trader.String(),
		"symbol", symbol,
		"amount", amount.String(),
	)
}

// ===== AUTHORIZATION METHODS =====

// IsAuthorizedMarketCreator checks if an address is authorized to create markets
// Markets can be created by:
// 1. Module governance authority
// 2. Addresses in the authorized market creators whitelist (set via governance)
func (k Keeper) IsAuthorizedMarketCreator(ctx sdk.Context, addr sdk.AccAddress) bool {
	// Check if address is in the whitelist
	store := ctx.KVStore(k.storeKey)
	key := types.GetAuthorizedMarketCreatorKey(addr)
	return store.Has(key)
}

// SetAuthorizedMarketCreator adds or removes an address from the market creator whitelist
// This should only be called via governance proposals
func (k Keeper) SetAuthorizedMarketCreator(ctx sdk.Context, addr sdk.AccAddress, authorized bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAuthorizedMarketCreatorKey(addr)

	if authorized {
		store.Set(key, []byte{1})
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"market_creator_authorized",
				sdk.NewAttribute("address", addr.String()),
			),
		)
	} else {
		store.Delete(key)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"market_creator_revoked",
				sdk.NewAttribute("address", addr.String()),
			),
		)
	}
}

// GetAllAuthorizedMarketCreators returns all authorized market creator addresses
func (k Keeper) GetAllAuthorizedMarketCreators(ctx sdk.Context) []sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.AuthorizedMarketCreatorPrefix)
	defer iterator.Close()

	var creators []sdk.AccAddress
	for ; iterator.Valid(); iterator.Next() {
		// Key format: prefix + address bytes
		addrBytes := iterator.Key()[len(types.AuthorizedMarketCreatorPrefix):]
		creators = append(creators, sdk.AccAddress(addrBytes))
	}
	return creators
}

// GetCompanyIDBySymbol returns company ID for a trading symbol
// Wrapper around equity keeper's GetCompanyBySymbol for convenience
func (k Keeper) GetCompanyIDBySymbol(ctx sdk.Context, symbol string) (uint64, bool) {
	if k.equityKeeper == nil {
		return 0, false
	}
	return k.equityKeeper.GetCompanyBySymbol(ctx, symbol)
}

// CancelOrdersForCompany cancels all open orders for a company's equity
// This should be called by the equity module when a trading halt is initiated
func (k Keeper) CancelOrdersForCompany(ctx sdk.Context, companyID uint64, symbol string) error {
	if symbol == "" {
		k.Logger(ctx).Error("cannot cancel orders - empty symbol", "company_id", companyID)
		return fmt.Errorf("empty symbol for company %d", companyID)
	}

	// Find all open orders for this company's equity trading pairs
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.OrderPrefix)
	defer iterator.Close()

	cancelledCount := 0
	var errors []error

	for ; iterator.Valid(); iterator.Next() {
		var order types.Order
		if err := json.Unmarshal(iterator.Value(), &order); err != nil {
			k.Logger(ctx).Error("failed to unmarshal order during halt cancellation", "error", err)
			continue
		}

		// Check if this order is for the halted company's equity
		if (order.BaseSymbol == symbol || order.QuoteSymbol == symbol) && order.IsFillable() {
			// Release locked funds
			if err := k.unlockOrderFunds(ctx, order); err != nil {
				k.Logger(ctx).Error("failed to unlock funds during halt cancellation",
					"order_id", order.ID,
					"error", err,
				)
				errors = append(errors, err)
				continue
			}

			// Update order status to cancelled
			order.Status = types.OrderStatusCancelled
			order.UpdatedAt = ctx.BlockTime()
			if err := k.SetOrder(ctx, order); err != nil {
				k.Logger(ctx).Error("failed to save cancelled order",
					"order_id", order.ID,
					"error", err,
				)
				errors = append(errors, err)
				continue
			}

			// Emit cancellation event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeOrderCancelled,
					sdk.NewAttribute("order_id", fmt.Sprintf("%d", order.ID)),
					sdk.NewAttribute("user", order.User),
					sdk.NewAttribute("market", order.MarketSymbol),
					sdk.NewAttribute("reason", "trading_halted"),
					sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyID)),
				),
			)

			cancelledCount++
		}
	}

	k.Logger(ctx).Info("cancelled orders for halted company",
		"company_id", companyID,
		"symbol", symbol,
		"orders_cancelled", cancelledCount,
		"errors", len(errors),
	)

	// Emit summary event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"trading_halt_orders_cancelled",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute("symbol", symbol),
			sdk.NewAttribute("orders_cancelled", fmt.Sprintf("%d", cancelledCount)),
			sdk.NewAttribute("errors", fmt.Sprintf("%d", len(errors))),
		),
	)

	if len(errors) > 0 {
		return fmt.Errorf("cancelled %d orders with %d errors", cancelledCount, len(errors))
	}

	return nil
}

// ExecuteSwapTransfer performs the actual asset transfer
func (k Keeper) ExecuteSwapTransfer(
	ctx sdk.Context,
	trader sdk.AccAddress,
	fromSymbol, toSymbol string,
	inputAmount, outputAmount math.LegacyDec,
) error {
	// Convert to Int for coin operations
	inputInt := inputAmount.TruncateInt()
	outputInt := outputAmount.TruncateInt()

	if fromSymbol == "HODL" && toSymbol != "HODL" {
		// HODL → Equity: Burn HODL, mint equity tokens
		hodlCoins := sdk.NewCoins(sdk.NewCoin("hodl", inputInt))
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, hodlCoins); err != nil {
			return err
		}

		// Mint equity tokens (simplified - in production would transfer actual shares)
		equityCoins := sdk.NewCoins(sdk.NewCoin(toSymbol, outputInt))
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, equityCoins); err != nil {
			return err
		}
		return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, trader, equityCoins)

	} else if fromSymbol != "HODL" && toSymbol == "HODL" {
		// Equity → HODL: Burn equity tokens, mint HODL
		equityCoins := sdk.NewCoins(sdk.NewCoin(fromSymbol, inputInt))
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, equityCoins); err != nil {
			return err
		}

		// Mint HODL tokens
		hodlCoins := sdk.NewCoins(sdk.NewCoin("hodl", outputInt))
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, hodlCoins); err != nil {
			return err
		}
		return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, trader, hodlCoins)

	} else {
		// Equity → Equity: Direct cross-equity swap
		// Use bank module to handle equity tokens as coins
		// In the Cosmos SDK, equity shares are represented as bank denoms

		// Verify both assets exist
		if !k.equityKeeper.IsSymbolTaken(ctx, fromSymbol) {
			return types.ErrAssetNotFound
		}
		if !k.equityKeeper.IsSymbolTaken(ctx, toSymbol) {
			return types.ErrAssetNotFound
		}

		// 1. Burn the FROM equity tokens from module account
		fromCoins := sdk.NewCoins(sdk.NewCoin(fromSymbol, inputInt))
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, fromCoins); err != nil {
			k.Logger(ctx).Error("failed to burn FROM equity",
				"from_symbol", fromSymbol,
				"amount", inputInt.String(),
				"error", err,
			)
			return err
		}

		// 2. Mint the TO equity tokens to module account
		toCoins := sdk.NewCoins(sdk.NewCoin(toSymbol, outputInt))
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, toCoins); err != nil {
			k.Logger(ctx).Error("failed to mint TO equity",
				"to_symbol", toSymbol,
				"amount", outputInt.String(),
				"error", err,
			)
			// Rollback: re-mint FROM tokens
			k.bankKeeper.MintCoins(ctx, types.ModuleName, fromCoins)
			return err
		}

		// 3. Transfer TO equity to trader
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, trader, toCoins); err != nil {
			k.Logger(ctx).Error("failed to send TO equity to trader",
				"to_symbol", toSymbol,
				"trader", trader.String(),
				"amount", outputInt.String(),
				"error", err,
			)
			// Rollback: burn TO tokens and re-mint FROM tokens
			k.bankKeeper.BurnCoins(ctx, types.ModuleName, toCoins)
			k.bankKeeper.MintCoins(ctx, types.ModuleName, fromCoins)
			return err
		}

		k.Logger(ctx).Info("executed cross-equity atomic swap",
			"trader", trader.String(),
			"from_symbol", fromSymbol,
			"to_symbol", toSymbol,
			"input_amount", inputInt.String(),
			"output_amount", outputInt.String(),
		)

		return nil
	}
}