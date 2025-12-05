# 📡 ShareHODL Blockchain - API Documentation

## API Endpoints Overview

### Base URLs
- **RPC**: `http://localhost:26657` (Tendermint RPC)
- **REST**: `http://localhost:1317` (Cosmos SDK REST API)  
- **gRPC**: `localhost:9090` (gRPC services)
- **WebSocket**: `ws://localhost:26657/websocket` (Real-time events)

## 1. Explorer API (Port 3001)

### Block Information
```typescript
// Get latest blocks
GET /cosmos/base/tendermint/v1beta1/blocks/latest
Response: {
  block_id: { hash: string, part_set_header: {...} },
  block: {
    header: {
      version: {...},
      chain_id: string,
      height: string,
      time: string,
      last_block_id: {...},
      validator_address: string
    },
    data: { txs: string[] },
    evidence: {...},
    last_commit: {...}
  }
}

// Get specific block by height
GET /cosmos/base/tendermint/v1beta1/blocks/{height}

// Get block transactions
GET /cosmos/tx/v1beta1/txs?events=tx.height={height}
```

### Transaction Information
```typescript
// Get transaction by hash
GET /cosmos/tx/v1beta1/txs/{hash}
Response: {
  tx: {
    body: {
      messages: [...],
      memo: string,
      timeout_height: string,
      extension_options: [...],
      non_critical_extension_options: [...]
    },
    auth_info: {
      signer_infos: [...],
      fee: {
        amount: [{ denom: string, amount: string }],
        gas_limit: string,
        payer: string,
        granter: string
      }
    },
    signatures: string[]
  },
  tx_response: {
    height: string,
    txhash: string,
    codespace: string,
    code: number,
    data: string,
    raw_log: string,
    logs: [...],
    info: string,
    gas_wanted: string,
    gas_used: string,
    tx: {...},
    timestamp: string
  }
}

// Search transactions
GET /cosmos/tx/v1beta1/txs?events=message.sender='{address}'
GET /cosmos/tx/v1beta1/txs?events=message.action='/cosmos.bank.v1beta1.MsgSend'
```

### Network Status
```typescript
// Get node status
GET /status (RPC endpoint)
Response: {
  jsonrpc: "2.0",
  id: -1,
  result: {
    node_info: {
      protocol_version: {...},
      id: string,
      listen_addr: string,
      network: string,
      version: string,
      channels: string,
      moniker: string,
      other: {...}
    },
    sync_info: {
      latest_block_hash: string,
      latest_app_hash: string,
      latest_block_height: string,
      latest_block_time: string,
      earliest_block_hash: string,
      earliest_app_hash: string,
      earliest_block_height: string,
      earliest_block_time: string,
      catching_up: boolean
    },
    validator_info: {
      address: string,
      pub_key: {...},
      voting_power: string
    }
  }
}
```

## 2. Trading API (Port 3002)

### Balance Queries
```typescript
// Get account balances
GET /cosmos/bank/v1beta1/balances/{address}
Response: {
  balances: [
    { denom: string, amount: string }
  ],
  pagination: {
    next_key: string,
    total: string
  }
}

// Get specific denomination balance
GET /cosmos/bank/v1beta1/balances/{address}/by_denom?denom={denom}
```

### DEX Operations
```typescript
// Execute atomic swap
POST /sharehodl.dex.v1.Msg/AtomicSwap
Body: {
  creator: string,
  from_denom: string,
  to_denom: string,
  amount: string,
  min_receive: string,
  timeout: string,
  slippage_tolerance: string
}

// Get market data
GET /sharehodl/dex/v1/markets
Response: {
  markets: [
    {
      base_denom: string,
      quote_denom: string,
      last_price: string,
      volume_24h: string,
      price_change_24h: string,
      high_24h: string,
      low_24h: string
    }
  ]
}

// Get order book
GET /sharehodl/dex/v1/orderbook/{market}
Response: {
  market: string,
  bids: [
    { price: string, quantity: string, total: string }
  ],
  asks: [
    { price: string, quantity: string, total: string }
  ],
  last_update: string
}
```

### HODL Stablecoin Operations
```typescript
// Mint HODL
POST /sharehodl.hodl.v1.Msg/MintHODL
Body: {
  creator: string,
  collateral_coins: [{ denom: string, amount: string }],
  hodl_amount: string
}

// Burn HODL
POST /sharehodl.hodl.v1.Msg/BurnHODL
Body: {
  creator: string,
  hodl_amount: string,
  collateral_return: [{ denom: string, amount: string }]
}

// Get collateral position
GET /sharehodl/hodl/v1/position/{address}
Response: {
  position: {
    owner: string,
    collateral: [{ denom: string, amount: string }],
    minted_hodl: string,
    collateral_ratio: string,
    liquidation_price: string,
    last_updated: string
  }
}
```

## 3. Governance API (Port 3003)

