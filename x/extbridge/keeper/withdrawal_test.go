package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

// TestRequestWithdrawal tests withdrawal request creation
func (suite *KeeperTestSuite) TestRequestWithdrawal() {
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

	// Set up user with HODL balance
	sender := sdk.AccAddress("sender1")
	hodlAmount := math.NewInt(10_000_000) // 10 HODL
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, hodlAmount))

	// Request withdrawal
	withdrawalID, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		hodlAmount,
	)

	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), withdrawalID)

	// Verify withdrawal was stored
	withdrawal, found := suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)
	suite.Require().True(found)
	suite.Require().Equal("ethereum-1", withdrawal.ChainID)
	suite.Require().Equal("USDT", withdrawal.AssetSymbol)
	suite.Require().Equal(sender.String(), withdrawal.Sender)
	suite.Require().Equal("0xrecipient", withdrawal.Recipient)
	suite.Require().Equal(types.WithdrawalStatusPending, withdrawal.Status)

	// Verify fee was calculated
	params := suite.keeper.GetParams(suite.ctx)
	expectedFee := math.LegacyNewDecFromInt(hodlAmount).Mul(params.BridgeFee).TruncateInt()
	suite.Require().Equal(expectedFee, withdrawal.FeeAmount)

	// Verify net HODL amount
	netHODL := hodlAmount.Sub(expectedFee)
	suite.Require().Equal(netHODL, withdrawal.HODLAmount.Sub(withdrawal.FeeAmount))

	// Verify timelock was set
	expectedExpiry := suite.ctx.BlockTime().Add(params.WithdrawalTimelockDuration())
	suite.Require().True(withdrawal.TimelockExpiry.Equal(expectedExpiry) || withdrawal.TimelockExpiry.After(expectedExpiry.Add(-1*time.Second)))
}

// TestRequestWithdrawalInsufficientBalance tests withdrawal with insufficient balance
func (suite *KeeperTestSuite) TestRequestWithdrawalInsufficientBalance() {
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

	// Set up user with low balance
	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(1_000_000))) // 1 HODL

	// Try to withdraw 10 HODL
	_, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(10_000_000),
	)

	suite.Require().Error(err)
	suite.Require().Equal(types.ErrInsufficientBalance, err)
}

// TestRequestWithdrawalUnsupportedChain tests withdrawal to unsupported chain
func (suite *KeeperTestSuite) TestRequestWithdrawalUnsupportedChain() {
	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(10_000_000)))

	// Try withdrawal to non-existent chain
	_, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"bitcoin-1",
		"BTC",
		"bc1recipient",
		math.NewInt(10_000_000),
	)

	suite.Require().Error(err)
	suite.Require().Equal(types.ErrChainNotSupported, err)
}

// TestRequestWithdrawalDisabledChain tests withdrawal to disabled chain
func (suite *KeeperTestSuite) TestRequestWithdrawalDisabledChain() {
	// Set up disabled chain
	chain := types.NewExternalChain(
		"ethereum-1",
		"Ethereum",
		"evm",
		false, // Disabled
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

	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(10_000_000)))

	// Try withdrawal
	_, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(10_000_000),
	)

	suite.Require().Error(err)
	suite.Require().Equal(types.ErrChainNotSupported, err)
}

// TestRequestWithdrawalRateLimit tests rate limiting
func (suite *KeeperTestSuite) TestRequestWithdrawalRateLimit() {
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

	// Set low rate limit for testing
	params := suite.keeper.GetParams(suite.ctx)
	params.MaxWithdrawalPerWindow = math.NewInt(10_000_000) // 10 HODL max per window
	err := suite.keeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	// Set up user with large balance
	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(100_000_000))) // 100 HODL

	// First withdrawal of 8 HODL should succeed
	_, err = suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(8_000_000),
	)
	suite.Require().NoError(err)

	// Second withdrawal of 8 HODL should fail (would exceed 10 HODL limit)
	_, err = suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(8_000_000),
	)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrRateLimitExceeded, err)

	// Withdrawal of 2 HODL should succeed (total 10 HODL)
	_, err = suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(2_000_000),
	)
	suite.Require().NoError(err)
}

// TestWithdrawalTimelock tests timelock enforcement
func (suite *KeeperTestSuite) TestWithdrawalTimelock() {
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

	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(10_000_000)))

	// Create withdrawal
	withdrawalID, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)
	suite.Require().NoError(err)

	withdrawal, _ := suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)

	// Check timelock not expired
	suite.Require().False(withdrawal.IsTimelockExpired(suite.ctx.BlockTime()))
	suite.Require().False(withdrawal.CanSign(suite.ctx.BlockTime()))

	// Move time forward past timelock
	params := suite.keeper.GetParams(suite.ctx)
	futureTime := suite.ctx.BlockTime().Add(params.WithdrawalTimelockDuration()).Add(1 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	// Check timelock expired
	suite.Require().True(withdrawal.IsTimelockExpired(futureTime))

	// Update withdrawal status to ready
	withdrawal.Status = types.WithdrawalStatusReady
	suite.keeper.SetWithdrawal(suite.ctx, withdrawal)

	// Now it can be signed
	withdrawal, _ = suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)
	suite.Require().True(withdrawal.CanSign(futureTime))
}

