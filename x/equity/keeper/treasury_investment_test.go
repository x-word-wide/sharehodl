package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// TreasuryInvestmentTestSuite is a test suite for treasury equity investment features
// NOTE: These tests are currently skipped pending full keeper initialization setup.
// The tests demonstrate the expected behavior and serve as documentation for the
// treasury equity investment feature.
type TreasuryInvestmentTestSuite struct {
	suite.Suite

	// Test accounts
	alice     sdk.AccAddress
	bob       sdk.AccAddress
	cofounder sdk.AccAddress

	// Test companies
	companyA    types.Company
	companyB    types.Company
	shareClassA types.ShareClass
	shareClassB types.ShareClass
}

// SetupTest initializes the test suite before each test
func (suite *TreasuryInvestmentTestSuite) SetupTest() {
	// Create test addresses
	suite.alice = sdk.AccAddress([]byte("alice_______________"))
	suite.bob = sdk.AccAddress([]byte("bob_________________"))
	suite.cofounder = sdk.AccAddress([]byte("cofounder___________"))

	// Setup test companies
	suite.companyA = types.NewCompany(
		1,
		"Company A",
		"CMPA",
		"First test company",
		suite.alice.String(),
		"https://companya.com",
		"Technology",
		"US",
		math.NewInt(1000000),
		math.NewInt(2000000),
		math.LegacyNewDec(1),
		sdk.NewCoins(sdk.NewInt64Coin("uhodl", 1000000)),
	)
	suite.companyA.Status = types.CompanyStatusActive

	suite.companyB = types.NewCompany(
		2,
		"Company B",
		"CMPB",
		"Second test company",
		suite.bob.String(),
		"https://companyb.com",
		"Finance",
		"US",
		math.NewInt(500000),
		math.NewInt(1000000),
		math.LegacyNewDec(1),
		sdk.NewCoins(sdk.NewInt64Coin("uhodl", 500000)),
	)
	suite.companyB.Status = types.CompanyStatusActive

	suite.shareClassA = types.NewShareClass(
		1,
		"COMMON",
		"Common Stock",
		"Standard voting shares",
		true,
		true,
		false,
		false,
		math.NewInt(1000000),
		math.LegacyNewDec(1),
		true,
	)

	suite.shareClassB = types.NewShareClass(
		2,
		"COMMON",
		"Common Stock",
		"Standard voting shares",
		true,
		true,
		false,
		false,
		math.NewInt(500000),
		math.LegacyNewDec(1),
		true,
	)
}

// =============================================================================
// Test: DepositEquityToTreasury
// =============================================================================

// TestDepositEquityToTreasury_Success tests successful equity deposit
func (suite *TreasuryInvestmentTestSuite) TestDepositEquityToTreasury_Success() {
	// TODO: This test requires a fully initialized keeper
	// For now, we'll define the test structure
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Company B treasury is initialized, cofounder owns Company A shares
	treasuryB := types.NewCompanyTreasury(2)
	suite.equityKeeper.SetCompanyTreasury(suite.ctx, treasuryB)

	// Cofounder owns 10,000 shares of Company A
	sharesDeposited := math.NewInt(10000)
	shareholding := types.NewShareholding(1, "COMMON", suite.cofounder.String(), sharesDeposited, math.LegacyNewDec(10))
	suite.equityKeeper.SetShareholding(suite.ctx, shareholding)

	// WHEN: Cofounder deposits Company A shares to Company B treasury
	err := suite.equityKeeper.DepositEquityToTreasury(
		suite.ctx,
		2,                           // Company B (treasury owner)
		suite.cofounder.String(),    // Depositor
		1,                           // Company A (target company)
		"COMMON",                    // Share class
		sharesDeposited,             // Shares to deposit
	)

	// THEN: Deposit succeeds
	require.NoError(suite.T(), err)

	// AND: Treasury investment holdings are updated
	updatedTreasury, found := suite.equityKeeper.GetCompanyTreasury(suite.ctx, 2)
	require.True(suite.T(), found)
	require.Equal(suite.T(), sharesDeposited, updatedTreasury.GetInvestmentShares(1, "COMMON"))

	// AND: Shares transferred from cofounder to module account
	moduleAddr := suite.accountKeeper.GetModuleAddress(types.ModuleName)
	moduleHolding, found := suite.equityKeeper.GetShareholding(suite.ctx, 1, "COMMON", moduleAddr.String())
	require.True(suite.T(), found)
	require.Equal(suite.T(), sharesDeposited, moduleHolding.Shares)

	// AND: Beneficial ownership is registered for treasury
	expectedBeneficialOwner := fmt.Sprintf("treasury:%d:%s", 2, suite.companyB.Symbol)
	beneficialOwners, err := suite.equityKeeper.GetBeneficialOwnersForModule(suite.ctx, moduleAddr.String(), 1, "COMMON")
	require.NoError(suite.T(), err)
	require.Len(suite.T(), beneficialOwners, 1)
	require.Equal(suite.T(), expectedBeneficialOwner, beneficialOwners[0].BeneficialOwner)
	require.Equal(suite.T(), sharesDeposited, beneficialOwners[0].Shares)
	*/
}

