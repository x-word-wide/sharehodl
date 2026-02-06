/**
 * ShareHODL Telegram Mini App Types
 */

// ============================================
// Chain & Network Types
// ============================================

export enum Chain {
  SHAREHODL = 'SHAREHODL',
  ETHEREUM = 'ETHEREUM',
  BITCOIN = 'BITCOIN',
  COSMOS = 'COSMOS',
  OSMOSIS = 'OSMOSIS',
  POLYGON = 'POLYGON',
  ARBITRUM = 'ARBITRUM',
  OPTIMISM = 'OPTIMISM',
  BASE = 'BASE',
  AVALANCHE = 'AVALANCHE',
  BNB = 'BNB',
  SOLANA = 'SOLANA',
  CELESTIA = 'CELESTIA'
}

export interface ChainConfig {
  chain: Chain;
  name: string;
  symbol: string;
  decimals: number;
  color: string;
  coinType: number;  // BIP44 coin type
  rpcUrl?: string;
  restUrl?: string;
  explorerUrl?: string;
  isTestnet?: boolean;
}

export const CHAIN_CONFIGS: Record<Chain, ChainConfig> = {
  [Chain.SHAREHODL]: {
    chain: Chain.SHAREHODL,
    name: 'ShareHODL',
    symbol: 'HODL',
    decimals: 6,
    color: '#1E40AF',
    coinType: 118,  // Cosmos
    rpcUrl: 'https://rpc.sharehodl.network',
    restUrl: 'https://api.sharehodl.network',
    explorerUrl: 'https://explorer.sharehodl.network'
  },
  [Chain.ETHEREUM]: {
    chain: Chain.ETHEREUM,
    name: 'Ethereum',
    symbol: 'ETH',
    decimals: 18,
    color: '#627EEA',
    coinType: 60
  },
  [Chain.BITCOIN]: {
    chain: Chain.BITCOIN,
    name: 'Bitcoin',
    symbol: 'BTC',
    decimals: 8,
    color: '#F7931A',
    coinType: 0
  },
  [Chain.COSMOS]: {
    chain: Chain.COSMOS,
    name: 'Cosmos Hub',
    symbol: 'ATOM',
    decimals: 6,
    color: '#2E3148',
    coinType: 118
  },
  [Chain.OSMOSIS]: {
    chain: Chain.OSMOSIS,
    name: 'Osmosis',
    symbol: 'OSMO',
    decimals: 6,
    color: '#5E12A0',
    coinType: 118
  },
  [Chain.POLYGON]: {
    chain: Chain.POLYGON,
    name: 'Polygon',
    symbol: 'POL',
    decimals: 18,
    color: '#8247E5',
    coinType: 60
  },
  [Chain.ARBITRUM]: {
    chain: Chain.ARBITRUM,
    name: 'Arbitrum',
    symbol: 'ETH',
    decimals: 18,
    color: '#28A0F0',
    coinType: 60
  },
  [Chain.OPTIMISM]: {
    chain: Chain.OPTIMISM,
    name: 'Optimism',
    symbol: 'ETH',
    decimals: 18,
    color: '#FF0420',
    coinType: 60
  },
  [Chain.BASE]: {
    chain: Chain.BASE,
    name: 'Base',
    symbol: 'ETH',
    decimals: 18,
    color: '#0052FF',
    coinType: 60
  },
  [Chain.AVALANCHE]: {
    chain: Chain.AVALANCHE,
    name: 'Avalanche',
    symbol: 'AVAX',
    decimals: 18,
    color: '#E84142',
    coinType: 60
  },
  [Chain.BNB]: {
    chain: Chain.BNB,
    name: 'BNB Chain',
    symbol: 'BNB',
    decimals: 18,
    color: '#F0B90B',
    coinType: 60
  },
  [Chain.SOLANA]: {
    chain: Chain.SOLANA,
    name: 'Solana',
    symbol: 'SOL',
    decimals: 9,
    color: '#00FFA3',
    coinType: 501
  },
  [Chain.CELESTIA]: {
    chain: Chain.CELESTIA,
    name: 'Celestia',
    symbol: 'TIA',
    decimals: 6,
    color: '#7B2BF9',
    coinType: 118
  }
};

