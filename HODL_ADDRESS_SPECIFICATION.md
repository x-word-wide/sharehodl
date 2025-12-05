# ShareHODL Hodl Address Specification

## Overview

ShareHODL blockchain uses a custom address format called "Hodl Addresses" to provide a distinctive and wallet-friendly addressing system. This specification defines the format, validation rules, and implementation guidelines for Hodl addresses.

## Address Format

### Structure
```
Hodl + 40 hexadecimal characters
```

### Examples
```
Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad
HodlA1B2c3D4e5F6789012345678901234567890aBcD
Hodlff1234567890abcdef1234567890abcdef123456
```

### Specifications

| Property | Value |
|----------|-------|
| **Total Length** | 44 characters |
| **Prefix** | `Hodl` (case-sensitive) |
| **Hex Portion Length** | 40 characters |
| **Encoding** | Hexadecimal (0-9, a-f, A-F) |
| **Case Sensitivity** | Case-insensitive for comparison |

## Technical Details

### Address Generation

1. **Random Generation**: Generate 20 random bytes (160 bits)
2. **Hex Encoding**: Convert bytes to 40-character hexadecimal string
3. **Prefix Addition**: Prepend `Hodl` to create final address
4. **Result**: 44-character Hodl address

```go
// Go implementation
func NewHodlAddress() (HodlAddress, error) {
    bytes := make([]byte, 20)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    hexStr := hex.EncodeToString(bytes)
    return HodlAddress("Hodl" + hexStr), nil
}
```

```typescript
// TypeScript implementation
function generateRandomHodlAddress(): string {
    const chars = '0123456789abcdef';
    let hexPart = '';
    
    for (let i = 0; i < 40; i++) {
        hexPart += chars[Math.floor(Math.random() * chars.length)];
    }
    
    return 'Hodl' + hexPart;
}
```

### Validation Rules

1. **Length Check**: Must be exactly 44 characters
2. **Prefix Check**: Must start with `Hodl`
3. **Hex Validation**: Characters 5-44 must be valid hexadecimal
4. **Character Set**: `[0-9a-fA-F]` for hex portion

```go
// Go validation
func ValidateHodlAddress(addr string) error {
    if len(addr) != 44 {
        return fmt.Errorf("invalid length: expected 44, got %d", len(addr))
    }
    
    if !strings.HasPrefix(addr, "Hodl") {
        return fmt.Errorf("invalid prefix: expected Hodl")
    }
    
    hexPart := addr[4:]
    if matched, _ := regexp.MatchString("^[0-9a-fA-F]+$", hexPart); !matched {
        return fmt.Errorf("invalid hex characters")
    }
    
    return nil
}
```

```typescript
// TypeScript validation
function validateHodlAddress(address: string): boolean {
    if (address.length !== 44) return false;
    if (!address.startsWith('Hodl')) return false;
    
    const hexPart = address.substring(4);
    const hexRegex = /^[0-9a-fA-F]+$/;
    return hexRegex.test(hexPart);
}
```

## Wallet Integration

### Trust Wallet Compatibility

Hodl addresses are designed to work seamlessly with Trust Wallet and other multi-chain wallets:

- **Format Similarity**: Uses familiar hex-based addressing like Ethereum
- **Length Consistency**: 44 characters total (similar to other blockchain addresses)
- **Prefix Recognition**: `Hodl` prefix clearly identifies ShareHODL network
- **Standard Hex**: 40-character hex portion follows established patterns

### Integration Guidelines

For wallet developers integrating ShareHODL support:

1. **Address Recognition**: Detect addresses starting with `Hodl`
2. **Validation**: Implement the validation rules specified above
3. **Display Format**: Show full address or use truncation (e.g., `Hodl46d0...54ad`)
4. **QR Codes**: Use full 44-character address in QR code generation
5. **Checksums**: Optional checksum validation (implementation dependent)

## Implementation Examples

### Backend Integration (Go)

```go
package types

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "regexp"
    "strings"
)

type HodlAddress string

const (
    HodlAddressPrefix = "Hodl"
    HodlAddressLength = 44
    HodlAddressHexLength = 40
)

func NewHodlAddress() (HodlAddress, error) {
    bytes := make([]byte, 20)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("failed to generate random bytes: %w", err)
    }
    
    hexStr := hex.EncodeToString(bytes)
    address := HodlAddressPrefix + hexStr
    
    return HodlAddress(address), nil
}

func (h HodlAddress) String() string {
    return string(h)
}

func (h HodlAddress) Bytes() ([]byte, error) {
    if err := ValidateHodlAddress(string(h)); err != nil {
        return nil, err
    }
    
    hexPart := string(h)[len(HodlAddressPrefix):]
    return hex.DecodeString(hexPart)
}
```

### Frontend Integration (TypeScript/JavaScript)

