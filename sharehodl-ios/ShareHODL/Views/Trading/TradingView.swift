import SwiftUI

struct TradingView: View {
    @State private var selectedPair = "HODL/STAKE"
    @State private var orderType = "buy"
    @State private var amount = ""
    @State private var price = ""

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(spacing: 20) {
                    // Market Selector
                    Picker("Trading Pair", selection: $selectedPair) {
                        Text("HODL/STAKE").tag("HODL/STAKE")
                        Text("HODL/USDC").tag("HODL/USDC")
                    }
                    .pickerStyle(.segmented)
                    .padding(.horizontal)

                    // Price Chart Placeholder
                    RoundedRectangle(cornerRadius: 12)
                        .fill(.ultraThinMaterial)
                        .frame(height: 200)
                        .overlay {
                            VStack {
                                Image(systemName: "chart.line.uptrend.xyaxis")
                                    .font(.largeTitle)
                                    .foregroundStyle(.secondary)
                                Text("Price Chart")
                                    .foregroundStyle(.secondary)
                            }
                        }
                        .padding(.horizontal)

                    // Order Book Preview
                    HStack {
                        // Bids
                        VStack(alignment: .leading) {
                            Text("Bids")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                            ForEach(0..<5) { i in
                                HStack {
                                    Text("1.00\(i)")
                                        .font(.caption.monospaced())
                                        .foregroundStyle(.green)
                                    Spacer()
                                    Text("\(100 - i * 10)")
                                        .font(.caption.monospaced())
                                }
                            }
                        }
                        .frame(maxWidth: .infinity)

                        Divider()

                        // Asks
                        VStack(alignment: .trailing) {
                            Text("Asks")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                            ForEach(0..<5) { i in
                                HStack {
                                    Text("\(50 + i * 10)")
                                        .font(.caption.monospaced())
                                    Spacer()
                                    Text("1.0\(i + 1)")
                                        .font(.caption.monospaced())
                                        .foregroundStyle(.red)
                                }
                            }
                        }
                        .frame(maxWidth: .infinity)
                    }
                    .padding()
                    .background(.ultraThinMaterial)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .padding(.horizontal)

                    // Order Form
                    VStack(spacing: 16) {
                        Picker("Order Type", selection: $orderType) {
                            Text("Buy").tag("buy")
                            Text("Sell").tag("sell")
                        }
                        .pickerStyle(.segmented)

                        TextField("Amount", text: $amount)
                            .keyboardType(.decimalPad)
                            .textFieldStyle(.roundedBorder)

                        TextField("Price", text: $price)
                            .keyboardType(.decimalPad)
                            .textFieldStyle(.roundedBorder)

                        Button {
                            // Place order
                        } label: {
                            Text(orderType == "buy" ? "Buy HODL" : "Sell HODL")
                                .font(.headline)
                                .frame(maxWidth: .infinity)
                                .padding()
                                .background(orderType == "buy" ? .green : .red)
                                .foregroundStyle(.white)
                                .clipShape(RoundedRectangle(cornerRadius: 12))
                        }
                    }
                    .padding()
                    .background(.ultraThinMaterial)
                    .clipShape(RoundedRectangle(cornerRadius: 12))
                    .padding(.horizontal)
                }
            }
            .navigationTitle("Trade")
        }
    }
}

#Preview {
    TradingView()
}
