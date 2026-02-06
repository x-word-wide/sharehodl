package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

// GetTSSSession retrieves a TSS session by ID
func (k Keeper) GetTSSSession(ctx sdk.Context, sessionID uint64) (types.TSSSession, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TSSSessionKey(sessionID))
	if bz == nil {
		return types.TSSSession{}, false
	}

	var session types.TSSSession
	if err := json.Unmarshal(bz, &session); err != nil {
		return types.TSSSession{}, false
	}
	return session, true
}

// SetTSSSession stores a TSS session
func (k Keeper) SetTSSSession(ctx sdk.Context, session types.TSSSession) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(session)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal TSS session", "error", err)
		return fmt.Errorf("failed to marshal TSS session: %w", err)
	}
	store.Set(types.TSSSessionKey(session.ID), bz)
	return nil
}

// GetAllTSSSessions returns all TSS sessions
func (k Keeper) GetAllTSSSessions(ctx sdk.Context) []types.TSSSession {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.TSSSessionPrefix)
	defer iterator.Close()

	sessions := []types.TSSSession{}
	for ; iterator.Valid(); iterator.Next() {
		var session types.TSSSession
		if err := json.Unmarshal(iterator.Value(), &session); err != nil {
			continue
		}
		sessions = append(sessions, session)
	}
	return sessions
}

// CreateTSSSession creates a new TSS signing session for a withdrawal
func (k Keeper) CreateTSSSession(ctx sdk.Context, withdrawalID uint64) (uint64, error) {
	withdrawal, found := k.GetWithdrawal(ctx, withdrawalID)
	if !found {
		return 0, types.ErrWithdrawalNotFound
	}

	if !withdrawal.CanSign(ctx.BlockTime()) {
		return 0, types.ErrWithdrawalNotReady
	}

	// Get eligible validators
	participants, err := k.GetEligibleValidators(ctx)
	if err != nil {
		return 0, err
	}

	if len(participants) == 0 {
		return 0, fmt.Errorf("no eligible validators for TSS")
	}

	// Calculate required signatures
	params := k.GetParams(ctx)
	requiredSigs := uint64(len(participants)) * params.TSSThreshold.MulInt64(100).TruncateInt().Uint64() / 100
	if requiredSigs == 0 {
		requiredSigs = 1
	}

	// Create message to sign (simplified - in production this would be the actual tx)
	message := []byte(fmt.Sprintf("withdrawal:%d:%s:%s:%s",
		withdrawalID,
		withdrawal.ChainID,
		withdrawal.Recipient,
		withdrawal.ExternalAmount.String(),
	))

	// Create session
	sessionID := k.GetNextTSSSessionID(ctx)
	participantAddrs := make([]string, len(participants))
	for i, p := range participants {
		participantAddrs[i] = p.String()
	}

	session := types.NewTSSSession(
		sessionID,
		withdrawalID,
		withdrawal.ChainID,
		participantAddrs,
		requiredSigs,
		1*time.Hour, // 1 hour timeout
		message,
		ctx.BlockTime(), // Use block time for determinism
	)

	if err := k.SetTSSSession(ctx, session); err != nil {
		return 0, err
	}

	// Update withdrawal
	withdrawal.Status = types.WithdrawalStatusSigning
	withdrawal.TSSSessionID = &sessionID
	if err := k.SetWithdrawal(ctx, withdrawal); err != nil {
		return 0, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"extbridge_tss_session_created",
			sdk.NewAttribute("session_id", fmt.Sprintf("%d", sessionID)),
			sdk.NewAttribute("withdrawal_id", fmt.Sprintf("%d", withdrawalID)),
			sdk.NewAttribute("participants", fmt.Sprintf("%d", len(participants))),
			sdk.NewAttribute("required_sigs", fmt.Sprintf("%d", requiredSigs)),
		),
	)

	return sessionID, nil
}

// SubmitTSSSignature submits a TSS signature share
func (k Keeper) SubmitTSSSignature(
	ctx sdk.Context,
	validator sdk.AccAddress,
	sessionID uint64,
	signatureData []byte,
) (bool, error) {
	// Verify validator eligibility
	if _, err := k.IsValidatorEligible(ctx, validator); err != nil {
		return false, err
	}

	// Get session
	session, found := k.GetTSSSession(ctx, sessionID)
	if !found {
		return false, types.ErrTSSSessionNotFound
	}

	// Check if session can accept signatures
	if !session.CanAddSignature() {
		if session.IsCompleted() {
			return false, types.ErrTSSSessionCompleted
		}
		if session.IsFailed() {
			return false, types.ErrTSSSessionFailed
		}
		return false, fmt.Errorf("session cannot accept signatures")
	}

	// Check timeout
	if session.IsTimeout(ctx.BlockTime()) {
		session.Status = types.TSSSessionStatusTimeout
		if err := k.SetTSSSession(ctx, session); err != nil {
			k.Logger(ctx).Error("failed to update TSS session on timeout", "error", err)
		}
		return false, types.ErrTSSTimeout
	}

	// Check if validator is a participant
	isParticipant := false
	for _, p := range session.Participants {
		if p == validator.String() {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return false, fmt.Errorf("validator not a participant in this session")
	}

	// Check for duplicate signature
	for _, share := range session.SignatureShares {
		if share.Validator == validator.String() {
			return false, fmt.Errorf("validator already submitted signature")
		}
	}

	// Add signature share
	share := types.NewTSSSignatureShare(sessionID, validator.String(), signatureData)
	session.SignatureShares = append(session.SignatureShares, share)
	session.ReceivedSigs++

	if session.Status == types.TSSSessionStatusPending {
		session.Status = types.TSSSessionStatusActive
	}

	// Check if we have enough signatures
	completed := false
	if session.HasEnoughSignatures() {
		// In production, this would combine the signature shares into a final signature
		session.CombinedSignature = []byte("combined_signature_placeholder")
		session.Status = types.TSSSessionStatusCompleted
		now := ctx.BlockTime()
		session.CompletedAt = &now
		completed = true

		// Update withdrawal status and burn escrowed HODL
		withdrawal, found := k.GetWithdrawal(ctx, session.WithdrawalID)
		if found {
			withdrawal.Status = types.WithdrawalStatusSigned
			if err := k.SetWithdrawal(ctx, withdrawal); err != nil {
				k.Logger(ctx).Error("failed to update withdrawal status after TSS", "error", err)
			}

			// Now burn the escrowed HODL since TSS signing succeeded
			hodlCoin := sdk.NewCoin(hodltypes.HODLDenom, withdrawal.HODLAmount)
			if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(hodlCoin)); err != nil {
				k.Logger(ctx).Error("failed to burn escrowed HODL after TSS completion", "error", err)
			}
		}
	}

	if err := k.SetTSSSession(ctx, session); err != nil {
		return false, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"extbridge_tss_signature_submitted",
			sdk.NewAttribute("session_id", fmt.Sprintf("%d", sessionID)),
			sdk.NewAttribute("validator", validator.String()),
			sdk.NewAttribute("received_sigs", fmt.Sprintf("%d", session.ReceivedSigs)),
			sdk.NewAttribute("required_sigs", fmt.Sprintf("%d", session.RequiredSigs)),
			sdk.NewAttribute("completed", fmt.Sprintf("%t", completed)),
		),
	)

	return completed, nil
}
