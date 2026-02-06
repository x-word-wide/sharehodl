package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/escrow/types"
)

// =============================================================================
// PHASE 2: ENHANCED APPEALS SYSTEM
// =============================================================================

// =============================================================================
// APPEAL STORAGE
// =============================================================================

// GetAppeal returns an appeal by ID
func (k Keeper) GetAppeal(ctx sdk.Context, appealID uint64) (types.Appeal, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetAppealKey(appealID))
	if bz == nil {
		return types.Appeal{}, false
	}

	var appeal types.Appeal
	if err := json.Unmarshal(bz, &appeal); err != nil {
		return types.Appeal{}, false
	}
	return appeal, true
}

// SetAppeal stores an appeal
func (k Keeper) SetAppeal(ctx sdk.Context, appeal types.Appeal) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(appeal)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal appeal", "error", err)
		return fmt.Errorf("failed to marshal appeal: %w", err)
	}
	store.Set(types.GetAppealKey(appeal.ID), bz)
	return nil
}

// DeleteAppeal removes an appeal
func (k Keeper) DeleteAppeal(ctx sdk.Context, appealID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetAppealKey(appealID))
}

// IndexAppeal creates indexes for efficient querying
func (k Keeper) IndexAppeal(ctx sdk.Context, appeal types.Appeal) {
	store := ctx.KVStore(k.storeKey)

	// Index by dispute
	if appeal.DisputeID > 0 {
		key := types.GetAppealByDisputeKey(appeal.DisputeID, appeal.ID)
		store.Set(key, sdk.Uint64ToBigEndian(appeal.ID))
	}

	// Index by report
	if appeal.ReportID > 0 {
		key := types.GetAppealByReportKey(appeal.ReportID, appeal.ID)
		store.Set(key, sdk.Uint64ToBigEndian(appeal.ID))
	}
}

// =============================================================================
// DISPUTE APPEALS (ENHANCED)
// =============================================================================

// SubmitDisputeAppeal submits an appeal for a dispute resolution
func (k Keeper) SubmitDisputeAppeal(
	ctx sdk.Context,
	disputeID uint64,
	appellant string,
	reason string,
) (uint64, error) {
	dispute, found := k.GetDispute(ctx, disputeID)
	if !found {
		return 0, types.ErrDisputeNotFound
	}

	// Validate appellant is a participant
	escrow, found := k.GetEscrow(ctx, dispute.EscrowID)
	if !found {
		return 0, types.ErrEscrowNotFound
	}

	if appellant != escrow.Sender && appellant != escrow.Recipient {
		return 0, types.ErrAppealerNotParticipant
	}

	// Check dispute was resolved (can't appeal ongoing)
	if dispute.Status != types.DisputeStatusResolved {
		return 0, types.ErrCannotAppeal
	}

	// Get existing appeals for this dispute
	existingAppeals := k.GetAppealsByDispute(ctx, disputeID)
	appealLevel := len(existingAppeals) + 1

	// Check max appeal level
	if appealLevel > 3 {
		return 0, types.ErrAppealLevelMaxed
	}

	// Create appeal
	appealID := k.GetNextAppealID(ctx)
	appeal := types.NewAppeal(
		appealID,
		disputeID,
		0, // Not a report appeal
		appellant,
		reason,
		appealLevel,
		dispute.Resolution,
		ctx.BlockTime(),
	)

	// Validate appeal
	if err := appeal.Validate(); err != nil {
		return 0, types.ErrInvalidAppeal
	}

	// Store appeal
	if err := k.SetAppeal(ctx, appeal); err != nil {
		return 0, err
	}

	// Create indexes
	k.IndexAppeal(ctx, appeal)

	// Update dispute status
	dispute.Status = types.DisputeStatusAppealed
	dispute.AppealsCount++
	k.SetDispute(ctx, dispute)

	// Assign appeal reviewers
	if err := k.assignAppealReviewers(ctx, &appeal); err != nil {
		k.Logger(ctx).Error("failed to assign appeal reviewers", "appeal_id", appealID, "error", err)
	} else {
		appeal.Status = types.AppealStatusReviewing
		k.SetAppeal(ctx, appeal)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAppealSubmitted,
			sdk.NewAttribute(types.AttributeKeyAppealID, fmt.Sprintf("%d", appealID)),
			sdk.NewAttribute(types.AttributeKeyDisputeID, fmt.Sprintf("%d", disputeID)),
			sdk.NewAttribute(types.AttributeKeyAppellant, appellant),
			sdk.NewAttribute(types.AttributeKeyAppealLevel, fmt.Sprintf("%d", appealLevel)),
		),
	)

	return appealID, nil
}

