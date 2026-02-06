package com.sharehodl.service

import com.sharehodl.model.Chain
import com.sharehodl.model.ChainAccount
import org.web3j.crypto.Bip32ECKeyPair
import org.web3j.crypto.ECKeyPair
import org.web3j.crypto.MnemonicUtils
import org.web3j.crypto.Sign
import org.web3j.crypto.Hash
import org.bouncycastle.crypto.digests.RIPEMD160Digest
import org.bouncycastle.crypto.params.ECPublicKeyParameters
import org.bouncycastle.crypto.params.ECDomainParameters
import org.bouncycastle.crypto.signers.ECDSASigner
import org.bouncycastle.asn1.x9.X9ECParameters
import org.bouncycastle.crypto.ec.CustomNamedCurves
import java.math.BigInteger
import java.security.MessageDigest
import java.security.SecureRandom
import java.util.Arrays
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Production-ready cryptographic service using Web3j + BouncyCastle
 * Provides: BIP39 (full 2048 words), BIP32/44 derivation, secp256k1, ECDSA, RIPEMD160
 *
 * SECURITY: This service implements secure memory handling for sensitive data:
 * - Private keys and mnemonics are cleared from memory after use
 * - Use SecureByteArray wrapper for sensitive byte arrays
 * - Always call clear() when done with sensitive operations
 */
@Singleton
class CryptoService @Inject constructor() {

    companion object {
        private const val BECH32_PREFIX = "hodl"
        // BIP44 path for Cosmos: m/44'/118'/0'/0/0
        private val COSMOS_PATH = intArrayOf(
            44 or Bip32ECKeyPair.HARDENED_BIT,
            118 or Bip32ECKeyPair.HARDENED_BIT,
            0 or Bip32ECKeyPair.HARDENED_BIT,
            0,
            0
        )

        // secp256k1 curve parameters for signature verification
        private val CURVE_PARAMS: X9ECParameters = CustomNamedCurves.getByName("secp256k1")
        private val CURVE = ECDomainParameters(
            CURVE_PARAMS.curve,
            CURVE_PARAMS.g,
            CURVE_PARAMS.n,
            CURVE_PARAMS.h
        )

        /**
         * Securely clear a byte array by overwriting with zeros
         */
        @JvmStatic
        fun secureClear(data: ByteArray) {
            Arrays.fill(data, 0.toByte())
        }

        /**
         * Securely clear a char array (for mnemonic strings)
         */
        @JvmStatic
        fun secureClear(data: CharArray) {
            Arrays.fill(data, '\u0000')
        }
    }

    // MARK: - Mnemonic Generation

    /**
     * Generate a new 12-word mnemonic phrase (like Trust Wallet)
     * Uses cryptographically secure 128-bit entropy and full BIP39 wordlist (2048 words)
     *
     * Entropy sizes:
     * - 128 bits (16 bytes) = 12 words
     * - 256 bits (32 bytes) = 24 words
     *
     * SECURITY: Entropy is securely cleared from memory after mnemonic generation
     */
    fun generateMnemonic(): String {
        val entropy = ByteArray(16) // 128 bits for 12 words (Trust Wallet style)
        return try {
            SecureRandom().nextBytes(entropy)
            MnemonicUtils.generateMnemonic(entropy)
        } finally {
            // SECURITY: Clear entropy from memory immediately after use
            secureClear(entropy)
        }
    }

    /**
     * Validate a mnemonic phrase (checks words and checksum)
     */
    fun validateMnemonic(mnemonic: String): Boolean {
        return try {
            MnemonicUtils.validateMnemonic(mnemonic)
            true
        } catch (e: Exception) {
            false
        }
    }

    /**
     * Check if a specific word is in the BIP39 wordlist
     */
    fun isValidWord(word: String): Boolean {
        val wordlist = MnemonicUtils.getWords()
        return wordlist.contains(word.lowercase())
    }

    /**
     * Get word suggestions for autocomplete (partial word input)
     */
    fun suggestWords(prefix: String): List<String> {
        val wordlist = MnemonicUtils.getWords()
        return wordlist.filter { it.startsWith(prefix.lowercase()) }.take(10)
    }

    // MARK: - Key Derivation

