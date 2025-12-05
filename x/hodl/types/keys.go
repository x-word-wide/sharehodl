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