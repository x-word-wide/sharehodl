import Foundation
import Security
import LocalAuthentication

/// Service for secure storage of sensitive data in iOS Keychain
final class KeychainService {
    static let shared = KeychainService()

    private let service = "com.sharehodl.wallet"

    private init() {}

    // MARK: - Keys
    private enum KeychainKey: String {
        case encryptedPrivateKey = "encrypted_private_key"
        case walletAddress = "wallet_address"
        case hasWallet = "has_wallet"
    }

    // MARK: - Public Methods

    /// Store encrypted private key in Keychain with biometric protection
    /// Uses biometryCurrentSet OR deviceOwnerAuthentication for fallback to passcode
    func storePrivateKey(_ privateKey: Data, withBiometrics: Bool = true) throws {
        let accessControl: SecAccessControl?

        if withBiometrics {
            var error: Unmanaged<CFError>?
            // Use userPresence which allows both biometrics and passcode fallback
            accessControl = SecAccessControlCreateWithFlags(
                nil,
                kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
                .userPresence,
                &error
            )
            if let error = error?.takeRetainedValue() {
                throw KeychainError.accessControlCreationFailed(error)
            }
        } else {
            accessControl = nil
        }

        var query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: KeychainKey.encryptedPrivateKey.rawValue,
            kSecValueData as String: privateKey
        ]

        if let ac = accessControl {
            query[kSecAttrAccessControl as String] = ac
        }

        // Delete existing item if any
        SecItemDelete(query as CFDictionary)

        let status = SecItemAdd(query as CFDictionary, nil)
        guard status == errSecSuccess else {
            throw KeychainError.unableToStore(status)
        }
    }

    /// Retrieve private key from Keychain (requires biometric auth)
    func retrievePrivateKey() throws -> Data {
        let context = LAContext()
        context.localizedReason = "Access your wallet"

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: KeychainKey.encryptedPrivateKey.rawValue,
            kSecReturnData as String: true,
            kSecUseAuthenticationContext as String: context
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess, let data = result as? Data else {
            throw KeychainError.unableToRetrieve(status)
        }

        return data
    }

    /// Store wallet address (not sensitive, no biometrics needed)
    /// Uses proper accessibility so address is only available when device unlocked
    func storeWalletAddress(_ address: String) throws {
        guard let addressData = address.data(using: .utf8) else {
            throw KeychainError.invalidData
        }

        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: KeychainKey.walletAddress.rawValue,
            kSecValueData as String: addressData,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]

        SecItemDelete(query as CFDictionary)

        let status = SecItemAdd(query as CFDictionary, nil)
        guard status == errSecSuccess else {
            throw KeychainError.unableToStore(status)
        }
    }

    /// Retrieve wallet address
    func retrieveWalletAddress() -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: KeychainKey.walletAddress.rawValue,
            kSecReturnData as String: true
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess, let data = result as? Data else {
            return nil
        }

        return String(data: data, encoding: .utf8)
    }

    /// Check if wallet exists
    var hasWallet: Bool {
        return retrieveWalletAddress() != nil
    }

    /// Delete all wallet data
    func deleteWallet() throws {
        let keys: [KeychainKey] = [.encryptedPrivateKey, .walletAddress, .hasWallet]

        for key in keys {
            let query: [String: Any] = [
                kSecClass as String: kSecClassGenericPassword,
                kSecAttrService as String: service,
                kSecAttrAccount as String: key.rawValue
            ]
            SecItemDelete(query as CFDictionary)
        }
    }

    // MARK: - Biometric Check

    /// Check if device supports biometrics
    var isBiometricsAvailable: Bool {
        let context = LAContext()
        var error: NSError?
        return context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error)
    }

    /// Get biometric type (Face ID or Touch ID)
    var biometricType: LABiometryType {
        let context = LAContext()
        _ = context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: nil)
        return context.biometryType
    }
}

// MARK: - Errors

enum KeychainError: LocalizedError {
    case unableToStore(OSStatus)
    case unableToRetrieve(OSStatus)
    case accessControlCreationFailed(Error)
    case invalidData
    case authenticationCancelled
    case biometricsNotAvailable

    /// User-facing error message (does not expose internal details)
    var errorDescription: String? {
        switch self {
        case .unableToStore:
            return "Unable to save wallet data securely"
        case .unableToRetrieve:
            return "Unable to access wallet. Please try again."
        case .accessControlCreationFailed:
            return "Security configuration failed"
        case .invalidData:
            return "Invalid wallet data"
        case .authenticationCancelled:
            return "Authentication was cancelled"
        case .biometricsNotAvailable:
            return "Biometric authentication is not available"
        }
    }

    /// Debug description with full details (for logging only)
    var debugDescription: String {
        switch self {
        case .unableToStore(let status):
            return "Keychain store failed: OSStatus \(status)"
        case .unableToRetrieve(let status):
            return "Keychain retrieve failed: OSStatus \(status)"
        case .accessControlCreationFailed(let error):
            return "Access control creation failed: \(error.localizedDescription)"
        case .invalidData:
            return "Invalid data format for keychain storage"
        case .authenticationCancelled:
            return "User cancelled authentication"
        case .biometricsNotAvailable:
            return "Biometrics not enrolled or not available"
        }
    }
}
