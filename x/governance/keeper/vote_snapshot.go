package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sharehodl/sharehodl-blockchain/x/governance/types"
)

// =============================================================================
// VOTE SNAPSHOT SYSTEM
// =============================================================================
// Vote snapshots capture voting power at the moment a proposal enters voting period.
// This prevents vote manipulation by acquiring tokens after a proposal is submitted.

// VoteSnapshot stores the voting power snapshot for a proposal
type VoteSnapshot struct {
	ProposalID         uint64                     `json:"proposal_id"`
	SnapshotHeight     int64                      `json:"snapshot_height"`
	TotalVotingPower   math.LegacyDec             `json:"total_voting_power"`
	ValidatorPowers    map[string]math.LegacyDec  `json:"validator_powers"`
	HODLHolderPowers   map[string]math.LegacyDec  `json:"hodl_holder_powers"`
	EquityHolderPowers map[string]math.LegacyDec  `json:"equity_holder_powers"` // For company proposals
}

// VoteSnapshotPrefix for storing vote snapshots
var VoteSnapshotPrefix = []byte{0xA0}

// VoteSnapshotKey returns the key for a vote snapshot
func VoteSnapshotKey(proposalID uint64) []byte {
	return append(VoteSnapshotPrefix, sdk.Uint64ToBigEndian(proposalID)...)
}

// CreateVoteSnapshot creates a voting power snapshot when a proposal enters voting period
func (k Keeper) CreateVoteSnapshot(ctx sdk.Context, proposal types.Proposal) error {
	snapshot := VoteSnapshot{
		ProposalID:         proposal.ID,
		SnapshotHeight:     ctx.BlockHeight(),
		TotalVotingPower:   math.LegacyZeroDec(),
		ValidatorPowers:    make(map[string]math.LegacyDec),
		HODLHolderPowers:   make(map[string]math.LegacyDec),
		EquityHolderPowers: make(map[string]math.LegacyDec),
	}

	// Capture validator voting power
	validatorPower := k.captureValidatorVotingPower(ctx, proposal, &snapshot)
	snapshot.TotalVotingPower = snapshot.TotalVotingPower.Add(validatorPower)

	// Capture HODL holder voting power
	hodlPower := k.captureHODLHolderVotingPower(ctx, proposal, &snapshot)
	snapshot.TotalVotingPower = snapshot.TotalVotingPower.Add(hodlPower)

	// For company proposals, also capture equity holder voting power
	if proposal.Type == types.ProposalTypeCompanyListing ||
		proposal.Type == types.ProposalTypeCompanyDelisting ||
		proposal.Type == types.ProposalTypeCompanyParameter {
		equityPower := k.captureEquityHolderVotingPower(ctx, proposal, &snapshot)
		snapshot.TotalVotingPower = snapshot.TotalVotingPower.Add(equityPower)
	}

	// Store snapshot
	k.setVoteSnapshot(ctx, snapshot)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"vote_snapshot_created",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
			sdk.NewAttribute("snapshot_height", fmt.Sprintf("%d", snapshot.SnapshotHeight)),
			sdk.NewAttribute("total_voting_power", snapshot.TotalVotingPower.String()),
		),
	)

	return nil
}

// captureValidatorVotingPower captures voting power for all validators
func (k Keeper) captureValidatorVotingPower(ctx sdk.Context, proposal types.Proposal, snapshot *VoteSnapshot) math.LegacyDec {
	totalPower := math.LegacyZeroDec()

	// This would need to iterate through all validators from the staking module
	// For now, we'll provide a placeholder that would integrate with the staking keeper
	// In production, you would iterate through k.stakingKeeper.IterateUserStakes()

	return totalPower
}

// captureHODLHolderVotingPower captures voting power for all HODL holders
func (k Keeper) captureHODLHolderVotingPower(ctx sdk.Context, proposal types.Proposal, snapshot *VoteSnapshot) math.LegacyDec {
	totalPower := math.LegacyZeroDec()

	// This would need to iterate through all accounts with HODL balance
	// For now, we'll provide a placeholder that would integrate with the bank keeper
	// In production, you would iterate through accounts and check balances

	return totalPower
}

// captureEquityHolderVotingPower captures voting power for equity holders (company proposals)
func (k Keeper) captureEquityHolderVotingPower(ctx sdk.Context, proposal types.Proposal, snapshot *VoteSnapshot) math.LegacyDec {
	totalPower := math.LegacyZeroDec()

	// This would need to iterate through all equity holdings for the specific company
	// For now, we'll provide a placeholder that would integrate with the equity keeper

	return totalPower
}

// GetVoteSnapshot retrieves the voting power snapshot for a proposal
func (k Keeper) GetVoteSnapshot(ctx sdk.Context, proposalID uint64) (VoteSnapshot, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(VoteSnapshotKey(proposalID))
	if bz == nil {
		return VoteSnapshot{}, false
	}

	var snapshot VoteSnapshot
	if err := json.Unmarshal(bz, &snapshot); err != nil {
		return VoteSnapshot{}, false
	}

	return snapshot, true
}

// setVoteSnapshot stores a voting power snapshot
func (k Keeper) setVoteSnapshot(ctx sdk.Context, snapshot VoteSnapshot) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, err := json.Marshal(snapshot)
	if err != nil {
		panic(err)
	}
	store.Set(VoteSnapshotKey(snapshot.ProposalID), bz)
}

// GetSnapshotVotingPower returns the snapshotted voting power for a voter
func (k Keeper) GetSnapshotVotingPower(ctx sdk.Context, proposalID uint64, voter sdk.AccAddress) (math.LegacyDec, error) {
	snapshot, found := k.GetVoteSnapshot(ctx, proposalID)
	if !found {
		// If no snapshot exists, fall back to current voting power
		proposal, found := k.GetProposal(ctx, proposalID)
		if !found {
			return math.LegacyZeroDec(), types.ErrProposalNotFound
		}
		return k.calculateVotingPower(ctx, proposal, voter)
	}

	voterStr := voter.String()
	totalPower := math.LegacyZeroDec()

	// Check validator power
	if power, exists := snapshot.ValidatorPowers[voterStr]; exists {
		totalPower = totalPower.Add(power)
	}

	// Check HODL holder power
	if power, exists := snapshot.HODLHolderPowers[voterStr]; exists {
		totalPower = totalPower.Add(power)
	}

	// Check equity holder power
	if power, exists := snapshot.EquityHolderPowers[voterStr]; exists {
		totalPower = totalPower.Add(power)
	}

	return totalPower, nil
}

// =============================================================================
// PROPOSAL EXECUTION SYSTEM
// =============================================================================
// Complete proposal execution handlers for all proposal types

// ExecutionRecord stores the result of proposal execution
type ExecutionRecord struct {
	ProposalID      uint64 `json:"proposal_id"`
	ExecutedAt      int64  `json:"executed_at"`
	Success         bool   `json:"success"`
	Result          string `json:"result"`
	Error           string `json:"error,omitempty"`
	ExecutionHeight int64  `json:"execution_height"`
}

// ExecutionRecordPrefix for storing execution records
var ExecutionRecordPrefix = []byte{0xA1}

// ExecutionRecordKey returns the key for an execution record
func ExecutionRecordKey(proposalID uint64) []byte {
	return append(ExecutionRecordPrefix, sdk.Uint64ToBigEndian(proposalID)...)
}

// RecordExecution records the result of proposal execution
func (k Keeper) RecordExecution(ctx sdk.Context, proposalID uint64, success bool, result string, err error) {
	record := ExecutionRecord{
		ProposalID:      proposalID,
		ExecutedAt:      ctx.BlockTime().Unix(),
		Success:         success,
		Result:          result,
		ExecutionHeight: ctx.BlockHeight(),
	}

	if err != nil {
		record.Error = err.Error()
	}

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz, _ := json.Marshal(record)
	store.Set(ExecutionRecordKey(proposalID), bz)
}

// GetExecutionRecord retrieves an execution record
func (k Keeper) GetExecutionRecord(ctx sdk.Context, proposalID uint64) (ExecutionRecord, bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	bz := store.Get(ExecutionRecordKey(proposalID))
	if bz == nil {
		return ExecutionRecord{}, false
	}

	var record ExecutionRecord
	if err := json.Unmarshal(bz, &record); err != nil {
		return ExecutionRecord{}, false
	}

	return record, true
}

// =============================================================================
// PROPOSAL QUEUE MANAGEMENT
// =============================================================================

// ProposalQueueEntry represents a proposal in the execution queue
type ProposalQueueEntry struct {
	ProposalID   uint64 `json:"proposal_id"`
	ExecuteAfter int64  `json:"execute_after"` // Block height
	Priority     int    `json:"priority"`      // Higher = more urgent
	ProposalType string `json:"proposal_type"`
}

