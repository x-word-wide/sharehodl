# Quick Test Guide - Keplr Integration

## Quick Verification

### 1. Build and Start (30 seconds)
```bash
# Build
go build -o ./build/sharehodld ./cmd/sharehodld

# Start node (in one terminal)
./build/sharehodld start

# Keep running for tests below
```

### 2. Test CLI Commands (30 seconds)
```bash
# In another terminal:

# Test tx commands
./build/sharehodld tx --help
./build/sharehodld tx bank --help

# Test query commands
./build/sharehodld query --help

# Expected: All commands show help without errors
```

### 3. Test REST API Account Query (10 seconds)
```bash
# Replace with an actual address from your genesis
curl http://localhost:1317/cosmos/auth/v1beta1/accounts/hodl1...

# Expected response (NOT an error):
# {
#   "account": {
#     "@type": "/cosmos.auth.v1beta1.BaseAccount",
#     "address": "hodl1...",
#     "pub_key": null,
#     "account_number": "0",
#     "sequence": "0"
#   }
# }

# NOT this error (which was the bug):
# {
#   "error": "no registered implementations of type types.AccountI"
# }
```

### 4. Test Keplr Integration (2 minutes)

#### Step 1: Add Chain
In browser console:
```javascript
await window.keplr.experimentalSuggestChain({
  chainId: "sharehodl-1",
  chainName: "ShareHODL",
  rpc: "http://localhost:26657",
  rest: "http://localhost:1317",
  bip44: { coinType: 118 },
  bech32Config: {
    bech32PrefixAccAddr: "hodl",
    bech32PrefixAccPub: "hodlpub",
    bech32PrefixValAddr: "hodlvaloper",
    bech32PrefixValPub: "hodlvaloperpub",
    bech32PrefixConsAddr: "hodlvalcons",
    bech32PrefixConsPub: "hodlvalconspub"
  },
  currencies: [{
    coinDenom: "HODL",
    coinMinimalDenom: "uhodl",
    coinDecimals: 6,
    coinGeckoId: "sharehodl"
  }],
  feeCurrencies: [{
    coinDenom: "HODL",
    coinMinimalDenom: "uhodl",
    coinDecimals: 6,
    coinGeckoId: "sharehodl",
    gasPriceStep: {
      low: 0.0025,
      average: 0.025,
      high: 0.04
    }
  }],
  stakeCurrency: {
    coinDenom: "HODL",
    coinMinimalDenom: "uhodl",
    coinDecimals: 6,
    coinGeckoId: "sharehodl"
  }
});
```

#### Step 2: Enable Chain
```javascript
await window.keplr.enable("sharehodl-1");
```

#### Step 3: Get Account Info
```javascript
const key = await window.keplr.getKey("sharehodl-1");
console.log("Address:", key.bech32Address);
console.log("Public Key:", key.pubKey);

// Expected: Address and public key displayed
// NOT: Error about account not found
```

#### Step 4: Get Account from REST
```javascript
const address = key.bech32Address;
const response = await fetch(`http://localhost:1317/cosmos/auth/v1beta1/accounts/${address}`);
const data = await response.json();
console.log(data);

// Expected: Account object with @type and address
// NOT: Error object
```

### 5. Test Transaction Signing (Optional)

```javascript
// Get offline signer
const offlineSigner = window.keplr.getOfflineSigner("sharehodl-1");
const accounts = await offlineSigner.getAccounts();
console.log("Accounts:", accounts);

// Create a simple transaction (you'll need SigningStargateClient)
// If this step works without errors, Keplr integration is complete!
```

## Success Indicators

✅ **All tests pass if:**
1. CLI commands show help without panics
2. `sharehodld tx bank --help` works
3. REST account query returns JSON (not error)
4. Keplr can add the chain
5. Keplr can enable the chain
6. Keplr can get account information

## Common Issues

### Issue: "command not found: sharehodld"
**Solution**: Use `./build/sharehodld` or add build/ to PATH

### Issue: "connection refused" on REST queries
**Solution**: Ensure node is running (`./build/sharehodld start`)

### Issue: Keplr says "chain not found"
**Solution**: Run the experimentalSuggestChain code first

### Issue: Account query returns empty
**Solution**: Make sure you're using an address that exists in genesis

## Get Test Address

```bash
# List keys
./build/sharehodld keys list --keyring-backend test

# Or get from genesis
cat genesis.json | grep -A 1 "alice"

# Use one of these addresses for testing
```

## Full E2E Test

```bash
# 1. Clean start
rm -rf ~/.sharehodl
./build/sharehodld init test --chain-id sharehodl-1

# 2. Add test key
./build/sharehodld keys add alice --keyring-backend test

# 3. Get address
ALICE_ADDR=$(./build/sharehodld keys show alice -a --keyring-backend test)
echo "Alice address: $ALICE_ADDR"

# 4. Add genesis account
./build/sharehodld genesis add-genesis-account $ALICE_ADDR 1000000000uhodl

# 5. Start node
./build/sharehodld start

# 6. Test account query (in another terminal)
curl http://localhost:1317/cosmos/auth/v1beta1/accounts/$ALICE_ADDR

# Expected: Account object, NOT error
```

## Quick Checklist

- [ ] Build succeeds
- [ ] Node starts
- [ ] CLI tx --help works
- [ ] CLI query --help works
- [ ] REST account query returns JSON
- [ ] Keplr can add chain
- [ ] Keplr can get account
- [ ] No "no registered implementations" error

## Time Estimate

- Build: 2 minutes
- CLI tests: 1 minute
- REST tests: 1 minute
- Keplr tests: 3 minutes
- **Total: ~7 minutes**

## Need Help?

See full documentation:
- `KEPLR_INTEGRATION_FIX_PLAN.md` - Problem analysis
- `KEPLR_INTEGRATION_IMPLEMENTATION.md` - Implementation details
- `KEPLR_INTEGRATION_COMPLETE.md` - Complete summary
