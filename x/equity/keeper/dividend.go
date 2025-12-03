package keeper

import (
	"encoding/json"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// Dividend management functions

// GetNextDividendID returns the next dividend ID and increments the counter
func (k Keeper) GetNextDividendID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DividendCounterKey)
	
	var counter uint64 = 1 // Start from 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}
	
	// Increment counter for next use
	store.Set(types.DividendCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetNextDividendPaymentID returns the next dividend payment ID and increments the counter
func (k Keeper) GetNextDividendPaymentID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DividendPaymentCounterKey)
	
	var counter uint64 = 1 // Start from 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}
	
	// Increment counter for next use
	store.Set(types.DividendPaymentCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetNextDividendClaimID returns the next dividend claim ID and increments the counter
func (k Keeper) GetNextDividendClaimID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DividendClaimCounterKey)
	
	var counter uint64 = 1 // Start from 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}
	
	// Increment counter for next use
	store.Set(types.DividendClaimCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetDividend returns a dividend by ID
func (k Keeper) GetDividend(ctx sdk.Context, dividendID uint64) (types.Dividend, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendKey(dividendID)
	bz := store.Get(key)
	if bz == nil {
		return types.Dividend{}, false
	}
	
	var dividend types.Dividend
	if err := json.Unmarshal(bz, &dividend); err != nil {
		return types.Dividend{}, false
	}
	return dividend, true
}

// SetDividend stores a dividend
func (k Keeper) SetDividend(ctx sdk.Context, dividend types.Dividend) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendKey(dividend.ID)
	bz, err := json.Marshal(dividend)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// DeleteDividend removes a dividend
func (k Keeper) DeleteDividend(ctx sdk.Context, dividendID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendKey(dividendID)
	store.Delete(key)
}

// GetDividendPayment returns a dividend payment by ID
func (k Keeper) GetDividendPayment(ctx sdk.Context, paymentID uint64) (types.DividendPayment, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendPaymentKey(paymentID)
	bz := store.Get(key)
	if bz == nil {
		return types.DividendPayment{}, false
	}
	
	var payment types.DividendPayment
	if err := json.Unmarshal(bz, &payment); err != nil {
		return types.DividendPayment{}, false
	}
	return payment, true
}

// SetDividendPayment stores a dividend payment
func (k Keeper) SetDividendPayment(ctx sdk.Context, payment types.DividendPayment) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendPaymentKey(payment.ID)
	bz, err := json.Marshal(payment)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// GetDividendSnapshot returns a dividend snapshot
func (k Keeper) GetDividendSnapshot(ctx sdk.Context, dividendID uint64) (types.DividendSnapshot, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendSnapshotKey(dividendID)
	bz := store.Get(key)
	if bz == nil {
		return types.DividendSnapshot{}, false
	}
	
	var snapshot types.DividendSnapshot
	if err := json.Unmarshal(bz, &snapshot); err != nil {
		return types.DividendSnapshot{}, false
	}
	return snapshot, true
}

// SetDividendSnapshot stores a dividend snapshot
func (k Keeper) SetDividendSnapshot(ctx sdk.Context, snapshot types.DividendSnapshot) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendSnapshotKey(snapshot.DividendID)
	bz, err := json.Marshal(snapshot)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// GetShareholderSnapshot returns a shareholder snapshot
func (k Keeper) GetShareholderSnapshot(ctx sdk.Context, dividendID uint64, shareholder string) (types.ShareholderSnapshot, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholderSnapshotKey(dividendID, shareholder)
	bz := store.Get(key)
	if bz == nil {
		return types.ShareholderSnapshot{}, false
	}
	
	var snapshot types.ShareholderSnapshot
	if err := json.Unmarshal(bz, &snapshot); err != nil {
		return types.ShareholderSnapshot{}, false
	}
	return snapshot, true
}

// SetShareholderSnapshot stores a shareholder snapshot
func (k Keeper) SetShareholderSnapshot(ctx sdk.Context, snapshot types.ShareholderSnapshot) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholderSnapshotKey(snapshot.DividendID, snapshot.Shareholder)
	bz, err := json.Marshal(snapshot)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// GetDividendPolicy returns the dividend policy for a company
func (k Keeper) GetDividendPolicy(ctx sdk.Context, companyID uint64) (types.DividendPolicy, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendPolicyKey(companyID)
	bz := store.Get(key)
	if bz == nil {
		return types.DividendPolicy{}, false
	}
	
	var policy types.DividendPolicy
	if err := json.Unmarshal(bz, &policy); err != nil {
		return types.DividendPolicy{}, false
	}
	return policy, true
}

// StoreDividendPolicy stores a dividend policy
func (k Keeper) StoreDividendPolicy(ctx sdk.Context, policy types.DividendPolicy) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendPolicyKey(policy.CompanyID)
	bz, err := json.Marshal(policy)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// Business logic methods

// DeclareDividend declares a new dividend
func (k Keeper) DeclareDividend(
	ctx sdk.Context,
	creator string,
	companyID uint64,
	classID string,
	dividendType types.DividendType,
	currency string,
	amountPerShare math.LegacyDec,
	exDividendDays, recordDays, paymentDays uint32,
	taxRate math.LegacyDec,
	description string,
	paymentMethod string,
	stockRatio math.LegacyDec,
) (uint64, error) {
	// Get company to verify creator is authorized
	company, found := k.GetCompany(ctx, companyID)
	if !found {
		return 0, types.ErrCompanyNotFound
	}
	
	// Only company founder can declare dividends (could add board governance later)
	if company.Founder != creator {
		return 0, types.ErrUnauthorized
	}
	
	// Company must be active
	if company.Status != types.CompanyStatusActive {
		return 0, types.ErrCompanyNotActive
	}
	
	// Calculate dividend dates
	now := ctx.BlockTime()
	declarationDate := now
	exDividendDate := now.AddDate(0, 0, int(exDividendDays))
	recordDate := now.AddDate(0, 0, int(recordDays))
	paymentDate := now.AddDate(0, 0, int(paymentDays))
	
	// Get next dividend ID
	dividendID := k.GetNextDividendID(ctx)
	
	// Create dividend
	dividend := types.NewDividend(
		dividendID,
		companyID,
		classID,
		dividendType,
		currency,
		amountPerShare,
		declarationDate,
		exDividendDate,
		recordDate,
		paymentDate,
		creator,
	)
	
	dividend.TaxRate = taxRate
	dividend.Description = description
	if dividendType == types.DividendTypeStock {
		dividend.StockRatio = stockRatio
	}
	
	// Validate dividend
	if err := dividend.Validate(); err != nil {
		return 0, err
	}
	
	// Calculate total dividend amount needed (simplified)
	var totalShares math.Int
	if classID == "" {
		// All share classes
		shareClasses := k.GetCompanyShareClasses(ctx, companyID)
		for _, class := range shareClasses {
			totalShares = totalShares.Add(class.OutstandingShares)
		}
	} else {
		// Specific share class
		totalShares = k.GetTotalShares(ctx, companyID, classID)
	}
	
	dividend.TotalAmount = amountPerShare.MulInt(totalShares)
	dividend.RemainingAmount = dividend.TotalAmount
	dividend.EligibleShares = totalShares
	
	// For cash dividends, verify founder has sufficient funds
	if dividendType == types.DividendTypeCash {
		creatorAddr, err := sdk.AccAddressFromBech32(creator)
		if err != nil {
			return 0, types.ErrUnauthorized
		}
		
		requiredPayment := sdk.NewCoins(sdk.NewCoin(currency, dividend.TotalAmount.TruncateInt()))
		balance := k.bankKeeper.GetAllBalances(ctx, creatorAddr)
		if !balance.IsAllGTE(requiredPayment) {
			return 0, types.ErrInsufficientDividendFunds
		}
		
		// Transfer funds to module account for dividend distribution
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, requiredPayment); err != nil {
			return 0, types.ErrInsufficientDividendFunds
		}
	}
	
	// Store dividend
	k.SetDividend(ctx, dividend)
	
	return dividendID, nil
}

