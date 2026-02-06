package com.sharehodl.service.tx

import com.sharehodl.config.FeeLevel
import com.sharehodl.config.NetworkConfig
import com.sharehodl.service.AccountInfo
import com.sharehodl.service.CryptoService
import org.json.JSONArray
import org.json.JSONObject
import java.security.MessageDigest
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Transaction builder for Cosmos SDK chains (ShareHODL)
 * Builds, signs, and encodes transactions in the SIGN_MODE_DIRECT format
 */
@Singleton
class CosmosTransactionBuilder @Inject constructor(
    private val cryptoService: CryptoService
) {

    companion object {
        private const val DENOM = "uhodl"
    }

    // ============================================
    // Message Types (Amino-compatible JSON)
    // ============================================

    /**
     * Build a MsgSend message for token transfers
     */
    fun buildMsgSend(
        fromAddress: String,
        toAddress: String,
        amount: String,
        denom: String = DENOM
    ): CosmosMessage {
        return CosmosMessage(
            typeUrl = "/cosmos.bank.v1beta1.MsgSend",
            value = JSONObject().apply {
                put("from_address", fromAddress)
                put("to_address", toAddress)
                put("amount", JSONArray().apply {
                    put(JSONObject().apply {
                        put("denom", denom)
                        put("amount", amount)
                    })
                })
            }
        )
    }

    /**
     * Build a MsgDelegate message for staking
     */
    fun buildMsgDelegate(
        delegatorAddress: String,
        validatorAddress: String,
        amount: String,
        denom: String = DENOM
    ): CosmosMessage {
        return CosmosMessage(
            typeUrl = "/cosmos.staking.v1beta1.MsgDelegate",
            value = JSONObject().apply {
                put("delegator_address", delegatorAddress)
                put("validator_address", validatorAddress)
                put("amount", JSONObject().apply {
                    put("denom", denom)
                    put("amount", amount)
                })
            }
        )
    }

    /**
     * Build a MsgUndelegate message for unstaking
     */
    fun buildMsgUndelegate(
        delegatorAddress: String,
        validatorAddress: String,
        amount: String,
        denom: String = DENOM
    ): CosmosMessage {
        return CosmosMessage(
            typeUrl = "/cosmos.staking.v1beta1.MsgUndelegate",
            value = JSONObject().apply {
                put("delegator_address", delegatorAddress)
                put("validator_address", validatorAddress)
                put("amount", JSONObject().apply {
                    put("denom", denom)
                    put("amount", amount)
                })
            }
        )
    }

    /**
     * Build a MsgWithdrawDelegatorReward message for claiming rewards
     */
    fun buildMsgWithdrawReward(
        delegatorAddress: String,
        validatorAddress: String
    ): CosmosMessage {
        return CosmosMessage(
            typeUrl = "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
            value = JSONObject().apply {
                put("delegator_address", delegatorAddress)
                put("validator_address", validatorAddress)
            }
        )
    }

    /**
     * Build a MsgVote message for governance voting
     */
    fun buildMsgVote(
        proposalId: String,
        voter: String,
        option: VoteOption
    ): CosmosMessage {
        return CosmosMessage(
            typeUrl = "/cosmos.gov.v1beta1.MsgVote",
            value = JSONObject().apply {
                put("proposal_id", proposalId)
                put("voter", voter)
                put("option", option.value)
            }
        )
    }

    // ============================================
    // Transaction Building
    // ============================================

    /**
     * Build an unsigned transaction
     */
    fun buildUnsignedTx(
        messages: List<CosmosMessage>,
        memo: String = "",
        feeAmount: String,
        gasLimit: Long = NetworkConfig.ShareHODL.DEFAULT_GAS_LIMIT
    ): UnsignedTransaction {
        val fee = Fee(
            amount = listOf(Coin(DENOM, feeAmount)),
            gasLimit = gasLimit
        )

        return UnsignedTransaction(
            messages = messages,
            memo = memo,
            fee = fee
        )
    }

    /**
     * Calculate fee based on gas estimate and fee level
     */
    fun calculateFee(
        gasLimit: Long = NetworkConfig.ShareHODL.DEFAULT_GAS_LIMIT,
        feeLevel: FeeLevel = FeeLevel.MEDIUM
    ): String {
        val gasPrice = when (feeLevel) {
            FeeLevel.LOW -> NetworkConfig.ShareHODL.GAS_PRICE_LOW.toDouble()
            FeeLevel.MEDIUM -> NetworkConfig.ShareHODL.GAS_PRICE_MEDIUM.toDouble()
            FeeLevel.HIGH -> NetworkConfig.ShareHODL.GAS_PRICE_HIGH.toDouble()
        }

        val feeAmount = (gasLimit * gasPrice).toLong()
        return feeAmount.toString()
    }

    // ============================================
    // Transaction Signing (SIGN_MODE_DIRECT)
    // ============================================

    /**
     * Sign a transaction and return the signed transaction bytes
     * Uses SIGN_MODE_DIRECT (protobuf encoding)
     */
    fun signTransaction(
        unsignedTx: UnsignedTransaction,
        accountInfo: AccountInfo,
        privateKey: ByteArray
    ): ByteArray {
        val chainId = NetworkConfig.ShareHODL.chainId

        // Derive public key from private key
        val publicKey = cryptoService.derivePublicKey(privateKey)

        // Build the SignDoc (what gets signed)
        val signDoc = buildSignDoc(
            unsignedTx = unsignedTx,
            chainId = chainId,
            accountNumber = accountInfo.accountNumber.toLong(),
            sequence = accountInfo.sequence.toLong()
        )

        // Hash the sign doc
        val signDocBytes = signDoc.toByteArray()
        val hash = sha256(signDocBytes)

        // Sign with ECDSA secp256k1
        val signature = cryptoService.signWithEcdsa(privateKey, hash)

        // Build the complete signed transaction
        return buildSignedTxBytes(unsignedTx, accountInfo, signature, publicKey)
    }

    /**
     * Build the SignDoc for SIGN_MODE_DIRECT
     * This is what gets hashed and signed
     */
    private fun buildSignDoc(
        unsignedTx: UnsignedTransaction,
        chainId: String,
        accountNumber: Long,
        sequence: Long
    ): SignDoc {
        // Build TxBody
        val txBody = TxBody(
            messages = unsignedTx.messages,
            memo = unsignedTx.memo
        )

        // Build AuthInfo
        val authInfo = AuthInfo(
            fee = unsignedTx.fee,
            sequence = sequence
        )

        return SignDoc(
            bodyBytes = txBody.encode(),
            authInfoBytes = authInfo.encode(),
            chainId = chainId,
            accountNumber = accountNumber
        )
    }

    /**
     * Build the final signed transaction bytes for broadcasting
     */
    private fun buildSignedTxBytes(
        unsignedTx: UnsignedTransaction,
        accountInfo: AccountInfo,
        signature: ByteArray,
        publicKey: ByteArray
    ): ByteArray {
        // Build the signed transaction JSON for amino encoding
        // Note: For production, use protobuf encoding
        val signedTx = JSONObject().apply {
            put("body", JSONObject().apply {
                put("messages", JSONArray().apply {
                    unsignedTx.messages.forEach { msg ->
                        put(JSONObject().apply {
                            put("@type", msg.typeUrl)
                            // Merge message value properties
                            val keys = msg.value.keys()
                            while (keys.hasNext()) {
                                val key = keys.next()
                                put(key, msg.value.get(key))
                            }
                        })
                    }
                })
                put("memo", unsignedTx.memo)
                put("timeout_height", "0")
                put("extension_options", JSONArray())
                put("non_critical_extension_options", JSONArray())
            })
            put("auth_info", JSONObject().apply {
                put("signer_infos", JSONArray().apply {
                    put(JSONObject().apply {
                        put("public_key", JSONObject().apply {
                            put("@type", "/cosmos.crypto.secp256k1.PubKey")
                            put("key", android.util.Base64.encodeToString(
                                publicKey,
                                android.util.Base64.NO_WRAP
                            ))
                        })
                        put("mode_info", JSONObject().apply {
                            put("single", JSONObject().apply {
                                put("mode", "SIGN_MODE_DIRECT")
                            })
                        })
                        put("sequence", accountInfo.sequence)
                    })
                })
                put("fee", JSONObject().apply {
                    put("amount", JSONArray().apply {
                        unsignedTx.fee.amount.forEach { coin ->
                            put(JSONObject().apply {
                                put("denom", coin.denom)
                                put("amount", coin.amount)
                            })
                        }
                    })
                    put("gas_limit", unsignedTx.fee.gasLimit.toString())
                    put("payer", "")
                    put("granter", "")
                })
            })
            put("signatures", JSONArray().apply {
                put(android.util.Base64.encodeToString(signature, android.util.Base64.NO_WRAP))
            })
        }

        return signedTx.toString().toByteArray()
    }

    private fun sha256(data: ByteArray): ByteArray {
        return MessageDigest.getInstance("SHA-256").digest(data)
    }
}

