package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

// TestRateLimitEnforcement tests rate limit enforcement
func (suite *KeeperTestSuite) TestRateLimitEnforcement() {
	// Set up chain and asset
	chain := types.NewExternalChain(
		"ethereum-1",
		"Ethereum",
		"evm",
		true,
		12,
		12,
		"0x1234567890abcdef",
		math.NewInt(100_000),
		math.NewInt(10_000_000_000),
	)
	suite.keeper.SetExternalChain(suite.ctx, chain)

	asset := types.NewExternalAsset(
		"ethereum-1",
		"USDT",
		"Tether USD",
		"0xdAC17F958D2ee523a2206206994597C13D831ec7",
		6,
		true,
		math.LegacyOneDec(),
		math.NewInt(1_000_000_000_000),
		math.NewInt(100_000_000_000),
	)
	suite.keeper.SetExternalAsset(suite.ctx, asset)

	// Set rate limit parameters
	params := suite.keeper.GetParams(suite.ctx)
	params.MaxWithdrawalPerWindow = math.NewInt(20_000_000) // 20 HODL max
	params.RateLimitWindow = 3600                            // 1 hour window
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(100_000_000)))

	// First withdrawal: 15 HODL (should succeed)
	amount1 := math.NewInt(15_000_000)
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDT", amount1)
	suite.Require().NoError(err)

	suite.keeper.UpdateRateLimit(suite.ctx, "ethereum-1", "USDT", amount1)

	// Second withdrawal: 10 HODL (should fail - would exceed 20 HODL limit)
	amount2 := math.NewInt(10_000_000)
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDT", amount2)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrRateLimitExceeded, err)

	// Third withdrawal: 5 HODL (should succeed - total 20 HODL)
	amount3 := math.NewInt(5_000_000)
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDT", amount3)
	suite.Require().NoError(err)

	suite.keeper.UpdateRateLimit(suite.ctx, "ethereum-1", "USDT", amount3)

	// Fourth withdrawal: 1 HODL (should fail - at limit)
	amount4 := math.NewInt(1_000_000)
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDT", amount4)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrRateLimitExceeded, err)

	// Verify rate limit state
	rateLimit := suite.keeper.GetRateLimit(suite.ctx, "ethereum-1", "USDT", suite.ctx.BlockTime())
	suite.Require().True(rateLimit.AmountUsed.Equal(math.NewInt(20_000_000)))
	suite.Require().Equal(uint64(2), rateLimit.TxCount)
	suite.Require().True(rateLimit.RemainingCapacity().IsZero())
}

// TestRateLimitWindowExpiration tests rate limit window reset
func (suite *KeeperTestSuite) TestRateLimitWindowExpiration() {
	// Set up chain and asset
	chain := types.NewExternalChain(
		"ethereum-1",
		"Ethereum",
		"evm",
		true,
		12,
		12,
		"0x1234567890abcdef",
		math.NewInt(100_000),
		math.NewInt(10_000_000_000),
	)
	suite.keeper.SetExternalChain(suite.ctx, chain)

	asset := types.NewExternalAsset(
		"ethereum-1",
		"USDT",
		"Tether USD",
		"0xdAC17F958D2ee523a2206206994597C13D831ec7",
		6,
		true,
		math.LegacyOneDec(),
		math.NewInt(1_000_000_000_000),
		math.NewInt(100_000_000_000),
	)
	suite.keeper.SetExternalAsset(suite.ctx, asset)

	// Set rate limit parameters
	params := suite.keeper.GetParams(suite.ctx)
	params.MaxWithdrawalPerWindow = math.NewInt(10_000_000) // 10 HODL max
	params.RateLimitWindow = 3600                            // 1 hour window
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Use up the entire rate limit
	amount := math.NewInt(10_000_000)
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDT", amount)
	suite.Require().NoError(err)

	suite.keeper.UpdateRateLimit(suite.ctx, "ethereum-1", "USDT", amount)

	// Verify limit reached
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDT", math.NewInt(1))
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrRateLimitExceeded, err)

	// Get rate limit before expiration
	rateLimit := suite.keeper.GetRateLimit(suite.ctx, "ethereum-1", "USDT", suite.ctx.BlockTime())
	suite.Require().False(rateLimit.IsExpired(suite.ctx.BlockTime()))
	suite.Require().Equal(math.NewInt(10_000_000), rateLimit.AmountUsed)

	// Move time forward past the window
	oldTime := suite.ctx.BlockTime()
	futureTime := oldTime.Add(2 * time.Hour)
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	// Old rate limit should be expired when checked at future time
	suite.Require().True(rateLimit.IsExpired(futureTime))

	// Getting rate limit at new time creates a fresh window
	newRateLimit := suite.keeper.GetRateLimit(suite.ctx, "ethereum-1", "USDT", futureTime)
	suite.Require().True(newRateLimit.AmountUsed.IsZero()) // New window, no usage

	// New withdrawal should succeed (new window)
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDT", amount)
	suite.Require().NoError(err)
}

