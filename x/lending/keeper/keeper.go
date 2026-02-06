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

	"github.com/sharehodl/sharehodl-blockchain/x/lending/types"
)

// Keeper of the lending store
type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      storetypes.StoreKey
	memKey        storetypes.StoreKey
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
	equityKeeper  types.EquityKeeper
	stakingKeeper types.UniversalStakingKeeper // For tier/reputation checks and validator oversight
}

// NewKeeper creates a new lending Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	equityKeeper types.EquityKeeper,
	stakingKeeper types.UniversalStakingKeeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		equityKeeper:  equityKeeper,
		stakingKeeper: stakingKeeper,
	}
}

// SetStakingKeeper sets the staking keeper (for late binding during app initialization)
func (k *Keeper) SetStakingKeeper(stakingKeeper types.UniversalStakingKeeper) {
	k.stakingKeeper = stakingKeeper
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// =============================================================================
// PARAMS MANAGEMENT - All values governance-controllable
// =============================================================================

// GetParams returns the current lending parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

// SetParams sets the lending parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}

// Param getters for convenience

// GetLoanDefaultSlashThreshold returns the governance-controllable slash threshold
func (k Keeper) GetLoanDefaultSlashThreshold(ctx sdk.Context) math.Int {
	return k.GetParams(ctx).LoanDefaultSlashThreshold
}

// GetDefaultLoanDuration returns the governance-controllable default loan duration
func (k Keeper) GetDefaultLoanDuration(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).GetDefaultLoanDuration()
}

// GetLenderUnbondingPeriod returns the governance-controllable lender unbonding period
func (k Keeper) GetLenderUnbondingPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).GetLenderUnbondingPeriod()
}

// GetBorrowerUnbondingPeriod returns the governance-controllable borrower unbonding period
func (k Keeper) GetBorrowerUnbondingPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).GetBorrowerUnbondingPeriod()
}

// IsLendingEnabled returns whether lending is enabled
func (k Keeper) IsLendingEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).LendingEnabled
}

// =============================================================================
// TIER CHECKS - Require Keeper+ tier for lending participation
// =============================================================================

// checkCanLend verifies a user has the required tier and reputation to lend
func (k Keeper) checkCanLend(ctx sdk.Context, address string) error {
	if k.stakingKeeper == nil {
		// If staking keeper not set, allow (for backwards compatibility during init)
		return nil
	}

	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	if !k.stakingKeeper.CanLend(ctx, addr) {
		return types.ErrInsufficientTierToLend
	}

	return nil
}

// checkCanBorrow verifies a user has the required tier and reputation to borrow
func (k Keeper) checkCanBorrow(ctx sdk.Context, address string) error {
	if k.stakingKeeper == nil {
		// If staking keeper not set, allow (for backwards compatibility during init)
		return nil
	}

	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	if !k.stakingKeeper.CanBorrow(ctx, addr) {
		return types.ErrInsufficientTierToBorrow
	}

	return nil
}

// notifyLoanSuccess notifies the staking system of successful loan repayment
func (k Keeper) notifyLoanSuccess(ctx sdk.Context, borrowerAddr string, loanID uint64) {
	if k.stakingKeeper == nil {
		return
	}

	addr, err := sdk.AccAddressFromBech32(borrowerAddr)
	if err != nil {
		return
	}

	if err := k.stakingKeeper.RewardSuccessfulLoan(ctx, addr, fmt.Sprintf("%d", loanID)); err != nil {
		k.Logger(ctx).Error("failed to reward successful loan", "error", err)
	}
}

// notifyLoanDefault notifies the staking system of loan default
func (k Keeper) notifyLoanDefault(ctx sdk.Context, borrowerAddr string, loanID uint64, defaultAmount math.Int) {
	if k.stakingKeeper == nil {
		return
	}

	addr, err := sdk.AccAddressFromBech32(borrowerAddr)
	if err != nil {
		return
	}

	// Penalize reputation
	if err := k.stakingKeeper.PenalizeLoanDefault(ctx, addr, fmt.Sprintf("%d", loanID)); err != nil {
		k.Logger(ctx).Error("failed to penalize loan default", "error", err)
	}

	// Slash stake if significant default (governance-controllable threshold)
	slashThreshold := k.GetLoanDefaultSlashThreshold(ctx)
	if defaultAmount.GT(slashThreshold) {
		if err := k.stakingKeeper.SlashForLoanDefault(ctx, addr, defaultAmount); err != nil {
			k.Logger(ctx).Error("failed to slash for loan default", "error", err)
		}
	}
}

// GetNextLoanID returns the next loan ID and increments the counter
func (k Keeper) GetNextLoanID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.LoanCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.LoanCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetNextPoolID returns the next pool ID and increments the counter
func (k Keeper) GetNextPoolID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PoolCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.PoolCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// ============ Loan Management ============

// GetLoan returns a loan by ID
func (k Keeper) GetLoan(ctx sdk.Context, loanID uint64) (types.Loan, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetLoanKey(loanID))
	if bz == nil {
		return types.Loan{}, false
	}

	var loan types.Loan
	if err := json.Unmarshal(bz, &loan); err != nil {
		return types.Loan{}, false
	}
	return loan, true
}

// SetLoan stores a loan
// SECURITY FIX: Returns error instead of panicking
func (k Keeper) SetLoan(ctx sdk.Context, loan types.Loan) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(loan)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal loan", "error", err)
		return fmt.Errorf("failed to marshal loan: %w", err)
	}
	store.Set(types.GetLoanKey(loan.ID), bz)
	return nil
}

// DeleteLoan removes a loan
func (k Keeper) DeleteLoan(ctx sdk.Context, loanID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetLoanKey(loanID))
}

// CreateLoan creates a new loan
func (k Keeper) CreateLoan(
	ctx sdk.Context,
	borrower string,
	principal math.Int,
	interestRate math.LegacyDec,
	duration time.Duration,
	collateral []types.Collateral,
) (types.Loan, error) {
	// TIER CHECK: Borrower must be Keeper+ tier
	if err := k.checkCanBorrow(ctx, borrower); err != nil {
		return types.Loan{}, err
	}

	loanID := k.GetNextLoanID(ctx)

	// Create loan
	loan := types.NewLoan(loanID, borrower, principal, interestRate, duration, collateral)

	// Validate loan
	if err := loan.Validate(); err != nil {
		return types.Loan{}, err
	}

	// Check collateral ratio meets minimum
	if loan.CollateralRatio.LT(loan.MinCollateralRatio) {
		return types.Loan{}, types.ErrCollateralRatioTooLow
	}

	// Lock collateral
	if err := k.lockCollateral(ctx, borrower, collateral); err != nil {
		return types.Loan{}, err
	}

	// Register beneficial owner for equity collateral (so borrower gets dividends)
	k.registerBeneficialOwnerForCollateral(ctx, borrower, loanID, collateral)

	// Store loan
	k.SetLoan(ctx, loan)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLoanCreated,
			sdk.NewAttribute(types.AttributeKeyLoanID, fmt.Sprintf("%d", loan.ID)),
			sdk.NewAttribute(types.AttributeKeyBorrower, borrower),
			sdk.NewAttribute(types.AttributeKeyAmount, principal.String()),
			sdk.NewAttribute(types.AttributeKeyInterestRate, interestRate.String()),
		),
	)

	return loan, nil
}