// TestDepositEquityToTreasury_UnauthorizedDepositor tests rejection by non-authorized user
func (suite *TreasuryInvestmentTestSuite) TestDepositEquityToTreasury_UnauthorizedDepositor() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Alice is not authorized for Company B
	// WHEN: Alice tries to deposit shares to Company B treasury
	err := suite.equityKeeper.DepositEquityToTreasury(
		suite.ctx,
		2,                       // Company B
		suite.alice.String(),    // Unauthorized depositor
		1,                       // Company A
		"COMMON",
		math.NewInt(1000),
	)

	// THEN: Deposit is rejected
	require.Error(suite.T(), err)
	require.ErrorIs(suite.T(), err, types.ErrNotAuthorizedProposer)
	*/
}

// TestDepositEquityToTreasury_InsufficientShares tests deposit with insufficient shares
func (suite *TreasuryInvestmentTestSuite) TestDepositEquityToTreasury_InsufficientShares() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Cofounder owns only 100 shares
	shareholding := types.NewShareholding(1, "COMMON", suite.cofounder.String(), math.NewInt(100), math.LegacyNewDec(10))
	suite.equityKeeper.SetShareholding(suite.ctx, shareholding)

	// WHEN: Cofounder tries to deposit 1000 shares (more than owned)
	err := suite.equityKeeper.DepositEquityToTreasury(
		suite.ctx,
		2,
		suite.cofounder.String(),
		1,
		"COMMON",
		math.NewInt(1000), // More than owned
	)

	// THEN: Deposit is rejected
	require.Error(suite.T(), err)
	require.ErrorIs(suite.T(), err, types.ErrInsufficientShares)
	*/
}

// TestDepositEquityToTreasury_FrozenTreasury tests deposit to frozen treasury
func (suite *TreasuryInvestmentTestSuite) TestDepositEquityToTreasury_FrozenTreasury() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Company B treasury is frozen
	suite.equityKeeper.FreezeTreasury(suite.ctx, 2, "Fraud investigation")

	// WHEN: Attempting to deposit
	err := suite.equityKeeper.DepositEquityToTreasury(
		suite.ctx,
		2,
		suite.cofounder.String(),
		1,
		"COMMON",
		math.NewInt(1000),
	)

	// THEN: Deposit is rejected
	require.Error(suite.T(), err)
	require.ErrorIs(suite.T(), err, types.ErrTreasuryFrozenForInvestigation)
	*/
}

// =============================================================================
// Test: Equity Withdrawal Flow
// =============================================================================

