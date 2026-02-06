/**
 * Secure Client-Side Cryptographic Operations
 *
 * SECURITY: All operations happen in the browser. Private keys NEVER leave the device.
 */

import * as bip39 from 'bip39';
import { ethers } from 'ethers';
import { Chain } from '../types';
import { logger } from './logger';

// ============================================
// Mnemonic Generation & Validation
// ============================================

/**
 * Generate a new BIP39 mnemonic phrase
 * Uses Web Crypto API for secure randomness
 */
export function generateMnemonic(strength: 128 | 256 = 256): string {
  // 256 bits = 24 words, 128 bits = 12 words
  return bip39.generateMnemonic(strength);
}

/**
 * Validate a mnemonic phrase
 */
export function validateMnemonic(mnemonic: string): boolean {
  return bip39.validateMnemonic(mnemonic);
}

/**
 * Convert mnemonic to seed
 */
export async function mnemonicToSeed(mnemonic: string, password = ''): Promise<Uint8Array> {
  const seed = await bip39.mnemonicToSeed(mnemonic, password);
  return new Uint8Array(seed);
}

// ============================================
// Key Derivation (BIP44)
// ============================================

/**
 * Derive wallet from mnemonic for a specific chain
 */
export function deriveWallet(mnemonic: string, coinType: number, accountIndex = 0): ethers.HDNodeWallet {
  const path = `m/44'/${coinType}'/${accountIndex}'/0/0`;
  return ethers.HDNodeWallet.fromPhrase(mnemonic, undefined, path);
}

/**
 * Get address for a specific chain from mnemonic
 */
export function getAddressFromMnemonic(mnemonic: string, chain: Chain): string {
  // For EVM chains, use ethers directly
  if (isEvmChain(chain)) {
    const wallet = deriveWallet(mnemonic, 60);
    return wallet.address;
  }

  // For Cosmos-based chains (including ShareHODL)
  if (isCosmosChain(chain)) {
    const wallet = deriveWallet(mnemonic, 118);
    const pubKeyBytes = ethers.getBytes(wallet.publicKey);
    return bech32Address(pubKeyBytes, getPrefix(chain));
  }

  // For Bitcoin (Native SegWit P2WPKH - bech32)
  if (chain === Chain.BITCOIN) {
    // Use BIP84 path for native segwit (m/84'/0'/0'/0/0)
    const path = `m/84'/0'/0'/0/0`;
    const wallet = ethers.HDNodeWallet.fromPhrase(mnemonic, undefined, path);

    // Get compressed public key (33 bytes)
    const pubKeyHex = wallet.publicKey;
    const pubKeyBytes = ethers.getBytes(pubKeyHex);

    // Hash160 = RIPEMD160(SHA256(pubkey))
    const sha256Hash = ethers.sha256(pubKeyBytes);
    const hash160 = ethers.ripemd160(sha256Hash);
    const hash160Bytes = ethers.getBytes(hash160);

    // P2WPKH witness program is just the 20-byte hash
    return bech32Encode('bc', convertToWitnessProgram(hash160Bytes));
  }

  // For Solana - Note: Proper Solana derivation requires Ed25519
  // This is a placeholder until Ed25519 library is added
  if (chain === Chain.SOLANA) {
    // Solana uses Ed25519, not secp256k1
    // For now, derive a deterministic identifier from the mnemonic
    // In production, use @solana/web3.js or tweetnacl for proper Ed25519 derivation
    const seedPhrase = mnemonic + '-solana-v1';
    const hash = ethers.keccak256(ethers.toUtf8Bytes(seedPhrase));
    // Return base58-like encoding (32 bytes as hex for now)
    return ethers.encodeBase58(ethers.getBytes(hash));
  }

  return '';
}

function isEvmChain(chain: Chain): boolean {
  return [
    Chain.ETHEREUM,
    Chain.POLYGON,
    Chain.ARBITRUM,
    Chain.OPTIMISM,
    Chain.BASE,
    Chain.AVALANCHE,
    Chain.BNB
  ].includes(chain);
}