// ActivateLoan activates a loan (funds borrower)
func (k Keeper) ActivateLoan(ctx sdk.Context, loanID uint64, lender string) error {
	// TIER CHECK: Lender must be Keeper+ tier
	if err := k.checkCanLend(ctx, lender); err != nil {
		return err
	}

	loan, found := k.GetLoan(ctx, loanID)
	if !found {
		return types.ErrLoanNotFound
	}

	if loan.Status != types.LoanStatusPending {
		return types.ErrInvalidLoan
	}

	lenderAddr, err := sdk.AccAddressFromBech32(lender)
	if err != nil {
		return err
	}

	borrowerAddr, err := sdk.AccAddressFromBech32(loan.Borrower)
	if err != nil {
		return err
	}

	// Transfer principal from lender to borrower
	principalCoins := sdk.NewCoins(sdk.NewCoin("hodl", loan.Principal))
	if err := k.bankKeeper.SendCoins(ctx, lenderAddr, borrowerAddr, principalCoins); err != nil {
		return fmt.Errorf("failed to transfer principal: %w", err)
	}

	// Update loan
	loan.Lender = lender
	loan.Status = types.LoanStatusActive
	loan.ActivatedAt = ctx.BlockTime()
	loan.MaturityAt = ctx.BlockTime().Add(loan.Duration)
	loan.LastAccrualAt = ctx.BlockTime()
	k.SetLoan(ctx, loan)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLoanActivated,
			sdk.NewAttribute(types.AttributeKeyLoanID, fmt.Sprintf("%d", loan.ID)),
			sdk.NewAttribute(types.AttributeKeyLender, lender),
		),
	)

	return nil
}

// RepayLoan makes a repayment on a loan
func (k Keeper) RepayLoan(ctx sdk.Context, loanID uint64, repayer string, amount math.Int) error {
	loan, found := k.GetLoan(ctx, loanID)
	if !found {
		return types.ErrLoanNotFound
	}

	if loan.Status != types.LoanStatusActive {
		return types.ErrLoanNotActive
	}

	// Only borrower can repay
	if loan.Borrower != repayer {
		return types.ErrNotBorrower
	}

	repayerAddr, err := sdk.AccAddressFromBech32(repayer)
	if err != nil {
		return err
	}

	lenderAddr, err := sdk.AccAddressFromBech32(loan.Lender)
	if err != nil {
		return err
	}

	// Accrue interest first
	k.accrueInterest(ctx, &loan)

	// Calculate total owed
	totalOwed := loan.TotalOwed

	// Check repayment doesn't exceed owed
	repaymentDec := math.LegacyNewDecFromInt(amount)
	if repaymentDec.GT(totalOwed) {
		return types.ErrRepaymentExceedsOwed
	}

	// Transfer repayment to lender
	repaymentCoins := sdk.NewCoins(sdk.NewCoin("hodl", amount))
	if err := k.bankKeeper.SendCoins(ctx, repayerAddr, lenderAddr, repaymentCoins); err != nil {
		return fmt.Errorf("failed to transfer repayment: %w", err)
	}

	// Update loan tracking
	// First pay interest, then principal
	if repaymentDec.LTE(loan.AccruedInterest) {
		loan.InterestPaid = loan.InterestPaid.Add(repaymentDec)
		loan.AccruedInterest = loan.AccruedInterest.Sub(repaymentDec)
	} else {
		// Pay all remaining interest, rest goes to principal
		remainingAfterInterest := repaymentDec.Sub(loan.AccruedInterest)
		loan.InterestPaid = loan.InterestPaid.Add(loan.AccruedInterest)
		loan.AccruedInterest = math.LegacyZeroDec()
		loan.PrincipalPaid = loan.PrincipalPaid.Add(remainingAfterInterest.TruncateInt())
	}

	// Update total owed
	loan.TotalOwed = loan.TotalOwed.Sub(repaymentDec)

	// Check if fully repaid
	if loan.TotalOwed.LTE(math.LegacyZeroDec()) {
		loan.Status = types.LoanStatusRepaid
		loan.RepaidAt = ctx.BlockTime()

		// Unregister beneficial owner for equity collateral
		k.unregisterBeneficialOwnerForCollateral(ctx, loan.Borrower, loan.ID, loan.Collateral)

		// Release collateral back to borrower
		k.releaseCollateral(ctx, loan.Borrower, loan.Collateral)

		// REWARD: Notify staking system of successful loan repayment
		k.notifyLoanSuccess(ctx, loan.Borrower, loan.ID)
	}

	k.SetLoan(ctx, loan)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLoanRepaid,
			sdk.NewAttribute(types.AttributeKeyLoanID, fmt.Sprintf("%d", loan.ID)),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute("remaining", loan.TotalOwed.String()),
		),
	)

	return nil
}

// LiquidateLoan liquidates an under-collateralized loan
func (k Keeper) LiquidateLoan(ctx sdk.Context, loanID uint64, liquidator string) error {
	loan, found := k.GetLoan(ctx, loanID)
	if !found {
		return types.ErrLoanNotFound
	}

	if loan.Status != types.LoanStatusActive {
		return types.ErrLoanNotActive
	}

	// Accrue interest first
	k.accrueInterest(ctx, &loan)

	// Update collateral value
	k.updateCollateralValue(ctx, &loan)

	// Check if liquidatable
	if !loan.IsLiquidatable() {
		return types.ErrLoanNotLiquidatable
	}

	liquidatorAddr, err := sdk.AccAddressFromBech32(liquidator)
	if err != nil {
		return err
	}

	lenderAddr, err := sdk.AccAddressFromBech32(loan.Lender)
	if err != nil {
		return err
	}

	// Calculate liquidation amounts
	totalOwed := loan.TotalOwed.TruncateInt()
	penalty := loan.CollateralValue.Mul(loan.LiquidationPenalty).TruncateInt()

	// Transfer collateral value minus penalty to lender
	lenderAmount := math.MinInt(totalOwed, loan.CollateralValue.TruncateInt().Sub(penalty))
	if lenderAmount.IsPositive() {
		lenderCoins := sdk.NewCoins(sdk.NewCoin("hodl", lenderAmount))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lenderAddr, lenderCoins); err != nil {
			k.Logger(ctx).Error("failed to send to lender during liquidation", "error", err)
		}
	}

	// Liquidator gets the penalty as reward
	if penalty.IsPositive() {
		liquidatorCoins := sdk.NewCoins(sdk.NewCoin("hodl", penalty))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidatorAddr, liquidatorCoins); err != nil {
			k.Logger(ctx).Error("failed to send penalty to liquidator", "error", err)
		}
	}

	// Unregister beneficial owner - borrower no longer owns the collateral
	k.unregisterBeneficialOwnerForCollateral(ctx, loan.Borrower, loan.ID, loan.Collateral)

	// Update loan status
	loan.Status = types.LoanStatusLiquidated
	k.SetLoan(ctx, loan)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLoanLiquidated,
			sdk.NewAttribute(types.AttributeKeyLoanID, fmt.Sprintf("%d", loan.ID)),
			sdk.NewAttribute("liquidator", liquidator),
			sdk.NewAttribute("penalty", penalty.String()),
		),
	)

	return nil
}

