package keeper

import (
	"encoding/json"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
)

// Proposal storage operations

// GetProposal retrieves a proposal by ID
func (k Keeper) GetProposal(ctx sdk.Context, proposalID uint64) (types.Proposal, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.ProposalKey(proposalID))
	if bz == nil {
		return types.Proposal{}, false
	}

	var proposal types.Proposal
	err := json.Unmarshal(bz, &proposal)
	if err != nil {
		return types.Proposal{}, false
	}

	return proposal, true
}

// setProposal stores a proposal
func (k Keeper) setProposal(ctx sdk.Context, proposal types.Proposal) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := json.Marshal(proposal)
	if err != nil {
		panic(err)
	}
	store.Set(types.ProposalKey(proposal.ID), bz)
}

// getNextProposalID returns the next available proposal ID
func (k Keeper) getNextProposalID(ctx sdk.Context) uint64 {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.ProposalIDCounterPrefix)
	if bz == nil {
		// Initialize counter
		store.Set(types.ProposalIDCounterPrefix, sdk.Uint64ToBigEndian(1))
		return 1
	}

	currentID := sdk.BigEndianToUint64(bz)
	nextID := currentID + 1
	store.Set(types.ProposalIDCounterPrefix, sdk.Uint64ToBigEndian(nextID))
	return nextID
}

// IterateProposals iterates over all proposals
func (k Keeper) IterateProposals(ctx sdk.Context, cb func(proposal types.Proposal) bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	proposalStore := prefix.NewStore(store, types.ProposalPrefix)
	
	iterator := proposalStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var proposal types.Proposal
		err := json.Unmarshal(iterator.Value(), &proposal)
		if err != nil {
			continue
		}

		if cb(proposal) {
			break
		}
	}
}

// Vote storage operations

// GetVote retrieves a vote
func (k Keeper) GetVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (types.Vote, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.VoteKey(proposalID, voter))
	if bz == nil {
		return types.Vote{}, false
	}

	var vote types.Vote
	err := json.Unmarshal(bz, &vote)
	if err != nil {
		return types.Vote{}, false
	}

	return vote, true
}

// setVote stores a vote
func (k Keeper) setVote(ctx sdk.Context, vote types.Vote) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	voterAddr, _ := sdk.AccAddressFromBech32(vote.Voter)
	bz, err := json.Marshal(vote)
	if err != nil {
		panic(err)
	}
	store.Set(types.VoteKey(vote.ProposalID, voterAddr), bz)
}

// hasVoted checks if a voter has already voted on a proposal
func (k Keeper) hasVoted(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) bool {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return store.Has(types.VoteKey(proposalID, voter)) || store.Has(types.WeightedVoteKey(proposalID, voter))
}

// GetWeightedVote retrieves a weighted vote
func (k Keeper) GetWeightedVote(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (types.WeightedVote, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.WeightedVoteKey(proposalID, voter))
	if bz == nil {
		return types.WeightedVote{}, false
	}

	var vote types.WeightedVote
	err := json.Unmarshal(bz, &vote)
	if err != nil {
		return types.WeightedVote{}, false
	}

	return vote, true
}

// setWeightedVote stores a weighted vote
func (k Keeper) setWeightedVote(ctx sdk.Context, vote types.WeightedVote) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	voterAddr, _ := sdk.AccAddressFromBech32(vote.Voter)
	bz, err := json.Marshal(vote)
	if err != nil {
		panic(err)
	}
	store.Set(types.WeightedVoteKey(vote.ProposalID, voterAddr), bz)
}

// IterateVotes iterates over all votes for a proposal
func (k Keeper) IterateVotes(ctx sdk.Context, proposalID uint64, cb func(vote types.Vote) bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	voteStore := prefix.NewStore(store, types.ProposalVoteIteratorKey(proposalID))
	
	iterator := voteStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var vote types.Vote
		err := json.Unmarshal(iterator.Value(), &vote)
		if err != nil {
			continue
		}

		if cb(vote) {
			break
		}
	}
}

// IterateWeightedVotes iterates over all weighted votes for a proposal
func (k Keeper) IterateWeightedVotes(ctx sdk.Context, proposalID uint64, cb func(vote types.WeightedVote) bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	voteStore := prefix.NewStore(store, types.ProposalWeightedVoteIteratorKey(proposalID))
	
	iterator := voteStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var vote types.WeightedVote
		err := json.Unmarshal(iterator.Value(), &vote)
		if err != nil {
			continue
		}

		if cb(vote) {
			break
		}
	}
}

