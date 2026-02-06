package types

import (
	"context"
)

// MsgServer defines the bridge module's Msg service
type MsgServer interface {
	SwapIn(context.Context, *SimpleMsgSwapIn) (*MsgSwapInResponse, error)
	SwapOut(context.Context, *SimpleMsgSwapOut) (*MsgSwapOutResponse, error)
	ValidatorDeposit(context.Context, *SimpleMsgValidatorDeposit) (*MsgValidatorDepositResponse, error)
	ProposeWithdrawal(context.Context, *SimpleMsgProposeWithdrawal) (*MsgProposeWithdrawalResponse, error)
	VoteWithdrawal(context.Context, *SimpleMsgVoteWithdrawal) (*MsgVoteWithdrawalResponse, error)
}

// Query service (placeholder for future implementation)
type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	SupportedAssets(context.Context, *QuerySupportedAssetsRequest) (*QuerySupportedAssetsResponse, error)
	ReserveBalances(context.Context, *QueryReserveBalancesRequest) (*QueryReserveBalancesResponse, error)
	SwapRecord(context.Context, *QuerySwapRecordRequest) (*QuerySwapRecordResponse, error)
	WithdrawalProposal(context.Context, *QueryWithdrawalProposalRequest) (*QueryWithdrawalProposalResponse, error)
}

// Query request/response types (placeholders)
type QueryParamsRequest struct{}
type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QuerySupportedAssetsRequest struct{}
type QuerySupportedAssetsResponse struct {
	Assets []SupportedAsset `json:"assets"`
}

type QueryReserveBalancesRequest struct{}
type QueryReserveBalancesResponse struct {
	Reserves []BridgeReserve `json:"reserves"`
}

type QuerySwapRecordRequest struct {
	SwapID uint64 `json:"swap_id"`
}
type QuerySwapRecordResponse struct {
	Record SwapRecord `json:"record"`
}

type QueryWithdrawalProposalRequest struct {
	ProposalID uint64 `json:"proposal_id"`
}
type QueryWithdrawalProposalResponse struct {
	Proposal ReserveWithdrawal `json:"proposal"`
}
