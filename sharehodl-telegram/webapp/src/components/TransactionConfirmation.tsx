/**
 * Transaction Confirmation Component - Premium Design
 *
 * Professional transaction confirmation flow with:
 * - Beautiful animated UI with SVG icons
 * - Slide-to-confirm gesture
 * - Face ID / PIN authentication
 * - Smart error handling for common issues
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

type AuthStep = 'review' | 'auth' | 'processing' | 'success' | 'error';

// SVG Icons for transaction types
const TxIcons = {
  stake: (color: string) => (
    <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
      <path d="M7 11V7a5 5 0 0 1 10 0v4" />
    </svg>
  ),
  unstake: (color: string) => (
    <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
      <path d="M7 11V7a5 5 0 0 1 9.9-1" />
    </svg>
  ),
  send: (color: string) => (
    <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 19V5" />
      <path d="M5 12l7-7 7 7" />
    </svg>
  ),
  claim: (color: string) => (
    <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 2v6" />
      <path d="M12 22v-6" />
      <path d="M4.93 4.93l4.24 4.24" />
      <path d="M14.83 14.83l4.24 4.24" />
      <path d="M2 12h6" />
      <path d="M16 12h6" />
      <path d="M4.93 19.07l4.24-4.24" />
      <path d="M14.83 9.17l4.24-4.24" />
    </svg>
  ),
  trade: (color: string) => (
    <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M16 3l4 4-4 4" />
      <path d="M20 7H4" />
      <path d="M8 21l-4-4 4-4" />
      <path d="M4 17h16" />
    </svg>
  ),
  escrow: (color: string) => (
    <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      <path d="M9 12l2 2 4-4" />
    </svg>
  ),
  loan: (color: string) => (
    <svg viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="12" y1="1" x2="12" y2="23" />
      <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
    </svg>
  ),
};

// Format number with commas
function formatWithCommas(value: string): string {
  const num = parseFloat(value);
  if (isNaN(num)) return value;
  return num.toLocaleString('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 6
  });
}

// Parse user-friendly error messages
function parseErrorMessage(error: string): { title: string; message: string; action?: string } {
  const errorLower = error.toLowerCase();

  if (error.includes('does not exist on chain')) {
    return {
      title: 'Wallet Not Activated',
      message: 'Your wallet needs to receive tokens first before you can send transactions.',
      action: 'Receive some HODL tokens to activate your wallet'
    };
  }

  if (errorLower.includes('insufficient funds')) {
    return {
      title: 'Insufficient Balance',
      message: 'You don\'t have enough tokens to complete this transaction.',
      action: 'Add more tokens to your wallet'
    };
  }

  if (errorLower.includes('sequence')) {
    return {
      title: 'Wallet Not Ready',
      message: 'Your wallet needs to be funded before sending transactions.',
      action: 'Receive tokens to activate your wallet'
    };
  }

  // Wallet data corruption errors
  if (errorLower.includes('corrupted') ||
      errorLower.includes('restore') ||
      errorLower.includes('incomplete') ||
      errorLower.includes('invalid characters') ||
      errorLower.includes('wallet data') ||
      errorLower.includes('string length') ||
      errorLower.includes('multiple of 4') ||
      errorLower.includes('base64') ||
      errorLower.includes('decode')) {
    return {
      title: 'Wallet Data Error',
      message: 'Your wallet data appears to be corrupted.',
      action: 'Please restore your wallet using your recovery phrase'
    };
  }

  return {
    title: 'Transaction Failed',
    message: error || 'Something went wrong. Please try again.'
  };
}

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
  const [errorInfo, setErrorInfo] = useState<{ title: string; message: string; action?: string } | null>(null);
  const [isBiometricAvailable, setIsBiometricAvailable] = useState(false);
  const [isAuthenticating, setIsAuthenticating] = useState(false);
  const [slideCompleted, setSlideCompleted] = useState(false);

  const sliderTrackRef = useRef<HTMLDivElement>(null);
  const pinInputRefs = useRef<(HTMLInputElement | null)[]>([]);
  const startXRef = useRef(0);
  const slideWidthRef = useRef(0);

  // Check for biometric availability
  useEffect(() => {
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
      biometricManager.authenticate({
        reason: `Confirm ${transaction.type} transaction`
      }, async (success: boolean, token?: string) => {
        if (success && token) {
          tg?.HapticFeedback?.notificationOccurred('success');
          await handleBiometricSuccess(token);
        } else {
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
    // Clear any stale error state from previous attempts
    setErrorInfo(null);
    setPinError(null);

    let transactionSucceeded = false;

    try {
      if (cachedPin) {
        const mnemonic = await getMnemonicForSigning(cachedPin);
        await onConfirm(mnemonic);
        // If we reach here, the transaction succeeded
        transactionSucceeded = true;
      } else {
        setStep('auth');
        setIsAuthenticating(false);
        return;
      }
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Transaction failed';
      setErrorInfo(parseErrorMessage(errorMsg));
      setStep('error');
      tg?.HapticFeedback?.notificationOccurred('error');
      return;
    }

    // Transaction succeeded - show success UI
    if (transactionSucceeded) {
      try {
        setStep('success');
        tg?.HapticFeedback?.notificationOccurred('success');
        setTimeout(onCancel, 2000);
      } catch {
        // Even if haptics fail, ensure success is shown
        setStep('success');
        setTimeout(onCancel, 2000);
      }
    }
  };

  const handleSlideStart = (e: React.TouchEvent | React.MouseEvent) => {
    if (step !== 'review' || slideCompleted) return;

    setIsSliding(true);
    const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX;
    startXRef.current = clientX;

    if (sliderTrackRef.current) {
      slideWidthRef.current = sliderTrackRef.current.offsetWidth - 64;
    }

    tg?.HapticFeedback?.impactOccurred('light');
  };

  const handleSlideMove = useCallback((e: TouchEvent | MouseEvent) => {
    if (!isSliding || slideCompleted) return;

    const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX;
    const deltaX = clientX - startXRef.current;
    const progress = Math.max(0, Math.min(100, (deltaX / slideWidthRef.current) * 100));
    setSlideProgress(progress);

    // Haptic feedback only at 95% (almost complete)
    if (progress >= 95 && progress < 97) {
      tg?.HapticFeedback?.impactOccurred('heavy');
    }
  }, [isSliding, slideCompleted, tg]);

  const handleSlideEnd = useCallback(() => {
    if (!isSliding || slideCompleted) return;

    setIsSliding(false);

    if (slideProgress >= 92) {
      setSlideCompleted(true);
      setSlideProgress(100);
      tg?.HapticFeedback?.notificationOccurred('success');
      setTimeout(() => setStep('auth'), 300);
    } else {
      setSlideProgress(0);
    }
  }, [isSliding, slideProgress, slideCompleted, tg]);

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

    if (value && index < 5) {
      pinInputRefs.current[index + 1]?.focus();
    }

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
    // Clear any stale error state from previous attempts
    setErrorInfo(null);
    setPinError(null);

    let transactionSucceeded = false;

    try {
      const mnemonic = await getMnemonicForSigning(enteredPin);
      await onConfirm(mnemonic);
      // If we reach here, the transaction succeeded
      transactionSucceeded = true;
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Transaction failed';

      // Check if this is a wallet data error (not a PIN error)
      const isWalletDataError =
        errorMsg.includes('string length') ||
        errorMsg.includes('corrupted') ||
        errorMsg.includes('multiple') ||
        errorMsg.includes('base64');

      // Check if PIN error (but exclude wallet data issues)
      const isPinError = !isWalletDataError && (
        errorMsg.includes('PIN') ||
        errorMsg.includes('Decryption') ||
        errorMsg.toLowerCase().includes('incorrect')
      );

      if (isPinError) {
        setStep('auth');
        setPin(['', '', '', '', '', '']);
        setPinError('Incorrect PIN. Please try again.');
        tg?.HapticFeedback?.notificationOccurred('error');
        setTimeout(() => pinInputRefs.current[0]?.focus(), 100);
      } else {
        setErrorInfo(parseErrorMessage(errorMsg));
        setStep('error');
        tg?.HapticFeedback?.notificationOccurred('error');
      }
      return; // Exit early on error
    }

    // Transaction succeeded - show success UI
    // This is separated to ensure success is always shown
    if (transactionSucceeded) {
      try {
        setStep('success');
        tg?.HapticFeedback?.notificationOccurred('success');
        setTimeout(onCancel, 2000);
      } catch {
        // Even if haptics fail, ensure success is shown
        setStep('success');
        setTimeout(onCancel, 2000);
      }
    }
  };

  const getTypeConfig = () => {
    const configs = {
      stake: { color: '#10b981', gradient: 'linear-gradient(135deg, #10b981 0%, #059669 100%)' },
      unstake: { color: '#f59e0b', gradient: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)' },
      send: { color: '#3b82f6', gradient: 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)' },
      claim: { color: '#8b5cf6', gradient: 'linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)' },
      trade: { color: '#06b6d4', gradient: 'linear-gradient(135deg, #06b6d4 0%, #0891b2 100%)' },
      escrow: { color: '#ec4899', gradient: 'linear-gradient(135deg, #ec4899 0%, #db2777 100%)' },
      loan: { color: '#f97316', gradient: 'linear-gradient(135deg, #f97316 0%, #ea580c 100%)' },
    };
    return configs[transaction.type] || configs.send;
  };

  const config = getTypeConfig();
  const formattedAmount = formatWithCommas(transaction.amount);
  const IconComponent = TxIcons[transaction.type] || TxIcons.send;

  return (
    <div className="tx-overlay">
      <div className="tx-backdrop" onClick={onCancel} />
      <div className="tx-sheet">
        {/* Drag indicator */}
        <div className="tx-drag-indicator" />

        {/* REVIEW STEP */}
        {step === 'review' && (
          <>
            {/* Hero Section */}
            <div className="tx-hero">
              <div className="tx-icon-wrap" style={{ background: `${config.color}15` }}>
                <div className="tx-icon">
                  {IconComponent(config.color)}
                </div>
              </div>
              <h2 className="tx-type-title">{transaction.title}</h2>
              <div className="tx-amount-display">
                <span className="tx-amount-value">{formattedAmount}</span>
                <span className="tx-amount-token">{transaction.token}</span>
              </div>
            </div>

            {/* Details Card */}
            <div className="tx-details-card">
              {transaction.recipient && (
                <div className="tx-detail-row">
                  <span className="tx-detail-label">To</span>
                  <span className="tx-detail-value tx-address">
                    {transaction.recipient.slice(0, 12)}...{transaction.recipient.slice(-6)}
                  </span>
                </div>
              )}
              {transaction.details?.map((detail, i) => (
                <div key={i} className="tx-detail-row">
                  <span className="tx-detail-label">{detail.label}</span>
                  <span className="tx-detail-value">{detail.value}</span>
                </div>
              ))}
              <div className="tx-detail-row">
                <span className="tx-detail-label">Network Fee</span>
                <span className="tx-detail-value tx-fee">{transaction.fee || '~0.001 HODL'}</span>
              </div>
            </div>

            {/* Warning */}
            {transaction.warning && (
              <div className="tx-warning-box">
                <div className="tx-warning-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                    <line x1="12" y1="9" x2="12" y2="13" />
                    <line x1="12" y1="17" x2="12.01" y2="17" />
                  </svg>
                </div>
                <span>{transaction.warning}</span>
              </div>
            )}

            {/* Slide to Confirm */}
            <div className="tx-slide-wrapper">
              <div
                ref={sliderTrackRef}
                className={`tx-slide-track ${slideCompleted ? 'completed' : ''}`}
                style={{
                  background: slideProgress > 0
                    ? `linear-gradient(90deg, ${config.color}30 ${slideProgress}%, #1f2937 ${slideProgress}%)`
                    : undefined
                }}
              >
                <div
                  className="tx-slide-thumb"
                  style={{
                    transform: `translateX(${(slideProgress / 100) * (slideWidthRef.current || 280)}px)`,
                    background: slideCompleted ? config.gradient : config.gradient
                  }}
                  onTouchStart={handleSlideStart}
                  onMouseDown={handleSlideStart}
                >
                  {slideCompleted ? (
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3">
                      <path d="M5 12l5 5L19 7" />
                    </svg>
                  ) : (
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                      <path d="M9 5l7 7-7 7" />
                    </svg>
                  )}
                </div>
                <span className="tx-slide-label" style={{ opacity: 1 - slideProgress / 80 }}>
                  Slide to confirm
                </span>
              </div>
            </div>

            {/* Cancel Button */}
            <button className="tx-cancel-btn" onClick={onCancel}>
              Cancel
            </button>
          </>
        )}

        {/* AUTH STEP */}
        {step === 'auth' && (
          <div className="tx-auth-section">
            {isAuthenticating ? (
              <div className="tx-biometric">
                <div className="tx-face-icon">
                  <svg viewBox="0 0 96 96" fill="none">
                    <rect x="4" y="4" width="24" height="4" rx="2" fill="#3b82f6"/>
                    <rect x="4" y="4" width="4" height="24" rx="2" fill="#3b82f6"/>
                    <rect x="68" y="4" width="24" height="4" rx="2" fill="#3b82f6"/>
                    <rect x="88" y="4" width="4" height="24" rx="2" fill="#3b82f6"/>
                    <rect x="4" y="88" width="24" height="4" rx="2" fill="#3b82f6"/>
                    <rect x="4" y="68" width="4" height="24" rx="2" fill="#3b82f6"/>
                    <rect x="68" y="88" width="24" height="4" rx="2" fill="#3b82f6"/>
                    <rect x="88" y="68" width="4" height="24" rx="2" fill="#3b82f6"/>
                    <circle cx="36" cy="40" r="4" fill="#3b82f6"/>
                    <circle cx="60" cy="40" r="4" fill="#3b82f6"/>
                    <path d="M32 60c6 8 20 8 32 0" stroke="#3b82f6" strokeWidth="3" strokeLinecap="round"/>
                  </svg>
                  <div className="tx-scan-line" />
                </div>
                <h3>Face ID</h3>
                <p>Look at your device to authenticate</p>
              </div>
            ) : (
              <>
                <div className="tx-pin-icon-wrap">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <rect x="3" y="11" width="18" height="11" rx="2"/>
                    <path d="M7 11V7a5 5 0 0110 0v4"/>
                    <circle cx="12" cy="16" r="1" fill="currentColor"/>
                  </svg>
                </div>
                <h3>Enter PIN</h3>
                <p>Authorize this transaction</p>

                <div className="tx-pin-dots">
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
                      className={`tx-pin-input ${digit ? 'filled' : ''} ${pinError ? 'error' : ''}`}
                      autoComplete="off"
                    />
                  ))}
                </div>

                {pinError && <p className="tx-pin-error">{pinError}</p>}

                {isBiometricAvailable && (
                  <button className="tx-face-btn" onClick={triggerBiometric}>
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
                      <rect x="4" y="4" width="3" height="1" rx="0.5"/>
                      <rect x="4" y="4" width="1" height="3" rx="0.5"/>
                      <rect x="17" y="4" width="3" height="1" rx="0.5"/>
                      <rect x="20" y="4" width="1" height="3" rx="0.5"/>
                      <rect x="4" y="19" width="3" height="1" rx="0.5"/>
                      <rect x="4" y="17" width="1" height="3" rx="0.5"/>
                      <rect x="17" y="19" width="3" height="1" rx="0.5"/>
                      <rect x="20" y="17" width="1" height="3" rx="0.5"/>
                      <circle cx="9" cy="10" r="1" fill="currentColor"/>
                      <circle cx="15" cy="10" r="1" fill="currentColor"/>
                      <path d="M9 15c1.5 2 4.5 2 6 0"/>
                    </svg>
                    Use Face ID
                  </button>
                )}

                <button className="tx-back-btn" onClick={() => { setStep('review'); setSlideProgress(0); setSlideCompleted(false); setErrorInfo(null); setPinError(null); }}>
                  Back
                </button>
              </>
            )}
          </div>
        )}

        {/* PROCESSING STEP */}
        {step === 'processing' && (
          <div className="tx-processing">
            <div className="tx-spinner" style={{ borderTopColor: config.color }} />
            <h3>Processing</h3>
            <p>Please wait while we confirm your transaction...</p>
          </div>
        )}

        {/* SUCCESS STEP */}
        {step === 'success' && (
          <div className="tx-success">
            <div className="tx-success-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                <path d="M5 12l5 5L19 7" />
              </svg>
            </div>
            <h3>Success!</h3>
            <p>Your {transaction.type} of {formattedAmount} {transaction.token} is complete</p>
          </div>
        )}

        {/* ERROR STEP */}
        {step === 'error' && errorInfo && (
          <div className="tx-error">
            <div className="tx-error-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="12" cy="12" r="10" />
                <path d="M12 8v4M12 16h.01" />
              </svg>
            </div>
            <h3>{errorInfo.title}</h3>
            <p>{errorInfo.message}</p>
            {errorInfo.action && (
              <div className="tx-error-action">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" width="18" height="18">
                  <circle cx="12" cy="12" r="10" />
                  <path d="M12 16v-4M12 8h.01" />
                </svg>
                <span>{errorInfo.action}</span>
              </div>
            )}
            <div className="tx-error-buttons">
              <button className="tx-retry-btn" onClick={() => { setStep('review'); setSlideProgress(0); setSlideCompleted(false); setErrorInfo(null); setPinError(null); }}>
                Try Again
              </button>
              <button className="tx-dismiss-btn" onClick={onCancel}>
                Dismiss
              </button>
            </div>
          </div>
        )}
      </div>

      <style>{styles}</style>
    </div>
  );
}