// TestRateLimitMultipleAssets tests independent rate limits per asset
func (suite *KeeperTestSuite) TestRateLimitMultipleAssets() {
	// Set up chain with multiple assets
	chain := types.NewExternalChain(
		"ethereum-1",
		"Ethereum",
		"evm",
		true,
		12,
		12,
		"0x1234567890abcdef",
		math.NewInt(100_000),
		math.NewInt(10_000_000_000),
	)
	suite.keeper.SetExternalChain(suite.ctx, chain)

	usdt := types.NewExternalAsset(
		"ethereum-1",
		"USDT",
		"Tether USD",
		"0xdAC17F958D2ee523a2206206994597C13D831ec7",
		6,
		true,
		math.LegacyOneDec(),
		math.NewInt(1_000_000_000_000),
		math.NewInt(100_000_000_000),
	)
	suite.keeper.SetExternalAsset(suite.ctx, usdt)

	usdc := types.NewExternalAsset(
		"ethereum-1",
		"USDC",
		"USD Coin",
		"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
		6,
		true,
		math.LegacyOneDec(),
		math.NewInt(1_000_000_000_000),
		math.NewInt(100_000_000_000),
	)
	suite.keeper.SetExternalAsset(suite.ctx, usdc)

	// Set rate limit
	params := suite.keeper.GetParams(suite.ctx)
	params.MaxWithdrawalPerWindow = math.NewInt(10_000_000) // 10 HODL per asset
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Max out USDT rate limit
	amount := math.NewInt(10_000_000)
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDT", amount)
	suite.Require().NoError(err)
	suite.keeper.UpdateRateLimit(suite.ctx, "ethereum-1", "USDT", amount)

	// USDT should be at limit
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDT", math.NewInt(1))
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrRateLimitExceeded, err)

	// USDC should still have full capacity (independent limit)
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDC", amount)
	suite.Require().NoError(err)
}

