package types

import (
	"encoding/binary"
)

// Module constants
const (
	// ModuleName defines the module name
	ModuleName = "governance"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_governance"
)

// Storage key prefixes
var (
	// Proposal storage
	ProposalPrefix              = []byte{0x01}
	ProposalIDCounterPrefix     = []byte{0x02}
	
	// Voting storage
	VotePrefix                  = []byte{0x10}
	WeightedVotePrefix          = []byte{0x11}
	VotingPowerPrefix          = []byte{0x12}
	TallyResultPrefix          = []byte{0x13}
	
	// Deposit storage
	DepositPrefix              = []byte{0x20}
	TotalDepositPrefix         = []byte{0x21}
	
	// Governance parameters
	ParamsPrefix               = []byte{0x30}
	
	// Proposal execution
	ExecutionPrefix            = []byte{0x40}
	ExecutionQueuePrefix       = []byte{0x41}
	
	// Indexing and queries
	ProposalByStatusPrefix     = []byte{0x50}
	ProposalBySubmitterPrefix  = []byte{0x51}
	ProposalByTypePrefix       = []byte{0x52}
	
	// Company and validator specific
	CompanyProposalPrefix      = []byte{0x60}
	ValidatorProposalPrefix    = []byte{0x61}
	
	// Emergency proposals
	EmergencyProposalPrefix    = []byte{0x70}
	EmergencyTypeIndexPrefix   = []byte{0x71}
	EmergencySeverityPrefix    = []byte{0x72}
	
	// Delegation
	DelegationPrefix           = []byte{0x80}
	DelegateIndexPrefix        = []byte{0x81}
	DelegatorIndexPrefix       = []byte{0x82}
	CompanyDelegationPrefix    = []byte{0x83}
	
	// Statistics and analytics
	GovernanceStatsPrefix      = []byte{0x90}
	ParticipationStatsPrefix   = []byte{0x91}

	// Time-based proposal queues for efficient end block processing
	// Key format: DepositEndTimePrefix + endTime (seconds) + proposalID
	DepositEndTimePrefix       = []byte{0xA0}
	// Key format: VotingEndTimePrefix + endTime (seconds) + proposalID
	VotingEndTimePrefix        = []byte{0xA1}

	// ==========================================================================
	// ANTI-SPAM RATE LIMITING KEYS
	// ==========================================================================

	// ProposerDailyCountPrefix stores proposal count per address per day
	// Key format: ProposerDailyCountPrefix + address + dayTimestamp
	ProposerDailyCountPrefix   = []byte{0xB0}

	// ProposerLastSubmitPrefix stores the timestamp of last proposal submission
	// Key format: ProposerLastSubmitPrefix + address
	ProposerLastSubmitPrefix   = []byte{0xB1}

	// BurnedFeesPrefix stores total fees burned from proposals
	BurnedFeesPrefix           = []byte{0xB2}
)

// Key construction functions

// ProposalKey returns the store key for a proposal
func ProposalKey(proposalID uint64) []byte {
	key := make([]byte, len(ProposalPrefix)+8)
	copy(key, ProposalPrefix)
	binary.BigEndian.PutUint64(key[len(ProposalPrefix):], proposalID)
	return key
}

// VoteKey returns the store key for a vote
func VoteKey(proposalID uint64, voter []byte) []byte {
	key := make([]byte, len(VotePrefix)+8+len(voter))
	copy(key, VotePrefix)
	binary.BigEndian.PutUint64(key[len(VotePrefix):], proposalID)
	copy(key[len(VotePrefix)+8:], voter)
	return key
}

// WeightedVoteKey returns the store key for a weighted vote
func WeightedVoteKey(proposalID uint64, voter []byte) []byte {
	key := make([]byte, len(WeightedVotePrefix)+8+len(voter))
	copy(key, WeightedVotePrefix)
	binary.BigEndian.PutUint64(key[len(WeightedVotePrefix):], proposalID)
	copy(key[len(WeightedVotePrefix)+8:], voter)
	return key
}

// DepositKey returns the store key for a deposit
func DepositKey(proposalID uint64, depositor []byte) []byte {
	key := make([]byte, len(DepositPrefix)+8+len(depositor))
	copy(key, DepositPrefix)
	binary.BigEndian.PutUint64(key[len(DepositPrefix):], proposalID)
	copy(key[len(DepositPrefix)+8:], depositor)
	return key
}

// TotalDepositKey returns the store key for total deposits of a proposal
func TotalDepositKey(proposalID uint64) []byte {
	key := make([]byte, len(TotalDepositPrefix)+8)
	copy(key, TotalDepositPrefix)
	binary.BigEndian.PutUint64(key[len(TotalDepositPrefix):], proposalID)
	return key
}

