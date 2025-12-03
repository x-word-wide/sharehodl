package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&SimpleMsgRegisterValidatorTier{}, "validator/SimpleMsgRegisterValidatorTier", nil)
	cdc.RegisterConcrete(&SimpleMsgSubmitBusinessVerification{}, "validator/SimpleMsgSubmitBusinessVerification", nil)
	cdc.RegisterConcrete(&SimpleMsgCompleteVerification{}, "validator/SimpleMsgCompleteVerification", nil)
	cdc.RegisterConcrete(&SimpleMsgStakeInBusiness{}, "validator/SimpleMsgStakeInBusiness", nil)
	cdc.RegisterConcrete(&SimpleMsgClaimVerificationRewards{}, "validator/SimpleMsgClaimVerificationRewards", nil)
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