#!/bin/bash
set -e

CHAIN_ID="${CHAIN_ID:-sharehodl-local-1}"
NODE_NAME="${NODE_NAME:-sharehodl-node}"
SHAREHODL_HOME="${SHAREHODL_HOME:-/root/.sharehodl}"

# Module account addresses (derived from SHA256 of module name, with hodl prefix)
BONDED_POOL="hodl1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3yafwf9"

# Initialize node if not already done
if [ ! -f "$SHAREHODL_HOME/config/genesis.json" ]; then
    echo "Initializing ShareHODL node..."
    sharehodld init "$NODE_NAME" --chain-id "$CHAIN_ID" --home "$SHAREHODL_HOME"

    GENESIS_FILE="$SHAREHODL_HOME/config/genesis.json"

    # Create operator key and capture output (SDK keyring bug workaround)
    echo "Creating operator key..."
    KEY_OUTPUT=$(echo "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about" | \
        sharehodld keys add operator --keyring-backend test --recover --home "$SHAREHODL_HOME" --output json 2>&1 || true)

    # Parse address from JSON output
    OPERATOR_ADDR=$(echo "$KEY_OUTPUT" | jq -r '.address' 2>/dev/null || echo "")

    # If JSON parsing failed, try to extract from text output (hodl prefix)
    if [ -z "$OPERATOR_ADDR" ] || [ "$OPERATOR_ADDR" = "null" ]; then
        OPERATOR_ADDR=$(echo "$KEY_OUTPUT" | grep -oE 'hodl[a-z0-9]{38}' | head -1)
    fi

    # Hardcoded fallback - this is the address for the mnemonic used (with hodl prefix)
    if [ -z "$OPERATOR_ADDR" ]; then
        OPERATOR_ADDR="hodl19rl4cm2hmr8afy4kldpxz3fka4jguq0adxdu08"
    fi
    echo "Operator address: $OPERATOR_ADDR"

    # Derive valoper address using bech32 conversion (proper checksum)
    HEX_BYTES=$(sharehodld keys parse "$OPERATOR_ADDR" 2>/dev/null | grep "bytes:" | cut -d' ' -f2)
    if [ -n "$HEX_BYTES" ]; then
        VALOPER_ADDR=$(sharehodld keys parse "$HEX_BYTES" 2>/dev/null | grep "hodlvaloper" | grep -v "pub" | tr -d ' -')
    fi
    if [ -z "$VALOPER_ADDR" ]; then
        VALOPER_ADDR="hodlvaloper19rl4cm2hmr8afy4kldpxz3fka4jguq0akulp68"
    fi
    echo "Valoper address: $VALOPER_ADDR"

    # Create user key
    echo "Creating user key..."
    USER_OUTPUT=$(echo "test test test test test test test test test test test junk" | \
        sharehodld keys add user --keyring-backend test --recover --home "$SHAREHODL_HOME" --output json 2>&1 || true)

    USER_ADDR=$(echo "$USER_OUTPUT" | jq -r '.address' 2>/dev/null || echo "")
    if [ -z "$USER_ADDR" ] || [ "$USER_ADDR" = "null" ]; then
        USER_ADDR=$(echo "$USER_OUTPUT" | grep -oE 'hodl[a-z0-9]{38}' | head -1)
    fi
    if [ -z "$USER_ADDR" ]; then
        USER_ADDR="hodl15yk64u7zc9g9k2yr2wmzeva5qgwxps6yh50wfv"
    fi
    echo "User address: $USER_ADDR"

    # Get validator consensus pubkey from priv_validator_key.json
    echo "Getting validator pubkey..."
    PUBKEY_KEY=$(jq -r '.pub_key.value' "$SHAREHODL_HOME/config/priv_validator_key.json")
    echo "Validator pubkey: $PUBKEY_KEY"

    echo ""
    echo "Setting up genesis..."

    # Add user account
    echo "  Adding user account..."
    jq --arg addr "$USER_ADDR" '.app_state.auth.accounts += [{
        "@type": "/cosmos.auth.v1beta1.BaseAccount",
        "address": $addr,
        "pub_key": null,
        "account_number": "0",
        "sequence": "0"
    }]' "$GENESIS_FILE" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "$GENESIS_FILE"

    # Add operator account
    echo "  Adding operator account..."
    jq --arg addr "$OPERATOR_ADDR" '.app_state.auth.accounts += [{
        "@type": "/cosmos.auth.v1beta1.BaseAccount",
        "address": $addr,
        "pub_key": null,
        "account_number": "1",
        "sequence": "0"
    }]' "$GENESIS_FILE" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "$GENESIS_FILE"

    # Add bonded_tokens_pool module account (required for staking)
    echo "  Adding bonded_tokens_pool module account..."
    jq --arg addr "$BONDED_POOL" '.app_state.auth.accounts += [{
        "@type": "/cosmos.auth.v1beta1.ModuleAccount",
        "base_account": {
            "address": $addr,
            "pub_key": null,
            "account_number": "2",
            "sequence": "0"
        },
        "name": "bonded_tokens_pool",
        "permissions": ["burner", "staking"]
    }]' "$GENESIS_FILE" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "$GENESIS_FILE"

    # Add balances
    echo "  Adding balances..."
    jq --arg user_addr "$USER_ADDR" --arg op_addr "$OPERATOR_ADDR" --arg bonded_pool "$BONDED_POOL" '
        .app_state.bank.balances += [
            {
                "address": $user_addr,
                "coins": [
                    {"denom": "stake", "amount": "1000000000000"},
                    {"denom": "uhodl", "amount": "1000000000000000"}
                ]
            },
            {
                "address": $op_addr,
                "coins": [
                    {"denom": "stake", "amount": "50000000000000"},
                    {"denom": "uhodl", "amount": "100000000000000"}
                ]
            },
            {
                "address": $bonded_pool,
                "coins": [
                    {"denom": "stake", "amount": "50000000000000"}
                ]
            }
        ]
    ' "$GENESIS_FILE" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "$GENESIS_FILE"

    # Create validator
    echo "  Adding validator..."
    jq --arg valoper "$VALOPER_ADDR" \
       --arg pubkey "$PUBKEY_KEY" \
       --arg moniker "$NODE_NAME" \
       '.app_state.staking.validators += [{
            "operator_address": $valoper,
            "consensus_pubkey": {
                "@type": "/cosmos.crypto.ed25519.PubKey",
                "key": $pubkey
            },
            "jailed": false,
            "status": "BOND_STATUS_BONDED",
            "tokens": "50000000000000",
            "delegator_shares": "50000000000000.000000000000000000",
            "description": {
                "moniker": $moniker,
                "identity": "",
                "website": "",
                "security_contact": "",
                "details": "Genesis Validator"
            },
            "unbonding_height": "0",
            "unbonding_time": "1970-01-01T00:00:00Z",
            "commission": {
                "commission_rates": {
                    "rate": "0.100000000000000000",
                    "max_rate": "0.200000000000000000",
                    "max_change_rate": "0.010000000000000000"
                },
                "update_time": "2026-02-01T00:00:00Z"
            },
            "min_self_delegation": "1"
        }]' "$GENESIS_FILE" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "$GENESIS_FILE"

    # Add delegation
    echo "  Adding delegation..."
    jq --arg delegator "$OPERATOR_ADDR" --arg valoper "$VALOPER_ADDR" \
       '.app_state.staking.delegations += [{
            "delegator_address": $delegator,
            "validator_address": $valoper,
            "shares": "50000000000000.000000000000000000"
        }]' "$GENESIS_FILE" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "$GENESIS_FILE"

    # Set staking power
    echo "  Setting staking power..."
    jq --arg valoper "$VALOPER_ADDR" '
        .app_state.staking.last_total_power = "50000000" |
        .app_state.staking.last_validator_powers += [{
            "address": $valoper,
            "power": "50000000"
        }]
    ' "$GENESIS_FILE" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "$GENESIS_FILE"

    # Update bank supply
    echo "  Updating bank supply..."
    jq '.app_state.bank.supply = [
        {"denom": "stake", "amount": "101000000000000"},
        {"denom": "uhodl", "amount": "1100000000000000"}
    ]' "$GENESIS_FILE" > "${GENESIS_FILE}.tmp" && mv "${GENESIS_FILE}.tmp" "$GENESIS_FILE"

    # Configure node
    echo "Configuring node..."
    sed -i 's/127.0.0.1:26657/0.0.0.0:26657/g' "$SHAREHODL_HOME/config/config.toml"
    sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = ["*"]/g' "$SHAREHODL_HOME/config/config.toml"
    sed -i 's/enable = false/enable = true/g' "$SHAREHODL_HOME/config/app.toml"
    sed -i 's/swagger = false/swagger = true/g' "$SHAREHODL_HOME/config/app.toml"
    sed -i 's/localhost:1317/0.0.0.0:1317/g' "$SHAREHODL_HOME/config/app.toml"
    sed -i 's/enabled-unsafe-cors = false/enabled-unsafe-cors = true/g' "$SHAREHODL_HOME/config/app.toml"

    echo ""
    echo "=========================================="
    echo "Node initialization complete!"
    echo "=========================================="
    echo "User: $USER_ADDR"
    echo "  - 1,000,000,000 HODL (1B uhodl)"
    echo "  - 1,000,000 stake"
    echo ""
    echo "Validator: $OPERATOR_ADDR"
    echo "  - 50,000,000 stake bonded"
    echo "=========================================="
fi

echo "Starting ShareHODL node..."
exec sharehodld start --home "$SHAREHODL_HOME" "$@"
