package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

// =============================================================================
// Company Queries
// =============================================================================

func (q queryServer) Company(goCtx context.Context, req *types.QueryCompanyRequest) (*types.QueryCompanyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	company, found := q.getCompany(ctx, req.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	return &types.QueryCompanyResponse{
		Company: company,
	}, nil
}

func (q queryServer) CompanyBySymbol(goCtx context.Context, req *types.QueryCompanyBySymbolRequest) (*types.QueryCompanyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	companies := q.GetAllCompanies(ctx)
	for _, company := range companies {
		if company.Symbol == req.Symbol {
			return &types.QueryCompanyResponse{
				Company: company,
			}, nil
		}
	}

	return nil, types.ErrCompanyNotFound
}

func (q queryServer) Companies(goCtx context.Context, req *types.QueryCompaniesRequest) (*types.QueryCompaniesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	companies := q.GetAllCompanies(ctx)

	// Apply filters
	filtered := []types.Company{}
	for _, company := range companies {
		if req.Status != "" && company.Status.String() != req.Status {
			continue
		}
		if req.Industry != "" && company.Industry != req.Industry {
			continue
		}
		filtered = append(filtered, company)
	}

	return &types.QueryCompaniesResponse{
		Companies: filtered,
		Pagination: &types.PageResponse{
			Total: uint64(len(filtered)),
		},
	}, nil
}

// =============================================================================
// Share Class Queries
// =============================================================================

func (q queryServer) ShareClass(goCtx context.Context, req *types.QueryShareClassRequest) (*types.QueryShareClassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	shareClass, found := q.getShareClass(ctx, req.CompanyID, req.ClassID)
	if !found {
		return nil, types.ErrShareClassNotFound
	}

	return &types.QueryShareClassResponse{
		ShareClass: shareClass,
	}, nil
}

func (q queryServer) ShareClasses(goCtx context.Context, req *types.QueryShareClassesRequest) (*types.QueryShareClassesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	shareClasses := q.GetCompanyShareClasses(ctx, req.CompanyID)

	return &types.QueryShareClassesResponse{
		ShareClasses: shareClasses,
	}, nil
}

// =============================================================================
// Shareholding Queries
// =============================================================================

func (q queryServer) Shareholding(goCtx context.Context, req *types.QueryShareholdingRequest) (*types.QueryShareholdingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	shareholding, found := q.getShareholding(ctx, req.CompanyID, req.ClassID, req.Owner)
	if !found {
		return nil, types.ErrShareholdingNotFound
	}

	return &types.QueryShareholdingResponse{
		Shareholding: shareholding,
	}, nil
}

func (q queryServer) ShareholdingsByOwner(goCtx context.Context, req *types.QueryShareholdingsByOwnerRequest) (*types.QueryShareholdingsByOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	shareholdings := q.GetOwnerShareholdings(ctx, req.Owner)

	return &types.QueryShareholdingsByOwnerResponse{
		Shareholdings: shareholdings,
		Pagination: &types.PageResponse{
			Total: uint64(len(shareholdings)),
		},
	}, nil
}

func (q queryServer) ShareholdingsByCompany(goCtx context.Context, req *types.QueryShareholdingsByCompanyRequest) (*types.QueryShareholdingsByCompanyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var shareholdings []types.Shareholding
	if req.ClassID != "" {
		shareholdings = q.GetCompanyShareholdings(ctx, req.CompanyID, req.ClassID)
	} else {
		// Get all share classes and aggregate
		shareClasses := q.GetCompanyShareClasses(ctx, req.CompanyID)
		for _, sc := range shareClasses {
			classHoldings := q.GetCompanyShareholdings(ctx, req.CompanyID, sc.ClassID)
			shareholdings = append(shareholdings, classHoldings...)
		}
	}

	return &types.QueryShareholdingsByCompanyResponse{
		Shareholdings: shareholdings,
		Pagination: &types.PageResponse{
			Total: uint64(len(shareholdings)),
		},
	}, nil
}

// =============================================================================
// Anti-Dilution Queries
// =============================================================================

func (q queryServer) AntiDilutionProvision(goCtx context.Context, req *types.QueryAntiDilutionProvisionRequest) (*types.QueryAntiDilutionProvisionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	provision, found := q.GetAntiDilutionProvision(ctx, req.CompanyID, req.ClassID)

	return &types.QueryAntiDilutionProvisionResponse{
		Provision: provision,
		Found:     found,
	}, nil
}

func (q queryServer) AntiDilutionProvisions(goCtx context.Context, req *types.QueryAntiDilutionProvisionsRequest) (*types.QueryAntiDilutionProvisionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	provisions := q.GetAntiDilutionProvisions(ctx, req.CompanyID)

	return &types.QueryAntiDilutionProvisionsResponse{
		Provisions: provisions,
	}, nil
}

func (q queryServer) IssuanceHistory(goCtx context.Context, req *types.QueryIssuanceHistoryRequest) (*types.QueryIssuanceHistoryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	records := q.GetIssuanceHistory(ctx, req.CompanyID, req.ClassID)

	return &types.QueryIssuanceHistoryResponse{
		Records: records,
		Pagination: &types.PageResponse{
			Total: uint64(len(records)),
		},
	}, nil
}

func (q queryServer) AdjustmentHistory(goCtx context.Context, req *types.QueryAdjustmentHistoryRequest) (*types.QueryAdjustmentHistoryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	adjustments := q.GetAdjustmentHistory(ctx, req.CompanyID, req.ClassID, req.Shareholder)

	return &types.QueryAdjustmentHistoryResponse{
		Adjustments: adjustments,
		Pagination: &types.PageResponse{
			Total: uint64(len(adjustments)),
		},
	}, nil
}

// =============================================================================
// Cap Table Query (Comprehensive Company View)
// =============================================================================

func (q queryServer) CompanyCapTable(goCtx context.Context, req *types.QueryCompanyCapTableRequest) (*types.QueryCompanyCapTableResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get company
	company, found := q.getCompany(ctx, req.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Get share classes with summaries
	shareClasses := q.GetCompanyShareClasses(ctx, req.CompanyID)
	classSummaries := make([]types.ShareClassSummary, 0, len(shareClasses))

	totalOutstanding := math.ZeroInt()
	for _, sc := range shareClasses {
		holdings := q.GetCompanyShareholdings(ctx, req.CompanyID, sc.ClassID)

		summary := types.ShareClassSummary{
			ClassID:           sc.ClassID,
			Name:              sc.Name,
			AuthorizedShares:  sc.AuthorizedShares,
			IssuedShares:      sc.IssuedShares,
			OutstandingShares: sc.OutstandingShares,
			HolderCount:       uint64(len(holdings)),
			HasAntiDilution:   sc.HasAntiDilution,
			AntiDilutionType:  sc.AntiDilutionType,
			VotingRights:      sc.VotingRights,
			DividendRights:    sc.DividendRights,
		}
		classSummaries = append(classSummaries, summary)
		totalOutstanding = totalOutstanding.Add(sc.OutstandingShares)
	}

	// Get top shareholders
	allHoldings := []types.Shareholding{}
	for _, sc := range shareClasses {
		holdings := q.GetCompanyShareholdings(ctx, req.CompanyID, sc.ClassID)
		allHoldings = append(allHoldings, holdings...)
	}

	// Aggregate holdings by owner and calculate ownership
	ownerTotals := make(map[string]math.Int)
	ownerClasses := make(map[string]string)
	for _, h := range allHoldings {
		if existing, ok := ownerTotals[h.Owner]; ok {
			ownerTotals[h.Owner] = existing.Add(h.Shares)
		} else {
			ownerTotals[h.Owner] = h.Shares
			ownerClasses[h.Owner] = h.ClassID
		}
	}

	topHolders := make([]types.ShareholderSummary, 0)
	for owner, shares := range ownerTotals {
		var ownershipPercent math.LegacyDec
		if totalOutstanding.IsPositive() {
			ownershipPercent = math.LegacyNewDecFromInt(shares).QuoInt(totalOutstanding).MulInt64(100)
		} else {
			ownershipPercent = math.LegacyZeroDec()
		}

		topHolders = append(topHolders, types.ShareholderSummary{
			Owner:            owner,
			TotalShares:      shares,
			OwnershipPercent: ownershipPercent,
			ClassID:          ownerClasses[owner],
			VotingPower:      ownershipPercent, // Simplified: assumes 1 share = 1 vote
		})
	}

	// Sort by shares (descending) and limit to top 10
	// Simple bubble sort for small lists
	for i := 0; i < len(topHolders); i++ {
		for j := i + 1; j < len(topHolders); j++ {
			if topHolders[j].TotalShares.GT(topHolders[i].TotalShares) {
				topHolders[i], topHolders[j] = topHolders[j], topHolders[i]
			}
		}
	}
	if len(topHolders) > 10 {
		topHolders = topHolders[:10]
	}

	// Calculate issuance statistics
	issuanceRecords := q.GetIssuanceHistory(ctx, req.CompanyID, "")
	issuanceStats := types.IssuanceStatistics{
		TotalIssuances:    uint64(len(issuanceRecords)),
		TotalSharesIssued: math.ZeroInt(),
		TotalRaised:       math.LegacyZeroDec(),
		DownRoundCount:    0,
	}

	for _, record := range issuanceRecords {
		issuanceStats.TotalSharesIssued = issuanceStats.TotalSharesIssued.Add(record.SharesIssued)
		issuanceStats.TotalRaised = issuanceStats.TotalRaised.Add(record.TotalRaised)
		if record.IsDownRound {
			issuanceStats.DownRoundCount++
		}
		issuanceStats.LastIssuancePrice = record.IssuePrice
	}

	if issuanceStats.TotalSharesIssued.IsPositive() {
		issuanceStats.AverageIssuePrice = issuanceStats.TotalRaised.QuoInt(issuanceStats.TotalSharesIssued)
	}

	// Calculate anti-dilution summary
	provisions := q.GetAntiDilutionProvisions(ctx, req.CompanyID)
	adjustments := q.GetAdjustmentHistory(ctx, req.CompanyID, "", "")

	antiDilutionSummary := types.AntiDilutionSummary{
		ProtectedClasses:    uint64(len(provisions)),
		TotalAdjustments:    uint64(len(adjustments)),
		TotalSharesAdjusted: math.ZeroInt(),
		ActiveProvisions:    0,
	}

	for _, p := range provisions {
		if p.IsActive {
			antiDilutionSummary.ActiveProvisions++
		}
	}

	for _, a := range adjustments {
		antiDilutionSummary.TotalSharesAdjusted = antiDilutionSummary.TotalSharesAdjusted.Add(a.SharesAdded)
	}

	return &types.QueryCompanyCapTableResponse{
		Company:       company,
		ShareClasses:  classSummaries,
		TopHolders:    topHolders,
		IssuanceStats: issuanceStats,
		AntiDilution:  antiDilutionSummary,
	}, nil
}

// =============================================================================
// Shareholder Alert Queries
// =============================================================================

// GetActiveAlertsForHolder returns all companies a holder owns shares in that have active warnings/investigations
func (q queryServer) GetActiveAlertsForHolder(ctx sdk.Context, holderAddress string) ([]types.ShareholderAlert, error) {
	// Get all shareholdings for this holder
	shareholdings := q.GetOwnerShareholdings(ctx, holderAddress)

	var alerts []types.ShareholderAlert

	// For each company the holder has shares in, check for active warnings
	for _, holding := range shareholdings {
		// Check for freeze warning
		if warning, found := q.GetFreezeWarningByCompany(ctx, holding.CompanyID); found {
			if warning.Status == types.FreezeWarningStatusPending ||
				warning.Status == types.FreezeWarningStatusResponded {
				company, _ := q.getCompany(ctx, holding.CompanyID)
				alerts = append(alerts, types.ShareholderAlert{
					CompanyID:     holding.CompanyID,
					CompanySymbol: company.Symbol,
					AlertType:     "fraud_investigation_warning",
					WarningID:     warning.ID,
					ExpiresAt:     warning.WarningExpiresAt,
					Message:       "Your holdings in this company are under fraud investigation. Trading may be halted.",
					SharesHeld:    holding.Shares,
					ClassID:       holding.ClassID,
				})
			}
		}

		// Check for trading halt
		if q.IsTradingHalted(ctx, holding.CompanyID) {
			company, _ := q.getCompany(ctx, holding.CompanyID)
			alerts = append(alerts, types.ShareholderAlert{
				CompanyID:     holding.CompanyID,
				CompanySymbol: company.Symbol,
				AlertType:     "trading_halted",
				Message:       "Trading is currently halted for this company due to ongoing investigation.",
				SharesHeld:    holding.Shares,
				ClassID:       holding.ClassID,
			})
		}
	}

	return alerts, nil
}
