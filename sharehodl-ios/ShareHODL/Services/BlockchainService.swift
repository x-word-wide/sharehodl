import Foundation

/// Service for interacting with ShareHODL blockchain
/// SECURITY: Enforces HTTPS for all production API calls
actor BlockchainService {
    static let shared = BlockchainService()

    // Chain configuration
    private let chainId = "sharehodl-1"
    private var restUrl = "http://localhost:1317"
    private var rpcUrl = "http://localhost:26657"

    // Security configuration
    private var enforceHTTPS = true
    private var isProduction = false

    private init() {}

    // MARK: - Configuration

    /// Configure blockchain endpoints
    /// - Parameters:
    ///   - restUrl: REST API URL
    ///   - rpcUrl: RPC URL
    ///   - isProduction: If true, enforces HTTPS (default: true)
    func configure(restUrl: String, rpcUrl: String, isProduction: Bool = true) throws {
        // Validate URLs
        guard let restURL = URL(string: restUrl),
              let rpcURL = URL(string: rpcUrl) else {
            throw BlockchainError.invalidConfiguration("Invalid URL format")
        }

        // Enforce HTTPS in production
        if isProduction {
            guard restURL.scheme == "https" && rpcURL.scheme == "https" else {
                throw BlockchainError.invalidConfiguration("HTTPS required for production endpoints")
            }
        }

        // Check for localhost (development only)
        let isLocalhost = restURL.host == "localhost" ||
                         restURL.host == "127.0.0.1" ||
                         rpcURL.host == "localhost" ||
                         rpcURL.host == "127.0.0.1"

        if isProduction && isLocalhost {
            throw BlockchainError.invalidConfiguration("Localhost not allowed in production")
        }

        self.restUrl = restUrl
        self.rpcUrl = rpcUrl
        self.isProduction = isProduction
        self.enforceHTTPS = isProduction
    }

    /// Configure for local development (HTTP allowed)
    func configureForDevelopment(restUrl: String = "http://localhost:1317", rpcUrl: String = "http://localhost:26657") {
        self.restUrl = restUrl
        self.rpcUrl = rpcUrl
        self.isProduction = false
        self.enforceHTTPS = false
    }

    /// Validate URL before making request
    private func validateURL(_ urlString: String) throws -> URL {
        guard let url = URL(string: urlString) else {
            throw BlockchainError.invalidConfiguration("Invalid URL")
        }

        if enforceHTTPS && url.scheme != "https" {
            throw BlockchainError.invalidConfiguration("HTTPS required")
        }

        return url
    }

    // MARK: - Account Queries

    /// Fetch account balances
    func fetchBalances(address: String) async throws -> [Balance] {
        let url = URL(string: "\(restUrl)/cosmos/bank/v1beta1/balances/\(address)")!

        let (data, response) = try await URLSession.shared.data(from: url)

        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw BlockchainError.requestFailed
        }

        let result = try JSONDecoder().decode(BalancesResponse.self, from: data)
        return result.balances
    }

    /// Fetch account info (for sequence number)
    func fetchAccount(address: String) async throws -> AccountInfo {
        let url = URL(string: "\(restUrl)/cosmos/auth/v1beta1/accounts/\(address)")!

        let (data, response) = try await URLSession.shared.data(from: url)

        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw BlockchainError.requestFailed
        }

        let result = try JSONDecoder().decode(AccountResponse.self, from: data)
        return result.account
    }

    // MARK: - Staking Queries

    /// Fetch all validators
    func fetchValidators() async throws -> [Validator] {
        let url = URL(string: "\(restUrl)/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED&pagination.limit=100")!

        let (data, response) = try await URLSession.shared.data(from: url)

        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw BlockchainError.requestFailed
        }

        let result = try JSONDecoder().decode(ValidatorsResponse.self, from: data)
        return result.validators
    }

    /// Fetch user delegations
    func fetchDelegations(address: String) async throws -> [Delegation] {
        let url = URL(string: "\(restUrl)/cosmos/staking/v1beta1/delegations/\(address)")!

        let (data, response) = try await URLSession.shared.data(from: url)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw BlockchainError.requestFailed
        }

        // 404 means no delegations
        if httpResponse.statusCode == 404 {
            return []
        }

        guard httpResponse.statusCode == 200 else {
            throw BlockchainError.requestFailed
        }

        let result = try JSONDecoder().decode(DelegationsResponse.self, from: data)
        return result.delegation_responses.map { $0.delegation }
    }

    /// Fetch staking rewards
    func fetchRewards(address: String) async throws -> RewardsResponse {
        let url = URL(string: "\(restUrl)/cosmos/distribution/v1beta1/delegators/\(address)/rewards")!

        let (data, response) = try await URLSession.shared.data(from: url)

        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw BlockchainError.requestFailed
        }

        return try JSONDecoder().decode(RewardsResponse.self, from: data)
    }

    // MARK: - Transaction Broadcasting

    /// Broadcast a signed transaction
    func broadcastTransaction(_ signedTx: Data) async throws -> BroadcastResult {
        let url = URL(string: "\(restUrl)/cosmos/tx/v1beta1/txs")!

        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        let body: [String: Any] = [
            "tx_bytes": signedTx.base64EncodedString(),
            "mode": "BROADCAST_MODE_SYNC"
        ]
        request.httpBody = try JSONSerialization.data(withJSONObject: body)

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw BlockchainError.broadcastFailed
        }

        return try JSONDecoder().decode(BroadcastResult.self, from: data)
    }

    // MARK: - Transaction History

    /// Fetch transaction history for an address
    func fetchTransactions(address: String, limit: Int = 20) async throws -> [Transaction] {
        // Query for sent transactions
        let sentUrl = URL(string: "\(restUrl)/cosmos/tx/v1beta1/txs?events=message.sender='\(address)'&pagination.limit=\(limit)")!

        let (sentData, sentResponse) = try await URLSession.shared.data(from: sentUrl)

        guard let httpResponse = sentResponse as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw BlockchainError.requestFailed
        }

        let sentResult = try JSONDecoder().decode(TransactionsResponse.self, from: sentData)

        // Query for received transactions
        let receivedUrl = URL(string: "\(restUrl)/cosmos/tx/v1beta1/txs?events=transfer.recipient='\(address)'&pagination.limit=\(limit)")!

        let (receivedData, _) = try await URLSession.shared.data(from: receivedUrl)
        let receivedResult = try JSONDecoder().decode(TransactionsResponse.self, from: receivedData)

        // Combine and sort by height
        var allTxs = sentResult.tx_responses + receivedResult.tx_responses
        allTxs.sort { ($0.height ?? "0") > ($1.height ?? "0") }

        return Array(allTxs.prefix(limit))
    }

    // MARK: - Network Status

    /// Check if blockchain is reachable
    func checkConnection() async -> Bool {
        guard let url = URL(string: "\(rpcUrl)/status") else {
            return false
        }

        do {
            let (_, response) = try await URLSession.shared.data(from: url)
            return (response as? HTTPURLResponse)?.statusCode == 200
        } catch {
            return false
        }
    }

    /// Get latest block height
    func getLatestBlockHeight() async throws -> Int64 {
        let url = URL(string: "\(rpcUrl)/status")!

        let (data, response) = try await URLSession.shared.data(from: url)

        guard let httpResponse = response as? HTTPURLResponse,
              httpResponse.statusCode == 200 else {
            throw BlockchainError.requestFailed
        }

        let result = try JSONDecoder().decode(StatusResponse.self, from: data)
        return Int64(result.result.sync_info.latest_block_height) ?? 0
    }
}

