'use client';

import React, { useState, useEffect } from 'react';
import { WalletButton, useWallet, useBlockchain } from '@repo/ui';
import { ArrowUpDown, TrendingUp, BarChart3, Zap, Shield, Clock, RefreshCw } from 'lucide-react';
import AdvancedTrading from './components/AdvancedTrading';

export default function Home() {
  const { connected, balances, address } = useWallet();
  const { networkStatus } = useBlockchain();
  const [view, setView] = useState('swap');
  const [fromAsset, setFromAsset] = useState('HODL');
  const [toAsset, setToAsset] = useState('APPLE');
  const [fromAmount, setFromAmount] = useState('');
  const [loading, setLoading] = useState(false);

  const mockAssets = {
    'HODL': { balance: 1250.50, price: 1.00 },
    'APPLE': { balance: 15.75, price: 185.25 },
    'TSLA': { balance: 8.25, price: 245.80 },
    'GOOGL': { balance: 3.10, price: 2750.00 },
    'MSFT': { balance: 12.50, price: 385.60 }
  };

  // Get real balance for an asset
  const getRealBalance = (asset: string) => {
    if (!connected || balances.length === 0) return mockAssets[asset]?.balance || 0;

    const denom = asset === 'HODL' ? 'uhodl' : `u${asset.toLowerCase()}`;
    const balance = balances.find(b => b.denom === denom);
    if (balance) {
      return parseInt(balance.amount) / 1000000;
    }
    return mockAssets[asset]?.balance || 0;
  };

  const calculateToAmount = () => {
    if (!fromAmount || fromAmount === '0') return '';
    const fromPrice = mockAssets[fromAsset]?.price || 0;
    const toPrice = mockAssets[toAsset]?.price || 0;
    if (fromPrice && toPrice) {
      const rate = fromPrice / toPrice;
      const amount = parseFloat(fromAmount) * rate * 0.97; // 3% slippage
      return amount.toFixed(6);
    }
    return '';
  };

  const executeSwap = async () => {
    if (!fromAmount || parseFloat(fromAmount) <= 0) return;

    setLoading(true);

    try {
      // Simulate swap execution
      await new Promise(resolve => setTimeout(resolve, 2000));
      alert(`Swap executed: ${fromAmount} ${fromAsset} for ${calculateToAmount()} ${toAsset}`);
      setFromAmount('');
    } catch (error) {
      console.error('Swap error:', error);
      alert('Swap failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  if (view === 'trading') {
    return (
      <div className="min-h-screen bg-gray-950 text-white">
        <header className="border-b border-gray-800 sticky top-0 z-50 bg-gray-950/95 backdrop-blur">
          <div className="container mx-auto px-4 py-4 flex justify-between items-center">
            <div className="flex items-center gap-2">
              <BarChart3 className="h-6 w-6 text-green-400" />
              <span className="text-2xl font-bold bg-gradient-to-r from-green-500 to-emerald-500 bg-clip-text text-transparent">
                Advanced Trading
              </span>
              {networkStatus?.connected && (
                <span className="text-xs px-2 py-1 bg-green-900/30 text-green-400 rounded-full">
                  Live
                </span>
              )}
            </div>
            <div className="flex items-center gap-4">
              <button
                onClick={() => setView('swap')}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
              >
                Simple Trading
              </button>
              <WalletButton />
            </div>
          </div>
        </header>
        <AdvancedTrading symbol="APPLE/HODL" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      {/* Header */}
      <header className="border-b border-gray-800 sticky top-0 z-50 bg-gray-950/95 backdrop-blur">
        <div className="container mx-auto px-4 py-4 flex justify-between items-center">
          <div className="flex items-center gap-2">
            <TrendingUp className="h-6 w-6 text-green-400" />
            <span className="text-2xl font-bold bg-gradient-to-r from-green-500 to-emerald-500 bg-clip-text text-transparent">
              ShareHODL Trading
            </span>
            {networkStatus?.connected && (
              <span className="text-xs px-2 py-1 bg-green-900/30 text-green-400 rounded-full">
                Live
              </span>
            )}
          </div>
          <div className="flex items-center gap-4">
            <nav className="hidden md:flex items-center gap-6 text-sm">
              <a href="http://localhost:3000" className="text-gray-400 hover:text-white transition">Home</a>
              <a href="http://localhost:3003" className="text-gray-400 hover:text-white transition">Explorer</a>
              <a href="http://localhost:3004" className="text-gray-400 hover:text-white transition">Wallet</a>
            </nav>
            <WalletButton />
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-8">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold mb-4 bg-gradient-to-r from-green-400 via-emerald-400 to-teal-400 bg-clip-text text-transparent">
            ShareHODL Trading
          </h1>
          <p className="text-gray-400 text-lg max-w-2xl mx-auto">
            Trade tokenized equities with instant settlement, ultra-low fees, and 24/7 access.
          </p>

          {/* View Toggle */}
          <div className="mt-6 flex justify-center gap-4">
            <button
              onClick={() => setView('swap')}
              className={`px-6 py-3 rounded-lg font-semibold transition-colors ${
                view === 'swap'
                  ? 'bg-blue-600 text-white'
                  : 'border border-gray-700 text-gray-400 hover:border-gray-600'
              }`}
            >
              Simple Trading
            </button>
            <button
              onClick={() => setView('trading')}
              className={`px-6 py-3 rounded-lg font-semibold transition-colors ${
                view === 'trading'
                  ? 'bg-green-600 text-white'
                  : 'border border-gray-700 text-gray-400 hover:border-gray-600'
              }`}
            >
              Advanced Trading
            </button>
          </div>
        </div>

        {/* Network Status */}
        <div className={`mb-8 p-4 rounded-xl border ${
          networkStatus?.connected
            ? 'bg-green-900/10 border-green-800'
            : 'bg-yellow-900/10 border-yellow-800'
        }`}>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className={`w-3 h-3 rounded-full ${
                networkStatus?.connected ? 'bg-green-400 animate-pulse' : 'bg-yellow-400'
              }`} />
              <div>
                <p className={`font-medium ${
                  networkStatus?.connected ? 'text-green-400' : 'text-yellow-400'
                }`}>
                  {networkStatus?.connected ? 'Connected to ShareHODL Network' : 'Connecting...'}
                </p>
                <p className="text-sm text-gray-500">
                  {networkStatus?.chainId || 'sharehodl-1'} | Block #{networkStatus?.latestBlockHeight || '---'}
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="grid gap-8 lg:grid-cols-3">
          {/* Swap Card */}
          <div className="lg:col-span-2">
            <div className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
              <h3 className="font-semibold mb-6 text-lg">Swap Assets</h3>

              <div className="space-y-4">
                {/* From Input */}
                <div>
                  <label className="block text-sm text-gray-400 mb-2">From</label>
                  <div className="flex gap-2">
                    <select
                      value={fromAsset}
                      onChange={(e) => setFromAsset(e.target.value)}
                      className="bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-green-500"
                    >
                      {Object.keys(mockAssets).map(asset => (
                        <option key={asset} value={asset}>{asset}</option>
                      ))}
                    </select>
                    <input
                      type="number"
                      value={fromAmount}
                      onChange={(e) => setFromAmount(e.target.value)}
                      placeholder="0.00"
                      className="flex-1 bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-green-500"
                    />
                    <button
                      onClick={() => setFromAmount(getRealBalance(fromAsset).toString())}
                      className="px-4 py-3 bg-green-600 hover:bg-green-700 text-white rounded-lg text-sm font-semibold transition-colors"
                    >
                      MAX
                    </button>
                  </div>
                  <p className="text-xs text-gray-500 mt-2">
                    Balance: {getRealBalance(fromAsset).toLocaleString(undefined, { minimumFractionDigits: 2 })} {fromAsset}
                  </p>
                </div>

                {/* Swap Button */}
                <div className="flex justify-center">
                  <button
                    onClick={() => {
                      const temp = fromAsset;
                      setFromAsset(toAsset);
                      setToAsset(temp);
                    }}
                    className="p-3 bg-gray-800 hover:bg-gray-700 rounded-full transition-colors"
                  >
                    <ArrowUpDown className="h-5 w-5 text-gray-400" />
                  </button>
                </div>

                {/* To Input */}
                <div>
                  <label className="block text-sm text-gray-400 mb-2">To</label>
                  <div className="flex gap-2">
                    <select
                      value={toAsset}
                      onChange={(e) => setToAsset(e.target.value)}
                      className="bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-green-500"
                    >
                      {Object.keys(mockAssets).map(asset => (
                        <option key={asset} value={asset}>{asset}</option>
                      ))}
                    </select>
                    <input
                      type="number"
                      value={calculateToAmount()}
                      readOnly
                      placeholder="0.00"
                      className="flex-1 bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 text-gray-400"
                    />
                  </div>
                  <p className="text-xs text-gray-500 mt-2">
                    Balance: {getRealBalance(toAsset).toLocaleString(undefined, { minimumFractionDigits: 2 })} {toAsset}
                  </p>
                </div>

                {/* Rate Info */}
                <div className="p-4 bg-gray-800 rounded-lg space-y-2">
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-400">Exchange Rate</span>
                    <span>1 {fromAsset} = {((mockAssets[fromAsset]?.price || 0) / (mockAssets[toAsset]?.price || 1)).toFixed(6)} {toAsset}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-400">Slippage Tolerance</span>
                    <span>3%</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-400">Trading Fee</span>
                    <span className="text-green-400">0.3%</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-400">Settlement</span>
                    <span className="text-green-400">~2 seconds</span>
                  </div>
                </div>

                {/* Action Button */}
                {!connected ? (
                  <div className="space-y-3">
                    <WalletButton />
                    <p className="text-center text-sm text-gray-500">Connect your wallet to start trading</p>
                  </div>
                ) : (
                  <button
                    onClick={executeSwap}
                    disabled={loading || !fromAmount || fromAsset === toAsset}
                    className="w-full bg-green-600 hover:bg-green-700 disabled:bg-gray-700 disabled:cursor-not-allowed text-white py-4 rounded-lg font-semibold transition-colors flex items-center justify-center gap-2"
                  >
                    {loading ? (
                      <>
                        <RefreshCw className="h-5 w-5 animate-spin" />
                        Processing...
                      </>
                    ) : (
                      'Swap'
                    )}
                  </button>
                )}
              </div>
            </div>
          </div>

          {/* Market Info Sidebar */}
          <div className="space-y-6">
            <div className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
              <h3 className="font-semibold mb-4">Market Overview</h3>
              <div className="space-y-4">
                {Object.entries(mockAssets).slice(1).map(([symbol, data]) => (
                  <div key={symbol} className="flex justify-between items-center">
                    <div>
                      <p className="font-medium">{symbol}</p>
                      <p className="text-sm text-gray-500">${data.price.toLocaleString()}</p>
                    </div>
                    <span className="text-green-400 text-sm">+2.4%</span>
                  </div>
                ))}
              </div>
            </div>

            <div className="border border-gray-800 rounded-xl p-6 bg-gray-900/50">
              <h3 className="font-semibold mb-4">Platform Stats</h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-gray-400">24h Volume</span>
                  <span>$12.5M</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Active Markets</span>
                  <span>4</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-400">Total Trades</span>
                  <span>24,891</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Why Trade Section */}
        <div className="mt-12 grid gap-6 md:grid-cols-4">
          <div className="border border-gray-800 rounded-xl p-6 text-center bg-gray-900/50">
            <Zap className="h-8 w-8 text-yellow-400 mx-auto mb-3" />
            <h3 className="font-semibold mb-2">2-Second Settlement</h3>
            <p className="text-sm text-gray-500">vs 2-3 days traditional</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-6 text-center bg-gray-900/50">
            <Clock className="h-8 w-8 text-blue-400 mx-auto mb-3" />
            <h3 className="font-semibold mb-2">24/7 Trading</h3>
            <p className="text-sm text-gray-500">Markets never close</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-6 text-center bg-gray-900/50">
            <TrendingUp className="h-8 w-8 text-green-400 mx-auto mb-3" />
            <h3 className="font-semibold mb-2">Ultra-Low Fees</h3>
            <p className="text-sm text-gray-500">0.3% trading fee</p>
          </div>
          <div className="border border-gray-800 rounded-xl p-6 text-center bg-gray-900/50">
            <Shield className="h-8 w-8 text-purple-400 mx-auto mb-3" />
            <h3 className="font-semibold mb-2">Circuit Breakers</h3>
            <p className="text-sm text-gray-500">Built-in protection</p>
          </div>
        </div>

        {/* Footer */}
        <div className="text-center text-gray-500 mt-12 pt-8 border-t border-gray-800">
          <p className="mb-2 text-gray-400">ShareHODL Trading Platform</p>
          <p className="text-sm mb-4">
            Fast, secure, and affordable trading for tokenized equities.
          </p>
          <div className="flex justify-center items-center gap-6">
            <a
              href="https://x.com/share_hodl"
              target="_blank"
              rel="noopener noreferrer"
              className="text-gray-500 hover:text-white transition-colors flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
              </svg>
              <span className="text-sm">@share_hodl</span>
            </a>
            <a
              href="https://github.com/x-word-wide/sharehodl"
              target="_blank"
              rel="noopener noreferrer"
              className="text-gray-500 hover:text-white transition-colors flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                <path fillRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clipRule="evenodd"/>
              </svg>
              <span className="text-sm">GitHub</span>
            </a>
          </div>
        </div>
      </main>
    </div>
  );
}
