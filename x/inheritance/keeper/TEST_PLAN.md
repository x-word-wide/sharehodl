# Inheritance Module Test Plan

## Overview
Comprehensive test suite for the ShareHODL inheritance module (dead man switch / next of kin system).

## Test Files Created

### 1. `keeper_test.go` - Basic Keeper Tests
**Status**: ✅ Created
**Coverage**: Plan CRUD operations, validation

**Key Tests**:
- `TestCreateInheritancePlan` - Plan creation
- `TestUpdateInheritancePlan` - Plan updates
- `TestCancelInheritancePlan` - Plan cancellation
- `TestGetPlansByOwner` - Owner queries
- `TestGetPlansByBeneficiary` - Beneficiary queries
- `TestPlanValidation` - Validation rules (percentages, grace period, etc.)
- `TestBeneficiaryValidation` - Beneficiary validation
- `TestSpecificAssetValidation` - Asset validation

**Critical Test Cases**:
- ❌ Reject plans with beneficiary percentages > 100%
- ❌ Reject plans with grace period < 30 days
- ❌ Reject plans with duplicate priorities
- ✅ Allow percentages < 100% (charity gets remainder)

---

### 2. `activity_tracker_test.go` - Activity Tracking Tests
**Status**: ✅ Created
**Coverage**: Activity recording, inactivity detection

**Key Tests**:
- `TestRecordActivity` - Activity logging
- `TestGetLastActivity` - Activity retrieval
- `TestIsInactive` - Inactivity detection
- `TestAutoCancelOnActivity` - False positive protection
- `TestInactivityCalculation` - Threshold calculations

**Critical Test Cases**:
- ✅ ANY signed transaction resets inactivity
- ✅ False positive protection (owner recovers wallet)
- ✅ Activity during grace period cancels trigger
- ⚠️ Test boundary conditions (exactly at threshold)

---

### 3. `switch_processor_test.go` - Switch Trigger Tests
**Status**: ✅ Created
**Coverage**: Dead man switch triggering, grace periods

**Key Tests**:
- `TestTriggerSwitch_Success` - Normal trigger flow
- `TestTriggerSwitch_OwnerBanned` - **CRITICAL**: Ban bypass prevention
- `TestTriggerSwitch_AlreadyTriggered` - Double trigger prevention
- `TestCancelTrigger_DuringGracePeriod` - Grace period cancellation
- `TestCancelTrigger_AfterGracePeriod` - Rejection after grace period
- `TestAutoCancelOnOwnerTransaction` - Auto-cancel on activity

**Critical Test Cases**:
- 🔒 **BANNED OWNER CANNOT TRIGGER** (prevents fraud funnel)
- ✅ Grace period minimum 30 days enforced
- ✅ Owner transaction auto-cancels during grace period
- ❌ Cannot cancel after grace period expires

---

### 4. `claim_processor_test.go` - Claim Processing Tests
**Status**: ✅ Created
**Coverage**: Beneficiary claims, cascading logic

**Key Tests**:
- `TestClaimInheritance_Success` - Normal claim flow
- `TestClaimInheritance_NotClaimable` - Premature claim rejection
- `TestClaimInheritance_BeneficiaryBanned` - **CRITICAL**: Cascade when banned
- `TestClaimInheritance_AlreadyClaimed` - Double claim prevention
- `TestClaimInheritance_NotBeneficiary` - Authorization check
- `TestCascadingBeneficiaries` - Cascade logic

**Critical Test Cases**:
- 🔒 **BANNED BENEFICIARY AUTO-SKIPS** (cascades to next)
- ✅ Priority-based claim ordering
- ✅ Claim windows expire sequentially
- ✅ Charity receives unclaimed remainder

**Cascade Logic**:
```
Plan: Ben1 (60%), Ben2 (40%), Charity
- Ben1 banned → Ben2 gets 100%, Charity gets 0%
- Ben1 + Ben2 banned → Charity gets 100%
```

---

