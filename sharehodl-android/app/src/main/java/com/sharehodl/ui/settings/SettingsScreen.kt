package com.sharehodl.ui.settings

import android.content.Context
import android.content.ContextWrapper
import android.widget.Toast
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
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
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.fragment.app.FragmentActivity
import com.sharehodl.config.NetworkConfig
import com.sharehodl.service.NetworkType
import com.sharehodl.service.OrderType
import com.sharehodl.ui.portfolio.*
import com.sharehodl.viewmodel.WalletViewModel

private fun Context.findActivity(): FragmentActivity? {
    var context = this
    while (context is ContextWrapper) {
        if (context is FragmentActivity) return context
        context = context.baseContext
    }
    return null
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen(
    viewModel: WalletViewModel,
    onLogout: () -> Unit
) {
    val walletAddress by viewModel.walletAddress.collectAsState()
    val recoveryPhrase by viewModel.recoveryPhrase.collectAsState()
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current
    val activity = context.findActivity()
    val uriHandler = LocalUriHandler.current

    // Dialog states
    var showDeleteDialog by remember { mutableStateOf(false) }
    var showLogoutDialog by remember { mutableStateOf(false) }
    var showRecoveryPhraseDialog by remember { mutableStateOf(false) }
    var showPriceAlertsDialog by remember { mutableStateOf(false) }
    var showOrderTypeDialog by remember { mutableStateOf(false) }
    var showNetworkDialog by remember { mutableStateOf(false) }
    var showChangePinDialog by remember { mutableStateOf(false) }
    var show2FADialog by remember { mutableStateOf(false) }

    // Settings state (persisted via SharedPreferences)
    val prefs = remember { context.getSharedPreferences("sharehodl_settings", Context.MODE_PRIVATE) }

    var priceAlertsEnabled by remember { mutableStateOf(prefs.getBoolean("price_alerts_enabled", false)) }
    var selectedOrderType by remember { mutableStateOf(OrderType.fromString(prefs.getString("default_order_type", "MARKET") ?: "MARKET")) }
    var tradeConfirmationsEnabled by remember { mutableStateOf(prefs.getBoolean("trade_confirmations", true)) }
    var biometricEnabled by remember { mutableStateOf(prefs.getBoolean("biometric_enabled", viewModel.isBiometricsAvailable())) }
    var twoFactorEnabled by remember { mutableStateOf(prefs.getBoolean("2fa_enabled", false)) }
    var selectedNetwork by remember { mutableStateOf(if (NetworkConfig.isProduction) NetworkType.MAINNET else NetworkType.TESTNET) }

    // Show recovery phrase dialog when phrase is loaded
    LaunchedEffect(recoveryPhrase) {
        if (recoveryPhrase != null) {
            showRecoveryPhraseDialog = true
        }
    }

    Scaffold(
        containerColor = DarkBackground,
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        "Settings",
                        fontWeight = FontWeight.Bold
                    )
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = DarkBackground,
                    titleContentColor = Color.White
                )
            )
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
            contentPadding = PaddingValues(16.dp)
        ) {
            // Account Card
            item {
                AccountCard(walletAddress = walletAddress ?: "")
            }

            item { Spacer(modifier = Modifier.height(24.dp)) }

            // Trading Preferences Section
            item {
                SettingsSectionHeader(title = "Trading Preferences")
            }

            item {
                SettingsToggleItem(
                    icon = Icons.Outlined.Notifications,
                    title = "Price Alerts",
                    subtitle = if (priceAlertsEnabled) "Enabled" else "Disabled",
                    isChecked = priceAlertsEnabled,
                    onToggle = { enabled ->
                        priceAlertsEnabled = enabled
                        prefs.edit().putBoolean("price_alerts_enabled", enabled).apply()
                        Toast.makeText(
                            context,
                            if (enabled) "Price alerts enabled" else "Price alerts disabled",
                            Toast.LENGTH_SHORT
                        ).show()
                    }
                )
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.TrendingUp,
                    title = "Default Order Type",
                    subtitle = selectedOrderType.displayName,
                    onClick = { showOrderTypeDialog = true }
                )
            }

            item {
                SettingsToggleItem(
                    icon = Icons.Outlined.Receipt,
                    title = "Trade Confirmations",
                    subtitle = if (tradeConfirmationsEnabled) "Required before placing orders" else "Disabled",
                    isChecked = tradeConfirmationsEnabled,
                    onToggle = { enabled ->
                        tradeConfirmationsEnabled = enabled
                        prefs.edit().putBoolean("trade_confirmations", enabled).apply()
                    }
                )
            }

            item { Spacer(modifier = Modifier.height(24.dp)) }

            // Security Section
            item {
                SettingsSectionHeader(title = "Security")
            }

            item {
                SettingsToggleItem(
                    icon = Icons.Outlined.Fingerprint,
                    title = "Biometric Authentication",
                    subtitle = if (!viewModel.isBiometricsAvailable()) "Not available on this device"
                    else if (biometricEnabled) "Enabled" else "Disabled",
                    isChecked = biometricEnabled && viewModel.isBiometricsAvailable(),
                    enabled = viewModel.isBiometricsAvailable(),
                    onToggle = { enabled ->
                        biometricEnabled = enabled
                        prefs.edit().putBoolean("biometric_enabled", enabled).apply()
                        Toast.makeText(
                            context,
                            if (enabled) "Biometric authentication enabled" else "Biometric authentication disabled",
                            Toast.LENGTH_SHORT
                        ).show()
                    }
                )
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.Lock,
                    title = "Change PIN",
                    subtitle = "Update your security PIN",
                    onClick = { showChangePinDialog = true }
                )
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.Key,
                    title = "View Recovery Phrase",
                    subtitle = "Requires authentication",
                    onClick = {
                        activity?.let { viewModel.viewRecoveryPhrase(it) }
                    }
                )
            }

            item {
                SettingsToggleItem(
                    icon = Icons.Outlined.Security,
                    title = "Two-Factor Authentication",
                    subtitle = if (twoFactorEnabled) "Enabled" else "Add extra security",
                    isChecked = twoFactorEnabled,
                    onToggle = { enabled ->
                        if (enabled) {
                            show2FADialog = true
                        } else {
                            twoFactorEnabled = false
                            prefs.edit().putBoolean("2fa_enabled", false).apply()
                            Toast.makeText(context, "2FA disabled", Toast.LENGTH_SHORT).show()
                        }
                    }
                )
            }

            item { Spacer(modifier = Modifier.height(24.dp)) }

            // Network Section
            item {
                SettingsSectionHeader(title = "Network")
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.Language,
                    title = "Network",
                    subtitle = selectedNetwork.displayName,
                    onClick = { showNetworkDialog = true }
                )
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.Storage,
                    title = "RPC Endpoint",
                    subtitle = if (selectedNetwork == NetworkType.MAINNET) "api.sharehodl.com" else "testnet-api.sharehodl.com",
                    onClick = {
                        Toast.makeText(context, "Custom RPC coming soon", Toast.LENGTH_SHORT).show()
                    }
                )
            }

            item { Spacer(modifier = Modifier.height(24.dp)) }

            // Legal & Support Section
            item {
                SettingsSectionHeader(title = "Legal & Support")
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.Info,
                    title = "App Version",
                    subtitle = "1.0.0 (Build 1)",
                    onClick = { }
                )
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.Description,
                    title = "Terms of Service",
                    subtitle = "View terms and conditions",
                    onClick = { uriHandler.openUri("https://sharehodl.com/terms") }
                )
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.PrivacyTip,
                    title = "Privacy Policy",
                    subtitle = "View privacy policy",
                    onClick = { uriHandler.openUri("https://sharehodl.com/privacy") }
                )
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.Gavel,
                    title = "Regulatory Information",
                    subtitle = "SEC and FINRA disclosures",
                    onClick = { uriHandler.openUri("https://sharehodl.com/disclosures") }
                )
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.Help,
                    title = "Help & Support",
                    subtitle = "Get help with the app",
                    onClick = { uriHandler.openUri("https://sharehodl.com/support") }
                )
            }

            item { Spacer(modifier = Modifier.height(24.dp)) }

            // Danger Zone
            item {
                SettingsSectionHeader(title = "Account")
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.Logout,
                    title = "Log Out",
                    subtitle = "Sign out of your account",
                    onClick = { showLogoutDialog = true },
                    isDanger = false
                )
            }

            item {
                SettingsItem(
                    icon = Icons.Outlined.DeleteForever,
                    title = "Delete Account",
                    subtitle = "Permanently remove your account",
                    onClick = { showDeleteDialog = true },
                    isDanger = true
                )
            }

            item { Spacer(modifier = Modifier.height(100.dp)) }
        }
    }

    // Delete Confirmation Dialog
    if (showDeleteDialog) {
        AlertDialog(
            onDismissRequest = { showDeleteDialog = false },
            containerColor = CardBackground,
            icon = {
                Box(
                    modifier = Modifier
                        .size(56.dp)
                        .clip(CircleShape)
                        .background(LossRed.copy(alpha = 0.15f)),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        Icons.Default.Warning,
                        contentDescription = null,
                        tint = LossRed,
                        modifier = Modifier.size(28.dp)
                    )
                }
            },
            title = {
                Text(
                    "Delete Wallet?",
                    color = Color.White,
                    fontWeight = FontWeight.Bold
                )
            },
            text = {
                Text(
                    "This will remove your wallet from this device. Make sure you have backed up your recovery phrase. This action cannot be undone.",
                    color = Color.White.copy(alpha = 0.7f)
                )
            },
            confirmButton = {
                Button(
                    onClick = {
                        viewModel.deleteWallet()
                        showDeleteDialog = false
                        onLogout()
                    },
                    colors = ButtonDefaults.buttonColors(
                        containerColor = LossRed
                    ),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text("Delete Wallet")
                }
            },
            dismissButton = {
                TextButton(
                    onClick = { showDeleteDialog = false },
                    colors = ButtonDefaults.textButtonColors(
                        contentColor = Color.White
                    )
                ) {
                    Text("Cancel")
                }
            }
        )
    }

    // Logout Confirmation Dialog
    if (showLogoutDialog) {
        AlertDialog(
            onDismissRequest = { showLogoutDialog = false },
            containerColor = CardBackground,
            icon = {
                Box(
                    modifier = Modifier
                        .size(56.dp)
                        .clip(CircleShape)
                        .background(PrimaryBlue.copy(alpha = 0.15f)),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        Icons.Default.Logout,
                        contentDescription = null,
                        tint = PrimaryBlue,
                        modifier = Modifier.size(28.dp)
                    )
                }
            },
            title = {
                Text(
                    "Log Out?",
                    color = Color.White,
                    fontWeight = FontWeight.Bold
                )
            },
            text = {
                Text(
                    "You will need to authenticate again to access your wallet.",
                    color = Color.White.copy(alpha = 0.7f)
                )
            },
            confirmButton = {
                Button(
                    onClick = {
                        showLogoutDialog = false
                        onLogout()
                    },
                    colors = ButtonDefaults.buttonColors(
                        containerColor = PrimaryBlue
                    ),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text("Log Out")
                }
            },
            dismissButton = {
                TextButton(
                    onClick = { showLogoutDialog = false },
                    colors = ButtonDefaults.textButtonColors(
                        contentColor = Color.White
                    )
                ) {
                    Text("Cancel")
                }
            }
        )
    }

    // Recovery Phrase Dialog
    if (showRecoveryPhraseDialog && recoveryPhrase != null) {
        AlertDialog(
            onDismissRequest = {
                showRecoveryPhraseDialog = false
                viewModel.clearRecoveryPhrase()
            },
            containerColor = CardBackground,
            icon = {
                Box(
                    modifier = Modifier
                        .size(56.dp)
                        .clip(CircleShape)
                        .background(PrimaryBlue.copy(alpha = 0.15f)),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        Icons.Default.Key,
                        contentDescription = null,
                        tint = PrimaryBlue,
                        modifier = Modifier.size(28.dp)
                    )
                }
            },
            title = {
                Text(
                    "Recovery Phrase",
                    color = Color.White,
                    fontWeight = FontWeight.Bold
                )
            },
            text = {
                Column {
                    // Warning Card
                    Card(
                        colors = CardDefaults.cardColors(
                            containerColor = LossRed.copy(alpha = 0.15f)
                        ),
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Row(
                            modifier = Modifier.padding(12.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Icon(
                                Icons.Default.Warning,
                                contentDescription = null,
                                tint = LossRed,
                                modifier = Modifier.size(20.dp)
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(
                                text = "Never share these words with anyone!",
                                style = MaterialTheme.typography.bodySmall,
                                color = LossRed
                            )
                        }
                    }

                    Spacer(modifier = Modifier.height(16.dp))

                    // Display words in a grid
                    val words = recoveryPhrase!!.split(" ")
                    Card(
                        colors = CardDefaults.cardColors(
                            containerColor = SurfaceColor
                        ),
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(12.dp)
                    ) {
                        Column(modifier = Modifier.padding(12.dp)) {
                            val rowCount = (words.size + 2) / 3
                            for (row in 0 until rowCount) {
                                Row(
                                    modifier = Modifier.fillMaxWidth(),
                                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                                ) {
                                    for (col in 0 until 3) {
                                        val index = row * 3 + col
                                        if (index < words.size) {
                                            Box(
                                                modifier = Modifier
                                                    .weight(1f)
                                                    .background(
                                                        CardBackground,
                                                        RoundedCornerShape(8.dp)
                                                    )
                                                    .padding(horizontal = 8.dp, vertical = 6.dp)
                                            ) {
                                                Row(verticalAlignment = Alignment.CenterVertically) {
                                                    Text(
                                                        text = "${index + 1}.",
                                                        style = MaterialTheme.typography.labelSmall,
                                                        color = Color.White.copy(alpha = 0.4f)
                                                    )
                                                    Spacer(modifier = Modifier.width(4.dp))
                                                    Text(
                                                        text = words[index],
                                                        style = MaterialTheme.typography.bodySmall,
                                                        fontWeight = FontWeight.Medium,
                                                        color = Color.White
                                                    )
                                                }
                                            }
                                        }
                                    }
                                }
                                if (row < rowCount - 1) {
                                    Spacer(modifier = Modifier.height(6.dp))
                                }
                            }
                        }
                    }
                }
            },
            confirmButton = {
                Button(
                    onClick = {
                        showRecoveryPhraseDialog = false
                        viewModel.clearRecoveryPhrase()
                    },
                    colors = ButtonDefaults.buttonColors(
                        containerColor = PrimaryBlue
                    ),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text("Done")
                }
            }
        )
    }

    // Order Type Selection Dialog
    if (showOrderTypeDialog) {
        AlertDialog(
            onDismissRequest = { showOrderTypeDialog = false },
            containerColor = CardBackground,
            title = {
                Text(
                    "Default Order Type",
                    color = Color.White,
                    fontWeight = FontWeight.Bold
                )
            },
            text = {
                Column {
                    OrderType.entries.forEach { orderType ->
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(vertical = 8.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            RadioButton(
                                selected = selectedOrderType == orderType,
                                onClick = {
                                    selectedOrderType = orderType
                                    prefs.edit().putString("default_order_type", orderType.name).apply()
                                },
                                colors = RadioButtonDefaults.colors(
                                    selectedColor = PrimaryBlue,
                                    unselectedColor = Color.White.copy(alpha = 0.5f)
                                )
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text(
                                text = orderType.displayName,
                                color = Color.White,
                                style = MaterialTheme.typography.bodyLarge
                            )
                        }
                    }
                }
            },
            confirmButton = {
                Button(
                    onClick = { showOrderTypeDialog = false },
                    colors = ButtonDefaults.buttonColors(
                        containerColor = PrimaryBlue
                    ),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text("Done")
                }
            }
        )
    }

    // Network Selection Dialog
    if (showNetworkDialog) {
        AlertDialog(
            onDismissRequest = { showNetworkDialog = false },
            containerColor = CardBackground,
            title = {
                Text(
                    "Select Network",
                    color = Color.White,
                    fontWeight = FontWeight.Bold
                )
            },
            text = {
                Column {
                    NetworkType.entries.forEach { network ->
                        Card(
                            onClick = {
                                selectedNetwork = network
                                NetworkConfig.setEnvironment(network == NetworkType.MAINNET)
                                showNetworkDialog = false
                                Toast.makeText(
                                    context,
                                    "Switched to ${network.displayName}",
                                    Toast.LENGTH_SHORT
                                ).show()
                            },
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(vertical = 4.dp),
                            colors = CardDefaults.cardColors(
                                containerColor = if (selectedNetwork == network)
                                    PrimaryBlue.copy(alpha = 0.15f) else SurfaceColor
                            ),
                            shape = RoundedCornerShape(12.dp)
                        ) {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(16.dp),
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Column(modifier = Modifier.weight(1f)) {
                                    Text(
                                        text = network.displayName,
                                        color = Color.White,
                                        fontWeight = FontWeight.Medium
                                    )
                                    Text(
                                        text = network.description,
                                        color = Color.White.copy(alpha = 0.5f),
                                        style = MaterialTheme.typography.bodySmall
                                    )
                                }
                                if (selectedNetwork == network) {
                                    Icon(
                                        Icons.Default.Check,
                                        contentDescription = null,
                                        tint = PrimaryBlue
                                    )
                                }
                            }
                        }
                    }
                }
            },
            confirmButton = { }
        )
    }

    // Change PIN Dialog
    if (showChangePinDialog) {
        var currentPin by remember { mutableStateOf("") }
        var newPin by remember { mutableStateOf("") }
        var confirmPin by remember { mutableStateOf("") }
        var pinError by remember { mutableStateOf<String?>(null) }

        AlertDialog(
            onDismissRequest = { showChangePinDialog = false },
            containerColor = CardBackground,
            title = {
                Text(
                    "Change PIN",
                    color = Color.White,
                    fontWeight = FontWeight.Bold
                )
            },
            text = {
                Column {
                    val hasExistingPin = prefs.getString("pin_hash", null) != null

                    if (hasExistingPin) {
                        OutlinedTextField(
                            value = currentPin,
                            onValueChange = { if (it.length <= 6) currentPin = it },
                            label = { Text("Current PIN") },
                            visualTransformation = PasswordVisualTransformation(),
                            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.NumberPassword),
                            singleLine = true,
                            modifier = Modifier.fillMaxWidth(),
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedTextColor = Color.White,
                                unfocusedTextColor = Color.White,
                                focusedBorderColor = PrimaryBlue,
                                unfocusedBorderColor = Color.White.copy(alpha = 0.3f),
                                focusedLabelColor = PrimaryBlue,
                                unfocusedLabelColor = Color.White.copy(alpha = 0.5f)
                            )
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                    }

                    OutlinedTextField(
                        value = newPin,
                        onValueChange = { if (it.length <= 6) newPin = it },
                        label = { Text("New PIN (6 digits)") },
                        visualTransformation = PasswordVisualTransformation(),
                        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.NumberPassword),
                        singleLine = true,
                        modifier = Modifier.fillMaxWidth(),
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedTextColor = Color.White,
                            unfocusedTextColor = Color.White,
                            focusedBorderColor = PrimaryBlue,
                            unfocusedBorderColor = Color.White.copy(alpha = 0.3f),
                            focusedLabelColor = PrimaryBlue,
                            unfocusedLabelColor = Color.White.copy(alpha = 0.5f)
                        )
                    )
                    Spacer(modifier = Modifier.height(12.dp))

                    OutlinedTextField(
                        value = confirmPin,
                        onValueChange = { if (it.length <= 6) confirmPin = it },
                        label = { Text("Confirm PIN") },
                        visualTransformation = PasswordVisualTransformation(),
                        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.NumberPassword),
                        singleLine = true,
                        modifier = Modifier.fillMaxWidth(),
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedTextColor = Color.White,
                            unfocusedTextColor = Color.White,
                            focusedBorderColor = PrimaryBlue,
                            unfocusedBorderColor = Color.White.copy(alpha = 0.3f),
                            focusedLabelColor = PrimaryBlue,
                            unfocusedLabelColor = Color.White.copy(alpha = 0.5f)
                        )
                    )

                    if (pinError != null) {
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = pinError!!,
                            color = LossRed,
                            style = MaterialTheme.typography.bodySmall
                        )
                    }
                }
            },
            confirmButton = {
                Button(
                    onClick = {
                        val hasExistingPin = prefs.getString("pin_hash", null) != null

                        // Validate
                        if (hasExistingPin) {
                            val storedHash = prefs.getString("pin_hash", null)
                            val currentHash = hashPin(currentPin)
                            if (storedHash != currentHash) {
                                pinError = "Current PIN is incorrect"
                                return@Button
                            }
                        }

                        if (newPin.length != 6) {
                            pinError = "PIN must be 6 digits"
                            return@Button
                        }

                        if (newPin != confirmPin) {
                            pinError = "PINs do not match"
                            return@Button
                        }

                        // Save new PIN
                        val newHash = hashPin(newPin)
                        prefs.edit().putString("pin_hash", newHash).apply()
                        showChangePinDialog = false
                        Toast.makeText(context, "PIN updated successfully", Toast.LENGTH_SHORT).show()
                    },
                    colors = ButtonDefaults.buttonColors(
                        containerColor = PrimaryBlue
                    ),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text("Save")
                }
            },
            dismissButton = {
                TextButton(
                    onClick = { showChangePinDialog = false },
                    colors = ButtonDefaults.textButtonColors(
                        contentColor = Color.White
                    )
                ) {
                    Text("Cancel")
                }
            }
        )
    }

    // 2FA Setup Dialog
    if (show2FADialog) {
        AlertDialog(
            onDismissRequest = { show2FADialog = false },
            containerColor = CardBackground,
            icon = {
                Box(
                    modifier = Modifier
                        .size(56.dp)
                        .clip(CircleShape)
                        .background(PrimaryBlue.copy(alpha = 0.15f)),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        Icons.Default.Security,
                        contentDescription = null,
                        tint = PrimaryBlue,
                        modifier = Modifier.size(28.dp)
                    )
                }
            },
            title = {
                Text(
                    "Enable Two-Factor Authentication",
                    color = Color.White,
                    fontWeight = FontWeight.Bold
                )
            },
            text = {
                Column {
                    Text(
                        "Two-factor authentication adds an extra layer of security to your account.",
                        color = Color.White.copy(alpha = 0.7f)
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    Text(
                        "You will need to use an authenticator app like Google Authenticator or Authy.",
                        color = Color.White.copy(alpha = 0.7f)
                    )
                }
            },
            confirmButton = {
                Button(
                    onClick = {
                        twoFactorEnabled = true
                        prefs.edit().putBoolean("2fa_enabled", true).apply()
                        show2FADialog = false
                        Toast.makeText(context, "2FA enabled", Toast.LENGTH_SHORT).show()
                    },
                    colors = ButtonDefaults.buttonColors(
                        containerColor = PrimaryBlue
                    ),
                    shape = RoundedCornerShape(12.dp)
                ) {
                    Text("Enable")
                }
            },
            dismissButton = {
                TextButton(
                    onClick = { show2FADialog = false },
                    colors = ButtonDefaults.textButtonColors(
                        contentColor = Color.White
                    )
                ) {
                    Text("Cancel")
                }
            }
        )
    }

    // Loading overlay
    if (uiState.isLoading) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(DarkBackground.copy(alpha = 0.8f)),
            contentAlignment = Alignment.Center
        ) {
            CircularProgressIndicator(color = PrimaryBlue)
        }
    }
}

