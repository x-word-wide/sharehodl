#!/bin/bash

# ShareHODL Node Health Check Script
# Performs comprehensive health checks for ShareHODL blockchain nodes

set -euo pipefail

# Configuration
SHAREHODL_HOME="${SHAREHODL_HOME:-/home/sharehodl/.sharehodl}"
NODE_ID="${NODE_ID:-0}"
RPC_PORT=$((26657 + NODE_ID))
API_PORT=$((1317 + NODE_ID))
TIMEOUT=10

# Health check functions
check_process() {
    if ! pgrep -f "sharehodld" >/dev/null; then
        echo "ERROR: ShareHODL process not running"
        exit 1
    fi
}

check_rpc_port() {
    if ! timeout $TIMEOUT curl -f -s "http://localhost:$RPC_PORT/status" >/dev/null; then
        echo "ERROR: RPC port $RPC_PORT not responding"
        exit 1
    fi
}

check_api_port() {
    if ! timeout $TIMEOUT curl -f -s "http://localhost:$API_PORT/node_info" >/dev/null; then
        echo "ERROR: API port $API_PORT not responding"
        exit 1
    fi
}

check_sync_status() {
    local status_response
    if ! status_response=$(timeout $TIMEOUT curl -f -s "http://localhost:$RPC_PORT/status"); then
        echo "ERROR: Failed to get sync status"
        exit 1
    fi
    
    local catching_up
    if ! catching_up=$(echo "$status_response" | jq -r '.result.sync_info.catching_up'); then
        echo "WARNING: Could not parse sync status"
        return 0
    fi
    
    # Allow some time for initial sync
    if [[ "$catching_up" == "true" ]]; then
        echo "INFO: Node is syncing"
    fi
}

check_peer_connections() {
    local net_info
    if ! net_info=$(timeout $TIMEOUT curl -f -s "http://localhost:$RPC_PORT/net_info"); then
        echo "WARNING: Could not get network info"
        return 0
    fi
    
    local peer_count
    if peer_count=$(echo "$net_info" | jq -r '.result.n_peers'); then
        if [[ "$peer_count" -eq 0 ]]; then
            echo "WARNING: No peers connected"
        else
            echo "INFO: Connected to $peer_count peers"
        fi
    fi
}

check_block_production() {
    local status_response
    if ! status_response=$(timeout $TIMEOUT curl -f -s "http://localhost:$RPC_PORT/status"); then
        echo "WARNING: Could not check block production"
        return 0
    fi
    
    local latest_block_height
    if latest_block_height=$(echo "$status_response" | jq -r '.result.sync_info.latest_block_height'); then
        if [[ "$latest_block_height" -eq 0 ]]; then
            echo "WARNING: No blocks produced yet"
        else
            echo "INFO: Latest block height: $latest_block_height"
        fi
    fi
}

check_disk_space() {
    local disk_usage
    if disk_usage=$(df "$SHAREHODL_HOME" | tail -1 | awk '{print $5}' | sed 's/%//'); then
        if [[ "$disk_usage" -gt 90 ]]; then
            echo "ERROR: Disk usage is $disk_usage%, critically high"
            exit 1
        elif [[ "$disk_usage" -gt 80 ]]; then
            echo "WARNING: Disk usage is $disk_usage%"
        fi
    fi
}

check_memory_usage() {
    local memory_usage
    if memory_usage=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}'); then
        if [[ "$memory_usage" -gt 95 ]]; then
            echo "ERROR: Memory usage is $memory_usage%, critically high"
            exit 1
        elif [[ "$memory_usage" -gt 85 ]]; then
            echo "WARNING: Memory usage is $memory_usage%"
        fi
    fi
}

# Run all health checks
main() {
    echo "INFO: Starting health check for ShareHODL node $NODE_ID"
    
    check_process
    check_rpc_port
    check_api_port
    check_sync_status
    check_peer_connections
    check_block_production
    check_disk_space
    check_memory_usage
    
    echo "SUCCESS: Health check passed"
    exit 0
}

# Run health check
main