### 5. `asset_handler_test.go` - Asset Transfer Tests
**Status**: ✅ Created
**Coverage**: HODL, equity, and mixed asset transfers

**Key Tests**:
- `TestTransferHODL` - HODL token transfers
- `TestTransferEquityShares` - Equity share transfers
- `TestTransferMultipleAssetTypes` - Mixed asset handling
- `TestTransferToCharity` - Charity fallback
- `TestLockedAssetsBlocked` - Lock enforcement during inheritance
- `TestAssetTransferAtomic` - Atomic transfer guarantees

**Critical Test Cases**:
- 🔒 Locked assets cannot be transferred during inheritance
- ✅ Atomic transfers (all or nothing)
- ✅ Correct percentage calculations with rounding
- ✅ Handles insufficient assets gracefully

---

### 6. `endblock_test.go` - EndBlock Processing Tests
**Status**: ✅ Created
**Coverage**: Automated processing, batch operations

**Key Tests**:
- `TestProcessInactivityChecks` - Inactivity monitoring
- `TestProcessGracePeriodExpiry` - Grace period expiration
- `TestProcessClaimWindowExpiry` - Claim window expiration
- `TestProcessUltraLongInactivity` - 50 year charity fallback
- `TestEndBlockBatchProcessing` - Batch efficiency
- `TestEndBlockPerformanceBenchmark` - Performance under load

**Critical Test Cases**:
- ⚡ Handle 1000+ active plans efficiently
- ✅ Process multiple plans in single block
- ✅ Gas optimization strategies
- ✅ Deterministic processing across validators

---

### 7. `integration_test.go` - End-to-End Integration Tests
**Status**: ✅ Created
**Coverage**: Complete workflows, edge cases

**Key Tests**:
- `TestFullInheritanceWorkflow` - Complete happy path
- `TestFalsePositiveProtection` - Owner recovers wallet
- `TestBannedOwnerCannotUseInheritance` - **CRITICAL**: Security test
- `TestBannedBeneficiaryCascades` - Cascade scenarios
- `Test50YearCharityFallback` - Ultra-long inactivity
- `TestMultipleBeneficiariesPartialClaims` - Complex scenarios

**Critical Test Cases**:
- 🔒 **FRAUDSTER CANNOT USE INHERITANCE TO BYPASS BAN**
- ✅ False positive protection works
- ✅ Beneficiary cascade logic correct
- ✅ 50 year threshold goes to charity
- ✅ Cross-module interactions work

---

## Critical Security Tests

### 1. Ban Bypass Prevention (`switch_processor_test.go`)
**Priority**: 🔴 CRITICAL
**Test**: `TestTriggerSwitch_OwnerBanned`

**Scenario**:
1. Bob creates inheritance plan with wife as beneficiary
2. Bob commits fraud and gets banned
3. Bob attempts to trigger inheritance switch
4. **MUST FAIL** with error: "owner is banned, inheritance disabled"

**Security Rationale**:
- Fraudsters cannot use inheritance to funnel assets to family
- Banned user's assets should be frozen/seized
- Inheritance would allow indirect asset retention

---

### 2. Beneficiary Ban Cascade (`claim_processor_test.go`)
**Priority**: 🔴 CRITICAL
**Test**: `TestClaimInheritance_BeneficiaryBanned`

**Scenario**:
1. Plan: Ben1 (60%), Ben2 (40%), Charity
2. Grace period expires
3. Ben1 is banned before claim
4. Ben1's claim auto-skips
5. Ben2 receives 100% (60% + 40%)

**Security Rationale**:
- Banned beneficiaries cannot receive inheritance
- Ban check happens at claim time (not trigger time)
- Cascade ensures assets don't go to banned users

---

## Test Execution Priority

### Phase 1: Foundation (Week 1)
1. ✅ `keeper_test.go` - Basic CRUD and validation
2. ✅ `activity_tracker_test.go` - Activity tracking

### Phase 2: Core Logic (Week 2)
3. ✅ `switch_processor_test.go` - Trigger logic
4. ✅ `claim_processor_test.go` - Claim logic

