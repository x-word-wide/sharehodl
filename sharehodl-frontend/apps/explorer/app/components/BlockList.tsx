import { Card, CardContent, CardHeader, CardTitle } from "@repo/ui";
import { Box } from "lucide-react";

const MOCK_BLOCKS = [
    { height: 123456, validator: "Validator One", txs: 12, time: "2s ago" },
    { height: 123455, validator: "Validator Two", txs: 5, time: "8s ago" },
    { height: 123454, validator: "Validator Three", txs: 24, time: "14s ago" },
    { height: 123453, validator: "Validator One", txs: 8, time: "20s ago" },
    { height: 123452, validator: "Validator Four", txs: 15, time: "26s ago" },
];

export function BlockList() {
    return (
        <Card className="h-full">
            <CardHeader>
                <CardTitle className="flex items-center gap-2 text-lg">
                    <Box className="h-5 w-5" />
                    Latest Blocks
                </CardTitle>
            </CardHeader>
            <CardContent>
                <div className="space-y-4">
                    {MOCK_BLOCKS.map((block) => (
                        <div
                            key={block.height}
                            className="flex items-center justify-between border-b pb-4 last:border-0 last:pb-0"
                        >
                            <div className="flex items-center gap-4">
                                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-secondary text-secondary-foreground">
                                    Bk
                                </div>
                                <div>
                                    <div className="font-medium text-primary">#{block.height}</div>
                                    <div className="text-xs text-muted-foreground">
                                        {block.time}
                                    </div>
                                </div>
                            </div>
                            <div className="text-right">
                                <div className="text-sm font-medium">
                                    {block.validator}
                                </div>
                                <div className="text-xs text-muted-foreground">
                                    {block.txs} txs
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </CardContent>
        </Card>
    );
}
