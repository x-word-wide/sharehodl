package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&SimpleMsgSwapIn{}, "bridge/SwapIn", nil)
	cdc.RegisterConcrete(&SimpleMsgSwapOut{}, "bridge/SwapOut", nil)
	cdc.RegisterConcrete(&SimpleMsgValidatorDeposit{}, "bridge/ValidatorDeposit", nil)
	cdc.RegisterConcrete(&SimpleMsgProposeWithdrawal{}, "bridge/ProposeWithdrawal", nil)
	cdc.RegisterConcrete(&SimpleMsgVoteWithdrawal{}, "bridge/VoteWithdrawal", nil)
}

// RegisterInterfaces registers the module interfaces
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// Simple message types don't need protobuf registration
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(Amino)
	Amino.Seal()
}
