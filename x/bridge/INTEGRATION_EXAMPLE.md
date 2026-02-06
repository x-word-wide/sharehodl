# Bridge Module Integration Example

## Complete Integration Guide

### Step 1: Update `app/app.go`

Add imports:
```go
import (
    // ... existing imports
    bridgekeeper "github.com/sharehodl/sharehodl-blockchain/x/bridge/keeper"
    bridgetypes "github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
    bridgemodule "github.com/sharehodl/sharehodl-blockchain/x/bridge"
)
```

Add to App struct:
```go
type ShareHodlApp struct {
    // ... existing keepers
    BridgeKeeper *bridgekeeper.Keeper
}
```

Add store keys:
```go
keys := storetypes.NewKVStoreKeys(
    authtypes.StoreKey,
    banktypes.StoreKey,
    stakingtypes.StoreKey,
    // ... other existing keys
    bridgetypes.StoreKey,
)

memKeys := storetypes.NewMemoryStoreKeys(
    // ... existing mem keys
    bridgetypes.MemStoreKey,
)
```

Create bridge keeper (after bank and validator keepers):
```go
app.BridgeKeeper = bridgekeeper.NewKeeper(
    appCodec,
    keys[bridgetypes.StoreKey],
    memKeys[bridgetypes.MemStoreKey],
    app.BankKeeper,
    app.ValidatorKeeper,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)
```

Add to module manager:
```go
app.ModuleManager = module.NewManager(
    auth.NewAppModule(appCodec, app.AccountKeeper, nil, app.GetSubspace(authtypes.ModuleName)),
    bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
    // ... other modules
    bridgemodule.NewAppModule(appCodec, *app.BridgeKeeper),
)
```

Set module order:
```go
app.ModuleManager.SetOrderBeginBlockers(
    // ... existing modules
    bridgetypes.ModuleName,
)

app.ModuleManager.SetOrderEndBlockers(
    // ... existing modules
    bridgetypes.ModuleName, // IMPORTANT: After validator module
)

app.ModuleManager.SetOrderInitGenesis(
    authtypes.ModuleName,
    banktypes.ModuleName,
    // ... other modules
    bridgetypes.ModuleName, // After bank, before anything that uses it
)
```

Add to module account permissions:
```go
maccPerms := map[string][]string{
    authtypes.FeeCollectorName:     nil,
    stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
    // ... other module accounts
    bridgetypes.ModuleName:         {authtypes.Minter, authtypes.Burner},
}
```

### Step 2: Update `genesis.json`

Add bridge module genesis state:
```json
{
  "app_state": {
    "auth": { ... },
    "bank": { ... },
    "bridge": {
      "params": {
        "swap_fee_rate": "0.001000000000000000",
        "min_reserve_ratio": "0.200000000000000000",
        "max_withdrawal_percent": "0.100000000000000000",
        "withdrawal_voting_period": "604800",
        "required_validator_approval": "0.667000000000000000",
        "bridge_enabled": true
      },
      "supported_assets": [
        {
          "denom": "uusdt",
          "swap_rate_to_hodl": "1.000000000000000000",
          "swap_rate_from_hodl": "1.000000000000000000",
          "min_swap_amount": "1000000",
          "max_swap_amount": "1000000000000",
          "enabled": true,
          "added_at": "2026-01-29T00:00:00Z"
        },
        {
          "denom": "uusdc",
          "swap_rate_to_hodl": "1.000000000000000000",
          "swap_rate_from_hodl": "1.000000000000000000",
          "min_swap_amount": "1000000",
          "max_swap_amount": "1000000000000",
          "enabled": true,
          "added_at": "2026-01-29T00:00:00Z"
        },
        {
          "denom": "udai",
          "swap_rate_to_hodl": "1.000000000000000000",
          "swap_rate_from_hodl": "1.000000000000000000",
          "min_swap_amount": "1000000",
          "max_swap_amount": "1000000000000",
          "enabled": true,
          "added_at": "2026-01-29T00:00:00Z"
        },
        {
          "denom": "ubtc",
          "swap_rate_to_hodl": "100000.000000000000000000",
          "swap_rate_from_hodl": "0.000010000000000000",
          "min_swap_amount": "100000",
          "max_swap_amount": "100000000000",
          "enabled": true,
          "added_at": "2026-01-29T00:00:00Z"
        },
        {
          "denom": "ueth",
          "swap_rate_to_hodl": "3500.000000000000000000",
          "swap_rate_from_hodl": "0.000285700000000000",
          "min_swap_amount": "1000000",
          "max_swap_amount": "10000000000000",
          "enabled": true,
          "added_at": "2026-01-29T00:00:00Z"
        }
      ],
      "reserves": [],
      "swap_records": [],
      "withdrawal_proposals": [],
      "next_swap_id": "1",
      "next_proposal_id": "1"
    }
  }
}
```

