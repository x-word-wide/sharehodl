#!/bin/bash

# ShareHODL Blockchain Deployment Script
# This script automates the deployment of ShareHODL blockchain infrastructure

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_DIR="$PROJECT_ROOT/deployment/config"
COMPOSE_FILE="$PROJECT_ROOT/deployment/docker-compose.yml"
TERRAFORM_DIR="$PROJECT_ROOT/deployment/terraform"

# Default values
ENVIRONMENT="development"
NETWORK="testnet"
NODES=3
VALIDATORS=3
GENESIS_ACCOUNTS=5
DEPLOYMENT_TYPE="docker"
MONITORING_ENABLED=true
SECURITY_SCANNING=true
BACKUP_ENABLED=true

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

# Help function
show_help() {
    cat << EOF
ShareHODL Blockchain Deployment Script

Usage: $0 [OPTIONS]

OPTIONS:
    -e, --environment ENVIRONMENT    Deployment environment (development, staging, production)
    -n, --network NETWORK           Network type (testnet, mainnet)
    -N, --nodes COUNT               Number of nodes to deploy (default: 3)
    -V, --validators COUNT          Number of validators (default: 3)
    -t, --type TYPE                 Deployment type (docker, kubernetes, terraform)
    -m, --monitoring                Enable monitoring stack
    -s, --security                  Enable security scanning
    -b, --backup                    Enable backup systems
    -h, --help                      Show this help message

EXAMPLES:
    # Development deployment with Docker
    $0 -e development -t docker

    # Production deployment with Kubernetes
    $0 -e production -n mainnet -t kubernetes -N 5 -V 3 -m -s -b

    # Staging deployment with Terraform
    $0 -e staging -t terraform --monitoring --security
EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--network)
                NETWORK="$2"
                shift 2
                ;;
            -N|--nodes)
                NODES="$2"
                shift 2
                ;;
            -V|--validators)
                VALIDATORS="$2"
                shift 2
                ;;
            -t|--type)
                DEPLOYMENT_TYPE="$2"
                shift 2
                ;;
            -m|--monitoring)
                MONITORING_ENABLED=true
                shift
                ;;
            -s|--security)
                SECURITY_SCANNING=true
                shift
                ;;
            -b|--backup)
                BACKUP_ENABLED=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Validate prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check if required tools are installed
    local required_tools=("docker" "docker-compose" "jq" "curl")
    
    if [[ "$DEPLOYMENT_TYPE" == "kubernetes" ]]; then
        required_tools+=("kubectl" "helm")
    elif [[ "$DEPLOYMENT_TYPE" == "terraform" ]]; then
        required_tools+=("terraform" "aws")
    fi

    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_error "$tool is required but not installed"
            exit 1
        fi
    done

    # Check Docker daemon
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi

    # Validate environment
    if [[ "$ENVIRONMENT" != "development" && "$ENVIRONMENT" != "staging" && "$ENVIRONMENT" != "production" ]]; then
        log_error "Invalid environment: $ENVIRONMENT"
        exit 1
    fi

    # Validate network
    if [[ "$NETWORK" != "testnet" && "$NETWORK" != "mainnet" ]]; then
        log_error "Invalid network: $NETWORK"
        exit 1
    fi

    # Validate node counts
    if [[ $NODES -lt 1 || $VALIDATORS -lt 1 || $VALIDATORS -gt $NODES ]]; then
        log_error "Invalid node/validator configuration"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

# Build ShareHODL binary
build_sharehodl() {
    log_info "Building ShareHODL blockchain binary..."

    cd "$PROJECT_ROOT"

    # Clean previous builds
    make clean || true

    # Build the binary
    if ! make build; then
        log_error "Failed to build ShareHODL binary"
        exit 1
    fi

    # Verify binary
    if [[ ! -f "build/sharehodld" ]]; then
        log_error "ShareHODL binary not found after build"
        exit 1
    fi

    # Test binary
    if ! ./build/sharehodld version &> /dev/null; then
        log_error "ShareHODL binary is not executable or corrupted"
        exit 1
    fi

    log_success "ShareHODL binary built successfully"
}

