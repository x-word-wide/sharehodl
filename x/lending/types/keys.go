package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "lending"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_lending"
)

// Store keys
var (
	// LoanPrefix stores loan data
	LoanPrefix = []byte{0x01}

	// LendingPoolPrefix stores lending pool data
	LendingPoolPrefix = []byte{0x02}

	// PoolDepositPrefix stores pool deposits
	PoolDepositPrefix = []byte{0x03}

	// LoanOfferPrefix stores loan offers
	LoanOfferPrefix = []byte{0x04}

	// LoanRequestPrefix stores loan requests
	LoanRequestPrefix = []byte{0x05}

	// LoanCounterKey stores the global loan counter
	LoanCounterKey = []byte{0x06}

	// PoolCounterKey stores the global pool counter
	PoolCounterKey = []byte{0x07}

	// UserLoansPrefix stores user loan indices
	UserLoansPrefix = []byte{0x08}

	// CollateralPricePrefix stores collateral prices
	CollateralPricePrefix = []byte{0x09}

	// ParamsKey stores module parameters
	ParamsKey = []byte{0x0A}

	// Stake-as-Trust-Ceiling: P2P lending stake prefixes

	// LenderStakePrefix stores lender security deposits
	LenderStakePrefix = []byte{0x10}

	// BorrowerStakePrefix stores borrower security deposits
	BorrowerStakePrefix = []byte{0x11}

	// OfferCounterKey stores the global offer counter
	OfferCounterKey = []byte{0x12}

	// RequestCounterKey stores the global request counter
	RequestCounterKey = []byte{0x13}

	// LenderOffersPrefix indexes offers by lender
	LenderOffersPrefix = []byte{0x14}

	// BorrowerRequestsPrefix indexes requests by borrower
	BorrowerRequestsPrefix = []byte{0x15}
)

// GetLoanKey returns the store key for a loan
func GetLoanKey(loanID uint64) []byte {
	return append(LoanPrefix, sdk.Uint64ToBigEndian(loanID)...)
}

// GetLendingPoolKey returns the store key for a lending pool
func GetLendingPoolKey(poolID uint64) []byte {
	return append(LendingPoolPrefix, sdk.Uint64ToBigEndian(poolID)...)
}

// GetPoolDepositKey returns the store key for a pool deposit
func GetPoolDepositKey(poolID uint64, user string) []byte {
	key := append(PoolDepositPrefix, sdk.Uint64ToBigEndian(poolID)...)
	return append(key, []byte(user)...)
}

// GetLoanOfferKey returns the store key for a loan offer
func GetLoanOfferKey(offerID uint64) []byte {
	return append(LoanOfferPrefix, sdk.Uint64ToBigEndian(offerID)...)
}

// GetLoanRequestKey returns the store key for a loan request
func GetLoanRequestKey(requestID uint64) []byte {
	return append(LoanRequestPrefix, sdk.Uint64ToBigEndian(requestID)...)
}

// GetUserLoansKey returns the store key for user's loans
func GetUserLoansKey(user string) []byte {
	return append(UserLoansPrefix, []byte(user)...)
}

// GetCollateralPriceKey returns the store key for collateral price
func GetCollateralPriceKey(denom string) []byte {
	return append(CollateralPricePrefix, []byte(denom)...)
}

// Stake-as-Trust-Ceiling: Key helper functions

// GetLenderStakeKey returns the store key for a lender's stake
func GetLenderStakeKey(address string) []byte {
	return append(LenderStakePrefix, []byte(address)...)
}

// GetBorrowerStakeKey returns the store key for a borrower's stake
func GetBorrowerStakeKey(address string) []byte {
	return append(BorrowerStakePrefix, []byte(address)...)
}

// GetLenderOffersKey returns the store key for a lender's offer index
func GetLenderOffersKey(lender string, offerID uint64) []byte {
	key := append(LenderOffersPrefix, []byte(lender)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(offerID)...)
}

// GetLenderOffersPrefixKey returns the prefix for all offers by a lender
func GetLenderOffersPrefixKey(lender string) []byte {
	key := append(LenderOffersPrefix, []byte(lender)...)
	return append(key, []byte(":")...)
}

// GetBorrowerRequestsKey returns the store key for a borrower's request index
func GetBorrowerRequestsKey(borrower string, requestID uint64) []byte {
	key := append(BorrowerRequestsPrefix, []byte(borrower)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(requestID)...)
}

// GetBorrowerRequestsPrefixKey returns the prefix for all requests by a borrower
func GetBorrowerRequestsPrefixKey(borrower string) []byte {
	key := append(BorrowerRequestsPrefix, []byte(borrower)...)
	return append(key, []byte(":")...)
}
