package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// =============================================================================
// Treasury CRUD Operations
// =============================================================================

// SetCompanyTreasury stores a company treasury record
func (k Keeper) SetCompanyTreasury(ctx sdk.Context, treasury types.CompanyTreasury) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyTreasuryKey(treasury.CompanyID)
	bz, err := json.Marshal(treasury)
	if err != nil {
		return fmt.Errorf("failed to marshal company treasury: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// GetCompanyTreasury retrieves a company treasury by company ID
func (k Keeper) GetCompanyTreasury(ctx sdk.Context, companyID uint64) (types.CompanyTreasury, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyTreasuryKey(companyID)
	bz := store.Get(key)
	if bz == nil {
		return types.CompanyTreasury{}, false
	}

	var treasury types.CompanyTreasury
	if err := json.Unmarshal(bz, &treasury); err != nil {
		return types.CompanyTreasury{}, false
	}
	return treasury, true
}

// DeleteCompanyTreasury removes a company treasury record
func (k Keeper) DeleteCompanyTreasury(ctx sdk.Context, companyID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyTreasuryKey(companyID)
	store.Delete(key)
}

// GetAllCompanyTreasuries returns all company treasuries
func (k Keeper) GetAllCompanyTreasuries(ctx sdk.Context) []types.CompanyTreasury {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CompanyTreasuryPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	treasuries := []types.CompanyTreasury{}
	for ; iterator.Valid(); iterator.Next() {
		var treasury types.CompanyTreasury
		if err := json.Unmarshal(iterator.Value(), &treasury); err == nil {
			treasuries = append(treasuries, treasury)
		}
	}

	return treasuries
}

// =============================================================================
// Treasury Withdrawal CRUD
// =============================================================================

// GetNextWithdrawalID returns the next withdrawal ID and increments the counter
func (k Keeper) GetNextWithdrawalID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.TreasuryWithdrawalCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.TreasuryWithdrawalCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// SetTreasuryWithdrawal stores a treasury withdrawal record
func (k Keeper) SetTreasuryWithdrawal(ctx sdk.Context, withdrawal types.TreasuryWithdrawal) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTreasuryWithdrawalKey(withdrawal.ID)
	bz, err := json.Marshal(withdrawal)
	if err != nil {
		return fmt.Errorf("failed to marshal treasury withdrawal: %w", err)
	}
	store.Set(key, bz)

	// Index by company ID
	indexKey := types.GetWithdrawalByCompanyKey(withdrawal.CompanyID, withdrawal.ID)
	store.Set(indexKey, sdk.Uint64ToBigEndian(withdrawal.ID))

	// Index by proposal ID if set
	if withdrawal.ProposalID > 0 {
		proposalIndexKey := types.GetWithdrawalByProposalKey(withdrawal.ProposalID)
		store.Set(proposalIndexKey, sdk.Uint64ToBigEndian(withdrawal.ID))
	}
	return nil
}

// GetTreasuryWithdrawal retrieves a treasury withdrawal by ID
func (k Keeper) GetTreasuryWithdrawal(ctx sdk.Context, withdrawalID uint64) (types.TreasuryWithdrawal, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTreasuryWithdrawalKey(withdrawalID)
	bz := store.Get(key)
	if bz == nil {
		return types.TreasuryWithdrawal{}, false
	}

	var withdrawal types.TreasuryWithdrawal
	if err := json.Unmarshal(bz, &withdrawal); err != nil {
		return types.TreasuryWithdrawal{}, false
	}
	return withdrawal, true
}

// GetWithdrawalsByCompany returns all withdrawals for a company
func (k Keeper) GetWithdrawalsByCompany(ctx sdk.Context, companyID uint64) []types.TreasuryWithdrawal {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetWithdrawalsByCompanyPrefix(companyID)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	withdrawals := []types.TreasuryWithdrawal{}
	for ; iterator.Valid(); iterator.Next() {
		withdrawalID := sdk.BigEndianToUint64(iterator.Value())
		if withdrawal, found := k.GetTreasuryWithdrawal(ctx, withdrawalID); found {
			withdrawals = append(withdrawals, withdrawal)
		}
	}

	return withdrawals
}

// GetWithdrawalByProposal retrieves a withdrawal by governance proposal ID
func (k Keeper) GetWithdrawalByProposal(ctx sdk.Context, proposalID uint64) (types.TreasuryWithdrawal, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetWithdrawalByProposalKey(proposalID)
	bz := store.Get(key)
	if bz == nil {
		return types.TreasuryWithdrawal{}, false
	}

	withdrawalID := sdk.BigEndianToUint64(bz)
	return k.GetTreasuryWithdrawal(ctx, withdrawalID)
}

// =============================================================================
// Treasury Operations
// =============================================================================

// DepositToTreasury adds funds to company treasury (called after primary sales)
func (k Keeper) DepositToTreasury(ctx sdk.Context, companyID uint64, amount sdk.Coins, depositor string) error {
	// Validate inputs
	if !amount.IsValid() || amount.IsZero() {
		return types.ErrInvalidWithdrawalAmount
	}

	// Check if company exists
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Get or create treasury
	treasury, exists := k.GetCompanyTreasury(ctx, companyID)
	if !exists {
		treasury = types.NewCompanyTreasury(companyID)
	}

	// Check if treasury is frozen
	if k.IsTreasuryFrozen(ctx, companyID) {
		return types.ErrTreasuryFrozenForInvestigation
	}

	// Transfer funds from depositor to module account
	depositorAddr, err := sdk.AccAddressFromBech32(depositor)
	if err != nil {
		return types.ErrUnauthorized
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositorAddr, types.ModuleName, amount); err != nil {
		return types.ErrInsufficientFunds
	}

	// Update treasury balances
	treasury.Balance = treasury.Balance.Add(amount...)
	treasury.TotalDeposited = treasury.TotalDeposited.Add(amount...)
	treasury.UpdatedAt = ctx.BlockTime()

	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryDeposit,
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyDepositAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor),
			sdk.NewAttribute(types.AttributeKeyTreasuryBalance, treasury.Balance.String()),
		),
	)

	k.Logger(ctx).Info("treasury deposit",
		"company_id", companyID,
		"amount", amount.String(),
		"depositor", depositor,
		"new_balance", treasury.Balance.String(),
	)

	return nil
}

