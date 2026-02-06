package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "inheritance"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_inheritance"
)

// KVStore keys
var (
	// ParamsKey tracks module parameters
	ParamsKey = []byte{0x00}

	// PlanCounterKey tracks the next plan ID
	PlanCounterKey = []byte{0x01}

	// InheritancePlanPrefix is the prefix for inheritance plans
	// Key: 0x02 | PlanID (8 bytes) -> InheritancePlan
	InheritancePlanPrefix = []byte{0x02}

	// OwnerPlanIndexPrefix indexes plans by owner address
	// Key: 0x03 | OwnerAddress | PlanID (8 bytes) -> []byte{1}
	OwnerPlanIndexPrefix = []byte{0x03}

	// BeneficiaryPlanIndexPrefix indexes plans by beneficiary address
	// Key: 0x04 | BeneficiaryAddress | PlanID (8 bytes) -> []byte{1}
	BeneficiaryPlanIndexPrefix = []byte{0x04}

	// LastActivityPrefix tracks last activity time for addresses
	// Key: 0x05 | OwnerAddress -> ActivityRecord
	LastActivityPrefix = []byte{0x05}

	// ActiveTriggerPrefix stores triggered switches pending claims
	// Key: 0x06 | PlanID (8 bytes) -> SwitchTrigger
	ActiveTriggerPrefix = []byte{0x06}

	// ClaimPrefix stores beneficiary claims
	// Key: 0x07 | PlanID (8 bytes) | BeneficiaryAddress -> BeneficiaryClaim
	ClaimPrefix = []byte{0x07}

	// LockedAssetPrefix stores assets locked during inheritance process
	// Key: 0x08 | PlanID (8 bytes) -> LockedAssetDuringInheritance
	LockedAssetPrefix = []byte{0x08}

	// ClaimLockPrefix stores claim locks to prevent double-claim race conditions
	// Key: 0x09 | PlanID (8 bytes) -> []byte{1}
	ClaimLockPrefix = []byte{0x09}
)

// InheritancePlanKey returns the key for an inheritance plan
func InheritancePlanKey(planID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, planID)
	return append(InheritancePlanPrefix, bz...)
}

// OwnerPlanIndexKey returns the index key for owner->plans mapping
func OwnerPlanIndexKey(owner sdk.AccAddress, planID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, planID)
	key := append(OwnerPlanIndexPrefix, owner.Bytes()...)
	return append(key, bz...)
}

// OwnerPlanIndexPrefix returns the prefix for all plans owned by an address
func OwnerPlanIndexPrefixForAddress(owner sdk.AccAddress) []byte {
	return append(OwnerPlanIndexPrefix, owner.Bytes()...)
}

// BeneficiaryPlanIndexKey returns the index key for beneficiary->plans mapping
func BeneficiaryPlanIndexKey(beneficiary sdk.AccAddress, planID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, planID)
	key := append(BeneficiaryPlanIndexPrefix, beneficiary.Bytes()...)
	return append(key, bz...)
}

// BeneficiaryPlanIndexPrefixForAddress returns the prefix for all plans with a beneficiary
func BeneficiaryPlanIndexPrefixForAddress(beneficiary sdk.AccAddress) []byte {
	return append(BeneficiaryPlanIndexPrefix, beneficiary.Bytes()...)
}

// LastActivityKey returns the key for tracking last activity
func LastActivityKey(owner sdk.AccAddress) []byte {
	return append(LastActivityPrefix, owner.Bytes()...)
}

// ActiveTriggerKey returns the key for an active switch trigger
func ActiveTriggerKey(planID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, planID)
	return append(ActiveTriggerPrefix, bz...)
}

// ClaimKey returns the key for a beneficiary claim
func ClaimKey(planID uint64, beneficiary sdk.AccAddress) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, planID)
	key := append(ClaimPrefix, bz...)
	return append(key, beneficiary.Bytes()...)
}

// ClaimPrefixForPlan returns the prefix for all claims for a plan
func ClaimPrefixForPlan(planID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, planID)
	return append(ClaimPrefix, bz...)
}

// LockedAssetKey returns the key for locked assets
func LockedAssetKey(planID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, planID)
	return append(LockedAssetPrefix, bz...)
}

// ClaimLockKey returns the key for a claim lock
func ClaimLockKey(planID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, planID)
	return append(ClaimLockPrefix, bz...)
}
