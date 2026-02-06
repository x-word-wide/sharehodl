package com.sharehodl.ui.trading

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.sharehodl.viewmodel.WalletViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TradingScreen(
    viewModel: WalletViewModel
) {
    var selectedTab by remember { mutableIntStateOf(0) }
    var orderType by remember { mutableStateOf(OrderType.Buy) }

    Scaffold(
        containerColor = MaterialTheme.colorScheme.background
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            // Price Header
            item {
                PriceHeader()
            }

            // Market Stats
            item {
                MarketStatsRow()
            }

            // Mini Chart Area
            item {
                ChartPlaceholder()
            }

            // Tabs
            item {
                TabRow(
                    selectedTabIndex = selectedTab,
                    containerColor = Color.Transparent,
                    modifier = Modifier.padding(horizontal = 16.dp)
                ) {
                    Tab(
                        selected = selectedTab == 0,
                        onClick = { selectedTab = 0 },
                        text = { Text("Trade") }
                    )
                    Tab(
                        selected = selectedTab == 1,
                        onClick = { selectedTab = 1 },
                        text = { Text("Order Book") }
                    )
                    Tab(
                        selected = selectedTab == 2,
                        onClick = { selectedTab = 2 },
                        text = { Text("My Orders") }
                    )
                }
            }

            // Content
            when (selectedTab) {
                0 -> item {
                    TradeForm(
                        orderType = orderType,
                        onOrderTypeChange = { orderType = it },
                        onPlaceOrder = { /* Handle order */ }
                    )
                }
                1 -> item {
                    OrderBookView()
                }
                2 -> item {
                    MyOrdersView()
                }
            }
        }
    }
}

@Composable
fun PriceHeader() {
    val isPositive = true
    val changeColor = if (isPositive) Color(0xFF22C55E) else Color(0xFFEF4444)

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(20.dp)
    ) {
        // Trading Pair
        Row(
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Token Icon
            Box(
                modifier = Modifier
                    .size(40.dp)
                    .clip(CircleShape)
                    .background(MaterialTheme.colorScheme.primary),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "H",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
            }
            Spacer(modifier = Modifier.width(12.dp))
            Column {
                Text(
                    text = "HODL/USDC",
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold
                )
                Text(
                    text = "ShareHODL Token",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        // Price
        Row(
            verticalAlignment = Alignment.Bottom
        ) {
            Text(
                text = "$1.00",
                style = MaterialTheme.typography.headlineLarge.copy(
                    fontWeight = FontWeight.Bold,
                    fontSize = 36.sp
                )
            )
            Spacer(modifier = Modifier.width(12.dp))
            Surface(
                color = changeColor.copy(alpha = 0.1f),
                shape = RoundedCornerShape(6.dp)
            ) {
                Row(
                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        if (isPositive) Icons.Default.ArrowDropUp else Icons.Default.ArrowDropDown,
                        contentDescription = null,
                        modifier = Modifier.size(18.dp),
                        tint = changeColor
                    )
                    Text(
                        text = "+0.00%",
                        style = MaterialTheme.typography.bodyMedium,
                        fontWeight = FontWeight.SemiBold,
                        color = changeColor
                    )
                }
            }
        }
    }
}

@Composable
fun MarketStatsRow() {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 20.dp),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        MarketStatItem(
            label = "24h High",
            value = "$1.02",
            valueColor = Color(0xFF22C55E)
        )
        MarketStatItem(
            label = "24h Low",
            value = "$0.98",
            valueColor = Color(0xFFEF4444)
        )
        MarketStatItem(
            label = "24h Volume",
            value = "$125.4K",
            valueColor = MaterialTheme.colorScheme.onSurface
        )
    }
}

@Composable
fun MarketStatItem(
    label: String,
    value: String,
    valueColor: Color
) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = value,
            style = MaterialTheme.typography.bodyLarge,
            fontWeight = FontWeight.SemiBold,
            color = valueColor
        )
    }
}

