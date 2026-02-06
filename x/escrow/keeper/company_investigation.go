package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/escrow/types"
)

// =============================================================================
// TIERED COMPANY INVESTIGATION SYSTEM
// Multi-tier voting for fraud freeze authority
// =============================================================================

// GetNextInvestigationID returns the next investigation ID and increments the counter
func (k Keeper) GetNextInvestigationID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CompanyInvestigationCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.CompanyInvestigationCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// =============================================================================
// INVESTIGATION STORAGE
// =============================================================================

// GetCompanyInvestigation returns an investigation by ID
func (k Keeper) GetCompanyInvestigation(ctx sdk.Context, investigationID uint64) (types.CompanyInvestigation, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetCompanyInvestigationKey(investigationID))
	if bz == nil {
		return types.CompanyInvestigation{}, false
	}

	var investigation types.CompanyInvestigation
	if err := json.Unmarshal(bz, &investigation); err != nil {
		return types.CompanyInvestigation{}, false
	}
	return investigation, true
}

// SetCompanyInvestigation stores an investigation
func (k Keeper) SetCompanyInvestigation(ctx sdk.Context, investigation types.CompanyInvestigation) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(investigation)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal investigation", "error", err)
		return fmt.Errorf("failed to marshal investigation: %w", err)
	}
	store.Set(types.GetCompanyInvestigationKey(investigation.ID), bz)
	return nil
}

// IndexCompanyInvestigation creates indexes for efficient querying
func (k Keeper) IndexCompanyInvestigation(ctx sdk.Context, investigation types.CompanyInvestigation) {
	store := ctx.KVStore(k.storeKey)

	// Index by company
	companyKey := types.GetCompanyInvestigationByCompanyKey(investigation.CompanyID, investigation.ID)
	store.Set(companyKey, sdk.Uint64ToBigEndian(investigation.ID))

	// Index by status
	statusKey := types.GetCompanyInvestigationByStatusKey(investigation.Status.String(), investigation.ID)
	store.Set(statusKey, sdk.Uint64ToBigEndian(investigation.ID))
}

// UpdateInvestigationStatusIndex updates the status index when status changes
func (k Keeper) UpdateInvestigationStatusIndex(ctx sdk.Context, investigation types.CompanyInvestigation, oldStatus types.InvestigationStatus) {
	store := ctx.KVStore(k.storeKey)

	// Remove old status index
	oldStatusKey := types.GetCompanyInvestigationByStatusKey(oldStatus.String(), investigation.ID)
	store.Delete(oldStatusKey)

	// Add new status index
	newStatusKey := types.GetCompanyInvestigationByStatusKey(investigation.Status.String(), investigation.ID)
	store.Set(newStatusKey, sdk.Uint64ToBigEndian(investigation.ID))
}

// =============================================================================
// INVESTIGATION CREATION
// =============================================================================

