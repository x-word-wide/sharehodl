package com.sharehodl.ui.onboarding

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import android.app.Activity
import android.content.Context
import android.content.ContextWrapper
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.fragment.app.FragmentActivity
import com.sharehodl.viewmodel.WalletViewModel

/**
 * Find FragmentActivity from Context
 * Compose's LocalContext may be a ContextWrapper, so we need to traverse
 */
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
fun OnboardingScreen(
    viewModel: WalletViewModel,
    onWalletCreated: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsState()
    val generatedMnemonic by viewModel.generatedMnemonic.collectAsState()

    var currentScreen by remember { mutableStateOf(OnboardingStep.Welcome) }
    var importMnemonic by remember { mutableStateOf("") }

    val context = LocalContext.current
    val activity = context.findActivity()

    // Navigate to main when wallet is created
    LaunchedEffect(uiState.hasWallet) {
        if (uiState.hasWallet) {
            onWalletCreated()
        }
    }

    // Show mnemonic backup when generated
    LaunchedEffect(uiState.showMnemonicBackup) {
        if (uiState.showMnemonicBackup && generatedMnemonic != null) {
            currentScreen = OnboardingStep.BackupMnemonic
        }
    }

    Box(modifier = Modifier.fillMaxSize()) {
        when (currentScreen) {
            OnboardingStep.Welcome -> WelcomeScreen(
                onCreateWallet = {
                    viewModel.createWallet()
                },
                onImportWallet = {
                    currentScreen = OnboardingStep.ImportWallet
                }
            )

            OnboardingStep.BackupMnemonic -> BackupMnemonicScreen(
                mnemonic = generatedMnemonic ?: "",
                onConfirmed = {
                    // Go to verification step instead of saving directly
                    currentScreen = OnboardingStep.VerifyMnemonic
                },
                onBack = {
                    viewModel.cancelMnemonicBackup()
                    currentScreen = OnboardingStep.Welcome
                }
            )

            OnboardingStep.VerifyMnemonic -> VerifyMnemonicScreen(
                mnemonic = generatedMnemonic ?: "",
                onVerified = {
                    android.util.Log.d("OnboardingScreen", "Mnemonic verified! Saving wallet...")
                    android.widget.Toast.makeText(context, "Saving wallet...", android.widget.Toast.LENGTH_SHORT).show()
                    val fragmentActivity = activity
                    if (fragmentActivity != null) {
                        viewModel.confirmMnemonicBackup(fragmentActivity)
                    } else {
                        android.util.Log.e("OnboardingScreen", "Could not find FragmentActivity")
                        android.widget.Toast.makeText(context, "Error: Activity not found", android.widget.Toast.LENGTH_LONG).show()
                    }
                },
                onBack = {
                    currentScreen = OnboardingStep.BackupMnemonic
                }
            )

            OnboardingStep.ImportWallet -> ImportWalletScreen(
                mnemonic = importMnemonic,
                onMnemonicChange = { importMnemonic = it },
                onImport = {
                    val fragmentActivity = activity
                    if (fragmentActivity != null) {
                        viewModel.importWallet(fragmentActivity, importMnemonic)
                    } else {
                        android.util.Log.e("OnboardingScreen", "Could not find FragmentActivity")
                    }
                },
                onBack = {
                    importMnemonic = ""
                    currentScreen = OnboardingStep.Welcome
                }
            )
        }

        // Loading Overlay
        if (uiState.isLoading) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(MaterialTheme.colorScheme.background.copy(alpha = 0.8f)),
                contentAlignment = Alignment.Center
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    CircularProgressIndicator()
                    Spacer(modifier = Modifier.height(16.dp))
                    Text("Please wait...")
                }
            }
        }

        // Error Snackbar
        uiState.error?.let { error ->
            Snackbar(
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(16.dp),
                action = {
                    TextButton(onClick = { viewModel.clearError() }) {
                        Text("Dismiss")
                    }
                }
            ) {
                Text(error)
            }
        }
    }
}

enum class OnboardingStep {
    Welcome,
    BackupMnemonic,
    VerifyMnemonic,
    ImportWallet
}

