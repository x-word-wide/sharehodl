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
	ErrInvalidOrderPrice    = errors.Register(ModuleName, 21, "invalid order price")
	ErrPriceTooLow          = errors.Register(ModuleName, 22, "price too low")
	ErrPriceTooHigh         = errors.Register(ModuleName, 23, "price too high")
	ErrInvalidTickSize      = errors.Register(ModuleName, 24, "invalid tick size")
	ErrInvalidSlippage      = errors.Register(ModuleName, 25, "invalid slippage tolerance")

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
	ErrTooFrequentUpdate   = errors.Register(ModuleName, 72, "price update too frequent")

	// Atomic swap errors
	ErrAssetNotFound       = errors.Register(ModuleName, 80, "asset not found")
	ErrSameAssetSwap       = errors.Register(ModuleName, 81, "cannot swap same asset")
	ErrNoMarketPrice       = errors.Register(ModuleName, 82, "no market price available")
	ErrInsufficientShares  = errors.Register(ModuleName, 83, "insufficient shares")
	ErrNotImplemented      = errors.Register(ModuleName, 84, "feature not implemented")
	ErrCrossEquitySwapNotImplemented = errors.Register(ModuleName, 85, "cross-equity swap not implemented")

	// Trading strategy errors
	ErrInvalidStrategyID     = errors.Register(ModuleName, 90, "invalid strategy ID")
	ErrInvalidStrategyTriggers = errors.Register(ModuleName, 91, "invalid strategy triggers")
	ErrTooManyTriggers      = errors.Register(ModuleName, 92, "too many triggers")

	// Trading controls and safety errors
	ErrOutsideTradingHours   = errors.Register(ModuleName, 100, "outside trading hours")
	ErrInvalidParameter      = errors.Register(ModuleName, 101, "invalid parameter")
	ErrOwnershipLimitExceeded = errors.Register(ModuleName, 102, "ownership limit exceeded")
	ErrDailyAcquisitionLimitExceeded = errors.Register(ModuleName, 103, "daily acquisition limit exceeded")
	ErrBoardApprovalRequired = errors.Register(ModuleName, 104, "board approval required")
	ErrTransferApprovalRequired = errors.Register(ModuleName, 105, "transfer approval required")
	ErrRestrictedInvestor    = errors.Register(ModuleName, 106, "restricted investor")
	ErrRightOfFirstRefusalPending = errors.Register(ModuleName, 107, "right of first refusal pending")
	ErrTakeoverDefenseTriggered = errors.Register(ModuleName, 108, "takeover defense triggered")
	ErrOrderTooLarge         = errors.Register(ModuleName, 109, "order size exceeds maximum limit")
	ErrOrderTooSmall         = errors.Register(ModuleName, 110, "order size below minimum limit")
	ErrDailyLimitExceeded    = errors.Register(ModuleName, 111, "daily trading limit exceeded")
	
	// Event types for advanced trading
	EventTypeCircuitBreakerTriggered = "circuit_breaker_triggered"
	EventTypeOrderRejected          = "order_rejected"
	EventTypeOrderPartiallyFilled   = "order_partially_filled"
	EventTypeOrderExpired          = "order_expired"
	EventTypeOrderCancelled        = "order_cancelled"
)