/**
 * Send Screen - Send crypto to another address
 */

import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Scan, ChevronDown, AlertCircle } from 'lucide-react';
import { useWalletStore } from '../services/walletStore';
import { Chain, CHAIN_CONFIGS } from '../types';

export function SendScreen() {
  const { chain: chainParam } = useParams();
  const navigate = useNavigate();
  const { accounts } = useWalletStore();
  const tg = window.Telegram?.WebApp;

  const [selectedChain, setSelectedChain] = useState<Chain>(
    chainParam ? (chainParam as Chain) : Chain.SHAREHODL
  );
  const [recipient, setRecipient] = useState('');
  const [amount, setAmount] = useState('');
  const [showChainPicker, setShowChainPicker] = useState(false);

  const account = accounts.find(a => a.chain === selectedChain);
  const config = CHAIN_CONFIGS[selectedChain];
  const balance = parseFloat(account?.balance || '0');

  const handleSend = () => {
    if (!recipient || !amount) {
      tg?.showAlert('Please fill in all fields');
      return;
    }

    if (parseFloat(amount) > balance) {
      tg?.showAlert('Insufficient balance');
      return;
    }

    tg?.HapticFeedback?.impactOccurred('heavy');
    tg?.showConfirm(
      `Send ${amount} ${config.symbol} to ${recipient.slice(0, 10)}...?`,
      (confirmed) => {
        if (confirmed) {
          tg?.showAlert('Transaction submitted!');
          navigate('/portfolio');
        }
      }
    );
  };

  const handleMax = () => {
    // Leave some for gas
    const maxAmount = selectedChain === Chain.SHAREHODL
      ? balance
      : Math.max(0, balance - 0.01);
    setAmount(maxAmount.toString());
  };

  return (
    <div className="flex flex-col min-h-screen bg-dark-bg p-4">
      <h1 className="text-xl font-bold text-white mb-6">Send</h1>

      {/* Chain selector */}
      <div className="mb-4">
        <label className="text-sm text-gray-400 mb-2 block">Asset</label>
        <button
          onClick={() => setShowChainPicker(true)}
          className="w-full input flex items-center justify-between"
        >
          <div className="flex items-center gap-3">
            <div
              className="w-8 h-8 rounded-full flex items-center justify-center"
              style={{ backgroundColor: `${config.color}20` }}
            >
              <span className="text-xs font-bold" style={{ color: config.color }}>
                {config.symbol.slice(0, 2)}
              </span>
            </div>
            <div>
              <span className="text-white">{config.symbol}</span>
              <span className="text-gray-500 ml-2 text-sm">{config.name}</span>
            </div>
          </div>
          <ChevronDown className="text-gray-500" size={20} />
        </button>
      </div>

      {/* Balance */}
      <div className="p-3 bg-dark-card rounded-xl mb-4">
        <div className="flex items-center justify-between">
          <span className="text-gray-400 text-sm">Available</span>
          <span className="text-white font-medium">
            {balance.toLocaleString()} {config.symbol}
          </span>
        </div>
      </div>

      {/* Recipient */}
      <div className="mb-4">
        <label className="text-sm text-gray-400 mb-2 block">Recipient Address</label>
        <div className="relative">
          <input
            type="text"
            value={recipient}
            onChange={(e) => setRecipient(e.target.value)}
            placeholder={`Enter ${config.name} address`}
            className="input pr-12"
          />
          <button
            onClick={() => tg?.showAlert('QR scanner coming soon')}
            className="absolute right-3 top-1/2 -translate-y-1/2"
          >
            <Scan className="text-gray-500" size={20} />
          </button>
        </div>
      </div>

      {/* Amount */}
      <div className="mb-4">
        <label className="text-sm text-gray-400 mb-2 block">Amount</label>
        <div className="relative">
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="0.00"
            className="input text-xl pr-20"
          />
          <button
            onClick={handleMax}
            className="absolute right-3 top-1/2 -translate-y-1/2 px-2 py-1 bg-primary/20 text-primary text-sm rounded"
          >
            MAX
          </button>
        </div>
      </div>

      {/* Warning for low balance */}
      {parseFloat(amount) > balance && (
        <div className="flex items-center gap-2 p-3 bg-accent-red/10 rounded-xl mb-4">
          <AlertCircle className="text-accent-red" size={18} />
          <span className="text-accent-red text-sm">Insufficient balance</span>
        </div>
      )}

      {/* Network fee */}
      <div className="p-3 bg-dark-card rounded-xl mb-6">
        <div className="flex items-center justify-between">
          <span className="text-gray-400 text-sm">Network Fee</span>
          <span className="text-white text-sm">~0.001 {config.symbol}</span>
        </div>
      </div>

      {/* Send button */}
      <div className="mt-auto">
        <button
          onClick={handleSend}
          disabled={!recipient || !amount || parseFloat(amount) > balance}
          className="w-full btn-primary"
        >
          Send {config.symbol}
        </button>
      </div>

      {/* Chain picker modal */}
      {showChainPicker && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-end">
          <div className="w-full bg-dark-card rounded-t-3xl p-4 max-h-[70vh] overflow-y-auto animate-slide-up">
            <h3 className="text-white font-semibold mb-4">Select Asset</h3>
            <div className="space-y-2">
              {accounts.map((acc) => {
                const cfg = CHAIN_CONFIGS[acc.chain];
                return (
                  <button
                    key={acc.chain}
                    onClick={() => {
                      setSelectedChain(acc.chain);
                      setShowChainPicker(false);
                    }}
                    className={`w-full p-3 rounded-xl flex items-center gap-3 ${
                      selectedChain === acc.chain ? 'bg-primary/20' : 'bg-dark-surface'
                    }`}
                  >
                    <div
                      className="w-10 h-10 rounded-full flex items-center justify-center"
                      style={{ backgroundColor: `${cfg.color}20` }}
                    >
                      <span className="text-sm font-bold" style={{ color: cfg.color }}>
                        {cfg.symbol.slice(0, 2)}
                      </span>
                    </div>
                    <div className="flex-1 text-left">
                      <span className="text-white font-medium">{cfg.symbol}</span>
                      <p className="text-gray-500 text-sm">{acc.balance}</p>
                    </div>
                  </button>
                );
              })}
            </div>
            <button
              onClick={() => setShowChainPicker(false)}
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
