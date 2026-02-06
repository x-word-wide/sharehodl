package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

// Keeper of the hodl store
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	memKey     storetypes.StoreKey
	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper

	// Authority for governance-controlled operations
	authority string
}

// Key prefixes for additional storage
var (
	CollateralWhitelistPrefix = []byte{0x12}
	PriceOraclePrefix         = []byte{0x13}
	BadDebtLimitKey           = []byte{0x14}
)

// GetMaxBadDebtLimit returns the governance-controllable bad debt circuit breaker limit
func (k Keeper) GetMaxBadDebtLimit(ctx sdk.Context) math.Int {
	params := k.GetParams(ctx)
	if params.MaxBadDebtLimit.IsNil() || !params.MaxBadDebtLimit.IsPositive() {
		return types.DefaultMaxBadDebtLimit
	}
	return params.MaxBadDebtLimit
}

// GetLiquidationPenalty returns the governance-controllable liquidation penalty
func (k Keeper) GetLiquidationPenalty(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.LiquidationPenalty.IsNil() {
		return math.LegacyNewDecWithPrec(10, 2) // 10% default
	}
	return params.LiquidationPenalty
}

// GetLiquidatorReward returns the governance-controllable liquidator reward
func (k Keeper) GetLiquidatorReward(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.LiquidatorReward.IsNil() {
		return math.LegacyNewDecWithPrec(5, 2) // 5% default
	}
	return params.LiquidatorReward
}

// GetMaxPriceDeviation returns the governance-controllable max price deviation
func (k Keeper) GetMaxPriceDeviation(ctx sdk.Context) math.LegacyDec {
	params := k.GetParams(ctx)
	if params.MaxPriceDeviation.IsNil() {
		return math.LegacyNewDecWithPrec(20, 2) // 20% default
	}
	return params.MaxPriceDeviation
}

// GetBlocksPerYear returns the governance-controllable blocks per year
func (k Keeper) GetBlocksPerYear(ctx sdk.Context) uint64 {
	params := k.GetParams(ctx)
	if params.BlocksPerYear == 0 {
		return types.DefaultBlocksPerYear
	}
	return params.BlocksPerYear
}

// NewKeeper creates a new hodl Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		authority:     authority,
	}
}

// GetAuthority returns the governance authority address
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetBalance returns the HODL balance for an address (governance compatibility)
func (k Keeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress) math.Int {
	coin := k.bankKeeper.GetBalance(ctx, addr, types.HODLDenom)
	return coin.Amount
}

// SendCoins sends HODL tokens from one address to another (governance compatibility)
func (k Keeper) SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount math.Int) error {
	coins := sdk.NewCoins(sdk.NewCoin(types.HODLDenom, amount))
	return k.bankKeeper.SendCoins(ctx, fromAddr, toAddr, coins)
}

// BurnTokens burns HODL tokens from an address (governance compatibility)
func (k Keeper) BurnTokens(ctx sdk.Context, addr sdk.AccAddress, amount math.Int) error {
	coins := sdk.NewCoins(sdk.NewCoin(types.HODLDenom, amount))
	// First send to module account, then burn
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, coins); err != nil {
		return err
	}
	return k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins)
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}
	
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)
}

// GetTotalSupply returns the total supply of HODL tokens (interface compatible)
func (k Keeper) GetTotalSupply(ctx context.Context) interface{} {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return k.getTotalSupply(sdkCtx)
}

// getTotalSupply returns the total supply of HODL tokens  
func (k Keeper) getTotalSupply(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.MintedSupplyKey)
	if bz == nil {
		return math.ZeroInt()
	}
	
	var supplyProto types.Supply
	k.cdc.MustUnmarshal(bz, &supplyProto)
	return supplyProto.Amount
}

// SetTotalSupply sets the total supply of HODL tokens
func (k Keeper) SetTotalSupply(ctx sdk.Context, supply math.Int) {
	store := ctx.KVStore(k.storeKey)
	supplyProto := types.Supply{Amount: supply}
	bz := k.cdc.MustMarshal(&supplyProto)
	store.Set(types.MintedSupplyKey, bz)
}

