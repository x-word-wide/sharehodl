'use client';

import React from 'react';
import { useWallet } from '../hooks/useWallet';

interface WalletButtonProps {
  className?: string;
  showBalance?: boolean;
}

export function WalletButton({ className = '', showBalance = false }: WalletButtonProps) {
  const { connected, connecting, address, balances, connect, disconnect, error } = useWallet();

  // Format address for display (show first and last 6 characters)
  const formatAddress = (addr: string) => {
    if (!addr) return '';
    return `${addr.slice(0, 10)}...${addr.slice(-6)}`;
  };

  // Get HODL balance
  const hodlBalance = balances.find(b => b.denom === 'uhodl');

  if (connecting) {
    return (
      <button
        className={`px-4 py-2 bg-gray-600 text-white rounded-lg cursor-wait flex items-center gap-2 ${className}`}
        disabled
      >
        <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
            fill="none"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
        Connecting...
      </button>
    );
  }

  if (connected && address) {
    return (
      <div className={`flex items-center gap-2 ${className}`}>
        {showBalance && hodlBalance && (
          <span className="text-sm text-gray-300 bg-gray-800 px-3 py-2 rounded-lg">
            {hodlBalance.displayAmount} HODL
          </span>
        )}
        <div className="relative group">
          <button
            className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg flex items-center gap-2 transition-colors"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            {formatAddress(address)}
          </button>
          <div className="absolute right-0 top-full mt-2 w-48 bg-gray-800 rounded-lg shadow-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-50">
            <div className="p-3 border-b border-gray-700">
              <p className="text-xs text-gray-400">Connected</p>
              <p className="text-sm text-white font-mono break-all">{address}</p>
            </div>
            {balances.length > 0 && (
              <div className="p-3 border-b border-gray-700">
                <p className="text-xs text-gray-400 mb-1">Balances</p>
                {balances.map((balance) => (
                  <div key={balance.denom} className="flex justify-between text-sm">
                    <span className="text-gray-300">{balance.symbol}</span>
                    <span className="text-white">{balance.displayAmount}</span>
                  </div>
                ))}
              </div>
            )}
            <button
              onClick={disconnect}
              className="w-full p-3 text-left text-red-400 hover:bg-gray-700 rounded-b-lg transition-colors"
            >
              Disconnect
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-end">
      <button
        onClick={connect}
        className={`px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg flex items-center gap-2 transition-colors ${className}`}
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
        Connect Wallet
      </button>
      {error && (
        <p className="text-xs text-red-400 mt-1">{error}</p>
      )}
    </div>
  );
}
