/**
 * Wallet State Management with Zustand
 *
 * Manages wallet state including accounts, balances, and secure storage
 * Supports multiple chains and tokens like Trust Wallet
 */

import { create } from 'zustand';
import {
  Chain,
  WalletAccount,
  AssetHolding,
  Token,
  CHAIN_CONFIGS,
  TOKENS,
  TokenType
} from '../types';
import {
  generateMnemonic,
  validateMnemonic,
  getAddressFromMnemonic,
  encryptData,
  decryptData
} from '../utils/crypto';
import {
  isLockedOut,
  getLockoutRemaining,
  recordFailedAttempt,
  resetFailedAttempts,
  getRemainingAttempts,
  updateLastActivity,
  shouldAutoLock,
  clearSecurityData,
  formatLockoutTime,
  getSecurityState,
  type SecurityState
} from '../utils/security';

// Storage keys
const STORAGE_KEYS = {
  ENCRYPTED_MNEMONIC: 'sh_encrypted_mnemonic',
  WALLET_INITIALIZED: 'sh_wallet_init',
  ACCOUNTS: 'sh_accounts',
  ASSETS: 'sh_assets',
  ENABLED_TOKENS: 'sh_enabled_tokens'
};

interface WalletStore {
  // State
  isInitialized: boolean;
  isLocked: boolean;
  isLoading: boolean;
  accounts: WalletAccount[];
  assets: AssetHolding[];
  enabledTokenIds: string[];
  totalBalanceUsd: number;
  error: string | null;

  // Security state
  securityState: SecurityState;
  remainingAttempts: number;

  // Actions
  initialize: () => Promise<void>;
  createWallet: (pin: string) => Promise<string>;
  completeWalletSetup: () => void;  // Call after seed phrase verification
  importWallet: (mnemonic: string, pin: string) => Promise<void>;
  unlockWallet: (pin: string) => Promise<void>;
  lockWallet: () => void;
  refreshBalances: () => Promise<void>;
  clearError: () => void;
  resetWallet: () => void;
  updateActivity: () => void;
  checkAutoLock: () => boolean;
  refreshSecurityState: () => void;

  // Token management
  enableToken: (tokenId: string) => void;
  disableToken: (tokenId: string) => void;
  getAssetByTokenId: (tokenId: string) => AssetHolding | undefined;
}

// Supported chains for initial account generation
const SUPPORTED_CHAINS: Chain[] = [
  Chain.SHAREHODL,
  Chain.ETHEREUM,
  Chain.BITCOIN,
  Chain.POLYGON,
  Chain.BNB,
  Chain.ARBITRUM,
  Chain.BASE,
  Chain.AVALANCHE,
  Chain.COSMOS,
  Chain.SOLANA,
];

// Default enabled tokens (shown by default in portfolio)
const DEFAULT_ENABLED_TOKENS: string[] = [
  'btc',           // Bitcoin
  'eth',           // Ethereum
  'matic',         // Polygon
  'bnb',           // BNB
  'usdt-eth',      // USDT on Ethereum
  'usdc-eth',      // USDC on Ethereum
  'usdt-polygon',  // USDT on Polygon
  'usdc-polygon',  // USDC on Polygon
  'usdt-bsc',      // USDT on BNB Chain
  'usdc-bsc',      // USDC on BNB Chain
  'hodl',          // ShareHODL native
];

