package types

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

const (
	// HodlAddressPrefix is the prefix for all ShareHODL addresses
	HodlAddressPrefix = "Hodl"
	
	// HodlAddressLength is the total length of a Hodl address (4 + 40 = 44 chars)
	HodlAddressLength = 44
	
	// HodlAddressHexLength is the length of the hex part (40 chars)
	HodlAddressHexLength = 40
)

// HodlAddress represents a ShareHODL blockchain address
type HodlAddress string

// NewHodlAddress creates a new random Hodl address
func NewHodlAddress() (HodlAddress, error) {
	// Generate 20 random bytes (40 hex chars)
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// Convert to hex and add Hodl prefix
	hexStr := hex.EncodeToString(bytes)
	address := HodlAddressPrefix + hexStr
	
	return HodlAddress(address), nil
}

// ParseHodlAddress creates a HodlAddress from a string and validates it
func ParseHodlAddress(addr string) (HodlAddress, error) {
	if err := ValidateHodlAddress(addr); err != nil {
		return "", err
	}
	return HodlAddress(addr), nil
}

// ValidateHodlAddress validates a Hodl address format
func ValidateHodlAddress(addr string) error {
	if len(addr) != HodlAddressLength {
		return fmt.Errorf("invalid address length: expected %d, got %d", HodlAddressLength, len(addr))
	}
	
	if !strings.HasPrefix(addr, HodlAddressPrefix) {
		return fmt.Errorf("invalid address prefix: expected %s", HodlAddressPrefix)
	}
	
	// Check if the hex part contains only valid hex characters
	hexPart := addr[len(HodlAddressPrefix):]
	if len(hexPart) != HodlAddressHexLength {
		return fmt.Errorf("invalid hex part length: expected %d, got %d", HodlAddressHexLength, len(hexPart))
	}
	
	// Validate hex characters (0-9, a-f, A-F)
	hexRegex := regexp.MustCompile("^[0-9a-fA-F]+$")
	if !hexRegex.MatchString(hexPart) {
		return fmt.Errorf("invalid hex characters in address")
	}
	
	return nil
}

// String returns the string representation of the address
func (h HodlAddress) String() string {
	return string(h)
}

// Bytes returns the raw bytes of the address (without prefix)
func (h HodlAddress) Bytes() ([]byte, error) {
	if err := ValidateHodlAddress(string(h)); err != nil {
		return nil, err
	}
	
	hexPart := string(h)[len(HodlAddressPrefix):]
	return hex.DecodeString(hexPart)
}

// ToLower returns the address in lowercase
func (h HodlAddress) ToLower() HodlAddress {
	return HodlAddress(strings.ToLower(string(h)))
}

// Equals checks if two addresses are equal (case-insensitive)
func (h HodlAddress) Equals(other HodlAddress) bool {
	return strings.EqualFold(string(h), string(other))
}

// IsEmpty checks if the address is empty
func (h HodlAddress) IsEmpty() bool {
	return len(string(h)) == 0
}

// HodlAddressFromBytes creates a Hodl address from raw bytes
func HodlAddressFromBytes(bytes []byte) (HodlAddress, error) {
	if len(bytes) != 20 {
		return "", fmt.Errorf("invalid byte length: expected 20, got %d", len(bytes))
	}
	
	hexStr := hex.EncodeToString(bytes)
	address := HodlAddressPrefix + hexStr
	
	return HodlAddress(address), nil
}

// IsValidHodlAddress is a convenience function to check if a string is a valid Hodl address
func IsValidHodlAddress(addr string) bool {
	return ValidateHodlAddress(addr) == nil
}