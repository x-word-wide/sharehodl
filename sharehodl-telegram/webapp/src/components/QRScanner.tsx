/**
 * QR Scanner Component - Premium Design
 * Supports both camera scanning (Telegram native) and gallery image selection
 */

import { useRef, useState, useCallback } from 'react';
import jsQR from 'jsqr';

// SECURITY: Maximum allowed QR data length
// ShareHODL addresses are ~43-50 chars, but allow for URIs like "hodl:address?amount=X"
const MAX_QR_DATA_LENGTH = 200;

// SECURITY: Allowed characters for address data (bech32 charset + URI chars)
const VALID_QR_CHARS = /^[a-zA-Z0-9:?=&.]+$/;

/**
 * Sanitize and validate QR code data
 * SECURITY: Prevents injection attacks and DoS via large payloads
 */
function sanitizeQRData(data: string): { valid: boolean; sanitized: string; error?: string } {
  // Check for empty data
  if (!data || typeof data !== 'string') {
    return { valid: false, sanitized: '', error: 'Empty QR code data' };
  }

  // Trim whitespace
  const trimmed = data.trim();

  // SECURITY: Check length to prevent DoS
  if (trimmed.length > MAX_QR_DATA_LENGTH) {
    return { valid: false, sanitized: '', error: 'QR code data too long' };
  }

  // SECURITY: Check for control characters and invalid content
  if (trimmed.includes('<') || trimmed.includes('>') || trimmed.includes('javascript:')) {
    return { valid: false, sanitized: '', error: 'Invalid QR code content' };
  }

  // SECURITY: Validate characters (allow bech32 charset + URI components)
  if (!VALID_QR_CHARS.test(trimmed)) {
    return { valid: false, sanitized: '', error: 'QR code contains invalid characters' };
  }

  // Extract address if it's a payment URI (hodl:address or sharehodl:address)
  let address = trimmed;
  if (trimmed.includes(':')) {
    const parts = trimmed.split(':');
    if (parts[0].toLowerCase() === 'hodl' || parts[0].toLowerCase() === 'sharehodl') {
      // Extract address from URI, stripping query params
      address = parts[1].split('?')[0];
    }
  }

  return { valid: true, sanitized: address };
}

interface QRScannerProps {
  onScan: (data: string) => void;
  onClose: () => void;
}

