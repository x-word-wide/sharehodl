package types

import (
	"fmt"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:         DefaultParams(),
		TotalSupply:    math.ZeroInt(),
		TotalCollateral: sdk.NewCoins(),
		Positions:      []CollateralPosition{},
	}
}

// Validate performs basic validation of the genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	if gs.TotalSupply.IsNegative() {
		return fmt.Errorf( "total supply cannot be negative")
	}

	if !gs.TotalCollateral.IsValid() {
		return fmt.Errorf( "total collateral")
	}

	// Validate positions
	seen := make(map[string]bool)
	for _, position := range gs.Positions {
		if seen[position.Owner] {
			return fmt.Errorf( "duplicate position for owner %s", position.Owner)
		}
		seen[position.Owner] = true

		if err := position.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// CollateralPosition represents a user's collateral position
type CollateralPosition struct {
	Owner           string      `json:"owner" yaml:"owner"`
	Collateral      sdk.Coins   `json:"collateral" yaml:"collateral"`
	MintedHODL      math.Int    `json:"minted_hodl" yaml:"minted_hodl"`
	StabilityDebt   math.LegacyDec `json:"stability_debt" yaml:"stability_debt"`
	LastUpdated     int64       `json:"last_updated" yaml:"last_updated"`
}

// NewCollateralPosition creates a new collateral position
func NewCollateralPosition(owner string, collateral sdk.Coins, mintedHODL math.Int, blockHeight int64) CollateralPosition {
	return CollateralPosition{
		Owner:         owner,
		Collateral:    collateral,
		MintedHODL:    mintedHODL,
		StabilityDebt: math.LegacyZeroDec(),
		LastUpdated:   blockHeight,
	}
}

// ProtoMessage implements proto.Message interface
func (m *CollateralPosition) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *CollateralPosition) Reset() { *m = CollateralPosition{} }

// String implements proto.Message interface
func (m *CollateralPosition) String() string {
	return m.Owner
}

// Validate validates the collateral position
func (cp CollateralPosition) Validate() error {
	_, err := sdk.AccAddressFromBech32(cp.Owner)
	if err != nil {
		return fmt.Errorf( "invalid owner address: %v", err)
	}

	if !cp.Collateral.IsValid() {
		return fmt.Errorf( "collateral")
	}

	if cp.MintedHODL.IsNegative() {
		return fmt.Errorf( "minted HODL cannot be negative")
	}

	if cp.StabilityDebt.IsNegative() {
		return fmt.Errorf( "stability debt cannot be negative")
	}

	return nil
}

// IsEmpty returns true if the position has no collateral or debt
func (cp CollateralPosition) IsEmpty() bool {
	return cp.Collateral.IsZero() && cp.MintedHODL.IsZero() && cp.StabilityDebt.IsZero()
}