package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/extbridge/types"
	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
)

// TestCreateTSSSession tests TSS session creation
func (suite *KeeperTestSuite) TestCreateTSSSession() {
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

	// Set up validators
	validators := []sdk.AccAddress{
		sdk.AccAddress("validator1"),
		sdk.AccAddress("validator2"),
		sdk.AccAddress("validator3"),
	}
	suite.stakingKeeper.totalVals = uint64(len(validators))
	for _, val := range validators {
		suite.stakingKeeper.validators[val.String()] = true
		suite.stakingKeeper.validatorTiers[val.String()] = 4
	}

	// Create withdrawal
	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(10_000_000)))

	withdrawalID, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)
	suite.Require().NoError(err)

	// Update withdrawal to ready status and move past timelock
	params := suite.keeper.GetParams(suite.ctx)
	futureTime := suite.ctx.BlockTime().Add(params.WithdrawalTimelockDuration()).Add(1 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	withdrawal, _ := suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)
	withdrawal.Status = types.WithdrawalStatusReady
	suite.keeper.SetWithdrawal(suite.ctx, withdrawal)

	// Create TSS session
	sessionID, err := suite.keeper.CreateTSSSession(suite.ctx, withdrawalID)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(1), sessionID)

	// Verify session was created
	session, found := suite.keeper.GetTSSSession(suite.ctx, sessionID)
	suite.Require().True(found)
	suite.Require().Equal(withdrawalID, session.WithdrawalID)
	suite.Require().Equal("ethereum-1", session.ChainID)
	suite.Require().Equal(types.TSSSessionStatusPending, session.Status)
	suite.Require().Len(session.Participants, 3)
	suite.Require().Equal(uint64(2), session.RequiredSigs) // 67% of 3 = 2
	suite.Require().Equal(uint64(0), session.ReceivedSigs)

	// Verify withdrawal status updated
	withdrawal, _ = suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)
	suite.Require().Equal(types.WithdrawalStatusSigning, withdrawal.Status)
	suite.Require().NotNil(withdrawal.TSSSessionID)
	suite.Require().Equal(sessionID, *withdrawal.TSSSessionID)
}

// TestCreateTSSSessionWithdrawalNotReady tests TSS creation with unready withdrawal
func (suite *KeeperTestSuite) TestCreateTSSSessionWithdrawalNotReady() {
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

	// Create withdrawal (still timelocked)
	withdrawalID, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)
	suite.Require().NoError(err)

	// Try to create TSS session - should fail (timelock not expired)
	_, err = suite.keeper.CreateTSSSession(suite.ctx, withdrawalID)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrWithdrawalNotReady, err)
}

// TestCreateTSSSessionNonExistentWithdrawal tests TSS creation with invalid withdrawal
func (suite *KeeperTestSuite) TestCreateTSSSessionNonExistentWithdrawal() {
	// Try to create TSS session for non-existent withdrawal
	_, err := suite.keeper.CreateTSSSession(suite.ctx, 999)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrWithdrawalNotFound, err)
}

// TestSubmitTSSSignature tests signature submission
func (suite *KeeperTestSuite) TestSubmitTSSSignature() {
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

	validators := []sdk.AccAddress{
		sdk.AccAddress("validator1"),
		sdk.AccAddress("validator2"),
		sdk.AccAddress("validator3"),
	}
	suite.stakingKeeper.totalVals = uint64(len(validators))
	for _, val := range validators {
		suite.stakingKeeper.validators[val.String()] = true
		suite.stakingKeeper.validatorTiers[val.String()] = 4
	}

	// Create withdrawal and TSS session
	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(10_000_000)))

	withdrawalID, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)
	suite.Require().NoError(err)

	params := suite.keeper.GetParams(suite.ctx)
	futureTime := suite.ctx.BlockTime().Add(params.WithdrawalTimelockDuration()).Add(1 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	withdrawal, _ := suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)
	withdrawal.Status = types.WithdrawalStatusReady
	suite.keeper.SetWithdrawal(suite.ctx, withdrawal)

	sessionID, err := suite.keeper.CreateTSSSession(suite.ctx, withdrawalID)
	suite.Require().NoError(err)

	// Submit first signature
	completed, err := suite.keeper.SubmitTSSSignature(
		suite.ctx,
		validators[0],
		sessionID,
		[]byte("signature_share_1"),
	)
	suite.Require().NoError(err)
	suite.Require().False(completed)

	// Verify session updated
	session, _ := suite.keeper.GetTSSSession(suite.ctx, sessionID)
	suite.Require().Equal(types.TSSSessionStatusActive, session.Status)
	suite.Require().Equal(uint64(1), session.ReceivedSigs)
	suite.Require().Len(session.SignatureShares, 1)

	// Submit second signature (should complete)
	completed, err = suite.keeper.SubmitTSSSignature(
		suite.ctx,
		validators[1],
		sessionID,
		[]byte("signature_share_2"),
	)
	suite.Require().NoError(err)
	suite.Require().True(completed)

	// Verify session completed
	session, _ = suite.keeper.GetTSSSession(suite.ctx, sessionID)
	suite.Require().Equal(types.TSSSessionStatusCompleted, session.Status)
	suite.Require().Equal(uint64(2), session.ReceivedSigs)
	suite.Require().NotEmpty(session.CombinedSignature)
	suite.Require().NotNil(session.CompletedAt)

	// Verify withdrawal status updated
	withdrawal, _ = suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)
	suite.Require().Equal(types.WithdrawalStatusSigned, withdrawal.Status)
}

