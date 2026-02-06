# Bridge Module

The Bridge module implements a **Reserve-Backed Bridge** system for the ShareHODL blockchain, enabling users to swap between HODL stablecoin and external assets (USDT, USDC, DAI, BTC, ETH).

## Overview

The bridge operates as a two-way swap mechanism backed by reserves:
- **Swap In**: Users deposit external assets → receive HODL
- **Swap Out**: Users deposit HODL → receive external assets
- **Validator Deposits**: Validators can deposit assets to increase reserves (no fees)
- **Governance Withdrawals**: Validators vote on reserve withdrawals (2/3 approval required)

## Key Features

### 1. Swap Operations
- **0.1% swap fee** on all user swaps (configurable via governance)
- Atomic swap execution
- Configurable min/max swap amounts per asset
- Reserve ratio enforcement (20% minimum by default)

### 2. Supported Assets
Default supported assets:
- `uusdt` - USDT (1:1 with HODL)
- `uusdc` - USDC (1:1 with HODL)
- `udai` - DAI (1:1 with HODL)
- `ubtc` - Bitcoin (1 BTC = 100,000 HODL)
- `ueth` - Ethereum (1 ETH = 3,500 HODL)

Each asset has:
- Configurable swap rates (to/from HODL)
- Min/max swap limits
- Enable/disable toggle

### 3. Reserve Management
- All swap-in assets go to module reserve
- Validator deposits add to reserve (fee-free)
- Reserve ratio monitored: `total_reserve_value / total_hodl_supply >= 20%`
- Withdrawals blocked if ratio would fall below minimum

### 4. Governance System
Validators can propose and vote on reserve withdrawals:
- **Proposal Requirements**:
  - Must be an active validator
  - Maximum 10% of reserve per proposal
  - Must maintain 20% minimum reserve ratio
  - 7-day voting period (configurable)

- **Voting**:
  - Only active validators can vote
  - Requires 66.7% approval (2/3 majority)
  - One vote per validator
  - Proposals expire after voting period

- **Execution**:
  - Approved proposals auto-execute at EndBlock
  - Expired proposals marked as expired
  - Executed proposals recorded with timestamp

## Message Types

### MsgSwapIn
Swap external asset for HODL.

```go
type MsgSwapIn struct {
    User        string   // User address
    InputDenom  string   // Asset denomination (e.g., "uusdt")
    InputAmount math.Int // Amount to swap
}
```

**Example**:
```bash
sharehodld tx bridge swap-in uusdt 1000000000 \
  --from alice \
  --chain-id sharehodl-1
```

### MsgSwapOut
Swap HODL for external asset.

```go
type MsgSwapOut struct {
    User        string   // User address
    HodlAmount  math.Int // HODL amount to swap
    OutputDenom string   // Desired asset (e.g., "uusdt")
}
```

**Example**:
```bash
sharehodld tx bridge swap-out 1000000000 uusdt \
  --from alice \
  --chain-id sharehodl-1
```

### MsgValidatorDeposit
Validator deposits assets to reserve (no fee).

```go
type MsgValidatorDeposit struct {
    Validator string   // Validator address
    Denom     string   // Asset denomination
    Amount    math.Int // Amount to deposit
}
```

**Example**:
```bash
sharehodld tx bridge validator-deposit ubtc 100000000 \
  --from validator1 \
  --chain-id sharehodl-1
```

### MsgProposeWithdrawal
Propose withdrawal from reserve.

```go
type MsgProposeWithdrawal struct {
    Proposer  string   // Validator proposing
    Recipient string   // Address to receive funds
    Denom     string   // Asset to withdraw
    Amount    math.Int // Amount to withdraw
    Reason    string   // Justification for withdrawal
}
```

**Example**:
```bash
sharehodld tx bridge propose-withdrawal \
  sharehodl1abc... \
  ubtc \
  50000000 \
  "Investment in mining operations" \
  --from validator1 \
  --chain-id sharehodl-1
```

### MsgVoteWithdrawal
Vote on withdrawal proposal.

```go
type MsgVoteWithdrawal struct {
    Validator  string // Validator voting
    ProposalID uint64 // Proposal ID
    Vote       bool   // true = yes, false = no
}
```

**Example**:
```bash
sharehodld tx bridge vote-withdrawal 1 true \
  --from validator1 \
  --chain-id sharehodl-1
```

## Parameters

All parameters are governance-controllable:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `swap_fee_rate` | 0.1% | Fee charged on swaps |
| `min_reserve_ratio` | 20% | Minimum reserve:HODL ratio |
| `max_withdrawal_percent` | 10% | Max withdrawal per proposal |
| `withdrawal_voting_period` | 7 days | Voting period for withdrawals |
| `required_validator_approval` | 66.7% | Approval threshold |
| `bridge_enabled` | true | Master enable/disable switch |

## State

### KVStore Keys