// RequestTreasuryWithdrawal creates a withdrawal request (requires governance proposal)
func (k Keeper) RequestTreasuryWithdrawal(
	ctx sdk.Context,
	companyID uint64,
	recipient string,
	amount sdk.Coins,
	purpose string,
	requestedBy string,
) (uint64, error) {
	// Validate inputs
	if !amount.IsValid() || amount.IsZero() {
		return 0, types.ErrInvalidWithdrawalAmount
	}
	if purpose == "" {
		return 0, types.ErrInvalidReason
	}

	// Check if requester is authorized
	if !k.CanRequestWithdrawal(ctx, companyID, requestedBy) {
		return 0, types.ErrNotTreasuryController
	}

	// Get treasury
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return 0, types.ErrTreasuryNotFound
	}

	// Check if treasury is frozen
	if k.IsTreasuryFrozen(ctx, companyID) {
		return 0, types.ErrTreasuryFrozenForInvestigation
	}

	// Check if there's already a pending withdrawal
	if k.HasPendingWithdrawal(ctx, companyID) {
		return 0, types.ErrWithdrawalPending
	}

	// Check if sufficient available balance
	if !treasury.CanWithdraw(amount) {
		return 0, types.ErrInsufficientTreasuryBalance
	}

	// Lock the amount in treasury
	if err := k.LockTreasuryAmount(ctx, companyID, amount); err != nil {
		return 0, err
	}

	// Create withdrawal request
	withdrawalID := k.GetNextWithdrawalID(ctx)
	withdrawal := types.NewTreasuryWithdrawal(
		withdrawalID,
		companyID,
		recipient,
		amount,
		purpose,
		requestedBy,
	)

	if err := k.SetTreasuryWithdrawal(ctx, withdrawal); err != nil {
		return 0, err
	}

	company, _ := k.getCompany(ctx, companyID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryWithdrawalRequest,
			sdk.NewAttribute(types.AttributeKeyWithdrawalID, fmt.Sprintf("%d", withdrawalID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyWithdrawalRecipient, recipient),
			sdk.NewAttribute(types.AttributeKeyWithdrawalAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyWithdrawalPurpose, purpose),
			sdk.NewAttribute(types.AttributeKeyWithdrawalProposer, requestedBy),
		),
	)

	k.Logger(ctx).Info("treasury withdrawal requested",
		"withdrawal_id", withdrawalID,
		"company_id", companyID,
		"amount", amount.String(),
		"recipient", recipient,
		"purpose", purpose,
	)

	return withdrawalID, nil
}

