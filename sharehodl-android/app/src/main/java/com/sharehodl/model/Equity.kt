package com.sharehodl.model

import androidx.compose.ui.graphics.Color

/**
 * Represents a tokenized equity (stock/share) on the ShareHODL blockchain
 */
data class Equity(
    val symbol: String,
    val companyName: String,
    val sector: Sector,
    val currentPrice: Double,
    val priceChange24h: Double,
    val priceChangePercent24h: Double,
    val marketCap: Long,
    val volume24h: Long,
    val high52Week: Double,
    val low52Week: Double,
    val dividendYield: Double? = null,
    val peRatio: Double? = null,
    val logoUrl: String? = null
) {
    val isPositiveChange: Boolean get() = priceChangePercent24h >= 0

    val formattedPrice: String get() = "$${String.format("%,.2f", currentPrice)}"

    val formattedChange: String get() = buildString {
        if (priceChangePercent24h >= 0) append("+")
        append(String.format("%.2f", priceChangePercent24h))
        append("%")
    }

    val formattedMarketCap: String get() = when {
        marketCap >= 1_000_000_000_000 -> "$${String.format("%.1f", marketCap / 1_000_000_000_000.0)}T"
        marketCap >= 1_000_000_000 -> "$${String.format("%.1f", marketCap / 1_000_000_000.0)}B"
        marketCap >= 1_000_000 -> "$${String.format("%.1f", marketCap / 1_000_000.0)}M"
        else -> "$${String.format("%,d", marketCap)}"
    }

    val formattedVolume: String get() = when {
        volume24h >= 1_000_000_000 -> "${String.format("%.1f", volume24h / 1_000_000_000.0)}B"
        volume24h >= 1_000_000 -> "${String.format("%.1f", volume24h / 1_000_000.0)}M"
        volume24h >= 1_000 -> "${String.format("%.1f", volume24h / 1_000.0)}K"
        else -> volume24h.toString()
    }
}

/**
 * User's holding of a specific equity
 */
data class EquityHolding(
    val equity: Equity,
    val shares: Double,
    val averageCost: Double,
    val purchaseDate: Long
) {
    val currentValue: Double get() = shares * equity.currentPrice
    val totalCost: Double get() = shares * averageCost
    val totalGainLoss: Double get() = currentValue - totalCost
    val totalGainLossPercent: Double get() = if (totalCost > 0) ((currentValue - totalCost) / totalCost) * 100 else 0.0
    val isProfit: Boolean get() = totalGainLoss >= 0

    val formattedShares: String get() = if (shares % 1 == 0.0) {
        shares.toInt().toString()
    } else {
        String.format("%.4f", shares)
    }

    val formattedValue: String get() = "$${String.format("%,.2f", currentValue)}"

    val formattedGainLoss: String get() = buildString {
        if (totalGainLoss >= 0) append("+")
        append("$${String.format("%,.2f", totalGainLoss)}")
    }

    val formattedGainLossPercent: String get() = buildString {
        if (totalGainLossPercent >= 0) append("+")
        append(String.format("%.2f", totalGainLossPercent))
        append("%")
    }
}

/**
 * Company profile with detailed information
 */
data class Company(
    val symbol: String,
    val name: String,
    val description: String,
    val sector: Sector,
    val industry: String,
    val employees: Int,
    val headquarters: String,
    val founded: Int,
    val ceo: String,
    val website: String,
    val totalShares: Long,
    val publicFloat: Long,
    val insiderOwnership: Double,
    val institutionalOwnership: Double
) {
    val formattedEmployees: String get() = when {
        employees >= 1_000_000 -> "${String.format("%.1f", employees / 1_000_000.0)}M"
        employees >= 1_000 -> "${String.format("%.1f", employees / 1_000.0)}K"
        else -> employees.toString()
    }
}

/**
 * Market sectors for categorization
 */
