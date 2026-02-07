/**
 * ShareHODL Blockchain Service
 *
 * Provides transaction signing and broadcasting using CosmJS.
 * SECURITY: Private keys are used only at signing time and immediately discarded.
 */

import { SigningStargateClient, StargateClient, GasPrice } from '@cosmjs/stargate';
import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import { Chain, CHAIN_CONFIGS } from '../types';
import { logger } from '../utils/logger';

// Chain configuration
const SHAREHODL_CONFIG = CHAIN_CONFIGS[Chain.SHAREHODL];
const RPC_URL = SHAREHODL_CONFIG.rpcUrl || 'https://rpc.sharehodl.com';
const REST_URL = SHAREHODL_CONFIG.restUrl || 'https://api.sharehodl.com';
const DENOM = 'uhodl';
const GAS_PRICE = GasPrice.fromString('0.025uhodl');
const ADDRESS_PREFIX = 'hodl';

// =============================================================================
// SECURITY: Rate Limiting
// =============================================================================

interface RateLimitEntry {
  count: number;
  resetTime: number;
}

class RateLimiter {
  private limits: Map<string, RateLimitEntry> = new Map();
  private readonly maxRequests: number;
  private readonly windowMs: number;

  constructor(maxRequests: number = 10, windowMs: number = 60000) {
    this.maxRequests = maxRequests;
    this.windowMs = windowMs;
  }

  /**
   * Check if a request should be allowed
   * @param key Identifier for rate limiting (e.g., function name or address)
   * @returns true if allowed, false if rate limited
   */
  check(key: string): boolean {
    const now = Date.now();
    const entry = this.limits.get(key);

    if (!entry || now >= entry.resetTime) {
      // New window
      this.limits.set(key, { count: 1, resetTime: now + this.windowMs });
      return true;
    }

    if (entry.count >= this.maxRequests) {
      return false;
    }

    entry.count++;
    return true;
  }

  /**
   * Get remaining time until rate limit resets
   */
  getRemainingTime(key: string): number {
    const entry = this.limits.get(key);
    if (!entry) return 0;
    return Math.max(0, entry.resetTime - Date.now());
  }
}

// SECURITY: Global rate limiters for different operation types
const queryRateLimiter = new RateLimiter(30, 60000);  // 30 queries per minute
const txRateLimiter = new RateLimiter(5, 60000);      // 5 transactions per minute

// Response types
export interface TransactionResult {
  success: boolean;
  txHash?: string;
  error?: string;
  gasUsed?: number;
  height?: number;
}

export interface BalanceResult {
  balance: string;
  denom: string;
}

export interface AccountInfo {
  address: string;
  balance: string;
  accountNumber?: number;
  sequence?: number;
}

/**
 * Create a wallet from mnemonic (for signing)
 * SECURITY: The wallet should be used immediately and not stored
 */
async function createWalletFromMnemonic(mnemonic: string): Promise<DirectSecp256k1HdWallet> {
  return DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
    prefix: ADDRESS_PREFIX,
  });
}

/**
 * Get a read-only client for querying the blockchain
 */
export async function getQueryClient(): Promise<StargateClient> {
  return StargateClient.connect(RPC_URL);
}

/**
 * Create a signing client from mnemonic
 * SECURITY: Use this only for the duration of signing, then disconnect
 */
async function getSigningClient(
  mnemonic: string
): Promise<{ client: SigningStargateClient; address: string }> {
  const wallet = await createWalletFromMnemonic(mnemonic);
  const [account] = await wallet.getAccounts();

  const client = await SigningStargateClient.connectWithSigner(RPC_URL, wallet, {
    gasPrice: GAS_PRICE,
  });

  return { client, address: account.address };
}

/**
 * Fetch account balance from blockchain
 * SECURITY: Rate limited to prevent API abuse
 */
