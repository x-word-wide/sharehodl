/**
 * Equity Detail Screen - ShareHODL PLC equity view
 * Shows holdings with chart, send/receive/buy/trade, transaction history, voting
 */

import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useWalletStore } from '../services/walletStore';
import { fetchTransactionHistory } from '../services/blockchainService';
import { Chain } from '../types';

interface TransactionItem {
  hash: string;
  type: 'SEND' | 'RECEIVE' | 'STAKE' | 'UNSTAKE' | 'CLAIM';
  amount: string;
  symbol: string;
  timestamp: number;
  height: number;
  counterparty?: string;
  fee?: string;
}

// Price chart data points (simulated stable price)
const generateChartData = () => {
  const points = [];
  const basePrice = 1.00;
  for (let i = 0; i < 30; i++) {
    // Very slight variation for stablecoin (within 0.1%)
    const variation = (Math.random() - 0.5) * 0.002;
    points.push(basePrice + variation);
  }
  return points;
};

// Time period options
type TimePeriod = '1D' | '1W' | '1M' | '3M' | '1Y' | 'ALL';

export function EquityDetailScreen() {
  const { equityId } = useParams();
  const navigate = useNavigate();
  const tg = window.Telegram?.WebApp;
  const { assets, accounts, refreshBalances } = useWalletStore();

  const [isRefreshing, setIsRefreshing] = useState(false);
  const [selectedPeriod, setSelectedPeriod] = useState<TimePeriod>('1M');
  const [chartData] = useState(generateChartData());
  const [transactions, setTransactions] = useState<TransactionItem[]>([]);
  const [isLoadingTxs, setIsLoadingTxs] = useState(true);
  const [selectedTx, setSelectedTx] = useState<TransactionItem | null>(null);

  // Get ShareHODL account and HODL balance
  // Use raw balance (not balanceFormatted which is "90.00M" -> parseFloat stops at "M")
  const sharehodlAccount = accounts.find(a => a.chain === Chain.SHAREHODL);
  const hodlAsset = assets.find(a => a.token.symbol === 'HODL');
  const hodlBalance = hodlAsset ? parseFloat(hodlAsset.balance) : 0;
  const hodlBalanceUsd = hodlAsset ? parseFloat(hodlAsset.balanceUsd) : 0;
  const address = sharehodlAccount?.address || '';

  // Format balance nicely: whole numbers without decimals, others with 2 decimals max
  const formatBalance = (bal: number): string => {
    if (bal >= 1_000_000) {
      return `${(bal / 1_000_000).toFixed(2)}M`;
    }
    if (bal >= 1_000) {
      return `${(bal / 1_000).toFixed(2)}K`;
    }
    if (Number.isInteger(bal)) {
      return bal.toLocaleString();
    }
    return bal.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 });
  };

  // Load transactions from blockchain
  const loadTransactions = useCallback(async () => {
    if (!address) {
      setTransactions([]);
      setIsLoadingTxs(false);
      return;
    }

    setIsLoadingTxs(true);
    try {
      const result = await fetchTransactionHistory(address, 20);
      setTransactions(result.transactions);
    } catch (error) {
      console.error('Failed to load transactions:', error);
    } finally {
      setIsLoadingTxs(false);
    }
  }, [address]);

  useEffect(() => {
    refreshBalances();
    loadTransactions();
  }, [refreshBalances, loadTransactions]);

  // Only handle sharehodl-plc for now
  if (equityId !== 'sharehodl-plc') {
    return (
      <div className="equity-detail not-found">
        <span className="not-found-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" width="64" height="64">
            <path d="M3 3v18h18" />
            <path d="M18 9l-5 5-4-4-6 6" />
          </svg>
        </span>
        <p className="not-found-title">Equity Not Found</p>
        <p className="not-found-desc">This equity is not available yet.</p>
        <button onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); navigate('/portfolio'); }}>
          Go Back
        </button>
        <style>{notFoundStyles}</style>
      </div>
    );
  }

  const handleRefresh = async () => {
    setIsRefreshing(true);
    tg?.HapticFeedback?.impactOccurred('light');
    await Promise.all([refreshBalances(), loadTransactions()]);
    setTimeout(() => setIsRefreshing(false), 500);
  };

  const handleSend = () => {
    tg?.HapticFeedback?.impactOccurred('medium');
    navigate('/send');
  };

  const handleReceive = () => {
    tg?.HapticFeedback?.impactOccurred('medium');
    navigate('/receive');
  };

  const handleBuy = () => {
    tg?.HapticFeedback?.impactOccurred('medium');
    navigate('/trade');
  };

  const handleTrade = () => {
    tg?.HapticFeedback?.impactOccurred('medium');
    navigate('/trade');
  };

  const handleVote = () => {
    tg?.HapticFeedback?.impactOccurred('medium');
    navigate('/governance');
  };

  // Simple SVG chart
  const renderChart = () => {
    const width = 320;
    const height = 120;
    const padding = 10;
    const chartWidth = width - padding * 2;
    const chartHeight = height - padding * 2;

    const minPrice = Math.min(...chartData);
    const maxPrice = Math.max(...chartData);
    const priceRange = maxPrice - minPrice || 0.01;

    const points = chartData.map((price, i) => {
      const x = padding + (i / (chartData.length - 1)) * chartWidth;
      const y = padding + chartHeight - ((price - minPrice) / priceRange) * chartHeight;
      return `${x},${y}`;
    }).join(' ');

    // Fill area under the line
    const fillPoints = `${padding},${height - padding} ${points} ${width - padding},${height - padding}`;

    return (
      <svg viewBox={`0 0 ${width} ${height}`} className="price-chart">
        <defs>
          <linearGradient id="chartGradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor="#1E40AF" stopOpacity="0.3" />
            <stop offset="100%" stopColor="#1E40AF" stopOpacity="0" />
          </linearGradient>
        </defs>
        <polygon points={fillPoints} fill="url(#chartGradient)" />
        <polyline
          points={points}
          fill="none"
          stroke="#3B82F6"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    );
  };

  const formatDate = (timestamp: number) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffDays = Math.floor((now.getTime() - date.getTime()) / 86400000);

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  const getTransactionIcon = (type: string) => {
    switch (type) {
      case 'RECEIVE':
        return (
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 19V5M5 12l7 7 7-7" />
          </svg>
        );
      case 'SEND':
        return (
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 5v14M5 12l7-7 7 7" />
          </svg>
        );
      case 'STAKE':
        return (
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
          </svg>
        );
      case 'CLAIM':
        return (
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="10" />
            <path d="M12 6v6l4 2" />
          </svg>
        );
      default:
        return (
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="10" />
          </svg>
        );
    }
  };

  return (
    <div className="equity-detail-screen">
      {/* Header */}
      <div className="detail-header">
        <button className="back-btn" onClick={() => navigate(-1)}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M15 18l-6-6 6-6" />
          </svg>
        </button>
        <div className="header-title">
          <span className="equity-name">ShareHODL PLC</span>
          <span className="equity-symbol">HODL</span>
        </div>
        <button className={`refresh-btn ${isRefreshing ? 'spinning' : ''}`} onClick={handleRefresh}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 12a9 9 0 11-9-9c2.52 0 4.93 1 6.74 2.74L21 8" />
            <path d="M21 3v5h-5" />
          </svg>
        </button>
      </div>

      {/* Balance Section */}
      <div className="balance-section">
        <div className="equity-icon">
          <span>SH</span>
        </div>
        <h1 className="balance-amount">{formatBalance(hodlBalance)} HODL</h1>
        <p className="balance-usd">${hodlBalanceUsd.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</p>
        <div className="price-info">
          <span className="current-price">$1.00 per share</span>
          <span className="price-badge stable">USD Pegged</span>
        </div>
      </div>

      {/* Chart Section */}
      <div className="chart-section">
        <div className="chart-container">
          {renderChart()}
        </div>
        <div className="period-selector">
          {(['1D', '1W', '1M', '3M', '1Y', 'ALL'] as TimePeriod[]).map((period) => (
            <button
              key={period}
              className={`period-btn ${selectedPeriod === period ? 'active' : ''}`}
              onClick={() => {
                tg?.HapticFeedback?.selectionChanged();
                setSelectedPeriod(period);
              }}
            >
              {period}
            </button>
          ))}
        </div>
      </div>

      {/* Action Buttons */}
      <div className="action-buttons">
        <button className="action-btn" onClick={handleReceive}>
          <div className="action-icon receive">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 19V5M5 12l7 7 7-7" />
            </svg>
          </div>
          <span>Receive</span>
        </button>
        <button className="action-btn" onClick={handleSend}>
          <div className="action-icon send">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 5v14M5 12l7-7 7 7" />
            </svg>
          </div>
          <span>Send</span>
        </button>
        <button className="action-btn" onClick={handleBuy}>
          <div className="action-icon buy">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 8v8M8 12h8" />
            </svg>
          </div>
          <span>Buy</span>
        </button>
        <button className="action-btn" onClick={handleTrade}>
          <div className="action-icon trade">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M7 10l5 5 5-5" />
              <path d="M17 14l-5-5-5 5" />
            </svg>
          </div>
          <span>Trade</span>
        </button>
      </div>

      {/* Governance/Vote Section */}
      <div className="vote-section" onClick={handleVote}>
        <div className="vote-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M12 2L2 7l10 5 10-5-10-5z" />
            <path d="M2 17l10 5 10-5" />
            <path d="M2 12l10 5 10-5" />
          </svg>
        </div>
        <div className="vote-info">
          <span className="vote-title">Governance & Voting</span>
          <span className="vote-desc">Participate in protocol decisions</span>
        </div>
        <div className="vote-arrow">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M9 18l6-6-6-6" />
          </svg>
        </div>
      </div>

      {/* Address Section */}
      {address && (
        <div className="address-section">
          <span className="address-label">Your HODL Address</span>
          <button
            className="address-value"
            onClick={() => {
              navigator.clipboard.writeText(address);
              tg?.HapticFeedback?.notificationOccurred('success');
              tg?.showAlert('Address copied!');
            }}
          >
            <span>{address.slice(0, 12)}...{address.slice(-10)}</span>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
              <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
            </svg>
          </button>
        </div>
      )}

      {/* Stats Section */}
      <div className="stats-section">
        <h3 className="section-title">Token Stats</h3>
        <div className="stats-grid">
          <div className="stat-item">
            <span className="stat-label">Price</span>
            <span className="stat-value">$1.00</span>
          </div>
          <div className="stat-item">
            <span className="stat-label">24h Change</span>
            <span className="stat-value stable">0.00%</span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Type</span>
            <span className="stat-value">Stablecoin</span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Chain</span>
            <span className="stat-value">ShareHODL</span>
          </div>
        </div>
      </div>

      {/* Transaction History */}
      <div className="transactions-section">
        <h3 className="section-title">Transaction History</h3>
        {isLoadingTxs ? (
          <div className="empty-transactions">
            <div className="spinner" />
            <p>Loading transactions...</p>
          </div>
        ) : transactions.length > 0 ? (
          <div className="transactions-list">
            {transactions.map((tx) => {
              const isIncoming = tx.type === 'RECEIVE' || tx.type === 'CLAIM';
              return (
                <div
                  key={tx.hash}
                  className="transaction-item"
                  onClick={() => {
                    tg?.HapticFeedback?.impactOccurred('light');
                    setSelectedTx(tx);
                  }}
                >
                  <div className={`tx-icon ${tx.type.toLowerCase()}`}>
                    {getTransactionIcon(tx.type)}
                  </div>
                  <div className="tx-info">
                    <span className="tx-type">
                      {tx.type === 'RECEIVE' ? 'Received' :
                       tx.type === 'SEND' ? 'Sent' :
                       tx.type === 'STAKE' ? 'Staked' :
                       tx.type === 'UNSTAKE' ? 'Unstaked' : 'Claimed'}
                    </span>
                    {tx.counterparty ? (
                      <span className="tx-counterparty">
                        {isIncoming ? 'From: ' : 'To: '}
                        {tx.counterparty.slice(0, 8)}...{tx.counterparty.slice(-6)}
                      </span>
                    ) : (
                      <span className="tx-date">{formatDate(tx.timestamp)}</span>
                    )}
                  </div>
                  <div className="tx-amount">
                    <span className={isIncoming ? 'positive' : 'negative'}>
                      {isIncoming ? '+' : '-'}{tx.amount} HODL
                    </span>
                    <span className="tx-date-small">{formatDate(tx.timestamp)}</span>
                  </div>
                  <div className="tx-chevron">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M9 18l6-6-6-6" />
                    </svg>
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="empty-transactions">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" width="48" height="48">
              <rect x="3" y="4" width="18" height="16" rx="2" />
              <path d="M7 8h10" />
              <path d="M7 12h6" />
              <path d="M7 16h4" />
            </svg>
            <p>No transactions yet</p>
            <span>Your HODL transactions will appear here</span>
          </div>
        )}
      </div>

      {/* Transaction Detail Modal */}
      {selectedTx && (
        <div className="tx-modal-overlay" onClick={() => setSelectedTx(null)}>
          <div className="tx-modal" onClick={(e) => e.stopPropagation()}>
            <div className="tx-modal-header">
              <div className={`tx-modal-icon ${selectedTx.type.toLowerCase()}`}>
                {getTransactionIcon(selectedTx.type)}
              </div>
              <h2 className="tx-modal-title">
                {selectedTx.type === 'RECEIVE' ? 'Received' :
                 selectedTx.type === 'SEND' ? 'Sent' :
                 selectedTx.type === 'STAKE' ? 'Staked' :
                 selectedTx.type === 'UNSTAKE' ? 'Unstaked' : 'Claimed'}
              </h2>
              <button className="tx-modal-close" onClick={() => setSelectedTx(null)}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M18 6L6 18M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div className="tx-modal-amount">
              <span className={selectedTx.type === 'RECEIVE' || selectedTx.type === 'CLAIM' ? 'positive' : 'negative'}>
                {selectedTx.type === 'RECEIVE' || selectedTx.type === 'CLAIM' ? '+' : '-'}{selectedTx.amount} HODL
              </span>
              <span className="tx-modal-usd">~${selectedTx.amount} USD</span>
            </div>

            <div className="tx-modal-status">
              <div className="status-badge success">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" width="14" height="14">
                  <path d="M20 6L9 17l-5-5" />
                </svg>
                Confirmed
              </div>
            </div>

            <div className="tx-modal-details">
              <div className="tx-modal-row">
                <span className="tx-modal-label">Transaction Hash</span>
                <button
                  className="tx-modal-value copyable"
                  onClick={() => {
                    navigator.clipboard.writeText(selectedTx.hash);
                    tg?.HapticFeedback?.notificationOccurred('success');
                    tg?.showAlert('Hash copied!');
                  }}
                >
                  <span className="hash-text">{selectedTx.hash}</span>
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" width="16" height="16">
                    <rect x="9" y="9" width="13" height="13" rx="2" />
                    <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
                  </svg>
                </button>
              </div>

              <div className="tx-modal-row">
                <span className="tx-modal-label">Block Height</span>
                <span className="tx-modal-value">#{selectedTx.height.toLocaleString()}</span>
              </div>

              <div className="tx-modal-row">
                <span className="tx-modal-label">Date & Time</span>
                <span className="tx-modal-value">
                  {new Date(selectedTx.timestamp).toLocaleDateString('en-US', {
                    year: 'numeric',
                    month: 'short',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit'
                  })}
                </span>
              </div>

              {selectedTx.counterparty && (
                <div className="tx-modal-row">
                  <span className="tx-modal-label">
                    {selectedTx.type === 'RECEIVE' || selectedTx.type === 'CLAIM' ? 'From Address' : 'To Address'}
                  </span>
                  <button
                    className="tx-modal-value copyable"
                    onClick={() => {
                      navigator.clipboard.writeText(selectedTx.counterparty!);
                      tg?.HapticFeedback?.notificationOccurred('success');
                      tg?.showAlert('Address copied!');
                    }}
                  >
                    <span className="hash-text">{selectedTx.counterparty}</span>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" width="16" height="16">
                      <rect x="9" y="9" width="13" height="13" rx="2" />
                      <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
                    </svg>
                  </button>
                </div>
              )}

              <div className="tx-modal-row">
                <span className="tx-modal-label">Network Fee</span>
                <span className="tx-modal-value">{selectedTx.fee || '~0.005'} HODL</span>
              </div>

              <div className="tx-modal-row">
                <span className="tx-modal-label">Asset</span>
                <div className="tx-modal-asset">
                  <div className="asset-icon">SH</div>
                  <span>HODL (ShareHODL)</span>
                </div>
              </div>

              <div className="tx-modal-row">
                <span className="tx-modal-label">Network</span>
                <span className="tx-modal-value">ShareHODL Mainnet</span>
              </div>
            </div>

            <button
              className="tx-modal-explorer-btn"
              onClick={() => {
                // Could open block explorer if available
                tg?.HapticFeedback?.impactOccurred('light');
                navigator.clipboard.writeText(selectedTx.hash);
                tg?.showAlert('Transaction hash copied!');
              }}
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" width="18" height="18">
                <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6" />
                <path d="M15 3h6v6" />
                <path d="M10 14L21 3" />
              </svg>
              Copy Full Transaction Hash
            </button>
          </div>
        </div>
      )}

      <style>{styles}</style>
    </div>
  );
}

const notFoundStyles = `
  .equity-detail.not-found {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    text-align: center;
    background: var(--tg-theme-bg-color, #0D1117);
  }
  .not-found-icon {
    color: var(--text-secondary, #8b949e);
    margin-bottom: 16px;
  }
  .not-found-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--text-primary, white);
    margin: 0 0 8px;
  }
  .not-found-desc {
    font-size: 14px;
    color: var(--text-secondary, #8b949e);
    margin: 0 0 24px;
    max-width: 280px;
    line-height: 1.5;
  }
  .equity-detail.not-found button {
    padding: 14px 28px;
    background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
    border: none;
    border-radius: 12px;
    color: white;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
    transition: transform 0.2s;
  }
  .equity-detail.not-found button:active {
    transform: scale(0.97);
  }
`;

const styles = `
  .equity-detail-screen {
    min-height: 100vh;
    padding-bottom: 100px;
    background: var(--tg-theme-bg-color, #0D1117);
  }

  .detail-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px;
  }

  .back-btn, .refresh-btn {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: var(--surface-bg, #161B22);
    border: 1px solid var(--border-color, #30363d);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .back-btn svg, .refresh-btn svg {
    width: 20px;
    height: 20px;
    color: var(--text-secondary, #8b949e);
  }

  .refresh-btn.spinning svg {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .header-title {
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .equity-name {
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary, white);
  }

  .equity-symbol {
    font-size: 12px;
    color: #3B82F6;
    font-weight: 500;
  }

  .balance-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 20px 16px 24px;
  }

  .equity-icon {
    width: 72px;
    height: 72px;
    border-radius: 50%;
    background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 24px;
    font-weight: 700;
    color: white;
    margin-bottom: 20px;
    box-shadow: 0 8px 24px rgba(30, 64, 175, 0.3);
  }

  .balance-amount {
    font-size: 36px;
    font-weight: 700;
    color: var(--text-primary, white);
    margin: 0;
  }

  .balance-usd {
    font-size: 18px;
    color: var(--text-secondary, #8b949e);
    margin: 8px 0 16px;
  }

  .price-info {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .current-price {
    font-size: 15px;
    color: var(--text-secondary, #8b949e);
  }

  .price-badge {
    font-size: 12px;
    font-weight: 600;
    padding: 4px 10px;
    border-radius: 8px;
  }

  .price-badge.stable {
    color: #10b981;
    background: rgba(16, 185, 129, 0.1);
  }

  .chart-section {
    padding: 0 16px 24px;
  }

  .chart-container {
    background: var(--surface-bg, #161B22);
    border: 1px solid var(--border-color, #30363d);
    border-radius: 16px;
    padding: 16px;
    margin-bottom: 12px;
  }

  .price-chart {
    width: 100%;
    height: 120px;
  }

  .period-selector {
    display: flex;
    gap: 8px;
    justify-content: center;
  }

  .period-btn {
    padding: 8px 14px;
    border-radius: 8px;
    border: none;
    background: var(--surface-bg, #161B22);
    color: var(--text-secondary, #8b949e);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .period-btn.active {
    background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
    color: white;
  }

  .action-buttons {
    display: flex;
    justify-content: center;
    gap: 16px;
    padding: 0 16px 24px;
  }

  .action-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    background: none;
    border: none;
    cursor: pointer;
  }

  .action-icon {
    width: 52px;
    height: 52px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.2s;
  }

  .action-icon.receive {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  }

  .action-icon.send {
    background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
  }

  .action-icon.buy {
    background: linear-gradient(135deg, #8B5CF6 0%, #7C3AED 100%);
  }

  .action-icon.trade {
    background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%);
  }

  .action-icon svg {
    width: 22px;
    height: 22px;
    color: white;
  }

  .action-btn span {
    font-size: 12px;
    font-weight: 500;
    color: var(--text-secondary, #8b949e);
  }

  .action-btn:active .action-icon {
    transform: scale(0.95);
  }

  .vote-section {
    margin: 0 16px 20px;
    padding: 16px;
    background: linear-gradient(135deg, rgba(139, 92, 246, 0.15) 0%, rgba(124, 58, 237, 0.1) 100%);
    border: 1px solid rgba(139, 92, 246, 0.3);
    border-radius: 16px;
    display: flex;
    align-items: center;
    gap: 14px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .vote-section:active {
    transform: scale(0.98);
  }

  .vote-icon {
    width: 44px;
    height: 44px;
    border-radius: 12px;
    background: rgba(139, 92, 246, 0.2);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .vote-icon svg {
    width: 24px;
    height: 24px;
    color: #8B5CF6;
  }

  .vote-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .vote-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--text-primary, white);
  }

  .vote-desc {
    font-size: 13px;
    color: var(--text-secondary, #8b949e);
  }

  .vote-arrow {
    width: 24px;
    height: 24px;
    color: var(--text-secondary, #8b949e);
  }

  .vote-arrow svg {
    width: 24px;
    height: 24px;
  }

  .address-section {
    margin: 0 16px 20px;
    padding: 16px;
    background: var(--surface-bg, #161B22);
    border: 1px solid var(--border-color, #30363d);
    border-radius: 14px;
  }

  .address-label {
    display: block;
    font-size: 13px;
    color: var(--text-secondary, #8b949e);
    margin-bottom: 8px;
  }

  .address-value {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    padding: 0;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-primary, white);
    font-size: 14px;
    font-family: monospace;
  }

  .address-value svg {
    width: 18px;
    height: 18px;
    color: var(--text-secondary, #8b949e);
  }

  .stats-section {
    padding: 0 16px 20px;
  }

  .section-title {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary, white);
    margin: 0 0 12px;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .stat-item {
    padding: 14px;
    background: var(--surface-bg, #161B22);
    border: 1px solid var(--border-color, #30363d);
    border-radius: 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .stat-label {
    font-size: 12px;
    color: var(--text-secondary, #8b949e);
  }

  .stat-value {
    font-size: 15px;
    font-weight: 600;
    color: var(--text-primary, white);
  }

  .stat-value.stable {
    color: #10b981;
  }

  .transactions-section {
    padding: 0 16px;
  }

  .transactions-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .transaction-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 14px;
    background: var(--surface-bg, #161B22);
    border: 1px solid var(--border-color, #30363d);
    border-radius: 14px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .transaction-item:active {
    background: rgba(255, 255, 255, 0.05);
    transform: scale(0.98);
  }

  .tx-icon {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .tx-icon.receive, .tx-icon.claim {
    background: rgba(16, 185, 129, 0.1);
    color: #10b981;
  }

  .tx-icon.send {
    background: rgba(239, 68, 68, 0.1);
    color: #ef4444;
  }

  .tx-icon.stake, .tx-icon.unstake {
    background: rgba(59, 130, 246, 0.1);
    color: #3B82F6;
  }

  .tx-icon svg {
    width: 20px;
    height: 20px;
  }

  .tx-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .tx-type {
    font-size: 15px;
    font-weight: 500;
    color: var(--text-primary, white);
  }

  .tx-date {
    font-size: 13px;
    color: var(--text-secondary, #8b949e);
  }

  .tx-counterparty {
    font-size: 12px;
    color: var(--text-secondary, #8b949e);
    font-family: monospace;
  }

  .tx-amount {
    text-align: right;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 2px;
  }

  .tx-date-small {
    font-size: 11px;
    color: var(--text-secondary, #8b949e);
  }

  .tx-amount .positive {
    color: #10b981;
    font-weight: 600;
  }

  .tx-amount .negative {
    color: #ef4444;
    font-weight: 600;
  }

  .tx-chevron {
    width: 20px;
    height: 20px;
    color: var(--text-secondary, #8b949e);
    flex-shrink: 0;
  }

  .tx-chevron svg {
    width: 20px;
    height: 20px;
  }

  .empty-transactions {
    padding: 40px 20px;
    text-align: center;
    background: var(--surface-bg, #161B22);
    border: 1px solid var(--border-color, #30363d);
    border-radius: 14px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .empty-transactions svg {
    color: var(--text-secondary, #8b949e);
    margin-bottom: 8px;
  }

  .empty-transactions p {
    font-size: 15px;
    font-weight: 500;
    color: var(--text-primary, white);
    margin: 0;
  }

  .empty-transactions span {
    font-size: 13px;
    color: var(--text-secondary, #8b949e);
  }

  /* Transaction Modal */
  .tx-modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.8);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: flex-end;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn 0.2s ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .tx-modal {
    width: 100%;
    max-width: 500px;
    max-height: 90vh;
    overflow-y: auto;
    background: var(--tg-theme-bg-color, #0D1117);
    border-radius: 24px 24px 0 0;
    padding: 20px;
    animation: slideUp 0.3s ease;
  }

  @keyframes slideUp {
    from { transform: translateY(100%); }
    to { transform: translateY(0); }
  }

  .tx-modal-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 20px;
  }

  .tx-modal-icon {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .tx-modal-icon.receive, .tx-modal-icon.claim {
    background: rgba(16, 185, 129, 0.15);
    color: #10b981;
  }

  .tx-modal-icon.send {
    background: rgba(239, 68, 68, 0.15);
    color: #ef4444;
  }

  .tx-modal-icon.stake, .tx-modal-icon.unstake {
    background: rgba(59, 130, 246, 0.15);
    color: #3B82F6;
  }

  .tx-modal-icon svg {
    width: 24px;
    height: 24px;
  }

  .tx-modal-title {
    flex: 1;
    font-size: 20px;
    font-weight: 600;
    color: var(--text-primary, white);
    margin: 0;
  }

  .tx-modal-close {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: var(--surface-bg, #161B22);
    border: 1px solid var(--border-color, #30363d);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    color: var(--text-secondary, #8b949e);
  }

  .tx-modal-close svg {
    width: 18px;
    height: 18px;
  }

  .tx-modal-amount {
    text-align: center;
    padding: 20px 0;
    border-bottom: 1px solid var(--border-color, #30363d);
    margin-bottom: 16px;
  }

  .tx-modal-amount .positive {
    font-size: 32px;
    font-weight: 700;
    color: #10b981;
    display: block;
  }

  .tx-modal-amount .negative {
    font-size: 32px;
    font-weight: 700;
    color: #ef4444;
    display: block;
  }

  .tx-modal-usd {
    font-size: 16px;
    color: var(--text-secondary, #8b949e);
    margin-top: 4px;
    display: block;
  }

  .tx-modal-status {
    display: flex;
    justify-content: center;
    margin-bottom: 20px;
  }

  .status-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 14px;
    border-radius: 20px;
    font-size: 13px;
    font-weight: 600;
  }

  .status-badge.success {
    background: rgba(16, 185, 129, 0.15);
    color: #10b981;
  }

  .tx-modal-details {
    background: var(--surface-bg, #161B22);
    border: 1px solid var(--border-color, #30363d);
    border-radius: 16px;
    overflow: hidden;
  }

  .tx-modal-row {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    padding: 14px 16px;
    border-bottom: 1px solid var(--border-color, #30363d);
    gap: 12px;
  }

  .tx-modal-row:last-child {
    border-bottom: none;
  }

  .tx-modal-label {
    font-size: 14px;
    color: var(--text-secondary, #8b949e);
    flex-shrink: 0;
  }

  .tx-modal-value {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary, white);
    text-align: right;
    word-break: break-all;
  }

  .tx-modal-value.copyable {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 0;
    background: none;
    border: none;
    cursor: pointer;
    color: #3B82F6;
    text-align: right;
  }

  .tx-modal-value.copyable:active {
    opacity: 0.7;
  }

  .tx-modal-value.copyable svg {
    flex-shrink: 0;
    margin-top: 2px;
    color: var(--text-secondary, #8b949e);
  }

  .hash-text {
    font-family: monospace;
    font-size: 12px;
    line-height: 1.4;
    word-break: break-all;
    max-width: 180px;
  }

  .tx-modal-asset {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .tx-modal-asset .asset-icon {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    font-weight: 700;
    color: white;
  }

  .tx-modal-explorer-btn {
    width: 100%;
    margin-top: 20px;
    padding: 16px;
    background: var(--surface-bg, #161B22);
    border: 1px solid var(--border-color, #30363d);
    border-radius: 14px;
    color: #3B82F6;
    font-size: 15px;
    font-weight: 600;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    cursor: pointer;
    transition: background 0.2s;
  }

  .tx-modal-explorer-btn:active {
    background: rgba(59, 130, 246, 0.1);
  }
`;
