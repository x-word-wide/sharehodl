package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types
const (
	TypeMsgCreatePlan   = "create_plan"
	TypeMsgUpdatePlan   = "update_plan"
	TypeMsgCancelPlan   = "cancel_plan"
	TypeMsgClaimAssets  = "claim_assets"
	TypeMsgCancelTrigger = "cancel_trigger"
)

// MsgCreatePlan creates a new inheritance plan
type MsgCreatePlan struct {
	Creator          string        `json:"creator"`
	Beneficiaries    []Beneficiary `json:"beneficiaries"`
	InactivityPeriod time.Duration `json:"inactivity_period"`
	GracePeriod      time.Duration `json:"grace_period"`
	ClaimWindow      time.Duration `json:"claim_window"`
	CharityAddress   string        `json:"charity_address,omitempty"`
}

// ValidateBasic performs basic validation
func (msg MsgCreatePlan) ValidateBasic() error {
	if msg.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return fmt.Errorf("invalid creator address: %v", err)
	}
	if len(msg.Beneficiaries) == 0 {
		return fmt.Errorf("must have at least one beneficiary")
	}
	if msg.InactivityPeriod <= 0 {
		return fmt.Errorf("inactivity period must be positive")
	}
	if msg.GracePeriod < 30*24*time.Hour {
		return fmt.Errorf("grace period must be at least 30 days")
	}
	if msg.ClaimWindow <= 0 {
		return fmt.Errorf("claim window must be positive")
	}

	// Validate beneficiaries
	totalPercentage := math.LegacyZeroDec()
	priorityMap := make(map[uint32]bool)
	for _, b := range msg.Beneficiaries {
		if err := b.Validate(); err != nil {
			return fmt.Errorf("invalid beneficiary: %w", err)
		}
		if priorityMap[b.Priority] {
			return fmt.Errorf("duplicate priority %d", b.Priority)
		}
		priorityMap[b.Priority] = true
		totalPercentage = totalPercentage.Add(b.Percentage)
	}

	if totalPercentage.GT(math.LegacyOneDec()) {
		return fmt.Errorf("total beneficiary percentage exceeds 100%%: %s", totalPercentage)
	}

	if msg.CharityAddress != "" {
		if _, err := sdk.AccAddressFromBech32(msg.CharityAddress); err != nil {
			return fmt.Errorf("invalid charity address: %v", err)
		}
	}

	return nil
}

// GetSigners returns the expected signers
func (msg MsgCreatePlan) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

// MsgUpdatePlan updates an existing inheritance plan
type MsgUpdatePlan struct {
	Creator          string        `json:"creator"`
	PlanID           uint64        `json:"plan_id"`
	Beneficiaries    []Beneficiary `json:"beneficiaries"`
	InactivityPeriod time.Duration `json:"inactivity_period"`
	GracePeriod      time.Duration `json:"grace_period"`
	ClaimWindow      time.Duration `json:"claim_window"`
	CharityAddress   string        `json:"charity_address,omitempty"`
}

// ValidateBasic performs basic validation
func (msg MsgUpdatePlan) ValidateBasic() error {
	if msg.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return fmt.Errorf("invalid creator address: %v", err)
	}
	if msg.PlanID == 0 {
		return fmt.Errorf("plan ID cannot be zero")
	}
	if len(msg.Beneficiaries) == 0 {
		return fmt.Errorf("must have at least one beneficiary")
	}
	if msg.InactivityPeriod <= 0 {
		return fmt.Errorf("inactivity period must be positive")
	}
	if msg.GracePeriod < 30*24*time.Hour {
		return fmt.Errorf("grace period must be at least 30 days")
	}
	if msg.ClaimWindow <= 0 {
		return fmt.Errorf("claim window must be positive")
	}

	// Validate beneficiaries
	totalPercentage := math.LegacyZeroDec()
	priorityMap := make(map[uint32]bool)
	for _, b := range msg.Beneficiaries {
		if err := b.Validate(); err != nil {
			return fmt.Errorf("invalid beneficiary: %w", err)
		}
		if priorityMap[b.Priority] {
			return fmt.Errorf("duplicate priority %d", b.Priority)
		}
		priorityMap[b.Priority] = true
		totalPercentage = totalPercentage.Add(b.Percentage)
	}

	if totalPercentage.GT(math.LegacyOneDec()) {
		return fmt.Errorf("total beneficiary percentage exceeds 100%%: %s", totalPercentage)
	}

	if msg.CharityAddress != "" {
		if _, err := sdk.AccAddressFromBech32(msg.CharityAddress); err != nil {
			return fmt.Errorf("invalid charity address: %v", err)
		}
	}

	return nil
}

// GetSigners returns the expected signers
func (msg MsgUpdatePlan) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

// MsgCancelPlan cancels an inheritance plan
type MsgCancelPlan struct {
	Creator string `json:"creator"`
	PlanID  uint64 `json:"plan_id"`
}

// ValidateBasic performs basic validation
func (msg MsgCancelPlan) ValidateBasic() error {
	if msg.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return fmt.Errorf("invalid creator address: %v", err)
	}
	if msg.PlanID == 0 {
		return fmt.Errorf("plan ID cannot be zero")
	}
	return nil
}

