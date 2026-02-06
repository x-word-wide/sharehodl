package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
)

// TestObserveDeposit tests deposit observation by validators
func (suite *KeeperTestSuite) TestObserveDeposit() {
	// Set up external chain and asset
	chain := types.NewExternalChain(
		"ethereum-1",
		"Ethereum",
		"evm",
		true,
		12,
		12,
		"0x1234567890abcdef",
		math.NewInt(100_000),      // Min 0.1 USDT
		math.NewInt(10_000_000_000), // Max 10,000 USDT
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

	// Set up eligible validators (need 10 to get RequiredAttestations = 7)
	validator := sdk.AccAddress("validator1")
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 4
	suite.stakingKeeper.totalVals = 10
	// Add more validators to reach 10 eligible validators
	for i := 2; i <= 10; i++ {
		v := sdk.AccAddress(fmt.Sprintf("validator%d", i))
		suite.stakingKeeper.validators[v.String()] = true
		suite.stakingKeeper.validatorTiers[v.String()] = 4
	}

	// Test successful deposit observation
	depositID, err := suite.keeper.ObserveDeposit(
		suite.ctx,
		validator,
		"ethereum-1",
		"USDT",
		"0xabcd1234",
		1000,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(1_000_000), // 1 USDT
	)

	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), depositID)

	// Verify deposit was stored
	deposit, found := suite.keeper.GetDeposit(suite.ctx, depositID)
	suite.Require().True(found)
	suite.Require().Equal("ethereum-1", deposit.ChainID)
	suite.Require().Equal("USDT", deposit.AssetSymbol)
	suite.Require().Equal("0xabcd1234", deposit.ExternalTxHash)
	suite.Require().Equal(types.DepositStatusPending, deposit.Status)
	suite.Require().Equal(uint64(0), deposit.Attestations)
	suite.Require().Equal(uint64(7), deposit.RequiredAttestations) // 67% of 10 = 7
}

// TestObserveDepositInvalidValidator tests deposit observation by ineligible validator
func (suite *KeeperTestSuite) TestObserveDepositInvalidValidator() {
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

	// Test with non-validator
	nonValidator := sdk.AccAddress("notvalidator")
	_, err := suite.keeper.ObserveDeposit(
		suite.ctx,
		nonValidator,
		"ethereum-1",
		"USDT",
		"0xabcd1234",
		1000,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(1_000_000),
	)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrNotValidator, err)

	// Test with validator below minimum tier
	lowTierValidator := sdk.AccAddress("lowtierval")
	suite.stakingKeeper.validators[lowTierValidator.String()] = true
	suite.stakingKeeper.validatorTiers[lowTierValidator.String()] = 2 // Below tier 4

	_, err = suite.keeper.ObserveDeposit(
		suite.ctx,
		lowTierValidator,
		"ethereum-1",
		"USDT",
		"0xabcd1234",
		1000,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(1_000_000),
	)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrInsufficientTier, err)
}

// TestObserveDepositAmountLimits tests deposit amount validation
func (suite *KeeperTestSuite) TestObserveDepositAmountLimits() {
	// Set up chain and asset
	chain := types.NewExternalChain(
		"ethereum-1",
		"Ethereum",
		"evm",
		true,
		12,
		12,
		"0x1234567890abcdef",
		math.NewInt(1_000_000),     // Min 1 USDT
		math.NewInt(100_000_000), // Max 100 USDT
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

	// Set up validator
	validator := sdk.AccAddress("validator1")
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 4
	suite.stakingKeeper.totalVals = 10

	// Test amount too small
	_, err := suite.keeper.ObserveDeposit(
		suite.ctx,
		validator,
		"ethereum-1",
		"USDT",
		"0xabcd1234",
		1000,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(500_000), // 0.5 USDT - below minimum
	)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrAmountTooSmall, err)

	// Test amount too large
	_, err = suite.keeper.ObserveDeposit(
		suite.ctx,
		validator,
		"ethereum-1",
		"USDT",
		"0xabcd5678",
		1001,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(200_000_000), // 200 USDT - above maximum
	)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrAmountTooLarge, err)
}

