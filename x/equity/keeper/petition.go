package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// =============================================================================
// SHAREHOLDER PETITION SYSTEM
// Allows shareholders without 10K HODL stake to collectively flag fraud concerns
// =============================================================================

// GetNextPetitionID returns the next petition ID and increments the counter
func (k Keeper) GetNextPetitionID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PetitionCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.PetitionCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetPetition returns a petition by ID
func (k Keeper) GetPetition(ctx sdk.Context, petitionID uint64) (types.ShareholderPetition, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPetitionKey(petitionID))
	if bz == nil {
		return types.ShareholderPetition{}, false
	}

	var petition types.ShareholderPetition
	if err := json.Unmarshal(bz, &petition); err != nil {
		return types.ShareholderPetition{}, false
	}
	return petition, true
}

// SetPetition stores a petition
func (k Keeper) SetPetition(ctx sdk.Context, petition types.ShareholderPetition) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(petition)
	if err != nil {
		return fmt.Errorf("failed to marshal petition: %w", err)
	}
	store.Set(types.GetPetitionKey(petition.ID), bz)
	return nil
}

// IndexPetition creates indexes for efficient querying
func (k Keeper) IndexPetition(ctx sdk.Context, petition types.ShareholderPetition) {
	store := ctx.KVStore(k.storeKey)

	// Index by company
	companyKey := types.GetPetitionByCompanyKey(petition.CompanyID, petition.ID)
	store.Set(companyKey, sdk.Uint64ToBigEndian(petition.ID))

	// Index by creator
	creatorKey := types.GetPetitionByCreatorKey(petition.Creator, petition.ID)
	store.Set(creatorKey, sdk.Uint64ToBigEndian(petition.ID))
}

// CreatePetition creates a new shareholder petition
// Any shareholder can create a petition regardless of stake tier
func (k Keeper) CreatePetition(
	ctx sdk.Context,
	creator string,
	companyID uint64,
	classID string,
	petitionType types.PetitionType,
	title, description string,
) (uint64, error) {
	// Validate creator address
	if _, err := sdk.AccAddressFromBech32(creator); err != nil {
		return 0, types.ErrInvalidAddress
	}

	// Check if company exists
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return 0, types.ErrCompanyNotFound
	}

	// Check if share class exists
	shareClass, found := k.getShareClass(ctx, companyID, classID)
	if !found {
		return 0, types.ErrShareClassNotFound
	}

	// REQUIREMENT 1: Creator must hold at least 1 share in the company
	shareholding, found := k.getShareholding(ctx, companyID, classID, creator)
	if !found || shareholding.Shares.LTE(math.ZeroInt()) {
		return 0, types.ErrNotShareholder
	}

	// Generate petition ID
	petitionID := k.GetNextPetitionID(ctx)

	// Create petition with default thresholds
	petition := types.NewShareholderPetition(
		petitionID,
		companyID,
		classID,
		petitionType,
		title,
		description,
		creator,
		types.DefaultShareholderThreshold, // 10% of shareholders
		types.DefaultAbsoluteThreshold,    // OR 100 signatures
		ctx.BlockTime(),
	)

	// Validate petition
	if err := petition.Validate(); err != nil {
		return 0, err
	}

	// Store petition
	if err := k.SetPetition(ctx, petition); err != nil {
		return 0, err
	}

	// Create indexes
	k.IndexPetition(ctx, petition)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"petition_created",
			sdk.NewAttribute("petition_id", fmt.Sprintf("%d", petitionID)),
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute("class_id", classID),
			sdk.NewAttribute("creator", creator),
			sdk.NewAttribute("petition_type", petitionType.String()),
			sdk.NewAttribute("title", title),
		),
	)

	k.Logger(ctx).Info("shareholder petition created",
		"petition_id", petitionID,
		"company_id", companyID,
		"company_name", company.Name,
		"class_id", classID,
		"creator", creator,
		"total_shareholders", len(k.GetCompanyShareholdings(ctx, companyID, classID)),
		"petition_type", petitionType.String())

	_ = shareClass // Suppress unused variable warning

	return petitionID, nil
}

