# ShareHODL Production Server

## Server Details
| Field | Value |
|-------|-------|
| IP Address | 178.63.13.190 |
| Provider | Hetzner (Server #387304) |
| OS | Ubuntu 22.04 LTS |
| SSH Port | 22 |

## SSH Connection
```bash
ssh root@178.63.13.190
```

## SSH Key (Authorized)
```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIUn1RkM5VHt2dc5YfvbX814RS+Fs8ukmJXIM7gnlIT/ sharehodl-server
```

## Domains (All with HTTPS/SSL)
| Domain | Service | Internal Port |
|--------|---------|---------------|
| https://sharehodl.com | Home Frontend | 3000 |
| https://rpc.sharehodl.com | Tendermint RPC | 26657 |
| https://api.sharehodl.com | REST API | 1317 |
| https://business.sharehodl.com | Business Portal | 3006 |
| https://gov.sharehodl.com | Governance Portal | 3005 |

## Services (Systemd)
| Service | Status | Port |
|---------|--------|------|
| sharehodld | Active | 26656, 26657, 1317, 9090 |
| sharehodl-home | Active | 3000 |
| nginx | Active | 80, 443 |

## Deployment

### Quick Deploy (Git Pull + Restart)
```bash
ssh root@178.63.13.190 '/root/deploy.sh all'
```

### Deploy Options
```bash
# Blockchain only
/root/deploy.sh blockchain

# Frontend only
/root/deploy.sh frontend

# Everything
/root/deploy.sh all
```

### Manual Commands
```bash
# Check blockchain status
curl https://rpc.sharehodl.com/status | jq '.result.sync_info'

# View blockchain logs
journalctl -u sharehodld -f

# Restart blockchain
systemctl restart sharehodld

# Restart frontend
systemctl restart sharehodl-home

# Check SSL certificates
certbot certificates
```

## Directory Structure
```
/root/
├── sharehodl-blockchain/      # Git repository
│   ├── sharehodl-frontend/    # Frontend monorepo
│   └── x/                     # Blockchain modules
├── go/bin/sharehodld          # Blockchain binary
└── deploy.sh                  # Deployment script
```

## Chain Configuration
- **Chain ID**: sharehodl-mainnet
- **Validator**: hodl12uly3carpzggk39k0sh3aalhwxzfycn3xhn9pc
- **Denom**: uhodl (1 HODL = 1,000,000 uhodl)

## Security Notes
- SSH key auth only (no password)
- All domains use HTTPS with Let's Encrypt
- Certbot auto-renews certificates
- Blockchain RPC is public (read-only operations)
