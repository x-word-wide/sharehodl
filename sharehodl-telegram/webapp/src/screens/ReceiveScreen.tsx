/**
 * Receive Screen - Show QR code and address for receiving
 */

import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { Copy, Check, ChevronDown, Share2 } from 'lucide-react';
import { QRCodeSVG } from 'qrcode.react';
import { useWalletStore } from '../services/walletStore';
import { Chain, CHAIN_CONFIGS } from '../types';

export function ReceiveScreen() {
  const { chain: chainParam } = useParams();
  const { accounts } = useWalletStore();
  const tg = window.Telegram?.WebApp;

  const [selectedChain, setSelectedChain] = useState<Chain>(
    chainParam ? (chainParam as Chain) : Chain.SHAREHODL
  );
  const [copied, setCopied] = useState(false);
  const [showChainPicker, setShowChainPicker] = useState(false);

  const account = accounts.find(a => a.chain === selectedChain);
  const config = CHAIN_CONFIGS[selectedChain];
  const address = account?.address || '';

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(address);
      setCopied(true);
      tg?.HapticFeedback?.notificationOccurred('success');
      setTimeout(() => setCopied(false), 2000);
    } catch {
      tg?.showAlert('Failed to copy');
    }
  };

  const handleShare = () => {
    if (navigator.share) {
      navigator.share({
        title: `My ${config.symbol} Address`,
        text: address
      });
    } else {
      handleCopy();
    }
  };

  return (
    <div className="flex flex-col min-h-screen bg-dark-bg p-4">
      <h1 className="text-xl font-bold text-white mb-6">Receive</h1>

      {/* Chain selector */}
      <button
        onClick={() => setShowChainPicker(true)}
        className="w-full input flex items-center justify-between mb-6"
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

      {/* QR Code */}
      <div className="bg-white p-6 rounded-2xl mb-6 mx-auto">
        <QRCodeSVG
          value={address}
          size={200}
          level="H"
          includeMargin={false}
          bgColor="#FFFFFF"
          fgColor="#000000"
        />
      </div>

      {/* Address */}
      <div className="bg-dark-card rounded-xl p-4 mb-4">
        <p className="text-gray-400 text-sm mb-2">Your {config.symbol} Address</p>
        <p className="text-white font-mono text-sm break-all">{address}</p>
      </div>

      {/* Actions */}
      <div className="flex gap-3">
        <button
          onClick={handleCopy}
          className="flex-1 btn-primary flex items-center justify-center gap-2"
        >
          {copied ? <Check size={18} /> : <Copy size={18} />}
          {copied ? 'Copied!' : 'Copy Address'}
        </button>
        <button
          onClick={handleShare}
          className="btn-secondary px-4"
        >
          <Share2 size={18} />
        </button>
      </div>

      {/* Warning */}
      <div className="mt-6 p-4 bg-accent-orange/10 border border-accent-orange/30 rounded-xl">
        <p className="text-accent-orange text-sm">
          Only send {config.symbol} to this address. Sending other tokens may result in permanent loss.
        </p>
      </div>

      {/* Chain picker modal */}
      {showChainPicker && (
        <div className="fixed inset-0 bg-black/60 z-50 flex items-end">
          <div className="w-full bg-dark-card rounded-t-3xl p-4 max-h-[70vh] overflow-y-auto animate-slide-up">
            <h3 className="text-white font-semibold mb-4">Select Network</h3>
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
                      <span className="text-white font-medium">{cfg.name}</span>
                      <p className="text-gray-500 text-sm">{cfg.symbol}</p>
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
