package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	equitytypes "github.com/sharehodl/sharehodl-blockchain/x/equity/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
)

// Keeper handles governance operations
type Keeper struct {
	cdc          codec.Codec
	storeService store.KVStoreService

	// Module keepers for integration
	equityKeeper         EquityKeeper
	hodlKeeper           HODLKeeper
	stakingKeeper        UniversalStakingKeeper
	bankKeeper           BankKeeper
	accountKeeper        AccountKeeper
	feeAbstractionKeeper FeeAbstractionKeeper

	// Authority for governance parameter updates
	authority string
}

// Interface definitions for module dependencies
type EquityKeeper interface {
	GetCompany(ctx context.Context, id uint64) (interface{}, bool)
	GetShareholding(ctx context.Context, companyID uint64, classID, shareholder string) (interface{}, bool)
	GetTotalShares(ctx sdk.Context, companyID uint64, classID string) math.Int
	SetDefaultCharityWallet(ctx sdk.Context, address string) error
	GetDefaultCharityWallet(ctx sdk.Context) string
	// GetBeneficialOwnershipsByOwner returns all beneficial ownership records for an owner
	// This enables LP providers to vote on company matters with their locked shares
	GetBeneficialOwnershipsByOwner(ctx sdk.Context, companyID uint64, classID, owner string) ([]equitytypes.BeneficialOwnership, error)
	// GetVotingSharesForCompany returns total voting shares (direct + beneficial ownership)
	// for a voter in a specific company - used for company proposal voting
	GetVotingSharesForCompany(ctx sdk.Context, companyID uint64, voter string) math.Int
	// ApproveDividendDistribution approves a dividend for distribution (governance only)
	ApproveDividendDistribution(ctx sdk.Context, dividendID uint64, proposalID uint64) error
	// RejectDividendDistribution rejects a dividend distribution (governance only)
	RejectDividendDistribution(ctx sdk.Context, dividendID uint64, proposalID uint64, reason string) error
}

// BeneficialOwnership is an alias to equitytypes.BeneficialOwnership for local use
type BeneficialOwnership = equitytypes.BeneficialOwnership

type HODLKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress) math.Int
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amount math.Int) error
	BurnTokens(ctx sdk.Context, addr sdk.AccAddress, amount math.Int) error
	// SECURITY FIX: Added for proper voting power calculation
	GetTotalSupply(ctx context.Context) interface{}
}

// ValidatorInfo is a simplified struct for validator data
type ValidatorInfo struct {
	Address string
	Active  bool
	Tier    int
}

type UniversalStakingKeeper interface {
	// GetUserTierInt returns the user's tier as int (0=Holder, 1=Keeper, 2=Warden, 3=Steward, 4=Archon, 5=Validator)
	GetUserTierInt(ctx sdk.Context, addr sdk.AccAddress) int
	// GetTotalStaked returns the total amount staked across all users
	GetTotalStaked(ctx sdk.Context) math.Int
}

// GetValidatorByAddress returns validator info by address (helper method for keeper)
func (k Keeper) GetValidatorByAddress(ctx sdk.Context, addr string) ValidatorInfo {
	if k.stakingKeeper == nil {
		return ValidatorInfo{Address: addr, Active: false, Tier: 0}
	}
	accAddr, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return ValidatorInfo{Address: addr, Active: false, Tier: 0}
	}
	tier := k.stakingKeeper.GetUserTierInt(ctx, accAddr)
	active := tier >= 5 // TierValidator = 5
	return ValidatorInfo{
		Address: addr,
		Active:  active,
		Tier:    tier,
	}
}

type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
}

// AccountKeeper defines the account keeper interface for module account detection
// Used to prevent treasury/escrow/DEX module accounts from voting
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// FeeAbstractionKeeper handles fee abstraction for users without HODL
// Allows shareholders to pay governance fees using equity tokens
type FeeAbstractionKeeper interface {
	// ProcessFeeAbstraction swaps equity for HODL to cover fees
	// Returns the total HODL processed (fee + markup)
	ProcessFeeAbstraction(ctx sdk.Context, payer sdk.AccAddress, requiredFeeUhodl math.Int) (math.Int, error)
	// IsEnabled returns whether fee abstraction is enabled
	IsEnabled(ctx sdk.Context) bool
	// WithdrawFromTreasury withdraws funds from treasury (governance only)
	WithdrawFromTreasury(ctx sdk.Context, recipient sdk.AccAddress, amount math.Int) error
	// GetTreasuryBalance returns the current treasury balance
	GetTreasuryBalance(ctx sdk.Context) math.Int
}