// TestSubmitTSSSignatureDuplicate tests duplicate signature prevention
func (suite *KeeperTestSuite) TestSubmitTSSSignatureDuplicate() {
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

	// Set up 3 validators so session needs 2 signatures (3*67/100 = 2)
	validator := sdk.AccAddress("validator1")
	validator2 := sdk.AccAddress("validator2")
	validator3 := sdk.AccAddress("validator3")
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 4
	suite.stakingKeeper.validators[validator2.String()] = true
	suite.stakingKeeper.validatorTiers[validator2.String()] = 4
	suite.stakingKeeper.validators[validator3.String()] = true
	suite.stakingKeeper.validatorTiers[validator3.String()] = 4
	suite.stakingKeeper.totalVals = 3

	// Create withdrawal and TSS session
	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(10_000_000)))

	withdrawalID, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)
	suite.Require().NoError(err)

	params := suite.keeper.GetParams(suite.ctx)
	futureTime := suite.ctx.BlockTime().Add(params.WithdrawalTimelockDuration()).Add(1 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	withdrawal, _ := suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)
	withdrawal.Status = types.WithdrawalStatusReady
	err = suite.keeper.SetWithdrawal(suite.ctx, withdrawal)
	suite.Require().NoError(err)

	sessionID, err := suite.keeper.CreateTSSSession(suite.ctx, withdrawalID)
	suite.Require().NoError(err)

	// Submit first signature (1/2 required, should not complete)
	completed, err := suite.keeper.SubmitTSSSignature(
		suite.ctx,
		validator,
		sessionID,
		[]byte("signature_share_1"),
	)
	suite.Require().NoError(err)
	suite.Require().False(completed) // Session not yet complete

	// Try to submit again from same validator - should fail
	_, err = suite.keeper.SubmitTSSSignature(
		suite.ctx,
		validator,
		sessionID,
		[]byte("signature_share_2"),
	)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "already submitted")
}

// TestSubmitTSSSignatureInvalidValidator tests signature from non-participant
func (suite *KeeperTestSuite) TestSubmitTSSSignatureInvalidValidator() {
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

	validator := sdk.AccAddress("validator1")
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 4
	suite.stakingKeeper.totalVals = 10

	// Create withdrawal and TSS session
	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(10_000_000)))

	withdrawalID, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)
	suite.Require().NoError(err)

	params := suite.keeper.GetParams(suite.ctx)
	futureTime := suite.ctx.BlockTime().Add(params.WithdrawalTimelockDuration()).Add(1 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	withdrawal, _ := suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)
	withdrawal.Status = types.WithdrawalStatusReady
	suite.keeper.SetWithdrawal(suite.ctx, withdrawal)

	sessionID, err := suite.keeper.CreateTSSSession(suite.ctx, withdrawalID)
	suite.Require().NoError(err)

	// Try to submit signature from non-participant
	nonParticipant := sdk.AccAddress("nonparticipant")
	suite.stakingKeeper.validators[nonParticipant.String()] = true
	suite.stakingKeeper.validatorTiers[nonParticipant.String()] = 4

	_, err = suite.keeper.SubmitTSSSignature(
		suite.ctx,
		nonParticipant,
		sessionID,
		[]byte("signature_share"),
	)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "not a participant")
}

// TestSubmitTSSSignatureNonExistentSession tests signature for invalid session
func (suite *KeeperTestSuite) TestSubmitTSSSignatureNonExistentSession() {
	validator := sdk.AccAddress("validator1")
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 4

	// Try to submit signature for non-existent session
	_, err := suite.keeper.SubmitTSSSignature(
		suite.ctx,
		validator,
		999,
		[]byte("signature_share"),
	)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrTSSSessionNotFound, err)
}

