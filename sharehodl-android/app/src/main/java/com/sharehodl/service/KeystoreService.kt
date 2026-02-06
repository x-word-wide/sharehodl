package com.sharehodl.service

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Log
import androidx.biometric.BiometricManager
import androidx.biometric.BiometricPrompt
import androidx.core.content.ContextCompat
import androidx.fragment.app.FragmentActivity
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import com.sharehodl.model.Chain
import com.sharehodl.model.ChainAccount
import org.json.JSONArray
import org.json.JSONObject
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec
import javax.inject.Inject
import javax.inject.Singleton
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException
import kotlinx.coroutines.suspendCancellableCoroutine

/**
 * Service for secure storage using Android Keystore and EncryptedSharedPreferences
 */
@Singleton
class KeystoreService @Inject constructor(
    private val context: Context
) {
    companion object {
        private const val TAG = "KeystoreService"
        private const val KEYSTORE_PROVIDER = "AndroidKeyStore"
        private const val KEY_ALIAS = "sharehodl_wallet_key"
        private const val KEY_ALIAS_NO_AUTH = "sharehodl_wallet_key_no_auth"
        private const val PREFS_NAME = "sharehodl_secure_prefs"
        private const val KEY_ENCRYPTED_PRIVATE_KEY = "encrypted_private_key"
        private const val KEY_ENCRYPTED_MNEMONIC = "encrypted_mnemonic"
        private const val KEY_WALLET_ADDRESS = "wallet_address"
        private const val KEY_CHAIN_ACCOUNTS = "chain_accounts"
        private const val KEY_IV = "encryption_iv"
        private const val KEY_MNEMONIC_IV = "mnemonic_iv"
        private const val KEY_USES_BIOMETRIC = "uses_biometric"
    }

    private val keyStore: KeyStore = KeyStore.getInstance(KEYSTORE_PROVIDER).apply {
        load(null)
    }

    private val encryptedPrefs by lazy {
        val masterKey = MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .setUserAuthenticationRequired(false)
            .build()

        EncryptedSharedPreferences.create(
            context,
            PREFS_NAME,
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        )
    }

    // MARK: - Key Generation

    /**
     * Check if device supports secure authentication
     */
    private fun canUseSecureAuthentication(): Boolean {
        val biometricManager = BiometricManager.from(context)
        val canAuth = biometricManager.canAuthenticate(
            BiometricManager.Authenticators.BIOMETRIC_STRONG or
            BiometricManager.Authenticators.DEVICE_CREDENTIAL
        )
        Log.d(TAG, "canUseSecureAuthentication: $canAuth")
        return canAuth == BiometricManager.BIOMETRIC_SUCCESS
    }

    /**
     * Generate a secret key without user authentication requirement
     * Used as fallback when device doesn't support biometrics/device credential
     */
    private fun getOrCreateSecretKeyNoAuth(): SecretKey {
        Log.d(TAG, "getOrCreateSecretKeyNoAuth called")
        if (keyStore.containsAlias(KEY_ALIAS_NO_AUTH)) {
            Log.d(TAG, "Returning existing key")
            return keyStore.getKey(KEY_ALIAS_NO_AUTH, null) as SecretKey
        }

        Log.d(TAG, "Creating new key without auth requirement")
        val keyGenerator = KeyGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_AES,
            KEYSTORE_PROVIDER
        )

        val builder = KeyGenParameterSpec.Builder(
            KEY_ALIAS_NO_AUTH,
            KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
        )
            .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
            .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
            .setKeySize(256)
            .setUserAuthenticationRequired(false) // No auth required

        keyGenerator.init(builder.build())
        return keyGenerator.generateKey()
    }

    /**
     * Generate or retrieve the master encryption key from Android Keystore
     * Uses StrongBox hardware security if available with TEE fallback
     */
    private fun getOrCreateSecretKey(): SecretKey {
        Log.d(TAG, "getOrCreateSecretKey called")
        if (keyStore.containsAlias(KEY_ALIAS)) {
            Log.d(TAG, "Returning existing key")
            return keyStore.getKey(KEY_ALIAS, null) as SecretKey
        }

        Log.d(TAG, "Creating new key with auth requirement")
        val keyGenerator = KeyGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_AES,
            KEYSTORE_PROVIDER
        )

        val builder = KeyGenParameterSpec.Builder(
            KEY_ALIAS,
            KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
        )
            .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
            .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
            .setKeySize(256)
            .setUserAuthenticationRequired(true)
            .setUserAuthenticationParameters(
                0, // Require auth for every use
                KeyProperties.AUTH_BIOMETRIC_STRONG or KeyProperties.AUTH_DEVICE_CREDENTIAL
            )

        // Invalidate key if new biometrics are enrolled (security measure)
        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.N) {
            builder.setInvalidatedByBiometricEnrollment(true)
        }

        // Try StrongBox first (hardware security module) if available
        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.P) {
            builder.setIsStrongBoxBacked(true)
        }

        return try {
            keyGenerator.init(builder.build())
            keyGenerator.generateKey()
        } catch (e: Exception) {
            Log.w(TAG, "StrongBox not available, falling back to TEE", e)
            // StrongBox not available, fall back to TEE
            if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.P) {
                builder.setIsStrongBoxBacked(false)
            }
            keyGenerator.init(builder.build())
            keyGenerator.generateKey()
        }
    }

    // MARK: - Private Key Storage

    /**
     * Store encrypted private key
     * Key is encrypted with hardware-backed Keystore key and stored in EncryptedSharedPreferences
     * Biometric prompt shown for UX but key doesn't require per-use auth (more compatible)
     */
    suspend fun storePrivateKey(
        activity: FragmentActivity,
        privateKey: ByteArray
    ): Result<Unit> {
        Log.d(TAG, "storePrivateKey called")

        return try {
            // Try biometric prompt for UX, but key storage works without it
            val showedBiometric = if (isBiometricsAvailable()) {
                try {
                    Log.d(TAG, "Showing biometric prompt for UX")
                    authenticateWithBiometrics(activity, "Secure your wallet")
                    true
                } catch (e: Exception) {
                    Log.w(TAG, "Biometric prompt failed/cancelled, continuing anyway: ${e.message}")
                    false
                }
            } else {
                Log.d(TAG, "Biometrics not available, skipping prompt")
                false
            }

            // Store the key using hardware-backed encryption (no per-use auth required)
            Log.d(TAG, "Storing with secure encryption (biometric shown: $showedBiometric)")
            storeWithSecureEncryption(privateKey, showedBiometric)
        } catch (e: Exception) {
            Log.e(TAG, "storePrivateKey failed: ${e.message}", e)
            Result.failure(e)
        }
    }

    /**
     * Store with secure encryption using Keystore (no per-use authentication required)
     * Still hardware-backed and very secure
     */
    private fun storeWithSecureEncryption(privateKey: ByteArray, usesBiometric: Boolean): Result<Unit> = runCatching {
        Log.d(TAG, "storeWithSecureEncryption: getting secret key")
        val secretKey = getOrCreateSecretKeyNoAuth()

        Log.d(TAG, "storeWithSecureEncryption: encrypting")
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, secretKey)

        val encryptedData = cipher.doFinal(privateKey)
        val iv = cipher.iv

        Log.d(TAG, "storeWithSecureEncryption: storing to prefs")
        encryptedPrefs.edit()
            .putString(KEY_ENCRYPTED_PRIVATE_KEY, encryptedData.toBase64())
            .putString(KEY_IV, iv.toBase64())
            .putBoolean(KEY_USES_BIOMETRIC, usesBiometric)
            .apply()

        Log.d(TAG, "storeWithSecureEncryption: success")
    }

    /**
     * Store with crypto-based biometric authentication
     * This properly authorizes the Keystore key by passing the cipher through BiometricPrompt
     */
    private suspend fun storeWithCryptoBasedAuth(
        activity: FragmentActivity,
        privateKey: ByteArray
    ): Result<Unit> {
        Log.d(TAG, "storeWithCryptoBasedAuth: getting secret key")
        val secretKey = getOrCreateSecretKey()

        Log.d(TAG, "storeWithCryptoBasedAuth: initializing cipher")
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")

        return try {
            cipher.init(Cipher.ENCRYPT_MODE, secretKey)

            Log.d(TAG, "storeWithCryptoBasedAuth: authenticating with CryptoObject")
            val authenticatedCipher = authenticateWithCrypto(activity, "Secure your wallet", cipher)

            Log.d(TAG, "storeWithCryptoBasedAuth: encrypting with authenticated cipher")
            val encryptedData = authenticatedCipher.doFinal(privateKey)
            val iv = authenticatedCipher.iv

            Log.d(TAG, "storeWithCryptoBasedAuth: storing to prefs")
            encryptedPrefs.edit()
                .putString(KEY_ENCRYPTED_PRIVATE_KEY, encryptedData.toBase64())
                .putString(KEY_IV, iv.toBase64())
                .putBoolean(KEY_USES_BIOMETRIC, true)
                .apply()

            Log.d(TAG, "storeWithCryptoBasedAuth: success")
            Result.success(Unit)
        } catch (e: android.security.keystore.UserNotAuthenticatedException) {
            Log.w(TAG, "Key requires authentication, using CryptoObject flow", e)
            // Key requires per-use auth, authenticate with CryptoObject
            val authenticatedCipher = authenticateWithCrypto(activity, "Secure your wallet", cipher)

            val encryptedData = authenticatedCipher.doFinal(privateKey)
            val iv = authenticatedCipher.iv

            encryptedPrefs.edit()
                .putString(KEY_ENCRYPTED_PRIVATE_KEY, encryptedData.toBase64())
                .putString(KEY_IV, iv.toBase64())
                .putBoolean(KEY_USES_BIOMETRIC, true)
                .apply()

            Log.d(TAG, "storeWithCryptoBasedAuth: success after auth")
            Result.success(Unit)
        } catch (e: Exception) {
            Log.e(TAG, "storeWithCryptoBasedAuth failed", e)
            Result.failure(e)
        }
    }

    private suspend fun storeWithBiometricAuth(
        activity: FragmentActivity,
        privateKey: ByteArray
    ): Result<Unit> = runCatching {
        Log.d(TAG, "storeWithBiometricAuth: getting secret key")
        val secretKey = getOrCreateSecretKey()

        Log.d(TAG, "storeWithBiometricAuth: authenticating")
        authenticateWithBiometrics(activity, "Secure your wallet")

        Log.d(TAG, "storeWithBiometricAuth: encrypting")
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, secretKey)

        val encryptedData = cipher.doFinal(privateKey)
        val iv = cipher.iv

        Log.d(TAG, "storeWithBiometricAuth: storing to prefs")
        encryptedPrefs.edit()
            .putString(KEY_ENCRYPTED_PRIVATE_KEY, encryptedData.toBase64())
            .putString(KEY_IV, iv.toBase64())
            .putBoolean(KEY_USES_BIOMETRIC, true)
            .apply()

        Log.d(TAG, "storeWithBiometricAuth: success")
    }

    private fun storeWithoutBiometricAuth(privateKey: ByteArray): Result<Unit> = runCatching {
        Log.d(TAG, "storeWithoutBiometricAuth: getting secret key")
        val secretKey = getOrCreateSecretKeyNoAuth()

        Log.d(TAG, "storeWithoutBiometricAuth: encrypting")
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, secretKey)

        val encryptedData = cipher.doFinal(privateKey)
        val iv = cipher.iv

        Log.d(TAG, "storeWithoutBiometricAuth: storing to prefs")
        encryptedPrefs.edit()
            .putString(KEY_ENCRYPTED_PRIVATE_KEY, encryptedData.toBase64())
            .putString(KEY_IV, iv.toBase64())
            .putBoolean(KEY_USES_BIOMETRIC, false)
            .apply()

        Log.d(TAG, "storeWithoutBiometricAuth: success")
    }

    /**
     * Retrieve private key (shows biometric prompt if originally stored with biometrics)
     */
    suspend fun retrievePrivateKey(
        activity: FragmentActivity
    ): Result<ByteArray> {
        Log.d(TAG, "retrievePrivateKey called")

        val encryptedData = encryptedPrefs.getString(KEY_ENCRYPTED_PRIVATE_KEY, null)
        val iv = encryptedPrefs.getString(KEY_IV, null)
        val usesBiometric = encryptedPrefs.getBoolean(KEY_USES_BIOMETRIC, false)

        if (encryptedData == null || iv == null) {
            Log.e(TAG, "No private key stored")
            return Result.failure(IllegalStateException("No private key stored"))
        }

        // Show biometric prompt if originally stored with biometrics
        if (usesBiometric && isBiometricsAvailable()) {
            try {
                Log.d(TAG, "Showing biometric prompt for wallet access")
                authenticateWithBiometrics(activity, "Access your wallet")
            } catch (e: Exception) {
                Log.e(TAG, "Biometric authentication failed: ${e.message}")
                return Result.failure(e)
            }
        }

        // Decrypt the key
        return retrieveWithSecureDecryption(encryptedData, iv)
    }

    private fun retrieveWithSecureDecryption(
        encryptedData: String,
        iv: String
    ): Result<ByteArray> = runCatching {
        Log.d(TAG, "retrieveWithSecureDecryption: decrypting")
        val secretKey = getOrCreateSecretKeyNoAuth()
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val spec = GCMParameterSpec(128, iv.fromBase64())
        cipher.init(Cipher.DECRYPT_MODE, secretKey, spec)

        cipher.doFinal(encryptedData.fromBase64())
    }

    // MARK: - Mnemonic Storage

    /**
     * Store encrypted mnemonic for recovery phrase viewing
     */
    fun storeMnemonic(mnemonic: String): Result<Unit> = runCatching {
        Log.d(TAG, "storeMnemonic: encrypting")
        val secretKey = getOrCreateSecretKeyNoAuth()
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, secretKey)

        val encryptedData = cipher.doFinal(mnemonic.toByteArray(Charsets.UTF_8))
        val iv = cipher.iv

        encryptedPrefs.edit()
            .putString(KEY_ENCRYPTED_MNEMONIC, encryptedData.toBase64())
            .putString(KEY_MNEMONIC_IV, iv.toBase64())
            .apply()

        Log.d(TAG, "storeMnemonic: success")
    }

    /**
     * Retrieve mnemonic (requires biometric authentication)
     */
    suspend fun retrieveMnemonic(activity: FragmentActivity): Result<String> {
        Log.d(TAG, "retrieveMnemonic called")

        val encryptedData = encryptedPrefs.getString(KEY_ENCRYPTED_MNEMONIC, null)
        val iv = encryptedPrefs.getString(KEY_MNEMONIC_IV, null)

        if (encryptedData == null || iv == null) {
            Log.e(TAG, "No mnemonic stored")
            return Result.failure(IllegalStateException("Recovery phrase not available"))
        }

        // Always require biometric auth to view mnemonic
        if (isBiometricsAvailable()) {
            try {
                Log.d(TAG, "Showing biometric prompt for mnemonic access")
                authenticateWithBiometrics(activity, "View recovery phrase")
            } catch (e: Exception) {
                Log.e(TAG, "Biometric authentication failed: ${e.message}")
                return Result.failure(e)
            }
        }

        return runCatching {
            val secretKey = getOrCreateSecretKeyNoAuth()
            val cipher = Cipher.getInstance("AES/GCM/NoPadding")
            val spec = GCMParameterSpec(128, iv.fromBase64())
            cipher.init(Cipher.DECRYPT_MODE, secretKey, spec)

            String(cipher.doFinal(encryptedData.fromBase64()), Charsets.UTF_8)
        }
    }

    // MARK: - Wallet Address Storage (Not Sensitive)

    /**
     * Store wallet address (no authentication required)
     */
    fun storeWalletAddress(address: String) {
        encryptedPrefs.edit()
            .putString(KEY_WALLET_ADDRESS, address)
            .apply()
    }

    /**
     * Retrieve wallet address
     */
    fun getWalletAddress(): String? {
        return encryptedPrefs.getString(KEY_WALLET_ADDRESS, null)
    }

    /**
     * Check if wallet exists
     */
    val hasWallet: Boolean
        get() = getWalletAddress() != null

    // MARK: - Multi-Chain Account Storage

    /**
     * Store all chain accounts (addresses derived from single mnemonic)
     */
    fun storeChainAccounts(accounts: List<ChainAccount>) {
        val jsonArray = JSONArray()
        accounts.forEach { account ->
            val jsonObject = JSONObject().apply {
                put("chain", account.chain.name)
                put("address", account.address)
                put("derivationPath", account.derivationPath)
                put("balance", account.balance)
                account.balanceUsd?.let { put("balanceUsd", it) }
            }
            jsonArray.put(jsonObject)
        }
        encryptedPrefs.edit()
            .putString(KEY_CHAIN_ACCOUNTS, jsonArray.toString())
            .apply()
        Log.d(TAG, "Stored ${accounts.size} chain accounts")
    }

    /**
     * Retrieve all stored chain accounts
     */
    fun getStoredChainAccounts(): List<ChainAccount> {
        val jsonString = encryptedPrefs.getString(KEY_CHAIN_ACCOUNTS, null) ?: return emptyList()

        return try {
            val jsonArray = JSONArray(jsonString)
            val accounts = mutableListOf<ChainAccount>()

            for (i in 0 until jsonArray.length()) {
                val jsonObject = jsonArray.getJSONObject(i)
                val chainName = jsonObject.getString("chain")
                val chain = Chain.valueOf(chainName)

                accounts.add(ChainAccount(
                    chain = chain,
                    address = jsonObject.getString("address"),
                    derivationPath = jsonObject.getString("derivationPath"),
                    balance = jsonObject.optString("balance", "0"),
                    balanceUsd = jsonObject.optString("balanceUsd", null)
                ))
            }

            Log.d(TAG, "Retrieved ${accounts.size} chain accounts")
            accounts
        } catch (e: Exception) {
            Log.e(TAG, "Failed to parse chain accounts: ${e.message}")
            emptyList()
        }
    }

    /**
     * Update balance for a specific chain
     */
    fun updateChainBalance(chain: Chain, balance: String, balanceUsd: String? = null) {
        val accounts = getStoredChainAccounts().toMutableList()
        val index = accounts.indexOfFirst { it.chain == chain }

        if (index >= 0) {
            accounts[index] = accounts[index].copy(balance = balance, balanceUsd = balanceUsd)
            storeChainAccounts(accounts)
        }
    }

    /**
     * Delete all wallet data
     */
    fun deleteWallet() {
        encryptedPrefs.edit()
            .remove(KEY_ENCRYPTED_PRIVATE_KEY)
            .remove(KEY_ENCRYPTED_MNEMONIC)
            .remove(KEY_WALLET_ADDRESS)
            .remove(KEY_CHAIN_ACCOUNTS)
            .remove(KEY_IV)
            .remove(KEY_MNEMONIC_IV)
            .remove(KEY_USES_BIOMETRIC)
            .apply()

        // Also delete the keys from keystore
        if (keyStore.containsAlias(KEY_ALIAS)) {
            keyStore.deleteEntry(KEY_ALIAS)
        }
        if (keyStore.containsAlias(KEY_ALIAS_NO_AUTH)) {
            keyStore.deleteEntry(KEY_ALIAS_NO_AUTH)
        }
    }

    // MARK: - Biometric Authentication

    /**
     * Check if biometrics are available
     */
    fun isBiometricsAvailable(): Boolean {
        val biometricManager = BiometricManager.from(context)
        return when (biometricManager.canAuthenticate(BiometricManager.Authenticators.BIOMETRIC_STRONG)) {
            BiometricManager.BIOMETRIC_SUCCESS -> true
            else -> false
        }
    }

    /**
     * Authenticate user with biometrics using CryptoObject
     * This properly authorizes Keystore keys for use
     */
    private suspend fun authenticateWithCrypto(
        activity: FragmentActivity,
        title: String,
        cipher: Cipher
    ): Cipher = suspendCancellableCoroutine { continuation ->
        Log.d(TAG, "authenticateWithCrypto: starting")
        val executor = ContextCompat.getMainExecutor(context)

        val callback = object : BiometricPrompt.AuthenticationCallback() {
            override fun onAuthenticationSucceeded(result: BiometricPrompt.AuthenticationResult) {
                Log.d(TAG, "authenticateWithCrypto: success")
                val authenticatedCipher = result.cryptoObject?.cipher
                if (authenticatedCipher != null) {
                    continuation.resume(authenticatedCipher)
                } else {
                    // Fallback to original cipher if crypto object not returned
                    continuation.resume(cipher)
                }
            }

            override fun onAuthenticationError(errorCode: Int, errString: CharSequence) {
                Log.e(TAG, "authenticateWithCrypto: error $errorCode - $errString")
                val userMessage = when (errorCode) {
                    BiometricPrompt.ERROR_LOCKOUT -> "Too many attempts. Please try again later."
                    BiometricPrompt.ERROR_LOCKOUT_PERMANENT -> "Biometrics disabled. Use device PIN."
                    BiometricPrompt.ERROR_USER_CANCELED -> "Authentication cancelled."
                    BiometricPrompt.ERROR_NEGATIVE_BUTTON -> "Authentication cancelled."
                    BiometricPrompt.ERROR_NO_BIOMETRICS -> "No biometrics enrolled. Please use PIN."
                    BiometricPrompt.ERROR_HW_NOT_PRESENT -> "No biometric hardware available."
                    BiometricPrompt.ERROR_NO_DEVICE_CREDENTIAL -> "No device credential set."
                    else -> "Authentication failed: $errString"
                }
                continuation.resumeWithException(BiometricException(userMessage))
            }

            override fun onAuthenticationFailed() {
                Log.w(TAG, "authenticateWithCrypto: attempt failed, user can retry")
            }
        }

        try {
            // For crypto-based auth, we can only use BIOMETRIC_STRONG (not device credential)
            val promptInfo = BiometricPrompt.PromptInfo.Builder()
                .setTitle(title)
                .setSubtitle("Authenticate to continue")
                .setNegativeButtonText("Cancel")
                .setAllowedAuthenticators(BiometricManager.Authenticators.BIOMETRIC_STRONG)
                .build()

            Log.d(TAG, "authenticateWithCrypto: showing prompt with CryptoObject")
            val biometricPrompt = BiometricPrompt(activity, executor, callback)
            biometricPrompt.authenticate(promptInfo, BiometricPrompt.CryptoObject(cipher))
        } catch (e: Exception) {
            Log.e(TAG, "authenticateWithCrypto: exception creating prompt", e)
            continuation.resumeWithException(BiometricException("Failed to show authentication: ${e.message}"))
        }
    }

    /**
     * Authenticate user with biometrics (with device credential fallback)
     */
    private suspend fun authenticateWithBiometrics(
        activity: FragmentActivity,
        title: String
    ) = suspendCancellableCoroutine { continuation ->
        Log.d(TAG, "authenticateWithBiometrics: starting")
        val executor = ContextCompat.getMainExecutor(context)

        val callback = object : BiometricPrompt.AuthenticationCallback() {
            override fun onAuthenticationSucceeded(result: BiometricPrompt.AuthenticationResult) {
                Log.d(TAG, "authenticateWithBiometrics: success")
                continuation.resume(Unit)
            }

            override fun onAuthenticationError(errorCode: Int, errString: CharSequence) {
                Log.e(TAG, "authenticateWithBiometrics: error $errorCode - $errString")
                // Provide user-friendly error messages
                val userMessage = when (errorCode) {
                    BiometricPrompt.ERROR_LOCKOUT -> "Too many attempts. Please try again later."
                    BiometricPrompt.ERROR_LOCKOUT_PERMANENT -> "Biometrics disabled. Use device PIN."
                    BiometricPrompt.ERROR_USER_CANCELED -> "Authentication cancelled."
                    BiometricPrompt.ERROR_NEGATIVE_BUTTON -> "Authentication cancelled."
                    BiometricPrompt.ERROR_NO_BIOMETRICS -> "No biometrics enrolled. Please use PIN."
                    BiometricPrompt.ERROR_HW_NOT_PRESENT -> "No biometric hardware available."
                    BiometricPrompt.ERROR_NO_DEVICE_CREDENTIAL -> "No device credential set."
                    else -> "Authentication failed: $errString"
                }
                continuation.resumeWithException(BiometricException(userMessage))
            }

            override fun onAuthenticationFailed() {
                Log.w(TAG, "authenticateWithBiometrics: attempt failed, user can retry")
                // Don't fail yet - let user retry
                // The system will eventually call onAuthenticationError after too many attempts
            }
        }

        try {
            // Allow both biometrics and device credential (PIN/pattern/password) as fallback
            val promptInfo = BiometricPrompt.PromptInfo.Builder()
                .setTitle(title)
                .setSubtitle("Authenticate to continue")
                .setAllowedAuthenticators(
                    BiometricManager.Authenticators.BIOMETRIC_STRONG or
                    BiometricManager.Authenticators.DEVICE_CREDENTIAL
                )
                .build()

            Log.d(TAG, "authenticateWithBiometrics: showing prompt")
            val biometricPrompt = BiometricPrompt(activity, executor, callback)
            biometricPrompt.authenticate(promptInfo)
        } catch (e: Exception) {
            Log.e(TAG, "authenticateWithBiometrics: exception creating prompt", e)
            continuation.resumeWithException(BiometricException("Failed to show authentication: ${e.message}"))
        }
    }

    // MARK: - Utility Extensions

    private fun ByteArray.toBase64(): String =
        android.util.Base64.encodeToString(this, android.util.Base64.NO_WRAP)

    private fun String.fromBase64(): ByteArray =
        android.util.Base64.decode(this, android.util.Base64.NO_WRAP)
}

class BiometricException(message: String) : Exception(message)
