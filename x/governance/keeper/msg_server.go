package keeper

import (
	"context"

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

	// Submit the proposal
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		msg.Submitter,
		msg.ProposalType,
		msg.Title,
		msg.Description,
		msg.InitialDeposit,
		msg.VotingPeriodDays,
		msg.QuorumRequired,
		msg.ThresholdRequired,
	)

	if err != nil {
		return &types.MsgSubmitProposalResponse{
			ProposalID: 0,
			Success:    false,
		}, err
	}

	return &types.MsgSubmitProposalResponse{
		ProposalID: proposalID,
		Success:    true,
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
		msg.Reason,
	)

	if err != nil {
		return &types.MsgVoteResponse{Success: false}, err
	}

	return &types.MsgVoteResponse{Success: true}, nil
}

// VoteWeighted handles weighted voting on proposals
func (ms msgServer) VoteWeighted(goCtx context.Context, msg *types.MsgVoteWeighted) (*types.MsgVoteWeightedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert message options to keeper options
	options := make([]types.WeightedVoteOption, len(msg.Options))
	for i, option := range msg.Options {
		options[i] = types.WeightedVoteOption{
			Option: option.Option,
			Weight: option.Weight,
		}
	}

	err := ms.Keeper.VoteWeighted(
		ctx,
		msg.Voter,
		msg.ProposalID,
		options,
		msg.Reason,
	)

	if err != nil {
		return &types.MsgVoteWeightedResponse{Success: false}, err
	}

	return &types.MsgVoteWeightedResponse{Success: true}, nil
}

// Deposit handles depositing tokens to proposals
func (ms msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.Deposit(
		ctx,
		msg.Depositor,
		msg.ProposalID,
		msg.Amount,
	)

	if err != nil {
		return &types.MsgDepositResponse{Success: false}, err
	}

	return &types.MsgDepositResponse{Success: true}, nil
}

// CancelProposal handles proposal cancellation
func (ms msgServer) CancelProposal(goCtx context.Context, msg *types.MsgCancelProposal) (*types.MsgCancelProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.CancelProposal(
		ctx,
		msg.Proposer,
		msg.ProposalID,
		msg.Reason,
	)

	if err != nil {
		return &types.MsgCancelProposalResponse{Success: false}, err
	}

	return &types.MsgCancelProposalResponse{Success: true}, nil
}

// SetGovernanceParams handles governance parameter updates (authority only)
func (ms msgServer) SetGovernanceParams(goCtx context.Context, msg *types.MsgSetGovernanceParams) (*types.MsgSetGovernanceParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify authority
	if msg.Authority != ms.Keeper.authority {
		return &types.MsgSetGovernanceParamsResponse{Success: false}, types.ErrInvalidAuthority
	}

	// Create governance parameters
	params := types.GovernanceParams{
		MinDeposit:             msg.MinDeposit,
		MaxDepositPeriodDays:   msg.MaxDepositPeriod,
		VotingPeriodDays:       msg.VotingPeriod,
		Quorum:                 msg.Quorum,
		Threshold:              msg.Threshold,
		VetoThreshold:          msg.VetoThreshold,
		BurnDeposits:           msg.BurnDeposits,
		BurnVoteVeto:           msg.BurnVoteVeto,
		MinInitialDepositRatio: msg.MinInitialDeposit,
	}

	// Store parameters
	ms.Keeper.SetGovernanceParams(ctx, params)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"set_governance_params",
			sdk.NewAttribute("authority", msg.Authority),
			sdk.NewAttribute("min_deposit", msg.MinDeposit.String()),
			sdk.NewAttribute("voting_period", sdk.Uint64ToBigEndian(uint64(msg.VotingPeriod)).String()),
			sdk.NewAttribute("quorum", msg.Quorum.String()),
			sdk.NewAttribute("threshold", msg.Threshold.String()),
		),
	)

	return &types.MsgSetGovernanceParamsResponse{Success: true}, nil
}

// Helper function to create company-specific proposal
func (ms msgServer) SubmitCompanyProposal(
	ctx sdk.Context,
	submitter string,
	proposalType string,
	title string,
	description string,
	companyID uint64,
	initialDeposit math.Int,
) (*types.MsgSubmitProposalResponse, error) {

	// Create company-specific proposal
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		submitter,
		proposalType,
		title,
		description,
		initialDeposit,
		14, // Default 14 days voting period for company proposals
		math.LegacyMustNewDecFromStr("0.25"), // 25% quorum for company proposals
		math.LegacyMustNewDecFromStr("0.5"),  // 50% threshold for company proposals
	)

	if err != nil {
		return &types.MsgSubmitProposalResponse{
			ProposalID: 0,
			Success:    false,
		}, err
	}

	// Store company-specific indexing
	store := runtime.KVStoreAdapter(ms.Keeper.storeService.OpenKVStore(ctx))
	store.Set(types.CompanyProposalKey(companyID, proposalID), []byte{})

	return &types.MsgSubmitProposalResponse{
		ProposalID: proposalID,
		Success:    true,
	}, nil
}