// =============================================================================
// PARAMS MANAGEMENT - All values governance-controllable
// =============================================================================

// GetExtendedParams returns the governance extended parameters
func (k Keeper) GetExtendedParams(ctx sdk.Context) types.ExtendedParams {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsPrefix)
	if err != nil || bz == nil {
		return types.DefaultExtendedParams()
	}

	var params types.ExtendedParams
	if err := json.Unmarshal(bz, &params); err != nil {
		return types.DefaultExtendedParams()
	}
	return params
}

// SetExtendedParams sets the governance extended parameters
func (k Keeper) SetExtendedParams(ctx sdk.Context, params types.ExtendedParams) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return store.Set(types.ParamsPrefix, bz)
}

// Convenience getters for governance parameters

// GetVotingPeriodDays returns governance-controllable voting period
func (k Keeper) GetVotingPeriodDays(ctx sdk.Context) uint64 {
	return k.GetExtendedParams(ctx).VotingPeriodDays
}

// GetValidatorVotingPeriodDays returns governance-controllable validator voting period
func (k Keeper) GetValidatorVotingPeriodDays(ctx sdk.Context) uint64 {
	return k.GetExtendedParams(ctx).ValidatorVotingPeriodDays
}

// GetQuorum returns governance-controllable quorum threshold
func (k Keeper) GetQuorum(ctx sdk.Context) math.LegacyDec {
	return k.GetExtendedParams(ctx).GetQuorum()
}

// GetCompanyQuorum returns governance-controllable company quorum
func (k Keeper) GetCompanyQuorum(ctx sdk.Context) math.LegacyDec {
	return k.GetExtendedParams(ctx).GetCompanyQuorum()
}

// GetValidatorQuorum returns governance-controllable validator quorum
func (k Keeper) GetValidatorQuorum(ctx sdk.Context) math.LegacyDec {
	return k.GetExtendedParams(ctx).GetValidatorQuorum()
}

// GetProtocolQuorum returns governance-controllable protocol quorum
func (k Keeper) GetProtocolQuorum(ctx sdk.Context) math.LegacyDec {
	return k.GetExtendedParams(ctx).GetProtocolQuorum()
}

// GetThreshold returns governance-controllable pass threshold
func (k Keeper) GetThreshold(ctx sdk.Context) math.LegacyDec {
	return k.GetExtendedParams(ctx).GetThreshold()
}

// GetCompanyThreshold returns governance-controllable company threshold
func (k Keeper) GetCompanyThreshold(ctx sdk.Context) math.LegacyDec {
	return k.GetExtendedParams(ctx).GetCompanyThreshold()
}

// GetValidatorThreshold returns governance-controllable validator threshold
func (k Keeper) GetValidatorThreshold(ctx sdk.Context) math.LegacyDec {
	return k.GetExtendedParams(ctx).GetValidatorThreshold()
}

// GetVetoThreshold returns governance-controllable veto threshold
func (k Keeper) GetVetoThreshold(ctx sdk.Context) math.LegacyDec {
	return k.GetExtendedParams(ctx).GetVetoThreshold()
}

// IsGovernanceEnabled returns whether governance is enabled
func (k Keeper) IsGovernanceEnabled(ctx sdk.Context) bool {
	return k.GetExtendedParams(ctx).GovernanceEnabled
}

// =============================================================================
// ANTI-SPAM PARAMETER GETTERS
// =============================================================================

// GetMinDeposit returns the minimum deposit required for proposals
func (k Keeper) GetMinDeposit(ctx sdk.Context) math.Int {
	return math.NewIntFromUint64(k.GetExtendedParams(ctx).MinDeposit)
}

// GetProposalFee returns the non-refundable proposal fee (burned on submission)
func (k Keeper) GetProposalFee(ctx sdk.Context) math.Int {
	return math.NewIntFromUint64(k.GetExtendedParams(ctx).ProposalFee)
}

// GetMaxProposalsPerDay returns the maximum proposals allowed per address per day
func (k Keeper) GetMaxProposalsPerDay(ctx sdk.Context) uint64 {
	return k.GetExtendedParams(ctx).MaxProposalsPerDay
}

// GetProposalCooldownHours returns the cooldown period between proposals in hours
func (k Keeper) GetProposalCooldownHours(ctx sdk.Context) uint64 {
	return k.GetExtendedParams(ctx).ProposalCooldownHours
}

// GetMinProposerStake returns the minimum HODL balance required to submit proposals
func (k Keeper) GetMinProposerStake(ctx sdk.Context) math.Int {
	return math.NewIntFromUint64(k.GetExtendedParams(ctx).MinProposerStake)
}

