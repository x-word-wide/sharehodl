package com.sharehodl.config

import com.sharehodl.model.Chain

/**
 * Network configuration for production and testnet environments
 * Centralized endpoint management for all supported chains
 */
object NetworkConfig {

    // Environment toggle - set to true for production
    var isProduction: Boolean = false
        private set

    fun setEnvironment(production: Boolean) {
        isProduction = production
    }

    // ============================================
    // ShareHODL Chain Configuration
    // ============================================

    object ShareHODL {
        const val CHAIN_ID_MAINNET = "sharehodl-1"
        const val CHAIN_ID_TESTNET = "sharehodl-testnet-1"

        const val REST_MAINNET = "https://api.sharehodl.com"
        const val REST_TESTNET = "https://testnet-api.sharehodl.com"

        const val RPC_MAINNET = "https://rpc.sharehodl.com"
        const val RPC_TESTNET = "https://testnet-rpc.sharehodl.com"

        const val GRPC_MAINNET = "grpc.sharehodl.com:9090"
        const val GRPC_TESTNET = "testnet-grpc.sharehodl.com:9090"

        const val DENOM = "uhodl"
        const val DISPLAY_DENOM = "HODL"
        const val DECIMALS = 6

        // Gas configuration
        const val DEFAULT_GAS_LIMIT = 200000L
        const val GAS_PRICE_LOW = "0.001"
        const val GAS_PRICE_MEDIUM = "0.0025"
        const val GAS_PRICE_HIGH = "0.005"

        val chainId: String get() = if (isProduction) CHAIN_ID_MAINNET else CHAIN_ID_TESTNET
        val restUrl: String get() = if (isProduction) REST_MAINNET else REST_TESTNET
        val rpcUrl: String get() = if (isProduction) RPC_MAINNET else RPC_TESTNET
        val grpcUrl: String get() = if (isProduction) GRPC_MAINNET else GRPC_TESTNET
    }

    // ============================================
    // EVM Chain Configuration (Ethereum, BNB, etc.)
    // ============================================

    object EVM {
        // API Keys - should be stored in BuildConfig in production
        var alchemyApiKey: String = ""
            private set
        var infuraApiKey: String = ""
            private set

        fun setApiKeys(alchemy: String, infura: String) {
            alchemyApiKey = alchemy
            infuraApiKey = infura
        }

        // Ethereum
        object Ethereum {
            const val CHAIN_ID_MAINNET = 1
            const val CHAIN_ID_SEPOLIA = 11155111

            fun rpcUrl(): String = if (isProduction) {
                "https://eth-mainnet.g.alchemy.com/v2/$alchemyApiKey"
            } else {
                "https://eth-sepolia.g.alchemy.com/v2/$alchemyApiKey"
            }

            val chainId: Int get() = if (isProduction) CHAIN_ID_MAINNET else CHAIN_ID_SEPOLIA
        }

        // BNB Smart Chain
        object BNB {
            const val CHAIN_ID_MAINNET = 56
            const val CHAIN_ID_TESTNET = 97

            val rpcUrl: String get() = if (isProduction) {
                "https://bsc-dataseed.binance.org"
            } else {
                "https://data-seed-prebsc-1-s1.binance.org:8545"
            }

            val chainId: Int get() = if (isProduction) CHAIN_ID_MAINNET else CHAIN_ID_TESTNET
        }

        // Polygon
        object Polygon {
            const val CHAIN_ID_MAINNET = 137
            const val CHAIN_ID_MUMBAI = 80001

            fun rpcUrl(): String = if (isProduction) {
                "https://polygon-mainnet.g.alchemy.com/v2/$alchemyApiKey"
            } else {
                "https://polygon-mumbai.g.alchemy.com/v2/$alchemyApiKey"
            }

            val chainId: Int get() = if (isProduction) CHAIN_ID_MAINNET else CHAIN_ID_MUMBAI
        }

        // Avalanche C-Chain
        object Avalanche {
            const val CHAIN_ID_MAINNET = 43114
            const val CHAIN_ID_FUJI = 43113

            val rpcUrl: String get() = if (isProduction) {
                "https://api.avax.network/ext/bc/C/rpc"
            } else {
                "https://api.avax-test.network/ext/bc/C/rpc"
            }

            val chainId: Int get() = if (isProduction) CHAIN_ID_MAINNET else CHAIN_ID_FUJI
        }
    }

    // ============================================
    // Bitcoin & UTXO Chain Configuration
    // ============================================

    object Bitcoin {
        const val NETWORK_MAINNET = "mainnet"
        const val NETWORK_TESTNET = "testnet"

