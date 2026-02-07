/**
 * P2P Trading Screen - Peer-to-peer marketplace
 * Shows Coming Soon state until P2P module is live
 */

import { useState, useEffect } from 'react';
import { Loader2 } from 'lucide-react';

const API_BASE = import.meta.env.VITE_SHAREHODL_REST || 'https://api.sharehodl.com';

interface P2PListing {
  id: string;
  type: 'BUY' | 'SELL';
  trader: {
    name: string;
    rating: number;
    trades: number;
    verified: boolean;
    online: boolean;
  };
  asset: string;
  currency: string;
  price: number;
  minAmount: number;
  maxAmount: number;
  available: number;
  paymentMethods: string[];
  avgReleaseTime: string;
}

export function P2PScreen() {
  const [isLoading, setIsLoading] = useState(true);
  const [listings, setListings] = useState<P2PListing[]>([]);
  const [isLive, setIsLive] = useState(false);

  useEffect(() => {
    const checkP2PStatus = async () => {
      try {
        // Try to fetch P2P listings from blockchain
        const response = await fetch(`${API_BASE}/sharehodl/p2p/v1/listings`);
        if (response.ok) {
          const data = await response.json();
          if (data.listings && data.listings.length > 0) {
            setListings(data.listings);
            setIsLive(true);
          }
        }
      } catch {
        // P2P module not live yet
        setIsLive(false);
      } finally {
        setIsLoading(false);
      }
    };

    checkP2PStatus();
  }, []);

  if (isLoading) {
    return (
      <div className="p2p-screen">
        <div className="loading-state">
          <Loader2 className="spin" size={40} />
          <p>Loading P2P marketplace...</p>
        </div>
        <style>{styles}</style>
      </div>
    );
  }

  // Coming Soon State
  if (!isLive) {
    return (
      <div className="p2p-screen">
        <div className="coming-soon">
          <div className="coming-soon-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
              <circle cx="9" cy="7" r="4" />
              <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
              <path d="M16 3.13a4 4 0 0 1 0 7.75" />
            </svg>
          </div>
          <h1 className="coming-soon-title">P2P Trading</h1>
          <p className="coming-soon-subtitle">Coming Soon</p>
          <div className="coming-soon-desc">
            <p>The ShareHODL peer-to-peer marketplace is under development.</p>
            <p className="features-title">Upcoming Features:</p>
            <ul className="features-list">
              <li>Buy and sell crypto directly with other users</li>
              <li>Escrow protection for secure trades</li>
              <li>Multiple payment methods supported</li>
              <li>Multi-currency support (USD, NGN, GBP)</li>
              <li>Verified trader badges and ratings</li>
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

  // Live State - Show real listings when API is ready
  return (
    <div className="p2p-screen">
      <div className="p2p-header">
        <h1 className="p2p-title">P2P Trading</h1>
        <div className="escrow-badge">
          <svg className="shield-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            <path d="M9 12l2 2 4-4" />
          </svg>
          <span>Escrow Protected</span>
        </div>
      </div>

      {listings.length === 0 ? (
        <div className="empty-state">
          <p>No P2P listings available yet</p>
        </div>
      ) : (
        <div className="listings">
          {listings.map((listing) => (
            <div key={listing.id} className="listing-card">
              <div className="trader-row">
                <div className="trader-avatar">
                  <span className="avatar-text">
                    {listing.trader.name.slice(0, 2).toUpperCase()}
                  </span>
                </div>
                <div className="trader-info">
                  <span className="trader-name">{listing.trader.name}</span>
                  <span className="trader-stats">{listing.trader.trades} trades</span>
                </div>
              </div>
              <div className="price-row">
                <span className="price-label">{listing.type}</span>
                <span className="price-value">{listing.price} {listing.currency}</span>
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
  .p2p-screen {
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

  .p2p-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
  }

  .p2p-title {
    font-size: 24px;
    font-weight: 700;
    color: white;
    margin: 0;
  }

  .escrow-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 12px;
    background: rgba(16, 185, 129, 0.1);
    border-radius: 20px;
    color: #10b981;
    font-size: 12px;
    font-weight: 500;
  }

  .shield-icon {
    width: 14px;
    height: 14px;
  }

  .empty-state {
    text-align: center;
    padding: 40px 20px;
    color: #8b949e;
    background: #161B22;
    border-radius: 14px;
  }

  .listings {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .listing-card {
    background: #161B22;
    border-radius: 14px;
    padding: 16px;
  }

  .trader-row {
    display: flex;
    gap: 12px;
    margin-bottom: 12px;
  }

  .trader-avatar {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .avatar-text {
    font-size: 16px;
    font-weight: 600;
    color: white;
  }

  .trader-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .trader-name {
    font-size: 15px;
    font-weight: 600;
    color: white;
  }

  .trader-stats {
    font-size: 13px;
    color: #8b949e;
  }

  .price-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .price-label {
    font-size: 12px;
    color: #8b949e;
  }

  .price-value {
    font-size: 18px;
    font-weight: 600;
    color: white;
  }
`;
