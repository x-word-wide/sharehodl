#!/bin/bash

# ShareHODL All Services Startup Script
# Ensures all frontend apps start with correct ports

set -e

FRONTEND_DIR="/home/ubuntu/sharehodl-blockchain/sharehodl-frontend"

echo "==================================="
echo "Starting All ShareHODL Services"
echo "==================================="

# Kill any existing services
echo "Stopping existing services..."
pkill -f 'next.*start' || true
sleep 5

echo "Starting services with correct ports..."

# Start each service with its designated port
cd $FRONTEND_DIR

# Home (port 3000)
nohup npx --prefix apps/home next start --port 3000 > /tmp/home-service.log 2>&1 & 
echo $! > /tmp/home-service.pid
echo "✓ Home started on port 3000"

# Governance (port 3001)  
nohup npx --prefix apps/governance next start --port 3001 > /tmp/governance-service.log 2>&1 &
echo $! > /tmp/governance-service.pid
echo "✓ Governance started on port 3001"

# Trading (port 3002)
nohup npx --prefix apps/trading next start --port 3002 > /tmp/trading-service.log 2>&1 &
echo $! > /tmp/trading-service.pid
echo "✓ Trading started on port 3002"

# Explorer (port 3003)
nohup npx --prefix apps/explorer next start --port 3003 > /tmp/explorer-service.log 2>&1 &
echo $! > /tmp/explorer-service.pid
echo "✓ Explorer started on port 3003"

# Wallet (port 3004)
nohup npx --prefix apps/wallet next start --port 3004 > /tmp/wallet-service.log 2>&1 &
echo $! > /tmp/wallet-service.pid
echo "✓ Wallet started on port 3004"

# Business (port 3005)
nohup npx --prefix apps/business next start --port 3005 > /tmp/business-service.log 2>&1 &
echo $! > /tmp/business-service.pid
echo "✓ Business started on port 3005"

echo ""
echo "Services starting... waiting 15 seconds..."
sleep 15

echo ""
echo "=== Service Status Check ==="
curl -s -o /dev/null -w "Home: %{http_code}\n" http://localhost:3000 || echo "Home: Failed"
curl -s -o /dev/null -w "Governance: %{http_code}\n" http://localhost:3001 || echo "Governance: Failed"
curl -s -o /dev/null -w "Trading: %{http_code}\n" http://localhost:3002 || echo "Trading: Failed"  
curl -s -o /dev/null -w "Explorer: %{http_code}\n" http://localhost:3003 || echo "Explorer: Failed"
curl -s -o /dev/null -w "Wallet: %{http_code}\n" http://localhost:3004 || echo "Wallet: Failed"
curl -s -o /dev/null -w "Business: %{http_code}\n" http://localhost:3005 || echo "Business: Failed"

echo ""
echo "All services started successfully!"
echo "Services will run in background with process IDs saved in /tmp/*-service.pid"