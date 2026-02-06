package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/types"
)

// Keeper manages the fee abstraction module state
type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      storetypes.StoreKey
	authority     string // governance module address
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	dexKeeper     types.DEXKeeper
	equityKeeper  types.EquityKeeper
}

// NewKeeper creates a new fee abstraction keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	authority string,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	dexKeeper types.DEXKeeper,
	equityKeeper types.EquityKeeper,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		authority:     authority,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		dexKeeper:     dexKeeper,
		equityKeeper:  equityKeeper,
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

// =============================================================================
// PARAMS MANAGEMENT - All values governance-controllable
// =============================================================================

// GetParams returns the current module parameters
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

// Param getters for convenience

// IsEnabled returns whether fee abstraction is enabled
func (k Keeper) IsEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).Enabled
}

// GetFeeMarkupBasisPoints returns the fee markup in basis points
func (k Keeper) GetFeeMarkupBasisPoints(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).FeeMarkupBasisPoints
}

// GetMaxSlippageBasisPoints returns the max slippage in basis points
func (k Keeper) GetMaxSlippageBasisPoints(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).MaxSlippageBasisPoints
}

// GetMinEquityValue returns the minimum equity value for fee abstraction
func (k Keeper) GetMinEquityValue(ctx sdk.Context) math.Int {
	params := k.GetParams(ctx)
	if params.MinEquityValueUhodl.IsNil() {
		return types.DefaultMinEquityValueUhodl
	}
	return params.MinEquityValueUhodl
}

// GetMaxFeeAbstractionPerBlock returns the DoS protection limit per block
func (k Keeper) GetMaxFeeAbstractionPerBlock(ctx sdk.Context) math.Int {
	params := k.GetParams(ctx)
	if params.MaxFeeAbstractionPerBlockUhodl.IsNil() {
		return types.DefaultMaxFeeAbstractionPerBlockUhodl
	}
	return params.MaxFeeAbstractionPerBlockUhodl
}

// GetTreasuryFundingMinimum returns the minimum treasury funding required
func (k Keeper) GetTreasuryFundingMinimum(ctx sdk.Context) math.Int {
	params := k.GetParams(ctx)
	if params.TreasuryFundingMinimum.IsNil() {
		return types.DefaultTreasuryFundingMinimum
	}
	return params.TreasuryFundingMinimum
}

// GetEnabledEquities returns the whitelist of enabled equities (empty = all)
func (k Keeper) GetEnabledEquities(ctx sdk.Context) []string {
	return k.GetParams(ctx).EnabledEquitySymbols
}

// IsEquityEnabled checks if a specific equity is enabled for fee abstraction
func (k Keeper) IsEquityEnabled(ctx sdk.Context, symbol string) bool {
	enabled := k.GetEnabledEquities(ctx)
	if len(enabled) == 0 {
		return true // Empty list means all equities are enabled
	}
	for _, s := range enabled {
		if s == symbol {
			return true
		}
	}
	return false
}

// =============================================================================
// BLOCK USAGE TRACKING (DoS Protection)
// =============================================================================

// GetBlockUsage returns the total fee abstraction usage for the current block
func (k Keeper) GetBlockUsage(ctx sdk.Context) math.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.BlockUsageKey)
	if bz == nil {
		return math.ZeroInt()
	}

	var usage struct {
		Height int64    `json:"height"`
		Amount math.Int `json:"amount"`
	}
	if err := json.Unmarshal(bz, &usage); err != nil {
		return math.ZeroInt()
	}

	// Reset if new block
	if usage.Height != ctx.BlockHeight() {
		return math.ZeroInt()
	}
	return usage.Amount
}

// IncrementBlockUsage adds to the block usage counter
func (k Keeper) IncrementBlockUsage(ctx sdk.Context, amount math.Int) {
	currentUsage := k.GetBlockUsage(ctx)
	usage := struct {
		Height int64    `json:"height"`
		Amount math.Int `json:"amount"`
	}{
		Height: ctx.BlockHeight(),
		Amount: currentUsage.Add(amount),
	}

	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(usage)
	store.Set(types.BlockUsageKey, bz)
}

// CheckBlockLimit checks if the block limit would be exceeded
func (k Keeper) CheckBlockLimit(ctx sdk.Context, amount math.Int) error {
	currentUsage := k.GetBlockUsage(ctx)
	maxUsage := k.GetMaxFeeAbstractionPerBlock(ctx)

	if currentUsage.Add(amount).GT(maxUsage) {
		return types.ErrBlockLimitExceeded
	}
	return nil
}

// =============================================================================
// RECORD TRACKING (Audit Trail)
// =============================================================================

// GetNextRecordID returns the next record ID and increments the counter
func (k Keeper) GetNextRecordID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.RecordCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz) + 1
	}

	store.Set(types.RecordCounterKey, sdk.Uint64ToBigEndian(counter))
	return counter
}

// SaveRecord saves a fee abstraction record
func (k Keeper) SaveRecord(ctx sdk.Context, record types.FeeAbstractionRecord) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := json.Marshal(record)
	store.Set(types.FeeAbstractionRecordKey(record.ID), bz)
}

// GetRecord retrieves a fee abstraction record by ID
func (k Keeper) GetRecord(ctx sdk.Context, id uint64) (types.FeeAbstractionRecord, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.FeeAbstractionRecordKey(id))
	if bz == nil {
		return types.FeeAbstractionRecord{}, false
	}

	var record types.FeeAbstractionRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return types.FeeAbstractionRecord{}, false
	}
	return record, true
}
