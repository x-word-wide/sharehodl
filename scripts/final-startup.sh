#!/bin/bash

# Final Service Startup Script
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Final ShareHODL Service Startup ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Ensuring clean slate..."
run_remote "pkill -f next || true && pkill -f pnpm || true"
sleep 5

echo "Step 2: Starting services with correct port configurations..."
echo "Starting Home on port 3000..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter home start > /tmp/home.log 2>&1 & echo \$! > /tmp/home.pid"

echo "Starting Governance on port 3001..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter governance start > /tmp/governance.log 2>&1 & echo \$! > /tmp/governance.pid"

echo "Starting Trading on port 3002..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter trading start > /tmp/trading.log 2>&1 & echo \$! > /tmp/trading.pid"

echo "Starting Explorer on port 3003..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter explorer start > /tmp/explorer.log 2>&1 & echo \$! > /tmp/explorer.pid"

echo "Starting Wallet on port 3004..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter wallet start > /tmp/wallet.log 2>&1 & echo \$! > /tmp/wallet.pid"

echo "Starting Business on port 3005..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter business start > /tmp/business.log 2>&1 & echo \$! > /tmp/business.pid"

echo ""
echo "Step 3: Waiting for services to initialize..."
sleep 45

echo ""
echo "Step 4: Testing all local ports..."
run_remote "curl -s -o /dev/null -w 'Home: %{http_code} | ' http://localhost:3000 && curl -s -o /dev/null -w 'Gov: %{http_code} | ' http://localhost:3001 && curl -s -o /dev/null -w 'Trade: %{http_code} | ' http://localhost:3002 && curl -s -o /dev/null -w 'Explorer: %{http_code} | ' http://localhost:3003 && curl -s -o /dev/null -w 'Wallet: %{http_code} | ' http://localhost:3004 && curl -s -o /dev/null -w 'Business: %{http_code}' http://localhost:3005"

echo ""
echo ""
echo "Step 5: Testing all HTTPS endpoints..."
echo "Governance: $(curl -s -o /dev/null -w '%{http_code}' https://gov.sharehodl.com)"
echo "Trading: $(curl -s -o /dev/null -w '%{http_code}' https://trade.sharehodl.com)"
echo "Wallet: $(curl -s -o /dev/null -w '%{http_code}' https://wallet.sharehodl.com)"
echo "Explorer: $(curl -s -o /dev/null -w '%{http_code}' https://scan.sharehodl.com)"
echo "Business: $(curl -s -o /dev/null -w '%{http_code}' https://business.sharehodl.com)"

echo ""
echo "=== Final startup complete! ==="
echo "All services should now be running with proper port configurations."
echo ""
echo "To restart all services in the future:"
echo "bash /home/ubuntu/sharehodl-blockchain/scripts/docker-restart-all.sh"