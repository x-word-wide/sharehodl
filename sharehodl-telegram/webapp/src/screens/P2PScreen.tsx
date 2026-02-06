/**
 * P2P Trading Screen - Professional peer-to-peer marketplace
 * Supports multiple assets (HODL, USDT, BTC) and currencies (USD, NGN, GBP)
 */

import { useState, useMemo } from 'react';

// SVG Icons for assets
const AssetIcons: Record<string, React.ReactNode> = {
  HODL: (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M12 2L2 7l10 5 10-5-10-5z" />
      <path d="M2 17l10 5 10-5" />
      <path d="M2 12l10 5 10-5" />
    </svg>
  ),
  USDT: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 .67 3 1.5S13.66 8 12 8s-3-.67-3-1.5S10.34 5 12 5zm4 10H8v-2h3V9.5h2V13h3v2z" />
    </svg>
  ),
  BTC: (
    <svg viewBox="0 0 24 24" fill="currentColor">
      <path d="M11.5 11.5v-2c1.1 0 2 .9 2 2h-2zm0 1h2c0 1.1-.9 2-2 2v-2zm1-9V2h-1v1.5h-1V2h-1v1.5c-1.93.23-3.5 1.9-3.5 3.95 0 1.58.92 2.94 2.25 3.58-.48.57-.75 1.3-.75 2.02 0 2.05 1.57 3.72 3.5 3.95v1.5h1V17h1v1.5h1V17c1.93-.23 3.5-1.9 3.5-3.95 0-.72-.27-1.45-.75-2.02 1.33-.64 2.25-2 2.25-3.58 0-2.05-1.57-3.72-3.5-3.95V2h-1v1.5h-1zm-1 11c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5v5zm0-6V5c-1.38 0-2.5 1.12-2.5 2.5s1.12 2.5 2.5 2.5z" />
    </svg>
  ),
};

// Currency symbols
const CURRENCIES = [
  { code: 'USD', symbol: '$', name: 'US Dollar', flag: 'US' },
  { code: 'NGN', symbol: '₦', name: 'Nigerian Naira', flag: 'NG' },
  { code: 'GBP', symbol: '£', name: 'British Pound', flag: 'GB' },
];

// Supported assets
const ASSETS = [
  { code: 'HODL', name: 'ShareHODL', color: '#3B82F6' },
  { code: 'USDT', name: 'Tether USD', color: '#26A17B' },
  { code: 'BTC', name: 'Bitcoin', color: '#F7931A' },
];

// Payment methods icons (SVG)
const PaymentIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <rect x="1" y="4" width="22" height="16" rx="2" ry="2" />
    <line x1="1" y1="10" x2="23" y2="10" />
  </svg>
);

const BankIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <path d="M3 21h18" />
    <path d="M3 10h18" />
    <path d="M5 6l7-3 7 3" />
    <path d="M4 10v11" />
    <path d="M20 10v11" />
    <path d="M8 14v3" />
    <path d="M12 14v3" />
    <path d="M16 14v3" />
  </svg>
);

const TransferIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
    <line x1="12" y1="1" x2="12" y2="23" />
    <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
  </svg>
);

interface Trader {
  name: string;
  rating: number;
  trades: number;
  verified: boolean;
  online: boolean;
}

interface P2PListing {
  id: string;
  type: 'BUY' | 'SELL';
  trader: Trader;
  asset: string;
  currency: string;
  price: number;
  minAmount: number;
  maxAmount: number;
  available: number;
  paymentMethods: string[];
  avgReleaseTime: string;
}

// Exchange rates (demo - in production would come from API)
const EXCHANGE_RATES: Record<string, Record<string, number>> = {
  HODL: { USD: 1.0, NGN: 1550, GBP: 0.79 },
  USDT: { USD: 1.0, NGN: 1550, GBP: 0.79 },
  BTC: { USD: 43500, NGN: 67425000, GBP: 34365 },
};

