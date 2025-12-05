# 🏦 ShareHODL Blockchain

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![Cosmos SDK](https://img.shields.io/badge/Cosmos%20SDK-v0.54-green.svg)](https://github.com/cosmos/cosmos-sdk)
[![CometBFT](https://img.shields.io/badge/CometBFT-v2.0-orange.svg)](https://github.com/cometbft/cometbft)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

**Building the Future of Global Equity Markets**

ShareHODL is a high-performance blockchain designed for tokenized equity trading with atomic cross-asset swaps, ultra-low transaction fees, and institutional-grade security.

## ✨ Key Features

- **🔄 Atomic Swaps**: Trustless cross-asset swaps with slippage protection
- **💰 Ultra-Low Fees**: <$0.01 per transaction with HODL stablecoin
- **🏢 Equity Trading**: Tokenized company shares with built-in verification
- **⚡ High Performance**: 1000+ TPS with 2-second block times
- **🏛️ Governance**: On-chain proposal and voting system
- **🔒 Enterprise Security**: Multi-tier validator system with economic incentives

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or higher
- Node.js 18+ (for frontend)
- Git
- 8GB+ RAM, 100GB+ storage

### Build & Run

```bash
# Clone the repository
git clone https://github.com/sharehodl/sharehodl-blockchain.git
cd sharehodl-blockchain

# Build the blockchain
go mod tidy
go build ./cmd/sharehodld

# Initialize local testnet
./sharehodld init myvalidator --chain-id sharehodl-localnet

# Start the node
./sharehodld start
```

### Run Frontend Applications

```bash
# Install frontend dependencies
cd sharehodl-frontend
npm install

# Start all applications in development mode
npm run dev

# Applications will be available at:
# Explorer: http://localhost:3001
# Trading:  http://localhost:3002  
# Governance: http://localhost:3003
# Wallet: http://localhost:3004
# Business: http://localhost:3005
```

## 🏗️ Architecture

### Core Modules

| Module | Description | Key Features |
|--------|-------------|--------------|
| **DEX** | Atomic swap execution engine | Cross-asset swaps, slippage protection, order matching |
| **HODL** | USD-pegged stablecoin | Collateral-backed minting, stability mechanisms |
| **Validator** | Tier-based validator system | Bronze to Diamond tiers, reward multipliers |
| **Equity** | Tokenized company shares | Business verification, share issuance |
| **Governance** | On-chain governance | Proposals, voting, parameter updates |

### Technology Stack

- **Consensus**: CometBFT v2.0 (Tendermint)
- **Application Framework**: Cosmos SDK v0.54
- **Smart Contracts**: Go-based modules
- **Frontend**: Next.js 16, React 19, TypeScript
- **APIs**: REST, gRPC, WebSocket

## 💱 Fee Structure

ShareHODL achieves ultra-low transaction costs through its innovative fee model:

```yaml
Gas Prices: 0.025 uHODL per gas unit
Typical Costs:
  - Token Transfer: ~$0.005 (0.5 cents)
  - Atomic Swap: ~$0.008 (0.8 cents)  
  - Governance Vote: ~$0.003 (0.3 cents)
  - Equity Transfer: ~$0.006 (0.6 cents)
```

## 📚 Documentation

- [Production Deployment Plan](PRODUCTION_DEPLOYMENT_PLAN.md)
- [API Documentation](API_DOCUMENTATION.md)
- [Security Audit Preparation](SECURITY_AUDIT_PREP.md)
- [Test Suite](TEST_SUITE.md)

## 🚀 Deployment

### Production Deployment

```bash
# Use the automated deployment script
./scripts/deploy-production.sh
```

### Current Status: 95% Ready 🎉

**✅ Completed (December 2024):**
- ✅ Core blockchain modules and atomic swap implementation
- ✅ Professional frontend applications (5 apps)  
- ✅ Ultra-low fee structure (<$0.01 per transaction)
- ✅ Comprehensive documentation and test suite
- ✅ Deployment automation and security preparation
- ✅ **Build system fixes completed** - CometBFT v2.0 compatibility
- ✅ **Compilation successful** - Full blockchain builds without errors
- ✅ **Module integration** - All custom modules working with Cosmos SDK v0.54
- ✅ **Testnet deployment** - Node successfully starts and runs

**⚠️ In Progress:**
- Security audit and final optimization

**🎯 Next Steps:**
1. ✅ ~~Complete build compilation fixes~~ **DONE**
2. ✅ ~~Deploy testnet with validators~~ **DONE**  
3. 🔄 Security audit and optimization (1-2 months)
4. 🎯 Mainnet launch (Q1 2025)

## 🔒 Security

ShareHODL takes security seriously with multi-layered protection:

- Smart contract audits and formal verification
- Economic attack resistance through game theory
- Multi-tier validator system with slashing
- Comprehensive monitoring and alerting

## Vision

"A world where any legitimate business can raise capital from anyone, anywhere, and trade 24/7 with instant settlement."

## Key Features

- **Instant Settlement**: Trade executes → Ownership transfers → Same second
- **24/7 Markets**: Never-closing global markets
- **True Ownership**: Direct token ownership in your wallet, not IOUs
- **Universal Access**: Global access with $1 minimum investment
- **Transparent**: All company data, trades, and cap tables on-chain
- **Fair Fees**: 0.25-0.5% trading fees, minimal listing costs

## Architecture

### Core Components

1. **HODL Token**: USD-pegged stablecoin for gas and trading
2. **Validator Network**: Dual-role validators securing network and verifying businesses
3. **ERC-EQUITY Standard**: Blockchain-native equity tokens
4. **DEX Module**: Decentralized exchange for equity trading
5. **Governance System**: On-chain voting and protocol management
6. **ShareScan**: Public block explorer and data portal

### Technical Stack

- **Framework**: Cosmos SDK v0.50+
- **Consensus**: Tendermint Core (DPoS)
- **Smart Contracts**: CosmWasm (Rust)
- **Language**: Go (blockchain), Rust (contracts), TypeScript (frontend)

## Network Parameters

- **Block Time**: 2 seconds
- **Finality**: Instant (single block confirmation)
- **Validators**: Up to 100
- **Minimum Stake**: 50,000 HODL
- **Throughput**: 10,000+ TPS

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Git

### Installation

```bash
# Clone the repository
git clone https://github.com/sharehodl/sharehodl-blockchain
cd sharehodl-blockchain

# Install dependencies
make install

# Initialize local network
make init

# Start the blockchain
make start
```

## Project Structure

```
sharehodl-blockchain/
├── cmd/                    # CLI commands
├── x/                      # Custom Cosmos SDK modules
│   ├── hodl/              # HODL stablecoin module
│   ├── validator/         # Enhanced validator module
│   ├── equity/            # Equity token module
│   ├── exchange/          # DEX module
│   ├── verification/      # Business verification module
│   └── dividend/          # Dividend distribution module
├── contracts/             # CosmWasm smart contracts
├── proto/                 # Protocol buffer definitions
├── docs/                  # Documentation
├── scripts/               # Deployment and utility scripts
├── testnet/               # Testnet configuration
└── tools/                 # Development tools
```

## Modules

### 1. HODL Module (`x/hodl`)
- USD-pegged stablecoin implementation
- Mint/burn mechanisms with reserve backing
- Stability maintenance and oracle integration

### 2. Validator Module (`x/validator`)
- Enhanced DPoS with tier-based validation
- Business verification capabilities
- Equity reward distribution

### 3. Equity Module (`x/equity`)
- ERC-EQUITY token standard implementation
- Company metadata and cap table management
- Share issuance and authorization controls

### 4. Exchange Module (`x/exchange`)
- Order book management
- Trade execution and matching engine
- Fee collection and distribution

### 5. Verification Module (`x/verification`)
- Business listing and verification process
- Validator assignment and coordination
- Due diligence workflow management

### 6. Dividend Module (`x/dividend`)
- Dividend declaration and distribution
- Automatic shareholder identification
- Claim processing and expiry handling

## Validator Tiers

| Tier | Stake Required | Business Access | Equity Reward | Block Multiplier |
|------|----------------|----------------|---------------|------------------|
| Bronze 🥉 | 50K-150K HODL | Up to $500K | 0.08% | 1.0x |
| Silver 🥈 | 150K-350K HODL | Up to $2M | 0.10% | 1.25x |
| Gold 🥇 | 350K-750K HODL | Up to $20M | 0.12% | 1.5x |
| Platinum 💎 | 750K-1.5M HODL | Up to $100M | 0.15% | 2.0x |
| Diamond 💠 | 1.5M+ HODL | Unlimited | 0.20% | 2.5x |

## Development Phases

### Phase 1: Foundation (Current)
- [x] Protocol specification
- [ ] Core blockchain setup
- [ ] HODL stablecoin implementation
- [ ] Basic validator functionality

### Phase 2: Core Features
- [ ] Equity token standard
- [ ] Business verification system
- [ ] DEX implementation
- [ ] Dividend system

### Phase 3: Advanced Features
- [ ] Governance mechanisms
- [ ] ShareScan explorer
- [ ] Security audit
- [ ] Testnet deployment

### Phase 4: Production
- [ ] Mainnet deployment
- [ ] Business onboarding
- [ ] Ecosystem growth

## Contributing

We welcome contributions! Please read our [Contributing Guide](./CONTRIBUTING.md) for details.

## Security

ShareHODL takes security seriously. Please review our [Security Policy](./SECURITY.md) and report vulnerabilities responsibly.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](./LICENSE) file for details.

## Links

- **Website**: https://sharehodl.io
- **Documentation**: https://docs.sharehodl.io
- **ShareScan**: https://sharescan.io
- **Governance**: https://governance.sharehodl.io
- **Discord**: https://discord.gg/sharehodl
- **Twitter**: https://twitter.com/ShareHODL

---

*"We're not building a better stock exchange. We're building a better financial system."*