import Foundation

/// BIP39 English wordlist (2048 words)
/// SECURITY: This is a critical security component. The app will crash if
/// the wordlist cannot be loaded, as using a reduced wordlist would
/// severely compromise wallet security.
enum BIP39Wordlist {
    /// Required word count for BIP39 English wordlist
    static let requiredWordCount = 2048

    /// The complete BIP39 English wordlist
    /// SECURITY: Validated at load time to ensure all 2048 words are present
    static let words: [String] = {
        let loadedWords = loadWordlist()

        // SECURITY: Crash if wordlist is incomplete
        // Using a reduced wordlist would make wallets trivially brute-forceable
        guard loadedWords.count == requiredWordCount else {
            fatalError("""
                SECURITY ERROR: BIP39 wordlist contains \(loadedWords.count) words but requires \(requiredWordCount).
                The app cannot safely generate wallets without the complete wordlist.
                Ensure bip39-english.txt is included in the app bundle.
                """)
        }

        // Validate first and last words as a checksum
        guard loadedWords.first == "abandon" && loadedWords.last == "zoo" else {
            fatalError("""
                SECURITY ERROR: BIP39 wordlist failed validation.
                Expected first word 'abandon' and last word 'zoo'.
                The wordlist file may be corrupted or tampered with.
                """)
        }

        return loadedWords
    }()

    /// Load the wordlist from the bundled resource file
    /// SECURITY: No fallback - fails if wordlist cannot be loaded
    private static func loadWordlist() -> [String] {
        // Try to load from bundle resource
        if let url = Bundle.main.url(forResource: "bip39-english", withExtension: "txt"),
           let content = try? String(contentsOf: url, encoding: .utf8) {
            let words = content.components(separatedBy: .newlines).filter { !$0.isEmpty }
            if words.count == requiredWordCount {
                return words
            }
        }

        // Fallback: try to load from module bundle (for SPM)
        #if SWIFT_PACKAGE
        if let url = Bundle.module.url(forResource: "bip39-english", withExtension: "txt"),
           let content = try? String(contentsOf: url, encoding: .utf8) {
            let words = content.components(separatedBy: .newlines).filter { !$0.isEmpty }
            if words.count == requiredWordCount {
                return words
            }
        }
        #endif

        // SECURITY: No fallback - return empty to trigger fatal error above
        // Using a reduced wordlist would make brute-forcing trivial
        return []
    }

    /// Check if the wordlist is properly loaded (for testing/validation)
    static var isValid: Bool {
        return words.count == requiredWordCount
    }
}
