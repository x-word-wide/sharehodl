package keeper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
func (k Keeper) SetDividend(ctx sdk.Context, dividend types.Dividend) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendKey(dividend.ID)
	bz, err := json.Marshal(dividend)
	if err != nil {
		return fmt.Errorf("failed to marshal dividend: %w", err)
	}
	store.Set(key, bz)
	return nil
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
func (k Keeper) SetDividendPayment(ctx sdk.Context, payment types.DividendPayment) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendPaymentKey(payment.ID)
	bz, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to marshal dividend payment: %w", err)
	}
	store.Set(key, bz)
	return nil
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
func (k Keeper) SetDividendSnapshot(ctx sdk.Context, snapshot types.DividendSnapshot) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendSnapshotKey(snapshot.DividendID)
	bz, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal dividend snapshot: %w", err)
	}
	store.Set(key, bz)
	return nil
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
func (k Keeper) SetShareholderSnapshot(ctx sdk.Context, snapshot types.ShareholderSnapshot) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholderSnapshotKey(snapshot.DividendID, snapshot.Shareholder)
	bz, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal shareholder snapshot: %w", err)
	}
	store.Set(key, bz)
	return nil
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
func (k Keeper) StoreDividendPolicy(ctx sdk.Context, policy types.DividendPolicy) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendPolicyKey(policy.CompanyID)
	bz, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal dividend policy: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// Business logic methods

// DeclareDividend declares a new dividend with mandatory audit document
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
	auditInfo types.AuditInfo, // MANDATORY audit document
) (uint64, error) {
	// Get company to verify creator is authorized
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return 0, types.ErrCompanyNotFound
	}

	// Use authorization system: founder OR authorized proposer can declare dividends
	if !k.CanProposeForCompany(ctx, companyID, creator) {
		return 0, types.ErrNotAuthorizedProposer
	}

	// Company must be active
	if company.Status != types.CompanyStatusActive {
		return 0, types.ErrCompanyNotActive
	}

	// Validate audit document (MANDATORY for dividend declaration)
	if err := auditInfo.Validate(); err != nil {
		return 0, err
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

	// For cash dividends, verify TREASURY has sufficient funds (NOT founder)
	if dividendType == types.DividendTypeCash {
		treasury, found := k.GetCompanyTreasury(ctx, companyID)
		if !found {
			return 0, types.ErrTreasuryNotFound
		}

		// Check treasury isn't frozen
		if k.IsTreasuryFrozen(ctx, companyID) {
			return 0, types.ErrTreasuryFrozenForInvestigation
		}

		requiredAmount := dividend.TotalAmount.TruncateInt()
		treasuryBalance := treasury.Balance.AmountOf(currency)

		if treasuryBalance.LT(requiredAmount) {
			return 0, fmt.Errorf("%w: need %s %s, treasury has %s",
				types.ErrInsufficientTreasuryBalance,
				requiredAmount.String(), currency,
				treasuryBalance.String())
		}

		// Lock dividend amount in treasury (deduct from available balance)
		// The funds are already in the module account, just track the allocation
		lockedCoins := sdk.NewCoins(sdk.NewCoin(currency, requiredAmount))
		treasury.Balance = treasury.Balance.Sub(lockedCoins...)
		treasury.UpdatedAt = ctx.BlockTime()
		if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
			return 0, err
		}

		// Note: Funds are already in equity module account from treasury deposits
		// We just deducted from treasury tracking, funds will be distributed from module
	}
	
	// Create audit record for this dividend (MANDATORY)
	auditID, err := k.CreateAuditForDividend(ctx, dividendID, companyID, creator, auditInfo)
	if err != nil {
		return 0, fmt.Errorf("failed to create audit record: %w", err)
	}

	// Link audit to dividend
	dividend.AuditID = auditID
	dividend.AuditHash = auditInfo.ContentHash

	// Store dividend
	if err := k.SetDividend(ctx, dividend); err != nil {
		return 0, err
	}

	k.Logger(ctx).Info("dividend declared with audit",
		"dividend_id", dividendID,
		"company_id", companyID,
		"audit_id", auditID,
		"audit_hash", auditInfo.ContentHash,
		"amount_per_share", amountPerShare.String(),
	)

	return dividendID, nil
}

