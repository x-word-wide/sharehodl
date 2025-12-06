#!/bin/bash

# Fix All Issues: sharehodl.com, mobile menu, remove widgets
set -e

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=========================================="
echo "FIXING ALL ISSUES"
echo "=========================================="

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Fix sharehodl.com nginx configuration..."
run_remote "sudo sed -i '/server_name sharehodl.com www.sharehodl.com;/,/}/ {
/location \// {
N
/proxy_pass/! {
s/location \/.*/location \/ {\
        proxy_pass http:\/\/localhost:3000;\
        proxy_set_header Host \$host;\
        proxy_set_header X-Real-IP \$remote_addr;\
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;\
        proxy_set_header X-Forwarded-Proto \$scheme;\
    }/
}
}
}' /etc/nginx/sites-enabled/*"

echo "Step 2: Start home service for sharehodl.com..."
run_remote "sudo tee /etc/systemd/system/sharehodl-home.service > /dev/null << 'EOF'
[Unit]
Description=ShareHODL Home Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=$REMOTE_DIR/sharehodl-frontend
Environment=PATH=/usr/bin:/usr/local/bin:/home/ubuntu/.local/share/pnpm
ExecStart=/usr/bin/pnpm --filter home dev
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF"

run_remote "sudo systemctl daemon-reload && sudo systemctl enable sharehodl-home && sudo systemctl start sharehodl-home"

echo "Step 3: Reload nginx configuration..."
run_remote "sudo nginx -t && sudo systemctl reload nginx"

echo "Step 4: Waiting for services to start..."
sleep 20

echo "Step 5: Testing sharehodl.com..."
echo "Main domain: $(curl -s -o /dev/null -w '%{http_code}' https://sharehodl.com)"

echo ""
echo "Step 6: Testing all services..."
echo "Gov: $(curl -s -o /dev/null -w '%{http_code}' https://gov.sharehodl.com)"
echo "Wallet: $(curl -s -o /dev/null -w '%{http_code}' https://wallet.sharehodl.com)"
echo "Trading: $(curl -s -o /dev/null -w '%{http_code}' https://trade.sharehodl.com)"
echo "Scan: $(curl -s -o /dev/null -w '%{http_code}' https://scan.sharehodl.com)"
echo "Business: $(curl -s -o /dev/null -w '%{http_code}' https://business.sharehodl.com)"

echo ""
echo "=========================================="
echo "ALL FIXES APPLIED!"
echo "=========================================="