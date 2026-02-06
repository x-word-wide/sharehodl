package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

// =============================================================================
// Dividend Audit Storage Methods
// =============================================================================

// GetNextAuditID returns the next audit ID and increments the counter
func (k Keeper) GetNextAuditID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DividendAuditCounterKey)

	var counter uint64 = 1 // Start from 1
	if bz != nil {
		counter = sdk.BigEndianToUint64(bz)
	}

	// Increment counter for next use
	store.Set(types.DividendAuditCounterKey, sdk.Uint64ToBigEndian(counter+1))
	return counter
}

// GetDividendAudit returns a dividend audit by ID
func (k Keeper) GetDividendAudit(ctx sdk.Context, auditID uint64) (types.DividendAudit, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendAuditKey(auditID)
	bz := store.Get(key)
	if bz == nil {
		return types.DividendAudit{}, false
	}

	var audit types.DividendAudit
	if err := json.Unmarshal(bz, &audit); err != nil {
		return types.DividendAudit{}, false
	}
	return audit, true
}

// SetDividendAudit stores a dividend audit
func (k Keeper) SetDividendAudit(ctx sdk.Context, audit types.DividendAudit) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDividendAuditKey(audit.ID)
	bz, err := json.Marshal(audit)
	if err != nil {
		return fmt.Errorf("failed to marshal dividend audit: %w", err)
	}
	store.Set(key, bz)

	// Create index: dividend_id -> audit_id
	indexKey := types.GetAuditByDividendKey(audit.DividendID)
	store.Set(indexKey, sdk.Uint64ToBigEndian(audit.ID))

	// Create index: company_id -> audit_id
	companyIndexKey := types.GetAuditByCompanyKey(audit.CompanyID, audit.ID)
	store.Set(companyIndexKey, sdk.Uint64ToBigEndian(audit.ID))
	return nil
}

// DeleteDividendAudit removes a dividend audit
func (k Keeper) DeleteDividendAudit(ctx sdk.Context, audit types.DividendAudit) {
	store := ctx.KVStore(k.storeKey)

	// Delete main record
	key := types.GetDividendAuditKey(audit.ID)
	store.Delete(key)

	// Delete dividend index
	indexKey := types.GetAuditByDividendKey(audit.DividendID)
	store.Delete(indexKey)

	// Delete company index
	companyIndexKey := types.GetAuditByCompanyKey(audit.CompanyID, audit.ID)
	store.Delete(companyIndexKey)
}

// GetAuditByDividend returns the audit for a specific dividend
func (k Keeper) GetAuditByDividend(ctx sdk.Context, dividendID uint64) (types.DividendAudit, bool) {
	store := ctx.KVStore(k.storeKey)
	indexKey := types.GetAuditByDividendKey(dividendID)
	bz := store.Get(indexKey)
	if bz == nil {
		return types.DividendAudit{}, false
	}

	auditID := sdk.BigEndianToUint64(bz)
	return k.GetDividendAudit(ctx, auditID)
}

// GetAuditsByCompany returns all audits for a company
func (k Keeper) GetAuditsByCompany(ctx sdk.Context, companyID uint64) []types.DividendAudit {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetAuditsByCompanyPrefix(companyID))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	audits := []types.DividendAudit{}
	for ; iterator.Valid(); iterator.Next() {
		auditID := sdk.BigEndianToUint64(iterator.Value())
		if audit, found := k.GetDividendAudit(ctx, auditID); found {
			audits = append(audits, audit)
		}
	}

	return audits
}

// =============================================================================
// Audit Verification Methods
// =============================================================================

// GetAuditVerification returns a specific verification
func (k Keeper) GetAuditVerification(ctx sdk.Context, auditID uint64, verifier string) (types.AuditVerification, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAuditVerificationKey(auditID, verifier)
	bz := store.Get(key)
	if bz == nil {
		return types.AuditVerification{}, false
	}

	var verification types.AuditVerification
	if err := json.Unmarshal(bz, &verification); err != nil {
		return types.AuditVerification{}, false
	}
	return verification, true
}

// SetAuditVerification stores an audit verification
func (k Keeper) SetAuditVerification(ctx sdk.Context, verification types.AuditVerification) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAuditVerificationKey(verification.AuditID, verification.Verifier)
	bz, err := json.Marshal(verification)
	if err != nil {
		return fmt.Errorf("failed to marshal audit verification: %w", err)
	}
	store.Set(key, bz)
	return nil
}

