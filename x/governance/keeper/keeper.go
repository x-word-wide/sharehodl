package keeper

import (
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
	equitytypes "github.com/sharehodl/sharehodl-blockchain/x/equity/types"
	hodltypes "github.com/sharehodl/sharehodl-blockchain/x/hodl/types"
	validatortypes "github.com/sharehodl/sharehodl-blockchain/x/validator/types"
)

// Keeper handles governance operations
type Keeper struct {
	cdc          codec.Codec
	storeService storetypes.KVStoreService

	// Module keepers for integration
	equityKeeper    EquityKeeper
	hodlKeeper      HODLKeeper
	validatorKeeper ValidatorKeeper
	bankKeeper      BankKeeper

	// Authority for governance parameter updates
	authority string
}

// Interface definitions for module dependencies
type EquityKeeper interface {
	GetCompany(ctx sdk.Context, id uint64) (equitytypes.Company, bool)
	GetShareholding(ctx sdk.Context, companyID uint64, shareholder string) (equitytypes.Shareholding, bool)
	GetTotalShares(ctx sdk.Context, companyID uint64, classID string) math.Int
	IterateShareholdings(ctx sdk.Context, companyID uint64, classID string, cb func(shareholding equitytypes.Shareholding) bool)
}

type HODLKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress) math.Int
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount math.Int) error
	BurnTokens(ctx sdk.Context, addr sdk.AccAddress, amount math.Int) error
}

type ValidatorKeeper interface {
	GetValidator(ctx sdk.Context, addr string) (validatortypes.Validator, bool)
	GetValidatorTier(ctx sdk.Context, addr string) validatortypes.ValidatorTier
	IsValidatorActive(ctx sdk.Context, addr string) bool
}

type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}

// NewKeeper creates a new governance Keeper
func NewKeeper(
	cdc codec.Codec,
	storeService storetypes.KVStoreService,
	equityKeeper EquityKeeper,
	hodlKeeper HODLKeeper,
	validatorKeeper ValidatorKeeper,
	bankKeeper BankKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:             cdc,
		storeService:    storeService,
		equityKeeper:    equityKeeper,
		hodlKeeper:      hodlKeeper,
		validatorKeeper: validatorKeeper,
		bankKeeper:      bankKeeper,
		authority:       authority,
	}
}

// Proposal operations

// SubmitProposal submits a new governance proposal
func (k Keeper) SubmitProposal(
	ctx sdk.Context,
	submitter string,
	proposalType string,
	title string,
	description string,
	initialDeposit math.Int,
	votingPeriodDays uint32,
	quorumRequired math.LegacyDec,
	thresholdRequired math.LegacyDec,
) (uint64, error) {
	submitterAddr, err := sdk.AccAddressFromBech32(submitter)
	if err != nil {
		return 0, types.ErrInvalidAddress
	}

	// Validate proposal type and requirements
	if err := k.validateProposalSubmission(ctx, proposalType, submitterAddr); err != nil {
		return 0, err
	}

	// Get next proposal ID
	proposalID := k.getNextProposalID(ctx)

	// Create proposal
	proposal := types.Proposal{
		ID:                proposalID,
		Type:              types.ProposalType(proposalType),
		Title:             title,
		Description:       description,
		Submitter:         submitter,
		SubmissionTime:    ctx.BlockTime(),
		DepositEndTime:    ctx.BlockTime().AddDate(0, 0, 7), // 7 days deposit period
		VotingStartTime:   ctx.BlockTime().AddDate(0, 0, 7),
		VotingEndTime:     ctx.BlockTime().AddDate(0, 0, 7+int(votingPeriodDays)),
		Status:            types.StatusDepositPeriod,
		TotalDeposit:      math.ZeroInt(),
		QuorumRequired:    quorumRequired,
		ThresholdRequired: thresholdRequired,
		CreatedAt:         ctx.BlockTime(),
		UpdatedAt:         ctx.BlockTime(),
	}

	// Store proposal
	k.setProposal(ctx, proposal)

	// Handle initial deposit if provided
	if !initialDeposit.IsZero() {
		err = k.addDeposit(ctx, proposalID, submitterAddr, initialDeposit)
		if err != nil {
			return 0, err
		}
	}

	// Index proposal
	k.indexProposal(ctx, proposal)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"submit_proposal",
			sdk.NewAttribute("proposal_id", sdk.Uint64ToBigEndian(proposalID).String()),
			sdk.NewAttribute("proposal_type", proposalType),
			sdk.NewAttribute("submitter", submitter),
			sdk.NewAttribute("title", title),
		),
	)

	return proposalID, nil
}

