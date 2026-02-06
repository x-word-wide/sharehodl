package equity

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/client/cli"
	"github.com/sharehodl/sharehodl-blockchain/x/equity/keeper"
	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.HasServices     = AppModule{}
	_ appmodule.AppModule    = AppModule{}
)

// AppModuleBasic defines the basic application module used by the equity module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// NewAppModuleBasic creates a new AppModuleBasic object
func NewAppModuleBasic(cdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

// Name returns the equity module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the equity module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register amino codec for types if needed
}

// RegisterInterfaces registers the equity module's interface types
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	// Register interfaces if needed
}

// DefaultGenesis returns the equity module's default genesis state.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	// Return default genesis state
	return json.RawMessage("{}")
}

// ValidateGenesis performs genesis state validation for the equity module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	// Validate genesis state
	return nil
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// Register gRPC gateway routes if needed
}

// GetTxCmd returns the equity module's root tx command.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns the equity module's root query command.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements the AppModule interface for the equity module.
type AppModule struct {
	AppModuleBasic

	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
		accountKeeper:  accountKeeper,
		bankKeeper:     bankKeeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the equity module's name.
func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

// RegisterServices registers a GRPC query service to respond to the module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Register message server
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	
	// Register query server when implemented
	// types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

// InitGenesis performs the equity module's genesis initialization. It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	// Initialize genesis state - simplified for now
}

// ExportGenesis returns the equity module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// Export genesis state - simplified for now
	return json.RawMessage("{}")
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock executes all ABCI BeginBlock logic respective to the equity module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	// BeginBlock logic - none needed for now
	return nil
}

// EndBlock executes all ABCI EndBlock logic respective to the equity module.
func (am AppModule) EndBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Process automatic dividend operations
	am.processDividendEndBlock(sdkCtx)

	// Process delisting-related operations
	am.processDelistingEndBlock(sdkCtx)

	// Process shareholder petition thresholds
	am.processPetitionEndBlock(sdkCtx)

	return nil
}

// processDelistingEndBlock handles delisting and compensation processing
func (am AppModule) processDelistingEndBlock(ctx sdk.Context) {
	// Process expired freeze warnings (execute freeze or escalate to Archon)
	am.keeper.ProcessExpiredFreezeWarnings(ctx)

	// Process expired trading halts (auto-resume)
	am.keeper.ProcessExpiredTradingHalts(ctx)

	// Process expired compensation claims (return remaining to community pool)
	am.keeper.ProcessExpiredCompensations(ctx)
}

// processPetitionEndBlock handles shareholder petition processing
func (am AppModule) processPetitionEndBlock(ctx sdk.Context) {
	// Process petition thresholds - convert to reports if threshold met
	am.keeper.ProcessPetitionThresholds(ctx)
}

// processDividendEndBlock handles automatic dividend processing
func (am AppModule) processDividendEndBlock(ctx sdk.Context) {
	// Process dividends that need record date snapshots
	am.processRecordDateSnapshots(ctx)

	// Process dividend payments in batches
	am.processDividendPayments(ctx)
}

// processRecordDateSnapshots creates snapshots for dividends that reached record date
func (am AppModule) processRecordDateSnapshots(ctx sdk.Context) {
	// Get all dividends in declared status
	dividends := am.keeper.GetAllDividendsByStatus(ctx, types.DividendStatusDeclared)

	for _, dividend := range dividends {
		if dividend.IsRecordDateReached(ctx.BlockTime()) {
			err := am.keeper.CreateRecordSnapshot(ctx, dividend.ID)
			if err != nil {
				ctx.Logger().Error("failed to create record snapshot",
					"dividend_id", dividend.ID,
					"error", err,
				)
			} else {
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						"dividend_snapshot_created",
						sdk.NewAttribute("dividend_id", fmt.Sprintf("%d", dividend.ID)),
						sdk.NewAttribute("company_id", fmt.Sprintf("%d", dividend.CompanyID)),
					),
				)
			}
		}
	}
}

// processDividendPayments processes dividend payments in batches
func (am AppModule) processDividendPayments(ctx sdk.Context) {
	// Get all dividends ready for processing
	dividends := am.keeper.GetAllDividendsByStatus(ctx, types.DividendStatusRecorded)

	// Process up to 10 dividends per block, 100 payments each
	const maxDividendsPerBlock = 10
	const batchSize = 100

	processed := 0
	for _, dividend := range dividends {
		if processed >= maxDividendsPerBlock {
			break
		}

		if dividend.IsReadyForPayment(ctx.BlockTime()) {
			paymentsProcessed, totalPaid, isComplete, err := am.keeper.ProcessDividendPayments(ctx, dividend.ID, uint32(batchSize))

			if err != nil {
				ctx.Logger().Error("failed to process dividend payments",
					"dividend_id", dividend.ID,
					"error", err,
				)
			} else {
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						"dividend_payments_processed",
						sdk.NewAttribute("dividend_id", fmt.Sprintf("%d", dividend.ID)),
						sdk.NewAttribute("payments_processed", fmt.Sprintf("%d", paymentsProcessed)),
						sdk.NewAttribute("total_paid", totalPaid.String()),
						sdk.NewAttribute("is_complete", fmt.Sprintf("%t", isComplete)),
					),
				)
			}

			processed++
		}
	}
}