/**
 * Equity Detail Screen - Portfolio view for equity holdings
 * Shows user's holdings with send/receive/trade like crypto assets
 */

import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

// Demo equity data with user holdings
const DEMO_EQUITIES: Record<string, {
  id: string;
  symbol: string;
  name: string;
  shares: number;
  pricePerShare: number;
  change24h: number;
  color: string;
  sector: string;
  avgCost: number;
  walletAddress: string;
}> = {
  'sharehodl-plc': {
    id: 'sharehodl-plc',
    symbol: 'SHDL',
    name: 'ShareHODL PLC',
    shares: 1250,
    pricePerShare: 12.50,
    change24h: 4.25,
    color: '#3B82F6',
    sector: 'Technology',
    avgCost: 10.00,
    walletAddress: 'sharehodl1qxp4z7v8k2m3n5j6h9f0g1d2s3a4w5e6r7t8y9u0i'
  },
  'property-mainnet': {
    id: 'property-mainnet',
    symbol: 'PROP',
    name: 'Property Mainnet',
    shares: 500,
    pricePerShare: 8.75,
    change24h: -1.30,
    color: '#10B981',
    sector: 'Real Estate',
    avgCost: 7.50,
    walletAddress: 'sharehodl1qxp4z7v8k2m3n5j6h9f0g1d2s3a4w5e6r7t8y9u0i'
  },
  'tech-ventures': {
    id: 'tech-ventures',
    symbol: 'TVNT',
    name: 'Tech Ventures Ltd',
    shares: 2000,
    pricePerShare: 3.20,
    change24h: 7.80,
    color: '#F59E0B',
    sector: 'Venture Capital',
    avgCost: 2.40,
    walletAddress: 'sharehodl1qxp4z7v8k2m3n5j6h9f0g1d2s3a4w5e6r7t8y9u0i'
  },
  'green-energy': {
    id: 'green-energy',
    symbol: 'GREN',
    name: 'Green Energy Corp',
    shares: 750,
    pricePerShare: 15.00,
    change24h: 2.15,
    color: '#059669',
    sector: 'Energy',
    avgCost: 12.80,
    walletAddress: 'sharehodl1qxp4z7v8k2m3n5j6h9f0g1d2s3a4w5e6r7t8y9u0i'
  },
  'fintech-global': {
    id: 'fintech-global',
    symbol: 'FNTK',
    name: 'FinTech Global',
    shares: 300,
    pricePerShare: 45.00,
    change24h: -0.50,
    color: '#6366F1',
    sector: 'Financial Services',
    avgCost: 38.00,
    walletAddress: 'sharehodl1qxp4z7v8k2m3n5j6h9f0g1d2s3a4w5e6r7t8y9u0i'
  }
};

// Demo transactions
const generateTransactions = (equity: typeof DEMO_EQUITIES[string]) => [
  {
    id: '1',
    type: 'BUY' as const,
    shares: 250,
    pricePerShare: equity.avgCost * 0.95,
    total: 250 * equity.avgCost * 0.95,
    date: Date.now() - 86400000 * 3,
    status: 'completed'
  },
  {
    id: '2',
    type: 'DIVIDEND' as const,
    shares: 0,
    pricePerShare: 0,
    total: equity.shares * 0.02,
    date: Date.now() - 86400000 * 15,
    status: 'completed'
  },
  {
    id: '3',
    type: 'BUY' as const,
    shares: 500,
    pricePerShare: equity.avgCost * 1.02,
    total: 500 * equity.avgCost * 1.02,
    date: Date.now() - 86400000 * 30,
    status: 'completed'
  },
  {
    id: '4',
    type: 'SELL' as const,
    shares: 100,
    pricePerShare: equity.pricePerShare * 0.98,
    total: 100 * equity.pricePerShare * 0.98,
    date: Date.now() - 86400000 * 45,
    status: 'completed'
  }
];

// Generate chart data
const generateChartData = (basePrice: number, change: number) => {
  const points = [];
  const variation = basePrice * 0.06;
  let price = basePrice - (change / 100) * basePrice;

  for (let i = 0; i < 30; i++) {
    const noise = (Math.random() - 0.5) * variation * 0.5;
    const trend = ((change / 100) * basePrice * (i / 29));
    price = basePrice - (change / 100) * basePrice + trend + noise;
    points.push(Math.max(price, basePrice * 0.5));
  }
  return points;
};