# Generate genesis configuration
generate_genesis() {
    log_info "Generating genesis configuration..."

    local genesis_file="$CONFIG_DIR/genesis.json"
    local validators_dir="$CONFIG_DIR/validators"
    
    mkdir -p "$CONFIG_DIR" "$validators_dir"

    # Initialize chain
    cd "$PROJECT_ROOT"
    ./build/sharehodld init "sharehodl-$ENVIRONMENT" --chain-id "sharehodl-$NETWORK-1" --home "$CONFIG_DIR"

    # Generate validator keys
    for i in $(seq 0 $((VALIDATORS - 1))); do
        local validator_dir="$validators_dir/validator-$i"
        mkdir -p "$validator_dir"
        
        # Generate validator key
        ./build/sharehodld keys add "validator-$i" --keyring-backend test --home "$validator_dir"
        
        # Get validator address
        local validator_addr=$(./build/sharehodl keys show "validator-$i" -a --keyring-backend test --home "$validator_dir")
        
        # Add validator to genesis
        ./build/sharehodld add-genesis-account "$validator_addr" 100000000000000uhodl,1000000000000ustake --home "$CONFIG_DIR"
        
        # Generate gentx
        ./build/sharehodld gentx "validator-$i" 1000000000ustake --chain-id "sharehodl-$NETWORK-1" --keyring-backend test --home "$validator_dir"
        
        # Copy gentx to genesis
        cp "$validator_dir/config/gentx"/*.json "$CONFIG_DIR/config/gentx/"
    done

    # Generate additional accounts
    for i in $(seq 0 $((GENESIS_ACCOUNTS - 1))); do
        local account_name="account-$i"
        ./build/sharehodl keys add "$account_name" --keyring-backend test --home "$CONFIG_DIR"
        local account_addr=$(./build/sharehodl keys show "$account_name" -a --keyring-backend test --home "$CONFIG_DIR")
        ./build/sharehodl add-genesis-account "$account_addr" 50000000000000uhodl --home "$CONFIG_DIR"
    done

    # Collect gentxs
    ./build/sharehodl collect-gentxs --home "$CONFIG_DIR"

    # Validate genesis
    if ! ./build/sharehodl validate-genesis --home "$CONFIG_DIR"; then
        log_error "Genesis validation failed"
        exit 1
    fi

    log_success "Genesis configuration generated"
}

# Deploy with Docker
deploy_docker() {
    log_info "Deploying ShareHODL with Docker..."

    # Create docker-compose file
    generate_docker_compose

    # Pull/build images
    docker-compose -f "$COMPOSE_FILE" pull
    docker-compose -f "$COMPOSE_FILE" build

    # Start services
    docker-compose -f "$COMPOSE_FILE" up -d

    # Wait for services to be ready
    wait_for_services_docker

    log_success "Docker deployment completed"
}

# Deploy with Kubernetes
deploy_kubernetes() {
    log_info "Deploying ShareHODL with Kubernetes..."

    local k8s_dir="$PROJECT_ROOT/deployment/kubernetes"
    
    # Apply namespace
    kubectl apply -f "$k8s_dir/namespace.yaml"

    # Apply configmaps and secrets
    kubectl apply -f "$k8s_dir/configmaps/"
    kubectl apply -f "$k8s_dir/secrets/"

    # Deploy persistent volumes
    kubectl apply -f "$k8s_dir/storage/"

    # Deploy services
    kubectl apply -f "$k8s_dir/services/"

    # Deploy deployments
    kubectl apply -f "$k8s_dir/deployments/"

    # Deploy ingress if enabled
    if [[ -f "$k8s_dir/ingress.yaml" ]]; then
        kubectl apply -f "$k8s_dir/ingress.yaml"
    fi

    # Wait for pods to be ready
    wait_for_services_kubernetes

    log_success "Kubernetes deployment completed"
}

# Deploy with Terraform
deploy_terraform() {
    log_info "Deploying ShareHODL with Terraform..."

    cd "$TERRAFORM_DIR"

    # Initialize Terraform
    terraform init

    # Plan deployment
    terraform plan -var="environment=$ENVIRONMENT" -var="network=$NETWORK" -var="node_count=$NODES"

    # Apply deployment
    if ! terraform apply -auto-approve -var="environment=$ENVIRONMENT" -var="network=$NETWORK" -var="node_count=$NODES"; then
        log_error "Terraform deployment failed"
        exit 1
    fi

    # Get outputs
    terraform output

    log_success "Terraform deployment completed"
}

# Generate Docker Compose configuration
generate_docker_compose() {
    log_info "Generating Docker Compose configuration..."

    cat > "$COMPOSE_FILE" << EOF
version: '3.8'

services:
  # ShareHODL Nodes
$(for i in $(seq 0 $((NODES - 1))); do
cat << NODE_EOF
  sharehodl-node-$i:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: sharehodl-node-$i
    environment:
      - NODE_ID=$i
      - ENVIRONMENT=$ENVIRONMENT
      - NETWORK=$NETWORK
    ports:
      - "$((26656 + i)):26656"
      - "$((26657 + i)):26657"
      - "$((1317 + i)):1317"
      - "$((9090 + i)):9090"
    volumes:
      - sharehodl-node-$i-data:/root/.sharehodl
      - ./deployment/config:/config:ro
    networks:
      - sharehodl-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:26657/status"]
      interval: 30s
      timeout: 10s
      retries: 3

NODE_EOF
done)

$(if [[ "$MONITORING_ENABLED" == "true" ]]; then
cat << MONITORING_EOF
  # Monitoring Stack
  prometheus:
    image: prom/prometheus:latest
    container_name: sharehodl-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./deployment/monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    networks:
      - sharehodl-network
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: sharehodl-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana
      - ./deployment/monitoring/grafana:/etc/grafana/provisioning:ro
    networks:
      - sharehodl-network
    restart: unless-stopped

  alertmanager:
    image: prom/alertmanager:latest
    container_name: sharehodl-alertmanager
    ports:
      - "9093:9093"
    volumes:
      - ./deployment/monitoring/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro
    networks:
      - sharehodl-network
    restart: unless-stopped

MONITORING_EOF
fi)

$(if [[ "$SECURITY_SCANNING" == "true" ]]; then
cat << SECURITY_EOF
  # Security Services
  vault:
    image: vault:latest
    container_name: sharehodl-vault
    ports:
      - "8200:8200"
    environment:
      - VAULT_DEV_ROOT_TOKEN_ID=sharehodl-dev-token
      - VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200
    cap_add:
      - IPC_LOCK
    networks:
      - sharehodl-network
    restart: unless-stopped

  security-scanner:
    build:
      context: .
      dockerfile: Dockerfile.security
    container_name: sharehodl-security-scanner
    environment:
      - SCAN_INTERVAL=3600
      - ALERT_WEBHOOK=http://alertmanager:9093/api/v1/alerts
    volumes:
      - ./:/app/source:ro
    networks:
      - sharehodl-network
    restart: unless-stopped

SECURITY_EOF
fi)

$(if [[ "$BACKUP_ENABLED" == "true" ]]; then
cat << BACKUP_EOF
  # Backup Service
  backup-service:
    image: alpine:latest
    container_name: sharehodl-backup
    environment:
      - BACKUP_INTERVAL=86400
      - S3_BUCKET=sharehodl-backups-$ENVIRONMENT
    volumes:
      - sharehodl-node-0-data:/data/node-0:ro
      - sharehodl-node-1-data:/data/node-1:ro
      - sharehodl-node-2-data:/data/node-2:ro
      - ./scripts/backup.sh:/backup.sh:ro
    entrypoint: ["sh", "/backup.sh"]
    networks:
      - sharehodl-network
    restart: unless-stopped

BACKUP_EOF
fi)

networks:
  sharehodl-network:
    driver: bridge

volumes:
$(for i in $(seq 0 $((NODES - 1))); do
echo "  sharehodl-node-$i-data:"
done)
$(if [[ "$MONITORING_ENABLED" == "true" ]]; then
echo "  prometheus-data:"
echo "  grafana-data:"
fi)
EOF

    log_success "Docker Compose configuration generated"
}

# Wait for Docker services
wait_for_services_docker() {
    log_info "Waiting for services to be ready..."

    local max_attempts=60
    local attempt=0

    while [[ $attempt -lt $max_attempts ]]; do
        if docker-compose -f "$COMPOSE_FILE" ps | grep -q "Up.*healthy"; then
            log_success "Services are ready"
            return 0
        fi
        
        log_info "Waiting for services... ($((attempt + 1))/$max_attempts)"
        sleep 10
        ((attempt++))
    done

    log_error "Services failed to start within timeout"
    return 1
}

# Wait for Kubernetes services
wait_for_services_kubernetes() {
    log_info "Waiting for Kubernetes pods to be ready..."

    if ! kubectl wait --for=condition=ready pod -l app=sharehodl-node --timeout=600s -n sharehodl; then
        log_error "Pods failed to start within timeout"
        return 1
    fi

    log_success "Kubernetes services are ready"
}

# Setup monitoring
setup_monitoring() {
    if [[ "$MONITORING_ENABLED" != "true" ]]; then
        return 0
    fi

    log_info "Setting up monitoring..."

    local monitoring_dir="$PROJECT_ROOT/deployment/monitoring"
    mkdir -p "$monitoring_dir"

    # Generate Prometheus configuration
    generate_prometheus_config

    # Generate Grafana dashboards
    generate_grafana_dashboards

    # Generate Alertmanager configuration
    generate_alertmanager_config

    log_success "Monitoring setup completed"
}

# Generate Prometheus configuration
generate_prometheus_config() {
    cat > "$PROJECT_ROOT/deployment/monitoring/prometheus.yml" << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "/etc/prometheus/alerts.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'sharehodl-nodes'
    static_configs:
$(for i in $(seq 0 $((NODES - 1))); do
echo "      - targets: ['sharehodl-node-$i:26657']"
done)
    metrics_path: '/metrics'
    scrape_interval: 30s

  - job_name: 'sharehodl-api'
    static_configs:
$(for i in $(seq 0 $((NODES - 1))); do
echo "      - targets: ['sharehodl-node-$i:1317']"
done)
    metrics_path: '/metrics'
    scrape_interval: 30s
EOF
}

# Generate Grafana dashboards
generate_grafana_dashboards() {
    local grafana_dir="$PROJECT_ROOT/deployment/monitoring/grafana"
    mkdir -p "$grafana_dir/dashboards" "$grafana_dir/datasources"

    # Datasource configuration
    cat > "$grafana_dir/datasources/prometheus.yml" << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF

    # Dashboard provisioning
    cat > "$grafana_dir/dashboards/dashboard.yml" << EOF
apiVersion: 1

providers:
  - name: 'ShareHODL Dashboards'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF
}

# Generate Alertmanager configuration
generate_alertmanager_config() {
    cat > "$PROJECT_ROOT/deployment/monitoring/alertmanager.yml" << EOF
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@sharehodl.io'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
  - name: 'web.hook'
    webhook_configs:
      - url: 'http://localhost:5001/'
EOF
}

# Run security checks
run_security_checks() {
    if [[ "$SECURITY_SCANNING" != "true" ]]; then
        return 0
    fi

    log_info "Running security checks..."

    # Build security scanner
    if [[ -f "$PROJECT_ROOT/Dockerfile.security" ]]; then
        docker build -f "$PROJECT_ROOT/Dockerfile.security" -t sharehodl-security "$PROJECT_ROOT"
        
        # Run security scan
        docker run --rm -v "$PROJECT_ROOT:/app/source:ro" sharehodl-security
    fi

    log_success "Security checks completed"
}

# Setup backup system
setup_backup() {
    if [[ "$BACKUP_ENABLED" != "true" ]]; then
        return 0
    fi

    log_info "Setting up backup system..."

    # Generate backup script
    generate_backup_script

    log_success "Backup system setup completed"
}

# Generate backup script
generate_backup_script() {
    cat > "$PROJECT_ROOT/scripts/backup.sh" << 'EOF'
#!/bin/bash

set -euo pipefail

BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d-%H%M%S)
S3_BUCKET="${S3_BUCKET:-sharehodl-backups}"

log_info() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] INFO: $1"
}

create_backup() {
    local node_id="$1"
    local data_dir="/data/node-$node_id"
    local backup_file="$BACKUP_DIR/sharehodl-node-$node_id-$DATE.tar.gz"
    
    if [[ -d "$data_dir" ]]; then
        log_info "Creating backup for node $node_id"
        tar -czf "$backup_file" -C "$data_dir" .
        
        # Upload to S3 if configured
        if command -v aws &> /dev/null && [[ -n "$S3_BUCKET" ]]; then
            log_info "Uploading backup to S3: $S3_BUCKET"
            aws s3 cp "$backup_file" "s3://$S3_BUCKET/node-$node_id/"
        fi
        
        log_info "Backup completed: $backup_file"
    fi
}

# Create backups directory
mkdir -p "$BACKUP_DIR"

# Backup all nodes
for i in {0..2}; do
    create_backup "$i"
done

# Cleanup old backups (keep last 7 days)
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +7 -delete

log_info "Backup process completed"

# Sleep if running in daemon mode
if [[ "${BACKUP_INTERVAL:-0}" -gt 0 ]]; then
    sleep "$BACKUP_INTERVAL"
    exec "$0"
fi
EOF

    chmod +x "$PROJECT_ROOT/scripts/backup.sh"
}

# Display deployment summary
show_deployment_summary() {
    log_success "ShareHODL deployment completed successfully!"
    
    echo
    echo "=== Deployment Summary ==="
    echo "Environment: $ENVIRONMENT"
    echo "Network: $NETWORK"
    echo "Deployment Type: $DEPLOYMENT_TYPE"
    echo "Nodes: $NODES"
    echo "Validators: $VALIDATORS"
    echo "Monitoring: $MONITORING_ENABLED"
    echo "Security Scanning: $SECURITY_SCANNING"
    echo "Backup: $BACKUP_ENABLED"
    echo
    
    case "$DEPLOYMENT_TYPE" in
        "docker")
            echo "=== Service URLs ==="
            for i in $(seq 0 $((NODES - 1))); do
                echo "Node $i RPC: http://localhost:$((26657 + i))"
                echo "Node $i API: http://localhost:$((1317 + i))"
            done
            if [[ "$MONITORING_ENABLED" == "true" ]]; then
                echo "Prometheus: http://localhost:9090"
                echo "Grafana: http://localhost:3000 (admin/admin)"
            fi
            echo
            echo "=== Management Commands ==="
            echo "View logs: docker-compose -f $COMPOSE_FILE logs -f"
            echo "Stop services: docker-compose -f $COMPOSE_FILE down"
            echo "Restart services: docker-compose -f $COMPOSE_FILE restart"
            ;;
        "kubernetes")
            echo "=== Kubernetes Resources ==="
            echo "Namespace: sharehodl"
            echo "View pods: kubectl get pods -n sharehodl"
            echo "View services: kubectl get svc -n sharehodl"
            echo "View logs: kubectl logs -f deployment/sharehodl-node-0 -n sharehodl"
            ;;
        "terraform")
            echo "=== Terraform Outputs ==="
            cd "$TERRAFORM_DIR"
            terraform output
            ;;
    esac
    
    echo
    echo "=== Next Steps ==="
    echo "1. Verify all services are running properly"
    echo "2. Check blockchain synchronization status"
    echo "3. Test API endpoints and functionality"
    echo "4. Configure monitoring dashboards"
    echo "5. Set up backup schedules"
    echo "6. Review security configurations"
}

# Main deployment function
main() {
    log_info "Starting ShareHODL blockchain deployment..."
    
    parse_args "$@"
    check_prerequisites
    build_sharehodl
    generate_genesis
    setup_monitoring
    run_security_checks
    setup_backup
    
    case "$DEPLOYMENT_TYPE" in
        "docker")
            deploy_docker
            ;;
        "kubernetes")
            deploy_kubernetes
            ;;
        "terraform")
            deploy_terraform
            ;;
        *)
            log_error "Invalid deployment type: $DEPLOYMENT_TYPE"
            exit 1
            ;;
    esac
    
    show_deployment_summary
}

# Error handling
trap 'log_error "Deployment failed at line $LINENO"' ERR

# Run main function with all arguments
main "$@"