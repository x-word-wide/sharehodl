#!/bin/bash

# ShareHODL Frontend Production Update Script
# This script pulls the latest code from GitHub and deploys frontend to production

set -e  # Exit on any error

REMOTE_HOST="54.198.239.167"
REMOTE_USER="ubuntu"
SSH_KEY="$HOME/.ssh/sharehodl-key-1764931859.pem"
REMOTE_DIR="/home/ubuntu/sharehodl-blockchain"

echo "=================================="
echo "ShareHODL Frontend Production Update"
echo "=================================="
echo ""

# Function to run commands on remote server
run_remote() {
    ssh -i "$SSH_KEY" "$REMOTE_USER@$REMOTE_HOST" "$1"
}

echo "Step 1: Pulling latest code from GitHub locally..."
git pull origin main
echo "✓ Local code updated"

echo ""
echo "Step 2: Copying updated frontend to production server..."
rsync -avz --exclude='node_modules' --exclude='.next' --exclude='dist' --exclude='.turbo' \
    -e "ssh -i $SSH_KEY" \
    ./sharehodl-frontend/ "$REMOTE_USER@$REMOTE_HOST:$REMOTE_DIR/sharehodl-frontend/"
echo "✓ Frontend code copied to production"

echo ""
echo "Step 3: Updating production git repository..."
run_remote "cd $REMOTE_DIR && git pull origin main"
echo "✓ Production git updated"

echo ""
echo "Step 4: Installing dependencies..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && pnpm install"
echo "✓ Dependencies installed"

echo ""
echo "Step 5: Building applications..."
run_remote "cd $REMOTE_DIR/sharehodl-frontend && pnpm build"
echo "✓ Applications built"

echo ""
echo "Step 6: Stopping existing services..."

# Kill existing processes gracefully
run_remote "pkill -f 'next.*governance' || true"
run_remote "pkill -f 'next.*trading' || true"  
run_remote "pkill -f 'next.*explorer' || true"
run_remote "pkill -f 'next.*wallet' || true"
run_remote "pkill -f 'next.*business' || true"

sleep 3
echo "✓ Services stopped"

echo ""
echo "Step 7: Starting services..."

# Start services in background
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter governance start > /tmp/governance.log 2>&1 & echo \$! > /tmp/governance.pid"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter trading start > /tmp/trading.log 2>&1 & echo \$! > /tmp/trading.pid"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter explorer start > /tmp/explorer.log 2>&1 & echo \$! > /tmp/explorer.pid"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter wallet start > /tmp/wallet.log 2>&1 & echo \$! > /tmp/wallet.pid"
run_remote "cd $REMOTE_DIR/sharehodl-frontend && nohup pnpm --filter business start > /tmp/business.log 2>&1 & echo \$! > /tmp/business.pid"

echo "✓ Services started"

echo ""
echo "Step 8: Waiting for services to initialize..."
sleep 15

echo ""
echo "Step 9: Checking service status..."
echo "Governance (port 3001):"
run_remote "curl -s -o /dev/null -w '%{http_code}' http://localhost:3001 || echo 'Failed'"

echo "Trading (port 3002):"
run_remote "curl -s -o /dev/null -w '%{http_code}' http://localhost:3002 || echo 'Failed'"

echo "Explorer (port 3003):"
run_remote "curl -s -o /dev/null -w '%{http_code}' http://localhost:3003 || echo 'Failed'"

echo "Wallet (port 3004):"
run_remote "curl -s -o /dev/null -w '%{http_code}' http://localhost:3004 || echo 'Failed'"

echo "Business (port 3005):"
run_remote "curl -s -o /dev/null -w '%{http_code}' http://localhost:3005 || echo 'Failed'"

echo ""
echo "Step 10: Testing HTTPS endpoints..."
echo "Testing https://gov.sharehodl.com..."
curl -s -o /dev/null -w "Status: %{http_code}\n" https://gov.sharehodl.com || echo "Failed"

echo "Testing https://trade.sharehodl.com..."
curl -s -o /dev/null -w "Status: %{http_code}\n" https://trade.sharehodl.com || echo "Failed"

echo "Testing https://scan.sharehodl.com..."
curl -s -o /dev/null -w "Status: %{http_code}\n" https://scan.sharehodl.com || echo "Failed"

echo "Testing https://wallet.sharehodl.com..."
curl -s -o /dev/null -w "Status: %{http_code}\n" https://wallet.sharehodl.com || echo "Failed"

echo "Testing https://business.sharehodl.com..."
curl -s -o /dev/null -w "Status: %{http_code}\n" https://business.sharehodl.com || echo "Failed"

echo "Testing redirect https://sharehodl.com..."
curl -s -o /dev/null -w "Status: %{http_code}\n" https://sharehodl.com || echo "Failed"

echo ""
echo "=================================="
echo "Frontend Deployment Complete!"
echo "=================================="
echo "All ShareHODL frontend services have been updated and restarted."
echo "Visit https://sharehodl.com to verify the deployment."
echo ""

# Optional: Show recent logs
echo "Recent logs:"
echo "Governance:"
run_remote "tail -n 3 /tmp/governance.log || echo 'No logs yet'"
echo "Trading:"
run_remote "tail -n 3 /tmp/trading.log || echo 'No logs yet'"
echo "Explorer:" 
run_remote "tail -n 3 /tmp/explorer.log || echo 'No logs yet'"

echo ""
echo "Frontend deployment script completed successfully!"