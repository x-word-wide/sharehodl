package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/dex/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) *msgServer {
	return &msgServer{Keeper: keeper}
}

// CreateMarket handles market creation
func (k msgServer) CreateMarket(goCtx context.Context, msg *types.SimpleMsgCreateMarket) (*types.MsgCreateMarketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// SECURITY FIX: Only authorized addresses can create markets
	// This prevents fake markets and symbol spoofing attacks
	if !k.Keeper.IsAuthorizedMarketCreator(ctx, creatorAddr) {
		return nil, errors.Wrap(types.ErrUnauthorized, "only authorized addresses can create markets; submit a governance proposal to request market creation authorization")
	}

	// Create market
	err = k.Keeper.CreateMarket(
		ctx,
		msg.BaseSymbol,
		msg.QuoteSymbol,
		msg.BaseAssetID,
		msg.QuoteAssetID,
		msg.MinOrderSize,
		msg.MaxOrderSize,
		msg.TickSize,
		msg.LotSize,
		msg.MakerFee,
		msg.TakerFee,
	)
	if err != nil {
		return nil, err
	}

	marketSymbol := msg.BaseSymbol + "/" + msg.QuoteSymbol

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"create_market",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("market_symbol", marketSymbol),
			sdk.NewAttribute("base_symbol", msg.BaseSymbol),
			sdk.NewAttribute("quote_symbol", msg.QuoteSymbol),
		),
	)

	return &types.MsgCreateMarketResponse{
		MarketSymbol: marketSymbol,
		Success:      true,
	}, nil
}

// PlaceOrder handles order placement
func (k msgServer) PlaceOrder(goCtx context.Context, msg *types.SimpleMsgPlaceOrder) (*types.MsgPlaceOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Parse order side
	var side types.OrderSide
	switch strings.ToLower(msg.Side) {
	case "buy":
		side = types.OrderSideBuy
	case "sell":
		side = types.OrderSideSell
	default:
		return nil, types.ErrInvalidOrderSide
	}

	// Parse order type
	var orderType types.OrderType
	switch strings.ToLower(msg.OrderType) {
	case "market":
		orderType = types.OrderTypeMarket
	case "limit":
		orderType = types.OrderTypeLimit
	case "stop":
		orderType = types.OrderTypeStop
	case "stop_limit":
		orderType = types.OrderTypeStopLimit
	default:
		return nil, types.ErrInvalidOrderType
	}

	// Parse time in force
	var timeInForce types.TimeInForce
	switch strings.ToLower(msg.TimeInForce) {
	case "gtc":
		timeInForce = types.TimeInForceGTC
	case "ioc":
		timeInForce = types.TimeInForceIOC
	case "fok":
		timeInForce = types.TimeInForceFOK
	case "gtd":
		timeInForce = types.TimeInForceGTD
	default:
		return nil, types.ErrInvalidTimeInForce
	}

	// Place order
	order, err := k.Keeper.PlaceOrder(
		ctx,
		msg.Creator,
		msg.MarketSymbol,
		side,
		orderType,
		timeInForce,
		msg.Quantity,
		msg.Price,
		msg.StopPrice,
		msg.ClientOrderID,
	)
	if err != nil {
		return nil, err
	}

	// Attempt to match order (simplified)
	trades := k.Keeper.MatchOrders(ctx, order)

	// Update order with any fills
	filledQuantity := math.ZeroInt()
	totalValue := math.LegacyZeroDec()
	totalFees := math.LegacyZeroDec()

	for _, trade := range trades {
		filledQuantity = filledQuantity.Add(trade.Quantity)
		totalValue = totalValue.Add(trade.Value)
		if trade.Buyer == order.User {
			totalFees = totalFees.Add(trade.BuyerFee)
		} else {
			totalFees = totalFees.Add(trade.SellerFee)
		}
	}

	// Update order status
	if filledQuantity.GT(math.ZeroInt()) {
		order.FilledQuantity = filledQuantity
		order.RemainingQuantity = order.Quantity.Sub(filledQuantity)
		order.TotalFees = totalFees

		if order.RemainingQuantity.IsZero() {
			order.Status = types.OrderStatusFilled
		} else {
			order.Status = types.OrderStatusPartiallyFilled
		}

		if !totalValue.IsZero() {
			order.AveragePrice = totalValue.QuoInt(filledQuantity)
		}

		k.Keeper.SetOrder(ctx, order)
	}

	// Store trades
	for _, trade := range trades {
		k.Keeper.SetTrade(ctx, trade)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"place_order",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("order_id", fmt.Sprintf("%d", order.ID)),
			sdk.NewAttribute("market_symbol", msg.MarketSymbol),
			sdk.NewAttribute("side", msg.Side),
			sdk.NewAttribute("order_type", msg.OrderType),
			sdk.NewAttribute("quantity", msg.Quantity.String()),
			sdk.NewAttribute("price", msg.Price.String()),
			sdk.NewAttribute("status", order.Status.String()),
			sdk.NewAttribute("filled_quantity", order.FilledQuantity.String()),
		),
	)

	return &types.MsgPlaceOrderResponse{
		OrderID:           order.ID,
		Status:            order.Status.String(),
		FilledQuantity:    order.FilledQuantity,
		RemainingQuantity: order.RemainingQuantity,
		AveragePrice:      order.AveragePrice,
		TotalFees:         order.TotalFees,
	}, nil
}

