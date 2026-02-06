package types

import (
	"cosmossdk.io/errors"
)

// x/escrow module sentinel errors
var (
	// Escrow errors
	ErrEscrowNotFound        = errors.Register(ModuleName, 1, "escrow not found")
	ErrEscrowAlreadyExists   = errors.Register(ModuleName, 2, "escrow already exists")
	ErrInvalidEscrow         = errors.Register(ModuleName, 3, "invalid escrow")
	ErrEscrowNotFunded       = errors.Register(ModuleName, 4, "escrow not funded")
	ErrEscrowAlreadyFunded   = errors.Register(ModuleName, 5, "escrow already funded")
	ErrEscrowExpired         = errors.Register(ModuleName, 6, "escrow has expired")
	ErrEscrowNotExpired      = errors.Register(ModuleName, 7, "escrow has not expired")
	ErrEscrowCompleted       = errors.Register(ModuleName, 8, "escrow already completed")
	ErrEscrowDisputed        = errors.Register(ModuleName, 9, "escrow is in dispute")
	ErrInvalidStatus         = errors.Register(ModuleName, 10, "invalid escrow status")

	// Dispute errors
	ErrDisputeNotFound       = errors.Register(ModuleName, 20, "dispute not found")
	ErrDisputeAlreadyExists  = errors.Register(ModuleName, 21, "dispute already exists for escrow")
	ErrInvalidDispute        = errors.Register(ModuleName, 22, "invalid dispute")
	ErrDisputeResolved       = errors.Register(ModuleName, 23, "dispute already resolved")
	ErrDisputeDeadlinePassed = errors.Register(ModuleName, 24, "dispute deadline has passed")
	ErrCannotAppeal          = errors.Register(ModuleName, 25, "cannot appeal dispute")
	ErrMaxAppealsReached     = errors.Register(ModuleName, 26, "maximum appeals reached")
	ErrInvalidResolution     = errors.Register(ModuleName, 27, "invalid dispute resolution")

	// Moderator errors
	ErrModeratorNotFound     = errors.Register(ModuleName, 30, "moderator not found")
	ErrModeratorAlreadyExists = errors.Register(ModuleName, 31, "moderator already registered")
	ErrModeratorNotActive    = errors.Register(ModuleName, 32, "moderator is not active")
	ErrInsufficientStake     = errors.Register(ModuleName, 33, "insufficient moderator stake")
	ErrModeratorNotAssigned  = errors.Register(ModuleName, 34, "moderator not assigned to escrow")
	ErrAlreadyVoted          = errors.Register(ModuleName, 35, "moderator already voted")
	ErrInvalidModeratorTier  = errors.Register(ModuleName, 36, "moderator tier too low for escrow value")

	// Asset errors
	ErrInsufficientFunds     = errors.Register(ModuleName, 40, "insufficient funds")
	ErrInvalidAsset          = errors.Register(ModuleName, 41, "invalid asset")
	ErrAssetTransferFailed   = errors.Register(ModuleName, 42, "asset transfer failed")

	// Authorization errors
	ErrUnauthorized          = errors.Register(ModuleName, 50, "unauthorized")
	ErrNotParticipant        = errors.Register(ModuleName, 51, "not a participant in escrow")
	ErrNotInitiator          = errors.Register(ModuleName, 52, "not the dispute initiator")

	// Condition errors
	ErrConditionNotMet       = errors.Register(ModuleName, 60, "condition not met")
	ErrInvalidCondition      = errors.Register(ModuleName, 61, "invalid condition")

	// Evidence errors
	ErrInvalidEvidence          = errors.Register(ModuleName, 70, "invalid evidence")
	ErrEvidenceAlreadySubmitted = errors.Register(ModuleName, 71, "evidence already submitted")

	// Stake-as-Trust-Ceiling errors
	ErrDisputeExceedsTrustCeiling  = errors.Register(ModuleName, 80, "dispute value exceeds moderator stake (trust ceiling)")
	ErrModeratorBlacklisted        = errors.Register(ModuleName, 81, "moderator is blacklisted")
	ErrModeratorHasActiveDisputes  = errors.Register(ModuleName, 82, "moderator has active disputes, cannot unstake")
	ErrUnbondingInProgress         = errors.Register(ModuleName, 83, "unbonding already in progress")
	ErrUnbondingNotFound           = errors.Register(ModuleName, 84, "unbonding request not found")
	ErrUnbondingNotComplete        = errors.Register(ModuleName, 85, "unbonding period not complete")
	ErrUnauthorizedValidatorAction = errors.Register(ModuleName, 86, "only validators can perform this action")
	ErrValidatorNotActive          = errors.Register(ModuleName, 87, "validator is not active")
	ErrCannotSlashBeyondStake      = errors.Register(ModuleName, 88, "cannot slash more than staked amount")
	ErrAlreadyBlacklisted          = errors.Register(ModuleName, 89, "moderator already blacklisted")
	ErrInvalidSlashFraction        = errors.Register(ModuleName, 90, "invalid slash fraction (must be 0-1)")
	ErrValidatorTierTooLow         = errors.Register(ModuleName, 91, "validator tier too low for this action")
	ErrModeratorNotBlacklisted     = errors.Register(ModuleName, 92, "moderator is not blacklisted")
	ErrBanNotExpired               = errors.Register(ModuleName, 93, "temporary ban has not expired")
	ErrInvalidWithdrawal           = errors.Register(ModuleName, 94, "invalid withdrawal request")

	// Universal staking tier requirement errors
	ErrInsufficientTierToModerate    = errors.Register(ModuleName, 100, "must be Warden tier or higher (100K HODL) to be a moderator")
	ErrInsufficientTierForLargeDispute = errors.Register(ModuleName, 101, "must be Steward tier or higher (1M HODL) to moderate large disputes")
	ErrStakingKeeperNotSet           = errors.Register(ModuleName, 102, "staking keeper not configured")

	// Phase 1: User Reporting System errors
	ErrReportNotFound                = errors.Register(ModuleName, 110, "report not found")
	ErrReportAlreadyExists           = errors.Register(ModuleName, 111, "report already exists")
	ErrInvalidReport                 = errors.Register(ModuleName, 112, "invalid report")
	ErrReporterTierTooLow            = errors.Register(ModuleName, 113, "must be Keeper tier or higher (10K HODL) to submit reports")
	ErrReportAlreadyResolved         = errors.Register(ModuleName, 114, "report already resolved")
	ErrNotAssignedReviewer           = errors.Register(ModuleName, 115, "not assigned as reviewer for this report")
	ErrReviewerAlreadyVoted          = errors.Register(ModuleName, 116, "reviewer already voted on this report")
	ErrInvalidReportTarget           = errors.Register(ModuleName, 117, "invalid report target")
	ErrReportDeadlinePassed          = errors.Register(ModuleName, 118, "report investigation deadline has passed")
	ErrInsufficientReviewerTier      = errors.Register(ModuleName, 119, "reviewer tier too low for this report")

	// Appeal errors (extended)
	ErrAppealNotFound                = errors.Register(ModuleName, 120, "appeal not found")
	ErrAppealAlreadyExists           = errors.Register(ModuleName, 121, "appeal already exists")
	ErrInvalidAppeal                 = errors.Register(ModuleName, 122, "invalid appeal")
	ErrAppealLevelMaxed              = errors.Register(ModuleName, 123, "maximum appeal level reached")
	ErrAppealDeadlinePassed          = errors.Register(ModuleName, 124, "appeal deadline has passed")
	ErrAppealerNotParticipant        = errors.Register(ModuleName, 125, "only dispute participants can file appeals")
	ErrAppealAlreadyResolved         = errors.Register(ModuleName, 126, "appeal already resolved")
	ErrNotAppealReviewer             = errors.Register(ModuleName, 127, "not assigned as appeal reviewer")
	ErrAppealReviewerAlreadyVoted    = errors.Register(ModuleName, 128, "reviewer already voted on this appeal")

	// Moderator metrics errors
	ErrModeratorMetricsNotFound      = errors.Register(ModuleName, 130, "moderator metrics not found")
	ErrAutoBlacklistTriggered        = errors.Register(ModuleName, 131, "moderator auto-blacklisted due to poor performance")

	// Escrow reserve errors
	ErrInsufficientReserve           = errors.Register(ModuleName, 140, "insufficient escrow reserve funds")
	ErrClaimNotFound                 = errors.Register(ModuleName, 141, "reserve claim not found")
	ErrClaimAlreadyProcessed         = errors.Register(ModuleName, 142, "reserve claim already processed")

	// Wrong resolution recovery errors
	ErrNotDisputeParticipant         = errors.Register(ModuleName, 150, "reporter must be a participant in the disputed escrow")
	ErrEscrowNotDisputed             = errors.Register(ModuleName, 151, "escrow was not disputed")
	ErrFundsAlreadyRecovered         = errors.Register(ModuleName, 152, "funds have already been recovered for this dispute")
	ErrRecoveryFailed                = errors.Register(ModuleName, 153, "failed to recover funds")

	// Voluntary return errors
	ErrNotCounterparty               = errors.Register(ModuleName, 160, "only the counterparty can respond to this report")
	ErrVoluntaryReturnExpired        = errors.Register(ModuleName, 161, "voluntary return grace period has expired")
	ErrReportNotPendingReturn        = errors.Register(ModuleName, 162, "report is not pending voluntary return")
	ErrInsufficientFundsForReturn    = errors.Register(ModuleName, 163, "insufficient funds to complete voluntary return")
	ErrVoluntaryReturnAlreadyDone    = errors.Register(ModuleName, 164, "voluntary return already completed")
	ErrAlreadyRejected               = errors.Register(ModuleName, 165, "counterparty already rejected this claim")

	// Reporter ban/spam errors
	ErrReporterBanned                = errors.Register(ModuleName, 170, "reporter is banned from submitting reports")
	ErrReporterRateLimited           = errors.Register(ModuleName, 171, "reporter has exceeded rate limit")
	ErrReporterCooldown              = errors.Register(ModuleName, 172, "reporter is in cooldown period after dismissed reports")
	ErrStakeAgeTooLow                = errors.Register(ModuleName, 173, "stake age too low - must be staked for at least 7 days to submit reports")

	// Company investigation edge cases
	ErrActiveInvestigationExists     = errors.Register(ModuleName, 1140, "company already has active investigation")
	ErrReviewerConflictOfInterest    = errors.Register(ModuleName, 1141, "reviewer has shareholding in company under investigation")
	ErrInsufficientReviewers         = errors.Register(ModuleName, 1142, "insufficient reviewers available for investigation")
	ErrReviewerRecentlyStaked        = errors.Register(ModuleName, 1143, "reviewer must be staked for at least 3 days to vote")
	ErrEvidenceLockedAfterVoting     = errors.Register(ModuleName, 1144, "cannot add evidence after voting has started")
	ErrRetaliatoryReportNotAllowed   = errors.Register(ModuleName, 1145, "cannot report user who has active report against you")
	ErrReportCooldownActive          = errors.Register(ModuleName, 1146, "must wait 7 days after being reported before filing user reports")
)