// GetCollateralPosition returns a user's collateral position
func (k Keeper) GetCollateralPosition(ctx sdk.Context, owner sdk.AccAddress) (types.CollateralPosition, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CollateralPositionKey(owner))
	if bz == nil {
		return types.CollateralPosition{}, false
	}
	
	var position types.CollateralPosition
	k.cdc.MustUnmarshal(bz, &position)
	return position, true
}

// SetCollateralPosition sets a user's collateral position
func (k Keeper) SetCollateralPosition(ctx sdk.Context, position types.CollateralPosition) {
	store := ctx.KVStore(k.storeKey)
	owner, _ := sdk.AccAddressFromBech32(position.Owner)
	bz := k.cdc.MustMarshal(&position)
	store.Set(types.CollateralPositionKey(owner), bz)
}

// DeleteCollateralPosition deletes a user's collateral position
func (k Keeper) DeleteCollateralPosition(ctx sdk.Context, owner sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.CollateralPositionKey(owner))
}

// IterateCollateralPositions iterates over all collateral positions
func (k Keeper) IterateCollateralPositions(ctx sdk.Context, cb func(position types.CollateralPosition) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.CollateralPositionPrefix)
	defer iterator.Close()
	
	for ; iterator.Valid(); iterator.Next() {
		var position types.CollateralPosition
		k.cdc.MustUnmarshal(iterator.Value(), &position)
		if cb(position) {
			break
		}
	}
}

// GetAllCollateralPositions returns all collateral positions
func (k Keeper) GetAllCollateralPositions(ctx sdk.Context) []types.CollateralPosition {
	positions := []types.CollateralPosition{}
	k.IterateCollateralPositions(ctx, func(position types.CollateralPosition) bool {
		positions = append(positions, position)
		return false
	})
	return positions
}

// =============================================================================
// LIQUIDATION MECHANISM
// =============================================================================

// NOTE: LiquidationPenalty and LiquidatorReward are now governance-controllable
// Use k.GetLiquidationPenalty(ctx) and k.GetLiquidatorReward(ctx) instead

// GetCollateralRatio calculates the collateral ratio for a position
// Returns the ratio as a decimal (e.g., 1.5 = 150%)
// SECURITY FIX: Includes stability debt in total debt calculation
func (k Keeper) GetCollateralRatio(ctx sdk.Context, position types.CollateralPosition) (math.LegacyDec, error) {
	// Calculate total debt = minted HODL + stability debt
	totalDebt := math.LegacyNewDecFromInt(position.MintedHODL).Add(position.StabilityDebt)

	// If no debt, return max ratio (position is fully collateralized)
	if totalDebt.IsZero() || totalDebt.IsNegative() {
		return math.LegacyNewDec(1000000), nil // Effectively infinite ratio
	}

	// Get collateral value in HODL terms
	collateralValue, err := k.GetCollateralValue(ctx, position.Collateral)
	if err != nil {
		return math.LegacyDec{}, err
	}

	// Calculate ratio: collateral_value / total_debt
	ratio := collateralValue.Quo(totalDebt)
	return ratio, nil
}

// GetTotalDebt returns the total debt for a position (minted + stability fees)
func (k Keeper) GetTotalDebt(ctx sdk.Context, position types.CollateralPosition) math.LegacyDec {
	return math.LegacyNewDecFromInt(position.MintedHODL).Add(position.StabilityDebt)
}

// GetCollateralValue returns the total value of collateral in HODL terms
func (k Keeper) GetCollateralValue(ctx sdk.Context, collateral sdk.Coins) (math.LegacyDec, error) {
	totalValue := math.LegacyZeroDec()

	for _, coin := range collateral {
		price, err := k.GetCollateralPrice(ctx, coin.Denom)
		if err != nil {
			return math.LegacyDec{}, fmt.Errorf("failed to get price for %s: %w", coin.Denom, err)
		}
		coinValue := math.LegacyNewDecFromInt(coin.Amount).Mul(price)
		totalValue = totalValue.Add(coinValue)
	}

	return totalValue, nil
}

