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
// Anti-Dilution Provision Management
// =============================================================================

// SetAntiDilutionProvision stores an anti-dilution provision
func (k Keeper) SetAntiDilutionProvision(ctx sdk.Context, provision types.AntiDilutionProvision) error {
	if err := provision.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetAntiDilutionProvisionKey(provision.CompanyID, provision.ClassID)
	bz, err := json.Marshal(provision)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetAntiDilutionProvision retrieves an anti-dilution provision
func (k Keeper) GetAntiDilutionProvision(ctx sdk.Context, companyID uint64, classID string) (types.AntiDilutionProvision, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAntiDilutionProvisionKey(companyID, classID)
	bz := store.Get(key)
	if bz == nil {
		return types.AntiDilutionProvision{}, false
	}

	var provision types.AntiDilutionProvision
	if err := json.Unmarshal(bz, &provision); err != nil {
		return types.AntiDilutionProvision{}, false
	}
	return provision, true
}

// DeleteAntiDilutionProvision removes an anti-dilution provision
func (k Keeper) DeleteAntiDilutionProvision(ctx sdk.Context, companyID uint64, classID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAntiDilutionProvisionKey(companyID, classID)
	store.Delete(key)
}

// RegisterAntiDilutionProvision creates a new anti-dilution provision for a share class
func (k Keeper) RegisterAntiDilutionProvision(
	ctx sdk.Context,
	companyID uint64,
	classID string,
	provisionType types.AntiDilutionType,
	triggerThreshold, adjustmentCap, minimumPrice math.LegacyDec,
	creator string,
) error {
	// Check if company exists
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Only founder can register anti-dilution provisions
	if company.Founder != creator {
		return types.ErrUnauthorized
	}

	// Check if share class exists
	_, found = k.getShareClass(ctx, companyID, classID)
	if !found {
		return types.ErrShareClassNotFound
	}

	// Check if provision already exists
	if _, exists := k.GetAntiDilutionProvision(ctx, companyID, classID); exists {
		return types.ErrAntiDilutionExists
	}

	// Create and store provision
	provision := types.NewAntiDilutionProvision(
		companyID,
		classID,
		provisionType,
		triggerThreshold,
		adjustmentCap,
		minimumPrice,
	)

	if err := k.SetAntiDilutionProvision(ctx, provision); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"anti_dilution_provision_registered",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute("class_id", classID),
			sdk.NewAttribute("provision_type", provisionType.String()),
		),
	)

	return nil
}

// =============================================================================
// Issuance Record Management
// =============================================================================

// GetNextIssuanceRecordID returns the next issuance record ID
func (k Keeper) GetNextIssuanceRecordID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.IssuanceRecordCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.IssuanceRecordCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// SetIssuanceRecord stores an issuance record
func (k Keeper) SetIssuanceRecord(ctx sdk.Context, record types.IssuanceRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetIssuanceRecordKey(record.ID)
	bz, err := json.Marshal(record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetIssuanceRecord retrieves an issuance record by ID
func (k Keeper) GetIssuanceRecord(ctx sdk.Context, recordID uint64) (types.IssuanceRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetIssuanceRecordKey(recordID)
	bz := store.Get(key)
	if bz == nil {
		return types.IssuanceRecord{}, false
	}

	var record types.IssuanceRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.IssuanceRecord{}, false
	}
	return record, true
}

// SetLastIssuancePrice stores the last issuance price for a share class
func (k Keeper) SetLastIssuancePrice(ctx sdk.Context, companyID uint64, classID string, price math.LegacyDec) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLastIssuancePriceKey(companyID, classID)
	bz, _ := price.Marshal()
	store.Set(key, bz)
}

// GetLastIssuancePrice retrieves the last issuance price for a share class
func (k Keeper) GetLastIssuancePrice(ctx sdk.Context, companyID uint64, classID string) (math.LegacyDec, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLastIssuancePriceKey(companyID, classID)
	bz := store.Get(key)
	if bz == nil {
		return math.LegacyZeroDec(), false
	}

	var price math.LegacyDec
	if err := price.Unmarshal(bz); err != nil {
		return math.LegacyZeroDec(), false
	}
	return price, true
}

