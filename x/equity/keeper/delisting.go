package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// =============================================================================
// Delisting Compensation Storage
// =============================================================================

// GetNextCompensationID returns the next compensation ID and increments the counter
func (k Keeper) GetNextCompensationID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DelistingCompensationCounter)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.DelistingCompensationCounter, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// SetDelistingCompensation stores a delisting compensation record
func (k Keeper) SetDelistingCompensation(ctx sdk.Context, comp types.DelistingCompensation) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDelistingCompensationKey(comp.ID)
	bz, err := json.Marshal(comp)
	if err != nil {
		return fmt.Errorf("failed to marshal delisting compensation: %w", err)
	}
	store.Set(key, bz)

	// Index by company ID
	indexKey := types.GetCompensationByCompanyKey(comp.CompanyID)
	store.Set(indexKey, sdk.Uint64ToBigEndian(comp.ID))
	return nil
}

// GetDelistingCompensation retrieves a delisting compensation by ID
func (k Keeper) GetDelistingCompensation(ctx sdk.Context, compensationID uint64) (types.DelistingCompensation, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDelistingCompensationKey(compensationID)
	bz := store.Get(key)
	if bz == nil {
		return types.DelistingCompensation{}, false
	}

	var comp types.DelistingCompensation
	if err := json.Unmarshal(bz, &comp); err != nil {
		return types.DelistingCompensation{}, false
	}
	return comp, true
}

// GetCompensationByCompany retrieves compensation for a specific company
func (k Keeper) GetCompensationByCompany(ctx sdk.Context, companyID uint64) (types.DelistingCompensation, bool) {
	store := ctx.KVStore(k.storeKey)
	indexKey := types.GetCompensationByCompanyKey(companyID)
	bz := store.Get(indexKey)
	if bz == nil {
		return types.DelistingCompensation{}, false
	}

	compensationID := sdk.BigEndianToUint64(bz)
	return k.GetDelistingCompensation(ctx, compensationID)
}

// =============================================================================
// Compensation Claims Storage
// =============================================================================

// GetNextClaimID returns the next claim ID and increments the counter
func (k Keeper) GetNextClaimID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CompensationClaimCounter)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.CompensationClaimCounter, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// SetCompensationClaim stores a compensation claim
func (k Keeper) SetCompensationClaim(ctx sdk.Context, claim types.CompensationClaim) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompensationClaimKey(claim.ID)
	bz, err := json.Marshal(claim)
	if err != nil {
		return fmt.Errorf("failed to marshal compensation claim: %w", err)
	}
	store.Set(key, bz)

	// Index by compensation ID
	compIndexKey := types.GetClaimsByCompensationKey(claim.CompensationID, claim.ID)
	store.Set(compIndexKey, sdk.Uint64ToBigEndian(claim.ID))

	// Index by claimant
	claimantIndexKey := types.GetClaimsByClaimantKey(claim.Claimant, claim.ID)
	store.Set(claimantIndexKey, sdk.Uint64ToBigEndian(claim.ID))
	return nil
}

// GetCompensationClaim retrieves a claim by ID
func (k Keeper) GetCompensationClaim(ctx sdk.Context, claimID uint64) (types.CompensationClaim, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompensationClaimKey(claimID)
	bz := store.Get(key)
	if bz == nil {
		return types.CompensationClaim{}, false
	}

	var claim types.CompensationClaim
	if err := json.Unmarshal(bz, &claim); err != nil {
		return types.CompensationClaim{}, false
	}
	return claim, true
}

// =============================================================================
// Trading Halt Storage
// =============================================================================

