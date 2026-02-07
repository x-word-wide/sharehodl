/**
 * Welcome Screen - Beautiful onboarding for new users
 * Inspired by Telegram Wallet design patterns
 */

import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const ONBOARDING_SLIDES = [
  {
    iconType: 'chart',
    title: 'The Equity Blockchain',
    description: 'Trade stocks directly on-chain. 24/7 markets, instant settlement, fractional shares'
  },
  {
    iconType: 'star',
    title: 'Stake & Earn Rewards',
    description: 'Stake HODL for up to 4x rewards, lower fees, and governance rights'
  },
  {
    iconType: 'shield',
    title: 'Your Keys, Your Stocks',
    description: 'Non-custodial ownership. Private keys never leave your device'
  },
  {
    iconType: 'globe',
    title: 'Multi-Chain Support',
    description: 'Bridge assets between ShareHODL, Ethereum, Bitcoin, and Cosmos'
  }
];

// SVG Icons for slides
const SlideIcon = ({ type }: { type: string }) => {
  switch (type) {
    case 'chart':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M3 3v18h18" />
          <path d="M18 9l-5 5-4-4-3 3" />
        </svg>
      );
    case 'star':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
        </svg>
      );
    case 'shield':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          <path d="M9 12l2 2 4-4" />
        </svg>
      );
    case 'globe':
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="12" cy="12" r="10" />
          <path d="M2 12h20" />
          <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
        </svg>
      );
    default:
      return null;
  }
};

