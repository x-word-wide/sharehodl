package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// =============================================================================
// Primary Sale Offer CRUD
// =============================================================================

// GetNextOfferID returns the next offer ID and increments the counter
func (k Keeper) GetNextOfferID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PrimarySaleOfferCounterKey)

	var counter uint64 = 1 // Start from 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	// Increment counter for next use
	store.Set(types.PrimarySaleOfferCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// SetPrimarySaleOffer stores a primary sale offer
func (k Keeper) SetPrimarySaleOffer(ctx sdk.Context, offer types.PrimarySaleOffer) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPrimarySaleOfferKey(offer.ID)
	bz, err := json.Marshal(offer)
	if err != nil {
		return fmt.Errorf("failed to marshal primary sale offer: %w", err)
	}
	store.Set(key, bz)

	// Update company index
	k.setOfferByCompanyIndex(ctx, offer.CompanyID, offer.ID)
	return nil
}

// GetPrimarySaleOffer returns a primary sale offer by ID
func (k Keeper) GetPrimarySaleOffer(ctx sdk.Context, offerID uint64) (types.PrimarySaleOffer, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPrimarySaleOfferKey(offerID)
	bz := store.Get(key)
	if bz == nil {
		return types.PrimarySaleOffer{}, false
	}

	var offer types.PrimarySaleOffer
	if err := json.Unmarshal(bz, &offer); err != nil {
		return types.PrimarySaleOffer{}, false
	}
	return offer, true
}

// GetOffersByCompany returns all offers for a company
func (k Keeper) GetOffersByCompany(ctx sdk.Context, companyID uint64) []types.PrimarySaleOffer {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetOffersByCompanyPrefix(companyID))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	offers := []types.PrimarySaleOffer{}
	for ; iterator.Valid(); iterator.Next() {
		offerID := sdk.BigEndianToUint64(iterator.Value())
		if offer, found := k.GetPrimarySaleOffer(ctx, offerID); found {
			offers = append(offers, offer)
		}
	}

	return offers
}

// GetActiveOfferBySymbol returns the active offer for a company symbol (for DEX integration)
func (k Keeper) GetActiveOfferBySymbol(ctx sdk.Context, symbol string) (types.PrimarySaleOffer, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetOfferBySymbolKey(symbol)
	bz := store.Get(key)
	if bz == nil {
		return types.PrimarySaleOffer{}, false
	}

	offerID := sdk.BigEndianToUint64(bz)
	return k.GetPrimarySaleOffer(ctx, offerID)
}

// GetAllActiveOffers returns all currently active offers
func (k Keeper) GetAllActiveOffers(ctx sdk.Context) []types.PrimarySaleOffer {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ActiveOffersPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	offers := []types.PrimarySaleOffer{}
	for ; iterator.Valid(); iterator.Next() {
		offerID := sdk.BigEndianToUint64(iterator.Key()[len(types.ActiveOffersPrefix):])
		if offer, found := k.GetPrimarySaleOffer(ctx, offerID); found {
			if offer.IsActive(ctx.BlockTime()) {
				offers = append(offers, offer)
			}
		}
	}

	return offers
}

// =============================================================================
// Primary Sale Offer Lifecycle
// =============================================================================

