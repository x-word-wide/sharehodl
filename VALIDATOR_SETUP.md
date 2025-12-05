# ShareHODL Validator Node Setup & Address Security

## Validator Node Architecture

### Hardware Requirements

**Minimum Specifications:**
- **CPU**: 8 cores, 3.0+ GHz (Intel Xeon or AMD EPYC)
- **RAM**: 32GB DDR4 (64GB recommended for Gold/Platinum tiers)
- **Storage**: 2TB NVMe SSD (10,000+ IOPS)
- **Network**: 1Gbps dedicated bandwidth, <50ms latency
- **Redundancy**: UPS backup, redundant internet connections

**Production Setup:**
- **Primary Node**: Main validator with hot keys
- **Backup Node**: Standby validator with encrypted keys
- **Sentry Nodes**: 2-3 nodes for DDoS protection
- **Monitoring**: 24/7 alerting and performance tracking

### Node Installation Process

**1. System Preparation**
```bash
# Update system and install dependencies
sudo apt update && sudo apt upgrade -y
sudo apt install build-essential git curl jq snapd -y

# Install Go 1.21+
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Create sharehodl user for security
sudo useradd -m -s /bin/bash sharehodl
sudo usermod -aG sudo sharehodl
```

**2. ShareHODL Binary Installation**
```bash
# Clone and build ShareHODL
git clone https://github.com/sharehodl/sharehodl-blockchain.git
cd sharehodl-blockchain
make install

# Verify installation
sharehodld version
# Output: sharehodl-1.0.0-commit_hash
```

**3. Node Initialization**
```bash
# Initialize validator node
sharehodld init "MyValidator" --chain-id sharehodl-mainnet-1

# Download genesis file
curl -s https://raw.githubusercontent.com/sharehodl/networks/main/mainnet/genesis.json > ~/.sharehodl/config/genesis.json

# Configure persistent peers
PEERS="node1@sharehodl-seed-1.com:26656,node2@sharehodl-seed-2.com:26656"
sed -i "s/persistent_peers = \"\"/persistent_peers = \"$PEERS\"/" ~/.sharehodl/config/config.toml
```

**4. Security Configuration**
```bash
# Generate validator keys securely
sharehodld keys add validator --keyring-backend file
sharehodld keys add operator --keyring-backend file

# Set up key management service (recommended for production)
# Option 1: Hardware Security Module (HSM)
# Option 2: HashiCorp Vault integration
# Option 3: AWS KMS for cloud deployments

# Configure firewall
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 26656/tcp # P2P
sudo ufw allow 26657/tcp # RPC (restrict to monitoring)
sudo ufw enable
```

**5. Systemd Service Setup**
```bash
# Create systemd service
sudo tee /etc/systemd/system/sharehodld.service > /dev/null <<EOF
[Unit]
Description=ShareHODL Validator
After=network.target

[Service]
Type=simple
User=sharehodl
WorkingDirectory=/home/sharehodl
ExecStart=/usr/local/bin/sharehodld start --home /home/sharehodl/.sharehodl
Restart=on-failure
RestartSec=10
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable sharehodld
sudo systemctl start sharehodld
```

## Address Generation Security

### Address Format: `sharehodl1...`

**Structure:**
- **Prefix**: `sharehodl` (human-readable identifier)
- **Separator**: `1` (bech32 separator)
- **Data**: 39 characters (160-bit hash + checksum)
- **Total Length**: 47 characters
- **Example**: `sharehodl1qjk2rmxlu4fsylvdj3w8m9e6qu6t8q5n9r2x7a8c`

### Cryptographic Foundation

**1. Private Key Generation**
```go
// 256-bit private key using secp256k1 curve
privateKey := secp256k1.GenPrivKey()
// Entropy: 256 bits of cryptographically secure randomness
// Source: OS entropy pool + additional randomness
```

**2. Public Key Derivation**
```go
publicKey := privateKey.PubKey()
// Uses elliptic curve point multiplication
// secp256k1: y² = x³ + 7 (same as Bitcoin/Ethereum)
```

**3. Address Generation Process**
```go
// Step 1: Get compressed public key (33 bytes)
compressedPubKey := publicKey.Bytes()

// Step 2: SHA256 hash
sha256Hash := sha256.Sum256(compressedPubKey)

// Step 3: RIPEMD160 hash (20 bytes = 160 bits)
ripemd160 := ripemd160.Sum256(sha256Hash[:])

// Step 4: Bech32 encoding with 'sharehodl' prefix
address := bech32.Encode("sharehodl", ripemd160)
```

