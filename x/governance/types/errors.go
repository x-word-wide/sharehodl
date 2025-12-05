package types

import (
	"cosmossdk.io/errors"
)

// Governance module error codes
const (
	DefaultCodespace = "governance"
)

// Governance module errors
var (
	// General errors
	ErrUnauthorized = errors.Register(DefaultCodespace, 1, "unauthorized")
	ErrInvalidAddress = errors.Register(DefaultCodespace, 2, "invalid address")

	// Proposal errors
	ErrProposalNotFound = errors.Register(DefaultCodespace, 100, "proposal not found")
	ErrInvalidProposal = errors.Register(DefaultCodespace, 101, "invalid proposal")
	ErrProposalNotActive = errors.Register(DefaultCodespace, 102, "proposal not in active status")
	ErrProposalClosed = errors.Register(DefaultCodespace, 103, "proposal voting period has ended")
	ErrProposalPassed = errors.Register(DefaultCodespace, 104, "proposal has already passed")
	ErrProposalRejected = errors.Register(DefaultCodespace, 105, "proposal has been rejected")
	ErrProposalCancelled = errors.Register(DefaultCodespace, 106, "proposal has been cancelled")
	ErrInvalidProposalType = errors.Register(DefaultCodespace, 107, "invalid proposal type")
	ErrInvalidProposalStatus = errors.Register(DefaultCodespace, 108, "invalid proposal status")
	ErrInvalidProposalContent = errors.Register(DefaultCodespace, 109, "invalid proposal content")
	ErrProposalTitleTooLong = errors.Register(DefaultCodespace, 110, "proposal title too long")
	ErrProposalDescriptionTooLong = errors.Register(DefaultCodespace, 111, "proposal description too long")
	ErrProposalAlreadyExists = errors.Register(DefaultCodespace, 112, "proposal already exists")
	ErrInvalidProposalID = errors.Register(DefaultCodespace, 113, "invalid proposal ID")

	// Voting errors
	ErrInvalidVote = errors.Register(DefaultCodespace, 200, "invalid vote")
	ErrVoteNotFound = errors.Register(DefaultCodespace, 201, "vote not found")
	ErrAlreadyVoted = errors.Register(DefaultCodespace, 202, "address has already voted on proposal")
	ErrInvalidVoteOption = errors.Register(DefaultCodespace, 203, "invalid vote option")
	ErrInvalidVoteWeight = errors.Register(DefaultCodespace, 204, "invalid vote weight")
	ErrVotingPeriodEnded = errors.Register(DefaultCodespace, 205, "voting period has ended")
	ErrVotingPeriodNotStarted = errors.Register(DefaultCodespace, 206, "voting period has not started")
	ErrInsufficientVotingPower = errors.Register(DefaultCodespace, 207, "insufficient voting power")
	ErrVetoVote = errors.Register(DefaultCodespace, 208, "veto vote cast")
	ErrInvalidWeightedVote = errors.Register(DefaultCodespace, 209, "invalid weighted vote options")

	// Deposit errors
	ErrInvalidDeposit = errors.Register(DefaultCodespace, 300, "invalid deposit")
	ErrDepositNotFound = errors.Register(DefaultCodespace, 301, "deposit not found")
	ErrInsufficientDeposit = errors.Register(DefaultCodespace, 302, "insufficient deposit amount")
	ErrMaxDepositExceeded = errors.Register(DefaultCodespace, 303, "maximum deposit exceeded")
	ErrDepositPeriodEnded = errors.Register(DefaultCodespace, 304, "deposit period has ended")
	ErrInvalidDepositAmount = errors.Register(DefaultCodespace, 305, "invalid deposit amount")
	ErrInsufficientFunds = errors.Register(DefaultCodespace, 306, "insufficient funds for deposit")

	// Governance parameter errors
	ErrInvalidGovernanceParams = errors.Register(DefaultCodespace, 400, "invalid governance parameters")
	ErrInvalidQuorum = errors.Register(DefaultCodespace, 401, "invalid quorum")
	ErrInvalidThreshold = errors.Register(DefaultCodespace, 402, "invalid threshold")
	ErrInvalidVetoThreshold = errors.Register(DefaultCodespace, 403, "invalid veto threshold")
	ErrInvalidVotingPeriod = errors.Register(DefaultCodespace, 404, "invalid voting period")
	ErrInvalidDepositPeriod = errors.Register(DefaultCodespace, 405, "invalid deposit period")
	ErrInvalidMinDeposit = errors.Register(DefaultCodespace, 406, "invalid minimum deposit")

	// Authority and permission errors
	ErrNotAuthorized = errors.Register(DefaultCodespace, 500, "not authorized to perform this action")
	ErrInvalidAuthority = errors.Register(DefaultCodespace, 501, "invalid authority address")
	ErrInsufficientPermissions = errors.Register(DefaultCodespace, 502, "insufficient permissions")
	ErrValidatorNotFound = errors.Register(DefaultCodespace, 503, "validator not found")
	ErrCompanyNotFound = errors.Register(DefaultCodespace, 504, "company not found")
	ErrInvalidReason = errors.Register(DefaultCodespace, 505, "invalid reason provided")

	// Execution errors
	ErrProposalExecutionFailed = errors.Register(DefaultCodespace, 600, "proposal execution failed")
	ErrInvalidExecutionHeight = errors.Register(DefaultCodespace, 601, "invalid execution height")
	ErrProposalNotReadyForExecution = errors.Register(DefaultCodespace, 602, "proposal not ready for execution")
	ErrExecutionTimeout = errors.Register(DefaultCodespace, 603, "proposal execution timeout")

	// Tally and counting errors
	ErrInvalidTallyResult = errors.Register(DefaultCodespace, 700, "invalid tally result")
	ErrQuorumNotReached = errors.Register(DefaultCodespace, 701, "quorum not reached")
	ErrThresholdNotMet = errors.Register(DefaultCodespace, 702, "threshold not met")
	ErrVetoThresholdExceeded = errors.Register(DefaultCodespace, 703, "veto threshold exceeded")
	ErrInvalidVotingPower = errors.Register(DefaultCodespace, 704, "invalid voting power calculation")

	// Module integration errors
	ErrEquityModuleNotFound = errors.Register(DefaultCodespace, 800, "equity module not found")
	ErrValidatorModuleNotFound = errors.Register(DefaultCodespace, 801, "validator module not found")
	ErrHODLModuleNotFound = errors.Register(DefaultCodespace, 802, "HODL module not found")
	ErrInvalidModuleAction = errors.Register(DefaultCodespace, 803, "invalid module action")
	ErrModuleCallFailed = errors.Register(DefaultCodespace, 804, "module call failed")

	// Delegation errors
	ErrDelegationNotFound = errors.Register(DefaultCodespace, 850, "delegation not found")
	ErrInvalidDelegate = errors.Register(DefaultCodespace, 851, "invalid delegate address")
	ErrDelegationNotRevocable = errors.Register(DefaultCodespace, 852, "delegation is not revocable")
	ErrInvalidDelegation = errors.Register(DefaultCodespace, 853, "invalid delegation parameters")
	ErrDelegationExpired = errors.Register(DefaultCodespace, 854, "delegation has expired")
	ErrCircularDelegation = errors.Register(DefaultCodespace, 855, "circular delegation detected")
	
	// Emergency proposal errors
	ErrInvalidEmergencyType = errors.Register(DefaultCodespace, 860, "invalid emergency type")
	ErrInvalidSeverityLevel = errors.Register(DefaultCodespace, 861, "invalid severity level")
	ErrEmergencyThresholdNotMet = errors.Register(DefaultCodespace, 862, "emergency threshold not met")
	ErrEmergencyExecutionFailed = errors.Register(DefaultCodespace, 863, "emergency execution failed")
	ErrValidatorNotActive = errors.Register(DefaultCodespace, 864, "validator is not active")
	ErrEmergencyTimeLimit = errors.Register(DefaultCodespace, 865, "emergency execution time limit exceeded")
	
	// Company governance errors
	ErrNotShareholder = errors.Register(DefaultCodespace, 870, "address is not a shareholder")
	ErrInsufficientOwnership = errors.Register(DefaultCodespace, 871, "insufficient ownership percentage")
	ErrShareClassNotFound = errors.Register(DefaultCodespace, 872, "share class not found")
	ErrInvalidShareClass = errors.Register(DefaultCodespace, 873, "invalid share class")
	ErrBoardApprovalRequired = errors.Register(DefaultCodespace, 874, "board approval required")
	ErrCompanyProposalNotFound = errors.Register(DefaultCodespace, 875, "company proposal not found")
	ErrInvalidCompanyProposalType = errors.Register(DefaultCodespace, 876, "invalid company proposal type")
	
	// State and storage errors
	ErrInvalidState = errors.Register(DefaultCodespace, 900, "invalid module state")
	ErrStorageCorruption = errors.Register(DefaultCodespace, 901, "storage corruption detected")
	ErrInvalidKeyFormat = errors.Register(DefaultCodespace, 902, "invalid key format")
	ErrSerializationFailed = errors.Register(DefaultCodespace, 903, "serialization failed")
	ErrDeserializationFailed = errors.Register(DefaultCodespace, 904, "deserialization failed")
)