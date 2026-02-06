package com.sharehodl

import android.os.Bundle
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.fragment.app.FragmentActivity
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material.icons.outlined.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavDestination.Companion.hierarchy
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.sharehodl.ui.activity.ActivityScreen
import com.sharehodl.ui.equity.EquityDetailScreen
import com.sharehodl.ui.market.MarketScreen
import com.sharehodl.ui.onboarding.OnboardingScreen
import com.sharehodl.ui.portfolio.DarkBackground
import com.sharehodl.ui.portfolio.CardBackground
import com.sharehodl.ui.portfolio.PortfolioScreen
import com.sharehodl.ui.portfolio.PrimaryBlue
import com.sharehodl.ui.settings.SettingsScreen
import com.sharehodl.ui.theme.ShareHODLTheme
import com.sharehodl.ui.trade.EquityTradeScreen
import com.sharehodl.ui.crypto.SendCryptoScreen
import com.sharehodl.ui.crypto.ReceiveCryptoScreen
import com.sharehodl.ui.crypto.TransactionHistorySection
import com.sharehodl.ui.p2p.P2PTradingScreen
import com.sharehodl.ui.lending.LendingScreen
import com.sharehodl.ui.inheritance.InheritanceScreen
import com.sharehodl.data.DemoCryptoData
import com.sharehodl.model.Chain
import com.sharehodl.model.ChainAccount
import com.sharehodl.viewmodel.WalletViewModel
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class MainActivity : FragmentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            ShareHODLTheme(darkTheme = true) {
                ShareHODLApp()
            }
        }
    }
}

/**
 * Navigation destinations - Equity focused
 */
sealed class Screen(
    val route: String,
    val title: String,
    val selectedIcon: ImageVector,
    val unselectedIcon: ImageVector
) {
    object Portfolio : Screen(
        "portfolio",
        "Portfolio",
        Icons.Filled.PieChart,
        Icons.Outlined.PieChart
    )
    object Market : Screen(
        "market",
        "Market",
        Icons.Filled.ShowChart,
        Icons.Outlined.ShowChart
    )
    object Trade : Screen(
        "trade",
        "Trade",
        Icons.Filled.SwapVert,
        Icons.Outlined.SwapVert
    )
    object Activity : Screen(
        "activity",
        "Activity",
        Icons.Filled.History,
        Icons.Outlined.History
    )
    object Settings : Screen(
        "settings",
        "Settings",
        Icons.Filled.Settings,
        Icons.Outlined.Settings
    )
    object Onboarding : Screen(
        "onboarding",
        "Welcome",
        Icons.Filled.AccountBalanceWallet,
        Icons.Outlined.AccountBalanceWallet
    )
    object EquityDetail : Screen(
        "equity/{symbol}",
        "Stock Detail",
        Icons.Filled.Info,
        Icons.Outlined.Info
    ) {
        fun createRoute(symbol: String) = "equity/$symbol"
    }
    object TradeWithSymbol : Screen(
        "trade/{symbol}",
        "Trade",
        Icons.Filled.SwapVert,
        Icons.Outlined.SwapVert
    ) {
        fun createRoute(symbol: String) = "trade/$symbol"
    }
    object CryptoDetail : Screen(
        "crypto/{chain}",
        "Crypto Detail",
        Icons.Filled.CurrencyBitcoin,
        Icons.Outlined.CurrencyBitcoin
    ) {
        fun createRoute(chain: Chain) = "crypto/${chain.name}"
    }
    object SendCrypto : Screen(
        "crypto/{chain}/send",
        "Send Crypto",
        Icons.Filled.Send,
        Icons.Outlined.Send
    ) {
        fun createRoute(chain: Chain) = "crypto/${chain.name}/send"
    }
    object ReceiveCrypto : Screen(
        "crypto/{chain}/receive",
        "Receive Crypto",
        Icons.Filled.QrCodeScanner,
        Icons.Outlined.QrCodeScanner
    ) {
        fun createRoute(chain: Chain) = "crypto/${chain.name}/receive"
    }
    object P2PTrading : Screen(
        "p2p",
        "P2P Trading",
        Icons.Filled.SwapHoriz,
        Icons.Outlined.SwapHoriz
    )
    object Lending : Screen(
        "lending",
        "Lending",
        Icons.Filled.AccountBalance,
        Icons.Outlined.AccountBalance
    )
    object Inheritance : Screen(
        "inheritance",
        "Inheritance",
        Icons.Filled.FamilyRestroom,
        Icons.Outlined.FamilyRestroom
    )
}

