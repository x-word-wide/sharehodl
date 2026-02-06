/**
 * Production-safe Logging Utility
 *
 * SECURITY: This wrapper prevents sensitive information from being logged in production.
 * In development mode, all logs are shown. In production, only errors are logged
 * and sensitive data patterns are filtered.
 */

type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LoggerConfig {
  /** Enable logging (defaults to development mode) */
  enabled: boolean;
  /** Minimum log level to display */
  minLevel: LogLevel;
  /** Prefix for all log messages */
  prefix: string;
}

const LOG_LEVELS: Record<LogLevel, number> = {
  debug: 0,
  info: 1,
  warn: 2,
  error: 3
};

// Detect if we're in production mode
const isProduction = import.meta.env.PROD || import.meta.env.MODE === 'production';

const defaultConfig: LoggerConfig = {
  enabled: true,
  minLevel: isProduction ? 'error' : 'debug',
  prefix: '[ShareHODL]'
};

let config: LoggerConfig = { ...defaultConfig };

/**
 * Patterns that indicate sensitive data - these will be redacted in production
 */
const SENSITIVE_PATTERNS = [
  /mnemonic/i,
  /seed/i,
  /private\s*key/i,
  /secret/i,
  /password/i,
  /pin/i,
  /token/i,
  /credential/i,
  /\b[a-zA-Z]{3,}\s+[a-zA-Z]{3,}\s+[a-zA-Z]{3,}\s+[a-zA-Z]{3,}/i, // Potential mnemonic phrases
];

/**
 * Check if a value might contain sensitive data
 */
function containsSensitiveData(value: unknown): boolean {
  if (typeof value !== 'string') {
    value = JSON.stringify(value);
  }
  return SENSITIVE_PATTERNS.some(pattern => pattern.test(value as string));
}

/**
 * Sanitize a value for logging, redacting sensitive information
 */
function sanitize(value: unknown): unknown {
  if (!isProduction) return value;

  if (typeof value === 'string') {
    if (containsSensitiveData(value)) {
      return '[REDACTED]';
    }
    return value;
  }

  if (typeof value === 'object' && value !== null) {
    if (Array.isArray(value)) {
      return value.map(sanitize);
    }

    const sanitized: Record<string, unknown> = {};
    for (const [key, val] of Object.entries(value)) {
      if (SENSITIVE_PATTERNS.some(pattern => pattern.test(key))) {
        sanitized[key] = '[REDACTED]';
      } else {
        sanitized[key] = sanitize(val);
      }
    }
    return sanitized;
  }

  return value;
}

/**
 * Check if a log should be displayed based on current configuration
 */
function shouldLog(level: LogLevel): boolean {
  if (!config.enabled) return false;
  return LOG_LEVELS[level] >= LOG_LEVELS[config.minLevel];
}

/**
 * Format arguments for logging
 */
function formatArgs(args: unknown[]): unknown[] {
  return args.map(arg => sanitize(arg));
}

/**
 * Production-safe logger
 */
export const logger = {
  /**
   * Debug level - only shown in development
   */
  debug: (...args: unknown[]) => {
    if (shouldLog('debug')) {
      console.log(config.prefix, ...formatArgs(args));
    }
  },

  /**
   * Info level - only shown in development
   */
  info: (...args: unknown[]) => {
    if (shouldLog('info')) {
      console.info(config.prefix, ...formatArgs(args));
    }
  },

  /**
   * Warning level - shown in development, hidden in production by default
   */
  warn: (...args: unknown[]) => {
    if (shouldLog('warn')) {
      console.warn(config.prefix, ...formatArgs(args));
    }
  },

  /**
   * Error level - always shown (unless explicitly disabled)
   * Sensitive data is still redacted
   */
  error: (...args: unknown[]) => {
    if (shouldLog('error')) {
      console.error(config.prefix, ...formatArgs(args));
    }
  },

  /**
   * Configure the logger
   */
  configure: (newConfig: Partial<LoggerConfig>) => {
    config = { ...config, ...newConfig };
  },

  /**
   * Get current configuration
   */
  getConfig: (): Readonly<LoggerConfig> => ({ ...config }),

  /**
   * Check if running in production mode
   */
  isProduction: () => isProduction,

  /**
   * Temporarily enable all logging (useful for debugging production issues)
   * Returns a function to restore previous settings
   */
  enableDebug: (): (() => void) => {
    const previousConfig = { ...config };
    config.minLevel = 'debug';
    return () => {
      config = previousConfig;
    };
  }
};

export default logger;