// GetDepositPeriodDays returns the deposit period in days
func (k Keeper) GetDepositPeriodDays(ctx sdk.Context) uint64 {
	return k.GetExtendedParams(ctx).DepositPeriodDays
}

// NewKeeper creates a new governance Keeper
func NewKeeper(
	cdc codec.Codec,
	storeService store.KVStoreService,
	equityKeeper EquityKeeper,
	hodlKeeper HODLKeeper,
	stakingKeeper UniversalStakingKeeper,
	bankKeeper BankKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeService:  storeService,
		equityKeeper:  equityKeeper,
		hodlKeeper:    hodlKeeper,
		stakingKeeper: stakingKeeper,
		bankKeeper:    bankKeeper,
		authority:     authority,
	}
}

// SetStakingKeeper sets the staking keeper (for late binding during app initialization)
func (k *Keeper) SetStakingKeeper(stakingKeeper UniversalStakingKeeper) {
	k.stakingKeeper = stakingKeeper
}

// SetFeeAbstractionKeeper sets the fee abstraction keeper (for late binding)
// This allows shareholders without HODL to pay proposal fees using equity tokens
func (k *Keeper) SetFeeAbstractionKeeper(feeAbsKeeper FeeAbstractionKeeper) {
	k.feeAbstractionKeeper = feeAbsKeeper
}

// SetAccountKeeper sets the account keeper (for late binding)
// Required for module account detection to prevent treasuries from voting
func (k *Keeper) SetAccountKeeper(accountKeeper AccountKeeper) {
	k.accountKeeper = accountKeeper
}

// Proposal operations

// SubmitProposal submits a new governance proposal
func (k Keeper) SubmitProposal(
	ctx sdk.Context,
	submitter string,
	proposalType types.ProposalType,
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

	// Validate proposal type and requirements (includes anti-spam checks)
	if err := k.validateProposalSubmission(ctx, proposalType.String(), submitterAddr); err != nil {
		return 0, err
	}

	// Charge and burn proposal fee (non-refundable anti-spam measure)
	if err := k.chargeProposalFee(ctx, submitterAddr); err != nil {
		return 0, err
	}

	// Update rate limiting counters
	k.updateProposerRateLimiting(ctx, submitterAddr)

	// Get next proposal ID
	proposalID := k.getNextProposalID(ctx)

	// Get parameters for deposit period and min deposit
	params := k.GetExtendedParams(ctx)
	depositPeriodDays := int(params.DepositPeriodDays)
	minDeposit := math.NewIntFromUint64(params.MinDeposit)

	// Create proposal
	proposal := types.Proposal{
		ID:               proposalID,
		Type:             proposalType,
		Title:            title,
		Description:      description,
		Proposer:         submitter,
		DepositEndTime:   ctx.BlockTime().AddDate(0, 0, depositPeriodDays),
		VotingStartTime:  ctx.BlockTime().AddDate(0, 0, depositPeriodDays),
		VotingEndTime:    ctx.BlockTime().AddDate(0, 0, depositPeriodDays+int(votingPeriodDays)),
		Status:           types.ProposalStatusDepositPeriod,
		TotalDeposit:     math.ZeroInt(),
		MinDeposit:       minDeposit,
		YesVotes:         math.ZeroInt(),
		NoVotes:          math.ZeroInt(),
		AbstainVotes:     math.ZeroInt(),
		NoWithVetoVotes:  math.ZeroInt(),
		TotalVotingPower: math.ZeroInt(),
		Quorum:           quorumRequired,
		Threshold:        thresholdRequired,
		VetoThreshold:    math.LegacyNewDecWithPrec(334, 3),
		CreatedAt:        ctx.BlockTime(),
		UpdatedAt:        ctx.BlockTime(),
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
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("proposal_type", proposalType.String()),
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
	option types.VoteOption,
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
		ProposalID:    proposalID,
		Voter:         voter,
		Option:        option,
		Weight:        math.LegacyOneDec(),
		VotingPower:   votingPower.TruncateInt(),
		Justification: reason,
		VotedAt:       ctx.BlockTime(),
	}

	// Store vote
	k.setVote(ctx, vote)

	// Update tally
	k.updateTallyResult(ctx, proposal, vote, math.LegacyOneDec())

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("voter", voter),
			sdk.NewAttribute("option", option.String()),
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
		ProposalID:  proposalID,
		Voter:       voter,
		Options:     options,
		VotingPower: votingPower.TruncateInt(),
		VotedAt:     ctx.BlockTime(),
	}

	// Store weighted vote
	k.setWeightedVote(ctx, weightedVote)

	// Update tally for each option
	for _, opt := range options {
		vote := types.Vote{
			ProposalID:  proposalID,
			Voter:       voter,
			Option:      opt.Option,
			Weight:      opt.Weight,
			VotingPower: votingPower.TruncateInt(),
			VotedAt:     ctx.BlockTime(),
		}
		k.updateTallyResult(ctx, proposal, vote, opt.Weight)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"weighted_vote",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
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
	if proposal.Status != types.ProposalStatusDepositPeriod {
		return types.ErrDepositPeriodEnded
	}

	if ctx.BlockTime().After(proposal.DepositEndTime) {
		return types.ErrDepositPeriodEnded
	}

	// Add deposit
	return k.addDeposit(ctx, proposalID, depositorAddr, amount)
}

