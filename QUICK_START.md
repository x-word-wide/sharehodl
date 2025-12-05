# ShareHODL Blockchain - Quick Start Guide

## 🚀 Get Started in 5 Minutes

This guide will get you running a local ShareHODL blockchain network with test data for development and testing.

## Prerequisites

- Go 1.21+
- Make
- Docker & Docker Compose (optional)
- Git

## Option 1: Native Setup (Recommended for Development)

### 1. Build the Project

```bash
# Clone and build
git clone https://github.com/sharehodl/sharehodl-blockchain.git
cd sharehodl-blockchain
make build

# Verify installation
./build/sharehodld version
```

### 2. Setup Local Network

```bash
# Run the setup script
./scripts/setup-localnet.sh

# This will:
# - Initialize a local node
# - Create test accounts (alice, bob, charlie, diana, companies)
# - Configure genesis with initial balances
# - Setup network parameters for fast development
```

### 3. Start the Network

```bash
# Start the blockchain
sharehodld start --home ~/.sharehodl

# Or use the generated startup script
~/.sharehodl/start-node.sh
```

### 4. Populate Test Data

In a new terminal:

```bash
# Wait for network to start (about 10 seconds)
# Then populate with test companies, shares, orders, and proposals
./scripts/populate-test-data.sh
```

### 5. Verify Setup

```bash
# Check network status
curl http://localhost:26657/status

# View test accounts
sharehodld keys list --home ~/.sharehodl

# Check account balance
sharehodl query bank balances $(sharehodld keys show alice -a --home ~/.sharehodl) --home ~/.sharehodl

# View test companies
sharehodld query equity company SHODL --home ~/.sharehodl

# Check order book
sharehodld query dex order-book SHODL/HODL --home ~/.sharehodl
```

## Option 2: Docker Setup (Recommended for Testing)

### 1. Quick Start with Docker

```bash
# Clone the repository
git clone https://github.com/sharehodl/sharehodl-blockchain.git
cd sharehodl-blockchain

# Start all services (blockchain + faucet + indexer)
docker-compose -f docker-compose.dev.yml up -d

# Check service status
docker-compose -f docker-compose.dev.yml ps
```

### 2. Access Services

- **Blockchain RPC**: http://localhost:26657
- **REST API**: http://localhost:1317  
- **Faucet**: http://localhost:8080
- **API Gateway**: http://localhost:3000
- **Frontend Dev**: http://localhost:3001 (if enabled)

### 3. Get Test Tokens

```bash
# Request tokens from faucet
curl -X POST http://localhost:8080/faucet \
  -H "Content-Type: application/json" \
  -d '{"address":"YOUR_BECH32_ADDRESS"}'

# Check faucet status
curl http://localhost:8080/status
```

## Network Endpoints

| Service | Endpoint | Description |
|---------|----------|-------------|
| RPC | http://localhost:26657 | Tendermint RPC |
| REST API | http://localhost:1317 | Cosmos SDK REST API |
| gRPC | localhost:9090 | gRPC endpoint |
| Faucet | http://localhost:8080 | Test token faucet |
| API Gateway | http://localhost:3000 | Unified API access |

## Test Accounts

The local network comes pre-configured with test accounts:

| Account | Role | Initial Balance |
|---------|------|----------------|
| `validator` | Validator | 100B HODL, 100B SHODL |
| `alice` | Individual | 10B HODL, 1B SHODL |
| `bob` | Individual | 10B HODL, 1B SHODL |
| `charlie` | Trader | 10B HODL, 1B SHODL |
| `diana` | Governance | 10B HODL, 1B SHODL |
| `sharehodl-corp` | Company | 50B HODL, 5B SHODL |
| `techstart` | Company | 25B HODL, 2.5B SHODL |
| `greenenergy` | Company | 75B HODL, 7.5B SHODL |

## Test Data

After running the populate script, you'll have:

### Companies
- **ShareHODL Corp** (SHODL) - Blockchain Technology
- **TechStart Inc** (TECH) - Software  
- **GreenEnergy LLC** (GREEN) - Renewable Energy
- **HealthTech Solutions** (HEALTH) - Healthcare Technology
- **FinanceFlow Corp** (FLOW) - Financial Technology

