package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// =============================================================================
// Freeze Warning Storage
// =============================================================================

// GetNextFreezeWarningID returns the next warning ID and increments the counter
func (k Keeper) GetNextFreezeWarningID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.FreezeWarningCounterKey)

	var counter uint64 = 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	store.Set(types.FreezeWarningCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// SetFreezeWarning stores a freeze warning
func (k Keeper) SetFreezeWarning(ctx sdk.Context, warning types.FreezeWarning) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetFreezeWarningKey(warning.ID)
	bz, err := json.Marshal(warning)
	if err != nil {
		return fmt.Errorf("failed to marshal freeze warning: %w", err)
	}
	store.Set(key, bz)

	// Index by company ID
	indexKey := types.GetFreezeWarningByCompanyKey(warning.CompanyID)
	store.Set(indexKey, sdk.Uint64ToBigEndian(warning.ID))
	return nil
}

// GetFreezeWarning retrieves a freeze warning by ID
func (k Keeper) GetFreezeWarning(ctx sdk.Context, warningID uint64) (types.FreezeWarning, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetFreezeWarningKey(warningID)
	bz := store.Get(key)
	if bz == nil {
		return types.FreezeWarning{}, false
	}

	var warning types.FreezeWarning
	if err := json.Unmarshal(bz, &warning); err != nil {
		return types.FreezeWarning{}, false
	}
	return warning, true
}

// GetFreezeWarningByCompany retrieves a freeze warning for a specific company
func (k Keeper) GetFreezeWarningByCompany(ctx sdk.Context, companyID uint64) (types.FreezeWarning, bool) {
	store := ctx.KVStore(k.storeKey)
	indexKey := types.GetFreezeWarningByCompanyKey(companyID)
	bz := store.Get(indexKey)
	if bz == nil {
		return types.FreezeWarning{}, false
	}

	warningID := sdk.BigEndianToUint64(bz)
	return k.GetFreezeWarning(ctx, warningID)
}

// DeleteFreezeWarning removes a freeze warning
func (k Keeper) DeleteFreezeWarning(ctx sdk.Context, warning types.FreezeWarning) {
	store := ctx.KVStore(k.storeKey)

	// Delete main record
	key := types.GetFreezeWarningKey(warning.ID)
	store.Delete(key)

	// Delete company index
	indexKey := types.GetFreezeWarningByCompanyKey(warning.CompanyID)
	store.Delete(indexKey)
}

// GetAllPendingFreezeWarnings returns all warnings in pending status
func (k Keeper) GetAllPendingFreezeWarnings(ctx sdk.Context) []types.FreezeWarning {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.FreezeWarningPrefix)
	defer iterator.Close()

	var warnings []types.FreezeWarning
	for ; iterator.Valid(); iterator.Next() {
		var warning types.FreezeWarning
		if err := json.Unmarshal(iterator.Value(), &warning); err != nil {
			k.Logger(ctx).Error("failed to unmarshal freeze warning", "error", err)
			continue
		}
		if warning.Status == types.FreezeWarningStatusPending || warning.Status == types.FreezeWarningStatusResponded {
			warnings = append(warnings, warning)
		}
	}
	return warnings
}

// =============================================================================
// Freeze Warning Operations
// =============================================================================

