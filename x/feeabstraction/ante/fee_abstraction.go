package ante

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/types"
)

// FeeAbstractionDecorator is an ante handler decorator that processes fee abstraction
// It intercepts transactions where the fee payer lacks sufficient HODL and attempts
// to pay fees using equity tokens instead
type FeeAbstractionDecorator struct {
	ak           types.AccountKeeper
	bk           types.BankKeeper
	feeabsKeeper *keeper.Keeper
}

// NewFeeAbstractionDecorator creates a new fee abstraction decorator
func NewFeeAbstractionDecorator(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	feeabsKeeper *keeper.Keeper,
) FeeAbstractionDecorator {
	return FeeAbstractionDecorator{
		ak:           ak,
		bk:           bk,
		feeabsKeeper: feeabsKeeper,
	}
}

// AnteHandle processes the transaction with fee abstraction
func (fad FeeAbstractionDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errors.Wrap(sdkerrors.ErrTxDecode, "tx must be FeeTx")
	}

	feePayerBytes := feeTx.FeePayer()
	feePayer := sdk.AccAddress(feePayerBytes)
	feeCoins := feeTx.GetFee()

	// Get required HODL amount
	requiredHODL := feeCoins.AmountOf("uhodl")
	if requiredHODL.IsZero() {
		// No HODL fee required, proceed normally
		return next(ctx, tx, simulate)
	}

	// Check if user has sufficient HODL
	hodlBalance := fad.bk.GetBalance(ctx, feePayer, "uhodl")

	if hodlBalance.Amount.GTE(requiredHODL) {
		// User has HODL - proceed normally
		return next(ctx, tx, simulate)
	}

	// User lacks HODL - attempt fee abstraction
	shortfall := requiredHODL.Sub(hodlBalance.Amount)

	// Process fee abstraction
	_, err := fad.feeabsKeeper.ProcessFeeAbstraction(ctx, feePayer, shortfall)
	if err != nil {
		// Fee abstraction failed - return detailed error for UX
		return ctx, errors.Wrapf(err,
			"insufficient HODL (%s) and fee abstraction failed",
			hodlBalance.Amount)
	}

	// Fee abstraction succeeded - mark context
	ctx = ctx.WithValue("fee_abstraction_used", true)
	ctx = ctx.WithValue("fee_abstraction_amount", shortfall)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"fee_abstraction_ante",
			sdk.NewAttribute("fee_payer", feePayer.String()),
			sdk.NewAttribute("shortfall", shortfall.String()),
		),
	)

	return next(ctx, tx, simulate)
}

// FeeAbstractionContextKey is the context key for fee abstraction status
type FeeAbstractionContextKey string

const (
	// ContextKeyFeeAbstractionUsed indicates if fee abstraction was used
	ContextKeyFeeAbstractionUsed FeeAbstractionContextKey = "fee_abstraction_used"
	
	// ContextKeyFeeAbstractionAmount is the amount that was abstracted
	ContextKeyFeeAbstractionAmount FeeAbstractionContextKey = "fee_abstraction_amount"
)

// WasFeeAbstractionUsed checks if fee abstraction was used in this transaction
func WasFeeAbstractionUsed(ctx sdk.Context) bool {
	val := ctx.Value(ContextKeyFeeAbstractionUsed)
	if val == nil {
		return false
	}
	used, ok := val.(bool)
	return ok && used
}

// GetFeeAbstractionAmount returns the amount that was fee abstracted
func GetFeeAbstractionAmount(ctx sdk.Context) math.Int {
	val := ctx.Value(ContextKeyFeeAbstractionAmount)
	if val == nil {
		return math.ZeroInt()
	}
	amount, ok := val.(math.Int)
	if !ok {
		return math.ZeroInt()
	}
	return amount
}
