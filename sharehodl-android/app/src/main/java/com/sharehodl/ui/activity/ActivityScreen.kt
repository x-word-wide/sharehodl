package com.sharehodl.ui.activity

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.outlined.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.hapticfeedback.HapticFeedbackType
import androidx.compose.ui.platform.LocalHapticFeedback
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.sharehodl.data.DemoData
import com.sharehodl.model.*
import com.sharehodl.ui.portfolio.*
import java.text.SimpleDateFormat
import java.util.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ActivityScreen(
    onOrderClick: (String) -> Unit = {},
    onTradeClick: (String) -> Unit = {}
) {
    val haptics = LocalHapticFeedback.current

    val trades = remember { DemoData.trades }
    val orders = remember { DemoData.orders }
    val dividends = remember { DemoData.dividends }

    var selectedTab by remember { mutableStateOf(0) }
    val tabs = listOf("Trades", "Orders", "Dividends")

    Scaffold(
        containerColor = DarkBackground,
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        "Activity",
                        fontWeight = FontWeight.Bold
                    )
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = DarkBackground,
                    titleContentColor = Color.White
                ),
                actions = {
                    IconButton(onClick = { /* Export */ }) {
                        Icon(
                            Icons.Outlined.Download,
                            contentDescription = "Export",
                            tint = Color.White
                        )
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            // Tab Selector
            TabRow(
                selectedTabIndex = selectedTab,
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                containerColor = CardBackground,
                contentColor = Color.White
            ) {
                tabs.forEachIndexed { index, title ->
                    Tab(
                        selected = selectedTab == index,
                        onClick = {
                            selectedTab = index
                            haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                        },
                        text = {
                            Text(
                                title,
                                fontWeight = if (selectedTab == index) FontWeight.SemiBold else FontWeight.Normal
                            )
                        }
                    )
                }
            }

            // Content based on selected tab
            LazyColumn(
                modifier = Modifier.fillMaxSize()
            ) {
                when (selectedTab) {
                    0 -> {
                        // Trades
                        if (trades.isEmpty()) {
                            item { EmptyActivityCard(type = "trades") }
                        } else {
                            items(trades) { trade ->
                                TradeHistoryItem(
                                    trade = trade,
                                    onClick = { onTradeClick(trade.id) }
                                )
                            }
                        }
                    }
                    1 -> {
                        // Open Orders
                        val openOrders = orders.filter { it.status == OrderStatus.PENDING || it.status == OrderStatus.PARTIAL }
                        if (openOrders.isEmpty()) {
                            item { EmptyActivityCard(type = "orders") }
                        } else {
                            items(openOrders) { order ->
                                OrderItem(
                                    order = order,
                                    onClick = { onOrderClick(order.id) }
                                )
                            }
                        }
                    }
                    2 -> {
                        // Dividends
                        if (dividends.isEmpty()) {
                            item { EmptyActivityCard(type = "dividends") }
                        } else {
                            items(dividends) { dividend ->
                                DividendItem(dividend = dividend)
                            }
                        }
                    }
                }

                // Bottom spacing
                item {
                    Spacer(modifier = Modifier.height(100.dp))
                }
            }
        }
    }
}

