import SwiftUI

struct GovernanceView: View {
    @State private var selectedTab = "active"

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Tab selector
                Picker("Proposals", selection: $selectedTab) {
                    Text("Active").tag("active")
                    Text("Passed").tag("passed")
                    Text("Rejected").tag("rejected")
                }
                .pickerStyle(.segmented)
                .padding()

                // Proposals list
                ScrollView {
                    VStack(spacing: 12) {
                        // Sample proposals
                        ProposalCard(
                            id: 1,
                            title: "Increase Validator Set Size",
                            status: .voting,
                            yesPercent: 65,
                            endDate: "2 days left"
                        )

                        ProposalCard(
                            id: 2,
                            title: "Community Pool Spend",
                            status: .voting,
                            yesPercent: 45,
                            endDate: "5 days left"
                        )

                        ProposalCard(
                            id: 3,
                            title: "Parameter Change: Min Commission",
                            status: .passed,
                            yesPercent: 78,
                            endDate: "Ended Jan 15"
                        )
                    }
                    .padding()
                }
            }
            .navigationTitle("Governance")
        }
    }
}

struct ProposalCard: View {
    let id: Int
    let title: String
    let status: ProposalStatus
    let yesPercent: Int
    let endDate: String

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("#\(id)")
                    .font(.caption)
                    .foregroundStyle(.secondary)

                Spacer()

                Text(status.rawValue)
                    .font(.caption)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(status.color.opacity(0.2))
                    .foregroundStyle(status.color)
                    .clipShape(Capsule())
            }

            Text(title)
                .font(.headline)

            // Voting progress
            VStack(alignment: .leading, spacing: 4) {
                HStack {
                    Text("Yes: \(yesPercent)%")
                        .font(.caption)
                        .foregroundStyle(.green)
                    Spacer()
                    Text("No: \(100 - yesPercent)%")
                        .font(.caption)
                        .foregroundStyle(.red)
                }

                GeometryReader { geo in
                    ZStack(alignment: .leading) {
                        RoundedRectangle(cornerRadius: 4)
                            .fill(.red.opacity(0.3))
                            .frame(height: 8)

                        RoundedRectangle(cornerRadius: 4)
                            .fill(.green)
                            .frame(width: geo.size.width * CGFloat(yesPercent) / 100, height: 8)
                    }
                }
                .frame(height: 8)
            }

            HStack {
                Text(endDate)
                    .font(.caption)
                    .foregroundStyle(.secondary)

                Spacer()

                if status == .voting {
                    Button("Vote") {
                        // Show voting sheet
                    }
                    .buttonStyle(.borderedProminent)
                    .controlSize(.small)
                }
            }
        }
        .padding()
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12))
    }
}

enum ProposalStatus: String {
    case voting = "Voting"
    case passed = "Passed"
    case rejected = "Rejected"

    var color: Color {
        switch self {
        case .voting: return .blue
        case .passed: return .green
        case .rejected: return .red
        }
    }
}

#Preview {
    GovernanceView()
}
