package types

const (
	// ModuleName defines the module name
	ModuleName = "feeabstraction"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName

	// RouterKey defines the module's message router key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// TreasuryPoolName is the module account name for fee treasury
	TreasuryPoolName = "feeabs_treasury"
)

// Key prefixes for store
var (
	// ParamsKey is the key for storing module parameters
	ParamsKey = []byte{0x01}

	// TreasuryGrantPrefix is the prefix for treasury grants
	TreasuryGrantPrefix = []byte{0x10}

	// FeeAbstractionRecordPrefix is the prefix for fee abstraction records (audit trail)
	FeeAbstractionRecordPrefix = []byte{0x20}

	// BlockUsageKey stores the fee abstraction usage for current block (DoS protection)
	BlockUsageKey = []byte{0x30}

	// RecordCounterKey stores the next record ID
	RecordCounterKey = []byte{0x40}
)

// TreasuryGrantKey returns the key for a treasury grant
func TreasuryGrantKey(grantee string) []byte {
	return append(TreasuryGrantPrefix, []byte(grantee)...)
}

// FeeAbstractionRecordKey returns the key for a fee abstraction record
func FeeAbstractionRecordKey(id uint64) []byte {
	bz := make([]byte, 8)
	bz[0] = byte(id >> 56)
	bz[1] = byte(id >> 48)
	bz[2] = byte(id >> 40)
	bz[3] = byte(id >> 32)
	bz[4] = byte(id >> 24)
	bz[5] = byte(id >> 16)
	bz[6] = byte(id >> 8)
	bz[7] = byte(id)
	return append(FeeAbstractionRecordPrefix, bz...)
}
