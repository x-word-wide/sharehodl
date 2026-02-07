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

// SECURITY: Persistent rate limiter that survives page reloads
class RateLimiter {
  private limits: Map<string, RateLimitEntry> = new Map();
  private readonly maxRequests: number;
  private readonly windowMs: number;
  private readonly storageKey: string;

  constructor(maxRequests: number = 10, windowMs: number = 60000, storageKey: string = 'sh_rate_limit') {
    this.maxRequests = maxRequests;
    this.windowMs = windowMs;
    this.storageKey = storageKey;
    this.loadFromStorage();
  }

  /**
   * Load rate limit state from localStorage
   */
  private loadFromStorage(): void {
    try {
      const stored = localStorage.getItem(this.storageKey);
      if (stored) {
        const data = JSON.parse(stored) as Record<string, RateLimitEntry>;
        const now = Date.now();
        // Only load entries that haven't expired
        for (const [key, entry] of Object.entries(data)) {
          if (entry.resetTime > now) {
            this.limits.set(key, entry);
          }
        }
      }
    } catch {
      // Ignore storage errors
    }
  }

  /**
   * Save rate limit state to localStorage
   */
  private saveToStorage(): void {
    try {
      const data: Record<string, RateLimitEntry> = {};
      const now = Date.now();
      // Only save entries that haven't expired
      for (const [key, entry] of this.limits.entries()) {
        if (entry.resetTime > now) {
          data[key] = entry;
        }
      }
      localStorage.setItem(this.storageKey, JSON.stringify(data));
    } catch {
      // Ignore storage errors
    }
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
      this.saveToStorage();
      return true;
    }

    if (entry.count >= this.maxRequests) {
      return false;
    }

    entry.count++;
    this.saveToStorage();
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
const queryRateLimiter = new RateLimiter(30, 60000, 'sh_query_rate_limit');  // 30 queries per minute
const txRateLimiter = new RateLimiter(5, 60000, 'sh_tx_rate_limit');        // 5 transactions per minute

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
    logger.warn('Rate limit exceeded for transaction history');
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

    // Query by transfer.recipient (received transactions) - check this first
    const receivedQuery = encodeURIComponent(`"transfer.recipient='${address}'"`);
    const receivedUrl = `${RPC_URL}/tx_search?query=${receivedQuery}&per_page=${limit}&order_by="desc"`;

    const receivedResponse = await fetch(receivedUrl);

    if (receivedResponse.ok) {
      const receivedData = await receivedResponse.json();
      const txs = receivedData.result?.txs || [];

      for (const tx of txs) {
        const txResult = await parseTendermintTx(tx, address, 'RECEIVE');
        if (txResult) {
          transactions.push(txResult);
        }
      }
    } else {
      logger.warn('Failed to fetch received transactions');
    }

    // Query by transfer.sender (sent transactions)
    const sentQuery = encodeURIComponent(`"transfer.sender='${address}'"`);
    const sentResponse = await fetch(
      `${RPC_URL}/tx_search?query=${sentQuery}&per_page=${limit}&order_by="desc"`
    );

