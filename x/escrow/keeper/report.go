package keeper

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/escrow/types"
)

// =============================================================================
// PHASE 1: USER REPORTING SYSTEM
// =============================================================================

// GetNextReportID returns the next report ID and increments the counter
func (k Keeper) GetNextReportID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ReportCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.ReportCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetNextAppealID returns the next appeal ID and increments the counter
func (k Keeper) GetNextAppealID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AppealCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.AppealCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// =============================================================================
// REPORT STORAGE
// =============================================================================

// GetReport returns a report by ID
func (k Keeper) GetReport(ctx sdk.Context, reportID uint64) (types.Report, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetReportKey(reportID))
	if bz == nil {
		return types.Report{}, false
	}

	var report types.Report
	if err := json.Unmarshal(bz, &report); err != nil {
		return types.Report{}, false
	}
	return report, true
}

// SetReport stores a report
func (k Keeper) SetReport(ctx sdk.Context, report types.Report) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(report)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal report", "error", err)
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	store.Set(types.GetReportKey(report.ID), bz)
	return nil
}

// DeleteReport removes a report
func (k Keeper) DeleteReport(ctx sdk.Context, reportID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetReportKey(reportID))
}

// IndexReport creates indexes for efficient querying
func (k Keeper) IndexReport(ctx sdk.Context, report types.Report) {
	store := ctx.KVStore(k.storeKey)

	// Index by target
	targetKey := types.GetReportByTargetKey(report.TargetType.String(), report.TargetID, report.ID)
	store.Set(targetKey, sdk.Uint64ToBigEndian(report.ID))

	// Index by reporter
	reporterKey := types.GetReportByReporterKey(report.Reporter, report.ID)
	store.Set(reporterKey, sdk.Uint64ToBigEndian(report.ID))

	// Index by status
	statusKey := types.GetReportByStatusKey(report.Status.String(), report.ID)
	store.Set(statusKey, sdk.Uint64ToBigEndian(report.ID))
}

// UpdateReportStatusIndex updates the status index when status changes
func (k Keeper) UpdateReportStatusIndex(ctx sdk.Context, report types.Report, oldStatus types.ReportStatus) {
	store := ctx.KVStore(k.storeKey)

	// Remove old status index
	oldStatusKey := types.GetReportByStatusKey(oldStatus.String(), report.ID)
	store.Delete(oldStatusKey)

	// Add new status index
	newStatusKey := types.GetReportByStatusKey(report.Status.String(), report.ID)
	store.Set(newStatusKey, sdk.Uint64ToBigEndian(report.ID))
}

// =============================================================================
// REPORT SUBMISSION
// =============================================================================

// SubmitReport creates a new fraud/misconduct report
func (k Keeper) SubmitReport(
	ctx sdk.Context,
	reporter string,
	reportType types.ReportType,
	targetType types.ReportTargetType,
	targetID string,
	reason string,
	severity int,
) (uint64, error) {
	// Validate reporter address
	reporterAddr, err := sdk.AccAddressFromBech32(reporter)
	if err != nil {
		return 0, types.ErrInvalidReport
	}

	// TIER CHECK: Must be Keeper+ (10K HODL) to submit reports
	if k.stakingKeeper == nil {
		return 0, types.ErrStakingKeeperNotSet
	}

	reporterTier := k.stakingKeeper.GetUserTierInt(ctx, reporterAddr)
	if reporterTier < types.StakeTierKeeper {
		return 0, types.ErrReporterTierTooLow
	}

	// ABUSE CHECK: Check if reporter is banned, rate-limited, or in cooldown
	if err := k.CheckReporterCanSubmit(ctx, reporter); err != nil {
		k.Logger(ctx).Info("report blocked",
			"reporter", reporter,
			"reason", err.Error())
		return 0, err
	}

	// NEW: Check for retaliatory/recursive reports (Issue #10)
	if targetType == types.ReportTargetTypeUser {
		targetAddr := targetID

		// Check if target has an active report against the reporter
		activeReports := k.GetActiveReportsByReporter(ctx, targetAddr)
		for _, activeReport := range activeReports {
			if activeReport.TargetType == types.ReportTargetTypeUser &&
				activeReport.TargetID == reporter {
				return 0, types.ErrRetaliatoryReportNotAllowed
			}
		}

		// Check cooldown - if reporter was recently a target, add cooldown
		recentlyTargeted := k.wasRecentlyTargeted(ctx, reporter, 7*24*time.Hour) // 7 day cooldown
		if recentlyTargeted {
			return 0, types.ErrReportCooldownActive
		}
	}

	// Validate target exists (basic validation)
	if err := k.validateReportTarget(ctx, targetType, targetID); err != nil {
		return 0, err
	}

	// Special validation for wrong resolution reports
	// Reporter must be a participant in the disputed escrow
	if reportType == types.ReportTypeWrongResolution {
		if targetType != types.ReportTargetTypeEscrow {
			return 0, types.ErrInvalidReportTarget
		}
		if err := k.validateWrongResolutionReport(ctx, reporter, targetID); err != nil {
			return 0, err
		}
	}

	// Create report
	reportID := k.GetNextReportID(ctx)
	report := types.NewReport(
		reportID,
		reportType,
		reporter,
		targetType,
		targetID,
		reason,
		severity,
		reporterTier,
		ctx.BlockTime(),
	)

	// Validate report
	if err := report.Validate(); err != nil {
		return 0, types.ErrInvalidReport
	}

	// Store report
	if err := k.SetReport(ctx, report); err != nil {
		return 0, err
	}

	// Create indexes
	k.IndexReport(ctx, report)

	// For WrongResolution reports, set up voluntary return grace period first
	// This gives the counterparty a chance to return funds without penalties
	if reportType == types.ReportTypeWrongResolution {
		if err := k.setupVoluntaryReturnPeriod(ctx, &report); err != nil {
			k.Logger(ctx).Error("failed to setup voluntary return period", "report_id", reportID, "error", err)
			// Fall through to normal flow if voluntary return setup fails
		} else {
			// Successfully set up voluntary return - don't assign reviewers yet
			k.SetReport(ctx, report)
			k.UpdateReportStatusIndex(ctx, report, types.ReportStatusOpen)

			// Emit voluntary return requested event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeVoluntaryReturnRequested,
					sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", reportID)),
					sdk.NewAttribute(types.AttributeKeyCounterparty, report.Counterparty),
					sdk.NewAttribute(types.AttributeKeyReturnAmount, report.AmountToReturn.String()),
					sdk.NewAttribute(types.AttributeKeyGracePeriodDeadline, report.VoluntaryReturnDeadline.String()),
				),
			)

			// Return early - don't assign reviewers until grace period expires
			return reportID, nil
		}
	}

	// For company fraud/scam reports, create tiered investigation (NO auto-freeze)
	// The freeze only happens after multi-tier voting approval + 24hr warning
	if (reportType == types.ReportTypeFraud || reportType == types.ReportTypeScam) && targetType == types.ReportTargetTypeCompany {
		var companyID uint64
		if _, err := fmt.Sscanf(targetID, "%d", &companyID); err == nil {
			// Create tiered investigation instead of immediate freeze
			investigationID, err := k.CreateCompanyInvestigation(ctx, companyID, reportID)
			if err != nil {
				k.Logger(ctx).Error("failed to create company investigation",
					"company_id", companyID,
					"report_id", reportID,
					"error", err)
			} else {
				k.Logger(ctx).Info("company investigation created - awaiting multi-tier review",
					"company_id", companyID,
					"report_id", reportID,
					"investigation_id", investigationID,
					"note", "NO freeze yet - requires Warden (2/3) + Steward (3/5) approval + 24hr warning")
			}
		}
	}

	// For non-WrongResolution reports (or if voluntary setup failed), assign reviewers
	if err := k.assignReviewers(ctx, &report); err != nil {
		k.Logger(ctx).Error("failed to assign reviewers", "report_id", reportID, "error", err)
		// Continue - report still created, reviewers can be assigned later
	} else {
		// Update report with assigned reviewers
		report.Status = types.ReportStatusUnderInvestigation
		k.SetReport(ctx, report)
		k.UpdateReportStatusIndex(ctx, report, types.ReportStatusOpen)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReportSubmitted,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", reportID)),
			sdk.NewAttribute(types.AttributeKeyReporter, reporter),
			sdk.NewAttribute(types.AttributeKeyReportType, reportType.String()),
			sdk.NewAttribute(types.AttributeKeyTargetType, targetType.String()),
			sdk.NewAttribute(types.AttributeKeyTargetID, targetID),
			sdk.NewAttribute(types.AttributeKeyPriority, fmt.Sprintf("%d", report.Priority)),
			sdk.NewAttribute(types.AttributeKeySeverity, fmt.Sprintf("%d", severity)),
		),
	)

	return reportID, nil
}

// validateReportTarget validates the target entity exists
func (k Keeper) validateReportTarget(ctx sdk.Context, targetType types.ReportTargetType, targetID string) error {
	switch targetType {
	case types.ReportTargetTypeEscrow:
		// Validate escrow exists
		// Parse targetID as uint64
		var escrowID uint64
		if _, err := fmt.Sscanf(targetID, "%d", &escrowID); err != nil {
			return types.ErrInvalidReportTarget
		}
		if _, found := k.GetEscrow(ctx, escrowID); !found {
			return types.ErrInvalidReportTarget
		}

	case types.ReportTargetTypeModerator:
		// Validate moderator exists
		if _, found := k.GetModerator(ctx, targetID); !found {
			return types.ErrInvalidReportTarget
		}

	case types.ReportTargetTypeUser:
		// Validate address format
		if _, err := sdk.AccAddressFromBech32(targetID); err != nil {
			return types.ErrInvalidReportTarget
		}

	case types.ReportTargetTypeCompany:
		// Company validation would be done via equity keeper if needed
		// For now, accept any non-empty company ID
		if targetID == "" {
			return types.ErrInvalidReportTarget
		}
	}

	return nil
}

// =============================================================================
// VOLUNTARY RETURN SYSTEM
// Gives counterparty (Bob) a chance to return funds without penalties
// =============================================================================

