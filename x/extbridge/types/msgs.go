package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types for extbridge module

// MsgObserveDeposit - Validator observes a deposit on external chain
type MsgObserveDeposit struct {
	Validator           string   `json:"validator" yaml:"validator"`
	ChainID             string   `json:"chain_id" yaml:"chain_id"`
	AssetSymbol         string   `json:"asset_symbol" yaml:"asset_symbol"`
	ExternalTxHash      string   `json:"external_tx_hash" yaml:"external_tx_hash"`
	ExternalBlockHeight uint64   `json:"external_block_height" yaml:"external_block_height"`
	Sender              string   `json:"sender" yaml:"sender"`       // External chain address
	Recipient           string   `json:"recipient" yaml:"recipient"` // ShareHODL address
	Amount              math.Int `json:"amount" yaml:"amount"`       // Amount in external asset
}

// ValidateBasic performs basic validation
func (msg MsgObserveDeposit) ValidateBasic() error {
	if msg.Validator == "" {
		return ErrNotValidator
	}
	if _, err := sdk.AccAddressFromBech32(msg.Validator); err != nil {
		return ErrNotValidator
	}
	if msg.ChainID == "" {
		return ErrInvalidChain
	}
	if msg.AssetSymbol == "" {
		return ErrInvalidAsset
	}
	if msg.ExternalTxHash == "" {
		return ErrInvalidTxHash
	}
	if msg.Sender == "" {
		return ErrInvalidExternalAddress
	}
	if msg.Recipient == "" {
		return ErrInvalidRecipient
	}
	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return ErrInvalidRecipient
	}
	if msg.Amount.IsNil() || !msg.Amount.IsPositive() {
		return ErrInvalidAmount
	}
	return nil
}

// GetSigners returns the signers
func (msg MsgObserveDeposit) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Validator)
	return []sdk.AccAddress{addr}
}

// MsgAttestDeposit - Validator attests to a deposit
type MsgAttestDeposit struct {
	Validator        string   `json:"validator" yaml:"validator"`
	DepositID        uint64   `json:"deposit_id" yaml:"deposit_id"`
	Approved         bool     `json:"approved" yaml:"approved"` // true = approve, false = reject
	ObservedTxHash   string   `json:"observed_tx_hash" yaml:"observed_tx_hash"` // Validator's observed tx hash
	ObservedAmount   math.Int `json:"observed_amount" yaml:"observed_amount"`   // Validator's observed amount
	Signature        []byte   `json:"signature,omitempty" yaml:"signature,omitempty"`
}

// ValidateBasic performs basic validation
func (msg MsgAttestDeposit) ValidateBasic() error {
	if msg.Validator == "" {
		return ErrNotValidator
	}
	if _, err := sdk.AccAddressFromBech32(msg.Validator); err != nil {
		return ErrNotValidator
	}
	if msg.DepositID == 0 {
		return ErrDepositNotFound
	}
	if msg.Approved {
		// When approving, validator must provide what they observed
		if msg.ObservedTxHash == "" {
			return ErrInvalidTxHash
		}
		if msg.ObservedAmount.IsNil() || !msg.ObservedAmount.IsPositive() {
			return ErrInvalidAmount
		}
	}
	return nil
}

// GetSigners returns the signers
func (msg MsgAttestDeposit) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Validator)
	return []sdk.AccAddress{addr}
}

// MsgRequestWithdrawal - User requests withdrawal to external chain
type MsgRequestWithdrawal struct {
	Sender      string   `json:"sender" yaml:"sender"`           // ShareHODL address
	ChainID     string   `json:"chain_id" yaml:"chain_id"`       // Destination chain
	AssetSymbol string   `json:"asset_symbol" yaml:"asset_symbol"` // Asset to receive
	Recipient   string   `json:"recipient" yaml:"recipient"`     // External chain address
	Amount      math.Int `json:"amount" yaml:"amount"`           // HODL amount to burn
}

// ValidateBasic performs basic validation
func (msg MsgRequestWithdrawal) ValidateBasic() error {
	if msg.Sender == "" {
		return ErrInvalidRecipient
	}
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return ErrInvalidRecipient
	}
	if msg.ChainID == "" {
		return ErrInvalidChain
	}
	if msg.AssetSymbol == "" {
		return ErrInvalidAsset
	}
	if msg.Recipient == "" {
		return ErrInvalidExternalAddress
	}
	if msg.Amount.IsNil() || !msg.Amount.IsPositive() {
		return ErrInvalidAmount
	}
	return nil
}

// GetSigners returns the signers
func (msg MsgRequestWithdrawal) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Sender)
	return []sdk.AccAddress{addr}
}

// MsgSubmitTSSSignature - Validator submits TSS signature share
type MsgSubmitTSSSignature struct {
	Validator     string `json:"validator" yaml:"validator"`
	SessionID     uint64 `json:"session_id" yaml:"session_id"`
	SignatureData []byte `json:"signature_data" yaml:"signature_data"`
}

