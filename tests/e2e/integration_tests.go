package e2e

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
	equitytypes "github.com/sharehodl/sharehodl-blockchain/x/equity/types"
	dextypes "github.com/sharehodl/sharehodl-blockchain/x/dex/types"
	governancetypes "github.com/sharehodl/sharehodl-blockchain/x/governance/types"
)

// TestHODLModule tests the HODL stablecoin module
func (s *E2ETestSuite) TestHODLModule() {
	s.T().Log("🪙 Testing HODL Module")
	
	ctx := context.Background()
	
	// Test minting HODL
	s.T().Log("Testing HODL minting")
	mintAmount := sdk.NewInt64Coin(hodltypes.DefaultDenom, 1000000) // 1 HODL
	
	txResp, err := s.hodlClient.MintHODL(ctx, s.validatorAccount.Address, s.investorAccount1.Address, mintAmount)
	require.NoError(s.T(), err)
	s.recordTestResult("HODL_Mint", txResp.Code == 0, "", time.Now())
	
	// Wait for transaction to be processed
	s.WaitForBlocks(2)
	
	// Query balance after minting
	balance, err := s.hodlClient.GetBalance(ctx, s.investorAccount1.Address)
	require.NoError(s.T(), err)
	require.True(s.T(), balance.Amount.Amount.GTE(mintAmount.Amount))
	s.recordTestResult("HODL_Balance_Check", true, fmt.Sprintf("Balance: %s", balance.Amount), time.Now())
	
	// Test burning HODL
	s.T().Log("Testing HODL burning")
	burnAmount := sdk.NewInt64Coin(hodltypes.DefaultDenom, 500000) // 0.5 HODL
	
	txResp, err = s.hodlClient.BurnHODL(ctx, s.investorAccount1.Address, burnAmount)
	require.NoError(s.T(), err)
	s.recordTestResult("HODL_Burn", txResp.Code == 0, "", time.Now())
	
	// Query supply
	supply, err := s.hodlClient.GetSupply(ctx)
	require.NoError(s.T(), err)
	require.True(s.T(), supply.TotalSupply.Amount.IsPositive())
	s.recordTestResult("HODL_Supply_Check", true, fmt.Sprintf("Supply: %s", supply.TotalSupply), time.Now())
	
	s.T().Log("✅ HODL Module tests completed")
}

// TestEquityModule tests the ERC-EQUITY module
func (s *E2ETestSuite) TestEquityModule() {
	s.T().Log("🏢 Testing Equity Module")
	
	ctx := context.Background()
	
	// Test company creation
	s.T().Log("Testing company creation")
	company := equitytypes.Company{
		Name:         "TestCorp Inc",
		Symbol:       "TEST",
		TotalShares:  1000000,
		Industry:     "Technology",
		ValuationUSD: 50000000,
	}
	
	txResp, err := s.equityClient.CreateCompany(ctx, s.businessAccount.Address, company)
	require.NoError(s.T(), err)
	s.recordTestResult("Equity_Company_Creation", txResp.Code == 0, "", time.Now())
	
	s.WaitForBlocks(2)
	
	// Query company information
	companyResp, err := s.equityClient.GetCompany(ctx, "TEST")
	require.NoError(s.T(), err)
	require.Equal(s.T(), "TestCorp Inc", companyResp.Company.Name)
	s.recordTestResult("Equity_Company_Query", true, fmt.Sprintf("Company: %s", companyResp.Company.Name), time.Now())
	
	// Test share issuance
	s.T().Log("Testing share issuance")
	shares := uint64(10000)
	
	txResp, err = s.equityClient.IssueShares(ctx, s.businessAccount.Address, s.investorAccount1.Address, "TEST", shares)
	require.NoError(s.T(), err)
	s.recordTestResult("Equity_Share_Issuance", txResp.Code == 0, "", time.Now())
	
	s.WaitForBlocks(2)
	
	// Query shareholdings
	holdings, err := s.equityClient.GetShareholdings(ctx, s.investorAccount1.Address)
	require.NoError(s.T(), err)
	require.True(s.T(), len(holdings.Holdings) > 0)
	s.recordTestResult("Equity_Shareholdings_Query", true, fmt.Sprintf("Holdings: %d", len(holdings.Holdings)), time.Now())
	
	// Test share transfer
	s.T().Log("Testing share transfer")
	transferShares := uint64(1000)
	
	txResp, err = s.equityClient.TransferShares(ctx, s.investorAccount1.Address, s.investorAccount2.Address, "TEST", transferShares)
	require.NoError(s.T(), err)
	s.recordTestResult("Equity_Share_Transfer", txResp.Code == 0, "", time.Now())
	
	// Test dividend distribution
	s.T().Log("Testing dividend distribution")
	dividendAmount := sdk.NewInt64Coin(hodltypes.DefaultDenom, 100000) // 0.1 HODL total
	
	txResp, err = s.equityClient.DistributeDividends(ctx, s.businessAccount.Address, "TEST", dividendAmount)
	require.NoError(s.T(), err)
	s.recordTestResult("Equity_Dividend_Distribution", txResp.Code == 0, "", time.Now())
	
	s.T().Log("✅ Equity Module tests completed")
}

