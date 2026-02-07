/**
 * Send Screen - Send crypto to another address
 *
 * SECURITY: PIN is required before signing transactions
 */

import { useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Scan, ChevronDown, AlertCircle, CheckCircle, Loader2, ExternalLink } from 'lucide-react';
import { useWalletStore } from '../services/walletStore';
import { Chain, CHAIN_CONFIGS } from '../types';
import { sendTokens, validateAddress, type TransactionResult } from '../services/blockchainService';
import { BottomSheet } from '../components/BottomSheet';

type SendStep = 'form' | 'pin' | 'sending' | 'success' | 'error';
const PIN_LENGTH = 6;

export function SendScreen() {
  const { chain: chainParam } = useParams();
  const navigate = useNavigate();
  const { accounts, getMnemonicForSigning } = useWalletStore();
  const tg = window.Telegram?.WebApp;

  const [selectedChain, setSelectedChain] = useState<Chain>(
    chainParam ? (chainParam as Chain) : Chain.SHAREHODL
  );
  const [recipient, setRecipient] = useState('');
  const [amount, setAmount] = useState('');
  const [memo, setMemo] = useState('');
  const [showChainPicker, setShowChainPicker] = useState(false);
  const [addressError, setAddressError] = useState<string | null>(null);

  // Transaction state
  const [step, setStep] = useState<SendStep>('form');
  const [pin, setPin] = useState('');
  const [txResult, setTxResult] = useState<TransactionResult | null>(null);
  const [shake, setShake] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);

  const account = accounts.find(a => a.chain === selectedChain);
  const config = CHAIN_CONFIGS[selectedChain];
  const balance = parseFloat(account?.balance || '0');

  // Validate address on change
  const handleRecipientChange = (value: string) => {
    setRecipient(value);
    setAddressError(null);

    if (value && selectedChain === Chain.SHAREHODL) {
      const validation = validateAddress(value);
      if (!validation.valid) {
        setAddressError(validation.error || 'Invalid address');
      }
    }
  };

  // Initiate send - show PIN dialog
  const handleSend = () => {
    if (!recipient || !amount) {
      tg?.showAlert('Please fill in all fields');
      return;
    }

    // Validate address for ShareHODL
    if (selectedChain === Chain.SHAREHODL) {
      const validation = validateAddress(recipient);
      if (!validation.valid) {
        setAddressError(validation.error || 'Invalid address');
        tg?.showAlert(validation.error || 'Invalid recipient address');
        return;
      }
    }

    if (parseFloat(amount) > balance) {
      tg?.showAlert('Insufficient balance');
      return;
    }

    // Show PIN confirmation
    tg?.HapticFeedback?.impactOccurred('medium');
    setStep('pin');
  };

  // Handle PIN entry
  const handleKeyPress = useCallback(async (key: string) => {
    if (isProcessing) return;
    tg?.HapticFeedback?.impactOccurred('light');

    if (key === 'delete') {
      setPin(prev => prev.slice(0, -1));
      return;
    }

    if (pin.length >= PIN_LENGTH) return;

    const newPin = pin + key;
    setPin(newPin);

    // Auto-submit when PIN is complete
    if (newPin.length === PIN_LENGTH) {
      await processTransaction(newPin);
    }
  }, [pin, isProcessing]);

  // Process the transaction
  const processTransaction = async (enteredPin: string) => {
    setIsProcessing(true);
    setStep('sending');
    tg?.HapticFeedback?.impactOccurred('heavy');

    try {
      // Get mnemonic for signing
      const mnemonic = await getMnemonicForSigning(enteredPin);

      // Send the transaction
      const result = await sendTokens(
        mnemonic,
        recipient,
        amount,
        memo || undefined
      );

      setTxResult(result);

      if (result.success) {
        tg?.HapticFeedback?.notificationOccurred('success');
        setStep('success');
      } else {
        tg?.HapticFeedback?.notificationOccurred('error');
        setStep('error');
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Transaction failed';

      // Check if it's a PIN error
      if (errorMessage.includes('PIN') || errorMessage.includes('Decryption')) {
        tg?.HapticFeedback?.notificationOccurred('error');
        setShake(true);
        setTimeout(() => {
          setShake(false);
          setPin('');
          setStep('pin');
        }, 500);
        setIsProcessing(false);
        return;
      }

      setTxResult({ success: false, error: errorMessage });
      tg?.HapticFeedback?.notificationOccurred('error');
      setStep('error');
    }

    setIsProcessing(false);
  };

  const handleMax = () => {
    // Leave some for gas (0.01 HODL for fees)
    const fee = selectedChain === Chain.SHAREHODL ? 0.01 : 0.01;
    const maxAmount = Math.max(0, balance - fee);
    setAmount(maxAmount.toString());
  };

  const handleBack = () => {
    if (step === 'pin') {
      setStep('form');
      setPin('');
    } else if (step === 'error') {
      setStep('form');
      setPin('');
      setTxResult(null);
    }
  };

  // Main form view
  if (step === 'form') {
    return (
      <BottomSheet title="Send" fullHeight onClose={() => navigate(-1)}>
        <div className="flex flex-col p-4 pb-8">

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
              onChange={(e) => handleRecipientChange(e.target.value)}
              placeholder={`Enter ${config.name} address`}
              className={`input pr-12 ${addressError ? 'border-accent-red' : ''}`}
            />
            <button
              onClick={() => tg?.showAlert('QR scanner coming soon')}
              className="absolute right-3 top-1/2 -translate-y-1/2"
            >
              <Scan className="text-gray-500" size={20} />
            </button>
          </div>
          {addressError && (
            <p className="text-accent-red text-sm mt-1">{addressError}</p>
          )}
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

        {/* Memo (optional) */}
        <div className="mb-4">
          <label className="text-sm text-gray-400 mb-2 block">Memo (optional)</label>
          <input
            type="text"
            value={memo}
            onChange={(e) => setMemo(e.target.value)}
            placeholder="Add a note"
            className="input"
          />
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
            <span className="text-gray-400 text-sm">Estimated Fee</span>
            <span className="text-white text-sm">~0.001 {config.symbol}</span>
          </div>
        </div>

        {/* Send button */}
        <div className="mt-auto">
          <button
            onClick={handleSend}
            disabled={!recipient || !amount || parseFloat(amount) > balance || !!addressError}
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
                        setRecipient('');
                        setAddressError(null);
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
      </BottomSheet>
    );
  }

  // PIN entry view
  if (step === 'pin') {
    return (
      <BottomSheet title="Confirm" fullHeight onClose={handleBack}>
        <div className="flex flex-col items-center justify-center p-4 pt-8">
          <div className="w-16 h-16 rounded-full bg-primary/20 flex items-center justify-center mb-4">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="w-8 h-8 text-primary">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
              <path d="M7 11V7a5 5 0 0 1 10 0v4" />
            </svg>
          </div>

          <h2 className="text-xl font-bold text-white mb-2">Confirm Transaction</h2>
          <p className="text-gray-400 text-center mb-6">
            Enter your PIN to send {amount} {config.symbol}
          </p>

          {/* Transaction summary */}
          <div className="w-full max-w-sm bg-dark-card rounded-xl p-4 mb-6">
            <div className="flex justify-between mb-2">
              <span className="text-gray-400">To</span>
              <span className="text-white font-mono text-sm">
                {recipient.slice(0, 12)}...{recipient.slice(-6)}
              </span>
            </div>
            <div className="flex justify-between mb-2">
              <span className="text-gray-400">Amount</span>
              <span className="text-white">{amount} {config.symbol}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-gray-400">Fee</span>
              <span className="text-white">~0.001 {config.symbol}</span>
            </div>
          </div>

          {/* PIN dots */}
          <div className={`flex gap-4 mb-8 ${shake ? 'animate-shake' : ''}`}>
            {Array.from({ length: PIN_LENGTH }).map((_, i) => (
              <div
                key={i}
                className={`w-4 h-4 rounded-full transition-all ${
                  i < pin.length ? 'bg-primary scale-110' : 'bg-gray-600'
                }`}
              />
            ))}
          </div>

          {/* Numpad */}
          <div className="grid grid-cols-3 gap-4 w-full max-w-[280px]">
            {['1', '2', '3', '4', '5', '6', '7', '8', '9', '', '0', 'delete'].map((key) =>
              key === '' ? (
                <div key="empty" className="w-[72px] h-[72px]" />
              ) : (
                <button
                  key={key}
                  onClick={() => handleKeyPress(key)}
                  disabled={isProcessing}
                  className="w-[72px] h-[72px] rounded-full bg-dark-card flex items-center justify-center text-white text-2xl transition-all active:bg-primary/20"
                >
                  {key === 'delete' ? '⌫' : key}
                </button>
              )
            )}
          </div>

          <style>{`
            @keyframes shake {
              0%, 100% { transform: translateX(0); }
              20%, 60% { transform: translateX(-10px); }
              40%, 80% { transform: translateX(10px); }
            }
            .animate-shake { animation: shake 0.5s ease-in-out; }
          `}</style>
        </div>
      </BottomSheet>
    );
  }

  // Sending view
  if (step === 'sending') {
    return (
      <BottomSheet fullHeight>
        <div className="flex flex-col items-center justify-center p-4 pt-16">
          <div className="w-20 h-20 rounded-full bg-primary/20 flex items-center justify-center mb-6">
            <Loader2 className="w-10 h-10 text-primary animate-spin" />
          </div>
          <h2 className="text-xl font-bold text-white mb-2">Sending Transaction</h2>
          <p className="text-gray-400 text-center">
            Please wait while your transaction is being processed...
          </p>
        </div>
      </BottomSheet>
    );
  }

  // Success view
  if (step === 'success' && txResult) {
    return (
      <BottomSheet fullHeight onClose={() => navigate('/portfolio')}>
        <div className="flex flex-col items-center justify-center p-4 pt-12">
          <div className="w-20 h-20 rounded-full bg-green-500/20 flex items-center justify-center mb-6">
            <CheckCircle className="w-10 h-10 text-green-500" />
          </div>

          <h2 className="text-xl font-bold text-white mb-2">Transaction Sent!</h2>
          <p className="text-gray-400 text-center mb-6">
            Your {amount} {config.symbol} has been sent successfully.
          </p>

        {/* Transaction details */}
        <div className="w-full max-w-sm bg-dark-card rounded-xl p-4 mb-6">
          <div className="flex justify-between mb-2">
            <span className="text-gray-400">To</span>
            <span className="text-white font-mono text-sm">
              {recipient.slice(0, 12)}...{recipient.slice(-6)}
            </span>
          </div>
          <div className="flex justify-between mb-2">
            <span className="text-gray-400">Amount</span>
            <span className="text-white">{amount} {config.symbol}</span>
          </div>
          {txResult.txHash && (
            <div className="flex justify-between items-center">
              <span className="text-gray-400">Tx Hash</span>
              <a
                href={`${config.explorerUrl}/tx/${txResult.txHash}`}
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1 text-primary text-sm"
              >
                {txResult.txHash.slice(0, 8)}...
                <ExternalLink size={14} />
              </a>
            </div>
          )}
        </div>

          <button
            onClick={() => navigate('/portfolio')}
            className="w-full max-w-sm btn-primary"
          >
            Done
          </button>
        </div>
      </BottomSheet>
    );
  }

  // Error view
  if (step === 'error') {
    return (
      <BottomSheet fullHeight onClose={() => navigate('/portfolio')}>
        <div className="flex flex-col items-center justify-center p-4 pt-12">
          <div className="w-20 h-20 rounded-full bg-red-500/20 flex items-center justify-center mb-6">
            <AlertCircle className="w-10 h-10 text-red-500" />
          </div>

          <h2 className="text-xl font-bold text-white mb-2">Transaction Failed</h2>
          <p className="text-gray-400 text-center mb-4">
            {txResult?.error || 'An error occurred while sending your transaction.'}
          </p>

          <div className="flex gap-4 w-full max-w-sm">
            <button
              onClick={handleBack}
              className="flex-1 btn-secondary"
            >
              Try Again
            </button>
            <button
              onClick={() => navigate('/portfolio')}
              className="flex-1 btn-primary"
            >
              Done
            </button>
          </div>
        </div>
      </BottomSheet>
    );
  }

  return null;
}
