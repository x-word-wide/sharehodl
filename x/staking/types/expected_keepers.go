package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

// GovernanceKeeper defines the expected governance keeper interface
// NOTE: Optional - can be nil. Methods will be added to governance keeper later.
type GovernanceKeeper interface {
	// Placeholder - methods to be defined when governance integration is complete
}

// FeeAbstractionKeeper defines the expected fee abstraction keeper interface
// NOTE: Optional - can be nil if not using fee abstraction integration.
type FeeAbstractionKeeper interface {
	// Get treasury balance available for distribution
	GetTreasuryBalance(ctx sdk.Context) math.Int
	// Withdraw from treasury for rewards
	WithdrawFromTreasury(ctx sdk.Context, recipient sdk.AccAddress, amount math.Int) error
}

// StakingHooks defines hooks for other modules to react to staking events
type StakingHooks interface {
	// Called when a user stakes
	AfterStake(ctx sdk.Context, staker sdk.AccAddress, amount math.Int, tier StakeTier) error
	// Called when a user unstakes
	AfterUnstake(ctx sdk.Context, staker sdk.AccAddress, amount math.Int, oldTier StakeTier) error
	// Called when tier changes
	AfterTierChange(ctx sdk.Context, staker sdk.AccAddress, oldTier, newTier StakeTier) error
	// Called when reputation changes
	AfterReputationChange(ctx sdk.Context, staker sdk.AccAddress, oldScore, newScore math.LegacyDec) error
	// Called when user is slashed
	AfterSlash(ctx sdk.Context, staker sdk.AccAddress, amount math.Int, reason string) error
}

// MultiStakingHooks combines multiple hooks
type MultiStakingHooks []StakingHooks

func (h MultiStakingHooks) AfterStake(ctx sdk.Context, staker sdk.AccAddress, amount math.Int, tier StakeTier) error {
	for _, hook := range h {
		if err := hook.AfterStake(ctx, staker, amount, tier); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiStakingHooks) AfterUnstake(ctx sdk.Context, staker sdk.AccAddress, amount math.Int, oldTier StakeTier) error {
	for _, hook := range h {
		if err := hook.AfterUnstake(ctx, staker, amount, oldTier); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiStakingHooks) AfterTierChange(ctx sdk.Context, staker sdk.AccAddress, oldTier, newTier StakeTier) error {
	for _, hook := range h {
		if err := hook.AfterTierChange(ctx, staker, oldTier, newTier); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiStakingHooks) AfterReputationChange(ctx sdk.Context, staker sdk.AccAddress, oldScore, newScore math.LegacyDec) error {
	for _, hook := range h {
		if err := hook.AfterReputationChange(ctx, staker, oldScore, newScore); err != nil {
			return err
		}
	}
	return nil
}

func (h MultiStakingHooks) AfterSlash(ctx sdk.Context, staker sdk.AccAddress, amount math.Int, reason string) error {
	for _, hook := range h {
		if err := hook.AfterSlash(ctx, staker, amount, reason); err != nil {
			return err
		}
	}
	return nil
}
