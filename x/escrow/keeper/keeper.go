package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/escrow/types"
)

// Keeper of the escrow store
type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       storetypes.StoreKey
	memKey         storetypes.StoreKey
	bankKeeper     types.BankKeeper
	accountKeeper  types.AccountKeeper
	equityKeeper   types.EquityKeeper
	stakingKeeper  types.UniversalStakingKeeper // For tier/reputation checks and validator oversight
}

// NewKeeper creates a new escrow Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	equityKeeper types.EquityKeeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		equityKeeper:  equityKeeper,
		stakingKeeper: nil, // Set later via SetStakingKeeper
	}
}

// SetStakingKeeper sets the staking keeper (for late binding during app initialization)
func (k *Keeper) SetStakingKeeper(stakingKeeper types.UniversalStakingKeeper) {
	k.stakingKeeper = stakingKeeper
}

// =============================================================================
// TIER CHECKS - Require Warden+ tier for moderator registration
// =============================================================================

// checkCanModerate verifies a user has the required tier to be a moderator
// Requires Warden+ tier (100K HODL)
func (k Keeper) checkCanModerate(ctx sdk.Context, address string) error {
	if k.stakingKeeper == nil {
		// SECURITY: Staking keeper is required - do not allow without tier checks
		return types.ErrStakingKeeperNotSet
	}

	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	if !k.stakingKeeper.CanModerate(ctx, addr) {
		return types.ErrInsufficientTierToModerate
	}

	return nil
}

// checkCanModerateLargeDispute verifies a user can handle large disputes
// Requires Steward+ tier (1M HODL) for disputes > 100K HODL
func (k Keeper) checkCanModerateLargeDispute(ctx sdk.Context, address string, disputeValue math.Int) error {
	if k.stakingKeeper == nil {
		// SECURITY: Staking keeper is required
		return types.ErrStakingKeeperNotSet
	}

	// Large disputes are > 100K HODL (100_000_000_000 uhodl)
	largeDisputeThreshold := math.NewInt(100_000_000_000)
	if disputeValue.LTE(largeDisputeThreshold) {
		return nil // Regular moderators can handle
	}

	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	if !k.stakingKeeper.CanModerateLargeDisputes(ctx, addr) {
		return types.ErrInsufficientTierForLargeDispute
	}

	return nil
}

// rewardModeratorReputation rewards a moderator for fair resolution
func (k Keeper) rewardModeratorReputation(ctx sdk.Context, moderator string, disputeID uint64) {
	if k.stakingKeeper == nil {
		return
	}

	addr, err := sdk.AccAddressFromBech32(moderator)
	if err != nil {
		return
	}

	if err := k.stakingKeeper.RewardSuccessfulDispute(ctx, addr, fmt.Sprintf("%d", disputeID)); err != nil {
		k.Logger(ctx).Error("failed to reward moderator", "error", err)
	}
}

// penalizeModeratorReputation penalizes a moderator for unfair resolution
func (k Keeper) penalizeModeratorReputation(ctx sdk.Context, moderator string, disputeID uint64) {
	if k.stakingKeeper == nil {
		return
	}

	addr, err := sdk.AccAddressFromBech32(moderator)
	if err != nil {
		return
	}

	if err := k.stakingKeeper.PenalizeBadDispute(ctx, addr, fmt.Sprintf("%d", disputeID)); err != nil {
		k.Logger(ctx).Error("failed to penalize moderator", "error", err)
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetNextEscrowID returns the next escrow ID and increments the counter
func (k Keeper) GetNextEscrowID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.EscrowCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.EscrowCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetNextDisputeID returns the next dispute ID and increments the counter
func (k Keeper) GetNextDisputeID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DisputeCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.DisputeCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// ============ Escrow Management ============

// GetEscrow returns an escrow by ID
func (k Keeper) GetEscrow(ctx sdk.Context, escrowID uint64) (types.Escrow, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetEscrowKey(escrowID))
	if bz == nil {
		return types.Escrow{}, false
	}

	var escrow types.Escrow
	if err := json.Unmarshal(bz, &escrow); err != nil {
		return types.Escrow{}, false
	}
	return escrow, true
}

// SetEscrow stores an escrow
// SECURITY FIX: Returns error instead of panicking
func (k Keeper) SetEscrow(ctx sdk.Context, escrow types.Escrow) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(escrow)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal escrow", "error", err)
		return fmt.Errorf("failed to marshal escrow: %w", err)
	}
	store.Set(types.GetEscrowKey(escrow.ID), bz)
	return nil
}

// DeleteEscrow removes an escrow
func (k Keeper) DeleteEscrow(ctx sdk.Context, escrowID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetEscrowKey(escrowID))
}

// CreateEscrow creates a new escrow agreement
func (k Keeper) CreateEscrow(
	ctx sdk.Context,
	sender, recipient string,
	moderator string,
	assets []types.EscrowAsset,
	description, terms string,
	expiresAt time.Time,
) (types.Escrow, error) {
	escrowID := k.GetNextEscrowID(ctx)

	escrow := types.NewEscrow(escrowID, sender, recipient, assets, description, terms, expiresAt, ctx.BlockTime())
	escrow.Moderator = moderator

	// Validate escrow
	if err := escrow.Validate(); err != nil {
		return types.Escrow{}, err
	}

	// Validate moderator if assigned
	if moderator != "" {
		mod, found := k.GetModerator(ctx, moderator)
		if !found {
			return types.Escrow{}, types.ErrModeratorNotFound
		}
		if !mod.Active {
			return types.Escrow{}, types.ErrModeratorNotActive
		}
		if mod.Blacklisted {
			return types.Escrow{}, types.ErrModeratorBlacklisted
		}

		// Stake-as-Trust-Ceiling: Check moderator can handle this escrow value
		// Trust ceiling = available stake - active dispute value
		if !mod.CanHandleDispute(escrow.TotalValue) {
			return types.Escrow{}, types.ErrDisputeExceedsTrustCeiling
		}

		// Update moderator's active dispute tracking
		mod.ActiveDisputeValue = mod.ActiveDisputeValue.Add(escrow.TotalValue)
		mod.ActiveDisputeCount++
		k.SetModerator(ctx, mod)
	}

	// Store escrow
	k.SetEscrow(ctx, escrow)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEscrowCreated,
			sdk.NewAttribute(types.AttributeKeyEscrowID, fmt.Sprintf("%d", escrow.ID)),
			sdk.NewAttribute(types.AttributeKeySender, sender),
			sdk.NewAttribute(types.AttributeKeyRecipient, recipient),
			sdk.NewAttribute(types.AttributeKeyModerator, moderator),
			sdk.NewAttribute(types.AttributeKeyAmount, escrow.TotalValue.String()),
		),
	)

	return escrow, nil
}

