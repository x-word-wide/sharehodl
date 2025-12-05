package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
	validatortypes "github.com/sharehodl/sharehodl-blockchain/x/validator/types"
	"fmt"
	"time"
)

// SubmitEmergencyProposal submits an emergency proposal with expedited procedures
func (k Keeper) SubmitEmergencyProposal(
	ctx sdk.Context,
	submitter string,
	emergencyType types.EmergencyType,
	title string,
	description string,
	justification string,
	severityLevel int,
	targetEntity string, // For validator/company specific emergencies
) (uint64, error) {
	submitterAddr, err := sdk.AccAddressFromBech32(submitter)
	if err != nil {
		return 0, types.ErrInvalidAddress
	}

	// Validate emergency proposal submission
	if err := k.validateEmergencyProposalSubmission(ctx, submitterAddr, emergencyType, severityLevel); err != nil {
		return 0, err
	}

	// Get next proposal ID
	proposalID := k.getNextProposalID(ctx)

	// Determine voting parameters based on emergency type and severity
	votingPeriod := k.getEmergencyVotingPeriod(emergencyType, severityLevel)
	quorum := k.getEmergencyQuorum(emergencyType, severityLevel)
	threshold := k.getEmergencyThreshold(emergencyType, severityLevel)
	requiredTier := k.getRequiredValidatorTier(emergencyType, severityLevel)

	// Create emergency proposal
	emergencyProposal := types.EmergencyProposal{
		ProposalID:     proposalID,
		EmergencyType:  emergencyType,
		SeverityLevel:  severityLevel,
		Justification:  justification,
		RequiredTier:   requiredTier.String(),
		FastTrack:      severityLevel >= 4, // Severity 4+ gets fast track
		ExecuteOnPass:  severityLevel >= 3, // Severity 3+ executes immediately
		TimeLimit:      k.getExecutionTimeLimit(emergencyType, severityLevel),
		CreatedAt:      ctx.BlockTime(),
	}

	// Create main governance proposal
	proposal := types.Proposal{
		ID:                proposalID,
		Type:              types.ProposalTypeEmergencyAction,
		Title:             fmt.Sprintf("[EMERGENCY:%s] %s", emergencyType.String(), title),
		Description:       description,
		Proposer:          submitter,
		Status:            types.ProposalStatusVotingPeriod, // Skip deposit period
		VotingStartTime:   ctx.BlockTime(),
		VotingEndTime:     ctx.BlockTime().Add(votingPeriod),
		DepositEndTime:    ctx.BlockTime(),
		TotalDeposit:      math.ZeroInt(),
		MinDeposit:        math.ZeroInt(),
		YesVotes:          math.ZeroInt(),
		NoVotes:           math.ZeroInt(),
		AbstainVotes:      math.ZeroInt(),
		NoWithVetoVotes:   math.ZeroInt(),
		TotalVotingPower:  math.ZeroInt(),
		Quorum:            quorum,
		Threshold:         threshold,
		VetoThreshold:     math.LegacyNewDecWithPrec(25, 2), // 25% for emergency
		ValidatorAddr:     targetEntity,
		CreatedAt:         ctx.BlockTime(),
		UpdatedAt:         ctx.BlockTime(),
	}

	// Store proposals
	k.setProposal(ctx, proposal)
	k.setEmergencyProposal(ctx, emergencyProposal)

	// Index proposals
	k.indexProposal(ctx, proposal)
	k.indexEmergencyProposal(ctx, emergencyProposal)

	// Immediately notify validators if severity is high
	if severityLevel >= 4 {
		k.broadcastEmergencyAlert(ctx, emergencyProposal)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"submit_emergency_proposal",
			sdk.NewAttribute("proposal_id", sdk.Uint64ToBigEndian(proposalID).String()),
			sdk.NewAttribute("emergency_type", emergencyType.String()),
			sdk.NewAttribute("severity_level", fmt.Sprintf("%d", severityLevel)),
			sdk.NewAttribute("submitter", submitter),
			sdk.NewAttribute("fast_track", fmt.Sprintf("%t", emergencyProposal.FastTrack)),
			sdk.NewAttribute("execute_on_pass", fmt.Sprintf("%t", emergencyProposal.ExecuteOnPass)),
		),
	)

	return proposalID, nil
}