// TestTSSSessionTimeout tests timeout handling
func (suite *KeeperTestSuite) TestTSSSessionTimeout() {
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

	validator := sdk.AccAddress("validator1")
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 4
	suite.stakingKeeper.totalVals = 10

	// Create withdrawal and TSS session
	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(10_000_000)))

	withdrawalID, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)
	suite.Require().NoError(err)

	params := suite.keeper.GetParams(suite.ctx)
	futureTime := suite.ctx.BlockTime().Add(params.WithdrawalTimelockDuration()).Add(1 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	withdrawal, _ := suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)
	withdrawal.Status = types.WithdrawalStatusReady
	suite.keeper.SetWithdrawal(suite.ctx, withdrawal)

	sessionID, err := suite.keeper.CreateTSSSession(suite.ctx, withdrawalID)
	suite.Require().NoError(err)

	// Move time forward past timeout (1 hour)
	timeoutTime := suite.ctx.BlockTime().Add(2 * time.Hour)
	suite.ctx = suite.ctx.WithBlockTime(timeoutTime)

	// Try to submit signature after timeout
	_, err = suite.keeper.SubmitTSSSignature(
		suite.ctx,
		validator,
		sessionID,
		[]byte("signature_share"),
	)
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrTSSTimeout, err)

	// Verify session marked as timeout
	session, _ := suite.keeper.GetTSSSession(suite.ctx, sessionID)
	suite.Require().Equal(types.TSSSessionStatusTimeout, session.Status)
}

// TestTSSSessionCompleted tests signature submission to completed session
func (suite *KeeperTestSuite) TestTSSSessionCompleted() {
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

	validators := []sdk.AccAddress{
		sdk.AccAddress("validator1"),
		sdk.AccAddress("validator2"),
		sdk.AccAddress("validator3"),
	}
	suite.stakingKeeper.totalVals = uint64(len(validators))
	for _, val := range validators {
		suite.stakingKeeper.validators[val.String()] = true
		suite.stakingKeeper.validatorTiers[val.String()] = 4
	}

	// Create withdrawal and TSS session
	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(10_000_000)))

	withdrawalID, err := suite.keeper.RequestWithdrawal(
		suite.ctx,
		sender,
		"ethereum-1",
		"USDT",
		"0xrecipient",
		math.NewInt(5_000_000),
	)
	suite.Require().NoError(err)

	params := suite.keeper.GetParams(suite.ctx)
	futureTime := suite.ctx.BlockTime().Add(params.WithdrawalTimelockDuration()).Add(1 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	withdrawal, _ := suite.keeper.GetWithdrawal(suite.ctx, withdrawalID)
	withdrawal.Status = types.WithdrawalStatusReady
	suite.keeper.SetWithdrawal(suite.ctx, withdrawal)

	sessionID, err := suite.keeper.CreateTSSSession(suite.ctx, withdrawalID)
	suite.Require().NoError(err)

	// Complete session with 2 signatures
	_, err = suite.keeper.SubmitTSSSignature(suite.ctx, validators[0], sessionID, []byte("sig1"))
	suite.Require().NoError(err)
	_, err = suite.keeper.SubmitTSSSignature(suite.ctx, validators[1], sessionID, []byte("sig2"))
	suite.Require().NoError(err)

	// Try to submit signature to completed session
	_, err = suite.keeper.SubmitTSSSignature(suite.ctx, validators[2], sessionID, []byte("sig3"))
	suite.Require().Error(err)
	suite.Require().Equal(types.ErrTSSSessionCompleted, err)
}

// TestGetAllTSSSessions tests retrieving all TSS sessions
func (suite *KeeperTestSuite) TestGetAllTSSSessions() {
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

	validator := sdk.AccAddress("validator1")
	suite.stakingKeeper.validators[validator.String()] = true
	suite.stakingKeeper.validatorTiers[validator.String()] = 4
	suite.stakingKeeper.totalVals = 10

	sender := sdk.AccAddress("sender1")
	suite.bankKeeper.balances[sender.String()] = sdk.NewCoins(sdk.NewCoin(hodltypes.HODLDenom, math.NewInt(100_000_000)))

	// Create multiple withdrawals and TSS sessions
	params := suite.keeper.GetParams(suite.ctx)
	futureTime := suite.ctx.BlockTime().Add(params.WithdrawalTimelockDuration()).Add(1 * time.Second)
	suite.ctx = suite.ctx.WithBlockTime(futureTime)

	for i := 0; i < 3; i++ {
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
		withdrawal.Status = types.WithdrawalStatusReady
		suite.keeper.SetWithdrawal(suite.ctx, withdrawal)

		_, err = suite.keeper.CreateTSSSession(suite.ctx, withdrawalID)
		suite.Require().NoError(err)
	}

	// Get all sessions
	sessions := suite.keeper.GetAllTSSSessions(suite.ctx)
	suite.Require().Len(sessions, 3)

	// Verify IDs are sequential
	suite.Require().Equal(uint64(1), sessions[0].ID)
	suite.Require().Equal(uint64(2), sessions[1].ID)
	suite.Require().Equal(uint64(3), sessions[2].ID)
}
