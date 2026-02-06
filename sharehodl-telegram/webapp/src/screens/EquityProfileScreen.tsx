/**
 * Equity Profile Screen - Premium equity details view
 */

import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

// Demo equity data
const DEMO_EQUITIES: Record<string, {
  id: string;
  symbol: string;
  name: string;
  shares: number;
  pricePerShare: number;
  change24h: number;
  color: string;
  description: string;
  marketCap: number;
  volume24h: number;
  high52w: number;
  low52w: number;
  sector: string;
  dividendYield: number;
  peRatio: number;
  avgCost: number;
}> = {
  'sharehodl-plc': {
    id: 'sharehodl-plc',
    symbol: 'SHDL',
    name: 'ShareHODL PLC',
    shares: 1250,
    pricePerShare: 12.50,
    change24h: 4.25,
    color: '#3B82F6',
    description: 'Pioneering blockchain technology for democratizing equity markets through tokenization.',
    marketCap: 125000000,
    volume24h: 2500000,
    high52w: 15.80,
    low52w: 8.20,
    sector: 'Technology',
    dividendYield: 2.5,
    peRatio: 18.5,
    avgCost: 10.00
  },
  'property-mainnet': {
    id: 'property-mainnet',
    symbol: 'PROP',
    name: 'Property Mainnet',
    shares: 500,
    pricePerShare: 8.75,
    change24h: -1.30,
    color: '#10B981',
    description: 'Tokenized real estate enabling fractional ownership of premium properties worldwide.',
    marketCap: 87500000,
    volume24h: 1200000,
    high52w: 12.40,
    low52w: 6.50,
    sector: 'Real Estate',
    dividendYield: 4.2,
    peRatio: 12.3,
    avgCost: 7.50
  },
  'tech-ventures': {
    id: 'tech-ventures',
    symbol: 'TVNT',
    name: 'Tech Ventures Ltd',
    shares: 2000,
    pricePerShare: 3.20,
    change24h: 7.80,
    color: '#F59E0B',
    description: 'Early-stage technology startup investments with high-growth potential.',
    marketCap: 32000000,
    volume24h: 450000,
    high52w: 4.80,
    low52w: 1.90,
    sector: 'Venture Capital',
    dividendYield: 0,
    peRatio: 25.0,
    avgCost: 2.40
  },
  'green-energy': {
    id: 'green-energy',
    symbol: 'GREN',
    name: 'Green Energy Corp',
    shares: 750,
    pricePerShare: 15.00,
    change24h: 2.15,
    color: '#059669',
    description: 'Renewable energy projects including solar and wind farm operations.',
    marketCap: 150000000,
    volume24h: 3200000,
    high52w: 18.50,
    low52w: 11.20,
    sector: 'Energy',
    dividendYield: 3.8,
    peRatio: 15.2,
    avgCost: 12.80
  },
  'fintech-global': {
    id: 'fintech-global',
    symbol: 'FNTK',
    name: 'FinTech Global',
    shares: 300,
    pricePerShare: 45.00,
    change24h: -0.50,
    color: '#6366F1',
    description: 'Innovative financial technology solutions for banking and payments.',
    marketCap: 450000000,
    volume24h: 8500000,
    high52w: 52.00,
    low52w: 32.50,
    sector: 'Financial Services',
    dividendYield: 1.5,
    peRatio: 22.0,
    avgCost: 38.00
  }
};

// Simulated price chart data points - more data points for smoother curve
const generateChartData = (basePrice: number, change: number) => {
  const points = [];
  const numPoints = 40;
  const variation = basePrice * 0.05;
  let startPrice = basePrice - (change / 100) * basePrice;

  for (let i = 0; i < numPoints; i++) {
    const progress = i / (numPoints - 1);
    const noise = (Math.random() - 0.5) * variation * 0.6;
    const trend = (change / 100) * basePrice * progress;
    const price = startPrice + trend + noise;
    points.push(Math.max(price, basePrice * 0.5));
  }
  return points;
};