    if (sentResponse.ok) {
      const sentData = await sentResponse.json();
      for (const tx of sentData.result?.txs || []) {
        if (!transactions.some(t => t.hash === tx.hash)) {
          const txResult = await parseTendermintTx(tx, address, 'SEND');
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
    logger.error('Failed to fetch transaction history');
    return { transactions: [], error: 'Failed to fetch transactions' };
  }
}

/**
 * Parse a Tendermint tx_search result into a transaction item
 * Uses coin_received/coin_spent events which are more reliable than transfer events
 */
async function parseTendermintTx(
  tx: { hash: string; height: string; tx_result?: { events?: Array<{ type: string; attributes: Array<{ key: string; value: string; index?: boolean }> }> } },
  userAddress: string,
  defaultType: 'SEND' | 'RECEIVE'
): Promise<{
  hash: string;
  type: 'SEND' | 'RECEIVE' | 'STAKE' | 'UNSTAKE' | 'CLAIM';
  amount: string;
  symbol: string;
  timestamp: number;
  height: number;
  counterparty?: string;
  fee?: string;
} | null> {
  try {
    const events = tx.tx_result?.events || [];
    let type: 'SEND' | 'RECEIVE' | 'STAKE' | 'UNSTAKE' | 'CLAIM' = defaultType;
    let amount = '0';
    let fee = '0';
    let counterparty = '';
    const userAddrLower = userAddress.toLowerCase();

    // Extract fee from tx_fee or fee event
    const feeEvent = events.find(e => e.type === 'tx' || e.type === 'fee');
    if (feeEvent) {
      const feeAttr = feeEvent.attributes.find(a => a.key === 'fee');
      if (feeAttr) {
        const match = feeAttr.value.match(/(\d+)/);
        if (match) {
          fee = (parseInt(match[1]) / 1_000_000).toFixed(4);
        }
      }
    }

    // Look for message type in events
    const messageEvent = events.find(e => e.type === 'message');
    if (messageEvent) {
      const messageAction = messageEvent.attributes.find(a => a.key === 'action');
      if (messageAction) {
        const action = messageAction.value;
        if (action.includes('MsgDelegate')) type = 'STAKE';
        else if (action.includes('MsgUndelegate')) type = 'UNSTAKE';
        else if (action.includes('MsgWithdrawDelegatorReward')) type = 'CLAIM';
      }
    }

    // Strategy 1: Use coin_received for RECEIVE, coin_spent for SEND
    // These events are more reliable and directly tied to the user's address
    if (defaultType === 'RECEIVE') {
      const coinReceivedEvents = events.filter(e => e.type === 'coin_received');
      for (const event of coinReceivedEvents) {
        const attrs: Record<string, string> = {};
        for (const attr of event.attributes) {
          attrs[attr.key] = attr.value;
        }
        if (attrs['receiver']?.toLowerCase() === userAddrLower && attrs['amount']) {
          const match = attrs['amount'].match(/(\d+)/);
          if (match) {
            const parsedAmt = parseInt(match[1]);
            // Take the largest amount (skip fee-sized amounts < 100000 = 0.1 HODL)
            if (parsedAmt > parseInt(amount) && parsedAmt > 100000) {
              amount = match[1];
            }
          }
        }
      }
    } else {
      // For SEND, use coin_spent but exclude fees (small amounts)
      const coinSpentEvents = events.filter(e => e.type === 'coin_spent');
      for (const event of coinSpentEvents) {
        const attrs: Record<string, string> = {};
        for (const attr of event.attributes) {
          attrs[attr.key] = attr.value;
        }
        if (attrs['spender']?.toLowerCase() === userAddrLower && attrs['amount']) {
          const match = attrs['amount'].match(/(\d+)/);
          if (match) {
            const parsedAmt = parseInt(match[1]);
            // Take the largest amount (the actual transfer, not fees)
            if (parsedAmt > parseInt(amount)) {
              amount = match[1];
            }
          }
        }
      }
    }

    // Strategy 2: Fallback to transfer events if coin_received/spent didn't work
    if (amount === '0') {
      const transferEvents = events.filter(e => e.type === 'transfer');
      let maxAmount = 0;

      for (const transferEvent of transferEvents) {
        const attrs: Record<string, string> = {};
        for (const attr of transferEvent.attributes) {
          attrs[attr.key] = attr.value;
        }

        const recipient = attrs['recipient'] || '';
        const sender = attrs['sender'] || '';

        const isUserTransfer = defaultType === 'RECEIVE'
          ? recipient.toLowerCase() === userAddrLower
          : sender.toLowerCase() === userAddrLower;

        if (isUserTransfer && attrs['amount']) {
          const match = attrs['amount'].match(/(\d+)/);
          if (match) {
            const parsedAmt = parseInt(match[1]);
            if (parsedAmt > maxAmount) {
              maxAmount = parsedAmt;
              amount = match[1];
              counterparty = defaultType === 'SEND' ? recipient : sender;
            }
          }
        }
      }
    }

    // For staking, get amount from delegate/unbond event
    if (type === 'STAKE' || type === 'UNSTAKE') {
      const stakingEvent = events.find(e => e.type === 'delegate' || e.type === 'unbond');
      if (stakingEvent) {
        for (const attr of stakingEvent.attributes) {
          if (attr.key === 'amount') {
            const match = attr.value.match(/(\d+)/);
            if (match) {
              amount = match[1];
            }
          }
          if (attr.key === 'validator') {
            counterparty = attr.value;
          }
        }
      }
    }

    const height = parseInt(tx.height);
    const parsedAmount = parseInt(amount) / 1_000_000;

    // Format amount nicely: no decimals for whole numbers, max 2 for others
    const formattedAmount = parsedAmount >= 1
      ? (Number.isInteger(parsedAmount) ? parsedAmount.toString() : parsedAmount.toFixed(2))
      : parsedAmount.toFixed(4);

    return {
      hash: tx.hash,
      type,
      amount: formattedAmount,
      symbol: 'HODL',
      timestamp: Date.now() - (height * 2000), // Approximate timestamp (2s blocks)
      height,
      counterparty,
      fee: fee !== '0' ? fee : undefined,
    };
  } catch (error) {
    logger.error('Error parsing transaction');
    return null;
  }
}

/**
 * Validate a ShareHODL address (bech32 format)
 */
export function validateAddress(address: string): { valid: boolean; error?: string } {
  if (!address) {
    return { valid: false, error: 'Address is required' };
  }

  // Trim any whitespace
  const cleanAddress = address.trim();

  // Check prefix
  if (!cleanAddress.startsWith(ADDRESS_PREFIX)) {
    return { valid: false, error: `Invalid address. Must start with '${ADDRESS_PREFIX}'` };
  }

  // Check for bech32 separator (must have '1' after prefix)
  const separatorIndex = cleanAddress.indexOf('1');
  if (separatorIndex !== ADDRESS_PREFIX.length) {
    return { valid: false, error: 'Invalid address format' };
  }

  // Check length (hodl addresses are typically 43-45 chars: hodl1 + 38-40 chars)
  if (cleanAddress.length < 43 || cleanAddress.length > 50) {
    return { valid: false, error: `Invalid address length (${cleanAddress.length} chars)` };
  }

  // Check for valid bech32 characters in data part
  const bech32Chars = 'qpzry9x8gf2tvdw0s3jn54khce6mua7l';
  const dataPart = cleanAddress.slice(separatorIndex + 1); // Skip prefix and separator

  if (dataPart.length === 0) {
    return { valid: false, error: 'Invalid address - missing data' };
  }

  for (const char of dataPart.toLowerCase()) {
    if (!bech32Chars.includes(char)) {
      return { valid: false, error: `Invalid character '${char}' in address` };
    }
  }

  // Check it doesn't contain mixed case (bech32 should be all lowercase or all uppercase)
  const hasUpper = /[A-Z]/.test(dataPart);
  const hasLower = /[a-z]/.test(dataPart);
  if (hasUpper && hasLower) {
    return { valid: false, error: 'Address has mixed case - invalid format' };
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
