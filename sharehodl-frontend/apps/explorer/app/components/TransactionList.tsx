import { Card, CardContent, CardHeader, CardTitle } from "@repo/ui";
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
        <Card className="h-full">
            <CardHeader>
                <CardTitle className="flex items-center gap-2 text-lg">
                    <ArrowRightLeft className="h-5 w-5" />
                    Latest Transactions
                </CardTitle>
            </CardHeader>
            <CardContent>
                <div className="space-y-4">
                    {MOCK_TXS.map((tx) => (
                        <div
                            key={tx.hash}
                            className="flex items-center justify-between border-b pb-4 last:border-0 last:pb-0"
                        >
                            <div className="flex items-center gap-4">
                                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-secondary text-secondary-foreground">
                                    Tx
                                </div>
                                <div>
                                    <div className="font-medium text-primary">{tx.hash}</div>
                                    <div className="text-xs text-muted-foreground">
                                        {tx.time}
                                    </div>
                                </div>
                            </div>
                            <div className="text-right">
                                <div className="text-sm font-medium">
                                    {tx.type}
                                </div>
                                <div className="text-xs text-muted-foreground">
                                    {tx.from}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </CardContent>
        </Card>
    );
}
