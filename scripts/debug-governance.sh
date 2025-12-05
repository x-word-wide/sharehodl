#!/bin/bash

# Debug Governance Service
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu" 
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"

echo "=== Debugging Governance Service ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Checking what's running on port 3001..."
run_remote "sudo lsof -i :3001 || echo 'Nothing on port 3001'"

echo ""
echo "Checking governance process..."
run_remote "ps aux | grep governance | grep -v grep || echo 'No governance process'"

echo ""
echo "Checking recent governance logs..."
run_remote "tail -n 20 /tmp/governance.log || echo 'No logs'"

echo ""
echo "Testing direct port access..."
run_remote "curl -s -o /dev/null -w 'Port 3001: %{http_code}' http://localhost:3001 || echo 'Port 3001: Failed'"

echo ""
echo "Checking nginx configuration..."
run_remote "sudo nginx -t && sudo systemctl status nginx | grep Active"

echo ""
echo "Checking if there are any Next.js processes..."
run_remote "ps aux | grep next | grep -v grep || echo 'No Next.js processes'"