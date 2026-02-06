/**
 * Trade Screen - Buy/Sell equities
 */

import { useState } from 'react';
import { Info } from 'lucide-react';

export function TradeScreen() {
  const [tradeType, setTradeType] = useState<'buy' | 'sell'>('buy');
  const [orderType, setOrderType] = useState<'market' | 'limit'>('market');
  const [symbol, setSymbol] = useState('AAPL');
  const [amount, setAmount] = useState('');
  const [limitPrice, setLimitPrice] = useState('');
  const tg = window.Telegram?.WebApp;

  const currentPrice = 189.45;
  const estimatedTotal = parseFloat(amount || '0') * (orderType === 'limit' ? parseFloat(limitPrice || '0') : currentPrice);

  const handleTrade = () => {
    tg?.HapticFeedback?.impactOccurred('heavy');
    tg?.showAlert(`${tradeType === 'buy' ? 'Buy' : 'Sell'} order placed for ${amount} ${symbol}`);
  };

  return (
    <div className="flex flex-col min-h-screen bg-dark-bg p-4">
      {/* Header */}
      <h1 className="text-xl font-bold text-white mb-4">Trade</h1>

      {/* Buy/Sell Toggle */}
      <div className="flex gap-2 p-1 bg-dark-surface rounded-xl mb-6">
        <button
          onClick={() => setTradeType('buy')}
          className={`flex-1 py-3 rounded-lg font-semibold transition-colors ${
            tradeType === 'buy'
              ? 'bg-accent-green text-white'
              : 'text-gray-500'
          }`}
        >
          Buy
        </button>
        <button
          onClick={() => setTradeType('sell')}
          className={`flex-1 py-3 rounded-lg font-semibold transition-colors ${
            tradeType === 'sell'
              ? 'bg-accent-red text-white'
              : 'text-gray-500'
          }`}
        >
          Sell
        </button>
      </div>

      {/* Symbol selector */}
      <div className="mb-4">
        <label className="text-sm text-gray-400 mb-2 block">Stock</label>
        <select
          value={symbol}
          onChange={(e) => setSymbol(e.target.value)}
          className="input"
        >
          <option value="AAPL">AAPL - Apple Inc.</option>
          <option value="GOOGL">GOOGL - Alphabet Inc.</option>
          <option value="MSFT">MSFT - Microsoft Corp.</option>
          <option value="AMZN">AMZN - Amazon.com Inc.</option>
          <option value="NVDA">NVDA - NVIDIA Corp.</option>
        </select>
      </div>

      {/* Current price */}
      <div className="p-4 bg-dark-card rounded-xl mb-4">
        <div className="flex items-center justify-between">
          <span className="text-gray-400">Current Price</span>
          <span className="text-white font-semibold text-xl">${currentPrice}</span>
        </div>
      </div>

      {/* Order type */}
      <div className="mb-4">
        <label className="text-sm text-gray-400 mb-2 block">Order Type</label>
        <div className="flex gap-2">
          <button
            onClick={() => setOrderType('market')}
            className={`flex-1 py-2 rounded-lg text-sm font-medium ${
              orderType === 'market'
                ? 'bg-primary text-white'
                : 'bg-dark-surface text-gray-400'
            }`}
          >
            Market
          </button>
          <button
            onClick={() => setOrderType('limit')}
            className={`flex-1 py-2 rounded-lg text-sm font-medium ${
              orderType === 'limit'
                ? 'bg-primary text-white'
                : 'bg-dark-surface text-gray-400'
            }`}
          >
            Limit
          </button>
        </div>
      </div>

      {/* Limit price */}
      {orderType === 'limit' && (
        <div className="mb-4">
          <label className="text-sm text-gray-400 mb-2 block">Limit Price</label>
          <input
            type="number"
            value={limitPrice}
            onChange={(e) => setLimitPrice(e.target.value)}
            placeholder="0.00"
            className="input"
          />
        </div>
      )}

      {/* Amount */}
      <div className="mb-4">
        <label className="text-sm text-gray-400 mb-2 block">Shares</label>
        <input
          type="number"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="0"
          className="input text-xl"
        />
      </div>

      {/* Quick amounts */}
      <div className="flex gap-2 mb-6">
        {['10', '50', '100', '500'].map((val) => (
          <button
            key={val}
            onClick={() => setAmount(val)}
            className="flex-1 py-2 bg-dark-surface rounded-lg text-gray-400 text-sm"
          >
            {val}
          </button>
        ))}
      </div>

      {/* Estimated total */}
      <div className="p-4 bg-dark-card rounded-xl mb-6">
        <div className="flex items-center justify-between mb-2">
          <span className="text-gray-400">Estimated Total</span>
          <span className="text-white font-semibold text-xl">
            ${estimatedTotal.toLocaleString('en-US', { minimumFractionDigits: 2 })}
          </span>
        </div>
        <div className="flex items-center gap-1 text-xs text-gray-500">
          <Info size={12} />
          <span>Plus ~0.3% trading fee</span>
        </div>
      </div>

      {/* Trade button */}
      <button
        onClick={handleTrade}
        disabled={!amount || parseFloat(amount) <= 0}
        className={`w-full py-4 rounded-xl font-semibold text-white transition-colors ${
          tradeType === 'buy'
            ? 'bg-accent-green disabled:bg-accent-green/50'
            : 'bg-accent-red disabled:bg-accent-red/50'
        }`}
      >
        {tradeType === 'buy' ? 'Buy' : 'Sell'} {symbol}
      </button>
    </div>
  );
}
