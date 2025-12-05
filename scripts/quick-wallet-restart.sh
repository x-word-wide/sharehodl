#!/bin/bash

# Quick Wallet Restart in Dev Mode
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Quick Wallet Restart ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Killing any existing wallet processes..."
run_remote "pkill -f wallet || true"
sleep 3

echo "Starting wallet in development mode..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter wallet dev > /tmp/wallet-dev.log 2>&1 & echo \$! > /tmp/wallet-dev.pid"

echo "Waiting for wallet to start..."
sleep 15

echo "Testing wallet..."
run_remote "curl -s -o /dev/null -w 'Local: %{http_code}' http://localhost:3004 || echo 'Failed'"

echo ""
echo "Testing HTTPS..."
curl -s -o /dev/null -w "HTTPS: %{http_code}\n" https://wallet.sharehodl.com

echo "Wallet restart complete!"