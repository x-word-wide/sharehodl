package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/equity/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) *msgServer {
	return &msgServer{Keeper: keeper}
}

// CreateCompany handles company creation
func (k msgServer) CreateCompany(goCtx context.Context, msg *types.SimpleMsgCreateCompany) (*types.MsgCreateCompanyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate creator address
	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid creator address: %v", err)
	}

	// Check if symbol is already taken
	if k.IsSymbolTaken(ctx, msg.Symbol) {
		return nil, types.ErrSymbolTaken
	}

	// Validate listing fee payment
	if !msg.ListingFee.IsZero() {
		balance := k.bankKeeper.GetAllBalances(ctx, creatorAddr)
		if !balance.IsAllGTE(msg.ListingFee) {
			return nil, types.ErrInsufficientFunds
		}

		// Transfer listing fee to module (could be used for protocol revenue)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creatorAddr, types.ModuleName, msg.ListingFee); err != nil {
			return nil, errors.Wrapf(err, "failed to transfer listing fee")
		}
	}

	// Get next company ID
	companyID := k.GetNextCompanyID(ctx)

	// Create company
	company := types.NewCompany(
		companyID,
		msg.Name,
		msg.Symbol,
		msg.Description,
		msg.Creator,
		msg.Website,
		msg.Industry,
		msg.Country,
		msg.TotalShares,
		msg.AuthorizedShares,
		msg.ParValue,
		msg.ListingFee,
	)

	// Validate company
	if err := company.Validate(); err != nil {
		return nil, errors.Wrapf(types.ErrCompanyNotFound, "invalid company: %v", err)
	}

	// Save company
	k.SetCompany(ctx, company)

	// Create default "COMMON" share class
	commonClass := types.NewShareClass(
		companyID,
		"COMMON",
		"Common Shares",
		"Common voting shares with dividend rights",
		true,  // voting rights
		true,  // dividend rights  
		false, // liquidation preference
		false, // conversion rights
		msg.AuthorizedShares,
		msg.ParValue,
		true, // transferable
	)

	k.SetShareClass(ctx, commonClass)

	// CRITICAL: Shares go to TREASURY, not founder
	// Founder controls treasury but doesn't personally own shares yet
	// This ensures all money flows through treasury and requires governance for withdrawals
	treasury := types.NewCompanyTreasury(companyID)

	// Initialize treasury share holdings
	if treasury.TreasuryShares == nil {
		treasury.TreasuryShares = make(map[string]math.Int)
	}
	if treasury.SharesSoldTotal == nil {
		treasury.SharesSoldTotal = make(map[string]math.Int)
	}

	// All initial shares go to treasury
	treasury.TreasuryShares["COMMON"] = msg.TotalShares
	treasury.SharesSoldTotal["COMMON"] = math.ZeroInt()

	if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
		return nil, err
	}

	// Initialize company authorization (owner can add co-founders as proposers later)
	auth := types.NewCompanyAuthorization(companyID, msg.Creator)
	if err := k.SetCompanyAuthorization(ctx, auth); err != nil {
		k.Logger(ctx).Error("failed to initialize company authorization", "company_id", companyID, "error", err)
		// Non-fatal: company is still created, just log the error
	}

	// DO NOT issue shares to founder - they stay in treasury
	// Founder must sell from treasury on DEX or via primary sale to get HODL

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"create_company",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute("symbol", msg.Symbol),
			sdk.NewAttribute("founder", msg.Creator),
			sdk.NewAttribute("total_shares", msg.TotalShares.String()),
		),
	)

	return &types.MsgCreateCompanyResponse{
		CompanyID: companyID,
	}, nil
}

// CreateShareClass handles share class creation
func (k msgServer) CreateShareClass(goCtx context.Context, msg *types.SimpleMsgCreateShareClass) (*types.MsgCreateShareClassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get company to verify creator is the founder
	company, found := k.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Only company founder can create new share classes
	if company.Founder != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	// Check if share class already exists
	if _, exists := k.GetShareClass(ctx, msg.CompanyID, msg.ClassID); exists {
		return nil, types.ErrShareClassAlreadyExists
	}

	// Create share class
	shareClass := types.NewShareClass(
		msg.CompanyID,
		msg.ClassID,
		msg.Name,
		msg.Description,
		msg.VotingRights,
		msg.DividendRights,
		msg.LiquidationPreference,
		msg.ConversionRights,
		msg.AuthorizedShares,
		msg.ParValue,
		msg.Transferable,
	)

	// Validate share class
	if err := shareClass.Validate(); err != nil {
		return nil, errors.Wrapf(types.ErrCompanyNotFound, "invalid share class: %v", err)
	}

	// Save share class
	k.SetShareClass(ctx, shareClass)

	// Update company's authorized shares total
	company.AuthorizedShares = company.AuthorizedShares.Add(msg.AuthorizedShares)
	company.UpdatedAt = time.Now()
	k.SetCompany(ctx, company)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"create_share_class",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("name", msg.Name),
			sdk.NewAttribute("authorized_shares", msg.AuthorizedShares.String()),
		),
	)

	return &types.MsgCreateShareClassResponse{
		Success: true,
	}, nil
}

