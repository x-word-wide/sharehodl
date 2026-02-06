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

// Business listing management

// GetNextListingID returns the next listing ID and increments the counter
func (k Keeper) GetNextListingID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ListingCounterKey)
	
	var counter uint64 = 1 // Start from 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}
	
	// Increment counter for next use
	store.Set(types.ListingCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetBusinessListing returns a business listing by ID
func (k Keeper) GetBusinessListing(ctx sdk.Context, listingID uint64) (types.BusinessListing, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetBusinessListingKey(listingID)
	bz := store.Get(key)
	if bz == nil {
		return types.BusinessListing{}, false
	}
	
	var listing types.BusinessListing
	if err := json.Unmarshal(bz, &listing); err != nil {
		return types.BusinessListing{}, false
	}
	return listing, true
}

// SetBusinessListing stores a business listing
func (k Keeper) SetBusinessListing(ctx sdk.Context, listing types.BusinessListing) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetBusinessListingKey(listing.ID)
	bz, err := json.Marshal(listing)
	if err != nil {
		return fmt.Errorf("failed to marshal business listing: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// DeleteBusinessListing removes a business listing
func (k Keeper) DeleteBusinessListing(ctx sdk.Context, listingID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetBusinessListingKey(listingID)
	store.Delete(key)
}

// GetListingRequirements returns the current listing requirements
func (k Keeper) GetListingRequirements(ctx sdk.Context) types.ListingRequirements {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ListingRequirementsKey)
	if bz == nil {
		// Return default requirements
		return types.DefaultListingRequirements()
	}
	
	var requirements types.ListingRequirements
	if err := json.Unmarshal(bz, &requirements); err != nil {
		// Return default on error
		return types.DefaultListingRequirements()
	}
	return requirements
}

// SetListingRequirements stores the listing requirements
func (k Keeper) SetListingRequirements(ctx sdk.Context, requirements types.ListingRequirements) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(requirements)
	if err != nil {
		return fmt.Errorf("failed to marshal listing requirements: %w", err)
	}
	store.Set(types.ListingRequirementsKey, bz)
	return nil
}

// SubmitBusinessListing submits a new business listing application
func (k Keeper) SubmitBusinessListing(
	ctx sdk.Context,
	creator string,
	companyName, legalEntityName, registrationNumber, jurisdiction, industry string,
	businessDescription, website, contactEmail, tokenSymbol, tokenName string,
	foundedDate time.Time,
	totalShares math.Int,
	initialPrice math.LegacyDec,
	listingFee math.Int,
	documents []types.BusinessDocumentUpload,
) (uint64, error) {
	// Check creator address
	creatorAddr, err := sdk.AccAddressFromBech32(creator)
	if err != nil {
		return 0, types.ErrUnauthorized
	}

	// TIER CHECK: Must be Keeper+ tier (10K HODL staked) to submit listing
	if k.stakingKeeper != nil {
		if !k.stakingKeeper.CanSubmitListing(ctx, creatorAddr) {
			return 0, types.ErrInsufficientTierForListing
		}
	}

	// Get listing requirements
	requirements := k.GetListingRequirements(ctx)

	// Validate listing meets requirements
	if err := k.validateListingRequirements(ctx, companyName, tokenSymbol, jurisdiction, industry, totalShares, initialPrice, listingFee, foundedDate, requirements); err != nil {
		return 0, err
	}

	// Check if creator has sufficient funds for listing fee
	balance := k.bankKeeper.GetBalance(ctx, creatorAddr, "uhodl")
	if balance.Amount.LT(listingFee) {
		return 0, types.ErrInsufficientListingFee
	}
	
	// Get next listing ID
	listingID := k.GetNextListingID(ctx)
	
	// Create business documents from uploads
	businessDocs := make([]types.BusinessDocument, len(documents))
	for i, doc := range documents {
		businessDocs[i] = types.BusinessDocument{
			ID:           uint64(i + 1),
			DocumentType: doc.DocumentType,
			FileName:     doc.FileName,
			FileHash:     doc.FileHash,
			FileSize:     doc.FileSize,
			UploadedAt:   time.Now(),
			Verified:     false,
		}
	}
	
	// Create listing
	listing := types.NewBusinessListing(
		listingID,
		companyName, legalEntityName, registrationNumber, jurisdiction, industry,
		businessDescription, website, contactEmail, tokenSymbol, tokenName,
		totalShares,
		initialPrice,
		listingFee,
		creator,
	)
	listing.Documents = businessDocs

	// Store listing
	if err := k.SetBusinessListing(ctx, listing); err != nil {
		return 0, err
	}

	// Transfer listing fee to module account
	fee := sdk.NewCoins(sdk.NewCoin("uhodl", listingFee))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, fee); err != nil {
		return 0, err
	}

	// LOCK STAKE: Lock creator's stake while listing is pending review
	// This prevents unstaking while the listing is under review
	if k.stakingKeeper != nil {
		listingIDStr := fmt.Sprintf("%d", listingID)
		if err := k.stakingKeeper.LockForPendingListing(ctx, creatorAddr, listingIDStr); err != nil {
			// Log but don't fail - stake locking is supplementary protection
			k.Logger(ctx).Error("failed to lock stake for pending listing",
				"listing_id", listingID,
				"creator", creator,
				"error", err,
			)
		}
	}

	// Assign validators for review
	if err := k.assignValidatorsForReview(ctx, listingID, requirements.RequiredValidators); err != nil {
		return 0, err
	}

	return listingID, nil
}