// GetProposal retrieves a proposal by ID
func (k Keeper) GetProposal(ctx sdk.Context, proposalID uint64) (types.Proposal, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.ProposalKey(proposalID)

	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return types.Proposal{}, false
	}

	var proposal types.Proposal
	if err := json.Unmarshal(bz, &proposal); err != nil {
		return types.Proposal{}, false
	}
	return proposal, true
}

// setProposal stores a proposal
func (k Keeper) setProposal(ctx sdk.Context, proposal types.Proposal) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.ProposalKey(proposal.ID)
	bz, _ := json.Marshal(proposal)
	store.Set(key, bz)

	// PERFORMANCE FIX: Index by end time for efficient end block processing
	if proposal.Status == types.ProposalStatusDepositPeriod {
		// Index by deposit end time
		depositEndKey := types.DepositEndTimeKey(proposal.DepositEndTime.Unix(), proposal.ID)
		store.Set(depositEndKey, []byte{0x01})
	} else if proposal.Status == types.ProposalStatusVotingPeriod {
		// Index by voting end time
		votingEndKey := types.VotingEndTimeKey(proposal.VotingEndTime.Unix(), proposal.ID)
		store.Set(votingEndKey, []byte{0x01})
	}
}

// setVote stores a vote
func (k Keeper) setVote(ctx sdk.Context, vote types.Vote) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.VoteKey(vote.ProposalID, []byte(vote.Voter))
	bz, _ := json.Marshal(vote)
	store.Set(key, bz)
}

// setWeightedVote stores a weighted vote
func (k Keeper) setWeightedVote(ctx sdk.Context, vote types.WeightedVote) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.WeightedVoteKey(vote.ProposalID, []byte(vote.Voter))
	bz, _ := json.Marshal(vote)
	store.Set(key, bz)
}

// hasVoted checks if an address has already voted on a proposal
func (k Keeper) hasVoted(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) bool {
	store := k.storeService.OpenKVStore(ctx)
	key := types.VoteKey(proposalID, voter.Bytes())
	has, _ := store.Has(key)
	return has
}

// getNextProposalID returns the next proposal ID
func (k Keeper) getNextProposalID(ctx sdk.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ProposalIDCounterPrefix)
	if err != nil || bz == nil {
		// Initialize counter
		nextBz, _ := json.Marshal(uint64(2))
		store.Set(types.ProposalIDCounterPrefix, nextBz)
		return 1
	}
	var id uint64
	if err := json.Unmarshal(bz, &id); err != nil {
		return 1
	}
	// Increment and store
	nextID := id + 1
	nextBz, _ := json.Marshal(nextID)
	store.Set(types.ProposalIDCounterPrefix, nextBz)
	return id
}

// addDeposit adds a deposit to a proposal
func (k Keeper) addDeposit(ctx sdk.Context, proposalID uint64, depositor sdk.AccAddress, amount math.Int) error {
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Update proposal total deposit
	proposal.TotalDeposit = proposal.TotalDeposit.Add(amount)

	// Add to deposits list
	deposit := types.Deposit{
		ProposalID:  proposalID,
		Depositor:   depositor.String(),
		Amount:      amount,
		Currency:    "uhodl",
		DepositedAt: ctx.BlockTime(),
	}
	proposal.Deposits = append(proposal.Deposits, deposit)

	// Check if min deposit is met
	if proposal.TotalDeposit.GTE(proposal.MinDeposit) && proposal.Status == types.ProposalStatusDepositPeriod {
		proposal.Status = types.ProposalStatusVotingPeriod
		proposal.VotingStartTime = ctx.BlockTime()
	}

	proposal.UpdatedAt = ctx.BlockTime()
	k.setProposal(ctx, proposal)

	return nil
}

