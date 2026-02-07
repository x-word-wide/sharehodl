/**
 * Market Screen - Professional equity marketplace
 * Browse, trade, and manage tokenized equities
 */

import { useState } from 'react';
import { EquitySector, SECTOR_COLORS } from '../types';

// Equities will be fetched from blockchain equity module when available
// Empty array until equity module API is exposed
const EQUITIES: Array<{
  symbol: string;
  name: string;
  price: number;
  change: number;
  marketCap: string;
  volume: string;
  sector: EquitySector;
}> = [];

interface SelectedEquity {
  symbol: string;
  name: string;
  price: number;
  change: number;
  marketCap: string;
  volume: string;
  sector: EquitySector;
}

export function MarketScreen() {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedSector, setSelectedSector] = useState<EquitySector | 'ALL'>('ALL');
  const [selectedEquity, setSelectedEquity] = useState<SelectedEquity | null>(null);
  const [tradeType, setTradeType] = useState<'buy' | 'sell'>('buy');
  const [orderType, setOrderType] = useState<'market' | 'limit'>('market');
  const [amount, setAmount] = useState('');
  const [limitPrice, setLimitPrice] = useState('');
  const tg = window.Telegram?.WebApp;

  const sectors = ['ALL', ...Object.values(EquitySector)] as const;

  const filteredEquities = EQUITIES.filter(equity => {
    const matchesSearch = equity.symbol.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          equity.name.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesSector = selectedSector === 'ALL' || equity.sector === selectedSector;
    return matchesSearch && matchesSector;
  });

  const handleTrade = () => {
    if (!selectedEquity || !amount) return;

    const shares = parseFloat(amount);
    const price = orderType === 'limit' ? parseFloat(limitPrice) : selectedEquity.price;
    const total = shares * price;

    tg?.showConfirm(
      `${tradeType === 'buy' ? 'Buy' : 'Sell'} ${shares} shares of ${selectedEquity.symbol} for $${total.toFixed(2)}?`,
      (confirmed) => {
        if (confirmed) {
          tg?.HapticFeedback?.notificationOccurred('success');
          tg?.showAlert(`Order placed successfully!\n${tradeType === 'buy' ? 'Bought' : 'Sold'} ${shares} ${selectedEquity.symbol}`);
          setSelectedEquity(null);
          setAmount('');
          setLimitPrice('');
        }
      }
    );
  };

  const formatSectorName = (sector: string): string => {
    if (sector === 'ALL') return 'All';
    return sector.charAt(0) + sector.slice(1).toLowerCase().replace('_', ' ');
  };

  return (
    <div className="market-screen">
      {/* Header */}
      <div className="market-header">
        <h1 className="title">Equity Market</h1>
        <p className="subtitle">Trade tokenized stocks 24/7</p>
      </div>

      {/* Search */}
      <div className="search-container">
        <div className="search-box">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="11" cy="11" r="8" />
            <path d="M21 21l-4.35-4.35" />
          </svg>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search stocks..."
          />
        </div>
      </div>

      {/* Sector Filter */}
      <div className="sector-filter">
        {sectors.map((sector) => (
          <button
            key={sector}
            onClick={() => {
              tg?.HapticFeedback?.selectionChanged();
              setSelectedSector(sector);
            }}
            className={`sector-btn ${selectedSector === sector ? 'active' : ''}`}
            style={selectedSector === sector && sector !== 'ALL'
              ? { background: `${SECTOR_COLORS[sector as EquitySector]}20`, borderColor: SECTOR_COLORS[sector as EquitySector] }
              : {}
            }
          >
            {formatSectorName(sector)}
          </button>
        ))}
      </div>

      {/* Market Stats - Only show when equities are available */}
      {EQUITIES.length > 0 && (
        <div className="market-stats">
          <div className="stat-card gainer">
            <span className="stat-label">Top Gainer</span>
            <span className="stat-symbol">-</span>
            <span className="stat-value positive">-</span>
          </div>
          <div className="stat-card loser">
            <span className="stat-label">Top Loser</span>
            <span className="stat-symbol">-</span>
            <span className="stat-value negative">-</span>
          </div>
          <div className="stat-card volume">
            <span className="stat-label">24h Volume</span>
            <span className="stat-symbol">$0</span>
            <span className="stat-value">Traded</span>
          </div>
        </div>
      )}

      {/* Equity List */}
      <div className="equity-list">
        {filteredEquities.map((equity) => {
          const isPositive = equity.change >= 0;
          const sectorColor = SECTOR_COLORS[equity.sector];

          return (
            <button
              key={equity.symbol}
              className="equity-card"
              onClick={() => {
                tg?.HapticFeedback?.impactOccurred('light');
                setSelectedEquity(equity);
              }}
            >
              <div className="equity-icon" style={{ background: `${sectorColor}20` }}>
                <span style={{ color: sectorColor }}>{equity.symbol.slice(0, 2)}</span>
              </div>
              <div className="equity-info">
                <div className="equity-name-row">
                  <span className="symbol">{equity.symbol}</span>
                  <span className="sector-tag" style={{ background: `${sectorColor}15`, color: sectorColor }}>
                    {equity.sector.slice(0, 4)}
                  </span>
                </div>
                <span className="name">{equity.name}</span>
              </div>
              <div className="equity-price">
                <span className="price">${equity.price.toFixed(2)}</span>
                <span className={`change ${isPositive ? 'positive' : 'negative'}`}>
                  {isPositive ? '+' : ''}{equity.change.toFixed(2)}%
                </span>
              </div>
            </button>
          );
        })}

        {filteredEquities.length === 0 && (
          <div className="empty-state">
            <span className="empty-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" width="48" height="48">
                <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
                <polyline points="17 6 23 6 23 12" />
              </svg>
            </span>
            <p className="empty-title">No Equities Available</p>
            <p className="empty-desc">Tokenized equities will appear here when companies are onboarded</p>
          </div>
        )}
      </div>

      {/* Trade Modal */}
      {selectedEquity && (
        <div className="modal-overlay" onClick={() => setSelectedEquity(null)}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            {/* Modal Header */}
            <div className="modal-header">
              <div className="equity-preview">
                <div
                  className="equity-icon-lg"
                  style={{ background: `${SECTOR_COLORS[selectedEquity.sector]}20` }}
                >
                  <span style={{ color: SECTOR_COLORS[selectedEquity.sector] }}>
                    {selectedEquity.symbol.slice(0, 2)}
                  </span>
                </div>
                <div>
                  <h2>{selectedEquity.symbol}</h2>
                  <p>{selectedEquity.name}</p>
                </div>
              </div>
              <button className="close-btn" onClick={() => setSelectedEquity(null)}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M18 6L6 18M6 6l12 12" />
                </svg>
              </button>
            </div>

            {/* Price Info */}
            <div className="price-info">
              <div className="current-price">
                <span className="label">Current Price</span>
                <span className="value">${selectedEquity.price.toFixed(2)}</span>
              </div>
              <div className={`price-change ${selectedEquity.change >= 0 ? 'positive' : 'negative'}`}>
                <span>{selectedEquity.change >= 0 ? '+' : ''}{selectedEquity.change.toFixed(2)}%</span>
                <span className="period">24h</span>
              </div>
            </div>

            {/* Market Data */}
            <div className="market-data">
              <div className="data-item">
                <span className="label">Market Cap</span>
                <span className="value">${selectedEquity.marketCap}</span>
              </div>
              <div className="data-item">
                <span className="label">Volume</span>
                <span className="value">{selectedEquity.volume}</span>
              </div>
            </div>

            {/* Trade Type Toggle */}
            <div className="trade-toggle">
              <button
                className={`toggle-btn buy ${tradeType === 'buy' ? 'active' : ''}`}
                onClick={() => setTradeType('buy')}
              >
                Buy
              </button>
              <button
                className={`toggle-btn sell ${tradeType === 'sell' ? 'active' : ''}`}
                onClick={() => setTradeType('sell')}
              >
                Sell
              </button>
            </div>

            {/* Order Type */}
            <div className="order-type">
              <button
                className={`order-btn ${orderType === 'market' ? 'active' : ''}`}
                onClick={() => setOrderType('market')}
              >
                Market Order
              </button>
              <button
                className={`order-btn ${orderType === 'limit' ? 'active' : ''}`}
                onClick={() => setOrderType('limit')}
              >
                Limit Order
              </button>
            </div>

            {/* Amount Input */}
            <div className="input-group">
              <label>Shares</label>
              <input
                type="text"
                inputMode="decimal"
                placeholder="0"
                value={amount}
                onChange={e => setAmount(e.target.value)}
              />
            </div>

            {/* Limit Price Input */}
            {orderType === 'limit' && (
              <div className="input-group">
                <label>Limit Price</label>
                <div className="price-input">
                  <span className="prefix">$</span>
                  <input
                    type="text"
                    inputMode="decimal"
                    placeholder={selectedEquity.price.toFixed(2)}
                    value={limitPrice}
                    onChange={e => setLimitPrice(e.target.value)}
                  />
                </div>
              </div>
            )}

            {/* Order Summary */}
            {amount && (
              <div className="order-summary">
                <div className="summary-row">
                  <span>Shares</span>
                  <span>{parseFloat(amount)}</span>
                </div>
                <div className="summary-row">
                  <span>Price per share</span>
                  <span>${orderType === 'limit' && limitPrice ? parseFloat(limitPrice).toFixed(2) : selectedEquity.price.toFixed(2)}</span>
                </div>
                <div className="summary-row total">
                  <span>Estimated Total</span>
                  <span>
                    ${((parseFloat(amount) || 0) * (orderType === 'limit' && limitPrice ? parseFloat(limitPrice) : selectedEquity.price)).toFixed(2)}
                  </span>
                </div>
              </div>
            )}

            {/* Trade Button */}
            <button
              className={`trade-btn ${tradeType}`}
              onClick={handleTrade}
              disabled={!amount || parseFloat(amount) <= 0}
            >
              {tradeType === 'buy' ? 'Buy' : 'Sell'} {selectedEquity.symbol}
            </button>
          </div>
        </div>
      )}

      <style>{styles}</style>
    </div>
  );
}

