# Shareholder Blacklist for Dividends - Implementation Guide

## Overview

This implementation adds shareholder blacklist functionality to the equity module, allowing companies to prevent specific shareholders from receiving dividends. When a blacklisted shareholder would receive a dividend, the funds are automatically redirected to a designated charity wallet or other fallback destination.

## Features

1. **Shareholder Blacklisting**: Company owners can blacklist shareholders from receiving dividends
2. **Temporary or Permanent**: Blacklists can expire after a duration or be permanent
3. **Dividend Redirection**: Blocked dividends automatically redirect to charity
4. **Configurable Fallback**: Multiple fallback options if charity wallet unavailable
5. **Audit Trail**: All blacklist actions and redirections are tracked via events

## Architecture

### Data Structures

#### ShareholderBlacklist
Tracks blacklisted addresses per company.

```go
type ShareholderBlacklist struct {
    CompanyID     uint64
    Address       string
    Reason        string
    BlacklistedBy string
    BlacklistedAt time.Time
    ExpiresAt     time.Time  // Zero value = permanent
    IsActive      bool
}
```

#### DividendRedirection
Defines where blocked dividends are sent.

```go
type DividendRedirection struct {
    CompanyID       uint64
    CharityWallet   string
    FallbackAction  string  // "community_pool", "burn", "pro_rata"
    Description     string
    SetBy           string
    SetAt           time.Time
    UpdatedAt       time.Time
}
```

### Store Keys

- `ShareholderBlacklistPrefix (0x62)`: company_id + address -> ShareholderBlacklist
- `DividendRedirectionPrefix (0x63)`: company_id -> DividendRedirection
- `DefaultCharityWalletKey (0x64)`: global default charity wallet

### Keeper Methods

#### Blacklist Management

```go
// Add shareholder to blacklist
BlacklistShareholder(ctx, companyID, address, reason, blacklistedBy, duration) error

// Remove shareholder from blacklist
UnblacklistShareholder(ctx, companyID, address) error

// Check if shareholder is blacklisted (auto-removes expired)
IsShareholderBlacklisted(ctx, companyID, address) bool

// Get blacklist entry
GetShareholderBlacklist(ctx, companyID, address) (ShareholderBlacklist, bool)

// Get all blacklisted shareholders for a company
GetBlacklistedShareholders(ctx, companyID) []ShareholderBlacklist
```

#### Redirection Configuration

```go
// Set company-specific dividend redirection
SetDividendRedirection(ctx, redirect) error

// Get dividend redirection config
GetDividendRedirection(ctx, companyID) (DividendRedirection, bool)

// Set protocol-level default charity wallet (governance only)
SetDefaultCharityWallet(ctx, address) error

// Get default charity wallet
GetDefaultCharityWallet(ctx) string
```

## Dividend Processing Flow

When processing dividend payments, the system now:

1. **Check Blacklist Status**: For each shareholder, check if blacklisted
2. **Determine Recipient**:
   - If NOT blacklisted: Send to original shareholder
   - If blacklisted: Redirect based on configuration
3. **Redirection Priority**:
   - Company-specific charity wallet (if configured)
   - Protocol default charity wallet (if set)
   - Community pool (final fallback)
