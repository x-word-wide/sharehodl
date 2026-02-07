/**
 * Lending Screen - DeFi lending protocol
 * Shows Coming Soon state until lending module is live
 */

import { useState, useEffect } from 'react';
import { Loader2 } from 'lucide-react';

const API_BASE = import.meta.env.VITE_SHAREHODL_REST || 'https://api.sharehodl.com';

interface LendingMarket {
  asset: string;
  totalSupply: number;
  totalBorrow: number;
  supplyApy: number;
  borrowApr: number;
  utilization: number;
}


export function LendingScreen() {
  const [isLoading, setIsLoading] = useState(true);
  const [markets, setMarkets] = useState<LendingMarket[]>([]);
  const [isLive, setIsLive] = useState(false);

  useEffect(() => {
    const checkLendingStatus = async () => {
      try {
        // Try to fetch lending markets from blockchain
        const response = await fetch(`${API_BASE}/sharehodl/lending/v1/markets`);
        if (response.ok) {
          const data = await response.json();
          if (data.markets && data.markets.length > 0) {
            setMarkets(data.markets);
            setIsLive(true);
          }
        }
      } catch {
        // Lending module not live yet
        setIsLive(false);
      } finally {
        setIsLoading(false);
      }
    };

    checkLendingStatus();
  }, []);

  if (isLoading) {
    return (
      <div className="lending-screen">
        <div className="loading-state">
          <Loader2 className="spin" size={40} />
          <p>Loading lending markets...</p>
        </div>
        <style>{styles}</style>
      </div>
    );
  }

  // Coming Soon State
  if (!isLive) {
    return (
      <div className="lending-screen">
        <div className="coming-soon">
          <div className="coming-soon-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 2L2 7l10 5 10-5-10-5z" />
              <path d="M2 17l10 5 10-5" />
              <path d="M2 12l10 5 10-5" />
            </svg>
          </div>
          <h1 className="coming-soon-title">Lending</h1>
          <p className="coming-soon-subtitle">Coming Soon</p>
          <div className="coming-soon-desc">
            <p>The ShareHODL DeFi lending protocol is under development.</p>
            <p className="features-title">Upcoming Features:</p>
            <ul className="features-list">
              <li>Supply assets and earn yield</li>
              <li>Borrow against your collateral</li>
              <li>Competitive APY rates</li>
              <li>Multi-asset collateral support</li>
              <li>Flash loan capabilities</li>
            </ul>
          </div>
          <div className="status-badge">
            <span className="status-dot" />
            <span>Development in Progress</span>
          </div>
        </div>
        <style>{styles}</style>
      </div>
    );
  }

  // Live State - Show real data
  return (
    <div className="lending-screen">
      <div className="lending-header">
        <h1 className="lending-title">Lending</h1>
      </div>

      {/* Markets List */}
      <div className="markets-section">
        <h2 className="section-title">Markets</h2>
        {markets.length === 0 ? (
          <div className="empty-state">
            <p>No lending markets available yet</p>
          </div>
        ) : (
          <div className="markets-list">
            {markets.map((market) => (
              <div key={market.asset} className="market-card">
                <div className="market-asset">
                  <span className="asset-name">{market.asset}</span>
                </div>
                <div className="market-stats">
                  <div className="stat">
                    <span className="label">Supply APY</span>
                    <span className="value green">{market.supplyApy.toFixed(2)}%</span>
                  </div>
                  <div className="stat">
                    <span className="label">Borrow APR</span>
                    <span className="value">{market.borrowApr.toFixed(2)}%</span>
                  </div>
                  <div className="stat">
                    <span className="label">Utilization</span>
                    <span className="value">{market.utilization.toFixed(0)}%</span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <style>{styles}</style>
    </div>
  );
}

const styles = `
  .lending-screen {
    min-height: 100vh;
    padding: 16px;
    padding-bottom: 100px;
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

  .coming-soon {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    min-height: 70vh;
    padding: 20px;
  }

  .coming-soon-icon {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(139, 92, 246, 0.2));
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 24px;
  }

  .coming-soon-icon svg {
    width: 40px;
    height: 40px;
    color: #3B82F6;
  }

  .coming-soon-title {
    font-size: 28px;
    font-weight: 700;
    color: white;
    margin: 0 0 8px;
  }

  .coming-soon-subtitle {
    font-size: 18px;
    color: #8b949e;
    margin: 0 0 24px;
  }

  .coming-soon-desc {
    max-width: 320px;
    color: #8b949e;
    font-size: 14px;
    line-height: 1.6;
  }

  .coming-soon-desc p {
    margin: 0 0 16px;
  }

  .features-title {
    color: white;
    font-weight: 600;
    margin-bottom: 8px !important;
  }

  .features-list {
    list-style: none;
    padding: 0;
    margin: 0;
    text-align: left;
  }

  .features-list li {
    padding: 8px 0;
    padding-left: 24px;
    position: relative;
  }

  .features-list li::before {
    content: '✓';
    position: absolute;
    left: 0;
    color: #10b981;
  }

  .status-badge {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 24px;
    padding: 8px 16px;
    background: rgba(245, 158, 11, 0.1);
    border-radius: 20px;
    font-size: 13px;
    color: #f59e0b;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #f59e0b;
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }

  .lending-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
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
    padding: 8px 12px;
    background: #161B22;
    border-radius: 12px;
  }

  .health-label {
    font-size: 12px;
    color: #8b949e;
  }

  .health-value {
    font-size: 16px;
    font-weight: 700;
  }

  .health-value.good { color: #10b981; }
  .health-value.warning { color: #f59e0b; }
  .health-value.danger { color: #ef4444; }

  .section-title {
    font-size: 16px;
    font-weight: 600;
    color: white;
    margin: 0 0 12px;
  }

  .markets-section {
    margin-bottom: 24px;
  }

  .empty-state {
    text-align: center;
    padding: 40px 20px;
    color: #8b949e;
    background: #161B22;
    border-radius: 14px;
  }

  .markets-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .market-card {
    background: #161B22;
    border-radius: 14px;
    padding: 16px;
  }

  .market-asset {
    margin-bottom: 12px;
  }

  .asset-name {
    font-size: 18px;
    font-weight: 600;
    color: white;
  }

  .market-stats {
    display: flex;
    justify-content: space-between;
  }

  .stat {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .stat .label {
    font-size: 12px;
    color: #8b949e;
  }

  .stat .value {
    font-size: 14px;
    font-weight: 600;
    color: white;
  }

  .stat .value.green {
    color: #10b981;
  }
`;