@Composable
fun TradeHistoryItem(
    trade: Trade,
    onClick: () -> Unit
) {
    val equity = DemoData.getEquity(trade.symbol)
    val isBuy = trade.side == OrderSide.BUY
    val sideColor = if (isBuy) GainGreen else LossRed
    val dateFormat = remember { SimpleDateFormat("MMM dd, yyyy HH:mm", Locale.getDefault()) }

    Surface(
        onClick = onClick,
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 6.dp),
        color = CardBackground,
        shape = RoundedCornerShape(16.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Icon
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .clip(CircleShape)
                    .background(sideColor.copy(alpha = 0.15f)),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    if (isBuy) Icons.Default.Add else Icons.Default.Remove,
                    contentDescription = null,
                    modifier = Modifier.size(24.dp),
                    tint = sideColor
                )
            }

            Spacer(modifier = Modifier.width(14.dp))

            // Trade Info
            Column(modifier = Modifier.weight(1f)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        text = if (isBuy) "Bought" else "Sold",
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.SemiBold,
                        color = Color.White
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = trade.symbol,
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.Bold,
                        color = equity?.sector?.color ?: PrimaryBlue
                    )
                }
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "${String.format("%.4f", trade.quantity).trimEnd('0').trimEnd('.')} shares @ $${String.format("%.2f", trade.price)}",
                    style = MaterialTheme.typography.bodySmall,
                    color = Color.White.copy(alpha = 0.6f)
                )
                Spacer(modifier = Modifier.height(2.dp))
                Text(
                    text = dateFormat.format(Date(trade.timestamp)),
                    style = MaterialTheme.typography.labelSmall,
                    color = Color.White.copy(alpha = 0.4f)
                )
            }

            // Total
            Column(horizontalAlignment = Alignment.End) {
                Text(
                    text = trade.formattedTotal,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                    color = Color.White
                )
                Text(
                    text = "Fee: $${String.format("%.2f", trade.fee)}",
                    style = MaterialTheme.typography.labelSmall,
                    color = Color.White.copy(alpha = 0.4f)
                )
            }
        }
    }
}

@Composable
fun OrderItem(
    order: Order,
    onClick: () -> Unit
) {
    val equity = DemoData.getEquity(order.symbol)
    val isBuy = order.side == OrderSide.BUY
    val sideColor = if (isBuy) GainGreen else LossRed
    val dateFormat = remember { SimpleDateFormat("MMM dd, HH:mm", Locale.getDefault()) }

    Surface(
        onClick = onClick,
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 6.dp),
        color = CardBackground,
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Icon
                Box(
                    modifier = Modifier
                        .size(48.dp)
                        .clip(CircleShape)
                        .background(sideColor.copy(alpha = 0.15f)),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        if (isBuy) Icons.Default.ShoppingCart else Icons.Default.Sell,
                        contentDescription = null,
                        modifier = Modifier.size(24.dp),
                        tint = sideColor
                    )
                }

                Spacer(modifier = Modifier.width(14.dp))

                // Order Info
                Column(modifier = Modifier.weight(1f)) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Surface(
                            color = sideColor.copy(alpha = 0.15f),
                            shape = RoundedCornerShape(4.dp)
                        ) {
                            Text(
                                text = if (isBuy) "BUY" else "SELL",
                                style = MaterialTheme.typography.labelSmall,
                                fontWeight = FontWeight.Bold,
                                color = sideColor,
                                modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp)
                            )
                        }
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = order.symbol,
                            style = MaterialTheme.typography.titleSmall,
                            fontWeight = FontWeight.Bold,
                            color = Color.White
                        )
                    }
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = "${order.type.name.lowercase().replaceFirstChar { it.uppercase() }} order",
                        style = MaterialTheme.typography.bodySmall,
                        color = Color.White.copy(alpha = 0.5f)
                    )
                }

                // Status Badge
                Surface(
                    color = when (order.status) {
                        OrderStatus.PENDING -> Color(0xFFF59E0B).copy(alpha = 0.15f)
                        OrderStatus.PARTIAL -> PrimaryBlue.copy(alpha = 0.15f)
                        OrderStatus.FILLED -> GainGreen.copy(alpha = 0.15f)
                        OrderStatus.CANCELLED -> NeutralGray.copy(alpha = 0.15f)
                    },
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Text(
                        text = order.status.name.lowercase().replaceFirstChar { it.uppercase() },
                        style = MaterialTheme.typography.labelSmall,
                        fontWeight = FontWeight.SemiBold,
                        color = when (order.status) {
                            OrderStatus.PENDING -> Color(0xFFF59E0B)
                            OrderStatus.PARTIAL -> PrimaryBlue
                            OrderStatus.FILLED -> GainGreen
                            OrderStatus.CANCELLED -> NeutralGray
                        },
                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp)
                    )
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // Order Details
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                OrderDetailColumn(label = "Quantity", value = order.formattedQuantity)
                OrderDetailColumn(label = "Price", value = order.formattedPrice)
                OrderDetailColumn(label = "Total", value = order.formattedTotal)
            }

            Spacer(modifier = Modifier.height(12.dp))

            // Cancel button
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Created ${dateFormat.format(Date(order.createdAt))}",
                    style = MaterialTheme.typography.labelSmall,
                    color = Color.White.copy(alpha = 0.4f)
                )

                TextButton(
                    onClick = { /* Cancel order */ },
                    colors = ButtonDefaults.textButtonColors(contentColor = LossRed)
                ) {
                    Text("Cancel Order")
                }
            }
        }
    }
}

