#!/bin/bash

# Create Systemd Services for Permanent Solution
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=========================================="
echo "CREATING SYSTEMD SERVICES - PERMANENT FIX"
echo "=========================================="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Creating systemd service files..."

# Create governance service
run_remote "sudo tee /etc/systemd/system/sharehodl-governance.service > /dev/null << 'EOF'
[Unit]
Description=ShareHODL Governance Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=$REMOTE_DIR/sharehodl-frontend
Environment=PATH=/usr/bin:/usr/local/bin
Environment=NODE_ENV=development
ExecStart=/usr/local/bin/pnpm --filter governance dev
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF"

# Create wallet service
run_remote "sudo tee /etc/systemd/system/sharehodl-wallet.service > /dev/null << 'EOF'
[Unit]
Description=ShareHODL Wallet Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=$REMOTE_DIR/sharehodl-frontend
Environment=PATH=/usr/bin:/usr/local/bin
Environment=NODE_ENV=development
ExecStart=/usr/local/bin/pnpm --filter wallet dev
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF"

# Create trading service
run_remote "sudo tee /etc/systemd/system/sharehodl-trading.service > /dev/null << 'EOF'
[Unit]
Description=ShareHODL Trading Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=$REMOTE_DIR/sharehodl-frontend
Environment=PATH=/usr/bin:/usr/local/bin
Environment=NODE_ENV=development
ExecStart=/usr/local/bin/pnpm --filter trading dev
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF"

echo "Reloading systemd and starting services..."
run_remote "sudo systemctl daemon-reload"
run_remote "sudo systemctl enable sharehodl-governance sharehodl-wallet sharehodl-trading"
run_remote "sudo systemctl start sharehodl-governance sharehodl-wallet sharehodl-trading"

echo ""
echo "Waiting for services to start..."
sleep 20

echo ""
echo "Checking service status..."
run_remote "sudo systemctl status sharehodl-governance --no-pager -l"
run_remote "sudo systemctl status sharehodl-wallet --no-pager -l"
run_remote "sudo systemctl status sharehodl-trading --no-pager -l"

echo ""
echo "Testing all services..."
echo "Governance: $(curl -s -o /dev/null -w '%{http_code}' https://gov.sharehodl.com)"
echo "Wallet: $(curl -s -o /dev/null -w '%{http_code}' https://wallet.sharehodl.com)"
echo "Trading: $(curl -s -o /dev/null -w '%{http_code}' https://trade.sharehodl.com)"

echo ""
echo "=========================================="
echo "SYSTEMD SERVICES CREATED!"
echo "=========================================="
echo "Services will now auto-start on boot and restart if they crash."
echo "To manage: sudo systemctl [start|stop|restart] sharehodl-governance"