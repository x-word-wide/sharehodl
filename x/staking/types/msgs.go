package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgUpdateParams = "update_params"
	TypeMsgStake        = "stake"
	TypeMsgUnstake      = "unstake"
	TypeMsgClaimRewards = "claim_rewards"
)

// MsgServer defines the msg service interface
type MsgServer interface {
	UpdateParams(ctx context.Context, msg *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
	Stake(ctx context.Context, msg *MsgStake) (*MsgStakeResponse, error)
	Unstake(ctx context.Context, msg *MsgUnstake) (*MsgUnstakeResponse, error)
	ClaimRewards(ctx context.Context, msg *MsgClaimRewards) (*MsgClaimRewardsResponse, error)
}

// ============================================================================
// MsgUpdateParams - Governance-controlled parameter update
// ============================================================================

// MsgUpdateParams is the message for updating staking parameters via governance
type MsgUpdateParams struct {
	// Authority is the address that controls the module (typically gov module)
	Authority string `json:"authority"`
	// Params are the new parameters to set
	Params Params `json:"params"`
}

// ProtoMessage implements proto.Message interface
func (m *MsgUpdateParams) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *MsgUpdateParams) Reset() { *m = MsgUpdateParams{} }

// String implements proto.Message interface
func (m *MsgUpdateParams) String() string { return "msg_update_params" }

// Route implements sdk.Msg
func (msg MsgUpdateParams) Route() string { return ModuleName }

// Type implements sdk.Msg
func (msg MsgUpdateParams) Type() string { return TypeMsgUpdateParams }

// GetSigners implements sdk.Msg
func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

// ValidateBasic implements sdk.Msg
func (msg MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}
	return msg.Params.Validate()
}

// MsgUpdateParamsResponse is the response type for MsgUpdateParams
type MsgUpdateParamsResponse struct{}

// ProtoMessage implements proto.Message interface
func (m *MsgUpdateParamsResponse) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *MsgUpdateParamsResponse) Reset() { *m = MsgUpdateParamsResponse{} }

// String implements proto.Message interface
func (m *MsgUpdateParamsResponse) String() string { return "msg_update_params_response" }

// ============================================================================
// MsgStake - Stake HODL to earn rewards and gain tier privileges
// ============================================================================

// MsgStake is the message for staking HODL
type MsgStake struct {
	Staker string   `json:"staker"`
	Amount math.Int `json:"amount"`
}

// ProtoMessage implements proto.Message interface
func (m *MsgStake) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *MsgStake) Reset() { *m = MsgStake{} }

// String implements proto.Message interface
func (m *MsgStake) String() string { return "msg_stake" }

// Route implements sdk.Msg
func (msg MsgStake) Route() string { return ModuleName }

// Type implements sdk.Msg
func (msg MsgStake) Type() string { return TypeMsgStake }

// GetSigners implements sdk.Msg
func (msg MsgStake) GetSigners() []sdk.AccAddress {
	staker, _ := sdk.AccAddressFromBech32(msg.Staker)
	return []sdk.AccAddress{staker}
}

// ValidateBasic implements sdk.Msg
func (msg MsgStake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Staker); err != nil {
		return err
	}
	if !msg.Amount.IsPositive() {
		return ErrInvalidAmount
	}
	return nil
}

// MsgStakeResponse is the response type for MsgStake
type MsgStakeResponse struct {
	NewTier     string   `json:"new_tier"`
	TotalStaked math.Int `json:"total_staked"`
}

// ProtoMessage implements proto.Message interface
func (m *MsgStakeResponse) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *MsgStakeResponse) Reset() { *m = MsgStakeResponse{} }

// String implements proto.Message interface
func (m *MsgStakeResponse) String() string { return "msg_stake_response" }

// ============================================================================
// MsgUnstake - Unstake HODL (may cause tier demotion)
// ============================================================================

// MsgUnstake is the message for unstaking HODL
type MsgUnstake struct {
	Staker string   `json:"staker"`
	Amount math.Int `json:"amount"`
}

// ProtoMessage implements proto.Message interface
func (m *MsgUnstake) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *MsgUnstake) Reset() { *m = MsgUnstake{} }

// String implements proto.Message interface
func (m *MsgUnstake) String() string { return "msg_unstake" }

// Route implements sdk.Msg
func (msg MsgUnstake) Route() string { return ModuleName }

// Type implements sdk.Msg
func (msg MsgUnstake) Type() string { return TypeMsgUnstake }

// GetSigners implements sdk.Msg
func (msg MsgUnstake) GetSigners() []sdk.AccAddress {
	staker, _ := sdk.AccAddressFromBech32(msg.Staker)
	return []sdk.AccAddress{staker}
}

// ValidateBasic implements sdk.Msg
func (msg MsgUnstake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Staker); err != nil {
		return err
	}
	if !msg.Amount.IsPositive() {
		return ErrInvalidAmount
	}
	return nil
}

// MsgUnstakeResponse is the response type for MsgUnstake
type MsgUnstakeResponse struct {
	NewTier         string   `json:"new_tier"`
	RemainingStaked math.Int `json:"remaining_staked"`
}

// ProtoMessage implements proto.Message interface
func (m *MsgUnstakeResponse) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *MsgUnstakeResponse) Reset() { *m = MsgUnstakeResponse{} }

// String implements proto.Message interface
func (m *MsgUnstakeResponse) String() string { return "msg_unstake_response" }

// ============================================================================
// MsgClaimRewards - Claim pending staking rewards
// ============================================================================

// MsgClaimRewards is the message for claiming staking rewards
type MsgClaimRewards struct {
	Staker string `json:"staker"`
}

// ProtoMessage implements proto.Message interface
func (m *MsgClaimRewards) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *MsgClaimRewards) Reset() { *m = MsgClaimRewards{} }

// String implements proto.Message interface
func (m *MsgClaimRewards) String() string { return "msg_claim_rewards" }

// Route implements sdk.Msg
func (msg MsgClaimRewards) Route() string { return ModuleName }

// Type implements sdk.Msg
func (msg MsgClaimRewards) Type() string { return TypeMsgClaimRewards }

// GetSigners implements sdk.Msg
func (msg MsgClaimRewards) GetSigners() []sdk.AccAddress {
	staker, _ := sdk.AccAddressFromBech32(msg.Staker)
	return []sdk.AccAddress{staker}
}

// ValidateBasic implements sdk.Msg
func (msg MsgClaimRewards) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Staker); err != nil {
		return err
	}
	return nil
}

// MsgClaimRewardsResponse is the response type for MsgClaimRewards
type MsgClaimRewardsResponse struct {
	ClaimedAmount math.Int `json:"claimed_amount"`
}

// ProtoMessage implements proto.Message interface
func (m *MsgClaimRewardsResponse) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *MsgClaimRewardsResponse) Reset() { *m = MsgClaimRewardsResponse{} }

// String implements proto.Message interface
func (m *MsgClaimRewardsResponse) String() string { return "msg_claim_rewards_response" }