const styles = `
  .market-screen {
    min-height: 100vh;
    padding-bottom: 100px;
  }

  .market-header {
    padding: 20px 16px;
  }

  .title {
    font-size: 24px;
    font-weight: 700;
    color: white;
    margin: 0 0 4px;
  }

  .subtitle {
    font-size: 14px;
    color: #8b949e;
    margin: 0;
  }

  .search-container {
    padding: 0 16px 16px;
  }

  .search-box {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: #161B22;
    border-radius: 12px;
  }

  .search-box svg {
    width: 20px;
    height: 20px;
    color: #8b949e;
    flex-shrink: 0;
  }

  .search-box input {
    flex: 1;
    background: transparent;
    border: none;
    color: white;
    font-size: 15px;
    outline: none;
  }

  .search-box input::placeholder {
    color: #8b949e;
  }

  .sector-filter {
    display: flex;
    gap: 8px;
    padding: 0 16px 16px;
    overflow-x: auto;
    scrollbar-width: none;
  }

  .sector-filter::-webkit-scrollbar {
    display: none;
  }

  .sector-btn {
    padding: 8px 14px;
    border-radius: 20px;
    border: 1px solid #30363d;
    background: transparent;
    color: #8b949e;
    font-size: 13px;
    font-weight: 500;
    white-space: nowrap;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .sector-btn.active {
    background: #1E40AF20;
    border-color: #1E40AF;
    color: #1E40AF;
  }

  .market-stats {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
    padding: 0 16px 16px;
  }

  .stat-card {
    display: flex;
    flex-direction: column;
    padding: 12px;
    background: #161B22;
    border-radius: 12px;
  }

  .stat-label {
    font-size: 11px;
    color: #8b949e;
  }

  .stat-symbol {
    font-size: 16px;
    font-weight: 700;
    color: white;
    margin: 4px 0;
  }

  .stat-value {
    font-size: 12px;
    font-weight: 500;
  }

  .stat-value.positive {
    color: #10b981;
  }

  .stat-value.negative {
    color: #f87171;
  }

  .equity-list {
    padding: 0 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .equity-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 14px;
    background: #161B22;
    border: none;
    border-radius: 14px;
    cursor: pointer;
    transition: all 0.2s ease;
    text-align: left;
    width: 100%;
  }

  .equity-card:active {
    transform: scale(0.98);
  }

  .equity-icon {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 700;
  }

  .equity-info {
    flex: 1;
    min-width: 0;
  }

  .equity-name-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .symbol {
    font-size: 15px;
    font-weight: 600;
    color: white;
  }

  .sector-tag {
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .name {
    font-size: 13px;
    color: #8b949e;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: block;
  }

  .equity-price {
    text-align: right;
  }

  .price {
    display: block;
    font-size: 15px;
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
    color: #f87171;
  }

  .empty-state {
    padding: 60px 40px;
    text-align: center;
    color: #8b949e;
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
    margin: 0;
    line-height: 1.5;
  }

  /* Modal */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.85);
    display: flex;
    align-items: flex-end;
    z-index: 100;
  }

  .modal-content {
    width: 100%;
    max-height: 90vh;
    background: #161B22;
    border-radius: 20px 20px 0 0;
    padding: 20px;
    overflow-y: auto;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 20px;
  }

  .equity-preview {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .equity-icon-lg {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    font-weight: 700;
  }

  .equity-preview h2 {
    font-size: 20px;
    font-weight: 700;
    color: white;
    margin: 0;
  }

  .equity-preview p {
    font-size: 14px;
    color: #8b949e;
    margin: 0;
  }

  .close-btn {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    border: none;
    background: #30363d;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .close-btn svg {
    width: 18px;
    height: 18px;
    color: #8b949e;
  }

  .price-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    background: #0D1117;
    border-radius: 12px;
    margin-bottom: 16px;
  }

  .current-price .label {
    display: block;
    font-size: 12px;
    color: #8b949e;
  }

  .current-price .value {
    font-size: 28px;
    font-weight: 700;
    color: white;
  }

  .price-change {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    padding: 8px 12px;
    border-radius: 8px;
  }

  .price-change.positive {
    background: rgba(16, 185, 129, 0.1);
    color: #10b981;
  }

  .price-change.negative {
    background: rgba(248, 113, 113, 0.1);
    color: #f87171;
  }

  .price-change span:first-child {
    font-size: 16px;
    font-weight: 600;
  }

  .price-change .period {
    font-size: 11px;
    opacity: 0.7;
  }

  .market-data {
    display: flex;
    gap: 16px;
    margin-bottom: 20px;
  }

  .data-item {
    flex: 1;
    padding: 12px;
    background: #0D1117;
    border-radius: 10px;
  }

  .data-item .label {
    display: block;
    font-size: 12px;
    color: #8b949e;
    margin-bottom: 4px;
  }

  .data-item .value {
    font-size: 16px;
    font-weight: 600;
    color: white;
  }

  .trade-toggle {
    display: flex;
    background: #0D1117;
    border-radius: 10px;
    padding: 4px;
    margin-bottom: 16px;
  }

  .toggle-btn {
    flex: 1;
    padding: 12px;
    border: none;
    border-radius: 8px;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    background: transparent;
    color: #8b949e;
  }

  .toggle-btn.buy.active {
    background: #10b981;
    color: white;
  }

  .toggle-btn.sell.active {
    background: #f87171;
    color: white;
  }

  .order-type {
    display: flex;
    gap: 8px;
    margin-bottom: 16px;
  }

  .order-btn {
    flex: 1;
    padding: 10px;
    border: 1px solid #30363d;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    background: transparent;
    color: #8b949e;
    transition: all 0.2s ease;
  }

  .order-btn.active {
    background: #1E40AF20;
    border-color: #1E40AF;
    color: #1E40AF;
  }

  .input-group {
    margin-bottom: 16px;
  }

  .input-group label {
    display: block;
    font-size: 13px;
    color: #8b949e;
    margin-bottom: 8px;
  }

  .input-group input {
    width: 100%;
    padding: 14px 16px;
    background: #0D1117;
    border: 1px solid #30363d;
    border-radius: 10px;
    font-size: 18px;
    font-weight: 600;
    color: white;
    outline: none;
  }

  .input-group input:focus {
    border-color: #1E40AF;
  }

  .price-input {
    position: relative;
  }

  .price-input .prefix {
    position: absolute;
    left: 16px;
    top: 50%;
    transform: translateY(-50%);
    color: #8b949e;
    font-size: 18px;
    font-weight: 600;
  }

  .price-input input {
    padding-left: 32px;
  }

  .order-summary {
    padding: 16px;
    background: #0D1117;
    border-radius: 12px;
    margin-bottom: 16px;
  }

  .summary-row {
    display: flex;
    justify-content: space-between;
    padding: 8px 0;
    color: #8b949e;
    font-size: 14px;
  }

  .summary-row.total {
    border-top: 1px solid #30363d;
    margin-top: 8px;
    padding-top: 16px;
    color: white;
    font-size: 16px;
    font-weight: 600;
  }

  .trade-btn {
    width: 100%;
    padding: 16px;
    border: none;
    border-radius: 12px;
    font-size: 16px;
    font-weight: 600;
    color: white;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .trade-btn.buy {
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  }

  .trade-btn.sell {
    background: linear-gradient(135deg, #f87171 0%, #dc2626 100%);
  }

  .trade-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;
