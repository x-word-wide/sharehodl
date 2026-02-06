#!/bin/bash

# ShareHODL Blockchain Deployment Script
# For Ubuntu 24.04 - Run as sharehodl user (NOT root)
# Usage: bash 02-deploy-sharehodl.sh

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Configuration
DEPLOY_DIR="/opt/sharehodl"
CHAIN_ID="sharehodl-mainnet"
NODE_NAME="sharehodl-mainnet-validator"
SERVER_IP="178.63.13.190"
GO_VERSION="1.23.5"

echo "=============================================="
echo "  ShareHODL Blockchain Deployment"
echo "  Production Environment"
echo "=============================================="
echo ""

# Check if running as non-root
if [[ $EUID -eq 0 ]]; then
   log_error "Do not run this script as root. Run as sharehodl user with sudo privileges."
fi

# Step 1: Install Docker
log_info "Step 1: Installing Docker..."
if command -v docker &> /dev/null; then
    log_warning "Docker already installed"
else
    # Install Docker prerequisites
    sudo apt-get update
    sudo apt-get install -y \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release

    # Add Docker's official GPG key
    sudo install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    sudo chmod a+r /etc/apt/keyrings/docker.gpg

    # Set up the repository
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
      sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    # Install Docker
    sudo apt-get update
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    # Add user to docker group
    sudo usermod -aG docker $USER

    log_success "Docker installed"
    log_warning "You may need to log out and back in for docker group to take effect"
fi

# Step 2: Install Docker Compose
log_info "Step 2: Verifying Docker Compose..."
if docker compose version &> /dev/null; then
    log_success "Docker Compose available"
else
    log_error "Docker Compose not available. Please reinstall Docker."
fi

# Step 3: Install Go (for building from source)
log_info "Step 3: Installing Go $GO_VERSION..."
if command -v go &> /dev/null; then
    CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
    log_warning "Go $CURRENT_GO already installed"
else
    cd /tmp
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
    rm "go${GO_VERSION}.linux-amd64.tar.gz"

    # Add to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
    echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

    log_success "Go $GO_VERSION installed"
fi

# Step 4: Install Node.js and pnpm
log_info "Step 4: Installing Node.js and pnpm..."
if command -v node &> /dev/null; then
    log_warning "Node.js already installed: $(node -v)"
else
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
    sudo apt-get install -y nodejs
    sudo npm install -g pnpm
    log_success "Node.js and pnpm installed"
fi

# Step 5: Clone or update repository
log_info "Step 5: Setting up ShareHODL repository..."
cd $DEPLOY_DIR

if [[ -d "$DEPLOY_DIR/sharehodl-blockchain" ]]; then
    log_warning "Repository exists, pulling latest..."
    cd sharehodl-blockchain
    git fetch origin
    git pull origin main || log_warning "Could not pull, continuing with existing code"
else
    log_info "Cloning repository..."
    # Replace with your actual repo URL
    git clone https://github.com/sharehodl/sharehodl-blockchain.git
    cd sharehodl-blockchain
fi

# Step 6: Create production environment file
log_info "Step 6: Creating production environment file..."
cat > .env <<EOF
# ShareHODL Production Environment
CHAIN_ID=$CHAIN_ID
NODE_NAME=$NODE_NAME
SERVER_IP=$SERVER_IP

# Database credentials (change these!)
POSTGRES_PASSWORD=$(openssl rand -base64 32)
REDIS_PASSWORD=$(openssl rand -base64 32)

# API Settings
MINIMUM_GAS_PRICES=0.0025uhodl
LOG_LEVEL=info

# Frontend URLs
NEXT_PUBLIC_API_URL=http://$SERVER_IP:1317
NEXT_PUBLIC_RPC_URL=http://$SERVER_IP:26657
NEXT_PUBLIC_CHAIN_ID=$CHAIN_ID
NEXT_PUBLIC_CHAIN_NAME=ShareHODL Mainnet
NEXT_PUBLIC_DENOM=uhodl
EOF

chmod 600 .env
log_success "Environment file created"
log_warning "IMPORTANT: Review and customize .env before deploying!"