// GetCollateralPrice returns the price of a collateral asset in HODL terms
// SECURITY: Only returns price for whitelisted collateral types
func (k Keeper) GetCollateralPrice(ctx sdk.Context, denom string) (math.LegacyDec, error) {
	// Native HODL is always 1:1
	if denom == "uhodl" || denom == "hodl" {
		return math.LegacyOneDec(), nil
	}

	// Check if collateral is whitelisted
	if !k.IsCollateralWhitelisted(ctx, denom) {
		return math.LegacyDec{}, fmt.Errorf("collateral %s is not whitelisted", denom)
	}

	store := ctx.KVStore(k.storeKey)
	priceKey := append([]byte{0x10}, []byte(denom)...)
	bz := store.Get(priceKey)

	if bz == nil {
		return math.LegacyDec{}, fmt.Errorf("no oracle price set for %s", denom)
	}

	var price math.LegacyDec
	if err := price.Unmarshal(bz); err != nil {
		return math.LegacyDec{}, fmt.Errorf("failed to unmarshal price for %s: %w", denom, err)
	}

	// Validate price is positive and reasonable
	if price.IsNil() || price.IsZero() || price.IsNegative() {
		return math.LegacyDec{}, fmt.Errorf("invalid price for %s: %s", denom, price.String())
	}

	return price, nil
}

// SetCollateralPrice sets the price of a collateral asset
// SECURITY: Requires governance authority
func (k Keeper) SetCollateralPrice(ctx sdk.Context, sender string, denom string, price math.LegacyDec) error {
	// Verify sender is governance authority
	if sender != k.authority {
		return fmt.Errorf("unauthorized: only governance can set collateral prices")
	}

	// Validate price
	if price.IsNil() || price.IsZero() || price.IsNegative() {
		return fmt.Errorf("invalid price: must be positive")
	}

	// Check collateral is whitelisted
	if !k.IsCollateralWhitelisted(ctx, denom) {
		return fmt.Errorf("collateral %s is not whitelisted", denom)
	}

	// Get current price to check deviation (governance-controllable limit)
	currentPrice, err := k.GetCollateralPrice(ctx, denom)
	if err == nil && !currentPrice.IsZero() {
		deviation := price.Sub(currentPrice).Abs().Quo(currentPrice)
		maxDeviation := k.GetMaxPriceDeviation(ctx)
		if deviation.GT(maxDeviation) {
			return fmt.Errorf("price deviation too large: %s (max %s)", deviation.String(), maxDeviation.String())
		}
	}

	store := ctx.KVStore(k.storeKey)
	priceKey := append([]byte{0x10}, []byte(denom)...)
	bz, _ := price.Marshal()
	store.Set(priceKey, bz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hodl_price_updated",
			sdk.NewAttribute("denom", denom),
			sdk.NewAttribute("price", price.String()),
			sdk.NewAttribute("updater", sender),
		),
	)

	return nil
}

// =============================================================================
// COLLATERAL WHITELIST MANAGEMENT
// =============================================================================

// IsCollateralWhitelisted checks if a collateral type is whitelisted
func (k Keeper) IsCollateralWhitelisted(ctx sdk.Context, denom string) bool {
	// Native HODL is always accepted
	if denom == "uhodl" || denom == "hodl" {
		return true
	}

	store := ctx.KVStore(k.storeKey)
	key := append(CollateralWhitelistPrefix, []byte(denom)...)
	return store.Has(key)
}

// AddCollateralToWhitelist adds a collateral type to the whitelist
// SECURITY: Requires governance authority
func (k Keeper) AddCollateralToWhitelist(ctx sdk.Context, sender string, denom string, initialPrice math.LegacyDec) error {
	if sender != k.authority {
		return fmt.Errorf("unauthorized: only governance can whitelist collateral")
	}

	if k.IsCollateralWhitelisted(ctx, denom) {
		return fmt.Errorf("collateral %s is already whitelisted", denom)
	}

	// Validate initial price
	if initialPrice.IsNil() || initialPrice.IsZero() || initialPrice.IsNegative() {
		return fmt.Errorf("invalid initial price: must be positive")
	}

	// Add to whitelist
	store := ctx.KVStore(k.storeKey)
	key := append(CollateralWhitelistPrefix, []byte(denom)...)
	store.Set(key, []byte{1})

	// Set initial price
	priceKey := append([]byte{0x10}, []byte(denom)...)
	bz, _ := initialPrice.Marshal()
	store.Set(priceKey, bz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hodl_collateral_whitelisted",
			sdk.NewAttribute("denom", denom),
			sdk.NewAttribute("initial_price", initialPrice.String()),
		),
	)

	return nil
}

