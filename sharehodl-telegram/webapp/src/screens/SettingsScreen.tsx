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
  CheckCircle,
  MoreVertical,
  Pencil,
  Trash2
} from 'lucide-react';
import { useWalletStore } from '../services/walletStore';

// Theme storage key
const THEME_KEY = 'sh_theme';
const BIOMETRIC_ENABLED_KEY = 'sh_biometric_enabled';

type Theme = 'dark' | 'light' | 'system';
type ModalType = 'none' | 'view-phrase' | 'change-pin' | 'wallets' | 'add-wallet' | 'edit-wallet' | 'setup-biometric';
type AddWalletMode = 'choose' | 'create' | 'import';

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
    importNewWallet,
    renameWallet,
    deleteWallet,
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
  const [addWalletMode, setAddWalletMode] = useState<AddWalletMode>('choose');
  const [importMnemonic, setImportMnemonic] = useState('');
  const [usePinFallback, setUsePinFallback] = useState(false);
  const [biometricAttempted, setBiometricAttempted] = useState(false);
  // Edit wallet state
  const [editingWallet, setEditingWallet] = useState<{ id: string; name: string } | null>(null);
  const [editWalletName, setEditWalletName] = useState('');
  const [deleteConfirmStep, setDeleteConfirmStep] = useState(0);
  // Biometric for add wallet flow
  const [addWalletBiometricAttempted, setAddWalletBiometricAttempted] = useState(false);
  const [addWalletUsePinFallback, setAddWalletUsePinFallback] = useState(false);

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
      // If BiometricManager exists, assume biometric is available
      // We'll check properly when user tries to enable it
      setBiometricAvailable(true);

      // Try to get more info, but don't block on it
      try {
        biometricManager.init(() => {
          // Update with actual values if init completes
          if (biometricManager.isBiometricAvailable !== undefined) {
            setBiometricAvailable(biometricManager.isAccessGranted || biometricManager.isBiometricAvailable);
          }
          if (biometricManager.biometricType) {
            const type = biometricManager.biometricType;
            setBiometricType(type === 'face' ? 'Face ID' : type === 'finger' ? 'Touch ID' : 'Biometric');
          }
        });
      } catch {
        // init failed, but we still allow trying biometric
      }
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
    // Dispatch custom event to notify App.tsx
    window.dispatchEvent(new Event('themechange'));
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
    setAddWalletMode('choose');
    setImportMnemonic('');
    setUsePinFallback(false);
    setBiometricAttempted(false);
    // Edit wallet state
    setEditingWallet(null);
    setEditWalletName('');
    setDeleteConfirmStep(0);
    // Add wallet biometric state
    setAddWalletBiometricAttempted(false);
    setAddWalletUsePinFallback(false);
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
        if (addWalletMode === 'create') {
          await handleAddWalletPinSubmit(updatedPin);
        } else if (addWalletMode === 'import') {
          await handleImportWalletPinSubmit(updatedPin);
        }
      } else if (activeModal === 'setup-biometric') {
        await handleBiometricSetupPin(updatedPin);
      } else if (activeModal === 'edit-wallet' && deleteConfirmStep === 2) {
        await handleDeleteWalletPinSubmit(updatedPin);
      }
    }
  }, [pin, newPin, confirmPin, pinStep, isLoading, activeModal, addWalletMode, deleteConfirmStep, tg]);

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
  // SECURITY: Uses Telegram's secure biometric storage - token IS the PIN
  const handleBiometricAuth = useCallback(() => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const biometricManager = (tg as any)?.BiometricManager;
    if (!biometricManager) {
      setUsePinFallback(true);
      return;
    }

    setBiometricAttempted(true);
    setIsLoading(true);

    // Timeout to prevent infinite loading
    const timeout = setTimeout(() => {
      setIsLoading(false);
      setError(`${biometricType} timed out. Please use PIN.`);
      setUsePinFallback(true);
    }, 30000);

    biometricManager.authenticate(
      { reason: 'Verify your identity to view recovery phrase' },
      async (success: boolean, token?: string) => {
        clearTimeout(timeout);

        if (success && token && token.length > 0) {
          try {
            // SECURITY: The token from Telegram's secure storage IS the PIN
            // This was stored when user enabled biometrics via updateBiometricToken(pin)
            const phrase = await getRecoveryPhrase(token);
            setRecoveryPhrase(phrase.split(' '));
            tg?.HapticFeedback?.notificationOccurred('success');
          } catch {
            tg?.HapticFeedback?.notificationOccurred('error');
            setError('Invalid credentials. Please use PIN.');
            setUsePinFallback(true);
          }
        } else if (success && (!token || token.length === 0)) {
          // Biometric verified but no token - biometrics not properly configured
          tg?.HapticFeedback?.notificationOccurred('error');
          setError(`${biometricType} not configured. Please re-enable in Settings or use PIN.`);
          setUsePinFallback(true);
        } else {
          // User cancelled or biometric failed
          tg?.HapticFeedback?.notificationOccurred('error');
          setError(`${biometricType} cancelled`);
          setUsePinFallback(true);
        }
        setIsLoading(false);
      }
    );
  }, [tg, getRecoveryPhrase, biometricType]);

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

  // Biometric authentication for add/import wallet
  const handleAddWalletBiometricAuth = useCallback(() => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const biometricManager = (tg as any)?.BiometricManager;
    if (!biometricManager) {
      setAddWalletUsePinFallback(true);
      return;
    }

    setAddWalletBiometricAttempted(true);
    setIsLoading(true);

    // Timeout to prevent infinite loading
    const timeout = setTimeout(() => {
      setIsLoading(false);
      setError(`${biometricType} timed out. Please use PIN.`);
      setAddWalletUsePinFallback(true);
    }, 30000);

    biometricManager.authenticate(
      { reason: 'Verify your identity to add wallet' },
      async (success: boolean, token?: string) => {
        clearTimeout(timeout);

        if (success && token && token.length > 0) {
          try {
            // Token from Telegram's secure storage IS the PIN
            const isValid = await verifyPin(token);
            if (!isValid) {
              throw new Error('Invalid credentials');
            }

            // PIN verified, proceed with add/import
            if (addWalletMode === 'create') {
              await handleAddWalletPinSubmit(token);
            } else if (addWalletMode === 'import') {
              await handleImportWalletPinSubmit(token);
            }
            tg?.HapticFeedback?.notificationOccurred('success');
          } catch {
            tg?.HapticFeedback?.notificationOccurred('error');
            setError('Invalid credentials. Please use PIN.');
            setAddWalletUsePinFallback(true);
          }
        } else if (success && (!token || token.length === 0)) {
          tg?.HapticFeedback?.notificationOccurred('error');
          setError(`${biometricType} not configured. Please use PIN.`);
          setAddWalletUsePinFallback(true);
        } else {
          tg?.HapticFeedback?.notificationOccurred('error');
          setError(`${biometricType} cancelled`);
          setAddWalletUsePinFallback(true);
        }
        setIsLoading(false);
      }
    );
  }, [tg, verifyPin, biometricType, addWalletMode]);

  // Trigger biometric for add wallet when reaching PIN step
  useEffect(() => {
    if (activeModal === 'add-wallet' && biometricEnabled && !addWalletBiometricAttempted && !addWalletUsePinFallback) {
      // For create mode, trigger when user has entered wallet name (moved past choose)
      // For import mode, trigger when mnemonic is valid (12+ words)
      const shouldTrigger =
        (addWalletMode === 'create' && !walletMnemonic && !success) ||
        (addWalletMode === 'import' && importMnemonic.trim().split(/\s+/).length >= 12 && !success);

      if (shouldTrigger && pin.length === 0 && !isLoading) {
        const timer = setTimeout(() => {
          handleAddWalletBiometricAuth();
        }, 300);
        return () => clearTimeout(timer);
      }
    }
  }, [activeModal, biometricEnabled, addWalletBiometricAttempted, addWalletUsePinFallback, addWalletMode, walletMnemonic, importMnemonic, pin.length, isLoading, success, handleAddWalletBiometricAuth]);

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

  // Import existing wallet
  const handleImportWalletPinSubmit = async (enteredPin: string) => {
    setIsLoading(true);
    try {
      await importNewWallet(newWalletName || `Imported Wallet ${wallets.length + 1}`, importMnemonic, enteredPin);
      tg?.HapticFeedback?.notificationOccurred('success');
      setSuccess('Wallet imported successfully!');
      setTimeout(() => closeModal(), 1500);
    } catch (err) {
      tg?.HapticFeedback?.notificationOccurred('error');
      setShake(true);
      setTimeout(() => {
        setShake(false);
        setPin('');
      }, 500);
      setError(err instanceof Error ? err.message : 'Failed to import wallet');
    }
    setIsLoading(false);
  };

  // Handle wallet rename
  const handleRenameWallet = () => {
    if (!editingWallet || !editWalletName.trim()) return;

    try {
      renameWallet(editingWallet.id, editWalletName.trim());
      tg?.HapticFeedback?.notificationOccurred('success');
      setSuccess('Wallet renamed successfully!');
      setTimeout(() => {
        setActiveModal('wallets');
        resetModalState();
      }, 1000);
    } catch (err) {
      tg?.HapticFeedback?.notificationOccurred('error');
      setError(err instanceof Error ? err.message : 'Failed to rename wallet');
    }
  };

  // Handle wallet delete with PIN verification
  const handleDeleteWalletPinSubmit = async (enteredPin: string) => {
    if (!editingWallet) return;

    setIsLoading(true);
    try {
      await deleteWallet(editingWallet.id, enteredPin);
      tg?.HapticFeedback?.notificationOccurred('success');
      setSuccess('Wallet deleted successfully!');
      setTimeout(() => {
        setActiveModal('wallets');
        resetModalState();
      }, 1500);
    } catch (err) {
      tg?.HapticFeedback?.notificationOccurred('error');
      setShake(true);
      setTimeout(() => {
        setShake(false);
        setPin('');
      }, 500);
      setError(err instanceof Error ? err.message : 'Failed to delete wallet');
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
  // SECURITY: PIN is stored ONLY in Telegram's secure biometric storage
  // Never store PIN in localStorage as it can be accessed by XSS attacks
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

      if (biometricManager && biometricManager.updateBiometricToken) {
        // Store PIN ONLY in Telegram's secure biometric storage
        // This is protected by the device's secure enclave
        biometricManager.updateBiometricToken(enteredPin, (tokenSaved: boolean) => {
          if (tokenSaved) {
            setBiometricEnabled(true);
            localStorage.setItem(BIOMETRIC_ENABLED_KEY, 'true');
            // Remove any old insecure storage
            localStorage.removeItem('sh_bio_pin');
            tg?.HapticFeedback?.notificationOccurred('success');
            setSuccess(`${biometricType} enabled successfully!`);
            setTimeout(() => closeModal(), 1500);
          } else {
            tg?.HapticFeedback?.notificationOccurred('error');
            setError(`Failed to set up ${biometricType}. Your device may not support secure storage.`);
            setPin('');
          }
          setIsLoading(false);
        });
      } else {
        // No biometric manager available - cannot securely store PIN
        tg?.HapticFeedback?.notificationOccurred('error');
        setError(`${biometricType} is not available on this device.`);
        setPin('');
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
                  <div className="wallet-actions">
                    {wallet.id === activeWalletId && (
                      <div className="wallet-active-badge">Active</div>
                    )}
                    <button
                      className="wallet-edit-btn"
                      onClick={(e) => {
                        e.stopPropagation();
                        tg?.HapticFeedback?.impactOccurred('light');
                        setEditingWallet({ id: wallet.id, name: wallet.name });
                        setEditWalletName(wallet.name);
                        setActiveModal('edit-wallet');
                      }}
                    >
                      <MoreVertical size={18} />
                    </button>
                  </div>
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

            {/* Success state for created wallet */}
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
            ) : success ? (
              /* Success state for imported wallet */
              <>
                <div className="modal-icon success">
                  <CheckCircle size={32} />
                </div>
                <h2 className="modal-title">{success}</h2>
              </>
            ) : addWalletMode === 'choose' ? (
              /* Choose between Create and Import */
              <>
                <div className="modal-icon">
                  <Wallet size={32} />
                </div>
                <h2 className="modal-title">Add Wallet</h2>
                <p className="modal-subtitle">Choose how you want to add a wallet</p>

                <div className="wallet-options">
                  <button
                    className="wallet-option"
                    onClick={() => {
                      tg?.HapticFeedback?.impactOccurred('light');
                      setAddWalletMode('create');
                    }}
                  >
                    <div className="wallet-option-icon create">
                      <Plus size={24} />
                    </div>
                    <div className="wallet-option-content">
                      <span className="wallet-option-title">Create New Wallet</span>
                      <span className="wallet-option-desc">Generate a new recovery phrase</span>
                    </div>
                    <ChevronRight size={20} className="wallet-option-chevron" />
                  </button>

                  <button
                    className="wallet-option"
                    onClick={() => {
                      tg?.HapticFeedback?.impactOccurred('light');
                      setAddWalletMode('import');
                    }}
                  >
                    <div className="wallet-option-icon import">
                      <Key size={24} />
                    </div>
                    <div className="wallet-option-content">
                      <span className="wallet-option-title">Import Existing Wallet</span>
                      <span className="wallet-option-desc">Use your recovery phrase</span>
                    </div>
                    <ChevronRight size={20} className="wallet-option-chevron" />
                  </button>
                </div>
              </>
            ) : addWalletMode === 'create' ? (
              /* Create new wallet flow */
              isLoading ? (
                <div className="loading-state">
                  <div className="spinner" />
                  <p>Creating wallet...</p>
                </div>
              ) : biometricEnabled && !addWalletUsePinFallback ? (
                /* Biometric option for create */
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

                  <p className="modal-subtitle">Verify your identity to create wallet</p>

                  <div className="biometric-prompt">
                    <button
                      className="biometric-button"
                      onClick={handleAddWalletBiometricAuth}
                      disabled={isLoading}
                    >
                      <Smartphone size={24} />
                      <span>Create with {biometricType}</span>
                    </button>
                  </div>

                  {error && (
                    <div className="modal-error">
                      <AlertCircle size={16} />
                      <span>{error}</span>
                    </div>
                  )}

                  <button
                    className="pin-fallback-button"
                    onClick={() => {
                      setAddWalletUsePinFallback(true);
                      setError('');
                    }}
                  >
                    <Lock size={16} />
                    <span>Use PIN instead</span>
                  </button>

                  <button
                    className="back-button"
                    onClick={() => {
                      setAddWalletMode('choose');
                      setAddWalletUsePinFallback(false);
                      setAddWalletBiometricAttempted(false);
                      setError('');
                    }}
                  >
                    ← Back
                  </button>
                </>
              ) : pin.length < PIN_LENGTH ? (
                /* PIN entry for create */
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

                  {biometricEnabled && (
                    <button
                      className="pin-fallback-button"
                      onClick={() => {
                        setAddWalletUsePinFallback(false);
                        setAddWalletBiometricAttempted(false);
                        setPin('');
                        setError('');
                      }}
                    >
                      <Smartphone size={16} />
                      <span>Use {biometricType} instead</span>
                    </button>
                  )}

                  <button
                    className="back-button"
                    onClick={() => {
                      setAddWalletMode('choose');
                      setPin('');
                      setError('');
                      setAddWalletUsePinFallback(false);
                      setAddWalletBiometricAttempted(false);
                    }}
                  >
                    ← Back
                  </button>
                </>
              ) : (
                <div className="loading-state">
                  <div className="spinner" />
                  <p>Creating wallet...</p>
                </div>
              )
            ) : addWalletMode === 'import' ? (
              /* Import wallet flow */
              !importMnemonic.trim() || importMnemonic.trim().split(/\s+/).length < 12 ? (
                <>
                  <div className="modal-icon">
                    <Key size={32} />
                  </div>
                  <h2 className="modal-title">Import Wallet</h2>
                  <p className="modal-subtitle">Enter your 12 or 24 word recovery phrase</p>

                  <div className="input-group">
                    <label>Wallet Name</label>
                    <input
                      type="text"
                      value={newWalletName}
                      onChange={(e) => setNewWalletName(e.target.value)}
                      placeholder={`Imported Wallet ${wallets.length + 1}`}
                      className="text-input"
                    />
                  </div>

                  <div className="input-group">
                    <label>Recovery Phrase</label>
                    <textarea
                      value={importMnemonic}
                      onChange={(e) => setImportMnemonic(e.target.value.toLowerCase())}
                      placeholder="Enter your recovery phrase, separated by spaces..."
                      className="text-input mnemonic-input"
                      rows={4}
                      autoCapitalize="none"
                      autoCorrect="off"
                      spellCheck={false}
                    />
                  </div>

                  {error && (
                    <div className="modal-error">
                      <AlertCircle size={16} />
                      <span>{error}</span>
                    </div>
                  )}

                  <button
                    className="modal-button primary"
                    onClick={() => {
                      const words = importMnemonic.trim().split(/\s+/);
                      if (words.length !== 12 && words.length !== 24) {
                        setError('Recovery phrase must be 12 or 24 words');
                        return;
                      }
                      setError('');
                      // Phrase is valid length, will proceed to PIN entry
                    }}
                    disabled={importMnemonic.trim().split(/\s+/).length < 12}
                  >
                    Continue
                  </button>

                  <button
                    className="back-button"
                    onClick={() => {
                      setAddWalletMode('choose');
                      setImportMnemonic('');
                      setError('');
                    }}
                  >
                    ← Back
                  </button>
                </>
              ) : isLoading ? (
                <div className="loading-state">
                  <div className="spinner" />
                  <p>Importing wallet...</p>
                </div>
              ) : biometricEnabled && !addWalletUsePinFallback ? (
                /* Biometric option for import */
                <>
                  <div className="modal-icon">
                    <Key size={32} />
                  </div>
                  <h2 className="modal-title">Import Wallet</h2>
                  <p className="modal-subtitle">Verify your identity to import wallet</p>

                  <div className="biometric-prompt">
                    <button
                      className="biometric-button"
                      onClick={handleAddWalletBiometricAuth}
                      disabled={isLoading}
                    >
                      <Smartphone size={24} />
                      <span>Import with {biometricType}</span>
                    </button>
                  </div>

                  {error && (
                    <div className="modal-error">
                      <AlertCircle size={16} />
                      <span>{error}</span>
                    </div>
                  )}

                  <button
                    className="pin-fallback-button"
                    onClick={() => {
                      setAddWalletUsePinFallback(true);
                      setError('');
                    }}
                  >
                    <Lock size={16} />
                    <span>Use PIN instead</span>
                  </button>

                  <button
                    className="back-button"
                    onClick={() => {
                      setImportMnemonic('');
                      setAddWalletUsePinFallback(false);
                      setAddWalletBiometricAttempted(false);
                      setError('');
                    }}
                  >
                    ← Back
                  </button>
                </>
              ) : pin.length < PIN_LENGTH ? (
                /* PIN entry for import */
                <>
                  <div className="modal-icon">
                    <Lock size={32} />
                  </div>
                  <h2 className="modal-title">Enter PIN</h2>
                  <p className="modal-subtitle">Verify your identity to import wallet</p>

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
                        setAddWalletUsePinFallback(false);
                        setAddWalletBiometricAttempted(false);
                        setPin('');
                        setError('');
                      }}
                    >
                      <Smartphone size={16} />
                      <span>Use {biometricType} instead</span>
                    </button>
                  )}

                  <button
                    className="back-button"
                    onClick={() => {
                      setPin('');
                      setError('');
                      setImportMnemonic('');
                      setAddWalletUsePinFallback(false);
                      setAddWalletBiometricAttempted(false);
                    }}
                  >
                    ← Back
                  </button>
                </>
              ) : (
                <div className="loading-state">
                  <div className="spinner" />
                  <p>Importing wallet...</p>
                </div>
              )
            ) : null}
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

      {/* Edit Wallet Modal */}
      {activeModal === 'edit-wallet' && editingWallet && (
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
            ) : deleteConfirmStep === 0 ? (
              /* Main edit options */
              <>
                <div className="modal-icon">
                  <Wallet size={32} />
                </div>
                <h2 className="modal-title">Edit Wallet</h2>
                <p className="modal-subtitle">{editingWallet.name}</p>

                <div className="edit-wallet-options">
                  <div className="input-group">
                    <label>Wallet Name</label>
                    <input
                      type="text"
                      value={editWalletName}
                      onChange={(e) => setEditWalletName(e.target.value)}
                      placeholder="Enter wallet name"
                      className="text-input"
                    />
                  </div>

                  <button
                    className="modal-button primary"
                    onClick={handleRenameWallet}
                    disabled={!editWalletName.trim() || editWalletName.trim() === editingWallet.name}
                  >
                    <Pencil size={18} />
                    <span>Save Name</span>
                  </button>

                  {wallets.length > 1 && (
                    <button
                      className="modal-button danger"
                      onClick={() => {
                        tg?.HapticFeedback?.impactOccurred('medium');
                        setDeleteConfirmStep(1);
                      }}
                    >
                      <Trash2 size={18} />
                      <span>Delete Wallet</span>
                    </button>
                  )}

                  {wallets.length <= 1 && (
                    <p className="delete-warning">Cannot delete the only wallet</p>
                  )}
                </div>

                <button
                  className="back-button"
                  onClick={() => {
                    setActiveModal('wallets');
                    resetModalState();
                  }}
                >
                  ← Back to Wallets
                </button>
              </>
            ) : deleteConfirmStep === 1 ? (
              /* Delete confirmation */
              <>
                <div className="modal-icon danger">
                  <AlertCircle size={32} />
                </div>
                <h2 className="modal-title">Delete Wallet?</h2>
                <p className="modal-subtitle warning">
                  This will permanently remove "{editingWallet.name}" from this device.
                </p>
                <p className="modal-subtitle">
                  Make sure you have backed up the recovery phrase!
                </p>

                {error && (
                  <div className="modal-error">
                    <AlertCircle size={16} />
                    <span>{error}</span>
                  </div>
                )}

                <div className="delete-confirm-buttons">
                  <button
                    className="modal-button secondary"
                    onClick={() => setDeleteConfirmStep(0)}
                  >
                    Cancel
                  </button>
                  <button
                    className="modal-button danger"
                    onClick={() => {
                      tg?.HapticFeedback?.impactOccurred('heavy');
                      setDeleteConfirmStep(2);
                    }}
                  >
                    Continue
                  </button>
                </div>
              </>
            ) : deleteConfirmStep === 2 ? (
              /* PIN entry for delete */
              <>
                <div className="modal-icon danger">
                  <Lock size={32} />
                </div>
                <h2 className="modal-title">Enter PIN to Delete</h2>
                <p className="modal-subtitle">Verify your identity to delete this wallet</p>

                {renderPinDots(pin.length)}

                {error && (
                  <div className="modal-error">
                    <AlertCircle size={16} />
                    <span>{error}</span>
                  </div>
                )}

                {renderNumpad(handlePinKey)}

                <button
                  className="back-button"
                  onClick={() => {
                    setDeleteConfirmStep(1);
                    setPin('');
                    setError('');
                  }}
                >
                  ← Back
                </button>
              </>
            ) : null}
          </div>
        </div>
      )}

      <style>{`
        .settings-screen {
          min-height: 100vh;
          padding-bottom: 100px;
          background-color: var(--tg-theme-bg-color);
        }

        .settings-header {
          padding: 16px;
        }

        .settings-title {
          font-size: 24px;
          font-weight: 700;
          color: var(--text-primary);
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
          color: var(--text-secondary);
          margin: 0 0 8px 4px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .settings-group-content {
          background: var(--surface-bg);
          backdrop-filter: blur(12px);
          -webkit-backdrop-filter: blur(12px);
          border: 1px solid var(--border-color);
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
          background: var(--input-bg);
          border-radius: 12px;
          color: var(--text-secondary);
          transition: all 0.2s ease;
        }

        .theme-option.selected .theme-option-icon {
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          color: white;
        }

        .theme-option-label {
          font-size: 13px;
          font-weight: 500;
          color: var(--text-secondary);
        }

        .theme-option.selected .theme-option-label {
          color: var(--text-primary);
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
          border-bottom: 1px solid var(--border-color);
          -webkit-tap-highlight-color: transparent;
        }

        .settings-item:last-child {
          border-bottom: none;
        }

        .settings-item:active {
          background: var(--input-bg);
        }

        .settings-item-icon {
          color: var(--text-secondary);
        }

        .settings-item-content {
          flex: 1;
        }

        .settings-item-title {
          font-size: 15px;
          font-weight: 500;
          color: var(--text-primary);
        }

        .settings-item-title.warning {
          color: #f59e0b;
        }

        .settings-item-title.danger {
          color: #ef4444;
        }

        .settings-item-subtitle {
          font-size: 13px;
          color: var(--text-secondary);
          margin-top: 2px;
        }

        .settings-item-chevron {
          color: var(--text-muted);
        }

        /* Settings Toggle */
        .settings-toggle {
          display: flex;
          align-items: center;
          gap: 14px;
          padding: 16px;
          border-bottom: 1px solid var(--border-color);
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
          background: var(--toggle-bg);
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
          color: var(--text-secondary);
          margin: 0;
        }

        .app-tagline {
          font-size: 12px;
          color: var(--text-muted);
          margin: 4px 0 0;
        }

        /* Modal */
        .modal-overlay {
          position: fixed;
          inset: 0;
          background: var(--overlay-bg);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 100;
          padding: 20px;
        }

        .modal-content {
          background: var(--modal-bg);
          border-radius: 20px;
          padding: 24px;
          width: 100%;
          max-width: 360px;
          max-height: 90vh;
          overflow-y: auto;
          position: relative;
          border: 1px solid var(--border-color);
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
          color: var(--text-secondary);
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
          color: var(--text-primary);
          text-align: center;
          margin: 0 0 8px;
        }

        .modal-subtitle {
          font-size: 14px;
          color: var(--text-secondary);
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
          background: var(--pin-dot-bg);
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
          background: var(--numpad-bg);
          color: var(--text-primary);
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
          color: var(--text-secondary);
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
          background: var(--pin-dot-bg);
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
          background: var(--input-bg);
          border-radius: 8px;
          padding: 8px;
          display: flex;
          align-items: center;
          gap: 6px;
        }

        .word-number {
          font-size: 11px;
          color: var(--text-secondary);
          min-width: 16px;
        }

        .word-text {
          font-size: 13px;
          color: var(--text-primary);
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
          color: var(--text-primary);
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
          background: var(--input-bg);
          border: none;
          border-radius: 12px;
          color: var(--text-primary);
          font-size: 14px;
          cursor: pointer;
        }

        .action-button.primary {
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          color: white;
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
          background: var(--input-bg);
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
          color: var(--text-primary);
          font-weight: 500;
        }

        .wallet-address {
          font-size: 12px;
          color: var(--text-secondary);
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

        /* Wallet Options (Create/Import choice) */
        .wallet-options {
          display: flex;
          flex-direction: column;
          gap: 12px;
          margin-top: 8px;
        }

        .wallet-option {
          display: flex;
          align-items: center;
          gap: 14px;
          padding: 16px;
          background: var(--input-bg);
          border: 1px solid var(--border-color);
          border-radius: 14px;
          cursor: pointer;
          transition: all 0.2s ease;
          text-align: left;
          width: 100%;
        }

        .wallet-option:active {
          transform: scale(0.98);
          background: var(--surface-bg);
        }

        .wallet-option-icon {
          width: 48px;
          height: 48px;
          border-radius: 12px;
          display: flex;
          align-items: center;
          justify-content: center;
          flex-shrink: 0;
        }

        .wallet-option-icon.create {
          background: linear-gradient(135deg, rgba(16, 185, 129, 0.2) 0%, rgba(16, 185, 129, 0.1) 100%);
          color: #10B981;
        }

        .wallet-option-icon.import {
          background: linear-gradient(135deg, rgba(59, 130, 246, 0.2) 0%, rgba(59, 130, 246, 0.1) 100%);
          color: #3B82F6;
        }

        .wallet-option-content {
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .wallet-option-title {
          font-size: 15px;
          font-weight: 600;
          color: var(--text-primary);
        }

        .wallet-option-desc {
          font-size: 13px;
          color: var(--text-secondary);
        }

        .wallet-option-chevron {
          color: var(--text-muted);
        }

        /* Mnemonic Input */
        .mnemonic-input {
          resize: none;
          font-family: monospace;
          line-height: 1.5;
        }

        /* Back Button */
        .back-button {
          display: block;
          width: 100%;
          margin-top: 16px;
          padding: 12px;
          background: transparent;
          border: none;
          color: var(--text-secondary);
          font-size: 14px;
          cursor: pointer;
          transition: color 0.2s ease;
        }

        .back-button:hover {
          color: var(--text-primary);
        }

        /* Input */
        .input-group {
          margin-bottom: 16px;
        }

        .input-group label {
          display: block;
          font-size: 13px;
          color: var(--text-secondary);
          margin-bottom: 8px;
        }

        .text-input {
          width: 100%;
          padding: 12px 16px;
          background: var(--input-bg);
          border: 1px solid var(--border-color);
          border-radius: 12px;
          color: var(--text-primary);
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
          border: 3px solid var(--pin-dot-bg);
          border-top-color: #3B82F6;
          border-radius: 50%;
          animation: spin 1s linear infinite;
          margin-bottom: 16px;
        }

        @keyframes spin {
          to { transform: rotate(360deg); }
        }

        .loading-state p {
          color: var(--text-secondary);
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
          color: var(--text-secondary);
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
          border: 1px solid var(--border-color);
          border-radius: 12px;
          color: var(--text-secondary);
          font-size: 14px;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .pin-fallback-button:hover {
          background: var(--input-bg);
          color: var(--text-primary);
        }

        /* Wallet Actions */
        .wallet-actions {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .wallet-edit-btn {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 32px;
          height: 32px;
          background: transparent;
          border: none;
          border-radius: 8px;
          color: var(--text-secondary);
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .wallet-edit-btn:active {
          background: var(--input-bg);
          color: var(--text-primary);
        }

        /* Edit Wallet Modal */
        .edit-wallet-options {
          display: flex;
          flex-direction: column;
          gap: 12px;
          margin-top: 8px;
        }

        .modal-button {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          width: 100%;
          padding: 14px;
          border: none;
          border-radius: 12px;
          font-size: 15px;
          font-weight: 500;
          cursor: pointer;
        }

        .modal-button.primary {
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          color: white;
        }

        .modal-button.primary:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .modal-button.secondary {
          background: var(--input-bg);
          color: var(--text-primary);
          border: 1px solid var(--border-color);
        }

        .modal-button.danger {
          background: rgba(239, 68, 68, 0.15);
          color: #ef4444;
          border: 1px solid rgba(239, 68, 68, 0.3);
        }

        .modal-button.danger:hover {
          background: rgba(239, 68, 68, 0.25);
        }

        .modal-icon.danger {
          background: rgba(239, 68, 68, 0.2);
          color: #ef4444;
        }

        .delete-warning {
          text-align: center;
          font-size: 13px;
          color: var(--text-secondary);
          margin-top: 8px;
        }

        .delete-confirm-buttons {
          display: flex;
          gap: 12px;
          margin-top: 16px;
        }

        .delete-confirm-buttons .modal-button {
          flex: 1;
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
