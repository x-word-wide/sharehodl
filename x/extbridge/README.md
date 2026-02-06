# External Bridge Module (extbridge)

The `extbridge` module enables ShareHODL blockchain to bridge with external blockchains (Bitcoin, Ethereum, etc.) allowing users to deposit external assets (USDT, BTC, ETH) and receive HODL, or burn HODL to withdraw external assets.

## Overview

The external bridge is a **validator-attested bridge** with the following security features:

- **Multi-Validator Attestation**: Requires 2/3+ of validators (Archon+ tier) to attest deposits
- **TSS (Threshold Signature Scheme)**: Withdrawals are signed by a threshold of validators
- **Timelock System**: Withdrawals have a mandatory delay before execution
- **Rate Limiting**: Per-asset daily withdrawal limits
- **Circuit Breaker**: Emergency pause mechanism for security incidents

## Architecture

```
External Chain (BTC/ETH/etc)
         │
         ├─ Deposit Transaction
         │
         ▼
   Validator Observers
         │
         ├─ Submit Observations
         │
         ▼
   Attestation Pool (2/3 threshold)
         │
         ├─ Mint HODL
         │
         ▼
      User Wallet

User Requests Withdrawal
         │
         ├─ Burn HODL
         │
         ▼
   Timelock Period (1 hour)
         │
         ├─ TSS Signing Session
         │
         ▼
   Broadcast to External Chain
```

## Key Components

### 1. Types

#### External Chain Configuration
- `ExternalChain`: Configuration for supported blockchains (Bitcoin, Ethereum, etc.)
- `ExternalAsset`: Configuration for supported assets (USDT, BTC, ETH, etc.)

#### Deposit Flow
- `Deposit`: Records external chain deposits
- `DepositAttestation`: Validator attestations for deposits
- `DepositStatus`: pending → attesting → completed/rejected

#### Withdrawal Flow
- `Withdrawal`: Records withdrawal requests
- `WithdrawalStatus`: pending → timelocked → ready → signing → signed → broadcast → completed
- `TSSSession`: Threshold signing session for withdrawals
- `TSSSignatureShare`: Individual validator signature shares

#### Security
- `RateLimit`: Per-asset daily withdrawal limits
- `CircuitBreaker`: Emergency pause mechanism
- `AnomalyDetection`: Suspicious activity tracking

### 2. Keeper Functions

#### Deposit Processing
- `ObserveDeposit()`: Validator observes external chain deposit
- `AttestDeposit()`: Validator attests to deposit validity
- `CompleteDeposit()`: Mint HODL after threshold reached

#### Withdrawal Processing
- `RequestWithdrawal()`: User requests withdrawal (burns HODL)
- `ProcessWithdrawals()`: Check timelock expiry (runs in EndBlock)
- `CreateTSSSession()`: Create signing session for ready withdrawals

#### TSS Coordination
- `SubmitTSSSignature()`: Validator submits signature share
- `CombineSignatures()`: Combine shares into final signature

#### Security
- `CheckRateLimit()`: Verify withdrawal within limits
- `IsOperationAllowed()`: Check circuit breaker status
- `IsValidatorEligible()`: Check validator tier (Archon+ required)

### 3. Messages

#### User Messages
- `MsgRequestWithdrawal`: Request withdrawal to external chain

#### Validator Messages
- `MsgObserveDeposit`: Submit deposit observation
- `MsgAttestDeposit`: Attest to deposit validity
- `MsgSubmitTSSSignature`: Submit TSS signature share

#### Governance Messages
- `MsgAddExternalChain`: Add new external blockchain
- `MsgAddExternalAsset`: Add new external asset
- `MsgUpdateCircuitBreaker`: Emergency pause/unpause

## Usage Examples

### User: Request Withdrawal

```bash
sharehodld tx extbridge request-withdrawal \
  --chain-id ethereum-1 \
  --asset USDT \
  --recipient 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb \
  --amount 1000000000uhodl \
  --from alice
```

### Validator: Observe Deposit

```bash
sharehodld tx extbridge observe-deposit \
  --chain-id ethereum-1 \
  --asset USDT \
  --tx-hash 0xabc123... \
  --block-height 15234567 \
  --sender 0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb \
  --recipient sharehodl1abc123... \
  --amount 1000000000 \
  --from validator1
```

### Validator: Attest Deposit

```bash
sharehodld tx extbridge attest-deposit \
  --deposit-id 42 \
  --approved true \
  --from validator1
```

### Validator: Submit TSS Signature

```bash
sharehodld tx extbridge submit-tss-signature \
  --session-id 10 \
  --signature-data <hex_encoded_signature> \
  --from validator1
```

### Governance: Add External Chain

```bash
sharehodld tx extbridge add-external-chain \
  --chain-id ethereum-1 \
  --chain-name "Ethereum Mainnet" \
  --chain-type evm \
  --confirmations 12 \
  --block-time 12 \
  --tss-address 0x... \
  --min-deposit 1000000 \
  --max-deposit 1000000000000 \
  --from gov
```

