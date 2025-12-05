#!/bin/bash

# Check Docker Status Script
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Checking Docker Status ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Checking Docker containers..."
run_remote "docker ps -a"

echo ""
echo "Checking Docker Compose status..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml ps || echo 'No compose services'"

echo ""
echo "Checking what's running on ports..."
run_remote "sudo netstat -tlnp | grep -E ':(3000|3001|3002|3003|3004|3005)' || echo 'No services on target ports'"

echo ""
echo "Checking Docker logs if containers exist..."
run_remote "cd $REMOTE_DIR && docker compose -f docker-compose.production.yml logs --tail=20 || echo 'No container logs'"