@Composable
fun OrderDetailColumn(label: String, value: String) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = Color.White.copy(alpha = 0.5f)
        )
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = value,
            style = MaterialTheme.typography.bodyMedium,
            fontWeight = FontWeight.Medium,
            color = Color.White
        )
    }
}

@Composable
fun DividendItem(dividend: Dividend) {
    val equity = DemoData.getEquity(dividend.symbol)
    val dateFormat = remember { SimpleDateFormat("MMM dd, yyyy", Locale.getDefault()) }

    val statusColor = when (dividend.status) {
        DividendStatus.ANNOUNCED -> Color(0xFFF59E0B)
        DividendStatus.EX_DATE_PASSED -> PrimaryBlue
        DividendStatus.PAID -> GainGreen
    }

    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 6.dp),
        color = CardBackground,
        shape = RoundedCornerShape(16.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Icon
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .clip(CircleShape)
                    .background(statusColor.copy(alpha = 0.15f)),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    Icons.Outlined.Payments,
                    contentDescription = null,
                    modifier = Modifier.size(24.dp),
                    tint = statusColor
                )
            }

            Spacer(modifier = Modifier.width(14.dp))

            // Dividend Info
            Column(modifier = Modifier.weight(1f)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        text = dividend.symbol,
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.Bold,
                        color = equity?.sector?.color ?: PrimaryBlue
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Surface(
                        color = statusColor.copy(alpha = 0.15f),
                        shape = RoundedCornerShape(4.dp)
                    ) {
                        Text(
                            text = when (dividend.status) {
                                DividendStatus.ANNOUNCED -> "Upcoming"
                                DividendStatus.EX_DATE_PASSED -> "Ex-Date Passed"
                                DividendStatus.PAID -> "Paid"
                            },
                            style = MaterialTheme.typography.labelSmall,
                            fontWeight = FontWeight.Medium,
                            color = statusColor,
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 2.dp)
                        )
                    }
                }
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "${dividend.frequency.name.lowercase().replaceFirstChar { it.uppercase() }} dividend",
                    style = MaterialTheme.typography.bodySmall,
                    color = Color.White.copy(alpha = 0.6f)
                )
                Spacer(modifier = Modifier.height(2.dp))
                Text(
                    text = "Pay date: ${dateFormat.format(Date(dividend.payDate))}",
                    style = MaterialTheme.typography.labelSmall,
                    color = Color.White.copy(alpha = 0.4f)
                )
            }

            // Amount
            Column(horizontalAlignment = Alignment.End) {
                Text(
                    text = dividend.formattedAmount,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                    color = Color.White
                )
                Text(
                    text = "per share",
                    style = MaterialTheme.typography.labelSmall,
                    color = Color.White.copy(alpha = 0.4f)
                )
            }
        }
    }
}

@Composable
fun EmptyActivityCard(type: String) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(48.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Box(
            modifier = Modifier
                .size(80.dp)
                .clip(CircleShape)
                .background(SurfaceColor),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                when (type) {
                    "trades" -> Icons.Outlined.Receipt
                    "orders" -> Icons.Outlined.Description
                    else -> Icons.Outlined.Payments
                },
                contentDescription = null,
                modifier = Modifier.size(40.dp),
                tint = Color.White.copy(alpha = 0.5f)
            )
        }
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = when (type) {
                "trades" -> "No Trade History"
                "orders" -> "No Open Orders"
                else -> "No Dividends"
            },
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
            color = Color.White
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = when (type) {
                "trades" -> "Your completed trades will appear here"
                "orders" -> "Your pending orders will appear here"
                else -> "Dividend payments will appear here"
            },
            style = MaterialTheme.typography.bodyMedium,
            color = Color.White.copy(alpha = 0.6f),
            textAlign = TextAlign.Center
        )
    }
}
