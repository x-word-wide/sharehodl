# GetAllHoldingsByAddress Implementation

## Overview
This document describes the implementation of the `GetAllHoldingsByAddress` method in the equity keeper, which enables the inheritance module to enumerate and transfer all equity holdings for a given address.

## Problem Statement
The inheritance module needed a way to retrieve all equity shareholdings for a deceased user's address in order to transfer them to beneficiaries according to the inheritance plan. Previously, there was no method to efficiently query all holdings for a single address across all companies and share classes.

## Solution

### 1. AddressHolding Type
A new struct type was added to `x/equity/types/equity.go`:

```go
type AddressHolding struct {
    CompanyID uint64   `json:"company_id"`
    ClassID   string   `json:"class_id"`
    Shares    math.Int `json:"shares"`
}
```

This struct provides a lightweight representation of a shareholding, containing only the essential information needed for transfer operations.

### 2. Equity Keeper Method
The method was added to `x/equity/keeper/keeper.go`:

```go
func (k Keeper) GetAllHoldingsByAddress(ctx sdk.Context, owner string) []interface{}
```

**Implementation Details:**
- Iterates through all shareholdings using the `ShareholdingPrefix`
- Filters shareholdings by the owner address
- Returns holdings as `[]interface{}` to avoid import cycles
- Handles JSON unmarshaling errors gracefully with logging

**Storage Iteration:**
The method uses the existing storage structure where shareholdings are keyed as:
```
ShareholdingPrefix + CompanyID (8 bytes) + ClassID + "|" + Owner
```

By iterating through all keys with the `ShareholdingPrefix`, the method can find all holdings regardless of company or share class.

### 3. Expected Keeper Interface
The method signature was added to `x/inheritance/types/expected_keepers.go`:

```go
type EquityKeeper interface {
    // ... existing methods ...

    // GetAllHoldingsByAddress returns all shareholdings for a given address
    GetAllHoldingsByAddress(ctx sdk.Context, owner string) []interface{}
}
```

**Design Decision: Why `[]interface{}`?**
- Avoids circular import dependencies between equity and inheritance modules
- Allows the inheritance module to work with the concrete type through type assertions
- Maintains loose coupling between modules

### 4. Inheritance Integration
Updated `x/inheritance/keeper/asset_handler.go` to use the new method:

```go
func (k Keeper) transferEquityShares(ctx sdk.Context, owner, beneficiary string, percentage math.LegacyDec) ([]types.TransferredAsset, error)
```

**Transfer Logic:**
1. Call `GetAllHoldingsByAddress` to get all holdings
2. Type assert each holding to extract `CompanyID`, `ClassID`, and `Shares`
3. Calculate transfer amount based on beneficiary's percentage
4. Call `TransferShares` for each holding
5. Record each successful transfer
6. Log errors for failed transfers but continue processing

**Error Handling:**
- Individual transfer failures don't stop the entire inheritance process
- All errors are logged for audit purposes
- Successfully transferred assets are tracked separately

## Performance Considerations

### Time Complexity
- O(n) where n is the total number of shareholdings in the system
- Requires full iteration through the shareholding store

### Optimization Opportunities
If performance becomes an issue with large numbers of shareholdings, consider:

1. **Secondary Index:** Create an owner-to-holdings index
   ```
   OwnerHoldingsPrefix + Owner -> []CompanyID + []ClassID
   ```

2. **Paginated Queries:** Add pagination support for addresses with many holdings

3. **Caching:** Cache holdings during inheritance processing to avoid repeated queries

## Usage Example

```go
// In inheritance module
holdings := k.equityKeeper.GetAllHoldingsByAddress(ctx, deceasedAddress)

for _, holdingInterface := range holdings {
    if h, ok := holdingInterface.(AddressHolding); ok {
        shareAmount := beneficiaryPercent.MulInt(h.Shares).TruncateInt()

        err := k.equityKeeper.TransferShares(
            ctx,
            h.CompanyID,
            h.ClassID,
            deceasedAddress,
            beneficiaryAddress,
            shareAmount,
        )

        if err != nil {
            // Log and continue
        }
    }
}
```

## Testing

### Unit Tests
Created `x/equity/keeper/address_holdings_test.go` with test cases for:
- Single holding retrieval
- Multiple holdings across companies
- Multiple share classes in same company
- Empty results for addresses with no holdings

### Integration Testing
The method is tested as part of the inheritance flow in the end-to-end tests.

## Security Considerations

1. **Privacy:** The method reveals all holdings for an address
   - Only called during legitimate inheritance processing
   - Access controlled through keeper boundaries

2. **Gas Costs:** Iteration can be expensive
   - Consider gas limits for addresses with many holdings
   - May need pagination in future

3. **Data Integrity:** Type assertions must be handled safely
   - Fallback logic for failed type assertions
   - Comprehensive error logging

## Future Enhancements

1. Add pagination support
2. Create owner-to-holdings secondary index
3. Add filtering by company or share class
4. Optimize with caching layer
5. Add metrics for monitoring iteration performance

## Related Files

- `x/equity/types/equity.go` - AddressHolding struct
- `x/equity/keeper/keeper.go` - GetAllHoldingsByAddress implementation
- `x/inheritance/types/expected_keepers.go` - Interface definition
- `x/inheritance/keeper/asset_handler.go` - Usage in transfers
- `x/equity/keeper/address_holdings_test.go` - Unit tests

## Migration Notes

**Breaking Changes:** None
- New method added to interface
- Existing code unaffected
- Backward compatible

**Deployment:**
- No state migration required
- Can be deployed without chain restart
- No genesis changes needed