// SetTradingHalt stores a trading halt
func (k Keeper) SetTradingHalt(ctx sdk.Context, halt types.TradingHalt) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradingHaltKey(halt.CompanyID)
	bz, err := json.Marshal(halt)
	if err != nil {
		return fmt.Errorf("failed to marshal trading halt: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// GetTradingHalt retrieves a trading halt for a company
func (k Keeper) GetTradingHalt(ctx sdk.Context, companyID uint64) (types.TradingHalt, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradingHaltKey(companyID)
	bz := store.Get(key)
	if bz == nil {
		return types.TradingHalt{}, false
	}

	var halt types.TradingHalt
	if err := json.Unmarshal(bz, &halt); err != nil {
		return types.TradingHalt{}, false
	}
	return halt, true
}

// DeleteTradingHalt removes a trading halt
func (k Keeper) DeleteTradingHalt(ctx sdk.Context, companyID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTradingHaltKey(companyID)
	store.Delete(key)
}

// IsTradingHalted checks if trading is halted for a company
func (k Keeper) IsTradingHalted(ctx sdk.Context, companyID uint64) bool {
	_, exists := k.GetTradingHalt(ctx, companyID)
	return exists
}

// =============================================================================
// Frozen Treasury Storage
// =============================================================================

// SetFrozenTreasury stores a frozen treasury record
func (k Keeper) SetFrozenTreasury(ctx sdk.Context, frozen types.FrozenTreasury) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetFrozenTreasuryKey(frozen.CompanyID)
	bz, err := json.Marshal(frozen)
	if err != nil {
		return fmt.Errorf("failed to marshal frozen treasury: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// GetFrozenTreasury retrieves a frozen treasury record
func (k Keeper) GetFrozenTreasury(ctx sdk.Context, companyID uint64) (types.FrozenTreasury, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetFrozenTreasuryKey(companyID)
	bz := store.Get(key)
	if bz == nil {
		return types.FrozenTreasury{}, false
	}

	var frozen types.FrozenTreasury
	if err := json.Unmarshal(bz, &frozen); err != nil {
		return types.FrozenTreasury{}, false
	}
	return frozen, true
}

// DeleteFrozenTreasury removes a frozen treasury record
func (k Keeper) DeleteFrozenTreasury(ctx sdk.Context, companyID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetFrozenTreasuryKey(companyID)
	store.Delete(key)
}

// =============================================================================
// Trading Halt Operations
// =============================================================================

// HaltTrading halts trading for a company (during fraud investigation)
func (k Keeper) HaltTrading(ctx sdk.Context, companyID uint64, reason string, initiator string, reportID uint64) error {
	// Check if company exists
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Check if already halted
	if k.IsTradingHalted(ctx, companyID) {
		return types.ErrTradingHalted
	}

	// Create trading halt (7-day investigation period)
	halt := types.NewTradingHalt(
		companyID,
		reason,
		initiator,
		reportID,
		7*24*time.Hour, // 7 days
		ctx.BlockTime(),
	)
	if err := k.SetTradingHalt(ctx, halt); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTradingHalted,
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyHaltReason, reason),
		),
	)

	k.Logger(ctx).Info("trading halted",
		"company_id", companyID,
		"symbol", company.Symbol,
		"reason", reason,
	)

	return nil
}

// ResumeTrading resumes trading for a company
func (k Keeper) ResumeTrading(ctx sdk.Context, companyID uint64, authority string) error {
	// Check if trading is halted
	if !k.IsTradingHalted(ctx, companyID) {
		return types.ErrTradingNotHalted
	}

	company, _ := k.getCompany(ctx, companyID)

	k.DeleteTradingHalt(ctx, companyID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTradingResumed,
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
		),
	)

	k.Logger(ctx).Info("trading resumed",
		"company_id", companyID,
		"symbol", company.Symbol,
	)

	return nil
}

// =============================================================================
// Treasury Freeze Operations
// =============================================================================

// FreezeTreasury freezes a company's treasury during investigation
func (k Keeper) FreezeTreasury(ctx sdk.Context, companyID uint64, amount math.Int, currency string, reason string, reportID uint64) error {
	// Check if already frozen
	if _, exists := k.GetFrozenTreasury(ctx, companyID); exists {
		return types.ErrTreasuryAlreadyFrozen
	}

	frozen := types.FrozenTreasury{
		CompanyID:    companyID,
		FrozenAmount: amount,
		Currency:     currency,
		FrozenAt:     ctx.BlockTime(),
		Reason:       reason,
		ReportID:     reportID,
	}
	if err := k.SetFrozenTreasury(ctx, frozen); err != nil {
		return err
	}

	company, _ := k.getCompany(ctx, companyID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryFrozen,
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyFrozenAmount, amount.String()),
		),
	)

	k.Logger(ctx).Info("treasury frozen",
		"company_id", companyID,
		"amount", amount.String(),
		"currency", currency,
	)

	return nil
}

// ReleaseTreasury releases frozen treasury funds
func (k Keeper) ReleaseTreasury(ctx sdk.Context, companyID uint64, destination string) (math.Int, error) {
	frozen, exists := k.GetFrozenTreasury(ctx, companyID)
	if !exists {
		return math.ZeroInt(), types.ErrTreasuryNotFrozen
	}

	// Update frozen record
	frozen.ReleasedAt = ctx.BlockTime()
	frozen.ReleasedTo = destination
	if err := k.SetFrozenTreasury(ctx, frozen); err != nil {
		return math.ZeroInt(), err
	}

	company, _ := k.getCompany(ctx, companyID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryReleased,
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyFrozenAmount, frozen.FrozenAmount.String()),
		),
	)

	k.Logger(ctx).Info("treasury released",
		"company_id", companyID,
		"amount", frozen.FrozenAmount.String(),
		"destination", destination,
	)

	return frozen.FrozenAmount, nil
}

