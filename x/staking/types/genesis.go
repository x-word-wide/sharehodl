package types

import (
	"cosmossdk.io/math"
)

// DefaultGenesisState returns default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		UserStakes: []UserStake{},
		EpochInfo: EpochInfo{
			CurrentEpoch:     0,
			TotalDistributed: math.ZeroInt(),
		},
	}
}

// GenesisState defines the universal staking module's genesis state
type GenesisState struct {
	Params     Params      `json:"params"`
	UserStakes []UserStake `json:"user_stakes"`
	EpochInfo  EpochInfo   `json:"epoch_info"`
}

// ProtoMessage implements proto.Message interface
func (m *GenesisState) ProtoMessage() {}

// Reset implements proto.Message interface
func (m *GenesisState) Reset() { *m = GenesisState{} }

// String implements proto.Message interface
func (m *GenesisState) String() string { return "staking_genesis" }

// Validate validates genesis state
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate user stakes
	seen := make(map[string]bool)
	for _, stake := range gs.UserStakes {
		if seen[stake.Owner] {
			return ErrAlreadyStaked
		}
		seen[stake.Owner] = true

		if !stake.StakedAmount.IsPositive() {
			return ErrInvalidAmount
		}
	}

	return nil
}
