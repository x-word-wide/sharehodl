#!/bin/bash

# Fix Docker Compose installation
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"

echo "=== Fixing Docker Compose Installation ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Removing broken docker-compose..."
run_remote "sudo rm -f /usr/local/bin/docker-compose"

echo "Step 2: Installing Docker Compose from apt..."
run_remote "sudo apt update && sudo apt install -y docker-compose-plugin"

echo "Step 3: Testing Docker Compose..."
run_remote "docker compose version"

echo ""
echo "Docker Compose installation fixed!"