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
        <div style={{border: "1px solid #ddd", padding: "1rem", borderRadius: "8px", height: "100%"}}>
            <div style={{marginBottom: "1rem"}}>
                <h3 style={{display: "flex", alignItems: "center", gap: "8px", fontSize: "18px", margin: "0"}}>
                    <Box style={{height: "20px", width: "20px"}} />
                    Latest Blocks
                </h3>
            </div>
            <div>
                <div style={{display: "flex", flexDirection: "column", gap: "16px"}}>
                    {MOCK_BLOCKS.map((block) => (
                        <div
                            key={block.height}
                            style={{display: "flex", alignItems: "center", justifyContent: "space-between", borderBottom: "1px solid #eee", paddingBottom: "16px"}}
                        >
                            <div style={{display: "flex", alignItems: "center", gap: "16px"}}>
                                <div style={{display: "flex", height: "40px", width: "40px", alignItems: "center", justifyContent: "center", borderRadius: "8px", backgroundColor: "#f5f5f5"}}>
                                    Bk
                                </div>
                                <div>
                                    <div style={{fontWeight: "500"}}>#{block.height}</div>
                                    <div style={{fontSize: "12px", color: "#666"}}>
                                        {block.time}
                                    </div>
                                </div>
                            </div>
                            <div style={{textAlign: "right"}}>
                                <div style={{fontSize: "14px", fontWeight: "500"}}>
                                    {block.validator}
                                </div>
                                <div style={{fontSize: "12px", color: "#666"}}>
                                    {block.txs} txs
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
}
