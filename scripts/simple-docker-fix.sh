#!/bin/bash

# Simple Docker Fix
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Simple Docker Fix ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Building Docker image..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && docker build -t sharehodl-frontend ."

echo ""
echo "Step 2: Stopping existing container..."
run_remote "docker stop sharehodl-frontend-container || true && docker rm sharehodl-frontend-container || true"

echo ""
echo "Step 3: Starting new container..."
run_remote "docker run -d --name sharehodl-frontend-container -p 3000:3000 -p 3001:3001 -p 3002:3002 -p 3003:3003 -p 3004:3004 -p 3005:3005 sharehodl-frontend"

echo ""
echo "Step 4: Waiting for services..."
sleep 60

echo ""
echo "Step 5: Testing services..."
echo "Governance: $(curl -s -o /dev/null -w '%{http_code}' https://gov.sharehodl.com)"
echo "Wallet: $(curl -s -o /dev/null -w '%{http_code}' https://wallet.sharehodl.com)"

echo ""
echo "Docker container started! Check logs with:"
echo "ssh ubuntu@$REMOTE_HOST 'docker logs sharehodl-frontend-container'"