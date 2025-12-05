package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
	equitytypes "github.com/sharehodl/sharehodl-blockchain/x/equity/types"
	"fmt"
	"time"
)

// SubmitCompanyProposal submits a company-specific governance proposal
func (k Keeper) SubmitCompanyProposal(
	ctx sdk.Context,
	submitter string,
	companyID uint64,
	proposalType types.CompanyProposalType,
	title string,
	description string,
	proposalData map[string]interface{},
	votingPeriod time.Duration,
	quorumRequirement math.LegacyDec,
	thresholdRequirement math.LegacyDec,
) (uint64, error) {
	submitterAddr, err := sdk.AccAddressFromBech32(submitter)
	if err != nil {
		return 0, types.ErrInvalidAddress
	}

	// Validate company exists
	company, found := k.equityKeeper.GetCompany(ctx, companyID)
	if !found {
		return 0, types.ErrCompanyNotFound
	}

	// Validate submitter eligibility
	if err := k.validateCompanyProposalSubmission(ctx, submitterAddr, company, proposalType); err != nil {
		return 0, err
	}

	// Get next proposal ID
	proposalID := k.getNextProposalID(ctx)

	// Determine voting period and requirements based on proposal type
	if votingPeriod == 0 {
		votingPeriod = k.getDefaultVotingPeriodForType(proposalType)
	}

	if quorumRequirement.IsZero() {
		quorumRequirement = k.getDefaultQuorumForType(proposalType)
	}

	if thresholdRequirement.IsZero() {
		thresholdRequirement = k.getDefaultThresholdForType(proposalType)
	}

	// Create company governance proposal
	companyProposal := types.CompanyGovernanceProposal{
		ProposalID:        proposalID,
		CompanyID:         companyID,
		CompanySymbol:     company.Symbol,
		Type:              proposalType,
		Title:             title,
		Description:       description,
		Proposer:          submitter,
		ShareClassVoting:  k.getShareClassVotingWeights(ctx, company),
		RequiredApproval:  thresholdRequirement,
		BoardApprovalReq:  k.requiresBoardApproval(proposalType),
		QuorumRequirement: quorumRequirement,
		VotingPeriod:      votingPeriod,
		ExecutionDelay:    k.getExecutionDelay(proposalType),
		ProposalData:      proposalData,
		CreatedAt:         ctx.BlockTime(),
		UpdatedAt:         ctx.BlockTime(),
	}

	// Create main governance proposal
	proposal := types.Proposal{
		ID:                proposalID,
		Type:              types.ProposalTypeCompanyGovernance,
		Title:             fmt.Sprintf("[%s] %s", company.Symbol, title),
		Description:       description,
		Proposer:          submitter,
		Status:            types.ProposalStatusVotingPeriod, // Skip deposit period for company proposals
		VotingStartTime:   ctx.BlockTime(),
		VotingEndTime:     ctx.BlockTime().Add(votingPeriod),
		DepositEndTime:    ctx.BlockTime(),
		TotalDeposit:      math.ZeroInt(),
		MinDeposit:        math.ZeroInt(), // No deposit required for company proposals
		YesVotes:          math.ZeroInt(),
		NoVotes:           math.ZeroInt(),
		AbstainVotes:      math.ZeroInt(),
		NoWithVetoVotes:   math.ZeroInt(),
		TotalVotingPower:  math.ZeroInt(),
		Quorum:            quorumRequirement,
		Threshold:         thresholdRequirement,
		VetoThreshold:     math.LegacyNewDecWithPrec(334, 3), // 33.4%
		CompanyID:         companyID,
		CreatedAt:         ctx.BlockTime(),
		UpdatedAt:         ctx.BlockTime(),
	}

	// Store both proposals
	k.setProposal(ctx, proposal)
	k.setCompanyProposal(ctx, companyProposal)

	// Index proposals
	k.indexProposal(ctx, proposal)
	k.indexCompanyProposal(ctx, companyProposal)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"submit_company_proposal",
			sdk.NewAttribute("proposal_id", sdk.Uint64ToBigEndian(proposalID).String()),
			sdk.NewAttribute("company_id", sdk.Uint64ToBigEndian(companyID).String()),
			sdk.NewAttribute("company_symbol", company.Symbol),
			sdk.NewAttribute("proposal_type", proposalType.String()),
			sdk.NewAttribute("submitter", submitter),
			sdk.NewAttribute("title", title),
		),
	)

	return proposalID, nil
}

