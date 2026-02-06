package main

import (
	"errors"
	"io"
	"os"

	"cosmossdk.io/log"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	cmtcfg "github.com/cometbft/cometbft/v2/config"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sharehodl/sharehodl-blockchain/app"
)

// NewRootCmd creates a new root command for sharehodld. It is called once in the
// main function.
func NewRootCmd() *cobra.Command {
	// Create encoding config with all module registrations
	encodingConfig := app.MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(app.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("")

	rootCmd := &cobra.Command{
		Use:   "sharehodld",
		Short: "ShareHODL Blockchain Daemon",
		Long: `ShareHODL is a purpose-built blockchain for a decentralized stock exchange.
It enables any legitimate business to raise capital from anyone, anywhere,
and trade 24/7 with instant settlement.`,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customCMTConfig := initCometBFTConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customCMTConfig)
		},
	}

	initRootCmd(rootCmd, encodingConfig.Codec, encodingConfig.TxConfig, encodingConfig.InterfaceRegistry, app.ModuleBasics)

	return rootCmd
}

// ShareHODL bech32 address prefixes
const (
	// Bech32PrefixAccAddr defines the bech32 prefix of an account's address
	Bech32PrefixAccAddr = "hodl"
	// Bech32PrefixAccPub defines the bech32 prefix of an account's public key
	Bech32PrefixAccPub = "hodlpub"
	// Bech32PrefixValAddr defines the bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = "hodlvaloper"
	// Bech32PrefixValPub defines the bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = "hodlvaloperpub"
	// Bech32PrefixConsAddr defines the bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = "hodlvalcons"
	// Bech32PrefixConsPub defines the bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = "hodlvalconspub"
)

// initRootCmd builds the root command for sharehodld.
func initRootCmd(
	rootCmd *cobra.Command,
	cdc codec.Codec,
	txConfig client.TxConfig,
	interfaceRegistry codectypes.InterfaceRegistry,
	basicManager module.BasicManager,
) {
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
	cfg.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	cfg.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)
	cfg.Seal()

	rootCmd.AddCommand(
		genutilcli.InitCmd(basicManager, app.DefaultNodeHome),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
	)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, appExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, and tx child commands
	rootCmd.AddCommand(
		genesisCommand(txConfig, basicManager),
		queryCommand(basicManager),
		txCommand(basicManager),
		keys.Commands(),
		hashCommand(), // Hash encoding/decoding utilities
	)
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}

// genesisCommand builds genesis-related `sharehodld genesis` command. Users may provide application specific commands as a parameter.
func genesisCommand(txConfig client.TxConfig, basicManager module.BasicManager, cmds ...*cobra.Command) *cobra.Command {
	cmd := genutilcli.GenesisCoreCommand(txConfig, basicManager, app.DefaultNodeHome)

	for _, subCmd := range cmds {
		cmd.AddCommand(subCmd)
	}
	return cmd
}

func queryCommand(basicManager module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.ValidatorCommand(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
	)

	// CRITICAL: Register module query commands
	// This adds commands like "sharehodld query bank balances"
	// which are essential for querying chain state via CLI
	//
	// NOTE: We manually add query commands instead of using basicManager.AddQueryCommands()
	// because some SDK modules require properly initialized codecs which aren't available
	// in the global ModuleBasics variable. Query commands generally don't need codecs.
	for _, module := range basicManager {
		if queryModule, ok := module.(interface{ GetQueryCmd() *cobra.Command }); ok {
			if queryCmd := queryModule.GetQueryCmd(); queryCmd != nil {
				cmd.AddCommand(queryCmd)
			}
		}
	}

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func txCommand(basicManager module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
	)

	// CRITICAL: Add bank tx commands directly with proper address codec.
	// The ModuleBasics bank module has nil address codec which causes panics.
	// We create the address codec here to match our chain's bech32 prefix.
	addressCodec := address.NewBech32Codec(Bech32PrefixAccAddr)
	cmd.AddCommand(bankcli.NewTxCmd(addressCodec))

	// Register other module transaction commands
	// Skip bank (already added above) and modules that panic due to nil codecs.
	for _, module := range basicManager {
		// Skip bank module - we added it manually above
		if module.Name() == "bank" {
			continue
		}
		if txModule, ok := module.(interface{ GetTxCmd() *cobra.Command }); ok {
			// Safely get command, skip if it panics (due to nil codec)
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Skip modules that panic due to uninitialized codecs
						// This is expected for SDK modules like staking in v0.54
					}
				}()
				if txCmd := txModule.GetTxCmd(); txCmd != nil {
					cmd.AddCommand(txCmd)
				}
			}()
		}
	}

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

// newApp creates the ShareHODL application.
func newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)

	return app.NewShareHODLApp(
		logger, db, traceStore, true,
		appOpts,
		baseappOptions...,
	)
}

// appExport creates a new ShareHODL app for export.
func appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var sharehodlApp *app.ShareHODLApp
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}

	// overwrite the FlagInvCheckPeriod
	viperAppOpts.Set(server.FlagInvCheckPeriod, 1)
	appOpts = viperAppOpts

	var loadLatest bool
	if height == -1 {
		loadLatest = true
	}

	sharehodlApp = app.NewShareHODLApp(
		logger,
		db,
		traceStore,
		loadLatest,
		appOpts,
	)

	if height != -1 {
		if err := sharehodlApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return sharehodlApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	type CustomAppConfig struct {
		serverconfig.Config
	}

	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()
	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In ShareHODL, we set default minimum gas prices.
	srvCfg.MinGasPrices = "0.0025uhodl"

	customAppConfig := CustomAppConfig{
		Config: *srvCfg,
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate

	return customAppTemplate, customAppConfig
}

// initCometBFTConfig helps to override default CometBFT Config values.
// return cmtcfg.DefaultConfig if no custom configuration is required for the application.
func initCometBFTConfig() *cmtcfg.Config {
	cfg := cmtcfg.DefaultConfig()

	// These fields should be configured for ShareHODL
	cfg.P2P.MaxNumInboundPeers = 40
	cfg.P2P.MaxNumOutboundPeers = 10
	cfg.P2P.FlushThrottleTimeout = 100000000 // 100ms in nanoseconds
	cfg.P2P.MaxPacketMsgPayloadSize = 1024

	// Consensus params for fast finality (CometBFT v2)
	cfg.Consensus.TimeoutPropose = 2000000000 // 2s in nanoseconds
	cfg.Consensus.TimeoutCommit = 2000000000  // 2s in nanoseconds

	return cfg
}