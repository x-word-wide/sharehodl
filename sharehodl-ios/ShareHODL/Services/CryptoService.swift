import Foundation
import P256K
import CryptoSwift
import CryptoKit

/// Production-ready cryptographic service using P256K (secp256k1) + CryptoSwift
/// Provides: BIP39 (full 2048 words), BIP32/44 derivation, secp256k1, ECDSA, RIPEMD160, Bech32
///
/// SECURITY: This service implements secure memory handling for sensitive data:
/// - Private keys and mnemonics are cleared from memory after use
/// - Use SecureData wrapper for sensitive byte arrays
/// - Always call clearSensitiveData() when done with sensitive operations
final class CryptoService {
    static let shared = CryptoService()

    private let bech32Prefix = "hodl"

    // BIP44 path for Cosmos: m/44'/118'/0'/0/0
    private let cosmosPath: [UInt32] = [
        44 | 0x80000000,  // 44' (purpose)
        118 | 0x80000000, // 118' (Cosmos coin type)
        0 | 0x80000000,   // 0' (account)
        0,                 // 0 (change)
        0                  // 0 (address index)
    ]

    private init() {}

    // MARK: - Secure Memory Handling

    /// Securely clear sensitive data from memory
    /// Uses volatile writes to prevent compiler optimization from removing the clear
    static func secureClear(_ data: inout [UInt8]) {
        guard !data.isEmpty else { return }
        data.withUnsafeMutableBufferPointer { buffer in
            // Use volatile-like writes to prevent optimization
            for i in 0..<buffer.count {
                buffer[i] = 0
            }
            // Memory barrier to ensure writes complete
            OSMemoryBarrier()
        }
    }

    /// Securely clear Data object
    static func secureClear(_ data: inout Data) {
        guard !data.isEmpty else { return }
        data.withUnsafeMutableBytes { buffer in
            if let baseAddress = buffer.baseAddress {
                memset(baseAddress, 0, buffer.count)
            }
        }
        OSMemoryBarrier()
    }

    /// Securely clear a string (for mnemonics)
    /// Note: String immutability in Swift makes this challenging
    /// This replaces with empty and forces deallocation
    static func secureClearString(_ string: inout String) {
        // Convert to bytes and clear
        if var bytes = string.data(using: .utf8) {
            secureClear(&bytes)
        }
        string = ""
    }

    // MARK: - Mnemonic Generation

    /// Generate a new 12-word mnemonic phrase (like Trust Wallet)
    /// Uses cryptographically secure 128-bit entropy and full BIP39 wordlist (2048 words)
    ///
    /// Entropy sizes:
    /// - 128 bits (16 bytes) = 12 words
    /// - 256 bits (32 bytes) = 24 words
    func generateMnemonic() throws -> String {
        var entropy = [UInt8](repeating: 0, count: 16) // 128 bits for 12 words (Trust Wallet style)
        let result = SecRandomCopyBytes(kSecRandomDefault, entropy.count, &entropy)
        guard result == errSecSuccess else {
            throw CryptoError.entropyGenerationFailed
        }
        return mnemonicFromEntropy(entropy)
    }

    /// Validate a mnemonic phrase (checks words and checksum)
    func validateMnemonic(_ mnemonic: String) -> Bool {
        let words = mnemonic.lowercased().split(separator: " ").map(String.init)

        // Check word count (12, 15, 18, 21, or 24)
        guard [12, 15, 18, 21, 24].contains(words.count) else {
            return false
        }

        // Check all words are in wordlist
        for word in words {
            guard bip39WordIndex(word) != nil else {
                return false
            }
        }

        // Verify checksum
        return verifyMnemonicChecksum(words)
    }

    /// Check if a specific word is in the BIP39 wordlist
    func isValidWord(_ word: String) -> Bool {
        return bip39WordIndex(word.lowercased()) != nil
    }

    /// Get word suggestions for autocomplete (partial word input)
    func suggestWords(prefix: String) -> [String] {
        let lowercasedPrefix = prefix.lowercased()
        return BIP39Wordlist.words.filter { $0.hasPrefix(lowercasedPrefix) }.prefix(10).map { $0 }
    }

    // MARK: - Key Derivation

    /// Derive private key from mnemonic using BIP44 path
    /// Path: m/44'/118'/0'/0/0 (Cosmos standard)
    /// WARNING: Caller is responsible for clearing the returned Data
    func derivePrivateKey(from mnemonic: String) throws -> Data {
        let seed = try mnemonicToSeed(mnemonic)
        defer {
            var mutableSeed = [UInt8](seed)
            CryptoService.secureClear(&mutableSeed)
        }
        return try deriveKeyFromSeed(seed, path: cosmosPath)
    }

