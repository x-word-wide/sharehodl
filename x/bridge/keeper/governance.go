package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
)

// ============================================================================
// Withdrawal Proposal Management
// ============================================================================

// GetNextProposalID returns the next withdrawal proposal ID
func (k Keeper) GetNextProposalID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.NextProposalIDKey)
	if bz == nil {
		return 1
	}
	return sdk.BigEndianToUint64(bz)
}

// SetNextProposalID sets the next withdrawal proposal ID
func (k Keeper) SetNextProposalID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.NextProposalIDKey, sdk.Uint64ToBigEndian(id))
}

// GetWithdrawalProposal retrieves a withdrawal proposal
func (k Keeper) GetWithdrawalProposal(ctx sdk.Context, proposalID uint64) (types.ReserveWithdrawal, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.WithdrawalProposalKey(proposalID))
	if bz == nil {
		return types.ReserveWithdrawal{}, false
	}

	var proposal types.ReserveWithdrawal
	if err := json.Unmarshal(bz, &proposal); err != nil {
		return types.ReserveWithdrawal{}, false
	}
	return proposal, true
}

// SetWithdrawalProposal stores a withdrawal proposal
func (k Keeper) SetWithdrawalProposal(ctx sdk.Context, proposal types.ReserveWithdrawal) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(proposal)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal withdrawal proposal", "error", err)
		return
	}
	store.Set(types.WithdrawalProposalKey(proposal.ID), bz)
}

// GetAllWithdrawalProposals returns all withdrawal proposals
func (k Keeper) GetAllWithdrawalProposals(ctx sdk.Context) []types.ReserveWithdrawal {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.WithdrawalProposalPrefix)
	defer iterator.Close()

	proposals := []types.ReserveWithdrawal{}
	for ; iterator.Valid(); iterator.Next() {
		var proposal types.ReserveWithdrawal
		if err := json.Unmarshal(iterator.Value(), &proposal); err != nil {
			continue
		}
		proposals = append(proposals, proposal)
	}
	return proposals
}

// ProposeWithdrawal creates a new withdrawal proposal
func (k Keeper) ProposeWithdrawal(
	ctx sdk.Context,
	proposer sdk.AccAddress,
	recipient sdk.AccAddress,
	denom string,
	amount math.Int,
	reason string,
) (uint64, error) {
	// Verify proposer is a validator
	if !k.IsValidator(ctx, proposer) {
		return 0, types.ErrNotValidator
	}

	// Validate asset is supported
	_, found := k.GetSupportedAsset(ctx, denom)
	if !found {
		return 0, types.ErrAssetNotSupported
	}

	params := k.GetParams(ctx)

	// Check withdrawal amount doesn't exceed maximum allowed
	reserveBalance := k.GetReserveBalance(ctx, denom)
	maxWithdrawal := params.MaxWithdrawalPercent.MulInt(reserveBalance).TruncateInt()

	if amount.GT(maxWithdrawal) {
		return 0, types.ErrExceedsMaxWithdrawal
	}

	// Check reserve balance is sufficient
	if reserveBalance.LT(amount) {
		return 0, types.ErrInsufficientReserve
	}

	// Check reserve ratio would be maintained after withdrawal
	if err := k.CheckReserveRatioAfterWithdrawal(ctx, denom, amount); err != nil {
		return 0, err
	}

	// Get total validator count
	totalValidators := k.GetValidatorCount(ctx)
	if totalValidators == 0 {
		return 0, fmt.Errorf("no active validators")
	}

	// Create proposal
	proposalID := k.GetNextProposalID(ctx)
	proposal := types.NewReserveWithdrawal(
		proposalID,
		proposer.String(),
		recipient.String(),
		denom,
		amount,
		reason,
		totalValidators,
		params.WithdrawalVotingPeriod,
	)

	// Validate proposal
	if err := proposal.Validate(); err != nil {
		return 0, err
	}

	// Store proposal
	k.SetWithdrawalProposal(ctx, proposal)
	k.SetNextProposalID(ctx, proposalID+1)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_withdrawal_proposed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("proposer", proposer.String()),
			sdk.NewAttribute("recipient", recipient.String()),
			sdk.NewAttribute("denom", denom),
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("reason", reason),
		),
	)

	k.Logger(ctx).Info("withdrawal proposal created",
		"proposal_id", proposalID,
		"proposer", proposer.String(),
		"amount", amount.String()+denom,
	)

	return proposalID, nil
}

