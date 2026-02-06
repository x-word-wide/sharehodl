/**
 * Settings Screen - Wallet settings and preferences
 */

import { useState, useEffect } from 'react';
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
  Check
} from 'lucide-react';
import { useWalletStore } from '../services/walletStore';

// Theme storage key
const THEME_KEY = 'sh_theme';
const BIOMETRIC_ENABLED_KEY = 'sh_biometric_enabled';

type Theme = 'dark' | 'light' | 'system';

export function SettingsScreen() {
  const { lockWallet, resetWallet } = useWalletStore();
  const tg = window.Telegram?.WebApp;

  // Helper functions for alerts/confirms with fallbacks
  const showAlert = (message: string) => {
    if (tg?.showAlert) {
      tg.showAlert(message);
    } else {
      alert(message);
    }
  };

  const showConfirm = (message: string, callback: (confirmed: boolean) => void) => {
    if (tg?.showConfirm) {
      tg.showConfirm(message, callback);
    } else {
      const confirmed = confirm(message);
      callback(confirmed);
    }
  };

  const openLink = (url: string) => {
    if (tg?.openLink) {
      tg.openLink(url);
    } else {
      window.open(url, '_blank');
    }
  };

  // Theme state
  const [theme, setTheme] = useState<Theme>(() => {
    const saved = localStorage.getItem(THEME_KEY);
    return (saved as Theme) || 'dark';
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
    const isDark = theme === 'dark' || (theme === 'system' && tg?.colorScheme === 'dark');

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
  }, [theme, tg?.colorScheme]);

  const handleThemeChange = (newTheme: Theme) => {
    tg?.HapticFeedback?.selectionChanged();
    setTheme(newTheme);
  };

  const handleBiometricToggle = async () => {
    tg?.HapticFeedback?.impactOccurred('light');
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const biometricManager = (tg as any)?.BiometricManager;

    if (!biometricEnabled) {
      // Enable biometric
      if (biometricManager && biometricManager.requestAccess) {
        biometricManager.requestAccess({ reason: 'Enable quick unlock with biometrics' }, (granted: boolean) => {
          if (granted) {
            setBiometricEnabled(true);
            localStorage.setItem(BIOMETRIC_ENABLED_KEY, 'true');
            tg?.HapticFeedback?.notificationOccurred('success');
            showAlert(`${biometricType} enabled for quick unlock`);
          } else {
            showAlert('Biometric access denied');
          }
        });
      } else {
        // Fallback for testing
        setBiometricEnabled(true);
        localStorage.setItem(BIOMETRIC_ENABLED_KEY, 'true');
        tg?.HapticFeedback?.notificationOccurred('success');
        showAlert('Biometric login enabled');
      }
    } else {
      // Disable biometric
      setBiometricEnabled(false);
      localStorage.setItem(BIOMETRIC_ENABLED_KEY, 'false');
      tg?.HapticFeedback?.selectionChanged();
      showAlert('Biometric login disabled');
    }
  };

  const handleChangePin = () => {
    tg?.HapticFeedback?.impactOccurred('light');
    showAlert('Change PIN - Coming soon!');
  };

  const handleViewRecoveryPhrase = () => {
    tg?.HapticFeedback?.impactOccurred('light');
    showAlert('Enter PIN to view recovery phrase');
  };

  const handle2FA = () => {
    tg?.HapticFeedback?.impactOccurred('light');
    showAlert('Two-Factor Authentication - Coming soon!');
  };

  const handleNetworkSelect = () => {
    tg?.HapticFeedback?.impactOccurred('light');
    showAlert('Network selection - Coming soon!');
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

  return (
    <div className="settings-screen">
      {/* Header */}
      <div className="settings-header">
        <h1 className="settings-title">Settings</h1>
      </div>

      {/* Settings Groups */}
      <div className="settings-content">
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
            onClick={handleChangePin}
          />
          <SettingsToggle
            icon={<Smartphone size={20} />}
            title={biometricType}
            subtitle="Quick unlock with biometrics"
            value={biometricEnabled}
            onChange={handleBiometricToggle}
          />
          <SettingsItem
            icon={<Key size={20} />}
            title="View Recovery Phrase"
            onClick={handleViewRecoveryPhrase}
          />
          <SettingsItem
            icon={<Shield size={20} />}
            title="Two-Factor Authentication"
            onClick={handle2FA}
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
            onClick={handleNetworkSelect}
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

        .theme-option:active {
          transform: scale(0.95);
        }

        .theme-option:hover:not(.selected) {
          background: rgba(48, 54, 61, 0.3);
          border-color: rgba(48, 54, 61, 0.5);
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

        .settings-item:hover {
          background: rgba(48, 54, 61, 0.2);
        }

        .settings-item:active {
          background: rgba(48, 54, 61, 0.5);
          transform: scale(0.98);
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
          transition: background 0.15s ease;
          -webkit-tap-highlight-color: transparent;
        }

        .settings-toggle:hover {
          background: rgba(48, 54, 61, 0.2);
        }

        .settings-toggle:active {
          background: rgba(48, 54, 61, 0.4);
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
          -webkit-tap-highlight-color: transparent;
        }

        .toggle-switch:active {
          transform: scale(0.95);
        }

        .toggle-switch.active {
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          box-shadow: 0 0 12px rgba(59, 130, 246, 0.4);
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
          box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
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
      {selected && <Check size={14} className="theme-option-check" style={{ color: '#3B82F6' }} />}
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
    <div className="settings-toggle">
      <span className="settings-item-icon">{icon}</span>
      <div className="settings-item-content">
        <span className="settings-item-title">{title}</span>
        {subtitle && <p className="settings-item-subtitle">{subtitle}</p>}
      </div>
      <button
        className={`toggle-switch ${value ? 'active' : ''} ${disabled ? 'disabled' : ''}`}
        onClick={() => !disabled && onChange(!value)}
        disabled={disabled}
      >
        <div className="toggle-knob" />
      </button>
    </div>
  );
}
