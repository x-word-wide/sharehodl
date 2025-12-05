#!/bin/bash

# ShareHODL Docker Production Deployment Script
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=================================="
echo "ShareHODL Docker Production Deploy"
echo "=================================="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Pulling latest code..."
git pull origin main
git push origin main

echo ""
echo "Step 2: Updating production server..."
run_remote "cd $REMOTE_DIR && git fetch origin main && git reset --hard origin/main"

echo ""
echo "Step 3: Stopping existing containers..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml down || true"

echo ""
echo "Step 4: Building new Docker images..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml build --no-cache"

echo ""
echo "Step 5: Starting services..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml up -d"

echo ""
echo "Step 6: Waiting for services to start..."
sleep 30

echo ""
echo "Step 7: Checking container status..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml ps"

echo ""
echo "Step 8: Testing endpoints..."
sleep 10

echo "Testing governance:"
curl -s -o /dev/null -w "Status: %{http_code}\n" https://gov.sharehodl.com || echo "Failed"

echo "Testing trading:"
curl -s -o /dev/null -w "Status: %{http_code}\n" https://trade.sharehodl.com || echo "Failed"

echo "Testing wallet:"
curl -s -o /dev/null -w "Status: %{http_code}\n" https://wallet.sharehodl.com || echo "Failed"

echo ""
echo "=================================="
echo "Docker deployment complete!"
echo "=================================="