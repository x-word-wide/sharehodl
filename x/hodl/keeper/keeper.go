package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

// Keeper of the hodl store
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	memKey     storetypes.StoreKey
	paramstore paramtypes.Subspace

	bankKeeper    types.BankKeeper
	accountKeeper types.AccountKeeper
}

// NewKeeper creates a new hodl Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    ps,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// GetTotalSupply returns the total supply of HODL tokens
func (k Keeper) GetTotalSupply(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.MintedSupplyKey)
	if bz == nil {
		return math.ZeroInt()
	}
	
	var supply math.Int
	k.cdc.MustUnmarshal(bz, &supply)
	return supply
}

// SetTotalSupply sets the total supply of HODL tokens
func (k Keeper) SetTotalSupply(ctx sdk.Context, supply math.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&supply)
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
	iterator := sdk.KVStorePrefixIterator(store, types.CollateralPositionPrefix)
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