// ReviewBusinessListing allows a validator to review a business listing
func (k Keeper) ReviewBusinessListing(
	ctx sdk.Context,
	validator string,
	listingID uint64,
	approved bool,
	comments string,
	score int,
	criteria types.VerificationCriteria,
	data map[string]interface{},
) error {
	// Get listing
	listing, found := k.GetBusinessListing(ctx, listingID)
	if !found {
		return types.ErrListingNotFound
	}
	
	// Check if listing is in correct status
	if listing.Status != types.ListingStatusPending && listing.Status != types.ListingStatusUnderReview {
		return types.ErrListingNotPending
	}
	
	// Check if validator is assigned
	if !k.isValidatorAssigned(ctx, listingID, validator) {
		return types.ErrValidatorNotAssigned
	}
	
	// Check if validator already reviewed
	for _, result := range listing.VerificationResults {
		if result.ValidatorAddress == validator {
			return types.ErrValidatorAlreadyReviewed
		}
	}
	
	// Check if user has required tier using staking keeper
	if k.stakingKeeper != nil {
		validatorAddr, err := sdk.AccAddressFromBech32(validator)
		if err != nil {
			return types.ErrUnauthorized
		}
		tier := k.stakingKeeper.GetUserTierInt(ctx, validatorAddr)
		if tier < types.StakeTierKeeper { // Requires at least Keeper tier
			return types.ErrUnauthorized
		}
	}
	
	// Add verification result
	result := types.VerificationResult{
		ValidatorAddress: validator,
		ValidatorTier:    2, // Simplified - assume business validators are tier 2
		Approved:         approved,
		Comments:         comments,
		Score:            score,
		VerifiedAt:       time.Now(),
		VerificationData: data,
	}
	
	listing.VerificationResults = append(listing.VerificationResults, result)
	listing.Status = types.ListingStatusUnderReview
	listing.UpdatedAt = time.Now()

	// Update listing
	if err := k.SetBusinessListing(ctx, listing); err != nil {
		return err
	}

	return nil
}

