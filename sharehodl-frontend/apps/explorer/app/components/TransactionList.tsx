import { ArrowRightLeft } from "lucide-react";

const MOCK_TXS = [
    { hash: "Hodl123abc456def789012345678901234567890abcd", type: "Send", from: "Hodl46d0...54ad", time: "2s ago" },
    { hash: "Hodl456def789012345678901234567890abcdef123", type: "Delegate", from: "HodlA1B2...aBcD", time: "5s ago" },
    { hash: "Hodl789abc012345678901234567890abcdef456789", type: "Vote", from: "Hodlff12...3456", time: "12s ago" },
    { hash: "Hodlabc123def456789012345678901234567890abc", type: "Swap", from: "Hodl0123...4567", time: "18s ago" },
    { hash: "Hodldef456789012345678901234567890abcdef123", type: "Mint", from: "Hodldead...beef", time: "25s ago" },
];

export function TransactionList() {
    return (
        <div style={{border: "1px solid #ddd", padding: "1rem", borderRadius: "8px", height: "100%"}}>
            <div style={{marginBottom: "1rem"}}>
                <h3 style={{display: "flex", alignItems: "center", gap: "8px", fontSize: "18px", margin: "0"}}>
                    <ArrowRightLeft style={{height: "20px", width: "20px"}} />
                    Latest Transactions
                </h3>
            </div>
            <div>
                <div style={{display: "flex", flexDirection: "column", gap: "16px"}}>
                    {MOCK_TXS.map((tx) => (
                        <div
                            key={tx.hash}
                            style={{display: "flex", alignItems: "center", justifyContent: "space-between", borderBottom: "1px solid #eee", paddingBottom: "16px"}}
                        >
                            <div style={{display: "flex", alignItems: "center", gap: "16px"}}>
                                <div style={{display: "flex", height: "40px", width: "40px", alignItems: "center", justifyContent: "center", borderRadius: "8px", backgroundColor: "#f5f5f5"}}>
                                    Tx
                                </div>
                                <div>
                                    <div style={{fontWeight: "500", fontSize: "14px", fontFamily: "monospace"}}>{tx.hash}</div>
                                    <div style={{fontSize: "12px", color: "#666"}}>
                                        {tx.time}
                                    </div>
                                </div>
                            </div>
                            <div style={{textAlign: "right"}}>
                                <div style={{fontSize: "14px", fontWeight: "500"}}>
                                    {tx.type}
                                </div>
                                <div style={{fontSize: "12px", color: "#666", fontFamily: "monospace"}}>
                                    {tx.from}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
}
