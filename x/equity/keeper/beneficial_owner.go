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

// BeneficialOwnership is an alias for types.BeneficialOwnership for backwards compatibility
type BeneficialOwnership = types.BeneficialOwnership

// DividendRecipient represents a recipient for dividend distribution
// Combines direct holders and beneficial owners from module accounts
type DividendRecipient struct {
	Address     string   `json:"address"`      // Recipient address
	Shares      math.Int `json:"shares"`       // Number of shares owned
	IsModule    bool     `json:"is_module"`    // true if this is a module account (skip, use beneficial owners)
	CompanyID   uint64   `json:"company_id"`   // Company ID
	ClassID     string   `json:"class_id"`     // Share class ID
}

// RegisterBeneficialOwner records that shares in a module account belong to a specific owner
// This should be called by escrow/lending/DEX modules when they lock shares
// Example: When Alice puts shares in escrow, the escrow module calls this to record Alice as the beneficial owner
func (k Keeper) RegisterBeneficialOwner(
	ctx sdk.Context,
	moduleAccount string,
	companyID uint64,
	classID string,
	beneficialOwner string,
	shares math.Int,
	refID uint64,
	refType string,
) error {
	// Validate inputs
	if moduleAccount == "" {
		return types.ErrInvalidInput.Wrap("module account cannot be empty")
	}
	if beneficialOwner == "" {
		return types.ErrInvalidInput.Wrap("beneficial owner cannot be empty")
	}
	if shares.IsNil() || shares.LTE(math.ZeroInt()) {
		return types.ErrInvalidInput.Wrap("shares must be positive")
	}
	if _, err := sdk.AccAddressFromBech32(beneficialOwner); err != nil {
		return types.ErrInvalidInput.Wrapf("invalid beneficial owner address: %v", err)
	}

	// Create beneficial ownership record
	ownership := BeneficialOwnership{
		ModuleAccount:   moduleAccount,
		CompanyID:       companyID,
		ClassID:         classID,
		BeneficialOwner: beneficialOwner,
		Shares:          shares,
		LockedAt:        ctx.BlockTime(),
		ReferenceID:     refID,
		ReferenceType:   refType,
	}

	// Validate the record
	if err := ownership.Validate(); err != nil {
		return types.ErrInvalidInput.Wrap(err.Error())
	}

	// Store the ownership record
	store := ctx.KVStore(k.storeKey)
	key := types.GetBeneficialOwnerKey(moduleAccount, companyID, classID, beneficialOwner, refID)
	bz, err := json.Marshal(ownership)
	if err != nil {
		return types.ErrInternalError.Wrap(err.Error())
	}
	store.Set(key, bz)

	// Create index for fast lookup by module
	indexKey := types.GetBeneficialOwnerByModuleKey(moduleAccount, companyID, classID)
	store.Set(indexKey, []byte{1}) // Just a marker, actual data is in the main key

	k.Logger(ctx).Info("registered beneficial owner",
		"module", moduleAccount,
		"company_id", companyID,
		"class_id", classID,
		"beneficial_owner", beneficialOwner,
		"shares", shares.String(),
		"ref_type", refType,
		"ref_id", refID,
	)

	return nil
}

// UnregisterBeneficialOwner removes a beneficial ownership record when shares are released
// Called when escrow completes, loan is repaid, or order is filled/cancelled
func (k Keeper) UnregisterBeneficialOwner(
	ctx sdk.Context,
	moduleAccount string,
	companyID uint64,
	classID string,
	beneficialOwner string,
	refID uint64,
) error {
	// Validate inputs
	if moduleAccount == "" {
		return types.ErrInvalidInput.Wrap("module account cannot be empty")
	}
	if beneficialOwner == "" {
		return types.ErrInvalidInput.Wrap("beneficial owner cannot be empty")
	}

	// Get the store
	store := ctx.KVStore(k.storeKey)
	key := types.GetBeneficialOwnerKey(moduleAccount, companyID, classID, beneficialOwner, refID)

	// Check if the record exists
	if !store.Has(key) {
		return types.ErrBeneficialOwnerNotFound.Wrapf(
			"no beneficial ownership found for module=%s company=%d class=%s owner=%s ref=%d",
			moduleAccount, companyID, classID, beneficialOwner, refID,
		)
	}

	// Delete the record
	store.Delete(key)

	k.Logger(ctx).Info("unregistered beneficial owner",
		"module", moduleAccount,
		"company_id", companyID,
		"class_id", classID,
		"beneficial_owner", beneficialOwner,
		"ref_id", refID,
	)

	return nil
}