private fun hashPin(pin: String): String {
    val bytes = pin.toByteArray()
    val digest = java.security.MessageDigest.getInstance("SHA-256")
    val hashBytes = digest.digest(bytes)
    return hashBytes.joinToString("") { "%02x".format(it) }
}

@Composable
fun AccountCard(walletAddress: String) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = CardBackground
        ),
        shape = RoundedCornerShape(16.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(20.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Avatar
            Box(
                modifier = Modifier
                    .size(56.dp)
                    .clip(CircleShape)
                    .background(PrimaryBlue.copy(alpha = 0.15f)),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    Icons.Default.Person,
                    contentDescription = null,
                    modifier = Modifier.size(28.dp),
                    tint = PrimaryBlue
                )
            }

            Spacer(modifier = Modifier.width(16.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = "My Account",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = if (walletAddress.length > 20)
                        "${walletAddress.take(10)}...${walletAddress.takeLast(8)}"
                    else walletAddress,
                    style = MaterialTheme.typography.bodySmall,
                    color = Color.White.copy(alpha = 0.5f)
                )
            }

            // Verified badge
            Surface(
                color = GainGreen.copy(alpha = 0.15f),
                shape = RoundedCornerShape(8.dp)
            ) {
                Row(
                    modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        Icons.Default.Verified,
                        contentDescription = null,
                        modifier = Modifier.size(14.dp),
                        tint = GainGreen
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "Verified",
                        style = MaterialTheme.typography.labelSmall,
                        fontWeight = FontWeight.Medium,
                        color = GainGreen
                    )
                }
            }
        }
    }
}

