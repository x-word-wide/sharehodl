package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "escrow"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_escrow"
)

// Store keys
var (
	// EscrowPrefix stores escrow data
	EscrowPrefix = []byte{0x01}

	// DisputePrefix stores dispute data
	DisputePrefix = []byte{0x02}

	// ModeratorPrefix stores moderator data
	ModeratorPrefix = []byte{0x03}

	// EscrowCounterKey stores the global escrow counter
	EscrowCounterKey = []byte{0x04}

	// DisputeCounterKey stores the global dispute counter
	DisputeCounterKey = []byte{0x05}

	// UserEscrowsPrefix stores user escrow indices
	UserEscrowsPrefix = []byte{0x06}

	// ModeratorEscrowsPrefix stores moderator escrow indices
	ModeratorEscrowsPrefix = []byte{0x07}

	// ParamsKey stores module parameters
	ParamsKey = []byte{0x08}

	// Stake-as-Trust-Ceiling: New store keys

	// UnbondingModeratorPrefix stores moderators in unbonding period
	UnbondingModeratorPrefix = []byte{0x09}

	// ModeratorBlacklistPrefix stores blacklisted moderators
	ModeratorBlacklistPrefix = []byte{0x0A}

	// ModeratorDisputeIndexPrefix indexes active disputes by moderator
	ModeratorDisputeIndexPrefix = []byte{0x0B}

	// ValidatorActionLogPrefix stores validator oversight action logs
	ValidatorActionLogPrefix = []byte{0x0C}

	// ValidatorActionCounterKey stores the global action counter
	ValidatorActionCounterKey = []byte{0x0D}

	// Phase 1: User Reporting System - New store keys

	// ReportPrefix stores report data
	ReportPrefix = []byte{0x0E}

	// ReportCounterKey stores the global report counter
	ReportCounterKey = []byte{0x0F}

	// ReportByTargetPrefix indexes reports by target entity
	ReportByTargetPrefix = []byte{0x10}

	// ReportByReporterPrefix indexes reports by reporter
	ReportByReporterPrefix = []byte{0x11}

	// AppealPrefix stores appeal data
	AppealPrefix = []byte{0x12}

	// AppealCounterKey stores the global appeal counter
	AppealCounterKey = []byte{0x13}

	// ModeratorMetricsPrefix stores moderator accountability metrics
	ModeratorMetricsPrefix = []byte{0x14}

	// EscrowReserveKey stores the escrow reserve fund
	EscrowReserveKey = []byte{0x15}

	// AppealByDisputePrefix indexes appeals by dispute ID
	AppealByDisputePrefix = []byte{0x16}

	// AppealByReportPrefix indexes appeals by report ID
	AppealByReportPrefix = []byte{0x17}

	// ReportByStatusPrefix indexes reports by status
	ReportByStatusPrefix = []byte{0x18}

	// ReporterHistoryPrefix stores reporter abuse/spam tracking
	ReporterHistoryPrefix = []byte{0x19}

	// ReporterBanPrefix indexes banned reporters
	ReporterBanPrefix = []byte{0x1A}

	// CompanyInvestigationPrefix stores company fraud investigations
	CompanyInvestigationPrefix = []byte{0x1B}

	// CompanyInvestigationCounterKey stores the global investigation counter
	CompanyInvestigationCounterKey = []byte{0x1C}

	// CompanyInvestigationByCompanyPrefix indexes investigations by company ID
	CompanyInvestigationByCompanyPrefix = []byte{0x1D}

	// CompanyInvestigationByStatusPrefix indexes investigations by status
	CompanyInvestigationByStatusPrefix = []byte{0x1E}
)

// GetEscrowKey returns the store key for an escrow
func GetEscrowKey(escrowID uint64) []byte {
	return append(EscrowPrefix, sdk.Uint64ToBigEndian(escrowID)...)
}

