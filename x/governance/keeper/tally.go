package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
)

// TallyProposal calculates the final tally for a proposal
func (k Keeper) TallyProposal(ctx sdk.Context, proposalID uint64) (types.TallyResult, error) {
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.TallyResult{}, types.ErrProposalNotFound
	}

	// Return the tally result from the proposal
	return proposal.CalculateTallyResult(), nil
}

// calculateTotalEligibleVotingPower calculates the total voting power eligible for a proposal
func (k Keeper) calculateTotalEligibleVotingPower(ctx sdk.Context, proposal types.Proposal) math.LegacyDec {
	totalPower := math.LegacyZeroDec()

	switch proposal.Type {
	case types.ProposalTypeCompanyListing, types.ProposalTypeCompanyDelisting, types.ProposalTypeCompanyParameter:
		// All validators and HODL holders can vote
		totalPower = k.calculateValidatorVotingPower(ctx)
		totalPower = totalPower.Add(k.calculateHODLHolderVotingPower(ctx))

	case types.ProposalTypeValidatorPromotion, types.ProposalTypeValidatorDemotion, types.ProposalTypeValidatorRemoval:
		// Only validators can vote
		totalPower = k.calculateValidatorVotingPower(ctx)

	case types.ProposalTypeProtocolParameter, types.ProposalTypeProtocolUpgrade:
		// All validators and HODL holders can vote with different weights
		totalPower = k.calculateValidatorVotingPower(ctx)
		totalPower = totalPower.Add(k.calculateHODLHolderVotingPower(ctx))

	case types.ProposalTypeTreasurySpend:
		// All HODL holders can vote
		totalPower = k.calculateHODLHolderVotingPower(ctx)

	case types.ProposalTypeEmergencyAction:
		// Only Gold tier validators can vote
		totalPower = k.calculateGoldValidatorVotingPower(ctx)
	}

	return totalPower
}

// calculateValidatorVotingPower calculates total voting power of all validators
// SECURITY FIX: Actually calculate voting power from validator stakes
func (k Keeper) calculateValidatorVotingPower(ctx sdk.Context) math.LegacyDec {
	totalPower := math.LegacyZeroDec()

	// Get total staked tokens from staking keeper
	totalStaked := k.stakingKeeper.GetTotalStaked(ctx)
	if totalStaked.IsPositive() {
		// Validator voting power = total staked / 1_000_000 (convert from micro)
		totalPower = math.LegacyNewDecFromInt(totalStaked).QuoInt64(1000000)
	}

	return totalPower
}

// calculateHODLHolderVotingPower calculates total voting power of all HODL holders
// SECURITY FIX: Actually calculate voting power from HODL supply
func (k Keeper) calculateHODLHolderVotingPower(ctx sdk.Context) math.LegacyDec {
	totalPower := math.LegacyZeroDec()

	// Get total HODL supply
	hodlSupply := k.hodlKeeper.GetTotalSupply(ctx)
	if hodlSupply != nil {
		if supply, ok := hodlSupply.(math.Int); ok && supply.IsPositive() {
			// HODL holder voting power = total supply / 1_000_000 (convert from uhodl)
			// This represents the aggregate voting power of all HODL holders
			totalPower = math.LegacyNewDecFromInt(supply).QuoInt64(1000000)
		}
	}

	return totalPower
}

// calculateGoldValidatorVotingPower calculates voting power of Archon tier validators and above
// Note: This is a simplified implementation that uses total staked as a proxy.
// In production, this would query specific tier stats from the staking module.
func (k Keeper) calculateGoldValidatorVotingPower(ctx sdk.Context) math.LegacyDec {
	// Use a fraction of total staked as a proxy for Archon+ voting power
	// This assumes approximately 10% of stakers are Archon tier or above
	totalStaked := k.stakingKeeper.GetTotalStaked(ctx)
	if totalStaked.IsPositive() {
		// Archon+ voting power = 10% of total staked / 1_000_000
		return math.LegacyNewDecFromInt(totalStaked).QuoInt64(10000000)
	}
	return math.LegacyZeroDec()
}

