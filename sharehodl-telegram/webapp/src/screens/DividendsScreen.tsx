/**
 * Dividends Screen - Shows pending and paid dividends
 * Governance-controlled dividend distribution with audit verification
 */

import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Loader2, Clock, CheckCircle, XCircle, FileText, Vote } from 'lucide-react';
import { useWalletStore } from '../services/walletStore';
import { Chain } from '../types';

const API_BASE = import.meta.env.VITE_SHAREHODL_REST || 'https://api.sharehodl.com';

interface PendingDividend {
  dividend_id: number;
  company_id: number;
  company_name: string;
  company_symbol: string;
  type: string;
  status: string;
  amount_per_share: string;
  currency: string;
  shares_held: string;
  estimated_amount: string;
  record_date: string;
  payment_date: string;
  proposal_id?: number;
  audit_hash?: string;
  description: string;
}

export function DividendsScreen() {
  const navigate = useNavigate();
  const { accounts } = useWalletStore();
  const [isLoading, setIsLoading] = useState(true);
  const [dividends, setDividends] = useState<PendingDividend[]>([]);
  const [filter, setFilter] = useState<'all' | 'pending' | 'approved' | 'paid'>('all');

  // Get ShareHODL address
  const sharehodlAccount = accounts.find(a => a.chain === Chain.SHAREHODL);
  const address = sharehodlAccount?.address || '';

  useEffect(() => {
    const fetchDividends = async () => {
      if (!address) {
        setIsLoading(false);
        return;
      }

      try {
        const response = await fetch(
          `${API_BASE}/sharehodl/equity/v1/dividends/pending?shareholder=${address}`
        );
        if (response.ok) {
          const data = await response.json();
          setDividends(data.dividends || []);
        }
      } catch (error) {
        console.error('Failed to fetch dividends:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchDividends();
  }, [address]);

  const filteredDividends = dividends.filter(d => {
    if (filter === 'all') return true;
    if (filter === 'pending') return d.status === 'pending_approval';
    if (filter === 'approved') return d.status === 'declared' || d.status === 'recorded';
    if (filter === 'paid') return d.status === 'paid';
    return true;
  });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'pending_approval':
        return { icon: Clock, color: '#f59e0b', label: 'Pending Vote' };
      case 'declared':
      case 'recorded':
        return { icon: CheckCircle, color: '#10b981', label: 'Approved' };
      case 'paid':
        return { icon: CheckCircle, color: '#3b82f6', label: 'Paid' };
      case 'rejected':
        return { icon: XCircle, color: '#ef4444', label: 'Rejected' };
      default:
        return { icon: Clock, color: '#8b949e', label: status };
    }
  };

  const formatAmount = (amount: string, currency: string) => {
    const num = parseFloat(amount) / 1_000_000; // Convert from micro
    return `${num.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} ${currency.toUpperCase()}`;
  };

  if (isLoading) {
    return (
      <div className="dividends-screen">
        <div className="header">
          <button className="back-btn" onClick={() => navigate(-1)}>
            <ArrowLeft size={24} />
          </button>
          <h1>Dividends</h1>
        </div>
        <div className="loading-state">
          <Loader2 className="spin" size={40} />
          <p>Loading dividends...</p>
        </div>
        <style>{styles}</style>
      </div>
    );
  }

  return (
    <div className="dividends-screen">
      <div className="header">
        <button className="back-btn" onClick={() => navigate(-1)}>
          <ArrowLeft size={24} />
        </button>
        <h1>Dividends</h1>
      </div>

      {/* Filter Tabs */}
      <div className="filter-tabs">
        {(['all', 'pending', 'approved', 'paid'] as const).map(tab => (
          <button
            key={tab}
            className={`filter-tab ${filter === tab ? 'active' : ''}`}
            onClick={() => setFilter(tab)}
          >
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </button>
        ))}
      </div>

      {/* Info Banner */}
      <div className="info-banner">
        <Vote size={18} />
        <p>Dividends require governance approval. Validators verify audit reports before distribution.</p>
      </div>

      {/* Dividends List */}
      {filteredDividends.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">
            <FileText size={48} />
          </div>
          <h3>No Dividends Found</h3>
          <p>
            {filter === 'pending'
              ? 'No pending dividend distributions for your holdings.'
              : filter === 'approved'
              ? 'No approved dividends awaiting payment.'
              : filter === 'paid'
              ? 'No dividend payments received yet.'
              : 'You have no dividend distributions for your equity holdings.'}
          </p>
        </div>
      ) : (
        <div className="dividends-list">
          {filteredDividends.map((dividend) => {
            const status = getStatusBadge(dividend.status);
            const StatusIcon = status.icon;

            return (
              <div key={dividend.dividend_id} className="dividend-card">
                <div className="dividend-header">
                  <div className="company-info">
                    <span className="company-symbol">{dividend.company_symbol}</span>
                    <span className="company-name">{dividend.company_name}</span>
                  </div>
                  <div className="status-badge" style={{ backgroundColor: `${status.color}20`, color: status.color }}>
                    <StatusIcon size={14} />
                    <span>{status.label}</span>
                  </div>
                </div>

                <div className="dividend-details">
                  <div className="detail-row">
                    <span className="label">Your Shares</span>
                    <span className="value">{parseInt(dividend.shares_held).toLocaleString()}</span>
                  </div>
                  <div className="detail-row">
                    <span className="label">Per Share</span>
                    <span className="value">{formatAmount(dividend.amount_per_share, dividend.currency)}</span>
                  </div>
                  <div className="detail-row highlight">
                    <span className="label">Estimated Payment</span>
                    <span className="value amount">{formatAmount(dividend.estimated_amount, dividend.currency)}</span>
                  </div>
                </div>

                <div className="dividend-dates">
                  <div className="date-item">
                    <span className="date-label">Record Date</span>
                    <span className="date-value">{dividend.record_date}</span>
                  </div>
                  <div className="date-item">
                    <span className="date-label">Payment Date</span>
                    <span className="date-value">{dividend.payment_date}</span>
                  </div>
                </div>

                {dividend.status === 'pending_approval' && dividend.proposal_id && (
                  <div className="proposal-link">
                    <Vote size={14} />
                    <span>Governance Proposal #{dividend.proposal_id}</span>
                  </div>
                )}

                {dividend.audit_hash && (
                  <div className="audit-info">
                    <FileText size={14} />
                    <span>Audit: {dividend.audit_hash.slice(0, 12)}...</span>
                  </div>
                )}

                {dividend.description && (
                  <p className="description">{dividend.description}</p>
                )}
              </div>
            );
          })}
        </div>
      )}

      <style>{styles}</style>
    </div>
  );
}

