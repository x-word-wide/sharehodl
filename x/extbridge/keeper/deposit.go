package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

// =========================================================================
// Deposit Management
// =========================================================================

// GetDeposit retrieves a deposit by ID
func (k Keeper) GetDeposit(ctx sdk.Context, depositID uint64) (types.Deposit, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DepositKey(depositID))
	if bz == nil {
		return types.Deposit{}, false
	}

	var deposit types.Deposit
	if err := json.Unmarshal(bz, &deposit); err != nil {
		return types.Deposit{}, false
	}
	return deposit, true
}

// SetDeposit stores a deposit
func (k Keeper) SetDeposit(ctx sdk.Context, deposit types.Deposit) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(deposit)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal deposit", "error", err)
		return fmt.Errorf("failed to marshal deposit: %w", err)
	}
	store.Set(types.DepositKey(deposit.ID), bz)

	// Also store by tx hash for lookups
	txHashKey := types.DepositByTxHashKey(deposit.ChainID, deposit.ExternalTxHash)
	store.Set(txHashKey, bz)
	return nil
}

// GetDepositByTxHash retrieves a deposit by external tx hash
func (k Keeper) GetDepositByTxHash(ctx sdk.Context, chainID string, txHash string) (types.Deposit, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DepositByTxHashKey(chainID, txHash))
	if bz == nil {
		return types.Deposit{}, false
	}

	var deposit types.Deposit
	if err := json.Unmarshal(bz, &deposit); err != nil {
		return types.Deposit{}, false
	}
	return deposit, true
}

// GetAllDeposits returns all deposits
func (k Keeper) GetAllDeposits(ctx sdk.Context) []types.Deposit {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.DepositPrefix)
	defer iterator.Close()

	deposits := []types.Deposit{}
	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		if err := json.Unmarshal(iterator.Value(), &deposit); err != nil {
			continue
		}
		deposits = append(deposits, deposit)
	}
	return deposits
}

// ObserveDeposit creates a new deposit observation
func (k Keeper) ObserveDeposit(
	ctx sdk.Context,
	validator sdk.AccAddress,
	chainID string,
	assetSymbol string,
	externalTxHash string,
	externalBlockHeight uint64,
	sender string,
	recipient string,
	amount math.Int,
) (uint64, error) {
	// Check circuit breaker
	if err := k.IsOperationAllowed(ctx, "deposit"); err != nil {
		return 0, err
	}

	// Check if bridging is enabled
	params := k.GetParams(ctx)
	if !params.BridgingEnabled {
		return 0, types.ErrBridgingDisabled
	}

	// Verify validator eligibility
	if _, err := k.IsValidatorEligible(ctx, validator); err != nil {
		return 0, err
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

	// Check for duplicate deposit
	if _, exists := k.GetDepositByTxHash(ctx, chainID, externalTxHash); exists {
		return 0, types.ErrDuplicateDeposit
	}

	// Verify amount is within limits
	if amount.LT(chain.MinDeposit) {
		return 0, types.ErrAmountTooSmall
	}
	if amount.GT(chain.MaxDeposit) {
		return 0, types.ErrAmountTooLarge
	}

	// Check deposit rate limits (similar to withdrawal rate limits)
	if err := k.CheckDepositRateLimit(ctx, chainID, assetSymbol, amount); err != nil {
		return 0, err
	}

	// Convert to HODL amount
	hodlAmount, err := k.ConvertToHODL(ctx, chainID, assetSymbol, amount)
	if err != nil {
		return 0, err
	}

	// Calculate required attestations
	requiredAttestations, err := k.CalculateRequiredAttestations(ctx)
	if err != nil {
		return 0, err
	}

	// Create deposit record
	depositID := k.GetNextDepositID(ctx)
	deposit := types.NewDeposit(
		depositID,
		chainID,
		assetSymbol,
		externalTxHash,
		externalBlockHeight,
		sender,
		recipient,
		amount,
		hodlAmount,
		requiredAttestations,
		ctx.BlockTime(),
	)

	if err := k.SetDeposit(ctx, deposit); err != nil {
		return 0, err
	}

	// Update deposit rate limit
	k.UpdateDepositRateLimit(ctx, chainID, assetSymbol, amount)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"extbridge_deposit_observed",
			sdk.NewAttribute("deposit_id", fmt.Sprintf("%d", depositID)),
			sdk.NewAttribute("chain_id", chainID),
			sdk.NewAttribute("asset", assetSymbol),
			sdk.NewAttribute("tx_hash", externalTxHash),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("hodl_amount", hodlAmount.String()),
			sdk.NewAttribute("recipient", recipient),
			sdk.NewAttribute("validator", validator.String()),
		),
	)

	k.Logger(ctx).Info("deposit observed",
		"deposit_id", depositID,
		"chain", chainID,
		"asset", assetSymbol,
		"amount", amount.String(),
	)

	return depositID, nil
}