export function EquityProfileScreen() {
  const { equityId } = useParams();
  const navigate = useNavigate();
  const tg = window.Telegram?.WebApp;
  const [selectedPeriod, setSelectedPeriod] = useState('1D');

  const equity = equityId ? DEMO_EQUITIES[equityId] : null;

  if (!equity) {
    return (
      <div className="not-found">
        <p>Equity not found</p>
        <button onClick={() => navigate('/portfolio')}>Go Back</button>
      </div>
    );
  }

  const totalValue = equity.shares * equity.pricePerShare;
  const totalCost = equity.shares * equity.avgCost;
  const profitLoss = totalValue - totalCost;
  const profitLossPercent = ((profitLoss / totalCost) * 100);
  const isPositive = equity.change24h >= 0;
  const isProfitable = profitLoss >= 0;

  const chartData = generateChartData(equity.pricePerShare, equity.change24h);
  const minPrice = Math.min(...chartData);
  const maxPrice = Math.max(...chartData);
  const priceRange = maxPrice - minPrice || 1;

  // Chart dimensions
  const chartWidth = 320;
  const chartHeight = 100;
  const padding = 2;

  // Generate SVG path with proper coordinates
  const chartPath = chartData.map((price, i) => {
    const x = padding + (i / (chartData.length - 1)) * (chartWidth - padding * 2);
    const y = padding + (1 - (price - minPrice) / priceRange) * (chartHeight - padding * 2);
    return `${i === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`;
  }).join(' ');

  // Area path for gradient fill
  const areaPath = `${chartPath} L ${chartWidth - padding} ${chartHeight - padding} L ${padding} ${chartHeight - padding} Z`;

  const handleTrade = (action: 'buy' | 'sell') => {
    tg?.HapticFeedback?.impactOccurred('medium');
    tg?.showAlert(`${action === 'buy' ? 'Buy' : 'Sell'} ${equity.symbol} - Coming soon!`);
  };

  return (
    <div className="equity-profile">
      {/* Header */}
      <header className="header">
        <button className="back-btn" onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); navigate(-1); }}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <path d="M15 18l-6-6 6-6" />
          </svg>
        </button>
        <div className="header-center">
          <span className="header-symbol">{equity.symbol}</span>
        </div>
        <button className="action-btn" onClick={() => tg?.showAlert('Add to watchlist')}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 5v14M5 12h14" />
          </svg>
        </button>
      </header>

      {/* Price Section */}
      <div className="price-section">
        <div className="company-row">
          <div className="company-logo" style={{ background: `${equity.color}20`, color: equity.color }}>
            {equity.symbol.slice(0, 2)}
          </div>
          <div className="company-details">
            <h1 className="company-name">{equity.name}</h1>
            <span className="company-sector">{equity.sector}</span>
          </div>
        </div>

        <div className="price-display">
          <span className="current-price">${equity.pricePerShare.toFixed(2)}</span>
          <div className={`change-badge ${isPositive ? 'up' : 'down'}`}>
            <svg viewBox="0 0 24 24" fill="currentColor" width="14" height="14">
              <path d={isPositive ? "M7 14l5-5 5 5H7z" : "M7 10l5 5 5-5H7z"} />
            </svg>
            <span>{isPositive ? '+' : ''}{equity.change24h.toFixed(2)}%</span>
          </div>
        </div>
      </div>

      {/* Chart */}
      <div className="chart-container">
        <svg viewBox={`0 0 ${chartWidth} ${chartHeight}`} preserveAspectRatio="none" className="chart">
          <defs>
            <linearGradient id={`gradient-${equity.id}`} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor={isPositive ? '#10b981' : '#ef4444'} stopOpacity="0.25" />
              <stop offset="100%" stopColor={isPositive ? '#10b981' : '#ef4444'} stopOpacity="0.02" />
            </linearGradient>
          </defs>
          <path
            d={areaPath}
            fill={`url(#gradient-${equity.id})`}
          />
          <path
            d={chartPath}
            fill="none"
            stroke={isPositive ? '#10b981' : '#ef4444'}
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>

        {/* Period Selector */}
        <div className="period-selector">
          {['1D', '1W', '1M', '3M', '1Y', 'ALL'].map((period) => (
            <button
              key={period}
              className={`period-btn ${selectedPeriod === period ? 'active' : ''}`}
              onClick={() => { tg?.HapticFeedback?.selectionChanged(); setSelectedPeriod(period); }}
            >
              {period}
            </button>
          ))}
        </div>
      </div>

      {/* Your Position */}
      <div className="position-card">
        <div className="position-header">
          <h2>Your Position</h2>
          <span className={`pnl-tag ${isProfitable ? 'profit' : 'loss'}`}>
            {isProfitable ? '+' : ''}{profitLossPercent.toFixed(2)}%
          </span>
        </div>

        <div className="position-grid">
          <div className="position-item main">
            <span className="item-value large">${totalValue.toLocaleString()}</span>
            <span className="item-label">Market Value</span>
          </div>
          <div className="position-item main">
            <span className="item-value large">{equity.shares.toLocaleString()}</span>
            <span className="item-label">Shares</span>
          </div>
        </div>

        <div className="position-details">
          <div className="detail-row">
            <span className="detail-label">Avg Cost</span>
            <span className="detail-value">${equity.avgCost.toFixed(2)}</span>
          </div>
          <div className="detail-row">
            <span className="detail-label">Total Return</span>
            <span className={`detail-value ${isProfitable ? 'profit' : 'loss'}`}>
              {isProfitable ? '+' : ''}${profitLoss.toLocaleString()}
            </span>
          </div>
        </div>
      </div>

      {/* Key Stats */}
      <div className="stats-card">
        <h2>Key Statistics</h2>
        <div className="stats-grid">
          <div className="stat">
            <span className="stat-label">Market Cap</span>
            <span className="stat-value">${(equity.marketCap / 1000000).toFixed(0)}M</span>
          </div>
          <div className="stat">
            <span className="stat-label">Volume</span>
            <span className="stat-value">${(equity.volume24h / 1000000).toFixed(1)}M</span>
          </div>
          <div className="stat">
            <span className="stat-label">52W High</span>
            <span className="stat-value">${equity.high52w.toFixed(2)}</span>
          </div>
          <div className="stat">
            <span className="stat-label">52W Low</span>
            <span className="stat-value">${equity.low52w.toFixed(2)}</span>
          </div>
          <div className="stat">
            <span className="stat-label">P/E Ratio</span>
            <span className="stat-value">{equity.peRatio.toFixed(1)}</span>
          </div>
          <div className="stat">
            <span className="stat-label">Dividend</span>
            <span className="stat-value">{equity.dividendYield > 0 ? `${equity.dividendYield}%` : '-'}</span>
          </div>
        </div>
      </div>

      {/* About */}
      <div className="about-card">
        <h2>About</h2>
        <p>{equity.description}</p>
      </div>

      {/* Trade Buttons - Fixed at bottom */}
      <div className="trade-bar">
        <button className="trade-btn sell" onClick={() => handleTrade('sell')}>
          Sell
        </button>
        <button className="trade-btn buy" onClick={() => handleTrade('buy')}>
          Buy
        </button>
      </div>

      <style>{`
        .equity-profile {
          min-height: 100vh;
          padding-bottom: 100px;
          background: #0D1117;
        }

        .not-found {
          min-height: 100vh;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 16px;
          color: white;
        }

        .not-found button {
          padding: 12px 24px;
          background: #3B82F6;
          border: none;
          border-radius: 10px;
          color: white;
          font-weight: 600;
        }

        /* Header */
        .header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 12px 16px;
          position: sticky;
          top: 0;
          background: rgba(13, 17, 23, 0.9);
          backdrop-filter: blur(20px);
          -webkit-backdrop-filter: blur(20px);
          z-index: 10;
        }

        .back-btn, .action-btn {
          width: 40px;
          height: 40px;
          border-radius: 12px;
          background: rgba(48, 54, 61, 0.6);
          border: none;
          display: flex;
          align-items: center;
          justify-content: center;
          color: white;
          cursor: pointer;
        }

        .back-btn svg, .action-btn svg {
          width: 22px;
          height: 22px;
        }

        .header-center {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .header-symbol {
          font-size: 17px;
          font-weight: 700;
          color: white;
        }

        /* Price Section */
        .price-section {
          padding: 16px 20px 8px;
        }

        .company-row {
          display: flex;
          align-items: center;
          gap: 14px;
          margin-bottom: 16px;
        }

        .company-logo {
          width: 52px;
          height: 52px;
          border-radius: 14px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 18px;
          font-weight: 700;
        }

        .company-details {
          flex: 1;
        }

        .company-name {
          font-size: 18px;
          font-weight: 600;
          color: white;
          margin: 0 0 4px;
        }

        .company-sector {
          font-size: 13px;
          color: #8b949e;
        }

        .price-display {
          display: flex;
          align-items: baseline;
          gap: 12px;
        }

        .current-price {
          font-size: 36px;
          font-weight: 700;
          color: white;
          letter-spacing: -1px;
        }

        .change-badge {
          display: inline-flex;
          align-items: center;
          gap: 4px;
          padding: 6px 10px;
          border-radius: 8px;
          font-size: 14px;
          font-weight: 600;
        }

        .change-badge.up {
          background: rgba(16, 185, 129, 0.15);
          color: #10b981;
        }

        .change-badge.down {
          background: rgba(239, 68, 68, 0.15);
          color: #ef4444;
        }

        .change-badge svg {
          width: 16px;
          height: 16px;
        }

        /* Chart */
        .chart-container {
          padding: 0 16px;
          margin-bottom: 24px;
        }

        .chart {
          width: 100%;
          height: 140px;
          background: rgba(22, 27, 34, 0.4);
          border-radius: 16px;
        }

        .period-selector {
          display: flex;
          gap: 4px;
          margin-top: 16px;
          padding: 4px;
          background: rgba(48, 54, 61, 0.4);
          border-radius: 10px;
        }

        .period-btn {
          flex: 1;
          padding: 8px 0;
          background: transparent;
          border: none;
          border-radius: 8px;
          font-size: 13px;
          font-weight: 600;
          color: #8b949e;
          cursor: pointer;
          transition: all 0.2s;
        }

        .period-btn.active {
          background: rgba(59, 130, 246, 0.2);
          color: #3B82F6;
        }

        /* Position Card */
        .position-card {
          margin: 0 16px 16px;
          padding: 20px;
          background: rgba(22, 27, 34, 0.8);
          border: 1px solid rgba(48, 54, 61, 0.6);
          border-radius: 16px;
        }

        .position-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
        }

        .position-header h2 {
          font-size: 15px;
          font-weight: 600;
          color: white;
          margin: 0;
        }

        .pnl-tag {
          padding: 4px 10px;
          border-radius: 6px;
          font-size: 13px;
          font-weight: 600;
        }

        .pnl-tag.profit {
          background: rgba(16, 185, 129, 0.15);
          color: #10b981;
        }

        .pnl-tag.loss {
          background: rgba(239, 68, 68, 0.15);
          color: #ef4444;
        }

        .position-grid {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 16px;
          margin-bottom: 16px;
        }

        .position-item {
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .item-value {
          font-size: 15px;
          font-weight: 600;
          color: white;
        }

        .item-value.large {
          font-size: 22px;
          font-weight: 700;
        }

        .item-label {
          font-size: 12px;
          color: #8b949e;
        }

        .position-details {
          padding-top: 16px;
          border-top: 1px solid rgba(48, 54, 61, 0.5);
          display: flex;
          flex-direction: column;
          gap: 10px;
        }

        .detail-row {
          display: flex;
          justify-content: space-between;
        }

        .detail-label {
          font-size: 14px;
          color: #8b949e;
        }

        .detail-value {
          font-size: 14px;
          font-weight: 600;
          color: white;
        }

        .detail-value.profit {
          color: #10b981;
        }

        .detail-value.loss {
          color: #ef4444;
        }

        /* Stats Card */
        .stats-card {
          margin: 0 16px 16px;
          padding: 20px;
          background: rgba(22, 27, 34, 0.8);
          border: 1px solid rgba(48, 54, 61, 0.6);
          border-radius: 16px;
        }

        .stats-card h2 {
          font-size: 15px;
          font-weight: 600;
          color: white;
          margin: 0 0 16px;
        }

        .stats-grid {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 16px;
        }

        .stat {
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .stat-label {
          font-size: 11px;
          color: #8b949e;
          text-transform: uppercase;
          letter-spacing: 0.3px;
        }

        .stat-value {
          font-size: 14px;
          font-weight: 600;
          color: white;
        }

        /* About Card */
        .about-card {
          margin: 0 16px 16px;
          padding: 20px;
          background: rgba(22, 27, 34, 0.8);
          border: 1px solid rgba(48, 54, 61, 0.6);
          border-radius: 16px;
        }

        .about-card h2 {
          font-size: 15px;
          font-weight: 600;
          color: white;
          margin: 0 0 12px;
        }

        .about-card p {
          font-size: 14px;
          color: #8b949e;
          line-height: 1.6;
          margin: 0;
        }

        /* Trade Bar */
        .trade-bar {
          position: fixed;
          bottom: 0;
          left: 0;
          right: 0;
          display: flex;
          gap: 12px;
          padding: 16px 20px;
          padding-bottom: calc(16px + env(safe-area-inset-bottom, 0));
          background: rgba(13, 17, 23, 0.95);
          backdrop-filter: blur(20px);
          -webkit-backdrop-filter: blur(20px);
          border-top: 1px solid rgba(48, 54, 61, 0.5);
        }

        .trade-btn {
          flex: 1;
          padding: 16px;
          border: none;
          border-radius: 12px;
          font-size: 16px;
          font-weight: 700;
          cursor: pointer;
          transition: transform 0.15s, opacity 0.15s;
        }

        .trade-btn:active {
          transform: scale(0.97);
          opacity: 0.9;
        }

        .trade-btn.buy {
          background: linear-gradient(135deg, #10b981 0%, #059669 100%);
          color: white;
        }

        .trade-btn.sell {
          background: rgba(239, 68, 68, 0.15);
          color: #ef4444;
          border: 1px solid rgba(239, 68, 68, 0.3);
        }
      `}</style>
    </div>
  );
}
