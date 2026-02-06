package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type constants
const (
	TypeMsgUpdateParams            = "update_params"
	TypeMsgGrantTreasuryAllowance  = "grant_treasury_allowance"
	TypeMsgRevokeTreasuryAllowance = "revoke_treasury_allowance"
	TypeMsgFundTreasury            = "fund_treasury"
)

// MsgUpdateParams is the message to update module parameters
// Only the governance module authority can execute this
type MsgUpdateParams struct {
	// Authority must be the governance module address
	Authority string `json:"authority" yaml:"authority"`

	// Params are the new parameters to set
	Params Params `json:"params" yaml:"params"`
}

// ProtoMessage implements the proto.Message interface
func (m *MsgUpdateParams) ProtoMessage() {}

// Reset implements the proto.Message interface
func (m *MsgUpdateParams) Reset() { *m = MsgUpdateParams{} }

// String implements the proto.Message interface
func (m MsgUpdateParams) String() string { return "MsgUpdateParams" }

// Route returns the message route
func (m MsgUpdateParams) Route() string { return RouterKey }

// Type returns the message type
func (m MsgUpdateParams) Type() string { return TypeMsgUpdateParams }

// GetSigners returns the signer addresses
func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns the bytes to sign
func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// ValidateBasic performs basic validation
func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidAddress
	}
	return m.Params.Validate()
}

// MsgGrantTreasuryAllowance grants a fee allowance from treasury
type MsgGrantTreasuryAllowance struct {
	// Authority must be the governance module address
	Authority string `json:"authority" yaml:"authority"`

	// Grantee is the address receiving the grant
	Grantee string `json:"grantee" yaml:"grantee"`

	// AllowedAmount is the maximum fee amount allowed
	AllowedAmount math.Int `json:"allowed_amount" yaml:"allowed_amount"`

	// ExpirationHeight is when the grant expires (0 = never)
	ExpirationHeight uint64 `json:"expiration_height" yaml:"expiration_height"`

	// AllowedEquities restricts which equities can be used (empty = all)
	AllowedEquities []string `json:"allowed_equities" yaml:"allowed_equities"`
}

// ProtoMessage implements the proto.Message interface
func (m *MsgGrantTreasuryAllowance) ProtoMessage() {}

// Reset implements the proto.Message interface
func (m *MsgGrantTreasuryAllowance) Reset() { *m = MsgGrantTreasuryAllowance{} }

// String implements the proto.Message interface
func (m MsgGrantTreasuryAllowance) String() string { return "MsgGrantTreasuryAllowance" }

// Route returns the message route
func (m MsgGrantTreasuryAllowance) Route() string { return RouterKey }

// Type returns the message type
func (m MsgGrantTreasuryAllowance) Type() string { return TypeMsgGrantTreasuryAllowance }

// GetSigners returns the signer addresses
func (m MsgGrantTreasuryAllowance) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns the bytes to sign
func (m MsgGrantTreasuryAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// ValidateBasic performs basic validation
func (m MsgGrantTreasuryAllowance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidAddress
	}
	if _, err := sdk.AccAddressFromBech32(m.Grantee); err != nil {
		return ErrInvalidAddress
	}
	if m.AllowedAmount.IsNil() || m.AllowedAmount.IsZero() || m.AllowedAmount.IsNegative() {
		return ErrZeroAmount
	}
	return nil
}

// MsgRevokeTreasuryAllowance revokes a treasury grant
type MsgRevokeTreasuryAllowance struct {
	// Authority must be the governance module address
	Authority string `json:"authority" yaml:"authority"`

	// Grantee is the address whose grant is being revoked
	Grantee string `json:"grantee" yaml:"grantee"`
}

// ProtoMessage implements the proto.Message interface
func (m *MsgRevokeTreasuryAllowance) ProtoMessage() {}

// Reset implements the proto.Message interface
func (m *MsgRevokeTreasuryAllowance) Reset() { *m = MsgRevokeTreasuryAllowance{} }

// String implements the proto.Message interface
func (m MsgRevokeTreasuryAllowance) String() string { return "MsgRevokeTreasuryAllowance" }

// Route returns the message route
func (m MsgRevokeTreasuryAllowance) Route() string { return RouterKey }

// Type returns the message type
func (m MsgRevokeTreasuryAllowance) Type() string { return TypeMsgRevokeTreasuryAllowance }

// GetSigners returns the signer addresses
func (m MsgRevokeTreasuryAllowance) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns the bytes to sign
func (m MsgRevokeTreasuryAllowance) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// ValidateBasic performs basic validation
func (m MsgRevokeTreasuryAllowance) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return ErrInvalidAddress
	}
	if _, err := sdk.AccAddressFromBech32(m.Grantee); err != nil {
		return ErrInvalidAddress
	}
	return nil
}

// MsgFundTreasury adds funds to the fee abstraction treasury
type MsgFundTreasury struct {
	// Funder is the address providing funds
	Funder string `json:"funder" yaml:"funder"`

	// Amount is the amount to add to treasury
	Amount math.Int `json:"amount" yaml:"amount"`
}

// ProtoMessage implements the proto.Message interface
func (m *MsgFundTreasury) ProtoMessage() {}

// Reset implements the proto.Message interface
func (m *MsgFundTreasury) Reset() { *m = MsgFundTreasury{} }

// String implements the proto.Message interface
func (m MsgFundTreasury) String() string { return "MsgFundTreasury" }

// Route returns the message route
func (m MsgFundTreasury) Route() string { return RouterKey }

// Type returns the message type
func (m MsgFundTreasury) Type() string { return TypeMsgFundTreasury }

// GetSigners returns the signer addresses
func (m MsgFundTreasury) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Funder)
	return []sdk.AccAddress{addr}
}

// GetSignBytes returns the bytes to sign
func (m MsgFundTreasury) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// ValidateBasic performs basic validation
func (m MsgFundTreasury) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Funder); err != nil {
		return ErrInvalidAddress
	}
	if m.Amount.IsNil() || m.Amount.IsZero() || m.Amount.IsNegative() {
		return ErrZeroAmount
	}
	return nil
}

// =============================================================================
// MESSAGE SERVER INTERFACE
// =============================================================================

// MsgServer is the interface for the msg server
type MsgServer interface {
	UpdateParams(ctx context.Context, msg *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
	GrantTreasuryAllowance(ctx context.Context, msg *MsgGrantTreasuryAllowance) (*MsgGrantTreasuryAllowanceResponse, error)
	RevokeTreasuryAllowance(ctx context.Context, msg *MsgRevokeTreasuryAllowance) (*MsgRevokeTreasuryAllowanceResponse, error)
	FundTreasury(ctx context.Context, msg *MsgFundTreasury) (*MsgFundTreasuryResponse, error)
}

// Response types

// MsgUpdateParamsResponse is the response from UpdateParams
type MsgUpdateParamsResponse struct{}

// MsgGrantTreasuryAllowanceResponse is the response from GrantTreasuryAllowance
type MsgGrantTreasuryAllowanceResponse struct{}

// MsgRevokeTreasuryAllowanceResponse is the response from RevokeTreasuryAllowance
type MsgRevokeTreasuryAllowanceResponse struct{}

// MsgFundTreasuryResponse is the response from FundTreasury
type MsgFundTreasuryResponse struct{}

// RegisterMsgServer is a placeholder for gRPC registration
// In a full implementation, this would be generated by protobuf
func RegisterMsgServer(registrar interface{}, srv MsgServer) {
	// Direct registration - the actual implementation uses
	// module.Configurator.MsgServer() which handles routing
}
