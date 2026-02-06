package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "hodl"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_hodl"

	// HODLDenom is the denomination for HODL tokens
	HODLDenom = "uhodl"

	// DefaultDenom is the default denomination (alias for HODLDenom)
	DefaultDenom = HODLDenom

	// DefaultStakingTokens is the default amount of staking tokens
	DefaultStakingTokens = 100000000000 // 100,000 HODL

	// DefaultBondedTokens is the default amount of bonded tokens
	DefaultBondedTokens = 10000000000 // 10,000 HODL
)

// KVStore keys
var (
	// ParamsKey tracks module parameters
	ParamsKey = []byte{0x00}
	
	// MintedSupplyKey tracks total minted HODL supply
	MintedSupplyKey = []byte{0x01}
	
	// CollateralKey tracks collateral backing HODL
	CollateralKey = []byte{0x02}
	
	// MintRatioKey tracks collateralization ratio
	MintRatioKey = []byte{0x03}
	
	// StabilityFeeKey tracks stability fee rate
	StabilityFeeKey = []byte{0x04}
	
	// CollateralPositionPrefix is the prefix for collateral positions
	CollateralPositionPrefix = []byte{0x05}
)

// CollateralPositionKey returns the key for a collateral position
func CollateralPositionKey(owner sdk.AccAddress) []byte {
	return append(CollateralPositionPrefix, owner.Bytes()...)
}