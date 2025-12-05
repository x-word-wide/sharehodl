#!/bin/bash

# Simple Docker Restart Script for ShareHODL
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Restarting All ShareHODL Docker Services ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Restarting Docker containers..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml restart"

echo "Waiting for services to restart..."
sleep 30

echo "Checking container status..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml ps"

echo ""
echo "Testing all endpoints..."
curl -s -o /dev/null -w "Governance: %{http_code} | " https://gov.sharehodl.com
curl -s -o /dev/null -w "Trading: %{http_code} | " https://trade.sharehodl.com  
curl -s -o /dev/null -w "Wallet: %{http_code} | " https://wallet.sharehodl.com
curl -s -o /dev/null -w "Explorer: %{http_code} | " https://scan.sharehodl.com
curl -s -o /dev/null -w "Business: %{http_code}\n" https://business.sharehodl.com

echo ""
echo "Docker restart complete!"