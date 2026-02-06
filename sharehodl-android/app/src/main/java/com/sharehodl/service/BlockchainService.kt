package com.sharehodl.service

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.json.JSONArray
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Service for interacting with ShareHODL blockchain via REST API
 * SECURITY: Enforces HTTPS for all production API calls
 */
@Singleton
class BlockchainService @Inject constructor() {

    companion object {
        // Default to localhost for development
        private const val DEFAULT_BASE_URL = "http://localhost:1317"
        private const val CHAIN_ID = "sharehodl-1"
        private const val DENOM = "uhodl"
    }

    private var baseUrl: String = DEFAULT_BASE_URL
    private var isProduction: Boolean = false
    private var enforceHTTPS: Boolean = false

    /**
     * Configure blockchain endpoints
     * @param baseUrl The REST API base URL
     * @param isProduction If true, enforces HTTPS (default: false)
     * @throws IllegalArgumentException if configuration is invalid
     */
    fun configure(baseUrl: String, isProduction: Boolean = false) {
        val url = URL(baseUrl)

        // Enforce HTTPS in production
        if (isProduction) {
            require(url.protocol == "https") {
                "HTTPS required for production endpoints"
            }
        }

        // Check for localhost (development only)
        val isLocalhost = url.host == "localhost" || url.host == "127.0.0.1"
        if (isProduction && isLocalhost) {
            throw IllegalArgumentException("Localhost not allowed in production")
        }

        this.baseUrl = baseUrl.trimEnd('/')
        this.isProduction = isProduction
        this.enforceHTTPS = isProduction
    }

    /**
     * Configure for local development (HTTP allowed)
     */
    fun configureForDevelopment(baseUrl: String = DEFAULT_BASE_URL) {
        this.baseUrl = baseUrl.trimEnd('/')
        this.isProduction = false
        this.enforceHTTPS = false
    }

    /**
     * Validate URL before making request
     */
    private fun validateUrl(path: String): URL {
        val fullUrl = "$baseUrl$path"
        val url = URL(fullUrl)

        if (enforceHTTPS && url.protocol != "https") {
            throw SecurityException("HTTPS required for API requests")
        }

        return url
    }

    // MARK: - Account & Balance

    /**
     * Get account balance for an address
     */
    suspend fun getBalance(address: String): Result<Balance> = withContext(Dispatchers.IO) {
        runCatching {
            val response = httpGet("/cosmos/bank/v1beta1/balances/$address")
            val balances = response.getJSONArray("balances")

            var amount = "0"
            for (i in 0 until balances.length()) {
                val balance = balances.getJSONObject(i)
                if (balance.getString("denom") == DENOM) {
                    amount = balance.getString("amount")
                    break
                }
            }

            Balance(
                denom = DENOM,
                amount = amount,
                displayAmount = formatAmount(amount)
            )
        }
    }

    /**
     * Get account info (sequence and account number)
     */
    suspend fun getAccount(address: String): Result<AccountInfo> = withContext(Dispatchers.IO) {
        runCatching {
            val response = httpGet("/cosmos/auth/v1beta1/accounts/$address")
            val account = response.getJSONObject("account")

            AccountInfo(
                address = address,
                accountNumber = account.optString("account_number", "0"),
                sequence = account.optString("sequence", "0")
            )
        }
    }

    // MARK: - Staking

    /**
     * Get all validators
     */
    suspend fun getValidators(): Result<List<Validator>> = withContext(Dispatchers.IO) {
        runCatching {
            val response = httpGet("/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED")
            val validators = response.getJSONArray("validators")

            val result = mutableListOf<Validator>()
            for (i in 0 until validators.length()) {
                val v = validators.getJSONObject(i)
                val description = v.getJSONObject("description")
                val commission = v.getJSONObject("commission")
                    .getJSONObject("commission_rates")

                result.add(Validator(
                    operatorAddress = v.getString("operator_address"),
                    moniker = description.getString("moniker"),
                    identity = description.optString("identity", ""),
                    website = description.optString("website", ""),
                    details = description.optString("details", ""),
                    tokens = v.getString("tokens"),
                    commissionRate = commission.getString("rate"),
                    status = v.getString("status"),
                    jailed = v.optBoolean("jailed", false)
                ))
            }

            result.sortedByDescending { it.tokens.toLongOrNull() ?: 0 }
        }
    }

    /**
     * Get delegations for an address
     */
    suspend fun getDelegations(delegatorAddress: String): Result<List<Delegation>> =
        withContext(Dispatchers.IO) {
            runCatching {
                val response = httpGet(
                    "/cosmos/staking/v1beta1/delegations/$delegatorAddress"
                )
                val delegations = response.getJSONArray("delegation_responses")

                val result = mutableListOf<Delegation>()
                for (i in 0 until delegations.length()) {
                    val d = delegations.getJSONObject(i)
                    val delegation = d.getJSONObject("delegation")
                    val balance = d.getJSONObject("balance")

                    result.add(Delegation(
                        delegatorAddress = delegation.getString("delegator_address"),
                        validatorAddress = delegation.getString("validator_address"),
                        shares = delegation.getString("shares"),
                        amount = balance.getString("amount"),
                        denom = balance.getString("denom")
                    ))
                }

                result
            }
        }