// RecordIssuance creates an issuance record and detects down rounds
func (k Keeper) RecordIssuance(
	ctx sdk.Context,
	companyID uint64,
	classID string,
	sharesIssued math.Int,
	issuePrice math.LegacyDec,
	recipient string,
	purpose string,
) (types.IssuanceRecord, error) {
	// Get previous price
	previousPrice, _ := k.GetLastIssuancePrice(ctx, companyID, classID)
	if previousPrice.IsZero() {
		// First issuance - use the share class par value
		shareClass, found := k.getShareClass(ctx, companyID, classID)
		if found {
			previousPrice = shareClass.ParValue
		}
	}

	// Create record
	recordID := k.GetNextIssuanceRecordID(ctx)
	record := types.NewIssuanceRecord(
		recordID,
		companyID,
		classID,
		sharesIssued,
		issuePrice,
		recipient,
		previousPrice,
		purpose,
	)

	// Store record
	if err := k.SetIssuanceRecord(ctx, record); err != nil {
		return types.IssuanceRecord{}, err
	}

	// Update last issuance price
	k.SetLastIssuancePrice(ctx, companyID, classID, issuePrice)

	return record, nil
}

// =============================================================================
// Anti-Dilution Adjustment Management
// =============================================================================

// GetNextAdjustmentID returns the next adjustment ID
func (k Keeper) GetNextAdjustmentID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AdjustmentCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.AdjustmentCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// SetAdjustment stores an adjustment record
func (k Keeper) SetAdjustment(ctx sdk.Context, adjustment types.AntiDilutionAdjustment) error {
	if err := adjustment.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetAntiDilutionAdjustmentKey(adjustment.ID)
	bz, err := json.Marshal(adjustment)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	return nil
}

// GetAdjustment retrieves an adjustment by ID
func (k Keeper) GetAdjustment(ctx sdk.Context, adjustmentID uint64) (types.AntiDilutionAdjustment, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAntiDilutionAdjustmentKey(adjustmentID)
	bz := store.Get(key)
	if bz == nil {
		return types.AntiDilutionAdjustment{}, false
	}

	var adjustment types.AntiDilutionAdjustment
	if err := json.Unmarshal(bz, &adjustment); err != nil {
		return types.AntiDilutionAdjustment{}, false
	}
	return adjustment, true
}

// =============================================================================
// Anti-Dilution Processing Logic
// =============================================================================

