#!/bin/bash

# Permanent Fix for ShareHODL Services
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=========================================="
echo "PERMANENT FIX FOR SHAREHODL SERVICES"
echo "=========================================="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Installing screen if not present..."
run_remote "sudo apt update && sudo apt install -y screen"

echo "Step 2: Killing all existing sessions and processes..."
run_remote "screen -wipe || true"
run_remote "pkill -f next || true && pkill -f pnpm || true"
sleep 5

echo "Step 3: Starting services in persistent screen sessions..."

echo "Starting Governance in screen..."
run_remote "screen -dmS governance bash -c 'cd $REMOTE_DIR/sharehodl-frontend && pnpm --filter governance dev'"

echo "Starting Wallet in screen..."  
run_remote "screen -dmS wallet bash -c 'cd $REMOTE_DIR/sharehodl-frontend && pnpm --filter wallet dev'"

echo "Starting Trading in screen..."
run_remote "screen -dmS trading bash -c 'cd $REMOTE_DIR/sharehodl-frontend && pnpm --filter trading dev'"

echo "Starting Explorer in screen..."
run_remote "screen -dmS explorer bash -c 'cd $REMOTE_DIR/sharehodl-frontend && pnpm --filter explorer dev'"

echo "Starting Business in screen..."
run_remote "screen -dmS business bash -c 'cd $REMOTE_DIR/sharehodl-frontend && pnpm --filter business dev'"

echo ""
echo "Step 4: Waiting for services to initialize..."
sleep 30

echo ""
echo "Step 5: Checking screen sessions..."
run_remote "screen -list"

echo ""
echo "Step 6: Testing all services..."
echo "Governance: $(curl -s -o /dev/null -w '%{http_code}' https://gov.sharehodl.com)"
echo "Trading: $(curl -s -o /dev/null -w '%{http_code}' https://trade.sharehodl.com)"
echo "Wallet: $(curl -s -o /dev/null -w '%{http_code}' https://wallet.sharehodl.com)"
echo "Explorer: $(curl -s -o /dev/null -w '%{http_code}' https://scan.sharehodl.com)"
echo "Business: $(curl -s -o /dev/null -w '%{http_code}' https://business.sharehodl.com)"

echo ""
echo "=========================================="
echo "PERMANENT FIX COMPLETE!"
echo "=========================================="
echo ""
echo "Services are now running in persistent screen sessions."
echo "They will survive SSH disconnections and stay running."
echo ""
echo "To check status: ssh ubuntu@$REMOTE_HOST 'screen -list'"
echo "To attach to a service: ssh ubuntu@$REMOTE_HOST 'screen -r governance'"
echo "To restart all: bash scripts/permanent-fix.sh"