// RemoveCollateralFromWhitelist removes a collateral type from the whitelist
// SECURITY: Requires governance authority
func (k Keeper) RemoveCollateralFromWhitelist(ctx sdk.Context, sender string, denom string) error {
	if sender != k.authority {
		return fmt.Errorf("unauthorized: only governance can remove collateral from whitelist")
	}

	if !k.IsCollateralWhitelisted(ctx, denom) {
		return fmt.Errorf("collateral %s is not whitelisted", denom)
	}

	// Check no positions use this collateral
	hasPositions := false
	k.IterateCollateralPositions(ctx, func(position types.CollateralPosition) bool {
		for _, coin := range position.Collateral {
			if coin.Denom == denom && coin.Amount.IsPositive() {
				hasPositions = true
				return true
			}
		}
		return false
	})

	if hasPositions {
		return fmt.Errorf("cannot remove collateral %s: positions still exist", denom)
	}

	// Remove from whitelist
	store := ctx.KVStore(k.storeKey)
	key := append(CollateralWhitelistPrefix, []byte(denom)...)
	store.Delete(key)

	// Remove price
	priceKey := append([]byte{0x10}, []byte(denom)...)
	store.Delete(priceKey)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hodl_collateral_removed",
			sdk.NewAttribute("denom", denom),
		),
	)

	return nil
}

// GetWhitelistedCollaterals returns all whitelisted collateral types
func (k Keeper) GetWhitelistedCollaterals(ctx sdk.Context) []string {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, CollateralWhitelistPrefix)
	defer iterator.Close()

	var collaterals []string
	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Key()[len(CollateralWhitelistPrefix):])
		collaterals = append(collaterals, denom)
	}
	return collaterals
}

// IsPositionLiquidatable checks if a position is below the liquidation threshold
func (k Keeper) IsPositionLiquidatable(ctx sdk.Context, position types.CollateralPosition) (bool, math.LegacyDec, error) {
	params := k.GetParams(ctx)

	ratio, err := k.GetCollateralRatio(ctx, position)
	if err != nil {
		return false, math.LegacyDec{}, err
	}

	// Position is liquidatable if ratio < liquidation_ratio (default 130%)
	return ratio.LT(params.LiquidationRatio), ratio, nil
}