// CancelOrder handles order cancellation
func (k msgServer) CancelOrder(goCtx context.Context, msg *types.SimpleMsgCancelOrder) (*types.MsgCancelOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Cancel order
	err = k.Keeper.CancelOrder(ctx, creatorAddr, msg.OrderID)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"cancel_order",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("order_id", fmt.Sprintf("%d", msg.OrderID)),
		),
	)

	return &types.MsgCancelOrderResponse{
		OrderID: msg.OrderID,
		Success: true,
	}, nil
}

// CancelAllOrders handles bulk order cancellation
func (k msgServer) CancelAllOrders(goCtx context.Context, msg *types.SimpleMsgCancelAllOrders) (*types.MsgCancelAllOrdersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Get all user orders
	orders := k.Keeper.getAllUserOrders(ctx, msg.Creator)
	
	cancelledCount := uint64(0)
	for _, order := range orders {
		// Filter by market if specified
		if msg.MarketSymbol != "" && order.MarketSymbol != msg.MarketSymbol {
			continue
		}

		// Cancel order
		if err := k.Keeper.CancelOrder(ctx, creatorAddr, order.ID); err == nil {
			cancelledCount++
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"cancel_all_orders",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("market_symbol", msg.MarketSymbol),
			sdk.NewAttribute("orders_cancelled", fmt.Sprintf("%d", cancelledCount)),
		),
	)

	return &types.MsgCancelAllOrdersResponse{
		OrdersCancelled: cancelledCount,
		Success:         true,
	}, nil
}

// CreateLiquidityPool handles liquidity pool creation
func (k msgServer) CreateLiquidityPool(goCtx context.Context, msg *types.SimpleMsgCreateLiquidityPool) (*types.MsgCreateLiquidityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Check if pool already exists
	baseSymbol, quoteSymbol := k.Keeper.parseMarketSymbol(msg.MarketSymbol)
	if _, exists := k.Keeper.GetLiquidityPool(ctx, baseSymbol, quoteSymbol); exists {
		return nil, types.ErrPoolAlreadyExists
	}

	// Check user has sufficient funds
	balance := k.bankKeeper.GetAllBalances(ctx, creatorAddr)
	
	if balance.AmountOf(baseSymbol).LT(msg.BaseAmount) || balance.AmountOf(quoteSymbol).LT(msg.QuoteAmount) {
		return nil, types.ErrInsufficientFunds
	}

	// Transfer funds to module (simplified)
	// In full implementation would transfer to pool contract

	// Calculate initial LP tokens (simplified)
	lpTokenSupply := msg.BaseAmount.Add(msg.QuoteAmount) // Simplified calculation

	// Create pool
	pool := types.LiquidityPool{
		MarketSymbol:     msg.MarketSymbol,
		BaseReserve:      msg.BaseAmount,
		QuoteReserve:     msg.QuoteAmount,
		LPTokenSupply:    lpTokenSupply,
		Fee:              msg.Fee,
		Volume24h:        math.ZeroInt(),
		FeesCollected24h: math.LegacyZeroDec(),
	}

	if err := k.Keeper.SetLiquidityPool(ctx, pool); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"create_liquidity_pool",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("market_symbol", msg.MarketSymbol),
			sdk.NewAttribute("base_amount", msg.BaseAmount.String()),
			sdk.NewAttribute("quote_amount", msg.QuoteAmount.String()),
			sdk.NewAttribute("lp_tokens_issued", lpTokenSupply.String()),
		),
	)

	return &types.MsgCreateLiquidityPoolResponse{
		PoolAddress:    "pool_" + msg.MarketSymbol, // Simplified
		LPTokensIssued: lpTokenSupply,
		Success:        true,
	}, nil
}