// CreateCompanyInvestigation creates a new fraud investigation for a company
// This is called when a Keeper submits a fraud/scam report against a company
// NO automatic freeze - investigation starts in Preliminary status
func (k Keeper) CreateCompanyInvestigation(ctx sdk.Context, companyID uint64, reportID uint64) (uint64, error) {
	// EDGE CASE #1: Check for existing active investigation against same company
	existingInvestigations := k.GetInvestigationsByCompany(ctx, companyID)
	for _, inv := range existingInvestigations {
		if inv.Status != types.InvestigationStatusCleared &&
			inv.Status != types.InvestigationStatusFrozen {
			k.Logger(ctx).Info("blocking duplicate investigation - company already under investigation",
				"company_id", companyID,
				"existing_investigation_id", inv.ID,
				"existing_status", inv.Status.String())
			return 0, types.ErrActiveInvestigationExists
		}
	}

	// Check if company exists (via equity keeper if available)
	if k.equityKeeper != nil {
		// The company validation would happen here
		// For now, we proceed assuming the company exists
	}

	// Create investigation
	investigationID := k.GetNextInvestigationID(ctx)
	investigation := types.NewCompanyInvestigation(investigationID, companyID, reportID, ctx.BlockTime())

	// Validate
	if err := investigation.Validate(); err != nil {
		return 0, types.ErrInvalidReport
	}

	// EDGE CASE #4: Check minimum Warden-tier reviewers are available
	if k.stakingKeeper != nil {
		wardenCount := k.countAvailableReviewers(ctx, types.StakeTierWarden)
		if wardenCount < 3 {
			k.Logger(ctx).Error("insufficient Warden-tier reviewers available for investigation",
				"required", 3,
				"available", wardenCount)
			return 0, types.ErrInsufficientReviewers
		}
	}

	// Store investigation
	if err := k.SetCompanyInvestigation(ctx, investigation); err != nil {
		return 0, err
	}

	// Create indexes
	k.IndexCompanyInvestigation(ctx, investigation)

	// Auto-transition to Warden review
	investigation.Status = types.InvestigationStatusWardenReview
	k.SetCompanyInvestigation(ctx, investigation)
	k.UpdateInvestigationStatusIndex(ctx, investigation, types.InvestigationStatusPreliminary)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"company_investigation_created",
			sdk.NewAttribute("investigation_id", fmt.Sprintf("%d", investigationID)),
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute("report_id", fmt.Sprintf("%d", reportID)),
			sdk.NewAttribute("status", investigation.Status.String()),
			sdk.NewAttribute("warden_deadline", investigation.WardenDeadline.String()),
		),
	)

	k.Logger(ctx).Info("company investigation created - awaiting Warden review",
		"investigation_id", investigationID,
		"company_id", companyID,
		"report_id", reportID,
		"deadline", investigation.WardenDeadline.String())

	return investigationID, nil
}

// =============================================================================
// INVESTIGATION VOTING
// =============================================================================