// =============================================================================
// Fraudulent Company Delisting
// =============================================================================

// DelistFraudulentCompany handles the full delisting process for a confirmed fraud
func (k Keeper) DelistFraudulentCompany(ctx sdk.Context, companyID uint64, reason types.DelistingReason, reportID uint64, authority string) error {
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	if company.Status == types.CompanyStatusDelisted {
		return types.ErrCompanyAlreadyDelisted
	}

	// Step 1: Calculate total outstanding shares across all classes
	totalShares := k.calculateTotalOutstandingShares(ctx, companyID)

	// Step 2: Create compensation record if fraud-related
	var compensationID uint64
	if reason.RequiresCompensation() {
		compensationID = k.createCompensationRecord(ctx, company, reason, totalShares, reportID)
	}

	// Step 3: Delist the company (this unlocks founder stake)
	if err := k.DelistCompany(ctx, companyID, authority, reason.String()); err != nil {
		return err
	}

	// Step 4: Build compensation pool from available sources (if fraud)
	if reason.RequiresCompensation() {
		k.buildCompensationPool(ctx, companyID, compensationID)
	}

	// Step 5: Resume trading halt removal (if was halted)
	k.DeleteTradingHalt(ctx, companyID)

	// Emit delisting event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCompanyDelisted,
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyDelistingReason, reason.String()),
		),
	)

	k.Logger(ctx).Info("company delisted for fraud",
		"company_id", companyID,
		"symbol", company.Symbol,
		"reason", reason.String(),
		"compensation_id", compensationID,
	)

	return nil
}

// calculateTotalOutstandingShares calculates total shares across all classes
func (k Keeper) calculateTotalOutstandingShares(ctx sdk.Context, companyID uint64) math.Int {
	total := math.ZeroInt()
	classes := k.GetCompanyShareClasses(ctx, companyID)
	for _, class := range classes {
		total = total.Add(class.OutstandingShares)
	}
	return total
}

// createCompensationRecord creates a new compensation record for a delisted company
func (k Keeper) createCompensationRecord(ctx sdk.Context, company types.Company, reason types.DelistingReason, totalShares math.Int, reportID uint64) uint64 {
	compensationID := k.GetNextCompensationID(ctx)

	comp := types.NewDelistingCompensation(
		compensationID,
		company.ID,
		company.Symbol,
		reason,
		totalShares,
		ctx.BlockTime(),
	)
	comp.ReportID = reportID

	if err := k.SetDelistingCompensation(ctx, comp); err != nil {
		k.Logger(ctx).Error("failed to set delisting compensation", "error", err)
		return 0
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCompensationCreated,
			sdk.NewAttribute(types.AttributeKeyCompensationID, fmt.Sprintf("%d", compensationID)),
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", company.ID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
		),
	)

	return compensationID
}

// buildCompensationPool builds the compensation pool using the waterfall approach
func (k Keeper) buildCompensationPool(ctx sdk.Context, companyID uint64, compensationID uint64) {
	comp, found := k.GetDelistingCompensation(ctx, compensationID)
	if !found {
		return
	}

	// Source 1: Frozen company treasury (100%)
	if frozen, exists := k.GetFrozenTreasury(ctx, companyID); exists {
		comp.AddTreasuryContribution(frozen.FrozenAmount)
		k.Logger(ctx).Info("added treasury to compensation pool",
			"company_id", companyID,
			"amount", frozen.FrozenAmount.String(),
		)
	}

	// Source 2: Verifier stakes would be slashed here
	// This requires integration with staking module
	// For now, we log that this should happen
	k.Logger(ctx).Info("verifier stakes should be slashed for compensation",
		"company_id", companyID,
	)

	// Save updated compensation
	if err := k.SetDelistingCompensation(ctx, comp); err != nil {
		k.Logger(ctx).Error("failed to set delisting compensation", "error", err)
	}
}

// =============================================================================
// Shareholder Compensation Claims
// =============================================================================