// AddLiquidity handles liquidity addition to AMM pools
func (k msgServer) AddLiquidity(goCtx context.Context, msg *types.SimpleMsgAddLiquidity) (*types.MsgAddLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Get market to find base/quote assets
	market, found := k.GetMarketBySymbol(ctx, msg.MarketSymbol)
	if !found {
		return nil, errors.Wrapf(types.ErrInvalidMarket, "market %s not found", msg.MarketSymbol)
	}

	// Get the liquidity pool
	pool, found := k.GetLiquidityPool(ctx, market.BaseSymbol, market.QuoteSymbol)
	if !found {
		return nil, errors.Wrapf(types.ErrPoolNotFound, "liquidity pool for %s not found", msg.MarketSymbol)
	}

	// Check user has sufficient funds
	balance := k.bankKeeper.GetAllBalances(ctx, creatorAddr)
	if balance.AmountOf(market.BaseSymbol).LT(msg.BaseAmount) {
		return nil, errors.Wrapf(types.ErrInsufficientFunds, "insufficient %s balance", market.BaseSymbol)
	}
	if balance.AmountOf(market.QuoteSymbol).LT(msg.QuoteAmount) {
		return nil, errors.Wrapf(types.ErrInsufficientFunds, "insufficient %s balance", market.QuoteSymbol)
	}

	// Calculate LP tokens to mint based on current pool ratio
	var lpTokensToMint math.Int
	if pool.BaseReserve.IsZero() || pool.QuoteReserve.IsZero() {
		// First liquidity provider - use geometric mean approximation
		// sqrt(base * quote) approximated as (base + quote) / 2 for simplicity
		lpTokensToMint = msg.BaseAmount.Add(msg.QuoteAmount).Quo(math.NewInt(2))
	} else {
		// Calculate proportional LP tokens using minimum ratio
		baseRatio := math.LegacyNewDecFromInt(msg.BaseAmount).Quo(math.LegacyNewDecFromInt(pool.BaseReserve))
		quoteRatio := math.LegacyNewDecFromInt(msg.QuoteAmount).Quo(math.LegacyNewDecFromInt(pool.QuoteReserve))

		var minRatio math.LegacyDec
		if baseRatio.LT(quoteRatio) {
			minRatio = baseRatio
		} else {
			minRatio = quoteRatio
		}
		lpTokensToMint = minRatio.MulInt(pool.LPTokenSupply).TruncateInt()
	}

	// Check minimum LP tokens received
	if lpTokensToMint.LT(msg.MinLPTokens) {
		return nil, errors.Wrapf(types.ErrSlippageExceeded, "LP tokens %s below minimum %s", lpTokensToMint.String(), msg.MinLPTokens.String())
	}

	// Transfer funds from provider to module
	coins := sdk.NewCoins(
		sdk.NewCoin(market.BaseSymbol, msg.BaseAmount),
		sdk.NewCoin(market.QuoteSymbol, msg.QuoteAmount),
	)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, coins); err != nil {
		return nil, errors.Wrap(err, "failed to transfer funds to pool")
	}

	// MINT LP TOKENS TO USER'S WALLET
	// LP token denom format: lp/{base}-{quote} (e.g., "lp/APPLE-HODL")
	lpTokenDenom := types.GetLPTokenDenom(market.BaseSymbol, market.QuoteSymbol)
	lpCoins := sdk.NewCoins(sdk.NewCoin(lpTokenDenom, lpTokensToMint))

	// Mint LP tokens to module first
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, lpCoins); err != nil {
		// Rollback: return deposited funds
		k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, coins)
		return nil, errors.Wrap(err, "failed to mint LP tokens")
	}

	// Send LP tokens from module to user
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, lpCoins); err != nil {
		// Rollback: burn minted tokens and return deposited funds
		k.bankKeeper.BurnCoins(ctx, types.ModuleName, lpCoins)
		k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, coins)
		return nil, errors.Wrap(err, "failed to send LP tokens to user")
	}

	// Update pool reserves
	pool.BaseReserve = pool.BaseReserve.Add(msg.BaseAmount)
	pool.QuoteReserve = pool.QuoteReserve.Add(msg.QuoteAmount)
	pool.LPTokenSupply = pool.LPTokenSupply.Add(lpTokensToMint)
	pool.UpdatedAt = ctx.BlockTime()
	k.SetLiquidityPool(ctx, pool)

	// REGISTER BENEFICIAL OWNERSHIP FOR EQUITY IN LIQUIDITY POOL
	// This ensures the LP provider can still:
	// 1. Vote on company matters
	// 2. Receive dividends
	// 3. Maintain their shareholder rights
	if k.equityKeeper != nil {
		companyID, found := k.equityKeeper.GetCompanyBySymbol(ctx, market.BaseSymbol)
		if found {
			// Register as beneficial owner - shares are in DEX module but user owns them
			lpPositionID := k.getNextLPPositionID(ctx)
			err := k.equityKeeper.RegisterBeneficialOwner(
				ctx,
				types.ModuleName,          // Module account holding the equity
				companyID,                 // Company ID
				"common",                  // Default share class
				msg.Creator,               // Beneficial owner (LP provider)
				msg.BaseAmount,            // Number of shares
				lpPositionID,              // Reference ID (LP position)
				"liquidity_provision",     // Reference type
			)
			if err != nil {
				k.Logger(ctx).Error("failed to register beneficial owner for LP",
					"provider", msg.Creator,
					"company_id", companyID,
					"shares", msg.BaseAmount.String(),
					"error", err,
				)
			} else {
				// Store LP position for tracking
				k.SetLPPosition(ctx, types.LPPosition{
					ID:              lpPositionID,
					Provider:        msg.Creator,
					MarketSymbol:    msg.MarketSymbol,
					BaseAmount:      msg.BaseAmount,
					QuoteAmount:     msg.QuoteAmount,
					LPTokenAmount:   lpTokensToMint,
					CompanyID:       companyID,
					CreatedAt:       ctx.BlockTime(),
				})
			}
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"add_liquidity",
			sdk.NewAttribute("market_symbol", msg.MarketSymbol),
			sdk.NewAttribute("provider", msg.Creator),
			sdk.NewAttribute("base_amount", msg.BaseAmount.String()),
			sdk.NewAttribute("quote_amount", msg.QuoteAmount.String()),
			sdk.NewAttribute("lp_tokens_minted", lpTokensToMint.String()),
			sdk.NewAttribute("lp_token_denom", lpTokenDenom),
			sdk.NewAttribute("beneficial_owner_registered", "true"),
		),
	)

	return &types.MsgAddLiquidityResponse{
		LPTokensReceived: lpTokensToMint,
		BaseUsed:         msg.BaseAmount,
		QuoteUsed:        msg.QuoteAmount,
		Success:          true,
	}, nil
}

