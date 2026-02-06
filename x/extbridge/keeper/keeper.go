package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
)

// Keeper of the extbridge store
type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       storetypes.StoreKey
	memKey         storetypes.StoreKey
	bankKeeper     types.BankKeeper
	accountKeeper  types.AccountKeeper
	stakingKeeper  types.StakingKeeper
	escrowKeeper   types.EscrowKeeper // For ban checking
	authority      string             // Governance authority
}

// NewKeeper creates a new extbridge Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper,
	escrowKeeper types.EscrowKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		stakingKeeper: stakingKeeper,
		escrowKeeper:  escrowKeeper,
		authority:     authority,
	}
}

// SetEscrowKeeper sets the escrow keeper (for late binding to avoid circular deps)
func (k *Keeper) SetEscrowKeeper(escrowKeeper types.EscrowKeeper) {
	k.escrowKeeper = escrowKeeper
}

// IsAddressBanned checks if an address is banned from using the bridge
func (k Keeper) IsAddressBanned(ctx sdk.Context, address string) (bool, string) {
	if k.escrowKeeper == nil {
		// If escrow keeper not set, allow operation (graceful degradation)
		return false, ""
	}
	return k.escrowKeeper.IsAddressBanned(ctx, address, ctx.BlockTime())
}

// GetAuthority returns the governance authority address
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams returns the module parameters
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

// SetParams sets the module parameters
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

// =========================================================================
// External Chain Management
// =========================================================================

// GetExternalChain retrieves an external chain configuration
func (k Keeper) GetExternalChain(ctx sdk.Context, chainID string) (types.ExternalChain, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ExternalChainKey(chainID))
	if bz == nil {
		return types.ExternalChain{}, false
	}

	var chain types.ExternalChain
	if err := json.Unmarshal(bz, &chain); err != nil {
		return types.ExternalChain{}, false
	}
	return chain, true
}

// SetExternalChain sets an external chain configuration
func (k Keeper) SetExternalChain(ctx sdk.Context, chain types.ExternalChain) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(chain)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal external chain", "error", err)
		return
	}
	store.Set(types.ExternalChainKey(chain.ChainID), bz)
}

// GetAllExternalChains returns all external chain configurations
func (k Keeper) GetAllExternalChains(ctx sdk.Context) []types.ExternalChain {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ExternalChainPrefix)
	defer iterator.Close()

	chains := []types.ExternalChain{}
	for ; iterator.Valid(); iterator.Next() {
		var chain types.ExternalChain
		if err := json.Unmarshal(iterator.Value(), &chain); err != nil {
			continue
		}
		chains = append(chains, chain)
	}
	return chains
}

// =========================================================================
// External Asset Management
// =========================================================================

// GetExternalAsset retrieves an external asset configuration
func (k Keeper) GetExternalAsset(ctx sdk.Context, chainID string, assetSymbol string) (types.ExternalAsset, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ExternalAssetKey(chainID, assetSymbol))
	if bz == nil {
		return types.ExternalAsset{}, false
	}

	var asset types.ExternalAsset
	if err := json.Unmarshal(bz, &asset); err != nil {
		return types.ExternalAsset{}, false
	}
	return asset, true
}

// SetExternalAsset sets an external asset configuration
func (k Keeper) SetExternalAsset(ctx sdk.Context, asset types.ExternalAsset) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(asset)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal external asset", "error", err)
		return
	}
	store.Set(types.ExternalAssetKey(asset.ChainID, asset.AssetSymbol), bz)
}

// GetAllExternalAssets returns all external asset configurations
func (k Keeper) GetAllExternalAssets(ctx sdk.Context) []types.ExternalAsset {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.ExternalAssetPrefix)
	defer iterator.Close()

	assets := []types.ExternalAsset{}
	for ; iterator.Valid(); iterator.Next() {
		var asset types.ExternalAsset
		if err := json.Unmarshal(iterator.Value(), &asset); err != nil {
			continue
		}
		assets = append(assets, asset)
	}
	return assets
}

// =========================================================================
// ID Management
// =========================================================================

// GetNextDepositID returns and increments the next deposit ID
func (k Keeper) GetNextDepositID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextDepositIDKey)
	if bz == nil {
		// First time: return 1 and store 2 for next call
		nextBz := make([]byte, 8)
		binary.BigEndian.PutUint64(nextBz, 2)
		store.Set(types.NextDepositIDKey, nextBz)
		return 1
	}

	id := binary.BigEndian.Uint64(bz)
	nextID := id + 1
	nextBz := make([]byte, 8)
	binary.BigEndian.PutUint64(nextBz, nextID)
	store.Set(types.NextDepositIDKey, nextBz)

	return id
}

// GetNextWithdrawalID returns and increments the next withdrawal ID
func (k Keeper) GetNextWithdrawalID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextWithdrawalIDKey)
	if bz == nil {
		// First time: return 1 and store 2 for next call
		nextBz := make([]byte, 8)
		binary.BigEndian.PutUint64(nextBz, 2)
		store.Set(types.NextWithdrawalIDKey, nextBz)
		return 1
	}

	id := binary.BigEndian.Uint64(bz)
	nextID := id + 1
	nextBz := make([]byte, 8)
	binary.BigEndian.PutUint64(nextBz, nextID)
	store.Set(types.NextWithdrawalIDKey, nextBz)

	return id
}