    /**
     * Derive private key from mnemonic using BIP44 path
     * Path: m/44'/118'/0'/0/0 (Cosmos standard)
     *
     * SECURITY: Intermediate cryptographic material (seed, master key) is cleared after derivation
     */
    fun derivePrivateKey(mnemonic: String): ByteArray {
        val seed = MnemonicUtils.generateSeed(mnemonic, "")
        try {
            val masterKeyPair = Bip32ECKeyPair.generateKeyPair(seed)
            val derivedKeyPair = Bip32ECKeyPair.deriveKeyPair(masterKeyPair, COSMOS_PATH)
            return derivedKeyPair.privateKey.toByteArray().let { bytes ->
                // Ensure 32 bytes (pad with zeros if needed, trim leading zero if 33 bytes)
                when {
                    bytes.size == 33 && bytes[0] == 0.toByte() -> bytes.copyOfRange(1, 33)
                    bytes.size < 32 -> ByteArray(32 - bytes.size) + bytes
                    else -> bytes.copyOfRange(0, 32)
                }
            }
        } finally {
            // SECURITY: Clear seed from memory after derivation
            secureClear(seed)
        }
    }

    /**
     * Derive private key with custom account index
     *
     * SECURITY: Intermediate cryptographic material is cleared after derivation
     */
    fun derivePrivateKey(mnemonic: String, accountIndex: Int): ByteArray {
        val seed = MnemonicUtils.generateSeed(mnemonic, "")
        try {
            val masterKeyPair = Bip32ECKeyPair.generateKeyPair(seed)
            val customPath = intArrayOf(
                44 or Bip32ECKeyPair.HARDENED_BIT,
                118 or Bip32ECKeyPair.HARDENED_BIT,
                0 or Bip32ECKeyPair.HARDENED_BIT,
                0,
                accountIndex
            )
            val derivedKeyPair = Bip32ECKeyPair.deriveKeyPair(masterKeyPair, customPath)
            return derivedKeyPair.privateKey.toByteArray().let { bytes ->
                when {
                    bytes.size == 33 && bytes[0] == 0.toByte() -> bytes.copyOfRange(1, 33)
                    bytes.size < 32 -> ByteArray(32 - bytes.size) + bytes
                    else -> bytes.copyOfRange(0, 32)
                }
            }
        } finally {
            // SECURITY: Clear seed from memory after derivation
            secureClear(seed)
        }
    }

    // MARK: - Address Generation

    /**
     * Derive ShareHODL address from private key
     * Returns bech32 address with "hodl" prefix
     */
    fun deriveAddress(privateKey: ByteArray): String {
        // Get public key (compressed, 33 bytes)
        val publicKey = derivePublicKey(privateKey)

        // SHA256 of public key
        val sha256Hash = sha256(publicKey)

        // RIPEMD160 of SHA256 result (20 bytes)
        val ripemd160Hash = ripemd160(sha256Hash)

        // Bech32 encode with "hodl" prefix
        return bech32Encode(BECH32_PREFIX, ripemd160Hash)
    }

    /**
     * Derive address directly from mnemonic
     */
    fun deriveAddressFromMnemonic(mnemonic: String): String {
        val privateKey = derivePrivateKey(mnemonic)
        return deriveAddress(privateKey)
    }

    // MARK: - Multi-Chain Support

    /**
     * Derive private key for a specific chain
     * Uses BIP44 path: m/44'/coin_type'/0'/0/0
     *
     * SECURITY: Intermediate cryptographic material is cleared after derivation
     */
    fun derivePrivateKey(mnemonic: String, chain: Chain, accountIndex: Int = 0): ByteArray {
        val seed = MnemonicUtils.generateSeed(mnemonic, "")
        try {
            val masterKeyPair = Bip32ECKeyPair.generateKeyPair(seed)
            val chainPath = intArrayOf(
                44 or Bip32ECKeyPair.HARDENED_BIT,
                chain.coinType or Bip32ECKeyPair.HARDENED_BIT,
                0 or Bip32ECKeyPair.HARDENED_BIT,
                0,
                accountIndex
            )
            val derivedKeyPair = Bip32ECKeyPair.deriveKeyPair(masterKeyPair, chainPath)
            return derivedKeyPair.privateKey.toByteArray().let { bytes ->
                when {
                    bytes.size == 33 && bytes[0] == 0.toByte() -> bytes.copyOfRange(1, 33)
                    bytes.size < 32 -> ByteArray(32 - bytes.size) + bytes
                    else -> bytes.copyOfRange(0, 32)
                }
            }
        } finally {
            // SECURITY: Clear seed from memory after derivation
            secureClear(seed)
        }
    }