```typescript
export class HodlAddress {
    private readonly address: string;

    constructor(address: string) {
        if (!this.validate(address)) {
            throw new Error('Invalid Hodl address format');
        }
        this.address = address.toLowerCase();
    }

    static generate(): HodlAddress {
        const chars = '0123456789abcdef';
        let hexPart = '';
        
        for (let i = 0; i < 40; i++) {
            hexPart += chars[Math.floor(Math.random() * chars.length)];
        }
        
        return new HodlAddress('Hodl' + hexPart);
    }

    toString(): string {
        return this.address;
    }

    toDisplayString(compact = false): string {
        if (compact) {
            return `${this.address.substring(0, 8)}...${this.address.substring(-8)}`;
        }
        return this.address;
    }

    private validate(address: string): boolean {
        if (address.length !== 44) return false;
        if (!address.startsWith('Hodl')) return false;
        
        const hexPart = address.substring(4);
        const hexRegex = /^[0-9a-fA-F]+$/;
        return hexRegex.test(hexPart);
    }
}
```

### React Component Example

```tsx
import React from 'react';
import { formatHodlAddress, validateHodlAddress } from '@repo/ui';

interface AddressDisplayProps {
    address: string;
    compact?: boolean;
}

export const AddressDisplay: React.FC<AddressDisplayProps> = ({ 
    address, 
    compact = false 
}) => {
    const isValid = validateHodlAddress(address);
    const displayAddress = isValid ? formatHodlAddress(address, compact) : address;
    
    return (
        <span className={`font-mono ${isValid ? 'text-green-600' : 'text-red-600'}`}>
            {displayAddress}
        </span>
    );
};
```

## Security Considerations

### Address Generation Security

1. **Random Source**: Use cryptographically secure random number generation
2. **Entropy**: Ensure 160 bits of entropy (20 random bytes)
3. **Collision Resistance**: 2^160 possible addresses provide collision resistance
4. **Private Key Derivation**: Follow established cryptographic standards

### Validation Security

1. **Input Sanitization**: Always validate address format before processing
2. **Case Handling**: Normalize addresses for comparison (case-insensitive)
3. **Length Checks**: Prevent buffer overflow attacks with strict length validation
4. **Character Validation**: Only accept valid hexadecimal characters

## Migration from Cosmos Addresses

### Transition Strategy

For existing ShareHODL networks using Cosmos-style addresses (`sharehodl1...`):

1. **Dual Support Period**: Support both formats during transition
2. **Address Mapping**: Maintain mapping between old and new formats
3. **Client Updates**: Update wallets and applications to support Hodl addresses
4. **Documentation Updates**: Update all documentation and examples

### Migration Tools

```go
// Convert existing Cosmos address to Hodl format
func CosmosToHodlAddress(cosmosAddr string) (HodlAddress, error) {
    // Extract the bech32 decoded bytes
    _, bz, err := bech32.DecodeAndConvert(cosmosAddr)
    if err != nil {
        return "", err
    }
    
    // Create Hodl address from same bytes
    return HodlAddressFromBytes(bz)
}
```

## Testing and Quality Assurance

### Test Vectors

```
Valid Addresses:
- Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad
- HodlA1B2c3D4e5F6789012345678901234567890aBcD
- Hodlff1234567890abcdef1234567890abcdef123456
- Hodl0000000000000000000000000000000000000000
- HodlFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF

Invalid Addresses:
- Hold46d0723646bcc9eb6bf1f382871c8b0fc32154ad  (wrong prefix)
- Hodl46d0723646bcc9eb6bf1f382871c8b0fc3215      (too short)
- Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154adee (too long)
- Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154gz   (invalid chars)
- hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad   (wrong case prefix)
```

### Unit Test Examples

```go
func TestHodlAddressValidation(t *testing.T) {
    tests := []struct{
        address string
        valid   bool
    }{
        {"Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad", true},
        {"Hold46d0723646bcc9eb6bf1f382871c8b0fc32154ad", false},
        {"Hodl46d072", false},
        {"", false},
    }
    
    for _, test := range tests {
        err := ValidateHodlAddress(test.address)
        if test.valid && err != nil {
            t.Errorf("Expected valid, got error: %v", err)
        }
        if !test.valid && err == nil {
            t.Errorf("Expected invalid, got no error")
        }
    }
}
```

## Future Considerations

### Potential Enhancements

1. **Checksums**: Optional checksum validation for error detection
2. **Address Types**: Different prefixes for different address types
3. **Vanity Addresses**: Tools for generating custom address patterns
4. **Hardware Wallet Support**: Integration with hardware wallets

### Backwards Compatibility

- Maintain support for legacy address formats during transition periods
- Provide conversion utilities between address formats
- Ensure API compatibility across address format changes

## Conclusion

The Hodl address format provides ShareHODL with a distinctive, wallet-friendly addressing system that maintains compatibility with existing infrastructure while establishing a unique brand identity. The format balances security, usability, and technical compatibility to create an optimal user experience across all ShareHODL applications and integrations.

For questions or contributions to this specification, please refer to the ShareHODL documentation or submit issues through the appropriate channels.