### Proposal Management
```typescript
// Get all proposals
GET /cosmos/gov/v1beta1/proposals
Response: {
  proposals: [
    {
      proposal_id: string,
      content: {
        type: string,
        value: {
          title: string,
          description: string
        }
      },
      status: string,
      final_tally_result: {
        yes: string,
        abstain: string,
        no: string,
        no_with_veto: string
      },
      submit_time: string,
      deposit_end_time: string,
      total_deposit: [{ denom: string, amount: string }],
      voting_start_time: string,
      voting_end_time: string
    }
  ],
  pagination: {...}
}

// Submit proposal
POST /cosmos.gov.v1beta1.Msg/SubmitProposal
Body: {
  content: {
    type: "/cosmos.gov.v1beta1.TextProposal",
    value: {
      title: string,
      description: string
    }
  },
  initial_deposit: [{ denom: string, amount: string }],
  proposer: string
}

// Vote on proposal
POST /cosmos.gov.v1beta1.Msg/Vote
Body: {
  proposal_id: string,
  voter: string,
  option: "VOTE_OPTION_YES" | "VOTE_OPTION_NO" | "VOTE_OPTION_ABSTAIN" | "VOTE_OPTION_NO_WITH_VETO"
}

// Get proposal votes
GET /cosmos/gov/v1beta1/proposals/{proposal_id}/votes
```

### Governance Parameters
```typescript
// Get governance parameters
GET /cosmos/gov/v1beta1/params/{param_type}
// param_type: "voting" | "tallying" | "deposit"

Response: {
  voting_params: {
    voting_period: string
  },
  deposit_params: {
    min_deposit: [{ denom: string, amount: string }],
    max_deposit_period: string
  },
  tally_params: {
    quorum: string,
    threshold: string,
    veto_threshold: string
  }
}
```

## 4. Wallet API (Port 3004)

### Account Management
```typescript
// Get account information
GET /cosmos/auth/v1beta1/accounts/{address}
Response: {
  account: {
    type: "/cosmos.auth.v1beta1.BaseAccount",
    value: {
      address: string,
      pub_key: {...},
      account_number: string,
      sequence: string
    }
  }
}

// Get transaction history
GET /cosmos/tx/v1beta1/txs?events=message.sender='{address}'&pagination.limit=50
GET /cosmos/tx/v1beta1/txs?events=transfer.recipient='{address}'&pagination.limit=50
```

### Transaction Broadcasting
```typescript
// Send transaction
POST /cosmos/tx/v1beta1/txs
Body: {
  tx_bytes: string, // base64 encoded signed transaction
  mode: "BROADCAST_MODE_SYNC" | "BROADCAST_MODE_ASYNC" | "BROADCAST_MODE_BLOCK"
}

Response: {
  tx_response: {
    height: string,
    txhash: string,
    codespace: string,
    code: number,
    data: string,
    raw_log: string,
    logs: [...],
    info: string,
    gas_wanted: string,
    gas_used: string,
    timestamp: string
  }
}

// Simulate transaction (estimate gas)
POST /cosmos/tx/v1beta1/simulate
Body: {
  tx_bytes: string // base64 encoded unsigned transaction
}

Response: {
  gas_info: {
    gas_wanted: string,
    gas_used: string
  },
  result: {
    data: string,
    log: string,
    events: [...]
  }
}
```

## 5. Business API (Port 3005)

### Company Verification
```typescript
// Register company
POST /sharehodl.equity.v1.Msg/RegisterCompany
Body: {
  owner: string,
  name: string,
  description: string,
  business_type: string,
  jurisdiction: string,
  registration_number: string,
  verification_documents: string[]
}

// Get company information
GET /sharehodl/equity/v1/company/{address}
Response: {
  company: {
    address: string,
    owner: string,
    name: string,
    description: string,
    business_type: string,
    jurisdiction: string,
    registration_number: string,
    verification_status: "PENDING" | "VERIFIED" | "REJECTED",
    verification_date: string,
    share_classes: [...]
  }
}

// Issue shares
POST /sharehodl.equity.v1.Msg/IssueShares
Body: {
  issuer: string,
  company_address: string,
  share_class: string,
  amount: string,
  recipient: string
}
```

### Validator Management
```typescript
// Get validator tier information
GET /sharehodl/validator/v1/tier/{validator_address}
Response: {
  tier: {
    validator_address: string,
    tier_level: "BRONZE" | "SILVER" | "GOLD" | "PLATINUM" | "DIAMOND",
    stake_amount: string,
    reward_multiplier: string,
    tier_since: string,
    next_tier_requirement: string
  }
}

// Update validator tier
POST /sharehodl.validator.v1.Msg/UpdateTier
Body: {
  validator_address: string,
  new_stake_amount: string
}
```

## 6. WebSocket Events (Real-time)

### Connection
```typescript
const ws = new WebSocket('ws://localhost:26657/websocket');

// Subscribe to events
ws.send(JSON.stringify({
  jsonrpc: "2.0",
  method: "subscribe",
  id: "1",
  params: {
    query: "tm.event='NewBlock'" // or other event types
  }
}));
```