// TestEquityWithdrawal_FullFlow tests the complete withdrawal workflow
func (suite *TreasuryInvestmentTestSuite) TestEquityWithdrawal_FullFlow() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Company B treasury holds 10,000 shares of Company A
	treasuryB := types.NewCompanyTreasury(2)
	if treasuryB.InvestmentHoldings == nil {
		treasuryB.InvestmentHoldings = make(map[uint64]map[string]math.Int)
	}
	if treasuryB.InvestmentHoldings[1] == nil {
		treasuryB.InvestmentHoldings[1] = make(map[string]math.Int)
	}
	treasuryB.InvestmentHoldings[1]["COMMON"] = math.NewInt(10000)
	suite.equityKeeper.SetCompanyTreasury(suite.ctx, treasuryB)

	sharesToWithdraw := math.NewInt(5000)

	// WHEN: Request withdrawal (creates pending request)
	requestID, err := suite.equityKeeper.RequestEquityWithdrawal(
		suite.ctx,
		2,                        // Company B (treasury owner)
		1,                        // Company A (target company)
		"COMMON",
		sharesToWithdraw,
		suite.bob.String(),       // Recipient
		"Selling to investor",
		suite.bob.String(),       // Requested by (company B founder)
	)

	// THEN: Request is created
	require.NoError(suite.T(), err)
	require.Greater(suite.T(), requestID, uint64(0))

	// AND: Shares are locked in treasury
	updatedTreasury, _ := suite.equityKeeper.GetCompanyTreasury(suite.ctx, 2)
	require.Equal(suite.T(), sharesToWithdraw, updatedTreasury.GetLockedEquity(1, "COMMON"))

	// AND: Request is in pending status
	request, found := suite.equityKeeper.GetEquityWithdrawalRequest(suite.ctx, requestID)
	require.True(suite.T(), found)
	require.Equal(suite.T(), types.WithdrawalStatusPending, request.Status)

	// WHEN: Governance approves withdrawal
	proposalID := uint64(100)
	err = suite.equityKeeper.ApproveEquityWithdrawal(suite.ctx, requestID, proposalID)
	require.NoError(suite.T(), err)

	// THEN: Request status is approved
	request, _ = suite.equityKeeper.GetEquityWithdrawalRequest(suite.ctx, requestID)
	require.Equal(suite.T(), types.WithdrawalStatusApproved, request.Status)
	require.Equal(suite.T(), proposalID, request.ProposalID)

	// WHEN: Execute withdrawal (after timelock)
	err = suite.equityKeeper.ExecuteEquityWithdrawal(suite.ctx, requestID)
	require.NoError(suite.T(), err)

	// THEN: Shares are transferred to recipient
	recipientHolding, found := suite.equityKeeper.GetShareholding(suite.ctx, 1, "COMMON", suite.bob.String())
	require.True(suite.T(), found)
	require.Equal(suite.T(), sharesToWithdraw, recipientHolding.Shares)

	// AND: Treasury investment holdings are reduced
	finalTreasury, _ := suite.equityKeeper.GetCompanyTreasury(suite.ctx, 2)
	expectedRemaining := math.NewInt(5000) // 10000 - 5000
	require.Equal(suite.T(), expectedRemaining, finalTreasury.GetInvestmentShares(1, "COMMON"))

	// AND: Shares are unlocked
	require.Equal(suite.T(), math.ZeroInt(), finalTreasury.GetLockedEquity(1, "COMMON"))

	// AND: Request status is executed
	request, _ = suite.equityKeeper.GetEquityWithdrawalRequest(suite.ctx, requestID)
	require.Equal(suite.T(), types.WithdrawalStatusExecuted, request.Status)
	require.False(suite.T(), request.ExecutedAt.IsZero())
	*/
}

