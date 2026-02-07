/**
 * Governance Screen - Voting and Delegation combined
 */

import { useState } from 'react';

// Demo proposals
const PROPOSALS = [
  {
    id: 1,
    title: 'Increase Staking Rewards to 5%',
    description: 'Proposal to increase base staking APY from 4% to 5% to attract more stakers.',
    status: 'active',
    votesFor: 1250000,
    votesAgainst: 450000,
    votesAbstain: 120000,
    endDate: '2026-02-12',
    proposer: 'hodl1abc...xyz'
  },
  {
    id: 2,
    title: 'Add USDC to Lending Markets',
    description: 'Enable USDC as collateral and borrowing asset in the lending protocol.',
    status: 'active',
    votesFor: 890000,
    votesAgainst: 120000,
    votesAbstain: 50000,
    endDate: '2026-02-15',
    proposer: 'hodl1def...uvw'
  },
  {
    id: 3,
    title: 'Reduce Transaction Fees',
    description: 'Lower network transaction fees from 0.1% to 0.05% to increase adoption.',
    status: 'passed',
    votesFor: 2100000,
    votesAgainst: 300000,
    votesAbstain: 80000,
    endDate: '2026-01-28',
    proposer: 'hodl1ghi...rst'
  }
];

// Demo validators for delegation
const VALIDATORS = [
  {
    id: 'val1',
    name: 'ShareHODL Genesis',
    address: 'hodl1val...abc',
    votingPower: 15.2,
    delegators: 1250,
    commission: 5,
    uptime: 99.9
  },
  {
    id: 'val2',
    name: 'Cosmos Guardian',
    address: 'hodl1val...def',
    votingPower: 12.8,
    delegators: 980,
    commission: 8,
    uptime: 99.5
  },
  {
    id: 'val3',
    name: 'StakeFlow',
    address: 'hodl1val...ghi',
    votingPower: 10.5,
    delegators: 720,
    commission: 10,
    uptime: 99.8
  }
];

// User's delegations
const USER_DELEGATIONS = [
  { validatorId: 'val1', validatorName: 'ShareHODL Genesis', amount: 500000 }
];