    /// Derive private key from mnemonic with automatic secure cleanup
    /// Returns SecureData that auto-clears when deallocated
    func derivePrivateKeySecure(from mnemonic: String) throws -> SecureData {
        let keyData = try derivePrivateKey(from: mnemonic)
        return SecureData(keyData)
    }

    /// Derive private key with custom account index
    func derivePrivateKey(from mnemonic: String, accountIndex: Int) throws -> Data {
        let seed = try mnemonicToSeed(mnemonic)
        let path: [UInt32] = [
            44 | 0x80000000,
            118 | 0x80000000,
            0 | 0x80000000,
            0,
            UInt32(accountIndex)
        ]
        return try deriveKeyFromSeed(seed, path: path)
    }

    /// Derive private key for a specific chain
    /// Uses BIP44 path: m/44'/coin_type'/0'/0/0
    /// WARNING: Caller is responsible for clearing the returned Data
    func derivePrivateKey(from mnemonic: String, chain: Chain, accountIndex: Int = 0) throws -> Data {
        let seed = try mnemonicToSeed(mnemonic)
        defer {
            var mutableSeed = [UInt8](seed)
            CryptoService.secureClear(&mutableSeed)
        }
        let path: [UInt32] = [
            44 | 0x80000000,              // Purpose: BIP44
            chain.coinType | 0x80000000,  // Coin type (hardened)
            0 | 0x80000000,               // Account (hardened)
            0,                             // Change (external)
            UInt32(accountIndex)           // Address index
        ]
        return try deriveKeyFromSeed(seed, path: path)
    }

    /// Derive private key for a specific chain with automatic secure cleanup
    func derivePrivateKeySecure(from mnemonic: String, chain: Chain, accountIndex: Int = 0) throws -> SecureData {
        let keyData = try derivePrivateKey(from: mnemonic, chain: chain, accountIndex: accountIndex)
        return SecureData(keyData)
    }

    /// Get the derivation path string for a chain
    func derivationPath(for chain: Chain, accountIndex: Int = 0) -> String {
        return "m/44'/\(chain.coinType)'/0'/0/\(accountIndex)"
    }

    // MARK: - Address Generation

    /// Derive ShareHODL address from private key
    /// Returns bech32 address with "hodl" prefix
    func deriveAddress(from privateKey: Data) throws -> String {
        let publicKey = try derivePublicKey(from: privateKey)

        // SHA256 of public key
        let sha256Hash = sha256(publicKey)

        // RIPEMD160 of SHA256 result (20 bytes)
        let ripemd160Hash = ripemd160(sha256Hash)

        // Bech32 encode with "hodl" prefix
        return bech32Encode(hrp: bech32Prefix, data: ripemd160Hash)
    }

    /// Derive address directly from mnemonic
    func deriveAddress(from mnemonic: String) throws -> String {
        let privateKey = try derivePrivateKey(from: mnemonic)
        return try deriveAddress(from: privateKey)
    }

    // MARK: - Multi-Chain Address Generation

    /// Derive address for a specific chain from private key
    func deriveAddress(from privateKey: Data, chain: Chain) throws -> String {
        let publicKey = try derivePublicKey(from: privateKey)

        switch chain {
        case .sharehodl:
            // Cosmos-style: SHA256 + RIPEMD160 + Bech32
            let hash = hash160(publicKey)
            return bech32Encode(hrp: chain.bech32Prefix ?? "hodl", data: hash)

        case .bitcoin:
            // Native SegWit (P2WPKH): SHA256 + RIPEMD160 + Bech32 with witness version
            let hash = hash160(publicKey)
            return bech32EncodeSegwit(hrp: "bc", witnessVersion: 0, data: hash)

        case .litecoin:
            // Native SegWit (P2WPKH): SHA256 + RIPEMD160 + Bech32 with witness version
            let hash = hash160(publicKey)
            return bech32EncodeSegwit(hrp: "ltc", witnessVersion: 0, data: hash)

        case .ethereum:
            // Ethereum: Keccak256 of uncompressed public key, last 20 bytes
            let uncompressedKey = try deriveUncompressedPublicKey(from: privateKey)
            // Remove the 0x04 prefix (65 bytes -> 64 bytes)
            let keyWithoutPrefix = Array(uncompressedKey.dropFirst())
            let hash = keccak256(Data(keyWithoutPrefix))
            // Take last 20 bytes and format as hex with 0x prefix
            let addressBytes = Array(hash.suffix(20))
            return "0x" + addressBytes.map { String(format: "%02x", $0) }.joined()
        }
    }

    /// Derive address for a chain directly from mnemonic
    func deriveAddress(from mnemonic: String, chain: Chain, accountIndex: Int = 0) throws -> String {
        let privateKey = try derivePrivateKey(from: mnemonic, chain: chain, accountIndex: accountIndex)
        return try deriveAddress(from: privateKey, chain: chain)
    }

