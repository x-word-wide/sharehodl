# ShareHODL Single VPS Deployment Guide

## 🚀 Deploy Everything on One Server

Yes! You can run the entire ShareHODL testnet on a single VPS. Perfect for testing and demonstrations.

## VPS Requirements

**Minimum Specs:**
- **CPU**: 4 cores
- **RAM**: 8GB
- **Storage**: 200GB SSD
- **Bandwidth**: 1TB/month
- **Cost**: $40-80/month

**Recommended Providers:**
- DigitalOcean: $80/month (8GB RAM, 4 cores)
- Vultr: $60/month (8GB RAM, 4 cores)  
- Hetzner: €40/month (8GB RAM, 4 cores)

## One-Command Deployment

**Step 1: Get a VPS**
1. Create Ubuntu 22.04 server
2. Note your server IP address
3. SSH into your server

**Step 2: Run Deploy Script**
```bash
# SSH into your VPS
ssh root@YOUR_VPS_IP

# Download and run deployment script
wget https://raw.githubusercontent.com/sharehodl/sharehodl-blockchain/main/scripts/deploy-single-vps.sh
chmod +x deploy-single-vps.sh
./deploy-single-vps.sh
```

**That's it! The script will:**
- Install all dependencies (Go, Node.js, Nginx)
- Build ShareHODL blockchain
- Create single-node testnet with validator
- Start all 5 frontend applications
- Configure reverse proxy
- Set up firewall and security

## Services After Deployment

**Your VPS will run:**
```
Blockchain Node:     :26657 (RPC)
REST API:           :1317  (API)
gRPC:               :9090  (gRPC)
Governance Portal:  :3001  (Frontend)
Trading Platform:   :3002  (Frontend) 
Explorer:           :3003  (Frontend)
Advanced Features:  :3004  (Frontend)
Business Portal:    :3005  (Frontend)
Web Gateway:        :80    (Nginx)
```

## Access Your Testnet

**After deployment, visit:**
- **Main Site**: `http://YOUR_VPS_IP`
- **Governance**: `http://YOUR_VPS_IP/governance`
- **Trading**: `http://YOUR_VPS_IP/trading` 
- **Explorer**: `http://YOUR_VPS_IP/explorer`
- **Business**: `http://YOUR_VPS_IP/business`
- **API**: `http://YOUR_VPS_IP/api/node_info`
- **RPC**: `http://YOUR_VPS_IP/rpc/status`

## Test Account Access

**Your validator has test tokens:**
```bash
# Get validator address (has 1T stake + 1T hodl)
sharehodld keys show validator -a --keyring-backend test

# Check balance
sharehodld query bank balances $(sharehodld keys show validator -a --keyring-backend test)

# Send tokens to test addresses in frontend
sharehodld tx bank send validator cosmos1... 1000000hodl --keyring-backend test --chain-id sharehodl-single-testnet
```

## Quick Health Check

**Verify deployment:**
```bash
# Check blockchain status
sudo systemctl status sharehodld
curl http://localhost:26657/status

# Check frontend apps
curl http://localhost:3001 # Governance
curl http://localhost:3002 # Trading
curl http://localhost:3003 # Explorer

# Check nginx
sudo systemctl status nginx
curl http://localhost/
```

## Custom Domain (Optional)

**Add your domain:**
```bash
# Point your domain to VPS IP
# Update nginx config
sudo nano /etc/nginx/sites-available/sharehodl
# Change: server_name _;
# To: server_name yourdomain.com;

# Get SSL certificate
sudo certbot --nginx -d yourdomain.com

# Restart nginx
sudo systemctl reload nginx
```

## Monitoring & Logs

**Check logs:**
```bash
# Blockchain logs
sudo journalctl -u sharehodld -f

# Frontend logs  
tail -f /tmp/frontend.log

# Nginx logs
sudo tail -f /var/log/nginx/access.log
```

## Costs

**Single VPS Total Cost:**
- VPS: $40-80/month
- Domain (optional): $12/year
- **Total: ~$50/month**

**Much cheaper than multi-node setup!**

## Scaling Later

**When ready to scale:**
1. Add more validator VPS
2. Set up load balancer
3. Separate frontend hosting
4. Add monitoring stack

**For now, single VPS is perfect for:**
- Testing all features
- Demo to investors  
- Community testing
- Development work
- Proof of concept

Your single VPS will run the complete ShareHODL ecosystem with all professional trading features, governance, and business tools!