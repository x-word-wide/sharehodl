#!/bin/bash

set -e

# ShareHODL blockchain initialization script

BINARY="sharehodld"
CHAIN_DIR="$HOME/.sharehodl"
CHAINID="sharehodl-1"
KEYRING="test"
KEY="validator"
MONIKER="sharehodl-validator"

echo "Building sharehodld binary..."
make build

echo "Initializing chain..."
$BINARY config keyring-backend $KEYRING
$BINARY config chain-id $CHAINID
$BINARY init $MONIKER --chain-id $CHAINID

echo "Adding validator key..."
if ! $BINARY keys show $KEY 2>/dev/null; then
    $BINARY keys add $KEY --keyring-backend $KEYRING
fi

echo "Setting up genesis..."
$BINARY add-genesis-account $KEY 100000000000000hodl,100000000stake --keyring-backend $KEYRING
$BINARY gentx $KEY 1000000stake --chain-id $CHAINID --keyring-backend $KEYRING

$BINARY collect-gentxs

echo "Validating genesis..."
$BINARY validate-genesis

echo "ShareHODL chain initialization completed!"
echo "Chain ID: $CHAINID"
echo "Validator key: $KEY"
echo "To start the chain, run: make start"