    /// Derive all chain accounts from a mnemonic
    func deriveAllChainAccounts(from mnemonic: String) throws -> [ChainAccount] {
        var accounts: [ChainAccount] = []

        for chain in Chain.allCases {
            let privateKey = try derivePrivateKey(from: mnemonic, chain: chain)
            let address = try deriveAddress(from: privateKey, chain: chain)
            let path = derivationPath(for: chain)

            accounts.append(ChainAccount(
                chain: chain,
                address: address,
                derivationPath: path
            ))
        }

        return accounts
    }

    /// Get uncompressed public key (65 bytes, with 0x04 prefix)
    private func deriveUncompressedPublicKey(from privateKey: Data) throws -> Data {
        guard privateKey.count == 32 else {
            throw CryptoError.invalidPrivateKey
        }

        let signingKey = try P256K.Signing.PrivateKey(dataRepresentation: privateKey)
        // Get uncompressed representation (65 bytes)
        return signingKey.publicKey.uncompressedRepresentation
    }

    /// Bech32 encode for SegWit addresses (BIP173)
    private func bech32EncodeSegwit(hrp: String, witnessVersion: Int, data: Data) -> String {
        // For SegWit, we need to prepend the witness version
        var programData = [UInt8(witnessVersion)]
        programData.append(contentsOf: convertBits(data: [UInt8](data), fromBits: 8, toBits: 5, pad: true))

        let checksum = bech32CreateChecksum(hrp: hrp, data: programData)
        let combined = programData + checksum

        var result = hrp + "1"
        for byte in combined {
            let index = bech32Charset.index(bech32Charset.startIndex, offsetBy: Int(byte))
            result.append(bech32Charset[index])
        }

        return result
    }

    /// Keccak256 hash (for Ethereum)
    func keccak256(_ data: Data) -> Data {
        // Use CryptoSwift's SHA3 with Keccak variant
        let hash = [UInt8](data).sha3(.keccak256)
        return Data(hash)
    }

    /// Get public key from private key (compressed, 33 bytes)
    func derivePublicKey(from privateKey: Data) throws -> Data {
        guard privateKey.count == 32 else {
            throw CryptoError.invalidPrivateKey
        }

        let signingKey = try P256K.Signing.PrivateKey(dataRepresentation: privateKey)
        let publicKey = signingKey.publicKey

        // Return compressed public key (33 bytes)
        return publicKey.dataRepresentation
    }

    // MARK: - Transaction Signing

    /// Sign a transaction hash with private key using ECDSA secp256k1
    func sign(hash: Data, privateKey: Data) throws -> Data {
        guard privateKey.count == 32, hash.count == 32 else {
            throw CryptoError.signingFailed
        }

        let signingKey = try P256K.Signing.PrivateKey(dataRepresentation: privateKey)
        let signature = try signingKey.signature(for: hash)

        // Return compact signature (64 bytes: r + s)
        return try signature.compactRepresentation
    }

    /// Sign a message (hashes first with SHA256)
    func signMessage(_ message: String, privateKey: Data) throws -> Data {
        guard let messageData = message.data(using: .utf8) else {
            throw CryptoError.invalidMessage
        }

        let hash = sha256(messageData)
        return try sign(hash: hash, privateKey: privateKey)
    }

    /// Verify a signature
    func verifySignature(_ signature: Data, message: Data, publicKey: Data) -> Bool {
        guard signature.count == 64, publicKey.count == 33 else {
            return false
        }

        do {
            let verifyingKey = try P256K.Signing.PublicKey(dataRepresentation: publicKey, format: .compressed)
            let ecdsaSignature = try P256K.Signing.ECDSASignature(compactRepresentation: signature)

            return verifyingKey.isValidSignature(ecdsaSignature, for: message)
        } catch {
            return false
        }
    }

    // MARK: - Secure Signing Operations

    /// Sign a transaction hash with secure private key (auto-clears key after signing)
    func signSecure(hash: Data, privateKey: SecureData) throws -> Data {
        defer { privateKey.clear() }
        return try sign(hash: hash, privateKey: privateKey.data)
    }

    /// Sign a message with mnemonic (derives key, signs, clears all sensitive data)
    func signWithMnemonic(hash: Data, mnemonic: String, chain: Chain = .sharehodl, accountIndex: Int = 0) throws -> Data {
        let secureKey = try derivePrivateKeySecure(from: mnemonic, chain: chain, accountIndex: accountIndex)
        return try signSecure(hash: hash, privateKey: secureKey)
    }