// Demo listings with multiple assets and currencies
const DEMO_LISTINGS: P2PListing[] = [
  // HODL listings
  {
    id: '1',
    type: 'SELL',
    trader: { name: 'CryptoKing', rating: 4.9, trades: 523, verified: true, online: true },
    asset: 'HODL',
    currency: 'USD',
    price: 1.02,
    minAmount: 100,
    maxAmount: 10000,
    available: 5000,
    paymentMethods: ['Bank Transfer', 'PayPal'],
    avgReleaseTime: '~5 min'
  },
  {
    id: '2',
    type: 'SELL',
    trader: { name: 'NairaTrade', rating: 4.8, trades: 312, verified: true, online: true },
    asset: 'HODL',
    currency: 'NGN',
    price: 1580,
    minAmount: 50000,
    maxAmount: 5000000,
    available: 3500,
    paymentMethods: ['Bank Transfer', 'Opay', 'Palmpay'],
    avgReleaseTime: '~3 min'
  },
  {
    id: '3',
    type: 'SELL',
    trader: { name: 'LondonTrader', rating: 4.7, trades: 189, verified: true, online: false },
    asset: 'HODL',
    currency: 'GBP',
    price: 0.81,
    minAmount: 50,
    maxAmount: 5000,
    available: 2000,
    paymentMethods: ['Bank Transfer', 'Revolut'],
    avgReleaseTime: '~15 min'
  },
  // USDT listings
  {
    id: '4',
    type: 'SELL',
    trader: { name: 'StableDealer', rating: 4.95, trades: 892, verified: true, online: true },
    asset: 'USDT',
    currency: 'USD',
    price: 1.001,
    minAmount: 100,
    maxAmount: 50000,
    available: 25000,
    paymentMethods: ['Bank Transfer', 'Zelle', 'Wire'],
    avgReleaseTime: '~2 min'
  },
  {
    id: '5',
    type: 'SELL',
    trader: { name: 'NaijaUSDT', rating: 4.85, trades: 645, verified: true, online: true },
    asset: 'USDT',
    currency: 'NGN',
    price: 1560,
    minAmount: 100000,
    maxAmount: 10000000,
    available: 15000,
    paymentMethods: ['Bank Transfer', 'Kuda', 'GTBank'],
    avgReleaseTime: '~5 min'
  },
  {
    id: '6',
    type: 'SELL',
    trader: { name: 'UKCrypto', rating: 4.9, trades: 421, verified: true, online: true },
    asset: 'USDT',
    currency: 'GBP',
    price: 0.80,
    minAmount: 50,
    maxAmount: 20000,
    available: 10000,
    paymentMethods: ['Bank Transfer', 'Faster Payments'],
    avgReleaseTime: '~3 min'
  },
  // BTC listings
  {
    id: '7',
    type: 'SELL',
    trader: { name: 'BTCWhale', rating: 4.92, trades: 1205, verified: true, online: true },
    asset: 'BTC',
    currency: 'USD',
    price: 43750,
    minAmount: 500,
    maxAmount: 100000,
    available: 2.5,
    paymentMethods: ['Wire Transfer', 'Bank Transfer'],
    avgReleaseTime: '~10 min'
  },
  {
    id: '8',
    type: 'SELL',
    trader: { name: 'AbujaBTC', rating: 4.75, trades: 378, verified: true, online: true },
    asset: 'BTC',
    currency: 'NGN',
    price: 68000000,
    minAmount: 500000,
    maxAmount: 50000000,
    available: 0.8,
    paymentMethods: ['Bank Transfer', 'Opay'],
    avgReleaseTime: '~15 min'
  },
  {
    id: '9',
    type: 'SELL',
    trader: { name: 'LondonBTC', rating: 4.88, trades: 567, verified: true, online: false },
    asset: 'BTC',
    currency: 'GBP',
    price: 34500,
    minAmount: 100,
    maxAmount: 50000,
    available: 1.2,
    paymentMethods: ['Bank Transfer', 'Revolut', 'Monzo'],
    avgReleaseTime: '~8 min'
  },
  // Buy orders
  {
    id: '10',
    type: 'BUY',
    trader: { name: 'QuickBuyer', rating: 4.95, trades: 892, verified: true, online: true },
    asset: 'HODL',
    currency: 'USD',
    price: 0.98,
    minAmount: 200,
    maxAmount: 20000,
    available: 15000,
    paymentMethods: ['Bank Transfer', 'PayPal', 'Zelle'],
    avgReleaseTime: '~2 min'
  },
  {
    id: '11',
    type: 'BUY',
    trader: { name: 'LagosBuyer', rating: 4.8, trades: 445, verified: true, online: true },
    asset: 'USDT',
    currency: 'NGN',
    price: 1540,
    minAmount: 50000,
    maxAmount: 5000000,
    available: 20000,
    paymentMethods: ['Bank Transfer', 'Opay'],
    avgReleaseTime: '~5 min'
  },
  {
    id: '12',
    type: 'BUY',
    trader: { name: 'BTCCollector', rating: 4.9, trades: 728, verified: true, online: true },
    asset: 'BTC',
    currency: 'GBP',
    price: 34200,
    minAmount: 200,
    maxAmount: 30000,
    available: 0.5,
    paymentMethods: ['Bank Transfer', 'Faster Payments'],
    avgReleaseTime: '~5 min'
  },
];

