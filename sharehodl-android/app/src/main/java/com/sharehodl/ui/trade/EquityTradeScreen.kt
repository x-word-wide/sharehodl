package com.sharehodl.ui.trade

import androidx.compose.animation.*
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
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
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.sharehodl.data.DemoData
import com.sharehodl.model.*
import com.sharehodl.ui.portfolio.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EquityTradeScreen(
    preSelectedSymbol: String? = null,
    onBack: (() -> Unit)? = null
) {
    val haptics = LocalHapticFeedback.current

    val equities = remember { DemoData.equities }
    val holdings = remember { DemoData.holdings }

    var selectedEquity by remember {
        mutableStateOf(
            preSelectedSymbol?.let { DemoData.getEquity(it) } ?: equities.firstOrNull()
        )
    }
    var isBuy by remember { mutableStateOf(true) }
    var orderType by remember { mutableStateOf(OrderType.MARKET) }
    var quantity by remember { mutableStateOf("") }
    var limitPrice by remember { mutableStateOf("") }

    var showEquitySelector by remember { mutableStateOf(false) }

    val quantityDouble = quantity.toDoubleOrNull() ?: 0.0
    val priceDouble = if (orderType == OrderType.LIMIT) {
        limitPrice.toDoubleOrNull() ?: 0.0
    } else {
        selectedEquity?.currentPrice ?: 0.0
    }
    val estimatedTotal = quantityDouble * priceDouble
    val tradingFee = estimatedTotal * 0.003 // 0.3% fee

    // Get user's holding for selected equity
    val userHolding = holdings.find { it.equity.symbol == selectedEquity?.symbol }

    Scaffold(
        containerColor = DarkBackground,
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        if (isBuy) "Buy Stock" else "Sell Stock",
                        fontWeight = FontWeight.Bold
                    )
                },
                navigationIcon = {
                    if (onBack != null) {
                        IconButton(onClick = onBack) {
                            Icon(
                                Icons.Default.ArrowBack,
                                contentDescription = "Back",
                                tint = Color.White
                            )
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = DarkBackground,
                    titleContentColor = Color.White
                )
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            LazyColumn(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
            ) {
                // Buy/Sell Toggle
                item {
                    BuySellToggle(
                        isBuy = isBuy,
                        onToggle = {
                            isBuy = it
                            haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                        }
                    )
                }

                // Equity Selector
                item {
                    EquitySelectorCard(
                        equity = selectedEquity,
                        onClick = { showEquitySelector = true }
                    )
                }

                // Order Type Selector
                item {
                    OrderTypeSelector(
                        orderType = orderType,
                        onSelect = {
                            orderType = it
                            haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                        }
                    )
                }

                // Quantity Input
                item {
                    QuantityInput(
                        quantity = quantity,
                        onQuantityChange = { quantity = it },
                        maxQuantity = if (!isBuy && userHolding != null) userHolding.shares else null
                    )
                }

                // Limit Price (if limit order)
                if (orderType == OrderType.LIMIT) {
                    item {
                        LimitPriceInput(
                            price = limitPrice,
                            onPriceChange = { limitPrice = it },
                            currentPrice = selectedEquity?.currentPrice ?: 0.0
                        )
                    }
                }

                // Order Summary
                item {
                    OrderSummaryCard(
                        equity = selectedEquity,
                        isBuy = isBuy,
                        quantity = quantityDouble,
                        price = priceDouble,
                        total = estimatedTotal,
                        fee = tradingFee
                    )
                }

                // Available Balance / Holdings
                item {
                    AvailableBalanceCard(
                        isBuy = isBuy,
                        holding = userHolding
                    )
                }
            }

            // Submit Button
            Surface(
                color = CardBackground,
                tonalElevation = 8.dp
            ) {
                Button(
                    onClick = {
                        haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                        // Submit order
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp)
                        .height(56.dp),
                    enabled = quantityDouble > 0 && selectedEquity != null &&
                            (orderType == OrderType.MARKET || priceDouble > 0),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = if (isBuy) GainGreen else LossRed
                    ),
                    shape = RoundedCornerShape(14.dp)
                ) {
                    Text(
                        text = if (isBuy) "Place Buy Order" else "Place Sell Order",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }
        }
    }

    // Equity Selector Bottom Sheet
    if (showEquitySelector) {
        EquitySelectorBottomSheet(
            equities = equities,
            selectedSymbol = selectedEquity?.symbol,
            onSelect = { equity ->
                selectedEquity = equity
                showEquitySelector = false
            },
            onDismiss = { showEquitySelector = false }
        )
    }
}