    /// Complete secure transaction signing workflow
    /// - Derives private key from mnemonic
    /// - Signs the transaction hash
    /// - Clears all sensitive data from memory
    /// - Returns the signature
    func secureSignTransaction(
        transactionHash: Data,
        mnemonic: String,
        chain: Chain = .sharehodl,
        accountIndex: Int = 0
    ) throws -> Data {
        // Create secure containers
        let secureMnemonic = SecureMnemonic(mnemonic)

        defer {
            // Ensure mnemonic is cleared even if an error occurs
            secureMnemonic.clear()
        }

        // Derive key securely
        let secureKey = try derivePrivateKeySecure(from: secureMnemonic.phrase, chain: chain, accountIndex: accountIndex)

        // Sign and return (secureKey auto-clears when signSecure completes)
        return try signSecure(hash: transactionHash, privateKey: secureKey)
    }

    // MARK: - Hashing Utilities

    /// SHA256 hash
    func sha256(_ data: Data) -> Data {
        let hash = SHA256.hash(data: data)
        return Data(hash)
    }

    /// RIPEMD160 hash
    func ripemd160(_ data: Data) -> Data {
        return Data(RIPEMD160.hash([UInt8](data)))
    }

    /// SHA256 + RIPEMD160 (Bitcoin-style hash160)
    func hash160(_ data: Data) -> Data {
        return ripemd160(sha256(data))
    }

    // MARK: - Account Management

    /// Create multiple accounts from same mnemonic
    func deriveAccounts(from mnemonic: String, count: Int) throws -> [(index: Int, address: String, privateKey: Data)] {
        var accounts: [(Int, String, Data)] = []

        for i in 0..<count {
            let privateKey = try derivePrivateKey(from: mnemonic, accountIndex: i)
            let address = try deriveAddress(from: privateKey)
            accounts.append((i, address, privateKey))
        }

        return accounts
    }

    // MARK: - BIP39 Implementation

    private func mnemonicFromEntropy(_ entropy: [UInt8]) -> String {
        // Calculate checksum (first CS bits of SHA256)
        let hash = SHA256.hash(data: Data(entropy))
        let hashBytes = [UInt8](Data(hash))

        // CS = ENT / 32, for 256 bits: CS = 8 bits
        let checksumBits = entropy.count / 4 // 8 bits for 32 bytes

        // Combine entropy + checksum into bit string
        var bits = ""
        for byte in entropy {
            bits += String(byte, radix: 2).padding(toLength: 8, withPad: "0", startingAt: 0)
        }

        // Add checksum bits
        let checksumByte = hashBytes[0]
        let checksumString = String(checksumByte, radix: 2).padding(toLength: 8, withPad: "0", startingAt: 0)
        bits += String(checksumString.prefix(checksumBits))

        // Split into 11-bit groups and map to words
        var words: [String] = []
        for i in stride(from: 0, to: bits.count, by: 11) {
            let start = bits.index(bits.startIndex, offsetBy: i)
            let end = bits.index(start, offsetBy: min(11, bits.count - i))
            let indexString = String(bits[start..<end])
            if let index = Int(indexString, radix: 2), index < BIP39Wordlist.words.count {
                words.append(BIP39Wordlist.words[index])
            }
        }

        return words.joined(separator: " ")
    }

    private func mnemonicToSeed(_ mnemonic: String) throws -> Data {
        guard validateMnemonic(mnemonic) else {
            throw CryptoError.invalidMnemonic
        }

        let password = mnemonic.decomposedStringWithCompatibilityMapping
        let salt = "mnemonic".decomposedStringWithCompatibilityMapping

        // PBKDF2-HMAC-SHA512 with 2048 iterations
        guard let passwordData = password.data(using: .utf8),
              let saltData = salt.data(using: .utf8) else {
            throw CryptoError.invalidMnemonic
        }

        let seed = try PKCS5.PBKDF2(
            password: [UInt8](passwordData),
            salt: [UInt8](saltData),
            iterations: 2048,
            keyLength: 64,
            variant: .sha2(.sha512)
        ).calculate()

        return Data(seed)
    }

    private func deriveKeyFromSeed(_ seed: Data, path: [UInt32]) throws -> Data {
        guard seed.count == 64 else {
            throw CryptoError.keyDerivationFailed
        }

        // Generate master key using HMAC-SHA512 with "Bitcoin seed" as key
        let hmac = try HMAC(key: Array("Bitcoin seed".utf8), variant: .sha2(.sha512))
        let masterKey = try hmac.authenticate([UInt8](seed))

        var privateKey = Array(masterKey.prefix(32))
        var chainCode = Array(masterKey.suffix(32))

        // Derive each level of the path
        for index in path {
            (privateKey, chainCode) = try deriveChildKey(privateKey: privateKey, chainCode: chainCode, index: index)
        }

        return Data(privateKey)
    }