export async function fetchBalance(address: string): Promise<BalanceResult> {
  // SECURITY: Rate limiting check
  if (!queryRateLimiter.check(`balance:${address}`)) {
    logger.warn('Rate limit exceeded for balance query', { address });
    return { balance: '0', denom: DENOM };
  }

  try {
    // Try RPC first
    const client = await getQueryClient();
    const balance = await client.getBalance(address, DENOM);
    await client.disconnect();

    return {
      balance: balance.amount,
      denom: balance.denom,
    };
  } catch (rpcError) {
    // Fallback to REST API
    try {
      const response = await fetch(
        `${REST_URL}/cosmos/bank/v1beta1/balances/${address}`
      );

      if (!response.ok) {
        throw new Error(`REST API error: ${response.status}`);
      }

      const data = await response.json();
      const balance = data.balances?.find(
        (b: { denom: string }) => b.denom === DENOM
      );

      return {
        balance: balance?.amount || '0',
        denom: DENOM,
      };
    } catch (restError) {
      logger.error('Failed to fetch balance:', restError);
      return { balance: '0', denom: DENOM };
    }
  }
}

/**
 * Get full account info including sequence number
 */
export async function getAccountInfo(address: string): Promise<AccountInfo> {
  try {
    const client = await getQueryClient();
    const account = await client.getAccount(address);
    const balance = await client.getBalance(address, DENOM);
    await client.disconnect();

    return {
      address,
      balance: balance.amount,
      accountNumber: account?.accountNumber,
      sequence: account?.sequence,
    };
  } catch (error) {
    logger.error('Failed to get account info:', error);
    return {
      address,
      balance: '0',
    };
  }
}

/**
 * Send tokens to another address
 * SECURITY: Mnemonic is used only for signing and then discarded
 * SECURITY: Rate limited to prevent transaction spam
 */
export async function sendTokens(
  mnemonic: string,
  recipientAddress: string,
  amount: string,
  memo?: string
): Promise<TransactionResult> {
  // SECURITY: Rate limiting check for transactions
  if (!txRateLimiter.check('sendTokens')) {
    const remainingMs = txRateLimiter.getRemainingTime('sendTokens');
    const remainingSecs = Math.ceil(remainingMs / 1000);
    return {
      success: false,
      error: `Rate limit exceeded. Please wait ${remainingSecs} seconds before trying again.`
    };
  }

  let signingClient: SigningStargateClient | null = null;

  try {
    // Validate inputs
    if (!mnemonic || !recipientAddress || !amount) {
      return { success: false, error: 'Missing required parameters' };
    }

    // Validate recipient address format
    if (!recipientAddress.startsWith(ADDRESS_PREFIX)) {
      return { success: false, error: `Invalid address format. Expected ${ADDRESS_PREFIX} prefix.` };
    }

    // Convert amount to micro units (uhodl)
    const amountInMicro = Math.floor(parseFloat(amount) * 1_000_000);
    if (amountInMicro <= 0) {
      return { success: false, error: 'Amount must be greater than 0' };
    }

    // Create signing client
    const { client, address: senderAddress } = await getSigningClient(mnemonic);
    signingClient = client;

    // Check balance
    const balance = await client.getBalance(senderAddress, DENOM);
    if (BigInt(balance.amount) < BigInt(amountInMicro)) {
      await client.disconnect();
      return { success: false, error: 'Insufficient balance' };
    }

    // Create and sign transaction
    const sendMsg = {
      typeUrl: '/cosmos.bank.v1beta1.MsgSend',
      value: {
        fromAddress: senderAddress,
        toAddress: recipientAddress,
        amount: [{ denom: DENOM, amount: amountInMicro.toString() }],
      },
    };

    // Estimate gas and send
    const result = await client.signAndBroadcast(
      senderAddress,
      [sendMsg],
      'auto',
      memo || ''
    );

    // Disconnect
    await client.disconnect();
    signingClient = null;

    // Check result
    if (result.code === 0) {
      return {
        success: true,
        txHash: result.transactionHash,
        gasUsed: Number(result.gasUsed),
        height: Number(result.height),
      };
    } else {
      return {
        success: false,
        error: `Transaction failed: ${result.rawLog}`,
        txHash: result.transactionHash,
      };
    }
  } catch (error) {
    // Ensure client is disconnected on error
    if (signingClient) {
      try {
        await signingClient.disconnect();
      } catch {
        // Ignore disconnect errors
      }
    }

    const errorMessage = error instanceof Error ? error.message : 'Unknown error';
    logger.error('Send transaction failed:', errorMessage);

    // Provide user-friendly error messages
    if (errorMessage.includes('insufficient funds')) {
      return { success: false, error: 'Insufficient balance for transaction and fees' };
    }
    if (errorMessage.includes('account sequence mismatch')) {
      return { success: false, error: 'Transaction conflict. Please wait a moment and try again.' };
    }
    if (errorMessage.includes('decoding bech32 failed')) {
      return { success: false, error: 'Invalid recipient address format' };
    }
    if (errorMessage.includes('does not exist on chain') || errorMessage.includes('account not found')) {
      return {
        success: false,
        error: "Account 'does not exist on chain'. Send some tokens there before trying to query sequence."
      };
    }

    return { success: false, error: errorMessage };
  }
}

