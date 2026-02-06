package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// SubmitProposal handles proposal submission
func (ms msgServer) SubmitProposal(goCtx context.Context, msg *types.MsgSubmitProposal) (*types.MsgSubmitProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get initial deposit amount (sum of all coins in uhodl)
	var initialDeposit math.Int
	if msg.InitialDeposit.IsValid() && !msg.InitialDeposit.IsZero() {
		initialDeposit = msg.InitialDeposit.AmountOf("uhodl")
	} else {
		initialDeposit = math.ZeroInt()
	}

	// Submit the proposal
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		msg.Proposer,
		msg.Type,
		msg.Title,
		msg.Description,
		initialDeposit,
		uint32(ms.Keeper.GetVotingPeriodDays(ctx)),     // Governance-controllable voting period
		ms.Keeper.GetQuorum(ctx),                     // Governance-controllable quorum
		ms.Keeper.GetThreshold(ctx),                  // Governance-controllable threshold
	)

	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitProposalResponse{
		ProposalID: proposalID,
	}, nil
}

// Vote handles voting on proposals
func (ms msgServer) Vote(goCtx context.Context, msg *types.MsgVote) (*types.MsgVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.Vote(
		ctx,
		msg.Voter,
		msg.ProposalID,
		msg.Option,
		"", // No reason field in MsgVote
	)

	if err != nil {
		return nil, err
	}

	return &types.MsgVoteResponse{}, nil
}

// VoteWeighted handles weighted voting on proposals
func (ms msgServer) VoteWeighted(goCtx context.Context, msg *types.MsgVoteWeighted) (*types.MsgVoteWeightedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.VoteWeighted(
		ctx,
		msg.Voter,
		msg.ProposalID,
		msg.Options,
		"", // No reason field in MsgVoteWeighted
	)

	if err != nil {
		return nil, err
	}

	return &types.MsgVoteWeightedResponse{}, nil
}

// Deposit handles depositing tokens to proposals
func (ms msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get amount in uhodl
	amount := msg.Amount.AmountOf("uhodl")

	err := ms.Keeper.Deposit(
		ctx,
		msg.Depositor,
		msg.ProposalID,
		amount,
	)

	if err != nil {
		return nil, err
	}

	return &types.MsgDepositResponse{}, nil
}

// CancelProposal handles proposal cancellation
func (ms msgServer) CancelProposal(goCtx context.Context, msg *types.MsgCancelProposal) (*types.MsgCancelProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.CancelProposal(
		ctx,
		msg.Proposer,
		msg.ProposalID,
	)

	if err != nil {
		return nil, err
	}

	return &types.MsgCancelProposalResponse{}, nil
}

// SetGovernanceParams handles governance parameter updates (authority only)
func (ms msgServer) SetGovernanceParams(goCtx context.Context, msg *types.MsgSetGovernanceParams) (*types.MsgSetGovernanceParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority
	if msg.Authority != ms.Keeper.authority {
		return nil, types.ErrInvalidAuthority
	}

	// Store parameters
	ms.Keeper.SetGovernanceParams(ctx, msg.Params)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"set_governance_params",
			sdk.NewAttribute("authority", msg.Authority),
		),
	)

	return &types.MsgSetGovernanceParamsResponse{}, nil
}

// Helper function to create company-specific proposal
func (ms msgServer) SubmitCompanyProposal(
	ctx sdk.Context,
	submitter string,
	proposalType types.ProposalType,
	title string,
	description string,
	companyID uint64,
	initialDeposit math.Int,
) (uint64, error) {

	// Create company-specific proposal with governance-controllable params
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		submitter,
		proposalType,
		title,
		description,
		initialDeposit,
		uint32(ms.Keeper.GetVotingPeriodDays(ctx)),  // Governance-controllable voting period
		ms.Keeper.GetCompanyQuorum(ctx),           // Governance-controllable company quorum
		ms.Keeper.GetCompanyThreshold(ctx),        // Governance-controllable company threshold
	)

	if err != nil {
		return 0, err
	}

	// Store company-specific indexing
	store := runtime.KVStoreAdapter(ms.Keeper.storeService.OpenKVStore(ctx))
	store.Set(types.CompanyProposalKey(companyID, proposalID), []byte{})

	return proposalID, nil
}