// LiquidatePosition liquidates an undercollateralized position
func (k Keeper) LiquidatePosition(ctx sdk.Context, liquidator sdk.AccAddress, ownerAddr sdk.AccAddress) error {
	position, found := k.GetCollateralPosition(ctx, ownerAddr)
	if !found {
		return fmt.Errorf("position not found for owner %s", ownerAddr.String())
	}

	// Check if position is liquidatable
	liquidatable, ratio, err := k.IsPositionLiquidatable(ctx, position)
	if err != nil {
		return err
	}
	if !liquidatable {
		return fmt.Errorf("position is not liquidatable, current ratio: %s", ratio.String())
	}

	// Calculate amounts - SECURITY FIX: Include stability debt in total debt
	principalDebt := position.MintedHODL
	totalDebt := math.LegacyNewDecFromInt(principalDebt).Add(position.StabilityDebt)

	// Get governance-controllable liquidation parameters
	liquidationPenalty := k.GetLiquidationPenalty(ctx)
	liquidatorReward := k.GetLiquidatorReward(ctx)
	debtWithPenalty := totalDebt.Mul(math.LegacyOneDec().Add(liquidationPenalty))

	// Calculate collateral to seize
	collateralValue, _ := k.GetCollateralValue(ctx, position.Collateral)

	// Determine how much collateral to give to liquidator
	var collateralToLiquidator sdk.Coins
	var collateralToProtocol sdk.Coins
	var collateralToOwner sdk.Coins

	// If collateral covers debt + penalty fully
	if collateralValue.GTE(debtWithPenalty) {
		// Liquidator gets debt value + liquidator reward
		liquidatorAmount := totalDebt.Mul(math.LegacyOneDec().Add(liquidatorReward))
		protocolAmount := totalDebt.Mul(liquidationPenalty.Sub(liquidatorReward))
		ownerAmount := collateralValue.Sub(liquidatorAmount).Sub(protocolAmount)

		// Distribute collateral proportionally
		for _, coin := range position.Collateral {
			price, _ := k.GetCollateralPrice(ctx, coin.Denom)
			coinValue := math.LegacyNewDecFromInt(coin.Amount).Mul(price)

			liqShare := liquidatorAmount.Quo(collateralValue).Mul(math.LegacyNewDecFromInt(coin.Amount))
			protShare := protocolAmount.Quo(collateralValue).Mul(math.LegacyNewDecFromInt(coin.Amount))
			ownerShare := ownerAmount.Quo(collateralValue).Mul(math.LegacyNewDecFromInt(coin.Amount))

			if liqShare.IsPositive() {
				collateralToLiquidator = collateralToLiquidator.Add(sdk.NewCoin(coin.Denom, liqShare.TruncateInt()))
			}
			if protShare.IsPositive() {
				collateralToProtocol = collateralToProtocol.Add(sdk.NewCoin(coin.Denom, protShare.TruncateInt()))
			}
			if ownerShare.IsPositive() && coinValue.IsPositive() {
				collateralToOwner = collateralToOwner.Add(sdk.NewCoin(coin.Denom, ownerShare.TruncateInt()))
			}
		}
	} else {
		// Collateral doesn't cover debt fully - liquidator gets everything, bad debt recorded
		collateralToLiquidator = position.Collateral
		k.RecordBadDebt(ctx, debtWithPenalty.Sub(collateralValue))
	}

	// Liquidator must provide HODL to cover the total debt (principal + stability fees)
	totalDebtInt := totalDebt.TruncateInt()
	hodlCoin := sdk.NewCoin("uhodl", totalDebtInt)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, liquidator, types.ModuleName, sdk.NewCoins(hodlCoin)); err != nil {
		return fmt.Errorf("liquidator doesn't have enough HODL: %w", err)
	}

	// Burn the HODL received from liquidator
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(hodlCoin)); err != nil {
		return fmt.Errorf("failed to burn HODL: %w", err)
	}

	// Update total supply (only principal affects supply, not stability fees)
	currentSupply := k.getTotalSupply(ctx)
	k.SetTotalSupply(ctx, currentSupply.Sub(principalDebt))

	// Transfer collateral to liquidator
	if collateralToLiquidator.IsValid() && !collateralToLiquidator.IsZero() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, liquidator, collateralToLiquidator); err != nil {
			return fmt.Errorf("failed to send collateral to liquidator: %w", err)
		}
	}

	// Return remaining collateral to owner if any
	if collateralToOwner.IsValid() && !collateralToOwner.IsZero() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, ownerAddr, collateralToOwner); err != nil {
			return fmt.Errorf("failed to return collateral to owner: %w", err)
		}
	}

	// Protocol fee goes to community pool or fee collector
	if collateralToProtocol.IsValid() && !collateralToProtocol.IsZero() {
		// Send to fee collector or burn
		feeCollector := k.accountKeeper.GetModuleAddress("fee_collector")
		if feeCollector != nil {
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, feeCollector, collateralToProtocol); err != nil {
				// Log error but don't fail the liquidation
				k.Logger(ctx).Error("failed to send protocol fee", "error", err)
			}
		}
	}

	// Delete the liquidated position
	k.DeleteCollateralPosition(ctx, ownerAddr)

	// Emit liquidation event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hodl_liquidation",
			sdk.NewAttribute("owner", position.Owner),
			sdk.NewAttribute("liquidator", liquidator.String()),
			sdk.NewAttribute("principal_debt", principalDebt.String()),
			sdk.NewAttribute("stability_debt", position.StabilityDebt.String()),
			sdk.NewAttribute("total_debt", totalDebt.String()),
			sdk.NewAttribute("collateral_ratio", ratio.String()),
			sdk.NewAttribute("collateral_to_liquidator", collateralToLiquidator.String()),
			sdk.NewAttribute("collateral_returned_to_owner", collateralToOwner.String()),
		),
	)

	k.Logger(ctx).Info("position liquidated",
		"owner", position.Owner,
		"liquidator", liquidator.String(),
		"total_debt", totalDebt.String(),
		"ratio", ratio.String(),
	)

	return nil
}