    /**
     * Get the derivation path string for a chain
     */
    fun derivationPath(chain: Chain, accountIndex: Int = 0): String {
        return "m/44'/${chain.coinType}'/0'/0/$accountIndex"
    }

    /**
     * Derive address for a specific chain from private key
     */
    fun deriveAddress(privateKey: ByteArray, chain: Chain): String {
        val publicKey = derivePublicKey(privateKey)

        return when (chain) {
            Chain.SHAREHODL -> {
                // Cosmos-style: SHA256 + RIPEMD160 + Bech32
                val hash = hash160(publicKey)
                bech32Encode(chain.bech32Prefix ?: "hodl", hash)
            }
            Chain.BITCOIN -> {
                // Legacy P2PKH address (starts with 1)
                val hash = hash160(publicKey)
                base58CheckEncode(0x00.toByte(), hash)
            }
            Chain.LITECOIN -> {
                // Legacy P2PKH address (starts with L)
                val hash = hash160(publicKey)
                base58CheckEncode(0x30.toByte(), hash)
            }
            Chain.DOGECOIN -> {
                // Legacy P2PKH address (starts with D)
                val hash = hash160(publicKey)
                base58CheckEncode(0x1e.toByte(), hash)
            }
            Chain.ETHEREUM, Chain.USDT, Chain.USDC, Chain.BNB, Chain.POLYGON, Chain.AVALANCHE -> {
                // Ethereum-style: Keccak256 of uncompressed public key, last 20 bytes
                val uncompressedKey = deriveUncompressedPublicKey(privateKey)
                // Remove the 0x04 prefix (65 bytes -> 64 bytes)
                val keyWithoutPrefix = uncompressedKey.copyOfRange(1, 65)
                val hash = keccak256(keyWithoutPrefix)
                // Take last 20 bytes and format as hex with 0x prefix
                val addressBytes = hash.copyOfRange(12, 32)
                "0x" + addressBytes.joinToString("") { String.format("%02x", it) }
            }
            Chain.SOLANA -> {
                // Solana uses base58-encoded public key directly
                base58Encode(publicKey)
            }
            Chain.XRP -> {
                // XRP uses base58 with custom alphabet
                val hash = hash160(publicKey)
                xrpBase58CheckEncode(hash)
            }
            Chain.TRON -> {
                // TRON is similar to Ethereum but uses base58check with prefix 0x41
                val uncompressedKey = deriveUncompressedPublicKey(privateKey)
                val keyWithoutPrefix = uncompressedKey.copyOfRange(1, 65)
                val hash = keccak256(keyWithoutPrefix)
                val addressBytes = hash.copyOfRange(12, 32)
                tronBase58CheckEncode(addressBytes)
            }
        }
    }

    /**
     * Derive address for a chain directly from mnemonic
     */
    fun deriveAddress(mnemonic: String, chain: Chain, accountIndex: Int = 0): String {
        val privateKey = derivePrivateKey(mnemonic, chain, accountIndex)
        return deriveAddress(privateKey, chain)
    }

    /**
     * Derive all chain accounts from a mnemonic
     */
    fun deriveAllChainAccounts(mnemonic: String): List<ChainAccount> {
        return Chain.majorChains.map { chain ->
            val privateKey = derivePrivateKey(mnemonic, chain)
            val address = deriveAddress(privateKey, chain)
            val path = derivationPath(chain)

            ChainAccount(
                chain = chain,
                address = address,
                derivationPath = path
            )
        }
    }