// ApproveTreasuryWithdrawal marks withdrawal as approved (called by governance)
func (k Keeper) ApproveTreasuryWithdrawal(ctx sdk.Context, withdrawalID uint64, proposalID uint64) error {
	withdrawal, found := k.GetTreasuryWithdrawal(ctx, withdrawalID)
	if !found {
		return types.ErrWithdrawalNotFound
	}

	if withdrawal.Status != types.WithdrawalStatusPending {
		return types.ErrClaimAlreadyProcessed
	}

	// Update withdrawal status
	withdrawal.Status = types.WithdrawalStatusApproved
	withdrawal.ProposalID = proposalID
	if err := k.SetTreasuryWithdrawal(ctx, withdrawal); err != nil {
		return err
	}

	company, _ := k.getCompany(ctx, withdrawal.CompanyID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryWithdrawalApproved,
			sdk.NewAttribute(types.AttributeKeyWithdrawalID, fmt.Sprintf("%d", withdrawalID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", withdrawal.CompanyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyWithdrawalAmount, withdrawal.Amount.String()),
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
		),
	)

	k.Logger(ctx).Info("treasury withdrawal approved",
		"withdrawal_id", withdrawalID,
		"company_id", withdrawal.CompanyID,
		"proposal_id", proposalID,
	)

	return nil
}

// ExecuteTreasuryWithdrawal transfers funds (called after timelock)
func (k Keeper) ExecuteTreasuryWithdrawal(ctx sdk.Context, withdrawalID uint64) error {
	withdrawal, found := k.GetTreasuryWithdrawal(ctx, withdrawalID)
	if !found {
		return types.ErrWithdrawalNotFound
	}

	if withdrawal.Status != types.WithdrawalStatusApproved {
		return types.ErrWithdrawalNotApproved
	}

	if !withdrawal.ExecutedAt.IsZero() {
		return types.ErrWithdrawalAlreadyExecuted
	}

	// Get treasury
	treasury, found := k.GetCompanyTreasury(ctx, withdrawal.CompanyID)
	if !found {
		return types.ErrTreasuryNotFound
	}

	// Check if treasury is frozen (might have been frozen after approval)
	if k.IsTreasuryFrozen(ctx, withdrawal.CompanyID) {
		return types.ErrTreasuryFrozenForInvestigation
	}

	// Transfer funds from module to recipient
	recipientAddr, err := sdk.AccAddressFromBech32(withdrawal.Recipient)
	if err != nil {
		return types.ErrUnauthorized
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, withdrawal.Amount); err != nil {
		return types.ErrInsufficientTreasuryBalance
	}

	// Update treasury balances
	treasury.Balance = treasury.Balance.Sub(withdrawal.Amount...)
	treasury.TotalWithdrawn = treasury.TotalWithdrawn.Add(withdrawal.Amount...)
	treasury.LockedAmount = treasury.LockedAmount.Sub(withdrawal.Amount...)
	treasury.UpdatedAt = ctx.BlockTime()
	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	// Update withdrawal status
	withdrawal.Status = types.WithdrawalStatusExecuted
	withdrawal.ExecutedAt = ctx.BlockTime()
	if err := k.SetTreasuryWithdrawal(ctx, withdrawal); err != nil {
		return err
	}

	company, _ := k.getCompany(ctx, withdrawal.CompanyID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryWithdrawalExecuted,
			sdk.NewAttribute(types.AttributeKeyWithdrawalID, fmt.Sprintf("%d", withdrawalID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", withdrawal.CompanyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyWithdrawalRecipient, withdrawal.Recipient),
			sdk.NewAttribute(types.AttributeKeyWithdrawalAmount, withdrawal.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyTreasuryBalance, treasury.Balance.String()),
		),
	)

	k.Logger(ctx).Info("treasury withdrawal executed",
		"withdrawal_id", withdrawalID,
		"company_id", withdrawal.CompanyID,
		"recipient", withdrawal.Recipient,
		"amount", withdrawal.Amount.String(),
	)

	return nil
}

// CancelTreasuryWithdrawal cancels a pending withdrawal
func (k Keeper) CancelTreasuryWithdrawal(ctx sdk.Context, withdrawalID uint64, reason string) error {
	withdrawal, found := k.GetTreasuryWithdrawal(ctx, withdrawalID)
	if !found {
		return types.ErrWithdrawalNotFound
	}

	if withdrawal.Status != types.WithdrawalStatusPending {
		return types.ErrClaimAlreadyProcessed
	}

	// Unlock the amount in treasury
	if err := k.UnlockTreasuryAmount(ctx, withdrawal.CompanyID, withdrawal.Amount); err != nil {
		return err
	}

	// Update withdrawal status
	withdrawal.Status = types.WithdrawalStatusCancelled
	if err := k.SetTreasuryWithdrawal(ctx, withdrawal); err != nil {
		return err
	}

	company, _ := k.getCompany(ctx, withdrawal.CompanyID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryWithdrawalCancelled,
			sdk.NewAttribute(types.AttributeKeyWithdrawalID, fmt.Sprintf("%d", withdrawalID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", withdrawal.CompanyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute("reason", reason),
		),
	)

	k.Logger(ctx).Info("treasury withdrawal cancelled",
		"withdrawal_id", withdrawalID,
		"company_id", withdrawal.CompanyID,
		"reason", reason,
	)

	return nil
}

// RejectTreasuryWithdrawal marks withdrawal as rejected (called by governance)
func (k Keeper) RejectTreasuryWithdrawal(ctx sdk.Context, withdrawalID uint64) error {
	withdrawal, found := k.GetTreasuryWithdrawal(ctx, withdrawalID)
	if !found {
		return types.ErrWithdrawalNotFound
	}

	if withdrawal.Status != types.WithdrawalStatusPending {
		return types.ErrClaimAlreadyProcessed
	}

	// Unlock the amount in treasury
	if err := k.UnlockTreasuryAmount(ctx, withdrawal.CompanyID, withdrawal.Amount); err != nil {
		return err
	}

	// Update withdrawal status
	withdrawal.Status = types.WithdrawalStatusRejected
	if err := k.SetTreasuryWithdrawal(ctx, withdrawal); err != nil {
		return err
	}

	company, _ := k.getCompany(ctx, withdrawal.CompanyID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryWithdrawalRejected,
			sdk.NewAttribute(types.AttributeKeyWithdrawalID, fmt.Sprintf("%d", withdrawalID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", withdrawal.CompanyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyWithdrawalStatus, "rejected"),
		),
	)

	k.Logger(ctx).Info("treasury withdrawal rejected",
		"withdrawal_id", withdrawalID,
		"company_id", withdrawal.CompanyID,
	)

	return nil
}

// =============================================================================
// Treasury Locking (for proposals/dividends)
// =============================================================================

// LockTreasuryAmount locks funds for pending proposal
func (k Keeper) LockTreasuryAmount(ctx sdk.Context, companyID uint64, amount sdk.Coins) error {
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return types.ErrTreasuryNotFound
	}

	// Check if sufficient available balance
	available := treasury.GetAvailableBalance()
	if !available.IsAllGTE(amount) {
		return types.ErrInsufficientTreasuryBalance
	}

	// Lock the amount
	treasury.LockedAmount = treasury.LockedAmount.Add(amount...)
	treasury.UpdatedAt = ctx.BlockTime()
	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryLocked,
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyTreasuryBalance, treasury.LockedAmount.String()),
		),
	)

	k.Logger(ctx).Info("treasury amount locked",
		"company_id", companyID,
		"amount", amount.String(),
		"total_locked", treasury.LockedAmount.String(),
	)

	return nil
}

