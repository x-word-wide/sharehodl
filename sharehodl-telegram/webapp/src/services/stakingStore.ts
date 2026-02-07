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

// API base URL for ShareHODL blockchain
export const API_BASE = import.meta.env.VITE_SHAREHODL_REST || 'https://api.sharehodl.com';

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

  // Fetch user's staking position from blockchain
  fetchStakingPosition: async (address: string) => {
    set({ isLoading: true, error: null });

    try {
      // Fetch delegations from blockchain
      const delegationsResponse = await fetch(`${API_BASE}/cosmos/staking/v1beta1/delegations/${address}`);
      const delegationsData = await delegationsResponse.json();

      // Fetch rewards from blockchain
      const rewardsResponse = await fetch(`${API_BASE}/cosmos/distribution/v1beta1/delegators/${address}/rewards`);
      const rewardsData = await rewardsResponse.json();

      // Fetch unbonding from blockchain
      const unbondingResponse = await fetch(`${API_BASE}/cosmos/staking/v1beta1/delegators/${address}/unbonding_delegations`);
      const unbondingData = await unbondingResponse.json();

      // Get validators list for tier info
      const validators = get().validators;

      // Parse delegations
      const delegations: Delegation[] = (delegationsData.delegation_responses || []).map((del: {
        delegation: { validator_address: string };
        balance: { amount: string };
      }) => {
        const validatorAddress = del.delegation.validator_address;
        const validator = validators.find(v => v.address === validatorAddress);
        const amount = parseInt(del.balance.amount) / 1_000_000; // uhodl to HODL

        // Find rewards for this validator
        const validatorRewards = (rewardsData.rewards || []).find((r: { validator_address: string }) =>
          r.validator_address === validatorAddress
        );
        const rewards = validatorRewards?.reward?.[0]?.amount
          ? parseFloat(validatorRewards.reward[0].amount) / 1_000_000
          : 0;

        return {
          validatorAddress,
          validatorName: validator?.name || validatorAddress.slice(0, 12) + '...',
          validatorTier: validator?.tier || ValidatorTier.BRONZE,
          amount,
          rewards,
          commission: validator?.commission || 0.05
        };
      });

      // Parse unbondings
      const unbondings: Unbonding[] = (unbondingData.unbonding_responses || []).flatMap((unbond: {
        validator_address: string;
        entries: Array<{ balance: string; completion_time: string }>;
      }) =>
        unbond.entries.map(entry => ({
          validatorAddress: unbond.validator_address,
          amount: parseInt(entry.balance) / 1_000_000,
          completionTime: new Date(entry.completion_time).getTime()
        }))
      );

      // Calculate total staked and rewards
      const stakedAmount = delegations.reduce((sum, d) => sum + d.amount, 0);
      const pendingRewards = delegations.reduce((sum, d) => sum + d.rewards, 0);

      // Calculate tier
      const tierConfig = calculateTier(stakedAmount);
      const { nextTier, progress } = calculateNextTierProgress(stakedAmount, tierConfig);

      const baseApr = 12; // 12% base APR
      const effectiveApr = baseApr * tierConfig.rewardMultiplier;

      const position: StakingPosition = {
        stakedAmount,
        pendingRewards,
        tier: tierConfig.tier,
        tierConfig,
        delegations,
        unbondings,
        apr: effectiveApr,
        nextTier,
        nextTierProgress: progress
      };

      set({ position, isLoading: false });
    } catch (error) {
      console.error('Failed to fetch staking position:', error);
      // If API fails, show empty position
      const tierConfig = calculateTier(0);
      set({
        position: {
          stakedAmount: 0,
          pendingRewards: 0,
          tier: tierConfig.tier,
          tierConfig,
          delegations: [],
          unbondings: [],
          apr: 12,
          nextTier: STAKING_TIERS[1],
          nextTierProgress: 0
        },
        isLoading: false
      });
    }
  },

  // Fetch list of validators from blockchain
  fetchValidators: async () => {
    set({ isLoading: true, error: null });

    try {
      // Fetch bonded validators
      const response = await fetch(`${API_BASE}/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED`);
      const data = await response.json();

      const validators: Validator[] = (data.validators || []).map((val: {
        operator_address: string;
        description: { moniker: string; details: string; website: string };
        commission: { commission_rates: { rate: string } };
        tokens: string;
        delegator_shares: string;
        jailed: boolean;
      }, index: number) => {
        const totalStaked = parseInt(val.tokens) / 1_000_000; // uhodl to HODL
        const commission = parseFloat(val.commission.commission_rates.rate);

        // Determine tier based on total staked
        let tier = ValidatorTier.BRONZE;
        if (totalStaked >= 10_000_000) tier = ValidatorTier.DIAMOND;
        else if (totalStaked >= 5_000_000) tier = ValidatorTier.PLATINUM;
        else if (totalStaked >= 1_000_000) tier = ValidatorTier.GOLD;
        else if (totalStaked >= 100_000) tier = ValidatorTier.SILVER;

        return {
          address: val.operator_address,
          name: val.description.moniker || `Validator ${index + 1}`,
          description: val.description.details || 'No description provided',
          website: val.description.website || '',
          commission,
          tier,
          totalStaked,
          delegatorCount: 0, // Would need additional query
          uptime: 99.9, // Would need slashing info
          isJailed: val.jailed,
          votingPower: 0 // Calculated separately
        };
      });

      // Sort by total staked (most staked first)
      validators.sort((a, b) => b.totalStaked - a.totalStaked);

      // Calculate voting power as percentage
      const totalStakedAll = validators.reduce((sum, v) => sum + v.totalStaked, 0);
      validators.forEach(v => {
        v.votingPower = totalStakedAll > 0 ? (v.totalStaked / totalStakedAll) * 100 : 0;
      });

      set({ validators, isLoading: false });
    } catch (error) {
      console.error('Failed to fetch validators:', error);
      // If API fails, show empty list
      set({ validators: [], isLoading: false });
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
// Export helper functions
export { calculateTier, calculateNextTierProgress };
