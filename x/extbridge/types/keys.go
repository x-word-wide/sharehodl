package types

import (
	"encoding/binary"
)

const (
	// ModuleName defines the module name
	ModuleName = "extbridge"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x01}

	// ExternalChainPrefix is the prefix for external chain configs
	ExternalChainPrefix = []byte{0x02}

	// ExternalAssetPrefix is the prefix for external asset configs
	ExternalAssetPrefix = []byte{0x03}

	// DepositPrefix is the prefix for deposit records
	DepositPrefix = []byte{0x04}

	// AttestationPrefix is the prefix for deposit attestations
	AttestationPrefix = []byte{0x05}

	// WithdrawalPrefix is the prefix for withdrawal records
	WithdrawalPrefix = []byte{0x06}

	// TSSSessionPrefix is the prefix for TSS signing sessions
	TSSSessionPrefix = []byte{0x07}

	// RateLimitPrefix is the prefix for rate limit tracking
	RateLimitPrefix = []byte{0x08}

	// NextDepositIDKey is the key for the next deposit ID counter
	NextDepositIDKey = []byte{0x09}

	// NextWithdrawalIDKey is the key for the next withdrawal ID counter
	NextWithdrawalIDKey = []byte{0x0A}

	// NextTSSSessionIDKey is the key for the next TSS session ID counter
	NextTSSSessionIDKey = []byte{0x0B}

	// CircuitBreakerKey is the key for circuit breaker status
	CircuitBreakerKey = []byte{0x0C}

	// TotalBridgedPrefix is the prefix for total bridged amount per asset
	TotalBridgedPrefix = []byte{0x0D}
)

// ExternalChainKey returns the key for an external chain config
func ExternalChainKey(chainID string) []byte {
	return append(ExternalChainPrefix, []byte(chainID)...)
}

// ExternalAssetKey returns the key for an external asset config
func ExternalAssetKey(chainID string, assetSymbol string) []byte {
	key := append(ExternalAssetPrefix, []byte(chainID)...)
	key = append(key, []byte("/")...)
	return append(key, []byte(assetSymbol)...)
}

// DepositKey returns the key for a deposit record
func DepositKey(depositID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, depositID)
	return append(DepositPrefix, bz...)
}

// DepositByTxHashKey returns the key for deposit lookup by external tx hash
func DepositByTxHashKey(chainID string, txHash string) []byte {
	key := append([]byte{0x14}, []byte(chainID)...)
	key = append(key, []byte("/")...)
	return append(key, []byte(txHash)...)
}

// AttestationKey returns the key for a deposit attestation
func AttestationKey(depositID uint64, validator string) []byte {
	depositBz := make([]byte, 8)
	binary.BigEndian.PutUint64(depositBz, depositID)
	key := append(AttestationPrefix, depositBz...)
	key = append(key, []byte("/")...)
	return append(key, []byte(validator)...)
}

// WithdrawalKey returns the key for a withdrawal record
func WithdrawalKey(withdrawalID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, withdrawalID)
	return append(WithdrawalPrefix, bz...)
}

// TSSSessionKey returns the key for a TSS session
func TSSSessionKey(sessionID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, sessionID)
	return append(TSSSessionPrefix, bz...)
}

// RateLimitKey returns the key for rate limit tracking
func RateLimitKey(chainID string, assetSymbol string, windowStart int64) []byte {
	key := append(RateLimitPrefix, []byte(chainID)...)
	key = append(key, []byte("/")...)
	key = append(key, []byte(assetSymbol)...)
	key = append(key, []byte("/")...)
	windowBz := make([]byte, 8)
	binary.BigEndian.PutUint64(windowBz, uint64(windowStart))
	return append(key, windowBz...)
}

// TotalBridgedKey returns the key for total bridged amount
func TotalBridgedKey(chainID string, assetSymbol string) []byte {
	key := append(TotalBridgedPrefix, []byte(chainID)...)
	key = append(key, []byte("/")...)
	return append(key, []byte(assetSymbol)...)
}