// TestEquityWithdrawal_WithoutApproval tests execution without approval
func (suite *TreasuryInvestmentTestSuite) TestEquityWithdrawal_WithoutApproval() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Withdrawal request is pending
	requestID := uint64(1)
	request := types.NewEquityWithdrawalRequest(
		requestID,
		2,
		1,
		"COMMON",
		math.NewInt(1000),
		suite.bob.String(),
		"Test withdrawal",
		suite.bob.String(),
	)
	suite.equityKeeper.SetEquityWithdrawalRequest(suite.ctx, request)

	// WHEN: Attempting to execute without approval
	err := suite.equityKeeper.ExecuteEquityWithdrawal(suite.ctx, requestID)

	// THEN: Execution is rejected
	require.Error(suite.T(), err)
	require.ErrorIs(suite.T(), err, types.ErrWithdrawalNotApproved)
	*/
}

// TestEquityWithdrawal_Rejection tests withdrawal rejection
func (suite *TreasuryInvestmentTestSuite) TestEquityWithdrawal_Rejection() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Pending withdrawal with locked shares
	treasuryB := types.NewCompanyTreasury(2)
	if treasuryB.InvestmentHoldings == nil {
		treasuryB.InvestmentHoldings = make(map[uint64]map[string]math.Int)
	}
	if treasuryB.InvestmentHoldings[1] == nil {
		treasuryB.InvestmentHoldings[1] = make(map[string]math.Int)
	}
	treasuryB.InvestmentHoldings[1]["COMMON"] = math.NewInt(10000)

	if treasuryB.LockedEquity == nil {
		treasuryB.LockedEquity = make(map[uint64]map[string]math.Int)
	}
	if treasuryB.LockedEquity[1] == nil {
		treasuryB.LockedEquity[1] = make(map[string]math.Int)
	}
	lockedShares := math.NewInt(3000)
	treasuryB.LockedEquity[1]["COMMON"] = lockedShares
	suite.equityKeeper.SetCompanyTreasury(suite.ctx, treasuryB)

	requestID := uint64(1)
	request := types.NewEquityWithdrawalRequest(
		requestID,
		2,
		1,
		"COMMON",
		lockedShares,
		suite.bob.String(),
		"Test withdrawal",
		suite.bob.String(),
	)
	suite.equityKeeper.SetEquityWithdrawalRequest(suite.ctx, request)

	// WHEN: Governance rejects withdrawal
	err := suite.equityKeeper.RejectEquityWithdrawal(suite.ctx, requestID)
	require.NoError(suite.T(), err)

	// THEN: Request status is rejected
	updatedRequest, found := suite.equityKeeper.GetEquityWithdrawalRequest(suite.ctx, requestID)
	require.True(suite.T(), found)
	require.Equal(suite.T(), types.WithdrawalStatusRejected, updatedRequest.Status)

	// AND: Shares are unlocked
	updatedTreasury, _ := suite.equityKeeper.GetCompanyTreasury(suite.ctx, 2)
	require.Equal(suite.T(), math.ZeroInt(), updatedTreasury.GetLockedEquity(1, "COMMON"))

	// AND: Investment holdings remain unchanged
	require.Equal(suite.T(), math.NewInt(10000), updatedTreasury.GetInvestmentShares(1, "COMMON"))
	*/
}

// TestEquityWithdrawal_UnauthorizedRequest tests unauthorized withdrawal request
func (suite *TreasuryInvestmentTestSuite) TestEquityWithdrawal_UnauthorizedRequest() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Alice is not authorized for Company B
	// WHEN: Alice tries to request withdrawal
	_, err := suite.equityKeeper.RequestEquityWithdrawal(
		suite.ctx,
		2,                        // Company B
		1,                        // Company A
		"COMMON",
		math.NewInt(1000),
		suite.alice.String(),     // Recipient
		"Unauthorized withdrawal",
		suite.alice.String(),     // Requested by (not authorized)
	)

	// THEN: Request is rejected
	require.Error(suite.T(), err)
	require.ErrorIs(suite.T(), err, types.ErrNotTreasuryController)
	*/
}

// =============================================================================
// Test: Cross-Company Dividends
// =============================================================================

