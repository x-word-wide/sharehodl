package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// ObserveDeposit handles deposit observations from validators
func (ms msgServer) ObserveDeposit(goCtx context.Context, msg *types.MsgObserveDeposit) (*types.MsgObserveDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validator, err := sdk.AccAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	depositID, err := ms.Keeper.ObserveDeposit(
		ctx,
		validator,
		msg.ChainID,
		msg.AssetSymbol,
		msg.ExternalTxHash,
		msg.ExternalBlockHeight,
		msg.Sender,
		msg.Recipient,
		msg.Amount,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgObserveDepositResponse{DepositID: depositID}, nil
}

// AttestDeposit handles deposit attestations from validators
func (ms msgServer) AttestDeposit(goCtx context.Context, msg *types.MsgAttestDeposit) (*types.MsgAttestDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validator, err := sdk.AccAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	attestations, completed, err := ms.Keeper.AttestDeposit(
		ctx,
		validator,
		msg.DepositID,
		msg.Approved,
		msg.ObservedTxHash,
		msg.ObservedAmount,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgAttestDepositResponse{
		Attestations: attestations,
		Completed:    completed,
	}, nil
}

// RequestWithdrawal handles withdrawal requests from users
func (ms msgServer) RequestWithdrawal(goCtx context.Context, msg *types.MsgRequestWithdrawal) (*types.MsgRequestWithdrawalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	withdrawalID, err := ms.Keeper.RequestWithdrawal(
		ctx,
		sender,
		msg.ChainID,
		msg.AssetSymbol,
		msg.Recipient,
		msg.Amount,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgRequestWithdrawalResponse{WithdrawalID: withdrawalID}, nil
}

// SubmitTSSSignature handles TSS signature submissions from validators
func (ms msgServer) SubmitTSSSignature(goCtx context.Context, msg *types.MsgSubmitTSSSignature) (*types.MsgSubmitTSSSignatureResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	validator, err := sdk.AccAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	completed, err := ms.Keeper.SubmitTSSSignature(
		ctx,
		validator,
		msg.SessionID,
		msg.SignatureData,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitTSSSignatureResponse{
		SessionID: msg.SessionID,
		Completed: completed,
	}, nil
}

// UpdateCircuitBreaker handles circuit breaker updates from governance
func (ms msgServer) UpdateCircuitBreaker(goCtx context.Context, msg *types.MsgUpdateCircuitBreaker) (*types.MsgUpdateCircuitBreakerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority
	if msg.Authority != ms.Keeper.GetAuthority() {
		return nil, types.ErrUnauthorized
	}

	cb := types.NewCircuitBreaker(
		msg.Enabled,
		msg.Reason,
		msg.Authority,
		msg.CanDeposit,
		msg.CanWithdraw,
		msg.CanAttest,
		nil,
	)

	ms.Keeper.SetCircuitBreaker(ctx, cb)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"extbridge_circuit_breaker_updated",
			sdk.NewAttribute("enabled", fmt.Sprintf("%t", msg.Enabled)),
			sdk.NewAttribute("reason", msg.Reason),
		),
	)

	return &types.MsgUpdateCircuitBreakerResponse{}, nil
}

// AddExternalChain handles adding new external chains (governance only)
func (ms msgServer) AddExternalChain(goCtx context.Context, msg *types.MsgAddExternalChain) (*types.MsgAddExternalChainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority
	if msg.Authority != ms.Keeper.GetAuthority() {
		return nil, types.ErrUnauthorized
	}

	chain := types.NewExternalChain(
		msg.ChainID,
		msg.ChainName,
		msg.ChainType,
		true, // Enabled by default
		msg.ConfirmationsReq,
		msg.BlockTime,
		msg.TSSAddress,
		msg.MinDeposit,
		msg.MaxDeposit,
	)

	if err := chain.Validate(); err != nil {
		return nil, err
	}

	ms.Keeper.SetExternalChain(ctx, chain)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"extbridge_chain_added",
			sdk.NewAttribute("chain_id", chain.ChainID),
			sdk.NewAttribute("chain_name", chain.ChainName),
		),
	)

	return &types.MsgAddExternalChainResponse{}, nil
}

// AddExternalAsset handles adding new external assets (governance only)
func (ms msgServer) AddExternalAsset(goCtx context.Context, msg *types.MsgAddExternalAsset) (*types.MsgAddExternalAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority
	if msg.Authority != ms.Keeper.GetAuthority() {
		return nil, types.ErrUnauthorized
	}

	// Verify chain exists
	if _, found := ms.Keeper.GetExternalChain(ctx, msg.ChainID); !found {
		return nil, types.ErrChainNotSupported
	}

	asset := types.NewExternalAsset(
		msg.ChainID,
		msg.AssetSymbol,
		msg.AssetName,
		msg.ContractAddress,
		msg.Decimals,
		true, // Enabled by default
		msg.ConversionRate,
		msg.DailyLimit,
		msg.PerTxLimit,
	)

	if err := asset.Validate(); err != nil {
		return nil, err
	}

	ms.Keeper.SetExternalAsset(ctx, asset)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"extbridge_asset_added",
			sdk.NewAttribute("chain_id", asset.ChainID),
			sdk.NewAttribute("asset_symbol", asset.AssetSymbol),
		),
	)

	return &types.MsgAddExternalAssetResponse{}, nil
}
