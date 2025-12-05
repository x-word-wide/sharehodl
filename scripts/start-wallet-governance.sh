#!/bin/bash

# Start Wallet and Governance Services
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Starting Wallet and Governance Services ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Cleaning up existing processes..."
run_remote "pkill -f 'wallet.*3004' || true"
run_remote "pkill -f 'governance.*3001' || true"
sleep 3

echo "Starting governance service on port 3001..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter governance dev > /tmp/governance-live.log 2>&1 & echo \$! > /tmp/governance-live.pid"

echo "Starting wallet service on port 3004..."  
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter wallet dev > /tmp/wallet-live.log 2>&1 & echo \$! > /tmp/wallet-live.pid"

echo "Waiting for services to start..."
sleep 20

echo "Testing services..."
run_remote "curl -s -o /dev/null -w 'Governance local: %{http_code} | ' http://localhost:3001 && curl -s -o /dev/null -w 'Wallet local: %{http_code}' http://localhost:3004"

echo ""
echo "Testing HTTPS endpoints..."
curl -s -o /dev/null -w "Governance: %{http_code} | " https://gov.sharehodl.com
curl -s -o /dev/null -w "Wallet: %{http_code}\n" https://wallet.sharehodl.com

echo ""
echo "Services started! Process IDs saved in /tmp/*-live.pid"