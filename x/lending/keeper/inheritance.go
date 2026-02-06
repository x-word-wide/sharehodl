package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/lending/types"
)

// TransferBorrowerPosition transfers a loan's borrower position to a new address
// This is called during inheritance to transfer loan obligations to beneficiaries
// The beneficiary inherits both the collateral AND the debt
func (k Keeper) TransferBorrowerPosition(ctx sdk.Context, loanID uint64, from, to string) error {
	loan, found := k.GetLoan(ctx, loanID)
	if !found {
		return types.ErrLoanNotFound
	}

	// Verify the from address is the current borrower
	if loan.Borrower != from {
		return fmt.Errorf("address %s is not the borrower of loan %d", from, loanID)
	}

	// Verify the loan is in an active or pending state
	if loan.Status != types.LoanStatusActive && loan.Status != types.LoanStatusPending {
		return fmt.Errorf("loan %d is not active or pending (status: %s)", loanID, loan.Status.String())
	}

	// Validate the new borrower address
	_, err := sdk.AccAddressFromBech32(to)
	if err != nil {
		return fmt.Errorf("invalid beneficiary address: %w", err)
	}

	// Update the loan's borrower
	loan.Borrower = to

	// NOTE: We do NOT transfer collateral here - it's already locked in the lending module
	// The beneficial ownership will be updated by the inheritance keeper if needed
	// The new borrower is now responsible for repaying the loan

	// Save the updated loan
	if err := k.SetLoan(ctx, loan); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"loan_borrower_transferred",
			sdk.NewAttribute(types.AttributeKeyLoanID, fmt.Sprintf("%d", loanID)),
			sdk.NewAttribute("from", from),
			sdk.NewAttribute("to", to),
			sdk.NewAttribute("principal", loan.Principal.String()),
			sdk.NewAttribute("collateral_value", loan.CollateralValue.String()),
			sdk.NewAttribute("total_owed", loan.TotalOwed.String()),
		),
	)

	k.Logger(ctx).Info("loan borrower position transferred",
		"loan_id", loanID,
		"from", from,
		"to", to,
		"principal", loan.Principal.String(),
		"collateral", loan.CollateralValue.String(),
		"note", "beneficiary now owns collateral AND owes the debt",
	)

	return nil
}

// TransferLenderPosition transfers a lending position to a new address
// This is called during inheritance to transfer lending positions to beneficiaries
// The beneficiary will receive future loan repayments
func (k Keeper) TransferLenderPosition(ctx sdk.Context, loanID uint64, from, to string) error {
	loan, found := k.GetLoan(ctx, loanID)
	if !found {
		return types.ErrLoanNotFound
	}

	// Verify the from address is the current lender
	if loan.Lender != from {
		return fmt.Errorf("address %s is not the lender of loan %d", from, loanID)
	}

	// Verify the loan is in an active state (not pending, as lender is set on activation)
	if loan.Status != types.LoanStatusActive {
		return fmt.Errorf("loan %d is not active (status: %s)", loanID, loan.Status.String())
	}

	// Validate the new lender address
	_, err := sdk.AccAddressFromBech32(to)
	if err != nil {
		return fmt.Errorf("invalid beneficiary address: %w", err)
	}

	// Update the loan's lender
	loan.Lender = to

	// Save the updated loan
	if err := k.SetLoan(ctx, loan); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"loan_lender_transferred",
			sdk.NewAttribute(types.AttributeKeyLoanID, fmt.Sprintf("%d", loanID)),
			sdk.NewAttribute("from", from),
			sdk.NewAttribute("to", to),
			sdk.NewAttribute("principal", loan.Principal.String()),
			sdk.NewAttribute("total_owed", loan.TotalOwed.String()),
		),
	)

	k.Logger(ctx).Info("loan lender position transferred",
		"loan_id", loanID,
		"from", from,
		"to", to,
		"principal", loan.Principal.String(),
		"note", "beneficiary will receive future loan repayments",
	)

	return nil
}
