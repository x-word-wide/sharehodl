#!/bin/bash

# Populate Test Data Script
# This script populates the local network with test companies, orders, and proposals

set -e

CHAIN_ID="sharehodl-local-1"
HOME_DIR="$HOME/.sharehodl"
NODE_URL="http://localhost:26657"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Wait for network to be ready
wait_for_network() {
    log_info "Waiting for network to be ready..."
    
    max_attempts=30
    attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s $NODE_URL/status > /dev/null 2>&1; then
            log_success "Network is ready!"
            return 0
        fi
        
        log_info "Attempt $attempt/$max_attempts - Network not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "Network failed to start after $max_attempts attempts"
    exit 1
}

# Execute transaction with retry
execute_tx() {
    local cmd="$1"
    local description="$2"
    local max_retries=3
    local retry=1
    
    while [ $retry -le $max_retries ]; do
        log_info "$description (attempt $retry/$max_retries)"
        
        if eval "$cmd --yes --gas auto --gas-adjustment 1.3 --broadcast-mode block"; then
            log_success "$description completed"
            return 0
        else
            log_warning "$description failed, retrying..."
            sleep 2
            retry=$((retry + 1))
        fi
    done
    
    log_error "$description failed after $max_retries attempts"
    return 1
}

# Create test companies
create_companies() {
    log_info "Creating test companies..."
    
    # ShareHODL Corp
    execute_tx \
        "sharehodld tx equity create-company 'ShareHODL Corp' 'SHODL' 1000000 'Blockchain Technology' 100000000 --from sharehodl-corp --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Creating ShareHODL Corp"
    
    sleep 3
    
    # TechStart Inc
    execute_tx \
        "sharehodld tx equity create-company 'TechStart Inc' 'TECH' 500000 'Software' 50000000 --from techstart --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Creating TechStart Inc"
    
    sleep 3
    
    # GreenEnergy LLC
    execute_tx \
        "sharehodld tx equity create-company 'GreenEnergy LLC' 'GREEN' 2000000 'Renewable Energy' 200000000 --from greenenergy --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Creating GreenEnergy LLC"
    
    sleep 3
    
    # HealthTech Solutions
    execute_tx \
        "sharehodld tx equity create-company 'HealthTech Solutions' 'HEALTH' 750000 'Healthcare Technology' 75000000 --from alice --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Creating HealthTech Solutions"
    
    sleep 3
    
    # FinanceFlow Corp
    execute_tx \
        "sharehodld tx equity create-company 'FinanceFlow Corp' 'FLOW' 1250000 'Financial Technology' 125000000 --from bob --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Creating FinanceFlow Corp"
    
    log_success "All test companies created"
}

# Issue initial shares
issue_shares() {
    log_info "Issuing initial shares..."
    
    # Issue SHODL shares
    execute_tx \
        "sharehodld tx equity issue-shares alice SHODL 50000 --from sharehodl-corp --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Issuing SHODL shares to Alice"
    
    sleep 2
    
    execute_tx \
        "sharehodld tx equity issue-shares bob SHODL 30000 --from sharehodl-corp --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Issuing SHODL shares to Bob"
    
    sleep 2
    
    # Issue TECH shares
    execute_tx \
        "sharehodld tx equity issue-shares charlie TECH 25000 --from techstart --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Issuing TECH shares to Charlie"
    
    sleep 2
    
    # Issue GREEN shares
    execute_tx \
        "sharehodld tx equity issue-shares diana GREEN 100000 --from greenenergy --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Issuing GREEN shares to Diana"
    
    log_success "Initial shares issued"
}

# Create liquidity pools
create_pools() {
    log_info "Creating liquidity pools..."
    
    # SHODL/HODL pool
    execute_tx \
        "sharehodld tx dex create-pool SHODL HODL 1000000000hodl 10000000000hodl --from sharehodl-corp --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Creating SHODL/HODL liquidity pool"
    
    sleep 3
    
    # TECH/HODL pool
    execute_tx \
        "sharehodld tx dex create-pool TECH HODL 500000000hodl 5000000000hodl --from techstart --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Creating TECH/HODL liquidity pool"
    
    sleep 3
    
    # GREEN/HODL pool
    execute_tx \
        "sharehodld tx dex create-pool GREEN HODL 2000000000hodl 10000000000hodl --from greenenergy --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Creating GREEN/HODL liquidity pool"
    
    log_success "Liquidity pools created"
}

