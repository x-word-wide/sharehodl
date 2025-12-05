'use client';

import React, { useState, useEffect } from 'react';
import { ShareHODLAPI, ShareHODLWebSocket } from './lib/api';
import AdvancedTrading from './components/AdvancedTrading';

export default function Home() {
  const [view, setView] = useState('swap'); // 'swap' or 'trading'
  const [fromAsset, setFromAsset] = useState('HODL');
  const [toAsset, setToAsset] = useState('APPLE');
  const [fromAmount, setFromAmount] = useState('');
  const [connected, setConnected] = useState(false);
  const [loading, setLoading] = useState(false);
  const [balances, setBalances] = useState<{ [key: string]: string }>({});
  const [networkStatus, setNetworkStatus] = useState<any>(null);

  const mockAssets = {
    'HODL': { balance: 1250.50, price: 1.00 },
    'APPLE': { balance: 15.75, price: 185.25 },
    'TSLA': { balance: 8.25, price: 245.80 },
    'GOOGL': { balance: 3.10, price: 2750.00 },
    'MSFT': { balance: 12.50, price: 385.60 }
  };

  // Initialize WebSocket connection and load data
  useEffect(() => {
    const ws = new ShareHODLWebSocket();
    
    // Try to connect to the blockchain
    const checkConnection = async () => {
      try {
        const status = await ShareHODLAPI.getNetworkStatus();
        if (status) {
          setNetworkStatus(status);
          setConnected(true);
          
          // Load real balances if connected
          const mockAddress = 'sharehodl1234567890abcdef1234567890abcdef12345678';
          const userBalances = await ShareHODLAPI.getBalances(mockAddress);
          const balanceMap: { [key: string]: string } = {};
          userBalances.forEach(balance => {
            balanceMap[balance.denom] = ShareHODLAPI.formatAmount(balance.amount);
          });
          setBalances(balanceMap);
        }
      } catch (error) {
        console.log('Blockchain not available, using mock data');
      }
    };

    checkConnection();
    
    // Set up real-time updates when blockchain is available
    if (connected) {
      ws.connect();
      ws.on('NewBlock', (data) => {
        console.log('New block:', data);
      });
    }

    return () => {
      ws.disconnect();
    };
  }, [connected]);

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

  const connectWallet = async () => {
    setLoading(true);
    await new Promise(resolve => setTimeout(resolve, 2000));
    setConnected(true);
    setLoading(false);
  };

  const executeSwap = async () => {
    if (!fromAmount || parseFloat(fromAmount) <= 0) return;
    
    setLoading(true);
    
    try {
      // Use real API if connected, otherwise simulate
      if (networkStatus) {
        const result = await ShareHODLAPI.executeSwap(
          fromAsset.toLowerCase(),
          toAsset.toLowerCase(), 
          ShareHODLAPI.toChainAmount(fromAmount),
          0.03 // 3% slippage
        );
        
        if (result.success) {
          alert(`Swap successful! Transaction: ${result.txHash}`);
        } else {
          alert(`Swap failed: ${result.error}`);
        }
      } else {
        // Mock execution
        await new Promise(resolve => setTimeout(resolve, 3000));
        alert(`Mock swap: ${fromAmount} ${fromAsset} for ${calculateToAmount()} ${toAsset}`);
      }
      
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
      <div className="min-h-screen bg-background">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between p-4 border-b gap-4">
          <h1 className="text-xl sm:text-2xl font-bold flex items-center gap-2">
            📈 ShareHODL Advanced Trading
          </h1>
          <button 
            onClick={() => setView('swap')}
            className="px-4 py-2 bg-blue-500 text-white rounded w-full sm:w-auto"
          >
            💱 Switch to Simple Trading
          </button>
        </div>
        <AdvancedTrading symbol="APPLE/HODL" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <main className="container mx-auto px-4 py-8">
        <div className="text-center mb-12">
          <h1 className="text-3xl sm:text-4xl font-bold mb-4 flex items-center justify-center gap-3">
            ShareHODL Trading
          </h1>
          <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
            Fast, secure, and affordable trading for everyone - from beginners to professionals.
          </p>
          <div className="mt-6 flex flex-col sm:flex-row justify-center gap-4">
            <button 
              onClick={() => setView('swap')}
              className={`px-6 py-3 rounded-lg font-semibold ${
                view === 'swap' 
                  ? 'bg-blue-500 text-white' 
                  : 'border border-blue-500 text-blue-500'
              }`}
            >
              💱 Simple Trading
            </button>
            <button 
              onClick={() => setView('trading')}
              className={`px-6 py-3 rounded-lg font-semibold ${
                view === 'trading' 
                  ? 'bg-green-500 text-white' 
                  : 'border border-green-500 text-green-500'
              }`}
            >
              📈 Advanced Trading
            </button>
          </div>
        </div>

        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <div className="border rounded-lg p-6">
              <h3 className="font-semibold mb-4">💱 Simple Trading - Swap Assets Instantly</h3>
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium mb-2">From</label>
                  <div className="flex gap-2">
                    <select 
                      value={fromAsset}
                      onChange={(e) => setFromAsset(e.target.value)}
                      className="border rounded px-3 py-2 bg-background"
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
                      className="flex-1 border rounded px-3 py-2 bg-background"
                    />
                    <button 
                      onClick={() => setFromAmount(mockAssets[fromAsset]?.balance.toString())}
                      className="px-3 py-2 bg-blue-500 text-white rounded text-sm"
                    >
                      MAX
                    </button>
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">
                    Balance: {mockAssets[fromAsset]?.balance} {fromAsset}
                  </p>
                </div>

                <div className="text-center">
                  <button 
                    onClick={() => {
                      const temp = fromAsset;
                      setFromAsset(toAsset);
                      setToAsset(temp);
                    }}
                    className="p-2 border rounded-full hover:bg-muted"
                  >
                    ↕
                  </button>
                </div>

                <div>
                  <label className="block text-sm font-medium mb-2">To</label>
                  <div className="flex gap-2">
                    <select 
                      value={toAsset}
                      onChange={(e) => setToAsset(e.target.value)}
                      className="border rounded px-3 py-2 bg-background"
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
                      className="flex-1 border rounded px-3 py-2 bg-muted text-muted-foreground"
                    />
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">
                    Balance: {mockAssets[toAsset]?.balance} {toAsset}
                  </p>
                </div>

                <div className="p-3 bg-muted rounded">
                  <div className="flex justify-between text-sm mb-1">
                    <span>Exchange Rate</span>
                    <span>1 {fromAsset} = {((mockAssets[fromAsset]?.price || 0) / (mockAssets[toAsset]?.price || 1)).toFixed(6)} {toAsset}</span>
                  </div>
                  <div className="flex justify-between text-sm mb-1">
                    <span>Slippage</span>
                    <span>3%</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Fee</span>
                    <span>0.3%</span>
                  </div>
                </div>

                {!connected ? (
                  <button 
                    onClick={connectWallet}
                    disabled={loading}
                    className="w-full bg-blue-500 text-white py-3 rounded font-semibold disabled:opacity-50"
                  >
                    {loading ? 'Connecting...' : '🔗 Connect Wallet'}
                  </button>
                ) : (
                  <button 
                    onClick={executeSwap}
                    disabled={loading || !fromAmount || fromAsset === toAsset}
                    className="w-full bg-blue-500 text-white py-3 rounded font-semibold disabled:opacity-50"
                  >
                    {loading ? 'Trading...' : '💱 Trade Now'}
                  </button>
                )}

                <div className={`p-3 rounded border ${
                  networkStatus 
                    ? 'bg-green-50 border-green-200' 
                    : 'bg-yellow-50 border-yellow-200'
                }`}>
                  <p className={`text-sm ${
                    networkStatus ? 'text-green-800' : 'text-yellow-800'
                  }`}>
                    {networkStatus ? (
                      <>
                        Connected to ShareHODL Network<br/>
                        Chain: {networkStatus.node_info?.network}<br/>
                        Height: {networkStatus.sync_info?.latest_block_height}
                      </>
                    ) : (
                      'Blockchain offline - Using mock data for demo'
                    )}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="grid gap-6 md:grid-cols-3 mt-8">
          <div className="border rounded-lg p-6">
            <div className="mb-4">
              <h3 className="font-semibold flex items-center gap-2">Market Volume</h3>
            </div>
            <div>
              <div className="text-2xl font-bold">$12.5M</div>
              <p className="text-sm text-muted-foreground">24h trading volume</p>
            </div>
          </div>

          <div className="border rounded-lg p-6">
            <div className="mb-4">
              <h3 className="font-semibold flex items-center gap-2">Active Markets</h3>
            </div>
            <div>
              <div className="text-2xl font-bold">4</div>
              <p className="text-sm text-muted-foreground">Live trading pairs</p>
            </div>
          </div>

          <div className="border rounded-lg p-6">
            <div className="mb-4">
              <h3 className="font-semibold flex items-center gap-2">📈 Advanced Tools</h3>
            </div>
            <div>
              <button 
                onClick={() => setView('trading')}
                className="w-full bg-green-500 text-white px-4 py-2 rounded font-semibold"
              >
                📊 Advanced Trading
              </button>
            </div>
          </div>
        </div>

        <div className="mt-12 space-y-4">
          <div className="p-6 border rounded-lg bg-gradient-to-r from-blue-50 to-green-50">
            <h3 className="font-bold text-lg mb-4 flex items-center gap-2">
              ⚡ Why Trade on ShareHODL?
            </h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div className="text-center">
                <div className="text-2xl mb-2">⚡</div>
                <div className="font-semibold">6-Second Settlement</div>
                <div className="text-gray-600">vs 2-3 days traditional</div>
              </div>
              <div className="text-center">
                <div className="text-2xl mb-2">🌍</div>
                <div className="font-semibold">24/7 Trading</div>
                <div className="text-gray-600">Never closes</div>
              </div>
              <div className="text-center">
                <div className="text-2xl mb-2">💰</div>
                <div className="font-semibold">Low Fees</div>
                <div className="text-gray-600">1000x cheaper</div>
              </div>
              <div className="text-center">
                <div className="text-2xl mb-2">🛡️</div>
                <div className="font-semibold">Safe Trading</div>
                <div className="text-gray-600">Built-in protection</div>
              </div>
            </div>
          </div>
          
          <div className="p-6 border rounded-lg bg-muted/50">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
              <div>
                <h4 className="font-semibold">📈 Want More Trading Features?</h4>
                <p className="text-sm text-gray-600">Access advanced order types, charts, and professional tools for serious traders</p>
              </div>
              <button 
                onClick={() => setView('trading')}
                className="bg-green-500 text-white px-6 py-3 rounded-lg font-semibold w-full sm:w-auto"
              >
                📊 Try Advanced Trading
              </button>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="text-center text-muted-foreground mt-12 pt-8 border-t">
          <p className="mb-2 font-semibold">
            ShareHODL Trading Platform
          </p>
          <p className="text-sm mb-4">
            Fast, secure, and affordable trading for everyone - beginners to professionals.
          </p>
          <div className="flex justify-center items-center gap-6 pt-4 border-t">
            <a 
              href="https://x.com/share_hodl" 
              target="_blank" 
              rel="noopener noreferrer"
              className="text-muted-foreground hover:text-foreground transition-colors flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
              </svg>
              <span className="text-sm">@share_hodl</span>
            </a>
            <a 
              href="https://github.com/x-word-wide/sharehodl" 
              target="_blank" 
              rel="noopener noreferrer"
              className="text-muted-foreground hover:text-foreground transition-colors flex items-center gap-2"
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
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