package app

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
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
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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
	"github.com/cosmos/gogoproto/proto"
	txsigning "cosmossdk.io/x/tx/signing"

	// ShareHODL modules
	hodlmodule "github.com/sharehodl/sharehodl-blockchain/x/hodl"
	hodlkeeper "github.com/sharehodl/sharehodl-blockchain/x/hodl/keeper"
	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"

	equitymodule "github.com/sharehodl/sharehodl-blockchain/x/equity"
	equitykeeper "github.com/sharehodl/sharehodl-blockchain/x/equity/keeper"
	equitytypes "github.com/sharehodl/sharehodl-blockchain/x/equity/types"

	dexmodule "github.com/sharehodl/sharehodl-blockchain/x/dex"
	dexkeeper "github.com/sharehodl/sharehodl-blockchain/x/dex/keeper"
	dextypes "github.com/sharehodl/sharehodl-blockchain/x/dex/types"

	escrowmodule "github.com/sharehodl/sharehodl-blockchain/x/escrow"
	escrowkeeper "github.com/sharehodl/sharehodl-blockchain/x/escrow/keeper"
	escrowtypes "github.com/sharehodl/sharehodl-blockchain/x/escrow/types"

	lendingmodule "github.com/sharehodl/sharehodl-blockchain/x/lending"
	lendingkeeper "github.com/sharehodl/sharehodl-blockchain/x/lending/keeper"
	lendingtypes "github.com/sharehodl/sharehodl-blockchain/x/lending/types"

	governancemodule "github.com/sharehodl/sharehodl-blockchain/x/governance"
	governancekeeper "github.com/sharehodl/sharehodl-blockchain/x/governance/keeper"
	governancetypes "github.com/sharehodl/sharehodl-blockchain/x/governance/types"

	agentmodule "github.com/sharehodl/sharehodl-blockchain/x/agent"
	agentkeeper "github.com/sharehodl/sharehodl-blockchain/x/agent/keeper"
	agenttypes "github.com/sharehodl/sharehodl-blockchain/x/agent/types"

	feeabstractionmodule "github.com/sharehodl/sharehodl-blockchain/x/feeabstraction"
	feeabstractionante "github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/ante"
	feeabstractionkeeper "github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/keeper"
	feeabstractiontypes "github.com/sharehodl/sharehodl-blockchain/x/feeabstraction/types"

	universalstakingmodule "github.com/sharehodl/sharehodl-blockchain/x/staking"
	universalstakingkeeper "github.com/sharehodl/sharehodl-blockchain/x/staking/keeper"
	universalstakingtypes "github.com/sharehodl/sharehodl-blockchain/x/staking/types"

	extbridgemodule "github.com/sharehodl/sharehodl-blockchain/x/extbridge"
	extbridgekeeper "github.com/sharehodl/sharehodl-blockchain/x/extbridge/keeper"
	extbridgetypes "github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"

	inheritancemodule "github.com/sharehodl/sharehodl-blockchain/x/inheritance"
	inheritancekeeper "github.com/sharehodl/sharehodl-blockchain/x/inheritance/keeper"
	inheritancetypes "github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"

	explorermodule "github.com/sharehodl/sharehodl-blockchain/x/explorer"
	explorerkeeper "github.com/sharehodl/sharehodl-blockchain/x/explorer/keeper"
	explorertypes "github.com/sharehodl/sharehodl-blockchain/x/explorer/types"

	validatormodule "github.com/sharehodl/sharehodl-blockchain/x/validator"
	validatorkeeper "github.com/sharehodl/sharehodl-blockchain/x/validator/keeper"
	validatortypes "github.com/sharehodl/sharehodl-blockchain/x/validator/types"

	bridgemodule "github.com/sharehodl/sharehodl-blockchain/x/bridge"
	bridgekeeper "github.com/sharehodl/sharehodl-blockchain/x/bridge/keeper"
	bridgetypes "github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
)