// GetBeneficialOwnershipsByOwner returns all beneficial ownership records for a specific beneficial owner
// This is used by governance to calculate voting power including shares in LP pools, escrow, etc.
func (k Keeper) GetBeneficialOwnershipsByOwner(
	ctx sdk.Context,
	companyID uint64,
	classID string,
	beneficialOwner string,
) ([]BeneficialOwnership, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.BeneficialOwnerPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	ownerships := []BeneficialOwnership{}
	for ; iterator.Valid(); iterator.Next() {
		var ownership BeneficialOwnership
		if err := json.Unmarshal(iterator.Value(), &ownership); err != nil {
			continue
		}
		// Match by company, class, and beneficial owner
		if ownership.CompanyID == companyID &&
			ownership.ClassID == classID &&
			ownership.BeneficialOwner == beneficialOwner {
			ownerships = append(ownerships, ownership)
		}
	}

	return ownerships, nil
}

// GetBeneficialOwnersForModule returns all beneficial ownership records for a specific module and share class
func (k Keeper) GetBeneficialOwnersForModule(
	ctx sdk.Context,
	moduleAccount string,
	companyID uint64,
	classID string,
) ([]BeneficialOwnership, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.BeneficialOwnerPrefix)

	// Create prefix for module + company + class
	keyPrefix := []byte(moduleAccount + "|")
	keyPrefix = append(keyPrefix, sdk.Uint64ToBigEndian(companyID)...)
	keyPrefix = append(keyPrefix, []byte(classID)...)

	iterator := store.Iterator(keyPrefix, storetypes.PrefixEndBytes(keyPrefix))
	defer iterator.Close()

	ownerships := []BeneficialOwnership{}
	for ; iterator.Valid(); iterator.Next() {
		var ownership BeneficialOwnership
		if err := json.Unmarshal(iterator.Value(), &ownership); err != nil {
			k.Logger(ctx).Error("failed to unmarshal beneficial ownership", "error", err)
			continue
		}
		ownerships = append(ownerships, ownership)
	}

	return ownerships, nil
}

// GetBeneficialOwnersForDividend returns all dividend recipients including beneficial owners
// This is the KEY function called by the dividend system to determine who should receive dividends
// It combines:
// 1. Direct shareholders (normal wallet holdings)
// 2. Beneficial owners of shares held in module accounts (escrow, lending, DEX)
// IMPORTANT: Treasury shares are EXCLUDED from dividends (standard corporate practice)
func (k Keeper) GetBeneficialOwnersForDividend(
	ctx sdk.Context,
	companyID uint64,
	classID string,
) []DividendRecipient {
	recipients := []DividendRecipient{}
	seenAddresses := make(map[string]math.Int) // Track aggregated shares per address

	// Get addresses to exclude from dividends (treasury, company founder holding treasury shares)
	excludedAddresses := k.GetDividendExcludedAddresses(ctx, companyID)

	// 1. Get all direct shareholdings for this company/class
	shareholdings := k.GetCompanyShareholdings(ctx, companyID, classID)

	for _, holding := range shareholdings {
		// Check if this address is excluded from dividends (treasury shares)
		if excludedAddresses[holding.Owner] {
			k.Logger(ctx).Debug("excluding treasury/excluded address from dividend",
				"address", holding.Owner,
				"shares", holding.VestedShares.String(),
			)
			continue
		}

		// Check if this is a module account
		isModule := k.isModuleAccount(holding.Owner)

		if !isModule {
			// Regular shareholder - add their vested shares
			if holding.VestedShares.GT(math.ZeroInt()) {
				if existing, found := seenAddresses[holding.Owner]; found {
					seenAddresses[holding.Owner] = existing.Add(holding.VestedShares)
				} else {
					seenAddresses[holding.Owner] = holding.VestedShares
				}
			}
		} else {
			// Module account - lookup beneficial owners
			k.Logger(ctx).Debug("found module account holding, looking up beneficial owners",
				"module", holding.Owner,
				"shares", holding.VestedShares.String(),
			)

			beneficialOwners, err := k.GetBeneficialOwnersForModule(ctx, holding.Owner, companyID, classID)
			if err != nil {
				k.Logger(ctx).Error("failed to get beneficial owners for module",
					"module", holding.Owner,
					"error", err,
				)
				continue
			}

			// Add beneficial owners
			for _, bo := range beneficialOwners {
				if bo.Shares.GT(math.ZeroInt()) {
					if existing, found := seenAddresses[bo.BeneficialOwner]; found {
						seenAddresses[bo.BeneficialOwner] = existing.Add(bo.Shares)
					} else {
						seenAddresses[bo.BeneficialOwner] = bo.Shares
					}

					k.Logger(ctx).Debug("added beneficial owner to dividend recipients",
						"owner", bo.BeneficialOwner,
						"shares", bo.Shares.String(),
						"module", bo.ModuleAccount,
						"ref_type", bo.ReferenceType,
					)
				}
			}
		}
	}

	// 2. Add treasury investment holdings from OTHER companies
	// When Company B's treasury holds Company A shares, Company B should receive dividends
	treasuryRecipients := k.GetTreasuryInvestmentHoldersForDividend(ctx, companyID, classID)
	for _, tr := range treasuryRecipients {
		// Use a special treasury identifier as the recipient
		// Format: "treasury_dividend:ownerCompanyID"
		treasuryKey := fmt.Sprintf("treasury_dividend:%d", tr.OwnerCompanyID)
		if existing, found := seenAddresses[treasuryKey]; found {
			seenAddresses[treasuryKey] = existing.Add(tr.Shares)
		} else {
			seenAddresses[treasuryKey] = tr.Shares
		}

		k.Logger(ctx).Debug("added treasury investment holder to dividend recipients",
			"owner_company_id", tr.OwnerCompanyID,
			"shares", tr.Shares.String(),
		)
	}

	// 3. Convert map to slice of recipients (ONCE, after all additions)
	for address, shares := range seenAddresses {
		recipients = append(recipients, DividendRecipient{
			Address:   address,
			Shares:    shares,
			IsModule:  false,
			CompanyID: companyID,
			ClassID:   classID,
		})
	}

	k.Logger(ctx).Info("calculated dividend recipients",
		"company_id", companyID,
		"class_id", classID,
		"total_recipients", len(recipients),
	)

	return recipients
}

