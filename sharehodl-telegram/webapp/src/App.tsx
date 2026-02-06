/**
 * ShareHODL Telegram Mini App
 *
 * Main application component with routing
 */

import { useEffect, useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom';
import { useWalletStore } from './services/walletStore';
import { logger } from './utils/logger';

// Theme storage key
const THEME_KEY = 'sh_theme';
type Theme = 'dark' | 'light' | 'system';

// Global theme hook - applies theme to document
function useGlobalTheme() {
  useEffect(() => {
    const applyTheme = () => {
      const saved = localStorage.getItem(THEME_KEY) as Theme | null;
      const theme = saved || 'system';
      const root = document.documentElement;
      const tg = window.Telegram?.WebApp;

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
    };

    // Apply on mount
    applyTheme();

    // Listen for localStorage changes (when settings change theme)
    const handleStorage = (e: StorageEvent) => {
      if (e.key === THEME_KEY) {
        applyTheme();
      }
    };
    window.addEventListener('storage', handleStorage);

    // Also listen for custom theme change event
    const handleThemeChange = () => applyTheme();
    window.addEventListener('themechange', handleThemeChange);

    // Listen for system theme changes
    const mediaQuery = window.matchMedia?.('(prefers-color-scheme: dark)');
    const handleMediaChange = () => {
      const saved = localStorage.getItem(THEME_KEY);
      if (!saved || saved === 'system') {
        applyTheme();
      }
    };
    mediaQuery?.addEventListener('change', handleMediaChange);

    return () => {
      window.removeEventListener('storage', handleStorage);
      window.removeEventListener('themechange', handleThemeChange);
      mediaQuery?.removeEventListener('change', handleMediaChange);
    };
  }, []);
}

// Screens
import { WelcomeScreen } from './screens/WelcomeScreen';
import { CreateWalletScreen } from './screens/CreateWalletScreen';
import { ImportWalletScreen } from './screens/ImportWalletScreen';
import { UnlockScreen } from './screens/UnlockScreen';
import { PortfolioScreen } from './screens/PortfolioScreen';
import { MarketScreen } from './screens/MarketScreen';
import { TradeScreen } from './screens/TradeScreen';
import { SendScreen } from './screens/SendScreen';
import { ReceiveScreen } from './screens/ReceiveScreen';
import { P2PScreen } from './screens/P2PScreen';
import { LendingScreen } from './screens/LendingScreen';
import { InheritanceScreen } from './screens/InheritanceScreen';
import { BridgeScreen } from './screens/BridgeScreen';
import { SettingsScreen } from './screens/SettingsScreen';
import { StakingScreen } from './screens/StakingScreen';
import { GovernanceScreen } from './screens/GovernanceScreen';
import { AssetDetailScreen } from './screens/AssetDetailScreen';
import { EquityDetailScreen } from './screens/EquityDetailScreen';
import { EquityProfileScreen } from './screens/EquityProfileScreen';

// Components
import { BottomNav } from './components/BottomNav';
import { LoadingScreen } from './components/LoadingScreen';

logger.debug('App.tsx loaded');

function AppContent() {
  logger.debug('AppContent rendering...');
  const navigate = useNavigate();
  const location = useLocation();
  const { isInitialized, isLocked, initialize } = useWalletStore();
  const [loading, setLoading] = useState(true);
  const [initError, setInitError] = useState<string | null>(null);

  // Apply global theme
  useGlobalTheme();

  // Initialize wallet on mount
  useEffect(() => {
    const init = async () => {
      try {
        logger.debug('Initializing wallet...');
        await initialize();
        logger.debug('Wallet initialized');
        setLoading(false);
      } catch (error) {
        logger.error('Initialization error:', error);
        setInitError(error instanceof Error ? error.message : 'Failed to initialize');
        setLoading(false);
      }
    };
    init();
  }, [initialize]);

  // Show error state
  if (initError) {
    return (
      <div style={{
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '20px',
        backgroundColor: '#0D1117',
        color: 'white',
        textAlign: 'center'
      }}>
        <h1 style={{ marginBottom: '16px' }}>Initialization Error</h1>
        <p style={{ color: '#8b949e', marginBottom: '20px' }}>{initError}</p>
        <button
          onClick={() => {
            localStorage.clear();
            window.location.reload();
          }}
          style={{
            padding: '12px 24px',
            background: 'linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%)',
            border: 'none',
            borderRadius: '12px',
            color: 'white',
            fontSize: '16px',
            cursor: 'pointer'
          }}
        >
          Reset & Retry
        </button>
      </div>
    );
  }

  // Handle Telegram deep links
  useEffect(() => {
    const tg = window.Telegram?.WebApp;
    const startParam = tg?.initDataUnsafe?.start_param;

    if (startParam) {
      // Parse screen from start_param (e.g., "screen=portfolio")
      const params = new URLSearchParams(startParam.replace(/_/g, '&'));
      const screen = params.get('screen');
      if (screen && !isLocked) {
        navigate(`/${screen}`);
      }
    }
  }, [isLocked, navigate]);

  // Configure Telegram back button
  useEffect(() => {
    const tg = window.Telegram?.WebApp;
    if (!tg) return;

    const mainRoutes = ['/', '/portfolio', '/market', '/trade', '/settings'];
    const isMainRoute = mainRoutes.includes(location.pathname);

    // Store handler reference for proper cleanup
    const handleBack = () => navigate(-1);

    if (isMainRoute) {
      tg.BackButton.hide();
    } else {
      tg.BackButton.show();
      tg.BackButton.onClick(handleBack);
    }

    return () => {
      tg.BackButton.offClick(handleBack);
    };
  }, [location.pathname, navigate]);

  // Disable Telegram's vertical swipe to close gesture
  useEffect(() => {
    const tg = window.Telegram?.WebApp;
    if (tg) {
      // Disable vertical swipes to prevent app from closing on swipe
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (tg as any).isVerticalSwipesEnabled = false;
    }
  }, []);

  if (loading) {
    return <LoadingScreen />;
  }

  // Not initialized - show welcome/create/import
  if (!isInitialized) {
    return (
      <Routes>
        <Route path="/" element={<WelcomeScreen />} />
        <Route path="/create" element={<CreateWalletScreen />} />
        <Route path="/import" element={<ImportWalletScreen />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    );
  }

  // IMPORTANT: Allow /create and /import routes to complete even after wallet is initialized
  // This is needed because createWallet sets isInitialized=true but the user still needs
  // to see and verify their seed phrase before being redirected
  if (location.pathname === '/create' || location.pathname === '/import') {
    return (
      <Routes>
        <Route path="/create" element={<CreateWalletScreen />} />
        <Route path="/import" element={<ImportWalletScreen />} />
      </Routes>
    );
  }

  // Locked - show unlock
  if (isLocked) {
    return <UnlockScreen />;
  }

  // Unlocked - show main app
  const showBottomNav = ['/', '/portfolio', '/market', '/trade', '/settings'].includes(location.pathname);

  return (
    <div className="flex flex-col min-h-screen">
      <main className="flex-1 pb-20">
        <Routes>
          <Route path="/" element={<Navigate to="/portfolio" replace />} />
          <Route path="/portfolio" element={<PortfolioScreen />} />
          <Route path="/market" element={<MarketScreen />} />
          <Route path="/trade" element={<TradeScreen />} />
          <Route path="/send" element={<SendScreen />} />
          <Route path="/send/:chain" element={<SendScreen />} />
          <Route path="/receive" element={<ReceiveScreen />} />
          <Route path="/receive/:chain" element={<ReceiveScreen />} />
          <Route path="/p2p" element={<P2PScreen />} />
          <Route path="/lending" element={<LendingScreen />} />
          <Route path="/inheritance" element={<InheritanceScreen />} />
          <Route path="/bridge" element={<BridgeScreen />} />
          <Route path="/staking" element={<StakingScreen />} />
          <Route path="/governance" element={<GovernanceScreen />} />
          <Route path="/asset/:tokenId" element={<AssetDetailScreen />} />
          <Route path="/equity/:equityId" element={<EquityDetailScreen />} />
          <Route path="/equity-profile/:equityId" element={<EquityProfileScreen />} />
          <Route path="/settings" element={<SettingsScreen />} />
          <Route path="*" element={<Navigate to="/portfolio" replace />} />
        </Routes>
      </main>

      {showBottomNav && <BottomNav />}
    </div>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AppContent />
    </BrowserRouter>
  );
}