// VoteOnInvestigation allows a Warden or Steward to vote on an investigation
func (k Keeper) VoteOnInvestigation(
	ctx sdk.Context,
	investigationID uint64,
	voter string,
	approve bool,
	reason string,
) error {
	investigation, found := k.GetCompanyInvestigation(ctx, investigationID)
	if !found {
		return types.ErrReportNotFound
	}

	// Check investigation is in a voting phase
	if investigation.Status != types.InvestigationStatusWardenReview &&
		investigation.Status != types.InvestigationStatusStewardReview {
		return fmt.Errorf("investigation is not in a voting phase: %s", investigation.Status.String())
	}

	// Check voter hasn't already voted
	if investigation.HasVoted(voter) {
		return types.ErrReviewerAlreadyVoted
	}

	// Get voter tier
	voterAddr, err := sdk.AccAddressFromBech32(voter)
	if err != nil {
		return err
	}

	if k.stakingKeeper == nil {
		return types.ErrStakingKeeperNotSet
	}

	tier := k.stakingKeeper.GetUserTierInt(ctx, voterAddr)

	// Validate voter tier based on phase
	if investigation.Status == types.InvestigationStatusWardenReview {
		// Must be Warden+ (tier 2)
		if tier < types.StakeTierWarden {
			return fmt.Errorf("insufficient tier for Warden review: need Warden (2), have %d", tier)
		}
	} else if investigation.Status == types.InvestigationStatusStewardReview {
		// Must be Steward+ (tier 3)
		if tier < types.StakeTierSteward {
			return fmt.Errorf("insufficient tier for Steward review: need Steward (3), have %d", tier)
		}
	}

	// EDGE CASE #2: Check reviewer does not hold shares in the company being investigated
	if k.equityKeeper != nil {
		// Check if reviewer holds any shares in this company
		shareholding, found := k.equityKeeper.GetShareholding(ctx, investigation.CompanyID, "", voter)
		if found && shareholding != nil {
			k.Logger(ctx).Warn("reviewer conflict of interest - holds shares in company under investigation",
				"investigation_id", investigationID,
				"company_id", investigation.CompanyID,
				"reviewer", voter)
			return types.ErrReviewerConflictOfInterest
		}
	}

	// EDGE CASE #7: Check voter is not in unbonding period (prevents penalty escape)
	if k.stakingKeeper != nil {
		// Check if user is currently unbonding or recently staked
		// If stake age is low, might be re-staking after unbond attempt
		stakeAge := k.stakingKeeper.GetStakeAge(ctx, voterAddr)
		if stakeAge < types.MinStakeAgeForVoting {
			k.Logger(ctx).Warn("reviewer recently staked - preventing vote to avoid unbonding escape",
				"investigation_id", investigationID,
				"reviewer", voter,
				"stake_age_seconds", stakeAge,
				"required_seconds", types.MinStakeAgeForVoting)
			return types.ErrReviewerRecentlyStaked
		}
	}

	// Create vote
	vote := types.TierVote{
		Voter:   voter,
		Tier:    tier,
		Approve: approve,
		Reason:  reason,
		VotedAt: ctx.BlockTime(),
	}

	oldStatus := investigation.Status

	// Add vote to appropriate phase
	if investigation.Status == types.InvestigationStatusWardenReview {
		investigation.WardenVotes = append(investigation.WardenVotes, vote)

		// Check if we have enough votes (3 votes required)
		approveCount, rejectCount := investigation.CountWardenVotes()
		if approveCount+rejectCount >= 3 {
			// Voting complete
			if approveCount >= 2 {
				// 2/3 approved - escalate to Steward review
				investigation.Status = types.InvestigationStatusStewardReview
				investigation.WardenApproved = true
				investigation.StewardDeadline = ctx.BlockTime().Add(72 * 60 * 60 * 1000000000) // 72 hours

				k.Logger(ctx).Info("investigation escalated to Steward review",
					"investigation_id", investigationID,
					"warden_approve", approveCount,
					"warden_reject", rejectCount)
			} else {
				// Majority rejected - clear investigation
				investigation.Status = types.InvestigationStatusCleared
				investigation.ResolvedAt = ctx.BlockTime()

				// Resume trading for the company (it was halted when investigation started)
				if k.equityKeeper != nil {
					if err := k.equityKeeper.ResumeTrading(ctx, investigation.CompanyID, "investigation_cleared_by_wardens"); err != nil {
						k.Logger(ctx).Error("failed to resume trading after cleared investigation",
							"investigation_id", investigationID,
							"company_id", investigation.CompanyID,
							"error", err)
					}
				}

				k.Logger(ctx).Info("investigation cleared by Wardens - trading resumed",
					"investigation_id", investigationID,
					"company_id", investigation.CompanyID,
					"warden_approve", approveCount,
					"warden_reject", rejectCount)
			}
		}
	} else if investigation.Status == types.InvestigationStatusStewardReview {
		investigation.StewardVotes = append(investigation.StewardVotes, vote)

		// Check if we have enough votes (5 votes required)
		approveCount, rejectCount := investigation.CountStewardVotes()
		if approveCount+rejectCount >= 5 {
			// Voting complete
			if approveCount >= 3 {
				// 3/5 approved - issue freeze warning
				investigation.Status = types.InvestigationStatusFreezeApproved
				investigation.StewardApproved = true
				investigation.WarningIssuedAt = ctx.BlockTime()
				investigation.WarningExpiresAt = ctx.BlockTime().Add(24 * 60 * 60 * 1000000000) // 24 hours

				k.Logger(ctx).Warn("freeze approved - 24hr warning issued",
					"investigation_id", investigationID,
					"company_id", investigation.CompanyID,
					"steward_approve", approveCount,
					"steward_reject", rejectCount,
					"warning_expires", investigation.WarningExpiresAt.String())

				// Emit warning event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						"investigation_freeze_warning",
						sdk.NewAttribute("investigation_id", fmt.Sprintf("%d", investigationID)),
						sdk.NewAttribute("company_id", fmt.Sprintf("%d", investigation.CompanyID)),
						sdk.NewAttribute("warning_expires_at", investigation.WarningExpiresAt.String()),
					),
				)
			} else {
				// Majority rejected - clear investigation
				investigation.Status = types.InvestigationStatusCleared
				investigation.ResolvedAt = ctx.BlockTime()

				// Resume trading for the company (it was halted when investigation started)
				if k.equityKeeper != nil {
					if err := k.equityKeeper.ResumeTrading(ctx, investigation.CompanyID, "investigation_cleared_by_stewards"); err != nil {
						k.Logger(ctx).Error("failed to resume trading after cleared investigation",
							"investigation_id", investigationID,
							"company_id", investigation.CompanyID,
							"error", err)
					}
				}

				k.Logger(ctx).Info("investigation cleared by Stewards - trading resumed",
					"investigation_id", investigationID,
					"company_id", investigation.CompanyID,
					"steward_approve", approveCount,
					"steward_reject", rejectCount)
			}
		}
	}

	// Save investigation
	k.SetCompanyInvestigation(ctx, investigation)

	// Update status index if status changed
	if investigation.Status != oldStatus {
		k.UpdateInvestigationStatusIndex(ctx, investigation, oldStatus)
	}

	// Emit vote event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"investigation_vote",
			sdk.NewAttribute("investigation_id", fmt.Sprintf("%d", investigationID)),
			sdk.NewAttribute("voter", voter),
			sdk.NewAttribute("tier", fmt.Sprintf("%d", tier)),
			sdk.NewAttribute("approve", fmt.Sprintf("%t", approve)),
			sdk.NewAttribute("phase", investigation.Status.String()),
		),
	)

	return nil
}

