# ShareHODL Blockchain Deployment Guide

## Server Details
- **Provider**: Spacecore/Hetzner
- **Server ID**: #387304
- **Hostname**: 387304.fsn.dedic.hosted-by-spacecore.pro
- **IP Address**: 178.63.13.190
- **Hardware**: Intel Core i5-12500 (6c/12t @ 4.60GHz), RAID1
- **OS**: Ubuntu 24.04 x64

## Quick Start (TL;DR)

```bash
# 1. SSH to server as root
ssh root@178.63.13.190

# 2. Upload and run hardening script
bash 01-server-hardening.sh sharehodl 2222

# 3. In NEW terminal, add your SSH key
ssh-copy-id -p 2222 sharehodl@178.63.13.190

# 4. Test SSH key login, then continue as sharehodl user
ssh -p 2222 sharehodl@178.63.13.190

# 5. Deploy blockchain
bash 02-deploy-sharehodl.sh

# 6. Setup genesis
bash 04-genesis-setup.sh

# 7. Start services
sudo systemctl start sharehodld
sudo systemctl start sharehodl-frontend

# 8. Verify deployment
bash 03-verify-deployment.sh
```

---

## Detailed Deployment Steps

### Phase 1: Initial Server Access

1. **Connect to the server**:
   ```bash
   ssh root@178.63.13.190
   # Password: CphKCyNAvJn6we
   ```

2. **Upload deployment scripts**:
   From your local machine:
   ```bash
   # Option A: Copy scripts directly
   scp -r scripts/deploy/* root@178.63.13.190:/root/

   # Option B: Clone repo on server
   ssh root@178.63.13.190 "apt update && apt install -y git && git clone <repo> /tmp/sharehodl"
   ```

### Phase 2: Server Hardening (CRITICAL)

**Run the hardening script as root:**

```bash
cd /root
chmod +x 01-server-hardening.sh
bash 01-server-hardening.sh sharehodl 2222
```

This script will:
- Create a new user `sharehodl` with sudo privileges
- Change SSH port to 2222
- Disable root SSH login
- Disable password authentication
- Configure UFW firewall
- Install and configure fail2ban
- Set kernel security parameters
- Enable automatic security updates

**IMPORTANT: Note the generated password!**

```
===========================================
  NEW USER CREDENTIALS
  Username: sharehodl
  Password: <generated-password>
===========================================
```

### Phase 3: Setup SSH Key Authentication

**Do NOT close your root terminal yet!**

1. **Open a NEW terminal on your local machine**

2. **Copy your SSH public key**:
   ```bash
   # If you don't have an SSH key, create one:
   ssh-keygen -t ed25519 -C "sharehodl-deployment"

   # Copy to server
   ssh-copy-id -p 2222 sharehodl@178.63.13.190
   ```

3. **Test SSH key login**:
   ```bash
   ssh -p 2222 sharehodl@178.63.13.190
   # Should login without password
   ```

4. **Verify sudo works**:
   ```bash
   sudo whoami
   # Should output: root
   ```

5. **Now restart SSH on the server** (in root terminal):
   ```bash
   systemctl restart sshd
   ```

6. **Verify you can still connect** in another terminal:
   ```bash
   ssh -p 2222 sharehodl@178.63.13.190
   ```

### Phase 4: Deploy ShareHODL Blockchain

**As the `sharehodl` user:**

```bash
# Navigate to deployment directory
cd /opt/sharehodl

# Upload or clone the scripts
sudo cp /root/02-deploy-sharehodl.sh .
sudo cp /root/03-verify-deployment.sh .
sudo cp /root/04-genesis-setup.sh .
sudo chown sharehodl:sharehodl *.sh
chmod +x *.sh

# Run deployment
bash 02-deploy-sharehodl.sh
```

**Note**: You may need to log out and back in after Docker installation for group changes to take effect.

### Phase 5: Setup Genesis and Start Chain

```bash
# Setup genesis accounts and validator
bash 04-genesis-setup.sh

# Start the blockchain
sudo systemctl start sharehodld

# Wait for it to start (check logs)
journalctl -u sharehodld -f
# Press Ctrl+C when you see blocks being produced

# Start frontend
sudo systemctl start sharehodl-frontend
```

### Phase 6: Verify Deployment

```bash
# Run verification script
bash 03-verify-deployment.sh

# Manual checks
curl http://localhost:26657/status | jq '.result.sync_info'
curl http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info | jq '.node_info'
```

---

## Service Management

### Start/Stop Services