// setupVoluntaryReturnPeriod sets up the grace period for voluntary return
func (k Keeper) setupVoluntaryReturnPeriod(ctx sdk.Context, report *types.Report) error {
	// Parse escrow ID
	var escrowID uint64
	if _, err := fmt.Sscanf(report.TargetID, "%d", &escrowID); err != nil {
		return types.ErrInvalidReportTarget
	}

	// Get escrow and dispute
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound
	}

	dispute, found := k.GetEscrowDispute(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotDisputed
	}

	// Determine who received the funds wrongly (counterparty)
	counterparty := k.determineWrongfulRecipient(escrow, dispute, report.Reporter)
	if counterparty == "" {
		return fmt.Errorf("could not determine counterparty")
	}

	// Calculate amount to return
	amountToReturn := k.calculateRecoveryAmount(escrow, dispute, report.Reporter)
	if amountToReturn.IsZero() {
		return fmt.Errorf("no amount to return")
	}

	// Set up voluntary return period (48 hours)
	gracePeriod := 48 * 60 * 60 * 1000000000 // 48 hours in nanoseconds
	report.Status = types.ReportStatusPendingVoluntaryReturn
	report.Counterparty = counterparty
	report.AmountToReturn = amountToReturn
	report.VoluntaryReturnDeadline = ctx.BlockTime().Add(48 * 60 * 60 * 1000000000) // 48 hours
	report.VoluntaryReturnComplete = false

	k.Logger(ctx).Info("voluntary return period started",
		"report_id", report.ID,
		"counterparty", counterparty,
		"amount", amountToReturn.String(),
		"deadline", report.VoluntaryReturnDeadline.String())

	// Suppress unused variable warning
	_ = gracePeriod

	return nil
}

// VoluntaryReturn allows the counterparty (Bob) to voluntarily return funds
// This resolves the report without penalties for anyone
func (k Keeper) VoluntaryReturn(ctx sdk.Context, reportID uint64, returner string) error {
	report, found := k.GetReport(ctx, reportID)
	if !found {
		return types.ErrReportNotFound
	}

	// Check report is pending voluntary return
	if report.Status != types.ReportStatusPendingVoluntaryReturn {
		return types.ErrReportNotPendingReturn
	}

	// Check returner is the counterparty
	if returner != report.Counterparty {
		return types.ErrNotCounterparty
	}

	// Check grace period hasn't expired
	if ctx.BlockTime().After(report.VoluntaryReturnDeadline) {
		return types.ErrVoluntaryReturnExpired
	}

	// Check not already done
	if report.VoluntaryReturnComplete {
		return types.ErrVoluntaryReturnAlreadyDone
	}

	// Check counterparty has sufficient funds
	returnerAddr, err := sdk.AccAddressFromBech32(returner)
	if err != nil {
		return err
	}

	balance := k.bankKeeper.GetBalance(ctx, returnerAddr, "uhodl")
	if balance.Amount.LT(report.AmountToReturn) {
		return types.ErrInsufficientFundsForReturn
	}

	// Execute the voluntary return
	reporterAddr, err := sdk.AccAddressFromBech32(report.Reporter)
	if err != nil {
		return err
	}

	// Transfer funds from counterparty to reporter (victim)
	coins := sdk.NewCoins(sdk.NewCoin("uhodl", report.AmountToReturn))
	if err := k.bankKeeper.SendCoins(ctx, returnerAddr, reporterAddr, coins); err != nil {
		return fmt.Errorf("failed to transfer funds: %w", err)
	}

	// Update report
	oldStatus := report.Status
	report.Status = types.ReportStatusVoluntarilyResolved
	report.VoluntaryReturnComplete = true
	report.ResolvedAt = ctx.BlockTime()
	report.Resolution = fmt.Sprintf("Voluntarily resolved: %s returned %s to %s",
		report.Counterparty, report.AmountToReturn.String(), report.Reporter)
	report.ResolvedBy = returner
	report.ActionsTaken = append(report.ActionsTaken, "voluntary_return_completed")

	// No penalties for anyone - this was an honest resolution
	report.ReporterReputationChange = 0 // No reward for voluntary resolution
	report.ReporterRewardAmount = math.ZeroInt()

	k.SetReport(ctx, report)
	k.UpdateReportStatusIndex(ctx, report, oldStatus)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVoluntaryReturnCompleted,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", reportID)),
			sdk.NewAttribute(types.AttributeKeyCounterparty, returner),
			sdk.NewAttribute(types.AttributeKeyReporter, report.Reporter),
			sdk.NewAttribute(types.AttributeKeyReturnAmount, report.AmountToReturn.String()),
		),
	)

	k.Logger(ctx).Info("voluntary return completed",
		"report_id", reportID,
		"returner", returner,
		"recipient", report.Reporter,
		"amount", report.AmountToReturn.String())

	return nil
}

// RejectVoluntaryReturn allows the counterparty (Bob) to reject the claim
// This immediately escalates to investigation - Bob is saying "I did nothing wrong"
func (k Keeper) RejectVoluntaryReturn(ctx sdk.Context, reportID uint64, rejector string, reason string) error {
	report, found := k.GetReport(ctx, reportID)
	if !found {
		return types.ErrReportNotFound
	}

	// Check report is pending voluntary return
	if report.Status != types.ReportStatusPendingVoluntaryReturn {
		return types.ErrReportNotPendingReturn
	}

	// Check rejector is the counterparty
	if rejector != report.Counterparty {
		return types.ErrNotCounterparty
	}

	// Check not already rejected or returned
	if report.CounterpartyRejected {
		return types.ErrAlreadyRejected
	}
	if report.VoluntaryReturnComplete {
		return types.ErrVoluntaryReturnAlreadyDone
	}

	// Mark as rejected
	report.CounterpartyRejected = true
	report.CounterpartyResponse = reason

	// Immediately escalate to investigation
	oldStatus := report.Status

	if err := k.assignReviewers(ctx, &report); err != nil {
		k.Logger(ctx).Error("failed to assign reviewers after rejection",
			"report_id", report.ID,
			"error", err)
		return err
	}

	report.Status = types.ReportStatusUnderInvestigation
	report.ActionsTaken = append(report.ActionsTaken,
		fmt.Sprintf("counterparty_rejected: %s says '%s'", report.Counterparty, reason))

	k.SetReport(ctx, report)
	k.UpdateReportStatusIndex(ctx, report, oldStatus)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVoluntaryReturnRejected,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", reportID)),
			sdk.NewAttribute(types.AttributeKeyCounterparty, rejector),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)

	k.Logger(ctx).Info("voluntary return rejected, escalating to investigation",
		"report_id", reportID,
		"counterparty", rejector,
		"reason", reason)

	return nil
}

// ProcessExpiredVoluntaryReturns escalates reports where grace period expired
// Called from EndBlock
func (k Keeper) ProcessExpiredVoluntaryReturns(ctx sdk.Context) {
	reports := k.GetReportsByStatus(ctx, types.ReportStatusPendingVoluntaryReturn)

	for _, report := range reports {
		// Skip if already rejected (will be handled separately)
		if report.CounterpartyRejected {
			continue
		}

		if ctx.BlockTime().After(report.VoluntaryReturnDeadline) {
			// Grace period expired - counterparty didn't respond
			// Escalate to full investigation
			k.Logger(ctx).Info("voluntary return period expired, escalating to investigation",
				"report_id", report.ID,
				"counterparty", report.Counterparty)

			oldStatus := report.Status

			// Assign reviewers and escalate
			if err := k.assignReviewers(ctx, &report); err != nil {
				k.Logger(ctx).Error("failed to assign reviewers after grace period",
					"report_id", report.ID,
					"error", err)
				continue
			}

			report.Status = types.ReportStatusUnderInvestigation
			report.ActionsTaken = append(report.ActionsTaken,
				fmt.Sprintf("voluntary_return_expired: %s did not respond within 48 hours", report.Counterparty))

			k.SetReport(ctx, report)
			k.UpdateReportStatusIndex(ctx, report, oldStatus)

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeVoluntaryReturnExpired,
					sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", report.ID)),
					sdk.NewAttribute(types.AttributeKeyCounterparty, report.Counterparty),
				),
			)
		}
	}
}

// =============================================================================
// WRONG RESOLUTION VALIDATION
// =============================================================================

// validateWrongResolutionReport validates that the reporter is a participant in the disputed escrow
func (k Keeper) validateWrongResolutionReport(ctx sdk.Context, reporter string, targetID string) error {
	// Parse escrow ID
	var escrowID uint64
	if _, err := fmt.Sscanf(targetID, "%d", &escrowID); err != nil {
		return types.ErrInvalidReportTarget
	}

	// Get escrow
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return types.ErrEscrowNotFound
	}

	// Check reporter is a participant
	if reporter != escrow.Sender && reporter != escrow.Recipient {
		return types.ErrNotDisputeParticipant
	}

	// Check escrow was disputed and resolved
	if escrow.Status != types.EscrowStatusResolved {
		// Check if there's a dispute
		dispute, found := k.GetEscrowDispute(ctx, escrowID)
		if !found {
			return types.ErrEscrowNotDisputed
		}
		if dispute.Status != types.DisputeStatusResolved && dispute.Status != types.DisputeStatusFinal {
			return types.ErrEscrowNotDisputed
		}
	}

	return nil
}

// assignReviewers assigns validators to investigate the report
func (k Keeper) assignReviewers(ctx sdk.Context, report *types.Report) error {
	if k.stakingKeeper == nil {
		return types.ErrStakingKeeperNotSet
	}

	// Get required tier based on priority
	requiredTier := k.getRequiredReviewerTier(report.Priority)

	// Get eligible moderators (validators with sufficient tier)
	moderators := k.GetActiveModerators(ctx)

	// Filter by tier and shuffle for fairness
	var eligible []string
	for _, mod := range moderators {
		modAddr, err := sdk.AccAddressFromBech32(mod.Address)
		if err != nil {
			continue
		}

		tier := k.stakingKeeper.GetUserTierInt(ctx, modAddr)
		if tier >= requiredTier {
			eligible = append(eligible, mod.Address)
		}
	}

	if len(eligible) < report.VotesRequired {
		// Not enough eligible reviewers - keep report in open status
		return fmt.Errorf("insufficient eligible reviewers: need %d, have %d", report.VotesRequired, len(eligible))
	}

	// Select reviewers (simple selection - first N eligible)
	// In production, this would use a randomized selection for fairness
	numReviewers := report.VotesRequired
	if numReviewers > len(eligible) {
		numReviewers = len(eligible)
	}

	report.AssignedReviewers = eligible[:numReviewers]

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReportReviewersAssigned,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", report.ID)),
			sdk.NewAttribute("reviewer_count", fmt.Sprintf("%d", len(report.AssignedReviewers))),
		),
	)

	return nil
}

// getRequiredReviewerTier returns the minimum tier for reviewing a report
func (k Keeper) getRequiredReviewerTier(priority int) int {
	switch priority {
	case 5:
		return types.StakeTierArchon // Emergency: Archon+
	case 4:
		return types.StakeTierSteward // High: Steward+
	default:
		return types.StakeTierWarden // Standard: Warden+
	}
}

// =============================================================================
// REPORT EVIDENCE
// =============================================================================

