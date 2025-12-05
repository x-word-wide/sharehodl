#!/bin/bash

# Emergency Wallet Service Fix
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Emergency Wallet Service Fix ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Check what's running..."
run_remote "ps aux | grep -E '(next|pnpm|node)' | grep -v grep || echo 'No services running'"

echo ""
echo "Step 2: Kill everything and clean up..."
run_remote "pkill -f next || true && pkill -f pnpm || true && pkill -f node || true"
sleep 5

echo ""
echo "Step 3: Check if port 3004 is free..."
run_remote "sudo ss -tlnp | grep :3004 || echo 'Port 3004 is free'"

echo ""
echo "Step 4: Try direct wallet startup..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && timeout 15s npx --prefix apps/wallet next start --port 3004 2>&1 || echo 'Direct start failed'"

echo ""
echo "Step 5: Try with full path..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend/apps/wallet && timeout 15s npx next start --port 3004 2>&1 || echo 'Full path start failed'"

echo ""
echo "Step 6: Check wallet logs if any..."
run_remote "tail -10 /tmp/wallet.log 2>/dev/null || echo 'No wallet logs'"

echo ""
echo "Emergency fix attempt complete."