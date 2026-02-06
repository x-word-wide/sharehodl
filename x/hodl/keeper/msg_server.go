package keeper

import (
	"context"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface  
func NewMsgServerImpl(keeper Keeper) *msgServer {
	return &msgServer{Keeper: keeper}
}

// Removed interface validation for now

// MintHODL handles the minting of HODL stablecoins against collateral
func (k msgServer) MintHODL(goCtx context.Context, msg *types.SimpleMsgMintHODL) (*types.MsgMintHODLResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// PERFORMANCE FIX: Validate amount is positive
	if !msg.HodlAmount.IsPositive() {
		return nil, errors.Wrap(types.ErrInvalidAmount, "HODL amount must be positive")
	}

	// PERFORMANCE FIX: Validate collateral is not empty or zero
	if msg.CollateralCoins.IsZero() || !msg.CollateralCoins.IsValid() {
		return nil, errors.Wrap(types.ErrInvalidAmount, "collateral amount must be positive and valid")
	}

	// Get parameters
	params := k.GetParams(ctx)
	if !params.MintingEnabled {
		return nil, types.ErrMintingDisabled
	}

	// SECURITY: Check circuit breaker
	if k.IsMintingPaused(ctx) {
		return nil, errors.Wrap(types.ErrMintingDisabled, "minting paused due to bad debt circuit breaker")
	}

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address: %v", err)
	}

	// SECURITY FIX: Validate all collateral types are whitelisted
	for _, coin := range msg.CollateralCoins {
		if !k.IsCollateralWhitelisted(ctx, coin.Denom) {
			return nil, errors.Wrapf(types.ErrInvalidCollateral,
				"collateral type %s is not whitelisted; only whitelisted collateral can be used for minting", coin.Denom)
		}
	}

	// Check if user has sufficient collateral
	balance := k.bankKeeper.GetAllBalances(ctx, creatorAddr)
	if !balance.IsAllGTE(msg.CollateralCoins) {
		return nil, errors.Wrapf(sdkerrors.ErrInsufficientFunds, "insufficient balance for collateral")
	}

	// SECURITY FIX: Calculate collateral value using proper oracle prices
	providedCollateralValue, err := k.GetCollateralValue(ctx, msg.CollateralCoins)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidCollateral, "failed to calculate collateral value: %v", err)
	}

	// Calculate required collateral value: hodlAmount * collateralRatio
	// hodlAmount is in uhodl, collateralValue is in HODL-equivalent terms
	requiredCollateralValue := math.LegacyNewDecFromInt(msg.HodlAmount).Mul(params.CollateralRatio)

	if providedCollateralValue.LT(requiredCollateralValue) {
		return nil, errors.Wrapf(types.ErrInsufficientCollateral,
			"provided collateral value %s is less than required %s (at %s%% ratio)",
			providedCollateralValue.String(), requiredCollateralValue.String(),
			params.CollateralRatio.Mul(math.LegacyNewDec(100)).String())
	}

	// Calculate mint fee
	mintFee := params.MintFee.MulInt(msg.HodlAmount).TruncateInt()
	hodlToMint := msg.HodlAmount.Sub(mintFee)

	// Transfer collateral to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, msg.CollateralCoins); err != nil {
		return nil, errors.Wrapf(err, "failed to transfer collateral to module")
	}

	// Mint HODL tokens to user
	hodlCoins := sdk.NewCoins(sdk.NewCoin("hodl", hodlToMint))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, hodlCoins); err != nil {
		return nil, errors.Wrapf(err, "failed to mint HODL tokens")
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, hodlCoins); err != nil {
		return nil, errors.Wrapf(err, "failed to send HODL tokens to user")
	}

	// Update or create collateral position
	position, exists := k.GetCollateralPosition(ctx, creatorAddr)
	if exists {
		position.Collateral = position.Collateral.Add(msg.CollateralCoins...)
		position.MintedHODL = position.MintedHODL.Add(hodlToMint)
	} else {
		position = types.NewCollateralPosition(msg.Creator, msg.CollateralCoins, hodlToMint, ctx.BlockHeight())
	}
	position.LastUpdated = ctx.BlockHeight()
	k.SetCollateralPosition(ctx, position)

	// Update total supply
	totalSupply := k.getTotalSupply(ctx)
	k.SetTotalSupply(ctx, totalSupply.Add(hodlToMint))

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"mint_hodl",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("collateral", msg.CollateralCoins.String()),
			sdk.NewAttribute("hodl_minted", hodlToMint.String()),
			sdk.NewAttribute("mint_fee", mintFee.String()),
		),
	)

	return &types.MsgMintHODLResponse{
		HodlMinted: hodlToMint,
		Fee:        mintFee,
	}, nil
}

