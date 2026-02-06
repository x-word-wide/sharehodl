package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/spf13/cobra"
)

// Bech32 prefixes for hash encoding (duplicated from app to avoid circular imports)
const (
	Bech32PrefixTxHash    = "sharetx"
	Bech32PrefixBlockHash = "shareblock"
)

// hashCommand creates the hash utility command
func hashCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hash",
		Short: "Hash encoding/decoding utilities",
		Long:  `Utilities for converting between hex and bech32 hash formats.`,
	}

	cmd.AddCommand(
		encodeTxHashCmd(),
		encodeBlockHashCmd(),
		decodeTxHashCmd(),
		decodeBlockHashCmd(),
	)

	return cmd
}

// encodeTxHashCmd creates the tx hash encoding command
func encodeTxHashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "encode-tx [hex-hash]",
		Short: "Encode a hex transaction hash to bech32 format",
		Long:  `Converts a hex transaction hash to bech32 format with 'sharetx' prefix.`,
		Example: `sharehodld hash encode-tx DF92CF72DE1CBD587E8AD933567D4912C51850E5B39276ADE5A9B1C961C136B6
# Output: sharetx1m7fv7uk7rj74sl52mye4vl2fztz3s589kwf8dt094xcujcwpx6mqjtu8gv`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hexHash := strings.TrimPrefix(strings.ToLower(args[0]), "0x")
			hashBytes, err := hex.DecodeString(hexHash)
			if err != nil {
				return fmt.Errorf("invalid hex hash: %w", err)
			}

			bech32Hash, err := bech32.ConvertAndEncode(Bech32PrefixTxHash, hashBytes)
			if err != nil {
				return fmt.Errorf("bech32 encoding failed: %w", err)
			}

			fmt.Println(bech32Hash)
			return nil
		},
	}
}

// encodeBlockHashCmd creates the block hash encoding command
func encodeBlockHashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "encode-block [hex-hash]",
		Short: "Encode a hex block hash to bech32 format",
		Long:  `Converts a hex block hash to bech32 format with 'shareblock' prefix.`,
		Example: `sharehodld hash encode-block 4D6DBF98A9D9F9F22843CFFDE4200F6C79D1953CB226E8735DC4A550547C44B9
# Output: shareblock1f4kmlx9fm8uly2zrel77ggq0d3uar9fukgnwsu6acjj4q4rugjus8gmjck`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hexHash := strings.TrimPrefix(strings.ToLower(args[0]), "0x")
			hashBytes, err := hex.DecodeString(hexHash)
			if err != nil {
				return fmt.Errorf("invalid hex hash: %w", err)
			}

			bech32Hash, err := bech32.ConvertAndEncode(Bech32PrefixBlockHash, hashBytes)
			if err != nil {
				return fmt.Errorf("bech32 encoding failed: %w", err)
			}

			fmt.Println(bech32Hash)
			return nil
		},
	}
}

// decodeTxHashCmd creates the tx hash decoding command
func decodeTxHashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decode-tx [bech32-hash]",
		Short: "Decode a bech32 transaction hash to hex format",
		Long:  `Converts a bech32 transaction hash (sharetx...) back to hex format.`,
		Example: `sharehodld hash decode-tx sharetx1m7fv7uk7rj74sl52mye4vl2fztz3s589kwf8dt094xcujcwpx6mqjtu8gv
# Output: DF92CF72DE1CBD587E8AD933567D4912C51850E5B39276ADE5A9B1C961C136B6`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix, hashBytes, err := bech32.DecodeAndConvert(args[0])
			if err != nil {
				return fmt.Errorf("bech32 decoding failed: %w", err)
			}

			if prefix != Bech32PrefixTxHash {
				return fmt.Errorf("invalid prefix: expected %s, got %s", Bech32PrefixTxHash, prefix)
			}

			fmt.Println(strings.ToUpper(hex.EncodeToString(hashBytes)))
			return nil
		},
	}
}

// decodeBlockHashCmd creates the block hash decoding command
func decodeBlockHashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decode-block [bech32-hash]",
		Short: "Decode a bech32 block hash to hex format",
		Long:  `Converts a bech32 block hash (shareblock...) back to hex format.`,
		Example: `sharehodld hash decode-block shareblock1f4kmlx9fm8uly2zrel77ggq0d3uar9fukgnwsu6acjj4q4rugjus8gmjck
# Output: 4D6DBF98A9D9F9F22843CFFDE4200F6C79D1953CB226E8735DC4A550547C44B9`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix, hashBytes, err := bech32.DecodeAndConvert(args[0])
			if err != nil {
				return fmt.Errorf("bech32 decoding failed: %w", err)
			}

			if prefix != Bech32PrefixBlockHash {
				return fmt.Errorf("invalid prefix: expected %s, got %s", Bech32PrefixBlockHash, prefix)
			}

			fmt.Println(strings.ToUpper(hex.EncodeToString(hashBytes)))
			return nil
		},
	}
}
