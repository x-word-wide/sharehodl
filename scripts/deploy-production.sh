#!/bin/bash

# ShareHODL Blockchain Production Deployment Script
# Version: 1.0
# Last Updated: December 2024

set -euo pipefail

# Configuration
CHAIN_ID="sharehodl-1"
MONIKER="${MONIKER:-sharehodl-validator}"
NODE_HOME="${NODE_HOME:-$HOME/.sharehodl}"
GENESIS_FILE="$NODE_HOME/config/genesis.json"
CONFIG_FILE="$NODE_HOME/config/config.toml"
APP_FILE="$NODE_HOME/config/app.toml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging
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

# Validate environment
validate_environment() {
    log_info "Validating deployment environment..."

    # Check Go version
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_GO="1.23"
    if [[ "$(printf '%s\n' "$REQUIRED_GO" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_GO" ]]; then
        log_error "Go version $GO_VERSION is too old. Required: $REQUIRED_GO+"
        exit 1
    fi
    log_success "Go version $GO_VERSION is compatible"

    # Check required tools
    for tool in jq wget curl; do
        if ! command -v $tool &> /dev/null; then
            log_error "$tool is not installed"
            exit 1
        fi
    done

    # Check system resources
    AVAILABLE_MEMORY=$(free -g | awk '/^Mem:/{print $7}')
    if [[ $AVAILABLE_MEMORY -lt 8 ]]; then
        log_warning "Available memory is ${AVAILABLE_MEMORY}GB. Recommended: 8GB+"
    fi

    AVAILABLE_DISK=$(df -BG / | tail -n1 | awk '{print $4}' | sed 's/G//')
    if [[ $AVAILABLE_DISK -lt 100 ]]; then
        log_warning "Available disk space is ${AVAILABLE_DISK}GB. Recommended: 100GB+"
    fi

    log_success "Environment validation passed"
}

# Build the blockchain binary
build_binary() {
    log_info "Building ShareHODL blockchain binary..."

    # Clean previous builds
    go clean -cache
    go mod tidy

    # Generate protobuf files
    log_info "Generating protobuf files..."
    if command -v buf &> /dev/null; then
        buf generate
    else
        log_warning "buf not found, installing..."
        go install github.com/bufbuild/buf/cmd/buf@latest
        ~/go/bin/buf generate
    fi

    # Build the main binary
    log_info "Compiling sharehodld binary..."
    go build -o build/sharehodld ./cmd/sharehodld

    # Verify build
    if [[ ! -f "build/sharehodld" ]]; then
        log_error "Binary build failed"
        exit 1
    fi

    ./build/sharehodld version
    log_success "Binary built successfully"
}

# Initialize validator node
init_validator() {
    log_info "Initializing validator node..."

    # Remove existing configuration
    if [[ -d "$NODE_HOME" ]]; then
        log_warning "Removing existing node configuration at $NODE_HOME"
        rm -rf "$NODE_HOME"
    fi

    # Initialize node
    ./build/sharehodld init "$MONIKER" --chain-id="$CHAIN_ID" --home="$NODE_HOME"

    log_success "Validator node initialized"
}

# Configure the node
configure_node() {
    log_info "Configuring node parameters..."

    # Update config.toml
    sed -i.bak "s/^moniker = .*/moniker = \"$MONIKER\"/" "$CONFIG_FILE"
    sed -i.bak 's/^timeout_commit = .*/timeout_commit = "2s"/' "$CONFIG_FILE"
    sed -i.bak 's/^timeout_propose = .*/timeout_propose = "2s"/' "$CONFIG_FILE"
    sed -i.bak 's/^peer_gossip_sleep_duration = .*/peer_gossip_sleep_duration = "25ms"/' "$CONFIG_FILE"
    sed -i.bak 's/^peer_query_maj23_sleep_duration = .*/peer_query_maj23_sleep_duration = "1s"/' "$CONFIG_FILE"

    # Update app.toml for low fees
    sed -i.bak 's/^minimum-gas-prices = .*/minimum-gas-prices = "0.025uhodl"/' "$APP_FILE"
    sed -i.bak 's/^pruning = .*/pruning = "custom"/' "$APP_FILE"
    sed -i.bak 's/^pruning-keep-recent = .*/pruning-keep-recent = "100"/' "$APP_FILE"
    sed -i.bak 's/^pruning-interval = .*/pruning-interval = "10"/' "$APP_FILE"

    log_success "Node configuration updated"
}

# Setup genesis file
setup_genesis() {
    log_info "Setting up genesis file..."

    # Copy the prepared genesis file
    if [[ -f "genesis.json" ]]; then
        cp genesis.json "$GENESIS_FILE"
    else
        log_error "Genesis file not found in project root"
        exit 1
    fi

    # Validate genesis
    ./build/sharehodld validate-genesis --home="$NODE_HOME"
    log_success "Genesis file validated"
}

# Setup systemd service
setup_systemd() {
    log_info "Setting up systemd service..."

    BINARY_PATH="$(pwd)/build/sharehodld"
    SERVICE_FILE="/etc/systemd/system/sharehodld.service"

    sudo tee "$SERVICE_FILE" > /dev/null <<EOF
[Unit]
Description=ShareHODL Blockchain Node
After=network.target

[Service]
Type=simple
User=$USER
ExecStart=$BINARY_PATH start --home=$NODE_HOME
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable sharehodld
    log_success "Systemd service configured"
}