// GetSigners returns the expected signers
func (msg MsgCancelPlan) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

// MsgClaimAssets allows a beneficiary to claim their inheritance
type MsgClaimAssets struct {
	Claimer string `json:"claimer"`
	PlanID  uint64 `json:"plan_id"`
}

// ValidateBasic performs basic validation
func (msg MsgClaimAssets) ValidateBasic() error {
	if msg.Claimer == "" {
		return fmt.Errorf("claimer cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Claimer); err != nil {
		return fmt.Errorf("invalid claimer address: %v", err)
	}
	if msg.PlanID == 0 {
		return fmt.Errorf("plan ID cannot be zero")
	}
	return nil
}

// GetSigners returns the expected signers
func (msg MsgClaimAssets) GetSigners() []sdk.AccAddress {
	claimer, _ := sdk.AccAddressFromBech32(msg.Claimer)
	return []sdk.AccAddress{claimer}
}

// MsgCancelTrigger allows owner to cancel a triggered switch
type MsgCancelTrigger struct {
	Creator string `json:"creator"`
	PlanID  uint64 `json:"plan_id"`
}

// ValidateBasic performs basic validation
func (msg MsgCancelTrigger) ValidateBasic() error {
	if msg.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return fmt.Errorf("invalid creator address: %v", err)
	}
	if msg.PlanID == 0 {
		return fmt.Errorf("plan ID cannot be zero")
	}
	return nil
}

// GetSigners returns the expected signers
func (msg MsgCancelTrigger) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

// Message response types
type MsgCreatePlanResponse struct {
	PlanID uint64 `json:"plan_id"`
}

type MsgUpdatePlanResponse struct{}

type MsgCancelPlanResponse struct{}

type MsgClaimAssetsResponse struct {
	TransferredAssets []TransferredAsset `json:"transferred_assets"`
}

type MsgCancelTriggerResponse struct{}

// ProtoMessage implementations
func (m *MsgCreatePlan) ProtoMessage()          {}
func (m *MsgCreatePlan) Reset()                 { *m = MsgCreatePlan{} }
func (m *MsgCreatePlan) String() string         { return fmt.Sprintf("MsgCreatePlan{Creator:%s}", m.Creator) }
func (m *MsgUpdatePlan) ProtoMessage()          {}
func (m *MsgUpdatePlan) Reset()                 { *m = MsgUpdatePlan{} }
func (m *MsgUpdatePlan) String() string         { return fmt.Sprintf("MsgUpdatePlan{Creator:%s,PlanID:%d}", m.Creator, m.PlanID) }
func (m *MsgCancelPlan) ProtoMessage()          {}
func (m *MsgCancelPlan) Reset()                 { *m = MsgCancelPlan{} }
func (m *MsgCancelPlan) String() string         { return fmt.Sprintf("MsgCancelPlan{Creator:%s,PlanID:%d}", m.Creator, m.PlanID) }
func (m *MsgClaimAssets) ProtoMessage()         {}
func (m *MsgClaimAssets) Reset()                { *m = MsgClaimAssets{} }
func (m *MsgClaimAssets) String() string        { return fmt.Sprintf("MsgClaimAssets{Claimer:%s,PlanID:%d}", m.Claimer, m.PlanID) }
func (m *MsgCancelTrigger) ProtoMessage()       {}
func (m *MsgCancelTrigger) Reset()              { *m = MsgCancelTrigger{} }
func (m *MsgCancelTrigger) String() string      { return fmt.Sprintf("MsgCancelTrigger{Creator:%s,PlanID:%d}", m.Creator, m.PlanID) }
func (m *MsgCreatePlanResponse) ProtoMessage()  {}
func (m *MsgCreatePlanResponse) Reset()         { *m = MsgCreatePlanResponse{} }
func (m *MsgCreatePlanResponse) String() string { return fmt.Sprintf("MsgCreatePlanResponse{PlanID:%d}", m.PlanID) }
func (m *MsgUpdatePlanResponse) ProtoMessage()  {}
func (m *MsgUpdatePlanResponse) Reset()         { *m = MsgUpdatePlanResponse{} }
func (m *MsgUpdatePlanResponse) String() string { return "MsgUpdatePlanResponse{}" }
func (m *MsgCancelPlanResponse) ProtoMessage()  {}
func (m *MsgCancelPlanResponse) Reset()         { *m = MsgCancelPlanResponse{} }
func (m *MsgCancelPlanResponse) String() string { return "MsgCancelPlanResponse{}" }
func (m *MsgClaimAssetsResponse) ProtoMessage() {}
func (m *MsgClaimAssetsResponse) Reset()        { *m = MsgClaimAssetsResponse{} }
func (m *MsgClaimAssetsResponse) String() string { return "MsgClaimAssetsResponse{}" }
func (m *MsgCancelTriggerResponse) ProtoMessage() {}
func (m *MsgCancelTriggerResponse) Reset()      { *m = MsgCancelTriggerResponse{} }
func (m *MsgCancelTriggerResponse) String() string { return "MsgCancelTriggerResponse{}" }
