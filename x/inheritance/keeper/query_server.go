package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sharehodl/sharehodl-blockchain/x/inheritance/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// Params returns module parameters
func (qs queryServer) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := qs.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// Plan returns a specific inheritance plan
func (qs queryServer) Plan(goCtx context.Context, req *types.QueryPlanRequest) (*types.QueryPlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	plan, found := qs.GetPlan(ctx, req.PlanID)
	if !found {
		return nil, types.ErrPlanNotFound
	}

	return &types.QueryPlanResponse{
		Plan: plan,
	}, nil
}

// PlansByOwner returns all plans owned by an address
func (qs queryServer) PlansByOwner(goCtx context.Context, req *types.QueryPlansByOwnerRequest) (*types.QueryPlansByOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	plans := qs.GetPlansByOwner(ctx, owner)

	return &types.QueryPlansByOwnerResponse{
		Plans: plans,
	}, nil
}

// PlansByBeneficiary returns all plans where address is a beneficiary
func (qs queryServer) PlansByBeneficiary(goCtx context.Context, req *types.QueryPlansByBeneficiaryRequest) (*types.QueryPlansByBeneficiaryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	beneficiary, err := sdk.AccAddressFromBech32(req.Beneficiary)
	if err != nil {
		return nil, err
	}

	plans := qs.GetPlansByBeneficiary(ctx, beneficiary)

	return &types.QueryPlansByBeneficiaryResponse{
		Plans: plans,
	}, nil
}

// Trigger returns an active trigger for a plan
func (qs queryServer) Trigger(goCtx context.Context, req *types.QueryTriggerRequest) (*types.QueryTriggerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	trigger, found := qs.GetActiveTrigger(ctx, req.PlanID)
	if !found {
		return nil, types.ErrTriggerNotFound
	}

	return &types.QueryTriggerResponse{
		Trigger: trigger,
	}, nil
}

// Claim returns a beneficiary claim
func (qs queryServer) Claim(goCtx context.Context, req *types.QueryClaimRequest) (*types.QueryClaimResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	beneficiary, err := sdk.AccAddressFromBech32(req.Beneficiary)
	if err != nil {
		return nil, err
	}

	claim, found := qs.GetClaim(ctx, req.PlanID, beneficiary)
	if !found {
		return nil, types.ErrClaimNotFound
	}

	return &types.QueryClaimResponse{
		Claim: claim,
	}, nil
}

// AllClaims returns all claims for a plan
func (qs queryServer) AllClaims(goCtx context.Context, req *types.QueryAllClaimsRequest) (*types.QueryAllClaimsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	claims := qs.GetAllClaimsForPlan(ctx, req.PlanID)

	return &types.QueryAllClaimsResponse{
		Claims: claims,
	}, nil
}

// LastActivity returns the last activity for an address
func (qs queryServer) LastActivity(goCtx context.Context, req *types.QueryLastActivityRequest) (*types.QueryLastActivityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	activity, found := qs.GetLastActivity(ctx, address)
	if !found {
		return &types.QueryLastActivityResponse{
			Activity: types.ActivityRecord{},
			Found:    false,
		}, nil
	}

	return &types.QueryLastActivityResponse{
		Activity: activity,
		Found:    true,
	}, nil
}