// GetAuditVerifications returns all verifications for an audit
func (k Keeper) GetAuditVerifications(ctx sdk.Context, auditID uint64) []types.AuditVerification {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetAuditVerificationsByAuditPrefix(auditID))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	verifications := []types.AuditVerification{}
	for ; iterator.Valid(); iterator.Next() {
		var verification types.AuditVerification
		if err := json.Unmarshal(iterator.Value(), &verification); err != nil {
			continue
		}
		verifications = append(verifications, verification)
	}

	return verifications
}

// =============================================================================
// Audit Business Logic
// =============================================================================

// CreateAuditForDividend creates a new audit record for a dividend declaration
func (k Keeper) CreateAuditForDividend(
	ctx sdk.Context,
	dividendID uint64,
	companyID uint64,
	submitter string,
	auditInfo types.AuditInfo,
) (uint64, error) {
	// Validate audit info
	if err := auditInfo.Validate(); err != nil {
		return 0, err
	}

	// Validate report date using block time
	if auditInfo.ReportDate.After(ctx.BlockTime()) {
		return 0, types.ErrAuditReportDateFuture
	}
	twoYearsAgo := ctx.BlockTime().AddDate(-2, 0, 0)
	if auditInfo.ReportDate.Before(twoYearsAgo) {
		return 0, types.ErrAuditReportTooOld
	}

	// Check if audit already exists for this dividend
	if _, found := k.GetAuditByDividend(ctx, dividendID); found {
		return 0, types.ErrAuditAlreadyExists
	}

	// Check content hash uniqueness - prevent duplicate audit documents
	if k.IsContentHashUsed(ctx, auditInfo.ContentHash) {
		return 0, types.ErrAuditContentHashDuplicate
	}

	// Get next audit ID
	auditID := k.GetNextAuditID(ctx)

	// Parse document type
	docType := types.ParseAuditDocumentType(auditInfo.DocumentType)

	// Create audit record using block time for consistency
	audit := types.NewDividendAudit(
		auditID,
		dividendID,
		companyID,
		submitter,
		auditInfo.ContentHash,
		auditInfo.StorageURI,
		docType,
		auditInfo.FileName,
		auditInfo.FileSize,
		auditInfo.Title,
		auditInfo.AuditorName,
		auditInfo.AuditorID,
		auditInfo.AuditPeriod,
		auditInfo.ReportDate,
		ctx.BlockTime(), // Use block time instead of time.Now()
	)

	// Set optional financial summary
	audit.ReportedRevenue = auditInfo.ReportedRevenue
	audit.ReportedProfit = auditInfo.ReportedProfit
	audit.ReportedDividend = auditInfo.ReportedDividend

	// Validate the full audit record
	if err := audit.Validate(); err != nil {
		return 0, types.ErrInvalidInput.Wrap(err.Error())
	}

	// Store audit
	if err := k.SetDividendAudit(ctx, audit); err != nil {
		return 0, err
	}

	// Store content hash for uniqueness check
	k.SetContentHashUsed(ctx, auditInfo.ContentHash, auditID)

	k.Logger(ctx).Info("created dividend audit",
		"audit_id", auditID,
		"dividend_id", dividendID,
		"company_id", companyID,
		"content_hash", auditInfo.ContentHash,
		"document_type", auditInfo.DocumentType,
		"auditor_name", auditInfo.AuditorName,
	)

	return auditID, nil
}

