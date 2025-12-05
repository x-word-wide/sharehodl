#!/bin/bash

# AWS ShareHODL Deployment Setup (SECURE)
# DO NOT PUT CREDENTIALS IN THIS FILE

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

log_info "Setting up AWS deployment for ShareHODL"
log_error "SECURITY: Never commit AWS credentials to Git!"

# Check if AWS CLI is installed
if ! command -v aws &> /dev/null; then
    log_info "Installing AWS CLI..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        brew install awscli
    elif [[ "$OSTYPE" == "linux"* ]]; then
        # Linux
        curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
        unzip awscliv2.zip
        sudo ./aws/install
    fi
fi

# Configure AWS (user will input credentials manually)
log_info "Configuring AWS CLI..."
log_info "You will be prompted to enter your AWS credentials"
log_info "Access Key ID: AKIA36Y72LEQWO5WDIIL"
log_info "Secret Access Key: [you have this]"
log_info "Default region: us-east-1"
log_info "Default output format: json"

aws configure

# Verify configuration
log_info "Testing AWS configuration..."
if aws sts get-caller-identity; then
    log_success "AWS configuration successful!"
else
    log_error "AWS configuration failed. Please check your credentials."
    exit 1
fi

# Create key pair for EC2 instances
log_info "Creating EC2 key pair..."
KEY_NAME="sharehodl-key-$(date +%s)"

aws ec2 create-key-pair \
    --key-name $KEY_NAME \
    --query 'KeyMaterial' \
    --output text > ~/.ssh/${KEY_NAME}.pem

# Secure the key file
chmod 400 ~/.ssh/${KEY_NAME}.pem

log_success "Key pair created: ~/.ssh/${KEY_NAME}.pem"
log_info "Key name for EC2: $KEY_NAME"

# Save key name for later use (not the key itself)
echo "SHAREHODL_KEY_NAME=$KEY_NAME" > aws-config.env
echo "SHAREHODL_KEY_PATH=~/.ssh/${KEY_NAME}.pem" >> aws-config.env

log_success "Configuration saved to aws-config.env"
log_info "Key file location: ~/.ssh/${KEY_NAME}.pem"
log_error "BACKUP THIS KEY FILE SECURELY - YOU CANNOT RETRIEVE IT FROM AWS!"

# Create security group
log_info "Creating security group..."
SECURITY_GROUP_ID=$(aws ec2 create-security-group \
    --group-name sharehodl-sg-$(date +%s) \
    --description "ShareHODL Testnet Security Group" \
    --query 'GroupId' \
    --output text)

# Add security group rules
aws ec2 authorize-security-group-ingress \
    --group-id $SECURITY_GROUP_ID \
    --protocol tcp --port 22 --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
    --group-id $SECURITY_GROUP_ID \
    --protocol tcp --port 80 --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
    --group-id $SECURITY_GROUP_ID \
    --protocol tcp --port 443 --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
    --group-id $SECURITY_GROUP_ID \
    --protocol tcp --port 26657 --cidr 0.0.0.0/0

aws ec2 authorize-security-group-ingress \
    --group-id $SECURITY_GROUP_ID \
    --protocol tcp --port 1317 --cidr 0.0.0.0/0

echo "SHAREHODL_SECURITY_GROUP=$SECURITY_GROUP_ID" >> aws-config.env

log_success "Security group created: $SECURITY_GROUP_ID"

# Create launch script
log_info "Creating EC2 launch script..."
cat > launch-ec2.sh << EOF
#!/bin/bash

# Load configuration
source aws-config.env

# Launch EC2 instance
INSTANCE_ID=\$(aws ec2 run-instances \\
    --image-id ami-0c7217cdde317cfec \\
    --count 1 \\
    --instance-type t3.xlarge \\
    --key-name \$SHAREHODL_KEY_NAME \\
    --security-group-ids \$SHAREHODL_SECURITY_GROUP \\
    --block-device-mappings '[{"DeviceName":"/dev/sda1","Ebs":{"VolumeSize":500,"VolumeType":"gp3","Iops":3000}}]' \\
    --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=ShareHODL-Testnet},{Key=Environment,Value=Testing}]' \\
    --query 'Instances[0].InstanceId' \\
    --output text)

echo "Instance launched: \$INSTANCE_ID"
echo "SHAREHODL_INSTANCE_ID=\$INSTANCE_ID" >> aws-config.env

# Wait for instance to be running
echo "Waiting for instance to be running..."
aws ec2 wait instance-running --instance-ids \$INSTANCE_ID

# Get public IP
PUBLIC_IP=\$(aws ec2 describe-instances \\
    --instance-ids \$INSTANCE_ID \\
    --query 'Reservations[0].Instances[0].PublicIpAddress' \\
    --output text)

echo "Instance is running!"
echo "Public IP: \$PUBLIC_IP"
echo "SHAREHODL_PUBLIC_IP=\$PUBLIC_IP" >> aws-config.env

echo ""
echo "To connect:"
echo "ssh -i \$SHAREHODL_KEY_PATH ubuntu@\$PUBLIC_IP"
echo ""
echo "To deploy ShareHODL:"
echo "scp -i \$SHAREHODL_KEY_PATH deploy-aws-ec2.sh ubuntu@\$PUBLIC_IP:~/"
echo "ssh -i \$SHAREHODL_KEY_PATH ubuntu@\$PUBLIC_IP"
echo "chmod +x deploy-aws-ec2.sh && sudo ./deploy-aws-ec2.sh"

EOF

chmod +x launch-ec2.sh

log_success "Setup complete!"
log_info "Next steps:"
log_info "1. Run: ./launch-ec2.sh"
log_info "2. SSH to your instance"
log_info "3. Run the deployment script"
log_info ""
log_info "Files created:"
log_info "- aws-config.env (safe to commit)"
log_info "- launch-ec2.sh (safe to commit)"
log_info "- ~/.ssh/${KEY_NAME}.pem (DO NOT COMMIT - BACKUP SECURELY!)"