// ============================================
// Data Classes
// ============================================

data class CosmosMessage(
    val typeUrl: String,
    val value: JSONObject
)

data class Coin(
    val denom: String,
    val amount: String
)

data class Fee(
    val amount: List<Coin>,
    val gasLimit: Long
)

data class UnsignedTransaction(
    val messages: List<CosmosMessage>,
    val memo: String,
    val fee: Fee
)

data class TxBody(
    val messages: List<CosmosMessage>,
    val memo: String
) {
    fun encode(): ByteArray {
        // Simplified encoding - use protobuf in production
        val json = JSONObject().apply {
            put("messages", JSONArray().apply {
                messages.forEach { msg ->
                    put(JSONObject().apply {
                        put("@type", msg.typeUrl)
                        val keys = msg.value.keys()
                        while (keys.hasNext()) {
                            val key = keys.next()
                            put(key, msg.value.get(key))
                        }
                    })
                }
            })
            put("memo", memo)
        }
        return json.toString().toByteArray()
    }
}

data class AuthInfo(
    val fee: Fee,
    val sequence: Long
) {
    fun encode(): ByteArray {
        val json = JSONObject().apply {
            put("fee", JSONObject().apply {
                put("amount", JSONArray().apply {
                    fee.amount.forEach { coin ->
                        put(JSONObject().apply {
                            put("denom", coin.denom)
                            put("amount", coin.amount)
                        })
                    }
                })
                put("gas_limit", fee.gasLimit.toString())
            })
            put("sequence", sequence.toString())
        }
        return json.toString().toByteArray()
    }
}

data class SignDoc(
    val bodyBytes: ByteArray,
    val authInfoBytes: ByteArray,
    val chainId: String,
    val accountNumber: Long
) {
    fun toByteArray(): ByteArray {
        // Combine all fields for signing
        val combined = JSONObject().apply {
            put("body_bytes", android.util.Base64.encodeToString(bodyBytes, android.util.Base64.NO_WRAP))
            put("auth_info_bytes", android.util.Base64.encodeToString(authInfoBytes, android.util.Base64.NO_WRAP))
            put("chain_id", chainId)
            put("account_number", accountNumber.toString())
        }
        return combined.toString().toByteArray()
    }
}

enum class VoteOption(val value: Int) {
    UNSPECIFIED(0),
    YES(1),
    ABSTAIN(2),
    NO(3),
    NO_WITH_VETO(4)
}