// VoteOnEmergencyProposal allows eligible validators to vote on emergency proposals
func (k Keeper) VoteOnEmergencyProposal(
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

	emergencyProposal, found := k.GetEmergencyProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Validate emergency voting conditions
	if err := k.validateEmergencyVoting(ctx, proposal, emergencyProposal, voterAddr); err != nil {
		return err
	}

	// Check if already voted
	if k.hasVoted(ctx, proposalID, voterAddr) {
		return types.ErrAlreadyVoted
	}

	// Calculate emergency voting power
	votingPower, err := k.calculateEmergencyVotingPower(ctx, emergencyProposal, voterAddr)
	if err != nil {
		return err
	}

	if votingPower.IsZero() {
		return types.ErrInsufficientVotingPower
	}

	// Create vote
	vote := types.Vote{
		ProposalID:    proposalID,
		Voter:         voter,
		Option:        option,
		VotingPower:   votingPower,
		Weight:        math.LegacyOneDec(),
		Justification: reason,
		VotedAt:       ctx.BlockTime(),
	}

	// Store vote
	k.setVote(ctx, vote)

	// Update tally
	k.updateTallyResult(ctx, proposal, vote, vote.Weight)

	// Check if proposal should be executed immediately
	if emergencyProposal.ExecuteOnPass && k.hasReachedEmergencyThreshold(ctx, proposal, emergencyProposal) {
		err = k.executeEmergencyProposal(ctx, proposal, emergencyProposal)
		if err != nil {
			// Log error but don't fail the vote
			ctx.Logger().Error("Failed to execute emergency proposal immediately", 
				"proposal_id", proposalID, "error", err)
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote_emergency_proposal",
			sdk.NewAttribute("proposal_id", sdk.Uint64ToBigEndian(proposalID).String()),
			sdk.NewAttribute("voter", voter),
			sdk.NewAttribute("option", option.String()),
			sdk.NewAttribute("voting_power", votingPower.String()),
			sdk.NewAttribute("emergency_type", emergencyProposal.EmergencyType.String()),
		),
	)

	return nil
}

// executeEmergencyProposal executes an emergency proposal based on its type
func (k Keeper) executeEmergencyProposal(
	ctx sdk.Context,
	proposal types.Proposal,
	emergencyProposal types.EmergencyProposal,
) error {
	switch emergencyProposal.EmergencyType {
	case types.EmergencyTypeSecurityBreach:
		return k.executeSecurityBreachResponse(ctx, proposal, emergencyProposal)
	case types.EmergencyTypeMarketHalt:
		return k.executeMarketHalt(ctx, proposal, emergencyProposal)
	case types.EmergencyTypeValidatorSlash:
		return k.executeValidatorSlashing(ctx, proposal, emergencyProposal)
	case types.EmergencyTypeProtocolUpgrade:
		return k.executeEmergencyProtocolUpgrade(ctx, proposal, emergencyProposal)
	case types.EmergencyTypeCompanyDelisting:
		return k.executeEmergencyDelisting(ctx, proposal, emergencyProposal)
	case types.EmergencyTypeSystemOutage:
		return k.executeSystemOutageResponse(ctx, proposal, emergencyProposal)
	default:
		return types.ErrInvalidEmergencyType
	}
}

// Validation functions

func (k Keeper) validateEmergencyProposalSubmission(
	ctx sdk.Context,
	submitter sdk.AccAddress,
	emergencyType types.EmergencyType,
	severityLevel int,
) error {
	// Validate severity level
	if severityLevel < 1 || severityLevel > 5 {
		return types.ErrInvalidSeverityLevel
	}

	// Check if submitter is a qualified validator
	validator, found := k.validatorKeeper.GetValidator(ctx, submitter.String())
	if !found {
		return types.ErrValidatorNotFound
	}

	tier := k.validatorKeeper.GetValidatorTier(ctx, validator.Address)
	requiredTier := k.getRequiredSubmitterTier(emergencyType, severityLevel)

	if tier < requiredTier {
		return types.ErrInsufficientPermissions
	}

	// Check validator status
	if !k.validatorKeeper.IsValidatorActive(ctx, validator.Address) {
		return types.ErrValidatorNotActive
	}

	return nil
}

