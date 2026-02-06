package com.sharehodl.ui.portfolio

import androidx.compose.animation.*
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
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
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.hapticfeedback.HapticFeedbackType
import androidx.compose.ui.platform.LocalHapticFeedback
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import com.sharehodl.data.DemoData
import com.sharehodl.model.*
import com.sharehodl.viewmodel.WalletViewModel

// Professional financial colors
val GainGreen = Color(0xFF00C853)
val LossRed = Color(0xFFFF1744)
val NeutralGray = Color(0xFF9E9E9E)
val PrimaryBlue = Color(0xFF1976D2)
val DarkBackground = Color(0xFF0D1117)
val CardBackground = Color(0xFF161B22)
val SurfaceColor = Color(0xFF21262D)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PortfolioScreen(
    onEquityClick: (String) -> Unit = {},
    onTradeClick: () -> Unit = {},
    onCryptoClick: (Chain) -> Unit = {},
    onP2PClick: () -> Unit = {},
    onLendingClick: () -> Unit = {},
    onInheritanceClick: () -> Unit = {},
    walletViewModel: WalletViewModel = hiltViewModel()
) {
    val haptics = LocalHapticFeedback.current

    val holdings = remember { DemoData.holdings }
    val portfolioSummary = remember { DemoData.getPortfolioSummary() }
    val watchlist = remember { DemoData.watchlist }

    // Real wallet data from user's phrase
    val chainAccounts by walletViewModel.chainAccounts.collectAsState()
    val walletAddress by walletViewModel.walletAddress.collectAsState()
    val balance by walletViewModel.balance.collectAsState()
    val uiState by walletViewModel.uiState.collectAsState()
    val totalPortfolioValue by walletViewModel.totalPortfolioValue.collectAsState()

    // Refresh balances on first load
    LaunchedEffect(Unit) {
        if (chainAccounts.isNotEmpty()) {
            walletViewModel.refreshMultiChainBalances()
        }
    }

    // Sort chain accounts to ensure ShareHODL is always first
    val sortedChainAccounts = remember(chainAccounts) {
        chainAccounts.sortedBy { account ->
            if (account.chain == Chain.SHAREHODL) 0 else Chain.majorChains.indexOf(account.chain) + 1
        }
    }

    var selectedTab by remember { mutableStateOf(0) }
    val tabs = listOf("Equity", "Crypto", "Watchlist")

    Scaffold(
        containerColor = DarkBackground,
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        "Portfolio",
                        fontWeight = FontWeight.Bold
                    )
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = DarkBackground,
                    titleContentColor = Color.White
                ),
                actions = {
                    IconButton(onClick = { /* Search */ }) {
                        Icon(
                            Icons.Default.Search,
                            contentDescription = "Search",
                            tint = Color.White
                        )
                    }
                    IconButton(onClick = { /* Notifications */ }) {
                        Icon(
                            Icons.Outlined.Notifications,
                            contentDescription = "Notifications",
                            tint = Color.White
                        )
                    }
                }
            )
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            // Portfolio Value Card
            item {
                PortfolioValueCard(
                    summary = portfolioSummary,
                    onTradeClick = onTradeClick
                )
            }

            // Quick Stats Row
            item {
                QuickStatsRow(summary = portfolioSummary)
            }

            // DeFi Services Section
            item {
                DefiServicesSection(
                    onP2PClick = onP2PClick,
                    onLendingClick = onLendingClick,
                    onInheritanceClick = onInheritanceClick
                )
            }

            // Tab Selector
            item {
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
            }

            // Content based on selected tab
            when (selectedTab) {
                0 -> {
                    // Create ShareHODL equity holding from real wallet balance
                    val shareHodlAccount = sortedChainAccounts.find { it.chain == Chain.SHAREHODL }
                    val shareHodlBalance = shareHodlAccount?.balance?.replace(",", "")?.toDoubleOrNull() ?: 0.0

                    // ShareHODL token as first equity (if user has balance)
                    if (shareHodlBalance > 0) {
                        item {
                            ShareHodlEquityCard(
                                balance = shareHodlBalance,
                                balanceUsd = shareHodlAccount?.balanceUsd ?: "$${String.format("%,.2f", shareHodlBalance)}",
                                onClick = {
                                    haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                                    onCryptoClick(Chain.SHAREHODL)
                                }
                            )
                        }
                    }

                    // Demo equity holdings
                    if (holdings.isEmpty() && shareHodlBalance <= 0) {
                        item { EmptyHoldingsCard() }
                    } else {
                        items(
                            items = holdings,
                            key = { it.equity.symbol }
                        ) { holding ->
                            EquityHoldingCard(
                                holding = holding,
                                onClick = {
                                    haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                                    onEquityClick(holding.equity.symbol)
                                }
                            )
                        }
                    }
                }
                1 -> {
                    // Real Crypto Holdings from wallet
                    if (sortedChainAccounts.isEmpty()) {
                        // No wallet created yet
                        item { EmptyCryptoCard() }
                    } else {
                        // Show wallet address header with total value
                        item {
                            CryptoWalletHeader(
                                address = sortedChainAccounts.firstOrNull()?.address ?: walletAddress ?: "",
                                totalBalance = walletViewModel.getFormattedPortfolioValue()
                            )
                        }

                        items(
                            items = sortedChainAccounts,
                            key = { it.chain.name }
                        ) { account ->
                            CryptoHoldingCard(
                                account = account,
                                onClick = {
                                    haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                                    onCryptoClick(account.chain)
                                }
                            )
                        }

                        // Pull to refresh hint
                        item {
                            if (uiState.isRefreshing) {
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(16.dp),
                                    contentAlignment = Alignment.Center
                                ) {
                                    CircularProgressIndicator(
                                        color = PrimaryBlue,
                                        modifier = Modifier.size(24.dp)
                                    )
                                }
                            }
                        }
                    }
                }
                2 -> {
                    // Watchlist
                    if (watchlist.isEmpty()) {
                        item { EmptyWatchlistCard() }
                    } else {
                        items(
                            items = watchlist,
                            key = { it.equity.symbol }
                        ) { item ->
                            WatchlistCard(
                                item = item,
                                onClick = {
                                    haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                                    onEquityClick(item.equity.symbol)
                                }
                            )
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

@Composable
fun PortfolioValueCard(
    summary: PortfolioSummary,
    onTradeClick: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        colors = CardDefaults.cardColors(containerColor = CardBackground),
        shape = RoundedCornerShape(20.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(
                    Brush.verticalGradient(
                        colors = listOf(
                            PrimaryBlue.copy(alpha = 0.15f),
                            CardBackground
                        )
                    )
                )
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(
                text = "Total Portfolio Value",
                style = MaterialTheme.typography.bodyMedium,
                color = Color.White.copy(alpha = 0.7f)
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = summary.formattedTotalValue,
                style = MaterialTheme.typography.displaySmall.copy(
                    fontWeight = FontWeight.Bold,
                    fontSize = 40.sp
                ),
                color = Color.White
            )

            Spacer(modifier = Modifier.height(12.dp))

            // Day change badge
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.Center
            ) {
                val changeColor = if (summary.isPositiveDay) GainGreen else LossRed

                Surface(
                    color = changeColor.copy(alpha = 0.15f),
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            if (summary.isPositiveDay) Icons.Default.TrendingUp else Icons.Default.TrendingDown,
                            contentDescription = null,
                            modifier = Modifier.size(18.dp),
                            tint = changeColor
                        )
                        Spacer(modifier = Modifier.width(6.dp))
                        Text(
                            text = "${summary.formattedDayChange} (${summary.formattedDayChangePercent})",
                            style = MaterialTheme.typography.bodyMedium,
                            fontWeight = FontWeight.SemiBold,
                            color = changeColor
                        )
                        Text(
                            text = " today",
                            style = MaterialTheme.typography.bodySmall,
                            color = changeColor.copy(alpha = 0.7f)
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            // Action buttons
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                PortfolioActionButton(
                    icon = Icons.Default.Add,
                    label = "Buy",
                    color = GainGreen,
                    onClick = onTradeClick
                )
                PortfolioActionButton(
                    icon = Icons.Default.Remove,
                    label = "Sell",
                    color = LossRed,
                    onClick = onTradeClick
                )
                PortfolioActionButton(
                    icon = Icons.Default.SwapHoriz,
                    label = "Transfer",
                    color = PrimaryBlue,
                    onClick = { }
                )
                PortfolioActionButton(
                    icon = Icons.Default.BarChart,
                    label = "Analytics",
                    color = Color(0xFFF59E0B),
                    onClick = { }
                )
            }
        }
    }
}

@Composable
fun PortfolioActionButton(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    color: Color,
    onClick: () -> Unit
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.clickable(onClick = onClick)
    ) {
        Surface(
            modifier = Modifier.size(52.dp),
            shape = CircleShape,
            color = color.copy(alpha = 0.15f)
        ) {
            Box(contentAlignment = Alignment.Center) {
                Icon(
                    icon,
                    contentDescription = label,
                    modifier = Modifier.size(24.dp),
                    tint = color
                )
            }
        }
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = Color.White.copy(alpha = 0.8f)
        )
    }
}

