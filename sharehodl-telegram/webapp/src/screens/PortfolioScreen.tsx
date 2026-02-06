/**
 * Portfolio Screen - Main wallet dashboard
 * Professional design inspired by Trust Wallet
 */

import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '../services/walletStore';
import { useStakingStore } from '../services/stakingStore';
import { Chain, AssetHolding, CHAIN_CONFIGS, TokenType } from '../types';

const SERVICES = [
  { id: 'staking', iconType: 'star', title: 'Staking', desc: 'Earn rewards', path: '/staking', color: '#1E40AF' },
  { id: 'governance', iconType: 'governance', title: 'Governance', desc: 'Vote & delegate', path: '/governance', color: '#8B5CF6' },
  { id: 'inheritance', iconType: 'inherit', title: 'Inherit', desc: 'Asset transfer', path: '/inheritance', color: '#14B8A6' },
  { id: 'p2p', iconType: 'users', title: 'P2P', desc: 'Peer trading', path: '/p2p', color: '#10B981' },
  { id: 'lending', iconType: 'coins', title: 'Lending', desc: 'Supply & borrow', path: '/lending', color: '#F59E0B' },
  { id: 'bridge', iconType: 'bridge', title: 'Bridge', desc: 'Cross-chain', path: '/bridge', color: '#3B82F6' }
];

// Demo equity holdings
const DEMO_EQUITIES = [
  {
    id: 'sharehodl-plc',
    symbol: 'SHDL',
    name: 'ShareHODL PLC',
    shares: 1250,
    pricePerShare: 12.50,
    change24h: 4.25,
    color: '#1E40AF'
  },
  {
    id: 'property-mainnet',
    symbol: 'PROP',
    name: 'Property Mainnet',
    shares: 500,
    pricePerShare: 8.75,
    change24h: -1.30,
    color: '#10B981'
  },
  {
    id: 'tech-ventures',
    symbol: 'TVNT',
    name: 'Tech Ventures Ltd',
    shares: 2000,
    pricePerShare: 3.20,
    change24h: 7.80,
    color: '#F59E0B'
  },
  {
    id: 'green-energy',
    symbol: 'GREN',
    name: 'Green Energy Corp',
    shares: 750,
    pricePerShare: 15.00,
    change24h: 2.15,
    color: '#059669'
  },
  {
    id: 'fintech-global',
    symbol: 'FNTK',
    name: 'FinTech Global',
    shares: 300,
    pricePerShare: 45.00,
    change24h: -0.50,
    color: '#6366F1'
  }
];

// Service icons
const ServiceIcon = ({ type, color }: { type: string; color: string }) => {
  const iconProps = { width: 24, height: 24, stroke: color, strokeWidth: 1.5, fill: 'none' };

  switch (type) {
    case 'star':
      return (
        <svg viewBox="0 0 24 24" width={24} height={24}>
          <polygon
            points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"
            fill={color}
            stroke={color}
            strokeWidth="0.5"
          />
        </svg>
      );
    case 'users':
      return (
        <svg viewBox="0 0 24 24" {...iconProps}>
          <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
          <circle cx="9" cy="7" r="4" />
          <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
          <path d="M16 3.13a4 4 0 0 1 0 7.75" />
        </svg>
      );
    case 'coins':
      return (
        <svg viewBox="0 0 24 24" {...iconProps}>
          <circle cx="8" cy="8" r="6" />
          <path d="M18.09 10.37A6 6 0 1 1 10.34 18" />
          <path d="M7 6h2v4H7z" />
        </svg>
      );
    case 'bridge':
      return (
        <svg viewBox="0 0 24 24" {...iconProps}>
          <path d="M4 18h16" />
          <path d="M4 18v-2a8 8 0 0 1 16 0v2" />
          <path d="M4 12h16" />
          <path d="M8 12v6" />
          <path d="M16 12v6" />
        </svg>
      );
    case 'governance':
      return (
        <svg viewBox="0 0 24 24" {...iconProps}>
          <path d="M12 2L2 7l10 5 10-5-10-5z" />
          <path d="M2 17l10 5 10-5" />
          <path d="M2 12l10 5 10-5" />
        </svg>
      );
    case 'inherit':
      return (
        <svg viewBox="0 0 24 24" {...iconProps}>
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          <path d="M12 8v4" />
          <path d="M12 16h.01" />
        </svg>
      );
    default:
      return null;
  }
};