### Security Measures

**Entropy Sources:**
- **OS Entropy**: `/dev/urandom` on Linux
- **Hardware RNG**: Intel RDRAND instruction
- **Additional Randomness**: Mouse movements, keyboard timings
- **Entropy Pool**: Minimum 256 bits before key generation

**Key Storage Security:**

**Option 1: Hardware Security Modules (HSM)**
```yaml
# Production validator setup
hsm:
  type: "aws-cloudhsm"
  cluster_id: "cluster-abc123"
  key_spec: "ECC_SECG_P256K1"
  usage: "SIGN_VERIFY"
```

**Option 2: Encrypted File Storage**
```bash
# Key encryption with scrypt + AES-256-GCM
sharehodld keys add validator \
  --keyring-backend file \
  --keyring-dir ~/.sharehodl/keys \
  --algo secp256k1
# Password: minimum 12 characters, mixed case, numbers, symbols
```

**Option 3: Distributed Key Management**
```yaml
# Shamir's Secret Sharing (3-of-5 threshold)
threshold_scheme:
  threshold: 3
  total_shares: 5
  shares_distributed_to:
    - "backup-location-1"
    - "backup-location-2" 
    - "backup-location-3"
    - "backup-location-4"
    - "backup-location-5"
```

## Validator Registration Process

**1. Stake Requirement**
```bash
# Minimum stake by tier (MAINNET)
Bronze:   50,000 HODL   ($50,000)
Silver:   100,000 HODL  ($100,000)
Gold:     250,000 HODL  ($250,000)
Platinum: 500,000 HODL  ($500,000)

# Testnet requirements (for testing)
Bronze:   1,000 HODL   ($1,000)
Silver:   5,000 HODL   ($5,000)
Gold:     10,000 HODL  ($10,000)
Platinum: 30,000 HODL  ($30,000)
```

**2. Business Verification**
```bash
# Submit validator registration
sharehodld tx staking create-validator \
  --amount=10000000000uhodl \
  --pubkey=$(sharehodld tendermint show-validator) \
  --moniker="MyValidator" \
  --chain-id=sharehodl-mainnet-1 \
  --commission-rate="0.05" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="10000000000" \
  --business-license="LICENSE_NUMBER" \
  --compliance-cert="CERT_HASH" \
  --from=validator
```

**3. Compliance Requirements**
- **Business Registration**: Incorporated entity required
- **Insurance**: $1M+ professional liability coverage
- **Compliance Officer**: Designated regulatory contact
- **Audit Trail**: 7-year transaction log retention
- **KYC Documentation**: Identity verification for all operators

## Network Security Features

**Sentry Node Architecture:**
```
Internet → Sentry Nodes → Private Network → Validator Node
```

**DDoS Protection:**
- **Rate Limiting**: 1000 requests/minute per IP
- **Connection Limits**: 50 inbound connections max
- **Firewall Rules**: Whitelist known validators only
- **Proxy Shields**: Cloudflare or AWS Shield integration

**Monitoring & Alerting:**
```yaml
# Prometheus metrics
metrics:
  enabled: true
  port: 9090
  
alerts:
  - name: "validator_down"
    condition: "up == 0"
    duration: "30s"
    
  - name: "missed_blocks"
    condition: "missed_blocks > 10"
    duration: "5m"
    
  - name: "disk_space_low"
    condition: "disk_free < 20%"
    duration: "1m"
```

## Backup & Recovery

**Key Backup Strategy:**
1. **Mnemonic Phrase**: 24-word BIP39 seed phrase
2. **Hardware Backup**: HSM key backup to secure location
3. **Geographic Distribution**: Keys stored in multiple jurisdictions
4. **Recovery Testing**: Monthly recovery drills

**Node Backup:**
```bash
# Automated daily backups
#!/bin/bash
DATE=$(date +%Y%m%d)
tar -czf /backup/sharehodl-$DATE.tar.gz ~/.sharehodl/data
aws s3 cp /backup/sharehodl-$DATE.tar.gz s3://validator-backups/
```

**Disaster Recovery:**
- **RTO**: 15 minutes (Recovery Time Objective)
- **RPO**: 1 hour (Recovery Point Objective)
- **Hot Standby**: Secondary validator in different data center
- **Automated Failover**: Switch to backup node if primary fails

---

*This setup ensures ShareHODL validators provide institutional-grade security, compliance, and reliability for professional equity trading.*