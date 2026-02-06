/**
 * Lending Screen - Modern DeFi lending with glassmorphism
 */

import { useState } from 'react';

// SVG Icons for assets
const AssetIcons: Record<string, React.ReactNode> = {
  HODL: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M12 2L2 7l10 5 10-5-10-5z" />
      <path d="M2 17l10 5 10-5" />
      <path d="M2 12l10 5 10-5" />
    </svg>
  ),
  ETH: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 1.75l-6.25 10.5L12 16l6.25-3.75L12 1.75zM5.75 13.5L12 22.25l6.25-8.75L12 17.25 5.75 13.5z" />
    </svg>
  ),
  USDT: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 .67 3 1.5S13.66 8 12 8s-3-.67-3-1.5S10.34 5 12 5zm4 10H8v-2h3V9.5h2V13h3v2z" />
    </svg>
  ),
  USDC: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <circle cx="12" cy="12" r="10" />
      <text x="12" y="16" textAnchor="middle" fill="#0D1117" fontSize="10" fontWeight="bold">$</text>
    </svg>
  ),
  BTC: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-1h-1v-2h1v-4h-1V8h1V7h2v1h1v2h-1v4h1v2h-1v1z" />
    </svg>
  ),
};

// Demo lending markets
const LENDING_MARKETS = [
  { asset: 'HODL', totalSupply: 5000000, totalBorrow: 2500000, supplyApy: 5.2, borrowApr: 8.5, utilization: 50, color: '#3B82F6' },
  { asset: 'ETH', totalSupply: 1200, totalBorrow: 800, supplyApy: 3.8, borrowApr: 6.2, utilization: 67, color: '#627EEA' },
  { asset: 'USDT', totalSupply: 2500000, totalBorrow: 1800000, supplyApy: 8.5, borrowApr: 12.0, utilization: 72, color: '#26A17B' },
  { asset: 'USDC', totalSupply: 3000000, totalBorrow: 2100000, supplyApy: 7.2, borrowApr: 10.5, utilization: 70, color: '#2775CA' },
];

// Demo user positions
const USER_POSITIONS = {
  supplied: [
    { asset: 'HODL', amount: 5000, value: 5000, apy: 5.2 },
    { asset: 'ETH', amount: 1.5, value: 5175, apy: 3.8 },
  ],
  borrowed: [
    { asset: 'USDT', amount: 2000, value: 2000, apr: 12.0 },
  ]
};

