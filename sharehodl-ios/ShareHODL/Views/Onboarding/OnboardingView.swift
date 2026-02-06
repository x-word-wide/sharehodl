import SwiftUI

struct OnboardingView: View {
    @EnvironmentObject var walletManager: WalletManager
    @State private var showCreateWallet = false
    @State private var showImportWallet = false

    var body: some View {
        NavigationStack {
            VStack(spacing: 32) {
                Spacer()

                // Logo
                VStack(spacing: 16) {
                    Image(systemName: "wallet.pass.fill")
                        .font(.system(size: 80))
                        .foregroundStyle(.blue)

                    Text("ShareHODL")
                        .font(.largeTitle.bold())

                    Text("Your Gateway to Tokenized Equity")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                }

                Spacer()

                // Features
                VStack(alignment: .leading, spacing: 16) {
                    FeatureRow(
                        icon: "lock.shield",
                        title: "Secure Storage",
                        description: "Keys protected by Face ID & Keychain"
                    )

                    FeatureRow(
                        icon: "bolt.fill",
                        title: "Instant Trading",
                        description: "2-second settlement on all trades"
                    )

                    FeatureRow(
                        icon: "chart.line.uptrend.xyaxis",
                        title: "Earn Rewards",
                        description: "Stake tokens and earn ~12.5% APR"
                    )
                }
                .padding(.horizontal)

                Spacer()

                // Action Buttons
                VStack(spacing: 12) {
                    Button {
                        showCreateWallet = true
                    } label: {
                        Text("Create New Wallet")
                            .font(.headline)
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(.blue)
                            .foregroundStyle(.white)
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                    }

                    Button {
                        showImportWallet = true
                    } label: {
                        Text("Import Existing Wallet")
                            .font(.headline)
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(.ultraThinMaterial)
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                    }
                }
                .padding(.horizontal)
                .padding(.bottom, 32)
            }
            .navigationDestination(isPresented: $showCreateWallet) {
                CreateWalletView()
            }
            .navigationDestination(isPresented: $showImportWallet) {
                ImportWalletView()
            }
        }
    }
}

struct FeatureRow: View {
    let icon: String
    let title: String
    let description: String

    var body: some View {
        HStack(spacing: 16) {
            Image(systemName: icon)
                .font(.title2)
                .foregroundStyle(.blue)
                .frame(width: 40)

            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.headline)
                Text(description)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()
        }
    }
}

// MARK: - Create Wallet View

struct CreateWalletView: View {
    @EnvironmentObject var walletManager: WalletManager
    @State private var mnemonic: String?
    @State private var isCreating = false
    @State private var hasBackedUp = false
    @State private var error: String?

    var body: some View {
        VStack(spacing: 24) {
            if let mnemonic = mnemonic {
                // Show mnemonic for backup
                mnemonicBackupView(mnemonic)
            } else {
                // Create wallet prompt
                createPromptView
            }
        }
        .padding()
        .navigationTitle("Create Wallet")
        .navigationBarTitleDisplayMode(.inline)
    }

    private var createPromptView: some View {
        VStack(spacing: 24) {
            Spacer()

            Image(systemName: "key.fill")
                .font(.system(size: 60))
                .foregroundStyle(.blue)

            Text("Generate Recovery Phrase")
                .font(.title2.bold())

            Text("You'll receive a 12-word recovery phrase. Write it down and store it safely. This is the ONLY way to recover your wallet.")
                .multilineTextAlignment(.center)
                .foregroundStyle(.secondary)

            Spacer()

            if let error = error {
                Text(error)
                    .foregroundStyle(.red)
                    .font(.caption)
            }

            Button {
                createWallet()
            } label: {
                if isCreating {
                    ProgressView()
                        .frame(maxWidth: .infinity)
                        .padding()
                } else {
                    Text("Generate Phrase")
                        .font(.headline)
                        .frame(maxWidth: .infinity)
                        .padding()
                }
            }
            .background(.blue)
            .foregroundStyle(.white)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .disabled(isCreating)
        }
    }

