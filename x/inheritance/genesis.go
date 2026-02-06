package inheritance

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

// GenesisState defines the inheritance module's genesis state
type GenesisState struct {
	Params          types.Params             `json:"params"`
	Plans           []types.InheritancePlan  `json:"plans"`
	Triggers        []types.SwitchTrigger    `json:"triggers"`
	Claims          []types.BeneficiaryClaim `json:"claims"`
	ActivityRecords []types.ActivityRecord   `json:"activity_records"`
}

// ProtoMessage implementations
func (m *GenesisState) ProtoMessage() {}
func (m *GenesisState) Reset()        { *m = GenesisState{} }
func (m *GenesisState) String() string { return "GenesisState{}" }

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:          types.DefaultParams(),
		Plans:           []types.InheritancePlan{},
		Triggers:        []types.SwitchTrigger{},
		Claims:          []types.BeneficiaryClaim{},
		ActivityRecords: []types.ActivityRecord{},
	}
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate plans
	planIDs := make(map[uint64]bool)
	for _, plan := range gs.Plans {
		if err := plan.Validate(); err != nil {
			return err
		}
		if planIDs[plan.PlanID] {
			return types.ErrInvalidPlan
		}
		planIDs[plan.PlanID] = true
	}

	// Validate triggers
	for _, trigger := range gs.Triggers {
		if err := trigger.Validate(); err != nil {
			return err
		}
		if !planIDs[trigger.PlanID] {
			return types.ErrPlanNotFound
		}
	}

	// Validate claims
	for _, claim := range gs.Claims {
		if err := claim.Validate(); err != nil {
			return err
		}
		if !planIDs[claim.PlanID] {
			return types.ErrPlanNotFound
		}
	}

	return nil
}

// InitGenesis initializes the inheritance module's state from a provided genesis state
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs GenesisState) {
	// Set params
	k.SetParams(ctx, gs.Params)

	// Initialize plans
	for _, plan := range gs.Plans {
		k.SetPlan(ctx, plan)

		// Create indices
		owner, _ := sdk.AccAddressFromBech32(plan.Owner)
		k.SetOwnerPlanIndex(ctx, owner, plan.PlanID)

		for _, beneficiary := range plan.Beneficiaries {
			bAddr, _ := sdk.AccAddressFromBech32(beneficiary.Address)
			k.SetBeneficiaryPlanIndex(ctx, bAddr, plan.PlanID)
		}
	}

	// Initialize triggers
	for _, trigger := range gs.Triggers {
		k.SetActiveTrigger(ctx, trigger)
	}

	// Initialize claims
	for _, claim := range gs.Claims {
		k.SetClaim(ctx, claim)
	}

	// Initialize activity records
	for _, activity := range gs.ActivityRecords {
		addr, _ := sdk.AccAddressFromBech32(activity.Address)
		store := ctx.KVStore(k.GetStoreKey())
		bz := k.GetCodec().MustMarshal(&activity)
		store.Set(types.LastActivityKey(addr), bz)
	}
}

// ExportGenesis returns the inheritance module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *GenesisState {
	return &GenesisState{
		Params:          k.GetParams(ctx),
		Plans:           k.GetAllPlans(ctx),
		Triggers:        k.GetAllActiveTriggers(ctx),
		Claims:          []types.BeneficiaryClaim{}, // Would need to iterate through all claims
		ActivityRecords: []types.ActivityRecord{},   // Would need to iterate through all activity records
	}
}