// CreateRecordSnapshot creates a snapshot of shareholders at the record date
func (k Keeper) CreateRecordSnapshot(ctx sdk.Context, dividendID uint64) error {
	// Get dividend
	dividend, found := k.GetDividend(ctx, dividendID)
	if !found {
		return types.ErrDividendNotFound
	}
	
	// Check if record date has been reached
	if !dividend.IsRecordDateReached(ctx.BlockTime()) {
		return types.ErrDividendSnapshotNotReady
	}
	
	// Check if snapshot already exists
	if _, exists := k.GetDividendSnapshot(ctx, dividendID); exists {
		return types.ErrDividendSnapshotExists
	}
	
	// Create dividend snapshot
	snapshot := types.NewDividendSnapshot(
		dividendID,
		dividend.CompanyID,
		dividend.ClassID,
		dividend.RecordDate,
	)
	
	// Get all shareholdings for the company/class
	var totalShares math.Int
	var totalShareholders uint64
	
	// Simplified: Get all shareholdings for the company
	// In a full implementation, would iterate through all shareholdings
	shareholdings := k.GetCompanyShareholdings(ctx, dividend.CompanyID, dividend.ClassID)
	
	for _, shareholding := range shareholdings {
		if shareholding.VestedShares.GT(math.ZeroInt()) {
			// Create shareholder snapshot
			shareholderSnapshot := types.NewShareholderSnapshot(
				dividendID,
				dividend.CompanyID,
				dividend.ClassID,
				shareholding.Owner,
				shareholding.VestedShares, // Use vested shares for dividends
				dividend.RecordDate,
			)
			k.SetShareholderSnapshot(ctx, shareholderSnapshot)
			
			totalShares = totalShares.Add(shareholding.VestedShares)
			totalShareholders++
		}
	}
	
	// Update snapshot totals
	snapshot.TotalShares = totalShares
	snapshot.TotalShareholders = totalShareholders
	snapshot.SnapshotTaken = true
	
	// Store snapshot
	k.SetDividendSnapshot(ctx, snapshot)
	
	// Update dividend with eligible shareholders count
	dividend.EligibleShares = totalShares
	dividend.ShareholdersEligible = totalShareholders
	dividend.Status = types.DividendStatusRecorded
	dividend.UpdatedAt = ctx.BlockTime()
	k.SetDividend(ctx, dividend)
	
	return nil
}