@Composable
fun QuickStatsRow(summary: PortfolioSummary) {
    LazyRow(
        modifier = Modifier.fillMaxWidth(),
        contentPadding = PaddingValues(horizontal = 16.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        item {
            QuickStatChip(
                label = "Total P&L",
                value = "$${String.format("%,.2f", summary.totalGainLoss)}",
                subValue = String.format("%.2f%%", summary.totalGainLossPercent),
                isPositive = summary.isPositiveTotal
            )
        }
        item {
            QuickStatChip(
                label = "Dividends",
                value = "$${String.format("%.2f", summary.dividendsEarned)}",
                subValue = "Earned",
                isPositive = true
            )
        }
        item {
            QuickStatChip(
                label = "Holdings",
                value = DemoData.holdings.size.toString(),
                subValue = "Stocks",
                isPositive = null
            )
        }
    }
}

@Composable
fun QuickStatChip(
    label: String,
    value: String,
    subValue: String,
    isPositive: Boolean?
) {
    val valueColor = when (isPositive) {
        true -> GainGreen
        false -> LossRed
        null -> Color.White
    }

    Surface(
        color = SurfaceColor,
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            horizontalAlignment = Alignment.Start
        ) {
            Text(
                text = label,
                style = MaterialTheme.typography.labelSmall,
                color = Color.White.copy(alpha = 0.5f)
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = value,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
                color = valueColor
            )
            Text(
                text = subValue,
                style = MaterialTheme.typography.labelSmall,
                color = valueColor.copy(alpha = 0.7f)
            )
        }
    }
}

