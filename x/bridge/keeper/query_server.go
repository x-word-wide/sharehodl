package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/bridge/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// Params queries the module params
func (qs queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := qs.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// SupportedAssets queries all supported assets
func (qs queryServer) SupportedAssets(goCtx context.Context, req *types.QuerySupportedAssetsRequest) (*types.QuerySupportedAssetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	assets := qs.Keeper.GetAllSupportedAssets(ctx)

	return &types.QuerySupportedAssetsResponse{
		Assets: assets,
	}, nil
}

// ReserveBalances queries all reserve balances
func (qs queryServer) ReserveBalances(goCtx context.Context, req *types.QueryReserveBalancesRequest) (*types.QueryReserveBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	reserves := qs.Keeper.GetAllReserveBalances(ctx)

	return &types.QueryReserveBalancesResponse{
		Reserves: reserves,
	}, nil
}

// SwapRecord queries a specific swap record
func (qs queryServer) SwapRecord(goCtx context.Context, req *types.QuerySwapRecordRequest) (*types.QuerySwapRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	record, found := qs.Keeper.GetSwapRecord(ctx, req.SwapID)
	if !found {
		return nil, types.ErrSwapFailed
	}

	return &types.QuerySwapRecordResponse{
		Record: record,
	}, nil
}

// WithdrawalProposal queries a specific withdrawal proposal
func (qs queryServer) WithdrawalProposal(goCtx context.Context, req *types.QueryWithdrawalProposalRequest) (*types.QueryWithdrawalProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	proposal, found := qs.Keeper.GetWithdrawalProposal(ctx, req.ProposalID)
	if !found {
		return nil, types.ErrProposalNotFound
	}

	return &types.QueryWithdrawalProposalResponse{
		Proposal: proposal,
	}, nil
}