// SignPetition allows a shareholder to sign/support a petition
func (k Keeper) SignPetition(
	ctx sdk.Context,
	petitionID uint64,
	signer string,
	comment string,
) error {
	// Validate signer address
	signerAddr, err := sdk.AccAddressFromBech32(signer)
	if err != nil {
		return types.ErrInvalidAddress
	}

	// Get petition
	petition, found := k.GetPetition(ctx, petitionID)
	if !found {
		return types.ErrPetitionNotFound
	}

	// Check petition is still open
	if petition.Status != types.PetitionStatusOpen {
		return types.ErrPetitionNotOpen
	}

	// Check not expired
	if petition.IsExpired(ctx.BlockTime()) {
		return types.ErrPetitionExpired
	}

	// REQUIREMENT 2: Signer must hold shares in the company
	shareholding, found := k.getShareholding(ctx, petition.CompanyID, petition.ClassID, signer)
	if !found || shareholding.Shares.LTE(math.ZeroInt()) {
		return types.ErrNotShareholder
	}

	// REQUIREMENT 6: Prevent duplicate signatures from same address
	if petition.HasSigned(signer) {
		return types.ErrAlreadySigned
	}

	// Create signature
	signature := types.PetitionSignature{
		Signer:     signer,
		SharesHeld: shareholding.Shares, // Record how many shares they hold
		SignedAt:   ctx.BlockTime(),
		Comment:    comment,
	}

	// Add signature to petition
	if err := petition.AddSignature(signature); err != nil {
		return err
	}

	// Save updated petition
	if err := k.SetPetition(ctx, petition); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"petition_signed",
			sdk.NewAttribute("petition_id", fmt.Sprintf("%d", petitionID)),
			sdk.NewAttribute("signer", signer),
			sdk.NewAttribute("shares_held", shareholding.Shares.String()),
			sdk.NewAttribute("signature_count", fmt.Sprintf("%d", petition.SignatureCount)),
		),
	)

	k.Logger(ctx).Info("petition signed",
		"petition_id", petitionID,
		"signer", signer,
		"shares_held", shareholding.Shares.String(),
		"total_signatures", petition.SignatureCount)

	_ = signerAddr // Suppress unused variable warning

	return nil
}

