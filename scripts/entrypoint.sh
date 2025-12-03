#!/bin/bash

# ShareHODL Node Entrypoint Script
# Handles initialization and startup of ShareHODL blockchain nodes

set -euo pipefail

# Configuration
SHAREHODL_HOME="${SHAREHODL_HOME:-/home/sharehodl/.sharehodl}"
CONFIG_DIR="${CONFIG_DIR:-/config}"
NODE_ID="${NODE_ID:-0}"
ENVIRONMENT="${ENVIRONMENT:-development}"
NETWORK="${NETWORK:-testnet}"
CHAIN_ID="${CHAIN_ID:-sharehodl-$NETWORK-1}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Initialize node if not already initialized
initialize_node() {
    if [[ ! -d "$SHAREHODL_HOME/config" ]]; then
        log_info "Initializing ShareHODL node..."
        
        sharehodld init "sharehodl-node-$NODE_ID" \
            --chain-id "$CHAIN_ID" \
            --home "$SHAREHODL_HOME"
        
        log_success "Node initialized"
    else
        log_info "Node already initialized, skipping initialization"
    fi
}

# Copy configuration files
setup_config() {
    log_info "Setting up configuration..."
    
    # Copy genesis file if available
    if [[ -f "$CONFIG_DIR/genesis.json" ]]; then
        cp "$CONFIG_DIR/genesis.json" "$SHAREHODL_HOME/config/"
        log_info "Genesis file copied"
    fi
    
    # Setup node-specific configuration
    setup_node_config
    
    # Setup seeds and persistent peers
    setup_networking
    
    # Setup keyring if validator keys exist
    setup_validator_keys
    
    log_success "Configuration setup completed"
}

# Setup node-specific configuration
setup_node_config() {
    local config_file="$SHAREHODL_HOME/config/config.toml"
    local app_file="$SHAREHODL_HOME/config/app.toml"
    
    if [[ -f "$config_file" ]]; then
        # Update node configuration
        sed -i "s/^moniker = .*/moniker = \"sharehodl-node-$NODE_ID\"/" "$config_file"
        
        # Set ports based on node ID
        local p2p_port=$((26656 + NODE_ID))
        local rpc_port=$((26657 + NODE_ID))
        local grpc_port=$((9090 + NODE_ID))
        local api_port=$((1317 + NODE_ID))
        
        sed -i "s/^laddr = \"tcp:\/\/0.0.0.0:26656\"/laddr = \"tcp:\/\/0.0.0.0:$p2p_port\"/" "$config_file"
        sed -i "s/^laddr = \"tcp:\/\/127.0.0.1:26657\"/laddr = \"tcp:\/\/0.0.0.0:$rpc_port\"/" "$config_file"
        
        # Enable prometheus metrics
        sed -i 's/^prometheus = false/prometheus = true/' "$config_file"
        sed -i 's/^prometheus_listen_addr = ":26660"/prometheus_listen_addr = ":26660"/' "$config_file"
        
        # Set logging level based on environment
        case "$ENVIRONMENT" in
            "development")
                sed -i 's/^log_level = .*/log_level = "debug"/' "$config_file"
                ;;
            "staging")
                sed -i 's/^log_level = .*/log_level = "info"/' "$config_file"
                ;;
            "production")
                sed -i 's/^log_level = .*/log_level = "warn"/' "$config_file"
                ;;
        esac
    fi
    
    if [[ -f "$app_file" ]]; then
        # Update app configuration
        sed -i "s/^address = \"tcp:\/\/0.0.0.0:1317\"/address = \"tcp:\/\/0.0.0.0:$api_port\"/" "$app_file"
        sed -i "s/^address = \"0.0.0.0:9090\"/address = \"0.0.0.0:$grpc_port\"/" "$app_file"
        
        # Enable API and gRPC
        sed -i 's/^enable = false/enable = true/' "$app_file"
        sed -i 's/^enabled-unsafe-cors = false/enabled-unsafe-cors = true/' "$app_file"
    fi
}

# Setup networking configuration
setup_networking() {
    local config_file="$SHAREHODL_HOME/config/config.toml"
    
    if [[ -f "$config_file" ]]; then
        # Set up seeds and persistent peers for multi-node setup
        local persistent_peers=""
        
        # Build persistent peers list (exclude self)
        for i in {0..2}; do
            if [[ $i -ne $NODE_ID ]]; then
                if [[ -n "$persistent_peers" ]]; then
                    persistent_peers="$persistent_peers,"
                fi
                # Use service names for Docker/K8s environments
                local peer_addr="sharehodl-node-$i:$((26656 + i))"
                # Generate a dummy node ID for peer connection
                local peer_id=$(echo -n "sharehodl-node-$i" | sha256sum | cut -c1-40)
                persistent_peers="$persistent_peers$peer_id@$peer_addr"
            fi
        done
        
        if [[ -n "$persistent_peers" ]]; then
            sed -i "s/^persistent_peers = .*/persistent_peers = \"$persistent_peers\"/" "$config_file"
        fi
        
        # Allow duplicate IP connections for local development
        if [[ "$ENVIRONMENT" == "development" ]]; then
            sed -i 's/^allow_duplicate_ip = false/allow_duplicate_ip = true/' "$config_file"
        fi
        
        # Set external address for production
        if [[ "$ENVIRONMENT" == "production" && -n "${EXTERNAL_IP:-}" ]]; then
            sed -i "s/^external_address = .*/external_address = \"$EXTERNAL_IP:$((26656 + NODE_ID))\"/" "$config_file"
        fi
    fi
}