// ProcessDividendPayments processes dividend payments in batches
func (k Keeper) ProcessDividendPayments(ctx sdk.Context, dividendID uint64, batchSize uint32) (uint32, math.LegacyDec, bool, error) {
	// Get dividend
	dividend, found := k.GetDividend(ctx, dividendID)
	if !found {
		return 0, math.LegacyZeroDec(), false, types.ErrDividendNotFound
	}
	
	// Check if dividend is ready for payment
	if !dividend.IsReadyForPayment(ctx.BlockTime()) {
		return 0, math.LegacyZeroDec(), false, types.ErrDividendNotProcessable
	}
	
	// Get dividend snapshot
	if _, found := k.GetDividendSnapshot(ctx, dividendID); !found {
		return 0, math.LegacyZeroDec(), false, types.ErrDividendSnapshotNotReady
	}
	
	// Process payments in batches
	paymentsProcessed := uint32(0)
	totalPaid := math.LegacyZeroDec()
	
	// Get shareholder snapshots (simplified - would use pagination in production)
	shareholderSnapshots := k.GetShareholderSnapshotsByDividend(ctx, dividendID)
	
	startIndex := dividend.ShareholdersPaid
	endIndex := startIndex + uint64(batchSize)
	if endIndex > uint64(len(shareholderSnapshots)) {
		endIndex = uint64(len(shareholderSnapshots))
	}
	
	for i := startIndex; i < endIndex; i++ {
		shareholderSnapshot := shareholderSnapshots[i]
		
		// Check if already paid
		if _, exists := k.GetDividendPaymentForShareholder(ctx, dividendID, shareholderSnapshot.Shareholder); exists {
			continue
		}
		
		// Calculate payment
		grossAmount := dividend.CalculatePayment(shareholderSnapshot.SharesHeld)
		taxWithheld := dividend.CalculateWithholding(grossAmount)
		netAmount := grossAmount.Sub(taxWithheld)
		
		// Create payment record
		paymentID := k.GetNextDividendPaymentID(ctx)
		payment := types.NewDividendPayment(
			paymentID,
			dividendID,
			dividend.CompanyID,
			dividend.ClassID,
			shareholderSnapshot.Shareholder,
			shareholderSnapshot.SharesHeld,
			grossAmount,
			dividend.Currency,
		)
		payment.TaxWithheld = taxWithheld
		payment.NetAmount = netAmount
		
		// Process payment based on dividend type
		if dividend.Type == types.DividendTypeCash {
			// Transfer funds from module to shareholder
			shareholderAddr, err := sdk.AccAddressFromBech32(shareholderSnapshot.Shareholder)
			if err != nil {
				payment.Status = "failed"
				payment.FailureReason = "invalid address"
			} else {
				paymentCoins := sdk.NewCoins(sdk.NewCoin(dividend.Currency, netAmount.TruncateInt()))
				if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, shareholderAddr, paymentCoins); err != nil {
					payment.Status = "failed"
					payment.FailureReason = err.Error()
				} else {
					payment.Status = "paid"
					payment.PaidAt = ctx.BlockTime()
					totalPaid = totalPaid.Add(grossAmount)
					paymentsProcessed++
				}
			}
		} else if dividend.Type == types.DividendTypeStock {
			// Issue new shares based on stock ratio
			newShares := dividend.StockRatio.MulInt(shareholderSnapshot.SharesHeld).TruncateInt()
			if err := k.IssueShares(
				ctx,
				dividend.CompanyID,
				dividend.ClassID,
				shareholderSnapshot.Shareholder,
				newShares,
				math.LegacyZeroDec(), // No cost for stock dividends
				"",
			); err != nil {
				payment.Status = "failed"
				payment.FailureReason = err.Error()
			} else {
				payment.Status = "paid"
				payment.PaidAt = ctx.BlockTime()
				dividend.NewSharesIssued = dividend.NewSharesIssued.Add(newShares)
				paymentsProcessed++
			}
		}
		
		// Store payment record
		k.SetDividendPayment(ctx, payment)
	}
	
	// Update dividend progress
	dividend.ShareholdersPaid += uint64(paymentsProcessed)
	dividend.SharesProcessed = dividend.SharesProcessed.Add(math.NewInt(int64(paymentsProcessed)))
	dividend.PaidAmount = dividend.PaidAmount.Add(totalPaid)
	dividend.RemainingAmount = dividend.TotalAmount.Sub(dividend.PaidAmount)
	dividend.UpdatedAt = ctx.BlockTime()
	
	// Check if dividend is complete
	isComplete := dividend.ShareholdersPaid >= dividend.ShareholdersEligible
	if isComplete {
		dividend.Status = types.DividendStatusPaid
		dividend.CompletedAt = ctx.BlockTime()
	} else {
		dividend.Status = types.DividendStatusProcessing
	}
	
	k.SetDividend(ctx, dividend)
	
	return paymentsProcessed, totalPaid, isComplete, nil
}