// CheckAndLiquidatePositions iterates through all positions and liquidates those below threshold
// This is called at the end of each block
func (k Keeper) CheckAndLiquidatePositions(ctx sdk.Context) {
	var positionsToLiquidate []types.CollateralPosition

	k.IterateCollateralPositions(ctx, func(position types.CollateralPosition) bool {
		liquidatable, ratio, err := k.IsPositionLiquidatable(ctx, position)
		if err != nil {
			k.Logger(ctx).Error("failed to check position", "owner", position.Owner, "error", err)
			return false
		}

		if liquidatable {
			k.Logger(ctx).Info("position eligible for liquidation",
				"owner", position.Owner,
				"ratio", ratio.String(),
			)
			positionsToLiquidate = append(positionsToLiquidate, position)
		}
		return false
	})

	// Emit event for positions that need liquidation
	for _, position := range positionsToLiquidate {
		ratio, _ := k.GetCollateralRatio(ctx, position)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"hodl_liquidation_warning",
				sdk.NewAttribute("owner", position.Owner),
				sdk.NewAttribute("collateral_ratio", ratio.String()),
				sdk.NewAttribute("debt_amount", position.MintedHODL.String()),
			),
		)
	}
}

// RecordBadDebt records bad debt that occurred during liquidation
// SECURITY: Includes circuit breaker for excessive bad debt
func (k Keeper) RecordBadDebt(ctx sdk.Context, amount math.LegacyDec) {
	store := ctx.KVStore(k.storeKey)
	badDebtKey := []byte{0x11}

	var currentBadDebt math.LegacyDec
	bz := store.Get(badDebtKey)
	if bz != nil {
		if err := currentBadDebt.Unmarshal(bz); err != nil {
			currentBadDebt = math.LegacyZeroDec()
		}
	} else {
		currentBadDebt = math.LegacyZeroDec()
	}

	newBadDebt := currentBadDebt.Add(amount)
	newBz, _ := newBadDebt.Marshal()
	store.Set(badDebtKey, newBz)

	// Circuit breaker check (governance-controllable limit)
	maxBadDebt := k.GetMaxBadDebtLimit(ctx)
	if newBadDebt.TruncateInt().GT(maxBadDebt) {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"hodl_circuit_breaker_triggered",
				sdk.NewAttribute("bad_debt", newBadDebt.String()),
				sdk.NewAttribute("limit", maxBadDebt.String()),
			),
		)
		k.Logger(ctx).Error("CIRCUIT BREAKER: Bad debt limit exceeded!",
			"bad_debt", newBadDebt.String(),
			"limit", maxBadDebt.String(),
		)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hodl_bad_debt",
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("total_bad_debt", newBadDebt.String()),
		),
	)
}

// IsMintingPaused checks if minting is paused due to bad debt circuit breaker
func (k Keeper) IsMintingPaused(ctx sdk.Context) bool {
	badDebt := k.GetBadDebt(ctx)
	return badDebt.TruncateInt().GT(k.GetMaxBadDebtLimit(ctx))
}

// GetBadDebt returns the total accumulated bad debt
func (k Keeper) GetBadDebt(ctx sdk.Context) math.LegacyDec {
	store := ctx.KVStore(k.storeKey)
	badDebtKey := []byte{0x11}

	bz := store.Get(badDebtKey)
	if bz == nil {
		return math.LegacyZeroDec()
	}

	var badDebt math.LegacyDec
	badDebt.Unmarshal(bz)
	return badDebt
}

// =============================================================================
// STABILITY FEE ACCRUAL
// =============================================================================

// AccrueStabilityFees accrues stability fees for a position
func (k Keeper) AccrueStabilityFees(ctx sdk.Context, position *types.CollateralPosition) {
	params := k.GetParams(ctx)

	if params.StabilityFee.IsZero() {
		return
	}

	// Calculate blocks since last update
	blocksSinceUpdate := ctx.BlockHeight() - position.LastUpdated
	if blocksSinceUpdate <= 0 {
		return
	}

	// Calculate fee: principal * rate * time (in blocks)
	// BlocksPerYear is governance-controllable (default ~5,256,000 at 6s blocks)
	blocksPerYear := math.LegacyNewDec(int64(k.GetBlocksPerYear(ctx)))
	timeRatio := math.LegacyNewDec(blocksSinceUpdate).Quo(blocksPerYear)

	fee := math.LegacyNewDecFromInt(position.MintedHODL).Mul(params.StabilityFee).Mul(timeRatio)
	position.StabilityDebt = position.StabilityDebt.Add(fee)
	position.LastUpdated = ctx.BlockHeight()
}

