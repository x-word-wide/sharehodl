package escrow

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/sharehodl/sharehodl-blockchain/x/escrow/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/escrow/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic implements the AppModuleBasic interface for the escrow module
type AppModuleBasic struct{}

// Name returns the escrow module's name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the escrow module's types on the LegacyAmino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {}

// DefaultGenesis returns default genesis state as raw bytes for the escrow module
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the escrow module
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config interface{}, bz json.RawMessage) error {
	return nil
}

// AppModule implements the AppModule interface for the escrow module
type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(k *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

// Name returns the escrow module's name
func (am AppModule) Name() string {
	return types.ModuleName
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface
func (am AppModule) IsAppModule() {}

// BeginBlock executes all ABCI BeginBlock logic for the escrow module
func (am AppModule) BeginBlock(ctx sdk.Context) error {
	return nil
}

// EndBlock executes all ABCI EndBlock logic for the escrow module
func (am AppModule) EndBlock(ctx sdk.Context) error {
	// Process expired escrows
	am.keeper.ProcessExpiredEscrows(ctx)

	// Process completed unbondings (Stake-as-Trust-Ceiling)
	am.keeper.ProcessCompletedUnbondings(ctx)

	// Process expired temporary bans (auto-lift when ban duration ends)
	am.keeper.ProcessExpiredBans(ctx)

	// Phase 1: User Reporting System processing
	// Check moderator metrics for auto-blacklist conditions
	am.keeper.ProcessModeratorMetrics(ctx)

	// Process expired voluntary returns (escalate to investigation if Bob didn't return)
	am.keeper.ProcessExpiredVoluntaryReturns(ctx)

	// Process expired reports (auto-dismiss if deadline passed)
	am.keeper.ProcessExpiredReports(ctx)

	// Process expired appeals (auto-resolve based on current votes)
	am.keeper.ProcessExpiredAppeals(ctx)

	// Process expired reporter bans (auto-lift temporary bans)
	am.keeper.ProcessExpiredReporterBans(ctx)

	return nil
}

// GenesisState represents the escrow module's genesis state
type GenesisState struct {
	Escrows               []types.Escrow              `json:"escrows"`
	Disputes              []types.Dispute             `json:"disputes"`
	Moderators            []types.Moderator           `json:"moderators"`
	UnbondingModerators   []types.UnbondingModerator  `json:"unbonding_moderators"`
	ModeratorBlacklists   []types.ModeratorBlacklist  `json:"moderator_blacklists"`
	ValidatorActions      []types.ValidatorAction     `json:"validator_actions"`

	// Phase 1: User Reporting System
	Reports               []types.Report              `json:"reports"`
	Appeals               []types.Appeal              `json:"appeals"`
	ModeratorMetrics      []types.ModeratorMetrics    `json:"moderator_metrics"`
	EscrowReserve         types.EscrowReserve         `json:"escrow_reserve"`
	ReporterHistories     []types.ReporterHistory     `json:"reporter_histories"`
}

// ProtoMessage implements proto.Message
func (gs *GenesisState) ProtoMessage() {}

// Reset implements proto.Message
func (gs *GenesisState) Reset() { *gs = GenesisState{} }

// String implements proto.Message
func (gs *GenesisState) String() string { return "escrow_genesis" }

// DefaultGenesisState returns the default genesis state for the escrow module
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Escrows:             []types.Escrow{},
		Disputes:            []types.Dispute{},
		Moderators:          []types.Moderator{},
		UnbondingModerators: []types.UnbondingModerator{},
		ModeratorBlacklists: []types.ModeratorBlacklist{},
		ValidatorActions:    []types.ValidatorAction{},

		// Phase 1: User Reporting System
		Reports:             []types.Report{},
		Appeals:             []types.Appeal{},
		ModeratorMetrics:    []types.ModeratorMetrics{},
		EscrowReserve:       types.NewEscrowReserve(),
		ReporterHistories:   []types.ReporterHistory{},
	}
}

// InitGenesis initializes the escrow module's state from a provided genesis state
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	for _, escrow := range genesisState.Escrows {
		am.keeper.SetEscrow(ctx, escrow)
	}

	for _, dispute := range genesisState.Disputes {
		am.keeper.SetDispute(ctx, dispute)
	}

	for _, moderator := range genesisState.Moderators {
		am.keeper.SetModerator(ctx, moderator)
	}

	for _, unbonding := range genesisState.UnbondingModerators {
		am.keeper.SetUnbondingModerator(ctx, unbonding)
	}

	for _, blacklist := range genesisState.ModeratorBlacklists {
		am.keeper.SetModeratorBlacklist(ctx, blacklist)
	}

	for _, action := range genesisState.ValidatorActions {
		am.keeper.SetValidatorAction(ctx, action)
	}

	// Phase 1: User Reporting System
	for _, report := range genesisState.Reports {
		am.keeper.SetReport(ctx, report)
		am.keeper.IndexReport(ctx, report)
	}

	for _, appeal := range genesisState.Appeals {
		am.keeper.SetAppeal(ctx, appeal)
		am.keeper.IndexAppeal(ctx, appeal)
	}

	for _, metrics := range genesisState.ModeratorMetrics {
		am.keeper.SetModeratorMetrics(ctx, metrics)
	}

	for _, history := range genesisState.ReporterHistories {
		am.keeper.SetReporterHistory(ctx, history)
	}

	am.keeper.SetEscrowReserve(ctx, genesisState.EscrowReserve)
}

// ExportGenesis returns the escrow module's exported genesis state
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := GenesisState{
		Escrows:             am.keeper.GetAllEscrows(ctx),
		Disputes:            am.keeper.GetAllDisputes(ctx),
		Moderators:          am.keeper.GetAllModerators(ctx),
		UnbondingModerators: am.keeper.GetAllUnbondingModerators(ctx),
		ModeratorBlacklists: am.keeper.GetAllBlacklistedModerators(ctx),
		ValidatorActions:    am.keeper.GetAllValidatorActions(ctx),

		// Phase 1: User Reporting System
		Reports:             am.keeper.GetAllReports(ctx),
		Appeals:             am.keeper.GetAllAppeals(ctx),
		ModeratorMetrics:    am.keeper.GetAllModeratorMetrics(ctx),
		EscrowReserve:       am.keeper.GetEscrowReserve(ctx),
		ReporterHistories:   am.keeper.GetAllReporterHistories(ctx),
	}
	return cdc.MustMarshalJSON(&gs)
}

// ConsensusVersion returns the escrow module's consensus version
func (am AppModule) ConsensusVersion() uint64 {
	return 1
}