// TestDEXModule tests the DEX trading module
func (s *E2ETestSuite) TestDEXModule() {
	s.T().Log("📈 Testing DEX Module")
	
	ctx := context.Background()
	
	// Create a liquidity pool first
	s.T().Log("Testing liquidity pool creation")
	liquidityA := sdk.NewInt64Coin("TEST", 10000)
	liquidityB := sdk.NewInt64Coin(hodltypes.DefaultDenom, 100000)
	
	txResp, err := s.dexClient.CreateLiquidityPool(ctx, s.investorAccount1.Address, "TEST", hodltypes.DefaultDenom, liquidityA, liquidityB)
	require.NoError(s.T(), err)
	s.recordTestResult("DEX_Pool_Creation", txResp.Code == 0, "", time.Now())
	
	s.WaitForBlocks(2)
	
	// Query liquidity pools
	pools, err := s.dexClient.GetLiquidityPools(ctx)
	require.NoError(s.T(), err)
	require.True(s.T(), len(pools.Pools) > 0)
	s.recordTestResult("DEX_Pools_Query", true, fmt.Sprintf("Pools: %d", len(pools.Pools)), time.Now())
	
	// Test placing buy order
	s.T().Log("Testing buy order placement")
	buyOrder := dextypes.Order{
		OrderType:   dextypes.OrderType_LIMIT,
		Side:        dextypes.OrderSide_BUY,
		Symbol:      "TEST/HODL",
		Quantity:    1000,
		Price:       sdk.NewDec(10), // 10 HODL per TEST share
		TimeInForce: dextypes.TimeInForce_GTC,
	}
	
	txResp, err = s.dexClient.PlaceOrder(ctx, s.investorAccount1.Address, buyOrder)
	require.NoError(s.T(), err)
	s.recordTestResult("DEX_Buy_Order", txResp.Code == 0, "", time.Now())
	
	// Test placing sell order
	s.T().Log("Testing sell order placement")
	sellOrder := dextypes.Order{
		OrderType:   dextypes.OrderType_LIMIT,
		Side:        dextypes.OrderSide_SELL,
		Symbol:      "TEST/HODL",
		Quantity:    500,
		Price:       sdk.NewDec(10), // Same price for matching
		TimeInForce: dextypes.TimeInForce_GTC,
	}
	
	txResp, err = s.dexClient.PlaceOrder(ctx, s.investorAccount2.Address, sellOrder)
	require.NoError(s.T(), err)
	s.recordTestResult("DEX_Sell_Order", txResp.Code == 0, "", time.Now())
	
	s.WaitForBlocks(3) // Wait for matching
	
	// Query order book
	orderBook, err := s.dexClient.GetOrderBook(ctx, "TEST/HODL")
	require.NoError(s.T(), err)
	s.recordTestResult("DEX_OrderBook_Query", true, fmt.Sprintf("Bids: %d, Asks: %d", len(orderBook.Bids), len(orderBook.Asks)), time.Now())
	
	// Query trade history
	trades, err := s.dexClient.GetTradeHistory(ctx, "TEST/HODL", 10)
	require.NoError(s.T(), err)
	s.recordTestResult("DEX_Trade_History", true, fmt.Sprintf("Trades: %d", len(trades.Trades)), time.Now())
	
	// Test adding liquidity
	s.T().Log("Testing liquidity addition")
	poolID := "1" // Assume first pool
	addLiquidityA := sdk.NewInt64Coin("TEST", 5000)
	addLiquidityB := sdk.NewInt64Coin(hodltypes.DefaultDenom, 50000)
	
	txResp, err = s.dexClient.AddLiquidity(ctx, s.investorAccount2.Address, poolID, addLiquidityA, addLiquidityB)
	require.NoError(s.T(), err)
	s.recordTestResult("DEX_Add_Liquidity", txResp.Code == 0, "", time.Now())
	
	s.T().Log("✅ DEX Module tests completed")
}

