# Freeze Warning System Implementation

## Overview

The Freeze Warning System provides a 24-hour advance notice before freezing a company's treasury during fraud investigations. This prevents griefing and gives legitimate companies a chance to respond with counter-evidence.

## Problem Statement

**Before**: When fraud was detected, the company's trading was halted and treasury frozen immediately, with no warning. This allowed for potential abuse:
- Griefing attacks where false reports could lock company funds instantly
- No opportunity for companies to defend themselves
- Legitimate companies could be harmed by malicious or mistaken reports

**After**: A structured warning system with:
- 24-hour warning period before treasury freeze
- Company can submit counter-evidence during warning period
- Escalation to Archon review if company responds
- Automatic freeze execution if no response

## Architecture

### Data Structures

#### FreezeWarning
```go
type FreezeWarning struct {
    ID                uint64              // Unique warning ID
    CompanyID         uint64              // Target company
    InvestigationID   uint64              // Links to investigation
    ReportID          uint64              // Original fraud report

    // Warning timeline
    Status            FreezeWarningStatus
    WarningIssuedAt   time.Time
    WarningExpiresAt  time.Time          // 24 hours after issued

    // Company response
    CompanyResponded      bool
    ResponseSubmittedAt   time.Time
    CounterEvidence       []Evidence
    ResponseText          string

    // Outcome
    FreezeExecuted    bool
    EscalatedToArchon bool
    ResolvedAt        time.Time
    Resolution        string
}
```

#### FreezeWarningStatus
```go
const (
    FreezeWarningStatusPending   // Warning active, awaiting response
    FreezeWarningStatusResponded // Company submitted counter-evidence
    FreezeWarningStatusExecuted  // Freeze executed (no response)
    FreezeWarningStatusEscalated // Escalated to Archon review
    FreezeWarningStatusCleared   // Cleared by Archons
)
```

### Storage Keys

```go
FreezeWarningPrefix        = []byte{0x3B}  // Store freeze warnings
FreezeWarningCounterKey    = []byte{0x3C}  // Counter for warning IDs
FreezeWarningByCompanyKey  = []byte{0x3D}  // Index: company ID -> warning ID
```

## Flow Diagrams

### Flow 1: Warning Issued, No Company Response

```
┌─────────────────────┐
│ Fraud Report        │
│ Confirmed by        │
│ Stewards            │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ IssueFreezeWarning  │
│ (24h timer starts)  │
└──────────┬──────────┘
           │
           │ Event: freeze_warning_issued
           │ → Company should monitor this
           │
           ▼
     Wait 24 hours
           │
           ▼
┌─────────────────────┐
│ EndBlock processes  │
│ expired warnings    │
└──────────┬──────────┘
           │
           │ No company response detected
           │
           ▼
┌─────────────────────┐
│ ExecuteFreezeAfter  │
│ Warning             │
│ - Halt trading      │
│ - Freeze treasury   │
└──────────┬──────────┘
           │
           ▼
    Company frozen
```

### Flow 2: Company Responds with Counter-Evidence

```
┌─────────────────────┐
│ Fraud Report        │
│ Confirmed by        │
│ Stewards            │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ IssueFreezeWarning  │
│ (24h timer starts)  │
└──────────┬──────────┘
           │
           │ Within 24 hours
           │
           ▼
┌─────────────────────┐
│ Company Founder     │
│ Responds with       │
│ Counter-Evidence    │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ RespondToFreeze     │
│ Warning             │
│ - Save evidence     │
│ - Mark as responded │
└──────────┬──────────┘
           │
           │ 24 hours pass
           │
           ▼
┌─────────────────────┐
│ EndBlock processes  │
│ expired warnings    │
└──────────┬──────────┘
           │
           │ Company responded detected
           │
           ▼
┌─────────────────────┐
│ EscalateToArchon    │
│ Review              │
│ - Archons evaluate  │
│ - Counter-evidence  │
└──────────┬──────────┘
           │
           ├──────────────┬──────────────┐
           │              │              │
           ▼              ▼              ▼
    Valid Evidence   Invalid       Needs More
    → Clear          → Freeze      Investigation
```

## API Reference

### Core Functions

#### IssueFreezeWarning
Issues a 24-hour warning before freeze. Called by Stewards after reviewing fraud evidence.

```go
func (k Keeper) IssueFreezeWarning(
    ctx sdk.Context,
    companyID uint64,
    investigationID uint64,
    reportID uint64,
) (uint64, error)
```

**Returns**: Warning ID
**Emits**: `freeze_warning_issued` event

**Example Event**:
```json
{
    "type": "freeze_warning_issued",
    "attributes": [
        {"key": "company_id", "value": "100"},
        {"key": "company_symbol", "value": "ACME"},
        {"key": "warning_id", "value": "1"},
        {"key": "report_id", "value": "500"},
        {"key": "warning_expires_at", "value": "2026-01-31T10:00:00Z"},
        {"key": "founder", "value": "sharehodl1founder..."}
    ]
}
```

#### RespondToFreezeWarning
Allows company founder to submit counter-evidence and response.