@Composable
fun EquityHoldingCard(
    holding: EquityHolding,
    onClick: () -> Unit
) {
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
            // Company Logo/Symbol
            Box(
                modifier = Modifier
                    .size(52.dp)
                    .clip(CircleShape)
                    .background(holding.equity.sector.color.copy(alpha = 0.2f)),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = holding.equity.symbol.take(2),
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = holding.equity.sector.color
                )
            }

            Spacer(modifier = Modifier.width(14.dp))

            // Company Info
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = holding.equity.symbol,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
                Text(
                    text = holding.equity.companyName,
                    style = MaterialTheme.typography.bodySmall,
                    color = Color.White.copy(alpha = 0.6f),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "${holding.formattedShares} shares",
                    style = MaterialTheme.typography.labelSmall,
                    color = Color.White.copy(alpha = 0.5f)
                )
            }

            // Value & Change
            Column(horizontalAlignment = Alignment.End) {
                Text(
                    text = holding.formattedValue,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    color = Color.White
                )
                Row(verticalAlignment = Alignment.CenterVertically) {
                    val changeColor = if (holding.isProfit) GainGreen else LossRed
                    Icon(
                        if (holding.isProfit) Icons.Default.TrendingUp else Icons.Default.TrendingDown,
                        contentDescription = null,
                        modifier = Modifier.size(14.dp),
                        tint = changeColor
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = holding.formattedGainLossPercent,
                        style = MaterialTheme.typography.bodySmall,
                        fontWeight = FontWeight.Medium,
                        color = changeColor
                    )
                }
                Text(
                    text = holding.formattedGainLoss,
                    style = MaterialTheme.typography.labelSmall,
                    color = if (holding.isProfit) GainGreen.copy(alpha = 0.7f) else LossRed.copy(alpha = 0.7f)
                )
            }
        }
    }
}

