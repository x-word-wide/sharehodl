/**
 * Import Wallet Screen - Import existing wallet with mnemonic
 * Professional PIN entry with numpad like CreateWalletScreen
 */

import { useState, useCallback, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { AlertCircle, Wallet } from 'lucide-react';
import { useWalletStore, generateRandomWalletName } from '../services/walletStore';
import { validateMnemonic } from '../utils/crypto';
import { validatePinComplexity, type PinValidationResult } from '../utils/security';
import { SecureMnemonic } from '../utils/secureMemory';

type Step = 'mnemonic' | 'name' | 'pin' | 'confirm-pin';
const PIN_LENGTH = 6;

export function ImportWalletScreen() {
  const navigate = useNavigate();
  const { importWallet, isLoading, error, clearError } = useWalletStore();
  const tg = window.Telegram?.WebApp;

  const [step, setStep] = useState<Step>('mnemonic');
  // SECURITY: Use SecureMnemonic for secure memory handling instead of plain useState
  const secureMnemonicRef = useRef<SecureMnemonic>(new SecureMnemonic());
  const [mnemonicInput, setMnemonicInput] = useState('');  // Only for display in input field
  const [walletName, setWalletName] = useState(() => generateRandomWalletName());
  const [pin, setPin] = useState('');
  const [confirmPin, setConfirmPin] = useState('');
  const [mnemonicError, setMnemonicError] = useState('');
  const [shake, setShake] = useState(false);
  const [pinValidation, setPinValidation] = useState<PinValidationResult | null>(null);
  const [showPinError, setShowPinError] = useState(false);

  // SECURITY: Clean up secure mnemonic on unmount
  useEffect(() => {
    const secureMnemonic = secureMnemonicRef.current;
    return () => {
      secureMnemonic.clear();
    };
  }, []);

  // Validate PIN as user types
  useEffect(() => {
    if (step === 'pin' && pin.length > 0) {
      setPinValidation(validatePinComplexity(pin));
    } else {
      setPinValidation(null);
      setShowPinError(false);
    }
  }, [pin, step]);

  const handleMnemonicSubmit = () => {
    clearError();
    setMnemonicError('');

    const cleanMnemonic = mnemonicInput.trim().toLowerCase();
    const words = cleanMnemonic.split(/\s+/);

    if (words.length !== 12 && words.length !== 24) {
      setMnemonicError('Please enter 12 or 24 words');
      return;
    }

    if (!validateMnemonic(cleanMnemonic)) {
      setMnemonicError('Invalid recovery phrase');
      return;
    }

    // SECURITY: Store in secure mnemonic and clear the input state
    secureMnemonicRef.current.set(cleanMnemonic);
    setMnemonicInput('');  // Clear the input display for security

    tg?.HapticFeedback?.impactOccurred('medium');
    setStep('name');
  };

  const handleNameSubmit = () => {
    tg?.HapticFeedback?.impactOccurred('medium');
    setStep('pin');
  };

  const currentPin = step === 'pin' ? pin : confirmPin;
  const setCurrentPin = step === 'pin' ? setPin : setConfirmPin;

  const handleKeyPress = useCallback(async (key: string) => {
    if (isLoading) return;
    tg?.HapticFeedback?.impactOccurred('light');

    if (key === 'delete') {
      setCurrentPin(prev => prev.slice(0, -1));
      return;
    }

    if (currentPin.length >= PIN_LENGTH) return;

    const newPin = currentPin + key;
    setCurrentPin(newPin);

    // Auto-advance when PIN is complete
    if (newPin.length === PIN_LENGTH) {
      if (step === 'pin') {
        // Validate PIN complexity before advancing
        const validation = validatePinComplexity(newPin);
        if (!validation.isValid) {
          tg?.HapticFeedback?.notificationOccurred('error');
          setShake(true);
          setShowPinError(true);
          setTimeout(() => {
            setShake(false);
            setPin('');
          }, 500);
          return;
        }
        setTimeout(() => setStep('confirm-pin'), 200);
      } else if (step === 'confirm-pin') {
        if (newPin === pin) {
          // PINs match, import wallet
          tg?.HapticFeedback?.notificationOccurred('success');
          try {
            const mnemonic = secureMnemonicRef.current.get();
            await importWallet(mnemonic, newPin, walletName);
            // SECURITY: Clear mnemonic from memory after import
            secureMnemonicRef.current.clear();
            navigate('/portfolio');
          } catch (e) {
            tg?.HapticFeedback?.notificationOccurred('error');
          }
        } else {
          // PINs don't match
          tg?.HapticFeedback?.notificationOccurred('error');
          setShake(true);
          setTimeout(() => {
            setShake(false);
            setConfirmPin('');
          }, 500);
        }
      }
    }
  }, [currentPin, step, pin, walletName, importWallet, tg, setCurrentPin, isLoading, navigate]);

  return (
    <div className="import-screen">
      {/* Progress indicator */}
      <div className="progress-bar">
        <div className="progress-track">
          <div
            className="progress-fill"
            style={{
              width: step === 'mnemonic' ? '25%' :
                     step === 'name' ? '50%' :
                     step === 'pin' ? '75%' : '100%'
            }}
          />
        </div>
      </div>

      {/* Mnemonic Step */}
      {step === 'mnemonic' && (
        <div className="mnemonic-step">
          <div className="step-header">
            <div className="step-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
              </svg>
            </div>
            <h1 className="step-title">Import Wallet</h1>
            <p className="step-subtitle">Enter your 12 or 24 word recovery phrase</p>
          </div>

          <div className="mnemonic-input-area">
            <textarea
              value={mnemonicInput}
              onChange={(e) => {
                setMnemonicInput(e.target.value);
                setMnemonicError('');
              }}
              placeholder="Enter your recovery phrase (words separated by spaces)"
              className="mnemonic-textarea"
              autoComplete="off"
              autoCorrect="off"
              autoCapitalize="off"
              spellCheck={false}
            />

            {mnemonicError && (
              <div className="error-message">
                <AlertCircle size={16} />
                <span>{mnemonicError}</span>
              </div>
            )}

            <div className="tips-card">
              <h3>Tips</h3>
              <ul>
                <li>Enter words separated by spaces</li>
                <li>Make sure all words are spelled correctly</li>
                <li>Words are case-insensitive</li>
              </ul>
            </div>
          </div>

          <button
            onClick={handleMnemonicSubmit}
            disabled={!mnemonicInput.trim()}
            className="continue-btn"
          >
            Continue
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M9 18l6-6-6-6" />
            </svg>
          </button>
        </div>
      )}

      {/* Name Step */}
      {step === 'name' && (
        <div className="name-step">
          <div className="step-header">
            <div className="step-icon">
              <Wallet size={28} color="white" />
            </div>
            <h1 className="step-title">Name Your Wallet</h1>
            <p className="step-subtitle">Give your wallet a memorable name</p>
          </div>

          <div className="name-input-area">
            <input
              type="text"
              value={walletName}
              onChange={(e) => setWalletName(e.target.value)}
              placeholder="Enter wallet name"
              className="name-input"
              autoComplete="off"
              autoCorrect="off"
              maxLength={30}
            />
            <p className="name-hint">
              You can change this later in Settings
            </p>
          </div>

          <button
            onClick={handleNameSubmit}
            disabled={!walletName.trim()}
            className="continue-btn"
          >
            Continue
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M9 18l6-6-6-6" />
            </svg>
          </button>
        </div>
      )}

      {/* PIN Steps */}
      {(step === 'pin' || step === 'confirm-pin') && (
        <div className="pin-step">
          <div className="pin-header">
            <div className="pin-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                <path d="M7 11V7a5 5 0 0110 0v4" />
              </svg>
            </div>
            <h1 className="pin-title">
              {step === 'pin' ? 'Create Passcode' : 'Confirm Passcode'}
            </h1>
            <p className="pin-subtitle">
              {step === 'pin'
                ? 'Set a 6-digit passcode to secure your wallet'
                : 'Enter your passcode again to confirm'
              }
            </p>
          </div>

          {/* PIN dots */}
          <div className={`pin-dots ${shake ? 'shake' : ''}`}>
            {Array.from({ length: PIN_LENGTH }).map((_, i) => (
              <div
                key={i}
                className={`pin-dot ${i < currentPin.length ? 'filled' : ''}`}
              />
            ))}
          </div>

          {/* PIN validation error */}
          {step === 'pin' && showPinError && pinValidation && !pinValidation.isValid && (
            <div className="pin-validation-error">
              <p className="pin-error">{pinValidation.errors[0]}</p>
              <p className="pin-hint">Try a unique combination of digits</p>
            </div>
          )}

          {/* PIN strength indicator */}
          {step === 'pin' && pin.length === PIN_LENGTH && pinValidation?.isValid && (
            <div className={`pin-strength ${pinValidation.strength}`}>
              <span className="strength-dot" />
              <span className="strength-text">
                {pinValidation.strength === 'strong' ? 'Strong passcode' : 'Acceptable passcode'}
              </span>
            </div>
          )}

          {step === 'confirm-pin' && shake && (
            <p className="pin-error">Passcodes don't match. Try again.</p>
          )}

          {/* Numpad */}
          <div className="numpad">
            {['1', '2', '3', '4', '5', '6', '7', '8', '9', '', '0', 'delete'].map((key) => (
              key === '' ? (
                <div key="empty" className="numpad-spacer" />
              ) : (
                <button
                  key={key}
                  className={`numpad-key ${key === 'delete' ? 'action' : ''}`}
                  onClick={() => handleKeyPress(key)}
                  disabled={isLoading}
                >
                  {key === 'delete' ? (
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M21 4H8l-7 8 7 8h13a2 2 0 002-2V6a2 2 0 00-2-2z" />
                      <line x1="18" y1="9" x2="12" y2="15" />
                      <line x1="12" y1="9" x2="18" y2="15" />
                    </svg>
                  ) : (
                    <span>{key}</span>
                  )}
                </button>
              )
            ))}
          </div>

          {isLoading && (
            <div className="loading-overlay">
              <div className="spinner" />
              <p>Importing wallet...</p>
            </div>
          )}
        </div>
      )}

      {error && (
        <div className="global-error">
          <p>{error}</p>
        </div>
      )}

      <style>{`
        .import-screen {
          min-height: 100vh;
          display: flex;
          flex-direction: column;
        }

        .progress-bar {
          padding: 16px 24px;
        }

        .progress-track {
          height: 4px;
          background: #30363d;
          border-radius: 2px;
          overflow: hidden;
        }

        .progress-fill {
          height: 100%;
          background: linear-gradient(90deg, #1E40AF, #3B82F6);
          border-radius: 2px;
          transition: width 0.3s ease;
        }

        /* Mnemonic Step */
        .mnemonic-step {
          flex: 1;
          display: flex;
          flex-direction: column;
          padding: 20px 24px 40px;
        }

        .step-header {
          text-align: center;
          margin-bottom: 24px;
        }

        .step-icon {
          width: 64px;
          height: 64px;
          margin: 0 auto 16px;
          border-radius: 50%;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .step-icon svg {
          width: 28px;
          height: 28px;
          color: white;
        }

        .step-title {
          font-size: 24px;
          font-weight: 700;
          color: white;
          margin: 0 0 8px;
        }

        .step-subtitle {
          font-size: 14px;
          color: #8b949e;
          margin: 0;
        }

        .mnemonic-input-area {
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: 16px;
        }

        .mnemonic-textarea {
          width: 100%;
          min-height: 140px;
          padding: 16px;
          background: #161B22;
          border: 1px solid #30363d;
          border-radius: 14px;
          color: white;
          font-size: 15px;
          line-height: 1.5;
          resize: none;
          outline: none;
          transition: border-color 0.2s ease;
        }

        .mnemonic-textarea:focus {
          border-color: #3B82F6;
        }

        .mnemonic-textarea::placeholder {
          color: #8b949e;
        }

        .error-message {
          display: flex;
          align-items: center;
          gap: 8px;
          color: #f87171;
          font-size: 14px;
        }

        .tips-card {
          padding: 16px;
          background: #161B22;
          border-radius: 14px;
        }

        .tips-card h3 {
          font-size: 14px;
          font-weight: 600;
          color: white;
          margin: 0 0 10px;
        }

        .tips-card ul {
          margin: 0;
          padding: 0;
          list-style: none;
        }

        .tips-card li {
          font-size: 13px;
          color: #8b949e;
          padding: 4px 0;
          padding-left: 16px;
          position: relative;
        }

        .tips-card li::before {
          content: '-';
          position: absolute;
          left: 0;
        }

        /* Name Step */
        .name-step {
          flex: 1;
          display: flex;
          flex-direction: column;
          padding: 20px 24px 40px;
        }

        .name-input-area {
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .name-input {
          width: 100%;
          padding: 16px;
          background: #161B22;
          border: 1px solid #30363d;
          border-radius: 14px;
          color: white;
          font-size: 18px;
          font-weight: 500;
          outline: none;
          transition: border-color 0.2s ease;
        }

        .name-input:focus {
          border-color: #3B82F6;
        }

        .name-input::placeholder {
          color: #8b949e;
        }

        .name-hint {
          font-size: 13px;
          color: #8b949e;
          text-align: center;
          margin: 0;
        }

        .continue-btn {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          width: 100%;
          padding: 16px;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          border: none;
          border-radius: 14px;
          font-size: 16px;
          font-weight: 600;
          color: white;
          cursor: pointer;
          margin-top: auto;
          transition: all 0.2s ease;
        }

        .continue-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .continue-btn:not(:disabled):active {
          transform: scale(0.98);
        }

        .continue-btn svg {
          width: 20px;
          height: 20px;
        }

        /* PIN Step */
        .pin-step {
          flex: 1;
          display: flex;
          flex-direction: column;
          padding: 20px 24px 40px;
        }

        .pin-header {
          text-align: center;
          margin-bottom: 32px;
        }

        .pin-icon {
          width: 64px;
          height: 64px;
          margin: 0 auto 16px;
          border-radius: 50%;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .pin-icon svg {
          width: 28px;
          height: 28px;
          color: white;
        }

        .pin-title {
          font-size: 24px;
          font-weight: 700;
          color: white;
          margin: 0 0 8px;
        }

        .pin-subtitle {
          font-size: 14px;
          color: #8b949e;
          margin: 0;
        }

        .pin-dots {
          display: flex;
          justify-content: center;
          gap: 16px;
          margin-bottom: 16px;
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
          background: #30363d;
          transition: all 0.15s ease;
        }

        .pin-dot.filled {
          background: linear-gradient(135deg, #1E40AF, #3B82F6);
          transform: scale(1.1);
          box-shadow: 0 0 12px rgba(30, 64, 175, 0.5);
        }

        .pin-error {
          text-align: center;
          color: #f87171;
          font-size: 14px;
          margin: 0;
        }

        .pin-validation-error {
          text-align: center;
          margin-bottom: 16px;
        }

        .pin-hint {
          text-align: center;
          color: #8b949e;
          font-size: 12px;
          margin: 4px 0 0;
        }

        .pin-strength {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          margin-bottom: 16px;
        }

        .strength-dot {
          width: 8px;
          height: 8px;
          border-radius: 50%;
        }

        .pin-strength.strong .strength-dot {
          background: #10b981;
        }

        .pin-strength.medium .strength-dot {
          background: #f59e0b;
        }

        .strength-text {
          font-size: 13px;
          color: #8b949e;
        }

        .pin-strength.strong .strength-text {
          color: #10b981;
        }

        .pin-strength.medium .strength-text {
          color: #f59e0b;
        }

        .numpad {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 12px;
          margin-top: auto;
          max-width: 280px;
          margin-left: auto;
          margin-right: auto;
          justify-items: center;
        }

        .numpad-key {
          width: 72px;
          height: 72px;
          border-radius: 50%;
          border: none;
          background: rgba(48, 54, 61, 0.5);
          color: white;
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

        .numpad-key span {
          display: flex;
          align-items: center;
          justify-content: center;
          line-height: 1;
        }

        .numpad-key:active {
          background: rgba(30, 64, 175, 0.3);
          transform: scale(0.95);
        }

        .numpad-key:disabled {
          opacity: 0.5;
        }

        .numpad-key.action {
          background: transparent;
        }

        .numpad-key svg {
          width: 24px;
          height: 24px;
          color: #8b949e;
        }

        .numpad-spacer {
          width: 72px;
          height: 72px;
        }

        .loading-overlay {
          position: fixed;
          inset: 0;
          background: rgba(13, 17, 23, 0.9);
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 16px;
          z-index: 100;
        }

        .loading-overlay .spinner {
          width: 40px;
          height: 40px;
          border: 3px solid #30363d;
          border-top-color: #3B82F6;
          border-radius: 50%;
          animation: spin 1s linear infinite;
        }

        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }

        .loading-overlay p {
          color: #8b949e;
          font-size: 14px;
        }

        .global-error {
          position: fixed;
          bottom: 100px;
          left: 16px;
          right: 16px;
          padding: 14px 16px;
          background: rgba(239, 68, 68, 0.15);
          border: 1px solid rgba(239, 68, 68, 0.3);
          border-radius: 12px;
          z-index: 50;
        }

        .global-error p {
          color: #f87171;
          font-size: 14px;
          margin: 0;
          text-align: center;
        }
      `}</style>
    </div>
  );
}
