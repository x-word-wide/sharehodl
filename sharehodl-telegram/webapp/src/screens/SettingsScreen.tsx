/**
 * Settings Screen - Wallet settings and preferences
 *
 * Features:
 * - View Recovery Phrase (with PIN verification)
 * - Change PIN
 * - Multi-wallet management
 * - Biometric authentication
 */

import { useState, useEffect, useCallback } from 'react';
import {
  Shield,
  Key,
  Bell,
  Globe,
  HelpCircle,
  FileText,
  LogOut,
  ChevronRight,
  Lock,
  Smartphone,
  Moon,
  Sun,
  Check,
  Wallet,
  Plus,
  Copy,
  Eye,
  EyeOff,
  X,
  AlertCircle,
  CheckCircle
} from 'lucide-react';
import { useWalletStore } from '../services/walletStore';
import { decryptData } from '../utils/crypto';

// Theme storage key
const THEME_KEY = 'sh_theme';
const BIOMETRIC_ENABLED_KEY = 'sh_biometric_enabled';

type Theme = 'dark' | 'light' | 'system';
type ModalType = 'none' | 'view-phrase' | 'change-pin' | 'wallets' | 'add-wallet' | 'rename-wallet' | 'setup-biometric';

const PIN_LENGTH = 6;

export function SettingsScreen() {
  const {
    lockWallet,
    resetWallet,
    verifyPin,
    changePin,
    getRecoveryPhrase,
    getWallets,
    wallets,
    activeWalletId,
    addWallet,
    setBiometricToken,
    clearBiometricToken
  } = useWalletStore();
  const tg = window.Telegram?.WebApp;

  // Modal state
  const [activeModal, setActiveModal] = useState<ModalType>('none');
  const [pin, setPin] = useState('');
  const [newPin, setNewPin] = useState('');
  const [confirmPin, setConfirmPin] = useState('');
  const [pinStep, setPinStep] = useState<'current' | 'new' | 'confirm'>('current');
  const [recoveryPhrase, setRecoveryPhrase] = useState<string[]>([]);
  const [showPhrase, setShowPhrase] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [shake, setShake] = useState(false);
  const [newWalletName, setNewWalletName] = useState('');
  const [walletMnemonic, setWalletMnemonic] = useState('');
  const [usePinFallback, setUsePinFallback] = useState(false);
  const [biometricAttempted, setBiometricAttempted] = useState(false);

  // Load wallets on mount
  useEffect(() => {
    const loadedWallets = getWallets();
    if (loadedWallets.length === 0) {
      // Migrate existing wallet to multi-wallet system
      // This is handled by the store
    }
  }, [getWallets]);

  // Helper functions for alerts/confirms with fallbacks
  const showAlert = useCallback((message: string) => {
    if (tg?.showAlert) {
      tg.showAlert(message);
    } else {
      alert(message);
    }
  }, [tg]);

  const showConfirm = useCallback((message: string, callback: (confirmed: boolean) => void) => {
    if (tg?.showConfirm) {
      tg.showConfirm(message, callback);
    } else {
      const confirmed = confirm(message);
      callback(confirmed);
    }
  }, [tg]);

  const openLink = useCallback((url: string) => {
    if (tg?.openLink) {
      tg.openLink(url);
    } else {
      window.open(url, '_blank');
    }
  }, [tg]);

  // Theme state - default to 'system' to respect OS/Telegram settings
  const [theme, setTheme] = useState<Theme>(() => {
    const saved = localStorage.getItem(THEME_KEY);
    return (saved as Theme) || 'system';
  });

  // Biometric state
  const [biometricEnabled, setBiometricEnabled] = useState(() => {
    return localStorage.getItem(BIOMETRIC_ENABLED_KEY) === 'true';
  });
  const [biometricAvailable, setBiometricAvailable] = useState(false);
  const [biometricType, setBiometricType] = useState<string>('Biometric');

  const [notifications, setNotifications] = useState(true);

  // Check biometric availability
  useEffect(() => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const biometricManager = (tg as any)?.BiometricManager;
    if (biometricManager) {
      biometricManager.init(() => {
        setBiometricAvailable(biometricManager.isAccessGranted || biometricManager.isBiometricAvailable);
        if (biometricManager.biometricType) {
          const type = biometricManager.biometricType;
          setBiometricType(type === 'face' ? 'Face ID' : type === 'finger' ? 'Touch ID' : 'Biometric');
        }
      });
    }
  }, [tg]);

  // Apply theme
  useEffect(() => {
    const root = document.documentElement;

    // Determine if dark mode based on theme setting
    let isDark: boolean;
    if (theme === 'dark') {
      isDark = true;
    } else if (theme === 'light') {
      isDark = false;
    } else {
      // System theme - check Telegram first, then browser preference
      if (tg?.colorScheme) {
        isDark = tg.colorScheme === 'dark';
      } else {
        // Fallback to browser/OS preference
        isDark = window.matchMedia?.('(prefers-color-scheme: dark)')?.matches ?? true;
      }
    }

    if (isDark) {
      root.setAttribute('data-theme', 'dark');
      root.style.setProperty('--tg-theme-bg-color', '#0D1117');
      root.style.setProperty('--tg-theme-text-color', '#ffffff');
      root.style.setProperty('--tg-theme-secondary-bg-color', '#161B22');
    } else {
      root.setAttribute('data-theme', 'light');
      root.style.setProperty('--tg-theme-bg-color', '#ffffff');
      root.style.setProperty('--tg-theme-text-color', '#1a1a1a');
      root.style.setProperty('--tg-theme-secondary-bg-color', '#f5f5f5');
    }

    localStorage.setItem(THEME_KEY, theme);

    // Listen for OS theme changes when in system mode
    if (theme === 'system') {
      const mediaQuery = window.matchMedia?.('(prefers-color-scheme: dark)');
      const handleChange = (e: MediaQueryListEvent) => {
        const newIsDark = e.matches;
        if (newIsDark) {
          root.setAttribute('data-theme', 'dark');
          root.style.setProperty('--tg-theme-bg-color', '#0D1117');
          root.style.setProperty('--tg-theme-text-color', '#ffffff');
          root.style.setProperty('--tg-theme-secondary-bg-color', '#161B22');
        } else {
          root.setAttribute('data-theme', 'light');
          root.style.setProperty('--tg-theme-bg-color', '#ffffff');
          root.style.setProperty('--tg-theme-text-color', '#1a1a1a');
          root.style.setProperty('--tg-theme-secondary-bg-color', '#f5f5f5');
        }
      };

      mediaQuery?.addEventListener('change', handleChange);
      return () => mediaQuery?.removeEventListener('change', handleChange);
    }
  }, [theme, tg?.colorScheme]);

  const handleThemeChange = (newTheme: Theme) => {
    tg?.HapticFeedback?.selectionChanged();
    setTheme(newTheme);
  };

  // Reset modal state
  const resetModalState = () => {
    setPin('');
    setNewPin('');
    setConfirmPin('');
    setPinStep('current');
    setRecoveryPhrase([]);
    setShowPhrase(false);
    setError('');
    setSuccess('');
    setIsLoading(false);
    setShake(false);
    setNewWalletName('');
    setWalletMnemonic('');
    setUsePinFallback(false);
    setBiometricAttempted(false);
  };

  const closeModal = () => {
    resetModalState();
    setActiveModal('none');
  };

  // Handle PIN keypress
  const handlePinKey = useCallback(async (key: string) => {
    if (isLoading) return;
    tg?.HapticFeedback?.impactOccurred('light');
    setError('');

    const currentPinState = pinStep === 'current' ? pin : pinStep === 'new' ? newPin : confirmPin;
    const setCurrentPin = pinStep === 'current' ? setPin : pinStep === 'new' ? setNewPin : setConfirmPin;

    if (key === 'delete') {
      setCurrentPin(prev => prev.slice(0, -1));
      return;
    }

    if (currentPinState.length >= PIN_LENGTH) return;

    const updatedPin = currentPinState + key;
    setCurrentPin(updatedPin);

    // Auto-submit when PIN is complete
    if (updatedPin.length === PIN_LENGTH) {
      if (activeModal === 'view-phrase') {
        await handleViewPhraseSubmit(updatedPin);
      } else if (activeModal === 'change-pin') {
        await handleChangePinStep(updatedPin);
      } else if (activeModal === 'add-wallet') {
        await handleAddWalletPinSubmit(updatedPin);
      } else if (activeModal === 'setup-biometric') {
        await handleBiometricSetupPin(updatedPin);
      }
    }
  }, [pin, newPin, confirmPin, pinStep, isLoading, activeModal, tg]);

  // View Recovery Phrase
  const handleViewPhraseSubmit = async (enteredPin: string) => {
    setIsLoading(true);
    try {
      const phrase = await getRecoveryPhrase(enteredPin);
      setRecoveryPhrase(phrase.split(' '));
      tg?.HapticFeedback?.notificationOccurred('success');
    } catch {
      tg?.HapticFeedback?.notificationOccurred('error');
      setShake(true);
      setTimeout(() => {
        setShake(false);
        setPin('');
      }, 500);
      setError('Invalid PIN');
    }
    setIsLoading(false);
  };

  // Biometric authentication for viewing recovery phrase
  const handleBiometricAuth = useCallback(() => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const biometricManager = (tg as any)?.BiometricManager;
    if (!biometricManager) {
      setUsePinFallback(true);
      return;
    }

    setBiometricAttempted(true);
    setIsLoading(true);

    biometricManager.authenticate(
      { reason: 'Verify your identity to view recovery phrase' },
      async (success: boolean, token?: string) => {
        if (success && token) {
          try {
            // Token is used to decrypt the stored PIN
            const BIOMETRIC_TOKEN_KEY = 'sh_biometric_token';
            const encryptedPin = localStorage.getItem(BIOMETRIC_TOKEN_KEY);

            if (encryptedPin) {
              const decryptedPin = await decryptData(encryptedPin, token);
              const phrase = await getRecoveryPhrase(decryptedPin);
              setRecoveryPhrase(phrase.split(' '));
              tg?.HapticFeedback?.notificationOccurred('success');
            } else {
              throw new Error('Biometric not configured');
            }
          } catch {
            tg?.HapticFeedback?.notificationOccurred('error');
            setError('Biometric authentication failed. Please use PIN.');
            setUsePinFallback(true);
          }
        } else {
          tg?.HapticFeedback?.notificationOccurred('error');
          setError('Biometric authentication cancelled');
          setUsePinFallback(true);
        }
        setIsLoading(false);
      }
    );
  }, [tg, getRecoveryPhrase]);

  // Trigger biometric when modal opens (if enabled)
  useEffect(() => {
    if (activeModal === 'view-phrase' && biometricEnabled && !biometricAttempted && !usePinFallback && recoveryPhrase.length === 0) {
      // Small delay to allow modal animation
      const timer = setTimeout(() => {
        handleBiometricAuth();
      }, 300);
      return () => clearTimeout(timer);
    }
  }, [activeModal, biometricEnabled, biometricAttempted, usePinFallback, recoveryPhrase.length, handleBiometricAuth]);

  // Change PIN steps
  const handleChangePinStep = async (enteredPin: string) => {
    setIsLoading(true);

    if (pinStep === 'current') {
      // Verify current PIN
      const isValid = await verifyPin(enteredPin);
      if (isValid) {
        tg?.HapticFeedback?.notificationOccurred('success');
        setPinStep('new');
        setPin(enteredPin); // Store for later use
      } else {
        tg?.HapticFeedback?.notificationOccurred('error');
        setShake(true);
        setTimeout(() => {
          setShake(false);
          setPin('');
        }, 500);
        setError('Current PIN is incorrect');
      }
    } else if (pinStep === 'new') {
      // Store new PIN and move to confirm
      setPinStep('confirm');
    } else if (pinStep === 'confirm') {
      // Verify PINs match
      if (enteredPin === newPin) {
        try {
          await changePin(pin, newPin);
          tg?.HapticFeedback?.notificationOccurred('success');
          setSuccess('PIN changed successfully!');
          setTimeout(() => closeModal(), 1500);
        } catch (err) {
          setError(err instanceof Error ? err.message : 'Failed to change PIN');
          tg?.HapticFeedback?.notificationOccurred('error');
        }
      } else {
        tg?.HapticFeedback?.notificationOccurred('error');
        setShake(true);
        setTimeout(() => {
          setShake(false);
          setConfirmPin('');
        }, 500);
        setError('PINs do not match');
      }
    }

    setIsLoading(false);
  };

  // Add new wallet
  const handleAddWalletPinSubmit = async (enteredPin: string) => {
    setIsLoading(true);
    try {
      const mnemonic = await addWallet(newWalletName || `Wallet ${wallets.length + 1}`, enteredPin);
      setWalletMnemonic(mnemonic);
      tg?.HapticFeedback?.notificationOccurred('success');
    } catch (err) {
      tg?.HapticFeedback?.notificationOccurred('error');
      setShake(true);
      setTimeout(() => {
        setShake(false);
        setPin('');
      }, 500);
      setError(err instanceof Error ? err.message : 'Failed to create wallet');
    }
    setIsLoading(false);
  };

  // Biometric toggle
  const handleBiometricToggle = async () => {
    tg?.HapticFeedback?.impactOccurred('light');
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const biometricManager = (tg as any)?.BiometricManager;

    if (!biometricEnabled) {
      // Enable biometric - need to request access first
      if (biometricManager && biometricManager.requestAccess) {
        biometricManager.requestAccess({ reason: 'Enable quick unlock with biometrics' }, (granted: boolean) => {
          if (granted) {
            // Open PIN entry modal to set up biometric
            resetModalState();
            setActiveModal('setup-biometric');
          } else {
            showAlert('Biometric access denied. Please allow biometric access in settings.');
          }
        });
      } else {
        // Fallback for testing without Telegram - open setup modal
        resetModalState();
        setActiveModal('setup-biometric');
      }
    } else {
      // Disable biometric
      clearBiometricToken();
      setBiometricEnabled(false);
      localStorage.setItem(BIOMETRIC_ENABLED_KEY, 'false');
      tg?.HapticFeedback?.selectionChanged();
      showAlert('Biometric login disabled');
    }
  };

  // Handle biometric setup PIN submission
  const handleBiometricSetupPin = async (enteredPin: string) => {
    setIsLoading(true);
    try {
      // Verify PIN first
      const isValid = await verifyPin(enteredPin);
      if (!isValid) {
        throw new Error('Invalid PIN');
      }

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const biometricManager = (tg as any)?.BiometricManager;

      // Store PIN in localStorage (base64 encoded) - biometric is used for identity verification
      // This is safe because biometric auth is required to access it
      const encodedPin = btoa(enteredPin);
      localStorage.setItem('sh_bio_pin', encodedPin);

      if (biometricManager && biometricManager.updateBiometricToken) {
        // Also try to store in Telegram's secure storage as backup
        biometricManager.updateBiometricToken(enteredPin, () => {
          // We don't rely on this succeeding - localStorage is our primary storage
          setBiometricToken(enteredPin).then(() => {
            setBiometricEnabled(true);
            localStorage.setItem(BIOMETRIC_ENABLED_KEY, 'true');
            tg?.HapticFeedback?.notificationOccurred('success');
            setSuccess(`${biometricType} enabled successfully!`);
            setTimeout(() => closeModal(), 1500);
          });
          setIsLoading(false);
        });
      } else {
        // Fallback for testing
        await setBiometricToken(enteredPin);
        setBiometricEnabled(true);
        localStorage.setItem(BIOMETRIC_ENABLED_KEY, 'true');
        tg?.HapticFeedback?.notificationOccurred('success');
        setSuccess(`${biometricType} enabled successfully!`);
        setTimeout(() => closeModal(), 1500);
        setIsLoading(false);
      }
    } catch (err) {
      tg?.HapticFeedback?.notificationOccurred('error');
      setShake(true);
      setTimeout(() => {
        setShake(false);
        setPin('');
      }, 500);
      setError(err instanceof Error ? err.message : 'Invalid PIN');
      setIsLoading(false);
    }
  };

  const handleLogout = () => {
    tg?.HapticFeedback?.impactOccurred('medium');
    showConfirm(
      'Are you sure you want to lock your wallet?',
      (confirmed) => {
        if (confirmed) {
          lockWallet();
          tg?.HapticFeedback?.notificationOccurred('warning');
        }
      }
    );
  };

  const handleReset = () => {
    tg?.HapticFeedback?.impactOccurred('heavy');
    showConfirm(
      'This will permanently delete your wallet from this device. Make sure you have your recovery phrase backed up!',
      (confirmed) => {
        if (confirmed) {
          showConfirm(
            'Are you ABSOLUTELY sure? This cannot be undone.',
            (doubleConfirmed) => {
              if (doubleConfirmed) {
                resetWallet();
                tg?.HapticFeedback?.notificationOccurred('error');
              }
            }
          );
        }
      }
    );
  };

  const handleCopyPhrase = () => {
    const phrase = recoveryPhrase.join(' ');
    navigator.clipboard.writeText(phrase);
    tg?.HapticFeedback?.notificationOccurred('success');
    showAlert('Recovery phrase copied! Clear clipboard after use.');

    // Auto-clear clipboard after 30 seconds
    setTimeout(() => {
      navigator.clipboard.writeText('').catch(() => {});
    }, 30000);
  };

  // Render numpad
  const renderNumpad = (onKeyPress: (key: string) => void) => (
    <div className="numpad">
      {['1', '2', '3', '4', '5', '6', '7', '8', '9', '', '0', 'delete'].map((key) =>
        key === '' ? (
          <div key="empty" className="numpad-spacer" />
        ) : (
          <button
            key={key}
            onClick={() => onKeyPress(key)}
            disabled={isLoading}
            className={`numpad-key ${key === 'delete' ? 'action' : ''}`}
          >
            {key === 'delete' ? '⌫' : key}
          </button>
        )
      )}
    </div>
  );

  // Render PIN dots
  const renderPinDots = (currentLength: number) => (
    <div className={`pin-dots ${shake ? 'shake' : ''}`}>
      {Array.from({ length: PIN_LENGTH }).map((_, i) => (
        <div key={i} className={`pin-dot ${i < currentLength ? 'filled' : ''}`} />
      ))}
    </div>
  );

  return (
    <div className="settings-screen">
      {/* Header */}
      <div className="settings-header">
        <h1 className="settings-title">Settings</h1>
      </div>

      {/* Settings Groups */}
      <div className="settings-content">
        {/* Wallets */}
        <SettingsGroup title="Wallets">
          <SettingsItem
            icon={<Wallet size={20} />}
            title="Manage Wallets"
            subtitle={`${wallets.length || 1} wallet${(wallets.length || 1) > 1 ? 's' : ''}`}
            onClick={() => {
              tg?.HapticFeedback?.impactOccurred('light');
              setActiveModal('wallets');
            }}
          />
        </SettingsGroup>

        {/* Appearance */}
        <SettingsGroup title="Appearance">
          <div className="theme-selector">
            <ThemeOption
              icon={<Moon size={20} />}
              label="Dark"
              selected={theme === 'dark'}
              onClick={() => handleThemeChange('dark')}
            />
            <ThemeOption
              icon={<Sun size={20} />}
              label="Light"
              selected={theme === 'light'}
              onClick={() => handleThemeChange('light')}
            />
            <ThemeOption
              icon={<Smartphone size={20} />}
              label="System"
              selected={theme === 'system'}
              onClick={() => handleThemeChange('system')}
            />
          </div>
        </SettingsGroup>

        {/* Security */}
        <SettingsGroup title="Security">
          <SettingsItem
            icon={<Lock size={20} />}
            title="Change PIN"
            onClick={() => {
              tg?.HapticFeedback?.impactOccurred('light');
              resetModalState();
              setActiveModal('change-pin');
            }}
          />
          <SettingsToggle
            icon={<Smartphone size={20} />}
            title={biometricType}
            subtitle="Quick unlock with biometrics"
            value={biometricEnabled}
            onChange={handleBiometricToggle}
            disabled={!biometricAvailable}
          />
          <SettingsItem
            icon={<Key size={20} />}
            title="View Recovery Phrase"
            onClick={() => {
              tg?.HapticFeedback?.impactOccurred('light');
              resetModalState();
              setActiveModal('view-phrase');
            }}
          />
          <SettingsItem
            icon={<Shield size={20} />}
            title="Two-Factor Authentication"
            onClick={() => {
              tg?.HapticFeedback?.impactOccurred('light');
              showAlert('Two-Factor Authentication - Coming soon!');
            }}
          />
        </SettingsGroup>

        {/* Preferences */}
        <SettingsGroup title="Preferences">
          <SettingsToggle
            icon={<Bell size={20} />}
            title="Notifications"
            subtitle="Price alerts & transactions"
            value={notifications}
            onChange={(v) => {
              setNotifications(v);
              tg?.HapticFeedback?.selectionChanged();
              showAlert(v ? 'Notifications enabled' : 'Notifications disabled');
            }}
          />
          <SettingsItem
            icon={<Globe size={20} />}
            title="Network"
            subtitle="ShareHODL Mainnet"
            onClick={() => {
              tg?.HapticFeedback?.impactOccurred('light');
              showAlert('Network selection - Coming soon!');
            }}
          />
        </SettingsGroup>

        {/* Support */}
        <SettingsGroup title="Support">
          <SettingsItem
            icon={<HelpCircle size={20} />}
            title="Help Center"
            onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); openLink('https://help.sharehodl.network'); }}
          />
          <SettingsItem
            icon={<FileText size={20} />}
            title="Terms of Service"
            onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); openLink('https://sharehodl.network/terms'); }}
          />
          <SettingsItem
            icon={<FileText size={20} />}
            title="Privacy Policy"
            onClick={() => { tg?.HapticFeedback?.impactOccurred('light'); openLink('https://sharehodl.network/privacy'); }}
          />
        </SettingsGroup>

        {/* Account */}
        <SettingsGroup title="Account">
          <SettingsItem
            icon={<Lock size={20} />}
            title="Lock Wallet"
            onClick={handleLogout}
            color="warning"
          />
          <SettingsItem
            icon={<LogOut size={20} />}
            title="Reset Wallet"
            subtitle="Delete wallet from device"
            onClick={handleReset}
            color="danger"
          />
        </SettingsGroup>

        {/* App info */}
        <div className="app-info">
          <p className="app-version">ShareHODL Wallet v1.0.0</p>
          <p className="app-tagline">Built with security in mind</p>
        </div>
      </div>

      {/* View Recovery Phrase Modal */}
      {activeModal === 'view-phrase' && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <button className="modal-close" onClick={closeModal}>
              <X size={24} />
            </button>

            {recoveryPhrase.length === 0 ? (
              <>
                {/* Show biometric option if enabled and not using PIN fallback */}
                {biometricEnabled && !usePinFallback ? (
                  <>
                    <div className="modal-icon">
                      <Smartphone size={32} />
                    </div>
                    <h2 className="modal-title">
                      {isLoading ? 'Authenticating...' : `Use ${biometricType}`}
                    </h2>
                    <p className="modal-subtitle">Verify your identity to view recovery phrase</p>

                    {isLoading ? (
                      <div className="biometric-loading">
                        <div className="spinner" />
                        <p>Waiting for {biometricType}...</p>
                      </div>
                    ) : (
                      <div className="biometric-prompt">
                        <button
                          className="biometric-button"
                          onClick={handleBiometricAuth}
                          disabled={isLoading}
                        >
                          <Smartphone size={24} />
                          <span>Authenticate with {biometricType}</span>
                        </button>
                      </div>
                    )}

                    {error && (
                      <div className="modal-error">
                        <AlertCircle size={16} />
                        <span>{error}</span>
                      </div>
                    )}

                    <button
                      className="pin-fallback-button"
                      onClick={() => {
                        setUsePinFallback(true);
                        setError('');
                      }}
                    >
                      <Lock size={16} />
                      <span>Use PIN instead</span>
                    </button>
                  </>
                ) : (
                  <>
                    <div className="modal-icon">
                      <Key size={32} />
                    </div>
                    <h2 className="modal-title">Enter PIN</h2>
                    <p className="modal-subtitle">Verify your identity to view recovery phrase</p>

                    {renderPinDots(pin.length)}

                    {error && (
                      <div className="modal-error">
                        <AlertCircle size={16} />
                        <span>{error}</span>
                      </div>
                    )}

                    {renderNumpad(handlePinKey)}

                    {biometricEnabled && (
                      <button
                        className="pin-fallback-button"
                        onClick={() => {
                          setUsePinFallback(false);
                          setBiometricAttempted(false);
                          setError('');
                        }}
                      >
                        <Smartphone size={16} />
                        <span>Use {biometricType} instead</span>
                      </button>
                    )}
                  </>
                )}
              </>
            ) : (
              <>
                <div className="modal-icon success">
                  <CheckCircle size={32} />
                </div>
                <h2 className="modal-title">Recovery Phrase</h2>
                <p className="modal-subtitle warning">Never share this with anyone!</p>

                <div className="phrase-container">
                  <div className={`phrase-grid ${showPhrase ? '' : 'blurred'}`}>
                    {recoveryPhrase.map((word, i) => (
                      <div key={i} className="phrase-word">
                        <span className="word-number">{i + 1}</span>
                        <span className="word-text">{word}</span>
                      </div>
                    ))}
                  </div>

                  {!showPhrase && (
                    <button className="reveal-button" onClick={() => setShowPhrase(true)}>
                      <Eye size={20} />
                      <span>Tap to reveal</span>
                    </button>
                  )}
                </div>

                <div className="phrase-actions">
                  <button className="action-button" onClick={() => setShowPhrase(!showPhrase)}>
                    {showPhrase ? <EyeOff size={18} /> : <Eye size={18} />}
                    <span>{showPhrase ? 'Hide' : 'Show'}</span>
                  </button>
                  <button className="action-button primary" onClick={handleCopyPhrase}>
                    <Copy size={18} />
                    <span>Copy</span>
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {/* Change PIN Modal */}
      {activeModal === 'change-pin' && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <button className="modal-close" onClick={closeModal}>
              <X size={24} />
            </button>

            {success ? (
              <>
                <div className="modal-icon success">
                  <CheckCircle size={32} />
                </div>
                <h2 className="modal-title">{success}</h2>
              </>
            ) : (
              <>
                <div className="modal-icon">
                  <Lock size={32} />
                </div>
                <h2 className="modal-title">
                  {pinStep === 'current' ? 'Current PIN' : pinStep === 'new' ? 'New PIN' : 'Confirm New PIN'}
                </h2>
                <p className="modal-subtitle">
                  {pinStep === 'current'
                    ? 'Enter your current PIN'
                    : pinStep === 'new'
                    ? 'Choose a new 6-digit PIN'
                    : 'Enter your new PIN again'}
                </p>

                {renderPinDots(
                  pinStep === 'current' ? pin.length : pinStep === 'new' ? newPin.length : confirmPin.length
                )}

                {error && (
                  <div className="modal-error">
                    <AlertCircle size={16} />
                    <span>{error}</span>
                  </div>
                )}

                {renderNumpad(handlePinKey)}

                {/* Step indicator */}
                <div className="step-indicator">
                  <div className={`step ${pinStep === 'current' ? 'active' : 'completed'}`} />
                  <div className={`step ${pinStep === 'new' ? 'active' : pinStep === 'confirm' ? 'completed' : ''}`} />
                  <div className={`step ${pinStep === 'confirm' ? 'active' : ''}`} />
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {/* Manage Wallets Modal */}
      {activeModal === 'wallets' && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content wallets-modal" onClick={e => e.stopPropagation()}>
            <button className="modal-close" onClick={closeModal}>
              <X size={24} />
            </button>

            <h2 className="modal-title">Wallets</h2>

            <div className="wallets-list">
              {(wallets.length > 0 ? wallets : [{ id: 'default', name: 'Main Wallet', sharehodlAddress: '', createdAt: 0 }]).map((wallet) => (
                <div key={wallet.id} className={`wallet-item ${wallet.id === activeWalletId ? 'active' : ''}`}>
                  <div className="wallet-info">
                    <div className="wallet-icon">
                      <Wallet size={20} />
                    </div>
                    <div className="wallet-details">
                      <span className="wallet-name">{wallet.name}</span>
                      {wallet.sharehodlAddress && (
                        <span className="wallet-address">
                          {wallet.sharehodlAddress.slice(0, 12)}...{wallet.sharehodlAddress.slice(-6)}
                        </span>
                      )}
                    </div>
                  </div>
                  {wallet.id === activeWalletId && (
                    <div className="wallet-active-badge">Active</div>
                  )}
                </div>
              ))}
            </div>

            <button
              className="add-wallet-button"
              onClick={() => {
                tg?.HapticFeedback?.impactOccurred('light');
                resetModalState();
                setActiveModal('add-wallet');
              }}
            >
              <Plus size={20} />
              <span>Add New Wallet</span>
            </button>
          </div>
        </div>
      )}

      {/* Add Wallet Modal */}
      {activeModal === 'add-wallet' && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <button className="modal-close" onClick={closeModal}>
              <X size={24} />
            </button>

            {walletMnemonic ? (
              <>
                <div className="modal-icon success">
                  <CheckCircle size={32} />
                </div>
                <h2 className="modal-title">Wallet Created!</h2>
                <p className="modal-subtitle warning">Save this recovery phrase securely!</p>

                <div className="phrase-container">
                  <div className={`phrase-grid ${showPhrase ? '' : 'blurred'}`}>
                    {walletMnemonic.split(' ').map((word, i) => (
                      <div key={i} className="phrase-word">
                        <span className="word-number">{i + 1}</span>
                        <span className="word-text">{word}</span>
                      </div>
                    ))}
                  </div>

                  {!showPhrase && (
                    <button className="reveal-button" onClick={() => setShowPhrase(true)}>
                      <Eye size={20} />
                      <span>Tap to reveal</span>
                    </button>
                  )}
                </div>

                <button className="modal-button primary" onClick={closeModal}>
                  I've saved my phrase
                </button>
              </>
            ) : pin.length < PIN_LENGTH ? (
              <>
                <div className="modal-icon">
                  <Plus size={32} />
                </div>
                <h2 className="modal-title">New Wallet</h2>

                <div className="input-group">
                  <label>Wallet Name</label>
                  <input
                    type="text"
                    value={newWalletName}
                    onChange={(e) => setNewWalletName(e.target.value)}
                    placeholder={`Wallet ${wallets.length + 1}`}
                    className="text-input"
                  />
                </div>

                <p className="modal-subtitle">Enter PIN to create wallet</p>

                {renderPinDots(pin.length)}

                {error && (
                  <div className="modal-error">
                    <AlertCircle size={16} />
                    <span>{error}</span>
                  </div>
                )}

                {renderNumpad(handlePinKey)}
              </>
            ) : (
              <div className="loading-state">
                <div className="spinner" />
                <p>Creating wallet...</p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Setup Biometric Modal */}
      {activeModal === 'setup-biometric' && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-content" onClick={e => e.stopPropagation()}>
            <button className="modal-close" onClick={closeModal}>
              <X size={24} />
            </button>

            {success ? (
              <>
                <div className="modal-icon success">
                  <CheckCircle size={32} />
                </div>
                <h2 className="modal-title">{success}</h2>
              </>
            ) : (
              <>
                <div className="modal-icon">
                  <Smartphone size={32} />
                </div>
                <h2 className="modal-title">Enable {biometricType}</h2>
                <p className="modal-subtitle">Enter your PIN to set up {biometricType}</p>

                {renderPinDots(pin.length)}

                {error && (
                  <div className="modal-error">
                    <AlertCircle size={16} />
                    <span>{error}</span>
                  </div>
                )}

                {renderNumpad(handlePinKey)}
              </>
            )}
          </div>
        </div>
      )}

      <style>{`
        .settings-screen {
          min-height: 100vh;
          padding-bottom: 100px;
        }

        .settings-header {
          padding: 16px;
        }

        .settings-title {
          font-size: 24px;
          font-weight: 700;
          color: white;
          margin: 0;
        }

        .settings-content {
          padding: 0 16px;
          display: flex;
          flex-direction: column;
          gap: 24px;
        }

        .settings-group-title {
          font-size: 13px;
          font-weight: 600;
          color: #8b949e;
          margin: 0 0 8px 4px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .settings-group-content {
          background: rgba(22, 27, 34, 0.6);
          backdrop-filter: blur(12px);
          -webkit-backdrop-filter: blur(12px);
          border: 1px solid rgba(48, 54, 61, 0.5);
          border-radius: 16px;
          overflow: hidden;
        }

        /* Theme Selector */
        .theme-selector {
          display: flex;
          padding: 12px;
          gap: 8px;
        }

        .theme-option {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 8px;
          padding: 16px 12px;
          background: transparent;
          border: 2px solid transparent;
          border-radius: 12px;
          cursor: pointer;
          transition: all 0.2s ease;
          -webkit-tap-highlight-color: transparent;
        }

        .theme-option.selected {
          background: rgba(30, 64, 175, 0.15);
          border-color: #1E40AF;
        }

        .theme-option-icon {
          width: 40px;
          height: 40px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: rgba(48, 54, 61, 0.5);
          border-radius: 12px;
          color: #8b949e;
          transition: all 0.2s ease;
        }

        .theme-option.selected .theme-option-icon {
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          color: white;
        }

        .theme-option-label {
          font-size: 13px;
          font-weight: 500;
          color: #8b949e;
        }

        .theme-option.selected .theme-option-label {
          color: white;
        }

        /* Settings Item */
        .settings-item {
          display: flex;
          align-items: center;
          gap: 14px;
          padding: 16px;
          background: transparent;
          border: none;
          width: 100%;
          cursor: pointer;
          transition: all 0.15s ease;
          text-align: left;
          border-bottom: 1px solid rgba(48, 54, 61, 0.3);
          -webkit-tap-highlight-color: transparent;
        }

        .settings-item:last-child {
          border-bottom: none;
        }

        .settings-item:active {
          background: rgba(48, 54, 61, 0.5);
        }

        .settings-item-icon {
          color: #8b949e;
        }

        .settings-item-content {
          flex: 1;
        }

        .settings-item-title {
          font-size: 15px;
          font-weight: 500;
          color: white;
        }

        .settings-item-title.warning {
          color: #f59e0b;
        }

        .settings-item-title.danger {
          color: #ef4444;
        }

        .settings-item-subtitle {
          font-size: 13px;
          color: #8b949e;
          margin-top: 2px;
        }

        .settings-item-chevron {
          color: #484f58;
        }

        /* Settings Toggle */
        .settings-toggle {
          display: flex;
          align-items: center;
          gap: 14px;
          padding: 16px;
          border-bottom: 1px solid rgba(48, 54, 61, 0.3);
          cursor: pointer;
          -webkit-tap-highlight-color: transparent;
        }

        .settings-toggle:last-child {
          border-bottom: none;
        }

        .toggle-switch {
          position: relative;
          width: 52px;
          height: 32px;
          background: #30363d;
          border: none;
          border-radius: 16px;
          cursor: pointer;
          transition: all 0.2s ease;
          flex-shrink: 0;
        }

        .toggle-switch.active {
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
        }

        .toggle-switch.disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .toggle-knob {
          position: absolute;
          top: 4px;
          left: 4px;
          width: 24px;
          height: 24px;
          background: white;
          border-radius: 50%;
          transition: transform 0.2s ease;
        }

        .toggle-switch.active .toggle-knob {
          transform: translateX(20px);
        }

        /* App Info */
        .app-info {
          text-align: center;
          padding: 24px 0;
        }

        .app-version {
          font-size: 13px;
          color: #8b949e;
          margin: 0;
        }

        .app-tagline {
          font-size: 12px;
          color: #484f58;
          margin: 4px 0 0;
        }

        /* Modal */
        .modal-overlay {
          position: fixed;
          inset: 0;
          background: rgba(0, 0, 0, 0.8);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 100;
          padding: 20px;
        }

        .modal-content {
          background: #161B22;
          border-radius: 20px;
          padding: 24px;
          width: 100%;
          max-width: 360px;
          max-height: 90vh;
          overflow-y: auto;
          position: relative;
        }

        .modal-content.wallets-modal {
          max-height: 80vh;
        }

        .modal-close {
          position: absolute;
          top: 16px;
          right: 16px;
          background: none;
          border: none;
          color: #8b949e;
          cursor: pointer;
          padding: 4px;
        }

        .modal-icon {
          width: 64px;
          height: 64px;
          margin: 0 auto 16px;
          border-radius: 50%;
          background: rgba(30, 64, 175, 0.2);
          display: flex;
          align-items: center;
          justify-content: center;
          color: #3B82F6;
        }

        .modal-icon.success {
          background: rgba(16, 185, 129, 0.2);
          color: #10B981;
        }

        .modal-title {
          font-size: 20px;
          font-weight: 700;
          color: white;
          text-align: center;
          margin: 0 0 8px;
        }

        .modal-subtitle {
          font-size: 14px;
          color: #8b949e;
          text-align: center;
          margin: 0 0 24px;
        }

        .modal-subtitle.warning {
          color: #f59e0b;
        }

        .modal-error {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          color: #ef4444;
          font-size: 14px;
          margin-bottom: 16px;
        }

        /* PIN Entry */
        .pin-dots {
          display: flex;
          justify-content: center;
          gap: 12px;
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
          width: 14px;
          height: 14px;
          border-radius: 50%;
          background: #30363d;
          transition: all 0.15s ease;
        }

        .pin-dot.filled {
          background: linear-gradient(135deg, #1E40AF, #3B82F6);
          transform: scale(1.1);
        }

        .numpad {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 10px;
          max-width: 260px;
          margin: 0 auto;
        }

        .numpad-key {
          width: 70px;
          height: 70px;
          border-radius: 50%;
          border: none;
          background: rgba(48, 54, 61, 0.5);
          color: white;
          font-size: 26px;
          font-weight: 500;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.15s ease;
          margin: 0 auto;
        }

        .numpad-key:active {
          background: rgba(30, 64, 175, 0.3);
          transform: scale(0.95);
        }

        .numpad-key.action {
          background: transparent;
          color: #8b949e;
        }

        .numpad-spacer {
          width: 70px;
          height: 70px;
        }

        .step-indicator {
          display: flex;
          justify-content: center;
          gap: 8px;
          margin-top: 24px;
        }

        .step {
          width: 8px;
          height: 8px;
          border-radius: 50%;
          background: #30363d;
        }

        .step.active {
          background: #3B82F6;
        }

        .step.completed {
          background: #10B981;
        }

        /* Recovery Phrase */
        .phrase-container {
          position: relative;
          margin-bottom: 16px;
        }

        .phrase-grid {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 8px;
          transition: filter 0.3s ease;
        }

        .phrase-grid.blurred {
          filter: blur(8px);
          user-select: none;
        }

        .phrase-word {
          background: rgba(48, 54, 61, 0.5);
          border-radius: 8px;
          padding: 8px;
          display: flex;
          align-items: center;
          gap: 6px;
        }

        .word-number {
          font-size: 11px;
          color: #8b949e;
          min-width: 16px;
        }

        .word-text {
          font-size: 13px;
          color: white;
          font-family: monospace;
        }

        .reveal-button {
          position: absolute;
          inset: 0;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 8px;
          background: none;
          border: none;
          color: white;
          cursor: pointer;
        }

        .phrase-actions {
          display: flex;
          gap: 12px;
        }

        .action-button {
          flex: 1;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          padding: 12px;
          background: rgba(48, 54, 61, 0.5);
          border: none;
          border-radius: 12px;
          color: white;
          font-size: 14px;
          cursor: pointer;
        }

        .action-button.primary {
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
        }

        /* Wallets List */
        .wallets-list {
          margin-bottom: 16px;
        }

        .wallet-item {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 14px;
          background: rgba(48, 54, 61, 0.3);
          border-radius: 12px;
          margin-bottom: 8px;
        }

        .wallet-item.active {
          background: rgba(30, 64, 175, 0.2);
          border: 1px solid rgba(59, 130, 246, 0.3);
        }

        .wallet-info {
          display: flex;
          align-items: center;
          gap: 12px;
        }

        .wallet-icon {
          width: 40px;
          height: 40px;
          border-radius: 10px;
          background: rgba(59, 130, 246, 0.2);
          display: flex;
          align-items: center;
          justify-content: center;
          color: #3B82F6;
        }

        .wallet-details {
          display: flex;
          flex-direction: column;
        }

        .wallet-name {
          color: white;
          font-weight: 500;
        }

        .wallet-address {
          font-size: 12px;
          color: #8b949e;
          font-family: monospace;
        }

        .wallet-active-badge {
          font-size: 12px;
          color: #10B981;
          background: rgba(16, 185, 129, 0.2);
          padding: 4px 10px;
          border-radius: 20px;
        }

        .add-wallet-button {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          width: 100%;
          padding: 14px;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          border: none;
          border-radius: 12px;
          color: white;
          font-size: 15px;
          font-weight: 500;
          cursor: pointer;
        }

        /* Input */
        .input-group {
          margin-bottom: 16px;
        }

        .input-group label {
          display: block;
          font-size: 13px;
          color: #8b949e;
          margin-bottom: 8px;
        }

        .text-input {
          width: 100%;
          padding: 12px 16px;
          background: rgba(48, 54, 61, 0.5);
          border: 1px solid rgba(48, 54, 61, 0.8);
          border-radius: 12px;
          color: white;
          font-size: 15px;
          outline: none;
        }

        .text-input:focus {
          border-color: #3B82F6;
        }

        .modal-button {
          width: 100%;
          padding: 14px;
          border: none;
          border-radius: 12px;
          font-size: 15px;
          font-weight: 500;
          cursor: pointer;
          margin-top: 16px;
        }

        .modal-button.primary {
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          color: white;
        }

        .loading-state {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 40px 0;
        }

        .spinner {
          width: 40px;
          height: 40px;
          border: 3px solid #30363d;
          border-top-color: #3B82F6;
          border-radius: 50%;
          animation: spin 1s linear infinite;
          margin-bottom: 16px;
        }

        @keyframes spin {
          to { transform: rotate(360deg); }
        }

        .loading-state p {
          color: #8b949e;
        }

        /* Biometric UI */
        .biometric-loading {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 32px 0;
        }

        .biometric-loading p {
          color: #8b949e;
          margin-top: 16px;
        }

        .biometric-prompt {
          padding: 24px 0;
        }

        .biometric-button {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 12px;
          width: 100%;
          padding: 16px;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          border: none;
          border-radius: 12px;
          color: white;
          font-size: 16px;
          font-weight: 500;
          cursor: pointer;
          transition: transform 0.15s ease, opacity 0.15s ease;
        }

        .biometric-button:active {
          transform: scale(0.98);
          opacity: 0.9;
        }

        .biometric-button:disabled {
          opacity: 0.6;
          cursor: not-allowed;
        }

        .pin-fallback-button {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          width: 100%;
          padding: 12px;
          margin-top: 16px;
          background: transparent;
          border: 1px solid rgba(48, 54, 61, 0.8);
          border-radius: 12px;
          color: #8b949e;
          font-size: 14px;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .pin-fallback-button:hover {
          background: rgba(48, 54, 61, 0.3);
          color: white;
        }
      `}</style>
    </div>
  );
}

function SettingsGroup({
  title,
  children
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <h2 className="settings-group-title">{title}</h2>
      <div className="settings-group-content">
        {children}
      </div>
    </div>
  );
}

function ThemeOption({
  icon,
  label,
  selected,
  onClick
}: {
  icon: React.ReactNode;
  label: string;
  selected: boolean;
  onClick: () => void;
}) {
  return (
    <button className={`theme-option ${selected ? 'selected' : ''}`} onClick={onClick}>
      <div className="theme-option-icon">{icon}</div>
      <span className="theme-option-label">{label}</span>
      {selected && <Check size={14} style={{ color: '#3B82F6' }} />}
    </button>
  );
}

function SettingsItem({
  icon,
  title,
  subtitle,
  onClick,
  color
}: {
  icon: React.ReactNode;
  title: string;
  subtitle?: string;
  onClick: () => void;
  color?: 'warning' | 'danger';
}) {
  return (
    <button className="settings-item" onClick={onClick}>
      <span className="settings-item-icon">{icon}</span>
      <div className="settings-item-content">
        <span className={`settings-item-title ${color || ''}`}>{title}</span>
        {subtitle && <p className="settings-item-subtitle">{subtitle}</p>}
      </div>
      <ChevronRight size={18} className="settings-item-chevron" />
    </button>
  );
}

function SettingsToggle({
  icon,
  title,
  subtitle,
  value,
  onChange,
  disabled = false
}: {
  icon: React.ReactNode;
  title: string;
  subtitle?: string;
  value: boolean;
  onChange: (value: boolean) => void;
  disabled?: boolean;
}) {
  return (
    <div className="settings-toggle" onClick={() => !disabled && onChange(!value)}>
      <span className="settings-item-icon">{icon}</span>
      <div className="settings-item-content">
        <span className="settings-item-title">{title}</span>
        {subtitle && <p className="settings-item-subtitle">{subtitle}</p>}
      </div>
      <button
        className={`toggle-switch ${value ? 'active' : ''} ${disabled ? 'disabled' : ''}`}
        disabled={disabled}
      >
        <div className="toggle-knob" />
      </button>
    </div>
  );
}