// indexProposal creates indexes for a proposal
func (k Keeper) indexProposal(ctx sdk.Context, proposal types.Proposal) {
	// Simple index by type
	store := k.storeService.OpenKVStore(ctx)
	typeKey := types.ProposalByTypeKey(proposal.Type.String(), proposal.ID)
	store.Set(typeKey, []byte{1})
}

// validateProposalSubmission validates proposal submission with anti-spam checks
func (k Keeper) validateProposalSubmission(ctx sdk.Context, proposalType string, submitter sdk.AccAddress) error {
	params := k.GetExtendedParams(ctx)

	// 1. Check minimum stake requirement
	if params.MinProposerStake > 0 {
		balance := k.hodlKeeper.GetBalance(ctx, submitter)
		minStake := math.NewIntFromUint64(params.MinProposerStake)
		if balance.LT(minStake) {
			return types.ErrInsufficientStake.Wrapf("required: %s uhodl, have: %s uhodl", minStake.String(), balance.String())
		}
	}

	// 2. Check rate limiting - max proposals per day
	if params.MaxProposalsPerDay > 0 {
		dayTimestamp := types.GetDayTimestamp(ctx.BlockTime().Unix())
		dailyCount := k.getProposerDailyCount(ctx, submitter, dayTimestamp)
		if dailyCount >= params.MaxProposalsPerDay {
			return types.ErrRateLimitExceeded.Wrapf("max %d proposals per day, already submitted %d", params.MaxProposalsPerDay, dailyCount)
		}
	}

	// 3. Check cooldown period
	if params.ProposalCooldownHours > 0 {
		lastSubmit := k.getProposerLastSubmitTime(ctx, submitter)
		if lastSubmit > 0 {
			cooldownSeconds := int64(params.ProposalCooldownHours * 3600)
			nextAllowed := lastSubmit + cooldownSeconds
			if ctx.BlockTime().Unix() < nextAllowed {
				remaining := nextAllowed - ctx.BlockTime().Unix()
				hours := remaining / 3600
				return types.ErrCooldownNotElapsed.Wrapf("must wait %d more hours before next proposal", hours+1)
			}
		}
	}

	return nil
}

// chargeProposalFee charges and burns the non-refundable proposal fee
// Supports fee abstraction: if user lacks HODL, they can pay with equity tokens
func (k Keeper) chargeProposalFee(ctx sdk.Context, submitter sdk.AccAddress) error {
	params := k.GetExtendedParams(ctx)

	if params.ProposalFee == 0 {
		return nil // No fee required
	}

	feeAmount := math.NewIntFromUint64(params.ProposalFee)

	// Check submitter's HODL balance
	balance := k.hodlKeeper.GetBalance(ctx, submitter)

	// If insufficient HODL, try fee abstraction (swap equity for HODL)
	if balance.LT(feeAmount) {
		shortfall := feeAmount.Sub(balance)

		// Check if fee abstraction is available
		if k.feeAbstractionKeeper != nil && k.feeAbstractionKeeper.IsEnabled(ctx) {
			// Use fee abstraction to swap equity for HODL
			_, err := k.feeAbstractionKeeper.ProcessFeeAbstraction(ctx, submitter, shortfall)
			if err != nil {
				// Fee abstraction failed - provide helpful error
				return types.ErrInsufficientProposalFee.Wrapf(
					"need %s uhodl, have %s uhodl; fee abstraction failed: %v",
					feeAmount.String(), balance.String(), err,
				)
			}

			// Fee abstraction succeeded - emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"proposal_fee_abstraction",
					sdk.NewAttribute("submitter", submitter.String()),
					sdk.NewAttribute("hodl_shortfall", shortfall.String()),
				),
			)
		} else {
			// No fee abstraction available
			return types.ErrInsufficientProposalFee.Wrapf(
				"required: %s uhodl, have: %s uhodl",
				feeAmount.String(), balance.String(),
			)
		}
	}

	// Now burn the fee (user has enough HODL after fee abstraction)
	if err := k.hodlKeeper.BurnTokens(ctx, submitter, feeAmount); err != nil {
		return types.ErrFeeBurnFailed.Wrap(err.Error())
	}

	// Track total burned fees
	k.addBurnedFees(ctx, feeAmount)

	// Emit event for fee burn
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"proposal_fee_burned",
			sdk.NewAttribute("submitter", submitter.String()),
			sdk.NewAttribute("amount", feeAmount.String()),
		),
	)

	return nil
}