// TestCrossCompanyDividend_TreasuryReceivesDividend tests treasury receiving dividends
func (suite *TreasuryInvestmentTestSuite) TestCrossCompanyDividend_TreasuryReceivesDividend() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Company B treasury holds 10,000 shares of Company A
	treasuryB := types.NewCompanyTreasury(2)
	if treasuryB.InvestmentHoldings == nil {
		treasuryB.InvestmentHoldings = make(map[uint64]map[string]math.Int)
	}
	if treasuryB.InvestmentHoldings[1] == nil {
		treasuryB.InvestmentHoldings[1] = make(map[string]math.Int)
	}
	treasuryB.InvestmentHoldings[1]["COMMON"] = math.NewInt(10000)
	treasuryB.Balance = sdk.NewCoins() // Start with zero
	suite.equityKeeper.SetCompanyTreasury(suite.ctx, treasuryB)

	// AND: Company A declares dividend of 1 HODL per share
	dividendPerShare := math.LegacyNewDec(1)
	totalDividend := dividendPerShare.MulInt(math.NewInt(10000)) // 10,000 HODL

	// WHEN: Treasury receives dividend credit
	dividendCoins := sdk.NewCoins(sdk.NewCoin("uhodl", totalDividend.TruncateInt()))
	err := suite.equityKeeper.CreditTreasuryDividend(
		suite.ctx,
		2,                 // Company B (recipient)
		1,                 // Company A (source)
		dividendCoins,
	)

	// THEN: Dividend is credited successfully
	require.NoError(suite.T(), err)

	// AND: Company B treasury balance is increased
	updatedTreasury, found := suite.equityKeeper.GetCompanyTreasury(suite.ctx, 2)
	require.True(suite.T(), found)
	require.Equal(suite.T(), dividendCoins, updatedTreasury.Balance)

	// AND: TotalDeposited is increased
	require.Equal(suite.T(), dividendCoins, updatedTreasury.TotalDeposited)
	*/
}

// TestCrossCompanyDividend_NoSelfDividend tests that company doesn't pay itself
func (suite *TreasuryInvestmentTestSuite) TestCrossCompanyDividend_NoSelfDividend() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Company A pays dividend
	// AND: GetBeneficialOwnersForDividend is called for Company A
	// WHEN: Checking treasury investment holders
	holders := suite.equityKeeper.GetTreasuryInvestmentHoldersForDividend(suite.ctx, 1, "COMMON")

	// THEN: Company A's own treasury is excluded
	for _, holder := range holders {
		require.NotEqual(suite.T(), uint64(1), holder.OwnerCompanyID,
			"Company should not receive dividend from itself")
	}
	*/
}

