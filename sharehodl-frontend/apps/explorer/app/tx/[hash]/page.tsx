import { Card, CardContent, CardHeader, CardTitle, Button, Header } from "@repo/ui";
import { Box, ArrowLeft, Clock, CheckCircle2, ArrowRightLeft, FileText, Coins } from "lucide-react";
import Link from "next/link";

export default function TransactionDetails({ params }: { params: { hash: string } }) {
    // Mock data
    const tx = {
        hash: params.hash,
        status: "Success",
        blockHeight: 123456,
        timestamp: "2023-12-03T10:30:00Z",
        from: "Hodl46d0723646bcc9eb6bf1f382871c8b0fc32154ad",
        to: "HodlA1B2c3D4e5F6789012345678901234567890aBcD",
        amount: "1,000 SHARE",
        fee: "0.0025 SHARE",
        gasUsed: "125,000",
        gasLimit: "200,000",
        memo: "Payment for services",
    };

    return (
        <div className="min-h-screen bg-background">
            <Header appName="Explorer" />

            <main className="container mx-auto px-4 py-8">
                <div className="mb-6">
                    <Link href="/" className="flex items-center text-sm text-muted-foreground hover:text-primary mb-4">
                        <ArrowLeft className="mr-1 h-4 w-4" />
                        Back to Dashboard
                    </Link>
                    <h1 className="text-3xl font-bold flex items-center gap-3">
                        <ArrowRightLeft className="h-8 w-8 text-muted-foreground" />
                        Transaction Details
                    </h1>
                </div>

                <div className="grid gap-6">
                    <Card>
                        <CardHeader>
                            <CardTitle>Transaction Information</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            <div className="grid gap-1">
                                <div className="text-sm font-medium text-muted-foreground">Transaction Hash</div>
                                <div className="font-mono text-sm break-all">{tx.hash}</div>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div className="grid gap-1">
                                    <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                                        <CheckCircle2 className="h-4 w-4" /> Status
                                    </div>
                                    <div className="flex items-center gap-2 text-green-500 font-medium">
                                        <CheckCircle2 className="h-4 w-4" />
                                        {tx.status}
                                    </div>
                                </div>

                                <div className="grid gap-1">
                                    <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                                        <Box className="h-4 w-4" /> Block
                                    </div>
                                    <Link href={`/blocks/${tx.blockHeight}`} className="text-primary hover:underline">
                                        #{tx.blockHeight}
                                    </Link>
                                </div>

                                <div className="grid gap-1">
                                    <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                                        <Clock className="h-4 w-4" /> Timestamp
                                    </div>
                                    <div>{new Date(tx.timestamp).toLocaleString()}</div>
                                </div>

                                <div className="grid gap-1">
                                    <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                                        <Coins className="h-4 w-4" /> Fee
                                    </div>
                                    <div>{tx.fee}</div>
                                </div>
                            </div>

                            <div className="border-t pt-6 grid gap-6 md:grid-cols-2">
                                <div className="grid gap-1">
                                    <div className="text-sm font-medium text-muted-foreground">From</div>
                                    <div className="font-mono text-sm text-primary hover:underline cursor-pointer">{tx.from}</div>
                                </div>
                                <div className="grid gap-1">
                                    <div className="text-sm font-medium text-muted-foreground">To</div>
                                    <div className="font-mono text-sm text-primary hover:underline cursor-pointer">{tx.to}</div>
                                </div>
                            </div>

                            <div className="border-t pt-6">
                                <div className="grid gap-1">
                                    <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                                        <FileText className="h-4 w-4" /> Memo
                                    </div>
                                    <div className="bg-muted p-3 rounded-md text-sm font-mono">
                                        {tx.memo}
                                    </div>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </main>
        </div>
    );
}