4. **Payment Execution**:
   - Cash dividends: Redirect payment to determined recipient
   - Stock dividends: Skip payment (stock can't be redirected to charity)
5. **Record Keeping**: Create payment record with status:
   - `"paid"`: Normal payment
   - `"redirected"`: Payment sent to charity (blacklisted shareholder)
   - `"skipped"`: Stock dividend for blacklisted shareholder
6. **Event Emission**: Emit `dividend_redirected` event for audit trail

### Code Integration

The blacklist check is integrated into `ProcessDividendPayments` in `/x/equity/keeper/dividend.go`:

```go
// Check if shareholder is blacklisted
if k.IsShareholderBlacklisted(ctx, dividend.CompanyID, shareholderSnapshot.Shareholder) {
    // Get redirection settings
    redirect, hasRedirect := k.GetDividendRedirection(ctx, dividend.CompanyID)

    if hasRedirect && redirect.CharityWallet != "" {
        recipientAddrStr = redirect.CharityWallet
    } else {
        recipientAddrStr = k.GetDefaultCharityWallet(ctx)
        if recipientAddrStr == "" {
            recipientAddrStr = k.GetCommunityPoolAddress(ctx)
        }
    }

    isRedirected = true
    payment.Status = "redirected"
}
```

## Message Types

### BlacklistShareholder

Adds a shareholder to the blacklist.

```go
type SimpleMsgBlacklistShareholder struct {
    Authority string  // Company owner or governance
    CompanyID uint64
    Address   string
    Reason    string
    Duration  int64   // Seconds, 0 = permanent
}
```

**Authorization**: Only company founder can blacklist shareholders

### UnblacklistShareholder

Removes a shareholder from the blacklist.

```go
type SimpleMsgUnblacklistShareholder struct {
    Authority string
    CompanyID uint64
    Address   string
}
```

**Authorization**: Only company founder can unblacklist shareholders

### SetDividendRedirection

Configures where blocked dividends are sent.

```go
type SimpleMsgSetDividendRedirection struct {
    Authority      string
    CompanyID      uint64
    CharityWallet  string
    FallbackAction string  // "community_pool", "burn", "pro_rata"
    Description    string
}
```

**Authorization**: Only company founder can configure redirection

### SetCharityWallet

Sets the protocol-level default charity wallet.

```go
type SimpleMsgSetCharityWallet struct {
    Authority string  // Governance only
    Wallet    string
}
```

**Authorization**: Governance module only (TODO: implement governance check)

## Events

### EventTypeShareholderBlacklisted

Emitted when a shareholder is added to the blacklist.

**Attributes**:
- `company_id`: Company ID
- `shareholder`: Blacklisted address
- `reason`: Reason for blacklisting
- `blacklisted_by`: Who performed the blacklist

### EventTypeShareholderUnblacklisted

Emitted when a shareholder is removed from the blacklist.

**Attributes**:
- `company_id`: Company ID
- `shareholder`: Unblacklisted address

### EventTypeDividendRedirected

Emitted when a dividend is redirected due to blacklist.

**Attributes**:
- `company_id`: Company ID
- `shareholder`: Original shareholder (blacklisted)
- `redirect_to`: Recipient address (charity)
- `redirect_amount`: Amount redirected
- `redirect_reason`: "shareholder_blacklisted"

## Error Codes

- `ErrShareholderBlacklisted (250)`: Shareholder is blacklisted
- `ErrNotBlacklisted (251)`: Shareholder is not blacklisted
- `ErrInvalidCharityWallet (252)`: Invalid charity wallet address
- `ErrNoRedirectionConfigured (253)`: No dividend redirection configured
- `ErrBlacklistExpired (254)`: Blacklist entry has expired
- `ErrBlacklistAlreadyExists (255)`: Blacklist entry already exists

## Usage Examples

### Example 1: Blacklist a Shareholder (Permanent)

```bash
sharehodld tx equity blacklist-shareholder \
  --authority sharehodl1company_founder... \
  --company-id 1 \
  --address sharehodl1bad_actor... \
  --reason "Fraudulent activity detected" \
  --duration 0 \
  --from company_founder
```

### Example 2: Blacklist a Shareholder (Temporary - 30 days)

```bash
sharehodld tx equity blacklist-shareholder \
  --authority sharehodl1company_founder... \
  --company-id 1 \
  --address sharehodl1suspended_shareholder... \
  --reason "Pending investigation" \
  --duration 2592000 \
  --from company_founder
```

### Example 3: Configure Dividend Redirection

```bash
sharehodld tx equity set-dividend-redirection \
  --authority sharehodl1company_founder... \
  --company-id 1 \
  --charity-wallet sharehodl1charity_org... \
  --fallback-action community_pool \
  --description "Red Cross donation" \
  --from company_founder
```

### Example 4: Set Default Charity Wallet (Governance)

```bash
sharehodld tx equity set-charity-wallet \
  --authority sharehodl1gov_module... \
  --wallet sharehodl1default_charity... \
  --from governance
```

### Example 5: Unblacklist a Shareholder

```bash
sharehodld tx equity unblacklist-shareholder \
  --authority sharehodl1company_founder... \
  --company-id 1 \
  --address sharehodl1reformed_shareholder... \
  --from company_founder
```

## Security Considerations

1. **Authorization**: Only company founders can manage blacklists for their companies
2. **Governance Override**: Future enhancement - allow governance to override blacklists
3. **Audit Trail**: All blacklist actions emit events for transparency
4. **Auto-Expiration**: Temporary blacklists automatically expire and are removed
5. **Validation**: All addresses are validated before blacklisting

## Future Enhancements

1. **Governance Control**: Allow governance to blacklist/unblacklist shareholders
2. **Multi-Signature**: Require multiple approvals for blacklisting
3. **Appeal Process**: Implement shareholder appeal mechanism
4. **Batch Operations**: Blacklist multiple shareholders in one transaction
5. **Blacklist Reasons**: Predefined reason codes with severity levels
6. **Notification System**: Notify shareholders when blacklisted
7. **Pro-Rata Redistribution**: Redistribute blocked dividends to other shareholders
8. **Burn Mechanism**: Implement token burning for blocked dividends

## Testing

Run unit tests:

```bash
cd x/equity
go test -v ./keeper -run TestShareholderBlacklist
go test -v ./keeper -run TestBlacklistExpiration
```

Integration tests should verify:
- Blacklist creation and removal
- Dividend redirection flow
- Expiration handling
- Event emission
- Authorization checks

## Files Modified

### New Files
- `/x/equity/types/blacklist.go` - Blacklist data structures and constructors
- `/x/equity/keeper/blacklist.go` - Blacklist keeper methods
- `/x/equity/keeper/blacklist_test.go` - Unit tests

### Modified Files
- `/x/equity/types/errors.go` - Added blacklist-related errors
- `/x/equity/types/key.go` - Added store keys and key functions
- `/x/equity/types/msg.go` - Added message types and validation
- `/x/equity/keeper/msg_server.go` - Added message handlers
- `/x/equity/keeper/dividend.go` - Integrated blacklist check in dividend processing

## API Reference

See inline code documentation in:
- `/x/equity/types/blacklist.go`
- `/x/equity/keeper/blacklist.go`

## Support

For questions or issues:
- Review inline code comments
- Check test cases for usage examples
- Consult the ShareHODL documentation