// FundEscrow funds an escrow with assets from the sender
func (k Keeper) FundEscrow(ctx sdk.Context, escrowID uint64, funder string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound
	}

	// Verify funder is the sender
	if escrow.Sender != funder {
		return types.ErrUnauthorized
	}

	// Check escrow status
	if escrow.Status != types.EscrowStatusPending {
		return types.ErrEscrowAlreadyFunded
	}

	// Check not expired
	if ctx.BlockTime().After(escrow.ExpiresAt) {
		return types.ErrEscrowExpired
	}

	funderAddr, err := sdk.AccAddressFromBech32(funder)
	if err != nil {
		return err
	}

	// Transfer assets to escrow module account
	for _, asset := range escrow.Assets {
		coins := sdk.NewCoins(sdk.NewCoin(asset.Denom, asset.Amount))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, funderAddr, types.ModuleName, coins); err != nil {
			return fmt.Errorf("failed to transfer %s: %w", asset.Denom, err)
		}

		// Register beneficial owner for equity assets
		// This ensures the sender continues to receive dividends while shares are in escrow
		if asset.AssetType == types.AssetTypeEquity && k.equityKeeper != nil {
			err := k.equityKeeper.RegisterBeneficialOwner(
				ctx,
				types.ModuleName,    // "escrow"
				asset.CompanyID,
				asset.ShareClass,
				escrow.Sender,       // The TRUE owner who deposited the shares
				asset.Amount,
				escrow.ID,           // Reference ID
				"escrow",            // Reference type
			)
			if err != nil {
				k.Logger(ctx).Error("failed to register beneficial owner",
					"escrow_id", escrow.ID,
					"company_id", asset.CompanyID,
					"class", asset.ShareClass,
					"owner", escrow.Sender,
					"error", err,
				)
				// Non-fatal: continue with escrow but log warning
				// The escrow still functions, but dividends may not be distributed correctly
			} else {
				k.Logger(ctx).Info("registered beneficial owner for escrowed shares",
					"escrow_id", escrow.ID,
					"company_id", asset.CompanyID,
					"class", asset.ShareClass,
					"owner", escrow.Sender,
					"shares", asset.Amount.String(),
				)
			}
		}
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusFunded
	escrow.FundedAt = ctx.BlockTime()
	k.SetEscrow(ctx, escrow)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEscrowFunded,
			sdk.NewAttribute(types.AttributeKeyEscrowID, fmt.Sprintf("%d", escrow.ID)),
			sdk.NewAttribute(types.AttributeKeySender, funder),
		),
	)

	return nil
}

// ReleaseEscrow releases escrow funds to the recipient
func (k Keeper) ReleaseEscrow(ctx sdk.Context, escrowID uint64, releaser string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound
	}

	// Only sender or moderator can release
	if escrow.Sender != releaser && escrow.Moderator != releaser {
		return types.ErrUnauthorized
	}

	// Check escrow is funded and not disputed
	if escrow.Status != types.EscrowStatusFunded {
		if escrow.Status == types.EscrowStatusDisputed {
			return types.ErrEscrowDisputed
		}
		return types.ErrEscrowNotFunded
	}

	recipientAddr, err := sdk.AccAddressFromBech32(escrow.Recipient)
	if err != nil {
		return err
	}

	// Calculate fees
	feeAmount := escrow.TotalValue.Mul(escrow.EscrowFee).TruncateInt()

	// Transfer assets to recipient (minus fees)
	for _, asset := range escrow.Assets {
		transferAmount := asset.Amount
		if asset.AssetType == types.AssetTypeHODL {
			// Deduct proportional fee from HODL
			assetFee := feeAmount.Mul(asset.Amount).Quo(escrow.TotalValue.TruncateInt())
			transferAmount = asset.Amount.Sub(assetFee)
		}

		if transferAmount.IsPositive() {
			coins := sdk.NewCoins(sdk.NewCoin(asset.Denom, transferAmount))
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, coins); err != nil {
				return fmt.Errorf("failed to release %s: %w", asset.Denom, err)
			}
		}

		// Unregister beneficial owner for equity assets
		// The shares are now leaving escrow and going to the recipient
		if asset.AssetType == types.AssetTypeEquity && k.equityKeeper != nil {
			err := k.equityKeeper.UnregisterBeneficialOwner(
				ctx,
				types.ModuleName,
				asset.CompanyID,
				asset.ShareClass,
				escrow.Sender,  // Original owner who deposited
				escrow.ID,
			)
			if err != nil {
				k.Logger(ctx).Error("failed to unregister beneficial owner",
					"escrow_id", escrow.ID,
					"company_id", asset.CompanyID,
					"class", asset.ShareClass,
					"owner", escrow.Sender,
					"error", err,
				)
				// Non-fatal: continue with release but log warning
			} else {
				k.Logger(ctx).Info("unregistered beneficial owner for released shares",
					"escrow_id", escrow.ID,
					"company_id", asset.CompanyID,
					"class", asset.ShareClass,
					"owner", escrow.Sender,
				)
			}
		}
	}

	// Pay moderator fee if moderator is assigned
	if escrow.Moderator != "" && escrow.ModeratorFee.IsPositive() {
		modFee := escrow.TotalValue.Mul(escrow.ModeratorFee).TruncateInt()
		if modFee.IsPositive() {
			modAddr, _ := sdk.AccAddressFromBech32(escrow.Moderator)
			modCoins := sdk.NewCoins(sdk.NewCoin("hodl", modFee))
			k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, modAddr, modCoins)
		}
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusReleased
	escrow.CompletedAt = ctx.BlockTime()
	k.SetEscrow(ctx, escrow)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEscrowReleased,
			sdk.NewAttribute(types.AttributeKeyEscrowID, fmt.Sprintf("%d", escrow.ID)),
			sdk.NewAttribute(types.AttributeKeyRecipient, escrow.Recipient),
			sdk.NewAttribute(types.AttributeKeyAmount, escrow.TotalValue.String()),
		),
	)

	return nil
}

// RefundEscrow refunds escrow funds to the sender
func (k Keeper) RefundEscrow(ctx sdk.Context, escrowID uint64, refunder string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound
	}

	// Only recipient or moderator can refund
	if escrow.Recipient != refunder && escrow.Moderator != refunder {
		return types.ErrUnauthorized
	}

	// Check escrow is funded and not disputed
	if escrow.Status != types.EscrowStatusFunded {
		if escrow.Status == types.EscrowStatusDisputed {
			return types.ErrEscrowDisputed
		}
		return types.ErrEscrowNotFunded
	}

	senderAddr, err := sdk.AccAddressFromBech32(escrow.Sender)
	if err != nil {
		return err
	}

	// Transfer all assets back to sender
	for _, asset := range escrow.Assets {
		coins := sdk.NewCoins(sdk.NewCoin(asset.Denom, asset.Amount))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, coins); err != nil {
			return fmt.Errorf("failed to refund %s: %w", asset.Denom, err)
		}

		// Unregister beneficial owner for equity assets
		// The shares are being returned to the original sender
		if asset.AssetType == types.AssetTypeEquity && k.equityKeeper != nil {
			err := k.equityKeeper.UnregisterBeneficialOwner(
				ctx,
				types.ModuleName,
				asset.CompanyID,
				asset.ShareClass,
				escrow.Sender,  // Original owner
				escrow.ID,
			)
			if err != nil {
				k.Logger(ctx).Error("failed to unregister beneficial owner on refund",
					"escrow_id", escrow.ID,
					"company_id", asset.CompanyID,
					"class", asset.ShareClass,
					"owner", escrow.Sender,
					"error", err,
				)
				// Non-fatal: continue with refund but log warning
			} else {
				k.Logger(ctx).Info("unregistered beneficial owner for refunded shares",
					"escrow_id", escrow.ID,
					"company_id", asset.CompanyID,
					"class", asset.ShareClass,
					"owner", escrow.Sender,
				)
			}
		}
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusRefunded
	escrow.CompletedAt = ctx.BlockTime()
	k.SetEscrow(ctx, escrow)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEscrowRefunded,
			sdk.NewAttribute(types.AttributeKeyEscrowID, fmt.Sprintf("%d", escrow.ID)),
			sdk.NewAttribute(types.AttributeKeySender, escrow.Sender),
		),
	)

	return nil
}

// CancelEscrow cancels a pending (unfunded) escrow
func (k Keeper) CancelEscrow(ctx sdk.Context, escrowID uint64, canceller string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound
	}

	// Only sender can cancel
	if escrow.Sender != canceller {
		return types.ErrUnauthorized
	}

	// Can only cancel pending escrows
	if escrow.Status != types.EscrowStatusPending {
		return types.ErrInvalidStatus
	}

	escrow.Status = types.EscrowStatusCancelled
	escrow.CompletedAt = ctx.BlockTime()
	k.SetEscrow(ctx, escrow)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEscrowCancelled,
			sdk.NewAttribute(types.AttributeKeyEscrowID, fmt.Sprintf("%d", escrow.ID)),
		),
	)

	return nil
}