function isCosmosChain(chain: Chain): boolean {
  return [
    Chain.SHAREHODL,
    Chain.COSMOS,
    Chain.OSMOSIS,
    Chain.CELESTIA
  ].includes(chain);
}

function getPrefix(chain: Chain): string {
  const prefixes: Partial<Record<Chain, string>> = {
    [Chain.SHAREHODL]: 'sharehodl',
    [Chain.COSMOS]: 'cosmos',
    [Chain.OSMOSIS]: 'osmo',
    [Chain.CELESTIA]: 'celestia'
  };
  return prefixes[chain] || 'cosmos';
}

// ============================================
// Address Generation
// ============================================

/**
 * Generate bech32 address for Cosmos chains
 */
function bech32Address(pubKey: Uint8Array, prefix: string): string {
  // SHA256 then RIPEMD160
  const sha256Hash = ethers.sha256(pubKey);
  const hash = ethers.ripemd160(sha256Hash);
  const hashBytes = ethers.getBytes(hash);

  // Convert to bech32
  return bech32Encode(prefix, hashBytes);
}

/**
 * Simplified bech32 encoding for Cosmos addresses
 */
function bech32Encode(prefix: string, data: Uint8Array): string {
  const CHARSET = 'qpzry9x8gf2tvdw0s3jn54khce6mua7l';

  // For Bitcoin segwit, first byte is witness version
  const isSegwit = prefix === 'bc' || prefix === 'tb';

  let converted: number[];
  if (isSegwit && data.length === 21) {
    // Segwit: version byte + converted program
    const version = data[0];
    const program = data.slice(1);
    converted = [version, ...convertBits(program, 8, 5, true)];
  } else {
    // Standard bech32: just convert the data
    converted = convertBits(data, 8, 5, true);
  }

  // Create checksum
  let result = prefix + '1';
  for (const byte of converted) {
    result += CHARSET[byte];
  }

  // Add checksum (use bech32m for segwit v1+, bech32 for v0)
  const checksum = isSegwit && data[0] > 0
    ? createChecksumBech32m(prefix, converted)
    : createChecksum(prefix, converted);
  for (const c of checksum) {
    result += CHARSET[c];
  }

  return result;
}

/**
 * Create bech32m checksum (for segwit v1+)
 */
function createChecksumBech32m(prefix: string, data: number[]): number[] {
  const BECH32M_CONST = 0x2bc830a3;
  const values = expandPrefix(prefix).concat(data).concat([0, 0, 0, 0, 0, 0]);
  const polymod = bech32Polymod(values) ^ BECH32M_CONST;
  const checksum: number[] = [];
  for (let i = 0; i < 6; i++) {
    checksum.push((polymod >> (5 * (5 - i))) & 31);
  }
  return checksum;
}

function convertBits(data: Uint8Array, fromBits: number, toBits: number, pad: boolean): number[] {
  let acc = 0;
  let bits = 0;
  const result: number[] = [];
  const maxv = (1 << toBits) - 1;

  for (const value of data) {
    acc = (acc << fromBits) | value;
    bits += fromBits;
    while (bits >= toBits) {
      bits -= toBits;
      result.push((acc >> bits) & maxv);
    }
  }

  if (pad && bits > 0) {
    result.push((acc << (toBits - bits)) & maxv);
  }

  return result;
}

function createChecksum(prefix: string, data: number[]): number[] {
  const values = expandPrefix(prefix).concat(data).concat([0, 0, 0, 0, 0, 0]);
  const polymod = bech32Polymod(values) ^ 1;
  const checksum: number[] = [];
  for (let i = 0; i < 6; i++) {
    checksum.push((polymod >> (5 * (5 - i))) & 31);
  }
  return checksum;
}

function expandPrefix(prefix: string): number[] {
  const result: number[] = [];
  for (const c of prefix) {
    result.push(c.charCodeAt(0) >> 5);
  }
  result.push(0);
  for (const c of prefix) {
    result.push(c.charCodeAt(0) & 31);
  }
  return result;
}