// SubmitReportEvidence adds evidence to a report
func (k Keeper) SubmitReportEvidence(
	ctx sdk.Context,
	reportID uint64,
	submitter string,
	hash string,
	description string,
) error {
	report, found := k.GetReport(ctx, reportID)
	if !found {
		return types.ErrReportNotFound
	}

	// Only reporter or assigned reviewers can submit evidence
	if submitter != report.Reporter && !k.isAssignedReviewer(report, submitter) {
		return types.ErrUnauthorized
	}

	// Check report is still active
	if report.Status == types.ReportStatusConfirmed || report.Status == types.ReportStatusDismissed {
		return types.ErrReportAlreadyResolved
	}

	// NEW: Prevent evidence addition after voting has started (Issue #5)
	if len(report.ReviewVotes) > 0 {
		return types.ErrEvidenceLockedAfterVoting
	}

	// Add evidence
	evidence := types.Evidence{
		Submitter:   submitter,
		Hash:        hash,
		Description: description,
		SubmittedAt: ctx.BlockTime(),
	}
	report.Evidence = append(report.Evidence, evidence)

	// Update priority based on new evidence count
	report.Priority = types.CalculateReportPriority(
		k.getReporterTier(ctx, report.Reporter),
		report.Severity,
		len(report.Evidence),
	)

	k.SetReport(ctx, report)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReportEvidenceAdded,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", reportID)),
			sdk.NewAttribute("submitter", submitter),
			sdk.NewAttribute("evidence_hash", hash),
		),
	)

	return nil
}

func (k Keeper) isAssignedReviewer(report types.Report, address string) bool {
	for _, reviewer := range report.AssignedReviewers {
		if reviewer == address {
			return true
		}
	}
	return false
}

func (k Keeper) getReporterTier(ctx sdk.Context, reporter string) int {
	if k.stakingKeeper == nil {
		return 0
	}
	addr, err := sdk.AccAddressFromBech32(reporter)
	if err != nil {
		return 0
	}
	return k.stakingKeeper.GetUserTierInt(ctx, addr)
}

// =============================================================================
// REPORT VOTING
// =============================================================================

// VoteOnReport allows a reviewer to vote on a report
func (k Keeper) VoteOnReport(
	ctx sdk.Context,
	reportID uint64,
	reviewer string,
	confirmed bool,
	comments string,
) error {
	report, found := k.GetReport(ctx, reportID)
	if !found {
		return types.ErrReportNotFound
	}

	// Check reviewer is assigned
	if !k.isAssignedReviewer(report, reviewer) {
		return types.ErrNotAssignedReviewer
	}

	// Check hasn't already voted
	for _, vote := range report.ReviewVotes {
		if vote.Reviewer == reviewer {
			return types.ErrReviewerAlreadyVoted
		}
	}

	// Check report is under investigation
	if report.Status != types.ReportStatusUnderInvestigation {
		return types.ErrReportAlreadyResolved
	}

	// Check deadline
	if ctx.BlockTime().After(report.DeadlineAt) {
		return types.ErrReportDeadlinePassed
	}

	// Get reviewer tier
	reviewerAddr, _ := sdk.AccAddressFromBech32(reviewer)
	tier := k.stakingKeeper.GetUserTierInt(ctx, reviewerAddr)

	// Check tier meets requirements
	requiredTier := k.getRequiredReviewerTier(report.Priority)
	if tier < requiredTier {
		return types.ErrInsufficientReviewerTier
	}

	// NEW: Snapshot evidence state on first vote (Issue #5)
	if len(report.ReviewVotes) == 0 {
		report.EvidenceSnapshot = k.hashEvidenceState(report.Evidence)
		report.EvidenceLockedAt = ctx.BlockTime()
	}

	// Add vote
	vote := types.ReviewVote{
		Reviewer:  reviewer,
		Tier:      tier,
		Confirmed: confirmed,
		Comments:  comments,
		VotedAt:   ctx.BlockTime(),
	}
	report.ReviewVotes = append(report.ReviewVotes, vote)

	// Check if we have enough votes to resolve
	if len(report.ReviewVotes) >= report.VotesRequired {
		k.resolveReport(ctx, &report)
	}

	k.SetReport(ctx, report)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReportVoted,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", reportID)),
			sdk.NewAttribute(types.AttributeKeyReviewer, reviewer),
			sdk.NewAttribute(types.AttributeKeyConfirmed, fmt.Sprintf("%t", confirmed)),
		),
	)

	return nil
}

// resolveReport resolves a report based on votes
func (k Keeper) resolveReport(ctx sdk.Context, report *types.Report) {
	// Count votes
	confirmCount := 0
	dismissCount := 0

	for _, vote := range report.ReviewVotes {
		if vote.Confirmed {
			confirmCount++
		} else {
			dismissCount++
		}
	}

	oldStatus := report.Status

	// Simple majority decides
	if confirmCount > dismissCount {
		// =====================================================================
		// REPORT CONFIRMED - Reporter was telling the truth
		// =====================================================================
		report.Status = types.ReportStatusConfirmed
		report.Resolution = fmt.Sprintf("Report confirmed by %d of %d reviewers", confirmCount, len(report.ReviewVotes))

		// Execute actions based on report type (recover funds, penalize moderators, etc.)
		k.executeReportActions(ctx, report)

		// Reward reporter and record confirmed report
		report.ReporterReputationChange = 2000 // +2000 reputation
		k.rewardReporter(ctx, report)
		k.RecordConfirmedReport(ctx, report.Reporter) // Track good behavior

	} else {
		// =====================================================================
		// REPORT DISMISSED - Reporter was WRONG (possibly fraudulent)
		// Apply ESCALATING penalties based on history
		// =====================================================================
		report.Status = types.ReportStatusDismissed
		report.Resolution = fmt.Sprintf("Report dismissed by %d of %d reviewers - claim was invalid", dismissCount, len(report.ReviewVotes))

		// Apply escalating penalties based on reporter's history
		// This handles:
		// - Base penalties (reputation loss)
		// - Escalated penalties for repeat offenders
		// - Stake slashing for serious/repeated abuse
		// - Temporary or permanent bans for spammers
		k.ApplyReporterPenalty(ctx, report)

		// Also apply the reputation penalty through staking keeper
		k.penalizeReporter(ctx, report)

		// Get reporter history to include in event
		history := k.GetOrCreateReporterHistory(ctx, report.Reporter)

		// Emit penalty event with escalation info
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeFraudulentReportPenalty,
				sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", report.ID)),
				sdk.NewAttribute(types.AttributeKeyReporter, report.Reporter),
				sdk.NewAttribute(types.AttributeKeyPenaltyAmount, fmt.Sprintf("%d", report.ReporterReputationChange)),
				sdk.NewAttribute("consecutive_dismissed", fmt.Sprintf("%d", history.ConsecutiveDismissed)),
				sdk.NewAttribute("total_dismissed", fmt.Sprintf("%d", history.DismissedReports)),
				sdk.NewAttribute("is_banned", fmt.Sprintf("%t", history.IsBanned)),
			),
		)

		// If this was a company fraud/scam report that was dismissed, restore the company
		if (report.ReportType == types.ReportTypeFraud || report.ReportType == types.ReportTypeScam) &&
			report.TargetType == types.ReportTargetTypeCompany {
			var companyID uint64
			if _, err := fmt.Sscanf(report.TargetID, "%d", &companyID); err == nil {
				if k.equityKeeper != nil {
					if err := k.equityKeeper.ConfirmFraudAndDelist(ctx, companyID, report.ID, false, "escrow_report_system"); err != nil {
						k.Logger(ctx).Error("failed to restore company after dismissed report",
							"company_id", companyID,
							"report_id", report.ID,
							"error", err)
					} else {
						report.ActionsTaken = append(report.ActionsTaken, fmt.Sprintf("company_%d_restored", companyID))
						k.Logger(ctx).Info("company restored after report dismissed",
							"company_id", companyID,
							"report_id", report.ID)
					}
				}
			}
		}
	}

	report.ResolvedAt = ctx.BlockTime()
	k.UpdateReportStatusIndex(ctx, *report, oldStatus)

	// Emit resolution event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReportResolved,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", report.ID)),
			sdk.NewAttribute(types.AttributeKeyStatus, report.Status.String()),
			sdk.NewAttribute(types.AttributeKeyConfirmed, fmt.Sprintf("%t", report.Status == types.ReportStatusConfirmed)),
		),
	)
}

// slashFraudulentReporter slashes the stake of someone who filed a fraudulent WrongResolution report
// This is more severe than a regular false report because they tried to steal funds
func (k Keeper) slashFraudulentReporter(ctx sdk.Context, report *types.Report) {
	if k.stakingKeeper == nil {
		return
	}

	reporterAddr, err := sdk.AccAddressFromBech32(report.Reporter)
	if err != nil {
		return
	}

	// Get reporter's tier - this determines if they have stake to slash
	tier := k.stakingKeeper.GetUserTierInt(ctx, reporterAddr)
	if tier < types.StakeTierKeeper {
		// No stake to slash
		return
	}

	// Slash 5% of their stake for fraudulent WrongResolution claim
	// This is done through the staking module
	// For now, we just record the action - actual slashing would go through staking keeper
	report.ActionsTaken = append(report.ActionsTaken,
		fmt.Sprintf("stake_slash_pending: 5%% of reporter stake for fraudulent claim"))

	k.Logger(ctx).Warn("fraudulent WrongResolution report detected",
		"reporter", report.Reporter,
		"report_id", report.ID,
		"counterparty", report.Counterparty,
		"attempted_amount", report.AmountToReturn.String())
}