# Place test orders
place_orders() {
    log_info "Placing test orders..."
    
    # Buy orders
    execute_tx \
        "sharehodld tx dex place-order LIMIT BUY SHODL/HODL 1000 100.0 GTC --from alice --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Placing SHODL buy order"
    
    sleep 2
    
    execute_tx \
        "sharehodld tx dex place-order LIMIT BUY TECH/HODL 2000 50.0 GTC --from bob --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Placing TECH buy order"
    
    sleep 2
    
    # Sell orders
    execute_tx \
        "sharehodld tx dex place-order LIMIT SELL GREEN/HODL 5000 25.0 GTC --from diana --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Placing GREEN sell order"
    
    log_success "Test orders placed"
}

# Submit governance proposals
submit_proposals() {
    log_info "Submitting governance proposals..."
    
    # Text proposal
    execute_tx \
        "sharehodld tx governance submit-proposal text 'ShareHODL Protocol Upgrade v1.1' 'Proposal to upgrade ShareHODL protocol to version 1.1 with enhanced DEX features' 10000000hodl --from alice --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Submitting protocol upgrade proposal"
    
    sleep 3
    
    # Parameter change proposal
    execute_tx \
        "sharehodld tx governance submit-proposal param-change 'Increase Block Gas Limit' 'Proposal to increase the maximum gas limit per block' 5000000hodl --from bob --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Submitting parameter change proposal"
    
    log_success "Governance proposals submitted"
}

# Mint test HODL tokens
mint_hodl_tokens() {
    log_info "Minting test HODL tokens..."
    
    # Mint HODL for testing
    execute_tx \
        "sharehodld tx hodl mint alice 1000000000hodl --from validator --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Minting HODL for Alice"
    
    sleep 2
    
    execute_tx \
        "sharehodld tx hodl mint bob 1000000000hodl --from validator --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Minting HODL for Bob"
    
    sleep 2
    
    execute_tx \
        "sharehodld tx hodl mint charlie 1000000000hodl --from validator --chain-id $CHAIN_ID --home $HOME_DIR" \
        "Minting HODL for Charlie"
    
    log_success "Test HODL tokens minted"
}

# Main execution
main() {
    log_info "========================================="
    log_info "   ShareHODL Test Data Population"
    log_info "========================================="
    
    wait_for_network
    
    log_info "Starting test data population..."
    sleep 2
    
    mint_hodl_tokens
    sleep 3
    
    create_companies
    sleep 5
    
    issue_shares
    sleep 5
    
    create_pools
    sleep 5
    
    place_orders
    sleep 3
    
    submit_proposals
    
    log_success "========================================="
    log_success "   Test Data Population Complete!"
    log_success "========================================="
    
    echo ""
    log_info "You can now:"
    echo "  - Query companies: sharehodld query equity company SHODL --home $HOME_DIR"
    echo "  - Check balances: sharehodld query bank balances \$(sharehodld keys show alice -a --home $HOME_DIR) --home $HOME_DIR"
    echo "  - View order book: sharehodld query dex order-book SHODL/HODL --home $HOME_DIR"
    echo "  - List proposals: sharehodld query governance proposals --home $HOME_DIR"
    echo "  - Access faucet: curl -X POST http://localhost:8080/faucet -d '{\"address\":\"YOUR_ADDRESS\"}'"
    echo ""
}

# Handle script arguments
case "${1:-}" in
    "companies")
        wait_for_network
        create_companies
        exit 0
        ;;
    "shares")
        wait_for_network
        issue_shares
        exit 0
        ;;
    "pools")
        wait_for_network
        create_pools
        exit 0
        ;;
    "orders")
        wait_for_network
        place_orders
        exit 0
        ;;
    "proposals")
        wait_for_network
        submit_proposals
        exit 0
        ;;
    *)
        main
        ;;
esac