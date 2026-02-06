import SwiftUI

struct StakingView: View {
    @EnvironmentObject var walletManager: WalletManager
    @State private var selectedValidator: Validator?
    @State private var delegateAmount = ""
    @State private var showDelegateSheet = false

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 20) {
                    // Staking Overview
                    stakingOverview

                    // Rewards Section
                    if Double(walletManager.totalRewards) ?? 0 > 0 {
                        rewardsCard
                    }

                    // My Delegations
                    delegationsSection

                    // Validators List
                    validatorsSection
                }
                .padding()
            }
            .navigationTitle("Staking")
            .refreshable {
                await walletManager.refreshData()
            }
            .sheet(isPresented: $showDelegateSheet) {
                DelegateSheet(validator: selectedValidator, amount: $delegateAmount)
            }
        }
    }

    // MARK: - Overview

    private var stakingOverview: some View {
        HStack(spacing: 16) {
            StatCard(
                title: "Total Staked",
                value: "0.00",
                subtitle: "STAKE",
                color: .green
            )

            StatCard(
                title: "Rewards",
                value: walletManager.totalRewards,
                subtitle: "STAKE",
                color: .yellow
            )

            StatCard(
                title: "APR",
                value: "~12.5%",
                subtitle: "Annual",
                color: .blue
            )
        }
    }

    private var rewardsCard: some View {
        HStack {
            VStack(alignment: .leading) {
                Text("Pending Rewards")
                    .font(.headline)
                Text("\(walletManager.totalRewards) STAKE")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Button("Claim All") {
                Task {
                    try? await walletManager.claimRewards()
                }
            }
            .buttonStyle(.borderedProminent)
            .tint(.yellow)
        }
        .padding()
        .background(.yellow.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }

    // MARK: - Delegations

    private var delegationsSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("My Delegations")
                .font(.headline)

            if walletManager.delegations.isEmpty {
                Text("No active delegations")
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(.ultraThinMaterial)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
            } else {
                ForEach(walletManager.delegations) { delegation in
                    DelegationRow(delegation: delegation)
                }
            }
        }
    }

    // MARK: - Validators

    private var validatorsSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Validators")
                    .font(.headline)
                Spacer()
                Text("\(walletManager.validators.count)")
                    .foregroundStyle(.secondary)
            }

            if walletManager.validators.isEmpty {
                Text("Loading validators...")
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity)
                    .padding()
            } else {
                ForEach(walletManager.validators.prefix(10)) { validator in
                    ValidatorRow(validator: validator) {
                        selectedValidator = validator
                        showDelegateSheet = true
                    }
                }
            }
        }
    }
}

// MARK: - Supporting Views

struct StatCard: View {
    let title: String
    let value: String
    let subtitle: String
    let color: Color

    var body: some View {
        VStack(spacing: 4) {
            Text(title)
                .font(.caption)
                .foregroundStyle(.secondary)
            Text(value)
                .font(.title3.bold())
            Text(subtitle)
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity)
        .padding()
        .background(color.opacity(0.1))
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

struct DelegationRow: View {
    let delegation: Delegation

    var body: some View {
        HStack {
            Circle()
                .fill(.blue.opacity(0.2))
                .frame(width: 40, height: 40)
                .overlay {
                    Text("V")
                        .font(.headline)
                        .foregroundStyle(.blue)
                }

            VStack(alignment: .leading) {
                Text(delegation.validator_address.prefix(20) + "...")
                    .font(.subheadline)
                Text(delegation.shares ?? "0")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Button("Manage") {
                // Show undelegate/redelegate options
            }
            .font(.caption)
            .buttonStyle(.bordered)
        }
        .padding()
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

struct ValidatorRow: View {
    let validator: Validator
    let onDelegate: () -> Void

    var body: some View {
        HStack {
            Circle()
                .fill(.blue.opacity(0.2))
                .frame(width: 40, height: 40)
                .overlay {
                    Text(String(validator.moniker.prefix(1)))
                        .font(.headline)
                        .foregroundStyle(.blue)
                }

            VStack(alignment: .leading) {
                Text(validator.moniker)
                    .font(.subheadline)
                Text(validator.commissionRate + " commission")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Button("Delegate") {
                onDelegate()
            }
            .font(.caption)
            .buttonStyle(.borderedProminent)
        }
        .padding()
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

struct DelegateSheet: View {
    let validator: Validator?
    @Binding var amount: String
    @Environment(\.dismiss) var dismiss

    var body: some View {
        NavigationStack {
            Form {
                if let validator = validator {
                    Section("Validator") {
                        Text(validator.moniker)
                        Text("Commission: \(validator.commissionRate)")
                            .foregroundStyle(.secondary)
                    }
                }

                Section("Amount") {
                    TextField("0.00", text: $amount)
                        .keyboardType(.decimalPad)
                }

                Section {
                    Text("Unbonding Period: 21 days")
                        .foregroundStyle(.secondary)
                }

                Section {
                    Button("Delegate") {
                        // TODO: Implement
                        dismiss()
                    }
                    .disabled(amount.isEmpty)
                }
            }
            .navigationTitle("Delegate")
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

#Preview {
    StakingView()
        .environmentObject(WalletManager())
}
