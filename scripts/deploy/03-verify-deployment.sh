#!/bin/bash

# ShareHODL Deployment Verification Script
# Checks all services and endpoints after deployment

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SERVER_IP="${1:-178.63.13.190}"

echo "=============================================="
echo "  ShareHODL Deployment Verification"
echo "  Server: $SERVER_IP"
echo "=============================================="
echo ""

PASS=0
FAIL=0

check() {
    local name="$1"
    local cmd="$2"

    printf "%-40s" "Checking $name..."
    if eval "$cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}[PASS]${NC}"
        ((PASS++))
    else
        echo -e "${RED}[FAIL]${NC}"
        ((FAIL++))
    fi
}

check_response() {
    local name="$1"
    local url="$2"
    local expected="$3"

    printf "%-40s" "Checking $name..."
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    if [[ "$response" == "$expected" ]]; then
        echo -e "${GREEN}[PASS]${NC} (HTTP $response)"
        ((PASS++))
    else
        echo -e "${RED}[FAIL]${NC} (HTTP $response, expected $expected)"
        ((FAIL++))
    fi
}

echo "=== System Services ==="
check "Docker service" "systemctl is-active docker"
check "ShareHODL blockchain" "systemctl is-active sharehodld"
check "ShareHODL frontend" "systemctl is-active sharehodl-frontend"
check "Nginx" "systemctl is-active nginx"
check "Fail2ban" "systemctl is-active fail2ban"
check "UFW firewall" "sudo ufw status | grep -q 'Status: active'"

echo ""
echo "=== Blockchain Endpoints ==="
check_response "RPC Status" "http://localhost:26657/status" "200"
check_response "RPC Health" "http://localhost:26657/health" "200"
check_response "REST API Node Info" "http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info" "200"
check_response "gRPC (via curl)" "http://localhost:9090" "200"

echo ""
echo "=== External Access ==="
check_response "External RPC" "http://$SERVER_IP:26657/status" "200"
check_response "External API" "http://$SERVER_IP:1317/cosmos/base/tendermint/v1beta1/node_info" "200"
check_response "Nginx (HTTP)" "http://$SERVER_IP/" "200"

echo ""
echo "=== Frontend Apps ==="
check_response "Home (3001)" "http://localhost:3001" "200"
check_response "Explorer (3002)" "http://localhost:3002" "200"
check_response "Trading (3003)" "http://localhost:3003" "200"
check_response "Wallet (3004)" "http://localhost:3004" "200"
check_response "Governance (3005)" "http://localhost:3005" "200"
check_response "Business (3006)" "http://localhost:3006" "200"

echo ""
echo "=== Blockchain Status ==="
if curl -s http://localhost:26657/status > /tmp/blockchain_status.json 2>/dev/null; then
    echo -e "${BLUE}Chain ID:${NC} $(jq -r '.result.node_info.network' /tmp/blockchain_status.json 2>/dev/null || echo 'N/A')"
    echo -e "${BLUE}Node:${NC} $(jq -r '.result.node_info.moniker' /tmp/blockchain_status.json 2>/dev/null || echo 'N/A')"
    echo -e "${BLUE}Latest Block:${NC} $(jq -r '.result.sync_info.latest_block_height' /tmp/blockchain_status.json 2>/dev/null || echo 'N/A')"
    echo -e "${BLUE}Catching Up:${NC} $(jq -r '.result.sync_info.catching_up' /tmp/blockchain_status.json 2>/dev/null || echo 'N/A')"
    rm -f /tmp/blockchain_status.json
else
    echo -e "${RED}Could not fetch blockchain status${NC}"
fi

echo ""
echo "=== Disk Usage ==="
df -h / | tail -1 | awk '{print "Root: " $3 " used of " $2 " (" $5 " full)"}'
df -h /opt 2>/dev/null | tail -1 | awk '{print "/opt: " $3 " used of " $2 " (" $5 " full)"}' || true

echo ""
echo "=== Memory Usage ==="
free -h | grep Mem | awk '{print "Memory: " $3 " used of " $2}'

echo ""
echo "=== Open Ports ==="
echo "Listening ports:"
sudo ss -tlnp | grep -E ':(22|80|443|1317|9090|26656|26657|3001|3002|3003|3004|3005|3006)' | awk '{print "  " $4}'

echo ""
echo "=============================================="
echo "  VERIFICATION SUMMARY"
echo "=============================================="
echo -e "  ${GREEN}Passed:${NC} $PASS"
echo -e "  ${RED}Failed:${NC} $FAIL"
echo ""

if [[ $FAIL -eq 0 ]]; then
    echo -e "${GREEN}All checks passed! Deployment is healthy.${NC}"
    exit 0
else
    echo -e "${YELLOW}Some checks failed. Review the output above.${NC}"
    exit 1
fi