@Composable
fun ChartPlaceholder() {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .height(180.dp)
            .padding(20.dp)
            .clip(RoundedCornerShape(16.dp))
            .background(
                Brush.verticalGradient(
                    colors = listOf(
                        Color(0xFF22C55E).copy(alpha = 0.1f),
                        Color(0xFF22C55E).copy(alpha = 0.02f)
                    )
                )
            ),
        contentAlignment = Alignment.Center
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Icon(
                Icons.Default.ShowChart,
                contentDescription = null,
                modifier = Modifier.size(48.dp),
                tint = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.5f)
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "Price Chart",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

enum class OrderType { Buy, Sell }

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TradeForm(
    orderType: OrderType,
    onOrderTypeChange: (OrderType) -> Unit,
    onPlaceOrder: () -> Unit
) {
    var price by remember { mutableStateOf("") }
    var amount by remember { mutableStateOf("") }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(20.dp)
    ) {
        // Buy/Sell Toggle
        Surface(
            shape = RoundedCornerShape(12.dp),
            color = MaterialTheme.colorScheme.surfaceVariant
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(4.dp)
            ) {
                // Buy Button
                Surface(
                    onClick = { onOrderTypeChange(OrderType.Buy) },
                    shape = RoundedCornerShape(10.dp),
                    color = if (orderType == OrderType.Buy) Color(0xFF22C55E) else Color.Transparent,
                    modifier = Modifier.weight(1f)
                ) {
                    Text(
                        text = "Buy",
                        modifier = Modifier.padding(vertical = 12.dp),
                        style = MaterialTheme.typography.bodyLarge,
                        fontWeight = FontWeight.SemiBold,
                        textAlign = TextAlign.Center,
                        color = if (orderType == OrderType.Buy) Color.White else MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }

                // Sell Button
                Surface(
                    onClick = { onOrderTypeChange(OrderType.Sell) },
                    shape = RoundedCornerShape(10.dp),
                    color = if (orderType == OrderType.Sell) Color(0xFFEF4444) else Color.Transparent,
                    modifier = Modifier.weight(1f)
                ) {
                    Text(
                        text = "Sell",
                        modifier = Modifier.padding(vertical = 12.dp),
                        style = MaterialTheme.typography.bodyLarge,
                        fontWeight = FontWeight.SemiBold,
                        textAlign = TextAlign.Center,
                        color = if (orderType == OrderType.Sell) Color.White else MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }

        Spacer(modifier = Modifier.height(20.dp))

        // Price Input
        OutlinedTextField(
            value = price,
            onValueChange = { price = it },
            label = { Text("Price") },
            placeholder = { Text("0.00") },
            suffix = { Text("USDC") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
            shape = RoundedCornerShape(12.dp)
        )

        Spacer(modifier = Modifier.height(12.dp))

        // Amount Input
        OutlinedTextField(
            value = amount,
            onValueChange = { amount = it },
            label = { Text("Amount") },
            placeholder = { Text("0.00") },
            suffix = { Text("HODL") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true,
            shape = RoundedCornerShape(12.dp)
        )

        Spacer(modifier = Modifier.height(16.dp))

        // Total
        Surface(
            color = MaterialTheme.colorScheme.surfaceVariant,
            shape = RoundedCornerShape(12.dp)
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "Total",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                val total = (price.toDoubleOrNull() ?: 0.0) * (amount.toDoubleOrNull() ?: 0.0)
                Text(
                    text = String.format("%.2f USDC", total),
                    style = MaterialTheme.typography.bodyLarge,
                    fontWeight = FontWeight.Bold
                )
            }
        }

        Spacer(modifier = Modifier.height(20.dp))

        // Place Order Button
        Button(
            onClick = onPlaceOrder,
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp),
            colors = ButtonDefaults.buttonColors(
                containerColor = if (orderType == OrderType.Buy) Color(0xFF22C55E) else Color(0xFFEF4444)
            ),
            enabled = price.isNotEmpty() && amount.isNotEmpty(),
            shape = RoundedCornerShape(12.dp)
        ) {
            Text(
                text = if (orderType == OrderType.Buy) "Place Buy Order" else "Place Sell Order",
                style = MaterialTheme.typography.bodyLarge,
                fontWeight = FontWeight.SemiBold
            )
        }

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = "Trading fee: 0.3%",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.align(Alignment.CenterHorizontally)
        )
    }
}

@Composable
fun OrderBookView() {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(20.dp)
    ) {
        // Header
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Text(
                "Price (USDC)",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Text(
                "Amount (HODL)",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Text(
                "Total",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }

        Spacer(modifier = Modifier.height(12.dp))

        // Sell Orders (Red)
        repeat(5) { index ->
            OrderBookRow(
                price = String.format("%.4f", 1.0 + (5 - index) * 0.001),
                amount = String.format("%.2f", (100..500).random().toDouble()),
                total = String.format("%.2f", (100..500).random().toDouble()),
                isBuy = false,
                fillPercent = (0.2f + index * 0.15f).coerceAtMost(1f)
            )
        }

        // Spread
        Surface(
            color = MaterialTheme.colorScheme.surfaceVariant,
            shape = RoundedCornerShape(8.dp),
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 8.dp)
        ) {
            Text(
                text = "Spread: $0.0010 (0.10%)",
                modifier = Modifier.padding(12.dp),
                style = MaterialTheme.typography.bodySmall,
                textAlign = TextAlign.Center,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }

        // Buy Orders (Green)
        repeat(5) { index ->
            OrderBookRow(
                price = String.format("%.4f", 1.0 - (index + 1) * 0.001),
                amount = String.format("%.2f", (100..500).random().toDouble()),
                total = String.format("%.2f", (100..500).random().toDouble()),
                isBuy = true,
                fillPercent = (0.2f + index * 0.15f).coerceAtMost(1f)
            )
        }
    }
}

@Composable
fun OrderBookRow(
    price: String,
    amount: String,
    total: String,
    isBuy: Boolean,
    fillPercent: Float
) {
    val color = if (isBuy) Color(0xFF22C55E) else Color(0xFFEF4444)

    Box(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp)
    ) {
        // Background fill
        Box(
            modifier = Modifier
                .fillMaxWidth(fillPercent)
                .height(32.dp)
                .align(if (isBuy) Alignment.CenterStart else Alignment.CenterEnd)
                .background(color.copy(alpha = 0.1f))
        )

        // Content
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 6.dp),
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Text(
                text = price,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
                color = color
            )
            Text(
                text = amount,
                style = MaterialTheme.typography.bodyMedium
            )
            Text(
                text = total,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@Composable
fun MyOrdersView() {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(40.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Box(
            modifier = Modifier
                .size(80.dp)
                .clip(CircleShape)
                .background(MaterialTheme.colorScheme.surfaceVariant),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                Icons.Default.Receipt,
                contentDescription = null,
                modifier = Modifier.size(40.dp),
                tint = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = "No open orders",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Medium
        )
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = "Your orders will appear here",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}