// RemoveLiquidity handles liquidity removal from AMM pools
func (k msgServer) RemoveLiquidity(goCtx context.Context, msg *types.SimpleMsgRemoveLiquidity) (*types.MsgRemoveLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Get market to find base/quote assets
	market, found := k.GetMarketBySymbol(ctx, msg.MarketSymbol)
	if !found {
		return nil, errors.Wrapf(types.ErrInvalidMarket, "market %s not found", msg.MarketSymbol)
	}

	// Get the liquidity pool
	pool, found := k.GetLiquidityPool(ctx, market.BaseSymbol, market.QuoteSymbol)
	if !found {
		return nil, errors.Wrapf(types.ErrPoolNotFound, "liquidity pool for %s not found", msg.MarketSymbol)
	}

	// Validate LP tokens amount
	if msg.LPTokenAmount.IsNil() || msg.LPTokenAmount.IsZero() {
		return nil, errors.Wrap(types.ErrInvalidOrderSize, "LP tokens amount must be positive")
	}

	if pool.LPTokenSupply.IsZero() {
		return nil, errors.Wrap(types.ErrInsufficientFunds, "pool has no LP tokens")
	}

	if msg.LPTokenAmount.GT(pool.LPTokenSupply) {
		return nil, errors.Wrap(types.ErrInsufficientFunds, "LP tokens exceed total supply")
	}

	// VERIFY USER OWNS LP TOKENS
	lpTokenDenom := types.GetLPTokenDenom(market.BaseSymbol, market.QuoteSymbol)
	userLPBalance := k.bankKeeper.GetBalance(ctx, creatorAddr, lpTokenDenom)
	if userLPBalance.Amount.LT(msg.LPTokenAmount) {
		return nil, errors.Wrapf(types.ErrInsufficientFunds,
			"insufficient LP tokens: have %s, need %s",
			userLPBalance.Amount.String(), msg.LPTokenAmount.String())
	}

	// Calculate proportional amounts to return
	lpRatio := math.LegacyNewDecFromInt(msg.LPTokenAmount).Quo(math.LegacyNewDecFromInt(pool.LPTokenSupply))
	baseReceived := lpRatio.MulInt(pool.BaseReserve).TruncateInt()
	quoteReceived := lpRatio.MulInt(pool.QuoteReserve).TruncateInt()

	// Check minimum amounts
	if baseReceived.LT(msg.MinBaseAmount) {
		return nil, errors.Wrapf(types.ErrSlippageExceeded, "base received %s below minimum %s", baseReceived.String(), msg.MinBaseAmount.String())
	}
	if quoteReceived.LT(msg.MinQuoteAmount) {
		return nil, errors.Wrapf(types.ErrSlippageExceeded, "quote received %s below minimum %s", quoteReceived.String(), msg.MinQuoteAmount.String())
	}

	// BURN LP TOKENS FROM USER'S WALLET
	lpCoins := sdk.NewCoins(sdk.NewCoin(lpTokenDenom, msg.LPTokenAmount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, lpCoins); err != nil {
		return nil, errors.Wrap(err, "failed to transfer LP tokens from user")
	}
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, lpCoins); err != nil {
		// Rollback: return LP tokens to user
		k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, lpCoins)
		return nil, errors.Wrap(err, "failed to burn LP tokens")
	}

	// Transfer assets back to provider
	coins := sdk.NewCoins(
		sdk.NewCoin(market.BaseSymbol, baseReceived),
		sdk.NewCoin(market.QuoteSymbol, quoteReceived),
	)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, coins); err != nil {
		return nil, errors.Wrap(err, "failed to transfer funds to provider")
	}

	// Update pool reserves
	pool.BaseReserve = pool.BaseReserve.Sub(baseReceived)
	pool.QuoteReserve = pool.QuoteReserve.Sub(quoteReceived)
	pool.LPTokenSupply = pool.LPTokenSupply.Sub(msg.LPTokenAmount)
	pool.UpdatedAt = ctx.BlockTime()
	k.SetLiquidityPool(ctx, pool)

	// UNREGISTER BENEFICIAL OWNERSHIP FOR WITHDRAWN EQUITY
	// This updates or removes the beneficial owner record
	if k.equityKeeper != nil {
		lpPosition, found := k.GetLPPositionByUserAndMarket(ctx, msg.Creator, msg.MarketSymbol)
		if found && lpPosition.CompanyID > 0 {
			// Check if this is a full or partial withdrawal
			if msg.LPTokenAmount.GTE(lpPosition.LPTokenAmount) {
				// Full withdrawal - unregister beneficial ownership completely
				err := k.equityKeeper.UnregisterBeneficialOwner(
					ctx,
					types.ModuleName,
					lpPosition.CompanyID,
					"common",
					msg.Creator,
					lpPosition.ID,
				)
				if err != nil {
					k.Logger(ctx).Error("failed to unregister beneficial owner",
						"provider", msg.Creator,
						"company_id", lpPosition.CompanyID,
						"error", err,
					)
				}
				// Delete the LP position
				k.DeleteLPPosition(ctx, lpPosition)
			} else {
				// Partial withdrawal - update beneficial ownership with reduced shares
				newBaseAmount := lpPosition.BaseAmount.Sub(baseReceived)
				newLPTokenAmount := lpPosition.LPTokenAmount.Sub(msg.LPTokenAmount)

				err := k.equityKeeper.UpdateBeneficialOwnerShares(
					ctx,
					types.ModuleName,
					lpPosition.CompanyID,
					"common",
					msg.Creator,
					lpPosition.ID,
					newBaseAmount,
				)
				if err != nil {
					k.Logger(ctx).Error("failed to update beneficial owner shares",
						"provider", msg.Creator,
						"company_id", lpPosition.CompanyID,
						"new_shares", newBaseAmount.String(),
						"error", err,
					)
				}
				// Update LP position
				k.UpdateLPPositionShares(ctx, lpPosition.ID, newBaseAmount, newLPTokenAmount)
			}
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"remove_liquidity",
			sdk.NewAttribute("market_symbol", msg.MarketSymbol),
			sdk.NewAttribute("provider", msg.Creator),
			sdk.NewAttribute("lp_tokens_burned", msg.LPTokenAmount.String()),
			sdk.NewAttribute("base_received", baseReceived.String()),
			sdk.NewAttribute("quote_received", quoteReceived.String()),
			sdk.NewAttribute("beneficial_owner_updated", "true"),
		),
	)

	return &types.MsgRemoveLiquidityResponse{
		BaseReceived:  baseReceived,
		QuoteReceived: quoteReceived,
		Success:       true,
	}, nil
}