// ProcessAntiDilutionAdjustments processes anti-dilution adjustments for a down round
func (k Keeper) ProcessAntiDilutionAdjustments(
	ctx sdk.Context,
	issuanceRecord types.IssuanceRecord,
) ([]types.AntiDilutionAdjustment, error) {
	// Only process if this is a down round
	if !issuanceRecord.IsDownRound {
		return nil, nil
	}

	// Get all share classes with anti-dilution provisions for this company
	adjustments := []types.AntiDilutionAdjustment{}

	// Get the provision for the issued class (protected classes typically are preferred shares)
	// In a typical scenario, preferred shares protect against common share down rounds
	provision, found := k.GetAntiDilutionProvision(ctx, issuanceRecord.CompanyID, issuanceRecord.ClassID)
	if !found || !provision.IsActive {
		return nil, nil
	}

	// Check if price drop exceeds trigger threshold
	if issuanceRecord.PriceDropPercent.LT(provision.TriggerThreshold) {
		return nil, nil
	}

	// Check minimum price
	if issuanceRecord.IssuePrice.LT(provision.MinimumPrice) {
		return nil, types.ErrBelowMinimumPrice
	}

	// Get all shareholders for this class
	shareholders := k.GetCompanyShareholdings(ctx, issuanceRecord.CompanyID, issuanceRecord.ClassID)
	if len(shareholders) == 0 {
		return nil, nil
	}

	// Get share class for total outstanding
	shareClass, found := k.getShareClass(ctx, issuanceRecord.CompanyID, issuanceRecord.ClassID)
	if !found {
		return nil, types.ErrShareClassNotFound
	}

	// Process each shareholder
	for _, holding := range shareholders {
		// Skip the recipient of the new issuance (they don't get protection on their own purchase)
		if holding.Owner == issuanceRecord.Recipient {
			continue
		}

		var adjustedShares math.Int
		var adjustmentRatio math.LegacyDec

		switch provision.ProvisionType {
		case types.AntiDilutionFullRatchet:
			adjustedShares, adjustmentRatio = types.CalculateFullRatchetAdjustment(
				holding.Shares,
				issuanceRecord.PreviousPrice,
				issuanceRecord.IssuePrice,
			)

		case types.AntiDilutionWeightedAverage, types.AntiDilutionBroadBasedWeightedAverage:
			adjustedShares, adjustmentRatio = types.CalculateWeightedAverageAdjustment(
				holding.Shares,
				issuanceRecord.PreviousPrice,
				issuanceRecord.IssuePrice,
				shareClass.OutstandingShares,
				issuanceRecord.SharesIssued,
			)

		default:
			continue
		}

		// Check if adjustment exceeds cap
		maxAdjustedShares := math.LegacyOneDec().Add(provision.AdjustmentCap).MulInt(holding.Shares).TruncateInt()
		if adjustedShares.GT(maxAdjustedShares) {
			adjustedShares = maxAdjustedShares
			adjustmentRatio = math.LegacyNewDecFromInt(adjustedShares).QuoInt(holding.Shares)
		}

		// Only create adjustment if shares actually increased
		if adjustedShares.LTE(holding.Shares) {
			continue
		}

		// Create adjustment record
		adjustmentID := k.GetNextAdjustmentID(ctx)
		adjustment := types.NewAntiDilutionAdjustment(
			adjustmentID,
			issuanceRecord.CompanyID,
			issuanceRecord.ClassID,
			issuanceRecord.ID,
			provision.ProvisionType,
			holding.Owner,
			holding.Shares,
			adjustedShares,
			issuanceRecord.PreviousPrice,
			issuanceRecord.IssuePrice,
			adjustmentRatio,
		)

		// Apply adjustment to shareholding
		if err := k.ApplyAdjustmentToShareholding(ctx, adjustment); err != nil {
			return nil, err
		}

		// Store adjustment record
		if err := k.SetAdjustment(ctx, adjustment); err != nil {
			return nil, err
		}

		adjustments = append(adjustments, adjustment)

		// Emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"anti_dilution_adjustment_applied",
				sdk.NewAttribute("adjustment_id", fmt.Sprintf("%d", adjustmentID)),
				sdk.NewAttribute("shareholder", holding.Owner),
				sdk.NewAttribute("original_shares", holding.Shares.String()),
				sdk.NewAttribute("adjusted_shares", adjustedShares.String()),
				sdk.NewAttribute("adjustment_ratio", adjustmentRatio.String()),
			),
		)
	}

	// Update provision statistics
	provision.TotalAdjustments += uint64(len(adjustments))
	provision.LastTriggerDate = ctx.BlockTime()
	provision.UpdatedAt = time.Now()
	k.SetAntiDilutionProvision(ctx, provision)

	return adjustments, nil
}

// ApplyAdjustmentToShareholding applies an anti-dilution adjustment to a shareholding
func (k Keeper) ApplyAdjustmentToShareholding(ctx sdk.Context, adjustment types.AntiDilutionAdjustment) error {
	holding, found := k.getShareholding(ctx, adjustment.CompanyID, adjustment.ClassID, adjustment.Shareholder)
	if !found {
		return types.ErrShareholdingNotFound
	}

	// Update shares
	sharesAdded := adjustment.AdjustedShares.Sub(holding.Shares)
	holding.Shares = adjustment.AdjustedShares
	holding.VestedShares = holding.VestedShares.Add(sharesAdded)

	// Recalculate cost basis (adjusted shares have zero cost for anti-dilution)
	// This maintains the original total cost but spreads it over more shares
	if holding.Shares.IsPositive() {
		holding.CostBasis = holding.TotalCost.QuoInt(holding.Shares)
	}

	holding.UpdatedAt = time.Now()
	k.SetShareholding(ctx, holding)

	// Update share class outstanding shares
	shareClass, found := k.getShareClass(ctx, adjustment.CompanyID, adjustment.ClassID)
	if found {
		shareClass.IssuedShares = shareClass.IssuedShares.Add(sharesAdded)
		shareClass.OutstandingShares = shareClass.OutstandingShares.Add(sharesAdded)
		shareClass.UpdatedAt = time.Now()
		k.SetShareClass(ctx, shareClass)
	}

	return nil
}

