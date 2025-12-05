'use client';

import React, { useState, useEffect } from 'react';
import { Navigation } from "@repo/ui";
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
        <Navigation />
        <div className="flex items-center justify-between p-4 border-b">
          <h1 className="text-2xl font-bold flex items-center gap-2">
            📊 ShareDEX Professional Trading
          </h1>
          <button 
            onClick={() => setView('swap')}
            className="px-4 py-2 bg-blue-500 text-white rounded"
          >
            Switch to Atomic Swaps
          </button>
        </div>
        <AdvancedTrading symbol="APPLE/HODL" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      <main className="container mx-auto px-4 py-8">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold mb-4 flex items-center justify-center gap-3">
            <span className="text-2xl">⇄</span>
            ShareDEX Trading Platform
          </h1>
          <p className="text-muted-foreground text-lg max-w-2xl mx-auto">
            Professional equity trading with institutional-grade features and atomic cross-asset swaps.
          </p>
          <div className="mt-6 flex justify-center gap-4">
            <button 
              onClick={() => setView('swap')}
              className={`px-6 py-2 rounded-lg font-semibold ${
                view === 'swap' 
                  ? 'bg-blue-500 text-white' 
                  : 'border border-blue-500 text-blue-500'
              }`}
            >
              Atomic Swaps
            </button>
            <button 
              onClick={() => setView('trading')}
              className={`px-6 py-2 rounded-lg font-semibold ${
                view === 'trading' 
                  ? 'bg-blue-500 text-white' 
                  : 'border border-blue-500 text-blue-500'
              }`}
            >
              Professional Trading
            </button>
          </div>
        </div>

        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <div className="border rounded-lg p-6">
              <h3 className="font-semibold mb-4">Atomic Swap</h3>
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
                    {loading ? 'Connecting...' : 'Connect Wallet'}
                  </button>
                ) : (
                  <button 
                    onClick={executeSwap}
                    disabled={loading || !fromAmount || fromAsset === toAsset}
                    className="w-full bg-blue-500 text-white py-3 rounded font-semibold disabled:opacity-50"
                  >
                    {loading ? 'Swapping...' : 'Execute Atomic Swap'}
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
                        🟢 Connected to ShareHODL Network<br/>
                        Chain: {networkStatus.node_info?.network}<br/>
                        Height: {networkStatus.sync_info?.latest_block_height}
                      </>
                    ) : (
                      '🟡 Blockchain offline - Using mock data for demo'
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
              <h3 className="font-semibold flex items-center gap-2">Professional Features</h3>
            </div>
            <div>
              <button 
                onClick={() => setView('trading')}
                className="w-full bg-green-500 text-white px-4 py-2 rounded font-semibold"
              >
                FOK/IOC Orders
              </button>
            </div>
          </div>
        </div>

        <div className="mt-12 space-y-4">
          <div className="p-6 border rounded-lg bg-gradient-to-r from-blue-50 to-green-50">
            <h3 className="font-bold text-lg mb-4 flex items-center gap-2">
              🚀 ShareHODL Trading Advantages
            </h3>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div className="text-center">
                <div className="text-2xl mb-2">⚡</div>
                <div className="font-semibold">6-Second Settlement</div>
                <div className="text-gray-600">vs T+2 traditional</div>
              </div>
              <div className="text-center">
                <div className="text-2xl mb-2">🕐</div>
                <div className="font-semibold">24/7 Trading</div>
                <div className="text-gray-600">vs 8hr/day traditional</div>
              </div>
              <div className="text-center">
                <div className="text-2xl mb-2">💰</div>
                <div className="font-semibold">$0.005 Fees</div>
                <div className="text-gray-600">vs $5-15+ traditional</div>
              </div>
              <div className="text-center">
                <div className="text-2xl mb-2">🛡️</div>
                <div className="font-semibold">Circuit Breakers</div>
                <div className="text-gray-600">Professional safeguards</div>
              </div>
            </div>
          </div>
          
          <div className="p-6 border rounded-lg bg-muted/50">
            <div className="flex items-center justify-between">
              <div>
                <h4 className="font-semibold">Ready for Professional Trading?</h4>
                <p className="text-sm text-gray-600">Access FOK/IOC orders, circuit breakers, and institutional features</p>
              </div>
              <button 
                onClick={() => setView('trading')}
                className="bg-green-500 text-white px-6 py-3 rounded-lg font-semibold"
              >
                Launch Pro Trading
              </button>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}