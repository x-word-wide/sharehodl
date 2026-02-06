'use client';

import { Box, RefreshCw } from "lucide-react";
import { useBlockchain } from "@repo/ui";

function formatTimeAgo(dateString: string): string {
    const date = new Date(dateString);
    const now = new Date();
    const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (seconds < 60) return `${seconds}s ago`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
}

export function BlockList() {
    const { blocks, loading, error, fetchBlocks } = useBlockchain();

    const handleRefresh = () => {
        fetchBlocks(10);
    };

    if (error) {
        return (
            <div className="border border-gray-700 p-4 rounded-lg h-full bg-gray-900">
                <div className="flex items-center justify-between mb-4">
                    <h3 className="flex items-center gap-2 text-lg font-semibold text-white">
                        <Box className="h-5 w-5" />
                        Latest Blocks
                    </h3>
                </div>
                <div className="text-center text-red-400 py-8">
                    <p>Failed to load blocks</p>
                    <p className="text-sm text-gray-500 mt-2">{error}</p>
                    <button
                        onClick={handleRefresh}
                        className="mt-4 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm"
                    >
                        Retry
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="border border-gray-700 p-4 rounded-lg h-full bg-gray-900">
            <div className="flex items-center justify-between mb-4">
                <h3 className="flex items-center gap-2 text-lg font-semibold text-white">
                    <Box className="h-5 w-5" />
                    Latest Blocks
                </h3>
                <button
                    onClick={handleRefresh}
                    className="p-2 hover:bg-gray-800 rounded-lg transition-colors"
                    disabled={loading}
                >
                    <RefreshCw className={`h-4 w-4 text-gray-400 ${loading ? 'animate-spin' : ''}`} />
                </button>
            </div>

            {loading && blocks.length === 0 ? (
                <div className="flex flex-col gap-4">
                    {[...Array(5)].map((_, i) => (
                        <div key={i} className="animate-pulse flex items-center justify-between pb-4 border-b border-gray-800">
                            <div className="flex items-center gap-4">
                                <div className="h-10 w-10 bg-gray-800 rounded-lg" />
                                <div>
                                    <div className="h-4 w-20 bg-gray-800 rounded mb-2" />
                                    <div className="h-3 w-16 bg-gray-800 rounded" />
                                </div>
                            </div>
                            <div className="text-right">
                                <div className="h-4 w-24 bg-gray-800 rounded mb-2" />
                                <div className="h-3 w-12 bg-gray-800 rounded" />
                            </div>
                        </div>
                    ))}
                </div>
            ) : blocks.length === 0 ? (
                <div className="text-center text-gray-500 py-8">
                    No blocks found
                </div>
            ) : (
                <div className="flex flex-col gap-4">
                    {blocks.map((block) => (
                        <div
                            key={block.height}
                            className="flex items-center justify-between pb-4 border-b border-gray-800 last:border-b-0 hover:bg-gray-800/50 -mx-2 px-2 rounded transition-colors cursor-pointer"
                        >
                            <div className="flex items-center gap-4">
                                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-900/30 text-blue-400 font-mono text-sm">
                                    Bk
                                </div>
                                <div>
                                    <div className="font-medium text-white">
                                        #{parseInt(block.height).toLocaleString()}
                                    </div>
                                    <div className="text-xs text-gray-500">
                                        {formatTimeAgo(block.time)}
                                    </div>
                                </div>
                            </div>
                            <div className="text-right">
                                <div className="text-sm font-medium text-gray-300">
                                    {block.proposer}
                                </div>
                                <div className="text-xs text-gray-500">
                                    {block.txCount} txs
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            <div className="mt-4 pt-4 border-t border-gray-800">
                <a
                    href="/blocks"
                    className="text-sm text-blue-400 hover:text-blue-300 transition-colors"
                >
                    View all blocks →
                </a>
            </div>
        </div>
    );
}