// AccrueAllStabilityFees accrues stability fees for all positions
// Called at the end of each block
func (k Keeper) AccrueAllStabilityFees(ctx sdk.Context) {
	k.IterateCollateralPositions(ctx, func(position types.CollateralPosition) bool {
		k.AccrueStabilityFees(ctx, &position)
		k.SetCollateralPosition(ctx, position)
		return false
	})
}

// PayStabilityDebt allows a user to pay off their stability debt
func (k Keeper) PayStabilityDebt(ctx sdk.Context, owner sdk.AccAddress, amount math.Int) error {
	position, found := k.GetCollateralPosition(ctx, owner)
	if !found {
		return fmt.Errorf("position not found for owner %s", owner.String())
	}

	// Accrue fees first
	k.AccrueStabilityFees(ctx, &position)

	// Check if payment exceeds debt
	debtInt := position.StabilityDebt.TruncateInt()
	if amount.GT(debtInt) {
		amount = debtInt
	}

	// Transfer HODL from user to module
	hodlCoin := sdk.NewCoin("uhodl", amount)
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, sdk.NewCoins(hodlCoin)); err != nil {
		return err
	}

	// Burn the HODL
	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(hodlCoin)); err != nil {
		return err
	}

	// Update position
	position.StabilityDebt = position.StabilityDebt.Sub(math.LegacyNewDecFromInt(amount))
	k.SetCollateralPosition(ctx, position)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hodl_stability_debt_paid",
			sdk.NewAttribute("owner", owner.String()),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("remaining_debt", position.StabilityDebt.String()),
		),
	)

	return nil
}

// =============================================================================
// ADDITIONAL COLLATERAL MANAGEMENT
// =============================================================================

// AddCollateral allows a user to add more collateral to their position
func (k Keeper) AddCollateral(ctx sdk.Context, owner sdk.AccAddress, collateral sdk.Coins) error {
	position, found := k.GetCollateralPosition(ctx, owner)
	if !found {
		return fmt.Errorf("position not found for owner %s", owner.String())
	}

	// Transfer collateral from user to module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, collateral); err != nil {
		return err
	}

	// Update position
	position.Collateral = position.Collateral.Add(collateral...)
	position.LastUpdated = ctx.BlockHeight()
	k.SetCollateralPosition(ctx, position)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hodl_collateral_added",
			sdk.NewAttribute("owner", owner.String()),
			sdk.NewAttribute("amount", collateral.String()),
			sdk.NewAttribute("total_collateral", position.Collateral.String()),
		),
	)

	return nil
}

// WithdrawCollateral allows a user to withdraw excess collateral
func (k Keeper) WithdrawCollateral(ctx sdk.Context, owner sdk.AccAddress, collateral sdk.Coins) error {
	position, found := k.GetCollateralPosition(ctx, owner)
	if !found {
		return fmt.Errorf("position not found for owner %s", owner.String())
	}

	// Check if user has enough collateral
	newCollateral, hasNeg := position.Collateral.SafeSub(collateral...)
	if hasNeg {
		return fmt.Errorf("insufficient collateral")
	}

	// Check if withdrawal would make position undercollateralized
	testPosition := position
	testPosition.Collateral = newCollateral

	params := k.GetParams(ctx)
	ratio, err := k.GetCollateralRatio(ctx, testPosition)
	if err != nil {
		return err
	}

	if ratio.LT(params.CollateralRatio) {
		return fmt.Errorf("withdrawal would make position undercollateralized, min ratio: %s, after withdrawal: %s",
			params.CollateralRatio.String(), ratio.String())
	}

	// Transfer collateral from module to user
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, collateral); err != nil {
		return err
	}

	// Update position
	position.Collateral = newCollateral
	position.LastUpdated = ctx.BlockHeight()
	k.SetCollateralPosition(ctx, position)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hodl_collateral_withdrawn",
			sdk.NewAttribute("owner", owner.String()),
			sdk.NewAttribute("amount", collateral.String()),
			sdk.NewAttribute("remaining_collateral", position.Collateral.String()),
		),
	)

	return nil
}