// Helper function to create validator-specific proposal
func (ms msgServer) SubmitValidatorProposal(
	ctx sdk.Context,
	submitter string,
	proposalType types.ProposalType,
	title string,
	description string,
	validatorAddr string,
	initialDeposit math.Int,
) (uint64, error) {

	// Create validator-specific proposal with governance-controllable params
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		submitter,
		proposalType,
		title,
		description,
		initialDeposit,
		uint32(ms.Keeper.GetValidatorVotingPeriodDays(ctx)), // Governance-controllable voting period
		ms.Keeper.GetValidatorQuorum(ctx),                 // Governance-controllable validator quorum
		ms.Keeper.GetValidatorThreshold(ctx),              // Governance-controllable validator threshold
	)

	if err != nil {
		return 0, err
	}

	// Store validator-specific indexing
	store := runtime.KVStoreAdapter(ms.Keeper.storeService.OpenKVStore(ctx))
	validatorAddrBytes, _ := sdk.AccAddressFromBech32(validatorAddr)
	store.Set(types.ValidatorProposalKey(validatorAddrBytes, proposalID), []byte{})

	return proposalID, nil
}

// Helper function to create protocol parameter proposal
func (ms msgServer) SubmitProtocolParameterProposal(
	ctx sdk.Context,
	submitter string,
	title string,
	description string,
	parameterChanges string, // JSON string of parameter changes
	initialDeposit math.Int,
) (uint64, error) {

	// Create protocol parameter proposal with governance-controllable params
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		submitter,
		types.ProposalTypeParameterChange,
		title,
		description,
		initialDeposit,
		uint32(ms.Keeper.GetVotingPeriodDays(ctx) + 7), // Protocol proposals get extended voting (base + 7 days)
		ms.Keeper.GetProtocolQuorum(ctx),           // Governance-controllable protocol quorum
		ms.Keeper.GetThreshold(ctx),                // Governance-controllable threshold
	)

	if err != nil {
		return 0, err
	}

	return proposalID, nil
}

// Helper function to create treasury spend proposal
func (ms msgServer) SubmitTreasurySpendProposal(
	ctx sdk.Context,
	submitter string,
	title string,
	description string,
	recipient string,
	amount math.Int,
	initialDeposit math.Int,
) (uint64, error) {

	// Create treasury spend proposal with governance-controllable params
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		submitter,
		types.ProposalTypeTreasurySpend,
		title,
		description,
		initialDeposit,
		uint32(ms.Keeper.GetVotingPeriodDays(ctx)),   // Governance-controllable voting period
		ms.Keeper.GetQuorum(ctx),                   // Governance-controllable quorum
		ms.Keeper.GetThreshold(ctx),                // Governance-controllable threshold
	)

	if err != nil {
		return 0, err
	}

	return proposalID, nil
}

// Helper function to create emergency action proposal
func (ms msgServer) SubmitEmergencyActionProposal(
	ctx sdk.Context,
	submitter string,
	title string,
	description string,
	action string, // JSON string describing emergency action
	initialDeposit math.Int,
) (uint64, error) {

	// Create emergency action proposal with expedited params
	// Emergency proposals have higher thresholds but shorter voting period
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		submitter,
		types.ProposalTypeEmergencyAction,
		title,
		description,
		initialDeposit,
		3,                                          // Emergency proposals: 3 days (expedited, not governance-controllable for security)
		ms.Keeper.GetValidatorQuorum(ctx),          // Use validator quorum (higher)
		ms.Keeper.GetValidatorThreshold(ctx),       // Use validator threshold (higher)
	)

	if err != nil {
		return 0, err
	}

	return proposalID, nil
}

// ProcessEndBlockProposals processes proposals in the end block
func (ms msgServer) ProcessEndBlockProposals(ctx sdk.Context) {
	ms.Keeper.EndBlocker(ctx)
}

// CancelProposal cancels a proposal (keeper method)
func (k Keeper) CancelProposal(ctx sdk.Context, proposer string, proposalID uint64) error {
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Only the proposer can cancel
	if proposal.Proposer != proposer {
		return types.ErrInvalidProposer
	}

	// Can only cancel during deposit period
	if proposal.Status != types.ProposalStatusDepositPeriod {
		return types.ErrProposalNotActive
	}

	// Update status to canceled
	proposal.Status = types.ProposalStatusCanceled
	proposal.UpdatedAt = ctx.BlockTime()
	k.setProposal(ctx, proposal)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"cancel_proposal",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("proposer", proposer),
		),
	)

	return nil
}