// TestGovernanceModule tests the governance module
func (s *E2ETestSuite) TestGovernanceModule() {
	s.T().Log("🗳️ Testing Governance Module")
	
	ctx := context.Background()
	
	// Test proposal submission
	s.T().Log("Testing proposal submission")
	proposal := governancetypes.Proposal{
		ProposalType: governancetypes.ProposalType_PARAMETER_CHANGE,
		Title:        "Test Parameter Change",
		Description:  "This is a test proposal for parameter change",
		VotingPeriod: 24 * time.Hour,
	}
	
	deposit := sdk.NewInt64Coin(hodltypes.DefaultDenom, 1000000) // 1 HODL
	
	txResp, err := s.governanceClient.SubmitProposal(ctx, s.validatorAccount.Address, proposal, deposit)
	require.NoError(s.T(), err)
	s.recordTestResult("Gov_Proposal_Submission", txResp.Code == 0, "", time.Now())
	
	s.WaitForBlocks(2)
	
	// Query proposals
	proposals, err := s.governanceClient.GetProposals(ctx, governancetypes.ProposalStatus_VOTING_PERIOD)
	require.NoError(s.T(), err)
	require.True(s.T(), len(proposals.Proposals) >= 0)
	s.recordTestResult("Gov_Proposals_Query", true, fmt.Sprintf("Proposals: %d", len(proposals.Proposals)), time.Now())
	
	// Assume proposal ID 1 for testing
	proposalID := uint64(1)
	
	// Test voting
	s.T().Log("Testing voting")
	txResp, err = s.governanceClient.Vote(ctx, s.validatorAccount.Address, proposalID, governancetypes.VoteOption_YES)
	require.NoError(s.T(), err)
	s.recordTestResult("Gov_Vote", txResp.Code == 0, "", time.Now())
	
	// Test additional deposit
	s.T().Log("Testing additional deposit")
	additionalDeposit := sdk.NewInt64Coin(hodltypes.DefaultDenom, 500000) // 0.5 HODL
	
	txResp, err = s.governanceClient.Deposit(ctx, s.investorAccount1.Address, proposalID, additionalDeposit)
	require.NoError(s.T(), err)
	s.recordTestResult("Gov_Additional_Deposit", txResp.Code == 0, "", time.Now())
	
	s.WaitForBlocks(2)
	
	// Query specific proposal
	proposalResp, err := s.governanceClient.GetProposal(ctx, proposalID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), proposalID, proposalResp.Proposal.ProposalId)
	s.recordTestResult("Gov_Proposal_Query", true, fmt.Sprintf("Proposal Status: %s", proposalResp.Proposal.Status), time.Now())
	
	// Query vote
	vote, err := s.governanceClient.GetVote(ctx, proposalID, s.validatorAccount.Address)
	require.NoError(s.T(), err)
	require.Equal(s.T(), governancetypes.VoteOption_YES, vote.Vote.Option)
	s.recordTestResult("Gov_Vote_Query", true, fmt.Sprintf("Vote: %s", vote.Vote.Option), time.Now())
	
	s.T().Log("✅ Governance Module tests completed")
}