### Event Types
```typescript
// New blocks
subscribe: "tm.event='NewBlock'"
response: {
  jsonrpc: "2.0",
  id: "1",
  result: {
    query: "tm.event='NewBlock'",
    data: {
      type: "tendermint/event/NewBlock",
      value: {
        block: {...},
        result_begin_block: {...},
        result_end_block: {...}
      }
    },
    events: {...}
  }
}

// New transactions
subscribe: "tm.event='Tx'"
subscribe: "message.action='/cosmos.bank.v1beta1.MsgSend'"
subscribe: "message.action='/sharehodl.dex.v1.MsgAtomicSwap'"

// Custom events
subscribe: "atomic_swap.from_denom='uhodl'"
subscribe: "company_registered.name='Apple Inc'"
```

## Frontend Integration Examples

### React Hook for Balance Query
```typescript
import { useQuery } from '@tanstack/react-query';

export function useBalance(address: string, denom?: string) {
  return useQuery({
    queryKey: ['balance', address, denom],
    queryFn: async () => {
      const url = denom 
        ? `/cosmos/bank/v1beta1/balances/${address}/by_denom?denom=${denom}`
        : `/cosmos/bank/v1beta1/balances/${address}`;
      
      const response = await fetch(`http://localhost:1317${url}`);
      return response.json();
    },
    enabled: !!address,
    refetchInterval: 5000 // Refresh every 5 seconds
  });
}
```

### WebSocket Hook for Real-time Data
```typescript
import { useEffect, useState } from 'react';

export function useBlockchainEvents() {
  const [latestBlock, setLatestBlock] = useState(null);
  const [newTransactions, setNewTransactions] = useState([]);

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:26657/websocket');
    
    ws.onopen = () => {
      // Subscribe to new blocks
      ws.send(JSON.stringify({
        jsonrpc: "2.0",
        method: "subscribe",
        id: "blocks",
        params: { query: "tm.event='NewBlock'" }
      }));
      
      // Subscribe to new transactions
      ws.send(JSON.stringify({
        jsonrpc: "2.0",
        method: "subscribe",
        id: "txs",
        params: { query: "tm.event='Tx'" }
      }));
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      if (data.id === "blocks") {
        setLatestBlock(data.result.data.value.block);
      } else if (data.id === "txs") {
        setNewTransactions(prev => [data.result.data.value.TxResult, ...prev.slice(0, 9)]);
      }
    };

    return () => ws.close();
  }, []);

  return { latestBlock, newTransactions };
}
```

### Transaction Signing
```typescript
import { SigningStargateClient } from '@cosmjs/stargate';

export async function executeAtomicSwap(
  signer: any,
  fromDenom: string,
  toDenom: string,
  amount: string
) {
  const client = await SigningStargateClient.connectWithSigner(
    "http://localhost:26657",
    signer
  );

  const msg = {
    typeUrl: "/sharehodl.dex.v1.MsgAtomicSwap",
    value: {
      creator: await signer.getFirstAddress(),
      fromDenom,
      toDenom,
      amount,
      minReceive: "0", // Calculate based on slippage
      timeout: (Date.now() + 3600000).toString(), // 1 hour
      slippageTolerance: "0.03" // 3%
    }
  };

  const fee = {
    amount: [{ denom: "uhodl", amount: "5000" }],
    gas: "200000"
  };

  return await client.signAndBroadcast(
    await signer.getFirstAddress(),
    [msg],
    fee,
    "Atomic swap via ShareDEX"
  );
}
```

## Error Handling

### Common HTTP Status Codes
```typescript
200: Success
400: Bad Request - Invalid parameters
404: Not Found - Resource doesn't exist
500: Internal Server Error - Node error
503: Service Unavailable - Node not synced

// Error response format
{
  code: number,
  message: string,
  details: string[]
}
```

### Transaction Error Codes
```typescript
// Cosmos SDK error codes
2: ErrTxDecode
3: ErrInvalidSequence
4: ErrUnauthorized
5: ErrInsufficientFunds
6: ErrUnknownRequest
7: ErrInvalidAddress
8: ErrInvalidPubKey
9: ErrUnknownAddress
10: ErrInvalidCoins
11: ErrOutOfGas

// ShareHODL custom error codes
1001: ErrInsufficientCollateral (HODL module)
1002: ErrSlippageExceeded (DEX module)
1003: ErrInvalidTier (Validator module)
1004: ErrCompanyNotVerified (Equity module)
```

## Rate Limiting & Best Practices

### API Rate Limits
- **REST API**: 100 requests/minute per IP
- **RPC API**: 1000 requests/minute per IP  
- **WebSocket**: 10 concurrent connections per IP

### Optimization Tips
```typescript
// Use pagination for large datasets
GET /cosmos/tx/v1beta1/txs?pagination.limit=20&pagination.offset=40

// Cache frequently accessed data
const blockCache = new Map();

// Batch multiple queries
const [balances, delegations, proposals] = await Promise.all([
  fetch('/cosmos/bank/v1beta1/balances/' + address),
  fetch('/cosmos/staking/v1beta1/delegations/' + address),
  fetch('/cosmos/gov/v1beta1/proposals')
]);

// Use WebSocket for real-time updates instead of polling
```

---

*API Documentation Version: 1.0*  
*Last Updated: December 2024*  
*Compatible with: Cosmos SDK v0.54, ShareHODL v1.0*