package com.sharehodl.ui.market

import androidx.compose.animation.*
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
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
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.sharehodl.data.DemoData
import com.sharehodl.model.*
import com.sharehodl.ui.portfolio.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MarketScreen(
    onEquityClick: (String) -> Unit = {}
) {
    val haptics = LocalHapticFeedback.current

    val equities = remember { DemoData.equities }
    val marketStats = remember { DemoData.getMarketStats() }
    val topGainers = remember { DemoData.getTopGainers() }
    val topLosers = remember { DemoData.getTopLosers() }
    val mostActive = remember { DemoData.getMostActive() }

    var searchQuery by remember { mutableStateOf("") }
    var selectedSector by remember { mutableStateOf<Sector?>(null) }
    var selectedTab by remember { mutableStateOf(0) }
    val tabs = listOf("All", "Gainers", "Losers", "Active")

    val filteredEquities = remember(searchQuery, selectedSector, selectedTab) {
        val baseList = when (selectedTab) {
            1 -> topGainers
            2 -> topLosers
            3 -> mostActive
            else -> if (selectedSector != null) {
                DemoData.getEquitiesBySector(selectedSector!!)
            } else {
                equities
            }
        }

        if (searchQuery.isNotEmpty()) {
            baseList.filter {
                it.symbol.contains(searchQuery, ignoreCase = true) ||
                it.companyName.contains(searchQuery, ignoreCase = true)
            }
        } else {
            baseList
        }
    }

    Scaffold(
        containerColor = DarkBackground,
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        "Markets",
                        fontWeight = FontWeight.Bold
                    )
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = DarkBackground,
                    titleContentColor = Color.White
                ),
                actions = {
                    IconButton(onClick = { /* Filters */ }) {
                        Icon(
                            Icons.Outlined.FilterList,
                            contentDescription = "Filter",
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
            // Search Bar
            item {
                SearchBar(
                    query = searchQuery,
                    onQueryChange = { searchQuery = it }
                )
            }

            // Market Overview Card
            item {
                MarketOverviewCard(stats = marketStats)
            }

            // Sector Filter Chips
            item {
                SectorFilterRow(
                    selectedSector = selectedSector,
                    onSectorSelected = { sector ->
                        selectedSector = if (selectedSector == sector) null else sector
                        haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                    }
                )
            }

            // Tab Selector
            item {
                ScrollableTabRow(
                    selectedTabIndex = selectedTab,
                    modifier = Modifier.padding(vertical = 8.dp),
                    containerColor = Color.Transparent,
                    contentColor = Color.White,
                    edgePadding = 16.dp,
                    divider = {}
                ) {
                    tabs.forEachIndexed { index, title ->
                        Tab(
                            selected = selectedTab == index,
                            onClick = {
                                selectedTab = index
                                selectedSector = null
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

            // Trending Section (only on All tab)
            if (selectedTab == 0 && searchQuery.isEmpty() && selectedSector == null) {
                item {
                    TrendingSection(
                        topGainer = marketStats.topGainer,
                        topLoser = marketStats.topLoser,
                        mostActive = marketStats.mostActive,
                        onEquityClick = onEquityClick
                    )
                }
            }

            // Equities List Header
            item {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp, vertical = 12.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = when {
                            selectedSector != null -> "${selectedSector!!.displayName} Stocks"
                            selectedTab == 1 -> "Top Gainers"
                            selectedTab == 2 -> "Top Losers"
                            selectedTab == 3 -> "Most Active"
                            else -> "All Stocks"
                        },
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        color = Color.White
                    )
                    Text(
                        text = "${filteredEquities.size} stocks",
                        style = MaterialTheme.typography.bodySmall,
                        color = Color.White.copy(alpha = 0.5f)
                    )
                }
            }

            // Equities List
            items(
                items = filteredEquities,
                key = { it.symbol }
            ) { equity ->
                EquityListItem(
                    equity = equity,
                    onClick = {
                        haptics.performHapticFeedback(HapticFeedbackType.LongPress)
                        onEquityClick(equity.symbol)
                    }
                )
            }

            // Empty state
            if (filteredEquities.isEmpty()) {
                item {
                    EmptySearchCard(query = searchQuery)
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
fun SearchBar(
    query: String,
    onQueryChange: (String) -> Unit
) {
    OutlinedTextField(
        value = query,
        onValueChange = onQueryChange,
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        placeholder = {
            Text(
                "Search stocks by name or symbol",
                color = Color.White.copy(alpha = 0.4f)
            )
        },
        leadingIcon = {
            Icon(
                Icons.Default.Search,
                contentDescription = null,
                tint = Color.White.copy(alpha = 0.5f)
            )
        },
        trailingIcon = {
            if (query.isNotEmpty()) {
                IconButton(onClick = { onQueryChange("") }) {
                    Icon(
                        Icons.Default.Clear,
                        contentDescription = "Clear",
                        tint = Color.White.copy(alpha = 0.5f)
                    )
                }
            }
        },
        colors = OutlinedTextFieldDefaults.colors(
            focusedBorderColor = PrimaryBlue,
            unfocusedBorderColor = SurfaceColor,
            focusedContainerColor = CardBackground,
            unfocusedContainerColor = CardBackground,
            cursorColor = PrimaryBlue,
            focusedTextColor = Color.White,
            unfocusedTextColor = Color.White
        ),
        shape = RoundedCornerShape(16.dp),
        singleLine = true,
        keyboardOptions = KeyboardOptions(imeAction = ImeAction.Search)
    )
}

@Composable
fun MarketOverviewCard(stats: MarketStats) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        colors = CardDefaults.cardColors(containerColor = CardBackground),
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(20.dp)
        ) {
            Text(
                text = "Market Overview",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
                color = Color.White
            )

            Spacer(modifier = Modifier.height(16.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                MarketStatItem(
                    label = "Total Market Cap",
                    value = formatLargeNumber(stats.totalMarketCap),
                    modifier = Modifier.weight(1f)
                )
                MarketStatItem(
                    label = "24h Volume",
                    value = formatLargeNumber(stats.totalVolume24h),
                    modifier = Modifier.weight(1f)
                )
                MarketStatItem(
                    label = "Active Stocks",
                    value = stats.activeCompanies.toString(),
                    modifier = Modifier.weight(1f)
                )
            }
        }
    }
}

@Composable
fun MarketStatItem(
    label: String,
    value: String,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier,
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = value,
            style = MaterialTheme.typography.titleLarge,
            fontWeight = FontWeight.Bold,
            color = Color.White
        )
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = label,
            style = MaterialTheme.typography.labelSmall,
            color = Color.White.copy(alpha = 0.5f)
        )
    }
}

@Composable
fun SectorFilterRow(
    selectedSector: Sector?,
    onSectorSelected: (Sector) -> Unit
) {
    LazyRow(
        modifier = Modifier.fillMaxWidth(),
        contentPadding = PaddingValues(horizontal = 16.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        items(Sector.entries.toList()) { sector ->
            SectorChip(
                sector = sector,
                isSelected = sector == selectedSector,
                onClick = { onSectorSelected(sector) }
            )
        }
    }
}

@Composable
fun SectorChip(
    sector: Sector,
    isSelected: Boolean,
    onClick: () -> Unit
) {
    Surface(
        onClick = onClick,
        color = if (isSelected) sector.color else SurfaceColor,
        shape = RoundedCornerShape(20.dp)
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 8.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(8.dp)
                    .clip(CircleShape)
                    .background(if (isSelected) Color.White else sector.color)
            )
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                text = sector.displayName,
                style = MaterialTheme.typography.labelMedium,
                fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
                color = if (isSelected) Color.White else Color.White.copy(alpha = 0.8f)
            )
        }
    }
}

@Composable
fun TrendingSection(
    topGainer: Equity?,
    topLoser: Equity?,
    mostActive: Equity?,
    onEquityClick: (String) -> Unit
) {
    Column(
        modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
    ) {
        Text(
            text = "Trending Today",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
            color = Color.White
        )

        Spacer(modifier = Modifier.height(12.dp))

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            topGainer?.let {
                TrendingCard(
                    equity = it,
                    label = "Top Gainer",
                    icon = Icons.Default.TrendingUp,
                    accentColor = GainGreen,
                    modifier = Modifier.weight(1f),
                    onClick = { onEquityClick(it.symbol) }
                )
            }
            topLoser?.let {
                TrendingCard(
                    equity = it,
                    label = "Top Loser",
                    icon = Icons.Default.TrendingDown,
                    accentColor = LossRed,
                    modifier = Modifier.weight(1f),
                    onClick = { onEquityClick(it.symbol) }
                )
            }
        }
    }
}