export const useWalletStore = create<WalletStore>((set, get) => ({
  // Initial state
  isInitialized: false,
  isLocked: true,
  isLoading: false,
  accounts: [],
  assets: [],
  enabledTokenIds: DEFAULT_ENABLED_TOKENS,
  totalBalanceUsd: 0,
  error: null,

  // Security state
  securityState: getSecurityState(),
  remainingAttempts: getRemainingAttempts(),

  // Initialize - check if wallet exists
  initialize: async () => {
    try {
      const initialized = localStorage.getItem(STORAGE_KEYS.WALLET_INITIALIZED);
      set({ isInitialized: initialized === 'true' });
    } catch (error) {
      console.error('Failed to initialize wallet:', error);
    }
  },

  // Create new wallet
  createWallet: async (pin: string) => {
    set({ isLoading: true, error: null });

    try {
      // Generate mnemonic
      const mnemonic = generateMnemonic(256); // 24 words

      // Encrypt and store mnemonic
      const encrypted = await encryptData(mnemonic, pin);
      localStorage.setItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC, encrypted);
      localStorage.setItem(STORAGE_KEYS.WALLET_INITIALIZED, 'true');

      // Generate accounts for all supported chains
      const accounts = generateAccounts(mnemonic);
      const enabledTokenIds = DEFAULT_ENABLED_TOKENS;

      // Generate initial assets from enabled tokens
      const assets = generateAssets(accounts, enabledTokenIds);

      // Store data (public info only)
      localStorage.setItem(STORAGE_KEYS.ACCOUNTS, JSON.stringify(accounts));
      localStorage.setItem(STORAGE_KEYS.ASSETS, JSON.stringify(assets));
      localStorage.setItem(STORAGE_KEYS.ENABLED_TOKENS, JSON.stringify(enabledTokenIds));

      set({
        isInitialized: true,
        isLocked: true,  // Keep locked until seed phrase backup is verified
        isLoading: false,
        accounts,
        assets,
        enabledTokenIds
      });

      return mnemonic;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to create wallet';
      set({ isLoading: false, error: message });
      throw error;
    }
  },

  // Complete wallet setup after seed phrase verification
  completeWalletSetup: () => {
    updateLastActivity();
    set({ isLocked: false });
  },

  // Import existing wallet
  importWallet: async (mnemonic: string, pin: string) => {
    set({ isLoading: true, error: null });

    try {
      // Validate mnemonic
      if (!validateMnemonic(mnemonic.trim())) {
        throw new Error('Invalid mnemonic phrase');
      }

      const cleanMnemonic = mnemonic.trim().toLowerCase();

      // Encrypt and store mnemonic
      const encrypted = await encryptData(cleanMnemonic, pin);
      localStorage.setItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC, encrypted);
      localStorage.setItem(STORAGE_KEYS.WALLET_INITIALIZED, 'true');

      // Generate accounts
      const accounts = generateAccounts(cleanMnemonic);
      const enabledTokenIds = DEFAULT_ENABLED_TOKENS;
      const assets = generateAssets(accounts, enabledTokenIds);

      localStorage.setItem(STORAGE_KEYS.ACCOUNTS, JSON.stringify(accounts));
      localStorage.setItem(STORAGE_KEYS.ASSETS, JSON.stringify(assets));
      localStorage.setItem(STORAGE_KEYS.ENABLED_TOKENS, JSON.stringify(enabledTokenIds));

      set({
        isInitialized: true,
        isLocked: false,
        isLoading: false,
        accounts,
        assets,
        enabledTokenIds
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to import wallet';
      set({ isLoading: false, error: message });
      throw error;
    }
  },

  // Unlock wallet with PIN (with brute force protection)
  unlockWallet: async (pin: string) => {
    // Check if locked out
    if (isLockedOut()) {
      const remaining = getLockoutRemaining();
      const message = `Too many failed attempts. Try again in ${formatLockoutTime(remaining)}`;
      set({
        error: message,
        securityState: getSecurityState(),
        remainingAttempts: getRemainingAttempts()
      });
      throw new Error(message);
    }

    set({ isLoading: true, error: null });

    try {
      const encrypted = localStorage.getItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC);
      if (!encrypted) {
        throw new Error('No wallet found');
      }

      // Try to decrypt to verify PIN
      await decryptData(encrypted, pin);

      // Success - reset failed attempts
      resetFailedAttempts();
      updateLastActivity();

      // Load accounts and assets
      const accountsJson = localStorage.getItem(STORAGE_KEYS.ACCOUNTS);
      const assetsJson = localStorage.getItem(STORAGE_KEYS.ASSETS);
      const enabledJson = localStorage.getItem(STORAGE_KEYS.ENABLED_TOKENS);

      let accounts = accountsJson ? JSON.parse(accountsJson) : [];
      const enabledTokenIds = enabledJson ? JSON.parse(enabledJson) : DEFAULT_ENABLED_TOKENS;

      // MIGRATION: If no assets exist, generate them from accounts
      // This handles wallets created before multi-asset support
      let assets = assetsJson ? JSON.parse(assetsJson) : [];
      if (assets.length === 0 && accounts.length > 0) {
        // Regenerate accounts to include any new chains
        const mnemonic = await decryptData(encrypted, pin);
        accounts = generateAccounts(mnemonic);
        assets = generateAssets(accounts, enabledTokenIds);

        // Save the migrated data
        localStorage.setItem(STORAGE_KEYS.ACCOUNTS, JSON.stringify(accounts));
        localStorage.setItem(STORAGE_KEYS.ASSETS, JSON.stringify(assets));
        localStorage.setItem(STORAGE_KEYS.ENABLED_TOKENS, JSON.stringify(enabledTokenIds));
      }

      set({
        isLocked: false,
        isLoading: false,
        accounts,
        assets,
        enabledTokenIds,
        securityState: getSecurityState(),
        remainingAttempts: getRemainingAttempts()
      });

      // Refresh balances in background
      get().refreshBalances();
    } catch (error) {
      // Record failed attempt
      const newSecurityState = recordFailedAttempt();
      const remaining = getRemainingAttempts();

      let errorMessage = 'Invalid PIN';
      if (newSecurityState.isLocked) {
        errorMessage = `Too many failed attempts. Try again in ${formatLockoutTime(Math.ceil(newSecurityState.lockoutRemainingMs / 1000))}`;
      } else if (remaining <= 2) {
        errorMessage = `Invalid PIN. ${remaining} attempt${remaining !== 1 ? 's' : ''} remaining`;
      }

      set({
        isLoading: false,
        error: errorMessage,
        securityState: newSecurityState,
        remainingAttempts: remaining
      });
      throw new Error(errorMessage);
    }
  },

  // Lock wallet
  lockWallet: () => {
    set({ isLocked: true });
  },

  // Refresh all balances
  refreshBalances: async () => {
    const { accounts, enabledTokenIds } = get();

    try {
      // Update account balances
      const updatedAccounts = await Promise.all(
        accounts.map(async (account) => {
          try {
            const balance = await fetchBalance(account.chain, account.address);
            const usdPrice = await fetchPrice(account.chain);
            const balanceNum = parseFloat(balance);
            const balanceUsd = (balanceNum * usdPrice).toFixed(2);

            return {
              ...account,
              balance,
              balanceUsd
            };
          } catch {
            return account;
          }
        })
      );

      // Generate updated assets with prices
      const updatedAssets = await generateAssetsWithPrices(updatedAccounts, enabledTokenIds);

      const totalUsd = updatedAssets.reduce((sum, asset) => {
        return sum + parseFloat(asset.balanceUsd || '0');
      }, 0);

      set({
        accounts: updatedAccounts,
        assets: updatedAssets,
        totalBalanceUsd: totalUsd
      });

      // Save updated data
      localStorage.setItem(STORAGE_KEYS.ACCOUNTS, JSON.stringify(updatedAccounts));
      localStorage.setItem(STORAGE_KEYS.ASSETS, JSON.stringify(updatedAssets));
    } catch (error) {
      console.error('Failed to refresh balances:', error);
    }
  },

  // Clear error
  clearError: () => set({ error: null }),

  // Reset wallet (dangerous!)
  resetWallet: () => {
    localStorage.removeItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC);
    localStorage.removeItem(STORAGE_KEYS.WALLET_INITIALIZED);
    localStorage.removeItem(STORAGE_KEYS.ACCOUNTS);
    localStorage.removeItem(STORAGE_KEYS.ASSETS);
    localStorage.removeItem(STORAGE_KEYS.ENABLED_TOKENS);

    // Also clear security data
    clearSecurityData();

    set({
      isInitialized: false,
      isLocked: true,
      accounts: [],
      assets: [],
      enabledTokenIds: DEFAULT_ENABLED_TOKENS,
      totalBalanceUsd: 0,
      error: null,
      securityState: getSecurityState(),
      remainingAttempts: getRemainingAttempts()
    });
  },

  // Update last activity timestamp (call on user interaction)
  updateActivity: () => {
    updateLastActivity();
  },

  // Check if auto-lock should trigger
  checkAutoLock: () => {
    if (shouldAutoLock() && !get().isLocked) {
      get().lockWallet();
      return true;
    }
    return false;
  },

  // Refresh security state from storage
  refreshSecurityState: () => {
    set({
      securityState: getSecurityState(),
      remainingAttempts: getRemainingAttempts()
    });
  },

  // Enable a token (add to portfolio)
  enableToken: (tokenId: string) => {
    const { enabledTokenIds, accounts } = get();
    if (enabledTokenIds.includes(tokenId)) return;

    const newEnabledTokenIds = [...enabledTokenIds, tokenId];
    const assets = generateAssets(accounts, newEnabledTokenIds);

    localStorage.setItem(STORAGE_KEYS.ENABLED_TOKENS, JSON.stringify(newEnabledTokenIds));
    localStorage.setItem(STORAGE_KEYS.ASSETS, JSON.stringify(assets));

    set({ enabledTokenIds: newEnabledTokenIds, assets });

    // Refresh to get balances
    get().refreshBalances();
  },

  // Disable a token (remove from portfolio)
  disableToken: (tokenId: string) => {
    const { enabledTokenIds, accounts } = get();
    const newEnabledTokenIds = enabledTokenIds.filter(id => id !== tokenId);
    const assets = generateAssets(accounts, newEnabledTokenIds);

    localStorage.setItem(STORAGE_KEYS.ENABLED_TOKENS, JSON.stringify(newEnabledTokenIds));
    localStorage.setItem(STORAGE_KEYS.ASSETS, JSON.stringify(assets));

    set({ enabledTokenIds: newEnabledTokenIds, assets });
  },

  // Get asset by token ID
  getAssetByTokenId: (tokenId: string) => {
    return get().assets.find(a => a.token.id === tokenId);
  }
}));