// assignAppealReviewers assigns reviewers to an appeal (Issue #12: Use unpredictable randomization)
func (k Keeper) assignAppealReviewers(ctx sdk.Context, appeal *types.Appeal) error {
	if k.stakingKeeper == nil {
		return types.ErrStakingKeeperNotSet
	}

	// Get all moderators and filter by tier
	moderators := k.GetActiveModerators(ctx)

	var eligible []string
	for _, mod := range moderators {
		modAddr, err := sdk.AccAddressFromBech32(mod.Address)
		if err != nil {
			continue
		}

		tier := k.stakingKeeper.GetUserTierInt(ctx, modAddr)
		if tier >= appeal.RequiredTier {
			// Exclude original moderators who voted in the dispute
			if appeal.DisputeID > 0 {
				if dispute, found := k.GetDispute(ctx, appeal.DisputeID); found {
					wasOriginalVoter := false
					for _, vote := range dispute.Votes {
						if vote.Moderator == mod.Address {
							wasOriginalVoter = true
							break
						}
					}
					if wasOriginalVoter {
						continue
					}
				}
			}
			eligible = append(eligible, mod.Address)
		}
	}

	if len(eligible) < appeal.ReviewerCount {
		return fmt.Errorf("insufficient eligible reviewers: need %d, have %d", appeal.ReviewerCount, len(eligible))
	}

	// NEW: Use block hash for deterministic but unpredictable randomization (Issue #12)
	blockHash := ctx.HeaderHash()
	seed := binary.BigEndian.Uint64(blockHash[:8])

	// Fisher-Yates shuffle with seed
	shuffled := make([]string, len(eligible))
	copy(shuffled, eligible)

	rng := rand.New(rand.NewSource(int64(seed + appeal.ID)))
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	appeal.AssignedReviewers = shuffled[:appeal.ReviewerCount]

	k.Logger(ctx).Info("reviewers assigned with randomization",
		"appeal_id", appeal.ID,
		"seed", seed,
		"reviewers", appeal.AssignedReviewers)

	return nil
}

// VoteOnAppeal allows a reviewer to vote on an appeal
func (k Keeper) VoteOnAppeal(
	ctx sdk.Context,
	appealID uint64,
	reviewer string,
	upholdOriginal bool,
	newResolution types.DisputeResolution,
	reasoning string,
) error {
	appeal, found := k.GetAppeal(ctx, appealID)
	if !found {
		return types.ErrAppealNotFound
	}

	// Check appeal is being reviewed
	if appeal.Status != types.AppealStatusReviewing && appeal.Status != types.AppealStatusOpen {
		return types.ErrAppealAlreadyResolved
	}

	// Check deadline
	if ctx.BlockTime().After(appeal.DeadlineAt) {
		return types.ErrAppealDeadlinePassed
	}

	// Check hasn't already voted
	for _, vote := range appeal.Votes {
		if vote.Reviewer == reviewer {
			return types.ErrAppealReviewerAlreadyVoted
		}
	}

	// Get reviewer tier
	reviewerAddr, err := sdk.AccAddressFromBech32(reviewer)
	if err != nil {
		return types.ErrUnauthorized
	}

	tier := k.stakingKeeper.GetUserTierInt(ctx, reviewerAddr)
	if tier < appeal.RequiredTier {
		return types.ErrInsufficientReviewerTier
	}

	// NEW: Snapshot evidence state on first vote (Issue #5)
	if len(appeal.Votes) == 0 {
		appeal.EvidenceSnapshot = k.hashEvidenceState(appeal.NewEvidence)
		appeal.EvidenceLockedAt = ctx.BlockTime()
	}

	// Add vote
	vote := types.AppealVote{
		Reviewer:       reviewer,
		Tier:           tier,
		UpholdOriginal: upholdOriginal,
		NewResolution:  newResolution,
		Reasoning:      reasoning,
		VotedAt:        ctx.BlockTime(),
	}
	appeal.Votes = append(appeal.Votes, vote)

	// Check if we have enough votes
	if len(appeal.Votes) >= appeal.ReviewerCount {
		k.resolveAppeal(ctx, &appeal)
	}

	k.SetAppeal(ctx, appeal)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAppealVoted,
			sdk.NewAttribute(types.AttributeKeyAppealID, fmt.Sprintf("%d", appealID)),
			sdk.NewAttribute(types.AttributeKeyReviewer, reviewer),
			sdk.NewAttribute(types.AttributeKeyUpheld, fmt.Sprintf("%t", upholdOriginal)),
		),
	)

	return nil
}

