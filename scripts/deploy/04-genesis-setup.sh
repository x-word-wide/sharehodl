#!/bin/bash

# ShareHODL Genesis Setup Script
# Sets up initial accounts and validator for new chain

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Configuration
DEPLOY_DIR="/opt/sharehodl/sharehodl-blockchain"
SHAREHODL_HOME="$HOME/.sharehodl"
SHAREHODLD="$DEPLOY_DIR/build/sharehodld"
CHAIN_ID="sharehodl-mainnet"
KEYRING="file"

echo "=============================================="
echo "  ShareHODL Genesis Setup"
echo "  Chain ID: $CHAIN_ID"
echo "=============================================="
echo ""

# Verify binary exists
if [[ ! -f "$SHAREHODLD" ]]; then
    log_error "sharehodld binary not found at $SHAREHODLD"
fi

# Verify chain is initialized
if [[ ! -d "$SHAREHODL_HOME/config" ]]; then
    log_error "Chain not initialized. Run 02-deploy-sharehodl.sh first."
fi

# Step 1: Create test accounts
log_info "Step 1: Creating test accounts..."

# Create accounts with mnemonics for testing
accounts=("alice" "bob" "charlie" "diana" "treasury")

for account in "${accounts[@]}"; do
    if $SHAREHODLD keys show "$account" --keyring-backend $KEYRING --home "$SHAREHODL_HOME" &>/dev/null; then
        log_warning "Account '$account' already exists"
    else
        log_info "Creating account: $account"
        $SHAREHODLD keys add "$account" --keyring-backend $KEYRING --home "$SHAREHODL_HOME"
    fi
done

# Step 2: Get validator address
log_info "Step 2: Getting validator address..."
VALIDATOR_ADDR=$($SHAREHODLD keys show validator -a --keyring-backend $KEYRING --home "$SHAREHODL_HOME")
log_info "Validator address: $VALIDATOR_ADDR"

# Step 3: Add genesis accounts
log_info "Step 3: Adding genesis accounts..."

# Add validator account with tokens
$SHAREHODLD genesis add-genesis-account "$VALIDATOR_ADDR" \
    100000000000000uhodl,100000000000000stake \
    --keyring-backend $KEYRING \
    --home "$SHAREHODL_HOME"

# Add test accounts with tokens
for account in "${accounts[@]}"; do
    ADDR=$($SHAREHODLD keys show "$account" -a --keyring-backend $KEYRING --home "$SHAREHODL_HOME")
    log_info "Adding genesis account: $account ($ADDR)"
    $SHAREHODLD genesis add-genesis-account "$ADDR" \
        10000000000000uhodl,10000000000000stake \
        --keyring-backend $KEYRING \
        --home "$SHAREHODL_HOME"
done

log_success "Genesis accounts added"

# Step 4: Create validator gentx
log_info "Step 4: Creating validator gentx..."
$SHAREHODLD genesis gentx validator 10000000000000stake \
    --keyring-backend $KEYRING \
    --chain-id "$CHAIN_ID" \
    --commission-rate 0.05 \
    --commission-max-rate 0.20 \
    --commission-max-change-rate 0.01 \
    --min-self-delegation 1 \
    --moniker "ShareHODL-Validator-1" \
    --details "ShareHODL Mainnet Validator" \
    --home "$SHAREHODL_HOME"

log_success "Validator gentx created"

# Step 5: Collect gentxs
log_info "Step 5: Collecting genesis transactions..."
$SHAREHODLD genesis collect-gentxs --home "$SHAREHODL_HOME"
log_success "Genesis transactions collected"

# Step 6: Validate genesis
log_info "Step 6: Validating genesis..."
$SHAREHODLD genesis validate --home "$SHAREHODL_HOME"
log_success "Genesis validated"

# Step 7: Display account addresses
echo ""
echo "=============================================="
echo "  GENESIS SETUP COMPLETE"
echo "=============================================="
echo ""
echo "Account addresses:"
echo ""
printf "%-12s %-50s %s\n" "Account" "Address" "Balance"
echo "--------------------------------------------------------------------------------"
echo "validator    $VALIDATOR_ADDR  100T HODL, 100T stake"
for account in "${accounts[@]}"; do
    ADDR=$($SHAREHODLD keys show "$account" -a --keyring-backend $KEYRING --home "$SHAREHODL_HOME")
    printf "%-12s %-50s %s\n" "$account" "$ADDR" "10T HODL, 10T stake"
done
echo ""
echo "Chain is ready to start!"
echo ""
echo "Start the blockchain:"
echo "  sudo systemctl start sharehodld"
echo ""
echo "Monitor logs:"
echo "  journalctl -u sharehodld -f"
echo ""
