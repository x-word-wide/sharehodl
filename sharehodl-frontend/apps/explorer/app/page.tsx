'use client';

import { useState, useCallback } from "react";
import { Card, CardContent, CardHeader, CardTitle, useBlockchain, WalletButton } from "@repo/ui";
import { BlockList } from "./components/BlockList";
import { TransactionList } from "./components/TransactionList";
import { Search, Activity, Box, Users, Zap, RefreshCw, Wifi, WifiOff } from "lucide-react";

export default function Home() {
  const { networkStatus, loading, refresh, search } = useBlockchain();
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResult, setSearchResult] = useState<{ type: string; data: any } | null>(null);
  const [searching, setSearching] = useState(false);

  const handleSearch = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    if (!searchQuery.trim()) return;

    setSearching(true);
    try {
      const result = await search(searchQuery.trim());
      setSearchResult(result);
    } catch (error) {
      console.error('Search error:', error);
    } finally {
      setSearching(false);
    }
  }, [searchQuery, search]);

  const formatBlockHeight = (height: string | undefined) => {
    if (!height) return '---';
    return parseInt(height).toLocaleString();
  };

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      <main className="container mx-auto px-4 py-8">
        {/* Header with Wallet */}
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold">ShareHODL Explorer</h1>
            <p className="text-gray-400">Explore blocks, transactions, and addresses</p>
          </div>
          <WalletButton showBalance />
        </div>

        {/* Network Status Banner */}
        <div className={`mb-8 p-4 rounded-lg flex items-center justify-between ${
          networkStatus?.connected ? 'bg-green-900/20 border border-green-800' : 'bg-red-900/20 border border-red-800'
        }`}>
          <div className="flex items-center gap-3">
            {networkStatus?.connected ? (
              <Wifi className="h-5 w-5 text-green-400" />
            ) : (
              <WifiOff className="h-5 w-5 text-red-400" />
            )}
            <div>
              <p className={`font-medium ${networkStatus?.connected ? 'text-green-400' : 'text-red-400'}`}>
                {networkStatus?.connected ? 'Connected to ShareHODL Network' : 'Disconnected from Network'}
              </p>
              <p className="text-sm text-gray-400">
                {networkStatus?.chainId || 'sharehodl-1'} | Block #{formatBlockHeight(networkStatus?.latestBlockHeight)}
              </p>
            </div>
          </div>
          <button
            onClick={refresh}
            disabled={loading}
            className="p-2 hover:bg-gray-800 rounded-lg transition-colors"
          >
            <RefreshCw className={`h-5 w-5 text-gray-400 ${loading ? 'animate-spin' : ''}`} />
          </button>
        </div>

        {/* Search Section */}
        <section className="mb-8">
          <form onSubmit={handleSearch} className="relative max-w-3xl mx-auto">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 h-5 w-5 text-gray-500" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search by block height, transaction hash, or address (hodl...)"
              className="w-full rounded-xl border border-gray-700 bg-gray-900 py-4 pl-12 pr-24 text-white placeholder:text-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <button
              type="submit"
              disabled={searching || !searchQuery.trim()}
              className="absolute right-2 top-1/2 -translate-y-1/2 px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white rounded-lg transition-colors"
            >
              {searching ? 'Searching...' : 'Search'}
            </button>
          </form>

          {/* Search Result Display */}
          {searchResult && (
            <div className="mt-4 max-w-3xl mx-auto p-4 bg-gray-900 rounded-lg border border-gray-700">
              {searchResult.type === 'unknown' ? (
                <p className="text-gray-400">No results found for "{searchQuery}"</p>
              ) : (
                <div>
                  <p className="text-sm text-gray-400 mb-2">
                    Found {searchResult.type}:
                  </p>
                  <pre className="text-sm text-gray-300 overflow-x-auto">
                    {JSON.stringify(searchResult.data, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          )}
        </section>

        {/* Network Stats */}
        <section className="mb-8 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card className="bg-gray-900 border-gray-700">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-gray-400">
                Network
              </CardTitle>
              <Activity className="h-4 w-4 text-blue-400" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-white">
                {networkStatus?.chainId || 'sharehodl-1'}
              </div>
              <p className="text-xs text-gray-500">
                {networkStatus?.catching_up ? 'Syncing...' : 'Fully synced'}
              </p>
            </CardContent>
          </Card>

          <Card className="bg-gray-900 border-gray-700">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-gray-400">
                Block Height
              </CardTitle>
              <Box className="h-4 w-4 text-green-400" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-white">
                {formatBlockHeight(networkStatus?.latestBlockHeight)}
              </div>
              <p className="text-xs text-gray-500">
                ~2 second blocks
              </p>
            </CardContent>
          </Card>

          <Card className="bg-gray-900 border-gray-700">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-gray-400">
                Validators
              </CardTitle>
              <Users className="h-4 w-4 text-purple-400" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-white">
                {networkStatus?.validatorCount || 1}
              </div>
              <p className="text-xs text-gray-500">
                Active validators
              </p>
            </CardContent>
          </Card>

          <Card className="bg-gray-900 border-gray-700">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-gray-400">
                Block Time
              </CardTitle>
              <Zap className="h-4 w-4 text-yellow-400" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-white">~2.0s</div>
              <p className="text-xs text-gray-500">
                Average block time
              </p>
            </CardContent>
          </Card>
        </section>

        {/* Blocks and Transactions */}
        <div className="grid gap-8 md:grid-cols-2">
          <BlockList />
          <TransactionList />
        </div>
      </main>
    </div>
  );
}