// determineProposalOutcome determines the outcome of a proposal based on tally and parameters
func (k Keeper) determineProposalOutcome(
	proposal types.Proposal,
	tally types.TallyResult,
) types.ProposalOutcome {

	// Check if voting period has ended
	if proposal.Status == types.ProposalStatusVotingPeriod {
		return types.OutcomeVotingInProgress
	}

	// Check for veto
	if tally.VetoPercentage.GTE(proposal.VetoThreshold) {
		return types.OutcomeVetoed
	}

	// Check quorum
	if tally.Turnout.LT(proposal.Quorum) {
		return types.OutcomeQuorumNotMet
	}

	// Check threshold
	if tally.YesPercentage.GTE(proposal.Threshold) {
		return types.OutcomePassed
	}

	return types.OutcomeRejected
}

// ProcessProposalResults processes the results of a finished proposal
func (k Keeper) ProcessProposalResults(ctx sdk.Context, proposalID uint64) error {
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Can only process proposals that have finished voting
	if proposal.Status != types.ProposalStatusVotingPeriod {
		return types.ErrInvalidProposalStatus
	}

	// Check if voting period has ended
	if ctx.BlockTime().Before(proposal.VotingEndTime) {
		return types.ErrVotingPeriodNotStarted
	}

	// Calculate final tally
	tally := proposal.CalculateTallyResult()
	outcome := k.determineProposalOutcome(proposal, tally)

	// Update proposal status based on outcome
	switch outcome {
	case types.OutcomePassed:
		proposal.Status = types.ProposalStatusPassed
		// Execute the proposal
		err := k.executeProposal(ctx, proposal)
		if err != nil {
			proposal.Status = types.ProposalStatusFailed
		}

		// Refund deposits on successful proposals
		k.refundProposalDeposits(ctx, proposalID)

	case types.OutcomeRejected, types.OutcomeQuorumNotMet:
		proposal.Status = types.ProposalStatusRejected

		// Refund deposits
		k.refundProposalDeposits(ctx, proposalID)

	case types.OutcomeVetoed:
		proposal.Status = types.ProposalStatusRejected

		// Refund deposits
		k.refundProposalDeposits(ctx, proposalID)
	}

	// Update proposal
	proposal.FinalizedAt = ctx.BlockTime()
	proposal.UpdatedAt = ctx.BlockTime()
	k.setProposal(ctx, proposal)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"proposal_finalized",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("status", proposal.Status.String()),
			sdk.NewAttribute("outcome", string(outcome)),
			sdk.NewAttribute("yes_votes", tally.YesVotes.String()),
			sdk.NewAttribute("no_votes", tally.NoVotes.String()),
			sdk.NewAttribute("abstain_votes", tally.AbstainVotes.String()),
			sdk.NewAttribute("veto_votes", tally.NoWithVetoVotes.String()),
			sdk.NewAttribute("turnout", tally.Turnout.String()),
		),
	)

	return nil
}

// executeProposal executes a passed proposal
func (k Keeper) executeProposal(ctx sdk.Context, proposal types.Proposal) error {
	switch proposal.Type {
	case types.ProposalTypeCompanyListing:
		return k.executeCompanyListingProposal(ctx, proposal)
	case types.ProposalTypeCompanyDelisting:
		return k.executeCompanyDelistingProposal(ctx, proposal)
	case types.ProposalTypeCompanyParameter:
		return k.executeCompanyParameterProposal(ctx, proposal)
	case types.ProposalTypeValidatorPromotion:
		return k.executeValidatorPromotionProposal(ctx, proposal)
	case types.ProposalTypeValidatorDemotion:
		return k.executeValidatorDemotionProposal(ctx, proposal)
	case types.ProposalTypeValidatorRemoval:
		return k.executeValidatorRemovalProposal(ctx, proposal)
	case types.ProposalTypeProtocolParameter:
		return k.executeProtocolParameterProposal(ctx, proposal)
	case types.ProposalTypeProtocolUpgrade:
		return k.executeProtocolUpgradeProposal(ctx, proposal)
	case types.ProposalTypeTreasurySpend:
		return k.executeTreasurySpendProposal(ctx, proposal)
	case types.ProposalTypeEmergencyAction:
		return k.executeEmergencyActionProposal(ctx, proposal)
	case types.ProposalTypeSetCharityWallet:
		return k.ExecuteSetCharityWalletProposal(ctx, proposal)
	case types.ProposalTypeDividendDistribution:
		return k.executeDividendDistributionProposal(ctx, proposal)
	default:
		return types.ErrInvalidProposalType
	}
}

