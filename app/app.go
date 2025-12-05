package app

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	dbm "github.com/cosmos/cosmos-db"
	abci "github.com/cometbft/cometbft/v2/abci/types"

	// ShareHODL modules
	hodlmodule "github.com/sharehodl/sharehodl-blockchain/x/hodl"
	hodlkeeper "github.com/sharehodl/sharehodl-blockchain/x/hodl/keeper"
	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
	
	validatormodule "github.com/sharehodl/sharehodl-blockchain/x/validator"
	validatorkeeper "github.com/sharehodl/sharehodl-blockchain/x/validator/keeper"
	validatortypes "github.com/sharehodl/sharehodl-blockchain/x/validator/types"
	
	equitymodule "github.com/sharehodl/sharehodl-blockchain/x/equity"
	equitykeeper "github.com/sharehodl/sharehodl-blockchain/x/equity/keeper"
	equitytypes "github.com/sharehodl/sharehodl-blockchain/x/equity/types"
	
	dexmodule "github.com/sharehodl/sharehodl-blockchain/x/dex"
	dexkeeper "github.com/sharehodl/sharehodl-blockchain/x/dex/keeper"
	dextypes "github.com/sharehodl/sharehodl-blockchain/x/dex/types"
)

const (
	Name = "sharehodl"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)
}

// ShareHODLApp extends ABCI appplication for ShareHODL blockchain
type ShareHODLApp struct {
	*baseapp.BaseApp

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper

	// ShareHODL keepers
	HODLKeeper            hodlkeeper.Keeper
	ValidatorKeeper       validatorkeeper.Keeper
	EquityKeeper          equitykeeper.Keeper
	DexKeeper             dexkeeper.Keeper

	// module manager
	MM               *module.Manager
	BasicManager     module.BasicManager
	configurator     module.Configurator
}

// NewShareHODLApp returns a reference to an initialized ShareHODLApp.
func NewShareHODLApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *ShareHODLApp {
	interfaceRegistry := types.NewInterfaceRegistry()
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := authtx.NewTxConfig(appCodec, authtx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	// basic manager - needs to be created early to register interfaces
	basicManager := module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(nil),
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		consensus.AppModuleBasic{},
		hodlmodule.NewAppModuleBasic(appCodec),
		validatormodule.NewAppModuleBasic(appCodec),
		equitymodule.NewAppModuleBasic(appCodec),
		dexmodule.NewAppModuleBasic(appCodec),
	)
	
	basicManager.RegisterInterfaces(interfaceRegistry)

	bApp := baseapp.NewBaseApp(Name, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		upgradetypes.StoreKey,
		consensusparamtypes.StoreKey,
		hodltypes.StoreKey,
		validatortypes.StoreKey,
		equitytypes.StoreKey,
		dextypes.StoreKey,
	)
	
	memKeys := storetypes.NewMemoryStoreKeys(hodltypes.MemStoreKey, validatortypes.MemStoreKey, equitytypes.MemStoreKey, dextypes.MemStoreKey)

	app := &ShareHODLApp{
		BaseApp:           bApp,
		cdc:               legacyAmino,
		appCodec:          appCodec,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
	}

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress("gov").String(),
		runtime.EventService{},
	)
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		address.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authtypes.NewModuleAddress("gov").String(),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		map[string]bool{},
		authtypes.NewModuleAddress("gov").String(),
		logger,
	)

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress("gov").String(),
		address.NewBech32Codec("cosmosvaloper"),
		address.NewBech32Codec("cosmosvalcons"),
	)

	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		map[int64]bool{},
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		DefaultNodeHome,
		app.BaseApp,
		authtypes.NewModuleAddress("gov").String(),
	)

	// Initialize HODL keeper
	app.HODLKeeper = *hodlkeeper.NewKeeper(
		appCodec,
		keys[hodltypes.StoreKey],
		memKeys[hodltypes.MemStoreKey],
		app.BankKeeper,
		app.AccountKeeper,
	)

	// Initialize Validator keeper
	app.ValidatorKeeper = *validatorkeeper.NewKeeper(
		appCodec,
		keys[validatortypes.StoreKey],
		memKeys[validatortypes.MemStoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
	)

	// Initialize Equity keeper
	app.EquityKeeper = *equitykeeper.NewKeeper(
		appCodec,
		keys[equitytypes.StoreKey],
		memKeys[equitytypes.MemStoreKey],
		app.BankKeeper,
		app.AccountKeeper,
		app.ValidatorKeeper,
	)

	// Initialize DEX keeper
	app.DexKeeper = *dexkeeper.NewKeeper(
		appCodec,
		keys[dextypes.StoreKey],
		memKeys[dextypes.MemStoreKey],
		app.BankKeeper,
		app.AccountKeeper,
		app.EquityKeeper,
		app.HODLKeeper,
	)

	/****  Module Manager ****/
	app.MM = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper,
			app.StakingKeeper,
			app,
			txConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil, nil),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, nil),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, nil),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		hodlmodule.NewAppModule(appCodec, app.HODLKeeper, app.AccountKeeper, app.BankKeeper),
		validatormodule.NewAppModule(appCodec, app.ValidatorKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		equitymodule.NewAppModule(appCodec, app.EquityKeeper, app.AccountKeeper, app.BankKeeper),
		dexmodule.NewAppModule(appCodec, app.DexKeeper, app.AccountKeeper, app.BankKeeper, app.EquityKeeper, app.HODLKeeper),
	)

	app.MM.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		stakingtypes.ModuleName,
		hodltypes.ModuleName,
		validatortypes.ModuleName,
		equitytypes.ModuleName,
		dextypes.ModuleName,
	)

	app.MM.SetOrderEndBlockers(
		stakingtypes.ModuleName,
		hodltypes.ModuleName,
		validatortypes.ModuleName,
		equitytypes.ModuleName,
		dextypes.ModuleName,
	)

	genesisModuleOrder := []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		upgradetypes.ModuleName,
		genutiltypes.ModuleName,
		consensusparamtypes.ModuleName,
		hodltypes.ModuleName,
		validatortypes.ModuleName,
		equitytypes.ModuleName,
		dextypes.ModuleName,
	}

	app.MM.SetOrderInitGenesis(genesisModuleOrder...)
	app.MM.SetOrderExportGenesis(genesisModuleOrder...)

	// initialize stores
	app.MountKVStores(keys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	// configure ante handler
	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: txConfig.SignModeHandler(),
		},
	)
	if err != nil {
		panic(err)
	}
	app.SetAnteHandler(anteHandler)

	// module configurator
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.MM.RegisterServices(app.configurator)

	// assign the basic manager that was created earlier
	app.BasicManager = basicManager

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(err)
		}
	}

	return app
}