const (
	Name = "sharehodl"

	// ShareHODL bech32 address prefixes
	Bech32PrefixAccAddr  = "hodl"
	Bech32PrefixAccPub   = "hodlpub"
	Bech32PrefixValAddr  = "hodlvaloper"
	Bech32PrefixValPub   = "hodlvaloperpub"
	Bech32PrefixConsAddr = "hodlvalcons"
	Bech32PrefixConsPub  = "hodlvalconspub"

	// ShareHODL bech32 hash prefixes (for tx and block hashes)
	Bech32PrefixTxHash    = "sharetx"
	Bech32PrefixBlockHash = "shareblock"
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:              nil,
		stakingtypes.BondedPoolName:             {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:          {authtypes.Burner, authtypes.Staking},
		escrowtypes.ModuleName:                  nil,
		lendingtypes.ModuleName:                 nil,
		governancetypes.ModuleName:              {authtypes.Burner},
		agenttypes.ModuleName:                   nil,
		feeabstractiontypes.ModuleName:          nil,
		feeabstractiontypes.TreasuryPoolName:    nil,
		universalstakingtypes.StakingPoolName:       {authtypes.Burner, authtypes.Staking},
		universalstakingtypes.RewardsPoolName:       nil,
		extbridgetypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
		inheritancetypes.ModuleName:                  nil, // Inheritance module doesn't need minting/burning
		validatortypes.ModuleName:                    {authtypes.Burner}, // Burns slashed tokens
		bridgetypes.ModuleName:                       {authtypes.Minter, authtypes.Burner}, // Bridge needs mint/burn for wrapped assets
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
	EquityKeeper          equitykeeper.Keeper
	DexKeeper             dexkeeper.Keeper
	EscrowKeeper          *escrowkeeper.Keeper
	LendingKeeper         *lendingkeeper.Keeper
	GovernanceKeeper      *governancekeeper.Keeper
	AgentKeeper              agentkeeper.Keeper
	FeeAbstractionKeeper     *feeabstractionkeeper.Keeper
	UniversalStakingKeeper   *universalstakingkeeper.Keeper
	ExtBridgeKeeper          *extbridgekeeper.Keeper
	InheritanceKeeper        *inheritancekeeper.Keeper
	ExplorerKeeper           *explorerkeeper.Keeper
	ValidatorKeeper          *validatorkeeper.Keeper
	BridgeKeeper             *bridgekeeper.Keeper

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
	// Create address codecs for our chain's bech32 prefixes
	addressCodec := address.NewBech32Codec(Bech32PrefixAccAddr)
	validatorAddressCodec := address.NewBech32Codec(Bech32PrefixValAddr)

	// CRITICAL: Create InterfaceRegistry with proper address codecs.
	// This is required for tx simulation (gas estimation) and proper
	// address conversion in gRPC queries. Without this, CLI and Keplr
	// transactions fail with "InterfaceRegistry requires a proper address codec".
	signingOptions := txsigning.Options{
		FileResolver:          proto.HybridResolver,
		AddressCodec:          addressCodec,
		ValidatorAddressCodec: validatorAddressCodec,
	}
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	if err != nil {
		panic(err)
	}

	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()

	// CRITICAL: SDK v0.54-alpha requires creating a SigningContext with proper
	// address codecs for signature verification to work. Without this, signature
	// verification fails with "unable to verify single signer signature".
	signingContext, err := txsigning.NewContext(signingOptions)
	if err != nil {
		panic(err)
	}

	txConfig, err := authtx.NewTxConfigWithOptions(appCodec, authtx.ConfigOptions{
		EnabledSignModes: authtx.DefaultSignModes,
		SigningContext:   signingContext,
	})
	if err != nil {
		panic(err)
	}

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
		equitymodule.NewAppModuleBasic(appCodec),
		dexmodule.NewAppModuleBasic(appCodec),
		escrowmodule.AppModuleBasic{},
		lendingmodule.AppModuleBasic{},
		governancemodule.NewAppModuleBasic(),
		agentmodule.NewAppModuleBasic(appCodec),
		feeabstractionmodule.NewAppModuleBasic(appCodec),
		universalstakingmodule.NewAppModuleBasic(appCodec),
		extbridgemodule.NewAppModuleBasic(appCodec),
		inheritancemodule.NewAppModuleBasic(appCodec),
		explorermodule.NewAppModuleBasic(appCodec),
		validatormodule.NewAppModuleBasic(appCodec),
		bridgemodule.NewAppModuleBasic(appCodec),
	)

	basicManager.RegisterInterfaces(interfaceRegistry)

	// Explicitly register SDK module interfaces for gRPC gateway
	// CRITICAL: These registrations enable proper serialization of interface types
	// over REST API, which is required for Keplr wallet integration
	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)

	// Register crypto key types for proper serialization
	interfaceRegistry.RegisterImplementations((*cryptotypes.PubKey)(nil),
		&ed25519.PubKey{},
		&secp256k1.PubKey{},
	)

	// CRITICAL: Register account interface implementations
	// This fixes "no registered implementations of type types.AccountI" error
	// that prevents Keplr from querying account information
	interfaceRegistry.RegisterImplementations((*authtypes.AccountI)(nil),
		&authtypes.BaseAccount{},
		&authtypes.ModuleAccount{},
	)

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
		equitytypes.StoreKey,
		dextypes.StoreKey,
		escrowtypes.StoreKey,
		lendingtypes.StoreKey,
		governancetypes.StoreKey,
		agenttypes.StoreKey,
		feeabstractiontypes.StoreKey,
		universalstakingtypes.StoreKey,
		extbridgetypes.StoreKey,
		inheritancetypes.StoreKey,
		explorertypes.StoreKey,
		validatortypes.StoreKey,
		bridgetypes.StoreKey,
	)
	
	memKeys := storetypes.NewMemoryStoreKeys(
		hodltypes.MemStoreKey,
		equitytypes.MemStoreKey,
		dextypes.MemStoreKey,
		escrowtypes.MemStoreKey,
		lendingtypes.MemStoreKey,
		governancetypes.MemStoreKey,
		feeabstractiontypes.MemStoreKey,
		universalstakingtypes.MemStoreKey,
		extbridgetypes.MemStoreKey,
		inheritancetypes.MemStoreKey,
		explorertypes.MemStoreKey,
		validatortypes.MemStoreKey,
		bridgetypes.MemStoreKey,
	)

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
	// CRITICAL: Use the SAME addressCodec instance that was used for InterfaceRegistry
	// and SigningContext. This ensures the ante handler's signature verification uses
	// the same codec instance. Different instances may cause verification failures.
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addressCodec,
		Bech32PrefixAccAddr,
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

	// CRITICAL: Use the SAME validatorAddressCodec instance that was used for
	// InterfaceRegistry and SigningContext to ensure consistency.
	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress("gov").String(),
		validatorAddressCodec,
		address.NewBech32Codec(Bech32PrefixConsAddr),
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
		authtypes.NewModuleAddress(hodltypes.ModuleName).String(),
	)

	// Initialize Equity keeper (UniversalStakingKeeper wired later via SetStakingKeeper)
	app.EquityKeeper = *equitykeeper.NewKeeper(
		appCodec,
		keys[equitytypes.StoreKey],
		memKeys[equitytypes.MemStoreKey],
		app.BankKeeper,
		app.AccountKeeper,
		authtypes.NewModuleAddress(governancetypes.ModuleName).String(),
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

	// Initialize Escrow keeper (staking keeper set later via SetStakingKeeper)
	app.EscrowKeeper = escrowkeeper.NewKeeper(
		appCodec,
		keys[escrowtypes.StoreKey],
		memKeys[escrowtypes.MemStoreKey],
		app.BankKeeper,
		app.AccountKeeper,
		app.EquityKeeper,
	)

	// Initialize Lending keeper (staking keeper set later via SetStakingKeeper)
	app.LendingKeeper = lendingkeeper.NewKeeper(
		appCodec,
		keys[lendingtypes.StoreKey],
		memKeys[lendingtypes.MemStoreKey],
		app.BankKeeper,
		app.AccountKeeper,
		app.EquityKeeper,
		nil, // UniversalStakingKeeper - set later via SetStakingKeeper
	)

	// Initialize Governance keeper (UniversalStakingKeeper set after initialization)
	app.GovernanceKeeper = governancekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[governancetypes.StoreKey]),
		app.EquityKeeper,
		app.HODLKeeper,
		nil, // UniversalStakingKeeper - set later after staking keeper is initialized
		app.BankKeeper,
		authtypes.NewModuleAddress("gov").String(),
	)

	// Initialize Agent keeper
	app.AgentKeeper = agentkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[agenttypes.StoreKey]),
		logger,
		authtypes.NewModuleAddress("gov").String(),
		app.BankKeeper,
	)

	// Initialize Fee Abstraction keeper with DEX and Equity adapters
	dexAdapter := feeabstractionkeeper.NewDEXKeeperAdapter(&app.DexKeeper)
	equityAdapter := feeabstractionkeeper.NewEquityKeeperAdapter(app.EquityKeeper)
	app.FeeAbstractionKeeper = feeabstractionkeeper.NewKeeper(
		appCodec,
		keys[feeabstractiontypes.StoreKey],
		authtypes.NewModuleAddress("gov").String(),
		app.AccountKeeper,
		app.BankKeeper,
		dexAdapter,
		equityAdapter,
	)

	// Initialize Universal Staking keeper
	app.UniversalStakingKeeper = universalstakingkeeper.NewKeeper(
		appCodec,
		keys[universalstakingtypes.StoreKey],
		memKeys[universalstakingtypes.MemStoreKey],
		authtypes.NewModuleAddress("gov").String(), // Governance authority for param updates
		app.BankKeeper,
		app.AccountKeeper,
		app.GovernanceKeeper,
		nil, // ValidatorKeeper removed - staking is now self-sufficient
		app.FeeAbstractionKeeper,
	)

	// Initialize Validator keeper (for validator tier management and business verification)
	app.ValidatorKeeper = validatorkeeper.NewKeeper(
		appCodec,
		keys[validatortypes.StoreKey],
		memKeys[validatortypes.MemStoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
	)

	// Wire universal staking into governance module
	app.GovernanceKeeper.SetStakingKeeper(app.UniversalStakingKeeper)

	// Wire account keeper into governance for module account detection
	// Prevents treasury/escrow/DEX module accounts from voting
	app.GovernanceKeeper.SetAccountKeeper(app.AccountKeeper)

	// Wire universal staking into lending module
	app.LendingKeeper.SetStakingKeeper(app.UniversalStakingKeeper)

	// Wire universal staking into escrow module (for moderator tier checks)
	app.EscrowKeeper.SetStakingKeeper(app.UniversalStakingKeeper)

	// Wire universal staking into equity module (for listing tier checks and stake locks)
	app.EquityKeeper.SetStakingKeeper(app.UniversalStakingKeeper)

	// Wire validator keeper into equity module (for audit verification)
	app.EquityKeeper.SetValidatorKeeper(NewValidatorKeeperAdapter(app.ValidatorKeeper))

	// TODO: Wire DEX and HODL keepers into agent module (requires adapters for interface compatibility)
	// app.AgentKeeper.SetDEXKeeper(&app.DexKeeper)
	// app.AgentKeeper.SetHODLKeeper(&app.HODLKeeper)

	// Initialize External Bridge keeper
	app.ExtBridgeKeeper = extbridgekeeper.NewKeeper(
		appCodec,
		keys[extbridgetypes.StoreKey],
		memKeys[extbridgetypes.MemStoreKey],
		app.BankKeeper,
		app.AccountKeeper,
		app.UniversalStakingKeeper,
		app.EscrowKeeper, // For ban checking
		authtypes.NewModuleAddress("gov").String(),
	)

	// Initialize Inheritance keeper (Dead Man Switch / Next of Kin)
	// CRITICAL: Adapters are used to match interface signatures between keepers
	banKeeperAdapter := NewBanKeeperAdapter(app.EscrowKeeper)
	equityKeeperAdapter := NewEquityKeeperAdapter(app.EquityKeeper)
	app.InheritanceKeeper = inheritancekeeper.NewKeeper(
		appCodec,
		keys[inheritancetypes.StoreKey],
		memKeys[inheritancetypes.MemStoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		equityKeeperAdapter, // Adapts equity keeper for inheritance interface
		banKeeperAdapter,    // Adapts escrow ban checks to inheritance interface
		app.HODLKeeper,      // For HODL token operations
		authtypes.NewModuleAddress("gov").String(),
	)

	// Wire staking, lending, and escrow keepers into inheritance (for position transfers)
	// Use adapters to match interface signatures
	stakingAdapter := NewStakingKeeperAdapter(app.UniversalStakingKeeper)
	lendingAdapter := NewLendingKeeperAdapter(app.LendingKeeper)
	escrowAdapter := NewEscrowKeeperAdapter(app.EscrowKeeper)
	app.InheritanceKeeper.SetStakingKeeper(stakingAdapter)
	app.InheritanceKeeper.SetLendingKeeper(lendingAdapter)
	app.InheritanceKeeper.SetEscrowKeeper(escrowAdapter)

	// Initialize Validator keeper
	app.ValidatorKeeper = validatorkeeper.NewKeeper(
		appCodec,
		keys[validatortypes.StoreKey],
		memKeys[validatortypes.MemStoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
	)

	// Initialize Bridge keeper (requires validator keeper for multi-sig validation)
	app.BridgeKeeper = bridgekeeper.NewKeeper(
		appCodec,
		keys[bridgetypes.StoreKey],
		memKeys[bridgetypes.MemStoreKey],
		app.BankKeeper,
		app.ValidatorKeeper,
		authtypes.NewModuleAddress("gov").String(),
	)

	// Initialize Explorer keeper (requires other keepers for indexing)
	// Note: TxDecoder is obtained from txConfig
	// Create adapters to match explorer keeper interfaces
	explorerEquityAdapter := NewExplorerEquityKeeperAdapter(app.EquityKeeper)
	explorerHODLAdapter := NewExplorerHODLKeeperAdapter(app.HODLKeeper)
	explorerStakingAdapter := NewExplorerStakingKeeperAdapter(app.UniversalStakingKeeper)
	explorerBankAdapter := NewExplorerBankKeeperAdapter(app.BankKeeper)
	app.ExplorerKeeper = explorerkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[explorertypes.StoreKey]),
		txConfig.TxDecoder(),
		explorerEquityAdapter,
		explorerHODLAdapter,
		explorerStakingAdapter,
		app.GovernanceKeeper,
		explorerBankAdapter,
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
		equitymodule.NewAppModule(appCodec, app.EquityKeeper, app.AccountKeeper, app.BankKeeper),
		dexmodule.NewAppModule(appCodec, app.DexKeeper, app.AccountKeeper, app.BankKeeper, app.EquityKeeper, app.HODLKeeper),
		escrowmodule.NewAppModule(app.EscrowKeeper),
		lendingmodule.NewAppModule(app.LendingKeeper),
		governancemodule.NewAppModule(app.GovernanceKeeper),
		agentmodule.NewAppModule(appCodec, app.AgentKeeper),
		feeabstractionmodule.NewAppModule(appCodec, app.FeeAbstractionKeeper, app.AccountKeeper, app.BankKeeper),
		universalstakingmodule.NewAppModule(appCodec, app.UniversalStakingKeeper, app.AccountKeeper, app.BankKeeper),
		extbridgemodule.NewAppModule(appCodec, *app.ExtBridgeKeeper),
		inheritancemodule.NewAppModule(appCodec, *app.InheritanceKeeper, app.AccountKeeper, app.BankKeeper),
		validatormodule.NewAppModule(appCodec, *app.ValidatorKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		explorermodule.NewAppModule(appCodec, app.ExplorerKeeper),
		bridgemodule.NewAppModule(appCodec, *app.BridgeKeeper),
	)

	app.MM.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		stakingtypes.ModuleName,
		validatortypes.ModuleName, // Validator module first for tier management
		hodltypes.ModuleName,
		equitytypes.ModuleName,
		dextypes.ModuleName,
		escrowtypes.ModuleName,
		lendingtypes.ModuleName,
		governancetypes.ModuleName,
		agenttypes.ModuleName,
		feeabstractiontypes.ModuleName,
		universalstakingtypes.ModuleName,
		extbridgetypes.ModuleName,
		inheritancetypes.ModuleName,
		explorertypes.ModuleName, // Explorer indexing after all state changes
		bridgetypes.ModuleName,
	)

	app.MM.SetOrderEndBlockers(
		stakingtypes.ModuleName,
		validatortypes.ModuleName, // Validator vesting and tier updates
		hodltypes.ModuleName,
		equitytypes.ModuleName,
		dextypes.ModuleName,
		escrowtypes.ModuleName,
		lendingtypes.ModuleName,
		governancetypes.ModuleName,
		agenttypes.ModuleName,
		feeabstractiontypes.ModuleName,
		universalstakingtypes.ModuleName,
		extbridgetypes.ModuleName,
		inheritancetypes.ModuleName, // Process inheritance: inactivity checks, grace periods, claims
		explorertypes.ModuleName,    // Explorer indexing after all state changes
		bridgetypes.ModuleName,
	)

	genesisModuleOrder := []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		upgradetypes.ModuleName,
		genutiltypes.ModuleName,
		consensusparamtypes.ModuleName,
		validatortypes.ModuleName, // Validator module before equity (equity needs validator keeper)
		hodltypes.ModuleName,
		equitytypes.ModuleName,
		dextypes.ModuleName,
		escrowtypes.ModuleName,
		lendingtypes.ModuleName,
		governancetypes.ModuleName,
		agenttypes.ModuleName,
		feeabstractiontypes.ModuleName,
		universalstakingtypes.ModuleName,
		validatortypes.ModuleName,
		extbridgetypes.ModuleName,
		inheritancetypes.ModuleName,
		bridgetypes.ModuleName,
		explorertypes.ModuleName, // Explorer last - no genesis state, indexes other modules
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

	// configure ante handler with fee abstraction
	anteHandler, err := NewAnteHandler(AnteHandlerOptions{
		AccountKeeper:        app.AccountKeeper,
		BankKeeper:           app.BankKeeper,
		SignModeHandler:      txConfig.SignModeHandler(),
		FeeAbstractionKeeper: app.FeeAbstractionKeeper,
	})
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
	// CRITICAL: Update client context with app's interface registry for proper type resolution
	// This ensures gRPC gateway can properly serialize interface types (AccountI, etc.)
	// which is required for Keplr wallet to query account information
	clientCtx = clientCtx.WithInterfaceRegistry(app.interfaceRegistry).
		WithCodec(app.appCodec).
		WithTxConfig(app.txConfig)

	// Register gRPC gateway routes with updated client context
	app.BasicManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// CRITICAL: Register tx service gRPC gateway routes for REST API
	// This enables /cosmos/tx/v1beta1/txs (broadcast) and /cosmos/tx/v1beta1/simulate endpoints
	// Without this, Keplr wallet cannot broadcast transactions via REST API
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
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
	equitymodule.NewAppModuleBasic(nil),
	dexmodule.NewAppModuleBasic(nil),
	escrowmodule.AppModuleBasic{},
	lendingmodule.AppModuleBasic{},
	governancemodule.NewAppModuleBasic(),
	feeabstractionmodule.NewAppModuleBasic(nil),
	universalstakingmodule.NewAppModuleBasic(nil),
	inheritancemodule.NewAppModuleBasic(nil),
	explorermodule.NewAppModuleBasic(nil),
	validatormodule.NewAppModuleBasic(nil),
	bridgemodule.NewAppModuleBasic(nil),
)

