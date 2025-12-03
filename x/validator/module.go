package validator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/v2/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sharehodl/sharehodl-blockchain/x/validator/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/validator/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface that defines the
// independent methods a Cosmos SDK module needs to implement.
type AppModuleBasic struct {
	cdc codec.BinaryCodec
}

func NewAppModuleBasic(cdc codec.BinaryCodec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// Name returns the name of the module as a string
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the amino codec for the module
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns a default GenesisState for the module
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis validates the GenesisState
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC Gateway routes here when we add query service
}

// GetTxCmd returns the root tx command for the module
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil // Will implement CLI commands later
}

// GetQueryCmd returns the root query command for the module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil // Will implement CLI commands later
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface that defines the inter-dependent methods that modules need to implement
type AppModule struct {
	AppModuleBasic

	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
}

func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
		stakingKeeper:  stakingKeeper,
	}
}

// RegisterServices registers a GRPC query service to respond to the module-specific GRPC queries
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	// Register query server when we implement it
}

// RegisterInvariants registers the invariants of the module
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs the module's genesis initialization
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)

	InitGenesis(ctx, am.keeper, genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the module's exported genesis state as raw JSON bytes
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion implements ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 2 }

// BeginBlock contains the logic that is automatically triggered at the beginning of each block
func (am AppModule) BeginBlock(ctx sdk.Context) (sdk.BeginBlock, error) {
	return sdk.BeginBlock{}, nil
}

// EndBlock contains the logic that is automatically triggered at the end of each block
func (am AppModule) EndBlock(ctx sdk.Context) (sdk.EndBlock, error) {
	return EndBlock(ctx, am.keeper)
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// EndBlock handles end block processing for validator reputation decay and verification timeouts
func EndBlock(ctx sdk.Context, k keeper.Keeper) (sdk.EndBlock, error) {
	// Handle verification timeouts
	handleVerificationTimeouts(ctx, k)
	
	// Update validator reputations (decay over time)
	updateValidatorReputations(ctx, k)
	
	return sdk.EndBlock{}, nil
}

// handleVerificationTimeouts checks for expired verifications and marks them as rejected
func handleVerificationTimeouts(ctx sdk.Context, k keeper.Keeper) {
	verifications := k.GetAllBusinessVerifications(ctx)
	
	for _, verification := range verifications {
		if verification.Status == types.VerificationPending || verification.Status == types.VerificationInProgress {
			if ctx.BlockTime().After(verification.VerificationDeadline) {
				verification.Status = types.VerificationRejected
				verification.DisputeReason = "Verification timeout - no response from assigned validators"
				k.SetBusinessVerification(ctx, verification)
				
				// Emit timeout event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						"verification_timeout",
						sdk.NewAttribute("business_id", verification.BusinessID),
						sdk.NewAttribute("company_name", verification.CompanyName),
					),
				)
			}
		}
	}
}

// updateValidatorReputations applies reputation decay over time
func updateValidatorReputations(ctx sdk.Context, k keeper.Keeper) {
	params := k.GetParams(ctx)
	tiers := k.GetAllValidatorTiers(ctx)
	
	for _, tierInfo := range tiers {
		// Apply monthly reputation decay if validator hasn't been active recently
		if !tierInfo.LastVerification.IsZero() {
			daysSinceLastVerification := int64(ctx.BlockTime().Sub(tierInfo.LastVerification).Hours() / 24)
			
			// Apply decay every 30 days of inactivity
			if daysSinceLastVerification > 30 {
				months := daysSinceLastVerification / 30
				for i := int64(0); i < months; i++ {
					// Reduce reputation by decay rate
					tierInfo.ReputationScore = tierInfo.ReputationScore.Mul(
						sdk.OneDec().Sub(params.ReputationDecayRate),
					)
				}
				
				// Save updated tier info
				k.SetValidatorTierInfo(ctx, tierInfo)
			}
		}
	}
}