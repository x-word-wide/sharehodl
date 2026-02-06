/**
 * Voting Screen - Governance proposals and voting
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
    endDate: '2026-02-12',
    proposer: 'sharehodl1abc...xyz'
  },
  {
    id: 2,
    title: 'Add USDC to Lending Markets',
    description: 'Enable USDC as collateral and borrowing asset in the lending protocol.',
    status: 'active',
    votesFor: 890000,
    votesAgainst: 120000,
    endDate: '2026-02-15',
    proposer: 'sharehodl1def...uvw'
  },
  {
    id: 3,
    title: 'Reduce Transaction Fees',
    description: 'Lower network transaction fees from 0.1% to 0.05% to increase adoption.',
    status: 'passed',
    votesFor: 2100000,
    votesAgainst: 300000,
    endDate: '2026-01-28',
    proposer: 'sharehodl1ghi...rst'
  },
  {
    id: 4,
    title: 'Community Treasury Allocation',
    description: 'Allocate 5M HODL from treasury to ecosystem development grants.',
    status: 'rejected',
    votesFor: 400000,
    votesAgainst: 1800000,
    endDate: '2026-01-20',
    proposer: 'sharehodl1jkl...opq'
  }
];

export function VotingScreen() {
  const tg = window.Telegram?.WebApp;
  const [filter, setFilter] = useState<'all' | 'active' | 'passed' | 'rejected'>('all');
  const [selectedProposal, setSelectedProposal] = useState<typeof PROPOSALS[0] | null>(null);

  const filteredProposals = filter === 'all'
    ? PROPOSALS
    : PROPOSALS.filter(p => p.status === filter);

  const handleVote = (proposalId: number, vote: 'for' | 'against' | 'abstain') => {
    tg?.HapticFeedback?.notificationOccurred('success');
    const voteLabel = vote === 'for' ? 'Yes' : vote === 'against' ? 'No' : 'Abstain';
    tg?.showAlert(`Vote submitted: ${voteLabel} on Proposal #${proposalId}`);
    setSelectedProposal(null);
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
    <div className="voting-screen">
      {/* Header */}
      <div className="header">
        <h1>Governance</h1>
        <p>Vote on proposals to shape the protocol</p>
      </div>

      {/* Stats */}
      <div className="stats-row">
        <div className="stat-card">
          <span className="stat-value">2.5M</span>
          <span className="stat-label">Your Voting Power</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{PROPOSALS.filter(p => p.status === 'active').length}</span>
          <span className="stat-label">Active Proposals</span>
        </div>
      </div>

      {/* Filters */}
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
          const totalVotes = proposal.votesFor + proposal.votesAgainst;
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

      {/* Vote Modal */}
      {selectedProposal && (
        <div className="modal-overlay" onClick={() => setSelectedProposal(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Vote on Proposal #{selectedProposal.id}</h2>
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
              <span className="value">2,500,000 HODL</span>
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
                <p>Voting has ended for this proposal</p>
                <span className="result" style={{ color: getStatusColor(selectedProposal.status) }}>
                  Result: {selectedProposal.status.toUpperCase()}
                </span>
              </div>
            )}
          </div>
        </div>
      )}

      <style>{`
        .voting-screen {
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

        .filters {
          display: flex;
          gap: 8px;
          margin-bottom: 16px;
          overflow-x: auto;
          padding-bottom: 4px;
        }

        .filter-btn {
          padding: 8px 16px;
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
          background: rgba(59, 130, 246, 0.2);
          color: #3B82F6;
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

        .vote-stats .for {
          color: #10b981;
        }

        .vote-stats .against {
          color: #ef4444;
        }

        .proposal-footer {
          padding-top: 12px;
          border-top: 1px solid rgba(48, 54, 61, 0.5);
        }

        .end-date {
          font-size: 12px;
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
      `}</style>
    </div>
  );
}