// Vote casts a vote on a proposal
func (k Keeper) Vote(
	ctx sdk.Context,
	voter string,
	proposalID uint64,
	option string,
	reason string,
) error {
	voterAddr, err := sdk.AccAddressFromBech32(voter)
	if err != nil {
		return types.ErrInvalidAddress
	}

	// Get proposal
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Validate voting conditions
	if err := k.validateVoting(ctx, proposal, voterAddr); err != nil {
		return err
	}

	// Check if already voted
	if k.hasVoted(ctx, proposalID, voterAddr) {
		return types.ErrAlreadyVoted
	}

	// Calculate voting power
	votingPower, err := k.calculateVotingPower(ctx, proposal, voterAddr)
	if err != nil {
		return err
	}

	if votingPower.IsZero() {
		return types.ErrInsufficientVotingPower
	}

	// Create vote
	vote := types.Vote{
		ProposalID:   proposalID,
		Voter:        voter,
		Option:       types.VoteOption(option),
		Weight:       math.LegacyOneDec(), // Full weight for simple vote
		VotingPower:  votingPower,
		Reason:       reason,
		SubmittedAt:  ctx.BlockTime(),
	}

	// Store vote
	k.setVote(ctx, vote)

	// Update tally
	k.updateTallyResult(ctx, proposal, vote, math.LegacyOneDec())

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote",
			sdk.NewAttribute("proposal_id", sdk.Uint64ToBigEndian(proposalID).String()),
			sdk.NewAttribute("voter", voter),
			sdk.NewAttribute("option", option),
			sdk.NewAttribute("voting_power", votingPower.String()),
		),
	)

	return nil
}

// VoteWeighted casts a weighted vote on a proposal
func (k Keeper) VoteWeighted(
	ctx sdk.Context,
	voter string,
	proposalID uint64,
	options []types.WeightedVoteOption,
	reason string,
) error {
	voterAddr, err := sdk.AccAddressFromBech32(voter)
	if err != nil {
		return types.ErrInvalidAddress
	}

	// Get proposal
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Validate voting conditions
	if err := k.validateVoting(ctx, proposal, voterAddr); err != nil {
		return err
	}

	// Check if already voted
	if k.hasVoted(ctx, proposalID, voterAddr) {
		return types.ErrAlreadyVoted
	}

	// Calculate voting power
	votingPower, err := k.calculateVotingPower(ctx, proposal, voterAddr)
	if err != nil {
		return err
	}

	if votingPower.IsZero() {
		return types.ErrInsufficientVotingPower
	}

	// Create weighted vote
	weightedVote := types.WeightedVote{
		ProposalID:   proposalID,
		Voter:        voter,
		Options:      options,
		VotingPower:  votingPower,
		Reason:       reason,
		SubmittedAt:  ctx.BlockTime(),
	}

	// Store weighted vote
	k.setWeightedVote(ctx, weightedVote)

	// Update tally for each option
	for _, option := range options {
		vote := types.Vote{
			ProposalID:  proposalID,
			Voter:       voter,
			Option:      types.VoteOption(option.Option),
			Weight:      option.Weight,
			VotingPower: votingPower,
		}
		k.updateTallyResult(ctx, proposal, vote, option.Weight)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"weighted_vote",
			sdk.NewAttribute("proposal_id", sdk.Uint64ToBigEndian(proposalID).String()),
			sdk.NewAttribute("voter", voter),
			sdk.NewAttribute("voting_power", votingPower.String()),
		),
	)

	return nil
}