// ============================================
// Token Types (ERC-20, BEP-20, etc.)
// ============================================

export enum TokenType {
  NATIVE = 'NATIVE',    // Native chain token (ETH, BTC, etc.)
  ERC20 = 'ERC20',      // Ethereum tokens
  BEP20 = 'BEP20',      // BNB Chain tokens
  SPL = 'SPL',          // Solana tokens
  CW20 = 'CW20',        // Cosmos tokens
}

export interface Token {
  id: string;           // Unique identifier
  symbol: string;
  name: string;
  chain: Chain;
  type: TokenType;
  decimals: number;
  contractAddress?: string;  // For non-native tokens
  logoUrl?: string;
  color: string;
  coingeckoId?: string;      // For price fetching
}

// Pre-defined popular tokens
export const TOKENS: Token[] = [
  // Native tokens
  {
    id: 'btc',
    symbol: 'BTC',
    name: 'Bitcoin',
    chain: Chain.BITCOIN,
    type: TokenType.NATIVE,
    decimals: 8,
    color: '#F7931A',
    coingeckoId: 'bitcoin'
  },
  {
    id: 'eth',
    symbol: 'ETH',
    name: 'Ethereum',
    chain: Chain.ETHEREUM,
    type: TokenType.NATIVE,
    decimals: 18,
    color: '#627EEA',
    coingeckoId: 'ethereum'
  },
  {
    id: 'matic',
    symbol: 'POL',
    name: 'Polygon',
    chain: Chain.POLYGON,
    type: TokenType.NATIVE,
    decimals: 18,
    color: '#8247E5',
    coingeckoId: 'matic-network'
  },
  {
    id: 'bnb',
    symbol: 'BNB',
    name: 'BNB',
    chain: Chain.BNB,
    type: TokenType.NATIVE,
    decimals: 18,
    color: '#F0B90B',
    coingeckoId: 'binancecoin'
  },
  {
    id: 'sol',
    symbol: 'SOL',
    name: 'Solana',
    chain: Chain.SOLANA,
    type: TokenType.NATIVE,
    decimals: 9,
    color: '#00FFA3',
    coingeckoId: 'solana'
  },
  {
    id: 'avax',
    symbol: 'AVAX',
    name: 'Avalanche',
    chain: Chain.AVALANCHE,
    type: TokenType.NATIVE,
    decimals: 18,
    color: '#E84142',
    coingeckoId: 'avalanche-2'
  },
  {
    id: 'atom',
    symbol: 'ATOM',
    name: 'Cosmos Hub',
    chain: Chain.COSMOS,
    type: TokenType.NATIVE,
    decimals: 6,
    color: '#2E3148',
    coingeckoId: 'cosmos'
  },
  {
    id: 'hodl',
    symbol: 'HODL',
    name: 'ShareHODL',
    chain: Chain.SHAREHODL,
    type: TokenType.NATIVE,
    decimals: 6,
    color: '#1E40AF',
  },
  // Stablecoins - Ethereum
  {
    id: 'usdt-eth',
    symbol: 'USDT',
    name: 'Tether USD',
    chain: Chain.ETHEREUM,
    type: TokenType.ERC20,
    decimals: 6,
    contractAddress: '0xdAC17F958D2ee523a2206206994597C13D831ec7',
    color: '#26A17B',
    coingeckoId: 'tether'
  },
  {
    id: 'usdc-eth',
    symbol: 'USDC',
    name: 'USD Coin',
    chain: Chain.ETHEREUM,
    type: TokenType.ERC20,
    decimals: 6,
    contractAddress: '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48',
    color: '#2775CA',
    coingeckoId: 'usd-coin'
  },
  // Stablecoins - Polygon
  {
    id: 'usdt-polygon',
    symbol: 'USDT',
    name: 'Tether USD',
    chain: Chain.POLYGON,
    type: TokenType.ERC20,
    decimals: 6,
    contractAddress: '0xc2132D05D31c914a87C6611C10748AEb04B58e8F',
    color: '#26A17B',
    coingeckoId: 'tether'
  },
  {
    id: 'usdc-polygon',
    symbol: 'USDC',
    name: 'USD Coin',
    chain: Chain.POLYGON,
    type: TokenType.ERC20,
    decimals: 6,
    contractAddress: '0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359',
    color: '#2775CA',
    coingeckoId: 'usd-coin'
  },
  // Stablecoins - BNB Chain
  {
    id: 'usdt-bsc',
    symbol: 'USDT',
    name: 'Tether USD',
    chain: Chain.BNB,
    type: TokenType.BEP20,
    decimals: 18,
    contractAddress: '0x55d398326f99059fF775485246999027B3197955',
    color: '#26A17B',
    coingeckoId: 'tether'
  },
  {
    id: 'usdc-bsc',
    symbol: 'USDC',
    name: 'USD Coin',
    chain: Chain.BNB,
    type: TokenType.BEP20,
    decimals: 18,
    contractAddress: '0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d',
    color: '#2775CA',
    coingeckoId: 'usd-coin'
  },
  // Stablecoins - Arbitrum
  {
    id: 'usdt-arb',
    symbol: 'USDT',
    name: 'Tether USD',
    chain: Chain.ARBITRUM,
    type: TokenType.ERC20,
    decimals: 6,
    contractAddress: '0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9',
    color: '#26A17B',
    coingeckoId: 'tether'
  },
  {
    id: 'usdc-arb',
    symbol: 'USDC',
    name: 'USD Coin',
    chain: Chain.ARBITRUM,
    type: TokenType.ERC20,
    decimals: 6,
    contractAddress: '0xaf88d065e77c8cC2239327C5EDb3A432268e5831',
    color: '#2775CA',
    coingeckoId: 'usd-coin'
  },
  // Stablecoins - Base
  {
    id: 'usdc-base',
    symbol: 'USDC',
    name: 'USD Coin',
    chain: Chain.BASE,
    type: TokenType.ERC20,
    decimals: 6,
    contractAddress: '0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913',
    color: '#2775CA',
    coingeckoId: 'usd-coin'
  },
  // Stablecoins - Avalanche
  {
    id: 'usdt-avax',
    symbol: 'USDT',
    name: 'Tether USD',
    chain: Chain.AVALANCHE,
    type: TokenType.ERC20,
    decimals: 6,
    contractAddress: '0x9702230A8Ea53601f5cD2dc00fDBc13d4dF4A8c7',
    color: '#26A17B',
    coingeckoId: 'tether'
  },
  {
    id: 'usdc-avax',
    symbol: 'USDC',
    name: 'USD Coin',
    chain: Chain.AVALANCHE,
    type: TokenType.ERC20,
    decimals: 6,
    contractAddress: '0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E',
    color: '#2775CA',
    coingeckoId: 'usd-coin'
  },
];