val bottomNavItems = listOf(
    Screen.Portfolio,
    Screen.Market,
    Screen.Trade,
    Screen.Activity,
    Screen.Settings
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ShareHODLApp(
    viewModel: WalletViewModel = hiltViewModel()
) {
    val navController = rememberNavController()
    val uiState by viewModel.uiState.collectAsState()

    // For demo purposes, always start with the main app (skip onboarding)
    val startDestination = Screen.Portfolio.route

    // Track if we're on a detail screen (to hide bottom nav)
    val navBackStackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route
    val showBottomNav = currentRoute in bottomNavItems.map { it.route }

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        containerColor = DarkBackground,
        bottomBar = {
            if (showBottomNav) {
                NavigationBar(
                    containerColor = CardBackground,
                    contentColor = Color.White
                ) {
                    val currentDestination = navBackStackEntry?.destination

                    bottomNavItems.forEach { screen ->
                        val selected = currentDestination?.hierarchy?.any { it.route == screen.route } == true

                        NavigationBarItem(
                            icon = {
                                Icon(
                                    if (selected) screen.selectedIcon else screen.unselectedIcon,
                                    contentDescription = screen.title
                                )
                            },
                            label = {
                                Text(
                                    screen.title,
                                    fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal
                                )
                            },
                            selected = selected,
                            onClick = {
                                navController.navigate(screen.route) {
                                    popUpTo(navController.graph.findStartDestination().id) {
                                        saveState = true
                                    }
                                    launchSingleTop = true
                                    restoreState = true
                                }
                            },
                            colors = NavigationBarItemDefaults.colors(
                                selectedIconColor = PrimaryBlue,
                                selectedTextColor = PrimaryBlue,
                                unselectedIconColor = Color.White.copy(alpha = 0.5f),
                                unselectedTextColor = Color.White.copy(alpha = 0.5f),
                                indicatorColor = PrimaryBlue.copy(alpha = 0.15f)
                            )
                        )
                    }
                }
            }
        }
    ) { innerPadding ->
        NavHost(
            navController = navController,
            startDestination = startDestination,
            modifier = Modifier.padding(innerPadding)
        ) {
            // Onboarding
            composable(Screen.Onboarding.route) {
                OnboardingScreen(
                    viewModel = viewModel,
                    onWalletCreated = {
                        navController.navigate(Screen.Portfolio.route) {
                            popUpTo(Screen.Onboarding.route) { inclusive = true }
                        }
                    }
                )
            }

            // Portfolio (Main screen)
            composable(Screen.Portfolio.route) {
                PortfolioScreen(
                    onEquityClick = { symbol ->
                        navController.navigate(Screen.EquityDetail.createRoute(symbol))
                    },
                    onTradeClick = {
                        navController.navigate(Screen.Trade.route)
                    },
                    onCryptoClick = { chain ->
                        navController.navigate(Screen.CryptoDetail.createRoute(chain))
                    },
                    onP2PClick = {
                        navController.navigate(Screen.P2PTrading.route)
                    },
                    onLendingClick = {
                        navController.navigate(Screen.Lending.route)
                    },
                    onInheritanceClick = {
                        navController.navigate(Screen.Inheritance.route)
                    },
                    walletViewModel = viewModel
                )
            }

            // Market
            composable(Screen.Market.route) {
                MarketScreen(
                    onEquityClick = { symbol ->
                        navController.navigate(Screen.EquityDetail.createRoute(symbol))
                    }
                )
            }

            // Trade (no pre-selected symbol)
            composable(Screen.Trade.route) {
                EquityTradeScreen(
                    preSelectedSymbol = null,
                    onBack = null
                )
            }

            // Trade with pre-selected symbol
            composable(
                route = Screen.TradeWithSymbol.route,
                arguments = listOf(navArgument("symbol") { type = NavType.StringType })
            ) { backStackEntry ->
                val symbol = backStackEntry.arguments?.getString("symbol")
                EquityTradeScreen(
                    preSelectedSymbol = symbol,
                    onBack = { navController.popBackStack() }
                )
            }

            // Activity
            composable(Screen.Activity.route) {
                ActivityScreen()
            }

            // Settings
            composable(Screen.Settings.route) {
                SettingsScreen(
                    viewModel = viewModel,
                    onLogout = {
                        navController.navigate(Screen.Onboarding.route) {
                            popUpTo(0) { inclusive = true }
                        }
                    }
                )
            }

            // Equity Detail
            composable(
                route = Screen.EquityDetail.route,
                arguments = listOf(navArgument("symbol") { type = NavType.StringType })
            ) { backStackEntry ->
                val symbol = backStackEntry.arguments?.getString("symbol") ?: ""
                EquityDetailScreen(
                    symbol = symbol,
                    onBack = { navController.popBackStack() },
                    onTrade = {
                        navController.navigate(Screen.TradeWithSymbol.createRoute(symbol))
                    }
                )
            }

            // Crypto Detail
            composable(
                route = Screen.CryptoDetail.route,
                arguments = listOf(navArgument("chain") { type = NavType.StringType })
            ) { backStackEntry ->
                val chainName = backStackEntry.arguments?.getString("chain") ?: ""
                val chain = try { Chain.valueOf(chainName) } catch (e: Exception) { Chain.SHAREHODL }
                CryptoDetailScreen(
                    chain = chain,
                    viewModel = viewModel,
                    onBack = { navController.popBackStack() },
                    onSend = { navController.navigate(Screen.SendCrypto.createRoute(chain)) },
                    onReceive = { navController.navigate(Screen.ReceiveCrypto.createRoute(chain)) }
                )
            }

            // Send Crypto
            composable(
                route = Screen.SendCrypto.route,
                arguments = listOf(navArgument("chain") { type = NavType.StringType })
            ) { backStackEntry ->
                val chainName = backStackEntry.arguments?.getString("chain") ?: ""
                val chain = try { Chain.valueOf(chainName) } catch (e: Exception) { Chain.SHAREHODL }
                val demoAccounts = remember { DemoCryptoData.getDemoChainAccounts() }
                val account = demoAccounts.find { it.chain == chain }

                SendCryptoScreen(
                    chain = chain,
                    account = account,
                    onBack = { navController.popBackStack() },
                    onSend = { toAddress, amount ->
                        // In production: sign and broadcast transaction
                        navController.popBackStack()
                    }
                )
            }

            // Receive Crypto
            composable(
                route = Screen.ReceiveCrypto.route,
                arguments = listOf(navArgument("chain") { type = NavType.StringType })
            ) { backStackEntry ->
                val chainName = backStackEntry.arguments?.getString("chain") ?: ""
                val chain = try { Chain.valueOf(chainName) } catch (e: Exception) { Chain.SHAREHODL }
                val demoAccounts = remember { DemoCryptoData.getDemoChainAccounts() }
                val account = demoAccounts.find { it.chain == chain }

                ReceiveCryptoScreen(
                    chain = chain,
                    account = account,
                    onBack = { navController.popBackStack() }
                )
            }

            // P2P Trading
            composable(Screen.P2PTrading.route) {
                P2PTradingScreen(
                    viewModel = viewModel,
                    onBack = { navController.popBackStack() }
                )
            }

            // Lending
            composable(Screen.Lending.route) {
                LendingScreen(
                    viewModel = viewModel,
                    onBack = { navController.popBackStack() }
                )
            }

            // Inheritance
            composable(Screen.Inheritance.route) {
                InheritanceScreen(
                    viewModel = viewModel,
                    onBack = { navController.popBackStack() }
                )
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CryptoDetailScreen(
    chain: Chain,
    viewModel: WalletViewModel,
    onBack: () -> Unit,
    onSend: () -> Unit = {},
    onReceive: () -> Unit = {}
) {
    // Use demo data for display
    val demoAccounts = remember { DemoCryptoData.getDemoChainAccounts() }
    val account = demoAccounts.find { it.chain == chain }
    val transactions = remember { DemoCryptoData.getDemoTransactions(chain) }

    Scaffold(
        containerColor = DarkBackground,
        topBar = {
            TopAppBar(
                title = { Text(chain.displayName, fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back", tint = Color.White)
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = DarkBackground,
                    titleContentColor = Color.White
                )
            )
        }
    ) { padding ->
        androidx.compose.foundation.lazy.LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            // Chain Info Card
            item {
                Card(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    colors = CardDefaults.cardColors(containerColor = CardBackground),
                    shape = androidx.compose.foundation.shape.RoundedCornerShape(20.dp)
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(24.dp),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        // Chain icon
                        Box(
                            modifier = Modifier
                                .size(80.dp)
                                .background(chain.color.copy(alpha = 0.2f), androidx.compose.foundation.shape.CircleShape),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = chain.symbol,
                                style = MaterialTheme.typography.headlineMedium,
                                fontWeight = FontWeight.Bold,
                                color = chain.color
                            )
                        }

                        Spacer(modifier = Modifier.height(16.dp))

                        Text(
                            text = chain.displayName,
                            style = MaterialTheme.typography.headlineSmall,
                            fontWeight = FontWeight.Bold,
                            color = Color.White
                        )

                        Spacer(modifier = Modifier.height(8.dp))

                        // Balance
                        Text(
                            text = account?.formattedBalance ?: "0 ${chain.symbol}",
                            style = MaterialTheme.typography.displaySmall,
                            fontWeight = FontWeight.Bold,
                            color = Color.White
                        )

                        if (account?.balanceUsd != null) {
                            Text(
                                text = "$$${account.balanceUsd}",
                                style = MaterialTheme.typography.titleMedium,
                                color = Color.White.copy(alpha = 0.6f)
                            )
                        }

                        Spacer(modifier = Modifier.height(24.dp))

                        // Action Buttons
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(12.dp)
                        ) {
                            Button(
                                onClick = onReceive,
                                colors = ButtonDefaults.buttonColors(containerColor = chain.color),
                                shape = androidx.compose.foundation.shape.RoundedCornerShape(12.dp),
                                modifier = Modifier.weight(1f)
                            ) {
                                Icon(Icons.Outlined.QrCodeScanner, contentDescription = null, modifier = Modifier.size(20.dp))
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("Receive")
                            }

                            OutlinedButton(
                                onClick = onSend,
                                colors = ButtonDefaults.outlinedButtonColors(contentColor = Color.White),
                                shape = androidx.compose.foundation.shape.RoundedCornerShape(12.dp),
                                modifier = Modifier.weight(1f)
                            ) {
                                Icon(Icons.Outlined.Send, contentDescription = null, modifier = Modifier.size(20.dp))
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("Send")
                            }
                        }
                    }
                }
            }

            // Address Card
            item {
                Card(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp),
                    colors = CardDefaults.cardColors(containerColor = CardBackground),
                    shape = androidx.compose.foundation.shape.RoundedCornerShape(16.dp)
                ) {
                    Column(
                        modifier = Modifier.padding(16.dp)
                    ) {
                        Text(
                            text = "Wallet Address",
                            style = MaterialTheme.typography.labelMedium,
                            color = Color.White.copy(alpha = 0.6f)
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = account?.address ?: "No address",
                            style = MaterialTheme.typography.bodyMedium,
                            color = Color.White,
                            fontWeight = FontWeight.Medium
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = "Path: ${account?.derivationPath ?: "-"}",
                            style = MaterialTheme.typography.labelSmall,
                            color = Color.White.copy(alpha = 0.4f)
                        )
                    }
                }
            }

            item { Spacer(modifier = Modifier.height(16.dp)) }

            // Transactions Header
            item {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text(
                        text = "Transactions",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        color = Color.White
                    )
                    Text(
                        text = "${transactions.size} total",
                        style = MaterialTheme.typography.bodySmall,
                        color = Color.White.copy(alpha = 0.5f)
                    )
                }
            }

            item { Spacer(modifier = Modifier.height(8.dp)) }

            // Transaction History
            item {
                TransactionHistorySection(
                    transactions = transactions,
                    chain = chain,
                    onTransactionClick = { /* Open in explorer */ }
                )
            }

            item { Spacer(modifier = Modifier.height(16.dp)) }

            // Chain Info
            item {
                Card(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp),
                    colors = CardDefaults.cardColors(containerColor = CardBackground),
                    shape = androidx.compose.foundation.shape.RoundedCornerShape(16.dp)
                ) {
                    Column(
                        modifier = Modifier.padding(16.dp)
                    ) {
                        Text(
                            text = "Chain Information",
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.SemiBold,
                            color = Color.White
                        )
                        Spacer(modifier = Modifier.height(12.dp))

                        ChainInfoRow("Symbol", chain.symbol)
                        ChainInfoRow("Decimals", chain.decimals.toString())
                        ChainInfoRow("Smallest Unit", chain.smallestUnit)
                        if (chain.isToken) {
                            ChainInfoRow("Type", "Token (${chain.parentChain})")
                        }
                    }
                }
            }

            item { Spacer(modifier = Modifier.height(100.dp)) }
        }
    }
}

@Composable
private fun ChainInfoRow(label: String, value: String) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp),
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
            fontWeight = FontWeight.Medium,
            color = Color.White
        )
    }
}
