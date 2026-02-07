/**
 * Delegation Screen - Delegate voting power to validators/representatives
 */

import { useState } from 'react';

// Demo validators/delegates
const DELEGATES = [
  {
    id: 'val1',
    name: 'ShareHODL Foundation',
    address: 'hodl1abc...xyz',
    votingPower: 15000000,
    delegators: 1250,
    commission: 5,
    isActive: true,
    avatar: 'SF'
  },
  {
    id: 'val2',
    name: 'Community Validator',
    address: 'hodl1def...uvw',
    votingPower: 8500000,
    delegators: 890,
    commission: 8,
    isActive: true,
    avatar: 'CV'
  },
  {
    id: 'val3',
    name: 'DeFi Alliance',
    address: 'hodl1ghi...rst',
    votingPower: 6200000,
    delegators: 650,
    commission: 10,
    isActive: true,
    avatar: 'DA'
  },
  {
    id: 'val4',
    name: 'Governance Guild',
    address: 'hodl1jkl...opq',
    votingPower: 4100000,
    delegators: 420,
    commission: 7,
    isActive: true,
    avatar: 'GG'
  },
  {
    id: 'val5',
    name: 'Staking Pro',
    address: 'hodl1mno...lmn',
    votingPower: 2800000,
    delegators: 310,
    commission: 12,
    isActive: false,
    avatar: 'SP'
  }
];

// Demo user delegations
const USER_DELEGATIONS = [
  { delegateId: 'val1', amount: 1500000 },
  { delegateId: 'val3', amount: 500000 }
];

