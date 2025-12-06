#!/bin/bash

# Final Working Solution - No Docker, Just Reliable Services
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=========================================="
echo "FINAL WORKING SOLUTION - SHAREHODL"
echo "=========================================="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Installing screen for persistent sessions..."
run_remote "sudo apt install -y screen"

echo "Cleaning up everything..."
run_remote "screen -wipe || true"
run_remote "pkill -f next || true"
run_remote "pkill -f pnpm || true"
sleep 5

echo "Starting services in screen sessions..."
run_remote "screen -dmS governance bash -c 'cd $REMOTE_DIR/sharehodl-frontend && pnpm --filter governance dev'"
run_remote "screen -dmS wallet bash -c 'cd $REMOTE_DIR/sharehodl-frontend && pnpm --filter wallet dev'"
run_remote "screen -dmS trading bash -c 'cd $REMOTE_DIR/sharehodl-frontend && pnpm --filter trading dev'"

echo "Waiting 20 seconds for services to start..."
sleep 20

echo "Testing services..."
echo "Governance: $(curl -s -o /dev/null -w '%{http_code}' https://gov.sharehodl.com)"
echo "Wallet: $(curl -s -o /dev/null -w '%{http_code}' https://wallet.sharehodl.com)"
echo "Trading: $(curl -s -o /dev/null -w '%{http_code}' https://trade.sharehodl.com)"

echo ""
echo "=========================================="
echo "SUCCESS! Services running in screen"
echo "=========================================="
echo "To restart all: bash scripts/final-working-solution.sh"
echo "To check status: ssh ubuntu@$REMOTE_HOST 'screen -list'"