// AttestDeposit allows a validator to attest to a deposit
func (k Keeper) AttestDeposit(
	ctx sdk.Context,
	validator sdk.AccAddress,
	depositID uint64,
	approved bool,
	observedTxHash string,
	observedAmount math.Int,
) (uint64, bool, error) {
	// Check circuit breaker
	if err := k.IsOperationAllowed(ctx, "attest"); err != nil {
		return 0, false, err
	}

	// Verify validator eligibility
	if _, err := k.IsValidatorEligible(ctx, validator); err != nil {
		return 0, false, err
	}

	// Get deposit
	deposit, found := k.GetDeposit(ctx, depositID)
	if !found {
		return 0, false, types.ErrDepositNotFound
	}

	// Check if deposit can be attested
	if !deposit.CanAttest() {
		if deposit.IsCompleted() {
			return 0, false, types.ErrDepositCompleted
		}
		if deposit.IsRejected() {
			return 0, false, types.ErrDepositRejected
		}
		return 0, false, fmt.Errorf("deposit cannot be attested")
	}

	// Check if validator already attested
	if k.HasAttestation(ctx, depositID, validator) {
		return 0, false, types.ErrAlreadyAttested
	}

	// If approving, verify the observed values match the deposit record
	if approved {
		if observedTxHash != deposit.ExternalTxHash {
			return 0, false, fmt.Errorf("observed tx hash does not match deposit record")
		}
		if !observedAmount.Equal(deposit.Amount) {
			return 0, false, fmt.Errorf("observed amount does not match deposit record")
		}
	}

	// Create attestation
	attestation := types.NewDepositAttestation(
		depositID,
		validator.String(),
		deposit.ChainID,
		deposit.AssetSymbol,
		deposit.ExternalTxHash,
		deposit.Amount,
		approved,
		ctx.BlockTime(),
	)

	k.SetAttestation(ctx, attestation)

	// Update deposit attestation count
	if approved {
		deposit.Attestations++
		if deposit.Status == types.DepositStatusPending {
			deposit.Status = types.DepositStatusAttesting
		}
	}

	if err := k.SetDeposit(ctx, deposit); err != nil {
		return 0, false, err
	}

	// Check if we have enough attestations
	completed := false
	if deposit.HasEnoughAttestations() {
		if err := k.CompleteDeposit(ctx, depositID); err != nil {
			k.Logger(ctx).Error("failed to complete deposit", "deposit_id", depositID, "error", err)
			return deposit.Attestations, false, err
		}
		completed = true
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"extbridge_deposit_attested",
			sdk.NewAttribute("deposit_id", fmt.Sprintf("%d", depositID)),
			sdk.NewAttribute("validator", validator.String()),
			sdk.NewAttribute("approved", fmt.Sprintf("%t", approved)),
			sdk.NewAttribute("attestations", fmt.Sprintf("%d", deposit.Attestations)),
			sdk.NewAttribute("required", fmt.Sprintf("%d", deposit.RequiredAttestations)),
			sdk.NewAttribute("completed", fmt.Sprintf("%t", completed)),
		),
	)

	return deposit.Attestations, completed, nil
}