export function WelcomeScreen() {
  const navigate = useNavigate();
  const tg = window.Telegram?.WebApp;
  const [currentSlide, setCurrentSlide] = useState(0);
  const [isAnimating, setIsAnimating] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [touchStart, setTouchStart] = useState<number | null>(null);

  // Auto-advance slides (pauses on user interaction)
  useEffect(() => {
    if (isPaused) return;

    const timer = setInterval(() => {
      setCurrentSlide((prev) => (prev + 1) % ONBOARDING_SLIDES.length);
    }, 4000);
    return () => clearInterval(timer);
  }, [isPaused]);

  const [touchStartY, setTouchStartY] = useState<number | null>(null);

  // Handle swipe gestures (horizontal only)
  const handleTouchStart = (e: React.TouchEvent) => {
    setTouchStart(e.touches[0].clientX);
    setTouchStartY(e.touches[0].clientY);
    setIsPaused(true); // Pause on touch
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    if (touchStart === null || touchStartY === null) return;

    const touchX = e.touches[0].clientX;
    const touchY = e.touches[0].clientY;
    const diffX = Math.abs(touchStart - touchX);
    const diffY = Math.abs(touchStartY - touchY);

    // If horizontal movement is greater than vertical, prevent default
    // This stops Telegram from interpreting it as a close gesture
    if (diffX > diffY && diffX > 10) {
      e.preventDefault();
    }
  };

  const handleTouchEnd = (e: React.TouchEvent) => {
    if (touchStart === null) return;

    const touchEnd = e.changedTouches[0].clientX;
    const touchEndY = e.changedTouches[0].clientY;
    const diffX = touchStart - touchEnd;
    const diffY = touchStartY !== null ? Math.abs(touchStartY - touchEndY) : 0;

    // Only trigger slide change for horizontal swipes (not vertical)
    if (Math.abs(diffX) > 50 && Math.abs(diffX) > diffY) {
      if (diffX > 0) {
        // Swipe left - next slide
        setCurrentSlide((prev) => (prev + 1) % ONBOARDING_SLIDES.length);
      } else {
        // Swipe right - previous slide
        setCurrentSlide((prev) => (prev - 1 + ONBOARDING_SLIDES.length) % ONBOARDING_SLIDES.length);
      }
      tg?.HapticFeedback?.selectionChanged();
    }

    setTouchStart(null);
    setTouchStartY(null);
  };

  // Resume auto-play after 5 seconds of no interaction
  useEffect(() => {
    if (!isPaused) return;

    const resumeTimer = setTimeout(() => {
      setIsPaused(false);
    }, 5000);

    return () => clearTimeout(resumeTimer);
  }, [isPaused, currentSlide]);

  const handleCreate = () => {
    tg?.HapticFeedback?.impactOccurred('medium');
    setIsAnimating(true);
    setTimeout(() => navigate('/create'), 150);
  };

  const handleImport = () => {
    tg?.HapticFeedback?.impactOccurred('light');
    setIsAnimating(true);
    setTimeout(() => navigate('/import'), 150);
  };

  const slide = ONBOARDING_SLIDES[currentSlide];

  return (
    <div className={`welcome-screen ${isAnimating ? 'fade-out' : ''}`}>
      {/* Animated background */}
      <div className="welcome-bg">
        <div className="welcome-gradient" />
        <div className="welcome-circles">
          <div className="circle circle-1" />
          <div className="circle circle-2" />
          <div className="circle circle-3" />
        </div>
      </div>

      {/* Content */}
      <div className="welcome-content">
        {/* Logo */}
        <div className="welcome-logo">
          <div className="logo-inner">
            <span className="logo-text">SH</span>
          </div>
          <div className="logo-ring" />
        </div>

        <h1 className="welcome-title">ShareHODL</h1>
        <p className="welcome-subtitle">The Blockchain for Stock Trading</p>

        {/* Onboarding slider with swipe support */}
        <div
          className="onboarding-slider"
          onTouchStart={handleTouchStart}
          onTouchMove={handleTouchMove}
          onTouchEnd={handleTouchEnd}
        >
          <div className="slide-content" key={currentSlide}>
            <div className="slide-icon">
              <SlideIcon type={slide.iconType} />
            </div>
            <h3 className="slide-title">{slide.title}</h3>
            <p className="slide-description">{slide.description}</p>
          </div>

          {/* Dots indicator with pause indicator */}
          <div className="slide-dots">
            {ONBOARDING_SLIDES.map((_, idx) => (
              <button
                key={idx}
                className={`dot ${idx === currentSlide ? 'active' : ''} ${isPaused && idx === currentSlide ? 'paused' : ''}`}
                onClick={() => {
                  tg?.HapticFeedback?.selectionChanged();
                  setCurrentSlide(idx);
                  setIsPaused(true);
                }}
              />
            ))}
          </div>

          {/* Swipe hint */}
          <p className="swipe-hint">Swipe to browse</p>
        </div>

        {/* Action buttons */}
        <div className="welcome-actions">
          <button className="btn-create" onClick={handleCreate}>
            <span className="btn-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="10" />
                <path d="M12 8v8M8 12h8" />
              </svg>
            </span>
            Create New Wallet
          </button>

          <button className="btn-import" onClick={handleImport}>
            <span className="btn-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4" />
                <polyline points="7 10 12 15 17 10" />
                <line x1="12" y1="15" x2="12" y2="3" />
              </svg>
            </span>
            Import Existing Wallet
          </button>
        </div>

        {/* Security badge */}
        <div className="security-badge">
          <svg className="lock-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
            <path d="M7 11V7a5 5 0 0110 0v4" />
          </svg>
          <span>End-to-end encrypted</span>
        </div>
      </div>

      <style>{`
        .welcome-screen {
          min-height: 100vh;
          display: flex;
          flex-direction: column;
          position: relative;
          overflow: hidden;
          transition: opacity 0.15s ease-out;
        }

        .welcome-screen.fade-out {
          opacity: 0;
        }

        .welcome-bg {
          position: absolute;
          inset: 0;
          z-index: 0;
        }

        .welcome-gradient {
          position: absolute;
          inset: 0;
          background: radial-gradient(ellipse at top, rgba(30, 64, 175, 0.15) 0%, transparent 60%),
                      radial-gradient(ellipse at bottom right, rgba(59, 130, 246, 0.1) 0%, transparent 50%);
        }

        .welcome-circles {
          position: absolute;
          inset: 0;
          overflow: hidden;
        }

        .circle {
          position: absolute;
          border-radius: 50%;
          border: 1px solid rgba(30, 64, 175, 0.1);
        }

        .circle-1 {
          width: 300px;
          height: 300px;
          top: -100px;
          right: -100px;
          animation: float 8s ease-in-out infinite;
        }

        .circle-2 {
          width: 200px;
          height: 200px;
          bottom: 20%;
          left: -80px;
          animation: float 6s ease-in-out infinite reverse;
        }

        .circle-3 {
          width: 150px;
          height: 150px;
          bottom: 10%;
          right: -50px;
          animation: float 10s ease-in-out infinite;
        }

        @keyframes float {
          0%, 100% { transform: translateY(0) rotate(0deg); }
          50% { transform: translateY(-20px) rotate(5deg); }
        }

        .welcome-content {
          position: relative;
          z-index: 1;
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          padding: 60px 24px 40px;
        }

        .welcome-logo {
          position: relative;
          width: 100px;
          height: 100px;
          margin-bottom: 20px;
        }

        .logo-inner {
          width: 100%;
          height: 100%;
          border-radius: 28px;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 50%, #60A5FA 100%);
          display: flex;
          align-items: center;
          justify-content: center;
          box-shadow: 0 20px 40px rgba(30, 64, 175, 0.3);
          animation: pulse-glow 3s ease-in-out infinite;
        }

        .logo-text {
          font-size: 36px;
          font-weight: 800;
          color: white;
          letter-spacing: -1px;
        }

        .logo-ring {
          position: absolute;
          inset: -8px;
          border-radius: 36px;
          border: 2px solid rgba(30, 64, 175, 0.3);
          animation: ring-pulse 3s ease-in-out infinite;
        }

        @keyframes pulse-glow {
          0%, 100% { box-shadow: 0 20px 40px rgba(30, 64, 175, 0.3); }
          50% { box-shadow: 0 25px 50px rgba(30, 64, 175, 0.4); }
        }

        @keyframes ring-pulse {
          0%, 100% { transform: scale(1); opacity: 0.5; }
          50% { transform: scale(1.05); opacity: 0.8; }
        }

        .welcome-title {
          font-size: 32px;
          font-weight: 700;
          color: white;
          margin: 0 0 8px;
          letter-spacing: -0.5px;
        }

        .welcome-subtitle {
          font-size: 15px;
          color: #8b949e;
          margin: 0 0 40px;
        }

        .onboarding-slider {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          width: 100%;
          max-width: 320px;
        }

        .slide-content {
          text-align: center;
          animation: slideIn 0.4s ease-out;
        }

        @keyframes slideIn {
          from {
            opacity: 0;
            transform: translateX(20px);
          }
          to {
            opacity: 1;
            transform: translateX(0);
          }
        }

        .slide-icon {
          width: 56px;
          height: 56px;
          margin: 0 auto 16px;
          padding: 12px;
          background: linear-gradient(135deg, rgba(30, 64, 175, 0.2) 0%, rgba(59, 130, 246, 0.2) 100%);
          border-radius: 16px;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .slide-icon svg {
          width: 32px;
          height: 32px;
          color: #3B82F6;
        }

        .slide-title {
          font-size: 20px;
          font-weight: 600;
          color: white;
          margin: 0 0 12px;
        }

        .slide-description {
          font-size: 14px;
          color: #8b949e;
          line-height: 1.5;
          margin: 0;
        }

        .slide-dots {
          display: flex;
          gap: 8px;
          margin-top: 24px;
        }

        .dot {
          width: 8px;
          height: 8px;
          border-radius: 4px;
          background: #30363d;
          border: none;
          padding: 0;
          cursor: pointer;
          transition: all 0.3s ease;
        }

        .dot.active {
          width: 24px;
          background: linear-gradient(90deg, #1E40AF, #3B82F6);
        }

        .dot.paused {
          animation: pulse-dot 1s ease-in-out infinite;
        }

        @keyframes pulse-dot {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }

        .swipe-hint {
          margin-top: 16px;
          font-size: 12px;
          color: #6b7280;
          text-align: center;
        }

        .welcome-actions {
          width: 100%;
          max-width: 320px;
          display: flex;
          flex-direction: column;
          gap: 12px;
          margin-top: auto;
        }

        .btn-create {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 10px;
          width: 100%;
          padding: 16px 24px;
          font-size: 16px;
          font-weight: 600;
          color: white;
          background: linear-gradient(135deg, #1E40AF 0%, #3B82F6 100%);
          border: none;
          border-radius: 14px;
          cursor: pointer;
          transition: all 0.2s ease;
          box-shadow: 0 4px 20px rgba(30, 64, 175, 0.3);
        }

        .btn-create:active {
          transform: scale(0.97);
        }

        .btn-import {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 10px;
          width: 100%;
          padding: 16px 24px;
          font-size: 16px;
          font-weight: 600;
          color: #8b949e;
          background: rgba(48, 54, 61, 0.5);
          border: 1px solid #30363d;
          border-radius: 14px;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .btn-import:active {
          transform: scale(0.97);
          background: rgba(48, 54, 61, 0.8);
        }

        .btn-icon {
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .btn-icon svg {
          width: 20px;
          height: 20px;
        }

        .security-badge {
          display: flex;
          align-items: center;
          gap: 8px;
          margin-top: 24px;
          padding: 10px 16px;
          background: rgba(16, 185, 129, 0.1);
          border-radius: 20px;
          color: #10b981;
          font-size: 13px;
        }

        .lock-icon {
          width: 14px;
          height: 14px;
        }
      `}</style>
    </div>
  );
}