// Helper to get token by ID
export function getTokenById(id: string): Token | undefined {
  return TOKENS.find(t => t.id === id);
}

// Helper to get all tokens for a chain
export function getTokensForChain(chain: Chain): Token[] {
  return TOKENS.filter(t => t.chain === chain);
}

// Helper to get native token for chain
export function getNativeToken(chain: Chain): Token | undefined {
  return TOKENS.find(t => t.chain === chain && t.type === TokenType.NATIVE);
}

// ============================================
// Wallet Types
// ============================================

export interface WalletAccount {
  chain: Chain;
  address: string;
  balance: string;
  balanceUsd?: string;
  derivationPath: string;
}

// Asset holding in wallet (like Trust Wallet)
export interface AssetHolding {
  token: Token;
  balance: string;          // Raw balance
  balanceFormatted: string; // Human readable
  balanceUsd: string;
  price: number;
  priceChange24h: number;   // Percentage
  address: string;          // Wallet address for this chain
}

export interface WalletState {
  isInitialized: boolean;
  isLocked: boolean;
  accounts: WalletAccount[];
  assets: AssetHolding[];   // All token holdings
  totalBalanceUsd: string;
}

// ============================================
// Equity Types
// ============================================

export interface Equity {
  symbol: string;
  companyName: string;
  sector: EquitySector;
  price: number;
  change24h: number;
  changePercent24h: number;
  marketCap: number;
  volume24h: number;
  description?: string;
}