// IssueShares handles share issuance
func (k msgServer) IssueShares(goCtx context.Context, msg *types.SimpleMsgIssueShares) (*types.MsgIssueSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get company to verify creator is the founder
	company, found := k.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Only company founder can issue shares (for now - could add board governance later)
	if company.Founder != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	// Company must be active to issue new shares
	if company.Status != types.CompanyStatusActive && company.Status != types.CompanyStatusPending {
		return nil, types.ErrCompanyNotActive
	}

	// Validate recipient address
	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid recipient address: %v", err)
	}

	// Calculate total cost
	totalCost := msg.Price.MulInt(msg.Shares)

	// If price > 0, require payment from recipient
	if !msg.Price.IsZero() {
		recipientAddr, _ := sdk.AccAddressFromBech32(msg.Recipient)
		requiredPayment := sdk.NewCoins(sdk.NewCoin(msg.PaymentDenom, totalCost.TruncateInt()))
		
		balance := k.bankKeeper.GetAllBalances(ctx, recipientAddr)
		if !balance.IsAllGTE(requiredPayment) {
			return nil, types.ErrInsufficientFunds
		}

		// Transfer payment to company founder (or could be to company treasury)
		founderAddr, _ := sdk.AccAddressFromBech32(company.Founder)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, recipientAddr, types.ModuleName, requiredPayment); err != nil {
			return nil, errors.Wrapf(err, "failed to transfer payment")
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, founderAddr, requiredPayment); err != nil {
			return nil, errors.Wrapf(err, "failed to send payment to founder")
		}
	}

	// Issue the shares
	if err := k.Keeper.IssueShares(ctx, msg.CompanyID, msg.ClassID, msg.Recipient, msg.Shares, msg.Price, msg.PaymentDenom); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"issue_shares",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("recipient", msg.Recipient),
			sdk.NewAttribute("shares", msg.Shares.String()),
			sdk.NewAttribute("price", msg.Price.String()),
			sdk.NewAttribute("total_cost", totalCost.String()),
		),
	)

	return &types.MsgIssueSharesResponse{
		SharesIssued: msg.Shares,
		TotalCost:    totalCost,
	}, nil
}

// TransferShares handles share transfers
func (k msgServer) TransferShares(goCtx context.Context, msg *types.SimpleMsgTransferShares) (*types.MsgTransferSharesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Only the owner can transfer their own shares (msg.Creator must equal msg.From)
	if msg.Creator != msg.From {
		return nil, types.ErrUnauthorized
	}

	// Get company to check if it's active
	company, found := k.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	if company.Status != types.CompanyStatusActive {
		return nil, types.ErrCompanyNotActive
	}

	// Transfer the shares
	if err := k.Keeper.TransferShares(ctx, msg.CompanyID, msg.ClassID, msg.From, msg.To, msg.Shares); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"transfer_shares",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("from", msg.From),
			sdk.NewAttribute("to", msg.To),
			sdk.NewAttribute("shares", msg.Shares.String()),
		),
	)

	return &types.MsgTransferSharesResponse{
		Success: true,
	}, nil
}

// UpdateCompany handles company information updates
func (k msgServer) UpdateCompany(goCtx context.Context, msg *types.SimpleMsgUpdateCompany) (*types.MsgUpdateCompanyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get company
	company, found := k.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Only company founder can update company info
	if company.Founder != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	// Update fields if provided
	if msg.Name != "" {
		company.Name = msg.Name
	}
	if msg.Description != "" {
		company.Description = msg.Description
	}
	if msg.Website != "" {
		company.Website = msg.Website
	}
	if msg.Industry != "" {
		company.Industry = msg.Industry
	}
	if msg.Country != "" {
		company.Country = msg.Country
	}

	company.UpdatedAt = time.Now()

	// Validate updated company
	if err := company.Validate(); err != nil {
		return nil, errors.Wrapf(types.ErrCompanyNotFound, "invalid company update: %v", err)
	}

	// Save company
	k.SetCompany(ctx, company)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"update_company",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("founder", msg.Creator),
		),
	)

	return &types.MsgUpdateCompanyResponse{
		Success: true,
	}, nil
}

// Business listing message handlers

// SubmitBusinessListing handles business listing submission
func (k msgServer) SubmitBusinessListing(goCtx context.Context, msg *types.MsgSubmitBusinessListing) (*types.MsgSubmitBusinessListingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert document uploads to business documents
	documents := make([]types.BusinessDocumentUpload, len(msg.Documents))
	copy(documents, msg.Documents)

	// Submit the listing
	listingID, err := k.Keeper.SubmitBusinessListing(
		ctx,
		msg.Creator,
		msg.CompanyName,
		msg.LegalEntityName,
		msg.RegistrationNumber,
		msg.Jurisdiction,
		msg.Industry,
		msg.BusinessDescription,
		msg.Website,
		msg.ContactEmail,
		msg.TokenSymbol,
		msg.TokenName,
		msg.FoundedDate,
		msg.TotalShares,
		msg.InitialPrice,
		msg.ListingFee,
		documents,
	)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"submit_business_listing",
			sdk.NewAttribute("listing_id", fmt.Sprintf("%d", listingID)),
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("company_name", msg.CompanyName),
			sdk.NewAttribute("token_symbol", msg.TokenSymbol),
			sdk.NewAttribute("listing_fee", msg.ListingFee.String()),
		),
	)

	return &types.MsgSubmitBusinessListingResponse{
		ListingID: listingID,
		Success:   true,
	}, nil
}

// ReviewBusinessListing handles validator review of business listings
func (k msgServer) ReviewBusinessListing(goCtx context.Context, msg *types.MsgReviewBusinessListing) (*types.MsgReviewBusinessListingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Review the listing
	err := k.Keeper.ReviewBusinessListing(
		ctx,
		msg.Validator,
		msg.ListingID,
		msg.Approved,
		msg.Comments,
		msg.Score,
		msg.Criteria,
		msg.Data,
	)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"review_business_listing",
			sdk.NewAttribute("listing_id", fmt.Sprintf("%d", msg.ListingID)),
			sdk.NewAttribute("validator", msg.Validator),
			sdk.NewAttribute("approved", fmt.Sprintf("%d", btoi(msg.Approved))),
			sdk.NewAttribute("score", fmt.Sprintf("%d", msg.Score)),
		),
	)

	return &types.MsgReviewBusinessListingResponse{
		Success: true,
	}, nil
}