// VoteOnCompanyProposal allows shareholders to vote on company proposals
func (k Keeper) VoteOnCompanyProposal(
	ctx sdk.Context,
	voter string,
	proposalID uint64,
	option types.VoteOption,
	reason string,
) error {
	voterAddr, err := sdk.AccAddressFromBech32(voter)
	if err != nil {
		return types.ErrInvalidAddress
	}

	// Get proposals
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	companyProposal, found := k.GetCompanyProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Validate voting conditions
	if err := k.validateCompanyVoting(ctx, proposal, companyProposal, voterAddr); err != nil {
		return err
	}

	// Check if already voted
	if k.hasVoted(ctx, proposalID, voterAddr) {
		return types.ErrAlreadyVoted
	}

	// Calculate shareholder voting power
	votingPower, err := k.calculateShareholderVotingPower(ctx, companyProposal, voterAddr)
	if err != nil {
		return err
	}

	if votingPower.IsZero() {
		return types.ErrInsufficientVotingPower
	}

	// Create vote
	vote := types.Vote{
		ProposalID:  proposalID,
		Voter:       voter,
		Option:      option,
		VotingPower: votingPower,
		Weight:      math.LegacyOneDec(),
		CompanyID:   companyProposal.CompanyID,
		Justification: reason,
		VotedAt:     ctx.BlockTime(),
	}

	// Store vote
	k.setVote(ctx, vote)

	// Update tally
	k.updateCompanyProposalTally(ctx, proposal, companyProposal, vote)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote_company_proposal",
			sdk.NewAttribute("proposal_id", sdk.Uint64ToBigEndian(proposalID).String()),
			sdk.NewAttribute("company_id", sdk.Uint64ToBigEndian(companyProposal.CompanyID).String()),
			sdk.NewAttribute("voter", voter),
			sdk.NewAttribute("option", option.String()),
			sdk.NewAttribute("voting_power", votingPower.String()),
		),
	)

	return nil
}

// validateCompanyProposalSubmission validates if submitter can submit company proposal
func (k Keeper) validateCompanyProposalSubmission(
	ctx sdk.Context,
	submitter sdk.AccAddress,
	company equitytypes.Company,
	proposalType types.CompanyProposalType,
) error {
	// Check if submitter is a shareholder
	shareholding, found := k.equityKeeper.GetShareholding(ctx, company.ID, submitter.String())
	if !found {
		return types.ErrNotAuthorized
	}

	// Calculate ownership percentage
	totalShares := k.equityKeeper.GetTotalShares(ctx, company.ID, shareholding.ClassID)
	if totalShares.IsZero() {
		return types.ErrInvalidState
	}

	ownershipPercent := math.LegacyNewDecFromInt(shareholding.Shares).Quo(math.LegacyNewDecFromInt(totalShares))

	// Check minimum ownership requirements based on proposal type
	minOwnership := k.getMinOwnershipForProposalType(proposalType)
	if ownershipPercent.LT(minOwnership) {
		return types.ErrInsufficientPermissions
	}

	// Special checks for sensitive proposal types
	switch proposalType {
	case types.CompanyProposalTypeMergerAcquisition:
		// Requires board member or 10%+ ownership
		if !k.isBoardMember(ctx, company, submitter) && ownershipPercent.LT(math.LegacyNewDecWithPrec(1, 1)) {
			return types.ErrInsufficientPermissions
		}
		
	case types.CompanyProposalTypeExecutiveComp:
		// Requires independent board member or 5%+ ownership
		if !k.isIndependentBoardMember(ctx, company, submitter) && ownershipPercent.LT(math.LegacyNewDecWithPrec(5, 2)) {
			return types.ErrInsufficientPermissions
		}
	}

	return nil
}