```go
func (k Keeper) RespondToFreezeWarning(
    ctx sdk.Context,
    warningID uint64,
    responder string,
    evidence []Evidence,
    response string,
) error
```

**Requirements**:
- Responder must be company founder
- Warning must be in Pending status
- Warning must not have expired
- Company cannot respond twice

**Emits**: `freeze_warning_response` event

#### ProcessExpiredFreezeWarnings
Processes warnings that have expired. Called from EndBlock.

```go
func (k Keeper) ProcessExpiredFreezeWarnings(ctx sdk.Context)
```

**Logic**:
- For each pending warning:
  - If expired AND no response → Execute freeze
  - If expired AND has response → Escalate to Archon

#### ExecuteFreezeAfterWarning
Executes the actual freeze after warning period expires without response.

```go
func (k Keeper) ExecuteFreezeAfterWarning(
    ctx sdk.Context,
    warningID uint64,
) error
```

**Actions**:
1. Halt trading (if not already halted)
2. Freeze company treasury
3. Update warning status to Executed
4. Emit `freeze_executed` event

#### EscalateToArchonReview
Escalates a warning with company response to Archon-tier review.

```go
func (k Keeper) EscalateToArchonReview(
    ctx sdk.Context,
    warningID uint64,
) error
```

**Actions**:
1. Update warning status to Escalated
2. Emit `freeze_escalated` event
3. Archons review counter-evidence manually

#### ClearFreezeWarning
Clears a freeze warning after Archon review. Called by Archons if counter-evidence is valid.

```go
func (k Keeper) ClearFreezeWarning(
    ctx sdk.Context,
    warningID uint64,
    clearedBy string,
    reason string,
) error
```

**Actions**:
1. Update warning status to Cleared
2. Resume trading if halted
3. Emit `freeze_warning_cleared` event

## Events

### freeze_warning_issued
Emitted when a freeze warning is issued to a company.

```
company_id: uint64
company_symbol: string
warning_id: uint64
report_id: uint64
warning_expires_at: timestamp
founder: address
```

**UI Integration**: Frontend should display prominent warning to company founder.

### freeze_warning_response
Emitted when company responds with counter-evidence.

```
warning_id: uint64
company_id: uint64
company_symbol: string
responder: address
evidence_count: uint64
```

### freeze_executed
Emitted when freeze is executed after warning expires.

```
warning_id: uint64
company_id: uint64
company_symbol: string
frozen_amount: uint64
```

### freeze_escalated
Emitted when warning with response is escalated to Archon review.

```
warning_id: uint64
company_id: uint64
company_symbol: string
evidence_count: uint64
response_text: string
```

### freeze_warning_cleared
Emitted when Archon clears a warning after reviewing counter-evidence.

```
warning_id: uint64
company_id: uint64
company_symbol: string
cleared_by: address
reason: string
```

## Integration Points

### Escrow Report System
When a fraud/scam report is submitted against a company:

1. `InitiateFraudInvestigation` is called immediately
   - Trading is halted as a precaution
   - Company status → Suspended
   - **Treasury is NOT frozen yet**

2. Stewards review the evidence
   - If valid → Issue freeze warning via `IssueFreezeWarning`
   - If invalid → Resume trading, restore company

### Governance Module
Archon-tier reviewers need access to:
- View freeze warnings with status Escalated
- Review company counter-evidence
- Call `ClearFreezeWarning` or approve freeze execution

### Frontend Applications

#### Business Dashboard (apps/business)
Must monitor events for the company's founder address:
```typescript
// Subscribe to freeze warning events
wsClient.subscribeEvent('freeze_warning_issued', (event) => {
  if (event.founder === currentUserAddress) {
    // Display urgent alert
    showFreezeWarningAlert({
      companyId: event.company_id,
      warningId: event.warning_id,
      expiresAt: event.warning_expires_at,
      reportId: event.report_id,
    })
  }
})
```

#### Warning Response UI
```typescript
interface RespondToWarningForm {
  warningId: number
  counterEvidence: {
    hash: string          // IPFS/Arweave hash
    description: string
    submitter: string
  }[]
  responseText: string     // Max 5000 chars
}

async function submitResponse(form: RespondToWarningForm) {
  const msg = {
    type: 'sharehodl.equity.MsgRespondToFreezeWarning',
    warningId: form.warningId,
    responder: userAddress,
    evidence: form.counterEvidence,
    response: form.responseText,
  }

  await wallet.signAndBroadcast([msg])
}
```

## Testing

### Unit Tests
See `freeze_warning_test.go` for comprehensive unit tests:
- Validation tests
- Expiration logic tests
- Status transitions

### Integration Test Scenarios

#### Scenario 1: Normal Freeze Execution
```go
// 1. Issue warning
warningID, _ := keeper.IssueFreezeWarning(ctx, companyID, invID, reportID)

// 2. Wait 24 hours (simulate)
ctx = ctx.WithBlockTime(ctx.BlockTime().Add(25 * time.Hour))

// 3. Process in EndBlock
keeper.ProcessExpiredFreezeWarnings(ctx)

// 4. Verify freeze executed
warning, _ := keeper.GetFreezeWarning(ctx, warningID)
require.Equal(t, types.FreezeWarningStatusExecuted, warning.Status)
require.True(t, keeper.IsTradingHalted(ctx, companyID))
```