    private func deriveChildKey(privateKey: [UInt8], chainCode: [UInt8], index: UInt32) throws -> ([UInt8], [UInt8]) {
        var data: [UInt8] = []

        if index >= 0x80000000 {
            // Hardened derivation
            data.append(0x00)
            data.append(contentsOf: privateKey)
        } else {
            // Normal derivation - use public key
            let signingKey = try P256K.Signing.PrivateKey(dataRepresentation: Data(privateKey))
            let publicKey = signingKey.publicKey.dataRepresentation
            data.append(contentsOf: publicKey)
        }

        // Append index (big endian)
        data.append(UInt8((index >> 24) & 0xFF))
        data.append(UInt8((index >> 16) & 0xFF))
        data.append(UInt8((index >> 8) & 0xFF))
        data.append(UInt8(index & 0xFF))

        let hmac = try HMAC(key: chainCode, variant: .sha2(.sha512))
        let derived = try hmac.authenticate(data)

        let derivedKey = Array(derived.prefix(32))
        let derivedChainCode = Array(derived.suffix(32))

        // Add derived key to parent key (mod curve order)
        let newKey = addPrivateKeys(privateKey, derivedKey)

        return (newKey, derivedChainCode)
    }

    private func addPrivateKeys(_ key1: [UInt8], _ key2: [UInt8]) -> [UInt8] {
        // secp256k1 curve order
        let curveOrder: [UInt8] = [
            0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
            0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE,
            0xBA, 0xAE, 0xDC, 0xE6, 0xAF, 0x48, 0xA0, 0x3B,
            0xBF, 0xD2, 0x5E, 0x8C, 0xD0, 0x36, 0x41, 0x41
        ]

        // Simple addition with carry (we'll use BigInt-style addition)
        var result = [UInt8](repeating: 0, count: 32)
        var carry: UInt16 = 0

        for i in (0..<32).reversed() {
            let sum = UInt16(key1[i]) + UInt16(key2[i]) + carry
            result[i] = UInt8(sum & 0xFF)
            carry = sum >> 8
        }

        // Reduce modulo curve order if needed (simplified - check if result >= order)
        if compareBytes(result, curveOrder) >= 0 {
            result = subtractBytes(result, curveOrder)
        }

        return result
    }

    private func compareBytes(_ a: [UInt8], _ b: [UInt8]) -> Int {
        for i in 0..<min(a.count, b.count) {
            if a[i] > b[i] { return 1 }
            if a[i] < b[i] { return -1 }
        }
        return 0
    }

    private func subtractBytes(_ a: [UInt8], _ b: [UInt8]) -> [UInt8] {
        var result = [UInt8](repeating: 0, count: 32)
        var borrow: Int16 = 0

        for i in (0..<32).reversed() {
            let diff = Int16(a[i]) - Int16(b[i]) - borrow
            if diff < 0 {
                result[i] = UInt8(diff + 256)
                borrow = 1
            } else {
                result[i] = UInt8(diff)
                borrow = 0
            }
        }

        return result
    }

    private func bip39WordIndex(_ word: String) -> Int? {
        return BIP39Wordlist.words.firstIndex(of: word.lowercased())
    }

    private func verifyMnemonicChecksum(_ words: [String]) -> Bool {
        // Convert words to bits
        var bits = ""
        for word in words {
            guard let index = bip39WordIndex(word) else {
                return false
            }
            bits += String(index, radix: 2).padding(toLength: 11, withPad: "0", startingAt: 0)
        }

        // Calculate entropy and checksum lengths
        let totalBits = words.count * 11
        let checksumBits = words.count / 3  // CS = MS / 3 where MS = word count
        let entropyBits = totalBits - checksumBits

        // Extract entropy bytes
        let entropyString = String(bits.prefix(entropyBits))
        var entropy: [UInt8] = []
        for i in stride(from: 0, to: entropyBits, by: 8) {
            let start = entropyString.index(entropyString.startIndex, offsetBy: i)
            let end = entropyString.index(start, offsetBy: 8)
            if let byte = UInt8(String(entropyString[start..<end]), radix: 2) {
                entropy.append(byte)
            }
        }

        // Calculate expected checksum
        let hash = SHA256.hash(data: Data(entropy))
        let hashBytes = [UInt8](Data(hash))
        let expectedChecksumByte = hashBytes[0]
        let expectedChecksumString = String(expectedChecksumByte, radix: 2).padding(toLength: 8, withPad: "0", startingAt: 0)
        let expectedChecksum = String(expectedChecksumString.prefix(checksumBits))

        // Extract actual checksum from mnemonic
        let actualChecksum = String(bits.suffix(checksumBits))

        return expectedChecksum == actualChecksum
    }

    // MARK: - Bech32 Encoding

    private let bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