// TestCircuitBreakerOperations tests circuit breaker functionality
func (suite *KeeperTestSuite) TestCircuitBreakerOperations() {
	// Initially no circuit breaker
	cb := suite.keeper.GetCircuitBreaker(suite.ctx)
	suite.Require().False(cb.Enabled)

	// All operations should be allowed
	suite.Require().NoError(suite.keeper.IsOperationAllowed(suite.ctx, "deposit"))
	suite.Require().NoError(suite.keeper.IsOperationAllowed(suite.ctx, "withdraw"))
	suite.Require().NoError(suite.keeper.IsOperationAllowed(suite.ctx, "attest"))

	// Enable circuit breaker blocking all operations
	duration := 1 * time.Hour
	cb = types.NewCircuitBreaker(
		true,
		"Emergency pause - suspected exploit",
		"cosmos1authority",
		false, // No deposits
		false, // No withdrawals
		false, // No attestations
		&duration,
	)
	suite.keeper.SetCircuitBreaker(suite.ctx, cb)

	// Verify all operations blocked
	err := suite.keeper.IsOperationAllowed(suite.ctx, "deposit")
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrCircuitBreakerActive, err)

	err = suite.keeper.IsOperationAllowed(suite.ctx, "withdraw")
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrCircuitBreakerActive, err)

	err = suite.keeper.IsOperationAllowed(suite.ctx, "attest")
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrCircuitBreakerActive, err)

	// Enable selective operations (deposits only)
	cb.CanDeposit = true
	suite.keeper.SetCircuitBreaker(suite.ctx, cb)

	// Deposits allowed, others blocked
	suite.Require().NoError(suite.keeper.IsOperationAllowed(suite.ctx, "deposit"))
	suite.Require().Error(suite.keeper.IsOperationAllowed(suite.ctx, "withdraw"))
	suite.Require().Error(suite.keeper.IsOperationAllowed(suite.ctx, "attest"))

	// Manually disable circuit breaker
	cb.Enabled = false
	suite.keeper.SetCircuitBreaker(suite.ctx, cb)

	// All operations allowed again
	suite.Require().NoError(suite.keeper.IsOperationAllowed(suite.ctx, "deposit"))
	suite.Require().NoError(suite.keeper.IsOperationAllowed(suite.ctx, "withdraw"))
	suite.Require().NoError(suite.keeper.IsOperationAllowed(suite.ctx, "attest"))
}

// TestCircuitBreakerAutoExpiry tests automatic expiration
func (suite *KeeperTestSuite) TestCircuitBreakerAutoExpiry() {
	// Enable circuit breaker with 1 hour expiry
	duration := 1 * time.Hour
	cb := types.NewCircuitBreaker(
		true,
		"Temporary pause",
		"cosmos1authority",
		false,
		false,
		false,
		&duration,
	)
	suite.keeper.SetCircuitBreaker(suite.ctx, cb)

	// Operations should be blocked
	err := suite.keeper.IsOperationAllowed(suite.ctx, "deposit")
	suite.Require().Error(err)

	// Verify not expired
	suite.Require().False(cb.IsExpired(suite.ctx.BlockTime()))

	// Move time forward past expiry
	futureTime := suite.ctx.BlockTime().Add(2 * time.Hour)
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	// Operations should now be allowed (auto-expired)
	err = suite.keeper.IsOperationAllowed(suite.ctx, "deposit")
	suite.Require().NoError(err)

	// Verify circuit breaker was disabled
	cb = suite.keeper.GetCircuitBreaker(suite.ctx)
	suite.Require().False(cb.Enabled)
}

// TestCircuitBreakerPermanent tests permanent circuit breaker (no expiry)
func (suite *KeeperTestSuite) TestCircuitBreakerPermanent() {
	// Enable circuit breaker with no expiry
	cb := types.NewCircuitBreaker(
		true,
		"Permanent shutdown",
		"cosmos1authority",
		false,
		false,
		false,
		nil, // No expiry
	)
	suite.keeper.SetCircuitBreaker(suite.ctx, cb)

	// Operations should be blocked
	err := suite.keeper.IsOperationAllowed(suite.ctx, "deposit")
	suite.Require().Error(err)

	// Verify no expiry set
	suite.Require().Nil(cb.ExpiresAt)
	suite.Require().False(cb.IsExpired(suite.ctx.BlockTime()))

	// Move time forward significantly
	futureTime := suite.ctx.BlockTime().Add(365 * 24 * time.Hour) // 1 year
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	// Should still be blocked (permanent)
	err = suite.keeper.IsOperationAllowed(suite.ctx, "deposit")
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrCircuitBreakerActive, err)
}