// UnlockTreasuryAmount releases locked funds
func (k Keeper) UnlockTreasuryAmount(ctx sdk.Context, companyID uint64, amount sdk.Coins) error {
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return types.ErrTreasuryNotFound
	}

	// Unlock the amount
	treasury.LockedAmount = treasury.LockedAmount.Sub(amount...)
	if treasury.LockedAmount.IsAnyNegative() {
		treasury.LockedAmount = sdk.NewCoins() // Reset to zero if negative
	}
	treasury.UpdatedAt = ctx.BlockTime()
	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryUnlocked,
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyTreasuryBalance, treasury.LockedAmount.String()),
		),
	)

	k.Logger(ctx).Info("treasury amount unlocked",
		"company_id", companyID,
		"amount", amount.String(),
		"remaining_locked", treasury.LockedAmount.String(),
	)

	return nil
}

// IsTreasuryLocked checks if treasury is locked (emergency flag)
func (k Keeper) IsTreasuryLocked(ctx sdk.Context, companyID uint64) bool {
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return false
	}
	return treasury.IsLocked
}

// =============================================================================
// Security Checks
// =============================================================================

// CanRequestWithdrawal checks if user can request withdrawal (must be owner or authorized proposer)
func (k Keeper) CanRequestWithdrawal(ctx sdk.Context, companyID uint64, requestor string) bool {
	// Use the new authorization system
	return k.CanProposeForCompany(ctx, companyID, requestor)
}

// HasPendingWithdrawal checks if company has pending withdrawal
func (k Keeper) HasPendingWithdrawal(ctx sdk.Context, companyID uint64) bool {
	withdrawals := k.GetWithdrawalsByCompany(ctx, companyID)
	for _, withdrawal := range withdrawals {
		if withdrawal.Status == types.WithdrawalStatusPending {
			return true
		}
	}
	return false
}

// IsTreasuryFrozen checks if treasury is frozen for investigation
func (k Keeper) IsTreasuryFrozen(ctx sdk.Context, companyID uint64) bool {
	_, exists := k.GetFrozenTreasury(ctx, companyID)
	return exists
}

// =============================================================================
// Treasury Share Management
// =============================================================================

// AddSharesToTreasury adds shares to treasury holdings (called during company creation or dilution)
func (k Keeper) AddSharesToTreasury(ctx sdk.Context, companyID uint64, classID string, shares math.Int) error {
	if shares.IsNil() || shares.LTE(math.ZeroInt()) {
		return types.ErrInsufficientShares
	}

	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		treasury = types.NewCompanyTreasury(companyID)
	}

	// Initialize maps if nil
	if treasury.TreasuryShares == nil {
		treasury.TreasuryShares = make(map[string]math.Int)
	}
	if treasury.SharesSoldTotal == nil {
		treasury.SharesSoldTotal = make(map[string]math.Int)
	}

	// Add shares to treasury holdings
	currentShares := treasury.GetTreasuryShares(classID)
	treasury.TreasuryShares[classID] = currentShares.Add(shares)
	treasury.UpdatedAt = ctx.BlockTime()

	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	k.Logger(ctx).Info("shares added to treasury",
		"company_id", companyID,
		"class_id", classID,
		"shares_added", shares.String(),
		"total_treasury_shares", treasury.TreasuryShares[classID].String(),
	)

	return nil
}

// TransferSharesFromTreasury transfers shares from treasury to a buyer
// This is called when treasury shares are sold (DEX or primary sale)
// Payment goes to treasury balance, shares go to buyer
func (k Keeper) TransferSharesFromTreasury(
	ctx sdk.Context,
	companyID uint64,
	classID string,
	buyer string,
	shares math.Int,
	pricePerShare math.LegacyDec,
	paymentDenom string,
) error {
	// Validate inputs
	if shares.IsNil() || shares.LTE(math.ZeroInt()) {
		return types.ErrInsufficientShares
	}

	// Get treasury
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return types.ErrTreasuryNotFound
	}

	// Check if treasury has enough shares
	if !treasury.CanSellShares(classID, shares) {
		return types.ErrInsufficientTreasuryShares
	}

	// Calculate total cost
	totalCost := pricePerShare.MulInt(shares)
	payment := sdk.NewCoins(sdk.NewCoin(paymentDenom, totalCost.TruncateInt()))

	// Transfer payment from buyer to module account (treasury holds it)
	buyerAddr, err := sdk.AccAddressFromBech32(buyer)
	if err != nil {
		return types.ErrUnauthorized
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, buyerAddr, types.ModuleName, payment); err != nil {
		return types.ErrInsufficientFunds
	}

	// Update treasury: reduce shares, increase balance
	treasury.TreasuryShares[classID] = treasury.TreasuryShares[classID].Sub(shares)

	// Track total sold
	soldTotal := treasury.GetSharesSold(classID)
	treasury.SharesSoldTotal[classID] = soldTotal.Add(shares)

	// Add payment to treasury balance
	treasury.Balance = treasury.Balance.Add(payment...)
	treasury.TotalDeposited = treasury.TotalDeposited.Add(payment...)
	treasury.UpdatedAt = ctx.BlockTime()

	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	// Issue shares to buyer using internal function
	// This updates ShareClass and creates/updates Shareholding for buyer
	if err := k.IssueShares(ctx, companyID, classID, buyer, shares, pricePerShare, paymentDenom); err != nil {
		// Rollback: refund buyer
		k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, buyerAddr, payment)
		return err
	}

	// Emit event
	company, _ := k.getCompany(ctx, companyID)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_share_sale",
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyShareClass, classID),
			sdk.NewAttribute("shares_sold", shares.String()),
			sdk.NewAttribute("buyer", buyer),
			sdk.NewAttribute("price_per_share", pricePerShare.String()),
			sdk.NewAttribute("total_cost", totalCost.String()),
			sdk.NewAttribute("payment_denom", paymentDenom),
			sdk.NewAttribute("treasury_balance", treasury.Balance.String()),
			sdk.NewAttribute("treasury_shares_remaining", treasury.TreasuryShares[classID].String()),
		),
	)

	k.Logger(ctx).Info("treasury shares sold",
		"company_id", companyID,
		"class_id", classID,
		"buyer", buyer,
		"shares", shares.String(),
		"price_per_share", pricePerShare.String(),
		"treasury_balance", treasury.Balance.String(),
	)

	return nil
}

