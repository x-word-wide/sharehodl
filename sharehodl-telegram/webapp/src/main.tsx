// Buffer polyfill for browser environment (required by bip39)
import { Buffer as BufferPolyfill } from 'buffer';
if (typeof globalThis.Buffer === 'undefined') {
  globalThis.Buffer = BufferPolyfill;
}

import React from 'react';
import ReactDOM from 'react-dom/client';
import '@telegram-apps/telegram-ui/dist/styles.css';
import App from './App';
import './index.css';
import { ErrorBoundary } from './components/ErrorBoundary';
import { logger } from './utils/logger';

logger.debug('main.tsx loaded');

// Initialize Telegram Web App
declare global {
  interface Window {
    Telegram: {
      WebApp: TelegramWebApp;
    };
  }

  interface TelegramWebApp {
    ready: () => void;
    expand: () => void;
    close: () => void;
    MainButton: MainButton;
    BackButton: BackButton;
    initData: string;
    initDataUnsafe: WebAppInitData;
    colorScheme: 'light' | 'dark';
    themeParams: ThemeParams;
    isExpanded: boolean;
    viewportHeight: number;
    viewportStableHeight: number;
    headerColor: string;
    backgroundColor: string;
    setHeaderColor: (color: string) => void;
    setBackgroundColor: (color: string) => void;
    enableClosingConfirmation: () => void;
    disableClosingConfirmation: () => void;
    onEvent: (eventType: string, callback: () => void) => void;
    offEvent: (eventType: string, callback: () => void) => void;
    sendData: (data: string) => void;
    openLink: (url: string) => void;
    openTelegramLink: (url: string) => void;
    showPopup: (params: PopupParams, callback?: (buttonId: string) => void) => void;
    showAlert: (message: string, callback?: () => void) => void;
    showConfirm: (message: string, callback?: (confirmed: boolean) => void) => void;
    HapticFeedback: HapticFeedback;
    CloudStorage: CloudStorage;
  }

  interface MainButton {
    text: string;
    color: string;
    textColor: string;
    isVisible: boolean;
    isActive: boolean;
    isProgressVisible: boolean;
    setText: (text: string) => MainButton;
    onClick: (callback: () => void) => MainButton;
    offClick: (callback: () => void) => MainButton;
    show: () => MainButton;
    hide: () => MainButton;
    enable: () => MainButton;
    disable: () => MainButton;
    showProgress: (leaveActive?: boolean) => MainButton;
    hideProgress: () => MainButton;
  }

  interface BackButton {
    isVisible: boolean;
    onClick: (callback: () => void) => BackButton;
    offClick: (callback: () => void) => BackButton;
    show: () => BackButton;
    hide: () => BackButton;
  }

  interface WebAppInitData {
    query_id?: string;
    user?: WebAppUser;
    receiver?: WebAppUser;
    chat?: WebAppChat;
    start_param?: string;
    can_send_after?: number;
    auth_date: number;
    hash: string;
  }

  interface WebAppUser {
    id: number;
    is_bot?: boolean;
    first_name: string;
    last_name?: string;
    username?: string;
    language_code?: string;
    is_premium?: boolean;
    photo_url?: string;
  }

  interface WebAppChat {
    id: number;
    type: string;
    title?: string;
    username?: string;
    photo_url?: string;
  }

  interface ThemeParams {
    bg_color?: string;
    text_color?: string;
    hint_color?: string;
    link_color?: string;
    button_color?: string;
    button_text_color?: string;
    secondary_bg_color?: string;
  }

  interface PopupParams {
    title?: string;
    message: string;
    buttons?: PopupButton[];
  }

  interface PopupButton {
    id?: string;
    type?: 'default' | 'ok' | 'close' | 'cancel' | 'destructive';
    text?: string;
  }

  interface HapticFeedback {
    impactOccurred: (style: 'light' | 'medium' | 'heavy' | 'rigid' | 'soft') => void;
    notificationOccurred: (type: 'error' | 'success' | 'warning') => void;
    selectionChanged: () => void;
  }

  interface CloudStorage {
    setItem: (key: string, value: string, callback?: (error: Error | null, stored: boolean) => void) => void;
    getItem: (key: string, callback: (error: Error | null, value: string | null) => void) => void;
    getItems: (keys: string[], callback: (error: Error | null, values: Record<string, string>) => void) => void;
    removeItem: (key: string, callback?: (error: Error | null, removed: boolean) => void) => void;
    removeItems: (keys: string[], callback?: (error: Error | null, removed: boolean) => void) => void;
    getKeys: (callback: (error: Error | null, keys: string[]) => void) => void;
  }
}

// Initialize Telegram WebApp
const tg = window.Telegram?.WebApp;
if (tg) {
  tg.ready();
  tg.expand();
  tg.setHeaderColor('#0D1117');
  tg.setBackgroundColor('#0D1117');
  tg.enableClosingConfirmation();
}

logger.debug('Starting app render...');

const rootElement = document.getElementById('root');
logger.debug('Root element:', rootElement);

if (rootElement) {
  try {
    ReactDOM.createRoot(rootElement).render(
      <React.StrictMode>
        <ErrorBoundary>
          <App />
        </ErrorBoundary>
      </React.StrictMode>
    );
    logger.debug('App rendered successfully');
  } catch (error) {
    logger.error('Error rendering app:', error);
    rootElement.innerHTML = `<div style="color: white; padding: 20px;">Error: ${error}</div>`;
  }
} else {
  logger.error('Root element not found!');
}
