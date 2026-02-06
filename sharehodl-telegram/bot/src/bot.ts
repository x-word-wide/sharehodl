/**
 * ShareHODL Telegram Bot
 *
 * Main entry point for the Telegram bot that launches the Mini App
 */

import { Telegraf, Markup } from 'telegraf';
import express from 'express';
import dotenv from 'dotenv';

dotenv.config();

const BOT_TOKEN = process.env.BOT_TOKEN;
const WEBAPP_URL = process.env.WEBAPP_URL || 'https://app.sharehodl.network';
const PORT = process.env.PORT || 3000;

if (!BOT_TOKEN) {
  throw new Error('BOT_TOKEN environment variable is required');
}

const bot = new Telegraf(BOT_TOKEN);

// ============================================
// Bot Commands
// ============================================

// /start - Welcome message with Mini App button
bot.start(async (ctx) => {
  const firstName = ctx.from?.first_name || 'there';

  await ctx.reply(
    `Welcome to ShareHODL, ${firstName}! 🚀\n\n` +
    `ShareHODL is the future of tokenized equity trading.\n\n` +
    `✅ Trade tokenized stocks 24/7\n` +
    `✅ Instant settlement\n` +
    `✅ Ultra-low fees (<$0.01)\n` +
    `✅ Multi-chain wallet\n` +
    `✅ P2P Trading & DeFi\n\n` +
    `Tap the button below to open your wallet:`,
    Markup.inlineKeyboard([
      [Markup.button.webApp('🚀 Open ShareHODL Wallet', WEBAPP_URL)],
      [Markup.button.url('📖 Learn More', 'https://sharehodl.network')]
    ])
  );
});

// /wallet - Quick access to wallet
bot.command('wallet', async (ctx) => {
  await ctx.reply(
    '💼 Access your ShareHODL wallet:',
    Markup.inlineKeyboard([
      [Markup.button.webApp('Open Wallet', `${WEBAPP_URL}?screen=portfolio`)]
    ])
  );
});

// /trade - Quick access to trading
bot.command('trade', async (ctx) => {
  await ctx.reply(
    '📈 Ready to trade? Choose an option:',
    Markup.inlineKeyboard([
      [Markup.button.webApp('🏦 Equity Trading', `${WEBAPP_URL}?screen=trade`)],
      [Markup.button.webApp('🤝 P2P Trading', `${WEBAPP_URL}?screen=p2p`)],
      [Markup.button.webApp('📊 Market', `${WEBAPP_URL}?screen=market`)]
    ])
  );
});

// /portfolio - View portfolio
bot.command('portfolio', async (ctx) => {
  await ctx.reply(
    '📊 View your portfolio:',
    Markup.inlineKeyboard([
      [Markup.button.webApp('View Portfolio', `${WEBAPP_URL}?screen=portfolio`)]
    ])
  );
});

// /send - Send tokens
bot.command('send', async (ctx) => {
  await ctx.reply(
    '📤 Send tokens:',
    Markup.inlineKeyboard([
      [Markup.button.webApp('Send Tokens', `${WEBAPP_URL}?screen=send`)]
    ])
  );
});

// /receive - Receive tokens
bot.command('receive', async (ctx) => {
  await ctx.reply(
    '📥 Receive tokens:',
    Markup.inlineKeyboard([
      [Markup.button.webApp('Receive Tokens', `${WEBAPP_URL}?screen=receive`)]
    ])
  );
});

// /defi - DeFi services menu
bot.command('defi', async (ctx) => {
  await ctx.reply(
    '🏦 DeFi Services:',
    Markup.inlineKeyboard([
      [Markup.button.webApp('🤝 P2P Trading', `${WEBAPP_URL}?screen=p2p`)],
      [Markup.button.webApp('💰 Lending', `${WEBAPP_URL}?screen=lending`)],
      [Markup.button.webApp('👨‍👩‍👧 Inheritance', `${WEBAPP_URL}?screen=inheritance`)]
    ])
  );
});

// /bridge - Crypto to HODL bridge
bot.command('bridge', async (ctx) => {
  await ctx.reply(
    '🌉 Bridge your crypto to HODL tokens:\n\n' +
    'Convert BTC, ETH, ATOM and more to HODL for equity trading.',
    Markup.inlineKeyboard([
      [Markup.button.webApp('Open Bridge', `${WEBAPP_URL}?screen=bridge`)]
    ])
  );
});

// /settings - Settings
bot.command('settings', async (ctx) => {
  await ctx.reply(
    '⚙️ Wallet Settings:',
    Markup.inlineKeyboard([
      [Markup.button.webApp('Open Settings', `${WEBAPP_URL}?screen=settings`)]
    ])
  );
});

// /help - Help menu
bot.command('help', async (ctx) => {
  await ctx.reply(
    '📚 ShareHODL Commands:\n\n' +
    '/start - Welcome message\n' +
    '/wallet - Open your wallet\n' +
    '/portfolio - View portfolio\n' +
    '/trade - Trading options\n' +
    '/send - Send tokens\n' +
    '/receive - Receive tokens\n' +
    '/defi - DeFi services\n' +
    '/bridge - Crypto to HODL bridge\n' +
    '/settings - Wallet settings\n' +
    '/help - This help message\n\n' +
    '🔒 Security Note:\n' +
    'Your private keys are stored securely on your device and never leave it.',
    Markup.inlineKeyboard([
      [Markup.button.webApp('🚀 Open ShareHODL', WEBAPP_URL)]
    ])
  );
});

// Handle inline button callbacks
bot.on('callback_query', async (ctx) => {
  await ctx.answerCbQuery();
});

// ============================================
// Express Server for Webhooks
// ============================================

const app = express();
app.use(express.json());

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'ok', bot: 'ShareHODL Telegram Bot' });
});

// API endpoints for Mini App
app.get('/api/price/:symbol', async (req, res) => {
  try {
    const { symbol } = req.params;
    // Return mock price for now - integrate with real price service
    const prices: Record<string, number> = {
      'HODL': 1.00,
      'BTC': 67500,
      'ETH': 3450,
      'ATOM': 8.50,
      'OSMO': 0.85
    };
    res.json({ symbol, price: prices[symbol] || 0 });
  } catch (error) {
    res.status(500).json({ error: 'Failed to fetch price' });
  }
});

// ============================================
// Start Bot
// ============================================

async function start() {
  // Set bot commands
  await bot.telegram.setMyCommands([
    { command: 'start', description: 'Welcome to ShareHODL' },
    { command: 'wallet', description: 'Open your wallet' },
    { command: 'portfolio', description: 'View portfolio' },
    { command: 'trade', description: 'Trading options' },
    { command: 'send', description: 'Send tokens' },
    { command: 'receive', description: 'Receive tokens' },
    { command: 'defi', description: 'DeFi services' },
    { command: 'bridge', description: 'Crypto to HODL bridge' },
    { command: 'settings', description: 'Wallet settings' },
    { command: 'help', description: 'Help & commands' }
  ]);

  // Start Express server
  app.listen(PORT, () => {
    console.log(`🚀 ShareHODL Bot API running on port ${PORT}`);
  });

  // Start bot with polling (use webhooks in production)
  await bot.launch();
  console.log('🤖 ShareHODL Telegram Bot started!');
}

// Graceful shutdown
process.once('SIGINT', () => bot.stop('SIGINT'));
process.once('SIGTERM', () => bot.stop('SIGTERM'));

start().catch(console.error);
