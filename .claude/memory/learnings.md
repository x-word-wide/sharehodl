# Project Learnings

> This file captures important learnings about the ShareHODL codebase. Update when discovering patterns, gotchas, or insights.

## Codebase Patterns

### Cosmos SDK Module Pattern
```
x/{module}/
├── keeper/           # Business logic
│   ├── keeper.go    # Core keeper with dependencies
│   ├── msg_server.go # Message handlers
│   └── query.go     # Query handlers
├── types/           # Data structures
│   ├── msgs.go      # Message definitions
│   ├── keys.go      # Store keys
│   └── errors.go    # Module errors
├── module.go        # Module registration
└── genesis.go       # Genesis handling
```

### Frontend App Pattern
```
apps/{app}/
├── app/             # Next.js App Router
│   ├── layout.tsx   # Root layout
│   ├── page.tsx     # Home page
│   └── [routes]/    # Route segments
├── components/      # App-specific components
└── lib/             # Utilities and hooks
```

## Important Gotchas

### HODL Module
- Collateral ratio is stored as basis points (15000 = 150%)
- Liquidation threshold is 130%, not 150%
- Stability fees are annual, calculated per-block
- Always check collateral ratio AFTER price updates

### DEX Module
- Order IDs are deterministic from tx hash
- Partial fills create new orders for remainder
- Slippage is max 50% (5000 basis points)
- Order matching happens in EndBlock

### Equity Module
- Share precision is 3 decimal places (0.001)
- Cap table changes require company admin signature
- Dividends must be in HODL tokens only
- Stock splits are processed in BeginBlock

### Validator Module
- Tier thresholds are in uhodl (micro HODL)
- Tier upgrades are immediate
- Tier downgrades have 7-day delay
- Slashing resets tier to Bronze

## Common Mistakes to Avoid

### Go Code
1. **Forgetting ValidateBasic()**: Always validate in both message and keeper
2. **Not emitting events**: Indexer relies on events for all state changes
3. **Panicking in handlers**: Use error returns, never panic
4. **Missing authorization**: Check msg.GetSigners() against required authority

### TypeScript Code
1. **Not handling loading states**: Always show skeleton/spinner
2. **Ignoring errors**: Always have error boundary and error UI
3. **Over-fetching**: Use specific queries, not getAll
4. **Missing types**: Never use `any`, define proper interfaces

### Testing
1. **Testing implementation not behavior**: Test what, not how
2. **Missing edge cases**: Empty, nil, max, min values
3. **No cleanup between tests**: State leaks cause flaky tests
4. **Hardcoded addresses**: Use generated or fixture addresses

## Performance Insights

### Blockchain
- KV store reads are expensive - batch when possible
- Event emission has overhead - emit only necessary data
- JSON marshaling is slow - use protobuf
- State iteration is O(n) - use indexes

### Frontend
- Bundle size matters - lazy load routes
- API calls are slow - cache aggressively
- Re-renders are expensive - memoize properly
- Images are heavy - use next/image

## Security Notes

### Known Attack Vectors
1. **Front-running**: Order timing manipulation
2. **Flash loans**: Not currently supported, but watch for
3. **Oracle manipulation**: Price feeds are trusted
4. **Governance attacks**: 33.4% quorum required

### Security-Critical Code Paths
- `x/hodl/keeper/msg_server.go` - Mint/burn operations
- `x/dex/keeper/matching.go` - Order matching engine
- `x/equity/keeper/transfers.go` - Share transfers
- `x/validator/keeper/slashing.go` - Slashing logic

## Debugging Tips

### Blockchain
- Check logs with: `docker logs sharehodl-node -f`
- Query state with: `sharehodld query {module} {query}`
- Debug txs with: `sharehodld query tx {hash}`

### Frontend
- Check network tab for API failures
- Use React DevTools for component state
- Check console for TypeScript errors
- Use debug mode: `DEBUG=* pnpm dev`

---
*Last updated: 2026-01-28*