// TreasuryInvestmentHolder represents a treasury that holds shares for dividend purposes
type TreasuryInvestmentHolder struct {
	OwnerCompanyID uint64   // Company that owns the treasury
	Shares         math.Int // Shares held in the dividend-paying company
}

// GetTreasuryInvestmentHoldersForDividend returns all treasuries holding shares in a company
// This is used to route dividends to other companies' treasuries when they hold investment shares
// IMPORTANT: This excludes the company's OWN treasury (no self-dividends)
func (k Keeper) GetTreasuryInvestmentHoldersForDividend(
	ctx sdk.Context,
	companyID uint64, // The company paying dividends
	classID string,
) []TreasuryInvestmentHolder {
	holders := []TreasuryInvestmentHolder{}

	// Get all company treasuries
	treasuries := k.GetAllCompanyTreasuries(ctx)

	for _, treasury := range treasuries {
		// Skip the company's own treasury (no self-dividends)
		if treasury.CompanyID == companyID {
			continue
		}

		// Check if this treasury has investment holdings in the dividend-paying company
		shares := treasury.GetInvestmentShares(companyID, classID)
		if shares.GT(math.ZeroInt()) {
			holders = append(holders, TreasuryInvestmentHolder{
				OwnerCompanyID: treasury.CompanyID,
				Shares:         shares,
			})

			k.Logger(ctx).Debug("found treasury investment holder",
				"owner_company_id", treasury.CompanyID,
				"target_company_id", companyID,
				"class_id", classID,
				"shares", shares.String(),
			)
		}
	}

	return holders
}

// GetDividendExcludedAddresses returns addresses that should NOT receive dividends
// This includes:
// 1. Protocol treasury (protocol-level treasury shares)
// 2. Company treasury module account
// Treasury shares don't receive dividends because it would be circular (company paying itself)
func (k Keeper) GetDividendExcludedAddresses(ctx sdk.Context, companyID uint64) map[string]bool {
	excluded := make(map[string]bool)

	// 1. Exclude protocol treasury (well-known address that holds HODL for trading)
	// This is the ShareHODL protocol treasury - shares held here are for sale, not dividends
	protocolTreasury := "sharehodl1treasury0000000000000000000000"
	excluded[protocolTreasury] = true

	// 2. Exclude equity module account (holds company treasury funds)
	equityModuleAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if equityModuleAddr != nil {
		excluded[equityModuleAddr.String()] = true
	}

	// 3. For company-specific exclusions, check if company has a designated treasury holder
	// Note: Company founder shares ARE eligible for dividends (they're an investor)
	// Only treasury shares (held for sale, not as investment) are excluded

	k.Logger(ctx).Debug("dividend excluded addresses",
		"company_id", companyID,
		"excluded_count", len(excluded),
	)

	return excluded
}