// SetGovernanceParams stores governance parameters
func (k Keeper) SetGovernanceParams(ctx sdk.Context, params types.GovernanceParams) {
	store := k.storeService.OpenKVStore(ctx)
	bz, _ := json.Marshal(params)
	store.Set(types.ParamsPrefix, bz)
}

// EndBlocker processes proposals that need attention
func (k Keeper) EndBlocker(ctx sdk.Context) {
	// Process expired deposit period proposals
	k.processExpiredDeposits(ctx)

	// Process ended voting period proposals
	k.processEndedVoting(ctx)
}

// processExpiredDeposits handles proposals with expired deposit periods
// PERFORMANCE FIX: Uses time-based index instead of scanning all proposals
func (k Keeper) processExpiredDeposits(ctx sdk.Context) {
	currentTime := ctx.BlockTime()
	params := k.GetGovernanceParams(ctx)
	store := k.storeService.OpenKVStore(ctx)

	// PERFORMANCE FIX: Only iterate proposals that have actually expired
	// Using time-based index for O(log n) instead of O(n) full scan
	iterator, _ := store.Iterator(
		types.DepositEndTimePrefix,
		types.DepositEndTimeRangePrefix(currentTime.Unix()+1),
	)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Extract proposal ID from key
		key := iterator.Key()
		if len(key) < len(types.DepositEndTimePrefix)+8+8 {
			continue
		}
		proposalID := sdk.BigEndianToUint64(key[len(key)-8:])

		// Get the proposal
		proposal, found := k.GetProposal(ctx, proposalID)
		if !found {
			continue
		}

		// Skip if not in deposit period (status may have changed)
		if proposal.Status != types.ProposalStatusDepositPeriod {
			// Clean up stale index entry
			store.Delete(key)
			continue
		}

		// Deposit period ended - check if minimum deposit was met
		if proposal.HasMetDeposit() {
			// Transition to voting period
			proposal.Status = types.ProposalStatusVotingPeriod
			proposal.VotingStartTime = currentTime
			proposal.VotingEndTime = currentTime.Add(params.VotingPeriod)
			proposal.UpdatedAt = currentTime

			// Remove from deposit end time index
			store.Delete(key)

			// Update proposal (will add to voting end time index)
			k.setProposal(ctx, proposal)

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"proposal_voting_started",
					sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
					sdk.NewAttribute("voting_end_time", proposal.VotingEndTime.String()),
				),
			)
		} else {
			// Deposit not met - fail the proposal and refund deposits
			proposal.Status = types.ProposalStatusFailed
			proposal.UpdatedAt = currentTime
			proposal.FinalizedAt = currentTime

			// Remove from deposit end time index
			store.Delete(key)

			k.setProposal(ctx, proposal)

			// Refund deposits
			k.refundProposalDeposits(ctx, proposal.ID)

			// Emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"proposal_deposit_failed",
					sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
					sdk.NewAttribute("total_deposit", proposal.TotalDeposit.String()),
					sdk.NewAttribute("min_deposit", proposal.MinDeposit.String()),
				),
			)
		}
	}
}

// processEndedVoting handles proposals with ended voting periods
// PERFORMANCE FIX: Uses time-based index instead of scanning all proposals
func (k Keeper) processEndedVoting(ctx sdk.Context) {
	currentTime := ctx.BlockTime()
	store := k.storeService.OpenKVStore(ctx)

	// PERFORMANCE FIX: Only iterate proposals that have actually ended
	// Using time-based index for O(log n) instead of O(n) full scan
	iterator, _ := store.Iterator(
		types.VotingEndTimePrefix,
		types.VotingEndTimeRangePrefix(currentTime.Unix()+1),
	)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Extract proposal ID from key
		key := iterator.Key()
		if len(key) < len(types.VotingEndTimePrefix)+8+8 {
			continue
		}
		proposalID := sdk.BigEndianToUint64(key[len(key)-8:])

		// Get the proposal
		proposal, found := k.GetProposal(ctx, proposalID)
		if !found {
			continue
		}

		// Skip if not in voting period (status may have changed)
		if proposal.Status != types.ProposalStatusVotingPeriod {
			// Clean up stale index entry
			store.Delete(key)
			continue
		}

		// Remove from voting end time index
		store.Delete(key)

		// Process the proposal results
		if err := k.ProcessProposalResults(ctx, proposal.ID); err != nil {
			ctx.Logger().Error("Failed to process proposal results",
				"proposal_id", proposal.ID,
				"error", err)
		}
	}
}
