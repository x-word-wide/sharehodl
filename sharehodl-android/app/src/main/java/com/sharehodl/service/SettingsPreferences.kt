package com.sharehodl.service

import android.content.Context
import android.content.SharedPreferences
import androidx.core.content.edit
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Manages user preferences and settings
 */
@Singleton
class SettingsPreferences @Inject constructor(
    private val context: Context
) {
    private val prefs: SharedPreferences by lazy {
        context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
    }

    companion object {
        private const val PREFS_NAME = "sharehodl_settings"

        // Trading Preferences
        private const val KEY_PRICE_ALERTS_ENABLED = "price_alerts_enabled"
        private const val KEY_DEFAULT_ORDER_TYPE = "default_order_type"
        private const val KEY_TRADE_CONFIRMATIONS = "trade_confirmations"

        // Security
        private const val KEY_BIOMETRIC_ENABLED = "biometric_enabled"
        private const val KEY_PIN_HASH = "pin_hash"
        private const val KEY_2FA_ENABLED = "2fa_enabled"
        private const val KEY_2FA_SECRET = "2fa_secret"

        // Network
        private const val KEY_NETWORK = "network"
        private const val KEY_CUSTOM_RPC = "custom_rpc"
    }

    // ============================================
    // Trading Preferences
    // ============================================

    var priceAlertsEnabled: Boolean
        get() = prefs.getBoolean(KEY_PRICE_ALERTS_ENABLED, false)
        set(value) = prefs.edit { putBoolean(KEY_PRICE_ALERTS_ENABLED, value) }

    var defaultOrderType: OrderType
        get() = OrderType.fromString(prefs.getString(KEY_DEFAULT_ORDER_TYPE, OrderType.MARKET.name) ?: OrderType.MARKET.name)
        set(value) = prefs.edit { putString(KEY_DEFAULT_ORDER_TYPE, value.name) }

    var tradeConfirmationsEnabled: Boolean
        get() = prefs.getBoolean(KEY_TRADE_CONFIRMATIONS, true)
        set(value) = prefs.edit { putBoolean(KEY_TRADE_CONFIRMATIONS, value) }

    // ============================================
    // Security Settings
    // ============================================

    var biometricEnabled: Boolean
        get() = prefs.getBoolean(KEY_BIOMETRIC_ENABLED, false)
        set(value) = prefs.edit { putBoolean(KEY_BIOMETRIC_ENABLED, value) }

    var pinHash: String?
        get() = prefs.getString(KEY_PIN_HASH, null)
        set(value) = prefs.edit { putString(KEY_PIN_HASH, value) }

    var twoFactorEnabled: Boolean
        get() = prefs.getBoolean(KEY_2FA_ENABLED, false)
        set(value) = prefs.edit { putBoolean(KEY_2FA_ENABLED, value) }

    var twoFactorSecret: String?
        get() = prefs.getString(KEY_2FA_SECRET, null)
        set(value) = prefs.edit { putString(KEY_2FA_SECRET, value) }

    // ============================================
    // Network Settings
    // ============================================

    var network: NetworkType
        get() = NetworkType.fromString(prefs.getString(KEY_NETWORK, NetworkType.MAINNET.name) ?: NetworkType.MAINNET.name)
        set(value) = prefs.edit { putString(KEY_NETWORK, value.name) }

    var customRpcEndpoint: String?
        get() = prefs.getString(KEY_CUSTOM_RPC, null)
        set(value) = prefs.edit { putString(KEY_CUSTOM_RPC, value) }

    // ============================================
    // PIN Management
    // ============================================

    fun hashPin(pin: String): String {
        val bytes = pin.toByteArray()
        val digest = java.security.MessageDigest.getInstance("SHA-256")
        val hashBytes = digest.digest(bytes)
        return hashBytes.joinToString("") { "%02x".format(it) }
    }

    fun verifyPin(pin: String): Boolean {
        val storedHash = pinHash ?: return false
        return hashPin(pin) == storedHash
    }

    fun setPin(newPin: String) {
        pinHash = hashPin(newPin)
    }

    fun hasPin(): Boolean = pinHash != null

    // ============================================
    // Clear All Settings
    // ============================================

    fun clearAll() {
        prefs.edit { clear() }
    }
}

/**
 * Order types for trading
 */
enum class OrderType(val displayName: String) {
    MARKET("Market Order"),
    LIMIT("Limit Order"),
    STOP_LOSS("Stop Loss"),
    STOP_LIMIT("Stop Limit");

    companion object {
        fun fromString(value: String): OrderType {
            return entries.find { it.name == value } ?: MARKET
        }
    }
}

/**
 * Network environments
 */
enum class NetworkType(val displayName: String, val description: String) {
    MAINNET("ShareHODL Mainnet", "Production network"),
    TESTNET("ShareHODL Testnet", "Test network for development");

    companion object {
        fun fromString(value: String): NetworkType {
            return entries.find { it.name == value } ?: MAINNET
        }
    }
}