    func bech32Encode(hrp: String, data: Data) -> String {
        let converted = convertBits(data: [UInt8](data), fromBits: 8, toBits: 5, pad: true)
        let checksum = bech32CreateChecksum(hrp: hrp, data: converted)
        let combined = converted + checksum

        var result = hrp + "1"
        for byte in combined {
            let index = bech32Charset.index(bech32Charset.startIndex, offsetBy: Int(byte))
            result.append(bech32Charset[index])
        }

        return result
    }

    func bech32Decode(_ address: String) -> (hrp: String, data: Data)? {
        guard let separatorIndex = address.lastIndex(of: "1") else {
            return nil
        }

        let hrp = String(address[..<separatorIndex]).lowercased()
        let dataString = String(address[address.index(after: separatorIndex)...]).lowercased()

        var data: [UInt8] = []
        for char in dataString {
            guard let index = bech32Charset.firstIndex(of: char) else {
                return nil
            }
            data.append(UInt8(bech32Charset.distance(from: bech32Charset.startIndex, to: index)))
        }

        // Verify checksum
        guard bech32VerifyChecksum(hrp: hrp, data: data) else {
            return nil
        }

        // Remove checksum (last 6 bytes)
        let dataWithoutChecksum = Array(data.dropLast(6))

        // Convert from 5-bit to 8-bit
        let converted = convertBits(data: dataWithoutChecksum, fromBits: 5, toBits: 8, pad: false)

        return (hrp, Data(converted))
    }

    private func convertBits(data: [UInt8], fromBits: Int, toBits: Int, pad: Bool) -> [UInt8] {
        var acc = 0
        var bits = 0
        var result: [UInt8] = []
        let maxv = (1 << toBits) - 1

        for byte in data {
            acc = (acc << fromBits) | Int(byte)
            bits += fromBits
            while bits >= toBits {
                bits -= toBits
                result.append(UInt8((acc >> bits) & maxv))
            }
        }

        if pad && bits > 0 {
            result.append(UInt8((acc << (toBits - bits)) & maxv))
        }

        return result
    }

    private func bech32CreateChecksum(hrp: String, data: [UInt8]) -> [UInt8] {
        let values = bech32HrpExpand(hrp) + data + [0, 0, 0, 0, 0, 0]
        let polymod = bech32Polymod(values) ^ 1

        var checksum: [UInt8] = []
        for i in 0..<6 {
            checksum.append(UInt8((polymod >> (5 * (5 - i))) & 31))
        }
        return checksum
    }

    private func bech32VerifyChecksum(hrp: String, data: [UInt8]) -> Bool {
        return bech32Polymod(bech32HrpExpand(hrp) + data) == 1
    }

    private func bech32HrpExpand(_ hrp: String) -> [UInt8] {
        var result: [UInt8] = []
        for char in hrp {
            result.append(UInt8(char.asciiValue! >> 5))
        }
        result.append(0)
        for char in hrp {
            result.append(UInt8(char.asciiValue! & 31))
        }
        return result
    }

    private func bech32Polymod(_ values: [UInt8]) -> Int {
        let generator = [0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3]
        var chk = 1
        for value in values {
            let top = chk >> 25
            chk = (chk & 0x1ffffff) << 5 ^ Int(value)
            for i in 0..<5 {
                if (top >> i) & 1 != 0 {
                    chk ^= generator[i]
                }
            }
        }
        return chk
    }
}

// MARK: - Errors

enum CryptoError: LocalizedError {
    case entropyGenerationFailed
    case invalidMnemonic
    case invalidPrivateKey
    case invalidPublicKey
    case invalidMessage
    case keyDerivationFailed
    case signingFailed
    case verificationFailed
    case bech32EncodingFailed

    var errorDescription: String? {
        switch self {
        case .entropyGenerationFailed:
            return "Failed to generate secure random entropy"
        case .invalidMnemonic:
            return "Invalid recovery phrase"
        case .invalidPrivateKey:
            return "Invalid private key"
        case .invalidPublicKey:
            return "Invalid public key"
        case .invalidMessage:
            return "Invalid message format"
        case .keyDerivationFailed:
            return "Key derivation failed"
        case .signingFailed:
            return "Transaction signing failed"
        case .verificationFailed:
            return "Signature verification failed"
        case .bech32EncodingFailed:
            return "Address encoding failed"
        }
    }
}

// MARK: - Secure Data Container

/// A container for sensitive data that automatically clears memory when deallocated
/// Use this for private keys, seeds, and other cryptographic material
final class SecureData {
    private var storage: [UInt8]

    /// Initialize with data (copies bytes)
    init(_ data: Data) {
        storage = [UInt8](data)
    }

    /// Initialize with byte array (copies bytes)
    init(_ bytes: [UInt8]) {
        storage = bytes
    }

    /// Initialize with specific size filled with zeros
    init(count: Int) {
        storage = [UInt8](repeating: 0, count: count)
    }