// ApproveListing handles listing approval
func (k msgServer) ApproveListing(goCtx context.Context, msg *types.MsgApproveListing) (*types.MsgApproveListingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Approve the listing and create the company
	companyID, err := k.Keeper.ApproveListing(ctx, msg.ListingID, msg.Authority)
	if err != nil {
		return nil, err
	}

	// Get the listing to retrieve token symbol
	listing, found := k.GetBusinessListing(ctx, msg.ListingID)
	if !found {
		return nil, types.ErrListingNotFound
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"approve_listing",
			sdk.NewAttribute("listing_id", fmt.Sprintf("%d", msg.ListingID)),
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute("token_symbol", listing.TokenSymbol),
			sdk.NewAttribute("authority", msg.Authority),
		),
	)

	return &types.MsgApproveListingResponse{
		TokenID:     companyID,
		TokenSymbol: listing.TokenSymbol,
		Success:     true,
	}, nil
}

// RejectListing handles listing rejection
func (k msgServer) RejectListing(goCtx context.Context, msg *types.MsgRejectListing) (*types.MsgRejectListingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Reject the listing
	err := k.Keeper.RejectListing(ctx, msg.ListingID, msg.Authority, msg.Reason)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"reject_listing",
			sdk.NewAttribute("listing_id", fmt.Sprintf("%d", msg.ListingID)),
			sdk.NewAttribute("authority", msg.Authority),
			sdk.NewAttribute("reason", msg.Reason),
		),
	)

	return &types.MsgRejectListingResponse{
		Success: true,
	}, nil
}

// SuspendListing handles listing suspension
func (k msgServer) SuspendListing(goCtx context.Context, msg *types.MsgSuspendListing) (*types.MsgSuspendListingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Suspend the listing
	err := k.Keeper.SuspendListing(ctx, msg.ListingID, msg.Authority, msg.Reason)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"suspend_listing",
			sdk.NewAttribute("listing_id", fmt.Sprintf("%d", msg.ListingID)),
			sdk.NewAttribute("authority", msg.Authority),
			sdk.NewAttribute("reason", msg.Reason),
		),
	)

	return &types.MsgSuspendListingResponse{
		Success: true,
	}, nil
}

// UpdateListingRequirements handles updating listing requirements
func (k msgServer) UpdateListingRequirements(goCtx context.Context, msg *types.MsgUpdateListingRequirements) (*types.MsgUpdateListingRequirementsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// CRITICAL: Verify caller is governance module authority
	// Only governance can update listing requirements
	if msg.Authority != k.GetAuthority() {
		return nil, errors.Wrapf(types.ErrUnauthorized, "only governance module can update listing requirements (expected %s, got %s)", k.GetAuthority(), msg.Authority)
	}

	// Update listing requirements
	k.SetListingRequirements(ctx, msg.Requirements)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"update_listing_requirements",
			sdk.NewAttribute("authority", msg.Authority),
			sdk.NewAttribute("minimum_listing_fee", msg.Requirements.MinimumListingFee.String()),
			sdk.NewAttribute("required_validators", fmt.Sprintf("%d", msg.Requirements.RequiredValidators)),
		),
	)

	return &types.MsgUpdateListingRequirementsResponse{
		Success: true,
	}, nil
}

// Helper function to convert bool to int for event attributes
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Dividend message handlers

// DeclareDividend handles dividend declaration with mandatory audit document
func (k msgServer) DeclareDividend(goCtx context.Context, msg *types.MsgDeclareDividend) (*types.MsgDeclareDividendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Parse dividend type
	var dividendType types.DividendType
	switch msg.DividendType {
	case "cash":
		dividendType = types.DividendTypeCash
	case "stock":
		dividendType = types.DividendTypeStock
	case "property":
		dividendType = types.DividendTypeProperty
	case "special":
		dividendType = types.DividendTypeSpecial
	default:
		return nil, types.ErrInvalidDividendType
	}

	// Validate audit document (MANDATORY - protects shareholders from fraudulent dividends)
	if err := msg.Audit.Validate(); err != nil {
		return nil, err
	}

	// Declare dividend with audit
	dividendID, err := k.Keeper.DeclareDividend(
		ctx,
		msg.Creator,
		msg.CompanyID,
		msg.ClassID,
		dividendType,
		msg.Currency,
		msg.AmountPerShare,
		msg.ExDividendDays,
		msg.RecordDays,
		msg.PaymentDays,
		msg.TaxRate,
		msg.Description,
		msg.PaymentMethod,
		msg.StockRatio,
		msg.Audit, // Pass audit info to keeper
	)
	if err != nil {
		return nil, err
	}

	// Emit event with audit info
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"declare_dividend",
			sdk.NewAttribute("dividend_id", fmt.Sprintf("%d", dividendID)),
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("dividend_type", msg.DividendType),
			sdk.NewAttribute("amount_per_share", msg.AmountPerShare.String()),
			sdk.NewAttribute("currency", msg.Currency),
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("audit_hash", msg.Audit.ContentHash),
			sdk.NewAttribute("auditor_name", msg.Audit.AuditorName),
		),
	)

	return &types.MsgDeclareDividendResponse{
		DividendID: dividendID,
		Success:    true,
	}, nil
}

// ProcessDividend handles dividend payment processing
func (k msgServer) ProcessDividend(goCtx context.Context, msg *types.MsgProcessDividend) (*types.MsgProcessDividendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Process dividend payments in batch
	paymentsProcessed, amountPaid, isComplete, err := k.Keeper.ProcessDividendPayments(
		ctx,
		msg.DividendID,
		msg.BatchSize,
	)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"process_dividend",
			sdk.NewAttribute("dividend_id", fmt.Sprintf("%d", msg.DividendID)),
			sdk.NewAttribute("payments_processed", fmt.Sprintf("%d", paymentsProcessed)),
			sdk.NewAttribute("amount_paid", amountPaid.String()),
			sdk.NewAttribute("is_complete", fmt.Sprintf("%d", btoi(isComplete))),
			sdk.NewAttribute("processor", msg.Creator),
		),
	)

	return &types.MsgProcessDividendResponse{
		PaymentsProcessed: paymentsProcessed,
		AmountPaid:        amountPaid,
		IsComplete:        isComplete,
		Success:           true,
	}, nil
}

