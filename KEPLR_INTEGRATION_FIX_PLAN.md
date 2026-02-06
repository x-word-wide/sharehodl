# Keplr Integration Fix Plan

## Problem Analysis

### Issue 1: gRPC Account Query Fails
**Error**: `no registered implementations of type types.AccountI`
**Endpoint**: `/cosmos/auth/v1beta1/accounts/{address}`
**Impact**: Keplr cannot retrieve account information needed for transaction signing

**Root Cause**:
1. Interface registry is created but not fully populated with all interface implementations
2. The `authtypes.RegisterInterfaces()` is called but might be missing crypto key registrations
3. The interface registry needs to be properly wired into the gRPC gateway client context

### Issue 2: CLI Module Commands Missing
**Error**: Commands like `sharehodld tx bank send` don't exist
**Impact**: Cannot test transactions via CLI, and debugging is difficult

**Root Cause**:
1. Module CLI commands are not registered in `cmd/sharehodld/root.go`
2. The `BasicManager.AddTxCommands()` and `BasicManager.AddQueryCommands()` are not called
3. Custom modules return `nil` for `GetTxCmd()` and `GetQueryCmd()` instead of implementing CLI

## Solution Design

### Fix 1: Complete Interface Registry Setup

**Location**: `app/app.go`

The interface registry needs proper registration in multiple places:

1. **During app creation** (line 205-211):
   - Already calling `std.RegisterInterfaces(interfaceRegistry)` ✓
   - Already calling `basicManager.RegisterInterfaces(interfaceRegistry)` ✓
   - Need to ensure crypto types are registered ✓ (lines 244-247)

2. **In RegisterAPIRoutes** (line 741-749):
   - Update client context with interface registry ✓ (line 744)
   - BUT: Need to ensure it's set BEFORE gateway registration

3. **In RegisterTxService** (line 763-765):
   - Pass interface registry to authtx.RegisterTxService ✓

**Required Changes**:
```go
// In NewShareHODLApp, after basicManager.RegisterInterfaces:
authtypes.RegisterInterfaces(interfaceRegistry)
banktypes.RegisterInterfaces(interfaceRegistry)
stakingtypes.RegisterInterfaces(interfaceRegistry)

// Register crypto implementations (already done at lines 244-247)
interfaceRegistry.RegisterImplementations((*cryptotypes.PubKey)(nil),
    &ed25519.PubKey{},
    &secp256k1.PubKey{},
)

// CRITICAL: Register account implementations
interfaceRegistry.RegisterImplementations((*authtypes.AccountI)(nil),
    &authtypes.BaseAccount{},
    &authtypes.ModuleAccount{},
)

// Register message implementations for all SDK modules
banktypes.RegisterInterfaces(interfaceRegistry)
stakingtypes.RegisterInterfaces(interfaceRegistry)
```

### Fix 2: Register Module CLI Commands

**Location**: `cmd/sharehodld/root.go`

The `queryCommand()` and `txCommand()` functions need to register module commands:

```go
func queryCommand() *cobra.Command {
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

    // ADD THIS: Register module query commands
    app.ModuleBasics.AddQueryCommands(cmd)

    cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")
    return cmd
}

func txCommand() *cobra.Command {
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

    // ADD THIS: Register module tx commands
    app.ModuleBasics.AddTxCommands(cmd)

    cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")
    return cmd
}
```

### Fix 3: Implement Custom Module CLI (Optional but Recommended)

For custom modules (hodl, equity, dex), we should implement CLI commands:

**Example for x/hodl/client/cli/tx.go**:
```go
package cli

import (
    "github.com/spf13/cobra"
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/client/flags"
    "github.com/cosmos/cosmos-sdk/client/tx"
    "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

func GetTxCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   types.ModuleName,
        Short: "HODL transaction subcommands",
        RunE:  client.ValidateCmd,
    }

    cmd.AddCommand(
        GetCmdMintHODL(),
        GetCmdBurnHODL(),
    )

    return cmd
}

func GetCmdMintHODL() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mint [amount]",
        Short: "Mint HODL tokens by depositing collateral",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
        },
    }
    flags.AddTxFlagsToCmd(cmd)
    return cmd
}
```

## Implementation Priority

### Phase 1: Critical Fixes (Keplr Integration)
1. ✅ Add complete interface registrations in `app/app.go`
2. ✅ Fix RegisterAPIRoutes to properly set interface registry
3. ✅ Register module CLI commands in `root.go`

### Phase 2: Enhanced CLI Support (Optional)
1. Create CLI packages for custom modules
2. Implement GetTxCmd() and GetQueryCmd() for each module
3. Add comprehensive command documentation

## Testing Plan

### Test 1: gRPC Account Query
```bash
# After fixes, this should return account info:
curl http://localhost:1317/cosmos/auth/v1beta1/accounts/hodl1...
```

### Test 2: CLI Commands
```bash
# Should list bank commands:
sharehodld tx bank --help

# Should work:
sharehodld tx bank send [from] [to] [amount] --chain-id sharehodl-1

# Should list balances:
sharehodld query bank balances [address]
```

### Test 3: Keplr Integration
1. Open Keplr wallet
2. Add ShareHODL chain
3. Try to send a transaction
4. Should succeed without "account not found" errors

## Expected Outcomes

After implementing these fixes:

1. **gRPC Gateway**: Account queries will serialize properly
2. **CLI**: Full module command support (bank, staking, custom modules)
3. **Keplr**: Can retrieve account info and submit transactions
4. **Development**: Easier debugging with CLI commands

## Code Locations

Files to modify:
- `/Users/sharehodl/.claude-worktrees/sharehodl-blockchain/exciting-robinson/app/app.go`
- `/Users/sharehodl/.claude-worktrees/sharehodl-blockchain/exciting-robinson/cmd/sharehodld/root.go`

Optional (Phase 2):
- `x/hodl/client/cli/tx.go` (create)
- `x/hodl/client/cli/query.go` (create)
- `x/hodl/module.go` (update GetTxCmd/GetQueryCmd)
- Similar for equity, dex, etc.