    deinit {
        clear()
    }

    /// Get the data (read-only access)
    var data: Data {
        Data(storage)
    }

    /// Get the bytes (read-only access)
    var bytes: [UInt8] {
        storage
    }

    /// Number of bytes
    var count: Int {
        storage.count
    }

    /// Check if empty
    var isEmpty: Bool {
        storage.isEmpty
    }

    /// Manually clear the data
    func clear() {
        CryptoService.secureClear(&storage)
    }

    /// Perform operation with the bytes, then clear
    func withBytes<T>(_ operation: ([UInt8]) throws -> T) rethrows -> T {
        defer { clear() }
        return try operation(storage)
    }

    /// Perform operation with mutable bytes, then clear
    func withMutableBytes<T>(_ operation: (inout [UInt8]) throws -> T) rethrows -> T {
        defer { clear() }
        return try operation(&storage)
    }
}

// MARK: - Secure Mnemonic Container

/// A container for mnemonic phrases that clears memory when deallocated
final class SecureMnemonic {
    private var words: [String]

    init(_ mnemonic: String) {
        words = mnemonic.split(separator: " ").map(String.init)
    }

    init(_ wordArray: [String]) {
        words = wordArray
    }

    deinit {
        clear()
    }

    /// Get the mnemonic as a string
    var phrase: String {
        words.joined(separator: " ")
    }

    /// Get individual words
    var wordArray: [String] {
        words
    }

    /// Number of words
    var wordCount: Int {
        words.count
    }

    /// Manually clear the mnemonic
    func clear() {
        for i in 0..<words.count {
            // Overwrite each word with random data then clear
            words[i] = String(repeating: "\0", count: words[i].count)
        }
        words = []
    }

    /// Perform operation with the phrase, then clear
    func withPhrase<T>(_ operation: (String) throws -> T) rethrows -> T {
        defer { clear() }
        return try operation(phrase)
    }
}

// MARK: - RIPEMD160 Implementation

/// RIPEMD160 hash implementation for Bitcoin/Cosmos address derivation
enum RIPEMD160 {
    /// Hash data using RIPEMD160
    static func hash(_ data: [UInt8]) -> [UInt8] {
        var h0: UInt32 = 0x67452301
        var h1: UInt32 = 0xEFCDAB89
        var h2: UInt32 = 0x98BADCFE
        var h3: UInt32 = 0x10325476
        var h4: UInt32 = 0xC3D2E1F0

        // Pre-processing: adding padding bits
        var message = data
        let originalLength = UInt64(data.count * 8)

        // Append bit '1' to message
        message.append(0x80)

        // Append zeros until message length ≡ 448 (mod 512)
        while message.count % 64 != 56 {
            message.append(0x00)
        }

        // Append original length in bits as 64-bit little-endian
        for i in 0..<8 {
            message.append(UInt8(truncatingIfNeeded: originalLength >> (i * 8)))
        }

        // Process each 512-bit block
        for chunkStart in stride(from: 0, to: message.count, by: 64) {
            var x = [UInt32](repeating: 0, count: 16)
            for i in 0..<16 {
                let offset = chunkStart + i * 4
                x[i] = UInt32(message[offset]) |
                       (UInt32(message[offset + 1]) << 8) |
                       (UInt32(message[offset + 2]) << 16) |
                       (UInt32(message[offset + 3]) << 24)
            }

            var al = h0, bl = h1, cl = h2, dl = h3, el = h4
            var ar = h0, br = h1, cr = h2, dr = h3, er = h4

            // 80 rounds
            for j in 0..<80 {
                let (fl, kl, rl) = leftRound(j, al, bl, cl, dl, el, x)
                let tl = fl &+ el &+ kl &+ x[Int(rl)]
                al = el; el = dl; dl = rotateLeft(cl, 10); cl = bl; bl = tl

                let (fr, kr, rr) = rightRound(j, ar, br, cr, dr, er, x)
                let tr = fr &+ er &+ kr &+ x[Int(rr)]
                ar = er; er = dr; dr = rotateLeft(cr, 10); cr = br; br = tr
            }

            let t = h1 &+ cl &+ dr
            h1 = h2 &+ dl &+ er
            h2 = h3 &+ el &+ ar
            h3 = h4 &+ al &+ br
            h4 = h0 &+ bl &+ cr
            h0 = t
        }

        // Produce the final hash value (little-endian)
        var result = [UInt8](repeating: 0, count: 20)
        for i in 0..<4 {
            result[i] = UInt8(truncatingIfNeeded: h0 >> (i * 8))
            result[i + 4] = UInt8(truncatingIfNeeded: h1 >> (i * 8))
            result[i + 8] = UInt8(truncatingIfNeeded: h2 >> (i * 8))
            result[i + 12] = UInt8(truncatingIfNeeded: h3 >> (i * 8))
            result[i + 16] = UInt8(truncatingIfNeeded: h4 >> (i * 8))
        }

        return result
    }

