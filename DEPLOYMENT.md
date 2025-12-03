# ShareHODL Blockchain Deployment Guide

This comprehensive guide covers deploying and operating the ShareHODL blockchain across different environments and platforms.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Environment Setup](#environment-setup)
4. [Deployment Options](#deployment-options)
5. [Configuration](#configuration)
6. [Monitoring & Observability](#monitoring--observability)
7. [Security Configuration](#security-configuration)
8. [Maintenance & Operations](#maintenance--operations)
9. [Troubleshooting](#troubleshooting)
10. [Appendices](#appendices)

## Prerequisites

### System Requirements

**Minimum Requirements:**
- CPU: 4 cores
- RAM: 8GB
- Storage: 100GB SSD
- Network: 1Gbps connection
- OS: Ubuntu 20.04+, CentOS 8+, or macOS 12+

**Recommended Production:**
- CPU: 8+ cores
- RAM: 32GB+
- Storage: 500GB+ NVMe SSD
- Network: 10Gbps connection
- OS: Ubuntu 22.04 LTS

### Software Dependencies

**Required:**
- Docker 24.0+
- Docker Compose 2.20+
- Git 2.30+
- Make 4.0+

**Optional (for specific deployment types):**
- Kubernetes 1.28+ (for K8s deployment)
- Terraform 1.5+ (for cloud deployment)
- Helm 3.12+ (for K8s package management)
- AWS CLI 2.0+ (for AWS deployment)

### Development Tools

```bash
# Install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install development tools
make tools

# Verify installation
go version
docker --version
make version
```

## Quick Start

### 1. Clone and Build

```bash
# Clone repository
git clone https://github.com/sharehodl/sharehodl-blockchain.git
cd sharehodl-blockchain

# Build the project
make build

# Verify build
./build/sharehodld version
```

### 2. Local Development Setup

```bash
# Setup development environment
make dev-setup

# Initialize local blockchain
make init

# Start local node
make start
```

### 3. Docker Development

```bash
# Deploy with Docker (3-node testnet)
./scripts/deploy.sh -e development -t docker -m -s

# Check status
docker-compose -f deployment/docker-compose.yml ps

# View logs
docker-compose -f deployment/docker-compose.yml logs -f sharehodl-node-0
```

### 4. Verify Deployment

```bash
# Check node status
curl http://localhost:26657/status

# Check API
curl http://localhost:1317/node_info

# Check metrics
curl http://localhost:26660/metrics
```

## Environment Setup

### Development Environment

**Purpose**: Local development and testing
**Resources**: Minimal
**Security**: Basic
**Monitoring**: Essential only

```bash
# Deploy development environment
./scripts/deploy.sh \
  --environment development \
  --network testnet \
  --type docker \
  --nodes 3 \
  --validators 3 \
  --monitoring

# Access services
echo "RPC: http://localhost:26657"
echo "API: http://localhost:1317"
echo "Grafana: http://localhost:3000 (admin/admin)"
echo "Prometheus: http://localhost:9090"
```

### Staging Environment

**Purpose**: Pre-production testing
**Resources**: Production-like
**Security**: Enhanced
**Monitoring**: Full stack

```bash
# Deploy staging environment
./scripts/deploy.sh \
  --environment staging \
  --network testnet \
  --type kubernetes \
  --nodes 5 \
  --validators 3 \
  --monitoring \
  --security \
  --backup

# Verify deployment
kubectl get pods -n sharehodl-staging
kubectl get svc -n sharehodl-staging
```

### Production Environment

**Purpose**: Live blockchain operation
**Resources**: High availability
**Security**: Maximum
**Monitoring**: Comprehensive

```bash
# Deploy production environment
./scripts/deploy.sh \
  --environment production \
  --network mainnet \
  --type terraform \
  --nodes 7 \
  --validators 5 \
  --monitoring \
  --security \
  --backup

# Verify deployment
terraform output
```

## Deployment Options

### Option 1: Docker Deployment

**Best for**: Development, small-scale testing

**Pros:**
- Quick setup
- Easy to manage
- Good for development

**Cons:**
- Limited scalability
- Single-host dependency

```bash
# Configuration
export SHAREHODL_ENV=development
export SHAREHODL_NETWORK=testnet
export SHAREHODL_NODES=3

# Deploy
./scripts/deploy.sh -e development -t docker

# Management commands
docker-compose -f deployment/docker-compose.yml up -d
docker-compose -f deployment/docker-compose.yml down
docker-compose -f deployment/docker-compose.yml restart sharehodl-node-0
```

### Option 2: Kubernetes Deployment

**Best for**: Staging, production with orchestration needs

**Pros:**
- High availability
- Auto-scaling
- Rolling updates
- Service discovery

**Cons:**
- Complex setup
- Requires K8s knowledge

```bash
# Prerequisites
kubectl cluster-info
helm version

# Deploy
./scripts/deploy.sh -e staging -t kubernetes

# Management
kubectl get all -n sharehodl
kubectl logs -f deployment/sharehodl-node-0 -n sharehodl
kubectl scale deployment sharehodl-node --replicas=5 -n sharehodl

# Updates
kubectl set image deployment/sharehodl-node sharehodl=sharehodl:v1.1.0 -n sharehodl
kubectl rollout status deployment/sharehodl-node -n sharehodl
kubectl rollout undo deployment/sharehodl-node -n sharehodl
```

### Option 3: Terraform Deployment

**Best for**: Production cloud deployments

**Pros:**
- Infrastructure as Code
- Multi-cloud support
- Disaster recovery
- Compliance

**Cons:**
- Cloud provider dependency
- Higher costs

```bash
# Prerequisites
terraform --version
aws configure  # or gcloud/az auth

# Deploy
./scripts/deploy.sh -e production -t terraform

# Management
cd deployment/terraform
terraform plan
terraform apply
terraform destroy  # WARNING: This will destroy everything
```

## Configuration

### Node Configuration

**config.toml** - CometBFT configuration:

```toml
# Network
moniker = "sharehodl-validator-1"
external_address = "your-public-ip:26656"
seeds = "seed1@seed1.sharehodl.io:26656,seed2@seed2.sharehodl.io:26656"
persistent_peers = ""

# Consensus
timeout_commit = "5s"
timeout_propose = "3s"
create_empty_blocks = true
create_empty_blocks_interval = "0s"

# P2P
max_num_inbound_peers = 40
max_num_outbound_peers = 10
pex = true
private_peer_ids = ""

# Mempool
size = 5000
max_txs_bytes = 1073741824  # 1GB
cache_size = 10000

# State Sync (for new nodes)
enable = true
rpc_servers = "rpc1.sharehodl.io:26657,rpc2.sharehodl.io:26657"
trust_height = 1000000
trust_hash = "trust_hash_here"

# Instrumentation
prometheus = true
prometheus_listen_addr = ":26660"
```

**app.toml** - Application configuration:

```toml
# Base Configuration
minimum-gas-prices = "0.025uhodl"
pruning = "custom"
pruning-keep-recent = "100"
pruning-interval = "10"

# API Configuration
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"
max-open-connections = 1000
rpc-read-timeout = 10
rpc-write-timeout = 0
rpc-max-body-bytes = 1000000
enabled-unsafe-cors = false

# gRPC Configuration
[grpc]
enable = true
address = "0.0.0.0:9090"

# State Sync Snapshots
[state-sync]
snapshot-interval = 1000
snapshot-keep-recent = 2
```

### Genesis Configuration

**genesis.json** - Network genesis state:

```json
{
  "genesis_time": "2024-01-01T00:00:00.000000000Z",
  "chain_id": "sharehodl-mainnet-1",
  "initial_height": "1",
  "consensus_params": {
    "block": {
      "max_bytes": "22020096",
      "max_gas": "10000000"
    },
    "evidence": {
      "max_age_num_blocks": "100000",
      "max_age_duration": "172800000000000"
    },
    "validator": {
      "pub_key_types": ["ed25519"]
    }
  },
  "app_hash": "",
  "app_state": {
    "auth": {
      "params": {
        "max_memo_characters": "256",
        "tx_sig_limit": "7",
        "tx_size_cost_per_byte": "10",
        "sig_verify_cost_ed25519": "590",
        "sig_verify_cost_secp256k1": "1000"
      }
    },
    "bank": {
      "params": {
        "send_enabled": [],
        "default_send_enabled": true
      },
      "balances": [],
      "supply": [],
      "denom_metadata": [
        {
          "description": "ShareHODL native token",
          "denom_units": [
            {
              "denom": "uhodl",
              "exponent": 0
            },
            {
              "denom": "hodl",
              "exponent": 6
            }
          ],
          "base": "uhodl",
          "display": "hodl",
          "name": "HODL",
          "symbol": "HODL"
        }
      ]
    }
  }
}
```

### Environment Variables

```bash
# Node Configuration
export SHAREHODL_HOME="/home/sharehodl/.sharehodl"
export SHAREHODL_CHAIN_ID="sharehodl-mainnet-1"
export SHAREHODL_MONIKER="validator-1"

# Network Configuration
export SHAREHODL_EXTERNAL_IP="your-public-ip"
export SHAREHODL_P2P_PORT="26656"
export SHAREHODL_RPC_PORT="26657"
export SHAREHODL_API_PORT="1317"
export SHAREHODL_GRPC_PORT="9090"

# Security Configuration
export SHAREHODL_KEYRING_BACKEND="file"
export SHAREHODL_ENABLE_SECURITY="true"
export SHAREHODL_AUDIT_LOG="true"

# Performance Configuration
export SHAREHODL_CACHE_SIZE="10000"
export SHAREHODL_MEMPOOL_SIZE="5000"
export SHAREHODL_MAX_CONNECTIONS="1000"

# Monitoring Configuration
export SHAREHODL_PROMETHEUS_ENABLED="true"
export SHAREHODL_METRICS_PORT="26660"
export SHAREHODL_LOG_LEVEL="info"
```

## Monitoring & Observability

### Prometheus Metrics

ShareHODL exposes comprehensive metrics for monitoring:

**Blockchain Metrics:**
```
# Block production
sharehodl_block_height
sharehodl_block_time_seconds
sharehodl_block_size_bytes

# Transaction metrics
sharehodl_tx_count_total
sharehodl_tx_fees_total
sharehodl_mempool_size

# Validator metrics
sharehodl_validator_power
sharehodl_validator_missed_blocks
sharehodl_validator_uptime_ratio
```

**Business Metrics:**
```
# Token metrics
sharehodl_hodl_supply_total
sharehodl_equity_tokens_count
sharehodl_equity_market_cap_usd

# DEX metrics
sharehodl_dex_trades_total
sharehodl_dex_volume_usd
sharehodl_dex_liquidity_usd

# Governance metrics
sharehodl_proposals_total
sharehodl_votes_total
sharehodl_governance_participation_rate
```

### Grafana Dashboards

**1. Blockchain Overview Dashboard:**
- Block production rate
- Transaction throughput
- Network health
- Validator performance

**2. Business Metrics Dashboard:**
- Trading volume and activity
- Token economics
- Governance participation
- Revenue metrics

**3. System Performance Dashboard:**
- CPU, memory, disk usage
- Network I/O
- Database performance
- Container health

**4. Security Dashboard:**
- Security events and alerts
- Failed authentication attempts
- Vulnerability scan results
- Compliance status

### Log Management

**Log Levels:**
- `error`: Critical errors requiring immediate attention
- `warn`: Warning conditions that should be monitored
- `info`: General information about system operation
- `debug`: Detailed information for troubleshooting

**Log Rotation:**
```bash
# Configure logrotate
cat > /etc/logrotate.d/sharehodl << EOF
/var/log/sharehodl/*.log {
    daily
    missingok
    rotate 30
    compress
    notifempty
    create 644 sharehodl sharehodl
    postrotate
        systemctl reload sharehodl
    endscript
}
EOF
```

### Alerting Rules

**Critical Alerts:**
- Node down
- Validator jailed
- High error rate
- Security incidents

**Warning Alerts:**
- High resource usage
- Slow response times
- Low peer count
- Failed transactions

**Configuration:**
```yaml
# Example alert rule
- alert: ShareHODLNodeDown
  expr: up{job="sharehodl-nodes"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "ShareHODL node {{ $labels.instance }} is down"
    description: "Node has been down for more than 1 minute"
    runbook_url: "https://docs.sharehodl.io/runbooks/node-down"
```

## Security Configuration

### Network Security

**Firewall Configuration:**
```bash
# Allow necessary ports
sudo ufw allow 22/tcp      # SSH
sudo ufw allow 26656/tcp   # P2P
sudo ufw allow 26657/tcp   # RPC (restrict to trusted IPs)
sudo ufw allow 1317/tcp    # API (restrict to trusted IPs)
sudo ufw allow 9090/tcp    # gRPC (restrict to trusted IPs)

# Enable firewall
sudo ufw enable
```

**Network Segmentation:**
```yaml
# Docker network configuration
networks:
  sharehodl-internal:
    driver: bridge
    internal: true
  sharehodl-external:
    driver: bridge
```

### Key Management

**Validator Key Security:**
```bash
# Generate validator key
sharehodld keys add validator --keyring-backend file

# Backup validator key (CRITICAL)
sharehodld keys export validator --keyring-backend file

# Set up Hardware Security Module (HSM) for production
sharehodld keys add validator --keyring-backend ledger
```

**Key Rotation:**
```bash
# Rotate node key
sharehodld unsafe-reset-all
sharehodl init new-node --chain-id sharehodl-mainnet-1

# Update validator commission
sharehodld tx staking edit-validator \
  --commission-rate="0.10" \
  --from=validator \
  --chain-id=sharehodl-mainnet-1
```

### Access Control

**RBAC Configuration:**
```yaml
# Kubernetes RBAC
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: sharehodl
  name: sharehodl-operator
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "watch", "list"]
- apiGroups: ["apps"]
  resources: ["deployments"]
  verbs: ["get", "list", "watch", "patch"]
```

**API Authentication:**
```bash
# Enable API authentication
export SHAREHODL_API_AUTH_ENABLED=true
export SHAREHODL_API_JWT_SECRET="your-jwt-secret"

# Configure rate limiting
export SHAREHODL_API_RATE_LIMIT=1000  # requests per minute
```

### Security Scanning

**Automated Security Scans:**
```bash
# Run comprehensive security scan
make sec-full

# Continuous security monitoring
docker run -d --name security-monitor \
  -v $(pwd):/app/source:ro \
  sharehodl-security

# Check scan results
curl http://localhost:8090/security/status
```

## Maintenance & Operations

### Backup Procedures

**Automated Backup:**
```bash
# Enable automated backups
export BACKUP_ENABLED=true
export BACKUP_INTERVAL=86400  # 24 hours
export S3_BUCKET=sharehodl-backups-prod

# Manual backup
./scripts/backup.sh
```

**Backup Verification:**
```bash
# Test backup restoration
./scripts/restore-test.sh backup-file.tar.gz

# Verify data integrity
sharehodld query bank total --node tcp://restored-node:26657
```

### Updates and Upgrades

**Rolling Updates:**
```bash
# Kubernetes rolling update
kubectl set image deployment/sharehodl-node \
  sharehodl=sharehodl:v1.1.0 -n sharehodl

# Monitor rollout
kubectl rollout status deployment/sharehodl-node -n sharehodl

# Rollback if needed
kubectl rollout undo deployment/sharehodl-node -n sharehodl
```

**Governance Upgrades:**
```bash
# Submit upgrade proposal
sharehodld tx gov submit-proposal software-upgrade v2.0.0 \
  --title="ShareHODL v2.0.0 Upgrade" \
  --description="Major protocol upgrade" \
  --upgrade-height=1000000 \
  --from=validator

# Vote on proposal
sharehodld tx gov vote 1 yes --from=validator
```

### Performance Tuning

**Database Optimization:**
```bash
# Compact database
sharehodld compact-db

# Prune old data
sharehodl prune --home ~/.sharehodl --pruning custom \
  --pruning-keep-recent 100 --pruning-interval 10
```

**Memory Management:**
```bash
# Configure memory limits
docker run --memory=8g --memory-swap=16g sharehodl:latest

# Monitor memory usage
docker stats sharehodl-node-0
```

### Health Checks

**Automated Health Checks:**
```bash
#!/bin/bash
# Health check script

# Check if node is running
if ! pgrep sharehodld > /dev/null; then
    echo "ERROR: ShareHODL node not running"
    exit 1
fi

# Check if node is syncing
LATEST_HEIGHT=$(curl -s localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
sleep 10
CURRENT_HEIGHT=$(curl -s localhost:26657/status | jq -r '.result.sync_info.latest_block_height')

if [ "$LATEST_HEIGHT" -eq "$CURRENT_HEIGHT" ]; then
    echo "WARNING: Node may not be syncing"
    exit 1
fi

echo "OK: Node is healthy"
```

## Troubleshooting

### Common Issues

**1. Node Not Syncing**
```bash
# Check peer connections
curl localhost:26657/net_info | jq '.result.n_peers'

# Check sync status
curl localhost:26657/status | jq '.result.sync_info'

# Solutions:
- Add more seed nodes
- Check firewall settings
- Verify network connectivity
- Reset node and re-sync
```

**2. High Memory Usage**
```bash
# Check memory usage
free -h
docker stats

# Solutions:
- Increase system memory
- Adjust cache settings
- Enable pruning
- Optimize database settings
```

**3. Failed Transactions**
```bash
# Check mempool
curl localhost:26657/num_unconfirmed_txs

# Check gas prices
sharehodld query bank total

# Solutions:
- Increase gas limit
- Check account balance
- Verify transaction format
- Check network congestion
```

**4. Validator Issues**
```bash
# Check validator status
sharehodl query staking validator $(sharehodl keys show validator --bech val -a)

# Check missed blocks
sharehodl query slashing signing-info $(sharehodl tendermint show-validator)

# Solutions:
- Check validator uptime
- Verify key access
- Monitor system resources
- Check network connectivity
```

### Debugging Tools

**Log Analysis:**
```bash
# View recent logs
journalctl -u sharehodl -f

# Filter by error level
grep "ERROR" /var/log/sharehodl/sharehodl.log

# Analyze performance
grep "block" /var/log/sharehodl/sharehodl.log | tail -100
```

**Network Debugging:**
```bash
# Test P2P connectivity
nc -zv peer-address 26656

# Check DNS resolution
nslookup seed.sharehodl.io

# Monitor network traffic
tcpdump -i any port 26656
```

**Performance Profiling:**
```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Enable profiling
export SHAREHODL_PROFILING_ENABLED=true
```

### Emergency Procedures

**Node Recovery:**
```bash
# Emergency stop
pkill sharehodld

# Backup current state
cp -r ~/.sharehodl ~/.sharehodl.backup

# Reset and resync
sharehodld unsafe-reset-all
sharehodld start --p2p.seeds="seed1:26656,seed2:26656"
```

**Network Recovery:**
```bash
# If network is halted
# 1. Coordinate with other validators
# 2. Agree on recovery height
# 3. Export genesis at recovery height
# 4. Restart network with new genesis
```

## Appendices

### A. Port Reference

| Service | Port | Protocol | Description |
|---------|------|----------|-------------|
| P2P | 26656 | TCP | Inter-node communication |
| RPC | 26657 | TCP | CometBFT RPC |
| API | 1317 | TCP | Cosmos SDK API |
| gRPC | 9090 | TCP | gRPC API |
| Metrics | 26660 | TCP | Prometheus metrics |
| Profile | 6060 | TCP | Go profiling |

### B. Directory Structure

```
~/.sharehodl/
├── config/
│   ├── config.toml       # CometBFT configuration
│   ├── app.toml          # Application configuration
│   ├── genesis.json      # Network genesis
│   ├── node_key.json     # Node P2P key
│   └── priv_validator_key.json # Validator key
├── data/
│   ├── application.db/   # Application state
│   ├── blockstore.db/    # Block storage
│   ├── evidence.db/      # Evidence database
│   ├── state.db/         # Consensus state
│   └── tx_index.db/      # Transaction index
└── keyring-test/         # Test keyring (dev only)
```

### C. Resource Requirements

**Development:**
- CPU: 2 cores
- RAM: 4GB
- Storage: 50GB
- Network: 100Mbps

**Staging:**
- CPU: 4 cores
- RAM: 16GB
- Storage: 200GB
- Network: 1Gbps

**Production:**
- CPU: 8+ cores
- RAM: 32GB+
- Storage: 1TB+ NVMe
- Network: 10Gbps

### D. Security Checklist

- [ ] Firewall configured correctly
- [ ] Validator keys secured (HSM/offline)
- [ ] Regular security scans enabled
- [ ] Access control implemented
- [ ] Monitoring and alerting active
- [ ] Backup and recovery tested
- [ ] Incident response plan ready
- [ ] Compliance requirements met

### E. Support Contacts

**Technical Support:**
- Email: support@sharehodl.io
- Discord: [ShareHODL Community](https://discord.gg/sharehodl)
- Documentation: https://docs.sharehodl.io

**Security Issues:**
- Email: security@sharehodl.io
- Bug Bounty: https://sharehodl.io/security/bug-bounty

**Emergency Contact:**
- Phone: +1-555-SHAREHODL
- Email: emergency@sharehodl.io

---

For more detailed information, please refer to the [ShareHODL Documentation](https://docs.sharehodl.io) or contact our support team.