// CreateRecordSnapshot creates a snapshot of shareholders at the record date
// This now uses the beneficial owner registry to ensure dividends go to the TRUE owners
// even when shares are held in escrow, lending, or DEX module accounts
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

	// Get all dividend recipients (includes beneficial owners from module accounts)
	recipients := k.GetBeneficialOwnersForDividend(ctx, dividend.CompanyID, dividend.ClassID)

	var totalShares math.Int
	var totalShareholders uint64

	// Create shareholder snapshots for each recipient
	for _, recipient := range recipients {
		if recipient.Shares.GT(math.ZeroInt()) {
			// Create shareholder snapshot
			shareholderSnapshot := types.NewShareholderSnapshot(
				dividendID,
				dividend.CompanyID,
				dividend.ClassID,
				recipient.Address,
				recipient.Shares,
				dividend.RecordDate,
			)
			if err := k.SetShareholderSnapshot(ctx, shareholderSnapshot); err != nil {
				return err
			}

			totalShares = totalShares.Add(recipient.Shares)
			totalShareholders++

			k.Logger(ctx).Debug("created shareholder snapshot for dividend",
				"dividend_id", dividendID,
				"shareholder", recipient.Address,
				"shares", recipient.Shares.String(),
			)
		}
	}

	// Update snapshot totals
	snapshot.TotalShares = totalShares
	snapshot.TotalShareholders = totalShareholders
	snapshot.SnapshotTaken = true

	// Store snapshot
	if err := k.SetDividendSnapshot(ctx, snapshot); err != nil {
		return err
	}

	// Update dividend with eligible shareholders count
	dividend.EligibleShares = totalShares
	dividend.ShareholdersEligible = totalShareholders
	dividend.Status = types.DividendStatusRecorded
	dividend.UpdatedAt = ctx.BlockTime()
	if err := k.SetDividend(ctx, dividend); err != nil {
		return err
	}

	k.Logger(ctx).Info("dividend snapshot created with beneficial owners",
		"dividend_id", dividendID,
		"company_id", dividend.CompanyID,
		"class_id", dividend.ClassID,
		"total_shareholders", totalShareholders,
		"total_shares", totalShares.String(),
	)

	return nil
}

