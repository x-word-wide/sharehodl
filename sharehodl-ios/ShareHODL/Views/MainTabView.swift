import SwiftUI

struct MainTabView: View {
    @EnvironmentObject var walletManager: WalletManager

    var body: some View {
        TabView {
            WalletView()
                .tabItem {
                    Label("Wallet", systemImage: "wallet.pass")
                }

            StakingView()
                .tabItem {
                    Label("Staking", systemImage: "chart.bar.fill")
                }

            TradingView()
                .tabItem {
                    Label("Trade", systemImage: "arrow.left.arrow.right")
                }

            GovernanceView()
                .tabItem {
                    Label("Governance", systemImage: "building.columns")
                }

            SettingsView()
                .tabItem {
                    Label("Settings", systemImage: "gear")
                }
        }
        .tint(.blue)
        .task {
            await walletManager.refreshData()
        }
    }
}

#Preview {
    MainTabView()
        .environmentObject(WalletManager())
}