// Event types
const (
	EventTypeEscrowCreated       = "escrow_created"
	EventTypeEscrowFunded        = "escrow_funded"
	EventTypeEscrowReleased      = "escrow_released"
	EventTypeEscrowRefunded      = "escrow_refunded"
	EventTypeEscrowCancelled     = "escrow_cancelled"
	EventTypeEscrowExpired       = "escrow_expired"
	EventTypeDisputeOpened       = "dispute_opened"
	EventTypeDisputeResolved     = "dispute_resolved"
	EventTypeDisputeAppealed     = "dispute_appealed"
	EventTypeModeratorVoted      = "moderator_voted"
	EventTypeModeratorRegistered = "moderator_registered"
	EventTypeEvidenceSubmitted   = "evidence_submitted"

	// Stake-as-Trust-Ceiling event types
	EventTypeModeratorUnstakeRequested = "moderator_unstake_requested"
	EventTypeModeratorUnstakeCompleted = "moderator_unstake_completed"
	EventTypeModeratorUnstakeCancelled = "moderator_unstake_cancelled"
	EventTypeModeratorSlashed          = "moderator_slashed"
	EventTypeModeratorBlacklisted      = "moderator_blacklisted"
	EventTypeModeratorUnblacklisted    = "moderator_unblacklisted"
	EventTypeModeratorStakeIncreased   = "moderator_stake_increased"
	EventTypeValidatorAction           = "validator_action"

	// Phase 1: User Reporting System event types
	EventTypeReportSubmitted           = "report_submitted"
	EventTypeReportEvidenceAdded       = "report_evidence_added"
	EventTypeReportReviewersAssigned   = "report_reviewers_assigned"
	EventTypeReportVoted               = "report_voted"
	EventTypeReportResolved            = "report_resolved"
	EventTypeReportAppealed            = "report_appealed"
	EventTypeAppealSubmitted           = "appeal_submitted"
	EventTypeAppealVoted               = "appeal_voted"
	EventTypeAppealResolved            = "appeal_resolved"
	EventTypeAppealEscalated           = "appeal_escalated"
	EventTypeModeratorWarning          = "moderator_warning"
	EventTypeModeratorAutoBlacklisted  = "moderator_auto_blacklisted"
	EventTypeReserveContribution       = "reserve_contribution"
	EventTypeReserveClaim              = "reserve_claim"

	// Wrong resolution recovery event types
	EventTypeFundsRecovered            = "funds_recovered"
	EventTypeWrongResolutionConfirmed  = "wrong_resolution_confirmed"
	EventTypeModeratorPenalized        = "moderator_penalized"

	// Voluntary return event types
	EventTypeVoluntaryReturnRequested  = "voluntary_return_requested"
	EventTypeVoluntaryReturnCompleted  = "voluntary_return_completed"
	EventTypeVoluntaryReturnExpired    = "voluntary_return_expired"
	EventTypeVoluntaryReturnRejected   = "voluntary_return_rejected"
	EventTypeFraudulentReportPenalty   = "fraudulent_report_penalty"

	// Report escalation event types
	EventTypeReportEscalated           = "report_escalated"
	EventTypeReportDeadlineExtended    = "report_deadline_extended"
	EventTypeReportStale               = "report_stale"
	EventTypeReportEscalatedToGovernance = "report_escalated_to_governance"
)