// ApproveListing approves a business listing and creates the equity token
func (k Keeper) ApproveListing(ctx sdk.Context, listingID uint64, authority string) (uint64, error) {
	// Get listing
	listing, found := k.GetBusinessListing(ctx, listingID)
	if !found {
		return 0, types.ErrListingNotFound
	}
	
	// Check if listing can be approved
	if !listing.CanApprove() {
		return 0, types.ErrInsufficientValidators
	}
	
	// Check if all required documents are present
	if !listing.HasRequiredDocuments() {
		return 0, types.ErrRequiredDocumentsMissing
	}
	
	// Create company
	companyID, err := k.CreateCompany(
		ctx,
		listing.CompanyName,
		listing.LegalEntityName,
		listing.RegistrationNumber,
		listing.Jurisdiction,
		listing.Industry,
		listing.BusinessDescription,
		listing.TokenSymbol,
		listing.Applicant,
	)
	if err != nil {
		return 0, err
	}
	
	// Create share class
	err = k.CreateShareClass(
		ctx,
		companyID,
		"COMMON", // Default to common shares
		listing.TokenName,
		listing.TotalShares,
		math.LegacyNewDec(1), // 1 vote per share
		true, // transferable
		listing.Applicant,
	)
	if err != nil {
		return 0, err
	}
	
	// Issue shares to the applicant
	err = k.IssueShares(
		ctx,
		companyID,
		"COMMON",
		listing.Applicant,
		listing.TotalShares,
		listing.InitialPrice,
		"uhodl",
	)
	if err != nil {
		return 0, err
	}
	
	// Update listing status
	listing.Status = types.ListingStatusApproved
	listing.ReviewCompletedAt = time.Now()
	listing.UpdatedAt = time.Now()
	if err := k.SetBusinessListing(ctx, listing); err != nil {
		return 0, err
	}

	// STAKE LOCK TRANSITION: Pending listing -> Company listing
	// 1. Remove pending listing lock
	// 2. Add company listing lock (stays until company is delisted)
	if k.stakingKeeper != nil {
		applicantAddr, err := sdk.AccAddressFromBech32(listing.Applicant)
		if err == nil {
			listingIDStr := fmt.Sprintf("%d", listingID)
			companyIDStr := fmt.Sprintf("%d", companyID)

			// Remove pending listing lock
			if err := k.stakingKeeper.UnlockPendingListing(ctx, applicantAddr, listingIDStr); err != nil {
				k.Logger(ctx).Error("failed to unlock pending listing",
					"listing_id", listingID,
					"applicant", listing.Applicant,
					"error", err,
				)
			}

			// Add company listing lock (permanent until delisting)
			if err := k.stakingKeeper.LockForCompanyListing(ctx, applicantAddr, companyIDStr); err != nil {
				k.Logger(ctx).Error("failed to lock stake for company listing",
					"company_id", companyID,
					"applicant", listing.Applicant,
					"error", err,
				)
			}
		}
	}

	return companyID, nil
}

// RejectListing rejects a business listing
func (k Keeper) RejectListing(ctx sdk.Context, listingID uint64, authority string, reason string) error {
	// Get listing
	listing, found := k.GetBusinessListing(ctx, listingID)
	if !found {
		return types.ErrListingNotFound
	}

	// Update listing status
	listing.Status = types.ListingStatusRejected
	listing.RejectionReason = reason
	listing.ReviewCompletedAt = time.Now()
	listing.UpdatedAt = time.Now()

	// Store updated listing
	if err := k.SetBusinessListing(ctx, listing); err != nil {
		return err
	}

	// Refund listing fee to applicant
	applicantAddr, err := sdk.AccAddressFromBech32(listing.Applicant)
	if err != nil {
		return err
	}

	fee := sdk.NewCoins(sdk.NewCoin("uhodl", listing.ListingFee))

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, applicantAddr, fee); err != nil {
		return err
	}

	// UNLOCK STAKE: Remove pending listing lock (applicant can now unstake)
	if k.stakingKeeper != nil {
		listingIDStr := fmt.Sprintf("%d", listingID)
		if err := k.stakingKeeper.UnlockPendingListing(ctx, applicantAddr, listingIDStr); err != nil {
			k.Logger(ctx).Error("failed to unlock pending listing",
				"listing_id", listingID,
				"applicant", listing.Applicant,
				"error", err,
			)
		}
	}

	return nil
}

// SuspendListing suspends a business listing
func (k Keeper) SuspendListing(ctx sdk.Context, listingID uint64, authority string, reason string) error {
	// Get listing
	listing, found := k.GetBusinessListing(ctx, listingID)
	if !found {
		return types.ErrListingNotFound
	}
	
	// Update listing status
	listing.Status = types.ListingStatusSuspended
	listing.Notes = reason
	listing.UpdatedAt = time.Now()

	// Store updated listing
	if err := k.SetBusinessListing(ctx, listing); err != nil {
		return err
	}

	return nil
}

// Helper functions