@Composable
fun BuySellToggle(
    isBuy: Boolean,
    onToggle: (Boolean) -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        colors = CardDefaults.cardColors(containerColor = CardBackground),
        shape = RoundedCornerShape(16.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(6.dp)
        ) {
            Surface(
                onClick = { onToggle(true) },
                modifier = Modifier.weight(1f),
                color = if (isBuy) GainGreen else Color.Transparent,
                shape = RoundedCornerShape(12.dp)
            ) {
                Box(
                    modifier = Modifier.padding(vertical = 14.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "Buy",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        color = if (isBuy) Color.White else Color.White.copy(alpha = 0.5f)
                    )
                }
            }

            Surface(
                onClick = { onToggle(false) },
                modifier = Modifier.weight(1f),
                color = if (!isBuy) LossRed else Color.Transparent,
                shape = RoundedCornerShape(12.dp)
            ) {
                Box(
                    modifier = Modifier.padding(vertical = 14.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "Sell",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        color = if (!isBuy) Color.White else Color.White.copy(alpha = 0.5f)
                    )
                }
            }
        }
    }
}

@Composable
fun EquitySelectorCard(
    equity: Equity?,
    onClick: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        colors = CardDefaults.cardColors(containerColor = CardBackground),
        shape = RoundedCornerShape(16.dp),
        onClick = onClick
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            if (equity != null) {
                Box(
                    modifier = Modifier
                        .size(48.dp)
                        .clip(CircleShape)
                        .background(equity.sector.color.copy(alpha = 0.2f)),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = equity.symbol.take(2),
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold,
                        color = equity.sector.color
                    )
                }

                Spacer(modifier = Modifier.width(14.dp))

                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = equity.symbol,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold,
                        color = Color.White
                    )
                    Text(
                        text = equity.companyName,
                        style = MaterialTheme.typography.bodySmall,
                        color = Color.White.copy(alpha = 0.6f)
                    )
                }

                Column(horizontalAlignment = Alignment.End) {
                    Text(
                        text = equity.formattedPrice,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        color = Color.White
                    )
                    val changeColor = if (equity.isPositiveChange) GainGreen else LossRed
                    Text(
                        text = equity.formattedChange,
                        style = MaterialTheme.typography.bodySmall,
                        color = changeColor
                    )
                }

                Spacer(modifier = Modifier.width(8.dp))

                Icon(
                    Icons.Default.KeyboardArrowDown,
                    contentDescription = "Select",
                    tint = Color.White.copy(alpha = 0.5f)
                )
            } else {
                Text(
                    text = "Select a stock",
                    style = MaterialTheme.typography.bodyLarge,
                    color = Color.White.copy(alpha = 0.5f),
                    modifier = Modifier.weight(1f)
                )
                Icon(
                    Icons.Default.KeyboardArrowDown,
                    contentDescription = "Select",
                    tint = Color.White.copy(alpha = 0.5f)
                )
            }
        }
    }
}