const styles = `
  .tx-overlay {
    position: fixed;
    inset: 0;
    z-index: 9999;
    display: flex;
    align-items: flex-end;
  }

  .tx-backdrop {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    animation: fadeIn 0.2s ease;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .tx-sheet {
    position: relative;
    width: 100%;
    max-height: 92vh;
    background: linear-gradient(180deg, #1a1f2e 0%, #0f1219 100%);
    border-radius: 28px 28px 0 0;
    padding: 12px 20px 32px;
    padding-bottom: max(32px, env(safe-area-inset-bottom));
    overflow-y: auto;
    animation: slideUp 0.35s cubic-bezier(0.32, 0.72, 0, 1);
    box-shadow: 0 -10px 40px rgba(0, 0, 0, 0.5);
  }

  @keyframes slideUp {
    from { transform: translateY(100%); }
    to { transform: translateY(0); }
  }

  .tx-drag-indicator {
    width: 40px;
    height: 4px;
    background: #3d4654;
    border-radius: 2px;
    margin: 0 auto 20px;
  }

  /* Hero Section */
  .tx-hero {
    text-align: center;
    padding: 20px 0 28px;
  }

  .tx-icon-wrap {
    width: 80px;
    height: 80px;
    border-radius: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 16px;
    animation: iconBounce 0.5s ease 0.2s both;
  }

  @keyframes iconBounce {
    0% { transform: scale(0); }
    60% { transform: scale(1.1); }
    100% { transform: scale(1); }
  }

  .tx-icon {
    width: 40px;
    height: 40px;
  }

  .tx-icon svg {
    width: 100%;
    height: 100%;
  }

  .tx-type-title {
    font-size: 15px;
    font-weight: 500;
    color: #8b95a8;
    margin: 0 0 12px;
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .tx-amount-display {
    display: flex;
    align-items: baseline;
    justify-content: center;
    gap: 10px;
  }

  .tx-amount-value {
    font-size: 48px;
    font-weight: 700;
    color: white;
    letter-spacing: -1px;
  }

  .tx-amount-token {
    font-size: 24px;
    font-weight: 600;
    color: #6b7689;
  }

  /* Details Card */
  .tx-details-card {
    background: #0f1318;
    border-radius: 16px;
    padding: 4px 16px;
    margin-bottom: 16px;
  }

  .tx-detail-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 0;
    border-bottom: 1px solid #1e242e;
  }

  .tx-detail-row:last-child {
    border-bottom: none;
  }

  .tx-detail-label {
    font-size: 14px;
    color: #6b7689;
  }

  .tx-detail-value {
    font-size: 14px;
    font-weight: 600;
    color: white;
  }

  .tx-detail-value.tx-address {
    font-family: 'SF Mono', 'Menlo', monospace;
    font-size: 13px;
    color: #60a5fa;
  }

  .tx-detail-value.tx-fee {
    color: #8b95a8;
  }

  /* Warning Box */
  .tx-warning-box {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 14px 16px;
    background: rgba(251, 191, 36, 0.08);
    border: 1px solid rgba(251, 191, 36, 0.2);
    border-radius: 14px;
    margin-bottom: 20px;
  }

  .tx-warning-icon {
    flex-shrink: 0;
    width: 20px;
    height: 20px;
    color: #fbbf24;
  }

  .tx-warning-icon svg {
    width: 100%;
    height: 100%;
  }

  .tx-warning-box span {
    font-size: 13px;
    color: #fbbf24;
    line-height: 1.5;
  }

  /* Slide to Confirm */
  .tx-slide-wrapper {
    padding: 8px 0 16px;
  }

  .tx-slide-track {
    position: relative;
    height: 64px;
    background: #1f2937;
    border-radius: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    user-select: none;
    -webkit-user-select: none;
    transition: background 0.15s ease;
  }

  .tx-slide-track.completed {
    background: rgba(16, 185, 129, 0.15);
  }

  .tx-slide-thumb {
    position: absolute;
    left: 4px;
    width: 56px;
    height: 56px;
    border-radius: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: grab;
    touch-action: none;
    transition: transform 0.05s linear;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
  }

  .tx-slide-thumb:active {
    cursor: grabbing;
  }

  .tx-slide-thumb svg {
    width: 24px;
    height: 24px;
    color: white;
  }

  .tx-slide-label {
    font-size: 16px;
    font-weight: 600;
    color: #6b7689;
    pointer-events: none;
    transition: opacity 0.15s ease;
  }

  /* Cancel Button */
  .tx-cancel-btn {
    width: 100%;
    padding: 16px;
    border: none;
    background: transparent;
    color: #6b7689;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    transition: color 0.2s;
  }

  .tx-cancel-btn:active {
    color: white;
  }

  /* Auth Section */
  .tx-auth-section {
    padding: 40px 0 20px;
    text-align: center;
  }

  .tx-auth-section h3 {
    font-size: 24px;
    font-weight: 700;
    color: white;
    margin: 0 0 8px;
  }

  .tx-auth-section p {
    font-size: 15px;
    color: #6b7689;
    margin: 0 0 32px;
  }

  .tx-pin-icon-wrap {
    width: 72px;
    height: 72px;
    background: rgba(59, 130, 246, 0.1);
    border-radius: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 20px;
  }

  .tx-pin-icon-wrap svg {
    width: 36px;
    height: 36px;
    color: #3b82f6;
  }

  .tx-pin-dots {
    display: flex;
    justify-content: center;
    gap: 14px;
    margin-bottom: 24px;
  }

  .tx-pin-input {
    width: 48px;
    height: 56px;
    border: 2px solid #2d3748;
    border-radius: 12px;
    background: #0f1318;
    font-size: 24px;
    font-weight: 700;
    color: white;
    text-align: center;
    outline: none;
    transition: all 0.2s ease;
    caret-color: transparent;
  }

  .tx-pin-input:focus {
    border-color: #3b82f6;
    box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.15);
  }

  .tx-pin-input.filled {
    border-color: #3b82f6;
    background: rgba(59, 130, 246, 0.08);
  }

  .tx-pin-input.error {
    border-color: #ef4444;
    animation: shakeInput 0.4s ease;
  }

  @keyframes shakeInput {
    0%, 100% { transform: translateX(0); }
    20%, 60% { transform: translateX(-6px); }
    40%, 80% { transform: translateX(6px); }
  }

  .tx-pin-error {
    font-size: 14px;
    color: #ef4444;
    margin: 0 0 20px;
  }

  .tx-face-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    width: 100%;
    padding: 16px;
    border: 1px solid #2d3748;
    border-radius: 14px;
    background: transparent;
    color: white;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    margin-bottom: 12px;
    transition: all 0.2s;
  }

  .tx-face-btn:active {
    background: #1f2937;
  }

  .tx-face-btn svg {
    width: 26px;
    height: 26px;
    color: #3b82f6;
  }

  .tx-back-btn {
    width: 100%;
    padding: 14px;
    border: none;
    background: transparent;
    color: #6b7689;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
  }

  /* Biometric */
  .tx-biometric {
    padding: 60px 0;
  }

  .tx-biometric h3 {
    margin-top: 24px;
  }

  .tx-face-icon {
    width: 120px;
    height: 120px;
    margin: 0 auto;
    position: relative;
  }

  .tx-face-icon svg {
    width: 100%;
    height: 100%;
  }

  .tx-scan-line {
    position: absolute;
    left: 20px;
    right: 20px;
    height: 2px;
    background: linear-gradient(90deg, transparent, #3b82f6, transparent);
    animation: scanLine 1.5s ease-in-out infinite;
  }

  @keyframes scanLine {
    0%, 100% { top: 20px; opacity: 0; }
    20% { opacity: 1; }
    80% { opacity: 1; }
    100% { top: 76px; opacity: 0; }
  }

  /* Processing */
  .tx-processing {
    padding: 80px 0;
    text-align: center;
  }

  .tx-processing h3 {
    font-size: 22px;
    font-weight: 700;
    color: white;
    margin: 24px 0 8px;
  }

  .tx-processing p {
    font-size: 15px;
    color: #6b7689;
    margin: 0;
  }

  .tx-spinner {
    width: 64px;
    height: 64px;
    border: 3px solid #2d3748;
    border-top-width: 3px;
    border-radius: 50%;
    margin: 0 auto;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* Success */
  .tx-success {
    padding: 80px 0;
    text-align: center;
    position: relative;
  }

  .tx-success h3 {
    font-size: 28px;
    font-weight: 700;
    color: white;
    margin: 24px 0 12px;
  }

  .tx-success p {
    font-size: 16px;
    color: #8b95a8;
    margin: 0;
    padding: 0 20px;
  }

  .tx-success-icon {
    width: 88px;
    height: 88px;
    background: rgba(16, 185, 129, 0.15);
    border-radius: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto;
    animation: successPop 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) both;
  }

  @keyframes successPop {
    0% { transform: scale(0); }
    100% { transform: scale(1); }
  }

  .tx-success-icon svg {
    width: 44px;
    height: 44px;
    color: #10b981;
    animation: checkDraw 0.4s ease 0.3s both;
  }

  @keyframes checkDraw {
    0% { stroke-dasharray: 50; stroke-dashoffset: 50; }
    100% { stroke-dashoffset: 0; }
  }

  /* Error */
  .tx-error {
    padding: 60px 0 20px;
    text-align: center;
  }

  .tx-error h3 {
    font-size: 24px;
    font-weight: 700;
    color: white;
    margin: 20px 0 8px;
  }

  .tx-error p {
    font-size: 15px;
    color: #8b95a8;
    margin: 0 0 16px;
    padding: 0 20px;
    line-height: 1.5;
  }

  .tx-error-icon {
    width: 80px;
    height: 80px;
    background: rgba(239, 68, 68, 0.12);
    border-radius: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto;
  }

  .tx-error-icon svg {
    width: 40px;
    height: 40px;
    color: #ef4444;
  }

  .tx-error-action {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 12px 16px;
    background: rgba(59, 130, 246, 0.08);
    border-radius: 12px;
    font-size: 14px;
    color: #60a5fa;
    margin: 0 0 24px;
  }

  .tx-error-action svg {
    flex-shrink: 0;
  }

  .tx-error-buttons {
    display: flex;
    gap: 12px;
  }

  .tx-retry-btn {
    flex: 1;
    padding: 16px;
    border: none;
    border-radius: 14px;
    background: #2d3748;
    color: white;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s;
  }

  .tx-retry-btn:active {
    background: #374151;
  }

  .tx-dismiss-btn {
    flex: 1;
    padding: 16px;
    border: none;
    border-radius: 14px;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
  }
`;