// TestRateLimitValidation tests rate limit validation
func (suite *KeeperTestSuite) TestRateLimitValidation() {
	// Valid rate limit
	validRL := types.NewRateLimit(
		"ethereum-1",
		"USDT",
		time.Now(),
		1*time.Hour,
		math.NewInt(10_000_000),
	)
	err := validRL.Validate()
	suite.Require().NoError(err)

	// Invalid - empty chain ID
	invalidRL := validRL
	invalidRL.ChainID = ""
	err = invalidRL.Validate()
	suite.Require().Error(err)

	// Invalid - empty asset symbol
	invalidRL = validRL
	invalidRL.AssetSymbol = ""
	err = invalidRL.Validate()
	suite.Require().Error(err)

	// Invalid - window end before start
	invalidRL = validRL
	invalidRL.WindowEnd = invalidRL.WindowStart.Add(-1 * time.Hour)
	err = invalidRL.Validate()
	suite.Require().Error(err)

	// Invalid - negative amount used
	invalidRL = validRL
	invalidRL.AmountUsed = math.NewInt(-1)
	err = invalidRL.Validate()
	suite.Require().Error(err)

	// Invalid - zero max amount
	invalidRL = validRL
	invalidRL.MaxAmount = math.ZeroInt()
	err = invalidRL.Validate()
	suite.Require().Error(err)
}

// TestCircuitBreakerValidation tests circuit breaker validation
func (suite *KeeperTestSuite) TestCircuitBreakerValidation() {
	duration := 1 * time.Hour

	// Valid circuit breaker
	validCB := types.NewCircuitBreaker(
		true,
		"Test reason",
		"cosmos1authority",
		true,
		true,
		true,
		&duration,
	)
	err := validCB.Validate()
	suite.Require().NoError(err)

	// Invalid - enabled but no reason
	invalidCB := validCB
	invalidCB.Reason = ""
	err = invalidCB.Validate()
	suite.Require().Error(err)

	// Invalid - enabled but no triggered by
	invalidCB = validCB
	invalidCB.TriggeredBy = ""
	err = invalidCB.Validate()
	suite.Require().Error(err)

	// Valid - disabled with no reason (ok when disabled)
	disabledCB := validCB
	disabledCB.Enabled = false
	disabledCB.Reason = ""
	err = disabledCB.Validate()
	suite.Require().NoError(err)
}

// TestRateLimitCapacity tests capacity calculations
func (suite *KeeperTestSuite) TestRateLimitCapacity() {
	rl := types.NewRateLimit(
		"ethereum-1",
		"USDT",
		time.Now(),
		1*time.Hour,
		math.NewInt(100_000_000), // 100 HODL max
	)

	// Initially full capacity
	suite.Require().True(rl.HasCapacity(math.NewInt(50_000_000)))
	suite.Require().True(rl.HasCapacity(math.NewInt(100_000_000)))
	suite.Require().False(rl.HasCapacity(math.NewInt(100_000_001)))
	suite.Require().True(rl.RemainingCapacity().Equal(math.NewInt(100_000_000)))

	// Use 60 HODL
	rl.AmountUsed = math.NewInt(60_000_000)
	suite.Require().True(rl.HasCapacity(math.NewInt(40_000_000)))
	suite.Require().False(rl.HasCapacity(math.NewInt(40_000_001)))
	suite.Require().True(rl.RemainingCapacity().Equal(math.NewInt(40_000_000)))

	// Use all capacity
	rl.AmountUsed = math.NewInt(100_000_000)
	suite.Require().False(rl.HasCapacity(math.NewInt(1)))
	suite.Require().True(rl.RemainingCapacity().IsZero())

	// Over capacity (shouldn't happen, but handle gracefully)
	rl.AmountUsed = math.NewInt(110_000_000)
	suite.Require().False(rl.HasCapacity(math.NewInt(1)))
	suite.Require().True(rl.RemainingCapacity().IsZero())
}

