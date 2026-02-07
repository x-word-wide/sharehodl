/**
 * Create Wallet Screen - Professional wallet creation flow
 * With numpad PIN entry like Telegram Wallet
 *
 * SECURITY: Includes PIN complexity validation to prevent weak passcodes
 * SECURITY: Uses SecureMnemonic for secure memory handling of recovery phrase
 */

import { useState, useCallback, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Wallet } from 'lucide-react';
import { useWalletStore, generateRandomWalletName } from '../services/walletStore';
import { validatePinComplexity, type PinValidationResult } from '../utils/security';
import { SecureMnemonic, scheduleSecureCleanup } from '../utils/secureMemory';

type Step = 'name' | 'pin' | 'confirm-pin' | 'mnemonic' | 'verify' | 'backup-confirm';
const PIN_LENGTH = 6;

// Quiz question type for seed phrase verification
interface QuizQuestion {
  position: number;      // 1-indexed word position
  correctWord: string;
  options: string[];     // 4 shuffled options including correct word
}

// Shuffle array utility
function shuffleArray<T>(array: T[]): T[] {
  const shuffled = [...array];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
}

// Generate verification quiz from mnemonic words
function generateVerificationQuiz(words: string[]): QuizQuestion[] {
  if (words.length < 12) return [];

  // Select 2 random positions (as user requested)
  const allPositions = words.map((_, i) => i);
  const selectedPositions = shuffleArray(allPositions).slice(0, 2).sort((a, b) => a - b);

  return selectedPositions.map(pos => {
    // Get 3 random decoy words from the mnemonic (excluding correct word)
    const decoys = words.filter((_, i) => i !== pos);
    const shuffledDecoys = shuffleArray(decoys).slice(0, 3);

    // Combine correct word with decoys and shuffle
    const options = shuffleArray([words[pos], ...shuffledDecoys]);

    return {
      position: pos + 1, // 1-indexed for display
      correctWord: words[pos],
      options
    };
  });
}