export function LendingScreen() {
  const [activeTab, setActiveTab] = useState<'supply' | 'borrow'>('supply');
  const [showDialog, setShowDialog] = useState(false);
  const [selectedMarket, setSelectedMarket] = useState<typeof LENDING_MARKETS[0] | null>(null);
  const tg = window.Telegram?.WebApp;

  // Calculate totals
  const totalSupplied = USER_POSITIONS.supplied.reduce((sum, p) => sum + p.value, 0);
  const totalBorrowed = USER_POSITIONS.borrowed.reduce((sum, p) => sum + p.value, 0);
  const healthFactor = totalBorrowed > 0 ? ((totalSupplied * 0.8) / totalBorrowed).toFixed(2) : '∞';
  const netApy = 4.2;

  const handleMarketClick = (market: typeof LENDING_MARKETS[0]) => {
    tg?.HapticFeedback?.impactOccurred('medium');
    setSelectedMarket(market);
    setShowDialog(true);
  };

  return (
    <div className="lending-screen">
      {/* Header */}
      <div className="lending-header">
        <h1 className="lending-title">Lending</h1>
        <div className="health-badge">
          <span className="health-label">Health</span>
          <span className={`health-value ${parseFloat(healthFactor) > 1.5 ? 'good' : parseFloat(healthFactor) > 1.2 ? 'warning' : 'danger'}`}>
            {healthFactor}
          </span>
        </div>
      </div>

      {/* Overview Cards */}
      <div className="overview-cards">
        <div className="overview-card supplied">
          <div className="card-bg" />
          <div className="card-content">
            <div className="card-icon supply-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
                <polyline points="17 6 23 6 23 12" />
              </svg>
            </div>
            <div className="card-info">
              <span className="card-label">Supplied</span>
              <span className="card-value">${totalSupplied.toLocaleString()}</span>
              <span className="card-apy">+{netApy}% Net APY</span>
            </div>
          </div>
        </div>
        <div className="overview-card borrowed">
          <div className="card-bg" />
          <div className="card-content">
            <div className="card-icon borrow-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <polyline points="23 18 13.5 8.5 8.5 13.5 1 6" />
                <polyline points="17 18 23 18 23 12" />
              </svg>
            </div>
            <div className="card-info">
              <span className="card-label">Borrowed</span>
              <span className="card-value">${totalBorrowed.toLocaleString()}</span>
              <span className="card-apr">-12.0% APR</span>
            </div>
          </div>
        </div>
      </div>

      {/* Your Positions */}
      {(USER_POSITIONS.supplied.length > 0 || USER_POSITIONS.borrowed.length > 0) && (
        <div className="positions-section">
          <h2 className="section-title">Your Positions</h2>
          <div className="positions-grid">
            {USER_POSITIONS.supplied.map((pos) => (
              <div key={`supply-${pos.asset}`} className="position-card supply">
                <div className="position-header">
                  <span className="position-type">Supplying</span>
                  <span className="position-apy">+{pos.apy}%</span>
                </div>
                <div className="position-amount">
                  {pos.amount.toLocaleString()} {pos.asset}
                </div>
                <div className="position-value">${pos.value.toLocaleString()}</div>
              </div>
            ))}
            {USER_POSITIONS.borrowed.map((pos) => (
              <div key={`borrow-${pos.asset}`} className="position-card borrow">
                <div className="position-header">
                  <span className="position-type">Borrowing</span>
                  <span className="position-apr">-{pos.apr}%</span>
                </div>
                <div className="position-amount">
                  {pos.amount.toLocaleString()} {pos.asset}
                </div>
                <div className="position-value">${pos.value.toLocaleString()}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Tab selector */}
      <div className="tab-container">
        <div className="tabs">
          <button
            className={`tab ${activeTab === 'supply' ? 'active' : ''}`}
            onClick={() => {
              tg?.HapticFeedback?.selectionChanged();
              setActiveTab('supply');
            }}
          >
            <span className="tab-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="12" cy="12" r="10" />
                <line x1="12" y1="8" x2="12" y2="16" />
                <line x1="8" y1="12" x2="16" y2="12" />
              </svg>
            </span>
            Supply
          </button>
          <button
            className={`tab ${activeTab === 'borrow' ? 'active' : ''}`}
            onClick={() => {
              tg?.HapticFeedback?.selectionChanged();
              setActiveTab('borrow');
            }}
          >
            <span className="tab-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <rect x="3" y="3" width="18" height="18" rx="2" />
                <line x1="3" y1="9" x2="21" y2="9" />
                <line x1="9" y1="21" x2="9" y2="9" />
              </svg>
            </span>
            Borrow
          </button>
        </div>
      </div>

      {/* Markets */}
      <div className="markets-section">
        <h2 className="section-title">
          {activeTab === 'supply' ? 'Supply Markets' : 'Borrow Markets'}
        </h2>
        <div className="markets-list">
          {LENDING_MARKETS.map((market) => (
            <MarketCard
              key={market.asset}
              market={market}
              isSupply={activeTab === 'supply'}
              onClick={() => handleMarketClick(market)}
            />
          ))}
        </div>
      </div>

      {/* Supply/Borrow Dialog */}
      {showDialog && selectedMarket && (
        <ActionDialog
          market={selectedMarket}
          isSupply={activeTab === 'supply'}
          onClose={() => setShowDialog(false)}
        />
      )}

      <style>{`
        .lending-screen {
          min-height: 100vh;
          padding: 16px;
          padding-bottom: 100px;
        }

        .lending-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 20px;
        }

        .lending-title {
          font-size: 24px;
          font-weight: 700;
          color: white;
          margin: 0;
        }

        .health-badge {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 8px 14px;
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(10px);
          border: 1px solid rgba(48, 54, 61, 0.5);
          border-radius: 20px;
        }

        .health-label {
          font-size: 12px;
          color: #8b949e;
        }

        .health-value {
          font-size: 14px;
          font-weight: 600;
        }

        .health-value.good { color: #10b981; }
        .health-value.warning { color: #f59e0b; }
        .health-value.danger { color: #ef4444; }

        /* Overview Cards */
        .overview-cards {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 12px;
          margin-bottom: 24px;
        }

        .overview-card {
          position: relative;
          padding: 18px;
          border-radius: 16px;
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(12px);
          border: 1px solid rgba(48, 54, 61, 0.5);
          overflow: hidden;
        }

        .card-bg {
          position: absolute;
          inset: 0;
          opacity: 0.1;
        }

        .overview-card.supplied .card-bg {
          background: radial-gradient(circle at 80% 20%, #10b981, transparent 60%);
        }

        .overview-card.borrowed .card-bg {
          background: radial-gradient(circle at 80% 20%, #f59e0b, transparent 60%);
        }

        .card-content {
          position: relative;
          display: flex;
          gap: 12px;
        }

        .card-icon {
          width: 32px;
          height: 32px;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .card-icon svg {
          width: 24px;
          height: 24px;
        }

        .card-icon.supply-icon {
          color: #10b981;
        }

        .card-icon.borrow-icon {
          color: #f59e0b;
        }

        .card-info {
          display: flex;
          flex-direction: column;
          gap: 2px;
        }

        .card-label {
          font-size: 12px;
          color: #8b949e;
        }

        .card-value {
          font-size: 20px;
          font-weight: 700;
          color: white;
        }

        .card-apy {
          font-size: 12px;
          color: #10b981;
          font-weight: 500;
        }

        .card-apr {
          font-size: 12px;
          color: #f59e0b;
          font-weight: 500;
        }

        /* Positions */
        .positions-section {
          margin-bottom: 24px;
        }

        .section-title {
          font-size: 14px;
          font-weight: 600;
          color: #8b949e;
          margin: 0 0 12px 4px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .positions-grid {
          display: flex;
          gap: 10px;
          overflow-x: auto;
          padding-bottom: 8px;
          -webkit-overflow-scrolling: touch;
        }

        .position-card {
          flex-shrink: 0;
          width: 140px;
          padding: 14px;
          border-radius: 14px;
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(10px);
          border: 1px solid rgba(48, 54, 61, 0.5);
        }

        .position-card.supply {
          border-left: 3px solid #10b981;
        }

        .position-card.borrow {
          border-left: 3px solid #f59e0b;
        }

        .position-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 8px;
        }

        .position-type {
          font-size: 11px;
          color: #8b949e;
        }

        .position-apy {
          font-size: 11px;
          font-weight: 600;
          color: #10b981;
        }

        .position-apr {
          font-size: 11px;
          font-weight: 600;
          color: #f59e0b;
        }

        .position-amount {
          font-size: 15px;
          font-weight: 600;
          color: white;
          margin-bottom: 2px;
        }

        .position-value {
          font-size: 13px;
          color: #8b949e;
        }

        /* Tabs */
        .tab-container {
          margin-bottom: 20px;
        }

        .tabs {
          display: flex;
          gap: 6px;
          padding: 5px;
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(10px);
          border: 1px solid rgba(48, 54, 61, 0.4);
          border-radius: 14px;
        }

        .tab {
          flex: 1;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          padding: 12px;
          border: none;
          border-radius: 10px;
          font-size: 14px;
          font-weight: 600;
          color: #8b949e;
          background: transparent;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .tab.active {
          background: linear-gradient(135deg, rgba(30, 64, 175, 0.8) 0%, rgba(59, 130, 246, 0.8) 100%);
          color: white;
          box-shadow: 0 2px 10px rgba(30, 64, 175, 0.3);
        }

        .tab-icon {
          width: 18px;
          height: 18px;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .tab-icon svg {
          width: 18px;
          height: 18px;
        }

        /* Markets */
        .markets-section {
          margin-bottom: 24px;
        }

        .markets-list {
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .market-card {
          padding: 16px;
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(12px);
          border: 1px solid rgba(48, 54, 61, 0.5);
          border-radius: 16px;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .market-card:active {
          transform: scale(0.98);
          background: rgba(22, 27, 34, 0.8);
        }

        .market-header {
          display: flex;
          justify-content: space-between;
          align-items: flex-start;
          margin-bottom: 12px;
        }

        .market-asset {
          display: flex;
          align-items: center;
          gap: 12px;
        }

        .asset-icon {
          width: 44px;
          height: 44px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          position: relative;
        }

        .asset-icon svg {
          width: 24px;
          height: 24px;
        }

        .asset-icon::after {
          content: '';
          position: absolute;
          inset: -4px;
          border-radius: 50%;
          background: inherit;
          opacity: 0.2;
          filter: blur(8px);
          z-index: -1;
        }

        .asset-info {
          display: flex;
          flex-direction: column;
        }

        .asset-name {
          font-size: 16px;
          font-weight: 600;
          color: white;
        }

        .asset-supply {
          font-size: 13px;
          color: #8b949e;
        }

        .market-rate {
          text-align: right;
        }

        .rate-value {
          font-size: 20px;
          font-weight: 700;
        }

        .rate-value.supply { color: #10b981; }
        .rate-value.borrow { color: #f59e0b; }

        .rate-label {
          font-size: 11px;
          color: #8b949e;
          text-transform: uppercase;
        }

        /* Utilization bar */
        .utilization-bar {
          margin-bottom: 14px;
        }

        .util-header {
          display: flex;
          justify-content: space-between;
          margin-bottom: 6px;
        }

        .util-label {
          font-size: 12px;
          color: #8b949e;
        }

        .util-value {
          font-size: 12px;
          color: #8b949e;
        }

        .util-track {
          height: 6px;
          background: rgba(48, 54, 61, 0.5);
          border-radius: 3px;
          overflow: hidden;
        }

        .util-fill {
          height: 100%;
          background: linear-gradient(90deg, #1E40AF, #3B82F6);
          border-radius: 3px;
          transition: width 0.3s ease;
        }

        .market-action {
          width: 100%;
          padding: 12px;
          border: none;
          border-radius: 10px;
          font-size: 14px;
          font-weight: 600;
          color: white;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .market-action.supply {
          background: linear-gradient(135deg, rgba(16, 185, 129, 0.8), rgba(5, 150, 105, 0.8));
        }

        .market-action.borrow {
          background: linear-gradient(135deg, rgba(245, 158, 11, 0.8), rgba(217, 119, 6, 0.8));
        }

        .market-action:active {
          transform: scale(0.97);
        }

        /* Dialog */
        .dialog-overlay {
          position: fixed;
          inset: 0;
          background: rgba(0, 0, 0, 0.7);
          z-index: 100;
          display: flex;
          align-items: flex-end;
          animation: fadeIn 0.2s ease;
        }

        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }

        .dialog-content {
          width: 100%;
          padding: 24px;
          background: #161B22;
          border-radius: 24px 24px 0 0;
          animation: slideUp 0.3s ease;
        }

        @keyframes slideUp {
          from { transform: translateY(100%); }
          to { transform: translateY(0); }
        }

        .dialog-header {
          display: flex;
          align-items: center;
          gap: 12px;
          margin-bottom: 24px;
        }

        .dialog-icon {
          width: 48px;
          height: 48px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .dialog-icon svg {
          width: 26px;
          height: 26px;
        }

        .dialog-title {
          flex: 1;
        }

        .dialog-title h3 {
          font-size: 18px;
          font-weight: 600;
          color: white;
          margin: 0 0 4px;
        }

        .dialog-title p {
          font-size: 13px;
          color: #8b949e;
          margin: 0;
        }

        .dialog-close {
          width: 32px;
          height: 32px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: rgba(48, 54, 61, 0.5);
          border: none;
          border-radius: 50%;
          cursor: pointer;
          color: #8b949e;
        }

        .dialog-input {
          margin-bottom: 20px;
        }

        .input-label {
          display: block;
          font-size: 13px;
          color: #8b949e;
          margin-bottom: 8px;
        }

        .input-field {
          width: 100%;
          padding: 16px;
          background: #0D1117;
          border: 1px solid #30363d;
          border-radius: 12px;
          font-size: 18px;
          color: white;
          outline: none;
        }

        .input-field:focus {
          border-color: #3B82F6;
        }

        .dialog-info {
          padding: 16px;
          background: rgba(48, 54, 61, 0.3);
          border-radius: 12px;
          margin-bottom: 20px;
        }

        .info-row {
          display: flex;
          justify-content: space-between;
          margin-bottom: 8px;
        }

        .info-row:last-child {
          margin-bottom: 0;
        }

        .info-label {
          font-size: 13px;
          color: #8b949e;
        }

        .info-value {
          font-size: 13px;
          font-weight: 500;
          color: white;
        }

        .info-value.highlight {
          color: #10b981;
        }

        .dialog-buttons {
          display: flex;
          gap: 12px;
        }

        .dialog-btn {
          flex: 1;
          padding: 16px;
          border: none;
          border-radius: 12px;
          font-size: 16px;
          font-weight: 600;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .dialog-btn.cancel {
          background: rgba(48, 54, 61, 0.5);
          color: white;
        }

        .dialog-btn.confirm {
          color: white;
        }

        .dialog-btn.confirm.supply {
          background: linear-gradient(135deg, #10b981, #059669);
        }

        .dialog-btn.confirm.borrow {
          background: linear-gradient(135deg, #f59e0b, #d97706);
        }

        .dialog-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .dialog-btn:active:not(:disabled) {
          transform: scale(0.97);
        }
      `}</style>
    </div>
  );
}

function MarketCard({
  market,
  isSupply,
  onClick
}: {
  market: typeof LENDING_MARKETS[0];
  isSupply: boolean;
  onClick: () => void;
}) {
  const rate = isSupply ? market.supplyApy : market.borrowApr;
  const total = isSupply ? market.totalSupply : market.totalBorrow;

  return (
    <div className="market-card" onClick={onClick}>
      <div className="market-header">
        <div className="market-asset">
          <div className="asset-icon" style={{ background: `${market.color}20`, color: market.color }}>
            {AssetIcons[market.asset] || AssetIcons.HODL}
          </div>
          <div className="asset-info">
            <span className="asset-name">{market.asset}</span>
            <span className="asset-supply">${(total / 1000000).toFixed(2)}M {isSupply ? 'supplied' : 'borrowed'}</span>
          </div>
        </div>
        <div className="market-rate">
          <span className={`rate-value ${isSupply ? 'supply' : 'borrow'}`}>{rate}%</span>
          <p className="rate-label">{isSupply ? 'APY' : 'APR'}</p>
        </div>
      </div>

      <div className="utilization-bar">
        <div className="util-header">
          <span className="util-label">Utilization</span>
          <span className="util-value">{market.utilization}%</span>
        </div>
        <div className="util-track">
          <div className="util-fill" style={{ width: `${market.utilization}%` }} />
        </div>
      </div>

      <button className={`market-action ${isSupply ? 'supply' : 'borrow'}`}>
        {isSupply ? 'Supply' : 'Borrow'} {market.asset}
      </button>
    </div>
  );
}

function ActionDialog({
  market,
  isSupply,
  onClose
}: {
  market: typeof LENDING_MARKETS[0];
  isSupply: boolean;
  onClose: () => void;
}) {
  const [amount, setAmount] = useState('');
  const tg = window.Telegram?.WebApp;

  const handleSubmit = () => {
    if (!amount) return;
    tg?.HapticFeedback?.notificationOccurred('success');
    tg?.showAlert(`Successfully ${isSupply ? 'supplied' : 'borrowed'} ${amount} ${market.asset}`);
    onClose();
  };

  return (
    <div className="dialog-overlay" onClick={onClose}>
      <div className="dialog-content" onClick={(e) => e.stopPropagation()}>
        <div className="dialog-header">
          <div className="dialog-icon" style={{ background: `${market.color}20`, color: market.color }}>
            {AssetIcons[market.asset] || AssetIcons.HODL}
          </div>
          <div className="dialog-title">
            <h3>{isSupply ? 'Supply' : 'Borrow'} {market.asset}</h3>
            <p>{isSupply ? 'Earn interest on your assets' : 'Borrow against your collateral'}</p>
          </div>
          <button className="dialog-close" onClick={onClose}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        <div className="dialog-input">
          <label className="input-label">Amount</label>
          <input
            type="number"
            className="input-field"
            placeholder="0.00"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
          />
        </div>

        <div className="dialog-info">
          <div className="info-row">
            <span className="info-label">{isSupply ? 'APY' : 'APR'}</span>
            <span className={`info-value ${isSupply ? 'highlight' : ''}`} style={{ color: isSupply ? '#10b981' : '#f59e0b' }}>
              {isSupply ? '+' : '-'}{isSupply ? market.supplyApy : market.borrowApr}%
            </span>
          </div>
          <div className="info-row">
            <span className="info-label">Utilization</span>
            <span className="info-value">{market.utilization}%</span>
          </div>
          {amount && (
            <div className="info-row">
              <span className="info-label">{isSupply ? 'Daily earnings' : 'Daily cost'}</span>
              <span className="info-value" style={{ color: isSupply ? '#10b981' : '#f59e0b' }}>
                {isSupply ? '+' : '-'}${((parseFloat(amount) * (isSupply ? market.supplyApy : market.borrowApr) / 100) / 365).toFixed(4)}
              </span>
            </div>
          )}
        </div>

        <div className="dialog-buttons">
          <button className="dialog-btn cancel" onClick={onClose}>
            Cancel
          </button>
          <button
            className={`dialog-btn confirm ${isSupply ? 'supply' : 'borrow'}`}
            onClick={handleSubmit}
            disabled={!amount}
          >
            {isSupply ? 'Supply' : 'Borrow'}
          </button>
        </div>
      </div>
    </div>
  );
}