// ============================================
// Helper Functions
// ============================================

function generateAccounts(mnemonic: string): WalletAccount[] {
  const accounts: WalletAccount[] = [];

  for (const chain of SUPPORTED_CHAINS) {
    const config = CHAIN_CONFIGS[chain];
    const address = getAddressFromMnemonic(mnemonic, chain);

    accounts.push({
      chain,
      address,
      balance: '0',
      balanceUsd: '0.00',
      derivationPath: `m/44'/${config.coinType}'/0'/0/0`
    });
  }

  return accounts;
}

// Generate asset holdings from enabled tokens
function generateAssets(accounts: WalletAccount[], enabledTokenIds: string[]): AssetHolding[] {
  const assets: AssetHolding[] = [];

  for (const tokenId of enabledTokenIds) {
    const token = TOKENS.find(t => t.id === tokenId);
    if (!token) continue;

    // Find the account for this token's chain
    const account = accounts.find(a => a.chain === token.chain);
    if (!account) continue;

    // Get demo balance for this token
    const balance = getDemoTokenBalance(token);
    const price = getDemoPrice(token);
    const balanceNum = parseFloat(balance);
    const balanceUsd = (balanceNum * price).toFixed(2);

    assets.push({
      token,
      balance,
      balanceFormatted: formatBalance(balance, token.decimals),
      balanceUsd,
      price,
      priceChange24h: getDemoPriceChange(token),
      address: account.address
    });
  }

  // Sort by USD value (highest first)
  return assets.sort((a, b) => parseFloat(b.balanceUsd) - parseFloat(a.balanceUsd));
}