// AddCollateral adds more collateral to a loan
func (k Keeper) AddCollateral(ctx sdk.Context, loanID uint64, adder string, collateral types.Collateral) error {
	loan, found := k.GetLoan(ctx, loanID)
	if !found {
		return types.ErrLoanNotFound
	}

	// Only borrower can add collateral
	if loan.Borrower != adder {
		return types.ErrNotBorrower
	}

	if loan.Status != types.LoanStatusActive && loan.Status != types.LoanStatusPending {
		return types.ErrLoanNotActive
	}

	// Validate and lock new collateral
	if err := collateral.Validate(); err != nil {
		return err
	}

	if err := k.lockCollateral(ctx, adder, []types.Collateral{collateral}); err != nil {
		return err
	}

	// Register beneficial owner for equity collateral
	k.registerBeneficialOwnerForCollateral(ctx, adder, loanID, []types.Collateral{collateral})

	// Add to loan collateral
	loan.Collateral = append(loan.Collateral, collateral)
	loan.CollateralValue = loan.CollateralValue.Add(collateral.Value)
	// SECURITY FIX: Prevent division by zero
	if loan.Principal.IsPositive() {
		loan.CollateralRatio = loan.CollateralValue.QuoInt(loan.Principal)
	} else {
		loan.CollateralRatio = math.LegacyZeroDec()
	}
	k.SetLoan(ctx, loan)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCollateralAdded,
			sdk.NewAttribute(types.AttributeKeyLoanID, fmt.Sprintf("%d", loan.ID)),
			sdk.NewAttribute("denom", collateral.Denom),
			sdk.NewAttribute("amount", collateral.Amount.String()),
		),
	)

	return nil
}

// ============ Pool Management ============

// GetLendingPool returns a lending pool by ID
func (k Keeper) GetLendingPool(ctx sdk.Context, poolID uint64) (types.LendingPool, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetLendingPoolKey(poolID))
	if bz == nil {
		return types.LendingPool{}, false
	}

	var pool types.LendingPool
	if err := json.Unmarshal(bz, &pool); err != nil {
		return types.LendingPool{}, false
	}
	return pool, true
}

// SetLendingPool stores a lending pool
// SECURITY FIX: Returns error instead of panicking
func (k Keeper) SetLendingPool(ctx sdk.Context, pool types.LendingPool) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(pool)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal lending pool", "error", err)
		return fmt.Errorf("failed to marshal lending pool: %w", err)
	}
	store.Set(types.GetLendingPoolKey(pool.ID), bz)
	return nil
}

// CreateLendingPool creates a new lending pool
func (k Keeper) CreateLendingPool(ctx sdk.Context, name string) (types.LendingPool, error) {
	poolID := k.GetNextPoolID(ctx)
	pool := types.NewLendingPool(poolID, name)

	if err := pool.Validate(); err != nil {
		return types.LendingPool{}, err
	}

	k.SetLendingPool(ctx, pool)
	return pool, nil
}

// DepositToPool deposits HODL into a lending pool
func (k Keeper) DepositToPool(ctx sdk.Context, poolID uint64, depositor string, amount math.Int) error {
	pool, found := k.GetLendingPool(ctx, poolID)
	if !found {
		return types.ErrPoolNotFound
	}

	depositorAddr, err := sdk.AccAddressFromBech32(depositor)
	if err != nil {
		return err
	}

	// Transfer HODL to module
	depositCoins := sdk.NewCoins(sdk.NewCoin("hodl", amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositorAddr, types.ModuleName, depositCoins); err != nil {
		return err
	}

	// Calculate pool tokens to mint
	var poolTokens math.Int
	if pool.TotalDeposits.IsZero() {
		poolTokens = amount
	} else {
		// poolTokens = amount * totalPoolTokens / totalDeposits
		poolTokens = amount.Mul(pool.PoolTokenSupply).Quo(pool.TotalDeposits)
	}

	// Update pool
	pool.TotalDeposits = pool.TotalDeposits.Add(amount)
	pool.AvailableLiquidity = pool.AvailableLiquidity.Add(amount)
	pool.PoolTokenSupply = pool.PoolTokenSupply.Add(poolTokens)
	pool.UpdatedAt = ctx.BlockTime()
	k.updatePoolRates(ctx, &pool)
	k.SetLendingPool(ctx, pool)

	// Store deposit record
	deposit := types.PoolDeposit{
		User:          depositor,
		PoolID:        poolID,
		DepositAmount: amount,
		PoolTokens:    poolTokens,
		DepositedAt:   ctx.BlockTime(),
		LastClaimAt:   ctx.BlockTime(),
	}
	k.setPoolDeposit(ctx, deposit)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePoolDeposit,
			sdk.NewAttribute(types.AttributeKeyPoolID, fmt.Sprintf("%d", poolID)),
			sdk.NewAttribute("depositor", depositor),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("pool_tokens", poolTokens.String()),
		),
	)

	return nil
}

// WithdrawFromPool withdraws HODL from a lending pool
func (k Keeper) WithdrawFromPool(ctx sdk.Context, poolID uint64, withdrawer string, poolTokens math.Int) error {
	pool, found := k.GetLendingPool(ctx, poolID)
	if !found {
		return types.ErrPoolNotFound
	}

	deposit, found := k.getPoolDeposit(ctx, poolID, withdrawer)
	if !found {
		return types.ErrDepositNotFound
	}

	if poolTokens.GT(deposit.PoolTokens) {
		return types.ErrInvalidWithdrawal
	}

	// SECURITY FIX: Check pool token supply is positive before division
	if pool.PoolTokenSupply.IsZero() {
		return types.ErrInvalidWithdrawal
	}

	// Calculate HODL to return
	// amount = poolTokens * totalDeposits / totalPoolTokens
	withdrawAmount := poolTokens.Mul(pool.TotalDeposits).Quo(pool.PoolTokenSupply)

	// Check liquidity
	if withdrawAmount.GT(pool.AvailableLiquidity) {
		return types.ErrInsufficientLiquidity
	}

	withdrawerAddr, err := sdk.AccAddressFromBech32(withdrawer)
	if err != nil {
		return err
	}

	// Transfer HODL to withdrawer
	withdrawCoins := sdk.NewCoins(sdk.NewCoin("hodl", withdrawAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawerAddr, withdrawCoins); err != nil {
		return err
	}

	// Update pool
	pool.TotalDeposits = pool.TotalDeposits.Sub(withdrawAmount)
	pool.AvailableLiquidity = pool.AvailableLiquidity.Sub(withdrawAmount)
	pool.PoolTokenSupply = pool.PoolTokenSupply.Sub(poolTokens)
	pool.UpdatedAt = ctx.BlockTime()
	k.updatePoolRates(ctx, &pool)
	k.SetLendingPool(ctx, pool)

	// Update deposit record
	deposit.PoolTokens = deposit.PoolTokens.Sub(poolTokens)
	if deposit.PoolTokens.IsZero() {
		k.deletePoolDeposit(ctx, poolID, withdrawer)
	} else {
		k.setPoolDeposit(ctx, deposit)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePoolWithdrawal,
			sdk.NewAttribute(types.AttributeKeyPoolID, fmt.Sprintf("%d", poolID)),
			sdk.NewAttribute("withdrawer", withdrawer),
			sdk.NewAttribute("amount", withdrawAmount.String()),
		),
	)

	return nil
}