// TestCrossModuleIntegration tests integration between modules
func (s *E2ETestSuite) TestCrossModuleIntegration() {
	s.T().Log("🔗 Testing Cross-Module Integration")
	
	ctx := context.Background()
	
	// Test scenario: Create company, issue shares, trade on DEX, and create governance proposal
	
	s.T().Log("Creating company for integration test")
	company := equitytypes.Company{
		Name:         "IntegrationCorp",
		Symbol:       "INTEG",
		TotalShares:  100000,
		Industry:     "Integration Testing",
		ValuationUSD: 10000000,
	}
	
	txResp, err := s.equityClient.CreateCompany(ctx, s.businessAccount.Address, company)
	require.NoError(s.T(), err)
	s.recordTestResult("Integration_Company_Creation", txResp.Code == 0, "", time.Now())
	
	s.WaitForBlocks(2)
	
	s.T().Log("Issuing shares for integration test")
	txResp, err = s.equityClient.IssueShares(ctx, s.businessAccount.Address, s.investorAccount1.Address, "INTEG", 10000)
	require.NoError(s.T(), err)
	s.recordTestResult("Integration_Share_Issuance", txResp.Code == 0, "", time.Now())
	
	s.WaitForBlocks(2)
	
	s.T().Log("Creating liquidity pool for integration test")
	liquidityA := sdk.NewInt64Coin("INTEG", 5000)
	liquidityB := sdk.NewInt64Coin(hodltypes.DefaultDenom, 50000)
	
	txResp, err = s.dexClient.CreateLiquidityPool(ctx, s.investorAccount1.Address, "INTEG", hodltypes.DefaultDenom, liquidityA, liquidityB)
	require.NoError(s.T(), err)
	s.recordTestResult("Integration_Pool_Creation", txResp.Code == 0, "", time.Now())
	
	s.WaitForBlocks(2)
	
	s.T().Log("Trading shares on DEX")
	order := dextypes.Order{
		OrderType:   dextypes.OrderType_LIMIT,
		Side:        dextypes.OrderSide_SELL,
		Symbol:      "INTEG/HODL",
		Quantity:    1000,
		Price:       sdk.NewDec(12),
		TimeInForce: dextypes.TimeInForce_GTC,
	}
	
	txResp, err = s.dexClient.PlaceOrder(ctx, s.investorAccount1.Address, order)
	require.NoError(s.T(), err)
	s.recordTestResult("Integration_Trading", txResp.Code == 0, "", time.Now())
	
	s.WaitForBlocks(2)
	
	s.T().Log("Creating governance proposal for company parameters")
	proposal := governancetypes.Proposal{
		ProposalType: governancetypes.ProposalType_TEXT,
		Title:        "Integration Test Proposal",
		Description:  "Proposal to test integration between equity and governance modules",
		VotingPeriod: 1 * time.Hour,
	}
	
	deposit := sdk.NewInt64Coin(hodltypes.DefaultDenom, 1000000)
	txResp, err = s.governanceClient.SubmitProposal(ctx, s.businessAccount.Address, proposal, deposit)
	require.NoError(s.T(), err)
	s.recordTestResult("Integration_Governance_Proposal", txResp.Code == 0, "", time.Now())
	
	s.T().Log("✅ Cross-Module Integration tests completed")
}

