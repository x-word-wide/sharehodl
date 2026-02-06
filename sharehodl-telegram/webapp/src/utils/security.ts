/**
 * Security Module - Brute Force Protection & PIN Security
 *
 * SECURITY FEATURES:
 * - Persistent brute force protection with exponential backoff
 * - Account lockout after max failed attempts
 * - PIN complexity validation (blocks weak PINs)
 * - Auto-lock timeout management
 * - Secure session management
 */

// Storage keys for security data
const SECURITY_KEYS = {
  FAILED_ATTEMPTS: 'sh_security_failed_attempts',
  LOCKOUT_UNTIL: 'sh_security_lockout_until',
  LAST_ACTIVITY: 'sh_security_last_activity',
  AUTO_LOCK_ENABLED: 'sh_security_auto_lock',
  AUTO_LOCK_TIMEOUT: 'sh_security_timeout'
};

// Security configuration
const SECURITY_CONFIG = {
  MAX_FAILED_ATTEMPTS: 5,          // Lock after 5 failed attempts
  INITIAL_LOCKOUT_MS: 30_000,      // 30 seconds initial lockout
  MAX_LOCKOUT_MS: 3600_000,        // 1 hour max lockout
  LOCKOUT_MULTIPLIER: 2,           // Double lockout time each failure
  AUTO_LOCK_DEFAULT_MS: 300_000,   // 5 minutes default auto-lock
  PIN_MIN_LENGTH: 6,
  PIN_MAX_LENGTH: 8
};

// Weak PINs that should be blocked
const WEAK_PINS = new Set([
  '000000', '111111', '222222', '333333', '444444',
  '555555', '666666', '777777', '888888', '999999',
  '123456', '654321', '123123', '112233', '121212',
  '696969', '000001', '100000', '420420', '112211',
  '131313', '141414', '151515', '161616', '171717',
  '181818', '191919', '101010', '102030', '010203',
  '123321', '789456', '456789', '987654', '147258',
  '258369', '369258', '159753', '357159', '246810'
]);

// ============================================
// Brute Force Protection
// ============================================

export interface SecurityState {
  failedAttempts: number;
  lockoutUntil: number | null;
  isLocked: boolean;
  lockoutRemainingMs: number;
}

/**
 * Get current security state
 */
export function getSecurityState(): SecurityState {
  const failedAttempts = parseInt(
    localStorage.getItem(SECURITY_KEYS.FAILED_ATTEMPTS) || '0',
    10
  );
  const lockoutUntil = localStorage.getItem(SECURITY_KEYS.LOCKOUT_UNTIL);
  const lockoutTimestamp = lockoutUntil ? parseInt(lockoutUntil, 10) : null;

  const now = Date.now();
  const isLocked = lockoutTimestamp !== null && lockoutTimestamp > now;
  const lockoutRemainingMs = isLocked ? lockoutTimestamp! - now : 0;

  return {
    failedAttempts,
    lockoutUntil: lockoutTimestamp,
    isLocked,
    lockoutRemainingMs
  };
}

/**
 * Check if user is currently locked out
 */
export function isLockedOut(): boolean {
  return getSecurityState().isLocked;
}

/**
 * Get remaining lockout time in seconds
 */
export function getLockoutRemaining(): number {
  const { lockoutRemainingMs } = getSecurityState();
  return Math.ceil(lockoutRemainingMs / 1000);
}

/**
 * Record a failed PIN attempt
 * Implements exponential backoff lockout
 */
export function recordFailedAttempt(): SecurityState {
  const state = getSecurityState();

  // If currently locked, don't increment (user shouldn't be able to try)
  if (state.isLocked) {
    return state;
  }

  const newFailedAttempts = state.failedAttempts + 1;
  localStorage.setItem(SECURITY_KEYS.FAILED_ATTEMPTS, newFailedAttempts.toString());

  // Check if we should lock the account
  if (newFailedAttempts >= SECURITY_CONFIG.MAX_FAILED_ATTEMPTS) {
    // Calculate lockout duration with exponential backoff
    const lockoutMultiplier = Math.pow(
      SECURITY_CONFIG.LOCKOUT_MULTIPLIER,
      Math.floor(newFailedAttempts / SECURITY_CONFIG.MAX_FAILED_ATTEMPTS) - 1
    );
    const lockoutDuration = Math.min(
      SECURITY_CONFIG.INITIAL_LOCKOUT_MS * lockoutMultiplier,
      SECURITY_CONFIG.MAX_LOCKOUT_MS
    );

    const lockoutUntil = Date.now() + lockoutDuration;
    localStorage.setItem(SECURITY_KEYS.LOCKOUT_UNTIL, lockoutUntil.toString());

    return {
      failedAttempts: newFailedAttempts,
      lockoutUntil,
      isLocked: true,
      lockoutRemainingMs: lockoutDuration
    };
  }

  return {
    ...state,
    failedAttempts: newFailedAttempts
  };
}