### Trading Pairs & Liquidity Pools
- SHODL/HODL
- TECH/HODL  
- GREEN/HODL

### Sample Orders
- Buy orders for major stocks
- Sell orders with competitive pricing
- Both limit and market orders

### Governance Proposals
- Protocol upgrade proposals
- Parameter change proposals
- Community funding proposals

## Common Commands

### Query Commands

```bash
# Company information
sharehodld query equity company SHODL --home ~/.sharehodl

# Account shareholdings  
sharehodld query equity shareholdings $(sharehodld keys show alice -a --home ~/.sharehodl) --home ~/.sharehodl

# Order book
sharehodld query dex order-book SHODL/HODL --home ~/.sharehodl

# Account balances
sharehodld query bank balances $(sharehodld keys show alice -a --home ~/.sharehodl) --home ~/.sharehodl

# Governance proposals
sharehodld query governance proposals --home ~/.sharehodl

# HODL supply
sharehodld query hodl supply --home ~/.sharehodl
```

### Transaction Commands

```bash
# Issue shares
sharehodld tx equity issue-shares alice SHODL 1000 --from sharehodl-corp --chain-id sharehodl-local-1 --home ~/.sharehodl

# Place trading order
sharehodld tx dex place-order LIMIT BUY SHODL/HODL 500 105.0 GTC --from alice --chain-id sharehodl-local-1 --home ~/.sharehodl

# Transfer shares
sharehodld tx equity transfer-shares bob SHODL 100 --from alice --chain-id sharehodl-local-1 --home ~/.sharehodl

# Mint HODL (validator only)
sharehodld tx hodl mint alice 1000000000hodl --from validator --chain-id sharehodl-local-1 --home ~/.sharehodl

# Submit governance proposal
sharehodld tx governance submit-proposal text "Test Proposal" "Description" 1000000hodl --from alice --chain-id sharehodl-local-1 --home ~/.sharehodl

# Vote on proposal
sharehodld tx governance vote 1 yes --from alice --chain-id sharehodl-local-1 --home ~/.sharehodl
```

## Development Tips

### Fast Block Times
Local network is configured with 1-second block times for rapid development.

### Auto Gas Estimation  
Add `--gas auto --gas-adjustment 1.3` to transactions for automatic gas estimation.

### Account Management
```bash
# List all accounts
sharehodld keys list --home ~/.sharehodl

# Show specific account address
sharehodld keys show alice -a --home ~/.sharehodl

# Create new account
sharehodld keys add newaccount --home ~/.sharehodl
```

### Reset Network
```bash
# Clean and restart
./scripts/setup-localnet.sh clean
./scripts/setup-localnet.sh
```

## Troubleshooting

### Network Won't Start
```bash
# Check for port conflicts
lsof -i :26657
lsof -i :1317

# Clean previous setup
rm -rf ~/.sharehodl
./scripts/setup-localnet.sh
```

### Transactions Failing
```bash
# Check account sequence
sharehodld query auth account $(sharehodld keys show alice -a --home ~/.sharehodl) --home ~/.sharehodl

# Verify account has funds
sharehodld query bank balances $(sharehodld keys show alice -a --home ~/.sharehodl) --home ~/.sharehodl
```

### Docker Issues
```bash
# View logs
docker-compose -f docker-compose.dev.yml logs sharehodl-node

# Restart services
docker-compose -f docker-compose.dev.yml restart

# Complete reset
docker-compose -f docker-compose.dev.yml down -v
docker-compose -f docker-compose.dev.yml up -d
```

## Next Steps

1. **Frontend Development**: Check `FRONTEND_DEVELOPMENT_PLAN.md` for UI development
2. **Testnet Deployment**: See `NETWORK_DEPLOYMENT_PLAN.md` for testnet setup
3. **Module Development**: Explore the custom modules in `x/` directory
4. **Testing**: Run E2E tests with `make test-e2e`

## Support

- **Documentation**: Check the `docs/` directory
- **Issues**: Report bugs on GitHub
- **Discord**: Join our development community

---

🎉 **You're ready to start developing on ShareHODL!**

The local network provides a complete testing environment with companies, trading, governance, and more. Explore the modules and start building!