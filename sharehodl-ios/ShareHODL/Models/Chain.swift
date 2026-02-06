import Foundation

/// Supported blockchain networks
enum Chain: String, CaseIterable, Identifiable, Codable {
    case sharehodl = "sharehodl"
    case bitcoin = "bitcoin"
    case litecoin = "litecoin"
    case ethereum = "ethereum"

    var id: String { rawValue }

    /// Display name
    var name: String {
        switch self {
        case .sharehodl: return "ShareHODL"
        case .bitcoin: return "Bitcoin"
        case .litecoin: return "Litecoin"
        case .ethereum: return "Ethereum"
        }
    }

    /// Token symbol
    var symbol: String {
        switch self {
        case .sharehodl: return "HODL"
        case .bitcoin: return "BTC"
        case .litecoin: return "LTC"
        case .ethereum: return "ETH"
        }
    }

    /// BIP44 coin type for key derivation
    /// https://github.com/satoshilabs/slips/blob/master/slip-0044.md
    var coinType: UInt32 {
        switch self {
        case .sharehodl: return 118  // Cosmos
        case .bitcoin: return 0
        case .litecoin: return 2
        case .ethereum: return 60
        }
    }

    /// Address prefix for Bech32 encoding (Cosmos-style chains)
    var bech32Prefix: String? {
        switch self {
        case .sharehodl: return "hodl"
        case .bitcoin: return "bc"      // Native SegWit
        case .litecoin: return "ltc"    // Native SegWit
        case .ethereum: return nil      // Ethereum uses hex addresses
        }
    }

    /// Whether this chain uses Cosmos-style addresses
    var isCosmosStyle: Bool {
        switch self {
        case .sharehodl: return true
        default: return false
        }
    }

    /// Whether this chain uses Bitcoin-style SegWit addresses
    var isBitcoinStyle: Bool {
        switch self {
        case .bitcoin, .litecoin: return true
        default: return false
        }
    }

    /// Whether this chain uses Ethereum-style addresses
    var isEthereumStyle: Bool {
        switch self {
        case .ethereum: return true
        default: return false
        }
    }

    /// Number of decimal places for display
    var decimals: Int {
        switch self {
        case .sharehodl: return 6   // uhodl
        case .bitcoin: return 8     // satoshis
        case .litecoin: return 8    // litoshis
        case .ethereum: return 18   // wei
        }
    }

    /// Smallest unit name
    var smallestUnit: String {
        switch self {
        case .sharehodl: return "uhodl"
        case .bitcoin: return "satoshi"
        case .litecoin: return "litoshi"
        case .ethereum: return "wei"
        }
    }

    /// Chain icon (SF Symbol name)
    var iconName: String {
        switch self {
        case .sharehodl: return "h.circle.fill"
        case .bitcoin: return "bitcoinsign.circle.fill"
        case .litecoin: return "l.circle.fill"
        case .ethereum: return "e.circle.fill"
        }
    }

    /// Chain color for UI
    var colorHex: String {
        switch self {
        case .sharehodl: return "#6366F1"  // Indigo
        case .bitcoin: return "#F7931A"    // Bitcoin orange
        case .litecoin: return "#345D9D"   // Litecoin blue
        case .ethereum: return "#627EEA"   // Ethereum purple
        }
    }

    /// Explorer URL template (replace {address} or {tx})
    var explorerAddressURL: String {
        switch self {
        case .sharehodl: return "https://explorer.sharehodl.com/address/{address}"
        case .bitcoin: return "https://mempool.space/address/{address}"
        case .litecoin: return "https://blockchair.com/litecoin/address/{address}"
        case .ethereum: return "https://etherscan.io/address/{address}"
        }
    }

    var explorerTxURL: String {
        switch self {
        case .sharehodl: return "https://explorer.sharehodl.com/tx/{tx}"
        case .bitcoin: return "https://mempool.space/tx/{tx}"
        case .litecoin: return "https://blockchair.com/litecoin/transaction/{tx}"
        case .ethereum: return "https://etherscan.io/tx/{tx}"
        }
    }
}

/// Represents a wallet account on a specific chain
struct ChainAccount: Identifiable, Codable {
    let id: UUID
    let chain: Chain
    let address: String
    let derivationPath: String
    var balance: String
    var balanceUSD: String?

    init(chain: Chain, address: String, derivationPath: String, balance: String = "0", balanceUSD: String? = nil) {
        self.id = UUID()
        self.chain = chain
        self.address = address
        self.derivationPath = derivationPath
        self.balance = balance
        self.balanceUSD = balanceUSD
    }

    /// Formatted balance with symbol
    var formattedBalance: String {
        "\(balance) \(chain.symbol)"
    }

    /// Short address for display (first 8...last 6)
    var shortAddress: String {
        guard address.count > 16 else { return address }
        let prefix = String(address.prefix(10))
        let suffix = String(address.suffix(6))
        return "\(prefix)...\(suffix)"
    }

    /// Explorer URL for this address
    var explorerURL: URL? {
        let urlString = chain.explorerAddressURL.replacingOccurrences(of: "{address}", with: address)
        return URL(string: urlString)
    }
}