// updateProposerRateLimiting updates the rate limiting counters after successful submission
func (k Keeper) updateProposerRateLimiting(ctx sdk.Context, submitter sdk.AccAddress) {
	store := k.storeService.OpenKVStore(ctx)
	currentTime := ctx.BlockTime().Unix()
	dayTimestamp := types.GetDayTimestamp(currentTime)

	// Increment daily count
	currentCount := k.getProposerDailyCount(ctx, submitter, dayTimestamp)
	newCount := currentCount + 1
	countKey := types.ProposerDailyCountKey(submitter.Bytes(), dayTimestamp)
	countBz, _ := json.Marshal(newCount)
	store.Set(countKey, countBz)

	// Update last submit time
	lastSubmitKey := types.ProposerLastSubmitKey(submitter.Bytes())
	timeBz, _ := json.Marshal(currentTime)
	store.Set(lastSubmitKey, timeBz)
}

// getProposerDailyCount returns the number of proposals submitted by an address today
func (k Keeper) getProposerDailyCount(ctx sdk.Context, submitter sdk.AccAddress, dayTimestamp int64) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	key := types.ProposerDailyCountKey(submitter.Bytes(), dayTimestamp)
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return 0
	}
	var count uint64
	if err := json.Unmarshal(bz, &count); err != nil {
		return 0
	}
	return count
}

// getProposerLastSubmitTime returns the timestamp of the last proposal submission
func (k Keeper) getProposerLastSubmitTime(ctx sdk.Context, submitter sdk.AccAddress) int64 {
	store := k.storeService.OpenKVStore(ctx)
	key := types.ProposerLastSubmitKey(submitter.Bytes())
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return 0
	}
	var timestamp int64
	if err := json.Unmarshal(bz, &timestamp); err != nil {
		return 0
	}
	return timestamp
}

// addBurnedFees adds to the total burned fees counter
func (k Keeper) addBurnedFees(ctx sdk.Context, amount math.Int) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.BurnedFeesKey()

	var totalBurned math.Int
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		totalBurned = math.ZeroInt()
	} else {
		if err := json.Unmarshal(bz, &totalBurned); err != nil {
			totalBurned = math.ZeroInt()
		}
	}

	totalBurned = totalBurned.Add(amount)
	newBz, _ := json.Marshal(totalBurned)
	store.Set(key, newBz)
}

// GetTotalBurnedFees returns the total fees burned from proposal submissions
func (k Keeper) GetTotalBurnedFees(ctx sdk.Context) math.Int {
	store := k.storeService.OpenKVStore(ctx)
	key := types.BurnedFeesKey()
	bz, err := store.Get(key)
	if err != nil || bz == nil {
		return math.ZeroInt()
	}
	var totalBurned math.Int
	if err := json.Unmarshal(bz, &totalBurned); err != nil {
		return math.ZeroInt()
	}
	return totalBurned
}

// validateVoting validates voting conditions including voter eligibility
// Different proposal types have different eligibility requirements:
// - Validator proposals: Only validators (tier >= 5) can vote
// - Company proposals: Only shareholders of that specific company can vote
// - Protocol proposals: Anyone with HODL balance can vote
func (k Keeper) validateVoting(ctx sdk.Context, proposal types.Proposal, voter sdk.AccAddress) error {
	if proposal.Status != types.ProposalStatusVotingPeriod {
		return types.ErrProposalNotActive
	}

	if ctx.BlockTime().Before(proposal.VotingStartTime) {
		return types.ErrVotingPeriodNotStarted
	}

	if ctx.BlockTime().After(proposal.VotingEndTime) {
		return types.ErrVotingPeriodEnded
	}

	// Check voter eligibility based on proposal type
	if err := k.checkVoterEligibility(ctx, proposal, voter); err != nil {
		return err
	}

	return nil
}