// Swap handles AMM swaps using constant product formula (x * y = k)
func (k msgServer) Swap(goCtx context.Context, msg *types.SimpleMsgSwap) (*types.MsgSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Get market to find base/quote assets
	market, found := k.GetMarketBySymbol(ctx, msg.MarketSymbol)
	if !found {
		return nil, errors.Wrapf(types.ErrInvalidMarket, "market %s not found", msg.MarketSymbol)
	}

	// Get the liquidity pool
	pool, found := k.GetLiquidityPool(ctx, market.BaseSymbol, market.QuoteSymbol)
	if !found {
		return nil, errors.Wrapf(types.ErrPoolNotFound, "liquidity pool for %s not found", msg.MarketSymbol)
	}

	// Validate input amount
	if msg.InputAmount.IsNil() || msg.InputAmount.IsZero() {
		return nil, errors.Wrap(types.ErrInvalidOrderSize, "input amount must be positive")
	}

	// Determine swap direction and reserves
	var inputReserve, outputReserve math.Int
	var outputAsset string
	if msg.InputAsset == market.BaseSymbol {
		inputReserve = pool.BaseReserve
		outputReserve = pool.QuoteReserve
		outputAsset = market.QuoteSymbol
	} else if msg.InputAsset == market.QuoteSymbol {
		inputReserve = pool.QuoteReserve
		outputReserve = pool.BaseReserve
		outputAsset = market.BaseSymbol
	} else {
		return nil, errors.Wrapf(types.ErrInvalidAsset, "input asset %s not in market %s", msg.InputAsset, msg.MarketSymbol)
	}

	// Check sufficient liquidity
	if outputReserve.IsZero() || inputReserve.IsZero() {
		return nil, errors.Wrap(types.ErrInsufficientLiquidity, "pool has insufficient reserves")
	}

	// Calculate output using constant product formula: x * y = k
	// outputAmount = (inputAmount * outputReserve) / (inputReserve + inputAmount)
	// Apply fee from pool configuration (default 0.3%)
	feeRate := pool.Fee
	if feeRate.IsNil() || feeRate.IsZero() {
		feeRate = math.LegacyNewDecWithPrec(3, 3) // 0.3% default
	}
	effectiveInput := math.LegacyNewDecFromInt(msg.InputAmount).Mul(math.LegacyOneDec().Sub(feeRate))

	// Calculate output amount
	numerator := effectiveInput.MulInt(outputReserve)
	denominator := math.LegacyNewDecFromInt(inputReserve).Add(effectiveInput)
	if denominator.IsZero() {
		return nil, errors.Wrap(types.ErrInsufficientLiquidity, "denominator is zero")
	}
	outputAmount := numerator.Quo(denominator).TruncateInt()

	// Calculate fee collected
	feeAmount := feeRate.MulInt(msg.InputAmount).TruncateInt()

	// Check slippage (minimum output)
	if outputAmount.LT(msg.MinOutputAmount) {
		return nil, errors.Wrapf(types.ErrSlippageExceeded, "output %s below minimum %s", outputAmount.String(), msg.MinOutputAmount.String())
	}

	// Calculate price impact
	spotPrice := math.LegacyNewDecFromInt(outputReserve).Quo(math.LegacyNewDecFromInt(inputReserve))
	executionPrice := math.LegacyNewDecFromInt(outputAmount).Quo(math.LegacyNewDecFromInt(msg.InputAmount))
	priceImpact := spotPrice.Sub(executionPrice).Quo(spotPrice).Abs()

	// Check user has sufficient funds
	balance := k.bankKeeper.GetBalance(ctx, creatorAddr, msg.InputAsset)
	if balance.Amount.LT(msg.InputAmount) {
		return nil, errors.Wrapf(types.ErrInsufficientFunds, "insufficient %s balance", msg.InputAsset)
	}

	// Transfer input from user to module
	inputCoins := sdk.NewCoins(sdk.NewCoin(msg.InputAsset, msg.InputAmount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, inputCoins); err != nil {
		return nil, errors.Wrap(err, "failed to transfer input to pool")
	}

	// Transfer output from module to user
	outputCoins := sdk.NewCoins(sdk.NewCoin(outputAsset, outputAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, outputCoins); err != nil {
		return nil, errors.Wrap(err, "failed to transfer output to trader")
	}

	// Update pool reserves
	if msg.InputAsset == market.BaseSymbol {
		pool.BaseReserve = pool.BaseReserve.Add(msg.InputAmount)
		pool.QuoteReserve = pool.QuoteReserve.Sub(outputAmount)
	} else {
		pool.QuoteReserve = pool.QuoteReserve.Add(msg.InputAmount)
		pool.BaseReserve = pool.BaseReserve.Sub(outputAmount)
	}
	pool.Volume24h = pool.Volume24h.Add(msg.InputAmount)
	pool.FeesCollected24h = pool.FeesCollected24h.Add(math.LegacyNewDecFromInt(feeAmount))
	pool.UpdatedAt = ctx.BlockTime()
	k.SetLiquidityPool(ctx, pool)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"swap",
			sdk.NewAttribute("market_symbol", msg.MarketSymbol),
			sdk.NewAttribute("trader", msg.Creator),
			sdk.NewAttribute("input_asset", msg.InputAsset),
			sdk.NewAttribute("input_amount", msg.InputAmount.String()),
			sdk.NewAttribute("output_asset", outputAsset),
			sdk.NewAttribute("output_amount", outputAmount.String()),
			sdk.NewAttribute("fee", feeAmount.String()),
			sdk.NewAttribute("price_impact", priceImpact.String()),
		),
	)

	return &types.MsgSwapResponse{
		OutputAmount: outputAmount,
		Fee:          feeRate,
		PriceImpact:  priceImpact,
		Success:      true,
	}, nil
}