// executeReportActions executes actions for a confirmed report
func (k Keeper) executeReportActions(ctx sdk.Context, report *types.Report) {
	switch report.ReportType {
	case types.ReportTypeFraud, types.ReportTypeScam:
		// If targeting a company, execute delisting process
		if report.TargetType == types.ReportTargetTypeCompany {
			// Parse company ID
			var companyID uint64
			if _, err := fmt.Sscanf(report.TargetID, "%d", &companyID); err == nil {
				// Trigger fraud delisting through equity module
				if k.equityKeeper != nil {
					if err := k.equityKeeper.ConfirmFraudAndDelist(ctx, companyID, report.ID, true, "escrow_report_system"); err != nil {
						k.Logger(ctx).Error("failed to delist fraudulent company",
							"company_id", companyID,
							"report_id", report.ID,
							"error", err)
						report.ActionsTaken = append(report.ActionsTaken, fmt.Sprintf("delisting_failed: %s", err.Error()))
					} else {
						report.ActionsTaken = append(report.ActionsTaken, fmt.Sprintf("company_%d_delisted", companyID))
						k.Logger(ctx).Info("fraudulent company delisted",
							"company_id", companyID,
							"report_id", report.ID,
							"reason", report.ReportType.String())
					}
				} else {
					report.ActionsTaken = append(report.ActionsTaken, "company_investigation_initiated_equity_keeper_missing")
				}
			} else {
				report.ActionsTaken = append(report.ActionsTaken, "company_investigation_initiated_invalid_id")
			}
		}

	case types.ReportTypeModeratorMisconduct:
		// Update moderator metrics
		if report.TargetType == types.ReportTargetTypeModerator {
			k.recordModeratorReportConfirmed(ctx, report.TargetID)
			report.ActionsTaken = append(report.ActionsTaken, "moderator_metrics_updated")
		}

	case types.ReportTypeCollusion:
		// Blacklist moderator if confirmed
		if report.TargetType == types.ReportTargetTypeModerator {
			k.recordModeratorReportConfirmed(ctx, report.TargetID)
			report.ActionsTaken = append(report.ActionsTaken, "collusion_recorded")
		}

	case types.ReportTypeMarketManipulation:
		// Would trigger DEX investigation
		report.ActionsTaken = append(report.ActionsTaken, "dex_investigation_flagged")

	case types.ReportTypeWrongResolution:
		// This is the key case - user reported a wrong dispute resolution
		// We need to recover funds and penalize moderators
		if report.TargetType == types.ReportTargetTypeEscrow {
			recoveredAmount, bonusRecovered, err := k.executeWrongResolutionRecovery(ctx, report)
			if err != nil {
				k.Logger(ctx).Error("failed to execute wrong resolution recovery",
					"report_id", report.ID,
					"escrow_id", report.TargetID,
					"error", err)
				report.ActionsTaken = append(report.ActionsTaken, fmt.Sprintf("recovery_failed: %s", err.Error()))
			} else {
				report.ActionsTaken = append(report.ActionsTaken, fmt.Sprintf("funds_recovered: %s", recoveredAmount.String()))

				// Reporter reward comes from EXCESS slashing, NOT from victim's funds
				// If we slashed more than needed for recovery, reporter gets up to 10% of that excess
				if bonusRecovered.IsPositive() {
					report.ReporterRewardAmount = bonusRecovered.QuoRaw(10) // 10% of excess

					// Cap reward at 1000 HODL (1000000000 uhodl)
					maxReward := math.NewInt(1000000000)
					if report.ReporterRewardAmount.GT(maxReward) {
						report.ReporterRewardAmount = maxReward
					}

					// Pay the reward to the reporter (if different from victim)
					if report.ReporterRewardAmount.IsPositive() {
						k.payReporterReward(ctx, report)
					}
				}
			}
		}
	}
}

// payReporterReward pays the HODL reward to the reporter
func (k Keeper) payReporterReward(ctx sdk.Context, report *types.Report) {
	reporterAddr, err := sdk.AccAddressFromBech32(report.Reporter)
	if err != nil {
		return
	}

	// The reward comes from the escrow module (excess from slashing)
	rewardCoins := sdk.NewCoins(sdk.NewCoin("uhodl", report.ReporterRewardAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, reporterAddr, rewardCoins); err != nil {
		k.Logger(ctx).Error("failed to pay reporter reward",
			"reporter", report.Reporter,
			"amount", report.ReporterRewardAmount.String(),
			"error", err)
		return
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"reporter_reward_paid",
			sdk.NewAttribute(types.AttributeKeyReporter, report.Reporter),
			sdk.NewAttribute(types.AttributeKeyRewardAmount, report.ReporterRewardAmount.String()),
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", report.ID)),
		),
	)
}

// executeWrongResolutionRecovery handles fund recovery when a wrong resolution is confirmed
// Returns: (amountRecoveredForVictim, bonusForRewards, error)
func (k Keeper) executeWrongResolutionRecovery(ctx sdk.Context, report *types.Report) (math.Int, math.Int, error) {
	// Parse escrow ID
	var escrowID uint64
	if _, err := fmt.Sscanf(report.TargetID, "%d", &escrowID); err != nil {
		return math.ZeroInt(), math.ZeroInt(), types.ErrInvalidReportTarget
	}

	// Get escrow and dispute
	escrow, found := k.GetEscrow(ctx, escrowID)
	if !found {
		return math.ZeroInt(), math.ZeroInt(), types.ErrEscrowNotFound
	}

	dispute, found := k.GetEscrowDispute(ctx, escrowID)
	if !found {
		return math.ZeroInt(), math.ZeroInt(), types.ErrEscrowNotDisputed
	}

	// Determine the wronged party (the reporter)
	wrongedParty := report.Reporter
	wrongedPartyAddr, err := sdk.AccAddressFromBech32(wrongedParty)
	if err != nil {
		return math.ZeroInt(), math.ZeroInt(), err
	}

	// Determine how much they should have received
	// Based on the dispute resolution, calculate what they lost
	recoveryAmount := k.calculateRecoveryAmount(escrow, dispute, wrongedParty)
	if recoveryAmount.IsZero() {
		return math.ZeroInt(), math.ZeroInt(), fmt.Errorf("no funds to recover")
	}

	// Try to recover funds from multiple sources (waterfall)
	// We try to recover MORE than needed so there's a bonus for rewards
	targetRecovery := recoveryAmount.MulRaw(110).QuoRaw(100) // Try to get 110% (10% extra for rewards)
	actualRecovered, sources := k.recoverFundsWaterfall(ctx, wrongedPartyAddr, recoveryAmount, targetRecovery, dispute)

	if actualRecovered.IsZero() {
		return math.ZeroInt(), math.ZeroInt(), types.ErrRecoveryFailed
	}

	// Calculate how much goes to victim vs bonus
	victimRecovery := math.MinInt(actualRecovered, recoveryAmount)
	bonusRecovery := math.ZeroInt()
	if actualRecovered.GT(recoveryAmount) {
		bonusRecovery = actualRecovered.Sub(recoveryAmount)
	}

	// If recovery is incomplete, emit event so victim knows to file governance proposal
	if victimRecovery.LT(recoveryAmount) {
		shortfall := recoveryAmount.Sub(victimRecovery)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"recovery_shortfall",
				sdk.NewAttribute(types.AttributeKeyEscrowID, fmt.Sprintf("%d", escrowID)),
				sdk.NewAttribute(types.AttributeKeyWrongedParty, wrongedParty),
				sdk.NewAttribute("shortfall_amount", shortfall.String()),
				sdk.NewAttribute("recovered_amount", victimRecovery.String()),
				sdk.NewAttribute("expected_amount", recoveryAmount.String()),
				sdk.NewAttribute("remedy", "File ProposalTypeCommunityPoolSpend via governance to recover remaining funds"),
			),
		)

		k.Logger(ctx).Warn("incomplete fund recovery - victim should file governance proposal",
			"escrow_id", escrowID,
			"victim", wrongedParty,
			"shortfall", shortfall.String(),
			"recovered", victimRecovery.String())
	}

	// Penalize the moderators who voted for the wrong resolution
	k.penalizeModeratorsForWrongResolution(ctx, dispute)

	// Mark dispute as overturned
	dispute.Status = types.DisputeStatusFinal
	k.SetDispute(ctx, dispute)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFundsRecovered,
			sdk.NewAttribute(types.AttributeKeyEscrowID, fmt.Sprintf("%d", escrowID)),
			sdk.NewAttribute(types.AttributeKeyDisputeID, fmt.Sprintf("%d", dispute.ID)),
			sdk.NewAttribute(types.AttributeKeyWrongedParty, wrongedParty),
			sdk.NewAttribute(types.AttributeKeyRecoveredAmount, victimRecovery.String()),
			sdk.NewAttribute("bonus_recovered", bonusRecovery.String()),
			sdk.NewAttribute(types.AttributeKeyRecoverySource, sources),
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", report.ID)),
		),
	)

	return victimRecovery, bonusRecovery, nil
}

// calculateRecoveryAmount calculates how much the wronged party should recover
func (k Keeper) calculateRecoveryAmount(escrow types.Escrow, dispute types.Dispute, wrongedParty string) math.Int {
	// The wronged party is the one who didn't receive funds but should have
	// Based on the original resolution:
	// - If ReleaseBuyer was chosen but reporter is sender, sender lost their funds
	// - If ReleaseSeller/Refund was chosen but reporter is recipient, recipient lost their share

	totalValue := escrow.TotalValue.TruncateInt()

	switch dispute.Resolution {
	case types.DisputeResolutionReleaseBuyer:
		// Funds went to recipient (buyer). If reporter is sender, they lost everything
		if wrongedParty == escrow.Sender {
			return totalValue
		}
	case types.DisputeResolutionReleaseSeller, types.DisputeResolutionRefund:
		// Funds went to sender. If reporter is recipient, they lost everything
		if wrongedParty == escrow.Recipient {
			return totalValue
		}
	case types.DisputeResolutionSplit:
		// Funds were split. Calculate what they should have gotten vs what they got
		// For simplicity, assume they got less than they should have
		// In a real case, we'd need more detailed tracking
		return totalValue.QuoRaw(2) // Assume they lost half
	}

	return math.ZeroInt()
}

