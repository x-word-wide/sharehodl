package types

import (
	"fmt"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                  DefaultParams(),
		ValidatorTiers:          []ValidatorTierInfo{},
		BusinessVerifications:   []BusinessVerification{},
		ValidatorBusinessStakes: []ValidatorBusinessStake{},
		VerificationRewards:     []VerificationRewards{},
	}
}

// Validate performs basic validation of the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate validator tiers
	seenValidators := make(map[string]bool)
	for _, tier := range gs.ValidatorTiers {
		if seenValidators[tier.ValidatorAddress] {
			return fmt.Errorf("duplicate validator tier for address %s", tier.ValidatorAddress)
		}
		seenValidators[tier.ValidatorAddress] = true

		if err := tier.Validate(); err != nil {
			return err
		}
	}

	// Validate business verifications
	seenBusinesses := make(map[string]bool)
	for _, verification := range gs.BusinessVerifications {
		if seenBusinesses[verification.BusinessID] {
			return fmt.Errorf("duplicate business verification for ID %s", verification.BusinessID)
		}
		seenBusinesses[verification.BusinessID] = true

		if err := verification.Validate(); err != nil {
			return err
		}
	}

	// Validate validator business stakes
	for _, stake := range gs.ValidatorBusinessStakes {
		if err := stake.Validate(); err != nil {
			return err
		}
	}

	// Validate verification rewards
	seenRewards := make(map[string]bool)
	for _, rewards := range gs.VerificationRewards {
		if seenRewards[rewards.ValidatorAddress] {
			return fmt.Errorf("duplicate verification rewards for validator %s", rewards.ValidatorAddress)
		}
		seenRewards[rewards.ValidatorAddress] = true

		if err := rewards.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the ValidatorTierInfo
func (vti ValidatorTierInfo) Validate() error {
	if _, err := sdk.ValAddressFromBech32(vti.ValidatorAddress); err != nil {
		return fmt.Errorf("invalid validator address: %v", err)
	}

	if vti.StakedAmount.IsNegative() {
		return fmt.Errorf("staked amount cannot be negative")
	}

	if vti.StakedAmount.LT(vti.Tier.GetStakeThreshold()) {
		return fmt.Errorf("staked amount %s is less than required for tier %s (%s)", 
			vti.StakedAmount.String(), vti.Tier.String(), vti.Tier.GetStakeThreshold().String())
	}

	if vti.SuccessfulVerifications > vti.VerificationCount {
		return fmt.Errorf("successful verifications cannot exceed total verifications")
	}

	if vti.TotalRewards.IsNegative() {
		return fmt.Errorf("total rewards cannot be negative")
	}

	if vti.ReputationScore.IsNegative() || vti.ReputationScore.GT(math.LegacyNewDec(100)) {
		return fmt.Errorf("reputation score must be between 0 and 100")
	}

	return nil
}

// Validate validates the BusinessVerification
func (bv BusinessVerification) Validate() error {
	if len(bv.BusinessID) == 0 {
		return fmt.Errorf("business ID cannot be empty")
	}

	if len(bv.CompanyName) == 0 {
		return fmt.Errorf("company name cannot be empty")
	}

	if len(bv.DocumentHashes) == 0 {
		return fmt.Errorf("at least one document hash required")
	}

	if bv.StakeRequired.IsNegative() {
		return fmt.Errorf("stake required cannot be negative")
	}

	if bv.VerificationFee.IsNegative() {
		return fmt.Errorf("verification fee cannot be negative")
	}

	// Validate assigned validators
	for _, valAddr := range bv.AssignedValidators {
		if _, err := sdk.ValAddressFromBech32(valAddr); err != nil {
			return fmt.Errorf("invalid assigned validator address %s: %v", valAddr, err)
		}
	}

	if bv.SubmittedAt.IsZero() {
		return fmt.Errorf("submitted at time cannot be zero")
	}

	if !bv.VerificationDeadline.IsZero() && bv.VerificationDeadline.Before(bv.SubmittedAt) {
		return fmt.Errorf("verification deadline cannot be before submission time")
	}

	return nil
}

// Validate validates the ValidatorBusinessStake
func (vbs ValidatorBusinessStake) Validate() error {
	if _, err := sdk.ValAddressFromBech32(vbs.ValidatorAddress); err != nil {
		return fmt.Errorf("invalid validator address: %v", err)
	}

	if len(vbs.BusinessID) == 0 {
		return fmt.Errorf("business ID cannot be empty")
	}

	if vbs.StakeAmount.IsNegative() {
		return fmt.Errorf("stake amount cannot be negative")
	}

	if vbs.EquityPercentage.IsNegative() || vbs.EquityPercentage.GT(math.LegacyOneDec()) {
		return fmt.Errorf("equity percentage must be between 0 and 1")
	}

	if vbs.VestingPeriod < 0 {
		return fmt.Errorf("vesting period cannot be negative")
	}

	if vbs.StakedAt.IsZero() {
		return fmt.Errorf("staked at time cannot be zero")
	}

	return nil
}

// Validate validates the VerificationRewards
func (vr VerificationRewards) Validate() error {
	if _, err := sdk.ValAddressFromBech32(vr.ValidatorAddress); err != nil {
		return fmt.Errorf("invalid validator address: %v", err)
	}

	if vr.TotalHODLRewards.IsNegative() {
		return fmt.Errorf("total HODL rewards cannot be negative")
	}

	if vr.TotalEquityRewards.IsNegative() {
		return fmt.Errorf("total equity rewards cannot be negative")
	}

	return nil
}