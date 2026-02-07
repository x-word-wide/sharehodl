/**
 * BottomSheet - Modal that slides up from bottom
 * Used for Send, Receive, and other overlay screens
 */

import { useEffect, useState, ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';

interface BottomSheetProps {
  children: ReactNode;
  title?: string;
  onClose?: () => void;
  fullHeight?: boolean;
}

export function BottomSheet({ children, title, onClose, fullHeight = false }: BottomSheetProps) {
  const navigate = useNavigate();
  const [isVisible, setIsVisible] = useState(false);
  const [isClosing, setIsClosing] = useState(false);
  const tg = window.Telegram?.WebApp;

  useEffect(() => {
    // Trigger enter animation
    requestAnimationFrame(() => {
      setIsVisible(true);
    });
  }, []);

  const handleClose = () => {
    tg?.HapticFeedback?.impactOccurred('light');
    setIsClosing(true);
    setTimeout(() => {
      if (onClose) {
        onClose();
      } else {
        navigate(-1);
      }
    }, 250);
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      handleClose();
    }
  };

  return (
    <div
      className={`bottom-sheet-overlay ${isVisible ? 'visible' : ''} ${isClosing ? 'closing' : ''}`}
      onClick={handleBackdropClick}
    >
      <div className={`bottom-sheet ${fullHeight ? 'full-height' : ''}`}>
        {/* Handle bar */}
        <div className="sheet-handle-container">
          <div className="sheet-handle" />
        </div>

        {/* Header */}
        {title && (
          <div className="sheet-header">
            <h2 className="sheet-title">{title}</h2>
            <button className="sheet-close" onClick={handleClose}>
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M18 6L6 18M6 6l12 12" />
              </svg>
            </button>
          </div>
        )}

        {/* Content */}
        <div className="sheet-content">
          {children}
        </div>
      </div>

      <style>{`
        .bottom-sheet-overlay {
          position: fixed;
          inset: 0;
          background: rgba(0, 0, 0, 0);
          z-index: 1000;
          display: flex;
          align-items: flex-end;
          transition: background 0.25s ease;
        }

        .bottom-sheet-overlay.visible {
          background: rgba(0, 0, 0, 0.5);
        }

        .bottom-sheet-overlay.closing {
          background: rgba(0, 0, 0, 0);
        }

        .bottom-sheet {
          width: 100%;
          max-height: 92vh;
          background: var(--tg-theme-bg-color, #0D1117);
          border-radius: 20px 20px 0 0;
          transform: translateY(100%);
          transition: transform 0.3s cubic-bezier(0.32, 0.72, 0, 1);
          overflow: hidden;
          display: flex;
          flex-direction: column;
        }

        .bottom-sheet.full-height {
          max-height: 95vh;
          min-height: 95vh;
        }

        .bottom-sheet-overlay.visible .bottom-sheet {
          transform: translateY(0);
        }

        .bottom-sheet-overlay.closing .bottom-sheet {
          transform: translateY(100%);
        }

        .sheet-handle-container {
          display: flex;
          justify-content: center;
          padding: 12px 0 8px;
        }

        .sheet-handle {
          width: 36px;
          height: 4px;
          background: var(--border-color, rgba(255, 255, 255, 0.2));
          border-radius: 2px;
        }

        .sheet-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 8px 16px 16px;
          border-bottom: 1px solid var(--border-color, rgba(255, 255, 255, 0.1));
        }

        .sheet-title {
          font-size: 18px;
          font-weight: 600;
          color: var(--text-primary, white);
          margin: 0;
        }

        .sheet-close {
          width: 32px;
          height: 32px;
          border-radius: 50%;
          background: var(--surface-bg, rgba(255, 255, 255, 0.1));
          border: none;
          display: flex;
          align-items: center;
          justify-content: center;
          cursor: pointer;
          transition: background 0.2s;
        }

        .sheet-close:active {
          background: var(--input-bg, rgba(255, 255, 255, 0.15));
        }

        .sheet-close svg {
          width: 18px;
          height: 18px;
          color: var(--text-secondary, #8b949e);
        }

        .sheet-content {
          flex: 1;
          overflow-y: auto;
          -webkit-overflow-scrolling: touch;
        }
      `}</style>
    </div>
  );
}