// CompleteDeposit completes a deposit and mints HODL
func (k Keeper) CompleteDeposit(ctx sdk.Context, depositID uint64) error {
	deposit, found := k.GetDeposit(ctx, depositID)
	if !found {
		return types.ErrDepositNotFound
	}

	if deposit.IsCompleted() {
		return types.ErrDepositCompleted
	}

	// Check if recipient is banned - funds should not be minted to banned addresses
	if banned, reason := k.IsAddressBanned(ctx, deposit.Recipient); banned {
		k.Logger(ctx).Warn("attempted deposit to banned address",
			"deposit_id", depositID,
			"recipient", deposit.Recipient,
			"reason", reason,
		)
		// Mark deposit as rejected instead of completing
		deposit.Status = types.DepositStatusRejected
		if err := k.SetDeposit(ctx, deposit); err != nil {
			return err
		}
		return types.ErrAddressBanned.Wrapf("recipient is banned: %s", reason)
	}

	// Mint HODL to recipient
	recipientAddr, err := sdk.AccAddressFromBech32(deposit.Recipient)
	if err != nil {
		return err
	}

	hodlCoin := sdk.NewCoin(hodltypes.HODLDenom, deposit.HODLAmount)
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(hodlCoin)); err != nil {
		k.Logger(ctx).Error("failed to mint HODL", "error", err)
		return types.ErrMintFailed
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, sdk.NewCoins(hodlCoin)); err != nil {
		k.Logger(ctx).Error("failed to send HODL", "error", err)
		// Rollback: burn the minted coins since we couldn't send them
		if burnErr := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(hodlCoin)); burnErr != nil {
			k.Logger(ctx).Error("failed to rollback minted HODL", "burn_error", burnErr, "original_error", err)
		}
		return types.ErrMintFailed
	}

	// Update deposit status
	now := ctx.BlockTime()
	deposit.Status = types.DepositStatusCompleted
	deposit.CompletedAt = &now
	if err := k.SetDeposit(ctx, deposit); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"extbridge_deposit_completed",
			sdk.NewAttribute("deposit_id", fmt.Sprintf("%d", depositID)),
			sdk.NewAttribute("recipient", deposit.Recipient),
			sdk.NewAttribute("hodl_minted", deposit.HODLAmount.String()),
			sdk.NewAttribute("external_amount", deposit.Amount.String()),
		),
	)

	k.Logger(ctx).Info("deposit completed",
		"deposit_id", depositID,
		"recipient", deposit.Recipient,
		"hodl_minted", deposit.HODLAmount.String(),
	)

	return nil
}

// =========================================================================
// Attestation Management
// =========================================================================

// SetAttestation stores an attestation
func (k Keeper) SetAttestation(ctx sdk.Context, attestation types.DepositAttestation) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(attestation)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal attestation", "error", err)
		return
	}
	store.Set(types.AttestationKey(attestation.DepositID, attestation.Validator), bz)
}

// GetAttestation retrieves an attestation
func (k Keeper) GetAttestation(ctx sdk.Context, depositID uint64, validator sdk.AccAddress) (types.DepositAttestation, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AttestationKey(depositID, validator.String()))
	if bz == nil {
		return types.DepositAttestation{}, false
	}

	var attestation types.DepositAttestation
	if err := json.Unmarshal(bz, &attestation); err != nil {
		return types.DepositAttestation{}, false
	}
	return attestation, true
}

// HasAttestation checks if a validator has attested to a deposit
func (k Keeper) HasAttestation(ctx sdk.Context, depositID uint64, validator sdk.AccAddress) bool {
	_, found := k.GetAttestation(ctx, depositID, validator)
	return found
}

// GetDepositAttestations returns all attestations for a deposit
func (k Keeper) GetDepositAttestations(ctx sdk.Context, depositID uint64) []types.DepositAttestation {
	// Note: This is a simplified implementation. In production, you'd want a better iteration strategy
	attestations := []types.DepositAttestation{}

	// Iterate through all eligible validators
	eligibleValidators, err := k.GetEligibleValidators(ctx)
	if err != nil {
		return attestations
	}

	for _, validator := range eligibleValidators {
		if attestation, found := k.GetAttestation(ctx, depositID, validator); found {
			attestations = append(attestations, attestation)
		}
	}

	return attestations
}
