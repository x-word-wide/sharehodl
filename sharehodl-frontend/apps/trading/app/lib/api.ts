// API client for ShareHODL blockchain
const BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:1317';
const RPC_URL = process.env.NEXT_PUBLIC_RPC_URL || 'http://localhost:26657';

export interface Balance {
  denom: string;
  amount: string;
}

export interface Market {
  base_denom: string;
  quote_denom: string;
  last_price: string;
  volume_24h: string;
  price_change_24h: string;
  high_24h: string;
  low_24h: string;
}

export interface SwapResult {
  success: boolean;
  txHash?: string;
  error?: string;
}

export class ShareHODLAPI {
  // Get account balances
  static async getBalances(address: string): Promise<Balance[]> {
    try {
      const response = await fetch(`${BASE_URL}/cosmos/bank/v1beta1/balances/${address}`);
      const data = await response.json();
      return data.balances || [];
    } catch (error) {
      console.error('Error fetching balances:', error);
      return [];
    }
  }

  // Get specific balance for a denom
  static async getBalance(address: string, denom: string): Promise<string> {
    try {
      const response = await fetch(`${BASE_URL}/cosmos/bank/v1beta1/balances/${address}/by_denom?denom=${denom}`);
      const data = await response.json();
      return data.balance?.amount || '0';
    } catch (error) {
      console.error('Error fetching balance:', error);
      return '0';
    }
  }

  // Get market data
  static async getMarkets(): Promise<Market[]> {
    try {
      // For now, return mock data - will connect to real API once blockchain is running
      return [
        {
          base_denom: 'uapple',
          quote_denom: 'uhodl',
          last_price: '185.25',
          volume_24h: '125890',
          price_change_24h: '2.5',
          high_24h: '187.50',
          low_24h: '180.75'
        },
        {
          base_denom: 'utsla',
          quote_denom: 'uhodl', 
          last_price: '245.80',
          volume_24h: '89340',
          price_change_24h: '-1.2',
          high_24h: '252.10',
          low_24h: '243.50'
        }
      ];
    } catch (error) {
      console.error('Error fetching markets:', error);
      return [];
    }
  }

  // Calculate swap amount with slippage
  static calculateSwapAmount(
    fromAmount: string,
    fromDenom: string,
    toDenom: string,
    slippage: number = 0.03
  ): { toAmount: string; rate: string; slippageAmount: string } {
    const mockPrices: { [key: string]: number } = {
      'uhodl': 1.00,
      'uapple': 185.25,
      'utsla': 245.80,
      'ugoogl': 2750.00,
      'umsft': 385.60
    };

    const fromPrice = mockPrices[fromDenom] || 1;
    const toPrice = mockPrices[toDenom] || 1;
    
    const rate = fromPrice / toPrice;
    const baseAmount = parseFloat(fromAmount) * rate;
    const slippageAmount = baseAmount * slippage;
    const finalAmount = baseAmount - slippageAmount;

    return {
      toAmount: finalAmount.toFixed(6),
      rate: rate.toFixed(6),
      slippageAmount: slippageAmount.toFixed(6)
    };
  }

  // Execute atomic swap (mock implementation)
  static async executeSwap(
    fromDenom: string,
    toDenom: string,
    amount: string,
    slippage: number = 0.03
  ): Promise<SwapResult> {
    try {
      // Mock implementation - will integrate with CosmJS when blockchain is ready
      console.log('Executing swap:', { fromDenom, toDenom, amount, slippage });
      
      // Simulate network delay
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      // Mock success
      return {
        success: true,
        txHash: `0x${Math.random().toString(16).substr(2, 64)}`
      };
    } catch (error) {
      console.error('Error executing swap:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error'
      };
    }
  }

  // Get network status
  static async getNetworkStatus() {
    try {
      const response = await fetch(`${RPC_URL}/status`);
      const data = await response.json();
      return data.result;
    } catch (error) {
      console.error('Error fetching network status:', error);
      return null;
    }
  }

  // Get latest block
  static async getLatestBlock() {
    try {
      const response = await fetch(`${BASE_URL}/cosmos/base/tendermint/v1beta1/blocks/latest`);
      const data = await response.json();
      return data.block;
    } catch (error) {
      console.error('Error fetching latest block:', error);
      return null;
    }
  }

  // Format amount for display
  static formatAmount(amount: string, decimals: number = 6): string {
    const num = parseFloat(amount) / Math.pow(10, decimals);
    return num.toLocaleString(undefined, { 
      minimumFractionDigits: 2, 
      maximumFractionDigits: 6 
    });
  }

  // Convert display amount to chain format
  static toChainAmount(amount: string, decimals: number = 6): string {
    const num = parseFloat(amount) * Math.pow(10, decimals);
    return Math.floor(num).toString();
  }
}

// WebSocket connection for real-time updates
export class ShareHODLWebSocket {
  private ws: WebSocket | null = null;
  private listeners: { [event: string]: ((data: any) => void)[] } = {};

  connect() {
    try {
      this.ws = new WebSocket(RPC_URL.replace('http', 'ws') + '/websocket');
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        // Subscribe to new blocks
        this.subscribe('tm.event=\'NewBlock\'');
        // Subscribe to new transactions
        this.subscribe('tm.event=\'Tx\'');
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.emit(data.id || 'message', data);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      this.ws.onclose = () => {
        console.log('WebSocket disconnected');
        // Attempt reconnection after 5 seconds
        setTimeout(() => this.connect(), 5000);
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };
    } catch (error) {
      console.error('Error connecting WebSocket:', error);
    }
  }

  subscribe(query: string, id?: string) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        jsonrpc: "2.0",
        method: "subscribe",
        id: id || query,
        params: { query }
      }));
    }
  }

  on(event: string, callback: (data: any) => void) {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(callback);
  }

  private emit(event: string, data: any) {
    if (this.listeners[event]) {
      this.listeners[event].forEach(callback => callback(data));
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}