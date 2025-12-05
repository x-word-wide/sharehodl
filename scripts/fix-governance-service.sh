#!/bin/bash

# Fix Governance Service Script
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Fixing Governance Service ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Ensuring latest code..."
run_remote "cd $REMOTE_DIR && git fetch origin main && git reset --hard origin/main"

echo "Step 2: Installing dependencies..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && pnpm install"

echo "Step 3: Building governance app..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && pnpm build --filter=governance"

echo "Step 4: Killing any existing governance processes..."
run_remote "pkill -f 'governance' || true"
run_remote "pkill -f '3001' || true"
sleep 5

echo "Step 5: Starting governance service directly with explicit port..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter governance start --port 3001 > /tmp/governance.log 2>&1 & echo \$! > /tmp/governance.pid"

echo "Step 6: Waiting for service to start..."
sleep 15

echo "Step 7: Testing local service..."
run_remote "curl -s -o /dev/null -w 'Local governance: %{http_code}' http://localhost:3001 || echo 'Failed'"

echo ""
echo "Step 8: Checking service logs..."
run_remote "tail -n 10 /tmp/governance.log"

echo ""
echo "Step 9: Testing HTTPS endpoint..."
curl -s -o /dev/null -w "HTTPS governance: %{http_code}\n" https://gov.sharehodl.com

echo ""
echo "Governance service fix complete!"