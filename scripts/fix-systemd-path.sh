#!/bin/bash

# Fix Systemd Service Paths
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Fixing Systemd Service Paths ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Stopping services..."
run_remote "sudo systemctl stop sharehodl-governance sharehodl-wallet sharehodl-trading || true"

echo "Updating service files with correct paths..."

# Update governance service
run_remote "sudo tee /etc/systemd/system/sharehodl-governance.service > /dev/null << 'EOF'
[Unit]
Description=ShareHODL Governance Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=$REMOTE_DIR/sharehodl-frontend
Environment=PATH=/usr/bin:/usr/local/bin:/home/ubuntu/.local/share/pnpm
ExecStart=/usr/bin/pnpm --filter governance dev
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF"

# Update wallet service  
run_remote "sudo tee /etc/systemd/system/sharehodl-wallet.service > /dev/null << 'EOF'
[Unit]
Description=ShareHODL Wallet Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=$REMOTE_DIR/sharehodl-frontend
Environment=PATH=/usr/bin:/usr/local/bin:/home/ubuntu/.local/share/pnpm
ExecStart=/usr/bin/pnpm --filter wallet dev
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF"

# Update trading service
run_remote "sudo tee /etc/systemd/system/sharehodl-trading.service > /dev/null << 'EOF'
[Unit]
Description=ShareHODL Trading Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=$REMOTE_DIR/sharehodl-frontend
Environment=PATH=/usr/bin:/usr/local/bin:/home/ubuntu/.local/share/pnpm
ExecStart=/usr/bin/pnpm --filter trading dev
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF"

echo "Reloading systemd..."
run_remote "sudo systemctl daemon-reload"

echo "Starting services..."
run_remote "sudo systemctl start sharehodl-governance"
run_remote "sudo systemctl start sharehodl-wallet" 
run_remote "sudo systemctl start sharehodl-trading"

echo "Waiting for startup..."
sleep 25

echo "Testing services..."
echo "Governance: $(curl -s -o /dev/null -w '%{http_code}' https://gov.sharehodl.com)"
echo "Wallet: $(curl -s -o /dev/null -w '%{http_code}' https://wallet.sharehodl.com)"
echo "Trading: $(curl -s -o /dev/null -w '%{http_code}' https://trade.sharehodl.com)"

echo "Systemd services fixed!"