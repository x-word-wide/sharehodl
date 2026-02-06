package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterLegacyAminoCodec registers the legacy amino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgUpdateParams{}, "feeabstraction/MsgUpdateParams", nil)
	cdc.RegisterConcrete(&MsgGrantTreasuryAllowance{}, "feeabstraction/MsgGrantTreasuryAllowance", nil)
	cdc.RegisterConcrete(&MsgRevokeTreasuryAllowance{}, "feeabstraction/MsgRevokeTreasuryAllowance", nil)
	cdc.RegisterConcrete(&MsgFundTreasury{}, "feeabstraction/MsgFundTreasury", nil)
}

// RegisterInterfaces registers the interface types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	// Note: This module uses legacy amino codec for message serialization.
	// Proto-based registration is not used to avoid type URL conflicts.
	// For full gRPC support, proper protobuf files would need to be generated.
	_ = registry // Silence unused parameter warning
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
}