// CreatePrimarySaleOffer creates a new share offering (founder only)
func (k Keeper) CreatePrimarySaleOffer(
	ctx sdk.Context,
	companyID uint64,
	classID string,
	sharesOffered math.Int,
	pricePerShare math.LegacyDec,
	quoteDenom string,
	minPurchase math.Int,
	maxPerAddress math.Int,
	startTime time.Time,
	endTime time.Time,
	creator string,
) (uint64, error) {
	// Validate creator address
	if _, err := sdk.AccAddressFromBech32(creator); err != nil {
		return 0, types.ErrUnauthorized
	}

	// Get company and verify creator can propose
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return 0, types.ErrCompanyNotFound
	}

	// Use authorization system: creator must be owner or authorized proposer
	if !k.CanProposeForCompany(ctx, companyID, creator) {
		return 0, types.ErrNotAuthorizedProposer
	}

	// Verify company is active
	if company.Status != types.CompanyStatusActive {
		return 0, types.ErrCompanyNotActive
	}

	// Get share class and validate
	shareClass, found := k.getShareClass(ctx, companyID, classID)
	if !found {
		return 0, types.ErrShareClassNotFound
	}

	// CRITICAL: Check if treasury has enough shares to offer
	// Shares must be in treasury before they can be sold
	treasury, found := k.GetCompanyTreasury(ctx, companyID)
	if !found {
		return 0, types.ErrTreasuryNotFound
	}

	treasuryShares := treasury.GetTreasuryShares(classID)
	if treasuryShares.LT(sharesOffered) {
		return 0, types.ErrInsufficientTreasuryShares
	}

	// Check if shares can be issued (not exceeding authorized)
	if shareClass.IssuedShares.Add(sharesOffered).GT(shareClass.AuthorizedShares) {
		return 0, types.ErrExceedsAuthorized
	}

	// Check for existing active offer for this share class
	existingOffers := k.GetOffersByCompany(ctx, companyID)
	for _, offer := range existingOffers {
		if offer.ClassID == classID && offer.Status == types.OfferStatusOpen {
			return 0, types.ErrPrimarySaleAlreadyExists
		}
	}

	// Validate time parameters
	if endTime.Before(startTime) {
		return 0, fmt.Errorf("end time must be after start time")
	}

	if startTime.Before(ctx.BlockTime()) {
		return 0, fmt.Errorf("start time cannot be in the past")
	}

	// Validate amounts
	if sharesOffered.IsNil() || sharesOffered.LTE(math.ZeroInt()) {
		return 0, fmt.Errorf("shares offered must be positive")
	}

	if pricePerShare.IsNil() || pricePerShare.LTE(math.LegacyZeroDec()) {
		return 0, types.ErrInvalidOfferPrice
	}

	// Get next offer ID
	offerID := k.GetNextOfferID(ctx)

	// Create offer
	offer := types.NewPrimarySaleOffer(
		offerID,
		companyID,
		classID,
		sharesOffered,
		pricePerShare,
		quoteDenom,
		minPurchase,
		maxPerAddress,
		startTime,
		endTime,
		creator,
	)

	// Validate
	if err := offer.Validate(); err != nil {
		return 0, err
	}

	// Store offer
	if err := k.SetPrimarySaleOffer(ctx, offer); err != nil {
		return 0, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePrimarySaleCreated,
			sdk.NewAttribute(types.AttributeKeyOfferID, fmt.Sprintf("%d", offerID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyOfferClassID, classID),
			sdk.NewAttribute(types.AttributeKeySharesOffered, sharesOffered.String()),
			sdk.NewAttribute(types.AttributeKeyOfferPrice, pricePerShare.String()),
			sdk.NewAttribute(types.AttributeKeyOfferStartTime, startTime.Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyOfferEndTime, endTime.Format(time.RFC3339)),
		),
	)

	k.Logger(ctx).Info("primary sale offer created",
		"offer_id", offerID,
		"company_id", companyID,
		"class_id", classID,
		"shares_offered", sharesOffered.String(),
		"price", pricePerShare.String(),
	)

	return offerID, nil
}