// Generate assets with live prices (for refresh)
async function generateAssetsWithPrices(
  accounts: WalletAccount[],
  enabledTokenIds: string[]
): Promise<AssetHolding[]> {
  const assets: AssetHolding[] = [];

  for (const tokenId of enabledTokenIds) {
    const token = TOKENS.find(t => t.id === tokenId);
    if (!token) continue;

    const account = accounts.find(a => a.chain === token.chain);
    if (!account) continue;

    try {
      // For native tokens, use account balance
      let balance: string;
      if (token.type === TokenType.NATIVE) {
        balance = await fetchBalance(token.chain, account.address);
      } else {
        // For tokens, fetch ERC20 balance (demo for now)
        balance = getDemoTokenBalance(token);
      }

      const price = await fetchTokenPrice(token);
      const priceChange = getDemoPriceChange(token);
      const balanceNum = parseFloat(balance);
      const balanceUsd = (balanceNum * price).toFixed(2);

      assets.push({
        token,
        balance,
        balanceFormatted: formatBalance(balance, token.decimals),
        balanceUsd,
        price,
        priceChange24h: priceChange,
        address: account.address
      });
    } catch {
      // Use demo data on error
      const balance = getDemoTokenBalance(token);
      const price = getDemoPrice(token);
      assets.push({
        token,
        balance,
        balanceFormatted: formatBalance(balance, token.decimals),
        balanceUsd: (parseFloat(balance) * price).toFixed(2),
        price,
        priceChange24h: getDemoPriceChange(token),
        address: account.address
      });
    }
  }

  return assets.sort((a, b) => parseFloat(b.balanceUsd) - parseFloat(a.balanceUsd));
}

function formatBalance(balance: string, _decimals: number): string {
  const num = parseFloat(balance);
  if (num === 0) return '0';
  if (num < 0.0001) return '<0.0001';
  if (num < 1) return num.toFixed(4);
  if (num < 1000) return num.toFixed(2);
  if (num < 1000000) return `${(num / 1000).toFixed(2)}K`;
  return `${(num / 1000000).toFixed(2)}M`;
}

async function fetchBalance(chain: Chain, address: string): Promise<string> {
  const config = CHAIN_CONFIGS[chain];

  // ShareHODL and Cosmos chains
  if (config.restUrl) {
    try {
      const response = await fetch(
        `${config.restUrl}/cosmos/bank/v1beta1/balances/${address}`
      );
      const data = await response.json();
      const balance = data.balances?.find((b: { denom: string }) =>
        b.denom === 'uhodl' || b.denom === 'uatom' || b.denom === 'uosmo'
      );
      if (balance) {
        return (parseFloat(balance.amount) / Math.pow(10, config.decimals)).toString();
      }
    } catch {
      // Fallback
    }
  }

  // Return demo balance for now
  return getDemoBalance(chain);
}