    private func mnemonicBackupView(_ mnemonic: String) -> some View {
        VStack(spacing: 24) {
            // Warning
            HStack {
                Image(systemName: "exclamationmark.triangle.fill")
                    .foregroundStyle(.orange)
                Text("Write down these words in order")
                    .font(.headline)
            }
            .padding()
            .background(.orange.opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 8))

            // Mnemonic words
            let words = mnemonic.split(separator: " ").map(String.init)
            LazyVGrid(columns: [
                GridItem(.flexible()),
                GridItem(.flexible()),
                GridItem(.flexible())
            ], spacing: 8) {
                ForEach(Array(words.enumerated()), id: \.offset) { index, word in
                    HStack {
                        Text("\(index + 1).")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                            .frame(width: 24, alignment: .leading)
                        Text(word)
                            .font(.body.monospaced())
                    }
                    .padding(8)
                    .background(.ultraThinMaterial)
                    .clipShape(RoundedRectangle(cornerRadius: 6))
                }
            }

            Spacer()

            // Confirmation
            Toggle("I have written down my recovery phrase", isOn: $hasBackedUp)

            Button {
                // Wallet already created, just dismiss
            } label: {
                Text("Continue")
                    .font(.headline)
                    .frame(maxWidth: .infinity)
                    .padding()
            }
            .background(hasBackedUp ? .blue : .gray)
            .foregroundStyle(.white)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .disabled(!hasBackedUp)
        }
    }

    private func createWallet() {
        isCreating = true
        error = nil

        Task {
            do {
                let newMnemonic = try await walletManager.createWallet()
                await MainActor.run {
                    mnemonic = newMnemonic
                    isCreating = false
                }
            } catch {
                await MainActor.run {
                    self.error = error.localizedDescription
                    isCreating = false
                }
            }
        }
    }
}

// MARK: - Import Wallet View

struct ImportWalletView: View {
    @EnvironmentObject var walletManager: WalletManager
    @State private var mnemonic = ""
    @State private var isImporting = false
    @State private var error: String?

    var body: some View {
        VStack(spacing: 24) {
            Text("Enter your 12, 15, 18, 21, or 24-word recovery phrase")
                .font(.headline)

            TextEditor(text: $mnemonic)
                .font(.body.monospaced())
                .frame(height: 150)
                .padding(8)
                .background(.ultraThinMaterial)
                .clipShape(RoundedRectangle(cornerRadius: 12))
                .autocapitalization(.none)
                .autocorrectionDisabled()

            Text("Enter words separated by spaces")
                .font(.caption)
                .foregroundStyle(.secondary)

            if let error = error {
                Text(error)
                    .foregroundStyle(.red)
                    .font(.caption)
            }

            Spacer()

            Button {
                importWallet()
            } label: {
                if isImporting {
                    ProgressView()
                        .frame(maxWidth: .infinity)
                        .padding()
                } else {
                    Text("Import Wallet")
                        .font(.headline)
                        .frame(maxWidth: .infinity)
                        .padding()
                }
            }
            .background(mnemonic.isEmpty ? .gray : .blue)
            .foregroundStyle(.white)
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .disabled(mnemonic.isEmpty || isImporting)
        }
        .padding()
        .navigationTitle("Import Wallet")
        .navigationBarTitleDisplayMode(.inline)
    }

    private func importWallet() {
        isImporting = true
        error = nil

        Task {
            do {
                try await walletManager.importWallet(mnemonic: mnemonic.lowercased().trimmingCharacters(in: .whitespacesAndNewlines))
            } catch {
                await MainActor.run {
                    self.error = error.localizedDescription
                    isImporting = false
                }
            }
        }
    }
}

#Preview {
    OnboardingView()
        .environmentObject(WalletManager())
}
