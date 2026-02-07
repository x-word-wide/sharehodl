package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// Expected keepers for dependency injection
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

type AccountKeeper interface {
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	SetAccount(context.Context, sdk.AccountI)
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
	GetModuleAddress(string) sdk.AccAddress
}

// Keeper of the equity store
type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        storetypes.StoreKey
	memKey          storetypes.StoreKey
	bankKeeper      BankKeeper
	accountKeeper   AccountKeeper
	stakingKeeper   types.UniversalStakingKeeper // For tier checks and stake locks
	validatorKeeper types.ValidatorKeeper        // For validator authorization checks
	authority       string                       // Governance module address for privileged operations
}

// NewKeeper creates a new equity Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper BankKeeper,
	accountKeeper AccountKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		stakingKeeper: nil, // Set later via SetStakingKeeper
		authority:     authority,
	}
}

// GetAuthority returns the governance authority address
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetStakingKeeper sets the staking keeper (for late binding during app initialization)
func (k *Keeper) SetStakingKeeper(stakingKeeper types.UniversalStakingKeeper) {
	k.stakingKeeper = stakingKeeper
}

// SetValidatorKeeper sets the validator keeper (for late binding during app initialization)
func (k *Keeper) SetValidatorKeeper(validatorKeeper types.ValidatorKeeper) {
	k.validatorKeeper = validatorKeeper
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Company management methods

// GetNextCompanyID returns the next company ID and increments the counter
func (k Keeper) GetNextCompanyID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CompanyCounterKey)
	
	var counter uint64 = 1 // Start from 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}
	
	// Increment counter for next use
	store.Set(types.CompanyCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetCompany returns a company by ID (interface compatible)
func (k Keeper) GetCompany(ctx context.Context, companyID uint64) (interface{}, bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	company, found := k.getCompany(sdkCtx, companyID)
	return company, found
}

func (k Keeper) getCompany(ctx sdk.Context, companyID uint64) (types.Company, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyKey(companyID)
	bz := store.Get(key)
	if bz == nil {
		return types.Company{}, false
	}
	
	var company types.Company
	if err := json.Unmarshal(bz, &company); err != nil {
		return types.Company{}, false
	}
	return company, true
}

// SetCompany stores a company
func (k Keeper) SetCompany(ctx sdk.Context, company types.Company) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyKey(company.ID)
	bz, err := json.Marshal(company)
	if err != nil {
		return fmt.Errorf("failed to marshal company: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// DeleteCompany removes a company
func (k Keeper) DeleteCompany(ctx sdk.Context, companyID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyKey(companyID)
	store.Delete(key)
}

// DelistCompany delists a company and unlocks the founder's stake
func (k Keeper) DelistCompany(ctx sdk.Context, companyID uint64, authority string, reason string) error {
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Update company status to delisted
	company.Status = types.CompanyStatusDelisted
	company.Description = company.Description + " [DELISTED: " + reason + "]"
	if err := k.SetCompany(ctx, company); err != nil {
		return err
	}

	// UNLOCK STAKE: Remove company listing lock so founder can unstake
	if k.stakingKeeper != nil {
		founderAddr, err := sdk.AccAddressFromBech32(company.Founder)
		if err == nil {
			companyIDStr := fmt.Sprintf("%d", companyID)
			if err := k.stakingKeeper.UnlockCompanyListing(ctx, founderAddr, companyIDStr); err != nil {
				k.Logger(ctx).Error("failed to unlock company listing",
					"company_id", companyID,
					"founder", company.Founder,
					"error", err,
				)
			} else {
				k.Logger(ctx).Info("company delisted, stake unlocked",
					"company_id", companyID,
					"founder", company.Founder,
				)
			}
		}
	}

	return nil
}

// IsSymbolTaken checks if a company symbol is already taken
func (k Keeper) IsSymbolTaken(ctx context.Context, symbol string) bool {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return k.isSymbolTaken(sdkCtx, symbol)
}

func (k Keeper) isSymbolTaken(ctx sdk.Context, symbol string) bool {
	companies := k.GetAllCompanies(ctx)
	for _, company := range companies {
		if company.Symbol == symbol {
			return true
		}
	}
	return false
}

// GetAllCompanies returns all companies
func (k Keeper) GetAllCompanies(ctx sdk.Context) []types.Company {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.CompanyPrefix)
	defer iterator.Close()

	var companies []types.Company
	for ; iterator.Valid(); iterator.Next() {
		var company types.Company
		if err := json.Unmarshal(iterator.Value(), &company); err != nil {
			k.Logger(ctx).Error("failed to unmarshal company", "error", err)
			continue
		}
		companies = append(companies, company)
	}
	return companies
}

// IterateCompanies iterates through all companies and calls the callback function
func (k Keeper) IterateCompanies(ctx sdk.Context, cb func(company types.Company) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.CompanyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var company types.Company
		if err := json.Unmarshal(iterator.Value(), &company); err != nil {
			k.Logger(ctx).Error("failed to unmarshal company", "error", err)
			continue
		}
		if cb(company) {
			break
		}
	}
}

// GetCompanyBySymbol returns company ID for a trading symbol
// Used by DEX to map equity symbols to company IDs for beneficial owner tracking
func (k Keeper) GetCompanyBySymbol(ctx sdk.Context, symbol string) (uint64, bool) {
	var foundID uint64
	var found bool

	k.IterateCompanies(ctx, func(company types.Company) bool {
		if company.Symbol == symbol {
			foundID = company.ID
			found = true
			return true // stop iteration
		}
		return false
	})

	return foundID, found
}

// Share class management methods


// SetShareClass stores a share class
func (k Keeper) SetShareClass(ctx sdk.Context, shareClass types.ShareClass) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareClassKey(shareClass.CompanyID, shareClass.ClassID)
	bz, err := json.Marshal(shareClass)
	if err != nil {
		return fmt.Errorf("failed to marshal share class: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// DeleteShareClass removes a share class
func (k Keeper) DeleteShareClass(ctx sdk.Context, companyID uint64, classID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareClassKey(companyID, classID)
	store.Delete(key)
}

// GetCompanyShareClasses returns all share classes for a company
func (k Keeper) GetCompanyShareClasses(ctx sdk.Context, companyID uint64) []types.ShareClass {
	store := ctx.KVStore(k.storeKey)
	// Construct prefix for company's share classes
	prefix := append(types.ShareClassPrefix, sdk.Uint64ToBigEndian(companyID)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var shareClasses []types.ShareClass
	for ; iterator.Valid(); iterator.Next() {
		var shareClass types.ShareClass
		if err := json.Unmarshal(iterator.Value(), &shareClass); err != nil {
			k.Logger(ctx).Error("failed to unmarshal share class", "error", err)
			continue
		}
		shareClasses = append(shareClasses, shareClass)
	}
	return shareClasses
}

// Shareholding management methods


// SetShareholding stores a shareholding
func (k Keeper) SetShareholding(ctx sdk.Context, shareholding types.Shareholding) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholdingKey(shareholding.CompanyID, shareholding.ClassID, shareholding.Owner)
	bz, err := json.Marshal(shareholding)
	if err != nil {
		return fmt.Errorf("failed to marshal shareholding: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// DeleteShareholding removes a shareholding
func (k Keeper) DeleteShareholding(ctx sdk.Context, companyID uint64, classID, owner string) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholdingKey(companyID, classID, owner)
	store.Delete(key)
}

// GetOwnerShareholdings returns all shareholdings for an owner
func (k Keeper) GetOwnerShareholdings(ctx sdk.Context, owner string) []types.Shareholding {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ShareholdingPrefix)
	defer iterator.Close()

	var shareholdings []types.Shareholding
	for ; iterator.Valid(); iterator.Next() {
		var shareholding types.Shareholding
		if err := json.Unmarshal(iterator.Value(), &shareholding); err != nil {
			k.Logger(ctx).Error("failed to unmarshal shareholding", "error", err)
			continue
		}
		// Filter by owner
		if shareholding.Owner == owner {
			shareholdings = append(shareholdings, shareholding)
		}
	}
	return shareholdings
}

// GetCompanyShareholdings returns all shareholdings for a company and share class
func (k Keeper) GetCompanyShareholdings(ctx sdk.Context, companyID uint64, classID string) []types.Shareholding {
	store := ctx.KVStore(k.storeKey)
	// Construct prefix for company's shareholdings with classID
	prefix := append(types.ShareholdingPrefix, sdk.Uint64ToBigEndian(companyID)...)
	prefix = append(prefix, []byte(classID)...)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var shareholdings []types.Shareholding
	for ; iterator.Valid(); iterator.Next() {
		var shareholding types.Shareholding
		if err := json.Unmarshal(iterator.Value(), &shareholding); err != nil {
			k.Logger(ctx).Error("failed to unmarshal shareholding", "error", err)
			continue
		}
		shareholdings = append(shareholdings, shareholding)
	}
	return shareholdings
}

// Business logic methods

// IssueShares issues new shares to a recipient
func (k Keeper) IssueShares(
	ctx sdk.Context,
	companyID uint64,
	classID string,
	recipient string,
	shares math.Int,
	pricePerShare math.LegacyDec,
	paymentDenom string,
) error {
	// Get share class
	shareClass, found := k.getShareClass(ctx, companyID, classID)
	if !found {
		return types.ErrShareClassNotFound
	}
	
	// Check if new issuance would exceed authorized shares
	newTotalIssued := shareClass.IssuedShares.Add(shares)
	if newTotalIssued.GT(shareClass.AuthorizedShares) {
		return types.ErrExceedsAuthorized
	}
	
	// Calculate total payment required
	totalPayment := pricePerShare.MulInt(shares)
	
	// Get or create shareholding
	shareholding, exists := k.getShareholding(ctx, companyID, classID, recipient)
	if exists {
		// Update existing shareholding
		weightedCostBasis := shareholding.CostBasis.MulInt(shareholding.Shares).Add(totalPayment)
		newTotalShares := shareholding.Shares.Add(shares)
		newCostBasis := weightedCostBasis.QuoInt(newTotalShares)
		
		shareholding.Shares = newTotalShares
		shareholding.VestedShares = shareholding.VestedShares.Add(shares) // Assume fully vested
		shareholding.CostBasis = newCostBasis
		shareholding.TotalCost = shareholding.TotalCost.Add(totalPayment)
	} else {
		// Create new shareholding
		shareholding = types.NewShareholding(companyID, classID, recipient, shares, pricePerShare)
	}
	
	// Update share class
	shareClass.IssuedShares = newTotalIssued
	shareClass.OutstandingShares = shareClass.OutstandingShares.Add(shares)
	
	// Save updates
	if err := k.SetShareholding(ctx, shareholding); err != nil {
		return err
	}
	if err := k.SetShareClass(ctx, shareClass); err != nil {
		return err
	}

	return nil
}

// TransferShares transfers shares between addresses
func (k Keeper) TransferShares(
	ctx sdk.Context,
	companyID uint64,
	classID string,
	from, to string,
	shares math.Int,
) error {
	// SECURITY: Validate share amount is positive
	if shares.IsNil() || !shares.IsPositive() {
		return types.ErrInsufficientShares
	}

	// SECURITY: Validate addresses are different
	if from == to {
		return types.ErrTransferRestricted
	}

	// Get share class to check transferability
	shareClass, found := k.getShareClass(ctx, companyID, classID)
	if !found {
		return types.ErrShareClassNotFound
	}

	if !shareClass.Transferable {
		return types.ErrTransferRestricted
	}

	// SECURITY: Check for trading halt - prevents transfers during fraud investigations
	if k.IsTradingHalted(ctx, companyID) {
		return types.ErrTradingHalted
	}

	// Get from shareholding
	fromHolding, found := k.getShareholding(ctx, companyID, classID, from)
	if !found {
		return types.ErrShareholdingNotFound
	}
	
	// Check if sufficient shares available (considering locked shares)
	availableShares := fromHolding.VestedShares.Sub(fromHolding.LockedShares)
	if availableShares.LT(shares) {
		return types.ErrInsufficientShares
	}
	
	// Update from holding
	fromHolding.Shares = fromHolding.Shares.Sub(shares)
	fromHolding.VestedShares = fromHolding.VestedShares.Sub(shares)
	
	// Get or create to shareholding
	toHolding, exists := k.getShareholding(ctx, companyID, classID, to)
	if exists {
		// Update existing holding - weighted average cost basis
		weightedCostBasis := toHolding.CostBasis.MulInt(toHolding.Shares).Add(fromHolding.CostBasis.MulInt(shares))
		newTotalShares := toHolding.Shares.Add(shares)
		newCostBasis := weightedCostBasis.QuoInt(newTotalShares)
		
		toHolding.Shares = newTotalShares
		toHolding.VestedShares = toHolding.VestedShares.Add(shares)
		toHolding.CostBasis = newCostBasis
		toHolding.TotalCost = toHolding.TotalCost.Add(fromHolding.CostBasis.MulInt(shares))
	} else {
		// Create new holding
		toHolding = types.NewShareholding(companyID, classID, to, shares, fromHolding.CostBasis)
	}
	
	// Save updates
	if fromHolding.Shares.IsZero() {
		k.DeleteShareholding(ctx, companyID, classID, from)
	} else {
		if err := k.SetShareholding(ctx, fromHolding); err != nil {
			return err
		}
	}
	if err := k.SetShareholding(ctx, toHolding); err != nil {
		return err
	}

	return nil
}

// GetTotalShares returns total outstanding shares for a share class
func (k Keeper) GetTotalShares(ctx sdk.Context, companyID uint64, classID string) math.Int {
	shareClass, found := k.getShareClass(ctx, companyID, classID)
	if !found {
		return math.ZeroInt()
	}
	return shareClass.OutstandingShares
}

// GetShareClass interface-compatible version
func (k Keeper) GetShareClass(ctx context.Context, companyID uint64, classID string) (interface{}, bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	shareClass, found := k.getShareClass(sdkCtx, companyID, classID)
	return shareClass, found
}

func (k Keeper) getShareClass(ctx sdk.Context, companyID uint64, classID string) (types.ShareClass, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareClassKey(companyID, classID)
	bz := store.Get(key)
	if bz == nil {
		return types.ShareClass{}, false
	}
	
	var shareClass types.ShareClass
	if err := json.Unmarshal(bz, &shareClass); err != nil {
		return types.ShareClass{}, false
	}
	return shareClass, true
}

// GetShareholding interface-compatible version
func (k Keeper) GetShareholding(ctx context.Context, companyID uint64, classID, owner string) (interface{}, bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	shareholding, found := k.getShareholding(sdkCtx, companyID, classID, owner)
	return shareholding, found
}

func (k Keeper) getShareholding(ctx sdk.Context, companyID uint64, classID, owner string) (types.Shareholding, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholdingKey(companyID, classID, owner)
	bz := store.Get(key)
	if bz == nil {
		return types.Shareholding{}, false
	}

	var shareholding types.Shareholding
	if err := json.Unmarshal(bz, &shareholding); err != nil {
		return types.Shareholding{}, false
	}
	return shareholding, true
}

// GetVotingSharesForCompany returns total voting shares for a voter in a specific company
// This includes:
// 1. Direct shareholdings (shares held directly by the voter)
// 2. Beneficial ownership (shares locked in DEX/escrow/lending but still owned by voter)
// Used by governance module for company proposal voting power
func (k Keeper) GetVotingSharesForCompany(ctx sdk.Context, companyID uint64, voter string) math.Int {
	totalShares := math.ZeroInt()

	// 1. Get direct shareholdings across all share classes
	shareholdings := k.GetOwnerShareholdings(ctx, voter)
	for _, sh := range shareholdings {
		if sh.CompanyID == companyID {
			totalShares = totalShares.Add(sh.Shares)
		}
	}

	// 2. Get beneficial ownership (shares locked in module accounts)
	// These are shares the voter owns but are locked in DEX pools, escrow, or lending
	beneficialOwnerships := k.GetAllBeneficialOwnershipsByOwner(ctx, voter)
	for _, bo := range beneficialOwnerships {
		if bo.CompanyID == companyID {
			totalShares = totalShares.Add(bo.Shares)
		}
	}

	return totalShares
}

// GetAllBeneficialOwnershipsByOwner returns all beneficial ownership records for an owner
// across all companies and share classes
func (k Keeper) GetAllBeneficialOwnershipsByOwner(ctx sdk.Context, owner string) []types.BeneficialOwnership {
	var ownerships []types.BeneficialOwnership

	// Iterate through all beneficial ownership records
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.BeneficialOwnerPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var bo types.BeneficialOwnership
		if err := json.Unmarshal(iterator.Value(), &bo); err != nil {
			continue
		}
		if bo.BeneficialOwner == owner {
			ownerships = append(ownerships, bo)
		}
	}

	return ownerships
}

// CreateCompany creates a new company - business logic version
func (k Keeper) CreateCompany(
	ctx sdk.Context,
	name, legalEntityName, registrationNumber, jurisdiction, industry string,
	description, symbol, founder string,
) (uint64, error) {
	// Check if symbol is already taken
	if k.IsSymbolTaken(ctx, symbol) {
		return 0, types.ErrSymbolTaken
	}
	
	// Get next company ID
	companyID := k.GetNextCompanyID(ctx)
	
	// Create company with minimal parameters
	company := types.Company{
		ID:          companyID,
		Name:        name,
		Symbol:      symbol,
		Description: description,
		Founder:     founder,
		Website:     "",
		Industry:    industry,
		Country:     jurisdiction,
		Status:      types.CompanyStatusActive,
		CreatedAt:   ctx.BlockTime(),
		UpdatedAt:   ctx.BlockTime(),
	}
	
	// Validate and store
	if err := company.Validate(); err != nil {
		return 0, err
	}

	if err := k.SetCompany(ctx, company); err != nil {
		return 0, err
	}
	return companyID, nil
}

// CreateShareClass creates a new share class - business logic version
func (k Keeper) CreateShareClass(
	ctx sdk.Context,
	companyID uint64,
	classID, name string,
	authorizedShares math.Int,
	votingMultiplier math.LegacyDec,
	transferable bool,
	creator string,
) error {
	// Check if company exists
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}
	
	// Only founder can create share classes
	if company.Founder != creator {
		return types.ErrUnauthorized
	}
	
	// Check if share class already exists
	if _, exists := k.getShareClass(ctx, companyID, classID); exists {
		return types.ErrShareClassAlreadyExists
	}
	
	// Create share class
	shareClass := types.ShareClass{
		CompanyID:             companyID,
		ClassID:               classID,
		Name:                  name,
		Description:           "",
		VotingRights:          true,
		DividendRights:        true,
		LiquidationPreference: false,
		ConversionRights:      false,
		AuthorizedShares:      authorizedShares,
		IssuedShares:          math.ZeroInt(),
		OutstandingShares:     math.ZeroInt(),
		ParValue:              math.LegacyNewDec(1),
		Transferable:          transferable,
		CreatedAt:             ctx.BlockTime(),
		UpdatedAt:             ctx.BlockTime(),
	}
	
	// Validate and store
	if err := shareClass.Validate(); err != nil {
		return err
	}

	if err := k.SetShareClass(ctx, shareClass); err != nil {
		return err
	}
	return nil
}

// GetAllHoldingsByAddress returns all shareholdings for a given address
// Used by inheritance module to transfer equity during inheritance
// Returns []interface{} to avoid import cycles with inheritance module
func (k Keeper) GetAllHoldingsByAddress(ctx sdk.Context, owner string) []interface{} {
	var holdings []interface{}

	// Iterate through all shareholdings
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ShareholdingPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var shareholding types.Shareholding
		if err := json.Unmarshal(iterator.Value(), &shareholding); err != nil {
			k.Logger(ctx).Error("failed to unmarshal shareholding", "error", err)
			continue
		}

		// Filter by owner
		if shareholding.Owner == owner {
			holdings = append(holdings, types.AddressHolding{
				CompanyID: shareholding.CompanyID,
				ClassID:   shareholding.ClassID,
				Shares:    shareholding.Shares,
			})
		}
	}

	return holdings
}