/**
 * Delegate tokens to a validator
 * SECURITY: Rate limited to prevent transaction spam
 */
export async function delegateTokens(
  mnemonic: string,
  validatorAddress: string,
  amount: string
): Promise<TransactionResult> {
  // SECURITY: Rate limiting check
  if (!txRateLimiter.check('delegateTokens')) {
    const remainingMs = txRateLimiter.getRemainingTime('delegateTokens');
    return {
      success: false,
      error: `Rate limit exceeded. Please wait ${Math.ceil(remainingMs / 1000)} seconds.`
    };
  }

  let signingClient: SigningStargateClient | null = null;

  try {
    const amountInMicro = Math.floor(parseFloat(amount) * 1_000_000);
    if (amountInMicro <= 0) {
      return { success: false, error: 'Amount must be greater than 0' };
    }

    const { client, address: delegatorAddress } = await getSigningClient(mnemonic);
    signingClient = client;

    const delegateMsg = {
      typeUrl: '/cosmos.staking.v1beta1.MsgDelegate',
      value: {
        delegatorAddress,
        validatorAddress,
        amount: { denom: DENOM, amount: amountInMicro.toString() },
      },
    };

    const result = await client.signAndBroadcast(
      delegatorAddress,
      [delegateMsg],
      'auto',
      ''
    );

    await client.disconnect();
    signingClient = null;

    if (result.code === 0) {
      return {
        success: true,
        txHash: result.transactionHash,
        gasUsed: Number(result.gasUsed),
        height: Number(result.height),
      };
    } else {
      return {
        success: false,
        error: `Delegation failed: ${result.rawLog}`,
      };
    }
  } catch (error) {
    if (signingClient) {
      try {
        await signingClient.disconnect();
      } catch {
        // Ignore disconnect errors
      }
    }

    const errorMessage = error instanceof Error ? error.message : 'Unknown error';

    // User-friendly error messages for delegation
    if (errorMessage.includes('does not exist on chain') || errorMessage.includes('account not found')) {
      return {
        success: false,
        error: "Account 'does not exist on chain'. Send some tokens there before trying to query sequence."
      };
    }
    if (errorMessage.includes('account sequence mismatch')) {
      return { success: false, error: 'Transaction conflict. Please wait a moment and try again.' };
    }
    if (errorMessage.includes('insufficient funds')) {
      return { success: false, error: 'Insufficient balance for staking and fees' };
    }

    return { success: false, error: errorMessage };
  }
}

/**
 * Undelegate tokens from a validator
 * SECURITY: Rate limited to prevent transaction spam
 */
