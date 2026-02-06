# Report Escalation System - Implementation Complete

## Overview
Implemented a comprehensive escalation system for fraud reports to **prevent false negatives** where real fraud could slip through due to insufficient votes or reviewer unavailability.

## Problem Solved
**Old behavior (DANGEROUS):**
```go
// Auto-dismissed if not enough votes - real fraud could slip through!
if len(report.ReviewVotes) < report.VotesRequired {
    report.Status = types.ReportStatusDismissed
    report.Resolution = "Auto-dismissed: deadline passed without sufficient votes"
}
```

**New behavior (SAFE):**
- If ANY reviewer saw merit (confirmed vote), ESCALATE to higher tier
- If ALL votes were "dismiss", then safe to auto-dismiss
- If NO votes at all, extend deadline and reassign reviewers
- Track stale reports that keep failing to get votes

## Files Modified

### 1. `/x/escrow/types/report.go`
Added escalation tracking fields to Report struct:
```go
// Escalation tracking (prevents false negatives)
EscalationCount   int          `json:"escalation_count"`              // How many times escalated
ExtensionCount    int          `json:"extension_count"`               // How many times deadline extended
OriginalTier      int          `json:"original_tier"`                 // Starting reviewer tier
CurrentTier       int          `json:"current_tier"`                  // Current reviewer tier
EscalatedAt       time.Time    `json:"escalated_at,omitempty"`        // When last escalated
PreviousVotes     []ReviewVote `json:"previous_votes,omitempty"`      // Votes from lower tiers
```

Added constants:
```go
const (
    MaxReportExtensions  = 2                       // Maximum deadline extensions before escalation
    MaxReportEscalations = 3                       // Maximum tier escalations before governance
    ExtensionDuration    = 3 * 24 * time.Hour      // 3 days per extension
)
```

### 2. `/x/escrow/types/errors.go`
Added new event types:
```go
// Report escalation event types
EventTypeReportEscalated           = "report_escalated"
EventTypeReportDeadlineExtended    = "report_deadline_extended"
EventTypeReportStale               = "report_stale"
EventTypeReportEscalatedToGovernance = "report_escalated_to_governance"
```

### 3. `/x/escrow/keeper/report.go`
Completely rewrote `ProcessExpiredReports()` with smart escalation logic:

```go
func (k Keeper) ProcessExpiredReports(ctx sdk.Context) {
    reports := k.GetReportsByStatus(ctx, types.ReportStatusUnderInvestigation)

    for _, report := range reports {
        if ctx.BlockTime().After(report.DeadlineAt) {
            confirmVotes := 0
            dismissVotes := 0

            for _, vote := range report.ReviewVotes {
                if vote.Confirmed {
                    confirmVotes++
                } else {
                    dismissVotes++
                }
            }

            if len(report.ReviewVotes) == 0 {
                // No votes at all - extend deadline and reassign
                k.extendReportDeadline(ctx, &report)
            } else if confirmVotes > 0 {
                // At least one reviewer saw merit - ESCALATE
                k.escalateReportToHigherTier(ctx, &report)
            } else {
                // All votes were dismiss - safe to auto-dismiss
                k.autoDismissReport(ctx, &report)
            }
        }
    }
}
```

Added three new methods:

#### `extendReportDeadline()`
- Extends deadline by 3 days (ExtensionDuration)
- Reassigns different reviewers (current ones didn't respond)
- Max 2 extensions before escalating
- Tracks extension count
- Emits `EventTypeReportDeadlineExtended`

#### `escalateReportToHigherTier()`
- Saves current votes to `PreviousVotes` history
- Escalates through tiers: Warden (2) → Steward (3) → Archon (4) → Governance
- Clears old votes and assigns new higher-tier reviewers
- Extends deadline for fresh review period
- Tracks escalation count
- Emits `EventTypeReportEscalated`
- Max 3 escalations before governance

#### `escalateReportToGovernance()`
- Final escalation level when even Archons can't reach consensus
- Changes status to `ReportStatusAppealed` (reusing for governance)
- Requires community governance vote
- Emits `EventTypeReportEscalatedToGovernance`

#### `autoDismissReport()` (NEW - extracted from old logic)
- Only called when ALL reviewers voted "dismiss"
- Applies penalties to reporter (existing penalty system)
- Logs all reviewer dismissal reasons
- Emits detailed resolution event

## Escalation Flow

### Scenario 1: No Votes (Reviewers Busy/Unavailable)
```
Report expires with 0 votes
  ↓
extendReportDeadline() - Extension 1 (3 days)
  ↓
Still 0 votes
  ↓
extendReportDeadline() - Extension 2 (3 days)
  ↓
Still 0 votes (max extensions reached)
  ↓
escalateReportToHigherTier() - Move to higher tier
```

### Scenario 2: Mixed Votes (At Least 1 "Confirmed")
```
Report expires with votes: 1 confirm, 2 dismiss
  ↓
escalateReportToHigherTier() - Warden → Steward
  ↓
Still mixed votes
  ↓
escalateReportToHigherTier() - Steward → Archon
  ↓
Still mixed votes
  ↓
escalateReportToHigherTier() - Archon → Governance
  ↓
escalateReportToGovernance() - Community vote required
```

### Scenario 3: All Dismiss (Safe to Auto-Dismiss)
```
Report expires with votes: 0 confirm, 5 dismiss
  ↓
autoDismissReport() - All reviewers agreed: no fraud
  ↓
Apply penalties to reporter for false report
```

## Benefits

### 1. **Prevents False Negatives**
Real fraud cannot slip through just because reviewers were busy or unavailable. If even ONE reviewer saw merit, the report escalates for more scrutiny.

### 2. **Intelligent Resource Allocation**
- Simple cases get handled quickly by lower tiers
- Complex/controversial cases automatically escalate to more experienced reviewers
- Community governance as final arbiter

### 3. **Full Audit Trail**
- `PreviousVotes` tracks all votes from all tiers
- `ActionsTaken` logs every extension and escalation
- Events emitted for transparency

### 4. **No False Auto-Dismissals**
Only auto-dismiss when ALL reviewers unanimously agreed the report was invalid. Any disagreement triggers escalation.

### 5. **Handles Edge Cases**
- Stale reports (no reviewers available) get marked
- Max escalations prevent infinite loops
- Governance as backstop for truly difficult cases

## Testing Scenarios

### Test Case 1: Subtle Fraud (Needs Expert Review)
1. User reports fraud
2. Warden tier reviewers split 2-2 (some see it, some don't)
3. Auto-escalates to Steward tier
4. Stewards confirm fraud 4-1
5. Report confirmed, actions executed

**Result:** Fraud caught despite initial disagreement

### Test Case 2: Busy Reviewers
1. User reports misconduct
2. No reviewers vote in 7 days (all busy)
3. Deadline extended +3 days, new reviewers assigned
4. Still no votes
5. Deadline extended again +3 days
6. Still no votes (max extensions)
7. Auto-escalates to higher tier
8. Higher tier reviews and resolves

**Result:** Report gets attention despite busy lower tier

### Test Case 3: False Report (Unanimous Dismiss)
1. User files bogus fraud report
2. All 5 Warden reviewers vote "dismiss"
3. Auto-dismissed with penalties to reporter
4. No escalation needed

**Result:** Fast dismissal when clearly invalid

### Test Case 4: Controversial Case
1. Complex fraud report
2. Wardens escalate to Stewards (split votes)
3. Stewards escalate to Archons (still split)
4. Archons escalate to Governance (max escalations)
5. Community votes via governance proposal

**Result:** Difficult cases get full community input

## Constants Summary

| Constant | Value | Purpose |
|----------|-------|---------|
| `MaxReportExtensions` | 2 | Max deadline extensions before escalation |
| `MaxReportEscalations` | 3 | Max tier escalations before governance |
| `ExtensionDuration` | 3 days | Time added per extension |

## Tier Escalation Path

```
Priority 1-3 (Standard):
Warden (2) → Steward (3) → Archon (4) → Governance

Priority 4 (High):
Steward (3) → Archon (4) → Governance

Priority 5 (Emergency):
Archon (4) → Governance
```

## Event Tracking

All escalations emit detailed events for monitoring:

```go
// Extension
EventTypeReportDeadlineExtended
  - report_id
  - extension_count
  - old_deadline
  - new_deadline

// Escalation
EventTypeReportEscalated
  - report_id
  - escalation_count
  - old_tier
  - new_tier
  - tier_name
  - previous_votes_count

// Governance
EventTypeReportEscalatedToGovernance
  - report_id
  - escalation_count
  - original_tier
  - report_type
  - target_type
  - target_id
```

## Integration Notes

### EndBlock
`ProcessExpiredReports()` is called from EndBlock to check for expired reports and apply escalation logic.

### Staking Keeper Integration
Uses existing `k.stakingKeeper.GetUserTierInt()` to determine reviewer tiers.

### Reporter Penalty System
Integrates with existing `ApplyReporterPenalty()` and `penalizeReporter()` for false reports.

### Moderator Assignment
Reuses existing `assignReviewers()` to find eligible reviewers at each tier.

## Security Considerations

1. **Max Escalations**: Prevents infinite escalation loops
2. **Max Extensions**: Limits deadline extensions to prevent DoS
3. **Vote History**: `PreviousVotes` provides full audit trail
4. **Governance Backstop**: Community can resolve truly controversial cases
5. **Unanimous Dismiss**: Only auto-dismiss with full consensus

## Future Enhancements

1. **Priority Boost**: Escalated reports could get priority in reviewer queues
2. **Reputation Weights**: Higher-tier votes could carry more weight
3. **Appeal System**: Allow reporter to appeal auto-dismissals
4. **Time-Based Urgency**: Older escalations get higher priority
5. **Reviewer Incentives**: Bonus rewards for handling escalated reports

## Conclusion

The escalation system ensures that **no legitimate fraud report slips through the cracks** while still efficiently handling clear false reports. It balances:

- **Safety**: Real fraud gets escalated for expert review
- **Efficiency**: Clear cases resolved quickly
- **Fairness**: Multiple tiers of review before dismissal
- **Transparency**: Full audit trail of all decisions

This implementation transforms the report system from a potential false-negative vulnerability into a robust, multi-tier fraud detection pipeline.