// ConfirmEscrow confirms readiness from sender or recipient
func (k Keeper) ConfirmEscrow(ctx sdk.Context, escrowID uint64, confirmer string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound
	}

	if escrow.Status != types.EscrowStatusFunded {
		return types.ErrEscrowNotFunded
	}

	if confirmer == escrow.Sender {
		escrow.SenderConfirmed = true
	} else if confirmer == escrow.Recipient {
		escrow.RecipientConfirmed = true
	} else {
		return types.ErrNotParticipant
	}

	// If both confirmed, auto-release
	if escrow.SenderConfirmed && escrow.RecipientConfirmed {
		k.SetEscrow(ctx, escrow)
		return k.ReleaseEscrow(ctx, escrowID, escrow.Sender)
	}

	k.SetEscrow(ctx, escrow)
	return nil
}

// ============ Dispute Management ============

// GetDispute returns a dispute by ID
func (k Keeper) GetDispute(ctx sdk.Context, disputeID uint64) (types.Dispute, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDisputeKey(disputeID))
	if bz == nil {
		return types.Dispute{}, false
	}

	var dispute types.Dispute
	if err := json.Unmarshal(bz, &dispute); err != nil {
		return types.Dispute{}, false
	}
	return dispute, true
}

// SetDispute stores a dispute
// SECURITY FIX: Returns error instead of panicking
func (k Keeper) SetDispute(ctx sdk.Context, dispute types.Dispute) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(dispute)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal dispute", "error", err)
		return fmt.Errorf("failed to marshal dispute: %w", err)
	}
	store.Set(types.GetDisputeKey(dispute.ID), bz)
	return nil
}

// GetEscrowDispute returns the dispute for an escrow
func (k Keeper) GetEscrowDispute(ctx sdk.Context, escrowID uint64) (types.Dispute, bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.DisputePrefix).Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var dispute types.Dispute
		if err := json.Unmarshal(iterator.Value(), &dispute); err != nil {
			continue
		}
		if dispute.EscrowID == escrowID {
			return dispute, true
		}
	}
	return types.Dispute{}, false
}

// OpenDispute opens a new dispute on an escrow
func (k Keeper) OpenDispute(ctx sdk.Context, escrowID uint64, initiator, reason string) (types.Dispute, error) {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.Dispute{}, types.ErrEscrowNotFound
	}

	// Verify initiator is a participant
	if initiator != escrow.Sender && initiator != escrow.Recipient {
		return types.Dispute{}, types.ErrNotParticipant
	}

	// Check escrow status
	if escrow.Status != types.EscrowStatusFunded {
		return types.Dispute{}, types.ErrEscrowNotFunded
	}

	// Check no existing dispute
	if _, found := k.GetEscrowDispute(ctx, escrowID); found {
		return types.Dispute{}, types.ErrDisputeAlreadyExists
	}

	// Create dispute
	disputeID := k.GetNextDisputeID(ctx)
	deadline := ctx.BlockTime().Add(7 * 24 * time.Hour) // 7 days for resolution
	dispute := types.NewDispute(disputeID, escrowID, initiator, reason, deadline, ctx.BlockTime())

	// Validate
	if err := dispute.Validate(); err != nil {
		return types.Dispute{}, err
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusDisputed
	k.SetEscrow(ctx, escrow)

	// Store dispute
	k.SetDispute(ctx, dispute)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDisputeOpened,
			sdk.NewAttribute(types.AttributeKeyDisputeID, fmt.Sprintf("%d", dispute.ID)),
			sdk.NewAttribute(types.AttributeKeyEscrowID, fmt.Sprintf("%d", escrowID)),
			sdk.NewAttribute("initiator", initiator),
			sdk.NewAttribute("reason", reason),
		),
	)

	return dispute, nil
}

// SubmitEvidence submits evidence to a dispute
func (k Keeper) SubmitEvidence(ctx sdk.Context, disputeID uint64, submitter, hash, description string) error {
	dispute, found := k.GetDispute(ctx, disputeID)
	if !found {
		return types.ErrDisputeNotFound
	}

	// Verify submitter is a participant
	escrow, _ := k.GetEscrow(ctx, dispute.EscrowID)
	if submitter != escrow.Sender && submitter != escrow.Recipient {
		return types.ErrNotParticipant
	}

	// Check dispute is still open
	if dispute.Status == types.DisputeStatusResolved || dispute.Status == types.DisputeStatusFinal {
		return types.ErrDisputeResolved
	}

	// Add evidence
	evidence := types.Evidence{
		Submitter:   submitter,
		Hash:        hash,
		Description: description,
		SubmittedAt: ctx.BlockTime(),
	}
	dispute.Evidence = append(dispute.Evidence, evidence)
	k.SetDispute(ctx, dispute)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEvidenceSubmitted,
			sdk.NewAttribute(types.AttributeKeyDisputeID, fmt.Sprintf("%d", disputeID)),
			sdk.NewAttribute("submitter", submitter),
		),
	)

	return nil
}

// VoteOnDispute allows a moderator to vote on a dispute
func (k Keeper) VoteOnDispute(
	ctx sdk.Context,
	disputeID uint64,
	moderator string,
	vote types.DisputeResolution,
	reason string,
) error {
	dispute, found := k.GetDispute(ctx, disputeID)
	if !found {
		return types.ErrDisputeNotFound
	}

	// Verify moderator exists and is active
	mod, found := k.GetModerator(ctx, moderator)
	if !found {
		return types.ErrModeratorNotFound
	}
	if !mod.Active {
		return types.ErrModeratorNotActive
	}

	// TIER CHECK: For large disputes, moderator must be Steward+ tier
	escrow, found := k.GetEscrow(ctx, dispute.EscrowID)
	if found {
		// Convert LegacyDec to Int for tier check
		escrowValue := escrow.TotalValue.TruncateInt()
		if err := k.checkCanModerateLargeDispute(ctx, moderator, escrowValue); err != nil {
			return err
		}
	}

	// Check dispute is in voting status
	if dispute.Status != types.DisputeStatusOpen && dispute.Status != types.DisputeStatusVoting {
		return types.ErrDisputeResolved
	}

	// Check moderator hasn't already voted
	for _, v := range dispute.Votes {
		if v.Moderator == moderator {
			return types.ErrAlreadyVoted
		}
	}

	// Check deadline
	if ctx.BlockTime().After(dispute.DeadlineAt) {
		return types.ErrDisputeDeadlinePassed
	}

	// Add vote
	modVote := types.ModeratorVote{
		Moderator: moderator,
		Vote:      vote,
		Reason:    reason,
		VotedAt:   ctx.BlockTime(),
	}
	dispute.Votes = append(dispute.Votes, modVote)
	dispute.Status = types.DisputeStatusVoting

	// Check if we have enough votes
	if len(dispute.Votes) >= dispute.VotesRequired {
		k.resolveDispute(ctx, &dispute)
	}

	k.SetDispute(ctx, dispute)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorVoted,
			sdk.NewAttribute(types.AttributeKeyDisputeID, fmt.Sprintf("%d", disputeID)),
			sdk.NewAttribute(types.AttributeKeyModerator, moderator),
			sdk.NewAttribute("vote", vote.String()),
		),
	)

	return nil
}