@Composable
fun WelcomeScreen(
    onCreateWallet: () -> Unit,
    onImportWallet: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(32.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        // Logo/Icon
        Box(
            modifier = Modifier
                .size(120.dp)
                .clip(RoundedCornerShape(60.dp))
                .background(MaterialTheme.colorScheme.primary),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                Icons.Default.AccountBalanceWallet,
                contentDescription = null,
                modifier = Modifier.size(64.dp),
                tint = MaterialTheme.colorScheme.onPrimary
            )
        }

        Spacer(modifier = Modifier.height(32.dp))

        Text(
            text = "ShareHODL",
            style = MaterialTheme.typography.headlineLarge,
            fontWeight = FontWeight.Bold
        )

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = "Your secure crypto wallet for the ShareHODL blockchain",
            style = MaterialTheme.typography.bodyLarge,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        Spacer(modifier = Modifier.height(64.dp))

        // Create Wallet Button
        Button(
            onClick = onCreateWallet,
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp)
        ) {
            Icon(Icons.Default.Add, contentDescription = null)
            Spacer(modifier = Modifier.width(8.dp))
            Text("Create New Wallet")
        }

        Spacer(modifier = Modifier.height(16.dp))

        // Import Wallet Button
        OutlinedButton(
            onClick = onImportWallet,
            modifier = Modifier
                .fillMaxWidth()
                .height(56.dp)
        ) {
            Icon(Icons.Default.FileDownload, contentDescription = null)
            Spacer(modifier = Modifier.width(8.dp))
            Text("Import Existing Wallet")
        }

        Spacer(modifier = Modifier.height(48.dp))

        // Security Note
        Row(
            verticalAlignment = Alignment.CenterVertically,
            modifier = Modifier.padding(horizontal = 16.dp)
        ) {
            Icon(
                Icons.Default.Security,
                contentDescription = null,
                modifier = Modifier.size(20.dp),
                tint = MaterialTheme.colorScheme.primary
            )
            Spacer(modifier = Modifier.width(8.dp))
            Text(
                text = "Your keys are stored securely with biometric protection",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BackupMnemonicScreen(
    mnemonic: String,
    onConfirmed: () -> Unit,
    onBack: () -> Unit
) {
    var hasAgreed by remember { mutableStateOf(false) }
    val words = mnemonic.split(" ")

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Backup Your Wallet") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(24.dp)
                .verticalScroll(rememberScrollState())
        ) {
            // Warning Card
            Card(
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.errorContainer
                )
            ) {
                Row(
                    modifier = Modifier.padding(16.dp),
                    verticalAlignment = Alignment.Top
                ) {
                    Icon(
                        Icons.Default.Warning,
                        contentDescription = null,
                        tint = MaterialTheme.colorScheme.error
                    )
                    Spacer(modifier = Modifier.width(12.dp))
                    Text(
                        text = "Write down these 12 words in order. This is your only way to recover your wallet. Never share these words with anyone!",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onErrorContainer
                    )
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            // Mnemonic Words Grid (3 rows x 4 columns = 12 words)
            Card(
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant
                )
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    // 3 rows for 12 words (4 words per row)
                    val rowCount = (words.size + 3) / 4 // Ceiling division
                    for (row in 0 until rowCount) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(8.dp)
                        ) {
                            for (col in 0 until 4) {
                                val index = row * 4 + col
                                if (index < words.size) {
                                    MnemonicWord(
                                        number = index + 1,
                                        word = words[index],
                                        modifier = Modifier.weight(1f)
                                    )
                                }
                            }
                        }
                        if (row < rowCount - 1) {
                            Spacer(modifier = Modifier.height(8.dp))
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            // Confirmation Checkbox
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.fillMaxWidth()
            ) {
                Checkbox(
                    checked = hasAgreed,
                    onCheckedChange = { hasAgreed = it }
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "I have written down my recovery phrase and stored it safely",
                    style = MaterialTheme.typography.bodyMedium
                )
            }

            Spacer(modifier = Modifier.height(24.dp))

            // Confirm Button
            Button(
                onClick = onConfirmed,
                enabled = hasAgreed,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp)
            ) {
                Icon(Icons.Default.Check, contentDescription = null)
                Spacer(modifier = Modifier.width(8.dp))
                Text("I've Backed Up My Phrase")
            }
        }
    }
}

