# ShareHODL Blockchain API Reference

## Overview

ShareHODL provides multiple API interfaces for interacting with the blockchain:
- **REST API**: HTTP-based queries and transactions
- **gRPC API**: High-performance binary protocol
- **WebSocket**: Real-time event streaming
- **CLI**: Command-line interface for operations

## Base URLs

**Mainnet**:
- REST: `https://api.sharehodl.com`
- gRPC: `grpc.sharehodl.com:9090`
- WebSocket: `wss://ws.sharehodl.com`

**Testnet**:
- REST: `https://api-testnet.sharehodl.com`
- gRPC: `grpc-testnet.sharehodl.com:9090`  
- WebSocket: `wss://ws-testnet.sharehodl.com`

## Authentication

Most read operations are public. Write operations require transaction signatures:

```bash
# Generate transaction
sharehodld tx bank send alice bob 1000uhodl \
  --chain-id sharehodl-mainnet \
  --from alice \
  --gas auto \
  --gas-prices 0.025uhodl
```

## Core APIs

### 1. Account Management

#### Get Account Information
```http
GET /cosmos/auth/v1beta1/accounts/{address}
```

**Example Response**:
```json
{
  "account": {
    "@type": "/cosmos.auth.v1beta1.BaseAccount",
    "address": "sharehodl1abc123...",
    "pub_key": null,
    "account_number": "1234",
    "sequence": "5"
  }
}
```

#### Get Account Balances
```http
GET /cosmos/bank/v1beta1/balances/{address}
```

**Example Response**:
```json
{
  "balances": [
    {
      "denom": "uhodl",
      "amount": "1000000000"
    },
    {
      "denom": "equity/AAPL/common",
      "amount": "100"
    }
  ],
  "pagination": {
    "next_key": null,
    "total": "2"
  }
}
```

### 2. Transaction APIs

#### Send Transaction
```http
POST /cosmos/tx/v1beta1/txs
```

**Request Body**:
```json
{
  "tx_bytes": "base64EncodedTransaction",
  "mode": "BROADCAST_MODE_SYNC"
}
```

#### Get Transaction
```http
GET /cosmos/tx/v1beta1/txs/{hash}
```

**Example Response**:
```json
{
  "tx": {
    "body": {
      "messages": [
        {
          "@type": "/cosmos.bank.v1beta1.MsgSend",
          "from_address": "sharehodl1abc...",
          "to_address": "sharehodl1def...",
          "amount": [{"denom": "uhodl", "amount": "1000000"}]
        }
      ]
    }
  },
  "tx_response": {
    "height": "12345",
    "txhash": "ABC123...",
    "code": 0,
    "gas_used": "45000",
    "gas_wanted": "50000"
  }
}
```

### 3. Block APIs

#### Get Latest Block
```http
GET /cosmos/base/tendermint/v1beta1/blocks/latest
```

#### Get Block by Height
```http
GET /cosmos/base/tendermint/v1beta1/blocks/{height}
```

**Example Response**:
```json
{
  "block_id": {
    "hash": "ABC123...",
    "part_set_header": {
      "total": 1,
      "hash": "DEF456..."
    }
  },
  "block": {
    "header": {
      "height": "12345",
      "time": "2024-12-04T20:00:00Z",
      "chain_id": "sharehodl-mainnet"
    },
    "data": {
      "txs": ["base64EncodedTx1", "base64EncodedTx2"]
    }
  }
}
```

## HODL Module APIs

### Get HODL Parameters
```http
GET /sharehodl/hodl/v1/params
```

**Response**:
```json
{
  "params": {
    "minting_enabled": true,
    "collateral_ratio": "1.500000000000000000",
    "stability_fee": "0.000500000000000000",
    "liquidation_ratio": "1.300000000000000000",
    "mint_fee": "0.000100000000000000",
    "burn_fee": "0.000100000000000000"
  }
}
```

### Get Total HODL Supply
```http
GET /sharehodl/hodl/v1/total_supply
```

**Response**:
```json
{
  "total_supply": "500000000000"
}
```

### Get Collateral Position
```http
GET /sharehodl/hodl/v1/position/{address}
```

**Response**:
```json
{
  "position": {
    "owner": "sharehodl1abc123...",
    "collateral": [
      {"denom": "uatom", "amount": "1000000000"}
    ],
    "minted_hodl": "666666666",
    "stability_debt": "1000000.000000000000000000",
    "last_updated": "12345"
  }
}
```

## DEX Module APIs

### Get DEX Parameters
```http
GET /sharehodl/dex/v1/params
```

### List Markets
```http
GET /sharehodl/dex/v1/markets
```

**Response**:
```json
{
  "markets": [
    {
      "base_denom": "equity/AAPL/common",
      "quote_denom": "uhodl",
      "tick_size": "0.000001",
      "min_quantity": "1"
    }
  ]
}
```

### Get Order Book
```http
GET /sharehodl/dex/v1/orderbook/{base_denom}/{quote_denom}
```

### Get Orders by Address
```http
GET /sharehodl/dex/v1/orders/{address}
```

### Get Trades
```http
GET /sharehodl/dex/v1/trades?market={base_denom}/{quote_denom}&limit=100
```

## Equity Module APIs

### List Companies
```http
GET /sharehodl/equity/v1/companies
```

**Response**:
```json
{
  "companies": [
    {
      "id": "1",
      "name": "Apple Inc.",
      "symbol": "AAPL",
      "registration_country": "US",
      "business_type": "technology",
      "verified": true,
      "verifier_tier": "gold",
      "created_at": "12345"
    }
  ]
}
```