// BorrowFromPool borrows HODL from a lending pool
func (k Keeper) BorrowFromPool(ctx sdk.Context, poolID uint64, borrower string, amount math.Int, collateral []types.Collateral) error {
	pool, found := k.GetLendingPool(ctx, poolID)
	if !found {
		return types.ErrPoolNotFound
	}

	// Check liquidity
	if amount.GT(pool.AvailableLiquidity) {
		return types.ErrInsufficientLiquidity
	}

	// Create loan against pool with governance-controllable duration
	interestRate := pool.CalculateBorrowRate()
	duration := k.GetDefaultLoanDuration(ctx)

	// CreateLoan will handle beneficial owner registration
	loan, err := k.CreateLoan(ctx, borrower, amount, interestRate, duration, collateral)
	if err != nil {
		return err
	}

	// Activate loan from pool
	borrowerAddr, err := sdk.AccAddressFromBech32(borrower)
	if err != nil {
		return err
	}

	// Transfer from module (pool) to borrower
	borrowCoins := sdk.NewCoins(sdk.NewCoin("hodl", amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrowerAddr, borrowCoins); err != nil {
		return err
	}

	// Update loan
	loan.Lender = "pool:" + fmt.Sprintf("%d", poolID)
	loan.Status = types.LoanStatusActive
	loan.ActivatedAt = ctx.BlockTime()
	loan.MaturityAt = ctx.BlockTime().Add(duration)
	k.SetLoan(ctx, loan)

	// Update pool
	pool.TotalBorrowed = pool.TotalBorrowed.Add(amount)
	pool.AvailableLiquidity = pool.AvailableLiquidity.Sub(amount)
	pool.UpdatedAt = ctx.BlockTime()
	k.updatePoolRates(ctx, &pool)
	k.SetLendingPool(ctx, pool)

	return nil
}

// ============ Helper Functions ============

// isEquityCollateral checks if a collateral item represents equity shares
func (k Keeper) isEquityCollateral(c types.Collateral) bool {
	return c.Type == types.CollateralTypeEquity && c.CompanyID > 0 && c.ShareClass != ""
}

// lockCollateral locks collateral in the module account
func (k Keeper) lockCollateral(ctx sdk.Context, owner string, collateral []types.Collateral) error {
	ownerAddr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return err
	}

	for _, c := range collateral {
		coins := sdk.NewCoins(sdk.NewCoin(c.Denom, c.Amount))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, ownerAddr, types.ModuleName, coins); err != nil {
			return fmt.Errorf("failed to lock %s: %w", c.Denom, err)
		}
	}

	return nil
}

// registerBeneficialOwnerForCollateral registers beneficial ownership for equity collateral
// This ensures the borrower continues to receive dividends on shares locked as collateral
func (k Keeper) registerBeneficialOwnerForCollateral(
	ctx sdk.Context,
	owner string,
	loanID uint64,
	collateral []types.Collateral,
) {
	if k.equityKeeper == nil {
		return
	}

	for _, c := range collateral {
		// Only register for equity collateral
		if !k.isEquityCollateral(c) {
			continue
		}

		err := k.equityKeeper.RegisterBeneficialOwner(
			ctx,
			types.ModuleName,    // "lending" module holds the shares
			c.CompanyID,
			c.ShareClass,
			owner,               // Borrower is the beneficial owner
			c.Amount,
			loanID,              // Reference to the loan
			"loan",              // Reference type
		)
		if err != nil {
			k.Logger(ctx).Error("failed to register beneficial owner for equity collateral",
				"error", err,
				"owner", owner,
				"loan_id", loanID,
				"company_id", c.CompanyID,
				"class", c.ShareClass,
				"shares", c.Amount.String(),
			)
		} else {
			k.Logger(ctx).Info("registered beneficial owner for equity collateral",
				"owner", owner,
				"loan_id", loanID,
				"company_id", c.CompanyID,
				"class", c.ShareClass,
				"shares", c.Amount.String(),
			)
		}
	}
}

// unregisterBeneficialOwnerForCollateral removes beneficial ownership when collateral is released
func (k Keeper) unregisterBeneficialOwnerForCollateral(
	ctx sdk.Context,
	owner string,
	loanID uint64,
	collateral []types.Collateral,
) {
	if k.equityKeeper == nil {
		return
	}

	for _, c := range collateral {
		// Only unregister for equity collateral
		if !k.isEquityCollateral(c) {
			continue
		}

		err := k.equityKeeper.UnregisterBeneficialOwner(
			ctx,
			types.ModuleName,
			c.CompanyID,
			c.ShareClass,
			owner,
			loanID,
		)
		if err != nil {
			k.Logger(ctx).Error("failed to unregister beneficial owner",
				"error", err,
				"owner", owner,
				"loan_id", loanID,
				"company_id", c.CompanyID,
				"class", c.ShareClass,
			)
		} else {
			k.Logger(ctx).Info("unregistered beneficial owner for equity collateral",
				"owner", owner,
				"loan_id", loanID,
				"company_id", c.CompanyID,
				"class", c.ShareClass,
			)
		}
	}
}

// releaseCollateral releases collateral back to owner
func (k Keeper) releaseCollateral(ctx sdk.Context, owner string, collateral []types.Collateral) error {
	ownerAddr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return err
	}

	for _, c := range collateral {
		coins := sdk.NewCoins(sdk.NewCoin(c.Denom, c.Amount))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, ownerAddr, coins); err != nil {
			k.Logger(ctx).Error("failed to release collateral", "error", err, "denom", c.Denom)
		}
	}

	return nil
}

// accrueInterest accrues interest on a loan
func (k Keeper) accrueInterest(ctx sdk.Context, loan *types.Loan) {
	interest := loan.CalculateAccruedInterest(ctx.BlockTime())
	loan.AccruedInterest = loan.AccruedInterest.Add(interest)
	loan.TotalOwed = loan.TotalOwed.Add(interest)
	loan.LastAccrualAt = ctx.BlockTime()
}