// SubmitCompensationClaim allows a shareholder to claim compensation
func (k Keeper) SubmitCompensationClaim(ctx sdk.Context, companyID uint64, claimant string, shareClass string) (uint64, error) {
	// Get compensation record
	comp, found := k.GetCompensationByCompany(ctx, companyID)
	if !found {
		return 0, types.ErrCompensationNotFound
	}

	// Check if claimable
	if !comp.IsClaimable(ctx.BlockTime()) {
		if ctx.BlockTime().After(comp.ClaimDeadline) {
			return 0, types.ErrCompensationExpired
		}
		return 0, types.ErrCompensationNotClaimable
	}

	// Get shareholder's shares at time of delisting
	shareholding, found := k.getShareholding(ctx, companyID, shareClass, claimant)
	if !found || shareholding.Shares.IsZero() {
		return 0, types.ErrNoSharesHeld
	}

	// Check if already claimed (by checking existing claims)
	if k.hasExistingClaim(ctx, comp.ID, claimant) {
		return 0, types.ErrClaimAlreadySubmitted
	}

	// During claim window, just register the claim (no immediate payment)
	// Payment happens pro-rata after the claim window closes
	if ctx.BlockTime().Before(comp.ClaimWindowEnd) {
		claimID := k.GetNextClaimID(ctx)
		claim := types.NewCompensationClaim(
			claimID,
			comp.ID,
			companyID,
			claimant,
			shareClass,
			shareholding.Shares,
			math.LegacyZeroDec(), // Don't calculate amount yet - will be pro-rata
			ctx.BlockTime(),
		)
		claim.Status = "registered" // Not yet paid

		if err := k.SetCompensationClaim(ctx, claim); err != nil {
			return 0, err
		}

		// Update compensation record - track total shares claimed
		comp.TotalClaims++
		comp.TotalClaimedShares = comp.TotalClaimedShares.Add(shareholding.Shares)
		if err := k.SetDelistingCompensation(ctx, comp); err != nil {
			return 0, err
		}

		k.Logger(ctx).Info("compensation claim registered (payment pending pro-rata distribution)",
			"claim_id", claimID,
			"company_id", companyID,
			"claimant", claimant,
			"shares", shareholding.Shares.String(),
		)

		return claimID, nil
	}

	// If claim window has ended, don't accept new claims
	return 0, types.ErrCompensationExpired
}

// hasExistingClaim checks if claimant already has a claim for this compensation
func (k Keeper) hasExistingClaim(ctx sdk.Context, compensationID uint64, claimant string) bool {
	store := ctx.KVStore(k.storeKey)

	// Iterate all claims for this compensation
	prefix := types.GetClaimsByCompensationPrefix(compensationID)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		claimID := sdk.BigEndianToUint64(iterator.Value())
		if claim, found := k.GetCompensationClaim(ctx, claimID); found {
			if claim.Claimant == claimant {
				return true // Duplicate found
			}
		}
	}

	return false
}

// ProcessCompensationClaim processes a pending claim and pays the claimant
func (k Keeper) ProcessCompensationClaim(ctx sdk.Context, claimID uint64) error {
	claim, found := k.GetCompensationClaim(ctx, claimID)
	if !found {
		return types.ErrClaimNotFound
	}

	if claim.Status != "pending" {
		return types.ErrClaimAlreadyProcessed
	}

	// Get compensation record
	comp, found := k.GetDelistingCompensation(ctx, claim.CompensationID)
	if !found {
		return types.ErrCompensationNotFound
	}

	// Check if enough funds in pool
	if comp.RemainingPool.LT(claim.ClaimAmount) {
		return types.ErrInsufficientCompensationPool
	}

	// Transfer compensation to claimant
	claimantAddr, err := sdk.AccAddressFromBech32(claim.Claimant)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin("uhodl", claim.ClaimAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, claimantAddr, coins); err != nil {
		// If module doesn't have funds, try minting (as last resort)
		k.Logger(ctx).Warn("failed to send from module, attempting mint",
			"error", err.Error(),
		)
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
			return err
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, claimantAddr, coins); err != nil {
			return err
		}
	}

	// Update claim
	claim.Status = "paid"
	claim.ProcessedAt = ctx.BlockTime()
	if err := k.SetCompensationClaim(ctx, claim); err != nil {
		return err
	}

	// Update compensation
	comp.ProcessClaim(claim.ClaimAmount)
	if err := k.SetDelistingCompensation(ctx, comp); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCompensationClaimed,
			sdk.NewAttribute(types.AttributeKeyCompensationID, fmt.Sprintf("%d", comp.ID)),
			sdk.NewAttribute(types.AttributeKeyClaimant, claim.Claimant),
			sdk.NewAttribute(types.AttributeKeyClaimAmount, claim.ClaimAmount.String()),
			sdk.NewAttribute(types.AttributeKeySharesHeld, claim.SharesHeld.String()),
		),
	)

	k.Logger(ctx).Info("compensation paid",
		"claim_id", claimID,
		"claimant", claim.Claimant,
		"amount", claim.ClaimAmount.String(),
	)

	return nil
}