// Helper function to create validator-specific proposal
func (ms msgServer) SubmitValidatorProposal(
	ctx sdk.Context,
	submitter string,
	proposalType string,
	title string,
	description string,
	validatorAddr string,
	initialDeposit math.Int,
) (*types.MsgSubmitProposalResponse, error) {

	// Create validator-specific proposal
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		submitter,
		proposalType,
		title,
		description,
		initialDeposit,
		7, // Default 7 days voting period for validator proposals
		math.LegacyMustNewDecFromStr("0.5"), // 50% quorum for validator proposals
		math.LegacyMustNewDecFromStr("0.67"), // 67% threshold for validator proposals
	)

	if err != nil {
		return &types.MsgSubmitProposalResponse{
			ProposalID: 0,
			Success:    false,
		}, err
	}

	// Store validator-specific indexing
	store := runtime.KVStoreAdapter(ms.Keeper.storeService.OpenKVStore(ctx))
	validatorAddrBytes, _ := sdk.AccAddressFromBech32(validatorAddr)
	store.Set(types.ValidatorProposalKey(validatorAddrBytes, proposalID), []byte{})

	return &types.MsgSubmitProposalResponse{
		ProposalID: proposalID,
		Success:    true,
	}, nil
}

// Helper function to create protocol parameter proposal
func (ms msgServer) SubmitProtocolParameterProposal(
	ctx sdk.Context,
	submitter string,
	title string,
	description string,
	parameterChanges string, // JSON string of parameter changes
	initialDeposit math.Int,
) (*types.MsgSubmitProposalResponse, error) {

	// Create protocol parameter proposal
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		submitter,
		"protocol_parameter",
		title,
		description,
		initialDeposit,
		21, // Default 21 days voting period for protocol proposals
		math.LegacyMustNewDecFromStr("0.4"), // 40% quorum for protocol proposals
		math.LegacyMustNewDecFromStr("0.5"), // 50% threshold for protocol proposals
	)

	if err != nil {
		return &types.MsgSubmitProposalResponse{
			ProposalID: 0,
			Success:    false,
		}, err
	}

	return &types.MsgSubmitProposalResponse{
		ProposalID: proposalID,
		Success:    true,
	}, nil
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
) (*types.MsgSubmitProposalResponse, error) {

	// Create treasury spend proposal
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		submitter,
		"treasury_spend",
		title,
		description,
		initialDeposit,
		14, // Default 14 days voting period for treasury proposals
		math.LegacyMustNewDecFromStr("0.33"), // 33% quorum for treasury proposals
		math.LegacyMustNewDecFromStr("0.5"),  // 50% threshold for treasury proposals
	)

	if err != nil {
		return &types.MsgSubmitProposalResponse{
			ProposalID: 0,
			Success:    false,
		}, err
	}

	return &types.MsgSubmitProposalResponse{
		ProposalID: proposalID,
		Success:    true,
	}, nil
}

// Helper function to create emergency action proposal
func (ms msgServer) SubmitEmergencyActionProposal(
	ctx sdk.Context,
	submitter string,
	title string,
	description string,
	action string, // JSON string describing emergency action
	initialDeposit math.Int,
) (*types.MsgSubmitProposalResponse, error) {

	// Create emergency action proposal
	proposalID, err := ms.Keeper.SubmitProposal(
		ctx,
		submitter,
		"emergency_action",
		title,
		description,
		initialDeposit,
		3, // Default 3 days voting period for emergency proposals
		math.LegacyMustNewDecFromStr("0.67"), // 67% quorum for emergency proposals
		math.LegacyMustNewDecFromStr("0.75"), // 75% threshold for emergency proposals
	)

	if err != nil {
		return &types.MsgSubmitProposalResponse{
			ProposalID: 0,
			Success:    false,
		}, err
	}

	return &types.MsgSubmitProposalResponse{
		ProposalID: proposalID,
		Success:    true,
	}, nil
}

// ProcessEndBlockProposals processes proposals in the end block
func (ms msgServer) ProcessEndBlockProposals(ctx sdk.Context) {
	ms.Keeper.EndBlocker(ctx)
}