// Attribute keys
const (
	AttributeKeyEscrowID   = "escrow_id"
	AttributeKeyDisputeID  = "dispute_id"
	AttributeKeySender     = "sender"
	AttributeKeyRecipient  = "recipient"
	AttributeKeyModerator  = "moderator"
	AttributeKeyStatus     = "status"
	AttributeKeyResolution = "resolution"
	AttributeKeyAmount     = "amount"
	AttributeKeyAsset      = "asset"

	// Stake-as-Trust-Ceiling attribute keys
	AttributeKeyValidator        = "validator"
	AttributeKeySlashAmount      = "slash_amount"
	AttributeKeySlashFraction    = "slash_fraction"
	AttributeKeyReason           = "reason"
	AttributeKeyPermanent        = "permanent"
	AttributeKeyUnbondingAmount  = "unbonding_amount"
	AttributeKeyCompletesAt      = "completes_at"
	AttributeKeyActionID         = "action_id"
	AttributeKeyAction           = "action"
	AttributeKeyTarget           = "target"
	AttributeKeyTrustCeiling     = "trust_ceiling"
	AttributeKeyAvailableStake   = "available_stake"
	AttributeKeyActiveDisputes   = "active_disputes"

	// Phase 1: User Reporting System attribute keys
	AttributeKeyReportID         = "report_id"
	AttributeKeyReporter         = "reporter"
	AttributeKeyReportType       = "report_type"
	AttributeKeyTargetType       = "target_type"
	AttributeKeyTargetID         = "target_id"
	AttributeKeyPriority         = "priority"
	AttributeKeySeverity         = "severity"
	AttributeKeyReviewer         = "reviewer"
	AttributeKeyConfirmed        = "confirmed"
	AttributeKeyAppealID         = "appeal_id"
	AttributeKeyAppealLevel      = "appeal_level"
	AttributeKeyAppellant        = "appellant"
	AttributeKeyUpheld           = "upheld"
	AttributeKeyOverturn         = "overturn"
	AttributeKeyWarningCount     = "warning_count"
	AttributeKeyOverturnRate     = "overturn_rate"
	AttributeKeyRewardAmount     = "reward_amount"
	AttributeKeyPenaltyAmount    = "penalty_amount"
	AttributeKeyClaimID          = "claim_id"
	AttributeKeyClaimant         = "claimant"

	// Wrong resolution recovery attribute keys
	AttributeKeyRecoveredAmount  = "recovered_amount"
	AttributeKeyRecoverySource   = "recovery_source"
	AttributeKeyWrongedParty     = "wronged_party"
	AttributeKeyOriginalResolution = "original_resolution"
	AttributeKeyCorrectResolution  = "correct_resolution"

	// Voluntary return attribute keys
	AttributeKeyCounterparty         = "counterparty"
	AttributeKeyReturnAmount         = "return_amount"
	AttributeKeyGracePeriodDeadline  = "grace_period_deadline"
	AttributeKeyVoluntaryReturn      = "voluntary_return"
)