        // Blockstream Esplora API (free, no API key needed)
        val esploraUrl: String get() = if (isProduction) {
            "https://blockstream.info/api"
        } else {
            "https://blockstream.info/testnet/api"
        }

        // Mempool.space API (alternative)
        val mempoolUrl: String get() = if (isProduction) {
            "https://mempool.space/api"
        } else {
            "https://mempool.space/testnet/api"
        }

        val network: String get() = if (isProduction) NETWORK_MAINNET else NETWORK_TESTNET
    }

    object Litecoin {
        // BlockCypher API
        val apiUrl: String get() = if (isProduction) {
            "https://api.blockcypher.com/v1/ltc/main"
        } else {
            "https://api.blockcypher.com/v1/ltc/test3"
        }
    }

    object Dogecoin {
        // BlockCypher API
        val apiUrl: String get() = if (isProduction) {
            "https://api.blockcypher.com/v1/doge/main"
        } else {
            // Dogecoin testnet is limited, use mainnet with caution
            "https://api.blockcypher.com/v1/doge/main"
        }
    }

    // ============================================
    // Other Chain Configuration
    // ============================================

    object Solana {
        var heliusApiKey: String = ""
            private set

        fun setApiKey(key: String) {
            heliusApiKey = key
        }

        val rpcUrl: String get() = if (isProduction) {
            if (heliusApiKey.isNotEmpty()) {
                "https://mainnet.helius-rpc.com/?api-key=$heliusApiKey"
            } else {
                "https://api.mainnet-beta.solana.com"
            }
        } else {
            "https://api.devnet.solana.com"
        }
    }

    object XRP {
        val rpcUrl: String get() = if (isProduction) {
            "https://xrplcluster.com"
        } else {
            "https://s.altnet.rippletest.net:51234"
        }
    }

    object Tron {
        val apiUrl: String get() = if (isProduction) {
            "https://api.trongrid.io"
        } else {
            "https://api.shasta.trongrid.io"
        }
    }

    // ============================================
    // Price API Configuration
    // ============================================

    object PriceApi {
        const val COINGECKO_URL = "https://api.coingecko.com/api/v3"
        const val CACHE_DURATION_MS = 60_000L // 1 minute cache

        // CoinGecko IDs for each chain
        val coinGeckoIds = mapOf(
            Chain.BITCOIN to "bitcoin",
            Chain.ETHEREUM to "ethereum",
            Chain.USDT to "tether",
            Chain.USDC to "usd-coin",
            Chain.BNB to "binancecoin",
            Chain.LITECOIN to "litecoin",
            Chain.DOGECOIN to "dogecoin",
            Chain.SOLANA to "solana",
            Chain.XRP to "ripple",
            Chain.TRON to "tron",
            Chain.POLYGON to "matic-network",
            Chain.AVALANCHE to "avalanche-2"
        )
    }

    // ============================================
    // Helper Functions
    // ============================================

    /**
     * Get the appropriate RPC/API URL for a given chain
     */
    fun getApiUrl(chain: Chain): String {
        return when (chain) {
            Chain.SHAREHODL -> ShareHODL.restUrl
            Chain.BITCOIN -> Bitcoin.esploraUrl
            Chain.ETHEREUM -> EVM.Ethereum.rpcUrl()
            Chain.USDT, Chain.USDC -> EVM.Ethereum.rpcUrl() // ERC-20 on Ethereum
            Chain.BNB -> EVM.BNB.rpcUrl
            Chain.LITECOIN -> Litecoin.apiUrl
            Chain.DOGECOIN -> Dogecoin.apiUrl
            Chain.SOLANA -> Solana.rpcUrl
            Chain.XRP -> XRP.rpcUrl
            Chain.TRON -> Tron.apiUrl
            Chain.POLYGON -> EVM.Polygon.rpcUrl()
            Chain.AVALANCHE -> EVM.Avalanche.rpcUrl
        }
    }

    /**
     * Get chain ID for EVM chains
     */
    fun getEvmChainId(chain: Chain): Int {
        return when (chain) {
            Chain.ETHEREUM, Chain.USDT, Chain.USDC -> EVM.Ethereum.chainId
            Chain.BNB -> EVM.BNB.chainId
            Chain.POLYGON -> EVM.Polygon.chainId
            Chain.AVALANCHE -> EVM.Avalanche.chainId
            else -> throw IllegalArgumentException("${chain.name} is not an EVM chain")
        }
    }
}

/**
 * Fee level for transactions
 */
enum class FeeLevel {
    LOW,
    MEDIUM,
    HIGH
}

/**
 * Estimated fee for a transaction
 */
data class EstimatedFee(
    val amount: String,
    val denom: String,
    val usdValue: String? = null,
    val gasLimit: Long = 0,
    val gasPrice: String = "0"
)
