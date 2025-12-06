#!/bin/bash

# Add Business Service
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Adding Business Service ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Creating business systemd service..."
run_remote "sudo tee /etc/systemd/system/sharehodl-business.service > /dev/null << 'EOF'
[Unit]
Description=ShareHODL Business Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=$REMOTE_DIR/sharehodl-frontend
Environment=PATH=/usr/bin:/usr/local/bin:/home/ubuntu/.local/share/pnpm
ExecStart=/usr/bin/pnpm --filter business dev
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF"

echo "Starting business service..."
run_remote "sudo systemctl daemon-reload"
run_remote "sudo systemctl enable sharehodl-business"
run_remote "sudo systemctl start sharehodl-business"

echo "Waiting for business service..."
sleep 15

echo "Testing all services now..."
echo "Governance: $(curl -s -o /dev/null -w '%{http_code}' https://gov.sharehodl.com)"
echo "Wallet: $(curl -s -o /dev/null -w '%{http_code}' https://wallet.sharehodl.com)"
echo "Trading: $(curl -s -o /dev/null -w '%{http_code}' https://trade.sharehodl.com)"
echo "Explorer/Scan: $(curl -s -o /dev/null -w '%{http_code}' https://scan.sharehodl.com)"
echo "Business: $(curl -s -o /dev/null -w '%{http_code}' https://business.sharehodl.com)"

echo "All services complete!"