@Composable
fun SettingsSectionHeader(title: String) {
    Text(
        text = title.uppercase(),
        style = MaterialTheme.typography.labelSmall,
        fontWeight = FontWeight.Bold,
        color = Color.White.copy(alpha = 0.4f),
        letterSpacing = 1.sp,
        modifier = Modifier.padding(vertical = 8.dp, horizontal = 4.dp)
    )
}

@Composable
fun SettingsItem(
    icon: ImageVector,
    title: String,
    subtitle: String,
    onClick: () -> Unit,
    isDanger: Boolean = false
) {
    Card(
        onClick = onClick,
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = CardBackground
        ),
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(40.dp)
                    .clip(CircleShape)
                    .background(
                        if (isDanger) LossRed.copy(alpha = 0.15f)
                        else SurfaceColor
                    ),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    icon,
                    contentDescription = null,
                    modifier = Modifier.size(20.dp),
                    tint = if (isDanger) LossRed else Color.White.copy(alpha = 0.8f)
                )
            }

            Spacer(modifier = Modifier.width(14.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = title,
                    style = MaterialTheme.typography.bodyLarge,
                    fontWeight = FontWeight.Medium,
                    color = if (isDanger) LossRed else Color.White
                )
                Text(
                    text = subtitle,
                    style = MaterialTheme.typography.bodySmall,
                    color = if (isDanger) LossRed.copy(alpha = 0.6f) else Color.White.copy(alpha = 0.5f)
                )
            }

            Icon(
                Icons.Default.ChevronRight,
                contentDescription = null,
                modifier = Modifier.size(20.dp),
                tint = Color.White.copy(alpha = 0.3f)
            )
        }
    }

    Spacer(modifier = Modifier.height(8.dp))
}