// TallyResultKey returns the store key for tally results
func TallyResultKey(proposalID uint64) []byte {
	key := make([]byte, len(TallyResultPrefix)+8)
	copy(key, TallyResultPrefix)
	binary.BigEndian.PutUint64(key[len(TallyResultPrefix):], proposalID)
	return key
}

// VotingPowerKey returns the store key for voting power at a specific height
func VotingPowerKey(proposalID uint64, voter []byte) []byte {
	key := make([]byte, len(VotingPowerPrefix)+8+len(voter))
	copy(key, VotingPowerPrefix)
	binary.BigEndian.PutUint64(key[len(VotingPowerPrefix):], proposalID)
	copy(key[len(VotingPowerPrefix)+8:], voter)
	return key
}

// ExecutionKey returns the store key for proposal execution data
func ExecutionKey(proposalID uint64) []byte {
	key := make([]byte, len(ExecutionPrefix)+8)
	copy(key, ExecutionPrefix)
	binary.BigEndian.PutUint64(key[len(ExecutionPrefix):], proposalID)
	return key
}

// ProposalByStatusKey returns the store key for indexing proposals by status
func ProposalByStatusKey(status string, proposalID uint64) []byte {
	statusBytes := []byte(status)
	key := make([]byte, len(ProposalByStatusPrefix)+len(statusBytes)+1+8)
	copy(key, ProposalByStatusPrefix)
	copy(key[len(ProposalByStatusPrefix):], statusBytes)
	key[len(ProposalByStatusPrefix)+len(statusBytes)] = 0x00 // separator
	binary.BigEndian.PutUint64(key[len(ProposalByStatusPrefix)+len(statusBytes)+1:], proposalID)
	return key
}

// ProposalBySubmitterKey returns the store key for indexing proposals by submitter
func ProposalBySubmitterKey(submitter []byte, proposalID uint64) []byte {
	key := make([]byte, len(ProposalBySubmitterPrefix)+len(submitter)+8)
	copy(key, ProposalBySubmitterPrefix)
	copy(key[len(ProposalBySubmitterPrefix):], submitter)
	binary.BigEndian.PutUint64(key[len(ProposalBySubmitterPrefix)+len(submitter):], proposalID)
	return key
}

// ProposalByTypeKey returns the store key for indexing proposals by type
func ProposalByTypeKey(proposalType string, proposalID uint64) []byte {
	typeBytes := []byte(proposalType)
	key := make([]byte, len(ProposalByTypePrefix)+len(typeBytes)+1+8)
	copy(key, ProposalByTypePrefix)
	copy(key[len(ProposalByTypePrefix):], typeBytes)
	key[len(ProposalByTypePrefix)+len(typeBytes)] = 0x00 // separator
	binary.BigEndian.PutUint64(key[len(ProposalByTypePrefix)+len(typeBytes)+1:], proposalID)
	return key
}

// CompanyProposalKey returns the store key for company-specific proposals
func CompanyProposalKey(companyID uint64, proposalID uint64) []byte {
	key := make([]byte, len(CompanyProposalPrefix)+16)
	copy(key, CompanyProposalPrefix)
	binary.BigEndian.PutUint64(key[len(CompanyProposalPrefix):], companyID)
	binary.BigEndian.PutUint64(key[len(CompanyProposalPrefix)+8:], proposalID)
	return key
}

// ValidatorProposalKey returns the store key for validator-specific proposals
func ValidatorProposalKey(validatorAddr []byte, proposalID uint64) []byte {
	key := make([]byte, len(ValidatorProposalPrefix)+len(validatorAddr)+8)
	copy(key, ValidatorProposalPrefix)
	copy(key[len(ValidatorProposalPrefix):], validatorAddr)
	binary.BigEndian.PutUint64(key[len(ValidatorProposalPrefix)+len(validatorAddr):], proposalID)
	return key
}

// Iterator helper functions

// ProposalVoteIteratorKey returns the prefix for iterating over votes for a proposal
func ProposalVoteIteratorKey(proposalID uint64) []byte {
	key := make([]byte, len(VotePrefix)+8)
	copy(key, VotePrefix)
	binary.BigEndian.PutUint64(key[len(VotePrefix):], proposalID)
	return key
}