// Name returns the name of the App
func (app *ShareHODLApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *ShareHODLApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.MM.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *ShareHODLApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.MM.EndBlock(ctx)
}

// InitChainer application update at chain initialization
func (app *ShareHODLApp) InitChainer(ctx sdk.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.MM.GetVersionMap())
	return app.MM.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *ShareHODLApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ExportAppStateAndValidators exports the state of the application for a genesis
// file.
func (app *ShareHODLApp) ExportAppStateAndValidators(
	forZeroHeight bool, jailAllowedAddrs, modulesToExport []string,
) (servertypes.ExportedApp, error) {
	ctx := app.NewContext(true)
	height := app.LastBlockHeight() + 1
	if forZeroHeight {
		height = 0
	}

	genState, err := app.MM.ExportGenesis(ctx, app.appCodec)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators, err := staking.WriteValidators(ctx, app.StakingKeeper)
	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          height,
		ConsensusParams: app.BaseApp.GetConsensusParams(ctx),
	}, err
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *ShareHODLApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	app.BasicManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// RegisterNodeService implements the Application.RegisterNodeService method.
func (app *ShareHODLApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *ShareHODLApp) RegisterTendermintService(clientCtx client.Context) {
	// This method is required by the Application interface but may be deprecated
	// For CometBFT v2, this might be handled differently
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *ShareHODLApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// GetTxConfig implements the TestingApp interface.
func (app *ShareHODLApp) GetTxConfig() client.TxConfig {
	return app.txConfig
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (app *ShareHODLApp) DefaultGenesis() map[string]json.RawMessage {
	return app.BasicManager.DefaultGenesis(app.appCodec)
}

// Configurator implements the TestingApp interface.
func (app *ShareHODLApp) Configurator() module.Configurator {
	return app.configurator
}

// GenesisState - The genesis state of the blockchain is represented here as a map of raw json
// messages keyed by a string module name.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	return GenesisState{}
}

// MakeEncodingConfig creates an EncodingConfig for sharehodl.
func MakeEncodingConfig() EncodingConfig {
	return MakeTestEncodingConfig()
}

// ModuleBasics defines the module BasicManager is in charge of setting up basic,
// non-dependant module elements, such as codec registration and genesis verification.
var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	genutil.NewAppModuleBasic(nil),
	bank.AppModuleBasic{},
	staking.AppModuleBasic{},
	upgrade.AppModuleBasic{},
	consensus.AppModuleBasic{},
	hodlmodule.NewAppModuleBasic(nil),
	validatormodule.NewAppModuleBasic(nil),
	equitymodule.NewAppModuleBasic(nil),
	dexmodule.NewAppModuleBasic(nil),
)