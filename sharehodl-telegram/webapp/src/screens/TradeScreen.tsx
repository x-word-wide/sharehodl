/**
 * Trade Screen - DEX Trading
 * Shows Coming Soon state until DEX module is live
 */

import { useState, useEffect } from 'react';
import { Loader2 } from 'lucide-react';

const API_BASE = import.meta.env.VITE_SHAREHODL_REST || 'https://api.sharehodl.com';

interface TradingPair {
  id: string;
  baseDenom: string;
  quoteDenom: string;
  baseSymbol: string;
  quoteSymbol: string;
  lastPrice: string;
  volume24h: string;
  priceChange24h: string;
}


export function TradeScreen() {
  const [isLoading, setIsLoading] = useState(true);
  const [pairs, setPairs] = useState<TradingPair[]>([]);
  const [isLive, setIsLive] = useState(false);

  useEffect(() => {
    const checkDexStatus = async () => {
      try {
        // Try to fetch trading pairs from DEX module
        const response = await fetch(`${API_BASE}/sharehodl/dex/v1/pairs`);
        if (response.ok) {
          const data = await response.json();
          if (data.pairs && data.pairs.length > 0) {
            setPairs(data.pairs);
            setIsLive(true);
          }
        }
      } catch {
        // DEX module not live yet
        setIsLive(false);
      } finally {
        setIsLoading(false);
      }
    };

    checkDexStatus();
  }, []);

  if (isLoading) {
    return (
      <div className="trade-screen">
        <div className="loading-state">
          <Loader2 className="spin" size={40} />
          <p>Loading trading pairs...</p>
        </div>
        <style>{styles}</style>
      </div>
    );
  }

  // Coming Soon State
  if (!isLive) {
    return (
      <div className="trade-screen">
        <div className="coming-soon">
          <div className="coming-soon-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M3 3v18h18" />
              <path d="M18.7 8l-5.1 5.2-2.8-2.7L7 14.3" />
            </svg>
          </div>
          <h1 className="coming-soon-title">Trading</h1>
          <p className="coming-soon-subtitle">Coming Soon</p>
          <div className="coming-soon-desc">
            <p>The ShareHODL decentralized exchange is under development.</p>
            <p className="features-title">Upcoming Features:</p>
            <ul className="features-list">
              <li>Trade tokenized equities 24/7</li>
              <li>Market and limit orders</li>
              <li>Ultra-low trading fees (0.3%)</li>
              <li>Instant settlement on-chain</li>
              <li>Advanced order book trading</li>
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

  // Live State - Show real trading pairs when API is ready
  return (
    <div className="trade-screen">
      <div className="trade-header">
        <h1 className="trade-title">Trading</h1>
      </div>

      {pairs.length === 0 ? (
        <div className="empty-state">
          <p>No trading pairs available yet</p>
        </div>
      ) : (
        <div className="pairs-list">
          {pairs.map((pair) => (
            <div key={pair.id} className="pair-card">
              <div className="pair-info">
                <span className="pair-symbol">{pair.baseSymbol}/{pair.quoteSymbol}</span>
              </div>
              <div className="pair-price">
                <span className="price">{pair.lastPrice}</span>
                <span className={`change ${parseFloat(pair.priceChange24h) >= 0 ? 'positive' : 'negative'}`}>
                  {parseFloat(pair.priceChange24h) >= 0 ? '+' : ''}{pair.priceChange24h}%
                </span>
              </div>
            </div>
          ))}
        </div>
      )}

      <style>{styles}</style>
    </div>
  );
}

const styles = `
  .trade-screen {
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

  .trade-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
  }

  .trade-title {
    font-size: 24px;
    font-weight: 700;
    color: white;
    margin: 0;
  }

  .empty-state {
    text-align: center;
    padding: 40px 20px;
    color: #8b949e;
    background: #161B22;
    border-radius: 14px;
  }

  .pairs-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .pair-card {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: #161B22;
    border-radius: 14px;
    padding: 16px;
  }

  .pair-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .pair-symbol {
    font-size: 16px;
    font-weight: 600;
    color: white;
  }

  .pair-price {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 4px;
  }

  .price {
    font-size: 16px;
    font-weight: 600;
    color: white;
  }

  .change {
    font-size: 13px;
    font-weight: 500;
  }

  .change.positive {
    color: #10b981;
  }

  .change.negative {
    color: #ef4444;
  }
`;
