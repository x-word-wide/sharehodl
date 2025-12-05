#!/bin/bash

# ShareHODL Local Network Setup Script
# This script initializes a local blockchain network for development and testing

set -e

# Configuration
CHAIN_ID="sharehodl-local-1"
NODE_NAME="sharehodl-node"
VALIDATOR_NAME="validator"
GENESIS_ACCOUNTS_FILE="scripts/localnet/genesis-accounts.json"
TEST_DATA_FILE="tests/e2e/testdata/test_data.json"
HOME_DIR="$HOME/.sharehodl"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
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

# Check if sharehodld binary exists
check_binary() {
    if ! command -v sharehodld &> /dev/null; then
        log_error "sharehodld binary not found. Please build the project first:"
        echo "  make build"
        exit 1
    fi
    log_success "sharehodld binary found"
}

# Clean previous setup
clean_setup() {
    log_info "Cleaning previous setup..."
    if [ -d "$HOME_DIR" ]; then
        rm -rf "$HOME_DIR"
        log_success "Removed existing home directory"
    fi
}

# Initialize node
init_node() {
    log_info "Initializing node..."
    sharehodld init $NODE_NAME --chain-id $CHAIN_ID --home $HOME_DIR
    log_success "Node initialized"
}

# Create test accounts
create_accounts() {
    log_info "Creating test accounts..."
    
    # Create validator account
    echo -e "\n\n\n\n\n\n\n\n\n\n\n\n" | sharehodld keys add $VALIDATOR_NAME --home $HOME_DIR
    VALIDATOR_ADDR=$(sharehodld keys show $VALIDATOR_NAME -a --home $HOME_DIR)
    log_success "Validator account created: $VALIDATOR_ADDR"
    
    # Create test accounts for businesses and users
    declare -a ACCOUNTS=("alice" "bob" "charlie" "diana" "sharehodl-corp" "techstart" "greenenergy")
    
    for account in "${ACCOUNTS[@]}"; do
        echo -e "\n\n\n\n\n\n\n\n\n\n\n\n" | sharehodld keys add $account --home $HOME_DIR
        addr=$(sharehodld keys show $account -a --home $HOME_DIR)
        log_success "Account created: $account ($addr)"
    done
}

# Configure genesis
configure_genesis() {
    log_info "Configuring genesis file..."
    
    GENESIS_FILE="$HOME_DIR/config/genesis.json"
    
    # Add validator to genesis
    sharehodld add-genesis-account $VALIDATOR_NAME 100000000000000hodl,100000000000000shodl --home $HOME_DIR
    
    # Add test accounts with initial balances
    declare -a ACCOUNTS=("alice" "bob" "charlie" "diana" "sharehodl-corp" "techstart" "greenenergy")
    
    for account in "${ACCOUNTS[@]}"; do
        sharehodld add-genesis-account $account 10000000000hodl,1000000000shodl --home $HOME_DIR
    done
    
    # Generate genesis transaction
    sharehodld gentx $VALIDATOR_NAME 10000000000hodl --chain-id $CHAIN_ID --home $HOME_DIR
    sharehodld collect-gentxs --home $HOME_DIR
    
    log_success "Genesis configured with test accounts"
}

# Configure node settings
configure_node() {
    log_info "Configuring node settings..."
    
    CONFIG_FILE="$HOME_DIR/config/config.toml"
    APP_FILE="$HOME_DIR/config/app.toml"
    
    # Update config.toml
    sed -i.bak 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $CONFIG_FILE
    sed -i.bak 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $CONFIG_FILE
    sed -i.bak 's/cors_allowed_origins = \[\]/cors_allowed_origins = ["*"]/g' $CONFIG_FILE
    sed -i.bak 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/g' $CONFIG_FILE
    
    # Update app.toml for API and gRPC
    sed -i.bak 's/enable = false/enable = true/g' $APP_FILE
    sed -i.bak 's/enabled-unsafe-cors = false/enabled-unsafe-cors = true/g' $APP_FILE
    sed -i.bak 's/address = "tcp:\/\/0.0.0.0:1317"/address = "tcp:\/\/0.0.0.0:1317"/g' $APP_FILE
    
    log_success "Node configuration updated"
}

# Setup test data
setup_test_data() {
    log_info "Setting up test data..."
    
    # Create test companies data script
    cat > $HOME_DIR/setup-test-data.sh << 'EOF'
#!/bin/bash
# This script will be run after the network starts to populate test data

CHAIN_ID="sharehodl-local-1"
HOME_DIR="$HOME/.sharehodl"

echo "Waiting for network to start..."
sleep 5

# Create test companies
echo "Creating test companies..."
sharehodld tx equity create-company "ShareHODL Corp" "SHODL" 1000000 "Blockchain Technology" 100000000 --from sharehodl-corp --chain-id $CHAIN_ID --home $HOME_DIR --yes

sharehodld tx equity create-company "TechStart Inc" "TECH" 500000 "Software" 50000000 --from techstart --chain-id $CHAIN_ID --home $HOME_DIR --yes

sharehodl tx equity create-company "GreenEnergy LLC" "GREEN" 2000000 "Renewable Energy" 200000000 --from greenenergy --chain-id $CHAIN_ID --home $HOME_DIR --yes

echo "Test data setup complete!"
EOF
    
    chmod +x $HOME_DIR/setup-test-data.sh
    log_success "Test data script created"
}

# Create startup script
create_startup_script() {
    log_info "Creating startup script..."
    
    cat > $HOME_DIR/start-node.sh << EOF
#!/bin/bash
echo "Starting ShareHODL local network..."
sharehodld start --home $HOME_DIR --pruning nothing --log_level info
EOF
    
    chmod +x $HOME_DIR/start-node.sh
    log_success "Startup script created"
}

# Main setup function
main() {
    log_info "========================================="
    log_info "   ShareHODL Local Network Setup"
    log_info "========================================="
    
    check_binary
    clean_setup
    init_node
    create_accounts
    configure_genesis
    configure_node
    setup_test_data
    create_startup_script
    
    log_success "========================================="
    log_success "   Setup Complete!"
    log_success "========================================="
    
    echo ""
    log_info "To start the network:"
    echo "  $HOME_DIR/start-node.sh"
    echo ""
    log_info "Or run directly:"
    echo "  sharehodld start --home $HOME_DIR"
    echo ""
    log_info "To populate test data (after network starts):"
    echo "  $HOME_DIR/setup-test-data.sh"
    echo ""
    log_info "Network endpoints:"
    echo "  RPC:  http://localhost:26657"
    echo "  API:  http://localhost:1317"
    echo "  gRPC: localhost:9090"
    echo ""
    log_info "View accounts:"
    echo "  sharehodld keys list --home $HOME_DIR"
}

# Handle script arguments
case "${1:-}" in
    "clean")
        clean_setup
        log_success "Cleanup complete"
        exit 0
        ;;
    "accounts")
        sharehodld keys list --home $HOME_DIR
        exit 0
        ;;
    "status")
        if [ -d "$HOME_DIR" ]; then
            log_info "Local network is configured"
            log_info "Home directory: $HOME_DIR"
            log_info "Chain ID: $CHAIN_ID"
        else
            log_warning "Local network not configured. Run setup first."
        fi
        exit 0
        ;;
    *)
        main
        ;;
esac