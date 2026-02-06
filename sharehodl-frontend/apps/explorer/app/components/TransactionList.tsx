'use client';

import { ArrowRightLeft, RefreshCw, CheckCircle, XCircle } from "lucide-react";
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

function formatHash(hash: string): string {
    if (!hash) return '';
    return `${hash.slice(0, 10)}...${hash.slice(-8)}`;
}

function formatAddress(address: string): string {
    if (!address) return '';
    return `${address.slice(0, 8)}...${address.slice(-4)}`;
}

export function TransactionList() {
    const { transactions, loading, fetchTransactions } = useBlockchain();

    const handleRefresh = () => {
        fetchTransactions(10);
    };

    return (
        <div className="border border-gray-700 p-4 rounded-lg h-full bg-gray-900">
            <div className="flex items-center justify-between mb-4">
                <h3 className="flex items-center gap-2 text-lg font-semibold text-white">
                    <ArrowRightLeft className="h-5 w-5" />
                    Latest Transactions
                </h3>
                <button
                    onClick={handleRefresh}
                    className="p-2 hover:bg-gray-800 rounded-lg transition-colors"
                    disabled={loading}
                >
                    <RefreshCw className={`h-4 w-4 text-gray-400 ${loading ? 'animate-spin' : ''}`} />
                </button>
            </div>

            {loading && transactions.length === 0 ? (
                <div className="flex flex-col gap-4">
                    {[...Array(5)].map((_, i) => (
                        <div key={i} className="animate-pulse flex items-center justify-between pb-4 border-b border-gray-800">
                            <div className="flex items-center gap-4">
                                <div className="h-10 w-10 bg-gray-800 rounded-lg" />
                                <div>
                                    <div className="h-4 w-32 bg-gray-800 rounded mb-2" />
                                    <div className="h-3 w-16 bg-gray-800 rounded" />
                                </div>
                            </div>
                            <div className="text-right">
                                <div className="h-4 w-20 bg-gray-800 rounded mb-2" />
                                <div className="h-3 w-24 bg-gray-800 rounded" />
                            </div>
                        </div>
                    ))}
                </div>
            ) : transactions.length === 0 ? (
                <div className="text-center text-gray-500 py-8">
                    <p>No transactions found</p>
                    <p className="text-xs mt-2">Transactions will appear here once activity occurs on the network</p>
                </div>
            ) : (
                <div className="flex flex-col gap-4">
                    {transactions.map((tx) => (
                        <div
                            key={tx.hash}
                            className="flex items-center justify-between pb-4 border-b border-gray-800 last:border-b-0 hover:bg-gray-800/50 -mx-2 px-2 rounded transition-colors cursor-pointer"
                        >
                            <div className="flex items-center gap-4">
                                <div className={`flex h-10 w-10 items-center justify-center rounded-lg ${
                                    tx.status === 'success'
                                        ? 'bg-green-900/30 text-green-400'
                                        : 'bg-red-900/30 text-red-400'
                                }`}>
                                    {tx.status === 'success' ? (
                                        <CheckCircle className="h-5 w-5" />
                                    ) : (
                                        <XCircle className="h-5 w-5" />
                                    )}
                                </div>
                                <div>
                                    <div className="font-medium text-white font-mono text-sm">
                                        {formatHash(tx.hash)}
                                    </div>
                                    <div className="text-xs text-gray-500">
                                        {formatTimeAgo(tx.time)}
                                    </div>
                                </div>
                            </div>
                            <div className="text-right">
                                <div className={`text-sm font-medium px-2 py-0.5 rounded ${
                                    tx.type === 'Send' ? 'bg-blue-900/30 text-blue-400' :
                                    tx.type === 'Delegate' ? 'bg-purple-900/30 text-purple-400' :
                                    tx.type === 'Vote' ? 'bg-orange-900/30 text-orange-400' :
                                    'bg-gray-800 text-gray-300'
                                }`}>
                                    {tx.type}
                                </div>
                                <div className="text-xs text-gray-500 font-mono mt-1">
                                    {tx.from ? formatAddress(tx.from) : `Block #${tx.height}`}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            <div className="mt-4 pt-4 border-t border-gray-800">
                <a
                    href="/transactions"
                    className="text-sm text-blue-400 hover:text-blue-300 transition-colors"
                >
                    View all transactions →
                </a>
            </div>
        </div>
    );
}
