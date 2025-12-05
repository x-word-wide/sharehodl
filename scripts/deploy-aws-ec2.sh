#!/bin/bash

# ShareHODL AWS EC2 Deployment Script
# Optimized for AWS infrastructure and services

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

# Get instance metadata
INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
REGION=$(curl -s http://169.254.169.254/latest/meta-data/placement/region)
AZ=$(curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone)

log_info "Deploying ShareHODL on AWS EC2"
log_info "Instance ID: $INSTANCE_ID"
log_info "Public IP: $PUBLIC_IP"
log_info "Region: $REGION"
log_info "Availability Zone: $AZ"

# Update system packages
log_info "Updating Ubuntu packages..."
sudo apt update && sudo apt upgrade -y

# Install AWS CLI and CloudWatch agent
log_info "Installing AWS tools..."
sudo apt install -y awscli amazon-cloudwatch-agent

# Install dependencies
log_info "Installing dependencies..."
sudo apt install -y build-essential git curl jq nginx certbot python3-certbot-nginx ufw htop

# Optimize for EC2
log_info "Configuring EC2 optimizations..."
# Enable enhanced networking
sudo modprobe ena
echo 'ena' | sudo tee -a /etc/modules

# Install Go 1.21
log_info "Installing Go 1.21..."
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPROXY=https://proxy.golang.org,direct' >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin

# Install Node.js 20 LTS
log_info "Installing Node.js and pnpm..."
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs
npm install -g pnpm

# Clone ShareHODL repository
log_info "Cloning ShareHODL repository..."
cd /home/ubuntu
git clone https://github.com/sharehodl/sharehodl-blockchain.git
cd sharehodl-blockchain

# Build ShareHODL blockchain
log_info "Building ShareHODL blockchain..."
make install

# Create sharehodl user for security
sudo useradd -m -s /bin/bash sharehodl
sudo usermod -aG sudo sharehodl

# Switch to sharehodl user for setup
sudo -u sharehodl bash << 'EOF'
cd /home/sharehodl

# Initialize testnet
sharehodld init sharehodl-aws-testnet --chain-id sharehodl-aws-testnet

# Create validator key
sharehodld keys add validator --keyring-backend test

# Get validator address  
VALIDATOR_ADDR=$(sharehodl keys show validator -a --keyring-backend test)
echo "Validator address: $VALIDATOR_ADDR"

# Create genesis account
sharehodl genesis add-genesis-account $VALIDATOR_ADDR 1000000000000stake,1000000000000hodl

# Generate validator genesis transaction
sharehodld genesis gentx validator 1000000000stake \
  --keyring-backend test \
  --chain-id sharehodl-aws-testnet \
  --commission-rate 0.1 \
  --commission-max-rate 0.2 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1

# Collect genesis transactions
sharehodld genesis collect-gentxs

# Configure for AWS networking
sed -i 's/enable = false/enable = true/' ~/.sharehodl/config/app.toml
sed -i 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/0.0.0.0:1317"/' ~/.sharehodl/config/app.toml
sed -i 's/address = "localhost:9090"/address = "0.0.0.0:9090"/' ~/.sharehodl/config/app.toml
sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' ~/.sharehodl/config/config.toml

# Enable CORS for frontend
sed -i 's/enabled-unsafe-cors = false/enabled-unsafe-cors = true/' ~/.sharehodl/config/app.toml
EOF

# Create systemd service for ShareHODL
log_info "Creating ShareHODL systemd service..."
sudo tee /etc/systemd/system/sharehodld.service > /dev/null <<EOF
[Unit]
Description=ShareHODL Blockchain Node
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=3
User=sharehodl
WorkingDirectory=/home/sharehodl
ExecStart=/usr/local/go/bin/sharehodld start --home /home/sharehodl/.sharehodl
StandardOutput=journal
StandardError=journal
SyslogIdentifier=sharehodld
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# Setup CloudWatch logging
log_info "Configuring CloudWatch logging..."
sudo tee /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json > /dev/null <<EOF
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/syslog",
            "log_group_name": "sharehodl-${INSTANCE_ID}",
            "log_stream_name": "syslog"
          }
        ]
      }
    }
  },
  "metrics": {
    "namespace": "ShareHODL/EC2",
    "metrics_collected": {
      "cpu": {
        "measurement": ["cpu_usage_idle", "cpu_usage_iowait", "cpu_usage_user", "cpu_usage_system"],
        "metrics_collection_interval": 60
      },
      "disk": {
        "measurement": ["used_percent"],
        "metrics_collection_interval": 60,
        "resources": ["*"]
      },
      "mem": {
        "measurement": ["mem_used_percent"],
        "metrics_collection_interval": 60
      }
    }
  }
}
EOF

# Start CloudWatch agent
sudo systemctl enable amazon-cloudwatch-agent
sudo systemctl start amazon-cloudwatch-agent

# Build and configure frontend
log_info "Setting up frontend applications..."
cd /home/ubuntu/sharehodl-blockchain/sharehodl-frontend

# Install dependencies
sudo -u ubuntu pnpm install