# Setup monitoring
setup_monitoring() {
    log_info "Setting up monitoring..."

    # Create monitoring configuration
    mkdir -p "$NODE_HOME/monitoring"

    # Prometheus configuration for metrics scraping
    tee "$NODE_HOME/monitoring/prometheus.yml" > /dev/null <<EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'sharehodl'
    static_configs:
      - targets: ['localhost:26660']
    scrape_interval: 5s
    metrics_path: /metrics

  - job_name: 'tendermint'
    static_configs:
      - targets: ['localhost:26660']
    scrape_interval: 5s
    metrics_path: /metrics

rule_files:
  - "alert_rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - localhost:9093
EOF

    # Alert rules
    tee "$NODE_HOME/monitoring/alert_rules.yml" > /dev/null <<EOF
groups:
  - name: sharehodl_alerts
    rules:
      - alert: NodeDown
        expr: up{job="sharehodl"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "ShareHODL node is down"

      - alert: HighBlockTime
        expr: tendermint_consensus_block_interval_seconds > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Block time is unusually high"

      - alert: LowPeerCount
        expr: tendermint_p2p_peers < 3
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low peer count detected"
EOF

    log_success "Monitoring configuration created"
}

# Security hardening
security_hardening() {
    log_info "Applying security hardening..."

    # Set secure file permissions
    chmod 600 "$NODE_HOME/config/priv_validator_key.json"
    chmod 600 "$NODE_HOME/config/node_key.json"
    chmod 644 "$NODE_HOME/config/genesis.json"
    chmod 644 "$CONFIG_FILE"
    chmod 644 "$APP_FILE"

    # Firewall rules (if ufw is available)
    if command -v ufw &> /dev/null; then
        sudo ufw default deny incoming
        sudo ufw default allow outgoing
        sudo ufw allow 22/tcp    # SSH
        sudo ufw allow 26656/tcp # P2P
        sudo ufw allow 26657/tcp # RPC (restrict in production)
        sudo ufw allow 1317/tcp  # REST API
        sudo ufw allow 9090/tcp  # gRPC
        sudo ufw --force enable
        log_success "Firewall configured"
    fi

    log_success "Security hardening applied"
}

# Deploy frontend applications
deploy_frontend() {
    log_info "Deploying frontend applications..."

    if [[ -d "sharehodl-frontend" ]]; then
        cd sharehodl-frontend

        # Install dependencies
        npm ci

        # Build all applications
        npm run build

        # TODO: Deploy to web server or CDN
        log_success "Frontend built successfully"
        cd ..
    else
        log_warning "Frontend directory not found, skipping deployment"
    fi
}

# Start the node
start_node() {
    log_info "Starting ShareHODL node..."

    # Start via systemd
    sudo systemctl start sharehodld
    sleep 5

    # Check status
    if sudo systemctl is-active --quiet sharehodld; then
        log_success "Node started successfully"
        
        # Show logs
        log_info "Node logs (last 20 lines):"
        sudo journalctl -u sharehodld -n 20 --no-pager
        
        # Show status
        sleep 10
        ./build/sharehodld status --node tcp://localhost:26657
    else
        log_error "Failed to start node"
        sudo journalctl -u sharehodld -n 50 --no-pager
        exit 1
    fi
}

# Health checks
health_checks() {
    log_info "Running health checks..."

    # Check if node is syncing
    STATUS=$(./build/sharehodld status --node tcp://localhost:26657 2>/dev/null || echo "ERROR")
    if [[ "$STATUS" == "ERROR" ]]; then
        log_error "Cannot connect to node"
        return 1
    fi

    CATCHING_UP=$(echo "$STATUS" | jq -r '.sync_info.catching_up')
    LATEST_HEIGHT=$(echo "$STATUS" | jq -r '.sync_info.latest_block_height')

    log_info "Node status:"
    echo "  Height: $LATEST_HEIGHT"
    echo "  Catching up: $CATCHING_UP"

    # Check API endpoints
    if curl -s http://localhost:1317/cosmos/base/tendermint/v1beta1/node_info >/dev/null; then
        log_success "REST API is accessible"
    else
        log_warning "REST API not accessible"
    fi

    if curl -s http://localhost:26657/status >/dev/null; then
        log_success "RPC endpoint is accessible"
    else
        log_warning "RPC endpoint not accessible"
    fi

    log_success "Health checks completed"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    rm -f "$CONFIG_FILE.bak" "$APP_FILE.bak"
}

# Main deployment function
main() {
    echo "🚀 ShareHODL Blockchain Production Deployment"
    echo "=============================================="
    
    log_info "Starting deployment process..."
    
    validate_environment
    build_binary
    init_validator
    configure_node
    setup_genesis
    setup_systemd
    setup_monitoring
    security_hardening
    deploy_frontend
    start_node
    health_checks
    cleanup
    
    echo
    log_success "🎉 Deployment completed successfully!"
    echo
    echo "Node Information:"
    echo "  Chain ID: $CHAIN_ID"
    echo "  Moniker: $MONIKER"
    echo "  Home: $NODE_HOME"
    echo "  RPC: http://localhost:26657"
    echo "  REST: http://localhost:1317"
    echo
    echo "Commands:"
    echo "  Status: ./build/sharehodld status"
    echo "  Logs: sudo journalctl -u sharehodld -f"
    echo "  Stop: sudo systemctl stop sharehodld"
    echo "  Restart: sudo systemctl restart sharehodld"
    echo
    echo "⚠️  Important: Backup your validator key at $NODE_HOME/config/priv_validator_key.json"
}

# Handle script termination
trap cleanup EXIT

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi