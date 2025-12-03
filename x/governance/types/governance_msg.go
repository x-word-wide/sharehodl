package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message types for governance operations

// MsgSubmitProposal represents a message to submit a governance proposal
type MsgSubmitProposal struct {
	Submitter        string         `json:"submitter"`
	ProposalType     string         `json:"proposal_type"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	InitialDeposit   math.Int       `json:"initial_deposit"`
	VotingPeriodDays uint32         `json:"voting_period_days"`
	QuorumRequired   math.LegacyDec `json:"quorum_required"`
	ThresholdRequired math.LegacyDec `json:"threshold_required"`
	
	// Type-specific fields (union-like structure)
	CompanyID        *uint64         `json:"company_id,omitempty"`          // For company proposals
	ValidatorAddress *string         `json:"validator_address,omitempty"`   // For validator proposals
	ParameterChanges *string         `json:"parameter_changes,omitempty"`   // For parameter changes (JSON)
	CodeID           *uint64         `json:"code_id,omitempty"`            // For software upgrades
	Amount           *math.Int       `json:"amount,omitempty"`             // For treasury/spending proposals
	Recipient        *string         `json:"recipient,omitempty"`          // For treasury proposals
}

// GetSigners implements the Msg interface
func (msg MsgSubmitProposal) GetSigners() []sdk.AccAddress {
	submitter, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{submitter}
}

// ValidateBasic implements the Msg interface
func (msg MsgSubmitProposal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.Title == "" || len(msg.Title) > 200 {
		return ErrInvalidProposal
	}

	if msg.Description == "" || len(msg.Description) > 5000 {
		return ErrInvalidProposal
	}

	validTypes := map[string]bool{
		"company_listing": true, "company_delisting": true, "company_parameter": true,
		"validator_promotion": true, "validator_demotion": true, "validator_removal": true,
		"protocol_parameter": true, "protocol_upgrade": true,
		"treasury_spend": true, "emergency_action": true,
	}
	if !validTypes[msg.ProposalType] {
		return ErrInvalidProposal
	}

	if msg.InitialDeposit.IsNegative() {
		return ErrInvalidProposal
	}

	if msg.VotingPeriodDays == 0 || msg.VotingPeriodDays > 90 {
		return ErrInvalidProposal
	}

	if msg.QuorumRequired.IsNegative() || msg.QuorumRequired.GT(math.LegacyOneDec()) {
		return ErrInvalidProposal
	}

	if msg.ThresholdRequired.IsNegative() || msg.ThresholdRequired.GT(math.LegacyOneDec()) {
		return ErrInvalidProposal
	}

	return nil
}

// MsgSubmitProposalResponse is the response type for MsgSubmitProposal
type MsgSubmitProposalResponse struct {
	ProposalID uint64 `json:"proposal_id"`
	Success    bool   `json:"success"`
}

// MsgVote represents a message to vote on a proposal
type MsgVote struct {
	Voter      string `json:"voter"`
	ProposalID uint64 `json:"proposal_id"`
	Option     string `json:"option"`     // "yes", "no", "abstain", "veto"
	Weight     math.LegacyDec `json:"weight,omitempty"` // For weighted voting
	Reason     string `json:"reason,omitempty"`
}

// GetSigners implements the Msg interface
func (msg MsgVote) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}

// ValidateBasic implements the Msg interface
func (msg MsgVote) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.ProposalID == 0 {
		return ErrProposalNotFound
	}

	validOptions := map[string]bool{
		"yes": true, "no": true, "abstain": true, "veto": true,
	}
	if !validOptions[msg.Option] {
		return ErrInvalidVote
	}

	if msg.Weight.IsNegative() || msg.Weight.GT(math.LegacyOneDec()) {
		return ErrInvalidVote
	}

	return nil
}

// MsgVoteResponse is the response type for MsgVote
type MsgVoteResponse struct {
	Success bool `json:"success"`
}

// MsgVoteWeighted represents a message to cast weighted votes on a proposal
type MsgVoteWeighted struct {
	Voter      string             `json:"voter"`
	ProposalID uint64             `json:"proposal_id"`
	Options    []WeightedVoteOption `json:"options"`
	Reason     string             `json:"reason,omitempty"`
}

// WeightedVoteOption represents a weighted vote option
type WeightedVoteOption struct {
	Option string         `json:"option"`
	Weight math.LegacyDec `json:"weight"`
}

// GetSigners implements the Msg interface
func (msg MsgVoteWeighted) GetSigners() []sdk.AccAddress {
	voter, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{voter}
}

// ValidateBasic implements the Msg interface
func (msg MsgVoteWeighted) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Voter)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.ProposalID == 0 {
		return ErrProposalNotFound
	}

	if len(msg.Options) == 0 || len(msg.Options) > 4 {
		return ErrInvalidVote
	}

	totalWeight := math.LegacyZeroDec()
	validOptions := map[string]bool{
		"yes": true, "no": true, "abstain": true, "veto": true,
	}

	for _, option := range msg.Options {
		if !validOptions[option.Option] {
			return ErrInvalidVote
		}
		if option.Weight.IsNegative() {
			return ErrInvalidVote
		}
		totalWeight = totalWeight.Add(option.Weight)
	}

	if !totalWeight.Equal(math.LegacyOneDec()) {
		return ErrInvalidVote
	}

	return nil
}

// MsgVoteWeightedResponse is the response type for MsgVoteWeighted
type MsgVoteWeightedResponse struct {
	Success bool `json:"success"`
}

// MsgDeposit represents a message to deposit tokens to a proposal
type MsgDeposit struct {
	Depositor  string   `json:"depositor"`
	ProposalID uint64   `json:"proposal_id"`
	Amount     math.Int `json:"amount"`
}

// GetSigners implements the Msg interface
func (msg MsgDeposit) GetSigners() []sdk.AccAddress {
	depositor, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{depositor}
}

// ValidateBasic implements the Msg interface
func (msg MsgDeposit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Depositor)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.ProposalID == 0 {
		return ErrProposalNotFound
	}

	if msg.Amount.IsNegative() || msg.Amount.IsZero() {
		return ErrInvalidDeposit
	}

	return nil
}

// MsgDepositResponse is the response type for MsgDeposit
type MsgDepositResponse struct {
	Success bool `json:"success"`
}

// MsgCancelProposal represents a message to cancel a proposal
type MsgCancelProposal struct {
	Proposer   string `json:"proposer"`
	ProposalID uint64 `json:"proposal_id"`
	Reason     string `json:"reason"`
}

// GetSigners implements the Msg interface
func (msg MsgCancelProposal) GetSigners() []sdk.AccAddress {
	proposer, err := sdk.AccAddressFromBech32(msg.Proposer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{proposer}
}

// ValidateBasic implements the Msg interface
func (msg MsgCancelProposal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Proposer)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.ProposalID == 0 {
		return ErrProposalNotFound
	}

	if msg.Reason == "" {
		return ErrInvalidReason
	}

	return nil
}

// MsgCancelProposalResponse is the response type for MsgCancelProposal
type MsgCancelProposalResponse struct {
	Success bool `json:"success"`
}

// MsgSetGovernanceParams represents a message to set governance parameters
type MsgSetGovernanceParams struct {
	Authority         string         `json:"authority"`
	MinDeposit        math.Int       `json:"min_deposit"`
	MaxDepositPeriod  uint32         `json:"max_deposit_period_days"`
	VotingPeriod      uint32         `json:"voting_period_days"`
	Quorum            math.LegacyDec `json:"quorum"`
	Threshold         math.LegacyDec `json:"threshold"`
	VetoThreshold     math.LegacyDec `json:"veto_threshold"`
	BurnDeposits      bool           `json:"burn_deposits"`
	BurnVoteVeto      bool           `json:"burn_vote_veto"`
	MinInitialDeposit math.LegacyDec `json:"min_initial_deposit_ratio"`
}

// GetSigners implements the Msg interface
func (msg MsgSetGovernanceParams) GetSigners() []sdk.AccAddress {
	authority, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{authority}
}

// ValidateBasic implements the Msg interface
func (msg MsgSetGovernanceParams) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Authority)
	if err != nil {
		return ErrUnauthorized
	}

	if msg.MinDeposit.IsNegative() {
		return ErrInvalidGovernanceParams
	}

	if msg.MaxDepositPeriod == 0 || msg.MaxDepositPeriod > 30 {
		return ErrInvalidGovernanceParams
	}

	if msg.VotingPeriod == 0 || msg.VotingPeriod > 90 {
		return ErrInvalidGovernanceParams
	}

	if msg.Quorum.IsNegative() || msg.Quorum.GT(math.LegacyOneDec()) {
		return ErrInvalidGovernanceParams
	}

	if msg.Threshold.IsNegative() || msg.Threshold.GT(math.LegacyOneDec()) {
		return ErrInvalidGovernanceParams
	}

	if msg.VetoThreshold.IsNegative() || msg.VetoThreshold.GT(math.LegacyOneDec()) {
		return ErrInvalidGovernanceParams
	}

	if msg.MinInitialDeposit.IsNegative() || msg.MinInitialDeposit.GT(math.LegacyOneDec()) {
		return ErrInvalidGovernanceParams
	}

	return nil
}

// MsgSetGovernanceParamsResponse is the response type for MsgSetGovernanceParams
type MsgSetGovernanceParamsResponse struct {
	Success bool `json:"success"`
}

// Helper functions

// NewMsgSubmitProposal creates a new MsgSubmitProposal instance
func NewMsgSubmitProposal(
	submitter string,
	proposalType string,
	title string,
	description string,
	initialDeposit math.Int,
	votingPeriodDays uint32,
	quorumRequired math.LegacyDec,
	thresholdRequired math.LegacyDec,
) *MsgSubmitProposal {
	return &MsgSubmitProposal{
		Submitter:         submitter,
		ProposalType:      proposalType,
		Title:             title,
		Description:       description,
		InitialDeposit:    initialDeposit,
		VotingPeriodDays:  votingPeriodDays,
		QuorumRequired:    quorumRequired,
		ThresholdRequired: thresholdRequired,
	}
}

// NewMsgVote creates a new MsgVote instance
func NewMsgVote(voter string, proposalID uint64, option string, reason string) *MsgVote {
	return &MsgVote{
		Voter:      voter,
		ProposalID: proposalID,
		Option:     option,
		Reason:     reason,
	}
}

// NewMsgVoteWeighted creates a new MsgVoteWeighted instance
func NewMsgVoteWeighted(
	voter string,
	proposalID uint64,
	options []WeightedVoteOption,
	reason string,
) *MsgVoteWeighted {
	return &MsgVoteWeighted{
		Voter:      voter,
		ProposalID: proposalID,
		Options:    options,
		Reason:     reason,
	}
}

// NewMsgDeposit creates a new MsgDeposit instance
func NewMsgDeposit(depositor string, proposalID uint64, amount math.Int) *MsgDeposit {
	return &MsgDeposit{
		Depositor:  depositor,
		ProposalID: proposalID,
		Amount:     amount,
	}
}