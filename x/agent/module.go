package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/sharehodl/sharehodl-blockchain/x/agent/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/agent/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

// AppModuleBasic implements the AppModuleBasic interface for the agent module
type AppModuleBasic struct {
	cdc codec.BinaryCodec
}

// NewAppModuleBasic creates a new AppModuleBasic
func NewAppModuleBasic(cdc codec.BinaryCodec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// Name returns the agent module's name
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the agent module's types for the given codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register message types if needed
}

// RegisterInterfaces registers the module's interface types
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	// Register interface types if needed
}

// DefaultGenesis returns the agent module's default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genesis := GenesisState{
		Agents:        []types.Agent{},
		Actions:       []types.AgentAction{},
		Subscriptions: []keeper.Subscription{},
		NextAgentId:   1,
		NextActionId:  1,
	}
	bz, _ := json.Marshal(genesis)
	return bz
}

// ValidateGenesis performs genesis state validation for the agent module
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data GenesisState
	if err := json.Unmarshal(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC gateway routes if needed
}

// ============================================================================
// AppModule
// ============================================================================

// AppModule implements the AppModule interface for the agent module
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule
func NewAppModule(cdc codec.BinaryCodec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface
func (am AppModule) IsAppModule() {}

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
}

// InitGenesis performs the agent module's genesis initialization
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState GenesisState
	if err := json.Unmarshal(data, &genesisState); err != nil {
		panic(fmt.Sprintf("failed to unmarshal agent genesis state: %v", err))
	}

	// Initialize agents
	for _, agent := range genesisState.Agents {
		am.keeper.SetAgent(ctx, agent)
		am.keeper.SetAgentByKey(ctx, agent.AgentKey, agent.ID)
	}

	// Initialize actions
	for _, action := range genesisState.Actions {
		am.keeper.SetAction(ctx, action)
	}

	// Initialize subscriptions
	for _, sub := range genesisState.Subscriptions {
		am.keeper.SetSubscription(ctx, sub)
	}

	// Set counters
	am.keeper.SetNextAgentID(ctx, genesisState.NextAgentId)
	am.keeper.SetNextActionID(ctx, genesisState.NextActionId)
}

// ExportGenesis returns the agent module's exported genesis state
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	var agents []types.Agent
	var actions []types.AgentAction
	var subscriptions []keeper.Subscription

	// Export agents
	am.keeper.IterateAgents(ctx, func(agent types.Agent) bool {
		agents = append(agents, agent)
		return false
	})

	// Export actions (could get large, may want to limit)
	// For now, we don't export historical actions

	// Export subscriptions
	// Would need an iterator for subscriptions

	genesis := GenesisState{
		Agents:        agents,
		Actions:       actions,
		Subscriptions: subscriptions,
		NextAgentId:   am.keeper.GetNextAgentID(ctx),
		NextActionId:  am.keeper.GetNextActionID(ctx),
	}

	bz, _ := json.Marshal(genesis)
	return bz
}

// ConsensusVersion implements AppModule/ConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock executes all ABCI BeginBlock logic respective to the agent module
func (am AppModule) BeginBlock(ctx context.Context) error {
	return nil
}

// EndBlock executes all ABCI EndBlock logic respective to the agent module
func (am AppModule) EndBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return am.endBlock(sdkCtx)
}

// endBlock handles end of block processing
func (am AppModule) endBlock(ctx sdk.Context) error {
	// Process any pending agent actions
	// Check for expired subscriptions
	// Auto-renew subscriptions if configured

	// Iterate through agents and check for expiry
	am.keeper.IterateAgents(ctx, func(agent types.Agent) bool {
		// Check if subscription expired
		if ctx.BlockTime().After(agent.ExpiresAt) && agent.IsActive {
			// Deactivate expired agents
			agent.IsActive = false
			am.keeper.SetAgent(ctx, agent)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"agent_expired",
					sdk.NewAttribute("agent_id", fmt.Sprintf("%d", agent.ID)),
					sdk.NewAttribute("owner", agent.Owner),
				),
			)
		}
		return false
	})

	return nil
}

// ============================================================================
// Genesis State
// ============================================================================

// GenesisState defines the agent module's genesis state
type GenesisState struct {
	Agents        []types.Agent        `json:"agents"`
	Actions       []types.AgentAction  `json:"actions"`
	Subscriptions []keeper.Subscription `json:"subscriptions"`
	NextAgentId   uint64               `json:"next_agent_id"`
	NextActionId  uint64               `json:"next_action_id"`
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	agentIDs := make(map[uint64]bool)
	agentKeys := make(map[string]bool)

	for _, agent := range gs.Agents {
		// Check for duplicate IDs
		if agentIDs[agent.ID] {
			return fmt.Errorf("duplicate agent ID: %d", agent.ID)
		}
		agentIDs[agent.ID] = true

		// Check for duplicate keys
		if agentKeys[agent.AgentKey] {
			return fmt.Errorf("duplicate agent key: %s", agent.AgentKey)
		}
		agentKeys[agent.AgentKey] = true

		// Validate agent
		if err := agent.Validate(); err != nil {
			return fmt.Errorf("invalid agent %d: %w", agent.ID, err)
		}
	}

	return nil
}

// RegisterMsgServer is a helper to register the message server
// This is used in module.go RegisterServices
func RegisterMsgServer(server interface{}, impl types.MsgServer) {
	// In a real implementation, this would register with gRPC
	// For now, we handle messages through the keeper
}