// updateCollateralValue updates the collateral value based on current prices
func (k Keeper) updateCollateralValue(ctx sdk.Context, loan *types.Loan) {
	totalValue := math.LegacyZeroDec()
	for i := range loan.Collateral {
		// Get current price (simplified - would use price oracle)
		price := k.getCollateralPrice(ctx, loan.Collateral[i].Denom)
		loan.Collateral[i].Value = price.MulInt(loan.Collateral[i].Amount)
		totalValue = totalValue.Add(loan.Collateral[i].Value)
	}

	loan.CollateralValue = totalValue
	outstandingPrincipal := loan.Principal.Sub(loan.PrincipalPaid)
	if !outstandingPrincipal.IsZero() {
		loan.CollateralRatio = totalValue.QuoInt(outstandingPrincipal)
	}
}

// getCollateralPrice gets the price of collateral (simplified)
func (k Keeper) getCollateralPrice(ctx sdk.Context, denom string) math.LegacyDec {
	if denom == "hodl" {
		return math.LegacyOneDec()
	}
	// For equity, would query DEX for market price
	// Simplified: return 1.0 for now
	return math.LegacyOneDec()
}

// updatePoolRates updates pool interest rates
func (k Keeper) updatePoolRates(ctx sdk.Context, pool *types.LendingPool) {
	if pool.TotalDeposits.IsZero() {
		pool.UtilizationRate = math.LegacyZeroDec()
	} else {
		pool.UtilizationRate = math.LegacyNewDecFromInt(pool.TotalBorrowed).QuoInt(pool.TotalDeposits)
	}
	pool.BorrowAPY = pool.CalculateBorrowRate()
	pool.SupplyAPY = pool.CalculateSupplyRate()
}

// Pool deposit helpers
func (k Keeper) getPoolDeposit(ctx sdk.Context, poolID uint64, user string) (types.PoolDeposit, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolDepositKey(poolID, user))
	if bz == nil {
		return types.PoolDeposit{}, false
	}

	var deposit types.PoolDeposit
	if err := json.Unmarshal(bz, &deposit); err != nil {
		return types.PoolDeposit{}, false
	}
	return deposit, true
}

func (k Keeper) setPoolDeposit(ctx sdk.Context, deposit types.PoolDeposit) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(deposit)
	store.Set(types.GetPoolDepositKey(deposit.PoolID, deposit.User), bz)
}

func (k Keeper) deletePoolDeposit(ctx sdk.Context, poolID uint64, user string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPoolDepositKey(poolID, user))
}

// ============ Periodic Tasks ============

// ProcessLoans processes all active loans (interest accrual, liquidation checks)
func (k Keeper) ProcessLoans(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.LoanPrefix).Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var loan types.Loan
		if err := json.Unmarshal(iterator.Value(), &loan); err != nil {
			continue
		}

		if loan.Status != types.LoanStatusActive {
			continue
		}

		// Accrue interest
		k.accrueInterest(ctx, &loan)

		// Update collateral value
		k.updateCollateralValue(ctx, &loan)

		// Check for default (overdue)
		if loan.IsOverdue(ctx.BlockTime()) {
			loan.Status = types.LoanStatusDefaulted

			// PENALTY: Notify staking system of loan default
			k.notifyLoanDefault(ctx, loan.Borrower, loan.ID, loan.TotalOwed.TruncateInt())

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeLoanDefaulted,
					sdk.NewAttribute(types.AttributeKeyLoanID, fmt.Sprintf("%d", loan.ID)),
				),
			)
		}

		k.SetLoan(ctx, loan)
	}
}

// GetAllLoans returns all loans
func (k Keeper) GetAllLoans(ctx sdk.Context) []types.Loan {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.LoanPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var loans []types.Loan
	for ; iterator.Valid(); iterator.Next() {
		var loan types.Loan
		if err := json.Unmarshal(iterator.Value(), &loan); err != nil {
			continue
		}
		loans = append(loans, loan)
	}

	return loans
}

// GetUserLoans returns all loans for a user
func (k Keeper) GetUserLoans(ctx sdk.Context, user string) []types.Loan {
	allLoans := k.GetAllLoans(ctx)
	var userLoans []types.Loan
	for _, loan := range allLoans {
		if loan.Borrower == user || loan.Lender == user {
			userLoans = append(userLoans, loan)
		}
	}
	return userLoans
}

// GetLiquidatableLoans returns all loans that can be liquidated
func (k Keeper) GetLiquidatableLoans(ctx sdk.Context) []types.Loan {
	allLoans := k.GetAllLoans(ctx)
	var liquidatable []types.Loan
	for _, loan := range allLoans {
		if loan.IsLiquidatable() {
			liquidatable = append(liquidatable, loan)
		}
	}
	return liquidatable
}

// ============ Stake-as-Trust-Ceiling: P2P Lending Stake Management ============

// GetLenderStake returns a lender's stake record
func (k Keeper) GetLenderStake(ctx sdk.Context, address string) (types.LenderStake, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetLenderStakeKey(address))
	if bz == nil {
		return types.LenderStake{}, false
	}

	var stake types.LenderStake
	if err := json.Unmarshal(bz, &stake); err != nil {
		return types.LenderStake{}, false
	}
	return stake, true
}

// SetLenderStake stores a lender's stake record
func (k Keeper) SetLenderStake(ctx sdk.Context, stake types.LenderStake) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(stake)
	if err != nil {
		return fmt.Errorf("failed to marshal lender stake: %w", err)
	}
	store.Set(types.GetLenderStakeKey(stake.Address), bz)
	return nil
}

// DeleteLenderStake removes a lender's stake record
func (k Keeper) DeleteLenderStake(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetLenderStakeKey(address))
}

// GetBorrowerStake returns a borrower's stake record
func (k Keeper) GetBorrowerStake(ctx sdk.Context, address string) (types.BorrowerStake, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetBorrowerStakeKey(address))
	if bz == nil {
		return types.BorrowerStake{}, false
	}

	var stake types.BorrowerStake
	if err := json.Unmarshal(bz, &stake); err != nil {
		return types.BorrowerStake{}, false
	}
	return stake, true
}

// SetBorrowerStake stores a borrower's stake record
func (k Keeper) SetBorrowerStake(ctx sdk.Context, stake types.BorrowerStake) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(stake)
	if err != nil {
		return fmt.Errorf("failed to marshal borrower stake: %w", err)
	}
	store.Set(types.GetBorrowerStakeKey(stake.Address), bz)
	return nil
}

// DeleteBorrowerStake removes a borrower's stake record
func (k Keeper) DeleteBorrowerStake(ctx sdk.Context, address string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetBorrowerStakeKey(address))
}