// IssueFreezeWarning issues a 24-hour warning before freeze
// Called by Stewards when they approve a freeze recommendation
func (k Keeper) IssueFreezeWarning(ctx sdk.Context, companyID, investigationID, reportID uint64) (uint64, error) {
	// Check if company exists
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return 0, types.ErrCompanyNotFound
	}

	// Check if there's already an active warning for this company
	if existingWarning, found := k.GetFreezeWarningByCompany(ctx, companyID); found {
		if existingWarning.Status == types.FreezeWarningStatusPending ||
		   existingWarning.Status == types.FreezeWarningStatusResponded {
			return 0, fmt.Errorf("company already has active freeze warning")
		}
	}

	// Create freeze warning with 24-hour deadline
	warningID := k.GetNextFreezeWarningID(ctx)
	warning := types.FreezeWarning{
		ID:               warningID,
		CompanyID:        companyID,
		InvestigationID:  investigationID,
		ReportID:         reportID,
		Status:           types.FreezeWarningStatusPending,
		WarningIssuedAt:  ctx.BlockTime(),
		WarningExpiresAt: ctx.BlockTime().Add(24 * time.Hour),
		CompanyResponded: false,
		FreezeExecuted:   false,
		EscalatedToArchon: false,
	}

	// Validate and store
	if err := warning.Validate(); err != nil {
		return 0, err
	}
	if err := k.SetFreezeWarning(ctx, warning); err != nil {
		return 0, err
	}

	// SECURITY: Impose strict treasury withdrawal limits during warning period
	// This prevents founders from draining treasury before freeze
	// Limits tightened to reduce front-running risk
	treasuryBalance := k.getTreasuryBalance(ctx, companyID)
	if treasuryBalance.IsPositive() {
		limit := types.TreasuryWithdrawalLimit{
			CompanyID:      companyID,
			MaxWithdrawal:  treasuryBalance.Quo(math.NewInt(20)), // 5% max per withdrawal (tightened from 10%)
			DailyLimit:     treasuryBalance.Quo(math.NewInt(10)), // 10% daily limit (tightened from 20%)
			WithdrawnToday: math.ZeroInt(),
			LastResetDate:  ctx.BlockTime(),
			WarningID:      warningID,
			CreatedAt:      ctx.BlockTime(),
		}
		if err := k.SetTreasuryWithdrawalLimit(ctx, limit); err != nil {
			k.Logger(ctx).Error("failed to set treasury withdrawal limit", "error", err)
		}
	}

	// Emit warning event (company should monitor this)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFreezeWarningIssued,
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute("warning_id", fmt.Sprintf("%d", warningID)),
			sdk.NewAttribute("report_id", fmt.Sprintf("%d", reportID)),
			sdk.NewAttribute("warning_expires_at", warning.WarningExpiresAt.String()),
			sdk.NewAttribute("founder", company.Founder),
		),
	)

	// Emit shareholder notification event
	// Indexers can use this to notify all shareholders of this company
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeShareholderAlert,
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyAlertType, "fraud_investigation_warning"),
			sdk.NewAttribute(types.AttributeKeyWarningID, fmt.Sprintf("%d", warningID)),
			sdk.NewAttribute(types.AttributeKeyExpiresAt, warning.WarningExpiresAt.String()),
			sdk.NewAttribute(types.AttributeKeyMessage, "Your holdings in this company are under fraud investigation. Trading may be halted."),
		),
	)

	k.Logger(ctx).Info("freeze warning issued",
		"company_id", companyID,
		"symbol", company.Symbol,
		"warning_id", warningID,
		"expires_at", warning.WarningExpiresAt,
	)

	return warningID, nil
}

// RespondToFreezeWarning allows company founder to respond with counter-evidence
func (k Keeper) RespondToFreezeWarning(
	ctx sdk.Context,
	warningID uint64,
	responder string,
	evidence []types.Evidence,
	response string,
) error {
	warning, found := k.GetFreezeWarning(ctx, warningID)
	if !found {
		return fmt.Errorf("freeze warning not found")
	}

	// Check if warning is still pending
	if warning.Status != types.FreezeWarningStatusPending {
		return fmt.Errorf("warning is no longer pending (status: %s)", warning.Status.String())
	}

	// Check if warning has expired
	if warning.IsExpired(ctx.BlockTime()) {
		return fmt.Errorf("warning period has expired")
	}

	// Validate responder is company founder OR authorized responder (Issue #11)
	company, found := k.getCompany(ctx, warning.CompanyID)
	if !found {
		return types.ErrCompanyNotFound
	}
	if !k.isAuthorizedResponder(ctx, company, responder) {
		return types.ErrUnauthorized
	}

	// Check if already responded
	if warning.CompanyResponded {
		return fmt.Errorf("company has already responded to this warning")
	}

	// Record response and evidence
	warning.CompanyResponded = true
	warning.ResponseSubmittedAt = ctx.BlockTime()
	warning.CounterEvidence = evidence
	warning.ResponseText = response
	warning.Status = types.FreezeWarningStatusResponded

	if err := k.SetFreezeWarning(ctx, warning); err != nil {
		return err
	}

	// Emit response event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFreezeWarningResponse,
			sdk.NewAttribute("warning_id", fmt.Sprintf("%d", warningID)),
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", warning.CompanyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute("responder", responder),
			sdk.NewAttribute("evidence_count", fmt.Sprintf("%d", len(evidence))),
		),
	)

	k.Logger(ctx).Info("company responded to freeze warning",
		"warning_id", warningID,
		"company_id", warning.CompanyID,
		"responder", responder,
		"evidence_count", len(evidence),
	)

	return nil
}