// ProcessPetitionThresholds checks all open petitions and converts those that met threshold
// Called from EndBlock
func (k Keeper) ProcessPetitionThresholds(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.PetitionPrefix).Iterator(nil, nil)
	defer iterator.Close()

	processedCount := 0
	convertedCount := 0
	expiredCount := 0

	for ; iterator.Valid(); iterator.Next() {
		var petition types.ShareholderPetition
		if err := json.Unmarshal(iterator.Value(), &petition); err != nil {
			k.Logger(ctx).Error("failed to unmarshal petition", "error", err)
			continue
		}

		// Only process open petitions
		if petition.Status != types.PetitionStatusOpen {
			continue
		}

		processedCount++

		// Check if expired
		if petition.IsExpired(ctx.BlockTime()) {
			k.Logger(ctx).Info("petition expired",
				"petition_id", petition.ID,
				"company_id", petition.CompanyID,
				"signature_count", petition.SignatureCount,
				"required", fmt.Sprintf("%d%% or %d", petition.ShareholderThreshold, petition.AbsoluteThreshold))

			petition.Status = types.PetitionStatusExpired
			k.SetPetition(ctx, petition)

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"petition_expired",
					sdk.NewAttribute("petition_id", fmt.Sprintf("%d", petition.ID)),
					sdk.NewAttribute("company_id", fmt.Sprintf("%d", petition.CompanyID)),
					sdk.NewAttribute("signature_count", fmt.Sprintf("%d", petition.SignatureCount)),
				),
			)

			expiredCount++
			continue
		}

		// REQUIREMENT 3: Check threshold
		// Get total shareholder count for company/class
		totalShareholders := len(k.GetCompanyShareholdings(ctx, petition.CompanyID, petition.ClassID))

		thresholdMet, reason := petition.CheckThreshold(totalShareholders)
		if thresholdMet && !petition.ThresholdMet {
			k.Logger(ctx).Info("petition threshold met",
				"petition_id", petition.ID,
				"company_id", petition.CompanyID,
				"reason", reason,
				"signature_count", petition.SignatureCount,
				"total_shareholders", totalShareholders)

			petition.ThresholdMet = true
			petition.Status = types.PetitionStatusThresholdMet

			// REQUIREMENT 4: Auto-create report in escrow module with ReportTypeFraud
			if err := k.convertPetitionToReport(ctx, &petition); err != nil {
				k.Logger(ctx).Error("failed to convert petition to report",
					"petition_id", petition.ID,
					"error", err)
				// Continue - mark as threshold met but not converted
			} else {
				petition.Status = types.PetitionStatusConverted
				petition.ConvertedAt = ctx.BlockTime()
				convertedCount++
			}

			k.SetPetition(ctx, petition)

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"petition_threshold_met",
					sdk.NewAttribute("petition_id", fmt.Sprintf("%d", petition.ID)),
					sdk.NewAttribute("company_id", fmt.Sprintf("%d", petition.CompanyID)),
					sdk.NewAttribute("reason", reason),
					sdk.NewAttribute("signature_count", fmt.Sprintf("%d", petition.SignatureCount)),
					sdk.NewAttribute("converted_to_report", fmt.Sprintf("%d", petition.ConvertedToReportID)),
				),
			)
		}
	}

	if processedCount > 0 {
		k.Logger(ctx).Info("processed petitions in EndBlock",
			"processed", processedCount,
			"converted", convertedCount,
			"expired", expiredCount)
	}
}

// convertPetitionToReport converts a petition that met threshold into a formal fraud report
// This creates a report in the escrow module with elevated priority
func (k Keeper) convertPetitionToReport(ctx sdk.Context, petition *types.ShareholderPetition) error {
	// TODO: Integration with escrow module to create report
	// For now, we'll emit an event that can be picked up by the escrow module
	// or create the report directly if we have access to the escrow keeper

	// REQUIREMENT 5: The converted report should have elevated priority (simulates Keeper-tier reporter)
	priority := petition.ConvertToPriority()

	company, found := k.getCompany(ctx, petition.CompanyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Construct report description
	reportDescription := fmt.Sprintf("SHAREHOLDER PETITION (ID: %d)\n\n", petition.ID)
	reportDescription += fmt.Sprintf("Petition: %s\n\n", petition.Title)
	reportDescription += fmt.Sprintf("Description: %s\n\n", petition.Description)
	reportDescription += fmt.Sprintf("Supported by %d shareholders holding %s shares\n\n",
		petition.SignatureCount,
		petition.GetTotalSharesSigned().String())
	reportDescription += "Top signatures:\n"

	// Include top 10 signatures with comments
	sigCount := 0
	for _, sig := range petition.Signatures {
		if sigCount >= 10 {
			break
		}
		if sig.Comment != "" {
			reportDescription += fmt.Sprintf("- %s (%s shares): %s\n",
				sig.Signer, sig.SharesHeld.String(), sig.Comment)
			sigCount++
		}
	}

	// Emit event for escrow module to pick up
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"petition_convert_to_report",
			sdk.NewAttribute("petition_id", fmt.Sprintf("%d", petition.ID)),
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", petition.CompanyID)),
			sdk.NewAttribute("company_name", company.Name),
			sdk.NewAttribute("petition_type", petition.PetitionType.String()),
			sdk.NewAttribute("title", petition.Title),
			sdk.NewAttribute("description", reportDescription),
			sdk.NewAttribute("priority", fmt.Sprintf("%d", priority)),
			sdk.NewAttribute("severity", "4"), // High severity by default
			sdk.NewAttribute("signature_count", fmt.Sprintf("%d", petition.SignatureCount)),
			sdk.NewAttribute("total_shares_signed", petition.GetTotalSharesSigned().String()),
			sdk.NewAttribute("creator", petition.Creator),
		),
	)

	k.Logger(ctx).Info("petition converted to fraud report",
		"petition_id", petition.ID,
		"company_id", petition.CompanyID,
		"company_name", company.Name,
		"signature_count", petition.SignatureCount,
		"priority", priority)

	// Store the report ID when created (placeholder for now)
	// In a full implementation, this would call the escrow keeper directly
	petition.ConvertedToReportID = 0 // Will be set by escrow module

	return nil
}

