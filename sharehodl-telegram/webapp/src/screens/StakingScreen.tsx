/**
 * Staking Screen - Professional staking interface
 * Delegate, undelegate, and claim rewards
 */

import { useEffect, useState } from 'react';
import { useStakingStore } from '../services/stakingStore';
import { useWalletStore } from '../services/walletStore';
import { Chain, VALIDATOR_TIER_COLORS, type Validator } from '../types';

type Tab = 'stake' | 'validators' | 'rewards';

export function StakingScreen() {
  const tg = window.Telegram?.WebApp;
  const { accounts } = useWalletStore();
  const {
    position,
    validators,
    isLoading,
    fetchStakingPosition,
    fetchValidators,
    delegate,
    undelegate,
    claimAllRewards
  } = useStakingStore();

  const [activeTab, setActiveTab] = useState<Tab>('stake');
  const [showDelegateModal, setShowDelegateModal] = useState(false);
  const [showUndelegateModal, setShowUndelegateModal] = useState(false);
  const [selectedValidator, setSelectedValidator] = useState<Validator | null>(null);
  const [amount, setAmount] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);

  // Get ShareHODL address
  const sharehodlAccount = accounts.find(a => a.chain === Chain.SHAREHODL);
  const address = sharehodlAccount?.address || '';

  useEffect(() => {
    if (address) {
      fetchStakingPosition(address);
      fetchValidators();
    }
  }, [address, fetchStakingPosition, fetchValidators]);

  const handleDelegate = async () => {
    if (!selectedValidator || !amount) return;

    setIsProcessing(true);
    try {
      await delegate(selectedValidator.address, parseFloat(amount));
      tg?.HapticFeedback?.notificationOccurred('success');
      setShowDelegateModal(false);
      setAmount('');
      setSelectedValidator(null);
    } catch {
      tg?.HapticFeedback?.notificationOccurred('error');
      tg?.showAlert('Failed to delegate. Please try again.');
    } finally {
      setIsProcessing(false);
    }
  };

  const handleUndelegate = async () => {
    if (!selectedValidator || !amount) return;

    setIsProcessing(true);
    try {
      await undelegate(selectedValidator.address, parseFloat(amount));
      tg?.HapticFeedback?.notificationOccurred('success');
      setShowUndelegateModal(false);
      setAmount('');
      setSelectedValidator(null);
    } catch {
      tg?.HapticFeedback?.notificationOccurred('error');
      tg?.showAlert('Failed to undelegate. Please try again.');
    } finally {
      setIsProcessing(false);
    }
  };

  const handleClaimRewards = async () => {
    if (!position?.pendingRewards) return;

    setIsProcessing(true);
    try {
      await claimAllRewards();
      tg?.HapticFeedback?.notificationOccurred('success');
      tg?.showAlert(`Claimed ${position.pendingRewards.toFixed(2)} HODL rewards!`);
    } catch {
      tg?.HapticFeedback?.notificationOccurred('error');
      tg?.showAlert('Failed to claim rewards. Please try again.');
    } finally {
      setIsProcessing(false);
    }
  };

  const formatNumber = (num: number): string => {
    if (num >= 1_000_000) return `${(num / 1_000_000).toFixed(2)}M`;
    if (num >= 1_000) return `${(num / 1_000).toFixed(1)}K`;
    return num.toFixed(2);
  };

  if (isLoading && !position) {
    return (
      <div className="loading-screen">
        <div className="loader" />
        <p>Loading staking data...</p>
        <style>{loadingStyles}</style>
      </div>
    );
  }

  return (
    <div className="staking-screen">
      {/* Header with tier */}
      {position && (
        <div className="staking-header">
          <div className="tier-badge" style={{ background: `${position.tierConfig.color}20`, borderColor: position.tierConfig.color }}>
            <span className="tier-icon">{position.tierConfig.icon}</span>
            <span className="tier-name" style={{ color: position.tierConfig.color }}>{position.tierConfig.name}</span>
          </div>
          <div className="staking-overview">
            <div className="staked-amount">
              <span className="label">Total Staked</span>
              <span className="value">{formatNumber(position.stakedAmount)} HODL</span>
            </div>
            <div className="apr-badge">
              <span className="apr-value">{position.apr.toFixed(1)}%</span>
              <span className="apr-label">APR</span>
            </div>
          </div>

          {/* Tier Progress */}
          {position.nextTier && (
            <div className="tier-progress">
              <div className="progress-header">
                <span className="current-tier">{position.tierConfig.name}</span>
                <span className="next-tier">{position.nextTier.name}</span>
              </div>
              <div className="progress-bar">
                <div
                  className="progress-fill"
                  style={{
                    width: `${position.nextTierProgress}%`,
                    background: `linear-gradient(90deg, ${position.tierConfig.color}, ${position.nextTier.color})`
                  }}
                />
              </div>
              <p className="progress-text">
                Stake {formatNumber(position.nextTier.minStake - position.stakedAmount)} more HODL for {position.nextTier.rewardMultiplier}x rewards
              </p>
            </div>
          )}
        </div>
      )}

      {/* Tabs */}
      <div className="tabs">
        {(['stake', 'validators', 'rewards'] as Tab[]).map(tab => (
          <button
            key={tab}
            className={`tab ${activeTab === tab ? 'active' : ''}`}
            onClick={() => { setActiveTab(tab); tg?.HapticFeedback?.selectionChanged(); }}
          >
            {tab === 'stake' ? 'My Stake' : tab === 'validators' ? 'Validators' : 'Rewards'}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="tab-content">
        {activeTab === 'stake' && position && (
          <div className="stake-tab">
            {/* Quick Actions */}
            <div className="quick-actions">
              <button
                className="action-btn primary"
                onClick={() => setShowDelegateModal(true)}
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M12 5v14M5 12h14" />
                </svg>
                Stake
              </button>
              <button
                className="action-btn secondary"
                onClick={handleClaimRewards}
                disabled={!position.pendingRewards || isProcessing}
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M12 2l3 7h7l-5.5 4 2 7L12 16l-6.5 4 2-7L2 9h7z" />
                </svg>
                Claim {position.pendingRewards.toFixed(2)} HODL
              </button>
            </div>

            {/* Delegations List */}
            <div className="section">
              <h3 className="section-title">Your Delegations</h3>
              {position.delegations.length === 0 ? (
                <div className="empty-state">
                  <p>No active delegations</p>
                  <p className="hint">Stake HODL to earn rewards</p>
                </div>
              ) : (
                <div className="delegations-list">
                  {position.delegations.map((del) => (
                    <div key={del.validatorAddress} className="delegation-card">
                      <div className="delegation-header">
                        <div className="validator-info">
                          <span
                            className="tier-dot"
                            style={{ background: VALIDATOR_TIER_COLORS[del.validatorTier] }}
                          />
                          <span className="validator-name">{del.validatorName}</span>
                        </div>
                        <span className="commission">{(del.commission * 100).toFixed(0)}% fee</span>
                      </div>
                      <div className="delegation-details">
                        <div className="detail">
                          <span className="label">Staked</span>
                          <span className="value">{formatNumber(del.amount)} HODL</span>
                        </div>
                        <div className="detail">
                          <span className="label">Rewards</span>
                          <span className="value reward">{del.rewards.toFixed(4)} HODL</span>
                        </div>
                      </div>
                      <div className="delegation-actions">
                        <button
                          className="action-sm"
                          onClick={() => {
                            const v = validators.find(v => v.address === del.validatorAddress);
                            if (v) {
                              setSelectedValidator(v);
                              setShowDelegateModal(true);
                            }
                          }}
                        >
                          + Stake
                        </button>
                        <button
                          className="action-sm danger"
                          onClick={() => {
                            const v = validators.find(v => v.address === del.validatorAddress);
                            if (v) {
                              setSelectedValidator(v);
                              setShowUndelegateModal(true);
                            }
                          }}
                        >
                          Unstake
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Unbonding */}
            {position.unbondings.length > 0 && (
              <div className="section">
                <h3 className="section-title">Unbonding</h3>
                <div className="unbonding-list">
                  {position.unbondings.map((unbond, i) => (
                    <div key={i} className="unbonding-item">
                      <span className="amount">{formatNumber(unbond.amount)} HODL</span>
                      <span className="completion">
                        Ready {new Date(unbond.completionTime).toLocaleDateString()}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {activeTab === 'validators' && (
          <div className="validators-tab">
            <div className="validators-list">
              {validators.map((validator) => (
                <div
                  key={validator.address}
                  className="validator-card"
                  onClick={() => {
                    setSelectedValidator(validator);
                    setShowDelegateModal(true);
                    tg?.HapticFeedback?.impactOccurred('light');
                  }}
                >
                  <div className="validator-header">
                    <div className="validator-main">
                      <span
                        className="tier-badge-sm"
                        style={{ background: VALIDATOR_TIER_COLORS[validator.tier] }}
                      >
                        {validator.tier}
                      </span>
                      <span className="validator-name">{validator.name}</span>
                    </div>
                    <span className="commission">{(validator.commission * 100).toFixed(0)}%</span>
                  </div>
                  <div className="validator-stats">
                    <div className="stat">
                      <span className="label">Staked</span>
                      <span className="value">{formatNumber(validator.totalStaked)}</span>
                    </div>
                    <div className="stat">
                      <span className="label">Uptime</span>
                      <span className="value">{validator.uptime.toFixed(1)}%</span>
                    </div>
                    <div className="stat">
                      <span className="label">Delegators</span>
                      <span className="value">{validator.delegatorCount.toLocaleString()}</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'rewards' && position && (
          <div className="rewards-tab">
            {/* Tier Benefits */}
            <div className="tier-benefits">
              <h3 className="section-title">Your {position.tierConfig.name} Benefits</h3>
              <div className="benefits-list">
                {position.tierConfig.benefits.map((benefit, i) => (
                  <div key={i} className="benefit-item">
                    <span className="check-icon" style={{ color: position.tierConfig.color }}>&#10003;</span>
                    <span>{benefit}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Reward Multiplier */}
            <div className="multiplier-card">
              <span className="multiplier-label">Your Reward Multiplier</span>
              <span className="multiplier-value" style={{ color: position.tierConfig.color }}>
                {position.tierConfig.rewardMultiplier}x
              </span>
              <span className="multiplier-desc">
                Base APR: 12% &rarr; Your APR: {position.apr.toFixed(1)}%
              </span>
            </div>

            {/* Rewards Summary */}
            <div className="rewards-summary">
              <div className="summary-item">
                <span className="label">Pending Rewards</span>
                <span className="value">{position.pendingRewards.toFixed(4)} HODL</span>
              </div>
              <button
                className="claim-btn"
                onClick={handleClaimRewards}
                disabled={!position.pendingRewards || isProcessing}
              >
                {isProcessing ? 'Claiming...' : 'Claim All Rewards'}
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Delegate Modal */}
      {showDelegateModal && (
        <div className="modal-overlay" onClick={() => setShowDelegateModal(false)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h2>Stake HODL</h2>
            {selectedValidator && (
              <div className="selected-validator">
                <span className="tier-dot" style={{ background: VALIDATOR_TIER_COLORS[selectedValidator.tier] }} />
                <span>{selectedValidator.name}</span>
                <span className="commission">{(selectedValidator.commission * 100).toFixed(0)}% commission</span>
              </div>
            )}
            <div className="input-group">
              <input
                type="number"
                placeholder="Amount to stake"
                value={amount}
                onChange={e => setAmount(e.target.value)}
              />
              <span className="suffix">HODL</span>
            </div>
            <div className="modal-actions">
              <button className="cancel-btn" onClick={() => setShowDelegateModal(false)}>Cancel</button>
              <button
                className="confirm-btn"
                onClick={handleDelegate}
                disabled={!amount || parseFloat(amount) <= 0 || isProcessing}
              >
                {isProcessing ? 'Staking...' : 'Stake'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Undelegate Modal */}
      {showUndelegateModal && (
        <div className="modal-overlay" onClick={() => setShowUndelegateModal(false)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <h2>Unstake HODL</h2>
            {selectedValidator && (
              <div className="selected-validator">
                <span className="tier-dot" style={{ background: VALIDATOR_TIER_COLORS[selectedValidator.tier] }} />
                <span>{selectedValidator.name}</span>
              </div>
            )}
            <div className="warning-card">
              <span>Unstaking takes 21 days. You won't earn rewards during this period.</span>
            </div>
            <div className="input-group">
              <input
                type="number"
                placeholder="Amount to unstake"
                value={amount}
                onChange={e => setAmount(e.target.value)}
              />
              <span className="suffix">HODL</span>
            </div>
            <div className="modal-actions">
              <button className="cancel-btn" onClick={() => setShowUndelegateModal(false)}>Cancel</button>
              <button
                className="confirm-btn danger"
                onClick={handleUndelegate}
                disabled={!amount || parseFloat(amount) <= 0 || isProcessing}
              >
                {isProcessing ? 'Unstaking...' : 'Unstake'}
              </button>
            </div>
          </div>
        </div>
      )}

      <style>{styles}</style>
    </div>
  );
}

const loadingStyles = `
  .loading-screen {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    color: #8b949e;
  }
  .loader {
    width: 40px;
    height: 40px;
    border: 3px solid #30363d;
    border-top-color: #1E40AF;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 16px;
  }
  @keyframes spin {
    to { transform: rotate(360deg); }
  }
`;

const styles = `
  .staking-screen {
    min-height: 100vh;
    padding-bottom: 100px;
  }

  .staking-header {
    padding: 20px 16px;
    background: linear-gradient(180deg, rgba(30, 64, 175, 0.1) 0%, transparent 100%);
  }

  .tier-badge {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    border-radius: 20px;
    border: 1px solid;
    margin-bottom: 16px;
  }

  .tier-icon {
    font-size: 16px;
  }

  .tier-name {
    font-size: 14px;
    font-weight: 600;
  }

  .staking-overview {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }

  .staked-amount .label {
    display: block;
    font-size: 13px;
    color: #8b949e;
  }

  .staked-amount .value {
    font-size: 28px;
    font-weight: 700;
    color: white;
  }

  .apr-badge {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 12px 16px;
    background: rgba(16, 185, 129, 0.1);
    border-radius: 12px;
  }

  .apr-value {
    font-size: 20px;
    font-weight: 700;
    color: #10b981;
  }

  .apr-label {
    font-size: 12px;
    color: #8b949e;
  }

  .tier-progress {
    background: #161B22;
    border-radius: 12px;
    padding: 16px;
  }

  .progress-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .current-tier, .next-tier {
    font-size: 12px;
    font-weight: 600;
  }

  .current-tier {
    color: #8b949e;
  }

  .next-tier {
    color: white;
  }

  .progress-bar {
    height: 6px;
    background: #30363d;
    border-radius: 3px;
    overflow: hidden;
    margin-bottom: 8px;
  }

  .progress-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.3s ease;
  }

  .progress-text {
    font-size: 12px;
    color: #8b949e;
    margin: 0;
  }

  .tabs {
    display: flex;
    padding: 0 16px;
    gap: 8px;
    background: #0D1117;
    border-bottom: 1px solid #30363d;
  }

  .tab {
    flex: 1;
    padding: 14px;
    border: none;
    background: transparent;
    color: #8b949e;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    transition: all 0.2s ease;
  }

  .tab.active {
    color: #1E40AF;
    border-bottom-color: #1E40AF;
  }

  .tab-content {
    padding: 16px;
  }

  .quick-actions {
    display: flex;
    gap: 12px;
    margin-bottom: 24px;
  }

  .action-btn {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 14px;
    border: none;
    border-radius: 12px;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .action-btn svg {
    width: 20px;
    height: 20px;
  }

  .action-btn.primary {
    background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
    color: white;
  }

  .action-btn.secondary {
    background: #161B22;
    color: white;
  }

  .action-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .section {
    margin-bottom: 24px;
  }

  .section-title {
    font-size: 16px;
    font-weight: 600;
    color: white;
    margin: 0 0 12px;
  }

  .empty-state {
    text-align: center;
    padding: 32px;
    color: #8b949e;
  }

  .empty-state .hint {
    font-size: 14px;
    margin-top: 4px;
  }

  .delegations-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .delegation-card {
    background: #161B22;
    border-radius: 14px;
    padding: 16px;
  }

  .delegation-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .validator-info {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .tier-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
  }

  .validator-name {
    font-size: 15px;
    font-weight: 600;
    color: white;
  }

  .commission {
    font-size: 13px;
    color: #8b949e;
  }

  .delegation-details {
    display: flex;
    gap: 24px;
    margin-bottom: 12px;
  }

  .detail {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .detail .label {
    font-size: 12px;
    color: #8b949e;
  }

  .detail .value {
    font-size: 15px;
    font-weight: 600;
    color: white;
  }

  .detail .value.reward {
    color: #10b981;
  }

  .delegation-actions {
    display: flex;
    gap: 8px;
  }

  .action-sm {
    padding: 8px 16px;
    border: none;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    background: #30363d;
    color: white;
    transition: all 0.2s ease;
  }

  .action-sm.danger {
    background: rgba(239, 68, 68, 0.1);
    color: #f87171;
  }

  .unbonding-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .unbonding-item {
    display: flex;
    justify-content: space-between;
    padding: 12px;
    background: #161B22;
    border-radius: 10px;
  }

  .unbonding-item .amount {
    font-weight: 600;
    color: white;
  }

  .unbonding-item .completion {
    color: #8b949e;
    font-size: 13px;
  }

  .validators-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .validator-card {
    background: #161B22;
    border-radius: 14px;
    padding: 16px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .validator-card:active {
    transform: scale(0.98);
  }

  .validator-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .validator-main {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .tier-badge-sm {
    padding: 3px 8px;
    border-radius: 6px;
    font-size: 10px;
    font-weight: 700;
    color: #0D1117;
    text-transform: uppercase;
  }

  .validator-stats {
    display: flex;
    gap: 16px;
  }

  .stat {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .stat .label {
    font-size: 11px;
    color: #8b949e;
  }

  .stat .value {
    font-size: 14px;
    font-weight: 600;
    color: white;
  }

  .tier-benefits {
    background: #161B22;
    border-radius: 14px;
    padding: 16px;
    margin-bottom: 16px;
  }

  .benefits-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .benefit-item {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 14px;
    color: white;
  }

  .check-icon {
    font-size: 16px;
    font-weight: 700;
  }

  .multiplier-card {
    background: linear-gradient(135deg, rgba(30, 64, 175, 0.1) 0%, rgba(59, 130, 246, 0.1) 100%);
    border-radius: 14px;
    padding: 20px;
    text-align: center;
    margin-bottom: 16px;
  }

  .multiplier-label {
    display: block;
    font-size: 13px;
    color: #8b949e;
    margin-bottom: 8px;
  }

  .multiplier-value {
    display: block;
    font-size: 48px;
    font-weight: 700;
    margin-bottom: 8px;
  }

  .multiplier-desc {
    font-size: 14px;
    color: #8b949e;
  }

  .rewards-summary {
    background: #161B22;
    border-radius: 14px;
    padding: 16px;
  }

  .summary-item {
    display: flex;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .summary-item .label {
    color: #8b949e;
  }

  .summary-item .value {
    font-weight: 600;
    color: #10b981;
  }

  .claim-btn {
    width: 100%;
    padding: 14px;
    border: none;
    border-radius: 12px;
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
    color: white;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
  }

  .claim-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Modal */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.8);
    display: flex;
    align-items: flex-end;
    z-index: 100;
  }

  .modal-content {
    width: 100%;
    background: #161B22;
    border-radius: 20px 20px 0 0;
    padding: 24px;
  }

  .modal-content h2 {
    font-size: 20px;
    font-weight: 700;
    color: white;
    margin: 0 0 16px;
    text-align: center;
  }

  .selected-validator {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px;
    background: #0D1117;
    border-radius: 10px;
    margin-bottom: 16px;
    color: white;
    font-weight: 500;
  }

  .selected-validator .commission {
    margin-left: auto;
  }

  .warning-card {
    padding: 12px;
    background: rgba(245, 158, 11, 0.1);
    border-radius: 10px;
    margin-bottom: 16px;
    color: #f59e0b;
    font-size: 13px;
  }

  .input-group {
    position: relative;
    margin-bottom: 16px;
  }

  .input-group input {
    width: 100%;
    padding: 16px;
    padding-right: 60px;
    background: #0D1117;
    border: 1px solid #30363d;
    border-radius: 12px;
    font-size: 18px;
    font-weight: 600;
    color: white;
    outline: none;
  }

  .input-group input:focus {
    border-color: #1E40AF;
  }

  .input-group .suffix {
    position: absolute;
    right: 16px;
    top: 50%;
    transform: translateY(-50%);
    color: #8b949e;
    font-weight: 600;
  }

  .modal-actions {
    display: flex;
    gap: 12px;
  }

  .cancel-btn, .confirm-btn {
    flex: 1;
    padding: 14px;
    border: none;
    border-radius: 12px;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
  }

  .cancel-btn {
    background: #30363d;
    color: white;
  }

  .confirm-btn {
    background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
    color: white;
  }

  .confirm-btn.danger {
    background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  }

  .confirm-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;