// ValidateBasic performs basic validation
func (msg MsgSubmitTSSSignature) ValidateBasic() error {
	if msg.Validator == "" {
		return ErrNotValidator
	}
	if _, err := sdk.AccAddressFromBech32(msg.Validator); err != nil {
		return ErrNotValidator
	}
	if msg.SessionID == 0 {
		return ErrInvalidTSSSession
	}
	if len(msg.SignatureData) == 0 {
		return ErrInvalidSignature
	}
	return nil
}

// GetSigners returns the signers
func (msg MsgSubmitTSSSignature) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Validator)
	return []sdk.AccAddress{addr}
}

// MsgUpdateCircuitBreaker - Governance updates circuit breaker
type MsgUpdateCircuitBreaker struct {
	Authority   string `json:"authority" yaml:"authority"`
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Reason      string `json:"reason" yaml:"reason"`
	CanDeposit  bool   `json:"can_deposit" yaml:"can_deposit"`
	CanWithdraw bool   `json:"can_withdraw" yaml:"can_withdraw"`
	CanAttest   bool   `json:"can_attest" yaml:"can_attest"`
}

// ValidateBasic performs basic validation
func (msg MsgUpdateCircuitBreaker) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidRecipient
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidRecipient
	}
	if msg.Enabled && msg.Reason == "" {
		return ErrInvalidAmount // Reusing error
	}
	return nil
}

// GetSigners returns the signers
func (msg MsgUpdateCircuitBreaker) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// MsgAddExternalChain - Governance adds new external chain
type MsgAddExternalChain struct {
	Authority        string   `json:"authority" yaml:"authority"`
	ChainID          string   `json:"chain_id" yaml:"chain_id"`
	ChainName        string   `json:"chain_name" yaml:"chain_name"`
	ChainType        string   `json:"chain_type" yaml:"chain_type"`
	ConfirmationsReq uint64   `json:"confirmations_required" yaml:"confirmations_required"`
	BlockTime        uint64   `json:"block_time" yaml:"block_time"`
	TSSAddress       string   `json:"tss_address" yaml:"tss_address"`
	MinDeposit       math.Int `json:"min_deposit" yaml:"min_deposit"`
	MaxDeposit       math.Int `json:"max_deposit" yaml:"max_deposit"`
}

// ValidateBasic performs basic validation
func (msg MsgAddExternalChain) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidRecipient
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidRecipient
	}
	if msg.ChainID == "" {
		return ErrInvalidChain
	}
	if msg.ChainName == "" {
		return ErrInvalidChain
	}
	if msg.ConfirmationsReq == 0 {
		return ErrInsufficientConfirmations
	}
	if msg.MinDeposit.IsNil() || msg.MinDeposit.IsNegative() {
		return ErrInvalidAmount
	}
	if msg.MaxDeposit.IsNil() || !msg.MaxDeposit.IsPositive() {
		return ErrInvalidAmount
	}
	return nil
}

// GetSigners returns the signers
func (msg MsgAddExternalChain) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// MsgAddExternalAsset - Governance adds new external asset
type MsgAddExternalAsset struct {
	Authority       string         `json:"authority" yaml:"authority"`
	ChainID         string         `json:"chain_id" yaml:"chain_id"`
	AssetSymbol     string         `json:"asset_symbol" yaml:"asset_symbol"`
	AssetName       string         `json:"asset_name" yaml:"asset_name"`
	ContractAddress string         `json:"contract_address" yaml:"contract_address"`
	Decimals        uint32         `json:"decimals" yaml:"decimals"`
	ConversionRate  math.LegacyDec `json:"conversion_rate" yaml:"conversion_rate"`
	DailyLimit      math.Int       `json:"daily_limit" yaml:"daily_limit"`
	PerTxLimit      math.Int       `json:"per_tx_limit" yaml:"per_tx_limit"`
}

// ValidateBasic performs basic validation
func (msg MsgAddExternalAsset) ValidateBasic() error {
	if msg.Authority == "" {
		return ErrInvalidRecipient
	}
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return ErrInvalidRecipient
	}
	if msg.ChainID == "" {
		return ErrInvalidChain
	}
	if msg.AssetSymbol == "" {
		return ErrInvalidAsset
	}
	if msg.Decimals == 0 || msg.Decimals > 18 {
		return ErrInvalidAmount
	}
	if msg.ConversionRate.IsNil() || !msg.ConversionRate.IsPositive() {
		return ErrConversionFailed
	}
	if msg.DailyLimit.IsNil() || !msg.DailyLimit.IsPositive() {
		return ErrInvalidAmount
	}
	return nil
}

// GetSigners returns the signers
func (msg MsgAddExternalAsset) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}
