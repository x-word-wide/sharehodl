# DEX Module

The DEX (Decentralized Exchange) module enables blockchain-native trading features on the ShareHODL blockchain, focusing on revolutionary capabilities that surpass traditional markets.

## Quick Start

### Installation

```bash
# Build the ShareHODL blockchain
make build

# Initialize a new chain
sharehodld init mynode --chain-id sharehodl-1

# Start the blockchain  
sharehodld start
```

### Basic Usage

```bash
# Place an atomic swap order
sharehodld tx dex atomic-swap \
  --from-symbol="HODL" \
  --to-symbol="APPLE" \
  --quantity=1000 \
  --max-slippage=0.05 \
  --from=alice

# Place a fractional order  
sharehodld tx dex fractional-order \
  --symbol="TSLA" \
  --side="buy" \
  --fractional-quantity=0.5 \
  --max-price=800 \
  --from=bob

# Create a trading strategy
sharehodld tx dex create-strategy \
  --name="Momentum" \
  --conditions="price_change_24h > 0.05" \
  --actions="buy APPLE 100 shares" \
  --max-exposure=10000 \
  --from=charlie
```

## Features

### 🚀 Blockchain-Native Advantages

- **Instant Settlement**: Execute trades in seconds, not T+2 days
- **24/7 Trading**: Global market access without exchange hours
- **Zero Counterparty Risk**: Atomic execution eliminates default risk
- **Fractional Trading**: Buy 0.1 shares for precise allocation
- **Programmable Strategies**: Automated execution with smart contracts

### ⚡ Core Functionality

1. **Atomic Cross-Asset Swaps**
   - HODL ↔ Equity swaps (APPLE, TSLA, etc.)
   - Cross-equity swaps (APPLE ↔ TSLA)
   - Real-time exchange rates
   - Configurable slippage protection (0-50%)

2. **Fractional Share Trading**
   - Sub-share precision (0.001 shares)
   - Micro-investing capabilities  
   - Dollar-cost averaging support
   - Precise portfolio rebalancing

3. **Programmable Trading Strategies**
   - Condition-based triggers
   - Automated order execution
   - Risk management controls
   - Multi-asset strategies

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    DEX Module                               │
├─────────────────────────────────────────────────────────────┤
│  Keeper Layer                                               │
│  ├── Atomic Swap Logic          ├── Order Management       │
│  ├── Exchange Rate Calculation  ├── Trade Execution        │
│  ├── Slippage Protection       ├── Strategy Automation     │
│  └── Asset Transfer Logic      └── Event Management        │
├─────────────────────────────────────────────────────────────┤
│  Message Layer                                              │
│  ├── MsgPlaceAtomicSwapOrder   ├── MsgPlaceFractionalOrder │
│  ├── MsgCreateTradingStrategy  ├── MsgExecuteStrategy      │
│  └── Input Validation         └── Error Handling          │
├─────────────────────────────────────────────────────────────┤
│  Storage Layer                                              │
│  ├── Order Books               ├── Trade History           │
│  ├── Market Data              ├── Strategy Storage         │
│  └── Counter Management       └── Asset Balances          │
└─────────────────────────────────────────────────────────────┘
```

## File Structure

```
x/dex/
├── keeper/
│   ├── keeper.go          # Core DEX logic and atomic swaps
│   ├── msg_server.go      # Message handlers for transactions
│   ├── keeper_test.go     # Keeper unit tests
│   └── msg_server_test.go # Message handler tests
├── types/
│   ├── dex.go            # Core types (Order, Trade, Market)
│   ├── msg.go            # Message types and validation
│   ├── errors.go         # Error definitions
│   ├── key.go            # Storage keys
│   ├── service.go        # gRPC service interface
│   └── atomic_swap_test.go # Type validation tests
├── module.go             # Module definition
├── integration_test.go   # End-to-end tests  
└── README.md            # This file
```

## Key Types

### Order

```go
type Order struct {
    ID                uint64         `json:"id"`
    MarketSymbol      string         `json:"market_symbol"`
    User              string         `json:"user"`
    Side              OrderSide      `json:"side"`        // Buy/Sell
    Type              OrderType      `json:"type"`        // AtomicSwap/Fractional/Programmatic
    Status            OrderStatus    `json:"status"`      // Pending/Filled/Cancelled
    Quantity          math.Int       `json:"quantity"`
    Price             math.LegacyDec `json:"price"`
    CreatedAt         time.Time      `json:"created_at"`
}
```

### Trade

```go
type Trade struct {
    ID           uint64         `json:"id"`
    MarketSymbol string         `json:"market_symbol"`
    Buyer        string         `json:"buyer"`
    Seller       string         `json:"seller"`
    Quantity     math.Int       `json:"quantity"`
    Price        math.LegacyDec `json:"price"`
    ExecutedAt   time.Time      `json:"executed_at"`
}
```

## Message Types

### Atomic Swap Order

Execute instant cross-asset swaps with zero counterparty risk.

```go
type MsgPlaceAtomicSwapOrder struct {
    Trader       string         `json:"trader"`
    FromSymbol   string         `json:"from_symbol"`   // Source asset
    ToSymbol     string         `json:"to_symbol"`     // Target asset
    Quantity     uint64         `json:"quantity"`      // Amount to swap
    MaxSlippage  math.LegacyDec `json:"max_slippage"`  // Price protection
    TimeInForce  TimeInForce    `json:"time_in_force"`
}
```

### Fractional Order

Trade partial shares for micro-investing and precise allocation.

```go
type MsgPlaceFractionalOrder struct {
    Trader             string         `json:"trader"`
    Symbol             string         `json:"symbol"`
    Side               OrderSide      `json:"side"`
    FractionalQuantity math.LegacyDec `json:"fractional_quantity"`
    Price              math.LegacyDec `json:"price"`
    MaxPrice           math.LegacyDec `json:"max_price"`
    TimeInForce        TimeInForce    `json:"time_in_force"`
}
```

### Trading Strategy

Create programmable strategies with automated execution.

```go
type MsgCreateTradingStrategy struct {
    Owner       string            `json:"owner"`
    Name        string            `json:"name"`
    Conditions  []TriggerCondition `json:"conditions"`
    Actions     []TradingAction   `json:"actions"`
    MaxExposure math.Int          `json:"max_exposure"`
}
```

## Testing

### Running Tests

```bash
# Run all DEX tests
go test ./x/dex/... -v