// ===== BLOCKCHAIN-NATIVE MESSAGE HANDLERS =====

// PlaceAtomicSwapOrder handles atomic cross-asset swap orders
func (k msgServer) PlaceAtomicSwapOrder(goCtx context.Context, msg *types.MsgPlaceAtomicSwapOrder) (*types.MsgPlaceAtomicSwapOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate trader address
	traderAddr, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid trader address: %v", err)
	}

	// Validate input parameters
	if msg.FromSymbol == "" || msg.ToSymbol == "" {
		return nil, types.ErrInvalidAsset
	}

	if msg.Quantity == 0 {
		return nil, types.ErrInvalidOrderSize
	}

	quantity := math.NewIntFromUint64(msg.Quantity)

	// Execute atomic swap
	order, err := k.ExecuteAtomicSwap(
		ctx,
		traderAddr,
		msg.FromSymbol,
		msg.ToSymbol,
		quantity,
		msg.MaxSlippage,
	)
	if err != nil {
		return nil, err
	}

	// Log successful atomic swap
	ctx.Logger().Info(
		"Atomic swap executed successfully",
		"trader", msg.Trader,
		"from", msg.FromSymbol,
		"to", msg.ToSymbol,
		"quantity", msg.Quantity,
		"order_id", fmt.Sprintf("%d", order.ID),
		"rate", order.AveragePrice.String(),
	)

	return &types.MsgPlaceAtomicSwapOrderResponse{
		OrderId:               fmt.Sprintf("%d", order.ID),
		EstimatedExchangeRate: order.AveragePrice,
	}, nil
}