// ClaimDividend handles dividend claims by shareholders
func (k msgServer) ClaimDividend(goCtx context.Context, msg *types.MsgClaimDividend) (*types.MsgClaimDividendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Process dividend claim
	amount, currency, taxWithheld, netAmount, err := k.Keeper.ClaimDividend(
		ctx,
		msg.Creator,
		msg.DividendID,
	)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"claim_dividend",
			sdk.NewAttribute("dividend_id", fmt.Sprintf("%d", msg.DividendID)),
			sdk.NewAttribute("claimant", msg.Creator),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("currency", currency),
			sdk.NewAttribute("tax_withheld", taxWithheld.String()),
			sdk.NewAttribute("net_amount", netAmount.String()),
		),
	)

	return &types.MsgClaimDividendResponse{
		Amount:      amount,
		Currency:    currency,
		TaxWithheld: taxWithheld,
		NetAmount:   netAmount,
		Success:     true,
	}, nil
}

// CancelDividend handles dividend cancellation
func (k msgServer) CancelDividend(goCtx context.Context, msg *types.MsgCancelDividend) (*types.MsgCancelDividendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get dividend to verify authorization
	dividend, found := k.GetDividend(ctx, msg.DividendID)
	if !found {
		return nil, types.ErrDividendNotFound
	}

	// Get company to verify creator is authorized
	company, found := k.getCompany(ctx, dividend.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Only company founder can cancel dividends
	if company.Founder != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	// Can only cancel if not yet paid
	if dividend.Status == types.DividendStatusPaid {
		return nil, types.ErrDividendAlreadyPaid
	}

	// If cash dividend with pending approval, release escrow back to treasury
	if dividend.Type == types.DividendTypeCash && dividend.Status == types.DividendStatusPendingApproval {
		// Use escrow system to properly return funds
		if err := k.ReturnDividendEscrowToTreasury(ctx, msg.DividendID); err != nil {
			k.Logger(ctx).Error("failed to return escrow to treasury on cancel", "error", err, "dividend_id", msg.DividendID)
			// Continue with cancellation - escrow may not exist for old dividends
		} else {
			k.Logger(ctx).Info("dividend cancelled - escrow returned to treasury",
				"dividend_id", msg.DividendID,
				"company_id", dividend.CompanyID,
			)
		}
	} else if dividend.Type == types.DividendTypeCash && dividend.RemainingAmount.GT(math.LegacyZeroDec()) {
		// For partially paid dividends, refund remaining funds directly
		refundCoins := sdk.NewCoins(sdk.NewCoin(dividend.Currency, dividend.RemainingAmount.TruncateInt()))

		// Get treasury and add funds back
		treasury, found := k.GetCompanyTreasury(ctx, dividend.CompanyID)
		if found {
			treasury.Balance = treasury.Balance.Add(refundCoins...)
			treasury.UpdatedAt = ctx.BlockTime()
			if err := k.SetCompanyTreasury(ctx, treasury); err != nil {
				k.Logger(ctx).Error("failed to refund to treasury", "error", err)
			} else {
				k.Logger(ctx).Info("dividend cancelled - remaining funds returned to treasury",
					"dividend_id", msg.DividendID,
					"company_id", dividend.CompanyID,
					"refund_amount", refundCoins.String(),
				)
			}
		}
	}

	// Cancel dividend
	dividend.Status = types.DividendStatusCancelled
	dividend.UpdatedAt = ctx.BlockTime()
	dividend.Notes = msg.Reason
	k.SetDividend(ctx, dividend)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"cancel_dividend",
			sdk.NewAttribute("dividend_id", fmt.Sprintf("%d", msg.DividendID)),
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", dividend.CompanyID)),
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("reason", msg.Reason),
		),
	)

	return &types.MsgCancelDividendResponse{
		Success: true,
	}, nil
}

// SetDividendPolicy handles setting company dividend policy
func (k msgServer) SetDividendPolicy(goCtx context.Context, msg *types.MsgSetDividendPolicy) (*types.MsgSetDividendPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get company to verify creator is authorized
	company, found := k.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Only company founder can set dividend policy
	if company.Founder != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	// Create dividend policy
	policy := types.DividendPolicy{
		CompanyID:               msg.CompanyID,
		Policy:                  msg.Policy,
		MinimumPayout:           msg.MinimumPayout,
		MaximumPayout:           msg.MaximumPayout,
		TargetPayoutRatio:       msg.TargetPayoutRatio,
		PaymentFrequency:        msg.PaymentFrequency,
		PaymentMethod:           msg.PaymentMethod,
		ReinvestmentAvailable:   msg.ReinvestmentAvailable,
		AutoReinvestmentDefault: msg.AutoReinvestmentDefault,
		MinimumSharesForPayment: msg.MinimumSharesForPayment,
		FractionalShareHandling: msg.FractionalShareHandling,
		TaxOptimization:         msg.TaxOptimization,
		Description:             msg.Description,
		EffectiveDate:           ctx.BlockTime(),
		LastUpdated:             ctx.BlockTime(),
		UpdatedBy:               msg.Creator,
	}

	// Store policy
	k.StoreDividendPolicy(ctx, policy)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"set_dividend_policy",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("policy", msg.Policy),
			sdk.NewAttribute("payment_frequency", msg.PaymentFrequency),
			sdk.NewAttribute("payment_method", msg.PaymentMethod),
			sdk.NewAttribute("creator", msg.Creator),
		),
	)

	return &types.MsgSetDividendPolicyResponse{
		Success: true,
	}, nil
}

