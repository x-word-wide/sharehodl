#!/bin/bash

# Fix Scan/Explorer Service
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Fixing Scan/Explorer Service ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Creating explorer systemd service..."
run_remote "sudo tee /etc/systemd/system/sharehodl-explorer.service > /dev/null << 'EOF'
[Unit]
Description=ShareHODL Explorer/Scan Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=$REMOTE_DIR/sharehodl-frontend
Environment=PATH=/usr/bin:/usr/local/bin:/home/ubuntu/.local/share/pnpm
ExecStart=/usr/bin/pnpm --filter explorer dev
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF"

echo "Reloading systemd and starting explorer..."
run_remote "sudo systemctl daemon-reload"
run_remote "sudo systemctl enable sharehodl-explorer"
run_remote "sudo systemctl start sharehodl-explorer"

echo "Waiting for explorer to start..."
sleep 20

echo "Testing explorer service..."
run_remote "curl -s -o /dev/null -w 'Local port 3003: %{http_code}' http://localhost:3003 || echo 'Failed'"

echo ""
echo "Testing scan.sharehodl.com..."
echo "Scan: $(curl -s -o /dev/null -w '%{http_code}' https://scan.sharehodl.com)"

echo ""
echo "Checking service status..."
run_remote "sudo systemctl status sharehodl-explorer --no-pager -l"

echo "Scan service fix complete!"