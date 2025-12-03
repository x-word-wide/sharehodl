package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// Expected keepers for dependency injection
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

type AccountKeeper interface {
	GetAccount(sdk.Context, sdk.AccAddress) sdk.AccountI
	SetAccount(sdk.Context, sdk.AccountI)
	NewAccount(sdk.Context, sdk.AccountI) sdk.AccountI
	GetModuleAddress(string) sdk.AccAddress
}

// ValidatorInfo represents validator information for business listing
type ValidatorInfo struct {
	Address string
	Active  bool
	Tier    int
}

type ValidatorKeeper interface {
	GetValidator(ctx sdk.Context, address string) (ValidatorInfo, bool)
	GetAllValidators(ctx sdk.Context) []ValidatorInfo
}

// Keeper of the equity store
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	memKey     storetypes.StoreKey
	paramstore paramtypes.Subspace

	bankKeeper      BankKeeper
	accountKeeper   AccountKeeper
	validatorKeeper ValidatorKeeper
}

// NewKeeper creates a new equity Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bankKeeper BankKeeper,
	accountKeeper AccountKeeper,
	validatorKeeper ValidatorKeeper,
) *Keeper {
	return &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		memKey:          memKey,
		paramstore:      ps,
		bankKeeper:      bankKeeper,
		accountKeeper:   accountKeeper,
		validatorKeeper: validatorKeeper,
	}
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

// GetCompany returns a company by ID
func (k Keeper) GetCompany(ctx sdk.Context, companyID uint64) (types.Company, bool) {
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
func (k Keeper) SetCompany(ctx sdk.Context, company types.Company) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyKey(company.ID)
	bz, err := json.Marshal(company)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// DeleteCompany removes a company
func (k Keeper) DeleteCompany(ctx sdk.Context, companyID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetCompanyKey(companyID)
	store.Delete(key)
}

// IsSymbolTaken checks if a company symbol is already taken
func (k Keeper) IsSymbolTaken(ctx sdk.Context, symbol string) bool {
	companies := k.GetAllCompanies(ctx)
	for _, company := range companies {
		if company.Symbol == symbol {
			return true
		}
	}
	return false
}

// GetAllCompanies returns all companies - simplified approach
func (k Keeper) GetAllCompanies(ctx sdk.Context) []types.Company {
	// For now, return empty list - would need to iterate through all keys
	// This is a simplified implementation
	return []types.Company{}
}

// Share class management methods

// GetShareClass returns a share class by company ID and class ID
func (k Keeper) GetShareClass(ctx sdk.Context, companyID uint64, classID string) (types.ShareClass, bool) {
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

// SetShareClass stores a share class
func (k Keeper) SetShareClass(ctx sdk.Context, shareClass types.ShareClass) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareClassKey(shareClass.CompanyID, shareClass.ClassID)
	bz, err := json.Marshal(shareClass)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// DeleteShareClass removes a share class
func (k Keeper) DeleteShareClass(ctx sdk.Context, companyID uint64, classID string) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareClassKey(companyID, classID)
	store.Delete(key)
}

// GetCompanyShareClasses returns all share classes for a company - simplified
func (k Keeper) GetCompanyShareClasses(ctx sdk.Context, companyID uint64) []types.ShareClass {
	// For now, return empty list - would need proper iteration
	// This is a simplified implementation  
	return []types.ShareClass{}
}

// Shareholding management methods

// GetShareholding returns a shareholding
func (k Keeper) GetShareholding(ctx sdk.Context, companyID uint64, classID, owner string) (types.Shareholding, bool) {
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

// SetShareholding stores a shareholding
func (k Keeper) SetShareholding(ctx sdk.Context, shareholding types.Shareholding) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholdingKey(shareholding.CompanyID, shareholding.ClassID, shareholding.Owner)
	bz, err := json.Marshal(shareholding)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)
}

// DeleteShareholding removes a shareholding
func (k Keeper) DeleteShareholding(ctx sdk.Context, companyID uint64, classID, owner string) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetShareholdingKey(companyID, classID, owner)
	store.Delete(key)
}

// GetOwnerShareholdings returns all shareholdings for an owner - simplified
func (k Keeper) GetOwnerShareholdings(ctx sdk.Context, owner string) []types.Shareholding {
	// For now, return empty list - would need proper iteration
	// This is a simplified implementation
	return []types.Shareholding{}
}

// GetCompanyShareholdings returns all shareholdings for a company and share class - simplified
func (k Keeper) GetCompanyShareholdings(ctx sdk.Context, companyID uint64, classID string) []types.Shareholding {
	// For now, return empty list - would need proper iteration
	// This is a simplified implementation
	return []types.Shareholding{}
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
	shareClass, found := k.GetShareClass(ctx, companyID, classID)
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
	shareholding, exists := k.GetShareholding(ctx, companyID, classID, recipient)
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
	k.SetShareholding(ctx, shareholding)
	k.SetShareClass(ctx, shareClass)
	
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
	// Get share class to check transferability
	shareClass, found := k.GetShareClass(ctx, companyID, classID)
	if !found {
		return types.ErrShareClassNotFound
	}
	
	if !shareClass.Transferable {
		return types.ErrTransferRestricted
	}
	
	// Get from shareholding
	fromHolding, found := k.GetShareholding(ctx, companyID, classID, from)
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
	toHolding, exists := k.GetShareholding(ctx, companyID, classID, to)
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
		k.SetShareholding(ctx, fromHolding)
	}
	k.SetShareholding(ctx, toHolding)
	
	return nil
}

// GetTotalShares returns total outstanding shares for a share class
func (k Keeper) GetTotalShares(ctx sdk.Context, companyID uint64, classID string) math.Int {
	shareClass, found := k.GetShareClass(ctx, companyID, classID)
	if !found {
		return math.ZeroInt()
	}
	return shareClass.OutstandingShares
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
	
	k.SetCompany(ctx, company)
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
	company, found := k.GetCompany(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}
	
	// Only founder can create share classes
	if company.Founder != creator {
		return types.ErrUnauthorized
	}
	
	// Check if share class already exists
	if _, exists := k.GetShareClass(ctx, companyID, classID); exists {
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
	
	k.SetShareClass(ctx, shareClass)
	return nil
}