@Composable
fun OrderTypeSelector(
    orderType: OrderType,
    onSelect: (OrderType) -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        colors = CardDefaults.cardColors(containerColor = CardBackground),
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Text(
                text = "Order Type",
                style = MaterialTheme.typography.labelMedium,
                color = Color.White.copy(alpha = 0.5f)
            )

            Spacer(modifier = Modifier.height(12.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp)
            ) {
                OrderType.entries.forEach { type ->
                    Surface(
                        onClick = { onSelect(type) },
                        modifier = Modifier.weight(1f),
                        color = if (orderType == type) PrimaryBlue else SurfaceColor,
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Column(
                            modifier = Modifier.padding(16.dp),
                            horizontalAlignment = Alignment.CenterHorizontally
                        ) {
                            Icon(
                                if (type == OrderType.MARKET) Icons.Outlined.Speed else Icons.Outlined.Timer,
                                contentDescription = null,
                                tint = Color.White,
                                modifier = Modifier.size(24.dp)
                            )
                            Spacer(modifier = Modifier.height(8.dp))
                            Text(
                                text = type.name.lowercase().replaceFirstChar { it.uppercase() },
                                style = MaterialTheme.typography.labelMedium,
                                fontWeight = if (orderType == type) FontWeight.SemiBold else FontWeight.Normal,
                                color = Color.White
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun QuantityInput(
    quantity: String,
    onQuantityChange: (String) -> Unit,
    maxQuantity: Double?
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        colors = CardDefaults.cardColors(containerColor = CardBackground),
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Quantity (Shares)",
                    style = MaterialTheme.typography.labelMedium,
                    color = Color.White.copy(alpha = 0.5f)
                )
                if (maxQuantity != null) {
                    TextButton(
                        onClick = {
                            onQuantityChange(maxQuantity.toString())
                        }
                    ) {
                        Text(
                            text = "Max: ${String.format("%.4f", maxQuantity).trimEnd('0').trimEnd('.')}",
                            style = MaterialTheme.typography.labelSmall,
                            color = PrimaryBlue
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(8.dp))

            OutlinedTextField(
                value = quantity,
                onValueChange = { value ->
                    if (value.isEmpty() || value.toDoubleOrNull() != null) {
                        onQuantityChange(value)
                    }
                },
                modifier = Modifier.fillMaxWidth(),
                placeholder = {
                    Text(
                        "0",
                        style = MaterialTheme.typography.headlineMedium,
                        color = Color.White.copy(alpha = 0.3f)
                    )
                },
                textStyle = MaterialTheme.typography.headlineMedium.copy(
                    color = Color.White,
                    fontWeight = FontWeight.SemiBold
                ),
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = PrimaryBlue,
                    unfocusedBorderColor = SurfaceColor,
                    cursorColor = PrimaryBlue
                ),
                shape = RoundedCornerShape(12.dp),
                singleLine = true,
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal)
            )

            Spacer(modifier = Modifier.height(12.dp))

            // Quick amount buttons
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                listOf(1, 5, 10, 25, 100).forEach { amount ->
                    Surface(
                        onClick = { onQuantityChange(amount.toString()) },
                        modifier = Modifier.weight(1f),
                        color = SurfaceColor,
                        shape = RoundedCornerShape(8.dp)
                    ) {
                        Text(
                            text = amount.toString(),
                            style = MaterialTheme.typography.labelMedium,
                            color = Color.White,
                            textAlign = TextAlign.Center,
                            modifier = Modifier.padding(vertical = 10.dp)
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun LimitPriceInput(
    price: String,
    onPriceChange: (String) -> Unit,
    currentPrice: Double
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        colors = CardDefaults.cardColors(containerColor = CardBackground),
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Limit Price",
                    style = MaterialTheme.typography.labelMedium,
                    color = Color.White.copy(alpha = 0.5f)
                )
                TextButton(
                    onClick = {
                        onPriceChange(String.format("%.2f", currentPrice))
                    }
                ) {
                    Text(
                        text = "Current: $${String.format("%.2f", currentPrice)}",
                        style = MaterialTheme.typography.labelSmall,
                        color = PrimaryBlue
                    )
                }
            }

            Spacer(modifier = Modifier.height(8.dp))

            OutlinedTextField(
                value = price,
                onValueChange = { value ->
                    if (value.isEmpty() || value.toDoubleOrNull() != null) {
                        onPriceChange(value)
                    }
                },
                modifier = Modifier.fillMaxWidth(),
                prefix = {
                    Text(
                        "$",
                        style = MaterialTheme.typography.headlineMedium,
                        color = Color.White.copy(alpha = 0.5f)
                    )
                },
                placeholder = {
                    Text(
                        "0.00",
                        style = MaterialTheme.typography.headlineMedium,
                        color = Color.White.copy(alpha = 0.3f)
                    )
                },
                textStyle = MaterialTheme.typography.headlineMedium.copy(
                    color = Color.White,
                    fontWeight = FontWeight.SemiBold
                ),
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = PrimaryBlue,
                    unfocusedBorderColor = SurfaceColor,
                    cursorColor = PrimaryBlue
                ),
                shape = RoundedCornerShape(12.dp),
                singleLine = true,
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Decimal)
            )
        }
    }
}

@Composable
fun OrderSummaryCard(
    equity: Equity?,
    isBuy: Boolean,
    quantity: Double,
    price: Double,
    total: Double,
    fee: Double
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        colors = CardDefaults.cardColors(containerColor = CardBackground),
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            Text(
                text = "Order Summary",
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold,
                color = Color.White
            )

            Spacer(modifier = Modifier.height(16.dp))

            SummaryRow(label = "Shares", value = String.format("%.4f", quantity).trimEnd('0').trimEnd('.'))
            SummaryRow(label = "Price per share", value = "$${String.format("%.2f", price)}")
            SummaryRow(label = "Subtotal", value = "$${String.format("%.2f", total)}")
            SummaryRow(label = "Trading Fee (0.3%)", value = "$${String.format("%.2f", fee)}")

            HorizontalDivider(
                modifier = Modifier.padding(vertical = 12.dp),
                color = SurfaceColor
            )

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "Total",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
                Text(
                    text = "$${String.format("%.2f", total + fee)}",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = if (isBuy) GainGreen else LossRed
                )
            }
        }
    }
}