export function EquityDetailScreen() {
  const { equityId } = useParams();
  const navigate = useNavigate();
  const tg = window.Telegram?.WebApp;
  const [isRefreshing, setIsRefreshing] = useState(false);

  const equity = equityId ? DEMO_EQUITIES[equityId] : null;

  if (!equity) {
    return (
      <div className="equity-detail not-found">
        <p>Equity not found</p>
        <button onClick={() => navigate('/portfolio')}>Go Back</button>
        <style>{notFoundStyles}</style>
      </div>
    );
  }

  const totalValue = equity.shares * equity.pricePerShare;
  const totalCost = equity.shares * equity.avgCost;
  const profitLoss = totalValue - totalCost;
  const profitLossPercent = ((profitLoss / totalCost) * 100);
  const isPositive = equity.change24h >= 0;
  const isProfitable = profitLoss >= 0;

  const transactions = generateTransactions(equity);
  const chartData = generateChartData(equity.pricePerShare, equity.change24h);

  // Build chart path
  const minPrice = Math.min(...chartData);
  const maxPrice = Math.max(...chartData);
  const range = maxPrice - minPrice || 1;

  const chartWidth = 300;
  const chartHeight = 80;

  const pathPoints = chartData.map((price, i) => {
    const x = (i / (chartData.length - 1)) * chartWidth;
    const y = chartHeight - ((price - minPrice) / range) * chartHeight;
    return { x, y };
  });

  const linePath = pathPoints.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x.toFixed(1)} ${p.y.toFixed(1)}`).join(' ');
  const areaPath = `${linePath} L ${chartWidth} ${chartHeight} L 0 ${chartHeight} Z`;

  const handleRefresh = async () => {
    setIsRefreshing(true);
    tg?.HapticFeedback?.impactOccurred('light');
    setTimeout(() => setIsRefreshing(false), 800);
  };

  const handleAction = (action: 'send' | 'receive' | 'trade') => {
    tg?.HapticFeedback?.impactOccurred('medium');
    tg?.showAlert(`${action.charAt(0).toUpperCase() + action.slice(1)} ${equity.symbol} - Coming soon!`);
  };

  const copyAddress = () => {
    navigator.clipboard.writeText(equity.walletAddress);
    tg?.HapticFeedback?.notificationOccurred('success');
    tg?.showAlert('Address copied!');
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

  return (
    <div className="equity-detail">
      {/* Header */}
      <header className="detail-header">
        <button className="back-btn" onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); navigate(-1); }}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <path d="M15 18l-6-6 6-6" />
          </svg>
        </button>
        <div className="header-title">
          <span className="equity-name">{equity.name}</span>
          <span className="equity-sector">{equity.sector}</span>
        </div>
        <button className={`refresh-btn ${isRefreshing ? 'spinning' : ''}`} onClick={handleRefresh}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 12a9 9 0 11-9-9c2.52 0 4.93 1 6.74 2.74L21 8" />
            <path d="M21 3v5h-5" />
          </svg>
        </button>
      </header>

      {/* Balance Section */}
      <div className="balance-section">
        <div className="equity-icon" style={{ background: `${equity.color}20`, color: equity.color }}>
          {equity.symbol.slice(0, 2)}
        </div>

        <h1 className="share-count">{equity.shares.toLocaleString()} shares</h1>
        <p className="value-usd">${totalValue.toLocaleString(undefined, { minimumFractionDigits: 2 })}</p>

        <div className="price-info">
          <span className="current-price">${equity.pricePerShare.toFixed(2)}/share</span>
          <span className={`price-change ${isPositive ? 'positive' : 'negative'}`}>
            {isPositive ? '+' : ''}{equity.change24h.toFixed(2)}%
          </span>
        </div>
      </div>

      {/* Mini Chart */}
      <div className="chart-section">
        <svg
          viewBox={`0 0 ${chartWidth} ${chartHeight}`}
          preserveAspectRatio="none"
          className="mini-chart"
        >
          <defs>
            <linearGradient id={`chart-grad-${equity.id}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={isPositive ? '#10b981' : '#ef4444'} stopOpacity="0.25" />
              <stop offset="100%" stopColor={isPositive ? '#10b981' : '#ef4444'} stopOpacity="0.02" />
            </linearGradient>
          </defs>
          <path d={areaPath} fill={`url(#chart-grad-${equity.id})`} />
          <path d={linePath} fill="none" stroke={isPositive ? '#10b981' : '#ef4444'} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>

      {/* Action Buttons */}
      <div className="action-buttons">
        <button className="action-btn" onClick={() => handleAction('send')}>
          <div className="action-icon send">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 5v14M5 12l7-7 7 7" />
            </svg>
          </div>
          <span>Send</span>
        </button>
        <button className="action-btn" onClick={() => handleAction('receive')}>
          <div className="action-icon receive">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 19V5M5 12l7 7 7-7" />
            </svg>
          </div>
          <span>Receive</span>
        </button>
        <button className="action-btn" onClick={() => handleAction('trade')}>
          <div className="action-icon trade">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M7 10l5 5 5-5" />
              <path d="M17 14l-5-5-5 5" />
            </svg>
          </div>
          <span>Trade</span>
        </button>
      </div>

      {/* Holdings Stats Card */}
      <div className="stats-card">
        <h3 className="card-title">Your Holdings</h3>
        <div className="stats-grid">
          <div className="stat-item">
            <span className="stat-label">Market Value</span>
            <span className="stat-value">${totalValue.toLocaleString()}</span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Total Shares</span>
            <span className="stat-value">{equity.shares.toLocaleString()}</span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Avg Cost</span>
            <span className="stat-value">${equity.avgCost.toFixed(2)}</span>
          </div>
          <div className="stat-item">
            <span className="stat-label">Total P/L</span>
            <span className={`stat-value ${isProfitable ? 'profit' : 'loss'}`}>
              {isProfitable ? '+' : ''}${profitLoss.toLocaleString(undefined, { maximumFractionDigits: 0 })}
              <small> ({isProfitable ? '+' : ''}{profitLossPercent.toFixed(1)}%)</small>
            </span>
          </div>
        </div>
      </div>

      {/* Wallet Address */}
      <div className="address-section">
        <span className="address-label">ShareHODL Equity Address</span>
        <button className="address-value" onClick={copyAddress}>
          <span>{equity.walletAddress.slice(0, 14)}...{equity.walletAddress.slice(-8)}</span>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="9" y="9" width="13" height="13" rx="2" />
            <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
          </svg>
        </button>
      </div>

      {/* Transaction History */}
      <div className="transactions-section">
        <h3 className="section-title">Transaction History</h3>
        <div className="transactions-list">
          {transactions.map((tx) => (
            <div key={tx.id} className="transaction-item">
              <div className={`tx-icon ${tx.type.toLowerCase()}`}>
                {tx.type === 'BUY' && (
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M12 19V5M5 12l7 7 7-7" />
                  </svg>
                )}
                {tx.type === 'SELL' && (
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M12 5v14M5 12l7-7 7 7" />
                  </svg>
                )}
                {tx.type === 'DIVIDEND' && (
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="12" cy="12" r="10" />
                    <path d="M12 6v12M6 12h12" />
                  </svg>
                )}
              </div>
              <div className="tx-info">
                <span className="tx-type">
                  {tx.type === 'BUY' && `Bought ${tx.shares} shares`}
                  {tx.type === 'SELL' && `Sold ${tx.shares} shares`}
                  {tx.type === 'DIVIDEND' && 'Dividend Received'}
                </span>
                <span className="tx-date">{formatDate(tx.date)}</span>
              </div>
              <div className="tx-amount">
                <span className={tx.type === 'SELL' ? 'positive' : tx.type === 'DIVIDEND' ? 'dividend' : 'negative'}>
                  {tx.type === 'SELL' || tx.type === 'DIVIDEND' ? '+' : '-'}${tx.total.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>

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
    gap: 16px;
    color: white;
  }
  .equity-detail.not-found button {
    padding: 12px 24px;
    background: #3B82F6;
    border: none;
    border-radius: 10px;
    color: white;
    font-weight: 600;
  }
`;

const styles = `
  .equity-detail {
    min-height: 100vh;
    padding-bottom: 100px;
    background: #0D1117;
  }

  .detail-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px;
    position: sticky;
    top: 0;
    background: rgba(13, 17, 23, 0.95);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    z-index: 10;
  }

  .back-btn, .refresh-btn {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    background: rgba(48, 54, 61, 0.6);
    border: none;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s;
  }

  .back-btn:active, .refresh-btn:active {
    transform: scale(0.95);
  }

  .back-btn svg, .refresh-btn svg {
    width: 20px;
    height: 20px;
    color: #8b949e;
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
    font-size: 16px;
    font-weight: 600;
    color: white;
  }

  .equity-sector {
    font-size: 12px;
    color: #8b949e;
  }

  .balance-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 20px 16px 16px;
  }

  .equity-icon {
    width: 72px;
    height: 72px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 26px;
    font-weight: 700;
    margin-bottom: 16px;
  }

  .share-count {
    font-size: 32px;
    font-weight: 700;
    color: white;
    margin: 0;
  }

  .value-usd {
    font-size: 18px;
    color: #8b949e;
    margin: 6px 0 12px;
  }

  .price-info {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .current-price {
    font-size: 14px;
    color: #8b949e;
  }

  .price-change {
    font-size: 13px;
    font-weight: 600;
    padding: 4px 10px;
    border-radius: 8px;
  }

  .price-change.positive {
    color: #10b981;
    background: rgba(16, 185, 129, 0.12);
  }

  .price-change.negative {
    color: #ef4444;
    background: rgba(239, 68, 68, 0.12);
  }

  .chart-section {
    padding: 0 16px;
    margin-bottom: 20px;
  }

  .mini-chart {
    width: 100%;
    height: 80px;
    border-radius: 12px;
    background: rgba(22, 27, 34, 0.5);
  }

  .action-buttons {
    display: flex;
    justify-content: center;
    gap: 32px;
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
    width: 56px;
    height: 56px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.2s;
  }

  .action-btn:active .action-icon {
    transform: scale(0.92);
  }

  .action-icon.send {
    background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
  }

  .action-icon.receive {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  }

  .action-icon.trade {
    background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%);
  }

  .action-icon svg {
    width: 24px;
    height: 24px;
    color: white;
  }

  .action-btn span {
    font-size: 13px;
    font-weight: 500;
    color: #8b949e;
  }

  .stats-card {
    margin: 0 16px 16px;
    padding: 18px;
    background: rgba(22, 27, 34, 0.8);
    border: 1px solid rgba(48, 54, 61, 0.5);
    border-radius: 16px;
  }

  .card-title {
    font-size: 15px;
    font-weight: 600;
    color: white;
    margin: 0 0 14px;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 14px;
  }

  .stat-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .stat-label {
    font-size: 12px;
    color: #8b949e;
  }

  .stat-value {
    font-size: 16px;
    font-weight: 600;
    color: white;
  }

  .stat-value.profit {
    color: #10b981;
  }

  .stat-value.loss {
    color: #ef4444;
  }

  .stat-value small {
    font-size: 12px;
    opacity: 0.8;
  }

  .address-section {
    margin: 0 16px 20px;
    padding: 14px 16px;
    background: rgba(22, 27, 34, 0.8);
    border: 1px solid rgba(48, 54, 61, 0.5);
    border-radius: 14px;
  }

  .address-label {
    display: block;
    font-size: 12px;
    color: #8b949e;
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
    color: white;
    font-size: 14px;
    font-family: 'SF Mono', monospace;
  }

  .address-value:active {
    opacity: 0.7;
  }

  .address-value svg {
    width: 18px;
    height: 18px;
    color: #8b949e;
  }

  .transactions-section {
    padding: 0 16px;
  }

  .section-title {
    font-size: 15px;
    font-weight: 600;
    color: white;
    margin: 0 0 12px;
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
    background: rgba(22, 27, 34, 0.7);
    border: 1px solid rgba(48, 54, 61, 0.4);
    border-radius: 14px;
  }

  .tx-icon {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .tx-icon.buy {
    background: rgba(16, 185, 129, 0.12);
    color: #10b981;
  }

  .tx-icon.sell {
    background: rgba(239, 68, 68, 0.12);
    color: #ef4444;
  }

  .tx-icon.dividend {
    background: rgba(59, 130, 246, 0.12);
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
  }

  .tx-type {
    font-size: 14px;
    font-weight: 500;
    color: white;
  }

  .tx-date {
    font-size: 12px;
    color: #8b949e;
  }

  .tx-amount {
    text-align: right;
  }

  .tx-amount .positive {
    color: #10b981;
    font-weight: 600;
    font-size: 14px;
  }

  .tx-amount .negative {
    color: #ef4444;
    font-weight: 600;
    font-size: 14px;
  }

  .tx-amount .dividend {
    color: #3B82F6;
    font-weight: 600;
    font-size: 14px;
  }
`;