export enum EquitySector {
  TECHNOLOGY = 'TECHNOLOGY',
  HEALTHCARE = 'HEALTHCARE',
  FINANCE = 'FINANCE',
  ENERGY = 'ENERGY',
  CONSUMER = 'CONSUMER',
  INDUSTRIAL = 'INDUSTRIAL',
  REAL_ESTATE = 'REAL_ESTATE',
  UTILITIES = 'UTILITIES',
  MATERIALS = 'MATERIALS',
  COMMUNICATION = 'COMMUNICATION'
}

export const SECTOR_COLORS: Record<EquitySector, string> = {
  [EquitySector.TECHNOLOGY]: '#1E40AF',
  [EquitySector.HEALTHCARE]: '#10B981',
  [EquitySector.FINANCE]: '#F59E0B',
  [EquitySector.ENERGY]: '#EF4444',
  [EquitySector.CONSUMER]: '#3B82F6',
  [EquitySector.INDUSTRIAL]: '#64748B',
  [EquitySector.REAL_ESTATE]: '#06B6D4',
  [EquitySector.UTILITIES]: '#84CC16',
  [EquitySector.MATERIALS]: '#F97316',
  [EquitySector.COMMUNICATION]: '#EC4899'
};

export interface EquityHolding {
  equity: Equity;
  shares: number;
  avgCost: number;
  currentValue: number;
  gainLoss: number;
  gainLossPercent: number;
}

// ============================================
// P2P Trading Types
// ============================================

export interface P2PListing {
  id: string;
  type: 'BUY' | 'SELL';
  asset: string;
  price: number;
  minAmount: number;
  maxAmount: number;
  availableAmount: number;
  paymentMethods: string[];
  trader: P2PTrader;
  createdAt: number;
}

export interface P2PTrader {
  id: string;
  name: string;
  completedTrades: number;
  rating: number;
  isVerified: boolean;
}

export interface P2POrder {
  id: string;
  listingId: string;
  amount: number;
  price: number;
  status: 'PENDING' | 'PAID' | 'COMPLETED' | 'CANCELLED' | 'DISPUTED';
  createdAt: number;
}

// ============================================
// Lending Types
// ============================================

export interface LendingMarket {
  asset: string;
  totalSupply: number;
  totalBorrow: number;
  supplyApy: number;
  borrowApr: number;
  utilizationRate: number;
  collateralFactor: number;
}

export interface LendingPosition {
  asset: string;
  type: 'SUPPLY' | 'BORROW';
  amount: number;
  apy: number;
  earnedOrOwed: number;
}

// ============================================
// Inheritance Types
// ============================================

export interface Beneficiary {
  id: string;
  name: string;
  address: string;
  allocationPercent: number;
}

export interface InheritancePlan {
  id: string;
  type: 'DEAD_MAN_SWITCH' | 'MULTI_SIG' | 'TIME_LOCKED';
  beneficiaries: Beneficiary[];
  assets: string[];
  status: 'ACTIVE' | 'TRIGGERED' | 'EXECUTED';
  lastCheckIn?: number;
  triggerDate?: number;
}

// ============================================
// Transaction Types
// ============================================

export interface Transaction {
  hash: string;
  type: 'SEND' | 'RECEIVE' | 'SWAP' | 'STAKE' | 'TRADE';
  chain: Chain;
  from: string;
  to: string;
  amount: string;
  symbol: string;
  fee?: string;
  status: 'PENDING' | 'SUCCESS' | 'FAILED';
  timestamp: number;
}

// ============================================
// Bridge Types
// ============================================

export interface BridgeQuote {
  fromChain: Chain;
  toChain: Chain;
  fromAmount: string;
  toAmount: string;
  fee: string;
  estimatedTime: number;
}

export interface BridgeTransaction {
  id: string;
  quote: BridgeQuote;
  status: 'PENDING' | 'CONFIRMING' | 'COMPLETED' | 'FAILED';
  txHash?: string;
  createdAt: number;
}

// ============================================
// Staking Types
// ============================================

/**
 * User staking tiers based on HODL holdings
 * Higher tiers get better rewards and platform benefits
 */
export enum StakingTier {
  NONE = 'NONE',
  HOLDER = 'HOLDER',       // 100 HODL
  KEEPER = 'KEEPER',       // 10,000 HODL
  WARDEN = 'WARDEN',       // 100,000 HODL
  STEWARD = 'STEWARD',     // 1,000,000 HODL
  ARCHON = 'ARCHON',       // 10,000,000 HODL
  VALIDATOR = 'VALIDATOR'  // 50,000,000 HODL (can run validator node)
}

