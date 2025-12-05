#!/bin/bash

# ShareHODL Remote SSL Configuration Script
# This script configures SSL on the remote EC2 instance

set -e

log_info() {
    echo -e "\033[0;34m[INFO]\033[0m $1"
}

log_success() {
    echo -e "\033[0;32m[SUCCESS]\033[0m $1"
}

log_error() {
    echo -e "\033[0;31m[ERROR]\033[0m $1"
}

# Check if AWS deployment config exists
if [ ! -f "aws-deployment.env" ]; then
    log_error "AWS deployment config not found!"
    log_info "Make sure EC2 instance is launched first"
    exit 1
fi

# Load configuration
source aws-deployment.env

log_info "Configuring SSL for ShareHODL domains on EC2: $SHAREHODL_PUBLIC_IP"

# Copy SSL setup script to instance
log_info "Copying SSL configuration script..."
scp -i $SHAREHODL_KEY_PATH -o StrictHostKeyChecking=no \
    scripts/setup-ssl-domains.sh \
    ubuntu@$SHAREHODL_PUBLIC_IP:~/setup-ssl-domains.sh

# Run SSL configuration
log_info "Setting up SSL certificates and domain configuration..."
ssh -i $SHAREHODL_KEY_PATH -o StrictHostKeyChecking=no ubuntu@$SHAREHODL_PUBLIC_IP << 'EOF'
chmod +x setup-ssl-domains.sh

# Wait for main deployment to complete first
echo "Waiting for main deployment to complete..."
while ! sudo systemctl is-active --quiet nginx; do
    echo "Waiting for Nginx to be active..."
    sleep 10
done

echo "Main deployment complete. Starting SSL configuration..."
sudo ./setup-ssl-domains.sh
EOF

log_success "SSL configuration completed!"
log_info ""
log_info "ShareHODL is now accessible with SSL at:"
log_info "  Main Site: https://sharehodl.com"
log_info "  Explorer: https://scan.sharehodl.com"
log_info "  Trading: https://trade.sharehodl.com"
log_info "  Governance: https://gov.sharehodl.com"  
log_info "  Business: https://business.sharehodl.com"
log_info "  Wallet: https://wallet.sharehodl.com"
log_info "  API: https://api.sharehodl.com"
log_info "  RPC: https://rpc.sharehodl.com"
log_info ""
log_info "SSL Features Enabled:"
log_info "  ✓ HTTPS enforced for all domains"
log_info "  ✓ Auto-renewal cron job configured"
log_info "  ✓ Security headers enabled"
log_info "  ✓ HTTP/2 support enabled"