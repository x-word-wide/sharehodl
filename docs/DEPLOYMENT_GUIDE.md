# ShareHODL Blockchain Deployment Guide

## Prerequisites

### System Requirements

**Minimum Hardware**:
- 4 CPU cores
- 8GB RAM
- 200GB SSD storage
- 100Mbps network connection

**Recommended Hardware (Production)**:
- 8+ CPU cores  
- 32GB RAM
- 1TB NVMe SSD storage
- 1Gbps network connection

### Software Dependencies

```bash
# Required software versions
Go: 1.23+
Git: 2.30+
Docker: 20.10+ (optional)
Ubuntu: 20.04+ or similar Linux distribution
```

## Installation

### 1. Clone Repository

```bash
git clone https://github.com/sharehodl/sharehodl-blockchain.git
cd sharehodl-blockchain
```

### 2. Build Binary

```bash
# Install dependencies
go mod download

# Build the blockchain binary
go build -o sharehodld ./cmd/sharehodld

# Verify build
./sharehodld version
```

### 3. Initialize Node

```bash
# Initialize node configuration
./sharehodld init <node-name> --chain-id <chain-id>

# Example for mainnet
./sharehodld init my-validator --chain-id sharehodl-mainnet

# Example for testnet  
./sharehodld init my-validator --chain-id sharehodl-testnet
```

## Configuration

### 1. Genesis Configuration

The genesis file contains the initial state of the blockchain. Key parameters:

```json
{
  "app_state": {
    "staking": {
      "params": {
        "bond_denom": "uhodl",
        "unbonding_time": "1814400s",
        "max_validators": 100
      }
    },
    "hodl": {
      "params": {
        "minting_enabled": true,
        "collateral_ratio": "1.500000000000000000",
        "stability_fee": "0.000500000000000000"
      }
    }
  }
}
```

### 2. Node Configuration

Edit `~/.sharehodl/config/config.toml`:

```toml
# Network configuration
[p2p]
laddr = "tcp://0.0.0.0:26656"
persistent_peers = "<peer-list>"
seeds = "<seed-list>"

# Consensus configuration  
[consensus]
timeout_propose = "2s"
timeout_commit = "2s"

# Mempool configuration
[mempool]
size = 5000
max_txs_bytes = 1073741824
```

Edit `~/.sharehodl/config/app.toml`:

```toml
# Gas price configuration
minimum-gas-prices = "0.025uhodl"

# API configuration
[api]
enable = true
address = "tcp://0.0.0.0:1317"

[grpc]
enable = true  
address = "0.0.0.0:9090"
```

### 3. Validator Setup (For Validators)

```bash
# Create validator key
./sharehodld keys add validator

# Add account to genesis (for new networks)
./sharehodl genesis add-genesis-account validator 1000000000000uhodl

# Create genesis transaction
./sharehodld genesis gentx validator 500000000000uhodl \
  --chain-id sharehodl-mainnet \
  --commission-rate 0.05 \
  --commission-max-rate 0.10 \
  --commission-max-change-rate 0.01

# Collect genesis transactions (for network coordinators)
./sharehodl genesis collect-gentxs
```

## Network Deployment

### Mainnet Deployment

1. **Download Genesis File**:
```bash
wget https://raw.githubusercontent.com/sharehodl/networks/main/mainnet/genesis.json \
  -O ~/.sharehodl/config/genesis.json
```

2. **Configure Peers**:
```bash
# Edit config.toml with mainnet peers
persistent_peers = "peer1@ip1:26656,peer2@ip2:26656"
seeds = "seed1@seed-ip1:26656,seed2@seed-ip2:26656"
```

3. **Start Node**:
```bash
# Start the node
./sharehodld start

# Or with systemd service
sudo systemctl start sharehodld
```

### Testnet Deployment

1. **Download Testnet Genesis**:
```bash
wget https://raw.githubusercontent.com/sharehodl/networks/main/testnet/genesis.json \
  -O ~/.sharehodl/config/genesis.json
```

2. **Configure for Testnet**:
```bash
# Set chain ID
sed -i 's/sharehodl-mainnet/sharehodl-testnet/g' ~/.sharehodl/config/genesis.json
```

3. **Start Testnet Node**:
```bash
./sharehodld start --minimum-gas-prices "0.025uhodl"
```

## Production Setup

### 1. Systemd Service

Create `/etc/systemd/system/sharehodld.service`:

