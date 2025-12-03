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