// PlaceFractionalOrder handles fractional share trading
func (k msgServer) PlaceFractionalOrder(goCtx context.Context, msg *types.MsgPlaceFractionalOrder) (*types.MsgPlaceFractionalOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate trader address
	traderAddr, err := sdk.AccAddressFromBech32(msg.Trader)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid trader address: %v", err)
	}

	// TODO: Implement fractional order logic
	_ = ctx // Mark as used until full implementation
	_ = traderAddr // Mark as used until full implementation

	// Validate fractional quantity is positive and not zero
	if msg.FractionalQuantity.IsZero() || msg.FractionalQuantity.IsNegative() {
		return nil, types.ErrInvalidOrderSize
	}

	// For now, return not implemented
	// TODO: Implement fractional share logic
	return nil, types.ErrNotImplemented

	// Implementation would:
	// 1. Validate minimum fraction requirements
	// 2. Check if fractional trading is enabled for the symbol
	// 3. Convert fractional shares to internal units
	// 4. Process order matching with fractional logic
	// 5. Handle partial fills and fractional remainder tracking
}

// CreateTradingStrategy handles programmable trading strategy creation
func (k msgServer) CreateTradingStrategy(goCtx context.Context, msg *types.MsgCreateTradingStrategy) (*types.MsgCreateTradingStrategyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid owner address: %v", err)
	}

	// Validate strategy parameters
	if msg.Name == "" {
		return nil, errors.Wrapf(types.ErrInvalidOrderID, "strategy name cannot be empty")
	}

	if len(msg.Conditions) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidOrderType, "strategy must have at least one condition")
	}

	if len(msg.Actions) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidOrderType, "strategy must have at least one action")
	}

	// Generate strategy ID
	strategyID := fmt.Sprintf("strategy_%s_%d", ownerAddr.String()[:8], ctx.BlockHeight())

	// Create trading strategy
	strategy := types.TradingStrategy{
		StrategyId:  strategyID,
		Name:        msg.Name,
		Owner:       msg.Owner,
		Conditions:  msg.Conditions,
		Actions:     msg.Actions,
		MaxExposure: msg.MaxExposure,
		IsActive:    true,
		CreatedAt:   ctx.BlockTime(),
	}

	// Store strategy
	if err := k.SetTradingStrategy(ctx, strategy); err != nil {
		return nil, err
	}

	// Emit strategy creation event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"trading_strategy_created",
			sdk.NewAttribute("owner", msg.Owner),
			sdk.NewAttribute("strategy_id", strategyID),
			sdk.NewAttribute("name", msg.Name),
			sdk.NewAttribute("conditions_count", fmt.Sprintf("%d", len(msg.Conditions))),
			sdk.NewAttribute("actions_count", fmt.Sprintf("%d", len(msg.Actions))),
			sdk.NewAttribute("max_exposure", msg.MaxExposure.String()),
		),
	)

	ctx.Logger().Info(
		"Trading strategy created",
		"owner", msg.Owner,
		"strategy_id", strategyID,
		"name", msg.Name,
	)

	return &types.MsgCreateTradingStrategyResponse{
		StrategyId: strategyID,
	}, nil
}

// ExecuteStrategy handles manual strategy execution
func (k msgServer) ExecuteStrategy(goCtx context.Context, msg *types.MsgExecuteStrategy) (*types.MsgExecuteStrategyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate executor address
	executorAddr, err := sdk.AccAddressFromBech32(msg.Executor)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid executor address: %v", err)
	}

	// Get strategy
	strategy, found := k.GetTradingStrategy(ctx, msg.StrategyId)
	if !found {
		return nil, types.ErrOrderNotFound
	}

	// Check if executor is the owner
	if strategy.Owner != msg.Executor {
		return nil, types.ErrUnauthorized
	}

	// Check if strategy is active
	if !strategy.IsActive {
		return nil, errors.Wrapf(types.ErrMarketInactive, "strategy is inactive")
	}

	// For manual execution or force execution, skip condition checks
	if !msg.ForceExecution {
		// Check if conditions are met
		if !k.CheckStrategyConditions(ctx, strategy) {
			return nil, errors.Wrapf(types.ErrInvalidOrderType, "strategy conditions not met")
		}
	}

	// Execute strategy actions
	orderIds, err := k.ExecuteStrategyActions(ctx, strategy, executorAddr)
	if err != nil {
		return nil, err
	}

	// Emit strategy execution event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"trading_strategy_executed",
			sdk.NewAttribute("executor", msg.Executor),
			sdk.NewAttribute("strategy_id", msg.StrategyId),
			sdk.NewAttribute("force_execution", fmt.Sprintf("%t", msg.ForceExecution)),
			sdk.NewAttribute("orders_placed", fmt.Sprintf("%d", len(orderIds))),
		),
	)

	ctx.Logger().Info(
		"Trading strategy executed",
		"executor", msg.Executor,
		"strategy_id", msg.StrategyId,
		"orders_placed", len(orderIds),
	)

	return &types.MsgExecuteStrategyResponse{
		ExecutedOrderIds: orderIds,
		OrdersPlaced:     uint64(len(orderIds)),
	}, nil
}

