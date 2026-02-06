#!/bin/bash

# Upload deployment scripts to server
# Run this from your LOCAL machine

SERVER_IP="178.63.13.190"
SSH_PORT="${1:-22}"  # Use 22 initially, then 2222 after hardening
SSH_USER="${2:-root}"

echo "=============================================="
echo "  Upload Deployment Scripts to Server"
echo "  Target: $SSH_USER@$SERVER_IP:$SSH_PORT"
echo "=============================================="

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo "Uploading scripts..."

scp -P "$SSH_PORT" \
    "$SCRIPT_DIR/01-server-hardening.sh" \
    "$SCRIPT_DIR/02-deploy-sharehodl.sh" \
    "$SCRIPT_DIR/03-verify-deployment.sh" \
    "$SCRIPT_DIR/04-genesis-setup.sh" \
    "$SCRIPT_DIR/DEPLOYMENT_GUIDE.md" \
    "$SSH_USER@$SERVER_IP:/root/"

if [[ $? -eq 0 ]]; then
    echo ""
    echo "Scripts uploaded successfully!"
    echo ""
    echo "Next steps:"
    echo "1. SSH to server: ssh $SSH_USER@$SERVER_IP"
    echo "2. Run: chmod +x /root/*.sh"
    echo "3. Run: bash /root/01-server-hardening.sh sharehodl 2222"
    echo ""
else
    echo "Upload failed!"
    exit 1
fi
