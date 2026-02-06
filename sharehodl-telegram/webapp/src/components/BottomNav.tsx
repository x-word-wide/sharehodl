/**
 * Bottom Navigation - Professional Telegram-style navigation
 */

import { useNavigate, useLocation } from 'react-router-dom';

interface NavItem {
  path: string;
  label: string;
  icon: string;
  activeIcon: string;
}

const navItems: NavItem[] = [
  {
    path: '/portfolio',
    label: 'Wallet',
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="M16 12h.01"/></svg>`,
    activeIcon: `<svg viewBox="0 0 24 24" fill="currentColor"><rect x="2" y="4" width="20" height="16" rx="2"/><circle cx="16" cy="12" r="1.5" fill="#0D1117"/></svg>`
  },
  {
    path: '/market',
    label: 'Market',
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 3v18h18"/><path d="M18 9l-5 5-4-4-3 3"/></svg>`,
    activeIcon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M3 3v18h18"/><path d="M18 9l-5 5-4-4-3 3"/></svg>`
  },
  {
    path: '/trade',
    label: 'Trade',
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M7 16V4M7 4L3 8M7 4l4 4"/><path d="M17 8v12m0 0l4-4m-4 4l-4-4"/></svg>`,
    activeIcon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M7 16V4M7 4L3 8M7 4l4 4"/><path d="M17 8v12m0 0l4-4m-4 4l-4-4"/></svg>`
  },
  {
    path: '/settings',
    label: 'Settings',
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z"/></svg>`,
    activeIcon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06a1.65 1.65 0 00.33-1.82 1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06a1.65 1.65 0 001.82.33H9a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06a1.65 1.65 0 00-.33 1.82V9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z"/></svg>`
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
                <span
                  className="nav-icon"
                  dangerouslySetInnerHTML={{ __html: isActive ? activeIcon : icon }}
                />
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
          background: rgba(13, 17, 23, 0.85);
          backdrop-filter: blur(24px);
          -webkit-backdrop-filter: blur(24px);
          border-top: 1px solid rgba(48, 54, 61, 0.5);
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
          color: #8b949e;
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
          color: #6b7280;
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
