'use client';

import { useState, useCallback, useEffect } from 'react';

// Auto-detect environment based on hostname
const getConfig = () => {
  // Allow explicit override via window config
  if (typeof window !== 'undefined' && (window as any).__SHAREHODL_CONFIG__) {
    return (window as any).__SHAREHODL_CONFIG__;
  }

  // Auto-detect production environment
  if (typeof window !== 'undefined') {
    const hostname = window.location.hostname;
    // Production: use sharehodl.com endpoints
    if (hostname.includes('sharehodl.com') || hostname.includes('sharehodl')) {
      return {
        rpcUrl: 'https://rpc.sharehodl.com',
        restUrl: 'https://api.sharehodl.com',
      };
    }
  }

  // Development: use localhost
  return {
    rpcUrl: 'http://localhost:26657',
    restUrl: 'http://localhost:1317',
  };
};

// Chain configuration for ShareHODL
export const SHAREHODL_CHAIN_CONFIG = {
  chainId: 'sharehodl-1',
  chainName: 'ShareHODL',
  rpc: getConfig().rpcUrl,
  rest: getConfig().restUrl,
  bip44: {
    coinType: 118,
  },
  bech32Config: {
    bech32PrefixAccAddr: 'hodl',
    bech32PrefixAccPub: 'hodlpub',
    bech32PrefixValAddr: 'hodlvaloper',
    bech32PrefixValPub: 'hodlvaloperpub',
    bech32PrefixConsAddr: 'hodlvalcons',
    bech32PrefixConsPub: 'hodlvalconspub',
  },
  currencies: [
    {
      coinDenom: 'HODL',
      coinMinimalDenom: 'uhodl',
      coinDecimals: 6,
    },
    {
      coinDenom: 'STAKE',
      coinMinimalDenom: 'stake',
      coinDecimals: 6,
    },
  ],
  feeCurrencies: [
    {
      coinDenom: 'HODL',
      coinMinimalDenom: 'uhodl',
      coinDecimals: 6,
      gasPriceStep: {
        low: 0.01,
        average: 0.025,
        high: 0.04,
      },
    },
  ],
  stakeCurrency: {
    coinDenom: 'STAKE',
    coinMinimalDenom: 'stake',
    coinDecimals: 6,
  },
};

export interface WalletBalance {
  denom: string;
  amount: string;
  displayAmount: string;
  symbol: string;
}

export interface WalletState {
  connected: boolean;
  connecting: boolean;
  address: string | null;
  balances: WalletBalance[];
  error: string | null;
}

// Keplr window type
declare global {
  interface Window {
    keplr?: {
      enable: (chainId: string) => Promise<void>;
      getOfflineSigner: (chainId: string) => any;
      getKey: (chainId: string) => Promise<{
        name: string;
        algo: string;
        pubKey: Uint8Array;
        address: Uint8Array;
        bech32Address: string;
      }>;
      experimentalSuggestChain: (chainInfo: any) => Promise<void>;
    };
    getOfflineSigner?: (chainId: string) => any;
  }
}

