package keeper

import (
	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
	"time"
)

// DelegateVotingPower delegates voting power to another address
func (k Keeper) DelegateVotingPower(
	ctx sdk.Context,
	delegator string,
	delegate string,
	companyID uint64,
	proposalType types.ProposalType,
	votingPower math.LegacyDec,
	expiryHeight uint64,
	revocable bool,
) error {
	delegatorAddr, err := sdk.AccAddressFromBech32(delegator)
	if err != nil {
		return types.ErrInvalidAddress
	}

	delegateAddr, err := sdk.AccAddressFromBech32(delegate)
	if err != nil {
		return types.ErrInvalidAddress
	}

	// Validate delegation parameters
	if votingPower.IsZero() || votingPower.IsNegative() {
		return types.ErrInvalidVotingPower
	}

	// Check if delegator has sufficient voting power
	availablePower, err := k.calculateAvailableVotingPower(ctx, delegatorAddr, companyID, proposalType)
	if err != nil {
		return err
	}

	if votingPower.GT(availablePower) {
		return types.ErrInsufficientVotingPower
	}

	// Create delegation
	delegation := types.VoteDelegation{
		Delegator:    delegator,
		Delegate:     delegate,
		CompanyID:    companyID,
		ProposalType: proposalType,
		VotingPower:  votingPower,
		ExpiryHeight: expiryHeight,
		Revocable:    revocable,
		CreatedAt:    ctx.BlockTime(),
	}

	// Store delegation
	k.setVoteDelegation(ctx, delegation)

	// Update delegation indexes
	k.indexDelegation(ctx, delegation)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"delegate_voting_power",
			sdk.NewAttribute("delegator", delegator),
			sdk.NewAttribute("delegate", delegate),
			sdk.NewAttribute("company_id", sdk.Uint64ToBigEndian(companyID).String()),
			sdk.NewAttribute("proposal_type", proposalType.String()),
			sdk.NewAttribute("voting_power", votingPower.String()),
			sdk.NewAttribute("expiry_height", sdk.Uint64ToBigEndian(expiryHeight).String()),
		),
	)

	return nil
}

// RevokeDelegation revokes a voting power delegation
func (k Keeper) RevokeDelegation(
	ctx sdk.Context,
	delegator string,
	delegate string,
	companyID uint64,
	proposalType types.ProposalType,
) error {
	delegatorAddr, err := sdk.AccAddressFromBech32(delegator)
	if err != nil {
		return types.ErrInvalidAddress
	}

	// Get delegation
	delegation, found := k.GetDelegation(ctx, delegatorAddr, companyID, proposalType)
	if !found {
		return types.ErrDelegationNotFound
	}

	// Validate revocation
	if delegation.Delegate != delegate {
		return types.ErrInvalidDelegate
	}

	if !delegation.Revocable {
		return types.ErrDelegationNotRevocable
	}

	// Remove delegation
	k.removeDelegation(ctx, delegation)
	k.removeFromDelegationIndexes(ctx, delegation)

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"revoke_delegation",
			sdk.NewAttribute("delegator", delegator),
			sdk.NewAttribute("delegate", delegate),
			sdk.NewAttribute("company_id", sdk.Uint64ToBigEndian(companyID).String()),
			sdk.NewAttribute("proposal_type", proposalType.String()),
		),
	)

	return nil
}