// GetTreasuryShareBalance returns the number of shares held in treasury for a class
func (k Keeper) GetTreasuryShareBalance(ctx sdk.Context, companyID uint64, classID string) math.Int {
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return math.ZeroInt()
	}
	return treasury.GetTreasuryShares(classID)
}

// =============================================================================
// Direct Treasury Deposits (for dividend funding)
// =============================================================================

// DepositToTreasuryDirect allows founder/co-founders to deposit HODL into treasury
// This is used to fund dividends, operational expenses, etc.
func (k Keeper) DepositToTreasuryDirect(ctx sdk.Context, companyID uint64, depositor string, amount sdk.Coins) error {
	// Check depositor is authorized (owner or proposer)
	if !k.CanProposeForCompany(ctx, companyID, depositor) {
		return types.ErrNotAuthorizedProposer
	}

	// Get treasury
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return types.ErrTreasuryNotFound
	}

	// Check treasury isn't frozen
	if k.IsTreasuryFrozen(ctx, companyID) {
		return types.ErrTreasuryFrozenForInvestigation
	}

	// Validate deposit amount
	if !amount.IsValid() || amount.IsZero() {
		return types.ErrInvalidWithdrawalAmount
	}

	// Transfer from depositor to module account
	depositorAddr, err := sdk.AccAddressFromBech32(depositor)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositorAddr, types.ModuleName, amount); err != nil {
		return fmt.Errorf("failed to deposit to treasury: %w", err)
	}

	// Update treasury balance
	treasury.Balance = treasury.Balance.Add(amount...)
	treasury.TotalDeposited = treasury.TotalDeposited.Add(amount...)
	treasury.UpdatedAt = ctx.BlockTime()
	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	// Get company for event
	company, _ := k.getCompany(ctx, companyID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryDeposit,
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor),
			sdk.NewAttribute(types.AttributeKeyDepositAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyTreasuryBalance, treasury.Balance.String()),
		),
	)

	k.Logger(ctx).Info("direct treasury deposit",
		"company_id", companyID,
		"depositor", depositor,
		"amount", amount.String(),
		"new_balance", treasury.Balance.String(),
	)

	return nil
}

// =============================================================================
// Treasury Equity Investment Operations
// =============================================================================

// DepositEquityToTreasury allows co-founders to deposit ANY company's equity into treasury
// This transfers shares from the depositor to the treasury as an investment
func (k Keeper) DepositEquityToTreasury(
	ctx sdk.Context,
	companyID uint64,       // Treasury owner company
	depositor string,       // Who is depositing
	targetCompanyID uint64, // Company whose shares are being deposited
	classID string,         // Share class
	shares math.Int,        // Number of shares
) error {
	// Validate inputs
	if shares.IsNil() || shares.LTE(math.ZeroInt()) {
		return types.ErrInsufficientShares
	}

	// Check depositor is authorized (co-founder/proposer)
	if !k.CanProposeForCompany(ctx, companyID, depositor) {
		return types.ErrNotAuthorizedProposer
	}

	// Get treasury
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		treasury = types.NewCompanyTreasury(companyID)
	}

	// Check treasury isn't frozen
	if k.IsTreasuryFrozen(ctx, companyID) {
		return types.ErrTreasuryFrozenForInvestigation
	}

	// Verify target company exists
	targetCompany, found := k.getCompany(ctx, targetCompanyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// SECURITY: Prevent circular dividends - company cannot deposit its own shares into its treasury
	// Self-held shares should not receive dividends (standard corporate practice)
	if companyID == targetCompanyID {
		return types.ErrInvalidInput.Wrap("company cannot deposit its own shares into treasury (prevents circular dividends)")
	}

	// Verify depositor owns the shares (use internal lowercase version that returns typed result)
	shareholding, found := k.getShareholding(ctx, targetCompanyID, classID, depositor)
	if !found || shareholding.Shares.LT(shares) {
		return types.ErrInsufficientShares
	}

	// Get treasury address (we'll use the module account as holder, but track beneficial ownership)
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	// Transfer shares from depositor to module account
	if err := k.TransferShares(ctx, targetCompanyID, classID, depositor, moduleAddr.String(), shares); err != nil {
		return fmt.Errorf("failed to transfer shares to treasury: %w", err)
	}

	// Record beneficial ownership - the company treasury is the beneficial owner
	// Use the treasury's company address as the beneficial owner for dividend routing
	ownerCompany, _ := k.getCompany(ctx, companyID)
	treasuryBeneficialOwner := fmt.Sprintf("treasury:%d:%s", companyID, ownerCompany.Symbol)

	if err := k.RegisterBeneficialOwner(
		ctx,
		moduleAddr.String(),       // moduleAccount
		targetCompanyID,           // companyID (whose shares)
		classID,                   // classID
		treasuryBeneficialOwner,   // beneficialOwner (treasury identifier)
		shares,                    // shares
		companyID,                 // refID (owner company ID)
		"treasury_investment",     // refType
	); err != nil {
		k.Logger(ctx).Error("failed to register treasury beneficial ownership", "error", err)
		// Continue anyway - the shares are held, just beneficial owner tracking failed
	}

	// Update treasury investment holdings
	if treasury.InvestmentHoldings == nil {
		treasury.InvestmentHoldings = make(map[uint64]map[string]math.Int)
	}
	if treasury.InvestmentHoldings[targetCompanyID] == nil {
		treasury.InvestmentHoldings[targetCompanyID] = make(map[string]math.Int)
	}

	currentShares := treasury.GetInvestmentShares(targetCompanyID, classID)
	treasury.InvestmentHoldings[targetCompanyID][classID] = currentShares.Add(shares)
	treasury.UpdatedAt = ctx.BlockTime()

	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	// Emit event (ownerCompany already retrieved above)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_equity_deposit",
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, ownerCompany.Symbol),
			sdk.NewAttribute("target_company_id", fmt.Sprintf("%d", targetCompanyID)),
			sdk.NewAttribute("target_company_symbol", targetCompany.Symbol),
			sdk.NewAttribute(types.AttributeKeyShareClass, classID),
			sdk.NewAttribute("shares_deposited", shares.String()),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor),
			sdk.NewAttribute("total_investment_shares", treasury.InvestmentHoldings[targetCompanyID][classID].String()),
		),
	)

	k.Logger(ctx).Info("equity deposited to treasury",
		"company_id", companyID,
		"target_company_id", targetCompanyID,
		"class_id", classID,
		"shares", shares.String(),
		"depositor", depositor,
	)

	return nil
}