// VoteOnWithdrawal allows validators to vote on a withdrawal proposal
func (k Keeper) VoteOnWithdrawal(ctx sdk.Context, validator sdk.AccAddress, proposalID uint64, vote bool) error {
	// Verify voter is a validator
	if !k.IsValidator(ctx, validator) {
		return types.ErrNotValidator
	}

	// Get proposal
	proposal, found := k.GetWithdrawalProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Check proposal is active
	if !proposal.IsActive() {
		if proposal.IsExpired() {
			return types.ErrProposalExpired
		}
		return types.ErrProposalNotActive
	}

	// Check if validator has already voted
	if k.HasVotedOnWithdrawal(ctx, proposalID, validator.String()) {
		return types.ErrProposalAlreadyVoted
	}

	// Record vote
	withdrawalVote := types.NewWithdrawalVote(proposalID, validator.String(), vote)
	k.SetWithdrawalVote(ctx, withdrawalVote)

	// Update proposal vote counts
	if vote {
		proposal.YesVotes++
	} else {
		proposal.NoVotes++
	}

	// Check if proposal should be approved or rejected
	params := k.GetParams(ctx)
	approvalRate := proposal.GetApprovalRate()

	if approvalRate.GTE(params.RequiredValidatorApproval) {
		proposal.Status = types.ProposalStatusApproved
	} else if proposal.NoVotes > proposal.TotalValidators/2 {
		// Rejected if majority votes no
		proposal.Status = types.ProposalStatusRejected
	}

	k.SetWithdrawalProposal(ctx, proposal)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_withdrawal_voted",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("validator", validator.String()),
			sdk.NewAttribute("vote", fmt.Sprintf("%t", vote)),
			sdk.NewAttribute("yes_votes", fmt.Sprintf("%d", proposal.YesVotes)),
			sdk.NewAttribute("no_votes", fmt.Sprintf("%d", proposal.NoVotes)),
			sdk.NewAttribute("status", proposal.Status.String()),
		),
	)

	k.Logger(ctx).Info("withdrawal vote recorded",
		"proposal_id", proposalID,
		"validator", validator.String(),
		"vote", vote,
	)

	return nil
}

// ExecuteWithdrawal executes an approved withdrawal proposal
func (k Keeper) ExecuteWithdrawal(ctx sdk.Context, proposalID uint64) error {
	// Get proposal
	proposal, found := k.GetWithdrawalProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Check proposal is approved
	if proposal.Status != types.ProposalStatusApproved {
		return types.ErrProposalNotApproved
	}

	// Check reserve still has sufficient balance
	reserveBalance := k.GetReserveBalance(ctx, proposal.Denom)
	if reserveBalance.LT(proposal.Amount) {
		return types.ErrInsufficientReserve
	}

	// Check reserve ratio would still be maintained
	if err := k.CheckReserveRatioAfterWithdrawal(ctx, proposal.Denom, proposal.Amount); err != nil {
		return err
	}

	// Subtract from reserve
	if err := k.SubtractFromReserve(ctx, proposal.Denom, proposal.Amount); err != nil {
		return err
	}

	// Send to recipient
	recipient, err := sdk.AccAddressFromBech32(proposal.Recipient)
	if err != nil {
		return err
	}

	coin := sdk.NewCoin(proposal.Denom, proposal.Amount)
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(coin)); err != nil {
		return fmt.Errorf("failed to send withdrawal to recipient: %w", err)
	}

	// Update proposal status
	now := ctx.BlockTime()
	proposal.Status = types.ProposalStatusExecuted
	proposal.ExecutedAt = &now
	k.SetWithdrawalProposal(ctx, proposal)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"bridge_withdrawal_executed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("recipient", proposal.Recipient),
			sdk.NewAttribute("denom", proposal.Denom),
			sdk.NewAttribute("amount", proposal.Amount.String()),
		),
	)

	k.Logger(ctx).Info("withdrawal executed",
		"proposal_id", proposalID,
		"recipient", proposal.Recipient,
		"amount", proposal.Amount.String()+proposal.Denom,
	)

	return nil
}

// GetWithdrawalVote retrieves a validator's vote on a proposal
func (k Keeper) GetWithdrawalVote(ctx sdk.Context, proposalID uint64, validator string) (types.WithdrawalVote, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.WithdrawalVoteKey(proposalID, validator))
	if bz == nil {
		return types.WithdrawalVote{}, false
	}

	var vote types.WithdrawalVote
	if err := json.Unmarshal(bz, &vote); err != nil {
		return types.WithdrawalVote{}, false
	}
	return vote, true
}

// SetWithdrawalVote stores a validator's vote
func (k Keeper) SetWithdrawalVote(ctx sdk.Context, vote types.WithdrawalVote) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(vote)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal withdrawal vote", "error", err)
		return
	}
	store.Set(types.WithdrawalVoteKey(vote.ProposalID, vote.Validator), bz)
}

// HasVotedOnWithdrawal checks if a validator has already voted on a proposal
func (k Keeper) HasVotedOnWithdrawal(ctx sdk.Context, proposalID uint64, validator string) bool {
	_, found := k.GetWithdrawalVote(ctx, proposalID, validator)
	return found
}

// ProcessExpiredProposals processes proposals that have expired
func (k Keeper) ProcessExpiredProposals(ctx sdk.Context) {
	proposals := k.GetAllWithdrawalProposals(ctx)

	for _, proposal := range proposals {
		if proposal.Status == types.ProposalStatusActive && proposal.IsExpired() {
			proposal.Status = types.ProposalStatusExpired
			k.SetWithdrawalProposal(ctx, proposal)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"bridge_withdrawal_expired",
					sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
				),
			)
		}
	}
}

// ExecuteApprovedProposals executes all approved proposals
func (k Keeper) ExecuteApprovedProposals(ctx sdk.Context) {
	proposals := k.GetAllWithdrawalProposals(ctx)

	for _, proposal := range proposals {
		if proposal.Status == types.ProposalStatusApproved {
			if err := k.ExecuteWithdrawal(ctx, proposal.ID); err != nil {
				k.Logger(ctx).Error("failed to execute approved withdrawal",
					"proposal_id", proposal.ID,
					"error", err,
				)
			}
		}
	}
}
