import { Card, CardContent, CardHeader, CardTitle, Button, Header } from "@repo/ui";
import { Box, ArrowLeft, Clock, Shield, Database, Activity } from "lucide-react";
import Link from "next/link";

export default function BlockDetails({ params }: { params: { height: string } }) {
    // Mock data - in a real app this would be fetched based on params.height
    const block = {
        height: params.height,
        hash: "Hodl7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa3d677284addd200126d9069",
        timestamp: "2023-12-03T10:30:00Z",
        validator: "Validator One",
        validatorAddress: "Hodl0123456789abcdef0123456789abcdef01234567",
        txCount: 12,
        gasUsed: "1,234,567",
        gasLimit: "10,000,000",
        size: "45.2 KB",
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
                        <Box className="h-8 w-8 text-muted-foreground" />
                        Block #{block.height}
                    </h1>
                </div>

                <div className="grid gap-6 md:grid-cols-3">
                    <Card className="md:col-span-2">
                        <CardHeader>
                            <CardTitle>Block Information</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            <div className="grid gap-1">
                                <div className="text-sm font-medium text-muted-foreground">Block Hash</div>
                                <div className="font-mono text-sm break-all">{block.hash}</div>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div className="grid gap-1">
                                    <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                                        <Clock className="h-4 w-4" /> Timestamp
                                    </div>
                                    <div>{new Date(block.timestamp).toLocaleString()}</div>
                                </div>

                                <div className="grid gap-1">
                                    <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                                        <Shield className="h-4 w-4" /> Proposer
                                    </div>
                                    <div className="text-primary hover:underline cursor-pointer">{block.validator}</div>
                                </div>

                                <div className="grid gap-1">
                                    <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                                        <Activity className="h-4 w-4" /> Transactions
                                    </div>
                                    <div>{block.txCount} transactions</div>
                                </div>

                                <div className="grid gap-1">
                                    <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                                        <Database className="h-4 w-4" /> Size
                                    </div>
                                    <div>{block.size}</div>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>Gas & Consensus</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid gap-1">
                                <div className="text-sm font-medium text-muted-foreground">Gas Used</div>
                                <div className="text-xl font-bold">{block.gasUsed}</div>
                                <div className="text-xs text-muted-foreground">
                                    {(parseInt(block.gasUsed.replace(/,/g, '')) / parseInt(block.gasLimit.replace(/,/g, '')) * 100).toFixed(2)}% of Limit
                                </div>
                                <div className="h-2 w-full bg-secondary rounded-full mt-1 overflow-hidden">
                                    <div
                                        className="h-full bg-primary"
                                        style={{ width: `${(parseInt(block.gasUsed.replace(/,/g, '')) / parseInt(block.gasLimit.replace(/,/g, '')) * 100)}%` }}
                                    />
                                </div>
                            </div>

                            <div className="grid gap-1 pt-4 border-t">
                                <div className="text-sm font-medium text-muted-foreground">Gas Limit</div>
                                <div>{block.gasLimit}</div>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </main>
        </div>
    );
}