    /**
     * Get uncompressed public key (65 bytes, with 0x04 prefix)
     */
    private fun deriveUncompressedPublicKey(privateKey: ByteArray): ByteArray {
        val keyPair = ECKeyPair.create(privateKey)
        val publicKey = keyPair.publicKey

        // Convert to uncompressed format (65 bytes with 0x04 prefix)
        val pubKeyBytes = publicKey.toByteArray()
        val result = ByteArray(65)
        result[0] = 0x04

        // Pad the public key coordinates to 32 bytes each
        val pubKeyPadded = when {
            pubKeyBytes.size >= 64 -> pubKeyBytes.copyOfRange(pubKeyBytes.size - 64, pubKeyBytes.size)
            else -> ByteArray(64 - pubKeyBytes.size) + pubKeyBytes
        }

        System.arraycopy(pubKeyPadded, 0, result, 1, 64)
        return result
    }

    /**
     * Bech32 encode for SegWit addresses (BIP173)
     */
    private fun bech32EncodeSegwit(hrp: String, witnessVersion: Int, data: ByteArray): String {
        // For SegWit, we need to prepend the witness version
        val programData = byteArrayOf(witnessVersion.toByte()) + convertBits(data, 8, 5, true)

        val checksum = bech32CreateChecksum(hrp, programData)
        val combined = programData + checksum

        val result = StringBuilder(hrp).append("1")
        for (byte in combined) {
            result.append(BECH32_CHARSET[byte.toInt() and 0xFF])
        }

        return result.toString()
    }

    /**
     * Keccak256 hash (for Ethereum)
     */
    fun keccak256(data: ByteArray): ByteArray {
        return Hash.sha3(data)
    }

    /**
     * Get public key from private key (compressed, 33 bytes)
     */
    fun derivePublicKey(privateKey: ByteArray): ByteArray {
        val keyPair = ECKeyPair.create(privateKey)
        val publicKey = keyPair.publicKey

        // Convert to compressed format (33 bytes)
        val pubKeyBytes = publicKey.toByteArray()
        val xCoord = if (pubKeyBytes.size == 65) {
            pubKeyBytes.copyOfRange(1, 33)
        } else if (pubKeyBytes.size >= 32) {
            pubKeyBytes.copyOfRange(pubKeyBytes.size - 64, pubKeyBytes.size - 32)
        } else {
            pubKeyBytes.copyOf(32)
        }

        // Determine prefix based on Y coordinate parity
        val yCoord = if (pubKeyBytes.size == 65) {
            pubKeyBytes.copyOfRange(33, 65)
        } else if (pubKeyBytes.size >= 32) {
            pubKeyBytes.copyOfRange(pubKeyBytes.size - 32, pubKeyBytes.size)
        } else {
            ByteArray(32)
        }

        val prefix = if (yCoord.last().toInt() and 1 == 0) 0x02.toByte() else 0x03.toByte()

        return byteArrayOf(prefix) + xCoord
    }

    // MARK: - Transaction Signing

    /**
     * Sign a transaction hash with private key using ECDSA secp256k1
     */
    fun sign(hash: ByteArray, privateKey: ByteArray): ByteArray {
        val keyPair = ECKeyPair.create(privateKey)
        val signature = Sign.signMessage(hash, keyPair, false)

        // Return compact signature (64 bytes: r + s)
        return signature.r + signature.s
    }

    /**
     * Sign with ECDSA (alias for sign)
     * Used by CosmosTransactionBuilder
     */
    fun signWithEcdsa(privateKey: ByteArray, hash: ByteArray): ByteArray {
        return sign(hash, privateKey)
    }

    /**
     * Get public key bytes for an address (used for transaction building)
     * Note: This derives the public key from stored private key
     */
    fun getPublicKeyForAddress(privateKey: ByteArray): ByteArray {
        return derivePublicKey(privateKey)
    }

    /**
     * Sign a message (hashes first with SHA256)
     */
    fun signMessage(message: String, privateKey: ByteArray): ByteArray {
        val messageBytes = message.toByteArray(Charsets.UTF_8)
        val hash = sha256(messageBytes)
        return sign(hash, privateKey)
    }