// VoteAsDelegate votes on behalf of delegators
func (k Keeper) VoteAsDelegate(
	ctx sdk.Context,
	delegate string,
	proposalID uint64,
	option types.VoteOption,
	reason string,
) error {
	delegateAddr, err := sdk.AccAddressFromBech32(delegate)
	if err != nil {
		return types.ErrInvalidAddress
	}

	// Get proposal
	proposal, found := k.GetProposal(ctx, proposalID)
	if !found {
		return types.ErrProposalNotFound
	}

	// Validate voting conditions
	if err := k.validateVoting(ctx, proposal, delegateAddr); err != nil {
		return err
	}

	// Get all delegations for this delegate for the proposal type
	delegations := k.GetDelegationsForDelegate(ctx, delegateAddr, proposal.Type)

	totalDelegatedPower := math.LegacyZeroDec()
	
	for _, delegation := range delegations {
		// Check if delegation is valid
		if k.isDelegationValid(ctx, delegation) {
			// Check if delegator hasn't already voted directly
			delegatorAddr, _ := sdk.AccAddressFromBech32(delegation.Delegator)
			if !k.hasVoted(ctx, proposalID, delegatorAddr) {
				totalDelegatedPower = totalDelegatedPower.Add(delegation.VotingPower)
				
				// Mark delegation as used
				delegation.LastUsed = ctx.BlockTime()
				k.setVoteDelegation(ctx, delegation)
			}
		}
	}

	if totalDelegatedPower.IsZero() {
		return types.ErrInsufficientVotingPower
	}

	// Calculate delegate's own voting power
	ownVotingPower, err := k.calculateVotingPower(ctx, proposal, delegateAddr)
	if err != nil {
		return err
	}

	// Create vote with combined power
	totalVotingPower := ownVotingPower.Add(totalDelegatedPower)

	vote := types.Vote{
		ProposalID:    proposalID,
		Voter:         delegate,
		Option:        option,
		Weight:        math.LegacyOneDec(),
		VotingPower:   totalVotingPower,
		DelegatedPower: totalDelegatedPower,
		Reason:        reason,
		SubmittedAt:   ctx.BlockTime(),
	}

	// Store vote
	k.setVote(ctx, vote)

	// Update tally
	k.updateTallyResult(ctx, proposal, vote, math.LegacyOneDec())

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote_as_delegate",
			sdk.NewAttribute("proposal_id", sdk.Uint64ToBigEndian(proposalID).String()),
			sdk.NewAttribute("delegate", delegate),
			sdk.NewAttribute("option", option.String()),
			sdk.NewAttribute("own_power", ownVotingPower.String()),
			sdk.NewAttribute("delegated_power", totalDelegatedPower.String()),
			sdk.NewAttribute("total_power", totalVotingPower.String()),
		),
	)

	return nil
}

// calculateAvailableVotingPower calculates available voting power after delegations
func (k Keeper) calculateAvailableVotingPower(
	ctx sdk.Context,
	addr sdk.AccAddress,
	companyID uint64,
	proposalType types.ProposalType,
) (math.LegacyDec, error) {
	// Calculate total voting power
	totalPower, err := k.calculateVotingPowerForType(ctx, addr, proposalType)
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	// Subtract already delegated power
	delegatedPower := k.getTotalDelegatedPower(ctx, addr, companyID, proposalType)
	
	availablePower := totalPower.Sub(delegatedPower)
	if availablePower.IsNegative() {
		return math.LegacyZeroDec(), nil
	}

	return availablePower, nil
}

// isDelegationValid checks if a delegation is still valid
func (k Keeper) isDelegationValid(ctx sdk.Context, delegation types.VoteDelegation) bool {
	// Check expiry height
	if delegation.ExpiryHeight > 0 && uint64(ctx.BlockHeight()) > delegation.ExpiryHeight {
		return false
	}

	// Check if delegator still has the voting power
	delegatorAddr, err := sdk.AccAddressFromBech32(delegation.Delegator)
	if err != nil {
		return false
	}

	currentPower, err := k.calculateVotingPowerForType(ctx, delegatorAddr, delegation.ProposalType)
	if err != nil {
		return false
	}

	return currentPower.GTE(delegation.VotingPower)
}

// Storage and indexing functions

// setVoteDelegation stores a vote delegation
func (k Keeper) setVoteDelegation(ctx sdk.Context, delegation types.VoteDelegation) {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	key := types.DelegationKey(delegation.Delegator, delegation.CompanyID, delegation.ProposalType)
	value := k.cdc.MustMarshal(&delegation)
	store.Set(key, value)
}

// GetDelegation retrieves a delegation
func (k Keeper) GetDelegation(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	companyID uint64,
	proposalType types.ProposalType,
) (types.VoteDelegation, bool) {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	key := types.DelegationKey(delegator.String(), companyID, proposalType)
	
	value := store.Get(key)
	if value == nil {
		return types.VoteDelegation{}, false
	}

	var delegation types.VoteDelegation
	k.cdc.MustUnmarshal(value, &delegation)
	return delegation, true
}