// TestCrossCompanyDividend_MultipleInvestors tests dividends to multiple treasury holders
func (suite *TreasuryInvestmentTestSuite) TestCrossCompanyDividend_MultipleInvestors() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Multiple companies hold Company A shares
	// Company B holds 10,000 shares
	treasuryB := types.NewCompanyTreasury(2)
	if treasuryB.InvestmentHoldings == nil {
		treasuryB.InvestmentHoldings = make(map[uint64]map[string]math.Int)
	}
	if treasuryB.InvestmentHoldings[1] == nil {
		treasuryB.InvestmentHoldings[1] = make(map[string]math.Int)
	}
	treasuryB.InvestmentHoldings[1]["COMMON"] = math.NewInt(10000)
	treasuryB.Balance = sdk.NewCoins()
	suite.equityKeeper.SetCompanyTreasury(suite.ctx, treasuryB)

	// Company C (ID 3) holds 5,000 shares
	companyC := types.NewCompany(
		3,
		"Company C",
		"CMPC",
		"Third test company",
		suite.cofounder.String(),
		"https://companyc.com",
		"Healthcare",
		"US",
		math.NewInt(100000),
		math.NewInt(200000),
		math.LegacyNewDec(1),
		sdk.NewCoins(sdk.NewInt64Coin("uhodl", 100000)),
	)
	suite.equityKeeper.SetCompany(suite.ctx, companyC)

	treasuryC := types.NewCompanyTreasury(3)
	if treasuryC.InvestmentHoldings == nil {
		treasuryC.InvestmentHoldings = make(map[uint64]map[string]math.Int)
	}
	if treasuryC.InvestmentHoldings[1] == nil {
		treasuryC.InvestmentHoldings[1] = make(map[string]math.Int)
	}
	treasuryC.InvestmentHoldings[1]["COMMON"] = math.NewInt(5000)
	treasuryC.Balance = sdk.NewCoins()
	suite.equityKeeper.SetCompanyTreasury(suite.ctx, treasuryC)

	// WHEN: Company A pays dividend (1 HODL per share)
	dividendPerShare := math.LegacyNewDec(1)

	// Company B should get 10,000 HODL
	dividendB := sdk.NewCoins(sdk.NewCoin("uhodl", math.NewInt(10000)))
	err := suite.equityKeeper.CreditTreasuryDividend(suite.ctx, 2, 1, dividendB)
	require.NoError(suite.T(), err)

	// Company C should get 5,000 HODL
	dividendC := sdk.NewCoins(sdk.NewCoin("uhodl", math.NewInt(5000)))
	err = suite.equityKeeper.CreditTreasuryDividend(suite.ctx, 3, 1, dividendC)
	require.NoError(suite.T(), err)

	// THEN: Both treasuries have correct balances
	updatedTreasuryB, _ := suite.equityKeeper.GetCompanyTreasury(suite.ctx, 2)
	require.Equal(suite.T(), dividendB, updatedTreasuryB.Balance)

	updatedTreasuryC, _ := suite.equityKeeper.GetCompanyTreasury(suite.ctx, 3)
	require.Equal(suite.T(), dividendC, updatedTreasuryC.Balance)
	*/
}

// =============================================================================
// Test: Treasury State Validation
// =============================================================================

// TestTreasuryInvestmentHoldings_AvailableShares tests available share calculation
func (suite *TreasuryInvestmentTestSuite) TestTreasuryInvestmentHoldings_AvailableShares() {
	// This test validates CompanyTreasury.GetAvailableInvestmentShares logic

	// GIVEN: Treasury with investment holdings and locked equity
	treasury := types.NewCompanyTreasury(1)

	// Initialize maps
	if treasury.InvestmentHoldings == nil {
		treasury.InvestmentHoldings = make(map[uint64]map[string]math.Int)
	}
	if treasury.InvestmentHoldings[2] == nil {
		treasury.InvestmentHoldings[2] = make(map[string]math.Int)
	}
	if treasury.LockedEquity == nil {
		treasury.LockedEquity = make(map[uint64]map[string]math.Int)
	}
	if treasury.LockedEquity[2] == nil {
		treasury.LockedEquity[2] = make(map[string]math.Int)
	}

	// Total investment: 10,000 shares
	treasury.InvestmentHoldings[2]["COMMON"] = math.NewInt(10000)

	// Locked: 3,000 shares
	treasury.LockedEquity[2]["COMMON"] = math.NewInt(3000)

	// WHEN: Calculating available shares
	available := treasury.GetAvailableInvestmentShares(2, "COMMON")

	// THEN: Available = Total - Locked = 7,000
	require.Equal(suite.T(), math.NewInt(7000), available)
}

// TestTreasuryInvestmentHoldings_NoHoldings tests empty investment holdings
func (suite *TreasuryInvestmentTestSuite) TestTreasuryInvestmentHoldings_NoHoldings() {
	// GIVEN: Treasury with no investment holdings
	treasury := types.NewCompanyTreasury(1)

	// WHEN: Getting investment shares
	shares := treasury.GetInvestmentShares(2, "COMMON")

	// THEN: Returns zero
	require.Equal(suite.T(), math.ZeroInt(), shares)
}