// GetPetitionsByCompany returns all petitions for a specific company
func (k Keeper) GetPetitionsByCompany(ctx sdk.Context, companyID uint64) []types.ShareholderPetition {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetPetitionsByCompanyPrefix(companyID)
	iterator := prefix.NewStore(store, prefixKey).Iterator(nil, nil)
	defer iterator.Close()

	var petitions []types.ShareholderPetition
	for ; iterator.Valid(); iterator.Next() {
		petitionID := sdk.BigEndianToUint64(iterator.Value())
		if petition, found := k.GetPetition(ctx, petitionID); found {
			petitions = append(petitions, petition)
		}
	}

	return petitions
}

// GetPetitionsByCreator returns all petitions created by a specific address
func (k Keeper) GetPetitionsByCreator(ctx sdk.Context, creator string) []types.ShareholderPetition {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetPetitionsByCreatorPrefix(creator)
	iterator := prefix.NewStore(store, prefixKey).Iterator(nil, nil)
	defer iterator.Close()

	var petitions []types.ShareholderPetition
	for ; iterator.Valid(); iterator.Next() {
		petitionID := sdk.BigEndianToUint64(iterator.Value())
		if petition, found := k.GetPetition(ctx, petitionID); found {
			petitions = append(petitions, petition)
		}
	}

	return petitions
}

// GetOpenPetitions returns all open petitions
func (k Keeper) GetOpenPetitions(ctx sdk.Context) []types.ShareholderPetition {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.PetitionPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var petitions []types.ShareholderPetition
	for ; iterator.Valid(); iterator.Next() {
		var petition types.ShareholderPetition
		if err := json.Unmarshal(iterator.Value(), &petition); err != nil {
			continue
		}
		if petition.Status == types.PetitionStatusOpen {
			petitions = append(petitions, petition)
		}
	}

	return petitions
}

// GetAllPetitions returns all petitions (for genesis export)
func (k Keeper) GetAllPetitions(ctx sdk.Context) []types.ShareholderPetition {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.PetitionPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var petitions []types.ShareholderPetition
	for ; iterator.Valid(); iterator.Next() {
		var petition types.ShareholderPetition
		if err := json.Unmarshal(iterator.Value(), &petition); err != nil {
			continue
		}
		petitions = append(petitions, petition)
	}

	return petitions
}

// WithdrawPetition allows the creator to withdraw their petition
func (k Keeper) WithdrawPetition(ctx sdk.Context, petitionID uint64, withdrawer string) error {
	petition, found := k.GetPetition(ctx, petitionID)
	if !found {
		return types.ErrPetitionNotFound
	}

	// Only creator can withdraw
	if petition.Creator != withdrawer {
		return types.ErrNotPetitionCreator
	}

	// Can only withdraw open petitions
	if petition.Status != types.PetitionStatusOpen {
		return types.ErrPetitionNotOpen
	}

	// Mark as withdrawn
	petition.Status = types.PetitionStatusWithdrawn
	if err := k.SetPetition(ctx, petition); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"petition_withdrawn",
			sdk.NewAttribute("petition_id", fmt.Sprintf("%d", petitionID)),
			sdk.NewAttribute("withdrawer", withdrawer),
			sdk.NewAttribute("signature_count", fmt.Sprintf("%d", petition.SignatureCount)),
		),
	)

	k.Logger(ctx).Info("petition withdrawn",
		"petition_id", petitionID,
		"withdrawer", withdrawer)

	return nil
}