// resolveDispute resolves a dispute based on votes
func (k Keeper) resolveDispute(ctx sdk.Context, dispute *types.Dispute) {
	// Count votes
	voteCounts := make(map[types.DisputeResolution]int)
	for _, v := range dispute.Votes {
		voteCounts[v.Vote]++
	}

	// Find winning resolution
	var maxVotes int
	var winningResolution types.DisputeResolution
	for resolution, count := range voteCounts {
		if count > maxVotes {
			maxVotes = count
			winningResolution = resolution
		}
	}

	dispute.Resolution = winningResolution
	dispute.Status = types.DisputeStatusResolved
	dispute.ResolvedAt = ctx.BlockTime()

	// Execute resolution
	escrow, _ := k.GetEscrow(ctx, dispute.EscrowID)
	switch winningResolution {
	case types.DisputeResolutionReleaseBuyer:
		// Release to recipient (buyer in P2P trade)
		k.ReleaseEscrow(ctx, dispute.EscrowID, escrow.Moderator)
	case types.DisputeResolutionReleaseSeller, types.DisputeResolutionRefund:
		// Refund to sender (seller)
		k.RefundEscrow(ctx, dispute.EscrowID, escrow.Moderator)
	case types.DisputeResolutionSplit:
		// Split funds - handled separately
		k.splitEscrowFunds(ctx, escrow, dispute)
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusResolved
	k.SetEscrow(ctx, escrow)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDisputeResolved,
			sdk.NewAttribute(types.AttributeKeyDisputeID, fmt.Sprintf("%d", dispute.ID)),
			sdk.NewAttribute(types.AttributeKeyResolution, winningResolution.String()),
		),
	)
}

// splitEscrowFunds splits escrow funds between parties
func (k Keeper) splitEscrowFunds(ctx sdk.Context, escrow types.Escrow, dispute *types.Dispute) error {
	senderAddr, _ := sdk.AccAddressFromBech32(escrow.Sender)
	recipientAddr, _ := sdk.AccAddressFromBech32(escrow.Recipient)

	// Default to 50/50 split if not specified
	senderShare := math.LegacyNewDecWithPrec(5, 1) // 0.5
	recipientShare := math.LegacyNewDecWithPrec(5, 1)

	if dispute.SenderAmount.IsPositive() && dispute.RecipientAmount.IsPositive() {
		total := dispute.SenderAmount.Add(dispute.RecipientAmount)
		senderShare = dispute.SenderAmount.Quo(total)
		recipientShare = dispute.RecipientAmount.Quo(total)
	}

	for _, asset := range escrow.Assets {
		senderAmount := senderShare.MulInt(asset.Amount).TruncateInt()
		recipientAmount := recipientShare.MulInt(asset.Amount).TruncateInt()

		if senderAmount.IsPositive() {
			senderCoins := sdk.NewCoins(sdk.NewCoin(asset.Denom, senderAmount))
			k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, senderCoins)
		}

		if recipientAmount.IsPositive() {
			recipientCoins := sdk.NewCoins(sdk.NewCoin(asset.Denom, recipientAmount))
			k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, recipientCoins)
		}

		// Unregister beneficial owner for equity assets after split
		// The original beneficial ownership is now invalid as shares are split between parties
		if asset.AssetType == types.AssetTypeEquity && k.equityKeeper != nil {
			err := k.equityKeeper.UnregisterBeneficialOwner(
				ctx,
				types.ModuleName,
				asset.CompanyID,
				asset.ShareClass,
				escrow.Sender,  // Original beneficial owner
				escrow.ID,
			)
			if err != nil {
				k.Logger(ctx).Error("failed to unregister beneficial owner after split",
					"escrow_id", escrow.ID,
					"company_id", asset.CompanyID,
					"class", asset.ShareClass,
					"owner", escrow.Sender,
					"error", err,
				)
				// Non-fatal: continue but log warning
			} else {
				k.Logger(ctx).Info("unregistered beneficial owner after split resolution",
					"escrow_id", escrow.ID,
					"company_id", asset.CompanyID,
					"class", asset.ShareClass,
					"sender_amount", senderAmount.String(),
					"recipient_amount", recipientAmount.String(),
				)
			}
		}
	}

	return nil
}

// AppealDispute appeals a dispute resolution
func (k Keeper) AppealDispute(ctx sdk.Context, disputeID uint64, appellant, reason string) error {
	dispute, found := k.GetDispute(ctx, disputeID)
	if !found {
		return types.ErrDisputeNotFound
	}

	// Check can appeal
	if dispute.Status != types.DisputeStatusResolved {
		return types.ErrCannotAppeal
	}
	if dispute.AppealsCount >= dispute.MaxAppeals {
		return types.ErrMaxAppealsReached
	}

	// Verify appellant is participant
	escrow, _ := k.GetEscrow(ctx, dispute.EscrowID)
	if appellant != escrow.Sender && appellant != escrow.Recipient {
		return types.ErrNotParticipant
	}

	// Reset for new round of voting
	dispute.Status = types.DisputeStatusAppealed
	dispute.AppealsCount++
	dispute.Votes = []types.ModeratorVote{} // Clear previous votes
	dispute.DeadlineAt = ctx.BlockTime().Add(7 * 24 * time.Hour)
	dispute.VotesRequired = 5 // Require more votes for appeal

	k.SetDispute(ctx, dispute)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDisputeAppealed,
			sdk.NewAttribute(types.AttributeKeyDisputeID, fmt.Sprintf("%d", disputeID)),
			sdk.NewAttribute("appellant", appellant),
		),
	)

	return nil
}

// ============ Moderator Management ============

// GetModerator returns a moderator by address
func (k Keeper) GetModerator(ctx sdk.Context, address string) (types.Moderator, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetModeratorKey(address))
	if bz == nil {
		return types.Moderator{}, false
	}

	var mod types.Moderator
	if err := json.Unmarshal(bz, &mod); err != nil {
		return types.Moderator{}, false
	}
	return mod, true
}

// SetModerator stores a moderator
// SECURITY FIX: Returns error instead of panicking
func (k Keeper) SetModerator(ctx sdk.Context, mod types.Moderator) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(mod)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal moderator", "error", err)
		return fmt.Errorf("failed to marshal moderator: %w", err)
	}
	store.Set(types.GetModeratorKey(mod.Address), bz)
	return nil
}

// RegisterModerator registers a new moderator
func (k Keeper) RegisterModerator(ctx sdk.Context, address string, stake math.Int) error {
	// TIER CHECK: Must be Warden+ tier (100K HODL) in universal staking to be moderator
	if err := k.checkCanModerate(ctx, address); err != nil {
		return err
	}

	// Check not already registered
	if _, found := k.GetModerator(ctx, address); found {
		return types.ErrModeratorAlreadyExists
	}

	// Minimum stake requirement (1000 HODL) - additional escrow-specific stake
	// Note: This is separate from universal staking tier requirement
	minStake := math.NewInt(1000000000) // 1000 HODL in uhodl
	if stake.LT(minStake) {
		return types.ErrInsufficientStake
	}

	// Transfer stake to module
	modAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}
	stakeCoins := sdk.NewCoins(sdk.NewCoin("hodl", stake))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, modAddr, types.ModuleName, stakeCoins); err != nil {
		return err
	}

	// Create moderator
	mod := types.NewModerator(address, stake, ctx.BlockTime())
	k.SetModerator(ctx, mod)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorRegistered,
			sdk.NewAttribute(types.AttributeKeyModerator, address),
			sdk.NewAttribute("stake", stake.String()),
		),
	)

	return nil
}

// UpdateModeratorStats updates moderator statistics after dispute resolution
func (k Keeper) UpdateModeratorStats(ctx sdk.Context, moderator string, successful bool) {
	mod, found := k.GetModerator(ctx, moderator)
	if !found {
		return
	}

	mod.DisputesHandled++

	// Update reputation
	if successful {
		// Increase reputation (max 100)
		mod.ReputationScore = mod.ReputationScore.Add(math.LegacyNewDec(1))
		if mod.ReputationScore.GT(math.LegacyNewDec(100)) {
			mod.ReputationScore = math.LegacyNewDec(100)
		}
	} else {
		// Decrease reputation
		mod.ReputationScore = mod.ReputationScore.Sub(math.LegacyNewDec(2))
		if mod.ReputationScore.LT(math.LegacyZeroDec()) {
			mod.ReputationScore = math.LegacyZeroDec()
		}
	}

	// Update success rate
	// (This is simplified - in production would track successful vs total)

	// Update tier based on disputes handled and reputation
	k.updateModeratorTier(ctx, &mod)

	k.SetModerator(ctx, mod)
}