| Prefix | Key | Value | Description |
|--------|-----|-------|-------------|
| `0x01` | - | `Params` | Module parameters |
| `0x02` | `denom` | `BridgeReserve` | Reserve balance per asset |
| `0x03` | `denom` | `SupportedAsset` | Supported asset config |
| `0x04` | `swapID` | `SwapRecord` | Swap transaction history |
| `0x05` | `proposalID` | `ReserveWithdrawal` | Withdrawal proposals |
| `0x06` | `proposalID + voter` | `WithdrawalVote` | Votes on proposals |
| `0x07` | - | `uint64` | Next swap ID counter |
| `0x08` | - | `uint64` | Next proposal ID counter |

## Events

### bridge_swap_in
Emitted when a user swaps external asset for HODL.

**Attributes**:
- `user`: User address
- `input_denom`: Asset swapped
- `input_amount`: Amount swapped
- `hodl_received`: HODL received (after fee)
- `fee`: Fee charged
- `swap_id`: Swap record ID

### bridge_swap_out
Emitted when a user swaps HODL for external asset.

**Attributes**:
- `user`: User address
- `hodl_amount`: HODL swapped
- `output_denom`: Asset received
- `output_received`: Amount received (after fee)
- `fee`: Fee charged
- `swap_id`: Swap record ID

### bridge_validator_deposit
Emitted when a validator deposits to reserve.

**Attributes**:
- `validator`: Validator address
- `denom`: Asset deposited
- `amount`: Amount deposited

### bridge_withdrawal_proposed
Emitted when a withdrawal is proposed.

**Attributes**:
- `proposal_id`: Proposal ID
- `proposer`: Validator proposing
- `recipient`: Recipient address
- `denom`: Asset to withdraw
- `amount`: Amount to withdraw
- `reason`: Withdrawal reason

### bridge_withdrawal_voted
Emitted when a validator votes.

**Attributes**:
- `proposal_id`: Proposal ID
- `validator`: Validator voting
- `vote`: Vote (true/false)
- `yes_votes`: Current yes votes
- `no_votes`: Current no votes
- `status`: Updated proposal status

### bridge_withdrawal_executed
Emitted when a withdrawal is executed.

**Attributes**:
- `proposal_id`: Proposal ID
- `recipient`: Recipient address
- `denom`: Asset withdrawn
- `amount`: Amount withdrawn

### bridge_withdrawal_expired
Emitted when a proposal expires.

**Attributes**:
- `proposal_id`: Proposal ID

## Security Considerations

1. **Reserve Ratio Enforcement**: All swaps and withdrawals must maintain minimum 20% reserve ratio
2. **Validator-Only Actions**: Only active validators can deposit or propose withdrawals
3. **Governance Approval**: 2/3 validator approval required for withdrawals
4. **Max Withdrawal Limit**: Maximum 10% of reserve per proposal prevents large unauthorized drains
5. **Atomic Operations**: All swaps are atomic - either complete fully or revert
6. **Fee Validation**: Swap fees prevent dust attacks and support protocol sustainability

## Integration

### App Integration

```go
// In app/app.go
import bridgekeeper "github.com/sharehodl/sharehodl-blockchain/x/bridge/keeper"
import bridgetypes "github.com/sharehodl/sharehodl-blockchain/x/bridge/types"

// Create keeper
app.BridgeKeeper = bridgekeeper.NewKeeper(
    appCodec,
    keys[bridgetypes.StoreKey],
    memKeys[bridgetypes.MemStoreKey],
    app.BankKeeper,
    app.ValidatorKeeper,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)

// Register module
app.ModuleManager = module.NewManager(
    // ... other modules
    bridge.NewAppModule(appCodec, app.BridgeKeeper),
)
```

### Genesis Integration

Add to `genesis.json`:

```json
{
  "bridge": {
    "params": {
      "swap_fee_rate": "0.001000000000000000",
      "min_reserve_ratio": "0.200000000000000000",
      "max_withdrawal_percent": "0.100000000000000000",
      "withdrawal_voting_period": 604800,
      "required_validator_approval": "0.667000000000000000",
      "bridge_enabled": true
    },
    "supported_assets": [],
    "reserves": [],
    "swap_records": [],
    "withdrawal_proposals": [],
    "next_swap_id": 1,
    "next_proposal_id": 1
  }
}
```

## Testing

Key test scenarios:
1. ✅ Swap in: user deposits USDT, receives HODL
2. ✅ Swap out: user deposits HODL, receives USDT
3. ✅ Fee calculation and deduction
4. ✅ Reserve ratio enforcement
5. ✅ Validator deposit (fee-free)
6. ✅ Withdrawal proposal creation
7. ✅ Withdrawal voting and approval
8. ✅ Withdrawal execution
9. ✅ Proposal expiration
10. ✅ Maximum withdrawal limit enforcement
11. ✅ Insufficient reserve rejection
12. ✅ Non-validator rejection

## Future Enhancements

- [ ] Oracle integration for dynamic price feeds
- [ ] Multi-signature withdrawal approvals
- [ ] Emergency pause mechanism
- [ ] Per-asset reserve ratios
- [ ] Tiered swap fees based on volume
- [ ] Cross-chain bridge integration
- [ ] Liquidity provider rewards
- [ ] Flash loan protection
