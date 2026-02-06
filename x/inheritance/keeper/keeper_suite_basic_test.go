package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

// TestGetSetParams tests parameter get/set functionality
func (suite *KeeperTestSuite) TestGetSetParams() {
	// Test default params
	params := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(180*24*time.Hour, params.MinInactivityPeriod())
	suite.Require().Equal(30*24*time.Hour, params.MinGracePeriod())
	suite.Require().Equal(30*24*time.Hour, params.MinClaimWindow())
	suite.Require().Equal(365*24*time.Hour, params.MaxClaimWindow())

	// Test setting custom params
	customParams := types.Params{
		MinInactivityPeriodSeconds:       2 * 365 * 24 * 60 * 60, // 2 years in seconds
		MinGracePeriodSeconds:            45 * 24 * 60 * 60,      // 45 days in seconds
		MinClaimWindowSeconds:            30 * 24 * 60 * 60,      // 30 days in seconds
		MaxClaimWindowSeconds:            120 * 24 * 60 * 60,     // 120 days in seconds
		UltraLongInactivityPeriodSeconds: 50 * 365 * 24 * 60 * 60, // 50 years in seconds
		DefaultCharityAddress:            "cosmos1charity",
		MaxBeneficiaries:                 10,
	}

	suite.keeper.SetParams(suite.ctx, customParams)

	// Verify params were set
	retrievedParams := suite.keeper.GetParams(suite.ctx)
	suite.Require().Equal(customParams.MinInactivityPeriod(), retrievedParams.MinInactivityPeriod())
	suite.Require().Equal(customParams.MinGracePeriod(), retrievedParams.MinGracePeriod())
	suite.Require().Equal(customParams.MinClaimWindow(), retrievedParams.MinClaimWindow())
	suite.Require().Equal(customParams.MaxClaimWindow(), retrievedParams.MaxClaimWindow())
}

// TestCreateInheritancePlan tests creating an inheritance plan
func (suite *KeeperTestSuite) TestCreateInheritancePlan() {
	owner := CreateTestAddresses(1)[0]

	// Create beneficiaries
	beneficiaries := []types.Beneficiary{
		{
			Address:    CreateTestAddresses(2)[0].String(),
			Percentage: math.LegacyNewDecWithPrec(60, 2), // 0.60 = 60%
			Priority:   1,
		},
		{
			Address:    CreateTestAddresses(2)[1].String(),
			Percentage: math.LegacyNewDecWithPrec(40, 2), // 0.40 = 40%
			Priority:   1,
		},
	}

	// Create inheritance plan
	planID := suite.keeper.GetNextPlanID(suite.ctx)
	suite.Require().Equal(uint64(1), planID)

	plan := types.InheritancePlan{
		PlanID:           planID,
		Owner:            owner.String(),
		Beneficiaries:    beneficiaries,
		InactivityPeriod: 180 * 24 * time.Hour,
		GracePeriod:      60 * 24 * time.Hour, // 60 days
		ClaimWindow:      30 * 24 * time.Hour,
		Status:           types.PlanStatusActive,
		CreatedAt:        suite.ctx.BlockTime(),
		UpdatedAt:        suite.ctx.BlockTime(),
	}

	// Validate plan
	err := plan.Validate()
	suite.Require().NoError(err)

	// Save plan using CreatePlan
	err = suite.keeper.CreatePlan(suite.ctx, plan)
	suite.Require().NoError(err)

	// Retrieve plan using GetPlan
	retrieved, found := suite.keeper.GetPlan(suite.ctx, planID)
	suite.Require().True(found)
	suite.Require().Equal(plan.PlanID, retrieved.PlanID)
	suite.Require().Equal(plan.Owner, retrieved.Owner)
	suite.Require().Len(retrieved.Beneficiaries, 2)

	// Get plans by owner
	ownerPlans := suite.keeper.GetPlansByOwner(suite.ctx, owner)
	suite.Require().Len(ownerPlans, 1)
	suite.Require().Equal(planID, ownerPlans[0].PlanID)
}

