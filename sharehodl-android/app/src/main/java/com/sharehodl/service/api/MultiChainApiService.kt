package com.sharehodl.service.api

import com.sharehodl.config.NetworkConfig
import com.sharehodl.model.Chain
import com.sharehodl.model.ChainAccount
import com.sharehodl.model.CryptoTransaction
import com.sharehodl.model.TransactionStatus
import com.sharehodl.model.TransactionType
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.json.JSONArray
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Multi-chain API service for fetching balances and transaction history
 * Provides unified interface across different blockchain networks
 */
@Singleton
class MultiChainApiService @Inject constructor() {

    companion object {
        private const val TIMEOUT_MS = 15000
    }

    // ============================================
    // Balance Fetching
    // ============================================

    /**
     * Fetch balance for a chain account
     * Returns updated ChainAccount with balance information
     */
    suspend fun fetchBalance(account: ChainAccount): Result<ChainAccount> = withContext(Dispatchers.IO) {
        runCatching {
            when {
                account.chain == Chain.SHAREHODL -> fetchShareHodlBalance(account)
                account.chain.isEthereumStyle -> fetchEvmBalance(account)
                account.chain.isBitcoinStyle -> fetchBitcoinStyleBalance(account)
                account.chain == Chain.SOLANA -> fetchSolanaBalance(account)
                account.chain == Chain.XRP -> fetchXrpBalance(account)
                account.chain == Chain.TRON -> fetchTronBalance(account)
                else -> account // Return unchanged for unsupported chains
            }
        }
    }

    /**
     * Fetch balances for all chain accounts
     */
    suspend fun fetchAllBalances(accounts: List<ChainAccount>): List<ChainAccount> {
        return accounts.map { account ->
            fetchBalance(account).getOrDefault(account)
        }
    }

    // ============================================
    // ShareHODL Chain
    // ============================================

    private suspend fun fetchShareHodlBalance(account: ChainAccount): ChainAccount {
        val response = httpGet("${NetworkConfig.ShareHODL.restUrl}/cosmos/bank/v1beta1/balances/${account.address}")
        val balances = response.getJSONArray("balances")

        var amount = "0"
        for (i in 0 until balances.length()) {
            val balance = balances.getJSONObject(i)
            if (balance.getString("denom") == NetworkConfig.ShareHODL.DENOM) {
                amount = balance.getString("amount")
                break
            }
        }

        // Convert from micro to display amount
        val displayAmount = formatMicroAmount(amount, 6)

        return account.copy(
            balance = displayAmount,
            balanceUsd = null // TODO: Integrate price service
        )
    }

    /**
     * Fetch ShareHODL transaction history
     */
    suspend fun fetchShareHodlTransactions(address: String, limit: Int = 20): Result<List<CryptoTransaction>> =
        withContext(Dispatchers.IO) {
            runCatching {
                val response = httpGet(
                    "${NetworkConfig.ShareHODL.restUrl}/cosmos/tx/v1beta1/txs?events=message.sender='$address'&limit=$limit"
                )

                val txResponses = response.getJSONArray("tx_responses")
                val transactions = mutableListOf<CryptoTransaction>()

                for (i in 0 until txResponses.length()) {
                    val tx = txResponses.getJSONObject(i)
                    parseShareHodlTransaction(tx, address)?.let { transactions.add(it) }
                }

                transactions.sortedByDescending { it.timestamp }
            }
        }

    private fun parseShareHodlTransaction(tx: JSONObject, userAddress: String): CryptoTransaction? {
        val hash = tx.getString("txhash")
        val timestamp = tx.optString("timestamp", "")
        val code = tx.optInt("code", 0)

        val txBody = tx.optJSONObject("tx")?.optJSONObject("body") ?: return null
        val messages = txBody.optJSONArray("messages") ?: return null
        if (messages.length() == 0) return null

        val msg = messages.getJSONObject(0)
        val msgType = msg.optString("@type", "")

        if (!msgType.contains("MsgSend")) return null

        val fromAddress = msg.optString("from_address", "")
        val toAddress = msg.optString("to_address", "")
        val amounts = msg.optJSONArray("amount")
        val amount = if (amounts != null && amounts.length() > 0) {
            amounts.getJSONObject(0).optString("amount", "0")
        } else "0"

        val type = if (fromAddress == userAddress) TransactionType.SEND else TransactionType.RECEIVE
        val displayAmount = formatMicroAmount(amount, 6)

        return CryptoTransaction(
            txHash = hash,
            chain = Chain.SHAREHODL,
            type = type,
            amount = displayAmount,
            amountUsd = null,
            fromAddress = fromAddress,
            toAddress = toAddress,
            timestamp = parseTimestamp(timestamp),
            status = if (code == 0) TransactionStatus.CONFIRMED else TransactionStatus.FAILED,
            confirmations = 1
        )
    }

