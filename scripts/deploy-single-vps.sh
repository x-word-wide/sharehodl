#!/bin/bash

# ShareHODL Single VPS Deployment Script
# Run this on a fresh Ubuntu 22.04 server

set -e

log_info() {
    echo -e "\033[0;34m[INFO]\033[0m $1"
}

log_success() {
    echo -e "\033[0;32m[SUCCESS]\033[0m $1"
}

log_error() {
    echo -e "\033[0;31m[ERROR]\033[0m $1"
}

# Update system
log_info "Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install dependencies
log_info "Installing dependencies..."
sudo apt install -y build-essential git curl jq nginx certbot python3-certbot-nginx ufw

# Install Go
log_info "Installing Go 1.21..."
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin

# Install Node.js for frontend
log_info "Installing Node.js and pnpm..."
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs
npm install -g pnpm

# Clone ShareHODL
log_info "Cloning ShareHODL repository..."
git clone https://github.com/sharehodl/sharehodl-blockchain.git
cd sharehodl-blockchain

# Build blockchain
log_info "Building ShareHODL blockchain..."
make install

# Initialize single-node testnet
log_info "Initializing single-node testnet..."
sharehodld init sharehodl-testnet --chain-id sharehodl-single-testnet

# Create validator account
log_info "Creating validator account..."
sharehodld keys add validator --keyring-backend test

# Get validator address
VALIDATOR_ADDR=$(sharehodld keys show validator -a --keyring-backend test)
log_info "Validator address: $VALIDATOR_ADDR"

# Add genesis account with tokens
log_info "Adding genesis account..."
sharehodld add-genesis-account $VALIDATOR_ADDR 1000000000000stake,1000000000000hodl

# Generate validator transaction
log_info "Generating validator transaction..."
sharehodld gentx validator 1000000000stake \
  --keyring-backend test \
  --chain-id sharehodl-single-testnet \
  --commission-rate 0.1 \
  --commission-max-rate 0.2 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1

# Collect genesis transactions
sharehodld collect-gentxs

# Configure API and gRPC
log_info "Configuring API services..."
sed -i 's/enable = false/enable = true/' ~/.sharehodl/config/app.toml
sed -i 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/0.0.0.0:1317"/' ~/.sharehodl/config/app.toml
sed -i 's/address = "localhost:9090"/address = "0.0.0.0:9090"/' ~/.sharehodl/config/app.toml

# Configure external access
sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' ~/.sharehodl/config/config.toml

# Create systemd service
log_info "Creating systemd service..."
sudo tee /etc/systemd/system/sharehodld.service > /dev/null <<EOF
[Unit]
Description=ShareHODL Blockchain
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME
ExecStart=$(which sharehodld) start --home $HOME/.sharehodl
Restart=on-failure
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# Start blockchain service
log_info "Starting ShareHODL blockchain..."
sudo systemctl daemon-reload
sudo systemctl enable sharehodld
sudo systemctl start sharehodld

# Build and start frontend
log_info "Building frontend applications..."
cd sharehodl-frontend

# Install frontend dependencies
pnpm install

# Start frontend in production mode
log_info "Starting frontend services..."
nohup pnpm dev > /tmp/frontend.log 2>&1 &

# Configure nginx reverse proxy
log_info "Configuring Nginx..."
sudo tee /etc/nginx/sites-available/sharehodl > /dev/null <<EOF
server {
    listen 80;
    server_name _;

    # Frontend apps
    location / {
        proxy_pass http://localhost:3001;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    location /governance {
        proxy_pass http://localhost:3001;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    location /trading {
        proxy_pass http://localhost:3002;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    location /explorer {
        proxy_pass http://localhost:3003;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    location /features {
        proxy_pass http://localhost:3004;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    location /business {
        proxy_pass http://localhost:3005;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    # Blockchain API
    location /api {
        proxy_pass http://localhost:1317;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    # RPC endpoint
    location /rpc {
        proxy_pass http://localhost:26657;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF

# Enable nginx site
sudo ln -sf /etc/nginx/sites-available/sharehodl /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl reload nginx

# Configure firewall
log_info "Configuring firewall..."
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 26657/tcp
sudo ufw allow 1317/tcp
sudo ufw --force enable

log_success "ShareHODL single-node testnet deployed successfully!"
log_info "Services running:"
log_info "  - Blockchain: http://YOUR_VPS_IP:26657/status"
log_info "  - API: http://YOUR_VPS_IP:1317/node_info"
log_info "  - Frontend: http://YOUR_VPS_IP"
log_info "  - Governance: http://YOUR_VPS_IP/governance"
log_info "  - Trading: http://YOUR_VPS_IP/trading"
log_info "  - Explorer: http://YOUR_VPS_IP/explorer"

log_info "To check status:"
log_info "  sudo systemctl status sharehodld"
log_info "  curl http://localhost:26657/status"

log_info "To get test tokens:"
log_info "  sharehodld keys show validator -a --keyring-backend test"
log_info "  # This address has 1T stake and 1T hodl tokens"