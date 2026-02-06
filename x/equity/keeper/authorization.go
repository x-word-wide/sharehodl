package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// =============================================================================
// Company Authorization CRUD Operations
// =============================================================================

// SetCompanyAuthorization stores a company authorization record
func (k Keeper) SetCompanyAuthorization(ctx sdk.Context, auth types.CompanyAuthorization) error {
	// Validate before saving
	if err := auth.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyAuthorizationKey(auth.CompanyID)
	bz, err := json.Marshal(auth)
	if err != nil {
		return fmt.Errorf("failed to marshal authorization: %w", err)
	}
	store.Set(key, bz)

	k.Logger(ctx).Info("company authorization saved",
		"company_id", auth.CompanyID,
		"owner", auth.Owner,
		"proposer_count", len(auth.AuthorizedProposers),
	)

	return nil
}

// GetCompanyAuthorization retrieves a company authorization by company ID
func (k Keeper) GetCompanyAuthorization(ctx sdk.Context, companyID uint64) (types.CompanyAuthorization, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyAuthorizationKey(companyID)
	bz := store.Get(key)
	if bz == nil {
		return types.CompanyAuthorization{}, false
	}

	var auth types.CompanyAuthorization
	if err := json.Unmarshal(bz, &auth); err != nil {
		k.Logger(ctx).Error("failed to unmarshal authorization", "company_id", companyID, "error", err)
		return types.CompanyAuthorization{}, false
	}
	return auth, true
}

// DeleteCompanyAuthorization removes a company authorization record
func (k Keeper) DeleteCompanyAuthorization(ctx sdk.Context, companyID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyAuthorizationKey(companyID)
	store.Delete(key)
}

// =============================================================================
// Authorization Check Functions
// =============================================================================

// IsCompanyOwner checks if the given address is the company owner
func (k Keeper) IsCompanyOwner(ctx sdk.Context, companyID uint64, address string) bool {
	auth, found := k.GetCompanyAuthorization(ctx, companyID)
	if !found {
		// Fallback to checking if founder matches (for backwards compatibility)
		company, found := k.getCompany(ctx, companyID)
		if !found {
			return false
		}
		return company.Founder == address
	}
	return auth.IsOwner(address)
}

// CanProposeForCompany checks if address can propose for company (owner OR proposer)
func (k Keeper) CanProposeForCompany(ctx sdk.Context, companyID uint64, address string) bool {
	auth, found := k.GetCompanyAuthorization(ctx, companyID)
	if !found {
		// Fallback to checking if founder matches (for backwards compatibility)
		company, found := k.getCompany(ctx, companyID)
		if !found {
			return false
		}
		return company.Founder == address
	}
	return auth.CanPropose(address)
}

// =============================================================================
// Authorization Management Functions
// =============================================================================

// AddAuthorizedProposer adds a new authorized proposer to a company
func (k Keeper) AddAuthorizedProposer(ctx sdk.Context, companyID uint64, owner string, newProposer string) error {
	// Get authorization record
	auth, found := k.GetCompanyAuthorization(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Verify owner
	if !auth.IsOwner(owner) {
		return types.ErrNotCompanyOwner
	}

	// Add proposer
	if err := auth.AddProposer(newProposer); err != nil {
		return err
	}

	// Save updated authorization
	if err := k.SetCompanyAuthorization(ctx, auth); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerAdded,
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyProposer, newProposer),
			sdk.NewAttribute(types.AttributeKeyProposerCount, fmt.Sprintf("%d", len(auth.AuthorizedProposers))),
		),
	)

	k.Logger(ctx).Info("authorized proposer added",
		"company_id", companyID,
		"proposer", newProposer,
		"total_proposers", len(auth.AuthorizedProposers),
	)

	return nil
}

