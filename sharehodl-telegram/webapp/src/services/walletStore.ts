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
import { logger } from '../utils/logger';
import {
  generateMnemonic,
  validateMnemonic,
  getAddressFromMnemonic,
  encryptData,
  decryptData
} from '../utils/crypto';
import { fetchBalance as fetchSharehodlBalance } from './blockchainService';
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

// ShareHODL-themed wallet name generator
const HODL_ADJECTIVES = [
  'Diamond', 'Platinum', 'Titanium', 'Quantum', 'Genesis', 'Prime', 'Alpha', 'Apex',
  'Stellar', 'Sovereign', 'Cardinal', 'Sentinel', 'Founder', 'Pioneer', 'Vanguard', 'Elite'
];
const HODL_NOUNS = [
  'HODL', 'Vault', 'Reserve', 'Treasury', 'Stake', 'Holdings', 'Capital', 'Assets',
  'Portfolio', 'Ledger', 'Chain', 'Block', 'Node', 'Validator', 'Stash', 'Fund'
];

export function generateRandomWalletName(): string {
  const adjective = HODL_ADJECTIVES[Math.floor(Math.random() * HODL_ADJECTIVES.length)];
  const noun = HODL_NOUNS[Math.floor(Math.random() * HODL_NOUNS.length)];
  return `${adjective} ${noun}`;
}

