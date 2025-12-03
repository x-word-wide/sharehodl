package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types for validator tier operations

// SimpleMsgRegisterValidatorTier registers a validator in a specific tier
type SimpleMsgRegisterValidatorTier struct {
	ValidatorAddress string   `json:"validator_address"`
	StakeAmount      math.Int `json:"stake_amount"`
	DesiredTier      ValidatorTier `json:"desired_tier"`
}

func (msg SimpleMsgRegisterValidatorTier) Route() string { return ModuleName }
func (msg SimpleMsgRegisterValidatorTier) Type() string { return "register_validator_tier" }
func (msg SimpleMsgRegisterValidatorTier) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return fmt.Errorf("invalid validator address: %v", err)
	}
	if !msg.StakeAmount.IsPositive() {
		return fmt.Errorf("stake amount must be positive")
	}
	if msg.DesiredTier > TierDiamond {
		return fmt.Errorf("invalid tier specified")
	}
	if msg.StakeAmount.LT(msg.DesiredTier.GetStakeThreshold()) {
		return fmt.Errorf("insufficient stake for desired tier: need %s, got %s", 
			msg.DesiredTier.GetStakeThreshold().String(), msg.StakeAmount.String())
	}
	return nil
}

func (msg SimpleMsgRegisterValidatorTier) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg SimpleMsgRegisterValidatorTier) GetSigners() []sdk.AccAddress {
	valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	accAddr := sdk.AccAddress(valAddr)
	return []sdk.AccAddress{accAddr}
}

// SimpleMsgSubmitBusinessVerification submits a business for verification
type SimpleMsgSubmitBusinessVerification struct {
	Submitter           string        `json:"submitter"`
	BusinessID          string        `json:"business_id"`
	CompanyName         string        `json:"company_name"`
	BusinessType        string        `json:"business_type"`
	RegistrationCountry string        `json:"registration_country"`
	DocumentHashes      []string      `json:"document_hashes"`
	RequiredTier        ValidatorTier `json:"required_tier"`
	VerificationFee     math.Int      `json:"verification_fee"`
}

func (msg SimpleMsgSubmitBusinessVerification) Route() string { return ModuleName }
func (msg SimpleMsgSubmitBusinessVerification) Type() string { return "submit_business_verification" }
func (msg SimpleMsgSubmitBusinessVerification) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Submitter); err != nil {
		return fmt.Errorf("invalid submitter address: %v", err)
	}
	if len(msg.BusinessID) == 0 {
		return fmt.Errorf("business ID cannot be empty")
	}
	if len(msg.CompanyName) == 0 {
		return fmt.Errorf("company name cannot be empty")
	}
	if len(msg.DocumentHashes) == 0 {
		return fmt.Errorf("at least one document hash required")
	}
	if !msg.VerificationFee.IsPositive() {
		return fmt.Errorf("verification fee must be positive")
	}
	return nil
}

func (msg SimpleMsgSubmitBusinessVerification) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg SimpleMsgSubmitBusinessVerification) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Submitter)
	return []sdk.AccAddress{addr}
}

// SimpleMsgCompleteVerification completes a business verification
type SimpleMsgCompleteVerification struct {
	ValidatorAddress  string                      `json:"validator_address"`
	BusinessID        string                      `json:"business_id"`
	VerificationResult BusinessVerificationStatus `json:"verification_result"`
	VerificationNotes string                      `json:"verification_notes"`
	StakeInBusiness   math.Int                    `json:"stake_in_business"`
}

func (msg SimpleMsgCompleteVerification) Route() string { return ModuleName }
func (msg SimpleMsgCompleteVerification) Type() string { return "complete_verification" }
func (msg SimpleMsgCompleteVerification) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return fmt.Errorf("invalid validator address: %v", err)
	}
	if len(msg.BusinessID) == 0 {
		return fmt.Errorf("business ID cannot be empty")
	}
	if msg.VerificationResult != VerificationApproved && msg.VerificationResult != VerificationRejected {
		return fmt.Errorf("verification result must be approved or rejected")
	}
	if msg.StakeInBusiness.IsNegative() {
		return fmt.Errorf("stake amount cannot be negative")
	}
	return nil
}

func (msg SimpleMsgCompleteVerification) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg SimpleMsgCompleteVerification) GetSigners() []sdk.AccAddress {
	valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	accAddr := sdk.AccAddress(valAddr)
	return []sdk.AccAddress{accAddr}
}

// SimpleMsgStakeInBusiness allows validators to stake in verified businesses
type SimpleMsgStakeInBusiness struct {
	ValidatorAddress string   `json:"validator_address"`
	BusinessID       string   `json:"business_id"`
	StakeAmount      math.Int `json:"stake_amount"`
	VestingPeriod    time.Duration `json:"vesting_period"`
}

func (msg SimpleMsgStakeInBusiness) Route() string { return ModuleName }
func (msg SimpleMsgStakeInBusiness) Type() string { return "stake_in_business" }
func (msg SimpleMsgStakeInBusiness) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return fmt.Errorf("invalid validator address: %v", err)
	}
	if len(msg.BusinessID) == 0 {
		return fmt.Errorf("business ID cannot be empty")
	}
	if !msg.StakeAmount.IsPositive() {
		return fmt.Errorf("stake amount must be positive")
	}
	if msg.VestingPeriod < 0 {
		return fmt.Errorf("vesting period cannot be negative")
	}
	return nil
}

func (msg SimpleMsgStakeInBusiness) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg SimpleMsgStakeInBusiness) GetSigners() []sdk.AccAddress {
	valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	accAddr := sdk.AccAddress(valAddr)
	return []sdk.AccAddress{accAddr}
}

// SimpleMsgClaimVerificationRewards claims verification rewards
type SimpleMsgClaimVerificationRewards struct {
	ValidatorAddress string `json:"validator_address"`
}

func (msg SimpleMsgClaimVerificationRewards) Route() string { return ModuleName }
func (msg SimpleMsgClaimVerificationRewards) Type() string { return "claim_verification_rewards" }
func (msg SimpleMsgClaimVerificationRewards) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return fmt.Errorf("invalid validator address: %v", err)
	}
	return nil
}

func (msg SimpleMsgClaimVerificationRewards) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg SimpleMsgClaimVerificationRewards) GetSigners() []sdk.AccAddress {
	valAddr, _ := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	accAddr := sdk.AccAddress(valAddr)
	return []sdk.AccAddress{accAddr}
}