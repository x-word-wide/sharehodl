package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	equitytypes "github.com/sharehodl/sharehodl-blockchain/x/equity/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
)

// companyInfo helper struct to extract company info from interface{}
type companyInfo struct {
	ID     uint64
	Symbol string
}

// extractCompanyInfo extracts company info from interface{}
func extractCompanyInfo(company interface{}) companyInfo {
	if c, ok := company.(equitytypes.Company); ok {
		return companyInfo{ID: c.ID, Symbol: c.Symbol}
	}
	// Handle map-like interface for JSON decoded data
	if m, ok := company.(map[string]interface{}); ok {
		info := companyInfo{}
		if id, ok := m["id"].(uint64); ok {
			info.ID = id
		}
		if sym, ok := m["symbol"].(string); ok {
			info.Symbol = sym
		}
		return info
	}
	return companyInfo{}
}

// shareholdingInfo helper struct to extract shareholding info from interface{}
type shareholdingInfo struct {
	ClassID string
	Shares  math.Int
}

// extractShareholdingInfo extracts shareholding info from interface{}
func extractShareholdingInfo(shareholding interface{}) shareholdingInfo {
	if s, ok := shareholding.(equitytypes.Shareholding); ok {
		return shareholdingInfo{ClassID: s.ClassID, Shares: s.Shares}
	}
	return shareholdingInfo{ClassID: "common", Shares: math.ZeroInt()}
}

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

	// Extract company info
	compInfo := extractCompanyInfo(company)

	// Validate submitter eligibility
	if err := k.validateCompanyProposalSubmission(ctx, submitterAddr, compInfo, proposalType); err != nil {
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
		CompanySymbol:     compInfo.Symbol,
		Type:              proposalType,
		Title:             title,
		Description:       description,
		Proposer:          submitter,
		ShareClassVoting:  k.getShareClassVotingWeights(ctx, compInfo),
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
		Title:             fmt.Sprintf("[%s] %s", compInfo.Symbol, title),
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
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute("company_symbol", compInfo.Symbol),
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
		VotingPower: votingPower.TruncateInt(),
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
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyProposal.CompanyID)),
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
	compInfo companyInfo,
	proposalType types.CompanyProposalType,
) error {
	// Check if submitter is a shareholder
	shareholdingIface, found := k.equityKeeper.GetShareholding(ctx, compInfo.ID, "common", submitter.String())
	if !found {
		return types.ErrNotAuthorized
	}

	sh := extractShareholdingInfo(shareholdingIface)

	// Calculate ownership percentage
	totalShares := k.equityKeeper.GetTotalShares(ctx, compInfo.ID, sh.ClassID)
	if totalShares.IsZero() {
		return types.ErrInvalidState
	}

	ownershipPercent := math.LegacyNewDecFromInt(sh.Shares).Quo(math.LegacyNewDecFromInt(totalShares))

	// Check minimum ownership requirements based on proposal type
	minOwnership := k.getMinOwnershipForProposalType(proposalType)
	if ownershipPercent.LT(minOwnership) {
		return types.ErrInsufficientPermissions
	}

	// Special checks for sensitive proposal types
	switch proposalType {
	case types.CompanyProposalTypeMergerAcquisition:
		// Requires board member or 10%+ ownership
		if !k.isBoardMember(ctx, compInfo, submitter) && ownershipPercent.LT(math.LegacyNewDecWithPrec(1, 1)) {
			return types.ErrInsufficientPermissions
		}

	case types.CompanyProposalTypeExecutiveComp:
		// Requires independent board member or 5%+ ownership
		if !k.isIndependentBoardMember(ctx, compInfo, submitter) && ownershipPercent.LT(math.LegacyNewDecWithPrec(5, 2)) {
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

	// Check if voter is a shareholder OR has beneficial ownership (LP positions)
	hasDirectShares := false
	hasBeneficialOwnership := false

	// 1. Check direct shareholding
	_, found := k.equityKeeper.GetShareholding(ctx, companyProposal.CompanyID, "common", voter.String())
	if found {
		hasDirectShares = true
	}

	// 2. Check beneficial ownership (shares in LP pools, escrow, etc.)
	if !hasDirectShares {
		ownerships, err := k.equityKeeper.GetBeneficialOwnershipsByOwner(ctx, companyProposal.CompanyID, "common", voter.String())
		if err == nil && len(ownerships) > 0 {
			hasBeneficialOwnership = true
		}
	}

	// Voter must have either direct shares or beneficial ownership
	if !hasDirectShares && !hasBeneficialOwnership {
		return types.ErrInsufficientVotingPower
	}

	return nil
}

// calculateShareholderVotingPower calculates voting power for shareholders
// This includes BOTH direct holdings AND beneficial ownership (LP positions, escrow, etc.)
// Equity holders can vote even when their shares are in liquidity pools or escrow
func (k Keeper) calculateShareholderVotingPower(
	ctx sdk.Context,
	companyProposal types.CompanyGovernanceProposal,
	voter sdk.AccAddress,
) (math.LegacyDec, error) {
	totalShares := math.ZeroInt()
	classID := "common"

	// 1. Get shareholder's DIRECT holdings
	shareholdingIface, found := k.equityKeeper.GetShareholding(ctx, companyProposal.CompanyID, classID, voter.String())
	if found {
		sh := extractShareholdingInfo(shareholdingIface)
		totalShares = totalShares.Add(sh.Shares)
		if sh.ClassID != "" {
			classID = sh.ClassID
		}
	}

	// 2. Get shareholder's BENEFICIAL OWNERSHIP (shares in LP pools, escrow, etc.)
	// This ensures LP providers can still vote on company matters
	beneficialOwnerships, err := k.equityKeeper.GetBeneficialOwnershipsByOwner(ctx, companyProposal.CompanyID, classID, voter.String())
	if err == nil {
		for _, ownership := range beneficialOwnerships {
			totalShares = totalShares.Add(ownership.Shares)
		}
	}

	// If no shares found (direct or beneficial), return zero voting power
	if totalShares.IsZero() {
		return math.LegacyZeroDec(), nil
	}

	// Get voting weight for this share class
	classWeight, exists := companyProposal.ShareClassVoting[classID]
	if !exists {
		classWeight = math.LegacyOneDec() // Default voting weight
	}

	// Calculate voting power based on TOTAL shares (direct + beneficial) and class weight
	votingPower := math.LegacyNewDecFromInt(totalShares).Mul(classWeight)

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

func (k Keeper) getShareClassVotingWeights(ctx sdk.Context, compInfo companyInfo) map[string]math.LegacyDec {
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

func (k Keeper) isBoardMember(ctx sdk.Context, compInfo companyInfo, addr sdk.AccAddress) bool {
	// Implementation would check if address is a board member
	// This would integrate with equity module's board tracking
	_ = compInfo // avoid unused warning
	return false
}

func (k Keeper) isIndependentBoardMember(ctx sdk.Context, compInfo companyInfo, addr sdk.AccAddress) bool {
	// Implementation would check if address is an independent board member
	_ = compInfo // avoid unused warning
	return false
}

// Storage functions for company proposals

func (k Keeper) setCompanyProposal(ctx sdk.Context, proposal types.CompanyGovernanceProposal) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	key := types.CompanyProposalKey(proposal.CompanyID, proposal.ProposalID)
	value, _ := json.Marshal(proposal)
	store.Set(key, value)
}

func (k Keeper) GetCompanyProposal(ctx sdk.Context, proposalID uint64) (types.CompanyGovernanceProposal, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))

	// We need to iterate to find by proposalID since we don't have companyID
	iterator := store.Iterator(types.CompanyProposalPrefix, types.PrefixEndBytes(types.CompanyProposalPrefix))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var proposal types.CompanyGovernanceProposal
		if err := json.Unmarshal(iterator.Value(), &proposal); err == nil {
			if proposal.ProposalID == proposalID {
				return proposal, true
			}
		}
	}

	return types.CompanyGovernanceProposal{}, false
}

func (k Keeper) indexCompanyProposal(ctx sdk.Context, proposal types.CompanyGovernanceProposal) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))

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