// ProcessDividendPayments processes dividend payments in batches
// PERFORMANCE FIX: Uses iterator with limits instead of loading all shareholders
func (k Keeper) ProcessDividendPayments(ctx sdk.Context, dividendID uint64, batchSize uint32) (uint32, math.LegacyDec, bool, error) {
	// Get dividend
	dividend, found := k.GetDividend(ctx, dividendID)
	if !found {
		return 0, math.LegacyZeroDec(), false, types.ErrDividendNotFound
	}

	// CRITICAL: Check audit verification status before processing payments
	if dividend.AuditID > 0 {
		audit, found := k.GetDividendAudit(ctx, dividend.AuditID)
		if !found {
			return 0, math.LegacyZeroDec(), false, types.ErrAuditNotFound
		}

		// Audit must be verified before dividends can be paid
		if audit.Status != types.AuditStatusVerified {
			return 0, math.LegacyZeroDec(), false, types.ErrAuditNotVerified
		}
	} else {
		// No audit linked - this should not happen for new dividends
		k.Logger(ctx).Warn("dividend has no audit linked - old dividend or missing audit",
			"dividend_id", dividendID,
		)
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

	// PERFORMANCE FIX: Use iterator with limits instead of loading all shareholders
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ShareholderSnapshotPrefix)
	prefixBytes := sdk.Uint64ToBigEndian(dividendID)
	iterator := store.Iterator(prefixBytes, append(prefixBytes, 0xFF))
	defer iterator.Close()

	// Skip already processed shareholders
	var skipped uint64
	for ; iterator.Valid() && skipped < dividend.ShareholdersPaid; iterator.Next() {
		skipped++
	}

	// Process up to batchSize shareholders
	for ; iterator.Valid() && paymentsProcessed < batchSize; iterator.Next() {
		var shareholderSnapshot types.ShareholderSnapshot
		if err := json.Unmarshal(iterator.Value(), &shareholderSnapshot); err != nil {
			k.Logger(ctx).Error("failed to unmarshal shareholder snapshot", "error", err)
			continue
		}

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

		// Check if this is a treasury dividend recipient (cross-company investment)
		// Format: "treasury_dividend:ownerCompanyID"
		if strings.HasPrefix(shareholderSnapshot.Shareholder, "treasury_dividend:") {
			// Parse treasury owner company ID
			parts := strings.Split(shareholderSnapshot.Shareholder, ":")
			if len(parts) == 2 {
				ownerCompanyID, err := strconv.ParseUint(parts[1], 10, 64)
				if err == nil {
					// Credit dividend to treasury balance instead of sending to address
					if dividend.Type == types.DividendTypeCash {
						paymentCoins := sdk.NewCoins(sdk.NewCoin(dividend.Currency, netAmount.TruncateInt()))
						if err := k.CreditTreasuryDividend(ctx, ownerCompanyID, dividend.CompanyID, paymentCoins); err != nil {
							payment.Status = "failed"
							payment.FailureReason = fmt.Sprintf("treasury credit failed: %v", err)
						} else {
							payment.Status = "paid_to_treasury"
							payment.PaidAt = ctx.BlockTime()
							totalPaid = totalPaid.Add(grossAmount)
							paymentsProcessed++

							k.Logger(ctx).Info("dividend paid to treasury",
								"dividend_id", dividendID,
								"recipient_treasury", ownerCompanyID,
								"amount", netAmount.String(),
							)
						}
					} else {
						// Stock dividends to treasuries - add to investment holdings
						payment.Status = "skipped"
						payment.FailureReason = "stock_dividend_to_treasury_not_supported"
					}
					if err := k.SetDividendPayment(ctx, payment); err != nil {
						k.Logger(ctx).Error("failed to set dividend payment", "error", err)
					}
					continue
				}
			}
		}

		// Check if shareholder is blacklisted
		var recipientAddr sdk.AccAddress
		var recipientAddrStr string
		var isRedirected bool

		if k.IsShareholderBlacklisted(ctx, dividend.CompanyID, shareholderSnapshot.Shareholder) {
			// Get redirection settings
			redirect, hasRedirect := k.GetDividendRedirection(ctx, dividend.CompanyID)

			if hasRedirect && redirect.CharityWallet != "" {
				recipientAddrStr = redirect.CharityWallet
			} else {
				// Use default charity wallet or fallback
				recipientAddrStr = k.GetDefaultCharityWallet(ctx)
				if recipientAddrStr == "" {
					// Final fallback: community pool via governance module
					recipientAddrStr = k.GetCommunityPoolAddress(ctx)
				}
			}

			isRedirected = true
			payment.Status = "redirected"
			payment.FailureReason = "shareholder_blacklisted"
		} else {
			recipientAddrStr = shareholderSnapshot.Shareholder
			isRedirected = false
		}

		// Process payment based on dividend type
		if dividend.Type == types.DividendTypeCash {
			// Transfer funds from module to recipient (shareholder or charity)
			var err error
			recipientAddr, err = sdk.AccAddressFromBech32(recipientAddrStr)
			if err != nil {
				payment.Status = "failed"
				payment.FailureReason = "invalid address"
			} else {
				paymentCoins := sdk.NewCoins(sdk.NewCoin(dividend.Currency, netAmount.TruncateInt()))
				if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, paymentCoins); err != nil {
					payment.Status = "failed"
					payment.FailureReason = err.Error()
				} else {
					if !isRedirected {
						payment.Status = "paid"
					}
					payment.PaidAt = ctx.BlockTime()
					totalPaid = totalPaid.Add(grossAmount)
					paymentsProcessed++

					// Emit redirection event if applicable
					if isRedirected {
						ctx.EventManager().EmitEvent(
							sdk.NewEvent(
								types.EventTypeDividendRedirected,
								sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", dividend.CompanyID)),
								sdk.NewAttribute(types.AttributeKeyShareholder, shareholderSnapshot.Shareholder),
								sdk.NewAttribute(types.AttributeKeyRedirectTo, recipientAddrStr),
								sdk.NewAttribute(types.AttributeKeyRedirectAmount, netAmount.String()),
								sdk.NewAttribute(types.AttributeKeyRedirectReason, "shareholder_blacklisted"),
							),
						)
					}
				}
			}
		} else if dividend.Type == types.DividendTypeStock {
			// For stock dividends, issue to original shareholder or skip if blacklisted
			// (Stock dividends cannot be easily redirected to charity)
			if isRedirected {
				payment.Status = "skipped"
				payment.FailureReason = "blacklisted_shareholder_stock_dividend"
			} else {
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
		}

		// Store payment record
		if err := k.SetDividendPayment(ctx, payment); err != nil {
			k.Logger(ctx).Error("failed to set dividend payment", "error", err)
			continue
		}
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

	if err := k.SetDividend(ctx, dividend); err != nil {
		return 0, math.LegacyZeroDec(), false, err
	}

	return paymentsProcessed, totalPaid, isComplete, nil
}

// ClaimDividend allows a shareholder to claim their dividend
func (k Keeper) ClaimDividend(ctx sdk.Context, claimant string, dividendID uint64) (math.LegacyDec, string, math.LegacyDec, math.LegacyDec, error) {
	// Get dividend
	dividend, found := k.GetDividend(ctx, dividendID)
	if !found {
		return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), types.ErrDividendNotFound
	}

	// CRITICAL: Check audit verification status before allowing claims
	if dividend.AuditID > 0 {
		audit, found := k.GetDividendAudit(ctx, dividend.AuditID)
		if !found {
			return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), types.ErrAuditNotFound
		}

		// Audit must be verified before dividends can be claimed
		if audit.Status != types.AuditStatusVerified {
			return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), types.ErrAuditNotVerified
		}
	} else {
		// No audit linked - this should not happen for new dividends
		k.Logger(ctx).Warn("dividend has no audit linked - old dividend or missing audit",
			"dividend_id", dividendID,
		)
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
	if err := k.SetDividendPayment(ctx, payment); err != nil {
		return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), err
	}

	// Update dividend totals
	dividend.PaidAmount = dividend.PaidAmount.Add(grossAmount)
	dividend.RemainingAmount = dividend.TotalAmount.Sub(dividend.PaidAmount)
	dividend.ShareholdersPaid++
	dividend.UpdatedAt = ctx.BlockTime()
	if err := k.SetDividend(ctx, dividend); err != nil {
		return math.LegacyZeroDec(), "", math.LegacyZeroDec(), math.LegacyZeroDec(), err
	}
	
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

// GetAllDividendsByStatus returns all dividends with a specific status
func (k Keeper) GetAllDividendsByStatus(ctx sdk.Context, status types.DividendStatus) []types.Dividend {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DividendPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	dividends := []types.Dividend{}
	for ; iterator.Valid(); iterator.Next() {
		var dividend types.Dividend
		if err := json.Unmarshal(iterator.Value(), &dividend); err == nil && dividend.Status == status {
			dividends = append(dividends, dividend)
		}
	}

	return dividends
}

// GetAllDividends returns all dividends
func (k Keeper) GetAllDividends(ctx sdk.Context) []types.Dividend {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DividendPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	dividends := []types.Dividend{}
	for ; iterator.Valid(); iterator.Next() {
		var dividend types.Dividend
		if err := json.Unmarshal(iterator.Value(), &dividend); err == nil {
			dividends = append(dividends, dividend)
		}
	}

	return dividends
}