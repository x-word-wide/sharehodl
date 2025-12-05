/**
 * ShareHODL Hodl Address Utilities
 * 
 * Provides utilities for handling ShareHODL blockchain addresses
 * with the format: Hodl + 40 hex characters
 * Example: Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad
 */

export const HODL_ADDRESS_PREFIX = 'Hodl';
export const HODL_ADDRESS_LENGTH = 44;
export const HODL_ADDRESS_HEX_LENGTH = 40;

/**
 * Validates a Hodl address format
 */
export function validateHodlAddress(address: string): boolean {
  try {
    // Check length
    if (address.length !== HODL_ADDRESS_LENGTH) {
      return false;
    }
    
    // Check prefix
    if (!address.startsWith(HODL_ADDRESS_PREFIX)) {
      return false;
    }
    
    // Check hex part
    const hexPart = address.substring(HODL_ADDRESS_PREFIX.length);
    if (hexPart.length !== HODL_ADDRESS_HEX_LENGTH) {
      return false;
    }
    
    // Validate hex characters (case insensitive)
    const hexRegex = /^[0-9a-fA-F]+$/;
    return hexRegex.test(hexPart);
  } catch {
    return false;
  }
}

/**
 * Generates a random Hodl address (for testing/demo purposes)
 */
export function generateRandomHodlAddress(): string {
  const chars = '0123456789abcdef';
  let hexPart = '';
  
  for (let i = 0; i < HODL_ADDRESS_HEX_LENGTH; i++) {
    hexPart += chars[Math.floor(Math.random() * chars.length)];
  }
  
  return HODL_ADDRESS_PREFIX + hexPart;
}

/**
 * Formats a Hodl address for display (adds spacing for readability)
 */
export function formatHodlAddress(address: string, compact = false): string {
  if (!validateHodlAddress(address)) {
    return address; // Return as-is if invalid
  }
  
  if (compact) {
    // Show first 8 and last 8 characters with ellipsis
    return `${address.substring(0, 8)}...${address.substring(address.length - 8)}`;
  }
  
  // Add spaces for better readability: Hodl 46d0 7236 46bc c9eb 6bf1 f382 871c 8b0f c321 54ad
  const prefix = address.substring(0, 4); // "Hodl"
  const hex = address.substring(4);
  const chunks = hex.match(/.{1,4}/g) || [];
  
  return `${prefix} ${chunks.join(' ')}`;
}

/**
 * Normalizes a Hodl address (converts to lowercase)
 */
export function normalizeHodlAddress(address: string): string {
  if (!validateHodlAddress(address)) {
    return address;
  }
  
  return address.toLowerCase();
}

/**
 * Compares two Hodl addresses (case insensitive)
 */
export function compareHodlAddresses(addr1: string, addr2: string): boolean {
  if (!validateHodlAddress(addr1) || !validateHodlAddress(addr2)) {
    return false;
  }
  
  return normalizeHodlAddress(addr1) === normalizeHodlAddress(addr2);
}

/**
 * Extracts the hex part from a Hodl address
 */
export function getHodlAddressHex(address: string): string | null {
  if (!validateHodlAddress(address)) {
    return null;
  }
  
  return address.substring(HODL_ADDRESS_PREFIX.length);
}

/**
 * Creates a Hodl address from hex string
 */
export function createHodlAddressFromHex(hex: string): string | null {
  // Validate hex input
  if (hex.length !== HODL_ADDRESS_HEX_LENGTH) {
    return null;
  }
  
  const hexRegex = /^[0-9a-fA-F]+$/;
  if (!hexRegex.test(hex)) {
    return null;
  }
  
  return HODL_ADDRESS_PREFIX + hex.toLowerCase();
}

/**
 * Type guard for Hodl addresses
 */
export function isHodlAddress(value: unknown): value is string {
  return typeof value === 'string' && validateHodlAddress(value);
}

/**
 * Sample Hodl addresses for testing/demo
 */
export const SAMPLE_HODL_ADDRESSES = [
  'Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad',
  'HodlA1B2c3D4e5F6789012345678901234567890aBcD',
  'Hodlff1234567890abcdef1234567890abcdef123456',
  'Hodl0123456789abcdef0123456789abcdef01234567',
  'Hodldeadbeefcafebabe1337133713371337deadbeef'
];

/**
 * Error messages for validation
 */
export const HODL_ADDRESS_ERRORS = {
  INVALID_LENGTH: `Address must be exactly ${HODL_ADDRESS_LENGTH} characters long`,
  INVALID_PREFIX: `Address must start with "${HODL_ADDRESS_PREFIX}"`,
  INVALID_HEX: 'Address contains invalid hexadecimal characters',
  EMPTY_ADDRESS: 'Address cannot be empty'
};

/**
 * Validates a Hodl address and returns detailed error message
 */
export function validateHodlAddressWithError(address: string): { valid: boolean; error?: string } {
  if (!address || address.length === 0) {
    return { valid: false, error: HODL_ADDRESS_ERRORS.EMPTY_ADDRESS };
  }
  
  if (address.length !== HODL_ADDRESS_LENGTH) {
    return { valid: false, error: HODL_ADDRESS_ERRORS.INVALID_LENGTH };
  }
  
  if (!address.startsWith(HODL_ADDRESS_PREFIX)) {
    return { valid: false, error: HODL_ADDRESS_ERRORS.INVALID_PREFIX };
  }
  
  const hexPart = address.substring(HODL_ADDRESS_PREFIX.length);
  const hexRegex = /^[0-9a-fA-F]+$/;
  if (!hexRegex.test(hexPart)) {
    return { valid: false, error: HODL_ADDRESS_ERRORS.INVALID_HEX };
  }
  
  return { valid: true };
}