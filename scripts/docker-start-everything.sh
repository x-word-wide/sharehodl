#!/bin/bash

# Ultimate Docker Start Everything Script
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=========================================="
echo "DOCKER START EVERYTHING - SHAREHODL"
echo "=========================================="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Updating code on production..."
git add . && git commit -m "feat: ultimate Docker solution with service verification" || true
git push origin main
run_remote "cd $REMOTE_DIR && git fetch origin main && git reset --hard origin/main"

echo ""
echo "Step 2: Stopping any existing containers..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml down || true"
run_remote "docker container prune -f"

echo ""
echo "Step 3: Building fresh Docker image..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml build --no-cache"

echo ""
echo "Step 4: Starting Docker container..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml up -d"

echo ""
echo "Step 5: Waiting for services to start (this takes time)..."
sleep 60

echo ""
echo "Step 6: Checking Docker container logs..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml logs --tail=20 sharehodl-frontend"

echo ""
echo "Step 7: Testing all services..."
echo "Governance: $(curl -s -o /dev/null -w '%{http_code}' https://gov.sharehodl.com)"
echo "Trading: $(curl -s -o /dev/null -w '%{http_code}' https://trade.sharehodl.com)"
echo "Wallet: $(curl -s -o /dev/null -w '%{http_code}' https://wallet.sharehodl.com)"
echo "Explorer: $(curl -s -o /dev/null -w '%{http_code}' https://scan.sharehodl.com)"
echo "Business: $(curl -s -o /dev/null -w '%{http_code}' https://business.sharehodl.com)"

echo ""
echo "=========================================="
echo "DOCKER START EVERYTHING COMPLETE!"
echo "=========================================="
echo ""
echo "All services should now be running in Docker!"
echo "To restart: bash scripts/docker-start-everything.sh"
echo "To check logs: docker compose -f docker-compose.production.yml logs -f"
echo "To restart just container: docker compose -f docker-compose.production.yml restart"