### Step 3: Usage Examples

#### Swap In (USDT → HODL)
```bash
# User swaps 100 USDT for HODL
sharehodld tx bridge swap-in uusdt 100000000 \
  --from alice \
  --chain-id sharehodl-1 \
  --gas auto \
  --gas-adjustment 1.5

# Response will include:
# - hodl_received: 99900000 (100 USDT - 0.1% fee)
# - fee_charged: 100000
# - swap_id: 1
```

#### Swap Out (HODL → USDT)
```bash
# User swaps 100 HODL for USDT
sharehodld tx bridge swap-out 100000000 uusdt \
  --from alice \
  --chain-id sharehodl-1 \
  --gas auto \
  --gas-adjustment 1.5

# Response will include:
# - asset_received: 99900000 (100 USDT - 0.1% fee)
# - fee_charged: 100000
# - swap_id: 2
```

#### Validator Deposit
```bash
# Validator deposits BTC to reserve (no fee)
sharehodld tx bridge validator-deposit ubtc 100000000 \
  --from validator1 \
  --chain-id sharehodl-1 \
  --gas auto
```

#### Propose Withdrawal
```bash
# Validator proposes withdrawal for investment
sharehodld tx bridge propose-withdrawal \
  sharehodl1recipient... \
  ubtc \
  50000000 \
  "Investment in Bitcoin mining operations" \
  --from validator1 \
  --chain-id sharehodl-1 \
  --gas auto

# Returns proposal_id: 1
```

#### Vote on Withdrawal
```bash
# Validator votes YES on proposal
sharehodld tx bridge vote-withdrawal 1 true \
  --from validator1 \
  --chain-id sharehodl-1 \
  --gas auto

# Validator votes NO on proposal
sharehodld tx bridge vote-withdrawal 1 false \
  --from validator2 \
  --chain-id sharehodl-1 \
  --gas auto
```

#### Query Examples
```bash
# Query bridge params
sharehodld query bridge params

# Query supported assets
sharehodld query bridge supported-assets

# Query reserve balances
sharehodld query bridge reserves

# Query specific swap record
sharehodld query bridge swap 1

# Query withdrawal proposal
sharehodld query bridge proposal 1
```

### Step 4: Testing

Create test file `x/bridge/integration_test.go`:
```go
package bridge_test

import (
    "testing"

    "cosmossdk.io/math"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/stretchr/testify/require"

    "github.com/sharehodl/sharehodl-blockchain/testutil"
    "github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
)

func TestSwapInIntegration(t *testing.T) {
    // Setup test app
    app := testutil.Setup(t)
    ctx := app.BaseApp.NewContext(false)

    // Create test user
    user := testutil.CreateTestAddrs(1)[0]

    // Fund user with USDT
    usdtCoin := sdk.NewCoin("uusdt", math.NewInt(100_000_000))
    require.NoError(t, app.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(usdtCoin)))
    require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, user, sdk.NewCoins(usdtCoin)))

    // Execute swap in
    hodlReceived, fee, err := app.BridgeKeeper.SwapIn(ctx, user, "uusdt", math.NewInt(100_000_000))
    require.NoError(t, err)

    // Verify amounts
    expectedFee := math.NewInt(100_000) // 0.1% of 100 USDT
    expectedHODL := math.NewInt(99_900_000)
    require.Equal(t, expectedFee, fee)
    require.Equal(t, expectedHODL, hodlReceived)

    // Verify user balance
    hodlBalance := app.BankKeeper.GetBalance(ctx, user, "uhodl")
    require.Equal(t, expectedHODL, hodlBalance.Amount)

    // Verify reserve balance
    reserveBalance := app.BridgeKeeper.GetReserveBalance(ctx, "uusdt")
    require.Equal(t, math.NewInt(100_000_000), reserveBalance)
}

func TestValidatorWithdrawalGovernance(t *testing.T) {
    app := testutil.Setup(t)
    ctx := app.BaseApp.NewContext(false)

    // Create validators
    validators := testutil.CreateTestValidators(3)

    // Add reserve balance
    app.BridgeKeeper.AddToReserve(ctx, "uusdt", math.NewInt(1_000_000_000))

    // Propose withdrawal
    proposalID, err := app.BridgeKeeper.ProposeWithdrawal(
        ctx,
        validators[0],
        validators[0],
        "uusdt",
        math.NewInt(100_000_000),
        "Test withdrawal",
    )
    require.NoError(t, err)
    require.Equal(t, uint64(1), proposalID)

    // Vote YES from 2/3 validators
    require.NoError(t, app.BridgeKeeper.VoteOnWithdrawal(ctx, validators[0], proposalID, true))
    require.NoError(t, app.BridgeKeeper.VoteOnWithdrawal(ctx, validators[1], proposalID, true))

    // Verify proposal is approved
    proposal, found := app.BridgeKeeper.GetWithdrawalProposal(ctx, proposalID)
    require.True(t, found)
    require.Equal(t, types.ProposalStatusApproved, proposal.Status)

    // Execute withdrawal
    require.NoError(t, app.BridgeKeeper.ExecuteWithdrawal(ctx, proposalID))

    // Verify reserve was reduced
    newReserve := app.BridgeKeeper.GetReserveBalance(ctx, "uusdt")
    require.Equal(t, math.NewInt(900_000_000), newReserve)
}
```