// Placeholder execution functions - these would be implemented based on proposal content
func (k Keeper) executeCompanyListingProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Implementation would parse proposal content and call equity module
	return nil
}

func (k Keeper) executeCompanyDelistingProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Implementation would parse proposal content and call equity module
	return nil
}

func (k Keeper) executeCompanyParameterProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Implementation would parse proposal content and update company parameters
	return nil
}

func (k Keeper) executeValidatorPromotionProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Implementation would parse proposal content and call validator module
	return nil
}

func (k Keeper) executeValidatorDemotionProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Implementation would parse proposal content and call validator module
	return nil
}

func (k Keeper) executeValidatorRemovalProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Implementation would parse proposal content and call validator module
	return nil
}

func (k Keeper) executeProtocolParameterProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Implementation would parse proposal content and update protocol parameters
	return nil
}

func (k Keeper) executeProtocolUpgradeProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Implementation would handle software upgrade logic
	return nil
}

func (k Keeper) executeTreasurySpendProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Extract recipient from proposal metadata with type assertion
	recipientVal, ok := proposal.Metadata["recipient"]
	if !ok {
		return fmt.Errorf("treasury spend proposal missing recipient")
	}
	recipientStr, ok := recipientVal.(string)
	if !ok || recipientStr == "" {
		return fmt.Errorf("treasury spend proposal recipient must be a non-empty string")
	}

	// Extract amount from proposal metadata with type assertion
	amountVal, ok := proposal.Metadata["amount"]
	if !ok {
		return fmt.Errorf("treasury spend proposal missing amount")
	}
	amountStr, ok := amountVal.(string)
	if !ok || amountStr == "" {
		return fmt.Errorf("treasury spend proposal amount must be a non-empty string")
	}

	// Parse recipient address
	recipient, err := sdk.AccAddressFromBech32(recipientStr)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}

	// Parse amount
	amount, ok := math.NewIntFromString(amountStr)
	if !ok {
		return fmt.Errorf("invalid amount: %s", amountStr)
	}

	// Check treasury has sufficient funds
	if k.feeAbstractionKeeper != nil {
		balance := k.feeAbstractionKeeper.GetTreasuryBalance(ctx)
		if balance.LT(amount) {
			return fmt.Errorf("insufficient treasury funds: have %s, need %s", balance.String(), amount.String())
		}

		// Execute the withdrawal via feeabstraction treasury
		if err := k.feeAbstractionKeeper.WithdrawFromTreasury(ctx, recipient, amount); err != nil {
			return fmt.Errorf("treasury withdrawal failed: %w", err)
		}
	}

	// Emit treasury spend event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_spend_executed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
			sdk.NewAttribute("recipient", recipientStr),
			sdk.NewAttribute("amount", amountStr),
			sdk.NewAttribute("title", proposal.Title),
		),
	)

	return nil
}

func (k Keeper) executeEmergencyActionProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Implementation would execute emergency actions
	return nil
}

