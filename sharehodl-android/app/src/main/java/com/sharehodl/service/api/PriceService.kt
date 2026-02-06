package com.sharehodl.service.api

import com.sharehodl.config.NetworkConfig
import com.sharehodl.model.Chain
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Service for fetching cryptocurrency prices
 * Uses CoinGecko API with caching to minimize API calls
 */
@Singleton
class PriceService @Inject constructor() {

    companion object {
        private const val TIMEOUT_MS = 10000
    }

    // Cache for prices with timestamps
    private val priceCache = mutableMapOf<Chain, CachedPrice>()
    private val cacheMutex = Mutex()

    data class CachedPrice(
        val usdPrice: Double,
        val timestamp: Long
    )

    // ============================================
    // Price Fetching
    // ============================================

    /**
     * Get USD price for a specific chain
     */
    suspend fun getUsdPrice(chain: Chain): Double? {
        // Check cache first
        cacheMutex.withLock {
            val cached = priceCache[chain]
            if (cached != null && !isCacheExpired(cached.timestamp)) {
                return cached.usdPrice
            }
        }

        // Fetch fresh prices
        return fetchPrices(listOf(chain))[chain]
    }

    /**
     * Get USD prices for multiple chains
     */
    suspend fun getUsdPrices(chains: List<Chain>): Map<Chain, Double> {
        val result = mutableMapOf<Chain, Double>()
        val chainsToFetch = mutableListOf<Chain>()

        // Check cache for each chain
        cacheMutex.withLock {
            for (chain in chains) {
                val cached = priceCache[chain]
                if (cached != null && !isCacheExpired(cached.timestamp)) {
                    result[chain] = cached.usdPrice
                } else {
                    chainsToFetch.add(chain)
                }
            }
        }

        // Fetch missing prices
        if (chainsToFetch.isNotEmpty()) {
            val fetched = fetchPrices(chainsToFetch)
            result.putAll(fetched)
        }

        return result
    }

    /**
     * Calculate USD value for a balance
     */
    suspend fun calculateUsdValue(chain: Chain, balance: String): String? {
        val price = getUsdPrice(chain) ?: return null
        val amount = balance.replace(",", "").toDoubleOrNull() ?: return null
        val usdValue = amount * price
        return formatUsdValue(usdValue)
    }

    /**
     * Refresh all prices (call periodically or on user action)
     */
    suspend fun refreshAllPrices(): Map<Chain, Double> {
        return fetchPrices(Chain.majorChains)
    }

    // ============================================
    // CoinGecko API
    // ============================================

    private suspend fun fetchPrices(chains: List<Chain>): Map<Chain, Double> = withContext(Dispatchers.IO) {
        val result = mutableMapOf<Chain, Double>()

        try {
            // Build comma-separated list of CoinGecko IDs
            val ids = chains.mapNotNull { NetworkConfig.PriceApi.coinGeckoIds[it] }
                .joinToString(",")

            if (ids.isEmpty()) return@withContext result

            val url = "${NetworkConfig.PriceApi.COINGECKO_URL}/simple/price?ids=$ids&vs_currencies=usd"
            val response = httpGet(url)

            // Parse response and update cache
            val now = System.currentTimeMillis()
            for (chain in chains) {
                val geckoId = NetworkConfig.PriceApi.coinGeckoIds[chain] ?: continue
                val priceObj = response.optJSONObject(geckoId)
                val usdPrice = priceObj?.optDouble("usd") ?: continue

                result[chain] = usdPrice
                cacheMutex.withLock {
                    priceCache[chain] = CachedPrice(usdPrice, now)
                }
            }

            // Add ShareHODL with fixed price (stablecoin pegged to $1)
            if (chains.contains(Chain.SHAREHODL)) {
                result[Chain.SHAREHODL] = 1.0
                cacheMutex.withLock {
                    priceCache[Chain.SHAREHODL] = CachedPrice(1.0, now)
                }
            }

        } catch (e: Exception) {
            // Return cached values on error
            cacheMutex.withLock {
                for (chain in chains) {
                    priceCache[chain]?.let { cached ->
                        result[chain] = cached.usdPrice
                    }
                }
            }
        }

        result
    }

    // ============================================
    // Price Formatting
    // ============================================

    /**
     * Format USD value for display
     */
    fun formatUsdValue(value: Double): String {
        return when {
            value >= 1_000_000 -> String.format("$%.2fM", value / 1_000_000)
            value >= 1_000 -> String.format("$%,.2f", value)
            value >= 1 -> String.format("$%.2f", value)
            value >= 0.01 -> String.format("$%.4f", value)
            else -> String.format("$%.6f", value)
        }
    }

    /**
     * Format price change percentage
     */
    fun formatPriceChange(change: Double): String {
        val sign = if (change >= 0) "+" else ""
        return "$sign${String.format("%.2f", change)}%"
    }

    // ============================================
    // Cache Management
    // ============================================

    private fun isCacheExpired(timestamp: Long): Boolean {
        return System.currentTimeMillis() - timestamp > NetworkConfig.PriceApi.CACHE_DURATION_MS
    }

    /**
     * Clear all cached prices
     */
    suspend fun clearCache() {
        cacheMutex.withLock {
            priceCache.clear()
        }
    }

    // ============================================
    // HTTP Utilities
    // ============================================

    private fun httpGet(urlString: String): JSONObject {
        val url = URL(urlString)
        val connection = url.openConnection() as HttpURLConnection

        return try {
            connection.requestMethod = "GET"
            connection.setRequestProperty("Accept", "application/json")
            connection.connectTimeout = TIMEOUT_MS
            connection.readTimeout = TIMEOUT_MS

            if (connection.responseCode == HttpURLConnection.HTTP_OK) {
                val response = connection.inputStream.bufferedReader().readText()
                JSONObject(response)
            } else {
                throw Exception("HTTP ${connection.responseCode}")
            }
        } finally {
            connection.disconnect()
        }
    }
}

/**
 * Price data for UI display
 */
data class PriceData(
    val chain: Chain,
    val usdPrice: Double,
    val priceChange24h: Double = 0.0,
    val lastUpdated: Long = System.currentTimeMillis()
)