// =============================================================================
// Equity Withdrawal Request CRUD
// =============================================================================

// GetNextEquityWithdrawalID returns the next equity withdrawal ID
func (k Keeper) GetNextEquityWithdrawalID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.EquityWithdrawalCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.EquityWithdrawalCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// SetEquityWithdrawalRequest stores an equity withdrawal request
func (k Keeper) SetEquityWithdrawalRequest(ctx sdk.Context, request types.EquityWithdrawalRequest) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEquityWithdrawalKey(request.ID)
	bz, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal equity withdrawal request: %w", err)
	}
	store.Set(key, bz)

	// Index by company ID
	indexKey := types.GetEquityWithdrawalByCompanyKey(request.CompanyID, request.ID)
	store.Set(indexKey, sdk.Uint64ToBigEndian(request.ID))

	// Index by proposal ID if set
	if request.ProposalID > 0 {
		proposalIndexKey := types.GetEquityWithdrawalByProposalKey(request.ProposalID)
		store.Set(proposalIndexKey, sdk.Uint64ToBigEndian(request.ID))
	}
	return nil
}

// GetEquityWithdrawalRequest retrieves an equity withdrawal request by ID
func (k Keeper) GetEquityWithdrawalRequest(ctx sdk.Context, requestID uint64) (types.EquityWithdrawalRequest, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetEquityWithdrawalKey(requestID)
	bz := store.Get(key)
	if bz == nil {
		return types.EquityWithdrawalRequest{}, false
	}

	var request types.EquityWithdrawalRequest
	if err := json.Unmarshal(bz, &request); err != nil {
		return types.EquityWithdrawalRequest{}, false
	}
	return request, true
}

// GetEquityWithdrawalsByCompany returns all equity withdrawals for a company
func (k Keeper) GetEquityWithdrawalsByCompany(ctx sdk.Context, companyID uint64) []types.EquityWithdrawalRequest {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetEquityWithdrawalsByCompanyPrefix(companyID)
	iterator := storetypes.KVStorePrefixIterator(store, prefixKey)
	defer iterator.Close()

	requests := []types.EquityWithdrawalRequest{}
	for ; iterator.Valid(); iterator.Next() {
		requestID := sdk.BigEndianToUint64(iterator.Value())
		if request, found := k.GetEquityWithdrawalRequest(ctx, requestID); found {
			requests = append(requests, request)
		}
	}

	return requests
}

