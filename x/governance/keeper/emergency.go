package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
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
	targetEntity string,
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
		ProposalID:    proposalID,
		EmergencyType: emergencyType,
		SeverityLevel: severityLevel,
		Justification: justification,
		RequiredTier:  fmt.Sprintf("%d", requiredTier),
		FastTrack:     severityLevel >= 4,
		ExecuteOnPass: severityLevel >= 3,
		TimeLimit:     k.getExecutionTimeLimit(emergencyType, severityLevel),
		CreatedAt:     ctx.BlockTime(),
	}

	// Create main governance proposal
	proposal := types.Proposal{
		ID:               proposalID,
		Type:             types.ProposalTypeEmergencyAction,
		Title:            fmt.Sprintf("[EMERGENCY] %s", title),
		Description:      description,
		Proposer:         submitter,
		Status:           types.ProposalStatusVotingPeriod,
		VotingStartTime:  ctx.BlockTime(),
		VotingEndTime:    ctx.BlockTime().Add(votingPeriod),
		DepositEndTime:   ctx.BlockTime(),
		TotalDeposit:     math.ZeroInt(),
		MinDeposit:       math.ZeroInt(),
		YesVotes:         math.ZeroInt(),
		NoVotes:          math.ZeroInt(),
		AbstainVotes:     math.ZeroInt(),
		NoWithVetoVotes:  math.ZeroInt(),
		TotalVotingPower: math.ZeroInt(),
		Quorum:           quorum,
		Threshold:        threshold,
		VetoThreshold:    math.LegacyNewDecWithPrec(25, 2),
		ValidatorAddr:    targetEntity,
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
	}

	// Store proposals
	k.setProposal(ctx, proposal)
	k.setEmergencyProposal(ctx, emergencyProposal)

	// Index proposals
	k.indexProposal(ctx, proposal)

	// Immediately notify validators if severity is high
	if severityLevel >= 4 {
		k.broadcastEmergencyAlert(ctx, emergencyProposal)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"submit_emergency_proposal",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("severity_level", fmt.Sprintf("%d", severityLevel)),
			sdk.NewAttribute("submitter", submitter),
			sdk.NewAttribute("fast_track", fmt.Sprintf("%t", emergencyProposal.FastTrack)),
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
		VotingPower:   votingPower.TruncateInt(),
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
			ctx.Logger().Error("Failed to execute emergency proposal immediately",
				"proposal_id", proposalID, "error", err)
		}
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote_emergency_proposal",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("voter", voter),
			sdk.NewAttribute("option", option.String()),
			sdk.NewAttribute("voting_power", votingPower.String()),
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
	validator := k.GetValidatorByAddress(ctx, submitter.String())
	if validator.Address == "" {
		return types.ErrValidatorNotFound
	}

	tier := validator.Tier
	requiredTier := k.getRequiredSubmitterTier(emergencyType, severityLevel)

	if tier < requiredTier {
		return types.ErrInsufficientPermissions
	}

	// Check validator status
	if !validator.Active {
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
	validator := k.GetValidatorByAddress(ctx, voter.String())
	if validator.Address == "" {
		return types.ErrValidatorNotFound
	}

	tier := validator.Tier
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
	validator := k.GetValidatorByAddress(ctx, voter.String())
	if validator.Address == "" {
		return math.LegacyZeroDec(), types.ErrValidatorNotFound
	}

	tier := validator.Tier

	// Base voting power from tier
	basePower := k.getTierMultiplier(tier)

	// Emergency multiplier based on type and severity
	emergencyMultiplier := k.getEmergencyMultiplier(emergencyProposal.EmergencyType, emergencyProposal.SeverityLevel)

	return basePower.Mul(emergencyMultiplier), nil
}

// Helper functions for emergency parameters

func (k Keeper) getEmergencyVotingPeriod(emergencyType types.EmergencyType, severityLevel int) time.Duration {
	baseHours := 24

	switch severityLevel {
	case 5:
		baseHours = 2
	case 4:
		baseHours = 6
	case 3:
		baseHours = 12
	case 2:
		baseHours = 18
	default:
		baseHours = 24
	}

	return time.Hour * time.Duration(baseHours)
}