export function CreateWalletScreen() {
  const navigate = useNavigate();
  const { createWallet, completeWalletSetup, isLoading } = useWalletStore();
  const tg = window.Telegram?.WebApp;

  const [step, setStep] = useState<Step>('name');
  const [walletName, setWalletName] = useState(() => generateRandomWalletName());
  const [pin, setPin] = useState('');
  const [confirmPin, setConfirmPin] = useState('');
  // SECURITY: Use SecureMnemonic ref instead of useState to minimize memory exposure
  const secureMnemonicRef = useRef(new SecureMnemonic());
  const [mnemonicVersion, setMnemonicVersion] = useState(0); // Force re-render when mnemonic changes
  const [showMnemonic, setShowMnemonic] = useState(false);
  const [copied, setCopied] = useState(false);
  const [confirmed, setConfirmed] = useState(false);
  const [shake, setShake] = useState(false);
  const [pinValidation, setPinValidation] = useState<PinValidationResult | null>(null);
  const [showPinError, setShowPinError] = useState(false);

  // Quiz state for seed phrase verification
  const [quizQuestions, setQuizQuestions] = useState<QuizQuestion[]>([]);
  const [selectedAnswers, setSelectedAnswers] = useState<Record<number, string>>({});
  const [quizAttempts, setQuizAttempts] = useState(0);
  const MAX_QUIZ_ATTEMPTS = 3;

  // SECURITY: Clear sensitive data from memory on unmount
  useEffect(() => {
    const secureMnemonic = secureMnemonicRef.current;
    return scheduleSecureCleanup(secureMnemonic, {
      clear: () => {
        setPin('');
        setConfirmPin('');
      }
    });
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
          // PINs match, create wallet
          tg?.HapticFeedback?.notificationOccurred('success');
          try {
            const generatedMnemonic = await createWallet(newPin, walletName);
            // SECURITY: Store in SecureMnemonic instead of React state
            secureMnemonicRef.current.set(generatedMnemonic);
            setMnemonicVersion(v => v + 1); // Trigger re-render
            setStep('mnemonic');
          } catch (error) {
            const message = error instanceof Error ? error.message : 'Failed to create wallet';
            tg?.showAlert(message);
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
  }, [currentPin, step, pin, walletName, createWallet, tg, setCurrentPin, isLoading]);

  const handleCopy = async () => {
    // SECURITY: Warn user about clipboard risks before copying
    const doCopy = async () => {
      try {
        const mnemonicText = secureMnemonicRef.current.get();
        await navigator.clipboard.writeText(mnemonicText);
        setCopied(true);
        tg?.HapticFeedback?.notificationOccurred('success');

        // SECURITY: Auto-clear clipboard after 15 seconds (reduced for better security)
        setTimeout(async () => {
          try {
            await navigator.clipboard.writeText('');
          } catch {
            // Clipboard clear failed, acceptable fallback
          }
        }, 15000);

        setTimeout(() => setCopied(false), 2000);
      } catch {
        tg?.showAlert('Failed to copy');
      }
    };

    // Show warning on first copy
    if (tg?.showConfirm) {
      tg.showConfirm(
        'Clipboard data can be accessed by other apps. The clipboard will be cleared after 15 seconds. Continue?',
        (confirmed) => {
          if (confirmed) doCopy();
        }
      );
    } else {
      // Fallback if Telegram confirm not available
      doCopy();
    }
  };

  // Handle starting the verification quiz
  const handleStartVerification = () => {
    const currentWords = secureMnemonicRef.current.getWords();
    const questions = generateVerificationQuiz(currentWords);
    setQuizQuestions(questions);
    setSelectedAnswers({});
    setStep('verify');
    tg?.HapticFeedback?.impactOccurred('medium');
  };

  // Handle quiz answer selection
  const handleAnswerSelect = (position: number, answer: string) => {
    if (selectedAnswers[position] !== undefined) return; // Already answered

    const question = quizQuestions.find(q => q.position === position);
    const isCorrect = question?.correctWord === answer;

    setSelectedAnswers(prev => ({ ...prev, [position]: answer }));
    tg?.HapticFeedback?.impactOccurred(isCorrect ? 'light' : 'heavy');
  };

  // Check if all answers are correct
  const allAnswersCorrect = quizQuestions.length > 0 &&
    quizQuestions.every(q => selectedAnswers[q.position] === q.correctWord);

  // Check if all questions have been answered
  const allQuestionsAnswered = quizQuestions.length > 0 &&
    Object.keys(selectedAnswers).length === quizQuestions.length;

  // Handle quiz completion
  const handleQuizComplete = () => {
    if (allAnswersCorrect) {
      tg?.HapticFeedback?.notificationOccurred('success');
      setStep('backup-confirm');
    } else {
      const newAttempts = quizAttempts + 1;
      setQuizAttempts(newAttempts);

      if (newAttempts >= MAX_QUIZ_ATTEMPTS) {
        // Force user back to view mnemonic again
        tg?.HapticFeedback?.notificationOccurred('error');
        tg?.showAlert('Please review your recovery phrase again carefully.');
        setStep('mnemonic');
        setSelectedAnswers({});
        setQuizAttempts(0);
        setShowMnemonic(true); // Show mnemonic by default this time
      } else {
        // Reset for retry with new questions
        tg?.HapticFeedback?.notificationOccurred('warning');
        const currentWords = secureMnemonicRef.current.getWords();
        setQuizQuestions(generateVerificationQuiz(currentWords));
        setSelectedAnswers({});
      }
    }
  };

  const handleComplete = () => {
    if (!confirmed) {
      tg?.HapticFeedback?.notificationOccurred('warning');
      return;
    }
    // SECURITY: Clear sensitive data from memory before navigation
    secureMnemonicRef.current.clear();
    setPin('');
    setConfirmPin('');
    // Complete wallet setup - this unlocks the wallet
    completeWalletSetup();
    tg?.HapticFeedback?.notificationOccurred('success');
    navigate('/portfolio');
  };

  // SECURITY: Get words from secure container (mnemonicVersion ensures re-render)
  const words = mnemonicVersion > 0 ? secureMnemonicRef.current.getWords() : [];

  return (
    <div className="create-screen">
      {/* Progress indicator */}
      <div className="progress-bar">
        <div className="progress-track">
          <div
            className="progress-fill"
            style={{
              width: step === 'name' ? '16%' :
                     step === 'pin' ? '32%' :
                     step === 'confirm-pin' ? '48%' :
                     step === 'mnemonic' ? '64%' :
                     step === 'verify' ? '82%' : '100%'
            }}
          />
        </div>
      </div>

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
            onClick={() => {
              tg?.HapticFeedback?.impactOccurred('medium');
              setStep('pin');
            }}
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
        </div>
      )}

      {/* Mnemonic Step */}
      {step === 'mnemonic' && (
        <div className="mnemonic-step">
          <div className="mnemonic-header">
            <div className="mnemonic-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
              </svg>
            </div>
            <h1 className="mnemonic-title">Recovery Phrase</h1>
            <p className="mnemonic-subtitle">
              Write down these 24 words in order. This is the only way to recover your wallet.
            </p>
          </div>

          {/* Warning */}
          <div className="warning-card">
            <span className="warning-icon">⚠️</span>
            <span className="warning-text">Never share your recovery phrase with anyone!</span>
          </div>

          {/* Mnemonic grid */}
          <div className="mnemonic-container">
            <div className={`mnemonic-grid ${!showMnemonic ? 'blurred' : ''}`}>
              {words.map((word, i) => (
                <div key={i} className="word-item">
                  <span className="word-num">{i + 1}</span>
                  <span className="word-text">{word}</span>
                </div>
              ))}
            </div>

            {!showMnemonic && (
              <button className="reveal-btn" onClick={() => setShowMnemonic(true)}>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                  <circle cx="12" cy="12" r="3" />
                </svg>
                Tap to reveal
              </button>
            )}
          </div>

          {/* Actions */}
          <div className="mnemonic-actions">
            <button
              className="action-btn secondary"
              onClick={() => setShowMnemonic(!showMnemonic)}
            >
              {showMnemonic ? '🙈 Hide' : '👁️ Show'}
            </button>
            <button className="action-btn secondary" onClick={handleCopy}>
              {copied ? '✓ Copied!' : '📋 Copy'}
            </button>
          </div>

          <button className="continue-btn" onClick={handleStartVerification}>
            I've saved my recovery phrase
          </button>
        </div>
      )}

      {/* Verify Step - Seed Phrase Quiz */}
      {step === 'verify' && (
        <div className="verify-step">
          <div className="verify-header">
            <div className="verify-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M9 12l2 2 4-4" />
                <circle cx="12" cy="12" r="10" />
              </svg>
            </div>
            <h1 className="verify-title">Verify Your Backup</h1>
            <p className="verify-subtitle">
              Select the correct words to confirm you've saved your recovery phrase
            </p>
          </div>

          {quizAttempts > 0 && (
            <div className="quiz-attempts-warning">
              <span className="warning-icon">⚠️</span>
              <span>{MAX_QUIZ_ATTEMPTS - quizAttempts} attempt{MAX_QUIZ_ATTEMPTS - quizAttempts !== 1 ? 's' : ''} remaining</span>
            </div>
          )}

          <div className="quiz-questions">
            {quizQuestions.map((question) => {
              const selectedAnswer = selectedAnswers[question.position];
              const isAnswered = selectedAnswer !== undefined;
              const isCorrect = selectedAnswer === question.correctWord;

              return (
                <div key={question.position} className="quiz-question">
                  <p className="question-label">
                    Word #{question.position}
                  </p>
                  <div className="options-grid">
                    {question.options.map((option) => {
                      let optionClass = 'option-btn';
                      if (isAnswered) {
                        if (option === question.correctWord) {
                          optionClass += ' correct';
                        } else if (option === selectedAnswer) {
                          optionClass += ' incorrect';
                        } else {
                          optionClass += ' disabled';
                        }
                      }

                      return (
                        <button
                          key={option}
                          className={optionClass}
                          onClick={() => handleAnswerSelect(question.position, option)}
                          disabled={isAnswered}
                        >
                          {option}
                        </button>
                      );
                    })}
                  </div>
                  {isAnswered && (
                    <p className={`answer-feedback ${isCorrect ? 'correct' : 'incorrect'}`}>
                      {isCorrect ? '✓ Correct!' : `✗ Incorrect - was "${question.correctWord}"`}
                    </p>
                  )}
                </div>
              );
            })}
          </div>

          {allQuestionsAnswered && (
            <button
              className={`verify-continue-btn ${allAnswersCorrect ? 'success' : 'retry'}`}
              onClick={handleQuizComplete}
            >
              {allAnswersCorrect ? 'Continue' : `Try Again (${MAX_QUIZ_ATTEMPTS - quizAttempts} left)`}
            </button>
          )}

          <button
            className="back-to-mnemonic-btn"
            onClick={() => {
              setStep('mnemonic');
              setSelectedAnswers({});
            }}
          >
            ← Back to Recovery Phrase
          </button>
        </div>
      )}

      {/* Backup Confirm Step */}
      {step === 'backup-confirm' && (
        <div className="confirm-step">
          <div className="confirm-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M22 11.08V12a10 10 0 11-5.93-9.14" />
              <path d="M22 4L12 14.01l-3-3" />
            </svg>
          </div>
          <h1 className="confirm-title">Almost Done!</h1>
          <p className="confirm-subtitle">
            Please confirm that you've securely saved your recovery phrase.
          </p>

          <label className="checkbox-card">
            <input
              type="checkbox"
              checked={confirmed}
              onChange={(e) => setConfirmed(e.target.checked)}
            />
            <span className="checkbox-custom" />
            <span className="checkbox-text">
              I understand that if I lose my recovery phrase, I will lose access to my wallet and all assets forever.
            </span>
          </label>

          <button
            className={`complete-btn ${confirmed ? 'active' : ''}`}
            onClick={handleComplete}
            disabled={!confirmed}
          >
            Complete Setup
          </button>
        </div>
      )}

      <style>{`
        .create-screen {
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

        /* Name Step */
        .name-step {
          flex: 1;
          display: flex;
          flex-direction: column;
          padding: 20px 24px 40px;
        }

        .step-header {
          text-align: center;
          margin-bottom: 32px;
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

        /* Mnemonic Step */
        .mnemonic-step {
          flex: 1;
          display: flex;
          flex-direction: column;
          padding: 20px 24px 40px;
        }

        .mnemonic-header {
          text-align: center;
          margin-bottom: 20px;
        }

        .mnemonic-icon {
          width: 64px;
          height: 64px;
          margin: 0 auto 16px;
          border-radius: 50%;
          background: linear-gradient(135deg, #10b981 0%, #059669 100%);
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .mnemonic-icon svg {
          width: 28px;
          height: 28px;
          color: white;
        }

        .mnemonic-title {
          font-size: 24px;
          font-weight: 700;
          color: white;
          margin: 0 0 8px;
        }

        .mnemonic-subtitle {
          font-size: 14px;
          color: #8b949e;
          margin: 0;
          line-height: 1.5;
        }

        .warning-card {
          display: flex;
          align-items: center;
          gap: 10px;
          padding: 12px 16px;
          background: rgba(245, 158, 11, 0.1);
          border-radius: 12px;
          margin-bottom: 20px;
        }

        .warning-icon {
          font-size: 18px;
        }

        .warning-text {
          font-size: 13px;
          color: #f59e0b;
          font-weight: 500;
        }

        .mnemonic-container {
          position: relative;
          flex: 1;
          min-height: 280px;
        }

        .mnemonic-grid {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 8px;
          padding: 16px;
          background: #161B22;
          border-radius: 16px;
          transition: filter 0.3s ease;
        }

        .mnemonic-grid.blurred {
          filter: blur(8px);
        }

        .word-item {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 10px;
          background: #0D1117;
          border-radius: 10px;
        }

        .word-num {
          font-size: 11px;
          color: #8b949e;
          min-width: 18px;
        }

        .word-text {
          font-size: 13px;
          font-weight: 500;
          color: white;
        }

        .reveal-btn {
          position: absolute;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 14px 24px;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          border: none;
          border-radius: 24px;
          font-size: 15px;
          font-weight: 600;
          color: white;
          cursor: pointer;
          box-shadow: 0 4px 20px rgba(30, 64, 175, 0.4);
        }

        .reveal-btn svg {
          width: 20px;
          height: 20px;
        }

        .mnemonic-actions {
          display: flex;
          gap: 12px;
          margin: 16px 0;
        }

        .action-btn {
          flex: 1;
          padding: 14px;
          border: none;
          border-radius: 12px;
          font-size: 15px;
          font-weight: 600;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .action-btn.secondary {
          background: #161B22;
          color: white;
        }

        .action-btn:active {
          transform: scale(0.97);
        }

        .continue-btn {
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
        }

        /* Confirm Step */
        .confirm-step {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          padding: 60px 24px 40px;
          text-align: center;
        }

        .confirm-icon {
          width: 80px;
          height: 80px;
          border-radius: 50%;
          background: linear-gradient(135deg, #10b981 0%, #059669 100%);
          display: flex;
          align-items: center;
          justify-content: center;
          margin-bottom: 24px;
        }

        .confirm-icon svg {
          width: 40px;
          height: 40px;
          color: white;
        }

        .confirm-title {
          font-size: 28px;
          font-weight: 700;
          color: white;
          margin: 0 0 12px;
        }

        .confirm-subtitle {
          font-size: 15px;
          color: #8b949e;
          margin: 0 0 40px;
          line-height: 1.5;
        }

        .checkbox-card {
          display: flex;
          gap: 16px;
          padding: 20px;
          background: #161B22;
          border-radius: 16px;
          cursor: pointer;
          text-align: left;
          margin-bottom: auto;
        }

        .checkbox-card input {
          display: none;
        }

        .checkbox-custom {
          width: 24px;
          height: 24px;
          border-radius: 8px;
          border: 2px solid #30363d;
          flex-shrink: 0;
          transition: all 0.2s ease;
          position: relative;
        }

        .checkbox-card input:checked + .checkbox-custom {
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          border-color: transparent;
        }

        .checkbox-card input:checked + .checkbox-custom::after {
          content: '';
          position: absolute;
          left: 7px;
          top: 3px;
          width: 6px;
          height: 12px;
          border: solid white;
          border-width: 0 2px 2px 0;
          transform: rotate(45deg);
        }

        .checkbox-text {
          font-size: 14px;
          color: #8b949e;
          line-height: 1.5;
        }

        .complete-btn {
          width: 100%;
          padding: 16px;
          background: #30363d;
          border: none;
          border-radius: 14px;
          font-size: 16px;
          font-weight: 600;
          color: #8b949e;
          cursor: not-allowed;
          margin-top: 24px;
          transition: all 0.3s ease;
        }

        .complete-btn.active {
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          color: white;
          cursor: pointer;
        }

        /* Verify Step */
        .verify-step {
          flex: 1;
          display: flex;
          flex-direction: column;
          padding: 20px 24px 40px;
        }

        .verify-header {
          text-align: center;
          margin-bottom: 24px;
        }

        .verify-icon {
          width: 64px;
          height: 64px;
          margin: 0 auto 16px;
          border-radius: 50%;
          background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .verify-icon svg {
          width: 32px;
          height: 32px;
          color: white;
        }

        .verify-title {
          font-size: 24px;
          font-weight: 700;
          color: white;
          margin: 0 0 8px;
        }

        .verify-subtitle {
          font-size: 14px;
          color: #8b949e;
          margin: 0;
          line-height: 1.5;
        }

        .quiz-attempts-warning {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          padding: 12px;
          background: rgba(239, 68, 68, 0.1);
          border-radius: 12px;
          margin-bottom: 20px;
          font-size: 14px;
          color: #ef4444;
        }

        .quiz-questions {
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: 20px;
          margin-bottom: 20px;
        }

        .quiz-question {
          background: #161B22;
          border-radius: 16px;
          padding: 16px;
        }

        .question-label {
          font-size: 14px;
          font-weight: 600;
          color: #8b949e;
          margin: 0 0 12px;
        }

        .options-grid {
          display: grid;
          grid-template-columns: repeat(2, 1fr);
          gap: 10px;
        }

        .option-btn {
          padding: 14px 12px;
          background: #0D1117;
          border: 2px solid #30363d;
          border-radius: 12px;
          font-size: 15px;
          font-weight: 500;
          color: white;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .option-btn:active:not(:disabled) {
          transform: scale(0.97);
        }

        .option-btn.correct {
          border-color: #10b981;
          background: rgba(16, 185, 129, 0.15);
          color: #10b981;
        }

        .option-btn.incorrect {
          border-color: #ef4444;
          background: rgba(239, 68, 68, 0.15);
          color: #ef4444;
        }

        .option-btn.disabled {
          opacity: 0.4;
          cursor: not-allowed;
        }

        .answer-feedback {
          margin: 12px 0 0;
          font-size: 13px;
          font-weight: 500;
          text-align: center;
        }

        .answer-feedback.correct {
          color: #10b981;
        }

        .answer-feedback.incorrect {
          color: #ef4444;
        }

        .verify-continue-btn {
          width: 100%;
          padding: 16px;
          border: none;
          border-radius: 14px;
          font-size: 16px;
          font-weight: 600;
          color: white;
          cursor: pointer;
          margin-bottom: 12px;
          transition: all 0.2s ease;
        }

        .verify-continue-btn.success {
          background: linear-gradient(135deg, #10b981 0%, #059669 100%);
        }

        .verify-continue-btn.retry {
          background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
        }

        .verify-continue-btn:active {
          transform: scale(0.98);
        }

        .back-to-mnemonic-btn {
          width: 100%;
          padding: 14px;
          background: transparent;
          border: 1px solid #30363d;
          border-radius: 14px;
          font-size: 14px;
          font-weight: 500;
          color: #8b949e;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .back-to-mnemonic-btn:active {
          background: rgba(48, 54, 61, 0.3);
        }
      `}</style>
    </div>
  );
}
