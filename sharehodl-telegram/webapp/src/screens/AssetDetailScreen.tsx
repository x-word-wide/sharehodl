/**
 * Asset Detail Screen - Shows details for a specific asset
 * Similar to Trust Wallet's asset view
 */

import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useWalletStore } from '../services/walletStore';
import { CHAIN_CONFIGS, getTokenById, Transaction } from '../types';

export function AssetDetailScreen() {
  const navigate = useNavigate();
  const { tokenId } = useParams<{ tokenId: string }>();
  const { assets, refreshBalances } = useWalletStore();
  const tg = window.Telegram?.WebApp;

  const [isRefreshing, setIsRefreshing] = useState(false);

  // Find the asset
  const asset = assets.find(a => a.token.id === tokenId);
  const token = tokenId ? getTokenById(tokenId) : undefined;

  useEffect(() => {
    if (!asset && !token) {
      navigate('/portfolio');
    }
  }, [asset, token, navigate]);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    tg?.HapticFeedback?.impactOccurred('light');
    await refreshBalances();
    setTimeout(() => setIsRefreshing(false), 500);
  };

  const handleSend = () => {
    tg?.HapticFeedback?.impactOccurred('medium');
    navigate(`/send?token=${tokenId}`);
  };

  const handleReceive = () => {
    tg?.HapticFeedback?.impactOccurred('medium');
    navigate(`/receive?token=${tokenId}`);
  };

  const handleSwap = () => {
    tg?.HapticFeedback?.impactOccurred('medium');
    navigate(`/trade?from=${tokenId}`);
  };

  if (!asset || !token) {
    return (
      <div className="asset-detail-screen">
        <div className="loading-state">
          <div className="spinner" />
          <p>Loading...</p>
        </div>
      </div>
    );
  }

  const chainConfig = CHAIN_CONFIGS[token.chain];
  const isPositive = asset.priceChange24h >= 0;

  // Real transactions will be fetched from blockchain
  // Empty array until transaction indexer is connected
  const transactions: Transaction[] = [];

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
    <div className="asset-detail-screen">
      {/* Header */}
      <div className="detail-header">
        <button className="back-btn" onClick={() => navigate(-1)}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M15 18l-6-6 6-6" />
          </svg>
        </button>
        <div className="header-title">
          <span className="token-name">{token.name}</span>
          <span className="chain-name" style={{ color: chainConfig.color }}>{chainConfig.name}</span>
        </div>
        <button className={`refresh-btn ${isRefreshing ? 'spinning' : ''}`} onClick={handleRefresh}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 12a9 9 0 11-9-9c2.52 0 4.93 1 6.74 2.74L21 8" />
            <path d="M21 3v5h-5" />
          </svg>
        </button>
      </div>

      {/* Balance Card */}
      <div className="balance-section">
        <div className="token-icon-large" style={{ background: `${token.color}20` }}>
          <span style={{ color: token.color }}>{token.symbol.slice(0, 2)}</span>
        </div>

        <h1 className="balance-amount">{asset.balanceFormatted} {token.symbol}</h1>
        <p className="balance-usd">${parseFloat(asset.balanceUsd).toLocaleString()}</p>

        <div className="price-info">
          <span className="current-price">${asset.price.toLocaleString()}</span>
          <span className={`price-change ${isPositive ? 'positive' : 'negative'}`}>
            {isPositive ? '▲' : '▼'} {Math.abs(asset.priceChange24h).toFixed(2)}%
          </span>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="action-buttons">
        <button className="action-btn" onClick={handleSend}>
          <div className="action-icon send">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 5v14M5 12l7-7 7 7" />
            </svg>
          </div>
          <span>Send</span>
        </button>
        <button className="action-btn" onClick={handleReceive}>
          <div className="action-icon receive">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 19V5M5 12l7 7 7-7" />
            </svg>
          </div>
          <span>Receive</span>
        </button>
        <button className="action-btn" onClick={handleSwap}>
          <div className="action-icon swap">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M7 10l5 5 5-5" />
              <path d="M17 14l-5-5-5 5" />
            </svg>
          </div>
          <span>Swap</span>
        </button>
      </div>

      {/* Address Section */}
      <div className="address-section">
        <span className="address-label">{chainConfig.name} Address</span>
        <button
          className="address-value"
          onClick={() => {
            navigator.clipboard.writeText(asset.address);
            tg?.HapticFeedback?.notificationOccurred('success');
            tg?.showAlert('Address copied!');
          }}
        >
          <span>{asset.address.slice(0, 12)}...{asset.address.slice(-10)}</span>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
            <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
          </svg>
        </button>
      </div>

      {/* Transactions */}
      <div className="transactions-section">
        <h3 className="section-title">Recent Transactions</h3>

        {transactions.length > 0 ? (
          <div className="transactions-list">
            {transactions.map((tx) => (
              <div key={tx.hash} className="transaction-item">
                <div className={`tx-icon ${tx.type.toLowerCase()}`}>
                  {tx.type === 'RECEIVE' ? (
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M12 19V5M5 12l7 7 7-7" />
                    </svg>
                  ) : (
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M12 5v14M5 12l7-7 7 7" />
                    </svg>
                  )}
                </div>
                <div className="tx-info">
                  <span className="tx-type">{tx.type === 'RECEIVE' ? 'Received' : 'Sent'}</span>
                  <span className="tx-date">{formatDate(tx.timestamp)}</span>
                </div>
                <div className="tx-amount">
                  <span className={tx.type === 'RECEIVE' ? 'positive' : 'negative'}>
                    {tx.type === 'RECEIVE' ? '+' : '-'}{tx.amount} {tx.symbol}
                  </span>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="empty-transactions">
            <p>No transactions yet</p>
          </div>
        )}
      </div>

      <style>{`
        .asset-detail-screen {
          min-height: 100vh;
          padding-bottom: 100px;
        }

        .loading-state {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 100vh;
          gap: 16px;
          color: #8b949e;
        }

        .spinner {
          width: 32px;
          height: 32px;
          border: 2px solid #30363d;
          border-top-color: #1E40AF;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          to { transform: rotate(360deg); }
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
          background: #161B22;
          border: none;
          display: flex;
          align-items: center;
          justify-content: center;
          cursor: pointer;
        }

        .back-btn svg, .refresh-btn svg {
          width: 20px;
          height: 20px;
          color: #8b949e;
        }

        .refresh-btn.spinning svg {
          animation: spin 1s linear infinite;
        }

        .header-title {
          display: flex;
          flex-direction: column;
          align-items: center;
        }

        .token-name {
          font-size: 18px;
          font-weight: 600;
          color: white;
        }

        .chain-name {
          font-size: 12px;
          font-weight: 500;
        }

        .balance-section {
          display: flex;
          flex-direction: column;
          align-items: center;
          padding: 20px 16px 32px;
        }

        .token-icon-large {
          width: 72px;
          height: 72px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 28px;
          font-weight: 700;
          margin-bottom: 20px;
        }

        .balance-amount {
          font-size: 36px;
          font-weight: 700;
          color: white;
          margin: 0;
        }

        .balance-usd {
          font-size: 18px;
          color: #8b949e;
          margin: 8px 0 16px;
        }

        .price-info {
          display: flex;
          align-items: center;
          gap: 12px;
        }

        .current-price {
          font-size: 15px;
          color: #8b949e;
        }

        .price-change {
          font-size: 14px;
          font-weight: 600;
          padding: 4px 10px;
          border-radius: 8px;
        }

        .price-change.positive {
          color: #10b981;
          background: rgba(16, 185, 129, 0.1);
        }

        .price-change.negative {
          color: #ef4444;
          background: rgba(239, 68, 68, 0.1);
        }

        .action-buttons {
          display: flex;
          justify-content: center;
          gap: 24px;
          padding: 0 16px 32px;
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
          transition: all 0.2s ease;
        }

        .action-icon.send {
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
        }

        .action-icon.receive {
          background: linear-gradient(135deg, #10b981 0%, #059669 100%);
        }

        .action-icon.swap {
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

        .action-btn:active .action-icon {
          transform: scale(0.95);
        }

        .address-section {
          margin: 0 16px 24px;
          padding: 16px;
          background: #161B22;
          border-radius: 14px;
        }

        .address-label {
          display: block;
          font-size: 13px;
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
          font-family: monospace;
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
          font-size: 16px;
          font-weight: 600;
          color: white;
          margin: 0 0 16px;
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
          background: #161B22;
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

        .tx-icon.receive {
          background: rgba(16, 185, 129, 0.1);
          color: #10b981;
        }

        .tx-icon.send {
          background: rgba(239, 68, 68, 0.1);
          color: #ef4444;
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
          font-size: 15px;
          font-weight: 500;
          color: white;
        }

        .tx-date {
          font-size: 13px;
          color: #8b949e;
        }

        .tx-amount {
          text-align: right;
        }

        .tx-amount .positive {
          color: #10b981;
          font-weight: 600;
        }

        .tx-amount .negative {
          color: #ef4444;
          font-weight: 600;
        }

        .empty-transactions {
          padding: 40px 20px;
          text-align: center;
          color: #8b949e;
        }
      `}</style>
    </div>
  );
}