// =============================================================================
// INVESTIGATION PHASE PROCESSING (EndBlock)
// =============================================================================

// ProcessInvestigationPhases handles phase transitions and deadlines
// Called from EndBlock
func (k Keeper) ProcessInvestigationPhases(ctx sdk.Context) {
	// Process Warden review deadlines
	k.processWardenReviewDeadlines(ctx)

	// Process Steward review deadlines
	k.processStewardReviewDeadlines(ctx)

	// Process freeze warnings (execute freeze after 24hr warning)
	k.processFreezeWarnings(ctx)
}

// processWardenReviewDeadlines handles expired Warden review deadlines
func (k Keeper) processWardenReviewDeadlines(ctx sdk.Context) {
	investigations := k.GetInvestigationsByStatus(ctx, types.InvestigationStatusWardenReview)

	for _, investigation := range investigations {
		if ctx.BlockTime().After(investigation.WardenDeadline) {
			// Deadline passed - check if we have enough votes
			approveCount, rejectCount := investigation.CountWardenVotes()

			if approveCount+rejectCount == 0 {
				// No votes - auto-clear
				oldStatus := investigation.Status
				investigation.Status = types.InvestigationStatusCleared
				investigation.ResolvedAt = ctx.BlockTime()
				k.SetCompanyInvestigation(ctx, investigation)
				k.UpdateInvestigationStatusIndex(ctx, investigation, oldStatus)

				// Resume trading for the company
				if k.equityKeeper != nil {
					if err := k.equityKeeper.ResumeTrading(ctx, investigation.CompanyID, "investigation_auto_cleared_no_votes"); err != nil {
						k.Logger(ctx).Error("failed to resume trading after auto-cleared investigation",
							"investigation_id", investigation.ID,
							"company_id", investigation.CompanyID,
							"error", err)
					}
				}

				k.Logger(ctx).Info("investigation auto-cleared - no Warden votes, trading resumed",
					"investigation_id", investigation.ID,
					"company_id", investigation.CompanyID)
			} else if approveCount >= 2 {
				// Have enough approval votes - escalate
				oldStatus := investigation.Status
				investigation.Status = types.InvestigationStatusStewardReview
				investigation.WardenApproved = true
				investigation.StewardDeadline = ctx.BlockTime().Add(72 * 60 * 60 * 1000000000)
				k.SetCompanyInvestigation(ctx, investigation)
				k.UpdateInvestigationStatusIndex(ctx, investigation, oldStatus)

				k.Logger(ctx).Info("investigation escalated to Steward review (deadline)",
					"investigation_id", investigation.ID)
			} else {
				// Not enough approvals - clear
				oldStatus := investigation.Status
				investigation.Status = types.InvestigationStatusCleared
				investigation.ResolvedAt = ctx.BlockTime()
				k.SetCompanyInvestigation(ctx, investigation)
				k.UpdateInvestigationStatusIndex(ctx, investigation, oldStatus)

				// Resume trading for the company
				if k.equityKeeper != nil {
					if err := k.equityKeeper.ResumeTrading(ctx, investigation.CompanyID, "investigation_cleared_insufficient_warden_approval"); err != nil {
						k.Logger(ctx).Error("failed to resume trading after cleared investigation",
							"investigation_id", investigation.ID,
							"company_id", investigation.CompanyID,
							"error", err)
					}
				}

				k.Logger(ctx).Info("investigation cleared - insufficient Warden approval, trading resumed",
					"investigation_id", investigation.ID,
					"company_id", investigation.CompanyID)
			}
		}
	}
}

