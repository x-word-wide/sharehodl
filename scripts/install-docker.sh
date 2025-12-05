#!/bin/bash

# Install Docker and Docker Compose on Ubuntu
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"

echo "=== Installing Docker on Production Server ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Updating package list..."
run_remote "sudo apt update"

echo "Step 2: Installing Docker..."
run_remote "sudo apt install -y docker.io"

echo "Step 3: Starting Docker service..."
run_remote "sudo systemctl start docker && sudo systemctl enable docker"

echo "Step 4: Adding user to docker group..."
run_remote "sudo usermod -aG docker ubuntu"

echo "Step 5: Installing Docker Compose..."
run_remote "sudo curl -L 'https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)' -o /usr/local/bin/docker-compose"
run_remote "sudo chmod +x /usr/local/bin/docker-compose"

echo "Step 6: Testing Docker installation..."
run_remote "docker --version && docker-compose --version"

echo ""
echo "Docker installation complete! You may need to log out and back in for group changes to take effect."