// TestTreasuryInvestmentHoldings_GetAllInvestments tests getting all holdings
func (suite *TreasuryInvestmentTestSuite) TestTreasuryInvestmentHoldings_GetAllInvestments() {
	// GIVEN: Treasury with multiple investment holdings
	treasury := types.NewCompanyTreasury(1)

	if treasury.InvestmentHoldings == nil {
		treasury.InvestmentHoldings = make(map[uint64]map[string]math.Int)
	}

	// Company 2, COMMON: 5,000 shares
	if treasury.InvestmentHoldings[2] == nil {
		treasury.InvestmentHoldings[2] = make(map[string]math.Int)
	}
	treasury.InvestmentHoldings[2]["COMMON"] = math.NewInt(5000)

	// Company 2, PREFERRED: 2,000 shares
	treasury.InvestmentHoldings[2]["PREFERRED"] = math.NewInt(2000)

	// Company 3, COMMON: 3,000 shares
	if treasury.InvestmentHoldings[3] == nil {
		treasury.InvestmentHoldings[3] = make(map[string]math.Int)
	}
	treasury.InvestmentHoldings[3]["COMMON"] = math.NewInt(3000)

	// WHEN: Getting all investment holdings
	holdings := treasury.GetAllInvestmentHoldings()

	// THEN: All holdings are returned
	require.Len(suite.T(), holdings, 3)

	// Verify holdings
	expectedHoldings := map[string]math.Int{
		"2:COMMON":    math.NewInt(5000),
		"2:PREFERRED": math.NewInt(2000),
		"3:COMMON":    math.NewInt(3000),
	}

	for _, holding := range holdings {
		key := fmt.Sprintf("%d:%s", holding.TargetCompanyID, holding.ClassID)
		expectedShares, exists := expectedHoldings[key]
		require.True(suite.T(), exists, "Unexpected holding: %s", key)
		require.Equal(suite.T(), expectedShares, holding.Shares)
	}
}

// =============================================================================
// Test: Edge Cases
// =============================================================================

// TestEquityWithdrawal_ExceedsAvailable tests withdrawal request exceeding available shares
func (suite *TreasuryInvestmentTestSuite) TestEquityWithdrawal_ExceedsAvailable() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// GIVEN: Treasury has 10,000 total, 3,000 locked, 7,000 available
	treasuryB := types.NewCompanyTreasury(2)
	if treasuryB.InvestmentHoldings == nil {
		treasuryB.InvestmentHoldings = make(map[uint64]map[string]math.Int)
	}
	if treasuryB.InvestmentHoldings[1] == nil {
		treasuryB.InvestmentHoldings[1] = make(map[string]math.Int)
	}
	treasuryB.InvestmentHoldings[1]["COMMON"] = math.NewInt(10000)

	if treasuryB.LockedEquity == nil {
		treasuryB.LockedEquity = make(map[uint64]map[string]math.Int)
	}
	if treasuryB.LockedEquity[1] == nil {
		treasuryB.LockedEquity[1] = make(map[string]math.Int)
	}
	treasuryB.LockedEquity[1]["COMMON"] = math.NewInt(3000)
	suite.equityKeeper.SetCompanyTreasury(suite.ctx, treasuryB)

	// WHEN: Requesting withdrawal of 8,000 shares (exceeds available)
	_, err := suite.equityKeeper.RequestEquityWithdrawal(
		suite.ctx,
		2,
		1,
		"COMMON",
		math.NewInt(8000), // Exceeds 7,000 available
		suite.bob.String(),
		"Over withdrawal",
		suite.bob.String(),
	)

	// THEN: Request is rejected
	require.Error(suite.T(), err)
	require.ErrorIs(suite.T(), err, types.ErrInsufficientTreasuryShares)
	*/
}

// TestDepositEquityToTreasury_ZeroShares tests deposit with zero shares
func (suite *TreasuryInvestmentTestSuite) TestDepositEquityToTreasury_ZeroShares() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// WHEN: Attempting to deposit zero shares
	err := suite.equityKeeper.DepositEquityToTreasury(
		suite.ctx,
		2,
		suite.cofounder.String(),
		1,
		"COMMON",
		math.ZeroInt(), // Zero shares
	)

	// THEN: Deposit is rejected
	require.Error(suite.T(), err)
	require.ErrorIs(suite.T(), err, types.ErrInsufficientShares)
	*/
}

