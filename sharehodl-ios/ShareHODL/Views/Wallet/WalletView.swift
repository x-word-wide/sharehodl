import SwiftUI

struct WalletView: View {
    @EnvironmentObject var walletManager: WalletManager
    @State private var showingSendSheet = false
    @State private var showingReceiveSheet = false

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 20) {
                    // Portfolio Card
                    portfolioCard

                    // Quick Actions
                    quickActions

                    // Assets List
                    assetsSection
                }
                .padding()
            }
            .navigationTitle("Wallet")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        Task {
                            await walletManager.refreshData()
                        }
                    } label: {
                        Image(systemName: "arrow.clockwise")
                    }
                    .disabled(walletManager.isLoading)
                }
            }
            .refreshable {
                await walletManager.refreshData()
            }
            .sheet(isPresented: $showingSendSheet) {
                SendView()
            }
            .sheet(isPresented: $showingReceiveSheet) {
                ReceiveView()
            }
        }
    }

    // MARK: - Portfolio Card

    private var portfolioCard: some View {
        VStack(spacing: 16) {
            // Connection status
            HStack {
                Circle()
                    .fill(walletManager.isConnected ? .green : .red)
                    .frame(width: 8, height: 8)
                Text(walletManager.isConnected ? "Connected" : "Offline")
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Spacer()
            }

            // Total balance
            VStack(spacing: 4) {
                Text("Total Balance")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                Text("$\(walletManager.totalBalance)")
                    .font(.system(size: 44, weight: .bold))
            }

            // Address
            if let address = walletManager.address {
                HStack {
                    Text(walletManager.formattedAddress)
                        .font(.system(.caption, design: .monospaced))
                        .foregroundStyle(.secondary)

                    Button {
                        UIPasteboard.general.string = address
                    } label: {
                        Image(systemName: "doc.on.doc")
                            .font(.caption)
                    }
                }
            }
        }
        .padding(24)
        .frame(maxWidth: .infinity)
        .background(
            LinearGradient(
                colors: [.blue.opacity(0.3), .purple.opacity(0.3)],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
        )
        .clipShape(RoundedRectangle(cornerRadius: 20))
    }

    // MARK: - Quick Actions

    private var quickActions: some View {
        HStack(spacing: 16) {
            ActionButton(
                title: "Send",
                icon: "arrow.up.circle.fill",
                color: .blue
            ) {
                showingSendSheet = true
            }

            ActionButton(
                title: "Receive",
                icon: "arrow.down.circle.fill",
                color: .green
            ) {
                showingReceiveSheet = true
            }

            ActionButton(
                title: "Stake",
                icon: "chart.bar.fill",
                color: .purple
            ) {
                // Navigate to staking tab
            }
        }
    }

    // MARK: - Assets Section

    private var assetsSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Assets")
                .font(.headline)

            if walletManager.balances.isEmpty {
                emptyAssetsView
            } else {
                ForEach(walletManager.balances) { balance in
                    AssetRow(balance: balance)
                }
            }
        }
    }

    private var emptyAssetsView: some View {
        VStack(spacing: 8) {
            Image(systemName: "tray")
                .font(.largeTitle)
                .foregroundStyle(.secondary)
            Text("No assets yet")
                .foregroundStyle(.secondary)
            Text("Receive tokens to get started")
                .font(.caption)
                .foregroundStyle(.tertiary)
        }
        .frame(maxWidth: .infinity)
        .padding(32)
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - Supporting Views

struct ActionButton: View {
    let title: String
    let icon: String
    let color: Color
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: 8) {
                Image(systemName: icon)
                    .font(.title2)
                Text(title)
                    .font(.caption)
            }
            .foregroundStyle(color)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 16)
            .background(color.opacity(0.1))
            .clipShape(RoundedRectangle(cornerRadius: 12))
        }
    }
}

struct AssetRow: View {
    let balance: Balance

    var body: some View {
        HStack {
            // Token icon
            Circle()
                .fill(balance.symbol == "HODL" ? .blue : .gray)
                .frame(width: 44, height: 44)
                .overlay {
                    Text(String(balance.symbol.prefix(1)))
                        .font(.headline)
                        .foregroundStyle(.white)
                }

            // Token info
            VStack(alignment: .leading, spacing: 2) {
                Text(balance.symbol)
                    .font(.headline)
                Text(balance.denom)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            // Amount
            VStack(alignment: .trailing, spacing: 2) {
                Text(balance.displayAmount)
                    .font(.headline)
                Text("$\(balance.displayAmount)")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .padding(12)
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

// MARK: - Send/Receive Views (Stubs)

struct SendView: View {
    @Environment(\.dismiss) var dismiss
    @State private var recipient = ""
    @State private var amount = ""

    var body: some View {
        NavigationStack {
            Form {
                Section("Recipient") {
                    TextField("hodl1...", text: $recipient)
                        .font(.system(.body, design: .monospaced))
                }

                Section("Amount") {
                    TextField("0.00", text: $amount)
                        .keyboardType(.decimalPad)
                }

                Section {
                    Button("Send") {
                        // TODO: Implement
                    }
                    .disabled(recipient.isEmpty || amount.isEmpty)
                }
            }
            .navigationTitle("Send")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .topBarLeading) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
            }
        }
    }
}

struct ReceiveView: View {
    @Environment(\.dismiss) var dismiss
    @EnvironmentObject var walletManager: WalletManager

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                // QR Code placeholder
                RoundedRectangle(cornerRadius: 12)
                    .fill(.ultraThinMaterial)
                    .frame(width: 200, height: 200)
                    .overlay {
                        Image(systemName: "qrcode")
                            .font(.system(size: 100))
                            .foregroundStyle(.secondary)
                    }

                // Address
                VStack(spacing: 8) {
                    Text("Your Address")
                        .font(.headline)

                    Text(walletManager.address ?? "")
                        .font(.system(.caption, design: .monospaced))
                        .multilineTextAlignment(.center)
                        .padding()
                        .background(.ultraThinMaterial)
                        .clipShape(RoundedRectangle(cornerRadius: 8))

                    Button {
                        UIPasteboard.general.string = walletManager.address
                    } label: {
                        Label("Copy Address", systemImage: "doc.on.doc")
                    }
                    .buttonStyle(.bordered)
                }

                Spacer()
            }
            .padding()
            .navigationTitle("Receive")
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
    WalletView()
        .environmentObject(WalletManager())
}
