# Shareholder Petition System

## Overview

The ShareHODL Shareholder Petition System allows regular shareholders (who don't have the 10K HODL Keeper stake required for direct fraud reporting) to collectively flag fraud concerns and unusual activity for investigation.

This system provides **democratic governance** while preventing **single-person griefing** through collective threshold requirements.

## Features

### Core Capabilities

1. **Accessible Reporting**: Any shareholder with at least 1 share can create a petition
2. **Collective Validation**: Requires 10% of shareholders OR 100 signatures to convert to formal report
3. **Automatic Conversion**: When threshold met, automatically creates elevated-priority fraud report
4. **Time-Limited**: 7-day window to gather signatures
5. **Transparent**: All signatures and comments are on-chain

### Petition Types

```go
const (
    PetitionTypeFraudConcern           // General fraud concerns
    PetitionTypeUnusualActivity        // Suspicious trading/financial activity
    PetitionTypeManagementMisconduct   // Director/executive misconduct
)
```

## How It Works

### 1. Create Petition

```go
// Any shareholder can create a petition
petitionID, err := keeper.CreatePetition(
    ctx,
    creator,        // Must hold >= 1 share
    companyID,
    classID,        // Which share class (e.g., "COMMON")
    petitionType,
    title,
    description,
)
```

**Requirements:**
- Creator must hold at least 1 share in the company
- Petition expires in 7 days if threshold not met

### 2. Gather Signatures

```go
// Other shareholders sign to show support
err := keeper.SignPetition(
    ctx,
    petitionID,
    signer,   // Must hold >= 1 share
    comment,  // Optional comment
)
```

**Anti-Griefing Measures:**
- Each address can only sign once
- Must hold shares in the company
- Signatures include share count (for transparency)

### 3. Threshold Check (Automatic)

Called from `EndBlock`, the system automatically checks all open petitions:

```go
func (k Keeper) ProcessPetitionThresholds(ctx sdk.Context) {
    // For each open petition:
    // 1. Check if expired (7 days)
    // 2. Check if threshold met:
    //    - 10% of total shareholders, OR
    //    - 100 absolute signatures (whichever is lower)
    // 3. If threshold met, convert to formal report
}
```

### 4. Automatic Conversion

When threshold is met:

```go
// Automatically creates report with:
- Report Type: ReportTypeFraud
- Priority: Based on signature count (3-5)
  - 200+ signatures = Priority 5 (Emergency)
  - 150+ signatures = Priority 4 (High)
  - 100+ signatures = Priority 3 (Medium)
- Severity: 4 (High)
- Description: Includes petition details + top 10 signatures with comments
```

**Elevated Priority**: The converted report simulates a Keeper-tier reporter, giving it higher priority in the investigation queue.

## Thresholds

### Default Configuration

```go
const (
    DefaultShareholderThreshold = 10     // 10% of shareholders
    DefaultAbsoluteThreshold    = 100    // OR 100 signatures
    PetitionExpirationDays      = 7      // 7 days to gather support
)
```

### Threshold Logic (Either/Or)

```go
func (p ShareholderPetition) CheckThreshold(totalShareholders int) (bool, string) {
    // Check absolute threshold (easier for large companies)
    if p.SignatureCount >= p.AbsoluteThreshold {
        return true, "absolute threshold met"
    }

    // Check percentage threshold
    requiredSignatures := int(float64(totalShareholders) * 0.10) // 10%
    if p.SignatureCount >= requiredSignatures {
        return true, "percentage threshold met"
    }

    return false, "threshold not met"
}
```

## Data Structures

### ShareholderPetition

```go
type ShareholderPetition struct {
    ID         uint64
    CompanyID  uint64
    ClassID    string  // Which share class

    PetitionType PetitionType
    Title        string
    Description  string
    Creator      string  // Original petitioner

    // Signatures
    Signatures     []PetitionSignature
    SignatureCount int

    // Thresholds (either/or)
    ShareholderThreshold int  // 10% of shareholders
    AbsoluteThreshold    int  // OR 100 signatures
    ThresholdMet         bool

    // Timeline
    CreatedAt   time.Time
    ExpiresAt   time.Time  // 7 days
    ConvertedAt time.Time

    // Conversion
    ConvertedToReportID uint64
    Status              PetitionStatus
}
```

### PetitionSignature

```go
type PetitionSignature struct {
    Signer     string
    SharesHeld math.Int  // How many shares they hold
    SignedAt   time.Time
    Comment    string    // Optional explanation
}
```

## Messages

### 1. CreatePetition

```go
type SimpleMsgCreatePetition struct {
    Creator      string
    CompanyID    uint64
    ClassID      string
    PetitionType PetitionType
    Title        string
    Description  string
}
```

### 2. SignPetition

```go
type SimpleMsgSignPetition struct {
    Signer     string
    PetitionID uint64
    Comment    string  // Optional
}
```

### 3. WithdrawPetition

```go
type SimpleMsgWithdrawPetition struct {
    Creator    string
    PetitionID uint64
}
```

## Keeper Methods

### Core Operations

```go
// Create a new petition
CreatePetition(ctx, creator, companyID, classID, petitionType, title, description) (uint64, error)

// Sign an existing petition
SignPetition(ctx, petitionID, signer, comment) error

// Withdraw a petition (creator only)
WithdrawPetition(ctx, petitionID, withdrawer) error

// Process thresholds (called from EndBlock)
ProcessPetitionThresholds(ctx)
```

### Queries

```go
// Get petition by ID
GetPetition(ctx, petitionID) (ShareholderPetition, bool)

// Get all petitions for a company
GetPetitionsByCompany(ctx, companyID) []ShareholderPetition

// Get petitions created by an address
GetPetitionsByCreator(ctx, creator) []ShareholderPetition

// Get all open petitions
GetOpenPetitions(ctx) []ShareholderPetition
```

## Events

### petition_created

```go
sdk.NewEvent(
    "petition_created",
    sdk.NewAttribute("petition_id", petitionID),
    sdk.NewAttribute("company_id", companyID),
    sdk.NewAttribute("class_id", classID),
    sdk.NewAttribute("creator", creator),
    sdk.NewAttribute("petition_type", petitionType),
    sdk.NewAttribute("title", title),
)
```

### petition_signed

```go
sdk.NewEvent(
    "petition_signed",
    sdk.NewAttribute("petition_id", petitionID),
    sdk.NewAttribute("signer", signer),
    sdk.NewAttribute("shares_held", sharesHeld),
    sdk.NewAttribute("signature_count", signatureCount),
)
```

### petition_threshold_met

```go
sdk.NewEvent(
    "petition_threshold_met",
    sdk.NewAttribute("petition_id", petitionID),
    sdk.NewAttribute("company_id", companyID),
    sdk.NewAttribute("reason", reason),
    sdk.NewAttribute("signature_count", signatureCount),
    sdk.NewAttribute("converted_to_report", reportID),
)
```

### petition_expired

```go
sdk.NewEvent(
    "petition_expired",
    sdk.NewAttribute("petition_id", petitionID),
    sdk.NewAttribute("company_id", companyID),
    sdk.NewAttribute("signature_count", signatureCount),
)
```

### petition_withdrawn

```go
sdk.NewEvent(
    "petition_withdrawn",
    sdk.NewAttribute("petition_id", petitionID),
    sdk.NewAttribute("withdrawer", withdrawer),
    sdk.NewAttribute("signature_count", signatureCount),
)
```

## Storage

### Keys

```go
// Main storage
PetitionPrefix = []byte{0x80}  // petition_id -> ShareholderPetition

// Counters
PetitionCounterKey = []byte{0x81}  // Global petition ID counter

// Indexes
PetitionByCompanyPrefix  = []byte{0x82}  // company_id -> []petition_id
PetitionByCreatorPrefix  = []byte{0x83}  // creator -> []petition_id
```

## Integration with Escrow Module

When a petition meets its threshold, it emits an event that can be consumed by the escrow module:

```go
ctx.EventManager().EmitEvent(
    sdk.NewEvent(
        "petition_convert_to_report",
        sdk.NewAttribute("petition_id", petitionID),
        sdk.NewAttribute("company_id", companyID),
        sdk.NewAttribute("petition_type", petitionType),
        sdk.NewAttribute("priority", priority),  // 3-5 based on signatures
        sdk.NewAttribute("severity", "4"),       // High
        sdk.NewAttribute("signature_count", signatureCount),
        // ... full petition details
    ),
)
```

The escrow module can then create a formal fraud report with elevated priority.

## Example Flow

### Scenario: Suspicious Activity Detection

1. **Alice** (shareholder with 100 shares) notices unusual trading:
   ```
   CreatePetition(
       creator: "alice",
       companyID: 42,
       classID: "COMMON",
       type: PetitionTypeUnusualActivity,
       title: "Suspicious Insider Trading Pattern",
       description: "Large sells right before bad news announcements",
   )
   ```

2. **Other shareholders** review and sign:
   ```
   SignPetition(petitionID: 1, signer: "bob", comment: "I noticed this too")
   SignPetition(petitionID: 1, signer: "carol", comment: "Very suspicious timing")
   SignPetition(petitionID: 1, signer: "dave", comment: "Agree, needs investigation")
   ... (97 more signatures)
   ```

3. **EndBlock** detects threshold met:
   ```
   ProcessPetitionThresholds() detects:
   - 100 signatures reached (absolute threshold)
   - Converts to Priority 3 fraud report
   - Trading may be halted pending investigation
   ```

4. **Escrow validators** investigate:
   - High-priority report assigned to Steward+ tier validators
   - Evidence reviewed from petition signatures
   - Decision made on company status

## Security Considerations

### Anti-Griefing Measures

1. **Minimum Share Requirement**: Must hold at least 1 share
2. **No Duplicate Signatures**: Each address can only sign once
3. **Collective Threshold**: Prevents single-person attacks
4. **Time Limit**: 7-day expiration prevents stale petitions
5. **On-Chain Transparency**: All signatures are public

### Attack Vectors & Mitigations

| Attack | Mitigation |
|--------|-----------|
| Sybil (multiple addresses) | Must hold shares in each address |
| Spam petitions | 1-share minimum + expiration |
| Coordinated false reporting | High threshold (10%/100) + validator review |
| Trading manipulation | Converted reports can halt trading if confirmed |

## Testing

### Unit Tests

```go
// test_petition.go
func TestCreatePetition(t *testing.T)
func TestSignPetition(t *testing.T)
func TestPetitionThreshold(t *testing.T)
func TestPetitionExpiration(t *testing.T)
func TestDuplicateSignature(t *testing.T)
func TestNonShareholderPetition(t *testing.T)
```

### Integration Tests

```go
// test_petition_integration.go
func TestPetitionToReportConversion(t *testing.T)
func TestMultipleShareClasses(t *testing.T)
func TestLargeShareholderBase(t *testing.T)
```

## Future Enhancements

1. **Weighted Voting**: Consider share count in threshold (not just signature count)
2. **Tiered Thresholds**: Different thresholds based on company size
3. **Cross-Class Petitions**: Allow petitions across multiple share classes
4. **Petition Amendments**: Allow updates to petition description
5. **Signature Withdrawal**: Allow unsigning before conversion

## References

- Escrow Module: `/x/escrow/keeper/report.go`
- Fraud Reporting: `/x/escrow/types/report.go`
- Equity Module: `/x/equity/keeper/keeper.go`
- EndBlock Processing: `/x/equity/module.go`