// ClaimDividend allows a shareholder to claim their dividend
func (k Keeper) ClaimDividend(ctx sdk.Context, claimant string, dividendID uint64) (math.LegacyDec, string, math.LegacyDec, math.LegacyDec, error) {
	// Get dividend
	dividend, found := k.GetDividend(ctx, dividendID)
	if !found {
		return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), types.ErrDividendNotFound
	}
	
	// Check if dividend is ready for claims
	if !dividend.IsReadyForPayment(ctx.BlockTime()) {
		return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), types.ErrDividendNotProcessable
	}
	
	// Check if already claimed/paid
	if _, exists := k.GetDividendPaymentForShareholder(ctx, dividendID, claimant); exists {
		return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), types.ErrDividendAlreadyClaimed
	}
	
	// Get shareholder snapshot to verify eligibility
	shareholderSnapshot, found := k.GetShareholderSnapshot(ctx, dividendID, claimant)
	if !found {
		return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), types.ErrDividendNotEligible
	}
	
	// Calculate payment
	grossAmount := dividend.CalculatePayment(shareholderSnapshot.SharesHeld)
	taxWithheld := dividend.CalculateWithholding(grossAmount)
	netAmount := grossAmount.Sub(taxWithheld)
	
	// Create payment record
	paymentID := k.GetNextDividendPaymentID(ctx)
	payment := types.NewDividendPayment(
		paymentID,
		dividendID,
		dividend.CompanyID,
		dividend.ClassID,
		claimant,
		shareholderSnapshot.SharesHeld,
		grossAmount,
		dividend.Currency,
	)
	payment.TaxWithheld = taxWithheld
	payment.NetAmount = netAmount
	
	// Process payment
	if dividend.Type == types.DividendTypeCash {
		// Transfer funds from module to claimant
		claimantAddr, err := sdk.AccAddressFromBech32(claimant)
		if err != nil {
			return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), types.ErrUnauthorized
		}
		
		paymentCoins := sdk.NewCoins(sdk.NewCoin(dividend.Currency, netAmount.TruncateInt()))
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, claimantAddr, paymentCoins); err != nil {
			return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), types.ErrDividendPaymentFailed
		}
		
		payment.Status = "paid"
		payment.PaidAt = ctx.BlockTime()
	} else {
		return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), types.ErrInvalidPaymentMethod
	}
	
	// Store payment record
	k.SetDividendPayment(ctx, payment)
	
	// Update dividend totals
	dividend.PaidAmount = dividend.PaidAmount.Add(grossAmount)
	dividend.RemainingAmount = dividend.TotalAmount.Sub(dividend.PaidAmount)
	dividend.ShareholdersPaid++
	dividend.UpdatedAt = ctx.BlockTime()
	k.SetDividend(ctx, dividend)
	
	return grossAmount, dividend.Currency, taxWithheld, netAmount, nil
}

