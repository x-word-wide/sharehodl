/**
 * Unlock Screen - Professional PIN entry with numpad
 * Inspired by Telegram Wallet passcode design
 *
 * SECURITY: Includes brute force protection with lockout display
 */

import { useState, useCallback, useEffect } from 'react';
import { useWalletStore } from '../services/walletStore';

const PIN_LENGTH = 6;
const BIOMETRIC_ENABLED_KEY = 'sh_biometric_enabled';

export function UnlockScreen() {
  const {
    unlockWallet,
    isLoading,
    error,
    clearError,
    resetWallet,
    securityState,
    remainingAttempts,
    refreshSecurityState
  } = useWalletStore();
  const tg = window.Telegram?.WebApp;

  const [pin, setPin] = useState('');
  const [shake, setShake] = useState(false);
  const [lockoutTimer, setLockoutTimer] = useState(0);

  // Biometric state
  const [biometricEnabled, setBiometricEnabled] = useState(false);
  const [biometricType, setBiometricType] = useState<string>('Biometric');
  const [biometricLoading, setBiometricLoading] = useState(false);

  // Track if we've already auto-triggered biometric
  const [autoTriggered, setAutoTriggered] = useState(false);

  // Check biometric availability and status on mount
  useEffect(() => {
    const enabled = localStorage.getItem(BIOMETRIC_ENABLED_KEY) === 'true';
    setBiometricEnabled(enabled);

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const biometricManager = (tg as any)?.BiometricManager;
    if (biometricManager) {
      biometricManager.init(() => {
        if (biometricManager.biometricType) {
          const type = biometricManager.biometricType;
          setBiometricType(type === 'face' ? 'Face ID' : type === 'finger' ? 'Touch ID' : 'Biometric');
        }
      });
    }
  }, [tg]);

  // Update lockout timer countdown
  useEffect(() => {
    refreshSecurityState();

    if (securityState.isLocked) {
      setLockoutTimer(Math.ceil(securityState.lockoutRemainingMs / 1000));

      const interval = setInterval(() => {
        setLockoutTimer(prev => {
          if (prev <= 1) {
            refreshSecurityState();
            return 0;
          }
          return prev - 1;
        });
      }, 1000);

      return () => clearInterval(interval);
    }
  }, [securityState.isLocked, securityState.lockoutRemainingMs, refreshSecurityState]);

  const formatTime = (seconds: number): string => {
    if (seconds < 60) return `${seconds}s`;
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const handleKeyPress = useCallback(async (key: string) => {
    if (isLoading || securityState.isLocked) return;

    tg?.HapticFeedback?.impactOccurred('light');

    if (key === 'delete') {
      setPin(prev => prev.slice(0, -1));
      clearError();
      return;
    }

    if (pin.length >= PIN_LENGTH) return;

    const newPin = pin + key;
    setPin(newPin);

    // Auto-submit when PIN is complete
    if (newPin.length === PIN_LENGTH) {
      try {
        await unlockWallet(newPin);
        tg?.HapticFeedback?.notificationOccurred('success');
      } catch {
        tg?.HapticFeedback?.notificationOccurred('error');
        setShake(true);
        setTimeout(() => {
          setShake(false);
          setPin('');
        }, 500);
      }
    }
  }, [pin, isLoading, unlockWallet, clearError, tg, securityState.isLocked]);

  const handleReset = () => {
    tg?.showConfirm(
      'This will delete your wallet from this device. Make sure you have your recovery phrase backed up.',
      (confirmed) => {
        if (confirmed) {
          resetWallet();
          tg?.HapticFeedback?.notificationOccurred('warning');
        }
      }
    );
  };

  const handleBiometric = useCallback(async () => {
    if (!biometricEnabled) {
      tg?.HapticFeedback?.impactOccurred('medium');
      tg?.showAlert(`${biometricType} is not enabled. Enable it in Settings.`);
      return;
    }

    tg?.HapticFeedback?.impactOccurred('medium');
    setBiometricLoading(true);
    clearError();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const biometricManager = (tg as any)?.BiometricManager;

    if (!biometricManager) {
      tg?.showAlert(`${biometricType} is only available in the Telegram app.`);
      setBiometricLoading(false);
      return;
    }

    // Timeout to prevent infinite loading
    const timeout = setTimeout(() => {
      setBiometricLoading(false);
      tg?.showAlert(`${biometricType} timed out. Please use PIN.`);
    }, 30000);

    try {
      // SECURITY: Only use the token from Telegram's secure biometric storage
      // Never fall back to localStorage as it can be compromised by XSS
      biometricManager.authenticate(
        { reason: 'Unlock your wallet' },
        async (success: boolean, token?: string) => {
          clearTimeout(timeout);

          if (success) {
            // The token from Telegram's secure storage IS the PIN
            if (token && token.length > 0) {
              try {
                await unlockWallet(token);
                tg?.HapticFeedback?.notificationOccurred('success');
              } catch {
                tg?.HapticFeedback?.notificationOccurred('error');
                tg?.showAlert(`${biometricType} failed. Please use PIN.`);
              }
            } else {
              // Token not returned - biometric storage may not be set up correctly
              tg?.HapticFeedback?.notificationOccurred('error');
              tg?.showAlert(`${biometricType} not configured. Please re-enable in Settings.`);
              // Reset biometric state
              localStorage.setItem(BIOMETRIC_ENABLED_KEY, 'false');
              localStorage.removeItem('sh_bio_pin'); // Clean up any old insecure storage
              setBiometricEnabled(false);
            }
          } else {
            // User cancelled or biometric failed
            tg?.HapticFeedback?.notificationOccurred('error');
          }
          setBiometricLoading(false);
        }
      );
    } catch {
      clearTimeout(timeout);
      tg?.showAlert(`${biometricType} error. Please use PIN.`);
      setBiometricLoading(false);
    }
  }, [biometricEnabled, biometricType, tg, clearError, unlockWallet]);

  // Auto-trigger biometric after a short delay when screen opens
  useEffect(() => {
    // Only auto-trigger if:
    // - Biometric is enabled
    // - Not already triggered
    // - Not locked out
    // - Not already loading
    if (!biometricEnabled || autoTriggered || securityState.isLocked || biometricLoading) {
      return;
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const biometricManager = (tg as any)?.BiometricManager;
    if (!biometricManager) {
      return;
    }

    // Wait for screen to render, then auto-trigger Face ID
    const timer = setTimeout(() => {
      setAutoTriggered(true);
      handleBiometric();
    }, 600); // 600ms delay for smooth UX

    return () => clearTimeout(timer);
  }, [biometricEnabled, autoTriggered, securityState.isLocked, biometricLoading, tg, handleBiometric]);

  const isLocked = securityState.isLocked && lockoutTimer > 0;

  return (
    <div className="pin-screen">
      {/* Background gradient */}
      <div className="pin-bg">
        <div className="pin-gradient" />
      </div>

      {/* Content */}
      <div className="pin-content">
        {/* Logo */}
        <div className="pin-logo">
          <div className={`logo-circle ${isLocked ? 'locked' : ''}`}>
            {isLocked ? (
              <svg className="lock-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                <path d="M7 11V7a5 5 0 0110 0v4" />
                <circle cx="12" cy="16" r="1" />
              </svg>
            ) : (
              <svg className="lock-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                <path d="M7 11V7a5 5 0 0110 0v4" />
              </svg>
            )}
          </div>
        </div>

        {isLocked ? (
          <>
            <h1 className="pin-title">Account Locked</h1>
            <p className="pin-subtitle">Too many failed attempts</p>
            <div className="lockout-timer">
              <div className="timer-circle">
                <span className="timer-text">{formatTime(lockoutTimer)}</span>
              </div>
              <p className="timer-label">Try again in</p>
            </div>
          </>
        ) : (
          <>
            <h1 className="pin-title">Welcome Back</h1>
            <p className="pin-subtitle">Enter your passcode</p>

            {/* PIN Dots */}
            <div className={`pin-dots ${shake ? 'shake' : ''}`}>
              {Array.from({ length: PIN_LENGTH }).map((_, i) => (
                <div
                  key={i}
                  className={`pin-dot ${i < pin.length ? 'filled' : ''} ${
                    isLoading && i < pin.length ? 'loading' : ''
                  }`}
                />
              ))}
            </div>

            {/* Error message with remaining attempts */}
            {error && (
              <div className="pin-error">
                <span>{error}</span>
                {remainingAttempts <= 3 && remainingAttempts > 0 && (
                  <span className="attempts-warning">
                    {remainingAttempts} attempt{remainingAttempts !== 1 ? 's' : ''} remaining before lockout
                  </span>
                )}
                {securityState.failedAttempts >= 3 && (
                  <button className="reset-link" onClick={handleReset}>
                    Forgot passcode?
                  </button>
                )}
              </div>
            )}

            {/* Numpad */}
            <div className="numpad">
              {['1', '2', '3', '4', '5', '6', '7', '8', '9', 'bio', '0', 'delete'].map((key) => (
                <button
                  key={key}
                  className={`numpad-key ${key === 'bio' || key === 'delete' ? 'action' : ''} ${key === 'bio' && biometricEnabled ? 'bio-enabled' : ''} ${key === 'bio' && biometricLoading ? 'bio-loading' : ''}`}
                  onClick={() => key === 'bio' ? handleBiometric() : handleKeyPress(key)}
                  disabled={isLoading || isLocked || (key === 'bio' && biometricLoading)}
                >
                  {key === 'delete' ? (
                    <svg className="delete-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M21 4H8l-7 8 7 8h13a2 2 0 002-2V6a2 2 0 00-2-2z" />
                      <line x1="18" y1="9" x2="12" y2="15" />
                      <line x1="12" y1="9" x2="18" y2="15" />
                    </svg>
                  ) : key === 'bio' ? (
                    biometricLoading ? (
                      <div className="bio-spinner" />
                    ) : (
                      <svg className="bio-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M12 11c0 3.517-1.009 6.799-2.753 9.571m-3.44-2.04l.054-.09A13.916 13.916 0 008 11a4 4 0 118 0c0 1.017-.07 2.019-.203 3m-2.118 6.844A21.88 21.88 0 0015.171 17m3.839 1.132c.645-2.266.99-4.659.99-7.132A8 8 0 008 4.07M3 15.364c.64-1.319 1-2.8 1-4.364 0-1.457.39-2.823 1.07-4" />
                      </svg>
                    )
                  ) : (
                    <span className="key-number">{key}</span>
                  )}
                </button>
              ))}
            </div>
          </>
        )}
      </div>

      <style>{`
        .pin-screen {
          min-height: 100vh;
          display: flex;
          flex-direction: column;
          position: relative;
          overflow: hidden;
          background-color: var(--tg-theme-bg-color);
        }

        .pin-bg {
          position: absolute;
          inset: 0;
          z-index: 0;
        }

        .pin-gradient {
          position: absolute;
          inset: 0;
          background: radial-gradient(ellipse at top, rgba(30, 64, 175, 0.1) 0%, transparent 50%);
        }

        .pin-content {
          position: relative;
          z-index: 1;
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          padding: 60px 24px 40px;
        }

        .pin-logo {
          margin-bottom: 24px;
        }

        .logo-circle {
          width: 80px;
          height: 80px;
          border-radius: 50%;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          display: flex;
          align-items: center;
          justify-content: center;
          box-shadow: 0 16px 40px rgba(30, 64, 175, 0.3);
          transition: all 0.3s ease;
        }

        .logo-circle.locked {
          background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
          box-shadow: 0 16px 40px rgba(239, 68, 68, 0.3);
        }

        .lock-icon {
          width: 32px;
          height: 32px;
          color: white;
        }

        .pin-title {
          font-size: 24px;
          font-weight: 700;
          color: var(--text-primary);
          margin: 0 0 8px;
        }

        .pin-subtitle {
          font-size: 15px;
          color: var(--text-secondary);
          margin: 0 0 40px;
        }

        .lockout-timer {
          display: flex;
          flex-direction: column;
          align-items: center;
          margin-top: 20px;
        }

        .timer-circle {
          width: 120px;
          height: 120px;
          border-radius: 50%;
          background: rgba(239, 68, 68, 0.1);
          border: 3px solid rgba(239, 68, 68, 0.3);
          display: flex;
          align-items: center;
          justify-content: center;
          margin-bottom: 16px;
        }

        .timer-text {
          font-size: 32px;
          font-weight: 700;
          color: #f87171;
          font-variant-numeric: tabular-nums;
        }

        .timer-label {
          font-size: 14px;
          color: var(--text-secondary);
          margin: 0;
        }

        .pin-dots {
          display: flex;
          gap: 16px;
          margin-bottom: 24px;
        }

        .pin-dots.shake {
          animation: shake 0.5s ease-in-out;
        }

        @keyframes shake {
          0%, 100% { transform: translateX(0); }
          20%, 60% { transform: translateX(-10px); }
          40%, 80% { transform: translateX(10px); }
        }

        .pin-dot {
          width: 16px;
          height: 16px;
          border-radius: 50%;
          background: var(--pin-dot-bg);
          transition: all 0.15s ease;
        }

        .pin-dot.filled {
          background: linear-gradient(135deg, #1E40AF, #3B82F6);
          transform: scale(1.1);
          box-shadow: 0 0 12px rgba(30, 64, 175, 0.5);
        }

        .pin-dot.loading {
          animation: pulse 0.8s ease-in-out infinite;
        }

        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }

        .pin-error {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 8px;
          margin-bottom: 24px;
          color: #f87171;
          font-size: 14px;
          text-align: center;
        }

        .attempts-warning {
          color: #fbbf24;
          font-size: 12px;
        }

        .reset-link {
          background: none;
          border: none;
          color: #3B82F6;
          font-size: 14px;
          cursor: pointer;
          padding: 0;
        }

        .numpad {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 16px;
          width: 100%;
          max-width: 280px;
          margin-top: auto;
          justify-items: center;
        }

        .numpad-key {
          width: 72px;
          height: 72px;
          border-radius: 50%;
          border: none;
          background: var(--numpad-bg);
          color: var(--text-primary);
          font-size: 28px;
          font-weight: 500;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.15s ease;
          -webkit-tap-highlight-color: transparent;
          user-select: none;
        }

        .numpad-key:active {
          background: rgba(30, 64, 175, 0.3);
          transform: scale(0.95);
        }

        .numpad-key:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .numpad-key.action {
          background: transparent;
        }

        .numpad-key.action:active {
          background: var(--input-bg);
        }

        .key-number {
          display: flex;
          align-items: center;
          justify-content: center;
          line-height: 1;
        }

        .delete-icon,
        .bio-icon {
          width: 24px;
          height: 24px;
          color: var(--text-secondary);
        }

        .numpad-key.bio-enabled .bio-icon {
          color: #3B82F6;
        }

        .numpad-key.bio-enabled {
          background: rgba(30, 64, 175, 0.1);
        }

        .numpad-key.bio-loading {
          opacity: 0.7;
        }

        .bio-spinner {
          width: 24px;
          height: 24px;
          border: 2px solid var(--pin-dot-bg);
          border-top-color: #3B82F6;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
}