// =============================================================================
// Fraud Detection Integration
// =============================================================================

// InitiateFraudInvestigation is called when a fraud/scam report is submitted
// This halts trading immediately as a precaution during investigation
// NOTE: Treasury freeze requires a separate 24-hour warning via IssueFreezeWarning
func (k Keeper) InitiateFraudInvestigation(ctx sdk.Context, companyID uint64, reportID uint64, initiator string) error {
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Halt trading immediately as investigation precaution
	// This protects investors from trading a potentially fraudulent asset
	if err := k.HaltTrading(ctx, companyID, "fraud investigation", initiator, reportID); err != nil {
		// Already halted is not an error for this flow
		if err != types.ErrTradingHalted {
			return err
		}
	}

	// Suspend company status
	company.Status = types.CompanyStatusSuspended
	k.SetCompany(ctx, company)

	k.Logger(ctx).Info("fraud investigation initiated - trading halted",
		"company_id", companyID,
		"symbol", company.Symbol,
		"report_id", reportID,
		"note", "treasury NOT frozen yet - Stewards must issue 24h warning first",
	)

	// NOTE: Treasury freeze does NOT happen here automatically
	// Stewards must review the investigation and explicitly issue a freeze warning via:
	// keeper.IssueFreezeWarning(ctx, companyID, investigationID, reportID)
	// This gives the company 24 hours to respond with counter-evidence before freeze

	return nil
}

// ConfirmFraudAndDelist is called when fraud is confirmed - triggers full delisting
func (k Keeper) ConfirmFraudAndDelist(ctx sdk.Context, companyID uint64, reportID uint64, isFraud bool, authority string) error {
	if !isFraud {
		// Not fraud - resume trading and restore company
		if err := k.ResumeTrading(ctx, companyID, authority); err != nil {
			if err != types.ErrTradingNotHalted {
				return err
			}
		}

		company, found := k.getCompany(ctx, companyID)
		if found && company.Status == types.CompanyStatusSuspended {
			company.Status = types.CompanyStatusActive
			k.SetCompany(ctx, company)
		}

		k.Logger(ctx).Info("fraud not confirmed, company restored",
			"company_id", companyID,
		)

		return nil
	}

	// Fraud confirmed - full delisting
	return k.DelistFraudulentCompany(ctx, companyID, types.DelistingReasonFraud, reportID, authority)
}

// =============================================================================
// Pro-Rata Compensation Distribution
// =============================================================================

// ProcessCompensationDistribution handles pro-rata distribution after claim window closes
// Called from EndBlock
func (k Keeper) ProcessCompensationDistribution(ctx sdk.Context) {
	// Find compensations ready for distribution
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.DelistingCompensationPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var comp types.DelistingCompensation
		if err := json.Unmarshal(iterator.Value(), &comp); err != nil {
			k.Logger(ctx).Error("failed to unmarshal compensation", "error", err)
			continue
		}

		// Check if distribution time has passed and not yet distributed
		if ctx.BlockTime().After(comp.DistributionAt) && !comp.IsDistributed {
			if err := k.distributeCompensationProRata(ctx, &comp); err != nil {
				k.Logger(ctx).Error("failed to distribute compensation",
					"compensation_id", comp.ID,
					"error", err,
				)
			}
		}
	}
}