    /**
     * Get staking rewards for an address
     */
    suspend fun getRewards(delegatorAddress: String): Result<Rewards> =
        withContext(Dispatchers.IO) {
            runCatching {
                val response = httpGet(
                    "/cosmos/distribution/v1beta1/delegators/$delegatorAddress/rewards"
                )

                val total = response.getJSONArray("total")
                var totalAmount = "0"
                for (i in 0 until total.length()) {
                    val reward = total.getJSONObject(i)
                    if (reward.getString("denom") == DENOM) {
                        totalAmount = reward.getString("amount").split(".")[0]
                        break
                    }
                }

                val rewardsList = response.getJSONArray("rewards")
                val validatorRewards = mutableListOf<ValidatorReward>()

                for (i in 0 until rewardsList.length()) {
                    val r = rewardsList.getJSONObject(i)
                    val validatorAddress = r.getString("validator_address")
                    val rewards = r.getJSONArray("reward")

                    var amount = "0"
                    for (j in 0 until rewards.length()) {
                        val reward = rewards.getJSONObject(j)
                        if (reward.getString("denom") == DENOM) {
                            amount = reward.getString("amount").split(".")[0]
                            break
                        }
                    }

                    validatorRewards.add(ValidatorReward(
                        validatorAddress = validatorAddress,
                        amount = amount
                    ))
                }

                Rewards(
                    totalAmount = totalAmount,
                    denom = DENOM,
                    validatorRewards = validatorRewards
                )
            }
        }

    // MARK: - Governance

    /**
     * Get governance proposals
     */
    suspend fun getProposals(): Result<List<Proposal>> = withContext(Dispatchers.IO) {
        runCatching {
            val response = httpGet("/cosmos/gov/v1/proposals")
            val proposals = response.getJSONArray("proposals")

            val result = mutableListOf<Proposal>()
            for (i in 0 until proposals.length()) {
                val p = proposals.getJSONObject(i)

                result.add(Proposal(
                    id = p.getString("id"),
                    title = p.optString("title", "Proposal ${p.getString("id")}"),
                    description = p.optString("summary", ""),
                    status = p.getString("status"),
                    votingEndTime = p.optString("voting_end_time", ""),
                    submitTime = p.optString("submit_time", "")
                ))
            }

            result.sortedByDescending { it.id.toLongOrNull() ?: 0 }
        }
    }

    // MARK: - Transactions

    /**
     * Get transaction history for an address
     */
    suspend fun getTransactions(address: String, limit: Int = 20): Result<List<Transaction>> =
        withContext(Dispatchers.IO) {
            runCatching {
                // Query transactions where address is sender or recipient
                val sentResponse = httpGet(
                    "/cosmos/tx/v1beta1/txs?events=message.sender='$address'&limit=$limit"
                )

                val txs = mutableListOf<Transaction>()
                val sentTxs = sentResponse.getJSONArray("tx_responses")

                for (i in 0 until sentTxs.length()) {
                    val tx = sentTxs.getJSONObject(i)
                    txs.add(parseTransaction(tx, address))
                }

                txs.sortedByDescending { it.height.toLongOrNull() ?: 0 }
            }
        }

    private fun parseTransaction(tx: JSONObject, userAddress: String): Transaction {
        val hash = tx.getString("txhash")
        val height = tx.getString("height")
        val timestamp = tx.optString("timestamp", "")
        val code = tx.optInt("code", 0)

        // Parse messages from the transaction
        var type = "Unknown"
        var amount = "0"
        var toAddress = ""
        var fromAddress = ""

        val txBody = tx.optJSONObject("tx")?.optJSONObject("body")
        val messages = txBody?.optJSONArray("messages")

        if (messages != null && messages.length() > 0) {
            val msg = messages.getJSONObject(0)
            val msgType = msg.optString("@type", "")

            when {
                msgType.contains("MsgSend") -> {
                    type = "Send"
                    fromAddress = msg.optString("from_address", "")
                    toAddress = msg.optString("to_address", "")
                    val amounts = msg.optJSONArray("amount")
                    if (amounts != null && amounts.length() > 0) {
                        amount = amounts.getJSONObject(0).optString("amount", "0")
                    }
                }
                msgType.contains("MsgDelegate") -> {
                    type = "Delegate"
                    amount = msg.optJSONObject("amount")?.optString("amount", "0") ?: "0"
                }
                msgType.contains("MsgUndelegate") -> {
                    type = "Undelegate"
                    amount = msg.optJSONObject("amount")?.optString("amount", "0") ?: "0"
                }
                msgType.contains("MsgWithdrawDelegatorReward") -> {
                    type = "Claim Rewards"
                }
                msgType.contains("MsgVote") -> {
                    type = "Vote"
                }
            }
        }

        return Transaction(
            hash = hash,
            height = height,
            timestamp = timestamp,
            type = type,
            amount = amount,
            fromAddress = fromAddress,
            toAddress = toAddress,
            success = code == 0
        )
    }

