/**
 * Transaction Confirmation Component
 *
 * A professional transaction confirmation flow with:
 * - Transaction details display
 * - Slide-to-confirm gesture
 * - Face ID / PIN authentication
 */

import { useState, useRef, useEffect, useCallback } from 'react';

export interface TransactionDetails {
  type: 'stake' | 'unstake' | 'send' | 'claim' | 'trade' | 'escrow' | 'loan';
  title: string;
  amount: string;
  token: string;
  recipient?: string;
  fee?: string;
  details?: Array<{ label: string; value: string }>;
  warning?: string;
}

interface TransactionConfirmationProps {
  transaction: TransactionDetails;
  onConfirm: (mnemonic: string) => Promise<void>;
  onCancel: () => void;
  getMnemonicForSigning: (pin: string) => Promise<string>;
  cachedPin: string | null;
}

type AuthStep = 'review' | 'slide' | 'auth' | 'processing' | 'success' | 'error';

export function TransactionConfirmation({
  transaction,
  onConfirm,
  onCancel,
  getMnemonicForSigning,
  cachedPin
}: TransactionConfirmationProps) {
  const tg = window.Telegram?.WebApp;

  const [step, setStep] = useState<AuthStep>('review');
  const [slideProgress, setSlideProgress] = useState(0);
  const [isSliding, setIsSliding] = useState(false);
  const [pin, setPin] = useState(['', '', '', '', '', '']);
  const [pinError, setPinError] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isBiometricAvailable, setIsBiometricAvailable] = useState(false);
  const [isAuthenticating, setIsAuthenticating] = useState(false);

  const sliderRef = useRef<HTMLDivElement>(null);
  const sliderTrackRef = useRef<HTMLDivElement>(null);
  const pinInputRefs = useRef<(HTMLInputElement | null)[]>([]);
  const startXRef = useRef(0);
  const slideWidthRef = useRef(0);

  // Check for biometric availability
  useEffect(() => {
    // Check if biometric is available via Telegram WebApp
    // BiometricManager is available in newer versions of Telegram WebApp API
    const biometricManager = (tg as any)?.BiometricManager;
    if (biometricManager?.isBiometricAvailable) {
      setIsBiometricAvailable(true);
    }
  }, [tg]);

  // Auto-trigger biometric when reaching auth step
  useEffect(() => {
    if (step === 'auth' && isBiometricAvailable && !isAuthenticating) {
      triggerBiometric();
    }
  }, [step, isBiometricAvailable]);

  // Focus first PIN input when auth step is reached
  useEffect(() => {
    if (step === 'auth' && !isBiometricAvailable) {
      setTimeout(() => pinInputRefs.current[0]?.focus(), 100);
    }
  }, [step, isBiometricAvailable]);

  const triggerBiometric = async () => {
    const biometricManager = (tg as any)?.BiometricManager;
    if (!biometricManager) return;

    setIsAuthenticating(true);

    try {
      // Request biometric authentication
      biometricManager.authenticate({
        reason: `Confirm ${transaction.type} transaction`
      }, async (success: boolean, token?: string) => {
        if (success && token) {
          tg?.HapticFeedback?.notificationOccurred('success');
          await handleBiometricSuccess(token);
        } else {
          // Fall back to PIN entry
          setIsAuthenticating(false);
          tg?.HapticFeedback?.notificationOccurred('error');
        }
      });
    } catch {
      setIsAuthenticating(false);
    }
  };

  const handleBiometricSuccess = async (_token: string) => {
    setStep('processing');
    try {
      // For biometric, we use the cached PIN if available
      // The token validates the biometric was successful
      if (cachedPin) {
        const mnemonic = await getMnemonicForSigning(cachedPin);
        await onConfirm(mnemonic);
        setStep('success');
        tg?.HapticFeedback?.notificationOccurred('success');
        setTimeout(onCancel, 1500);
      } else {
        // No cached PIN, need to enter manually
        setStep('auth');
        setIsAuthenticating(false);
      }
    } catch (error) {
      setStep('error');
      setErrorMessage(error instanceof Error ? error.message : 'Transaction failed');
      tg?.HapticFeedback?.notificationOccurred('error');
    }
  };

  const handleSlideStart = (e: React.TouchEvent | React.MouseEvent) => {
    if (step !== 'review') return;

    setIsSliding(true);
    const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX;
    startXRef.current = clientX;

    if (sliderTrackRef.current) {
      slideWidthRef.current = sliderTrackRef.current.offsetWidth - 60; // 60 = slider button width
    }

    tg?.HapticFeedback?.impactOccurred('light');
  };

  const handleSlideMove = useCallback((e: TouchEvent | MouseEvent) => {
    if (!isSliding) return;

    const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX;
    const deltaX = clientX - startXRef.current;
    const progress = Math.max(0, Math.min(100, (deltaX / slideWidthRef.current) * 100));
    setSlideProgress(progress);

    // Haptic feedback at milestones
    if (progress > 25 && progress < 30) {
      tg?.HapticFeedback?.impactOccurred('light');
    } else if (progress > 50 && progress < 55) {
      tg?.HapticFeedback?.impactOccurred('medium');
    } else if (progress > 75 && progress < 80) {
      tg?.HapticFeedback?.impactOccurred('heavy');
    }
  }, [isSliding, tg]);

  const handleSlideEnd = useCallback(() => {
    if (!isSliding) return;

    setIsSliding(false);

    if (slideProgress >= 95) {
      // Completed slide
      setSlideProgress(100);
      tg?.HapticFeedback?.notificationOccurred('success');
      setTimeout(() => {
        setStep('auth');
      }, 200);
    } else {
      // Reset
      setSlideProgress(0);
    }
  }, [isSliding, slideProgress, tg]);

  // Add/remove global listeners for slide
  useEffect(() => {
    if (isSliding) {
      window.addEventListener('mousemove', handleSlideMove);
      window.addEventListener('mouseup', handleSlideEnd);
      window.addEventListener('touchmove', handleSlideMove);
      window.addEventListener('touchend', handleSlideEnd);
    }

    return () => {
      window.removeEventListener('mousemove', handleSlideMove);
      window.removeEventListener('mouseup', handleSlideEnd);
      window.removeEventListener('touchmove', handleSlideMove);
      window.removeEventListener('touchend', handleSlideEnd);
    };
  }, [isSliding, handleSlideMove, handleSlideEnd]);

  const handlePinChange = (index: number, value: string) => {
    if (!/^\d*$/.test(value)) return;

    const newPin = [...pin];
    newPin[index] = value.slice(-1);
    setPin(newPin);
    setPinError(null);

    // Auto-focus next input
    if (value && index < 5) {
      pinInputRefs.current[index + 1]?.focus();
    }

    // Auto-submit when all digits entered
    if (value && index === 5 && newPin.every(d => d !== '')) {
      handlePinSubmit(newPin.join(''));
    }

    tg?.HapticFeedback?.impactOccurred('light');
  };

  const handlePinKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !pin[index] && index > 0) {
      pinInputRefs.current[index - 1]?.focus();
    }
  };

  const handlePinSubmit = async (enteredPin: string) => {
    setStep('processing');

    try {
      const mnemonic = await getMnemonicForSigning(enteredPin);
      await onConfirm(mnemonic);
      setStep('success');
      tg?.HapticFeedback?.notificationOccurred('success');
      setTimeout(onCancel, 1500);
    } catch (error) {
      setStep('auth');
      setPin(['', '', '', '', '', '']);
      setPinError('Invalid PIN. Please try again.');
      tg?.HapticFeedback?.notificationOccurred('error');
      setTimeout(() => pinInputRefs.current[0]?.focus(), 100);
    }
  };

  const getTypeIcon = () => {
    switch (transaction.type) {
      case 'stake': return '🔒';
      case 'unstake': return '🔓';
      case 'send': return '↗️';
      case 'claim': return '🎁';
      case 'trade': return '💱';
      case 'escrow': return '🤝';
      case 'loan': return '💰';
      default: return '📝';
    }
  };

  const getTypeColor = () => {
    switch (transaction.type) {
      case 'stake': return '#10b981';
      case 'unstake': return '#f59e0b';
      case 'send': return '#3b82f6';
      case 'claim': return '#8b5cf6';
      case 'trade': return '#06b6d4';
      case 'escrow': return '#ec4899';
      case 'loan': return '#f97316';
      default: return '#6b7280';
    }
  };

  return (
    <div className="tx-confirm-overlay">
      <div className="tx-confirm-container">
        {/* Header */}
        <div className="tx-confirm-header">
          <button className="close-btn" onClick={onCancel}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
          <h2>{step === 'success' ? 'Success!' : step === 'error' ? 'Failed' : 'Confirm Transaction'}</h2>
          <div className="spacer" />
        </div>

        {/* Content based on step */}
        {(step === 'review' || step === 'slide') && (
          <>
            {/* Transaction Summary */}
            <div className="tx-summary">
              <div className="tx-icon" style={{ backgroundColor: `${getTypeColor()}20` }}>
                <span>{getTypeIcon()}</span>
              </div>
              <h3 className="tx-title">{transaction.title}</h3>
              <div className="tx-amount">
                <span className="amount">{transaction.amount}</span>
                <span className="token">{transaction.token}</span>
              </div>
            </div>

            {/* Transaction Details */}
            <div className="tx-details">
              {transaction.recipient && (
                <div className="detail-row">
                  <span className="label">To</span>
                  <span className="value address">{formatAddress(transaction.recipient)}</span>
                </div>
              )}
              {transaction.details?.map((detail, i) => (
                <div key={i} className="detail-row">
                  <span className="label">{detail.label}</span>
                  <span className="value">{detail.value}</span>
                </div>
              ))}
              <div className="detail-row">
                <span className="label">Network Fee</span>
                <span className="value">{transaction.fee || '~0.001 HODL'}</span>
              </div>
            </div>

            {/* Warning */}
            {transaction.warning && (
              <div className="tx-warning">
                <svg viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
                  <path d="M12 9v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span>{transaction.warning}</span>
              </div>
            )}

            {/* Slide to Confirm */}
            <div className="slide-container">
              <div
                ref={sliderTrackRef}
                className="slide-track"
                style={{
                  background: slideProgress > 0
                    ? `linear-gradient(90deg, ${getTypeColor()}40 ${slideProgress}%, #30363d ${slideProgress}%)`
                    : '#30363d'
                }}
              >
                <div
                  ref={sliderRef}
                  className="slide-button"
                  style={{
                    left: `${slideProgress}%`,
                    backgroundColor: slideProgress > 90 ? getTypeColor() : '#1E40AF'
                  }}
                  onTouchStart={handleSlideStart}
                  onMouseDown={handleSlideStart}
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                    <path d="M9 5l7 7-7 7" />
                  </svg>
                </div>
                <span className="slide-text" style={{ opacity: 1 - slideProgress / 100 }}>
                  Slide to confirm
                </span>
              </div>
            </div>
          </>
        )}

        {step === 'auth' && (
          <div className="auth-container">
            {isAuthenticating ? (
              <div className="biometric-scanning">
                <div className="face-id-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <path d="M9 4H5a1 1 0 00-1 1v4" />
                    <path d="M15 4h4a1 1 0 011 1v4" />
                    <path d="M9 20H5a1 1 0 01-1-1v-4" />
                    <path d="M15 20h4a1 1 0 001-1v-4" />
                    <circle cx="9" cy="9" r="1" fill="currentColor" />
                    <circle cx="15" cy="9" r="1" fill="currentColor" />
                    <path d="M9 15c.83.67 2 1 3 1s2.17-.33 3-1" />
                  </svg>
                </div>
                <p className="scanning-text">Scanning Face ID...</p>
                <div className="scanning-animation">
                  <div className="scan-line" />
                </div>
              </div>
            ) : (
              <>
                <div className="pin-header">
                  <div className="lock-icon">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                      <path d="M7 11V7a5 5 0 0110 0v4" />
                    </svg>
                  </div>
                  <h3>Enter PIN to Confirm</h3>
                  <p>Enter your 6-digit PIN to authorize this transaction</p>
                </div>

                <div className="pin-input-container">
                  {pin.map((digit, i) => (
                    <input
                      key={i}
                      ref={el => pinInputRefs.current[i] = el}
                      type="password"
                      inputMode="numeric"
                      maxLength={1}
                      value={digit}
                      onChange={e => handlePinChange(i, e.target.value)}
                      onKeyDown={e => handlePinKeyDown(i, e)}
                      className={`pin-digit ${digit ? 'filled' : ''} ${pinError ? 'error' : ''}`}
                      autoComplete="off"
                    />
                  ))}
                </div>

                {pinError && (
                  <p className="pin-error">{pinError}</p>
                )}

                {isBiometricAvailable && (
                  <button className="biometric-btn" onClick={triggerBiometric}>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                      <path d="M9 4H5a1 1 0 00-1 1v4" />
                      <path d="M15 4h4a1 1 0 011 1v4" />
                      <path d="M9 20H5a1 1 0 01-1-1v-4" />
                      <path d="M15 20h4a1 1 0 001-1v-4" />
                      <circle cx="9" cy="9" r="1" fill="currentColor" />
                      <circle cx="15" cy="9" r="1" fill="currentColor" />
                      <path d="M9 15c.83.67 2 1 3 1s2.17-.33 3-1" />
                    </svg>
                    Use Face ID
                  </button>
                )}
              </>
            )}
          </div>
        )}

        {step === 'processing' && (
          <div className="processing-container">
            <div className="processing-spinner" />
            <p>Processing transaction...</p>
          </div>
        )}

        {step === 'success' && (
          <div className="success-container">
            <div className="success-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <path d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h3>Transaction Confirmed</h3>
            <p>Your {transaction.type} transaction was successful</p>
          </div>
        )}

        {step === 'error' && (
          <div className="error-container">
            <div className="error-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <path d="M6 18L18 6M6 6l12 12" />
              </svg>
            </div>
            <h3>Transaction Failed</h3>
            <p>{errorMessage || 'An error occurred'}</p>
            <button className="retry-btn" onClick={() => setStep('review')}>
              Try Again
            </button>
          </div>
        )}
      </div>

      <style>{styles}</style>
    </div>
  );
}