@Composable
fun SummaryRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 6.dp),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.bodyMedium,
            color = Color.White.copy(alpha = 0.6f)
        )
        Text(
            text = value,
            style = MaterialTheme.typography.bodyMedium,
            color = Color.White
        )
    }
}

@Composable
fun AvailableBalanceCard(
    isBuy: Boolean,
    holding: EquityHolding?
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        colors = CardDefaults.cardColors(containerColor = SurfaceColor),
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Icon(
                if (isBuy) Icons.Outlined.AccountBalanceWallet else Icons.Outlined.Inventory,
                contentDescription = null,
                tint = Color.White.copy(alpha = 0.5f),
                modifier = Modifier.size(20.dp)
            )
            Spacer(modifier = Modifier.width(12.dp))
            Text(
                text = if (isBuy) "Available Balance" else "Available to Sell",
                style = MaterialTheme.typography.bodySmall,
                color = Color.White.copy(alpha = 0.5f),
                modifier = Modifier.weight(1f)
            )
            Text(
                text = if (isBuy) "$10,000.00" else "${holding?.formattedShares ?: "0"} shares",
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.SemiBold,
                color = Color.White
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EquitySelectorBottomSheet(
    equities: List<Equity>,
    selectedSymbol: String?,
    onSelect: (Equity) -> Unit,
    onDismiss: () -> Unit
) {
    ModalBottomSheet(
        onDismissRequest = onDismiss,
        containerColor = CardBackground
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 32.dp)
        ) {
            Text(
                text = "Select Stock",
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold,
                color = Color.White,
                modifier = Modifier.padding(horizontal = 24.dp, vertical = 16.dp)
            )

            equities.forEach { equity ->
                Surface(
                    onClick = { onSelect(equity) },
                    color = if (equity.symbol == selectedSymbol)
                        PrimaryBlue.copy(alpha = 0.15f)
                    else
                        Color.Transparent
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 24.dp, vertical = 14.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Box(
                            modifier = Modifier
                                .size(40.dp)
                                .clip(CircleShape)
                                .background(equity.sector.color.copy(alpha = 0.2f)),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = equity.symbol.take(1),
                                style = MaterialTheme.typography.titleSmall,
                                fontWeight = FontWeight.Bold,
                                color = equity.sector.color
                            )
                        }

                        Spacer(modifier = Modifier.width(12.dp))

                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = equity.symbol,
                                style = MaterialTheme.typography.bodyLarge,
                                fontWeight = FontWeight.SemiBold,
                                color = Color.White
                            )
                            Text(
                                text = equity.companyName,
                                style = MaterialTheme.typography.bodySmall,
                                color = Color.White.copy(alpha = 0.5f)
                            )
                        }

                        Column(horizontalAlignment = Alignment.End) {
                            Text(
                                text = equity.formattedPrice,
                                style = MaterialTheme.typography.bodyMedium,
                                fontWeight = FontWeight.Medium,
                                color = Color.White
                            )
                            Text(
                                text = equity.formattedChange,
                                style = MaterialTheme.typography.labelSmall,
                                color = if (equity.isPositiveChange) GainGreen else LossRed
                            )
                        }

                        if (equity.symbol == selectedSymbol) {
                            Spacer(modifier = Modifier.width(12.dp))
                            Icon(
                                Icons.Default.CheckCircle,
                                contentDescription = "Selected",
                                tint = PrimaryBlue
                            )
                        }
                    }
                }
            }
        }
    }
}