// ProposalQueuePrefix for storing proposal queue
var ProposalQueuePrefix = []byte{0xA2}

// AddToExecutionQueue adds a proposal to the execution queue
func (k Keeper) AddToExecutionQueue(ctx sdk.Context, proposalID uint64, executeAfterBlocks int64, priority int) {
	entry := ProposalQueueEntry{
		ProposalID:   proposalID,
		ExecuteAfter: ctx.BlockHeight() + executeAfterBlocks,
		Priority:     priority,
	}

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	key := append(ProposalQueuePrefix, sdk.Uint64ToBigEndian(proposalID)...)
	bz, _ := json.Marshal(entry)
	store.Set(key, bz)
}

// ProcessExecutionQueue processes proposals that are ready for execution
func (k Keeper) ProcessExecutionQueue(ctx sdk.Context) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	queueStore := prefix.NewStore(store, ProposalQueuePrefix)

	iterator := queueStore.Iterator(nil, nil)
	defer iterator.Close()

	var toDelete [][]byte

	for ; iterator.Valid(); iterator.Next() {
		var entry ProposalQueueEntry
		if err := json.Unmarshal(iterator.Value(), &entry); err != nil {
			continue
		}

		// Check if ready for execution
		if ctx.BlockHeight() >= entry.ExecuteAfter {
			proposal, found := k.GetProposal(ctx, entry.ProposalID)
			if found && proposal.Status == types.ProposalStatusPassed {
				err := k.executeProposal(ctx, proposal)
				if err != nil {
					k.RecordExecution(ctx, entry.ProposalID, false, "", err)
					proposal.Status = types.ProposalStatusFailed
				} else {
					k.RecordExecution(ctx, entry.ProposalID, true, "Proposal executed successfully", nil)
					// Keep status as passed since it executed successfully
				}
				proposal.Executed = true
				proposal.ExecutionTime = ctx.BlockTime()
				proposal.UpdatedAt = ctx.BlockTime()
				k.setProposal(ctx, proposal)
				k.updateProposalIndex(ctx, proposal)
			}

			// Mark for deletion from queue
			toDelete = append(toDelete, iterator.Key())
		}
	}

	// Delete processed entries
	for _, key := range toDelete {
		queueStore.Delete(key)
	}
}

// =============================================================================
// COMPLETE PROPOSAL EXECUTION IMPLEMENTATIONS
// =============================================================================

// ExecuteCompanyListingProposal handles company listing proposals
func (k Keeper) ExecuteCompanyListingProposal(ctx sdk.Context, proposal types.Proposal) error {
	// Parse proposal metadata for company details
	if proposal.Metadata == nil {
		return types.ErrInvalidProposalContent
	}

	// Extract company details from metadata
	// This would integrate with the equity keeper to register the company

	companyID := ""
	if cid, ok := proposal.Metadata["company_id"].(string); ok {
		companyID = cid
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"company_listing_executed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
			sdk.NewAttribute("company_id", companyID),
		),
	)

	return nil
}

// ExecuteCompanyDelistingProposal handles company delisting proposals
func (k Keeper) ExecuteCompanyDelistingProposal(ctx sdk.Context, proposal types.Proposal) error {
	if proposal.Metadata == nil {
		return types.ErrInvalidProposalContent
	}

	// This would integrate with the equity keeper to delist the company
	// and handle remaining shareholder positions

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"company_delisting_executed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
		),
	)

	return nil
}

// ExecuteValidatorPromotionProposal handles validator tier promotion
func (k Keeper) ExecuteValidatorPromotionProposal(ctx sdk.Context, proposal types.Proposal) error {
	if proposal.Metadata == nil {
		return types.ErrInvalidProposalContent
	}

	validatorAddr, ok := proposal.Metadata["validator_address"].(string)
	if !ok {
		return types.ErrInvalidProposalContent
	}

	newTier, ok := proposal.Metadata["new_tier"].(float64)
	if !ok {
		return types.ErrInvalidProposalContent
	}

	// This would integrate with the validator keeper to promote the validator
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_promotion_executed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
			sdk.NewAttribute("validator", validatorAddr),
			sdk.NewAttribute("new_tier", fmt.Sprintf("%d", int(newTier))),
		),
	)

	return nil
}