// OpenPrimarySaleOffer activates an offer for trading
func (k Keeper) OpenPrimarySaleOffer(ctx sdk.Context, offerID uint64, opener string) error {
	// Get offer
	offer, found := k.GetPrimarySaleOffer(ctx, offerID)
	if !found {
		return types.ErrPrimarySaleNotFound
	}

	// Get company and verify opener is founder
	company, found := k.getCompany(ctx, offer.CompanyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	if company.Founder != opener {
		return types.ErrUnauthorized
	}

	// Check offer is in draft status
	if offer.Status != types.OfferStatusDraft {
		return fmt.Errorf("offer is not in draft status")
	}

	// Update status
	offer.Status = types.OfferStatusOpen
	if err := k.SetPrimarySaleOffer(ctx, offer); err != nil {
		return err
	}

	// Add to active offers index
	k.SetActiveOfferIndex(ctx, offerID, company.Symbol)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePrimarySaleOpened,
			sdk.NewAttribute(types.AttributeKeyOfferID, fmt.Sprintf("%d", offerID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", offer.CompanyID)),
			sdk.NewAttribute(types.AttributeKeyOfferStatus, offer.Status.String()),
		),
	)

	k.Logger(ctx).Info("primary sale offer opened",
		"offer_id", offerID,
		"company_id", offer.CompanyID,
	)

	return nil
}

// ClosePrimarySaleOffer closes an offer (manually or auto when sold out)
func (k Keeper) ClosePrimarySaleOffer(ctx sdk.Context, offerID uint64, closer string) error {
	// Get offer
	offer, found := k.GetPrimarySaleOffer(ctx, offerID)
	if !found {
		return types.ErrPrimarySaleNotFound
	}

	// Get company
	company, found := k.getCompany(ctx, offer.CompanyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Verify closer is founder (if manually closed)
	if closer != "" && company.Founder != closer {
		return types.ErrUnauthorized
	}

	// Check offer is open
	if offer.Status != types.OfferStatusOpen {
		return fmt.Errorf("offer is not open")
	}

	// Update status
	offer.Status = types.OfferStatusClosed
	if err := k.SetPrimarySaleOffer(ctx, offer); err != nil {
		return err
	}

	// Remove from active offers index
	k.RemoveActiveOfferIndex(ctx, offerID, company.Symbol)

	// Calculate total revenue
	totalRevenue := offer.PricePerShare.MulInt(offer.SharesSold)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePrimarySaleClosed,
			sdk.NewAttribute(types.AttributeKeyOfferID, fmt.Sprintf("%d", offerID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", offer.CompanyID)),
			sdk.NewAttribute(types.AttributeKeySharesSold, offer.SharesSold.String()),
			sdk.NewAttribute(types.AttributeKeyTotalRevenue, totalRevenue.String()),
			sdk.NewAttribute(types.AttributeKeyOfferStatus, offer.Status.String()),
		),
	)

	k.Logger(ctx).Info("primary sale offer closed",
		"offer_id", offerID,
		"company_id", offer.CompanyID,
		"shares_sold", offer.SharesSold.String(),
		"total_revenue", totalRevenue.String(),
	)

	return nil
}

// CancelPrimarySaleOffer cancels an offer (founder only, before any sales)
func (k Keeper) CancelPrimarySaleOffer(ctx sdk.Context, offerID uint64, canceller string) error {
	// Get offer
	offer, found := k.GetPrimarySaleOffer(ctx, offerID)
	if !found {
		return types.ErrPrimarySaleNotFound
	}

	// Get company and verify canceller is founder
	company, found := k.getCompany(ctx, offer.CompanyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	if company.Founder != canceller {
		return types.ErrUnauthorized
	}

	// Can only cancel if no shares sold yet
	if offer.SharesSold.GT(math.ZeroInt()) {
		return fmt.Errorf("cannot cancel offer with shares already sold")
	}

	// Can only cancel draft or open offers
	if offer.Status != types.OfferStatusDraft && offer.Status != types.OfferStatusOpen {
		return fmt.Errorf("offer cannot be cancelled in current status")
	}

	// Update status
	offer.Status = types.OfferStatusCancelled
	if err := k.SetPrimarySaleOffer(ctx, offer); err != nil {
		return err
	}

	// Remove from active offers if it was open
	if offer.Status == types.OfferStatusOpen {
		k.RemoveActiveOfferIndex(ctx, offerID, company.Symbol)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePrimarySaleClosed,
			sdk.NewAttribute(types.AttributeKeyOfferID, fmt.Sprintf("%d", offerID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", offer.CompanyID)),
			sdk.NewAttribute(types.AttributeKeyOfferStatus, "cancelled"),
		),
	)

	k.Logger(ctx).Info("primary sale offer cancelled",
		"offer_id", offerID,
		"company_id", offer.CompanyID,
	)

	return nil
}

// =============================================================================
// Primary Sale Purchase (CRITICAL - funds go to treasury)
// =============================================================================

// BuyFromPrimarySale executes a purchase from primary offering
// - Issues NEW shares to buyer
// - Routes payment to company TREASURY (not seller)
// - Updates offer.SharesSold counter
func (k Keeper) BuyFromPrimarySale(
	ctx sdk.Context,
	offerID uint64,
	buyer string,
	shareQuantity math.Int,
) (math.Int, error) {
	// Validate buyer address
	buyerAddr, err := sdk.AccAddressFromBech32(buyer)
	if err != nil {
		return math.ZeroInt(), types.ErrUnauthorized
	}

	// Get offer
	offer, found := k.GetPrimarySaleOffer(ctx, offerID)
	if !found {
		return math.ZeroInt(), types.ErrPrimarySaleNotFound
	}

	// Validate purchase
	if err := k.ValidatePurchase(ctx, offer, buyer, shareQuantity, ctx.BlockTime()); err != nil {
		return math.ZeroInt(), err
	}

	// Calculate purchase cost (ceiling to handle decimal prices)
	purchaseCostDec := offer.PricePerShare.MulInt(shareQuantity)
	purchaseCost := purchaseCostDec.Ceil().TruncateInt()

	// Check buyer has sufficient balance
	balance := k.bankKeeper.GetBalance(ctx, buyerAddr, offer.QuoteDenom)
	if balance.Amount.LT(purchaseCost) {
		return math.ZeroInt(), types.ErrInsufficientFunds
	}

	// CRITICAL: Transfer shares from treasury to buyer
	// This function handles:
	// - Payment transfer (buyer -> module/treasury)
	// - Treasury balance update (increase HODL balance)
	// - Treasury share reduction (decrease treasury shares)
	// - Share issuance to buyer
	if err := k.TransferSharesFromTreasury(ctx, offer.CompanyID, offer.ClassID, buyer, shareQuantity, offer.PricePerShare, offer.QuoteDenom); err != nil {
		return math.ZeroInt(), fmt.Errorf("failed to transfer shares from treasury: %w", err)
	}

	// Update offer sold counter
	offer.SharesSold = offer.SharesSold.Add(shareQuantity)
	if err := k.SetPrimarySaleOffer(ctx, offer); err != nil {
		return math.ZeroInt(), err
	}

	// Track buyer purchase amount for MaxPerAddress enforcement
	if err := k.addBuyerPurchaseAmount(ctx, offerID, buyer, shareQuantity); err != nil {
		return math.ZeroInt(), err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePrimarySalePurchase,
			sdk.NewAttribute(types.AttributeKeyOfferID, fmt.Sprintf("%d", offerID)),
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", offer.CompanyID)),
			sdk.NewAttribute(types.AttributeKeyOfferClassID, offer.ClassID),
			sdk.NewAttribute(types.AttributeKeyOfferBuyer, buyer),
			sdk.NewAttribute(types.AttributeKeyOfferSharesPurchased, shareQuantity.String()),
			sdk.NewAttribute(types.AttributeKeyPurchaseCost, purchaseCost.String()),
			sdk.NewAttribute(types.AttributeKeySharesSold, offer.SharesSold.String()),
			sdk.NewAttribute(types.AttributeKeySharesRemaining, offer.RemainingShares().String()),
		),
	)

	k.Logger(ctx).Info("primary sale purchase completed",
		"offer_id", offerID,
		"company_id", offer.CompanyID,
		"buyer", buyer,
		"shares", shareQuantity.String(),
		"cost", purchaseCost.String(),
	)

	// Auto-close if sold out
	if offer.RemainingShares().IsZero() {
		if err := k.ClosePrimarySaleOffer(ctx, offerID, ""); err != nil {
			k.Logger(ctx).Error("failed to auto-close sold out offer",
				"offer_id", offerID,
				"error", err,
			)
		}
	}

	return purchaseCost, nil
}

// ValidatePurchase checks if purchase is valid
func (k Keeper) ValidatePurchase(
	ctx sdk.Context,
	offer types.PrimarySaleOffer,
	buyer string,
	quantity math.Int,
	currentTime time.Time,
) error {
	// Check offer is active
	if !offer.IsActive(currentTime) {
		if offer.Status != types.OfferStatusOpen {
			return types.ErrPrimarySaleNotActive
		}
		if currentTime.Before(offer.StartTime) {
			return types.ErrOfferNotStarted
		}
		if currentTime.After(offer.EndTime) {
			return types.ErrPrimarySaleEnded
		}
		if offer.RemainingShares().IsZero() {
			return types.ErrPrimarySaleSoldOut
		}
		return types.ErrPrimarySaleNotActive
	}

	// Validate quantity
	if quantity.IsNil() || quantity.LTE(math.ZeroInt()) {
		return fmt.Errorf("purchase quantity must be positive")
	}

	// Check minimum purchase
	if !offer.MinPurchase.IsZero() && quantity.LT(offer.MinPurchase) {
		return types.ErrBelowMinPurchase
	}

	// Check if enough shares available
	if quantity.GT(offer.RemainingShares()) {
		return types.ErrInsufficientSharesAvailable
	}

	// Check maximum per address (if set)
	if !offer.MaxPerAddress.IsZero() {
		currentPurchases := k.GetBuyerPurchaseAmount(ctx, offer.ID, buyer)
		totalAfterPurchase := currentPurchases.Add(quantity)
		if totalAfterPurchase.GT(offer.MaxPerAddress) {
			return types.ErrExceedsMaxPurchase
		}
	}

	return nil
}

// GetBuyerPurchaseAmount returns how many shares buyer has purchased from this offer
func (k Keeper) GetBuyerPurchaseAmount(ctx sdk.Context, offerID uint64, buyer string) math.Int {
	store := ctx.KVStore(k.storeKey)
	key := types.GetBuyerPurchaseKey(offerID, buyer)
	bz := store.Get(key)
	if bz == nil {
		return math.ZeroInt()
	}

	// Deserialize the math.Int from bytes
	var amount math.Int
	err := amount.Unmarshal(bz)
	if err != nil {
		// If deserialization fails, return zero
		return math.ZeroInt()
	}
	return amount
}

// addBuyerPurchaseAmount adds to the buyer's purchase tracking
func (k Keeper) addBuyerPurchaseAmount(ctx sdk.Context, offerID uint64, buyer string, amount math.Int) error {
	currentAmount := k.GetBuyerPurchaseAmount(ctx, offerID, buyer)
	newAmount := currentAmount.Add(amount)

	store := ctx.KVStore(k.storeKey)
	key := types.GetBuyerPurchaseKey(offerID, buyer)

	// Serialize the math.Int to bytes
	bz, err := newAmount.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal buyer purchase amount: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// =============================================================================
// Index Management
// =============================================================================

// SetActiveOfferIndex adds offer to active offers index
func (k Keeper) SetActiveOfferIndex(ctx sdk.Context, offerID uint64, symbol string) {
	store := ctx.KVStore(k.storeKey)

	// Add to general active offers index
	activeKey := types.GetActiveOfferKey(offerID)
	store.Set(activeKey, []byte{0x01})

	// Add to symbol index (for DEX lookup)
	symbolKey := types.GetOfferBySymbolKey(symbol)
	store.Set(symbolKey, sdk.Uint64ToBigEndian(offerID))
}

// RemoveActiveOfferIndex removes offer from active index
func (k Keeper) RemoveActiveOfferIndex(ctx sdk.Context, offerID uint64, symbol string) {
	store := ctx.KVStore(k.storeKey)

	// Remove from active offers index
	activeKey := types.GetActiveOfferKey(offerID)
	store.Delete(activeKey)

	// Remove from symbol index
	symbolKey := types.GetOfferBySymbolKey(symbol)
	store.Delete(symbolKey)
}

// setOfferByCompanyIndex sets the company->offer index
func (k Keeper) setOfferByCompanyIndex(ctx sdk.Context, companyID uint64, offerID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetOfferByCompanyKey(companyID, offerID)
	store.Set(key, sdk.Uint64ToBigEndian(offerID))
}

// =============================================================================
// Helper Functions
// =============================================================================

// IsOfferActive checks if offer is open and within time window
func (k Keeper) IsOfferActive(ctx sdk.Context, offerID uint64) bool {
	offer, found := k.GetPrimarySaleOffer(ctx, offerID)
	if !found {
		return false
	}
	return offer.IsActive(ctx.BlockTime())
}

// GetRemainingShares returns unsold shares in offer
func (k Keeper) GetRemainingShares(ctx sdk.Context, offerID uint64) math.Int {
	offer, found := k.GetPrimarySaleOffer(ctx, offerID)
	if !found {
		return math.ZeroInt()
	}
	return offer.RemainingShares()
}

// ProcessExpiredOffers closes offers past their end time (called in EndBlock)
func (k Keeper) ProcessExpiredOffers(ctx sdk.Context) {
	currentTime := ctx.BlockTime()
	activeOffers := k.GetAllActiveOffers(ctx)

	for _, offer := range activeOffers {
		if currentTime.After(offer.EndTime) {
			k.Logger(ctx).Info("auto-closing expired offer",
				"offer_id", offer.ID,
				"company_id", offer.CompanyID,
				"end_time", offer.EndTime,
			)

			if err := k.ClosePrimarySaleOffer(ctx, offer.ID, ""); err != nil {
				k.Logger(ctx).Error("failed to auto-close expired offer",
					"offer_id", offer.ID,
					"error", err,
				)
			}
		}
	}
}

// =============================================================================
// Treasury Integration (CRITICAL)
// =============================================================================

// depositToTreasury routes primary sale proceeds to company treasury
// This is the CRITICAL function that ensures funds go to treasury, not a personal wallet
func (k Keeper) depositToTreasury(ctx sdk.Context, companyID uint64, amount sdk.Coins, depositor string) error {
	// Get or create treasury
	treasury, found := k.getCompanyTreasury(ctx, companyID)
	if !found {
		// Create new treasury if it doesn't exist
		treasury = types.NewCompanyTreasury(companyID)
	}

	// Check if treasury is locked
	if treasury.IsLocked {
		return types.ErrTreasuryLocked
	}

	// Get company treasury module account
	treasuryAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if treasuryAddr == nil {
		return fmt.Errorf("treasury module account not found")
	}

	// Funds are already in module account from buyer->module transfer
	// We just update the treasury balance tracking

	// Update treasury balance
	treasury.Balance = treasury.Balance.Add(amount...)
	treasury.TotalDeposited = treasury.TotalDeposited.Add(amount...)
	treasury.UpdatedAt = ctx.BlockTime()

	// Save treasury
	if err := k.setCompanyTreasury(ctx, treasury); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTreasuryDeposit,
			sdk.NewAttribute(types.AttributeKeyTreasuryCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyDepositAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyDepositor, depositor),
			sdk.NewAttribute(types.AttributeKeyTreasuryBalance, treasury.Balance.String()),
		),
	)

	k.Logger(ctx).Info("primary sale proceeds deposited to treasury",
		"company_id", companyID,
		"amount", amount.String(),
		"depositor", depositor,
		"new_balance", treasury.Balance.String(),
	)

	return nil
}

// getCompanyTreasury retrieves a company treasury
func (k Keeper) getCompanyTreasury(ctx sdk.Context, companyID uint64) (types.CompanyTreasury, bool) {
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

// setCompanyTreasury stores a company treasury
func (k Keeper) setCompanyTreasury(ctx sdk.Context, treasury types.CompanyTreasury) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyTreasuryKey(treasury.CompanyID)
	bz, err := json.Marshal(treasury)
	if err != nil {
		return fmt.Errorf("failed to marshal company treasury: %w", err)
	}
	store.Set(key, bz)
	return nil
}