// checkVoterEligibility verifies that the voter is eligible based on proposal type
func (k Keeper) checkVoterEligibility(ctx sdk.Context, proposal types.Proposal, voter sdk.AccAddress) error {
	// ==========================================================================
	// CRITICAL: Block module accounts from voting
	// Treasury, escrow, DEX, and other module accounts must not vote
	// ==========================================================================
	if k.isModuleAccount(voter) {
		return types.ErrNotEligibleToVote.Wrapf(
			"module accounts cannot vote",
		)
	}

	switch proposal.Type {
	// ==========================================================================
	// VALIDATOR-ONLY PROPOSALS
	// Only active validators (tier 5) can vote on these
	// ==========================================================================
	case types.ProposalTypeSoftwareUpgrade,
		types.ProposalTypeValidatorTierChange,
		types.ProposalTypeValidatorPromotion,
		types.ProposalTypeValidatorDemotion,
		types.ProposalTypeValidatorRemoval,
		types.ProposalTypeEmergencyAction:

		tier := k.stakingKeeper.GetUserTierInt(ctx, voter)
		if tier < 5 { // TierValidator = 5
			return types.ErrValidatorRequired.Wrapf(
				"voter tier %d, required tier 5 (Validator)",
				tier,
			)
		}

	// ==========================================================================
	// COMPANY SHAREHOLDER PROPOSALS
	// Only shareholders of the specific company can vote
	// ==========================================================================
	case types.ProposalTypeCompanyGovernance,
		types.ProposalTypeCompanyListing,
		types.ProposalTypeCompanyDelisting,
		types.ProposalTypeCompanyParameter:

		// Must have a company ID set
		if proposal.CompanyID == 0 {
			// If no company ID, fall back to HODL holder requirement
			return k.checkHODLBalance(ctx, voter)
		}

		// Check if voter holds shares in this company (direct + beneficial)
		equityShares := k.equityKeeper.GetVotingSharesForCompany(ctx, proposal.CompanyID, voter.String())
		if equityShares.IsZero() {
			return types.ErrShareholderRequired.Wrapf(
				"voter has no shares in company %d",
				proposal.CompanyID,
			)
		}

	// ==========================================================================
	// PROTOCOL/TREASURY PROPOSALS
	// Anyone with HODL balance can vote
	// ==========================================================================
	case types.ProposalTypeParameterChange,
		types.ProposalTypeCommunityPoolSpend,
		types.ProposalTypeSetCharityWallet,
		types.ProposalTypeFeeAbstractionParams,
		types.ProposalTypeTreasuryGrant,
		types.ProposalTypeTreasuryFunding,
		types.ProposalTypeTreasurySpend,
		types.ProposalTypeProtocolParameter,
		types.ProposalTypeProtocolUpgrade,
		types.ProposalTypeListingRequirement,
		types.ProposalTypeTradingHalt:

		return k.checkHODLBalance(ctx, voter)

	// ==========================================================================
	// TEXT PROPOSALS
	// Anyone with HODL balance or shares in related company can vote
	// ==========================================================================
	case types.ProposalTypeText:
		// If company-specific text proposal
		if proposal.CompanyID > 0 {
			equityShares := k.equityKeeper.GetVotingSharesForCompany(ctx, proposal.CompanyID, voter.String())
			if !equityShares.IsZero() {
				return nil // Shareholder can vote
			}
		}
		// Otherwise, require HODL balance
		return k.checkHODLBalance(ctx, voter)

	default:
		// Unknown proposal type - require HODL balance as fallback
		return k.checkHODLBalance(ctx, voter)
	}

	return nil
}

// checkHODLBalance verifies voter has non-zero HODL balance
func (k Keeper) checkHODLBalance(ctx sdk.Context, voter sdk.AccAddress) error {
	hodlBalance := k.hodlKeeper.GetBalance(ctx, voter)
	if hodlBalance.IsZero() {
		return types.ErrHODLHolderRequired.Wrapf(
			"voter has zero HODL balance",
		)
	}
	return nil
}

// isModuleAccount checks if the given address is a known module account
// Module accounts (treasury, escrow, DEX, lending, etc.) must not vote
// This is a defense-in-depth check - module accounts also fail other checks
func (k Keeper) isModuleAccount(voter sdk.AccAddress) bool {
	if k.accountKeeper == nil {
		return false // No account keeper available, skip check
	}

	// List of module names whose accounts should be blocked from voting
	// These are module accounts that hold user funds but should not vote
	blockedModules := []string{
		"equity",           // Company treasury
		"hodl",             // HODL stablecoin module
		"dex",              // DEX orderbook and LP pools
		"escrow",           // Escrow holdings
		"lending",          // Lending collateral
		"staking",          // Staked funds
		"distribution",     // Reward distribution
		"fee_collector",    // Collected fees
		"bonded_tokens_pool",
		"not_bonded_tokens_pool",
		"governance",       // Governance deposits
		"feeabstraction",   // Fee abstraction treasury
		"bridge",           // Bridge reserves
		"inheritance",      // Inheritance holdings
	}

	for _, moduleName := range blockedModules {
		moduleAddr := k.accountKeeper.GetModuleAddress(moduleName)
		if moduleAddr != nil && voter.Equals(moduleAddr) {
			return true
		}
	}

	return false
}

