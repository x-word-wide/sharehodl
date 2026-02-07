/**
 * Send Screen - Premium Send Flow
 *
 * SECURITY: PIN is required before signing transactions
 */

import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useWalletStore } from '../services/walletStore';
import { Chain, CHAIN_CONFIGS } from '../types';
import { sendTokens, validateAddress, type TransactionResult } from '../services/blockchainService';
import { BottomSheet } from '../components/BottomSheet';
import { TransactionConfirmation, TransactionDetails } from '../components/TransactionConfirmation';
import { QRScanner } from '../components/QRScanner';

type SendStep = 'form' | 'confirm' | 'success' | 'error';

export function SendScreen() {
  const { chain: chainParam } = useParams();
  const navigate = useNavigate();
  const { accounts, getMnemonicForSigning, _cachedPin } = useWalletStore();
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
  const [txResult, setTxResult] = useState<TransactionResult | null>(null);
  const [pendingTransaction, setPendingTransaction] = useState<TransactionDetails | null>(null);
  const [showQRScanner, setShowQRScanner] = useState(false);

  const account = accounts.find(a => a.chain === selectedChain);
  const config = CHAIN_CONFIGS[selectedChain];
  const balance = parseFloat(account?.balance || '0');

  // Validate address on change
  const handleRecipientChange = (value: string) => {
    const cleanedValue = value.trim().replace(/[\n\r]/g, '');
    setRecipient(cleanedValue);
    setAddressError(null);

    if (cleanedValue && selectedChain === Chain.SHAREHODL) {
      const validation = validateAddress(cleanedValue);
      if (!validation.valid) {
        setAddressError(validation.error || 'Invalid address');
      }
    }
  };

  // Handle QR scan result
  const handleQRScan = (data: string) => {
    setShowQRScanner(false);
    const cleanedAddress = data.trim().replace(/[\n\r]/g, '');
    handleRecipientChange(cleanedAddress);
    tg?.HapticFeedback?.notificationOccurred('success');
  };

  // Initiate send - show confirmation
  const handleSend = () => {
    if (!recipient || !amount) {
      tg?.showAlert('Please fill in all fields');
      return;
    }

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

    const transaction: TransactionDetails = {
      type: 'send',
      title: `Send ${config.symbol}`,
      amount: amount,
      token: config.symbol,
      recipient: recipient,
      fee: `~0.001 ${config.symbol}`,
      details: memo ? [{ label: 'Memo', value: memo }] : undefined,
    };

    setPendingTransaction(transaction);
    setStep('confirm');
    tg?.HapticFeedback?.impactOccurred('medium');
  };

  // Process the transaction after confirmation
  const handleTransactionConfirm = async (mnemonic: string) => {
    const result = await sendTokens(
      mnemonic,
      recipient,
      amount,
      memo || undefined
    );

    setTxResult(result);

    if (result.success) {
      setStep('success');
    } else {
      throw new Error(result.error || 'Transaction failed');
    }
  };

  const handleTransactionCancel = () => {
    setPendingTransaction(null);
    setStep('form');
  };

  const handleMax = () => {
    const fee = selectedChain === Chain.SHAREHODL ? 0.01 : 0.01;
    const maxAmount = Math.max(0, balance - fee);
    setAmount(maxAmount.toString());
  };

  // SECURITY: Validate and sanitize amount input
  // Only allows digits and at most one decimal point, with max 6 decimal places
  const handleAmountChange = (value: string) => {
    // Remove any characters that aren't digits or decimal point
    let sanitized = value.replace(/[^\d.]/g, '');

    // Ensure only one decimal point
    const parts = sanitized.split('.');
    if (parts.length > 2) {
      sanitized = parts[0] + '.' + parts.slice(1).join('');
    }

    // Limit decimal places to 6 (micro units precision)
    if (parts.length === 2 && parts[1].length > 6) {
      sanitized = parts[0] + '.' + parts[1].slice(0, 6);
    }

    // Prevent unreasonably large values (max 999,999,999)
    const numValue = parseFloat(sanitized);
    if (!isNaN(numValue) && numValue > 999999999) {
      return; // Don't update if value is too large
    }

    setAmount(sanitized);
  };

  // Main form view
  if (step === 'form') {
    return (
      <BottomSheet title="Send" fullHeight onClose={() => navigate(-1)}>
        <div className="send-screen">
          {/* Asset selector */}
          <div className="send-section">
            <label className="send-label">Asset</label>
            <button
              onClick={() => setShowChainPicker(true)}
              className="send-asset-picker"
            >
              <div
                className="send-asset-icon"
                style={{ backgroundColor: `${config.color}20` }}
              >
                <span style={{ color: config.color }}>{config.symbol.slice(0, 2)}</span>
              </div>
              <div className="send-asset-info">
                <span className="send-asset-symbol">{config.symbol}</span>
                <span className="send-asset-name">{config.name}</span>
              </div>
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="send-chevron">
                <path d="M6 9l6 6 6-6" />
              </svg>
            </button>
          </div>

          {/* Balance display */}
          <div className="send-balance-card">
            <span className="send-balance-label">Available Balance</span>
            <span className="send-balance-value">{balance.toLocaleString()} {config.symbol}</span>
          </div>

          {/* Recipient section */}
          <div className="send-section">
            <label className="send-label">Recipient Address</label>

            {recipient ? (
              <div className={`send-address-display ${addressError ? 'has-error' : ''}`}>
                <span className="send-address-text">
                  {recipient.slice(0, 14)}...{recipient.slice(-10)}
                </span>
                <button
                  className="send-address-clear"
                  onClick={() => { setRecipient(''); setAddressError(null); }}
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M18 6L6 18M6 6l12 12" />
                  </svg>
                </button>
              </div>
            ) : (
              <div className="send-address-empty">
                <span>Paste or scan recipient address</span>
              </div>
            )}

            {/* Address action buttons */}
            <div className="send-address-actions">
              <button
                className="send-action-btn"
                onClick={async () => {
                  try {
                    const text = await navigator.clipboard.readText();
                    if (text) {
                      handleRecipientChange(text);
                      tg?.HapticFeedback?.impactOccurred('light');
                    }
                  } catch {
                    tg?.showAlert('Unable to access clipboard');
                  }
                }}
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="9" y="9" width="13" height="13" rx="2" />
                  <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
                </svg>
                <span>Paste</span>
              </button>
              <button
                className="send-action-btn"
                onClick={() => setShowQRScanner(true)}
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="3" y="3" width="7" height="7" rx="1" />
                  <rect x="14" y="3" width="7" height="7" rx="1" />
                  <rect x="3" y="14" width="7" height="7" rx="1" />
                  <rect x="14" y="14" width="3" height="3" />
                  <rect x="18" y="14" width="3" height="3" />
                  <rect x="14" y="18" width="3" height="3" />
                  <rect x="18" y="18" width="3" height="3" />
                </svg>
                <span>Scan QR</span>
              </button>
            </div>

            {addressError && (
              <div className="send-error-msg">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <circle cx="12" cy="12" r="10" />
                  <path d="M12 8v4M12 16h.01" />
                </svg>
                <span>{addressError}</span>
              </div>
            )}
          </div>

          {/* Amount section */}
          <div className="send-section">
            <label className="send-label">Amount</label>
            <div className="send-amount-wrap">
              <input
                type="text"
                inputMode="decimal"
                value={amount}
                onChange={(e) => handleAmountChange(e.target.value)}
                placeholder="0.00"
                className="send-amount-input"
              />
              <button className="send-max-btn" onClick={handleMax}>
                MAX
              </button>
            </div>
          </div>

          {/* Memo section */}
          <div className="send-section">
            <label className="send-label">Memo (optional)</label>
            <input
              type="text"
              value={memo}
              onChange={(e) => setMemo(e.target.value)}
              placeholder="Add a note"
              className="send-input"
            />
          </div>

          {/* Insufficient balance warning */}
          {parseFloat(amount) > balance && (
            <div className="send-warning">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
                <line x1="12" y1="9" x2="12" y2="13" />
                <line x1="12" y1="17" x2="12.01" y2="17" />
              </svg>
              <span>Insufficient balance</span>
            </div>
          )}

          {/* Fee info */}
          <div className="send-fee-card">
            <span>Network Fee</span>
            <span>~0.001 {config.symbol}</span>
          </div>

          {/* Fixed bottom send button */}
          <div className="send-bottom-fixed">
            <button
              onClick={handleSend}
              disabled={!recipient || !amount || parseFloat(amount) > balance || !!addressError}
              className="send-btn-primary"
            >
              Send {config.symbol}
            </button>
          </div>

          {/* QR Scanner modal */}
          {showQRScanner && (
            <QRScanner
              onScan={handleQRScan}
              onClose={() => setShowQRScanner(false)}
            />
          )}

          {/* Chain picker modal */}
          {showChainPicker && (
            <div className="send-picker-modal">
              <div className="send-picker-overlay" onClick={() => setShowChainPicker(false)} />
              <div className="send-picker-sheet">
                <div className="send-picker-handle" />
                <h3 className="send-picker-title">Select Asset</h3>
                <div className="send-picker-list">
                  {accounts.map((acc) => {
                    const cfg = CHAIN_CONFIGS[acc.chain];
                    const isSelected = selectedChain === acc.chain;
                    return (
                      <button
                        key={acc.chain}
                        onClick={() => {
                          setSelectedChain(acc.chain);
                          setShowChainPicker(false);
                          setRecipient('');
                          setAddressError(null);
                        }}
                        className={`send-picker-item ${isSelected ? 'selected' : ''}`}
                      >
                        <div
                          className="send-picker-icon"
                          style={{ backgroundColor: `${cfg.color}20` }}
                        >
                          <span style={{ color: cfg.color }}>{cfg.symbol.slice(0, 2)}</span>
                        </div>
                        <div className="send-picker-info">
                          <span className="send-picker-symbol">{cfg.symbol}</span>
                          <span className="send-picker-balance">{acc.balance}</span>
                        </div>
                        {isSelected && (
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" className="send-picker-check">
                            <path d="M5 12l5 5L19 7" />
                          </svg>
                        )}
                      </button>
                    );
                  })}
                </div>
                <button
                  className="send-picker-cancel"
                  onClick={() => setShowChainPicker(false)}
                >
                  Cancel
                </button>
              </div>
            </div>
          )}
        </div>

        <style>{`
          .send-screen {
            display: flex;
            flex-direction: column;
            padding: 16px;
            padding-bottom: 100px;
            gap: 20px;
          }

          .send-section {
            display: flex;
            flex-direction: column;
            gap: 10px;
          }

          .send-label {
            font-size: 13px;
            font-weight: 500;
            color: #8b949e;
            text-transform: uppercase;
            letter-spacing: 0.5px;
          }

          /* Asset picker */
          .send-asset-picker {
            display: flex;
            align-items: center;
            gap: 14px;
            padding: 14px 16px;
            background: #0f1318;
            border: 1px solid #2d3748;
            border-radius: 16px;
            cursor: pointer;
          }

          .send-asset-picker:active {
            background: #1a1f2a;
          }

          .send-asset-icon {
            width: 44px;
            height: 44px;
            border-radius: 12px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 14px;
            font-weight: 700;
          }

          .send-asset-info {
            flex: 1;
            text-align: left;
          }

          .send-asset-symbol {
            display: block;
            font-size: 16px;
            font-weight: 600;
            color: white;
          }

          .send-asset-name {
            display: block;
            font-size: 13px;
            color: #6b7689;
            margin-top: 2px;
          }

          .send-chevron {
            width: 20px;
            height: 20px;
            color: #6b7689;
          }

          /* Balance card */
          .send-balance-card {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 16px;
            background: linear-gradient(135deg, rgba(59, 130, 246, 0.08) 0%, rgba(37, 99, 235, 0.05) 100%);
            border: 1px solid rgba(59, 130, 246, 0.15);
            border-radius: 14px;
          }

          .send-balance-label {
            font-size: 14px;
            color: #8b949e;
          }

          .send-balance-value {
            font-size: 16px;
            font-weight: 600;
            color: white;
          }

          /* Address display */
          .send-address-display {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 16px;
            background: rgba(59, 130, 246, 0.08);
            border: 1px solid rgba(59, 130, 246, 0.3);
            border-radius: 14px;
          }

          .send-address-display.has-error {
            background: rgba(239, 68, 68, 0.08);
            border-color: rgba(239, 68, 68, 0.3);
          }

          .send-address-text {
            font-family: 'SF Mono', 'Menlo', monospace;
            font-size: 14px;
            color: white;
          }

          .send-address-clear {
            width: 32px;
            height: 32px;
            border-radius: 50%;
            background: #2d3748;
            border: none;
            display: flex;
            align-items: center;
            justify-content: center;
            cursor: pointer;
          }

          .send-address-clear svg {
            width: 16px;
            height: 16px;
            color: #8b949e;
          }

          .send-address-empty {
            padding: 18px 16px;
            background: #0f1318;
            border: 1px solid #2d3748;
            border-radius: 14px;
            text-align: center;
          }

          .send-address-empty span {
            font-size: 15px;
            color: #6b7689;
          }

          /* Address action buttons */
          .send-address-actions {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 12px;
          }

          .send-action-btn {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 10px;
            padding: 14px;
            background: #0f1318;
            border: 1px solid #2d3748;
            border-radius: 14px;
            cursor: pointer;
            transition: all 0.2s;
          }

          .send-action-btn:active {
            background: #1a1f2a;
            border-color: #3b82f6;
          }

          .send-action-btn svg {
            width: 20px;
            height: 20px;
            color: #6b7689;
          }

          .send-action-btn span {
            font-size: 15px;
            font-weight: 500;
            color: #8b949e;
          }

          /* Error message */
          .send-error-msg {
            display: flex;
            align-items: center;
            gap: 10px;
            padding: 12px 14px;
            background: rgba(239, 68, 68, 0.1);
            border-radius: 12px;
          }

          .send-error-msg svg {
            width: 18px;
            height: 18px;
            color: #ef4444;
            flex-shrink: 0;
          }

          .send-error-msg span {
            font-size: 14px;
            color: #ef4444;
          }

          /* Amount input */
          .send-amount-wrap {
            position: relative;
          }

          .send-amount-input {
            width: 100%;
            padding: 18px 80px 18px 18px;
            background: #0f1318;
            border: 1px solid #2d3748;
            border-radius: 14px;
            font-size: 20px;
            font-weight: 600;
            color: white;
            outline: none;
          }

          .send-amount-input::placeholder {
            color: #4b5563;
          }

          .send-amount-input:focus {
            border-color: #3b82f6;
          }

          .send-max-btn {
            position: absolute;
            right: 14px;
            top: 50%;
            transform: translateY(-50%);
            padding: 8px 14px;
            background: rgba(59, 130, 246, 0.15);
            border: none;
            border-radius: 8px;
            font-size: 13px;
            font-weight: 600;
            color: #3b82f6;
            cursor: pointer;
          }

          .send-max-btn:active {
            background: rgba(59, 130, 246, 0.25);
          }

          /* Standard input */
          .send-input {
            width: 100%;
            padding: 16px;
            background: #0f1318;
            border: 1px solid #2d3748;
            border-radius: 14px;
            font-size: 16px;
            color: white;
            outline: none;
          }

          .send-input::placeholder {
            color: #4b5563;
          }

          .send-input:focus {
            border-color: #3b82f6;
          }

          /* Warning */
          .send-warning {
            display: flex;
            align-items: center;
            gap: 12px;
            padding: 14px 16px;
            background: rgba(239, 68, 68, 0.1);
            border: 1px solid rgba(239, 68, 68, 0.2);
            border-radius: 14px;
          }

          .send-warning svg {
            width: 20px;
            height: 20px;
            color: #ef4444;
            flex-shrink: 0;
          }

          .send-warning span {
            font-size: 14px;
            color: #ef4444;
          }

          /* Fee card */
          .send-fee-card {
            display: flex;
            justify-content: space-between;
            padding: 14px 16px;
            background: #0f1318;
            border-radius: 12px;
          }

          .send-fee-card span:first-child {
            font-size: 14px;
            color: #6b7689;
          }

          .send-fee-card span:last-child {
            font-size: 14px;
            color: #8b949e;
          }

          /* Fixed bottom button */
          .send-bottom-fixed {
            position: fixed;
            bottom: 0;
            left: 0;
            right: 0;
            padding: 16px 20px;
            padding-bottom: max(20px, env(safe-area-inset-bottom));
            background: linear-gradient(180deg, transparent 0%, #0f1219 30%);
            z-index: 100;
          }

          .send-btn-primary {
            width: 100%;
            padding: 18px;
            background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
            border: none;
            border-radius: 16px;
            font-size: 17px;
            font-weight: 600;
            color: white;
            cursor: pointer;
            transition: all 0.2s;
          }

          .send-btn-primary:disabled {
            opacity: 0.4;
            cursor: not-allowed;
          }

          .send-btn-primary:not(:disabled):active {
            transform: scale(0.98);
          }

          /* Chain picker modal */
          .send-picker-modal {
            position: fixed;
            inset: 0;
            z-index: 1000;
            display: flex;
            align-items: flex-end;
          }

          .send-picker-overlay {
            position: absolute;
            inset: 0;
            background: rgba(0, 0, 0, 0.8);
          }

          .send-picker-sheet {
            position: relative;
            width: 100%;
            max-height: 70vh;
            background: #161B22;
            border-radius: 24px 24px 0 0;
            padding: 12px 20px 24px;
            overflow-y: auto;
            animation: pickerSlideUp 0.3s ease;
          }

          @keyframes pickerSlideUp {
            from { transform: translateY(100%); }
            to { transform: translateY(0); }
          }

          .send-picker-handle {
            width: 40px;
            height: 4px;
            background: #3d4654;
            border-radius: 2px;
            margin: 0 auto 16px;
          }

          .send-picker-title {
            font-size: 18px;
            font-weight: 600;
            color: white;
            margin: 0 0 16px;
          }

          .send-picker-list {
            display: flex;
            flex-direction: column;
            gap: 8px;
          }

          .send-picker-item {
            display: flex;
            align-items: center;
            gap: 14px;
            padding: 14px;
            background: #0f1318;
            border: 1px solid transparent;
            border-radius: 14px;
            cursor: pointer;
            text-align: left;
          }

          .send-picker-item.selected {
            background: rgba(59, 130, 246, 0.1);
            border-color: rgba(59, 130, 246, 0.3);
          }

          .send-picker-icon {
            width: 44px;
            height: 44px;
            border-radius: 12px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 14px;
            font-weight: 700;
          }

          .send-picker-info {
            flex: 1;
          }

          .send-picker-symbol {
            display: block;
            font-size: 16px;
            font-weight: 600;
            color: white;
          }

          .send-picker-balance {
            display: block;
            font-size: 13px;
            color: #6b7689;
            margin-top: 2px;
          }

          .send-picker-check {
            width: 24px;
            height: 24px;
            color: #3b82f6;
          }

          .send-picker-cancel {
            width: 100%;
            margin-top: 16px;
            padding: 16px;
            background: #2d3748;
            border: none;
            border-radius: 14px;
            font-size: 16px;
            font-weight: 600;
            color: white;
            cursor: pointer;
          }
        `}</style>
      </BottomSheet>
    );
  }

  // Transaction confirmation view
  if (step === 'confirm' && pendingTransaction) {
    return (
      <TransactionConfirmation
        transaction={pendingTransaction}
        onConfirm={handleTransactionConfirm}
        onCancel={handleTransactionCancel}
        getMnemonicForSigning={getMnemonicForSigning}
        cachedPin={_cachedPin}
      />
    );
  }

  // Success view
  if (step === 'success' && txResult) {
    return (
      <BottomSheet fullHeight onClose={() => navigate('/portfolio')}>
        <div className="send-result">
          <div className="send-result-icon success">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path d="M5 12l5 5L19 7" />
            </svg>
          </div>

          <h2 className="send-result-title">Transaction Sent!</h2>
          <p className="send-result-desc">
            Your {amount} {config.symbol} has been sent successfully.
          </p>

          <div className="send-result-card">
            <div className="send-result-row">
              <span>To</span>
              <span className="mono">{recipient.slice(0, 12)}...{recipient.slice(-6)}</span>
            </div>
            <div className="send-result-row">
              <span>Amount</span>
              <span>{amount} {config.symbol}</span>
            </div>
            {txResult.txHash && (
              <div className="send-result-row">
                <span>Tx Hash</span>
                <a
                  href={`${config.explorerUrl}/tx/${txResult.txHash}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="send-result-link"
                >
                  {txResult.txHash.slice(0, 8)}...
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6" />
                    <path d="M15 3h6v6" />
                    <path d="M10 14L21 3" />
                  </svg>
                </a>
              </div>
            )}
          </div>

          <button className="send-result-btn" onClick={() => navigate('/portfolio')}>
            Done
          </button>
        </div>

        <style>{`
          .send-result {
            display: flex;
            flex-direction: column;
            align-items: center;
            padding: 48px 24px;
            text-align: center;
          }

          .send-result-icon {
            width: 88px;
            height: 88px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-bottom: 24px;
          }

          .send-result-icon.success {
            background: rgba(16, 185, 129, 0.15);
          }

          .send-result-icon.success svg {
            width: 44px;
            height: 44px;
            color: #10b981;
          }

          .send-result-icon.error {
            background: rgba(239, 68, 68, 0.15);
          }

          .send-result-icon.error svg {
            width: 44px;
            height: 44px;
            color: #ef4444;
          }

          .send-result-title {
            font-size: 24px;
            font-weight: 700;
            color: white;
            margin: 0 0 8px;
          }

          .send-result-desc {
            font-size: 15px;
            color: #8b949e;
            margin: 0 0 32px;
          }

          .send-result-card {
            width: 100%;
            max-width: 320px;
            background: #0f1318;
            border-radius: 16px;
            padding: 4px 16px;
            margin-bottom: 32px;
          }

          .send-result-row {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 14px 0;
            border-bottom: 1px solid #1e242e;
          }

          .send-result-row:last-child {
            border-bottom: none;
          }

          .send-result-row span:first-child {
            font-size: 14px;
            color: #6b7689;
          }

          .send-result-row span:last-child,
          .send-result-row .mono {
            font-size: 14px;
            font-weight: 500;
            color: white;
          }

          .send-result-row .mono {
            font-family: 'SF Mono', 'Menlo', monospace;
            font-size: 13px;
            color: #60a5fa;
          }

          .send-result-link {
            display: flex;
            align-items: center;
            gap: 6px;
            font-size: 14px;
            color: #3b82f6;
            text-decoration: none;
          }

          .send-result-link svg {
            width: 14px;
            height: 14px;
          }

          .send-result-btn {
            width: 100%;
            max-width: 320px;
            padding: 18px;
            background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
            border: none;
            border-radius: 16px;
            font-size: 17px;
            font-weight: 600;
            color: white;
            cursor: pointer;
          }

          .send-result-buttons {
            display: flex;
            gap: 12px;
            width: 100%;
            max-width: 320px;
          }

          .send-result-buttons button {
            flex: 1;
            padding: 16px;
            border: none;
            border-radius: 14px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
          }

          .send-result-buttons .secondary {
            background: #2d3748;
            color: white;
          }

          .send-result-buttons .primary {
            background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
            color: white;
          }
        `}</style>
      </BottomSheet>
    );
  }

  // Error view
  if (step === 'error') {
    return (
      <BottomSheet fullHeight onClose={() => navigate('/portfolio')}>
        <div className="send-result">
          <div className="send-result-icon error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 8v4M12 16h.01" />
            </svg>
          </div>

          <h2 className="send-result-title">Transaction Failed</h2>
          <p className="send-result-desc">
            {txResult?.error || 'An error occurred while sending your transaction.'}
          </p>

          <div className="send-result-buttons">
            <button
              className="secondary"
              onClick={() => { setStep('form'); setTxResult(null); }}
            >
              Try Again
            </button>
            <button
              className="primary"
              onClick={() => navigate('/portfolio')}
            >
              Done
            </button>
          </div>
        </div>

        <style>{`
          .send-result {
            display: flex;
            flex-direction: column;
            align-items: center;
            padding: 48px 24px;
            text-align: center;
          }

          .send-result-icon {
            width: 88px;
            height: 88px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-bottom: 24px;
          }

          .send-result-icon.error {
            background: rgba(239, 68, 68, 0.15);
          }

          .send-result-icon.error svg {
            width: 44px;
            height: 44px;
            color: #ef4444;
          }

          .send-result-title {
            font-size: 24px;
            font-weight: 700;
            color: white;
            margin: 0 0 8px;
          }

          .send-result-desc {
            font-size: 15px;
            color: #8b949e;
            margin: 0 0 32px;
            max-width: 280px;
          }

          .send-result-buttons {
            display: flex;
            gap: 12px;
            width: 100%;
            max-width: 320px;
          }

          .send-result-buttons button {
            flex: 1;
            padding: 16px;
            border: none;
            border-radius: 14px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
          }

          .send-result-buttons .secondary {
            background: #2d3748;
            color: white;
          }

          .send-result-buttons .primary {
            background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
            color: white;
          }
        `}</style>
      </BottomSheet>
    );
  }

  return null;
}
