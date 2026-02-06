package feeabstraction

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/types"
)

// GenesisState defines the fee abstraction module's genesis state
type GenesisState struct {
	// Params are the module parameters
	Params types.Params `json:"params"`

	// TreasuryGrants are the existing treasury grants
	TreasuryGrants []types.TreasuryFeeGrant `json:"treasury_grants"`
}

// ProtoMessage implements the proto.Message interface
func (gs *GenesisState) ProtoMessage() {}

// Reset implements the proto.Message interface
func (gs *GenesisState) Reset() { *gs = GenesisState{} }

// String implements the proto.Message interface
func (gs GenesisState) String() string { return "GenesisState" }

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:         types.DefaultParams(),
		TreasuryGrants: []types.TreasuryFeeGrant{},
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate treasury grants
	grantees := make(map[string]bool)
	for _, grant := range gs.TreasuryGrants {
		if grant.Grantee == "" {
			return fmt.Errorf("grant has empty grantee")
		}
		if grantees[grant.Grantee] {
			return fmt.Errorf("duplicate grant for grantee: %s", grant.Grantee)
		}
		grantees[grant.Grantee] = true
	}

	return nil
}

// InitGenesis initializes the module's genesis state
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genState *GenesisState) {
	// Set params
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(fmt.Errorf("failed to set params: %w", err))
	}

	// Set treasury grants
	for _, grant := range genState.TreasuryGrants {
		if err := k.SetTreasuryGrant(ctx, grant); err != nil {
			panic(fmt.Errorf("failed to set treasury grant: %w", err))
		}
	}

	k.Logger(ctx).Info("fee abstraction genesis initialized",
		"enabled", genState.Params.Enabled,
		"markup_bp", genState.Params.FeeMarkupBasisPoints,
		"num_grants", len(genState.TreasuryGrants),
	)
}

// ExportGenesis exports the module's genesis state
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *GenesisState {
	return &GenesisState{
		Params:         k.GetParams(ctx),
		TreasuryGrants: k.GetAllTreasuryGrants(ctx),
	}
}