# Setup validator keys if available
setup_validator_keys() {
    local validator_dir="$CONFIG_DIR/validators/validator-$NODE_ID"
    
    if [[ -d "$validator_dir" ]]; then
        log_info "Setting up validator keys..."
        
        # Copy validator key files
        if [[ -f "$validator_dir/config/priv_validator_key.json" ]]; then
            cp "$validator_dir/config/priv_validator_key.json" "$SHAREHODL_HOME/config/"
        fi
        
        if [[ -f "$validator_dir/config/node_key.json" ]]; then
            cp "$validator_dir/config/node_key.json" "$SHAREHODL_HOME/config/"
        fi
        
        # Copy keyring files
        if [[ -d "$validator_dir/keyring-test" ]]; then
            cp -r "$validator_dir/keyring-test" "$SHAREHODL_HOME/"
        fi
        
        log_success "Validator keys configured"
    fi
}

# Wait for other nodes to be ready (for multi-node deployments)
wait_for_peers() {
    if [[ "$NODE_ID" != "0" ]]; then
        log_info "Waiting for peer nodes to be ready..."
        
        local max_attempts=30
        local attempt=0
        
        while [[ $attempt -lt $max_attempts ]]; do
            # Check if node-0 is responsive
            if curl -f -s "http://sharehodl-node-0:26657/status" >/dev/null 2>&1; then
                log_success "Peer nodes are ready"
                return 0
            fi
            
            log_info "Waiting for peers... ($((attempt + 1))/$max_attempts)"
            sleep 10
            ((attempt++))
        done
        
        log_warning "Peers not ready, continuing anyway"
    fi
}

# Run pre-start checks
run_pre_start_checks() {
    log_info "Running pre-start checks..."
    
    # Check if genesis file exists
    if [[ ! -f "$SHAREHODL_HOME/config/genesis.json" ]]; then
        log_error "Genesis file not found"
        exit 1
    fi
    
    # Validate genesis file
    if ! sharehodld validate-genesis --home "$SHAREHODL_HOME" >/dev/null 2>&1; then
        log_error "Genesis validation failed"
        exit 1
    fi
    
    # Check configuration files
    if [[ ! -f "$SHAREHODL_HOME/config/config.toml" ]]; then
        log_error "config.toml not found"
        exit 1
    fi
    
    if [[ ! -f "$SHAREHODL_HOME/config/app.toml" ]]; then
        log_error "app.toml not found"
        exit 1
    fi
    
    log_success "Pre-start checks passed"
}

# Start ShareHODL node
start_node() {
    log_info "Starting ShareHODL node..."
    
    # Start with appropriate flags based on environment
    local start_cmd="sharehodld start --home $SHAREHODL_HOME"
    
    # Add environment-specific flags
    case "$ENVIRONMENT" in
        "development")
            start_cmd="$start_cmd --log_level debug"
            ;;
        "production")
            start_cmd="$start_cmd --log_level warn"
            ;;
    esac
    
    # Add pruning configuration for production
    if [[ "$ENVIRONMENT" == "production" ]]; then
        start_cmd="$start_cmd --pruning custom --pruning-keep-recent 100 --pruning-interval 10"
    fi
    
    log_success "ShareHODL node starting with command: $start_cmd"
    exec $start_cmd
}

# Handle different commands
handle_command() {
    case "$1" in
        "start")
            initialize_node
            setup_config
            wait_for_peers
            run_pre_start_checks
            start_node
            ;;
        "init")
            initialize_node
            setup_config
            log_success "Node initialization completed"
            ;;
        "validate")
            run_pre_start_checks
            log_success "Validation completed"
            ;;
        "version")
            sharehodld version
            ;;
        "keys")
            shift
            sharehodld keys "$@" --home "$SHAREHODL_HOME"
            ;;
        "status")
            sharehodld status --home "$SHAREHODL_HOME"
            ;;
        *)
            log_info "Running custom command: $*"
            exec sharehodld "$@" --home "$SHAREHODL_HOME"
            ;;
    esac
}

# Create necessary directories
mkdir -p "$SHAREHODL_HOME/config" "$SHAREHODL_HOME/data"

# Handle the command
if [[ $# -eq 0 ]]; then
    handle_command "start"
else
    handle_command "$@"
fi