# Tiered Freeze Authority Implementation

## Overview
Implemented multi-tier voting system for fraud freeze authority to prevent single Keeper from immediately freezing company treasury.

## Problem Solved
**Old Behavior** (`x/escrow/keeper/report.go:224-240`):
- ANY Keeper submits fraud report → Company treasury frozen IMMEDIATELY
- Major vulnerability: Malicious/mistaken Keeper could halt legitimate companies

**New Behavior**:
- Keeper submits report → Status: "Preliminary Investigation" (NO freeze)
- Multi-tier voting required before freeze takes effect

## Implementation Flow

### Phase 1: Warden Review (48 hours)
- 3 Wardens (Tier 2: 10K HODL stake) vote
- 2/3 approval required to escalate to Steward review
- If rejected → Investigation cleared, no action

### Phase 2: Steward Review (72 hours)
- 5 Stewards (Tier 3: 50K HODL stake) vote
- 3/5 approval required to approve freeze
- If rejected → Investigation cleared, no action

### Phase 3: 24-Hour Warning
- If Stewards approve → Warning issued to company
- Company has 24 hours notice before freeze
- Company can prepare/respond

### Phase 4: Freeze Execution
- After 24-hour warning expires → Treasury frozen, trading halted
- Full investigation by Archons can proceed

## Files Modified

### 1. `/x/escrow/types/report.go`
**Added:**
- `InvestigationStatus` enum (6 statuses)
- `TierVote` struct for tier-specific voting
- `CompanyInvestigation` struct to track multi-tier voting process
- Helper methods: `CountWardenVotes()`, `CountStewardVotes()`, `HasVoted()`

### 2. `/x/escrow/types/keys.go`
**Added:**
- `CompanyInvestigationPrefix` - Main storage prefix
- `CompanyInvestigationCounterKey` - ID counter
- `CompanyInvestigationByCompanyPrefix` - Index by company
- `CompanyInvestigationByStatusPrefix` - Index by status
- Key helper functions

### 3. `/x/escrow/keeper/company_investigation.go` (NEW FILE)
**Implements:**
- `CreateCompanyInvestigation()` - Start investigation (no freeze)
- `VoteOnInvestigation()` - Warden/Steward voting
- `ProcessInvestigationPhases()` - EndBlock phase transitions
- `GetInvestigationsByStatus()` - Query helpers
- `GetInvestigationsByCompany()` - Query helpers

### 4. `/x/escrow/keeper/report.go`
**Modified:**
- Lines 223-242: Changed from immediate `InitiateFraudInvestigation()` to `CreateCompanyInvestigation()`
- Removed auto-freeze vulnerability
- Added tiered investigation flow

## Data Structures

### InvestigationStatus
```go
const (
    InvestigationStatusPreliminary    // Report filed, no action yet
    InvestigationStatusWardenReview   // Awaiting Warden review (48hr)
    InvestigationStatusStewardReview  // Escalated to Steward review (72hr)
    InvestigationStatusFreezeApproved // Freeze approved, pending 24hr warning
    InvestigationStatusFrozen         // Treasury frozen, trading halted
    InvestigationStatusCleared        // Investigation cleared, no fraud
)
```

### TierVote
```go
type TierVote struct {
    Voter     string    // Validator address
    Tier      int       // Validator tier (Warden=2, Steward=3, Archon=4)
    Approve   bool      // true = escalate/freeze, false = clear
    Reason    string    // Voting rationale
    VotedAt   time.Time
}
```

### CompanyInvestigation
```go
type CompanyInvestigation struct {
    ID        uint64
    CompanyID uint64
    ReportID  uint64
    Status    InvestigationStatus

    // Warden Review Phase
    WardenVotes    []TierVote
    WardenDeadline time.Time
    WardenApproved bool

    // Steward Review Phase
    StewardVotes    []TierVote
    StewardDeadline time.Time
    StewardApproved bool

    // Warning Phase
    WarningIssuedAt  time.Time
    WarningExpiresAt time.Time
    CompanyNotified  bool

    // Timeline
    CreatedAt  time.Time
    FrozenAt   time.Time
    ResolvedAt time.Time
}
```

## Keeper Methods

### CreateCompanyInvestigation
```go
func (k Keeper) CreateCompanyInvestigation(ctx, companyID, reportID uint64) (uint64, error)
```
- Creates investigation with Preliminary status
- Auto-transitions to Warden Review
- Sets 48-hour deadline for Warden votes
- NO freeze occurs

### VoteOnInvestigation
```go
func (k Keeper) VoteOnInvestigation(ctx, investigationID uint64, voter string, approve bool, reason string) error
```
- Validates voter tier (Warden for phase 1, Steward for phase 2)
- Records vote
- Checks if quorum reached
- Auto-transitions phases based on votes

### ProcessInvestigationPhases (EndBlock)
```go
func (k Keeper) ProcessInvestigationPhases(ctx)
```
- Processes expired Warden review deadlines
- Processes expired Steward review deadlines
- Executes freeze after 24-hour warning expires
- Calls `equityKeeper.InitiateFraudInvestigation()` only when approved

