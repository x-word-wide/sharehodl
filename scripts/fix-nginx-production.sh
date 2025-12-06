#!/bin/bash

# Fix Nginx and Production Services
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=========================================="
echo "FIXING NGINX AND PRODUCTION SERVICES"
echo "=========================================="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Check current nginx configuration..."
run_remote "sudo nginx -t && echo 'Nginx config is valid'"

echo ""
echo "Step 2: Check what ports nginx expects..."
run_remote "sudo cat /etc/nginx/sites-enabled/* | grep -E 'proxy_pass|listen' || echo 'No nginx config found'"

echo ""
echo "Step 3: Kill all existing processes..."
run_remote "sudo pkill -f next || true && sudo pkill -f pnpm || true && sudo pkill -f node || true"
sleep 5

echo ""
echo "Step 4: Check if ports are free..."
run_remote "sudo ss -tlnp | grep -E ':(3000|3001|3002|3003|3004|3005)' || echo 'All ports are free'"

echo ""
echo "Step 5: Start services with explicit ports..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter governance dev --port 3001 > /tmp/gov-fixed.log 2>&1 & echo 'Started governance'"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter wallet dev --port 3004 > /tmp/wallet-fixed.log 2>&1 & echo 'Started wallet'"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter trading dev --port 3002 > /tmp/trading-fixed.log 2>&1 & echo 'Started trading'"

echo ""
echo "Step 6: Wait for services to start..."
sleep 25

echo ""
echo "Step 7: Test local ports..."
run_remote "curl -s -o /dev/null -w 'Port 3001: %{http_code} | ' http://localhost:3001 && curl -s -o /dev/null -w 'Port 3004: %{http_code} | ' http://localhost:3004 && curl -s -o /dev/null -w 'Port 3002: %{http_code}' http://localhost:3002"

echo ""
echo ""
echo "Step 8: Test HTTPS endpoints..."
echo "Governance: $(curl -s -o /dev/null -w '%{http_code}' https://gov.sharehodl.com)"
echo "Wallet: $(curl -s -o /dev/null -w '%{http_code}' https://wallet.sharehodl.com)"
echo "Trading: $(curl -s -o /dev/null -w '%{http_code}' https://trade.sharehodl.com)"

echo ""
echo "Step 9: Check service logs..."
run_remote "tail -5 /tmp/gov-fixed.log 2>/dev/null || echo 'No gov logs'"
run_remote "tail -5 /tmp/wallet-fixed.log 2>/dev/null || echo 'No wallet logs'"

echo ""
echo "=========================================="
echo "NGINX PRODUCTION FIX COMPLETE"
echo "=========================================="