// resolveAppeal resolves an appeal based on votes
func (k Keeper) resolveAppeal(ctx sdk.Context, appeal *types.Appeal) {
	// Count votes
	upholdCount := 0
	overturnCount := 0

	// Track proposed new resolutions
	resolutionVotes := make(map[types.DisputeResolution]int)

	for _, vote := range appeal.Votes {
		if vote.UpholdOriginal {
			upholdCount++
		} else {
			overturnCount++
			resolutionVotes[vote.NewResolution]++
		}
	}

	appeal.ResolvedAt = ctx.BlockTime()

	if upholdCount > overturnCount {
		// Original decision upheld
		appeal.Status = types.AppealStatusUpheld
		appeal.NewResolution = appeal.OriginalResolution

		// Reward original moderators
		if appeal.DisputeID > 0 {
			k.rewardOriginalModerators(ctx, appeal.DisputeID)
		}
	} else {
		// Original decision overturned
		appeal.Status = types.AppealStatusOverturned

		// Find most popular new resolution
		var maxVotes int
		for resolution, count := range resolutionVotes {
			if count > maxVotes {
				maxVotes = count
				appeal.NewResolution = resolution
			}
		}

		// Penalize original moderators (reputation first)
		if appeal.DisputeID > 0 {
			k.penalizeOriginalModerators(ctx, appeal.DisputeID)
		}

		// Execute new resolution
		if appeal.DisputeID > 0 {
			k.executeAppealResolution(ctx, appeal)
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAppealResolved,
			sdk.NewAttribute(types.AttributeKeyAppealID, fmt.Sprintf("%d", appeal.ID)),
			sdk.NewAttribute(types.AttributeKeyStatus, appeal.Status.String()),
			sdk.NewAttribute(types.AttributeKeyOverturn, fmt.Sprintf("%t", appeal.Status == types.AppealStatusOverturned)),
		),
	)
}

// rewardOriginalModerators rewards moderators whose decision was upheld
func (k Keeper) rewardOriginalModerators(ctx sdk.Context, disputeID uint64) {
	dispute, found := k.GetDispute(ctx, disputeID)
	if !found {
		return
	}

	for _, vote := range dispute.Votes {
		if vote.Vote == dispute.Resolution {
			// This moderator voted correctly
			k.rewardModeratorReputation(ctx, vote.Moderator, disputeID)

			// Update metrics
			metrics := k.GetOrCreateModeratorMetrics(ctx, vote.Moderator)
			metrics.RecordDecision(false, ctx.BlockTime()) // Not overturned
			k.SetModeratorMetrics(ctx, metrics)
		}
	}
}

// penalizeOriginalModerators penalizes moderators whose decision was overturned
func (k Keeper) penalizeOriginalModerators(ctx sdk.Context, disputeID uint64) {
	dispute, found := k.GetDispute(ctx, disputeID)
	if !found {
		return
	}

	for _, vote := range dispute.Votes {
		if vote.Vote == dispute.Resolution {
			// This moderator voted for the overturned resolution
			k.penalizeModeratorReputation(ctx, vote.Moderator, disputeID)

			// Update metrics - this is an overturn
			metrics := k.GetOrCreateModeratorMetrics(ctx, vote.Moderator)
			metrics.RecordDecision(true, ctx.BlockTime()) // Overturned
			k.SetModeratorMetrics(ctx, metrics)

			// Check for auto-blacklist
			if shouldBlacklist, reason := metrics.ShouldAutoBlacklist(); shouldBlacklist {
				k.autoBlacklistModerator(ctx, vote.Moderator, reason, &metrics)
			}
			k.SetModeratorMetrics(ctx, metrics)
		}
	}
}

// executeAppealResolution executes the new resolution after an appeal
func (k Keeper) executeAppealResolution(ctx sdk.Context, appeal *types.Appeal) {
	if appeal.DisputeID == 0 {
		return
	}

	dispute, found := k.GetDispute(ctx, appeal.DisputeID)
	if !found {
		return
	}

	escrow, found := k.GetEscrow(ctx, dispute.EscrowID)
	if !found {
		return
	}

	// Update dispute with new resolution
	dispute.Resolution = appeal.NewResolution
	dispute.Status = types.DisputeStatusFinal
	k.SetDispute(ctx, dispute)

	// Execute the new resolution
	switch appeal.NewResolution {
	case types.DisputeResolutionReleaseBuyer:
		k.ReleaseEscrow(ctx, dispute.EscrowID, escrow.Moderator)
	case types.DisputeResolutionReleaseSeller, types.DisputeResolutionRefund:
		k.RefundEscrow(ctx, dispute.EscrowID, escrow.Moderator)
	case types.DisputeResolutionSplit:
		k.splitEscrowFunds(ctx, escrow, &dispute)
	}

	// Update escrow status
	escrow.Status = types.EscrowStatusResolved
	k.SetEscrow(ctx, escrow)
}