// validateCompanyVoting validates voting eligibility for company proposals
func (k Keeper) validateCompanyVoting(
	ctx sdk.Context,
	proposal types.Proposal,
	companyProposal types.CompanyGovernanceProposal,
	voter sdk.AccAddress,
) error {
	// Check proposal status
	if proposal.Status != types.ProposalStatusVotingPeriod {
		return types.ErrProposalNotActive
	}

	// Check voting period
	if ctx.BlockTime().Before(proposal.VotingStartTime) || ctx.BlockTime().After(proposal.VotingEndTime) {
		return types.ErrVotingPeriodEnded
	}

	// Check if voter is a shareholder
	_, found := k.equityKeeper.GetShareholding(ctx, companyProposal.CompanyID, voter.String())
	if !found {
		return types.ErrInsufficientVotingPower
	}

	return nil
}

// calculateShareholderVotingPower calculates voting power for shareholders
func (k Keeper) calculateShareholderVotingPower(
	ctx sdk.Context,
	companyProposal types.CompanyGovernanceProposal,
	voter sdk.AccAddress,
) (math.LegacyDec, error) {
	// Get shareholder's holdings
	shareholding, found := k.equityKeeper.GetShareholding(ctx, companyProposal.CompanyID, voter.String())
	if !found {
		return math.LegacyZeroDec(), nil
	}

	// Get voting weight for this share class
	classWeight, exists := companyProposal.ShareClassVoting[shareholding.ClassID]
	if !exists {
		classWeight = math.LegacyOneDec() // Default voting weight
	}

	// Calculate voting power based on shares and class weight
	votingPower := math.LegacyNewDecFromInt(shareholding.Shares).Mul(classWeight)

	// Apply any voting caps or restrictions
	maxVotingPower := k.getMaxVotingPowerPerShareholder(ctx, companyProposal)
	if votingPower.GT(maxVotingPower) {
		votingPower = maxVotingPower
	}

	return votingPower, nil
}

// updateCompanyProposalTally updates vote tallies for company proposals
func (k Keeper) updateCompanyProposalTally(
	ctx sdk.Context,
	proposal types.Proposal,
	companyProposal types.CompanyGovernanceProposal,
	vote types.Vote,
) {
	// Update main proposal tally
	k.updateTallyResult(ctx, proposal, vote, vote.Weight)

	// Update company-specific tally metrics
	k.updateCompanyTallyMetrics(ctx, companyProposal, vote)
}

// Helper functions

func (k Keeper) getDefaultVotingPeriodForType(proposalType types.CompanyProposalType) time.Duration {
	switch proposalType {
	case types.CompanyProposalTypeBoardElection:
		return time.Hour * 24 * 14 // 2 weeks
	case types.CompanyProposalTypeMergerAcquisition:
		return time.Hour * 24 * 30 // 30 days
	case types.CompanyProposalTypeBylaw:
		return time.Hour * 24 * 21 // 3 weeks
	case types.CompanyProposalTypeExecutiveComp:
		return time.Hour * 24 * 10 // 10 days
	default:
		return time.Hour * 24 * 7 // 1 week default
	}
}

func (k Keeper) getDefaultQuorumForType(proposalType types.CompanyProposalType) math.LegacyDec {
	switch proposalType {
	case types.CompanyProposalTypeMergerAcquisition:
		return math.LegacyNewDecWithPrec(67, 2) // 67%
	case types.CompanyProposalTypeBylaw:
		return math.LegacyNewDecWithPrec(6, 1)  // 60%
	case types.CompanyProposalTypeExecutiveComp:
		return math.LegacyNewDecWithPrec(5, 1)  // 50%
	default:
		return math.LegacyNewDecWithPrec(334, 3) // 33.4%
	}
}

func (k Keeper) getDefaultThresholdForType(proposalType types.CompanyProposalType) math.LegacyDec {
	switch proposalType {
	case types.CompanyProposalTypeMergerAcquisition:
		return math.LegacyNewDecWithPrec(75, 2) // 75%
	case types.CompanyProposalTypeBylaw:
		return math.LegacyNewDecWithPrec(67, 2) // 67%
	case types.CompanyProposalTypeShareBuyback:
		return math.LegacyNewDecWithPrec(67, 2) // 67%
	default:
		return math.LegacyNewDecWithPrec(5, 1) // 50%
	}
}