### Get Company Details
```http
GET /sharehodl/equity/v1/companies/{company_id}
```

### List Share Classes
```http
GET /sharehodl/equity/v1/share-classes/{company_id}
```

**Response**:
```json
{
  "share_classes": [
    {
      "id": "common",
      "company_id": "1",
      "name": "Common Stock",
      "symbol": "AAPL",
      "total_shares": "1000000",
      "issued_shares": "800000",
      "share_price": "150.00",
      "currency": "USD"
    }
  ]
}
```

### Get Shareholdings
```http
GET /sharehodl/equity/v1/shareholdings/{address}
```

## Validator Module APIs

### Get Validator Tiers
```http
GET /sharehodl/validator/v1/tiers
```

### Get Validator Tier Info
```http
GET /sharehodl/validator/v1/validator/{validator_address}
```

**Response**:
```json
{
  "validator_info": {
    "validator_address": "sharehodlvaloper1abc...",
    "tier": "gold",
    "stake_amount": "1000000000000",
    "reward_multiplier": "1.500000000000000000",
    "business_verifications": 25,
    "reputation_score": "4.8"
  }
}
```

## Transaction Types

### 1. HODL Transactions

#### Mint HODL
```bash
sharehodld tx hodl mint 1000000uhodl \
  --from alice \
  --collateral 1500000uatom \
  --chain-id sharehodl-mainnet
```

#### Burn HODL  
```bash
sharehodld tx hodl burn 500000uhodl \
  --from alice \
  --chain-id sharehodl-mainnet
```

### 2. DEX Transactions

#### Create Order
```bash
sharehodld tx dex create-order \
  --base-denom "equity/AAPL/common" \
  --quote-denom "uhodl" \
  --side "buy" \
  --quantity "100" \
  --price "150000000" \
  --from alice
```

#### Cancel Order
```bash
sharehodld tx dex cancel-order \
  --order-id "12345" \
  --from alice
```

### 3. Equity Transactions

#### Register Company
```bash
sharehodld tx equity register-company \
  --name "Apple Inc." \
  --symbol "AAPL" \
  --country "US" \
  --business-type "technology" \
  --from company-admin
```

#### Issue Shares
```bash
sharehodld tx equity issue-shares \
  --company-id "1" \
  --share-class "common" \
  --amount "1000000" \
  --recipient "sharehodl1abc..." \
  --from company-admin
```

## WebSocket Events

### Connection
```javascript
const ws = new WebSocket('wss://ws.sharehodl.com');

ws.onopen = () => {
  // Subscribe to events
  ws.send(JSON.stringify({
    method: 'subscribe',
    params: ['tm.event=NewBlock']
  }));
};
```

### Available Subscriptions

#### New Blocks
```json
{"method": "subscribe", "params": ["tm.event='NewBlock'"]}
```

#### New Transactions  
```json
{"method": "subscribe", "params": ["tm.event='Tx'"]}
```

#### DEX Events
```json
{"method": "subscribe", "params": ["dex.order_matched"]}
```

#### Equity Events
```json
{"method": "subscribe", "params": ["equity.shares_issued"]}
```

**Example Event**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "query": "tm.event='NewBlock'",
    "data": {
      "type": "tendermint/event/NewBlock",
      "value": {
        "block": {
          "header": {
            "height": "12346",
            "time": "2024-12-04T20:01:00Z"
          }
        }
      }
    }
  }
}
```

## Error Codes

| Code | Message | Description |
|------|---------|-------------|
| 0 | Success | Transaction executed successfully |
| 2 | Invalid Request | Malformed request |
| 3 | Insufficient Funds | Account balance too low |
| 4 | Unauthorized | Invalid signature or permissions |
| 5 | Not Found | Resource not found |
| 11 | Out of Gas | Transaction ran out of gas |
| 12 | Memo Too Large | Transaction memo exceeds limit |
| 13 | Insufficient Fee | Fee too low for transaction |
| 19 | Tx Too Large | Transaction size exceeds limit |

## Rate Limits

**Public Endpoints**:
- 100 requests per minute per IP
- Burst limit: 200 requests

**Authenticated Endpoints**:
- 1000 requests per minute per account
- Burst limit: 2000 requests

## SDKs and Libraries

### JavaScript/TypeScript
```bash
npm install @sharehodl/sharehodl-js
```

```typescript
import { ShareHODLClient } from '@sharehodl/sharehodl-js';

const client = new ShareHODLClient('https://api.sharehodl.com');
const balance = await client.getBalance('sharehodl1abc...');
```

### Go
```bash
go get github.com/sharehodl/sharehodl-go-sdk
```

```go
import "github.com/sharehodl/sharehodl-go-sdk/client"

client := client.NewHTTPClient("https://api.sharehodl.com")
balance, err := client.GetBalance("sharehodl1abc...")
```

### Python
```bash
pip install sharehodl-python
```

```python
from sharehodl import Client

client = Client("https://api.sharehodl.com")
balance = client.get_balance("sharehodl1abc...")
```

## Testing

Use testnet for development:

```bash
# Configure for testnet
sharehodld config chain-id sharehodl-testnet
sharehodld config node https://rpc-testnet.sharehodl.com:443

# Get test tokens from faucet
curl -X POST https://faucet-testnet.sharehodl.com/claim \
  -H "Content-Type: application/json" \
  -d '{"address": "sharehodl1abc..."}'
```

---

This API reference provides comprehensive documentation for integrating with the ShareHODL blockchain.