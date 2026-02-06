package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

// GetWithdrawal retrieves a withdrawal by ID
func (k Keeper) GetWithdrawal(ctx sdk.Context, withdrawalID uint64) (types.Withdrawal, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.WithdrawalKey(withdrawalID))
	if bz == nil {
		return types.Withdrawal{}, false
	}

	var withdrawal types.Withdrawal
	if err := json.Unmarshal(bz, &withdrawal); err != nil {
		return types.Withdrawal{}, false
	}
	return withdrawal, true
}

// SetWithdrawal stores a withdrawal
func (k Keeper) SetWithdrawal(ctx sdk.Context, withdrawal types.Withdrawal) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(withdrawal)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal withdrawal", "error", err)
		return fmt.Errorf("failed to marshal withdrawal: %w", err)
	}
	store.Set(types.WithdrawalKey(withdrawal.ID), bz)
	return nil
}

// GetAllWithdrawals returns all withdrawals
func (k Keeper) GetAllWithdrawals(ctx sdk.Context) []types.Withdrawal {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.WithdrawalPrefix)
	defer iterator.Close()

	withdrawals := []types.Withdrawal{}
	for ; iterator.Valid(); iterator.Next() {
		var withdrawal types.Withdrawal
		if err := json.Unmarshal(iterator.Value(), &withdrawal); err != nil {
			continue
		}
		withdrawals = append(withdrawals, withdrawal)
	}
	return withdrawals
}

// RequestWithdrawal creates a new withdrawal request
func (k Keeper) RequestWithdrawal(
	ctx sdk.Context,
	sender sdk.AccAddress,
	chainID string,
	assetSymbol string,
	recipient string,
	hodlAmount math.Int,
) (uint64, error) {
	// Check if sender is banned
	if banned, reason := k.IsAddressBanned(ctx, sender.String()); banned {
		k.Logger(ctx).Warn("banned address attempted withdrawal",
			"address", sender.String(),
			"reason", reason,
		)
		return 0, types.ErrAddressBanned.Wrapf("sender is banned: %s", reason)
	}

	// Check circuit breaker
	if err := k.IsOperationAllowed(ctx, "withdraw"); err != nil {
		return 0, err
	}

	// Check if bridging is enabled
	params := k.GetParams(ctx)
	if !params.BridgingEnabled {
		return 0, types.ErrBridgingDisabled
	}

	// Verify chain exists and is enabled
	chain, found := k.GetExternalChain(ctx, chainID)
	if !found {
		return 0, types.ErrChainNotSupported
	}
	if !chain.Enabled {
		return 0, types.ErrChainNotSupported
	}

	// Verify asset exists and is enabled
	asset, found := k.GetExternalAsset(ctx, chainID, assetSymbol)
	if !found {
		return 0, types.ErrAssetNotSupported
	}
	if !asset.Enabled {
		return 0, types.ErrAssetNotSupported
	}

	// Check sender has enough HODL
	balance := k.bankKeeper.GetBalance(ctx, sender, hodltypes.HODLDenom)
	if balance.Amount.LT(hodlAmount) {
		return 0, types.ErrInsufficientBalance
	}

	// Calculate fee
	feeAmount := math.LegacyNewDecFromInt(hodlAmount).Mul(params.BridgeFee).TruncateInt()
	netHODL := hodlAmount.Sub(feeAmount)

	// Convert to external asset amount
	externalAmount, err := k.ConvertFromHODL(ctx, chainID, assetSymbol, netHODL)
	if err != nil {
		return 0, err
	}

	// Check rate limits
	if err := k.CheckRateLimit(ctx, chainID, assetSymbol, hodlAmount); err != nil {
		return 0, err
	}

	// Lock HODL in module account (escrow) - DO NOT burn yet
	// Coins will be burned only after successful TSS signing
	hodlCoin := sdk.NewCoin(hodltypes.HODLDenom, hodlAmount)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, sdk.NewCoins(hodlCoin)); err != nil {
		return 0, types.ErrInsufficientBalance
	}

	// Create withdrawal record
	withdrawalID := k.GetNextWithdrawalID(ctx)
	withdrawal := types.NewWithdrawal(
		withdrawalID,
		chainID,
		assetSymbol,
		sender.String(),
		recipient,
		hodlAmount,
		externalAmount,
		feeAmount,
		params.WithdrawalTimelockDuration(),
	)

	if err := k.SetWithdrawal(ctx, withdrawal); err != nil {
		return 0, err
	}

	// Update rate limit
	k.UpdateRateLimit(ctx, chainID, assetSymbol, hodlAmount)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"extbridge_withdrawal_requested",
			sdk.NewAttribute("withdrawal_id", fmt.Sprintf("%d", withdrawalID)),
			sdk.NewAttribute("chain_id", chainID),
			sdk.NewAttribute("asset", assetSymbol),
			sdk.NewAttribute("sender", sender.String()),
			sdk.NewAttribute("recipient", recipient),
			sdk.NewAttribute("hodl_burned", hodlAmount.String()),
			sdk.NewAttribute("external_amount", externalAmount.String()),
			sdk.NewAttribute("fee", feeAmount.String()),
			sdk.NewAttribute("timelock_expiry", withdrawal.TimelockExpiry.Format(time.RFC3339)),
		),
	)

	k.Logger(ctx).Info("withdrawal requested",
		"withdrawal_id", withdrawalID,
		"chain", chainID,
		"asset", assetSymbol,
		"amount", externalAmount.String(),
	)

	return withdrawalID, nil
}

