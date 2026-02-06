package types

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgServer defines the agent module message server interface
type MsgServer interface {
	CreateAgent(context.Context, *MsgCreateAgent) (*MsgCreateAgentResponse, error)
	UpdateAgent(context.Context, *MsgUpdateAgent) (*MsgUpdateAgentResponse, error)
	DeleteAgent(context.Context, *MsgDeleteAgent) (*MsgDeleteAgentResponse, error)
	ActivateAgent(context.Context, *MsgActivateAgent) (*MsgActivateAgentResponse, error)
	ExecuteAction(context.Context, *MsgExecuteAction) (*MsgExecuteActionResponse, error)
	Subscribe(context.Context, *MsgSubscribe) (*MsgSubscribeResponse, error)
	RegenerateAgentKey(context.Context, *MsgRegenerateAgentKey) (*MsgRegenerateAgentKeyResponse, error)
	SetPermissions(context.Context, *MsgSetPermissions) (*MsgSetPermissionsResponse, error)
}

// ============================================================================
// MsgCreateAgent
// ============================================================================

type MsgCreateAgent struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (msg MsgCreateAgent) Route() string { return ModuleName }
func (msg MsgCreateAgent) Type() string  { return "create_agent" }

func (msg MsgCreateAgent) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if msg.Name == "" {
		return fmt.Errorf("agent name cannot be empty")
	}
	if len(msg.Name) > 100 {
		return fmt.Errorf("agent name too long (max 100 characters)")
	}
	return nil
}

func (msg MsgCreateAgent) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgCreateAgent) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

type MsgCreateAgentResponse struct {
	AgentId  uint64 `json:"agent_id"`
	AgentKey string `json:"agent_key"`
}

// ============================================================================
// MsgUpdateAgent
// ============================================================================

