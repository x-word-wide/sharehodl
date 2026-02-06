package app

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"
	txsigning "cosmossdk.io/x/tx/signing"
)

// EncodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

// MakeTestEncodingConfig creates an EncodingConfig for CLI and testing
func MakeTestEncodingConfig() EncodingConfig {
	// Create address codecs for our chain's bech32 prefixes
	addressCodec := address.NewBech32Codec(Bech32PrefixAccAddr)
	validatorAddressCodec := address.NewBech32Codec(Bech32PrefixValAddr)

	// CRITICAL: Create InterfaceRegistry with proper address codecs.
	// This is required for tx signing, simulation (gas estimation) and proper
	// address conversion. Without this, CLI transactions fail with
	// "InterfaceRegistry requires a proper address codec".
	signingOptions := txsigning.Options{
		FileResolver:          proto.HybridResolver,
		AddressCodec:          addressCodec,
		ValidatorAddressCodec: validatorAddressCodec,
	}
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     proto.HybridResolver,
		SigningOptions: signingOptions,
	})
	if err != nil {
		panic(err)
	}

	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()

	// CRITICAL: SDK v0.54-alpha requires creating a SigningContext with proper
	// address codecs for signature verification to work. Without this, signature
	// verification fails with "unable to verify single signer signature".
	signingContext, err := txsigning.NewContext(signingOptions)
	if err != nil {
		panic(err)
	}

	txConfig, err := tx.NewTxConfigWithOptions(protoCodec, tx.ConfigOptions{
		EnabledSignModes: tx.DefaultSignModes,
		SigningContext:   signingContext,
	})
	if err != nil {
		panic(err)
	}

	std.RegisterLegacyAminoCodec(legacyAmino)
	std.RegisterInterfaces(interfaceRegistry)

	// CRITICAL: Register all module interfaces for proper type resolution
	// This ensures the encoding config has all necessary type registrations
	// for CLI commands and testing
	ModuleBasics.RegisterLegacyAminoCodec(legacyAmino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             protoCodec,
		TxConfig:          txConfig,
		Amino:             legacyAmino,
	}
}