@Composable
fun TrendingCard(
    equity: Equity,
    label: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    accentColor: Color,
    modifier: Modifier = Modifier,
    onClick: () -> Unit
) {
    Surface(
        onClick = onClick,
        modifier = modifier,
        color = CardBackground,
        shape = RoundedCornerShape(16.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    icon,
                    contentDescription = null,
                    modifier = Modifier.size(16.dp),
                    tint = accentColor
                )
                Spacer(modifier = Modifier.width(4.dp))
                Text(
                    text = label,
                    style = MaterialTheme.typography.labelSmall,
                    color = accentColor
                )
            }

            Spacer(modifier = Modifier.height(12.dp))

            Row(verticalAlignment = Alignment.CenterVertically) {
                Box(
                    modifier = Modifier
                        .size(36.dp)
                        .clip(CircleShape)
                        .background(equity.sector.color.copy(alpha = 0.2f)),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = equity.symbol.take(1),
                        style = MaterialTheme.typography.labelLarge,
                        fontWeight = FontWeight.Bold,
                        color = equity.sector.color
                    )
                }
                Spacer(modifier = Modifier.width(10.dp))
                Column {
                    Text(
                        text = equity.symbol,
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.Bold,
                        color = Color.White
                    )
                    Text(
                        text = equity.formattedChange,
                        style = MaterialTheme.typography.bodySmall,
                        fontWeight = FontWeight.SemiBold,
                        color = accentColor
                    )
                }
            }
        }
    }
}