@Composable
fun WatchlistCard(
    item: WatchlistItem,
    onClick: () -> Unit
) {
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
            // Company Logo/Symbol
            Box(
                modifier = Modifier
                    .size(52.dp)
                    .clip(CircleShape)
                    .background(item.equity.sector.color.copy(alpha = 0.2f)),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = item.equity.symbol.take(2),
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = item.equity.sector.color
                )
            }

            Spacer(modifier = Modifier.width(14.dp))

            // Company Info
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = item.equity.symbol,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
                Text(
                    text = item.equity.companyName,
                    style = MaterialTheme.typography.bodySmall,
                    color = Color.White.copy(alpha = 0.6f),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                if (item.alertPrice != null) {
                    Spacer(modifier = Modifier.height(4.dp))
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Icon(
                            Icons.Outlined.NotificationsActive,
                            contentDescription = null,
                            modifier = Modifier.size(12.dp),
                            tint = Color(0xFFF59E0B)
                        )
                        Spacer(modifier = Modifier.width(4.dp))
                        Text(
                            text = "Alert at $${String.format("%.2f", item.alertPrice)}",
                            style = MaterialTheme.typography.labelSmall,
                            color = Color(0xFFF59E0B)
                        )
                    }
                }
            }

            // Price & Change
            Column(horizontalAlignment = Alignment.End) {
                Text(
                    text = item.equity.formattedPrice,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    color = Color.White
                )
                val changeColor = if (item.equity.isPositiveChange) GainGreen else LossRed
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        if (item.equity.isPositiveChange) Icons.Default.TrendingUp else Icons.Default.TrendingDown,
                        contentDescription = null,
                        modifier = Modifier.size(14.dp),
                        tint = changeColor
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = item.equity.formattedChange,
                        style = MaterialTheme.typography.bodySmall,
                        fontWeight = FontWeight.Medium,
                        color = changeColor
                    )
                }
            }
        }
    }
}

@Composable
fun EmptyHoldingsCard() {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(32.dp),
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
                Icons.Outlined.AccountBalance,
                contentDescription = null,
                modifier = Modifier.size(40.dp),
                tint = Color.White.copy(alpha = 0.5f)
            )
        }
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = "No Holdings Yet",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
            color = Color.White
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = "Start building your portfolio by\npurchasing your first stock",
            style = MaterialTheme.typography.bodyMedium,
            color = Color.White.copy(alpha = 0.6f),
            textAlign = TextAlign.Center
        )
        Spacer(modifier = Modifier.height(24.dp))
        Button(
            onClick = { },
            colors = ButtonDefaults.buttonColors(containerColor = PrimaryBlue),
            shape = RoundedCornerShape(12.dp)
        ) {
            Icon(Icons.Default.Add, contentDescription = null)
            Spacer(modifier = Modifier.width(8.dp))
            Text("Browse Stocks")
        }
    }
}

@Composable
fun EmptyWatchlistCard() {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(32.dp),
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
                Icons.Outlined.Visibility,
                contentDescription = null,
                modifier = Modifier.size(40.dp),
                tint = Color.White.copy(alpha = 0.5f)
            )
        }
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = "Watchlist Empty",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
            color = Color.White
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = "Add stocks to your watchlist\nto track their performance",
            style = MaterialTheme.typography.bodyMedium,
            color = Color.White.copy(alpha = 0.6f),
            textAlign = TextAlign.Center
        )
    }
}