// Deposit adds tokens to a proposal's deposit
func (k Keeper) Deposit(
	ctx sdk.Context,
	depositor string,
	proposalID uint64,
	amount math.Int,
) error {
	depositorAddr, err := sdk.AccAddressFromBech32(depositor)
	if err != nil {
		return types.ErrInvalidAddress
	}

	// Get proposal
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Validate deposit conditions
	if proposal.Status != types.StatusDepositPeriod {
		return types.ErrDepositPeriodEnded
	}

	if ctx.BlockTime().After(proposal.DepositEndTime) {
		return types.ErrDepositPeriodEnded
	}

	// Add deposit
	err = k.addDeposit(ctx, proposalID, depositorAddr, amount)
	if err != nil {
		return err
	}

	// Check if minimum deposit reached
	params := k.getGovernanceParams(ctx)
	if proposal.TotalDeposit.GTE(params.MinDeposit) && proposal.Status == types.StatusDepositPeriod {
		// Move to voting period
		proposal.Status = types.StatusVotingPeriod
		proposal.VotingStartTime = ctx.BlockTime()
		k.setProposal(ctx, proposal)
		k.updateProposalIndex(ctx, proposal)
	}

	return nil
}

// CancelProposal cancels a proposal (only by submitter during deposit period)
func (k Keeper) CancelProposal(
	ctx sdk.Context,
	proposer string,
	proposalID uint64,
	reason string,
) error {
	// Get proposal
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Validate cancellation
	if proposal.Submitter != proposer {
		return types.ErrNotAuthorized
	}

	if proposal.Status != types.StatusDepositPeriod {
		return types.ErrInvalidProposalStatus
	}

	// Cancel proposal
	proposal.Status = types.StatusCancelled
	proposal.UpdatedAt = ctx.BlockTime()
	k.setProposal(ctx, proposal)
	k.updateProposalIndex(ctx, proposal)

	// Refund deposits
	k.refundProposalDeposits(ctx, proposalID)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"cancel_proposal",
			sdk.NewAttribute("proposal_id", sdk.Uint64ToBigEndian(proposalID).String()),
			sdk.NewAttribute("proposer", proposer),
			sdk.NewAttribute("reason", reason),
		),
	)

	return nil
}

// Helper functions for validation and calculations

func (k Keeper) validateProposalSubmission(ctx sdk.Context, proposalType string, submitter sdk.AccAddress) error {
	switch proposalType {
	case "company_listing", "company_delisting", "company_parameter":
		// Company proposals require validator tier Bronze or higher
		validator, found := k.validatorKeeper.GetValidator(ctx, submitter.String())
		if !found {
			return types.ErrNotAuthorized
		}
		tier := k.validatorKeeper.GetValidatorTier(ctx, validator.Address)
		if tier < validatortypes.TierBronze {
			return types.ErrInsufficientPermissions
		}

	case "validator_promotion", "validator_demotion", "validator_removal":
		// Validator proposals require existing Gold tier validator
		validator, found := k.validatorKeeper.GetValidator(ctx, submitter.String())
		if !found {
			return types.ErrNotAuthorized
		}
		tier := k.validatorKeeper.GetValidatorTier(ctx, validator.Address)
		if tier < validatortypes.TierGold {
			return types.ErrInsufficientPermissions
		}

	case "protocol_parameter", "protocol_upgrade":
		// Protocol proposals require Gold tier validator or governance authority
		if submitter.String() != k.authority {
			validator, found := k.validatorKeeper.GetValidator(ctx, submitter.String())
			if !found {
				return types.ErrNotAuthorized
			}
			tier := k.validatorKeeper.GetValidatorTier(ctx, validator.Address)
			if tier < validatortypes.TierGold {
				return types.ErrInsufficientPermissions
			}
		}

	case "treasury_spend":
		// Treasury proposals require significant HODL stake
		balance := k.hodlKeeper.GetBalance(ctx, submitter)
		minStake := math.NewInt(100000) // 100,000 HODL minimum
		if balance.LT(minStake) {
			return types.ErrInsufficientPermissions
		}

	case "emergency_action":
		// Emergency proposals require Gold tier validator
		validator, found := k.validatorKeeper.GetValidator(ctx, submitter.String())
		if !found {
			return types.ErrNotAuthorized
		}
		tier := k.validatorKeeper.GetValidatorTier(ctx, validator.Address)
		if tier < validatortypes.TierGold {
			return types.ErrInsufficientPermissions
		}

	default:
		return types.ErrInvalidProposalType
	}

	return nil
}

