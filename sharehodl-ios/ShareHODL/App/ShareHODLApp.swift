import SwiftUI

@main
struct ShareHODLApp: App {
    @StateObject private var walletManager = WalletManager()

    var body: some Scene {
        WindowGroup {
            if walletManager.hasWallet {
                MainTabView()
                    .environmentObject(walletManager)
            } else {
                OnboardingView()
                    .environmentObject(walletManager)
            }
        }
    }
}