// Deposit storage operations

// GetDeposit retrieves a deposit
func (k Keeper) GetDeposit(ctx sdk.Context, proposalID uint64, depositor sdk.AccAddress) (types.Deposit, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.DepositKey(proposalID, depositor))
	if bz == nil {
		return types.Deposit{}, false
	}

	var deposit types.Deposit
	err := json.Unmarshal(bz, &deposit)
	if err != nil {
		return types.Deposit{}, false
	}

	return deposit, true
}

// setDeposit stores a deposit
func (k Keeper) setDeposit(ctx sdk.Context, deposit types.Deposit) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	depositorAddr, _ := sdk.AccAddressFromBech32(deposit.Depositor)
	bz, err := json.Marshal(deposit)
	if err != nil {
		panic(err)
	}
	store.Set(types.DepositKey(deposit.ProposalID, depositorAddr), bz)
}

// addDeposit adds tokens to a proposal deposit
func (k Keeper) addDeposit(ctx sdk.Context, proposalID uint64, depositor sdk.AccAddress, amount math.Int) error {
	// Transfer tokens from depositor to module
	hodlCoin := sdk.NewCoin("hodl", amount)
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleName, sdk.NewCoins(hodlCoin))
	if err != nil {
		return types.ErrInsufficientFunds
	}

	// Get or create deposit record
	deposit, found := k.GetDeposit(ctx, proposalID, depositor)
	if found {
		deposit.Amount = deposit.Amount.Add(amount)
		deposit.UpdatedAt = ctx.BlockTime()
	} else {
		deposit = types.Deposit{
			ProposalID: proposalID,
			Depositor:  depositor.String(),
			Amount:     amount,
			DepositedAt: ctx.BlockTime(),
			UpdatedAt:   ctx.BlockTime(),
		}
	}

	k.setDeposit(ctx, deposit)

	// Update proposal total deposit
	proposal, found := k.GetProposal(ctx, proposalID)
	if found {
		proposal.TotalDeposit = proposal.TotalDeposit.Add(amount)
		proposal.UpdatedAt = ctx.BlockTime()
		k.setProposal(ctx, proposal)
	}

	return nil
}

// IterateDeposits iterates over all deposits for a proposal
func (k Keeper) IterateDeposits(ctx sdk.Context, proposalID uint64, cb func(deposit types.Deposit) bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	depositStore := prefix.NewStore(store, types.ProposalDepositIteratorKey(proposalID))
	
	iterator := depositStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit types.Deposit
		err := json.Unmarshal(iterator.Value(), &deposit)
		if err != nil {
			continue
		}

		if cb(deposit) {
			break
		}
	}
}

// refundProposalDeposits refunds all deposits for a proposal
func (k Keeper) refundProposalDeposits(ctx sdk.Context, proposalID uint64) {
	k.IterateDeposits(ctx, proposalID, func(deposit types.Deposit) bool {
		depositorAddr, _ := sdk.AccAddressFromBech32(deposit.Depositor)
		hodlCoin := sdk.NewCoin("hodl", deposit.Amount)
		
		// Refund tokens
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositorAddr, sdk.NewCoins(hodlCoin))
		if err == nil {
			// Remove deposit record after successful refund
			store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
			store.Delete(types.DepositKey(deposit.ProposalID, depositorAddr))
		}
		
		return false // Continue iterating
	})
}

// burnProposalDeposits burns all deposits for a proposal
func (k Keeper) burnProposalDeposits(ctx sdk.Context, proposalID uint64) {
	totalBurn := math.ZeroInt()
	
	k.IterateDeposits(ctx, proposalID, func(deposit types.Deposit) bool {
		totalBurn = totalBurn.Add(deposit.Amount)
		
		// Remove deposit record
		store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
		depositorAddr, _ := sdk.AccAddressFromBech32(deposit.Depositor)
		store.Delete(types.DepositKey(deposit.ProposalID, depositorAddr))
		
		return false // Continue iterating
	})

	if !totalBurn.IsZero() {
		hodlCoin := sdk.NewCoin("hodl", totalBurn)
		k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(hodlCoin))
	}
}

// Tally storage operations

// GetTallyResult retrieves tally results for a proposal
func (k Keeper) GetTallyResult(ctx sdk.Context, proposalID uint64) (types.TallyResult, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.TallyResultKey(proposalID))
	if bz == nil {
		return types.TallyResult{}, false
	}

	var tally types.TallyResult
	err := json.Unmarshal(bz, &tally)
	if err != nil {
		return types.TallyResult{}, false
	}

	return tally, true
}