@Composable
fun EquityListItem(
    equity: Equity,
    onClick: () -> Unit
) {
    Surface(
        onClick = onClick,
        modifier = Modifier.fillMaxWidth(),
        color = Color.Transparent
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Company Logo/Symbol
            Box(
                modifier = Modifier
                    .size(48.dp)
                    .clip(CircleShape)
                    .background(equity.sector.color.copy(alpha = 0.15f)),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = equity.symbol.take(2),
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.Bold,
                    color = equity.sector.color
                )
            }

            Spacer(modifier = Modifier.width(14.dp))

            // Company Info
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = equity.symbol,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
                Text(
                    text = equity.companyName,
                    style = MaterialTheme.typography.bodySmall,
                    color = Color.White.copy(alpha = 0.5f),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
            }

            // Price & Change
            Column(horizontalAlignment = Alignment.End) {
                Text(
                    text = equity.formattedPrice,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                    color = Color.White
                )
                val changeColor = if (equity.isPositiveChange) GainGreen else LossRed
                Surface(
                    color = changeColor.copy(alpha = 0.1f),
                    shape = RoundedCornerShape(4.dp)
                ) {
                    Text(
                        text = equity.formattedChange,
                        style = MaterialTheme.typography.labelSmall,
                        fontWeight = FontWeight.SemiBold,
                        color = changeColor,
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp)
                    )
                }
            }
        }
    }

    HorizontalDivider(
        modifier = Modifier.padding(start = 78.dp, end = 16.dp),
        thickness = 0.5.dp,
        color = SurfaceColor
    )
}

@Composable
fun EmptySearchCard(query: String) {
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
                Icons.Outlined.SearchOff,
                contentDescription = null,
                modifier = Modifier.size(40.dp),
                tint = Color.White.copy(alpha = 0.5f)
            )
        }
        Spacer(modifier = Modifier.height(16.dp))
        Text(
            text = if (query.isNotEmpty()) "No Results Found" else "No Stocks Available",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
            color = Color.White
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = if (query.isNotEmpty())
                "We couldn't find any stocks matching \"$query\""
            else
                "Check back later for available stocks",
            style = MaterialTheme.typography.bodyMedium,
            color = Color.White.copy(alpha = 0.6f),
            textAlign = TextAlign.Center
        )
    }
}

private fun formatLargeNumber(value: Long): String {
    return when {
        value >= 1_000_000_000_000 -> "$${String.format("%.1f", value / 1_000_000_000_000.0)}T"
        value >= 1_000_000_000 -> "$${String.format("%.1f", value / 1_000_000_000.0)}B"
        value >= 1_000_000 -> "$${String.format("%.1f", value / 1_000_000.0)}M"
        else -> "$${String.format("%,d", value)}"
    }
}