### Step 5: Event Handling

In your application, listen for bridge events:
```go
// Listen for swap events
app.EventManager().SubscribeToEvent("bridge_swap_in", func(event sdk.Event) {
    for _, attr := range event.Attributes {
        if attr.Key == "user" {
            // Track user swaps
        }
        if attr.Key == "hodl_received" {
            // Track HODL minted
        }
    }
})

// Listen for withdrawal events
app.EventManager().SubscribeToEvent("bridge_withdrawal_executed", func(event sdk.Event) {
    // Track withdrawals for auditing
})
```

### Step 6: Monitoring

Key metrics to monitor:
```go
// Reserve ratio
ratio := app.BridgeKeeper.GetReserveRatio(ctx)
if ratio.LT(types.DefaultMinReserveRatio) {
    // Alert: Reserve ratio below minimum
}

// Total reserves
reserves := app.BridgeKeeper.GetAllReserveBalances(ctx)
totalValue := app.BridgeKeeper.GetTotalReserveValue(ctx)

// Active withdrawal proposals
proposals := app.BridgeKeeper.GetAllWithdrawalProposals(ctx)
activeProposals := filterActive(proposals)

// Recent swaps
swaps := app.BridgeKeeper.GetAllSwapRecords(ctx)
recentSwaps := filterRecent(swaps, 24*time.Hour)
```

## Complete Integration Checklist

- [ ] Add imports to `app/app.go`
- [ ] Add BridgeKeeper to App struct
- [ ] Add bridge store keys
- [ ] Create bridge keeper instance
- [ ] Add to module manager
- [ ] Set module order (BeginBlock, EndBlock, InitGenesis)
- [ ] Add module account with mint/burn permissions
- [ ] Update `genesis.json` with bridge state
- [ ] Build and test
- [ ] Run integration tests
- [ ] Monitor events and metrics
- [ ] Document for users

## Common Issues

### Issue 1: Module account not found
**Solution**: Ensure bridge module is added to `maccPerms` with `authtypes.Minter, authtypes.Burner`

### Issue 2: Reserve ratio violation
**Solution**: Ensure sufficient reserves before enabling swaps. Validators should deposit initial reserves.

### Issue 3: Validator not found
**Solution**: Ensure ValidatorKeeper is properly initialized and passed to BridgeKeeper

### Issue 4: Genesis validation fails
**Solution**: Check all asset swap rates are positive and params are within valid ranges

## Production Deployment

Before mainnet:
1. ✅ Security audit
2. ✅ Testnet deployment
3. ✅ Load testing
4. ✅ Reserve ratio monitoring
5. ✅ Governance testing
6. ✅ Emergency pause mechanism
7. ✅ Oracle integration (for dynamic rates)
8. ✅ Multi-signature withdrawals
9. ✅ Rate limiting
10. ✅ Circuit breakers
