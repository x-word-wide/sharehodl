package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

var (
	// Pre-computed valid test addresses
	testOwner        = sdk.AccAddress([]byte("inheritance_owner___")).String()
	testBeneficiary1 = sdk.AccAddress([]byte("beneficiary_one_____")).String()
	testBeneficiary2 = sdk.AccAddress([]byte("beneficiary_two_____")).String()
	testBeneficiary3 = sdk.AccAddress([]byte("beneficiary_three___")).String()
	testCharity      = sdk.AccAddress([]byte("charity_wallet______")).String()
)

// TestCreateInheritancePlan tests creating a new inheritance plan
func TestCreateInheritancePlan(t *testing.T) {
	t.Skip("Requires full keeper setup")

	// Test cases to implement:
	// 1. Create valid plan with 1 year inactivity period
	// 2. Create valid plan with multiple beneficiaries
	// 3. Create valid plan with specific asset allocations
	// 4. Reject plan with invalid owner address
	// 5. Reject plan with no beneficiaries
	// 6. Reject plan with grace period < 30 days
	// 7. Reject plan with invalid charity address
	// 8. Reject plan with beneficiary percentages > 100%
	// 9. Verify plan is stored correctly
	// 10. Verify plan is indexed by owner
	// 11. Verify plan is indexed by beneficiaries
}

// TestUpdateInheritancePlan tests updating an existing plan
func TestUpdateInheritancePlan(t *testing.T) {
	t.Skip("Requires full keeper setup")

	// Test cases to implement:
	// 1. Update beneficiaries
	// 2. Update inactivity period
	// 3. Update grace period (must remain >= 30 days)
	// 4. Update charity address
	// 5. Reject update from non-owner
	// 6. Reject update of triggered plan
	// 7. Reject update of completed plan
	// 8. Reject update with invalid percentages
	// 9. Verify indexes are updated correctly
	// 10. Verify UpdatedAt timestamp is set
}

// TestCancelInheritancePlan tests cancelling a plan
func TestCancelInheritancePlan(t *testing.T) {
	t.Skip("Requires full keeper setup")

	// Test cases to implement:
	// 1. Cancel active plan by owner
	// 2. Reject cancellation from non-owner
	// 3. Reject cancellation of already cancelled plan
	// 4. Reject cancellation of completed plan
	// 5. Allow cancellation of triggered plan (during grace period)
	// 6. Verify plan status changes to cancelled
	// 7. Verify indexes are cleaned up
	// 8. Verify locked assets are released
}

// TestGetPlansByOwner tests querying plans by owner
func TestGetPlansByOwner(t *testing.T) {
	t.Skip("Requires full keeper setup")

	// Test cases to implement:
	// 1. Get all plans for owner with multiple plans
	// 2. Get plans for owner with no plans
	// 3. Verify pagination works correctly
	// 4. Verify only owner's plans are returned
	// 5. Verify cancelled plans are included
	// 6. Verify completed plans are included
}

// TestGetPlansByBeneficiary tests querying plans by beneficiary
func TestGetPlansByBeneficiary(t *testing.T) {
	t.Skip("Requires full keeper setup")

	// Test cases to implement:
	// 1. Get all plans where address is beneficiary
	// 2. Get plans for address that is not a beneficiary
	// 3. Verify address can be beneficiary in multiple plans
	// 4. Verify only active/triggered plans are returned
	// 5. Verify cancelled plans are excluded
	// 6. Verify completed plans are included
}

