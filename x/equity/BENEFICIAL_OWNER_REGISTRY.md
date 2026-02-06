# Beneficial Owner Registry

## Overview

The Beneficial Owner Registry tracks the **true owner** of shares when they are held in module accounts (escrow, lending, DEX). This ensures that dividends are paid to the correct owners even when shares are locked in various blockchain modules.

## Problem It Solves

Without this system:
- Alice puts 1000 shares in escrow → Escrow module now "owns" the shares
- Company declares dividend → Dividend goes to escrow module (wrong!)
- Alice doesn't receive her dividend even though she's the true owner

With this system:
- Alice puts 1000 shares in escrow → Escrow module calls `RegisterBeneficialOwner`
- Company declares dividend → System checks beneficial owner registry
- Alice receives dividend for her 1000 shares (correct!)

## Architecture

### Data Structure

```go
type BeneficialOwnership struct {
    ModuleAccount   string    // Module holding shares: "escrow", "lending", "dex"
    CompanyID       uint64    // Company ID
    ClassID         string    // Share class ID
    BeneficialOwner string    // TRUE owner (who should receive dividends)
    Shares          math.Int  // Number of shares locked
    LockedAt        time.Time // When shares were locked
    ReferenceID     uint64    // ID in locking module (escrow_id, loan_id, order_id)
    ReferenceType   string    // "escrow", "loan", or "order"
}
```

### Core Functions

#### For Module Integration

```go
// Call when locking shares (escrow created, loan taken, order placed)
RegisterBeneficialOwner(
    ctx sdk.Context,
    moduleAccount string,    // "escrow", "lending", "dex"
    companyID uint64,
    classID string,
    beneficialOwner string,  // User who still owns the shares
    shares math.Int,
    refID uint64,            // Your module's internal ID
    refType string,          // "escrow", "loan", "order"
) error

// Call when releasing shares (escrow completed, loan repaid, order filled)
UnregisterBeneficialOwner(
    ctx sdk.Context,
    moduleAccount string,
    companyID uint64,
    classID string,
    beneficialOwner string,
    refID uint64,
) error

// Call on partial updates (partial order fill, partial loan repayment)
UpdateBeneficialOwnerShares(
    ctx sdk.Context,
    moduleAccount string,
    companyID uint64,
    classID string,
    beneficialOwner string,
    refID uint64,
    newShares math.Int,     // Updated share count
) error
```

#### For Dividend Distribution

```go
// Called by dividend system to get ALL owners (direct + beneficial)
GetBeneficialOwnersForDividend(
    ctx sdk.Context,
    companyID uint64,
    classID string,
) []DividendRecipient
```

## Integration Guide

### For Escrow Module

```go
// When creating escrow
func (k Keeper) CreateEscrow(ctx sdk.Context, seller string, shares math.Int, ...) (uint64, error) {
    escrowID := k.GetNextEscrowID(ctx)

    // Transfer shares to escrow module
    // ... transfer logic ...

    // Register beneficial owner
    if err := k.equityKeeper.RegisterBeneficialOwner(
        ctx,
        "escrow",               // module name
        companyID,
        classID,
        seller,                 // seller is still the beneficial owner
        shares,
        escrowID,               // reference to this escrow
        "escrow",               // reference type
    ); err != nil {
        return 0, err
    }

    return escrowID, nil
}

// When completing escrow
func (k Keeper) CompleteEscrow(ctx sdk.Context, escrowID uint64) error {
    escrow := k.GetEscrow(ctx, escrowID)

    // Transfer shares to buyer
    // ... transfer logic ...

    // Unregister beneficial owner (seller no longer owns the shares)
    if err := k.equityKeeper.UnregisterBeneficialOwner(
        ctx,
        "escrow",
        escrow.CompanyID,
        escrow.ClassID,
        escrow.Seller,
        escrowID,
    ); err != nil {
        return err
    }

    return nil
}
```

### For Lending Module

```go
// When creating loan with share collateral
func (k Keeper) CreateLoan(ctx sdk.Context, borrower string, collateralShares math.Int, ...) (uint64, error) {
    loanID := k.GetNextLoanID(ctx)

    // Lock collateral shares
    // ... transfer to lending module ...

    // Register beneficial owner
    if err := k.equityKeeper.RegisterBeneficialOwner(
        ctx,
        "lending",
        companyID,
        classID,
        borrower,               // borrower still owns the shares
        collateralShares,
        loanID,
        "loan",
    ); err != nil {
        return 0, err
    }

    return loanID, nil
}

// When repaying loan
func (k Keeper) RepayLoan(ctx sdk.Context, loanID uint64) error {
    loan := k.GetLoan(ctx, loanID)

    // Return collateral to borrower
    // ... transfer logic ...

    // Unregister beneficial owner
    if err := k.equityKeeper.UnregisterBeneficialOwner(
        ctx,
        "lending",
        loan.CompanyID,
        loan.ClassID,
        loan.Borrower,
        loanID,
    ); err != nil {
        return err
    }

    return nil
}
```

### For DEX Module