// Helper methods for trading strategies

// SetTradingStrategy stores a trading strategy
func (k Keeper) SetTradingStrategy(ctx sdk.Context, strategy types.TradingStrategy) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradingStrategyKey(strategy.StrategyId)

	bz, err := json.Marshal(strategy)
	if err != nil {
		return fmt.Errorf("failed to marshal trading strategy: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// GetTradingStrategy retrieves a trading strategy
func (k Keeper) GetTradingStrategy(ctx sdk.Context, strategyId string) (types.TradingStrategy, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradingStrategyKey(strategyId)

	bz := store.Get(key)
	if bz == nil {
		return types.TradingStrategy{}, false
	}

	var strategy types.TradingStrategy
	if err := json.Unmarshal(bz, &strategy); err != nil {
		// Log error but return not found instead of panicking
		ctx.Logger().Error("failed to unmarshal trading strategy", "error", err)
		return types.TradingStrategy{}, false
	}
	return strategy, true
}

// CheckStrategyConditions checks if strategy conditions are met
func (k Keeper) CheckStrategyConditions(ctx sdk.Context, strategy types.TradingStrategy) bool {
	for _, condition := range strategy.Conditions {
		if !k.CheckTriggerCondition(ctx, condition) {
			return false
		}
	}
	return true
}

// CheckTriggerCondition checks a single trigger condition
func (k Keeper) CheckTriggerCondition(ctx sdk.Context, condition types.TriggerCondition) bool {
	// Get current price for the symbol
	currentPrice := k.GetAssetPrice(ctx, condition.Symbol)
	
	switch condition.ConditionType {
	case "ABOVE":
		return currentPrice.GT(condition.PriceThreshold)
	case "BELOW":
		return currentPrice.LT(condition.PriceThreshold)
	case "EQUALS":
		// Use small tolerance for equality
		tolerance := math.LegacyNewDecWithPrec(1, 4) // 0.01%
		diff := currentPrice.Sub(condition.PriceThreshold).Abs()
		maxDiff := condition.PriceThreshold.Mul(tolerance)
		return diff.LTE(maxDiff)
	default:
		return false
	}
}

// ExecuteStrategyActions executes strategy actions
func (k Keeper) ExecuteStrategyActions(ctx sdk.Context, strategy types.TradingStrategy, executor sdk.AccAddress) ([]string, error) {
	var orderIds []string
	
	for _, action := range strategy.Actions {
		orderId, err := k.ExecuteTradingAction(ctx, action, executor, strategy.MaxExposure)
		if err != nil {
			// Log error but continue with other actions
			ctx.Logger().Error(
				"Failed to execute strategy action",
				"strategy_id", strategy.StrategyId,
				"action", action,
				"error", err,
			)
			continue
		}
		orderIds = append(orderIds, orderId)
	}
	
	return orderIds, nil
}

// ExecuteTradingAction executes a single trading action
func (k Keeper) ExecuteTradingAction(ctx sdk.Context, action types.TradingAction, executor sdk.AccAddress, maxExposure math.Int) (string, error) {
	// Convert fractional quantity to uint64 for order placement
	quantity := action.Quantity.TruncateInt().Uint64()
	
	if quantity == 0 {
		return "", types.ErrInvalidOrderSize
	}

	// Calculate price (current price + offset)
	currentPrice := k.GetAssetPrice(ctx, action.Symbol)
	orderPrice := currentPrice.Add(action.PriceOffset)
	
	// Create and place order
	orderID := k.GetNextOrderID(ctx)
	orderIdStr := fmt.Sprintf("strategy_order_%d", orderID)
	
	order := types.Order{
		ID:               orderID,
		User:             executor.String(),
		Type:             types.OrderTypeMarket, // Simplified for strategy execution
		Side:             types.OrderSideBuy,    // Simplified for now
		MarketSymbol:     action.Symbol + "/HODL",
		BaseSymbol:       action.Symbol,
		QuoteSymbol:      "HODL", 
		Quantity:         math.NewIntFromUint64(quantity),
		Price:            orderPrice,
		Status:           types.OrderStatusPending,
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
		ClientOrderID:    fmt.Sprintf("strategy_%s", executor.String()[:8]),
	}
	
	// Store the order
	k.SetOrder(ctx, order)
	
	// For now, immediately mark as filled for demonstration
	// In production, would go through order matching
	order.Status = types.OrderStatusFilled
	order.FilledQuantity = math.NewIntFromUint64(quantity)
	order.AveragePrice = orderPrice
	k.SetOrder(ctx, order)
	
	return orderIdStr, nil
}