    // ============================================
    // EVM Chains (Ethereum, BNB, Polygon, Avalanche)
    // ============================================

    private suspend fun fetchEvmBalance(account: ChainAccount): ChainAccount {
        val rpcUrl = NetworkConfig.getApiUrl(account.chain)

        // For tokens (USDT, USDC), use ERC-20 balanceOf call
        if (account.chain.isToken) {
            return fetchErc20Balance(account, rpcUrl)
        }

        // For native coins, use eth_getBalance
        val requestBody = JSONObject().apply {
            put("jsonrpc", "2.0")
            put("method", "eth_getBalance")
            put("params", JSONArray().apply {
                put(account.address)
                put("latest")
            })
            put("id", 1)
        }

        val response = httpPost(rpcUrl, requestBody.toString())
        val balanceHex = response.optString("result", "0x0")
        val balanceWei = parseHexToBigInteger(balanceHex)

        // Convert from wei to display amount
        val displayAmount = formatWeiAmount(balanceWei.toString(), account.chain.decimals)

        return account.copy(
            balance = displayAmount,
            balanceUsd = null
        )
    }

    private suspend fun fetchErc20Balance(account: ChainAccount, rpcUrl: String): ChainAccount {
        // ERC-20 contract addresses
        val contractAddress = when (account.chain) {
            Chain.USDT -> if (NetworkConfig.isProduction) "0xdAC17F958D2ee523a2206206994597C13D831ec7" else "0x..."
            Chain.USDC -> if (NetworkConfig.isProduction) "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" else "0x..."
            else -> return account
        }

        // ERC-20 balanceOf method signature: 0x70a08231
        val data = "0x70a08231000000000000000000000000${account.address.removePrefix("0x")}"

        val requestBody = JSONObject().apply {
            put("jsonrpc", "2.0")
            put("method", "eth_call")
            put("params", JSONArray().apply {
                put(JSONObject().apply {
                    put("to", contractAddress)
                    put("data", data)
                })
                put("latest")
            })
            put("id", 1)
        }

        val response = httpPost(rpcUrl, requestBody.toString())
        val balanceHex = response.optString("result", "0x0")
        val balance = parseHexToBigInteger(balanceHex)

        val displayAmount = formatWeiAmount(balance.toString(), account.chain.decimals)

        return account.copy(
            balance = displayAmount,
            balanceUsd = null
        )
    }

    // ============================================
    // Bitcoin-Style Chains (BTC, LTC, DOGE)
    // ============================================

    private suspend fun fetchBitcoinStyleBalance(account: ChainAccount): ChainAccount {
        return when (account.chain) {
            Chain.BITCOIN -> fetchBitcoinBalance(account)
            Chain.LITECOIN, Chain.DOGECOIN -> fetchBlockcypherBalance(account)
            else -> account
        }
    }

    private suspend fun fetchBitcoinBalance(account: ChainAccount): ChainAccount {
        // Use Blockstream Esplora API
        val response = httpGet("${NetworkConfig.Bitcoin.esploraUrl}/address/${account.address}")

        val chainStats = response.getJSONObject("chain_stats")
        val mempoolStats = response.getJSONObject("mempool_stats")

        val confirmedBalance = chainStats.optLong("funded_txo_sum", 0) - chainStats.optLong("spent_txo_sum", 0)
        val unconfirmedBalance = mempoolStats.optLong("funded_txo_sum", 0) - mempoolStats.optLong("spent_txo_sum", 0)
        val totalSatoshis = confirmedBalance + unconfirmedBalance

        val displayAmount = formatSatoshiAmount(totalSatoshis)

        return account.copy(
            balance = displayAmount,
            balanceUsd = null
        )
    }

    private suspend fun fetchBlockcypherBalance(account: ChainAccount): ChainAccount {
        val apiUrl = when (account.chain) {
            Chain.LITECOIN -> NetworkConfig.Litecoin.apiUrl
            Chain.DOGECOIN -> NetworkConfig.Dogecoin.apiUrl
            else -> return account
        }

        val response = httpGet("$apiUrl/addrs/${account.address}/balance")
        val balance = response.optLong("balance", 0)
        val unconfirmedBalance = response.optLong("unconfirmed_balance", 0)
        val totalSatoshis = balance + unconfirmedBalance

        val displayAmount = formatSatoshiAmount(totalSatoshis)

        return account.copy(
            balance = displayAmount,
            balanceUsd = null
        )
    }

