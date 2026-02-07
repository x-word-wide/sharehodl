package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// Module error codes (2-99 range to avoid conflicts)
var (
	ErrFeeAbstractionDisabled = sdkerrors.Register(ModuleName, 2, "fee abstraction is disabled")

	ErrInsufficientEquityValue = sdkerrors.Register(ModuleName, 3, "equity value insufficient to cover fee + markup")

	ErrNoEligibleEquity = sdkerrors.Register(ModuleName, 4, "no eligible equity for fee abstraction")

	ErrSwapFailed = sdkerrors.Register(ModuleName, 5, "equity-to-HODL swap failed")

	ErrSlippageExceeded = sdkerrors.Register(ModuleName, 6, "swap slippage exceeded maximum")

	ErrBlockLimitExceeded = sdkerrors.Register(ModuleName, 7, "block fee abstraction limit exceeded")

	ErrEquityNotWhitelisted = sdkerrors.Register(ModuleName, 8, "equity not whitelisted for fee abstraction")

	ErrInsufficientTreasuryFunds = sdkerrors.Register(ModuleName, 9, "insufficient treasury funds")

	ErrInvalidAuthority = sdkerrors.Register(ModuleName, 10, "invalid governance authority")

	ErrGrantNotFound = sdkerrors.Register(ModuleName, 11, "treasury grant not found")

	ErrGrantExpired = sdkerrors.Register(ModuleName, 12, "treasury grant has expired")

	ErrGrantExhausted = sdkerrors.Register(ModuleName, 13, "treasury grant allowance exhausted")

	ErrInvalidGrant = sdkerrors.Register(ModuleName, 14, "invalid treasury grant")

	ErrMarketNotFound = sdkerrors.Register(ModuleName, 15, "market not found for equity")

	ErrMarketInactive = sdkerrors.Register(ModuleName, 16, "market is inactive or halted")

	ErrPriceUnavailable = sdkerrors.Register(ModuleName, 17, "equity price is unavailable")

	ErrInvalidParams = sdkerrors.Register(ModuleName, 18, "invalid module parameters")

	ErrZeroAmount = sdkerrors.Register(ModuleName, 19, "amount cannot be zero")

	ErrInvalidAddress = sdkerrors.Register(ModuleName, 20, "invalid address")

	ErrNotWhitelisted = sdkerrors.Register(ModuleName, 21, "address not whitelisted to fund treasury")

	ErrTreasuryWithdrawalNotAllowed = sdkerrors.Register(ModuleName, 22, "direct treasury withdrawals not allowed - use governance proposal")
)
