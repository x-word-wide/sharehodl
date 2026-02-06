package governance

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/sharehodl/sharehodl-blockchain/x/governance/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic implements the AppModuleBasic interface for the governance module
type AppModuleBasic struct{}

// NewAppModuleBasic creates a new AppModuleBasic
func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{}
}

// Name returns the governance module's name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the governance module's types on the LegacyAmino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {}

// DefaultGenesis returns default genesis state as raw bytes for the governance module
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	gs := DefaultGenesisState()
	bz, _ := json.Marshal(gs)
	return bz
}

// ValidateGenesis performs genesis state validation for the governance module
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config interface{}, bz json.RawMessage) error {
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return err
	}
	return gs.Validate()
}

// AppModule implements the AppModule interface for the governance module
type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(),
		keeper:         k,
	}
}

// Name returns the governance module's name
func (am AppModule) Name() string {
	return types.ModuleName
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface
func (am AppModule) IsAppModule() {}

// BeginBlock executes all ABCI BeginBlock logic for the governance module
func (am AppModule) BeginBlock(ctx sdk.Context) error {
	return nil
}

// EndBlock executes all ABCI EndBlock logic for the governance module
func (am AppModule) EndBlock(ctx sdk.Context) error {
	// Process voting periods that have ended
	am.processEndedVotingPeriods(ctx)

	// Process execution queue
	am.keeper.ProcessExecutionQueue(ctx)

	return nil
}

// processEndedVotingPeriods processes all proposals whose voting period has ended
func (am AppModule) processEndedVotingPeriods(ctx sdk.Context) {
	// Use the keeper's IterateProposals to process proposals
	am.keeper.IterateProposals(ctx, func(proposal types.Proposal) bool {
		// Check if voting period has ended
		if proposal.Status == types.ProposalStatusVotingPeriod && ctx.BlockTime().After(proposal.VotingEndTime) {
			// Process the proposal results
			am.keeper.ProcessProposalResults(ctx, proposal.ID)
		}

		// Check deposit period
		if proposal.Status == types.ProposalStatusDepositPeriod && ctx.BlockTime().After(proposal.DepositEndTime) {
			// Check if minimum deposit was reached (use proposal's MinDeposit field)
			if proposal.TotalDeposit.LT(proposal.MinDeposit) {
				// Update proposal status to rejected
				proposal.Status = types.ProposalStatusRejected
				proposal.UpdatedAt = ctx.BlockTime()
				am.keeper.SetProposalExternal(ctx, proposal)
			}
		}

		return false // Continue iteration
	})
}

// GenesisState defines the governance module's genesis state
type GenesisState struct {
	Params    types.GovernanceParams `json:"params"`
	Proposals []types.Proposal       `json:"proposals"`
	Votes     []types.Vote           `json:"votes"`
	Deposits  []types.Deposit        `json:"deposits"`
}

// DefaultGenesisState returns the default genesis state for the governance module
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:    types.DefaultParams(),
		Proposals: []types.Proposal{},
		Votes:     []types.Vote{},
		Deposits:  []types.Deposit{},
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	return nil
}

// InitGenesis initializes the governance module's state from a provided genesis state
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState GenesisState
	if err := json.Unmarshal(data, &genesisState); err != nil {
		panic(err)
	}

	// Set params
	am.keeper.SetGovernanceParams(ctx, genesisState.Params)

	// Set proposals
	for _, proposal := range genesisState.Proposals {
		am.keeper.SetProposalExternal(ctx, proposal)
	}
}

// ExportGenesis returns the governance module's exported genesis state
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	proposals := am.keeper.GetAllProposals(ctx)
	gs := GenesisState{
		Params:    am.keeper.GetGovernanceParams(ctx),
		Proposals: proposals,
		Votes:     []types.Vote{},
		Deposits:  []types.Deposit{},
	}
	bz, _ := json.Marshal(gs)
	return bz
}

// ConsensusVersion returns the governance module's consensus version
func (am AppModule) ConsensusVersion() uint64 {
	return 1
}