## Events Emitted

### company_investigation_created
- investigation_id
- company_id
- report_id
- status
- warden_deadline

### investigation_vote
- investigation_id
- voter
- tier
- approve
- phase

### investigation_freeze_warning
- investigation_id
- company_id
- warning_expires_at

### company_frozen
- investigation_id
- company_id
- report_id
- frozen_at

## Voting Requirements

### Warden Review
- **Quorum**: 3 Wardens must vote
- **Approval**: 2/3 (2 out of 3) to escalate
- **Timeout**: 48 hours
- **On Rejection**: Investigation cleared

### Steward Review
- **Quorum**: 5 Stewards must vote
- **Approval**: 3/5 (3 out of 5) to approve freeze
- **Timeout**: 72 hours
- **On Rejection**: Investigation cleared

### Warning Period
- **Duration**: 24 hours after Steward approval
- **Purpose**: Give company notice before freeze
- **After Expiry**: Freeze executed automatically

## Security Improvements

### Before
- 1 Keeper (10K HODL) → Immediate freeze
- No review process
- No warning period
- High risk of abuse/mistakes

### After
- 2/3 Wardens (10K HODL each) → Escalate to Steward review
- 3/5 Stewards (50K HODL each) → Approve freeze
- 24-hour warning period
- Multi-tier consensus required
- ~150K-250K HODL at stake across voters

## Total Timeline

### Minimum (if all votes happen immediately)
- Warden review: ~1 hour (if 3 votes come quickly)
- Steward review: ~1 hour (if 5 votes come quickly)
- Warning period: 24 hours (mandatory)
- **Total: ~26 hours minimum**

### Maximum (if deadlines reached)
- Warden review: 48 hours
- Steward review: 72 hours
- Warning period: 24 hours
- **Total: 144 hours (6 days) maximum**

### Typical (some votes, some timeouts)
- Warden review: 12-24 hours
- Steward review: 24-48 hours
- Warning period: 24 hours
- **Total: ~2-4 days typical**

## Integration Points

### Equity Keeper
- `InitiateFraudInvestigation(ctx, companyID, reportID, initiator)` now only called:
  - After Steward approval (3/5 votes)
  - After 24-hour warning expires
  - From `ProcessInvestigationPhases()` EndBlock

### Staking Keeper
- `GetUserTierInt(ctx, address)` used to validate voter tier
- Ensures Wardens are tier 2+ (10K HODL)
- Ensures Stewards are tier 3+ (50K HODL)

## Testing Checklist

- [ ] Create investigation without freeze
- [ ] Warden voting (2/3 approval → escalate)
- [ ] Warden voting (< 2/3 → clear)
- [ ] Steward voting (3/5 approval → warning)
- [ ] Steward voting (< 3/5 → clear)
- [ ] 24-hour warning period
- [ ] Freeze execution after warning
- [ ] Deadline expiration handling
- [ ] Double-vote prevention
- [ ] Tier validation
- [ ] Query by status
- [ ] Query by company

## Future Enhancements

1. **Archon Override**: Add ability for 7/9 Archons to immediately freeze in emergency
2. **Vote Delegation**: Allow Wardens/Stewards to delegate votes
3. **Evidence Attachment**: Link investigation to report evidence
4. **Notification System**: On-chain notifications for companies under investigation
5. **Appeal Process**: Allow companies to appeal to Archons if frozen
6. **Metrics Tracking**: Track investigation outcomes for voter accountability

## Migration Notes

**No migration required** - This is a new system that runs in parallel with existing report system.

Existing fraud reports will continue to work, but new company fraud/scam reports will use the tiered investigation system.

## Deployment Steps

1. Deploy updated keeper code
2. Initialize investigation counter to 0
3. Monitor first investigations
4. Verify EndBlock processing
5. Verify freeze execution after approval

## Monitoring

Watch for these events in production:
- `company_investigation_created` - Track new investigations
- `investigation_vote` - Track voting activity
- `investigation_freeze_warning` - Alert on imminent freezes
- `company_frozen` - Alert on executed freezes

## Related Files

- `/x/escrow/types/report.go` - Type definitions
- `/x/escrow/types/keys.go` - Storage keys
- `/x/escrow/keeper/company_investigation.go` - Core implementation
- `/x/escrow/keeper/report.go` - Integration with report system
- `/x/equity/keeper/delisting.go` - Freeze execution

## Summary

This implementation successfully moves from a single-point-of-failure freeze authority to a multi-tier consensus system with proper safeguards:

1. **Decentralized Authority**: 2 Warden + 3 Steward approval required
2. **Economic Stake**: ~150K-250K HODL at risk across voters
3. **Warning Period**: 24-hour notice before freeze
4. **Reversible Process**: Can be cleared at any phase
5. **Transparent**: All votes and status changes tracked on-chain

The company treasury is now protected from hasty or malicious freeze actions while still allowing legitimate fraud investigations to proceed through a proper review process.
