package types

import (
	"context"

	"google.golang.org/grpc"
)

// MsgServer defines the Msg service
type MsgServer interface {
	CreatePlan(goCtx context.Context, msg *MsgCreatePlan) (*MsgCreatePlanResponse, error)
	UpdatePlan(goCtx context.Context, msg *MsgUpdatePlan) (*MsgUpdatePlanResponse, error)
	CancelPlan(goCtx context.Context, msg *MsgCancelPlan) (*MsgCancelPlanResponse, error)
	ClaimAssets(goCtx context.Context, msg *MsgClaimAssets) (*MsgClaimAssetsResponse, error)
	CancelTrigger(goCtx context.Context, msg *MsgCancelTrigger) (*MsgCancelTriggerResponse, error)
}

// QueryServer defines the Query service
type QueryServer interface {
	Params(goCtx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
	Plan(goCtx context.Context, req *QueryPlanRequest) (*QueryPlanResponse, error)
	PlansByOwner(goCtx context.Context, req *QueryPlansByOwnerRequest) (*QueryPlansByOwnerResponse, error)
	PlansByBeneficiary(goCtx context.Context, req *QueryPlansByBeneficiaryRequest) (*QueryPlansByBeneficiaryResponse, error)
	Trigger(goCtx context.Context, req *QueryTriggerRequest) (*QueryTriggerResponse, error)
	Claim(goCtx context.Context, req *QueryClaimRequest) (*QueryClaimResponse, error)
	AllClaims(goCtx context.Context, req *QueryAllClaimsRequest) (*QueryAllClaimsResponse, error)
	LastActivity(goCtx context.Context, req *QueryLastActivityRequest) (*QueryLastActivityResponse, error)
}

// Query request and response types
type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QueryPlanRequest struct {
	PlanID uint64 `json:"plan_id"`
}

type QueryPlanResponse struct {
	Plan InheritancePlan `json:"plan"`
}

type QueryPlansByOwnerRequest struct {
	Owner string `json:"owner"`
}

type QueryPlansByOwnerResponse struct {
	Plans []InheritancePlan `json:"plans"`
}

type QueryPlansByBeneficiaryRequest struct {
	Beneficiary string `json:"beneficiary"`
}

type QueryPlansByBeneficiaryResponse struct {
	Plans []InheritancePlan `json:"plans"`
}

type QueryTriggerRequest struct {
	PlanID uint64 `json:"plan_id"`
}

type QueryTriggerResponse struct {
	Trigger SwitchTrigger `json:"trigger"`
}

type QueryClaimRequest struct {
	PlanID      uint64 `json:"plan_id"`
	Beneficiary string `json:"beneficiary"`
}

type QueryClaimResponse struct {
	Claim BeneficiaryClaim `json:"claim"`
}

type QueryAllClaimsRequest struct {
	PlanID uint64 `json:"plan_id"`
}

type QueryAllClaimsResponse struct {
	Claims []BeneficiaryClaim `json:"claims"`
}

type QueryLastActivityRequest struct {
	Address string `json:"address"`
}

type QueryLastActivityResponse struct {
	Activity ActivityRecord `json:"activity"`
	Found    bool           `json:"found"`
}

// ProtoMessage implementations
func (m *QueryParamsRequest) ProtoMessage()             {}
func (m *QueryParamsRequest) Reset()                    { *m = QueryParamsRequest{} }
func (m *QueryParamsResponse) ProtoMessage()            {}
func (m *QueryParamsResponse) Reset()                   { *m = QueryParamsResponse{} }
func (m *QueryPlanRequest) ProtoMessage()               {}
func (m *QueryPlanRequest) Reset()                      { *m = QueryPlanRequest{} }
func (m *QueryPlanResponse) ProtoMessage()              {}
func (m *QueryPlanResponse) Reset()                     { *m = QueryPlanResponse{} }
func (m *QueryPlansByOwnerRequest) ProtoMessage()       {}
func (m *QueryPlansByOwnerRequest) Reset()              { *m = QueryPlansByOwnerRequest{} }
func (m *QueryPlansByOwnerResponse) ProtoMessage()      {}
func (m *QueryPlansByOwnerResponse) Reset()             { *m = QueryPlansByOwnerResponse{} }
func (m *QueryPlansByBeneficiaryRequest) ProtoMessage() {}
func (m *QueryPlansByBeneficiaryRequest) Reset()        { *m = QueryPlansByBeneficiaryRequest{} }
func (m *QueryPlansByBeneficiaryResponse) ProtoMessage() {}
func (m *QueryPlansByBeneficiaryResponse) Reset()       { *m = QueryPlansByBeneficiaryResponse{} }
func (m *QueryTriggerRequest) ProtoMessage()            {}
func (m *QueryTriggerRequest) Reset()                   { *m = QueryTriggerRequest{} }
func (m *QueryTriggerResponse) ProtoMessage()           {}
func (m *QueryTriggerResponse) Reset()                  { *m = QueryTriggerResponse{} }
func (m *QueryClaimRequest) ProtoMessage()              {}
func (m *QueryClaimRequest) Reset()                     { *m = QueryClaimRequest{} }
func (m *QueryClaimResponse) ProtoMessage()             {}
func (m *QueryClaimResponse) Reset()                    { *m = QueryClaimResponse{} }
func (m *QueryAllClaimsRequest) ProtoMessage()          {}
func (m *QueryAllClaimsRequest) Reset()                 { *m = QueryAllClaimsRequest{} }
func (m *QueryAllClaimsResponse) ProtoMessage()         {}
func (m *QueryAllClaimsResponse) Reset()                { *m = QueryAllClaimsResponse{} }
func (m *QueryLastActivityRequest) ProtoMessage()       {}
func (m *QueryLastActivityRequest) Reset()              { *m = QueryLastActivityRequest{} }
func (m *QueryLastActivityResponse) ProtoMessage()      {}
func (m *QueryLastActivityResponse) Reset()             { *m = QueryLastActivityResponse{} }

// RegisterMsgServer registers the msg server
func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	// Registration would be done by gRPC
	// This is a placeholder for module compatibility
}

// RegisterQueryServer registers the query server
func RegisterQueryServer(s grpc.ServiceRegistrar, srv QueryServer) {
	// Registration would be done by gRPC
	// This is a placeholder for module compatibility
}