/**
 * Reset failed attempts on successful unlock
 */
export function resetFailedAttempts(): void {
  localStorage.removeItem(SECURITY_KEYS.FAILED_ATTEMPTS);
  localStorage.removeItem(SECURITY_KEYS.LOCKOUT_UNTIL);
}

/**
 * Get number of remaining attempts before lockout
 */
export function getRemainingAttempts(): number {
  const { failedAttempts } = getSecurityState();
  return Math.max(0, SECURITY_CONFIG.MAX_FAILED_ATTEMPTS - failedAttempts);
}

// ============================================
// PIN Complexity Validation
// ============================================

export interface PinValidationResult {
  isValid: boolean;
  errors: string[];
  strength: 'weak' | 'medium' | 'strong';
}

/**
 * Validate PIN complexity
 */
export function validatePinComplexity(pin: string): PinValidationResult {
  const errors: string[] = [];

  // Check length
  if (pin.length < SECURITY_CONFIG.PIN_MIN_LENGTH) {
    errors.push(`PIN must be at least ${SECURITY_CONFIG.PIN_MIN_LENGTH} digits`);
  }
  if (pin.length > SECURITY_CONFIG.PIN_MAX_LENGTH) {
    errors.push(`PIN must be at most ${SECURITY_CONFIG.PIN_MAX_LENGTH} digits`);
  }

  // Check if all digits
  if (!/^\d+$/.test(pin)) {
    errors.push('PIN must contain only numbers');
  }

  // Check for weak PINs
  if (WEAK_PINS.has(pin)) {
    errors.push('This PIN is too common and easily guessed');
  }

  // Check for sequential patterns
  if (isSequential(pin)) {
    errors.push('PIN contains sequential numbers');
  }

  // Check for repeated patterns
  if (hasRepeatedPattern(pin)) {
    errors.push('PIN contains repeated patterns');
  }

  // Calculate strength
  let strength: 'weak' | 'medium' | 'strong' = 'strong';
  if (errors.length > 0) {
    strength = 'weak';
  } else if (hasModerateWeakness(pin)) {
    strength = 'medium';
  }

  return {
    isValid: errors.length === 0,
    errors,
    strength
  };
}

/**
 * Check if PIN is sequential (ascending or descending)
 */
function isSequential(pin: string): boolean {
  if (pin.length < 3) return false;

  const digits = pin.split('').map(Number);

  // Check ascending
  let isAscending = true;
  let isDescending = true;

  for (let i = 1; i < digits.length; i++) {
    if (digits[i] !== digits[i - 1] + 1) isAscending = false;
    if (digits[i] !== digits[i - 1] - 1) isDescending = false;
  }

  return isAscending || isDescending;
}

/**
 * Check if PIN has repeated patterns like 123123 or 121212
 */
function hasRepeatedPattern(pin: string): boolean {
  if (pin.length < 4) return false;

  // Check for 2-digit patterns
  if (pin.length >= 4) {
    const pattern2 = pin.slice(0, 2);
    if (pin === pattern2.repeat(pin.length / 2)) return true;
  }

  // Check for 3-digit patterns
  if (pin.length >= 6) {
    const pattern3 = pin.slice(0, 3);
    if (pin === pattern3.repeat(pin.length / 3)) return true;
  }

  // Check if all digits are the same
  if (new Set(pin).size === 1) return true;

  return false;
}

/**
 * Check for moderate weaknesses (not blocking, but affects strength)
 */
function hasModerateWeakness(pin: string): boolean {
  const digits = pin.split('').map(Number);

  // Check if less than 4 unique digits
  if (new Set(digits).size < 4) return true;

  // Check for common patterns that aren't blocked
  const commonPatterns = [
    /^(\d)\1\1/,    // Starts with 3 same digits
    /(\d)\1\1$/,    // Ends with 3 same digits
  ];

  return commonPatterns.some(pattern => pattern.test(pin));
}