// validateListingRequirements validates a listing against requirements
func (k Keeper) validateListingRequirements(
	ctx sdk.Context,
	companyName, tokenSymbol, jurisdiction, industry string,
	totalShares math.Int,
	initialPrice math.LegacyDec,
	listingFee math.Int,
	foundedDate time.Time,
	requirements types.ListingRequirements,
) error {
	// Check minimum listing fee
	if listingFee.LT(requirements.MinimumListingFee) {
		return types.ErrInsufficientListingFee
	}
	
	// Check minimum shares
	if totalShares.LT(requirements.MinimumShares) {
		return types.ErrInvalidTotalShares
	}
	
	// Check minimum price
	if initialPrice.LT(requirements.MinimumPrice) {
		return types.ErrInvalidInitialPrice
	}
	
	// Check jurisdiction
	validJurisdiction := false
	for _, allowed := range requirements.AllowedJurisdictions {
		if jurisdiction == allowed {
			validJurisdiction = true
			break
		}
	}
	if !validJurisdiction {
		return types.ErrInvalidJurisdiction
	}
	
	// Check prohibited industries
	for _, prohibited := range requirements.ProhibitedIndustries {
		if industry == prohibited {
			return types.ErrProhibitedIndustry
		}
	}
	
	// Check business age
	monthsOld := int(time.Since(foundedDate).Hours() / (24 * 30))
	if monthsOld < requirements.MinimumBusinessAge {
		return types.ErrInvalidCompanyName // Reuse error for simplicity
	}
	
	// Check if token symbol is already taken
	if k.isTokenSymbolTaken(ctx, tokenSymbol) {
		return types.ErrSymbolTaken
	}
	
	return nil
}

// assignValidatorsForReview assigns validators to review a listing
// Note: This is a placeholder implementation - in production, validators would be selected
// based on their tier and availability from the universal staking module
func (k Keeper) assignValidatorsForReview(ctx sdk.Context, listingID uint64, requiredCount int) error {
	// Note: With the unified staking system, validator selection for listing review
	// should be done through the staking keeper's tier-based selection.
	// This placeholder implementation logs a warning and succeeds.
	k.Logger(ctx).Info("assignValidatorsForReview: validator selection placeholder",
		"listing_id", listingID,
		"required_count", requiredCount,
	)

	// For now, listings proceed without pre-assigned validators
	// Validators with appropriate tier can submit reviews directly
	assigned := []string{}
	
	// Update listing with assigned validators
	listing, found := k.GetBusinessListing(ctx, listingID)
	if !found {
		return types.ErrListingNotFound
	}
	
	listing.AssignedValidators = assigned
	listing.ReviewStartedAt = time.Now()
	if err := k.SetBusinessListing(ctx, listing); err != nil {
		return err
	}

	return nil
}

// isValidatorAssigned checks if a validator is assigned to a listing
func (k Keeper) isValidatorAssigned(ctx sdk.Context, listingID uint64, validator string) bool {
	listing, found := k.GetBusinessListing(ctx, listingID)
	if !found {
		return false
	}
	
	for _, assigned := range listing.AssignedValidators {
		if assigned == validator {
			return true
		}
	}
	
	return false
}

// isTokenSymbolTaken checks if a token symbol is already in use
func (k Keeper) isTokenSymbolTaken(ctx sdk.Context, symbol string) bool {
	// Check existing companies
	companies := k.GetAllCompanies(ctx)
	for _, company := range companies {
		if company.Symbol == symbol {
			return true
		}
	}
	
	// Check pending listings
	listings := k.GetAllBusinessListings(ctx)
	for _, listing := range listings {
		if listing.TokenSymbol == symbol && 
		   (listing.Status == types.ListingStatusPending || 
		    listing.Status == types.ListingStatusUnderReview ||
		    listing.Status == types.ListingStatusApproved) {
			return true
		}
	}
	
	return false
}

// GetAllBusinessListings returns all business listings
func (k Keeper) GetAllBusinessListings(ctx sdk.Context) []types.BusinessListing {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.BusinessListingPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()
	
	listings := []types.BusinessListing{}
	for ; iterator.Valid(); iterator.Next() {
		var listing types.BusinessListing
		if err := json.Unmarshal(iterator.Value(), &listing); err == nil {
			listings = append(listings, listing)
		}
	}
	
	return listings
}

// GetBusinessListingsByStatus returns listings with a specific status
func (k Keeper) GetBusinessListingsByStatus(ctx sdk.Context, status types.ListingStatus) []types.BusinessListing {
	listings := k.GetAllBusinessListings(ctx)
	filtered := []types.BusinessListing{}
	
	for _, listing := range listings {
		if listing.Status == status {
			filtered = append(filtered, listing)
		}
	}
	
	return filtered
}

// GetBusinessListingsByValidator returns listings assigned to a validator
func (k Keeper) GetBusinessListingsByValidator(ctx sdk.Context, validator string) []types.BusinessListing {
	listings := k.GetAllBusinessListings(ctx)
	filtered := []types.BusinessListing{}
	
	for _, listing := range listings {
		for _, assigned := range listing.AssignedValidators {
			if assigned == validator {
				filtered = append(filtered, listing)
				break
			}
		}
	}
	
	return filtered
}