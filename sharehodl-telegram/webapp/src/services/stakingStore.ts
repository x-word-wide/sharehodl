/**
 * Staking State Management
 *
 * Manages staking positions, delegations, rewards, and tier calculations
 */

import { create } from 'zustand';
import {
  StakingTierConfig,
  StakingPosition,
  Delegation,
  Unbonding,
  Validator,
  ValidatorTier,
  STAKING_TIERS
} from '../types';

// API base URL (will be configured for production)
// TODO: Replace demo data with actual API calls
export const API_BASE = import.meta.env.VITE_API_URL || 'https://api.sharehodl.network';

interface StakingStore {
  // State
  position: StakingPosition | null;
  validators: Validator[];
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchStakingPosition: (address: string) => Promise<void>;
  fetchValidators: () => Promise<void>;
  delegate: (validatorAddress: string, amount: number) => Promise<void>;
  undelegate: (validatorAddress: string, amount: number) => Promise<void>;
  redelegate: (fromValidator: string, toValidator: string, amount: number) => Promise<void>;
  claimRewards: (validatorAddress?: string) => Promise<void>;
  claimAllRewards: () => Promise<void>;
  clearError: () => void;
}

/**
 * Calculate staking tier based on staked amount
 */
function calculateTier(stakedAmount: number): StakingTierConfig {
  // Find the highest tier the user qualifies for
  for (let i = STAKING_TIERS.length - 1; i >= 0; i--) {
    if (stakedAmount >= STAKING_TIERS[i].minStake) {
      return STAKING_TIERS[i];
    }
  }
  return STAKING_TIERS[0]; // NONE tier
}

/**
 * Calculate progress to next tier
 */
function calculateNextTierProgress(stakedAmount: number, currentTier: StakingTierConfig): { nextTier?: StakingTierConfig; progress: number } {
  const currentIndex = STAKING_TIERS.findIndex(t => t.tier === currentTier.tier);

  if (currentIndex === STAKING_TIERS.length - 1) {
    // Already at max tier
    return { progress: 100 };
  }

  const nextTier = STAKING_TIERS[currentIndex + 1];
  const currentMin = currentTier.minStake;
  const nextMin = nextTier.minStake;

  const progress = Math.min(100, ((stakedAmount - currentMin) / (nextMin - currentMin)) * 100);

  return { nextTier, progress };
}

