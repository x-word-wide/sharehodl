package keeper

import (
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

	// Get current tally or initialize
	tally, found := k.GetTallyResult(ctx, proposalID)
	if !found {
		tally = types.TallyResult{
			ProposalID:    proposalID,
			YesVotes:      math.LegacyZeroDec(),
			NoVotes:       math.LegacyZeroDec(),
			AbstainVotes:  math.LegacyZeroDec(),
			VetoVotes:     math.LegacyZeroDec(),
			TotalVotes:    math.LegacyZeroDec(),
		}
	}

	// Calculate total eligible voting power for quorum
	totalEligiblePower := k.calculateTotalEligibleVotingPower(ctx, proposal)
	tally.TotalEligibleVotes = totalEligiblePower

	// Calculate participation rate
	if !totalEligiblePower.IsZero() {
		tally.ParticipationRate = tally.TotalVotes.Quo(totalEligiblePower)
	} else {
		tally.ParticipationRate = math.LegacyZeroDec()
	}

	// Calculate vote percentages
	if !tally.TotalVotes.IsZero() {
		tally.YesPercentage = tally.YesVotes.Quo(tally.TotalVotes)
		tally.NoPercentage = tally.NoVotes.Quo(tally.TotalVotes)
		tally.AbstainPercentage = tally.AbstainVotes.Quo(tally.TotalVotes)
		tally.VetoPercentage = tally.VetoVotes.Quo(tally.TotalVotes)
	} else {
		tally.YesPercentage = math.LegacyZeroDec()
		tally.NoPercentage = math.LegacyZeroDec()
		tally.AbstainPercentage = math.LegacyZeroDec()
		tally.VetoPercentage = math.LegacyZeroDec()
	}

	// Determine outcome
	params := k.getGovernanceParams(ctx)
	tally.Outcome = k.determineProposalOutcome(proposal, tally, params)

	// Store final tally
	k.setTallyResult(ctx, proposalID, tally)

	return tally, nil
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
func (k Keeper) calculateValidatorVotingPower(ctx sdk.Context) math.LegacyDec {
	totalPower := math.LegacyZeroDec()
	
	// This would need to iterate through all validators
	// For now, we'll use a placeholder implementation
	// In a real implementation, you'd iterate through the validator set
	
	return totalPower
}

// calculateHODLHolderVotingPower calculates total voting power of all HODL holders
func (k Keeper) calculateHODLHolderVotingPower(ctx sdk.Context) math.LegacyDec {
	totalPower := math.LegacyZeroDec()
	
	// This would need to get total HODL supply and calculate voting power
	// For now, we'll use a placeholder implementation
	
	return totalPower
}

// calculateGoldValidatorVotingPower calculates voting power of Gold tier validators only
func (k Keeper) calculateGoldValidatorVotingPower(ctx sdk.Context) math.LegacyDec {
	totalPower := math.LegacyZeroDec()
	
	// This would iterate through validators and only count Gold tier
	// For now, we'll use a placeholder implementation
	
	return totalPower
}

// determineProposalOutcome determines the outcome of a proposal based on tally and parameters
func (k Keeper) determineProposalOutcome(
	proposal types.Proposal,
	tally types.TallyResult,
	params types.GovernanceParams,
) types.ProposalOutcome {
	
	// Check if voting period has ended
	if proposal.Status == types.StatusVotingPeriod {
		return types.OutcomeVotingInProgress
	}

	// Check for veto
	if tally.VetoPercentage.GTE(params.VetoThreshold) {
		return types.OutcomeVetoed
	}

	// Check quorum
	quorum := proposal.QuorumRequired
	if quorum.IsZero() {
		quorum = params.Quorum
	}
	
	if tally.ParticipationRate.LT(quorum) {
		return types.OutcomeQuorumNotMet
	}

	// Check threshold
	threshold := proposal.ThresholdRequired
	if threshold.IsZero() {
		threshold = params.Threshold
	}

	// Calculate yes votes as percentage of non-abstain votes
	nonAbstainVotes := tally.YesVotes.Add(tally.NoVotes)
	if nonAbstainVotes.IsZero() {
		return types.OutcomeRejected
	}

	yesPercentageOfNonAbstain := tally.YesVotes.Quo(nonAbstainVotes)
	
	if yesPercentageOfNonAbstain.GTE(threshold) {
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
	if proposal.Status != types.StatusVotingPeriod {
		return types.ErrInvalidProposalStatus
	}

	// Check if voting period has ended
	if ctx.BlockTime().Before(proposal.VotingEndTime) {
		return types.ErrVotingPeriodNotStarted
	}

	// Calculate final tally
	tally, err := k.TallyProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	// Update proposal status based on outcome
	params := k.getGovernanceParams(ctx)
	
	switch tally.Outcome {
	case types.OutcomePassed:
		proposal.Status = types.StatusPassed
		// Execute the proposal
		err = k.executeProposal(ctx, proposal)
		if err != nil {
			proposal.Status = types.StatusExecutionFailed
		} else {
			proposal.Status = types.StatusExecuted
		}
		
		// Refund deposits on successful proposals
		k.refundProposalDeposits(ctx, proposalID)

	case types.OutcomeRejected, types.OutcomeQuorumNotMet:
		proposal.Status = types.StatusRejected
		
		// Burn or refund deposits based on parameters
		if params.BurnDeposits {
			k.burnProposalDeposits(ctx, proposalID)
		} else {
			k.refundProposalDeposits(ctx, proposalID)
		}

	case types.OutcomeVetoed:
		proposal.Status = types.StatusVetoed
		
		// Always burn deposits on vetoed proposals
		if params.BurnVoteVeto {
			k.burnProposalDeposits(ctx, proposalID)
		} else {
			k.refundProposalDeposits(ctx, proposalID)
		}
	}

	// Update proposal
	proposal.UpdatedAt = ctx.BlockTime()
	k.setProposal(ctx, proposal)
	k.updateProposalIndex(ctx, proposal)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"proposal_finalized",
			sdk.NewAttribute("proposal_id", sdk.Uint64ToBigEndian(proposalID).String()),
			sdk.NewAttribute("status", string(proposal.Status)),
			sdk.NewAttribute("outcome", string(tally.Outcome)),
			sdk.NewAttribute("yes_votes", tally.YesVotes.String()),
			sdk.NewAttribute("no_votes", tally.NoVotes.String()),
			sdk.NewAttribute("abstain_votes", tally.AbstainVotes.String()),
			sdk.NewAttribute("veto_votes", tally.VetoVotes.String()),
			sdk.NewAttribute("participation_rate", tally.ParticipationRate.String()),
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
	// Implementation would transfer funds from treasury
	return nil
}

func (k Keeper) executeEmergencyActionProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Implementation would execute emergency actions
	return nil
}

// EndBlocker processes proposals that have reached their end time
func (k Keeper) EndBlocker(ctx sdk.Context) {
	// Process all proposals that have finished voting
	k.IterateProposals(ctx, func(proposal types.Proposal) bool {
		if proposal.Status == types.StatusVotingPeriod && 
		   ctx.BlockTime().After(proposal.VotingEndTime) {
			
			err := k.ProcessProposalResults(ctx, proposal.ID)
			if err != nil {
				// Log error but continue processing other proposals
				ctx.Logger().Error("Failed to process proposal results", 
					"proposal_id", proposal.ID, "error", err)
			}
		}
		
		// Check for proposals that didn't reach minimum deposit
		if proposal.Status == types.StatusDepositPeriod && 
		   ctx.BlockTime().After(proposal.DepositEndTime) {
			
			params := k.getGovernanceParams(ctx)
			if proposal.TotalDeposit.LT(params.MinDeposit) {
				// Reject proposal due to insufficient deposit
				proposal.Status = types.StatusRejected
				proposal.UpdatedAt = ctx.BlockTime()
				k.setProposal(ctx, proposal)
				k.updateProposalIndex(ctx, proposal)
				
				// Refund or burn deposits based on parameters
				if params.BurnDeposits {
					k.burnProposalDeposits(ctx, proposal.ID)
				} else {
					k.refundProposalDeposits(ctx, proposal.ID)
				}
			}
		}
		
		return false // Continue iterating
	})
}