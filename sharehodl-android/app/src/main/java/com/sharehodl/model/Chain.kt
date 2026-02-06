package com.sharehodl.model

import androidx.compose.ui.graphics.Color

/**
 * Supported blockchain networks
 */
enum class Chain(
    val displayName: String,
    val symbol: String,
    val coinType: Int,  // BIP44 coin type
    val bech32Prefix: String?,
    val decimals: Int,
    val smallestUnit: String,
    val colorHex: Long,
    val explorerAddressUrl: String,
    val explorerTxUrl: String,
    val isToken: Boolean = false,  // True for tokens like USDT, USDC
    val parentChain: String? = null  // Parent chain for tokens
) {
    SHAREHODL(
        displayName = "ShareHODL",
        symbol = "HODL",
        coinType = 118,  // Cosmos
        bech32Prefix = "hodl",
        decimals = 6,
        smallestUnit = "uhodl",
        colorHex = 0xFF6366F1,  // Indigo
        explorerAddressUrl = "https://explorer.sharehodl.com/address/{address}",
        explorerTxUrl = "https://explorer.sharehodl.com/tx/{tx}"
    ),
    BITCOIN(
        displayName = "Bitcoin",
        symbol = "BTC",
        coinType = 0,
        bech32Prefix = null,  // Legacy addresses (1..., 3...)
        decimals = 8,
        smallestUnit = "satoshi",
        colorHex = 0xFFF7931A,  // Bitcoin orange
        explorerAddressUrl = "https://mempool.space/address/{address}",
        explorerTxUrl = "https://mempool.space/tx/{tx}"
    ),
    ETHEREUM(
        displayName = "Ethereum",
        symbol = "ETH",
        coinType = 60,
        bech32Prefix = null,  // Ethereum uses 0x hex addresses
        decimals = 18,
        smallestUnit = "wei",
        colorHex = 0xFF627EEA,  // Ethereum purple
        explorerAddressUrl = "https://etherscan.io/address/{address}",
        explorerTxUrl = "https://etherscan.io/tx/{tx}"
    ),
    USDT(
        displayName = "Tether",
        symbol = "USDT",
        coinType = 60,  // ERC-20 on Ethereum
        bech32Prefix = null,
        decimals = 6,
        smallestUnit = "micro",
        colorHex = 0xFF26A17B,  // Tether green
        explorerAddressUrl = "https://etherscan.io/address/{address}",
        explorerTxUrl = "https://etherscan.io/tx/{tx}",
        isToken = true,
        parentChain = "ETH"
    ),
    USDC(
        displayName = "USD Coin",
        symbol = "USDC",
        coinType = 60,  // ERC-20 on Ethereum
        bech32Prefix = null,
        decimals = 6,
        smallestUnit = "micro",
        colorHex = 0xFF2775CA,  // USDC blue
        explorerAddressUrl = "https://etherscan.io/address/{address}",
        explorerTxUrl = "https://etherscan.io/tx/{tx}",
        isToken = true,
        parentChain = "ETH"
    ),
    BNB(
        displayName = "BNB",
        symbol = "BNB",
        coinType = 714,  // BSC
        bech32Prefix = null,
        decimals = 18,
        smallestUnit = "jager",
        colorHex = 0xFFF3BA2F,  // Binance yellow
        explorerAddressUrl = "https://bscscan.com/address/{address}",
        explorerTxUrl = "https://bscscan.com/tx/{tx}"
    ),
    LITECOIN(
        displayName = "Litecoin",
        symbol = "LTC",
        coinType = 2,
        bech32Prefix = null,  // Legacy addresses (L..., M...)
        decimals = 8,
        smallestUnit = "litoshi",
        colorHex = 0xFF345D9D,  // Litecoin blue
        explorerAddressUrl = "https://blockchair.com/litecoin/address/{address}",
        explorerTxUrl = "https://blockchair.com/litecoin/transaction/{tx}"
    ),
    DOGECOIN(
        displayName = "Dogecoin",
        symbol = "DOGE",
        coinType = 3,
        bech32Prefix = null,
        decimals = 8,
        smallestUnit = "koinu",
        colorHex = 0xFFC2A633,  // Doge gold
        explorerAddressUrl = "https://dogechain.info/address/{address}",
        explorerTxUrl = "https://dogechain.info/tx/{tx}"
    ),
    SOLANA(
        displayName = "Solana",
        symbol = "SOL",
        coinType = 501,
        bech32Prefix = null,
        decimals = 9,
        smallestUnit = "lamport",
        colorHex = 0xFF9945FF,  // Solana purple
        explorerAddressUrl = "https://solscan.io/account/{address}",
        explorerTxUrl = "https://solscan.io/tx/{tx}"
    ),
    XRP(
        displayName = "XRP",
        symbol = "XRP",
        coinType = 144,
        bech32Prefix = null,
        decimals = 6,
        smallestUnit = "drop",
        colorHex = 0xFF23292F,  // XRP dark
        explorerAddressUrl = "https://xrpscan.com/account/{address}",
        explorerTxUrl = "https://xrpscan.com/tx/{tx}"
    ),
    TRON(
        displayName = "TRON",
        symbol = "TRX",
        coinType = 195,
        bech32Prefix = null,
        decimals = 6,
        smallestUnit = "sun",
        colorHex = 0xFFFF0013,  // Tron red
        explorerAddressUrl = "https://tronscan.org/#/address/{address}",
        explorerTxUrl = "https://tronscan.org/#/transaction/{tx}"
    ),
    POLYGON(
        displayName = "Polygon",
        symbol = "MATIC",
        coinType = 60,  // Uses Ethereum path
        bech32Prefix = null,
        decimals = 18,
        smallestUnit = "wei",
        colorHex = 0xFF8247E5,  // Polygon purple
        explorerAddressUrl = "https://polygonscan.com/address/{address}",
        explorerTxUrl = "https://polygonscan.com/tx/{tx}"
    ),
    AVALANCHE(
        displayName = "Avalanche",
        symbol = "AVAX",
        coinType = 9000,
        bech32Prefix = null,
        decimals = 18,
        smallestUnit = "nAVAX",
        colorHex = 0xFFE84142,  // Avalanche red
        explorerAddressUrl = "https://snowtrace.io/address/{address}",
        explorerTxUrl = "https://snowtrace.io/tx/{tx}"
    );

    val color: Color get() = Color(colorHex)

    val isCosmosStyle: Boolean get() = this == SHAREHODL
    val isBitcoinStyle: Boolean get() = this == BITCOIN || this == LITECOIN || this == DOGECOIN
    val isEthereumStyle: Boolean get() = this == ETHEREUM || this == BNB || this == POLYGON || this == AVALANCHE || isToken

    fun getExplorerUrl(address: String): String =
        explorerAddressUrl.replace("{address}", address)

    fun getTxExplorerUrl(txHash: String): String =
        explorerTxUrl.replace("{tx}", txHash)

    companion object {
        fun fromCoinType(coinType: Int): Chain? =
            entries.find { it.coinType == coinType }

        // Major chains for the main wallet view
        val majorChains = listOf(SHAREHODL, BITCOIN, ETHEREUM, USDT, USDC, BNB, LITECOIN, DOGECOIN, SOLANA, XRP, TRON, POLYGON, AVALANCHE)
    }
}

