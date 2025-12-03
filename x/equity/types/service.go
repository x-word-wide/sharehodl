package types

import (
	"context"
)

// MsgServer defines the module's message service.
type MsgServer interface {
	CreateCompany(context.Context, *SimpleMsgCreateCompany) (*MsgCreateCompanyResponse, error)
	CreateShareClass(context.Context, *SimpleMsgCreateShareClass) (*MsgCreateShareClassResponse, error)
	IssueShares(context.Context, *SimpleMsgIssueShares) (*MsgIssueSharesResponse, error)
	TransferShares(context.Context, *SimpleMsgTransferShares) (*MsgTransferSharesResponse, error)
	UpdateCompany(context.Context, *SimpleMsgUpdateCompany) (*MsgUpdateCompanyResponse, error)
}

// RegisterMsgServer registers the message server
func RegisterMsgServer(server interface{}, impl MsgServer) {
	// In a real implementation this would register the gRPC server
	// For our simplified version, we'll handle this in the module
}