// VerifyAudit allows a validator to verify or reject an audit
func (k Keeper) VerifyAudit(
	ctx sdk.Context,
	auditID uint64,
	verifier string,
	approved bool,
	comments string,
) error {
	// CRITICAL: Verify that verifier is actually a validator
	verifierAddr, err := sdk.AccAddressFromBech32(verifier)
	if err != nil {
		return types.ErrInvalidInput.Wrap("invalid verifier address")
	}

	// Check if verifier is a validator using the validator keeper
	if k.validatorKeeper != nil {
		if !k.validatorKeeper.IsValidator(ctx, verifierAddr) {
			return types.ErrNotValidator
		}
	}
	// If validator keeper not set, log warning but allow (for testing/development)
	if k.validatorKeeper == nil {
		k.Logger(ctx).Warn("validator keeper not set - skipping validator authorization check",
			"verifier", verifier,
			"audit_id", auditID,
		)
	}

	// Get audit
	audit, found := k.GetDividendAudit(ctx, auditID)
	if !found {
		return types.ErrAuditNotFound
	}

	// Check if already verified by this verifier
	if _, found := k.GetAuditVerification(ctx, auditID, verifier); found {
		return types.ErrAuditAlreadyVerified
	}

	// Create verification record
	verification := types.AuditVerification{
		AuditID:    auditID,
		Verifier:   verifier,
		Approved:   approved,
		Comments:   comments,
		VerifiedAt: ctx.BlockTime(),
	}

	// Validate verification
	if err := verification.Validate(); err != nil {
		return types.ErrInvalidInput.Wrap(err.Error())
	}

	// Store verification
	if err := k.SetAuditVerification(ctx, verification); err != nil {
		return err
	}

	// Update audit status based on verifications
	verifications := k.GetAuditVerifications(ctx, auditID)
	verifications = append(verifications, verification) // Include the new one

	approvalCount := 0
	rejectionCount := 0
	for _, v := range verifications {
		if v.Approved {
			approvalCount++
		} else {
			rejectionCount++
		}
	}

	// Add verifier to the list
	audit.VerifiedBy = append(audit.VerifiedBy, verifier)

	// Determine status based on verification counts
	// Require at least 2 approvals for verification, 2 rejections for rejection
	if approvalCount >= 2 {
		audit.Status = types.AuditStatusVerified
		audit.VerifiedAt = ctx.BlockTime()
	} else if rejectionCount >= 2 {
		audit.Status = types.AuditStatusRejected
		audit.RejectionReason = comments
	}

	audit.UpdatedAt = ctx.BlockTime()
	if err := k.SetDividendAudit(ctx, audit); err != nil {
		return err
	}

	k.Logger(ctx).Info("audit verification submitted",
		"audit_id", auditID,
		"verifier", verifier,
		"approved", approved,
		"new_status", audit.Status.String(),
	)

	return nil
}

// DisputeAudit marks an audit as disputed
func (k Keeper) DisputeAudit(ctx sdk.Context, auditID uint64, reason string) error {
	audit, found := k.GetDividendAudit(ctx, auditID)
	if !found {
		return types.ErrAuditNotFound
	}

	audit.Status = types.AuditStatusDisputed
	audit.RejectionReason = reason
	audit.UpdatedAt = ctx.BlockTime()

	if err := k.SetDividendAudit(ctx, audit); err != nil {
		return err
	}

	k.Logger(ctx).Info("audit disputed",
		"audit_id", auditID,
		"reason", reason,
	)

	return nil
}

// GetAuditSummary returns a simplified audit summary for a dividend
func (k Keeper) GetAuditSummary(ctx sdk.Context, dividendID uint64) (types.AuditSummary, bool) {
	audit, found := k.GetAuditByDividend(ctx, dividendID)
	if !found {
		return types.AuditSummary{}, false
	}

	return audit.ToSummary(), true
}

// GetAllAudits returns all dividend audits
func (k Keeper) GetAllAudits(ctx sdk.Context) []types.DividendAudit {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DividendAuditPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	audits := []types.DividendAudit{}
	for ; iterator.Valid(); iterator.Next() {
		var audit types.DividendAudit
		if err := json.Unmarshal(iterator.Value(), &audit); err != nil {
			continue
		}
		audits = append(audits, audit)
	}

	return audits
}

// GetAuditsByStatus returns all audits with a specific status
func (k Keeper) GetAuditsByStatus(ctx sdk.Context, status types.AuditVerificationStatus) []types.DividendAudit {
	audits := k.GetAllAudits(ctx)
	filtered := []types.DividendAudit{}
	for _, audit := range audits {
		if audit.Status == status {
			filtered = append(filtered, audit)
		}
	}
	return filtered
}

// =============================================================================
// Content Hash Uniqueness Methods
// =============================================================================

// IsContentHashUsed checks if a content hash is already used by another audit
func (k Keeper) IsContentHashUsed(ctx sdk.Context, contentHash string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAuditContentHashKey(contentHash)
	return store.Has(key)
}

// SetContentHashUsed marks a content hash as used by storing the audit ID
func (k Keeper) SetContentHashUsed(ctx sdk.Context, contentHash string, auditID uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAuditContentHashKey(contentHash)
	store.Set(key, sdk.Uint64ToBigEndian(auditID))
}

// GetAuditIDByContentHash returns the audit ID that uses this content hash
func (k Keeper) GetAuditIDByContentHash(ctx sdk.Context, contentHash string) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAuditContentHashKey(contentHash)
	bz := store.Get(key)
	if bz == nil {
		return 0, false
	}
	return sdk.BigEndianToUint64(bz), true
}