// ProposalWeightedVoteIteratorKey returns the prefix for iterating over weighted votes for a proposal
func ProposalWeightedVoteIteratorKey(proposalID uint64) []byte {
	key := make([]byte, len(WeightedVotePrefix)+8)
	copy(key, WeightedVotePrefix)
	binary.BigEndian.PutUint64(key[len(WeightedVotePrefix):], proposalID)
	return key
}

// ProposalDepositIteratorKey returns the prefix for iterating over deposits for a proposal
func ProposalDepositIteratorKey(proposalID uint64) []byte {
	key := make([]byte, len(DepositPrefix)+8)
	copy(key, DepositPrefix)
	binary.BigEndian.PutUint64(key[len(DepositPrefix):], proposalID)
	return key
}

// StatusIteratorKey returns the prefix for iterating over proposals by status
func StatusIteratorKey(status string) []byte {
	statusBytes := []byte(status)
	key := make([]byte, len(ProposalByStatusPrefix)+len(statusBytes)+1)
	copy(key, ProposalByStatusPrefix)
	copy(key[len(ProposalByStatusPrefix):], statusBytes)
	key[len(ProposalByStatusPrefix)+len(statusBytes)] = 0x00 // separator
	return key
}

// SubmitterIteratorKey returns the prefix for iterating over proposals by submitter
func SubmitterIteratorKey(submitter []byte) []byte {
	key := make([]byte, len(ProposalBySubmitterPrefix)+len(submitter))
	copy(key, ProposalBySubmitterPrefix)
	copy(key[len(ProposalBySubmitterPrefix):], submitter)
	return key
}

// TypeIteratorKey returns the prefix for iterating over proposals by type
func TypeIteratorKey(proposalType string) []byte {
	typeBytes := []byte(proposalType)
	key := make([]byte, len(ProposalByTypePrefix)+len(typeBytes)+1)
	copy(key, ProposalByTypePrefix)
	copy(key[len(ProposalByTypePrefix):], typeBytes)
	key[len(ProposalByTypePrefix)+len(typeBytes)] = 0x00 // separator
	return key
}

// CompanyIteratorKey returns the prefix for iterating over proposals by company
func CompanyIteratorKey(companyID uint64) []byte {
	key := make([]byte, len(CompanyProposalPrefix)+8)
	copy(key, CompanyProposalPrefix)
	binary.BigEndian.PutUint64(key[len(CompanyProposalPrefix):], companyID)
	return key
}

// ValidatorIteratorKey returns the prefix for iterating over proposals by validator
func ValidatorIteratorKey(validatorAddr []byte) []byte {
	key := make([]byte, len(ValidatorProposalPrefix)+len(validatorAddr))
	copy(key, ValidatorProposalPrefix)
	copy(key[len(ValidatorProposalPrefix):], validatorAddr)
	return key
}

// GetProposalIDFromKey extracts proposal ID from various key types
func GetProposalIDFromKey(key []byte, prefixLen int) uint64 {
	if len(key) < prefixLen+8 {
		return 0
	}
	return binary.BigEndian.Uint64(key[prefixLen:prefixLen+8])
}

// GetCompanyIDFromKey extracts company ID from company proposal key
func GetCompanyIDFromKey(key []byte) uint64 {
	if len(key) < len(CompanyProposalPrefix)+8 {
		return 0
	}
	return binary.BigEndian.Uint64(key[len(CompanyProposalPrefix):len(CompanyProposalPrefix)+8])
}

// SplitVoteKey extracts proposal ID and voter address from vote key
func SplitVoteKey(key []byte) (uint64, []byte) {
	if len(key) < len(VotePrefix)+8 {
		return 0, nil
	}
	proposalID := binary.BigEndian.Uint64(key[len(VotePrefix):len(VotePrefix)+8])
	voter := key[len(VotePrefix)+8:]
	return proposalID, voter
}

// SplitDepositKey extracts proposal ID and depositor address from deposit key
func SplitDepositKey(key []byte) (uint64, []byte) {
	if len(key) < len(DepositPrefix)+8 {
		return 0, nil
	}
	proposalID := binary.BigEndian.Uint64(key[len(DepositPrefix):len(DepositPrefix)+8])
	depositor := key[len(DepositPrefix)+8:]
	return proposalID, depositor
}

// Emergency proposal key functions

// EmergencyProposalKey returns the store key for emergency proposals
func EmergencyProposalKey(proposalID uint64) []byte {
	key := make([]byte, len(EmergencyProposalPrefix)+8)
	copy(key, EmergencyProposalPrefix)
	binary.BigEndian.PutUint64(key[len(EmergencyProposalPrefix):], proposalID)
	return key
}

