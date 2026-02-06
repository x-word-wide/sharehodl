package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
)

// Keeper of the bridge store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	memKey   storetypes.StoreKey

	bankKeeper      types.BankKeeper
	validatorKeeper types.ValidatorKeeper

	// Authority for governance-controlled operations
	authority string
}

// NewKeeper creates a new bridge Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	validatorKeeper types.ValidatorKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		memKey:          memKey,
		bankKeeper:      bankKeeper,
		validatorKeeper: validatorKeeper,
		authority:       authority,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the governance authority address
func (k Keeper) GetAuthority() string {
	return k.authority
}

// ============================================================================
// Params Management
// ============================================================================

// GetParams returns all module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return types.DefaultParams()
	}

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

// SetParams sets all module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}

// ============================================================================
// Supported Assets Management
// ============================================================================

// GetSupportedAsset retrieves a supported asset configuration
func (k Keeper) GetSupportedAsset(ctx sdk.Context, denom string) (types.SupportedAsset, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.SupportedAssetKey(denom))
	if bz == nil {
		return types.SupportedAsset{}, false
	}

	var asset types.SupportedAsset
	if err := json.Unmarshal(bz, &asset); err != nil {
		return types.SupportedAsset{}, false
	}
	return asset, true
}

// SetSupportedAsset sets a supported asset configuration
func (k Keeper) SetSupportedAsset(ctx sdk.Context, asset types.SupportedAsset) {
	if err := asset.Validate(); err != nil {
		k.Logger(ctx).Error("invalid supported asset", "error", err)
		return
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(asset)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal supported asset", "error", err)
		return
	}
	store.Set(types.SupportedAssetKey(asset.Denom), bz)
}

// IsAssetSupported checks if an asset is supported for bridging
func (k Keeper) IsAssetSupported(ctx sdk.Context, denom string) bool {
	asset, found := k.GetSupportedAsset(ctx, denom)
	return found && asset.Enabled
}

// GetAllSupportedAssets returns all supported assets
func (k Keeper) GetAllSupportedAssets(ctx sdk.Context) []types.SupportedAsset {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SupportedAssetPrefix)
	defer iterator.Close()

	assets := []types.SupportedAsset{}
	for ; iterator.Valid(); iterator.Next() {
		var asset types.SupportedAsset
		if err := json.Unmarshal(iterator.Value(), &asset); err != nil {
			continue
		}
		assets = append(assets, asset)
	}
	return assets
}

// ============================================================================
// Reserve Management
// ============================================================================

// GetReserveBalance returns the balance of a specific asset in the reserve
func (k Keeper) GetReserveBalance(ctx sdk.Context, denom string) math.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BridgeReserveKey(denom))
	if bz == nil {
		return math.ZeroInt()
	}

	var reserve types.BridgeReserve
	if err := json.Unmarshal(bz, &reserve); err != nil {
		return math.ZeroInt()
	}
	return reserve.Balance
}

// SetReserveBalance sets the balance of a specific asset in the reserve
func (k Keeper) SetReserveBalance(ctx sdk.Context, denom string, balance math.Int) {
	if balance.IsNegative() {
		k.Logger(ctx).Error("attempted to set negative reserve balance", "denom", denom, "balance", balance)
		return
	}

	reserve := types.NewBridgeReserve(denom, balance)
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(reserve)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal reserve", "error", err)
		return
	}
	store.Set(types.BridgeReserveKey(denom), bz)
}

// AddToReserve adds assets to the reserve
func (k Keeper) AddToReserve(ctx sdk.Context, denom string, amount math.Int) {
	currentBalance := k.GetReserveBalance(ctx, denom)
	newBalance := currentBalance.Add(amount)
	k.SetReserveBalance(ctx, denom, newBalance)

	k.Logger(ctx).Info("added to reserve",
		"denom", denom,
		"amount", amount,
		"new_balance", newBalance,
	)
}

// SubtractFromReserve subtracts assets from the reserve
func (k Keeper) SubtractFromReserve(ctx sdk.Context, denom string, amount math.Int) error {
	currentBalance := k.GetReserveBalance(ctx, denom)
	if currentBalance.LT(amount) {
		return types.ErrInsufficientReserve
	}

	newBalance := currentBalance.Sub(amount)
	k.SetReserveBalance(ctx, denom, newBalance)

	k.Logger(ctx).Info("subtracted from reserve",
		"denom", denom,
		"amount", amount,
		"new_balance", newBalance,
	)

	return nil
}

// GetAllReserveBalances returns all reserve balances
func (k Keeper) GetAllReserveBalances(ctx sdk.Context) []types.BridgeReserve {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.BridgeReservePrefix)
	defer iterator.Close()

	reserves := []types.BridgeReserve{}
	for ; iterator.Valid(); iterator.Next() {
		var reserve types.BridgeReserve
		if err := json.Unmarshal(iterator.Value(), &reserve); err != nil {
			continue
		}
		reserves = append(reserves, reserve)
	}
	return reserves
}

// ============================================================================
// Swap Record Management
// ============================================================================

// GetNextSwapID returns the next swap ID
func (k Keeper) GetNextSwapID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextSwapIDKey)
	if bz == nil {
		return 1
	}
	return sdk.BigEndianToUint64(bz)
}

// SetNextSwapID sets the next swap ID
func (k Keeper) SetNextSwapID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextSwapIDKey, sdk.Uint64ToBigEndian(id))
}

// GetSwapRecord retrieves a swap record
func (k Keeper) GetSwapRecord(ctx sdk.Context, swapID uint64) (types.SwapRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.SwapRecordKey(swapID))
	if bz == nil {
		return types.SwapRecord{}, false
	}

	var record types.SwapRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.SwapRecord{}, false
	}
	return record, true
}

// SetSwapRecord stores a swap record
func (k Keeper) SetSwapRecord(ctx sdk.Context, record types.SwapRecord) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(record)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal swap record", "error", err)
		return
	}
	store.Set(types.SwapRecordKey(record.ID), bz)
}

// GetAllSwapRecords returns all swap records
func (k Keeper) GetAllSwapRecords(ctx sdk.Context) []types.SwapRecord {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.SwapRecordPrefix)
	defer iterator.Close()

	records := []types.SwapRecord{}
	for ; iterator.Valid(); iterator.Next() {
		var record types.SwapRecord
		if err := json.Unmarshal(iterator.Value(), &record); err != nil {
			continue
		}
		records = append(records, record)
	}
	return records
}

// ============================================================================
// Validator Checks
// ============================================================================

// IsValidator checks if an address is a validator
func (k Keeper) IsValidator(ctx sdk.Context, addr sdk.AccAddress) bool {
	if k.validatorKeeper == nil {
		return false
	}
	return k.validatorKeeper.IsValidatorActive(ctx, addr.String())
}

// GetValidatorCount returns the total number of active validators
func (k Keeper) GetValidatorCount(ctx sdk.Context) uint64 {
	if k.validatorKeeper == nil {
		return 0
	}

	count := uint64(0)
	k.validatorKeeper.IterateValidators(ctx, func(valAddr string) bool {
		count++
		return false
	})
	return count
}