// recoverFundsWaterfall attempts to recover funds from multiple sources
// victimNeed: minimum amount needed for the victim
// targetAmount: ideal amount including bonus for reporter rewards
// Returns actual recovered amount and sources used
//
// RECOVERY ORDER (Waterfall):
// 1. Clawback from wrongful recipient (Bob) - they got funds they shouldn't have
// 2. Escrow Reserve Fund - only if clawback insufficient
// 3. Moderator stake slashing - covers remainder + bonus for rewards
func (k Keeper) recoverFundsWaterfall(
	ctx sdk.Context,
	victim sdk.AccAddress,
	victimNeed math.Int,
	targetAmount math.Int,
	dispute types.Dispute,
) (math.Int, string) {
	totalRecovered := math.ZeroInt()
	sources := []string{}

	// Get escrow to find the wrongful recipient
	escrow, found := k.GetEscrow(ctx, dispute.EscrowID)
	if !found {
		return math.ZeroInt(), "escrow_not_found"
	}

	// Determine who wrongfully received the funds
	wrongfulRecipient := k.determineWrongfulRecipient(escrow, dispute, victim.String())

	// =========================================================================
	// SOURCE 1: CLAWBACK FROM WRONGFUL RECIPIENT (Bob)
	// This is the PRIMARY source - Bob got money he shouldn't have
	// =========================================================================
	if wrongfulRecipient != "" {
		clawbackAmount, clawbackSuccess := k.clawbackFromWrongfulRecipient(ctx, wrongfulRecipient, victimNeed)
		if clawbackSuccess && clawbackAmount.IsPositive() {
			// Transfer clawed back funds to victim
			coins := sdk.NewCoins(sdk.NewCoin("uhodl", clawbackAmount))
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, victim, coins); err == nil {
				totalRecovered = totalRecovered.Add(clawbackAmount)
				sources = append(sources, fmt.Sprintf("clawback:%s", clawbackAmount.String()))
			}
		}
	}

	// =========================================================================
	// SOURCE 2: ESCROW RESERVE FUND
	// Only used if clawback didn't cover everything (Bob spent the money)
	// =========================================================================
	if totalRecovered.LT(victimNeed) {
		reserve := k.GetEscrowReserve(ctx)
		if reserve.TotalFunds.IsPositive() {
			reserveNeeded := math.MinInt(victimNeed.Sub(totalRecovered), reserve.TotalFunds)
			if reserveNeeded.IsPositive() {
				coins := sdk.NewCoins(sdk.NewCoin("uhodl", reserveNeeded))
				if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, victim, coins); err == nil {
					reserve.TotalFunds = reserve.TotalFunds.Sub(reserveNeeded)
					reserve.TotalPaidOut = reserve.TotalPaidOut.Add(reserveNeeded)
					k.SetEscrowReserve(ctx, reserve)

					totalRecovered = totalRecovered.Add(reserveNeeded)
					sources = append(sources, fmt.Sprintf("reserve:%s", reserveNeeded.String()))
				}
			}
		}
	}

	// =========================================================================
	// SOURCE 3: MODERATOR STAKE SLASHING
	// Covers any remaining victim need + bonus for reporter rewards
	// Moderators are liable for their bad decision
	// =========================================================================
	remaining := targetAmount.Sub(totalRecovered)
	if remaining.IsPositive() {
		slashRecovered := k.slashModeratorsForRecovery(ctx, dispute, remaining)
		if slashRecovered.IsPositive() {
			// First, pay remaining victim need
			victimRemaining := victimNeed.Sub(totalRecovered)
			if victimRemaining.IsPositive() {
				toVictim := math.MinInt(slashRecovered, victimRemaining)
				coins := sdk.NewCoins(sdk.NewCoin("uhodl", toVictim))
				// SlashModerator burns coins, so we mint equivalent for recovery
				if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err == nil {
					if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, victim, coins); err == nil {
						totalRecovered = totalRecovered.Add(toVictim)
						slashRecovered = slashRecovered.Sub(toVictim)
					}
				}
			}

			// Any remaining goes to module for reporter reward
			if slashRecovered.IsPositive() {
				coins := sdk.NewCoins(sdk.NewCoin("uhodl", slashRecovered))
				if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err == nil {
					totalRecovered = totalRecovered.Add(slashRecovered)
				}
			}

			sources = append(sources, fmt.Sprintf("moderator_slash:%s", slashRecovered.String()))
		}
	}

	return totalRecovered, fmt.Sprintf("[%s]", joinSources(sources))
}

// determineWrongfulRecipient figures out who received funds they shouldn't have
func (k Keeper) determineWrongfulRecipient(escrow types.Escrow, dispute types.Dispute, victim string) string {
	switch dispute.Resolution {
	case types.DisputeResolutionReleaseBuyer:
		// Recipient (buyer) got the funds
		// If victim is sender, then recipient wrongfully received
		if victim == escrow.Sender {
			return escrow.Recipient
		}
	case types.DisputeResolutionReleaseSeller, types.DisputeResolutionRefund:
		// Sender got the funds
		// If victim is recipient, then sender wrongfully received
		if victim == escrow.Recipient {
			return escrow.Sender
		}
	case types.DisputeResolutionSplit:
		// Both got some - more complex, return empty for now
		// In a real implementation, we'd track exact amounts
		return ""
	}
	return ""
}

// clawbackFromWrongfulRecipient attempts to recover funds from someone who received them wrongly
// Returns (amount recovered, success)
func (k Keeper) clawbackFromWrongfulRecipient(ctx sdk.Context, wrongfulRecipient string, amount math.Int) (math.Int, bool) {
	recipientAddr, err := sdk.AccAddressFromBech32(wrongfulRecipient)
	if err != nil {
		return math.ZeroInt(), false
	}

	// Check how much they have
	balance := k.bankKeeper.GetBalance(ctx, recipientAddr, "uhodl")

	// Try to clawback up to what they have or what we need
	clawbackAmount := math.MinInt(amount, balance.Amount)

	if clawbackAmount.IsZero() || clawbackAmount.IsNegative() {
		// They have no funds - they spent it all
		k.Logger(ctx).Warn("wrongful recipient has insufficient funds for clawback",
			"recipient", wrongfulRecipient,
			"needed", amount.String(),
			"available", balance.Amount.String())
		return math.ZeroInt(), false
	}

	// Execute clawback - transfer from wrongful recipient to escrow module
	coins := sdk.NewCoins(sdk.NewCoin("uhodl", clawbackAmount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, recipientAddr, types.ModuleName, coins); err != nil {
		k.Logger(ctx).Error("failed to clawback from wrongful recipient",
			"recipient", wrongfulRecipient,
			"amount", clawbackAmount.String(),
			"error", err)
		return math.ZeroInt(), false
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"funds_clawback",
			sdk.NewAttribute("wrongful_recipient", wrongfulRecipient),
			sdk.NewAttribute(types.AttributeKeyAmount, clawbackAmount.String()),
		),
	)

	k.Logger(ctx).Info("clawed back funds from wrongful recipient",
		"recipient", wrongfulRecipient,
		"amount", clawbackAmount.String())

	return clawbackAmount, true
}

func joinSources(sources []string) string {
	result := ""
	for i, s := range sources {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

// slashModeratorsForRecovery slashes moderators who voted for the wrong resolution
// Returns the total amount recovered from slashing
//
// IMPORTANT: Moderators are FULLY LIABLE up to their stake for wrong decisions.
// This is why we have stake-as-trust-ceiling - they should never moderate
// transactions larger than what they can cover if they're wrong.
func (k Keeper) slashModeratorsForRecovery(ctx sdk.Context, dispute types.Dispute, neededAmount math.Int) math.Int {
	totalSlashed := math.ZeroInt()
	remaining := neededAmount

	// Count how many moderators voted wrong
	wrongVoterCount := 0
	for _, vote := range dispute.Votes {
		if vote.Vote == dispute.Resolution {
			wrongVoterCount++
		}
	}

	if wrongVoterCount == 0 {
		return math.ZeroInt()
	}

	// Calculate per-moderator share of liability
	// Each wrong-voting moderator is equally liable
	perModeratorLiability := neededAmount.QuoRaw(int64(wrongVoterCount))

	// Slash each moderator who voted for the wrong resolution
	for _, vote := range dispute.Votes {
		if vote.Vote == dispute.Resolution {
			// This moderator voted for the resolution that was overturned
			mod, found := k.GetModerator(ctx, vote.Moderator)
			if !found {
				continue
			}

			// Calculate how much to slash from this moderator
			// They are liable for their share, up to their full stake
			slashAmount := math.MinInt(perModeratorLiability, mod.Stake)

			// But don't slash more than what's still needed
			slashAmount = math.MinInt(slashAmount, remaining)

			if slashAmount.IsPositive() {
				// Calculate slash fraction based on actual amount
				slashFraction := math.LegacyNewDecFromInt(slashAmount).Quo(math.LegacyNewDecFromInt(mod.Stake))

				// Execute the slash
				if err := k.SlashModerator(ctx, "system", vote.Moderator, slashFraction, "wrong dispute resolution - funds recovery"); err != nil {
					k.Logger(ctx).Error("failed to slash moderator for recovery",
						"moderator", vote.Moderator,
						"amount", slashAmount.String(),
						"error", err)
					continue
				}

				totalSlashed = totalSlashed.Add(slashAmount)
				remaining = remaining.Sub(slashAmount)

				k.Logger(ctx).Info("slashed moderator for fund recovery",
					"moderator", vote.Moderator,
					"amount", slashAmount.String(),
					"remaining_needed", remaining.String())

				// If we have enough, stop slashing
				if remaining.IsZero() || remaining.IsNegative() {
					break
				}
			}
		}
	}

	return totalSlashed
}

// penalizeModeratorsForWrongResolution penalizes all moderators who voted for the wrong resolution
func (k Keeper) penalizeModeratorsForWrongResolution(ctx sdk.Context, dispute types.Dispute) {
	for _, vote := range dispute.Votes {
		if vote.Vote == dispute.Resolution {
			// This moderator voted for the resolution that was overturned

			// Update metrics
			metrics := k.GetOrCreateModeratorMetrics(ctx, vote.Moderator)
			metrics.RecordDecision(true, ctx.BlockTime()) // Mark as overturned

			// Check for auto-blacklist
			if shouldBlacklist, reason := metrics.ShouldAutoBlacklist(); shouldBlacklist {
				k.autoBlacklistModerator(ctx, vote.Moderator, reason, &metrics)
			}
			k.SetModeratorMetrics(ctx, metrics)

			// Penalize reputation via staking keeper
			k.penalizeModeratorReputation(ctx, vote.Moderator, dispute.ID)

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeModeratorPenalized,
					sdk.NewAttribute(types.AttributeKeyModerator, vote.Moderator),
					sdk.NewAttribute(types.AttributeKeyDisputeID, fmt.Sprintf("%d", dispute.ID)),
					sdk.NewAttribute(types.AttributeKeyReason, "wrong_resolution_overturned"),
				),
			)
		}
	}
}

// rewardReporter rewards a reporter for a confirmed report
func (k Keeper) rewardReporter(ctx sdk.Context, report *types.Report) {
	if k.stakingKeeper == nil {
		return
	}

	reporterAddr, err := sdk.AccAddressFromBech32(report.Reporter)
	if err != nil {
		return
	}

	// Reward reputation via staking keeper
	if err := k.stakingKeeper.RewardSuccessfulDispute(ctx, reporterAddr, fmt.Sprintf("report_%d", report.ID)); err != nil {
		k.Logger(ctx).Error("failed to reward reporter", "error", err)
	}
}

// penalizeReporter penalizes a reporter for a dismissed report (reputation only)
func (k Keeper) penalizeReporter(ctx sdk.Context, report *types.Report) {
	if k.stakingKeeper == nil {
		return
	}

	reporterAddr, err := sdk.AccAddressFromBech32(report.Reporter)
	if err != nil {
		return
	}

	// Penalize reputation (no slashing for good-faith false reports)
	if err := k.stakingKeeper.PenalizeBadDispute(ctx, reporterAddr, fmt.Sprintf("report_%d", report.ID)); err != nil {
		k.Logger(ctx).Error("failed to penalize reporter", "error", err)
	}
}

// =============================================================================
// REPORT QUERIES
// =============================================================================

