package types

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// MsgClient is the client API for Msg service.
type MsgClient interface {
	// ObserveDeposit allows validators to observe deposits on external chains
	ObserveDeposit(ctx context.Context, in *MsgObserveDeposit, opts ...grpc.CallOption) (*MsgObserveDepositResponse, error)
	// AttestDeposit allows validators to attest to deposits
	AttestDeposit(ctx context.Context, in *MsgAttestDeposit, opts ...grpc.CallOption) (*MsgAttestDepositResponse, error)
	// RequestWithdrawal allows users to request withdrawals to external chains
	RequestWithdrawal(ctx context.Context, in *MsgRequestWithdrawal, opts ...grpc.CallOption) (*MsgRequestWithdrawalResponse, error)
	// SubmitTSSSignature allows validators to submit TSS signature shares
	SubmitTSSSignature(ctx context.Context, in *MsgSubmitTSSSignature, opts ...grpc.CallOption) (*MsgSubmitTSSSignatureResponse, error)
	// UpdateCircuitBreaker allows governance to update the circuit breaker
	UpdateCircuitBreaker(ctx context.Context, in *MsgUpdateCircuitBreaker, opts ...grpc.CallOption) (*MsgUpdateCircuitBreakerResponse, error)
	// AddExternalChain allows governance to add a new external chain
	AddExternalChain(ctx context.Context, in *MsgAddExternalChain, opts ...grpc.CallOption) (*MsgAddExternalChainResponse, error)
	// AddExternalAsset allows governance to add a new external asset
	AddExternalAsset(ctx context.Context, in *MsgAddExternalAsset, opts ...grpc.CallOption) (*MsgAddExternalAssetResponse, error)
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// ObserveDeposit allows validators to observe deposits on external chains
	ObserveDeposit(context.Context, *MsgObserveDeposit) (*MsgObserveDepositResponse, error)
	// AttestDeposit allows validators to attest to deposits
	AttestDeposit(context.Context, *MsgAttestDeposit) (*MsgAttestDepositResponse, error)
	// RequestWithdrawal allows users to request withdrawals to external chains
	RequestWithdrawal(context.Context, *MsgRequestWithdrawal) (*MsgRequestWithdrawalResponse, error)
	// SubmitTSSSignature allows validators to submit TSS signature shares
	SubmitTSSSignature(context.Context, *MsgSubmitTSSSignature) (*MsgSubmitTSSSignatureResponse, error)
	// UpdateCircuitBreaker allows governance to update the circuit breaker
	UpdateCircuitBreaker(context.Context, *MsgUpdateCircuitBreaker) (*MsgUpdateCircuitBreakerResponse, error)
	// AddExternalChain allows governance to add a new external chain
	AddExternalChain(context.Context, *MsgAddExternalChain) (*MsgAddExternalChainResponse, error)
	// AddExternalAsset allows governance to add a new external asset
	AddExternalAsset(context.Context, *MsgAddExternalAsset) (*MsgAddExternalAssetResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct{}

func (*UnimplementedMsgServer) ObserveDeposit(context.Context, *MsgObserveDeposit) (*MsgObserveDepositResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ObserveDeposit not implemented")
}
func (*UnimplementedMsgServer) AttestDeposit(context.Context, *MsgAttestDeposit) (*MsgAttestDepositResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AttestDeposit not implemented")
}
func (*UnimplementedMsgServer) RequestWithdrawal(context.Context, *MsgRequestWithdrawal) (*MsgRequestWithdrawalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestWithdrawal not implemented")
}
func (*UnimplementedMsgServer) SubmitTSSSignature(context.Context, *MsgSubmitTSSSignature) (*MsgSubmitTSSSignatureResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitTSSSignature not implemented")
}
func (*UnimplementedMsgServer) UpdateCircuitBreaker(context.Context, *MsgUpdateCircuitBreaker) (*MsgUpdateCircuitBreakerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCircuitBreaker not implemented")
}
func (*UnimplementedMsgServer) AddExternalChain(context.Context, *MsgAddExternalChain) (*MsgAddExternalChainResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddExternalChain not implemented")
}
func (*UnimplementedMsgServer) AddExternalAsset(context.Context, *MsgAddExternalAsset) (*MsgAddExternalAssetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddExternalAsset not implemented")
}

// Response types
type MsgObserveDepositResponse struct {
	DepositID uint64 `json:"deposit_id" yaml:"deposit_id"`
}

type MsgAttestDepositResponse struct {
	Attestations uint64 `json:"attestations" yaml:"attestations"`
	Completed    bool   `json:"completed" yaml:"completed"`
}

type MsgRequestWithdrawalResponse struct {
	WithdrawalID uint64 `json:"withdrawal_id" yaml:"withdrawal_id"`
}

type MsgSubmitTSSSignatureResponse struct {
	SessionID uint64 `json:"session_id" yaml:"session_id"`
	Completed bool   `json:"completed" yaml:"completed"`
}

type MsgUpdateCircuitBreakerResponse struct{}

type MsgAddExternalChainResponse struct{}

type MsgAddExternalAssetResponse struct{}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "sharehodl.extbridge.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "sharehodl/extbridge/v1/tx.proto",
}