const styles = `
  .dividends-screen {
    min-height: 100vh;
    padding: 16px;
    padding-bottom: 100px;
  }

  .header {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 20px;
  }

  .header h1 {
    font-size: 24px;
    font-weight: 700;
    color: white;
    margin: 0;
  }

  .back-btn {
    background: none;
    border: none;
    color: white;
    padding: 8px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 60vh;
    color: #8b949e;
    gap: 16px;
  }

  .spin {
    animation: spin 1s linear infinite;
    color: #3B82F6;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .filter-tabs {
    display: flex;
    gap: 8px;
    margin-bottom: 16px;
    overflow-x: auto;
  }

  .filter-tab {
    padding: 8px 16px;
    background: #161B22;
    border: none;
    border-radius: 20px;
    color: #8b949e;
    font-size: 14px;
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.2s;
  }

  .filter-tab.active {
    background: #3B82F6;
    color: white;
  }

  .info-banner {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 12px 16px;
    background: rgba(59, 130, 246, 0.1);
    border-radius: 12px;
    margin-bottom: 20px;
    color: #3B82F6;
  }

  .info-banner p {
    margin: 0;
    font-size: 13px;
    line-height: 1.5;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 60px 20px;
    color: #8b949e;
  }

  .empty-icon {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    background: rgba(139, 148, 158, 0.1);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 16px;
  }

  .empty-state h3 {
    font-size: 18px;
    font-weight: 600;
    color: white;
    margin: 0 0 8px;
  }

  .empty-state p {
    font-size: 14px;
    margin: 0;
    max-width: 280px;
  }

  .dividends-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .dividend-card {
    background: #161B22;
    border-radius: 16px;
    padding: 16px;
  }

  .dividend-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 16px;
  }

  .company-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .company-symbol {
    font-size: 18px;
    font-weight: 700;
    color: white;
  }

  .company-name {
    font-size: 13px;
    color: #8b949e;
  }

  .status-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 500;
  }

  .dividend-details {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 16px;
    padding-bottom: 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .detail-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .detail-row .label {
    font-size: 14px;
    color: #8b949e;
  }

  .detail-row .value {
    font-size: 14px;
    font-weight: 500;
    color: white;
  }

  .detail-row.highlight .value.amount {
    font-size: 16px;
    font-weight: 700;
    color: #10b981;
  }

  .dividend-dates {
    display: flex;
    gap: 16px;
    margin-bottom: 12px;
  }

  .date-item {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .date-label {
    font-size: 11px;
    color: #8b949e;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .date-value {
    font-size: 13px;
    color: white;
    font-weight: 500;
  }

  .proposal-link, .audit-info {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: #8b949e;
    margin-top: 8px;
  }

  .proposal-link {
    color: #8B5CF6;
  }

  .description {
    margin: 12px 0 0;
    font-size: 13px;
    color: #8b949e;
    line-height: 1.5;
  }
`;