// Token icon component with fallback
const TokenIcon = ({ symbol, color, size = 44 }: { symbol: string; color: string; size?: number }) => {
  return (
    <div
      className="token-icon"
      style={{
        width: size,
        height: size,
        borderRadius: '50%',
        background: `${color}20`,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        fontSize: size * 0.36,
        fontWeight: 700,
        color: color
      }}
    >
      {symbol.slice(0, 2).toUpperCase()}
    </div>
  );
};

export function PortfolioScreen() {
  const navigate = useNavigate();
  const { accounts, assets, totalBalanceUsd, refreshBalances, isLoading } = useWalletStore();
  const { position, fetchStakingPosition } = useStakingStore();
  const tg = window.Telegram?.WebApp;

  const [isRefreshing, setIsRefreshing] = useState(false);
  const [selectedTab, setSelectedTab] = useState<'crypto' | 'equity'>('equity');

  // Get ShareHODL address for staking
  const sharehodlAccount = accounts.find(a => a.chain === Chain.SHAREHODL);
  const address = sharehodlAccount?.address || '';

  useEffect(() => {
    refreshBalances();
    if (address) {
      fetchStakingPosition(address);
    }
  }, [address, refreshBalances, fetchStakingPosition]);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    tg?.HapticFeedback?.impactOccurred('light');
    await refreshBalances();
    setTimeout(() => setIsRefreshing(false), 500);
  };

  // Count unique chains with assets
  const chainCount = new Set(assets.map(a => a.token.chain)).size;

  const userName = tg?.initDataUnsafe?.user?.first_name || 'there';

  return (
    <div className="portfolio-screen">
      {/* Header */}
      <div className="portfolio-header">
        <div className="header-left">
          <div className="greeting-row">
            <h1 className="greeting">Hello, {userName}</h1>
            {position && (
              <div
                className="tier-badge"
                style={{
                  background: `${position.tierConfig.color}20`,
                  borderColor: position.tierConfig.color
                }}
                onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); navigate('/staking'); }}
              >
                <span className="tier-icon">{position.tierConfig.icon}</span>
                <span className="tier-name" style={{ color: position.tierConfig.color }}>
                  {position.tierConfig.name}
                </span>
              </div>
            )}
          </div>
          <p className="greeting-sub">
            {position && position.stakedAmount > 0
              ? `${position.apr.toFixed(1)}% APR on staking`
              : 'Welcome back'}
          </p>
        </div>
        <button className={`refresh-btn ${isRefreshing ? 'spinning' : ''}`} onClick={handleRefresh}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M21 12a9 9 0 11-9-9c2.52 0 4.93 1 6.74 2.74L21 8" />
            <path d="M21 3v5h-5" />
          </svg>
        </button>
      </div>

      {/* Balance Card */}
      <div className="balance-card">
        <div className="balance-bg" />
        <div className="balance-content">
          <p className="balance-label">Total Balance</p>
          <h2 className="balance-amount">
            <span className="currency">$</span>
            {totalBalanceUsd.toLocaleString('en-US', { minimumFractionDigits: 2 })}
          </h2>
          <div className="balance-info">
            <div className="info-chip">
              <span className="chip-emoji">🔗</span>
              <span>{chainCount} chains</span>
            </div>
            <div className="info-chip">
              <span className="chip-emoji">💎</span>
              <span>{assets.length} assets</span>
            </div>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="quick-actions">
          <button className="action-btn primary" onClick={() => { tg?.HapticFeedback?.impactOccurred('medium'); navigate('/send'); }}>
            <span className="action-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 5v14M5 12l7-7 7 7" />
              </svg>
            </span>
            <span className="action-label">Send</span>
          </button>
          <button className="action-btn" onClick={() => { tg?.HapticFeedback?.impactOccurred('medium'); navigate('/receive'); }}>
            <span className="action-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 19V5M5 12l7 7 7-7" />
              </svg>
            </span>
            <span className="action-label">Receive</span>
          </button>
          <button className="action-btn" onClick={() => { tg?.HapticFeedback?.impactOccurred('medium'); navigate('/trade'); }}>
            <span className="action-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M7 10l5 5 5-5" />
                <path d="M17 14l-5-5-5 5" />
              </svg>
            </span>
            <span className="action-label">Trade</span>
          </button>
        </div>
      </div>

      {/* Staking Card */}
      {position && position.stakedAmount > 0 && (
        <div className="staking-card" onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); navigate('/staking'); }}>
          <div className="staking-info">
            <div className="staking-header-row">
              <span className="staking-label">Staked Balance</span>
              <span
                className="multiplier-badge"
                style={{ background: `${position.tierConfig.color}20`, color: position.tierConfig.color }}
              >
                {position.tierConfig.rewardMultiplier}x rewards
              </span>
            </div>
            <span className="staking-amount">{position.stakedAmount.toLocaleString()} HODL</span>
            {position.pendingRewards > 0 && (
              <span className="pending-rewards">
                +{position.pendingRewards.toFixed(4)} HODL pending
              </span>
            )}
          </div>
          <div className="staking-arrow">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M9 18l6-6-6-6" />
            </svg>
          </div>
        </div>
      )}

      {/* Services */}
      <div className="services-section">
        <h3 className="section-title">DeFi Services</h3>
        <div className="services-grid">
          {SERVICES.map((service) => (
            <button
              key={service.id}
              className="service-card"
              onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); navigate(service.path); }}
            >
              <div className="service-icon" style={{ background: `${service.color}15` }}>
                <ServiceIcon type={service.iconType} color={service.color} />
              </div>
              <span className="service-title">{service.title}</span>
              <span className="service-desc">{service.desc}</span>
            </button>
          ))}
        </div>
      </div>

      {/* Assets Tab */}
      <div className="assets-section">
        <div className="assets-tabs">
          <button
            className={`tab ${selectedTab === 'equity' ? 'active' : ''}`}
            onClick={() => setSelectedTab('equity')}
          >
            Equity
          </button>
          <button
            className={`tab ${selectedTab === 'crypto' ? 'active' : ''}`}
            onClick={() => setSelectedTab('crypto')}
          >
            Crypto
          </button>
        </div>

        {/* Assets List */}
        <div className="assets-list">
          {isLoading && assets.length === 0 ? (
            <div className="loading-state">
              <div className="spinner" />
              <p>Loading assets...</p>
            </div>
          ) : selectedTab === 'crypto' ? (
            assets.length > 0 ? (
              assets.map((asset: AssetHolding) => {
                const priceChange = asset.priceChange24h;
                const isPositive = priceChange >= 0;
                const chainConfig = CHAIN_CONFIGS[asset.token.chain];

                return (
                  <button
                    key={asset.token.id}
                    className="asset-item"
                    onClick={() => {
                      tg?.HapticFeedback?.impactOccurred('light');
                      navigate(`/asset/${asset.token.id}`);
                    }}
                  >
                    <TokenIcon symbol={asset.token.symbol} color={asset.token.color} />
                    <div className="asset-info">
                      <div className="asset-name-row">
                        <span className="asset-symbol">{asset.token.symbol}</span>
                        {asset.token.type !== TokenType.NATIVE && (
                          <span className="chain-badge" style={{ color: chainConfig.color }}>
                            {chainConfig.name}
                          </span>
                        )}
                      </div>
                      <div className="asset-price-row">
                        <span className="asset-price">${asset.price.toLocaleString()}</span>
                        <span className={`price-change ${isPositive ? 'positive' : 'negative'}`}>
                          {isPositive ? '+' : ''}{priceChange.toFixed(2)}%
                        </span>
                      </div>
                    </div>
                    <div className="asset-balance">
                      <span className="balance-amount">{asset.balanceFormatted}</span>
                      <span className="balance-usd-value">${parseFloat(asset.balanceUsd).toLocaleString()}</span>
                    </div>
                  </button>
                );
              })
            ) : (
              <div className="empty-state">
                <span className="empty-icon">💰</span>
                <p className="empty-title">No Assets Yet</p>
                <p className="empty-desc">Receive crypto to see it here</p>
                <button className="empty-btn" onClick={() => navigate('/receive')}>
                  Receive
                </button>
              </div>
            )
          ) : (
            DEMO_EQUITIES.map((equity) => {
              const totalValue = equity.shares * equity.pricePerShare;
              const isPositive = equity.change24h >= 0;

              return (
                <button
                  key={equity.id}
                  className="asset-item"
                  onClick={() => {
                    tg?.HapticFeedback?.impactOccurred('light');
                    navigate(`/equity/${equity.id}`);
                  }}
                >
                  <TokenIcon symbol={equity.symbol} color={equity.color} />
                  <div className="asset-info">
                    <div className="asset-name-row">
                      <span className="asset-symbol equity-name">{equity.name}</span>
                    </div>
                    <div className="asset-price-row">
                      <span className="asset-price">${equity.pricePerShare.toFixed(2)}/share</span>
                      <span className={`price-change ${isPositive ? 'positive' : 'negative'}`}>
                        {isPositive ? '+' : ''}{equity.change24h.toFixed(2)}%
                      </span>
                    </div>
                  </div>
                  <div className="asset-balance">
                    <span className="balance-amount">{equity.shares.toLocaleString()}</span>
                    <span className="balance-usd-value">${totalValue.toLocaleString()}</span>
                  </div>
                </button>
              );
            })
          )}
        </div>

        {/* Add Token Button */}
        {selectedTab === 'crypto' && (
          <button className="add-token-btn" onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); navigate('/add-token'); }}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 8v8M8 12h8" />
            </svg>
            <span>Add Token</span>
          </button>
        )}
      </div>

      <style>{`
        .portfolio-screen {
          min-height: 100vh;
          padding-bottom: 100px;
        }

        .portfolio-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 20px 16px 16px;
        }

        .greeting {
          font-size: 20px;
          font-weight: 700;
          color: white;
          margin: 0;
        }

        .greeting-row {
          display: flex;
          align-items: center;
          gap: 10px;
        }

        .tier-badge {
          display: inline-flex;
          align-items: center;
          gap: 4px;
          padding: 4px 10px;
          border-radius: 12px;
          border: 1px solid;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .tier-badge:active {
          transform: scale(0.95);
        }

        .tier-icon {
          font-size: 12px;
        }

        .tier-name {
          font-size: 12px;
          font-weight: 600;
        }

        .greeting-sub {
          font-size: 14px;
          color: #8b949e;
          margin: 4px 0 0;
        }

        .refresh-btn {
          width: 44px;
          height: 44px;
          border-radius: 50%;
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(10px);
          -webkit-backdrop-filter: blur(10px);
          border: 1px solid rgba(48, 54, 61, 0.4);
          display: flex;
          align-items: center;
          justify-content: center;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .refresh-btn:active {
          transform: scale(0.95);
          background: rgba(22, 27, 34, 0.8);
        }

        .refresh-btn svg {
          width: 20px;
          height: 20px;
          color: #8b949e;
        }

        .refresh-btn.spinning svg {
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }

        .balance-card {
          margin: 0 16px 20px;
          padding: 24px;
          border-radius: 20px;
          background: rgba(22, 27, 34, 0.75);
          backdrop-filter: blur(16px);
          -webkit-backdrop-filter: blur(16px);
          border: 1px solid rgba(48, 54, 61, 0.6);
          box-shadow: 0 8px 32px rgba(0, 0, 0, 0.25);
          position: relative;
          overflow: hidden;
        }

        .balance-bg {
          position: absolute;
          inset: 0;
          background:
            radial-gradient(circle at 20% 80%, rgba(30, 64, 175, 0.2) 0%, transparent 40%),
            radial-gradient(circle at 80% 20%, rgba(59, 130, 246, 0.15) 0%, transparent 40%),
            radial-gradient(circle at 50% 50%, rgba(16, 185, 129, 0.05) 0%, transparent 50%);
          pointer-events: none;
        }

        .balance-content {
          position: relative;
          z-index: 1;
        }

        .balance-label {
          font-size: 14px;
          color: #8b949e;
          margin: 0 0 8px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .balance-amount {
          font-size: 44px;
          font-weight: 800;
          color: white;
          margin: 0;
          line-height: 1;
          letter-spacing: -1px;
        }

        .currency {
          font-size: 28px;
          font-weight: 600;
          opacity: 0.6;
          margin-right: 2px;
        }

        .balance-info {
          display: flex;
          align-items: center;
          gap: 10px;
          margin-top: 14px;
        }

        .info-chip {
          display: inline-flex;
          align-items: center;
          gap: 6px;
          padding: 6px 12px;
          background: rgba(255, 255, 255, 0.06);
          border-radius: 20px;
          font-size: 13px;
          color: #8b949e;
          transition: background 0.2s ease;
        }

        .chip-emoji {
          font-size: 12px;
        }

        .quick-actions {
          display: flex;
          gap: 12px;
          margin-top: 24px;
          position: relative;
          z-index: 1;
        }

        .action-btn {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 8px;
          padding: 16px 14px;
          background: rgba(48, 54, 61, 0.4);
          backdrop-filter: blur(8px);
          -webkit-backdrop-filter: blur(8px);
          border: 1px solid rgba(255, 255, 255, 0.05);
          border-radius: 16px;
          cursor: pointer;
          transition: all 0.2s ease;
          position: relative;
          overflow: hidden;
        }

        .action-btn::before {
          content: '';
          position: absolute;
          inset: 0;
          background: radial-gradient(circle at center, rgba(255, 255, 255, 0.1) 0%, transparent 70%);
          opacity: 0;
          transition: opacity 0.3s;
        }

        .action-btn:active::before {
          opacity: 1;
        }

        .action-btn.primary {
          background: linear-gradient(135deg, rgba(30, 64, 175, 0.9) 0%, rgba(59, 130, 246, 0.9) 100%);
          border: none;
          box-shadow: 0 4px 20px rgba(30, 64, 175, 0.3);
        }

        .action-btn:active {
          transform: scale(0.97);
        }

        .action-icon {
          width: 40px;
          height: 40px;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .action-icon svg {
          width: 24px;
          height: 24px;
          color: white;
        }

        .action-label {
          font-size: 13px;
          font-weight: 600;
          color: white;
        }

        .staking-card {
          margin: 0 16px 20px;
          padding: 18px;
          background: linear-gradient(135deg, rgba(30, 64, 175, 0.15) 0%, rgba(59, 130, 246, 0.1) 100%);
          backdrop-filter: blur(10px);
          -webkit-backdrop-filter: blur(10px);
          border: 1px solid rgba(30, 64, 175, 0.3);
          border-radius: 16px;
          display: flex;
          align-items: center;
          justify-content: space-between;
          cursor: pointer;
          transition: all 0.2s ease;
          box-shadow: 0 4px 20px rgba(30, 64, 175, 0.15);
        }

        .staking-card:active {
          transform: scale(0.98);
          background: linear-gradient(135deg, rgba(30, 64, 175, 0.2) 0%, rgba(59, 130, 246, 0.15) 100%);
        }

        .staking-info {
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .staking-header-row {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .staking-label {
          font-size: 13px;
          color: #8b949e;
        }

        .multiplier-badge {
          padding: 2px 8px;
          border-radius: 8px;
          font-size: 11px;
          font-weight: 600;
        }

        .staking-amount {
          font-size: 20px;
          font-weight: 700;
          color: white;
        }

        .pending-rewards {
          font-size: 13px;
          color: #10b981;
        }

        .staking-arrow {
          width: 24px;
          height: 24px;
          color: #8b949e;
        }

        .staking-arrow svg {
          width: 24px;
          height: 24px;
        }

        .services-section {
          padding: 0 16px;
          margin-bottom: 24px;
        }

        .section-title {
          font-size: 16px;
          font-weight: 600;
          color: white;
          margin: 0 0 12px;
        }

        .services-grid {
          display: flex;
          gap: 10px;
          overflow-x: auto;
          padding-bottom: 8px;
          scrollbar-width: none;
          -ms-overflow-style: none;
        }

        .services-grid::-webkit-scrollbar {
          display: none;
        }

        .service-card {
          display: flex;
          flex-direction: column;
          align-items: center;
          min-width: 78px;
          padding: 14px 10px;
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(12px);
          -webkit-backdrop-filter: blur(12px);
          border: 1px solid rgba(48, 54, 61, 0.5);
          border-radius: 16px;
          cursor: pointer;
          transition: all 0.25s ease;
          flex-shrink: 0;
        }

        .service-card:active {
          transform: scale(0.95);
          background: rgba(22, 27, 34, 0.8);
        }

        .service-icon {
          width: 42px;
          height: 42px;
          border-radius: 12px;
          display: flex;
          align-items: center;
          justify-content: center;
          margin-bottom: 8px;
          font-size: 18px;
          position: relative;
        }

        .service-icon::after {
          content: '';
          position: absolute;
          inset: -4px;
          border-radius: 18px;
          background: inherit;
          opacity: 0.2;
          filter: blur(8px);
          z-index: -1;
        }

        .service-title {
          font-size: 12px;
          font-weight: 600;
          color: white;
          text-align: center;
        }

        .service-desc {
          font-size: 10px;
          color: #8b949e;
          margin-top: 2px;
        }

        .assets-section {
          padding: 0 16px;
        }

        .assets-tabs {
          display: flex;
          gap: 6px;
          padding: 5px;
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(10px);
          -webkit-backdrop-filter: blur(10px);
          border: 1px solid rgba(48, 54, 61, 0.4);
          border-radius: 14px;
          margin-bottom: 16px;
        }

        .tab {
          flex: 1;
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

        .assets-list {
          display: flex;
          flex-direction: column;
          gap: 10px;
        }

        .asset-item {
          display: flex;
          align-items: center;
          gap: 14px;
          padding: 16px;
          background: rgba(22, 27, 34, 0.5);
          backdrop-filter: blur(8px);
          -webkit-backdrop-filter: blur(8px);
          border: 1px solid rgba(48, 54, 61, 0.4);
          border-radius: 16px;
          cursor: pointer;
          transition: all 0.2s ease;
          text-align: left;
          width: 100%;
        }

        .asset-item:active {
          transform: scale(0.98);
          background: rgba(22, 27, 34, 0.7);
          border-color: rgba(48, 54, 61, 0.6);
        }

        .asset-icon {
          width: 44px;
          height: 44px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 16px;
          font-weight: 700;
        }

        .asset-info {
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: 4px;
          min-width: 0;
        }

        .asset-name-row {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .asset-symbol {
          font-size: 16px;
          font-weight: 600;
          color: white;
        }

        .asset-symbol.equity-name {
          font-size: 14px;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
          max-width: 160px;
        }

        .chain-badge {
          font-size: 10px;
          font-weight: 500;
          padding: 2px 6px;
          background: rgba(255, 255, 255, 0.05);
          border-radius: 4px;
        }

        .asset-price-row {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .asset-price {
          font-size: 13px;
          color: #8b949e;
        }

        .price-change {
          font-size: 12px;
          font-weight: 500;
        }

        .price-change.positive {
          color: #10b981;
        }

        .price-change.negative {
          color: #ef4444;
        }

        .asset-name {
          font-size: 13px;
          color: #8b949e;
        }

        .asset-balance {
          text-align: right;
          flex-shrink: 0;
        }

        .asset-balance .balance-amount {
          display: block;
          font-size: 16px;
          font-weight: 600;
          color: white;
        }

        .balance-usd-value {
          display: block;
          font-size: 13px;
          color: #8b949e;
        }

        .add-token-btn {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          width: 100%;
          padding: 14px;
          margin-top: 12px;
          background: transparent;
          border: 1px dashed #30363d;
          border-radius: 14px;
          color: #8b949e;
          font-size: 14px;
          font-weight: 500;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .add-token-btn:active {
          background: rgba(48, 54, 61, 0.3);
        }

        .add-token-btn svg {
          width: 20px;
          height: 20px;
        }

        .loading-state {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 40px 20px;
          gap: 16px;
          color: #8b949e;
        }

        .loading-state .spinner {
          width: 32px;
          height: 32px;
          border: 2px solid #30363d;
          border-top-color: #1E40AF;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        .empty-state {
          padding: 40px 20px;
          text-align: center;
        }

        .empty-icon {
          font-size: 48px;
          display: block;
          margin-bottom: 16px;
        }

        .empty-title {
          font-size: 18px;
          font-weight: 600;
          color: white;
          margin: 0 0 8px;
        }

        .empty-desc {
          font-size: 14px;
          color: #8b949e;
          margin: 0 0 20px;
        }

        .empty-btn {
          padding: 12px 24px;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          border: none;
          border-radius: 12px;
          font-size: 15px;
          font-weight: 600;
          color: white;
          cursor: pointer;
        }
      `}</style>
    </div>
  );
}
