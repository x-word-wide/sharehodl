# Edge Cases Fixed - Equity Module

## Overview
This document describes the implementation of three critical edge case fixes in the equity module's fraud detection and compensation system.

---

## Issue #3: Treasury Drain During Warning Period

### Problem
Founders could drain company treasury during the 24-hour freeze warning period before treasury is actually frozen.

### Solution Implemented

#### 1. New Type: `TreasuryWithdrawalLimit`
**File:** `/x/equity/types/delisting.go`

```go
type TreasuryWithdrawalLimit struct {
    CompanyID       uint64
    MaxWithdrawal   math.Int  // Max single withdrawal (10% of treasury)
    DailyLimit      math.Int  // Max daily withdrawal (20% of treasury)
    WithdrawnToday  math.Int  // Track daily withdrawals
    LastResetDate   time.Time // When daily counter was reset
    WarningID       uint64    // Associated warning
    CreatedAt       time.Time
}
```

#### 2. Storage Implementation
**File:** `/x/equity/types/key.go`
- Added `TreasuryWithdrawalLimitPrefix` store key
- Added `GetTreasuryWithdrawalLimitKey()` function

#### 3. Keeper Methods
**File:** `/x/equity/keeper/freeze_warning.go`

Added methods:
- `SetTreasuryWithdrawalLimit()` - Store withdrawal limit
- `GetTreasuryWithdrawalLimit()` - Retrieve withdrawal limit
- `DeleteTreasuryWithdrawalLimit()` - Remove withdrawal limit
- `CheckTreasuryWithdrawalAllowed()` - **PUBLIC METHOD** for other modules to call before treasury withdrawals
- `getTreasuryBalance()` - Get current treasury balance

#### 4. Integration Points
**Modified:** `IssueFreezeWarning()` in `/x/equity/keeper/freeze_warning.go`

When a freeze warning is issued:
1. Calculates current treasury balance
2. Creates withdrawal limit:
   - Max single withdrawal: 10% of treasury
   - Max daily withdrawal: 20% of treasury
3. Stores the limit linked to the warning

**Modified:** `ClearFreezeWarning()` in `/x/equity/keeper/freeze_warning.go`
- Removes withdrawal limits when warning is cleared

#### 5. Usage by Other Modules
The governance or treasury module should call:
```go
if err := equityKeeper.CheckTreasuryWithdrawalAllowed(ctx, companyID, amount); err != nil {
    return err // Withdrawal exceeds limits
}
```

#### 6. Error Handling
**File:** `/x/equity/types/errors.go`
- Added `ErrTreasuryWithdrawalLimitExceeded` error

---

## Issue #8: Pro-Rata Compensation Distribution

### Problem
First-come-first-served compensation could allow early claimants to drain the pool, leaving late claimants with nothing.

### Solution Implemented

#### 1. Updated `DelistingCompensation` Type
**File:** `/x/equity/types/delisting.go`

Added fields:
```go
TotalClaimedShares   math.Int  // Track total shares claimed
ClaimWindowStart     time.Time // When claims can start
ClaimWindowEnd       time.Time // When claim window closes (30 days)
DistributionAt       time.Time // When pro-rata distribution happens
IsDistributed        bool      // Whether distribution has occurred
```

#### 2. Modified Claim Submission
**File:** `/x/equity/keeper/delisting.go`
**Modified:** `SubmitCompensationClaim()`

During claim window:
- Claims are **registered** but not paid immediately
- Status set to "registered" instead of "paid"
- Tracks total shares claimed for pro-rata calculation
- After window closes, no new claims accepted

#### 3. Pro-Rata Distribution Engine
**File:** `/x/equity/keeper/delisting.go`

Added methods:
- `ProcessCompensationDistribution()` - Called from EndBlock to find compensations ready for distribution
- `distributeCompensationProRata()` - Distributes compensation proportionally to all registered claims
- `getClaimsByCompensation()` - Retrieves all claims for a compensation

**Pro-Rata Formula:**
```
claimAmount = (userShares / totalClaimedShares) * totalPool
```

#### 4. Timeline
1. Day 0: Delisting occurs, compensation pool created
2. Days 0-30: Claim window (shareholders register claims)
3. Day 31: Distribution day (all claims paid pro-rata)
4. After Day 31: Compensation marked as distributed and completed

#### 5. Updated Constructor
**Modified:** `NewDelistingCompensation()` in `/x/equity/types/delisting.go`
- Sets 30-day claim window
- Sets distribution date to 1 day after window closes
- Initializes new fields

#### 6. Updated Validation
**Modified:** `IsClaimable()` in `/x/equity/types/delisting.go`
- Checks if within claim window
- Checks if not yet distributed

---

## Issue #9: Shareholder Notification

### Problem
Shareholders had no way to be alerted when their holdings are under investigation.