// CreateRecordSnapshot handles creating shareholder snapshots
func (k msgServer) CreateRecordSnapshot(goCtx context.Context, msg *types.MsgCreateRecordSnapshot) (*types.MsgCreateRecordSnapshotResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create record snapshot
	err := k.Keeper.CreateRecordSnapshot(ctx, msg.DividendID)
	if err != nil {
		return nil, err
	}

	// Get snapshot to return statistics
	snapshot, found := k.GetDividendSnapshot(ctx, msg.DividendID)
	if !found {
		return nil, types.ErrDividendSnapshotNotReady
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"create_record_snapshot",
			sdk.NewAttribute("dividend_id", fmt.Sprintf("%d", msg.DividendID)),
			sdk.NewAttribute("total_shareholders", fmt.Sprintf("%d", snapshot.TotalShareholders)),
			sdk.NewAttribute("total_shares", snapshot.TotalShares.String()),
			sdk.NewAttribute("creator", msg.Creator),
		),
	)

	return &types.MsgCreateRecordSnapshotResponse{
		TotalShareholders: snapshot.TotalShareholders,
		TotalShares:       snapshot.TotalShares,
		SnapshotComplete:  true,
		Success:           true,
	}, nil
}

// EnableDividendReinvestment handles DRIP settings
func (k msgServer) EnableDividendReinvestment(goCtx context.Context, msg *types.MsgEnableDividendReinvestment) (*types.MsgEnableDividendReinvestmentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get company to verify authorization
	company, found := k.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Only company founder can enable/disable DRIP
	if company.Founder != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	// Update dividend policy to enable/disable reinvestment
	policy, found := k.GetDividendPolicy(ctx, msg.CompanyID)
	if !found {
		// Create default policy if none exists
		policy = types.DividendPolicy{
			CompanyID:        msg.CompanyID,
			Policy:           "variable",
			PaymentFrequency: "quarterly",
			PaymentMethod:    "automatic",
			EffectiveDate:    ctx.BlockTime(),
			LastUpdated:      ctx.BlockTime(),
			UpdatedBy:        msg.Creator,
		}
	}

	policy.ReinvestmentAvailable = msg.Enabled
	policy.LastUpdated = ctx.BlockTime()
	policy.UpdatedBy = msg.Creator

	k.StoreDividendPolicy(ctx, policy)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"enable_dividend_reinvestment",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("enabled", fmt.Sprintf("%d", btoi(msg.Enabled))),
			sdk.NewAttribute("creator", msg.Creator),
		),
	)

	return &types.MsgEnableDividendReinvestmentResponse{
		Success: true,
	}, nil
}

// =============================================================================
// Anti-Dilution Message Handlers
// =============================================================================

// RegisterAntiDilution registers anti-dilution protection for a share class
func (k msgServer) RegisterAntiDilution(goCtx context.Context, msg *types.SimpleMsgRegisterAntiDilution) (*types.MsgRegisterAntiDilutionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Register the anti-dilution provision
	err := k.Keeper.RegisterAntiDilutionProvision(
		ctx,
		msg.CompanyID,
		msg.ClassID,
		msg.ProvisionType,
		msg.TriggerThreshold,
		msg.AdjustmentCap,
		msg.MinimumPrice,
		msg.Creator,
	)
	if err != nil {
		return nil, err
	}

	// Update share class to mark it has anti-dilution
	shareClass, found := k.getShareClass(ctx, msg.CompanyID, msg.ClassID)
	if found {
		shareClass.HasAntiDilution = true
		shareClass.AntiDilutionType = msg.ProvisionType
		shareClass.UpdatedAt = ctx.BlockTime()
		k.SetShareClass(ctx, shareClass)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"register_anti_dilution",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("provision_type", msg.ProvisionType.String()),
			sdk.NewAttribute("trigger_threshold", msg.TriggerThreshold.String()),
			sdk.NewAttribute("adjustment_cap", msg.AdjustmentCap.String()),
			sdk.NewAttribute("minimum_price", msg.MinimumPrice.String()),
			sdk.NewAttribute("creator", msg.Creator),
		),
	)

	return &types.MsgRegisterAntiDilutionResponse{
		Success:       true,
		ProvisionType: msg.ProvisionType.String(),
	}, nil
}

// UpdateAntiDilution updates an existing anti-dilution provision
func (k msgServer) UpdateAntiDilution(goCtx context.Context, msg *types.SimpleMsgUpdateAntiDilution) (*types.MsgUpdateAntiDilutionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Get company to verify creator is the founder
	company, found := k.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	if company.Founder != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	// Get existing provision
	provision, found := k.GetAntiDilutionProvision(ctx, msg.CompanyID, msg.ClassID)
	if !found {
		return nil, types.ErrAntiDilutionNotFound
	}

	// Update fields if provided
	if msg.IsActive != nil {
		provision.IsActive = *msg.IsActive
	}
	if !msg.TriggerThreshold.IsNil() && !msg.TriggerThreshold.IsZero() {
		provision.TriggerThreshold = msg.TriggerThreshold
	}
	if !msg.AdjustmentCap.IsNil() && !msg.AdjustmentCap.IsZero() {
		provision.AdjustmentCap = msg.AdjustmentCap
	}
	if !msg.MinimumPrice.IsNil() && !msg.MinimumPrice.IsNegative() {
		provision.MinimumPrice = msg.MinimumPrice
	}

	provision.UpdatedAt = ctx.BlockTime()

	// Save updated provision
	if err := k.SetAntiDilutionProvision(ctx, provision); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"update_anti_dilution",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("creator", msg.Creator),
		),
	)

	return &types.MsgUpdateAntiDilutionResponse{
		Success: true,
	}, nil
}

// =============================================================================
// Treasury Share Dilution Handlers
// =============================================================================

