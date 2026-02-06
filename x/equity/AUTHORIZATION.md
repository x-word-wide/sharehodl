# Company Authorization System

## Overview

The Company Authorization system enables co-founders and delegated governance for ShareHODL companies. It provides a role-based access control system for treasury operations, primary sales, and share dilution proposals.

## Key Concepts

### Roles

1. **Owner (Role 2)**: Full control over the company
   - Add/remove authorized proposers
   - Transfer ownership (two-step process)
   - All proposer privileges

2. **Proposer (Role 1)**: Can create proposals
   - Request treasury withdrawals
   - Create primary sale offers
   - Propose share dilution
   - Cannot modify authorization

3. **None (Role 0)**: No special privileges
   - Can only interact as a shareholder

### Authorization Structure

```go
type CompanyAuthorization struct {
    CompanyID            uint64    // Company this applies to
    Owner                string    // Current owner (founder initially)
    AuthorizedProposers  []string  // Co-founders/proposers
    PendingOwnerTransfer string    // Two-step ownership transfer
    CreatedAt            time.Time
    UpdatedAt            time.Time
}
```

## Use Cases

### 1. Co-Founder Scenario

Alice and Bob start a company together:

```go
// Alice creates the company (becomes owner)
CreateCompany(founder: alice, ...)

// Alice adds Bob as authorized proposer
MsgAddAuthorizedProposer{
    Creator: alice,
    CompanyID: 1,
    NewProposer: bob,
}

// Now both Alice and Bob can:
// - Request treasury withdrawals
// - Create primary sale offers
// - Propose share dilution

// Only Alice can:
// - Add/remove proposers
// - Transfer ownership
```

### 2. CEO Transition

Company transfers control from founder to hired CEO:

```go
// Step 1: Founder initiates transfer
MsgInitiateOwnershipTransfer{
    Creator: founder,
    CompanyID: 1,
    NewOwner: newCEO,
}
// Status: Pending transfer to newCEO

// Step 2: New CEO accepts
MsgAcceptOwnershipTransfer{
    Creator: newCEO,
    CompanyID: 1,
}
// Status: newCEO is now owner, founder loses privileges
```

### 3. Board of Directors

Multiple board members can propose, but chairman controls access:

```go
// Chairman (owner) adds board members as proposers
AddAuthorizedProposer(director1)
AddAuthorizedProposer(director2)
AddAuthorizedProposer(director3)

// Any board member can now propose treasury withdrawals
// But only chairman can add/remove board members
```

## API Reference

### Query Authorization

```go
// Check if address is owner
IsCompanyOwner(ctx, companyID, address) bool

// Check if address can propose (owner OR proposer)
CanProposeForCompany(ctx, companyID, address) bool

// Get full authorization record
GetCompanyAuthorization(ctx, companyID) (CompanyAuthorization, bool)
```

### Manage Proposers

```go
// Add proposer (owner only)
AddAuthorizedProposer(ctx, companyID, owner, newProposer) error

// Remove proposer (owner only)
RemoveAuthorizedProposer(ctx, companyID, owner, proposerToRemove) error
```

### Transfer Ownership

```go
// Step 1: Current owner initiates
InitiateOwnershipTransfer(ctx, companyID, currentOwner, newOwner) error

// Step 2: New owner accepts
AcceptOwnershipTransfer(ctx, companyID, newOwner) error

// Cancel pending transfer (current owner only)
CancelOwnershipTransfer(ctx, companyID, currentOwner) error
```

## Message Types

### MsgAddAuthorizedProposer

Adds a new authorized proposer to a company.

**Signer**: Company owner only

```json
{
  "creator": "sharehodl1owner...",
  "company_id": 1,
  "new_proposer": "sharehodl1cofounder..."
}
```

**Errors**:
- `ErrNotCompanyOwner`: Signer is not the company owner
- `ErrProposerAlreadyExists`: Address is already a proposer
- `ErrCompanyNotFound`: Company does not exist

### MsgRemoveAuthorizedProposer

Removes an authorized proposer from a company.

**Signer**: Company owner only

```json
{
  "creator": "sharehodl1owner...",
  "company_id": 1,
  "proposer_to_remove": "sharehodl1cofounder..."
}
```

**Errors**:
- `ErrNotCompanyOwner`: Signer is not the company owner
- `ErrProposerNotFound`: Address is not a proposer
- `ErrCannotRemoveSelf`: Cannot remove owner (they implicitly have privileges)

### MsgInitiateOwnershipTransfer

Starts a two-step ownership transfer.

**Signer**: Current company owner only

```json
{
  "creator": "sharehodl1oldowner...",
  "company_id": 1,
  "new_owner": "sharehodl1newowner..."
}
```

**Errors**:
- `ErrNotCompanyOwner`: Signer is not the current owner
- `ErrOwnershipTransferPending`: Another transfer is already pending

### MsgAcceptOwnershipTransfer

Completes an ownership transfer (must be called by pending owner).

**Signer**: Pending new owner only

```json
{
  "creator": "sharehodl1newowner...",
  "company_id": 1
}
```

**Errors**:
- `ErrNoOwnershipTransferPending`: No transfer is pending
- `ErrUnauthorized`: Signer is not the pending owner

### MsgCancelOwnershipTransfer

Cancels a pending ownership transfer.

**Signer**: Current company owner only

```json
{
  "creator": "sharehodl1owner...",
  "company_id": 1
}
```