// ============================================
// CRYPTO SECTION COMPONENTS
// ============================================

@Composable
fun CryptoWalletHeader(
    address: String,
    totalBalance: String
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
                .background(
                    Brush.horizontalGradient(
                        colors = listOf(
                            Color(0xFF6366F1).copy(alpha = 0.2f),
                            Color(0xFFF7931A).copy(alpha = 0.1f)
                        )
                    )
                )
                .padding(16.dp)
        ) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column {
                    Text(
                        text = "Your Wallet",
                        style = MaterialTheme.typography.labelMedium,
                        color = Color.White.copy(alpha = 0.6f)
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = if (address.length > 20) "${address.take(10)}...${address.takeLast(6)}" else address,
                        style = MaterialTheme.typography.bodySmall,
                        color = Color.White.copy(alpha = 0.8f),
                        fontWeight = FontWeight.Medium
                    )
                }

                // Copy address button
                Surface(
                    color = SurfaceColor,
                    shape = RoundedCornerShape(8.dp)
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            Icons.Outlined.ContentCopy,
                            contentDescription = "Copy",
                            modifier = Modifier.size(16.dp),
                            tint = Color.White.copy(alpha = 0.7f)
                        )
                        Spacer(modifier = Modifier.width(6.dp))
                        Text(
                            text = "Copy",
                            style = MaterialTheme.typography.labelSmall,
                            color = Color.White.copy(alpha = 0.7f)
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Total Crypto Balance",
                    style = MaterialTheme.typography.bodySmall,
                    color = Color.White.copy(alpha = 0.6f)
                )
                Text(
                    text = totalBalance,
                    style = MaterialTheme.typography.titleMedium,
                    color = Color.White,
                    fontWeight = FontWeight.Bold
                )
            }
        }
    }
}

@Composable
fun CryptoHoldingCard(
    account: ChainAccount,
    onClick: () -> Unit
) {
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
            // Chain Logo
            Box(
                modifier = Modifier
                    .size(52.dp)
                    .clip(CircleShape)
                    .background(account.chain.color.copy(alpha = 0.2f)),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = account.chain.symbol.take(2),
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = account.chain.color
                )
            }

            Spacer(modifier = Modifier.width(14.dp))

            // Chain Info
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = account.chain.displayName,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
                Text(
                    text = account.shortAddress,
                    style = MaterialTheme.typography.bodySmall,
                    color = Color.White.copy(alpha = 0.6f),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                Spacer(modifier = Modifier.height(4.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        Icons.Outlined.Route,
                        contentDescription = null,
                        modifier = Modifier.size(12.dp),
                        tint = Color.White.copy(alpha = 0.4f)
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = account.derivationPath,
                        style = MaterialTheme.typography.labelSmall,
                        color = Color.White.copy(alpha = 0.4f)
                    )
                }
            }

            // Balance
            Column(horizontalAlignment = Alignment.End) {
                Text(
                    text = account.formattedBalance,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    color = Color.White
                )
                if (account.balanceUsd != null) {
                    Text(
                        text = "$$${account.balanceUsd}",
                        style = MaterialTheme.typography.bodySmall,
                        color = Color.White.copy(alpha = 0.6f)
                    )
                }
                Spacer(modifier = Modifier.height(4.dp))
                // Chain tag
                Surface(
                    color = account.chain.color.copy(alpha = 0.15f),
                    shape = RoundedCornerShape(6.dp)
                ) {
                    Text(
                        text = account.chain.symbol,
                        style = MaterialTheme.typography.labelSmall,
                        color = account.chain.color,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp)
                    )
                }
            }
        }
    }
}