// TestCircuitBreakerOperationAllowed tests operation permission logic
func (suite *KeeperTestSuite) TestCircuitBreakerOperationAllowed() {
	duration := 1 * time.Hour

	// Circuit breaker with all operations allowed
	cb := types.NewCircuitBreaker(
		true,
		"Test",
		"authority",
		true,
		true,
		true,
		&duration,
	)

	suite.Require().True(cb.IsOperationAllowed("deposit"))
	suite.Require().True(cb.IsOperationAllowed("withdraw"))
	suite.Require().True(cb.IsOperationAllowed("attest"))
	suite.Require().False(cb.IsOperationAllowed("unknown"))

	// Disable withdrawals
	cb.CanWithdraw = false
	suite.Require().True(cb.IsOperationAllowed("deposit"))
	suite.Require().False(cb.IsOperationAllowed("withdraw"))
	suite.Require().True(cb.IsOperationAllowed("attest"))

	// Disable all
	cb.CanDeposit = false
	cb.CanAttest = false
	suite.Require().False(cb.IsOperationAllowed("deposit"))
	suite.Require().False(cb.IsOperationAllowed("withdraw"))
	suite.Require().False(cb.IsOperationAllowed("attest"))

	// Disabled circuit breaker allows all
	cb.Enabled = false
	suite.Require().True(cb.IsOperationAllowed("deposit"))
	suite.Require().True(cb.IsOperationAllowed("withdraw"))
	suite.Require().True(cb.IsOperationAllowed("attest"))
}

// TestRateLimitPerChain tests that rate limits are per-chain
func (suite *KeeperTestSuite) TestRateLimitPerChain() {
	// Set up two chains
	ethChain := types.NewExternalChain(
		"ethereum-1",
		"Ethereum",
		"evm",
		true,
		12,
		12,
		"0x1234567890abcdef",
		math.NewInt(100_000),
		math.NewInt(10_000_000_000),
	)
	suite.keeper.SetExternalChain(suite.ctx, ethChain)

	bscChain := types.NewExternalChain(
		"bsc-1",
		"Binance Smart Chain",
		"evm",
		true,
		3,
		3,
		"0xabcdef1234567890",
		math.NewInt(100_000),
		math.NewInt(10_000_000_000),
	)
	suite.keeper.SetExternalChain(suite.ctx, bscChain)

	// Same asset on both chains
	ethUsdt := types.NewExternalAsset(
		"ethereum-1",
		"USDT",
		"Tether USD",
		"0xdAC17F958D2ee523a2206206994597C13D831ec7",
		6,
		true,
		math.LegacyOneDec(),
		math.NewInt(1_000_000_000_000),
		math.NewInt(100_000_000_000),
	)
	suite.keeper.SetExternalAsset(suite.ctx, ethUsdt)

	bscUsdt := types.NewExternalAsset(
		"bsc-1",
		"USDT",
		"Tether USD",
		"0x55d398326f99059fF775485246999027B3197955",
		18,
		true,
		math.LegacyOneDec(),
		math.NewInt(1_000_000_000_000),
		math.NewInt(100_000_000_000),
	)
	suite.keeper.SetExternalAsset(suite.ctx, bscUsdt)

	// Set rate limit
	params := suite.keeper.GetParams(suite.ctx)
	params.MaxWithdrawalPerWindow = math.NewInt(10_000_000) // 10 HODL
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Max out Ethereum rate limit
	amount := math.NewInt(10_000_000)
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDT", amount)
	suite.Require().NoError(err)
	suite.keeper.UpdateRateLimit(suite.ctx, "ethereum-1", "USDT", amount)

	// Ethereum should be at limit
	err = suite.keeper.CheckRateLimit(suite.ctx, "ethereum-1", "USDT", math.NewInt(1))
	suite.Require().Error(err)

	// BSC should still have full capacity (independent limit)
	err = suite.keeper.CheckRateLimit(suite.ctx, "bsc-1", "USDT", amount)
	suite.Require().NoError(err)
}