export async function undelegateTokens(
  mnemonic: string,
  validatorAddress: string,
  amount: string
): Promise<TransactionResult> {
  // SECURITY: Rate limiting check
  if (!txRateLimiter.check('undelegateTokens')) {
    const remainingMs = txRateLimiter.getRemainingTime('undelegateTokens');
    return {
      success: false,
      error: `Rate limit exceeded. Please wait ${Math.ceil(remainingMs / 1000)} seconds.`
    };
  }

  let signingClient: SigningStargateClient | null = null;

  try {
    const amountInMicro = Math.floor(parseFloat(amount) * 1_000_000);
    if (amountInMicro <= 0) {
      return { success: false, error: 'Amount must be greater than 0' };
    }

    const { client, address: delegatorAddress } = await getSigningClient(mnemonic);
    signingClient = client;

    const undelegateMsg = {
      typeUrl: '/cosmos.staking.v1beta1.MsgUndelegate',
      value: {
        delegatorAddress,
        validatorAddress,
        amount: { denom: DENOM, amount: amountInMicro.toString() },
      },
    };

    const result = await client.signAndBroadcast(
      delegatorAddress,
      [undelegateMsg],
      'auto',
      ''
    );

    await client.disconnect();
    signingClient = null;

    if (result.code === 0) {
      return {
        success: true,
        txHash: result.transactionHash,
        gasUsed: Number(result.gasUsed),
        height: Number(result.height),
      };
    } else {
      return {
        success: false,
        error: `Undelegation failed: ${result.rawLog}`,
      };
    }
  } catch (error) {
    if (signingClient) {
      try {
        await signingClient.disconnect();
      } catch {
        // Ignore disconnect errors
      }
    }

    const errorMessage = error instanceof Error ? error.message : 'Unknown error';

    // User-friendly error messages for undelegation
    if (errorMessage.includes('does not exist on chain') || errorMessage.includes('account not found')) {
      return {
        success: false,
        error: "Account 'does not exist on chain'. Send some tokens there before trying to query sequence."
      };
    }
    if (errorMessage.includes('account sequence mismatch')) {
      return { success: false, error: 'Transaction conflict. Please wait a moment and try again.' };
    }

    return { success: false, error: errorMessage };
  }
}

/**
 * Claim staking rewards from a validator
 * SECURITY: Rate limited to prevent transaction spam
 */
export async function claimRewards(
  mnemonic: string,
  validatorAddress: string
): Promise<TransactionResult> {
  // SECURITY: Rate limiting check
  if (!txRateLimiter.check('claimRewards')) {
    const remainingMs = txRateLimiter.getRemainingTime('claimRewards');
    return {
      success: false,
      error: `Rate limit exceeded. Please wait ${Math.ceil(remainingMs / 1000)} seconds.`
    };
  }

  let signingClient: SigningStargateClient | null = null;

  try {
    const { client, address: delegatorAddress } = await getSigningClient(mnemonic);
    signingClient = client;

    const claimMsg = {
      typeUrl: '/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward',
      value: {
        delegatorAddress,
        validatorAddress,
      },
    };

    const result = await client.signAndBroadcast(
      delegatorAddress,
      [claimMsg],
      'auto',
      ''
    );

    await client.disconnect();
    signingClient = null;

    if (result.code === 0) {
      return {
        success: true,
        txHash: result.transactionHash,
        gasUsed: Number(result.gasUsed),
        height: Number(result.height),
      };
    } else {
      return {
        success: false,
        error: `Claim failed: ${result.rawLog}`,
      };
    }
  } catch (error) {
    if (signingClient) {
      try {
        await signingClient.disconnect();
      } catch {
        // Ignore disconnect errors
      }
    }

    const errorMessage = error instanceof Error ? error.message : 'Unknown error';

    // User-friendly error messages for reward claims
    if (errorMessage.includes('does not exist on chain') || errorMessage.includes('account not found')) {
      return {
        success: false,
        error: "Account 'does not exist on chain'. Send some tokens there before trying to query sequence."
      };
    }
    if (errorMessage.includes('account sequence mismatch')) {
      return { success: false, error: 'Transaction conflict. Please wait a moment and try again.' };
    }

    return { success: false, error: errorMessage };
  }
}

