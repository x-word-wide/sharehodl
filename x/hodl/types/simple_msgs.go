package types

import (
	"fmt"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Simple message types that bypass protobuf for now

type SimpleMsgMintHODL struct {
	Creator         string      `json:"creator"`
	CollateralCoins sdk.Coins   `json:"collateral_coins"`
	HodlAmount      math.Int    `json:"hodl_amount"`
}

func (msg SimpleMsgMintHODL) Route() string { return ModuleName }
func (msg SimpleMsgMintHODL) Type() string { return "mint_hodl" }
func (msg SimpleMsgMintHODL) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return fmt.Errorf("invalid creator address: %v", err)
	}
	if !msg.CollateralCoins.IsValid() {
		return fmt.Errorf("invalid collateral coins")
	}
	if !msg.HodlAmount.IsPositive() {
		return fmt.Errorf("HODL amount must be positive")
	}
	return nil
}

func (msg SimpleMsgMintHODL) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg SimpleMsgMintHODL) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{addr}
}

type SimpleMsgBurnHODL struct {
	Creator    string   `json:"creator"`
	HodlAmount math.Int `json:"hodl_amount"`
}

func (msg SimpleMsgBurnHODL) Route() string { return ModuleName }
func (msg SimpleMsgBurnHODL) Type() string { return "burn_hodl" }
func (msg SimpleMsgBurnHODL) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return fmt.Errorf("invalid creator address: %v", err)
	}
	if !msg.HodlAmount.IsPositive() {
		return fmt.Errorf("HODL amount must be positive")
	}
	return nil
}

func (msg SimpleMsgBurnHODL) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg SimpleMsgBurnHODL) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{addr}
}

// MsgLiquidate is used to liquidate undercollateralized positions
type MsgLiquidate struct {
	Liquidator    string `json:"liquidator"`
	PositionOwner string `json:"position_owner"`
}

func (msg MsgLiquidate) Route() string { return ModuleName }
func (msg MsgLiquidate) Type() string  { return "liquidate" }
func (msg MsgLiquidate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Liquidator); err != nil {
		return fmt.Errorf("invalid liquidator address: %v", err)
	}
	if _, err := sdk.AccAddressFromBech32(msg.PositionOwner); err != nil {
		return fmt.Errorf("invalid position owner address: %v", err)
	}
	return nil
}

func (msg MsgLiquidate) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgLiquidate) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Liquidator)
	return []sdk.AccAddress{addr}
}

// MsgLiquidateResponse is the response for liquidation
type MsgLiquidateResponse struct {
	DebtRepaid       math.Int       `json:"debt_repaid"`
	CollateralSeized math.Int       `json:"collateral_seized"`
	LiquidationRatio math.LegacyDec `json:"liquidation_ratio"`
}

// MsgAddCollateral adds collateral to an existing position
type MsgAddCollateral struct {
	Owner      string    `json:"owner"`
	Collateral sdk.Coins `json:"collateral"`
}

func (msg MsgAddCollateral) Route() string { return ModuleName }
func (msg MsgAddCollateral) Type() string  { return "add_collateral" }
func (msg MsgAddCollateral) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if !msg.Collateral.IsValid() || msg.Collateral.IsZero() {
		return fmt.Errorf("invalid collateral")
	}
	return nil
}

func (msg MsgAddCollateral) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgAddCollateral) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

// MsgAddCollateralResponse is the response for adding collateral
type MsgAddCollateralResponse struct {
	NewCollateralRatio math.LegacyDec `json:"new_collateral_ratio"`
}

// MsgWithdrawCollateral withdraws excess collateral from a position
type MsgWithdrawCollateral struct {
	Owner      string    `json:"owner"`
	Collateral sdk.Coins `json:"collateral"`
}

func (msg MsgWithdrawCollateral) Route() string { return ModuleName }
func (msg MsgWithdrawCollateral) Type() string  { return "withdraw_collateral" }
func (msg MsgWithdrawCollateral) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if !msg.Collateral.IsValid() || msg.Collateral.IsZero() {
		return fmt.Errorf("invalid collateral")
	}
	return nil
}

func (msg MsgWithdrawCollateral) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgWithdrawCollateral) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

// MsgWithdrawCollateralResponse is the response for withdrawing collateral
type MsgWithdrawCollateralResponse struct {
	NewCollateralRatio math.LegacyDec `json:"new_collateral_ratio"`
}

// MsgPayStabilityDebt pays off accumulated stability fees
type MsgPayStabilityDebt struct {
	Owner  string   `json:"owner"`
	Amount math.Int `json:"amount"`
}

func (msg MsgPayStabilityDebt) Route() string { return ModuleName }
func (msg MsgPayStabilityDebt) Type() string  { return "pay_stability_debt" }
func (msg MsgPayStabilityDebt) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
		return fmt.Errorf("invalid owner address: %v", err)
	}
	if !msg.Amount.IsPositive() {
		return fmt.Errorf("amount must be positive")
	}
	return nil
}

func (msg MsgPayStabilityDebt) GetSignBytes() []byte {
	return []byte(fmt.Sprintf("%+v", msg))
}

func (msg MsgPayStabilityDebt) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Owner)
	return []sdk.AccAddress{addr}
}

// MsgPayStabilityDebtResponse is the response for paying stability debt
type MsgPayStabilityDebtResponse struct {
	RemainingDebt math.LegacyDec `json:"remaining_debt"`
}