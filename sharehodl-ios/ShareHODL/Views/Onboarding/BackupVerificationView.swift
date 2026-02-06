import SwiftUI

/// View for verifying user has backed up their recovery phrase
struct BackupVerificationView: View {
    @Environment(\.dismiss) var dismiss
    @Binding var isVerified: Bool

    let mnemonic: String

    @State private var currentStep: VerificationStep = .showPhrase
    @State private var challenges: [WordChallenge] = []
    @State private var currentChallengeIndex = 0
    @State private var userInput = ""
    @State private var showError = false
    @State private var suggestions: [String] = []

    private let verificationService = BackupVerificationService.shared

    enum VerificationStep {
        case showPhrase
        case verifyWords
        case complete
    }

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Progress indicator
                progressIndicator

                // Content based on current step
                switch currentStep {
                case .showPhrase:
                    showPhraseView
                case .verifyWords:
                    verifyWordsView
                case .complete:
                    completionView
                }
            }
            .navigationTitle(navigationTitle)
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    if currentStep != .complete {
                        Button("Cancel") {
                            dismiss()
                        }
                    }
                }
            }
        }
        .interactiveDismissDisabled(currentStep == .verifyWords)
    }

    // MARK: - Navigation Title

    private var navigationTitle: String {
        switch currentStep {
        case .showPhrase: return "Backup Recovery Phrase"
        case .verifyWords: return "Verify Backup"
        case .complete: return "Backup Complete"
        }
    }

    // MARK: - Progress Indicator

    private var progressIndicator: some View {
        HStack(spacing: 8) {
            ForEach(0..<3, id: \.self) { index in
                Rectangle()
                    .fill(stepColor(for: index))
                    .frame(height: 4)
                    .clipShape(Capsule())
            }
        }
        .padding(.horizontal)
        .padding(.top, 8)
    }

    private func stepColor(for index: Int) -> Color {
        let currentIndex: Int
        switch currentStep {
        case .showPhrase: currentIndex = 0
        case .verifyWords: currentIndex = 1
        case .complete: currentIndex = 2
        }
        return index <= currentIndex ? .blue : .gray.opacity(0.3)
    }

    // MARK: - Show Phrase View

    private var showPhraseView: some View {
        ScrollView {
            VStack(spacing: 24) {
                // Warning
                warningBanner

                // Instructions
                Text("Write down these 24 words in order and store them in a safe place. This is the ONLY way to recover your wallet.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal)

                // Word grid
                wordGrid

                // Security tips
                securityTips

                // Continue button
                Button {
                    challenges = verificationService.generateChallenge(mnemonic: mnemonic, challengeCount: 3)
                    currentChallengeIndex = 0
                    userInput = ""
                    withAnimation {
                        currentStep = .verifyWords
                    }
                } label: {
                    Text("I've Written It Down")
                        .font(.headline)
                        .foregroundStyle(.white)
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(.blue)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                }
                .padding(.horizontal)
            }
            .padding(.vertical, 24)
        }
    }

    private var warningBanner: some View {
        HStack(spacing: 12) {
            Image(systemName: "exclamationmark.triangle.fill")
                .foregroundStyle(.orange)
                .font(.title2)

            VStack(alignment: .leading, spacing: 4) {
                Text("Never share your recovery phrase")
                    .font(.subheadline.bold())
                Text("Anyone with this phrase can steal your funds")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(.orange.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .padding(.horizontal)
    }

    private var wordGrid: some View {
        let words = mnemonic.split(separator: " ").map(String.init)
        let columns = [GridItem(.flexible()), GridItem(.flexible()), GridItem(.flexible())]

        return LazyVGrid(columns: columns, spacing: 12) {
            ForEach(Array(words.enumerated()), id: \.offset) { index, word in
                HStack(spacing: 8) {
                    Text("\(index + 1)")
                        .font(.caption2)
                        .foregroundStyle(.secondary)
                        .frame(width: 20)

                    Text(word)
                        .font(.system(.body, design: .monospaced))
                        .fontWeight(.medium)
                }
                .padding(.vertical, 8)
                .padding(.horizontal, 12)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(.gray.opacity(0.1))
                .clipShape(RoundedRectangle(cornerRadius: 8))
            }
        }
        .padding(.horizontal)
    }

    private var securityTips: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Security Tips")
                .font(.headline)

            SecurityTipRow(icon: "pencil", text: "Write it on paper, not digitally")
            SecurityTipRow(icon: "lock.fill", text: "Store in a secure location")
            SecurityTipRow(icon: "eye.slash", text: "Never share with anyone")
            SecurityTipRow(icon: "camera.fill", text: "Don't take screenshots")
        }
        .padding()
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(.gray.opacity(0.05))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .padding(.horizontal)
    }

    // MARK: - Verify Words View

    private var verifyWordsView: some View {
        VStack(spacing: 32) {
            Spacer()

            // Challenge progress
            Text("Verify word \(currentChallengeIndex + 1) of \(challenges.count)")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            // Current challenge
            if currentChallengeIndex < challenges.count {
                let challenge = challenges[currentChallengeIndex]

                VStack(spacing: 16) {
                    Text("What is \(challenge.displayText)?")
                        .font(.title2.bold())

                    // Input field
                    TextField("Enter word", text: $userInput)
                        .font(.system(.title3, design: .monospaced))
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()
                        .padding()
                        .background(.gray.opacity(0.1))
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                        .padding(.horizontal, 40)
                        .onChange(of: userInput) { newValue in
                            showError = false
                            suggestions = verificationService.suggestWords(prefix: newValue)
                        }

                    // Suggestions
                    if !suggestions.isEmpty && !userInput.isEmpty {
                        ScrollView(.horizontal, showsIndicators: false) {
                            HStack(spacing: 8) {
                                ForEach(suggestions, id: \.self) { word in
                                    Button {
                                        userInput = word
                                        suggestions = []
                                    } label: {
                                        Text(word)
                                            .font(.subheadline)
                                            .padding(.horizontal, 12)
                                            .padding(.vertical, 6)
                                            .background(.blue.opacity(0.1))
                                            .foregroundStyle(.blue)
                                            .clipShape(Capsule())
                                    }
                                }
                            }
                            .padding(.horizontal, 40)
                        }
                    }

                    // Error message
                    if showError {
                        Text("Incorrect word. Please try again.")
                            .font(.caption)
                            .foregroundStyle(.red)
                    }
                }
            }

            Spacer()

            // Verify button
            Button {
                verifyCurrentWord()
            } label: {
                Text("Verify")
                    .font(.headline)
                    .foregroundStyle(.white)
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(userInput.isEmpty ? .gray : .blue)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
            }
            .disabled(userInput.isEmpty)
            .padding(.horizontal)
            .padding(.bottom, 24)
        }
    }

    private func verifyCurrentWord() {
        guard currentChallengeIndex < challenges.count else { return }

        let challenge = challenges[currentChallengeIndex]
        let isCorrect = userInput.lowercased().trimmingCharacters(in: .whitespaces) ==
                        challenge.correctWord.lowercased()

        if isCorrect {
            // Move to next challenge or complete
            if currentChallengeIndex < challenges.count - 1 {
                currentChallengeIndex += 1
                userInput = ""
                suggestions = []
            } else {
                // All challenges completed
                verificationService.markBackupVerified()
                withAnimation {
                    currentStep = .complete
                }
            }
        } else {
            showError = true
        }
    }

    // MARK: - Completion View

    private var completionView: some View {
        VStack(spacing: 32) {
            Spacer()

            Image(systemName: "checkmark.circle.fill")
                .font(.system(size: 80))
                .foregroundStyle(.green)

            VStack(spacing: 8) {
                Text("Backup Verified!")
                    .font(.title.bold())

                Text("Your recovery phrase has been verified. Keep it safe - it's the only way to recover your wallet.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, 40)
            }

            Spacer()

            Button {
                isVerified = true
                dismiss()
            } label: {
                Text("Done")
                    .font(.headline)
                    .foregroundStyle(.white)
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(.green)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
            }
            .padding(.horizontal)
            .padding(.bottom, 24)
        }
    }
}

// MARK: - Supporting Views

struct SecurityTipRow: View {
    let icon: String
    let text: String

    var body: some View {
        HStack(spacing: 12) {
            Image(systemName: icon)
                .foregroundStyle(.blue)
                .frame(width: 24)

            Text(text)
                .font(.subheadline)
        }
    }
}

#Preview {
    BackupVerificationView(
        isVerified: .constant(false),
        mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"
    )
}