@Composable
fun SettingsToggleItem(
    icon: ImageVector,
    title: String,
    subtitle: String,
    isChecked: Boolean,
    onToggle: (Boolean) -> Unit,
    enabled: Boolean = true
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = CardBackground
        ),
        shape = RoundedCornerShape(12.dp)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(40.dp)
                    .clip(CircleShape)
                    .background(SurfaceColor),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    icon,
                    contentDescription = null,
                    modifier = Modifier.size(20.dp),
                    tint = if (enabled) Color.White.copy(alpha = 0.8f) else Color.White.copy(alpha = 0.3f)
                )
            }

            Spacer(modifier = Modifier.width(14.dp))

            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = title,
                    style = MaterialTheme.typography.bodyLarge,
                    fontWeight = FontWeight.Medium,
                    color = if (enabled) Color.White else Color.White.copy(alpha = 0.5f)
                )
                Text(
                    text = subtitle,
                    style = MaterialTheme.typography.bodySmall,
                    color = Color.White.copy(alpha = 0.5f)
                )
            }

            Switch(
                checked = isChecked,
                onCheckedChange = onToggle,
                enabled = enabled,
                colors = SwitchDefaults.colors(
                    checkedThumbColor = Color.White,
                    checkedTrackColor = PrimaryBlue,
                    uncheckedThumbColor = Color.White.copy(alpha = 0.8f),
                    uncheckedTrackColor = SurfaceColor,
                    disabledCheckedThumbColor = Color.White.copy(alpha = 0.5f),
                    disabledCheckedTrackColor = PrimaryBlue.copy(alpha = 0.3f),
                    disabledUncheckedThumbColor = Color.White.copy(alpha = 0.3f),
                    disabledUncheckedTrackColor = SurfaceColor.copy(alpha = 0.5f)
                )
            )
        }
    }

    Spacer(modifier = Modifier.height(8.dp))
}