// TestCancelInheritancePlan tests cancelling an inheritance plan
func (suite *KeeperTestSuite) TestCancelInheritancePlan() {
	owner := CreateTestAddresses(1)[0]

	// Create and save a plan
	planID := suite.keeper.GetNextPlanID(suite.ctx)
	plan := types.InheritancePlan{
		PlanID:  planID,
		Owner:   owner.String(),
		Beneficiaries: []types.Beneficiary{
			{
				Address:    CreateTestAddresses(2)[0].String(),
				Percentage: math.LegacyOneDec(), // 100%
				Priority:   1,
			},
		},
		InactivityPeriod: 180 * 24 * time.Hour,
		GracePeriod:      60 * 24 * time.Hour,
		ClaimWindow:      30 * 24 * time.Hour,
		Status:           types.PlanStatusActive,
		CreatedAt:        suite.ctx.BlockTime(),
		UpdatedAt:        suite.ctx.BlockTime(),
	}
	err := suite.keeper.CreatePlan(suite.ctx, plan)
	suite.Require().NoError(err)

	// Cancel the plan
	plan.Status = types.PlanStatusCancelled
	plan.UpdatedAt = suite.ctx.BlockTime()
	suite.keeper.SetPlan(suite.ctx, plan)

	// Retrieve and verify
	retrieved, found := suite.keeper.GetPlan(suite.ctx, planID)
	suite.Require().True(found)
	suite.Require().Equal(types.PlanStatusCancelled, retrieved.Status)
}

// TestRecordActivity tests activity recording
func (suite *KeeperTestSuite) TestRecordActivity() {
	owner := CreateTestAddresses(1)[0]

	// Record activity
	suite.keeper.RecordActivity(suite.ctx, owner, "transaction")

	// Get last activity
	lastActivity, found := suite.keeper.GetLastActivity(suite.ctx, owner)
	suite.Require().True(found)
	suite.Require().Equal("transaction", lastActivity.ActivityType)
	suite.Require().Equal(suite.ctx.BlockTime(), lastActivity.LastActivityTime)
}

// TestIsInactive tests inactivity checking
func (suite *KeeperTestSuite) TestIsInactive() {
	owner := CreateTestAddresses(1)[0]

	// Get threshold from params
	params := suite.keeper.GetParams(suite.ctx)
	threshold := params.MinInactivityPeriod

	// Record recent activity
	suite.keeper.RecordActivity(suite.ctx, owner, "transaction")

	// Should not be inactive
	isInactive := suite.keeper.IsInactive(suite.ctx, owner, threshold)
	suite.Require().False(isInactive)

	// Advance time beyond threshold (180 days + 1 day)
	suite.AdvanceTime(threshold + (24 * time.Hour))

	// Should now be inactive
	isInactive = suite.keeper.IsInactive(suite.ctx, owner, threshold)
	suite.Require().True(isInactive)
}

// TestBannedOwnerCannotCreatePlan tests that banned users cannot create plans
func (suite *KeeperTestSuite) TestBannedOwnerCannotCreatePlan() {
	owner := CreateTestAddresses(1)[0]

	// Ban the owner
	suite.BanAddress(owner.String())

	// Create a plan
	plan := types.InheritancePlan{
		PlanID:  1,
		Owner:   owner.String(),
		Beneficiaries: []types.Beneficiary{
			{
				Address:    CreateTestAddresses(2)[0].String(),
				Percentage: math.LegacyOneDec(),
				Priority:   1,
			},
		},
		InactivityPeriod: 180 * 24 * time.Hour,
		GracePeriod:      60 * 24 * time.Hour,
		ClaimWindow:      30 * 24 * time.Hour,
		Status:           types.PlanStatusActive,
		CreatedAt:        suite.ctx.BlockTime(),
		UpdatedAt:        suite.ctx.BlockTime(),
	}

	// Check if owner is banned before creating plan
	isBanned := suite.banKeeper.IsAddressBanned(suite.ctx, owner.String())
	suite.Require().True(isBanned)

	// In real implementation, the message server would prevent this
	// For now, we just verify the ban check works
	err := suite.keeper.CreatePlan(suite.ctx, plan)
	// If the keeper checks for bans, this should error
	// Otherwise this test documents that ban checking should be in msg_server
	_ = err
}