// GetDisputeKey returns the store key for a dispute
func GetDisputeKey(disputeID uint64) []byte {
	return append(DisputePrefix, sdk.Uint64ToBigEndian(disputeID)...)
}

// GetModeratorKey returns the store key for a moderator
func GetModeratorKey(address string) []byte {
	return append(ModeratorPrefix, []byte(address)...)
}

// GetUserEscrowsKey returns the store key for user's escrows
func GetUserEscrowsKey(user string) []byte {
	return append(UserEscrowsPrefix, []byte(user)...)
}

// GetModeratorEscrowsKey returns the store key for moderator's escrows
func GetModeratorEscrowsKey(moderator string) []byte {
	return append(ModeratorEscrowsPrefix, []byte(moderator)...)
}

// GetEscrowDisputeKey returns the store key for escrow's dispute
func GetEscrowDisputeKey(escrowID uint64) []byte {
	key := append(DisputePrefix, []byte("escrow:")...)
	return append(key, sdk.Uint64ToBigEndian(escrowID)...)
}

// Stake-as-Trust-Ceiling: Key helper functions

// GetUnbondingModeratorKey returns the store key for an unbonding moderator
func GetUnbondingModeratorKey(address string) []byte {
	return append(UnbondingModeratorPrefix, []byte(address)...)
}

// GetModeratorBlacklistKey returns the store key for a blacklisted moderator
func GetModeratorBlacklistKey(address string) []byte {
	return append(ModeratorBlacklistPrefix, []byte(address)...)
}

// GetModeratorDisputeIndexKey returns the store key for moderator's dispute index
func GetModeratorDisputeIndexKey(moderator string, disputeID uint64) []byte {
	key := append(ModeratorDisputeIndexPrefix, []byte(moderator)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(disputeID)...)
}

// GetModeratorDisputeIndexPrefixKey returns the prefix for all disputes of a moderator
func GetModeratorDisputeIndexPrefixKey(moderator string) []byte {
	key := append(ModeratorDisputeIndexPrefix, []byte(moderator)...)
	return append(key, []byte(":")...)
}

// GetValidatorActionLogKey returns the store key for a validator action log entry
func GetValidatorActionLogKey(actionID uint64) []byte {
	return append(ValidatorActionLogPrefix, sdk.Uint64ToBigEndian(actionID)...)
}

// GetValidatorActionLogByValidatorKey returns the store key for actions by a specific validator
func GetValidatorActionLogByValidatorKey(validator string, actionID uint64) []byte {
	key := append(ValidatorActionLogPrefix, []byte("validator:")...)
	key = append(key, []byte(validator)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(actionID)...)
}

// Phase 1: User Reporting System - Key helper functions

// GetReportKey returns the store key for a report
func GetReportKey(reportID uint64) []byte {
	return append(ReportPrefix, sdk.Uint64ToBigEndian(reportID)...)
}

// GetReportByTargetKey returns the store key for target's reports index
func GetReportByTargetKey(targetType, targetID string, reportID uint64) []byte {
	key := append(ReportByTargetPrefix, []byte(targetType)...)
	key = append(key, []byte(":")...)
	key = append(key, []byte(targetID)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(reportID)...)
}

// GetReportByTargetPrefixKey returns the prefix for all reports against a target
func GetReportByTargetPrefixKey(targetType, targetID string) []byte {
	key := append(ReportByTargetPrefix, []byte(targetType)...)
	key = append(key, []byte(":")...)
	key = append(key, []byte(targetID)...)
	return append(key, []byte(":")...)
}

// GetReportByReporterKey returns the store key for reporter's reports index
func GetReportByReporterKey(reporter string, reportID uint64) []byte {
	key := append(ReportByReporterPrefix, []byte(reporter)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(reportID)...)
}

// GetReportByReporterPrefixKey returns the prefix for all reports by a reporter
func GetReportByReporterPrefixKey(reporter string) []byte {
	key := append(ReportByReporterPrefix, []byte(reporter)...)
	return append(key, []byte(":")...)
}