### Phase 3: Advanced (Week 3)
5. ✅ `asset_handler_test.go` - Asset transfers
6. ✅ `endblock_test.go` - Automated processing

### Phase 4: Integration (Week 4)
7. ✅ `integration_test.go` - End-to-end workflows

---

## Test Coverage Targets

| Component | Target | Priority |
|-----------|--------|----------|
| Keeper methods | >90% | High |
| Validation logic | 100% | Critical |
| Ban checks | 100% | Critical |
| Cascade logic | 100% | High |
| Asset transfers | >85% | High |
| EndBlock processing | >80% | Medium |
| Edge cases | >75% | Medium |

---

## Test Data

### Valid Test Addresses
```go
testOwner        = sdk.AccAddress([]byte("inheritance_owner___")).String()
testBeneficiary1 = sdk.AccAddress([]byte("beneficiary_one_____")).String()
testBeneficiary2 = sdk.AccAddress([]byte("beneficiary_two_____")).String()
testBeneficiary3 = sdk.AccAddress([]byte("beneficiary_three___")).String()
testCharity      = sdk.AccAddress([]byte("charity_wallet______")).String()
```

### Standard Test Plan
```go
plan := types.InheritancePlan{
    PlanID: 1,
    Owner: testOwner,
    Beneficiaries: []types.Beneficiary{
        {Address: testBeneficiary1, Priority: 1, Percentage: 0.60},
        {Address: testBeneficiary2, Priority: 2, Percentage: 0.40},
    },
    InactivityPeriod: 365 * 24 * time.Hour, // 1 year
    GracePeriod: 30 * 24 * time.Hour,       // 30 days
    ClaimWindow: 180 * 24 * time.Hour,      // 180 days
    CharityAddress: testCharity,
    Status: types.PlanStatusActive,
}
```

---

## Running Tests

### Run all tests
```bash
go test ./x/inheritance/keeper/... -v
```

### Run specific test file
```bash
go test ./x/inheritance/keeper/keeper_test.go -v
```

### Run with coverage
```bash
go test ./x/inheritance/keeper/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run integration tests only
```bash
go test ./x/inheritance/keeper/integration_test.go -v
```

---

## Known Limitations

### Tests Marked `t.Skip("Requires full keeper setup")`
Most tests are currently marked with `t.Skip` because they require:
1. Full keeper initialization with mock dependencies
2. Mock bank keeper, equity keeper, escrow keeper
3. Test context with block time management
4. State store setup

### Next Steps
1. Create keeper test framework with mocks
2. Implement mock keepers for dependencies
3. Remove `t.Skip` and implement full test logic
4. Add integration with CI/CD

---

## Dependencies

### Required Mock Keepers
- `BankKeeper` - For HODL transfers and balance queries
- `EquityKeeper` - For share transfers
- `EscrowKeeper` - For ban status checks (ReporterHistory)
- `ValidatorKeeper` - For tier checks (if needed)

### Test Utilities Needed
- `setupTestKeeper(t)` - Initialize keeper with mocks
- `createTestPlan(owner, years)` - Create test plan
- `createTestBeneficiary(addr, priority, pct)` - Create beneficiary
- `simulateBlockTime(ctx, duration)` - Time travel for tests

---

## Success Criteria

### Must Pass Before Production
- ✅ All validation tests pass (100%)
- ✅ All security tests pass (ban bypass prevention)
- ✅ All cascade logic tests pass
- ✅ Integration tests pass for happy path
- ✅ Performance tests meet targets (1000+ plans)
- ✅ No memory leaks or resource exhaustion

### Nice to Have
- Edge case coverage >75%
- Stress tests with 10,000+ plans
- Migration tests for upgrades
- Fuzzing tests for input validation

---

**Test Suite Status**: 📋 Planning Complete, Implementation Pending
**Security Review**: ⏳ Awaiting Implementation
**Last Updated**: 2026-01-30