// Helper functions

// GetShareholderSnapshotsByDividend returns all shareholder snapshots for a dividend
func (k Keeper) GetShareholderSnapshotsByDividend(ctx sdk.Context, dividendID uint64) []types.ShareholderSnapshot {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ShareholderSnapshotPrefix)
	prefixBytes := sdk.Uint64ToBigEndian(dividendID)
	iterator := store.Iterator(prefixBytes, append(prefixBytes, 0xFF))
	defer iterator.Close()
	
	snapshots := []types.ShareholderSnapshot{}
	for ; iterator.Valid(); iterator.Next() {
		var snapshot types.ShareholderSnapshot
		if err := json.Unmarshal(iterator.Value(), &snapshot); err == nil {
			snapshots = append(snapshots, snapshot)
		}
	}
	
	return snapshots
}

// GetDividendPaymentForShareholder returns a dividend payment for a specific shareholder
func (k Keeper) GetDividendPaymentForShareholder(ctx sdk.Context, dividendID uint64, shareholder string) (types.DividendPayment, bool) {
	// Simplified: iterate through all payments to find the one for this shareholder
	// In production, would use a more efficient index
	payments := k.GetDividendPaymentsByDividend(ctx, dividendID)
	for _, payment := range payments {
		if payment.Shareholder == shareholder {
			return payment, true
		}
	}
	return types.DividendPayment{}, false
}

// GetDividendPaymentsByDividend returns all payments for a dividend
func (k Keeper) GetDividendPaymentsByDividend(ctx sdk.Context, dividendID uint64) []types.DividendPayment {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DividendPaymentPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	
	payments := []types.DividendPayment{}
	for ; iterator.Valid(); iterator.Next() {
		var payment types.DividendPayment
		if err := json.Unmarshal(iterator.Value(), &payment); err == nil && payment.DividendID == dividendID {
			payments = append(payments, payment)
		}
	}
	
	return payments
}

// GetDividendsByCompany returns all dividends for a company
func (k Keeper) GetDividendsByCompany(ctx sdk.Context, companyID uint64) []types.Dividend {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DividendPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	
	dividends := []types.Dividend{}
	for ; iterator.Valid(); iterator.Next() {
		var dividend types.Dividend
		if err := json.Unmarshal(iterator.Value(), &dividend); err == nil && dividend.CompanyID == companyID {
			dividends = append(dividends, dividend)
		}
	}
	
	return dividends
}