package keeper

import (
	"context"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

	// Get parameters
	params := k.GetParams(ctx)
	if !params.MintingEnabled {
		return nil, types.ErrMintingDisabled
	}

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(errors.ErrInvalidAddress, "invalid creator address: %v", err)
	}

	// Check if user has sufficient collateral
	balance := k.bankKeeper.GetAllBalances(ctx, creatorAddr)
	if !balance.IsAllGTE(msg.CollateralCoins) {
		return nil, errors.Wrapf(errors.ErrInsufficientFunds, "insufficient balance for collateral")
	}

	// Calculate required collateral value (simplified - in real implementation would use oracle prices)
	// For now, assume 1:1 USD peg and require 150% collateral
	requiredCollateralValue := msg.HodlAmount.Mul(params.CollateralRatio.TruncateInt())
	providedCollateralValue := msg.CollateralCoins.AmountOf("stake") // Assuming stake token as collateral
	
	if providedCollateralValue.LT(requiredCollateralValue) {
		return nil, types.ErrInsufficientCollateral
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
	totalSupply := k.GetTotalSupply(ctx)
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

	// Get parameters
	params := k.GetParams(ctx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(errors.ErrInvalidAddress, "invalid creator address: %v", err)
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
		return nil, errors.Wrap(errors.ErrInvalidRequest, "cannot burn more HODL than minted")
	}

	collateralRatio := sdk.NewDecFromInt(msg.HodlAmount).QuoInt(position.MintedHODL)
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
	totalSupply := k.GetTotalSupply(ctx)
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