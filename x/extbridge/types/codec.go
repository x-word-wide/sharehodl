package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterCodec registers the necessary x/extbridge interfaces and concrete types
// on the provided LegacyAmino codec.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgObserveDeposit{}, "extbridge/MsgObserveDeposit", nil)
	cdc.RegisterConcrete(&MsgAttestDeposit{}, "extbridge/MsgAttestDeposit", nil)
	cdc.RegisterConcrete(&MsgRequestWithdrawal{}, "extbridge/MsgRequestWithdrawal", nil)
	cdc.RegisterConcrete(&MsgSubmitTSSSignature{}, "extbridge/MsgSubmitTSSSignature", nil)
	cdc.RegisterConcrete(&MsgUpdateCircuitBreaker{}, "extbridge/MsgUpdateCircuitBreaker", nil)
	cdc.RegisterConcrete(&MsgAddExternalChain{}, "extbridge/MsgAddExternalChain", nil)
	cdc.RegisterConcrete(&MsgAddExternalAsset{}, "extbridge/MsgAddExternalAsset", nil)
}

// RegisterInterfaces registers the x/extbridge interfaces types with the interface registry
// Note: Using JSON encoding, not protobuf, so this is simplified
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// Not using protobuf gRPC - messages are handled directly through keeper
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(Amino)
	Amino.Seal()
}