async function fetchPrice(chain: Chain): Promise<number> {
  // Demo prices - in production, fetch from CoinGecko or similar
  return getDemoChainPrice(chain);
}

async function fetchTokenPrice(token: Token): Promise<number> {
  // In production, use CoinGecko API with token.coingeckoId
  return getDemoPrice(token);
}

function getDemoChainPrice(chain: Chain): number {
  const prices: Record<Chain, number> = {
    [Chain.SHAREHODL]: 1.0,
    [Chain.ETHEREUM]: 3450,
    [Chain.BITCOIN]: 67500,
    [Chain.COSMOS]: 8.50,
    [Chain.OSMOSIS]: 0.85,
    [Chain.POLYGON]: 0.58,
    [Chain.ARBITRUM]: 3450,
    [Chain.OPTIMISM]: 3450,
    [Chain.BASE]: 3450,
    [Chain.AVALANCHE]: 35,
    [Chain.BNB]: 580,
    [Chain.SOLANA]: 145,
    [Chain.CELESTIA]: 8.20
  };

  return prices[chain] || 0;
}

function getDemoBalance(chain: Chain): string {
  const balances: Record<Chain, string> = {
    [Chain.SHAREHODL]: '10000.00',
    [Chain.ETHEREUM]: '0.5',
    [Chain.BITCOIN]: '0.01',
    [Chain.COSMOS]: '25.5',
    [Chain.OSMOSIS]: '100.0',
    [Chain.POLYGON]: '150.0',
    [Chain.ARBITRUM]: '0.1',
    [Chain.OPTIMISM]: '0.1',
    [Chain.BASE]: '0.05',
    [Chain.AVALANCHE]: '5.0',
    [Chain.BNB]: '0.5',
    [Chain.SOLANA]: '2.0',
    [Chain.CELESTIA]: '50.0'
  };

  return balances[chain] || '0';
}

// Token-specific demo balances
function getDemoTokenBalance(token: Token): string {
  const balances: Record<string, string> = {
    'btc': '0.015',
    'eth': '0.5',
    'matic': '250.0',
    'bnb': '1.2',
    'sol': '3.5',
    'avax': '8.0',
    'atom': '25.5',
    'hodl': '10000.0',
    // Stablecoins
    'usdt-eth': '500.0',
    'usdc-eth': '750.0',
    'usdt-polygon': '200.0',
    'usdc-polygon': '300.0',
    'usdt-bsc': '150.0',
    'usdc-bsc': '250.0',
    'usdt-arb': '100.0',
    'usdc-arb': '175.0',
    'usdc-base': '125.0',
    'usdt-avax': '80.0',
    'usdc-avax': '120.0',
  };

  return balances[token.id] || '0';
}

// Demo prices for tokens
function getDemoPrice(token: Token): number {
  const prices: Record<string, number> = {
    'btc': 67500,
    'eth': 3450,
    'matic': 0.58,
    'bnb': 580,
    'sol': 145,
    'avax': 35,
    'atom': 8.50,
    'hodl': 1.0,
    // Stablecoins
    'usdt-eth': 1.0,
    'usdc-eth': 1.0,
    'usdt-polygon': 1.0,
    'usdc-polygon': 1.0,
    'usdt-bsc': 1.0,
    'usdc-bsc': 1.0,
    'usdt-arb': 1.0,
    'usdc-arb': 1.0,
    'usdc-base': 1.0,
    'usdt-avax': 1.0,
    'usdc-avax': 1.0,
  };

  return prices[token.id] || 0;
}

// Demo 24h price changes
function getDemoPriceChange(token: Token): number {
  const changes: Record<string, number> = {
    'btc': 2.5,
    'eth': 3.2,
    'matic': -1.8,
    'bnb': 1.5,
    'sol': 5.2,
    'avax': -0.8,
    'atom': 4.1,
    'hodl': 0.5,
    // Stablecoins are stable
    'usdt-eth': 0.01,
    'usdc-eth': -0.02,
    'usdt-polygon': 0.01,
    'usdc-polygon': -0.01,
    'usdt-bsc': 0.02,
    'usdc-bsc': -0.01,
    'usdt-arb': 0.01,
    'usdc-arb': 0.0,
    'usdc-base': -0.01,
    'usdt-avax': 0.01,
    'usdc-avax': 0.0,
  };

  return changes[token.id] || 0;
}