enum class Sector(
    val displayName: String,
    val colorHex: Long
) {
    TECHNOLOGY("Technology", 0xFF3B82F6),
    HEALTHCARE("Healthcare", 0xFF10B981),
    FINANCE("Finance", 0xFF8B5CF6),
    CONSUMER("Consumer", 0xFFF59E0B),
    ENERGY("Energy", 0xFFEF4444),
    INDUSTRIAL("Industrial", 0xFF6B7280),
    REAL_ESTATE("Real Estate", 0xFF14B8A6),
    UTILITIES("Utilities", 0xFF06B6D4),
    MATERIALS("Materials", 0xFFEC4899),
    COMMUNICATIONS("Communications", 0xFF6366F1);

    val color: Color get() = Color(colorHex)
}

/**
 * Dividend payment record
 */
data class Dividend(
    val symbol: String,
    val amount: Double,
    val exDate: Long,
    val payDate: Long,
    val recordDate: Long,
    val frequency: DividendFrequency,
    val status: DividendStatus
) {
    val formattedAmount: String get() = "$${String.format("%.4f", amount)}"
}

enum class DividendFrequency {
    MONTHLY, QUARTERLY, SEMI_ANNUAL, ANNUAL
}

enum class DividendStatus {
    ANNOUNCED, EX_DATE_PASSED, PAID
}

/**
 * Trade order for the order book
 */
data class Order(
    val id: String,
    val symbol: String,
    val side: OrderSide,
    val type: OrderType,
    val quantity: Double,
    val price: Double,
    val filledQuantity: Double = 0.0,
    val status: OrderStatus,
    val createdAt: Long,
    val updatedAt: Long
) {
    val isFullyFilled: Boolean get() = filledQuantity >= quantity
    val remainingQuantity: Double get() = quantity - filledQuantity

    val formattedQuantity: String get() = if (quantity % 1 == 0.0) {
        quantity.toInt().toString()
    } else {
        String.format("%.4f", quantity)
    }

    val formattedPrice: String get() = "$${String.format("%.2f", price)}"

    val formattedTotal: String get() = "$${String.format("%,.2f", quantity * price)}"
}

enum class OrderSide {
    BUY, SELL
}

enum class OrderType {
    MARKET, LIMIT
}

enum class OrderStatus {
    PENDING, PARTIAL, FILLED, CANCELLED
}

/**
 * Trade execution record
 */
data class Trade(
    val id: String,
    val symbol: String,
    val side: OrderSide,
    val quantity: Double,
    val price: Double,
    val fee: Double,
    val timestamp: Long
) {
    val total: Double get() = quantity * price
    val formattedTotal: String get() = "$${String.format("%,.2f", total)}"
}

/**
 * Market statistics
 */
data class MarketStats(
    val totalMarketCap: Long,
    val totalVolume24h: Long,
    val activeCompanies: Int,
    val topGainer: Equity?,
    val topLoser: Equity?,
    val mostActive: Equity?
)

/**
 * Portfolio summary
 */
data class PortfolioSummary(
    val totalValue: Double,
    val totalCost: Double,
    val dayChange: Double,
    val dayChangePercent: Double,
    val totalGainLoss: Double,
    val totalGainLossPercent: Double,
    val dividendsEarned: Double
) {
    val isPositiveDay: Boolean get() = dayChange >= 0
    val isPositiveTotal: Boolean get() = totalGainLoss >= 0

    val formattedTotalValue: String get() = "$${String.format("%,.2f", totalValue)}"
    val formattedDayChange: String get() = buildString {
        if (dayChange >= 0) append("+")
        append("$${String.format("%,.2f", dayChange)}")
    }
    val formattedDayChangePercent: String get() = buildString {
        if (dayChangePercent >= 0) append("+")
        append(String.format("%.2f", dayChangePercent))
        append("%")
    }
}

/**
 * Watchlist item
 */
data class WatchlistItem(
    val equity: Equity,
    val addedAt: Long,
    val alertPrice: Double? = null,
    val notes: String? = null
)
