#!/bin/bash
# ShareHODL Telegram Webapp Deploy Script
# Usage: ./deploy.sh

set -e

SERVER="root@178.63.13.190"
REMOTE_PATH="/var/www/sharehodl-telegram"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEBAPP_DIR="$SCRIPT_DIR/webapp"

echo "=== ShareHODL Telegram Webapp Deployment ==="
echo ""

# Check if we're in the right directory
if [ ! -d "$WEBAPP_DIR" ]; then
    echo "Error: webapp directory not found"
    exit 1
fi

cd "$WEBAPP_DIR"

# Build
echo "Building webapp..."
npm run build

if [ ! -d "dist" ]; then
    echo "Error: Build failed - dist directory not found"
    exit 1
fi

echo "Build complete!"
echo ""

# Deploy
echo "Deploying to $SERVER..."
ssh $SERVER "rm -rf $REMOTE_PATH/*"
scp -r dist/* $SERVER:$REMOTE_PATH/

echo ""
echo "=== Deployment Complete ==="
echo "Webapp URL: https://t.me/ShareHODLBot"
echo ""

# Optional: Check blockchain status
echo "Blockchain status:"
ssh $SERVER "curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height'" 2>/dev/null || echo "Could not fetch"