// RemoveAuthorizedProposer removes an authorized proposer from a company
func (k Keeper) RemoveAuthorizedProposer(ctx sdk.Context, companyID uint64, owner string, proposerToRemove string) error {
	// Get authorization record
	auth, found := k.GetCompanyAuthorization(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Verify owner
	if !auth.IsOwner(owner) {
		return types.ErrNotCompanyOwner
	}

	// Cannot remove owner (they implicitly have proposer privileges)
	if auth.IsOwner(proposerToRemove) {
		return types.ErrCannotRemoveSelf
	}

	// Remove proposer
	if err := auth.RemoveProposer(proposerToRemove); err != nil {
		return err
	}

	// Save updated authorization
	if err := k.SetCompanyAuthorization(ctx, auth); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposerRemoved,
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyProposer, proposerToRemove),
			sdk.NewAttribute(types.AttributeKeyProposerCount, fmt.Sprintf("%d", len(auth.AuthorizedProposers))),
		),
	)

	k.Logger(ctx).Info("authorized proposer removed",
		"company_id", companyID,
		"proposer", proposerToRemove,
		"remaining_proposers", len(auth.AuthorizedProposers),
	)

	return nil
}

// InitiateOwnershipTransfer starts a two-step ownership transfer
func (k Keeper) InitiateOwnershipTransfer(ctx sdk.Context, companyID uint64, currentOwner string, newOwner string) error {
	// Get authorization record
	auth, found := k.GetCompanyAuthorization(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Verify current owner
	if !auth.IsOwner(currentOwner) {
		return types.ErrNotCompanyOwner
	}

	// Initiate transfer
	if err := auth.InitiateOwnershipTransfer(newOwner); err != nil {
		return err
	}

	// Save updated authorization
	if err := k.SetCompanyAuthorization(ctx, auth); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOwnershipTransferInitiated,
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyOldOwner, currentOwner),
			sdk.NewAttribute(types.AttributeKeyPendingOwner, newOwner),
		),
	)

	k.Logger(ctx).Info("ownership transfer initiated",
		"company_id", companyID,
		"current_owner", currentOwner,
		"pending_owner", newOwner,
	)

	return nil
}

// AcceptOwnershipTransfer completes the ownership transfer (called by new owner)
func (k Keeper) AcceptOwnershipTransfer(ctx sdk.Context, companyID uint64, newOwner string) error {
	// Get authorization record
	auth, found := k.GetCompanyAuthorization(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Store old owner for event
	oldOwner := auth.Owner

	// Accept transfer
	if err := auth.AcceptOwnershipTransfer(newOwner); err != nil {
		return err
	}

	// Save updated authorization
	if err := k.SetCompanyAuthorization(ctx, auth); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOwnershipTransferAccepted,
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyOldOwner, oldOwner),
			sdk.NewAttribute(types.AttributeKeyNewOwner, newOwner),
		),
	)

	k.Logger(ctx).Info("ownership transfer accepted",
		"company_id", companyID,
		"old_owner", oldOwner,
		"new_owner", newOwner,
	)

	return nil
}

// CancelOwnershipTransfer cancels a pending ownership transfer
func (k Keeper) CancelOwnershipTransfer(ctx sdk.Context, companyID uint64, currentOwner string) error {
	// Get authorization record
	auth, found := k.GetCompanyAuthorization(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Verify current owner
	if !auth.IsOwner(currentOwner) {
		return types.ErrNotCompanyOwner
	}

	// Store pending owner for event
	pendingOwner := auth.PendingOwnerTransfer

	// Cancel transfer
	if err := auth.CancelOwnershipTransfer(); err != nil {
		return err
	}

	// Save updated authorization
	if err := k.SetCompanyAuthorization(ctx, auth); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOwnershipTransferCancelled,
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyOldOwner, currentOwner),
			sdk.NewAttribute(types.AttributeKeyPendingOwner, pendingOwner),
		),
	)

	k.Logger(ctx).Info("ownership transfer cancelled",
		"company_id", companyID,
		"owner", currentOwner,
		"cancelled_pending_owner", pendingOwner,
	)

	return nil
}