// TestProcessWithdrawals tests automatic processing of timelocked withdrawals
func (suite *KeeperTestSuite) TestProcessWithdrawals() {
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

	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(20_000_000)))

	// Create two withdrawals
	withdrawalID1, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)
	suite.Require().NoError(err)

	withdrawalID2, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)
	suite.Require().NoError(err)

	// Verify both are pending
	w1, _ := suite.keeper.GetWithdrawal(suite.ctx, withdrawalID1)
	w2, _ := suite.keeper.GetWithdrawal(suite.ctx, withdrawalID2)
	suite.Require().Equal(types.WithdrawalStatusPending, w1.Status)
	suite.Require().Equal(types.WithdrawalStatusPending, w2.Status)

	// Move time forward past timelock
	params := suite.keeper.GetParams(suite.ctx)
	futureTime := suite.ctx.BlockTime().Add(params.WithdrawalTimelockDuration()).Add(1 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	// Process withdrawals
	suite.keeper.ProcessWithdrawals(suite.ctx)

	// Verify both are now ready
	w1, _ = suite.keeper.GetWithdrawal(suite.ctx, withdrawalID1)
	w2, _ = suite.keeper.GetWithdrawal(suite.ctx, withdrawalID2)
	suite.Require().Equal(types.WithdrawalStatusReady, w1.Status)
	suite.Require().Equal(types.WithdrawalStatusReady, w2.Status)
}

// TestWithdrawalStatusTransitions tests withdrawal status state machine
func (suite *KeeperTestSuite) TestWithdrawalStatusTransitions() {
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

	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(10_000_000)))

	// Create withdrawal
	withdrawalID, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)
	suite.Require().NoError(err)

	withdrawal, _ := suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)

	// Test initial state
	suite.Require().Equal(types.WithdrawalStatusPending, withdrawal.Status)
	suite.Require().False(withdrawal.IsCompleted())
	suite.Require().False(withdrawal.IsCancelled())
	suite.Require().False(withdrawal.IsFailed())
	suite.Require().False(withdrawal.IsFinal())

	// Test ready state
	withdrawal.Status = types.WithdrawalStatusReady
	suite.Require().False(withdrawal.IsFinal())

	// Test completed state
	withdrawal.Status = types.WithdrawalStatusCompleted
	now := time.Now()
	withdrawal.CompletedAt = &now
	suite.Require().True(withdrawal.IsCompleted())
	suite.Require().True(withdrawal.IsFinal())

	// Test cancelled state
	withdrawal.Status = types.WithdrawalStatusCancelled
	suite.Require().True(withdrawal.IsCancelled())
	suite.Require().True(withdrawal.IsFinal())

	// Test failed state
	withdrawal.Status = types.WithdrawalStatusFailed
	withdrawal.FailureReason = "External chain failure"
	suite.Require().True(withdrawal.IsFailed())
	suite.Require().True(withdrawal.IsFinal())
}

// TestWithdrawalWithCircuitBreaker tests withdrawal with circuit breaker
func (suite *KeeperTestSuite) TestWithdrawalWithCircuitBreaker() {
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

	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(10_000_000)))

	// Enable circuit breaker blocking withdrawals
	duration := 1 * time.Hour
	cb := types.NewCircuitBreaker(
		true,
		"Security incident",
		"cosmos1authority",
		true,
		false, // Withdrawals disabled
		true,
		&duration,
	)
	suite.keeper.SetCircuitBreaker(suite.ctx, cb)

	// Attempt withdrawal - should fail
	_, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)

	suite.Require().Error(err)
	suite.Require().Equal(types.ErrCircuitBreakerActive, err)
}

// TestGetAllWithdrawals tests retrieving all withdrawals
func (suite *KeeperTestSuite) TestGetAllWithdrawals() {
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

	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(100_000_000)))

	// Create multiple withdrawals
	for i := 0; i < 3; i++ {
		_, err := suite.keeper.RequestWithdrawal(
			suite.ctx,
			sender,
			"ethereum-1",
			"USDT",
			"0xrecipient",
			math.NewInt(5_000_000),
		)
		suite.Require().NoError(err)
	}

	// Get all withdrawals
	withdrawals := suite.keeper.GetAllWithdrawals(suite.ctx)
	suite.Require().Len(withdrawals, 3)

	// Verify IDs are sequential
	suite.Require().Equal(uint64(1), withdrawals[0].ID)
	suite.Require().Equal(uint64(2), withdrawals[1].ID)
	suite.Require().Equal(uint64(3), withdrawals[2].ID)
}