    /**
     * Fetch Bitcoin transaction history
     */
    suspend fun fetchBitcoinTransactions(address: String): Result<List<CryptoTransaction>> =
        withContext(Dispatchers.IO) {
            runCatching {
                val response = httpGet("${NetworkConfig.Bitcoin.esploraUrl}/address/$address/txs")

                // Response is a JSONArray directly
                val txArray = JSONArray(response.toString())
                val transactions = mutableListOf<CryptoTransaction>()

                for (i in 0 until minOf(txArray.length(), 20)) {
                    val tx = txArray.getJSONObject(i)
                    parseBitcoinTransaction(tx, address)?.let { transactions.add(it) }
                }

                transactions
            }
        }

    private fun parseBitcoinTransaction(tx: JSONObject, userAddress: String): CryptoTransaction? {
        val txid = tx.getString("txid")
        val status = tx.getJSONObject("status")
        val confirmed = status.optBoolean("confirmed", false)
        val blockTime = status.optLong("block_time", System.currentTimeMillis() / 1000)

        // Calculate amount based on inputs/outputs
        val vins = tx.getJSONArray("vin")
        val vouts = tx.getJSONArray("vout")

        var sentAmount = 0L
        var receivedAmount = 0L

        // Check inputs (what we spent)
        for (i in 0 until vins.length()) {
            val vin = vins.getJSONObject(i)
            val prevout = vin.optJSONObject("prevout")
            if (prevout != null) {
                val scriptPubKey = prevout.optString("scriptpubkey_address", "")
                if (scriptPubKey == userAddress) {
                    sentAmount += prevout.optLong("value", 0)
                }
            }
        }

        // Check outputs (what we received)
        for (i in 0 until vouts.length()) {
            val vout = vouts.getJSONObject(i)
            val scriptPubKey = vout.optString("scriptpubkey_address", "")
            if (scriptPubKey == userAddress) {
                receivedAmount += vout.optLong("value", 0)
            }
        }

        val type = if (sentAmount > receivedAmount) TransactionType.SEND else TransactionType.RECEIVE
        val netAmount = if (type == TransactionType.SEND) sentAmount - receivedAmount else receivedAmount

        return CryptoTransaction(
            txHash = txid,
            chain = Chain.BITCOIN,
            type = type,
            amount = formatSatoshiAmount(netAmount),
            amountUsd = null,
            fromAddress = if (type == TransactionType.SEND) userAddress else "multiple",
            toAddress = if (type == TransactionType.RECEIVE) userAddress else "multiple",
            timestamp = blockTime * 1000,
            status = if (confirmed) TransactionStatus.CONFIRMED else TransactionStatus.PENDING,
            confirmations = if (confirmed) 1 else 0
        )
    }

    // ============================================
    // Solana
    // ============================================

    private suspend fun fetchSolanaBalance(account: ChainAccount): ChainAccount {
        val requestBody = JSONObject().apply {
            put("jsonrpc", "2.0")
            put("method", "getBalance")
            put("params", JSONArray().apply {
                put(account.address)
            })
            put("id", 1)
        }

        val response = httpPost(NetworkConfig.Solana.rpcUrl, requestBody.toString())
        val result = response.optJSONObject("result")
        val lamports = result?.optLong("value", 0) ?: 0

        // Convert from lamports to SOL (9 decimals)
        val displayAmount = formatLamportsAmount(lamports)

        return account.copy(
            balance = displayAmount,
            balanceUsd = null
        )
    }

    // ============================================
    // XRP
    // ============================================

    private suspend fun fetchXrpBalance(account: ChainAccount): ChainAccount {
        val requestBody = JSONObject().apply {
            put("method", "account_info")
            put("params", JSONArray().apply {
                put(JSONObject().apply {
                    put("account", account.address)
                    put("ledger_index", "validated")
                })
            })
        }

        val response = httpPost(NetworkConfig.XRP.rpcUrl, requestBody.toString())
        val result = response.optJSONObject("result")
        val accountData = result?.optJSONObject("account_data")
        val drops = accountData?.optString("Balance", "0") ?: "0"

        // Convert from drops to XRP (6 decimals)
        val displayAmount = formatMicroAmount(drops, 6)

        return account.copy(
            balance = displayAmount,
            balanceUsd = null
        )
    }