// ProcessExpiredFreezeWarnings processes warnings that have expired
// Called from EndBlock
func (k Keeper) ProcessExpiredFreezeWarnings(ctx sdk.Context) {
	warnings := k.GetAllPendingFreezeWarnings(ctx)

	for _, warning := range warnings {
		if warning.IsExpired(ctx.BlockTime()) {
			k.handleExpiredWarning(ctx, warning)
		}
	}
}

// handleExpiredWarning handles an expired freeze warning
func (k Keeper) handleExpiredWarning(ctx sdk.Context, warning types.FreezeWarning) {
	if !warning.CompanyResponded {
		// No response → Execute freeze immediately
		k.Logger(ctx).Info("executing freeze after warning period expired with no response",
			"warning_id", warning.ID,
			"company_id", warning.CompanyID,
		)

		if err := k.ExecuteFreezeAfterWarning(ctx, warning.ID); err != nil {
			k.Logger(ctx).Error("failed to execute freeze after warning",
				"warning_id", warning.ID,
				"error", err,
			)
		}
	} else {
		// Company responded → Escalate to Archon review
		k.Logger(ctx).Info("escalating freeze warning to Archon review",
			"warning_id", warning.ID,
			"company_id", warning.CompanyID,
		)

		if err := k.EscalateToArchonReview(ctx, warning.ID); err != nil {
			k.Logger(ctx).Error("failed to escalate warning to Archon",
				"warning_id", warning.ID,
				"error", err,
			)
		}
	}
}

// ExecuteFreezeAfterWarning performs the actual freeze after warning period
func (k Keeper) ExecuteFreezeAfterWarning(ctx sdk.Context, warningID uint64) error {
	warning, found := k.GetFreezeWarning(ctx, warningID)
	if !found {
		return fmt.Errorf("warning not found")
	}

	company, found := k.getCompany(ctx, warning.CompanyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Halt trading (if not already halted)
	if !k.IsTradingHalted(ctx, warning.CompanyID) {
		if err := k.HaltTrading(ctx, warning.CompanyID, "freeze warning expired", "freeze_warning_system", warning.ReportID); err != nil {
			k.Logger(ctx).Error("failed to halt trading",
				"company_id", warning.CompanyID,
				"error", err,
			)
		}
	}

	// Freeze treasury
	// Get company treasury balance
	companyAddr := sdk.MustAccAddressFromBech32(company.Founder) // Using founder as treasury for now
	treasuryBalance := k.bankKeeper.GetBalance(ctx, companyAddr, "uhodl")

	if treasuryBalance.Amount.IsPositive() {
		if err := k.FreezeTreasury(ctx, warning.CompanyID, treasuryBalance.Amount, "uhodl", "freeze warning expired", warning.ReportID); err != nil {
			k.Logger(ctx).Error("failed to freeze treasury",
				"company_id", warning.CompanyID,
				"error", err,
			)
		}
	}

	// Update warning status
	warning.Status = types.FreezeWarningStatusExecuted
	warning.FreezeExecuted = true
	warning.ResolvedAt = ctx.BlockTime()
	warning.Resolution = "Freeze executed after warning period expired"
	if err := k.SetFreezeWarning(ctx, warning); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFreezeExecuted,
			sdk.NewAttribute("warning_id", fmt.Sprintf("%d", warningID)),
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", warning.CompanyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute(types.AttributeKeyFrozenAmount, treasuryBalance.Amount.String()),
		),
	)

	k.Logger(ctx).Info("freeze executed after warning period",
		"warning_id", warningID,
		"company_id", warning.CompanyID,
		"symbol", company.Symbol,
	)

	return nil
}