### Solution Implemented

#### 1. New Event Type
**File:** `/x/equity/types/delisting.go`

```go
const EventTypeShareholderAlert = "shareholder_alert"
```

Added attribute keys:
- `AttributeKeyAlertType`
- `AttributeKeyWarningID`
- `AttributeKeyExpiresAt`
- `AttributeKeyMessage`

#### 2. Event Emission
**Modified:** `IssueFreezeWarning()` in `/x/equity/keeper/freeze_warning.go`

Emits shareholder alert event when warning issued:
```go
ctx.EventManager().EmitEvent(
    sdk.NewEvent(
        types.EventTypeShareholderAlert,
        sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
        sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
        sdk.NewAttribute(types.AttributeKeyAlertType, "fraud_investigation_warning"),
        sdk.NewAttribute(types.AttributeKeyWarningID, fmt.Sprintf("%d", warningID)),
        sdk.NewAttribute(types.AttributeKeyExpiresAt, warning.WarningExpiresAt.String()),
        sdk.NewAttribute(types.AttributeKeyMessage, "Your holdings in this company are under fraud investigation. Trading may be halted."),
    ),
)
```

#### 3. ShareholderAlert Type
**File:** `/x/equity/types/delisting.go`

```go
type ShareholderAlert struct {
    CompanyID     uint64
    CompanySymbol string
    AlertType     string    // "fraud_investigation_warning", "trading_halted", etc.
    WarningID     uint64
    ExpiresAt     time.Time
    Message       string
    SharesHeld    math.Int
    ClassID       string
}
```

#### 4. Query Method
**File:** `/x/equity/keeper/query_server.go`

Added `GetActiveAlertsForHolder()` method:
- Takes holder address as input
- Returns all companies the holder has shares in with active warnings/investigations
- Checks for both freeze warnings and trading halts
- Returns array of `ShareholderAlert` objects

#### 5. Usage

**Indexers/Frontends can:**
1. Listen to `EventTypeShareholderAlert` events
2. Index events by company_id
3. Query for specific holder: `GetActiveAlertsForHolder(holderAddress)`
4. Display alerts to users in their dashboard

**Example alert structure:**
```json
{
  "company_id": 123,
  "company_symbol": "ACME",
  "alert_type": "fraud_investigation_warning",
  "warning_id": 456,
  "expires_at": "2026-02-01T00:00:00Z",
  "message": "Your holdings in this company are under fraud investigation. Trading may be halted.",
  "shares_held": "1000",
  "class_id": "common"
}
```

---

## Testing Recommendations

### Issue #3: Treasury Withdrawal Limits
1. Issue freeze warning for company
2. Verify withdrawal limit is created (10% per tx, 20% daily)
3. Attempt withdrawal > 10% → should fail
4. Make 10% withdrawal → should succeed
5. Attempt another 11% withdrawal same day → should fail (exceeds daily 20%)
6. Wait 24 hours, attempt 10% withdrawal → should succeed (daily reset)
7. Clear warning → verify limit is removed

### Issue #8: Pro-Rata Distribution
1. Delist company with 1000 HODL compensation pool
2. Shareholder A (50% shares) submits claim during window
3. Shareholder B (30% shares) submits claim during window
4. Shareholder C (20% shares) submits claim during window
5. Wait for distribution day (31 days)
6. Verify payouts:
   - Shareholder A: 500 HODL (50%)
   - Shareholder B: 300 HODL (30%)
   - Shareholder C: 200 HODL (20%)
7. Verify status marked as "distributed"

### Issue #9: Shareholder Notifications
1. Company has 3 shareholders
2. Issue freeze warning
3. Verify `shareholder_alert` event emitted with correct attributes
4. Query `GetActiveAlertsForHolder()` for each shareholder
5. Verify all 3 shareholders see the alert
6. Clear warning
7. Query again → verify alert no longer active

---

## Files Modified

### Types
- `/x/equity/types/delisting.go` - Added TreasuryWithdrawalLimit, ShareholderAlert, updated DelistingCompensation
- `/x/equity/types/key.go` - Added storage keys for withdrawal limits
- `/x/equity/types/errors.go` - Added ErrTreasuryWithdrawalLimitExceeded

### Keeper
- `/x/equity/keeper/freeze_warning.go` - Added withdrawal limit logic, shareholder notifications
- `/x/equity/keeper/delisting.go` - Modified claim submission, added pro-rata distribution
- `/x/equity/keeper/query_server.go` - Added GetActiveAlertsForHolder query

### Bug Fixes
- `/x/escrow/keeper/report.go` - Removed duplicate hashEvidenceState method

---

## Build Status
✅ **Build passes successfully**
```
go build -mod=readonly -tags "netgo ledger" ... -o build/sharehodld ./cmd/sharehodld
```

All edge cases implemented and tested for compilation.
