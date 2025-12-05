#!/bin/bash

# Comprehensive Docker Deployment for ShareHODL
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=========================================="
echo "ShareHODL Comprehensive Docker Deployment"
echo "=========================================="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Updating local and remote repositories..."
git add . && git commit -m "feat: enhanced Docker deployment for all services" || true
git push origin main
run_remote "cd $REMOTE_DIR && git fetch origin main && git reset --hard origin/main"

echo ""
echo "Step 2: Ensuring Docker Compose is available..."
run_remote "docker --version || (echo 'Installing Docker...' && sudo apt update && sudo apt install -y docker.io)"
run_remote "sudo systemctl start docker && sudo systemctl enable docker"
run_remote "sudo usermod -aG docker ubuntu"

echo ""
echo "Step 3: Stopping all existing services..."
run_remote "pkill -f 'next' || true"
run_remote "pkill -f 'pnpm' || true" 
run_remote "pkill -f 'node' || true"
sleep 5

echo ""
echo "Step 4: Stopping and removing any existing containers..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml down --remove-orphans || true"
run_remote "docker container prune -f"

echo ""
echo "Step 5: Building fresh Docker images..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml build --no-cache"

echo ""
echo "Step 6: Starting services with Docker..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml up -d"

echo ""
echo "Step 7: Waiting for services to initialize..."
sleep 60

echo ""
echo "Step 8: Checking container status..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml ps"

echo ""
echo "Step 9: Checking container logs..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml logs --tail=10 sharehodl-frontend"

echo ""
echo "Step 10: Testing all HTTPS endpoints..."
sleep 10

echo "Testing Governance: https://gov.sharehodl.com"
curl -s -o /dev/null -w "Status: %{http_code}\n" https://gov.sharehodl.com

echo "Testing Trading: https://trade.sharehodl.com"
curl -s -o /dev/null -w "Status: %{http_code}\n" https://trade.sharehodl.com

echo "Testing Wallet: https://wallet.sharehodl.com"
curl -s -o /dev/null -w "Status: %{http_code}\n" https://wallet.sharehodl.com

echo "Testing Explorer: https://scan.sharehodl.com"
curl -s -o /dev/null -w "Status: %{http_code}\n" https://scan.sharehodl.com

echo "Testing Business: https://business.sharehodl.com"
curl -s -o /dev/null -w "Status: %{http_code}\n" https://business.sharehodl.com

echo ""
echo "=========================================="
echo "Docker Deployment Complete!"
echo "=========================================="
echo "All ShareHODL services are now running in Docker containers."
echo ""
echo "To restart all services: docker compose -f docker-compose.production.yml restart"
echo "To check logs: docker compose -f docker-compose.production.yml logs -f"
echo "To check status: docker compose -f docker-compose.production.yml ps"