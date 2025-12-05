#!/bin/bash

# ShareHODL EC2 Launch Script
# This script launches EC2 instance and deploys ShareHODL

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
    log_info "Run: ./aws-config.local.sh first"
    exit 1
fi

# Load configuration
source aws-deployment.env

log_info "Launching ShareHODL EC2 instance..."

# Launch EC2 instance
INSTANCE_ID=$(aws ec2 run-instances \
    --image-id ami-0c7217cdde317cfec \
    --count 1 \
    --instance-type t3.xlarge \
    --key-name $SHAREHODL_KEY_NAME \
    --security-group-ids $SHAREHODL_SECURITY_GROUP \
    --block-device-mappings '[{"DeviceName":"/dev/sda1","Ebs":{"VolumeSize":500,"VolumeType":"gp3","Iops":3000}}]' \
    --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=ShareHODL-Testnet},{Key=Environment,Value=Testing},{Key=Application,Value=ShareHODL}]' \
    --query 'Instances[0].InstanceId' \
    --output text)

echo "Instance launched: $INSTANCE_ID"
echo "SHAREHODL_INSTANCE_ID=$INSTANCE_ID" >> aws-deployment.env

# Wait for instance to be running
log_info "Waiting for instance to be running..."
aws ec2 wait instance-running --instance-ids $INSTANCE_ID

# Get public IP
PUBLIC_IP=$(aws ec2 describe-instances \
    --instance-ids $INSTANCE_ID \
    --query 'Reservations[0].Instances[0].PublicIpAddress' \
    --output text)

echo "Instance is running!"
echo "Public IP: $PUBLIC_IP"
echo "SHAREHODL_PUBLIC_IP=$PUBLIC_IP" >> aws-deployment.env

log_success "EC2 instance launched successfully!"
log_info ""
log_info "Instance Details:"
log_info "  Instance ID: $INSTANCE_ID"
log_info "  Public IP: $PUBLIC_IP"
log_info "  Key File: $SHAREHODL_KEY_PATH"
log_info ""
log_info "Next Steps:"
log_info "1. Wait 2-3 minutes for instance to fully boot"
log_info "2. Connect via SSH:"
log_info "   ssh -i $SHAREHODL_KEY_PATH ubuntu@$PUBLIC_IP"
log_info ""
log_info "3. Deploy ShareHODL (run this on the instance):"
log_info "   wget https://raw.githubusercontent.com/sharehodl/sharehodl-blockchain/main/scripts/deploy-aws-ec2.sh"
log_info "   chmod +x deploy-aws-ec2.sh"
log_info "   sudo ./deploy-aws-ec2.sh"
log_info ""
log_info "4. Or deploy automatically:"
log_info "   ./deploy-to-ec2.sh"