```bash
# Blockchain
sudo systemctl start sharehodld
sudo systemctl stop sharehodld
sudo systemctl restart sharehodld

# Frontend
sudo systemctl start sharehodl-frontend
sudo systemctl stop sharehodl-frontend
sudo systemctl restart sharehodl-frontend

# Nginx
sudo systemctl reload nginx
```

### View Logs

```bash
# Blockchain logs
journalctl -u sharehodld -f

# Frontend logs
journalctl -u sharehodl-frontend -f

# Nginx logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

### Check Status

```bash
# All services
sudo systemctl status sharehodld sharehodl-frontend nginx

# Blockchain status
curl -s http://localhost:26657/status | jq '.result.sync_info'

# Node info
curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info
```

---

## Endpoints

| Service | Internal | External |
|---------|----------|----------|
| Frontend Home | http://localhost:3001 | http://178.63.13.190 |
| Explorer | http://localhost:3002 | http://178.63.13.190/explorer |
| Trading | http://localhost:3003 | http://178.63.13.190/trading |
| Wallet | http://localhost:3004 | http://178.63.13.190/wallet |
| Governance | http://localhost:3005 | http://178.63.13.190/governance |
| Business | http://localhost:3006 | http://178.63.13.190/business |
| REST API | http://localhost:1317 | http://178.63.13.190:1317 |
| RPC | http://localhost:26657 | http://178.63.13.190:26657 |
| gRPC | localhost:9090 | 178.63.13.190:9090 |
| P2P | localhost:26656 | 178.63.13.190:26656 |

---

## Firewall Rules

The UFW firewall is configured with these rules:

| Port | Protocol | Purpose |
|------|----------|---------|
| 2222 | TCP | SSH (custom port) |
| 80 | TCP | HTTP (Nginx) |
| 443 | TCP | HTTPS (Nginx) |
| 26656 | TCP | Tendermint P2P |
| 26657 | TCP | Tendermint RPC |
| 1317 | TCP | Cosmos REST API |
| 9090 | TCP | gRPC |

---

## Security Best Practices

1. **SSH Key Only**: Password authentication is disabled. Always use SSH keys.

2. **Non-Standard SSH Port**: SSH runs on port 2222 to reduce automated attacks.

3. **Fail2ban**: Automatically bans IPs with too many failed login attempts.

4. **UFW Firewall**: Only necessary ports are open.

5. **Automatic Updates**: Security updates are automatically installed.

6. **Limited User**: Services run as non-root `sharehodl` user.

---

## Backup Procedures

### Backup Validator Keys

```bash
# Critical files to backup
cp ~/.sharehodl/config/priv_validator_key.json /secure/backup/
cp ~/.sharehodl/config/node_key.json /secure/backup/
```

### Backup Chain Data

```bash
# Stop the node first
sudo systemctl stop sharehodld

# Backup data directory
tar -czvf sharehodl-backup-$(date +%Y%m%d).tar.gz ~/.sharehodl/data

# Restart
sudo systemctl start sharehodld
```

---

## Troubleshooting

### Blockchain won't start

```bash
# Check logs
journalctl -u sharehodld -n 100 --no-pager

# Validate genesis
/opt/sharehodl/sharehodl-blockchain/build/sharehodld genesis validate --home ~/.sharehodl

# Check config
cat ~/.sharehodl/config/config.toml | grep -E "^laddr|^external"
```

### Frontend not accessible

```bash
# Check if running
sudo systemctl status sharehodl-frontend

# Check nginx
sudo nginx -t
sudo systemctl status nginx

# Check ports
sudo ss -tlnp | grep -E ':(3001|3002|3003|3004|3005|3006)'
```

### Can't connect via SSH

1. Use the Hetzner/Spacecore web console
2. Check if SSH is running: `systemctl status sshd`
3. Check firewall: `ufw status`
4. Emergency revert: `bash /root/revert-ssh.sh`

---

## SSL/TLS Setup (Optional)

Once you have a domain name:

```bash
# Install certificate
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com

# Auto-renewal is already configured
sudo systemctl status certbot.timer
```

---

## Monitoring (Optional)

### Setup Prometheus + Grafana

```bash
# Prometheus metrics are enabled on port 26660
curl http://localhost:26660/metrics

# Install Prometheus (optional)
# Install Grafana (optional)
# Import Cosmos SDK dashboards
```

---

## Support

- **Logs Location**: `/var/log/sharehodl/`
- **Config Location**: `~/.sharehodl/config/`
- **Binary Location**: `/opt/sharehodl/sharehodl-blockchain/build/sharehodld`
- **Service Files**: `/etc/systemd/system/sharehodld.service`
