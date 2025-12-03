package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "validator"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_validator"
)

// Validator tier thresholds in HODL tokens
const (
	BronzeTierThreshold   = 50000   // 50K HODL
	SilverTierThreshold   = 150000  // 150K HODL  
	GoldTierThreshold     = 350000  // 350K HODL
	PlatinumTierThreshold = 750000  // 750K HODL
	DiamondTierThreshold  = 1500000 // 1.5M HODL
)

// KVStore keys
var (
	// ValidatorTierKey tracks validator tier information
	ValidatorTierPrefix = []byte{0x01}
	
	// BusinessVerificationKey tracks business verification requests
	BusinessVerificationPrefix = []byte{0x02}
	
	// VerificationRewardsKey tracks rewards earned by validators
	VerificationRewardsPrefix = []byte{0x03}
	
	// ValidatorBusinessStakeKey tracks validator stakes in businesses
	ValidatorBusinessStakePrefix = []byte{0x04}
	
	// ActiveVerificationsKey tracks ongoing verifications
	ActiveVerificationsPrefix = []byte{0x05}
)

// ValidatorTierKey returns the key for a validator tier record
func ValidatorTierKey(valAddr sdk.ValAddress) []byte {
	return append(ValidatorTierPrefix, valAddr.Bytes()...)
}

// BusinessVerificationKey returns the key for a business verification
func BusinessVerificationKey(businessID string) []byte {
	return append(BusinessVerificationPrefix, []byte(businessID)...)
}

// VerificationRewardsKey returns the key for validator verification rewards
func VerificationRewardsKey(valAddr sdk.ValAddress) []byte {
	return append(VerificationRewardsPrefix, valAddr.Bytes()...)
}

// ValidatorBusinessStakeKey returns the key for validator business stakes
func ValidatorBusinessStakeKey(valAddr sdk.ValAddress, businessID string) []byte {
	key := append(ValidatorBusinessStakePrefix, valAddr.Bytes()...)
	return append(key, []byte(businessID)...)
}

// ActiveVerificationKey returns the key for active verifications
func ActiveVerificationKey(verificationID string) []byte {
	return append(ActiveVerificationsPrefix, []byte(verificationID)...)
}