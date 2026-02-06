package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/escrow/types"
)

// TransferBuyerPosition transfers escrow buyer position for inheritance
// The beneficiary becomes the buyer and expects delivery
func (k Keeper) TransferBuyerPosition(ctx sdk.Context, escrowID uint64, from, to string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound
	}

	// Verify the from address is the current sender (buyer)
	if escrow.Sender != from {
		return fmt.Errorf("address %s is not the sender of escrow %d", from, escrowID)
	}

	// SECURITY FIX: Block transfer during disputed state to prevent abuse
	if escrow.Status == types.EscrowStatusDisputed {
		return fmt.Errorf("cannot transfer escrow position during active dispute")
	}

	// Verify the escrow is in a transferable state
	if escrow.Status != types.EscrowStatusFunded {
		return fmt.Errorf("escrow %d cannot be transferred (status: %s)", escrowID, escrow.Status.String())
	}

	// Validate the new buyer address
	_, err := sdk.AccAddressFromBech32(to)
	if err != nil {
		return fmt.Errorf("invalid beneficiary address: %w", err)
	}

	// Update the escrow's sender (buyer)
	escrow.Sender = to

	// Reset confirmations since we have a new party
	escrow.SenderConfirmed = false

	// NOTE: Beneficial ownership for equity in escrow needs to be updated
	// This will be handled by the inheritance keeper calling equityKeeper methods
	// For now, we just update the escrow record

	// Save the updated escrow
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"escrow_buyer_transferred",
			sdk.NewAttribute(types.AttributeKeyEscrowID, fmt.Sprintf("%d", escrowID)),
			sdk.NewAttribute("from", from),
			sdk.NewAttribute("to", to),
			sdk.NewAttribute("value", escrow.TotalValue.String()),
			sdk.NewAttribute("status", escrow.Status.String()),
		),
	)

	k.Logger(ctx).Info("escrow buyer position transferred",
		"escrow_id", escrowID,
		"from", from,
		"to", to,
		"value", escrow.TotalValue.String(),
		"note", "beneficiary expects delivery and can release/dispute",
	)

	return nil
}

// TransferSellerPosition transfers escrow seller position for inheritance
// The beneficiary becomes the seller and must deliver goods/shares
func (k Keeper) TransferSellerPosition(ctx sdk.Context, escrowID uint64, from, to string) error {
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound
	}

	// Verify the from address is the current recipient (seller)
	if escrow.Recipient != from {
		return fmt.Errorf("address %s is not the recipient of escrow %d", from, escrowID)
	}

	// SECURITY FIX: Block transfer during disputed state to prevent abuse
	if escrow.Status == types.EscrowStatusDisputed {
		return fmt.Errorf("cannot transfer escrow position during active dispute")
	}

	// Verify the escrow is in a transferable state
	if escrow.Status != types.EscrowStatusFunded {
		return fmt.Errorf("escrow %d cannot be transferred (status: %s)", escrowID, escrow.Status.String())
	}

	// Validate the new seller address
	_, err := sdk.AccAddressFromBech32(to)
	if err != nil {
		return fmt.Errorf("invalid beneficiary address: %w", err)
	}

	// Update the escrow's recipient (seller)
	escrow.Recipient = to

	// Reset confirmations since we have a new party
	escrow.RecipientConfirmed = false

	// Save the updated escrow
	if err := k.SetEscrow(ctx, escrow); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"escrow_seller_transferred",
			sdk.NewAttribute(types.AttributeKeyEscrowID, fmt.Sprintf("%d", escrowID)),
			sdk.NewAttribute("from", from),
			sdk.NewAttribute("to", to),
			sdk.NewAttribute("value", escrow.TotalValue.String()),
			sdk.NewAttribute("status", escrow.Status.String()),
		),
	)

	k.Logger(ctx).Info("escrow seller position transferred",
		"escrow_id", escrowID,
		"from", from,
		"to", to,
		"value", escrow.TotalValue.String(),
		"note", "beneficiary must deliver and receives payment on completion",
	)

	return nil
}
