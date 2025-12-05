#!/bin/bash

# Service Restart Script
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Restarting ShareHODL Services ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Running the start-all-services script on production..."
run_remote "cd $REMOTE_DIR && bash scripts/start-all-services.sh"

echo ""
echo "Step 2: Testing all services after restart..."
sleep 20

echo "Testing governance:"
curl -s -o /dev/null -w "Status: %{http_code}\n" https://gov.sharehodl.com

echo "Testing wallet:"  
curl -s -o /dev/null -w "Status: %{http_code}\n" https://wallet.sharehodl.com

echo "Testing trading:"
curl -s -o /dev/null -w "Status: %{http_code}\n" https://trade.sharehodl.com

echo ""
echo "Service restart complete!"