export function GovernanceScreen() {
  const tg = window.Telegram?.WebApp;
  const [activeTab, setActiveTab] = useState<'proposals' | 'delegation'>('proposals');
  const [filter, setFilter] = useState<'all' | 'active' | 'passed' | 'rejected'>('all');
  const [selectedProposal, setSelectedProposal] = useState<typeof PROPOSALS[0] | null>(null);
  const [selectedValidator, setSelectedValidator] = useState<typeof VALIDATORS[0] | null>(null);
  const [delegateAmount, setDelegateAmount] = useState('');

  const filteredProposals = filter === 'all'
    ? PROPOSALS
    : PROPOSALS.filter(p => p.status === filter);

  const userVotingPower = 2500000;
  const userDelegated = USER_DELEGATIONS.reduce((sum, d) => sum + d.amount, 0);
  const userAvailable = userVotingPower - userDelegated;

  const handleVote = (proposalId: number, vote: 'for' | 'against' | 'abstain') => {
    tg?.HapticFeedback?.notificationOccurred('success');
    const voteLabel = vote === 'for' ? 'Yes' : vote === 'against' ? 'No' : 'Abstain';
    tg?.showAlert(`Vote submitted: ${voteLabel} on Proposal #${proposalId}`);
    setSelectedProposal(null);
  };

  const handleDelegate = () => {
    if (!selectedValidator || !delegateAmount) return;
    tg?.HapticFeedback?.notificationOccurred('success');
    tg?.showAlert(`Delegated ${parseInt(delegateAmount).toLocaleString()} HODL to ${selectedValidator.name}`);
    setSelectedValidator(null);
    setDelegateAmount('');
  };

  const handleUndelegate = (validatorName: string) => {
    tg?.HapticFeedback?.notificationOccurred('success');
    tg?.showAlert(`Undelegation from ${validatorName} initiated. Unbonding period: 21 days`);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return '#3B82F6';
      case 'passed': return '#10b981';
      case 'rejected': return '#ef4444';
      default: return '#8b949e';
    }
  };

  return (
    <div className="governance-screen">
      {/* Header */}
      <div className="header">
        <h1>Governance</h1>
        <p>Vote on proposals & delegate voting power</p>
      </div>

      {/* Stats */}
      <div className="stats-row">
        <div className="stat-card">
          <span className="stat-value">{(userVotingPower / 1000000).toFixed(1)}M</span>
          <span className="stat-label">Voting Power</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{PROPOSALS.filter(p => p.status === 'active').length}</span>
          <span className="stat-label">Active Proposals</span>
        </div>
      </div>

      {/* Tabs */}
      <div className="tabs">
        <button
          className={`tab ${activeTab === 'proposals' ? 'active' : ''}`}
          onClick={() => { tg?.HapticFeedback?.selectionChanged(); setActiveTab('proposals'); }}
        >
          Proposals
        </button>
        <button
          className={`tab ${activeTab === 'delegation' ? 'active' : ''}`}
          onClick={() => { tg?.HapticFeedback?.selectionChanged(); setActiveTab('delegation'); }}
        >
          Delegation
        </button>
      </div>

      {activeTab === 'proposals' ? (
        <>
          {/* Proposal Filters */}
          <div className="filters">
            {(['all', 'active', 'passed', 'rejected'] as const).map((f) => (
              <button
                key={f}
                className={`filter-btn ${filter === f ? 'active' : ''}`}
                onClick={() => { tg?.HapticFeedback?.selectionChanged(); setFilter(f); }}
              >
                {f.charAt(0).toUpperCase() + f.slice(1)}
              </button>
            ))}
          </div>

          {/* Proposals List */}
          <div className="proposals-list">
            {filteredProposals.map((proposal) => {
              const totalVotes = proposal.votesFor + proposal.votesAgainst + proposal.votesAbstain;
              const forPercent = totalVotes > 0 ? (proposal.votesFor / totalVotes) * 100 : 0;

              return (
                <button
                  key={proposal.id}
                  className="proposal-card"
                  onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); setSelectedProposal(proposal); }}
                >
                  <div className="proposal-header">
                    <span className="proposal-id">#{proposal.id}</span>
                    <span className="proposal-status" style={{ background: `${getStatusColor(proposal.status)}20`, color: getStatusColor(proposal.status) }}>
                      {proposal.status}
                    </span>
                  </div>
                  <h3 className="proposal-title">{proposal.title}</h3>
                  <p className="proposal-desc">{proposal.description}</p>

                  <div className="vote-progress">
                    <div className="progress-bar">
                      <div className="progress-fill" style={{ width: `${forPercent}%` }} />
                    </div>
                    <div className="vote-stats">
                      <span className="for">Yes: {(proposal.votesFor / 1000000).toFixed(1)}M</span>
                      <span className="against">No: {(proposal.votesAgainst / 1000000).toFixed(1)}M</span>
                    </div>
                  </div>

                  <div className="proposal-footer">
                    <span className="end-date">
                      {proposal.status === 'active' ? `Ends ${proposal.endDate}` : `Ended ${proposal.endDate}`}
                    </span>
                  </div>
                </button>
              );
            })}
          </div>
        </>
      ) : (
        <>
          {/* Delegation Overview */}
          <div className="delegation-overview">
            <div className="delegation-stat">
              <span className="label">Delegated</span>
              <span className="value">{(userDelegated / 1000000).toFixed(2)}M HODL</span>
            </div>
            <div className="delegation-stat">
              <span className="label">Available</span>
              <span className="value">{(userAvailable / 1000000).toFixed(2)}M HODL</span>
            </div>
          </div>

          {/* Current Delegations */}
          {USER_DELEGATIONS.length > 0 && (
            <div className="section">
              <h3 className="section-title">Your Delegations</h3>
              <div className="delegations-list">
                {USER_DELEGATIONS.map((del) => (
                  <div key={del.validatorId} className="delegation-item">
                    <div className="del-info">
                      <span className="del-name">{del.validatorName}</span>
                      <span className="del-amount">{(del.amount / 1000000).toFixed(2)}M HODL</span>
                    </div>
                    <button
                      className="undelegate-btn"
                      onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); handleUndelegate(del.validatorName); }}
                    >
                      Undelegate
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Validators */}
          <div className="section">
            <h3 className="section-title">Validators</h3>
            <div className="validators-list">
              {VALIDATORS.map((validator) => (
                <button
                  key={validator.id}
                  className="validator-card"
                  onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); setSelectedValidator(validator); }}
                >
                  <div className="validator-icon">
                    {validator.name.slice(0, 2).toUpperCase()}
                  </div>
                  <div className="validator-info">
                    <span className="validator-name">{validator.name}</span>
                    <div className="validator-stats">
                      <span>{validator.votingPower}% power</span>
                      <span>{validator.commission}% fee</span>
                    </div>
                  </div>
                  <div className="validator-arrow">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M9 18l6-6-6-6" />
                    </svg>
                  </div>
                </button>
              ))}
            </div>
          </div>
        </>
      )}

      {/* Vote Modal */}
      {selectedProposal && (
        <div className="modal-overlay" onClick={() => setSelectedProposal(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Proposal #{selectedProposal.id}</h2>
              <button className="close-btn" onClick={() => setSelectedProposal(null)}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>

            <h3 className="modal-title">{selectedProposal.title}</h3>
            <p className="modal-desc">{selectedProposal.description}</p>

            <div className="voting-power-info">
              <span className="label">Your Voting Power</span>
              <span className="value">{userVotingPower.toLocaleString()} HODL</span>
            </div>

            {selectedProposal.status === 'active' ? (
              <div className="vote-buttons">
                <button className="vote-btn yes" onClick={() => handleVote(selectedProposal.id, 'for')}>
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                  Yes
                </button>
                <button className="vote-btn no" onClick={() => handleVote(selectedProposal.id, 'against')}>
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                  </svg>
                  No
                </button>
                <button className="vote-btn abstain" onClick={() => handleVote(selectedProposal.id, 'abstain')}>
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="12" cy="12" r="10" />
                    <line x1="8" y1="12" x2="16" y2="12" />
                  </svg>
                  Abstain
                </button>
              </div>
            ) : (
              <div className="voting-closed">
                <p>Voting has ended</p>
                <span className="result" style={{ color: getStatusColor(selectedProposal.status) }}>
                  {selectedProposal.status.toUpperCase()}
                </span>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Delegate Modal */}
      {selectedValidator && (
        <div className="modal-overlay" onClick={() => setSelectedValidator(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Delegate to Validator</h2>
              <button className="close-btn" onClick={() => setSelectedValidator(null)}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>

            <div className="validator-detail">
              <div className="validator-icon large">
                {selectedValidator.name.slice(0, 2).toUpperCase()}
              </div>
              <h3>{selectedValidator.name}</h3>
              <div className="validator-detail-stats">
                <div className="detail-stat">
                  <span className="label">Voting Power</span>
                  <span className="value">{selectedValidator.votingPower}%</span>
                </div>
                <div className="detail-stat">
                  <span className="label">Commission</span>
                  <span className="value">{selectedValidator.commission}%</span>
                </div>
                <div className="detail-stat">
                  <span className="label">Uptime</span>
                  <span className="value">{selectedValidator.uptime}%</span>
                </div>
              </div>
            </div>

            <div className="delegate-input-section">
              <label>Amount to Delegate</label>
              <div className="input-row">
                <input
                  type="number"
                  placeholder="0"
                  value={delegateAmount}
                  onChange={(e) => setDelegateAmount(e.target.value)}
                />
                <button className="max-btn" onClick={() => setDelegateAmount(userAvailable.toString())}>
                  MAX
                </button>
              </div>
              <span className="available">Available: {userAvailable.toLocaleString()} HODL</span>
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
        .governance-screen {
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

        .stats-row {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 12px;
          margin-bottom: 20px;
        }

        .stat-card {
          padding: 16px;
          background: rgba(22, 27, 34, 0.8);
          border: 1px solid rgba(48, 54, 61, 0.6);
          border-radius: 14px;
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .stat-value {
          font-size: 22px;
          font-weight: 700;
          color: white;
        }

        .stat-label {
          font-size: 12px;
          color: #8b949e;
        }

        .tabs {
          display: flex;
          gap: 4px;
          padding: 4px;
          background: rgba(48, 54, 61, 0.4);
          border-radius: 12px;
          margin-bottom: 16px;
        }

        .tab {
          flex: 1;
          padding: 12px;
          background: transparent;
          border: none;
          border-radius: 10px;
          font-size: 14px;
          font-weight: 600;
          color: #8b949e;
          cursor: pointer;
          transition: all 0.2s;
        }

        .tab.active {
          background: linear-gradient(135deg, #8B5CF6 0%, #7C3AED 100%);
          color: white;
        }

        .filters {
          display: flex;
          gap: 8px;
          margin-bottom: 16px;
          overflow-x: auto;
          padding-bottom: 4px;
        }

        .filter-btn {
          padding: 8px 14px;
          background: rgba(48, 54, 61, 0.5);
          border: none;
          border-radius: 20px;
          font-size: 13px;
          font-weight: 600;
          color: #8b949e;
          white-space: nowrap;
          cursor: pointer;
          transition: all 0.2s;
        }

        .filter-btn.active {
          background: rgba(139, 92, 246, 0.2);
          color: #8B5CF6;
        }

        .proposals-list {
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .proposal-card {
          padding: 16px;
          background: rgba(22, 27, 34, 0.8);
          border: 1px solid rgba(48, 54, 61, 0.6);
          border-radius: 16px;
          text-align: left;
          cursor: pointer;
          transition: all 0.2s;
        }

        .proposal-card:active {
          transform: scale(0.98);
        }

        .proposal-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 10px;
        }

        .proposal-id {
          font-size: 12px;
          font-weight: 600;
          color: #8b949e;
        }

        .proposal-status {
          padding: 4px 10px;
          border-radius: 12px;
          font-size: 11px;
          font-weight: 600;
          text-transform: uppercase;
        }

        .proposal-title {
          font-size: 16px;
          font-weight: 600;
          color: white;
          margin: 0 0 8px;
        }

        .proposal-desc {
          font-size: 13px;
          color: #8b949e;
          margin: 0 0 14px;
          line-height: 1.5;
        }

        .vote-progress {
          margin-bottom: 12px;
        }

        .progress-bar {
          height: 6px;
          background: rgba(239, 68, 68, 0.3);
          border-radius: 3px;
          overflow: hidden;
          margin-bottom: 8px;
        }

        .progress-fill {
          height: 100%;
          background: #10b981;
          border-radius: 3px;
        }

        .vote-stats {
          display: flex;
          justify-content: space-between;
          font-size: 12px;
        }

        .vote-stats .for { color: #10b981; }
        .vote-stats .against { color: #ef4444; }

        .proposal-footer {
          padding-top: 12px;
          border-top: 1px solid rgba(48, 54, 61, 0.5);
        }

        .end-date {
          font-size: 12px;
          color: #8b949e;
        }

        /* Delegation Styles */
        .delegation-overview {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 12px;
          margin-bottom: 20px;
        }

        .delegation-stat {
          padding: 16px;
          background: rgba(139, 92, 246, 0.1);
          border: 1px solid rgba(139, 92, 246, 0.3);
          border-radius: 14px;
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .delegation-stat .label {
          font-size: 12px;
          color: #8b949e;
        }

        .delegation-stat .value {
          font-size: 16px;
          font-weight: 700;
          color: white;
        }

        .section {
          margin-bottom: 20px;
        }

        .section-title {
          font-size: 15px;
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
          justify-content: space-between;
          align-items: center;
          padding: 14px;
          background: rgba(22, 27, 34, 0.8);
          border: 1px solid rgba(48, 54, 61, 0.6);
          border-radius: 12px;
        }

        .del-info {
          display: flex;
          flex-direction: column;
          gap: 2px;
        }

        .del-name {
          font-size: 14px;
          font-weight: 600;
          color: white;
        }

        .del-amount {
          font-size: 13px;
          color: #8b949e;
        }

        .undelegate-btn {
          padding: 8px 14px;
          background: rgba(239, 68, 68, 0.15);
          border: 1px solid rgba(239, 68, 68, 0.3);
          border-radius: 8px;
          font-size: 13px;
          font-weight: 600;
          color: #ef4444;
          cursor: pointer;
        }

        .validators-list {
          display: flex;
          flex-direction: column;
          gap: 10px;
        }

        .validator-card {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 14px;
          background: rgba(22, 27, 34, 0.8);
          border: 1px solid rgba(48, 54, 61, 0.6);
          border-radius: 14px;
          cursor: pointer;
          transition: all 0.2s;
        }

        .validator-card:active {
          transform: scale(0.98);
        }

        .validator-icon {
          width: 44px;
          height: 44px;
          background: linear-gradient(135deg, #8B5CF6 0%, #7C3AED 100%);
          border-radius: 12px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 14px;
          font-weight: 700;
          color: white;
        }

        .validator-icon.large {
          width: 64px;
          height: 64px;
          font-size: 20px;
          border-radius: 16px;
        }

        .validator-info {
          flex: 1;
        }

        .validator-name {
          display: block;
          font-size: 15px;
          font-weight: 600;
          color: white;
          margin-bottom: 4px;
        }

        .validator-stats {
          display: flex;
          gap: 12px;
          font-size: 12px;
          color: #8b949e;
        }

        .validator-arrow {
          width: 20px;
          height: 20px;
          color: #8b949e;
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
          margin-bottom: 16px;
        }

        .modal-header h2 {
          font-size: 16px;
          font-weight: 600;
          color: #8b949e;
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

        .modal-title {
          font-size: 20px;
          font-weight: 700;
          color: white;
          margin: 0 0 12px;
        }

        .modal-desc {
          font-size: 14px;
          color: #8b949e;
          line-height: 1.6;
          margin: 0 0 20px;
        }

        .voting-power-info {
          display: flex;
          justify-content: space-between;
          padding: 14px;
          background: rgba(48, 54, 61, 0.4);
          border-radius: 12px;
          margin-bottom: 20px;
        }

        .voting-power-info .label {
          font-size: 14px;
          color: #8b949e;
        }

        .voting-power-info .value {
          font-size: 14px;
          font-weight: 600;
          color: white;
        }

        .vote-buttons {
          display: grid;
          grid-template-columns: 1fr 1fr 1fr;
          gap: 10px;
        }

        .vote-btn {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 6px;
          padding: 14px 8px;
          border: none;
          border-radius: 14px;
          font-size: 14px;
          font-weight: 700;
          cursor: pointer;
          transition: all 0.2s;
        }

        .vote-btn:active {
          transform: scale(0.95);
        }

        .vote-btn svg {
          width: 22px;
          height: 22px;
        }

        .vote-btn.yes {
          background: linear-gradient(135deg, #10b981 0%, #059669 100%);
          color: white;
        }

        .vote-btn.no {
          background: rgba(239, 68, 68, 0.15);
          color: #ef4444;
          border: 1px solid rgba(239, 68, 68, 0.3);
        }

        .vote-btn.abstain {
          background: rgba(139, 148, 158, 0.15);
          color: #8b949e;
          border: 1px solid rgba(139, 148, 158, 0.3);
        }

        .voting-closed {
          text-align: center;
          padding: 20px;
          background: rgba(48, 54, 61, 0.3);
          border-radius: 14px;
        }

        .voting-closed p {
          font-size: 14px;
          color: #8b949e;
          margin: 0 0 8px;
        }

        .voting-closed .result {
          font-size: 16px;
          font-weight: 700;
        }

        /* Delegate Modal */
        .validator-detail {
          text-align: center;
          margin-bottom: 20px;
        }

        .validator-detail .validator-icon {
          margin: 0 auto 12px;
        }

        .validator-detail h3 {
          font-size: 18px;
          font-weight: 700;
          color: white;
          margin: 0 0 16px;
        }

        .validator-detail-stats {
          display: flex;
          justify-content: center;
          gap: 20px;
        }

        .detail-stat {
          display: flex;
          flex-direction: column;
          gap: 2px;
        }

        .detail-stat .label {
          font-size: 11px;
          color: #8b949e;
        }

        .detail-stat .value {
          font-size: 14px;
          font-weight: 600;
          color: white;
        }

        .delegate-input-section {
          margin-bottom: 20px;
        }

        .delegate-input-section label {
          display: block;
          font-size: 13px;
          color: #8b949e;
          margin-bottom: 8px;
        }

        .input-row {
          display: flex;
          gap: 10px;
        }

        .input-row input {
          flex: 1;
          padding: 14px;
          background: rgba(48, 54, 61, 0.5);
          border: 1px solid rgba(48, 54, 61, 0.8);
          border-radius: 12px;
          font-size: 16px;
          color: white;
          outline: none;
        }

        .input-row input:focus {
          border-color: #8B5CF6;
        }

        .max-btn {
          padding: 14px 18px;
          background: rgba(139, 92, 246, 0.2);
          border: 1px solid rgba(139, 92, 246, 0.4);
          border-radius: 12px;
          font-size: 13px;
          font-weight: 700;
          color: #8B5CF6;
          cursor: pointer;
        }

        .available {
          display: block;
          font-size: 12px;
          color: #8b949e;
          margin-top: 8px;
        }

        .delegate-submit-btn {
          width: 100%;
          padding: 16px;
          background: linear-gradient(135deg, #8B5CF6 0%, #7C3AED 100%);
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
