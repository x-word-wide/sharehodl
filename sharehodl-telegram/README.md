# ShareHODL Telegram Mini App

A secure, feature-rich Telegram Mini App for the ShareHODL blockchain platform.

## Architecture

```
sharehodl-telegram/
├── bot/                    # Telegram Bot Backend (Node.js)
│   ├── src/
│   │   ├── bot.ts         # Main bot logic
│   │   ├── handlers/      # Command handlers
│   │   └── services/      # API services
│   └── package.json
│
└── webapp/                 # Mini App Frontend (React + Vite)
    ├── src/
    │   ├── components/    # Reusable UI components
    │   ├── screens/       # App screens
    │   ├── services/      # Blockchain & API services
    │   ├── hooks/         # Custom React hooks
    │   ├── utils/         # Crypto utilities
    │   └── types/         # TypeScript types
    └── package.json
```

## Security Model

- All cryptographic operations happen **client-side only**
- Private keys and mnemonics **never** leave the device
- Uses Web Crypto API for secure random number generation
- Encrypted local storage with user PIN
- Session timeouts for inactive users

## Features

### Core Wallet
- Create new wallet (BIP39 mnemonic)
- Import existing wallet
- Multi-chain support (ShareHODL, Ethereum, Bitcoin, Cosmos chains)
- Send & receive tokens
- Real-time balance updates

### Equity Trading (Primary Focus)
- View tokenized equity holdings
- Real-time market data
- Buy/Sell equity tokens
- Portfolio analytics

### DeFi Services
- P2P Trading marketplace
- Lending & Borrowing
- Inheritance planning

### ShareHODL Network
- Native HODL token support
- Crypto-to-HODL bridge
- Staking rewards
- Governance participation

## Development

### Prerequisites
- Node.js 18+
- pnpm (recommended) or npm
- Telegram Bot Token (from @BotFather)

### Setup Bot
```bash
cd bot
pnpm install
cp .env.example .env
# Add your BOT_TOKEN to .env
pnpm dev
```

### Setup Web App
```bash
cd webapp
pnpm install
pnpm dev
```

### Deploy
```bash
# Build webapp
cd webapp && pnpm build

# Deploy to your hosting (Vercel, Netlify, etc.)
# Update bot webhook URL
```

## Environment Variables

### Bot (.env)
```
BOT_TOKEN=your_telegram_bot_token
WEBAPP_URL=https://your-webapp-url.com
SHAREHODL_RPC=https://rpc.sharehodl.network
```

### Webapp (.env)
```
VITE_SHAREHODL_RPC=https://rpc.sharehodl.network
VITE_SHAREHODL_REST=https://api.sharehodl.network
VITE_COINGECKO_API=https://api.coingecko.com/api/v3
```

## License
MIT - ShareHODL Team