func (k Keeper) getMinOwnershipForProposalType(proposalType types.CompanyProposalType) math.LegacyDec {
	switch proposalType {
	case types.CompanyProposalTypeBoardElection:
		return math.LegacyNewDecWithPrec(1, 2) // 1%
	case types.CompanyProposalTypeMergerAcquisition:
		return math.LegacyNewDecWithPrec(5, 2) // 5%
	case types.CompanyProposalTypeExecutiveComp:
		return math.LegacyNewDecWithPrec(3, 2) // 3%
	default:
		return math.LegacyNewDecWithPrec(5, 3) // 0.5%
	}
}

func (k Keeper) requiresBoardApproval(proposalType types.CompanyProposalType) bool {
	switch proposalType {
	case types.CompanyProposalTypeMergerAcquisition,
		 types.CompanyProposalTypeCapitalStructure,
		 types.CompanyProposalTypeAssetSale:
		return true
	default:
		return false
	}
}

func (k Keeper) getExecutionDelay(proposalType types.CompanyProposalType) time.Duration {
	switch proposalType {
	case types.CompanyProposalTypeMergerAcquisition:
		return time.Hour * 24 * 7 // 1 week delay
	case types.CompanyProposalTypeBylaw:
		return time.Hour * 24 * 3 // 3 days delay
	default:
		return time.Hour * 24 // 1 day delay
	}
}

func (k Keeper) getShareClassVotingWeights(ctx sdk.Context, company equitytypes.Company) map[string]math.LegacyDec {
	weights := make(map[string]math.LegacyDec)
	
	// Default implementation - would be enhanced based on actual share class structure
	weights["common"] = math.LegacyOneDec()
	weights["preferred"] = math.LegacyNewDec(2) // Preferred shares get 2x voting power
	
	return weights
}

func (k Keeper) getMaxVotingPowerPerShareholder(ctx sdk.Context, companyProposal types.CompanyGovernanceProposal) math.LegacyDec {
	// Default cap at 25% to prevent single shareholder dominance
	return math.LegacyNewDecWithPrec(25, 2)
}

func (k Keeper) isBoardMember(ctx sdk.Context, company equitytypes.Company, addr sdk.AccAddress) bool {
	// Implementation would check if address is a board member
	// This would integrate with equity module's board tracking
	return false
}

func (k Keeper) isIndependentBoardMember(ctx sdk.Context, company equitytypes.Company, addr sdk.AccAddress) bool {
	// Implementation would check if address is an independent board member
	return false
}

// Storage functions for company proposals

func (k Keeper) setCompanyProposal(ctx sdk.Context, proposal types.CompanyGovernanceProposal) {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	key := types.CompanyProposalKey(proposal.ProposalID)
	value := k.cdc.MustMarshal(&proposal)
	store.Set(key, value)
}

func (k Keeper) GetCompanyProposal(ctx sdk.Context, proposalID uint64) (types.CompanyGovernanceProposal, bool) {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	key := types.CompanyProposalKey(proposalID)
	
	value := store.Get(key)
	if value == nil {
		return types.CompanyGovernanceProposal{}, false
	}

	var proposal types.CompanyGovernanceProposal
	k.cdc.MustUnmarshal(value, &proposal)
	return proposal, true
}

func (k Keeper) indexCompanyProposal(ctx sdk.Context, proposal types.CompanyGovernanceProposal) {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	
	// Index by company
	companyKey := types.CompanyProposalIndexKey(proposal.CompanyID)
	companyIndexKey := append(companyKey, sdk.Uint64ToBigEndian(proposal.ProposalID)...)
	store.Set(companyIndexKey, []byte{})
	
	// Index by type
	typeKey := types.CompanyProposalTypeIndexKey(proposal.Type)
	typeIndexKey := append(typeKey, sdk.Uint64ToBigEndian(proposal.ProposalID)...)
	store.Set(typeIndexKey, []byte{})
}

func (k Keeper) updateCompanyTallyMetrics(ctx sdk.Context, proposal types.CompanyGovernanceProposal, vote types.Vote) {
	// Implementation would track company-specific voting metrics
	// This could include tracking by share class, insider vs. outsider votes, etc.
}