// EscalateToArchonReview escalates a freeze warning with company response to Archon review
func (k Keeper) EscalateToArchonReview(ctx sdk.Context, warningID uint64) error {
	warning, found := k.GetFreezeWarning(ctx, warningID)
	if !found {
		return fmt.Errorf("warning not found")
	}

	company, found := k.getCompany(ctx, warning.CompanyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Update warning status
	warning.Status = types.FreezeWarningStatusEscalated
	warning.EscalatedToArchon = true
	warning.ResolvedAt = ctx.BlockTime()
	warning.Resolution = fmt.Sprintf("Escalated to Archon review - company submitted %d pieces of counter-evidence", len(warning.CounterEvidence))
	if err := k.SetFreezeWarning(ctx, warning); err != nil {
		return err
	}

	// Emit escalation event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFreezeEscalated,
			sdk.NewAttribute("warning_id", fmt.Sprintf("%d", warningID)),
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", warning.CompanyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute("evidence_count", fmt.Sprintf("%d", len(warning.CounterEvidence))),
			sdk.NewAttribute("response_text", warning.ResponseText),
		),
	)

	k.Logger(ctx).Info("freeze warning escalated to Archon review",
		"warning_id", warningID,
		"company_id", warning.CompanyID,
		"symbol", company.Symbol,
		"evidence_count", len(warning.CounterEvidence),
	)

	// TODO: In a full implementation, this would create a governance proposal
	// or assign Archon-tier reviewers to evaluate the counter-evidence.
	// For now, we just mark it as escalated and log it.

	return nil
}

// ClearFreezeWarning clears a freeze warning (called by Archons after reviewing counter-evidence)
func (k Keeper) ClearFreezeWarning(ctx sdk.Context, warningID uint64, clearedBy string, reason string) error {
	warning, found := k.GetFreezeWarning(ctx, warningID)
	if !found {
		return fmt.Errorf("warning not found")
	}

	company, found := k.getCompany(ctx, warning.CompanyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	// Update warning status
	warning.Status = types.FreezeWarningStatusCleared
	warning.ResolvedAt = ctx.BlockTime()
	warning.Resolution = fmt.Sprintf("Cleared by %s: %s", clearedBy, reason)
	if err := k.SetFreezeWarning(ctx, warning); err != nil {
		return err
	}

	// Resume trading if halted
	if k.IsTradingHalted(ctx, warning.CompanyID) {
		if err := k.ResumeTrading(ctx, warning.CompanyID, clearedBy); err != nil {
			k.Logger(ctx).Error("failed to resume trading",
				"company_id", warning.CompanyID,
				"error", err,
			)
		}
	}

	// Remove treasury withdrawal limits
	k.DeleteTreasuryWithdrawalLimit(ctx, warning.CompanyID)

	// Emit cleared event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFreezeWarningCleared,
			sdk.NewAttribute("warning_id", fmt.Sprintf("%d", warningID)),
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", warning.CompanyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute("cleared_by", clearedBy),
			sdk.NewAttribute(types.AttributeKeyReason, reason),
		),
	)

	k.Logger(ctx).Info("freeze warning cleared",
		"warning_id", warningID,
		"company_id", warning.CompanyID,
		"symbol", company.Symbol,
		"cleared_by", clearedBy,
	)

	return nil
}

// =============================================================================
// Treasury Withdrawal Limit Operations
// =============================================================================