// isModuleAccount checks if an address is a module account
// Module accounts are special accounts used by blockchain modules (escrow, lending, DEX, etc.)
func (k Keeper) isModuleAccount(address string) bool {
	// Try to parse the address
	addr, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return false
	}

	// Check against known module accounts
	// These are the common module accounts in Cosmos SDK apps
	moduleNames := []string{
		"escrow",
		"lending",
		"dex",
		types.ModuleName, // equity module itself
		"hodl",
		"validator",
		"governance",
		"explorer",
		"distribution",
		"fee_collector",
		"bonded_tokens_pool",
		"not_bonded_tokens_pool",
	}

	for _, moduleName := range moduleNames {
		moduleAddr := k.accountKeeper.GetModuleAddress(moduleName)
		if moduleAddr != nil && moduleAddr.Equals(addr) {
			return true
		}
	}

	return false
}

// GetBeneficialOwnerForReference retrieves a specific beneficial ownership record by reference
func (k Keeper) GetBeneficialOwnerForReference(
	ctx sdk.Context,
	moduleAccount string,
	companyID uint64,
	classID string,
	beneficialOwner string,
	refID uint64,
) (BeneficialOwnership, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetBeneficialOwnerKey(moduleAccount, companyID, classID, beneficialOwner, refID)

	bz := store.Get(key)
	if bz == nil {
		return BeneficialOwnership{}, false
	}

	var ownership BeneficialOwnership
	if err := json.Unmarshal(bz, &ownership); err != nil {
		k.Logger(ctx).Error("failed to unmarshal beneficial ownership", "error", err)
		return BeneficialOwnership{}, false
	}

	return ownership, true
}

// GetAllBeneficialOwnerships returns all beneficial ownership records (useful for debugging/admin)
func (k Keeper) GetAllBeneficialOwnerships(ctx sdk.Context) []BeneficialOwnership {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.BeneficialOwnerPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	ownerships := []BeneficialOwnership{}
	for ; iterator.Valid(); iterator.Next() {
		var ownership BeneficialOwnership
		if err := json.Unmarshal(iterator.Value(), &ownership); err != nil {
			k.Logger(ctx).Error("failed to unmarshal beneficial ownership", "error", err)
			continue
		}
		ownerships = append(ownerships, ownership)
	}

	return ownerships
}

// UpdateBeneficialOwnerShares updates the share count for an existing beneficial ownership record
// This is used when partial fills occur in DEX orders or partial loan repayments
func (k Keeper) UpdateBeneficialOwnerShares(
	ctx sdk.Context,
	moduleAccount string,
	companyID uint64,
	classID string,
	beneficialOwner string,
	refID uint64,
	newShares math.Int,
) error {
	// Validate inputs
	if newShares.IsNil() || newShares.LT(math.ZeroInt()) {
		return types.ErrInvalidInput.Wrap("shares cannot be negative")
	}

	// Get existing record
	ownership, found := k.GetBeneficialOwnerForReference(ctx, moduleAccount, companyID, classID, beneficialOwner, refID)
	if !found {
		return types.ErrBeneficialOwnerNotFound.Wrapf(
			"no beneficial ownership found for module=%s company=%d class=%s owner=%s ref=%d",
			moduleAccount, companyID, classID, beneficialOwner, refID,
		)
	}

	// If new shares is zero, delete the record
	if newShares.IsZero() {
		return k.UnregisterBeneficialOwner(ctx, moduleAccount, companyID, classID, beneficialOwner, refID)
	}

	// Update shares
	oldShares := ownership.Shares
	ownership.Shares = newShares

	// Validate updated record
	if err := ownership.Validate(); err != nil {
		return types.ErrInvalidInput.Wrap(err.Error())
	}

	// Store updated record
	store := ctx.KVStore(k.storeKey)
	key := types.GetBeneficialOwnerKey(moduleAccount, companyID, classID, beneficialOwner, refID)
	bz, err := json.Marshal(ownership)
	if err != nil {
		return types.ErrInternalError.Wrap(err.Error())
	}
	store.Set(key, bz)

	k.Logger(ctx).Info("updated beneficial owner shares",
		"module", moduleAccount,
		"company_id", companyID,
		"class_id", classID,
		"beneficial_owner", beneficialOwner,
		"ref_id", refID,
		"old_shares", oldShares.String(),
		"new_shares", newShares.String(),
	)

	return nil
}
