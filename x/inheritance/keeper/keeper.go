package keeper

import (
	"encoding/binary"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

// Keeper of the inheritance store
type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      storetypes.StoreKey
	memKey        storetypes.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	equityKeeper  types.EquityKeeper
	banKeeper     types.BanKeeper
	hodlKeeper    types.HODLKeeper
	stakingKeeper types.StakingKeeper
	lendingKeeper types.LendingKeeper
	escrowKeeper  types.EscrowKeeper

	// Authority for governance-controlled operations
	authority string
}

// NewKeeper creates a new inheritance Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	equityKeeper types.EquityKeeper,
	banKeeper types.BanKeeper,
	hodlKeeper types.HODLKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		equityKeeper:  equityKeeper,
		banKeeper:     banKeeper,
		hodlKeeper:    hodlKeeper,
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

// GetStoreKey returns the store key
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// GetCodec returns the codec
func (k Keeper) GetCodec() codec.BinaryCodec {
	return k.cdc
}

// SetStakingKeeper sets the staking keeper (for late binding during app initialization)
func (k *Keeper) SetStakingKeeper(stakingKeeper types.StakingKeeper) {
	k.stakingKeeper = stakingKeeper
}

// SetLendingKeeper sets the lending keeper (for late binding during app initialization)
func (k *Keeper) SetLendingKeeper(lendingKeeper types.LendingKeeper) {
	k.lendingKeeper = lendingKeeper
}

// SetEscrowKeeper sets the escrow keeper (for late binding during app initialization)
func (k *Keeper) SetEscrowKeeper(escrowKeeper types.EscrowKeeper) {
	k.escrowKeeper = escrowKeeper
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

// GetNextPlanID returns and increments the plan counter
func (k Keeper) GetNextPlanID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PlanCounterKey)

	var planID uint64
	if bz == nil {
		planID = 1
	} else {
		planID = binary.BigEndian.Uint64(bz)
	}

	// Increment and store
	nextID := planID + 1
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, nextID)
	store.Set(types.PlanCounterKey, bz)

	return planID
}

// CreatePlan creates a new inheritance plan
func (k Keeper) CreatePlan(ctx sdk.Context, plan types.InheritancePlan) error {
	// Validate the plan
	if err := plan.Validate(); err != nil {
		return err
	}

	// Check if owner is banned - banned users cannot create plans
	if k.banKeeper != nil && k.banKeeper.IsAddressBanned(ctx, plan.Owner) {
		return types.ErrOwnerBanned
	}

	// Validate against params
	params := k.GetParams(ctx)
	if plan.InactivityPeriod < params.MinInactivityPeriod {
		return types.ErrInvalidInactivity
	}
	if plan.GracePeriod < params.MinGracePeriod {
		return types.ErrInvalidGracePeriod
	}
	if plan.ClaimWindow < params.MinClaimWindow || plan.ClaimWindow > params.MaxClaimWindow {
		return types.ErrInvalidClaimWindow
	}
	if uint32(len(plan.Beneficiaries)) > params.MaxBeneficiaries {
		return types.ErrTooManyBeneficiaries
	}

	// Store the plan
	k.SetPlan(ctx, plan)

	// Create indices
	owner, _ := sdk.AccAddressFromBech32(plan.Owner)
	k.SetOwnerPlanIndex(ctx, owner, plan.PlanID)

	for _, beneficiary := range plan.Beneficiaries {
		bAddr, _ := sdk.AccAddressFromBech32(beneficiary.Address)
		k.SetBeneficiaryPlanIndex(ctx, bAddr, plan.PlanID)
	}

	k.Logger(ctx).Info("inheritance plan created",
		"plan_id", plan.PlanID,
		"owner", plan.Owner,
		"beneficiaries", len(plan.Beneficiaries),
	)

	return nil
}

// GetPlan returns a plan by ID
func (k Keeper) GetPlan(ctx sdk.Context, planID uint64) (types.InheritancePlan, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.InheritancePlanKey(planID))
	if bz == nil {
		return types.InheritancePlan{}, false
	}

	var plan types.InheritancePlan
	k.cdc.MustUnmarshal(bz, &plan)
	return plan, true
}