// ProposeShareDilution creates a governance proposal to add shares to treasury
func (k msgServer) ProposeShareDilution(goCtx context.Context, msg *types.SimpleMsgProposeShareDilution) (*types.MsgProposeShareDilutionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Verify company exists
	_, found := k.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Use authorization system: creator must be owner or authorized proposer
	if !k.CanProposeForCompany(ctx, msg.CompanyID, msg.Creator) {
		return nil, types.ErrNotAuthorizedProposer
	}

	// Verify share class exists
	shareClass, found := k.getShareClass(ctx, msg.CompanyID, msg.ClassID)
	if !found {
		return nil, types.ErrShareClassNotFound
	}

	// Check if new shares would exceed authorized
	currentIssued := shareClass.IssuedShares
	newAuthorizedNeeded := currentIssued.Add(msg.AdditionalShares)
	if newAuthorizedNeeded.GT(shareClass.AuthorizedShares) {
		return nil, types.ErrExceedsAuthorized
	}

	// NOTE: In a real implementation, this would create a governance proposal
	// For now, we'll emit an event indicating a proposal should be created
	// The governance module would then vote on it

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"propose_share_dilution",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("additional_shares", msg.AdditionalShares.String()),
			sdk.NewAttribute("purpose", msg.Purpose),
			sdk.NewAttribute("proposer", msg.Creator),
			sdk.NewAttribute("proposal_title", msg.ProposalTitle),
		),
	)

	return &types.MsgProposeShareDilutionResponse{
		ProposalID: 0, // Would be set by governance module
		Success:    true,
	}, nil
}

// ExecuteShareDilution executes approved share dilution (called by governance after vote passes)
func (k msgServer) ExecuteShareDilution(goCtx context.Context, msg *types.SimpleMsgExecuteShareDilution) (*types.MsgExecuteShareDilutionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// CRITICAL: Verify caller is governance module authority
	// Only governance can execute share dilution after a proposal passes
	if msg.Authority != k.GetAuthority() {
		return nil, errors.Wrapf(types.ErrUnauthorized, "only governance module can execute share dilution (expected %s, got %s)", k.GetAuthority(), msg.Authority)
	}

	// Verify company exists
	_, found := k.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Get share class
	shareClass, found := k.getShareClass(ctx, msg.CompanyID, msg.ClassID)
	if !found {
		return nil, types.ErrShareClassNotFound
	}

	// Add shares to treasury
	if err := k.Keeper.AddSharesToTreasury(ctx, msg.CompanyID, msg.ClassID, msg.AdditionalShares); err != nil {
		return nil, err
	}

	// Update share class authorized shares if needed
	newAuthorizedNeeded := shareClass.AuthorizedShares.Add(msg.AdditionalShares)
	if newAuthorizedNeeded.GT(shareClass.AuthorizedShares) {
		shareClass.AuthorizedShares = newAuthorizedNeeded
		shareClass.UpdatedAt = ctx.BlockTime()
		k.SetShareClass(ctx, shareClass)
	}

	// Get updated treasury shares
	treasury, _ := k.GetCompanyTreasury(ctx, msg.CompanyID)
	newTreasuryShares := treasury.GetTreasuryShares(msg.ClassID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"execute_share_dilution",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("shares_added", msg.AdditionalShares.String()),
			sdk.NewAttribute("new_treasury_shares", newTreasuryShares.String()),
			sdk.NewAttribute("new_authorized", shareClass.AuthorizedShares.String()),
		),
	)

	k.Logger(ctx).Info("share dilution executed",
		"company_id", msg.CompanyID,
		"class_id", msg.ClassID,
		"shares_added", msg.AdditionalShares.String(),
		"new_treasury_shares", newTreasuryShares.String(),
	)

	return &types.MsgExecuteShareDilutionResponse{
		NewTreasuryShares: newTreasuryShares,
		TotalAuthorized:   shareClass.AuthorizedShares,
		Success:           true,
	}, nil
}

// IssueSharesWithProtection issues shares with anti-dilution processing
func (k msgServer) IssueSharesWithProtection(goCtx context.Context, msg *types.SimpleMsgIssueSharesWithProtection) (*types.MsgIssueSharesWithProtectionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Get company to verify creator is the founder
	company, found := k.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Only company founder can issue shares
	if company.Founder != msg.Creator {
		return nil, types.ErrUnauthorized
	}

	// Company must be active
	if company.Status != types.CompanyStatusActive && company.Status != types.CompanyStatusPending {
		return nil, types.ErrCompanyNotActive
	}

	// Validate recipient address
	if _, err := sdk.AccAddressFromBech32(msg.Recipient); err != nil {
		return nil, errors.Wrapf(types.ErrUnauthorized, "invalid recipient address: %v", err)
	}

	// Calculate total cost
	totalCost := msg.Price.MulInt(msg.Shares)

	// If price > 0, require payment from recipient
	if !msg.Price.IsZero() {
		recipientAddr, _ := sdk.AccAddressFromBech32(msg.Recipient)
		requiredPayment := sdk.NewCoins(sdk.NewCoin(msg.PaymentDenom, totalCost.TruncateInt()))

		balance := k.bankKeeper.GetAllBalances(ctx, recipientAddr)
		if !balance.IsAllGTE(requiredPayment) {
			return nil, types.ErrInsufficientFunds
		}

		// Transfer payment to company founder
		founderAddr, _ := sdk.AccAddressFromBech32(company.Founder)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, recipientAddr, types.ModuleName, requiredPayment); err != nil {
			return nil, errors.Wrapf(err, "failed to transfer payment")
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, founderAddr, requiredPayment); err != nil {
			return nil, errors.Wrapf(err, "failed to send payment to founder")
		}
	}

	// Issue shares with anti-dilution protection processing
	adjustments, err := k.Keeper.IssueSharesWithAntiDilution(
		ctx,
		msg.CompanyID,
		msg.ClassID,
		msg.Recipient,
		msg.Shares,
		msg.Price,
		msg.PaymentDenom,
		msg.Purpose,
	)
	if err != nil {
		return nil, err
	}

	// Check if this was a down round
	record, _ := k.GetIssuanceRecord(ctx, k.GetNextIssuanceRecordID(ctx)-1)
	isDownRound := record.IsDownRound

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"issue_shares_with_protection",
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", msg.CompanyID)),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("recipient", msg.Recipient),
			sdk.NewAttribute("shares", msg.Shares.String()),
			sdk.NewAttribute("price", msg.Price.String()),
			sdk.NewAttribute("is_down_round", fmt.Sprintf("%d", btoi(isDownRound))),
			sdk.NewAttribute("adjustments_count", fmt.Sprintf("%d", len(adjustments))),
			sdk.NewAttribute("purpose", msg.Purpose),
		),
	)

	return &types.MsgIssueSharesWithProtectionResponse{
		SharesIssued:     msg.Shares,
		TotalCost:        totalCost,
		IsDownRound:      isDownRound,
		AdjustmentsCount: uint64(len(adjustments)),
	}, nil
}

