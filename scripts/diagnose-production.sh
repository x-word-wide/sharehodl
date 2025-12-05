#!/bin/bash

# Diagnose Production Issues
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=== Diagnosing Production Issues ==="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "1. Checking Node.js and pnpm versions..."
run_remote "node --version && pnpm --version"

echo ""
echo "2. Checking if code is up to date..."
run_remote "cd $REMOTE_DIR && git log --oneline -1"

echo ""
echo "3. Checking if pnpm install worked..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && ls -la node_modules/.pnpm || echo 'No pnpm modules'"

echo ""
echo "4. Checking if builds exist..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && ls -la apps/*/\.next || echo 'No builds found'"

echo ""
echo "5. Testing a simple build of one app..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && pnpm build --filter=governance 2>&1 | tail -10"

echo ""
echo "6. Testing if governance can start standalone..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && timeout 10 npx --prefix apps/governance next start --port 3001 2>&1 || echo 'Start command failed'"

echo ""
echo "Production diagnosis complete!"