// setTallyResult stores tally results
func (k Keeper) setTallyResult(ctx sdk.Context, proposalID uint64, tally types.TallyResult) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := json.Marshal(tally)
	if err != nil {
		panic(err)
	}
	store.Set(types.TallyResultKey(proposalID), bz)
}

// updateTallyResult updates tally results with a new vote
func (k Keeper) updateTallyResult(ctx sdk.Context, proposal types.Proposal, vote types.Vote, weight math.LegacyDec) {
	tally, found := k.GetTallyResult(ctx, proposal.ID)
	if !found {
		tally = types.TallyResult{
			ProposalID:    proposal.ID,
			YesVotes:      math.LegacyZeroDec(),
			NoVotes:       math.LegacyZeroDec(),
			AbstainVotes:  math.LegacyZeroDec(),
			VetoVotes:     math.LegacyZeroDec(),
			TotalVotes:    math.LegacyZeroDec(),
		}
	}

	votePower := vote.VotingPower.Mul(weight)
	tally.TotalVotes = tally.TotalVotes.Add(votePower)

	switch vote.Option {
	case types.OptionYes:
		tally.YesVotes = tally.YesVotes.Add(votePower)
	case types.OptionNo:
		tally.NoVotes = tally.NoVotes.Add(votePower)
	case types.OptionAbstain:
		tally.AbstainVotes = tally.AbstainVotes.Add(votePower)
	case types.OptionVeto:
		tally.VetoVotes = tally.VetoVotes.Add(votePower)
	}

	k.setTallyResult(ctx, proposal.ID, tally)
}

// Governance parameters

// getGovernanceParams retrieves governance parameters
func (k Keeper) getGovernanceParams(ctx sdk.Context) types.GovernanceParams {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(types.ParamsPrefix)
	if bz == nil {
		// Return default parameters
		return types.GovernanceParams{
			MinDeposit:             math.NewInt(10000),           // 10,000 HODL
			MaxDepositPeriodDays:   7,                           // 7 days
			VotingPeriodDays:       14,                          // 14 days
			Quorum:                 math.LegacyMustNewDecFromStr("0.334"), // 33.4%
			Threshold:              math.LegacyMustNewDecFromStr("0.5"),   // 50%
			VetoThreshold:          math.LegacyMustNewDecFromStr("0.334"), // 33.4%
			BurnDeposits:           true,
			BurnVoteVeto:           true,
			MinInitialDepositRatio: math.LegacyMustNewDecFromStr("0.25"),  // 25%
		}
	}

	var params types.GovernanceParams
	err := json.Unmarshal(bz, &params)
	if err != nil {
		panic(err)
	}

	return params
}

// SetGovernanceParams stores governance parameters
func (k Keeper) SetGovernanceParams(ctx sdk.Context, params types.GovernanceParams) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	store.Set(types.ParamsPrefix, bz)
}

// Indexing operations

// indexProposal creates indexes for efficient querying
func (k Keeper) indexProposal(ctx sdk.Context, proposal types.Proposal) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	
	// Index by status
	store.Set(types.ProposalByStatusKey(string(proposal.Status), proposal.ID), []byte{})
	
	// Index by submitter
	submitterAddr, _ := sdk.AccAddressFromBech32(proposal.Submitter)
	store.Set(types.ProposalBySubmitterKey(submitterAddr, proposal.ID), []byte{})
	
	// Index by type
	store.Set(types.ProposalByTypeKey(string(proposal.Type), proposal.ID), []byte{})
}

// updateProposalIndex updates proposal indexes when status changes
func (k Keeper) updateProposalIndex(ctx sdk.Context, proposal types.Proposal) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	
	// Remove old status index (we need to iterate to find it)
	statusStore := prefix.NewStore(store, types.ProposalByStatusPrefix)
	statusIter := statusStore.Iterator(nil, nil)
	defer statusIter.Close()
	
	for ; statusIter.Valid(); statusIter.Next() {
		proposalID := types.GetProposalIDFromKey(statusIter.Key(), len(types.ProposalByStatusPrefix))
		if proposalID == proposal.ID {
			statusStore.Delete(statusIter.Key())
			break
		}
	}
	
	// Add new status index
	store.Set(types.ProposalByStatusKey(string(proposal.Status), proposal.ID), []byte{})
}