// BurnHODL handles the burning of HODL stablecoins to redeem collateral
func (k msgServer) BurnHODL(goCtx context.Context, msg *types.SimpleMsgBurnHODL) (*types.MsgBurnHODLResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// PERFORMANCE FIX: Validate amount is positive
	if !msg.HodlAmount.IsPositive() {
		return nil, errors.Wrap(types.ErrInvalidAmount, "HODL amount must be positive")
	}

	// Get parameters
	params := k.GetParams(ctx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address: %v", err)
	}

	// Check if user has sufficient HODL balance
	hodlBalance := k.bankKeeper.GetBalance(ctx, creatorAddr, "hodl")
	if hodlBalance.Amount.LT(msg.HodlAmount) {
		return nil, types.ErrInsufficientHODLBalance
	}

	// Get user's collateral position
	position, exists := k.GetCollateralPosition(ctx, creatorAddr)
	if !exists {
		return nil, types.ErrPositionNotFound
	}

	// Calculate how much collateral to return (proportional to HODL burned)
	if position.MintedHODL.LT(msg.HodlAmount) {
		return nil, errors.Wrap(sdkerrors.ErrInvalidRequest, "cannot burn more HODL than minted")
	}

	collateralRatio := math.LegacyNewDecFromInt(msg.HodlAmount).QuoInt(position.MintedHODL)
	collateralToReturn := sdk.NewCoins()
	
	for _, coin := range position.Collateral {
		amount := collateralRatio.MulInt(coin.Amount).TruncateInt()
		if amount.IsPositive() {
			collateralToReturn = collateralToReturn.Add(sdk.NewCoin(coin.Denom, amount))
		}
	}

	// Calculate burn fee
	burnFee := params.BurnFee.MulInt(msg.HodlAmount).TruncateInt()
	
	// Burn HODL tokens from user
	hodlCoins := sdk.NewCoins(sdk.NewCoin("hodl", msg.HodlAmount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, hodlCoins); err != nil {
		return nil, errors.Wrapf(err, "failed to send HODL to module for burning")
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, hodlCoins); err != nil {
		return nil, errors.Wrapf(err, "failed to burn HODL tokens")
	}

	// Return collateral to user (minus fee)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, collateralToReturn); err != nil {
		return nil, errors.Wrapf(err, "failed to return collateral to user")
	}

	// Update position
	position.Collateral = position.Collateral.Sub(collateralToReturn...)
	position.MintedHODL = position.MintedHODL.Sub(msg.HodlAmount)
	position.LastUpdated = ctx.BlockHeight()

	if position.IsEmpty() {
		k.DeleteCollateralPosition(ctx, creatorAddr)
	} else {
		k.SetCollateralPosition(ctx, position)
	}

	// Update total supply
	totalSupply := k.getTotalSupply(ctx)
	k.SetTotalSupply(ctx, totalSupply.Sub(msg.HodlAmount))

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"burn_hodl",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("hodl_burned", msg.HodlAmount.String()),
			sdk.NewAttribute("collateral_returned", collateralToReturn.String()),
			sdk.NewAttribute("burn_fee", burnFee.String()),
		),
	)

	return &types.MsgBurnHODLResponse{
		CollateralReturned: collateralToReturn,
		Fee:               burnFee,
	}, nil
}

