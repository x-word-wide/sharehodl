package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterCodec registers the necessary x/inheritance interfaces and concrete types
// on the provided LegacyAmino codec.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreatePlan{}, "inheritance/CreatePlan", nil)
	cdc.RegisterConcrete(&MsgUpdatePlan{}, "inheritance/UpdatePlan", nil)
	cdc.RegisterConcrete(&MsgCancelPlan{}, "inheritance/CancelPlan", nil)
	cdc.RegisterConcrete(&MsgClaimAssets{}, "inheritance/ClaimAssets", nil)
	cdc.RegisterConcrete(&MsgCancelTrigger{}, "inheritance/CancelTrigger", nil)
}

// RegisterInterfaces registers the x/inheritance interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// Note: This module uses legacy amino codec for message serialization.
	// Proto-based registration is not used to avoid type URL conflicts.
	_ = registry // Silence unused parameter warning
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(Amino)
	Amino.Seal()
}
