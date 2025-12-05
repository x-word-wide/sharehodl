# Atomic Swap API Reference

## Overview

This document provides a complete API reference for ShareHODL's atomic cross-asset swap functionality. The API enables instant, trustless exchanges between equity shares and HODL stablecoins with zero counterparty risk.

## Table of Contents

- [Message Types](#message-types)
- [Query Endpoints](#query-endpoints)  
- [Transaction Examples](#transaction-examples)
- [Response Formats](#response-formats)
- [Error Codes](#error-codes)
- [Event Types](#event-types)
- [SDK Integration](#sdk-integration)

## Message Types

### MsgPlaceAtomicSwapOrder

Execute instant cross-asset swaps with atomic settlement.

#### Request

```protobuf
service Msg {
  rpc PlaceAtomicSwapOrder(MsgPlaceAtomicSwapOrder) returns (MsgPlaceAtomicSwapOrderResponse);
}

message MsgPlaceAtomicSwapOrder {
  string trader = 1;                    // Trader's wallet address (bech32)
  string from_symbol = 2;               // Source asset symbol (HODL, APPLE, TSLA)
  string to_symbol = 3;                 // Target asset symbol 
  uint64 quantity = 4;                  // Amount to swap (in asset base units)
  string max_slippage = 5;              // Maximum acceptable slippage (0.00-0.50)
  TimeInForce time_in_force = 6;        // Order duration (GTC, IOC, FOK)
}
```

#### Response

```protobuf
message MsgPlaceAtomicSwapOrderResponse {
  string order_id = 1;                  // Unique order identifier
  string estimated_exchange_rate = 2;   // Final exchange rate used
}
```

#### CLI Usage

```bash
sharehodld tx dex atomic-swap \
  --from-symbol="HODL" \
  --to-symbol="APPLE" \
  --quantity=1000 \
  --max-slippage=0.05 \
  --time-in-force="GTC" \
  --from=trader1 \
  --gas=200000 \
  --chain-id=sharehodl-1
```

#### JSON-RPC Example

```bash
curl -X POST http://localhost:26657 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "broadcast_tx_commit",
    "params": {
      "tx": "CpwBCpkBCh0vc2hhcmVob2RsLmRleC52MS5Nc2dBdG9taWNTd2Fw..."
    },
    "id": 1
  }'
```

### MsgPlaceFractionalOrder

Trade fractional shares for micro-investing and precise portfolio allocation.

#### Request

```protobuf
message MsgPlaceFractionalOrder {
  string trader = 1;                    // Trader's wallet address
  string symbol = 2;                    // Asset symbol (APPLE, TSLA, etc.)
  OrderSide side = 3;                   // BUY or SELL
  string fractional_quantity = 4;       // Fractional shares (e.g., "0.5" for half share)
  string price = 5;                     // Limit price per share
  string max_price = 6;                 // Maximum acceptable price (buy orders)
  TimeInForce time_in_force = 7;        // Order duration
}
```

#### Response

```protobuf
message MsgPlaceFractionalOrderResponse {
  string order_id = 1;                  // Order identifier
  string effective_price = 2;           // Price achieved
}
```

#### CLI Usage

```bash
sharehodld tx dex fractional-order \
  --symbol="TSLA" \
  --side="buy" \
  --fractional-quantity=0.25 \
  --price=800 \
  --max-price=850 \
  --from=investor1 \
  --gas=150000
```

### MsgCreateTradingStrategy

Create programmable trading strategies with automated execution.

#### Request

```protobuf
message MsgCreateTradingStrategy {
  string owner = 1;                     // Strategy owner address
  string name = 2;                      // Human-readable strategy name
  repeated TriggerCondition conditions = 3; // Execution triggers
  repeated TradingAction actions = 4;   // Actions to execute
  string max_exposure = 5;              // Maximum capital at risk
}

message TriggerCondition {
  string symbol = 1;                    // Asset symbol to monitor
  string price_threshold = 2;           // Trigger price level
  string condition_type = 3;            // "ABOVE", "BELOW", "EQUALS"
  bool is_percentage = 4;               // Whether threshold is percentage-based
}

message TradingAction {
  OrderType action_type = 1;            // Order type (MARKET, LIMIT)
  string symbol = 2;                    // Asset to trade
  OrderSide side = 3;                   // BUY or SELL
  string quantity = 4;                  // Amount to trade
  string price_offset = 5;              // Price adjustment from market
}
```

#### Response

```protobuf
message MsgCreateTradingStrategyResponse {
  string strategy_id = 1;               // Unique strategy identifier
}
```

#### CLI Usage

```bash
sharehodld tx dex create-strategy \
  --name="Momentum Strategy" \
  --trigger-symbol="APPLE" \
  --trigger-threshold="1.05" \
  --trigger-type="ABOVE" \
  --action-type="market" \
  --action-symbol="APPLE" \
  --action-side="buy" \
  --action-quantity="100" \
  --max-exposure=50000 \
  --from=strategist1 \
  --gas=300000
```

## Query Endpoints

### Get Order Information

#### Request

```
GET /sharehodl/dex/v1/orders/{order_id}
```

#### Response

```json
{
  "order": {
    "id": "12345",
    "market_symbol": "APPLE/HODL",
    "base_symbol": "APPLE",
    "quote_symbol": "HODL",
    "user": "sharehodl1xyz...",
    "side": "BUY",
    "type": "ATOMIC_SWAP",
    "status": "FILLED",
    "quantity": "100",
    "filled_quantity": "100", 
    "price": "150.50",
    "average_price": "150.48",
    "created_at": "2023-12-04T10:30:00Z",
    "updated_at": "2023-12-04T10:30:02Z"
  }
}
```

#### CLI Usage

```bash
sharehodld query dex order 12345
```

### Get Market Data

#### Request

```
GET /sharehodl/dex/v1/markets/{base_symbol}/{quote_symbol}
```

#### Response

```json
{
  "market": {
    "market_symbol": "APPLE/HODL",
    "base_symbol": "APPLE",
    "quote_symbol": "HODL", 
    "last_price": "150.48",
    "volume_24h": "50000",
    "high_24h": "152.00",
    "low_24h": "148.50",
    "price_change_24h": "2.30",
    "active": true
  }
}
```

#### CLI Usage

```bash
sharehodl query dex market APPLE HODL
```

### Get Order Book

#### Request

```
GET /sharehodl/dex/v1/orderbook/{base_symbol}/{quote_symbol}
```

#### Response

```json
{
  "order_book": {
    "market_symbol": "APPLE/HODL",
    "bids": [
      {
        "price": "150.45",
        "quantity": "500",
        "order_count": 3
      }
    ],
    "asks": [
      {
        "price": "150.55", 
        "quantity": "300",
        "order_count": 2
      }
    ],
    "last_updated": "2023-12-04T10:30:00Z"
  }
}
```

### Get User Orders

#### Request

```
GET /sharehodl/dex/v1/orders?user={address}&status={status}&limit={limit}
```

#### Response

```json
{
  "orders": [
    {
      "id": "12345",
      "market_symbol": "APPLE/HODL",
      "status": "FILLED",
      "quantity": "100",
      "price": "150.50"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "1"
  }
}
```

### Get Trading Strategy

#### Request

```
GET /sharehodl/dex/v1/strategies/{strategy_id}
```

#### Response

```json
{
  "strategy": {
    "strategy_id": "strategy_12345",
    "owner": "sharehodl1abc...",
    "name": "Momentum Strategy",
    "conditions": [
      {
        "symbol": "APPLE",
        "price_threshold": "1.05",
        "condition_type": "ABOVE",
        "is_percentage": true
      }
    ],
    "actions": [
      {
        "action_type": "MARKET",
        "symbol": "APPLE",
        "side": "BUY", 
        "quantity": "100",
        "price_offset": "5.00"
      }
    ],
    "max_exposure": "50000",
    "active": true
  }
}
```

### Get Exchange Rate

#### Request

```
GET /sharehodl/dex/v1/rate/{from_symbol}/{to_symbol}
```

#### Response

```json
{
  "exchange_rate": {
    "from_symbol": "HODL",
    "to_symbol": "APPLE",
    "rate": "0.006644",
    "timestamp": "2023-12-04T10:30:00Z"
  }
}
```

#### CLI Usage

```bash
sharehodl query dex rate HODL APPLE
```

## Transaction Examples

### Complete Atomic Swap Flow

#### 1. Check Balance

```bash
sharehodld query bank balances sharehodl1xyz... 
```

#### 2. Get Current Exchange Rate

```bash
sharehodld query dex rate HODL APPLE
```

#### 3. Execute Swap

```bash
sharehodld tx dex atomic-swap \
  --from-symbol="HODL" \
  --to-symbol="APPLE" \
  --quantity=1000 \
  --max-slippage=0.03 \
  --from=trader1 \
  --yes
```

#### 4. Verify Execution

```bash
# Get order details
sharehodld query dex order [order-id]

# Check updated balance
sharehodld query bank balances sharehodl1xyz...
```

### Fractional Investment Strategy

#### 1. Dollar-Cost Averaging Setup

```bash
# Buy $100 worth of APPLE every week (0.67 shares at $150)
sharehodld tx dex fractional-order \
  --symbol="APPLE" \
  --side="buy" \
  --fractional-quantity=0.67 \
  --max-price=150 \
  --from=investor1
```

#### 2. Portfolio Rebalancing

```bash
# Sell partial position (1.5 shares)
sharehodld tx dex fractional-order \
  --symbol="TSLA" \
  --side="sell" \
  --fractional-quantity=1.5 \
  --price=800 \
  --from=investor1
```

### Automated Trading Strategy

#### 1. Momentum Strategy Creation

```bash
sharehodld tx dex create-strategy \
  --name="TSLA Momentum" \
  --trigger-symbol="TSLA" \
  --trigger-threshold="1.10" \
  --trigger-type="ABOVE" \
  --action-type="market" \
  --action-symbol="TSLA" \
  --action-side="buy" \
  --action-quantity="50" \
  --max-exposure=100000 \
  --from=trader1
```

#### 2. Strategy Monitoring

```bash
# Check strategy status
sharehodld query dex strategy [strategy-id]

# List all strategies for user
sharehodld query dex strategies sharehodl1abc...
```

## Response Formats

### Success Response

```json
{
  "code": 0,
  "data": "CgUIARIBMA==",
  "log": "",
  "info": "",
  "gas_wanted": "200000",
  "gas_used": "85432",
  "events": [
    {
      "type": "atomic_swap_executed",
      "attributes": [
        {
          "key": "trader",
          "value": "sharehodl1xyz..."
        },
        {
          "key": "from_symbol", 
          "value": "HODL"
        },
        {
          "key": "to_symbol",
          "value": "APPLE"
        },
        {
          "key": "order_id",
          "value": "12345"
        }
      ]
    }
  ],
  "codespace": ""
}
```

### Error Response

```json
{
  "code": 42,
  "data": "",
  "log": "slippage exceeded: actual=0.08, max=0.05: slippage exceeded", 
  "info": "",
  "gas_wanted": "200000",
  "gas_used": "35000",
  "events": [],
  "codespace": "dex"
}
```

## Error Codes

| Code | Error | Description |
|------|-------|-------------|
| 1 | market_not_found | Specified market does not exist |
| 14 | invalid_order_size | Order quantity is zero or negative |
| 20 | invalid_price | Price is zero, negative, or malformed |
| 25 | invalid_slippage | Slippage tolerance outside valid range (0-50%) |
| 30 | insufficient_funds | Account lacks sufficient balance |
| 42 | slippage_exceeded | Price movement exceeds tolerance |
| 70 | unauthorized | Invalid or missing trader address |
| 80 | asset_not_found | Asset symbol not supported |
| 81 | same_asset_swap | Cannot swap asset with itself |
| 82 | no_market_price | Price data unavailable for asset |
| 83 | insufficient_shares | Insufficient equity shares to swap |

### Error Handling Best Practices

```javascript
// Example error handling in JavaScript
try {
  const response = await client.tx.dex.atomicSwap({
    trader: userAddress,
    fromSymbol: "HODL", 
    toSymbol: "APPLE",
    quantity: 1000,
    maxSlippage: "0.05"
  });
  
  console.log(`Swap successful: ${response.orderID}`);
  
} catch (error) {
  switch (error.code) {
    case 42:
      console.error("Slippage exceeded - try increasing tolerance");
      break;
    case 30:
      console.error("Insufficient funds - check balance");
      break;
    case 82:
      console.error("Price unavailable - try again later");
      break;
    default:
      console.error(`Swap failed: ${error.message}`);
  }
}
```

## Event Types

### AtomicSwapExecuted

Emitted when an atomic swap completes successfully.

```json
{
  "type": "atomic_swap_executed",
  "attributes": [
    {"key": "trader", "value": "sharehodl1xyz..."},
    {"key": "from_symbol", "value": "HODL"},
    {"key": "to_symbol", "value": "APPLE"}, 
    {"key": "input_amount", "value": "1000"},
    {"key": "output_amount", "value": "6.644"},
    {"key": "exchange_rate", "value": "0.006644"},
    {"key": "slippage", "value": "0.024"},
    {"key": "order_id", "value": "12345"}
  ]
}
```

### FractionalOrderPlaced

Emitted when a fractional order is placed.

```json
{
  "type": "fractional_order_placed",
  "attributes": [
    {"key": "trader", "value": "sharehodl1abc..."},
    {"key": "symbol", "value": "TSLA"},
    {"key": "side", "value": "buy"},
    {"key": "fractional_quantity", "value": "0.25"},
    {"key": "order_id", "value": "67890"}
  ]
}
```

### TradingStrategyCreated

Emitted when a new trading strategy is created.

```json
{
  "type": "trading_strategy_created", 
  "attributes": [
    {"key": "owner", "value": "sharehodl1def..."},
    {"key": "strategy_id", "value": "strategy_12345"},
    {"key": "name", "value": "Momentum Strategy"},
    {"key": "max_exposure", "value": "50000"}
  ]
}
```

## SDK Integration

### Go SDK

```go
package main

import (
    "context"
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/sharehodl/sharehodl-blockchain/x/dex/types"
)

func executeAtomicSwap(ctx context.Context, client client.Context) error {
    msg := &types.MsgPlaceAtomicSwapOrder{
        Trader:      "sharehodl1xyz...",
        FromSymbol:  "HODL",
        ToSymbol:    "APPLE", 
        Quantity:    1000,
        MaxSlippage: sdk.MustNewDecFromStr("0.05"),
        TimeInForce: types.TimeInForceGTC,
    }
    
    // Build and broadcast transaction
    txBuilder := client.TxBuilder()
    txBuilder.SetMsgs(msg)
    txBuilder.SetGasLimit(200000)
    
    txBytes, err := client.TxConfig.TxEncoder()(txBuilder.GetTx())
    if err != nil {
        return err
    }
    
    res, err := client.BroadcastTxCommit(ctx, txBytes)
    if err != nil {
        return err
    }
    
    if res.Code != 0 {
        return fmt.Errorf("tx failed: %s", res.Log)
    }
    
    // Extract order ID from events
    for _, event := range res.Events {
        if event.Type == "atomic_swap_executed" {
            for _, attr := range event.Attributes {
                if attr.Key == "order_id" {
                    fmt.Printf("Order ID: %s\n", attr.Value)
                }
            }
        }
    }
    
    return nil
}
```

### JavaScript SDK

```javascript
import { SigningStargateClient } from "@cosmjs/stargate";
import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";

async function executeAtomicSwap() {
  // Create wallet and client
  const wallet = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic);
  const client = await SigningStargateClient.connectWithSigner(
    "http://localhost:26657", 
    wallet
  );
  
  const [firstAccount] = await wallet.getAccounts();
  
  // Prepare atomic swap message
  const msg = {
    typeUrl: "/sharehodl.dex.v1.MsgPlaceAtomicSwapOrder",
    value: {
      trader: firstAccount.address,
      fromSymbol: "HODL",
      toSymbol: "APPLE",
      quantity: "1000", 
      maxSlippage: "0.05",
      timeInForce: "GTC"
    }
  };
  
  // Execute transaction
  const fee = {
    amount: [{ denom: "hodl", amount: "5000" }],
    gas: "200000"
  };
  
  const result = await client.signAndBroadcast(
    firstAccount.address,
    [msg],
    fee,
    "Atomic swap: HODL -> APPLE"
  );
  
  if (result.code === 0) {
    console.log("Swap successful:", result.transactionHash);
    
    // Parse events for order ID
    result.events.forEach(event => {
      if (event.type === "atomic_swap_executed") {
        event.attributes.forEach(attr => {
          if (attr.key === "order_id") {
            console.log("Order ID:", attr.value);
          }
        });
      }
    });
  } else {
    console.error("Swap failed:", result.rawLog);
  }
}
```

### Python SDK

```python
import asyncio
from cosmpy.aerial.client import LedgerClient
from cosmpy.aerial.config import NetworkConfig
from cosmpy.aerial.wallet import LocalWallet

async def execute_atomic_swap():
    # Setup client and wallet
    network_config = NetworkConfig.fetchai_mainnet()
    client = LedgerClient(network_config)
    wallet = LocalWallet.from_mnemonic(mnemonic)
    
    # Prepare message
    msg = {
        "trader": str(wallet.address()),
        "from_symbol": "HODL",
        "to_symbol": "APPLE",
        "quantity": 1000,
        "max_slippage": "0.05",
        "time_in_force": "GTC"
    }
    
    # Execute transaction
    tx = await client.broadcast_tx_async(
        wallet=wallet,
        msgs=[msg],
        msg_type="/sharehodl.dex.v1.MsgPlaceAtomicSwapOrder",
        gas_limit=200000
    )
    
    if tx.code == 0:
        print(f"Swap successful: {tx.hash}")
        
        # Extract order ID from events
        for event in tx.events:
            if event.type == "atomic_swap_executed":
                for attr in event.attributes:
                    if attr.key == "order_id":
                        print(f"Order ID: {attr.value}")
    else:
        print(f"Swap failed: {tx.log}")

# Run the swap
asyncio.run(execute_atomic_swap())
```

## Rate Limiting

### Default Limits

- **Atomic Swaps**: 100 per minute per address
- **Fractional Orders**: 200 per minute per address  
- **Strategy Creation**: 10 per hour per address
- **Queries**: 1000 per minute per IP

### Headers

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1672531200
```

## Pagination

All list endpoints support pagination using cursor-based pagination.

### Request Parameters

- `limit`: Maximum number of items (default: 50, max: 500)
- `offset`: Starting position (default: 0)
- `order`: Sort order ("asc" or "desc", default: "desc")

### Response Format

```json
{
  "data": [...],
  "pagination": {
    "next_key": "eyJpZCI6MTIzNDV9",
    "total": "1500",
    "limit": "50",
    "offset": "100"
  }
}
```

---

**For more information:**
- [Main Documentation](./ATOMIC_SWAPS.md)
- [DEX Module README](./x/dex/README.md)
- [GitHub Repository](https://github.com/sharehodl/sharehodl-blockchain)