type MsgUpdateAgent struct {
	Owner       string `json:"owner"`
	AgentId     uint64 `json:"agent_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (msg MsgUpdateAgent) Route() string { return ModuleName }
func (msg MsgUpdateAgent) Type() string  { return "update_agent" }

func (msg MsgUpdateAgent) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if msg.AgentId == 0 {
		return fmt.Errorf("agent ID cannot be zero")
	}
	return nil
}

func (msg MsgUpdateAgent) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgUpdateAgent) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

type MsgUpdateAgentResponse struct{}

// ============================================================================
// MsgDeleteAgent
// ============================================================================

type MsgDeleteAgent struct {
	Owner   string `json:"owner"`
	AgentId uint64 `json:"agent_id"`
}

func (msg MsgDeleteAgent) Route() string { return ModuleName }
func (msg MsgDeleteAgent) Type() string  { return "delete_agent" }

func (msg MsgDeleteAgent) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if msg.AgentId == 0 {
		return fmt.Errorf("agent ID cannot be zero")
	}
	return nil
}

func (msg MsgDeleteAgent) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgDeleteAgent) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

type MsgDeleteAgentResponse struct{}

// ============================================================================
// MsgActivateAgent
// ============================================================================

type MsgActivateAgent struct {
	Owner   string `json:"owner"`
	AgentId uint64 `json:"agent_id"`
	Active  bool   `json:"active"`
}

func (msg MsgActivateAgent) Route() string { return ModuleName }
func (msg MsgActivateAgent) Type() string  { return "activate_agent" }

func (msg MsgActivateAgent) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if msg.AgentId == 0 {
		return fmt.Errorf("agent ID cannot be zero")
	}
	return nil
}

func (msg MsgActivateAgent) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgActivateAgent) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

type MsgActivateAgentResponse struct{}

// ============================================================================
// MsgExecuteAction
// ============================================================================

type MsgExecuteAction struct {
	AgentKey string                 `json:"agent_key"`
	Module   string                 `json:"module"`
	Action   string                 `json:"action"`
	Params   map[string]interface{} `json:"params"`
}

func (msg MsgExecuteAction) Route() string { return ModuleName }
func (msg MsgExecuteAction) Type() string  { return "execute_action" }

func (msg MsgExecuteAction) ValidateBasic() error {
	if msg.AgentKey == "" {
		return fmt.Errorf("agent key cannot be empty")
	}
	if msg.Module == "" {
		return fmt.Errorf("module cannot be empty")
	}
	if msg.Action == "" {
		return fmt.Errorf("action cannot be empty")
	}
	return nil
}

func (msg MsgExecuteAction) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgExecuteAction) GetSigners() []sdk.AccAddress {
	// Agent actions are signed by the agent key, not a user
	// The actual signer is verified during action execution
	return []sdk.AccAddress{}
}

type MsgExecuteActionResponse struct {
	ActionId uint64 `json:"action_id"`
	Status   string `json:"status"`
	Result   string `json:"result"`
}

// ============================================================================
// MsgSubscribe
// ============================================================================

type MsgSubscribe struct {
	Owner     string `json:"owner"`
	Tier      string `json:"tier"`
	AutoRenew bool   `json:"auto_renew"`
}

func (msg MsgSubscribe) Route() string { return ModuleName }
func (msg MsgSubscribe) Type() string  { return "subscribe" }

func (msg MsgSubscribe) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	validTiers := map[string]bool{"basic": true, "pro": true, "enterprise": true}
	if !validTiers[msg.Tier] {
		return fmt.Errorf("invalid subscription tier: %s (must be basic, pro, or enterprise)", msg.Tier)
	}
	return nil
}

func (msg MsgSubscribe) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgSubscribe) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

type MsgSubscribeResponse struct {
	ExpiresAt int64 `json:"expires_at"`
}

// ============================================================================
// MsgRegenerateAgentKey
// ============================================================================

type MsgRegenerateAgentKey struct {
	Owner   string `json:"owner"`
	AgentId uint64 `json:"agent_id"`
}

func (msg MsgRegenerateAgentKey) Route() string { return ModuleName }
func (msg MsgRegenerateAgentKey) Type() string  { return "regenerate_agent_key" }

func (msg MsgRegenerateAgentKey) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if msg.AgentId == 0 {
		return fmt.Errorf("agent ID cannot be zero")
	}
	return nil
}

func (msg MsgRegenerateAgentKey) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgRegenerateAgentKey) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

type MsgRegenerateAgentKeyResponse struct {
	NewAgentKey string `json:"new_agent_key"`
}

// ============================================================================
// MsgSetPermissions
// ============================================================================

type MsgSetPermissions struct {
	Owner       string       `json:"owner"`
	AgentId     uint64       `json:"agent_id"`
	Permissions []Permission `json:"permissions"`
}

func (msg MsgSetPermissions) Route() string { return ModuleName }
func (msg MsgSetPermissions) Type() string  { return "set_permissions" }

func (msg MsgSetPermissions) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if msg.AgentId == 0 {
		return fmt.Errorf("agent ID cannot be zero")
	}
	for i, perm := range msg.Permissions {
		if perm.Module == "" {
			return fmt.Errorf("permission %d: module cannot be empty", i)
		}
		if len(perm.Actions) == 0 {
			return fmt.Errorf("permission %d: actions cannot be empty", i)
		}
	}
	return nil
}

func (msg MsgSetPermissions) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgSetPermissions) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

type MsgSetPermissionsResponse struct{}

// ============================================================================
// MsgSetSpendingLimits
// ============================================================================

type MsgSetSpendingLimits struct {
	Owner      string   `json:"owner"`
	AgentId    uint64   `json:"agent_id"`
	DailyLimit math.Int `json:"daily_limit"`
	PerTxLimit math.Int `json:"per_tx_limit"`
}

func (msg MsgSetSpendingLimits) Route() string { return ModuleName }
func (msg MsgSetSpendingLimits) Type() string  { return "set_spending_limits" }

func (msg MsgSetSpendingLimits) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if msg.AgentId == 0 {
		return fmt.Errorf("agent ID cannot be zero")
	}
	if msg.DailyLimit.IsNegative() {
		return fmt.Errorf("daily limit cannot be negative")
	}
	if msg.PerTxLimit.IsNegative() {
		return fmt.Errorf("per-tx limit cannot be negative")
	}
	return nil
}

func (msg MsgSetSpendingLimits) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgSetSpendingLimits) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

type MsgSetSpendingLimitsResponse struct{}

// ============================================================================
// Message Registration
// ============================================================================

// RegisterMsgServer registers the MsgServer implementation
// This is called by the module's RegisterServices method
func RegisterMsgServer(s interface{}, srv MsgServer) {
	// In gRPC-based implementation, this would register with the gRPC server
	// For now, this is a placeholder that allows the module to compile
	// The actual message routing is handled by the Cosmos SDK message router
}
