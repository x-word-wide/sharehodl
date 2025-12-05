package keeper

import (
	"encoding/json"
	"fmt"

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

// GetMarket returns a market by symbol
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

// SetMarket stores a market
func (k Keeper) SetMarket(ctx sdk.Context, market types.Market) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetMarketKey(market.BaseSymbol, market.QuoteSymbol)
	bz, err := json.Marshal(market)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
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
func (k Keeper) SetOrder(ctx sdk.Context, order types.Order) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetOrderKey(order.ID)
	bz, err := json.Marshal(order)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// DeleteOrder removes an order
func (k Keeper) DeleteOrder(ctx sdk.Context, orderID uint64) {
	store := ctx.KVStore(k.storeKey)
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
func (k Keeper) SetTrade(ctx sdk.Context, trade types.Trade) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradeKey(trade.ID)
	bz, err := json.Marshal(trade)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// Order book management

// GetOrderBook returns the order book for a market
func (k Keeper) GetOrderBook(ctx sdk.Context, baseSymbol, quoteSymbol string) types.OrderBook {
	// For simplified implementation, return empty order book
	// In full implementation, would aggregate all open orders
	return types.OrderBook{
		MarketSymbol: baseSymbol + "/" + quoteSymbol,
		Bids:        []types.OrderBookLevel{},
		Asks:        []types.OrderBookLevel{},
	}
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
	k.SetMarket(ctx, market)
	return nil
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
	
	// Store order
	k.SetOrder(ctx, order)
	
	// For simplified implementation, return order without matching
	// In full implementation would match against existing orders
	
	return order, nil
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
	
	// Update order status
	order.Status = types.OrderStatusCancelled
	k.SetOrder(ctx, order)
	
	// In full implementation would release locked funds
	
	return nil
}

// MatchOrders performs order matching (simplified implementation)
func (k Keeper) MatchOrders(ctx sdk.Context, newOrder types.Order) []types.Trade {
	// Simplified matching - return empty trades
	// Full implementation would:
	// 1. Find matching orders from opposite side
	// 2. Execute trades at appropriate prices
	// 3. Update order quantities and status
	// 4. Transfer assets between users
	// 5. Calculate and collect fees
	
	return []types.Trade{}
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
func (k Keeper) SetLiquidityPool(ctx sdk.Context, pool types.LiquidityPool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLiquidityPoolKey(pool.MarketSymbol, "")
	bz, err := json.Marshal(pool)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
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

// getAllUserOrders returns all orders for a user (simplified implementation)
func (k Keeper) getAllUserOrders(ctx sdk.Context, user string) []types.Order {
	// Simplified - return empty list
	// Full implementation would iterate through user's orders
	return []types.Order{}
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

	// Apply atomic swap fee (0.1%)
	fee := math.LegacyNewDecWithPrec(1, 3) // 0.001 = 0.1%
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
	} else {
		// Lock equity shares - check if trader owns enough
		shareholding, found := k.equityKeeper.GetShareholding(ctx, 0, symbol, trader.String()) // Simplified company ID
		if !found {
			return types.ErrInsufficientShares
		}
		// Note: shareholding type checking would be needed in full implementation
		_ = shareholding // Mark as used for now

		// In production, would track locked shares separately
		// For now, transfer to module account
		return types.ErrNotImplemented // TODO: Implement equity locking
	}
}

// UnlockAssetForSwap unlocks asset if swap fails
func (k Keeper) UnlockAssetForSwap(ctx sdk.Context, trader sdk.AccAddress, symbol string, amount math.Int) {
	if symbol == "HODL" {
		// Return HODL tokens
		hodlCoins := sdk.NewCoins(sdk.NewCoin("hodl", amount))
		k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, trader, hodlCoins)
	}
	// TODO: Handle equity unlocking
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
		// Equity → Equity: Cross-equity swap via HODL
		// TODO: Implement direct equity-to-equity swaps
		return types.ErrCrossEquitySwapNotImplemented
	}
}