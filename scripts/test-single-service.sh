#!/bin/bash

# Test Single Service Startup
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Testing Single Service Startup ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Testing governance service startup in foreground..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && npx --prefix apps/governance next start --port 3001" &
REMOTE_PID=$!

echo "Waiting 30 seconds for service to start..."
sleep 30

echo "Testing local connection..."
run_remote "curl -s -o /dev/null -w 'Status: %{http_code}' http://localhost:3001" || echo "Failed"

echo ""
echo "Testing HTTPS connection..." 
curl -s -o /dev/null -w "Status: %{http_code}\n" https://gov.sharehodl.com

# Kill the remote process
kill $REMOTE_PID 2>/dev/null || true

echo "Single service test complete!"