# Step 7: Build blockchain binary
log_info "Step 7: Building ShareHODL blockchain..."
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
make build
log_success "Blockchain binary built"

# Step 8: Initialize the blockchain
log_info "Step 8: Initializing blockchain..."
SHAREHODL_HOME="$HOME/.sharehodl"

if [[ -d "$SHAREHODL_HOME/config" ]]; then
    log_warning "Blockchain already initialized at $SHAREHODL_HOME"
else
    ./build/sharehodld init "$NODE_NAME" --chain-id "$CHAIN_ID" --home "$SHAREHODL_HOME"

    # Create validator key
    ./build/sharehodld keys add validator --keyring-backend file --home "$SHAREHODL_HOME"

    # Save validator address
    VALIDATOR_ADDR=$(./build/sharehodld keys show validator -a --keyring-backend file --home "$SHAREHODL_HOME")
    echo "VALIDATOR_ADDRESS=$VALIDATOR_ADDR" >> .env

    log_success "Blockchain initialized"
    log_info "Validator address: $VALIDATOR_ADDR"
fi

# Step 9: Configure the blockchain
log_info "Step 9: Configuring blockchain..."
CONFIG_FILE="$SHAREHODL_HOME/config/config.toml"
APP_FILE="$SHAREHODL_HOME/config/app.toml"

# Enable API
sed -i 's/enable = false/enable = true/g' "$APP_FILE"
sed -i 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/0.0.0.0:1317"/' "$APP_FILE"
sed -i 's/address = "localhost:9090"/address = "0.0.0.0:9090"/' "$APP_FILE"

# Configure RPC
sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' "$CONFIG_FILE"

# Set external address
sed -i "s/external_address = \"\"/external_address = \"$SERVER_IP:26656\"/" "$CONFIG_FILE"

# Enable prometheus metrics
sed -i 's/prometheus = false/prometheus = true/' "$CONFIG_FILE"

# Set log level
sed -i 's/log_level = "info"/log_level = "warn"/' "$CONFIG_FILE"

log_success "Blockchain configured"

# Step 10: Create systemd service
log_info "Step 10: Creating systemd service..."
sudo tee /etc/systemd/system/sharehodld.service > /dev/null <<EOF
[Unit]
Description=ShareHODL Blockchain Node
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME
ExecStart=$DEPLOY_DIR/sharehodl-blockchain/build/sharehodld start --home $SHAREHODL_HOME
Restart=on-failure
RestartSec=5
LimitNOFILE=65535
StandardOutput=journal
StandardError=journal
SyslogIdentifier=sharehodld

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=$SHAREHODL_HOME /var/log/sharehodl

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable sharehodld
log_success "Systemd service created"