// executeDividendDistributionProposal approves a dividend for distribution
// This is called when validators approve a dividend proposal with verified audit
func (k Keeper) executeDividendDistributionProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Extract dividend_id from proposal metadata with type assertion
	dividendIDVal, ok := proposal.Metadata["dividend_id"]
	if !ok {
		return fmt.Errorf("dividend distribution proposal missing dividend_id")
	}

	// Handle both string and float64 (JSON number) types
	var dividendID uint64
	switch v := dividendIDVal.(type) {
	case string:
		var err error
		dividendID, err = parseUint64(v)
		if err != nil {
			return fmt.Errorf("invalid dividend_id: %w", err)
		}
	case float64:
		dividendID = uint64(v)
	case uint64:
		dividendID = v
	default:
		return fmt.Errorf("dividend_id must be a number or string, got %T", dividendIDVal)
	}

	if dividendID == 0 {
		return fmt.Errorf("dividend_id cannot be zero")
	}

	// Call equity keeper to approve the dividend
	if k.equityKeeper == nil {
		return fmt.Errorf("equity keeper not available")
	}

	if err := k.equityKeeper.ApproveDividendDistribution(ctx, dividendID, proposal.ID); err != nil {
		return fmt.Errorf("failed to approve dividend distribution: %w", err)
	}

	// Emit dividend approval event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"dividend_distribution_approved",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
			sdk.NewAttribute("dividend_id", fmt.Sprintf("%d", dividendID)),
			sdk.NewAttribute("title", proposal.Title),
		),
	)

	return nil
}

// parseUint64 parses a string to uint64
func parseUint64(s string) (uint64, error) {
	var result uint64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// refundProposalDeposits refunds deposits to depositors
func (k Keeper) refundProposalDeposits(ctx sdk.Context, proposalID uint64) {
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return
	}

	for _, deposit := range proposal.Deposits {
		depositorAddr, err := sdk.AccAddressFromBech32(deposit.Depositor)
		if err != nil {
			continue
		}

		coins := sdk.NewCoins(sdk.NewCoin(deposit.Currency, deposit.Amount))
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositorAddr, coins)
		if err != nil {
			ctx.Logger().Error("Failed to refund deposit",
				"depositor", deposit.Depositor,
				"amount", deposit.Amount.String(),
				"error", err)
		}
	}
}

// burnProposalDeposits burns deposits
func (k Keeper) burnProposalDeposits(ctx sdk.Context, proposalID uint64) {
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return
	}

	for _, deposit := range proposal.Deposits {
		coins := sdk.NewCoins(sdk.NewCoin(deposit.Currency, deposit.Amount))
		err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins)
		if err != nil {
			ctx.Logger().Error("Failed to burn deposit",
				"depositor", deposit.Depositor,
				"amount", deposit.Amount.String(),
				"error", err)
		}
	}
}

// GetGovernanceParams retrieves governance parameters (exported)
func (k Keeper) GetGovernanceParams(ctx sdk.Context) types.GovernanceParams {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsPrefix)
	if err != nil || bz == nil {
		return types.DefaultParams()
	}

	var params types.GovernanceParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultParams()
	}
	return params
}

// IterateProposals iterates through all proposals
func (k Keeper) IterateProposals(ctx sdk.Context, fn func(proposal types.Proposal) bool) {
	store := k.storeService.OpenKVStore(ctx)

	iter, err := store.Iterator(types.ProposalPrefix, types.PrefixEndBytes(types.ProposalPrefix))
	if err != nil {
		return
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var proposal types.Proposal
		if err := json.Unmarshal(iter.Value(), &proposal); err != nil {
			continue
		}
		if fn(proposal) {
			break
		}
	}
}

// updateProposalIndex updates proposal indexes after status change
func (k Keeper) updateProposalIndex(ctx sdk.Context, proposal types.Proposal) {
	// Update status index
	store := k.storeService.OpenKVStore(ctx)
	statusKey := types.ProposalByStatusKey(proposal.Status.String(), proposal.ID)
	store.Set(statusKey, []byte{1})
}