// EmergencyTypeIndexKey returns the index key for emergency proposals by type
func EmergencyTypeIndexKey(emergencyType uint32, proposalID uint64) []byte {
	key := make([]byte, len(EmergencyTypeIndexPrefix)+4+8)
	copy(key, EmergencyTypeIndexPrefix)
	binary.BigEndian.PutUint32(key[len(EmergencyTypeIndexPrefix):], emergencyType)
	binary.BigEndian.PutUint64(key[len(EmergencyTypeIndexPrefix)+4:], proposalID)
	return key
}

// EmergencySeverityKey returns the index key for emergency proposals by severity
func EmergencySeverityKey(severity int, proposalID uint64) []byte {
	key := make([]byte, len(EmergencySeverityPrefix)+4+8)
	copy(key, EmergencySeverityPrefix)
	binary.BigEndian.PutUint32(key[len(EmergencySeverityPrefix):], uint32(severity))
	binary.BigEndian.PutUint64(key[len(EmergencySeverityPrefix)+4:], proposalID)
	return key
}

// Delegation key functions

// DelegationKey returns the store key for vote delegations
func DelegationKey(delegator string, companyID uint64, proposalType uint32) []byte {
	delegatorBytes := []byte(delegator)
	key := make([]byte, len(DelegationPrefix)+len(delegatorBytes)+1+8+4)
	copy(key, DelegationPrefix)
	copy(key[len(DelegationPrefix):], delegatorBytes)
	key[len(DelegationPrefix)+len(delegatorBytes)] = 0x00 // separator
	binary.BigEndian.PutUint64(key[len(DelegationPrefix)+len(delegatorBytes)+1:], companyID)
	binary.BigEndian.PutUint32(key[len(DelegationPrefix)+len(delegatorBytes)+1+8:], proposalType)
	return key
}

// DelegateIndexKey returns the index key for delegations by delegate
func DelegateIndexKey(delegate string, proposalType uint32, delegator string) []byte {
	delegateBytes := []byte(delegate)
	delegatorBytes := []byte(delegator)
	key := make([]byte, len(DelegateIndexPrefix)+len(delegateBytes)+1+4+len(delegatorBytes)+1)
	copy(key, DelegateIndexPrefix)
	copy(key[len(DelegateIndexPrefix):], delegateBytes)
	key[len(DelegateIndexPrefix)+len(delegateBytes)] = 0x00 // separator
	binary.BigEndian.PutUint32(key[len(DelegateIndexPrefix)+len(delegateBytes)+1:], proposalType)
	copy(key[len(DelegateIndexPrefix)+len(delegateBytes)+1+4:], delegatorBytes)
	key[len(DelegateIndexPrefix)+len(delegateBytes)+1+4+len(delegatorBytes)] = 0x00 // separator
	return key
}

// DelegatorIndexKey returns the index key for delegations by delegator
func DelegatorIndexKey(delegator string, proposalType uint32) []byte {
	delegatorBytes := []byte(delegator)
	key := make([]byte, len(DelegatorIndexPrefix)+len(delegatorBytes)+1+4)
	copy(key, DelegatorIndexPrefix)
	copy(key[len(DelegatorIndexPrefix):], delegatorBytes)
	key[len(DelegatorIndexPrefix)+len(delegatorBytes)] = 0x00 // separator
	binary.BigEndian.PutUint32(key[len(DelegatorIndexPrefix)+len(delegatorBytes)+1:], proposalType)
	return key
}

// CompanyDelegationKey returns the index key for company-specific delegations
func CompanyDelegationKey(companyID uint64, proposalType uint32, delegator string) []byte {
	delegatorBytes := []byte(delegator)
	key := make([]byte, len(CompanyDelegationPrefix)+8+4+len(delegatorBytes)+1)
	copy(key, CompanyDelegationPrefix)
	binary.BigEndian.PutUint64(key[len(CompanyDelegationPrefix):], companyID)
	binary.BigEndian.PutUint32(key[len(CompanyDelegationPrefix)+8:], proposalType)
	copy(key[len(CompanyDelegationPrefix)+8+4:], delegatorBytes)
	key[len(CompanyDelegationPrefix)+8+4+len(delegatorBytes)] = 0x00 // separator
	return key
}

// Statistics key functions

// GovernanceStatsKey returns the store key for governance statistics
func GovernanceStatsKey() []byte {
	return GovernanceStatsPrefix
}

// ParticipationStatsKey returns the store key for participation statistics
func ParticipationStatsKey(participant []byte) []byte {
	key := make([]byte, len(ParticipationStatsPrefix)+len(participant))
	copy(key, ParticipationStatsPrefix)
	copy(key[len(ParticipationStatsPrefix):], participant)
	return key
}