// DepositLenderStake deposits security stake for a lender
func (k Keeper) DepositLenderStake(ctx sdk.Context, address string, amount math.Int) error {
	lenderAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	// Transfer stake to module
	stakeCoins := sdk.NewCoins(sdk.NewCoin("hodl", amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, lenderAddr, types.ModuleName, stakeCoins); err != nil {
		return err
	}

	// Update or create stake record
	stake, found := k.GetLenderStake(ctx, address)
	if !found {
		stake = types.NewLenderStake(address, math.ZeroInt())
	}

	stake.TotalStake = stake.TotalStake.Add(amount)
	stake.AvailableStake = stake.AvailableStake.Add(amount)
	if err := k.SetLenderStake(ctx, stake); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLenderStakeDeposited,
			sdk.NewAttribute(types.AttributeKeyLender, address),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyTrustCeiling, stake.TotalStake.String()),
		),
	)

	return nil
}

// DepositBorrowerStake deposits security stake for a borrower
func (k Keeper) DepositBorrowerStake(ctx sdk.Context, address string, amount math.Int) error {
	borrowerAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	// Transfer stake to module
	stakeCoins := sdk.NewCoins(sdk.NewCoin("hodl", amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, borrowerAddr, types.ModuleName, stakeCoins); err != nil {
		return err
	}

	// Update or create stake record
	stake, found := k.GetBorrowerStake(ctx, address)
	if !found {
		stake = types.NewBorrowerStake(address, math.ZeroInt())
	}

	stake.TotalStake = stake.TotalStake.Add(amount)
	stake.AvailableStake = stake.AvailableStake.Add(amount)
	if err := k.SetBorrowerStake(ctx, stake); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBorrowerStakeDeposited,
			sdk.NewAttribute(types.AttributeKeyBorrower, address),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyTrustCeiling, stake.TotalStake.String()),
		),
	)

	return nil
}

// GetLenderTrustCeiling returns maximum offer value a lender can post
func (k Keeper) GetLenderTrustCeiling(ctx sdk.Context, address string) (math.Int, error) {
	stake, found := k.GetLenderStake(ctx, address)
	if !found {
		return math.ZeroInt(), types.ErrLenderStakeNotFound
	}
	return stake.GetTrustCeiling(), nil
}

// GetBorrowerTrustCeiling returns maximum request value a borrower can post
func (k Keeper) GetBorrowerTrustCeiling(ctx sdk.Context, address string) (math.Int, error) {
	stake, found := k.GetBorrowerStake(ctx, address)
	if !found {
		return math.ZeroInt(), types.ErrBorrowerStakeNotFound
	}
	return stake.GetTrustCeiling(), nil
}

// CanLenderPostOffer checks if lender can post an offer of given value
func (k Keeper) CanLenderPostOffer(ctx sdk.Context, address string, offerValue math.Int) error {
	stake, found := k.GetLenderStake(ctx, address)
	if !found {
		return types.ErrLenderStakeNotFound
	}

	if !stake.CanPostOffer(offerValue) {
		return types.ErrOfferExceedsTrustCeiling
	}

	return nil
}

// CanBorrowerPostRequest checks if borrower can post a request of given value
func (k Keeper) CanBorrowerPostRequest(ctx sdk.Context, address string, requestValue math.Int) error {
	stake, found := k.GetBorrowerStake(ctx, address)
	if !found {
		return types.ErrBorrowerStakeNotFound
	}

	if !stake.CanPostRequest(requestValue) {
		return types.ErrRequestExceedsTrustCeiling
	}

	return nil
}

// LockLenderStake locks stake when an offer is created
func (k Keeper) LockLenderStake(ctx sdk.Context, address string, amount math.Int) error {
	stake, found := k.GetLenderStake(ctx, address)
	if !found {
		return types.ErrLenderStakeNotFound
	}

	if stake.AvailableStake.LT(amount) {
		return types.ErrInsufficientLenderStake
	}

	stake.AvailableStake = stake.AvailableStake.Sub(amount)
	stake.LockedStake = stake.LockedStake.Add(amount)
	stake.ActiveOffers++
	stake.TotalOfferValue = stake.TotalOfferValue.Add(amount)

	if err := k.SetLenderStake(ctx, stake); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStakeLocked,
			sdk.NewAttribute(types.AttributeKeyLender, address),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyAvailableStake, stake.AvailableStake.String()),
		),
	)

	return nil
}

// ReleaseLenderStake releases stake when offer is cancelled or filled
func (k Keeper) ReleaseLenderStake(ctx sdk.Context, address string, amount math.Int) error {
	stake, found := k.GetLenderStake(ctx, address)
	if !found {
		return types.ErrLenderStakeNotFound
	}

	// Release from locked back to available
	if stake.LockedStake.LT(amount) {
		amount = stake.LockedStake // Cap at locked amount
	}

	stake.LockedStake = stake.LockedStake.Sub(amount)
	stake.AvailableStake = stake.AvailableStake.Add(amount)
	if stake.ActiveOffers > 0 {
		stake.ActiveOffers--
	}
	stake.TotalOfferValue = stake.TotalOfferValue.Sub(amount)
	if stake.TotalOfferValue.IsNegative() {
		stake.TotalOfferValue = math.ZeroInt()
	}

	if err := k.SetLenderStake(ctx, stake); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStakeReleased,
			sdk.NewAttribute(types.AttributeKeyLender, address),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyAvailableStake, stake.AvailableStake.String()),
		),
	)

	return nil
}

// LockBorrowerStake locks stake when a request is created
func (k Keeper) LockBorrowerStake(ctx sdk.Context, address string, amount math.Int) error {
	stake, found := k.GetBorrowerStake(ctx, address)
	if !found {
		return types.ErrBorrowerStakeNotFound
	}

	if stake.AvailableStake.LT(amount) {
		return types.ErrInsufficientBorrowerStake
	}

	stake.AvailableStake = stake.AvailableStake.Sub(amount)
	stake.LockedStake = stake.LockedStake.Add(amount)
	stake.ActiveRequests++
	stake.TotalRequestValue = stake.TotalRequestValue.Add(amount)

	if err := k.SetBorrowerStake(ctx, stake); err != nil {
		return err
	}

	return nil
}

// ReleaseBorrowerStake releases stake when request is cancelled or filled
func (k Keeper) ReleaseBorrowerStake(ctx sdk.Context, address string, amount math.Int) error {
	stake, found := k.GetBorrowerStake(ctx, address)
	if !found {
		return types.ErrBorrowerStakeNotFound
	}

	if stake.LockedStake.LT(amount) {
		amount = stake.LockedStake
	}

	stake.LockedStake = stake.LockedStake.Sub(amount)
	stake.AvailableStake = stake.AvailableStake.Add(amount)
	if stake.ActiveRequests > 0 {
		stake.ActiveRequests--
	}
	stake.TotalRequestValue = stake.TotalRequestValue.Sub(amount)
	if stake.TotalRequestValue.IsNegative() {
		stake.TotalRequestValue = math.ZeroInt()
	}

	return k.SetBorrowerStake(ctx, stake)
}

