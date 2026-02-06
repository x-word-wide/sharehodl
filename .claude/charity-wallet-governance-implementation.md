# Charity Wallet Governance Implementation

## Overview
The protocol default charity wallet is now controlled by validator governance. Only validators can propose and vote to change the default charity wallet address used for redirecting dividends from blacklisted shareholders.

## Implementation Summary

### Files Modified

1. **x/governance/types/governance.go**
   - Added `ProposalTypeSetCharityWallet` (type 9)
   - Added "set_charity_wallet" to ProposalType String() method

2. **x/governance/keeper/keeper.go**
   - Added `SetDefaultCharityWallet(ctx, address)` to EquityKeeper interface
   - Added `GetDefaultCharityWallet(ctx)` to EquityKeeper interface
   - Added `ExecuteSetCharityWalletProposal(ctx, proposal)` method

3. **x/governance/keeper/tally.go**
   - Added `ProposalTypeSetCharityWallet` case to executeProposal() switch

4. **x/equity/keeper/keeper.go**
   - Added `SetDefaultCharityWallet(ctx, address)` method
   - Added `GetDefaultCharityWallet(ctx)` method

5. **x/equity/keeper/msg_server.go**
   - Updated `SetCharityWallet()` to require authority field
   - Added governance authority check
   - Added event emission for transparency

6. **x/equity/types/key.go**
   - Already has `DefaultCharityWalletKey` defined (0x64)

## Proposal Flow

### 1. Validator Submits Proposal

```bash
# Using CLI (example)
sharehodld tx governance submit-proposal \
  --type="set_charity_wallet" \
  --title="Update Default Charity Wallet to Red Cross" \
  --description="Proposal to update the protocol's default charity wallet to the Red Cross address for redirecting dividends from blacklisted shareholders." \
  --metadata='{"wallet":"sharehodl1charity123..."}' \
  --from=validator1 \
  --chain-id=sharehodl-1
```

### 2. Validators Vote

```bash
# Validators vote on the proposal
sharehodld tx governance vote <proposal-id> yes \
  --from=validator1 \
  --chain-id=sharehodl-1
```

### 3. Proposal Passes

- **Quorum**: 33.4% of validator voting power must participate
- **Threshold**: 50% of votes must be "yes"
- **Voting Period**: 7 days (configurable via governance)

### 4. Automatic Execution

When the proposal passes, `ExecuteSetCharityWalletProposal()` is called:

1. Extracts wallet address from proposal metadata
2. Validates address format (Bech32)
3. Calls `equityKeeper.SetDefaultCharityWallet(ctx, address)`
4. Emits `charity_wallet_updated` event

## Key Features

### Governance-Only Control
- Only governance module can change the default charity wallet
- Individual companies can still set their own charity wallet (overrides default)
- Validators control the protocol-level default

### Transparency
- All changes emit events for tracking
- Proposal ID linked to wallet changes
- Full audit trail of governance decisions

### Security
- Address validation before setting
- Atomic execution (success or revert)
- Error handling with detailed messages

## API Examples

### Submit Proposal (Programmatic)

```go
// Create proposal metadata
metadata := map[string]interface{}{
    "wallet": "sharehodl1charity123abc...",
    "description": "Update charity wallet to Red Cross",
}

// Submit proposal
proposalID, err := governanceKeeper.SubmitProposal(
    ctx,
    submitter,
    types.ProposalTypeSetCharityWallet,
    "Update Default Charity Wallet",
    "Proposal to update the default charity wallet for dividend redirection",
    initialDeposit,
    7, // 7 day voting period
    math.LegacyNewDecWithPrec(334, 3), // 33.4% quorum
    math.LegacyNewDecWithPrec(5, 1),   // 50% threshold
)
```

### Get Current Charity Wallet

```go
// Query current default charity wallet
wallet := equityKeeper.GetDefaultCharityWallet(ctx)
if wallet == "" {
    // No default set, companies must configure individually
}
```

## Event Schema

### charity_wallet_updated
```json
{
  "type": "charity_wallet_updated",
  "attributes": [
    {
      "key": "wallet",
      "value": "sharehodl1charity123abc..."
    },
    {
      "key": "proposal_id",
      "value": "42"
    }
  ]
}
```

### set_default_charity_wallet
```json
{
  "type": "set_default_charity_wallet",
  "attributes": [
    {
      "key": "wallet",
      "value": "sharehodl1charity123abc..."
    },
    {
      "key": "authority",
      "value": "sharehodl1gov..."
    }
  ]
}
```

## Integration with Dividend System

When a shareholder is blacklisted and a dividend is declared:

1. Check if company has custom charity wallet configured
2. If not, fall back to protocol default (`GetDefaultCharityWallet()`)
3. If neither set, use configured fallback action (hold, burn, etc.)

## Testing

### Unit Tests Needed
- [ ] Proposal submission with valid metadata
- [ ] Proposal submission with invalid wallet address
- [ ] Proposal execution updates wallet correctly
- [ ] GetDefaultCharityWallet returns correct value
- [ ] Only governance authority can call SetCharityWallet

### Integration Tests Needed
- [ ] Full proposal flow (submit -> vote -> execute)
- [ ] Dividend redirection uses default wallet
- [ ] Company-specific wallet overrides default
- [ ] Event emission verification

## Migration Notes

For existing chains upgrading to this feature:

1. No existing data migration needed
2. DefaultCharityWalletKey starts empty
3. First governance vote sets initial value
4. Backwards compatible with existing blacklist system

## Security Considerations

- Only governance module can change default charity wallet
- Individual companies maintain autonomy (can override default)
- Address validation prevents invalid addresses
- Events provide transparency for community oversight
- No direct user control (requires validator consensus)

## Future Enhancements

- Add charity wallet registry (verified charity addresses)
- Add proposal type for charity whitelist
- Add metrics for charity donations over time
- Add multi-sig support for charity wallets
