/**
 * Equity Profile Screen - Premium equity details view
 */

import { useParams, useNavigate } from 'react-router-dom';

// Equity data type - will be populated from blockchain when equity module is connected
interface EquityData {
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
}

// Empty equities - will be fetched from blockchain equity module
const EQUITIES: Record<string, EquityData> = {};

export function EquityProfileScreen() {
  const { equityId } = useParams();
  const navigate = useNavigate();
  const tg = window.Telegram?.WebApp;

  // All equities currently show "not found" since we're connected to real blockchain
  const equity = equityId ? EQUITIES[equityId] : null;

  // Always show not found for now - no demo data
  return (
    <div className="not-found">
      <span className="not-found-icon">📊</span>
      <p className="not-found-title">Equity Profile Not Found</p>
      <p className="not-found-desc">
        {equity ? 'Loading equity profile...' : 'Equity profiles will be available when the equity module is connected.'}
      </p>
      <button onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); navigate('/portfolio'); }}>
        Go Back
      </button>
      <style>{`
        .not-found {
          min-height: 100vh;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 40px 20px;
          text-align: center;
          background: #0D1117;
        }
        .not-found-icon {
          font-size: 64px;
          margin-bottom: 16px;
        }
        .not-found-title {
          font-size: 20px;
          font-weight: 600;
          color: white;
          margin: 0 0 8px;
        }
        .not-found-desc {
          font-size: 14px;
          color: #8b949e;
          margin: 0 0 24px;
          max-width: 280px;
          line-height: 1.5;
        }
        .not-found button {
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
        .not-found button:active {
          transform: scale(0.97);
        }
      `}</style>
    </div>
  );
}