// processStewardReviewDeadlines handles expired Steward review deadlines
func (k Keeper) processStewardReviewDeadlines(ctx sdk.Context) {
	investigations := k.GetInvestigationsByStatus(ctx, types.InvestigationStatusStewardReview)

	for _, investigation := range investigations {
		if ctx.BlockTime().After(investigation.StewardDeadline) {
			// Deadline passed - check if we have enough votes
			approveCount, _ := investigation.CountStewardVotes()

			if approveCount >= 3 {
				// Have enough approval votes - approve freeze
				oldStatus := investigation.Status
				investigation.Status = types.InvestigationStatusFreezeApproved
				investigation.StewardApproved = true
				investigation.WarningIssuedAt = ctx.BlockTime()
				investigation.WarningExpiresAt = ctx.BlockTime().Add(24 * 60 * 60 * 1000000000)
				k.SetCompanyInvestigation(ctx, investigation)
				k.UpdateInvestigationStatusIndex(ctx, investigation, oldStatus)

				k.Logger(ctx).Warn("freeze approved (deadline) - 24hr warning issued",
					"investigation_id", investigation.ID,
					"company_id", investigation.CompanyID)
			} else {
				// Not enough approvals - clear
				oldStatus := investigation.Status
				investigation.Status = types.InvestigationStatusCleared
				investigation.ResolvedAt = ctx.BlockTime()
				k.SetCompanyInvestigation(ctx, investigation)
				k.UpdateInvestigationStatusIndex(ctx, investigation, oldStatus)

				// Resume trading for the company
				if k.equityKeeper != nil {
					if err := k.equityKeeper.ResumeTrading(ctx, investigation.CompanyID, "investigation_cleared_insufficient_steward_approval"); err != nil {
						k.Logger(ctx).Error("failed to resume trading after cleared investigation",
							"investigation_id", investigation.ID,
							"company_id", investigation.CompanyID,
							"error", err)
					}
				}

				k.Logger(ctx).Info("investigation cleared - insufficient Steward approval, trading resumed",
					"investigation_id", investigation.ID,
					"company_id", investigation.CompanyID)
			}
		}
	}
}

