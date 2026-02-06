import Foundation
import Security

/// Service for verifying user has backed up their recovery phrase
/// SECURITY: Uses Keychain for storing backup verification status
final class BackupVerificationService {
    static let shared = BackupVerificationService()

    private let keychainService = "com.sharehodl.backupverification"
    private let backupVerifiedKey = "backup_verified"
    private let backupVerifiedDateKey = "backup_verified_date"

    private init() {}

    // MARK: - Keychain Helpers

    private func setKeychainValue(_ value: String, forKey key: String) -> Bool {
        let data = value.data(using: .utf8)!

        // First try to delete any existing item
        let deleteQuery: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key
        ]
        SecItemDelete(deleteQuery as CFDictionary)

        // Add new item
        let addQuery: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly
        ]

        let status = SecItemAdd(addQuery as CFDictionary, nil)
        return status == errSecSuccess
    }

    private func getKeychainValue(forKey key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne
        ]

        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess,
              let data = result as? Data,
              let value = String(data: data, encoding: .utf8) else {
            return nil
        }

        return value
    }

    private func deleteKeychainValue(forKey key: String) {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: keychainService,
            kSecAttrAccount as String: key
        ]
        SecItemDelete(query as CFDictionary)
    }

    // MARK: - Backup Status

    /// Check if backup has been verified
    var isBackupVerified: Bool {
        return getKeychainValue(forKey: backupVerifiedKey) == "true"
    }

    /// Date when backup was verified
    var backupVerifiedDate: Date? {
        guard let timestampString = getKeychainValue(forKey: backupVerifiedDateKey),
              let timestamp = Double(timestampString) else {
            return nil
        }
        return Date(timeIntervalSince1970: timestamp)
    }

    /// Mark backup as verified
    func markBackupVerified() {
        _ = setKeychainValue("true", forKey: backupVerifiedKey)
        _ = setKeychainValue(String(Date().timeIntervalSince1970), forKey: backupVerifiedDateKey)
    }

    /// Reset backup verification (for testing or wallet reset)
    func resetBackupVerification() {
        deleteKeychainValue(forKey: backupVerifiedKey)
        deleteKeychainValue(forKey: backupVerifiedDateKey)
    }

    // MARK: - Verification Challenge

    /// Generate a verification challenge with random word positions
    /// - Parameters:
    ///   - mnemonic: The full recovery phrase
    ///   - challengeCount: Number of words to verify (default 3)
    /// - Returns: Array of word challenges
    func generateChallenge(mnemonic: String, challengeCount: Int = 3) -> [WordChallenge] {
        let words = mnemonic.split(separator: " ").map(String.init)
        guard words.count >= challengeCount else { return [] }

        // Generate random unique positions
        var positions = Set<Int>()
        while positions.count < challengeCount {
            let randomPosition = Int.random(in: 0..<words.count)
            positions.insert(randomPosition)
        }

        // Create challenges sorted by position
        return positions.sorted().map { position in
            WordChallenge(
                wordNumber: position + 1,  // 1-indexed for user display
                correctWord: words[position],
                userAnswer: ""
            )
        }
    }

    /// Verify user answers against the challenge
    /// - Parameters:
    ///   - challenges: The word challenges with user answers
    /// - Returns: True if all answers are correct
    func verifyChallenge(_ challenges: [WordChallenge]) -> Bool {
        return challenges.allSatisfy { challenge in
            challenge.userAnswer.lowercased().trimmingCharacters(in: .whitespaces) ==
            challenge.correctWord.lowercased()
        }
    }

    /// Get word suggestions for autocomplete during verification
    func suggestWords(prefix: String) -> [String] {
        return CryptoService.shared.suggestWords(prefix: prefix)
    }
}

/// Represents a single word verification challenge
struct WordChallenge: Identifiable {
    let id = UUID()
    let wordNumber: Int      // 1-indexed position (e.g., "Word #5")
    let correctWord: String  // The correct answer
    var userAnswer: String   // User's input

    /// Check if user's answer is correct
    var isCorrect: Bool {
        userAnswer.lowercased().trimmingCharacters(in: .whitespaces) ==
        correctWord.lowercased()
    }

    /// Display string for the challenge
    var displayText: String {
        "Word #\(wordNumber)"
    }
}
