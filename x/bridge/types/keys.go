package types

import "encoding/binary"

const (
	// ModuleName defines the module name
	ModuleName = "bridge"

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

	// BridgeReservePrefix is the prefix for reserve balances by denom
	BridgeReservePrefix = []byte{0x02}

	// SupportedAssetPrefix is the prefix for supported assets
	SupportedAssetPrefix = []byte{0x03}

	// SwapRecordPrefix is the prefix for swap records
	SwapRecordPrefix = []byte{0x04}

	// WithdrawalProposalPrefix is the prefix for withdrawal proposals
	WithdrawalProposalPrefix = []byte{0x05}

	// WithdrawalVotePrefix is the prefix for withdrawal votes
	WithdrawalVotePrefix = []byte{0x06}

	// NextSwapIDKey is the key for the next swap ID counter
	NextSwapIDKey = []byte{0x07}

	// NextProposalIDKey is the key for the next withdrawal proposal ID counter
	NextProposalIDKey = []byte{0x08}

	// TotalReserveValueKey is the key for cached total reserve value
	TotalReserveValueKey = []byte{0x09}
)

// BridgeReserveKey returns the key for a specific asset in the reserve
func BridgeReserveKey(denom string) []byte {
	return append(BridgeReservePrefix, []byte(denom)...)
}

// SupportedAssetKey returns the key for a supported asset
func SupportedAssetKey(denom string) []byte {
	return append(SupportedAssetPrefix, []byte(denom)...)
}

// SwapRecordKey returns the key for a swap record
func SwapRecordKey(swapID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, swapID)
	return append(SwapRecordPrefix, bz...)
}

// WithdrawalProposalKey returns the key for a withdrawal proposal
func WithdrawalProposalKey(proposalID uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, proposalID)
	return append(WithdrawalProposalPrefix, bz...)
}

// WithdrawalVoteKey returns the key for a withdrawal vote
func WithdrawalVoteKey(proposalID uint64, voter string) []byte {
	proposalBz := make([]byte, 8)
	binary.BigEndian.PutUint64(proposalBz, proposalID)
	key := append(WithdrawalVotePrefix, proposalBz...)
	return append(key, []byte(voter)...)
}