// TestDepositEquityToTreasury_NonexistentCompany tests deposit to non-existent company
func (suite *TreasuryInvestmentTestSuite) TestDepositEquityToTreasury_NonexistentCompany() {
	suite.T().Skip("Requires keeper initialization")

	/*
	// WHEN: Attempting to deposit to non-existent target company
	err := suite.equityKeeper.DepositEquityToTreasury(
		suite.ctx,
		2,
		suite.cofounder.String(),
		999,                // Non-existent company
		"COMMON",
		math.NewInt(1000),
	)

	// THEN: Deposit is rejected
	require.Error(suite.T(), err)
	require.ErrorIs(suite.T(), err, types.ErrCompanyNotFound)
	*/
}

// =============================================================================
// Test Suite Runner
// =============================================================================

// TestTreasuryInvestmentTestSuite runs the test suite
func TestTreasuryInvestmentTestSuite(t *testing.T) {
	suite.Run(t, new(TreasuryInvestmentTestSuite))
}

// =============================================================================
// Additional Unit Tests (without suite)
// =============================================================================

// TestEquityWithdrawalRequest_Validate tests validation of withdrawal requests
func TestEquityWithdrawalRequest_Validate(t *testing.T) {
	alice := sdk.AccAddress([]byte("alice_______________")).String()

	tests := []struct {
		name        string
		request     types.EquityWithdrawalRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: types.NewEquityWithdrawalRequest(
				1,
				2,
				1,
				"COMMON",
				math.NewInt(1000),
				alice,
				"Test withdrawal",
				alice,
			),
			expectError: false,
		},
		{
			name: "zero company ID",
			request: types.EquityWithdrawalRequest{
				ID:              1,
				CompanyID:       0, // Invalid
				TargetCompanyID: 1,
				ClassID:         "COMMON",
				Shares:          math.NewInt(1000),
				Recipient:       alice,
				Purpose:         "Test",
				RequestedBy:     alice,
			},
			expectError: true,
			errorMsg:    "company ID cannot be zero",
		},
		{
			name: "zero target company ID",
			request: types.EquityWithdrawalRequest{
				ID:              1,
				CompanyID:       2,
				TargetCompanyID: 0, // Invalid
				ClassID:         "COMMON",
				Shares:          math.NewInt(1000),
				Recipient:       alice,
				Purpose:         "Test",
				RequestedBy:     alice,
			},
			expectError: true,
			errorMsg:    "target company ID cannot be zero",
		},
		{
			name: "empty class ID",
			request: types.EquityWithdrawalRequest{
				ID:              1,
				CompanyID:       2,
				TargetCompanyID: 1,
				ClassID:         "", // Invalid
				Shares:          math.NewInt(1000),
				Recipient:       alice,
				Purpose:         "Test",
				RequestedBy:     alice,
			},
			expectError: true,
			errorMsg:    "class ID cannot be empty",
		},
		{
			name: "zero shares",
			request: types.EquityWithdrawalRequest{
				ID:              1,
				CompanyID:       2,
				TargetCompanyID: 1,
				ClassID:         "COMMON",
				Shares:          math.ZeroInt(), // Invalid
				Recipient:       alice,
				Purpose:         "Test",
				RequestedBy:     alice,
			},
			expectError: true,
			errorMsg:    "shares must be positive",
		},
		{
			name: "empty purpose",
			request: types.EquityWithdrawalRequest{
				ID:              1,
				CompanyID:       2,
				TargetCompanyID: 1,
				ClassID:         "COMMON",
				Shares:          math.NewInt(1000),
				Recipient:       alice,
				Purpose:         "", // Invalid
				RequestedBy:     alice,
			},
			expectError: true,
			errorMsg:    "purpose cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