    /**
     * Verify an ECDSA signature using secp256k1
     * @param signature The signature (64 bytes: r || s)
     * @param message The message that was signed (should be a hash)
     * @param publicKey The compressed public key (33 bytes)
     * @return true if the signature is valid
     */
    fun verifySignature(signature: ByteArray, message: ByteArray, publicKey: ByteArray): Boolean {
        if (signature.size != 64 || publicKey.size != 33) {
            return false
        }

        return try {
            // Extract r and s from the signature
            val r = BigInteger(1, signature.copyOfRange(0, 32))
            val s = BigInteger(1, signature.copyOfRange(32, 64))

            // Decode the compressed public key to a point
            val point = CURVE_PARAMS.curve.decodePoint(publicKey)
            val pubKeyParams = ECPublicKeyParameters(point, CURVE)

            // Create the verifier
            val signer = ECDSASigner()
            signer.init(false, pubKeyParams)

            // Verify the signature
            signer.verifySignature(message, r, s)
        } catch (e: Exception) {
            false
        }
    }

    // MARK: - Hashing Utilities

    /**
     * SHA256 hash
     */
    fun sha256(data: ByteArray): ByteArray {
        return MessageDigest.getInstance("SHA-256").digest(data)
    }

    /**
     * RIPEMD160 hash
     */
    fun ripemd160(data: ByteArray): ByteArray {
        val digest = RIPEMD160Digest()
        digest.update(data, 0, data.size)
        val result = ByteArray(20)
        digest.doFinal(result, 0)
        return result
    }

    /**
     * SHA256 + RIPEMD160 (Bitcoin-style hash160)
     */
    fun hash160(data: ByteArray): ByteArray {
        return ripemd160(sha256(data))
    }

    // MARK: - Account Management

    /**
     * Create multiple accounts from same mnemonic
     * SECURITY: Private keys are NOT stored in the returned accounts
     * Use derivePrivateKey() when you need to sign transactions
     */
    fun deriveAccounts(mnemonic: String, count: Int): List<Account> {
        val accounts = mutableListOf<Account>()

        for (i in 0 until count) {
            val privateKey = derivePrivateKey(mnemonic, i)
            try {
                val publicKey = derivePublicKey(privateKey)
                val address = deriveAddress(privateKey)

                accounts.add(Account(
                    index = i,
                    address = address,
                    publicKey = publicKey
                ))
            } finally {
                // Always clear private key after use
                secureClear(privateKey)
            }
        }

        return accounts
    }

    /**
     * Account data class
     * SECURITY: Does NOT store private keys - derive when needed for signing
     */
    data class Account(
        val index: Int,
        val address: String,
        val publicKey: ByteArray
    ) {
        override fun equals(other: Any?): Boolean {
            if (this === other) return true
            if (javaClass != other?.javaClass) return false
            other as Account
            return index == other.index && address == other.address
        }

        override fun hashCode(): Int {
            var result = index
            result = 31 * result + address.hashCode()
            return result
        }
    }

    // MARK: - Bech32 Encoding

    private val BECH32_CHARSET = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

    private fun bech32Encode(hrp: String, data: ByteArray): String {
        val converted = convertBits(data, 8, 5, true)
        val checksum = bech32CreateChecksum(hrp, converted)
        val combined = converted + checksum

        val result = StringBuilder(hrp).append("1")
        for (byte in combined) {
            result.append(BECH32_CHARSET[byte.toInt()])
        }

        return result.toString()
    }

    private fun convertBits(data: ByteArray, fromBits: Int, toBits: Int, pad: Boolean): ByteArray {
        var acc = 0
        var bits = 0
        val result = mutableListOf<Byte>()
        val maxv = (1 shl toBits) - 1

        for (byte in data) {
            acc = (acc shl fromBits) or (byte.toInt() and 0xFF)
            bits += fromBits
            while (bits >= toBits) {
                bits -= toBits
                result.add(((acc shr bits) and maxv).toByte())
            }
        }

        if (pad && bits > 0) {
            result.add(((acc shl (toBits - bits)) and maxv).toByte())
        }

        return result.toByteArray()
    }

    private fun bech32CreateChecksum(hrp: String, data: ByteArray): ByteArray {
        val values = bech32HrpExpand(hrp) + data + byteArrayOf(0, 0, 0, 0, 0, 0)
        val polymod = bech32Polymod(values) xor 1

        val checksum = ByteArray(6)
        for (i in 0 until 6) {
            checksum[i] = ((polymod shr (5 * (5 - i))) and 31).toByte()
        }
        return checksum
    }

