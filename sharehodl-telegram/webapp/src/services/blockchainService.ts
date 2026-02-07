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
 */
export async function fetchBalance(address: string): Promise<BalanceResult> {
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
 */
export async function sendTokens(
  mnemonic: string,
  recipientAddress: string,
  amount: string,
  memo?: string
): Promise<TransactionResult> {
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
      return { success: false, error: 'Transaction conflict. Please try again.' };
    }
    if (errorMessage.includes('decoding bech32 failed')) {
      return { success: false, error: 'Invalid recipient address format' };
    }

    return { success: false, error: errorMessage };
  }
}

/**
 * Delegate tokens to a validator
 */
export async function delegateTokens(
  mnemonic: string,
  validatorAddress: string,
  amount: string
): Promise<TransactionResult> {
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
    return { success: false, error: errorMessage };
  }
}

/**
 * Undelegate tokens from a validator
 */
export async function undelegateTokens(
  mnemonic: string,
  validatorAddress: string,
  amount: string
): Promise<TransactionResult> {
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
    return { success: false, error: errorMessage };
  }
}

/**
 * Claim staking rewards from a validator
 */
export async function claimRewards(
  mnemonic: string,
  validatorAddress: string
): Promise<TransactionResult> {
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
    return { success: false, error: errorMessage };
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