// RequestLenderUnstake initiates unbonding for lender stake (14-day period)
func (k Keeper) RequestLenderUnstake(ctx sdk.Context, address string, amount math.Int) error {
	stake, found := k.GetLenderStake(ctx, address)
	if !found {
		return types.ErrLenderStakeNotFound
	}

	// Check no active offers
	if stake.ActiveOffers > 0 {
		return types.ErrActiveOffersExist
	}

	// Check not already unbonding
	if !stake.UnbondingAmount.IsZero() {
		return types.ErrStakeUnbondingInProgress
	}

	if amount.GT(stake.AvailableStake) {
		return types.ErrInsufficientLenderStake
	}

	stake.AvailableStake = stake.AvailableStake.Sub(amount)
	stake.UnbondingAmount = amount
	stake.UnbondingAt = ctx.BlockTime()

	if err := k.SetLenderStake(ctx, stake); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLenderUnstakeRequested,
			sdk.NewAttribute(types.AttributeKeyLender, address),
			sdk.NewAttribute(types.AttributeKeyUnbondingAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletesAt, stake.UnbondingAt.Add(k.GetLenderUnbondingPeriod(ctx)).String()),
		),
	)

	return nil
}

// CompleteLenderUnstake finalizes lender unstaking after unbonding period
func (k Keeper) CompleteLenderUnstake(ctx sdk.Context, address string) error {
	stake, found := k.GetLenderStake(ctx, address)
	if !found {
		return types.ErrLenderStakeNotFound
	}

	if stake.UnbondingAmount.IsZero() {
		return types.ErrInvalidWithdrawal
	}

	if !stake.IsUnbondingComplete(ctx.BlockTime()) {
		return types.ErrStakeUnbondingNotComplete
	}

	// Transfer unbonded stake back to lender
	lenderAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}
	unstakeCoins := sdk.NewCoins(sdk.NewCoin("hodl", stake.UnbondingAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lenderAddr, unstakeCoins); err != nil {
		return err
	}

	withdrawnAmount := stake.UnbondingAmount

	// Update stake
	stake.TotalStake = stake.TotalStake.Sub(stake.UnbondingAmount)
	stake.UnbondingAmount = math.ZeroInt()
	stake.UnbondingAt = time.Time{}

	// Delete stake record if completely withdrawn
	if stake.TotalStake.IsZero() {
		k.DeleteLenderStake(ctx, address)
	} else {
		if err := k.SetLenderStake(ctx, stake); err != nil {
			return err
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLenderStakeWithdrawn,
			sdk.NewAttribute(types.AttributeKeyLender, address),
			sdk.NewAttribute(types.AttributeKeyAmount, withdrawnAmount.String()),
		),
	)

	return nil
}

// RequestBorrowerUnstake initiates unbonding for borrower stake
func (k Keeper) RequestBorrowerUnstake(ctx sdk.Context, address string, amount math.Int) error {
	stake, found := k.GetBorrowerStake(ctx, address)
	if !found {
		return types.ErrBorrowerStakeNotFound
	}

	// Check no active requests
	if stake.ActiveRequests > 0 {
		return types.ErrActiveRequestsExist
	}

	// Check not already unbonding
	if !stake.UnbondingAmount.IsZero() {
		return types.ErrStakeUnbondingInProgress
	}

	if amount.GT(stake.AvailableStake) {
		return types.ErrInsufficientBorrowerStake
	}

	stake.AvailableStake = stake.AvailableStake.Sub(amount)
	stake.UnbondingAmount = amount
	stake.UnbondingAt = ctx.BlockTime()

	if err := k.SetBorrowerStake(ctx, stake); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBorrowerUnstakeRequested,
			sdk.NewAttribute(types.AttributeKeyBorrower, address),
			sdk.NewAttribute(types.AttributeKeyUnbondingAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletesAt, stake.UnbondingAt.Add(k.GetBorrowerUnbondingPeriod(ctx)).String()),
		),
	)

	return nil
}

// CompleteBorrowerUnstake finalizes borrower unstaking after unbonding period
func (k Keeper) CompleteBorrowerUnstake(ctx sdk.Context, address string) error {
	stake, found := k.GetBorrowerStake(ctx, address)
	if !found {
		return types.ErrBorrowerStakeNotFound
	}

	if stake.UnbondingAmount.IsZero() {
		return types.ErrInvalidWithdrawal
	}

	if !stake.IsUnbondingComplete(ctx.BlockTime()) {
		return types.ErrStakeUnbondingNotComplete
	}

	// Transfer unbonded stake back to borrower
	borrowerAddr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}
	unstakeCoins := sdk.NewCoins(sdk.NewCoin("hodl", stake.UnbondingAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrowerAddr, unstakeCoins); err != nil {
		return err
	}

	withdrawnAmount := stake.UnbondingAmount

	// Update stake
	stake.TotalStake = stake.TotalStake.Sub(stake.UnbondingAmount)
	stake.UnbondingAmount = math.ZeroInt()
	stake.UnbondingAt = time.Time{}

	// Delete stake record if completely withdrawn
	if stake.TotalStake.IsZero() {
		k.DeleteBorrowerStake(ctx, address)
	} else {
		if err := k.SetBorrowerStake(ctx, stake); err != nil {
			return err
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBorrowerStakeWithdrawn,
			sdk.NewAttribute(types.AttributeKeyBorrower, address),
			sdk.NewAttribute(types.AttributeKeyAmount, withdrawnAmount.String()),
		),
	)

	return nil
}

// GetNextOfferID returns the next offer ID and increments the counter
func (k Keeper) GetNextOfferID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.OfferCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.OfferCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetNextRequestID returns the next request ID and increments the counter
func (k Keeper) GetNextRequestID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.RequestCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.RequestCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetLoanOffer returns a loan offer by ID
func (k Keeper) GetLoanOffer(ctx sdk.Context, offerID uint64) (types.LoanOffer, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetLoanOfferKey(offerID))
	if bz == nil {
		return types.LoanOffer{}, false
	}

	var offer types.LoanOffer
	if err := json.Unmarshal(bz, &offer); err != nil {
		return types.LoanOffer{}, false
	}
	return offer, true
}

// SetLoanOffer stores a loan offer
func (k Keeper) SetLoanOffer(ctx sdk.Context, offer types.LoanOffer) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(offer)
	if err != nil {
		return fmt.Errorf("failed to marshal loan offer: %w", err)
	}
	store.Set(types.GetLoanOfferKey(offer.ID), bz)
	return nil
}

// DeleteLoanOffer removes a loan offer
func (k Keeper) DeleteLoanOffer(ctx sdk.Context, offerID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetLoanOfferKey(offerID))
}

// GetLoanRequest returns a loan request by ID
func (k Keeper) GetLoanRequest(ctx sdk.Context, requestID uint64) (types.LoanRequest, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetLoanRequestKey(requestID))
	if bz == nil {
		return types.LoanRequest{}, false
	}

	var request types.LoanRequest
	if err := json.Unmarshal(bz, &request); err != nil {
		return types.LoanRequest{}, false
	}
	return request, true
}