// updateModeratorTier updates a moderator's tier based on performance
func (k Keeper) updateModeratorTier(ctx sdk.Context, mod *types.Moderator) {
	if mod.DisputesHandled >= 200 && mod.ReputationScore.GTE(math.LegacyNewDec(95)) {
		mod.Tier = types.ModeratorTierPlatinum
		mod.MaxEscrowValue = math.LegacyNewDec(1000000) // 1M HODL
	} else if mod.DisputesHandled >= 50 && mod.ReputationScore.GTE(math.LegacyNewDec(90)) {
		mod.Tier = types.ModeratorTierGold
		mod.MaxEscrowValue = math.LegacyNewDec(100000) // 100k HODL
	} else if mod.DisputesHandled >= 10 && mod.ReputationScore.GTE(math.LegacyNewDec(80)) {
		mod.Tier = types.ModeratorTierSilver
		mod.MaxEscrowValue = math.LegacyNewDec(50000) // 50k HODL
	} else {
		mod.Tier = types.ModeratorTierBronze
		mod.MaxEscrowValue = math.LegacyNewDec(10000) // 10k HODL
	}
}

// ============ Expiry Handling ============

// ProcessExpiredEscrows processes all expired escrows
func (k Keeper) ProcessExpiredEscrows(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.EscrowPrefix).Iterator(nil, nil)
	defer iterator.Close()

	currentTime := ctx.BlockTime()

	for ; iterator.Valid(); iterator.Next() {
		var escrow types.Escrow
		if err := json.Unmarshal(iterator.Value(), &escrow); err != nil {
			continue
		}

		// Check if expired and still pending or funded
		if currentTime.After(escrow.ExpiresAt) {
			if escrow.Status == types.EscrowStatusPending {
				escrow.Status = types.EscrowStatusExpired
				k.SetEscrow(ctx, escrow)
			} else if escrow.Status == types.EscrowStatusFunded {
				// Auto-refund expired funded escrows
				k.RefundEscrow(ctx, escrow.ID, escrow.Recipient)
				escrow.Status = types.EscrowStatusExpired
				k.SetEscrow(ctx, escrow)
			}
		}
	}
}

// GetAllModerators returns all registered moderators
func (k Keeper) GetAllModerators(ctx sdk.Context) []types.Moderator {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.ModeratorPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var moderators []types.Moderator
	for ; iterator.Valid(); iterator.Next() {
		var mod types.Moderator
		if err := json.Unmarshal(iterator.Value(), &mod); err != nil {
			continue
		}
		moderators = append(moderators, mod)
	}

	return moderators
}

// GetActiveModerators returns all active moderators
func (k Keeper) GetActiveModerators(ctx sdk.Context) []types.Moderator {
	allMods := k.GetAllModerators(ctx)
	var active []types.Moderator
	for _, mod := range allMods {
		if mod.Active && !mod.Blacklisted {
			active = append(active, mod)
		}
	}
	return active
}

// HasActiveModerators checks if there are any active moderators available
// Returns false if all moderators are blacklisted or inactive
func (k Keeper) HasActiveModerators(ctx sdk.Context) bool {
	return len(k.GetActiveModerators(ctx)) > 0
}

// GetMinimumModeratorsForDispute returns the minimum number of moderators needed
// If we're below this threshold, disputes should auto-escalate to governance
func (k Keeper) GetMinimumModeratorsForDispute(disputeValue math.Int) int {
	// Standard disputes need 3 moderators
	// Large disputes (>100K HODL) need 5 moderators
	largeThreshold := math.NewInt(100_000_000_000)
	if disputeValue.GT(largeThreshold) {
		return 5
	}
	return 3
}

// CheckModeratorAvailability checks if enough moderators are available for a dispute
// Returns (hasEnough, available, required)
func (k Keeper) CheckModeratorAvailability(ctx sdk.Context, disputeValue math.Int) (bool, int, int) {
	active := k.GetActiveModerators(ctx)
	required := k.GetMinimumModeratorsForDispute(disputeValue)
	return len(active) >= required, len(active), required
}

// EmergencyDisputeResolution handles disputes when no moderators are available
// This auto-splits funds 50/50 and emits an emergency event
func (k Keeper) EmergencyDisputeResolution(ctx sdk.Context, escrowID uint64, reason string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound
	}

	senderAddr, err := sdk.AccAddressFromBech32(escrow.Sender)
	if err != nil {
		return err
	}
	recipientAddr, err := sdk.AccAddressFromBech32(escrow.Recipient)
	if err != nil {
		return err
	}

	// Split each asset 50/50 as emergency fallback
	for _, asset := range escrow.Assets {
		if asset.AssetType == types.AssetTypeHODL {
			// Split HODL tokens
			halfAmount := asset.Amount.Quo(math.NewInt(2))
			if halfAmount.IsPositive() {
				senderCoins := sdk.NewCoins(sdk.NewCoin(asset.Denom, halfAmount))
				if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, senderAddr, senderCoins); err != nil {
					return err
				}

				recipientCoins := sdk.NewCoins(sdk.NewCoin(asset.Denom, halfAmount))
				if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, recipientCoins); err != nil {
					return err
				}
			}
		} else if asset.AssetType == types.AssetTypeEquity && k.equityKeeper != nil {
			// Split equity shares
			halfShares := asset.Amount.Quo(math.NewInt(2))
			if halfShares.IsPositive() {
				// Transfer half to sender
				if err := k.equityKeeper.TransferShares(ctx, asset.CompanyID, asset.ShareClass,
					types.ModuleName, escrow.Sender, halfShares); err != nil {
					k.Logger(ctx).Error("emergency equity transfer to sender failed", "error", err)
				}
				// Transfer half to recipient
				if err := k.equityKeeper.TransferShares(ctx, asset.CompanyID, asset.ShareClass,
					types.ModuleName, escrow.Recipient, halfShares); err != nil {
					k.Logger(ctx).Error("emergency equity transfer to recipient failed", "error", err)
				}

				// Unregister beneficial owner after emergency split
				if err := k.equityKeeper.UnregisterBeneficialOwner(
					ctx,
					types.ModuleName,
					asset.CompanyID,
					asset.ShareClass,
					escrow.Sender,
					escrowID,
				); err != nil {
					k.Logger(ctx).Error("failed to unregister beneficial owner in emergency resolution",
						"escrow_id", escrowID,
						"error", err,
					)
				}
			}
		}
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusResolved
	k.SetEscrow(ctx, escrow)

	// Emit emergency event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"emergency_dispute_resolution",
			sdk.NewAttribute(types.AttributeKeyEscrowID, fmt.Sprintf("%d", escrowID)),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
			sdk.NewAttribute(types.AttributeKeyResolution, "50/50 split - no moderators available"),
		),
	)

	k.Logger(ctx).Warn("emergency dispute resolution executed",
		"escrow_id", escrowID,
		"reason", reason,
		"resolution", "50/50 split")

	return nil
}

// ============ Stake-as-Trust-Ceiling: Unbonding Management ============

// GetUnbondingModerator returns an unbonding moderator record
func (k Keeper) GetUnbondingModerator(ctx sdk.Context, address string) (types.UnbondingModerator, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetUnbondingModeratorKey(address))
	if bz == nil {
		return types.UnbondingModerator{}, false
	}

	var unbonding types.UnbondingModerator
	if err := json.Unmarshal(bz, &unbonding); err != nil {
		return types.UnbondingModerator{}, false
	}
	return unbonding, true
}