// SetTreasuryWithdrawalLimit stores a treasury withdrawal limit
func (k Keeper) SetTreasuryWithdrawalLimit(ctx sdk.Context, limit types.TreasuryWithdrawalLimit) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTreasuryWithdrawalLimitKey(limit.CompanyID)
	bz, err := json.Marshal(limit)
	if err != nil {
		return fmt.Errorf("failed to marshal treasury withdrawal limit: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// GetTreasuryWithdrawalLimit retrieves a treasury withdrawal limit
func (k Keeper) GetTreasuryWithdrawalLimit(ctx sdk.Context, companyID uint64) (types.TreasuryWithdrawalLimit, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTreasuryWithdrawalLimitKey(companyID)
	bz := store.Get(key)
	if bz == nil {
		return types.TreasuryWithdrawalLimit{}, false
	}

	var limit types.TreasuryWithdrawalLimit
	if err := json.Unmarshal(bz, &limit); err != nil {
		return types.TreasuryWithdrawalLimit{}, false
	}
	return limit, true
}

// DeleteTreasuryWithdrawalLimit removes a treasury withdrawal limit
func (k Keeper) DeleteTreasuryWithdrawalLimit(ctx sdk.Context, companyID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetTreasuryWithdrawalLimitKey(companyID)
	store.Delete(key)
}

// CheckTreasuryWithdrawalAllowed checks if a treasury withdrawal is allowed
// Returns error if withdrawal exceeds limits during freeze warning period
func (k Keeper) CheckTreasuryWithdrawalAllowed(ctx sdk.Context, companyID uint64, amount math.Int) error {
	// Check if there's an active withdrawal limit
	limit, found := k.GetTreasuryWithdrawalLimit(ctx, companyID)
	if !found {
		return nil // No limit in place, withdrawal allowed
	}

	// Check if we need to reset daily counter (new day)
	if !limit.LastResetDate.Equal(ctx.BlockTime().Truncate(24 * time.Hour)) {
		// Reset daily counter
		limit.WithdrawnToday = math.ZeroInt()
		limit.LastResetDate = ctx.BlockTime().Truncate(24 * time.Hour)
	}

	// Check single withdrawal limit (10% per withdrawal)
	if amount.GT(limit.MaxWithdrawal) {
		return types.ErrTreasuryWithdrawalLimitExceeded
	}

	// Check daily limit (20% per day)
	totalWithdrawnToday := limit.WithdrawnToday.Add(amount)
	if totalWithdrawnToday.GT(limit.DailyLimit) {
		return types.ErrTreasuryWithdrawalLimitExceeded
	}

	// Update withdrawn today counter
	limit.WithdrawnToday = totalWithdrawnToday
	if err := k.SetTreasuryWithdrawalLimit(ctx, limit); err != nil {
		return err
	}

	return nil
}

// getTreasuryBalance gets the treasury balance for a company
func (k Keeper) getTreasuryBalance(ctx sdk.Context, companyID uint64) math.Int {
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return math.ZeroInt()
	}

	// Using founder address as treasury for now
	companyAddr, err := sdk.AccAddressFromBech32(company.Founder)
	if err != nil {
		return math.ZeroInt()
	}

	treasuryBalance := k.bankKeeper.GetBalance(ctx, companyAddr, "uhodl")
	return treasuryBalance.Amount
}

// =============================================================================
// AUTHORIZED RESPONDERS (Issue #11)
// =============================================================================

// isAuthorizedResponder checks if address can respond on behalf of company
func (k Keeper) isAuthorizedResponder(ctx sdk.Context, company types.Company, responder string) bool {
	// Founder always authorized
	if company.Founder == responder {
		return true
	}

	// Check company's authorized responders list
	for _, auth := range company.AuthorizedResponders {
		if auth == responder {
			return true
		}
	}

	// TODO: Could add check if responder is a major shareholder (>10% ownership)
	// This would require checking shareholdings, which could be added later

	return false
}

// AuthorizeCompanyResponder allows founder to authorize someone to respond to warnings
func (k Keeper) AuthorizeCompanyResponder(ctx sdk.Context, companyID uint64, founder string, newResponder string) error {
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	if company.Founder != founder {
		return types.ErrUnauthorized
	}

	// Validate new responder address
	if _, err := sdk.AccAddressFromBech32(newResponder); err != nil {
		return fmt.Errorf("invalid responder address: %w", err)
	}

	// Check if already authorized
	for _, auth := range company.AuthorizedResponders {
		if auth == newResponder {
			return fmt.Errorf("address already authorized")
		}
	}

	// Add to authorized responders
	company.AuthorizedResponders = append(company.AuthorizedResponders, newResponder)
	company.UpdatedAt = ctx.BlockTime()
	if err := k.SetCompany(ctx, company); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeResponderAuthorized,
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute("founder", founder),
			sdk.NewAttribute("responder", newResponder),
		),
	)

	k.Logger(ctx).Info("authorized company responder",
		"company_id", companyID,
		"responder", newResponder,
	)

	return nil
}

// RevokeCompanyResponder allows founder to revoke authorization
func (k Keeper) RevokeCompanyResponder(ctx sdk.Context, companyID uint64, founder string, responder string) error {
	company, found := k.getCompany(ctx, companyID)
	if !found {
		return types.ErrCompanyNotFound
	}

	if company.Founder != founder {
		return types.ErrUnauthorized
	}

	// Remove from authorized responders
	newAuthorized := []string{}
	found = false
	for _, auth := range company.AuthorizedResponders {
		if auth != responder {
			newAuthorized = append(newAuthorized, auth)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("responder not found in authorized list")
	}

	company.AuthorizedResponders = newAuthorized
	company.UpdatedAt = ctx.BlockTime()
	if err := k.SetCompany(ctx, company); err != nil {
		return err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeResponderRevoked,
			sdk.NewAttribute(types.AttributeKeyCompanyID, fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute(types.AttributeKeyCompanySymbol, company.Symbol),
			sdk.NewAttribute("founder", founder),
			sdk.NewAttribute("responder", responder),
		),
	)

	k.Logger(ctx).Info("revoked company responder",
		"company_id", companyID,
		"responder", responder,
	)

	return nil
}