// ============================================
// Auto-Lock Management
// ============================================

/**
 * Update last activity timestamp
 */
export function updateLastActivity(): void {
  localStorage.setItem(SECURITY_KEYS.LAST_ACTIVITY, Date.now().toString());
}

/**
 * Get last activity timestamp
 */
export function getLastActivity(): number {
  const timestamp = localStorage.getItem(SECURITY_KEYS.LAST_ACTIVITY);
  return timestamp ? parseInt(timestamp, 10) : Date.now();
}

/**
 * Check if auto-lock should trigger
 */
export function shouldAutoLock(): boolean {
  const enabled = localStorage.getItem(SECURITY_KEYS.AUTO_LOCK_ENABLED);
  if (enabled === 'false') return false;

  const timeout = getAutoLockTimeout();
  const lastActivity = getLastActivity();
  const now = Date.now();

  return now - lastActivity > timeout;
}

/**
 * Get auto-lock timeout in milliseconds
 */
export function getAutoLockTimeout(): number {
  const timeout = localStorage.getItem(SECURITY_KEYS.AUTO_LOCK_TIMEOUT);
  return timeout ? parseInt(timeout, 10) : SECURITY_CONFIG.AUTO_LOCK_DEFAULT_MS;
}

/**
 * Set auto-lock timeout
 */
export function setAutoLockTimeout(timeoutMs: number): void {
  localStorage.setItem(SECURITY_KEYS.AUTO_LOCK_TIMEOUT, timeoutMs.toString());
}

/**
 * Enable/disable auto-lock
 */
export function setAutoLockEnabled(enabled: boolean): void {
  localStorage.setItem(SECURITY_KEYS.AUTO_LOCK_ENABLED, enabled.toString());
}

/**
 * Check if auto-lock is enabled
 */
export function isAutoLockEnabled(): boolean {
  const enabled = localStorage.getItem(SECURITY_KEYS.AUTO_LOCK_ENABLED);
  return enabled !== 'false'; // Default to true
}

// ============================================
// Security Utilities
// ============================================

/**
 * Format lockout time for display
 */
export function formatLockoutTime(seconds: number): string {
  if (seconds < 60) {
    return `${seconds} second${seconds !== 1 ? 's' : ''}`;
  }
  if (seconds < 3600) {
    const minutes = Math.ceil(seconds / 60);
    return `${minutes} minute${minutes !== 1 ? 's' : ''}`;
  }
  const hours = Math.ceil(seconds / 3600);
  return `${hours} hour${hours !== 1 ? 's' : ''}`;
}

/**
 * Clear all security data (for wallet reset)
 */
export function clearSecurityData(): void {
  localStorage.removeItem(SECURITY_KEYS.FAILED_ATTEMPTS);
  localStorage.removeItem(SECURITY_KEYS.LOCKOUT_UNTIL);
  localStorage.removeItem(SECURITY_KEYS.LAST_ACTIVITY);
  localStorage.removeItem(SECURITY_KEYS.AUTO_LOCK_ENABLED);
  localStorage.removeItem(SECURITY_KEYS.AUTO_LOCK_TIMEOUT);
}

/**
 * Create auto-lock monitor that calls lockCallback when timeout expires
 * Also listens for visibility changes to lock when app goes to background
 *
 * SECURITY: Reduced default check interval from 10s to 2s for faster response
 */
export function createAutoLockMonitor(
  lockCallback: () => void,
  checkIntervalMs: number = 2000,  // Reduced from 10s to 2s for better security
  lockOnHidden: boolean = true     // Lock immediately when tab/app is hidden
): () => void {
  const checkAutoLock = () => {
    if (shouldAutoLock()) {
      lockCallback();
    }
  };

  // SECURITY: Lock immediately when user switches away from app
  const handleVisibilityChange = () => {
    if (lockOnHidden && document.hidden) {
      // User switched to another app/tab - lock after a brief delay
      // The delay allows for brief tab switches without locking
      setTimeout(() => {
        if (document.hidden) {
          lockCallback();
        }
      }, 1000);  // 1 second grace period
    }
  };

  const intervalId = setInterval(checkAutoLock, checkIntervalMs);

  // Add visibility change listener
  if (typeof document !== 'undefined') {
    document.addEventListener('visibilitychange', handleVisibilityChange);
  }

  // Return cleanup function
  return () => {
    clearInterval(intervalId);
    if (typeof document !== 'undefined') {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    }
  };
}