// RequestEquityWithdrawal creates a withdrawal request (requires governance vote)
func (k Keeper) RequestEquityWithdrawal(
	ctx sdk.Context,
	companyID uint64,
	targetCompanyID uint64,
	classID string,
	shares math.Int,
	recipient string,
	purpose string,
	requestedBy string,
) (uint64, error) {
	// Validate inputs
	if shares.IsNil() || shares.LTE(math.ZeroInt()) {
		return 0, types.ErrInsufficientShares
	}
	if purpose == "" {
		return 0, types.ErrInvalidReason
	}

	// Check requester is authorized
	if !k.CanRequestWithdrawal(ctx, companyID, requestedBy) {
		return 0, types.ErrNotTreasuryController
	}

	// Get treasury
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return 0, types.ErrTreasuryNotFound
	}

	// Check treasury isn't frozen
	if k.IsTreasuryFrozen(ctx, companyID) {
		return 0, types.ErrTreasuryFrozenForInvestigation
	}

	// Check if there are enough available shares (not locked)
	available := treasury.GetAvailableInvestmentShares(targetCompanyID, classID)
	if available.LT(shares) {
		return 0, types.ErrInsufficientTreasuryShares
	}

	// Lock the shares in treasury
	if err := k.LockTreasuryEquity(ctx, companyID, targetCompanyID, classID, shares); err != nil {
		return 0, err
	}

	// Create withdrawal request
	requestID := k.GetNextEquityWithdrawalID(ctx)
	request := types.NewEquityWithdrawalRequest(
		requestID,
		companyID,
		targetCompanyID,
		classID,
		shares,
		recipient,
		purpose,
		requestedBy,
	)

	if err := k.SetEquityWithdrawalRequest(ctx, request); err != nil {
		return 0, err
	}

	ownerCompany, _ := k.getCompany(ctx, companyID)
	targetCompany, _ := k.getCompany(ctx, targetCompanyID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_equity_withdrawal_request",
			sdk.NewAttribute("request_id", fmt.Sprintf("%d", requestID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, ownerCompany.Symbol),
			sdk.NewAttribute("target_company_id", fmt.Sprintf("%d", targetCompanyID)),
			sdk.NewAttribute("target_company_symbol", targetCompany.Symbol),
			sdk.NewAttribute(types.AttributeKeyShareClass, classID),
			sdk.NewAttribute("shares", shares.String()),
			sdk.NewAttribute(types.AttributeKeyWithdrawalRecipient, recipient),
			sdk.NewAttribute(types.AttributeKeyWithdrawalPurpose, purpose),
			sdk.NewAttribute(types.AttributeKeyWithdrawalProposer, requestedBy),
		),
	)

	k.Logger(ctx).Info("equity withdrawal requested",
		"request_id", requestID,
		"company_id", companyID,
		"target_company_id", targetCompanyID,
		"shares", shares.String(),
		"recipient", recipient,
	)

	return requestID, nil
}

// ApproveEquityWithdrawal marks withdrawal as approved (called by governance)
func (k Keeper) ApproveEquityWithdrawal(ctx sdk.Context, requestID uint64, proposalID uint64) error {
	request, found := k.GetEquityWithdrawalRequest(ctx, requestID)
	if !found {
		return types.ErrWithdrawalNotFound
	}

	if request.Status != types.WithdrawalStatusPending {
		return types.ErrClaimAlreadyProcessed
	}

	request.Status = types.WithdrawalStatusApproved
	request.ProposalID = proposalID
	if err := k.SetEquityWithdrawalRequest(ctx, request); err != nil {
		return err
	}

	k.Logger(ctx).Info("equity withdrawal approved",
		"request_id", requestID,
		"company_id", request.CompanyID,
		"proposal_id", proposalID,
	)

	return nil
}

// ExecuteEquityWithdrawal transfers shares (called after timelock)
func (k Keeper) ExecuteEquityWithdrawal(ctx sdk.Context, requestID uint64) error {
	request, found := k.GetEquityWithdrawalRequest(ctx, requestID)
	if !found {
		return types.ErrWithdrawalNotFound
	}

	if request.Status != types.WithdrawalStatusApproved {
		return types.ErrWithdrawalNotApproved
	}

	if !request.ExecutedAt.IsZero() {
		return types.ErrWithdrawalAlreadyExecuted
	}

	// Get treasury
	treasury, found := k.GetCompanyTreasury(ctx, request.CompanyID)
	if !found {
		return types.ErrTreasuryNotFound
	}

	// Check treasury isn't frozen
	if k.IsTreasuryFrozen(ctx, request.CompanyID) {
		return types.ErrTreasuryFrozenForInvestigation
	}

	// Get module address
	moduleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)

	// Transfer shares from module to recipient
	if err := k.TransferShares(ctx, request.TargetCompanyID, request.ClassID, moduleAddr.String(), request.Recipient, request.Shares); err != nil {
		return fmt.Errorf("failed to transfer shares: %w", err)
	}

	// Update treasury: remove from investment holdings and locked
	if treasury.InvestmentHoldings != nil && treasury.InvestmentHoldings[request.TargetCompanyID] != nil {
		current := treasury.GetInvestmentShares(request.TargetCompanyID, request.ClassID)
		newAmount := current.Sub(request.Shares)
		if newAmount.IsNegative() {
			newAmount = math.ZeroInt()
		}
		treasury.InvestmentHoldings[request.TargetCompanyID][request.ClassID] = newAmount
	}

	// Unlock the equity
	if err := k.UnlockTreasuryEquity(ctx, request.CompanyID, request.TargetCompanyID, request.ClassID, request.Shares); err != nil {
		k.Logger(ctx).Error("failed to unlock treasury equity", "error", err)
	}

	treasury.UpdatedAt = ctx.BlockTime()
	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	// Update request status
	request.Status = types.WithdrawalStatusExecuted
	request.ExecutedAt = ctx.BlockTime()
	if err := k.SetEquityWithdrawalRequest(ctx, request); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_equity_withdrawal_executed",
			sdk.NewAttribute("request_id", fmt.Sprintf("%d", requestID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", request.CompanyID)),
			sdk.NewAttribute("target_company_id", fmt.Sprintf("%d", request.TargetCompanyID)),
			sdk.NewAttribute(types.AttributeKeyShareClass, request.ClassID),
			sdk.NewAttribute("shares", request.Shares.String()),
			sdk.NewAttribute(types.AttributeKeyWithdrawalRecipient, request.Recipient),
		),
	)

	k.Logger(ctx).Info("equity withdrawal executed",
		"request_id", requestID,
		"company_id", request.CompanyID,
		"target_company_id", request.TargetCompanyID,
		"shares", request.Shares.String(),
		"recipient", request.Recipient,
	)

	return nil
}

