#!/bin/bash

# Test Docker Setup Locally
set -e

echo "=== Testing ShareHODL Docker Setup Locally ==="

echo "Step 1: Building Docker image locally..."
docker compose -f docker-compose.local.yml build

echo ""
echo "Step 2: Starting local Docker services..."
docker compose -f docker-compose.local.yml up -d

echo ""
echo "Step 3: Waiting for services to start..."
sleep 60

echo ""
echo "Step 4: Testing local services..."
echo "Testing local services on localhost:"

curl -s -o /dev/null -w "Home (3000): %{http_code} | " http://localhost:3000 || echo "Home: Failed"
curl -s -o /dev/null -w "Governance (3001): %{http_code} | " http://localhost:3001 || echo "Governance: Failed"
curl -s -o /dev/null -w "Trading (3002): %{http_code} | " http://localhost:3002 || echo "Trading: Failed"
curl -s -o /dev/null -w "Explorer (3003): %{http_code} | " http://localhost:3003 || echo "Explorer: Failed"
curl -s -o /dev/null -w "Wallet (3004): %{http_code} | " http://localhost:3004 || echo "Wallet: Failed"
curl -s -o /dev/null -w "Business (3005): %{http_code}\n" http://localhost:3005 || echo "Business: Failed"

echo ""
echo "Step 5: Checking container logs..."
docker compose -f docker-compose.local.yml logs --tail=10

echo ""
echo "Step 6: Checking container status..."
docker compose -f docker-compose.local.yml ps

echo ""
echo "Local Docker test complete!"
echo "If all services show 200 status, the Docker setup is working correctly."
echo ""
echo "To stop: docker compose -f docker-compose.local.yml down"