// GetReportsByTarget returns all reports for a specific target
func (k Keeper) GetReportsByTarget(ctx sdk.Context, targetType, targetID string) []types.Report {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetReportByTargetPrefixKey(targetType, targetID)
	iterator := prefix.NewStore(store, prefixKey).Iterator(nil, nil)
	defer iterator.Close()

	var reports []types.Report
	for ; iterator.Valid(); iterator.Next() {
		reportID := sdk.BigEndianToUint64(iterator.Value())
		if report, found := k.GetReport(ctx, reportID); found {
			reports = append(reports, report)
		}
	}

	return reports
}

// GetReportsByReporter returns all reports by a specific reporter
func (k Keeper) GetReportsByReporter(ctx sdk.Context, reporter string) []types.Report {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetReportByReporterPrefixKey(reporter)
	iterator := prefix.NewStore(store, prefixKey).Iterator(nil, nil)
	defer iterator.Close()

	var reports []types.Report
	for ; iterator.Valid(); iterator.Next() {
		reportID := sdk.BigEndianToUint64(iterator.Value())
		if report, found := k.GetReport(ctx, reportID); found {
			reports = append(reports, report)
		}
	}

	return reports
}

// GetReportsByStatus returns all reports with a specific status
func (k Keeper) GetReportsByStatus(ctx sdk.Context, status types.ReportStatus) []types.Report {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetReportByStatusPrefixKey(status.String())
	iterator := prefix.NewStore(store, prefixKey).Iterator(nil, nil)
	defer iterator.Close()

	var reports []types.Report
	for ; iterator.Valid(); iterator.Next() {
		reportID := sdk.BigEndianToUint64(iterator.Value())
		if report, found := k.GetReport(ctx, reportID); found {
			reports = append(reports, report)
		}
	}

	return reports
}

// GetAllReports returns all reports for genesis export
func (k Keeper) GetAllReports(ctx sdk.Context) []types.Report {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.ReportPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var reports []types.Report
	for ; iterator.Valid(); iterator.Next() {
		var report types.Report
		if err := json.Unmarshal(iterator.Value(), &report); err != nil {
			continue
		}
		reports = append(reports, report)
	}

	return reports
}

// =============================================================================
// MODERATOR METRICS TRACKING
// =============================================================================

// GetModeratorMetrics returns metrics for a moderator
func (k Keeper) GetModeratorMetrics(ctx sdk.Context, address string) (types.ModeratorMetrics, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetModeratorMetricsKey(address))
	if bz == nil {
		return types.ModeratorMetrics{}, false
	}

	var metrics types.ModeratorMetrics
	if err := json.Unmarshal(bz, &metrics); err != nil {
		return types.ModeratorMetrics{}, false
	}
	return metrics, true
}

// SetModeratorMetrics stores moderator metrics
func (k Keeper) SetModeratorMetrics(ctx sdk.Context, metrics types.ModeratorMetrics) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal moderator metrics: %w", err)
	}
	store.Set(types.GetModeratorMetricsKey(metrics.Address), bz)
	return nil
}

// GetOrCreateModeratorMetrics returns metrics, creating if needed
func (k Keeper) GetOrCreateModeratorMetrics(ctx sdk.Context, address string) types.ModeratorMetrics {
	metrics, found := k.GetModeratorMetrics(ctx, address)
	if !found {
		metrics = types.NewModeratorMetrics(address)
		k.SetModeratorMetrics(ctx, metrics)
	}
	return metrics
}

// RecordModeratorDecision records a moderator's decision for metrics
func (k Keeper) RecordModeratorDecision(ctx sdk.Context, moderator string, overturned bool) {
	metrics := k.GetOrCreateModeratorMetrics(ctx, moderator)
	metrics.RecordDecision(overturned, ctx.BlockTime())

	// Check if auto-blacklist should trigger
	if shouldBlacklist, reason := metrics.ShouldAutoBlacklist(); shouldBlacklist {
		k.autoBlacklistModerator(ctx, moderator, reason, &metrics)
	}

	k.SetModeratorMetrics(ctx, metrics)

	// If overturned, also update moderator reputation
	if overturned {
		k.penalizeModeratorReputation(ctx, moderator, 0)
	}
}

// recordModeratorReportConfirmed records a confirmed report against a moderator
func (k Keeper) recordModeratorReportConfirmed(ctx sdk.Context, moderator string) {
	metrics := k.GetOrCreateModeratorMetrics(ctx, moderator)
	metrics.ReportsAgainst++
	metrics.ConfirmedReports++

	// Check if auto-blacklist should trigger
	if shouldBlacklist, reason := metrics.ShouldAutoBlacklist(); shouldBlacklist {
		k.autoBlacklistModerator(ctx, moderator, reason, &metrics)
	}

	k.SetModeratorMetrics(ctx, metrics)
}

// autoBlacklistModerator automatically blacklists a moderator based on metrics
func (k Keeper) autoBlacklistModerator(ctx sdk.Context, moderator, reason string, metrics *types.ModeratorMetrics) {
	// Determine ban parameters based on metrics
	permanent := metrics.ConfirmedReports >= 2 // Permanent for 2+ confirmed fraud reports
	slashFraction := math.LegacyZeroDec()

	// Slash only after 3+ overturns
	if metrics.ConsecutiveOverturns >= 3 {
		slashFraction = math.LegacyNewDecWithPrec(5, 2) // 5% slash for 3 overturns
		if metrics.ConsecutiveOverturns >= 5 {
			slashFraction = math.LegacyNewDecWithPrec(10, 2) // 10% slash for 5+ overturns
		}
	}

	// Execute slash if needed
	if slashFraction.IsPositive() {
		k.SlashModerator(ctx, "system", moderator, slashFraction, reason)
	}

	// Blacklist moderator
	// For auto-blacklist, use 14-day temporary ban unless it's permanent
	banDuration := 14 * 24 * 60 * 60 * 1000000000 // 14 days in nanoseconds (as time.Duration)
	k.BlacklistModerator(ctx, "system", moderator, reason, permanent, 0) // Use 0 for now, let ProcessExpiredBans handle

	// If temporary, set the unban time
	if !permanent {
		if blacklist, found := k.GetModeratorBlacklist(ctx, moderator); found {
			blacklist.UnbanAt = ctx.BlockTime().Add(14 * 24 * 60 * 60 * 1000000000) // 14 days
			k.SetModeratorBlacklist(ctx, blacklist)
		}
	}

	// Issue warning
	metrics.WarningCount++
	metrics.LastWarningAt = ctx.BlockTime()

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeModeratorAutoBlacklisted,
			sdk.NewAttribute(types.AttributeKeyModerator, moderator),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
			sdk.NewAttribute(types.AttributeKeyPermanent, fmt.Sprintf("%t", permanent)),
			sdk.NewAttribute(types.AttributeKeyOverturnRate, metrics.OverturnRate.String()),
			sdk.NewAttribute(types.AttributeKeyWarningCount, fmt.Sprintf("%d", metrics.WarningCount)),
		),
	)

	// Suppress unused variable warning
	_ = banDuration
}

// ProcessModeratorMetrics checks all active moderators for auto-blacklist conditions
// Called from EndBlock
func (k Keeper) ProcessModeratorMetrics(ctx sdk.Context) {
	moderators := k.GetAllModerators(ctx)

	for _, mod := range moderators {
		if mod.Blacklisted {
			continue
		}

		metrics := k.GetOrCreateModeratorMetrics(ctx, mod.Address)

		// Check auto-blacklist conditions
		if shouldBlacklist, reason := metrics.ShouldAutoBlacklist(); shouldBlacklist {
			k.autoBlacklistModerator(ctx, mod.Address, reason, &metrics)
			k.SetModeratorMetrics(ctx, metrics)
		}
	}
}

// GetAllModeratorMetrics returns all moderator metrics for genesis export
func (k Keeper) GetAllModeratorMetrics(ctx sdk.Context) []types.ModeratorMetrics {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.ModeratorMetricsPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var allMetrics []types.ModeratorMetrics
	for ; iterator.Valid(); iterator.Next() {
		var metrics types.ModeratorMetrics
		if err := json.Unmarshal(iterator.Value(), &metrics); err != nil {
			continue
		}
		allMetrics = append(allMetrics, metrics)
	}

	return allMetrics
}

// =============================================================================
// ESCROW RESERVE FUND
// =============================================================================

// GetEscrowReserve returns the escrow reserve fund
func (k Keeper) GetEscrowReserve(ctx sdk.Context) types.EscrowReserve {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.EscrowReserveKey)
	if bz == nil {
		return types.NewEscrowReserve()
	}

	var reserve types.EscrowReserve
	if err := json.Unmarshal(bz, &reserve); err != nil {
		return types.NewEscrowReserve()
	}
	return reserve
}

// SetEscrowReserve stores the escrow reserve fund
func (k Keeper) SetEscrowReserve(ctx sdk.Context, reserve types.EscrowReserve) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(reserve)
	if err != nil {
		return fmt.Errorf("failed to marshal escrow reserve: %w", err)
	}
	store.Set(types.EscrowReserveKey, bz)
	return nil
}

// ContributeToReserve adds funds to the escrow reserve (called from escrow fees)
func (k Keeper) ContributeToReserve(ctx sdk.Context, amount math.Int) error {
	reserve := k.GetEscrowReserve(ctx)
	reserve.TotalFunds = reserve.TotalFunds.Add(amount)

	if err := k.SetEscrowReserve(ctx, reserve); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReserveContribution,
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute("total_reserve", reserve.TotalFunds.String()),
		),
	)

	return nil
}

// =============================================================================
// EXPIRED REPORTS PROCESSING
// =============================================================================

// ProcessExpiredReports processes reports that have passed their deadline
// Called from EndBlock
//
// NEW BEHAVIOR: Instead of auto-dismissing, we use escalation to prevent false negatives:
// 1. If ANY vote was "confirmed" (fraud likely), ESCALATE to higher tier
// 2. If ALL votes were "dismiss", then safe to auto-dismiss
// 3. If NO votes at all, extend deadline once and reassign reviewers
// 4. Track stale reports that keep failing to get votes
func (k Keeper) ProcessExpiredReports(ctx sdk.Context) {
	reports := k.GetReportsByStatus(ctx, types.ReportStatusUnderInvestigation)

	for _, report := range reports {
		if ctx.BlockTime().After(report.DeadlineAt) {
			// Count vote types
			confirmVotes := 0
			dismissVotes := 0

			for _, vote := range report.ReviewVotes {
				if vote.Confirmed {
					confirmVotes++
				} else {
					dismissVotes++
				}
			}

			// Decision logic based on votes received
			if len(report.ReviewVotes) == 0 {
				// NO VOTES AT ALL - extend deadline and reassign
				k.extendReportDeadline(ctx, &report)
			} else if confirmVotes > 0 {
				// AT LEAST ONE REVIEWER SAW MERIT - ESCALATE
				// This prevents false negatives where real fraud slips through
				k.escalateReportToHigherTier(ctx, &report)
			} else if dismissVotes > 0 {
				// ALL votes were dismiss - safe to auto-dismiss
				k.autoDismissReport(ctx, &report)
			} else {
				// Shouldn't happen, but treat as no votes
				k.extendReportDeadline(ctx, &report)
			}
		}
	}
}