// distributeCompensationProRata distributes compensation to all claimants pro-rata
func (k Keeper) distributeCompensationProRata(ctx sdk.Context, comp *types.DelistingCompensation) error {
	if comp.IsDistributed {
		return fmt.Errorf("compensation already distributed")
	}

	// If no claims submitted, return
	if comp.TotalClaims == 0 || comp.TotalClaimedShares.IsZero() {
		k.Logger(ctx).Info("no claims submitted for compensation",
			"compensation_id", comp.ID,
			"company_id", comp.CompanyID,
		)
		comp.IsDistributed = true
		comp.Status = types.CompensationStatusCompleted
		if err := k.SetDelistingCompensation(ctx, *comp); err != nil {
			return err
		}
		return nil
	}

	// Calculate pro-rata compensation per share
	// claimAmount = (userShares / totalClaimedShares) * totalPool
	proRataPerShare := math.LegacyNewDecFromInt(comp.TotalPool).QuoInt(comp.TotalClaimedShares)

	// Get all claims for this compensation
	claims := k.getClaimsByCompensation(ctx, comp.ID)

	// Pay each claim proportionally
	for _, claim := range claims {
		if claim.Status != "registered" {
			continue // Skip already processed claims
		}

		// Calculate pro-rata amount
		claimAmount := proRataPerShare.MulInt(claim.SharesHeld).TruncateInt()

		// Transfer compensation to claimant
		claimantAddr, err := sdk.AccAddressFromBech32(claim.Claimant)
		if err != nil {
			k.Logger(ctx).Error("invalid claimant address",
				"claim_id", claim.ID,
				"claimant", claim.Claimant,
				"error", err,
			)
			continue
		}

		coins := sdk.NewCoins(sdk.NewCoin("uhodl", claimAmount))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, claimantAddr, coins); err != nil {
			// If module doesn't have funds, try minting (as last resort)
			k.Logger(ctx).Warn("failed to send from module, attempting mint",
				"claim_id", claim.ID,
				"error", err.Error(),
			)
			if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
				k.Logger(ctx).Error("failed to mint coins", "error", err)
				continue
			}
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, claimantAddr, coins); err != nil {
				k.Logger(ctx).Error("failed to send after mint", "error", err)
				continue
			}
		}

		// Update claim
		claim.ClaimAmount = claimAmount
		claim.Status = "paid"
		claim.ProcessedAt = ctx.BlockTime()
		if err := k.SetCompensationClaim(ctx, claim); err != nil {
			k.Logger(ctx).Error("failed to set compensation claim", "error", err)
			continue
		}

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompensationClaimed,
				sdk.NewAttribute(types.AttributeKeyCompensationID, fmt.Sprintf("%d", comp.ID)),
				sdk.NewAttribute(types.AttributeKeyClaimant, claim.Claimant),
				sdk.NewAttribute(types.AttributeKeyClaimAmount, claimAmount.String()),
				sdk.NewAttribute(types.AttributeKeySharesHeld, claim.SharesHeld.String()),
			),
		)

		// Update compensation totals
		comp.ProcessedClaims++
		comp.ClaimedAmount = comp.ClaimedAmount.Add(claimAmount)
		comp.RemainingPool = comp.RemainingPool.Sub(claimAmount)

		k.Logger(ctx).Info("compensation paid pro-rata",
			"claim_id", claim.ID,
			"claimant", claim.Claimant,
			"shares", claim.SharesHeld.String(),
			"amount", claimAmount.String(),
		)
	}

	// Mark compensation as distributed
	comp.IsDistributed = true
	comp.Status = types.CompensationStatusCompleted
	comp.CompletedAt = ctx.BlockTime()
	if err := k.SetDelistingCompensation(ctx, *comp); err != nil {
		return err
	}

	k.Logger(ctx).Info("pro-rata distribution completed",
		"compensation_id", comp.ID,
		"company_id", comp.CompanyID,
		"total_paid", comp.ClaimedAmount.String(),
		"remaining", comp.RemainingPool.String(),
	)

	return nil
}

// getClaimsByCompensation retrieves all claims for a compensation
func (k Keeper) getClaimsByCompensation(ctx sdk.Context, compensationID uint64) []types.CompensationClaim {
	store := ctx.KVStore(k.storeKey)
	prefix := types.GetClaimsByCompensationPrefix(compensationID)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var claims []types.CompensationClaim
	for ; iterator.Valid(); iterator.Next() {
		claimID := sdk.BigEndianToUint64(iterator.Value())
		if claim, found := k.GetCompensationClaim(ctx, claimID); found {
			claims = append(claims, claim)
		}
	}
	return claims
}

// =============================================================================
// EndBlock Processing
// =============================================================================

// ProcessExpiredCompensations marks expired compensations and returns remaining to community pool
func (k Keeper) ProcessExpiredCompensations(ctx sdk.Context) {
	// This would iterate all compensations and check deadlines
	// For now, this is a placeholder for EndBlock processing
}

// ProcessExpiredTradingHalts auto-resumes trading if halt period expires
func (k Keeper) ProcessExpiredTradingHalts(ctx sdk.Context) {
	// This would iterate all trading halts and check resume times
	// For now, this is a placeholder for EndBlock processing
}
