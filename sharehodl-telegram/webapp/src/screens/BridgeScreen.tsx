/**
 * Bridge Screen - Convert crypto to HODL tokens
 */

import { useState } from 'react';
import { ArrowDown, Info, Clock, Shield, ChevronDown } from 'lucide-react';
import { Chain, CHAIN_CONFIGS } from '../types';

// Supported bridge assets
const BRIDGE_ASSETS = [
  { chain: Chain.ETHEREUM, rate: 3450 },
  { chain: Chain.BITCOIN, rate: 67500 },
  { chain: Chain.COSMOS, rate: 8.50 },
  { chain: Chain.OSMOSIS, rate: 0.85 },
  { chain: Chain.POLYGON, rate: 0.58 },
];

export function BridgeScreen() {
  const [fromChain, setFromChain] = useState<Chain>(Chain.ETHEREUM);
  const [amount, setAmount] = useState('');
  const [showFromPicker, setShowFromPicker] = useState(false);
  const tg = window.Telegram?.WebApp;

  const fromConfig = CHAIN_CONFIGS[fromChain];
  const rate = BRIDGE_ASSETS.find(a => a.chain === fromChain)?.rate || 0;
  const hodlAmount = parseFloat(amount || '0') * rate;
  const fee = hodlAmount * 0.001; // 0.1% fee
  const receiveAmount = hodlAmount - fee;

  const handleBridge = () => {
    if (!amount || parseFloat(amount) <= 0) {
      tg?.showAlert('Please enter an amount');
      return;
    }

    tg?.HapticFeedback?.impactOccurred('heavy');
    tg?.showConfirm(
      `Bridge ${amount} ${fromConfig.symbol} to ${receiveAmount.toFixed(2)} HODL?`,
      (confirmed) => {
        if (confirmed) {
          tg?.showAlert('Bridge transaction submitted!');
          setAmount('');
        }
      }
    );
  };

  return (
    <div className="flex flex-col min-h-screen bg-dark-bg p-4">
      <h1 className="text-xl font-bold text-white mb-2">Bridge to HODL</h1>
      <p className="text-gray-400 text-sm mb-6">
        Convert your crypto to HODL tokens for equity trading
      </p>

      {/* From */}
      <div className="mb-2">
        <label className="text-sm text-gray-400 mb-2 block">From</label>
        <div className="p-4 bg-dark-card rounded-xl">
          <div className="flex items-center justify-between mb-3">
            <button
              onClick={() => setShowFromPicker(true)}
              className="flex items-center gap-2 p-2 bg-dark-surface rounded-lg"
            >
              <div
                className="w-6 h-6 rounded-full flex items-center justify-center"
                style={{ backgroundColor: `${fromConfig.color}20` }}
              >
                <span className="text-xs font-bold" style={{ color: fromConfig.color }}>
                  {fromConfig.symbol.slice(0, 1)}
                </span>
              </div>
              <span className="text-white font-medium">{fromConfig.symbol}</span>
              <ChevronDown size={16} className="text-gray-500" />
            </button>
            <span className="text-gray-500 text-sm">Balance: 0.5 {fromConfig.symbol}</span>
          </div>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="0.00"
            className="w-full bg-transparent text-2xl text-white outline-none"
          />
          <p className="text-gray-500 text-sm mt-1">
            ~${(parseFloat(amount || '0') * rate).toLocaleString()}
          </p>
        </div>
      </div>

      {/* Arrow */}
      <div className="flex justify-center my-2">
        <div className="w-10 h-10 rounded-full bg-dark-card flex items-center justify-center">
          <ArrowDown className="text-gray-400" size={20} />
        </div>
      </div>

      {/* To */}
      <div className="mb-6">
        <label className="text-sm text-gray-400 mb-2 block">To</label>
        <div className="p-4 bg-dark-card rounded-xl">
          <div className="flex items-center gap-2 p-2 bg-dark-surface rounded-lg w-fit mb-3">
            <div className="w-6 h-6 rounded-full bg-primary/20 flex items-center justify-center">
              <span className="text-xs font-bold text-primary">H</span>
            </div>
            <span className="text-white font-medium">HODL</span>
          </div>
          <p className="text-2xl text-white font-medium">
            {receiveAmount > 0 ? receiveAmount.toFixed(2) : '0.00'}
          </p>
          <p className="text-gray-500 text-sm mt-1">ShareHODL Token</p>
        </div>
      </div>

      {/* Rate & Fee info */}
      <div className="space-y-3 mb-6">
        <div className="flex items-center justify-between p-3 bg-dark-card rounded-xl">
          <div className="flex items-center gap-2">
            <Info size={16} className="text-gray-500" />
            <span className="text-gray-400 text-sm">Exchange Rate</span>
          </div>
          <span className="text-white text-sm">
            1 {fromConfig.symbol} = {rate.toLocaleString()} HODL
          </span>
        </div>

        <div className="flex items-center justify-between p-3 bg-dark-card rounded-xl">
          <div className="flex items-center gap-2">
            <Shield size={16} className="text-gray-500" />
            <span className="text-gray-400 text-sm">Bridge Fee</span>
          </div>
          <span className="text-white text-sm">0.1% ({fee.toFixed(2)} HODL)</span>
        </div>

        <div className="flex items-center justify-between p-3 bg-dark-card rounded-xl">
          <div className="flex items-center gap-2">
            <Clock size={16} className="text-gray-500" />
            <span className="text-gray-400 text-sm">Estimated Time</span>
          </div>
          <span className="text-white text-sm">~5 minutes</span>
        </div>
      </div>

      {/* Bridge button */}
      <button
        onClick={handleBridge}
        disabled={!amount || parseFloat(amount) <= 0}
        className="w-full btn-primary mt-auto"
      >
        Bridge to HODL
      </button>

      {/* Info */}
      <p className="text-xs text-gray-500 text-center mt-4">
        HODL is pegged 1:1 to USD. Use it to trade tokenized equities.
      </p>

      {/* From chain picker */}
      {showFromPicker && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-end">
          <div className="w-full bg-dark-card rounded-t-3xl p-4 max-h-[60vh] overflow-y-auto animate-slide-up">
            <h3 className="text-white font-semibold mb-4">Select Asset</h3>
            <div className="space-y-2">
              {BRIDGE_ASSETS.map(({ chain, rate }) => {
                const config = CHAIN_CONFIGS[chain];
                return (
                  <button
                    key={chain}
                    onClick={() => {
                      setFromChain(chain);
                      setShowFromPicker(false);
                    }}
                    className={`w-full p-3 rounded-xl flex items-center gap-3 ${
                      fromChain === chain ? 'bg-primary/20' : 'bg-dark-surface'
                    }`}
                  >
                    <div
                      className="w-10 h-10 rounded-full flex items-center justify-center"
                      style={{ backgroundColor: `${config.color}20` }}
                    >
                      <span className="text-sm font-bold" style={{ color: config.color }}>
                        {config.symbol.slice(0, 2)}
                      </span>
                    </div>
                    <div className="flex-1 text-left">
                      <span className="text-white font-medium">{config.symbol}</span>
                      <p className="text-gray-500 text-sm">{config.name}</p>
                    </div>
                    <span className="text-gray-400 text-sm">
                      1 = {rate.toLocaleString()} HODL
                    </span>
                  </button>
                );
              })}
            </div>
            <button
              onClick={() => setShowFromPicker(false)}
              className="w-full mt-4 btn-secondary"
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