# Run specific test suites
go test ./x/dex/types/ -v      # Type validation tests
go test ./x/dex/keeper/ -v     # Keeper logic tests  
go test ./x/dex/ -v            # Integration tests

# Run with coverage
go test ./x/dex/... -v -cover
```

### Test Coverage

- **Message Validation**: 15+ test cases
- **Atomic Swap Logic**: 8 scenarios including edge cases
- **Slippage Protection**: 4 price protection tests
- **Error Handling**: 10+ failure modes
- **Integration**: End-to-end workflows

## Examples

### Basic Atomic Swap

```bash
# Swap 1000 HODL for APPLE shares
sharehodld tx dex atomic-swap \
  --from-symbol="HODL" \
  --to-symbol="APPLE" \
  --quantity=1000 \
  --max-slippage=0.03 \
  --gas=200000 \
  --from=trader1
```

### Fractional Investment

```bash
# Buy 0.25 TSLA shares
sharehodld tx dex fractional-order \
  --symbol="TSLA" \
  --side="buy" \
  --fractional-quantity=0.25 \
  --max-price=850 \
  --gas=150000 \
  --from=investor1
```

### Momentum Strategy

```bash
# Create automated momentum trading
sharehodld tx dex create-strategy \
  --name="APPLE Momentum" \
  --trigger-symbol="APPLE" \
  --trigger-threshold="1.05" \
  --trigger-type="ABOVE" \
  --action-type="market" \
  --action-symbol="APPLE" \
  --action-side="buy" \
  --action-quantity="100" \
  --max-exposure=50000 \
  --gas=300000 \
  --from=strategist1
```

## Configuration

### Module Parameters

```yaml
# app.toml
[dex]
max_slippage_tolerance = "0.50"        # 50% maximum
min_fractional_quantity = "0.001"      # 0.001 shares minimum  
max_strategy_triggers = 10             # Max triggers per strategy
fee_rate = "0.001"                    # 0.1% trading fee
```

### Genesis Configuration

```json
{
  "dex": {
    "params": {
      "max_slippage_tolerance": "0.50",
      "min_fractional_quantity": "0.001",
      "trading_fee_rate": "0.001"
    },
    "markets": [],
    "orders": [],
    "trades": []
  }
}
```

## Security

### Slippage Protection

- Maximum slippage capped at 50% to prevent extreme losses
- Real-time price validation at execution
- User-configurable tolerance levels

### Asset Safety

- Atomic operations prevent partial failures
- Balance verification before execution
- Automatic rollback on any error

### Access Control

- Only asset owners can initiate swaps
- Strategy creators maintain full control
- Multi-signature support for large trades

## Performance

### Benchmarks

| Operation | Execution Time | Gas Usage |
|-----------|----------------|-----------|
| Atomic Swap | < 100ms | ~80,000 |
| Fractional Order | < 50ms | ~60,000 |
| Strategy Creation | < 200ms | ~120,000 |

### Optimization

- Efficient storage with minimal state bloat
- Lazy loading of order book data
- Batch processing for multiple operations
- Event indexing for fast queries

## Troubleshooting

### Common Issues

1. **Slippage Exceeded**
   - Increase `max_slippage` parameter
   - Check market volatility
   - Use smaller order sizes

2. **Insufficient Funds**
   - Verify account balances
   - Account for trading fees
   - Check asset allowances

3. **Invalid Asset**
   - Ensure assets are supported
   - Check symbol spelling
   - Verify market exists

### Debug Commands

```bash
# Check account balance
sharehodld query bank balances [address]

# View order status
sharehodld query dex order [order-id]

# Check market data  
sharehodld query dex market [base-symbol] [quote-symbol]

# List trading strategies
sharehodld query dex strategies [owner]
```

## Contributing

### Development Setup

```bash
# Clone repository
git clone https://github.com/sharehodl/sharehodl-blockchain.git
cd sharehodl-blockchain

# Install dependencies
go mod download

# Run tests
make test

# Build binary
make build
```

### Code Standards

- Follow Go best practices
- Include comprehensive tests
- Document public functions
- Use meaningful variable names
- Handle errors gracefully

## Resources

- [Atomic Swaps Documentation](../ATOMIC_SWAPS.md)
- [ShareHODL Whitepaper](../docs/whitepaper.md)
- [API Reference](../docs/api.md)
- [GitHub Repository](https://github.com/sharehodl/sharehodl-blockchain)

---

**Need Help?**
- Join our [Discord](https://discord.gg/sharehodl)
- Visit [Docs](https://docs.sharehodl.com)
- Open an [Issue](https://github.com/sharehodl/sharehodl-blockchain/issues)