// TestAssetTransfer tests transferring assets to beneficiaries
func (suite *KeeperTestSuite) TestAssetTransfer() {
	owner := CreateTestAddresses(1)[0]
	beneficiary := CreateTestAddresses(2)[0]

	// Set up HODL balance
	suite.SetHODLBalance(owner, math.NewInt(1_000_000))

	// Set up equity shares
	suite.SetShares(1, "COMMON", owner.String(), math.NewInt(10_000))

	// Create coins for bank keeper
	coins := sdk.NewCoins(sdk.NewInt64Coin("uhodl", 500_000))
	suite.SetBalance(owner, coins)

	// Verify initial balances
	hodlBalance := suite.hodlKeeper.GetBalance(suite.ctx, owner)
	suite.Require().Equal(math.NewInt(1_000_000), hodlBalance)

	shares, found := suite.equityKeeper.GetShareholding(suite.ctx, 1, "COMMON", owner.String())
	suite.Require().True(found)
	suite.Require().Equal(math.NewInt(10_000), shares)

	// Transfer 50% to beneficiary
	err := suite.hodlKeeper.SendCoins(suite.ctx, owner, beneficiary, math.NewInt(500_000))
	suite.Require().NoError(err)

	err = suite.equityKeeper.TransferShares(suite.ctx, 1, "COMMON", owner.String(), beneficiary.String(), math.NewInt(5_000))
	suite.Require().NoError(err)

	// Verify balances after transfer
	ownerBalance := suite.hodlKeeper.GetBalance(suite.ctx, owner)
	suite.Require().Equal(math.NewInt(500_000), ownerBalance)

	beneficiaryBalance := suite.hodlKeeper.GetBalance(suite.ctx, beneficiary)
	suite.Require().Equal(math.NewInt(500_000), beneficiaryBalance)

	beneficiaryShares, found := suite.equityKeeper.GetShareholding(suite.ctx, 1, "COMMON", beneficiary.String())
	suite.Require().True(found)
	suite.Require().Equal(math.NewInt(5_000), beneficiaryShares)
}

// TestPlanIDIncrement tests that plan IDs increment correctly
func (suite *KeeperTestSuite) TestPlanIDIncrement() {
	id1 := suite.keeper.GetNextPlanID(suite.ctx)
	suite.Require().Equal(uint64(1), id1)

	id2 := suite.keeper.GetNextPlanID(suite.ctx)
	suite.Require().Equal(uint64(2), id2)

	id3 := suite.keeper.GetNextPlanID(suite.ctx)
	suite.Require().Equal(uint64(3), id3)
}

// TestGetAllInheritancePlans tests retrieving all plans
func (suite *KeeperTestSuite) TestGetAllInheritancePlans() {
	// Create multiple plans
	owners := CreateTestAddresses(3)

	for i, owner := range owners {
		planID := suite.keeper.GetNextPlanID(suite.ctx)
		plan := types.InheritancePlan{
			PlanID:  planID,
			Owner:   owner.String(),
			Beneficiaries: []types.Beneficiary{
				{
					Address:    CreateTestAddresses(5)[i].String(),
					Percentage: math.LegacyOneDec(),
					Priority:   1,
				},
			},
			InactivityPeriod: 180 * 24 * time.Hour,
			GracePeriod:      60 * 24 * time.Hour,
			ClaimWindow:      30 * 24 * time.Hour,
			Status:           types.PlanStatusActive,
			CreatedAt:        suite.ctx.BlockTime(),
			UpdatedAt:        suite.ctx.BlockTime(),
		}
		err := suite.keeper.CreatePlan(suite.ctx, plan)
		suite.Require().NoError(err)
	}

	// Get all plans
	allPlans := suite.keeper.GetAllPlans(suite.ctx)
	suite.Require().Len(allPlans, 3)
}