// TestObserveDepositDuplicate tests duplicate deposit prevention
func (suite *KeeperTestSuite) TestObserveDepositDuplicate() {
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

	// Set up validator
	validator := sdk.AccAddress("validator1")
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 4
	suite.stakingKeeper.totalVals = 10

	// First observation succeeds
	txHash := "0xuniquetxhash"
	_, err := suite.keeper.ObserveDeposit(
		suite.ctx,
		validator,
		"ethereum-1",
		"USDT",
		txHash,
		1000,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(1_000_000),
	)
	suite.Require().NoError(err)

	// Second observation with same tx hash should fail
	_, err = suite.keeper.ObserveDeposit(
		suite.ctx,
		validator,
		"ethereum-1",
		"USDT",
		txHash,
		1000,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(1_000_000),
	)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrDuplicateDeposit, err)
}

// TestAttestDeposit tests deposit attestation
func (suite *KeeperTestSuite) TestAttestDeposit() {
	// Set up chain, asset, and validators
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

	// Set up 3 validators
	suite.stakingKeeper.totalVals = 3
	validators := []sdk.AccAddress{
		sdk.AccAddress("validator1"),
		sdk.AccAddress("validator2"),
		sdk.AccAddress("validator3"),
	}
	for _, val := range validators {
		suite.stakingKeeper.validators[val.String()] = true
		suite.stakingKeeper.validatorTiers[val.String()] = 4
	}

	// Create deposit
	depositID, err := suite.keeper.ObserveDeposit(
		suite.ctx,
		validators[0],
		"ethereum-1",
		"USDT",
		"0xtxhash",
		1000,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(1_000_000),
	)
	suite.Require().NoError(err)

	// First attestation (1/3 required - 67% of 3 = 3) - must provide observed values matching deposit
	attestCount, completed, err := suite.keeper.AttestDeposit(suite.ctx, validators[0], depositID, true, "0xtxhash", math.NewInt(1_000_000))
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), attestCount)
	suite.Require().False(completed)

	deposit, _ := suite.keeper.GetDeposit(suite.ctx, depositID)
	suite.Require().Equal(types.DepositStatusAttesting, deposit.Status)

	// Second attestation (2/3 required) - must provide observed values matching deposit
	attestCount, completed, err = suite.keeper.AttestDeposit(suite.ctx, validators[1], depositID, true, "0xtxhash", math.NewInt(1_000_000))
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(2), attestCount)
	suite.Require().False(completed) // Not yet complete, need 3

	// Third attestation (3/3 required - should complete) - must provide observed values matching deposit
	attestCount, completed, err = suite.keeper.AttestDeposit(suite.ctx, validators[2], depositID, true, "0xtxhash", math.NewInt(1_000_000))
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(3), attestCount)
	suite.Require().True(completed)

	// Verify deposit completed
	deposit, _ = suite.keeper.GetDeposit(suite.ctx, depositID)
	suite.Require().Equal(types.DepositStatusCompleted, deposit.Status)
	suite.Require().NotNil(deposit.CompletedAt)
}

// TestAttestDepositDuplicate tests duplicate attestation prevention
func (suite *KeeperTestSuite) TestAttestDepositDuplicate() {
	// Set up chain, asset, and validator
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

	// Set up multiple validators so deposit doesn't auto-complete on first attestation
	validator := sdk.AccAddress("validator1")
	validator2 := sdk.AccAddress("validator2")
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 4
	suite.stakingKeeper.validators[validator2.String()] = true
	suite.stakingKeeper.validatorTiers[validator2.String()] = 4
	suite.stakingKeeper.totalVals = 2

	// Create deposit
	depositID, err := suite.keeper.ObserveDeposit(
		suite.ctx,
		validator,
		"ethereum-1",
		"USDT",
		"0xtxhash",
		1000,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(1_000_000),
	)
	suite.Require().NoError(err)

	// First attestation succeeds - with matching observed values (1/2 required)
	_, completed, err := suite.keeper.AttestDeposit(suite.ctx, validator, depositID, true, "0xtxhash", math.NewInt(1_000_000))
	suite.Require().NoError(err)
	suite.Require().False(completed) // Not yet complete

	// Second attestation from same validator should fail with ErrAlreadyAttested
	_, _, err = suite.keeper.AttestDeposit(suite.ctx, validator, depositID, true, "0xtxhash", math.NewInt(1_000_000))
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrAlreadyAttested, err)
}