// =============================================================================
// REPORT ESCALATION SYSTEM - Prevents False Negatives
// =============================================================================

// extendReportDeadline extends deadline by 3 days and reassigns reviewers
// Used when no votes were received - gives reviewers more time
func (k Keeper) extendReportDeadline(ctx sdk.Context, report *types.Report) {
	// Check if we've exceeded maximum extensions
	if report.ExtensionCount >= types.MaxReportExtensions {
		// Too many extensions without votes - escalate instead
		k.Logger(ctx).Warn("report exceeded max extensions, escalating",
			"report_id", report.ID,
			"extension_count", report.ExtensionCount)
		k.escalateReportToHigherTier(ctx, report)
		return
	}

	// Extend deadline
	oldDeadline := report.DeadlineAt
	report.ExtensionCount++
	report.DeadlineAt = ctx.BlockTime().Add(types.ExtensionDuration)

	// Reassign different reviewers (current ones didn't respond)
	// Clear old assignments and assign fresh reviewers
	report.AssignedReviewers = []string{}
	if err := k.assignReviewers(ctx, report); err != nil {
		k.Logger(ctx).Error("failed to reassign reviewers during extension",
			"report_id", report.ID,
			"error", err)
		// Mark as stale if we can't assign reviewers
		report.ActionsTaken = append(report.ActionsTaken,
			fmt.Sprintf("stale_report: no reviewers available for extension %d", report.ExtensionCount))
		return
	}

	report.ActionsTaken = append(report.ActionsTaken,
		fmt.Sprintf("deadline_extended: extension %d, new deadline %s",
			report.ExtensionCount, report.DeadlineAt.Format(time.RFC3339)))

	k.SetReport(ctx, *report)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReportDeadlineExtended,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", report.ID)),
			sdk.NewAttribute("extension_count", fmt.Sprintf("%d", report.ExtensionCount)),
			sdk.NewAttribute("old_deadline", oldDeadline.Format(time.RFC3339)),
			sdk.NewAttribute("new_deadline", report.DeadlineAt.Format(time.RFC3339)),
		),
	)

	k.Logger(ctx).Info("report deadline extended",
		"report_id", report.ID,
		"extension_count", report.ExtensionCount,
		"new_deadline", report.DeadlineAt.Format(time.RFC3339))
}

// escalateReportToHigherTier moves report to next tier for review
// Used when at least one reviewer saw merit - prevents false negatives
func (k Keeper) escalateReportToHigherTier(ctx sdk.Context, report *types.Report) {
	// Check if we've exceeded maximum escalations
	if report.EscalationCount >= types.MaxReportEscalations {
		// Max escalations reached - escalate to governance
		k.escalateReportToGovernance(ctx, report)
		return
	}

	// Save current votes to history
	report.PreviousVotes = append(report.PreviousVotes, report.ReviewVotes...)

	// Determine next tier
	oldTier := report.CurrentTier
	var newTier int
	var tierName string

	switch oldTier {
	case types.StakeTierWarden: // 2 -> 3
		newTier = types.StakeTierSteward
		tierName = "Steward"
	case types.StakeTierSteward: // 3 -> 4
		newTier = types.StakeTierArchon
		tierName = "Archon"
	case types.StakeTierArchon: // 4 -> governance
		k.escalateReportToGovernance(ctx, report)
		return
	default:
		// Already at or above Archon - escalate to governance
		k.escalateReportToGovernance(ctx, report)
		return
	}

	// Update report for higher tier review
	report.EscalationCount++
	report.CurrentTier = newTier
	report.EscalatedAt = ctx.BlockTime()
	report.ReviewVotes = []types.ReviewVote{} // Clear votes for fresh review
	report.AssignedReviewers = []string{} // Clear old assignments

	// Assign new higher-tier reviewers
	if err := k.assignReviewers(ctx, report); err != nil {
		k.Logger(ctx).Error("failed to assign higher-tier reviewers",
			"report_id", report.ID,
			"new_tier", newTier,
			"error", err)
		// If we can't assign higher tier, try governance
		k.escalateReportToGovernance(ctx, report)
		return
	}

	// Extend deadline for new reviewers (give them fresh time)
	report.DeadlineAt = ctx.BlockTime().Add(types.ExtensionDuration)

	report.ActionsTaken = append(report.ActionsTaken,
		fmt.Sprintf("escalated_to_tier_%d: %s tier review (escalation %d)",
			newTier, tierName, report.EscalationCount))

	k.SetReport(ctx, *report)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReportEscalated,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", report.ID)),
			sdk.NewAttribute("escalation_count", fmt.Sprintf("%d", report.EscalationCount)),
			sdk.NewAttribute("old_tier", fmt.Sprintf("%d", oldTier)),
			sdk.NewAttribute("new_tier", fmt.Sprintf("%d", newTier)),
			sdk.NewAttribute("tier_name", tierName),
			sdk.NewAttribute("previous_votes_count", fmt.Sprintf("%d", len(report.PreviousVotes))),
		),
	)

	k.Logger(ctx).Info("report escalated to higher tier",
		"report_id", report.ID,
		"old_tier", oldTier,
		"new_tier", newTier,
		"tier_name", tierName,
		"escalation_count", report.EscalationCount)
}

// escalateReportToGovernance escalates report to on-chain governance vote
// Final escalation level when even Archons can't reach consensus
func (k Keeper) escalateReportToGovernance(ctx sdk.Context, report *types.Report) {
	// Mark report as requiring governance decision
	oldStatus := report.Status
	report.Status = types.ReportStatusAppealed // Reuse appeal status for governance
	report.Resolution = fmt.Sprintf(
		"Escalated to governance: report reached max escalations (%d), requires community vote",
		report.EscalationCount,
	)

	report.ActionsTaken = append(report.ActionsTaken,
		fmt.Sprintf("escalated_to_governance: max escalations reached (%d), community vote required",
			report.EscalationCount))

	k.SetReport(ctx, *report)
	k.UpdateReportStatusIndex(ctx, *report, oldStatus)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReportEscalatedToGovernance,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", report.ID)),
			sdk.NewAttribute("escalation_count", fmt.Sprintf("%d", report.EscalationCount)),
			sdk.NewAttribute("original_tier", fmt.Sprintf("%d", report.OriginalTier)),
			sdk.NewAttribute("report_type", report.ReportType.String()),
			sdk.NewAttribute("target_type", report.TargetType.String()),
			sdk.NewAttribute("target_id", report.TargetID),
		),
	)

	k.Logger(ctx).Warn("report escalated to governance",
		"report_id", report.ID,
		"report_type", report.ReportType.String(),
		"target_type", report.TargetType.String(),
		"target_id", report.TargetID,
		"escalation_count", report.EscalationCount,
		"reason", "max escalations reached, requires community vote")
}

// autoDismissReport safely dismisses when all votes were dismiss
// Only called when ALL reviewers agreed the report was invalid
func (k Keeper) autoDismissReport(ctx sdk.Context, report *types.Report) {
	oldStatus := report.Status
	report.Status = types.ReportStatusDismissed
	report.Resolution = fmt.Sprintf(
		"Auto-dismissed: deadline passed with all %d reviewers voting to dismiss (no fraud detected)",
		len(report.ReviewVotes),
	)
	report.ResolvedAt = ctx.BlockTime()

	// Apply escalating penalties to reporter for false report
	k.ApplyReporterPenalty(ctx, report)
	k.penalizeReporter(ctx, report)

	// Log dismissal reasons from reviewers
	dismissalReasons := []string{}
	for _, vote := range report.ReviewVotes {
		if vote.Comments != "" {
			dismissalReasons = append(dismissalReasons,
				fmt.Sprintf("%s: %s", vote.Reviewer, vote.Comments))
		}
	}
	if len(dismissalReasons) > 0 {
		report.ActionsTaken = append(report.ActionsTaken,
			fmt.Sprintf("reviewer_consensus: all voted dismiss - %s",
				fmt.Sprintf("%v", dismissalReasons)))
	}

	k.SetReport(ctx, *report)
	k.UpdateReportStatusIndex(ctx, *report, oldStatus)

	// If this was a company fraud/scam report, restore the company
	if (report.ReportType == types.ReportTypeFraud || report.ReportType == types.ReportTypeScam) &&
		report.TargetType == types.ReportTargetTypeCompany {
		var companyID uint64
		if _, err := fmt.Sscanf(report.TargetID, "%d", &companyID); err == nil {
			if k.equityKeeper != nil {
				if err := k.equityKeeper.ConfirmFraudAndDelist(ctx, companyID, report.ID, false, "escrow_report_system_auto_dismiss"); err != nil {
					k.Logger(ctx).Error("failed to restore company after auto-dismissed report",
						"company_id", companyID,
						"report_id", report.ID,
						"error", err)
				} else {
					report.ActionsTaken = append(report.ActionsTaken, fmt.Sprintf("company_%d_restored_trading_resumed", companyID))
					k.SetReport(ctx, *report)
					k.Logger(ctx).Info("company restored after report auto-dismissed - trading resumed",
						"company_id", companyID,
						"report_id", report.ID)
				}
			}
		}
	}

	// Emit resolution event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReportResolved,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", report.ID)),
			sdk.NewAttribute(types.AttributeKeyStatus, report.Status.String()),
			sdk.NewAttribute(types.AttributeKeyConfirmed, "false"),
			sdk.NewAttribute("dismiss_votes", fmt.Sprintf("%d", len(report.ReviewVotes))),
			sdk.NewAttribute("auto_dismissed", "true"),
		),
	)

	k.Logger(ctx).Info("report auto-dismissed with full reviewer consensus",
		"report_id", report.ID,
		"votes_received", len(report.ReviewVotes),
		"votes_required", report.VotesRequired)
}

// =============================================================================
// PRIORITY SORTED REPORTS
// =============================================================================

// GetOpenReportsByPriority returns open reports sorted by priority (highest first)
func (k Keeper) GetOpenReportsByPriority(ctx sdk.Context) []types.Report {
	reports := k.GetReportsByStatus(ctx, types.ReportStatusOpen)

	// Sort by priority descending
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Priority > reports[j].Priority
	})

	return reports
}