// =============================================================================
// REPORT APPEALS
// =============================================================================

// SubmitReportAppeal submits an appeal for a dismissed report
func (k Keeper) SubmitReportAppeal(
	ctx sdk.Context,
	reportID uint64,
	appellant string,
	reason string,
) (uint64, error) {
	report, found := k.GetReport(ctx, reportID)
	if !found {
		return 0, types.ErrReportNotFound
	}

	// Only reporter can appeal
	if appellant != report.Reporter {
		return 0, types.ErrAppealerNotParticipant
	}

	// Can only appeal dismissed reports
	if report.Status != types.ReportStatusDismissed {
		return 0, types.ErrCannotAppeal
	}

	// Get existing appeals for this report
	existingAppeals := k.GetAppealsByReport(ctx, reportID)
	appealLevel := len(existingAppeals) + 1

	// Check max appeal level
	if appealLevel > 2 { // Reports can only be appealed twice
		return 0, types.ErrAppealLevelMaxed
	}

	// Create appeal
	appealID := k.GetNextAppealID(ctx)
	appeal := types.NewAppeal(
		appealID,
		0, // Not a dispute appeal
		reportID,
		appellant,
		reason,
		appealLevel,
		types.DisputeResolutionNone, // Not applicable for report appeals
		ctx.BlockTime(),
	)

	// Store appeal
	if err := k.SetAppeal(ctx, appeal); err != nil {
		return 0, err
	}

	// Create indexes
	k.IndexAppeal(ctx, appeal)

	// Update report status
	oldStatus := report.Status
	report.Status = types.ReportStatusAppealed
	k.SetReport(ctx, report)
	k.UpdateReportStatusIndex(ctx, report, oldStatus)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReportAppealed,
			sdk.NewAttribute(types.AttributeKeyReportID, fmt.Sprintf("%d", reportID)),
			sdk.NewAttribute(types.AttributeKeyAppealID, fmt.Sprintf("%d", appealID)),
			sdk.NewAttribute(types.AttributeKeyAppellant, appellant),
		),
	)

	return appealID, nil
}

// =============================================================================
// APPEAL ESCALATION
// =============================================================================

// EscalateAppeal escalates an appeal to a higher tier
func (k Keeper) EscalateAppeal(ctx sdk.Context, appealID uint64) error {
	appeal, found := k.GetAppeal(ctx, appealID)
	if !found {
		return types.ErrAppealNotFound
	}

	// Can only escalate appeals that weren't resolved or were upheld
	if appeal.Status == types.AppealStatusOverturned {
		return types.ErrCannotAppeal
	}

	// Check if can escalate
	if appeal.AppealLevel >= 3 {
		return types.ErrAppealLevelMaxed
	}

	// Create new escalated appeal
	newAppealID := k.GetNextAppealID(ctx)
	newAppeal := types.NewAppeal(
		newAppealID,
		appeal.DisputeID,
		appeal.ReportID,
		appeal.Appellant,
		fmt.Sprintf("Escalated from appeal %d: %s", appeal.ID, appeal.AppealReason),
		appeal.AppealLevel+1,
		appeal.OriginalResolution,
		ctx.BlockTime(),
	)

	// Copy evidence
	newAppeal.NewEvidence = append(newAppeal.NewEvidence, appeal.NewEvidence...)

	// Store new appeal
	if err := k.SetAppeal(ctx, newAppeal); err != nil {
		return err
	}

	// Mark old appeal as escalated
	appeal.Status = types.AppealStatusEscalated
	k.SetAppeal(ctx, appeal)

	// Create indexes
	k.IndexAppeal(ctx, newAppeal)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAppealEscalated,
			sdk.NewAttribute(types.AttributeKeyAppealID, fmt.Sprintf("%d", newAppealID)),
			sdk.NewAttribute("original_appeal_id", fmt.Sprintf("%d", appealID)),
			sdk.NewAttribute(types.AttributeKeyAppealLevel, fmt.Sprintf("%d", newAppeal.AppealLevel)),
		),
	)

	return nil
}

// =============================================================================
// APPEAL QUERIES
// =============================================================================