// SetUnbondingModerator stores an unbonding moderator record
func (k Keeper) SetUnbondingModerator(ctx sdk.Context, unbonding types.UnbondingModerator) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(unbonding)
	if err != nil {
		return fmt.Errorf("failed to marshal unbonding moderator: %w", err)
	}
	store.Set(types.GetUnbondingModeratorKey(unbonding.Address), bz)
	return nil
}

// DeleteUnbondingModerator removes an unbonding moderator record
func (k Keeper) DeleteUnbondingModerator(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetUnbondingModeratorKey(address))
}

// RequestUnstake initiates the 14-day unbonding period for a moderator
func (k Keeper) RequestUnstake(ctx sdk.Context, address string, amount math.Int) error {
	mod, found := k.GetModerator(ctx, address)
	if !found {
		return types.ErrModeratorNotFound
	}

	// Check moderator is not blacklisted
	if mod.Blacklisted {
		return types.ErrModeratorBlacklisted
	}

	// Check no active disputes
	if mod.ActiveDisputeCount > 0 {
		return types.ErrModeratorHasActiveDisputes
	}

	// Check not already unbonding
	if _, found := k.GetUnbondingModerator(ctx, address); found {
		return types.ErrUnbondingInProgress
	}

	// Validate amount
	availableStake := mod.GetAvailableStake()
	if amount.GT(availableStake) {
		return types.ErrInsufficientStake
	}

	// Create unbonding record
	unbonding := types.NewUnbondingModerator(address, amount, ctx.BlockTime())
	if err := k.SetUnbondingModerator(ctx, unbonding); err != nil {
		return err
	}

	// Update moderator unbonding stake
	mod.UnbondingStake = amount
	k.SetModerator(ctx, mod)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorUnstakeRequested,
			sdk.NewAttribute(types.AttributeKeyModerator, address),
			sdk.NewAttribute(types.AttributeKeyUnbondingAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletesAt, unbonding.CompletesAt.String()),
		),
	)

	return nil
}

// CompleteUnstake finalizes unstaking after unbonding period
func (k Keeper) CompleteUnstake(ctx sdk.Context, address string) error {
	unbonding, found := k.GetUnbondingModerator(ctx, address)
	if !found {
		return types.ErrUnbondingNotFound
	}

	// Check unbonding period complete
	if !unbonding.IsUnbondingComplete(ctx.BlockTime()) {
		return types.ErrUnbondingNotComplete
	}

	// Check not slashed during unbonding
	if unbonding.Status == types.UnbondingStatusSlashed {
		k.DeleteUnbondingModerator(ctx, address)
		return types.ErrModeratorBlacklisted
	}

	mod, found := k.GetModerator(ctx, address)
	if !found {
		return types.ErrModeratorNotFound
	}

	// Transfer unbonded stake back to moderator
	modAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}
	unstakeCoins := sdk.NewCoins(sdk.NewCoin("hodl", unbonding.UnbondingAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, modAddr, unstakeCoins); err != nil {
		return err
	}

	// Update moderator stake
	mod.Stake = mod.Stake.Sub(unbonding.UnbondingAmount)
	mod.UnbondingStake = math.ZeroInt()

	// If stake drops below minimum, deactivate moderator
	minStake := math.NewInt(1000000000) // 1000 HODL
	if mod.Stake.LT(minStake) {
		mod.Active = false
	}

	// Update MaxEscrowValue to match new stake (trust ceiling = stake)
	mod.MaxEscrowValue = math.LegacyNewDecFromInt(mod.Stake)

	k.SetModerator(ctx, mod)
	k.DeleteUnbondingModerator(ctx, address)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorUnstakeCompleted,
			sdk.NewAttribute(types.AttributeKeyModerator, address),
			sdk.NewAttribute(types.AttributeKeyAmount, unbonding.UnbondingAmount.String()),
		),
	)

	return nil
}

// CancelUnstake cancels an in-progress unbonding request
func (k Keeper) CancelUnstake(ctx sdk.Context, address string) error {
	unbonding, found := k.GetUnbondingModerator(ctx, address)
	if !found {
		return types.ErrUnbondingNotFound
	}

	if unbonding.Status != types.UnbondingStatusActive {
		return types.ErrInvalidStatus
	}

	mod, found := k.GetModerator(ctx, address)
	if !found {
		return types.ErrModeratorNotFound
	}

	mod.UnbondingStake = math.ZeroInt()
	k.SetModerator(ctx, mod)
	k.DeleteUnbondingModerator(ctx, address)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorUnstakeCancelled,
			sdk.NewAttribute(types.AttributeKeyModerator, address),
		),
	)

	return nil
}

// IncreaseModeratorStake allows a moderator to increase their stake
func (k Keeper) IncreaseModeratorStake(ctx sdk.Context, address string, amount math.Int) error {
	mod, found := k.GetModerator(ctx, address)
	if !found {
		return types.ErrModeratorNotFound
	}

	if mod.Blacklisted {
		return types.ErrModeratorBlacklisted
	}

	// Transfer additional stake to module
	modAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}
	stakeCoins := sdk.NewCoins(sdk.NewCoin("hodl", amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, modAddr, types.ModuleName, stakeCoins); err != nil {
		return err
	}

	// Update moderator stake
	mod.Stake = mod.Stake.Add(amount)
	mod.MaxEscrowValue = math.LegacyNewDecFromInt(mod.Stake) // Update trust ceiling
	mod.Active = true                                        // Re-activate if was deactivated due to low stake

	k.SetModerator(ctx, mod)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorStakeIncreased,
			sdk.NewAttribute(types.AttributeKeyModerator, address),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyTrustCeiling, mod.MaxEscrowValue.String()),
		),
	)

	return nil
}

// ReleaseModeratorDisputeValue releases dispute value when an escrow is completed/cancelled
func (k Keeper) ReleaseModeratorDisputeValue(ctx sdk.Context, moderator string, value math.LegacyDec) error {
	mod, found := k.GetModerator(ctx, moderator)
	if !found {
		return types.ErrModeratorNotFound
	}

	// Decrease active dispute tracking
	mod.ActiveDisputeValue = mod.ActiveDisputeValue.Sub(value)
	if mod.ActiveDisputeValue.LT(math.LegacyZeroDec()) {
		mod.ActiveDisputeValue = math.LegacyZeroDec()
	}
	if mod.ActiveDisputeCount > 0 {
		mod.ActiveDisputeCount--
	}

	k.SetModerator(ctx, mod)
	return nil
}

// GetAllUnbondingModerators returns all moderators in unbonding
func (k Keeper) GetAllUnbondingModerators(ctx sdk.Context) []types.UnbondingModerator {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.UnbondingModeratorPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var unbondings []types.UnbondingModerator
	for ; iterator.Valid(); iterator.Next() {
		var unbonding types.UnbondingModerator
		if err := json.Unmarshal(iterator.Value(), &unbonding); err != nil {
			continue
		}
		unbondings = append(unbondings, unbonding)
	}

	return unbondings
}

// ============ Stake-as-Trust-Ceiling: Validator Oversight ============

// GetModeratorBlacklist returns a blacklist entry for a moderator
func (k Keeper) GetModeratorBlacklist(ctx sdk.Context, address string) (types.ModeratorBlacklist, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetModeratorBlacklistKey(address))
	if bz == nil {
		return types.ModeratorBlacklist{}, false
	}

	var blacklist types.ModeratorBlacklist
	if err := json.Unmarshal(bz, &blacklist); err != nil {
		return types.ModeratorBlacklist{}, false
	}
	return blacklist, true
}

// SetModeratorBlacklist stores a blacklist entry
func (k Keeper) SetModeratorBlacklist(ctx sdk.Context, blacklist types.ModeratorBlacklist) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(blacklist)
	if err != nil {
		return fmt.Errorf("failed to marshal blacklist: %w", err)
	}
	store.Set(types.GetModeratorBlacklistKey(blacklist.Address), bz)
	return nil
}