### Governance: Add External Asset

```bash
sharehodld tx extbridge add-external-asset \
  --chain-id ethereum-1 \
  --asset-symbol USDT \
  --asset-name "Tether USD" \
  --contract-address 0xdac17f958d2ee523a2206206994597c13d831ec7 \
  --decimals 6 \
  --conversion-rate 1.0 \
  --daily-limit 10000000000000 \
  --per-tx-limit 1000000000000 \
  --from gov
```

## Security Features

### 1. Multi-Validator Attestation
- Deposits require 2/3+ validator attestations (configurable via params)
- Only validators with Archon+ tier (10M+ HODL staked) can attest
- Prevents single point of failure or malicious validator attacks

### 2. TSS Signing
- Withdrawals signed by threshold of validators (2/3+)
- No single validator can sign withdrawal alone
- Private key never exists in complete form

### 3. Timelock
- All withdrawals have mandatory 1-hour delay (configurable)
- Allows time to detect and halt malicious withdrawals
- Can be cancelled during timelock period if fraud detected

### 4. Rate Limiting
- Per-asset daily withdrawal limits
- Prevents rapid draining of bridge reserves
- Configurable via governance

### 5. Circuit Breaker
- Emergency pause mechanism
- Can disable deposits, withdrawals, or attestations independently
- Triggered manually by governance or automatically on anomaly detection

### 6. Amount Limits
- Per-transaction min/max limits per chain
- Daily limits per asset
- Protects against large unexpected withdrawals

## Parameters

```go
type Params struct {
    BridgingEnabled          bool    // Master switch
    AttestationThreshold     Dec     // 0.67 = 2/3
    MinValidatorTier         int32   // 4 = Archon
    WithdrawalTimelock       uint64  // 3600 seconds (1 hour)
    RateLimitWindow          uint64  // 86400 seconds (24 hours)
    MaxWithdrawalPerWindow   Int     // 1M HODL
    BridgeFee                Dec     // 0.001 = 0.1%
    TSSThreshold             Dec     // 0.67 = 2/3
    EmergencyPauseEnabled    bool    // Circuit breaker
}
```

## Events

### Deposit Events
- `extbridge_deposit_observed`: New deposit observed
- `extbridge_deposit_attested`: Validator attested
- `extbridge_deposit_completed`: Deposit completed, HODL minted

### Withdrawal Events
- `extbridge_withdrawal_requested`: Withdrawal requested
- `extbridge_withdrawal_ready`: Timelock expired
- `extbridge_withdrawal_completed`: Withdrawal signed and broadcast

### TSS Events
- `extbridge_tss_session_created`: TSS session started
- `extbridge_tss_signature_submitted`: Validator submitted signature

### Security Events
- `extbridge_circuit_breaker_updated`: Circuit breaker changed
- `extbridge_rate_limit_exceeded`: Rate limit violation

## Module Integration

The extbridge module is wired into the ShareHODL app:

1. **Module Account**: Has minter and burner permissions
2. **Keeper Dependencies**:
   - `BankKeeper`: Mint/burn HODL
   - `AccountKeeper`: Manage accounts
   - `StakingKeeper`: Check validator tiers
3. **Begin/End Block**: Processes timelocked withdrawals
4. **Governance**: Parameter updates and circuit breaker control

## Development Status

- [x] Core types implemented
- [x] Keeper structure implemented
- [x] Deposit flow (observe, attest, complete)
- [x] Withdrawal flow (request, timelock, ready)
- [x] TSS coordination (create session, submit signatures)
- [x] Security features (rate limits, circuit breaker)
- [x] Message handlers
- [x] Module wiring in app.go
- [ ] Query endpoints (TODO)
- [ ] CLI commands (TODO)
- [ ] External chain observers (TODO)
- [ ] TSS implementation (TODO - currently placeholder)
- [ ] Integration tests (TODO)
- [ ] Security audit (TODO)

## Future Enhancements

1. **Multi-Chain Support**: Add more chains (Solana, Polygon, BSC)
2. **Asset Variety**: Support more assets per chain
3. **Dynamic Conversion Rates**: Oracle-based pricing
4. **Liquidity Pools**: Reserve management
5. **Fast Withdrawals**: Reduced timelock for small amounts
6. **Batch Processing**: Group withdrawals for efficiency

## Security Considerations

### Before Production
1. **Security Audit**: Full audit by reputable firm
2. **TSS Implementation**: Complete TSS library integration
3. **External Observers**: Reliable external chain monitoring
4. **Anomaly Detection**: Machine learning-based fraud detection
5. **Insurance Fund**: Reserve pool for potential losses
6. **Gradual Rollout**: Start with low limits, increase over time

### Operational Security
1. **Validator Diversity**: Geographic and infrastructure diversity
2. **Key Rotation**: Regular TSS key rotation
3. **Monitoring**: 24/7 monitoring of bridge activity
4. **Incident Response**: Documented response procedures
5. **Regular Audits**: Periodic security audits

## License

Copyright (c) ShareHODL 2025