// CompanyProposalIndexKey returns the index key for company proposals
func CompanyProposalIndexKey(companyID uint64) []byte {
	key := make([]byte, len(CompanyProposalPrefix)+8)
	copy(key, CompanyProposalPrefix)
	binary.BigEndian.PutUint64(key[len(CompanyProposalPrefix):], companyID)
	return key
}

// CompanyProposalTypeIndexKey returns the index key for company proposals by type
func CompanyProposalTypeIndexKey(proposalType CompanyProposalType) []byte {
	typeBytes := []byte(proposalType.String())
	key := make([]byte, len(CompanyProposalPrefix)+len(typeBytes)+1)
	copy(key, CompanyProposalPrefix)
	copy(key[len(CompanyProposalPrefix):], typeBytes)
	key[len(CompanyProposalPrefix)+len(typeBytes)] = 0x00
	return key
}

// PrefixEndBytes returns the end bytes for a prefix iterator
func PrefixEndBytes(prefix []byte) []byte {
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		if end[i] < 0xff {
			end[i]++
			return end[:i+1]
		}
	}
	return nil // All 0xff bytes, no end
}

// DepositEndTimeKey returns the key for deposit end time queue
// This allows efficient querying of proposals with expired deposit periods
func DepositEndTimeKey(endTime int64, proposalID uint64) []byte {
	key := make([]byte, len(DepositEndTimePrefix)+8+8)
	copy(key, DepositEndTimePrefix)
	binary.BigEndian.PutUint64(key[len(DepositEndTimePrefix):], uint64(endTime))
	binary.BigEndian.PutUint64(key[len(DepositEndTimePrefix)+8:], proposalID)
	return key
}

// VotingEndTimeKey returns the key for voting end time queue
// This allows efficient querying of proposals with ended voting periods
func VotingEndTimeKey(endTime int64, proposalID uint64) []byte {
	key := make([]byte, len(VotingEndTimePrefix)+8+8)
	copy(key, VotingEndTimePrefix)
	binary.BigEndian.PutUint64(key[len(VotingEndTimePrefix):], uint64(endTime))
	binary.BigEndian.PutUint64(key[len(VotingEndTimePrefix)+8:], proposalID)
	return key
}

// DepositEndTimeRangePrefix returns the prefix for iterating proposals expiring before a time
func DepositEndTimeRangePrefix(maxTime int64) []byte {
	key := make([]byte, len(DepositEndTimePrefix)+8)
	copy(key, DepositEndTimePrefix)
	binary.BigEndian.PutUint64(key[len(DepositEndTimePrefix):], uint64(maxTime))
	return key
}

// VotingEndTimeRangePrefix returns the prefix for iterating proposals ending before a time
func VotingEndTimeRangePrefix(maxTime int64) []byte {
	key := make([]byte, len(VotingEndTimePrefix)+8)
	copy(key, VotingEndTimePrefix)
	binary.BigEndian.PutUint64(key[len(VotingEndTimePrefix):], uint64(maxTime))
	return key
}

// ==========================================================================
// ANTI-SPAM KEY FUNCTIONS
// ==========================================================================

// ProposerDailyCountKey returns the key for tracking daily proposal count per address
// dayTimestamp should be the Unix timestamp at the start of the day (midnight UTC)
func ProposerDailyCountKey(proposer []byte, dayTimestamp int64) []byte {
	key := make([]byte, len(ProposerDailyCountPrefix)+len(proposer)+8)
	copy(key, ProposerDailyCountPrefix)
	copy(key[len(ProposerDailyCountPrefix):], proposer)
	binary.BigEndian.PutUint64(key[len(ProposerDailyCountPrefix)+len(proposer):], uint64(dayTimestamp))
	return key
}

// ProposerLastSubmitKey returns the key for tracking when an address last submitted a proposal
func ProposerLastSubmitKey(proposer []byte) []byte {
	key := make([]byte, len(ProposerLastSubmitPrefix)+len(proposer))
	copy(key, ProposerLastSubmitPrefix)
	copy(key[len(ProposerLastSubmitPrefix):], proposer)
	return key
}

// BurnedFeesKey returns the key for tracking total burned fees
func BurnedFeesKey() []byte {
	return BurnedFeesPrefix
}

// GetDayTimestamp returns the Unix timestamp for the start of the day (midnight UTC)
func GetDayTimestamp(timestamp int64) int64 {
	// 86400 seconds per day
	return (timestamp / 86400) * 86400
}