// TestPlanValidation tests validation of plan parameters
func TestPlanValidation(t *testing.T) {
	tests := []struct {
		name    string
		plan    types.InheritancePlan
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid plan",
			plan: types.InheritancePlan{
				PlanID: 1,
				Owner:  testOwner,
				Beneficiaries: []types.Beneficiary{
					{
						Address:    testBeneficiary1,
						Priority:   1,
						Percentage: math.LegacyNewDecWithPrec(50, 2), // 0.50
					},
					{
						Address:    testBeneficiary2,
						Priority:   2,
						Percentage: math.LegacyNewDecWithPrec(50, 2), // 0.50
					},
				},
				InactivityPeriod: 365 * 24 * time.Hour,
				GracePeriod:      30 * 24 * time.Hour,
				ClaimWindow:      180 * 24 * time.Hour,
				CharityAddress:   testCharity,
				Status:           types.PlanStatusActive,
			},
			wantErr: false,
		},
		{
			name: "invalid - no beneficiaries",
			plan: types.InheritancePlan{
				PlanID:           1,
				Owner:            testOwner,
				Beneficiaries:    []types.Beneficiary{},
				InactivityPeriod: 365 * 24 * time.Hour,
				GracePeriod:      30 * 24 * time.Hour,
				ClaimWindow:      180 * 24 * time.Hour,
			},
			wantErr: true,
			errMsg:  "must have at least one beneficiary",
		},
		{
			name: "invalid - percentages exceed 100%",
			plan: types.InheritancePlan{
				PlanID: 1,
				Owner:  testOwner,
				Beneficiaries: []types.Beneficiary{
					{
						Address:    testBeneficiary1,
						Priority:   1,
						Percentage: math.LegacyNewDecWithPrec(60, 2), // 0.60
					},
					{
						Address:    testBeneficiary2,
						Priority:   2,
						Percentage: math.LegacyNewDecWithPrec(60, 2), // 0.60
					},
				},
				InactivityPeriod: 365 * 24 * time.Hour,
				GracePeriod:      30 * 24 * time.Hour,
				ClaimWindow:      180 * 24 * time.Hour,
			},
			wantErr: true,
			errMsg:  "total beneficiary percentage must equal 100%", // Security fix: must be exactly 100%
		},
		{
			name: "invalid - grace period too short",
			plan: types.InheritancePlan{
				PlanID: 1,
				Owner:  testOwner,
				Beneficiaries: []types.Beneficiary{
					{
						Address:    testBeneficiary1,
						Priority:   1,
						Percentage: math.LegacyOneDec(),
					},
				},
				InactivityPeriod: 365 * 24 * time.Hour,
				GracePeriod:      7 * 24 * time.Hour, // Only 7 days
				ClaimWindow:      180 * 24 * time.Hour,
			},
			wantErr: true,
			errMsg:  "grace period must be at least 30 days",
		},
		{
			name: "invalid - duplicate priorities",
			plan: types.InheritancePlan{
				PlanID: 1,
				Owner:  testOwner,
				Beneficiaries: []types.Beneficiary{
					{
						Address:    testBeneficiary1,
						Priority:   1,
						Percentage: math.LegacyNewDecWithPrec(50, 2),
					},
					{
						Address:    testBeneficiary2,
						Priority:   1, // Duplicate priority
						Percentage: math.LegacyNewDecWithPrec(50, 2),
					},
				},
				InactivityPeriod: 365 * 24 * time.Hour,
				GracePeriod:      30 * 24 * time.Hour,
				ClaimWindow:      180 * 24 * time.Hour,
			},
			wantErr: true,
			errMsg:  "duplicate priority",
		},
		{
			name: "invalid - percentages under 100% (must be exactly 100%)",
			plan: types.InheritancePlan{
				PlanID: 1,
				Owner:  testOwner,
				Beneficiaries: []types.Beneficiary{
					{
						Address:    testBeneficiary1,
						Priority:   1,
						Percentage: math.LegacyNewDecWithPrec(30, 2), // 0.30
					},
					{
						Address:    testBeneficiary2,
						Priority:   2,
						Percentage: math.LegacyNewDecWithPrec(20, 2), // 0.20
					},
				},
				InactivityPeriod: 365 * 24 * time.Hour,
				GracePeriod:      30 * 24 * time.Hour,
				ClaimWindow:      180 * 24 * time.Hour,
				CharityAddress:   testCharity,
			},
			wantErr: true,
			errMsg:  "total beneficiary percentage must equal 100%", // Security: must be exactly 100%
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.plan.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestBeneficiaryValidation tests beneficiary validation
func TestBeneficiaryValidation(t *testing.T) {
	tests := []struct {
		name        string
		beneficiary types.Beneficiary
		wantErr     bool
		errMsg      string
	}{
		{
			name: "valid beneficiary",
			beneficiary: types.Beneficiary{
				Address:    testBeneficiary1,
				Priority:   1,
				Percentage: math.LegacyNewDecWithPrec(50, 2),
			},
			wantErr: false,
		},
		{
			name: "invalid - empty address",
			beneficiary: types.Beneficiary{
				Address:    "",
				Priority:   1,
				Percentage: math.LegacyNewDecWithPrec(50, 2),
			},
			wantErr: true,
			errMsg:  "beneficiary address cannot be empty",
		},
		{
			name: "invalid - zero priority",
			beneficiary: types.Beneficiary{
				Address:    testBeneficiary1,
				Priority:   0,
				Percentage: math.LegacyNewDecWithPrec(50, 2),
			},
			wantErr: true,
			errMsg:  "priority must be positive",
		},
		{
			name: "invalid - negative percentage",
			beneficiary: types.Beneficiary{
				Address:    testBeneficiary1,
				Priority:   1,
				Percentage: math.LegacyNewDecWithPrec(-10, 2),
			},
			wantErr: true,
			errMsg:  "percentage must be between 0 and 1",
		},
		{
			name: "invalid - percentage over 100%",
			beneficiary: types.Beneficiary{
				Address:    testBeneficiary1,
				Priority:   1,
				Percentage: math.LegacyNewDecWithPrec(150, 2),
			},
			wantErr: true,
			errMsg:  "percentage must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.beneficiary.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSpecificAssetValidation tests specific asset validation
func TestSpecificAssetValidation(t *testing.T) {
	tests := []struct {
		name    string
		asset   types.SpecificAsset
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid HODL asset",
			asset: types.SpecificAsset{
				AssetType: types.AssetTypeHODL,
				Denom:     "uhodl",
				Amount:    math.NewInt(1000000),
			},
			wantErr: false,
		},
		{
			name: "valid equity asset",
			asset: types.SpecificAsset{
				AssetType: types.AssetTypeEquity,
				CompanyID: 1,
				ClassID:   "common",
				Shares:    math.NewInt(100),
			},
			wantErr: false,
		},
		{
			name: "invalid - token with no denom",
			asset: types.SpecificAsset{
				AssetType: types.AssetTypeHODL,
				Denom:     "",
				Amount:    math.NewInt(1000000),
			},
			wantErr: true,
			errMsg:  "denom cannot be empty",
		},
		{
			name: "invalid - equity with no company ID",
			asset: types.SpecificAsset{
				AssetType: types.AssetTypeEquity,
				CompanyID: 0,
				ClassID:   "common",
				Shares:    math.NewInt(100),
			},
			wantErr: true,
			errMsg:  "company ID cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.asset.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper function to create a test inheritance plan
func createTestPlan(owner string, inactivityYears int) types.InheritancePlan {
	return types.InheritancePlan{
		PlanID: 1,
		Owner:  owner,
		Beneficiaries: []types.Beneficiary{
			{
				Address:    testBeneficiary1,
				Priority:   1,
				Percentage: math.LegacyNewDecWithPrec(60, 2), // 60%
			},
			{
				Address:    testBeneficiary2,
				Priority:   2,
				Percentage: math.LegacyNewDecWithPrec(40, 2), // 40%
			},
		},
		InactivityPeriod: time.Duration(inactivityYears) * 365 * 24 * time.Hour,
		GracePeriod:      30 * 24 * time.Hour,
		ClaimWindow:      180 * 24 * time.Hour,
		CharityAddress:   testCharity,
		Status:           types.PlanStatusActive,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// Helper function to create a test beneficiary
func createTestBeneficiary(address string, priority uint32, percentage math.LegacyDec) types.Beneficiary {
	return types.Beneficiary{
		Address:    address,
		Priority:   priority,
		Percentage: percentage,
	}
}