@Composable
fun EmptyCryptoCard() {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(32.dp),
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
                Icons.Outlined.CurrencyBitcoin,
                contentDescription = null,
                modifier = Modifier.size(40.dp),
                tint = Color(0xFFF7931A)
            )
        }
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = "No Crypto Balance",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
            color = Color.White
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = "Your wallet is set up!\nReceive crypto to start building\nyour portfolio",
            style = MaterialTheme.typography.bodyMedium,
            color = Color.White.copy(alpha = 0.6f),
            textAlign = TextAlign.Center
        )
        Spacer(modifier = Modifier.height(24.dp))
        Row(
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            OutlinedButton(
                onClick = { },
                colors = ButtonDefaults.outlinedButtonColors(contentColor = Color.White),
                shape = RoundedCornerShape(12.dp)
            ) {
                Icon(Icons.Outlined.QrCodeScanner, contentDescription = null, modifier = Modifier.size(18.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text("Receive")
            }
            Button(
                onClick = { },
                colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF6366F1)),
                shape = RoundedCornerShape(12.dp)
            ) {
                Icon(Icons.Default.ShoppingCart, contentDescription = null, modifier = Modifier.size(18.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text("Buy Crypto")
            }
        }
    }
}

@Composable
fun NoCryptoWalletCard(
    onCreateWallet: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Box(
            modifier = Modifier
                .size(80.dp)
                .clip(CircleShape)
                .background(
                    Brush.linearGradient(
                        colors = listOf(
                            Color(0xFF6366F1).copy(alpha = 0.3f),
                            Color(0xFFF7931A).copy(alpha = 0.3f)
                        )
                    )
                ),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                Icons.Outlined.Wallet,
                contentDescription = null,
                modifier = Modifier.size(40.dp),
                tint = Color.White
            )
        }
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = "Multi-Chain Wallet",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
            color = Color.White
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = "Create or import your wallet to\nhold crypto across multiple chains",
            style = MaterialTheme.typography.bodyMedium,
            color = Color.White.copy(alpha = 0.6f),
            textAlign = TextAlign.Center
        )
        Spacer(modifier = Modifier.height(16.dp))

        // Supported chains row
        Row(
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Chain.entries.forEach { chain ->
                Surface(
                    color = chain.color.copy(alpha = 0.2f),
                    shape = CircleShape
                ) {
                    Box(
                        modifier = Modifier
                            .size(36.dp)
                            .padding(8.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = chain.symbol.take(1),
                            style = MaterialTheme.typography.labelSmall,
                            fontWeight = FontWeight.Bold,
                            color = chain.color
                        )
                    }
                }
            }
        }

        Spacer(modifier = Modifier.height(24.dp))

        Row(
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            OutlinedButton(
                onClick = onCreateWallet,
                colors = ButtonDefaults.outlinedButtonColors(contentColor = Color.White),
                shape = RoundedCornerShape(12.dp)
            ) {
                Icon(Icons.Outlined.Download, contentDescription = null, modifier = Modifier.size(18.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text("Import Wallet")
            }
            Button(
                onClick = onCreateWallet,
                colors = ButtonDefaults.buttonColors(containerColor = Color(0xFF6366F1)),
                shape = RoundedCornerShape(12.dp)
            ) {
                Icon(Icons.Default.Add, contentDescription = null, modifier = Modifier.size(18.dp))
                Spacer(modifier = Modifier.width(8.dp))
                Text("Create Wallet")
            }
        }
    }
}

/**
 * Special card for ShareHODL token in equity section
 * Shows the user's real HODL balance prominently
 */