// RejectEquityWithdrawal marks withdrawal as rejected and unlocks shares
func (k Keeper) RejectEquityWithdrawal(ctx sdk.Context, requestID uint64) error {
	request, found := k.GetEquityWithdrawalRequest(ctx, requestID)
	if !found {
		return types.ErrWithdrawalNotFound
	}

	if request.Status != types.WithdrawalStatusPending {
		return types.ErrClaimAlreadyProcessed
	}

	// Unlock the shares
	if err := k.UnlockTreasuryEquity(ctx, request.CompanyID, request.TargetCompanyID, request.ClassID, request.Shares); err != nil {
		return err
	}

	request.Status = types.WithdrawalStatusRejected
	if err := k.SetEquityWithdrawalRequest(ctx, request); err != nil {
		return err
	}

	k.Logger(ctx).Info("equity withdrawal rejected",
		"request_id", requestID,
		"company_id", request.CompanyID,
	)

	return nil
}

// =============================================================================
// Treasury Equity Locking (for pending proposals)
// =============================================================================

// LockTreasuryEquity locks equity for pending withdrawal proposal
func (k Keeper) LockTreasuryEquity(ctx sdk.Context, companyID, targetCompanyID uint64, classID string, shares math.Int) error {
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return types.ErrTreasuryNotFound
	}

	// Check if there are enough available shares
	available := treasury.GetAvailableInvestmentShares(targetCompanyID, classID)
	if available.LT(shares) {
		return types.ErrInsufficientTreasuryShares
	}

	// Initialize locked equity map if needed
	if treasury.LockedEquity == nil {
		treasury.LockedEquity = make(map[uint64]map[string]math.Int)
	}
	if treasury.LockedEquity[targetCompanyID] == nil {
		treasury.LockedEquity[targetCompanyID] = make(map[string]math.Int)
	}

	// Lock the shares
	currentLocked := treasury.GetLockedEquity(targetCompanyID, classID)
	treasury.LockedEquity[targetCompanyID][classID] = currentLocked.Add(shares)
	treasury.UpdatedAt = ctx.BlockTime()

	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	k.Logger(ctx).Info("treasury equity locked",
		"company_id", companyID,
		"target_company_id", targetCompanyID,
		"class_id", classID,
		"shares_locked", shares.String(),
	)

	return nil
}

// UnlockTreasuryEquity releases locked equity
func (k Keeper) UnlockTreasuryEquity(ctx sdk.Context, companyID, targetCompanyID uint64, classID string, shares math.Int) error {
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return types.ErrTreasuryNotFound
	}

	if treasury.LockedEquity == nil || treasury.LockedEquity[targetCompanyID] == nil {
		return nil // Nothing locked
	}

	currentLocked := treasury.GetLockedEquity(targetCompanyID, classID)
	newLocked := currentLocked.Sub(shares)
	if newLocked.IsNegative() {
		newLocked = math.ZeroInt()
	}
	treasury.LockedEquity[targetCompanyID][classID] = newLocked
	treasury.UpdatedAt = ctx.BlockTime()

	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	k.Logger(ctx).Info("treasury equity unlocked",
		"company_id", companyID,
		"target_company_id", targetCompanyID,
		"class_id", classID,
		"shares_unlocked", shares.String(),
	)

	return nil
}

// GetTreasuryInvestmentHoldings returns all investment holdings for a treasury
func (k Keeper) GetTreasuryInvestmentHoldings(ctx sdk.Context, companyID uint64) []types.InvestmentHolding {
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return []types.InvestmentHolding{}
	}
	return treasury.GetAllInvestmentHoldings()
}

// =============================================================================
// Treasury Dividend Credit (for cross-company investments)
// =============================================================================

// CreditTreasuryDividend credits dividend payment to a company's treasury balance
// This is called when a company's treasury holds shares in another company that pays dividends
func (k Keeper) CreditTreasuryDividend(
	ctx sdk.Context,
	recipientCompanyID uint64, // Company whose treasury receives the dividend
	sourceCompanyID uint64,    // Company that paid the dividend
	amount sdk.Coins,          // Dividend amount (in HODL or other currency)
) error {
	// Get treasury
	treasury, found := k.GetCompanyTreasury(ctx, recipientCompanyID)
	if !found {
		return types.ErrTreasuryNotFound
	}

	// Credit the dividend to treasury balance
	treasury.Balance = treasury.Balance.Add(amount...)
	treasury.TotalDeposited = treasury.TotalDeposited.Add(amount...)
	treasury.UpdatedAt = ctx.BlockTime()

	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	// Get company info for event
	recipientCompany, _ := k.getCompany(ctx, recipientCompanyID)
	sourceCompany, _ := k.getCompany(ctx, sourceCompanyID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_dividend_received",
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", recipientCompanyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, recipientCompany.Symbol),
			sdk.NewAttribute("source_company_id", fmt.Sprintf("%d", sourceCompanyID)),
			sdk.NewAttribute("source_company_symbol", sourceCompany.Symbol),
			sdk.NewAttribute("dividend_amount", amount.String()),
			sdk.NewAttribute(types.AttributeKeyTreasuryBalance, treasury.Balance.String()),
		),
	)

	k.Logger(ctx).Info("treasury dividend credited",
		"recipient_company_id", recipientCompanyID,
		"source_company_id", sourceCompanyID,
		"amount", amount.String(),
		"new_balance", treasury.Balance.String(),
	)

	return nil
}
