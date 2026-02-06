/**
 * Secure Memory Handling Utilities
 *
 * SECURITY: Provides secure erasure and handling of sensitive data in memory.
 * Note: In JavaScript, true secure erasure is limited due to garbage collection
 * and string immutability. These utilities provide best-effort protection.
 */

/**
 * Securely overwrite a Uint8Array with random data then zeros
 */
export function secureWipeArray(arr: Uint8Array | number[]): void {
  if (!arr || arr.length === 0) return;

  try {
    // First pass: overwrite with random data
    if (arr instanceof Uint8Array && typeof crypto !== 'undefined' && crypto.getRandomValues) {
      crypto.getRandomValues(arr);
    } else {
      for (let i = 0; i < arr.length; i++) {
        arr[i] = Math.floor(Math.random() * 256);
      }
    }

    // Second pass: overwrite with zeros
    for (let i = 0; i < arr.length; i++) {
      arr[i] = 0;
    }

    // Third pass: overwrite with ones
    for (let i = 0; i < arr.length; i++) {
      arr[i] = 0xFF;
    }

    // Final pass: zeros
    for (let i = 0; i < arr.length; i++) {
      arr[i] = 0;
    }
  } catch {
    // Best effort - at minimum zero out
    for (let i = 0; i < arr.length; i++) {
      arr[i] = 0;
    }
  }
}

/**
 * Create a mnemonic container that can be securely wiped
 * Uses a Uint8Array internally for better memory control
 */
export class SecureMnemonic {
  private data: Uint8Array | null = null;
  private _isCleared = false;

  constructor(mnemonic?: string) {
    if (mnemonic) {
      this.set(mnemonic);
    }
  }

  set(mnemonic: string): void {
    // Clear existing data first
    this.clear();

    // Store as bytes
    const encoder = new TextEncoder();
    this.data = encoder.encode(mnemonic);
    this._isCleared = false;
  }

  get(): string {
    if (this._isCleared || !this.data) {
      return '';
    }
    const decoder = new TextDecoder();
    return decoder.decode(this.data);
  }

  getWords(): string[] {
    return this.get().split(' ').filter(w => w.length > 0);
  }

  get isCleared(): boolean {
    return this._isCleared;
  }

  clear(): void {
    if (this.data) {
      secureWipeArray(this.data);
      this.data = null;
    }
    this._isCleared = true;
  }
}

/**
 * Secure PIN container - minimizes PIN exposure in memory
 */
export class SecurePin {
  private data: Uint8Array | null = null;

  set(pin: string): void {
    this.clear();
    const encoder = new TextEncoder();
    this.data = encoder.encode(pin);
  }

  get(): string {
    if (!this.data) return '';
    const decoder = new TextDecoder();
    return decoder.decode(this.data);
  }

  get length(): number {
    return this.data?.length || 0;
  }

  clear(): void {
    if (this.data) {
      secureWipeArray(this.data);
      this.data = null;
    }
  }
}

/**
 * Schedule cleanup of sensitive data
 * Returns a cleanup function that should be called on component unmount
 */
export function scheduleSecureCleanup(
  ...items: (SecureMnemonic | SecurePin | { clear: () => void })[]
): () => void {
  return () => {
    for (const item of items) {
      try {
        item.clear();
      } catch {
        // Best effort cleanup
      }
    }
  };
}

/**
 * Create a temporary mnemonic that auto-clears after a timeout
 */
export function createTemporaryMnemonic(
  mnemonic: string,
  timeoutMs: number = 300000 // 5 minutes default
): { getMnemonic: () => string; clear: () => void } {
  const secure = new SecureMnemonic(mnemonic);

  const timeoutId = setTimeout(() => {
    secure.clear();
  }, timeoutMs);

  return {
    getMnemonic: () => secure.get(),
    clear: () => {
      clearTimeout(timeoutId);
      secure.clear();
    }
  };
}
