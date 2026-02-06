import SwiftUI

struct SettingsView: View {
    @EnvironmentObject var walletManager: WalletManager
    @State private var showDeleteConfirmation = false
    @State private var useBiometrics = true
    @State private var showBackupPhrase = false

    var body: some View {
        NavigationStack {
            List {
                // Wallet Section
                Section("Wallet") {
                    HStack {
                        Text("Address")
                        Spacer()
                        Text(walletManager.formattedAddress)
                            .font(.system(.caption, design: .monospaced))
                            .foregroundStyle(.secondary)
                    }

                    Button {
                        showBackupPhrase = true
                    } label: {
                        Label("Backup Recovery Phrase", systemImage: "key")
                    }
                }

                // Security Section
                Section("Security") {
                    Toggle(isOn: $useBiometrics) {
                        Label("Use Face ID", systemImage: "faceid")
                    }

                    NavigationLink {
                        Text("Change PIN")
                    } label: {
                        Label("Change PIN", systemImage: "lock")
                    }
                }

                // Network Section
                Section("Network") {
                    HStack {
                        Text("Chain")
                        Spacer()
                        Text("sharehodl-1")
                            .foregroundStyle(.secondary)
                    }

                    HStack {
                        Text("Status")
                        Spacer()
                        HStack(spacing: 4) {
                            Circle()
                                .fill(walletManager.isConnected ? .green : .red)
                                .frame(width: 8, height: 8)
                            Text(walletManager.isConnected ? "Connected" : "Offline")
                                .foregroundStyle(.secondary)
                        }
                    }
                }

                // About Section
                Section("About") {
                    HStack {
                        Text("Version")
                        Spacer()
                        Text("1.0.0")
                            .foregroundStyle(.secondary)
                    }

                    Link(destination: URL(string: "https://sharehodl.com")!) {
                        Label("Website", systemImage: "globe")
                    }

                    Link(destination: URL(string: "https://x.com/share_hodl")!) {
                        Label("Twitter", systemImage: "bird")
                    }
                }

                // Danger Zone
                Section {
                    Button(role: .destructive) {
                        showDeleteConfirmation = true
                    } label: {
                        Label("Delete Wallet", systemImage: "trash")
                    }
                } footer: {
                    Text("This will remove all wallet data from this device. Make sure you have backed up your recovery phrase.")
                }
            }
            .navigationTitle("Settings")
            .alert("Delete Wallet?", isPresented: $showDeleteConfirmation) {
                Button("Cancel", role: .cancel) {}
                Button("Delete", role: .destructive) {
                    try? walletManager.deleteWallet()
                }
            } message: {
                Text("This action cannot be undone. Make sure you have your recovery phrase backed up.")
            }
            .sheet(isPresented: $showBackupPhrase) {
                BackupPhraseView()
            }
        }
    }
}

struct BackupPhraseView: View {
    @Environment(\.dismiss) var dismiss
    @State private var isAuthenticated = false

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                if isAuthenticated {
                    // Would show actual mnemonic after biometric auth
                    Text("Recovery phrase would be shown here after biometric authentication")
                        .multilineTextAlignment(.center)
                        .foregroundStyle(.secondary)
                } else {
                    VStack(spacing: 16) {
                        Image(systemName: "faceid")
                            .font(.system(size: 60))
                            .foregroundStyle(.blue)

                        Text("Authenticate to view your recovery phrase")
                            .multilineTextAlignment(.center)

                        Button("Authenticate") {
                            // Trigger biometric auth
                            isAuthenticated = true
                        }
                        .buttonStyle(.borderedProminent)
                    }
                }

                Spacer()
            }
            .padding()
            .navigationTitle("Recovery Phrase")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
    }
}

#Preview {
    SettingsView()
        .environmentObject(WalletManager())
}