```ini
[Unit]
Description=ShareHODL Blockchain Node
After=network-online.target

[Service]
Type=simple
User=sharehodl
WorkingDirectory=/home/sharehodl
ExecStart=/home/sharehodl/sharehodld start
Restart=on-failure
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable sharehodld
sudo systemctl start sharehodld
```

### 2. Monitoring Setup

**Prometheus Configuration** (`prometheus.yml`):
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'sharehodl'
    static_configs:
      - targets: ['localhost:26660']
```

**Grafana Dashboards**:
- Node health metrics
- Transaction throughput
- Block production time
- Validator performance

### 3. Backup Strategy

**Daily State Backup**:
```bash
#!/bin/bash
# backup-state.sh
DATE=$(date +%Y%m%d)
./sharehodld export > backup-$DATE.json
tar -czf backup-$DATE.tar.gz ~/.sharehodl/data/
```

**Automatic Snapshots**:
```bash
# Enable state sync snapshots
snapshot-interval = 1000
snapshot-keep-recent = 2
```

## Security Configuration

### 1. Firewall Rules

```bash
# Allow necessary ports
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 26656/tcp # P2P
sudo ufw allow 26657/tcp # RPC (optional)
sudo ufw allow 1317/tcp  # API (optional)
sudo ufw allow 9090/tcp  # gRPC (optional)

sudo ufw enable
```

### 2. Key Management

**Hardware Security Module (HSM)**:
```bash
# Configure HSM for validator signing
[priv_validator_key_file]
key_type = "tendermint/PrivValidatorKey"
priv_key = "priv-validator-key.json"

# For production, use external HSM
[priv_validator_laddr]
laddr = "tcp://127.0.0.1:26659"
```

**Key Backup**:
```bash
# Backup validator keys securely
cp ~/.sharehodl/config/priv_validator_key.json /secure/backup/location/
cp ~/.sharehodl/config/node_key.json /secure/backup/location/
```

### 3. Remote Signer Setup

For enhanced security, use a remote signer:

```bash
# Start remote signer on separate machine
tmkms start -c tmkms.toml

# Configure node to use remote signer
priv_validator_laddr = "tcp://0.0.0.0:26659"
```

## Scaling and Performance

### 1. Database Optimization

**PostgreSQL Backend** (optional):
```bash
# Configure for high-performance database
app_db_backend = "pebbledb"  # or "rocksdb"
iavl-cache-size = 781250
```

### 2. Hardware Scaling

**Vertical Scaling**:
- Increase CPU cores for faster transaction processing
- Add RAM for larger state cache
- Use NVMe storage for faster I/O

**Horizontal Scaling**:
- Run multiple full nodes behind load balancer
- Separate RPC/API nodes from consensus nodes
- Use state sync for fast bootstrapping

### 3. Network Optimization

```toml
# Increase peer limits for better connectivity
max_num_inbound_peers = 80
max_num_outbound_peers = 80

# Optimize mempool for high throughput
size = 10000
max_txs_bytes = 2147483648
```

## Maintenance

### 1. Regular Updates

```bash
# Update to latest version
git pull origin main
go build -o sharehodld ./cmd/sharehodld

# Restart with new binary
sudo systemctl restart sharehodld
```

### 2. State Pruning

```bash
# Configure pruning to save disk space
pruning = "custom"
pruning-keep-recent = "100"
pruning-keep-every = "0"  
pruning-interval = "10"
```

### 3. Log Management

```bash
# Configure log rotation
log_level = "info"
log_format = "json"

# Use logrotate for file management
/var/log/sharehodl/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
```

## Troubleshooting

### Common Issues

1. **Node Not Syncing**:
```bash
# Check peer connections
./sharehodld status | jq '.sync_info'

# Verify genesis file
shasum -a 256 ~/.sharehodl/config/genesis.json
```

2. **High Memory Usage**:
```bash
# Reduce IAVL cache size
iavl-cache-size = 100000

# Enable pruning
pruning = "everything"
```

3. **Transaction Failures**:
```bash
# Check minimum gas prices
minimum-gas-prices = "0.025uhodl"

# Verify account balance
./sharehodld query bank balances <address>
```

### Performance Monitoring

```bash
# Monitor node health
./sharehodld status

# Check validator status  
./sharehodld query staking validator <validator-address>

# Monitor system resources
htop
iostat -x 1
```

## Support

For deployment support:
- **Documentation**: https://docs.sharehodl.com
- **Discord**: https://discord.gg/sharehodl
- **GitHub Issues**: https://github.com/sharehodl/sharehodl-blockchain/issues

---

This deployment guide covers all aspects of running ShareHODL nodes from development to production environments.