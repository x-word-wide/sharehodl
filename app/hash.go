package app

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// HashCodec provides encoding/decoding for transaction and block hashes
// using bech32 format with custom prefixes.
type HashCodec struct {
	prefix string
}

// NewTxHashCodec creates a codec for transaction hashes with "sharetx" prefix
func NewTxHashCodec() *HashCodec {
	return &HashCodec{prefix: Bech32PrefixTxHash}
}

// NewBlockHashCodec creates a codec for block hashes with "shareblock" prefix
func NewBlockHashCodec() *HashCodec {
	return &HashCodec{prefix: Bech32PrefixBlockHash}
}

// Encode converts a hex hash string to bech32 format
// Input: "4D6DBF98A9D9F9F22843CFFDE4200F6C79D1953CB226E8735DC4A550547C44B9"
// Output: "sharetx1f4kmlv92e8uncggsl0a7yqp0dvjawn2uvjxlp6thv99g2glycjun3ky"
func (c *HashCodec) Encode(hexHash string) (string, error) {
	// Remove any "0x" prefix if present
	hexHash = strings.TrimPrefix(strings.ToLower(hexHash), "0x")

	// Decode hex string to bytes
	hashBytes, err := hex.DecodeString(hexHash)
	if err != nil {
		return "", fmt.Errorf("invalid hex hash: %w", err)
	}

	// Encode to bech32
	bech32Hash, err := bech32.ConvertAndEncode(c.prefix, hashBytes)
	if err != nil {
		return "", fmt.Errorf("bech32 encoding failed: %w", err)
	}

	return bech32Hash, nil
}

// Decode converts a bech32 hash back to hex format
// Input: "sharetx1f4kmlv92e8uncggsl0a7yqp0dvjawn2uvjxlp6thv99g2glycjun3ky"
// Output: "4D6DBF98A9D9F9F22843CFFDE4200F6C79D1953CB226E8735DC4A550547C44B9"
func (c *HashCodec) Decode(bech32Hash string) (string, error) {
	prefix, hashBytes, err := bech32.DecodeAndConvert(bech32Hash)
	if err != nil {
		return "", fmt.Errorf("bech32 decoding failed: %w", err)
	}

	if prefix != c.prefix {
		return "", fmt.Errorf("invalid prefix: expected %s, got %s", c.prefix, prefix)
	}

	return strings.ToUpper(hex.EncodeToString(hashBytes)), nil
}

// MustEncode encodes a hex hash to bech32, panicking on error
func (c *HashCodec) MustEncode(hexHash string) string {
	result, err := c.Encode(hexHash)
	if err != nil {
		panic(err)
	}
	return result
}

// Global codec instances for convenience
var (
	TxHashCodec    = NewTxHashCodec()
	BlockHashCodec = NewBlockHashCodec()
)

// EncodeTxHash converts a hex transaction hash to bech32 format with "sharetx" prefix
func EncodeTxHash(hexHash string) (string, error) {
	return TxHashCodec.Encode(hexHash)
}

// DecodeTxHash converts a bech32 transaction hash back to hex format
func DecodeTxHash(bech32Hash string) (string, error) {
	return TxHashCodec.Decode(bech32Hash)
}

// EncodeBlockHash converts a hex block hash to bech32 format with "shareblock" prefix
func EncodeBlockHash(hexHash string) (string, error) {
	return BlockHashCodec.Encode(hexHash)
}

// DecodeBlockHash converts a bech32 block hash back to hex format
func DecodeBlockHash(bech32Hash string) (string, error) {
	return BlockHashCodec.Decode(bech32Hash)
}