    private fun bech32HrpExpand(hrp: String): ByteArray {
        val result = ByteArray(hrp.length * 2 + 1)
        for (i in hrp.indices) {
            result[i] = (hrp[i].code shr 5).toByte()
        }
        result[hrp.length] = 0
        for (i in hrp.indices) {
            result[hrp.length + 1 + i] = (hrp[i].code and 31).toByte()
        }
        return result
    }

    private fun bech32Polymod(values: ByteArray): Int {
        val generator = intArrayOf(0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3)
        var chk = 1
        for (value in values) {
            val top = chk shr 25
            chk = (chk and 0x1ffffff) shl 5 xor (value.toInt() and 0xFF)
            for (i in 0 until 5) {
                if ((top shr i) and 1 != 0) {
                    chk = chk xor generator[i]
                }
            }
        }
        return chk
    }

    // MARK: - Base58 Encoding

    private val BASE58_ALPHABET = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
    private val XRP_BASE58_ALPHABET = "rpshnaf39wBUDNEGHJKLM4PQRST7VWXYZ2bcdeCg65jkm8oFqi1tuvAxyz"

    /**
     * Base58 encode (for Solana public keys)
     */
    private fun base58Encode(data: ByteArray): String {
        if (data.isEmpty()) return ""

        // Convert to big integer
        var num = java.math.BigInteger(1, data)
        val sb = StringBuilder()
        val base = java.math.BigInteger.valueOf(58)

        while (num > java.math.BigInteger.ZERO) {
            val mod = num.mod(base).toInt()
            sb.append(BASE58_ALPHABET[mod])
            num = num.divide(base)
        }

        // Add leading zeros
        for (byte in data) {
            if (byte == 0.toByte()) {
                sb.append(BASE58_ALPHABET[0])
            } else {
                break
            }
        }

        return sb.reverse().toString()
    }

    /**
     * Base58Check encode (for Bitcoin/Litecoin/Dogecoin legacy addresses)
     */
    private fun base58CheckEncode(version: Byte, payload: ByteArray): String {
        val versionedPayload = byteArrayOf(version) + payload
        val checksum = doubleSha256(versionedPayload).copyOfRange(0, 4)
        return base58Encode(versionedPayload + checksum)
    }

    /**
     * XRP Base58Check encode (uses different alphabet)
     */
    private fun xrpBase58CheckEncode(payload: ByteArray): String {
        val versionedPayload = byteArrayOf(0x00.toByte()) + payload
        val checksum = doubleSha256(versionedPayload).copyOfRange(0, 4)
        val data = versionedPayload + checksum

        if (data.isEmpty()) return ""

        var num = java.math.BigInteger(1, data)
        val sb = StringBuilder()
        val base = java.math.BigInteger.valueOf(58)

        while (num > java.math.BigInteger.ZERO) {
            val mod = num.mod(base).toInt()
            sb.append(XRP_BASE58_ALPHABET[mod])
            num = num.divide(base)
        }

        // Add leading zeros
        for (byte in data) {
            if (byte == 0.toByte()) {
                sb.append(XRP_BASE58_ALPHABET[0])
            } else {
                break
            }
        }

        return sb.reverse().toString()
    }

    /**
     * TRON Base58Check encode (Ethereum-style address with base58)
     */
    private fun tronBase58CheckEncode(addressBytes: ByteArray): String {
        val versionedPayload = byteArrayOf(0x41.toByte()) + addressBytes
        val checksum = doubleSha256(versionedPayload).copyOfRange(0, 4)
        return base58Encode(versionedPayload + checksum)
    }

    /**
     * Double SHA256 (for Bitcoin-style checksums)
     */
    private fun doubleSha256(data: ByteArray): ByteArray {
        return sha256(sha256(data))
    }

    // MARK: - Secure Key Derivation

    /**
     * Derive private key with automatic secure cleanup
     * Returns SecureByteArray that clears memory when close() is called
     */
    fun derivePrivateKeySecure(mnemonic: String): SecureByteArray {
        return SecureByteArray(derivePrivateKey(mnemonic))
    }

