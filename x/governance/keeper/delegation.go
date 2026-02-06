package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
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

	if _, err := sdk.AccAddressFromBech32(delegate); err != nil {
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

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"delegate_voting_power",
			sdk.NewAttribute("delegator", delegator),
			sdk.NewAttribute("delegate", delegate),
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyID)),
			sdk.NewAttribute("proposal_type", proposalType.String()),
			sdk.NewAttribute("voting_power", votingPower.String()),
			sdk.NewAttribute("expiry_height", fmt.Sprintf("%d", expiryHeight)),
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

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"revoke_delegation",
			sdk.NewAttribute("delegator", delegator),
			sdk.NewAttribute("delegate", delegate),
			sdk.NewAttribute("company_id", fmt.Sprintf("%d", companyID)),
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

	// Track delegators whose voting power was used
	var votedDelegators []sdk.AccAddress

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

				// Track this delegator
				votedDelegators = append(votedDelegators, delegatorAddr)
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
		ProposalID:  proposalID,
		Voter:       delegate,
		Option:      option,
		Weight:      math.LegacyOneDec(),
		VotingPower: totalVotingPower.TruncateInt(),
	}

	// Store vote
	k.setVote(ctx, vote)

	// Update tally
	k.updateTallyResult(ctx, proposal, vote, math.LegacyOneDec())

	// SECURITY FIX: Mark all delegators as having voted to prevent double voting
	// When a delegate votes on behalf of delegators, the delegators should not be
	// able to vote again directly on the same proposal
	for _, delegatorAddr := range votedDelegators {
		delegatorVote := types.Vote{
			ProposalID:  proposalID,
			Voter:       delegatorAddr.String(),
			Option:      option,
			Weight:      math.LegacyZeroDec(), // Zero weight since power already counted via delegate
			VotingPower: math.ZeroInt(),
			Justification: fmt.Sprintf("Voted via delegate: %s", delegate),
			VotedAt:     ctx.BlockTime(),
		}
		k.setVote(ctx, delegatorVote)
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote_as_delegate",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute("delegate", delegate),
			sdk.NewAttribute("option", option.String()),
			sdk.NewAttribute("own_power", ownVotingPower.String()),
			sdk.NewAttribute("delegated_power", totalDelegatedPower.String()),
			sdk.NewAttribute("total_power", totalVotingPower.String()),
			sdk.NewAttribute("delegators_count", fmt.Sprintf("%d", len(votedDelegators))),
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
	store := k.storeService.OpenKVStore(ctx)
	key := types.DelegationKey(delegation.Delegator, delegation.CompanyID, uint32(delegation.ProposalType))
	value, _ := json.Marshal(delegation)
	store.Set(key, value)
}

// GetDelegation retrieves a delegation
func (k Keeper) GetDelegation(
	ctx sdk.Context,
	delegator sdk.AccAddress,
	companyID uint64,
	proposalType types.ProposalType,
) (types.VoteDelegation, bool) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.DelegationKey(delegator.String(), companyID, uint32(proposalType))

	value, err := store.Get(key)
	if err != nil || value == nil {
		return types.VoteDelegation{}, false
	}

	var delegation types.VoteDelegation
	if err := json.Unmarshal(value, &delegation); err != nil {
		return types.VoteDelegation{}, false
	}
	return delegation, true
}

// GetDelegationsForDelegate gets all delegations for a delegate
func (k Keeper) GetDelegationsForDelegate(
	ctx sdk.Context,
	delegate sdk.AccAddress,
	proposalType types.ProposalType,
) []types.VoteDelegation {
	store := k.storeService.OpenKVStore(ctx)

	var delegations []types.VoteDelegation

	// Use prefix iteration - need to build prefix without delegator
	prefixBytes := append(types.DelegateIndexPrefix, []byte(delegate.String())...)
	prefixBytes = append(prefixBytes, 0x00) // separator
	iter, err := store.Iterator(prefixBytes, types.PrefixEndBytes(prefixBytes))
	if err != nil {
		return delegations
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var delegation types.VoteDelegation
		if err := json.Unmarshal(iter.Value(), &delegation); err != nil {
			continue
		}
		delegations = append(delegations, delegation)
	}

	return delegations
}

// removeDelegation removes a delegation from storage
func (k Keeper) removeDelegation(ctx sdk.Context, delegation types.VoteDelegation) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.DelegationKey(delegation.Delegator, delegation.CompanyID, uint32(delegation.ProposalType))
	store.Delete(key)

	// Also remove from delegate index
	delegateKey := types.DelegateIndexKey(delegation.Delegate, uint32(delegation.ProposalType), delegation.Delegator)
	store.Delete(delegateKey)
}

// getTotalDelegatedPower gets total power already delegated by an address
func (k Keeper) getTotalDelegatedPower(
	ctx sdk.Context,
	addr sdk.AccAddress,
	companyID uint64,
	proposalType types.ProposalType,
) math.LegacyDec {
	store := k.storeService.OpenKVStore(ctx)

	totalDelegated := math.LegacyZeroDec()

	// Use prefix iteration for delegator's delegations
	prefixBytes := append(types.DelegatorIndexPrefix, []byte(addr.String())...)
	prefixBytes = append(prefixBytes, 0x00) // separator
	iter, err := store.Iterator(prefixBytes, types.PrefixEndBytes(prefixBytes))
	if err != nil {
		return totalDelegated
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var delegation types.VoteDelegation
		if err := json.Unmarshal(iter.Value(), &delegation); err != nil {
			continue
		}

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