/**
 * Represents a wallet account on a specific chain
 */
data class ChainAccount(
    val chain: Chain,
    val address: String,
    val derivationPath: String,
    val balance: String = "0",
    val balanceUsd: String? = null,
    val transactions: List<CryptoTransaction> = emptyList()
) {
    val formattedBalance: String get() = "$balance ${chain.symbol}"

    val shortAddress: String get() {
        if (address.length <= 16) return address
        return "${address.take(10)}...${address.takeLast(6)}"
    }

    val explorerUrl: String get() = chain.getExplorerUrl(address)
}

/**
 * Represents a crypto transaction
 */
data class CryptoTransaction(
    val txHash: String,
    val chain: Chain,
    val type: TransactionType,
    val amount: String,
    val amountUsd: String? = null,
    val fromAddress: String,
    val toAddress: String,
    val timestamp: Long,
    val status: TransactionStatus,
    val fee: String? = null,
    val confirmations: Int = 0
) {
    val formattedAmount: String get() = "$amount ${chain.symbol}"

    val shortTxHash: String get() {
        if (txHash.length <= 16) return txHash
        return "${txHash.take(8)}...${txHash.takeLast(8)}"
    }

    val explorerUrl: String get() = chain.getTxExplorerUrl(txHash)

    val formattedDate: String get() {
        val date = java.util.Date(timestamp)
        val format = java.text.SimpleDateFormat("MMM dd, yyyy HH:mm", java.util.Locale.getDefault())
        return format.format(date)
    }

    val isReceived: Boolean get() = type == TransactionType.RECEIVE
}

enum class TransactionType {
    SEND,
    RECEIVE,
    SWAP,
    CONTRACT_CALL
}

enum class TransactionStatus {
    PENDING,
    CONFIRMED,
    FAILED
}
