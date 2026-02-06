/**
 * Bottom Navigation - Professional Telegram-style navigation
 * SECURITY: Uses React SVG components instead of dangerouslySetInnerHTML
 */

import { useNavigate, useLocation } from 'react-router-dom';
import { Wallet, TrendingUp, ArrowUpDown, Settings } from 'lucide-react';
import { ReactNode } from 'react';

interface NavItem {
  path: string;
  label: string;
  icon: ReactNode;
  activeIcon: ReactNode;
}

// SECURITY: Use proper React components instead of raw HTML strings
const navItems: NavItem[] = [
  {
    path: '/portfolio',
    label: 'Wallet',
    icon: <Wallet size={24} strokeWidth={2} />,
    activeIcon: <Wallet size={24} strokeWidth={2.5} />
  },
  {
    path: '/market',
    label: 'Market',
    icon: <TrendingUp size={24} strokeWidth={2} />,
    activeIcon: <TrendingUp size={24} strokeWidth={2.5} />
  },
  {
    path: '/trade',
    label: 'Trade',
    icon: <ArrowUpDown size={24} strokeWidth={2} />,
    activeIcon: <ArrowUpDown size={24} strokeWidth={2.5} />
  },
  {
    path: '/settings',
    label: 'Settings',
    icon: <Settings size={24} strokeWidth={2} />,
    activeIcon: <Settings size={24} strokeWidth={2.5} />
  }
];

export function BottomNav() {
  const navigate = useNavigate();
  const location = useLocation();
  const tg = window.Telegram?.WebApp;

  const handleNavClick = (path: string) => {
    tg?.HapticFeedback?.selectionChanged();
    navigate(path);
  };

  return (
    <nav className="bottom-nav">
      <div className="nav-glass-bg" />
      <div className="nav-container">
        {navItems.map(({ path, label, icon, activeIcon }) => {
          const isActive = location.pathname === path;

          return (
            <button
              key={path}
              onClick={() => handleNavClick(path)}
              className={`nav-item ${isActive ? 'active' : ''}`}
            >
              <span className="nav-icon-wrapper">
                {isActive && <span className="nav-icon-glow" />}
                <span className="nav-icon">
                  {isActive ? activeIcon : icon}
                </span>
              </span>
              <span className="nav-label">{label}</span>
            </button>
          );
        })}
      </div>

      <style>{`
        .bottom-nav {
          position: fixed;
          bottom: 0;
          left: 0;
          right: 0;
          z-index: 50;
          padding-bottom: env(safe-area-inset-bottom, 0);
        }

        .nav-glass-bg {
          position: absolute;
          inset: 0;
          background: var(--glass-bg);
          backdrop-filter: blur(24px);
          -webkit-backdrop-filter: blur(24px);
          border-top: 1px solid var(--glass-border);
        }

        .nav-container {
          position: relative;
          display: flex;
          justify-content: space-around;
          align-items: center;
          padding: 10px 0 8px;
        }

        .nav-item {
          position: relative;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          flex: 1;
          padding: 6px 0;
          border: none;
          background: transparent;
          cursor: pointer;
          transition: all 0.2s ease;
          -webkit-tap-highlight-color: transparent;
        }

        .nav-item:active {
          transform: scale(0.95);
        }

        .nav-icon-wrapper {
          position: relative;
          width: 48px;
          height: 32px;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .nav-icon-glow {
          position: absolute;
          inset: 0;
          background: radial-gradient(
            ellipse at center,
            rgba(59, 130, 246, 0.35) 0%,
            transparent 70%
          );
          border-radius: 12px;
          animation: glowPulse 2s ease-in-out infinite;
        }

        @keyframes glowPulse {
          0%, 100% { opacity: 0.8; }
          50% { opacity: 1; }
        }

        .nav-icon {
          position: relative;
          width: 24px;
          height: 24px;
          display: flex;
          align-items: center;
          justify-content: center;
          color: var(--text-secondary);
          transition: all 0.25s ease;
        }

        .nav-icon svg {
          width: 24px;
          height: 24px;
        }

        .nav-item.active .nav-icon {
          color: #3B82F6;
          transform: translateY(-2px);
        }

        .nav-label {
          font-size: 11px;
          font-weight: 500;
          color: var(--text-muted);
          margin-top: 4px;
          transition: all 0.25s ease;
        }

        .nav-item.active .nav-label {
          color: #3B82F6;
          font-weight: 700;
        }
      `}</style>
    </nav>
  );
}