function bech32Polymod(values: number[]): number {
  const GEN = [0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3];
  let chk = 1;
  for (const v of values) {
    const top = chk >> 25;
    chk = ((chk & 0x1ffffff) << 5) ^ v;
    for (let i = 0; i < 5; i++) {
      if ((top >> i) & 1) {
        chk ^= GEN[i];
      }
    }
  }
  return chk;
}

/**
 * Convert hash160 to witness program (version 0 P2WPKH)
 */
function convertToWitnessProgram(hash160: Uint8Array): Uint8Array {
  // Witness version 0 + 20 byte hash
  const program = new Uint8Array(21);
  program[0] = 0; // version 0
  program.set(hash160, 1);
  return program;
}

// ============================================
// Encryption for Local Storage
// ============================================

/**
 * Check if Web Crypto API is available (requires secure context)
 */
function isWebCryptoAvailable(): boolean {
  return typeof crypto !== 'undefined' &&
         typeof crypto.subtle !== 'undefined' &&
         typeof crypto.subtle.importKey === 'function';
}

/**
 * Encrypt data with AES-GCM using a PIN-derived key
 */
export async function encryptData(data: string, pin: string): Promise<string> {
  // Check for Web Crypto API availability
  if (!isWebCryptoAvailable()) {
    throw new Error('Secure encryption not available. Please ensure you are using HTTPS.');
  }

  const encoder = new TextEncoder();
  const dataBuffer = encoder.encode(data);

  try {
    // Derive key from PIN
    const keyMaterial = await crypto.subtle.importKey(
      'raw',
      encoder.encode(pin),
      'PBKDF2',
      false,
      ['deriveKey']
    );

    const salt = crypto.getRandomValues(new Uint8Array(16));
    const key = await crypto.subtle.deriveKey(
      {
        name: 'PBKDF2',
        salt,
        iterations: 600000, // OWASP recommended minimum for PBKDF2-SHA256
        hash: 'SHA-256'
      },
      keyMaterial,
      { name: 'AES-GCM', length: 256 },
      false,
      ['encrypt']
    );

    const iv = crypto.getRandomValues(new Uint8Array(12));
    const encrypted = await crypto.subtle.encrypt(
      { name: 'AES-GCM', iv },
      key,
      dataBuffer
    );

    // Combine salt + iv + encrypted data
    const result = new Uint8Array(salt.length + iv.length + encrypted.byteLength);
    result.set(salt, 0);
    result.set(iv, salt.length);
    result.set(new Uint8Array(encrypted), salt.length + iv.length);

    return btoa(String.fromCharCode(...result));
  } catch (error) {
    // SECURITY: Don't log sensitive details in production
    logger.error('Encryption operation failed');
    throw new Error('Failed to encrypt wallet data. Please try again.');
  }
}

/**
 * Decrypt data with AES-GCM
 */
export async function decryptData(encryptedData: string, pin: string): Promise<string> {
  // Check for Web Crypto API availability
  if (!isWebCryptoAvailable()) {
    throw new Error('Secure decryption not available. Please ensure you are using HTTPS.');
  }

  const encoder = new TextEncoder();
  const decoder = new TextDecoder();

  try {
    const data = Uint8Array.from(atob(encryptedData), c => c.charCodeAt(0));

    const salt = data.slice(0, 16);
    const iv = data.slice(16, 28);
    const encrypted = data.slice(28);

    const keyMaterial = await crypto.subtle.importKey(
      'raw',
      encoder.encode(pin),
      'PBKDF2',
      false,
      ['deriveKey']
    );

    const key = await crypto.subtle.deriveKey(
      {
        name: 'PBKDF2',
        salt,
        iterations: 600000, // OWASP recommended minimum for PBKDF2-SHA256
        hash: 'SHA-256'
      },
      keyMaterial,
      { name: 'AES-GCM', length: 256 },
      false,
      ['decrypt']
    );

    const decrypted = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv },
      key,
      encrypted
    );

    return decoder.decode(decrypted);
  } catch (error) {
    // Don't log specific error to avoid leaking info
    throw new Error('Decryption failed');
  }
}