/**
 * Validator sub-tiers for those who run validator nodes
 */
export enum ValidatorTier {
  BRONZE = 'BRONZE',       // 50,000 HODL min stake
  SILVER = 'SILVER',       // 100,000 HODL
  GOLD = 'GOLD',           // 250,000 HODL
  PLATINUM = 'PLATINUM',   // 500,000 HODL
  DIAMOND = 'DIAMOND'      // 1,000,000 HODL
}

export interface StakingTierConfig {
  tier: StakingTier;
  name: string;
  minStake: number;        // in HODL
  rewardMultiplier: number;
  color: string;
  icon: string;
  benefits: string[];
}

export const STAKING_TIERS: StakingTierConfig[] = [
  {
    tier: StakingTier.NONE,
    name: 'No Stake',
    minStake: 0,
    rewardMultiplier: 0,
    color: '#64748B',
    icon: '○',
    benefits: []
  },
  {
    tier: StakingTier.HOLDER,
    name: 'Holder',
    minStake: 100,
    rewardMultiplier: 1.0,
    color: '#10B981',
    icon: '◐',
    benefits: ['Basic staking rewards', 'P2P trading access']
  },
  {
    tier: StakingTier.KEEPER,
    name: 'Keeper',
    minStake: 10_000,
    rewardMultiplier: 1.5,
    color: '#1E40AF',
    icon: '◑',
    benefits: ['1.5x reward boost', 'Reduced trading fees', 'Priority support']
  },
  {
    tier: StakingTier.WARDEN,
    name: 'Warden',
    minStake: 100_000,
    rewardMultiplier: 2.0,
    color: '#3B82F6',
    icon: '◕',
    benefits: ['2x reward boost', 'Zero trading fees', 'Governance voting', 'Early access features']
  },
  {
    tier: StakingTier.STEWARD,
    name: 'Steward',
    minStake: 1_000_000,
    rewardMultiplier: 2.5,
    color: '#F59E0B',
    icon: '⬤',
    benefits: ['2.5x reward boost', 'VIP support', 'Proposal creation', 'Exclusive airdrops']
  },
  {
    tier: StakingTier.ARCHON,
    name: 'Archon',
    minStake: 10_000_000,
    rewardMultiplier: 3.0,
    color: '#EF4444',
    icon: '◈',
    benefits: ['3x reward boost', 'Council membership', 'Protocol governance', 'Revenue sharing']
  },
  {
    tier: StakingTier.VALIDATOR,
    name: 'Validator',
    minStake: 50_000_000,
    rewardMultiplier: 4.0,
    color: '#EC4899',
    icon: '◆',
    benefits: ['4x reward boost', 'Run validator node', 'Block rewards', 'Network governance']
  }
];

export interface StakingPosition {
  stakedAmount: number;      // in HODL
  pendingRewards: number;    // unclaimed rewards
  tier: StakingTier;
  tierConfig: StakingTierConfig;
  delegations: Delegation[];
  unbondings: Unbonding[];
  apr: number;               // current APR
  nextTier?: StakingTierConfig;
  nextTierProgress?: number; // 0-100%
}

export interface Delegation {
  validatorAddress: string;
  validatorName: string;
  validatorTier: ValidatorTier;
  amount: number;
  rewards: number;
  commission: number;
}

export interface Unbonding {
  validatorAddress: string;
  amount: number;
  completionTime: number;
}

export interface Validator {
  address: string;
  name: string;
  description?: string;
  website?: string;
  commission: number;
  tier: ValidatorTier;
  totalStaked: number;
  delegatorCount: number;
  uptime: number;
  isJailed: boolean;
  votingPower: number;
}

export const VALIDATOR_TIER_COLORS: Record<ValidatorTier, string> = {
  [ValidatorTier.BRONZE]: '#CD7F32',
  [ValidatorTier.SILVER]: '#C0C0C0',
  [ValidatorTier.GOLD]: '#FFD700',
  [ValidatorTier.PLATINUM]: '#E5E4E2',
  [ValidatorTier.DIAMOND]: '#B9F2FF'
};