func (k Keeper) getEmergencyQuorum(emergencyType types.EmergencyType, severityLevel int) math.LegacyDec {
	switch severityLevel {
	case 5:
		return math.LegacyNewDecWithPrec(33, 2)
	case 4:
		return math.LegacyNewDecWithPrec(40, 2)
	case 3:
		return math.LegacyNewDecWithPrec(50, 2)
	case 2:
		return math.LegacyNewDecWithPrec(60, 2)
	default:
		return math.LegacyNewDecWithPrec(67, 2)
	}
}

func (k Keeper) getEmergencyThreshold(emergencyType types.EmergencyType, severityLevel int) math.LegacyDec {
	switch emergencyType {
	case types.EmergencyTypeSecurityBreach, types.EmergencyTypeSystemOutage:
		return math.LegacyNewDecWithPrec(60, 2)
	case types.EmergencyTypeMarketHalt:
		return math.LegacyNewDecWithPrec(67, 2)
	case types.EmergencyTypeValidatorSlash:
		return math.LegacyNewDecWithPrec(75, 2)
	default:
		return math.LegacyNewDecWithPrec(67, 2)
	}
}

func (k Keeper) getRequiredValidatorTier(emergencyType types.EmergencyType, severityLevel int) int {
	switch severityLevel {
	case 5, 4:
		return 4 // Archon
	case 3:
		return 3 // Steward
	default:
		return 2 // Warden
	}
}

func (k Keeper) getRequiredSubmitterTier(emergencyType types.EmergencyType, severityLevel int) int {
	switch severityLevel {
	case 5:
		return 5 // Validator
	case 4:
		return 4 // Archon
	case 3:
		return 3 // Steward
	default:
		return 2 // Warden
	}
}

func (k Keeper) getEmergencyMultiplier(emergencyType types.EmergencyType, severityLevel int) math.LegacyDec {
	baseMultiplier := math.LegacyNewDec(int64(severityLevel))

	switch emergencyType {
	case types.EmergencyTypeSecurityBreach:
		return baseMultiplier.Mul(math.LegacyNewDecWithPrec(15, 1))
	case types.EmergencyTypeSystemOutage:
		return baseMultiplier.Mul(math.LegacyNewDecWithPrec(12, 1))
	default:
		return baseMultiplier
	}
}

func (k Keeper) getExecutionTimeLimit(emergencyType types.EmergencyType, severityLevel int) time.Duration {
	switch severityLevel {
	case 5:
		return time.Minute * 15
	case 4:
		return time.Hour * 1
	case 3:
		return time.Hour * 4
	default:
		return time.Hour * 24
	}
}

func (k Keeper) hasReachedEmergencyThreshold(
	ctx sdk.Context,
	proposal types.Proposal,
	emergencyProposal types.EmergencyProposal,
) bool {
	tally := proposal.CalculateTallyResult()

	if tally.Turnout.LT(proposal.Quorum) {
		return false
	}

	return tally.YesPercentage.GTE(proposal.Threshold)
}

// Alert and notification functions

func (k Keeper) broadcastEmergencyAlert(ctx sdk.Context, proposal types.EmergencyProposal) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"emergency_alert",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ProposalID)),
			sdk.NewAttribute("severity_level", fmt.Sprintf("%d", proposal.SeverityLevel)),
			sdk.NewAttribute("alert_time", ctx.BlockTime().String()),
		),
	)
}

// Storage functions

func (k Keeper) setEmergencyProposal(ctx sdk.Context, proposal types.EmergencyProposal) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.EmergencyProposalKey(proposal.ProposalID)
	value, _ := json.Marshal(proposal)
	store.Set(key, value)
}

func (k Keeper) GetEmergencyProposal(ctx sdk.Context, proposalID uint64) (types.EmergencyProposal, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.EmergencyProposalKey(proposalID)

	value, err := store.Get(key)
	if err != nil || value == nil {
		return types.EmergencyProposal{}, false
	}

	var proposal types.EmergencyProposal
	if err := json.Unmarshal(value, &proposal); err != nil {
		return types.EmergencyProposal{}, false
	}
	return proposal, true
}

// Execution implementations (placeholders)

func (k Keeper) executeSecurityBreachResponse(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	return nil
}

func (k Keeper) executeMarketHalt(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	return nil
}

func (k Keeper) executeValidatorSlashing(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	return nil
}

func (k Keeper) executeEmergencyProtocolUpgrade(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	return nil
}

func (k Keeper) executeEmergencyDelisting(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	return nil
}

func (k Keeper) executeSystemOutageResponse(ctx sdk.Context, proposal types.Proposal, emergency types.EmergencyProposal) error {
	return nil
}