// Generate cryptographically secure wallet ID
function generateSecureWalletId(): string {
  const randomBytes = new Uint8Array(12);
  crypto.getRandomValues(randomBytes);
  const randomPart = Array.from(randomBytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join('')
    .slice(0, 12);
  return `wallet_${Date.now()}_${randomPart}`;
}

// Storage keys
const STORAGE_KEYS = {
  ENCRYPTED_MNEMONIC: 'sh_encrypted_mnemonic',
  WALLET_INITIALIZED: 'sh_wallet_init',
  ACCOUNTS: 'sh_accounts',
  ASSETS: 'sh_assets',
  ENABLED_TOKENS: 'sh_enabled_tokens',
  // Multi-wallet support
  WALLETS: 'sh_wallets',
  ACTIVE_WALLET_ID: 'sh_active_wallet',
  // Biometric
  BIOMETRIC_TOKEN: 'sh_biometric_token'
};

// SECURITY: PIN cache timeout (5 minutes)
// After this period, cached PIN is cleared to reduce exposure window
const PIN_CACHE_TIMEOUT_MS = 5 * 60 * 1000; // 5 minutes

// Check if cached PIN has expired
function isPinCacheExpired(timestamp: number | null): boolean {
  if (!timestamp) return true;
  return Date.now() - timestamp > PIN_CACHE_TIMEOUT_MS;
}

// Wallet metadata for multi-wallet support
export interface WalletMetadata {
  id: string;
  name: string;
  createdAt: number;
  sharehodlAddress: string;
}

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

  // Cached PIN for transaction signing (cleared on lock or timeout)
  _cachedPin: string | null;
  _pinCacheTimestamp: number | null;

  // Actions
  initialize: () => Promise<void>;
  createWallet: (pin: string, name?: string) => Promise<string>;
  completeWalletSetup: () => void;  // Call after seed phrase verification
  importWallet: (mnemonic: string, pin: string, name?: string) => Promise<void>;
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

  // Transaction signing
  getMnemonicForSigning: (pin: string) => Promise<string>;
  getSharehodlAddress: () => string | undefined;

  // Multi-wallet support
  wallets: WalletMetadata[];
  activeWalletId: string | null;
  getWallets: () => WalletMetadata[];
  switchWallet: (walletId: string, pin: string) => Promise<void>;
  addWallet: (name: string, pin: string) => Promise<string>;
  importNewWallet: (name: string, mnemonic: string, pin: string) => Promise<void>;
  renameWallet: (walletId: string, newName: string) => void;
  deleteWallet: (walletId: string, pin: string) => Promise<void>;

  // Security settings
  verifyPin: (pin: string) => Promise<boolean>;
  changePin: (currentPin: string, newPin: string) => Promise<void>;
  getRecoveryPhrase: (pin: string) => Promise<string>;

  // Biometric
  setBiometricToken: (pin: string) => Promise<void>;
  unlockWithBiometric: (token: string) => Promise<void>;
  clearBiometricToken: () => void;
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

  // Cached PIN for transaction signing (cleared on lock or timeout)
  // SECURITY: PIN is auto-cleared after 5 minutes of inactivity
  _cachedPin: null,
  _pinCacheTimestamp: null,

  // Multi-wallet support
  wallets: [],
  activeWalletId: null,

  // Initialize - check if wallet exists
  initialize: async () => {
    try {
      const initialized = localStorage.getItem(STORAGE_KEYS.WALLET_INITIALIZED);
      set({ isInitialized: initialized === 'true' });
    } catch (error) {
      logger.error('Failed to initialize wallet:', error);
    }
  },

  // Create new wallet
  createWallet: async (pin: string, name?: string) => {
    set({ isLoading: true, error: null });

    try {
      // Generate mnemonic
      const mnemonic = generateMnemonic(256); // 24 words

      // Generate a wallet ID for this wallet
      const walletId = generateSecureWalletId();

      // Encrypt and store mnemonic (both main key and wallet-specific key)
      const encrypted = await encryptData(mnemonic, pin);
      localStorage.setItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC, encrypted);
      localStorage.setItem(`${STORAGE_KEYS.ENCRYPTED_MNEMONIC}_${walletId}`, encrypted);
      localStorage.setItem(STORAGE_KEYS.WALLET_INITIALIZED, 'true');

      // Generate accounts for all supported chains
      const accounts = generateAccounts(mnemonic);
      const enabledTokenIds = DEFAULT_ENABLED_TOKENS;
      const sharehodlAddress = accounts.find(a => a.chain === Chain.SHAREHODL)?.address || '';

      // Generate initial assets from enabled tokens
      const assets = generateAssets(accounts, enabledTokenIds);

      // Create wallet entry with provided name or generate random name
      const walletName = name?.trim() || generateRandomWalletName();
      const newWallet: WalletMetadata = {
        id: walletId,
        name: walletName,
        createdAt: Date.now(),
        sharehodlAddress
      };

      // Save wallet list and active wallet
      localStorage.setItem(STORAGE_KEYS.WALLETS, JSON.stringify([newWallet]));
      localStorage.setItem(STORAGE_KEYS.ACTIVE_WALLET_ID, walletId);

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
        enabledTokenIds,
        wallets: [newWallet],
        activeWalletId: walletId
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
  importWallet: async (mnemonic: string, pin: string, name?: string) => {
    set({ isLoading: true, error: null });

    try {
      // Validate mnemonic
      if (!validateMnemonic(mnemonic.trim())) {
        throw new Error('Invalid mnemonic phrase');
      }

      const cleanMnemonic = mnemonic.trim().toLowerCase();

      // Generate a wallet ID for this wallet
      const walletId = generateSecureWalletId();

      // Encrypt and store mnemonic (both main key and wallet-specific key)
      const encrypted = await encryptData(cleanMnemonic, pin);
      localStorage.setItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC, encrypted);
      localStorage.setItem(`${STORAGE_KEYS.ENCRYPTED_MNEMONIC}_${walletId}`, encrypted);
      localStorage.setItem(STORAGE_KEYS.WALLET_INITIALIZED, 'true');

      // Generate accounts
      const accounts = generateAccounts(cleanMnemonic);
      const enabledTokenIds = DEFAULT_ENABLED_TOKENS;
      const assets = generateAssets(accounts, enabledTokenIds);
      const sharehodlAddress = accounts.find(a => a.chain === Chain.SHAREHODL)?.address || '';

      // Create wallet entry with provided name or generate random name
      const walletName = name?.trim() || generateRandomWalletName();
      const newWallet: WalletMetadata = {
        id: walletId,
        name: walletName,
        createdAt: Date.now(),
        sharehodlAddress
      };

      // Save wallet list and active wallet
      localStorage.setItem(STORAGE_KEYS.WALLETS, JSON.stringify([newWallet]));
      localStorage.setItem(STORAGE_KEYS.ACTIVE_WALLET_ID, walletId);

      localStorage.setItem(STORAGE_KEYS.ACCOUNTS, JSON.stringify(accounts));
      localStorage.setItem(STORAGE_KEYS.ASSETS, JSON.stringify(assets));
      localStorage.setItem(STORAGE_KEYS.ENABLED_TOKENS, JSON.stringify(enabledTokenIds));

      set({
        isInitialized: true,
        isLocked: false,
        isLoading: false,
        accounts,
        assets,
        enabledTokenIds,
        wallets: [newWallet],
        activeWalletId: walletId
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

      // Load wallets list and active wallet ID
      const walletsJson = localStorage.getItem(STORAGE_KEYS.WALLETS);
      let wallets: WalletMetadata[] = walletsJson ? JSON.parse(walletsJson) : [];
      let activeWalletId = localStorage.getItem(STORAGE_KEYS.ACTIVE_WALLET_ID);

      // MIGRATION: If wallets list is empty but we have encrypted mnemonic,
      // this is an old wallet that needs to be added to the multi-wallet list
      if (wallets.length === 0 && encrypted) {
        const sharehodlAddress = accounts.find((a: WalletAccount) => a.chain === Chain.SHAREHODL)?.address || '';
        const mainWalletId = `wallet_main_${Date.now()}`;
        const mainWallet: WalletMetadata = {
          id: mainWalletId,
          name: generateRandomWalletName(),
          createdAt: Date.now(),
          sharehodlAddress
        };
        wallets = [mainWallet];
        activeWalletId = mainWalletId;

        // Store the main wallet's mnemonic with its ID
        const encryptedKey = `${STORAGE_KEYS.ENCRYPTED_MNEMONIC}_${mainWalletId}`;
        localStorage.setItem(encryptedKey, encrypted);

        // Save the migrated wallet list
        localStorage.setItem(STORAGE_KEYS.WALLETS, JSON.stringify(wallets));
        localStorage.setItem(STORAGE_KEYS.ACTIVE_WALLET_ID, mainWalletId);
      }

      set({
        isLocked: false,
        isLoading: false,
        accounts,
        assets,
        enabledTokenIds,
        wallets,
        activeWalletId,
        securityState: getSecurityState(),
        remainingAttempts: getRemainingAttempts(),
        // SECURITY: Cache PIN for transaction signing with timestamp
        _cachedPin: pin,
        _pinCacheTimestamp: Date.now()
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
    // SECURITY: Clear cached PIN and timestamp on lock
    set({ isLocked: true, _cachedPin: null, _pinCacheTimestamp: null });
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
      logger.error('Failed to refresh balances:', error);
    }
  },

  // Clear error
  clearError: () => set({ error: null }),

  // Reset wallet (dangerous!) - Complete cleanup of ALL wallet data
  resetWallet: () => {
    // Get list of wallets to clear individual mnemonic keys
    try {
      const walletsJson = localStorage.getItem(STORAGE_KEYS.WALLETS);
      if (walletsJson) {
        const wallets: WalletMetadata[] = JSON.parse(walletsJson);
        // Remove each wallet's individual encrypted mnemonic
        for (const wallet of wallets) {
          const walletKey = `${STORAGE_KEYS.ENCRYPTED_MNEMONIC}_${wallet.id}`;
          localStorage.removeItem(walletKey);
        }
      }
    } catch (e) {
      // Ignore parsing errors, continue with cleanup
    }

    // Clear all wallet-related storage keys
    localStorage.removeItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC);
    localStorage.removeItem(STORAGE_KEYS.WALLET_INITIALIZED);
    localStorage.removeItem(STORAGE_KEYS.ACCOUNTS);
    localStorage.removeItem(STORAGE_KEYS.ASSETS);
    localStorage.removeItem(STORAGE_KEYS.ENABLED_TOKENS);
    localStorage.removeItem(STORAGE_KEYS.WALLETS);
    localStorage.removeItem(STORAGE_KEYS.ACTIVE_WALLET_ID);
    localStorage.removeItem(STORAGE_KEYS.BIOMETRIC_TOKEN);

    // Clear security data (failed attempts, lockout, etc.)
    clearSecurityData();

    // Reset state to initial values
    set({
      isInitialized: false,
      isLocked: true,
      accounts: [],
      assets: [],
      wallets: [],
      activeWalletId: null,
      enabledTokenIds: DEFAULT_ENABLED_TOKENS,
      totalBalanceUsd: 0,
      error: null,
      // SECURITY: Clear PIN cache on reset
      _cachedPin: null,
      _pinCacheTimestamp: null,
      securityState: getSecurityState(),
      remainingAttempts: getRemainingAttempts()
    });
  },

  // Update last activity timestamp (call on user interaction)
  updateActivity: () => {
    updateLastActivity();
  },

  // Check if auto-lock should trigger
  // SECURITY: Also clears stale PIN cache
  checkAutoLock: () => {
    const state = get();

    // SECURITY: Clear stale PIN cache even if not auto-locking
    if (state._cachedPin && isPinCacheExpired(state._pinCacheTimestamp)) {
      set({ _cachedPin: null, _pinCacheTimestamp: null });
    }

    if (shouldAutoLock() && !state.isLocked) {
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
  },

  // Get mnemonic for transaction signing
  // SECURITY: This should only be called when needed for signing
  getMnemonicForSigning: async (pin: string): Promise<string> => {
    const encrypted = localStorage.getItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC);
    if (!encrypted) {
      throw new Error('No wallet found');
    }

    try {
      const mnemonic = await decryptData(encrypted, pin);
      return mnemonic;
    } catch (error) {
      throw new Error('Invalid PIN');
    }
  },

  // Get the ShareHODL address from accounts
  getSharehodlAddress: (): string | undefined => {
    const { accounts } = get();
    const sharehodlAccount = accounts.find(a => a.chain === Chain.SHAREHODL);
    return sharehodlAccount?.address;
  },

  // ============================================
  // Multi-Wallet Support
  // ============================================

  getWallets: (): WalletMetadata[] => {
    try {
      const walletsJson = localStorage.getItem(STORAGE_KEYS.WALLETS);
      return walletsJson ? JSON.parse(walletsJson) : [];
    } catch {
      return [];
    }
  },

  switchWallet: async (walletId: string, pin: string): Promise<void> => {
    const wallets = get().getWallets();
    const wallet = wallets.find(w => w.id === walletId);
    if (!wallet) throw new Error('Wallet not found');

    // Get encrypted mnemonic for this wallet
    const encryptedKey = `${STORAGE_KEYS.ENCRYPTED_MNEMONIC}_${walletId}`;
    const encrypted = localStorage.getItem(encryptedKey);
    if (!encrypted) throw new Error('Wallet data not found');

    try {
      const mnemonic = await decryptData(encrypted, pin);
      const accounts = generateAccounts(mnemonic);
      const enabledTokenIds = DEFAULT_ENABLED_TOKENS;
      const assets = generateAssets(accounts, enabledTokenIds);

      localStorage.setItem(STORAGE_KEYS.ACTIVE_WALLET_ID, walletId);
      localStorage.setItem(STORAGE_KEYS.ACCOUNTS, JSON.stringify(accounts));
      localStorage.setItem(STORAGE_KEYS.ASSETS, JSON.stringify(assets));

      set({
        activeWalletId: walletId,
        accounts,
        assets,
        enabledTokenIds,
        // SECURITY: Update PIN cache with timestamp
        _cachedPin: pin,
        _pinCacheTimestamp: Date.now()
      });

      get().refreshBalances();
    } catch {
      throw new Error('Invalid PIN');
    }
  },

  addWallet: async (name: string, pin: string): Promise<string> => {
    set({ isLoading: true, error: null });

    try {
      // Generate new mnemonic
      const mnemonic = generateMnemonic(256);
      const walletId = generateSecureWalletId();

      // Encrypt and store
      const encrypted = await encryptData(mnemonic, pin);
      const encryptedKey = `${STORAGE_KEYS.ENCRYPTED_MNEMONIC}_${walletId}`;
      localStorage.setItem(encryptedKey, encrypted);

      // Generate accounts to get address
      const accounts = generateAccounts(mnemonic);
      const sharehodlAddress = accounts.find(a => a.chain === Chain.SHAREHODL)?.address || '';

      // Add to wallets list
      const wallets = get().getWallets();
      const newWallet: WalletMetadata = {
        id: walletId,
        name: name || `Wallet ${wallets.length + 1}`,
        createdAt: Date.now(),
        sharehodlAddress
      };
      wallets.push(newWallet);
      localStorage.setItem(STORAGE_KEYS.WALLETS, JSON.stringify(wallets));

      // If this is the first wallet, also set as primary mnemonic
      if (wallets.length === 1) {
        localStorage.setItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC, encrypted);
        localStorage.setItem(STORAGE_KEYS.WALLET_INITIALIZED, 'true');
      }

      // Always switch to the new wallet
      localStorage.setItem(STORAGE_KEYS.ACTIVE_WALLET_ID, walletId);

      // Generate assets for the new wallet
      const enabledTokenIds = get().enabledTokenIds;
      const assets = generateAssets(accounts, enabledTokenIds);

      // Save account and asset data for the new wallet
      localStorage.setItem(STORAGE_KEYS.ACCOUNTS, JSON.stringify(accounts));
      localStorage.setItem(STORAGE_KEYS.ASSETS, JSON.stringify(assets));

      set({
        isLoading: false,
        wallets,
        activeWalletId: walletId,
        accounts,
        assets
      });

      // Refresh balances for new wallet
      get().refreshBalances();

      return mnemonic;
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to create wallet';
      set({ isLoading: false, error: message });
      throw error;
    }
  },

  importNewWallet: async (name: string, mnemonic: string, pin: string): Promise<void> => {
    set({ isLoading: true, error: null });

    try {
      // Clean and normalize the mnemonic (lowercase, collapse whitespace for BIP39 validation)
      const cleanMnemonic = mnemonic.trim().toLowerCase().replace(/\s+/g, ' ');

      if (!validateMnemonic(cleanMnemonic)) {
        throw new Error('Invalid mnemonic phrase');
      }
      const walletId = generateSecureWalletId();

      // Encrypt and store
      const encrypted = await encryptData(cleanMnemonic, pin);
      const encryptedKey = `${STORAGE_KEYS.ENCRYPTED_MNEMONIC}_${walletId}`;
      localStorage.setItem(encryptedKey, encrypted);

      // Generate accounts
      const accounts = generateAccounts(cleanMnemonic);
      const sharehodlAddress = accounts.find(a => a.chain === Chain.SHAREHODL)?.address || '';

      // Add to wallets list
      const wallets = get().getWallets();
      const newWallet: WalletMetadata = {
        id: walletId,
        name: name || `Imported Wallet ${wallets.length + 1}`,
        createdAt: Date.now(),
        sharehodlAddress
      };
      wallets.push(newWallet);
      localStorage.setItem(STORAGE_KEYS.WALLETS, JSON.stringify(wallets));

      // If this is the first wallet, set as primary mnemonic
      if (wallets.length === 1) {
        localStorage.setItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC, encrypted);
        localStorage.setItem(STORAGE_KEYS.WALLET_INITIALIZED, 'true');
      }

      // Always switch to the new wallet
      localStorage.setItem(STORAGE_KEYS.ACTIVE_WALLET_ID, walletId);

      // Generate assets for the new wallet
      const enabledTokenIds = get().enabledTokenIds;
      const assets = generateAssets(accounts, enabledTokenIds);

      // Save account and asset data for the new wallet
      localStorage.setItem(STORAGE_KEYS.ACCOUNTS, JSON.stringify(accounts));
      localStorage.setItem(STORAGE_KEYS.ASSETS, JSON.stringify(assets));

      set({
        isLoading: false,
        wallets,
        activeWalletId: walletId,
        accounts,
        assets
      });

      // Refresh balances for new wallet
      get().refreshBalances();
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to import wallet';
      set({ isLoading: false, error: message });
      throw error;
    }
  },

  renameWallet: (walletId: string, newName: string): void => {
    const wallets = get().getWallets();
    const walletIndex = wallets.findIndex(w => w.id === walletId);
    if (walletIndex === -1) return;

    wallets[walletIndex].name = newName;
    localStorage.setItem(STORAGE_KEYS.WALLETS, JSON.stringify(wallets));
    set({ wallets });
  },

  deleteWallet: async (walletId: string, pin: string): Promise<void> => {
    // Verify PIN first
    const isValid = await get().verifyPin(pin);
    if (!isValid) throw new Error('Invalid PIN');

    const wallets = get().getWallets();
    if (wallets.length <= 1) {
      throw new Error('Cannot delete the only wallet');
    }

    // Remove wallet encrypted data
    const encryptedKey = `${STORAGE_KEYS.ENCRYPTED_MNEMONIC}_${walletId}`;
    localStorage.removeItem(encryptedKey);

    // Remove from wallets list
    const newWallets = wallets.filter(w => w.id !== walletId);
    localStorage.setItem(STORAGE_KEYS.WALLETS, JSON.stringify(newWallets));

    // If this was the active wallet, switch to first available
    if (get().activeWalletId === walletId && newWallets.length > 0) {
      await get().switchWallet(newWallets[0].id, pin);
    }

    set({ wallets: newWallets });
  },

  // ============================================
  // Security Settings
  // ============================================

  verifyPin: async (pin: string): Promise<boolean> => {
    const encrypted = localStorage.getItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC);
    if (!encrypted) return false;

    try {
      await decryptData(encrypted, pin);
      return true;
    } catch {
      return false;
    }
  },

  changePin: async (currentPin: string, newPin: string): Promise<void> => {
    // Verify current PIN
    const isValid = await get().verifyPin(currentPin);
    if (!isValid) throw new Error('Current PIN is incorrect');

    // Get and re-encrypt main mnemonic
    const encrypted = localStorage.getItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC);
    if (!encrypted) throw new Error('No wallet found');

    const mnemonic = await decryptData(encrypted, currentPin);
    const newEncrypted = await encryptData(mnemonic, newPin);
    localStorage.setItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC, newEncrypted);

    // Re-encrypt all wallet mnemonics
    const wallets = get().getWallets();
    for (const wallet of wallets) {
      const walletKey = `${STORAGE_KEYS.ENCRYPTED_MNEMONIC}_${wallet.id}`;
      const walletEncrypted = localStorage.getItem(walletKey);
      if (walletEncrypted) {
        try {
          const walletMnemonic = await decryptData(walletEncrypted, currentPin);
          const newWalletEncrypted = await encryptData(walletMnemonic, newPin);
          localStorage.setItem(walletKey, newWalletEncrypted);
        } catch {
          // Skip if can't decrypt (might be corrupted)
        }
      }
    }

    // Clear biometric token since PIN changed
    localStorage.removeItem(STORAGE_KEYS.BIOMETRIC_TOKEN);

    // SECURITY: Update cached PIN with timestamp
    set({ _cachedPin: newPin, _pinCacheTimestamp: Date.now() });
  },

  getRecoveryPhrase: async (pin: string): Promise<string> => {
    const encrypted = localStorage.getItem(STORAGE_KEYS.ENCRYPTED_MNEMONIC);
    if (!encrypted) throw new Error('No wallet found');

    try {
      return await decryptData(encrypted, pin);
    } catch {
      throw new Error('Invalid PIN');
    }
  },

  // ============================================
  // Biometric Authentication
  // ============================================

  setBiometricToken: async (pin: string): Promise<void> => {
    // Verify PIN first
    const isValid = await get().verifyPin(pin);
    if (!isValid) throw new Error('Invalid PIN');

    // Generate a secure random token
    const tokenBytes = new Uint8Array(32);
    crypto.getRandomValues(tokenBytes);
    const token = btoa(String.fromCharCode(...tokenBytes));

    // SECURITY: Create payload with expiration (30 days)
    const expiresAt = Date.now() + (30 * 24 * 60 * 60 * 1000);
    const payload = JSON.stringify({ pin, expiresAt });

    // Encrypt the payload with the token for later retrieval
    const encryptedPayload = await encryptData(payload, token);
    localStorage.setItem(STORAGE_KEYS.BIOMETRIC_TOKEN, encryptedPayload);
  },

  unlockWithBiometric: async (token: string): Promise<void> => {
    const encryptedPayload = localStorage.getItem(STORAGE_KEYS.BIOMETRIC_TOKEN);
    if (!encryptedPayload) throw new Error('Biometric not set up');

    try {
      // Decrypt the payload using the biometric token
      const payloadStr = await decryptData(encryptedPayload, token);
      const payload = JSON.parse(payloadStr);

      // SECURITY: Check if biometric token has expired
      if (payload.expiresAt && Date.now() > payload.expiresAt) {
        // Clear expired token
        localStorage.removeItem(STORAGE_KEYS.BIOMETRIC_TOKEN);
        throw new Error('Biometric authentication has expired. Please set up again.');
      }

      // Use the PIN to unlock normally
      await get().unlockWallet(payload.pin);
    } catch (error) {
      // Clear invalid token
      if (error instanceof SyntaxError) {
        localStorage.removeItem(STORAGE_KEYS.BIOMETRIC_TOKEN);
      }
      throw error instanceof Error ? error : new Error('Biometric authentication failed');
    }
  },

  clearBiometricToken: (): void => {
    localStorage.removeItem(STORAGE_KEYS.BIOMETRIC_TOKEN);
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

    // Get balance for this token (0 until fetched from blockchain)
    const balance = getTokenBalance(token);
    const price = getTokenPrice(token);
    const balanceNum = parseFloat(balance);
    const balanceUsd = (balanceNum * price).toFixed(2);

    assets.push({
      token,
      balance,
      balanceFormatted: formatBalance(balance, token.decimals),
      balanceUsd,
      price,
      priceChange24h: getPriceChange(token),
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
        // For tokens, ERC20 balance fetching not yet implemented
        balance = getTokenBalance(token);
      }

      const price = await fetchTokenPrice(token);
      const priceChange = getPriceChange(token);
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
      // Return 0 balance on error
      const balance = getTokenBalance(token);
      const price = getTokenPrice(token);
      assets.push({
        token,
        balance,
        balanceFormatted: formatBalance(balance, token.decimals),
        balanceUsd: (parseFloat(balance) * price).toFixed(2),
        price,
        priceChange24h: getPriceChange(token),
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

  // Use blockchain service for ShareHODL chain (real RPC connection)
  if (chain === Chain.SHAREHODL) {
    try {
      const result = await fetchSharehodlBalance(address);
      const balanceNum = parseFloat(result.balance) / Math.pow(10, config.decimals);
      return balanceNum.toString();
    } catch (error) {
      logger.error('Failed to fetch ShareHODL balance:', error);
      // Return 0 on error (blockchain not reachable)
      return getRealBalance(chain);
    }
  }

  // Other Cosmos chains - use REST API
  if (config.restUrl) {
    try {
      const response = await fetch(
        `${config.restUrl}/cosmos/bank/v1beta1/balances/${address}`
      );
      const data = await response.json();
      const balance = data.balances?.find((b: { denom: string }) =>
        b.denom === 'uatom' || b.denom === 'uosmo'
      );
      if (balance) {
        return (parseFloat(balance.amount) / Math.pow(10, config.decimals)).toString();
      }
    } catch {
      // Fallback
    }
  }

  // Return 0 for chains without real integration
  return getRealBalance(chain);
}

async function fetchPrice(chain: Chain): Promise<number> {
  // Reference prices - in production, fetch from CoinGecko or similar
  return getChainPrice(chain);
}

async function fetchTokenPrice(token: Token): Promise<number> {
  // In production, use CoinGecko API with token.coingeckoId
  return getTokenPrice(token);
}

// Reference prices for USD conversion (these are static for now)
// In production, these would be fetched from a price oracle like CoinGecko
function getChainPrice(chain: Chain): number {
  const prices: Record<Chain, number> = {
    [Chain.SHAREHODL]: 1.0,  // HODL is pegged to $1
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

// Real balance - returns 0 until fetched from blockchain
function getRealBalance(_chain: Chain): string {
  // All balances start at 0 until fetched from blockchain
  return '0';
}

// Token-specific balance - returns 0 until fetched from blockchain
function getTokenBalance(_token: Token): string {
  // All balances start at 0 until fetched from blockchain
  return '0';
}

// Token prices (for USD conversion)
// In production, these would be fetched from a price API
function getTokenPrice(token: Token): number {
  const prices: Record<string, number> = {
    'btc': 67500,
    'eth': 3450,
    'matic': 0.58,
    'bnb': 580,
    'sol': 145,
    'avax': 35,
    'atom': 8.50,
    'hodl': 1.0,  // HODL is pegged to $1
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

// Price change - returns 0 until we have real price history
function getPriceChange(_token: Token): number {
  // No price change data until connected to price API
  return 0;
}