// SetLoanRequest stores a loan request
func (k Keeper) SetLoanRequest(ctx sdk.Context, request types.LoanRequest) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal loan request: %w", err)
	}
	store.Set(types.GetLoanRequestKey(request.ID), bz)
	return nil
}

// DeleteLoanRequest removes a loan request
func (k Keeper) DeleteLoanRequest(ctx sdk.Context, requestID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetLoanRequestKey(requestID))
}

// ============ Validator Oversight: P2P Lending Slashing ============

// canSlashAsValidator checks if an address has validator oversight permissions
// Warden+ (tier 2+) can perform slashing oversight
func (k Keeper) canSlashAsValidator(ctx sdk.Context, addr string) (bool, int, error) {
	if k.stakingKeeper == nil {
		return false, 0, types.ErrUnauthorizedValidatorAction
	}
	accAddr, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return false, 0, err
	}
	tier := k.stakingKeeper.GetUserTierInt(ctx, accAddr)
	// Warden+ (tier 2+) can perform slashing
	return tier >= 2, tier, nil
}

// SlashLenderStake allows validators to slash a lender's stake
// Warden+ validators (tier 2+) can slash lenders
func (k Keeper) SlashLenderStake(
	ctx sdk.Context,
	validatorAddr string,
	lenderAddr string,
	slashFraction math.LegacyDec,
	reason string,
) error {
	// Verify caller has validator oversight permissions (Warden+ tier)
	canSlash, tier, err := k.canSlashAsValidator(ctx, validatorAddr)
	if err != nil {
		return err
	}
	if !canSlash {
		return types.ErrValidatorTierTooLow
	}

	k.Logger(ctx).Info("validator slashing lender stake",
		"validator", validatorAddr,
		"tier", tier,
		"lender", lenderAddr,
	)

	// Validate slash fraction
	if slashFraction.LT(math.LegacyZeroDec()) || slashFraction.GT(math.LegacyOneDec()) {
		return types.ErrInvalidSlashFraction
	}

	stake, found := k.GetLenderStake(ctx, lenderAddr)
	if !found {
		return types.ErrLenderStakeNotFound
	}

	// Calculate slash amount (from both available and unbonding)
	totalSlashable := stake.TotalStake
	slashAmount := slashFraction.MulInt(totalSlashable).TruncateInt()

	if slashAmount.IsZero() {
		return nil // Nothing to slash
	}

	// Burn slashed tokens FIRST - must succeed before modifying state
	slashCoins := sdk.NewCoins(sdk.NewCoin("hodl", slashAmount))
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, slashCoins); err != nil {
		k.Logger(ctx).Error("failed to burn slashed lender tokens, aborting", "error", err)
		return fmt.Errorf("failed to burn slashed tokens: %w", err)
	}

	// Update stake ONLY after successful burn
	stake.TotalStake = stake.TotalStake.Sub(slashAmount)
	if stake.TotalStake.IsNegative() {
		stake.TotalStake = math.ZeroInt()
	}
	stake.AvailableStake = stake.TotalStake.Sub(stake.LockedStake).Sub(stake.UnbondingAmount)
	if stake.AvailableStake.IsNegative() {
		stake.AvailableStake = math.ZeroInt()
	}

	if stake.TotalStake.IsZero() {
		k.DeleteLenderStake(ctx, lenderAddr)
	} else {
		if err := k.SetLenderStake(ctx, stake); err != nil {
			return err
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLenderSlashed,
			sdk.NewAttribute(types.AttributeKeyLender, lenderAddr),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
			sdk.NewAttribute(types.AttributeKeySlashAmount, slashAmount.String()),
			sdk.NewAttribute(types.AttributeKeySlashFraction, slashFraction.String()),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)

	k.Logger(ctx).Info("lender stake slashed",
		"lender", lenderAddr,
		"validator", validatorAddr,
		"amount", slashAmount.String(),
		"reason", reason,
	)

	return nil
}

// SlashBorrowerStake allows validators to slash a borrower's stake
// Warden+ validators (tier 2+) can slash borrowers
func (k Keeper) SlashBorrowerStake(
	ctx sdk.Context,
	validatorAddr string,
	borrowerAddr string,
	slashFraction math.LegacyDec,
	reason string,
) error {
	// Verify caller has validator oversight permissions (Warden+ tier)
	canSlash, tier, err := k.canSlashAsValidator(ctx, validatorAddr)
	if err != nil {
		return err
	}
	if !canSlash {
		return types.ErrValidatorTierTooLow
	}

	k.Logger(ctx).Info("validator slashing borrower stake",
		"validator", validatorAddr,
		"tier", tier,
		"borrower", borrowerAddr,
	)

	// Validate slash fraction
	if slashFraction.LT(math.LegacyZeroDec()) || slashFraction.GT(math.LegacyOneDec()) {
		return types.ErrInvalidSlashFraction
	}

	stake, found := k.GetBorrowerStake(ctx, borrowerAddr)
	if !found {
		return types.ErrBorrowerStakeNotFound
	}

	// Calculate slash amount
	totalSlashable := stake.TotalStake
	slashAmount := slashFraction.MulInt(totalSlashable).TruncateInt()

	if slashAmount.IsZero() {
		return nil // Nothing to slash
	}

	// Burn slashed tokens FIRST - must succeed before modifying state
	slashCoins := sdk.NewCoins(sdk.NewCoin("hodl", slashAmount))
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, slashCoins); err != nil {
		k.Logger(ctx).Error("failed to burn slashed borrower tokens, aborting", "error", err)
		return fmt.Errorf("failed to burn slashed tokens: %w", err)
	}

	// Update stake ONLY after successful burn
	stake.TotalStake = stake.TotalStake.Sub(slashAmount)
	if stake.TotalStake.IsNegative() {
		stake.TotalStake = math.ZeroInt()
	}
	stake.AvailableStake = stake.TotalStake.Sub(stake.LockedStake).Sub(stake.UnbondingAmount)
	if stake.AvailableStake.IsNegative() {
		stake.AvailableStake = math.ZeroInt()
	}

	if stake.TotalStake.IsZero() {
		k.DeleteBorrowerStake(ctx, borrowerAddr)
	} else {
		if err := k.SetBorrowerStake(ctx, stake); err != nil {
			return err
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBorrowerSlashed,
			sdk.NewAttribute(types.AttributeKeyBorrower, borrowerAddr),
			sdk.NewAttribute(types.AttributeKeyValidator, validatorAddr),
			sdk.NewAttribute(types.AttributeKeySlashAmount, slashAmount.String()),
			sdk.NewAttribute(types.AttributeKeySlashFraction, slashFraction.String()),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)

	k.Logger(ctx).Info("borrower stake slashed",
		"borrower", borrowerAddr,
		"validator", validatorAddr,
		"amount", slashAmount.String(),
		"reason", reason,
	)

	return nil
}
