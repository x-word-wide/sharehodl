# ShareHODL Blockchain - Project Intelligence

## Project Overview
ShareHODL is a high-performance blockchain platform for tokenized equity trading built on Cosmos SDK v0.54 with CometBFT v2.0 consensus.

**Vision**: Democratize global equity markets with 24/7 trading, instant settlement, and ultra-low fees (<$0.01/tx)

**Key Stats**:
- Target: 1000+ TPS, 2-second blocks
- Status: ~95% complete, pending security audit
- Target Launch: Q1 2025

## Tech Stack

| Layer | Technology |
|-------|------------|
| Blockchain | Go 1.23.5, Cosmos SDK v0.54, CometBFT v2.0 |
| Frontend | Next.js 16, React 19, TypeScript, Turborepo (pnpm) |
| Database | PostgreSQL 15, Redis 7 |
| Deployment | Docker, Docker Compose |

## Architecture

### Core Blockchain Modules (`x/`)
- **hodl/** - USD-pegged stablecoin (150% collateral, liquidation at 130%)
- **dex/** - Order book trading with atomic swaps, 0.3% fees
- **equity/** - Tokenized company shares, cap table management
- **validator/** - 5-tier DPoS (Bronze→Diamond), stake-based rewards
- **governance/** - On-chain voting (33.4% quorum, 50% threshold)
- **explorer/** - Block/transaction indexing

### Frontend Apps (`sharehodl-frontend/apps/`)
| App | Port | Purpose |
|-----|------|---------|
| home | 3001 | Landing, company directory |
| explorer | 3002 | Block browser, tx viewer |
| trading | 3003 | DEX interface |
| wallet | 3004 | Account management |
| governance | 3005 | Proposal voting |
| business | 3006 | Company onboarding |

### Services (`services/`)
- **indexer/** - PostgreSQL blockchain indexer

## Key Files Reference

### Blockchain
- `app/app.go` - Main application wiring
- `x/*/keeper/` - Business logic per module
- `x/*/types/` - Data structures and messages
- `proto/sharehodl/` - Protocol buffer definitions
- `genesis.json` - Initial chain state

### Frontend
- `sharehodl-frontend/turbo.json` - Monorepo config
- `sharehodl-frontend/apps/*/` - Individual Next.js apps
- `sharehodl-frontend/packages/ui/` - Shared components

### Deployment
- `docker-compose.dev.yml` - Development environment
- `docker-compose.production.yml` - Production setup
- `Makefile` - Build automation

## Development Commands

```bash
# Build blockchain
make build
make install

# Run local testnet
make localnet-start

# Generate protobuf
make proto-gen

# Frontend development
cd sharehodl-frontend && pnpm install && pnpm dev

# Docker development
docker-compose -f docker-compose.dev.yml up

# Run tests
make test
cd sharehodl-frontend && pnpm test
```

## Coding Standards

### Go (Blockchain)
- Follow Cosmos SDK patterns (keeper, types, module.go)
- Use protocol buffers for all message types
- Emit events for all state changes
- Maintain 150% collateral ratio for HODL operations
- All keeper methods must validate inputs

### TypeScript (Frontend)
- Use TypeScript strict mode
- Follow Next.js App Router conventions
- Shared components go in `packages/ui/`
- Use Tailwind CSS for styling
- API calls through dedicated hooks

### Testing
- Unit tests for all keeper methods
- Integration tests for module interactions
- E2E tests for critical user flows
- Test coverage target: >80%

## Security Requirements

- Never expose private keys in logs
- Validate all user inputs at boundaries
- Use parameterized queries for database
- Implement rate limiting on public endpoints
- Follow OWASP top 10 guidelines
- All financial operations need atomic transactions

## Agent Team

This project uses a coordinated AI agent team. See `.claude/agents/` for specialized agents:

- **Architect** - System design, feature planning, architecture decisions
- **Security Auditor** - Security review, vulnerability detection
- **Test Engineer** - Test creation, coverage analysis
- **Performance Optimizer** - Bottleneck identification, optimization
- **Documentation Specialist** - Docs maintenance, API documentation
- **Implementation Engineer** - Code implementation, bug fixes

## Progress Tracking

See `.claude/memory/` for:
- `progress.md` - Current sprint progress
- `decisions.md` - Architecture decision records
- `learnings.md` - Project-specific learnings
- `issues.md` - Known issues and blockers

## Quick Reference

### Ports
- 26657: Tendermint RPC
- 26656: P2P
- 1317: REST API
- 9090: gRPC
- 3001-3006: Frontend apps
- 5432: PostgreSQL
- 6379: Redis

### Genesis Accounts
| Account | Address | Balance |
|---------|---------|---------|
| **Validator** | `hodl1xwmmwhh95fdzw3an97v6369244477e9pvzm0wx` | 100M HODL (10M staked) |
| **Treasury** | `hodl1dd0st03tk3fwd47r7h7rxcjchw40zph02qah7l` | 1B HODL |

### Genesis Validator
- Address: `hodl1xwmmwhh95fdzw3an97v6369244477e9pvzm0wx`
- Operator: `hodlvaloper1xwmmwhh95fdzw3an97v6369244477e9phcfjmx`
- Moniker: "ShareHODL Foundation"
- Commission: 5%
- Self-delegation: 10M HODL
- Keyring: `file` backend (production secure)

### Chain ID
- `sharehodl-1`

### Token Denomination
- `uhodl` (micro HODL, 1 HODL = 1,000,000 uhodl)

## Deployment

### Production Server
- **IP**: 178.63.13.190
- **User**: root
- **Blockchain Data**: /root/.sharehodl/
- **Keyring Passphrase**: /root/.sharehodl/keyring-passphrase (chmod 600)
- **Telegram Webapp Path**: /var/www/sharehodl-telegram/

### Quick Deploy Commands

```bash
# Deploy Telegram Webapp (one command)
cd sharehodl-telegram/webapp && npm run build && \
ssh root@178.63.13.190 "rm -rf /var/www/sharehodl-telegram/*" && \
scp -r dist/* root@178.63.13.190:/var/www/sharehodl-telegram/

# Check blockchain status
ssh root@178.63.13.190 "curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height'"

# View blockchain logs
ssh root@178.63.13.190 "journalctl -u sharehodld -f"

# Restart blockchain
ssh root@178.63.13.190 "systemctl restart sharehodld"
```

### Deploy Telegram Webapp (Step by Step)
```bash
# 1. Build locally
cd sharehodl-telegram/webapp
npm run build

# 2. Clear cache and deploy
ssh root@178.63.13.190 "rm -rf /var/www/sharehodl-telegram/*"
scp -r dist/* root@178.63.13.190:/var/www/sharehodl-telegram/
```

### Blockchain Management
```bash
# View validator status
ssh root@178.63.13.190 "curl -s http://localhost:1317/cosmos/staking/v1beta1/validators | jq '.validators[0]'"

# Check account balances
ssh root@178.63.13.190 "curl -s http://localhost:1317/cosmos/bank/v1beta1/balances/hodl1xwmmwhh95fdzw3an97v6369244477e9pvzm0wx"

# Service management
ssh root@178.63.13.190 "systemctl status sharehodld"
ssh root@178.63.13.190 "systemctl restart sharehodld"
ssh root@178.63.13.190 "systemctl stop sharehodld"
```

### Security Notes
- Keyring uses `file` backend with encrypted passphrase
- Passphrase stored at `/root/.sharehodl/keyring-passphrase` (root only, chmod 600)
- Never share mnemonics or passphrases in chat/logs
- For key operations, SSH directly to server
