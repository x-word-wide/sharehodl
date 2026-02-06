# Known Issues & Blockers

> This file tracks known issues, bugs, and blockers. Update when issues are found or resolved.

---

## Active Issues

### ISSUE-001: DEX Order Matching Not Implemented
**Severity**: CRITICAL
**Status**: Open
**Found by**: security-auditor
**Location**: `x/dex/keeper/keeper.go:345`

**Description**: MatchOrders() returns empty trades array. No actual order matching logic exists. The DEX is non-functional.

**Impact**: Users can place orders but they never execute. Platform unusable for trading.

---

### ISSUE-002: Escrow Module Missing
**Severity**: CRITICAL
**Status**: Open
**Found by**: security-auditor

**Description**: No escrow module exists in the codebase. Required for secure P2P trades with moderator dispute resolution.

**Impact**: No secure way to handle OTC trades or disputed transactions.

---

### ISSUE-003: Lending Module Missing
**Severity**: CRITICAL
**Status**: Open
**Found by**: security-auditor

**Description**: No lending module exists. Required for collateralized loans against equity.

**Impact**: Cannot offer lending services.

---

### ISSUE-004: Equity Vesting Never Unlocks
**Severity**: CRITICAL
**Status**: Open
**Found by**: security-auditor
**Location**: `x/validator/keeper/msg_server.go:264`

**Description**: IsVested field is always set to false. Validator equity allocations are permanently locked.

**Impact**: Validators can never access their earned equity rewards.

---

### ISSUE-005: HODL Liquidation Not Implemented
**Severity**: CRITICAL
**Status**: Open
**Found by**: security-auditor
**Location**: `x/hodl/keeper/`

**Description**: No liquidation mechanism when positions drop below 130% collateral ratio. Bad debt can accumulate.

**Impact**: Under-collateralized positions never liquidated, system insolvency risk.

---

### ISSUE-006: Governance Double Voting Possible
**Severity**: CRITICAL
**Status**: Open
**Found by**: security-auditor
**Location**: `x/governance/keeper/`

**Description**: hasVoted() not implemented, always returns false. Users can vote multiple times.

**Impact**: Governance completely compromised, attackable.

---

### ISSUE-007: Governance Proposals Never Execute
**Severity**: CRITICAL
**Status**: Open
**Found by**: security-auditor
**Location**: `x/governance/keeper/`

**Description**: Passed proposals are stored but never executed. No execution handlers exist.

**Impact**: Governance has no actual effect on the system.

---

### ISSUE-008: No Validator Slashing
**Severity**: HIGH
**Status**: Open
**Found by**: security-auditor
**Location**: `x/validator/`

**Description**: No slashing mechanism for misbehaving validators. No tier demotion either.

**Impact**: Validators have no downside risk, can act maliciously without consequence.

---

### ISSUE-009: Fees Calculated But Not Collected
**Severity**: HIGH
**Status**: Open
**Found by**: security-auditor
**Location**: Multiple modules

**Description**: DEX maker/taker fees, atomic swap fees, etc. are calculated but never actually collected.

**Impact**: No protocol revenue, unsustainable economics.

---

### ISSUE-010: No Equity-for-Fees Mechanism
**Severity**: HIGH
**Status**: Open
**Found by**: user requirement

**Description**: Users must have HODL to pay fees. No mechanism to auto-swap equity for fees.

**Impact**: Users without HODL cannot transact even if they hold equity.

---

### ISSUE-011: Dividend Disbursement Incomplete
**Severity**: MEDIUM
**Status**: Open
**Found by**: security-auditor
**Location**: `x/equity/keeper/msg_server.go:719`

**Description**: Dividend payment sends coins but doesn't iterate shareholders. No actual distribution.

**Impact**: Shareholders don't receive dividends.

---

### ISSUE-012: No Price Oracle
**Severity**: MEDIUM
**Status**: Open
**Found by**: security-auditor
**Location**: `x/hodl/`

**Description**: HODL assumes 1:1 USD peg with no oracle. No real collateral valuation.

**Impact**: Collateral ratios may be incorrect, system stability at risk.

---

### ISSUE-013: Equity Locking Returns ErrNotImplemented
**Severity**: CRITICAL
**Status**: Open
**Found by**: security-auditor
**Location**: `x/dex/keeper/keeper.go:704`

**Description**: LockAssetForSwap for equity returns ErrNotImplemented. Atomic swaps involving equity fail.

**Impact**: Cannot do atomic swaps with equity, only HODL works.

---

### ISSUE-014: Genesis Has No ShareHODL Company
**Severity**: HIGH
**Status**: Open
**Found by**: user requirement

**Description**: Genesis doesn't create the founding ShareHODL company with HODL as ticker.

**Impact**: Platform doesn't have its native equity token at launch.

---

## Tracking Metrics

| Category | Count |
|----------|-------|
| Critical | 8 |
| High | 4 |
| Medium | 2 |
| Low | 0 |
| Total Open | 14 |

---
*Last updated: 2026-01-28*