/**
 * Fetch transaction history for an address
 * Uses Tendermint RPC tx_search for reliable querying
 */
export async function fetchTransactionHistory(
  address: string,
  limit: number = 20
): Promise<{
  transactions: Array<{
    hash: string;
    type: 'SEND' | 'RECEIVE' | 'STAKE' | 'UNSTAKE' | 'CLAIM';
    amount: string;
    symbol: string;
    timestamp: number;
    height: number;
    counterparty?: string;
  }>;
  error?: string;
}> {
  // SECURITY: Rate limiting check
  if (!queryRateLimiter.check(`txHistory:${address}`)) {
    logger.warn('Rate limit exceeded for transaction history', { address });
    return { transactions: [], error: 'Rate limit exceeded' };
  }

  try {
    const transactions: Array<{
      hash: string;
      type: 'SEND' | 'RECEIVE' | 'STAKE' | 'UNSTAKE' | 'CLAIM';
      amount: string;
      symbol: string;
      timestamp: number;
      height: number;
      counterparty?: string;
    }> = [];

    // Use Tendermint RPC tx_search - more reliable than REST API
    // Query by transfer.sender (sent transactions)
    const sentQuery = encodeURIComponent(`"transfer.sender='${address}'"`);
    const sentResponse = await fetch(
      `${RPC_URL}/tx_search?query=${sentQuery}&per_page=${limit}&order_by="desc"`
    );

    if (sentResponse.ok) {
      const sentData = await sentResponse.json();
      for (const tx of sentData.result?.txs || []) {
        const txResult = await parseTendermintTx(tx, address, 'SEND');
        if (txResult) {
          transactions.push(txResult);
        }
      }
    }

    // Query by transfer.recipient (received transactions)
    const receivedQuery = encodeURIComponent(`"transfer.recipient='${address}'"`);
    const receivedResponse = await fetch(
      `${RPC_URL}/tx_search?query=${receivedQuery}&per_page=${limit}&order_by="desc"`
    );

    if (receivedResponse.ok) {
      const receivedData = await receivedResponse.json();
      for (const tx of receivedData.result?.txs || []) {
        // Skip if already added
        if (!transactions.some(t => t.hash === tx.hash)) {
          const txResult = await parseTendermintTx(tx, address, 'RECEIVE');
          if (txResult) {
            transactions.push(txResult);
          }
        }
      }
    }

    // Also query by message.sender for staking transactions
    const msgQuery = encodeURIComponent(`"message.sender='${address}'"`);
    const msgResponse = await fetch(
      `${RPC_URL}/tx_search?query=${msgQuery}&per_page=${limit}&order_by="desc"`
    );

    if (msgResponse.ok) {
      const msgData = await msgResponse.json();
      for (const tx of msgData.result?.txs || []) {
        // Skip if already added
        if (!transactions.some(t => t.hash === tx.hash)) {
          const txResult = await parseTendermintTx(tx, address, 'SEND');
          if (txResult) {
            transactions.push(txResult);
          }
        }
      }
    }

    // Sort by height descending (newest first)
    transactions.sort((a, b) => b.height - a.height);

    return { transactions: transactions.slice(0, limit) };
  } catch (error) {
    logger.error('Failed to fetch transaction history:', error);
    return { transactions: [], error: 'Failed to fetch transactions' };
  }
}

/**
 * Parse a Tendermint tx_search result into a transaction item
 */