// =============================================================================
// REPORTER HISTORY - Anti-Spam & Abuse Tracking
// =============================================================================

// GetReporterHistory returns the history for a reporter
func (k Keeper) GetReporterHistory(ctx sdk.Context, address string) (types.ReporterHistory, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetReporterHistoryKey(address))
	if bz == nil {
		return types.ReporterHistory{}, false
	}

	var history types.ReporterHistory
	if err := json.Unmarshal(bz, &history); err != nil {
		return types.ReporterHistory{}, false
	}
	return history, true
}

// SetReporterHistory stores reporter history
func (k Keeper) SetReporterHistory(ctx sdk.Context, history types.ReporterHistory) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal reporter history: %w", err)
	}
	store.Set(types.GetReporterHistoryKey(history.Address), bz)

	// Also index if banned
	if history.IsBanned {
		store.Set(types.GetReporterBanKey(history.Address), []byte{1})
	} else {
		store.Delete(types.GetReporterBanKey(history.Address))
	}

	return nil
}

// GetOrCreateReporterHistory returns history, creating if needed
func (k Keeper) GetOrCreateReporterHistory(ctx sdk.Context, address string) types.ReporterHistory {
	history, found := k.GetReporterHistory(ctx, address)
	if !found {
		history = types.NewReporterHistory(address)
		k.SetReporterHistory(ctx, history)
	}
	return history
}

// CheckReporterCanSubmit checks if a reporter is allowed to submit reports
// Returns error if blocked/banned/rate-limited
func (k Keeper) CheckReporterCanSubmit(ctx sdk.Context, reporter string) error {
	history, found := k.GetReporterHistory(ctx, reporter)
	if !found {
		// New reporter, always allowed
		return nil
	}

	// Check for expired bans and unban if needed
	if history.IsBanned && !history.BanExpiresAt.IsZero() && ctx.BlockTime().After(history.BanExpiresAt) {
		history.Unban()
		k.SetReporterHistory(ctx, history)
		k.Logger(ctx).Info("reporter ban expired",
			"reporter", reporter,
			"was_banned_for", history.BanReason)
	}

	// Check if blocked
	blocked, reason := history.ShouldBlockReport(ctx.BlockTime())
	if blocked {
		if history.IsBanned {
			return types.ErrReporterBanned
		}
		if history.ConsecutiveDismissed >= 3 {
			return types.ErrReporterCooldown
		}
		return fmt.Errorf("%w: %s", types.ErrReporterRateLimited, reason)
	}

	// ANTI-SYBIL: Check stake age (prevents banned users from creating new accounts)
	// Must be staked at Keeper+ tier for at least 7 days to submit reports
	if k.stakingKeeper != nil {
		reporterAddr, err := sdk.AccAddressFromBech32(reporter)
		if err == nil {
			stakeAge := k.stakingKeeper.GetStakeAge(ctx, reporterAddr)
			if stakeAge < types.MinStakeAgeForReporting {
				daysRemaining := (types.MinStakeAgeForReporting - stakeAge) / (24 * 60 * 60)
				k.Logger(ctx).Info("reporter stake age too low",
					"reporter", reporter,
					"stake_age_seconds", stakeAge,
					"required_seconds", types.MinStakeAgeForReporting,
					"days_remaining", daysRemaining)
				return types.ErrStakeAgeTooLow
			}
		}
	}

	return nil
}

// ApplyReporterPenalty applies escalating penalties based on reporter history
func (k Keeper) ApplyReporterPenalty(ctx sdk.Context, report *types.Report) {
	history := k.GetOrCreateReporterHistory(ctx, report.Reporter)

	// Record this dismissed report
	history.RecordReport(false, ctx.BlockTime())

	// Get escalated penalty based on history
	reputationPenalty, slashPercent, shouldBan, banDuration := history.GetEscalatedPenalty(report.ReportType)

	// Apply reputation penalty (escalated)
	report.ReporterReputationChange = -int(reputationPenalty)
	history.TotalReputationLost += int64(reputationPenalty)

	// Apply stake slash if warranted
	if slashPercent.IsPositive() && k.stakingKeeper != nil {
		reporterAddr, err := sdk.AccAddressFromBech32(report.Reporter)
		if err == nil {
			// Record the slash for tracking
			report.ActionsTaken = append(report.ActionsTaken,
				fmt.Sprintf("stake_slash_applied: %s%%", slashPercent.MulInt64(100).String()))

			// Actual slashing would go through staking keeper
			// For now, track it in history
			// slashAmount would be calculated from user's stake
			k.Logger(ctx).Warn("applying escalated stake slash for false report",
				"reporter", report.Reporter,
				"slash_percent", slashPercent.String(),
				"consecutive_dismissed", history.ConsecutiveDismissed)

			_ = reporterAddr // Would use for actual slashing
		}
	}

	// Issue warning
	history.WarningCount++

	// Apply ban if warranted
	if shouldBan {
		banReason := fmt.Sprintf("repeated false reports (%d consecutive dismissed)", history.ConsecutiveDismissed)
		if history.FalseReportRate.GT(math.LegacyNewDecWithPrec(80, 2)) {
			banReason = fmt.Sprintf("systematic abuse (%.0f%% false report rate)", history.FalseReportRate.MulInt64(100).MustFloat64())
		}

		history.Ban(banReason, banDuration, ctx.BlockTime())

		report.ActionsTaken = append(report.ActionsTaken, fmt.Sprintf("reporter_banned: %s", banReason))

		// Emit ban event
		banType := "temporary"
		if banDuration == 0 {
			banType = "permanent"
		}
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"reporter_banned",
				sdk.NewAttribute(types.AttributeKeyReporter, report.Reporter),
				sdk.NewAttribute(types.AttributeKeyReason, banReason),
				sdk.NewAttribute("ban_type", banType),
				sdk.NewAttribute("consecutive_dismissed", fmt.Sprintf("%d", history.ConsecutiveDismissed)),
				sdk.NewAttribute("total_dismissed", fmt.Sprintf("%d", history.DismissedReports)),
			),
		)

		k.Logger(ctx).Warn("reporter banned for abuse",
			"reporter", report.Reporter,
			"reason", banReason,
			"ban_type", banType,
			"duration", banDuration.String())
	}

	// Save updated history
	k.SetReporterHistory(ctx, history)
}

// RecordConfirmedReport records a confirmed report in reporter history
func (k Keeper) RecordConfirmedReport(ctx sdk.Context, reporter string) {
	history := k.GetOrCreateReporterHistory(ctx, reporter)
	history.RecordReport(true, ctx.BlockTime())
	k.SetReporterHistory(ctx, history)
}

// GetAllReporterHistories returns all reporter histories for genesis export
func (k Keeper) GetAllReporterHistories(ctx sdk.Context) []types.ReporterHistory {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.ReporterHistoryPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var histories []types.ReporterHistory
	for ; iterator.Valid(); iterator.Next() {
		var history types.ReporterHistory
		if err := json.Unmarshal(iterator.Value(), &history); err != nil {
			continue
		}
		histories = append(histories, history)
	}

	return histories
}

// ProcessExpiredReporterBans lifts expired temporary bans
// Called from EndBlock
func (k Keeper) ProcessExpiredReporterBans(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.ReporterBanPrefix).Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		address := string(iterator.Key())
		history, found := k.GetReporterHistory(ctx, address)
		if !found {
			continue
		}

		// Check if temporary ban has expired
		if history.IsBanned && !history.BanExpiresAt.IsZero() && ctx.BlockTime().After(history.BanExpiresAt) {
			history.Unban()
			k.SetReporterHistory(ctx, history)

			k.Logger(ctx).Info("reporter temporary ban expired",
				"reporter", address,
				"was_banned_for", history.BanReason)
		}
	}
}

// =============================================================================
// BAN CHECKING FOR EXTERNAL MODULES (e.g., extbridge)
// =============================================================================

// IsAddressBanned checks if an address is banned from the platform
// Used by extbridge to prevent banned addresses from using the bridge
// Returns (isBanned bool, reason string)
func (k Keeper) IsAddressBanned(ctx sdk.Context, address string, blockTime time.Time) (bool, string) {
	// Check reporter history for ban status
	history, found := k.GetReporterHistory(ctx, address)
	if !found {
		// No history means not banned
		return false, ""
	}

	// Check if permanently banned
	if history.IsBanned && history.BanExpiresAt.IsZero() {
		return true, fmt.Sprintf("permanently banned: %s", history.BanReason)
	}

	// Check if temporarily banned (and ban is still active)
	if history.IsBanned && blockTime.Before(history.BanExpiresAt) {
		remaining := history.BanExpiresAt.Sub(blockTime)
		return true, fmt.Sprintf("temporarily banned until %s (%v remaining): %s",
			history.BanExpiresAt.Format(time.RFC3339),
			remaining.Round(time.Hour),
			history.BanReason,
		)
	}

	// Check moderator blacklist status
	mod, modFound := k.GetModerator(ctx, address)
	if modFound && mod.Blacklisted {
		return true, "blacklisted as moderator"
	}

	return false, ""
}

// =============================================================================
// RECURSIVE REPORT PREVENTION (Issue #10)
// =============================================================================

// GetActiveReportsByReporter returns all non-resolved reports by this reporter
func (k Keeper) GetActiveReportsByReporter(ctx sdk.Context, reporter string) []types.Report {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetReportByReporterPrefixKey(reporter)
	iterator := prefix.NewStore(store, prefixKey).Iterator(nil, nil)
	defer iterator.Close()

	var activeReports []types.Report
	for ; iterator.Valid(); iterator.Next() {
		reportID := sdk.BigEndianToUint64(iterator.Value())
		if report, found := k.GetReport(ctx, reportID); found {
			// Only include active reports (not resolved)
			if report.Status != types.ReportStatusConfirmed &&
				report.Status != types.ReportStatusDismissed &&
				report.Status != types.ReportStatusVoluntarilyResolved {
				activeReports = append(activeReports, report)
			}
		}
	}

	return activeReports
}

// wasRecentlyTargeted checks if address was target of a report within cooldown period
func (k Keeper) wasRecentlyTargeted(ctx sdk.Context, address string, cooldown time.Duration) bool {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetReportByTargetPrefixKey(types.ReportTargetTypeUser.String(), address)
	iterator := prefix.NewStore(store, prefixKey).Iterator(nil, nil)
	defer iterator.Close()

	cutoffTime := ctx.BlockTime().Add(-cooldown)

	for ; iterator.Valid(); iterator.Next() {
		reportID := sdk.BigEndianToUint64(iterator.Value())
		if report, found := k.GetReport(ctx, reportID); found {
			// Check if report was created within the cooldown period
			if report.CreatedAt.After(cutoffTime) {
				return true
			}
		}
	}

	return false
}
