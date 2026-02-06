# ShareHODL Blockchain - Complete System Documentation

**Version:** 1.0
**Last Updated:** 2026-01-29
**Status:** Pre-Production (Security Audit Complete)

---

## Table of Contents

1. [Overview](#overview)
2. [The 6-Tier Staking System](#the-6-tier-staking-system)
3. [HODL Stablecoin](#hodl-stablecoin)
4. [Company Equity System](#company-equity-system)
5. [DEX Trading](#dex-trading)
6. [P2P Escrow & Moderation](#p2p-escrow--moderation)
7. [Appeal System](#appeal-system)
8. [Wrong Resolution Recovery](#wrong-resolution-recovery)
9. [Reporter Abuse Prevention](#reporter-abuse-prevention)
10. [Lending System](#lending-system)
11. [Governance](#governance)
12. [System Architecture](#system-architecture)
13. [Security Model](#security-model)
14. [Module Reference](#module-reference)

---

## Overview

ShareHODL is a **self-governing blockchain** for tokenized equity trading built on Cosmos SDK v0.54 with CometBFT v2.0 consensus.

### Vision
Democratize global equity markets with:
- 24/7 trading
- Instant settlement
- Ultra-low fees (<$0.01/tx)
- No admin keys - fully community governed

### Key Statistics
- Target: 1000+ TPS
- Block time: 2 seconds
- Native token: HODL (USD-pegged stablecoin)

---

## The 6-Tier Staking System

Everything in ShareHODL is built on **stake-based trust**. Users stake HODL tokens to unlock capabilities:

### Tier Structure

| Tier | Name | Stake Required | Capabilities |
|------|------|----------------|--------------|
| 0 | Holder | Any amount | Trade on DEX, hold shares, vote |
| 1 | Keeper | 10,000 HODL | Lend/borrow, submit listings, file reports |
| 2 | Warden | 100,000 HODL | Moderate escrow disputes, verify small businesses |
| 3 | Steward | 1,000,000 HODL | Moderate large disputes (>100K), verify medium businesses |
| 4 | Archon | 5,000,000 HODL | Slash/blacklist moderators, final appeal reviews |
| 5 | Validator | 10,000,000 HODL | Run validator node, full oversight |

### Stake-as-Trust-Ceiling Principle

The core security model:
- **You can only moderate transactions up to your stake value**
- If you vote incorrectly and get overturned, **your stake is slashed to compensate the victim**
- This ensures moderators have "skin in the game"

### Reputation System

Each user has a reputation score that:
- Starts at 100 (neutral)
- Increases with successful moderations, on-time loan repayments
- Decreases with overturned decisions, defaults, false reports
- Affects voting weight in governance
- Below 50 = auto-deactivated from moderation

---

## HODL Stablecoin

HODL is the USD-pegged stablecoin that powers the ecosystem.

### Minting Process

```
User deposits $150 collateral (150% ratio required)
        │
        ▼
   HODL Module mints 100 HODL
        │
        ▼
   User receives 100 HODL tokens
```

### Collateral Requirements

| Parameter | Value |
|-----------|-------|
| Minimum Collateral Ratio | 150% |
| Liquidation Threshold | 130% |
| Liquidation Reward | 10% |
| Stability Fee | Variable (governance-controlled) |

### Liquidation Process

When collateral ratio drops to 130%:
1. Position becomes liquidatable
2. Any user can trigger liquidation
3. Liquidator repays the debt
4. Liquidator receives collateral + 10% bonus
5. Remaining collateral returned to original owner

### Supported Collateral Types

- HODL tokens (self-referential for leverage)
- Equity tokens (company shares)
- External assets (via IBC bridges - future)

---

## Company Equity System

Companies can tokenize their shares on ShareHODL.

### Listing Process

```
1. SUBMIT LISTING
   └── Founder (Keeper+ tier) submits application
   └── Documents: Registration, financials, governance
   └── Listing fee paid in HODL

2. VALIDATOR REVIEW
   └── 3+ validators assigned based on company size
   └── Each reviews documents and scores 1-100
   └── 70+ average score required to pass
   └── Validator stakes locked during review

3. APPROVAL/REJECTION
   └── Approved: Company goes Active, shares tradeable
   └── Rejected: Founder notified with reasons

4. ONGOING OPERATIONS
   └── Cap table management
   └── Dividend distribution
   └── Anti-dilution protections
```

### Share Classes

Companies can create multiple share classes:
- **Common**: Standard voting rights
- **Preferred**: Priority dividends, liquidation preference
- **Restricted**: Vesting schedules, transfer restrictions

### Dividend System

Fully automated dividend distribution:
1. Company declares dividend (amount, record date, payment date)
2. System snapshots all shareholders at record date
3. EndBlocker processes payments in batches
4. Shareholders receive proportional payments automatically

### Anti-Dilution Protections

Two types available:
- **Full Ratchet**: Price adjusts to lowest subsequent round
- **Weighted Average**: Price adjusts based on money raised

---

## DEX Trading

Decentralized exchange for trading company shares.

### Order Types

| Type | Description |
|------|-------------|
| Market | Execute immediately at best available price |
| Limit | Execute at specified price or better |
| Stop | Trigger market order when price reaches threshold |
| Stop-Limit | Trigger limit order when price reaches threshold |

### Time-in-Force Options

| Option | Description |
|--------|-------------|
| IOC | Immediate-or-Cancel: Fill what's possible, cancel rest |
| FOK | Fill-or-Kill: Complete fill or nothing |
| GTC | Good-til-Cancelled: Stays open indefinitely |
| GTD | Good-til-Date: Expires at specified time |

### Fees

- Trading fee: 0.3% per trade
- Maker rebate: Available for limit orders that add liquidity
- Fee abstraction: Pay fees in any supported token

### Fraud Detection

Built-in detection for:
- **Wash Trading**: Same entity on both sides
- **Spoofing**: Large orders placed then cancelled
- **Layering**: Multiple orders to create false depth
- **Front-running**: Transaction ordering manipulation

### Circuit Breakers

Automatic trading halts when:
- Price moves >10% in 1 minute
- Unusual volume detected
- Manual governance trigger

---

## P2P Escrow & Moderation

For off-chain deals (buying shares with fiat, physical goods, etc.)

### Escrow Flow

```
STEP 1: CREATE ESCROW
┌─────────────────────────────────────────────┐
│ • Seller deposits asset (shares/tokens)     │
│ • 3 moderators randomly assigned (Warden+)  │
│ • Conditions defined (e.g., "Pay $1000")    │
│ • Deadline set                              │
└─────────────────────────────────────────────┘
              │
STEP 2: BUYER FULFILLS CONDITIONS (off-chain)
              │
STEP 3: RESOLUTION
              │
    ┌─────────┴─────────┐
    │                   │
    ▼                   ▼
BOTH AGREE          DISPUTE
Seller confirms  →  Either party
payment received    claims problem
    │                   │
    ▼                   ▼
Release to         Moderator
Buyer              Voting
```

### Moderator Voting

When dispute occurs:
1. 3 assigned moderators review evidence
2. Each votes: Release to Buyer, Refund to Seller, or Split
3. Simple majority wins
4. 7-day deadline (auto-resolves based on votes if expired)

### Moderator Selection

Moderators are selected based on:
- Minimum tier requirement (Warden+ for standard, Steward+ for large)
- Stake must exceed transaction value (trust ceiling)
- Reputation score
- No conflicts of interest (not participant in escrow)

---

## Appeal System

Three levels of appeal for disputed decisions.

### Level Structure

| Level | Reviewers | Tier Required | Deadline | Can Overturn |
|-------|-----------|---------------|----------|--------------|
| 1 (Standard) | 5 | Warden+ | 5 days | Yes |
| 2 (Escalated) | 7 | Steward+ | 7 days | Yes |
| 3 (Final) | 9 | Archon+ | 10 days | Final decision |

### Appeal Process

1. Losing party files appeal with reason and new evidence
2. Higher-tier reviewers assigned
3. Reviewers vote to uphold or overturn
4. If overturned, original moderators penalized

### Moderator Penalties on Overturn

| Offense | Reputation | Stake Slash | Other |
|---------|------------|-------------|-------|
| 1st overturn | -2500 | None | Warning issued |
| 2nd overturn (30 days) | -2500 | 5% | Formal warning |
| 3rd overturn (30 days) | -2500 | 10% | 14-day blacklist |
| 4th+ overturn | -2500 | 15% | 28-day blacklist |

---

## Wrong Resolution Recovery

If you lose funds due to a bad moderator decision.

### Process Flow

```
1. REPORT SUBMITTED
   └── User files "WrongResolution" report
   └── Must be participant in original escrow
   └── Requires Keeper+ tier

2. VOLUNTARY RETURN PERIOD (48 hours)
   └── Counterparty notified
   └── Can RETURN funds → Case closed, no penalties
   └── Can REJECT → Escalates to investigation
   └── Ignores → Escalates after 48 hours

3. INVESTIGATION
   └── Keepers review evidence and vote
   └── Majority decides validity

4. RESOLUTION
   If Valid (Victim was wronged):
   └── Funds recovered via WATERFALL:
       1. Clawback from wrongful recipient
       2. Escrow reserve fund
       3. Slash moderators who voted wrong
   └── Victim compensated
   └── Moderators penalized

   If Invalid (False claim):
   └── Reporter penalized (escalating)
   └── Eventually banned if repeated
```

### Fund Recovery Waterfall

Priority order for recovering funds:

1. **Clawback from Wrongful Recipient**
   - First attempt to take back funds from Bob if he still has them
   - Most efficient - no new tokens created

2. **Escrow Reserve Fund**
   - Funded by 0.1% of all escrow fees
   - Used when clawback insufficient

3. **Moderator Stake Slashing**
   - Moderators who voted wrong have stake slashed
   - Slashed amount goes to victim
   - Last resort

---

## Reporter Abuse Prevention

Prevents spam and fraudulent reports.

### Escalating Penalties

| Offense | Reputation Loss | Stake Slash | Ban |
|---------|-----------------|-------------|-----|
| 1st false report | -1000 | None | None |
| 2nd false report | -1500 | 10% | None |
| 3rd false report | -2000 | 15% | 7 days |
| 4th false report | -3000 | 25% | 30 days |
| 5th+ false report | -5000 | 50% | Permanent |

### Rate Limiting

- Maximum 3 reports per day
- Cooldown after dismissed reports (1 day per consecutive dismissal)
- Banned users cannot submit any reports

### Anti-Sybil Protection

- **7-day minimum stake age** required to submit reports
- Prevents banned users from creating new accounts to bypass bans
- Ensures reporters have long-term commitment to the network

### False Report Rate Tracking

If >80% of reports are false with 10+ total reports → **Permanent Ban**

---

## Lending System

P2P and pool-based lending.

### Pool Lending

```
LENDERS                          BORROWERS
   │                                 │
   │ Deposit HODL                    │ Deposit Collateral
   ▼                                 ▼
┌─────────────────────────────────────────┐
│              LENDING POOL               │
│  • Utilization-based interest rates     │
│  • Dynamic APY for lenders              │
│  • 150% collateral requirement          │
│  • Liquidation at 125%                  │
└─────────────────────────────────────────┘
   │                                 │
   │ Earn Interest                   │ Borrow HODL
   ▼                                 ▼
```

### P2P Lending

```
LENDER                           BORROWER
   │                                 │
   │ Posts Offer                     │ Accepts with Collateral
   │ "10K HODL at 5% APR"            │
   ▼                                 ▼
┌─────────────────────────────────────────┐
│           P2P LOAN AGREEMENT            │
│  • Trust ceiling: Offer ≤ Lender stake  │
│  • Both deposit security stakes         │
│  • 14-day unbonding to exit             │
└─────────────────────────────────────────┘
```

### Default Handling

When borrower defaults:
1. Collateral liquidated
2. Lender receives collateral value
3. Borrower's reputation penalized
4. Borrower's security stake slashed

### Lending Parameters

| Parameter | Value |
|-----------|-------|
| Minimum Collateral Ratio | 150% |
| Liquidation Threshold | 125% |
| Minimum Loan Duration | 1 day |
| Maximum Loan Duration | 365 days |
| Interest Rate Range | 1% - 50% APR |

---

## Governance

All parameters are controlled by token holders.

### Proposal Types

| Type | Description | Quorum | Threshold |
|------|-------------|--------|-----------|
| Protocol Parameter | Change fees, ratios, timeouts | 33.4% | 50% |
| Treasury Spending | Allocate community funds | 33.4% | 50% |
| Company Delisting | Remove fraudulent company | 33.4% | 60% |
| Validator Change | Add/remove validators | 40% | 66.7% |
| Protocol Upgrade | Software upgrades | 40% | 66.7% |

### Emergency Proposals

Expedited voting for urgent matters:

| Severity | Required Tier | Voting Period |
|----------|---------------|---------------|
| Level 1 (Critical) | Archon+ | 24 hours |
| Level 2 (Urgent) | Steward+ | 48 hours |
| Level 3 (Important) | Warden+ | 72 hours |

### Vote Weighting

```
Vote Weight = Stake Amount × Reputation Multiplier

Reputation Multiplier:
  • 150+ reputation: 1.5x
  • 100-149 reputation: 1.0x
  • 50-99 reputation: 0.5x
  • <50 reputation: Cannot vote
```

### Vote Snapshots

- Voting power captured at proposal creation time
- Prevents manipulation via stake movement during voting
- Ensures fair representation

---

## System Architecture

### Module Dependency Graph

```
                         ┌─────────────┐
                         │  GOVERNANCE │
                         │  Controls   │
                         │  all params │
                         └──────┬──────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│    STAKING    │◄──────│     HODL      │──────►│     DEX       │
│  6-tier trust │       │  Stablecoin   │       │  Order book   │
│  Reputation   │       │  Collateral   │       │  Trading      │
└───────┬───────┘       └───────┬───────┘       └───────┬───────┘
        │                       │                       │
        │               ┌───────┴───────┐               │
        │               │               │               │
        ▼               ▼               ▼               ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│    ESCROW     │ │    LENDING    │ │    EQUITY     │ │  FEE ABSTRACT │
│  P2P trades   │ │  Loans/pools  │ │  Companies    │ │  Multi-token  │
│  Moderation   │ │  Collateral   │ │  Shares       │ │  fee payment  │
│  Disputes     │ │  Liquidation  │ │  Dividends    │ │               │
└───────────────┘ └───────────────┘ └───────────────┘ └───────────────┘
        │                                   │
        └───────────────────────────────────┘
                        │
                        ▼
         ┌─────────────────────────────┐
         │  REPORTING & MODERATION     │
         │  • Fraud reports            │
         │  • Appeals                  │
         │  • Wrong resolution claims  │
         │  • Auto-penalties           │
         │  • Company delisting        │
         └─────────────────────────────┘
```

### Cross-Module Communication

Modules communicate via **Keeper Interfaces** defined in `expected_keepers.go`:
- No circular dependencies
- Clean separation of concerns
- Testable in isolation

---

## Security Model

### Stake-Based Security

| Attack Vector | Protection |
|---------------|------------|
| Sybil Attack | Minimum stake requirements per tier + 7-day stake age for reporters |
| Moderator Collusion | Random selection, multiple moderators |
| False Reports | Escalating penalties, bans |
| Stake Flight | 14-day unbonding period |
| Vote Manipulation | Snapshots at proposal creation |
| All Moderators Blacklisted | Emergency 50/50 split resolution |
| Recovery Shortfall | Event emitted, victim can file governance proposal |

### Economic Security

| Mechanism | Purpose |
|-----------|---------|
| Trust Ceiling | Moderators can't moderate above their stake |
| Slashing | Penalize bad behavior |
| Reputation | Track long-term trustworthiness |
| Unbonding | Prevent escape before penalties |

### Decentralization

- **No Admin Keys**: All authority flows through governance
- **Governance-Controlled Parameters**: Community decides all settings
- **Validator Diversity**: Multiple independent validators required
- **Open Source**: All code publicly auditable

---

## Module Reference

### Core Modules

| Module | Path | Purpose |
|--------|------|---------|
| hodl | `x/hodl/` | USD-pegged stablecoin |
| dex | `x/dex/` | Order book trading |
| equity | `x/equity/` | Company shares, dividends |
| escrow | `x/escrow/` | P2P trades, disputes |
| governance | `x/governance/` | On-chain voting |
| staking | `x/staking/` | 6-tier stake system |
| lending | `x/lending/` | Loans and pools |
| feeabstraction | `x/feeabstraction/` | Multi-token fees |
| explorer | `x/explorer/` | Block indexing |

### Key Files

| File | Purpose |
|------|---------|
| `app/app.go` | Main application wiring |
| `x/*/keeper/keeper.go` | Business logic |
| `x/*/types/types.go` | Data structures |
| `x/*/types/errors.go` | Error definitions |
| `x/*/types/keys.go` | Storage keys |
| `x/*/module.go` | Module lifecycle |

### Configuration

| File | Purpose |
|------|---------|
| `genesis.json` | Initial chain state |
| `config.toml` | Node configuration |
| `app.toml` | Application settings |

---

## Appendix: Quick Reference

### Port Reference

| Port | Service |
|------|---------|
| 26657 | Tendermint RPC |
| 26656 | P2P |
| 1317 | REST API |
| 9090 | gRPC |
| 3001-3006 | Frontend apps |
| 8080 | Faucet |

### Token Denominations

| Token | Denomination | Precision |
|-------|--------------|-----------|
| HODL | uhodl | 6 decimals |
| 1 HODL = 1,000,000 uhodl |

### Chain Information

| Parameter | Value |
|-----------|-------|
| Chain ID | sharehodl-1 |
| Block Time | 2 seconds |
| Unbonding Period | 14 days |

---

## Contact & Support

- **GitHub**: https://github.com/sharehodl/sharehodl-blockchain
- **Documentation**: This file + inline code comments
- **Issues**: GitHub Issues for bug reports

---

*Document generated: 2026-01-29*
*System Version: Pre-Production*