function formatAddress(address: string): string {
  if (address.length <= 16) return address;
  return `${address.slice(0, 10)}...${address.slice(-6)}`;
}

const styles = `
  .tx-confirm-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.9);
    z-index: 1000;
    display: flex;
    align-items: flex-end;
    animation: fadeIn 0.2s ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .tx-confirm-container {
    width: 100%;
    max-height: 90vh;
    background: #161B22;
    border-radius: 24px 24px 0 0;
    padding: 20px;
    padding-bottom: max(20px, env(safe-area-inset-bottom));
    animation: slideUp 0.3s ease;
    overflow-y: auto;
  }

  @keyframes slideUp {
    from { transform: translateY(100%); }
    to { transform: translateY(0); }
  }

  .tx-confirm-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 24px;
  }

  .tx-confirm-header h2 {
    font-size: 18px;
    font-weight: 700;
    color: white;
    margin: 0;
  }

  .close-btn {
    width: 36px;
    height: 36px;
    border: none;
    background: #30363d;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .close-btn svg {
    width: 20px;
    height: 20px;
    color: #8b949e;
  }

  .spacer {
    width: 36px;
  }

  /* Transaction Summary */
  .tx-summary {
    text-align: center;
    padding: 24px 0;
    border-bottom: 1px solid #30363d;
    margin-bottom: 20px;
  }

  .tx-icon {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 16px;
    font-size: 28px;
  }

  .tx-title {
    font-size: 16px;
    font-weight: 600;
    color: #8b949e;
    margin: 0 0 12px;
  }

  .tx-amount {
    display: flex;
    align-items: baseline;
    justify-content: center;
    gap: 8px;
  }

  .tx-amount .amount {
    font-size: 36px;
    font-weight: 700;
    color: white;
  }

  .tx-amount .token {
    font-size: 20px;
    font-weight: 600;
    color: #8b949e;
  }

  /* Transaction Details */
  .tx-details {
    background: #0D1117;
    border-radius: 16px;
    padding: 16px;
    margin-bottom: 20px;
  }

  .detail-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;
    border-bottom: 1px solid #21262d;
  }

  .detail-row:last-child {
    border-bottom: none;
  }

  .detail-row .label {
    font-size: 14px;
    color: #8b949e;
  }

  .detail-row .value {
    font-size: 14px;
    font-weight: 600;
    color: white;
  }

  .detail-row .value.address {
    font-family: monospace;
    font-size: 13px;
    color: #58a6ff;
  }

  /* Warning */
  .tx-warning {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 14px 16px;
    background: rgba(245, 158, 11, 0.1);
    border: 1px solid rgba(245, 158, 11, 0.3);
    border-radius: 12px;
    margin-bottom: 24px;
  }

  .tx-warning svg {
    flex-shrink: 0;
    color: #f59e0b;
  }

  .tx-warning span {
    font-size: 13px;
    color: #f59e0b;
  }

  /* Slide to Confirm */
  .slide-container {
    padding: 8px 0;
  }

  .slide-track {
    position: relative;
    height: 60px;
    border-radius: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    user-select: none;
    -webkit-user-select: none;
  }

  .slide-button {
    position: absolute;
    left: 0;
    top: 50%;
    transform: translateY(-50%);
    width: 56px;
    height: 56px;
    margin-left: 2px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: grab;
    transition: left 0.1s ease-out, background-color 0.2s ease;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .slide-button:active {
    cursor: grabbing;
  }

  .slide-button svg {
    width: 24px;
    height: 24px;
    color: white;
  }

  .slide-text {
    font-size: 16px;
    font-weight: 600;
    color: #8b949e;
    pointer-events: none;
    transition: opacity 0.2s ease;
  }

  /* Auth Container */
  .auth-container {
    padding: 32px 0;
    text-align: center;
  }

  .pin-header {
    margin-bottom: 32px;
  }

  .lock-icon {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    background: rgba(30, 64, 175, 0.1);
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 20px;
  }

  .lock-icon svg {
    width: 32px;
    height: 32px;
    color: #3b82f6;
  }

  .pin-header h3 {
    font-size: 20px;
    font-weight: 700;
    color: white;
    margin: 0 0 8px;
  }

  .pin-header p {
    font-size: 14px;
    color: #8b949e;
    margin: 0;
  }

  .pin-input-container {
    display: flex;
    justify-content: center;
    gap: 12px;
    margin-bottom: 24px;
  }

  .pin-digit {
    width: 48px;
    height: 56px;
    border: 2px solid #30363d;
    border-radius: 12px;
    background: #0D1117;
    font-size: 24px;
    font-weight: 700;
    color: white;
    text-align: center;
    outline: none;
    transition: all 0.2s ease;
  }

  .pin-digit:focus {
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
  }

  .pin-digit.filled {
    border-color: #3b82f6;
  }

  .pin-digit.error {
    border-color: #ef4444;
    animation: shake 0.5s ease;
  }

  @keyframes shake {
    0%, 100% { transform: translateX(0); }
    20%, 60% { transform: translateX(-8px); }
    40%, 80% { transform: translateX(8px); }
  }

  .pin-error {
    font-size: 14px;
    color: #ef4444;
    margin: 0 0 16px;
  }

  .biometric-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    width: 100%;
    padding: 16px;
    border: 1px solid #30363d;
    border-radius: 14px;
    background: transparent;
    color: white;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .biometric-btn:active {
    background: #21262d;
  }

  .biometric-btn svg {
    width: 28px;
    height: 28px;
    color: #3b82f6;
  }

  /* Biometric Scanning */
  .biometric-scanning {
    padding: 40px 0;
    text-align: center;
  }

  .face-id-icon {
    width: 100px;
    height: 100px;
    margin: 0 auto 24px;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .face-id-icon svg {
    width: 80px;
    height: 80px;
    color: #3b82f6;
  }

  .scanning-text {
    font-size: 18px;
    font-weight: 600;
    color: white;
    margin: 0 0 24px;
  }

  .scanning-animation {
    width: 120px;
    height: 120px;
    margin: 0 auto;
    position: relative;
    border-radius: 20px;
    overflow: hidden;
    background: rgba(59, 130, 246, 0.1);
  }

  .scan-line {
    position: absolute;
    width: 100%;
    height: 3px;
    background: linear-gradient(90deg, transparent, #3b82f6, transparent);
    animation: scanMove 2s ease-in-out infinite;
  }

  @keyframes scanMove {
    0%, 100% { top: 0; }
    50% { top: calc(100% - 3px); }
  }

  /* Processing */
  .processing-container {
    padding: 60px 0;
    text-align: center;
  }

  .processing-spinner {
    width: 56px;
    height: 56px;
    border: 4px solid #30363d;
    border-top-color: #3b82f6;
    border-radius: 50%;
    margin: 0 auto 24px;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .processing-container p {
    font-size: 16px;
    color: #8b949e;
    margin: 0;
  }

  /* Success */
  .success-container {
    padding: 60px 0;
    text-align: center;
  }

  .success-icon {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    background: rgba(16, 185, 129, 0.1);
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 24px;
    animation: successPop 0.4s ease;
  }

  @keyframes successPop {
    0% { transform: scale(0); }
    50% { transform: scale(1.1); }
    100% { transform: scale(1); }
  }

  .success-icon svg {
    width: 40px;
    height: 40px;
    color: #10b981;
  }

  .success-container h3 {
    font-size: 22px;
    font-weight: 700;
    color: white;
    margin: 0 0 8px;
  }

  .success-container p {
    font-size: 14px;
    color: #8b949e;
    margin: 0;
  }

  /* Error */
  .error-container {
    padding: 60px 0;
    text-align: center;
  }

  .error-icon {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    background: rgba(239, 68, 68, 0.1);
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 24px;
  }

  .error-icon svg {
    width: 40px;
    height: 40px;
    color: #ef4444;
  }

  .error-container h3 {
    font-size: 22px;
    font-weight: 700;
    color: white;
    margin: 0 0 8px;
  }

  .error-container p {
    font-size: 14px;
    color: #8b949e;
    margin: 0 0 24px;
  }

  .retry-btn {
    padding: 14px 32px;
    border: none;
    border-radius: 12px;
    background: #30363d;
    color: white;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
  }
`;
