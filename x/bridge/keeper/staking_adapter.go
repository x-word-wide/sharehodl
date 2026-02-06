package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
	stakingtypes "github.com/sharehodl/sharehodl-blockchain/x/staking/types"
)

// StakingKeeperAdapter adapts the UniversalStakingKeeper to the bridge's ValidatorKeeper interface
type StakingKeeperAdapter struct {
	stakingKeeper StakingKeeper
}

// StakingKeeper defines the interface we need from the staking module
type StakingKeeper interface {
	GetUserTier(ctx sdk.Context, addr sdk.AccAddress) stakingtypes.StakeTier
	HasMinimumTier(ctx sdk.Context, addr sdk.AccAddress, minTier stakingtypes.StakeTier) bool
	IterateUserStakes(ctx sdk.Context, fn func(stake stakingtypes.UserStake) bool)
}

// NewStakingKeeperAdapter creates a new adapter
func NewStakingKeeperAdapter(stakingKeeper StakingKeeper) *StakingKeeperAdapter {
	return &StakingKeeperAdapter{
		stakingKeeper: stakingKeeper,
	}
}

// Ensure StakingKeeperAdapter implements ValidatorKeeper
var _ types.ValidatorKeeper = (*StakingKeeperAdapter)(nil)

// IsValidatorActive checks if an address has Archon tier or higher (validator-level stake)
// For bridge governance, we require Archon+ tier to participate in reserve governance
func (a *StakingKeeperAdapter) IsValidatorActive(ctx sdk.Context, addr string) bool {
	accAddr, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return false
	}

	// Require Archon tier (4) or higher for bridge governance
	// This means 100K+ HODL staked
	return a.stakingKeeper.HasMinimumTier(ctx, accAddr, stakingtypes.TierArchon)
}

// IterateValidators iterates over all users with Archon+ tier
func (a *StakingKeeperAdapter) IterateValidators(ctx sdk.Context, fn func(valAddr string) bool) {
	a.stakingKeeper.IterateUserStakes(ctx, func(stake stakingtypes.UserStake) bool {
		// Only include Archon+ tier users as "validators" for bridge governance
		if stake.Tier >= stakingtypes.TierArchon {
			return fn(stake.Owner)
		}
		return false
	})
}