// calculateVotingPower calculates voting power for an address
// Voting power sources:
// 1. HODL balance - base voting power for all proposals
// 2. Equity shares - for company-specific proposals, shareholders can vote with their shares
// 3. Tier multiplier - staking tier provides a multiplier bonus
func (k Keeper) calculateVotingPower(ctx sdk.Context, proposal types.Proposal, voter sdk.AccAddress) (math.LegacyDec, error) {
	// Get HODL balance as base voting power
	hodlBalance := k.hodlKeeper.GetBalance(ctx, voter)
	votingPower := math.LegacyNewDecFromInt(hodlBalance)

	// For company-specific proposals, add equity voting power
	// Shareholders can vote on company matters with their shares
	if proposal.CompanyID > 0 {
		equityShares := k.equityKeeper.GetVotingSharesForCompany(ctx, proposal.CompanyID, voter.String())
		if !equityShares.IsZero() {
			// Add equity shares to voting power
			// Each share = 1 voting power unit (same weight as 1 uhodl)
			votingPower = votingPower.Add(math.LegacyNewDecFromInt(equityShares))
		}
	}

	// Add validator tier bonus if applicable
	tier := k.stakingKeeper.GetUserTierInt(ctx, voter)
	tierMultiplier := k.getTierMultiplier(tier)
	if !tierMultiplier.IsZero() {
		votingPower = votingPower.Mul(tierMultiplier)
	}

	return votingPower, nil
}

// getTierMultiplier returns multiplier for staking tier
func (k Keeper) getTierMultiplier(tier int) math.LegacyDec {
	switch tier {
	case 5: // Validator
		return math.LegacyNewDec(4)
	case 4: // Archon
		return math.LegacyNewDec(3)
	case 3: // Steward
		return math.LegacyNewDecWithPrec(25, 1) // 2.5x
	case 2: // Warden
		return math.LegacyNewDec(2)
	case 1: // Keeper
		return math.LegacyNewDecWithPrec(15, 1) // 1.5x
	default: // Holder
		return math.LegacyOneDec()
	}
}

// updateTallyResult updates the vote tally for a proposal
func (k Keeper) updateTallyResult(ctx sdk.Context, proposal types.Proposal, vote types.Vote, weight math.LegacyDec) {
	weightedPower := math.LegacyNewDecFromInt(vote.VotingPower).Mul(weight).TruncateInt()

	switch vote.Option {
	case types.VoteOptionYes:
		proposal.YesVotes = proposal.YesVotes.Add(weightedPower)
	case types.VoteOptionNo:
		proposal.NoVotes = proposal.NoVotes.Add(weightedPower)
	case types.VoteOptionAbstain:
		proposal.AbstainVotes = proposal.AbstainVotes.Add(weightedPower)
	case types.VoteOptionNoWithVeto:
		proposal.NoWithVetoVotes = proposal.NoWithVetoVotes.Add(weightedPower)
	}

	proposal.TotalVotingPower = proposal.TotalVotingPower.Add(weightedPower)
	proposal.UpdatedAt = ctx.BlockTime()

	k.setProposal(ctx, proposal)
}

// GetAuthority returns the module authority
func (k Keeper) GetAuthority() string {
	return k.authority
}

// SetProposalExternal is the exported method to set a proposal (for genesis init)
func (k Keeper) SetProposalExternal(ctx sdk.Context, proposal types.Proposal) {
	k.setProposal(ctx, proposal)
}

// GetAllProposals returns all proposals from the store
func (k Keeper) GetAllProposals(ctx sdk.Context) []types.Proposal {
	var proposals []types.Proposal

	k.IterateProposals(ctx, func(proposal types.Proposal) bool {
		proposals = append(proposals, proposal)
		return false
	})

	return proposals
}

// ExecuteSetCharityWalletProposal executes a proposal to set the default charity wallet
func (k Keeper) ExecuteSetCharityWalletProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Extract wallet address from proposal metadata
	walletStr, ok := proposal.Metadata["wallet"].(string)
	if !ok || walletStr == "" {
		return fmt.Errorf("invalid proposal metadata: wallet address not found or empty")
	}

	// Validate address format
	_, err := sdk.AccAddressFromBech32(walletStr)
	if err != nil {
		return fmt.Errorf("invalid charity wallet address: %w", err)
	}

	// Call equity keeper to set the wallet
	if err := k.equityKeeper.SetDefaultCharityWallet(ctx, walletStr); err != nil {
		return fmt.Errorf("failed to set default charity wallet: %w", err)
	}

	// Emit event for transparency
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"charity_wallet_updated",
			sdk.NewAttribute("wallet", walletStr),
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
		),
	)

	return nil
}