// =============================================================================
// Company Authorization Message Handlers
// =============================================================================

// AddAuthorizedProposer adds a new authorized proposer to a company
func (k msgServer) AddAuthorizedProposer(goCtx context.Context, msg *types.MsgAddAuthorizedProposer) (*types.MsgAddAuthorizedProposerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Add proposer
	if err := k.Keeper.AddAuthorizedProposer(ctx, msg.CompanyID, msg.Creator, msg.NewProposer); err != nil {
		return nil, err
	}

	return &types.MsgAddAuthorizedProposerResponse{
		Success: true,
	}, nil
}

// RemoveAuthorizedProposer removes an authorized proposer from a company
func (k msgServer) RemoveAuthorizedProposer(goCtx context.Context, msg *types.MsgRemoveAuthorizedProposer) (*types.MsgRemoveAuthorizedProposerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Remove proposer
	if err := k.Keeper.RemoveAuthorizedProposer(ctx, msg.CompanyID, msg.Creator, msg.ProposerToRemove); err != nil {
		return nil, err
	}

	return &types.MsgRemoveAuthorizedProposerResponse{
		Success: true,
	}, nil
}

// InitiateOwnershipTransfer initiates a two-step ownership transfer
func (k msgServer) InitiateOwnershipTransfer(goCtx context.Context, msg *types.MsgInitiateOwnershipTransfer) (*types.MsgInitiateOwnershipTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Initiate transfer
	if err := k.Keeper.InitiateOwnershipTransfer(ctx, msg.CompanyID, msg.Creator, msg.NewOwner); err != nil {
		return nil, err
	}

	return &types.MsgInitiateOwnershipTransferResponse{
		Success: true,
	}, nil
}

// AcceptOwnershipTransfer accepts a pending ownership transfer
func (k msgServer) AcceptOwnershipTransfer(goCtx context.Context, msg *types.MsgAcceptOwnershipTransfer) (*types.MsgAcceptOwnershipTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Accept transfer
	if err := k.Keeper.AcceptOwnershipTransfer(ctx, msg.CompanyID, msg.Creator); err != nil {
		return nil, err
	}

	return &types.MsgAcceptOwnershipTransferResponse{
		Success:  true,
		NewOwner: msg.Creator,
	}, nil
}

// CancelOwnershipTransfer cancels a pending ownership transfer
func (k msgServer) CancelOwnershipTransfer(goCtx context.Context, msg *types.MsgCancelOwnershipTransfer) (*types.MsgCancelOwnershipTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Cancel transfer
	if err := k.Keeper.CancelOwnershipTransfer(ctx, msg.CompanyID, msg.Creator); err != nil {
		return nil, err
	}

	return &types.MsgCancelOwnershipTransferResponse{
		Success: true,
	}, nil
}
// =============================================================================
// Shareholder Blacklist Message Handlers
// =============================================================================

// BlacklistShareholder adds a shareholder to the blacklist
func (k msgServer) BlacklistShareholder(goCtx context.Context, msg *types.SimpleMsgBlacklistShareholder) (*types.MsgBlacklistShareholderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	// Get company to verify authority
	company, found := k.Keeper.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Only company founder can blacklist shareholders (could add governance later)
	if company.Founder != msg.Authority {
		return nil, types.ErrUnauthorized
	}

	// Convert duration from seconds to time.Duration
	var duration time.Duration
	if msg.Duration > 0 {
		duration = time.Duration(msg.Duration) * time.Second
	}

	// Blacklist the shareholder
	if err := k.Keeper.BlacklistShareholder(
		ctx,
		msg.CompanyID,
		msg.Address,
		msg.Reason,
		msg.Authority,
		duration,
	); err != nil {
		return nil, err
	}

	return &types.MsgBlacklistShareholderResponse{
		Success: true,
	}, nil
}

// UnblacklistShareholder removes a shareholder from the blacklist
func (k msgServer) UnblacklistShareholder(goCtx context.Context, msg *types.SimpleMsgUnblacklistShareholder) (*types.MsgUnblacklistShareholderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	// Get company to verify authority
	company, found := k.Keeper.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Only company founder can unblacklist shareholders
	if company.Founder != msg.Authority {
		return nil, types.ErrUnauthorized
	}

	// Unblacklist the shareholder
	if err := k.Keeper.UnblacklistShareholder(ctx, msg.CompanyID, msg.Address); err != nil {
		return nil, err
	}

	return &types.MsgUnblacklistShareholderResponse{
		Success: true,
	}, nil
}

// SetDividendRedirection configures where blocked dividends should be redirected
func (k msgServer) SetDividendRedirection(goCtx context.Context, msg *types.SimpleMsgSetDividendRedirection) (*types.MsgSetDividendRedirectionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	// Get company to verify authority
	company, found := k.Keeper.getCompany(ctx, msg.CompanyID)
	if !found {
		return nil, types.ErrCompanyNotFound
	}

	// Only company founder can configure redirection
	if company.Founder != msg.Authority {
		return nil, types.ErrUnauthorized
	}

	// Create redirection config
	redirect := types.NewDividendRedirection(
		msg.CompanyID,
		msg.CharityWallet,
		msg.FallbackAction,
		msg.Description,
		msg.Authority,
	)

	// Set dividend redirection
	if err := k.Keeper.SetDividendRedirection(ctx, redirect); err != nil {
		return nil, err
	}

	return &types.MsgSetDividendRedirectionResponse{
		Success: true,
	}, nil
}