// TestAttestDepositCompleted tests attestation of completed deposit
func (suite *KeeperTestSuite) TestAttestDepositCompleted() {
	// Set up chain, asset, and validators
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

	// Set up 3 validators (67% of 3 = 3 required attestations)
	suite.stakingKeeper.totalVals = 3
	val1 := sdk.AccAddress("validator1")
	val2 := sdk.AccAddress("validator2")
	val3 := sdk.AccAddress("validator3")
	val4 := sdk.AccAddress("validator4")

	for _, val := range []sdk.AccAddress{val1, val2, val3, val4} {
		suite.stakingKeeper.validators[val.String()] = true
		suite.stakingKeeper.validatorTiers[val.String()] = 4
	}

	// Create deposit - with 4 eligible validators, need 3 attestations (67% of 4 = 2.68 → 3)
	depositID, err := suite.keeper.ObserveDeposit(
		suite.ctx,
		val1,
		"ethereum-1",
		"USDT",
		"0xtxhash",
		1000,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(1_000_000),
	)
	suite.Require().NoError(err)

	// Complete deposit with 3 attestations - with matching observed values
	_, _, err = suite.keeper.AttestDeposit(suite.ctx, val1, depositID, true, "0xtxhash", math.NewInt(1_000_000))
	suite.Require().NoError(err)
	_, _, err = suite.keeper.AttestDeposit(suite.ctx, val2, depositID, true, "0xtxhash", math.NewInt(1_000_000))
	suite.Require().NoError(err)
	_, completed, err := suite.keeper.AttestDeposit(suite.ctx, val3, depositID, true, "0xtxhash", math.NewInt(1_000_000))
	suite.Require().NoError(err)
	suite.Require().True(completed)

	// Try to attest completed deposit - should fail
	_, _, err = suite.keeper.AttestDeposit(suite.ctx, val4, depositID, true, "0xtxhash", math.NewInt(1_000_000))
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrDepositCompleted, err)
}

// TestGetDepositByTxHash tests deposit lookup by transaction hash
func (suite *KeeperTestSuite) TestGetDepositByTxHash() {
	// Set up chain, asset, and validator
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

	validator := sdk.AccAddress("validator1")
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 4
	suite.stakingKeeper.totalVals = 10

	// Create deposit
	txHash := "0xuniquetxhash123"
	depositID, err := suite.keeper.ObserveDeposit(
		suite.ctx,
		validator,
		"ethereum-1",
		"USDT",
		txHash,
		1000,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(1_000_000),
	)
	suite.Require().NoError(err)

	// Retrieve by tx hash
	deposit, found := suite.keeper.GetDepositByTxHash(suite.ctx, "ethereum-1", txHash)
	suite.Require().True(found)
	suite.Require().Equal(depositID, deposit.ID)
	suite.Require().Equal(txHash, deposit.ExternalTxHash)

	// Test non-existent tx hash
	_, found = suite.keeper.GetDepositByTxHash(suite.ctx, "ethereum-1", "0xnonexistent")
	suite.Require().False(found)
}

// TestDepositWithCircuitBreaker tests deposit observation with circuit breaker active
func (suite *KeeperTestSuite) TestDepositWithCircuitBreaker() {
	// Set up chain, asset, and validator
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

	validator := sdk.AccAddress("validator1")
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 4
	suite.stakingKeeper.totalVals = 10

	// Enable circuit breaker blocking deposits
	duration := 1 * time.Hour
	cb := types.NewCircuitBreaker(
		true,
		"Security incident",
		"cosmos1authority",
		false, // Deposits disabled
		true,
		true,
		&duration,
	)
	suite.keeper.SetCircuitBreaker(suite.ctx, cb)

	// Attempt deposit observation - should fail
	_, err := suite.keeper.ObserveDeposit(
		suite.ctx,
		validator,
		"ethereum-1",
		"USDT",
		"0xtxhash",
		1000,
		"0xsender",
		sdk.AccAddress("recipient1").String(),
		math.NewInt(1_000_000),
	)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrCircuitBreakerActive, err)
}