@Composable
fun MnemonicWord(
    number: Int,
    word: String,
    modifier: Modifier = Modifier
) {
    Surface(
        modifier = modifier,
        color = MaterialTheme.colorScheme.surface,
        shape = RoundedCornerShape(8.dp)
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 6.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = "$number.",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Spacer(modifier = Modifier.width(4.dp))
            Text(
                text = word,
                style = MaterialTheme.typography.bodySmall,
                fontWeight = FontWeight.Medium
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun VerifyMnemonicScreen(
    mnemonic: String,
    onVerified: () -> Unit,
    onBack: () -> Unit
) {
    val words = mnemonic.split(" ")

    // Pick 3 random word indices to verify (deterministic based on mnemonic hash)
    val verifyIndices = remember(mnemonic) {
        val seed = mnemonic.hashCode()
        val random = java.util.Random(seed.toLong())
        (0 until words.size).shuffled(random).take(3).sorted()
    }

    var word1 by remember { mutableStateOf("") }
    var word2 by remember { mutableStateOf("") }
    var word3 by remember { mutableStateOf("") }
    var showError by remember { mutableStateOf(false) }

    val isCorrect = word1.trim().lowercase() == words.getOrNull(verifyIndices[0])?.lowercase() &&
            word2.trim().lowercase() == words.getOrNull(verifyIndices[1])?.lowercase() &&
            word3.trim().lowercase() == words.getOrNull(verifyIndices[2])?.lowercase()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Verify Recovery Phrase") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(24.dp)
                .verticalScroll(rememberScrollState())
        ) {
            // Instructions
            Card(
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.primaryContainer
                )
            ) {
                Row(
                    modifier = Modifier.padding(16.dp),
                    verticalAlignment = Alignment.Top
                ) {
                    Icon(
                        Icons.Default.Info,
                        contentDescription = null,
                        tint = MaterialTheme.colorScheme.primary
                    )
                    Spacer(modifier = Modifier.width(12.dp))
                    Text(
                        text = "Please enter the following words from your recovery phrase to confirm you've backed it up correctly.",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onPrimaryContainer
                    )
                }
            }

            Spacer(modifier = Modifier.height(32.dp))

            // Word 1
            OutlinedTextField(
                value = word1,
                onValueChange = {
                    word1 = it
                    showError = false
                },
                label = { Text("Word #${verifyIndices[0] + 1}") },
                placeholder = { Text("Enter word ${verifyIndices[0] + 1}") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                isError = showError && word1.trim().lowercase() != words.getOrNull(verifyIndices[0])?.lowercase()
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Word 2
            OutlinedTextField(
                value = word2,
                onValueChange = {
                    word2 = it
                    showError = false
                },
                label = { Text("Word #${verifyIndices[1] + 1}") },
                placeholder = { Text("Enter word ${verifyIndices[1] + 1}") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                isError = showError && word2.trim().lowercase() != words.getOrNull(verifyIndices[1])?.lowercase()
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Word 3
            OutlinedTextField(
                value = word3,
                onValueChange = {
                    word3 = it
                    showError = false
                },
                label = { Text("Word #${verifyIndices[2] + 1}") },
                placeholder = { Text("Enter word ${verifyIndices[2] + 1}") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                isError = showError && word3.trim().lowercase() != words.getOrNull(verifyIndices[2])?.lowercase()
            )

            if (showError) {
                Spacer(modifier = Modifier.height(16.dp))
                Text(
                    text = "One or more words are incorrect. Please check your backup and try again.",
                    color = MaterialTheme.colorScheme.error,
                    style = MaterialTheme.typography.bodySmall
                )
            }

            Spacer(modifier = Modifier.height(32.dp))

            // Verify Button
            Button(
                onClick = {
                    if (isCorrect) {
                        onVerified()
                    } else {
                        showError = true
                    }
                },
                enabled = word1.isNotBlank() && word2.isNotBlank() && word3.isNotBlank(),
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp)
            ) {
                Icon(Icons.Default.Check, contentDescription = null)
                Spacer(modifier = Modifier.width(8.dp))
                Text("Verify & Create Wallet")
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ImportWalletScreen(
    mnemonic: String,
    onMnemonicChange: (String) -> Unit,
    onImport: () -> Unit,
    onBack: () -> Unit
) {
    val wordCount = if (mnemonic.isBlank()) 0 else mnemonic.trim().split("\\s+".toRegex()).size

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Import Wallet") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(24.dp)
        ) {
            Text(
                text = "Enter your recovery phrase",
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold
            )

            Spacer(modifier = Modifier.height(8.dp))

            Text(
                text = "Enter your 12, 15, 18, 21, or 24 word recovery phrase to restore your wallet.",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            Spacer(modifier = Modifier.height(24.dp))

            // Mnemonic Input
            OutlinedTextField(
                value = mnemonic,
                onValueChange = onMnemonicChange,
                label = { Text("Recovery Phrase") },
                placeholder = { Text("Enter your words separated by spaces") },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(200.dp),
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Text),
                supportingText = {
                    Text("$wordCount words entered")
                }
            )

            Spacer(modifier = Modifier.height(24.dp))

            // Info Card
            Card(
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.primaryContainer
                )
            ) {
                Row(
                    modifier = Modifier.padding(16.dp),
                    verticalAlignment = Alignment.Top
                ) {
                    Icon(
                        Icons.Default.Info,
                        contentDescription = null,
                        tint = MaterialTheme.colorScheme.primary
                    )
                    Spacer(modifier = Modifier.width(12.dp))
                    Text(
                        text = "Your recovery phrase will be encrypted and stored securely. Make sure you're in a private location.",
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onPrimaryContainer
                    )
                }
            }

            Spacer(modifier = Modifier.weight(1f))

            // Import Button
            Button(
                onClick = onImport,
                enabled = wordCount in listOf(12, 15, 18, 21, 24),
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp)
            ) {
                Icon(Icons.Default.FileDownload, contentDescription = null)
                Spacer(modifier = Modifier.width(8.dp))
                Text("Import Wallet")
            }
        }
    }
}