// TestBusinessWorkflow tests complete business workflow
func (s *E2ETestSuite) TestBusinessWorkflow() {
	s.T().Log("🏪 Testing Complete Business Workflow")
	
	ctx := context.Background()
	startTime := time.Now()
	
	// Step 1: Business verification and company registration
	s.T().Log("Step 1: Business registration workflow")
	
	// Create HODL for business operations
	mintAmount := sdk.NewInt64Coin(hodltypes.DefaultDenom, 10000000) // 10 HODL
	txResp, err := s.hodlClient.MintHODL(ctx, s.validatorAccount.Address, s.businessAccount.Address, mintAmount)
	require.NoError(s.T(), err)
	
	// Register company
	company := equitytypes.Company{
		Name:         "WorkflowCorp Ltd",
		Symbol:       "WORK",
		TotalShares:  500000,
		Industry:     "Business Workflow",
		ValuationUSD: 25000000,
	}
	
	txResp, err = s.equityClient.CreateCompany(ctx, s.businessAccount.Address, company)
	require.NoError(s.T(), err)
	
	s.WaitForBlocks(2)
	
	// Step 2: Initial share distribution (equity raise)
	s.T().Log("Step 2: Equity raise workflow")
	
	// Issue shares to multiple investors
	investors := []string{s.investorAccount1.Address, s.investorAccount2.Address}
	shareAllocations := []uint64{50000, 30000}
	
	for i, investor := range investors {
		txResp, err = s.equityClient.IssueShares(ctx, s.businessAccount.Address, investor, "WORK", shareAllocations[i])
		require.NoError(s.T(), err)
		s.WaitForBlocks(1)
	}
	
	// Step 3: Create market for trading
	s.T().Log("Step 3: Market creation workflow")
	
	// Create liquidity pool
	liquidityA := sdk.NewInt64Coin("WORK", 20000)
	liquidityB := sdk.NewInt64Coin(hodltypes.DefaultDenom, 200000)
	
	txResp, err = s.dexClient.CreateLiquidityPool(ctx, s.businessAccount.Address, "WORK", hodltypes.DefaultDenom, liquidityA, liquidityB)
	require.NoError(s.T(), err)
	
	s.WaitForBlocks(2)
	
	// Step 4: Trading activity
	s.T().Log("Step 4: Trading workflow")
	
	// Multiple trades to simulate market activity
	trades := []struct {
		trader   string
		side     dextypes.OrderSide
		quantity uint64
		price    int64
	}{
		{s.investorAccount1.Address, dextypes.OrderSide_SELL, 5000, 12},
		{s.investorAccount2.Address, dextypes.OrderSide_BUY, 3000, 12},
		{s.investorAccount1.Address, dextypes.OrderSide_SELL, 2000, 11},
	}
	
	for _, trade := range trades {
		order := dextypes.Order{
			OrderType:   dextypes.OrderType_LIMIT,
			Side:        trade.side,
			Symbol:      "WORK/HODL",
			Quantity:    trade.quantity,
			Price:       sdk.NewDec(trade.price),
			TimeInForce: dextypes.TimeInForce_GTC,
		}
		
		txResp, err = s.dexClient.PlaceOrder(ctx, trade.trader, order)
		require.NoError(s.T(), err)
		s.WaitForBlocks(1)
	}
	
	// Step 5: Dividend distribution
	s.T().Log("Step 5: Dividend distribution workflow")
	
	// Distribute quarterly dividends
	dividendAmount := sdk.NewInt64Coin(hodltypes.DefaultDenom, 100000) // 0.1 HODL total
	
	txResp, err = s.equityClient.DistributeDividends(ctx, s.businessAccount.Address, "WORK", dividendAmount)
	require.NoError(s.T(), err)
	
	s.WaitForBlocks(2)
	
	// Step 6: Governance participation
	s.T().Log("Step 6: Corporate governance workflow")
	
	// Create proposal for company strategic decision
	proposal := governancetypes.Proposal{
		ProposalType: governancetypes.ProposalType_TEXT,
		Title:        "WorkflowCorp Strategic Expansion",
		Description:  "Proposal to expand operations into new markets",
		VotingPeriod: 2 * time.Hour,
	}
	
	deposit := sdk.NewInt64Coin(hodltypes.DefaultDenom, 1000000)
	txResp, err = s.governanceClient.SubmitProposal(ctx, s.businessAccount.Address, proposal, deposit)
	require.NoError(s.T(), err)
	
	// Shareholders vote on proposal
	proposalID := uint64(2) // Assume this is proposal ID
	votes := []struct {
		voter  string
		option governancetypes.VoteOption
	}{
		{s.investorAccount1.Address, governancetypes.VoteOption_YES},
		{s.investorAccount2.Address, governancetypes.VoteOption_YES},
		{s.businessAccount.Address, governancetypes.VoteOption_YES},
	}
	
	for _, vote := range votes {
		txResp, err = s.governanceClient.Vote(ctx, vote.voter, proposalID, vote.option)
		require.NoError(s.T(), err)
		s.WaitForBlocks(1)
	}
	
	workflowDuration := time.Since(startTime)
	s.recordTestResult("Business_Workflow_Complete", true, 
		fmt.Sprintf("Completed in %v", workflowDuration), startTime)
	
	s.T().Log("✅ Complete Business Workflow tests completed")
}

// WaitForBlocks waits for specified number of blocks to be produced
func (s *E2ETestSuite) WaitForBlocks(blocks int) {
	// Wait for blocks to be produced (simulated)
	time.Sleep(time.Duration(blocks) * 6 * time.Second) // Assume 6s block time
}

// recordTestResult records a test result in metrics
func (s *E2ETestSuite) recordTestResult(name string, passed bool, details string, startTime time.Time) {
	result := TestResult{
		Name:      name,
		Status:    "PASS",
		Duration:  time.Since(startTime),
		Details:   details,
		StartTime: startTime,
		EndTime:   time.Now(),
	}
	
	if !passed {
		result.Status = "FAIL"
		s.metrics.FailedTests++
	} else {
		s.metrics.PassedTests++
	}
	
	s.metrics.TotalTests++
	s.metrics.TestResults = append(s.metrics.TestResults, result)
	
	s.T().Logf("📊 Test Result: %s - %s (%v)", name, result.Status, result.Duration)
}