async function parseTendermintTx(
  tx: { hash: string; height: string; tx_result?: { events?: Array<{ type: string; attributes: Array<{ key: string; value: string }> }> } },
  _userAddress: string,
  defaultType: 'SEND' | 'RECEIVE'
): Promise<{
  hash: string;
  type: 'SEND' | 'RECEIVE' | 'STAKE' | 'UNSTAKE' | 'CLAIM';
  amount: string;
  symbol: string;
  timestamp: number;
  height: number;
  counterparty?: string;
} | null> {
  try {
    const events = tx.tx_result?.events || [];
    let type: 'SEND' | 'RECEIVE' | 'STAKE' | 'UNSTAKE' | 'CLAIM' = defaultType;
    let amount = '0';
    let counterparty = '';

    // Look for message type in events
    const messageEvent = events.find(e => e.type === 'message');
    const messageAction = messageEvent?.attributes.find(a => {
      // Tendermint RPC returns base64 encoded values
      const decodedKey = atob(a.key);
      return decodedKey === 'action';
    });

    if (messageAction) {
      const action = atob(messageAction.value);
      if (action.includes('MsgDelegate')) type = 'STAKE';
      else if (action.includes('MsgUndelegate')) type = 'UNSTAKE';
      else if (action.includes('MsgWithdrawDelegatorReward')) type = 'CLAIM';
    }

    // Get amount from transfer event
    const transferEvent = events.find(e => e.type === 'transfer');
    if (transferEvent) {
      for (const attr of transferEvent.attributes) {
        const key = atob(attr.key);
        const value = atob(attr.value);
        if (key === 'amount') {
          // Parse amount like "100000000uhodl"
          const match = value.match(/(\d+)/);
          if (match) {
            amount = match[1];
          }
        }
        if (key === 'recipient' && defaultType === 'SEND') {
          counterparty = value;
        }
        if (key === 'sender' && defaultType === 'RECEIVE') {
          counterparty = value;
        }
      }
    }

    // For staking, get amount from delegate/unbond event
    if (type === 'STAKE' || type === 'UNSTAKE') {
      const stakingEvent = events.find(e => e.type === 'delegate' || e.type === 'unbond');
      if (stakingEvent) {
        for (const attr of stakingEvent.attributes) {
          const key = atob(attr.key);
          const value = atob(attr.value);
          if (key === 'amount') {
            const match = value.match(/(\d+)/);
            if (match) {
              amount = match[1];
            }
          }
          if (key === 'validator') {
            counterparty = value;
          }
        }
      }
    }

    const height = parseInt(tx.height);

    return {
      hash: tx.hash,
      type,
      amount: (parseInt(amount) / 1_000_000).toFixed(6),
      symbol: 'HODL',
      timestamp: Date.now() - (height * 2000), // Approximate timestamp (2s blocks)
      height,
      counterparty,
    };
  } catch {
    return null;
  }
}

/**
 * Validate a ShareHODL address
 */
export function validateAddress(address: string): { valid: boolean; error?: string } {
  if (!address) {
    return { valid: false, error: 'Address is required' };
  }

  // Check prefix
  if (!address.startsWith(ADDRESS_PREFIX)) {
    return { valid: false, error: `Address must start with '${ADDRESS_PREFIX}'` };
  }

  // Check length (hodl addresses are typically 43-45 chars)
  if (address.length < 40 || address.length > 60) {
    return { valid: false, error: 'Invalid address length' };
  }

  // Check for valid bech32 characters
  const bech32Chars = 'qpzry9x8gf2tvdw0s3jn54khce6mua7l';
  const dataPart = address.slice(ADDRESS_PREFIX.length + 1); // Skip prefix and separator
  for (const char of dataPart.toLowerCase()) {
    if (!bech32Chars.includes(char)) {
      return { valid: false, error: 'Invalid address characters' };
    }
  }

  return { valid: true };
}

/**
 * Format amount from micro units to display units
 */
export function formatAmount(microAmount: string | number): string {
  const amount = typeof microAmount === 'string' ? parseInt(microAmount) : microAmount;
  return (amount / 1_000_000).toFixed(6);
}

/**
 * Convert display amount to micro units
 */
export function toMicroAmount(amount: string | number): string {
  const value = typeof amount === 'string' ? parseFloat(amount) : amount;
  return Math.floor(value * 1_000_000).toString();
}