// GetAppealsByDispute returns all appeals for a dispute
func (k Keeper) GetAppealsByDispute(ctx sdk.Context, disputeID uint64) []types.Appeal {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetAppealByDisputePrefixKey(disputeID)
	iterator := prefix.NewStore(store, prefixKey).Iterator(nil, nil)
	defer iterator.Close()

	var appeals []types.Appeal
	for ; iterator.Valid(); iterator.Next() {
		appealID := sdk.BigEndianToUint64(iterator.Value())
		if appeal, found := k.GetAppeal(ctx, appealID); found {
			appeals = append(appeals, appeal)
		}
	}

	return appeals
}

// GetAppealsByReport returns all appeals for a report
func (k Keeper) GetAppealsByReport(ctx sdk.Context, reportID uint64) []types.Appeal {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetAppealByReportPrefixKey(reportID)
	iterator := prefix.NewStore(store, prefixKey).Iterator(nil, nil)
	defer iterator.Close()

	var appeals []types.Appeal
	for ; iterator.Valid(); iterator.Next() {
		appealID := sdk.BigEndianToUint64(iterator.Value())
		if appeal, found := k.GetAppeal(ctx, appealID); found {
			appeals = append(appeals, appeal)
		}
	}

	return appeals
}

// GetAllAppeals returns all appeals for genesis export
func (k Keeper) GetAllAppeals(ctx sdk.Context) []types.Appeal {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.AppealPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var appeals []types.Appeal
	for ; iterator.Valid(); iterator.Next() {
		var appeal types.Appeal
		if err := json.Unmarshal(iterator.Value(), &appeal); err != nil {
			continue
		}
		appeals = append(appeals, appeal)
	}

	return appeals
}

// =============================================================================
// EXPIRED APPEALS PROCESSING
// =============================================================================

// ProcessExpiredAppeals processes appeals that have passed their deadline
// Called from EndBlock
func (k Keeper) ProcessExpiredAppeals(ctx sdk.Context) {
	appeals := k.GetAllAppeals(ctx)

	for _, appeal := range appeals {
		// Skip already resolved appeals
		if appeal.Status != types.AppealStatusOpen && appeal.Status != types.AppealStatusReviewing {
			continue
		}

		if ctx.BlockTime().After(appeal.DeadlineAt) {
			// Auto-resolve based on current votes
			if len(appeal.Votes) > 0 {
				k.resolveAppeal(ctx, &appeal)
				k.SetAppeal(ctx, appeal)
			} else {
				// No votes - uphold original by default
				appeal.Status = types.AppealStatusUpheld
				appeal.NewResolution = appeal.OriginalResolution
				appeal.ResolvedAt = ctx.BlockTime()
				k.SetAppeal(ctx, appeal)

				k.Logger(ctx).Info("auto-upheld expired appeal with no votes",
					"appeal_id", appeal.ID,
				)
			}
		}
	}
}

// =============================================================================
// ADD EVIDENCE TO APPEAL
// =============================================================================

// AddAppealEvidence adds evidence to an appeal
func (k Keeper) AddAppealEvidence(
	ctx sdk.Context,
	appealID uint64,
	submitter string,
	hash string,
	description string,
) error {
	appeal, found := k.GetAppeal(ctx, appealID)
	if !found {
		return types.ErrAppealNotFound
	}

	// Only appellant can add evidence
	if submitter != appeal.Appellant {
		return types.ErrUnauthorized
	}

	// Check appeal is still open
	if appeal.Status != types.AppealStatusOpen && appeal.Status != types.AppealStatusReviewing {
		return types.ErrAppealAlreadyResolved
	}

	// NEW: Prevent evidence addition after voting has started (Issue #5)
	if len(appeal.Votes) > 0 {
		return types.ErrEvidenceLockedAfterVoting
	}

	// Add evidence
	evidence := types.Evidence{
		Submitter:   submitter,
		Hash:        hash,
		Description: description,
		SubmittedAt: ctx.BlockTime(),
	}
	appeal.NewEvidence = append(appeal.NewEvidence, evidence)

	k.SetAppeal(ctx, appeal)

	return nil
}

// hashEvidenceState creates a deterministic hash of evidence for snapshot (Issue #5)
func (k Keeper) hashEvidenceState(evidence []types.Evidence) string {
	if len(evidence) == 0 {
		return ""
	}

	// Create a stable representation of evidence
	var evidenceStr string
	for _, e := range evidence {
		evidenceStr += fmt.Sprintf("%s:%s:%s;", e.Submitter, e.Hash, e.SubmittedAt.String())
	}

	// Use simple string representation as hash (could use crypto/sha256 for production)
	return evidenceStr
}