func (k Keeper) validateEmergencyVoting(
	ctx sdk.Context,
	proposal types.Proposal,
	emergencyProposal types.EmergencyProposal,
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

	// Check if voter meets tier requirement
	validator, found := k.validatorKeeper.GetValidator(ctx, voter.String())
	if !found {
		return types.ErrValidatorNotFound
	}

	tier := k.validatorKeeper.GetValidatorTier(ctx, validator.Address)
	requiredTier := k.getRequiredValidatorTier(emergencyProposal.EmergencyType, emergencyProposal.SeverityLevel)

	if tier < requiredTier {
		return types.ErrInsufficientPermissions
	}

	return nil
}

// Calculation functions

func (k Keeper) calculateEmergencyVotingPower(
	ctx sdk.Context,
	emergencyProposal types.EmergencyProposal,
	voter sdk.AccAddress,
) (math.LegacyDec, error) {
	validator, found := k.validatorKeeper.GetValidator(ctx, voter.String())
	if !found {
		return math.LegacyZeroDec(), types.ErrValidatorNotFound
	}

	tier := k.validatorKeeper.GetValidatorTier(ctx, validator.Address)
	
	// Base voting power from tier
	basePower := k.getValidatorTierMultiplier(tier)
	
	// Emergency multiplier based on type and severity
	emergencyMultiplier := k.getEmergencyMultiplier(emergencyProposal.EmergencyType, emergencyProposal.SeverityLevel)
	
	return basePower.Mul(emergencyMultiplier), nil
}

// Helper functions for emergency parameters

func (k Keeper) getEmergencyVotingPeriod(emergencyType types.EmergencyType, severityLevel int) time.Duration {
	baseHours := 24
	
	// Reduce voting period based on severity
	switch severityLevel {
	case 5:
		baseHours = 2  // 2 hours for most critical
	case 4:
		baseHours = 6  // 6 hours for critical
	case 3:
		baseHours = 12 // 12 hours for high
	case 2:
		baseHours = 18 // 18 hours for medium
	default:
		baseHours = 24 // 24 hours for low
	}
	
	return time.Hour * time.Duration(baseHours)
}

func (k Keeper) getEmergencyQuorum(emergencyType types.EmergencyType, severityLevel int) math.LegacyDec {
	// Lower quorum for higher severity
	switch severityLevel {
	case 5:
		return math.LegacyNewDecWithPrec(33, 2) // 33%
	case 4:
		return math.LegacyNewDecWithPrec(40, 2) // 40%
	case 3:
		return math.LegacyNewDecWithPrec(50, 2) // 50%
	case 2:
		return math.LegacyNewDecWithPrec(60, 2) // 60%
	default:
		return math.LegacyNewDecWithPrec(67, 2) // 67%
	}
}

func (k Keeper) getEmergencyThreshold(emergencyType types.EmergencyType, severityLevel int) math.LegacyDec {
	// Different thresholds based on emergency type
	switch emergencyType {
	case types.EmergencyTypeSecurityBreach, types.EmergencyTypeSystemOutage:
		return math.LegacyNewDecWithPrec(60, 2) // 60% for security issues
	case types.EmergencyTypeMarketHalt:
		return math.LegacyNewDecWithPrec(67, 2) // 67% for market halts
	case types.EmergencyTypeValidatorSlash:
		return math.LegacyNewDecWithPrec(75, 2) // 75% for validator punishment
	default:
		return math.LegacyNewDecWithPrec(67, 2) // 67% default
	}
}

func (k Keeper) getRequiredValidatorTier(emergencyType types.EmergencyType, severityLevel int) validatortypes.ValidatorTier {
	switch severityLevel {
	case 5, 4:
		return validatortypes.TierGold // Highest severity requires Gold+
	case 3:
		return validatortypes.TierSilver // High severity requires Silver+
	default:
		return validatortypes.TierBronze // Lower severity requires Bronze+
	}
}

func (k Keeper) getRequiredSubmitterTier(emergencyType types.EmergencyType, severityLevel int) validatortypes.ValidatorTier {
	// Submitters need higher tier than voters
	switch severityLevel {
	case 5:
		return validatortypes.TierPlatinum // Most critical requires Platinum
	case 4:
		return validatortypes.TierGold     // Critical requires Gold
	case 3:
		return validatortypes.TierSilver   // High requires Silver
	default:
		return validatortypes.TierBronze   // Lower requires Bronze
	}
}

