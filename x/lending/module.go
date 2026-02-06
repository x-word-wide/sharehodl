package lending

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/sharehodl/sharehodl-blockchain/x/lending/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/lending/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic implements the AppModuleBasic interface for the lending module
type AppModuleBasic struct{}

// Name returns the lending module's name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the lending module's types on the LegacyAmino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {}

// DefaultGenesis returns default genesis state as raw bytes for the lending module
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the lending module
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config interface{}, bz json.RawMessage) error {
	return nil
}

// AppModule implements the AppModule interface for the lending module
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

// Name returns the lending module's name
func (am AppModule) Name() string {
	return types.ModuleName
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface
func (am AppModule) IsAppModule() {}

// BeginBlock executes all ABCI BeginBlock logic for the lending module
func (am AppModule) BeginBlock(ctx sdk.Context) error {
	return nil
}

// EndBlock executes all ABCI EndBlock logic for the lending module
func (am AppModule) EndBlock(ctx sdk.Context) error {
	// Process loans (interest accrual, status updates)
	am.keeper.ProcessLoans(ctx)
	return nil
}

// GenesisState represents the lending module's genesis state
type GenesisState struct {
	Loans        []types.Loan        `json:"loans"`
	LendingPools []types.LendingPool `json:"lending_pools"`
}

// ProtoMessage implements proto.Message
func (gs *GenesisState) ProtoMessage() {}

// Reset implements proto.Message
func (gs *GenesisState) Reset() { *gs = GenesisState{} }

// String implements proto.Message
func (gs *GenesisState) String() string { return "lending_genesis" }

// DefaultGenesisState returns the default genesis state for the lending module
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Loans:        []types.Loan{},
		LendingPools: []types.LendingPool{},
	}
}

// InitGenesis initializes the lending module's state from a provided genesis state
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	for _, loan := range genesisState.Loans {
		am.keeper.SetLoan(ctx, loan)
	}

	for _, pool := range genesisState.LendingPools {
		am.keeper.SetLendingPool(ctx, pool)
	}
}

// ExportGenesis returns the lending module's exported genesis state
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := GenesisState{
		Loans:        am.keeper.GetAllLoans(ctx),
		LendingPools: []types.LendingPool{},
	}
	return cdc.MustMarshalJSON(&gs)
}

// ConsensusVersion returns the lending module's consensus version
func (am AppModule) ConsensusVersion() uint64 {
	return 1
}
