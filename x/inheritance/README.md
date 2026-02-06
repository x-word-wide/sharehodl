# Inheritance Module (Dead Man Switch)

The inheritance module enables ShareHODL users to create inheritance plans that automatically transfer their assets to designated beneficiaries after a period of inactivity.

## Features

- **Dead Man Switch**: Automatically triggers after owner inactivity
- **Grace Period**: Minimum 30-day warning period before asset transfers
- **Priority-Based Distribution**: Beneficiaries claim in priority order
- **Percentage & Specific Allocations**: Flexible asset distribution
- **Cascading Claims**: Skips banned/invalid beneficiaries automatically
- **False Positive Protection**: Any owner activity cancels triggers
- **Ultra-Long Inactivity**: After 50 years, assets go to charity
- **Ban Integration**: Banned owners cannot use the system

## Key Concepts

### Inheritance Plan
A plan defines:
- Owner (plan creator)
- Beneficiaries (with priority and percentage)
- Inactivity period (how long before trigger)
- Grace period (warning period, min 30 days)
- Claim window (time each beneficiary has to claim)
- Charity address (for unclaimed assets)

### Plan Lifecycle
1. **Active** - Monitoring for owner inactivity
2. **Triggered** - Dead man switch activated, grace period started
3. **Executing** - Grace period expired, claims being processed
4. **Completed** - All assets transferred
5. **Cancelled** - Cancelled by owner

### Activity Tracking
The module tracks all owner transactions. Any transaction automatically cancels triggered switches (false positive protection).

### Beneficiary Claims
After grace period:
1. Claims open in priority order (1, 2, 3...)
2. Each beneficiary gets a claim window
3. If beneficiary is banned → skip to next priority
4. If claim window expires → cascade to next priority
5. Assets transferred (specific allocations + percentage)

## Messages

### MsgCreatePlan
Create a new inheritance plan.

```go
{
  "creator": "sharehodl1...",
  "beneficiaries": [
    {
      "address": "sharehodl1...",
      "priority": 1,
      "percentage": "0.5", // 50%
      "specific_assets": []
    }
  ],
  "inactivity_period": "8760h", // 365 days
  "grace_period": "720h", // 30 days
  "claim_window": "2160h", // 90 days
  "charity_address": "sharehodl1..."
}
```

### MsgUpdatePlan
Update an existing plan (only when active).

### MsgCancelPlan
Cancel a plan permanently.

### MsgClaimAssets
Beneficiary claims their inheritance (during claim window).

### MsgCancelTrigger
Owner manually cancels a triggered switch.

## Queries

- `Params` - Get module parameters
- `Plan` - Get plan by ID
- `PlansByOwner` - Get all plans for an owner
- `PlansByBeneficiary` - Get all plans where address is beneficiary
- `Trigger` - Get active trigger for a plan
- `Claim` - Get claim status for a beneficiary
- `AllClaims` - Get all claims for a plan
- `LastActivity` - Get last activity for an address

## Parameters

- `min_inactivity_period`: Minimum time before switch can trigger (default: 180 days)
- `min_grace_period`: Minimum grace period (default: 30 days, enforced)
- `min_claim_window`: Minimum claim window (default: 30 days)
- `max_claim_window`: Maximum claim window (default: 365 days)
- `ultra_long_inactivity_period`: Time before charity transfer (default: 50 years)
- `default_charity_address`: Fallback charity address
- `max_beneficiaries`: Maximum beneficiaries per plan (default: 10)

## Events

All state changes emit events for indexing:
- `inheritance_plan_created`
- `inheritance_plan_updated`
- `inheritance_switch_triggered`
- `inheritance_switch_cancelled`
- `inheritance_assets_claimed`
- `inheritance_beneficiary_skipped`
- And more...

## Asset Types

Supports:
1. **HODL** - Native stablecoin
2. **Equity** - Company shares
3. **Custom Tokens** - Any token on the chain

## Security

- **Ban Check**: Banned owners cannot trigger switches
- **Activity Recording**: All transactions record activity
- **Beneficiary Validation**: Ban status checked before every transfer
- **Atomic Transfers**: All-or-nothing asset transfers
- **Time Locks**: Grace periods and claim windows enforced
- **Owner Control**: Can cancel at any time before grace period expires

## Example Workflow

1. Alice creates plan with Bob (priority 1) and Charlie (priority 2)
2. Alice becomes inactive for 365 days
3. Switch triggers, 30-day grace period starts
4. Grace period expires, Bob's claim window opens (90 days)
5. Bob claims → receives 50% of Alice's assets
6. Bob's window expires, Charlie's window opens
7. Charlie claims → receives remaining assets
8. Plan marked as completed

## False Positive Protection

If Alice signs ANY transaction during grace period or claim windows:
- All active triggers automatically cancelled
- Plan reverts to "Active" status
- Process resets completely

## CLI Examples

Coming soon after protobuf generation.

## Integration

See `IMPLEMENTATION.md` for integration details.