// =============================================================================
// Enhanced Issue Shares with Anti-Dilution
// =============================================================================

// IssueSharesWithAntiDilution issues new shares and processes any anti-dilution adjustments
func (k Keeper) IssueSharesWithAntiDilution(
	ctx sdk.Context,
	companyID uint64,
	classID string,
	recipient string,
	shares math.Int,
	pricePerShare math.LegacyDec,
	paymentDenom string,
	purpose string,
) ([]types.AntiDilutionAdjustment, error) {
	// First, issue the shares using existing logic
	if err := k.IssueShares(ctx, companyID, classID, recipient, shares, pricePerShare, paymentDenom); err != nil {
		return nil, err
	}

	// Record the issuance
	record, err := k.RecordIssuance(ctx, companyID, classID, shares, pricePerShare, recipient, purpose)
	if err != nil {
		return nil, err
	}

	// Process anti-dilution adjustments if this was a down round
	adjustments, err := k.ProcessAntiDilutionAdjustments(ctx, record)
	if err != nil {
		return nil, err
	}

	return adjustments, nil
}

// =============================================================================
// Query Methods
// =============================================================================

// GetAntiDilutionProvisions returns all provisions for a company
func (k Keeper) GetAntiDilutionProvisions(ctx sdk.Context, companyID uint64) []types.AntiDilutionProvision {
	provisions := []types.AntiDilutionProvision{}

	// Get all share classes for the company
	shareClasses := k.GetCompanyShareClasses(ctx, companyID)
	for _, sc := range shareClasses {
		if provision, found := k.GetAntiDilutionProvision(ctx, companyID, sc.ClassID); found {
			provisions = append(provisions, provision)
		}
	}

	return provisions
}

// GetShareholderAdjustments returns all adjustments for a shareholder
func (k Keeper) GetShareholderAdjustments(ctx sdk.Context, companyID uint64, classID, shareholder string) []types.AntiDilutionAdjustment {
	// This is a simplified implementation - would need iteration in production
	return []types.AntiDilutionAdjustment{}
}

// GetIssuanceHistory returns all issuance records for a company
func (k Keeper) GetIssuanceHistory(ctx sdk.Context, companyID uint64, classID string) []types.IssuanceRecord {
	store := ctx.KVStore(k.storeKey)
	records := []types.IssuanceRecord{}

	// Iterate through all issuance records
	// In production, we'd use a secondary index by companyID
	iterator := storetypes.KVStorePrefixIterator(store, types.IssuanceRecordPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record types.IssuanceRecord
		if err := json.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}

		// Filter by company and optionally by class
		if record.CompanyID == companyID {
			if classID == "" || record.ClassID == classID {
				records = append(records, record)
			}
		}
	}

	return records
}

// GetAdjustmentHistory returns all anti-dilution adjustments matching filters
func (k Keeper) GetAdjustmentHistory(ctx sdk.Context, companyID uint64, classID, shareholder string) []types.AntiDilutionAdjustment {
	store := ctx.KVStore(k.storeKey)
	adjustments := []types.AntiDilutionAdjustment{}

	// Iterate through all adjustment records
	iterator := storetypes.KVStorePrefixIterator(store, types.AntiDilutionAdjustmentPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var adjustment types.AntiDilutionAdjustment
		if err := json.Unmarshal(iterator.Value(), &adjustment); err != nil {
			continue
		}

		// Apply filters
		if companyID != 0 && adjustment.CompanyID != companyID {
			continue
		}
		if classID != "" && adjustment.ClassID != classID {
			continue
		}
		if shareholder != "" && adjustment.Shareholder != shareholder {
			continue
		}

		adjustments = append(adjustments, adjustment)
	}

	return adjustments
}
