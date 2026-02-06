import Foundation
import SwiftUI

/// Main wallet manager - handles wallet state and operations
@MainActor
class WalletManager: ObservableObject {
    // MARK: - Published State

    @Published var address: String?
    @Published var balances: [Balance] = []
    @Published var delegations: [Delegation] = []
    @Published var totalRewards: String = "0"
    @Published var validators: [Validator] = []

    @Published var isLoading = false
    @Published var error: String?
    @Published var isConnected = false

    // MARK: - Services

    private let keychainService = KeychainService.shared
    private let cryptoService = CryptoService.shared
    private let blockchainService = BlockchainService.shared

    // MARK: - Computed Properties

    var hasWallet: Bool {
        keychainService.hasWallet
    }

    var formattedAddress: String {
        guard let address = address else { return "" }
        if address.count > 20 {
            return "\(address.prefix(12))...\(address.suffix(8))"
        }
        return address
    }

    var totalBalance: String {
        let total = balances.reduce(0.0) { sum, balance in
            sum + (Double(balance.amount) ?? 0) / 1_000_000
        }
        return String(format: "%.2f", total)
    }

    // MARK: - Initialization

    init() {
        // Load saved wallet address
        if let savedAddress = keychainService.retrieveWalletAddress() {
            self.address = savedAddress
        }
    }

    // MARK: - Wallet Creation

    /// Create a new wallet
    func createWallet() async throws -> String {
        isLoading = true
        defer { isLoading = false }

        // Generate mnemonic
        let mnemonic = try cryptoService.generateMnemonic()

        // Derive keys and address
        let privateKey = try cryptoService.derivePrivateKey(from: mnemonic)
        let address = try cryptoService.deriveAddress(from: privateKey)

        // Store securely
        try keychainService.storePrivateKey(privateKey)
        try keychainService.storeWalletAddress(address)

        self.address = address

        // Return mnemonic for user to back up (only shown once!)
        return mnemonic
    }

    /// Import wallet from mnemonic
    func importWallet(mnemonic: String) async throws {
        isLoading = true
        defer { isLoading = false }

        // Validate mnemonic
        guard cryptoService.validateMnemonic(mnemonic) else {
            throw WalletError.invalidMnemonic
        }

        // Derive keys and address
        let privateKey = try cryptoService.derivePrivateKey(from: mnemonic)
        let address = try cryptoService.deriveAddress(from: privateKey)

        // Store securely
        try keychainService.storePrivateKey(privateKey)
        try keychainService.storeWalletAddress(address)

        self.address = address

        // Fetch initial data
        await refreshData()
    }

    // MARK: - Data Fetching

    /// Refresh all wallet data
    func refreshData() async {
        guard let address = address else { return }

        isLoading = true
        error = nil

        do {
            // Check connection
            isConnected = await blockchainService.checkConnection()

            // Fetch balances
            balances = try await blockchainService.fetchBalances(address: address)

            // Fetch delegations
            delegations = try await blockchainService.fetchDelegations(address: address)

            // Fetch rewards
            let rewards = try await blockchainService.fetchRewards(address: address)
            totalRewards = rewards.totalRewards

            // Fetch validators
            validators = try await blockchainService.fetchValidators()

        } catch {
            self.error = error.localizedDescription
        }

        isLoading = false
    }

    // MARK: - Transactions

    /// Send tokens
    func sendTokens(to recipient: String, amount: String, denom: String) async throws {
        guard let _ = address else {
            throw WalletError.noWallet
        }

        isLoading = true
        defer { isLoading = false }

        // Get private key (requires biometric auth)
        let privateKey = try keychainService.retrievePrivateKey()

        // Build, sign, and broadcast transaction
        // TODO: Implement proper Cosmos transaction building
        // This requires amino/protobuf encoding

        // For now, just show it would work
        print("Would send \(amount) \(denom) to \(recipient)")

        // Clear private key from memory
        // (In Swift, this is handled by ARC, but we could zero the bytes)

        await refreshData()
    }

    /// Delegate tokens to validator
    func delegate(to validator: String, amount: String) async throws {
        guard let _ = address else {
            throw WalletError.noWallet
        }

        isLoading = true
        defer { isLoading = false }

        // Similar to send - build MsgDelegate, sign, broadcast
        print("Would delegate \(amount) to \(validator)")

        await refreshData()
    }

    /// Claim staking rewards
    func claimRewards(from validator: String? = nil) async throws {
        guard let _ = address else {
            throw WalletError.noWallet
        }

        isLoading = true
        defer { isLoading = false }

        // Build MsgWithdrawDelegatorReward, sign, broadcast
        if let validator = validator {
            print("Would claim rewards from \(validator)")
        } else {
            print("Would claim all rewards")
        }

        await refreshData()
    }

    // MARK: - Wallet Management

    /// Delete wallet (reset app)
    func deleteWallet() throws {
        try keychainService.deleteWallet()
        address = nil
        balances = []
        delegations = []
        totalRewards = "0"
        validators = []
    }
}

// MARK: - Errors

enum WalletError: LocalizedError {
    case noWallet
    case invalidMnemonic
    case transactionFailed(String)

    var errorDescription: String? {
        switch self {
        case .noWallet:
            return "No wallet configured"
        case .invalidMnemonic:
            return "Invalid mnemonic phrase"
        case .transactionFailed(let reason):
            return "Transaction failed: \(reason)"
        }
    }
}