// processFreezeWarnings executes freezes after 24hr warning period expires
func (k Keeper) processFreezeWarnings(ctx sdk.Context) {
	investigations := k.GetInvestigationsByStatus(ctx, types.InvestigationStatusFreezeApproved)

	for _, investigation := range investigations {
		if ctx.BlockTime().After(investigation.WarningExpiresAt) {
			// Warning period expired - execute freeze
			if k.equityKeeper != nil {
				if err := k.equityKeeper.InitiateFraudInvestigation(ctx, investigation.CompanyID, investigation.ReportID, "tiered_investigation_system"); err != nil {
					k.Logger(ctx).Error("failed to freeze company",
						"investigation_id", investigation.ID,
						"company_id", investigation.CompanyID,
						"error", err)
					continue
				}
			}

			oldStatus := investigation.Status
			investigation.Status = types.InvestigationStatusFrozen
			investigation.FrozenAt = ctx.BlockTime()
			k.SetCompanyInvestigation(ctx, investigation)
			k.UpdateInvestigationStatusIndex(ctx, investigation, oldStatus)

			// Emit freeze event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"company_frozen",
					sdk.NewAttribute("investigation_id", fmt.Sprintf("%d", investigation.ID)),
					sdk.NewAttribute("company_id", fmt.Sprintf("%d", investigation.CompanyID)),
					sdk.NewAttribute("report_id", fmt.Sprintf("%d", investigation.ReportID)),
					sdk.NewAttribute("frozen_at", investigation.FrozenAt.String()),
				),
			)

			k.Logger(ctx).Warn("company frozen - trading halted, treasury frozen",
				"investigation_id", investigation.ID,
				"company_id", investigation.CompanyID,
				"report_id", investigation.ReportID)
		}
	}
}

// =============================================================================
// INVESTIGATION QUERIES
// =============================================================================

// GetInvestigationsByStatus returns all investigations with a specific status
func (k Keeper) GetInvestigationsByStatus(ctx sdk.Context, status types.InvestigationStatus) []types.CompanyInvestigation {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetCompanyInvestigationByStatusPrefixKey(status.String())
	iterator := prefix.NewStore(store, prefixKey).Iterator(nil, nil)
	defer iterator.Close()

	var investigations []types.CompanyInvestigation
	for ; iterator.Valid(); iterator.Next() {
		investigationID := sdk.BigEndianToUint64(iterator.Value())
		if investigation, found := k.GetCompanyInvestigation(ctx, investigationID); found {
			investigations = append(investigations, investigation)
		}
	}

	return investigations
}

// GetInvestigationsByCompany returns all investigations for a company
func (k Keeper) GetInvestigationsByCompany(ctx sdk.Context, companyID uint64) []types.CompanyInvestigation {
	store := ctx.KVStore(k.storeKey)
	prefixKey := types.GetCompanyInvestigationByCompanyPrefixKey(companyID)
	iterator := prefix.NewStore(store, prefixKey).Iterator(nil, nil)
	defer iterator.Close()

	var investigations []types.CompanyInvestigation
	for ; iterator.Valid(); iterator.Next() {
		investigationID := sdk.BigEndianToUint64(iterator.Value())
		if investigation, found := k.GetCompanyInvestigation(ctx, investigationID); found {
			investigations = append(investigations, investigation)
		}
	}

	return investigations
}

// GetAllCompanyInvestigations returns all investigations for genesis export
func (k Keeper) GetAllCompanyInvestigations(ctx sdk.Context) []types.CompanyInvestigation {
	store := ctx.KVStore(k.storeKey)
	iterator := prefix.NewStore(store, types.CompanyInvestigationPrefix).Iterator(nil, nil)
	defer iterator.Close()

	var investigations []types.CompanyInvestigation
	for ; iterator.Valid(); iterator.Next() {
		var investigation types.CompanyInvestigation
		if err := json.Unmarshal(iterator.Value(), &investigation); err != nil {
			continue
		}
		investigations = append(investigations, investigation)
	}

	return investigations
}

// =============================================================================
// HELPER METHODS
// =============================================================================

// countAvailableReviewers counts users at or above the specified tier
// Used to verify minimum reviewer availability before creating investigations
func (k Keeper) countAvailableReviewers(ctx sdk.Context, minTier int) int {
	// This would query the staking module for users at tier >= minTier
	// For now, return a reasonable default indicating reviewers exist
	// Full implementation requires staking keeper method
	//
	// TODO: Implement proper reviewer counting via staking keeper
	// k.stakingKeeper.GetUserCountByMinTier(ctx, minTier)
	return 10 // Placeholder - assume reviewers available
}