export function QRScanner({ onScan, onClose }: QRScannerProps) {
  const tg = window.Telegram?.WebApp;
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Use Telegram's native QR scanner
  const handleCameraScan = useCallback(() => {
    if (tg?.showScanQrPopup) {
      tg.showScanQrPopup(
        { text: 'Point camera at QR code' },
        (result: string) => {
          if (result) {
            // SECURITY: Sanitize QR data before processing
            const { valid, sanitized, error: sanitizeError } = sanitizeQRData(result);
            if (!valid) {
              tg.HapticFeedback?.notificationOccurred('error');
              setError(sanitizeError || 'Invalid QR code');
              return true; // Close scanner and show error
            }
            tg.HapticFeedback?.notificationOccurred('success');
            onScan(sanitized);
            return true; // Close the scanner
          }
          return false; // Keep scanner open
        }
      );
    } else {
      setError('Camera scanner not available');
    }
  }, [tg, onScan]);

  // Handle gallery image selection
  const handleGallerySelect = useCallback(() => {
    fileInputRef.current?.click();
  }, []);

  // Process selected image file
  const handleFileChange = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setIsProcessing(true);
    setError(null);

    try {
      const img = new Image();
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');

      await new Promise<void>((resolve, reject) => {
        img.onload = () => {
          const maxSize = 1000;
          let { width, height } = img;
          if (width > maxSize || height > maxSize) {
            if (width > height) {
              height = (height / width) * maxSize;
              width = maxSize;
            } else {
              width = (width / height) * maxSize;
              height = maxSize;
            }
          }

          canvas.width = width;
          canvas.height = height;
          ctx?.drawImage(img, 0, 0, width, height);

          const imageData = ctx?.getImageData(0, 0, width, height);
          if (imageData) {
            const code = jsQR(imageData.data, imageData.width, imageData.height);
            if (code) {
              // SECURITY: Sanitize QR data before processing
              const { valid, sanitized, error: sanitizeError } = sanitizeQRData(code.data);
              if (!valid) {
                reject(new Error(sanitizeError || 'Invalid QR code'));
                return;
              }
              tg?.HapticFeedback?.notificationOccurred('success');
              onScan(sanitized);
              resolve();
            } else {
              reject(new Error('No QR code found in image'));
            }
          } else {
            reject(new Error('Failed to process image'));
          }
        };

        img.onerror = () => reject(new Error('Failed to load image'));
        img.src = URL.createObjectURL(file);
      });
    } catch (err) {
      tg?.HapticFeedback?.notificationOccurred('error');
      setError(err instanceof Error ? err.message : 'Failed to scan QR code');
    } finally {
      setIsProcessing(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  }, [tg, onScan]);

  return (
    <div className="qr-modal">
      <div className="qr-backdrop" onClick={onClose} />
      <div className="qr-sheet">
        {/* Drag handle */}
        <div className="qr-handle" />

        {/* Header */}
        <div className="qr-header">
          <div className="qr-header-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <rect x="3" y="3" width="7" height="7" rx="1" />
              <rect x="14" y="3" width="7" height="7" rx="1" />
              <rect x="3" y="14" width="7" height="7" rx="1" />
              <rect x="14" y="14" width="3" height="3" />
              <rect x="18" y="14" width="3" height="3" />
              <rect x="14" y="18" width="3" height="3" />
              <rect x="18" y="18" width="3" height="3" />
            </svg>
          </div>
          <h2>Scan QR Code</h2>
          <p>Choose how to scan the recipient's address</p>
        </div>

        {/* Options */}
        <div className="qr-options">
          <button
            className="qr-btn qr-btn-primary"
            onClick={handleCameraScan}
            disabled={isProcessing}
          >
            <div className="qr-btn-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M23 19a2 2 0 01-2 2H3a2 2 0 01-2-2V8a2 2 0 012-2h4l2-3h6l2 3h4a2 2 0 012 2z" />
                <circle cx="12" cy="13" r="4" />
              </svg>
            </div>
            <div className="qr-btn-text">
              <span className="qr-btn-title">Open Camera</span>
              <span className="qr-btn-desc">Scan QR code directly</span>
            </div>
            <div className="qr-btn-arrow">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M9 18l6-6-6-6" />
              </svg>
            </div>
          </button>

          <button
            className="qr-btn qr-btn-secondary"
            onClick={handleGallerySelect}
            disabled={isProcessing}
          >
            <div className="qr-btn-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                <circle cx="8.5" cy="8.5" r="1.5" />
                <path d="M21 15l-5-5L5 21" />
              </svg>
            </div>
            <div className="qr-btn-text">
              <span className="qr-btn-title">Choose from Photos</span>
              <span className="qr-btn-desc">Select image with QR code</span>
            </div>
            <div className="qr-btn-arrow">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M9 18l6-6-6-6" />
              </svg>
            </div>
          </button>
        </div>

        {/* Processing state */}
        {isProcessing && (
          <div className="qr-status qr-processing">
            <div className="qr-spinner" />
            <span>Processing image...</span>
          </div>
        )}

        {/* Error state */}
        {error && (
          <div className="qr-status qr-error">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 8v4M12 16h.01" />
            </svg>
            <span>{error}</span>
          </div>
        )}

        {/* Cancel button */}
        <button className="qr-cancel" onClick={onClose}>
          Cancel
        </button>

        {/* Hidden file input */}
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          onChange={handleFileChange}
          style={{ display: 'none' }}
        />
      </div>

      <style>{`
        .qr-modal {
          position: fixed;
          inset: 0;
          z-index: 9999;
          display: flex;
          align-items: flex-end;
        }

        .qr-backdrop {
          position: absolute;
          inset: 0;
          background: rgba(0, 0, 0, 0.85);
          backdrop-filter: blur(8px);
          -webkit-backdrop-filter: blur(8px);
        }

        .qr-sheet {
          position: relative;
          width: 100%;
          background: linear-gradient(180deg, #1a1f2e 0%, #0f1219 100%);
          border-radius: 28px 28px 0 0;
          padding: 12px 20px 32px;
          padding-bottom: max(32px, env(safe-area-inset-bottom));
          animation: qrSlideUp 0.35s cubic-bezier(0.32, 0.72, 0, 1);
        }

        @keyframes qrSlideUp {
          from { transform: translateY(100%); }
          to { transform: translateY(0); }
        }

        .qr-handle {
          width: 40px;
          height: 4px;
          background: #3d4654;
          border-radius: 2px;
          margin: 0 auto 24px;
        }

        .qr-header {
          text-align: center;
          margin-bottom: 28px;
        }

        .qr-header-icon {
          width: 64px;
          height: 64px;
          background: rgba(59, 130, 246, 0.12);
          border-radius: 20px;
          display: flex;
          align-items: center;
          justify-content: center;
          margin: 0 auto 16px;
        }

        .qr-header-icon svg {
          width: 32px;
          height: 32px;
          color: #3b82f6;
        }

        .qr-header h2 {
          font-size: 22px;
          font-weight: 700;
          color: white;
          margin: 0 0 8px;
        }

        .qr-header p {
          font-size: 15px;
          color: #6b7689;
          margin: 0;
        }

        .qr-options {
          display: flex;
          flex-direction: column;
          gap: 12px;
          margin-bottom: 20px;
        }

        .qr-btn {
          display: flex;
          align-items: center;
          gap: 16px;
          padding: 18px 16px;
          border-radius: 16px;
          border: none;
          cursor: pointer;
          text-align: left;
          transition: all 0.2s;
        }

        .qr-btn:active {
          transform: scale(0.98);
        }

        .qr-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .qr-btn-primary {
          background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(37, 99, 235, 0.1) 100%);
          border: 1px solid rgba(59, 130, 246, 0.3);
        }

        .qr-btn-primary:active {
          background: linear-gradient(135deg, rgba(59, 130, 246, 0.25) 0%, rgba(37, 99, 235, 0.2) 100%);
        }

        .qr-btn-secondary {
          background: #0f1318;
          border: 1px solid #2d3748;
        }

        .qr-btn-secondary:active {
          background: #1a1f2a;
        }

        .qr-btn-icon {
          width: 52px;
          height: 52px;
          border-radius: 14px;
          display: flex;
          align-items: center;
          justify-content: center;
          flex-shrink: 0;
        }

        .qr-btn-primary .qr-btn-icon {
          background: rgba(59, 130, 246, 0.2);
        }

        .qr-btn-primary .qr-btn-icon svg {
          width: 26px;
          height: 26px;
          color: #3b82f6;
        }

        .qr-btn-secondary .qr-btn-icon {
          background: rgba(16, 185, 129, 0.15);
        }

        .qr-btn-secondary .qr-btn-icon svg {
          width: 26px;
          height: 26px;
          color: #10b981;
        }

        .qr-btn-text {
          flex: 1;
        }

        .qr-btn-title {
          display: block;
          font-size: 16px;
          font-weight: 600;
          color: white;
          margin-bottom: 3px;
        }

        .qr-btn-desc {
          display: block;
          font-size: 13px;
          color: #6b7689;
        }

        .qr-btn-arrow {
          width: 24px;
          height: 24px;
          color: #4b5563;
        }

        .qr-btn-arrow svg {
          width: 100%;
          height: 100%;
        }

        .qr-status {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 12px;
          padding: 16px;
          border-radius: 14px;
          margin-bottom: 16px;
          font-size: 14px;
        }

        .qr-processing {
          background: rgba(59, 130, 246, 0.1);
          color: #8b95a8;
        }

        .qr-error {
          background: rgba(239, 68, 68, 0.1);
          color: #f87171;
        }

        .qr-error svg {
          width: 18px;
          height: 18px;
          flex-shrink: 0;
        }

        .qr-spinner {
          width: 20px;
          height: 20px;
          border: 2px solid #2d3748;
          border-top-color: #3b82f6;
          border-radius: 50%;
          animation: qrSpin 1s linear infinite;
        }

        @keyframes qrSpin {
          to { transform: rotate(360deg); }
        }

        .qr-cancel {
          width: 100%;
          padding: 16px;
          border: none;
          background: transparent;
          color: #6b7689;
          font-size: 16px;
          font-weight: 600;
          cursor: pointer;
        }

        .qr-cancel:active {
          color: white;
        }
      `}</style>
    </div>
  );
}