// GetDelegationsForDelegate gets all delegations for a delegate
func (k Keeper) GetDelegationsForDelegate(
	ctx sdk.Context,
	delegate sdk.AccAddress,
	proposalType types.ProposalType,
) []types.VoteDelegation {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	delegateStore := prefix.NewStore(store, types.DelegateIndexKey(delegate.String(), proposalType))
	
	var delegations []types.VoteDelegation
	
	iterator := delegateStore.Iterator(nil, nil)
	defer iterator.Close()
	
	for ; iterator.Valid(); iterator.Next() {
		var delegation types.VoteDelegation
		k.cdc.MustUnmarshal(iterator.Value(), &delegation)
		delegations = append(delegations, delegation)
	}
	
	return delegations
}

// removeDelegation removes a delegation from storage
func (k Keeper) removeDelegation(ctx sdk.Context, delegation types.VoteDelegation) {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	key := types.DelegationKey(delegation.Delegator, delegation.CompanyID, delegation.ProposalType)
	store.Delete(key)
}

// indexDelegation creates indexes for efficient delegation queries
func (k Keeper) indexDelegation(ctx sdk.Context, delegation types.VoteDelegation) {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	
	// Index by delegate
	delegateKey := types.DelegateIndexKey(delegation.Delegate, delegation.ProposalType)
	delegateIndexKey := append(delegateKey, []byte(delegation.Delegator)...)
	value := k.cdc.MustMarshal(&delegation)
	store.Set(delegateIndexKey, value)
	
	// Index by company
	if delegation.CompanyID > 0 {
		companyKey := types.CompanyDelegationIndexKey(delegation.CompanyID, delegation.ProposalType)
		companyIndexKey := append(companyKey, []byte(delegation.Delegator)...)
		store.Set(companyIndexKey, value)
	}
}

// removeFromDelegationIndexes removes delegation from all indexes
func (k Keeper) removeFromDelegationIndexes(ctx sdk.Context, delegation types.VoteDelegation) {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	
	// Remove from delegate index
	delegateKey := types.DelegateIndexKey(delegation.Delegate, delegation.ProposalType)
	delegateIndexKey := append(delegateKey, []byte(delegation.Delegator)...)
	store.Delete(delegateIndexKey)
	
	// Remove from company index
	if delegation.CompanyID > 0 {
		companyKey := types.CompanyDelegationIndexKey(delegation.CompanyID, delegation.ProposalType)
		companyIndexKey := append(companyKey, []byte(delegation.Delegator)...)
		store.Delete(companyIndexKey)
	}
}

// getTotalDelegatedPower gets total power already delegated by an address
func (k Keeper) getTotalDelegatedPower(
	ctx sdk.Context,
	addr sdk.AccAddress,
	companyID uint64,
	proposalType types.ProposalType,
) math.LegacyDec {
	store := ctx.KVStore(k.storeService.OpenKVStore(ctx))
	delegatorStore := prefix.NewStore(store, types.DelegatorIndexKey(addr.String()))
	
	totalDelegated := math.LegacyZeroDec()
	
	iterator := delegatorStore.Iterator(nil, nil)
	defer iterator.Close()
	
	for ; iterator.Valid(); iterator.Next() {
		var delegation types.VoteDelegation
		k.cdc.MustUnmarshal(iterator.Value(), &delegation)
		
		// Filter by company and proposal type if specified
		if (companyID == 0 || delegation.CompanyID == companyID) &&
		   (proposalType == 0 || delegation.ProposalType == proposalType) &&
		   k.isDelegationValid(ctx, delegation) {
			totalDelegated = totalDelegated.Add(delegation.VotingPower)
		}
	}
	
	return totalDelegated
}

// calculateVotingPowerForType calculates voting power for a specific proposal type
func (k Keeper) calculateVotingPowerForType(
	ctx sdk.Context,
	addr sdk.AccAddress,
	proposalType types.ProposalType,
) (math.LegacyDec, error) {
	// Create a dummy proposal to use existing calculation logic
	dummyProposal := types.Proposal{
		Type: proposalType,
	}
	
	return k.calculateVotingPower(ctx, dummyProposal, addr)
}