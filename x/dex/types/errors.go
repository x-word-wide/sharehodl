package types

import (
	"cosmossdk.io/errors"
)

// x/dex module sentinel errors
var (
	// Market errors
	ErrMarketNotFound       = errors.Register(ModuleName, 1, "market not found")
	ErrMarketAlreadyExists  = errors.Register(ModuleName, 2, "market already exists")
	ErrInvalidMarket        = errors.Register(ModuleName, 3, "invalid market")
	ErrMarketInactive       = errors.Register(ModuleName, 4, "market is inactive")
	ErrTradingHalted        = errors.Register(ModuleName, 5, "trading is halted")

	// Order errors
	ErrOrderNotFound        = errors.Register(ModuleName, 10, "order not found")
	ErrInvalidOrderID       = errors.Register(ModuleName, 11, "invalid order ID")
	ErrInvalidOrderSide     = errors.Register(ModuleName, 12, "invalid order side")
	ErrInvalidOrderType     = errors.Register(ModuleName, 13, "invalid order type")
	ErrInvalidOrderSize     = errors.Register(ModuleName, 14, "invalid order size")
	ErrInvalidTimeInForce   = errors.Register(ModuleName, 15, "invalid time in force")
	ErrOrderExpired         = errors.Register(ModuleName, 16, "order has expired")
	ErrOrderAlreadyCancelled = errors.Register(ModuleName, 17, "order already cancelled")
	ErrOrderAlreadyFilled   = errors.Register(ModuleName, 18, "order already filled")
	ErrCannotCancelOrder    = errors.Register(ModuleName, 19, "cannot cancel order")

	// Price errors
	ErrInvalidPrice         = errors.Register(ModuleName, 20, "invalid price")
	ErrPriceTooLow          = errors.Register(ModuleName, 21, "price too low")
	ErrPriceTooHigh         = errors.Register(ModuleName, 22, "price too high")
	ErrInvalidTickSize      = errors.Register(ModuleName, 23, "invalid tick size")

	// Balance errors
	ErrInsufficientFunds    = errors.Register(ModuleName, 30, "insufficient funds")
	ErrInsufficientBalance  = errors.Register(ModuleName, 31, "insufficient balance")
	ErrInvalidAsset         = errors.Register(ModuleName, 32, "invalid asset")

	// Trading errors
	ErrNoMatchingOrders     = errors.Register(ModuleName, 40, "no matching orders")
	ErrTradeExecutionFailed = errors.Register(ModuleName, 41, "trade execution failed")
	ErrSlippageExceeded     = errors.Register(ModuleName, 42, "slippage exceeded")

	// Liquidity pool errors
	ErrPoolNotFound         = errors.Register(ModuleName, 50, "liquidity pool not found")
	ErrPoolAlreadyExists    = errors.Register(ModuleName, 51, "liquidity pool already exists")
	ErrInvalidLiquidity     = errors.Register(ModuleName, 52, "invalid liquidity amount")
	ErrInsufficientLiquidity = errors.Register(ModuleName, 53, "insufficient liquidity")
	ErrInvalidSwapAmount    = errors.Register(ModuleName, 54, "invalid swap amount")
	ErrSwapRatioTooHigh     = errors.Register(ModuleName, 55, "swap ratio too high")

	// Fee errors
	ErrInvalidFee          = errors.Register(ModuleName, 60, "invalid fee")
	ErrFeeTooHigh          = errors.Register(ModuleName, 61, "fee too high")

	// Authorization errors
	ErrUnauthorized        = errors.Register(ModuleName, 70, "unauthorized")
	ErrPermissionDenied    = errors.Register(ModuleName, 71, "permission denied")
)