```go
// When placing sell order
func (k Keeper) PlaceOrder(ctx sdk.Context, seller string, shares math.Int, ...) (uint64, error) {
    orderID := k.GetNextOrderID(ctx)

    // Lock shares in DEX
    // ... transfer to dex module ...

    // Register beneficial owner
    if err := k.equityKeeper.RegisterBeneficialOwner(
        ctx,
        "dex",
        companyID,
        classID,
        seller,                 // seller still owns the shares
        shares,
        orderID,
        "order",
    ); err != nil {
        return 0, err
    }

    return orderID, nil
}

// When order is partially filled
func (k Keeper) FillOrder(ctx sdk.Context, orderID uint64, fillAmount math.Int) error {
    order := k.GetOrder(ctx, orderID)
    remainingShares := order.Shares.Sub(fillAmount)

    // Transfer filled shares to buyer
    // ... transfer logic ...

    if remainingShares.IsZero() {
        // Order fully filled - unregister
        if err := k.equityKeeper.UnregisterBeneficialOwner(
            ctx,
            "dex",
            order.CompanyID,
            order.ClassID,
            order.Seller,
            orderID,
        ); err != nil {
            return err
        }
    } else {
        // Partial fill - update share count
        if err := k.equityKeeper.UpdateBeneficialOwnerShares(
            ctx,
            "dex",
            order.CompanyID,
            order.ClassID,
            order.Seller,
            orderID,
            remainingShares,    // new share count
        ); err != nil {
            return err
        }
    }

    return nil
}

// When cancelling order
func (k Keeper) CancelOrder(ctx sdk.Context, orderID uint64) error {
    order := k.GetOrder(ctx, orderID)

    // Return shares to seller
    // ... transfer logic ...

    // Unregister beneficial owner
    if err := k.equityKeeper.UnregisterBeneficialOwner(
        ctx,
        "dex",
        order.CompanyID,
        order.ClassID,
        order.Seller,
        orderID,
    ); err != nil {
        return err
    }

    return nil
}
```

## Dividend Distribution Flow

### Before (Without Beneficial Owner Registry)

```
1. Company has 10,000 shares outstanding
2. Alice owns 1,000 shares directly in wallet
3. Bob owns 2,000 shares, but 1,000 are in escrow
4. Dividend snapshot sees:
   - Alice: 1,000 shares
   - Bob: 1,000 shares (only wallet)
   - Escrow module: 1,000 shares
5. Dividend payment:
   - Alice gets 1,000 shares worth ✓
   - Bob gets 1,000 shares worth ✗ (should be 2,000)
   - Escrow module gets 1,000 shares worth ✗ (shouldn't get anything)
```

### After (With Beneficial Owner Registry)

```
1. Company has 10,000 shares outstanding
2. Alice owns 1,000 shares directly in wallet
3. Bob owns 2,000 shares:
   - 1,000 in wallet
   - 1,000 in escrow (registry records: Bob is beneficial owner)
4. Dividend snapshot uses GetBeneficialOwnersForDividend():
   - Alice: 1,000 shares (from wallet)
   - Bob: 2,000 shares (1,000 wallet + 1,000 from registry)
   - Escrow module: skipped (not a real owner)
5. Dividend payment:
   - Alice gets 1,000 shares worth ✓
   - Bob gets 2,000 shares worth ✓
   - Escrow module gets nothing ✓
```

## Testing Scenarios

### Scenario 1: Shares in Escrow
```
1. Alice owns 1,000 shares
2. Alice creates escrow for 500 shares
3. Registry records: escrow module holds 500 shares for Alice
4. Company declares dividend
5. Alice receives dividend for all 1,000 shares
```

### Scenario 2: Shares on DEX
```
1. Bob owns 2,000 shares
2. Bob places sell order for 1,000 shares
3. Registry records: dex module holds 1,000 shares for Bob
4. Company declares dividend
5. Bob receives dividend for all 2,000 shares
6. Order partially fills (300 shares sold)
7. Registry updated: dex now holds 700 shares for Bob
8. Next dividend: Bob receives dividend for 1,700 shares (1,000 wallet + 700 on DEX)
```

### Scenario 3: Shares as Loan Collateral
```
1. Charlie owns 3,000 shares
2. Charlie takes loan using 1,500 shares as collateral
3. Registry records: lending module holds 1,500 shares for Charlie
4. Company declares dividend
5. Charlie receives dividend for all 3,000 shares
6. Charlie repays loan
7. Registry unregisters the 1,500 shares
8. Next dividend: Charlie receives dividend for 3,000 shares (all in wallet now)
```

### Scenario 4: Multiple Concurrent Operations
```
1. Diana owns 5,000 shares
2. Diana creates escrow for 1,000 shares
3. Diana places DEX order for 2,000 shares
4. Diana uses 1,000 shares as loan collateral
5. Diana has 1,000 shares in wallet
6. Registry records:
   - escrow: 1,000 shares
   - dex: 2,000 shares
   - lending: 1,000 shares
7. Company declares dividend
8. Diana receives dividend for all 5,000 shares
```

## Storage Keys

```
BeneficialOwnerPrefix (0x50): module|companyID|classID|owner|refID -> BeneficialOwnership
BeneficialOwnerByModulePrefix (0x51): module|companyID|classID -> index
```

## Error Codes

- `ErrBeneficialOwnerNotFound (260)`: Record not found
- `ErrInvalidInput (261)`: Invalid parameter
- `ErrInternalError (262)`: Internal error

## Best Practices

1. **Always register** when locking shares in a module account
2. **Always unregister** when returning shares to owner
3. **Use UpdateBeneficialOwnerShares** for partial operations
4. **Use unique reference IDs** from your module (escrow_id, loan_id, order_id)
5. **Handle errors properly** - failed registration means user won't get dividends

## Security Considerations

1. Only the locking module should be able to register/unregister its own records
2. Validation ensures:
   - Shares > 0
   - Valid addresses
   - Valid reference types
3. Multiple registrations for same user are supported (different reference IDs)
4. Registry is authoritative for dividend distribution

## Future Enhancements

Potential future improvements:
1. Add expiry dates for stale records
2. Add query endpoints for users to see their beneficial ownership
3. Add events for registration/unregistration
4. Add admin tools to fix incorrect registrations
5. Support for other module types beyond escrow/lending/dex

## Contact

For questions or issues with the beneficial owner registry, contact the equity module maintainers.