// DeleteModeratorBlacklist removes a blacklist entry
func (k Keeper) DeleteModeratorBlacklist(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetModeratorBlacklistKey(address))
}

// GetNextValidatorActionID returns the next action ID for audit logging
func (k Keeper) GetNextValidatorActionID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ValidatorActionCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.ValidatorActionCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// LogValidatorAction logs a validator oversight action for audit trail
func (k Keeper) LogValidatorAction(ctx sdk.Context, validator, action, target, reason string, amount math.Int) {
	actionID := k.GetNextValidatorActionID(ctx)
	actionLog := types.ValidatorAction{
		ID:        actionID,
		Validator: validator,
		Action:    action,
		Target:    target,
		Reason:    reason,
		Amount:    amount,
		Timestamp: ctx.BlockTime(),
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(actionLog)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal validator action", "error", err)
		return
	}
	store.Set(types.GetValidatorActionLogKey(actionID), bz)

	// Emit event for audit
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeValidatorAction,
			sdk.NewAttribute(types.AttributeKeyActionID, fmt.Sprintf("%d", actionID)),
			sdk.NewAttribute(types.AttributeKeyValidator, validator),
			sdk.NewAttribute(types.AttributeKeyAction, action),
			sdk.NewAttribute(types.AttributeKeyTarget, target),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)
}

// SlashModerator allows validators to slash a moderator's stake
// Warden+ tier (tier >= 2) can slash moderators
func (k Keeper) SlashModerator(
	ctx sdk.Context,
	validatorAddr string,
	moderatorAddr string,
	slashFraction math.LegacyDec,
	reason string,
) error {
	// Verify caller has validator oversight capability via staking tier
	if k.stakingKeeper == nil {
		return types.ErrUnauthorizedValidatorAction
	}

	valAddr, err := sdk.AccAddressFromBech32(validatorAddr)
	if err != nil {
		return err
	}

	// Check tier - Warden+ (tier >= 2) can slash moderators
	tier := k.stakingKeeper.GetUserTierInt(ctx, valAddr)
	if tier < types.StakeTierWarden {
		return types.ErrValidatorTierTooLow
	}

	// Validate slash fraction
	if slashFraction.LT(math.LegacyZeroDec()) || slashFraction.GT(math.LegacyOneDec()) {
		return types.ErrInvalidSlashFraction
	}

	mod, found := k.GetModerator(ctx, moderatorAddr)
	if !found {
		return types.ErrModeratorNotFound
	}

	// Calculate slash amount (from both staked and unbonding)
	totalSlashable := mod.Stake
	if !mod.UnbondingStake.IsNil() {
		totalSlashable = totalSlashable.Add(mod.UnbondingStake)
	}
	slashAmount := slashFraction.MulInt(totalSlashable).TruncateInt()

	if slashAmount.IsZero() {
		return nil // Nothing to slash
	}

	// Burn slashed tokens FIRST - CRITICAL: Must succeed before modifying state
	slashCoins := sdk.NewCoins(sdk.NewCoin("hodl", slashAmount))
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, slashCoins); err != nil {
		// BurnCoins failure is critical - abort slashing to maintain consistency
		// The moderator's stake must match the module account balance
		k.Logger(ctx).Error("failed to burn slashed tokens, aborting slash operation", "error", err)
		return fmt.Errorf("failed to burn slashed tokens: %w", err)
	}

	// Update moderator state ONLY after successful burn
	// Slash from stake first, then from unbonding
	remainingSlash := slashAmount
	if remainingSlash.LTE(mod.Stake) {
		mod.Stake = mod.Stake.Sub(remainingSlash)
		remainingSlash = math.ZeroInt()
	} else {
		remainingSlash = remainingSlash.Sub(mod.Stake)
		mod.Stake = math.ZeroInt()
	}

	// Slash from unbonding if needed
	if remainingSlash.IsPositive() && !mod.UnbondingStake.IsNil() && mod.UnbondingStake.IsPositive() {
		if remainingSlash.LTE(mod.UnbondingStake) {
			mod.UnbondingStake = mod.UnbondingStake.Sub(remainingSlash)
		} else {
			mod.UnbondingStake = math.ZeroInt()
		}
	}

	// Update unbonding record if in progress
	if unbonding, found := k.GetUnbondingModerator(ctx, moderatorAddr); found {
		unbonding.UnbondingAmount = mod.UnbondingStake
		if unbonding.UnbondingAmount.IsZero() {
			unbonding.Status = types.UnbondingStatusSlashed
		}
		k.SetUnbondingModerator(ctx, unbonding)
	}

	// HIGH PRIORITY FIX: Check if moderator has active disputes exceeding new trust ceiling
	// This enforces the trust ceiling retroactively after slashing
	newTrustCeiling := math.LegacyNewDecFromInt(mod.Stake)
	if mod.ActiveDisputeValue.GT(newTrustCeiling) {
		// Log warning - moderator now over-committed
		k.Logger(ctx).Warn("moderator over-committed after slash",
			"moderator", moderatorAddr,
			"active_disputes", mod.ActiveDisputeValue.String(),
			"new_ceiling", newTrustCeiling.String(),
		)
		// Force deactivate to prevent new dispute assignments
		mod.Active = false
	}

	// Update trust ceiling
	mod.MaxEscrowValue = math.LegacyNewDecFromInt(mod.Stake)

	// Deactivate if below minimum stake
	minStake := math.NewInt(1000000000) // 1000 HODL
	if mod.Stake.LT(minStake) {
		mod.Active = false
	}

	k.SetModerator(ctx, mod)

	// Log action for audit trail
	k.LogValidatorAction(ctx, validatorAddr, "slash", moderatorAddr, reason, slashAmount)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorSlashed,
			sdk.NewAttribute(types.AttributeKeyModerator, moderatorAddr),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
			sdk.NewAttribute(types.AttributeKeySlashAmount, slashAmount.String()),
			sdk.NewAttribute(types.AttributeKeySlashFraction, slashFraction.String()),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)

	return nil
}

// BlacklistModerator bans a moderator from participating
// Steward+ tier (tier >= 3) can blacklist moderators
func (k Keeper) BlacklistModerator(
	ctx sdk.Context,
	validatorAddr string,
	moderatorAddr string,
	reason string,
	permanent bool,
	banDuration time.Duration,
) error {
	// Verify caller has validator oversight capability via staking tier
	if k.stakingKeeper == nil {
		return types.ErrUnauthorizedValidatorAction
	}

	valAddr, err := sdk.AccAddressFromBech32(validatorAddr)
	if err != nil {
		return err
	}

	// Check tier - Steward+ (tier >= 3) can blacklist moderators
	tier := k.stakingKeeper.GetUserTierInt(ctx, valAddr)
	if tier < types.StakeTierSteward {
		return types.ErrValidatorTierTooLow
	}

	mod, found := k.GetModerator(ctx, moderatorAddr)
	if !found {
		return types.ErrModeratorNotFound
	}
	if mod.Blacklisted {
		return types.ErrAlreadyBlacklisted
	}

	// Create blacklist entry
	blacklist := types.NewModeratorBlacklist(moderatorAddr, validatorAddr, reason, permanent, banDuration, ctx.BlockTime())
	if err := k.SetModeratorBlacklist(ctx, blacklist); err != nil {
		return err
	}

	// Update moderator
	mod.Active = false
	mod.Blacklisted = true
	mod.BlacklistedAt = ctx.BlockTime()
	k.SetModerator(ctx, mod)

	// Log action for audit trail
	k.LogValidatorAction(ctx, validatorAddr, "blacklist", moderatorAddr, reason, math.ZeroInt())

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorBlacklisted,
			sdk.NewAttribute(types.AttributeKeyModerator, moderatorAddr),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
			sdk.NewAttribute(types.AttributeKeyPermanent, fmt.Sprintf("%t", permanent)),
		),
	)

	return nil
}