    // MARK: - Transaction Broadcasting

    /**
     * Broadcast a signed transaction
     */
    suspend fun broadcastTransaction(signedTxBytes: ByteArray): Result<BroadcastResult> =
        withContext(Dispatchers.IO) {
            runCatching {
                val txBase64 = android.util.Base64.encodeToString(
                    signedTxBytes,
                    android.util.Base64.NO_WRAP
                )

                val requestBody = JSONObject().apply {
                    put("tx_bytes", txBase64)
                    put("mode", "BROADCAST_MODE_SYNC")
                }

                val response = httpPost("/cosmos/tx/v1beta1/txs", requestBody.toString())
                val txResponse = response.getJSONObject("tx_response")

                BroadcastResult(
                    txHash = txResponse.getString("txhash"),
                    code = txResponse.optInt("code", 0),
                    rawLog = txResponse.optString("raw_log", "")
                )
            }
        }

    // MARK: - HTTP Utilities

    private fun httpGet(endpoint: String): JSONObject {
        val url = validateUrl(endpoint)
        val connection = url.openConnection() as HttpURLConnection

        return try {
            connection.requestMethod = "GET"
            connection.setRequestProperty("Content-Type", "application/json")
            connection.connectTimeout = 10000
            connection.readTimeout = 10000

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

    private fun httpPost(endpoint: String, body: String): JSONObject {
        val url = validateUrl(endpoint)
        val connection = url.openConnection() as HttpURLConnection

        return try {
            connection.requestMethod = "POST"
            connection.setRequestProperty("Content-Type", "application/json")
            connection.doOutput = true
            connection.connectTimeout = 10000
            connection.readTimeout = 10000

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

    // MARK: - Utilities

    private fun formatAmount(microAmount: String): String {
        val amount = microAmount.toLongOrNull() ?: 0
        val hodl = amount / 1_000_000.0
        return String.format("%.6f", hodl)
    }
}

// MARK: - Data Models

data class Balance(
    val denom: String,
    val amount: String,
    val displayAmount: String
)

data class AccountInfo(
    val address: String,
    val accountNumber: String,
    val sequence: String
)

data class Validator(
    val operatorAddress: String,
    val moniker: String,
    val identity: String,
    val website: String,
    val details: String,
    val tokens: String,
    val commissionRate: String,
    val status: String,
    val jailed: Boolean
) {
    val displayCommission: String
        get() {
            val rate = commissionRate.toDoubleOrNull() ?: 0.0
            return String.format("%.2f%%", rate * 100)
        }

    val displayTokens: String
        get() {
            val amount = tokens.toLongOrNull() ?: 0
            val hodl = amount / 1_000_000.0
            return String.format("%.2f HODL", hodl)
        }
}

data class Delegation(
    val delegatorAddress: String,
    val validatorAddress: String,
    val shares: String,
    val amount: String,
    val denom: String
) {
    val displayAmount: String
        get() {
            val amt = amount.toLongOrNull() ?: 0
            val hodl = amt / 1_000_000.0
            return String.format("%.6f HODL", hodl)
        }
}

data class Rewards(
    val totalAmount: String,
    val denom: String,
    val validatorRewards: List<ValidatorReward>
) {
    val displayTotal: String
        get() {
            val amount = totalAmount.toLongOrNull() ?: 0
            val hodl = amount / 1_000_000.0
            return String.format("%.6f HODL", hodl)
        }
}

data class ValidatorReward(
    val validatorAddress: String,
    val amount: String
)

data class Proposal(
    val id: String,
    val title: String,
    val description: String,
    val status: String,
    val votingEndTime: String,
    val submitTime: String
) {
    val displayStatus: String
        get() = when (status) {
            "PROPOSAL_STATUS_VOTING_PERIOD" -> "Voting"
            "PROPOSAL_STATUS_PASSED" -> "Passed"
            "PROPOSAL_STATUS_REJECTED" -> "Rejected"
            "PROPOSAL_STATUS_DEPOSIT_PERIOD" -> "Deposit"
            else -> status.replace("PROPOSAL_STATUS_", "")
        }
}

data class Transaction(
    val hash: String,
    val height: String,
    val timestamp: String,
    val type: String,
    val amount: String,
    val fromAddress: String,
    val toAddress: String,
    val success: Boolean
) {
    val displayAmount: String
        get() {
            val amt = amount.toLongOrNull() ?: 0
            if (amt == 0L) return ""
            val hodl = amt / 1_000_000.0
            return String.format("%.6f HODL", hodl)
        }

    val shortHash: String
        get() = if (hash.length > 12) "${hash.take(6)}...${hash.takeLast(6)}" else hash
}

data class BroadcastResult(
    val txHash: String,
    val code: Int,
    val rawLog: String
) {
    val success: Boolean get() = code == 0
}
