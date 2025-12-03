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