// GetReportByStatusKey returns the store key for status index
func GetReportByStatusKey(status string, reportID uint64) []byte {
	key := append(ReportByStatusPrefix, []byte(status)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(reportID)...)
}

// GetReportByStatusPrefixKey returns the prefix for all reports with a status
func GetReportByStatusPrefixKey(status string) []byte {
	key := append(ReportByStatusPrefix, []byte(status)...)
	return append(key, []byte(":")...)
}

// GetAppealKey returns the store key for an appeal
func GetAppealKey(appealID uint64) []byte {
	return append(AppealPrefix, sdk.Uint64ToBigEndian(appealID)...)
}

// GetAppealByDisputeKey returns the store key for dispute's appeals
func GetAppealByDisputeKey(disputeID uint64, appealID uint64) []byte {
	key := append(AppealByDisputePrefix, sdk.Uint64ToBigEndian(disputeID)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(appealID)...)
}

// GetAppealByDisputePrefixKey returns the prefix for all appeals on a dispute
func GetAppealByDisputePrefixKey(disputeID uint64) []byte {
	key := append(AppealByDisputePrefix, sdk.Uint64ToBigEndian(disputeID)...)
	return append(key, []byte(":")...)
}

// GetAppealByReportKey returns the store key for report's appeals
func GetAppealByReportKey(reportID uint64, appealID uint64) []byte {
	key := append(AppealByReportPrefix, sdk.Uint64ToBigEndian(reportID)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(appealID)...)
}

// GetAppealByReportPrefixKey returns the prefix for all appeals on a report
func GetAppealByReportPrefixKey(reportID uint64) []byte {
	key := append(AppealByReportPrefix, sdk.Uint64ToBigEndian(reportID)...)
	return append(key, []byte(":")...)
}

// GetModeratorMetricsKey returns the store key for moderator metrics
func GetModeratorMetricsKey(address string) []byte {
	return append(ModeratorMetricsPrefix, []byte(address)...)
}

// GetReporterHistoryKey returns the store key for reporter history
func GetReporterHistoryKey(address string) []byte {
	return append(ReporterHistoryPrefix, []byte(address)...)
}

// GetReporterBanKey returns the store key for banned reporter
func GetReporterBanKey(address string) []byte {
	return append(ReporterBanPrefix, []byte(address)...)
}

// GetCompanyInvestigationKey returns the store key for a company investigation
func GetCompanyInvestigationKey(investigationID uint64) []byte {
	return append(CompanyInvestigationPrefix, sdk.Uint64ToBigEndian(investigationID)...)
}

// GetCompanyInvestigationByCompanyKey returns the store key for company's investigation index
func GetCompanyInvestigationByCompanyKey(companyID uint64, investigationID uint64) []byte {
	key := append(CompanyInvestigationByCompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(investigationID)...)
}

// GetCompanyInvestigationByCompanyPrefixKey returns the prefix for all investigations of a company
func GetCompanyInvestigationByCompanyPrefixKey(companyID uint64) []byte {
	key := append(CompanyInvestigationByCompanyPrefix, sdk.Uint64ToBigEndian(companyID)...)
	return append(key, []byte(":")...)
}

// GetCompanyInvestigationByStatusKey returns the store key for status index
func GetCompanyInvestigationByStatusKey(status string, investigationID uint64) []byte {
	key := append(CompanyInvestigationByStatusPrefix, []byte(status)...)
	key = append(key, []byte(":")...)
	return append(key, sdk.Uint64ToBigEndian(investigationID)...)
}

// GetCompanyInvestigationByStatusPrefixKey returns the prefix for all investigations with a status
func GetCompanyInvestigationByStatusPrefixKey(status string) []byte {
	key := append(CompanyInvestigationByStatusPrefix, []byte(status)...)
	return append(key, []byte(":")...)
}
