# ShareHODL Documentation

Welcome to the comprehensive documentation for ShareHODL Blockchain - a purpose-built blockchain for decentralized equity trading.

## Quick Navigation

### 📚 Core Documentation
- [**Technical Overview**](TECHNICAL_OVERVIEW.md) - Architecture and design principles
- [**Module Documentation**](MODULES.md) - Detailed module specifications
- [**API Reference**](API_REFERENCE.md) - Complete API documentation
- [**Security**](SECURITY.md) - Security architecture and best practices

### 🚀 Getting Started
- [**Deployment Guide**](DEPLOYMENT_GUIDE.md) - Production deployment instructions
- [**Developer Setup**](../README.md#getting-started) - Local development environment
- [**Testing Guide**](../TEST_SUITE.md) - Running tests and validation

### 📋 Reference Materials
- [**Production Plan**](../PRODUCTION_DEPLOYMENT_PLAN.md) - Complete deployment roadmap
- [**Security Audit**](../SECURITY_AUDIT_PREP.md) - Security audit preparation
- [**Frontend Apps**](../sharehodl-frontend/README.md) - User interface applications

## Architecture Overview

ShareHODL is built on the Cosmos SDK v0.54 with CometBFT v2.0 consensus, featuring five custom modules:

```
┌─────────────────────────────────────────────────┐
│              Frontend Applications              │ 
│  Explorer | Trading | Governance | Wallet | Biz │
├─────────────────────────────────────────────────┤
│                 API Layer                       │
│           gRPC | REST | WebSocket               │
├─────────────────────────────────────────────────┤
│               Custom Modules                    │
│   HODL  │  DEX   │ Equity │ Validator │  Gov   │
├─────────────────────────────────────────────────┤
│             Cosmos SDK v0.54                    │
│   Auth | Bank | Staking | Governance | IBC     │
├─────────────────────────────────────────────────┤
│            CometBFT v2.0 Consensus              │
│      Tendermint | P2P | Mempool | State        │
└─────────────────────────────────────────────────┘
```

## Key Features

### 🔄 Ultra-Low Fees
- **0.025 uHODL per gas unit** (~$0.005 per transaction)
- **HODL Stablecoin**: USD-pegged for price stability
- **Atomic Swaps**: Cross-asset trading with guaranteed execution

### 🏢 Equity Trading
- **Tokenized Shares**: Blockchain-native equity tokens
- **Business Verification**: Multi-tier validator verification system
- **Instant Settlement**: Trade execution and ownership transfer in seconds

### ⭐ Validator Tiers
- **Bronze to Diamond**: Stake-based validator tiers
- **Reward Multipliers**: Up to 3x rewards for top-tier validators
- **Business Verification**: Validators earn equity for due diligence

### 🔒 Security
- **Economic Security**: Stake-based consensus with slashing
- **Audited Code**: Comprehensive security audits (scheduled)
- **Governance**: On-chain parameter updates and emergency procedures

## Module Summary

| Module | Purpose | Key Features |
|--------|---------|--------------|
| **HODL** | USD-pegged stablecoin | Collateral-backed minting, stability mechanisms |
| **DEX** | Decentralized exchange | Order book, atomic swaps, slippage protection |
| **Equity** | Share tokenization | Business verification, cap table management |
| **Validator** | Enhanced validation | Tier system, business verification rewards |
| **Governance** | Protocol governance | Parameter updates, emergency procedures |

## Development Status

### ✅ Completed (December 2024)
- [x] Core blockchain modules implementation
- [x] Frontend applications (5 apps) 
- [x] Build system compatibility (CometBFT v2.0)
- [x] Module integration and testing
- [x] Deployment automation
- [x] Comprehensive documentation

### ⏳ In Progress
- [ ] Security audits
- [ ] Performance optimization
- [ ] Mainnet preparation

### 🎯 Upcoming (Q1 2025)
- [ ] External security audits
- [ ] Mainnet launch
- [ ] Business onboarding
- [ ] Ecosystem development

## Quick Start Examples

### Running a Node
```bash
# Initialize node
./sharehodld init my-node --chain-id sharehodl-mainnet

# Start node
./sharehodld start --minimum-gas-prices="0.025uhodl"
```

### Basic API Usage
```bash
# Get account balance
curl https://api.sharehodl.com/cosmos/bank/v1beta1/balances/{address}

# Query HODL parameters
curl https://api.sharehodl.com/sharehodl/hodl/v1/params
```

### Creating Transactions
```bash
# Send HODL tokens
sharehodld tx bank send alice bob 1000uhodl \
  --from alice \
  --chain-id sharehodl-mainnet \
  --gas-prices 0.025uhodl

# Create DEX order
sharehodld tx dex create-order \
  --base-denom "equity/AAPL/common" \
  --quote-denom "uhodl" \
  --side "buy" \
  --quantity "100" \
  --price "150000000" \
  --from trader
```

## Performance Metrics

### Network Performance
- **Throughput**: 1,000+ TPS
- **Block Time**: 2 seconds
- **Finality**: Instant (single block)
- **Validators**: Up to 100

### Economic Efficiency
- **Gas Price**: 0.025 uHODL per gas unit
- **Transfer Fee**: ~$0.005 (0.5 cents)
- **Trade Fee**: ~$0.008 (0.8 cents)
- **Business Listing**: ~$0.013 (1.3 cents)

## Community and Support

### Resources
- **Website**: https://sharehodl.com
- **Discord**: https://discord.gg/sharehodl
- **GitHub**: https://github.com/sharehodl/sharehodl-blockchain
- **Documentation**: https://docs.sharehodl.com

### Contributing
We welcome contributions! See our [Contributing Guide](../CONTRIBUTING.md) for:
- Code style standards
- Testing requirements
- Pull request process
- Community guidelines

### Security
For security-related matters:
- **Security Team**: security@sharehodl.com
- **Bug Bounty**: Up to $50,000 for critical vulnerabilities
- **Coordinated Disclosure**: 90-day window

## Documentation Structure

```
docs/
├── README.md                 # This file - documentation index
├── TECHNICAL_OVERVIEW.md     # Architecture and technical details
├── MODULES.md               # Detailed module documentation
├── API_REFERENCE.md         # Complete API documentation
├── DEPLOYMENT_GUIDE.md      # Production deployment guide
└── SECURITY.md              # Security architecture and procedures
```

## Version Information

- **Cosmos SDK**: v0.54.0
- **CometBFT**: v2.0.0
- **Go Version**: 1.23+
- **Protocol Version**: 1.0.0
- **Documentation Version**: 1.0.0 (December 2024)

---

**Ready to build the future of global equity markets? Start with our [Technical Overview](TECHNICAL_OVERVIEW.md) or jump into the [Deployment Guide](DEPLOYMENT_GUIDE.md).**