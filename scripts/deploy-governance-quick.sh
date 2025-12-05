#!/bin/bash

# Quick Governance App Deployment Script
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Quick Governance Deployment ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Updating code on production server..."
run_remote "cd $REMOTE_DIR && git fetch origin main && git reset --hard origin/main"

echo "Step 2: Installing dependencies..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && pnpm install"

echo "Step 3: Building governance app..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && pnpm build --filter=governance"

echo "Step 4: Stopping governance service..."
run_remote "pkill -f 'governance.*3001' || true"
sleep 3

echo "Step 5: Starting governance service on port 3001..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup npx --prefix apps/governance next start --port 3001 > /tmp/governance.log 2>&1 & echo \$! > /tmp/governance.pid"

echo "Step 6: Waiting for service to start..."
sleep 10

echo "Step 7: Testing governance service..."
run_remote "curl -s -o /dev/null -w 'Local: %{http_code}' http://localhost:3001 || echo 'Failed'"

echo ""
echo "Step 8: Testing HTTPS endpoint..."
curl -s -o /dev/null -w "HTTPS: %{http_code}\n" https://gov.sharehodl.com || echo "Failed"

echo ""
echo "Governance deployment complete!"