// GetNextTSSSessionID returns and increments the next TSS session ID
func (k Keeper) GetNextTSSSessionID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextTSSSessionIDKey)
	if bz == nil {
		// First time: return 1 and store 2 for next call
		nextBz := make([]byte, 8)
		binary.BigEndian.PutUint64(nextBz, 2)
		store.Set(types.NextTSSSessionIDKey, nextBz)
		return 1
	}

	id := binary.BigEndian.Uint64(bz)
	nextID := id + 1
	nextBz := make([]byte, 8)
	binary.BigEndian.PutUint64(nextBz, nextID)
	store.Set(types.NextTSSSessionIDKey, nextBz)

	return id
}

// =========================================================================
// Validator Tier Checks
// =========================================================================

// IsValidatorEligible checks if a validator is eligible to participate
func (k Keeper) IsValidatorEligible(ctx sdk.Context, validator sdk.AccAddress) (bool, error) {
	params := k.GetParams(ctx)

	// Check if validator
	if !k.stakingKeeper.IsValidator(ctx, validator) {
		return false, types.ErrNotValidator
	}

	// Check tier
	tier, err := k.stakingKeeper.GetValidatorTier(ctx, validator)
	if err != nil {
		return false, err
	}

	if tier < params.MinValidatorTier {
		return false, types.ErrInsufficientTier
	}

	return true, nil
}

// GetEligibleValidators returns all validators eligible to participate
func (k Keeper) GetEligibleValidators(ctx sdk.Context) ([]sdk.AccAddress, error) {
	params := k.GetParams(ctx)
	return k.stakingKeeper.GetValidatorsByTier(ctx, params.MinValidatorTier)
}

// CalculateRequiredAttestations calculates how many attestations are required
func (k Keeper) CalculateRequiredAttestations(ctx sdk.Context) (uint64, error) {
	params := k.GetParams(ctx)

	// Use eligible validators (with sufficient tier) instead of total validators
	eligibleValidators, err := k.GetEligibleValidators(ctx)
	if err != nil {
		return 0, err
	}

	totalEligible := uint64(len(eligibleValidators))
	if totalEligible == 0 {
		return 0, fmt.Errorf("no eligible validators available")
	}

	// Calculate threshold (e.g., 2/3 of eligible validators)
	threshold := params.AttestationThreshold
	required := math.LegacyNewDec(int64(totalEligible)).Mul(threshold).Ceil().TruncateInt().Uint64()

	// Ensure at least 1 attestation
	if required == 0 {
		required = 1
	}

	return required, nil
}

// =========================================================================
// Circuit Breaker
// =========================================================================

// GetCircuitBreaker retrieves the circuit breaker state
func (k Keeper) GetCircuitBreaker(ctx sdk.Context) types.CircuitBreaker {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CircuitBreakerKey)
	if bz == nil {
		return types.CircuitBreaker{Enabled: false}
	}

	var cb types.CircuitBreaker
	if err := json.Unmarshal(bz, &cb); err != nil {
		return types.CircuitBreaker{Enabled: false}
	}
	return cb
}

// SetCircuitBreaker sets the circuit breaker state
func (k Keeper) SetCircuitBreaker(ctx sdk.Context, cb types.CircuitBreaker) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(cb)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal circuit breaker", "error", err)
		return
	}
	store.Set(types.CircuitBreakerKey, bz)
}

// IsOperationAllowed checks if an operation is allowed by circuit breaker
func (k Keeper) IsOperationAllowed(ctx sdk.Context, operation string) error {
	cb := k.GetCircuitBreaker(ctx)

	// Check if expired
	if cb.IsExpired(ctx.BlockTime()) {
		cb.Enabled = false
		k.SetCircuitBreaker(ctx, cb)
		return nil
	}

	if !cb.IsOperationAllowed(operation) {
		return types.ErrCircuitBreakerActive
	}

	return nil
}

// =========================================================================
// Asset Conversion
// =========================================================================

// ConvertToHODL converts external asset amount to HODL amount
func (k Keeper) ConvertToHODL(ctx sdk.Context, chainID string, assetSymbol string, amount math.Int) (math.Int, error) {
	asset, found := k.GetExternalAsset(ctx, chainID, assetSymbol)
	if !found {
		return math.Int{}, types.ErrAssetNotSupported
	}

	// HODL amount = external amount * conversion rate
	hodlAmount := math.LegacyNewDecFromInt(amount).Mul(asset.ConversionRate).TruncateInt()
	return hodlAmount, nil
}

// ConvertFromHODL converts HODL amount to external asset amount
func (k Keeper) ConvertFromHODL(ctx sdk.Context, chainID string, assetSymbol string, hodlAmount math.Int) (math.Int, error) {
	asset, found := k.GetExternalAsset(ctx, chainID, assetSymbol)
	if !found {
		return math.Int{}, types.ErrAssetNotSupported
	}

	// External amount = HODL amount / conversion rate
	externalAmount := math.LegacyNewDecFromInt(hodlAmount).Quo(asset.ConversionRate).TruncateInt()
	return externalAmount, nil
}
