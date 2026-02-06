'use client';

import { useState, useCallback, useEffect } from 'react';
import { SHAREHODL_CHAIN_CONFIG } from './useWallet';

export interface Validator {
  operatorAddress: string;
  moniker: string;
  tokens: string;
  commission: string;
  votingPower: string;
  status: string;
}

export interface Delegation {
  validatorAddress: string;
  validatorMoniker: string;
  amount: string;
  rewards: string;
}

export interface StakingState {
  validators: Validator[];
  delegations: Delegation[];
  totalStaked: string;
  totalRewards: string;
  unbonding: string;
  loading: boolean;
  error: string | null;
}

// Get REST URL
const getRestUrl = () => {
  if (typeof window !== 'undefined' && (window as any).__SHAREHODL_CONFIG__?.restUrl) {
    return (window as any).__SHAREHODL_CONFIG__.restUrl;
  }
  return 'http://localhost:1317';
};

export function useStaking(address: string | null) {
  const [state, setState] = useState<StakingState>({
    validators: [],
    delegations: [],
    totalStaked: '0',
    totalRewards: '0',
    unbonding: '0',
    loading: false,
    error: null,
  });

  const [txStatus, setTxStatus] = useState<{
    loading: boolean;
    error: string | null;
    success: string | null;
  }>({ loading: false, error: null, success: null });

  // Fetch all validators
  const fetchValidators = useCallback(async (): Promise<Validator[]> => {
    try {
      const response = await fetch(
        `${getRestUrl()}/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED&pagination.limit=100`
      );
      const data = await response.json();

      if (!data.validators) return [];

      // Calculate total tokens for voting power
      const totalTokens = data.validators.reduce(
        (sum: number, v: any) => sum + parseFloat(v.tokens || '0'),
        0
      );

      return data.validators.map((v: any) => ({
        operatorAddress: v.operator_address,
        moniker: v.description?.moniker || 'Unknown',
        tokens: v.tokens,
        commission: (parseFloat(v.commission?.commission_rates?.rate || '0') * 100).toFixed(1),
        votingPower: ((parseFloat(v.tokens || '0') / totalTokens) * 100).toFixed(2),
        status: v.status,
      }));
    } catch (error) {
      console.error('Error fetching validators:', error);
      return [];
    }
  }, []);

  // Fetch user's delegations
  const fetchDelegations = useCallback(async (userAddress: string): Promise<Delegation[]> => {
    try {
      // Fetch delegations
      const delegationsRes = await fetch(
        `${getRestUrl()}/cosmos/staking/v1beta1/delegations/${userAddress}`
      );
      const delegationsData = await delegationsRes.json();

      // Fetch rewards
      const rewardsRes = await fetch(
        `${getRestUrl()}/cosmos/distribution/v1beta1/delegators/${userAddress}/rewards`
      );
      const rewardsData = await rewardsRes.json();

      if (!delegationsData.delegation_responses) return [];

      // Get validator info for each delegation
      const delegations = await Promise.all(
        delegationsData.delegation_responses.map(async (d: any) => {
          // Try to get validator moniker
          let moniker = d.delegation.validator_address;
          try {
            const valRes = await fetch(
              `${getRestUrl()}/cosmos/staking/v1beta1/validators/${d.delegation.validator_address}`
            );
            const valData = await valRes.json();
            moniker = valData.validator?.description?.moniker || d.delegation.validator_address;
          } catch {
            // Use address if can't fetch moniker
          }

          // Find rewards for this validator
          const validatorRewards = rewardsData.rewards?.find(
            (r: any) => r.validator_address === d.delegation.validator_address
          );
          const rewardAmount = validatorRewards?.reward?.[0]?.amount || '0';

          return {
            validatorAddress: d.delegation.validator_address,
            validatorMoniker: moniker,
            amount: d.balance?.amount || '0',
            rewards: rewardAmount,
          };
        })
      );

      return delegations;
    } catch (error) {
      console.error('Error fetching delegations:', error);
      return [];
    }
  }, []);

  // Fetch unbonding delegations
  const fetchUnbonding = useCallback(async (userAddress: string): Promise<string> => {
    try {
      const response = await fetch(
        `${getRestUrl()}/cosmos/staking/v1beta1/delegators/${userAddress}/unbonding_delegations`
      );
      const data = await response.json();

      if (!data.unbonding_responses) return '0';

      let total = 0;
      data.unbonding_responses.forEach((u: any) => {
        u.entries?.forEach((e: any) => {
          total += parseFloat(e.balance || '0');
        });
      });

      return total.toString();
    } catch (error) {
      console.error('Error fetching unbonding:', error);
      return '0';
    }
  }, []);

  // Refresh all staking data
  const refresh = useCallback(async () => {
    setState(prev => ({ ...prev, loading: true, error: null }));

    try {
      const validators = await fetchValidators();

      let delegations: Delegation[] = [];
      let unbonding = '0';

      if (address) {
        delegations = await fetchDelegations(address);
        unbonding = await fetchUnbonding(address);
      }

      // Calculate totals
      const totalStaked = delegations.reduce(
        (sum, d) => sum + parseFloat(d.amount),
        0
      ).toString();

      const totalRewards = delegations.reduce(
        (sum, d) => sum + parseFloat(d.rewards),
        0
      ).toString();

      setState({
        validators,
        delegations,
        totalStaked,
        totalRewards,
        unbonding,
        loading: false,
        error: null,
      });
    } catch (error) {
      setState(prev => ({
        ...prev,
        loading: false,
        error: error instanceof Error ? error.message : 'Failed to fetch staking data',
      }));
    }
  }, [address, fetchValidators, fetchDelegations, fetchUnbonding]);

  // Delegate tokens
  const delegate = useCallback(async (validatorAddress: string, amount: string) => {
    if (!address || !window.keplr) {
      setTxStatus({ loading: false, error: 'Wallet not connected', success: null });
      return false;
    }

    setTxStatus({ loading: true, error: null, success: null });

    try {
      const chainId = SHAREHODL_CHAIN_CONFIG.chainId;
      await window.keplr.enable(chainId);
      const offlineSigner = window.keplr.getOfflineSigner(chainId);
      const accounts = await offlineSigner.getAccounts();

      // Convert amount to uhodl (6 decimals)
      const amountInMicro = (parseFloat(amount) * 1_000_000).toFixed(0);

      const msg = {
        typeUrl: '/cosmos.staking.v1beta1.MsgDelegate',
        value: {
          delegatorAddress: accounts[0].address,
          validatorAddress: validatorAddress,
          amount: {
            denom: 'stake',
            amount: amountInMicro,
          },
        },
      };

      const fee = {
        amount: [{ denom: 'uhodl', amount: '5000' }],
        gas: '200000',
      };

      // Use signAndBroadcast via Keplr
      const { SigningStargateClient } = await import('@cosmjs/stargate');
      const client = await SigningStargateClient.connectWithSigner(
        SHAREHODL_CHAIN_CONFIG.rpc,
        offlineSigner
      );

      const result = await client.signAndBroadcast(
        accounts[0].address,
        [msg],
        fee,
        ''
      );

      if (result.code === 0) {
        setTxStatus({ loading: false, error: null, success: `Delegated ${amount} STAKE successfully!` });
        await refresh();
        return true;
      } else {
        setTxStatus({ loading: false, error: `Transaction failed: ${result.rawLog}`, success: null });
        return false;
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Delegation failed';
      setTxStatus({ loading: false, error: errorMsg, success: null });
      return false;
    }
  }, [address, refresh]);

  // Undelegate tokens
  const undelegate = useCallback(async (validatorAddress: string, amount: string) => {
    if (!address || !window.keplr) {
      setTxStatus({ loading: false, error: 'Wallet not connected', success: null });
      return false;
    }

    setTxStatus({ loading: true, error: null, success: null });

    try {
      const chainId = SHAREHODL_CHAIN_CONFIG.chainId;
      await window.keplr.enable(chainId);
      const offlineSigner = window.keplr.getOfflineSigner(chainId);
      const accounts = await offlineSigner.getAccounts();

      const amountInMicro = (parseFloat(amount) * 1_000_000).toFixed(0);

      const msg = {
        typeUrl: '/cosmos.staking.v1beta1.MsgUndelegate',
        value: {
          delegatorAddress: accounts[0].address,
          validatorAddress: validatorAddress,
          amount: {
            denom: 'stake',
            amount: amountInMicro,
          },
        },
      };

      const fee = {
        amount: [{ denom: 'uhodl', amount: '5000' }],
        gas: '200000',
      };

      const { SigningStargateClient } = await import('@cosmjs/stargate');
      const client = await SigningStargateClient.connectWithSigner(
        SHAREHODL_CHAIN_CONFIG.rpc,
        offlineSigner
      );

      const result = await client.signAndBroadcast(
        accounts[0].address,
        [msg],
        fee,
        ''
      );

      if (result.code === 0) {
        setTxStatus({ loading: false, error: null, success: `Undelegation initiated! Tokens will be available in 21 days.` });
        await refresh();
        return true;
      } else {
        setTxStatus({ loading: false, error: `Transaction failed: ${result.rawLog}`, success: null });
        return false;
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Undelegation failed';
      setTxStatus({ loading: false, error: errorMsg, success: null });
      return false;
    }
  }, [address, refresh]);

  // Claim rewards
  const claimRewards = useCallback(async (validatorAddress?: string) => {
    if (!address || !window.keplr) {
      setTxStatus({ loading: false, error: 'Wallet not connected', success: null });
      return false;
    }

    setTxStatus({ loading: true, error: null, success: null });

    try {
      const chainId = SHAREHODL_CHAIN_CONFIG.chainId;
      await window.keplr.enable(chainId);
      const offlineSigner = window.keplr.getOfflineSigner(chainId);
      const accounts = await offlineSigner.getAccounts();

      // If no specific validator, claim from all
      const validatorsToClaim = validatorAddress
        ? [validatorAddress]
        : state.delegations.map(d => d.validatorAddress);

      const msgs = validatorsToClaim.map(valAddr => ({
        typeUrl: '/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward',
        value: {
          delegatorAddress: accounts[0].address,
          validatorAddress: valAddr,
        },
      }));

      const fee = {
        amount: [{ denom: 'uhodl', amount: '10000' }],
        gas: (100000 + 50000 * msgs.length).toString(),
      };

      const { SigningStargateClient } = await import('@cosmjs/stargate');
      const client = await SigningStargateClient.connectWithSigner(
        SHAREHODL_CHAIN_CONFIG.rpc,
        offlineSigner
      );

      const result = await client.signAndBroadcast(
        accounts[0].address,
        msgs,
        fee,
        ''
      );

      if (result.code === 0) {
        setTxStatus({ loading: false, error: null, success: 'Rewards claimed successfully!' });
        await refresh();
        return true;
      } else {
        setTxStatus({ loading: false, error: `Transaction failed: ${result.rawLog}`, success: null });
        return false;
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Claim failed';
      setTxStatus({ loading: false, error: errorMsg, success: null });
      return false;
    }
  }, [address, state.delegations, refresh]);

  // Clear transaction status
  const clearTxStatus = useCallback(() => {
    setTxStatus({ loading: false, error: null, success: null });
  }, []);

  // Auto-refresh on mount and when address changes
  useEffect(() => {
    refresh();
  }, [refresh]);

  return {
    ...state,
    txStatus,
    refresh,
    delegate,
    undelegate,
    claimRewards,
    clearTxStatus,
  };
}
