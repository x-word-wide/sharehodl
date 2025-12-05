#!/bin/bash

# ShareHODL Automated EC2 Deployment Script
# This script automatically deploys ShareHODL to the launched EC2 instance

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
    log_info "Run: ./launch-sharehodl-ec2.sh first"
    exit 1
fi

# Load configuration
source aws-deployment.env

# Check if we have all required variables
if [ -z "$SHAREHODL_PUBLIC_IP" ] || [ -z "$SHAREHODL_KEY_PATH" ]; then
    log_error "Missing instance details. Make sure EC2 instance is launched."
    exit 1
fi

log_info "Deploying ShareHODL to EC2 instance: $SHAREHODL_PUBLIC_IP"

# Wait for SSH to be ready
log_info "Waiting for SSH to be ready..."
sleep 60

# Test SSH connection
log_info "Testing SSH connection..."
ssh -i $SHAREHODL_KEY_PATH -o StrictHostKeyChecking=no ubuntu@$SHAREHODL_PUBLIC_IP "echo 'SSH connection successful'"

# Copy deployment script to instance
log_info "Copying deployment script to instance..."
scp -i $SHAREHODL_KEY_PATH -o StrictHostKeyChecking=no \
    scripts/deploy-aws-ec2.sh \
    ubuntu@$SHAREHODL_PUBLIC_IP:~/deploy-aws-ec2.sh

# Run deployment on instance
log_info "Running deployment script on instance..."
ssh -i $SHAREHODL_KEY_PATH -o StrictHostKeyChecking=no ubuntu@$SHAREHODL_PUBLIC_IP << 'EOF'
chmod +x deploy-aws-ec2.sh
sudo ./deploy-aws-ec2.sh
EOF

log_success "ShareHODL deployment completed!"
log_info ""
log_info "Access your ShareHODL testnet:"
log_info "  Frontend: http://$SHAREHODL_PUBLIC_IP"
log_info "  Governance: http://$SHAREHODL_PUBLIC_IP/governance"
log_info "  Trading: http://$SHAREHODL_PUBLIC_IP/trading"
log_info "  Explorer: http://$SHAREHODL_PUBLIC_IP/explorer"
log_info "  Business: http://$SHAREHODL_PUBLIC_IP/business"
log_info "  API: http://$SHAREHODL_PUBLIC_IP/api/node_info"
log_info "  Health: http://$SHAREHODL_PUBLIC_IP/health"
log_info ""
log_info "SSH Access:"
log_info "  ssh -i $SHAREHODL_KEY_PATH ubuntu@$SHAREHODL_PUBLIC_IP"