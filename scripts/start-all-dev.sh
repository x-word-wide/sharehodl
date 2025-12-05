#!/bin/bash

# Start All Services in Development Mode  
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Starting All Services in Development Mode ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Starting all apps in development mode..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter home dev > /tmp/home-dev.log 2>&1 & echo \$! > /tmp/home-dev.pid"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter governance dev > /tmp/governance-dev.log 2>&1 & echo \$! > /tmp/governance-dev.pid" 
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter trading dev > /tmp/trading-dev.log 2>&1 & echo \$! > /tmp/trading-dev.pid"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter explorer dev > /tmp/explorer-dev.log 2>&1 & echo \$! > /tmp/explorer-dev.pid"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter wallet dev > /tmp/wallet-dev.log 2>&1 & echo \$! > /tmp/wallet-dev.pid"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter business dev > /tmp/business-dev.log 2>&1 & echo \$! > /tmp/business-dev.pid"

echo ""
echo "Waiting 45 seconds for all services to start..."
sleep 45

echo ""
echo "Testing all services..."
echo "Governance: $(curl -s -o /dev/null -w '%{http_code}' https://gov.sharehodl.com)"
echo "Trading: $(curl -s -o /dev/null -w '%{http_code}' https://trade.sharehodl.com)"  
echo "Wallet: $(curl -s -o /dev/null -w '%{http_code}' https://wallet.sharehodl.com)"
echo "Explorer: $(curl -s -o /dev/null -w '%{http_code}' https://scan.sharehodl.com)"
echo "Business: $(curl -s -o /dev/null -w '%{http_code}' https://business.sharehodl.com)"

echo ""
echo "All services started in development mode!"
echo "This works around the production build startup issue."