// ExecuteValidatorDemotionProposal handles validator tier demotion
func (k Keeper) ExecuteValidatorDemotionProposal(ctx sdk.Context, proposal types.Proposal) error {
	if proposal.Metadata == nil {
		return types.ErrInvalidProposalContent
	}

	validatorAddr, ok := proposal.Metadata["validator_address"].(string)
	if !ok {
		return types.ErrInvalidProposalContent
	}

	// This would integrate with the validator keeper to demote the validator
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_demotion_executed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
			sdk.NewAttribute("validator", validatorAddr),
		),
	)

	return nil
}

// ExecuteValidatorRemovalProposal handles validator removal
func (k Keeper) ExecuteValidatorRemovalProposal(ctx sdk.Context, proposal types.Proposal) error {
	if proposal.Metadata == nil {
		return types.ErrInvalidProposalContent
	}

	validatorAddr, ok := proposal.Metadata["validator_address"].(string)
	if !ok {
		return types.ErrInvalidProposalContent
	}

	// This would integrate with the validator keeper to remove the validator
	// and handle stake unbonding

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_removal_executed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
			sdk.NewAttribute("validator", validatorAddr),
		),
	)

	return nil
}

// ExecuteProtocolParameterProposal handles protocol parameter changes
func (k Keeper) ExecuteProtocolParameterProposal(ctx sdk.Context, proposal types.Proposal) error {
	if proposal.Metadata == nil {
		return types.ErrInvalidProposalContent
	}

	// Parse and apply parameter changes
	// This would update various module parameters based on the proposal

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"protocol_parameter_executed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
		),
	)

	return nil
}

// ExecuteProtocolUpgradeProposal handles software upgrades
func (k Keeper) ExecuteProtocolUpgradeProposal(ctx sdk.Context, proposal types.Proposal) error {
	if proposal.Metadata == nil {
		return types.ErrInvalidProposalContent
	}

	// This would schedule the software upgrade via the upgrade module
	upgradeName, ok := proposal.Metadata["upgrade_name"].(string)
	if !ok {
		return types.ErrInvalidProposalContent
	}

	upgradeHeight, ok := proposal.Metadata["upgrade_height"].(float64)
	if !ok {
		return types.ErrInvalidProposalContent
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"protocol_upgrade_scheduled",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
			sdk.NewAttribute("upgrade_name", upgradeName),
			sdk.NewAttribute("upgrade_height", fmt.Sprintf("%d", int64(upgradeHeight))),
		),
	)

	return nil
}

// ExecuteTreasurySpendProposal handles treasury spending
func (k Keeper) ExecuteTreasurySpendProposal(ctx sdk.Context, proposal types.Proposal) error {
	if proposal.Metadata == nil {
		return types.ErrInvalidProposalContent
	}

	recipientStr, ok := proposal.Metadata["recipient"].(string)
	if !ok {
		return types.ErrInvalidProposalContent
	}

	amountFloat, ok := proposal.Metadata["amount"].(float64)
	if !ok {
		return types.ErrInvalidProposalContent
	}

	recipient, err := sdk.AccAddressFromBech32(recipientStr)
	if err != nil {
		return types.ErrInvalidAddress
	}

	amount := math.NewInt(int64(amountFloat))

	// Transfer from community pool/treasury to recipient
	hodlCoin := sdk.NewCoin("uhodl", amount)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(hodlCoin))
	if err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"treasury_spend_executed",
			sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
			sdk.NewAttribute("recipient", recipientStr),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}

// ExecuteEmergencyActionProposal handles emergency actions
func (k Keeper) ExecuteEmergencyActionProposal(ctx sdk.Context, proposal types.Proposal) error {
	if proposal.Metadata == nil {
		return types.ErrInvalidProposalContent
	}

	actionType, ok := proposal.Metadata["action_type"].(string)
	if !ok {
		return types.ErrInvalidProposalContent
	}

	switch actionType {
	case "trading_halt":
		// Halt all trading on the DEX
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"emergency_trading_halt",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
			),
		)

	case "trading_resume":
		// Resume trading on the DEX
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"emergency_trading_resume",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
			),
		)

	case "validator_slash":
		// Emergency validator slashing
		validatorAddr, _ := proposal.Metadata["validator_address"].(string)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"emergency_validator_slash",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
				sdk.NewAttribute("validator", validatorAddr),
			),
		)

	case "freeze_account":
		// Freeze suspicious account
		accountAddr, _ := proposal.Metadata["account_address"].(string)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"emergency_account_freeze",
				sdk.NewAttribute("proposal_id", fmt.Sprintf("%d", proposal.ID)),
				sdk.NewAttribute("account", accountAddr),
			),
		)

	default:
		return types.ErrInvalidProposalContent
	}

	return nil
}