**Errors**:
- `ErrNotCompanyOwner`: Signer is not the current owner
- `ErrNoOwnershipTransferPending`: No transfer is pending

## Events

### proposer_added
```json
{
  "type": "proposer_added",
  "attributes": [
    {"key": "company_id", "value": "1"},
    {"key": "proposer", "value": "sharehodl1cofounder..."},
    {"key": "proposer_count", "value": "2"}
  ]
}
```

### proposer_removed
```json
{
  "type": "proposer_removed",
  "attributes": [
    {"key": "company_id", "value": "1"},
    {"key": "proposer", "value": "sharehodl1cofounder..."},
    {"key": "proposer_count", "value": "1"}
  ]
}
```

### ownership_transfer_initiated
```json
{
  "type": "ownership_transfer_initiated",
  "attributes": [
    {"key": "company_id", "value": "1"},
    {"key": "old_owner", "value": "sharehodl1founder..."},
    {"key": "pending_owner", "value": "sharehodl1ceo..."}
  ]
}
```

### ownership_transfer_accepted
```json
{
  "type": "ownership_transfer_accepted",
  "attributes": [
    {"key": "company_id", "value": "1"},
    {"key": "old_owner", "value": "sharehodl1founder..."},
    {"key": "new_owner", "value": "sharehodl1ceo..."}
  ]
}
```

### ownership_transfer_cancelled
```json
{
  "type": "ownership_transfer_cancelled",
  "attributes": [
    {"key": "company_id", "value": "1"},
    {"key": "old_owner", "value": "sharehodl1founder..."},
    {"key": "pending_owner", "value": "sharehodl1ceo..."}
  ]
}
```

## Integration Points

The authorization system automatically applies to:

### Treasury Withdrawals
`CanRequestWithdrawal` now checks `CanProposeForCompany` instead of just checking founder.

```go
// Before: Only founder could request
if company.Founder != requestor { return false }

// After: Owner OR proposer can request
return k.CanProposeForCompany(ctx, companyID, requestor)
```

### Primary Sales
`CreatePrimarySaleOffer` now checks `CanProposeForCompany`.

```go
// Before: Only founder could create
if company.Founder != creator { return error }

// After: Owner OR proposer can create
if !k.CanProposeForCompany(ctx, companyID, creator) {
    return ErrNotAuthorizedProposer
}
```

### Share Dilution
`ProposeShareDilution` now checks `CanProposeForCompany`.

```go
// Before: Only founder could propose
if company.Founder != msg.Creator { return error }

// After: Owner OR proposer can propose
if !k.CanProposeForCompany(ctx, msg.CompanyID, msg.Creator) {
    return ErrNotAuthorizedProposer
}
```

## Backwards Compatibility

The system maintains backwards compatibility:

1. **Existing companies without authorization records**: Falls back to checking `company.Founder`
2. **Automatic initialization**: New companies get authorization records automatically in `CreateCompany`
3. **Migration path**: Existing companies can have authorization records created retroactively

## Security Considerations

### Two-Step Ownership Transfer

Ownership transfer requires two transactions to prevent accidents:

1. Current owner initiates: Sets `PendingOwnerTransfer`
2. New owner accepts: Completes the transfer

This prevents:
- Typos in address causing loss of control
- Unwilling recipients becoming owners
- Front-running attacks

### Role Separation

- **Owner role**: Cannot be removed, only transferred
- **Proposer role**: Can be added/removed by owner
- **Clear hierarchy**: Owner has all proposer privileges + management rights

### Access Control

All authorization-changing operations:
- Verify signer is current owner
- Validate addresses before adding
- Emit events for transparency
- Cannot create circular references

## Error Codes

| Code | Error | Description |
|------|-------|-------------|
| 240 | `ErrNotCompanyOwner` | Signer is not the company owner |
| 241 | `ErrNotAuthorizedProposer` | Signer cannot propose for this company |
| 242 | `ErrProposerAlreadyExists` | Address is already an authorized proposer |
| 243 | `ErrProposerNotFound` | Address is not in proposer list |
| 244 | `ErrOwnershipTransferPending` | Another transfer is pending |
| 245 | `ErrNoOwnershipTransferPending` | No transfer to accept/cancel |
| 246 | `ErrCannotRemoveSelf` | Owner cannot be removed as proposer |

## Best Practices

### For Founders

1. **Add co-founders early**: Add authorized proposers when company is created
2. **Document roles**: Keep off-chain records of who has what authority
3. **Test transfers**: Use testnet to practice ownership transfers before mainnet

### For Co-Founders

1. **Accept limitations**: Understand proposers cannot modify authorization
2. **Coordinate proposals**: Avoid duplicate treasury withdrawal requests
3. **Communication**: Ensure owner is aware of proposals

### For Developers

1. **Use helper functions**: Always use `CanProposeForCompany` instead of checking founder
2. **Handle fallback**: Account for companies without authorization records
3. **Emit events**: Log all authorization changes for auditability

## Future Enhancements

Potential extensions (not yet implemented):

1. **Multi-sig ownership**: Require M-of-N signatures for owner actions
2. **Time-locked transfers**: Delay period before transfer completes
3. **Role hierarchy**: Additional roles (CFO, CTO) with specific permissions
4. **Expiring proposers**: Time-limited proposer roles
5. **Proposal limits**: Rate limiting per proposer
6. **Audit log**: On-chain record of all authorization changes