// RefundWithdrawal refunds escrowed HODL for a failed or timed-out withdrawal
func (k Keeper) RefundWithdrawal(ctx sdk.Context, withdrawalID uint64) error {
	withdrawal, found := k.GetWithdrawal(ctx, withdrawalID)
	if !found {
		return types.ErrWithdrawalNotFound
	}

	// Only refund if withdrawal failed or timed out
	if withdrawal.Status != types.WithdrawalStatusFailed && withdrawal.Status != types.WithdrawalStatusTimeout {
		return fmt.Errorf("withdrawal cannot be refunded in current status: %s", withdrawal.Status)
	}

	// Refund escrowed HODL to sender
	senderAddr, err := sdk.AccAddressFromBech32(withdrawal.Sender)
	if err != nil {
		return err
	}

	hodlCoin := sdk.NewCoin(hodltypes.HODLDenom, withdrawal.HODLAmount)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, sdk.NewCoins(hodlCoin)); err != nil {
		k.Logger(ctx).Error("failed to refund HODL", "error", err)
		return err
	}

	// Update withdrawal status
	withdrawal.Status = types.WithdrawalStatusRefunded
	if err := k.SetWithdrawal(ctx, withdrawal); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"extbridge_withdrawal_refunded",
			sdk.NewAttribute("withdrawal_id", fmt.Sprintf("%d", withdrawalID)),
			sdk.NewAttribute("sender", withdrawal.Sender),
			sdk.NewAttribute("refunded_amount", withdrawal.HODLAmount.String()),
		),
	)

	k.Logger(ctx).Info("withdrawal refunded",
		"withdrawal_id", withdrawalID,
		"sender", withdrawal.Sender,
		"refunded_amount", withdrawal.HODLAmount.String(),
	)

	return nil
}

// ProcessWithdrawals processes withdrawals that are ready for TSS signing
func (k Keeper) ProcessWithdrawals(ctx sdk.Context) {
	withdrawals := k.GetAllWithdrawals(ctx)
	currentTime := ctx.BlockTime()

	for _, withdrawal := range withdrawals {
		// Only process pending or timelocked withdrawals
		if withdrawal.Status != types.WithdrawalStatusPending && withdrawal.Status != types.WithdrawalStatusTimelocked {
			continue
		}

		// Check if timelock has expired
		if !withdrawal.IsTimelockExpired(currentTime) {
			continue
		}

		// Update status to ready
		withdrawal.Status = types.WithdrawalStatusReady
		if err := k.SetWithdrawal(ctx, withdrawal); err != nil {
			k.Logger(ctx).Error("failed to update withdrawal status", "error", err)
			continue
		}

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"extbridge_withdrawal_ready",
				sdk.NewAttribute("withdrawal_id", fmt.Sprintf("%d", withdrawal.ID)),
				sdk.NewAttribute("chain_id", withdrawal.ChainID),
			),
		)
	}
}