    // ============================================
    // TRON
    // ============================================

    private suspend fun fetchTronBalance(account: ChainAccount): ChainAccount {
        val response = httpGet("${NetworkConfig.Tron.apiUrl}/v1/accounts/${account.address}")

        val data = response.optJSONArray("data")
        val balance = if (data != null && data.length() > 0) {
            data.getJSONObject(0).optLong("balance", 0)
        } else 0L

        // Convert from sun to TRX (6 decimals)
        val displayAmount = formatMicroAmount(balance.toString(), 6)

        return account.copy(
            balance = displayAmount,
            balanceUsd = null
        )
    }

    // ============================================
    // Transaction History (Generic)
    // ============================================

    /**
     * Fetch transaction history for any chain
     */
    suspend fun fetchTransactions(account: ChainAccount, limit: Int = 20): Result<List<CryptoTransaction>> {
        return when (account.chain) {
            Chain.SHAREHODL -> fetchShareHodlTransactions(account.address, limit)
            Chain.BITCOIN -> fetchBitcoinTransactions(account.address)
            // TODO: Add more chain-specific implementations
            else -> Result.success(emptyList())
        }
    }

    // ============================================
    // Helper Functions
    // ============================================

    private fun httpGet(urlString: String): JSONObject {
        val url = URL(urlString)
        val connection = url.openConnection() as HttpURLConnection

        return try {
            connection.requestMethod = "GET"
            connection.setRequestProperty("Content-Type", "application/json")
            connection.connectTimeout = TIMEOUT_MS
            connection.readTimeout = TIMEOUT_MS

            if (connection.responseCode == HttpURLConnection.HTTP_OK) {
                val response = connection.inputStream.bufferedReader().readText()
                JSONObject(response)
            } else {
                throw Exception("HTTP ${connection.responseCode}: ${connection.responseMessage}")
            }
        } finally {
            connection.disconnect()
        }
    }

    private fun httpPost(urlString: String, body: String): JSONObject {
        val url = URL(urlString)
        val connection = url.openConnection() as HttpURLConnection

        return try {
            connection.requestMethod = "POST"
            connection.setRequestProperty("Content-Type", "application/json")
            connection.doOutput = true
            connection.connectTimeout = TIMEOUT_MS
            connection.readTimeout = TIMEOUT_MS

            connection.outputStream.bufferedWriter().use { it.write(body) }

            if (connection.responseCode == HttpURLConnection.HTTP_OK) {
                val response = connection.inputStream.bufferedReader().readText()
                JSONObject(response)
            } else {
                val error = connection.errorStream?.bufferedReader()?.readText() ?: ""
                throw Exception("HTTP ${connection.responseCode}: $error")
            }
        } finally {
            connection.disconnect()
        }
    }

    private fun formatMicroAmount(microAmount: String, decimals: Int): String {
        val amount = microAmount.toLongOrNull() ?: 0
        val divisor = Math.pow(10.0, decimals.toDouble())
        return String.format("%.${minOf(decimals, 6)}f", amount / divisor)
    }

    private fun formatWeiAmount(weiAmount: String, decimals: Int): String {
        val amount = weiAmount.toBigIntegerOrNull() ?: java.math.BigInteger.ZERO
        val divisor = java.math.BigInteger.TEN.pow(decimals)
        val wholePart = amount.divide(divisor)
        val fractionalPart = amount.mod(divisor)

        return if (fractionalPart == java.math.BigInteger.ZERO) {
            wholePart.toString()
        } else {
            val fractionalStr = fractionalPart.toString().padStart(decimals, '0').take(6).trimEnd('0')
            if (fractionalStr.isEmpty()) wholePart.toString() else "$wholePart.$fractionalStr"
        }
    }

    private fun formatSatoshiAmount(satoshis: Long): String {
        val btc = satoshis / 100_000_000.0
        return String.format("%.8f", btc)
    }

    private fun formatLamportsAmount(lamports: Long): String {
        val sol = lamports / 1_000_000_000.0
        return String.format("%.9f", sol)
    }

    private fun parseHexToBigInteger(hex: String): java.math.BigInteger {
        val cleanHex = hex.removePrefix("0x")
        return if (cleanHex.isEmpty()) java.math.BigInteger.ZERO
        else java.math.BigInteger(cleanHex, 16)
    }

    private fun parseTimestamp(timestamp: String): Long {
        return try {
            java.time.Instant.parse(timestamp).toEpochMilli()
        } catch (e: Exception) {
            System.currentTimeMillis()
        }
    }
}