// SetCharityWallet sets the protocol-level default charity wallet (governance only)
func (k msgServer) SetCharityWallet(goCtx context.Context, msg *types.SimpleMsgSetCharityWallet) (*types.MsgSetCharityWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	// SECURITY: Verify caller is governance module authority
	// Only governance can set default charity wallet
	if msg.Authority != k.GetAuthority() {
		return nil, errors.Wrapf(types.ErrUnauthorized, "only governance module can set charity wallet (expected %s, got %s)", k.GetAuthority(), msg.Authority)
	}

	// Set default charity wallet
	if err := k.Keeper.SetDefaultCharityWallet(ctx, msg.Wallet); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"set_default_charity_wallet",
			sdk.NewAttribute("wallet", msg.Wallet),
			sdk.NewAttribute("authority", msg.Authority),
		),
	)

	return &types.MsgSetCharityWalletResponse{
		Success: true,
	}, nil
}

// =============================================================================
// Treasury Deposit Message Handlers
// =============================================================================

// DepositToTreasury allows founder/co-founders to deposit HODL into treasury
func (k msgServer) DepositToTreasury(goCtx context.Context, msg *types.SimpleMsgDepositToTreasury) (*types.MsgDepositToTreasuryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.Keeper.DepositToTreasuryDirect(ctx, msg.CompanyID, msg.Depositor, msg.Amount); err != nil {
		return nil, err
	}

	return &types.MsgDepositToTreasuryResponse{
		Success: true,
	}, nil
}

// =============================================================================
// Dividend Audit Message Handlers
// =============================================================================

// VerifyAudit handles validator verification of dividend audits
func (k msgServer) VerifyAudit(goCtx context.Context, msg *types.MsgVerifyAudit) (*types.MsgVerifyAuditResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Verify the audit
	err := k.Keeper.VerifyAudit(
		ctx,
		msg.AuditID,
		msg.Verifier,
		msg.Approved,
		msg.Comments,
	)
	if err != nil {
		return nil, err
	}

	// Get updated audit to return status
	audit, found := k.GetDividendAudit(ctx, msg.AuditID)
	if !found {
		return nil, types.ErrAuditNotFound
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"verify_audit",
			sdk.NewAttribute("audit_id", fmt.Sprintf("%d", msg.AuditID)),
			sdk.NewAttribute("verifier", msg.Verifier),
			sdk.NewAttribute("approved", fmt.Sprintf("%t", msg.Approved)),
			sdk.NewAttribute("status", audit.Status.String()),
		),
	)

	return &types.MsgVerifyAuditResponse{
		Success: true,
		Status:  audit.Status.String(),
	}, nil
}

// DisputeAudit handles shareholder disputes of dividend audits
func (k msgServer) DisputeAudit(goCtx context.Context, msg *types.MsgDisputeAudit) (*types.MsgDisputeAuditResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Get audit to verify shareholder status
	audit, found := k.GetDividendAudit(ctx, msg.AuditID)
	if !found {
		return nil, types.ErrAuditNotFound
	}

	// Verify disputer is actually a shareholder of the company
	shareholding, found := k.getShareholding(ctx, audit.CompanyID, "COMMON", msg.Disputer)
	if !found || shareholding.Shares.IsZero() {
		return nil, types.ErrUnauthorized.Wrap("only shareholders can dispute audits")
	}

	// Dispute the audit
	err := k.Keeper.DisputeAudit(ctx, msg.AuditID, msg.Reason)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"dispute_audit",
			sdk.NewAttribute("audit_id", fmt.Sprintf("%d", msg.AuditID)),
			sdk.NewAttribute("disputer", msg.Disputer),
			sdk.NewAttribute("reason", msg.Reason),
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", audit.CompanyID)),
		),
	)

	return &types.MsgDisputeAuditResponse{
		Success: true,
	}, nil
}

// =============================================================================
// Shareholder Petition Handlers
// =============================================================================

// CreatePetition handles creating a new shareholder petition
func (k msgServer) CreatePetition(goCtx context.Context, msg *types.SimpleMsgCreatePetition) (*types.MsgCreatePetitionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Create petition
	petitionID, err := k.Keeper.CreatePetition(
		ctx,
		msg.Creator,
		msg.CompanyID,
		msg.ClassID,
		msg.PetitionType,
		msg.Title,
		msg.Description,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreatePetitionResponse{
		PetitionID: petitionID,
		Success:    true,
	}, nil
}

// SignPetition handles signing/supporting a petition
func (k msgServer) SignPetition(goCtx context.Context, msg *types.SimpleMsgSignPetition) (*types.MsgSignPetitionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Sign petition
	if err := k.Keeper.SignPetition(ctx, msg.PetitionID, msg.Signer, msg.Comment); err != nil {
		return nil, err
	}

	// Get updated petition to return signature count
	petition, found := k.Keeper.GetPetition(ctx, msg.PetitionID)
	if !found {
		return nil, types.ErrPetitionNotFound
	}

	return &types.MsgSignPetitionResponse{
		Success:        true,
		SignatureCount: petition.SignatureCount,
	}, nil
}

// WithdrawPetition handles withdrawing a petition
func (k msgServer) WithdrawPetition(goCtx context.Context, msg *types.SimpleMsgWithdrawPetition) (*types.MsgWithdrawPetitionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate basic message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Withdraw petition
	if err := k.Keeper.WithdrawPetition(ctx, msg.PetitionID, msg.Creator); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawPetitionResponse{
		Success: true,
	}, nil
}