export function useWallet() {
  const [state, setState] = useState<WalletState>({
    connected: false,
    connecting: false,
    address: null,
    balances: [],
    error: null,
  });

  // Check if Keplr is installed
  const isKeplrInstalled = useCallback(() => {
    return typeof window !== 'undefined' && window.keplr !== undefined;
  }, []);

  // Suggest chain to Keplr (for custom chains like ShareHODL)
  const suggestChain = useCallback(async () => {
    if (!window.keplr) return false;

    try {
      await window.keplr.experimentalSuggestChain({
        chainId: SHAREHODL_CHAIN_CONFIG.chainId,
        chainName: SHAREHODL_CHAIN_CONFIG.chainName,
        rpc: SHAREHODL_CHAIN_CONFIG.rpc,
        rest: SHAREHODL_CHAIN_CONFIG.rest,
        bip44: SHAREHODL_CHAIN_CONFIG.bip44,
        bech32Config: SHAREHODL_CHAIN_CONFIG.bech32Config,
        currencies: SHAREHODL_CHAIN_CONFIG.currencies,
        feeCurrencies: SHAREHODL_CHAIN_CONFIG.feeCurrencies,
        stakeCurrency: SHAREHODL_CHAIN_CONFIG.stakeCurrency,
      });
      return true;
    } catch (error) {
      console.error('Error suggesting chain:', error);
      return false;
    }
  }, []);

  // Fetch balances from chain
  const fetchBalances = useCallback(async (address: string): Promise<WalletBalance[]> => {
    try {
      const response = await fetch(
        `${SHAREHODL_CHAIN_CONFIG.rest}/cosmos/bank/v1beta1/balances/${address}`
      );
      const data = await response.json();

      if (!data.balances) return [];

      return data.balances.map((balance: { denom: string; amount: string }) => {
        const currency = SHAREHODL_CHAIN_CONFIG.currencies.find(
          c => c.coinMinimalDenom === balance.denom
        );
        const decimals = currency?.coinDecimals || 6;
        const displayAmount = (parseFloat(balance.amount) / Math.pow(10, decimals)).toFixed(2);

        return {
          denom: balance.denom,
          amount: balance.amount,
          displayAmount,
          symbol: currency?.coinDenom || balance.denom.toUpperCase(),
        };
      });
    } catch (error) {
      console.error('Error fetching balances:', error);
      return [];
    }
  }, []);

  // Connect wallet
  const connect = useCallback(async () => {
    if (!isKeplrInstalled()) {
      setState(prev => ({
        ...prev,
        error: 'Keplr wallet not installed. Please install Keplr extension.',
      }));
      // Open Keplr website
      window.open('https://www.keplr.app/', '_blank');
      return;
    }

    setState(prev => ({ ...prev, connecting: true, error: null }));

    try {
      // First, suggest the chain
      await suggestChain();

      // Enable the chain
      await window.keplr!.enable(SHAREHODL_CHAIN_CONFIG.chainId);

      // Get the key/address
      const key = await window.keplr!.getKey(SHAREHODL_CHAIN_CONFIG.chainId);
      const address = key.bech32Address;

      // Fetch balances
      const balances = await fetchBalances(address);

      setState({
        connected: true,
        connecting: false,
        address,
        balances,
        error: null,
      });

      // Save connection state
      localStorage.setItem('sharehodl_wallet_connected', 'true');
    } catch (error) {
      console.error('Error connecting wallet:', error);
      setState(prev => ({
        ...prev,
        connecting: false,
        error: error instanceof Error ? error.message : 'Failed to connect wallet',
      }));
    }
  }, [isKeplrInstalled, suggestChain, fetchBalances]);

  // Disconnect wallet
  const disconnect = useCallback(() => {
    setState({
      connected: false,
      connecting: false,
      address: null,
      balances: [],
      error: null,
    });
    localStorage.removeItem('sharehodl_wallet_connected');
  }, []);

  // Refresh balances
  const refreshBalances = useCallback(async () => {
    if (!state.address) return;
    const balances = await fetchBalances(state.address);
    setState(prev => ({ ...prev, balances }));
  }, [state.address, fetchBalances]);

  // Get signer for transactions
  const getSigner = useCallback(async () => {
    if (!window.keplr || !state.connected) {
      throw new Error('Wallet not connected');
    }
    return window.keplr.getOfflineSigner(SHAREHODL_CHAIN_CONFIG.chainId);
  }, [state.connected]);

  // Auto-reconnect on page load
  useEffect(() => {
    const tryAutoReconnect = async () => {
      const wasConnected = localStorage.getItem('sharehodl_wallet_connected');
      if (wasConnected !== 'true') return;

      // Wait for Keplr to be injected (can take a moment)
      let attempts = 0;
      while (!window.keplr && attempts < 10) {
        await new Promise(resolve => setTimeout(resolve, 100));
        attempts++;
      }

      if (window.keplr) {
        connect();
      }
    };

    tryAutoReconnect();
  }, []);

  // Listen for Keplr account changes
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const handleAccountChange = () => {
      if (state.connected) {
        connect();
      }
    };

    window.addEventListener('keplr_keystorechange', handleAccountChange);
    return () => {
      window.removeEventListener('keplr_keystorechange', handleAccountChange);
    };
  }, [state.connected, connect]);

  return {
    ...state,
    connect,
    disconnect,
    refreshBalances,
    getSigner,
    isKeplrInstalled,
    chainConfig: SHAREHODL_CHAIN_CONFIG,
  };
}
