package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "equity"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key  
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_equity"
)

// Store keys
var (
	// CompanyPrefix stores Company information
	CompanyPrefix = []byte{0x01}
	
	// ShareClassPrefix stores share class information
	ShareClassPrefix = []byte{0x02}
	
	// ShareholdingPrefix stores individual shareholdings
	ShareholdingPrefix = []byte{0x03}
	
	// VestingPrefix stores vesting schedules
	VestingPrefix = []byte{0x05}
	
	// CompanyCounterKey stores the global company counter
	CompanyCounterKey = []byte{0x06}
	
	// Business listing prefixes
	BusinessListingPrefix = []byte{0x07}
	ListingCounterKey     = []byte{0x08}
	ListingRequirementsKey = []byte{0x09}
	ValidatorAssignmentPrefix = []byte{0x0A}
	
	// Dividend system prefixes
	DividendPrefix = []byte{0x04}
	DividendCounterKey = []byte{0x0C}
	DividendPaymentPrefix = []byte{0x0D}
	DividendClaimPrefix = []byte{0x0E}
	DividendSnapshotPrefix = []byte{0x0F}
	ShareholderSnapshotPrefix = []byte{0x10}
	DividendPolicyPrefix = []byte{0x11}
	DividendStatisticsPrefix = []byte{0x12}
	DividendPaymentCounterKey = []byte{0x13}
	DividendClaimCounterKey = []byte{0x14}
)

// GetCompanyKey returns the store key for a company
func GetCompanyKey(companyID uint64) []byte {
	return append(CompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetShareClassKey returns the store key for a share class
func GetShareClassKey(companyID uint64, classID string) []byte {
	key := append(ShareClassPrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, []byte(classID)...)
}

// GetShareholdingKey returns the store key for a shareholding
func GetShareholdingKey(companyID uint64, classID string, owner string) []byte {
	key := append(ShareholdingPrefix, sdk.Uint64ToBigEndian(companyID)...)
	key = append(key, []byte(classID)...)
	key = append(key, []byte("|")...) // separator
	return append(key, []byte(owner)...)
}


// GetVestingKey returns the store key for a vesting schedule
func GetVestingKey(companyID uint64, classID string, owner string) []byte {
	key := append(VestingPrefix, sdk.Uint64ToBigEndian(companyID)...)
	key = append(key, []byte(classID)...)
	key = append(key, []byte("|")...) // separator
	return append(key, []byte(owner)...)
}

// GetBusinessListingKey returns the store key for a business listing
func GetBusinessListingKey(listingID uint64) []byte {
	return append(BusinessListingPrefix, sdk.Uint64ToBigEndian(listingID)...)
}

// GetValidatorAssignmentKey returns the store key for validator assignment
func GetValidatorAssignmentKey(listingID uint64, validator string) []byte {
	key := append(ValidatorAssignmentPrefix, sdk.Uint64ToBigEndian(listingID)...)
	return append(key, []byte(validator)...)
}

// GetDividendKey returns the store key for a dividend
func GetDividendKey(dividendID uint64) []byte {
	return append(DividendPrefix, sdk.Uint64ToBigEndian(dividendID)...)
}

// GetDividendPaymentKey returns the store key for a dividend payment
func GetDividendPaymentKey(paymentID uint64) []byte {
	return append(DividendPaymentPrefix, sdk.Uint64ToBigEndian(paymentID)...)
}

// GetDividendClaimKey returns the store key for a dividend claim
func GetDividendClaimKey(claimID uint64) []byte {
	return append(DividendClaimPrefix, sdk.Uint64ToBigEndian(claimID)...)
}

// GetDividendSnapshotKey returns the store key for a dividend snapshot
func GetDividendSnapshotKey(dividendID uint64) []byte {
	return append(DividendSnapshotPrefix, sdk.Uint64ToBigEndian(dividendID)...)
}

// GetShareholderSnapshotKey returns the store key for a shareholder snapshot
func GetShareholderSnapshotKey(dividendID uint64, shareholder string) []byte {
	key := append(ShareholderSnapshotPrefix, sdk.Uint64ToBigEndian(dividendID)...)
	return append(key, []byte(shareholder)...)
}

// GetDividendPolicyKey returns the store key for dividend policy
func GetDividendPolicyKey(companyID uint64) []byte {
	return append(DividendPolicyPrefix, sdk.Uint64ToBigEndian(companyID)...)
}

// GetDividendStatisticsKey returns the store key for dividend statistics
func GetDividendStatisticsKey(companyID uint64, classID string) []byte {
	key := append(DividendStatisticsPrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, []byte(classID)...)
}