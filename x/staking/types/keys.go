package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "universalstaking"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName

	// StakingPoolName is the module account for staked tokens
	StakingPoolName = "staking_pool"

	// RewardsPoolName is the module account for pending rewards
	RewardsPoolName = "rewards_pool"
)

// Key prefixes
var (
	ParamsKey                  = []byte{0x01}
	UserStakePrefix            = []byte{0x10}
	TierCountPrefix            = []byte{0x20}
	ReputationHistoryPrefix    = []byte{0x30}
	EpochInfoKey               = []byte{0x40}
	PendingRewardsPrefix       = []byte{0x50}
	UserLocksPrefix            = []byte{0x60} // Stake locks for users
	UnbondingRequestPrefix     = []byte{0x70} // Pending unstake requests
	InheritanceUnbondingPrefix = []byte{0x80} // Inheritance unbonding requests (recipient != owner)
	UserCommitmentsPrefix      = []byte{0x90} // Stake-as-Trust-Ceiling commitments
)

// GetUserStakeKey returns the key for a user's stake
func GetUserStakeKey(addr sdk.AccAddress) []byte {
	return append(UserStakePrefix, addr.Bytes()...)
}

// GetTierCountKey returns the key for tier member count
func GetTierCountKey(tier StakeTier) []byte {
	return append(TierCountPrefix, byte(tier))
}

// GetReputationHistoryKey returns the key for reputation history
func GetReputationHistoryKey(addr sdk.AccAddress) []byte {
	return append(ReputationHistoryPrefix, addr.Bytes()...)
}

// GetPendingRewardsKey returns the key for pending rewards
func GetPendingRewardsKey(addr sdk.AccAddress) []byte {
	return append(PendingRewardsPrefix, addr.Bytes()...)
}

// GetUserLocksKey returns the key for a user's stake locks
func GetUserLocksKey(addr sdk.AccAddress) []byte {
	return append(UserLocksPrefix, addr.Bytes()...)
}

// GetUnbondingRequestKey returns the key for a user's unbonding request
func GetUnbondingRequestKey(addr sdk.AccAddress) []byte {
	return append(UnbondingRequestPrefix, addr.Bytes()...)
}

// GetInheritanceUnbondingKey returns the key for an inheritance unbonding request
func GetInheritanceUnbondingKey(addr sdk.AccAddress) []byte {
	return append(InheritanceUnbondingPrefix, addr.Bytes()...)
}

// GetUserCommitmentsKey returns the key for a user's stake commitments
func GetUserCommitmentsKey(addr sdk.AccAddress) []byte {
	return append(UserCommitmentsPrefix, addr.Bytes()...)
}