@Composable
fun ShareHodlEquityCard(
    balance: Double,
    balanceUsd: String,
    onClick: () -> Unit
) {
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
                .background(
                    Brush.horizontalGradient(
                        colors = listOf(
                            Color(0xFF6366F1).copy(alpha = 0.15f),
                            CardBackground
                        )
                    )
                )
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // ShareHODL Logo
            Box(
                modifier = Modifier
                    .size(52.dp)
                    .clip(CircleShape)
                    .background(
                        Brush.linearGradient(
                            colors = listOf(
                                Color(0xFF6366F1),
                                Color(0xFF8B5CF6)
                            )
                        )
                    ),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "SH",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
            }

            Spacer(modifier = Modifier.width(14.dp))

            // ShareHODL Info
            Column(modifier = Modifier.weight(1f)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        text = "HODL",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold,
                        color = Color.White
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Surface(
                        color = Color(0xFF6366F1).copy(alpha = 0.2f),
                        shape = RoundedCornerShape(4.dp)
                    ) {
                        Text(
                            text = "Native",
                            style = MaterialTheme.typography.labelSmall,
                            color = Color(0xFF6366F1),
                            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                        )
                    }
                }
                Text(
                    text = "ShareHODL Token",
                    style = MaterialTheme.typography.bodySmall,
                    color = Color.White.copy(alpha = 0.6f),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "${String.format("%,.4f", balance)} HODL",
                    style = MaterialTheme.typography.labelSmall,
                    color = Color.White.copy(alpha = 0.5f)
                )
            }

            // Value
            Column(horizontalAlignment = Alignment.End) {
                Text(
                    text = balanceUsd,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    color = Color.White
                )
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Icon(
                        Icons.Default.TrendingUp,
                        contentDescription = null,
                        modifier = Modifier.size(14.dp),
                        tint = GainGreen
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "$1.00",
                        style = MaterialTheme.typography.bodySmall,
                        fontWeight = FontWeight.Medium,
                        color = Color.White.copy(alpha = 0.6f)
                    )
                }
                Text(
                    text = "Stablecoin",
                    style = MaterialTheme.typography.labelSmall,
                    color = Color(0xFF6366F1)
                )
            }
        }
    }
}

// ============================================
// DEFI SERVICES SECTION
// ============================================

@Composable
fun DefiServicesSection(
    onP2PClick: () -> Unit,
    onLendingClick: () -> Unit,
    onInheritanceClick: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 12.dp)
    ) {
        Text(
            text = "DeFi Services",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold,
            color = Color.White,
            modifier = Modifier.padding(bottom = 12.dp)
        )

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            DefiServiceCard(
                modifier = Modifier.weight(1f),
                icon = Icons.Default.SwapHoriz,
                title = "P2P Trading",
                subtitle = "Trade directly",
                gradientColors = listOf(
                    Color(0xFF6366F1),
                    Color(0xFF8B5CF6)
                ),
                onClick = onP2PClick
            )

            DefiServiceCard(
                modifier = Modifier.weight(1f),
                icon = Icons.Default.AccountBalance,
                title = "Lending",
                subtitle = "Earn interest",
                gradientColors = listOf(
                    Color(0xFF10B981),
                    Color(0xFF059669)
                ),
                onClick = onLendingClick
            )

            DefiServiceCard(
                modifier = Modifier.weight(1f),
                icon = Icons.Default.FamilyRestroom,
                title = "Inheritance",
                subtitle = "Asset transfer",
                gradientColors = listOf(
                    Color(0xFFF59E0B),
                    Color(0xFFD97706)
                ),
                onClick = onInheritanceClick
            )
        }
    }
}

@Composable
fun DefiServiceCard(
    modifier: Modifier = Modifier,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    title: String,
    subtitle: String,
    gradientColors: List<Color>,
    onClick: () -> Unit
) {
    val haptics = LocalHapticFeedback.current

    Surface(
        onClick = {
            haptics.performHapticFeedback(HapticFeedbackType.LongPress)
            onClick()
        },
        modifier = modifier,
        color = CardBackground,
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .background(
                    Brush.verticalGradient(
                        colors = listOf(
                            gradientColors[0].copy(alpha = 0.15f),
                            CardBackground
                        )
                    )
                )
                .padding(12.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Box(
                modifier = Modifier
                    .size(44.dp)
                    .clip(CircleShape)
                    .background(
                        Brush.linearGradient(colors = gradientColors)
                    ),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    icon,
                    contentDescription = title,
                    modifier = Modifier.size(22.dp),
                    tint = Color.White
                )
            }

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = title,
                style = MaterialTheme.typography.labelMedium,
                fontWeight = FontWeight.SemiBold,
                color = Color.White,
                textAlign = TextAlign.Center
            )

            Text(
                text = subtitle,
                style = MaterialTheme.typography.labelSmall,
                color = Color.White.copy(alpha = 0.6f),
                textAlign = TextAlign.Center
            )
        }
    }
}