export function P2PScreen() {
  const [tradeType, setTradeType] = useState<'BUY' | 'SELL'>('BUY');
  const [selectedAsset, setSelectedAsset] = useState<string>('HODL');
  const [selectedCurrency, setSelectedCurrency] = useState<string>('USD');
  const [showAssetSelector, setShowAssetSelector] = useState(false);
  const [showCurrencySelector, setShowCurrencySelector] = useState(false);
  const [selectedListing, setSelectedListing] = useState<P2PListing | null>(null);
  const [amount, setAmount] = useState('');
  const tg = window.Telegram?.WebApp;

  // Get current currency info
  const currentCurrency = CURRENCIES.find(c => c.code === selectedCurrency) || CURRENCIES[0];
  const currentAsset = ASSETS.find(a => a.code === selectedAsset) || ASSETS[0];

  // Filter listings based on selections
  const filteredListings = useMemo(() => {
    return DEMO_LISTINGS.filter(listing =>
      listing.type === (tradeType === 'BUY' ? 'SELL' : 'BUY') &&
      listing.asset === selectedAsset &&
      listing.currency === selectedCurrency
    );
  }, [tradeType, selectedAsset, selectedCurrency]);

  const handleTrade = (listing: P2PListing) => {
    tg?.HapticFeedback?.impactOccurred('medium');
    setSelectedListing(listing);
  };

  const handleConfirmTrade = () => {
    tg?.HapticFeedback?.notificationOccurred('success');
    tg?.showAlert(`Trade initiated with ${selectedListing?.trader.name}. Escrow activated.`);
    setSelectedListing(null);
    setAmount('');
  };

  const formatPrice = (price: number, currency: string) => {
    const curr = CURRENCIES.find(c => c.code === currency);
    if (currency === 'NGN') {
      return `${curr?.symbol}${price.toLocaleString()}`;
    }
    return `${curr?.symbol}${price.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  };

  const formatAmount = (amount: number, currency: string) => {
    const curr = CURRENCIES.find(c => c.code === currency);
    if (currency === 'NGN') {
      return `${curr?.symbol}${amount.toLocaleString()}`;
    }
    return `${curr?.symbol}${amount.toLocaleString()}`;
  };

  return (
    <div className="p2p-screen">
      {/* Header */}
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

      {/* Asset & Currency Selectors */}
      <div className="selector-row">
        <button
          className="selector-btn"
          onClick={() => {
            tg?.HapticFeedback?.selectionChanged();
            setShowAssetSelector(true);
          }}
        >
          <div className="selector-icon" style={{ background: `${currentAsset.color}20`, color: currentAsset.color }}>
            {AssetIcons[selectedAsset]}
          </div>
          <div className="selector-info">
            <span className="selector-label">Asset</span>
            <span className="selector-value">{selectedAsset}</span>
          </div>
          <svg className="chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="6 9 12 15 18 9" />
          </svg>
        </button>

        <button
          className="selector-btn"
          onClick={() => {
            tg?.HapticFeedback?.selectionChanged();
            setShowCurrencySelector(true);
          }}
        >
          <div className="selector-flag">{currentCurrency.flag}</div>
          <div className="selector-info">
            <span className="selector-label">Currency</span>
            <span className="selector-value">{currentCurrency.code}</span>
          </div>
          <svg className="chevron" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="6 9 12 15 18 9" />
          </svg>
        </button>
      </div>

      {/* Trade type toggle */}
      <div className="trade-toggle">
        <button
          className={`toggle-btn ${tradeType === 'BUY' ? 'active buy' : ''}`}
          onClick={() => {
            tg?.HapticFeedback?.selectionChanged();
            setTradeType('BUY');
          }}
        >
          <svg className="toggle-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 19V5" />
            <path d="M5 12l7-7 7 7" />
          </svg>
          Buy {selectedAsset}
        </button>
        <button
          className={`toggle-btn ${tradeType === 'SELL' ? 'active sell' : ''}`}
          onClick={() => {
            tg?.HapticFeedback?.selectionChanged();
            setTradeType('SELL');
          }}
        >
          <svg className="toggle-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 5v14" />
            <path d="M19 12l-7 7-7-7" />
          </svg>
          Sell {selectedAsset}
        </button>
      </div>

      {/* Market rate info */}
      <div className="market-rate-card">
        <div className="rate-row">
          <span className="rate-label">Market Rate</span>
          <span className="rate-value">
            1 {selectedAsset} = {formatPrice(EXCHANGE_RATES[selectedAsset][selectedCurrency], selectedCurrency)}
          </span>
        </div>
      </div>

      {/* Listings */}
      <div className="listings">
        {filteredListings.length === 0 ? (
          <div className="empty-state">
            <svg className="empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <circle cx="11" cy="11" r="8" />
              <path d="M21 21l-4.35-4.35" />
            </svg>
            <p className="empty-text">No listings found</p>
            <p className="empty-hint">Try selecting a different asset or currency</p>
          </div>
        ) : (
          filteredListings.map((listing) => (
            <div key={listing.id} className="listing-card" onClick={() => handleTrade(listing)}>
              {/* Trader info */}
              <div className="trader-row">
                <div className="trader-avatar">
                  <span className="avatar-text">
                    {listing.trader.name.slice(0, 2).toUpperCase()}
                  </span>
                  {listing.trader.online && <span className="online-dot" />}
                </div>
                <div className="trader-info">
                  <div className="trader-name-row">
                    <span className="trader-name">{listing.trader.name}</span>
                    {listing.trader.verified && (
                      <svg className="verified-icon" viewBox="0 0 24 24" fill="#10b981">
                        <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    )}
                  </div>
                  <div className="trader-stats">
                    <span className="rating">
                      <svg className="star-icon" viewBox="0 0 24 24" fill="#f59e0b">
                        <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                      </svg>
                      {listing.trader.rating}
                    </span>
                    <span className="divider">|</span>
                    <span className="trades">{listing.trader.trades} trades</span>
                    <span className="divider">|</span>
                    <span className="release-time">{listing.avgReleaseTime}</span>
                  </div>
                </div>
              </div>

              {/* Price and amount */}
              <div className="price-row">
                <div className="price-col">
                  <span className="price-label">Price</span>
                  <span className="price-value">{formatPrice(listing.price, listing.currency)}</span>
                </div>
                <div className="amount-col">
                  <span className="price-label">Available</span>
                  <span className="price-value">{listing.available.toLocaleString()} {listing.asset}</span>
                </div>
              </div>

              {/* Limits */}
              <div className="limits-row">
                <span className="limits-label">Limit</span>
                <span className="limits-value">
                  {formatAmount(listing.minAmount, listing.currency)} - {formatAmount(listing.maxAmount, listing.currency)}
                </span>
              </div>

              {/* Payment methods */}
              <div className="payment-row">
                {listing.paymentMethods.map((method) => (
                  <span key={method} className="payment-tag">
                    {method.includes('Bank') ? <BankIcon /> : method.includes('Transfer') ? <TransferIcon /> : <PaymentIcon />}
                    {method}
                  </span>
                ))}
              </div>

              {/* Trade button */}
              <button className={`trade-btn ${tradeType === 'BUY' ? 'buy' : 'sell'}`}>
                {tradeType === 'BUY' ? 'Buy' : 'Sell'} {listing.asset}
              </button>
            </div>
          ))
        )}
      </div>

      {/* Create ad button */}
      <button className="create-ad-btn" onClick={() => tg?.showAlert('Create listing coming soon!')}>
        <svg className="plus-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <line x1="12" y1="5" x2="12" y2="19" />
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
        Post an Ad
      </button>

      {/* Asset Selector Modal */}
      {showAssetSelector && (
        <div className="modal-overlay" onClick={() => setShowAssetSelector(false)}>
          <div className="selector-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">Select Asset</h3>
              <button className="modal-close" onClick={() => setShowAssetSelector(false)}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>
            <div className="selector-options">
              {ASSETS.map((asset) => (
                <button
                  key={asset.code}
                  className={`selector-option ${selectedAsset === asset.code ? 'selected' : ''}`}
                  onClick={() => {
                    tg?.HapticFeedback?.selectionChanged();
                    setSelectedAsset(asset.code);
                    setShowAssetSelector(false);
                  }}
                >
                  <div className="option-icon" style={{ background: `${asset.color}20`, color: asset.color }}>
                    {AssetIcons[asset.code]}
                  </div>
                  <div className="option-info">
                    <span className="option-name">{asset.code}</span>
                    <span className="option-full">{asset.name}</span>
                  </div>
                  {selectedAsset === asset.code && (
                    <svg className="check-icon" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2">
                      <polyline points="20 6 9 17 4 12" />
                    </svg>
                  )}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Currency Selector Modal */}
      {showCurrencySelector && (
        <div className="modal-overlay" onClick={() => setShowCurrencySelector(false)}>
          <div className="selector-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3 className="modal-title">Select Currency</h3>
              <button className="modal-close" onClick={() => setShowCurrencySelector(false)}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>
            <div className="selector-options">
              {CURRENCIES.map((currency) => (
                <button
                  key={currency.code}
                  className={`selector-option ${selectedCurrency === currency.code ? 'selected' : ''}`}
                  onClick={() => {
                    tg?.HapticFeedback?.selectionChanged();
                    setSelectedCurrency(currency.code);
                    setShowCurrencySelector(false);
                  }}
                >
                  <div className="option-flag">{currency.flag}</div>
                  <div className="option-info">
                    <span className="option-name">{currency.code}</span>
                    <span className="option-full">{currency.name}</span>
                  </div>
                  <span className="option-symbol">{currency.symbol}</span>
                  {selectedCurrency === currency.code && (
                    <svg className="check-icon" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2">
                      <polyline points="20 6 9 17 4 12" />
                    </svg>
                  )}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Trade modal */}
      {selectedListing && (
        <div className="modal-overlay" onClick={() => setSelectedListing(null)}>
          <div className="trade-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">
                {tradeType === 'BUY' ? 'Buy' : 'Sell'} {selectedListing.asset}
              </h2>
              <button className="modal-close" onClick={() => setSelectedListing(null)}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              </button>
            </div>

            {/* Trader summary */}
            <div className="modal-trader">
              <div className="trader-avatar large">
                <span className="avatar-text">
                  {selectedListing.trader.name.slice(0, 2).toUpperCase()}
                </span>
              </div>
              <div className="modal-trader-info">
                <span className="trader-name">{selectedListing.trader.name}</span>
                <span className="trader-stats">
                  <svg className="star-icon" viewBox="0 0 24 24" fill="#f59e0b">
                    <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                  </svg>
                  {selectedListing.trader.rating} | {selectedListing.trader.trades} trades
                </span>
              </div>
            </div>

            {/* Amount input */}
            <div className="modal-input-group">
              <label className="input-label">Amount ({selectedListing.currency})</label>
              <div className="input-wrapper">
                <span className="input-prefix">{currentCurrency.symbol}</span>
                <input
                  type="number"
                  className="modal-input"
                  placeholder="0.00"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                />
                <button className="max-btn" onClick={() => setAmount(selectedListing.maxAmount.toString())}>
                  MAX
                </button>
              </div>
              <span className="input-hint">
                Limit: {formatAmount(selectedListing.minAmount, selectedListing.currency)} - {formatAmount(selectedListing.maxAmount, selectedListing.currency)}
              </span>
            </div>

            {/* You will receive */}
            <div className="receive-summary">
              <span className="receive-label">You will {tradeType === 'BUY' ? 'receive' : 'send'}</span>
              <span className="receive-value">
                {amount ? (parseFloat(amount) / selectedListing.price).toFixed(selectedListing.asset === 'BTC' ? 8 : 2) : '0.00'} {selectedListing.asset}
              </span>
            </div>

            {/* Payment methods */}
            <div className="modal-payment">
              <span className="payment-label">Payment Methods</span>
              <div className="payment-tags">
                {selectedListing.paymentMethods.map((method) => (
                  <span key={method} className="payment-tag">
                    {method.includes('Bank') ? <BankIcon /> : method.includes('Transfer') ? <TransferIcon /> : <PaymentIcon />}
                    {method}
                  </span>
                ))}
              </div>
            </div>

            {/* Escrow notice */}
            <div className="escrow-notice">
              <svg className="shield-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
              </svg>
              <span>Funds will be held in escrow until trade is complete</span>
            </div>

            {/* Confirm button */}
            <button
              className={`confirm-btn ${tradeType === 'BUY' ? 'buy' : 'sell'}`}
              onClick={handleConfirmTrade}
              disabled={!amount || parseFloat(amount) < selectedListing.minAmount}
            >
              {tradeType === 'BUY' ? 'Buy' : 'Sell'} {selectedListing.asset}
            </button>
          </div>
        </div>
      )}

      <style>{`
        .p2p-screen {
          min-height: 100vh;
          padding: 16px;
          padding-bottom: 100px;
        }

        .p2p-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
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

        /* Selector Row */
        .selector-row {
          display: flex;
          gap: 10px;
          margin-bottom: 16px;
        }

        .selector-btn {
          flex: 1;
          display: flex;
          align-items: center;
          gap: 10px;
          padding: 12px;
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(10px);
          -webkit-backdrop-filter: blur(10px);
          border: 1px solid rgba(48, 54, 61, 0.5);
          border-radius: 12px;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .selector-btn:active {
          transform: scale(0.98);
          background: rgba(22, 27, 34, 0.8);
        }

        .selector-icon {
          width: 36px;
          height: 36px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .selector-icon svg {
          width: 20px;
          height: 20px;
        }

        .selector-flag {
          width: 36px;
          height: 36px;
          border-radius: 50%;
          background: rgba(48, 54, 61, 0.5);
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 18px;
        }

        .selector-info {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: flex-start;
        }

        .selector-label {
          font-size: 11px;
          color: #8b949e;
        }

        .selector-value {
          font-size: 15px;
          font-weight: 600;
          color: white;
        }

        .chevron {
          width: 16px;
          height: 16px;
          color: #8b949e;
        }

        /* Trade Toggle */
        .trade-toggle {
          display: flex;
          gap: 8px;
          padding: 4px;
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(10px);
          -webkit-backdrop-filter: blur(10px);
          border: 1px solid rgba(48, 54, 61, 0.4);
          border-radius: 14px;
          margin-bottom: 16px;
        }

        .toggle-btn {
          flex: 1;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          padding: 12px;
          border: none;
          border-radius: 10px;
          font-size: 15px;
          font-weight: 600;
          color: #8b949e;
          background: transparent;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .toggle-btn.active.buy {
          background: linear-gradient(135deg, #10b981 0%, #059669 100%);
          color: white;
        }

        .toggle-btn.active.sell {
          background: linear-gradient(135deg, #f87171 0%, #dc2626 100%);
          color: white;
        }

        .toggle-icon {
          width: 18px;
          height: 18px;
        }

        /* Market Rate Card */
        .market-rate-card {
          padding: 14px 16px;
          background: rgba(30, 64, 175, 0.1);
          backdrop-filter: blur(8px);
          -webkit-backdrop-filter: blur(8px);
          border: 1px solid rgba(30, 64, 175, 0.2);
          border-radius: 12px;
          margin-bottom: 20px;
        }

        .rate-row {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .rate-label {
          font-size: 13px;
          color: #8b949e;
        }

        .rate-value {
          font-size: 14px;
          font-weight: 600;
          color: white;
        }

        /* Empty State */
        .empty-state {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 48px 24px;
          text-align: center;
        }

        .empty-icon {
          width: 64px;
          height: 64px;
          color: #30363d;
          margin-bottom: 16px;
        }

        .empty-text {
          font-size: 16px;
          font-weight: 600;
          color: #8b949e;
          margin: 0 0 8px;
        }

        .empty-hint {
          font-size: 13px;
          color: #484f58;
          margin: 0;
        }

        /* Listings */
        .listings {
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .listing-card {
          padding: 16px;
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(12px);
          -webkit-backdrop-filter: blur(12px);
          border: 1px solid rgba(48, 54, 61, 0.5);
          border-radius: 16px;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .listing-card:active {
          transform: scale(0.98);
          background: rgba(22, 27, 34, 0.8);
        }

        .trader-row {
          display: flex;
          gap: 12px;
          margin-bottom: 16px;
        }

        .trader-avatar {
          position: relative;
          width: 44px;
          height: 44px;
          border-radius: 50%;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .trader-avatar.large {
          width: 56px;
          height: 56px;
        }

        .avatar-text {
          font-size: 16px;
          font-weight: 600;
          color: white;
        }

        .online-dot {
          position: absolute;
          bottom: 2px;
          right: 2px;
          width: 10px;
          height: 10px;
          border-radius: 50%;
          background: #10b981;
          border: 2px solid #161B22;
        }

        .trader-info {
          flex: 1;
        }

        .trader-name-row {
          display: flex;
          align-items: center;
          gap: 6px;
          margin-bottom: 4px;
        }

        .trader-name {
          font-size: 15px;
          font-weight: 600;
          color: white;
        }

        .verified-icon {
          width: 16px;
          height: 16px;
        }

        .trader-stats {
          display: flex;
          align-items: center;
          gap: 6px;
          font-size: 13px;
          color: #8b949e;
        }

        .star-icon {
          width: 12px;
          height: 12px;
        }

        .rating {
          display: flex;
          align-items: center;
          gap: 3px;
        }

        .divider {
          color: #30363d;
        }

        .price-row {
          display: flex;
          gap: 24px;
          margin-bottom: 12px;
        }

        .price-col, .amount-col {
          display: flex;
          flex-direction: column;
          gap: 4px;
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

        .limits-row {
          display: flex;
          gap: 8px;
          margin-bottom: 12px;
          font-size: 13px;
        }

        .limits-label {
          color: #8b949e;
        }

        .limits-value {
          color: white;
        }

        .payment-row {
          display: flex;
          flex-wrap: wrap;
          gap: 8px;
          margin-bottom: 16px;
        }

        .payment-tag {
          display: flex;
          align-items: center;
          gap: 5px;
          padding: 6px 10px;
          background: rgba(48, 54, 61, 0.5);
          border-radius: 8px;
          font-size: 12px;
          color: #8b949e;
        }

        .trade-btn {
          width: 100%;
          padding: 14px;
          border: none;
          border-radius: 12px;
          font-size: 15px;
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

        .trade-btn:active {
          transform: scale(0.98);
        }

        .create-ad-btn {
          position: fixed;
          bottom: 90px;
          left: 16px;
          right: 16px;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          padding: 16px;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          border: none;
          border-radius: 14px;
          font-size: 16px;
          font-weight: 600;
          color: white;
          cursor: pointer;
          box-shadow: 0 4px 20px rgba(30, 64, 175, 0.3);
        }

        .plus-icon {
          width: 20px;
          height: 20px;
        }

        /* Modal */
        .modal-overlay {
          position: fixed;
          inset: 0;
          background: rgba(0, 0, 0, 0.8);
          display: flex;
          align-items: flex-end;
          z-index: 100;
          animation: fadeIn 0.2s ease;
        }

        @keyframes fadeIn {
          from { opacity: 0; }
          to { opacity: 1; }
        }

        .selector-modal, .trade-modal {
          width: 100%;
          max-height: 90vh;
          padding: 24px;
          background: #161B22;
          border-radius: 24px 24px 0 0;
          animation: slideUp 0.3s ease;
          overflow-y: auto;
        }

        @keyframes slideUp {
          from { transform: translateY(100%); }
          to { transform: translateY(0); }
        }

        .modal-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 24px;
        }

        .modal-title {
          font-size: 20px;
          font-weight: 700;
          color: white;
          margin: 0;
        }

        .modal-close {
          width: 32px;
          height: 32px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: rgba(48, 54, 61, 0.5);
          border: none;
          border-radius: 50%;
          cursor: pointer;
        }

        .modal-close svg {
          width: 18px;
          height: 18px;
          color: #8b949e;
        }

        /* Selector Options */
        .selector-options {
          display: flex;
          flex-direction: column;
          gap: 8px;
        }

        .selector-option {
          display: flex;
          align-items: center;
          gap: 14px;
          padding: 14px;
          background: rgba(22, 27, 34, 0.6);
          border: 1px solid rgba(48, 54, 61, 0.5);
          border-radius: 14px;
          cursor: pointer;
          transition: all 0.2s ease;
          width: 100%;
        }

        .selector-option.selected {
          background: rgba(30, 64, 175, 0.15);
          border-color: rgba(30, 64, 175, 0.5);
        }

        .selector-option:active {
          transform: scale(0.98);
        }

        .option-icon {
          width: 44px;
          height: 44px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .option-icon svg {
          width: 24px;
          height: 24px;
        }

        .option-flag {
          width: 44px;
          height: 44px;
          border-radius: 50%;
          background: rgba(48, 54, 61, 0.5);
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 24px;
        }

        .option-info {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: flex-start;
        }

        .option-name {
          font-size: 16px;
          font-weight: 600;
          color: white;
        }

        .option-full {
          font-size: 13px;
          color: #8b949e;
        }

        .option-symbol {
          font-size: 18px;
          font-weight: 600;
          color: #8b949e;
        }

        .check-icon {
          width: 20px;
          height: 20px;
        }

        /* Trade Modal */
        .modal-trader {
          display: flex;
          align-items: center;
          gap: 16px;
          padding: 16px;
          background: #0D1117;
          border-radius: 14px;
          margin-bottom: 24px;
        }

        .modal-trader-info {
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .modal-input-group {
          margin-bottom: 20px;
        }

        .input-label {
          display: block;
          font-size: 14px;
          color: #8b949e;
          margin-bottom: 8px;
        }

        .input-wrapper {
          display: flex;
          align-items: center;
          background: #0D1117;
          border: 1px solid #30363d;
          border-radius: 12px;
          overflow: hidden;
        }

        .input-prefix {
          padding: 0 12px;
          font-size: 18px;
          color: #8b949e;
        }

        .modal-input {
          flex: 1;
          padding: 14px 0;
          background: transparent;
          border: none;
          font-size: 18px;
          color: white;
          outline: none;
        }

        .max-btn {
          padding: 8px 16px;
          margin: 6px;
          background: rgba(30, 64, 175, 0.2);
          border: none;
          border-radius: 8px;
          font-size: 12px;
          font-weight: 600;
          color: #3B82F6;
          cursor: pointer;
        }

        .input-hint {
          display: block;
          font-size: 12px;
          color: #8b949e;
          margin-top: 8px;
        }

        .receive-summary {
          display: flex;
          justify-content: space-between;
          padding: 16px;
          background: #0D1117;
          border-radius: 12px;
          margin-bottom: 20px;
        }

        .receive-label {
          font-size: 14px;
          color: #8b949e;
        }

        .receive-value {
          font-size: 18px;
          font-weight: 600;
          color: white;
        }

        .modal-payment {
          margin-bottom: 20px;
        }

        .payment-label {
          display: block;
          font-size: 14px;
          color: #8b949e;
          margin-bottom: 12px;
        }

        .payment-tags {
          display: flex;
          flex-wrap: wrap;
          gap: 8px;
        }

        .escrow-notice {
          display: flex;
          align-items: center;
          gap: 10px;
          padding: 14px;
          background: rgba(16, 185, 129, 0.1);
          border-radius: 12px;
          margin-bottom: 24px;
          font-size: 13px;
          color: #10b981;
        }

        .confirm-btn {
          width: 100%;
          padding: 16px;
          border: none;
          border-radius: 14px;
          font-size: 16px;
          font-weight: 600;
          color: white;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .confirm-btn.buy {
          background: linear-gradient(135deg, #10b981 0%, #059669 100%);
        }

        .confirm-btn.sell {
          background: linear-gradient(135deg, #f87171 0%, #dc2626 100%);
        }

        .confirm-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }
      `}</style>
    </div>
  );
}
