package types

import (
	"cosmossdk.io/errors"
)

// x/hodl module sentinel errors
var (
	ErrMintingDisabled          = errors.Register(ModuleName, 1, "HODL minting is disabled")
	ErrInsufficientCollateral   = errors.Register(ModuleName, 2, "insufficient collateral for minting")
	ErrInvalidCollateralType    = errors.Register(ModuleName, 3, "invalid collateral type")
	ErrCollateralRatioTooLow    = errors.Register(ModuleName, 4, "collateral ratio below minimum")
	ErrInsufficientHODLBalance  = errors.Register(ModuleName, 5, "insufficient HODL balance for burning")
	ErrInvalidBurnAmount        = errors.Register(ModuleName, 6, "invalid burn amount")
	ErrLiquidationInProgress    = errors.Register(ModuleName, 7, "position is being liquidated")
	ErrPositionNotFound         = errors.Register(ModuleName, 8, "collateral position not found")
	ErrInvalidCollateral        = errors.Register(ModuleName, 9, "invalid collateral")
	ErrInvalidAmount            = errors.Register(ModuleName, 10, "invalid amount - must be positive")
)