export function DelegationScreen() {
  const tg = window.Telegram?.WebApp;
  const [selectedDelegate, setSelectedDelegate] = useState<typeof DELEGATES[0] | null>(null);
  const [delegateAmount, setDelegateAmount] = useState('');
  const [showDelegateModal, setShowDelegateModal] = useState(false);

  const totalDelegated = USER_DELEGATIONS.reduce((sum, d) => sum + d.amount, 0);
  const availableToDelegate = 2500000 - totalDelegated; // Assume 2.5M total

  const handleDelegate = () => {
    if (!selectedDelegate || !delegateAmount) return;
    tg?.HapticFeedback?.notificationOccurred('success');
    tg?.showAlert(`Delegated ${parseInt(delegateAmount).toLocaleString()} HODL to ${selectedDelegate.name}`);
    setShowDelegateModal(false);
    setDelegateAmount('');
    setSelectedDelegate(null);
  };

  const handleUndelegate = (delegateId: string) => {
    const delegate = DELEGATES.find(d => d.id === delegateId);
    tg?.HapticFeedback?.notificationOccurred('warning');
    tg?.showAlert(`Undelegation from ${delegate?.name} initiated. Tokens will be available in 21 days.`);
  };

  return (
    <div className="delegation-screen">
      {/* Header */}
      <div className="header">
        <h1>Delegation</h1>
        <p>Delegate your voting power to representatives</p>
      </div>

      {/* Overview */}
      <div className="overview-card">
        <div className="overview-row">
          <div className="overview-item">
            <span className="overview-value">{(totalDelegated / 1000000).toFixed(1)}M</span>
            <span className="overview-label">Delegated</span>
          </div>
          <div className="overview-item">
            <span className="overview-value">{(availableToDelegate / 1000000).toFixed(1)}M</span>
            <span className="overview-label">Available</span>
          </div>
          <div className="overview-item">
            <span className="overview-value">{USER_DELEGATIONS.length}</span>
            <span className="overview-label">Delegates</span>
          </div>
        </div>
      </div>

      {/* Your Delegations */}
      {USER_DELEGATIONS.length > 0 && (
        <div className="section">
          <h2 className="section-title">Your Delegations</h2>
          <div className="delegations-list">
            {USER_DELEGATIONS.map((del) => {
              const delegate = DELEGATES.find(d => d.id === del.delegateId);
              if (!delegate) return null;
              return (
                <div key={del.delegateId} className="delegation-item">
                  <div className="delegate-avatar" style={{ background: '#3B82F620', color: '#3B82F6' }}>
                    {delegate.avatar}
                  </div>
                  <div className="delegation-info">
                    <span className="delegate-name">{delegate.name}</span>
                    <span className="delegation-amount">{(del.amount / 1000000).toFixed(1)}M HODL</span>
                  </div>
                  <button className="undelegate-btn" onClick={() => handleUndelegate(del.delegateId)}>
                    Undelegate
                  </button>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Available Delegates */}
      <div className="section">
        <h2 className="section-title">Validators</h2>
        <div className="delegates-list">
          {DELEGATES.map((delegate) => {
            const userDelegation = USER_DELEGATIONS.find(d => d.delegateId === delegate.id);
            return (
              <button
                key={delegate.id}
                className={`delegate-card ${!delegate.isActive ? 'inactive' : ''}`}
                onClick={() => {
                  if (delegate.isActive) {
                    tg?.HapticFeedback?.impactOccurred('light');
                    setSelectedDelegate(delegate);
                    setShowDelegateModal(true);
                  }
                }}
              >
                <div className="delegate-avatar" style={{ background: delegate.isActive ? '#3B82F620' : '#8b949e20', color: delegate.isActive ? '#3B82F6' : '#8b949e' }}>
                  {delegate.avatar}
                </div>
                <div className="delegate-info">
                  <div className="delegate-header">
                    <span className="delegate-name">{delegate.name}</span>
                    {!delegate.isActive && <span className="inactive-badge">Inactive</span>}
                  </div>
                  <div className="delegate-stats">
                    <span>{(delegate.votingPower / 1000000).toFixed(1)}M VP</span>
                    <span className="dot">-</span>
                    <span>{delegate.delegators} delegators</span>
                    <span className="dot">-</span>
                    <span>{delegate.commission}% fee</span>
                  </div>
                </div>
                {userDelegation && (
                  <div className="delegated-badge">
                    <svg viewBox="0 0 24 24" fill="currentColor">
                      <path d="M9 12l2 2 4-4" />
                      <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="2" />
                    </svg>
                  </div>
                )}
              </button>
            );
          })}
        </div>
      </div>

      {/* Delegate Modal */}
      {showDelegateModal && selectedDelegate && (
        <div className="modal-overlay" onClick={() => setShowDelegateModal(false)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Delegate to {selectedDelegate.name}</h2>
              <button className="close-btn" onClick={() => setShowDelegateModal(false)}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>

            <div className="delegate-preview">
              <div className="delegate-avatar large" style={{ background: '#3B82F620', color: '#3B82F6' }}>
                {selectedDelegate.avatar}
              </div>
              <div className="delegate-details">
                <span className="name">{selectedDelegate.name}</span>
                <span className="address">{selectedDelegate.address}</span>
              </div>
            </div>

            <div className="delegate-meta">
              <div className="meta-item">
                <span className="label">Commission</span>
                <span className="value">{selectedDelegate.commission}%</span>
              </div>
              <div className="meta-item">
                <span className="label">Voting Power</span>
                <span className="value">{(selectedDelegate.votingPower / 1000000).toFixed(1)}M</span>
              </div>
              <div className="meta-item">
                <span className="label">Delegators</span>
                <span className="value">{selectedDelegate.delegators}</span>
              </div>
            </div>

            <div className="amount-input-group">
              <label>Amount to Delegate</label>
              <div className="input-row">
                <input
                  type="text"
                  inputMode="numeric"
                  value={delegateAmount}
                  onChange={(e) => setDelegateAmount(e.target.value.replace(/\D/g, ''))}
                  placeholder="0"
                  className="amount-input"
                />
                <span className="currency">HODL</span>
              </div>
              <div className="balance-row">
                <span>Available: {availableToDelegate.toLocaleString()} HODL</span>
                <button
                  className="max-btn"
                  onClick={() => setDelegateAmount(availableToDelegate.toString())}
                >
                  MAX
                </button>
              </div>
            </div>

            <button
              className="delegate-submit-btn"
              onClick={handleDelegate}
              disabled={!delegateAmount || parseInt(delegateAmount) <= 0}
            >
              Delegate
            </button>
          </div>
        </div>
      )}

      <style>{`
        .delegation-screen {
          min-height: 100vh;
          padding: 16px;
          padding-bottom: 100px;
        }

        .header {
          margin-bottom: 20px;
        }

        .header h1 {
          font-size: 24px;
          font-weight: 700;
          color: white;
          margin: 0 0 4px;
        }

        .header p {
          font-size: 14px;
          color: #8b949e;
          margin: 0;
        }

        .overview-card {
          padding: 20px;
          background: rgba(22, 27, 34, 0.8);
          border: 1px solid rgba(48, 54, 61, 0.6);
          border-radius: 16px;
          margin-bottom: 24px;
        }

        .overview-row {
          display: flex;
          justify-content: space-around;
        }

        .overview-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 4px;
        }

        .overview-value {
          font-size: 24px;
          font-weight: 700;
          color: white;
        }

        .overview-label {
          font-size: 12px;
          color: #8b949e;
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

        .delegations-list {
          display: flex;
          flex-direction: column;
          gap: 10px;
        }

        .delegation-item {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 14px;
          background: rgba(59, 130, 246, 0.1);
          border: 1px solid rgba(59, 130, 246, 0.3);
          border-radius: 14px;
        }

        .delegate-avatar {
          width: 44px;
          height: 44px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 14px;
          font-weight: 700;
          flex-shrink: 0;
        }

        .delegate-avatar.large {
          width: 56px;
          height: 56px;
          font-size: 18px;
        }

        .delegation-info {
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: 2px;
        }

        .delegate-name {
          font-size: 15px;
          font-weight: 600;
          color: white;
        }

        .delegation-amount {
          font-size: 13px;
          color: #8b949e;
        }

        .undelegate-btn {
          padding: 8px 14px;
          background: rgba(239, 68, 68, 0.15);
          border: none;
          border-radius: 10px;
          font-size: 13px;
          font-weight: 600;
          color: #ef4444;
          cursor: pointer;
        }

        .delegates-list {
          display: flex;
          flex-direction: column;
          gap: 10px;
        }

        .delegate-card {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 14px;
          background: rgba(22, 27, 34, 0.8);
          border: 1px solid rgba(48, 54, 61, 0.6);
          border-radius: 14px;
          text-align: left;
          cursor: pointer;
          transition: all 0.2s;
        }

        .delegate-card:active {
          transform: scale(0.98);
        }

        .delegate-card.inactive {
          opacity: 0.6;
          cursor: not-allowed;
        }

        .delegate-info {
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .delegate-header {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .inactive-badge {
          padding: 2px 8px;
          background: rgba(239, 68, 68, 0.15);
          border-radius: 8px;
          font-size: 10px;
          font-weight: 600;
          color: #ef4444;
        }

        .delegate-stats {
          display: flex;
          align-items: center;
          gap: 6px;
          font-size: 12px;
          color: #8b949e;
        }

        .dot {
          opacity: 0.5;
        }

        .delegated-badge {
          width: 28px;
          height: 28px;
          color: #10b981;
        }

        .delegated-badge svg {
          width: 100%;
          height: 100%;
        }

        /* Modal */
        .modal-overlay {
          position: fixed;
          inset: 0;
          background: rgba(0, 0, 0, 0.7);
          display: flex;
          align-items: flex-end;
          justify-content: center;
          z-index: 100;
        }

        .modal-content {
          width: 100%;
          max-width: 500px;
          background: #161B22;
          border-radius: 20px 20px 0 0;
          padding: 24px;
          padding-bottom: calc(24px + env(safe-area-inset-bottom, 0));
        }

        .modal-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 20px;
        }

        .modal-header h2 {
          font-size: 18px;
          font-weight: 700;
          color: white;
          margin: 0;
        }

        .close-btn {
          width: 32px;
          height: 32px;
          background: rgba(48, 54, 61, 0.5);
          border: none;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          cursor: pointer;
        }

        .close-btn svg {
          width: 18px;
          height: 18px;
          color: #8b949e;
        }

        .delegate-preview {
          display: flex;
          align-items: center;
          gap: 14px;
          padding: 16px;
          background: rgba(48, 54, 61, 0.3);
          border-radius: 14px;
          margin-bottom: 16px;
        }

        .delegate-details {
          display: flex;
          flex-direction: column;
          gap: 2px;
        }

        .delegate-details .name {
          font-size: 16px;
          font-weight: 600;
          color: white;
        }

        .delegate-details .address {
          font-size: 13px;
          color: #8b949e;
        }

        .delegate-meta {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 12px;
          margin-bottom: 20px;
        }

        .meta-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 4px;
          padding: 12px;
          background: rgba(48, 54, 61, 0.3);
          border-radius: 10px;
        }

        .meta-item .label {
          font-size: 11px;
          color: #8b949e;
        }

        .meta-item .value {
          font-size: 15px;
          font-weight: 600;
          color: white;
        }

        .amount-input-group {
          margin-bottom: 20px;
        }

        .amount-input-group label {
          display: block;
          font-size: 14px;
          color: #8b949e;
          margin-bottom: 8px;
        }

        .input-row {
          display: flex;
          align-items: center;
          background: rgba(48, 54, 61, 0.4);
          border-radius: 12px;
          padding: 4px;
        }

        .amount-input {
          flex: 1;
          background: transparent;
          border: none;
          padding: 14px;
          font-size: 20px;
          font-weight: 600;
          color: white;
          outline: none;
        }

        .amount-input::placeholder {
          color: #484f58;
        }

        .currency {
          padding: 10px 16px;
          background: rgba(48, 54, 61, 0.6);
          border-radius: 10px;
          font-size: 14px;
          font-weight: 600;
          color: #8b949e;
          margin-right: 4px;
        }

        .balance-row {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-top: 10px;
          font-size: 13px;
          color: #8b949e;
        }

        .max-btn {
          padding: 4px 10px;
          background: rgba(59, 130, 246, 0.2);
          border: none;
          border-radius: 6px;
          font-size: 12px;
          font-weight: 700;
          color: #3B82F6;
          cursor: pointer;
        }

        .delegate-submit-btn {
          width: 100%;
          padding: 16px;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          border: none;
          border-radius: 14px;
          font-size: 16px;
          font-weight: 700;
          color: white;
          cursor: pointer;
          transition: all 0.2s;
        }

        .delegate-submit-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .delegate-submit-btn:not(:disabled):active {
          transform: scale(0.98);
        }
      `}</style>
    </div>
  );
}