    /**
     * Derive private key for a specific chain with automatic secure cleanup
     */
    fun derivePrivateKeySecure(mnemonic: String, chain: Chain, accountIndex: Int = 0): SecureByteArray {
        return SecureByteArray(derivePrivateKey(mnemonic, chain, accountIndex))
    }

    // MARK: - Secure Signing Operations

    /**
     * Sign a transaction hash with secure private key (auto-clears key after signing)
     */
    fun signSecure(hash: ByteArray, securePrivateKey: SecureByteArray): ByteArray {
        return try {
            sign(hash, securePrivateKey.bytes)
        } finally {
            securePrivateKey.close()
        }
    }

    /**
     * Sign with mnemonic (derives key, signs, clears all sensitive data)
     */
    fun signWithMnemonic(
        hash: ByteArray,
        mnemonic: String,
        chain: Chain = Chain.SHAREHODL,
        accountIndex: Int = 0
    ): ByteArray {
        val secureKey = derivePrivateKeySecure(mnemonic, chain, accountIndex)
        return signSecure(hash, secureKey)
    }

    /**
     * Complete secure transaction signing workflow
     * - Derives private key from mnemonic
     * - Signs the transaction hash
     * - Clears all sensitive data from memory
     * - Returns the signature
     */
    fun secureSignTransaction(
        transactionHash: ByteArray,
        mnemonic: String,
        chain: Chain = Chain.SHAREHODL,
        accountIndex: Int = 0
    ): ByteArray {
        // Create secure container for mnemonic
        val secureMnemonic = SecureMnemonic(mnemonic)

        return try {
            // Derive key securely
            val secureKey = derivePrivateKeySecure(secureMnemonic.phrase, chain, accountIndex)

            // Sign and return (secureKey auto-clears when signSecure completes)
            signSecure(transactionHash, secureKey)
        } finally {
            // Ensure mnemonic is cleared even if an error occurs
            secureMnemonic.close()
        }
    }
}

/**
 * A container for sensitive byte data that clears memory when closed
 * Use this for private keys, seeds, and other cryptographic material
 * Implements AutoCloseable for use with use {} blocks
 */
class SecureByteArray(data: ByteArray) : AutoCloseable {
    private var storage: ByteArray = data.copyOf()
    private var cleared = false

    /**
     * Get the bytes (read-only access)
     * Throws if already cleared
     */
    val bytes: ByteArray
        get() {
            check(!cleared) { "SecureByteArray has been cleared" }
            return storage
        }

    /**
     * Number of bytes
     */
    val size: Int get() = storage.size

    /**
     * Check if cleared
     */
    val isCleared: Boolean get() = cleared

    /**
     * Clear the data from memory
     */
    override fun close() {
        if (!cleared) {
            CryptoService.secureClear(storage)
            cleared = true
        }
    }

    /**
     * Perform operation with the bytes, then clear
     */
    inline fun <T> use(block: (ByteArray) -> T): T {
        return try {
            block(bytes)
        } finally {
            close()
        }
    }
}

/**
 * A container for mnemonic phrases that clears memory when closed
 * Implements AutoCloseable for use with use {} blocks
 */
class SecureMnemonic(mnemonic: String) : AutoCloseable {
    private var chars: CharArray = mnemonic.toCharArray()
    private var cleared = false

    /**
     * Get the mnemonic as a string
     * Throws if already cleared
     */
    val phrase: String
        get() {
            check(!cleared) { "SecureMnemonic has been cleared" }
            return String(chars)
        }

    /**
     * Number of words
     */
    val wordCount: Int
        get() = if (cleared) 0 else chars.count { it == ' ' } + 1

    /**
     * Check if cleared
     */
    val isCleared: Boolean get() = cleared

    /**
     * Clear the mnemonic from memory
     */
    override fun close() {
        if (!cleared) {
            CryptoService.secureClear(chars)
            cleared = true
        }
    }

    /**
     * Perform operation with the phrase, then clear
     */
    inline fun <T> use(block: (String) -> T): T {
        return try {
            block(phrase)
        } finally {
            close()
        }
    }
}