// UnblacklistModerator removes a moderator from the blacklist
// Only works for temporary bans that have expired, or requires Archon+ tier
func (k Keeper) UnblacklistModerator(
	ctx sdk.Context,
	validatorAddr string,
	moderatorAddr string,
) error {
	// Verify caller has validator oversight capability via staking tier
	if k.stakingKeeper == nil {
		return types.ErrUnauthorizedValidatorAction
	}

	valAddr, err := sdk.AccAddressFromBech32(validatorAddr)
	if err != nil {
		return err
	}

	mod, found := k.GetModerator(ctx, moderatorAddr)
	if !found {
		return types.ErrModeratorNotFound
	}
	if !mod.Blacklisted {
		return types.ErrModeratorNotBlacklisted
	}

	blacklist, found := k.GetModeratorBlacklist(ctx, moderatorAddr)
	if !found {
		return types.ErrModeratorNotBlacklisted
	}

	// Check if ban can be lifted
	tier := k.stakingKeeper.GetUserTierInt(ctx, valAddr)
	if blacklist.Permanent {
		// Permanent bans can only be lifted by Archon+ (tier >= 4)
		if tier < types.StakeTierArchon {
			return types.ErrValidatorTierTooLow
		}
	} else {
		// Temporary bans can be lifted by Steward+ after expiry, or Archon+ anytime
		if tier < types.StakeTierArchon && !blacklist.IsBanExpired(ctx.BlockTime()) {
			return types.ErrBanNotExpired
		}
	}

	// Remove blacklist
	k.DeleteModeratorBlacklist(ctx, moderatorAddr)

	// Update moderator
	mod.Blacklisted = false
	mod.BlacklistedAt = time.Time{}

	// Re-activate if has sufficient stake
	minStake := math.NewInt(1000000000) // 1000 HODL
	if mod.Stake.GTE(minStake) {
		mod.Active = true
	}

	k.SetModerator(ctx, mod)

	// Log action for audit trail
	k.LogValidatorAction(ctx, validatorAddr, "unblacklist", moderatorAddr, "ban lifted", math.ZeroInt())

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorUnblacklisted,
			sdk.NewAttribute(types.AttributeKeyModerator, moderatorAddr),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
		),
	)

	return nil
}

// GetAllBlacklistedModerators returns all blacklisted moderators
func (k Keeper) GetAllBlacklistedModerators(ctx sdk.Context) []types.ModeratorBlacklist {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.ModeratorBlacklistPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var blacklisted []types.ModeratorBlacklist
	for ; iterator.Valid(); iterator.Next() {
		var bl types.ModeratorBlacklist
		if err := json.Unmarshal(iterator.Value(), &bl); err != nil {
			continue
		}
		blacklisted = append(blacklisted, bl)
	}

	return blacklisted
}

// ============ EndBlocker Processing Functions ============

// ProcessCompletedUnbondings processes all unbondings that have completed the 14-day period
// This is called from EndBlock to automatically complete unbondings
func (k Keeper) ProcessCompletedUnbondings(ctx sdk.Context) {
	unbondings := k.GetAllUnbondingModerators(ctx)
	currentTime := ctx.BlockTime()

	for _, unbonding := range unbondings {
		// Skip if not active or already processed
		if unbonding.Status != types.UnbondingStatusActive {
			continue
		}

		// Check if unbonding period is complete
		if unbonding.IsUnbondingComplete(currentTime) {
			// Auto-complete the unbonding
			if err := k.CompleteUnstake(ctx, unbonding.Address); err != nil {
				k.Logger(ctx).Error("failed to auto-complete unbonding",
					"moderator", unbonding.Address,
					"error", err,
				)
				continue
			}

			k.Logger(ctx).Info("auto-completed moderator unbonding",
				"moderator", unbonding.Address,
				"amount", unbonding.UnbondingAmount.String(),
			)
		}
	}
}

// ProcessExpiredBans processes all temporary bans that have expired
// This is called from EndBlock to automatically lift expired bans
func (k Keeper) ProcessExpiredBans(ctx sdk.Context) {
	blacklists := k.GetAllBlacklistedModerators(ctx)
	currentTime := ctx.BlockTime()

	for _, blacklist := range blacklists {
		// Skip permanent bans
		if blacklist.Permanent {
			continue
		}

		// Check if ban has expired
		if blacklist.IsBanExpired(currentTime) {
			mod, found := k.GetModerator(ctx, blacklist.Address)
			if !found {
				continue
			}

			// Remove from blacklist
			k.DeleteModeratorBlacklist(ctx, blacklist.Address)

			// Update moderator
			mod.Blacklisted = false
			// Re-activate if has sufficient stake
			minStake := math.NewInt(1000000000) // 1000 HODL
			if mod.Stake.GTE(minStake) {
				mod.Active = true
			}
			k.SetModerator(ctx, mod)

			// Log action
			k.LogValidatorAction(ctx, "system", "auto_unblacklist", blacklist.Address, "temporary ban expired", math.ZeroInt())

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeModeratorUnblacklisted,
					sdk.NewAttribute(types.AttributeKeyModerator, blacklist.Address),
					sdk.NewAttribute(types.AttributeKeyReason, "temporary ban expired"),
				),
			)

			k.Logger(ctx).Info("auto-lifted expired moderator ban",
				"moderator", blacklist.Address,
				"original_reason", blacklist.Reason,
			)
		}
	}
}

// ============ Validator Action Audit Trail ============

// SetValidatorAction stores a validator action for audit trail
func (k Keeper) SetValidatorAction(ctx sdk.Context, action types.ValidatorAction) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(action)
	if err != nil {
		return fmt.Errorf("failed to marshal validator action: %w", err)
	}
	store.Set(types.GetValidatorActionLogKey(action.ID), bz)
	return nil
}

// GetValidatorAction returns a validator action by ID
func (k Keeper) GetValidatorAction(ctx sdk.Context, id uint64) (types.ValidatorAction, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetValidatorActionLogKey(id))
	if bz == nil {
		return types.ValidatorAction{}, false
	}

	var action types.ValidatorAction
	if err := json.Unmarshal(bz, &action); err != nil {
		return types.ValidatorAction{}, false
	}
	return action, true
}

// GetAllValidatorActions returns all validator actions for export
func (k Keeper) GetAllValidatorActions(ctx sdk.Context) []types.ValidatorAction {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.ValidatorActionLogPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var actions []types.ValidatorAction
	for ; iterator.Valid(); iterator.Next() {
		var action types.ValidatorAction
		if err := json.Unmarshal(iterator.Value(), &action); err != nil {
			continue
		}
		actions = append(actions, action)
	}

	return actions
}

// ============ Genesis Export Helpers ============

// GetAllEscrows returns all escrows for genesis export
func (k Keeper) GetAllEscrows(ctx sdk.Context) []types.Escrow {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.EscrowPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var escrows []types.Escrow
	for ; iterator.Valid(); iterator.Next() {
		var escrow types.Escrow
		if err := json.Unmarshal(iterator.Value(), &escrow); err != nil {
			continue
		}
		escrows = append(escrows, escrow)
	}

	return escrows
}

// GetAllDisputes returns all disputes for genesis export
func (k Keeper) GetAllDisputes(ctx sdk.Context) []types.Dispute {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.DisputePrefix).Iterator(nil, nil)
	defer iterator.Close()

	var disputes []types.Dispute
	for ; iterator.Valid(); iterator.Next() {
		var dispute types.Dispute
		if err := json.Unmarshal(iterator.Value(), &dispute); err != nil {
			continue
		}
		disputes = append(disputes, dispute)
	}

	return disputes
}
