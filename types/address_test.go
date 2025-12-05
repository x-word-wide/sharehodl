package types

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestNewHodlAddress(t *testing.T) {
	addr, err := NewHodlAddress()
	if err != nil {
		t.Fatalf("Failed to create new Hodl address: %v", err)
	}
	
	if err := ValidateHodlAddress(string(addr)); err != nil {
		t.Fatalf("Generated invalid address: %v", err)
	}
	
	if len(string(addr)) != HodlAddressLength {
		t.Fatalf("Invalid address length: expected %d, got %d", HodlAddressLength, len(string(addr)))
	}
	
	if !strings.HasPrefix(string(addr), HodlAddressPrefix) {
		t.Fatalf("Address doesn't have correct prefix: %s", string(addr))
	}
}

func TestValidateHodlAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		valid   bool
	}{
		{
			name:    "valid address lowercase",
			address: "Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad",
			valid:   true,
		},
		{
			name:    "valid address uppercase",
			address: "Hodl46D0723646BCC9EB6BF1F382871C8B0FC32154AD",
			valid:   true,
		},
		{
			name:    "valid address mixed case",
			address: "HodlA1B2c3D4e5F6789012345678901234567890aBcD",
			valid:   true,
		},
		{
			name:    "invalid prefix",
			address: "Hold46d0723646bcc9eb6bf1f382871c8b0fc32154ad",
			valid:   false,
		},
		{
			name:    "too short",
			address: "Hodl46d0723646bcc9eb6bf1f382871c8b0fc3215",
			valid:   false,
		},
		{
			name:    "too long",
			address: "Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154adee",
			valid:   false,
		},
		{
			name:    "invalid characters",
			address: "Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154gz",
			valid:   false,
		},
		{
			name:    "empty string",
			address: "",
			valid:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHodlAddress(tt.address)
			if tt.valid && err != nil {
				t.Errorf("Expected valid address, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid address, got no error")
			}
		})
	}
}

func TestParseHodlAddress(t *testing.T) {
	validAddr := "Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad"
	
	addr, err := ParseHodlAddress(validAddr)
	if err != nil {
		t.Fatalf("Failed to parse valid address: %v", err)
	}
	
	if string(addr) != validAddr {
		t.Fatalf("Parsed address doesn't match input: expected %s, got %s", validAddr, string(addr))
	}
}

func TestHodlAddressBytes(t *testing.T) {
	testHex := "46d0723646bcc9eb6bf1f382871c8b0fc32154ad"
	addr := HodlAddress(HodlAddressPrefix + testHex)
	
	bytes, err := addr.Bytes()
	if err != nil {
		t.Fatalf("Failed to get bytes from address: %v", err)
	}
	
	expectedBytes, _ := hex.DecodeString(testHex)
	if len(bytes) != len(expectedBytes) {
		t.Fatalf("Byte length mismatch: expected %d, got %d", len(expectedBytes), len(bytes))
	}
	
	for i, b := range bytes {
		if b != expectedBytes[i] {
			t.Fatalf("Byte mismatch at position %d: expected %x, got %x", i, expectedBytes[i], b)
		}
	}
}

func TestHodlAddressFromBytes(t *testing.T) {
	testBytes := []byte{0x46, 0xd0, 0x72, 0x36, 0x46, 0xbc, 0xc9, 0xeb, 0x6b, 0xf1, 0xf3, 0x82, 0x87, 0x1c, 0x8b, 0x0f, 0xc3, 0x21, 0x54, 0xad}
	
	addr, err := HodlAddressFromBytes(testBytes)
	if err != nil {
		t.Fatalf("Failed to create address from bytes: %v", err)
	}
	
	expected := "Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad"
	if string(addr) != expected {
		t.Fatalf("Address mismatch: expected %s, got %s", expected, string(addr))
	}
}

func TestHodlAddressEquals(t *testing.T) {
	addr1 := HodlAddress("Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad")
	addr2 := HodlAddress("Hodl46D0723646BCC9EB6BF1F382871C8B0FC32154AD")
	addr3 := HodlAddress("HodlA1B2c3D4e5F6789012345678901234567890aBcD")
	
	if !addr1.Equals(addr2) {
		t.Fatal("Case-insensitive comparison should be equal")
	}
	
	if addr1.Equals(addr3) {
		t.Fatal("Different addresses should not be equal")
	}
}

func TestHodlAddressToLower(t *testing.T) {
	addr := HodlAddress("Hodl46D0723646BCC9EB6BF1F382871C8B0FC32154AD")
	expected := HodlAddress("hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad")
	
	if addr.ToLower() != expected {
		t.Fatalf("ToLower failed: expected %s, got %s", expected, addr.ToLower())
	}
}

func TestIsValidHodlAddress(t *testing.T) {
	validAddr := "Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad"
	invalidAddr := "Hold46d0723646bcc9eb6bf1f382871c8b0fc32154ad"
	
	if !IsValidHodlAddress(validAddr) {
		t.Fatal("Valid address should return true")
	}
	
	if IsValidHodlAddress(invalidAddr) {
		t.Fatal("Invalid address should return false")
	}
}