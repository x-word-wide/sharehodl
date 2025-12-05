#!/bin/bash

# Clean Restart All Services
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu" 
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Clean Restart of All Services ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Killing ALL Next.js processes..."
run_remote "pkill -f 'next' || true"
run_remote "pkill -f 'pnpm' || true"
run_remote "pkill -f 'node' || true"
sleep 5

echo "Step 2: Checking what's still using ports..."
run_remote "sudo lsof -i :3000 || echo 'Port 3000 free'"
run_remote "sudo lsof -i :3001 || echo 'Port 3001 free'"  
run_remote "sudo lsof -i :3002 || echo 'Port 3002 free'"
run_remote "sudo lsof -i :3003 || echo 'Port 3003 free'"
run_remote "sudo lsof -i :3004 || echo 'Port 3004 free'"
run_remote "sudo lsof -i :3005 || echo 'Port 3005 free'"

echo ""
echo "Step 3: Starting governance service first on port 3001..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup npx --prefix apps/governance next start --port 3001 > /tmp/governance.log 2>&1 & echo \$! > /tmp/governance.pid"
sleep 10

echo "Step 4: Testing governance..."
run_remote "curl -s -o /dev/null -w 'Governance: %{http_code}' http://localhost:3001 || echo 'Failed'"

echo ""
echo "Step 5: Starting other services..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup npx --prefix apps/trading next start --port 3002 > /tmp/trading.log 2>&1 & echo \$! > /tmp/trading.pid"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup npx --prefix apps/wallet next start --port 3004 > /tmp/wallet.log 2>&1 & echo \$! > /tmp/wallet.pid"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup npx --prefix apps/explorer next start --port 3003 > /tmp/explorer.log 2>&1 & echo \$! > /tmp/explorer.pid"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup npx --prefix apps/business next start --port 3005 > /tmp/business.log 2>&1 & echo \$! > /tmp/business.pid"

echo ""
echo "Step 6: Waiting for all services to start..."
sleep 15

echo ""
echo "Step 7: Testing all services..."
run_remote "curl -s -o /dev/null -w 'Governance: %{http_code} | ' http://localhost:3001 && curl -s -o /dev/null -w 'Trading: %{http_code} | ' http://localhost:3002 && curl -s -o /dev/null -w 'Wallet: %{http_code} | ' http://localhost:3004 && curl -s -o /dev/null -w 'Explorer: %{http_code} | ' http://localhost:3003 && curl -s -o /dev/null -w 'Business: %{http_code}' http://localhost:3005"

echo ""
echo "Clean restart complete!"