// AnteHandlerOptions are the options required for constructing a default SDK AnteHandler
// with fee abstraction support.
type AnteHandlerOptions struct {
	AccountKeeper        authkeeper.AccountKeeper
	BankKeeper           bankkeeper.Keeper
	SignModeHandler      *txsigning.HandlerMap
	FeeAbstractionKeeper *feeabstractionkeeper.Keeper
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer. This version includes fee abstraction support.
func NewAnteHandler(options AnteHandlerOptions) (sdk.AnteHandler, error) {
	return sdk.ChainAnteDecorators(
		ante.NewSetUpContextDecorator(),
		ante.NewExtensionOptionsDecorator(nil),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		// Fee abstraction decorator - BEFORE standard fee deduction
		// This intercepts txs with insufficient HODL and swaps equity if available
		feeabstractionante.NewFeeAbstractionDecorator(
			options.AccountKeeper,
			options.BankKeeper,
			options.FeeAbstractionKeeper,
		),
		ante.NewDeductFeeDecorator(
			options.AccountKeeper,
			options.BankKeeper,
			nil, // feegrant keeper
			nil, // txFeeChecker
		),
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, ante.DefaultSigVerificationGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
	), nil
}

// BanKeeperAdapter wraps the escrow keeper to provide the BanKeeper interface
// for the inheritance module
type BanKeeperAdapter struct {
	escrowKeeper *escrowkeeper.Keeper
}

// NewBanKeeperAdapter creates a new adapter for the ban keeper interface
func NewBanKeeperAdapter(escrowKeeper *escrowkeeper.Keeper) *BanKeeperAdapter {
	return &BanKeeperAdapter{escrowKeeper: escrowKeeper}
}

// IsAddressBanned checks if an address is banned (adapts escrow keeper's signature)
func (a *BanKeeperAdapter) IsAddressBanned(ctx sdk.Context, address string) bool {
	if a.escrowKeeper == nil {
		return false
	}
	banned, _ := a.escrowKeeper.IsAddressBanned(ctx, address, ctx.BlockTime())
	return banned
}

// EquityKeeperAdapter wraps the equity keeper for inheritance module
type EquityKeeperAdapter struct {
	equityKeeper equitykeeper.Keeper
}

// NewEquityKeeperAdapter creates a new adapter for the equity keeper interface
func NewEquityKeeperAdapter(equityKeeper equitykeeper.Keeper) *EquityKeeperAdapter {
	return &EquityKeeperAdapter{equityKeeper: equityKeeper}
}

// GetShareholding returns the shareholding for an owner (converts sdk.Context to context.Context)
func (a *EquityKeeperAdapter) GetShareholding(ctx sdk.Context, companyID uint64, classID string, owner string) (interface{}, bool) {
	return a.equityKeeper.GetShareholding(ctx, companyID, classID, owner)
}

// TransferShares transfers shares from one owner to another
func (a *EquityKeeperAdapter) TransferShares(ctx sdk.Context, companyID uint64, classID string, from string, to string, shares math.Int) error {
	return a.equityKeeper.TransferShares(ctx, companyID, classID, from, to, shares)
}

// GetCompany returns company information (converts sdk.Context to context.Context)
func (a *EquityKeeperAdapter) GetCompany(ctx context.Context, companyID uint64) (interface{}, bool) {
	return a.equityKeeper.GetCompany(ctx, companyID)
}

// GetAllHoldingsByAddress returns all holdings for an address
func (a *EquityKeeperAdapter) GetAllHoldingsByAddress(ctx sdk.Context, owner string) []interface{} {
	return a.equityKeeper.GetAllHoldingsByAddress(ctx, owner)
}

// StakingKeeperAdapter wraps the staking keeper for inheritance module
type StakingKeeperAdapter struct {
	stakingKeeper *universalstakingkeeper.Keeper
}

// NewStakingKeeperAdapter creates a new adapter for the staking keeper interface
func NewStakingKeeperAdapter(stakingKeeper *universalstakingkeeper.Keeper) *StakingKeeperAdapter {
	return &StakingKeeperAdapter{stakingKeeper: stakingKeeper}
}

// GetUserStake returns a user's stake as interface{} for inheritance module compatibility
func (a *StakingKeeperAdapter) GetUserStake(ctx sdk.Context, owner sdk.AccAddress) (interface{}, bool) {
	if a.stakingKeeper == nil {
		return nil, false
	}
	stake, found := a.stakingKeeper.GetUserStake(ctx, owner)
	if !found {
		return nil, false
	}
	return stake, true
}

// UnstakeForInheritance initiates unstaking with beneficiary as recipient
func (a *StakingKeeperAdapter) UnstakeForInheritance(ctx sdk.Context, owner sdk.AccAddress, recipient string, amount math.Int) error {
	if a.stakingKeeper == nil {
		return nil
	}
	return a.stakingKeeper.UnstakeForInheritance(ctx, owner, recipient, amount)
}

// LendingKeeperAdapter wraps the lending keeper for inheritance module
type LendingKeeperAdapter struct {
	lendingKeeper *lendingkeeper.Keeper
}

// NewLendingKeeperAdapter creates a new adapter for the lending keeper interface
func NewLendingKeeperAdapter(lendingKeeper *lendingkeeper.Keeper) *LendingKeeperAdapter {
	return &LendingKeeperAdapter{lendingKeeper: lendingKeeper}
}

// GetUserLoans returns all loans where user is borrower or lender as []interface{}
func (a *LendingKeeperAdapter) GetUserLoans(ctx sdk.Context, user string) []interface{} {
	if a.lendingKeeper == nil {
		return nil
	}
	loans := a.lendingKeeper.GetUserLoans(ctx, user)
	result := make([]interface{}, len(loans))
	for i, loan := range loans {
		result[i] = loan
	}
	return result
}

// TransferBorrowerPosition transfers a loan's borrower position
func (a *LendingKeeperAdapter) TransferBorrowerPosition(ctx sdk.Context, loanID uint64, from, to string) error {
	if a.lendingKeeper == nil {
		return nil
	}
	return a.lendingKeeper.TransferBorrowerPosition(ctx, loanID, from, to)
}

// TransferLenderPosition transfers a lending position
func (a *LendingKeeperAdapter) TransferLenderPosition(ctx sdk.Context, loanID uint64, from, to string) error {
	if a.lendingKeeper == nil {
		return nil
	}
	return a.lendingKeeper.TransferLenderPosition(ctx, loanID, from, to)
}

// EscrowKeeperAdapter wraps the escrow keeper for inheritance module
type EscrowKeeperAdapter struct {
	escrowKeeper *escrowkeeper.Keeper
}

// NewEscrowKeeperAdapter creates a new adapter for the escrow keeper interface
func NewEscrowKeeperAdapter(escrowKeeper *escrowkeeper.Keeper) *EscrowKeeperAdapter {
	return &EscrowKeeperAdapter{escrowKeeper: escrowKeeper}
}

// GetAllEscrows returns all escrows as []interface{}
func (a *EscrowKeeperAdapter) GetAllEscrows(ctx sdk.Context) []interface{} {
	if a.escrowKeeper == nil {
		return nil
	}
	escrows := a.escrowKeeper.GetAllEscrows(ctx)
	result := make([]interface{}, len(escrows))
	for i, escrow := range escrows {
		result[i] = escrow
	}
	return result
}

// TransferBuyerPosition transfers escrow buyer position
func (a *EscrowKeeperAdapter) TransferBuyerPosition(ctx sdk.Context, escrowID uint64, from, to string) error {
	if a.escrowKeeper == nil {
		return nil
	}
	return a.escrowKeeper.TransferBuyerPosition(ctx, escrowID, from, to)
}

// TransferSellerPosition transfers escrow seller position
func (a *EscrowKeeperAdapter) TransferSellerPosition(ctx sdk.Context, escrowID uint64, from, to string) error {
	if a.escrowKeeper == nil {
		return nil
	}
	return a.escrowKeeper.TransferSellerPosition(ctx, escrowID, from, to)
}

// ValidatorKeeperAdapter wraps the validator keeper to provide the ValidatorKeeper interface
// for the equity module (audit verification)
type ValidatorKeeperAdapter struct {
	validatorKeeper *validatorkeeper.Keeper
}

// NewValidatorKeeperAdapter creates a new adapter for the validator keeper interface
func NewValidatorKeeperAdapter(validatorKeeper *validatorkeeper.Keeper) *ValidatorKeeperAdapter {
	return &ValidatorKeeperAdapter{validatorKeeper: validatorKeeper}
}

// IsValidator checks if an address is currently a validator
func (a *ValidatorKeeperAdapter) IsValidator(ctx sdk.Context, addr sdk.AccAddress) bool {
	if a.validatorKeeper == nil {
		return false
	}

	// Convert AccAddress to validator address string for the validator keeper
	valAddr := sdk.ValAddress(addr.Bytes()).String()
	return a.validatorKeeper.IsValidatorActive(ctx, valAddr)
}