#### Scenario 2: Company Responds, Escalates
```go
// 1. Issue warning
warningID, _ := keeper.IssueFreezeWarning(ctx, companyID, invID, reportID)

// 2. Company responds within 24h
evidence := []types.Evidence{{...}}
keeper.RespondToFreezeWarning(ctx, warningID, founder, evidence, "We are legitimate")

// 3. Wait 24 hours
ctx = ctx.WithBlockTime(ctx.BlockTime().Add(25 * time.Hour))

// 4. Process in EndBlock
keeper.ProcessExpiredFreezeWarnings(ctx)

// 5. Verify escalated
warning, _ := keeper.GetFreezeWarning(ctx, warningID)
require.Equal(t, types.FreezeWarningStatusEscalated, warning.Status)
require.True(t, warning.EscalatedToArchon)
```

#### Scenario 3: Archon Clears Warning
```go
// ... (after escalation from Scenario 2)

// Archon reviews and clears
keeper.ClearFreezeWarning(ctx, warningID, archonAddr, "Evidence validated")

// Verify cleared
warning, _ := keeper.GetFreezeWarning(ctx, warningID)
require.Equal(t, types.FreezeWarningStatusCleared, warning.Status)
require.False(t, keeper.IsTradingHalted(ctx, companyID))
```

## Security Considerations

### Griefing Prevention
1. **Stake Requirements**: Only Keeper+ tier (10K HODL staked) can submit fraud reports
2. **Stake Age**: Must be staked for 7+ days to prevent Sybil attacks
3. **Reporter History**: Tracked to penalize false reporters
4. **Warning Period**: 24 hours gives company time to respond

### Access Control
- Only company founder can respond to freeze warning
- Only Archon-tier reviewers can clear warnings
- Only Stewards can issue freeze warnings

### Edge Cases Handled
1. **Company already has active warning**: Reject new warning
2. **Warning expires while processing**: Status checks prevent double-execution
3. **Company responds after expiration**: Rejected with error
4. **Multiple responses**: Only first response accepted

## Performance

### Storage Impact
- 1 warning per company maximum (active warnings)
- Historical warnings kept for audit trail
- Index by company ID for fast lookup

### EndBlock Processing
```go
// Worst case: All warnings expire in same block
warnings := k.GetAllPendingFreezeWarnings(ctx)  // O(n) where n = active warnings
for _, warning := range warnings {               // Typically < 10 per block
    if warning.IsExpired(ctx.BlockTime()) {
        k.handleExpiredWarning(ctx, warning)     // O(1)
    }
}
```

**Typical Load**: 0-5 warnings processed per block
**Max Load**: ~100 warnings (unlikely in practice)

## Future Enhancements

### Version 2.0 Features
1. **Configurable Warning Period**: Allow governance to adjust from 24h
2. **Partial Evidence Acceptance**: Archons can request additional evidence
3. **Appeal Process**: Company can appeal Archon decision
4. **Automatic Evidence Verification**: IPFS content verification
5. **Multi-signature Response**: Require multiple company officers to respond

### Governance Integration
```go
// Future: Create governance proposal for Archon review
type ArchonReviewProposal struct {
    WarningID       uint64
    CompanyEvidence []Evidence
    ReviewDeadline  time.Time
    MinArchonVotes  uint64    // e.g., 3 of 5 Archons
}
```

## Troubleshooting

### Warning Not Issued
**Problem**: `IssueFreezeWarning` returns error
**Checks**:
1. Company exists? `keeper.GetCompany(ctx, companyID)`
2. Already has warning? `keeper.GetFreezeWarningByCompany(ctx, companyID)`
3. Report exists? Verify reportID is valid

### Company Cannot Respond
**Problem**: `RespondToFreezeWarning` returns error
**Checks**:
1. Warning still pending? Check status
2. Not expired? Check `warning.WarningExpiresAt`
3. Responder is founder? Verify against `company.Founder`
4. Already responded? Check `warning.CompanyResponded`

### Freeze Not Executing
**Problem**: Warning expired but freeze didn't happen
**Checks**:
1. EndBlock running? Check module EndBlock called
2. Warning actually expired? Check timestamps
3. Status correct? Should be Pending or Responded

## References

### Related Files
- `/x/equity/types/delisting.go` - FreezeWarning types
- `/x/equity/types/key.go` - Storage keys
- `/x/equity/keeper/freeze_warning.go` - Implementation
- `/x/equity/keeper/freeze_warning_test.go` - Tests
- `/x/equity/keeper/delisting.go` - Integration with fraud system
- `/x/equity/module.go` - EndBlock registration

### Related Systems
- Escrow Report System (`x/escrow/keeper/report.go`)
- Company Investigation Flow
- Governance Archon Tier
- Staking Tier System

## Changelog

### v1.0.0 (2026-01-30)
- Initial implementation
- 24-hour warning period
- Company response capability
- Archon escalation flow
- EndBlock automation