# Step 11: Setup log rotation
log_info "Step 11: Setting up log rotation..."
sudo tee /etc/logrotate.d/sharehodl > /dev/null <<EOF
/var/log/sharehodl/*.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 $USER $USER
    sharedscripts
    postrotate
        systemctl reload sharehodld > /dev/null 2>&1 || true
    endscript
}
EOF
log_success "Log rotation configured"

# Step 12: Build and prepare frontend
log_info "Step 12: Building frontend applications..."
cd $DEPLOY_DIR/sharehodl-blockchain/sharehodl-frontend

# Install dependencies
pnpm install

# Build for production
pnpm build || log_warning "Frontend build failed, check errors above"

log_success "Frontend built"

# Step 13: Create frontend systemd service
log_info "Step 13: Creating frontend service..."
sudo tee /etc/systemd/system/sharehodl-frontend.service > /dev/null <<EOF
[Unit]
Description=ShareHODL Frontend Applications
After=network.target sharehodld.service
Wants=sharehodld.service

[Service]
Type=simple
User=$USER
WorkingDirectory=$DEPLOY_DIR/sharehodl-blockchain/sharehodl-frontend
ExecStart=/usr/bin/pnpm start
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=sharehodl-frontend

# Environment
Environment=NODE_ENV=production
Environment=NEXT_PUBLIC_API_URL=http://$SERVER_IP:1317
Environment=NEXT_PUBLIC_RPC_URL=http://$SERVER_IP:26657
Environment=NEXT_PUBLIC_CHAIN_ID=$CHAIN_ID

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable sharehodl-frontend
log_success "Frontend service created"

# Step 14: Install and configure Nginx
log_info "Step 14: Installing and configuring Nginx..."
sudo apt install -y nginx certbot python3-certbot-nginx

# Create nginx config
sudo tee /etc/nginx/sites-available/sharehodl > /dev/null <<EOF
# ShareHODL Production Nginx Configuration

# Rate limiting
limit_req_zone \$binary_remote_addr zone=api_limit:10m rate=100r/s;
limit_conn_zone \$binary_remote_addr zone=conn_limit:10m;

# Upstream servers
upstream blockchain_api {
    server 127.0.0.1:1317;
    keepalive 32;
}

upstream blockchain_rpc {
    server 127.0.0.1:26657;
    keepalive 32;
}

upstream frontend_home {
    server 127.0.0.1:3001;
}

upstream frontend_explorer {
    server 127.0.0.1:3002;
}

upstream frontend_trading {
    server 127.0.0.1:3003;
}

upstream frontend_wallet {
    server 127.0.0.1:3004;
}

upstream frontend_governance {
    server 127.0.0.1:3005;
}

upstream frontend_business {
    server 127.0.0.1:3006;
}

# Main server
server {
    listen 80;
    server_name $SERVER_IP _;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Home / Landing
    location / {
        proxy_pass http://frontend_home;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Explorer
    location /explorer {
        proxy_pass http://frontend_explorer;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Trading
    location /trading {
        proxy_pass http://frontend_trading;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Wallet
    location /wallet {
        proxy_pass http://frontend_wallet;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Governance
    location /governance {
        proxy_pass http://frontend_governance;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Business Portal
    location /business {
        proxy_pass http://frontend_business;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Blockchain REST API
    location /api/ {
        limit_req zone=api_limit burst=50 nodelay;
        limit_conn conn_limit 20;

        rewrite ^/api/(.*)$ /\$1 break;
        proxy_pass http://blockchain_api;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }

    # Cosmos SDK API
    location /cosmos/ {
        limit_req zone=api_limit burst=50 nodelay;
        proxy_pass http://blockchain_api;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    # RPC endpoint
    location /rpc/ {
        limit_req zone=api_limit burst=20 nodelay;
        rewrite ^/rpc/(.*)$ /\$1 break;
        proxy_pass http://blockchain_rpc;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # WebSocket for real-time updates
    location /websocket {
        proxy_pass http://blockchain_rpc;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 86400;
    }

    # Health check
    location /health {
        access_log off;
        return 200 'OK';
        add_header Content-Type text/plain;
    }
}
EOF

# Enable site
sudo ln -sf /etc/nginx/sites-available/sharehodl /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Test and reload nginx
sudo nginx -t && sudo systemctl reload nginx
log_success "Nginx configured"

# Summary
echo ""
echo "=============================================="
echo "  DEPLOYMENT COMPLETE"
echo "=============================================="
echo ""
log_success "ShareHODL blockchain deployment completed!"
echo ""
echo "NEXT STEPS:"
echo ""
echo "1. Review configuration:"
echo "   cat $DEPLOY_DIR/sharehodl-blockchain/.env"
echo ""
echo "2. Initialize genesis with test accounts (if new chain):"
echo "   Run the genesis setup script"
echo ""
echo "3. Start the blockchain:"
echo "   sudo systemctl start sharehodld"
echo ""
echo "4. Start the frontend:"
echo "   sudo systemctl start sharehodl-frontend"
echo ""
echo "5. Check status:"
echo "   sudo systemctl status sharehodld"
echo "   curl http://localhost:26657/status"
echo ""
echo "ENDPOINTS:"
echo "  - Frontend: http://$SERVER_IP"
echo "  - RPC: http://$SERVER_IP:26657"
echo "  - REST API: http://$SERVER_IP:1317"
echo "  - gRPC: $SERVER_IP:9090"
echo ""
echo "MONITORING:"
echo "  - Logs: journalctl -u sharehodld -f"
echo "  - Status: curl http://localhost:26657/status | jq"
echo ""