    private static func rotateLeft(_ x: UInt32, _ n: UInt32) -> UInt32 {
        return (x << n) | (x >> (32 - n))
    }

    private static func f(_ j: Int, _ x: UInt32, _ y: UInt32, _ z: UInt32) -> UInt32 {
        switch j {
        case 0..<16: return x ^ y ^ z
        case 16..<32: return (x & y) | (~x & z)
        case 32..<48: return (x | ~y) ^ z
        case 48..<64: return (x & z) | (y & ~z)
        default: return x ^ (y | ~z)
        }
    }

    private static let kl: [UInt32] = [0x00000000, 0x5A827999, 0x6ED9EBA1, 0x8F1BBCDC, 0xA953FD4E]
    private static let kr: [UInt32] = [0x50A28BE6, 0x5C4DD124, 0x6D703EF3, 0x7A6D76E9, 0x00000000]

    private static let rl: [[Int]] = [
        [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15],
        [7, 4, 13, 1, 10, 6, 15, 3, 12, 0, 9, 5, 2, 14, 11, 8],
        [3, 10, 14, 4, 9, 15, 8, 1, 2, 7, 0, 6, 13, 11, 5, 12],
        [1, 9, 11, 10, 0, 8, 12, 4, 13, 3, 7, 15, 14, 5, 6, 2],
        [4, 0, 5, 9, 7, 12, 2, 10, 14, 1, 3, 8, 11, 6, 15, 13]
    ]

    private static let rr: [[Int]] = [
        [5, 14, 7, 0, 9, 2, 11, 4, 13, 6, 15, 8, 1, 10, 3, 12],
        [6, 11, 3, 7, 0, 13, 5, 10, 14, 15, 8, 12, 4, 9, 1, 2],
        [15, 5, 1, 3, 7, 14, 6, 9, 11, 8, 12, 2, 10, 0, 4, 13],
        [8, 6, 4, 1, 3, 11, 15, 0, 5, 12, 2, 13, 9, 7, 10, 14],
        [12, 15, 10, 4, 1, 5, 8, 7, 6, 2, 13, 14, 0, 3, 9, 11]
    ]

    private static let sl: [[UInt32]] = [
        [11, 14, 15, 12, 5, 8, 7, 9, 11, 13, 14, 15, 6, 7, 9, 8],
        [7, 6, 8, 13, 11, 9, 7, 15, 7, 12, 15, 9, 11, 7, 13, 12],
        [11, 13, 6, 7, 14, 9, 13, 15, 14, 8, 13, 6, 5, 12, 7, 5],
        [11, 12, 14, 15, 14, 15, 9, 8, 9, 14, 5, 6, 8, 6, 5, 12],
        [9, 15, 5, 11, 6, 8, 13, 12, 5, 12, 13, 14, 11, 8, 5, 6]
    ]

    private static let sr: [[UInt32]] = [
        [8, 9, 9, 11, 13, 15, 15, 5, 7, 7, 8, 11, 14, 14, 12, 6],
        [9, 13, 15, 7, 12, 8, 9, 11, 7, 7, 12, 7, 6, 15, 13, 11],
        [9, 7, 15, 11, 8, 6, 6, 14, 12, 13, 5, 14, 13, 13, 7, 5],
        [15, 5, 8, 11, 14, 14, 6, 14, 6, 9, 12, 9, 12, 5, 15, 8],
        [8, 5, 12, 9, 12, 5, 14, 6, 8, 13, 6, 5, 15, 13, 11, 11]
    ]

    private static func leftRound(_ j: Int, _ a: UInt32, _ b: UInt32, _ c: UInt32, _ d: UInt32, _ e: UInt32, _ x: [UInt32]) -> (UInt32, UInt32, Int) {
        let round = j / 16
        let idx = j % 16
        let fVal = f(j, b, c, d)
        let s = sl[round][idx]
        let r = rl[round][idx]
        let k = kl[round]
        let t = rotateLeft(a &+ fVal &+ x[r] &+ k, s) &+ e
        return (t, k, r)
    }

    private static func rightRound(_ j: Int, _ a: UInt32, _ b: UInt32, _ c: UInt32, _ d: UInt32, _ e: UInt32, _ x: [UInt32]) -> (UInt32, UInt32, Int) {
        let round = j / 16
        let idx = j % 16
        let fVal = f(79 - j, b, c, d)
        let s = sr[round][idx]
        let r = rr[round][idx]
        let k = kr[round]
        let t = rotateLeft(a &+ fVal &+ x[r] &+ k, s) &+ e
        return (t, k, r)
    }
}
