package keeper

import (
	"context"
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

	// Issue initial shares to founder (typically 100% initially)
	if err := k.Keeper.IssueShares(ctx, companyID, "COMMON", msg.Creator, msg.TotalShares, msg.ParValue, "hodl"); err != nil {
		return nil, errors.Wrapf(err, "failed to issue initial shares to founder")
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"create_company",
			sdk.NewAttribute("company_id", string(rune(companyID))),
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
	company, found := k.GetCompany(ctx, msg.CompanyID)
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
			sdk.NewAttribute("company_id", string(rune(msg.CompanyID))),
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
	company, found := k.GetCompany(ctx, msg.CompanyID)
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
			sdk.NewAttribute("company_id", string(rune(msg.CompanyID))),
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
	company, found := k.GetCompany(ctx, msg.CompanyID)
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
			sdk.NewAttribute("company_id", string(rune(msg.CompanyID))),
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
	company, found := k.GetCompany(ctx, msg.CompanyID)
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
			sdk.NewAttribute("company_id", string(rune(msg.CompanyID))),
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
			sdk.NewAttribute("listing_id", string(rune(listingID))),
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
			sdk.NewAttribute("listing_id", string(rune(msg.ListingID))),
			sdk.NewAttribute("validator", msg.Validator),
			sdk.NewAttribute("approved", string(rune(btoi(msg.Approved)))),
			sdk.NewAttribute("score", string(rune(msg.Score))),
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
			sdk.NewAttribute("listing_id", string(rune(msg.ListingID))),
			sdk.NewAttribute("company_id", string(rune(companyID))),
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
			sdk.NewAttribute("listing_id", string(rune(msg.ListingID))),
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
			sdk.NewAttribute("listing_id", string(rune(msg.ListingID))),
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

	// Update listing requirements
	k.SetListingRequirements(ctx, msg.Requirements)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"update_listing_requirements",
			sdk.NewAttribute("authority", msg.Authority),
			sdk.NewAttribute("minimum_listing_fee", msg.Requirements.MinimumListingFee.String()),
			sdk.NewAttribute("required_validators", string(rune(msg.Requirements.RequiredValidators))),
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

// DeclareDividend handles dividend declaration
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

	// Declare dividend
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
	)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"declare_dividend",
			sdk.NewAttribute("dividend_id", string(rune(dividendID))),
			sdk.NewAttribute("company_id", string(rune(msg.CompanyID))),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("dividend_type", msg.DividendType),
			sdk.NewAttribute("amount_per_share", msg.AmountPerShare.String()),
			sdk.NewAttribute("currency", msg.Currency),
			sdk.NewAttribute("creator", msg.Creator),
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
			sdk.NewAttribute("dividend_id", string(rune(msg.DividendID))),
			sdk.NewAttribute("payments_processed", string(rune(paymentsProcessed))),
			sdk.NewAttribute("amount_paid", amountPaid.String()),
			sdk.NewAttribute("is_complete", string(rune(btoi(isComplete)))),
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
			sdk.NewAttribute("dividend_id", string(rune(msg.DividendID))),
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
	company, found := k.GetCompany(ctx, dividend.CompanyID)
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

	// Cancel dividend
	dividend.Status = types.DividendStatusCancelled
	dividend.UpdatedAt = ctx.BlockTime()
	dividend.Notes = msg.Reason
	k.SetDividend(ctx, dividend)

	// If cash dividend, refund remaining funds to creator
	if dividend.Type == types.DividendTypeCash && dividend.RemainingAmount.GT(math.LegacyZeroDec()) {
		creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
		if err == nil {
			refundCoins := sdk.NewCoins(sdk.NewCoin(dividend.Currency, dividend.RemainingAmount.TruncateInt()))
			k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, creatorAddr, refundCoins)
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"cancel_dividend",
			sdk.NewAttribute("dividend_id", string(rune(msg.DividendID))),
			sdk.NewAttribute("company_id", string(rune(dividend.CompanyID))),
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
	company, found := k.GetCompany(ctx, msg.CompanyID)
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
			sdk.NewAttribute("company_id", string(rune(msg.CompanyID))),
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
			sdk.NewAttribute("dividend_id", string(rune(msg.DividendID))),
			sdk.NewAttribute("total_shareholders", string(rune(snapshot.TotalShareholders))),
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
	company, found := k.GetCompany(ctx, msg.CompanyID)
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
			CompanyID:       msg.CompanyID,
			Policy:          "variable",
			PaymentFrequency: "quarterly",
			PaymentMethod:   "automatic",
			EffectiveDate:   ctx.BlockTime(),
			LastUpdated:     ctx.BlockTime(),
			UpdatedBy:       msg.Creator,
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
			sdk.NewAttribute("company_id", string(rune(msg.CompanyID))),
			sdk.NewAttribute("class_id", msg.ClassID),
			sdk.NewAttribute("enabled", string(rune(btoi(msg.Enabled)))),
			sdk.NewAttribute("creator", msg.Creator),
		),
	)

	return &types.MsgEnableDividendReinvestmentResponse{
		Success: true,
	}, nil
}