func (k Keeper) validateVoting(ctx sdk.Context, proposal types.Proposal, voter sdk.AccAddress) error {
	// Check proposal status
	if proposal.Status != types.StatusVotingPeriod {
		return types.ErrProposalNotActive
	}

	// Check voting period
	if ctx.BlockTime().Before(proposal.VotingStartTime) {
		return types.ErrVotingPeriodNotStarted
	}

	if ctx.BlockTime().After(proposal.VotingEndTime) {
		return types.ErrVotingPeriodEnded
	}

	return nil
}

func (k Keeper) calculateVotingPower(ctx sdk.Context, proposal types.Proposal, voter sdk.AccAddress) (math.LegacyDec, error) {
	// Voting power calculation based on proposal type
	switch proposal.Type {
	case types.ProposalTypeCompanyListing, types.ProposalTypeCompanyDelisting, types.ProposalTypeCompanyParameter:
		// For company proposals, voting power based on validator tier and HODL stake
		validator, isValidator := k.validatorKeeper.GetValidator(ctx, voter.String())
		if isValidator {
			tier := k.validatorKeeper.GetValidatorTier(ctx, validator.Address)
			tierMultiplier := k.getValidatorTierMultiplier(tier)
			hodlBalance := k.hodlKeeper.GetBalance(ctx, voter)
			stakePower := math.LegacyNewDecFromInt(hodlBalance).Quo(math.LegacyNewDec(1000)) // 1000 HODL = 1 vote
			return stakePower.Mul(tierMultiplier), nil
		} else {
			// Non-validators have reduced voting power
			hodlBalance := k.hodlKeeper.GetBalance(ctx, voter)
			return math.LegacyNewDecFromInt(hodlBalance).Quo(math.LegacyNewDec(5000)), nil // 5000 HODL = 1 vote
		}

	case types.ProposalTypeValidatorPromotion, types.ProposalTypeValidatorDemotion, types.ProposalTypeValidatorRemoval:
		// Validator proposals: only validators can vote
		validator, isValidator := k.validatorKeeper.GetValidator(ctx, voter.String())
		if !isValidator {
			return math.LegacyZeroDec(), nil
		}
		tier := k.validatorKeeper.GetValidatorTier(ctx, validator.Address)
		return k.getValidatorTierMultiplier(tier), nil

	case types.ProposalTypeProtocolParameter, types.ProposalTypeProtocolUpgrade:
		// Protocol proposals: combined validator tier and stake
		validator, isValidator := k.validatorKeeper.GetValidator(ctx, voter.String())
		hodlBalance := k.hodlKeeper.GetBalance(ctx, voter)
		basePower := math.LegacyNewDecFromInt(hodlBalance).Quo(math.LegacyNewDec(2000)) // 2000 HODL = 1 vote
		
		if isValidator {
			tier := k.validatorKeeper.GetValidatorTier(ctx, validator.Address)
			tierMultiplier := k.getValidatorTierMultiplier(tier)
			return basePower.Mul(tierMultiplier), nil
		}
		return basePower, nil

	case types.ProposalTypeTreasurySpend:
		// Treasury proposals: HODL stake voting
		hodlBalance := k.hodlKeeper.GetBalance(ctx, voter)
		return math.LegacyNewDecFromInt(hodlBalance).Quo(math.LegacyNewDec(1000)), nil // 1000 HODL = 1 vote

	case types.ProposalTypeEmergencyAction:
		// Emergency proposals: only Gold tier validators
		validator, isValidator := k.validatorKeeper.GetValidator(ctx, voter.String())
		if !isValidator {
			return math.LegacyZeroDec(), nil
		}
		tier := k.validatorKeeper.GetValidatorTier(ctx, validator.Address)
		if tier < validatortypes.TierGold {
			return math.LegacyZeroDec(), nil
		}
		return math.LegacyNewDec(1), nil

	default:
		return math.LegacyZeroDec(), types.ErrInvalidProposalType
	}
}

func (k Keeper) getValidatorTierMultiplier(tier validatortypes.ValidatorTier) math.LegacyDec {
	switch tier {
	case validatortypes.TierBronze:
		return math.LegacyNewDec(2)
	case validatortypes.TierSilver:
		return math.LegacyNewDec(5)
	case validatortypes.TierGold:
		return math.LegacyNewDec(10)
	case validatortypes.TierPlatinum:
		return math.LegacyNewDec(20)
	default:
		return math.LegacyOneDec()
	}
}