# Create frontend systemd services
for app in governance trading explorer features business; do
  port=$((3000 + $(echo $app | wc -c)))
  
  sudo tee /etc/systemd/system/sharehodl-$app.service > /dev/null <<EOF
[Unit]
Description=ShareHODL $app Frontend
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=3
User=ubuntu
WorkingDirectory=/home/ubuntu/sharehodl-blockchain/sharehodl-frontend
Environment=NODE_ENV=production
Environment=PORT=$port
ExecStart=/usr/bin/pnpm dev
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

  sudo systemctl enable sharehodl-$app
done

# Configure Nginx with AWS best practices
log_info "Configuring Nginx for AWS..."
sudo tee /etc/nginx/sites-available/sharehodl > /dev/null <<EOF
server {
    listen 80;
    server_name $PUBLIC_IP _;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";

    # Frontend applications
    location / {
        proxy_pass http://localhost:3001;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /governance {
        proxy_pass http://localhost:3001;
        include proxy_params;
    }

    location /trading {
        proxy_pass http://localhost:3002;
        include proxy_params;
    }

    location /explorer {
        proxy_pass http://localhost:3003;
        include proxy_params;
    }

    location /features {
        proxy_pass http://localhost:3004;
        include proxy_params;
    }

    location /business {
        proxy_pass http://localhost:3005;
        include proxy_params;
    }

    # API endpoints
    location /api/ {
        proxy_pass http://localhost:1317/;
        include proxy_params;
    }

    location /rpc/ {
        proxy_pass http://localhost:26657/;
        include proxy_params;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Health check for ALB
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}
EOF

# Create proxy_params file
sudo tee /etc/nginx/proxy_params > /dev/null <<EOF
proxy_set_header Host \$host;
proxy_set_header X-Real-IP \$remote_addr;
proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto \$scheme;
proxy_buffering off;
EOF

# Enable nginx site
sudo ln -sf /etc/nginx/sites-available/sharehodl /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Configure security groups via AWS CLI (if IAM permissions available)
log_info "Configuring security groups..."
if aws ec2 describe-instances --instance-ids $INSTANCE_ID &>/dev/null; then
    SECURITY_GROUP=$(aws ec2 describe-instances --instance-ids $INSTANCE_ID --query 'Reservations[0].Instances[0].SecurityGroups[0].GroupId' --output text)
    
    # Allow required ports
    aws ec2 authorize-security-group-ingress --group-id $SECURITY_GROUP --protocol tcp --port 80 --cidr 0.0.0.0/0 2>/dev/null || true
    aws ec2 authorize-security-group-ingress --group-id $SECURITY_GROUP --protocol tcp --port 443 --cidr 0.0.0.0/0 2>/dev/null || true
    aws ec2 authorize-security-group-ingress --group-id $SECURITY_GROUP --protocol tcp --port 26657 --cidr 0.0.0.0/0 2>/dev/null || true
    aws ec2 authorize-security-group-ingress --group-id $SECURITY_GROUP --protocol tcp --port 1317 --cidr 0.0.0.0/0 2>/dev/null || true
fi

# Configure UFW firewall
log_info "Configuring firewall..."
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 26657/tcp
sudo ufw allow 1317/tcp
sudo ufw allow 9090/tcp
sudo ufw --force enable

# Start all services
log_info "Starting all services..."
sudo systemctl daemon-reload

# Start blockchain
sudo systemctl enable sharehodld
sudo systemctl start sharehodld

# Wait for blockchain to start
sleep 10

# Start frontend services
sudo systemctl start sharehodl-governance
sudo systemctl start sharehodl-trading  
sudo systemctl start sharehodl-explorer
sudo systemctl start sharehodl-features
sudo systemctl start sharehodl-business

# Start nginx
sudo nginx -t && sudo systemctl enable nginx && sudo systemctl start nginx

# Create instance tags
aws ec2 create-tags --resources $INSTANCE_ID --tags Key=Name,Value=ShareHODL-Testnet Key=Environment,Value=Testing Key=Application,Value=ShareHODL 2>/dev/null || true

log_success "ShareHODL AWS deployment completed!"
log_info "Instance Details:"
log_info "  Instance ID: $INSTANCE_ID"
log_info "  Public IP: $PUBLIC_IP"
log_info "  Region: $REGION"
log_info ""
log_info "Access your ShareHODL testnet:"
log_info "  Frontend: http://$PUBLIC_IP"
log_info "  API: http://$PUBLIC_IP/api/node_info"
log_info "  RPC: http://$PUBLIC_IP/rpc/status" 
log_info "  Health: http://$PUBLIC_IP/health"
log_info ""
log_info "Service Management:"
log_info "  sudo systemctl status sharehodld"
log_info "  sudo journalctl -u sharehodld -f"
log_info "  aws logs describe-log-groups --log-group-name-prefix sharehodl"
log_info ""
log_info "Test Account:"
log_info "  sudo -u sharehodl sharehodld keys show validator -a --keyring-backend test"

# Display validator info
log_info "Getting validator address for testing..."
VALIDATOR_ADDR=$(sudo -u sharehodl sharehodl keys show validator -a --keyring-backend test)
log_success "Test account address: $VALIDATOR_ADDR"
log_info "This account has 1T HODL and 1T STAKE tokens for testing"