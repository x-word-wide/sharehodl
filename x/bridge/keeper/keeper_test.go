package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
)

// TestParamsValidation tests parameter validation
func TestParamsValidation(t *testing.T) {
	tests := []struct {
		name    string
		params  types.Params
		wantErr bool
	}{
		{
			name:    "default params are valid",
			params:  types.DefaultParams(),
			wantErr: false,
		},
		{
			name: "negative swap fee rate",
			params: types.NewParams(
				math.LegacyNewDec(-1),
				types.DefaultMinReserveRatio,
				types.DefaultMaxWithdrawalPercent,
				types.DefaultRequiredValidatorApproval,
				types.DefaultWithdrawalVotingPeriod,
				true,
			),
			wantErr: true,
		},
		{
			name: "swap fee rate > 100%",
			params: types.NewParams(
				math.LegacyNewDec(2),
				types.DefaultMinReserveRatio,
				types.DefaultMaxWithdrawalPercent,
				types.DefaultRequiredValidatorApproval,
				types.DefaultWithdrawalVotingPeriod,
				true,
			),
			wantErr: true,
		},
		{
			name: "negative min reserve ratio",
			params: types.NewParams(
				types.DefaultSwapFeeRate,
				math.LegacyNewDec(-1),
				types.DefaultMaxWithdrawalPercent,
				types.DefaultRequiredValidatorApproval,
				types.DefaultWithdrawalVotingPeriod,
				true,
			),
			wantErr: true,
		},
		{
			name: "validator approval < 50%",
			params: types.NewParams(
				types.DefaultSwapFeeRate,
				types.DefaultMinReserveRatio,
				types.DefaultMaxWithdrawalPercent,
				math.LegacyNewDecWithPrec(40, 2),
				types.DefaultWithdrawalVotingPeriod,
				true,
			),
			wantErr: true,
		},
		{
			name: "zero voting period",
			params: types.NewParams(
				types.DefaultSwapFeeRate,
				types.DefaultMinReserveRatio,
				types.DefaultMaxWithdrawalPercent,
				types.DefaultRequiredValidatorApproval,
				0,
				true,
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSupportedAssetValidation tests supported asset validation
func TestSupportedAssetValidation(t *testing.T) {
	tests := []struct {
		name    string
		asset   types.SupportedAsset
		wantErr bool
	}{
		{
			name: "valid asset",
			asset: types.NewSupportedAsset(
				"uusdt",
				math.LegacyOneDec(),
				math.LegacyOneDec(),
				math.NewInt(1_000_000),
				math.NewInt(1_000_000_000),
				true,
			),
			wantErr: false,
		},
		{
			name: "empty denom",
			asset: types.NewSupportedAsset(
				"",
				math.LegacyOneDec(),
				math.LegacyOneDec(),
				math.NewInt(1_000_000),
				math.NewInt(1_000_000_000),
				true,
			),
			wantErr: true,
		},
		{
			name: "zero swap rate to HODL",
			asset: types.NewSupportedAsset(
				"uusdt",
				math.LegacyZeroDec(),
				math.LegacyOneDec(),
				math.NewInt(1_000_000),
				math.NewInt(1_000_000_000),
				true,
			),
			wantErr: true,
		},
		{
			name: "negative swap rate from HODL",
			asset: types.NewSupportedAsset(
				"uusdt",
				math.LegacyOneDec(),
				math.LegacyNewDec(-1),
				math.NewInt(1_000_000),
				math.NewInt(1_000_000_000),
				true,
			),
			wantErr: true,
		},
		{
			name: "max < min swap amount",
			asset: types.NewSupportedAsset(
				"uusdt",
				math.LegacyOneDec(),
				math.LegacyOneDec(),
				math.NewInt(1_000_000_000),
				math.NewInt(1_000_000),
				true,
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.asset.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSwapRecordValidation tests swap record validation
func TestSwapRecordValidation(t *testing.T) {
	validAddr := "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"

	tests := []struct {
		name    string
		record  types.SwapRecord
		wantErr bool
	}{
		{
			name: "valid swap record",
			record: types.NewSwapRecord(
				1,
				validAddr,
				types.SwapDirectionIn,
				"uusdt",
				math.NewInt(1000000),
				"uhodl",
				math.NewInt(999000),
				math.NewInt(1000),
				math.LegacyOneDec(),
				time.Now(),
				100,
			),
			wantErr: false,
		},
		{
			name: "empty user address",
			record: types.NewSwapRecord(
				1,
				"",
				types.SwapDirectionIn,
				"uusdt",
				math.NewInt(1000000),
				"uhodl",
				math.NewInt(999000),
				math.NewInt(1000),
				math.LegacyOneDec(),
				time.Now(),
				100,
			),
			wantErr: true,
		},
		{
			name: "negative input amount",
			record: types.NewSwapRecord(
				1,
				validAddr,
				types.SwapDirectionIn,
				"uusdt",
				math.NewInt(-1000000),
				"uhodl",
				math.NewInt(999000),
				math.NewInt(1000),
				math.LegacyOneDec(),
				time.Now(),
				100,
			),
			wantErr: true,
		},
		{
			name: "zero swap rate",
			record: types.NewSwapRecord(
				1,
				validAddr,
				types.SwapDirectionIn,
				"uusdt",
				math.NewInt(1000000),
				"uhodl",
				math.NewInt(999000),
				math.NewInt(1000),
				math.LegacyZeroDec(),
				time.Now(),
				100,
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestReserveWithdrawalValidation tests withdrawal proposal validation
func TestReserveWithdrawalValidation(t *testing.T) {
	validAddr := "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"

	tests := []struct {
		name     string
		proposal types.ReserveWithdrawal
		wantErr  bool
	}{
		{
			name: "valid withdrawal proposal",
			proposal: types.NewReserveWithdrawal(
				1,
				validAddr,
				validAddr,
				"uusdt",
				math.NewInt(1000000),
				"Need funds for investment",
				10,
				604800,
			),
			wantErr: false,
		},
		{
			name: "empty proposer",
			proposal: types.NewReserveWithdrawal(
				1,
				"",
				validAddr,
				"uusdt",
				math.NewInt(1000000),
				"Need funds for investment",
				10,
				604800,
			),
			wantErr: true,
		},
		{
			name: "empty recipient",
			proposal: types.NewReserveWithdrawal(
				1,
				validAddr,
				"",
				"uusdt",
				math.NewInt(1000000),
				"Need funds for investment",
				10,
				604800,
			),
			wantErr: true,
		},
		{
			name: "zero amount",
			proposal: types.NewReserveWithdrawal(
				1,
				validAddr,
				validAddr,
				"uusdt",
				math.ZeroInt(),
				"Need funds for investment",
				10,
				604800,
			),
			wantErr: true,
		},
		{
			name: "empty reason",
			proposal: types.NewReserveWithdrawal(
				1,
				validAddr,
				validAddr,
				"uusdt",
				math.NewInt(1000000),
				"",
				10,
				604800,
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.proposal.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestProposalStatusAndRates tests proposal status methods
func TestProposalStatusAndRates(t *testing.T) {
	validAddr := "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"
	now := time.Now()

	proposal := types.ReserveWithdrawal{
		ID:              1,
		Proposer:        validAddr,
		Recipient:       validAddr,
		Denom:           "uusdt",
		Amount:          math.NewInt(1000000),
		Reason:          "Test",
		Status:          types.ProposalStatusActive,
		YesVotes:        7,
		NoVotes:         3,
		TotalValidators: 10,
		CreatedAt:       now,
		VotingEndsAt:    now.Add(24 * time.Hour),
	}

	// Test IsActive
	require.True(t, proposal.IsActive())

	// Test IsExpired
	require.False(t, proposal.IsExpired())

	// Test GetApprovalRate
	approvalRate := proposal.GetApprovalRate()
	expectedRate := math.LegacyNewDec(7).QuoInt64(10)
	require.True(t, approvalRate.Equal(expectedRate))

	// Test expired proposal
	expiredProposal := proposal
	expiredProposal.VotingEndsAt = now.Add(-1 * time.Hour)
	require.False(t, expiredProposal.IsActive())
	require.True(t, expiredProposal.IsExpired())

	// Test non-active proposal
	executedProposal := proposal
	executedProposal.Status = types.ProposalStatusExecuted
	require.False(t, executedProposal.IsActive())
}

// TestGenesisValidation tests genesis state validation
func TestGenesisValidation(t *testing.T) {
	tests := []struct {
		name    string
		genesis types.GenesisState
		wantErr bool
	}{
		{
			name:    "default genesis is valid",
			genesis: *types.DefaultGenesis(),
			wantErr: false,
		},
		{
			name: "invalid params",
			genesis: types.GenesisState{
				Params: types.NewParams(
					math.LegacyNewDec(-1),
					types.DefaultMinReserveRatio,
					types.DefaultMaxWithdrawalPercent,
					types.DefaultRequiredValidatorApproval,
					types.DefaultWithdrawalVotingPeriod,
					true,
				),
				SupportedAssets:     []types.SupportedAsset{},
				Reserves:            []types.BridgeReserve{},
				SwapRecords:         []types.SwapRecord{},
				WithdrawalProposals: []types.ReserveWithdrawal{},
				NextSwapID:          1,
				NextProposalID:      1,
			},
			wantErr: true,
		},
		{
			name: "duplicate supported assets",
			genesis: types.GenesisState{
				Params: types.DefaultParams(),
				SupportedAssets: []types.SupportedAsset{
					types.NewSupportedAsset("uusdt", math.LegacyOneDec(), math.LegacyOneDec(), math.NewInt(1), math.NewInt(1000), true),
					types.NewSupportedAsset("uusdt", math.LegacyOneDec(), math.LegacyOneDec(), math.NewInt(1), math.NewInt(1000), true),
				},
				Reserves:            []types.BridgeReserve{},
				SwapRecords:         []types.SwapRecord{},
				WithdrawalProposals: []types.ReserveWithdrawal{},
				NextSwapID:          1,
				NextProposalID:      1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.genesis.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
