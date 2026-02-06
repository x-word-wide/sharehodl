package com.sharehodl.viewmodel

import androidx.fragment.app.FragmentActivity
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.sharehodl.config.FeeLevel
import com.sharehodl.config.NetworkConfig
import com.sharehodl.model.Chain
import com.sharehodl.model.ChainAccount
import com.sharehodl.model.CryptoTransaction
import com.sharehodl.service.*
import com.sharehodl.service.api.MultiChainApiService
import com.sharehodl.service.api.PriceService
import com.sharehodl.service.tx.CosmosTransactionBuilder
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

/**
 * Main ViewModel for wallet state management
 *
 * SECURITY: Private keys are NEVER stored in UI state.
 * Keys are derived, used, and cleared in a single atomic operation.
 *
 * MULTI-CHAIN: One mnemonic derives addresses for all supported networks
 * (like Trust Wallet). ShareHODL is the primary network.
 */
@HiltViewModel
class WalletViewModel @Inject constructor(
    private val cryptoService: CryptoService,
    private val keystoreService: KeystoreService,
    private val blockchainService: BlockchainService,
    private val multiChainApiService: MultiChainApiService,
    private val priceService: PriceService,
    private val cosmosTransactionBuilder: CosmosTransactionBuilder
) : ViewModel() {

    // MARK: - State

    private val _uiState = MutableStateFlow(WalletUiState())
    val uiState: StateFlow<WalletUiState> = _uiState.asStateFlow()

    private val _walletAddress = MutableStateFlow<String?>(null)
    val walletAddress: StateFlow<String?> = _walletAddress.asStateFlow()

    // Multi-chain accounts - all derived from single mnemonic
    private val _chainAccounts = MutableStateFlow<List<ChainAccount>>(emptyList())
    val chainAccounts: StateFlow<List<ChainAccount>> = _chainAccounts.asStateFlow()

    // Currently selected chain (default: ShareHODL)
    private val _selectedChain = MutableStateFlow(Chain.SHAREHODL)
    val selectedChain: StateFlow<Chain> = _selectedChain.asStateFlow()

    private val _balance = MutableStateFlow<Balance?>(null)
    val balance: StateFlow<Balance?> = _balance.asStateFlow()

    private val _validators = MutableStateFlow<List<Validator>>(emptyList())
    val validators: StateFlow<List<Validator>> = _validators.asStateFlow()

    private val _delegations = MutableStateFlow<List<Delegation>>(emptyList())
    val delegations: StateFlow<List<Delegation>> = _delegations.asStateFlow()

    private val _rewards = MutableStateFlow<Rewards?>(null)
    val rewards: StateFlow<Rewards?> = _rewards.asStateFlow()

    private val _proposals = MutableStateFlow<List<Proposal>>(emptyList())
    val proposals: StateFlow<List<Proposal>> = _proposals.asStateFlow()

    private val _transactions = MutableStateFlow<List<Transaction>>(emptyList())
    val transactions: StateFlow<List<Transaction>> = _transactions.asStateFlow()

    // Multi-chain crypto transactions
    private val _cryptoTransactions = MutableStateFlow<Map<Chain, List<CryptoTransaction>>>(emptyMap())
    val cryptoTransactions: StateFlow<Map<Chain, List<CryptoTransaction>>> = _cryptoTransactions.asStateFlow()

    // USD prices for all chains
    private val _prices = MutableStateFlow<Map<Chain, Double>>(emptyMap())
    val prices: StateFlow<Map<Chain, Double>> = _prices.asStateFlow()

    // Total portfolio value in USD
    private val _totalPortfolioValue = MutableStateFlow(0.0)
    val totalPortfolioValue: StateFlow<Double> = _totalPortfolioValue.asStateFlow()

    // SECURITY: Mnemonic stored temporarily using SecureMnemonic wrapper
    // It is cleared from memory after backup confirmation
    private var pendingSecureMnemonic: SecureMnemonic? = null

    private val _generatedMnemonic = MutableStateFlow<String?>(null)
    val generatedMnemonic: StateFlow<String?> = _generatedMnemonic.asStateFlow()

    init {
        checkExistingWallet()
    }

    // MARK: - Wallet Management

    /**
     * Check if wallet exists on startup and load all chain accounts
     */
    private fun checkExistingWallet() {
        val address = keystoreService.getWalletAddress()
        if (address != null) {
            _walletAddress.value = address
            _uiState.value = _uiState.value.copy(hasWallet = true)

            // Load all chain accounts from stored mnemonic
            loadChainAccounts()
            refreshData()
        }
    }

    /**
     * Load all chain accounts from stored mnemonic
     */
    private fun loadChainAccounts() {
        viewModelScope.launch {
            val storedAccounts = keystoreService.getStoredChainAccounts()
            if (storedAccounts.isNotEmpty()) {
                _chainAccounts.value = storedAccounts
            }
        }
    }

    /**
     * Select a different chain to view
     */
    fun selectChain(chain: Chain) {
        _selectedChain.value = chain
        // Update current address to selected chain
        val account = _chainAccounts.value.find { it.chain == chain }
        if (account != null) {
            _walletAddress.value = account.address
        }
        // Refresh data for the selected chain
        refreshData()
    }

    /**
     * Get the account for currently selected chain
     */
    fun getCurrentChainAccount(): ChainAccount? {
        return _chainAccounts.value.find { it.chain == _selectedChain.value }
    }

    /**
     * Generate a new wallet
     * SECURITY: Private key is NOT stored in UI state
     */
    fun createWallet() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)

            try {
                // Generate mnemonic
                val mnemonic = cryptoService.generateMnemonic()

                // Store mnemonic securely (will be cleared after backup confirmation)
                pendingSecureMnemonic?.close() // Clear any existing
                pendingSecureMnemonic = SecureMnemonic(mnemonic)

                // Show mnemonic to user (this copy will be cleared when UI is dismissed)
                _generatedMnemonic.value = mnemonic

                // Derive address only for display (private key NOT stored)
                val secureKey = cryptoService.derivePrivateKeySecure(mnemonic)
                val address = try {
                    cryptoService.deriveAddress(secureKey.bytes)
                } finally {
                    secureKey.close() // Immediately clear private key
                }

                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    showMnemonicBackup = true,
                    pendingAddress = address
                )
            } catch (e: Exception) {
                pendingSecureMnemonic?.close()
                pendingSecureMnemonic = null
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = "Failed to create wallet: ${e.message}"
                )
            }
        }
    }

    /**
     * Confirm mnemonic backup and store wallet
     * SECURITY: Derives private key fresh, stores it, then clears from memory
     * MULTI-CHAIN: Derives and stores addresses for all supported chains
     */
    fun confirmMnemonicBackup(activity: FragmentActivity) {
        android.util.Log.d("WalletViewModel", "confirmMnemonicBackup called")

        viewModelScope.launch {
            val state = _uiState.value
            val address = state.pendingAddress
            val secureMnemonic = pendingSecureMnemonic

            android.util.Log.d("WalletViewModel", "address=$address, hasMnemonic=${secureMnemonic != null}")

            if (secureMnemonic == null || secureMnemonic.isCleared || address == null) {
                android.util.Log.e("WalletViewModel", "No pending wallet data")
                _uiState.value = _uiState.value.copy(error = "No pending wallet. Please try again.")
                return@launch
            }

            _uiState.value = _uiState.value.copy(isLoading = true, error = null)

            try {
                // Derive private key fresh from mnemonic
                android.util.Log.d("WalletViewModel", "Deriving private key...")
                val secureKey = cryptoService.derivePrivateKeySecure(secureMnemonic.phrase)

                try {
                    // Store the private key securely
                    android.util.Log.d("WalletViewModel", "Storing private key...")
                    val result = keystoreService.storePrivateKey(activity, secureKey.bytes)

                    if (result.isSuccess) {
                        android.util.Log.d("WalletViewModel", "Storage successful!")
                        keystoreService.storeWalletAddress(address)

                        // Store mnemonic for recovery phrase viewing
                        keystoreService.storeMnemonic(secureMnemonic.phrase)

                        // Derive all chain accounts from the mnemonic
                        android.util.Log.d("WalletViewModel", "Deriving multi-chain accounts...")
                        val allAccounts = cryptoService.deriveAllChainAccounts(secureMnemonic.phrase)
                        _chainAccounts.value = allAccounts

                        // Store chain accounts for later retrieval
                        keystoreService.storeChainAccounts(allAccounts)

                        _walletAddress.value = address
                        _generatedMnemonic.value = null

                        // Clear secure mnemonic
                        pendingSecureMnemonic?.close()
                        pendingSecureMnemonic = null

                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            hasWallet = true,
                            showMnemonicBackup = false,
                            pendingAddress = null
                        )
                        refreshData()
                    } else {
                        val error = result.exceptionOrNull()
                        android.util.Log.e("WalletViewModel", "Storage failed: ${error?.message}", error)
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            error = "Failed to store wallet: ${error?.message ?: "Unknown error"}"
                        )
                    }
                } finally {
                    // Always clear private key from memory
                    secureKey.close()
                }
            } catch (e: Exception) {
                android.util.Log.e("WalletViewModel", "Exception: ${e.message}", e)
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = "Failed to confirm wallet: ${e.message}"
                )
            }
        }
    }

    /**
     * Import existing wallet from mnemonic
     * SECURITY: Private key is derived, stored, and cleared in one operation
     */
    fun importWallet(activity: FragmentActivity, mnemonic: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)

            // Wrap mnemonic for secure handling
            val secureMnemonic = SecureMnemonic(mnemonic)

            try {
                // Validate mnemonic
                if (!cryptoService.validateMnemonic(secureMnemonic.phrase)) {
                    secureMnemonic.close()
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        error = "Invalid mnemonic phrase"
                    )
                    return@launch
                }

                // Derive key securely
                val secureKey = cryptoService.derivePrivateKeySecure(secureMnemonic.phrase)

                try {
                    val address = cryptoService.deriveAddress(secureKey.bytes)

                    // Store securely
                    keystoreService.storePrivateKey(activity, secureKey.bytes)
                        .onSuccess {
                            keystoreService.storeWalletAddress(address)
                            _walletAddress.value = address
                            _uiState.value = _uiState.value.copy(
                                isLoading = false,
                                hasWallet = true
                            )
                            refreshData()
                        }
                        .onFailure { e ->
                            _uiState.value = _uiState.value.copy(
                                isLoading = false,
                                error = "Failed to store wallet: ${e.message}"
                            )
                        }
                } finally {
                    // Always clear private key
                    secureKey.close()
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = "Failed to import wallet: ${e.message}"
                )
            } finally {
                // Always clear mnemonic
                secureMnemonic.close()
            }
        }
    }

    /**
     * Delete wallet and all data
     */
    fun deleteWallet() {
        keystoreService.deleteWallet()
        _walletAddress.value = null
        _chainAccounts.value = emptyList()
        _selectedChain.value = Chain.SHAREHODL
        _balance.value = null
        _delegations.value = emptyList()
        _rewards.value = null
        _transactions.value = emptyList()
        _uiState.value = WalletUiState()
    }

    // MARK: - Data Fetching

    /**
     * Refresh all wallet data (ShareHODL chain only)
     */
    fun refreshData() {
        val address = _walletAddress.value ?: return

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isRefreshing = true)

            // Fetch balance
            blockchainService.getBalance(address)
                .onSuccess { _balance.value = it }
                .onFailure { /* Silent fail */ }

            // Fetch delegations
            blockchainService.getDelegations(address)
                .onSuccess { _delegations.value = it }
                .onFailure { /* Silent fail */ }

            // Fetch rewards
            blockchainService.getRewards(address)
                .onSuccess { _rewards.value = it }
                .onFailure { /* Silent fail */ }

            // Fetch transactions
            blockchainService.getTransactions(address)
                .onSuccess { _transactions.value = it }
                .onFailure { /* Silent fail */ }

            _uiState.value = _uiState.value.copy(isRefreshing = false)
        }
    }

    /**
     * Refresh multi-chain balances (all chains)
     * This fetches real balances from each blockchain network
     */
    fun refreshMultiChainBalances() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isRefreshing = true)

            try {
                val currentAccounts = _chainAccounts.value
                if (currentAccounts.isEmpty()) {
                    _uiState.value = _uiState.value.copy(isRefreshing = false)
                    return@launch
                }

                // Fetch balances for all chains
                val updatedAccounts = multiChainApiService.fetchAllBalances(currentAccounts)

                // Fetch USD prices
                val prices = priceService.getUsdPrices(Chain.majorChains)
                _prices.value = prices

                // Update accounts with USD values
                val accountsWithUsd = updatedAccounts.map { account ->
                    val price = prices[account.chain]
                    if (price != null) {
                        val balance = account.balance.replace(",", "").toDoubleOrNull() ?: 0.0
                        val usdValue = balance * price
                        account.copy(balanceUsd = priceService.formatUsdValue(usdValue))
                    } else {
                        account
                    }
                }

                _chainAccounts.value = accountsWithUsd

                // Calculate total portfolio value
                val totalValue = accountsWithUsd.sumOf { account ->
                    val balance = account.balance.replace(",", "").toDoubleOrNull() ?: 0.0
                    val price = prices[account.chain] ?: 0.0
                    balance * price
                }
                _totalPortfolioValue.value = totalValue

            } catch (e: Exception) {
                android.util.Log.e("WalletViewModel", "Error refreshing multi-chain balances: ${e.message}")
            }

            _uiState.value = _uiState.value.copy(isRefreshing = false)
        }
    }

    /**
     * Refresh transactions for a specific chain
     */
    fun refreshChainTransactions(chain: Chain) {
        val account = _chainAccounts.value.find { it.chain == chain } ?: return

        viewModelScope.launch {
            multiChainApiService.fetchTransactions(account)
                .onSuccess { txs ->
                    val currentTxs = _cryptoTransactions.value.toMutableMap()
                    currentTxs[chain] = txs
                    _cryptoTransactions.value = currentTxs
                }
        }
    }

    /**
     * Get formatted total portfolio value
     */
    fun getFormattedPortfolioValue(): String {
        return priceService.formatUsdValue(_totalPortfolioValue.value)
    }

    /**
     * Fetch validators
     */
    fun fetchValidators() {
        viewModelScope.launch {
            blockchainService.getValidators()
                .onSuccess { _validators.value = it }
                .onFailure {
                    _uiState.value = _uiState.value.copy(
                        error = "Failed to load validators"
                    )
                }
        }
    }

    /**
     * Fetch governance proposals
     */
    fun fetchProposals() {
        viewModelScope.launch {
            blockchainService.getProposals()
                .onSuccess { _proposals.value = it }
                .onFailure {
                    _uiState.value = _uiState.value.copy(
                        error = "Failed to load proposals"
                    )
                }
        }
    }

    // MARK: - Staking Operations

    /**
     * Delegate tokens to a validator
     * Builds, signs, and broadcasts a real delegation transaction
     */
    fun delegate(
        activity: FragmentActivity,
        validatorAddress: String,
        amount: String,
        feeLevel: FeeLevel = FeeLevel.MEDIUM
    ) {
        val delegatorAddress = _walletAddress.value ?: return

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)

            try {
                // 1. Get account info
                val accountInfo = blockchainService.getAccount(delegatorAddress).getOrThrow()

                // 2. Get private key with biometric auth
                val privateKey = keystoreService.retrievePrivateKey(activity).getOrThrow()

                try {
                    // 3. Convert amount to micro units
                    val amountInMicro = (amount.toDouble() * 1_000_000).toLong().toString()

                    // 4. Calculate fee
                    val feeAmount = cosmosTransactionBuilder.calculateFee(feeLevel = feeLevel)

                    // 5. Build the delegation message
                    val msgDelegate = cosmosTransactionBuilder.buildMsgDelegate(
                        delegatorAddress = delegatorAddress,
                        validatorAddress = validatorAddress,
                        amount = amountInMicro
                    )

                    // 6. Build unsigned transaction
                    val unsignedTx = cosmosTransactionBuilder.buildUnsignedTx(
                        messages = listOf(msgDelegate),
                        memo = "",
                        feeAmount = feeAmount
                    )

                    // 7. Sign the transaction
                    val signedTxBytes = cosmosTransactionBuilder.signTransaction(
                        unsignedTx = unsignedTx,
                        accountInfo = accountInfo,
                        privateKey = privateKey
                    )

                    // 8. Broadcast the transaction
                    val result = blockchainService.broadcastTransaction(signedTxBytes).getOrThrow()

                    if (result.success) {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            successMessage = "Delegation submitted successfully!"
                        )
                        refreshData()
                    } else {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            error = "Delegation failed: ${result.rawLog}"
                        )
                    }

                } finally {
                    CryptoService.secureClear(privateKey)
                }

            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = "Failed to delegate: ${e.message}"
                )
            }
        }
    }

    /**
     * Undelegate from a validator
     * Note: Unbonding period is 21 days
     */
    fun undelegate(
        activity: FragmentActivity,
        validatorAddress: String,
        amount: String,
        feeLevel: FeeLevel = FeeLevel.MEDIUM
    ) {
        val delegatorAddress = _walletAddress.value ?: return

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)

            try {
                val accountInfo = blockchainService.getAccount(delegatorAddress).getOrThrow()
                val privateKey = keystoreService.retrievePrivateKey(activity).getOrThrow()

                try {
                    val amountInMicro = (amount.toDouble() * 1_000_000).toLong().toString()
                    val feeAmount = cosmosTransactionBuilder.calculateFee(feeLevel = feeLevel)

                    val msgUndelegate = cosmosTransactionBuilder.buildMsgUndelegate(
                        delegatorAddress = delegatorAddress,
                        validatorAddress = validatorAddress,
                        amount = amountInMicro
                    )

                    val unsignedTx = cosmosTransactionBuilder.buildUnsignedTx(
                        messages = listOf(msgUndelegate),
                        memo = "",
                        feeAmount = feeAmount
                    )

                    val signedTxBytes = cosmosTransactionBuilder.signTransaction(
                        unsignedTx = unsignedTx,
                        accountInfo = accountInfo,
                        privateKey = privateKey
                    )

                    val result = blockchainService.broadcastTransaction(signedTxBytes).getOrThrow()

                    if (result.success) {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            successMessage = "Undelegation started! Tokens will be available in 21 days."
                        )
                        refreshData()
                    } else {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            error = "Undelegation failed: ${result.rawLog}"
                        )
                    }

                } finally {
                    CryptoService.secureClear(privateKey)
                }

            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = "Failed to undelegate: ${e.message}"
                )
            }
        }
    }

    /**
     * Claim staking rewards from all validators
     */
    fun claimRewards(activity: FragmentActivity, feeLevel: FeeLevel = FeeLevel.MEDIUM) {
        val delegatorAddress = _walletAddress.value ?: return
        val delegations = _delegations.value

        if (delegations.isEmpty()) {
            _uiState.value = _uiState.value.copy(error = "No delegations to claim rewards from")
            return
        }

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)

            try {
                val accountInfo = blockchainService.getAccount(delegatorAddress).getOrThrow()
                val privateKey = keystoreService.retrievePrivateKey(activity).getOrThrow()

                try {
                    val feeAmount = cosmosTransactionBuilder.calculateFee(feeLevel = feeLevel)

                    // Build withdraw reward messages for all validators
                    val messages = delegations.map { delegation ->
                        cosmosTransactionBuilder.buildMsgWithdrawReward(
                            delegatorAddress = delegatorAddress,
                            validatorAddress = delegation.validatorAddress
                        )
                    }

                    val unsignedTx = cosmosTransactionBuilder.buildUnsignedTx(
                        messages = messages,
                        memo = "",
                        feeAmount = feeAmount,
                        gasLimit = NetworkConfig.ShareHODL.DEFAULT_GAS_LIMIT * messages.size
                    )

                    val signedTxBytes = cosmosTransactionBuilder.signTransaction(
                        unsignedTx = unsignedTx,
                        accountInfo = accountInfo,
                        privateKey = privateKey
                    )

                    val result = blockchainService.broadcastTransaction(signedTxBytes).getOrThrow()

                    if (result.success) {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            successMessage = "Rewards claimed successfully!"
                        )
                        refreshData()
                    } else {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            error = "Failed to claim rewards: ${result.rawLog}"
                        )
                    }

                } finally {
                    CryptoService.secureClear(privateKey)
                }

            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = "Failed to claim rewards: ${e.message}"
                )
            }
        }
    }

    // MARK: - Send

    /**
     * Send ShareHODL tokens to another address
     * Builds, signs, and broadcasts a real transaction
     */
    fun send(
        activity: FragmentActivity,
        toAddress: String,
        amount: String,
        feeLevel: FeeLevel = FeeLevel.MEDIUM
    ) {
        val fromAddress = _walletAddress.value ?: return

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)

            try {
                // 1. Get account info (sequence number, account number)
                val accountInfo = blockchainService.getAccount(fromAddress).getOrThrow()

                // 2. Get private key with biometric auth
                val privateKey = keystoreService.retrievePrivateKey(activity).getOrThrow()

                try {
                    // 3. Convert amount to micro units (uhodl)
                    val amountInMicro = (amount.toDouble() * 1_000_000).toLong().toString()

                    // 4. Calculate fee
                    val feeAmount = cosmosTransactionBuilder.calculateFee(
                        gasLimit = NetworkConfig.ShareHODL.DEFAULT_GAS_LIMIT,
                        feeLevel = feeLevel
                    )

                    // 5. Build the message
                    val msgSend = cosmosTransactionBuilder.buildMsgSend(
                        fromAddress = fromAddress,
                        toAddress = toAddress,
                        amount = amountInMicro
                    )

                    // 6. Build unsigned transaction
                    val unsignedTx = cosmosTransactionBuilder.buildUnsignedTx(
                        messages = listOf(msgSend),
                        memo = "",
                        feeAmount = feeAmount
                    )

                    // 7. Sign the transaction
                    val signedTxBytes = cosmosTransactionBuilder.signTransaction(
                        unsignedTx = unsignedTx,
                        accountInfo = accountInfo,
                        privateKey = privateKey
                    )

                    // 8. Broadcast the transaction
                    val result = blockchainService.broadcastTransaction(signedTxBytes).getOrThrow()

                    if (result.success) {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            successMessage = "Transaction submitted! Hash: ${result.txHash.take(12)}..."
                        )
                        refreshData()
                    } else {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            error = "Transaction failed: ${result.rawLog}"
                        )
                    }

                } finally {
                    // Always clear private key from memory
                    CryptoService.secureClear(privateKey)
                }

            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = "Failed to send: ${e.message}"
                )
            }
        }
    }

    /**
     * Estimate fee for a transaction
     */
    fun estimateFee(feeLevel: FeeLevel = FeeLevel.MEDIUM): String {
        val feeAmount = cosmosTransactionBuilder.calculateFee(
            gasLimit = NetworkConfig.ShareHODL.DEFAULT_GAS_LIMIT,
            feeLevel = feeLevel
        )
        val feeInHodl = feeAmount.toLong() / 1_000_000.0
        return String.format("%.6f HODL", feeInHodl)
    }

    // MARK: - Recovery Phrase

    private val _recoveryPhrase = MutableStateFlow<String?>(null)
    val recoveryPhrase: StateFlow<String?> = _recoveryPhrase.asStateFlow()

    /**
     * Retrieve recovery phrase for viewing (requires authentication)
     */
    fun viewRecoveryPhrase(activity: FragmentActivity) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)

            keystoreService.retrieveMnemonic(activity)
                .onSuccess { mnemonic ->
                    _recoveryPhrase.value = mnemonic
                    _uiState.value = _uiState.value.copy(isLoading = false)
                }
                .onFailure { e ->
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        error = "Failed to retrieve recovery phrase: ${e.message}"
                    )
                }
        }
    }

    fun clearRecoveryPhrase() {
        _recoveryPhrase.value = null
    }

    // MARK: - Utilities

    fun clearError() {
        _uiState.value = _uiState.value.copy(error = null)
    }

    fun clearSuccessMessage() {
        _uiState.value = _uiState.value.copy(successMessage = null)
    }

    fun cancelMnemonicBackup() {
        // Clear secure mnemonic from memory
        pendingSecureMnemonic?.close()
        pendingSecureMnemonic = null

        _generatedMnemonic.value = null
        _uiState.value = _uiState.value.copy(
            showMnemonicBackup = false,
            pendingAddress = null
        )
    }

    /**
     * Clean up resources when ViewModel is cleared
     */
    override fun onCleared() {
        super.onCleared()
        // Ensure any pending sensitive data is cleared
        pendingSecureMnemonic?.close()
        pendingSecureMnemonic = null
    }

    fun isBiometricsAvailable(): Boolean = keystoreService.isBiometricsAvailable()
}

/**
 * UI State for wallet screens
 * SECURITY: No sensitive data (private keys, seeds) is stored here
 */
data class WalletUiState(
    val hasWallet: Boolean = false,
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val error: String? = null,
    val successMessage: String? = null,
    val showMnemonicBackup: Boolean = false,
    val pendingAddress: String? = null
)