// MARK: - Response Types

struct BalancesResponse: Decodable {
    let balances: [Balance]
}

struct Balance: Decodable, Identifiable {
    var id: String { denom }
    let denom: String
    let amount: String

    var displayAmount: String {
        let value = (Double(amount) ?? 0) / 1_000_000
        return String(format: "%.2f", value)
    }

    var symbol: String {
        denom.replacingOccurrences(of: "u", with: "").uppercased()
    }
}

struct AccountResponse: Decodable {
    let account: AccountInfo
}

struct AccountInfo: Decodable {
    let address: String?
    let account_number: String?
    let sequence: String?

    var sequenceNumber: UInt64 {
        UInt64(sequence ?? "0") ?? 0
    }

    var accountNumber: UInt64 {
        UInt64(account_number ?? "0") ?? 0
    }
}

struct ValidatorsResponse: Decodable {
    let validators: [Validator]
}

struct Validator: Decodable, Identifiable {
    var id: String { operator_address }
    let operator_address: String
    let description: ValidatorDescription?
    let tokens: String?
    let commission: ValidatorCommission?

    var moniker: String {
        description?.moniker ?? "Unknown"
    }

    var commissionRate: String {
        guard let rate = commission?.commission_rates?.rate,
              let rateValue = Double(rate) else {
            return "0%"
        }
        return String(format: "%.1f%%", rateValue * 100)
    }
}

struct ValidatorDescription: Decodable {
    let moniker: String?
    let website: String?
    let details: String?
}

struct ValidatorCommission: Decodable {
    let commission_rates: CommissionRates?
}

struct CommissionRates: Decodable {
    let rate: String?
    let max_rate: String?
}

struct DelegationsResponse: Decodable {
    let delegation_responses: [DelegationResponse]
}

struct DelegationResponse: Decodable {
    let delegation: Delegation
    let balance: Balance?
}

struct Delegation: Decodable, Identifiable {
    var id: String { "\(delegator_address)-\(validator_address)" }
    let delegator_address: String
    let validator_address: String
    let shares: String?
}

struct RewardsResponse: Decodable {
    let rewards: [ValidatorReward]?
    let total: [Balance]?

    var totalRewards: String {
        guard let total = total, let first = total.first else {
            return "0"
        }
        return first.displayAmount
    }
}

struct ValidatorReward: Decodable {
    let validator_address: String
    let reward: [Balance]?
}

struct TransactionsResponse: Decodable {
    let tx_responses: [Transaction]
}

struct Transaction: Decodable, Identifiable {
    var id: String { txhash ?? UUID().uuidString }
    let txhash: String?
    let height: String?
    let timestamp: String?
    let code: Int?

    var isSuccess: Bool {
        code == 0
    }
}

struct BroadcastResult: Decodable {
    let tx_response: TxResponse?
}

struct TxResponse: Decodable {
    let txhash: String?
    let code: Int?
    let raw_log: String?

    var isSuccess: Bool {
        code == 0
    }
}

struct StatusResponse: Decodable {
    let result: StatusResult
}

struct StatusResult: Decodable {
    let sync_info: SyncInfo
}

struct SyncInfo: Decodable {
    let latest_block_height: String
}

// MARK: - Errors

enum BlockchainError: LocalizedError {
    case requestFailed
    case broadcastFailed
    case invalidResponse
    case invalidConfiguration(String)

    var errorDescription: String? {
        switch self {
        case .requestFailed:
            return "Network request failed"
        case .broadcastFailed:
            return "Transaction broadcast failed"
        case .invalidResponse:
            return "Invalid response from server"
        case .invalidConfiguration(let message):
            return "Configuration error: \(message)"
        }
    }
}
