#!/bin/bash

# Try Development Mode
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Trying Development Mode ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Starting wallet in development mode on port 3004..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && timeout 30s pnpm --filter wallet dev 2>&1 || echo 'Dev mode failed'" &

sleep 20

echo ""
echo "Testing wallet dev mode..."
run_remote "curl -s -o /dev/null -w 'Wallet dev: %{http_code}' http://localhost:3004 || echo 'Failed'"

echo ""
echo "Testing HTTPS..."
curl -s -o /dev/null -w "HTTPS wallet: %{http_code}\n" https://wallet.sharehodl.com

echo ""
echo "Dev mode test complete."