func (k Keeper) getEmergencyMultiplier(emergencyType types.EmergencyType, severityLevel int) math.LegacyDec {
	// Higher severity gets higher multiplier
	baseMultiplier := math.LegacyNewDec(int64(severityLevel))
	
	// Type-specific adjustments
	switch emergencyType {
	case types.EmergencyTypeSecurityBreach:
		return baseMultiplier.Mul(math.LegacyNewDecWithPrec(15, 1)) // 1.5x
	case types.EmergencyTypeSystemOutage:
		return baseMultiplier.Mul(math.LegacyNewDecWithPrec(12, 1)) // 1.2x
	default:
		return baseMultiplier
	}
}

func (k Keeper) getExecutionTimeLimit(emergencyType types.EmergencyType, severityLevel int) time.Duration {
	switch severityLevel {
	case 5:
		return time.Minute * 15 // 15 minutes for most critical
	case 4:
		return time.Hour * 1    // 1 hour for critical
	case 3:
		return time.Hour * 4    // 4 hours for high
	default:
		return time.Hour * 24   // 24 hours for lower severity
	}
}

func (k Keeper) hasReachedEmergencyThreshold(
	ctx sdk.Context,
	proposal types.Proposal,
	emergencyProposal types.EmergencyProposal,
) bool {
	tally := proposal.CalculateTallyResult()
	
	// Check if threshold is met
	if tally.Turnout.LT(proposal.Quorum) {
		return false
	}
	
	return tally.YesPercentage.GTE(proposal.Threshold)
}

// Alert and notification functions

func (k Keeper) broadcastEmergencyAlert(ctx sdk.Context, proposal types.EmergencyProposal) {
	// Implementation would send alerts to all validators
	// This could use the event system or a separate notification mechanism
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"emergency_alert",
			sdk.NewAttribute("proposal_id", sdk.Uint64ToBigEndian(proposal.ProposalID).String()),
			sdk.NewAttribute("emergency_type", proposal.EmergencyType.String()),
			sdk.NewAttribute("severity_level", fmt.Sprintf("%d", proposal.SeverityLevel)),
			sdk.NewAttribute("alert_time", ctx.BlockTime().String()),
		),
	)
}

// Storage functions

func (k Keeper) setEmergencyProposal(ctx sdk.Context, proposal types.EmergencyProposal) {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	key := types.EmergencyProposalKey(proposal.ProposalID)
	value := k.cdc.MustMarshal(&proposal)
	store.Set(key, value)
}

func (k Keeper) GetEmergencyProposal(ctx sdk.Context, proposalID uint64) (types.EmergencyProposal, bool) {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	key := types.EmergencyProposalKey(proposalID)
	
	value := store.Get(key)
	if value == nil {
		return types.EmergencyProposal{}, false
	}

	var proposal types.EmergencyProposal
	k.cdc.MustUnmarshal(value, &proposal)
	return proposal, true
}

func (k Keeper) indexEmergencyProposal(ctx sdk.Context, proposal types.EmergencyProposal) {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	
	// Index by emergency type
	typeKey := types.EmergencyTypeIndexKey(proposal.EmergencyType)
	typeIndexKey := append(typeKey, sdk.Uint64ToBigEndian(proposal.ProposalID)...)
	store.Set(typeIndexKey, []byte{})
	
	// Index by severity
	severityKey := types.EmergencySeverityIndexKey(proposal.SeverityLevel)
	severityIndexKey := append(severityKey, sdk.Uint64ToBigEndian(proposal.ProposalID)...)
	store.Set(severityIndexKey, []byte{})
}

// Execution implementations (placeholders)

func (k Keeper) executeSecurityBreachResponse(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	// Implementation would handle security breach response
	return nil
}

func (k Keeper) executeMarketHalt(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	// Implementation would halt trading
	return nil
}

func (k Keeper) executeValidatorSlashing(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	// Implementation would slash validator
	return nil
}

func (k Keeper) executeEmergencyProtocolUpgrade(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	// Implementation would trigger protocol upgrade
	return nil
}

func (k Keeper) executeEmergencyDelisting(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	// Implementation would delist company
	return nil
}

func (k Keeper) executeSystemOutageResponse(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	// Implementation would handle system outage
	return nil
}