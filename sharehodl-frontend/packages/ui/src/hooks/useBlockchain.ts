'use client';

import { useState, useCallback, useEffect, useRef } from 'react';

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

// Lazy getters to ensure config is evaluated at runtime (after SSR hydration)
const getRpcUrl = () => getConfig().rpcUrl;
const getRestUrl = () => getConfig().restUrl;

export interface Block {
  height: string;
  hash: string;
  time: string;
  proposer: string;
  txCount: number;
}

export interface Transaction {
  hash: string;
  height: string;
  time: string;
  type: string;
  status: 'success' | 'failed';
  fee: string;
  from?: string;
  to?: string;
  amount?: string;
}

export interface NetworkStatus {
  connected: boolean;
  chainId: string;
  latestBlockHeight: string;
  latestBlockTime: string;
  catching_up: boolean;
  validatorCount: number;
}

export function useBlockchain() {
  const [blocks, setBlocks] = useState<Block[]>([]);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [networkStatus, setNetworkStatus] = useState<NetworkStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  // Fetch network status
  const fetchNetworkStatus = useCallback(async () => {
    try {
      const response = await fetch(`${getRpcUrl()}/status`);
      const data = await response.json();

      if (data.result) {
        const status: NetworkStatus = {
          connected: true,
          chainId: data.result.node_info?.network || 'unknown',
          latestBlockHeight: data.result.sync_info?.latest_block_height || '0',
          latestBlockTime: data.result.sync_info?.latest_block_time || '',
          catching_up: data.result.sync_info?.catching_up || false,
          validatorCount: 0, // Will be fetched separately
        };
        setNetworkStatus(status);
        return status;
      }
    } catch (err) {
      console.error('Error fetching network status:', err);
      setNetworkStatus(prev => prev ? { ...prev, connected: false } : null);
      setError('Failed to connect to blockchain');
    }
    return null;
  }, []);

  // Fetch recent blocks
  const fetchBlocks = useCallback(async (count: number = 10) => {
    try {
      // First get the latest block height
      const statusResponse = await fetch(`${getRpcUrl()}/status`);
      const statusData = await statusResponse.json();
      const latestHeight = parseInt(statusData.result?.sync_info?.latest_block_height || '0');

      if (latestHeight === 0) return [];

      const newBlocks: Block[] = [];

      // Fetch last N blocks
      const startHeight = Math.max(1, latestHeight - count + 1);

      for (let height = latestHeight; height >= startHeight; height--) {
        try {
          const blockResponse = await fetch(`${getRpcUrl()}/block?height=${height}`);
          const blockData = await blockResponse.json();

          if (blockData.result?.block) {
            const block = blockData.result.block;
            newBlocks.push({
              height: block.header.height,
              hash: blockData.result.block_id?.hash || '',
              time: block.header.time,
              proposer: block.header.proposer_address?.substring(0, 12) + '...' || 'Unknown',
              txCount: block.data?.txs?.length || 0,
            });
          }
        } catch (err) {
          console.error(`Error fetching block ${height}:`, err);
        }
      }

      setBlocks(newBlocks);
      return newBlocks;
    } catch (err) {
      console.error('Error fetching blocks:', err);
      setError('Failed to fetch blocks');
      return [];
    }
  }, []);

  // Fetch recent transactions
  const fetchTransactions = useCallback(async (count: number = 10) => {
    try {
      // Search for recent transactions
      const response = await fetch(
        `${getRestUrl()}/cosmos/tx/v1beta1/txs?events=tx.height>0&pagination.limit=${count}&order_by=ORDER_BY_DESC`
      );
      const data = await response.json();

      if (data.tx_responses) {
        const txs: Transaction[] = data.tx_responses.map((tx: any) => {
          // Parse transaction type from messages
          const messages = tx.tx?.body?.messages || [];
          const firstMsg = messages[0] || {};
          const typeUrl = firstMsg['@type'] || '';
          const type = typeUrl.split('.').pop() || 'Unknown';

          // Extract from/to addresses if available
          const from = firstMsg.from_address || firstMsg.sender || firstMsg.delegator_address || '';
          const to = firstMsg.to_address || firstMsg.receiver || firstMsg.validator_address || '';

          // Get amount if available
          const amount = firstMsg.amount?.[0]?.amount
            ? `${(parseFloat(firstMsg.amount[0].amount) / 1000000).toFixed(2)} ${firstMsg.amount[0].denom?.replace('u', '').toUpperCase()}`
            : '';

          return {
            hash: tx.txhash,
            height: tx.height,
            time: tx.timestamp,
            type: type.replace('Msg', ''),
            status: tx.code === 0 ? 'success' : 'failed',
            fee: tx.tx?.auth_info?.fee?.amount?.[0]?.amount || '0',
            from,
            to,
            amount,
          };
        });

        setTransactions(txs);
        return txs;
      }
      return [];
    } catch (err) {
      console.error('Error fetching transactions:', err);
      // Don't set error for this - transactions might just be empty
      return [];
    }
  }, []);

  // Fetch a specific block by height
  const fetchBlock = useCallback(async (height: string): Promise<Block | null> => {
    try {
      const response = await fetch(`${getRpcUrl()}/block?height=${height}`);
      const data = await response.json();

      if (data.result?.block) {
        const block = data.result.block;
        return {
          height: block.header.height,
          hash: data.result.block_id?.hash || '',
          time: block.header.time,
          proposer: block.header.proposer_address || 'Unknown',
          txCount: block.data?.txs?.length || 0,
        };
      }
      return null;
    } catch (err) {
      console.error('Error fetching block:', err);
      return null;
    }
  }, []);

  // Fetch a specific transaction by hash
  const fetchTransaction = useCallback(async (hash: string): Promise<Transaction | null> => {
    try {
      const response = await fetch(`${getRestUrl()}/cosmos/tx/v1beta1/txs/${hash}`);
      const data = await response.json();

      if (data.tx_response) {
        const tx = data.tx_response;
        const messages = tx.tx?.body?.messages || [];
        const firstMsg = messages[0] || {};
        const typeUrl = firstMsg['@type'] || '';

        return {
          hash: tx.txhash,
          height: tx.height,
          time: tx.timestamp,
          type: typeUrl.split('.').pop()?.replace('Msg', '') || 'Unknown',
          status: tx.code === 0 ? 'success' : 'failed',
          fee: tx.tx?.auth_info?.fee?.amount?.[0]?.amount || '0',
          from: firstMsg.from_address || firstMsg.sender || '',
          to: firstMsg.to_address || firstMsg.receiver || '',
          amount: firstMsg.amount?.[0]?.amount || '',
        };
      }
      return null;
    } catch (err) {
      console.error('Error fetching transaction:', err);
      return null;
    }
  }, []);

  // Search for blocks/transactions/addresses
  const search = useCallback(async (query: string): Promise<{
    type: 'block' | 'transaction' | 'address' | 'unknown';
    data: any;
  }> => {
    // Check if it's a block height (numeric)
    if (/^\d+$/.test(query)) {
      const block = await fetchBlock(query);
      if (block) {
        return { type: 'block', data: block };
      }
    }

    // Check if it's a transaction hash (64 hex chars)
    if (/^[A-Fa-f0-9]{64}$/.test(query)) {
      const tx = await fetchTransaction(query);
      if (tx) {
        return { type: 'transaction', data: tx };
      }
    }

    // Check if it's an address (starts with hodl)
    if (query.startsWith('hodl')) {
      // Return address type - can add account lookup later
      return { type: 'address', data: { address: query } };
    }

    return { type: 'unknown', data: null };
  }, [fetchBlock, fetchTransaction]);

  // WebSocket for real-time updates
  const connectWebSocket = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    try {
      const wsUrl = getRpcUrl().replace('http', 'ws') + '/websocket';
      wsRef.current = new WebSocket(wsUrl);

      wsRef.current.onopen = () => {
        console.log('WebSocket connected');
        // Subscribe to new blocks
        wsRef.current?.send(JSON.stringify({
          jsonrpc: '2.0',
          method: 'subscribe',
          id: 'blocks',
          params: { query: "tm.event='NewBlock'" }
        }));
      };

      wsRef.current.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.result?.data?.value?.block) {
            // New block received - refresh data
            fetchBlocks(10);
            fetchTransactions(10);
            fetchNetworkStatus();
          }
        } catch (err) {
          console.error('Error parsing WebSocket message:', err);
        }
      };

      wsRef.current.onclose = () => {
        console.log('WebSocket disconnected, reconnecting...');
        setTimeout(connectWebSocket, 5000);
      };

      wsRef.current.onerror = (err) => {
        console.error('WebSocket error:', err);
      };
    } catch (err) {
      console.error('Error connecting WebSocket:', err);
    }
  }, [fetchBlocks, fetchTransactions, fetchNetworkStatus]);

  // Initial data fetch
  useEffect(() => {
    const init = async () => {
      setLoading(true);
      await fetchNetworkStatus();
      await fetchBlocks(10);
      await fetchTransactions(10);
      setLoading(false);
      connectWebSocket();
    };

    init();

    // Cleanup WebSocket on unmount
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  // Refresh all data
  const refresh = useCallback(async () => {
    setLoading(true);
    await Promise.all([
      fetchNetworkStatus(),
      fetchBlocks(10),
      fetchTransactions(10),
    ]);
    setLoading(false);
  }, [fetchNetworkStatus, fetchBlocks, fetchTransactions]);

  return {
    blocks,
    transactions,
    networkStatus,
    loading,
    error,
    fetchBlocks,
    fetchTransactions,
    fetchBlock,
    fetchTransaction,
    fetchNetworkStatus,
    search,
    refresh,
  };
}
