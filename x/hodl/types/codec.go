package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&SimpleMsgMintHODL{}, "hodl/SimpleMsgMintHODL", nil)
	cdc.RegisterConcrete(&SimpleMsgBurnHODL{}, "hodl/SimpleMsgBurnHODL", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// Simplified registration for now
	// Will add protobuf implementations later
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(Amino)
	Amino.Seal()
}