export const useStakingStore = create<StakingStore>((set, get) => ({
  position: null,
  validators: [],
  isLoading: false,
  error: null,

  // Fetch user's staking position
  fetchStakingPosition: async (address: string) => {
    set({ isLoading: true, error: null });

    try {
      // In production, fetch from blockchain API
      // For now, use demo data
      const demoPosition = createDemoPosition(address);
      set({ position: demoPosition, isLoading: false });

      // TODO: Replace with actual API call
      // const response = await fetch(`${API_BASE}/cosmos/staking/v1beta1/delegations/${address}`);
      // const data = await response.json();
      // ...process data
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to fetch staking position';
      set({ isLoading: false, error: message });
    }
  },

  // Fetch list of validators
  fetchValidators: async () => {
    set({ isLoading: true, error: null });

    try {
      // In production, fetch from blockchain API
      const demoValidators = createDemoValidators();
      set({ validators: demoValidators, isLoading: false });

      // TODO: Replace with actual API call
      // const response = await fetch(`${API_BASE}/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED`);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to fetch validators';
      set({ isLoading: false, error: message });
    }
  },

  // Delegate to a validator
  delegate: async (validatorAddress: string, amount: number) => {
    set({ isLoading: true, error: null });

    try {
      // TODO: Implement actual delegation transaction
      // This would require:
      // 1. Get mnemonic from encrypted storage
      // 2. Create delegation message
      // 3. Sign and broadcast transaction

      // For demo, update local state
      const { position, validators } = get();
      if (position) {
        const validator = validators.find(v => v.address === validatorAddress);
        const existingDelegation = position.delegations.find(d => d.validatorAddress === validatorAddress);

        let newDelegations: Delegation[];
        if (existingDelegation) {
          newDelegations = position.delegations.map(d =>
            d.validatorAddress === validatorAddress
              ? { ...d, amount: d.amount + amount }
              : d
          );
        } else {
          newDelegations = [
            ...position.delegations,
            {
              validatorAddress,
              validatorName: validator?.name || 'Unknown Validator',
              validatorTier: validator?.tier || ValidatorTier.BRONZE,
              amount,
              rewards: 0,
              commission: validator?.commission || 0.05
            }
          ];
        }

        const newStakedAmount = position.stakedAmount + amount;
        const tierConfig = calculateTier(newStakedAmount);
        const { nextTier, progress } = calculateNextTierProgress(newStakedAmount, tierConfig);

        set({
          position: {
            ...position,
            stakedAmount: newStakedAmount,
            tier: tierConfig.tier,
            tierConfig,
            delegations: newDelegations,
            nextTier,
            nextTierProgress: progress
          },
          isLoading: false
        });
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to delegate';
      set({ isLoading: false, error: message });
      throw error;
    }
  },

  // Undelegate from a validator
  undelegate: async (validatorAddress: string, amount: number) => {
    set({ isLoading: true, error: null });

    try {
      const { position } = get();
      if (position) {
        // Create unbonding entry (21 day unbonding period)
        const completionTime = Date.now() + 21 * 24 * 60 * 60 * 1000;
        const newUnbonding: Unbonding = {
          validatorAddress,
          amount,
          completionTime
        };

        // Update delegation
        const newDelegations = position.delegations
          .map(d => d.validatorAddress === validatorAddress
            ? { ...d, amount: d.amount - amount }
            : d
          )
          .filter(d => d.amount > 0);

        const newStakedAmount = position.stakedAmount - amount;
        const tierConfig = calculateTier(newStakedAmount);
        const { nextTier, progress } = calculateNextTierProgress(newStakedAmount, tierConfig);

        set({
          position: {
            ...position,
            stakedAmount: newStakedAmount,
            tier: tierConfig.tier,
            tierConfig,
            delegations: newDelegations,
            unbondings: [...position.unbondings, newUnbonding],
            nextTier,
            nextTierProgress: progress
          },
          isLoading: false
        });
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to undelegate';
      set({ isLoading: false, error: message });
      throw error;
    }
  },

  // Redelegate between validators
  redelegate: async (fromValidator: string, toValidator: string, amount: number) => {
    set({ isLoading: true, error: null });

    try {
      const { position, validators } = get();
      if (position) {
        const toValidatorInfo = validators.find(v => v.address === toValidator);

        const newDelegations = position.delegations
          .map(d => {
            if (d.validatorAddress === fromValidator) {
              return { ...d, amount: d.amount - amount };
            }
            if (d.validatorAddress === toValidator) {
              return { ...d, amount: d.amount + amount };
            }
            return d;
          })
          .filter(d => d.amount > 0);

        // Add new delegation if didn't exist
        if (!position.delegations.find(d => d.validatorAddress === toValidator)) {
          newDelegations.push({
            validatorAddress: toValidator,
            validatorName: toValidatorInfo?.name || 'Unknown Validator',
            validatorTier: toValidatorInfo?.tier || ValidatorTier.BRONZE,
            amount,
            rewards: 0,
            commission: toValidatorInfo?.commission || 0.05
          });
        }

        set({
          position: { ...position, delegations: newDelegations },
          isLoading: false
        });
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to redelegate';
      set({ isLoading: false, error: message });
      throw error;
    }
  },

  // Claim rewards from a specific validator
  claimRewards: async (validatorAddress?: string) => {
    set({ isLoading: true, error: null });

    try {
      const { position } = get();
      if (position) {
        let claimedAmount = 0;

        const newDelegations = position.delegations.map(d => {
          if (!validatorAddress || d.validatorAddress === validatorAddress) {
            claimedAmount += d.rewards;
            return { ...d, rewards: 0 };
          }
          return d;
        });

        set({
          position: {
            ...position,
            delegations: newDelegations,
            pendingRewards: position.pendingRewards - claimedAmount
          },
          isLoading: false
        });

        return;
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to claim rewards';
      set({ isLoading: false, error: message });
      throw error;
    }
  },

  // Claim all pending rewards
  claimAllRewards: async () => {
    return get().claimRewards();
  },

  clearError: () => set({ error: null })
}));

// ============================================
// Demo Data Generators
// ============================================

function createDemoPosition(address: string): StakingPosition {
  // Generate consistent demo data based on address
  const seed = address.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
  const stakedAmount = (seed % 50000) + 5000; // 5,000 - 55,000 HODL

  const tierConfig = calculateTier(stakedAmount);
  const { nextTier, progress } = calculateNextTierProgress(stakedAmount, tierConfig);

  const baseApr = 12; // 12% base APR
  const effectiveApr = baseApr * tierConfig.rewardMultiplier;

  const delegations: Delegation[] = [
    {
      validatorAddress: 'sharehodlvaloper1abc123',
      validatorName: 'ShareHODL Foundation',
      validatorTier: ValidatorTier.PLATINUM,
      amount: stakedAmount * 0.6,
      rewards: stakedAmount * 0.6 * (effectiveApr / 100) * (7 / 365), // ~1 week rewards
      commission: 0.03
    },
    {
      validatorAddress: 'sharehodlvaloper1def456',
      validatorName: 'Cosmos Validators',
      validatorTier: ValidatorTier.GOLD,
      amount: stakedAmount * 0.4,
      rewards: stakedAmount * 0.4 * (effectiveApr / 100) * (7 / 365),
      commission: 0.05
    }
  ];

  const pendingRewards = delegations.reduce((sum, d) => sum + d.rewards, 0);

  return {
    stakedAmount,
    pendingRewards,
    tier: tierConfig.tier,
    tierConfig,
    delegations,
    unbondings: [],
    apr: effectiveApr,
    nextTier,
    nextTierProgress: progress
  };
}

function createDemoValidators(): Validator[] {
  return [
    {
      address: 'sharehodlvaloper1abc123',
      name: 'ShareHODL Foundation',
      description: 'Official ShareHODL Foundation validator. Supporting network security and decentralization.',
      website: 'https://sharehodl.network',
      commission: 0.03,
      tier: ValidatorTier.PLATINUM,
      totalStaked: 15_000_000,
      delegatorCount: 2450,
      uptime: 99.98,
      isJailed: false,
      votingPower: 12.5
    },
    {
      address: 'sharehodlvaloper1def456',
      name: 'Cosmos Validators',
      description: 'Professional validator service with 99.9% uptime guarantee.',
      website: 'https://cosmosvalidators.io',
      commission: 0.05,
      tier: ValidatorTier.GOLD,
      totalStaked: 8_500_000,
      delegatorCount: 1820,
      uptime: 99.95,
      isJailed: false,
      votingPower: 8.2
    },
    {
      address: 'sharehodlvaloper1ghi789',
      name: 'Stake Capital',
      description: 'Enterprise-grade staking infrastructure.',
      website: 'https://stakecapital.com',
      commission: 0.05,
      tier: ValidatorTier.GOLD,
      totalStaked: 6_200_000,
      delegatorCount: 1340,
      uptime: 99.92,
      isJailed: false,
      votingPower: 6.1
    },
    {
      address: 'sharehodlvaloper1jkl012',
      name: 'Figment',
      description: 'Leading Web3 infrastructure provider.',
      website: 'https://figment.io',
      commission: 0.08,
      tier: ValidatorTier.SILVER,
      totalStaked: 4_100_000,
      delegatorCount: 980,
      uptime: 99.88,
      isJailed: false,
      votingPower: 4.0
    },
    {
      address: 'sharehodlvaloper1mno345',
      name: 'Chorus One',
      description: 'Institutional staking provider.',
      website: 'https://chorus.one',
      commission: 0.10,
      tier: ValidatorTier.SILVER,
      totalStaked: 3_500_000,
      delegatorCount: 720,
      uptime: 99.85,
      isJailed: false,
      votingPower: 3.4
    },
    {
      address: 'sharehodlvaloper1pqr678',
      name: 'Everstake',
      description: 'Multi-chain staking platform.',
      website: 'https://everstake.one',
      commission: 0.05,
      tier: ValidatorTier.BRONZE,
      totalStaked: 2_200_000,
      delegatorCount: 540,
      uptime: 99.80,
      isJailed: false,
      votingPower: 2.1
    },
    {
      address: 'sharehodlvaloper1stu901',
      name: 'P2P Validator',
      description: 'Non-custodial staking provider.',
      website: 'https://p2p.org',
      commission: 0.07,
      tier: ValidatorTier.BRONZE,
      totalStaked: 1_800_000,
      delegatorCount: 380,
      uptime: 99.75,
      isJailed: false,
      votingPower: 1.8
    },
    {
      address: 'sharehodlvaloper1vwx234',
      name: 'Blockdaemon',
      description: 'Enterprise blockchain infrastructure.',
      website: 'https://blockdaemon.com',
      commission: 0.06,
      tier: ValidatorTier.BRONZE,
      totalStaked: 1_500_000,
      delegatorCount: 290,
      uptime: 99.70,
      isJailed: false,
      votingPower: 1.5
    }
  ];
}

// Export helper functions
export { calculateTier, calculateNextTierProgress };