// Liquidate handles the liquidation of undercollateralized positions
func (k msgServer) Liquidate(goCtx context.Context, msg *types.MsgLiquidate) (*types.MsgLiquidateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate liquidator address
	liquidatorAddr, err := sdk.AccAddressFromBech32(msg.Liquidator)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid liquidator address: %v", err)
	}

	// Validate position owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.PositionOwner)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid position owner address: %v", err)
	}

	// Get the position
	position, found := k.GetCollateralPosition(ctx, ownerAddr)
	if !found {
		return nil, types.ErrPositionNotFound
	}

	// Calculate debt and collateral info before liquidation
	totalDebt := k.GetTotalDebt(ctx, position)
	collateralValue, _ := k.GetCollateralValue(ctx, position.Collateral)
	ratio, _ := k.GetCollateralRatio(ctx, position)

	// Execute liquidation
	if err := k.LiquidatePosition(ctx, liquidatorAddr, ownerAddr); err != nil {
		return nil, errors.Wrapf(err, "liquidation failed")
	}

	return &types.MsgLiquidateResponse{
		DebtRepaid:       totalDebt.TruncateInt(),
		CollateralSeized: collateralValue.TruncateInt(),
		LiquidationRatio: ratio,
	}, nil
}

// AddCollateral handles adding collateral to an existing position
func (k msgServer) AddCollateral(goCtx context.Context, msg *types.MsgAddCollateral) (*types.MsgAddCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// PERFORMANCE FIX: Validate collateral is not empty or zero
	if msg.Collateral.IsZero() || !msg.Collateral.IsValid() {
		return nil, errors.Wrap(types.ErrInvalidAmount, "collateral amount must be positive and valid")
	}

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address: %v", err)
	}

	// Execute add collateral
	if err := k.Keeper.AddCollateral(ctx, ownerAddr, msg.Collateral); err != nil {
		return nil, errors.Wrapf(err, "failed to add collateral")
	}

	// Get updated ratio
	position, _ := k.GetCollateralPosition(ctx, ownerAddr)
	newRatio, _ := k.GetCollateralRatio(ctx, position)

	return &types.MsgAddCollateralResponse{
		NewCollateralRatio: newRatio,
	}, nil
}

// WithdrawCollateral handles withdrawing excess collateral from a position
func (k msgServer) WithdrawCollateral(goCtx context.Context, msg *types.MsgWithdrawCollateral) (*types.MsgWithdrawCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// PERFORMANCE FIX: Validate collateral is not empty or zero
	if msg.Collateral.IsZero() || !msg.Collateral.IsValid() {
		return nil, errors.Wrap(types.ErrInvalidAmount, "collateral amount must be positive and valid")
	}

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address: %v", err)
	}

	// Execute withdraw collateral
	if err := k.Keeper.WithdrawCollateral(ctx, ownerAddr, msg.Collateral); err != nil {
		return nil, errors.Wrapf(err, "failed to withdraw collateral")
	}

	// Get updated ratio
	position, found := k.GetCollateralPosition(ctx, ownerAddr)
	var newRatio math.LegacyDec
	if found {
		newRatio, _ = k.GetCollateralRatio(ctx, position)
	}

	return &types.MsgWithdrawCollateralResponse{
		NewCollateralRatio: newRatio,
	}, nil
}

// PayStabilityDebt handles paying off accumulated stability fees
func (k msgServer) PayStabilityDebt(goCtx context.Context, msg *types.MsgPayStabilityDebt) (*types.MsgPayStabilityDebtResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// PERFORMANCE FIX: Validate amount is positive
	if !msg.Amount.IsPositive() {
		return nil, errors.Wrap(types.ErrInvalidAmount, "payment amount must be positive")
	}

	// Validate owner address
	ownerAddr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address: %v", err)
	}

	// Execute pay stability debt
	if err := k.Keeper.PayStabilityDebt(ctx, ownerAddr, msg.Amount); err != nil {
		return nil, errors.Wrapf(err, "failed to pay stability debt")
	}

	// Get remaining debt
	position, _ := k.GetCollateralPosition(ctx, ownerAddr)

	return &types.MsgPayStabilityDebtResponse{
		RemainingDebt: position.StabilityDebt,
	}, nil
}