// SetPlan stores a plan
func (k Keeper) SetPlan(ctx sdk.Context, plan types.InheritancePlan) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&plan)
	store.Set(types.InheritancePlanKey(plan.PlanID), bz)
}

// DeletePlan deletes a plan
func (k Keeper) DeletePlan(ctx sdk.Context, planID uint64) {
	store := ctx.KVStore(k.storeKey)
	plan, found := k.GetPlan(ctx, planID)
	if !found {
		return
	}

	// Delete the plan
	store.Delete(types.InheritancePlanKey(planID))

	// Delete indices
	owner, _ := sdk.AccAddressFromBech32(plan.Owner)
	k.DeleteOwnerPlanIndex(ctx, owner, planID)

	for _, beneficiary := range plan.Beneficiaries {
		bAddr, _ := sdk.AccAddressFromBech32(beneficiary.Address)
		k.DeleteBeneficiaryPlanIndex(ctx, bAddr, planID)
	}
}

// SetOwnerPlanIndex sets the owner->plan index
func (k Keeper) SetOwnerPlanIndex(ctx sdk.Context, owner sdk.AccAddress, planID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.OwnerPlanIndexKey(owner, planID), []byte{1})
}

// DeleteOwnerPlanIndex deletes the owner->plan index
func (k Keeper) DeleteOwnerPlanIndex(ctx sdk.Context, owner sdk.AccAddress, planID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.OwnerPlanIndexKey(owner, planID))
}

// SetBeneficiaryPlanIndex sets the beneficiary->plan index
func (k Keeper) SetBeneficiaryPlanIndex(ctx sdk.Context, beneficiary sdk.AccAddress, planID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.BeneficiaryPlanIndexKey(beneficiary, planID), []byte{1})
}

// DeleteBeneficiaryPlanIndex deletes the beneficiary->plan index
func (k Keeper) DeleteBeneficiaryPlanIndex(ctx sdk.Context, beneficiary sdk.AccAddress, planID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.BeneficiaryPlanIndexKey(beneficiary, planID))
}

// GetPlansByOwner returns all plans owned by an address
func (k Keeper) GetPlansByOwner(ctx sdk.Context, owner sdk.AccAddress) []types.InheritancePlan {
	store := ctx.KVStore(k.storeKey)
	prefix := types.OwnerPlanIndexPrefixForAddress(owner)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var plans []types.InheritancePlan
	for ; iterator.Valid(); iterator.Next() {
		// Extract plan ID from key (last 8 bytes)
		key := iterator.Key()
		planIDBz := key[len(key)-8:]
		planID := binary.BigEndian.Uint64(planIDBz)

		plan, found := k.GetPlan(ctx, planID)
		if found {
			plans = append(plans, plan)
		}
	}

	return plans
}

// GetPlansByBeneficiary returns all plans where address is a beneficiary
func (k Keeper) GetPlansByBeneficiary(ctx sdk.Context, beneficiary sdk.AccAddress) []types.InheritancePlan {
	store := ctx.KVStore(k.storeKey)
	prefix := types.BeneficiaryPlanIndexPrefixForAddress(beneficiary)
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	var plans []types.InheritancePlan
	for ; iterator.Valid(); iterator.Next() {
		// Extract plan ID from key (last 8 bytes)
		key := iterator.Key()
		planIDBz := key[len(key)-8:]
		planID := binary.BigEndian.Uint64(planIDBz)

		plan, found := k.GetPlan(ctx, planID)
		if found {
			plans = append(plans, plan)
		}
	}

	return plans
}

// IteratePlans iterates over all plans
func (k Keeper) IteratePlans(ctx sdk.Context, cb func(plan types.InheritancePlan) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.InheritancePlanPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var plan types.InheritancePlan
		k.cdc.MustUnmarshal(iterator.Value(), &plan)
		if cb(plan) {
			break
		}
	}
}

// GetAllPlans returns all plans
func (k Keeper) GetAllPlans(ctx sdk.Context) []types.InheritancePlan {
	var plans []types.InheritancePlan
	k.IteratePlans(ctx, func(plan types.InheritancePlan) bool {
		plans = append(plans, plan)
		return false
	})
	return plans
}
